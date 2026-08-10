// ABOUTME: Lint passes for the workflow-level inputs block (DIP155-DIP159).
// ABOUTME: dippin lints the declaration; the engine validates supplied values.
package validator

import (
	"fmt"
	"regexp"
	"strconv"
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

// --- DIP158: invalid or inapplicable input constraint ---

// inputConstraintTypes maps each constraint attribute to the input types it
// applies to. A constraint set on any other type is inapplicable (DIP158).
var inputConstraintTypes = map[string][]string{
	"options":    {"enum"},
	"pattern":    {"text", "secret"},
	"max_length": {"text", "secret"},
	"multiline":  {"text", "secret"},
	"min":        {"number"},
	"max":        {"number"},
}

// lintInputConstraints checks DIP158: an input's constraints must be valid and
// applicable to its type — enum default in options, min ≤ max, a well-formed
// pattern, and no constraint on a type that has no such constraint.
func lintInputConstraints(w *ir.Workflow) []Diagnostic {
	var diags []Diagnostic
	for _, in := range w.Inputs {
		diags = append(diags, checkEnumDefault(in)...)
		diags = append(diags, checkMinMax(in)...)
		diags = append(diags, checkPattern(in)...)
		diags = append(diags, checkInapplicableConstraints(in)...)
	}
	return diags
}

// checkEnumDefault flags an enum whose default is not one of its options.
func checkEnumDefault(in *ir.Input) []Diagnostic {
	if in.Type != "enum" || !in.HasDefault {
		return nil
	}
	for _, o := range in.Options {
		if o == in.Default {
			return nil
		}
	}
	return []Diagnostic{inputConstraintDiag(in, fmt.Sprintf("enum default %q is not one of its options", in.Default))}
}

// checkMinMax flags a number input whose min exceeds its max.
func checkMinMax(in *ir.Input) []Diagnostic {
	if in.Type != "number" {
		return nil
	}
	lo, hi, ok := parseBounds(in)
	if !ok || lo <= hi {
		return nil
	}
	return []Diagnostic{inputConstraintDiag(in, fmt.Sprintf("min (%s) is greater than max (%s)", in.Min, in.Max))}
}

// parseBounds parses both bounds as floats; ok is false unless both are present
// and numeric.
func parseBounds(in *ir.Input) (lo, hi float64, ok bool) {
	if in.Min == "" || in.Max == "" {
		return 0, 0, false
	}
	l, e1 := strconv.ParseFloat(in.Min, 64)
	h, e2 := strconv.ParseFloat(in.Max, 64)
	if e1 != nil || e2 != nil {
		return 0, 0, false
	}
	return l, h, true
}

// checkPattern flags a malformed pattern regex.
func checkPattern(in *ir.Input) []Diagnostic {
	if in.Pattern == "" {
		return nil
	}
	if _, err := regexp.Compile(in.Pattern); err != nil {
		return []Diagnostic{inputConstraintDiag(in, fmt.Sprintf("pattern %q is not a valid regular expression", in.Pattern))}
	}
	return nil
}

// checkInapplicableConstraints flags each set constraint that does not apply to
// the input's type (e.g. max_length on a bool, options on a text).
func checkInapplicableConstraints(in *ir.Input) []Diagnostic {
	var out []Diagnostic
	for _, c := range setConstraints(in) {
		if !typeAllowsConstraint(in.Type, c) {
			out = append(out, inputConstraintDiag(in,
				fmt.Sprintf("%s does not apply to a %q input", c, in.Type)))
		}
	}
	return out
}

// setConstraints returns the names of the constraints this input actually sets.
func setConstraints(in *ir.Input) []string {
	checks := []struct {
		name string
		set  bool
	}{
		{"options", len(in.Options) > 0},
		{"pattern", in.Pattern != ""},
		{"max_length", in.MaxLength > 0},
		{"multiline", in.Multiline},
		{"min", in.Min != ""},
		{"max", in.Max != ""},
	}
	var set []string
	for _, c := range checks {
		if c.set {
			set = append(set, c.name)
		}
	}
	return set
}

// typeAllowsConstraint reports whether constraint applies to typ.
func typeAllowsConstraint(typ, constraint string) bool {
	for _, t := range inputConstraintTypes[constraint] {
		if t == typ {
			return true
		}
	}
	return false
}

// inputConstraintDiag builds a DIP158 error for one input.
func inputConstraintDiag(in *ir.Input, detail string) Diagnostic {
	return Diagnostic{
		Code:     DIP158,
		Severity: SeverityError,
		Message:  fmt.Sprintf("input %q: %s", in.Name, detail),
		Location: in.Source,
		Help:     "correct the constraint, or move it to an input whose type supports it",
	}
}

// --- DIP159: declared input never referenced (dead input) ---

// lintDeadInputs checks DIP159: a declared input that no prompt, edge condition,
// or tool command references. Advisory (a host may still collect it), mirroring
// DIP107 for dead node outputs.
func lintDeadInputs(w *ir.Workflow) []Diagnostic {
	if len(w.Inputs) == 0 {
		return nil
	}
	referenced := referencedInputNames(w)
	var diags []Diagnostic
	for _, in := range w.Inputs {
		if referenced[in.Name] {
			continue
		}
		diags = append(diags, Diagnostic{
			Code:     DIP159,
			Severity: SeverityWarning,
			Message:  fmt.Sprintf("declared input %q is never referenced", in.Name),
			Location: in.Source,
			Help:     fmt.Sprintf("reference it as ${inputs.%s} in a prompt or edge condition, or remove the declaration", in.Name),
		})
	}
	return diags
}

// referencedInputNames collects every input name mentioned as ${inputs.x} in a
// prompt or tool command, or as inputs.x in an edge condition. A tool-command
// mention counts (the author clearly intends the input; DIP157 separately warns
// it won't interpolate there) so DIP157 and DIP159 don't double-flag it.
func referencedInputNames(w *ir.Workflow) map[string]bool {
	seen := map[string]bool{}
	for _, n := range w.Nodes {
		collectInputRefs(seen, nodePrompt(n))
		if cfg, ok := n.Config.(ir.ToolConfig); ok {
			collectInputRefs(seen, cfg.Command)
		}
	}
	for _, e := range w.Edges {
		collectConditionInputRefs(seen, e)
	}
	return seen
}

// collectInputRefs adds every ${inputs.x} name found in text to seen.
func collectInputRefs(seen map[string]bool, text string) {
	if text == "" {
		return
	}
	for _, m := range varRefPattern.FindAllStringSubmatch(text, -1) {
		if name, ok := inputRefName(m[1]); ok {
			seen[name] = true
		}
	}
}

// collectConditionInputRefs adds every inputs.x name found in an edge condition.
func collectConditionInputRefs(seen map[string]bool, e *ir.Edge) {
	if e.Condition == nil || e.Condition.Parsed == nil {
		return
	}
	for _, cmp := range extractComparisons(e.Condition.Parsed) {
		if name, ok := inputRefName(cmp.Variable); ok {
			seen[name] = true
		}
	}
}

// inputRefName returns the input name from an inputs.-prefixed reference.
func inputRefName(ref string) (string, bool) {
	if !strings.HasPrefix(ref, inputsPrefix) {
		return "", false
	}
	name := strings.TrimPrefix(ref, inputsPrefix)
	if name == "" {
		return "", false
	}
	return name, true
}
