package ir_test

import (
	"testing"

	"github.com/2389-research/dippin-lang/ir"
)

func TestConditionExprInterface(t *testing.T) {
	// Verify all 4 condition types satisfy the sealed interface.
	var _ ir.ConditionExpr = ir.CondCompare{}
	var _ ir.ConditionExpr = ir.CondAnd{}
	var _ ir.ConditionExpr = ir.CondOr{}
	var _ ir.ConditionExpr = ir.CondNot{}
}

func TestEdgesFromParallel(t *testing.T) {
	w := &ir.Workflow{
		Nodes: []*ir.Node{
			{ID: "split", Kind: ir.NodeParallel, Config: ir.ParallelConfig{Targets: []string{"a", "b"}}},
			{ID: "a", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "A"}},
			{ID: "b", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "B"}},
		},
	}
	edges := w.EdgesFrom("split")
	if len(edges) != 2 {
		t.Fatalf("EdgesFrom(split) = %d, want 2", len(edges))
	}
}

func TestEdgesToFanIn(t *testing.T) {
	w := &ir.Workflow{
		Nodes: []*ir.Node{
			{ID: "a", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "A"}},
			{ID: "b", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "B"}},
			{ID: "join", Kind: ir.NodeFanIn, Config: ir.FanInConfig{Sources: []string{"a", "b"}}},
		},
	}
	edges := w.EdgesTo("join")
	if len(edges) != 2 {
		t.Fatalf("EdgesTo(join) = %d, want 2", len(edges))
	}
}

func TestEdgesDedup(t *testing.T) {
	w := &ir.Workflow{
		Nodes: []*ir.Node{
			{ID: "split", Kind: ir.NodeParallel, Config: ir.ParallelConfig{Targets: []string{"a"}}},
			{ID: "a", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "A"}},
		},
		Edges: []*ir.Edge{
			{From: "split", To: "a"},
		},
	}
	edges := w.EdgesFrom("split")
	// Explicit edge + implicit parallel edge should dedup to 1
	if len(edges) != 1 {
		t.Errorf("EdgesFrom(split) = %d, want 1 (dedup)", len(edges))
	}
}

func TestNodeByIDNotFound(t *testing.T) {
	w := &ir.Workflow{
		Nodes: []*ir.Node{
			{ID: "a", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "A"}},
		},
	}
	if w.Node("nonexistent") != nil {
		t.Error("expected nil for nonexistent node")
	}
}

func TestEdgesFromNonexistent(t *testing.T) {
	w := &ir.Workflow{}
	edges := w.EdgesFrom("ghost")
	if len(edges) != 0 {
		t.Errorf("EdgesFrom(ghost) = %d, want 0", len(edges))
	}
}

func TestEdgesToNonexistent(t *testing.T) {
	w := &ir.Workflow{}
	edges := w.EdgesTo("ghost")
	if len(edges) != 0 {
		t.Errorf("EdgesTo(ghost) = %d, want 0", len(edges))
	}
}
