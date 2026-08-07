// ABOUTME: Lint passes for the workflow-level inputs block (DIP155-DIP157).
// ABOUTME: dippin lints the declaration; the engine validates supplied values.
package validator

import (
	"fmt"
	"strings"

	"github.com/2389-research/dippin-lang/ir"
)

// knownInputTypes is the v1 closed set of input types. A type outside this set
// is carried through the parser and IR verbatim and diagnosed here — that keeps
// a .dip using a future type parseable, formattable and packable on an older
// dippin, so only the lint complains. See issue #190.
var knownInputTypes = map[string]bool{
	"text":   true,
	"number": true,
	"bool":   true,
	"enum":   true,
	"file":   true,
	"secret": true,
}

// lintUnknownInputType checks DIP155: an input's declared type must be one this
// dippin recognizes.
func lintUnknownInputType(w *ir.Workflow) []Diagnostic {
	var diags []Diagnostic
	for _, in := range w.Inputs {
		if knownInputTypes[in.Type] {
			continue
		}
		diags = append(diags, Diagnostic{
			Code:     DIP155,
			Severity: SeverityError,
			Message:  fmt.Sprintf("input %q declares unrecognized type %q", in.Name, in.Type),
			Location: in.Source,
			Help:     "valid types are text, number, bool, enum, file, secret — or upgrade dippin if this type is newer than this build",
		})
	}
	return diags
}

// inputsPrefix is the namespace prefix for declared-input references.
const inputsPrefix = "inputs."

// lintUndeclaredInputRef checks DIP156: every ${inputs.x} in a prompt and every
// bare inputs.x in an edge condition must resolve to a declared input. inputs is
// the only closed namespace in the language — ctx is open, so a typo there is
// undetectable, which is precisely why caller input does not live in ctx.
func lintUndeclaredInputRef(w *ir.Workflow) []Diagnostic {
	var diags []Diagnostic
	diags = append(diags, undeclaredInputRefsInPrompts(w)...)
	diags = append(diags, undeclaredInputRefsInConditions(w)...)
	return diags
}

// undeclaredInputRefsInPrompts scans ${inputs.x} references in node prompts.
func undeclaredInputRefsInPrompts(w *ir.Workflow) []Diagnostic {
	var diags []Diagnostic
	for _, n := range w.Nodes {
		if prompt := nodePrompt(n); prompt != "" {
			diags = append(diags, undeclaredInputRefsInNodePrompt(w, n, prompt)...)
		}
	}
	return diags
}

// undeclaredInputRefsInNodePrompt scans a single node's prompt text for
// undeclared ${inputs.x} references.
func undeclaredInputRefsInNodePrompt(w *ir.Workflow, n *ir.Node, prompt string) []Diagnostic {
	var diags []Diagnostic
	for _, m := range varRefPattern.FindAllStringSubmatch(prompt, -1) {
		name, ok := undeclaredInputName(w, m[1])
		if !ok {
			continue
		}
		diags = append(diags, Diagnostic{
			Code:     DIP156,
			Severity: SeverityError,
			Message:  fmt.Sprintf("node %q references undeclared input ${inputs.%s}", n.ID, name),
			Location: n.Source,
			Help:     "declare it in the workflow's inputs block, or correct the name",
		})
	}
	return diags
}

// undeclaredInputRefsInConditions scans bare inputs.x variables in edge
// conditions. Conditions reference variables without ${} — see docs/context.md.
func undeclaredInputRefsInConditions(w *ir.Workflow) []Diagnostic {
	var diags []Diagnostic
	for _, e := range w.Edges {
		if e.Condition == nil || e.Condition.Parsed == nil {
			continue
		}
		diags = append(diags, undeclaredInputRefsInEdge(w, e)...)
	}
	return diags
}

// undeclaredInputRefsInEdge scans a single edge's parsed condition for
// undeclared inputs.x references.
func undeclaredInputRefsInEdge(w *ir.Workflow, e *ir.Edge) []Diagnostic {
	var diags []Diagnostic
	for _, cmp := range extractComparisons(e.Condition.Parsed) {
		name, ok := undeclaredInputName(w, cmp.Variable)
		if !ok {
			continue
		}
		diags = append(diags, Diagnostic{
			Code:     DIP156,
			Severity: SeverityError,
			Message:  fmt.Sprintf("edge %s → %s references undeclared input %q", e.From, e.To, "inputs."+name),
			Location: e.Source,
			Help:     "declare it in the workflow's inputs block, or correct the name",
		})
	}
	return diags
}

// undeclaredInputName returns the input name from an inputs.-prefixed reference
// when that input is not declared. ok is false for any other reference.
func undeclaredInputName(w *ir.Workflow, ref string) (string, bool) {
	if !strings.HasPrefix(ref, inputsPrefix) {
		return "", false
	}
	name := strings.TrimPrefix(ref, inputsPrefix)
	if name == "" || w.Input(name) != nil {
		return "", false
	}
	return name, true
}

// lintInputInToolCommand checks DIP157: an ${inputs.x} reference inside a tool
// node's command body never interpolates. The runtime keeps the whole inputs.
// namespace off its shell-interpolation allowlist (the same mechanism that
// blocks LLM-origin ctx.* keys from reaching a shell), so the reference is dead
// text that expands to nothing — silently, for every input type. See #190.
func lintInputInToolCommand(w *ir.Workflow) []Diagnostic {
	var diags []Diagnostic
	for _, n := range w.Nodes {
		cfg, ok := n.Config.(ir.ToolConfig)
		if !ok || cfg.Command == "" {
			continue
		}
		diags = append(diags, inputRefsInCommand(w, n, cfg.Command)...)
	}
	return diags
}

// inputRefsInCommand reports every inputs.-namespaced reference in one command body.
func inputRefsInCommand(w *ir.Workflow, n *ir.Node, command string) []Diagnostic {
	var diags []Diagnostic
	for _, m := range varRefPattern.FindAllStringSubmatch(command, -1) {
		if !strings.HasPrefix(m[1], inputsPrefix) {
			continue
		}
		name := strings.TrimPrefix(m[1], inputsPrefix)
		diags = append(diags, Diagnostic{
			Code:     DIP157,
			Severity: SeverityError,
			Message:  fmt.Sprintf("tool %q references ${inputs.%s}, which never interpolates in a command", n.ID, name),
			Location: n.Source,
			Help:     inputInCommandHelp(w, name),
		})
	}
	return diags
}

// inputInCommandHelp tailors the fix hint, sharpening it for a secret.
func inputInCommandHelp(w *ir.Workflow, name string) string {
	if in := w.Input(name); in != nil && in.Type == "secret" {
		return "the runtime never expands inputs into a shell — a secret least of all; pass it through the runtime's credential mechanism instead"
	}
	return "the runtime keeps the inputs namespace off its shell allowlist; route the value through an agent or a declared context key instead"
}
