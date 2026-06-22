package simulate

import (
	"bytes"
	"strings"
	"testing"

	"github.com/2389-research/dippin-lang/event"
	"github.com/2389-research/dippin-lang/ir"
)

// --- ParseCondition: unexpected trailing token message ---

func TestParseCondition_UnexpectedTokenMsg(t *testing.T) {
	_, err := ParseCondition("ctx.x = a garbage")
	if err == nil {
		t.Fatal("expected error for trailing token, got nil")
	}
	if !strings.Contains(err.Error(), "unexpected token") {
		t.Errorf("error = %q, want 'unexpected token'", err.Error())
	}
}

// --- ParseCondition: incomplete or/and/not ---

func TestParseCondition_IncompleteOr(t *testing.T) {
	_, err := ParseCondition("ctx.x = a or")
	if err == nil {
		t.Fatal("expected error for incomplete or expression")
	}
}

func TestParseCondition_IncompleteAnd(t *testing.T) {
	_, err := ParseCondition("ctx.x = a and")
	if err == nil {
		t.Fatal("expected error for incomplete and expression")
	}
}

func TestParseCondition_BareNot(t *testing.T) {
	_, err := ParseCondition("not")
	if err == nil {
		t.Fatal("expected error for bare 'not'")
	}
}

// --- evalCompare: unknown operator returns false ---

func TestEvalCompare_UnknownOp(t *testing.T) {
	s := &simulator{ctx: map[string]string{"x": "test"}}
	result := s.evalCompare(ir.CondCompare{Variable: "x", Op: "bogus", Value: "test"})
	if result {
		t.Error("expected false for unknown operator")
	}
}

// --- evalCompare: 'in' operator ---

func TestEvalCompare_InOperatorMatch(t *testing.T) {
	s := &simulator{ctx: map[string]string{"x": "b"}}
	if !s.evalCompare(ir.CondCompare{Variable: "x", Op: "in", Value: "a,b,c"}) {
		t.Error("expected true for 'in' match")
	}
}

func TestEvalCompare_InOperatorNoMatch(t *testing.T) {
	s := &simulator{ctx: map[string]string{"x": "z"}}
	if s.evalCompare(ir.CondCompare{Variable: "x", Op: "in", Value: "a,b,c"}) {
		t.Error("expected false for 'in' non-match")
	}
}

// --- resolveVariable: unknown namespace ---

func TestResolveVariable_UnknownNamespace(t *testing.T) {
	s := &simulator{
		workflow: &ir.Workflow{},
		ctx:      make(map[string]string),
	}
	val := s.resolveVariable("unknown.var")
	if val != "" {
		t.Errorf("expected empty, got %q", val)
	}
}

// --- resolveFidelity ---

func TestResolveFidelity_Defaults(t *testing.T) {
	if f := resolveFidelity("", "full"); f != "full" {
		t.Errorf("expected 'full', got %q", f)
	}
	if f := resolveFidelity("compact", "full"); f != "compact" {
		t.Errorf("expected 'compact', got %q", f)
	}
}

// --- stepNode: missing node ---

func TestStepNode_MissingNodeError(t *testing.T) {
	ResetRunCounter()
	w := minimalWorkflow()
	s := &simulator{
		workflow:   w,
		opts:       Options{},
		ctx:        make(map[string]string),
		visited:    make(map[string]bool),
		nodeVisits: make(map[string]int),
	}
	_, _, err := s.stepNode("NonExistent")
	if err == nil {
		t.Fatal("expected error for missing node")
	}
}

// --- advanceToNext: dead end ---

func TestAdvanceToNext_DeadEndResult(t *testing.T) {
	ResetRunCounter()
	w := &ir.Workflow{
		Name:  "DeadEnd2",
		Start: "A",
		Exit:  "B",
		Nodes: []*ir.Node{
			{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "A."}},
			{ID: "B", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "B."}},
		},
		Edges: []*ir.Edge{},
	}
	res, err := Run(w, Options{})
	if err != nil {
		t.Fatalf("Run error: %v", err)
	}
	if res.Status != "dead_end" {
		t.Errorf("Status = %q, want dead_end", res.Status)
	}
}

// --- forceLoopExit: unconditional fallback ---

func TestForceLoopExit_UnconditionalEdge(t *testing.T) {
	ResetRunCounter()
	w := &ir.Workflow{
		Name:  "LoopUnc2",
		Start: "A",
		Exit:  "B",
		Nodes: []*ir.Node{
			{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "A.", AutoStatus: true}},
			{ID: "B", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "B."}},
		},
		Edges: []*ir.Edge{
			{From: "A", To: "A", Condition: &ir.Condition{
				Raw: "ctx.outcome = success",
			}},
			{From: "A", To: "B"},
		},
	}
	res, err := Run(w, Options{MaxNodeVisits: 1})
	if err != nil {
		t.Fatalf("Run error: %v", err)
	}
	if res.Status != "success" {
		t.Errorf("Status = %q, want success", res.Status)
	}
}

// --- forceLoopExit: all conds match, falls to first edge ---

func TestForceLoopExit_AllCondMatchFallback(t *testing.T) {
	ResetRunCounter()
	w := &ir.Workflow{
		Name:  "AllMatch2",
		Start: "A",
		Exit:  "B",
		Nodes: []*ir.Node{
			{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "A.", AutoStatus: true}},
			{ID: "B", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "B."}},
		},
		Edges: []*ir.Edge{
			{From: "A", To: "B", Condition: &ir.Condition{
				Raw: "ctx.outcome = success",
			}},
			{From: "A", To: "A", Condition: &ir.Condition{
				Raw: "ctx.outcome = success",
			}},
		},
	}
	res, err := Run(w, Options{MaxNodeVisits: 1})
	if err != nil {
		t.Fatalf("Run error: %v", err)
	}
	if res.Status != "success" {
		t.Errorf("Status = %q, want success", res.Status)
	}
}

// --- firstUnconditional: all conditional ---

func TestFirstUnconditional_AllConditional(t *testing.T) {
	s := &simulator{ctx: make(map[string]string)}
	edges := []*ir.Edge{
		{From: "A", To: "B", Condition: &ir.Condition{
			Raw: "ctx.x = y",
		}},
	}
	e := s.firstUnconditional(edges)
	if e != nil {
		t.Error("expected nil for all-conditional edges")
	}
}

// --- Human interactive: EOF with default ---

func TestHumanInteractive_EOFWithDefault(t *testing.T) {
	ResetRunCounter()
	w := &ir.Workflow{
		Name:  "HumanEOFDef",
		Start: "H",
		Exit:  "Done",
		Nodes: []*ir.Node{
			{ID: "H", Kind: ir.NodeHuman, Label: "Ask", Config: ir.HumanConfig{
				Mode: "choice", Default: "approve",
			}},
			{ID: "Done", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "End."}},
		},
		Edges: []*ir.Edge{{From: "H", To: "Done"}},
	}
	input := strings.NewReader("")
	var stderr bytes.Buffer
	res, err := Run(w, Options{Interactive: true, Stdin: input, Stderr: &stderr})
	if err != nil {
		t.Fatalf("Run error: %v", err)
	}
	var found bool
	for _, ev := range res.Events {
		if cu, ok := ev.(event.ContextUpdate); ok && cu.Key == "human_response" {
			if cu.Value != "approve" {
				t.Errorf("human_response = %q, want approve", cu.Value)
			}
			found = true
		}
	}
	if !found {
		t.Error("expected context_update for human_response")
	}
}

// --- Interactive prompt with nil stderr ---

func TestHumanInteractive_NilStderr(t *testing.T) {
	ResetRunCounter()
	w := &ir.Workflow{
		Name:  "HumanNoStderr2",
		Start: "H",
		Exit:  "Done",
		Nodes: []*ir.Node{
			{ID: "H", Kind: ir.NodeHuman, Label: "", Config: ir.HumanConfig{Mode: "choice"}},
			{ID: "Done", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "End."}},
		},
		Edges: []*ir.Edge{{From: "H", To: "Done"}},
	}
	input := strings.NewReader("yes\n")
	res, err := Run(w, Options{Interactive: true, Stdin: input, Stderr: nil})
	if err != nil {
		t.Fatalf("Run error: %v", err)
	}
	if res.Status != "success" {
		t.Errorf("Status = %q, want success", res.Status)
	}
}

// --- RunAllPaths: no exit ---

func TestRunAllPaths_NoExit(t *testing.T) {
	ResetRunCounter()
	w := &ir.Workflow{Name: "bad2", Start: "A"}
	_, err := RunAllPaths(w, nil)
	if err == nil {
		t.Fatal("expected error for missing exit node")
	}
}

func TestRunAllPaths_BadCondition(t *testing.T) {
	ResetRunCounter()
	w := &ir.Workflow{
		Name:  "BadCond2",
		Start: "A",
		Exit:  "B",
		Nodes: []*ir.Node{
			{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "A."}},
			{ID: "B", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "B."}},
		},
		Edges: []*ir.Edge{
			{From: "A", To: "B", Condition: &ir.Condition{Raw: "bad missing op"}},
		},
	}
	_, err := RunAllPaths(w, nil)
	if err == nil {
		t.Fatal("expected error for invalid condition")
	}
}

// --- seedConditionContext ---

func TestSeedConditionContext_OrSeedsLeft(t *testing.T) {
	ctx := make(map[string]string)
	seedConditionContext(ir.CondOr{
		Left:  ir.CondCompare{Variable: "ctx.x", Op: "=", Value: "a"},
		Right: ir.CondCompare{Variable: "ctx.y", Op: "=", Value: "b"},
	}, ctx)
	if ctx["x"] != "a" {
		t.Errorf("ctx[x] = %q, want 'a'", ctx["x"])
	}
}

func TestSeedConditionContext_AndBoth(t *testing.T) {
	ctx := make(map[string]string)
	seedConditionContext(ir.CondAnd{
		Left:  ir.CondCompare{Variable: "ctx.x", Op: "=", Value: "a"},
		Right: ir.CondCompare{Variable: "ctx.y", Op: "=", Value: "b"},
	}, ctx)
	if ctx["x"] != "a" || ctx["y"] != "b" {
		t.Errorf("ctx = %v", ctx)
	}
}

func TestSeedConditionContext_NotSkips(t *testing.T) {
	ctx := make(map[string]string)
	seedConditionContext(ir.CondNot{Inner: ir.CondCompare{Variable: "ctx.x", Op: "=", Value: "a"}}, ctx)
	if _, ok := ctx["x"]; ok {
		t.Error("NOT should not seed context")
	}
}

func TestSeedConditionContext_NilSafe(t *testing.T) {
	ctx := make(map[string]string)
	seedConditionContext(nil, ctx)
	if len(ctx) != 0 {
		t.Error("nil expr should not seed context")
	}
}

func TestSeedCompareContext_InFirst(t *testing.T) {
	ctx := make(map[string]string)
	seedCompareContext(ir.CondCompare{Variable: "ctx.x", Op: "in", Value: "a, b, c"}, ctx)
	if ctx["x"] != "a" {
		t.Errorf("ctx[x] = %q, want 'a'", ctx["x"])
	}
}

func TestSeedCompareContext_ContainsSkips(t *testing.T) {
	ctx := make(map[string]string)
	seedCompareContext(ir.CondCompare{Variable: "ctx.x", Op: "contains", Value: "test"}, ctx)
	if _, ok := ctx["x"]; ok {
		t.Error("contains should not seed context")
	}
}

func TestSeedCompareContext_BareVar(t *testing.T) {
	ctx := make(map[string]string)
	seedCompareContext(ir.CondCompare{Variable: "outcome", Op: "=", Value: "done"}, ctx)
	if ctx["outcome"] != "done" {
		t.Errorf("ctx[outcome] = %q, want 'done'", ctx["outcome"])
	}
}

// --- appendNodeEvents ---

func TestAppendNodeEvents_ParallelNode(t *testing.T) {
	pe := &pathEnumerator{workflow: &ir.Workflow{Name: "Test"}}
	node := &ir.Node{ID: "P", Kind: ir.NodeParallel, Config: ir.ParallelConfig{Targets: []string{"A", "B"}}}
	events := pe.appendNodeEvents(node, nil)
	if len(events) != 3 {
		t.Fatalf("events = %d, want 3", len(events))
	}
}

func TestAppendNodeEvents_FanInNode(t *testing.T) {
	pe := &pathEnumerator{workflow: &ir.Workflow{Name: "Test"}}
	node := &ir.Node{ID: "J", Kind: ir.NodeFanIn, Config: ir.FanInConfig{Sources: []string{"A", "B"}}}
	events := pe.appendNodeEvents(node, nil)
	if len(events) != 3 {
		t.Fatalf("events = %d, want 3", len(events))
	}
}

// --- assignRunID ---

func TestAssignRunID_NoPipelineStartEvt(t *testing.T) {
	events := []event.Event{event.NodeEnter{Event: event.TypeNodeEnter, Node: "A"}}
	result := assignRunID(events, "test-id")
	if len(result) != 1 {
		t.Fatalf("events = %d, want 1", len(result))
	}
}

// --- shouldExplore ---

func TestShouldExplore_Limits(t *testing.T) {
	pe := &pathEnumerator{results: make([]*Result, maxPathResults)}
	if pe.shouldExplore(&pathState{nodeID: "A", depth: 0, visited: make(map[string]int)}) {
		t.Error("should not explore when maxPathResults reached")
	}

	pe2 := &pathEnumerator{}
	if pe2.shouldExplore(&pathState{nodeID: "A", depth: maxPathDepth + 1, visited: make(map[string]int)}) {
		t.Error("should not explore when maxPathDepth exceeded")
	}

	pe3 := &pathEnumerator{}
	if pe3.shouldExplore(&pathState{nodeID: "A", depth: 0, visited: map[string]int{"A": 2}}) {
		t.Error("should not explore when node visited 2 times")
	}
}

// --- Per-node scenario ---

func TestPerNodeScenario_Override(t *testing.T) {
	ResetRunCounter()
	w := conditionalWorkflow()
	res, err := Run(w, Options{Scenario: map[string]string{"Check.outcome": "fail"}})
	if err != nil {
		t.Fatalf("Run error: %v", err)
	}
	assertPathContains(t, res.Path, "PathB")
}

// --- validateSimInput: exit not found ---

func TestValidateSimInput_ExitNotFoundError(t *testing.T) {
	w := &ir.Workflow{
		Name: "bad3", Start: "A", Exit: "Ghost",
		Nodes: []*ir.Node{{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "begin"}}},
	}
	_, err := Run(w, Options{})
	if err == nil {
		t.Fatal("expected error for exit node not found")
	}
}

// --- resolveConditionalNext: no match falls through ---

func TestResolveConditionalNext_FallbackToFirst(t *testing.T) {
	ResetRunCounter()
	w := &ir.Workflow{
		Name: "NoMatch2", Start: "A", Exit: "C",
		Nodes: []*ir.Node{
			{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "A."}},
			{ID: "B", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "B."}},
			{ID: "C", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "C."}},
		},
		Edges: []*ir.Edge{
			{From: "A", To: "B", Condition: &ir.Condition{
				Raw: "ctx.x = y",
			}},
			{From: "A", To: "C", Condition: &ir.Condition{
				Raw: "ctx.x = z",
			}},
			{From: "B", To: "C"},
		},
	}
	res, err := Run(w, Options{Scenario: map[string]string{"x": "neither"}})
	if err != nil {
		t.Fatalf("Run error: %v", err)
	}
	assertPathContains(t, res.Path, "B")
}

// --- tokenizeCondition ---

func TestTokenizeCondition_OperatorAfterWord(t *testing.T) {
	tokens := tokenizeCondition("ctx.x!=y")
	if len(tokens) != 3 {
		t.Errorf("expected 3 tokens, got %d: %v", len(tokens), tokens)
	}
}

func TestTryTokenizeWordCond_EmptyInput(t *testing.T) {
	tok, n := tryTokenizeWordCond("", 0)
	if n != 0 || tok != "" {
		t.Errorf("expected empty, got %q/%d", tok, n)
	}
}
