package pricing

import "sort"

// rankedModel pairs a concrete model id with its in-family rank for resolution.
type rankedModel struct {
	id   string
	rank int
}

// aliasEligible reports whether a model may be the target of a family alias:
// it must be priced, not deprecated, and not a preview. This is what keeps
// "opus@latest" from ever resolving to a retired or unpriced model.
func aliasEligible(p ModelPrice) bool {
	return p.Priced && !p.Deprecated && p.Maturity != "preview"
}

// familyCandidates returns the eligible models of one family within a provider,
// newest first (by Rank, then id for a deterministic tie-break).
func familyCandidates(provider, family string) []rankedModel {
	var c []rankedModel
	for id, p := range index.byProvider[canonicalProvider(provider)] {
		if p.Family == family && aliasEligible(p) {
			c = append(c, rankedModel{id, p.Rank})
		}
	}
	sort.Slice(c, func(i, j int) bool {
		if c[i].rank != c[j].rank {
			return c[i].rank > c[j].rank
		}
		return c[i].id > c[j].id
	})
	return c
}

// ResolveAlias resolves a family reference to a concrete model id. Given
// provider "anthropic", family "opus", and a selector, it returns the model the
// alias points at:
//
//	"latest" / "sota" -> the newest eligible model in the family
//	"stable"          -> the second-newest (one release back), or the newest if
//	                     the family has only one eligible member
//
// It returns ("", false) if the provider/family has no eligible member or the
// selector is not understood. This is the resolver half of the drift-resistant
// model-reference design (#264); the surface syntax and fmt-time pinning that
// call it are a later phase.
func ResolveAlias(provider, family, selector string) (string, bool) {
	c := familyCandidates(provider, family)
	if len(c) == 0 {
		return "", false
	}
	switch selector {
	case "latest", "sota":
		return c[0].id, true
	case "stable":
		return c[min(1, len(c)-1)].id, true
	}
	return "", false
}
