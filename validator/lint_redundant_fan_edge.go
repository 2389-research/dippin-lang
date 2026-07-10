// ABOUTME: DIP153 — an edges-block edge redundantly repeats a parallel fan-out
// or fan_in join already declared inline on a node's config. The inline list is
// the single source of truth; such an unconditional, attribute-free edge conveys
// nothing new and is stripped by `dippin fmt` (rejected under `dip 2`).
package validator

import (
	"fmt"

	"github.com/2389-research/dippin-lang/ir"
)

// lintRedundantFanEdge flags edges that duplicate an inline parallel/fan_in fork
// (DIP153). Detection delegates to ir.IsRedundantFanEdge so the lint, the
// formatter strip, and the dip-2 parser rejection share one definition.
func lintRedundantFanEdge(w *ir.Workflow) []Diagnostic {
	var diags []Diagnostic
	for _, e := range w.Edges {
		if !ir.IsRedundantFanEdge(w, e) {
			continue
		}
		diags = append(diags, Diagnostic{
			Code:     DIP153,
			Severity: SeverityWarning,
			Message: fmt.Sprintf(
				"edges-block edge '%s -> %s' redundantly repeats the inline parallel/fan_in fork; the inline list is authoritative — run 'dippin fmt' to remove it (rejected under 'dip 2')",
				e.From, e.To),
			Location: e.Source,
		})
	}
	return diags
}
