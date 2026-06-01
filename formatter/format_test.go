package formatter

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/2389-research/dippin-lang/ir"
	"github.com/2389-research/dippin-lang/parser"
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

func assertIdempotent(t *testing.T, w *ir.Workflow) {
	t.Helper()
	first := Format(w)
	second := Format(w)
	if first != second {
		t.Errorf("Format is not idempotent:\nfirst:\n%s\nsecond:\n%s", first, second)
	}
}

// --- Table-driven tests ---

func TestFormatHappyPath(t *testing.T) {
	tests := []struct {
		name     string
		workflow *ir.Workflow
		checks   []string // substrings that must appear
	}{
		{
			name:     "minimal workflow",
			workflow: minimalWorkflow(),
			checks: []string{
				"workflow minimal",
				"  start: Begin",
				"  exit: End",
				"  human Begin",
				"    mode: freeform",
				"  agent End",
				"    prompt:",
				"      Done.",
				"  edges",
				"    Begin -> End",
			},
		},
		{
			name:     "full ask_and_execute",
			workflow: askAndExecuteWorkflow(),
			checks: []string{
				"workflow ask_and_execute",
				`  goal: "Ask user for a task, implement it, review, ship"`,
				"  start: AskUser",
				"  exit: Done",
				"  defaults",
				"    model: claude-opus-4-6",
				"    provider: anthropic",
				"    retry_policy: standard",
				"    fidelity: summary:high",
				"    max_restarts: 5",
				"  human AskUser",
				"  agent Interpret",
				"    reads: human_response",
				"    writes: plan",
				"    prompt:",
				"      You are a senior software architect.",
				"      ${ctx.human_response}",
				"  parallel ImplementFanOut -> ImplementClaude, ImplementCodex",
				"  agent ImplementClaude",
				"    model: gpt-5.4",
				"    provider: openai",
				"  agent ImplementCodex",
				"    model: gpt-5.3-codex",
				"  fan_in ImplementJoin <- ImplementClaude, ImplementCodex",
				"  agent Validate",
				"    goal_gate: true",
				"    auto_status: true",
				"    max_retries: 2",
				"  human Approve",
				"    mode: choice",
				"    default: Yes",
				"  agent Done",
				"  edges",
				"    AskUser -> Interpret",
				"    Validate -> Approve  when ctx.outcome = success",
				`    Validate -> Interpret  when ctx.outcome = fail  label: retry  restart: true`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := Format(tt.workflow)
			for _, check := range tt.checks {
				assertContains(t, output, check)
			}
			assertIdempotent(t, tt.workflow)
		})
	}
}

func TestFormatAgentAllFields(t *testing.T) {
	w := &ir.Workflow{
		Name:  "agent_all_fields",
		Start: "A",
		Exit:  "A",
		Nodes: []*ir.Node{
			{
				ID:      "A",
				Kind:    ir.NodeAgent,
				Label:   "My Agent Node",
				Classes: []string{"primary", "reviewer"},
				Config: ir.AgentConfig{
					Model:           "gpt-5.4",
					Provider:        "openai",
					ReasoningEffort: "high",
					Fidelity:        "full",
					GoalGate:        true,
					AutoStatus:      true,
					MaxTurns:        10,
					Prompt:          "Do the thing.\nDo it well.",
				},
				Retry: ir.RetryConfig{
					Policy:      "aggressive",
					MaxRetries:  5,
					RetryTarget: "Fallback",
				},
				IO: ir.NodeIO{
					Reads:  []string{"input", "context"},
					Writes: []string{"output"},
				},
			},
		},
	}

	output := Format(w)

	// Verify canonical order: label, class, model, provider, reasoning_effort,
	// fidelity, goal_gate, auto_status, max_turns, retry_policy, max_retries,
	// retry_target, reads, writes, prompt
	lines := strings.Split(output, "\n")
	fieldOrder := []string{
		"label:", "class:", "model:", "provider:", "reasoning_effort:",
		"fidelity:", "goal_gate:", "auto_status:", "max_turns:",
		"retry_policy:", "max_retries:", "retry_target:",
		"reads:", "writes:", "prompt:",
	}
	lastIdx := -1
	for _, field := range fieldOrder {
		for i, l := range lines {
			if strings.Contains(strings.TrimSpace(l), field) {
				if i <= lastIdx {
					t.Errorf("field %q at line %d appears before or at previous field line %d", field, i, lastIdx)
				}
				lastIdx = i
				break
			}
		}
	}

	assertContains(t, output, `label: "My Agent Node"`)
	assertContains(t, output, "class: primary, reviewer")
	assertContains(t, output, "model: gpt-5.4")
	assertContains(t, output, "provider: openai")
	assertContains(t, output, "reasoning_effort: high")
	assertContains(t, output, "fidelity: full")
	assertContains(t, output, "goal_gate: true")
	assertContains(t, output, "auto_status: true")
	assertContains(t, output, "max_turns: 10")
	assertContains(t, output, "retry_policy: aggressive")
	assertContains(t, output, "max_retries: 5")
	assertContains(t, output, "retry_target: Fallback")
	assertContains(t, output, "reads: input, context")
	assertContains(t, output, "writes: output")
	assertContains(t, output, "prompt:")
	assertContains(t, output, "Do the thing.")
	assertIdempotent(t, w)
}

func TestFormatHumanAllFields(t *testing.T) {
	w := &ir.Workflow{
		Name:  "human_all",
		Start: "H",
		Exit:  "H",
		Nodes: []*ir.Node{
			{
				ID:    "H",
				Kind:  ir.NodeHuman,
				Label: "User Input",
				Config: ir.HumanConfig{
					Mode:    "choice",
					Default: "Yes",
				},
				IO: ir.NodeIO{
					Reads:  []string{"prev_output"},
					Writes: []string{"human_response"},
				},
			},
		},
	}

	output := Format(w)
	assertContains(t, output, `label: "User Input"`)
	assertContains(t, output, "mode: choice")
	assertContains(t, output, "default: Yes")
	assertContains(t, output, "reads: prev_output")
	assertContains(t, output, "writes: human_response")
	assertIdempotent(t, w)
}

func TestFormatHumanWithPrompt(t *testing.T) {
	prompt := "Please review the proposed changes below.\n\nIf everything looks correct, approve to continue.\nOtherwise, reject and provide feedback."
	w := &ir.Workflow{
		Name:  "human_prompt",
		Start: "Gate",
		Exit:  "Gate",
		Nodes: []*ir.Node{
			{
				ID:   "Gate",
				Kind: ir.NodeHuman,
				Config: ir.HumanConfig{
					Mode:   "choice",
					Prompt: prompt,
				},
			},
		},
	}

	output := Format(w)
	assertContains(t, output, "mode: choice")
	assertContains(t, output, "prompt:")
	assertContains(t, output, "Please review the proposed changes below.")
	assertContains(t, output, "If everything looks correct, approve to continue.")
	assertIdempotent(t, w)
}

func TestFormatHumanPromptRoundtrip(t *testing.T) {
	input := `workflow HumanPromptRT
  start: Gate
  exit: Done

  human Gate
    mode: choice
    prompt:
      Review the changes below.

      Approve or reject.

  agent Done
    prompt: "Ship it."

  edges
    Gate -> Done
`
	w1, err := parser.NewParser(input, "test.dip").Parse()
	if err != nil {
		t.Fatalf("first parse failed: %v", err)
	}

	formatted := Format(w1)
	w2, err := parser.NewParser(formatted, "test.dip").Parse()
	if err != nil {
		t.Fatalf("second parse failed: %v\nformatted:\n%s", err, formatted)
	}

	reformatted := Format(w2)
	if formatted != reformatted {
		t.Errorf("formatter not idempotent\nfirst:\n%s\nsecond:\n%s", formatted, reformatted)
	}

	cfg := w2.Nodes[0].Config.(ir.HumanConfig)
	if cfg.Prompt == "" {
		t.Fatal("prompt is empty after round-trip")
	}
	if cfg.Mode != "choice" {
		t.Errorf("mode = %q, want choice", cfg.Mode)
	}
}

func TestFormatToolMultilineCommand(t *testing.T) {
	w := &ir.Workflow{
		Name:  "tool_test",
		Start: "T",
		Exit:  "T",
		Nodes: []*ir.Node{
			{
				ID:   "T",
				Kind: ir.NodeTool,
				Config: ir.ToolConfig{
					Command: "#!/bin/sh\nset -eu\nif pytest --tb=short 2>&1; then\n  printf 'pass'\nelse\n  printf 'fail'\n  exit 1\nfi",
					Timeout: 60 * time.Second,
				},
				IO: ir.NodeIO{Writes: []string{"test_result"}},
			},
		},
	}

	output := Format(w)
	assertContains(t, output, "timeout: 1m")
	assertContains(t, output, "writes: test_result")
	assertContains(t, output, "command:")
	assertContains(t, output, "      #!/bin/sh")
	assertContains(t, output, "      set -eu")
	assertContains(t, output, "      if pytest --tb=short 2>&1; then")
	assertContains(t, output, "        printf 'pass'")
	assertContains(t, output, "        printf 'fail'")
	assertIdempotent(t, w)
}

func TestFormatFieldOrdering(t *testing.T) {
	t.Run("prompt is always last", func(t *testing.T) {
		w := &ir.Workflow{
			Name:  "prompt_last",
			Start: "A",
			Exit:  "A",
			Nodes: []*ir.Node{
				{
					ID:   "A",
					Kind: ir.NodeAgent,
					Config: ir.AgentConfig{
						Prompt:   "Do work.",
						Model:    "gpt-5.4",
						Provider: "openai",
					},
					IO: ir.NodeIO{Reads: []string{"input"}},
				},
			},
		}

		output := Format(w)
		promptIdx := strings.Index(output, "prompt:")
		readsIdx := strings.Index(output, "reads:")
		modelIdx := strings.Index(output, "model:")
		if promptIdx < readsIdx {
			t.Error("prompt should appear after reads")
		}
		if promptIdx < modelIdx {
			t.Error("prompt should appear after model")
		}
	})

	t.Run("command is always last for tool", func(t *testing.T) {
		w := &ir.Workflow{
			Name:  "cmd_last",
			Start: "T",
			Exit:  "T",
			Nodes: []*ir.Node{
				{
					ID:   "T",
					Kind: ir.NodeTool,
					Config: ir.ToolConfig{
						Command: "echo hello",
						Timeout: 30 * time.Second,
					},
					IO: ir.NodeIO{Writes: []string{"out"}},
				},
			},
		}
		output := Format(w)
		commandIdx := strings.Index(output, "command:")
		timeoutIdx := strings.Index(output, "timeout:")
		writesIdx := strings.Index(output, "writes:")
		if commandIdx < timeoutIdx {
			t.Error("command should appear after timeout")
		}
		if commandIdx < writesIdx {
			t.Error("command should appear after writes")
		}
	})
}

func TestFormatMultilineContent(t *testing.T) {
	tests := []struct {
		name     string
		prompt   string
		contains []string
		excludes []string
	}{
		{
			name:   "blank lines preserved",
			prompt: "First line.\n\nThird line.",
			contains: []string{
				"      First line.",
				"      Third line.",
			},
		},
		{
			name:   "variable references preserved",
			prompt: "Process ${ctx.human_response} now.\nUse ${graph.goal} as guide.",
			contains: []string{
				"${ctx.human_response}",
				"${graph.goal}",
			},
		},
		{
			name:   "trailing whitespace stripped",
			prompt: "line with spaces   \nnext line",
			contains: []string{
				"      line with spaces",
				"      next line",
			},
			excludes: []string{
				"line with spaces   ",
			},
		},
		{
			name:   "trailing blank lines in prompt stripped",
			prompt: "content\n\n\n",
			contains: []string{
				"      content",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &ir.Workflow{
				Name:  "test",
				Start: "A",
				Exit:  "A",
				Nodes: []*ir.Node{
					{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: tt.prompt}},
				},
			}
			output := Format(w)
			for _, c := range tt.contains {
				assertContains(t, output, c)
			}
			for _, e := range tt.excludes {
				assertNotContains(t, output, e)
			}
		})
	}
}

func TestFormatEdges(t *testing.T) {
	tests := []struct {
		name   string
		edges  []*ir.Edge
		checks []string
	}{
		{
			name: "simple edge",
			edges: []*ir.Edge{
				{From: "A", To: "B"},
			},
			checks: []string{"    A -> B"},
		},
		{
			name: "conditional edge",
			edges: []*ir.Edge{
				{From: "A", To: "B", Condition: &ir.Condition{
					Raw:    "ctx.outcome = success",
					Parsed: ir.CondCompare{Variable: "ctx.outcome", Op: "=", Value: "success"},
				}},
			},
			checks: []string{"    A -> B  when ctx.outcome = success"},
		},
		{
			name: "edge with all attributes",
			edges: []*ir.Edge{
				{
					From:  "A",
					To:    "B",
					Label: "retry",
					Condition: &ir.Condition{
						Raw:    "ctx.outcome = fail",
						Parsed: ir.CondCompare{Variable: "ctx.outcome", Op: "=", Value: "fail"},
					},
					Weight:  10,
					Restart: true,
				},
			},
			checks: []string{
				`    A -> B  when ctx.outcome = fail  label: retry  weight: 10  restart: true`,
			},
		},
		{
			name: "complex condition with AND",
			edges: []*ir.Edge{
				{From: "A", To: "B", Condition: &ir.Condition{
					Raw: "ctx.outcome = success and ctx.tool_stdout != empty",
					Parsed: ir.CondAnd{
						Left:  ir.CondCompare{Variable: "ctx.outcome", Op: "=", Value: "success"},
						Right: ir.CondCompare{Variable: "ctx.tool_stdout", Op: "!=", Value: "empty"},
					},
				}},
			},
			checks: []string{
				"    A -> B  when ctx.outcome = success and ctx.tool_stdout != empty",
			},
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
				Edges: tt.edges,
			}
			output := Format(w)
			for _, check := range tt.checks {
				assertContains(t, output, check)
			}
		})
	}
}

func TestFormatDefaults(t *testing.T) {
	t.Run("defaults with some fields", func(t *testing.T) {
		w := &ir.Workflow{
			Name:  "test",
			Start: "A",
			Exit:  "A",
			Defaults: ir.WorkflowDefaults{
				Model:       "gpt-5.4",
				Provider:    "openai",
				MaxRetries:  3,
				MaxRestarts: 5,
			},
			Nodes: []*ir.Node{
				{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "go."}},
			},
		}
		output := Format(w)
		assertContains(t, output, "  defaults")
		assertContains(t, output, "    model: gpt-5.4")
		assertContains(t, output, "    provider: openai")
		assertContains(t, output, "    max_retries: 3")
		assertContains(t, output, "    max_restarts: 5")
		// Fields not set should not appear
		assertNotContains(t, output, "retry_policy:")
		assertNotContains(t, output, "fidelity:")
		assertNotContains(t, output, "cache_tools:")
		assertNotContains(t, output, "compaction:")
	})

	t.Run("no defaults omits block", func(t *testing.T) {
		w := &ir.Workflow{
			Name:  "test",
			Start: "A",
			Exit:  "A",
			Nodes: []*ir.Node{
				{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "go."}},
			},
		}
		output := Format(w)
		assertNotContains(t, output, "defaults")
	})
}

func TestFormatSpecialCases(t *testing.T) {
	t.Run("empty workflow", func(t *testing.T) {
		w := &ir.Workflow{
			Name:  "empty",
			Start: "S",
			Exit:  "E",
		}
		output := Format(w)
		assertContains(t, output, "workflow empty")
		assertContains(t, output, "  start: S")
		assertContains(t, output, "  exit: E")
		// Should have exactly one trailing newline
		if !strings.HasSuffix(output, "\n") {
			t.Error("output should end with newline")
		}
		trimmed := strings.TrimRight(output, "\n")
		if strings.HasSuffix(trimmed, "\n") {
			t.Error("output should have exactly one trailing newline")
		}
	})

	t.Run("parallel and fan_in inline", func(t *testing.T) {
		w := &ir.Workflow{
			Name:  "par",
			Start: "P",
			Exit:  "J",
			Nodes: []*ir.Node{
				{ID: "P", Kind: ir.NodeParallel, Config: ir.ParallelConfig{Targets: []string{"A", "B", "C"}}},
				{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "a."}},
				{ID: "B", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "b."}},
				{ID: "C", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "c."}},
				{ID: "J", Kind: ir.NodeFanIn, Config: ir.FanInConfig{Sources: []string{"A", "B", "C"}}},
			},
		}
		output := Format(w)
		assertContains(t, output, "  parallel P -> A, B, C")
		assertContains(t, output, "  fan_in J <- A, B, C")
	})

	t.Run("subgraph with params", func(t *testing.T) {
		w := &ir.Workflow{
			Name:  "sub",
			Start: "S",
			Exit:  "S",
			Nodes: []*ir.Node{
				{
					ID:   "S",
					Kind: ir.NodeSubgraph,
					Config: ir.SubgraphConfig{
						Ref:    "./review.dip",
						Params: map[string]string{"model": "gpt-5.4", "strict": "true"},
					},
				},
			},
		}
		output := Format(w)
		assertContains(t, output, "  subgraph S")
		assertContains(t, output, "    ref: ./review.dip")
		assertContains(t, output, "    params:")
		assertContains(t, output, "      model: gpt-5.4")
		assertContains(t, output, "      strict: true")
	})

	t.Run("idempotency", func(t *testing.T) {
		assertIdempotent(t, askAndExecuteWorkflow())
		assertIdempotent(t, minimalWorkflow())
	})
}

func TestFormatEdgeCases(t *testing.T) {
	t.Run("zero-value agent config", func(t *testing.T) {
		w := &ir.Workflow{
			Name:  "test",
			Start: "A",
			Exit:  "A",
			Nodes: []*ir.Node{
				{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{}},
			},
		}
		output := Format(w)
		assertContains(t, output, "  agent A")
		// No fields should be emitted for zero-value config
		assertNotContains(t, output, "model:")
		assertNotContains(t, output, "prompt:")
		assertNotContains(t, output, "goal_gate:")
		assertNotContains(t, output, "auto_status:")
	})

	t.Run("workflow with goal", func(t *testing.T) {
		w := &ir.Workflow{
			Name:  "test",
			Goal:  "Do something useful",
			Start: "A",
			Exit:  "A",
			Nodes: []*ir.Node{
				{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "go."}},
			},
		}
		output := Format(w)
		assertContains(t, output, `  goal: "Do something useful"`)
	})

	t.Run("workflow without goal", func(t *testing.T) {
		w := &ir.Workflow{
			Name:  "test",
			Start: "A",
			Exit:  "A",
			Nodes: []*ir.Node{
				{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "go."}},
			},
		}
		output := Format(w)
		assertNotContains(t, output, "goal:")
	})

	t.Run("node with classes", func(t *testing.T) {
		w := &ir.Workflow{
			Name:  "test",
			Start: "A",
			Exit:  "A",
			Nodes: []*ir.Node{
				{
					ID:      "A",
					Kind:    ir.NodeAgent,
					Classes: []string{"cls1", "cls2"},
					Config:  ir.AgentConfig{Prompt: "go."},
				},
			},
		}
		output := Format(w)
		assertContains(t, output, "    class: cls1, cls2")
	})

	t.Run("reads and writes with multiple keys", func(t *testing.T) {
		w := &ir.Workflow{
			Name:  "test",
			Start: "A",
			Exit:  "A",
			Nodes: []*ir.Node{
				{
					ID:   "A",
					Kind: ir.NodeAgent,
					Config: ir.AgentConfig{
						Prompt: "go.",
					},
					IO: ir.NodeIO{
						Reads:  []string{"key1", "key2", "key3"},
						Writes: []string{"out1", "out2"},
					},
				},
			},
		}
		output := Format(w)
		assertContains(t, output, "    reads: key1, key2, key3")
		assertContains(t, output, "    writes: out1, out2")
	})
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
		{time.Hour + 30*time.Minute + 15*time.Second, "1h30m15s"},
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

func TestFormatConditions(t *testing.T) {
	tests := []struct {
		name string
		expr ir.ConditionExpr
		want string
	}{
		{
			name: "simple compare",
			expr: ir.CondCompare{Variable: "ctx.outcome", Op: "=", Value: "success"},
			want: "ctx.outcome = success",
		},
		{
			name: "AND condition",
			expr: ir.CondAnd{
				Left:  ir.CondCompare{Variable: "ctx.x", Op: "=", Value: "1"},
				Right: ir.CondCompare{Variable: "ctx.y", Op: "=", Value: "2"},
			},
			want: "ctx.x = 1 and ctx.y = 2",
		},
		{
			name: "OR condition",
			expr: ir.CondOr{
				Left:  ir.CondCompare{Variable: "ctx.x", Op: "=", Value: "1"},
				Right: ir.CondCompare{Variable: "ctx.x", Op: "=", Value: "2"},
			},
			want: "ctx.x = 1 or ctx.x = 2",
		},
		{
			name: "NOT condition",
			expr: ir.CondNot{
				Inner: ir.CondCompare{Variable: "ctx.done", Op: "=", Value: "true"},
			},
			want: "not ctx.done = true",
		},
		{
			name: "nested AND/OR — AND inside OR needs parens",
			expr: ir.CondOr{
				Left: ir.CondAnd{
					Left:  ir.CondCompare{Variable: "ctx.x", Op: "=", Value: "1"},
					Right: ir.CondCompare{Variable: "ctx.y", Op: "=", Value: "2"},
				},
				Right: ir.CondCompare{Variable: "ctx.z", Op: "=", Value: "3"},
			},
			want: "(ctx.x = 1 and ctx.y = 2) or ctx.z = 3",
		},
		{
			name: "NOT of compound — parens around inner AND",
			expr: ir.CondNot{
				Inner: ir.CondAnd{
					Left:  ir.CondCompare{Variable: "ctx.x", Op: "=", Value: "1"},
					Right: ir.CondCompare{Variable: "ctx.y", Op: "=", Value: "2"},
				},
			},
			want: "not (ctx.x = 1 and ctx.y = 2)",
		},
		{
			name: "NOT of OR — parens around inner OR",
			expr: ir.CondNot{
				Inner: ir.CondOr{
					Left:  ir.CondCompare{Variable: "ctx.a", Op: "=", Value: "1"},
					Right: ir.CondCompare{Variable: "ctx.b", Op: "=", Value: "2"},
				},
			},
			want: "not (ctx.a = 1 or ctx.b = 2)",
		},
		{
			name: "OR inside AND needs parens",
			expr: ir.CondAnd{
				Left: ir.CondOr{
					Left:  ir.CondCompare{Variable: "ctx.a", Op: "=", Value: "1"},
					Right: ir.CondCompare{Variable: "ctx.b", Op: "=", Value: "2"},
				},
				Right: ir.CondCompare{Variable: "ctx.c", Op: "=", Value: "3"},
			},
			want: "(ctx.a = 1 or ctx.b = 2) and ctx.c = 3",
		},
		{
			name: "nil condition",
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

func TestFormatTrailingNewline(t *testing.T) {
	w := minimalWorkflow()
	output := Format(w)
	if !strings.HasSuffix(output, "\n") {
		t.Error("output must end with a newline")
	}
	// Must be exactly one trailing newline
	trimmed := strings.TrimRight(output, "\n")
	if output != trimmed+"\n" {
		t.Errorf("output should have exactly one trailing newline, got %d",
			len(output)-len(trimmed))
	}
}

func TestFormatNoTrailingWhitespace(t *testing.T) {
	w := askAndExecuteWorkflow()
	output := Format(w)
	for i, line := range strings.Split(output, "\n") {
		if line != strings.TrimRight(line, " \t") {
			t.Errorf("line %d has trailing whitespace: %q", i+1, line)
		}
	}
}

func TestFormatBlankLineSeparation(t *testing.T) {
	w := askAndExecuteWorkflow()
	output := Format(w)

	// No triple-blank-lines
	if strings.Contains(output, "\n\n\n") {
		t.Error("output contains triple blank lines")
	}

	// Blank line between workflow header and defaults
	assertContains(t, output, "  exit: Done\n\n  defaults")
	// Blank line between defaults and first node
	assertContains(t, output, "    max_restarts: 5\n\n  human AskUser")
	// Blank line between nodes
	assertContains(t, output, "    mode: freeform\n\n  agent Interpret")
	// Blank line before edges
	assertContains(t, output, "      Ship it.\n\n  edges")
}

func TestFormatQuoting(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"simple", "simple"},
		{"two words", `"two words"`},
		{"path/to/file.dip", "path/to/file.dip"},
		{"has:colon", "has:colon"},
		{"has spaces and stuff!", `"has spaces and stuff!"`},
		{"", `""`},
		{"under_score", "under_score"},
		{"dash-case", "dash-case"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := quoteValue(tt.input)
			if got != tt.want {
				t.Errorf("quoteValue(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestFormatDefaultsCacheToolsAndCompaction(t *testing.T) {
	w := &ir.Workflow{
		Name:  "test",
		Start: "A",
		Exit:  "A",
		Defaults: ir.WorkflowDefaults{
			CacheTools: true,
			Compaction: "rolling",
		},
		Nodes: []*ir.Node{
			{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "go."}},
		},
	}
	output := Format(w)
	assertContains(t, output, "    cache_tools: true")
	assertContains(t, output, "    compaction: rolling")
}

func TestFormatDefaultsRestartTarget(t *testing.T) {
	w := &ir.Workflow{
		Name:  "test",
		Start: "A",
		Exit:  "A",
		Defaults: ir.WorkflowDefaults{
			RestartTarget: "Begin",
		},
		Nodes: []*ir.Node{
			{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "go."}},
		},
	}
	output := Format(w)
	assertContains(t, output, "    restart_target: Begin")
}

func TestFormatSubgraphNoParams(t *testing.T) {
	w := &ir.Workflow{
		Name:  "sub",
		Start: "S",
		Exit:  "S",
		Nodes: []*ir.Node{
			{
				ID:   "S",
				Kind: ir.NodeSubgraph,
				Config: ir.SubgraphConfig{
					Ref: "./other.dip",
				},
			},
		},
	}
	output := Format(w)
	assertContains(t, output, "    ref: ./other.dip")
	assertNotContains(t, output, "params:")
}

func TestFormatEdgeWeightOnly(t *testing.T) {
	w := &ir.Workflow{
		Name:  "test",
		Start: "A",
		Exit:  "B",
		Nodes: []*ir.Node{
			{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "go."}},
			{ID: "B", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "done."}},
		},
		Edges: []*ir.Edge{
			{From: "A", To: "B", Weight: 5},
		},
	}
	output := Format(w)
	assertContains(t, output, "    A -> B  weight: 5")
}

func TestFormatNilWorkflowConfig(t *testing.T) {
	// A node with nil config should not panic
	w := &ir.Workflow{
		Name:  "test",
		Start: "A",
		Exit:  "A",
		Nodes: []*ir.Node{
			{ID: "A", Kind: ir.NodeAgent}, // Config is nil
		},
	}
	output := Format(w)
	assertContains(t, output, "  agent A")
}

func TestFormatLabelOnSubgraph(t *testing.T) {
	w := &ir.Workflow{
		Name:  "sub",
		Start: "S",
		Exit:  "S",
		Nodes: []*ir.Node{
			{
				ID:    "S",
				Kind:  ir.NodeSubgraph,
				Label: "Code Review",
				Config: ir.SubgraphConfig{
					Ref: "./review.dip",
				},
			},
		},
	}
	output := Format(w)
	assertContains(t, output, `    label: "Code Review"`)
	assertContains(t, output, "    ref: ./review.dip")
}

func TestFormatBaseDelay(t *testing.T) {
	w := &ir.Workflow{
		Name:  "test",
		Start: "A",
		Exit:  "A",
		Nodes: []*ir.Node{
			{
				ID:     "A",
				Kind:   ir.NodeAgent,
				Config: ir.AgentConfig{Prompt: "go."},
				Retry: ir.RetryConfig{
					Policy:    "aggressive",
					BaseDelay: 500 * time.Millisecond,
				},
			},
		},
	}
	output := Format(w)
	assertContains(t, output, "retry_policy: aggressive")
	assertContains(t, output, "base_delay: 500ms")
}

func TestFormatToolOutputs(t *testing.T) {
	w := &ir.Workflow{
		Name:  "tool_outputs",
		Start: "T",
		Exit:  "T",
		Nodes: []*ir.Node{
			{
				ID:   "T",
				Kind: ir.NodeTool,
				Config: ir.ToolConfig{
					Outputs: []string{"complete", "continue", "error"},
					Command: "echo done",
					Timeout: 30 * time.Second,
				},
			},
		},
	}
	output := Format(w)
	assertContains(t, output, "outputs: complete, continue, error")
	// outputs should appear before timeout and command
	outputsIdx := strings.Index(output, "outputs:")
	timeoutIdx := strings.Index(output, "timeout:")
	commandIdx := strings.Index(output, "command:")
	if outputsIdx > timeoutIdx {
		t.Error("outputs should appear before timeout")
	}
	if outputsIdx > commandIdx {
		t.Error("outputs should appear before command")
	}
	assertIdempotent(t, w)
}

func TestFormatToolWithoutOutputs(t *testing.T) {
	w := &ir.Workflow{
		Name:  "tool_no_outputs",
		Start: "T",
		Exit:  "T",
		Nodes: []*ir.Node{
			{
				ID:   "T",
				Kind: ir.NodeTool,
				Config: ir.ToolConfig{
					Command: "echo done",
				},
			},
		},
	}
	output := Format(w)
	assertNotContains(t, output, "outputs:")
}

func TestFormatBaseDelay_Zero_Omitted(t *testing.T) {
	w := &ir.Workflow{
		Name:  "test",
		Start: "A",
		Exit:  "A",
		Nodes: []*ir.Node{
			{
				ID:     "A",
				Kind:   ir.NodeAgent,
				Config: ir.AgentConfig{Prompt: "go."},
				Retry:  ir.RetryConfig{Policy: "standard"},
			},
		},
	}
	output := Format(w)
	assertNotContains(t, output, "base_delay")
}

// --- Phase 2: New coverage tests ---

func TestFormatParallelBlock(t *testing.T) {
	tests := []struct {
		name   string
		node   *ir.Node
		checks []string
	}{
		{
			name: "branches with all fields",
			node: &ir.Node{
				ID:   "Split",
				Kind: ir.NodeParallel,
				Config: ir.ParallelConfig{
					Branches: []ir.BranchConfig{
						{Target: "A", Model: "gpt-5.4", Provider: "openai", Fidelity: "full"},
						{Target: "B", Model: "claude-opus-4-6", Provider: "anthropic", Fidelity: "summary:high"},
					},
				},
			},
			checks: []string{
				"  parallel Split",
				"    branch: A",
				"      model: gpt-5.4",
				"      provider: openai",
				"      fidelity: full",
				"    branch: B",
				"      model: claude-opus-4-6",
				"      provider: anthropic",
				"      fidelity: summary:high",
			},
		},
		{
			name: "branches with model only",
			node: &ir.Node{
				ID:   "Fan",
				Kind: ir.NodeParallel,
				Config: ir.ParallelConfig{
					Branches: []ir.BranchConfig{
						{Target: "worker_a", Model: "o1"},
						{Target: "worker_b", Model: "claude-3"},
					},
				},
			},
			checks: []string{
				"  parallel Fan",
				"    branch: worker_a",
				"      model: o1",
				"    branch: worker_b",
				"      model: claude-3",
			},
		},
		{
			name: "branch with no config fields",
			node: &ir.Node{
				ID:   "P",
				Kind: ir.NodeParallel,
				Config: ir.ParallelConfig{
					Branches: []ir.BranchConfig{
						{Target: "X"},
						{Target: "Y"},
					},
				},
			},
			checks: []string{
				"  parallel P",
				"    branch: X",
				"    branch: Y",
			},
		},
		{
			name: "single branch",
			node: &ir.Node{
				ID:   "S",
				Kind: ir.NodeParallel,
				Config: ir.ParallelConfig{
					Branches: []ir.BranchConfig{
						{Target: "Only", Model: "gpt-5.4"},
					},
				},
			},
			checks: []string{
				"  parallel S",
				"    branch: Only",
				"      model: gpt-5.4",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &ir.Workflow{
				Name:  "par_test",
				Start: tt.node.ID,
				Exit:  tt.node.ID,
				Nodes: []*ir.Node{tt.node},
			}
			output := Format(w)
			for _, check := range tt.checks {
				assertContains(t, output, check)
			}
			assertIdempotent(t, w)
		})
	}
}

func TestFormatStylesheet(t *testing.T) {
	tests := []struct {
		name   string
		rules  []ir.StylesheetRule
		checks []string
	}{
		{
			name: "all selector types sorted by specificity",
			rules: []ir.StylesheetRule{
				{Selector: ir.StyleSelector{Kind: "id", Value: "Done"}, Properties: map[string]string{"max_retries": "5"}},
				{Selector: ir.StyleSelector{Kind: "universal"}, Properties: map[string]string{"temperature": "0.7"}},
				{Selector: ir.StyleSelector{Kind: "class", Value: "coder"}, Properties: map[string]string{"model": "o1"}},
				{Selector: ir.StyleSelector{Kind: "kind", Value: "agent"}, Properties: map[string]string{"fidelity": "full"}},
			},
			checks: []string{
				"  stylesheet:",
				"    *",
				"      temperature: 0.7",
				"    agent",
				"      fidelity: full",
				"    .coder",
				"      model: o1",
				"    #Done",
				"      max_retries: 5",
			},
		},
		{
			name: "same-specificity alphabetical sort",
			rules: []ir.StylesheetRule{
				{Selector: ir.StyleSelector{Kind: "class", Value: "reviewer"}, Properties: map[string]string{"model": "gpt-5.4"}},
				{Selector: ir.StyleSelector{Kind: "class", Value: "coder"}, Properties: map[string]string{"model": "o1"}},
			},
			checks: []string{
				"    .coder",
				"    .reviewer",
			},
		},
		{
			name: "single universal rule",
			rules: []ir.StylesheetRule{
				{Selector: ir.StyleSelector{Kind: "universal"}, Properties: map[string]string{"temperature": "0.5"}},
			},
			checks: []string{
				"  stylesheet:",
				"    *",
				"      temperature: 0.5",
			},
		},
		{
			name: "multiple properties sorted alphabetically",
			rules: []ir.StylesheetRule{
				{Selector: ir.StyleSelector{Kind: "universal"}, Properties: map[string]string{
					"temperature":      "0.7",
					"model":            "gpt-5.4",
					"reasoning_effort": "medium",
				}},
			},
			checks: []string{
				"      model: gpt-5.4",
				"      reasoning_effort: medium",
				"      temperature: 0.7",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &ir.Workflow{
				Name:       "style_test",
				Start:      "A",
				Exit:       "A",
				Stylesheet: tt.rules,
				Nodes: []*ir.Node{
					{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "go."}},
				},
			}
			output := Format(w)
			for _, check := range tt.checks {
				assertContains(t, output, check)
			}
			assertIdempotent(t, w)
		})
	}
}

func TestFormatFloat(t *testing.T) {
	tests := []struct {
		input float64
		want  string
	}{
		{0.8, "0.8"},
		{1.0, "1.0"},
		{0.75, "0.75"},
		{0.0, "0.0"},
		{100.0, "100.0"},
	}
	for _, tt := range tests {
		got := formatFloat(tt.input)
		if got != tt.want {
			t.Errorf("formatFloat(%v) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestFormatAgentCompactionThreshold(t *testing.T) {
	w := &ir.Workflow{
		Name:  "compaction_test",
		Start: "A",
		Exit:  "A",
		Nodes: []*ir.Node{
			{
				ID:   "A",
				Kind: ir.NodeAgent,
				Config: ir.AgentConfig{
					Prompt:              "Go.",
					CompactionThreshold: 0.8,
				},
			},
		},
	}
	output := Format(w)
	assertContains(t, output, "compaction_threshold: 0.8")
	assertIdempotent(t, w)
}

func TestFormatEdgeConditionRawOnly(t *testing.T) {
	w := &ir.Workflow{
		Name:  "raw_cond",
		Start: "A",
		Exit:  "B",
		Nodes: []*ir.Node{
			{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "go."}},
			{ID: "B", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "done."}},
		},
		Edges: []*ir.Edge{
			{From: "A", To: "B", Condition: &ir.Condition{
				Raw:    "ctx.outcome = success",
				Parsed: nil, // Parsed not populated
			}},
		},
	}
	output := Format(w)
	assertContains(t, output, "A -> B  when ctx.outcome = success")
	assertIdempotent(t, w)
}

func TestFormatHumanInterview(t *testing.T) {
	data, err := os.ReadFile("../parser/testdata/human_interview.dip")
	if err != nil {
		t.Fatalf("read human_interview.dip: %v", err)
	}
	w, err := parser.NewParser(string(data), "human_interview.dip").Parse()
	if err != nil {
		t.Fatalf("first parse: %v", err)
	}
	got := Format(w)
	w2, err := parser.NewParser(got, "formatted.dip").Parse()
	if err != nil {
		t.Fatalf("second parse: %v\nformatted:\n%s", err, got)
	}
	gate := w2.Nodes[0]
	for _, n := range w2.Nodes {
		if n.ID == "Gate" {
			gate = n
		}
	}
	cfg := gate.Config.(ir.HumanConfig)
	if cfg.QuestionsKey != "interview_questions" {
		t.Errorf("round-trip QuestionsKey = %q", cfg.QuestionsKey)
	}
	if cfg.AnswersKey != "interview_answers" {
		t.Errorf("round-trip AnswersKey = %q", cfg.AnswersKey)
	}
}

func TestFormatRoundTrip(t *testing.T) {
	// Parse all_features.dip, format, parse again, format again, assert identical.
	data, err := os.ReadFile("../parser/testdata/all_features.dip")
	if err != nil {
		t.Fatalf("read all_features.dip: %v", err)
	}

	p1 := parser.NewParser(string(data), "all_features.dip")
	w1, err := p1.Parse()
	if err != nil {
		t.Fatalf("first parse: %v", err)
	}

	formatted1 := Format(w1)

	p2 := parser.NewParser(formatted1, "formatted.dip")
	w2, err := p2.Parse()
	if err != nil {
		t.Fatalf("second parse: %v", err)
	}

	formatted2 := Format(w2)

	if formatted1 != formatted2 {
		t.Errorf("round-trip not stable:\n--- first ---\n%s\n--- second ---\n%s", formatted1, formatted2)
	}
}

func TestFormatResponseFormat(t *testing.T) {
	w := &ir.Workflow{
		Name:  "resp_fmt_test",
		Start: "A",
		Exit:  "A",
		Nodes: []*ir.Node{
			{
				ID:   "A",
				Kind: ir.NodeAgent,
				Config: ir.AgentConfig{
					Provider:       "openai",
					ResponseFormat: "json_object",
					Prompt:         "Return JSON.",
				},
			},
		},
	}

	output := Format(w)
	assertContains(t, output, "response_format: json_object")
	assertContains(t, output, "provider: openai")
	assertContains(t, output, "prompt:")

	// response_format appears after provider, before prompt
	providerIdx := strings.Index(output, "provider:")
	responseFormatIdx := strings.Index(output, "response_format:")
	promptIdx := strings.Index(output, "prompt:")
	if responseFormatIdx < providerIdx {
		t.Error("response_format should appear after provider")
	}
	if responseFormatIdx > promptIdx {
		t.Error("response_format should appear before prompt")
	}
	assertIdempotent(t, w)
}

func TestFormatResponseSchema(t *testing.T) {
	schema := `{
  "type": "object",
  "properties": {
    "result": {"type": "string"}
  }
}`
	w := &ir.Workflow{
		Name:  "resp_schema_test",
		Start: "A",
		Exit:  "A",
		Nodes: []*ir.Node{
			{
				ID:   "A",
				Kind: ir.NodeAgent,
				Config: ir.AgentConfig{
					ResponseFormat: "json_schema",
					ResponseSchema: schema,
					Prompt:         "Return structured output.",
				},
			},
		},
	}

	output := Format(w)
	assertContains(t, output, "response_format: json_schema")
	assertContains(t, output, "response_schema:")
	assertContains(t, output, `"type": "object"`)
	assertContains(t, output, `"result": {"type": "string"}`)
	assertIdempotent(t, w)
}

func TestFormatAgentParams(t *testing.T) {
	w := &ir.Workflow{
		Name:  "agent_params_test",
		Start: "A",
		Exit:  "A",
		Nodes: []*ir.Node{
			{
				ID:   "A",
				Kind: ir.NodeAgent,
				Config: ir.AgentConfig{
					Params: map[string]string{
						"backend":         "claude-code",
						"permission_mode": "auto",
					},
					Prompt: "Do work.",
				},
			},
		},
	}

	output := Format(w)
	assertContains(t, output, "params:")
	assertContains(t, output, "backend: claude-code")
	assertContains(t, output, "permission_mode: auto")

	// params appear before prompt
	paramsIdx := strings.Index(output, "params:")
	promptIdx := strings.Index(output, "prompt:")
	if paramsIdx > promptIdx {
		t.Error("params should appear before prompt")
	}
	assertIdempotent(t, w)
}

func TestFormatAgentNoResponseFieldsWhenUnset(t *testing.T) {
	w := &ir.Workflow{
		Name:  "no_response_fields",
		Start: "A",
		Exit:  "A",
		Nodes: []*ir.Node{
			{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{
				Prompt: "Just a plain prompt.",
			}},
		},
	}
	output := Format(w)
	assertNotContains(t, output, "response_format:")
	assertNotContains(t, output, "response_schema:")
	assertNotContains(t, output, "params:")
}

func TestFormatAllStructuredOutputFieldsTogether(t *testing.T) {
	schema := "{\"type\":\"object\",\"properties\":{\"answer\":{\"type\":\"string\"}},\"required\":[\"answer\"]}"
	w := &ir.Workflow{
		Name:  "all_structured",
		Start: "A",
		Exit:  "A",
		Nodes: []*ir.Node{
			{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{
				ResponseFormat: "json_schema",
				ResponseSchema: schema,
				Params:         map[string]string{"backend": "claude-code"},
				Prompt:         "Return structured output.",
			}},
		},
	}
	output := Format(w)
	assertContains(t, output, "response_format: json_schema")
	assertContains(t, output, "response_schema:")
	assertContains(t, output, "params:")
	assertContains(t, output, "backend: claude-code")

	// Field ordering
	rfIdx := strings.Index(output, "response_format:")
	rsIdx := strings.Index(output, "response_schema:")
	pIdx := strings.Index(output, "params:")
	prIdx := strings.Index(output, "prompt:")
	if rfIdx > rsIdx {
		t.Error("response_format should appear before response_schema")
	}
	if pIdx > prIdx {
		t.Error("params should appear before prompt")
	}
	assertIdempotent(t, w)
}

func TestFormatConditionalNode(t *testing.T) {
	w := &ir.Workflow{
		Name:  "cond_test",
		Start: "Start",
		Exit:  "Done",
		Nodes: []*ir.Node{
			{ID: "Start", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "Go."}},
			{ID: "Check", Kind: ir.NodeConditional, Label: "Evaluate", Config: ir.ConditionalConfig{}},
			{ID: "Done", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "Done."}},
		},
		Edges: []*ir.Edge{
			{From: "Start", To: "Check"},
			{From: "Check", To: "Done"},
		},
	}
	output := Format(w)
	if !strings.Contains(output, "conditional Check") {
		t.Errorf("expected 'conditional Check' in output, got:\n%s", output)
	}
	if !strings.Contains(output, "label: Evaluate") {
		t.Errorf("expected 'label: Evaluate' in output, got:\n%s", output)
	}
}

func TestFormatVarsBlock(t *testing.T) {
	w := &ir.Workflow{
		Name:  "Test",
		Start: "A",
		Exit:  "B",
		Vars: map[string]string{
			"env":     "production",
			"api_url": "https://example.com/api",
			"retries": "3",
		},
		Nodes: []*ir.Node{
			{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "do it"}},
			{ID: "B", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "done"}},
		},
		Edges: []*ir.Edge{{From: "A", To: "B"}},
	}
	out := Format(w)

	// vars block must appear
	if !strings.Contains(out, "vars\n") {
		t.Errorf("expected vars block header, got:\n%s", out)
	}
	// keys must be sorted
	apiIdx := strings.Index(out, "api_url:")
	envIdx := strings.Index(out, "env:")
	retriesIdx := strings.Index(out, "retries:")
	if apiIdx < 0 || envIdx < 0 || retriesIdx < 0 {
		t.Fatalf("missing key in output:\n%s", out)
	}
	if !(apiIdx < envIdx && envIdx < retriesIdx) {
		t.Errorf("vars keys not sorted: api_url@%d env@%d retries@%d\n%s", apiIdx, envIdx, retriesIdx, out)
	}
	// values present
	if !strings.Contains(out, "api_url: https://example.com/api") {
		t.Errorf("api_url value missing:\n%s", out)
	}
	if !strings.Contains(out, "env: production") {
		t.Errorf("env value missing:\n%s", out)
	}
}

func TestFormatVarsEmptyOmitted(t *testing.T) {
	w := &ir.Workflow{
		Name:  "Test",
		Start: "A",
		Exit:  "B",
		Vars:  nil,
		Nodes: []*ir.Node{
			{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "do it"}},
			{ID: "B", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "done"}},
		},
		Edges: []*ir.Edge{{From: "A", To: "B"}},
	}
	out := Format(w)
	if strings.Contains(out, "vars") {
		t.Errorf("expected no vars block for nil vars, got:\n%s", out)
	}

	// Also test empty map
	w.Vars = map[string]string{}
	out = Format(w)
	if strings.Contains(out, "vars") {
		t.Errorf("expected no vars block for empty vars map, got:\n%s", out)
	}
}

func TestFormatManagerLoop_RoundTrip(t *testing.T) {
	src := `workflow W
  start: M
  exit: M

  manager_loop M
    label: "Quality Gate"
    subgraph_ref: quality_loop
    poll_interval: 30s
    max_cycles: 12
    stop_condition: stack.child.cycles = 10
    steer_condition: stack.child.cycles = 5
    steer_context: hint=speed_up, priority=high, urgency=low

  edges
    M -> M
`
	p := parser.NewParser(src, "test.dip")
	w, err := p.Parse()
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	out := Format(w)

	// Round-trip: reparse the formatted output.
	p2 := parser.NewParser(out, "formatted.dip")
	w2, err := p2.Parse()
	if err != nil {
		t.Fatalf("reparse of formatted output: err=%v\n%s", err, out)
	}
	mNode := w2.Node("M")
	if mNode == nil {
		t.Fatalf("node M not found after round-trip")
	}
	cfg, ok := mNode.Config.(ir.ManagerLoopConfig)
	if !ok {
		t.Fatalf("Config = %T, want ManagerLoopConfig", mNode.Config)
	}
	if cfg.SubgraphRef != "quality_loop" {
		t.Errorf("SubgraphRef = %q", cfg.SubgraphRef)
	}
	if cfg.MaxCycles != 12 {
		t.Errorf("MaxCycles = %d", cfg.MaxCycles)
	}
	if cfg.PollInterval != 30*time.Second {
		t.Errorf("PollInterval = %v", cfg.PollInterval)
	}
	if cfg.StopCondition == nil || cfg.StopCondition.Raw != "stack.child.cycles = 10" {
		t.Errorf("StopCondition = %+v", cfg.StopCondition)
	}
	if cfg.SteerCondition == nil || cfg.SteerCondition.Raw != "stack.child.cycles = 5" {
		t.Errorf("SteerCondition = %+v", cfg.SteerCondition)
	}
	if cfg.SteerContext["hint"] != "speed_up" || cfg.SteerContext["priority"] != "high" || cfg.SteerContext["urgency"] != "low" {
		t.Errorf("SteerContext round-trip failed: %v", cfg.SteerContext)
	}
	if mNode.Label != "Quality Gate" {
		t.Errorf("Label = %q, want %q", mNode.Label, "Quality Gate")
	}

	// Idempotency: formatting the formatted output produces the same string.
	if Format(w2) != out {
		t.Errorf("format is not idempotent\n---first---\n%s\n---second---\n%s", out, Format(w2))
	}
}

// TestFormatManagerLoop_ParsedConditionFallback verifies that a ManagerLoopConfig
// with only Parsed (no Raw) on StopCondition still emits a stop_condition line.
// This covers programmatically-built IR where Raw is never populated.
func TestFormatManagerLoop_ParsedConditionFallback(t *testing.T) {
	w := &ir.Workflow{
		Name:  "W",
		Start: "M",
		Exit:  "M",
		Nodes: []*ir.Node{
			{
				ID:   "M",
				Kind: ir.NodeManagerLoop,
				Config: ir.ManagerLoopConfig{
					SubgraphRef: "inner.dip",
					MaxCycles:   5,
					// Parsed set, Raw intentionally empty — simulates programmatic IR construction.
					StopCondition: &ir.Condition{
						Parsed: ir.CondCompare{Variable: "stack.child.outcome", Op: "=", Value: "success"},
					},
				},
			},
		},
		Edges: []*ir.Edge{{From: "M", To: "M"}},
	}
	out := Format(w)
	if !strings.Contains(out, "stop_condition:") {
		t.Errorf("expected stop_condition in formatted output, got:\n%s", out)
	}
	if !strings.Contains(out, "stack.child.outcome = success") {
		t.Errorf("expected condition text in output, got:\n%s", out)
	}
}

func TestFormatRequiresSingle(t *testing.T) {
	w := &ir.Workflow{
		Name:     "req_single",
		Start:    "A",
		Exit:     "A",
		Requires: []string{"git"},
		Nodes: []*ir.Node{
			{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "Do."}},
		},
		Edges: []*ir.Edge{{From: "A", To: "A"}},
	}
	out := Format(w)
	assertContains(t, out, "  requires: git")
	assertIdempotent(t, w)
}

func TestFormatRequiresMulti(t *testing.T) {
	w := &ir.Workflow{
		Name:     "req_multi",
		Goal:     "Multi-dep workflow",
		Start:    "A",
		Exit:     "A",
		Requires: []string{"git", "docker", "jq"},
		Nodes: []*ir.Node{
			{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "Do."}},
		},
		Edges: []*ir.Edge{{From: "A", To: "A"}},
	}
	out := Format(w)
	assertContains(t, out, "  requires: git, docker, jq")

	// requires must sit between goal and start in canonical order.
	lines := strings.Split(out, "\n")
	idxGoal, idxReq, idxStart := -1, -1, -1
	for i, l := range lines {
		trim := strings.TrimSpace(l)
		switch {
		case strings.HasPrefix(trim, "goal:"):
			idxGoal = i
		case strings.HasPrefix(trim, "requires:"):
			idxReq = i
		case strings.HasPrefix(trim, "start:"):
			idxStart = i
		}
	}
	if !(idxGoal < idxReq && idxReq < idxStart) {
		t.Errorf("expected goal < requires < start; got %d, %d, %d", idxGoal, idxReq, idxStart)
	}
	assertIdempotent(t, w)
}

func TestFormatRequiresOmittedWhenNilOrEmpty(t *testing.T) {
	w := &ir.Workflow{
		Name:  "no_req",
		Start: "A",
		Exit:  "A",
		Nodes: []*ir.Node{
			{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "Do."}},
		},
		Edges: []*ir.Edge{{From: "A", To: "A"}},
	}
	assertNotContains(t, Format(w), "requires:")
	w.Requires = []string{}
	assertNotContains(t, Format(w), "requires:")
}

func TestFormatRequiresRoundTrip(t *testing.T) {
	w1 := &ir.Workflow{
		Name:     "req_rt",
		Start:    "A",
		Exit:     "A",
		Requires: []string{"git", "docker"},
		Nodes: []*ir.Node{
			{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "Do."}},
		},
		Edges: []*ir.Edge{{From: "A", To: "A"}},
	}
	formatted := Format(w1)
	w2, err := parser.NewParser(formatted, "rt").Parse()
	if err != nil {
		t.Fatalf("re-parse: %v\nformatted:\n%s", err, formatted)
	}
	if len(w2.Requires) != len(w1.Requires) {
		t.Fatalf("Requires len: %d vs %d", len(w1.Requires), len(w2.Requires))
	}
	for i := range w1.Requires {
		if w1.Requires[i] != w2.Requires[i] {
			t.Errorf("Requires[%d]: %q vs %q", i, w1.Requires[i], w2.Requires[i])
		}
	}
}

func TestFormat_AgentToolAccess(t *testing.T) {
	src := `workflow X
  start: A
  exit: A

  agent A
    prompt: "x"
    tool_access: none
`
	p := parser.NewParser(src, "test.dip")
	w, err := p.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	out := Format(w)
	if !strings.Contains(out, "tool_access: none") {
		t.Errorf("formatted output missing tool_access line:\n%s", out)
	}
}

func TestFormat_AgentToolAccessOmitted(t *testing.T) {
	src := `workflow X
  start: A
  exit: A

  agent A
    prompt: "x"
`
	p := parser.NewParser(src, "test.dip")
	w, err := p.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	out := Format(w)
	if strings.Contains(out, "tool_access") {
		t.Errorf("omitted tool_access should not appear in formatted output:\n%s", out)
	}
}

// TestFormat_AgentToolAccessWhitespaceOnly verifies that a whitespace-only
// tool_access value is treated as omitted on round-trip. The validator
// canonicalizes via TrimSpace + ToLower so whitespace passes lint as "empty",
// but pre-fix the formatter preserved the literal whitespace, producing a
// non-empty tool_access line that contradicted lint semantics. The fix gates
// emission on TrimSpace-non-empty.
func TestFormat_AgentToolAccessWhitespaceOnly(t *testing.T) {
	src := `workflow X
  start: A
  exit: A

  agent A
    prompt: "x"
    tool_access: "   "
`
	p := parser.NewParser(src, "test.dip")
	w, err := p.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	out := Format(w)
	if strings.Contains(out, "tool_access") {
		t.Errorf("whitespace-only tool_access should not appear in formatted output:\n%s", out)
	}
}

func TestFormat_ToolCommandFile(t *testing.T) {
	src := `workflow X
  start: A
  exit: A

  tool A
    command_file: scripts/setup.sh
`
	p := parser.NewParser(src, "test.dip")
	w, err := p.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	out := Format(w)
	if !strings.Contains(out, "command_file: scripts/setup.sh") {
		t.Errorf("formatted output missing command_file directive:\n%s", out)
	}
	// And it should NOT also emit a multi-line command: block
	if strings.Contains(out, "command:") {
		t.Errorf("formatted output should not contain inline command: when file form used:\n%s", out)
	}
}

func TestFormat_AgentInlineSystemPromptRoundTrip(t *testing.T) {
	// Regression: pre-#65, writeAgentFields silently dropped SystemPrompt
	// on format. This test pins the fix.
	wf := &ir.Workflow{
		Name:  "W",
		Goal:  "test",
		Start: "A",
		Exit:  "A",
		Nodes: []*ir.Node{
			{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{
				Model:        "claude-sonnet-4-6",
				SystemPrompt: "You are a code reviewer.",
			}},
		},
	}
	out := Format(wf)
	if !strings.Contains(out, "system_prompt:") {
		t.Errorf("formatter dropped inline system_prompt; output:\n%s", out)
	}
	if !strings.Contains(out, "You are a code reviewer.") {
		t.Errorf("formatter dropped system_prompt content; output:\n%s", out)
	}
}

func TestFormat_AgentPromptFile(t *testing.T) {
	wf := &ir.Workflow{
		Name:  "W",
		Goal:  "test",
		Start: "A",
		Exit:  "A",
		Nodes: []*ir.Node{
			{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{
				Model:      "claude-sonnet-4-6",
				PromptFile: "prompts/task.md",
			}},
		},
	}
	out := Format(wf)
	if !strings.Contains(out, "prompt_file: prompts/task.md") {
		t.Errorf("expected `prompt_file: prompts/task.md` in output; got:\n%s", out)
	}
	if strings.Contains(out, "\n    prompt:") {
		t.Errorf("formatter emitted inline prompt: when PromptFile was set; got:\n%s", out)
	}
}

func TestFormat_AgentSystemPromptFile(t *testing.T) {
	wf := &ir.Workflow{
		Name:  "W",
		Goal:  "test",
		Start: "A",
		Exit:  "A",
		Nodes: []*ir.Node{
			{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{
				Model:            "claude-sonnet-4-6",
				SystemPromptFile: "prompts/persona.md",
			}},
		},
	}
	out := Format(wf)
	if !strings.Contains(out, "system_prompt_file: prompts/persona.md") {
		t.Errorf("expected `system_prompt_file: prompts/persona.md` in output; got:\n%s", out)
	}
	if strings.Contains(out, "\n    system_prompt:") {
		t.Errorf("formatter emitted inline system_prompt: when SystemPromptFile was set; got:\n%s", out)
	}
}

func TestFormat_AgentCrossSlotMix(t *testing.T) {
	wf := &ir.Workflow{
		Name:  "W",
		Goal:  "test",
		Start: "A",
		Exit:  "A",
		Nodes: []*ir.Node{
			{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{
				Model:        "claude-sonnet-4-6",
				PromptFile:   "prompts/task.md",
				SystemPrompt: "You are a reviewer.",
			}},
		},
	}
	out := Format(wf)
	if !strings.Contains(out, "prompt_file: prompts/task.md") {
		t.Errorf("missing prompt_file: line; got:\n%s", out)
	}
	if !strings.Contains(out, "system_prompt:") {
		t.Errorf("missing inline system_prompt:; got:\n%s", out)
	}
	if !strings.Contains(out, "You are a reviewer.") {
		t.Errorf("missing system_prompt content; got:\n%s", out)
	}
}

func TestFormat_ToolCommandInline(t *testing.T) {
	src := `workflow X
  start: A
  exit: A

  tool A
    command: echo hi
`
	p := parser.NewParser(src, "test.dip")
	w, err := p.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	out := Format(w)
	if !strings.Contains(out, "command:") {
		t.Errorf("formatted output missing inline command:\n%s", out)
	}
	if strings.Contains(out, "command_file:") {
		t.Errorf("formatted output should not contain command_file when inline used:\n%s", out)
	}
}

func TestFormatAgentWritablePaths(t *testing.T) {
	w := &ir.Workflow{
		Name: "T", Start: "A", Exit: "A",
		Nodes: []*ir.Node{
			{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{
				Prompt:        "x",
				WritablePaths: []string{"workspace/**", ".ai/sprints/**"},
			}},
		},
	}
	out := Format(w)
	if !strings.Contains(out, "writable_paths: workspace/**, .ai/sprints/**") {
		t.Errorf("formatted output missing writable_paths; got:\n%s", out)
	}
}

func TestFormatBranchWritablePathsOnly(t *testing.T) {
	w := &ir.Workflow{
		Name: "T", Start: "split", Exit: "join",
		Nodes: []*ir.Node{
			{ID: "a", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "a"}},
			{ID: "split", Kind: ir.NodeParallel, Config: ir.ParallelConfig{
				Targets:  []string{"a"},
				Branches: []ir.BranchConfig{{Target: "a", WritablePaths: []string{"workspace/**"}}},
			}},
			{ID: "join", Kind: ir.NodeFanIn, Config: ir.FanInConfig{Sources: []string{"a"}}},
		},
	}
	out := Format(w)
	if !strings.Contains(out, "writable_paths: workspace/**") {
		t.Errorf("formatted output missing per-branch writable_paths; got:\n%s", out)
	}
}

// A branch that sets ONLY tool_access must still be emitted (regression guard
// for the writeBranch early-return that previously checked only model/provider/fidelity).
func TestFormatBranchToolAccessOnly(t *testing.T) {
	w := &ir.Workflow{
		Name:  "T",
		Start: "split",
		Exit:  "join",
		Nodes: []*ir.Node{
			{ID: "a", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "a"}},
			{ID: "split", Kind: ir.NodeParallel, Config: ir.ParallelConfig{
				Targets:  []string{"a"},
				Branches: []ir.BranchConfig{{Target: "a", ToolAccess: "none"}},
			}},
			{ID: "join", Kind: ir.NodeFanIn, Config: ir.FanInConfig{Sources: []string{"a"}}},
		},
	}
	out := Format(w)
	if !strings.Contains(out, "tool_access: none") {
		t.Errorf("formatted output missing per-branch tool_access; got:\n%s", out)
	}
}

func TestFormatAgentWritablePathsRoundTrips(t *testing.T) {
	src := `workflow X
  start: A
  exit: A

  agent A
    prompt: "x"
    writable_paths: workspace/**, .ai/sprints/**
`
	w1, err := parser.NewParser(src, "rt.dip").Parse()
	if err != nil {
		t.Fatalf("parse1: %v", err)
	}
	w2, err := parser.NewParser(Format(w1), "rt.dip").Parse()
	if err != nil {
		t.Fatalf("parse2: %v", err)
	}
	got := w2.Node("A").Config.(ir.AgentConfig).WritablePaths
	if len(got) != 2 || got[0] != "workspace/**" || got[1] != ".ai/sprints/**" {
		t.Errorf("WritablePaths after format round-trip = %v, want [workspace/** .ai/sprints/**]", got)
	}
}
