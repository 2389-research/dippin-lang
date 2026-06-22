package migrate

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/2389-research/dippin-lang/ir"
)

// ============================================================
// DOT Parser Tests (10 cases)
// ============================================================

func TestParseDOTSimpleDigraph(t *testing.T) {
	input := `digraph G { A -> B; }`
	dg, err := parseDOT(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if dg.Name != "G" {
		t.Errorf("name = %q, want %q", dg.Name, "G")
	}
	if len(dg.Nodes) != 2 {
		t.Errorf("nodes = %d, want 2", len(dg.Nodes))
	}
	if len(dg.Edges) != 1 {
		t.Errorf("edges = %d, want 1", len(dg.Edges))
	}
	if dg.Edges[0].From != "A" || dg.Edges[0].To != "B" {
		t.Errorf("edge = %s->%s, want A->B", dg.Edges[0].From, dg.Edges[0].To)
	}
}

func TestParseDOTNodeWithAttributes(t *testing.T) {
	input := `digraph G {
		A [shape=box, label="My Agent"];
	}`
	dg, err := parseDOT(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(dg.Nodes) != 1 {
		t.Fatalf("nodes = %d, want 1", len(dg.Nodes))
	}
	n := dg.Nodes[0]
	if n.ID != "A" {
		t.Errorf("ID = %q, want %q", n.ID, "A")
	}
	if n.Attrs["shape"] != "box" {
		t.Errorf("shape = %q, want %q", n.Attrs["shape"], "box")
	}
	if n.Attrs["label"] != "My Agent" {
		t.Errorf("label = %q, want %q", n.Attrs["label"], "My Agent")
	}
}

func TestParseDOTEdgeWithAttributes(t *testing.T) {
	input := `digraph G {
		A -> B [label="yes", condition="outcome=success"];
	}`
	dg, err := parseDOT(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(dg.Edges) != 1 {
		t.Fatalf("edges = %d, want 1", len(dg.Edges))
	}
	e := dg.Edges[0]
	if e.Attrs["label"] != "yes" {
		t.Errorf("label = %q, want %q", e.Attrs["label"], "yes")
	}
	if e.Attrs["condition"] != "outcome=success" {
		t.Errorf("condition = %q, want %q", e.Attrs["condition"], "outcome=success")
	}
}

func TestParseDOTGraphAttributes(t *testing.T) {
	input := `digraph G {
		graph [goal="test", rankdir=LR];
	}`
	dg, err := parseDOT(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if dg.GraphAttrs["goal"] != "test" {
		t.Errorf("goal = %q, want %q", dg.GraphAttrs["goal"], "test")
	}
	if dg.GraphAttrs["rankdir"] != "LR" {
		t.Errorf("rankdir = %q, want %q", dg.GraphAttrs["rankdir"], "LR")
	}
}

func TestParseDOTQuotedStringsWithEscapes(t *testing.T) {
	input := `digraph G {
		A [label="line1\nline2\"quoted\""];
	}`
	dg, err := parseDOT(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(dg.Nodes) != 1 {
		t.Fatalf("nodes = %d, want 1", len(dg.Nodes))
	}
	want := "line1\nline2\"quoted\""
	if dg.Nodes[0].Attrs["label"] != want {
		t.Errorf("label = %q, want %q", dg.Nodes[0].Attrs["label"], want)
	}
}

func TestParseDOTComments(t *testing.T) {
	input := `digraph G {
		// This is a line comment
		A [shape=box];
		/* This is a
		   block comment */
		B [shape=hexagon];
		A -> B;
	}`
	dg, err := parseDOT(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(dg.Nodes) != 2 {
		t.Errorf("nodes = %d, want 2", len(dg.Nodes))
	}
	if len(dg.Edges) != 1 {
		t.Errorf("edges = %d, want 1", len(dg.Edges))
	}
}

func TestParseDOTEmptyGraph(t *testing.T) {
	input := `digraph G {}`
	dg, err := parseDOT(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if dg.Name != "G" {
		t.Errorf("name = %q, want %q", dg.Name, "G")
	}
	if len(dg.Nodes) != 0 {
		t.Errorf("nodes = %d, want 0", len(dg.Nodes))
	}
	if len(dg.Edges) != 0 {
		t.Errorf("edges = %d, want 0", len(dg.Edges))
	}
}

func TestParseDOTMultipleEdges(t *testing.T) {
	input := `digraph G { A -> B; A -> C; B -> C; }`
	dg, err := parseDOT(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(dg.Edges) != 3 {
		t.Errorf("edges = %d, want 3", len(dg.Edges))
	}
}

func TestParseDOTMissingSemicolons(t *testing.T) {
	input := `digraph G {
		A [shape=box]
		B [shape=hexagon]
		A -> B
	}`
	dg, err := parseDOT(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(dg.Nodes) != 2 {
		t.Errorf("nodes = %d, want 2", len(dg.Nodes))
	}
	if len(dg.Edges) != 1 {
		t.Errorf("edges = %d, want 1", len(dg.Edges))
	}
}

func TestParseDOTMalformed(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"not a digraph", `graph G { A -> B; }`},
		{"missing closing brace", `digraph G { A -> B;`},
		{"missing opening brace", `digraph G A -> B; }`},
		{"empty string", ``},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseDOT(tt.input)
			if err == nil {
				t.Error("expected error, got nil")
			}
		})
	}
}

// ============================================================
// Migration Tests (17 cases)
// ============================================================

func TestMigrateShapeToKindMapping(t *testing.T) {
	tests := []struct {
		shape    string
		wantKind ir.NodeKind
	}{
		{"box", ir.NodeAgent},
		{"hexagon", ir.NodeHuman},
		{"parallelogram", ir.NodeTool},
		{"component", ir.NodeParallel},
		{"tripleoctagon", ir.NodeFanIn},
		{"tab", ir.NodeSubgraph},
		{"Mdiamond", ir.NodeAgent},      // Start marker
		{"Msquare", ir.NodeAgent},       // Exit marker
		{"diamond", ir.NodeConditional}, // Diamond → conditional
		{"", ir.NodeAgent},              // Missing shape → default
	}
	for _, tt := range tests {
		t.Run("shape_"+tt.shape, func(t *testing.T) {
			shapeAttr := ""
			if tt.shape != "" {
				shapeAttr = `, shape=` + tt.shape
			}
			dot := `digraph G {
				Start [shape=Mdiamond];
				Exit [shape=Msquare];
				N [label="Test"` + shapeAttr + `];
				Start -> N;
				N -> Exit;
			}`
			w, err := Migrate(dot)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			n := w.Node("N")
			if n == nil {
				t.Fatal("node N not found")
			}
			if n.Kind != tt.wantKind {
				t.Errorf("kind = %q, want %q", n.Kind, tt.wantKind)
			}
		})
	}
}

func TestMigrateStartExitIdentification(t *testing.T) {
	dot := `digraph G {
		Start [shape=Mdiamond, label="Start"];
		Exit [shape=Msquare, label="Exit"];
		A [shape=box];
		Start -> A;
		A -> Exit;
	}`
	w, err := Migrate(dot)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if w.Start != "Start" {
		t.Errorf("start = %q, want %q", w.Start, "Start")
	}
	if w.Exit != "Exit" {
		t.Errorf("exit = %q, want %q", w.Exit, "Exit")
	}
	// Start and Exit should exist as nodes.
	if w.Node("Start") == nil {
		t.Error("Start node not found in IR")
	}
	if w.Node("Exit") == nil {
		t.Error("Exit node not found in IR")
	}
}

func TestMigratePromptUnescaping(t *testing.T) {
	dot := `digraph G {
		Start [shape=Mdiamond];
		A [shape=box, prompt="line1\nline2\n\"code\""];
		Exit [shape=Msquare];
		Start -> A;
		A -> Exit;
	}`
	w, err := Migrate(dot)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	n := w.Node("A")
	if n == nil {
		t.Fatal("node A not found")
	}
	cfg := n.Config.(ir.AgentConfig)
	want := "line1\nline2\n\"code\""
	if cfg.Prompt != want {
		t.Errorf("prompt = %q, want %q", cfg.Prompt, want)
	}
}

func TestMigrateToolCommandUnescaping(t *testing.T) {
	dot := `digraph G {
		Start [shape=Mdiamond];
		T [shape=parallelogram, tool_command="set -eu\necho hello"];
		Exit [shape=Msquare];
		Start -> T;
		T -> Exit;
	}`
	w, err := Migrate(dot)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	n := w.Node("T")
	if n == nil {
		t.Fatal("node T not found")
	}
	cfg := n.Config.(ir.ToolConfig)
	want := "set -eu\necho hello"
	if cfg.Command != want {
		t.Errorf("command = %q, want %q", cfg.Command, want)
	}
}

func TestMigrateConditionNamespacePrefixing(t *testing.T) {
	tests := []struct {
		name    string
		condRaw string
		wantVar string
	}{
		{"bare outcome", "outcome=success", "ctx.outcome"},
		{"context. prefix", "context.tool_stdout=all_complete", "ctx.tool_stdout"},
		{"ctx. prefix kept", "ctx.outcome=success", "ctx.outcome"},
		{"graph. prefix kept", "graph.goal=done", "graph.goal"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dot := `digraph G {
				Start [shape=Mdiamond];
				A [shape=box];
				B [shape=box];
				Exit [shape=Msquare];
				Start -> A;
				A -> B [condition="` + tt.condRaw + `"];
				B -> Exit;
			}`
			w, err := Migrate(dot)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			edges := w.EdgesFrom("A")
			if len(edges) != 1 {
				t.Fatalf("edges = %d, want 1", len(edges))
			}
			e := edges[0]
			if e.Condition == nil {
				t.Fatal("expected condition")
			}
			cc, ok := e.Condition.Parsed.(ir.CondCompare)
			if !ok {
				t.Fatalf("expected CondCompare, got %T", e.Condition.Parsed)
			}
			if cc.Variable != tt.wantVar {
				t.Errorf("variable = %q, want %q", cc.Variable, tt.wantVar)
			}
		})
	}
}

func TestMigrateComplexCondition(t *testing.T) {
	dot := `digraph G {
		Start [shape=Mdiamond];
		A [shape=box];
		B [shape=box];
		Exit [shape=Msquare];
		Start -> A;
		A -> B [condition="outcome=success && tool_stdout contains pass"];
		B -> Exit;
	}`
	w, err := Migrate(dot)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	edges := w.EdgesFrom("A")
	if len(edges) != 1 {
		t.Fatalf("edges = %d, want 1", len(edges))
	}
	cond := edges[0].Condition
	if cond == nil {
		t.Fatal("expected condition")
	}
	and, ok := cond.Parsed.(ir.CondAnd)
	if !ok {
		t.Fatalf("expected CondAnd, got %T", cond.Parsed)
	}
	left, ok := and.Left.(ir.CondCompare)
	if !ok {
		t.Fatalf("expected CondCompare left, got %T", and.Left)
	}
	if left.Variable != "ctx.outcome" || left.Op != "=" || left.Value != "success" {
		t.Errorf("left = %+v", left)
	}
	right, ok := and.Right.(ir.CondCompare)
	if !ok {
		t.Fatalf("expected CondCompare right, got %T", and.Right)
	}
	if right.Variable != "ctx.tool_stdout" || right.Op != "contains" || right.Value != "pass" {
		t.Errorf("right = %+v", right)
	}
}

func TestMigrateConditionWithNegation(t *testing.T) {
	dot := `digraph G {
		Start [shape=Mdiamond];
		A [shape=box];
		B [shape=box];
		Exit [shape=Msquare];
		Start -> A;
		A -> B [condition="not outcome=fail"];
		B -> Exit;
	}`
	w, err := Migrate(dot)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	edges := w.EdgesFrom("A")
	if len(edges) != 1 {
		t.Fatalf("edges = %d, want 1", len(edges))
	}
	cond := edges[0].Condition
	if cond == nil {
		t.Fatal("expected condition")
	}
	notExpr, ok := cond.Parsed.(ir.CondNot)
	if !ok {
		t.Fatalf("expected CondNot, got %T", cond.Parsed)
	}
	inner, ok := notExpr.Inner.(ir.CondCompare)
	if !ok {
		t.Fatalf("expected CondCompare inner, got %T", notExpr.Inner)
	}
	if inner.Variable != "ctx.outcome" || inner.Op != "=" || inner.Value != "fail" {
		t.Errorf("inner = %+v", inner)
	}
}

func TestMigrateRestartEdge(t *testing.T) {
	tests := []struct {
		name string
		attr string
	}{
		{"restart=true", `restart=true`},
		{"loop_restart=true", `loop_restart=true`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dot := `digraph G {
				Start [shape=Mdiamond];
				A [shape=box];
				B [shape=box];
				Exit [shape=Msquare];
				Start -> A;
				A -> B [` + tt.attr + `];
				B -> Exit;
			}`
			w, err := Migrate(dot)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			edges := w.EdgesFrom("A")
			if len(edges) != 1 {
				t.Fatalf("edges = %d, want 1", len(edges))
			}
			if !edges[0].Restart {
				t.Error("expected restart=true on edge")
			}
		})
	}
}

func TestMigrateOverrideEdge(t *testing.T) {
	dot := `digraph G {
		Start [shape=Mdiamond];
		A [shape=box];
		B [shape=box];
		Exit [shape=Msquare];
		Start -> A;
		A -> B [override=true];
		B -> Exit;
	}`
	w, err := Migrate(dot)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	edges := w.EdgesFrom("A")
	if len(edges) != 1 {
		t.Fatalf("edges = %d, want 1", len(edges))
	}
	if !edges[0].Override {
		t.Error("expected override=true on edge")
	}
}

func TestMigrateChoiceEdge(t *testing.T) {
	dot := `digraph G {
		Start [shape=Mdiamond];
		A [shape=box];
		B [shape=box];
		Exit [shape=Msquare];
		Start -> A;
		A -> B [label="Approve", choice="yes"];
		B -> Exit;
	}`
	w, err := Migrate(dot)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	edges := w.EdgesFrom("A")
	if len(edges) != 1 {
		t.Fatalf("edges = %d, want 1", len(edges))
	}
	if edges[0].Choice != "yes" {
		t.Errorf("expected choice=yes on edge, got %q", edges[0].Choice)
	}
}

func TestMigrateGraphDefaults(t *testing.T) {
	dot := `digraph G {
		graph [goal="Test the system", default_max_retry=3, max_restarts=7, default_fidelity="summary:high"];
		Start [shape=Mdiamond];
		Exit [shape=Msquare];
		Start -> Exit;
	}`
	w, err := Migrate(dot)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if w.Goal != "Test the system" {
		t.Errorf("goal = %q, want %q", w.Goal, "Test the system")
	}
	if w.Defaults.MaxRetries != 3 {
		t.Errorf("max_retries = %d, want 3", w.Defaults.MaxRetries)
	}
	if w.Defaults.MaxRestarts != 7 {
		t.Errorf("max_restarts = %d, want 7", w.Defaults.MaxRestarts)
	}
	if w.Defaults.Fidelity != "summary:high" {
		t.Errorf("fidelity = %q, want %q", w.Defaults.Fidelity, "summary:high")
	}
}

func TestMigrateToolSafetyDefaults(t *testing.T) {
	dot := `digraph G {
		graph [tool_commands_allow="git *,make *", tool_denylist_add="rm -rf /"];
		Start [shape=Mdiamond];
		Exit [shape=Msquare];
		Start -> Exit;
	}`
	w, err := Migrate(dot)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if w.Defaults.ToolCommandsAllow != "git *,make *" {
		t.Errorf("tool_commands_allow = %q, want %q", w.Defaults.ToolCommandsAllow, "git *,make *")
	}
	if w.Defaults.ToolDenylistAdd != "rm -rf /" {
		t.Errorf("tool_denylist_add = %q, want %q", w.Defaults.ToolDenylistAdd, "rm -rf /")
	}
	// Must not leak into Vars.
	if _, ok := w.Vars["tool_commands_allow"]; ok {
		t.Errorf("tool_commands_allow should route to Defaults, not Vars")
	}
	if _, ok := w.Vars["tool_denylist_add"]; ok {
		t.Errorf("tool_denylist_add should route to Defaults, not Vars")
	}
}

func TestMigrateParallelInference(t *testing.T) {
	dot := `digraph G {
		Start [shape=Mdiamond];
		P [shape=component, label="Fan Out"];
		A [shape=box];
		B [shape=box];
		Exit [shape=Msquare];
		Start -> P;
		P -> A;
		P -> B;
		A -> Exit;
		B -> Exit;
	}`
	w, err := Migrate(dot)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	n := w.Node("P")
	if n == nil {
		t.Fatal("node P not found")
	}
	cfg, ok := n.Config.(ir.ParallelConfig)
	if !ok {
		t.Fatalf("expected ParallelConfig, got %T", n.Config)
	}
	if len(cfg.Targets) != 2 {
		t.Fatalf("targets = %d, want 2", len(cfg.Targets))
	}
	// Targets should be A and B (in edge order).
	if cfg.Targets[0] != "A" || cfg.Targets[1] != "B" {
		t.Errorf("targets = %v, want [A B]", cfg.Targets)
	}
}

func TestMigrateFanInInference(t *testing.T) {
	dot := `digraph G {
		Start [shape=Mdiamond];
		A [shape=box];
		B [shape=box];
		J [shape=tripleoctagon, label="Join"];
		Exit [shape=Msquare];
		Start -> A;
		Start -> B;
		A -> J;
		B -> J;
		J -> Exit;
	}`
	w, err := Migrate(dot)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	n := w.Node("J")
	if n == nil {
		t.Fatal("node J not found")
	}
	cfg, ok := n.Config.(ir.FanInConfig)
	if !ok {
		t.Fatalf("expected FanInConfig, got %T", n.Config)
	}
	if len(cfg.Sources) != 2 {
		t.Fatalf("sources = %d, want 2", len(cfg.Sources))
	}
	if cfg.Sources[0] != "A" || cfg.Sources[1] != "B" {
		t.Errorf("sources = %v, want [A B]", cfg.Sources)
	}
}

func TestMigrateDiamondDisambiguation(t *testing.T) {
	tests := []struct {
		name     string
		attrs    string
		wantKind ir.NodeKind
	}{
		{"bare diamond", `shape=diamond, label="Route?"`, ir.NodeConditional},
		{"diamond with prompt", `shape=diamond, prompt="Choose wisely"`, ir.NodeAgent},
		{"diamond with tool_command", `shape=diamond, tool_command="echo test"`, ir.NodeTool},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dot := `digraph G {
				Start [shape=Mdiamond];
				D [` + tt.attrs + `];
				Exit [shape=Msquare];
				Start -> D;
				D -> Exit;
			}`
			w, err := Migrate(dot)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			n := w.Node("D")
			if n == nil {
				t.Fatal("node D not found")
			}
			if n.Kind != tt.wantKind {
				t.Errorf("kind = %q, want %q", n.Kind, tt.wantKind)
			}
		})
	}
}

func TestMigrateEdgeWeight(t *testing.T) {
	dot := `digraph G {
		Start [shape=Mdiamond];
		A [shape=box];
		B [shape=box];
		Exit [shape=Msquare];
		Start -> A;
		A -> B [weight=10];
		B -> Exit;
	}`
	w, err := Migrate(dot)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	edges := w.EdgesFrom("A")
	if len(edges) != 1 {
		t.Fatalf("edges = %d, want 1", len(edges))
	}
	if edges[0].Weight != 10 {
		t.Errorf("weight = %d, want 10", edges[0].Weight)
	}
}

func TestMigrateDurationParsing(t *testing.T) {
	tests := []struct {
		name    string
		timeout string
		wantDur time.Duration
	}{
		{"30s", "30s", 30 * time.Second},
		{"1h30m", "1h30m", 90 * time.Minute},
		{"5m", "5m", 5 * time.Minute},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dot := `digraph G {
				Start [shape=Mdiamond];
				T [shape=parallelogram, tool_command="echo test", timeout="` + tt.timeout + `"];
				Exit [shape=Msquare];
				Start -> T;
				T -> Exit;
			}`
			w, err := Migrate(dot)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			n := w.Node("T")
			if n == nil {
				t.Fatal("node T not found")
			}
			cfg := n.Config.(ir.ToolConfig)
			if cfg.Timeout != tt.wantDur {
				t.Errorf("timeout = %v, want %v", cfg.Timeout, tt.wantDur)
			}
		})
	}
}

func TestMigrateEmptyNodeDefaultsToAgent(t *testing.T) {
	dot := `digraph G {
		Start [shape=Mdiamond];
		N [];
		Exit [shape=Msquare];
		Start -> N;
		N -> Exit;
	}`
	w, err := Migrate(dot)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	n := w.Node("N")
	if n == nil {
		t.Fatal("node N not found")
	}
	if n.Kind != ir.NodeAgent {
		t.Errorf("kind = %q, want %q", n.Kind, ir.NodeAgent)
	}
}

func TestMigrateToSourceRoundTrip(t *testing.T) {
	dot := `digraph test_flow {
		graph [goal="Simple test"];
		Start [shape=Mdiamond, label="Start"];
		Exit [shape=Msquare, label="Exit"];
		Worker [shape=box, label="Worker", prompt="Do the work."];
		Start -> Worker;
		Worker -> Exit [condition="outcome=success"];
	}`
	source, err := MigrateToSource(dot)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Verify it looks like valid .dip source.
	if !strings.HasPrefix(source, "workflow test_flow") {
		t.Errorf("source should start with 'workflow test_flow', got:\n%s", source)
	}
	if !strings.Contains(source, "start: Start") {
		t.Errorf("source should contain 'start: Start', got:\n%s", source)
	}
	if !strings.Contains(source, "exit: Exit") {
		t.Errorf("source should contain 'exit: Exit', got:\n%s", source)
	}
	if !strings.Contains(source, "Do the work.") {
		t.Errorf("source should contain prompt text, got:\n%s", source)
	}
	if !strings.Contains(source, "edges") {
		t.Errorf("source should contain edges section, got:\n%s", source)
	}
	if !strings.Contains(source, "on success") {
		t.Errorf("source should contain on-shorthand routing condition, got:\n%s", source)
	}
}

func TestMigrateLegacyAttributeNames(t *testing.T) {
	dot := `digraph G {
		Start [shape=Mdiamond];
		A [shape=box, llm_model="claude-opus-4-6", llm_provider="anthropic", prompt="Test"];
		Exit [shape=Msquare];
		Start -> A;
		A -> Exit;
	}`
	w, err := Migrate(dot)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	n := w.Node("A")
	if n == nil {
		t.Fatal("node A not found")
	}
	cfg := n.Config.(ir.AgentConfig)
	if cfg.Model != "claude-opus-4-6" {
		t.Errorf("model = %q, want %q", cfg.Model, "claude-opus-4-6")
	}
	if cfg.Provider != "anthropic" {
		t.Errorf("provider = %q, want %q", cfg.Provider, "anthropic")
	}
}

func TestMigrateHumanTimeoutDOT(t *testing.T) {
	dot := `digraph G {
		Start [shape=Mdiamond];
		H [shape=hexagon, mode=choice, timeout="5s", timeout_action="fail"];
		Exit [shape=Msquare];
		Start -> H;
		H -> Exit;
	}`
	w, err := Migrate(dot)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	n := w.Node("H")
	if n == nil {
		t.Fatal("node H not found")
	}
	cfg, ok := n.Config.(ir.HumanConfig)
	if !ok {
		t.Fatalf("expected HumanConfig, got %T", n.Config)
	}
	if cfg.Timeout != 5*time.Second {
		t.Errorf("timeout = %v, want 5s", cfg.Timeout)
	}
	if cfg.TimeoutAction != "fail" {
		t.Errorf("timeout_action = %q, want %q", cfg.TimeoutAction, "fail")
	}
}

func TestMigrateHumanTimeoutInvalidDuration(t *testing.T) {
	dot := `digraph G {
		Start [shape=Mdiamond];
		H [shape=hexagon, mode=choice, timeout="notaduration"];
		Exit [shape=Msquare];
		Start -> H;
		H -> Exit;
	}`
	_, err := Migrate(dot)
	if err == nil {
		t.Error("expected error for invalid timeout duration, got nil")
	}
}

func TestMigrateHumanTimeoutActionInvalid(t *testing.T) {
	dot := `digraph G {
		Start [shape=Mdiamond];
		H [shape=hexagon, mode=choice, timeout_action="explode"];
		Exit [shape=Msquare];
		Start -> H;
		H -> Exit;
	}`
	_, err := Migrate(dot)
	if err == nil {
		t.Error("expected error for invalid timeout_action, got nil")
	}
}

func TestMigrateBudgetDefaults(t *testing.T) {
	dot := `digraph G {
		graph [max_total_tokens="2000000", max_cost_cents="500", max_wall_time="30m"];
		Start [shape=Mdiamond];
		Exit [shape=Msquare];
		Start -> Exit;
	}`
	w, err := Migrate(dot)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if w.Defaults.MaxTotalTokens != 2000000 {
		t.Errorf("max_total_tokens = %d, want 2000000", w.Defaults.MaxTotalTokens)
	}
	if w.Defaults.MaxCostCents != 500 {
		t.Errorf("max_cost_cents = %d, want 500", w.Defaults.MaxCostCents)
	}
	if w.Defaults.MaxWallTime != 30*time.Minute {
		t.Errorf("max_wall_time = %v, want 30m", w.Defaults.MaxWallTime)
	}
}

// ============================================================
// Parity Checker Tests (8 cases)
// ============================================================

func makeTestWorkflow() *ir.Workflow {
	return &ir.Workflow{
		Name:  "test",
		Start: "A",
		Exit:  "C",
		Defaults: ir.WorkflowDefaults{
			Model:      "gpt-5.4",
			MaxRetries: 3,
		},
		Nodes: []*ir.Node{
			{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "Do A."}},
			{ID: "B", Kind: ir.NodeTool, Config: ir.ToolConfig{Command: "echo B"}},
			{ID: "C", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "Done."}},
		},
		Edges: []*ir.Edge{
			{From: "A", To: "B"},
			{From: "B", To: "C", Condition: &ir.Condition{
				Raw: "ctx.outcome = success",
			}},
		},
	}
}

func TestCheckParityIdentical(t *testing.T) {
	a := makeTestWorkflow()
	b := makeTestWorkflow()
	diffs := CheckParity(a, b)
	if len(diffs) != 0 {
		t.Errorf("expected 0 differences, got %d:", len(diffs))
		for _, d := range diffs {
			t.Logf("  %s: %s", d.Kind, d.Message)
		}
	}
}

func TestCheckParityMissingNode(t *testing.T) {
	a := makeTestWorkflow()
	b := makeTestWorkflow()
	// Remove node B from b.
	b.Nodes = []*ir.Node{b.Nodes[0], b.Nodes[2]}

	diffs := CheckParity(a, b)
	found := false
	for _, d := range diffs {
		if d.Kind == "node_missing" && strings.Contains(d.Message, "B") {
			found = true
		}
	}
	if !found {
		t.Error("expected node_missing difference for B")
	}
}

func TestCheckParityExtraNode(t *testing.T) {
	a := makeTestWorkflow()
	b := makeTestWorkflow()
	b.Nodes = append(b.Nodes, &ir.Node{ID: "X", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "Extra."}})

	diffs := CheckParity(a, b)
	found := false
	for _, d := range diffs {
		if d.Kind == "node_extra" && strings.Contains(d.Message, "X") {
			found = true
		}
	}
	if !found {
		t.Error("expected node_extra difference for X")
	}
}

func TestCheckParityStartMismatch(t *testing.T) {
	a := makeTestWorkflow()
	b := makeTestWorkflow()
	b.Start = "B"

	diffs := CheckParity(a, b)
	found := false
	for _, d := range diffs {
		if d.Kind == "start_mismatch" {
			found = true
		}
	}
	if !found {
		t.Error("expected start_mismatch difference")
	}
}

func TestCheckParityExitMismatch(t *testing.T) {
	a := makeTestWorkflow()
	b := makeTestWorkflow()
	b.Exit = "B"

	diffs := CheckParity(a, b)
	found := false
	for _, d := range diffs {
		if d.Kind == "exit_mismatch" {
			found = true
		}
	}
	if !found {
		t.Error("expected exit_mismatch difference")
	}
}

func TestCheckParityEdgeMissing(t *testing.T) {
	a := makeTestWorkflow()
	b := makeTestWorkflow()
	b.Edges = b.Edges[:1] // Remove second edge.

	diffs := CheckParity(a, b)
	found := false
	for _, d := range diffs {
		if d.Kind == "edge_missing" {
			found = true
		}
	}
	if !found {
		t.Error("expected edge_missing difference")
	}
}

func TestCheckParityConfigMismatch(t *testing.T) {
	a := makeTestWorkflow()
	b := makeTestWorkflow()
	// Change prompt on node A.
	b.Nodes[0].Config = ir.AgentConfig{Prompt: "Completely different prompt."}

	diffs := CheckParity(a, b)
	found := false
	for _, d := range diffs {
		if d.Kind == "config_mismatch" && strings.Contains(d.Message, "A") && strings.Contains(d.Message, "prompt") {
			found = true
		}
	}
	if !found {
		t.Error("expected config_mismatch for node A prompt")
	}
}

func TestCheckParityKindMismatch(t *testing.T) {
	a := makeTestWorkflow()
	b := makeTestWorkflow()
	b.Nodes[1].Kind = ir.NodeAgent // B was tool, now agent.
	b.Nodes[1].Config = ir.AgentConfig{Prompt: "Now an agent."}

	diffs := CheckParity(a, b)
	found := false
	for _, d := range diffs {
		if d.Kind == "kind_mismatch" && strings.Contains(d.Message, "B") {
			found = true
		}
	}
	if !found {
		t.Error("expected kind_mismatch for node B")
	}
}

func TestCheckParityWhitespaceTolerantPrompt(t *testing.T) {
	a := makeTestWorkflow()
	b := makeTestWorkflow()
	// Change prompt to add trailing whitespace — should still be considered equal.
	b.Nodes[0].Config = ir.AgentConfig{Prompt: "Do A.  "}

	diffs := CheckParity(a, b)
	// Filter for prompt-specific config_mismatch.
	for _, d := range diffs {
		if d.Kind == "config_mismatch" && strings.Contains(d.Message, "prompt") {
			t.Errorf("unexpected prompt mismatch with whitespace difference: %s", d.Message)
		}
	}
}

func TestCheckParityDefaultsMismatch(t *testing.T) {
	a := makeTestWorkflow()
	b := makeTestWorkflow()
	b.Defaults.MaxRetries = 10

	diffs := CheckParity(a, b)
	found := false
	for _, d := range diffs {
		if d.Kind == "defaults_mismatch" && strings.Contains(d.Message, "max_retries") {
			found = true
		}
	}
	if !found {
		t.Error("expected defaults_mismatch for max_retries")
	}
}

// ============================================================
// Integration Test: build_dippin.dot
// ============================================================

func TestMigrateBuildDippinDOT(t *testing.T) {
	data, err := os.ReadFile("../build_dippin.dot")
	if err != nil {
		t.Skipf("skipping integration test: %v", err)
	}

	w, err := Migrate(string(data))
	if err != nil {
		t.Fatalf("Migrate(build_dippin.dot) error: %v", err)
	}

	// Verify workflow name.
	if w.Name != "BuildDippin" {
		t.Errorf("name = %q, want %q", w.Name, "BuildDippin")
	}

	// Verify goal was extracted.
	if w.Goal == "" {
		t.Error("goal should not be empty")
	}
	if !strings.Contains(w.Goal, "Dippin toolchain") {
		t.Errorf("goal = %q, expected it to mention 'Dippin toolchain'", w.Goal)
	}

	// Verify start/exit.
	if w.Start != "Start" {
		t.Errorf("start = %q, want %q", w.Start, "Start")
	}
	if w.Exit != "Exit" {
		t.Errorf("exit = %q, want %q", w.Exit, "Exit")
	}

	// Verify defaults.
	if w.Defaults.MaxRetries != 3 {
		t.Errorf("defaults.max_retries = %d, want 3", w.Defaults.MaxRetries)
	}
	if w.Defaults.MaxRestarts != 7 {
		t.Errorf("defaults.max_restarts = %d, want 7", w.Defaults.MaxRestarts)
	}
	if w.Defaults.Fidelity != "summary:high" {
		t.Errorf("defaults.fidelity = %q, want %q", w.Defaults.Fidelity, "summary:high")
	}

	// Verify expected nodes exist.
	expectedNodes := []string{
		"Start", "Exit", "SetupWorkspace", "PickNextComponent",
		"CheckComplete", "PlanComponent", "ImplementParallel",
		"ImplementClaude", "ImplementGPT", "ImplementGemini",
		"ImplementJoin", "ValidateBuild", "ReviewParallel",
		"ReviewClaude", "ReviewGPT", "ReviewGemini",
		"ReviewsJoin", "ReviewAnalysis", "CommitWork",
		"MarkComplete", "FailureSummary", "MarkFailed",
	}
	for _, id := range expectedNodes {
		if w.Node(id) == nil {
			t.Errorf("expected node %q not found", id)
		}
	}

	// Verify node kinds.
	kindChecks := map[string]ir.NodeKind{
		"SetupWorkspace":    ir.NodeTool,
		"PickNextComponent": ir.NodeTool,
		"PlanComponent":     ir.NodeAgent,
		"ImplementParallel": ir.NodeParallel,
		"ImplementClaude":   ir.NodeAgent,
		"ImplementJoin":     ir.NodeFanIn,
		"ReviewParallel":    ir.NodeParallel,
		"ReviewsJoin":       ir.NodeFanIn,
		"ReviewAnalysis":    ir.NodeAgent,
		"CheckComplete":     ir.NodeConditional, // diamond → conditional
	}
	for id, wantKind := range kindChecks {
		n := w.Node(id)
		if n == nil {
			continue
		}
		if n.Kind != wantKind {
			t.Errorf("node %q kind = %q, want %q", id, n.Kind, wantKind)
		}
	}

	// Verify edges exist.
	if len(w.Edges) == 0 {
		t.Error("expected edges")
	}

	// Verify ImplementParallel targets are inferred.
	implPar := w.Node("ImplementParallel")
	if implPar != nil {
		cfg, ok := implPar.Config.(ir.ParallelConfig)
		if !ok {
			t.Errorf("ImplementParallel config type = %T, want ParallelConfig", implPar.Config)
		} else if len(cfg.Targets) != 3 {
			t.Errorf("ImplementParallel targets = %d, want 3", len(cfg.Targets))
		}
	}

	// Verify ImplementJoin sources are inferred.
	implJoin := w.Node("ImplementJoin")
	if implJoin != nil {
		cfg, ok := implJoin.Config.(ir.FanInConfig)
		if !ok {
			t.Errorf("ImplementJoin config type = %T, want FanInConfig", implJoin.Config)
		} else if len(cfg.Sources) != 3 {
			t.Errorf("ImplementJoin sources = %d, want 3", len(cfg.Sources))
		}
	}

	// Verify a restart edge exists (MarkComplete -> PickNextComponent).
	restartFound := false
	for _, e := range w.Edges {
		if e.From == "MarkComplete" && e.To == "PickNextComponent" && e.Restart {
			restartFound = true
		}
	}
	if !restartFound {
		t.Error("expected restart edge MarkComplete -> PickNextComponent")
	}

	// Verify conditions with context. prefix were normalized.
	for _, e := range w.Edges {
		if e.From == "CheckComplete" && e.To == "Exit" {
			if e.Condition == nil {
				t.Error("expected condition on CheckComplete -> Exit")
				break
			}
			cc, ok := e.Condition.Parsed.(ir.CondCompare)
			if !ok {
				t.Errorf("expected CondCompare, got %T", e.Condition.Parsed)
				break
			}
			if cc.Variable != "ctx.tool_stdout" {
				t.Errorf("variable = %q, want %q", cc.Variable, "ctx.tool_stdout")
			}
			break
		}
	}

	// Verify PlanComponent has a model from llm_model attribute.
	plan := w.Node("PlanComponent")
	if plan != nil {
		cfg, ok := plan.Config.(ir.AgentConfig)
		if !ok {
			t.Errorf("PlanComponent config type = %T, want AgentConfig", plan.Config)
		} else {
			if cfg.Model != "claude-opus-4-6" {
				t.Errorf("PlanComponent model = %q, want %q", cfg.Model, "claude-opus-4-6")
			}
			if cfg.Provider != "anthropic" {
				t.Errorf("PlanComponent provider = %q, want %q", cfg.Provider, "anthropic")
			}
		}
	}

	// Verify ReviewAnalysis has goal_gate and retry_target.
	ra := w.Node("ReviewAnalysis")
	if ra != nil {
		cfg, ok := ra.Config.(ir.AgentConfig)
		if !ok {
			t.Errorf("ReviewAnalysis config type = %T, want AgentConfig", ra.Config)
		} else {
			if !cfg.GoalGate {
				t.Error("ReviewAnalysis should have goal_gate=true")
			}
		}
		if ra.Retry.RetryTarget != "ImplementClaude" {
			t.Errorf("ReviewAnalysis retry_target = %q, want %q", ra.Retry.RetryTarget, "ImplementClaude")
		}
	}

	// Verify MigrateToSource doesn't error.
	source, err := MigrateToSource(string(data))
	if err != nil {
		t.Fatalf("MigrateToSource(build_dippin.dot) error: %v", err)
	}
	if !strings.HasPrefix(source, "workflow BuildDippin") {
		t.Errorf("source should start with 'workflow BuildDippin', got:\n%.100s...", source)
	}
}

// ============================================================
// Additional helper/edge case tests
// ============================================================

func TestAddNamespacePrefix(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"outcome", "ctx.outcome"},
		{"tool_stdout", "ctx.tool_stdout"},
		{"ctx.outcome", "ctx.outcome"},
		{"graph.goal", "graph.goal"},
		{"context.tool_stdout", "ctx.tool_stdout"},
		{"context.outcome", "ctx.outcome"},
		{"custom_var", "ctx.custom_var"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := addNamespacePrefix(tt.input)
			if got != tt.want {
				t.Errorf("addNamespacePrefix(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseConditionEdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		wantErr bool
		check   func(*ir.Condition) error
	}{
		{
			name: "simple equals",
			raw:  "outcome=success",
			check: func(c *ir.Condition) error {
				cc, ok := c.Parsed.(ir.CondCompare)
				if !ok {
					return fmt.Errorf("got %T, want CondCompare", c.Parsed)
				}
				if cc.Variable != "ctx.outcome" || cc.Op != "=" || cc.Value != "success" {
					return fmt.Errorf("got %+v", cc)
				}
				return nil
			},
		},
		{
			name: "not equals",
			raw:  "outcome!=fail",
			check: func(c *ir.Condition) error {
				cc, ok := c.Parsed.(ir.CondCompare)
				if !ok {
					return fmt.Errorf("got %T, want CondCompare", c.Parsed)
				}
				if cc.Op != "!=" {
					return fmt.Errorf("op = %q, want !=", cc.Op)
				}
				return nil
			},
		},
		{
			name: "contains operator",
			raw:  "tool_stdout contains pass",
			check: func(c *ir.Condition) error {
				cc, ok := c.Parsed.(ir.CondCompare)
				if !ok {
					return fmt.Errorf("got %T, want CondCompare", c.Parsed)
				}
				if cc.Op != "contains" {
					return fmt.Errorf("op = %q, want contains", cc.Op)
				}
				return nil
			},
		},
		{
			name: "OR condition",
			raw:  "outcome=success || outcome=partial",
			check: func(c *ir.Condition) error {
				_, ok := c.Parsed.(ir.CondOr)
				if !ok {
					return fmt.Errorf("got %T, want CondOr", c.Parsed)
				}
				return nil
			},
		},
		{
			name: "bang prefix negation",
			raw:  "!outcome=fail",
			check: func(c *ir.Condition) error {
				_, ok := c.Parsed.(ir.CondNot)
				if !ok {
					return fmt.Errorf("got %T, want CondNot", c.Parsed)
				}
				return nil
			},
		},
		{
			name:    "empty string",
			raw:     "",
			wantErr: false,
			check: func(c *ir.Condition) error {
				if c != nil {
					return fmt.Errorf("expected nil for empty condition, got %+v", c)
				}
				return nil
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := parseCondition(tt.raw)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.check != nil {
				if err := tt.check(c); err != nil {
					t.Error(err)
				}
			}
		})
	}
}

func TestMigrateNodeLabel(t *testing.T) {
	dot := `digraph G {
		Start [shape=Mdiamond, label="Begin Here"];
		A [shape=box, label="My Special Node", prompt="Do it."];
		Exit [shape=Msquare, label="The End"];
		Start -> A;
		A -> Exit;
	}`
	w, err := Migrate(dot)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	n := w.Node("A")
	if n == nil {
		t.Fatal("node A not found")
	}
	if n.Label != "My Special Node" {
		t.Errorf("label = %q, want %q", n.Label, "My Special Node")
	}
}

func TestMigrateWorkflowName(t *testing.T) {
	dot := `digraph MyWorkflow { A [shape=Mdiamond]; B [shape=Msquare]; A -> B; }`
	w, err := Migrate(dot)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if w.Name != "MyWorkflow" {
		t.Errorf("name = %q, want %q", w.Name, "MyWorkflow")
	}
}

func TestMigrateQuotedGraphName(t *testing.T) {
	dot := `digraph "my workflow" { A [shape=Mdiamond]; B [shape=Msquare]; A -> B; }`
	w, err := Migrate(dot)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if w.Name != "my workflow" {
		t.Errorf("name = %q, want %q", w.Name, "my workflow")
	}
}

func TestMigrateRetryConfig(t *testing.T) {
	dot := `digraph G {
		Start [shape=Mdiamond];
		A [shape=box, prompt="Try", max_retries=3, retry_policy="aggressive", retry_target="B", fallback_target="C"];
		B [shape=box, prompt="Retry here"];
		C [shape=box, prompt="Fallback"];
		Exit [shape=Msquare];
		Start -> A;
		A -> B;
		B -> Exit;
		A -> C;
		C -> Exit;
	}`
	w, err := Migrate(dot)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	n := w.Node("A")
	if n == nil {
		t.Fatal("node A not found")
	}
	if n.Retry.MaxRetries != 3 {
		t.Errorf("max_retries = %d, want 3", n.Retry.MaxRetries)
	}
	if n.Retry.Policy != "aggressive" {
		t.Errorf("policy = %q, want %q", n.Retry.Policy, "aggressive")
	}
	if n.Retry.RetryTarget != "B" {
		t.Errorf("retry_target = %q, want %q", n.Retry.RetryTarget, "B")
	}
	if n.Retry.FallbackTarget != "C" {
		t.Errorf("fallback_target = %q, want %q", n.Retry.FallbackTarget, "C")
	}
}

func TestMigrateSubgraphNode(t *testing.T) {
	dot := `digraph G {
		Start [shape=Mdiamond];
		S [shape=tab, ref="./review.dip", label="Review Sub"];
		Exit [shape=Msquare];
		Start -> S;
		S -> Exit;
	}`
	w, err := Migrate(dot)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	n := w.Node("S")
	if n == nil {
		t.Fatal("node S not found")
	}
	if n.Kind != ir.NodeSubgraph {
		t.Errorf("kind = %q, want %q", n.Kind, ir.NodeSubgraph)
	}
	cfg, ok := n.Config.(ir.SubgraphConfig)
	if !ok {
		t.Fatalf("config type = %T, want SubgraphConfig", n.Config)
	}
	if cfg.Ref != "./review.dip" {
		t.Errorf("ref = %q, want %q", cfg.Ref, "./review.dip")
	}
}

func TestMigrateHumanNode(t *testing.T) {
	dot := `digraph G {
		Start [shape=Mdiamond];
		H [shape=hexagon, mode="choice", default="Yes", label="Approval"];
		Exit [shape=Msquare];
		Start -> H;
		H -> Exit;
	}`
	w, err := Migrate(dot)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	n := w.Node("H")
	if n == nil {
		t.Fatal("node H not found")
	}
	if n.Kind != ir.NodeHuman {
		t.Errorf("kind = %q, want %q", n.Kind, ir.NodeHuman)
	}
	cfg, ok := n.Config.(ir.HumanConfig)
	if !ok {
		t.Fatalf("config type = %T, want HumanConfig", n.Config)
	}
	if cfg.Mode != "choice" {
		t.Errorf("mode = %q, want %q", cfg.Mode, "choice")
	}
	if cfg.Default != "Yes" {
		t.Errorf("default = %q, want %q", cfg.Default, "Yes")
	}
}

func TestMigrateParallelExplicitTargets(t *testing.T) {
	dot := `digraph G {
		Start [shape=Mdiamond];
		P [shape=component, targets="A,B,C"];
		A [shape=box];
		B [shape=box];
		C [shape=box];
		Exit [shape=Msquare];
		Start -> P;
		P -> A;
		P -> B;
		P -> C;
		A -> Exit;
		B -> Exit;
		C -> Exit;
	}`
	w, err := Migrate(dot)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	n := w.Node("P")
	if n == nil {
		t.Fatal("node P not found")
	}
	cfg, ok := n.Config.(ir.ParallelConfig)
	if !ok {
		t.Fatalf("config type = %T, want ParallelConfig", n.Config)
	}
	// Explicit targets from attribute should be used.
	if len(cfg.Targets) != 3 {
		t.Fatalf("targets = %d, want 3", len(cfg.Targets))
	}
	if cfg.Targets[0] != "A" || cfg.Targets[1] != "B" || cfg.Targets[2] != "C" {
		t.Errorf("targets = %v, want [A B C]", cfg.Targets)
	}
}

func TestMigrateVersionIsSet(t *testing.T) {
	dot := `digraph G { A [shape=Mdiamond]; B [shape=Msquare]; A -> B; }`
	w, err := Migrate(dot)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if w.Version != "1" {
		t.Errorf("version = %q, want %q", w.Version, "1")
	}
}

func TestMigrateAgentConfigFields(t *testing.T) {
	dot := `digraph G {
		Start [shape=Mdiamond];
		A [shape=box, prompt="Do it.", reasoning_effort="high", fidelity="full", goal_gate=true, auto_status=true];
		Exit [shape=Msquare];
		Start -> A;
		A -> Exit;
	}`
	w, err := Migrate(dot)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	n := w.Node("A")
	if n == nil {
		t.Fatal("node A not found")
	}
	cfg, ok := n.Config.(ir.AgentConfig)
	if !ok {
		t.Fatalf("config type = %T, want AgentConfig", n.Config)
	}
	if cfg.ReasoningEffort != "high" {
		t.Errorf("reasoning_effort = %q, want %q", cfg.ReasoningEffort, "high")
	}
	if cfg.Fidelity != "full" {
		t.Errorf("fidelity = %q, want %q", cfg.Fidelity, "full")
	}
	if !cfg.GoalGate {
		t.Error("expected goal_gate=true")
	}
	if !cfg.AutoStatus {
		t.Error("expected auto_status=true")
	}
}

func TestMigrateEdgeLabel(t *testing.T) {
	dot := `digraph G {
		Start [shape=Mdiamond];
		A [shape=box];
		B [shape=box];
		Exit [shape=Msquare];
		Start -> A;
		A -> B [label="proceed", condition="outcome=success"];
		B -> Exit;
	}`
	w, err := Migrate(dot)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	edges := w.EdgesFrom("A")
	if len(edges) != 1 {
		t.Fatalf("edges = %d, want 1", len(edges))
	}
	if edges[0].Label != "proceed" {
		t.Errorf("label = %q, want %q", edges[0].Label, "proceed")
	}
}

func TestNormalizeWhitespace(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"hello  world", "hello world"},
		{"  leading", "leading"},
		{"trailing  ", "trailing"},
		{"a\n\tb", "a b"},
		{"", ""},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := normalizeWhitespace(tt.input)
			if got != tt.want {
				t.Errorf("normalizeWhitespace(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseDOTDefaultNodeEdgeAttrs(t *testing.T) {
	input := `digraph G {
		node [fontname="Helvetica"];
		edge [fontname="Helvetica"];
		A [shape=box];
		B [shape=box];
		A -> B;
	}`
	dg, err := parseDOT(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Default node attrs should be merged into each node.
	if dg.Nodes[0].Attrs["fontname"] != "Helvetica" {
		t.Errorf("node A fontname = %q, want %q", dg.Nodes[0].Attrs["fontname"], "Helvetica")
	}
	// Default edge attrs should be merged into each edge.
	if dg.Edges[0].Attrs["fontname"] != "Helvetica" {
		t.Errorf("edge fontname = %q, want %q", dg.Edges[0].Attrs["fontname"], "Helvetica")
	}
}

func TestParseDOTDOTLeftJustify(t *testing.T) {
	// \l in DOT means left-justified newline — should be converted to \n.
	input := `digraph G {
		A [label="first\lsecond"];
	}`
	dg, err := parseDOT(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "first\nsecond"
	if dg.Nodes[0].Attrs["label"] != want {
		t.Errorf("label = %q, want %q", dg.Nodes[0].Attrs["label"], want)
	}
}

func TestCheckParityEdgeExtra(t *testing.T) {
	a := makeTestWorkflow()
	b := makeTestWorkflow()
	b.Edges = append(b.Edges, &ir.Edge{From: "C", To: "A"})

	diffs := CheckParity(a, b)
	found := false
	for _, d := range diffs {
		if d.Kind == "edge_extra" {
			found = true
		}
	}
	if !found {
		t.Error("expected edge_extra difference")
	}
}

// ============================================================
// Parity: structural config comparison tests
// ============================================================

func TestCheckParityHumanConfigMatch(t *testing.T) {
	a := &ir.Workflow{
		Name: "test", Start: "H", Exit: "D",
		Nodes: []*ir.Node{
			{ID: "H", Kind: ir.NodeHuman, Config: ir.HumanConfig{Mode: "choice"}},
			{ID: "D", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "Done."}},
		},
		Edges: []*ir.Edge{{From: "H", To: "D"}},
	}
	b := &ir.Workflow{
		Name: "test", Start: "H", Exit: "D",
		Nodes: []*ir.Node{
			{ID: "H", Kind: ir.NodeHuman, Config: ir.HumanConfig{Mode: "choice"}},
			{ID: "D", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "Done."}},
		},
		Edges: []*ir.Edge{{From: "H", To: "D"}},
	}
	diffs := CheckParity(a, b)
	if len(diffs) != 0 {
		t.Errorf("expected 0 diffs, got %d: %v", len(diffs), diffs)
	}
}

func TestCheckParityHumanConfigMismatch(t *testing.T) {
	a := &ir.Workflow{
		Name: "test", Start: "H", Exit: "D",
		Nodes: []*ir.Node{
			{ID: "H", Kind: ir.NodeHuman, Config: ir.HumanConfig{Mode: "choice"}},
			{ID: "D", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "Done."}},
		},
		Edges: []*ir.Edge{{From: "H", To: "D"}},
	}
	b := &ir.Workflow{
		Name: "test", Start: "H", Exit: "D",
		Nodes: []*ir.Node{
			{ID: "H", Kind: ir.NodeHuman, Config: ir.HumanConfig{Mode: "freeform"}},
			{ID: "D", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "Done."}},
		},
		Edges: []*ir.Edge{{From: "H", To: "D"}},
	}
	diffs := CheckParity(a, b)
	found := false
	for _, d := range diffs {
		if d.Kind == "config_mismatch" && strings.Contains(d.Message, "mode") {
			found = true
		}
	}
	if !found {
		t.Error("expected config_mismatch for mode")
	}
}

func TestCheckParityHumanConfigTypeMismatch(t *testing.T) {
	a := &ir.Workflow{
		Name: "test", Start: "H", Exit: "D",
		Nodes: []*ir.Node{
			{ID: "H", Kind: ir.NodeHuman, Config: ir.HumanConfig{Mode: "choice"}},
			{ID: "D", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "Done."}},
		},
		Edges: []*ir.Edge{{From: "H", To: "D"}},
	}
	b := &ir.Workflow{
		Name: "test", Start: "H", Exit: "D",
		Nodes: []*ir.Node{
			{ID: "H", Kind: ir.NodeHuman, Config: ir.AgentConfig{Prompt: "wrong type"}},
			{ID: "D", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "Done."}},
		},
		Edges: []*ir.Edge{{From: "H", To: "D"}},
	}
	diffs := CheckParity(a, b)
	found := false
	for _, d := range diffs {
		if d.Kind == "config_mismatch" && strings.Contains(d.Message, "type mismatch") {
			found = true
		}
	}
	if !found {
		t.Error("expected config_mismatch for type mismatch")
	}
}

func TestCheckParityParallelConfigMatch(t *testing.T) {
	a := &ir.Workflow{
		Name: "test", Start: "P", Exit: "D",
		Nodes: []*ir.Node{
			{ID: "P", Kind: ir.NodeParallel, Config: ir.ParallelConfig{Targets: []string{"A", "B"}}},
			{ID: "D", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "Done."}},
		},
		Edges: []*ir.Edge{{From: "P", To: "D"}},
	}
	b := &ir.Workflow{
		Name: "test", Start: "P", Exit: "D",
		Nodes: []*ir.Node{
			{ID: "P", Kind: ir.NodeParallel, Config: ir.ParallelConfig{Targets: []string{"A", "B"}}},
			{ID: "D", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "Done."}},
		},
		Edges: []*ir.Edge{{From: "P", To: "D"}},
	}
	diffs := CheckParity(a, b)
	if len(diffs) != 0 {
		t.Errorf("expected 0 diffs, got %d: %v", len(diffs), diffs)
	}
}

func TestCheckParityParallelConfigMismatch(t *testing.T) {
	a := &ir.Workflow{
		Name: "test", Start: "P", Exit: "D",
		Nodes: []*ir.Node{
			{ID: "P", Kind: ir.NodeParallel, Config: ir.ParallelConfig{Targets: []string{"A", "B"}}},
			{ID: "D", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "Done."}},
		},
		Edges: []*ir.Edge{{From: "P", To: "D"}},
	}
	b := &ir.Workflow{
		Name: "test", Start: "P", Exit: "D",
		Nodes: []*ir.Node{
			{ID: "P", Kind: ir.NodeParallel, Config: ir.ParallelConfig{Targets: []string{"X", "Y"}}},
			{ID: "D", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "Done."}},
		},
		Edges: []*ir.Edge{{From: "P", To: "D"}},
	}
	diffs := CheckParity(a, b)
	found := false
	for _, d := range diffs {
		if d.Kind == "config_mismatch" && strings.Contains(d.Message, "targets") {
			found = true
		}
	}
	if !found {
		t.Error("expected config_mismatch for targets")
	}
}

func TestCheckParityParallelConfigTypeMismatch(t *testing.T) {
	a := &ir.Workflow{
		Name: "test", Start: "P", Exit: "D",
		Nodes: []*ir.Node{
			{ID: "P", Kind: ir.NodeParallel, Config: ir.ParallelConfig{Targets: []string{"A"}}},
			{ID: "D", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "Done."}},
		},
		Edges: []*ir.Edge{{From: "P", To: "D"}},
	}
	b := &ir.Workflow{
		Name: "test", Start: "P", Exit: "D",
		Nodes: []*ir.Node{
			{ID: "P", Kind: ir.NodeParallel, Config: ir.AgentConfig{Prompt: "wrong"}},
			{ID: "D", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "Done."}},
		},
		Edges: []*ir.Edge{{From: "P", To: "D"}},
	}
	diffs := CheckParity(a, b)
	found := false
	for _, d := range diffs {
		if d.Kind == "config_mismatch" && strings.Contains(d.Message, "type mismatch") {
			found = true
		}
	}
	if !found {
		t.Error("expected config_mismatch for type mismatch")
	}
}

func TestCheckParityFanInConfigMatch(t *testing.T) {
	a := &ir.Workflow{
		Name: "test", Start: "J", Exit: "D",
		Nodes: []*ir.Node{
			{ID: "J", Kind: ir.NodeFanIn, Config: ir.FanInConfig{Sources: []string{"A", "B"}}},
			{ID: "D", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "Done."}},
		},
		Edges: []*ir.Edge{{From: "J", To: "D"}},
	}
	b := &ir.Workflow{
		Name: "test", Start: "J", Exit: "D",
		Nodes: []*ir.Node{
			{ID: "J", Kind: ir.NodeFanIn, Config: ir.FanInConfig{Sources: []string{"A", "B"}}},
			{ID: "D", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "Done."}},
		},
		Edges: []*ir.Edge{{From: "J", To: "D"}},
	}
	diffs := CheckParity(a, b)
	if len(diffs) != 0 {
		t.Errorf("expected 0 diffs, got %d: %v", len(diffs), diffs)
	}
}

func TestCheckParityFanInConfigMismatch(t *testing.T) {
	a := &ir.Workflow{
		Name: "test", Start: "J", Exit: "D",
		Nodes: []*ir.Node{
			{ID: "J", Kind: ir.NodeFanIn, Config: ir.FanInConfig{Sources: []string{"A", "B"}}},
			{ID: "D", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "Done."}},
		},
		Edges: []*ir.Edge{{From: "J", To: "D"}},
	}
	b := &ir.Workflow{
		Name: "test", Start: "J", Exit: "D",
		Nodes: []*ir.Node{
			{ID: "J", Kind: ir.NodeFanIn, Config: ir.FanInConfig{Sources: []string{"X"}}},
			{ID: "D", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "Done."}},
		},
		Edges: []*ir.Edge{{From: "J", To: "D"}},
	}
	diffs := CheckParity(a, b)
	found := false
	for _, d := range diffs {
		if d.Kind == "config_mismatch" && strings.Contains(d.Message, "sources") {
			found = true
		}
	}
	if !found {
		t.Error("expected config_mismatch for sources")
	}
}

func TestCheckParityFanInConfigTypeMismatch(t *testing.T) {
	a := &ir.Workflow{
		Name: "test", Start: "J", Exit: "D",
		Nodes: []*ir.Node{
			{ID: "J", Kind: ir.NodeFanIn, Config: ir.FanInConfig{Sources: []string{"A"}}},
			{ID: "D", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "Done."}},
		},
		Edges: []*ir.Edge{{From: "J", To: "D"}},
	}
	b := &ir.Workflow{
		Name: "test", Start: "J", Exit: "D",
		Nodes: []*ir.Node{
			{ID: "J", Kind: ir.NodeFanIn, Config: ir.AgentConfig{Prompt: "wrong"}},
			{ID: "D", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "Done."}},
		},
		Edges: []*ir.Edge{{From: "J", To: "D"}},
	}
	diffs := CheckParity(a, b)
	found := false
	for _, d := range diffs {
		if d.Kind == "config_mismatch" && strings.Contains(d.Message, "type mismatch") {
			found = true
		}
	}
	if !found {
		t.Error("expected config_mismatch for type mismatch")
	}
}

func TestCheckParitySubgraphConfigMatch(t *testing.T) {
	a := &ir.Workflow{
		Name: "test", Start: "S", Exit: "D",
		Nodes: []*ir.Node{
			{ID: "S", Kind: ir.NodeSubgraph, Config: ir.SubgraphConfig{Ref: "./sub.dip"}},
			{ID: "D", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "Done."}},
		},
		Edges: []*ir.Edge{{From: "S", To: "D"}},
	}
	b := &ir.Workflow{
		Name: "test", Start: "S", Exit: "D",
		Nodes: []*ir.Node{
			{ID: "S", Kind: ir.NodeSubgraph, Config: ir.SubgraphConfig{Ref: "./sub.dip"}},
			{ID: "D", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "Done."}},
		},
		Edges: []*ir.Edge{{From: "S", To: "D"}},
	}
	diffs := CheckParity(a, b)
	if len(diffs) != 0 {
		t.Errorf("expected 0 diffs, got %d: %v", len(diffs), diffs)
	}
}

func TestCheckParitySubgraphConfigMismatch(t *testing.T) {
	a := &ir.Workflow{
		Name: "test", Start: "S", Exit: "D",
		Nodes: []*ir.Node{
			{ID: "S", Kind: ir.NodeSubgraph, Config: ir.SubgraphConfig{Ref: "./sub.dip"}},
			{ID: "D", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "Done."}},
		},
		Edges: []*ir.Edge{{From: "S", To: "D"}},
	}
	b := &ir.Workflow{
		Name: "test", Start: "S", Exit: "D",
		Nodes: []*ir.Node{
			{ID: "S", Kind: ir.NodeSubgraph, Config: ir.SubgraphConfig{Ref: "./other.dip"}},
			{ID: "D", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "Done."}},
		},
		Edges: []*ir.Edge{{From: "S", To: "D"}},
	}
	diffs := CheckParity(a, b)
	found := false
	for _, d := range diffs {
		if d.Kind == "config_mismatch" && strings.Contains(d.Message, "ref") {
			found = true
		}
	}
	if !found {
		t.Error("expected config_mismatch for ref")
	}
}

func TestCheckParitySubgraphConfigTypeMismatch(t *testing.T) {
	a := &ir.Workflow{
		Name: "test", Start: "S", Exit: "D",
		Nodes: []*ir.Node{
			{ID: "S", Kind: ir.NodeSubgraph, Config: ir.SubgraphConfig{Ref: "./sub.dip"}},
			{ID: "D", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "Done."}},
		},
		Edges: []*ir.Edge{{From: "S", To: "D"}},
	}
	b := &ir.Workflow{
		Name: "test", Start: "S", Exit: "D",
		Nodes: []*ir.Node{
			{ID: "S", Kind: ir.NodeSubgraph, Config: ir.AgentConfig{Prompt: "wrong"}},
			{ID: "D", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "Done."}},
		},
		Edges: []*ir.Edge{{From: "S", To: "D"}},
	}
	diffs := CheckParity(a, b)
	found := false
	for _, d := range diffs {
		if d.Kind == "config_mismatch" && strings.Contains(d.Message, "type mismatch") {
			found = true
		}
	}
	if !found {
		t.Error("expected config_mismatch for type mismatch")
	}
}

// ============================================================
// Parity: agent behavior and defaults coverage
// ============================================================

func TestCheckParityAgentBehaviorMismatch(t *testing.T) {
	a := &ir.Workflow{
		Name: "test", Start: "A", Exit: "A",
		Nodes: []*ir.Node{
			{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{
				Prompt: "Do it.", GoalGate: true, AutoStatus: false,
			}},
		},
	}
	b := &ir.Workflow{
		Name: "test", Start: "A", Exit: "A",
		Nodes: []*ir.Node{
			{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{
				Prompt: "Do it.", GoalGate: false, AutoStatus: true,
			}},
		},
	}
	diffs := CheckParity(a, b)
	foundGoalGate := false
	foundAutoStatus := false
	for _, d := range diffs {
		if strings.Contains(d.Message, "goal_gate") {
			foundGoalGate = true
		}
		if strings.Contains(d.Message, "auto_status") {
			foundAutoStatus = true
		}
	}
	if !foundGoalGate {
		t.Error("expected config_mismatch for goal_gate")
	}
	if !foundAutoStatus {
		t.Error("expected config_mismatch for auto_status")
	}
}

func TestCheckParityAgentModelProviderMismatch(t *testing.T) {
	a := &ir.Workflow{
		Name: "test", Start: "A", Exit: "A",
		Nodes: []*ir.Node{
			{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{
				Prompt: "Do it.", Model: "gpt-5.4", Provider: "openai",
			}},
		},
	}
	b := &ir.Workflow{
		Name: "test", Start: "A", Exit: "A",
		Nodes: []*ir.Node{
			{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{
				Prompt: "Do it.", Model: "claude-opus-4-6", Provider: "anthropic",
			}},
		},
	}
	diffs := CheckParity(a, b)
	foundModel := false
	foundProvider := false
	for _, d := range diffs {
		if strings.Contains(d.Message, "model") {
			foundModel = true
		}
		if strings.Contains(d.Message, "provider") {
			foundProvider = true
		}
	}
	if !foundModel {
		t.Error("expected config_mismatch for model")
	}
	if !foundProvider {
		t.Error("expected config_mismatch for provider")
	}
}

func TestCheckParityToolConfigTypeMismatch(t *testing.T) {
	a := &ir.Workflow{
		Name: "test", Start: "T", Exit: "T",
		Nodes: []*ir.Node{
			{ID: "T", Kind: ir.NodeTool, Config: ir.ToolConfig{Command: "echo hi"}},
		},
	}
	b := &ir.Workflow{
		Name: "test", Start: "T", Exit: "T",
		Nodes: []*ir.Node{
			{ID: "T", Kind: ir.NodeTool, Config: ir.AgentConfig{Prompt: "wrong"}},
		},
	}
	diffs := CheckParity(a, b)
	found := false
	for _, d := range diffs {
		if d.Kind == "config_mismatch" && strings.Contains(d.Message, "type mismatch") {
			found = true
		}
	}
	if !found {
		t.Error("expected config_mismatch for type mismatch")
	}
}

func TestCheckParityDefaultsFieldMismatch(t *testing.T) {
	a := &ir.Workflow{
		Name: "test", Start: "A", Exit: "A",
		Defaults: ir.WorkflowDefaults{
			Model: "gpt-5.4", Provider: "openai", Fidelity: "full",
			MaxRetries: 3, MaxRestarts: 5,
		},
		Nodes: []*ir.Node{{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "Do."}}},
	}
	b := &ir.Workflow{
		Name: "test", Start: "A", Exit: "A",
		Defaults: ir.WorkflowDefaults{
			Model: "claude-opus-4-6", Provider: "anthropic", Fidelity: "summary",
			MaxRetries: 1, MaxRestarts: 2,
		},
		Nodes: []*ir.Node{{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "Do."}}},
	}
	diffs := CheckParity(a, b)
	fields := map[string]bool{}
	for _, d := range diffs {
		if d.Kind == "defaults_mismatch" {
			fields[d.PathA] = true
		}
	}
	for _, f := range []string{"defaults.model", "defaults.provider", "defaults.fidelity", "defaults.max_retries", "defaults.max_restarts"} {
		if !fields[f] {
			t.Errorf("expected defaults_mismatch for %s", f)
		}
	}
}

// ============================================================
// Migration: max_turns and cmd_timeout coverage
// ============================================================

func TestMigrateMaxTurns(t *testing.T) {
	dot := `digraph G {
		Start [shape=Mdiamond];
		A [shape=box, prompt="Do it.", max_turns=5];
		Exit [shape=Msquare];
		Start -> A;
		A -> Exit;
	}`
	w, err := Migrate(dot)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	n := w.Node("A")
	if n == nil {
		t.Fatal("node A not found")
	}
	cfg := n.Config.(ir.AgentConfig)
	if cfg.MaxTurns != 5 {
		t.Errorf("max_turns = %d, want 5", cfg.MaxTurns)
	}
}

func TestMigrateMaxTurnsInvalid(t *testing.T) {
	dot := `digraph G {
		Start [shape=Mdiamond];
		A [shape=box, prompt="Do it.", max_turns=abc];
		Exit [shape=Msquare];
		Start -> A;
		A -> Exit;
	}`
	w, err := Migrate(dot)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	n := w.Node("A")
	if n == nil {
		t.Fatal("node A not found")
	}
	cfg := n.Config.(ir.AgentConfig)
	if cfg.MaxTurns != 0 {
		t.Errorf("max_turns = %d, want 0 for invalid input", cfg.MaxTurns)
	}
}

func TestMigrateCmdTimeout(t *testing.T) {
	dot := `digraph G {
		Start [shape=Mdiamond];
		A [shape=box, prompt="Do it.", cmd_timeout="30s"];
		Exit [shape=Msquare];
		Start -> A;
		A -> Exit;
	}`
	w, err := Migrate(dot)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	n := w.Node("A")
	if n == nil {
		t.Fatal("node A not found")
	}
	cfg := n.Config.(ir.AgentConfig)
	if cfg.CmdTimeout != 30*time.Second {
		t.Errorf("cmd_timeout = %v, want 30s", cfg.CmdTimeout)
	}
}

func TestMigrateCmdTimeoutInvalid(t *testing.T) {
	dot := `digraph G {
		Start [shape=Mdiamond];
		A [shape=box, prompt="Do it.", cmd_timeout="notaduration"];
		Exit [shape=Msquare];
		Start -> A;
		A -> Exit;
	}`
	w, err := Migrate(dot)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	n := w.Node("A")
	if n == nil {
		t.Fatal("node A not found")
	}
	cfg := n.Config.(ir.AgentConfig)
	if cfg.CmdTimeout != 0 {
		t.Errorf("cmd_timeout = %v, want 0 for invalid input", cfg.CmdTimeout)
	}
}

func TestMigrateCacheToolsAndCompaction(t *testing.T) {
	dot := `digraph G {
		Start [shape=Mdiamond];
		A [shape=box, prompt="Do it.", cache_tools=true, compaction="aggressive"];
		Exit [shape=Msquare];
		Start -> A;
		A -> Exit;
	}`
	w, err := Migrate(dot)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	n := w.Node("A")
	if n == nil {
		t.Fatal("node A not found")
	}
	cfg := n.Config.(ir.AgentConfig)
	if !cfg.CacheTools {
		t.Error("expected cache_tools=true")
	}
	if cfg.Compaction != "aggressive" {
		t.Errorf("compaction = %q, want %q", cfg.Compaction, "aggressive")
	}
}

func TestMigrateSystemPrompt(t *testing.T) {
	dot := `digraph G {
		Start [shape=Mdiamond];
		A [shape=box, prompt="Do it.", system_prompt="You are a helper."];
		Exit [shape=Msquare];
		Start -> A;
		A -> Exit;
	}`
	w, err := Migrate(dot)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	n := w.Node("A")
	if n == nil {
		t.Fatal("node A not found")
	}
	cfg := n.Config.(ir.AgentConfig)
	if cfg.SystemPrompt != "You are a helper." {
		t.Errorf("system_prompt = %q, want %q", cfg.SystemPrompt, "You are a helper.")
	}
}

func TestMigrateFanInExplicitSources(t *testing.T) {
	dot := `digraph G {
		Start [shape=Mdiamond];
		J [shape=tripleoctagon, sources="A,B"];
		A [shape=box];
		B [shape=box];
		Exit [shape=Msquare];
		Start -> A;
		Start -> B;
		A -> J;
		B -> J;
		J -> Exit;
	}`
	w, err := Migrate(dot)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	n := w.Node("J")
	if n == nil {
		t.Fatal("node J not found")
	}
	cfg, ok := n.Config.(ir.FanInConfig)
	if !ok {
		t.Fatalf("config type = %T, want FanInConfig", n.Config)
	}
	if len(cfg.Sources) != 2 || cfg.Sources[0] != "A" || cfg.Sources[1] != "B" {
		t.Errorf("sources = %v, want [A B]", cfg.Sources)
	}
}

func TestMigrateGraphDefaultsModelProvider(t *testing.T) {
	dot := `digraph G {
		graph [model="claude-sonnet-4-6", provider="anthropic", max_retries=2];
		Start [shape=Mdiamond];
		Exit [shape=Msquare];
		Start -> Exit;
	}`
	w, err := Migrate(dot)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if w.Defaults.Model != "claude-sonnet-4-6" {
		t.Errorf("defaults.model = %q, want %q", w.Defaults.Model, "claude-sonnet-4-6")
	}
	if w.Defaults.Provider != "anthropic" {
		t.Errorf("defaults.provider = %q, want %q", w.Defaults.Provider, "anthropic")
	}
	if w.Defaults.MaxRetries != 2 {
		t.Errorf("defaults.max_retries = %d, want 2", w.Defaults.MaxRetries)
	}
}

func TestMigrateConditionOr(t *testing.T) {
	dot := `digraph G {
		Start [shape=Mdiamond];
		A [shape=box];
		B [shape=box];
		Exit [shape=Msquare];
		Start -> A;
		A -> B [condition="outcome=success || outcome=partial"];
		B -> Exit;
	}`
	w, err := Migrate(dot)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	edges := w.EdgesFrom("A")
	if len(edges) != 1 {
		t.Fatalf("edges = %d, want 1", len(edges))
	}
	_, ok := edges[0].Condition.Parsed.(ir.CondOr)
	if !ok {
		t.Fatalf("expected CondOr, got %T", edges[0].Condition.Parsed)
	}
}

func TestMigrateToSourceError(t *testing.T) {
	_, err := MigrateToSource("not a valid dot graph")
	if err == nil {
		t.Error("expected error for invalid DOT input")
	}
}

func TestMigrateGraphAttrsToVars(t *testing.T) {
	input := `digraph pipeline {
  graph [goal="Run the pipeline", model="claude-opus-4-6", max_retries=3, max_restarts=2, api_url="example.com/api", env=production];
  A [shape=box, label="Agent A"];
  B [shape=Msquare, label="Done"];
  A -> B;
}`
	w, err := Migrate(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Known attrs go to Defaults
	if w.Defaults.MaxRetries != 3 {
		t.Errorf("MaxRetries = %d, want 3", w.Defaults.MaxRetries)
	}
	if w.Defaults.MaxRestarts != 2 {
		t.Errorf("MaxRestarts = %d, want 2", w.Defaults.MaxRestarts)
	}
	if w.Defaults.Model != "claude-opus-4-6" {
		t.Errorf("Defaults.Model = %q, want %q", w.Defaults.Model, "claude-opus-4-6")
	}

	// Unknown attrs go to Vars
	if w.Vars["api_url"] != "example.com/api" {
		t.Errorf("Vars[api_url] = %q, want %q", w.Vars["api_url"], "example.com/api")
	}
	if w.Vars["env"] != "production" {
		t.Errorf("Vars[env] = %q, want %q", w.Vars["env"], "production")
	}

	// goal is a known handler — should NOT appear in Vars
	if _, ok := w.Vars["goal"]; ok {
		t.Errorf("goal should not be in Vars, but it is: %q", w.Vars["goal"])
	}
}

func TestMigrate_ManagerLoop(t *testing.T) {
	dot := `digraph W {
  S [shape=Mdiamond, label="S"];
  M [shape=house, label="Supervisor", subgraph_ref="inner", poll_interval="30s", max_cycles="12", stop_condition="stack.child.cycles = 10", steer_condition="stack.child.cycles = 5", steer_context="hint=speed_up,priority=high"];
  E [shape=Msquare, label="E"];
  S -> M;
  M -> E;
}`
	w, err := Migrate(dot)
	if err != nil {
		t.Fatalf("Migrate: %v", err)
	}
	n := w.Node("M")
	if n == nil {
		t.Fatalf("node M not found in workflow")
	}
	if n.Kind != ir.NodeManagerLoop {
		t.Fatalf("Kind = %v, want NodeManagerLoop", n.Kind)
	}
	cfg, ok := n.Config.(ir.ManagerLoopConfig)
	if !ok {
		t.Fatalf("Config = %T, want ManagerLoopConfig", n.Config)
	}
	if cfg.SubgraphRef != "inner" {
		t.Errorf("SubgraphRef = %q, want %q", cfg.SubgraphRef, "inner")
	}
	if cfg.MaxCycles != 12 {
		t.Errorf("MaxCycles = %d, want 12", cfg.MaxCycles)
	}
	if cfg.PollInterval != 30*time.Second {
		t.Errorf("PollInterval = %v, want 30s", cfg.PollInterval)
	}
	if cfg.StopCondition == nil || cfg.StopCondition.Raw != "stack.child.cycles = 10" {
		t.Errorf("StopCondition = %+v", cfg.StopCondition)
	}
	if cfg.SteerCondition == nil || cfg.SteerCondition.Raw != "stack.child.cycles = 5" {
		t.Errorf("SteerCondition = %+v", cfg.SteerCondition)
	}
	if cfg.SteerContext["hint"] != "speed_up" || cfg.SteerContext["priority"] != "high" {
		t.Errorf("SteerContext = %v", cfg.SteerContext)
	}
}

func TestMigrate_ManagerLoop_Minimal(t *testing.T) {
	// A manager_loop node with only subgraph_ref set should migrate cleanly
	// with zero-value PollInterval, MaxCycles, and nil condition pointers.
	// Exercises the attr-presence guards in applyManagerLoopScalarMigrate
	// and applyManagerLoopConditionMigrate.
	dot := `digraph W {
  S [shape=Mdiamond, label="S"];
  M [shape=house, label="Minimal", subgraph_ref="inner"];
  E [shape=Msquare, label="E"];
  S -> M;
  M -> E;
}`
	w, err := Migrate(dot)
	if err != nil {
		t.Fatalf("Migrate: %v", err)
	}
	n := w.Node("M")
	if n == nil || n.Kind != ir.NodeManagerLoop {
		t.Fatalf("node M wrong: %+v", n)
	}
	cfg, ok := n.Config.(ir.ManagerLoopConfig)
	if !ok {
		t.Fatalf("Config = %T", n.Config)
	}
	if cfg.SubgraphRef != "inner" {
		t.Errorf("SubgraphRef = %q", cfg.SubgraphRef)
	}
	if cfg.PollInterval != 0 {
		t.Errorf("PollInterval = %v, want 0", cfg.PollInterval)
	}
	if cfg.MaxCycles != 0 {
		t.Errorf("MaxCycles = %d, want 0", cfg.MaxCycles)
	}
	if cfg.StopCondition != nil {
		t.Errorf("StopCondition = %+v, want nil", cfg.StopCondition)
	}
	if cfg.SteerCondition != nil {
		t.Errorf("SteerCondition = %+v, want nil", cfg.SteerCondition)
	}
	if cfg.SteerContext == nil {
		t.Errorf("SteerContext should be non-nil empty map")
	}
	if len(cfg.SteerContext) != 0 {
		t.Errorf("SteerContext should be empty; got %v", cfg.SteerContext)
	}
}

func TestMigrate_ManagerLoop_AsStartNode(t *testing.T) {
	// A manager_loop node at Start gets shape=Mdiamond from export, not shape=house.
	// migrate must still reconstruct it as NodeManagerLoop by looking at attrs.
	dot := `digraph W {
  Supervise [shape=Mdiamond, label="Supervisor", subgraph_ref="inner", max_cycles="5", stop_condition="stack.child.outcome = success"];
  Done [shape=Msquare, label="Done"];
  Supervise -> Done;
}`
	w, err := Migrate(dot)
	if err != nil {
		t.Fatalf("Migrate: %v", err)
	}
	n := w.Node("Supervise")
	if n == nil || n.Kind != ir.NodeManagerLoop {
		t.Fatalf("Supervise kind = %v, want NodeManagerLoop", n)
	}
	cfg, ok := n.Config.(ir.ManagerLoopConfig)
	if !ok {
		t.Fatalf("Config = %T", n.Config)
	}
	if cfg.SubgraphRef != "inner" || cfg.MaxCycles != 5 {
		t.Errorf("fields lost on migrate through Mdiamond: %+v", cfg)
	}
}

func TestMigrate_ManagerLoop_PartialConfigAtStartNode(t *testing.T) {
	// A manager_loop with only poll_interval + max_cycles set (no subgraph_ref yet)
	// must still be recognized as NodeManagerLoop when at start, so the config
	// isn't silently dropped during migrate.
	dot := `digraph W {
  Supervise [shape=Mdiamond, label="Supervisor", poll_interval="10s", max_cycles="5"];
  Done [shape=Msquare, label="Done"];
  Supervise -> Done;
}`
	w, err := Migrate(dot)
	if err != nil {
		t.Fatalf("Migrate: %v", err)
	}
	n := w.Node("Supervise")
	if n == nil || n.Kind != ir.NodeManagerLoop {
		t.Fatalf("Supervise kind = %v, want NodeManagerLoop (partial manager_loop attrs should still resolve)", n)
	}
}

func TestBuildToolConfigRoutingAttrs(t *testing.T) {
	attrs := map[string]string{
		"tool_command":   "echo hi",
		"timeout":        "30s",
		"marker_grep":    "^pass$",
		"route_required": "true",
		"output_limit":   "8192",
	}
	cfg, err := buildToolConfig(attrs)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.MarkerGrep != "^pass$" {
		t.Errorf("MarkerGrep = %q", cfg.MarkerGrep)
	}
	if !cfg.RouteRequired {
		t.Error("RouteRequired = false")
	}
	if cfg.OutputLimit != 8192 {
		t.Errorf("OutputLimit = %d", cfg.OutputLimit)
	}
}

func TestBuildToolConfigOutputLimitInvalid(t *testing.T) {
	attrs := map[string]string{"output_limit": "not_a_number"}
	_, err := buildToolConfig(attrs)
	if err == nil {
		t.Error("expected error for non-numeric output_limit, got nil")
	}
}

func TestBuildToolConfigOutputLimitNegative(t *testing.T) {
	attrs := map[string]string{"output_limit": "-1"}
	_, err := buildToolConfig(attrs)
	if err == nil {
		t.Error("expected error for negative output_limit, got nil")
	}
}

func TestBuildToolConfigOutputLimitZero(t *testing.T) {
	attrs := map[string]string{"tool_command": "echo test", "output_limit": "0"}
	cfg, err := buildToolConfig(attrs)
	if err != nil {
		t.Fatalf("unexpected error for output_limit=0: %v", err)
	}
	if cfg.OutputLimit != 0 {
		t.Errorf("OutputLimit = %d, want 0 (engine default)", cfg.OutputLimit)
	}
}

func TestMigrate_ToolAccessAttr(t *testing.T) {
	dot := `digraph X {
  graph [rankdir=TB];
  A [shape=box, label=A, prompt="x", tool_access="none"];
}`
	w, err := Migrate(dot)
	if err != nil {
		t.Fatalf("migrate error: %v", err)
	}
	node := w.Node("A")
	if node == nil {
		t.Fatalf("node A not found")
	}
	cfg, ok := node.Config.(ir.AgentConfig)
	if !ok {
		t.Fatalf("expected AgentConfig, got %T", node.Config)
	}
	if cfg.ToolAccess != "none" {
		t.Errorf("ToolAccess = %q, want %q", cfg.ToolAccess, "none")
	}
}

func TestMigrateParallelBranches(t *testing.T) {
	dot := `digraph G {
		Start [shape=Mdiamond];
		P [shape=component, targets="fast,accurate", branches="target=fast;model=claude-haiku-4-5;provider=anthropic;fidelity=summary,target=accurate;model=claude-opus-4-7;provider=anthropic;fidelity=full"];
		fast [shape=box];
		accurate [shape=box];
		Exit [shape=Msquare];
		Start -> P;
		P -> fast;
		P -> accurate;
		fast -> Exit;
		accurate -> Exit;
	}`
	w, err := Migrate(dot)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	cfg, ok := w.Node("P").Config.(ir.ParallelConfig)
	if !ok {
		t.Fatalf("config type = %T, want ParallelConfig", w.Node("P").Config)
	}
	if len(cfg.Branches) != 2 {
		t.Fatalf("branches = %d, want 2", len(cfg.Branches))
	}
	if len(cfg.Targets) != 2 || cfg.Targets[0] != "fast" || cfg.Targets[1] != "accurate" {
		t.Errorf("targets = %v, want [fast accurate]", cfg.Targets)
	}
	b0 := cfg.Branches[0]
	if b0.Target != "fast" || b0.Model != "claude-haiku-4-5" || b0.Provider != "anthropic" || b0.Fidelity != "summary" {
		t.Errorf("branch[0] = %+v", b0)
	}
	b1 := cfg.Branches[1]
	if b1.Target != "accurate" || b1.Model != "claude-opus-4-7" || b1.Provider != "anthropic" || b1.Fidelity != "full" {
		t.Errorf("branch[1] = %+v", b1)
	}
}

// Inline form (no branches=) must leave Branches nil so the formatter keeps
// inline output.
func TestMigrateParallelInlineNoBranches(t *testing.T) {
	cfg := buildParallelConfig(map[string]string{"targets": "A,B"})
	if cfg.Branches != nil {
		t.Errorf("branches = %+v, want nil", cfg.Branches)
	}
	if len(cfg.Targets) != 2 {
		t.Errorf("targets = %v, want [A B]", cfg.Targets)
	}
}

// A branch token with no target= is dropped (a zero-target branch would corrupt
// the edge mapping); unknown keys are skipped.
func TestMigrateParallelBranchesMalformed(t *testing.T) {
	cfg := buildParallelConfig(map[string]string{
		"branches": "model=x,target=B;bogus=1;model=y",
	})
	if len(cfg.Branches) != 1 {
		t.Fatalf("branches = %d, want 1 (target-less branch dropped)", len(cfg.Branches))
	}
	if cfg.Branches[0].Target != "B" || cfg.Branches[0].Model != "y" {
		t.Errorf("branch[0] = %+v, want {Target:B Model:y}", cfg.Branches[0])
	}
}

// inferParallelFanIn rebuilds ParallelConfig when Targets is empty; it must not
// discard Branches in the process.
func TestInferParallelTargetsPreservesBranches(t *testing.T) {
	w := &ir.Workflow{
		Nodes: []*ir.Node{
			{ID: "P", Kind: ir.NodeParallel, Config: ir.ParallelConfig{
				Branches: []ir.BranchConfig{{Target: "A", Model: "m"}},
			}},
		},
		Edges: []*ir.Edge{{From: "P", To: "A"}},
	}
	inferParallelFanIn(w)
	cfg := w.Node("P").Config.(ir.ParallelConfig)
	if len(cfg.Branches) != 1 || cfg.Branches[0].Model != "m" {
		t.Errorf("branches lost during inference: %+v", cfg.Branches)
	}
	if len(cfg.Targets) != 1 || cfg.Targets[0] != "A" {
		t.Errorf("targets = %v, want [A]", cfg.Targets)
	}
}

// A parallel node that is also the workflow start (shape overridden to Mdiamond)
// must be recovered as NodeParallel with its branches intact, not degraded to
// NodeAgent.
func TestMigrateParallelAsStartNode(t *testing.T) {
	dot := `digraph G {
		split [shape=Mdiamond, targets="fast,accurate", branches="target=fast;model=claude-haiku-4-5,target=accurate;model=claude-opus-4-7"];
		fast [shape=box];
		accurate [shape=box];
		join [shape=Msquare];
		split -> fast;
		split -> accurate;
		fast -> join;
		accurate -> join;
	}`
	w, err := Migrate(dot)
	if err != nil {
		t.Fatalf("migrate: %v", err)
	}
	n := w.Node("split")
	if n.Kind != ir.NodeParallel {
		t.Errorf("kind = %v, want parallel", n.Kind)
	}
	cfg, ok := n.Config.(ir.ParallelConfig)
	if !ok {
		t.Fatalf("config type = %T, want ParallelConfig", n.Config)
	}
	if len(cfg.Branches) != 2 {
		t.Fatalf("branches = %d, want 2", len(cfg.Branches))
	}
	if cfg.Branches[0].Model != "claude-haiku-4-5" || cfg.Branches[1].Model != "claude-opus-4-7" {
		t.Errorf("branch models = %q, %q", cfg.Branches[0].Model, cfg.Branches[1].Model)
	}
}

func TestMigrateBranchToolAccess(t *testing.T) {
	dot := `digraph G {
		Start [shape=Mdiamond];
		P [shape=component, targets="a,b", branches="target=a;tool_access=none,target=b;model=claude-haiku-4-5"];
		a [shape=box];
		b [shape=box];
		Exit [shape=Msquare];
		Start -> P;
		P -> a;
		P -> b;
		a -> Exit;
		b -> Exit;
	}`
	w, err := Migrate(dot)
	if err != nil {
		t.Fatalf("migrate: %v", err)
	}
	cfg, ok := w.Node("P").Config.(ir.ParallelConfig)
	if !ok {
		t.Fatalf("config type = %T, want ParallelConfig", w.Node("P").Config)
	}
	if len(cfg.Branches) != 2 {
		t.Fatalf("branches = %d, want 2", len(cfg.Branches))
	}
	if cfg.Branches[0].ToolAccess != "none" {
		t.Errorf("branch[0].ToolAccess = %q, want none", cfg.Branches[0].ToolAccess)
	}
	if cfg.Branches[1].ToolAccess != "" {
		t.Errorf("branch[1].ToolAccess = %q, want empty", cfg.Branches[1].ToolAccess)
	}
}

func TestMigrate_AgentLastResponseTruncate_RoundTrip(t *testing.T) {
	dot := `digraph X {
  graph [rankdir=TB];
  A [shape=box, label=A, prompt="x", last_response_truncate="4096"];
}`
	w, err := Migrate(dot)
	if err != nil {
		t.Fatalf("migrate error: %v", err)
	}
	node := w.Node("A")
	if node == nil {
		t.Fatalf("node A not found")
	}
	cfg, ok := node.Config.(ir.AgentConfig)
	if !ok {
		t.Fatalf("expected AgentConfig, got %T", node.Config)
	}
	if cfg.LastResponseTruncate != 4096 {
		t.Errorf("last_response_truncate lost on DOT round-trip: got %d, want 4096", cfg.LastResponseTruncate)
	}
}

func TestMigrateBranchLastResponseTruncate(t *testing.T) {
	dot := `digraph G {
		Start [shape=Mdiamond];
		P [shape=component, targets="a,b", branches="target=a;last_response_truncate=4096,target=b;model=claude-haiku-4-5"];
		a [shape=box];
		b [shape=box];
		Exit [shape=Msquare];
		Start -> P;
		P -> a;
		P -> b;
		a -> Exit;
		b -> Exit;
	}`
	w, err := Migrate(dot)
	if err != nil {
		t.Fatalf("migrate: %v", err)
	}
	cfg, ok := w.Node("P").Config.(ir.ParallelConfig)
	if !ok {
		t.Fatalf("config type = %T, want ParallelConfig", w.Node("P").Config)
	}
	if len(cfg.Branches) != 2 {
		t.Fatalf("branches = %d, want 2", len(cfg.Branches))
	}
	if cfg.Branches[0].LastResponseTruncate != 4096 {
		t.Errorf("branch[0].LastResponseTruncate = %d, want 4096", cfg.Branches[0].LastResponseTruncate)
	}
	if cfg.Branches[1].LastResponseTruncate != 0 {
		t.Errorf("branch[1].LastResponseTruncate = %d, want 0", cfg.Branches[1].LastResponseTruncate)
	}
}
