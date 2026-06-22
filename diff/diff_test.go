package diff_test

import (
	"os"
	"testing"

	"github.com/2389-research/dippin-lang/diff"
	"github.com/2389-research/dippin-lang/ir"
	"github.com/2389-research/dippin-lang/parser"
)

func loadFixture(t *testing.T, path string) *ir.Workflow {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read %s: %v", path, err)
	}
	p := parser.NewParser(string(data), path)
	w, err := p.Parse()
	if err != nil {
		t.Fatalf("failed to parse %s: %v", path, err)
	}
	return w
}

func TestCompare_V1toV2(t *testing.T) {
	v1 := loadFixture(t, "testdata/v1.dip")
	v2 := loadFixture(t, "testdata/v2.dip")

	report := diff.Compare(v1, v2)

	// Review was added in v2.
	if len(report.NodesAdded) != 1 || report.NodesAdded[0] != "Review" {
		t.Errorf("expected [Review] added, got %v", report.NodesAdded)
	}

	// No nodes removed.
	if len(report.NodesRemoved) != 0 {
		t.Errorf("expected no removals, got %v", report.NodesRemoved)
	}

	// Process node was modified (model and prompt changed).
	foundProcess := false
	for _, nd := range report.NodesModified {
		if nd.NodeID == "Process" {
			foundProcess = true
			if len(nd.Changes) == 0 {
				t.Error("expected field changes for Process node")
			}
		}
	}
	if !foundProcess {
		t.Error("expected Process node to appear in modifications")
	}
}

func TestCompare_EdgesChanged(t *testing.T) {
	v1 := loadFixture(t, "testdata/v1.dip")
	v2 := loadFixture(t, "testdata/v2.dip")

	report := diff.Compare(v1, v2)

	// Process -> Done was removed, Process -> Review and Review -> Done added.
	if len(report.EdgesAdded) < 1 {
		t.Errorf("expected edges added, got %v", report.EdgesAdded)
	}
	if len(report.EdgesRemoved) < 1 {
		t.Errorf("expected edges removed, got %v", report.EdgesRemoved)
	}
}

func TestCompare_CostDelta(t *testing.T) {
	v1 := loadFixture(t, "testdata/v1.dip")
	v2 := loadFixture(t, "testdata/v2.dip")

	report := diff.Compare(v1, v2)

	// v2 has an extra node and downgraded Process, cost should change.
	if report.CostDelta.OldCost.Expected == 0 && report.CostDelta.NewCost.Expected == 0 {
		t.Error("expected non-zero cost values")
	}
}

func TestCompare_Identical(t *testing.T) {
	v1 := loadFixture(t, "testdata/v1.dip")

	report := diff.Compare(v1, v1)

	if len(report.NodesAdded) != 0 {
		t.Errorf("expected no nodes added, got %v", report.NodesAdded)
	}
	if len(report.NodesRemoved) != 0 {
		t.Errorf("expected no nodes removed, got %v", report.NodesRemoved)
	}
	if len(report.NodesModified) != 0 {
		t.Errorf("expected no nodes modified, got %v", report.NodesModified)
	}
	if len(report.EdgesAdded) != 0 {
		t.Errorf("expected no edges added, got %v", report.EdgesAdded)
	}
	if len(report.EdgesRemoved) != 0 {
		t.Errorf("expected no edges removed, got %v", report.EdgesRemoved)
	}
}

func TestCompare_FieldChanges(t *testing.T) {
	v1 := loadFixture(t, "testdata/v1.dip")
	v2 := loadFixture(t, "testdata/v2.dip")

	report := diff.Compare(v1, v2)

	for _, nd := range report.NodesModified {
		for _, c := range nd.Changes {
			if c.Field == "" {
				t.Error("field name should not be empty")
			}
		}
	}
}

func TestCompare_NodeModifiedConfigChanges(t *testing.T) {
	old := &ir.Workflow{
		Name: "test", Start: "A", Exit: "B",
		Nodes: []*ir.Node{
			{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{
				Prompt: "Original prompt.", Model: "gpt-5.4", Provider: "openai",
				MaxTurns: 3, ReasoningEffort: "low", Fidelity: "summary",
			}},
			{ID: "B", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "Done."}},
		},
		Edges: []*ir.Edge{{From: "A", To: "B"}},
	}
	new := &ir.Workflow{
		Name: "test", Start: "A", Exit: "B",
		Nodes: []*ir.Node{
			{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{
				Prompt: "Changed prompt.", Model: "claude-opus-4-6", Provider: "anthropic",
				MaxTurns: 10, ReasoningEffort: "high", Fidelity: "full",
			}},
			{ID: "B", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "Done."}},
		},
		Edges: []*ir.Edge{{From: "A", To: "B"}},
	}

	report := diff.Compare(old, new)

	if len(report.NodesModified) != 1 {
		t.Fatalf("expected 1 modified node, got %d", len(report.NodesModified))
	}
	nd := report.NodesModified[0]
	if nd.NodeID != "A" {
		t.Errorf("expected modified node A, got %s", nd.NodeID)
	}

	fieldNames := map[string]bool{}
	for _, c := range nd.Changes {
		fieldNames[c.Field] = true
	}
	for _, f := range []string{"model", "provider", "prompt", "max_turns", "reasoning_effort", "fidelity"} {
		if !fieldNames[f] {
			t.Errorf("expected field change for %q", f)
		}
	}
}

func TestCompare_LabelChange(t *testing.T) {
	old := &ir.Workflow{
		Name: "test", Start: "A", Exit: "A",
		Nodes: []*ir.Node{
			{ID: "A", Kind: ir.NodeAgent, Label: "Old Label", Config: ir.AgentConfig{Prompt: "Do."}},
		},
	}
	new := &ir.Workflow{
		Name: "test", Start: "A", Exit: "A",
		Nodes: []*ir.Node{
			{ID: "A", Kind: ir.NodeAgent, Label: "New Label", Config: ir.AgentConfig{Prompt: "Do."}},
		},
	}

	report := diff.Compare(old, new)
	if len(report.NodesModified) != 1 {
		t.Fatalf("expected 1 modified node, got %d", len(report.NodesModified))
	}
	found := false
	for _, c := range report.NodesModified[0].Changes {
		if c.Field == "label" {
			found = true
		}
	}
	if !found {
		t.Error("expected label change")
	}
}

func TestCompare_KindChange(t *testing.T) {
	old := &ir.Workflow{
		Name: "test", Start: "A", Exit: "A",
		Nodes: []*ir.Node{
			{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "Do."}},
		},
	}
	new := &ir.Workflow{
		Name: "test", Start: "A", Exit: "A",
		Nodes: []*ir.Node{
			{ID: "A", Kind: ir.NodeHuman, Config: ir.HumanConfig{Mode: "freeform"}},
		},
	}

	report := diff.Compare(old, new)
	if len(report.NodesModified) != 1 {
		t.Fatalf("expected 1 modified node, got %d", len(report.NodesModified))
	}
	found := false
	for _, c := range report.NodesModified[0].Changes {
		if c.Field == "kind" {
			found = true
		}
	}
	if !found {
		t.Error("expected kind change")
	}
}

func TestCompare_EdgeWithConditions(t *testing.T) {
	old := &ir.Workflow{
		Name: "test", Start: "A", Exit: "B",
		Nodes: []*ir.Node{
			{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "Do."}},
			{ID: "B", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "Done."}},
		},
		Edges: []*ir.Edge{
			{From: "A", To: "B", Condition: &ir.Condition{Raw: "ctx.outcome = success"}},
		},
	}
	new := &ir.Workflow{
		Name: "test", Start: "A", Exit: "B",
		Nodes: []*ir.Node{
			{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "Do."}},
			{ID: "B", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "Done."}},
		},
		Edges: []*ir.Edge{
			{From: "A", To: "B", Condition: &ir.Condition{Raw: "ctx.outcome = fail"}},
		},
	}

	report := diff.Compare(old, new)
	// Old edge removed, new edge added (different condition = different key).
	if len(report.EdgesAdded) != 1 {
		t.Errorf("expected 1 edge added, got %d", len(report.EdgesAdded))
	}
	if len(report.EdgesRemoved) != 1 {
		t.Errorf("expected 1 edge removed, got %d", len(report.EdgesRemoved))
	}
}

func TestCompare_NonAgentConfigNotCompared(t *testing.T) {
	old := &ir.Workflow{
		Name: "test", Start: "T", Exit: "T",
		Nodes: []*ir.Node{
			{ID: "T", Kind: ir.NodeTool, Config: ir.ToolConfig{Command: "echo hi"}},
		},
	}
	new := &ir.Workflow{
		Name: "test", Start: "T", Exit: "T",
		Nodes: []*ir.Node{
			{ID: "T", Kind: ir.NodeTool, Config: ir.ToolConfig{Command: "echo bye"}},
		},
	}

	report := diff.Compare(old, new)
	// compareConfigFields only handles AgentConfig, so ToolConfig diffs are not reported.
	if len(report.NodesModified) != 0 {
		t.Errorf("expected 0 modified nodes for tool config change, got %d", len(report.NodesModified))
	}
}

func TestCompare_MultipleModifiedNodesSorted(t *testing.T) {
	old := &ir.Workflow{
		Name: "test", Start: "A", Exit: "C",
		Nodes: []*ir.Node{
			{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "A1", Model: "m1"}},
			{ID: "B", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "B1", Model: "m1"}},
			{ID: "C", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "C1"}},
		},
		Edges: []*ir.Edge{{From: "A", To: "B"}, {From: "B", To: "C"}},
	}
	new := &ir.Workflow{
		Name: "test", Start: "A", Exit: "C",
		Nodes: []*ir.Node{
			{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "A1", Model: "m2"}},
			{ID: "B", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "B1", Model: "m2"}},
			{ID: "C", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "C1"}},
		},
		Edges: []*ir.Edge{{From: "A", To: "B"}, {From: "B", To: "C"}},
	}

	report := diff.Compare(old, new)
	if len(report.NodesModified) != 2 {
		t.Fatalf("expected 2 modified nodes, got %d", len(report.NodesModified))
	}
	// Should be sorted by NodeID.
	if report.NodesModified[0].NodeID != "A" || report.NodesModified[1].NodeID != "B" {
		t.Errorf("expected sorted [A, B], got [%s, %s]",
			report.NodesModified[0].NodeID, report.NodesModified[1].NodeID)
	}
}
