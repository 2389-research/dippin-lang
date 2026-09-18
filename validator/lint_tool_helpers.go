package validator

import (
	"regexp"
	"strings"

	"github.com/2389-research/dippin-lang/ir"
	"mvdan.cc/sh/v3/syntax"
)

// ctxVarPattern matches ${ctx.*} references in tool commands.
var ctxVarPattern = regexp.MustCompile(`\$\{ctx\.[^}]+\}`)

// lintToolCtxVars flags ${ctx.*} references in tool commands that won't
// resolve at parse time. DIP124: tool command references runtime variable.
func lintToolCtxVars(w *ir.Workflow) []Diagnostic {
	var diags []Diagnostic
	for _, n := range w.Nodes {
		diags = append(diags, checkToolCtxVars(n)...)
	}
	return diags
}

// checkToolCtxVars checks a single tool node for ${ctx.*} references.
func checkToolCtxVars(n *ir.Node) []Diagnostic {
	cfg, ok := n.Config.(ir.ToolConfig)
	if !ok || cfg.Command == "" {
		return nil
	}
	matches := ctxVarPattern.FindAllString(cfg.Command, -1)
	if len(matches) == 0 {
		return nil
	}
	var diags []Diagnostic
	for _, m := range matches {
		diags = append(diags, Diagnostic{
			Code:     DIP124,
			Severity: SeverityWarning,
			Message:  "tool command references " + m + " which expands to empty at runtime",
			Location: n.Source,
		})
	}
	return diags
}

// placeholderDummy replaces every ${...} placeholder before shell parsing.
// Dippin placeholders like ${graph.workflow_dir} or ${params.foo} contain a
// "." inside the braces, which is not valid shell parameter-expansion syntax
// and causes the mvdan shell parser to fail, misdirecting extraction to the
// fallback word-splitter. Substituting a shell-safe dummy word lets the
// parser succeed so the normal skip-set/assignment extraction logic runs.
const placeholderDummy = "__dip_placeholder__"

// extractBinary parses a shell command and returns the first non-builtin,
// non-preamble command name. Uses a proper shell AST parser to correctly
// handle variable assignments, pipes, subshells, command substitution, etc.
// Shell builtins and preamble commands (mkdir) are skipped to find the
// primary tool binary. Falls back to token-based extraction on parse errors.
// Dippin ${ns.key} placeholders are substituted with a dummy word first so
// they don't break the shell parse (see placeholderDummy); if the extracted
// binary name still contains the dummy (whether it *is* the placeholder or
// the placeholder was concatenated with literal text, e.g. "${p.x}suffix"),
// extraction returns "" since the real binary can't be known before
// expansion. Note: varRefPattern also matches plain shell ${VAR} expansions
// (it doesn't distinguish a dippin ${ns.key} from e.g. ${HOME} or ${TOOL}),
// so a command whose first word is a plain shell variable expansion
// (${TOOL} --flag, ${HOME}/bin/tool) now yields "" instead of the walker
// skipping past it to a later command. This is an acceptable, intentional
// tradeoff: it drops a hint rather than ever fabricating a wrong one.
func extractBinary(command string) string {
	sanitized := varRefPattern.ReplaceAllString(command, placeholderDummy)
	parser := syntax.NewParser(syntax.KeepComments(false))
	prog, err := parser.Parse(strings.NewReader(sanitized), "")
	var bin string
	if err != nil {
		bin = extractBinaryFallback(sanitized)
	} else {
		syntax.Walk(prog, walkForBinary(&bin))
	}
	if strings.Contains(bin, placeholderDummy) {
		return ""
	}
	return bin
}

// extractBinaryFallback performs best-effort extraction when shell parsing
// fails. Skips builtins and preamble, returns the first plausible binary.
func extractBinaryFallback(command string) string {
	for _, field := range strings.Fields(command) {
		if !isSkippableCommand(field) {
			return field
		}
	}
	return ""
}

// walkForBinary returns a walk function that captures the first non-builtin,
// non-preamble command binary into bin.
func walkForBinary(bin *string) func(syntax.Node) bool {
	return func(node syntax.Node) bool {
		if *bin != "" {
			return false
		}
		name := callExprBinary(node)
		if name != "" && !isSkippableCommand(name) {
			*bin = name
			return false
		}
		return true
	}
}

// callExprBinary returns the literal binary name of a CallExpr node.
// Handles "command" specially: "command -v foo" is a query (returns ""),
// "command foo" executes foo (returns "foo").
func callExprBinary(node syntax.Node) string {
	call, ok := node.(*syntax.CallExpr)
	if !ok || len(call.Args) == 0 {
		return ""
	}
	name := extractWordLiteral(call.Args[0])
	if name == "command" {
		return commandTarget(call.Args[1:])
	}
	return name
}

// commandTarget resolves the actual binary from "command" arguments.
// "command -v foo" / "command -V foo" → "" (query only, not execution).
// "command foo args..." → "foo" (executes foo).
func commandTarget(args []*syntax.Word) string {
	for _, arg := range args {
		lit := extractWordLiteral(arg)
		if lit == "" {
			return ""
		}
		if !strings.HasPrefix(lit, "-") {
			return lit // first non-flag arg is the binary
		}
		if isCommandQueryFlag(lit) {
			return "" // -v/-V means this is a lookup, not execution
		}
	}
	return ""
}

// isCommandQueryFlag returns true if the flag makes "command" a query
// rather than an execution (i.e., -v or -V).
func isCommandQueryFlag(flag string) bool {
	return strings.ContainsAny(flag, "vV")
}

// extractWordLiteral returns the literal string of a simple Word,
// or "" if it contains expansions/substitutions.
func extractWordLiteral(w *syntax.Word) string {
	if len(w.Parts) != 1 {
		return ""
	}
	lit, ok := w.Parts[0].(*syntax.Lit)
	if !ok {
		return ""
	}
	return lit.Value
}

// shellBuiltins are commands handled by the shell, not found on PATH.
var shellBuiltins = map[string]bool{
	"echo": true, "printf": true, "test": true, "[": true,
	"if": true, "then": true, "else": true, "fi": true,
	"for": true, "while": true, "do": true, "done": true,
	"case": true, "esac": true, "read": true, "eval": true,
	"exec": true, "exit": true, "return": true, "shift": true,
	"trap": true, "wait": true, "true": true, "false": true,
	"source": true, ".": true, "local": true, "declare": true,
	"set": true, "cd": true, "export": true, "unset": true,
}

// preambleCommands are external setup binaries skipped when finding
// the primary tool binary. Matches documented DIP125 behavior.
var preambleCommands = map[string]bool{
	"mkdir": true,
}

// isShellBuiltin returns true if the command is a shell builtin.
func isShellBuiltin(cmd string) bool {
	return shellBuiltins[cmd]
}

// isSkippableCommand returns true if the command should be skipped
// when looking for the primary tool binary (builtins + preamble).
func isSkippableCommand(cmd string) bool {
	return shellBuiltins[cmd] || preambleCommands[cmd]
}

// strQuote wraps a string in double quotes for diagnostic messages.
func strQuote(s string) string {
	return "\"" + s + "\""
}
