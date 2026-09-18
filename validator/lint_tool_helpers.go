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

// placeholderDummy replaces every dippin namespace placeholder before shell
// parsing. Dippin placeholders like ${graph.workflow_dir} or ${params.foo}
// contain a "." inside the braces, which is not valid shell
// parameter-expansion syntax and causes the mvdan shell parser to fail,
// misdirecting extraction to the fallback word-splitter. Substituting a
// shell-safe dummy word lets the parser succeed so the normal
// skip-set/assignment extraction logic runs.
const placeholderDummy = "__dip_placeholder__"

// placeholderPattern matches only dippin namespace references of the form
// ${ident.ident[.ident...]} (${graph.workflow_dir}, ${params.foo},
// ${ctx.x}, ${ctx.node.id.key}, ...) — the shape that's invalid shell
// syntax and needs the placeholderDummy substitution below. This is
// deliberately narrower than the package-wide varRefPattern (lint_context.go),
// which matches ANY ${...} body and is used elsewhere (DIP106) to validate
// prompt/command variable references generally. Rewriting every ${...} here
// would also clobber valid shell expansions that happen to appear in a
// command body — ${VAR}, ${VAR:-default}, ${#VAR} — which parse fine as-is
// and must reach the AST walk unchanged so it can still see past them (e.g.
// to a command substitution like ${CACHE:-$(helper)}).
var placeholderPattern = regexp.MustCompile(`\$\{[A-Za-z_][A-Za-z0-9_]*(?:\.[A-Za-z0-9_-]+)+\}`)

// extractBinary parses a shell command and returns the first non-builtin,
// non-preamble command name. Uses a proper shell AST parser to correctly
// handle variable assignments, pipes, subshells, command substitution, etc.
// Shell builtins and preamble commands (mkdir) are skipped to find the
// primary tool binary. Falls back to token-based extraction on parse errors.
// Dippin ${ns.key} placeholders are substituted with a dummy word first so
// they don't break the shell parse (see placeholderDummy/placeholderPattern);
// if the extracted binary name still contains the dummy (whether it *is*
// the placeholder or the placeholder was concatenated with literal text,
// e.g. "${p.x}suffix"), extraction returns "" since the real binary can't
// be known before expansion.
//
// Two additional cases return "" because the symbol space becomes
// unknowable: (1) if a "." or "source" command appears, in source order,
// before the first non-builtin command in the body, the script may be
// loading functions from disk that the walk has no way to resolve — DIP125
// is skipped for that node (a "."/"source" appearing *after* the first real
// command does not suppress it — that first command is still checkable,
// and a "." inside the body of an as-yet-uncalled function still counts as
// "before" if it's reached first in source order); (2) if the candidate
// binary name matches a function defined by a FuncDecl anywhere in the
// body, it's a shell function, not a PATH binary. Note that
// syntax.Walk returning false only prunes that node's children — it does
// not stop the walk at siblings, so the walk may well continue past a "."
// and still set bin to some later command. The "" this doc promises comes
// from extractBinary's `sawSource ||` override below, not from the walk
// itself; that flag, once set, is the load-bearing invariant.
func extractBinary(command string) string {
	sanitized := placeholderPattern.ReplaceAllString(command, placeholderDummy)
	bin, sawSource := parseBinary(sanitized)
	if sawSource || strings.Contains(bin, placeholderDummy) {
		return ""
	}
	return bin
}

// parseBinary shell-parses sanitized and returns the candidate binary name
// plus whether a "."/"source" command was reached before it (see
// extractBinary's doc comment). Falls back to token-based extraction on
// parse errors.
func parseBinary(sanitized string) (string, bool) {
	parser := syntax.NewParser(syntax.KeepComments(false))
	prog, err := parser.Parse(strings.NewReader(sanitized), "")
	if err != nil {
		return extractBinaryFallback(sanitized), false
	}
	var bin string
	var sawSource bool
	syntax.Walk(prog, func(node syntax.Node) bool {
		return visitForBinary(node, &bin, &sawSource)
	})
	if bin != "" && bodyDefinesFunc(prog, bin) {
		bin = ""
	}
	return bin, sawSource
}

// bodyDefinesFunc reports whether the parsed command body declares a shell
// function named name (a FuncDecl), meaning name is not a PATH binary.
func bodyDefinesFunc(prog *syntax.File, name string) bool {
	found := false
	syntax.Walk(prog, func(node syntax.Node) bool {
		if found {
			return false
		}
		if fd, ok := node.(*syntax.FuncDecl); ok && fd.Name != nil && fd.Name.Value == name {
			found = true
			return false
		}
		return true
	})
	return found
}

// extractBinaryFallback performs best-effort extraction when shell parsing
// fails. Skips builtins and preamble, returns the first plausible binary. A
// "."/"source" token, wherever it falls, makes the symbol space unknowable
// and short-circuits to "" — mirroring the AST path's sawSource override —
// so a malformed script that sources a lib before its real command can't
// false-positive on a function name defined by that lib.
func extractBinaryFallback(command string) string {
	for _, field := range strings.Fields(command) {
		if isSourceCommand(field) {
			return ""
		}
		if !isSkippableCommand(field) {
			return field
		}
	}
	return ""
}

// visitForBinary is a single syntax.Walk step that captures the first
// non-builtin, non-preamble command binary into bin. If a "." or "source"
// command is reached (in walk order) before any such binary, it sets
// sawSource — but returning false here only prunes that CallExpr's own
// children, it does NOT stop the walk at later siblings, so bin may still
// end up set by a command that textually follows the "."/"source". The
// actual "" result for that case comes from parseBinary/extractBinary's
// `sawSource ||` override, not from this function refusing to set bin. It's
// pulled out as a plain function (rather than a closure returned from a
// wrapper) so its branching doesn't stack on top of an enclosing closure's
// nesting level for cognitive-complexity purposes.
func visitForBinary(node syntax.Node, bin *string, sawSource *bool) bool {
	if *bin != "" {
		return false
	}
	name := callExprBinary(node)
	if name == "" {
		return true
	}
	if isSourceCommand(name) {
		*sawSource = true
		return false
	}
	if isSkippableCommand(name) {
		return true
	}
	*bin = name
	return false
}

// isSourceCommand reports whether name loads another file's definitions
// into the current shell ("." or "source"), which makes the symbol space
// unknowable to a static walk.
func isSourceCommand(name string) bool {
	return name == "." || name == "source"
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
	"echo": true, "printf": true, "test": true, "[": true, "[[": true,
	"if": true, "then": true, "else": true, "fi": true,
	"for": true, "while": true, "do": true, "done": true,
	"case": true, "esac": true, "read": true, "eval": true,
	"exec": true, "exit": true, "return": true, "shift": true,
	"trap": true, "wait": true, "true": true, "false": true,
	"source": true, ".": true, ":": true, "local": true, "declare": true,
	"set": true, "cd": true, "export": true, "unset": true,
	"alias": true, "break": true, "continue": true, "getopts": true,
	"readonly": true, "times": true, "type": true, "ulimit": true,
	"umask": true, "hash": true, "pwd": true, "kill": true,
	"jobs": true, "fg": true, "bg": true, "let": true, "typeset": true,
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
