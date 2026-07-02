package simulate

import (
	"testing"

	"github.com/2389-research/dippin-lang/ir"
)

// nonExhaustiveElseWorkflow: Route has a single guard (ctx.flag = yes) that
// never matches (flag is unset), no unconditional edge, and a section
// `else -> Cleanup`. The guard set is non-exhaustive, so the no-match case
// routes to Cleanup.
func nonExhaustiveElseWorkflow() *ir.Workflow {
	return &ir.Workflow{
		Name: "NonExhaustiveElse", Version: "1", Start: "Route", Exit: "Done",
		ElseTarget: "Cleanup",
		Nodes: []*ir.Node{
			{ID: "Route", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "route"}},
			{ID: "Win", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "win"}},
			{ID: "Cleanup", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "cleanup"}},
			{ID: "Done", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "done"}},
		},
		Edges: []*ir.Edge{
			{From: "Route", To: "Win", Condition: &ir.Condition{Raw: "ctx.flag = yes"}},
			{From: "Win", To: "Done"},
			{From: "Cleanup", To: "Done"},
		},
	}
}

// partitionElseWorkflow: Route guards on a gold/silver complete partition
// (which EdgesExhaustive treats as exhaustive) plus `else -> Cleanup`. A
// concrete scenario value the guards don't cover (tier=bronze) must still route
// to Cleanup — the engine falls to else on any unmatched outcome.
func partitionElseWorkflow() *ir.Workflow {
	return &ir.Workflow{
		Name: "PartitionElse", Version: "1", Start: "Route", Exit: "Done",
		ElseTarget: "Cleanup",
		Nodes: []*ir.Node{
			{ID: "Route", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "route"}},
			{ID: "Gold", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "gold"}},
			{ID: "Silver", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "silver"}},
			{ID: "Cleanup", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "cleanup"}},
			{ID: "Done", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "done"}},
		},
		Edges: []*ir.Edge{
			{From: "Route", To: "Gold", Condition: &ir.Condition{Raw: "ctx.tier = gold"}},
			{From: "Route", To: "Silver", Condition: &ir.Condition{Raw: "ctx.tier = silver"}},
			{From: "Gold", To: "Done"},
			{From: "Silver", To: "Done"},
			{From: "Cleanup", To: "Done"},
		},
	}
}

// exhaustiveElseWorkflow: Route guards on the exhaustive ctx.outcome
// success/fail set with `else -> Cleanup`. Used by the path enumerator, which
// trusts the declared partition and must not enumerate an unreachable else path.
func exhaustiveElseWorkflow() *ir.Workflow {
	return &ir.Workflow{
		Name: "ExhaustiveElse", Version: "1", Start: "Route", Exit: "Done",
		ElseTarget: "Cleanup",
		Nodes: []*ir.Node{
			{ID: "Route", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "route"}},
			{ID: "Win", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "win"}},
			{ID: "Lose", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "lose"}},
			{ID: "Cleanup", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "cleanup"}},
			{ID: "Done", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "done"}},
		},
		Edges: []*ir.Edge{
			{From: "Route", To: "Win", Condition: &ir.Condition{Raw: "ctx.outcome = success"}},
			{From: "Route", To: "Lose", Condition: &ir.Condition{Raw: "ctx.outcome = fail"}},
			{From: "Win", To: "Done"},
			{From: "Lose", To: "Done"},
			{From: "Cleanup", To: "Done"},
		},
	}
}

func pathContains(path []string, id string) bool {
	for _, p := range path {
		if p == id {
			return true
		}
	}
	return false
}

func anyPathContains(results []*Result, id string) bool {
	for _, r := range results {
		if pathContains(r.Path, id) {
			return true
		}
	}
	return false
}

func TestSimulate_RoutesNoMatchToElse(t *testing.T) {
	res, err := Run(nonExhaustiveElseWorkflow(), Options{})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !pathContains(res.Path, "Cleanup") {
		t.Errorf("unmatched guard did not route to else target Cleanup; path=%v", res.Path)
	}
	if pathContains(res.Path, "Win") {
		t.Errorf("guard should not have matched; path=%v", res.Path)
	}
}

// A concrete scenario value that matches no guard (bronze, against a gold/silver
// partition) must route to the else default — even though the partition looks
// statically exhaustive — because that is what the engine does.
func TestSimulate_UnmatchedScenarioValueRoutesToElse(t *testing.T) {
	res, err := Run(partitionElseWorkflow(), Options{Scenario: map[string]string{"tier": "bronze"}})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !pathContains(res.Path, "Cleanup") {
		t.Errorf("unmatched scenario value did not route to else target Cleanup; path=%v", res.Path)
	}
	if pathContains(res.Path, "Gold") || pathContains(res.Path, "Silver") {
		t.Errorf("no partition guard should have matched tier=bronze; path=%v", res.Path)
	}
}

func TestRunAllPaths_EnumeratesElseBranch(t *testing.T) {
	results, err := RunAllPaths(nonExhaustiveElseWorkflow(), nil)
	if err != nil {
		t.Fatalf("RunAllPaths: %v", err)
	}
	if !anyPathContains(results, "Cleanup") {
		t.Errorf("path enumerator did not emit the else branch (Cleanup); %d paths", len(results))
	}
	if !anyPathContains(results, "Win") {
		t.Errorf("path enumerator dropped the guard branch (Win); %d paths", len(results))
	}
}

func TestRunAllPaths_NoElseBranchForExhaustive(t *testing.T) {
	results, err := RunAllPaths(exhaustiveElseWorkflow(), nil)
	if err != nil {
		t.Fatalf("RunAllPaths: %v", err)
	}
	if anyPathContains(results, "Cleanup") {
		t.Errorf("exhaustive node should not enumerate an else branch; %d paths reached Cleanup", len(results))
	}
}
