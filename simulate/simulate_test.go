package simulate

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/2389-research/dippin-lang/event"
	"github.com/2389-research/dippin-lang/ir"
)

func init() {
	ResetRunCounter()
}

// --- Fixtures ---

func minimalWorkflow() *ir.Workflow {
	return &ir.Workflow{
		Name:    "Minimal",
		Version: "1",
		Start:   "Start",
		Exit:    "Done",
		Nodes: []*ir.Node{
			{ID: "Start", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "Begin."}},
			{ID: "Done", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "End."}},
		},
		Edges: []*ir.Edge{
			{From: "Start", To: "Done"},
		},
	}
}

func conditionalWorkflow() *ir.Workflow {
	return &ir.Workflow{
		Name:    "Conditional",
		Version: "1",
		Start:   "Check",
		Exit:    "Done",
		Defaults: ir.WorkflowDefaults{
			Model:    "claude-opus-4-6",
			Provider: "anthropic",
		},
		Nodes: []*ir.Node{
			{ID: "Check", Kind: ir.NodeAgent, Config: ir.AgentConfig{
				Prompt:     "Check something.",
				AutoStatus: true,
			}},
			{ID: "PathA", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "Path A."}},
			{ID: "PathB", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "Path B."}},
			{ID: "Done", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "Done."}},
		},
		Edges: []*ir.Edge{
			{From: "Check", To: "PathA", Condition: &ir.Condition{
				Raw: "ctx.outcome = success",
			}},
			{From: "Check", To: "PathB", Condition: &ir.Condition{
				Raw: "ctx.outcome = fail",
			}},
			{From: "PathA", To: "Done"},
			{From: "PathB", To: "Done"},
		},
	}
}

func parallelWorkflow() *ir.Workflow {
	return &ir.Workflow{
		Name:    "Parallel",
		Version: "1",
		Start:   "Start",
		Exit:    "Done",
		Nodes: []*ir.Node{
			{ID: "Start", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "Go."}},
			{ID: "FanOut", Kind: ir.NodeParallel, Config: ir.ParallelConfig{
				Targets: []string{"WorkerA", "WorkerB"},
			}},
			{ID: "WorkerA", Kind: ir.NodeAgent, Config: ir.AgentConfig{
				Prompt: "Worker A.", Model: "gpt-5.2", Provider: "openai",
			}},
			{ID: "WorkerB", Kind: ir.NodeAgent, Config: ir.AgentConfig{
				Prompt: "Worker B.", Model: "gemini-3-flash-preview", Provider: "gemini",
			}},
			{ID: "Join", Kind: ir.NodeFanIn, Config: ir.FanInConfig{
				Sources: []string{"WorkerA", "WorkerB"},
			}},
			{ID: "Done", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "Done."}},
		},
		Edges: []*ir.Edge{
			{From: "Start", To: "FanOut"},
			{From: "FanOut", To: "WorkerA"},
			{From: "FanOut", To: "WorkerB"},
			{From: "WorkerA", To: "Join"},
			{From: "WorkerB", To: "Join"},
			{From: "Join", To: "Done"},
		},
	}
}

func humanWorkflow() *ir.Workflow {
	return &ir.Workflow{
		Name:    "HumanGate",
		Version: "1",
		Start:   "Start",
		Exit:    "Done",
		Nodes: []*ir.Node{
			{ID: "Start", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "Begin."}},
			{ID: "Ask", Kind: ir.NodeHuman, Label: "What do you want?", Config: ir.HumanConfig{Mode: "freeform"}},
			{ID: "Done", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "Done."}},
		},
		Edges: []*ir.Edge{
			{From: "Start", To: "Ask"},
			{From: "Ask", To: "Done"},
		},
	}
}

func toolWorkflow() *ir.Workflow {
	return &ir.Workflow{
		Name:    "ToolTest",
		Version: "1",
		Start:   "Run",
		Exit:    "Done",
		Nodes: []*ir.Node{
			{ID: "Run", Kind: ir.NodeTool, Config: ir.ToolConfig{
				Command: "echo hello", Timeout: 30 * time.Second,
			}},
			{ID: "Done", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "Done."}},
		},
		Edges: []*ir.Edge{
			{From: "Run", To: "Done"},
		},
	}
}

func complexConditionWorkflow() *ir.Workflow {
	return &ir.Workflow{
		Name:    "ComplexCond",
		Version: "1",
		Start:   "Check",
		Exit:    "Done",
		Nodes: []*ir.Node{
			{ID: "Check", Kind: ir.NodeTool, Config: ir.ToolConfig{Command: "check"}},
			{ID: "PathA", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "Path A."}},
			{ID: "PathB", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "Path B."}},
			{ID: "Done", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "Done."}},
		},
		Edges: []*ir.Edge{
			{From: "Check", To: "PathA", Condition: &ir.Condition{
				Raw: "ctx.outcome = success and ctx.tool_stdout != empty",
			}},
			{From: "Check", To: "PathB", Condition: &ir.Condition{
				Raw: "not ctx.outcome = success",
			}},
			{From: "PathA", To: "Done"},
			{From: "PathB", To: "Done"},
		},
	}
}

func restartWorkflow() *ir.Workflow {
	return &ir.Workflow{
		Name:    "RestartLoop",
		Version: "1",
		Start:   "Impl",
		Exit:    "Done",
		Nodes: []*ir.Node{
			{ID: "Impl", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "Implement."}},
			{ID: "Review", Kind: ir.NodeAgent, Config: ir.AgentConfig{
				Prompt: "Review.", AutoStatus: true,
			}},
			{ID: "Done", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "Done."}},
		},
		Edges: []*ir.Edge{
			{From: "Impl", To: "Review"},
			{From: "Review", To: "Done", Condition: &ir.Condition{
				Raw: "ctx.outcome = success",
			}},
			{From: "Review", To: "Impl", Label: "retry", Restart: true, Condition: &ir.Condition{
				Raw: "ctx.outcome = fail",
			}},
		},
	}
}

func subgraphWorkflow() *ir.Workflow {
	return &ir.Workflow{
		Name:    "WithSubgraph",
		Version: "1",
		Start:   "Build",
		Exit:    "Done",
		Nodes: []*ir.Node{
			{ID: "Build", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "Build."}},
			{ID: "Sub", Kind: ir.NodeSubgraph, Config: ir.SubgraphConfig{
				Ref:    "./review.dip",
				Params: map[string]string{"model": "gpt-5.4"},
			}},
			{ID: "Done", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "Done."}},
		},
		Edges: []*ir.Edge{
			{From: "Build", To: "Sub"},
			{From: "Sub", To: "Done"},
		},
	}
}

func unconditionalFallbackWorkflow() *ir.Workflow {
	return &ir.Workflow{
		Name:    "Fallback",
		Version: "1",
		Start:   "Check",
		Exit:    "Done",
		Nodes: []*ir.Node{
			{ID: "Check", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "Check."}},
			{ID: "Special", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "Special."}},
			{ID: "Default", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "Default."}},
			{ID: "Done", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "Done."}},
		},
		Edges: []*ir.Edge{
			{From: "Check", To: "Special", Condition: &ir.Condition{
				Raw: "ctx.outcome = special",
			}},
			{From: "Check", To: "Default"}, // unconditional fallback
			{From: "Special", To: "Done"},
			{From: "Default", To: "Done"},
		},
	}
}

// --- Tests ---

func TestRunMinimal(t *testing.T) {
	ResetRunCounter()
	w := minimalWorkflow()

	res, err := Run(w, Options{})
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}
	if res.Status != "success" {
		t.Errorf("Status = %q, want success", res.Status)
	}
	if res.NodesVisited != 2 {
		t.Errorf("NodesVisited = %d, want 2", res.NodesVisited)
	}
	if len(res.Path) != 2 || res.Path[0] != "Start" || res.Path[1] != "Done" {
		t.Errorf("Path = %v, want [Start Done]", res.Path)
	}

	// Verify event sequence.
	assertEventSequence(t, res.Events, []event.Type{
		event.TypePipelineStart,
		event.TypeNodeEnter, // Start
		event.TypeNodeExit,  // Start
		event.TypeEdgeTraverse,
		event.TypeNodeEnter, // Done
		event.TypeNodeExit,  // Done
		event.TypePipelineEnd,
	})
}

func TestRunConditional_Success(t *testing.T) {
	ResetRunCounter()
	w := conditionalWorkflow()

	res, err := Run(w, Options{
		Scenario: map[string]string{"outcome": "success"},
	})
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}
	if res.Status != "success" {
		t.Errorf("Status = %q, want success", res.Status)
	}
	assertPathContains(t, res.Path, "PathA")
	assertPathNotContains(t, res.Path, "PathB")
}

func TestRunConditional_Fail(t *testing.T) {
	ResetRunCounter()
	w := conditionalWorkflow()

	res, err := Run(w, Options{
		Scenario: map[string]string{"outcome": "fail"},
	})
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}
	if res.Status != "success" {
		t.Errorf("Status = %q, want success", res.Status)
	}
	assertPathContains(t, res.Path, "PathB")
	assertPathNotContains(t, res.Path, "PathA")
}

func TestRunConditional_DefaultAutoStatus(t *testing.T) {
	ResetRunCounter()
	w := conditionalWorkflow()

	// No scenario — AutoStatus defaults to "success".
	res, err := Run(w, Options{})
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}
	assertPathContains(t, res.Path, "PathA")
}

func TestRunParallel(t *testing.T) {
	ResetRunCounter()
	w := parallelWorkflow()

	res, err := Run(w, Options{})
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}
	if res.Status != "success" {
		t.Errorf("Status = %q, want success", res.Status)
	}

	// Should have parallel_start and parallel_end events.
	var hasParStart, hasParEnd bool
	for _, ev := range res.Events {
		switch ev.EventType() {
		case event.TypeParallelStart:
			hasParStart = true
			ps := ev.(event.ParallelStart)
			if len(ps.Targets) != 2 {
				t.Errorf("ParallelStart targets = %d, want 2", len(ps.Targets))
			}
		case event.TypeParallelEnd:
			hasParEnd = true
			pe := ev.(event.ParallelEnd)
			if len(pe.Sources) != 2 {
				t.Errorf("ParallelEnd sources = %d, want 2", len(pe.Sources))
			}
		}
	}
	if !hasParStart {
		t.Error("missing parallel_start event")
	}
	if !hasParEnd {
		t.Error("missing parallel_end event")
	}
}

func TestRunHumanAutoSuccess(t *testing.T) {
	ResetRunCounter()
	w := humanWorkflow()

	res, err := Run(w, Options{})
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}
	if res.Status != "success" {
		t.Errorf("Status = %q, want success", res.Status)
	}
	assertPathContains(t, res.Path, "Ask")
}

func TestRunHumanInteractive(t *testing.T) {
	ResetRunCounter()
	w := humanWorkflow()

	input := strings.NewReader("build a spaceship\n")
	var stderr bytes.Buffer

	res, err := Run(w, Options{
		Interactive: true,
		Stdin:       input,
		Stderr:      &stderr,
	})
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}
	if res.Status != "success" {
		t.Errorf("Status = %q, want success", res.Status)
	}

	// Should have emitted a context_update for human_response.
	var found bool
	for _, ev := range res.Events {
		if cu, ok := ev.(event.ContextUpdate); ok && cu.Key == "human_response" {
			if cu.Value != "build a spaceship" {
				t.Errorf("human_response = %q, want %q", cu.Value, "build a spaceship")
			}
			found = true
		}
	}
	if !found {
		t.Error("expected context_update for human_response")
	}

	// Stderr should contain the prompt.
	if !strings.Contains(stderr.String(), "What do you want?") {
		t.Errorf("stderr = %q, expected human prompt", stderr.String())
	}
}

func TestRunTool(t *testing.T) {
	ResetRunCounter()
	w := toolWorkflow()

	res, err := Run(w, Options{})
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}
	if res.Status != "success" {
		t.Errorf("Status = %q, want success", res.Status)
	}

	// Verify tool node emits command in enter event.
	var foundCommand bool
	for _, ev := range res.Events {
		if ne, ok := ev.(event.NodeEnter); ok && ne.Node == "Run" {
			if ne.Command != "echo hello" {
				t.Errorf("command = %q, want %q", ne.Command, "echo hello")
			}
			foundCommand = true
		}
	}
	if !foundCommand {
		t.Error("expected node_enter event for tool node with command")
	}
}

func TestRunComplexCondition(t *testing.T) {
	ResetRunCounter()
	w := complexConditionWorkflow()

	// outcome=success AND tool_stdout!=empty → PathA
	res, err := Run(w, Options{
		Scenario: map[string]string{
			"outcome":     "success",
			"tool_stdout": "some output",
		},
	})
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}
	assertPathContains(t, res.Path, "PathA")

	// outcome=fail → NOT(outcome=success) → PathB
	ResetRunCounter()
	res2, err := Run(w, Options{
		Scenario: map[string]string{"outcome": "fail"},
	})
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}
	assertPathContains(t, res2.Path, "PathB")
}

func TestRunSubgraph(t *testing.T) {
	ResetRunCounter()
	w := subgraphWorkflow()

	res, err := Run(w, Options{})
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}
	if res.Status != "success" {
		t.Errorf("Status = %q, want success", res.Status)
	}
	assertPathContains(t, res.Path, "Sub")

	// Verify subgraph label.
	for _, ev := range res.Events {
		if ne, ok := ev.(event.NodeEnter); ok && ne.Node == "Sub" {
			if ne.Label != "subgraph:./review.dip" {
				t.Errorf("subgraph label = %q, want %q", ne.Label, "subgraph:./review.dip")
			}
		}
	}
}

func TestRunUnconditionalFallback(t *testing.T) {
	ResetRunCounter()
	w := unconditionalFallbackWorkflow()

	// No scenario match → should take unconditional edge to Default.
	res, err := Run(w, Options{})
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}
	assertPathContains(t, res.Path, "Default")
	assertPathNotContains(t, res.Path, "Special")
}

func TestRunUnconditionalFallback_ConditionMatches(t *testing.T) {
	ResetRunCounter()
	w := unconditionalFallbackWorkflow()

	res, err := Run(w, Options{
		Scenario: map[string]string{"outcome": "special"},
	})
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}
	assertPathContains(t, res.Path, "Special")
	assertPathNotContains(t, res.Path, "Default")
}

func TestRunRestart_FailThenSuccess(t *testing.T) {
	ResetRunCounter()
	w := restartWorkflow()

	// Scenario: outcome=fail, which will cause a loop.
	// Since AutoStatus is true and we set outcome=fail, Review → Impl (restart).
	// On second visit, outcome is still fail, so it loops again.
	// But the loop cap (maxSteps/visited) should prevent infinite loop.
	res, err := Run(w, Options{
		Scenario: map[string]string{"outcome": "fail"},
	})
	if err != nil {
		// Expected — should hit step limit or complete.
		t.Logf("Expected error for infinite fail loop: %v", err)
		return
	}

	// If it completed, it must have found a path somehow.
	t.Logf("Completed with status=%s, path=%v", res.Status, res.Path)
}

func TestRunRestart_Success(t *testing.T) {
	ResetRunCounter()
	w := restartWorkflow()

	res, err := Run(w, Options{
		Scenario: map[string]string{"outcome": "success"},
	})
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}
	if res.Status != "success" {
		t.Errorf("Status = %q, want success", res.Status)
	}
	assertPathContains(t, res.Path, "Done")
}

func TestRunMissingStart(t *testing.T) {
	ResetRunCounter()
	w := &ir.Workflow{Name: "bad", Exit: "Done"}
	_, err := Run(w, Options{})
	if err == nil {
		t.Fatal("expected error for missing start node")
	}
}

func TestRunMissingExit(t *testing.T) {
	ResetRunCounter()
	w := &ir.Workflow{Name: "bad", Start: "Start"}
	_, err := Run(w, Options{})
	if err == nil {
		t.Fatal("expected error for missing exit node")
	}
}

func TestRunStartNodeNotFound(t *testing.T) {
	ResetRunCounter()
	w := &ir.Workflow{
		Name:  "bad",
		Start: "Ghost",
		Exit:  "Done",
		Nodes: []*ir.Node{
			{ID: "Done", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "end"}},
		},
	}
	_, err := Run(w, Options{})
	if err == nil {
		t.Fatal("expected error for start node not found")
	}
}

// --- AllPaths Tests ---

func TestRunAllPaths_Conditional(t *testing.T) {
	ResetRunCounter()
	w := conditionalWorkflow()

	results, err := RunAllPaths(w, Options{})
	if err != nil {
		t.Fatalf("RunAllPaths() error: %v", err)
	}

	// Should find at least 2 paths (through PathA and PathB).
	if len(results) < 2 {
		t.Errorf("RunAllPaths found %d paths, want at least 2", len(results))
	}

	var pathAFound, pathBFound bool
	for _, r := range results {
		for _, n := range r.Path {
			if n == "PathA" {
				pathAFound = true
			}
			if n == "PathB" {
				pathBFound = true
			}
		}
	}
	if !pathAFound {
		t.Error("RunAllPaths missing path through PathA")
	}
	if !pathBFound {
		t.Error("RunAllPaths missing path through PathB")
	}
}

func TestRunAllPaths_Minimal(t *testing.T) {
	ResetRunCounter()
	w := minimalWorkflow()

	results, err := RunAllPaths(w, Options{})
	if err != nil {
		t.Fatalf("RunAllPaths() error: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("RunAllPaths found %d paths, want 1", len(results))
	}
}

// --- EmitJSONL Tests ---

func TestEmitJSONL(t *testing.T) {
	ResetRunCounter()
	w := minimalWorkflow()

	res, err := Run(w, Options{})
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	var buf bytes.Buffer
	if err := EmitJSONL(&buf, res.Events); err != nil {
		t.Fatalf("EmitJSONL() error: %v", err)
	}

	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) != len(res.Events) {
		t.Errorf("JSONL lines = %d, want %d", len(lines), len(res.Events))
	}

	// Each line should be valid JSON.
	for i, line := range lines {
		var m map[string]interface{}
		if err := json.Unmarshal([]byte(line), &m); err != nil {
			t.Errorf("line %d not valid JSON: %v\nline: %s", i, err, line)
		}
		// Each line should have an "event" field.
		if _, ok := m["event"]; !ok {
			t.Errorf("line %d missing 'event' field", i)
		}
		// Each line should have a "timestamp" field.
		if _, ok := m["timestamp"]; !ok {
			t.Errorf("line %d missing 'timestamp' field", i)
		}
	}
}

func TestEmitJSONL_FirstLine(t *testing.T) {
	ResetRunCounter()
	w := minimalWorkflow()

	res, err := Run(w, Options{})
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	var buf bytes.Buffer
	if err := EmitJSONL(&buf, res.Events); err != nil {
		t.Fatalf("EmitJSONL() error: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	var first map[string]interface{}
	if err := json.Unmarshal([]byte(lines[0]), &first); err != nil {
		t.Fatalf("first line not valid JSON: %v", err)
	}

	if first["event"] != "pipeline_start" {
		t.Errorf("first event = %q, want pipeline_start", first["event"])
	}
	if first["workflow"] != "Minimal" {
		t.Errorf("workflow = %q, want Minimal", first["workflow"])
	}
}

func TestEmitJSONL_LastLine(t *testing.T) {
	ResetRunCounter()
	w := minimalWorkflow()

	res, err := Run(w, Options{})
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	var buf bytes.Buffer
	if err := EmitJSONL(&buf, res.Events); err != nil {
		t.Fatalf("EmitJSONL() error: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	var last map[string]interface{}
	if err := json.Unmarshal([]byte(lines[len(lines)-1]), &last); err != nil {
		t.Fatalf("last line not valid JSON: %v", err)
	}

	if last["event"] != "pipeline_end" {
		t.Errorf("last event = %q, want pipeline_end", last["event"])
	}
	if last["status"] != "success" {
		t.Errorf("status = %q, want success", last["status"])
	}
}

func TestEdgeTraverseIncludesCondition(t *testing.T) {
	ResetRunCounter()
	w := conditionalWorkflow()

	res, err := Run(w, Options{
		Scenario: map[string]string{"outcome": "success"},
	})
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	var found bool
	for _, ev := range res.Events {
		if et, ok := ev.(event.EdgeTraverse); ok && et.From == "Check" {
			if et.Condition != "ctx.outcome = success" {
				t.Errorf("edge condition = %q, want %q", et.Condition, "ctx.outcome = success")
			}
			found = true
		}
	}
	if !found {
		t.Error("expected edge_traverse event from Check")
	}
}

func TestNodeEnterIncludesModel(t *testing.T) {
	ResetRunCounter()
	w := parallelWorkflow()

	res, err := Run(w, Options{})
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	for _, ev := range res.Events {
		if ne, ok := ev.(event.NodeEnter); ok && ne.Node == "WorkerA" {
			if ne.Model != "gpt-5.2" {
				t.Errorf("WorkerA model = %q, want %q", ne.Model, "gpt-5.2")
			}
			if ne.Provider != "openai" {
				t.Errorf("WorkerA provider = %q, want %q", ne.Provider, "openai")
			}
			return
		}
	}
	t.Error("expected node_enter for WorkerA")
}

func TestContextUpdateEmitted(t *testing.T) {
	ResetRunCounter()
	w := conditionalWorkflow()

	res, err := Run(w, Options{})
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	// AutoStatus on Check should produce a context_update.
	var found bool
	for _, ev := range res.Events {
		if cu, ok := ev.(event.ContextUpdate); ok && cu.Key == "outcome" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected context_update for 'outcome' from AutoStatus node")
	}
}

// --- Multi-gate workflow for scenario stickiness tests ---

// multiGateWorkflow creates a workflow with two sequential phases, each
// followed by a conditional gate. This tests that scenario values remain
// effective across multiple gates, not just the first one.
//
//	Start → Phase1(tool) → Gate1 ─[outcome=success]→ Phase2(agent,auto_status)
//	                             └─[outcome=fail]──→ Fail1 → Done
//	Phase2 → Gate2 ─[outcome=success]→ Success → Done
//	               └─[outcome=fail]──→ Fail2 → Done
func multiGateWorkflow() *ir.Workflow {
	return &ir.Workflow{
		Name:    "MultiGate",
		Version: "1",
		Start:   "Start",
		Exit:    "Done",
		Nodes: []*ir.Node{
			{ID: "Start", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "Begin."}},
			{ID: "Phase1", Kind: ir.NodeTool, Config: ir.ToolConfig{Command: "build"}},
			{ID: "Gate1", Kind: ir.NodeAgent, Config: ir.AgentConfig{
				Prompt: "Gate 1.", AutoStatus: true,
			}},
			{ID: "Phase2", Kind: ir.NodeAgent, Config: ir.AgentConfig{
				Prompt: "Phase 2.", AutoStatus: true,
			}},
			{ID: "Gate2", Kind: ir.NodeAgent, Config: ir.AgentConfig{
				Prompt: "Gate 2.", AutoStatus: true,
			}},
			{ID: "Fail1", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "Fail at phase 1."}},
			{ID: "Fail2", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "Fail at phase 2."}},
			{ID: "Success", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "All passed."}},
			{ID: "Done", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "Done."}},
		},
		Edges: []*ir.Edge{
			{From: "Start", To: "Phase1"},
			{From: "Phase1", To: "Gate1"},
			{From: "Gate1", To: "Phase2", Condition: &ir.Condition{
				Raw: "ctx.outcome = success",
			}},
			{From: "Gate1", To: "Fail1", Condition: &ir.Condition{
				Raw: "ctx.outcome = fail",
			}},
			{From: "Phase2", To: "Gate2"},
			{From: "Gate2", To: "Success", Condition: &ir.Condition{
				Raw: "ctx.outcome = success",
			}},
			{From: "Gate2", To: "Fail2", Condition: &ir.Condition{
				Raw: "ctx.outcome = fail",
			}},
			{From: "Fail1", To: "Done"},
			{From: "Fail2", To: "Done"},
			{From: "Success", To: "Done"},
		},
	}
}

func TestRunScenario_StickyAcrossMultipleGates(t *testing.T) {
	ResetRunCounter()
	w := multiGateWorkflow()

	// With outcome=fail, Gate1 should take the fail branch immediately.
	res, err := Run(w, Options{
		Scenario: map[string]string{"outcome": "fail"},
	})
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}
	if res.Status != "success" {
		t.Errorf("Status = %q, want success", res.Status)
	}
	assertPathContains(t, res.Path, "Fail1")
	assertPathNotContains(t, res.Path, "Phase2")
	assertPathNotContains(t, res.Path, "Success")
}

func TestRunScenario_NoScenarioDefaultsToSuccess(t *testing.T) {
	ResetRunCounter()
	w := multiGateWorkflow()

	// Without scenario, tool/auto_status defaults seed outcome=success.
	res, err := Run(w, Options{})
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}
	assertPathContains(t, res.Path, "Phase2")
	assertPathContains(t, res.Path, "Success")
	assertPathNotContains(t, res.Path, "Fail1")
	assertPathNotContains(t, res.Path, "Fail2")
}

func TestRunScenario_ToolStdoutOverride(t *testing.T) {
	ResetRunCounter()

	// Workflow with a tool node whose gate checks tool_stdout.
	w := &ir.Workflow{
		Name:    "ToolStdout",
		Version: "1",
		Start:   "Run",
		Exit:    "Done",
		Nodes: []*ir.Node{
			{ID: "Run", Kind: ir.NodeTool, Config: ir.ToolConfig{Command: "test"}},
			{ID: "OK", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "OK."}},
			{ID: "Bad", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "Bad."}},
			{ID: "Done", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "Done."}},
		},
		Edges: []*ir.Edge{
			{From: "Run", To: "OK", Condition: &ir.Condition{
				Raw: "ctx.tool_stdout = success",
			}},
			{From: "Run", To: "Bad", Condition: &ir.Condition{
				Raw: "ctx.tool_stdout = fail",
			}},
			{From: "OK", To: "Done"},
			{From: "Bad", To: "Done"},
		},
	}

	res, err := Run(w, Options{
		Scenario: map[string]string{"tool_stdout": "fail"},
	})
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}
	assertPathContains(t, res.Path, "Bad")
	assertPathNotContains(t, res.Path, "OK")
}

// --- Loop Detection Tests ---

// toolGatedLoopWorkflow creates a workflow with a tool-gated loop:
// Start → PickNext(tool) → Work → PickNext (restart)
// PickNext → Done when ctx.tool_stdout contains all-done
// The tool default (tool_stdout=success) never contains "all-done",
// so without MaxNodeVisits the loop runs to maxSteps.
func toolGatedLoopWorkflow() *ir.Workflow {
	return &ir.Workflow{
		Name:  "ToolLoop",
		Start: "Start",
		Exit:  "Done",
		Nodes: []*ir.Node{
			{ID: "Start", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "Go."}},
			{ID: "PickNext", Kind: ir.NodeTool, Config: ir.ToolConfig{Command: "pick"}},
			{ID: "Work", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "Do work."}},
			{ID: "Done", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "Done."}},
		},
		Edges: []*ir.Edge{
			{From: "Start", To: "PickNext"},
			{From: "PickNext", To: "Work", Condition: &ir.Condition{
				Raw: "ctx.tool_stdout not contains all-done",
			}},
			{From: "PickNext", To: "Done", Condition: &ir.Condition{
				Raw: "ctx.tool_stdout contains all-done",
			}},
			{From: "Work", To: "PickNext", Restart: true},
		},
	}
}

func TestMaxNodeVisits_BreaksToolLoop(t *testing.T) {
	ResetRunCounter()
	w := toolGatedLoopWorkflow()

	res, err := Run(w, Options{MaxNodeVisits: 3})
	if err != nil {
		t.Fatalf("expected loop to break, got error: %v", err)
	}
	if res.Status != "success" {
		t.Errorf("Status = %q, want success", res.Status)
	}
	assertPathContains(t, res.Path, "Done")
}

func TestMaxNodeVisits_Zero_UsesMaxSteps(t *testing.T) {
	ResetRunCounter()
	w := toolGatedLoopWorkflow()

	// Without MaxNodeVisits, should hit maxSteps error.
	_, err := Run(w, Options{MaxNodeVisits: 0})
	if err == nil {
		t.Fatal("expected maxSteps error without MaxNodeVisits")
	}
	if !strings.Contains(err.Error(), "500 steps") {
		t.Errorf("error = %q, want maxSteps message", err.Error())
	}
}

func TestPerNodeScenario_InLoop(t *testing.T) {
	ResetRunCounter()
	w := restartWorkflow()

	// Per-node scenario: Review.outcome=fail causes a loop.
	// With MaxNodeVisits, the loop should break and exit.
	res, err := Run(w, Options{
		Scenario:      map[string]string{"Review.outcome": "fail"},
		MaxNodeVisits: 3,
	})
	if err != nil {
		t.Fatalf("expected loop to break, got error: %v", err)
	}
	// After loop breaking, should reach Done.
	assertPathContains(t, res.Path, "Done")
	// Review should have been visited multiple times.
	reviewCount := 0
	for _, n := range res.Path {
		if n == "Review" {
			reviewCount++
		}
	}
	if reviewCount < 2 {
		t.Errorf("Review visited %d times, expected at least 2", reviewCount)
	}
}

// --- Condition evaluation tests ---

func orConditionWorkflow() *ir.Workflow {
	return &ir.Workflow{
		Name:    "OrCond",
		Version: "1",
		Start:   "Check",
		Exit:    "Done",
		Nodes: []*ir.Node{
			{ID: "Check", Kind: ir.NodeAgent, Config: ir.AgentConfig{
				Prompt: "Check.", AutoStatus: true,
			}},
			{ID: "PathA", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "A."}},
			{ID: "PathB", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "B."}},
			{ID: "Done", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "Done."}},
		},
		Edges: []*ir.Edge{
			{From: "Check", To: "PathA", Condition: &ir.Condition{
				Raw: "ctx.outcome = success or ctx.outcome = partial",
			}},
			{From: "Check", To: "PathB", Condition: &ir.Condition{
				Raw: "ctx.outcome = fail",
			}},
			{From: "PathA", To: "Done"},
			{From: "PathB", To: "Done"},
		},
	}
}

func TestEvalCondition_Or(t *testing.T) {
	ResetRunCounter()
	w := orConditionWorkflow()

	// "partial" matches the OR's right branch.
	res, err := Run(w, Options{
		Scenario: map[string]string{"outcome": "partial"},
	})
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}
	assertPathContains(t, res.Path, "PathA")
	assertPathNotContains(t, res.Path, "PathB")

	// "success" matches the OR's left branch.
	ResetRunCounter()
	res2, err := Run(w, Options{
		Scenario: map[string]string{"outcome": "success"},
	})
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}
	assertPathContains(t, res2.Path, "PathA")
}

func inConditionWorkflow() *ir.Workflow {
	return &ir.Workflow{
		Name:    "InCond",
		Version: "1",
		Start:   "Check",
		Exit:    "Done",
		Nodes: []*ir.Node{
			{ID: "Check", Kind: ir.NodeAgent, Config: ir.AgentConfig{
				Prompt: "Check.", AutoStatus: true,
			}},
			{ID: "PathA", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "A."}},
			{ID: "PathB", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "B."}},
			{ID: "Done", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "Done."}},
		},
		Edges: []*ir.Edge{
			{From: "Check", To: "PathA", Condition: &ir.Condition{
				Raw: "ctx.outcome in success,partial,done",
			}},
			{From: "Check", To: "PathB", Condition: &ir.Condition{
				Raw: "ctx.outcome = fail",
			}},
			{From: "PathA", To: "Done"},
			{From: "PathB", To: "Done"},
		},
	}
}

func TestEvalCondition_In(t *testing.T) {
	ResetRunCounter()
	w := inConditionWorkflow()

	// "partial" is in the comma-separated list.
	res, err := Run(w, Options{
		Scenario: map[string]string{"outcome": "partial"},
	})
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}
	assertPathContains(t, res.Path, "PathA")
	assertPathNotContains(t, res.Path, "PathB")

	// "fail" doesn't match the in-list.
	ResetRunCounter()
	res2, err := Run(w, Options{
		Scenario: map[string]string{"outcome": "fail"},
	})
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}
	assertPathContains(t, res2.Path, "PathB")
	assertPathNotContains(t, res2.Path, "PathA")
}

// --- forceLoopExit tests ---

// forceLoopExitWithUnconditional tests the path where forceLoopExit
// finds no non-matching conditional but does find an unconditional edge.
func unconditionalExitLoopWorkflow() *ir.Workflow {
	return &ir.Workflow{
		Name:  "UnconditionalExitLoop",
		Start: "Start",
		Exit:  "Done",
		Nodes: []*ir.Node{
			{ID: "Start", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "Go."}},
			{ID: "Gate", Kind: ir.NodeAgent, Config: ir.AgentConfig{
				Prompt: "Gate.", AutoStatus: true,
			}},
			{ID: "Work", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "Work."}},
			{ID: "Done", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "Done."}},
		},
		Edges: []*ir.Edge{
			{From: "Start", To: "Gate"},
			// One conditional that always matches (loops), plus an unconditional fallback.
			{From: "Gate", To: "Work", Condition: &ir.Condition{
				Raw: "ctx.outcome = success",
			}},
			{From: "Gate", To: "Done"}, // unconditional
			{From: "Work", To: "Gate", Restart: true},
		},
	}
}

func TestForceLoopExit_TakesUnconditionalEdge(t *testing.T) {
	ResetRunCounter()
	w := unconditionalExitLoopWorkflow()

	res, err := Run(w, Options{MaxNodeVisits: 2})
	if err != nil {
		t.Fatalf("expected loop to break, got error: %v", err)
	}
	if res.Status != "success" {
		t.Errorf("Status = %q, want success", res.Status)
	}
	assertPathContains(t, res.Path, "Done")
}

// --- seedConditionContext / path enumerator tests ---

func TestRunAllPaths_OrCondition(t *testing.T) {
	ResetRunCounter()
	w := orConditionWorkflow()

	results, err := RunAllPaths(w, Options{})
	if err != nil {
		t.Fatalf("RunAllPaths() error: %v", err)
	}
	// Should find paths through both edges.
	if len(results) < 2 {
		t.Errorf("RunAllPaths found %d paths, want at least 2", len(results))
	}
}

func TestRunAllPaths_InCondition(t *testing.T) {
	ResetRunCounter()
	w := inConditionWorkflow()

	results, err := RunAllPaths(w, Options{})
	if err != nil {
		t.Fatalf("RunAllPaths() error: %v", err)
	}
	if len(results) < 2 {
		t.Errorf("RunAllPaths found %d paths, want at least 2", len(results))
	}
}

func TestRunAllPaths_ComplexCondition(t *testing.T) {
	ResetRunCounter()
	w := complexConditionWorkflow()

	results, err := RunAllPaths(w, Options{})
	if err != nil {
		t.Fatalf("RunAllPaths() error: %v", err)
	}
	if len(results) < 2 {
		t.Errorf("RunAllPaths found %d paths, want at least 2", len(results))
	}
}

func TestRunAllPaths_RestartLoop(t *testing.T) {
	ResetRunCounter()
	w := restartWorkflow()

	results, err := RunAllPaths(w, Options{})
	if err != nil {
		t.Fatalf("RunAllPaths() error: %v", err)
	}
	// Should find paths that include the restart edge.
	if len(results) < 2 {
		t.Errorf("RunAllPaths found %d paths, want at least 2", len(results))
	}
}

func TestRunAllPaths_Parallel(t *testing.T) {
	ResetRunCounter()
	w := parallelWorkflow()

	results, err := RunAllPaths(w, Options{})
	if err != nil {
		t.Fatalf("RunAllPaths() error: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("RunAllPaths found 0 paths")
	}
	// Verify parallel/fan_in events appear in path results.
	var hasParStart, hasParEnd bool
	for _, r := range results {
		for _, ev := range r.Events {
			switch ev.EventType() {
			case event.TypeParallelStart:
				hasParStart = true
			case event.TypeParallelEnd:
				hasParEnd = true
			}
		}
	}
	if !hasParStart {
		t.Error("RunAllPaths missing parallel_start event")
	}
	if !hasParEnd {
		t.Error("RunAllPaths missing parallel_end event")
	}
}

func TestRunAllPaths_MissingStart(t *testing.T) {
	ResetRunCounter()
	w := &ir.Workflow{Name: "bad", Exit: "Done"}
	_, err := RunAllPaths(w, Options{})
	if err == nil {
		t.Fatal("expected error for missing start node")
	}
}

// --- Additional condition parsing edge cases ---

func TestParseCondition_Empty(t *testing.T) {
	_, err := ParseCondition("")
	if err == nil {
		t.Fatal("expected error for empty condition")
	}
}

func TestParseCondition_TrailingTokens(t *testing.T) {
	_, err := ParseCondition("ctx.x = y extra")
	if err == nil {
		t.Fatal("expected error for trailing tokens")
	}
}

func TestParseCondition_DoubleEquals(t *testing.T) {
	expr, err := ParseCondition("ctx.x == value")
	if err != nil {
		t.Fatalf("ParseCondition() error: %v", err)
	}
	cmp, ok := expr.(ir.CondCompare)
	if !ok {
		t.Fatalf("expected CondCompare, got %T", expr)
	}
	if cmp.Op != "==" {
		t.Errorf("Op = %q, want ==", cmp.Op)
	}
}

func TestParseCondition_NotEquals(t *testing.T) {
	expr, err := ParseCondition("ctx.x != value")
	if err != nil {
		t.Fatalf("ParseCondition() error: %v", err)
	}
	cmp, ok := expr.(ir.CondCompare)
	if !ok {
		t.Fatalf("expected CondCompare, got %T", expr)
	}
	if cmp.Op != "!=" {
		t.Errorf("Op = %q, want !=", cmp.Op)
	}
}

func TestParseCondition_QuotedValue(t *testing.T) {
	expr, err := ParseCondition(`ctx.x = "hello world"`)
	if err != nil {
		t.Fatalf("ParseCondition() error: %v", err)
	}
	cmp, ok := expr.(ir.CondCompare)
	if !ok {
		t.Fatalf("expected CondCompare, got %T", expr)
	}
	if cmp.Value != "hello world" {
		t.Errorf("Value = %q, want %q", cmp.Value, "hello world")
	}
}

func TestParseCondition_SingleQuotedValue(t *testing.T) {
	expr, err := ParseCondition("ctx.x = 'hello world'")
	if err != nil {
		t.Fatalf("ParseCondition() error: %v", err)
	}
	cmp, ok := expr.(ir.CondCompare)
	if !ok {
		t.Fatalf("expected CondCompare, got %T", expr)
	}
	if cmp.Value != "hello world" {
		t.Errorf("Value = %q, want %q", cmp.Value, "hello world")
	}
}

// --- Condition evaluation edge cases ---

func TestEvalCondition_StartsWith(t *testing.T) {
	ResetRunCounter()
	w := &ir.Workflow{
		Name:  "StartsWith",
		Start: "Check",
		Exit:  "Done",
		Nodes: []*ir.Node{
			{ID: "Check", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "Check."}},
			{ID: "Match", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "Match."}},
			{ID: "NoMatch", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "No."}},
			{ID: "Done", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "Done."}},
		},
		Edges: []*ir.Edge{
			{From: "Check", To: "Match", Condition: &ir.Condition{
				Raw: "ctx.value startswith hello",
			}},
			{From: "Check", To: "NoMatch"},
			{From: "Match", To: "Done"},
			{From: "NoMatch", To: "Done"},
		},
	}

	res, err := Run(w, Options{Scenario: map[string]string{"value": "hello world"}})
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}
	assertPathContains(t, res.Path, "Match")
}

func TestEvalCondition_EndsWith(t *testing.T) {
	ResetRunCounter()
	w := &ir.Workflow{
		Name:  "EndsWith",
		Start: "Check",
		Exit:  "Done",
		Nodes: []*ir.Node{
			{ID: "Check", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "Check."}},
			{ID: "Match", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "Match."}},
			{ID: "NoMatch", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "No."}},
			{ID: "Done", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "Done."}},
		},
		Edges: []*ir.Edge{
			{From: "Check", To: "Match", Condition: &ir.Condition{
				Raw: "ctx.value endswith world",
			}},
			{From: "Check", To: "NoMatch"},
			{From: "Match", To: "Done"},
			{From: "NoMatch", To: "Done"},
		},
	}

	res, err := Run(w, Options{Scenario: map[string]string{"value": "hello world"}})
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}
	assertPathContains(t, res.Path, "Match")
}

func TestEvalCondition_Contains(t *testing.T) {
	ResetRunCounter()
	w := &ir.Workflow{
		Name:  "Contains",
		Start: "Check",
		Exit:  "Done",
		Nodes: []*ir.Node{
			{ID: "Check", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "Check."}},
			{ID: "Match", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "Match."}},
			{ID: "NoMatch", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "No."}},
			{ID: "Done", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "Done."}},
		},
		Edges: []*ir.Edge{
			{From: "Check", To: "Match", Condition: &ir.Condition{
				Raw: "ctx.value contains middle",
			}},
			{From: "Check", To: "NoMatch"},
			{From: "Match", To: "Done"},
			{From: "NoMatch", To: "Done"},
		},
	}

	res, err := Run(w, Options{Scenario: map[string]string{"value": "start middle end"}})
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}
	assertPathContains(t, res.Path, "Match")
}

// --- EmitJSONL edge case ---

func TestEmitJSONL_Empty(t *testing.T) {
	var buf bytes.Buffer
	if err := EmitJSONL(&buf, nil); err != nil {
		t.Fatalf("EmitJSONL(nil) error: %v", err)
	}
	if buf.Len() != 0 {
		t.Errorf("expected empty output for nil events, got %q", buf.String())
	}
}

// --- Interactive edge cases ---

func TestRunHumanInteractive_EOF(t *testing.T) {
	ResetRunCounter()
	w := humanWorkflow()

	// Empty reader → EOF, should return default value.
	input := strings.NewReader("")
	res, err := Run(w, Options{
		Interactive: true,
		Stdin:       input,
	})
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}
	if res.Status != "success" {
		t.Errorf("Status = %q, want success", res.Status)
	}
}

func TestRunHumanInteractive_ChoiceMode(t *testing.T) {
	ResetRunCounter()
	w := &ir.Workflow{
		Name:  "HumanChoice",
		Start: "Start",
		Exit:  "Done",
		Nodes: []*ir.Node{
			{ID: "Start", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "Begin."}},
			{ID: "Ask", Kind: ir.NodeHuman, Label: "", Config: ir.HumanConfig{
				Mode:    "choice",
				Default: "approve",
			}},
			{ID: "Done", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "Done."}},
		},
		Edges: []*ir.Edge{
			{From: "Start", To: "Ask"},
			{From: "Ask", To: "Done"},
		},
	}

	input := strings.NewReader("reject\n")
	var stderr bytes.Buffer
	res, err := Run(w, Options{
		Interactive: true,
		Stdin:       input,
		Stderr:      &stderr,
	})
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}
	if res.Status != "success" {
		t.Errorf("Status = %q, want success", res.Status)
	}
	// Choice mode should say "Enter choice:".
	if !strings.Contains(stderr.String(), "Enter choice") {
		t.Errorf("stderr = %q, expected 'Enter choice'", stderr.String())
	}
}

// --- EnsureConditionsParsed edge cases ---

func TestEnsureConditionsParsed_AlreadyParsed(t *testing.T) {
	w := &ir.Workflow{
		Edges: []*ir.Edge{
			{From: "A", To: "B", Condition: &ir.Condition{
				Raw: "ctx.x = y",
			}},
		},
	}
	err := EnsureConditionsParsed(w)
	if err != nil {
		t.Fatalf("EnsureConditionsParsed() error: %v", err)
	}
}

func TestEnsureConditionsParsed_NilCondition(t *testing.T) {
	w := &ir.Workflow{
		Edges: []*ir.Edge{
			{From: "A", To: "B"},
		},
	}
	err := EnsureConditionsParsed(w)
	if err != nil {
		t.Fatalf("EnsureConditionsParsed() error: %v", err)
	}
}

// --- Fidelity resolution ---

func TestResolveFidelity(t *testing.T) {
	ResetRunCounter()
	w := &ir.Workflow{
		Name:  "Fidelity",
		Start: "A",
		Exit:  "A",
		Defaults: ir.WorkflowDefaults{
			Fidelity: "high",
		},
		Nodes: []*ir.Node{
			{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{
				Prompt:   "Test.",
				Fidelity: "", // Should use default.
			}},
		},
	}

	res, err := Run(w, Options{})
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}
	for _, ev := range res.Events {
		if ne, ok := ev.(event.NodeEnter); ok && ne.Node == "A" {
			if ne.Fidelity != "high" {
				t.Errorf("Fidelity = %q, want high", ne.Fidelity)
			}
		}
	}
	_ = res
}

func TestNodeFidelityOverridesDefault(t *testing.T) {
	ResetRunCounter()
	w := &ir.Workflow{
		Name:  "FidelityOverride",
		Start: "A",
		Exit:  "A",
		Defaults: ir.WorkflowDefaults{
			Fidelity: "high",
		},
		Nodes: []*ir.Node{
			{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{
				Prompt:   "Test.",
				Fidelity: "low",
			}},
		},
	}

	res, err := Run(w, Options{})
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}
	for _, ev := range res.Events {
		if ne, ok := ev.(event.NodeEnter); ok && ne.Node == "A" {
			if ne.Fidelity != "low" {
				t.Errorf("Fidelity = %q, want low", ne.Fidelity)
			}
		}
	}
}

// --- Dead-end single run ---

func TestRun_DeadEnd(t *testing.T) {
	ResetRunCounter()
	w := &ir.Workflow{
		Name:  "DeadEnd",
		Start: "Start",
		Exit:  "Done",
		Nodes: []*ir.Node{
			{ID: "Start", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "Go."}},
			{ID: "Sink", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "Stuck."}},
			{ID: "Done", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "Done."}},
		},
		Edges: []*ir.Edge{
			{From: "Start", To: "Sink"},
		},
	}

	res, err := Run(w, Options{})
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}
	if res.Status != "dead_end" {
		t.Errorf("Status = %q, want dead_end", res.Status)
	}
}

// --- Exit node not found ---

func TestRunExitNodeNotFound(t *testing.T) {
	ResetRunCounter()
	w := &ir.Workflow{
		Name:  "bad",
		Start: "Start",
		Exit:  "Ghost",
		Nodes: []*ir.Node{
			{ID: "Start", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "go"}},
		},
	}
	_, err := Run(w, Options{})
	if err == nil {
		t.Fatal("expected error for exit node not found")
	}
}

// --- seedCompareContext edge cases via RunAllPaths ---

func TestRunAllPaths_NotCondition(t *testing.T) {
	ResetRunCounter()
	w := &ir.Workflow{
		Name:  "NotCond",
		Start: "A",
		Exit:  "C",
		Nodes: []*ir.Node{
			{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "A."}},
			{ID: "B", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "B."}},
			{ID: "C", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "C."}},
		},
		Edges: []*ir.Edge{
			{From: "A", To: "B", Condition: &ir.Condition{
				Raw: "not ctx.done = true",
			}},
			{From: "A", To: "C"},
			{From: "B", To: "C"},
		},
	}

	results, err := RunAllPaths(w, Options{})
	if err != nil {
		t.Fatalf("RunAllPaths() error: %v", err)
	}
	if len(results) < 2 {
		t.Errorf("RunAllPaths found %d paths, want at least 2", len(results))
	}
}

// --- Issue 4: Tool defaults mask fallback edges ---

func TestToolDefaultClearing_FallbackEdge(t *testing.T) {
	ResetRunCounter()
	w := &ir.Workflow{
		Name:  "ToolFallback",
		Start: "T",
		Exit:  "Done",
		Nodes: []*ir.Node{
			{ID: "T", Kind: ir.NodeTool, Config: ir.ToolConfig{Command: "echo test"}},
			{ID: "Special", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "Special."}},
			{ID: "Fallback", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "Fallback."}},
			{ID: "Done", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "Done."}},
		},
		Edges: []*ir.Edge{
			{From: "T", To: "Special", Condition: &ir.Condition{
				Raw: "ctx.tool_stdout = has_data",
			}},
			{From: "T", To: "Fallback"}, // unconditional fallback
			{From: "Special", To: "Done"},
			{From: "Fallback", To: "Done"},
		},
	}

	// Without clearing: tool auto-seeds tool_stdout=success, condition doesn't match
	// has_data, so fallback is taken. But outcome=success is also seeded.
	// This test verifies that setting tool_stdout="" suppresses the default
	// so no conditional matches and the fallback fires.
	res, err := Run(w, Options{Scenario: map[string]string{"T.tool_stdout": ""}})
	if err != nil {
		t.Fatalf("Run error: %v", err)
	}
	assertPathContains(t, res.Path, "Fallback")
	assertPathNotContains(t, res.Path, "Special")
}

func TestToolDefaultClearing_GlobalEmpty(t *testing.T) {
	ResetRunCounter()
	w := &ir.Workflow{
		Name:  "ToolGlobalClear",
		Start: "T",
		Exit:  "Done",
		Nodes: []*ir.Node{
			{ID: "T", Kind: ir.NodeTool, Config: ir.ToolConfig{Command: "echo test"}},
			{ID: "Match", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "Match."}},
			{ID: "Done", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "Done."}},
		},
		Edges: []*ir.Edge{
			{From: "T", To: "Match", Condition: &ir.Condition{
				Raw: "ctx.outcome = success",
			}},
			{From: "T", To: "Done"}, // fallback
			{From: "Match", To: "Done"},
		},
	}

	// Global outcome="" should suppress the tool's outcome=success default.
	res, err := Run(w, Options{Scenario: map[string]string{"outcome": ""}})
	if err != nil {
		t.Fatalf("Run error: %v", err)
	}
	// outcome is empty, condition doesn't match, fallback to Done.
	assertPathNotContains(t, res.Path, "Match")
}

// --- Issue 1: preferred_label for human gate edge selection ---

func TestPreferredLabel_MatchesEdge(t *testing.T) {
	ResetRunCounter()
	w := &ir.Workflow{
		Name:  "HumanLabel",
		Start: "Gate",
		Exit:  "Done",
		Nodes: []*ir.Node{
			{ID: "Gate", Kind: ir.NodeHuman, Config: ir.HumanConfig{Mode: "choice"}},
			{ID: "YesPath", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "Yes."}},
			{ID: "NoPath", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "No."}},
			{ID: "Done", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "Done."}},
		},
		Edges: []*ir.Edge{
			{From: "Gate", To: "YesPath", Label: "[Y] Yes"},
			{From: "Gate", To: "NoPath", Label: "[N] No"},
			{From: "YesPath", To: "Done"},
			{From: "NoPath", To: "Done"},
		},
	}

	// preferred_label="no" should match "[N] No" (case-insensitive substring).
	res, err := Run(w, Options{Scenario: map[string]string{"preferred_label": "no"}})
	if err != nil {
		t.Fatalf("Run error: %v", err)
	}
	assertPathContains(t, res.Path, "NoPath")
	assertPathNotContains(t, res.Path, "YesPath")
}

func TestPreferredLabel_PerNode(t *testing.T) {
	ResetRunCounter()
	w := &ir.Workflow{
		Name:  "PerNodeLabel",
		Start: "Gate",
		Exit:  "Done",
		Nodes: []*ir.Node{
			{ID: "Gate", Kind: ir.NodeHuman, Config: ir.HumanConfig{Mode: "choice"}},
			{ID: "Approve", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "Approved."}},
			{ID: "Reject", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "Rejected."}},
			{ID: "Done", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "Done."}},
		},
		Edges: []*ir.Edge{
			{From: "Gate", To: "Approve", Label: "[A] Approve"},
			{From: "Gate", To: "Reject", Label: "[R] Reject"},
			{From: "Approve", To: "Done"},
			{From: "Reject", To: "Done"},
		},
	}

	// Per-node preferred_label using NodeID.preferred_label syntax.
	res, err := Run(w, Options{Scenario: map[string]string{"Gate.preferred_label": "reject"}})
	if err != nil {
		t.Fatalf("Run error: %v", err)
	}
	assertPathContains(t, res.Path, "Reject")
	assertPathNotContains(t, res.Path, "Approve")
}

func TestPreferredLabel_NoMatch_TakesFirst(t *testing.T) {
	ResetRunCounter()
	w := &ir.Workflow{
		Name:  "LabelNoMatch",
		Start: "Gate",
		Exit:  "Done",
		Nodes: []*ir.Node{
			{ID: "Gate", Kind: ir.NodeHuman, Config: ir.HumanConfig{Mode: "choice"}},
			{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "A."}},
			{ID: "B", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "B."}},
			{ID: "Done", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "Done."}},
		},
		Edges: []*ir.Edge{
			{From: "Gate", To: "A", Label: "Option A"},
			{From: "Gate", To: "B", Label: "Option B"},
			{From: "A", To: "Done"},
			{From: "B", To: "Done"},
		},
	}

	// preferred_label doesn't match any edge — falls through to first edge.
	res, err := Run(w, Options{Scenario: map[string]string{"preferred_label": "nonexistent"}})
	if err != nil {
		t.Fatalf("Run error: %v", err)
	}
	assertPathContains(t, res.Path, "A")
}

func TestPreferredLabel_ClearedAfterUse(t *testing.T) {
	ResetRunCounter()
	w := &ir.Workflow{
		Name:  "LabelCleared",
		Start: "Gate1",
		Exit:  "Done",
		Nodes: []*ir.Node{
			{ID: "Gate1", Kind: ir.NodeHuman, Config: ir.HumanConfig{Mode: "choice"}},
			{ID: "Mid", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "Mid."}},
			{ID: "Gate2", Kind: ir.NodeHuman, Config: ir.HumanConfig{Mode: "choice"}},
			{ID: "X", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "X."}},
			{ID: "Y", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "Y."}},
			{ID: "Done", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "Done."}},
		},
		Edges: []*ir.Edge{
			{From: "Gate1", To: "Mid", Label: "[A] Alpha"},
			{From: "Mid", To: "Gate2"},
			{From: "Gate2", To: "X", Label: "[X] First"},
			{From: "Gate2", To: "Y", Label: "[Y] Second"},
			{From: "X", To: "Done"},
			{From: "Y", To: "Done"},
		},
	}

	// preferred_label="alpha" affects Gate1 but should NOT leak to Gate2.
	// Gate2 should take the first edge (X) since preferred_label was cleared.
	res, err := Run(w, Options{Scenario: map[string]string{"preferred_label": "alpha"}})
	if err != nil {
		t.Fatalf("Run error: %v", err)
	}
	assertPathContains(t, res.Path, "Mid")
	assertPathContains(t, res.Path, "X") // Gate2 takes first edge, not leaking "alpha"
}

// --- Helpers ---

func assertEventSequence(t *testing.T, events []event.Event, expected []event.Type) {
	t.Helper()
	if len(events) != len(expected) {
		types := make([]event.Type, len(events))
		for i, e := range events {
			types[i] = e.EventType()
		}
		t.Fatalf("event count = %d, want %d\ngot:  %v\nwant: %v", len(events), len(expected), types, expected)
	}
	for i, want := range expected {
		if got := events[i].EventType(); got != want {
			t.Errorf("event[%d] = %q, want %q", i, got, want)
		}
	}
}

func assertPathContains(t *testing.T, path []string, nodeID string) {
	t.Helper()
	for _, n := range path {
		if n == nodeID {
			return
		}
	}
	t.Errorf("path %v does not contain %q", path, nodeID)
}

func assertPathNotContains(t *testing.T, path []string, nodeID string) {
	t.Helper()
	for _, n := range path {
		if n == nodeID {
			t.Errorf("path %v should not contain %q", path, nodeID)
			return
		}
	}
}

func interviewWorkflow() *ir.Workflow {
	return &ir.Workflow{
		Name:  "InterviewTest",
		Start: "Start",
		Exit:  "Done",
		Nodes: []*ir.Node{
			{ID: "Start", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "Ask questions."}},
			{ID: "Ask", Kind: ir.NodeHuman, Label: "Answer questions", Config: ir.HumanConfig{
				Mode:         "interview",
				QuestionsKey: "questions",
				AnswersKey:   "answers",
			}},
			{ID: "Done", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "Done."}},
		},
		Edges: []*ir.Edge{
			{From: "Start", To: "Ask"},
			{From: "Ask", To: "Done"},
		},
	}
}

func TestRunHumanInteractive_InterviewMode(t *testing.T) {
	ResetRunCounter()
	w := interviewWorkflow()
	input := strings.NewReader("answer one\nanswer two\n\n")
	var stderr bytes.Buffer
	res, err := Run(w, Options{
		Interactive: true,
		Stdin:       input,
		Stderr:      &stderr,
	})
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}
	if res.Status != "success" {
		t.Errorf("Status = %q, want success", res.Status)
	}
	if !strings.Contains(stderr.String(), "interview") {
		t.Error("expected 'interview' in stderr prompt")
	}
	// Verify answers stored in custom key.
	var foundAnswers bool
	for _, ev := range res.Events {
		if cu, ok := ev.(event.ContextUpdate); ok && cu.Key == "answers" {
			foundAnswers = true
			if !strings.Contains(cu.Value, "answer one") {
				t.Errorf("answers missing 'answer one': %s", cu.Value)
			}
		}
	}
	if !foundAnswers {
		t.Error("expected context update for answers key")
	}
}

func TestRunHumanAutoSuccess_InterviewMode(t *testing.T) {
	w := interviewWorkflow()
	res, err := Run(w, Options{})
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}
	if res.Status != "success" {
		t.Errorf("non-interactive interview should auto-succeed, got %q", res.Status)
	}
}

func TestRunHumanInteractive_InterviewEOF(t *testing.T) {
	w := interviewWorkflow()
	input := strings.NewReader("")
	res, err := Run(w, Options{
		Interactive: true,
		Stdin:       input,
	})
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}
	if res.Status != "success" {
		t.Errorf("EOF on interview should succeed, got %q", res.Status)
	}
}
