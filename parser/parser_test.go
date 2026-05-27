package parser

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/2389-research/dippin-lang/formatter"
	"github.com/2389-research/dippin-lang/ir"
)

func TestParseSimpleWorkflow(t *testing.T) {
	input := `workflow Test
  goal: "Test"
  start: A
  exit: B

  agent A
    prompt: "Do A."

  agent B
    prompt: "Do B."

  edges
    A -> B
`
	p := NewParser(input, "test.dip")
	w, err := p.Parse()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if w.Name != "Test" {
		t.Errorf("name = %q, want %q", w.Name, "Test")
	}
	if len(w.Nodes) != 2 {
		t.Errorf("nodes = %d, want 2", len(w.Nodes))
	}
	if len(w.Edges) != 1 {
		t.Errorf("edges = %d, want 1", len(w.Edges))
	}
	if w.Start != "A" || w.Exit != "B" {
		t.Errorf("start/exit = %s/%s, want A/B", w.Start, w.Exit)
	}
}

func TestParseComplexWorkflow(t *testing.T) {
	input := `workflow Complex
  goal: "Complex"
  start: P
  exit: J

  parallel P -> A, B

  agent A
    prompt:
      Do A.
      More text.

  agent B
    prompt: Do B.

  fan_in J <- A, B

  edges
    P -> A
    P -> B
    A -> J
    B -> J
`
	p := NewParser(input, "test.dip")
	w, err := p.Parse()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(w.Nodes) != 4 {
		t.Errorf("nodes = %d, want 4", len(w.Nodes))
	}

	// Check parallel node
	var parallelNode *ir.Node
	for _, n := range w.Nodes {
		if n.Kind == ir.NodeParallel {
			parallelNode = n
			break
		}
	}
	if parallelNode == nil {
		t.Fatal("parallel node not found")
	}
	if len(parallelNode.Config.(ir.ParallelConfig).Targets) != 2 {
		t.Errorf("parallel targets = %d, want 2", len(parallelNode.Config.(ir.ParallelConfig).Targets))
	}

	// Check multiline prompt
	var agentANode *ir.Node
	for _, n := range w.Nodes {
		if n.ID == "A" {
			agentANode = n
			break
		}
	}
	if agentANode == nil {
		t.Fatal("agent A not found")
	}
	prompt := agentANode.Config.(ir.AgentConfig).Prompt
	if !strings.Contains(prompt, "Do A.") || !strings.Contains(prompt, "More text.") {
		t.Errorf("prompt = %q, want it to contain 'Do A.' and 'More text.'", prompt)
	}
}

func TestParseShellScriptCommand(t *testing.T) {
	input := `workflow Build
  goal: "Build"
  start: Setup
  exit: Done

  tool Setup
    command:
      #!/bin/sh
      set -eu
      if [ -f go.mod ]; then
        go build ./... 2>&1 || exit 1
        go test ./... 2>&1
      fi
      if [ -f package.json ]; then
        npm test 2>&1 || { echo "failed"; exit 1; }
      fi
      printf 'done'

  agent Done
    prompt: "Finish."

  edges
    Setup -> Done
`
	p := NewParser(input, "test.dip")
	w, err := p.Parse()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(w.Nodes) != 2 {
		t.Fatalf("nodes = %d, want 2", len(w.Nodes))
	}

	var toolNode *ir.Node
	for _, n := range w.Nodes {
		if n.Kind == ir.NodeTool {
			toolNode = n
			break
		}
	}
	if toolNode == nil {
		t.Fatal("tool node not found")
	}

	cmd := toolNode.Config.(ir.ToolConfig).Command
	// Must preserve shell syntax verbatim
	for _, want := range []string{
		"#!/bin/sh",
		"set -eu",
		"if [ -f go.mod ]; then",
		"  go build ./... 2>&1 || exit 1",
		"fi",
		`{ echo "failed"; exit 1; }`,
		"printf 'done'",
	} {
		if !strings.Contains(cmd, want) {
			t.Errorf("command missing %q\ngot:\n%s", want, cmd)
		}
	}
}

func TestParseColonInValue(t *testing.T) {
	input := `workflow Test
  goal: "Test"
  start: A
  exit: A

  defaults
    fidelity: summary:medium
    max_retries: 3

  agent A
    fidelity: summary:high
    prompt: "Do it."

  edges
    A -> A
`
	p := NewParser(input, "test.dip")
	w, err := p.Parse()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if w.Defaults.Fidelity != "summary:medium" {
		t.Errorf("defaults fidelity = %q, want %q", w.Defaults.Fidelity, "summary:medium")
	}

	if len(w.Nodes) != 1 {
		t.Fatalf("nodes = %d, want 1", len(w.Nodes))
	}
	cfg := w.Nodes[0].Config.(ir.AgentConfig)
	if cfg.Fidelity != "summary:high" {
		t.Errorf("agent fidelity = %q, want %q", cfg.Fidelity, "summary:high")
	}
}

func TestParseInvalidIntegerEmitsDiagnostic(t *testing.T) {
	input := `workflow Test
  goal: "Test"
  start: A
  exit: A

  defaults
    max_retries: three

  agent A
    max_turns: unlimited
    prompt: "Do it."

  edges
    A -> A
`
	p := NewParser(input, "test.dip")
	_, err := p.Parse()
	if err == nil {
		t.Fatal("expected error for invalid integers, got nil")
	}
	errStr := err.Error()
	if !strings.Contains(errStr, "invalid integer") {
		t.Errorf("expected 'invalid integer' diagnostic, got: %s", errStr)
	}
	if !strings.Contains(errStr, `"three"`) {
		t.Errorf("expected diagnostic to mention %q, got: %s", "three", errStr)
	}
	if !strings.Contains(errStr, `"unlimited"`) {
		t.Errorf("expected diagnostic to mention %q, got: %s", "unlimited", errStr)
	}
}

func TestParseInvalidDurationEmitsDiagnostic(t *testing.T) {
	input := `workflow Test
  goal: "Test"
  start: A
  exit: A

  tool A
    timeout: 60
    command:
      echo done

  edges
    A -> A
`
	p := NewParser(input, "test.dip")
	_, err := p.Parse()
	if err == nil {
		t.Fatal("expected error for invalid duration, got nil")
	}
	errStr := err.Error()
	if !strings.Contains(errStr, "invalid duration") {
		t.Errorf("expected 'invalid duration' diagnostic, got: %s", errStr)
	}
	if !strings.Contains(errStr, `"60"`) {
		t.Errorf("expected diagnostic to mention %q, got: %s", "60", errStr)
	}
}

func TestParseTabIndentation(t *testing.T) {
	// Verify that tab-indented files parse without panicking.
	// Tabs count as 1 byte of indentation (not 8).
	input := "workflow Test\n\tgoal: \"Test\"\n\tstart: A\n\texit: A\n\n\tagent A\n\t\tprompt: \"Do it.\"\n\n\tedges\n\t\tA -> A\n"
	p := NewParser(input, "test.dip")
	w, err := p.Parse()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if w.Name != "Test" {
		t.Errorf("name = %q, want %q", w.Name, "Test")
	}
	if len(w.Nodes) != 1 {
		t.Errorf("nodes = %d, want 1", len(w.Nodes))
	}
}

func TestParseMultilineBlockWithBlankLines(t *testing.T) {
	input := `workflow Test
  goal: "Test"
  start: A
  exit: A

  agent A
    prompt:
      Paragraph one.

      Paragraph two after blank line.

      Paragraph three.

  edges
    A -> A
`
	p := NewParser(input, "test.dip")
	w, err := p.Parse()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(w.Nodes) != 1 {
		t.Fatalf("nodes = %d, want 1", len(w.Nodes))
	}
	prompt := w.Nodes[0].Config.(ir.AgentConfig).Prompt
	if !strings.Contains(prompt, "Paragraph one.") {
		t.Errorf("prompt missing 'Paragraph one.', got: %q", prompt)
	}
	if !strings.Contains(prompt, "Paragraph two after blank line.") {
		t.Errorf("prompt missing 'Paragraph two after blank line.', got: %q", prompt)
	}
	if !strings.Contains(prompt, "Paragraph three.") {
		t.Errorf("prompt missing 'Paragraph three.', got: %q", prompt)
	}
	// Verify blank lines are preserved
	if strings.Count(prompt, "\n\n") < 2 {
		t.Errorf("expected at least 2 blank lines preserved in prompt, got: %q", prompt)
	}
}

func TestParseEdgeWithQuotedConditionValue(t *testing.T) {
	input := `workflow Test
  goal: "Test"
  start: A
  exit: B

  agent A
    prompt: "Do it."

  agent B
    prompt: "Done."

  edges
    A -> B when ctx.outcome == "success"
`
	p := NewParser(input, "test.dip")
	w, err := p.Parse()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(w.Edges) != 1 {
		t.Fatalf("edges = %d, want 1", len(w.Edges))
	}
	cond := w.Edges[0].Condition
	if cond == nil {
		t.Fatal("expected condition, got nil")
	}
	raw := cond.Raw
	if !strings.Contains(raw, "ctx.outcome") {
		t.Errorf("condition raw missing 'ctx.outcome', got: %q", raw)
	}
	if !strings.Contains(raw, `"success"`) {
		t.Errorf("condition raw missing quoted value, got: %q", raw)
	}
}

func TestParseComparisonOperators(t *testing.T) {
	input := `workflow Test
  goal: "Test"
  start: A
  exit: D

  agent A
    prompt: "A"
  agent B
    prompt: "B"
  agent C
    prompt: "C"
  agent D
    prompt: "D"

  edges
    A -> B when ctx.retries <= 3
    A -> C when ctx.score >= 80
    A -> D when ctx.count < 10
`
	p := NewParser(input, "test.dip")
	w, err := p.Parse()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(w.Edges) != 3 {
		t.Fatalf("edges = %d, want 3", len(w.Edges))
	}
	tests := []struct {
		idx  int
		want string
	}{
		{0, "<="},
		{1, ">="},
		{2, "<"},
	}
	for _, tt := range tests {
		raw := w.Edges[tt.idx].Condition.Raw
		if !strings.Contains(raw, tt.want) {
			t.Errorf("edge %d condition missing %q, got: %q", tt.idx, tt.want, raw)
		}
	}
}

func TestParseStylesheet(t *testing.T) {
	input := `workflow Test
  start: A
  exit: A

  agent A
    class: coder
    prompt: "Do it."

  stylesheet:
    *
      temperature: 0.7
    agent
      fidelity: full
    .coder
      model: o1
      reasoning_effort: medium
    #A
      max_retries: 5

  edges
    A -> A
`
	p := NewParser(input, "test.dip")
	w, err := p.Parse()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(w.Stylesheet) != 4 {
		t.Fatalf("stylesheet rules = %d, want 4", len(w.Stylesheet))
	}

	// Check universal selector
	found := false
	for _, r := range w.Stylesheet {
		if r.Selector.Kind == "universal" {
			found = true
			if r.Properties["temperature"] != "0.7" {
				t.Errorf("universal temperature = %q, want 0.7", r.Properties["temperature"])
			}
		}
	}
	if !found {
		t.Error("universal selector not found")
	}

	// Check class selector
	for _, r := range w.Stylesheet {
		if r.Selector.Kind == "class" && r.Selector.Value == "coder" {
			if r.Properties["model"] != "o1" {
				t.Errorf(".coder model = %q, want o1", r.Properties["model"])
			}
		}
	}

	// Check ID selector
	for _, r := range w.Stylesheet {
		if r.Selector.Kind == "id" && r.Selector.Value == "A" {
			if r.Properties["max_retries"] != "5" {
				t.Errorf("#A max_retries = %q, want 5", r.Properties["max_retries"])
			}
		}
	}
}

func TestParseParallelBlockForm(t *testing.T) {
	input := `workflow Test
  start: split
  exit: join

  agent worker_a
    prompt: "A"

  agent worker_b
    prompt: "B"

  parallel split
    branch: worker_a
      model: o1
    branch: worker_b
      model: claude-3

  fan_in join <- worker_a, worker_b

  edges
    join -> join
`
	p := NewParser(input, "test.dip")
	w, err := p.Parse()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var pNode *ir.Node
	for _, n := range w.Nodes {
		if n.Kind == ir.NodeParallel {
			pNode = n
			break
		}
	}
	if pNode == nil {
		t.Fatal("parallel node not found")
	}

	cfg := pNode.Config.(ir.ParallelConfig)
	if len(cfg.Targets) != 2 {
		t.Fatalf("targets = %d, want 2", len(cfg.Targets))
	}
	if len(cfg.Branches) != 2 {
		t.Fatalf("branches = %d, want 2", len(cfg.Branches))
	}
	if cfg.Branches[0].Model != "o1" {
		t.Errorf("branch[0].model = %q, want o1", cfg.Branches[0].Model)
	}
	if cfg.Branches[1].Model != "claude-3" {
		t.Errorf("branch[1].model = %q, want claude-3", cfg.Branches[1].Model)
	}
}

func TestParseParallelInlineStillWorks(t *testing.T) {
	input := `workflow Test
  start: P
  exit: J

  agent A
    prompt: "A"
  agent B
    prompt: "B"

  parallel P -> A, B
  fan_in J <- A, B

  edges
    J -> J
`
	p := NewParser(input, "test.dip")
	w, err := p.Parse()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var pNode *ir.Node
	for _, n := range w.Nodes {
		if n.Kind == ir.NodeParallel {
			pNode = n
			break
		}
	}
	if pNode == nil {
		t.Fatal("parallel node not found")
	}
	cfg := pNode.Config.(ir.ParallelConfig)
	if len(cfg.Targets) != 2 {
		t.Errorf("targets = %d, want 2", len(cfg.Targets))
	}
	if len(cfg.Branches) != 0 {
		t.Errorf("branches = %d, want 0 (inline form)", len(cfg.Branches))
	}
}

func TestParseOnResume(t *testing.T) {
	input := `workflow Test
  start: A
  exit: A

  defaults
    fidelity: full
    on_resume: degrade

  agent A
    prompt: "Do it."

  edges
    A -> A
`
	p := NewParser(input, "test.dip")
	w, err := p.Parse()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if w.Defaults.Fidelity != "full" {
		t.Errorf("fidelity = %q, want full", w.Defaults.Fidelity)
	}
	if w.Defaults.OnResume != "degrade" {
		t.Errorf("on_resume = %q, want degrade", w.Defaults.OnResume)
	}
}

func TestParseCompactionFields(t *testing.T) {
	input := `workflow Test
  start: A
  exit: A

  agent A
    compaction: sliding_window
    compaction_threshold: 0.8
    cache_tools: true
    prompt: "Do it."

  edges
    A -> A
`
	p := NewParser(input, "test.dip")
	w, err := p.Parse()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cfg := w.Nodes[0].Config.(ir.AgentConfig)
	if cfg.Compaction != "sliding_window" {
		t.Errorf("compaction = %q, want sliding_window", cfg.Compaction)
	}
	if cfg.CompactionThreshold != 0.8 {
		t.Errorf("compaction_threshold = %f, want 0.8", cfg.CompactionThreshold)
	}
	if !cfg.CacheTools {
		t.Error("cache_tools = false, want true")
	}
}

func TestParseToolOutputs(t *testing.T) {
	input := `workflow Test
  start: T
  exit: T

  tool T
    outputs: complete, continue, error
    timeout: 30s
    command:
      echo done

  edges
    T -> T
`
	p := NewParser(input, "test.dip")
	w, err := p.Parse()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(w.Nodes) != 1 {
		t.Fatalf("nodes = %d, want 1", len(w.Nodes))
	}
	cfg := w.Nodes[0].Config.(ir.ToolConfig)
	if len(cfg.Outputs) != 3 {
		t.Fatalf("outputs = %d, want 3", len(cfg.Outputs))
	}
	want := []string{"complete", "continue", "error"}
	for i, v := range want {
		if cfg.Outputs[i] != v {
			t.Errorf("outputs[%d] = %q, want %q", i, cfg.Outputs[i], v)
		}
	}
}

func TestParseToolWithoutOutputs(t *testing.T) {
	input := `workflow Test
  start: T
  exit: T

  tool T
    timeout: 30s
    command:
      echo done

  edges
    T -> T
`
	p := NewParser(input, "test.dip")
	w, err := p.Parse()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	cfg := w.Nodes[0].Config.(ir.ToolConfig)
	if len(cfg.Outputs) != 0 {
		t.Errorf("outputs = %d, want 0", len(cfg.Outputs))
	}
}

func TestParseSubgraphParams(t *testing.T) {
	input := `workflow Test
  start: Build
  exit: Done

  agent Build
    prompt: "Build."

  subgraph Review
    ref: ./review.dip
    params:
      language: python
      strict: true

  agent Done
    prompt: "Done."

  edges
    Build -> Review
    Review -> Done
`
	p := NewParser(input, "test.dip")
	w, err := p.Parse()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var sg *ir.Node
	for _, n := range w.Nodes {
		if n.Kind == ir.NodeSubgraph {
			sg = n
			break
		}
	}
	if sg == nil {
		t.Fatal("subgraph node not found")
	}

	cfg := sg.Config.(ir.SubgraphConfig)
	if cfg.Ref != "./review.dip" {
		t.Errorf("ref = %q, want ./review.dip", cfg.Ref)
	}
	if cfg.Params["language"] != "python" {
		t.Errorf("params[language] = %q, want python", cfg.Params["language"])
	}
	if cfg.Params["strict"] != "true" {
		t.Errorf("params[strict] = %q, want true", cfg.Params["strict"])
	}
}

func TestParseEdgeAttributes(t *testing.T) {
	input := readTestdata(t, "edge_attributes.dip")
	p := NewParser(input, "edge_attributes.dip")
	w, err := p.Parse()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(w.Edges) != 3 {
		t.Fatalf("edges = %d, want 3", len(w.Edges))
	}

	// Edge 0: A -> B with when, label, weight
	e0 := w.Edges[0]
	if e0.From != "A" || e0.To != "B" {
		t.Errorf("edge[0] = %s -> %s, want A -> B", e0.From, e0.To)
	}
	if e0.Condition == nil || !strings.Contains(e0.Condition.Raw, "ctx.status") {
		t.Errorf("edge[0] condition = %v, want ctx.status condition", e0.Condition)
	}
	if e0.Label != "approved" {
		t.Errorf("edge[0] label = %q, want %q", e0.Label, "approved")
	}
	if e0.Weight != 5 {
		t.Errorf("edge[0] weight = %d, want 5", e0.Weight)
	}

	// Edge 1: B -> C with restart
	e1 := w.Edges[1]
	if !e1.Restart {
		t.Error("edge[1] restart = false, want true")
	}

	// Edge 2: C -> D with label only
	e2 := w.Edges[2]
	if e2.Label != "final" {
		t.Errorf("edge[2] label = %q, want %q", e2.Label, "final")
	}
}

func TestParseRetryFields(t *testing.T) {
	input := readTestdata(t, "retry_fields.dip")
	p := NewParser(input, "retry_fields.dip")
	w, err := p.Parse()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(w.Nodes) != 1 {
		t.Fatalf("nodes = %d, want 1", len(w.Nodes))
	}
	n := w.Nodes[0]
	if n.Retry.Policy != "aggressive" {
		t.Errorf("retry_policy = %q, want aggressive", n.Retry.Policy)
	}
	if n.Retry.MaxRetries != 5 {
		t.Errorf("max_retries = %d, want 5", n.Retry.MaxRetries)
	}
	if n.Retry.BaseDelay.String() != "2s" {
		t.Errorf("base_delay = %v, want 2s", n.Retry.BaseDelay)
	}
	if n.Retry.RetryTarget != "A" {
		t.Errorf("retry_target = %q, want A", n.Retry.RetryTarget)
	}
	if n.Retry.FallbackTarget != "A" {
		t.Errorf("fallback_target = %q, want A", n.Retry.FallbackTarget)
	}
}

func TestParseParallelBranchFields(t *testing.T) {
	input := readTestdata(t, "parallel_branches.dip")
	p := NewParser(input, "parallel_branches.dip")
	w, err := p.Parse()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var pNode *ir.Node
	for _, n := range w.Nodes {
		if n.Kind == ir.NodeParallel {
			pNode = n
			break
		}
	}
	if pNode == nil {
		t.Fatal("parallel node not found")
	}

	cfg := pNode.Config.(ir.ParallelConfig)
	if len(cfg.Branches) != 2 {
		t.Fatalf("branches = %d, want 2", len(cfg.Branches))
	}

	b0 := cfg.Branches[0]
	if b0.Target != "fast" {
		t.Errorf("branch[0].target = %q, want fast", b0.Target)
	}
	if b0.Model != "claude-haiku-4-5" {
		t.Errorf("branch[0].model = %q, want claude-haiku-4-5", b0.Model)
	}
	if b0.Provider != "anthropic" {
		t.Errorf("branch[0].provider = %q, want anthropic", b0.Provider)
	}
	if b0.Fidelity != "summary" {
		t.Errorf("branch[0].fidelity = %q, want summary", b0.Fidelity)
	}

	b1 := cfg.Branches[1]
	if b1.Target != "accurate" {
		t.Errorf("branch[1].target = %q, want accurate", b1.Target)
	}
	if b1.Fidelity != "full" {
		t.Errorf("branch[1].fidelity = %q, want full", b1.Fidelity)
	}
}

func TestParseDefaultsComplex(t *testing.T) {
	input := readTestdata(t, "defaults_complex.dip")
	p := NewParser(input, "defaults_complex.dip")
	w, err := p.Parse()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	d := w.Defaults
	if d.MaxRetries != 3 {
		t.Errorf("max_retries = %d, want 3", d.MaxRetries)
	}
	if d.MaxRestarts != 10 {
		t.Errorf("max_restarts = %d, want 10", d.MaxRestarts)
	}
	if !d.CacheTools {
		t.Error("cache_tools = false, want true")
	}
	if d.RetryPolicy != "standard" {
		t.Errorf("retry_policy = %q, want standard", d.RetryPolicy)
	}
	if d.RestartTarget != "A" {
		t.Errorf("restart_target = %q, want A", d.RestartTarget)
	}
}

func TestParseHumanAndAgentFields(t *testing.T) {
	input := readTestdata(t, "human_node.dip")
	p := NewParser(input, "human_node.dip")
	w, err := p.Parse()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Find human node
	var humanNode, agentNode *ir.Node
	for _, n := range w.Nodes {
		if n.Kind == ir.NodeHuman {
			humanNode = n
		}
		if n.Kind == ir.NodeAgent {
			agentNode = n
		}
	}
	if humanNode == nil {
		t.Fatal("human node not found")
	}
	hcfg := humanNode.Config.(ir.HumanConfig)
	if hcfg.Mode != "choice" {
		t.Errorf("human mode = %q, want choice", hcfg.Mode)
	}
	if hcfg.Default != "approve" {
		t.Errorf("human default = %q, want approve", hcfg.Default)
	}

	if agentNode == nil {
		t.Fatal("agent node not found")
	}
	acfg := agentNode.Config.(ir.AgentConfig)
	if acfg.SystemPrompt != "You are helpful." {
		t.Errorf("system_prompt = %q, want 'You are helpful.'", acfg.SystemPrompt)
	}
	if acfg.Model != "claude-sonnet-4-6" {
		t.Errorf("model = %q, want claude-sonnet-4-6", acfg.Model)
	}
	if acfg.Provider != "anthropic" {
		t.Errorf("provider = %q, want anthropic", acfg.Provider)
	}
	if !acfg.GoalGate {
		t.Error("goal_gate = false, want true")
	}
	if !acfg.AutoStatus {
		t.Error("auto_status = false, want true")
	}
	if acfg.ReasoningEffort != "high" {
		t.Errorf("reasoning_effort = %q, want high", acfg.ReasoningEffort)
	}
	if acfg.MaxTurns != 10 {
		t.Errorf("max_turns = %d, want 10", acfg.MaxTurns)
	}
	if agentNode.Label != "Main Agent" {
		t.Errorf("label = %q, want 'Main Agent'", agentNode.Label)
	}
	if len(agentNode.Classes) != 2 {
		t.Errorf("classes = %d, want 2", len(agentNode.Classes))
	}
	if len(agentNode.IO.Reads) != 2 {
		t.Errorf("reads = %d, want 2", len(agentNode.IO.Reads))
	}
	if len(agentNode.IO.Writes) != 2 {
		t.Errorf("writes = %d, want 2", len(agentNode.IO.Writes))
	}
}

func TestParseHumanPrompt(t *testing.T) {
	input := readTestdata(t, "human_prompt.dip")
	p := NewParser(input, "human_prompt.dip")
	w, err := p.Parse()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var humanNode *ir.Node
	for _, n := range w.Nodes {
		if n.Kind == ir.NodeHuman {
			humanNode = n
			break
		}
	}
	if humanNode == nil {
		t.Fatal("human node not found")
	}

	cfg := humanNode.Config.(ir.HumanConfig)
	if cfg.Mode != "choice" {
		t.Errorf("mode = %q, want choice", cfg.Mode)
	}
	if cfg.Default != "approve" {
		t.Errorf("default = %q, want approve", cfg.Default)
	}
	if cfg.Prompt == "" {
		t.Fatal("prompt is empty, want multiline content")
	}
	for _, want := range []string{
		"Please review the proposed changes below.",
		"If everything looks correct, approve to continue.",
		"## Changes",
		"${ctx.diff_summary}",
	} {
		if !strings.Contains(cfg.Prompt, want) {
			t.Errorf("prompt missing %q\ngot:\n%s", want, cfg.Prompt)
		}
	}
	// Verify blank lines are preserved
	if strings.Count(cfg.Prompt, "\n\n") < 1 {
		t.Errorf("expected blank lines in prompt, got: %q", cfg.Prompt)
	}
}

func TestTokenKindString(t *testing.T) {
	tok := Token{
		Type:     TokenIdentifier,
		Value:    "hello",
		Location: ir.SourceLocation{Line: 1, Column: 5},
	}
	s := tok.String()
	if !strings.Contains(s, "hello") {
		t.Errorf("Token.String() = %q, expected it to contain 'hello'", s)
	}
	if !strings.Contains(s, "1:5") {
		t.Errorf("Token.String() = %q, expected it to contain '1:5'", s)
	}
}

func readTestdata(t *testing.T, name string) string {
	t.Helper()
	data, err := os.ReadFile("testdata/" + name)
	if err != nil {
		t.Fatalf("failed to read testdata/%s: %v", name, err)
	}
	return string(data)
}

func parseFixture(t *testing.T, name string) *ir.Workflow {
	t.Helper()
	src := readTestdata(t, name)
	p := NewParser(src, name)
	w, err := p.Parse()
	if err != nil {
		t.Fatalf("parse %s: %v", name, err)
	}
	return w
}

func findNode(t *testing.T, w *ir.Workflow, id string) *ir.Node {
	t.Helper()
	for _, n := range w.Nodes {
		if n.ID == id {
			return n
		}
	}
	t.Fatalf("node %q not found", id)
	return nil
}

func TestParseResponseFormat(t *testing.T) {
	w := parseFixture(t, "response_format.dip")

	// JsonAgent: response_format=json_object, empty response_schema
	jsonAgent := findNode(t, w, "JsonAgent")
	jcfg := jsonAgent.Config.(ir.AgentConfig)
	if jcfg.ResponseFormat != "json_object" {
		t.Errorf("JsonAgent ResponseFormat = %q, want json_object", jcfg.ResponseFormat)
	}
	if jcfg.ResponseSchema != "" {
		t.Errorf("JsonAgent ResponseSchema = %q, want empty", jcfg.ResponseSchema)
	}

	// SchemaAgent: response_format=json_schema, valid JSON response_schema
	schemaAgent := findNode(t, w, "SchemaAgent")
	scfg := schemaAgent.Config.(ir.AgentConfig)
	if scfg.ResponseFormat != "json_schema" {
		t.Errorf("SchemaAgent ResponseFormat = %q, want json_schema", scfg.ResponseFormat)
	}
	if scfg.ResponseSchema == "" {
		t.Fatal("SchemaAgent ResponseSchema is empty, want valid JSON")
	}
	var schemaObj map[string]interface{}
	if err := json.Unmarshal([]byte(scfg.ResponseSchema), &schemaObj); err != nil {
		t.Errorf("SchemaAgent ResponseSchema is not valid JSON: %v\ngot: %q", err, scfg.ResponseSchema)
	}

	// ParamsAgent: params with backend and permission_mode
	paramsAgent := findNode(t, w, "ParamsAgent")
	pcfg := paramsAgent.Config.(ir.AgentConfig)
	if pcfg.Params == nil {
		t.Fatal("ParamsAgent Params is nil")
	}
	if pcfg.Params["backend"] != "claude-code" {
		t.Errorf("ParamsAgent Params[backend] = %q, want claude-code", pcfg.Params["backend"])
	}
	if pcfg.Params["permission_mode"] != "auto" {
		t.Errorf("ParamsAgent Params[permission_mode] = %q, want auto", pcfg.Params["permission_mode"])
	}
}

func TestRoundtripResponseFormatFields(t *testing.T) {
	w1 := parseFixture(t, "response_format.dip")
	formatted := formatter.Format(w1)
	w2, err := NewParser(formatted, "roundtrip").Parse()
	if err != nil {
		t.Fatalf("failed to parse formatted output: %v", err)
	}

	// Verify response_format survives
	jsonNode := findNode(t, w2, "JsonAgent")
	jcfg := jsonNode.Config.(ir.AgentConfig)
	if jcfg.ResponseFormat != "json_object" {
		t.Errorf("round-trip: JsonAgent ResponseFormat = %q, want json_object", jcfg.ResponseFormat)
	}

	// Verify response_schema JSON survives with valid structure
	schemaNode := findNode(t, w2, "SchemaAgent")
	scfg := schemaNode.Config.(ir.AgentConfig)
	if scfg.ResponseFormat != "json_schema" {
		t.Errorf("round-trip: SchemaAgent ResponseFormat = %q, want json_schema", scfg.ResponseFormat)
	}
	if !json.Valid([]byte(scfg.ResponseSchema)) {
		t.Errorf("round-trip: SchemaAgent ResponseSchema is not valid JSON: %s", scfg.ResponseSchema)
	}

	// Verify params survive
	paramsNode := findNode(t, w2, "ParamsAgent")
	pcfg2 := paramsNode.Config.(ir.AgentConfig)
	if pcfg2.Params["backend"] != "claude-code" {
		t.Errorf("round-trip: ParamsAgent Params[backend] = %q, want claude-code", pcfg2.Params["backend"])
	}
	if pcfg2.Params["permission_mode"] != "auto" {
		t.Errorf("round-trip: ParamsAgent Params[permission_mode] = %q, want auto", pcfg2.Params["permission_mode"])
	}
}

func TestParseHumanInterview(t *testing.T) {
	w := parseFixture(t, "human_interview.dip")
	gate := findNode(t, w, "Gate")
	cfg, ok := gate.Config.(ir.HumanConfig)
	if !ok {
		t.Fatal("Gate is not HumanConfig")
	}
	if cfg.Mode != "interview" {
		t.Errorf("Mode = %q, want interview", cfg.Mode)
	}
	if cfg.QuestionsKey != "interview_questions" {
		t.Errorf("QuestionsKey = %q, want interview_questions", cfg.QuestionsKey)
	}
	if cfg.AnswersKey != "interview_answers" {
		t.Errorf("AnswersKey = %q, want interview_answers", cfg.AnswersKey)
	}
	if cfg.Prompt == "" {
		t.Error("expected non-empty prompt")
	}
}

func TestParseBracketEdgeSyntaxEmitsDiagnostic(t *testing.T) {
	input := `workflow Test
  goal: "Test"
  start: A
  exit: B

  agent A
    prompt: "Do A."

  agent B
    prompt: "Do B."

  edges
    A -> B [label: "go" condition: "ctx.x = 1"]
`
	p := NewParser(input, "test.dip")
	_, err := p.Parse()
	if err == nil {
		t.Fatal("expected parse error for bracket edge syntax, got nil")
	}
	errStr := err.Error()
	if !strings.Contains(errStr, "bracket") {
		t.Errorf("expected error to mention 'bracket', got: %s", errStr)
	}
	if !strings.Contains(errStr, "when") {
		t.Errorf("expected error to suggest 'when' keyword, got: %s", errStr)
	}
}

func TestParseBracketEdgeSyntaxRemainingEdgesParsed(t *testing.T) {
	// Even when a bracket-annotated edge triggers a diagnostic, subsequent edges
	// in the same block should still be parsed correctly.
	input := `workflow Test
  goal: "Test"
  start: A
  exit: C

  agent A
    prompt: "Do A."

  agent B
    prompt: "Do B."

  agent C
    prompt: "Do C."

  edges
    A -> B [label: "go"]
    B -> C
`
	p := NewParser(input, "test.dip")
	w, err := p.Parse()
	// Expect a parse error due to bracket syntax on the first edge.
	if err == nil {
		t.Fatal("expected parse error for bracket edge syntax, got nil")
	}
	// The second edge (B -> C) should still have been parsed.
	found := false
	for _, e := range w.Edges {
		if e.From == "B" && e.To == "C" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected edge B -> C to be parsed after bracket syntax error")
	}
}

func TestParseConditionalNode(t *testing.T) {
	input := `workflow Test
  goal: "Conditional routing"
  start: A
  exit: Done

  agent A
    prompt: "Analyze input."

  conditional Route
    label: "Route by Outcome"

  agent Done
    prompt: "Finish."

  edges
    A -> Route
    Route -> Done
`
	p := NewParser(input, "test.dip")
	w, err := p.Parse()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	route := w.Node("Route")
	if route == nil {
		t.Fatal("node Route not found")
	}
	if route.Kind != ir.NodeConditional {
		t.Errorf("kind = %q, want %q", route.Kind, ir.NodeConditional)
	}
	if _, ok := route.Config.(ir.ConditionalConfig); !ok {
		t.Errorf("config type = %T, want ConditionalConfig", route.Config)
	}
	if route.Label != "Route by Outcome" {
		t.Errorf("label = %q, want %q", route.Label, "Route by Outcome")
	}
}

func TestNestedRetryBlockError(t *testing.T) {
	input := `workflow Test
  start: A
  exit: B
  tool A
    label: "Verify"
    command: "verify.sh"
    timeout: 30s
    retry
      policy: aggressive
      max_retries: 5
      retry_target: process
  agent B
    prompt: done
  edges
    A -> B
`
	p := NewParser(input, "test.dip")
	_, err := p.Parse()
	if err == nil {
		t.Fatal("expected parse error for nested retry block")
	}
	if !strings.Contains(err.Error(), "retry") {
		t.Errorf("error should mention retry, got: %s", err.Error())
	}
	if !strings.Contains(err.Error(), "retry_policy") {
		t.Errorf("error should suggest retry_policy, got: %s", err.Error())
	}
	if !strings.Contains(err.Error(), "max_retries") {
		t.Errorf("error should suggest max_retries, got: %s", err.Error())
	}
}

func TestNestedRetryBlockRestOfWorkflowParses(t *testing.T) {
	input := `workflow Test
  start: A
  exit: B
  tool A
    label: "Verify"
    command: "verify.sh"
    timeout: 30s
    retry
      policy: aggressive
      max_retries: 5
      retry_target: process
  agent B
    prompt: done
  edges
    A -> B
`
	p := NewParser(input, "test.dip")
	w, _ := p.Parse()
	// Even with the error, the parser should recover and parse remaining nodes and edges.
	nodeB := w.Node("B")
	if nodeB == nil {
		t.Fatal("node B should be parsed despite retry error in node A")
	}
	if len(w.Edges) == 0 {
		t.Error("edges should be parsed despite retry error in node A")
	}
}

func TestParseAgentBackendField(t *testing.T) {
	input := `workflow Test
  start: A
  exit: B
  agent A
    backend: claude-code
    prompt: do it
  agent B
    prompt: done
  edges
    A -> B
`
	p := NewParser(input, "test.dip")
	w, err := p.Parse()
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	node := w.Node("A")
	if node == nil {
		t.Fatal("node A not found")
	}
	cfg, ok := node.Config.(ir.AgentConfig)
	if !ok {
		t.Fatalf("expected AgentConfig, got %T", node.Config)
	}
	if cfg.Backend != "claude-code" {
		t.Errorf("expected backend 'claude-code', got %q", cfg.Backend)
	}
}

func TestParseUnrecognizedAgentFieldEmitsDiagnostic(t *testing.T) {
	input := `workflow Test
  start: A
  exit: B
  agent A
    foo: bar
    prompt: do it
  agent B
    prompt: done
  edges
    A -> B
`
	p := NewParser(input, "test.dip")
	_, err := p.Parse()
	if err == nil {
		t.Fatal("expected parse error for unrecognized field 'foo'")
	}
	if !strings.Contains(err.Error(), "unrecognized") {
		t.Errorf("error should mention 'unrecognized', got: %s", err.Error())
	}
	if !strings.Contains(err.Error(), "foo") {
		t.Errorf("error should mention field name 'foo', got: %s", err.Error())
	}
	if !strings.Contains(err.Error(), "params") {
		t.Errorf("error should suggest 'params:', got: %s", err.Error())
	}
}

func TestParseUnrecognizedToolFieldEmitsDiagnostic(t *testing.T) {
	input := `workflow Test
  start: A
  exit: B
  tool A
    foo: bar
    command: echo hi
    timeout: 5s
  agent B
    prompt: done
  edges
    A -> B
`
	p := NewParser(input, "test.dip")
	_, err := p.Parse()
	if err == nil {
		t.Fatal("expected parse error for unrecognized tool field 'foo'")
	}
	if !strings.Contains(err.Error(), "unrecognized") {
		t.Errorf("error should mention 'unrecognized', got: %s", err.Error())
	}
	if !strings.Contains(err.Error(), "tool") {
		t.Errorf("error should mention node kind 'tool', got: %s", err.Error())
	}
	if strings.Contains(err.Error(), "params") {
		t.Errorf("tool nodes don't support params — hint should NOT mention params, got: %s", err.Error())
	}
}

func TestParseUnrecognizedSubgraphFieldSuggestsParams(t *testing.T) {
	input := `workflow Test
  start: A
  exit: B
  subgraph A
    foo: bar
    ref: child.dip
  agent B
    prompt: done
  edges
    A -> B
`
	p := NewParser(input, "test.dip")
	_, err := p.Parse()
	if err == nil {
		t.Fatal("expected parse error for unrecognized subgraph field 'foo'")
	}
	if !strings.Contains(err.Error(), "unrecognized") {
		t.Errorf("error should mention 'unrecognized', got: %s", err.Error())
	}
	if !strings.Contains(err.Error(), "params") {
		t.Errorf("subgraph nodes support params — hint should mention params, got: %s", err.Error())
	}
}

func TestParseAgentWorkingDirField(t *testing.T) {
	input := `workflow Test
  start: A
  exit: B
  agent A
    working_dir: .ai/worktrees/claude
    prompt: do it
  agent B
    prompt: done
  edges
    A -> B
`
	p := NewParser(input, "test.dip")
	w, err := p.Parse()
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	node := w.Node("A")
	if node == nil {
		t.Fatal("node A not found")
	}
	cfg, ok := node.Config.(ir.AgentConfig)
	if !ok {
		t.Fatalf("expected AgentConfig, got %T", node.Config)
	}
	if cfg.WorkingDir != ".ai/worktrees/claude" {
		t.Errorf("expected working_dir '.ai/worktrees/claude', got %q", cfg.WorkingDir)
	}
}

func TestParseVarsBlock(t *testing.T) {
	input := `workflow Test
  start: A
  exit: B

  vars
    api_url: https://example.com/api
    env: production
    retries: 3

  agent A
    prompt: do it
  agent B
    prompt: done
  edges
    A -> B
`
	p := NewParser(input, "test.dip")
	w, err := p.Parse()
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	if len(w.Vars) != 3 {
		t.Fatalf("expected 3 vars, got %d", len(w.Vars))
	}
	if w.Vars["api_url"] != "https://example.com/api" {
		t.Errorf("api_url = %q, want %q", w.Vars["api_url"], "https://example.com/api")
	}
	if w.Vars["env"] != "production" {
		t.Errorf("env = %q, want %q", w.Vars["env"], "production")
	}
	if w.Vars["retries"] != "3" {
		t.Errorf("retries = %q, want %q", w.Vars["retries"], "3")
	}
}

func TestParseVarsDuplicateKey(t *testing.T) {
	input := `workflow Test
  start: A
  exit: B

  vars
    key: first
    key: second

  agent A
    prompt: do it
  agent B
    prompt: done
  edges
    A -> B
`
	p := NewParser(input, "test.dip")
	_, err := p.Parse()
	if err == nil {
		t.Fatal("expected error for duplicate vars key, got nil")
	}
	if !strings.Contains(err.Error(), "duplicate vars key") {
		t.Errorf("error %q does not mention duplicate vars key", err.Error())
	}
}

func TestParseVarsRoundTrip(t *testing.T) {
	input := `workflow Test
  start: A
  exit: B

  vars
    source_ref: "references/sdk-python/src"
    target_name: my-crate

  agent A
    prompt: go

  agent B
    prompt: done

  edges
    A -> B
`
	p := NewParser(input, "test.dip")
	w, err := p.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	formatted := formatter.Format(w)
	p2 := NewParser(formatted, "roundtrip.dip")
	w2, err := p2.Parse()
	if err != nil {
		t.Fatalf("re-parse error: %v\nformatted:\n%s", err, formatted)
	}
	if len(w2.Vars) != 2 {
		t.Errorf("expected 2 vars after round-trip, got %d", len(w2.Vars))
	}
	if w2.Vars["source_ref"] != "references/sdk-python/src" {
		t.Errorf("source_ref lost in round-trip: %q", w2.Vars["source_ref"])
	}
	if w2.Vars["target_name"] != "my-crate" {
		t.Errorf("target_name lost in round-trip: %q", w2.Vars["target_name"])
	}
}

func TestParseDefaultsBudget(t *testing.T) {
	w := parseFixture(t, "defaults_budget.dip")
	d := w.Defaults
	if d.MaxTotalTokens != 500000 {
		t.Errorf("max_total_tokens = %d, want 500000", d.MaxTotalTokens)
	}
	if d.MaxCostCents != 1000 {
		t.Errorf("max_cost_cents = %d, want 1000", d.MaxCostCents)
	}
	if d.MaxWallTime != 30*time.Minute {
		t.Errorf("max_wall_time = %v, want 30m0s", d.MaxWallTime)
	}
}

func TestParseHumanTimeout(t *testing.T) {
	w := parseFixture(t, "human_timeout.dip")
	gate := findNode(t, w, "Gate")
	cfg, ok := gate.Config.(ir.HumanConfig)
	if !ok {
		t.Fatal("Gate is not HumanConfig")
	}
	if cfg.Timeout != 5*time.Minute {
		t.Errorf("Timeout = %v, want 5m0s", cfg.Timeout)
	}
	if cfg.TimeoutAction != "default" {
		t.Errorf("TimeoutAction = %q, want default", cfg.TimeoutAction)
	}
}

func TestParseHumanTimeoutRoundTrip(t *testing.T) {
	w1 := parseFixture(t, "human_timeout.dip")
	formatted := formatter.Format(w1)
	w2, err := NewParser(formatted, "roundtrip").Parse()
	if err != nil {
		t.Fatalf("re-parse error: %v\nformatted:\n%s", err, formatted)
	}
	gate := findNode(t, w2, "Gate")
	cfg, ok := gate.Config.(ir.HumanConfig)
	if !ok {
		t.Fatal("round-trip: Gate is not HumanConfig")
	}
	if cfg.Timeout != 5*time.Minute {
		t.Errorf("round-trip: Timeout = %v, want 5m0s", cfg.Timeout)
	}
	if cfg.TimeoutAction != "default" {
		t.Errorf("round-trip: TimeoutAction = %q, want default", cfg.TimeoutAction)
	}
}

func TestParseHumanTimeoutActionInvalid(t *testing.T) {
	src := `workflow Test
  start: Gate
  exit: Gate

  human Gate
    mode: choice
    timeout_action: explode
`
	p := NewParser(src, "test.dip")
	_, err := p.Parse()
	if err == nil {
		t.Fatal("expected parse error for invalid timeout_action, got nil")
	}
	if !strings.Contains(err.Error(), "invalid timeout_action") {
		t.Errorf("expected error to mention invalid timeout_action, got: %v", err)
	}
}

func TestParseDefaultsBudgetRoundTrip(t *testing.T) {
	w1 := parseFixture(t, "defaults_budget.dip")
	formatted := formatter.Format(w1)
	w2, err := NewParser(formatted, "roundtrip").Parse()
	if err != nil {
		t.Fatalf("re-parse error: %v\nformatted:\n%s", err, formatted)
	}
	d := w2.Defaults
	if d.MaxTotalTokens != 500000 {
		t.Errorf("round-trip: max_total_tokens = %d, want 500000", d.MaxTotalTokens)
	}
	if d.MaxCostCents != 1000 {
		t.Errorf("round-trip: max_cost_cents = %d, want 1000", d.MaxCostCents)
	}
	if d.MaxWallTime != 30*time.Minute {
		t.Errorf("round-trip: max_wall_time = %v, want 30m0s", d.MaxWallTime)
	}
}

func TestParseManagerLoopNode(t *testing.T) {
	src := `workflow W
  start: M
  exit: M

  manager_loop M
    label: "Quality Gate Supervisor"
    subgraph_ref: quality_loop
    poll_interval: 30s
    max_cycles: 12
    stop_condition: stack.child.cycles = 10
    steer_condition: stack.child.cycles = 5

  edges
    M -> M
`
	p := NewParser(src, "")
	w, err := p.Parse()
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	if len(w.Nodes) != 1 {
		t.Fatalf("got %d nodes, want 1", len(w.Nodes))
	}
	n := w.Nodes[0]
	if n.Kind != ir.NodeManagerLoop {
		t.Errorf("Kind = %q, want %q", n.Kind, ir.NodeManagerLoop)
	}
	cfg, ok := n.Config.(ir.ManagerLoopConfig)
	if !ok {
		t.Fatalf("Config = %T, want ManagerLoopConfig", n.Config)
	}
	if cfg.SubgraphRef != "quality_loop" {
		t.Errorf("SubgraphRef = %q, want %q", cfg.SubgraphRef, "quality_loop")
	}
	if cfg.PollInterval != 30*time.Second {
		t.Errorf("PollInterval = %v, want 30s", cfg.PollInterval)
	}
	if cfg.MaxCycles != 12 {
		t.Errorf("MaxCycles = %d, want 12", cfg.MaxCycles)
	}
	if n.Label != "Quality Gate Supervisor" {
		t.Errorf("Label = %q", n.Label)
	}
	if cfg.StopCondition == nil || cfg.StopCondition.Raw != "stack.child.cycles = 10" {
		t.Errorf("StopCondition.Raw = %+v", cfg.StopCondition)
	}
	if cfg.SteerCondition == nil || cfg.SteerCondition.Raw != "stack.child.cycles = 5" {
		t.Errorf("SteerCondition.Raw = %+v", cfg.SteerCondition)
	}
}

func TestParseManagerLoop_SteerContextInline(t *testing.T) {
	src := `workflow W
  start: M
  exit: M

  manager_loop M
    subgraph_ref: inner
    steer_context: hint=speed_up, priority=high

  edges
    M -> M
`
	p := NewParser(src, "")
	w, err := p.Parse()
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	cfg := w.Nodes[0].Config.(ir.ManagerLoopConfig)
	if cfg.SteerContext["hint"] != "speed_up" || cfg.SteerContext["priority"] != "high" {
		t.Errorf("SteerContext = %v", cfg.SteerContext)
	}
}

func TestParseManagerLoop_SteerContextBlock(t *testing.T) {
	src := `workflow W
  start: M
  exit: M

  manager_loop M
    subgraph_ref: inner
    steer_context:
      hint: speed_up
      priority: high

  edges
    M -> M
`
	p := NewParser(src, "")
	w, err := p.Parse()
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	cfg := w.Nodes[0].Config.(ir.ManagerLoopConfig)
	if cfg.SteerContext["hint"] != "speed_up" || cfg.SteerContext["priority"] != "high" {
		t.Errorf("SteerContext = %v", cfg.SteerContext)
	}
}

func TestParseManagerLoop_UnknownFieldHint(t *testing.T) {
	src := `workflow W
  start: M
  exit: M

  manager_loop M
    subgraph_ref: inner
    bogus_field: value

  edges
    M -> M
`
	p := NewParser(src, "")
	_, err := p.Parse()
	if err == nil {
		t.Fatalf("expected parse error containing unknown-field hint; got nil")
	}
	if !strings.Contains(err.Error(), "manager_loop") || !strings.Contains(err.Error(), "bogus_field") {
		t.Errorf("expected error mentioning manager_loop and bogus_field; got %v", err)
	}
}

func TestParseManagerLoop_SteerContextInline_CommaInValueTruncates(t *testing.T) {
	// Documents the limitation of the inline form: a value containing a comma
	// is split at that comma. Users needing comma-containing values must use
	// the block form. This test exists so future refactors don't silently
	// change the behavior without a deliberate decision.
	src := `workflow W
  start: M
  exit: M

  manager_loop M
    subgraph_ref: inner
    steer_context: hint="speed, up", priority=high

  edges
    M -> M
`
	p := NewParser(src, "")
	_, _ = p.Parse() // non-nil err expected (malformed entry diagnostic)
	cfg := p.workflow.Nodes[0].Config.(ir.ManagerLoopConfig)
	// hint is truncated to `"speed` (the part before the comma, quote preserved).
	if cfg.SteerContext["hint"] != `"speed` {
		t.Errorf("hint = %q, want %q (documents inline-form truncation)", cfg.SteerContext["hint"], `"speed`)
	}
	// priority parses normally because it's after the second comma.
	if cfg.SteerContext["priority"] != "high" {
		t.Errorf("priority = %q, want %q", cfg.SteerContext["priority"], "high")
	}
	// A "malformed entry" diagnostic is emitted for the middle fragment.
	gotMalformed := false
	for _, d := range p.diagnostics {
		if strings.Contains(d, "must be key=value") {
			gotMalformed = true
		}
	}
	if !gotMalformed {
		t.Errorf("expected malformed-entry diagnostic; got %v", p.diagnostics)
	}
}

func TestParseManagerLoop_SteerContextBlock_SingleEntry(t *testing.T) {
	// Regression: a single-entry block-form steer_context produces a raw
	// block value with no embedded newline. The pre-fix parser routed that
	// to the inline CSV handler and emitted a "must be key=value" diagnostic.
	src := `workflow W
  start: M
  exit: M

  manager_loop M
    subgraph_ref: inner
    steer_context:
      hint: halfway_through

  edges
    M -> M
`
	p := NewParser(src, "")
	_, err := p.Parse()
	if err != nil {
		t.Fatalf("Parse returned error (diagnostics): %v", err)
	}
	if len(p.diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics on single-entry block form: %v", p.diagnostics)
	}
	cfg := p.workflow.Nodes[0].Config.(ir.ManagerLoopConfig)
	if cfg.SteerContext["hint"] != "halfway_through" {
		t.Errorf("SteerContext[hint] = %q, want %q", cfg.SteerContext["hint"], "halfway_through")
	}
	if len(cfg.SteerContext) != 1 {
		t.Errorf("SteerContext size = %d, want 1: %v", len(cfg.SteerContext), cfg.SteerContext)
	}
}

func TestParseManagerLoop_SteerContextKeyWithColon(t *testing.T) {
	// Keys containing ':' break the block-form round-trip (splitKeyValue splits
	// on first colon), so the parser rejects them with a diagnostic and drops
	// the entry. Valid keys in the same entry are still parsed.
	//
	// Note: isSteerContextBlockForm routes to block form when ':' comes before '='
	// in the raw string. To get a ':'-containing key into the inline parser, '='
	// must appear before ':' in the first token — e.g. "hint=ok, ns:weird=value".
	src := `workflow W
  start: M
  exit: M

  manager_loop M
    subgraph_ref: inner
    steer_context: hint=ok, ns:weird=value

  edges
    M -> M
`
	p := NewParser(src, "")
	_, _ = p.Parse()
	cfg := p.workflow.Nodes[0].Config.(ir.ManagerLoopConfig)
	if _, bad := cfg.SteerContext["ns:weird"]; bad {
		t.Errorf("key with ':' should have been dropped, got %v", cfg.SteerContext)
	}
	if cfg.SteerContext["hint"] != "ok" {
		t.Errorf("hint should still parse: %v", cfg.SteerContext)
	}
	found := false
	for _, d := range p.diagnostics {
		if strings.Contains(d, `"ns:weird"`) && strings.Contains(d, ":") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected diagnostic about key with ':'; got %v", p.diagnostics)
	}
}

func TestParseDefaultsToolSafety(t *testing.T) {
	w := parseFixture(t, "defaults_tool_safety.dip")
	d := w.Defaults
	if d.ToolCommandsAllow != "git *,make *,npm test" {
		t.Errorf("tool_commands_allow = %q, want %q", d.ToolCommandsAllow, "git *,make *,npm test")
	}
	if d.ToolDenylistAdd != "rm -rf /,dd *" {
		t.Errorf("tool_denylist_add = %q, want %q", d.ToolDenylistAdd, "rm -rf /,dd *")
	}
}

func TestParseDefaultsToolSafetyRoundTrip(t *testing.T) {
	w1 := parseFixture(t, "defaults_tool_safety.dip")
	formatted := formatter.Format(w1)
	w2, err := NewParser(formatted, "roundtrip").Parse()
	if err != nil {
		t.Fatalf("re-parse error: %v\nformatted:\n%s", err, formatted)
	}
	d := w2.Defaults
	if d.ToolCommandsAllow != "git *,make *,npm test" {
		t.Errorf("round-trip: tool_commands_allow = %q, want %q", d.ToolCommandsAllow, "git *,make *,npm test")
	}
	if d.ToolDenylistAdd != "rm -rf /,dd *" {
		t.Errorf("round-trip: tool_denylist_add = %q, want %q", d.ToolDenylistAdd, "rm -rf /,dd *")
	}
}

func TestParseRequiresSingle(t *testing.T) {
	input := `workflow Test
  goal: "Test"
  requires: git
  start: A
  exit: A

  agent A
    prompt: "Do it."

  edges
    A -> A
`
	w, err := NewParser(input, "test.dip").Parse()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(w.Requires) != 1 || w.Requires[0] != "git" {
		t.Errorf("Requires = %#v, want [git]", w.Requires)
	}
}

func TestParseRequiresMulti(t *testing.T) {
	input := `workflow Test
  goal: "Test"
  requires: git, docker, jq
  start: A
  exit: A

  agent A
    prompt: "Do it."

  edges
    A -> A
`
	w, err := NewParser(input, "test.dip").Parse()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []string{"git", "docker", "jq"}
	if len(w.Requires) != len(want) {
		t.Fatalf("Requires = %#v, want %#v", w.Requires, want)
	}
	for i, v := range want {
		if w.Requires[i] != v {
			t.Errorf("Requires[%d] = %q, want %q", i, w.Requires[i], v)
		}
	}
}

func TestParseRequiresTrimsAndDropsEmpties(t *testing.T) {
	input := `workflow Test
  goal: "Test"
  requires:   git ,  ,  docker  ,
  start: A
  exit: A

  agent A
    prompt: "Do it."

  edges
    A -> A
`
	w, err := NewParser(input, "test.dip").Parse()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []string{"git", "docker"}
	if len(w.Requires) != len(want) {
		t.Fatalf("Requires = %#v, want %#v", w.Requires, want)
	}
	for i, v := range want {
		if w.Requires[i] != v {
			t.Errorf("Requires[%d] = %q, want %q", i, w.Requires[i], v)
		}
	}
}

func TestParseRequiresMissingIsNil(t *testing.T) {
	input := `workflow Test
  goal: "Test"
  start: A
  exit: A

  agent A
    prompt: "Do it."

  edges
    A -> A
`
	w, err := NewParser(input, "test.dip").Parse()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if w.Requires != nil {
		t.Errorf("Requires = %#v, want nil", w.Requires)
	}
}

func TestParseRequiresUnknownEntriesAccepted(t *testing.T) {
	// Unknown / future dependency names must not produce parser diagnostics.
	input := `workflow Test
  goal: "Test"
  requires: future-thing-9000, quantum-entangler
  start: A
  exit: A

  agent A
    prompt: "Do it."

  edges
    A -> A
`
	w, err := NewParser(input, "test.dip").Parse()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(w.Requires) != 2 {
		t.Errorf("Requires = %#v, want 2 entries", w.Requires)
	}
}

func TestParseRequiresRoundTrip(t *testing.T) {
	input := `workflow Test
  goal: "Test"
  requires: git, docker, jq
  start: A
  exit: A

  agent A
    prompt: "Do it."

  edges
    A -> A
`
	w1, err := NewParser(input, "test.dip").Parse()
	if err != nil {
		t.Fatalf("first parse: %v", err)
	}
	formatted := formatter.Format(w1)
	w2, err := NewParser(formatted, "roundtrip").Parse()
	if err != nil {
		t.Fatalf("re-parse: %v\nformatted:\n%s", err, formatted)
	}
	if len(w1.Requires) != len(w2.Requires) {
		t.Fatalf("Requires len: %d vs %d", len(w1.Requires), len(w2.Requires))
	}
	for i := range w1.Requires {
		if w1.Requires[i] != w2.Requires[i] {
			t.Errorf("Requires[%d]: %q vs %q", i, w1.Requires[i], w2.Requires[i])
		}
	}
}

func TestParseRequiresFromFixture(t *testing.T) {
	w := parseFixture(t, "workflow_requires.dip")
	want := []string{"git", "docker", "jq"}
	if len(w.Requires) != len(want) {
		t.Fatalf("Requires = %#v, want %#v", w.Requires, want)
	}
	for i, v := range want {
		if w.Requires[i] != v {
			t.Errorf("Requires[%d] = %q, want %q", i, w.Requires[i], v)
		}
	}
}

func TestParseAgent_ToolAccess(t *testing.T) {
	cases := []struct {
		name string
		src  string
		want string
	}{
		{
			name: "explicit none",
			src: `workflow X
  start: A
  exit: A

  agent A
    prompt: "x"
    tool_access: none
`,
			want: "none",
		},
		{
			name: "case variant",
			src: `workflow X
  start: A
  exit: A

  agent A
    prompt: "x"
    tool_access: None
`,
			want: "None",
		},
		{
			name: "invalid value stored verbatim",
			src: `workflow X
  start: A
  exit: A

  agent A
    prompt: "x"
    tool_access: foo
`,
			want: "foo",
		},
		{
			name: "quoted",
			src: `workflow X
  start: A
  exit: A

  agent A
    prompt: "x"
    tool_access: "none"
`,
			want: "none",
		},
		{
			name: "omitted (default)",
			src: `workflow X
  start: A
  exit: A

  agent A
    prompt: "x"
`,
			want: "",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p := NewParser(tc.src, "test.dip")
			w, err := p.Parse()
			if err != nil {
				t.Fatalf("parse error: %v", err)
			}
			node := w.Node("A")
			if node == nil {
				t.Fatalf("node A not found")
			}
			cfg, ok := node.Config.(ir.AgentConfig)
			if !ok {
				t.Fatalf("expected AgentConfig, got %T", node.Config)
			}
			if cfg.ToolAccess != tc.want {
				t.Errorf("ToolAccess = %q, want %q", cfg.ToolAccess, tc.want)
			}
		})
	}
}

func TestParseTool_CommandFile(t *testing.T) {
	cases := []struct {
		name             string
		src              string
		wantFile         string
		wantCommand      string
		wantDiagContains string // substring expected in any parser diagnostic; "" = no diagnostic expected
	}{
		{
			name: "command_file only",
			src: `workflow X
  start: A
  exit: A

  tool A
    command_file: scripts/setup.sh
`,
			wantFile:    "scripts/setup.sh",
			wantCommand: "",
		},
		{
			name: "command only (unchanged behavior)",
			src: `workflow X
  start: A
  exit: A

  tool A
    command: echo hi
`,
			wantFile:    "",
			wantCommand: "echo hi",
		},
		{
			name: "both set (command first)",
			src: `workflow X
  start: A
  exit: A

  tool A
    command: echo hi
    command_file: scripts/setup.sh
`,
			wantDiagContains: "both `command` and `command_file`",
		},
		{
			name: "both set (command_file first)",
			src: `workflow X
  start: A
  exit: A

  tool A
    command_file: scripts/setup.sh
    command: echo hi
`,
			wantDiagContains: "both `command` and `command_file`",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p := NewParser(tc.src, "test.dip")
			w, _ := p.Parse() // some cases expect diagnostics; don't fail on parse error
			if tc.wantDiagContains != "" {
				found := false
				for _, d := range p.diagnostics {
					if strings.Contains(d, tc.wantDiagContains) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected diagnostic containing %q; got %v", tc.wantDiagContains, p.diagnostics)
				}
				return
			}
			if w == nil {
				t.Fatalf("unexpected nil workflow")
			}
			node := w.Node("A")
			if node == nil {
				t.Fatalf("node A not found")
			}
			cfg, ok := node.Config.(ir.ToolConfig)
			if !ok {
				t.Fatalf("expected ToolConfig, got %T", node.Config)
			}
			if cfg.CommandFile != tc.wantFile {
				t.Errorf("CommandFile = %q, want %q", cfg.CommandFile, tc.wantFile)
			}
			if cfg.Command != tc.wantCommand {
				t.Errorf("Command = %q, want %q", cfg.Command, tc.wantCommand)
			}
		})
	}
}

func TestParser_AgentPromptFile(t *testing.T) {
	src := `workflow W
  goal: "test"
  start: A
  exit: A

  agent A
    model: claude-sonnet-4-6
    prompt_file: prompts/task.md
`
	w, err := NewParser(src, "test.dip").Parse()
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if len(w.Nodes) != 1 {
		t.Fatalf("got %d nodes, want 1", len(w.Nodes))
	}
	cfg, ok := w.Nodes[0].Config.(ir.AgentConfig)
	if !ok {
		t.Fatalf("node 0 config is not AgentConfig: %T", w.Nodes[0].Config)
	}
	if cfg.PromptFile != "prompts/task.md" {
		t.Errorf("PromptFile = %q, want %q", cfg.PromptFile, "prompts/task.md")
	}
	if cfg.Prompt != "" {
		t.Errorf("Prompt = %q, want empty (parser must not load file)", cfg.Prompt)
	}
}

func TestParser_AgentSystemPromptFile(t *testing.T) {
	src := `workflow W
  goal: "test"
  start: A
  exit: A

  agent A
    model: claude-sonnet-4-6
    system_prompt_file: prompts/persona.md
`
	w, err := NewParser(src, "test.dip").Parse()
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	cfg := w.Nodes[0].Config.(ir.AgentConfig)
	if cfg.SystemPromptFile != "prompts/persona.md" {
		t.Errorf("SystemPromptFile = %q, want %q", cfg.SystemPromptFile, "prompts/persona.md")
	}
	if cfg.SystemPrompt != "" {
		t.Errorf("SystemPrompt = %q, want empty (parser must not load file)", cfg.SystemPrompt)
	}
}

func TestParser_AgentPromptConflict(t *testing.T) {
	src := `workflow W
  goal: "test"
  start: A
  exit: A

  agent A
    model: claude-sonnet-4-6
    prompt: "inline"
    prompt_file: prompts/task.md
`
	_, err := NewParser(src, "test.dip").Parse()
	if err == nil {
		t.Fatal("expected diagnostic for prompt + prompt_file conflict, got nil error")
	}
	if !strings.Contains(err.Error(), "both `prompt` and `prompt_file`") {
		t.Errorf("diagnostic message missing expected phrase; got %q", err.Error())
	}
}

func TestParser_AgentSystemPromptConflict(t *testing.T) {
	src := `workflow W
  goal: "test"
  start: A
  exit: A

  agent A
    model: claude-sonnet-4-6
    system_prompt: "inline persona"
    system_prompt_file: prompts/persona.md
`
	_, err := NewParser(src, "test.dip").Parse()
	if err == nil {
		t.Fatal("expected diagnostic for system_prompt + system_prompt_file conflict, got nil error")
	}
	if !strings.Contains(err.Error(), "both `system_prompt` and `system_prompt_file`") {
		t.Errorf("diagnostic message missing expected phrase; got %q", err.Error())
	}
}

func TestParser_AgentCrossSlotMixOK(t *testing.T) {
	src := `workflow W
  goal: "test"
  start: A
  exit: A

  agent A
    model: claude-sonnet-4-6
    prompt_file: prompts/task.md
    system_prompt: "you are a code reviewer"
`
	w, err := NewParser(src, "test.dip").Parse()
	if err != nil {
		t.Fatalf("mixed slots should not error: %v", err)
	}
	cfg := w.Nodes[0].Config.(ir.AgentConfig)
	if cfg.PromptFile != "prompts/task.md" {
		t.Errorf("PromptFile = %q", cfg.PromptFile)
	}
	if cfg.SystemPrompt != "you are a code reviewer" {
		t.Errorf("SystemPrompt = %q", cfg.SystemPrompt)
	}
}
