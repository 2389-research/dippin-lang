package validator

import (
	"testing"

	"github.com/2389-research/dippin-lang/ir"
)

// --- checkSelectorRef: class and ID selector validation ---

func TestCheckSelectorRef_MissingClass(t *testing.T) {
	rule := ir.StylesheetRule{
		Selector:   ir.StyleSelector{Kind: "class", Value: "nonexistent"},
		Properties: map[string]string{"model": "o1"},
	}
	classes := map[string]bool{"coder": true}
	nodeIDs := map[string]bool{"A": true}

	diags := checkSelectorRef(rule, classes, nodeIDs)
	if len(diags) != 1 {
		t.Fatalf("expected 1 diagnostic, got %d", len(diags))
	}
	if diags[0].Code != DIP117 {
		t.Errorf("code = %q, want DIP117", diags[0].Code)
	}
}

func TestCheckSelectorRef_ValidClass(t *testing.T) {
	rule := ir.StylesheetRule{
		Selector:   ir.StyleSelector{Kind: "class", Value: "coder"},
		Properties: map[string]string{"model": "o1"},
	}
	classes := map[string]bool{"coder": true}
	nodeIDs := map[string]bool{"A": true}

	diags := checkSelectorRef(rule, classes, nodeIDs)
	if len(diags) != 0 {
		t.Errorf("expected 0 diagnostics, got %d", len(diags))
	}
}

func TestCheckSelectorRef_MissingNodeID(t *testing.T) {
	rule := ir.StylesheetRule{
		Selector:   ir.StyleSelector{Kind: "id", Value: "Ghost"},
		Properties: map[string]string{"model": "o1"},
	}
	classes := map[string]bool{}
	nodeIDs := map[string]bool{"A": true}

	diags := checkSelectorRef(rule, classes, nodeIDs)
	if len(diags) != 1 {
		t.Fatalf("expected 1 diagnostic, got %d", len(diags))
	}
	if diags[0].Code != DIP118 {
		t.Errorf("code = %q, want DIP118", diags[0].Code)
	}
}

func TestCheckSelectorRef_ValidNodeID(t *testing.T) {
	rule := ir.StylesheetRule{
		Selector:   ir.StyleSelector{Kind: "id", Value: "A"},
		Properties: map[string]string{"model": "o1"},
	}
	classes := map[string]bool{}
	nodeIDs := map[string]bool{"A": true}

	diags := checkSelectorRef(rule, classes, nodeIDs)
	if len(diags) != 0 {
		t.Errorf("expected 0 diagnostics, got %d", len(diags))
	}
}

func TestCheckSelectorRef_UnknownKind(t *testing.T) {
	rule := ir.StylesheetRule{
		Selector:   ir.StyleSelector{Kind: "universal", Value: "*"},
		Properties: map[string]string{"model": "o1"},
	}
	diags := checkSelectorRef(rule, map[string]bool{}, map[string]bool{})
	if len(diags) != 0 {
		t.Errorf("expected 0 diagnostics for universal selector, got %d", len(diags))
	}
}

// --- lintOnResume ---

func TestLintOnResume_InvalidValue(t *testing.T) {
	w := &ir.Workflow{
		Defaults: ir.WorkflowDefaults{
			OnResume: "bogus",
			Fidelity: "full",
		},
	}
	diags := lintOnResume(w)
	if len(diags) != 1 {
		t.Fatalf("expected 1 diagnostic, got %d", len(diags))
	}
	if diags[0].Code != DIP116 {
		t.Errorf("code = %q, want DIP116", diags[0].Code)
	}
}

func TestLintOnResume_WithoutFidelity(t *testing.T) {
	w := &ir.Workflow{
		Defaults: ir.WorkflowDefaults{
			OnResume: "preserve",
			Fidelity: "", // no fidelity
		},
	}
	diags := lintOnResume(w)
	if len(diags) != 1 {
		t.Fatalf("expected 1 diagnostic (on_resume without fidelity), got %d", len(diags))
	}
}

func TestLintOnResume_InvalidValueAndNoFidelity(t *testing.T) {
	w := &ir.Workflow{
		Defaults: ir.WorkflowDefaults{
			OnResume: "bogus",
			Fidelity: "",
		},
	}
	diags := lintOnResume(w)
	if len(diags) != 2 {
		t.Fatalf("expected 2 diagnostics, got %d", len(diags))
	}
}

func TestLintOnResume_EmptyOnResume(t *testing.T) {
	w := &ir.Workflow{
		Defaults: ir.WorkflowDefaults{OnResume: ""},
	}
	diags := lintOnResume(w)
	if len(diags) != 0 {
		t.Errorf("expected 0 diagnostics for empty on_resume, got %d", len(diags))
	}
}

func TestLintOnResume_ValidPreserve(t *testing.T) {
	w := &ir.Workflow{
		Defaults: ir.WorkflowDefaults{
			OnResume: "preserve",
			Fidelity: "full",
		},
	}
	diags := lintOnResume(w)
	if len(diags) != 0 {
		t.Errorf("expected 0 diagnostics for valid on_resume, got %d", len(diags))
	}
}

// --- lintStylesheetRefs integration ---

func TestLintStylesheetRefs_Integration(t *testing.T) {
	w := &ir.Workflow{
		Nodes: []*ir.Node{
			{ID: "A", Kind: ir.NodeAgent, Classes: []string{"coder"}, Config: ir.AgentConfig{Prompt: "test"}},
		},
		Stylesheet: []ir.StylesheetRule{
			{Selector: ir.StyleSelector{Kind: "class", Value: "missing"}, Properties: map[string]string{}},
			{Selector: ir.StyleSelector{Kind: "id", Value: "Ghost"}, Properties: map[string]string{}},
			{Selector: ir.StyleSelector{Kind: "class", Value: "coder"}, Properties: map[string]string{}},
			{Selector: ir.StyleSelector{Kind: "id", Value: "A"}, Properties: map[string]string{}},
		},
	}
	diags := lintStylesheetRefs(w)
	// Should get 2 diagnostics: missing class + missing ID.
	if len(diags) != 2 {
		t.Fatalf("expected 2 diagnostics, got %d", len(diags))
	}
}

// --- collectClassesAndIDs ---

func TestCollectClassesAndIDs_Empty(t *testing.T) {
	w := &ir.Workflow{}
	classes, ids := collectClassesAndIDs(w)
	if len(classes) != 0 || len(ids) != 0 {
		t.Error("expected empty sets for empty workflow")
	}
}

// --- extractSignedComparisons edge cases ---

func TestExtractSignedComparisons_NilExpr(t *testing.T) {
	// nil expression should return nil without panic.
	result := extractSignedComparisons(nil, false)
	if len(result) != 0 {
		t.Errorf("expected 0 comparisons for nil expr, got %d", len(result))
	}
}

func TestExtractSignedComparisons_Negated(t *testing.T) {
	expr := ir.CondNot{Inner: ir.CondCompare{Variable: "ctx.x", Op: "=", Value: "y"}}
	result := extractSignedComparisons(expr, false)
	if len(result) != 1 {
		t.Fatalf("expected 1 comparison, got %d", len(result))
	}
	if !result[0].negated {
		t.Error("expected negated comparison inside CondNot")
	}
}

func TestExtractSignedComparisons_OrBranches(t *testing.T) {
	expr := ir.CondOr{
		Left:  ir.CondCompare{Variable: "ctx.x", Op: "=", Value: "a"},
		Right: ir.CondCompare{Variable: "ctx.y", Op: "=", Value: "b"},
	}
	result := extractSignedComparisons(expr, false)
	if len(result) != 2 {
		t.Fatalf("expected 2 comparisons, got %d", len(result))
	}
}

// --- stripNamespace ---

func TestStripNamespace(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"ctx.outcome", "outcome"},
		{"outcome", "outcome"},
		{"graph.goal", "goal"},
	}
	for _, tt := range tests {
		got := stripNamespace(tt.input)
		if got != tt.want {
			t.Errorf("stripNamespace(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
