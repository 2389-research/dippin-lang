package cost

import (
	"fmt"
	"testing"

	"github.com/2389-research/dippin-lang/ir"
)

func testPricing() PricingTable {
	return DefaultPricing()
}

// TestDefaultPricingNewAnthropicModels guards the #116 pricing refresh: the new
// flagship IDs must be priced so `dippin cost` can estimate them. Mythos
// Preview is intentionally absent (no published per-token price).
func TestDefaultPricingNewAnthropicModels(t *testing.T) {
	anthropic := DefaultPricing()["anthropic"]
	want := map[string]ModelPrice{
		"claude-opus-4-8": {InputPer1M: 5.00, OutputPer1M: 25.00},
		"claude-fable-5":  {InputPer1M: 10.00, OutputPer1M: 50.00},
		"claude-mythos-5": {InputPer1M: 10.00, OutputPer1M: 50.00},
	}
	for model, price := range want {
		got, ok := anthropic[model]
		if !ok {
			t.Errorf("%s missing from anthropic pricing", model)
			continue
		}
		if got != price {
			t.Errorf("%s price = %+v, want %+v", model, got, price)
		}
	}
}

func TestSingleAgentNode(t *testing.T) {
	w := &ir.Workflow{
		Start: "a1",
		Exit:  "a1",
		Defaults: ir.WorkflowDefaults{
			Provider: "anthropic",
		},
		Nodes: []*ir.Node{
			{
				ID:   "a1",
				Kind: ir.NodeAgent,
				Config: ir.AgentConfig{
					Prompt:   "Write a haiku about Go.",
					Model:    "claude-sonnet-4-6",
					MaxTurns: 9,
				},
			},
		},
	}

	r := Analyze(w, testPricing())

	nc, ok := r.Nodes["a1"]
	if !ok {
		t.Fatal("expected node a1 in report")
	}
	if nc.Model != "claude-sonnet-4-6" {
		t.Errorf("model = %q, want claude-sonnet-4-6", nc.Model)
	}
	if nc.Provider != "anthropic" {
		t.Errorf("provider = %q, want anthropic", nc.Provider)
	}
	if nc.Cost.Min <= 0 {
		t.Errorf("expected min cost > 0, got %f", nc.Cost.Min)
	}
	if nc.Cost.Max < nc.Cost.Expected || nc.Cost.Expected < nc.Cost.Min {
		t.Errorf("cost ordering wrong: min=%f expected=%f max=%f",
			nc.Cost.Min, nc.Cost.Expected, nc.Cost.Max)
	}
}

func TestDefaultModelFromWorkflow(t *testing.T) {
	w := &ir.Workflow{
		Start: "a1",
		Exit:  "a1",
		Defaults: ir.WorkflowDefaults{
			Model:    "claude-haiku-3-5",
			Provider: "anthropic",
		},
		Nodes: []*ir.Node{
			{
				ID:     "a1",
				Kind:   ir.NodeAgent,
				Config: ir.AgentConfig{Prompt: "Hello world"},
			},
		},
	}

	r := Analyze(w, testPricing())

	nc := r.Nodes["a1"]
	if nc.Model != "claude-haiku-3-5" {
		t.Errorf("model = %q, want claude-haiku-3-5", nc.Model)
	}
	if nc.Provider != "anthropic" {
		t.Errorf("provider = %q, want anthropic", nc.Provider)
	}
	if nc.Cost.Min <= 0 {
		t.Errorf("expected cost > 0 for known model, got %f", nc.Cost.Min)
	}
}

func TestToolNodeZeroCost(t *testing.T) {
	w := &ir.Workflow{
		Start: "t1",
		Exit:  "t1",
		Nodes: []*ir.Node{
			{
				ID:     "t1",
				Kind:   ir.NodeTool,
				Config: ir.ToolConfig{Command: "echo hello"},
			},
		},
	}

	r := Analyze(w, testPricing())

	nc := r.Nodes["t1"]
	if nc.Cost.Min != 0 || nc.Cost.Expected != 0 || nc.Cost.Max != 0 {
		t.Errorf("tool node should have $0 cost, got %+v", nc.Cost)
	}
}

func TestParallelBranchesSummed(t *testing.T) {
	w := &ir.Workflow{
		Start: "p1",
		Exit:  "join",
		Defaults: ir.WorkflowDefaults{
			Provider: "anthropic",
		},
		Nodes: []*ir.Node{
			{
				ID:   "p1",
				Kind: ir.NodeParallel,
				Config: ir.ParallelConfig{
					Targets: []string{"a1", "a2"},
				},
			},
			{
				ID:   "a1",
				Kind: ir.NodeAgent,
				Config: ir.AgentConfig{
					Prompt: "Branch one task",
					Model:  "claude-haiku-3-5",
				},
			},
			{
				ID:   "a2",
				Kind: ir.NodeAgent,
				Config: ir.AgentConfig{
					Prompt: "Branch two task",
					Model:  "claude-haiku-3-5",
				},
			},
			{
				ID:   "join",
				Kind: ir.NodeFanIn,
				Config: ir.FanInConfig{
					Sources: []string{"a1", "a2"},
				},
			},
		},
	}

	r := Analyze(w, testPricing())

	a1Cost := r.Nodes["a1"].Cost.Expected
	a2Cost := r.Nodes["a2"].Cost.Expected
	if a1Cost <= 0 || a2Cost <= 0 {
		t.Fatalf("expected both branches to have cost > 0")
	}

	// Total should include both branches.
	if r.Total.Expected < a1Cost+a2Cost {
		t.Errorf("total expected=%f should be >= a1+a2=%f",
			r.Total.Expected, a1Cost+a2Cost)
	}
}

func TestRestartLoopMultiplier(t *testing.T) {
	w := &ir.Workflow{
		Start: "a1",
		Exit:  "done",
		Defaults: ir.WorkflowDefaults{
			Provider: "anthropic",
		},
		Nodes: []*ir.Node{
			{
				ID:   "a1",
				Kind: ir.NodeAgent,
				Config: ir.AgentConfig{
					Prompt: "Do work",
					Model:  "claude-sonnet-4-6",
				},
			},
			{
				ID:     "done",
				Kind:   ir.NodeTool,
				Config: ir.ToolConfig{Command: "echo done"},
			},
		},
		Edges: []*ir.Edge{
			{From: "done", To: "a1", Restart: true},
		},
	}

	// Get base cost without restart.
	wNoLoop := &ir.Workflow{
		Start: "a1",
		Exit:  "a1",
		Defaults: ir.WorkflowDefaults{
			Provider: "anthropic",
		},
		Nodes: []*ir.Node{
			{
				ID:   "a1",
				Kind: ir.NodeAgent,
				Config: ir.AgentConfig{
					Prompt: "Do work",
					Model:  "claude-sonnet-4-6",
				},
			},
		},
	}

	rNoLoop := Analyze(wNoLoop, testPricing())
	baseCost := rNoLoop.Nodes["a1"].Cost.Max

	rLoop := Analyze(w, testPricing())
	loopCost := rLoop.Nodes["a1"].Cost.Max

	// Loop cost should be strictly greater due to multiplier.
	if loopCost <= baseCost {
		t.Errorf("loop max cost (%f) should be > base max cost (%f)", loopCost, baseCost)
	}
}

func TestUnknownModelZeroCost(t *testing.T) {
	w := &ir.Workflow{
		Start: "a1",
		Exit:  "a1",
		Defaults: ir.WorkflowDefaults{
			Provider: "mystery-corp",
		},
		Nodes: []*ir.Node{
			{
				ID:   "a1",
				Kind: ir.NodeAgent,
				Config: ir.AgentConfig{
					Prompt: "Explain quantum computing",
					Model:  "mystery-model-9000",
				},
			},
		},
	}

	r := Analyze(w, testPricing())

	nc := r.Nodes["a1"]
	if nc.Cost.Min != 0 || nc.Cost.Expected != 0 || nc.Cost.Max != 0 {
		t.Errorf("unknown model should have $0 cost, got %+v", nc.Cost)
	}
	if len(r.Assumptions) == 0 {
		t.Error("expected assumptions about unknown model")
	}
}

func TestBuildLoopRangeWithMaxRestarts(t *testing.T) {
	// When MaxRestarts is set on the workflow, buildLoopRange should be used.
	w := &ir.Workflow{
		Start: "a1",
		Exit:  "done",
		Defaults: ir.WorkflowDefaults{
			Provider:    "anthropic",
			MaxRestarts: 8,
		},
		Nodes: []*ir.Node{
			{
				ID:   "a1",
				Kind: ir.NodeAgent,
				Config: ir.AgentConfig{
					Prompt: "Do work",
					Model:  "claude-sonnet-4-6",
				},
			},
			{
				ID:     "done",
				Kind:   ir.NodeTool,
				Config: ir.ToolConfig{Command: "echo done"},
			},
		},
		Edges: []*ir.Edge{
			{From: "done", To: "a1", Restart: true},
		},
	}

	r := Analyze(w, testPricing())
	nc := r.Nodes["a1"]
	// With MaxRestarts=8, expected = max(8/2, 2) = 4, max = 8.
	// Cost should be multiplied accordingly.
	if nc.Cost.Max <= 0 {
		t.Errorf("expected positive cost with loop multiplier, got %f", nc.Cost.Max)
	}
}

func TestBuildLoopRangeSmallMaxRestarts(t *testing.T) {
	// When MaxRestarts is small (e.g., 2), expected should clamp to 2.
	w := &ir.Workflow{
		Start: "a1",
		Exit:  "done",
		Defaults: ir.WorkflowDefaults{
			Provider:    "anthropic",
			MaxRestarts: 2,
		},
		Nodes: []*ir.Node{
			{
				ID:   "a1",
				Kind: ir.NodeAgent,
				Config: ir.AgentConfig{
					Prompt: "Do work",
					Model:  "claude-sonnet-4-6",
				},
			},
			{
				ID:     "done",
				Kind:   ir.NodeTool,
				Config: ir.ToolConfig{Command: "echo done"},
			},
		},
		Edges: []*ir.Edge{
			{From: "done", To: "a1", Restart: true},
		},
	}

	r := Analyze(w, testPricing())
	nc := r.Nodes["a1"]
	if nc.Cost.Max <= 0 {
		t.Errorf("expected positive cost, got %f", nc.Cost.Max)
	}
}

func TestGetModelProviderNonAgentNode(t *testing.T) {
	// getModelProvider for a non-agent node falls back to workflow defaults.
	w := &ir.Workflow{
		Start: "h1",
		Exit:  "h1",
		Defaults: ir.WorkflowDefaults{
			Model:    "claude-sonnet-4-6",
			Provider: "anthropic",
		},
		Nodes: []*ir.Node{
			{
				ID:     "h1",
				Kind:   ir.NodeHuman,
				Config: ir.HumanConfig{Mode: "freeform"},
			},
		},
	}

	r := Analyze(w, testPricing())
	nc := r.Nodes["h1"]
	// Human nodes are non-agent, so cost should be zero.
	if nc.Cost.Min != 0 || nc.Cost.Expected != 0 || nc.Cost.Max != 0 {
		t.Errorf("human node should have $0 cost, got %+v", nc.Cost)
	}
}

func TestEstimateTurnsHighMaxTurns(t *testing.T) {
	// When maxTurns is high enough that expected = maxTurns/3 >= 3.
	w := &ir.Workflow{
		Start: "a1",
		Exit:  "a1",
		Defaults: ir.WorkflowDefaults{
			Provider: "anthropic",
		},
		Nodes: []*ir.Node{
			{
				ID:   "a1",
				Kind: ir.NodeAgent,
				Config: ir.AgentConfig{
					Prompt:   "Do complex work",
					Model:    "claude-sonnet-4-6",
					MaxTurns: 30,
				},
			},
		},
	}

	r := Analyze(w, testPricing())
	nc := r.Nodes["a1"]
	// With MaxTurns=30, expected = 30/3 = 10, which is >= 3.
	if nc.Cost.Max <= nc.Cost.Expected {
		t.Errorf("max should be > expected: max=%f expected=%f", nc.Cost.Max, nc.Cost.Expected)
	}
}

func TestSortTopCostsTruncation(t *testing.T) {
	// More than 5 agent nodes should truncate TopCosts to 5.
	nodes := make([]*ir.Node, 7)
	for i := range nodes {
		nodes[i] = &ir.Node{
			ID:   fmt.Sprintf("a%d", i),
			Kind: ir.NodeAgent,
			Config: ir.AgentConfig{
				Prompt: fmt.Sprintf("Task %d with some text to vary cost", i),
				Model:  "claude-sonnet-4-6",
			},
		}
	}

	w := &ir.Workflow{
		Start: "a0",
		Exit:  "a6",
		Defaults: ir.WorkflowDefaults{
			Provider: "anthropic",
		},
		Nodes: nodes,
	}

	r := Analyze(w, testPricing())
	if len(r.TopCosts) != 5 {
		t.Errorf("expected 5 top costs, got %d", len(r.TopCosts))
	}
}

func TestEstimateTurnsZeroMaxTurns(t *testing.T) {
	// When MaxTurns is 0, defaults to 10, then expected = 10/3 = 3.
	ac := ir.AgentConfig{Prompt: "test", MaxTurns: 0}
	turns := estimateTurns(ac)
	if turns.Max != 10 {
		t.Errorf("max = %d, want 10", turns.Max)
	}
	if turns.Expected != 3 {
		t.Errorf("expected = %d, want 3", turns.Expected)
	}
	if turns.Min != 3 {
		t.Errorf("min = %d, want 3", turns.Min)
	}
}

func TestEstimateTurnsLowMaxTurns(t *testing.T) {
	// When MaxTurns is 6, expected = 6/3 = 2, which is < 3, so clamped to 3.
	ac := ir.AgentConfig{Prompt: "test", MaxTurns: 6}
	turns := estimateTurns(ac)
	if turns.Max != 6 {
		t.Errorf("max = %d, want 6", turns.Max)
	}
	if turns.Expected != 3 {
		t.Errorf("expected = %d, want 3 (clamped)", turns.Expected)
	}
}

func TestTopCostsSorting(t *testing.T) {
	w := &ir.Workflow{
		Start: "cheap",
		Exit:  "expensive",
		Defaults: ir.WorkflowDefaults{
			Provider: "anthropic",
		},
		Nodes: []*ir.Node{
			{
				ID:   "cheap",
				Kind: ir.NodeAgent,
				Config: ir.AgentConfig{
					Prompt: "Hi",
					Model:  "claude-haiku-3-5",
				},
			},
			{
				ID:   "expensive",
				Kind: ir.NodeAgent,
				Config: ir.AgentConfig{
					Prompt: "Write a comprehensive analysis of the entire history of computing.",
					Model:  "claude-opus-4-6",
				},
			},
		},
	}

	r := Analyze(w, testPricing())

	if len(r.TopCosts) < 2 {
		t.Fatalf("expected at least 2 top costs, got %d", len(r.TopCosts))
	}
	if r.TopCosts[0].Cost.Max < r.TopCosts[1].Cost.Max {
		t.Errorf("top costs not sorted descending: %f < %f",
			r.TopCosts[0].Cost.Max, r.TopCosts[1].Cost.Max)
	}
	if r.TopCosts[0].NodeID != "expensive" {
		t.Errorf("expected top cost to be 'expensive', got %q", r.TopCosts[0].NodeID)
	}
}

// TestLookupPriceDottedDashedEquivalence covers #188: the Anthropic pricing
// keys are dashed (claude-haiku-4-5) but the Vercel-gateway-documented ID a
// .dip must carry is dotted (claude-haiku-4.5). Both spellings must resolve to
// the same price; a genuinely unknown model must still miss.
func TestLookupPriceDottedDashedEquivalence(t *testing.T) {
	pricing := DefaultPricing()

	dashed, ok := lookupPrice("anthropic", "claude-haiku-4-5", pricing)
	if !ok {
		t.Fatal("dashed claude-haiku-4-5 should be known")
	}
	dotted, ok := lookupPrice("anthropic", "claude-haiku-4.5", pricing)
	if !ok {
		t.Fatal("dotted claude-haiku-4.5 must resolve (issue #188)")
	}
	if dotted != dashed {
		t.Errorf("dotted price %+v != dashed price %+v", dotted, dashed)
	}

	// Exact dotted keys (OpenAI/Gemini) must still resolve unchanged.
	if _, ok := lookupPrice("openai", "gpt-5.5", pricing); !ok {
		t.Error("exact dotted OpenAI key gpt-5.5 regressed")
	}

	// A genuinely unknown model must still miss (no false positive from normalization).
	if _, ok := lookupPrice("anthropic", "claude-nonexistent-9-9", pricing); ok {
		t.Error("unknown model must not resolve")
	}
}

// TestNewFrontierProvidersPriced covers #189: current frontier models across
// all providers — including the new zai/moonshot/minimax providers — must
// resolve to a non-zero price. Guards against the "unknown model → $0" gap.
func TestNewFrontierProvidersPriced(t *testing.T) {
	p := DefaultPricing()
	cases := []struct{ provider, model string }{
		{"anthropic", "claude-opus-5"},
		{"anthropic", "claude-sonnet-5"},
		{"openai", "gpt-5.6-sol"},
		{"openai", "gpt-5.6-terra"},
		{"gemini", "gemini-3.6-flash"},
		{"xai", "grok-4.5"},
		{"grok", "grok-4.5"},
		{"zai", "glm-5.2"},
		{"moonshot", "kimi-k3"},
		{"kimi", "kimi-k3"},
		{"minimax", "MiniMax-M3"},
	}
	for _, c := range cases {
		price, ok := lookupPrice(c.provider, c.model, p)
		if !ok {
			t.Errorf("%s/%s not in pricing table", c.provider, c.model)
			continue
		}
		if price.InputPer1M <= 0 || price.OutputPer1M <= 0 {
			t.Errorf("%s/%s priced at %+v, want non-zero", c.provider, c.model, price)
		}
	}
}
