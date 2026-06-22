// Package optimize analyzes workflows for model cost optimization opportunities.
//
// It applies rule-based heuristics to identify nodes where a cheaper model
// could be used without sacrificing quality, or where a more capable model
// is warranted for complex prompts.
package optimize

import (
	"github.com/2389-research/dippin-lang/cost"
	"github.com/2389-research/dippin-lang/ir"
)

// Report is the full optimization analysis result.
type Report struct {
	Suggestions   []Suggestion   `json:"suggestions"`
	CurrentCost   cost.CostRange `json:"current_cost"`
	OptimizedCost cost.CostRange `json:"optimized_cost"`
	Savings       cost.CostRange `json:"savings"`
}

// Suggestion is a single optimization recommendation.
type Suggestion struct {
	NodeID       string         `json:"node_id"`
	Rule         string         `json:"rule"`
	Message      string         `json:"message"`
	CurrentModel string         `json:"current_model"`
	SuggestModel string         `json:"suggest_model"`
	Savings      cost.CostRange `json:"savings"`
}

// Analyze produces an optimization Report for the given workflow.
func Analyze(w *ir.Workflow, pricing cost.PricingTable) *Report {
	costReport := cost.Analyze(w, pricing)
	r := &Report{CurrentCost: costReport.Total}

	for _, n := range w.Nodes {
		ac, ok := n.Config.(ir.AgentConfig)
		if !ok {
			continue
		}
		model, provider := resolveModelProvider(ac, w)
		ctx := ruleContext{node: n, config: ac, model: model, provider: provider, workflow: w, pricing: pricing, costReport: costReport}
		r.Suggestions = append(r.Suggestions, applyRules(ctx)...)
	}

	r.OptimizedCost = estimateOptimizedCost(r, costReport)
	r.Savings = computeSavings(r.CurrentCost, r.OptimizedCost)
	return r
}

// ruleContext bundles the data each rule needs.
type ruleContext struct {
	node       *ir.Node
	config     ir.AgentConfig
	model      string
	provider   string
	workflow   *ir.Workflow
	pricing    cost.PricingTable
	costReport *cost.Report
}

// resolveModelProvider gets the effective model and provider for an agent node.
func resolveModelProvider(ac ir.AgentConfig, w *ir.Workflow) (string, string) {
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

// applyRules runs all optimization rules against a single node.
func applyRules(ctx ruleContext) []Suggestion {
	rules := []func(ruleContext) *Suggestion{
		ruleSimplePromptExpensiveModel,
		ruleHighIterationRetry,
		ruleBookkeepingNode,
		ruleComplexPromptCheapModel,
	}
	var suggestions []Suggestion
	for _, rule := range rules {
		if s := rule(ctx); s != nil {
			suggestions = append(suggestions, *s)
		}
	}
	return suggestions
}

// estimateOptimizedCost applies all suggestion savings to the current cost.
func estimateOptimizedCost(r *Report, costReport *cost.Report) cost.CostRange {
	totalSavings := cost.CostRange{}
	for _, s := range r.Suggestions {
		totalSavings.Min += s.Savings.Min
		totalSavings.Expected += s.Savings.Expected
		totalSavings.Max += s.Savings.Max
	}
	return cost.CostRange{
		Min:      clampZero(r.CurrentCost.Min - totalSavings.Min),
		Expected: clampZero(r.CurrentCost.Expected - totalSavings.Expected),
		Max:      clampZero(r.CurrentCost.Max - totalSavings.Max),
	}
}

// computeSavings calculates the difference between current and optimized cost.
func computeSavings(current, optimized cost.CostRange) cost.CostRange {
	return cost.CostRange{
		Min:      current.Min - optimized.Min,
		Expected: current.Expected - optimized.Expected,
		Max:      current.Max - optimized.Max,
	}
}

func clampZero(v float64) float64 {
	if v < 0 {
		return 0
	}
	return v
}
