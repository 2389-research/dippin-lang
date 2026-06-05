package validator

import (
	"strings"
	"testing"
	"time"

	"github.com/2389-research/dippin-lang/ir"
)

func budgetTestWorkflow(d ir.WorkflowDefaults) *ir.Workflow {
	return &ir.Workflow{
		Name: "B", Start: "A", Exit: "A", Defaults: d,
		Nodes: []*ir.Node{{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "x"}}},
		Edges: []*ir.Edge{{From: "A", To: "A"}},
	}
}

func countDiagCode(diags []Diagnostic, code string) int {
	n := 0
	for _, d := range diags {
		if d.Code == code {
			n++
		}
	}
	return n
}

func TestDIP145FiresOnNegativeBudgets(t *testing.T) {
	w := budgetTestWorkflow(ir.WorkflowDefaults{
		MaxTotalTokens: -1,
		MaxCostCents:   -5,
		MaxWallTime:    -30 * time.Second,
		StallTimeout:   -5 * time.Minute,
	})
	got := countDiagCode(lintBudgetRanges(w), DIP145)
	if got != 4 {
		t.Errorf("DIP145 count = %d, want 4 (one per negative field)", got)
	}
}

func TestDIP145SilentOnUnsetZero(t *testing.T) {
	w := budgetTestWorkflow(ir.WorkflowDefaults{})
	if got := countDiagCode(lintBudgetRanges(w), DIP145); got != 0 {
		t.Errorf("DIP145 fired on unset (0) budgets: %d, want 0", got)
	}
}

func TestDIP145SilentOnValidPositive(t *testing.T) {
	w := budgetTestWorkflow(ir.WorkflowDefaults{
		MaxTotalTokens: 500000,
		MaxCostCents:   1000,
		MaxWallTime:    30 * time.Minute,
		StallTimeout:   5 * time.Minute,
	})
	if got := countDiagCode(lintBudgetRanges(w), DIP145); got != 0 {
		t.Errorf("DIP145 fired on valid positive budgets: %d, want 0", got)
	}
}

func TestDIP145MessageNamesFieldAndValue(t *testing.T) {
	w := budgetTestWorkflow(ir.WorkflowDefaults{MaxCostCents: -5})
	diags := lintBudgetRanges(w)
	if len(diags) != 1 {
		t.Fatalf("want 1 diag, got %d", len(diags))
	}
	msg := diags[0].Message
	if !strings.Contains(msg, "max_cost_cents") || !strings.Contains(msg, "-5") {
		t.Errorf("message must name field and value, got: %q", msg)
	}
}

func TestDIP145DurationMessageIsReadable(t *testing.T) {
	w := budgetTestWorkflow(ir.WorkflowDefaults{StallTimeout: -5 * time.Minute})
	diags := lintBudgetRanges(w)
	if len(diags) != 1 {
		t.Fatalf("want 1 diag, got %d", len(diags))
	}
	msg := diags[0].Message
	if !strings.Contains(msg, "-5m") {
		t.Errorf("duration message must be human-readable (contain -5m), got: %q", msg)
	}
	if strings.Contains(msg, "300000000000") {
		t.Errorf("duration message must NOT show raw nanoseconds, got: %q", msg)
	}
}
