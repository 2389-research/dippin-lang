package optimize

import (
	"github.com/2389-research/dippin-lang/cost"
	"github.com/2389-research/dippin-lang/ir"
)

// ruleSimplePromptExpensiveModel flags simple prompts on expensive models.
func ruleSimplePromptExpensiveModel(ctx ruleContext) *Suggestion {
	if !isExpensiveModel(ctx.model) {
		return nil
	}
	score := promptComplexity(ctx.config.Prompt)
	if score >= complexityThresholdHigh {
		return nil
	}
	suggested := suggestCheaperModel(ctx.provider)
	savings := estimateNodeSavings(ctx, suggested)
	return &Suggestion{
		NodeID:       ctx.node.ID,
		Rule:         "simple-prompt-expensive-model",
		Message:      "simple prompt does not need an expensive model",
		CurrentModel: ctx.model,
		SuggestModel: suggested,
		Savings:      savings,
	}
}

// ruleHighIterationRetry flags nodes in retry loops that use expensive models.
func ruleHighIterationRetry(ctx ruleContext) *Suggestion {
	if !isInRetryLoop(ctx.node.ID, ctx.workflow) {
		return nil
	}
	if isCheapModel(ctx.model) {
		return nil
	}
	suggested := suggestCheaperModel(ctx.provider)
	savings := estimateNodeSavings(ctx, suggested)
	return &Suggestion{
		NodeID:       ctx.node.ID,
		Rule:         "high-iteration-retry",
		Message:      "node in retry loop — consider a cheaper model for mechanical iterations",
		CurrentModel: ctx.model,
		SuggestModel: suggested,
		Savings:      savings,
	}
}

// ruleBookkeepingNode flags bookkeeping nodes that use expensive models.
func ruleBookkeepingNode(ctx ruleContext) *Suggestion {
	if !isBookkeepingPrompt(ctx.config.Prompt) {
		return nil
	}
	if isCheapModel(ctx.model) {
		return nil
	}
	suggested := suggestCheaperModel(ctx.provider)
	savings := estimateNodeSavings(ctx, suggested)
	return &Suggestion{
		NodeID:       ctx.node.ID,
		Rule:         "bookkeeping-node",
		Message:      "bookkeeping task (summary/cleanup/commit) can use a cheaper model",
		CurrentModel: ctx.model,
		SuggestModel: suggested,
		Savings:      savings,
	}
}

// ruleComplexPromptCheapModel flags complex prompts on cheap models.
func ruleComplexPromptCheapModel(ctx ruleContext) *Suggestion {
	if !isCheapModel(ctx.model) {
		return nil
	}
	score := promptComplexity(ctx.config.Prompt)
	if score < complexityThresholdHigh {
		return nil
	}
	suggested := suggestStrongerModel(ctx.provider)
	return &Suggestion{
		NodeID:       ctx.node.ID,
		Rule:         "complex-prompt-cheap-model",
		Message:      "complex prompt may benefit from a more capable model",
		CurrentModel: ctx.model,
		SuggestModel: suggested,
		Savings:      cost.CostRange{}, // upgrade costs more, no savings
	}
}

// isInRetryLoop checks if a node is targeted by a restart edge.
func isInRetryLoop(nodeID string, w *ir.Workflow) bool {
	for _, e := range w.Edges {
		if e.Restart && e.To == nodeID {
			return true
		}
	}
	return false
}

// estimateNodeSavings estimates cost savings from switching models.
func estimateNodeSavings(ctx ruleContext, suggestedModel string) cost.CostRange {
	currentPrice := lookupPrice(ctx.provider, ctx.model, ctx.pricing)
	suggestedPrice := lookupPrice(ctx.provider, suggestedModel, ctx.pricing)

	ratio := savingsRatio(currentPrice, suggestedPrice)
	nodeCost := estimateNodeCost(ctx)
	return cost.CostRange{
		Min:      nodeCost.Min * ratio,
		Expected: nodeCost.Expected * ratio,
		Max:      nodeCost.Max * ratio,
	}
}

// savingsRatio computes approximate savings fraction from price difference.
func savingsRatio(current, suggested cost.ModelPrice) float64 {
	if current.InputPer1M == 0 {
		return 0
	}
	avgCurrent := (current.InputPer1M + current.OutputPer1M) / 2
	avgSuggested := (suggested.InputPer1M + suggested.OutputPer1M) / 2
	ratio := (avgCurrent - avgSuggested) / avgCurrent
	if ratio < 0 {
		return 0
	}
	return ratio
}

// estimateNodeCost gets the cost range for a node from the precomputed report.
// cost.Analyze writes a NodeCost for every node in the workflow, and ctx.node
// is iterated from that same workflow, so the lookup always succeeds.
func estimateNodeCost(ctx ruleContext) cost.CostRange {
	return ctx.costReport.Nodes[ctx.node.ID].Cost
}

// lookupPrice finds pricing for a provider/model pair.
func lookupPrice(provider, model string, pricing cost.PricingTable) cost.ModelPrice {
	if models, ok := pricing[provider]; ok {
		if p, ok := models[model]; ok {
			return p
		}
	}
	return cost.ModelPrice{}
}
