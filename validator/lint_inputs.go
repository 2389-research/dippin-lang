// ABOUTME: Lint passes for the workflow-level inputs block (DIP155-DIP157).
// ABOUTME: dippin lints the declaration; the engine validates supplied values.
package validator

import (
	"fmt"

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
