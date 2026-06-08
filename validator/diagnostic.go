// Package validator performs graph structure validation on Dippin IR workflows.
//
// It implements checks DIP001 through DIP010, covering structural correctness
// of the workflow graph: start/exit existence, edge validity, reachability,
// cycle detection, parallel/fan-in pairing, duplicate detection, and edge
// condition parseability.
//
// The validator takes a *ir.Workflow and returns a Result containing all
// diagnostics found. It always runs all checks and never short-circuits, so a
// single pass reports everything. Validate has one documented side effect: it
// lazily populates edge Condition.Parsed for parseable conditions (idempotent),
// so callers should not assume the workflow is left unmutated.
package validator

import (
	"fmt"
	"strings"

	"github.com/2389-research/dippin-lang/ir"
)

// Severity levels for diagnostics.
type Severity int

const (
	SeverityError   Severity = iota // Must fix — workflow cannot execute
	SeverityWarning                 // Should fix — likely a bug (used by linter, not this component)
	SeverityInfo                    // Informational
	SeverityHint                    // Suggestion
)

// String returns a human-readable severity label.
func (s Severity) String() string {
	switch s {
	case SeverityError:
		return "error"
	case SeverityWarning:
		return "warning"
	case SeverityInfo:
		return "info"
	case SeverityHint:
		return "hint"
	default:
		return "unknown"
	}
}

// Diagnostic represents a single validation finding.
type Diagnostic struct {
	Code     string            // "DIP001", "DIP002", etc.
	Severity Severity          // Error, warning, etc.
	Message  string            // Human-readable explanation
	Location ir.SourceLocation // Where in the source (may be zero-value if unavailable)
	Help     string            // Optional "did you mean X?" or explanation
	Fix      string            // Optional suggested replacement text
}

// String returns a formatted diagnostic string matching the spec format:
//
//	error[DIP003]: unknown node reference "InterpretX" in edge
//	  --> pipeline.dip:45:5
//	  = help: did you mean "Interpret"?
func (d Diagnostic) String() string {
	var b strings.Builder
	fmt.Fprintf(&b, "%s[%s]: %s", d.Severity, d.Code, d.Message)

	file := d.Location.File
	if file == "" {
		file = "<unknown>"
	}
	fmt.Fprintf(&b, "\n  --> %s:%d:%d", file, d.Location.Line, d.Location.Column)

	if d.Help != "" {
		fmt.Fprintf(&b, "\n  = help: %s", d.Help)
	}
	if d.Fix != "" {
		fmt.Fprintf(&b, "\n  = fix: %s", d.Fix)
	}
	return b.String()
}

// Result holds the outcome of a validation pass.
type Result struct {
	Diagnostics []Diagnostic
}

// Errors returns only error-severity diagnostics.
func (r Result) Errors() []Diagnostic {
	var out []Diagnostic
	for _, d := range r.Diagnostics {
		if d.Severity == SeverityError {
			out = append(out, d)
		}
	}
	return out
}

// HasErrors returns true if any error-severity diagnostics exist.
func (r Result) HasErrors() bool {
	for _, d := range r.Diagnostics {
		if d.Severity == SeverityError {
			return true
		}
	}
	return false
}
