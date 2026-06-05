package validator

import (
	"fmt"

	"github.com/2389-research/dippin-lang/ir"
)

// lintBudgetRanges checks DIP145: a graph-level budget default must not be
// negative. Zero means "unset / no limit" (the formatter emits only non-zero
// values), so only values < 0 are flagged. Integer fields use %d; duration
// fields use time.Duration.String() so messages are human-readable.
func lintBudgetRanges(w *ir.Workflow) []Diagnostic {
	d := w.Defaults
	checks := []struct {
		name string
		neg  bool
		val  string
	}{
		{"max_total_tokens", d.MaxTotalTokens < 0, fmt.Sprintf("%d", d.MaxTotalTokens)},
		{"max_cost_cents", d.MaxCostCents < 0, fmt.Sprintf("%d", d.MaxCostCents)},
		{"max_wall_time", d.MaxWallTime < 0, d.MaxWallTime.String()},
		{"stall_timeout", d.StallTimeout < 0, d.StallTimeout.String()},
	}
	var diags []Diagnostic
	for _, c := range checks {
		if c.neg {
			diags = append(diags, Diagnostic{
				Code:     DIP145,
				Severity: SeverityWarning,
				Message:  fmt.Sprintf("workflow budget default %s is %s; budgets cannot be negative", c.name, c.val),
				Help:     "use a positive cap (e.g. max_cost_cents: 1000 for $10.00), or omit it / set 0 for no limit",
			})
		}
	}
	return diags
}
