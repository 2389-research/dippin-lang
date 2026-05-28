package ir_test

import (
	"testing"
	"time"

	"github.com/2389-research/dippin-lang/ir"
)

// Fixture: minimal valid workflow (two nodes, one edge).
func minimalWorkflow() *ir.Workflow {
	return &ir.Workflow{
		Name:    "minimal",
		Version: "1",
		Start:   "Begin",
		Exit:    "End",
		Nodes: []*ir.Node{
			{ID: "Begin", Kind: ir.NodeHuman, Config: ir.HumanConfig{Mode: "freeform"}},
			{ID: "End", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "Done."}},
		},
		Edges: []*ir.Edge{
			{From: "Begin", To: "End"},
		},
	}
}

// Fixture: ask-and-execute (the canonical example from the design doc).
func askAndExecuteWorkflow() *ir.Workflow {
	return &ir.Workflow{
		Name:    "ask_and_execute",
		Version: "1",
		Goal:    "Ask user for a task, implement it, review, ship",
		Start:   "AskUser",
		Exit:    "Done",
		Defaults: ir.WorkflowDefaults{
			Model:       "claude-opus-4-6",
			Provider:    "anthropic",
			RetryPolicy: "standard",
			Fidelity:    "summary:high",
			MaxRestarts: 5,
		},
		Nodes: []*ir.Node{
			{ID: "AskUser", Kind: ir.NodeHuman, Config: ir.HumanConfig{Mode: "freeform"}},
			{
				ID: "Interpret", Kind: ir.NodeAgent,
				IO: ir.NodeIO{Reads: []string{"human_response"}, Writes: []string{"plan"}},
				Config: ir.AgentConfig{
					Prompt: "You are a senior software architect.\n\nRead the user's request below and produce a clear,\nactionable implementation plan.\n\n## User Request\n${ctx.human_response}",
				},
			},
			{ID: "ImplementFanOut", Kind: ir.NodeParallel, Config: ir.ParallelConfig{Targets: []string{"ImplementClaude", "ImplementCodex"}}},
			{
				ID: "ImplementClaude", Kind: ir.NodeAgent,
				Config: ir.AgentConfig{Prompt: "Implement the plan.", Model: "gpt-5.4", Provider: "openai"},
			},
			{
				ID: "ImplementCodex", Kind: ir.NodeAgent,
				Config: ir.AgentConfig{Prompt: "Implement the plan.", Model: "gpt-5.3-codex", Provider: "openai"},
			},
			{ID: "ImplementJoin", Kind: ir.NodeFanIn, Config: ir.FanInConfig{Sources: []string{"ImplementClaude", "ImplementCodex"}}},
			{
				ID: "Validate", Kind: ir.NodeAgent,
				Config: ir.AgentConfig{
					Prompt:     "Review the implementations. Run tests.\nRespond with STATUS: success or STATUS: fail.",
					AutoStatus: true,
					GoalGate:   true,
				},
				Retry: ir.RetryConfig{MaxRetries: 2},
			},
			{ID: "Approve", Kind: ir.NodeHuman, Config: ir.HumanConfig{Mode: "choice", Default: "Yes"}},
			{ID: "Done", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "Ship it."}},
		},
		Edges: []*ir.Edge{
			{From: "AskUser", To: "Interpret"},
			{From: "Interpret", To: "ImplementFanOut"},
			{From: "ImplementJoin", To: "Validate"},
			{From: "Validate", To: "Approve", Condition: &ir.Condition{
				Raw:    "ctx.outcome = success",
				Parsed: ir.CondCompare{Variable: "ctx.outcome", Op: "=", Value: "success"},
			}},
			{From: "Validate", To: "Interpret", Label: "retry", Restart: true, Condition: &ir.Condition{
				Raw:    "ctx.outcome = fail",
				Parsed: ir.CondCompare{Variable: "ctx.outcome", Op: "=", Value: "fail"},
			}},
			{From: "Approve", To: "Done"},
		},
	}
}

// Fixture: tool node with multiline command.
func toolWorkflow() *ir.Workflow {
	return &ir.Workflow{
		Name:    "tool_test",
		Version: "1",
		Start:   "Check",
		Exit:    "Report",
		Nodes: []*ir.Node{
			{
				ID: "Check", Kind: ir.NodeTool,
				IO: ir.NodeIO{Writes: []string{"test_result"}},
				Config: ir.ToolConfig{
					Command: "#!/bin/sh\nset -eu\nif pytest --tb=short 2>&1; then\n  printf 'pass'\nelse\n  printf 'fail'\n  exit 1\nfi",
					Timeout: 60 * time.Second,
				},
			},
			{ID: "Report", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "Report results."}},
		},
		Edges: []*ir.Edge{
			{From: "Check", To: "Report"},
		},
	}
}

// Fixture: subgraph composition.
func subgraphWorkflow() *ir.Workflow {
	return &ir.Workflow{
		Name:    "with_subgraph",
		Version: "1",
		Start:   "Build",
		Exit:    "Done",
		Nodes: []*ir.Node{
			{ID: "Build", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "Build the feature."}},
			{
				ID: "Review", Kind: ir.NodeSubgraph,
				Config: ir.SubgraphConfig{
					Ref:    "./review.dip",
					Params: map[string]string{"model": "gpt-5.4", "strict": "true"},
				},
			},
			{ID: "Done", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "Done."}},
		},
		Edges: []*ir.Edge{
			{From: "Build", To: "Review"},
			{From: "Review", To: "Done", Condition: &ir.Condition{
				Raw:    "ctx.outcome = success",
				Parsed: ir.CondCompare{Variable: "ctx.outcome", Op: "=", Value: "success"},
			}},
			{From: "Review", To: "Build", Restart: true, Condition: &ir.Condition{
				Raw:    "ctx.outcome = fail",
				Parsed: ir.CondCompare{Variable: "ctx.outcome", Op: "=", Value: "fail"},
			}},
		},
	}
}

// Fixture: complex conditions with AND/OR/NOT.
func complexConditionWorkflow() *ir.Workflow {
	return &ir.Workflow{
		Name:    "complex_conditions",
		Version: "1",
		Start:   "Check",
		Exit:    "Done",
		Nodes: []*ir.Node{
			{ID: "Check", Kind: ir.NodeTool, Config: ir.ToolConfig{Command: "echo check"}},
			{ID: "PathA", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "Path A."}},
			{ID: "PathB", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "Path B."}},
			{ID: "Done", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "Done."}},
		},
		Edges: []*ir.Edge{
			{From: "Check", To: "PathA", Condition: &ir.Condition{
				Raw: "ctx.outcome = success and ctx.tool_stdout != empty",
				Parsed: ir.CondAnd{
					Left:  ir.CondCompare{Variable: "ctx.outcome", Op: "=", Value: "success"},
					Right: ir.CondCompare{Variable: "ctx.tool_stdout", Op: "!=", Value: "empty"},
				},
			}},
			{From: "Check", To: "PathB", Condition: &ir.Condition{
				Raw: "not ctx.outcome = success",
				Parsed: ir.CondNot{
					Inner: ir.CondCompare{Variable: "ctx.outcome", Op: "=", Value: "success"},
				},
			}},
			{From: "PathA", To: "Done"},
			{From: "PathB", To: "Done"},
		},
	}
}

func TestWorkflowNodeLookup(t *testing.T) {
	w := minimalWorkflow()

	if n := w.Node("Begin"); n == nil {
		t.Fatal("expected to find node Begin")
	} else if n.Kind != ir.NodeHuman {
		t.Errorf("Begin kind = %q, want %q", n.Kind, ir.NodeHuman)
	}

	if n := w.Node("Nonexistent"); n != nil {
		t.Errorf("expected nil for nonexistent node, got %+v", n)
	}
}

func TestWorkflowEdgesFrom(t *testing.T) {
	w := askAndExecuteWorkflow()

	edges := w.EdgesFrom("Validate")
	if len(edges) != 2 {
		t.Fatalf("EdgesFrom(Validate) = %d edges, want 2", len(edges))
	}

	targets := map[string]bool{}
	for _, e := range edges {
		targets[e.To] = true
	}
	if !targets["Approve"] || !targets["Interpret"] {
		t.Errorf("unexpected edge targets: %v", targets)
	}
}

func TestWorkflowEdgesTo(t *testing.T) {
	w := askAndExecuteWorkflow()

	edges := w.EdgesTo("Interpret")
	if len(edges) != 2 {
		t.Fatalf("EdgesTo(Interpret) = %d edges, want 2", len(edges))
	}

	sources := map[string]bool{}
	for _, e := range edges {
		sources[e.From] = true
	}
	if !sources["AskUser"] || !sources["Validate"] {
		t.Errorf("unexpected edge sources: %v", sources)
	}
}

func TestRestartEdge(t *testing.T) {
	w := askAndExecuteWorkflow()

	edges := w.EdgesFrom("Validate")
	var restartEdge *ir.Edge
	for _, e := range edges {
		if e.Restart {
			restartEdge = e
			break
		}
	}

	if restartEdge == nil {
		t.Fatal("expected a restart edge from Validate")
	}
	if restartEdge.To != "Interpret" {
		t.Errorf("restart edge target = %q, want Interpret", restartEdge.To)
	}
	if restartEdge.Label != "retry" {
		t.Errorf("restart edge label = %q, want retry", restartEdge.Label)
	}
}

func TestNodeKinds(t *testing.T) {
	w := askAndExecuteWorkflow()

	expected := map[string]ir.NodeKind{
		"AskUser":         ir.NodeHuman,
		"Interpret":       ir.NodeAgent,
		"ImplementFanOut": ir.NodeParallel,
		"ImplementClaude": ir.NodeAgent,
		"ImplementCodex":  ir.NodeAgent,
		"ImplementJoin":   ir.NodeFanIn,
		"Validate":        ir.NodeAgent,
		"Approve":         ir.NodeHuman,
		"Done":            ir.NodeAgent,
	}

	for id, wantKind := range expected {
		n := w.Node(id)
		if n == nil {
			t.Errorf("node %q not found", id)
			continue
		}
		if n.Kind != wantKind {
			t.Errorf("node %q kind = %q, want %q", id, n.Kind, wantKind)
		}
	}
}

func TestNodeConfigTypes(t *testing.T) {
	w := askAndExecuteWorkflow()

	// Agent node should have AgentConfig
	interpret := w.Node("Interpret")
	if _, ok := interpret.Config.(ir.AgentConfig); !ok {
		t.Errorf("Interpret config type = %T, want AgentConfig", interpret.Config)
	}

	// Human node should have HumanConfig
	ask := w.Node("AskUser")
	if _, ok := ask.Config.(ir.HumanConfig); !ok {
		t.Errorf("AskUser config type = %T, want HumanConfig", ask.Config)
	}

	// Parallel node should have ParallelConfig
	fanOut := w.Node("ImplementFanOut")
	if pc, ok := fanOut.Config.(ir.ParallelConfig); !ok {
		t.Errorf("ImplementFanOut config type = %T, want ParallelConfig", fanOut.Config)
	} else if len(pc.Targets) != 2 {
		t.Errorf("ParallelConfig targets = %d, want 2", len(pc.Targets))
	}
}

func TestToolConfig(t *testing.T) {
	w := toolWorkflow()

	check := w.Node("Check")
	tc, ok := check.Config.(ir.ToolConfig)
	if !ok {
		t.Fatalf("Check config type = %T, want ToolConfig", check.Config)
	}
	if tc.Timeout != 60*time.Second {
		t.Errorf("timeout = %v, want 60s", tc.Timeout)
	}
	if tc.Command == "" {
		t.Error("command should not be empty")
	}
}

func TestSubgraphConfig(t *testing.T) {
	w := subgraphWorkflow()

	review := w.Node("Review")
	sc, ok := review.Config.(ir.SubgraphConfig)
	if !ok {
		t.Fatalf("Review config type = %T, want SubgraphConfig", review.Config)
	}
	if sc.Ref != "./review.dip" {
		t.Errorf("ref = %q, want ./review.dip", sc.Ref)
	}
	if sc.Params["model"] != "gpt-5.4" {
		t.Errorf("params[model] = %q, want gpt-5.4", sc.Params["model"])
	}
}

func TestNodeIO(t *testing.T) {
	w := askAndExecuteWorkflow()

	interpret := w.Node("Interpret")
	if len(interpret.IO.Reads) != 1 || interpret.IO.Reads[0] != "human_response" {
		t.Errorf("Interpret reads = %v, want [human_response]", interpret.IO.Reads)
	}
	if len(interpret.IO.Writes) != 1 || interpret.IO.Writes[0] != "plan" {
		t.Errorf("Interpret writes = %v, want [plan]", interpret.IO.Writes)
	}
}

func TestConditionAST(t *testing.T) {
	w := complexConditionWorkflow()

	edges := w.EdgesFrom("Check")
	if len(edges) != 2 {
		t.Fatalf("EdgesFrom(Check) = %d, want 2", len(edges))
	}

	// First edge: AND condition
	andEdge := edges[0]
	if andEdge.Condition == nil {
		t.Fatal("expected condition on first edge")
	}
	if _, ok := andEdge.Condition.Parsed.(ir.CondAnd); !ok {
		t.Errorf("first edge condition type = %T, want CondAnd", andEdge.Condition.Parsed)
	}

	// Second edge: NOT condition
	notEdge := edges[1]
	if notEdge.Condition == nil {
		t.Fatal("expected condition on second edge")
	}
	if _, ok := notEdge.Condition.Parsed.(ir.CondNot); !ok {
		t.Errorf("second edge condition type = %T, want CondNot", notEdge.Condition.Parsed)
	}
}

func TestConditionalConfigImplementsNodeConfig(t *testing.T) {
	// Compile-time interface satisfaction check; if ConditionalConfig does
	// not implement NodeConfig, this assignment won't compile.
	var _ ir.NodeConfig = ir.ConditionalConfig{}
}

func TestNodeIDs(t *testing.T) {
	w := minimalWorkflow()
	ids := w.NodeIDs()
	if len(ids) != 2 {
		t.Fatalf("NodeIDs() = %d, want 2", len(ids))
	}
	if ids[0] != "Begin" || ids[1] != "End" {
		t.Errorf("NodeIDs() = %v, want [Begin End]", ids)
	}
}

func TestManagerLoopConfig_ImplementsNodeConfig(t *testing.T) {
	// Compile-time interface satisfaction check; if ManagerLoopConfig does
	// not implement NodeConfig, this assignment won't compile.
	var _ ir.NodeConfig = ir.ManagerLoopConfig{}

	cfg := ir.ManagerLoopConfig{
		SubgraphRef:    "quality_loop",
		PollInterval:   30 * time.Second,
		MaxCycles:      12,
		StopCondition:  &ir.Condition{Raw: "stack.child.cycles = 10"},
		SteerCondition: &ir.Condition{Raw: "stack.child.cycles = 5"},
		SteerContext:   map[string]string{"hint": "speed_up", "priority": "high"},
	}

	if cfg.SubgraphRef != "quality_loop" {
		t.Errorf("SubgraphRef = %q, want %q", cfg.SubgraphRef, "quality_loop")
	}
	if cfg.PollInterval != 30*time.Second {
		t.Errorf("PollInterval = %v, want 30s", cfg.PollInterval)
	}
	if got := cfg.SteerContext["hint"]; got != "speed_up" {
		t.Errorf("SteerContext[hint] = %q, want %q", got, "speed_up")
	}
	if cfg.MaxCycles != 12 {
		t.Errorf("MaxCycles = %d, want 12", cfg.MaxCycles)
	}
	if cfg.StopCondition == nil || cfg.StopCondition.Raw != "stack.child.cycles = 10" {
		t.Errorf("StopCondition.Raw = %q, want %q", cfg.StopCondition.Raw, "stack.child.cycles = 10")
	}
	if cfg.SteerCondition == nil || cfg.SteerCondition.Raw != "stack.child.cycles = 5" {
		t.Errorf("SteerCondition.Raw = %q, want %q", cfg.SteerCondition.Raw, "stack.child.cycles = 5")
	}
}

func TestNodeManagerLoop_IsNodeKind(t *testing.T) {
	if ir.NodeManagerLoop != "manager_loop" {
		t.Errorf("NodeManagerLoop = %q, want %q", ir.NodeManagerLoop, "manager_loop")
	}
}
