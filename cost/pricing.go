package cost

import "github.com/2389-research/dippin-lang/pricing"

// DefaultPricing returns a PricingTable with current model prices (USD per 1M
// tokens), projected from the embedded pricing catalog
// (github.com/2389-research/dippin-lang/pricing), which is the single source of
// truth for prices and their provenance. This adapter keeps cost's historical
// provider→model→{in,out} shape — including the alias provider keys
// (google/xai/kimi) — so every existing caller keeps working unchanged while
// the data lives in one place (pricing/prices.json).
//
// Only priced entries are projected; a known-but-unpriced model (e.g. Qwen) is
// absent here and prices at $0, exactly as an out-of-table model did before.
func DefaultPricing() PricingTable {
	catalog := pricing.Providers() // canonical providers, priced-only
	table := make(PricingTable, len(catalog)+len(pricing.ProviderAliases()))
	for provider, models := range catalog {
		table[provider] = projectModels(models)
	}
	for alias, canonical := range pricing.ProviderAliases() {
		if models, ok := catalog[canonical]; ok {
			table[alias] = projectModels(models)
		}
	}
	return table
}

// projectModels maps the pricing catalog's ModelPrice to cost's leaner
// {InputPer1M, OutputPer1M} form (dippin's estimator models base input/output
// only; cache tiers live in the pricing layer for runtime consumers).
func projectModels(models map[string]pricing.ModelPrice) map[string]ModelPrice {
	out := make(map[string]ModelPrice, len(models))
	for id, p := range models {
		out[id] = ModelPrice{InputPer1M: p.InputPerM, OutputPer1M: p.OutputPerM}
	}
	return out
}
