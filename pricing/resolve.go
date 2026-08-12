package pricing

import (
	"regexp"
	"sort"
	"strings"
)

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

// familyMember reports whether the entry stored under key id is a rankable
// member of family: a canonical (non-alias) key, in the family, and eligible.
func familyMember(id string, p ModelPrice, family string) bool {
	return !isAliasKey("", id, p) && p.Family == family && aliasEligible(p)
}

// familyCandidates returns the eligible models of one family within a provider,
// newest first (by Rank, then id for a deterministic tie-break).
func familyCandidates(provider, family string) []rankedModel {
	return rankFamily(index.byProvider[canonicalProvider(provider)], family)
}

// rankFamily is the testable core of familyCandidates. It skips alias keys
// (byProvider holds every model under its canonical id AND each of its aliases,
// so an aliased model would otherwise be counted more than once — which could
// let "stable" resolve to a second spelling of the newest model instead of the
// true one-release-back model).
func rankFamily(models map[string]ModelPrice, family string) []rankedModel {
	var c []rankedModel
	for id, p := range models {
		if familyMember(id, p, family) {
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

// aliasRe matches a family-alias model reference: an optional "provider/"
// prefix, a family tag, and one of the fixed selectors after "@".
var aliasRe = regexp.MustCompile(`^([a-z0-9]+/)?[a-z0-9.\-]+@(latest|sota|stable)$`)

// parseAliasRef splits an alias model value into (provider, family, selector).
// provider is the "provider/" prefix if present, else nodeProvider. The caller
// must have confirmed modelValue is an alias (aliasRe matched).
func parseAliasRef(nodeProvider, modelValue string) (provider, family, selector string) {
	provider = nodeProvider
	ref := modelValue
	if i := strings.IndexByte(ref, '/'); i >= 0 {
		provider = ref[:i]
		ref = ref[i+1:]
	}
	at := strings.IndexByte(ref, '@')
	return provider, ref[:at], ref[at+1:]
}

// ResolveModelRef is the shared entry point for the author-facing family@selector
// alias syntax (#264). Given a node's provider and its model value:
//
//   - If modelValue is a family alias ([provider/]family@selector), isAlias is
//     true. The optional provider/ prefix overrides nodeProvider; provider
//     reports which provider the alias resolved under (the prefix if present,
//     else nodeProvider) so callers can reconcile a cross-provider alias.
//     resolved reports whether the alias points at a concrete, eligible model;
//     concrete is that model id (empty when resolved is false).
//   - Otherwise isAlias is false (concrete="", provider="", resolved=false): an
//     ordinary concrete model id, handled by the existing catalog paths.
func ResolveModelRef(nodeProvider, modelValue string) (concrete, provider string, resolved, isAlias bool) {
	if !aliasRe.MatchString(modelValue) {
		return "", "", false, false
	}
	prov, family, selector := parseAliasRef(nodeProvider, modelValue)
	id, ok := ResolveAlias(prov, family, selector)
	return id, prov, ok, true
}
