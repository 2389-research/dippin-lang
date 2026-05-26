package export

import (
	"strings"
	"testing"
	"time"

	"github.com/2389-research/dippin-lang/ir"
)

// --- Fixtures ---

func minimalWorkflow() *ir.Workflow {
	return &ir.Workflow{
		Name:  "minimal",
		Start: "Begin",
		Exit:  "End",
		Nodes: []*ir.Node{
			{ID: "Begin", Kind: ir.NodeHuman, Config: ir.HumanConfig{Mode: "freeform"}},
			{ID: "End", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "Done."}},
		},
		Edges: []*ir.Edge{
			{From: "Begin", To: "End"},
		},
	}
}

func askAndExecuteWorkflow() *ir.Workflow {
	return &ir.Workflow{
		Name:  "ask_and_execute",
		Goal:  "Ask user for a task, implement it, review, ship",
		Start: "AskUser",
		Exit:  "Done",
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

func toolWorkflow() *ir.Workflow {
	return &ir.Workflow{
		Name:  "tool_test",
		Start: "Check",
		Exit:  "Report",
		Nodes: []*ir.Node{
			{
				ID: "Check", Kind: ir.NodeTool,
				IO: ir.NodeIO{Writes: []string{"test_result"}},
				Config: ir.ToolConfig{
					Command: "#!/bin/sh\nset -eu\npytest --tb=short",
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

func subgraphWorkflow() *ir.Workflow {
	return &ir.Workflow{
		Name:  "with_subgraph",
		Start: "Build",
		Exit:  "Done",
		Nodes: []*ir.Node{
			{ID: "Build", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "Build the feature."}},
			{
				ID: "Review", Kind: ir.NodeSubgraph,
				Config: ir.SubgraphConfig{
					Ref:    "./review.dip",
					Params: map[string]string{"model": "gpt-5.4"},
				},
			},
			{ID: "Done", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "Done."}},
		},
		Edges: []*ir.Edge{
			{From: "Build", To: "Review"},
			{From: "Review", To: "Done"},
		},
	}
}

// --- Test helpers ---

func assertContains(t *testing.T, output, substr string) {
	t.Helper()
	if !strings.Contains(output, substr) {
		t.Errorf("output does not contain %q\n\ngot:\n%s", substr, output)
	}
}

func assertNotContains(t *testing.T, output, substr string) {
	t.Helper()
	if strings.Contains(output, substr) {
		t.Errorf("output unexpectedly contains %q\n\ngot:\n%s", substr, output)
	}
}

// --- Tests ---

func TestExportDOTMinimal(t *testing.T) {
	out := ExportDOT(minimalWorkflow(), ExportOptions{})

	assertContains(t, out, "digraph minimal {")
	assertContains(t, out, "graph [rankdir=TB];")
	assertContains(t, out, `Begin [label="Begin", mode="freeform", shape="Mdiamond"];`)
	assertContains(t, out, `End [label="End", shape="Msquare"];`)
	assertContains(t, out, "Begin -> End;")
	assertContains(t, out, "}\n")
}

func TestExportDOTFullWorkflow(t *testing.T) {
	out := ExportDOT(askAndExecuteWorkflow(), ExportOptions{})

	// Verify digraph structure.
	assertContains(t, out, "digraph ask_and_execute {")

	// Start node (human) gets Mdiamond shape and emits mode.
	assertContains(t, out, `AskUser [label="AskUser", mode="freeform", shape="Mdiamond"];`)

	// Exit node gets Msquare shape.
	assertContains(t, out, `Done [label="Done", shape="Msquare"];`)

	// Regular agent node gets box shape.
	assertContains(t, out, `Interpret [label="Interpret", shape="box"];`)

	// Human node (non-start) emits mode and default as semantic attrs.
	assertContains(t, out, `Approve [default="Yes", label="Approve", mode="choice", shape="hexagon"];`)

	// Parallel node gets component shape and emits targets.
	assertContains(t, out, `ImplementFanOut [label="ImplementFanOut", shape="component", targets="ImplementClaude,ImplementCodex"];`)

	// Fan-in node gets tripleoctagon shape and emits sources.
	assertContains(t, out, `ImplementJoin [label="ImplementJoin", shape="tripleoctagon", sources="ImplementClaude,ImplementCodex"];`)

	// Simple edge.
	assertContains(t, out, "AskUser -> Interpret;")

	// Conditional edge.
	assertContains(t, out, `Validate -> Approve [condition="outcome = success", label="outcome = success"];`)

	// Restart edge with condition and label.
	assertContains(t, out, `Validate -> Interpret`)
	assertContains(t, out, `restart="true"`)
	assertContains(t, out, `style="dashed"`)
}

func TestExportDOTNodeShapes(t *testing.T) {
	tests := []struct {
		name     string
		kind     ir.NodeKind
		config   ir.NodeConfig
		wantAttr string
	}{
		{"agent", ir.NodeAgent, ir.AgentConfig{Prompt: "go."}, `shape="box"`},
		{"human", ir.NodeHuman, ir.HumanConfig{}, `shape="hexagon"`},
		{"tool", ir.NodeTool, ir.ToolConfig{Command: "echo"}, `shape="parallelogram"`},
		{"parallel", ir.NodeParallel, ir.ParallelConfig{Targets: []string{"A"}}, `shape="component"`},
		{"fan_in", ir.NodeFanIn, ir.FanInConfig{Sources: []string{"A"}}, `shape="tripleoctagon"`},
		{"subgraph", ir.NodeSubgraph, ir.SubgraphConfig{Ref: "x.dip"}, `shape="tab"`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &ir.Workflow{
				Name:  "test",
				Start: "Start",
				Exit:  "Exit",
				Nodes: []*ir.Node{
					{ID: "Start", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "go."}},
					{ID: "N", Kind: tt.kind, Config: tt.config},
					{ID: "Exit", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "done."}},
				},
			}
			out := ExportDOT(w, ExportOptions{})
			// N should get its kind-based shape (not overridden by start/exit).
			if !strings.Contains(out, tt.wantAttr) {
				t.Errorf("node N missing %s in output:\n%s", tt.wantAttr, out)
			}
		})
	}
}

func TestExportDOTStartExitShapeOverride(t *testing.T) {
	w := &ir.Workflow{
		Name:  "test",
		Start: "S",
		Exit:  "E",
		Nodes: []*ir.Node{
			{ID: "S", Kind: ir.NodeHuman, Config: ir.HumanConfig{}},
			{ID: "E", Kind: ir.NodeTool, Config: ir.ToolConfig{Command: "echo done"}},
		},
		Edges: []*ir.Edge{{From: "S", To: "E"}},
	}
	out := ExportDOT(w, ExportOptions{})

	// Start node should be Mdiamond regardless of kind.
	assertContains(t, out, `S [label="S", shape="Mdiamond"];`)
	// Exit node should be Msquare regardless of kind.
	assertContains(t, out, `E [label="E", shape="Msquare"];`)
}

func TestExportDOTNodeLabel(t *testing.T) {
	w := &ir.Workflow{
		Name:  "test",
		Start: "A",
		Exit:  "B",
		Nodes: []*ir.Node{
			{ID: "A", Kind: ir.NodeAgent, Label: "My Agent", Config: ir.AgentConfig{Prompt: "go."}},
			{ID: "B", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "done."}},
		},
	}
	out := ExportDOT(w, ExportOptions{})

	// Node with a label uses the label text.
	assertContains(t, out, `label="My Agent"`)
	// Node without a label uses the ID.
	assertContains(t, out, `B [label="B"`)
}

func TestExportDOTRankDir(t *testing.T) {
	tests := []struct {
		name    string
		rankDir string
		want    string
	}{
		{"default", "", "graph [rankdir=TB];"},
		{"LR", "LR", "graph [rankdir=LR];"},
		{"TB explicit", "TB", "graph [rankdir=TB];"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := minimalWorkflow()
			out := ExportDOT(w, ExportOptions{RankDir: tt.rankDir})
			assertContains(t, out, tt.want)
		})
	}
}

func TestExportDOTIncludePrompts(t *testing.T) {
	t.Run("prompts included", func(t *testing.T) {
		w := &ir.Workflow{
			Name:  "test",
			Start: "A",
			Exit:  "A",
			Nodes: []*ir.Node{
				{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{
					Prompt:   "Line one.\nLine two.",
					Model:    "gpt-5.4",
					Provider: "openai",
				}},
			},
		}
		out := ExportDOT(w, ExportOptions{IncludePrompts: true})
		assertContains(t, out, `model="gpt-5.4"`)
		assertContains(t, out, `prompt="Line one.`+`\n`+`Line two."`)
		assertContains(t, out, `provider="openai"`)
	})

	t.Run("prompts excluded by default", func(t *testing.T) {
		w := &ir.Workflow{
			Name:  "test",
			Start: "A",
			Exit:  "A",
			Nodes: []*ir.Node{
				{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{
					Prompt: "secret prompt",
					Model:  "gpt-5.4",
				}},
			},
		}
		out := ExportDOT(w, ExportOptions{})
		assertNotContains(t, out, "prompt=")
		assertNotContains(t, out, "model=")
	})
}

func TestExportDOTToolCommand(t *testing.T) {
	out := ExportDOT(toolWorkflow(), ExportOptions{IncludePrompts: true})
	// Multiline command should be escaped.
	assertContains(t, out, `tool_command="#!/bin/sh\nset -eu\npytest --tb=short"`)
	assertContains(t, out, `timeout="1m"`)
}

func TestExportDOTHumanConfig(t *testing.T) {
	w := &ir.Workflow{
		Name:  "test",
		Start: "H",
		Exit:  "H",
		Nodes: []*ir.Node{
			{ID: "H", Kind: ir.NodeHuman, Config: ir.HumanConfig{Mode: "choice", Default: "Yes"}},
		},
	}
	out := ExportDOT(w, ExportOptions{IncludePrompts: true})
	assertContains(t, out, `default="Yes"`)
	assertContains(t, out, `mode="choice"`)
}

func TestExportDOT_AgentToolAccess(t *testing.T) {
	w := &ir.Workflow{
		Name:  "X",
		Start: "A",
		Exit:  "A",
		Nodes: []*ir.Node{
			{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{
				Prompt:     "x",
				ToolAccess: "none",
			}},
		},
	}
	out := ExportDOT(w, ExportOptions{IncludePrompts: true})
	if !strings.Contains(out, `tool_access="none"`) && !strings.Contains(out, `tool_access=none`) {
		t.Errorf("DOT output missing tool_access attribute:\n%s", out)
	}
}

// TestExportDOT_AgentToolAccess_WithoutPrompts verifies that the runtime
// safety field survives a minimal (IncludePrompts: false) export. Tracker
// reads tool_access from graph.Attrs regardless of prompt inclusion; if the
// field were prompt-gated, a minimal export would silently drop it and the
// safety primitive wouldn't bind at runtime.
func TestExportDOT_AgentToolAccess_WithoutPrompts(t *testing.T) {
	w := &ir.Workflow{
		Name:  "X",
		Start: "A",
		Exit:  "A",
		Nodes: []*ir.Node{
			{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{
				Prompt:     "x",
				ToolAccess: "none",
			}},
		},
	}
	out := ExportDOT(w, ExportOptions{})
	if !strings.Contains(out, `tool_access="none"`) && !strings.Contains(out, `tool_access=none`) {
		t.Errorf("DOT output missing tool_access attribute under IncludePrompts: false:\n%s", out)
	}
	// Prompts must NOT appear (negative assertion confirms IncludePrompts: false honored).
	if strings.Contains(out, `prompt="x"`) {
		t.Errorf("prompt unexpectedly present under IncludePrompts: false:\n%s", out)
	}
}

func TestExportDOTSubgraphConfig(t *testing.T) {
	// Tests the export package's handling of un-flattened subgraph nodes.
	// In production, the CLI calls flatten.Flatten before ExportDOT,
	// so subgraph nodes should not appear in normal export-dot output.
	// This test verifies the fallback rendering for direct ExportDOT calls.
	out := ExportDOT(subgraphWorkflow(), ExportOptions{IncludePrompts: true})
	assertContains(t, out, `ref="./review.dip"`)
	assertContains(t, out, `shape="tab"`)
}

func TestExportDOTParallelConfig(t *testing.T) {
	w := &ir.Workflow{
		Name:  "test",
		Start: "S",
		Exit:  "J",
		Nodes: []*ir.Node{
			{ID: "S", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "go."}},
			{ID: "P", Kind: ir.NodeParallel, Config: ir.ParallelConfig{Targets: []string{"A", "B"}}},
			{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "a."}},
			{ID: "B", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "b."}},
			{ID: "J", Kind: ir.NodeFanIn, Config: ir.FanInConfig{Sources: []string{"A", "B"}}},
		},
	}
	out := ExportDOT(w, ExportOptions{IncludePrompts: true})
	assertContains(t, out, `targets="A,B"`)
	assertContains(t, out, `sources="A,B"`)
}

func TestExportDOTHighlightGoalGates(t *testing.T) {
	w := &ir.Workflow{
		Name:  "test",
		Start: "A",
		Exit:  "B",
		Nodes: []*ir.Node{
			{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "go.", GoalGate: true}},
			{ID: "B", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "done."}},
		},
	}

	t.Run("highlighting enabled", func(t *testing.T) {
		out := ExportDOT(w, ExportOptions{HighlightGoalGates: true})
		assertContains(t, out, `fillcolor="#ffcccc"`)
		assertContains(t, out, `style="filled"`)
	})

	t.Run("highlighting disabled", func(t *testing.T) {
		out := ExportDOT(w, ExportOptions{HighlightGoalGates: false})
		assertNotContains(t, out, "fillcolor")
		assertNotContains(t, out, `style="filled"`)
	})
}

func TestExportDOTEdgeConditions(t *testing.T) {
	tests := []struct {
		name      string
		condition *ir.Condition
		wantAttr  string
	}{
		{
			name: "simple compare",
			condition: &ir.Condition{
				Raw:    "ctx.outcome = success",
				Parsed: ir.CondCompare{Variable: "ctx.outcome", Op: "=", Value: "success"},
			},
			wantAttr: `condition="outcome = success"`,
		},
		{
			name: "AND condition",
			condition: &ir.Condition{
				Raw: "ctx.x = 1 and ctx.y = 2",
				Parsed: ir.CondAnd{
					Left:  ir.CondCompare{Variable: "ctx.x", Op: "=", Value: "1"},
					Right: ir.CondCompare{Variable: "ctx.y", Op: "=", Value: "2"},
				},
			},
			wantAttr: `condition="x = 1 and y = 2"`,
		},
		{
			name: "OR condition",
			condition: &ir.Condition{
				Raw: "ctx.x = 1 or ctx.x = 2",
				Parsed: ir.CondOr{
					Left:  ir.CondCompare{Variable: "ctx.x", Op: "=", Value: "1"},
					Right: ir.CondCompare{Variable: "ctx.x", Op: "=", Value: "2"},
				},
			},
			wantAttr: `condition="x = 1 or x = 2"`,
		},
		{
			name: "NOT condition",
			condition: &ir.Condition{
				Raw: "not ctx.done = true",
				Parsed: ir.CondNot{
					Inner: ir.CondCompare{Variable: "ctx.done", Op: "=", Value: "true"},
				},
			},
			wantAttr: `condition="not done = true"`,
		},
		{
			name: "nested AND in OR — parenthesized",
			condition: &ir.Condition{
				Raw: "(ctx.x = 1 and ctx.y = 2) or ctx.z = 3",
				Parsed: ir.CondOr{
					Left: ir.CondAnd{
						Left:  ir.CondCompare{Variable: "ctx.x", Op: "=", Value: "1"},
						Right: ir.CondCompare{Variable: "ctx.y", Op: "=", Value: "2"},
					},
					Right: ir.CondCompare{Variable: "ctx.z", Op: "=", Value: "3"},
				},
			},
			wantAttr: `condition="(x = 1 and y = 2) or z = 3"`,
		},
		{
			name: "NOT of compound — parenthesized",
			condition: &ir.Condition{
				Raw: "not (ctx.x = 1 and ctx.y = 2)",
				Parsed: ir.CondNot{
					Inner: ir.CondAnd{
						Left:  ir.CondCompare{Variable: "ctx.x", Op: "=", Value: "1"},
						Right: ir.CondCompare{Variable: "ctx.y", Op: "=", Value: "2"},
					},
				},
			},
			wantAttr: `condition="not (x = 1 and y = 2)"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &ir.Workflow{
				Name:  "test",
				Start: "A",
				Exit:  "B",
				Nodes: []*ir.Node{
					{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "go."}},
					{ID: "B", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "done."}},
				},
				Edges: []*ir.Edge{
					{From: "A", To: "B", Condition: tt.condition},
				},
			}
			out := ExportDOT(w, ExportOptions{})
			assertContains(t, out, tt.wantAttr)
		})
	}
}

func TestExportDOTEdgeRestart(t *testing.T) {
	w := &ir.Workflow{
		Name:  "test",
		Start: "A",
		Exit:  "B",
		Nodes: []*ir.Node{
			{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "go."}},
			{ID: "B", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "done."}},
		},
		Edges: []*ir.Edge{
			{From: "A", To: "B", Restart: true},
		},
	}
	out := ExportDOT(w, ExportOptions{})
	assertContains(t, out, `restart="true"`)
	assertContains(t, out, `style="dashed"`)
}

func TestExportDOTEdgeWeight(t *testing.T) {
	w := &ir.Workflow{
		Name:  "test",
		Start: "A",
		Exit:  "B",
		Nodes: []*ir.Node{
			{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "go."}},
			{ID: "B", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "done."}},
		},
		Edges: []*ir.Edge{
			{From: "A", To: "B", Weight: 10},
		},
	}
	out := ExportDOT(w, ExportOptions{})
	assertContains(t, out, `weight="10"`)
}

func TestExportDOTEdgeLabelWithCondition(t *testing.T) {
	// When an edge has both a label and a condition, the explicit label is used
	// and the condition is a separate attribute.
	w := &ir.Workflow{
		Name:  "test",
		Start: "A",
		Exit:  "B",
		Nodes: []*ir.Node{
			{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "go."}},
			{ID: "B", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "done."}},
		},
		Edges: []*ir.Edge{
			{
				From:  "A",
				To:    "B",
				Label: "retry",
				Condition: &ir.Condition{
					Raw:    "ctx.outcome = fail",
					Parsed: ir.CondCompare{Variable: "ctx.outcome", Op: "=", Value: "fail"},
				},
			},
		},
	}
	out := ExportDOT(w, ExportOptions{})
	assertContains(t, out, `label="retry"`)
	assertContains(t, out, `condition="outcome = fail"`)
}

func TestExportDOTEdgeConditionAsLabel(t *testing.T) {
	// When an edge has a condition but no explicit label, the condition
	// text is used as the label for visual display.
	w := &ir.Workflow{
		Name:  "test",
		Start: "A",
		Exit:  "B",
		Nodes: []*ir.Node{
			{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "go."}},
			{ID: "B", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "done."}},
		},
		Edges: []*ir.Edge{
			{
				From: "A",
				To:   "B",
				Condition: &ir.Condition{
					Raw:    "ctx.outcome = success",
					Parsed: ir.CondCompare{Variable: "ctx.outcome", Op: "=", Value: "success"},
				},
			},
		},
	}
	out := ExportDOT(w, ExportOptions{})
	assertContains(t, out, `label="outcome = success"`)
	assertContains(t, out, `condition="outcome = success"`)
}

func TestExportDOTEmptyWorkflow(t *testing.T) {
	w := &ir.Workflow{
		Name:  "empty",
		Start: "S",
		Exit:  "E",
	}
	out := ExportDOT(w, ExportOptions{})
	assertContains(t, out, "digraph empty {")
	assertContains(t, out, "}\n")
	// No node or edge lines (just the preamble and closing).
	assertNotContains(t, out, "->")
}

func TestExportDOTNoName(t *testing.T) {
	w := &ir.Workflow{
		Start: "S",
		Exit:  "E",
	}
	out := ExportDOT(w, ExportOptions{})
	assertContains(t, out, "digraph workflow {")
}

func TestExportDOTAllEdgeAttributes(t *testing.T) {
	w := &ir.Workflow{
		Name:  "test",
		Start: "A",
		Exit:  "B",
		Nodes: []*ir.Node{
			{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "go."}},
			{ID: "B", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "done."}},
		},
		Edges: []*ir.Edge{
			{
				From:    "A",
				To:      "B",
				Label:   "retry",
				Weight:  5,
				Restart: true,
				Condition: &ir.Condition{
					Raw:    "ctx.outcome = fail",
					Parsed: ir.CondCompare{Variable: "ctx.outcome", Op: "=", Value: "fail"},
				},
			},
		},
	}
	out := ExportDOT(w, ExportOptions{})
	assertContains(t, out, `condition="outcome = fail"`)
	assertContains(t, out, `label="retry"`)
	assertContains(t, out, `restart="true"`)
	assertContains(t, out, `style="dashed"`)
	assertContains(t, out, `weight="5"`)
}

func TestExportDOTExecutionPathOrder(t *testing.T) {
	w := &ir.Workflow{
		Name:  "test",
		Start: "A",
		Exit:  "C",
		Nodes: []*ir.Node{
			{ID: "A", Kind: ir.NodeAgent, Label: "Start Agent", Config: ir.AgentConfig{Prompt: "go."}},
			{ID: "B", Kind: ir.NodeAgent, Label: "Middle", Config: ir.AgentConfig{Prompt: "work."}},
			{ID: "C", Kind: ir.NodeAgent, Label: "End", Config: ir.AgentConfig{Prompt: "done."}},
		},
		Edges: []*ir.Edge{
			{From: "A", To: "B"},
			{From: "B", To: "C"},
		},
	}
	out := ExportDOT(w, ExportOptions{
		ExecutionPath: []string{"A", "B", "A", "C"},
	})
	// A visited twice: steps 1 and 3.
	assertContains(t, out, `[1,3] Start Agent`)
	assertContains(t, out, `[2] Middle`)
	assertContains(t, out, `[4] End`)
	assertContains(t, out, `fillcolor="#e0f0ff"`)
	assertContains(t, out, `style="bold,filled"`)
}

func TestExportDOTEmptyExecutionPath(t *testing.T) {
	w := minimalWorkflow()
	out := ExportDOT(w, ExportOptions{ExecutionPath: []string{}})
	// No execution order annotations.
	assertNotContains(t, out, "bold,filled")
}

func TestNodeShapeUnknownKind(t *testing.T) {
	// An unknown NodeKind should default to "box".
	w := &ir.Workflow{
		Name:  "test",
		Start: "S",
		Exit:  "E",
		Nodes: []*ir.Node{
			{ID: "S", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "go."}},
			{ID: "N", Kind: ir.NodeKind("unknown_kind"), Config: ir.AgentConfig{Prompt: "x."}},
			{ID: "E", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "done."}},
		},
	}
	out := ExportDOT(w, ExportOptions{})
	// N should get the default "box" shape.
	assertContains(t, out, `N [label="N", shape="box"];`)
}

func TestExportDOTEdgeConditionRawFallback(t *testing.T) {
	// When Parsed is nil, Raw should be used as the condition string.
	w := &ir.Workflow{
		Name:  "test",
		Start: "A",
		Exit:  "B",
		Nodes: []*ir.Node{
			{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "go."}},
			{ID: "B", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "done."}},
		},
		Edges: []*ir.Edge{
			{
				From: "A",
				To:   "B",
				Condition: &ir.Condition{
					Raw:    "ctx.outcome = success",
					Parsed: nil,
				},
			},
		},
	}
	out := ExportDOT(w, ExportOptions{})
	assertContains(t, out, `condition="outcome = success"`)
}

func TestFormatDurationSubSecond(t *testing.T) {
	// Sub-second durations should fall through to Go's default formatting.
	w := &ir.Workflow{
		Name:  "test",
		Start: "T",
		Exit:  "T",
		Nodes: []*ir.Node{
			{
				ID: "T", Kind: ir.NodeTool,
				Config: ir.ToolConfig{
					Command: "echo fast",
					Timeout: 500 * time.Millisecond,
				},
			},
		},
	}
	out := ExportDOT(w, ExportOptions{IncludePrompts: true})
	assertContains(t, out, `timeout="500ms"`)
}

func TestFormatDurationHoursMinutesSeconds(t *testing.T) {
	w := &ir.Workflow{
		Name:  "test",
		Start: "T",
		Exit:  "T",
		Nodes: []*ir.Node{
			{
				ID: "T", Kind: ir.NodeTool,
				Config: ir.ToolConfig{
					Command: "echo slow",
					Timeout: 1*time.Hour + 30*time.Minute + 15*time.Second,
				},
			},
		},
	}
	out := ExportDOT(w, ExportOptions{IncludePrompts: true})
	assertContains(t, out, `timeout="1h30m15s"`)
}

func TestExportDOTEdgeEmptyCondition(t *testing.T) {
	// An edge with a Condition but empty Raw and nil Parsed should not
	// produce a condition attribute.
	w := &ir.Workflow{
		Name:  "test",
		Start: "A",
		Exit:  "B",
		Nodes: []*ir.Node{
			{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "go."}},
			{ID: "B", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "done."}},
		},
		Edges: []*ir.Edge{
			{
				From:      "A",
				To:        "B",
				Condition: &ir.Condition{Raw: "", Parsed: nil},
			},
		},
	}
	out := ExportDOT(w, ExportOptions{})
	assertNotContains(t, out, "condition=")
}

func TestExportDOTIdempotent(t *testing.T) {
	workflows := []*ir.Workflow{
		minimalWorkflow(),
		askAndExecuteWorkflow(),
		toolWorkflow(),
		subgraphWorkflow(),
	}
	for _, w := range workflows {
		first := ExportDOT(w, ExportOptions{IncludePrompts: true})
		second := ExportDOT(w, ExportOptions{IncludePrompts: true})
		if first != second {
			t.Errorf("ExportDOT is not idempotent for workflow %q", w.Name)
		}
	}
}

func TestExportDOTDeterministicAttrOrder(t *testing.T) {
	// Attributes should always appear in sorted order.
	w := &ir.Workflow{
		Name:  "test",
		Start: "A",
		Exit:  "B",
		Nodes: []*ir.Node{
			{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "go."}},
			{ID: "B", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "done."}},
		},
		Edges: []*ir.Edge{
			{
				From:    "A",
				To:      "B",
				Label:   "go",
				Weight:  3,
				Restart: true,
			},
		},
	}
	out := ExportDOT(w, ExportOptions{})

	// Find the edge line.
	for _, line := range strings.Split(out, "\n") {
		if strings.Contains(line, "A -> B") {
			// Verify attribute order: label, restart, style, weight (alphabetical).
			labelIdx := strings.Index(line, "label=")
			restartIdx := strings.Index(line, "restart=")
			styleIdx := strings.Index(line, "style=")
			weightIdx := strings.Index(line, "weight=")
			if labelIdx > restartIdx || restartIdx > styleIdx || styleIdx > weightIdx {
				t.Errorf("attributes not in sorted order: %s", line)
			}
			return
		}
	}
	t.Error("could not find edge line A -> B")
}

func TestExportDOTSpecialCharactersInLabel(t *testing.T) {
	w := &ir.Workflow{
		Name:  "test",
		Start: "A",
		Exit:  "A",
		Nodes: []*ir.Node{
			{ID: "A", Kind: ir.NodeAgent, Label: `Say "hello" & goodbye`, Config: ir.AgentConfig{Prompt: "go."}},
		},
	}
	out := ExportDOT(w, ExportOptions{})
	// Internal quotes should be escaped.
	assertContains(t, out, `label="Say \"hello\" & goodbye"`)
}

func TestExportDOTValidDOTSyntax(t *testing.T) {
	// Basic structural validation: the output should be parseable as DOT.
	out := ExportDOT(askAndExecuteWorkflow(), ExportOptions{IncludePrompts: true})

	// Should start with digraph and end with }
	if !strings.HasPrefix(out, "digraph ") {
		t.Error("output should start with 'digraph '")
	}
	if !strings.HasSuffix(out, "}\n") {
		t.Error("output should end with '}\\n'")
	}

	// Count braces — should be balanced.
	opens := strings.Count(out, "{")
	closes := strings.Count(out, "}")
	if opens != closes {
		t.Errorf("unbalanced braces: %d opens, %d closes", opens, closes)
	}

	// Every node and edge statement should end with semicolon.
	for _, line := range strings.Split(out, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || trimmed == "}" || strings.HasPrefix(trimmed, "digraph") {
			continue
		}
		if !strings.HasSuffix(trimmed, ";") {
			t.Errorf("line does not end with semicolon: %q", trimmed)
		}
	}
}

func TestExportDOTNilConditionParsed(t *testing.T) {
	// Edge with Condition but nil Parsed should fall back to Raw string.
	w := &ir.Workflow{
		Name:  "test",
		Start: "A",
		Exit:  "B",
		Nodes: []*ir.Node{
			{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "go."}},
			{ID: "B", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "done."}},
		},
		Edges: []*ir.Edge{
			{From: "A", To: "B", Condition: &ir.Condition{Raw: "ctx.x = 1"}},
		},
	}
	out := ExportDOT(w, ExportOptions{})
	assertContains(t, out, `condition="x = 1"`)
}

func TestExportDOTNilConfig(t *testing.T) {
	// A node with nil config should not panic.
	w := &ir.Workflow{
		Name:  "test",
		Start: "A",
		Exit:  "A",
		Nodes: []*ir.Node{
			{ID: "A", Kind: ir.NodeAgent},
		},
	}
	out := ExportDOT(w, ExportOptions{IncludePrompts: true})
	assertContains(t, out, `A [`)
}

func TestExportDOTGoalGateNonAgent(t *testing.T) {
	// HighlightGoalGates should only apply to agent nodes with GoalGate: true.
	w := &ir.Workflow{
		Name:  "test",
		Start: "H",
		Exit:  "H",
		Nodes: []*ir.Node{
			{ID: "H", Kind: ir.NodeHuman, Config: ir.HumanConfig{Mode: "freeform"}},
		},
	}
	out := ExportDOT(w, ExportOptions{HighlightGoalGates: true})
	assertNotContains(t, out, "fillcolor")
}

func TestExportDOTNoSubgraphNodes(t *testing.T) {
	// After flattening, a workflow with subgraph refs should have no
	// shape="tab" nodes in the output. This test verifies that if the
	// caller passes an already-flattened workflow, no tab shapes appear.
	w := &ir.Workflow{
		Name: "flat", Start: "A", Exit: "B",
		Nodes: []*ir.Node{
			{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "go."}},
			{ID: "B", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "done."}},
		},
		Edges: []*ir.Edge{{From: "A", To: "B"}},
	}
	out := ExportDOT(w, ExportOptions{})
	assertNotContains(t, out, `shape="tab"`)
}

// --- Unit tests for internal helpers ---

func TestDotID(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"AskUser", "AskUser"},
		{"simple_name", "simple_name"},
		{"has space", `"has space"`},
		{"123start", `"123start"`},
		{"", `""`},
		{"with-dash", `"with-dash"`},
		{"with.dot", `"with.dot"`},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := dotID(tt.input)
			if got != tt.want {
				t.Errorf("dotID(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestDotQuote(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"hello", `"hello"`},
		{`say "hi"`, `"say \"hi\""`},
		{`path\to`, `"path\\to"`},
		{"", `""`},
		{`line1\nline2`, `"line1\nline2"`},         // DOT escape sequence preserved
		{`left\lalign`, `"left\lalign"`},           // DOT \l preserved
		{`real\\backslash`, `"real\\\\backslash"`}, // Actual backslash-backslash escaped
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := dotQuote(tt.input)
			if got != tt.want {
				t.Errorf("dotQuote(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestEscapeNewlines(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"no newlines", "no newlines"},
		{"line1\nline2", `line1\nline2`},
		{"a\nb\nc", `a\nb\nc`},
		{"", ""},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := escapeNewlines(tt.input)
			if got != tt.want {
				t.Errorf("escapeNewlines(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		input time.Duration
		want  string
	}{
		{30 * time.Second, "30s"},
		{5 * time.Minute, "5m"},
		{time.Hour, "1h"},
		{90 * time.Minute, "1h30m"},
		{0, "0s"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := formatDuration(tt.input)
			if got != tt.want {
				t.Errorf("formatDuration(%v) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestFormatConditionExport(t *testing.T) {
	tests := []struct {
		name string
		expr ir.ConditionExpr
		want string
	}{
		{
			name: "simple compare",
			expr: ir.CondCompare{Variable: "ctx.outcome", Op: "=", Value: "success"},
			want: "outcome = success",
		},
		{
			name: "AND",
			expr: ir.CondAnd{
				Left:  ir.CondCompare{Variable: "ctx.x", Op: "=", Value: "1"},
				Right: ir.CondCompare{Variable: "ctx.y", Op: "=", Value: "2"},
			},
			want: "x = 1 and y = 2",
		},
		{
			name: "OR",
			expr: ir.CondOr{
				Left:  ir.CondCompare{Variable: "ctx.x", Op: "=", Value: "1"},
				Right: ir.CondCompare{Variable: "ctx.x", Op: "=", Value: "2"},
			},
			want: "x = 1 or x = 2",
		},
		{
			name: "NOT",
			expr: ir.CondNot{Inner: ir.CondCompare{Variable: "ctx.done", Op: "=", Value: "true"}},
			want: "not done = true",
		},
		{
			name: "AND inside OR parenthesized",
			expr: ir.CondOr{
				Left: ir.CondAnd{
					Left:  ir.CondCompare{Variable: "ctx.x", Op: "=", Value: "1"},
					Right: ir.CondCompare{Variable: "ctx.y", Op: "=", Value: "2"},
				},
				Right: ir.CondCompare{Variable: "ctx.z", Op: "=", Value: "3"},
			},
			want: "(x = 1 and y = 2) or z = 3",
		},
		{
			name: "NOT of compound parenthesized",
			expr: ir.CondNot{
				Inner: ir.CondAnd{
					Left:  ir.CondCompare{Variable: "ctx.x", Op: "=", Value: "1"},
					Right: ir.CondCompare{Variable: "ctx.y", Op: "=", Value: "2"},
				},
			},
			want: "not (x = 1 and y = 2)",
		},
		{
			name: "nil",
			expr: nil,
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatCondition(tt.expr)
			if got != tt.want {
				t.Errorf("formatCondition() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestSortStrings(t *testing.T) {
	tests := []struct {
		input []string
		want  []string
	}{
		{[]string{"c", "a", "b"}, []string{"a", "b", "c"}},
		{[]string{"z"}, []string{"z"}},
		{nil, nil},
		{[]string{}, []string{}},
		{[]string{"a", "a"}, []string{"a", "a"}},
	}
	for _, tt := range tests {
		got := make([]string, len(tt.input))
		copy(got, tt.input)
		sortStrings(got)
		if len(got) != len(tt.want) {
			t.Errorf("sortStrings(%v) length = %d, want %d", tt.input, len(got), len(tt.want))
			continue
		}
		for i := range got {
			if got[i] != tt.want[i] {
				t.Errorf("sortStrings(%v)[%d] = %q, want %q", tt.input, i, got[i], tt.want[i])
			}
		}
	}
}

func TestIsSimpleDOTID(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"AskUser", true},
		{"node_1", true},
		{"A", true},
		{"123", false},
		{"has space", false},
		{"has-dash", false},
		{"has.dot", false},
		{"", false},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := isSimpleDOTID(tt.input)
			if got != tt.want {
				t.Errorf("isSimpleDOTID(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestFormatDOTAttrs(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		got := formatDOTAttrs(map[string]string{})
		if got != "" {
			t.Errorf("formatDOTAttrs({}) = %q, want empty", got)
		}
	})

	t.Run("single", func(t *testing.T) {
		got := formatDOTAttrs(map[string]string{"shape": "box"})
		if got != `[shape="box"]` {
			t.Errorf(`formatDOTAttrs = %q, want [shape="box"]`, got)
		}
	})

	t.Run("sorted keys", func(t *testing.T) {
		got := formatDOTAttrs(map[string]string{"z": "1", "a": "2"})
		// a should come before z.
		aIdx := strings.Index(got, "a=")
		zIdx := strings.Index(got, "z=")
		if aIdx > zIdx {
			t.Errorf("keys not sorted: %s", got)
		}
	})
}

func TestExportConditionalNodeDiamond(t *testing.T) {
	w := &ir.Workflow{
		Name:  "cond_test",
		Start: "Start",
		Exit:  "Done",
		Nodes: []*ir.Node{
			{ID: "Start", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "Go."}},
			{ID: "Route", Kind: ir.NodeConditional, Label: "Branch", Config: ir.ConditionalConfig{}},
			{ID: "Done", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "Done."}},
		},
		Edges: []*ir.Edge{
			{From: "Start", To: "Route"},
			{From: "Route", To: "Done"},
		},
	}
	dot := ExportDOT(w, ExportOptions{})
	if !strings.Contains(dot, `shape="diamond"`) && !strings.Contains(dot, `shape=diamond`) {
		t.Errorf("expected diamond shape for conditional node, got:\n%s", dot)
	}
}

func TestExportVarsAsGraphAttrs(t *testing.T) {
	w := &ir.Workflow{
		Name:  "Test",
		Start: "A",
		Exit:  "B",
		Vars: map[string]string{
			"env":     "production",
			"api_url": "https://example.com/api",
		},
		Nodes: []*ir.Node{
			{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "do it"}},
			{ID: "B", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "done"}},
		},
		Edges: []*ir.Edge{{From: "A", To: "B"}},
	}
	out := ExportDOT(w, ExportOptions{})

	if !strings.Contains(out, `api_url="https://example.com/api"`) {
		t.Errorf("expected api_url graph attr, got:\n%s", out)
	}
	if !strings.Contains(out, `env="production"`) {
		t.Errorf("expected env graph attr, got:\n%s", out)
	}
	// Sorted: api_url must appear before env
	apiIdx := strings.Index(out, "api_url=")
	envIdx := strings.Index(out, "env=")
	if apiIdx < 0 || envIdx < 0 {
		t.Fatalf("keys missing from output:\n%s", out)
	}
	if apiIdx > envIdx {
		t.Errorf("vars not sorted: api_url@%d should precede env@%d\n%s", apiIdx, envIdx, out)
	}
}

func TestExportDOT_ManagerLoop(t *testing.T) {
	w := &ir.Workflow{
		Name:  "W",
		Start: "S",
		Exit:  "E",
		Nodes: []*ir.Node{
			{ID: "S", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "start"}},
			{ID: "M", Kind: ir.NodeManagerLoop, Label: "Supervisor", Config: ir.ManagerLoopConfig{
				SubgraphRef:    "inner",
				PollInterval:   30 * time.Second,
				MaxCycles:      12,
				StopCondition:  &ir.Condition{Raw: "stack.child.cycles = 10"},
				SteerCondition: &ir.Condition{Raw: "stack.child.cycles = 5"},
				SteerContext:   map[string]string{"hint": "speed_up", "priority": "high"},
			}},
			{ID: "E", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "end"}},
		},
		Edges: []*ir.Edge{
			{From: "S", To: "M"},
			{From: "M", To: "E"},
		},
	}
	out := ExportDOT(w, ExportOptions{})
	assertContains(t, out, `shape="house"`)
	assertContains(t, out, `subgraph_ref="inner"`)
	assertContains(t, out, `poll_interval="30s"`)
	assertContains(t, out, `max_cycles="12"`)
	assertContains(t, out, `stop_condition="stack.child.cycles = 10"`)
	assertContains(t, out, `steer_condition="stack.child.cycles = 5"`)
	// steer_context flattens to canonical sorted "k=v,k=v"
	assertContains(t, out, `steer_context="hint=speed_up,priority=high"`)
}

func TestExportDOT_ManagerLoop_MinimalOmitsEmptyAttrs(t *testing.T) {
	// A manager_loop with only subgraph_ref set should emit shape=house and
	// subgraph_ref but no poll_interval, max_cycles, stop/steer_condition,
	// or steer_context attrs. Exercises the guard clauses in
	// applyManagerLoopScalarAttrs and applyManagerLoopConditionAttrs.
	w := &ir.Workflow{
		Name:  "W",
		Start: "S",
		Exit:  "E",
		Nodes: []*ir.Node{
			{ID: "S", Kind: ir.NodeAgent, Config: ir.AgentConfig{}},
			{ID: "M", Kind: ir.NodeManagerLoop, Config: ir.ManagerLoopConfig{
				SubgraphRef: "inner",
			}},
			{ID: "E", Kind: ir.NodeAgent, Config: ir.AgentConfig{}},
		},
		Edges: []*ir.Edge{{From: "S", To: "M"}, {From: "M", To: "E"}},
	}
	out := ExportDOT(w, ExportOptions{})
	assertContains(t, out, `shape="house"`)
	assertContains(t, out, `subgraph_ref="inner"`)
	for _, attr := range []string{"poll_interval", "max_cycles", "stop_condition", "steer_condition", "steer_context"} {
		if strings.Contains(out, attr+"=") {
			t.Errorf("unexpected %s attr in minimal output:\n%s", attr, out)
		}
	}
}

func TestExportVarsSkipsDefaultsCollision(t *testing.T) {
	w := &ir.Workflow{
		Name:  "Test",
		Start: "A",
		Exit:  "B",
		Vars: map[string]string{
			"model":   "should-be-skipped",
			"env":     "staging",
			"rankdir": "also-skipped",
		},
		Nodes: []*ir.Node{
			{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "do it"}},
			{ID: "B", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "done"}},
		},
		Edges: []*ir.Edge{{From: "A", To: "B"}},
	}
	out := ExportDOT(w, ExportOptions{})

	// "model" and "rankdir" are reserved — they must not appear as extra graph attrs
	if strings.Contains(out, `model="should-be-skipped"`) {
		t.Errorf("reserved key 'model' should be skipped, got:\n%s", out)
	}
	if strings.Contains(out, `rankdir="also-skipped"`) {
		t.Errorf("reserved key 'rankdir' should be skipped, got:\n%s", out)
	}
	// Non-reserved key must appear
	if !strings.Contains(out, `env="staging"`) {
		t.Errorf("expected env graph attr, got:\n%s", out)
	}
}

func TestExportToolSafetyDefaults(t *testing.T) {
	w := &ir.Workflow{
		Name:  "ToolSafety",
		Start: "A",
		Exit:  "B",
		Defaults: ir.WorkflowDefaults{
			ToolCommandsAllow: "git *,make *",
			ToolDenylistAdd:   "rm -rf /",
		},
		Nodes: []*ir.Node{
			{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "a"}},
			{ID: "B", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "b"}},
		},
		Edges: []*ir.Edge{{From: "A", To: "B"}},
	}
	out := ExportDOT(w, ExportOptions{})

	if !strings.Contains(out, `tool_commands_allow="git *,make *"`) {
		t.Errorf("expected tool_commands_allow graph attr, got:\n%s", out)
	}
	if !strings.Contains(out, `tool_denylist_add="rm -rf /"`) {
		t.Errorf("expected tool_denylist_add graph attr, got:\n%s", out)
	}
}

func TestExportToolSafetyDefaultsOmitEmpty(t *testing.T) {
	w := &ir.Workflow{
		Name:  "Empty",
		Start: "A",
		Exit:  "B",
		Nodes: []*ir.Node{
			{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "a"}},
			{ID: "B", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "b"}},
		},
		Edges: []*ir.Edge{{From: "A", To: "B"}},
	}
	out := ExportDOT(w, ExportOptions{})

	if strings.Contains(out, "tool_commands_allow") {
		t.Errorf("empty field should not emit tool_commands_allow:\n%s", out)
	}
	if strings.Contains(out, "tool_denylist_add") {
		t.Errorf("empty field should not emit tool_denylist_add:\n%s", out)
	}
}

func TestExportToolSafetyVarsCollision(t *testing.T) {
	// A user mis-specifies the keys in vars:; export sets them from Defaults and
	// must not also emit them from Vars (double-emit would be invalid DOT).
	w := &ir.Workflow{
		Name:  "Collision",
		Start: "A",
		Exit:  "B",
		Defaults: ir.WorkflowDefaults{
			ToolCommandsAllow: "git *",
		},
		Vars: map[string]string{
			"tool_commands_allow": "should-be-skipped",
			"tool_denylist_add":   "also-skipped",
		},
		Nodes: []*ir.Node{
			{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "a"}},
			{ID: "B", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "b"}},
		},
		Edges: []*ir.Edge{{From: "A", To: "B"}},
	}
	out := ExportDOT(w, ExportOptions{})

	if strings.Contains(out, `tool_commands_allow="should-be-skipped"`) {
		t.Errorf("vars 'tool_commands_allow' should be skipped, got:\n%s", out)
	}
	if strings.Contains(out, `tool_denylist_add="also-skipped"`) {
		t.Errorf("vars 'tool_denylist_add' should be skipped, got:\n%s", out)
	}
	if !strings.Contains(out, `tool_commands_allow="git *"`) {
		t.Errorf("expected tool_commands_allow from Defaults, got:\n%s", out)
	}
}

func TestExportDOTToolRoutingFields(t *testing.T) {
	wf := &ir.Workflow{
		Name:  "tool_routing_test",
		Goal:  "test routing",
		Start: "T",
		Exit:  "T",
		Nodes: []*ir.Node{
			{
				ID:   "T",
				Kind: ir.NodeTool,
				Config: ir.ToolConfig{
					Command:       "echo pass",
					Timeout:       30 * time.Second,
					MarkerGrep:    "^pass$",
					RouteRequired: true,
					OutputLimit:   8192,
				},
			},
		},
		Edges: []*ir.Edge{},
	}
	out := ExportDOT(wf, ExportOptions{})
	for _, want := range []string{
		`marker_grep="^pass$"`,
		`route_required="true"`,
		`output_limit="8192"`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("DOT output missing %q. Full output:\n%s", want, out)
		}
	}
}

func TestExportDOTToolRoutingOmitWhenZero(t *testing.T) {
	wf := &ir.Workflow{
		Name:  "tool_omit_test",
		Goal:  "test omission",
		Start: "T",
		Exit:  "T",
		Nodes: []*ir.Node{
			{
				ID:   "T",
				Kind: ir.NodeTool,
				Config: ir.ToolConfig{
					Command: "echo pass",
					Timeout: 30 * time.Second,
					// MarkerGrep, RouteRequired, OutputLimit all zero/default
				},
			},
		},
		Edges: []*ir.Edge{},
	}
	out := ExportDOT(wf, ExportOptions{})
	for _, unwanted := range []string{"marker_grep", "route_required", "output_limit"} {
		if strings.Contains(out, unwanted) {
			t.Errorf("DOT output unexpectedly contains %q. Full output:\n%s", unwanted, out)
		}
	}
}
