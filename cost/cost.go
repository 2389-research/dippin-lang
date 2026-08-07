// Package cost estimates workflow execution costs based on model pricing.
//
// It analyzes an IR workflow and produces a Report with per-node cost
// estimates, aggregated totals by provider, and a sorted list of the
// most expensive nodes. All estimates use heuristic token counts and
// turn ranges.
package cost

import (
	"fmt"
	"sort"
	"strings"

	"github.com/2389-research/dippin-lang/ir"
)

// ModelPrice holds per-token pricing for a model.
type ModelPrice struct {
	InputPer1M  float64 `json:"input_per_1m"`
	OutputPer1M float64 `json:"output_per_1m"`
}

// PricingTable maps provider -> model -> price.
type PricingTable map[string]map[string]ModelPrice

// Report is the full cost analysis result.
type Report struct {
	Nodes       map[string]NodeCost  `json:"nodes"`
	Total       CostRange            `json:"total"`
	ByProvider  map[string]CostRange `json:"by_provider"`
	TopCosts    []NodeCost           `json:"top_costs"`
	Assumptions []string             `json:"assumptions"`
}

// NodeCost is the estimated cost for a single node.
type NodeCost struct {
	NodeID   string        `json:"node_id"`
	Model    string        `json:"model"`
	Provider string        `json:"provider"`
	Kind     string        `json:"kind"`
	Cost     CostRange     `json:"cost"`
	Turns    TurnRange     `json:"turns"`
	Tokens   TokenEstimate `json:"tokens"`
}

// CostRange holds min/expected/max cost in USD.
type CostRange struct {
	Min      float64 `json:"min"`
	Expected float64 `json:"expected"`
	Max      float64 `json:"max"`
}

// TurnRange holds estimated turn counts.
type TurnRange struct {
	Min      int `json:"min"`
	Expected int `json:"expected"`
	Max      int `json:"max"`
}

// TokenEstimate holds token usage estimates.
type TokenEstimate struct {
	PromptTokens  int `json:"prompt_tokens"`
	InputPerTurn  int `json:"input_per_turn"`
	OutputPerTurn int `json:"output_per_turn"`
}

// Analyze produces a cost Report for the given workflow and pricing table.
func Analyze(w *ir.Workflow, pricing PricingTable) *Report {
	r := &Report{
		Nodes:      make(map[string]NodeCost),
		ByProvider: make(map[string]CostRange),
	}

	for _, n := range w.Nodes {
		nc := estimateNodeCost(n, w, pricing, r)
		r.Nodes[n.ID] = nc
	}

	r.Total = aggregateTotal(r.Nodes)
	r.ByProvider = aggregateByProvider(r.Nodes)
	r.TopCosts = sortTopCosts(r.Nodes)
	return r
}

// aggregateTotal sums all node costs into a single CostRange.
func aggregateTotal(nodes map[string]NodeCost) CostRange {
	var total CostRange
	for _, nc := range nodes {
		total = AddCostRange(total, nc.Cost)
	}
	return total
}

// aggregateByProvider groups and sums costs by provider.
func aggregateByProvider(nodes map[string]NodeCost) map[string]CostRange {
	byProv := make(map[string]CostRange)
	for _, nc := range nodes {
		if nc.Provider == "" {
			continue
		}
		byProv[nc.Provider] = AddCostRange(byProv[nc.Provider], nc.Cost)
	}
	return byProv
}

// AddCostRange adds two CostRange values.
func AddCostRange(a, b CostRange) CostRange {
	return CostRange{
		Min:      a.Min + b.Min,
		Expected: a.Expected + b.Expected,
		Max:      a.Max + b.Max,
	}
}

// topCostLimit caps how many of the most expensive nodes Analyze reports.
const topCostLimit = 5

// sortTopCosts returns the topCostLimit nodes sorted by max cost descending.
func sortTopCosts(nodes map[string]NodeCost) []NodeCost {
	all := make([]NodeCost, 0, len(nodes))
	for _, nc := range nodes {
		all = append(all, nc)
	}
	sort.Slice(all, func(i, j int) bool {
		return all[i].Cost.Max > all[j].Cost.Max
	})
	if len(all) > topCostLimit {
		all = all[:topCostLimit]
	}
	return all
}

// estimateNodeCost computes cost for a single node.
func estimateNodeCost(n *ir.Node, w *ir.Workflow, pricing PricingTable, r *Report) NodeCost {
	nc := NodeCost{
		NodeID: n.ID,
		Kind:   string(n.Kind),
	}

	ac, ok := n.Config.(ir.AgentConfig)
	if !ok {
		return nc // non-agent nodes have zero cost
	}

	nc.Model, nc.Provider = getModelProvider(ac, w)
	nc.Tokens = estimateTokens(ac, nc.Model)
	nc.Turns = estimateTurns(ac)

	price, found := lookupPrice(nc.Provider, nc.Model, pricing)
	if !found {
		r.Assumptions = append(r.Assumptions,
			fmt.Sprintf("unknown model %q (provider %q): cost set to $0", nc.Model, nc.Provider))
		return nc
	}

	nc.Cost = computeCostRange(nc.Tokens, nc.Turns, price)
	nc.Cost = applyLoopMultiplier(n.ID, w, nc.Cost)
	return nc
}

// getModelProvider resolves the model and provider for an agent node.
func getModelProvider(ac ir.AgentConfig, w *ir.Workflow) (string, string) {
	model := ac.Model
	if model == "" {
		model = w.Defaults.Model
	}
	provider := ac.Provider
	if provider == "" {
		provider = w.Defaults.Provider
	}
	return model, provider
}

// estimateTokens produces a TokenEstimate from the agent config and model.
func estimateTokens(ac ir.AgentConfig, model string) TokenEstimate {
	promptTokens := len(ac.Prompt) / 4
	return TokenEstimate{
		PromptTokens:  promptTokens,
		InputPerTurn:  promptTokens + 500 + 1500,
		OutputPerTurn: outputTokensForModel(model),
	}
}

// estimateTurns produces a TurnRange from the agent config.
func estimateTurns(ac ir.AgentConfig) TurnRange {
	maxTurns := ac.MaxTurns
	if maxTurns <= 0 {
		maxTurns = 10
	}
	expected := maxTurns / 3
	if expected < 3 {
		expected = 3
	}
	return TurnRange{Min: 3, Expected: expected, Max: maxTurns}
}

// modelTierTokens maps model name substrings to estimated output tokens per turn.
var modelTierTokens = []struct {
	pattern string
	tokens  int
}{
	{"opus", 1500},
	{"sonnet", 1000},
	{"gpt-5", 1000},
	{"codex", 1000},
	{"gpt-4o", 1000},
	{"haiku", 400},
	{"mini", 400},
}

// outputTokensForModel returns an estimated output-tokens-per-turn for a model.
func outputTokensForModel(model string) int {
	lower := strings.ToLower(model)
	for _, tier := range modelTierTokens {
		if strings.Contains(lower, tier.pattern) {
			return tier.tokens
		}
	}
	return 800 // default mid-tier
}

// lookupPrice finds pricing for a provider/model pair.
//
// An exact key match wins. On a miss it retries with a version-separator-
// insensitive comparison so a dotted ID (claude-haiku-4.5, the Vercel AI
// Gateway spelling) resolves to a dashed catalog key (claude-haiku-4-5), and
// vice versa (issue #188). No two keys in any provider collide under this
// normalization, so the fallback can only turn a miss into the intended hit.
func lookupPrice(provider, model string, pricing PricingTable) (ModelPrice, bool) {
	models, ok := pricing[provider]
	if !ok {
		return ModelPrice{}, false
	}
	if p, ok := models[model]; ok {
		return p, true
	}
	want := canonicalModelID(model)
	for k, p := range models {
		if canonicalModelID(k) == want {
			return p, true
		}
	}
	return ModelPrice{}, false
}

// canonicalModelID folds the model version separator so dotted and dashed
// spellings compare equal (claude-haiku-4.5 == claude-haiku-4-5). See #188.
func canonicalModelID(model string) string {
	return strings.ReplaceAll(model, ".", "-")
}

// computeCostRange calculates cost for each turn scenario.
func computeCostRange(tokens TokenEstimate, turns TurnRange, price ModelPrice) CostRange {
	calc := func(numTurns int) float64 {
		input := float64(tokens.InputPerTurn*numTurns) * price.InputPer1M / 1_000_000
		output := float64(tokens.OutputPerTurn*numTurns) * price.OutputPer1M / 1_000_000
		return input + output
	}
	return CostRange{
		Min:      calc(turns.Min),
		Expected: calc(turns.Expected),
		Max:      calc(turns.Max),
	}
}

// applyLoopMultiplier scales cost if the node is in a restart cycle.
func applyLoopMultiplier(nodeID string, w *ir.Workflow, cost CostRange) CostRange {
	mult := findLoopMultiplier(nodeID, w)
	if mult.Max <= 1 {
		return cost
	}
	return CostRange{
		Min:      cost.Min * float64(mult.Min),
		Expected: cost.Expected * float64(mult.Expected),
		Max:      cost.Max * float64(mult.Max),
	}
}

// findLoopMultiplier checks if nodeID is the target of a restart edge.
func findLoopMultiplier(nodeID string, w *ir.Workflow) TurnRange {
	if !isRestartTarget(nodeID, w) {
		return TurnRange{Min: 1, Expected: 1, Max: 1}
	}
	maxRestarts := w.Defaults.MaxRestarts
	if maxRestarts > 0 {
		return buildLoopRange(maxRestarts)
	}
	return TurnRange{Min: 2, Expected: 5, Max: 10}
}

// isRestartTarget returns true if any restart edge points to nodeID.
func isRestartTarget(nodeID string, w *ir.Workflow) bool {
	for _, e := range w.Edges {
		if e.Restart && e.To == nodeID {
			return true
		}
	}
	return false
}

// buildLoopRange creates a TurnRange from a maxRestarts value.
func buildLoopRange(maxRestarts int) TurnRange {
	expected := maxRestarts / 2
	if expected < 2 {
		expected = 2
	}
	return TurnRange{Min: 2, Expected: expected, Max: maxRestarts}
}
