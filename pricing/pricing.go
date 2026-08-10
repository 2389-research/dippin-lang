// Package pricing is the single source of truth for LLM model pricing and the
// model catalog. It is a leaf package: it imports nothing from the rest of
// dippin, so cost, validator, and downstream consumers (e.g. tracker) can all
// depend on it without pulling in analysis machinery.
//
// The data lives in prices.json (embedded below), not in Go literals, so a
// price change is a reviewable data diff decoupled from grammar releases. Each
// entry records its published-price Source URL and the AsOf date it was
// verified — provenance travels with the number.
package pricing

import (
	_ "embed"
	"sort"
	"strings"
)

//go:embed prices.json
var pricesJSON []byte

// ModelPrice is the price of one model, carried the way providers publish it.
// Cache fields follow the "zero means use the other / default" convention so a
// new entry prices cache traffic sanely without stating both. Prices are USD
// per 1M tokens.
type ModelPrice struct {
	InputPerM       float64
	OutputPerM      float64
	CachedInputPerM float64 // OpenAI-style absolute cached-input price (0 = use CacheReadMult)
	CacheReadMult   float64 // Anthropic/Gemini "0.1x" convention (0 = provider default)
	CacheWriteMult  float64
	Aliases         []string
	Priced          bool   // false = in the catalog but no established price (e.g. Qwen, free tiers)
	Deprecated      bool   // true = retired on the first-party provider API (still priced for passthrough platforms like Bedrock/Vertex, so kept in the catalog); a consumer treating the catalog as a first-party allowlist should filter these out
	Source          string // published-price URL
	AsOf            string // YYYY-MM-DD the price was verified against Source
}

// Usage is a neutral token-count struct so the one Cost implementation serves
// both dippin's estimator (projected counts) and a runtime (observed counts)
// without either depending on the other's usage type.
type Usage struct {
	Input, Output, CacheRead, CacheWrite, Reasoning int
}

// Cost returns the USD cost of one Usage under one ModelPrice. It is the single
// cost calculation both dippin and downstream consumers call — there is nothing
// to drift because there is one implementation.
func Cost(u Usage, p ModelPrice) float64 {
	perM := func(n int, rate float64) float64 { return float64(n) * rate / 1_000_000 }
	c := perM(u.Input, p.InputPerM) + perM(u.Output, p.OutputPerM)
	c += perM(u.Reasoning, p.OutputPerM) // reasoning tokens bill as output
	c += cachedInputCost(u.CacheRead, p)
	if p.CacheWriteMult > 0 {
		c += perM(u.CacheWrite, p.InputPerM*p.CacheWriteMult)
	}
	return c
}

// cachedInputCost prices cache-read traffic: an absolute cached-input rate wins,
// else the read multiplier against the input rate, else zero.
func cachedInputCost(cacheRead int, p ModelPrice) float64 {
	switch {
	case p.CachedInputPerM > 0:
		return float64(cacheRead) * p.CachedInputPerM / 1_000_000
	case p.CacheReadMult > 0:
		return float64(cacheRead) * p.InputPerM * p.CacheReadMult / 1_000_000
	default:
		return 0
	}
}

// CanonicalModelID folds the version separator so dotted and dashed spellings
// compare equal (claude-haiku-4.5 == claude-haiku-4-5). See issue #188.
func CanonicalModelID(model string) string {
	return strings.ReplaceAll(model, ".", "-")
}

// Lookup resolves a model ID across all providers, honoring aliases and the
// version-separator fold. found is false for a model not in the catalog; the
// caller sets policy for unknowns (a runtime treats unknown as $0 + warning; a
// linter can flag it louder). A found-but-unpriced entry returns Priced=false.
func Lookup(model string) (ModelPrice, bool) {
	if p, ok := index.byModel[model]; ok {
		return p, true
	}
	if p, ok := index.byCanonModel[CanonicalModelID(model)]; ok {
		return p, true
	}
	return ModelPrice{}, false
}

// LookupProvider resolves (provider, model), applying provider aliases
// (google→gemini, xai→grok, kimi→moonshot) and the version-separator fold.
func LookupProvider(provider, model string) (ModelPrice, bool) {
	prov := canonicalProvider(provider)
	models, ok := index.byProvider[prov]
	if !ok {
		return ModelPrice{}, false
	}
	if p, ok := models[model]; ok {
		return p, true
	}
	want := CanonicalModelID(model)
	for id, p := range models {
		if CanonicalModelID(id) == want {
			return p, true
		}
	}
	return ModelPrice{}, false
}

// canonicalProvider resolves a provider alias to its canonical key.
func canonicalProvider(provider string) string {
	if c, ok := index.providerAliases[provider]; ok {
		return c
	}
	return provider
}

// Providers returns the canonical provider→model→price catalog (priced entries
// only). Callers that need alias provider keys (google/xai/kimi) should also
// consult ProviderAliases.
func Providers() map[string]map[string]ModelPrice {
	out := make(map[string]map[string]ModelPrice, len(index.byProvider))
	for prov, models := range index.byProvider {
		m := make(map[string]ModelPrice, len(models))
		for id, p := range models {
			if p.Priced {
				m[id] = p
			}
		}
		out[prov] = m
	}
	return out
}

// ProviderAliases returns alias→canonical provider mappings.
func ProviderAliases() map[string]string {
	out := make(map[string]string, len(index.providerAliases))
	for a, c := range index.providerAliases {
		out[a] = c
	}
	return out
}

// KnownProvider reports whether a provider (alias or canonical) is in the catalog.
func KnownProvider(provider string) bool {
	_, ok := index.byProvider[canonicalProvider(provider)]
	return ok
}

// ProviderNames returns every provider key — canonical and alias — sorted.
// Used for the "known providers" diagnostic help.
func ProviderNames() []string {
	seen := map[string]bool{}
	for p := range index.byProvider {
		seen[p] = true
	}
	for a := range index.providerAliases {
		seen[a] = true
	}
	return sortedKeys(seen)
}

// ModelIDs returns every model ID (priced and unpriced) known for a provider
// (alias-resolved), sorted. Used for the "known models for <provider>" help.
func ModelIDs(provider string) []string {
	models, ok := index.byProvider[canonicalProvider(provider)]
	if !ok {
		return nil
	}
	seen := map[string]bool{}
	for id := range models {
		seen[id] = true
	}
	return sortedKeys(seen)
}

func sortedKeys(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
