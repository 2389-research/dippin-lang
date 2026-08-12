package pricing

import "sort"

// noCacheSignal reports that an entry carries no cache-rate information at all —
// a consumer estimating cost would have to guess one.
func noCacheSignal(p ModelPrice) bool {
	return p.CachedInputPerM == 0 && p.CacheReadMult == 0 && p.CacheWriteMult == 0
}

// cacheGapEntry is a priced, non-deprecated model with no cache signal.
func cacheGapEntry(p ModelPrice) bool {
	return p.Priced && !p.Deprecated && noCacheSignal(p)
}

// CacheGaps returns "provider/model" for every priced, non-deprecated catalog
// entry that carries no cache-rate signal (CachedInputPerM, CacheReadMult, and
// CacheWriteMult all zero). Sorted for determinism.
//
// This lets a downstream consumer know exactly which models still need a
// cache-rate overlay instead of hard-coding a guess: the goal is to drive this
// list down to only models whose providers publish no cache price at all, at
// which point the consumer can retire its overlay for everything else. The
// package's own test asserts this set never grows silently, so a newly added
// priced model can't ship without either a cache rate or a deliberate entry on
// the known-unverifiable allowlist.
func CacheGaps() []string {
	var out []string
	for prov, models := range index.byProvider {
		for id, p := range models {
			// Skip alias keys so a model isn't listed under both its id and an
			// alias spelling (byProvider holds it under every name).
			if !isAliasKey("", id, p) && cacheGapEntry(p) {
				out = append(out, prov+"/"+id)
			}
		}
	}
	sort.Strings(out)
	return out
}
