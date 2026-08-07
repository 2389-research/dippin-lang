package pricing

import (
	"sort"
	"time"
)

// Staleness flags one catalog entry whose provenance needs attention: its AsOf
// date is older than the allowed age, or its Source URL is empty/unparseable.
type Staleness struct {
	Provider string
	Model    string
	AsOf     string
	AgeDays  int    // days since AsOf; -1 if AsOf is missing/unparseable
	Reason   string // "stale" | "no-source" | "bad-date"
}

// StaleEntries reports every catalog entry that is overdue for re-verification
// as of now, using maxAge as the freshness window. now and maxAge are
// parameters (not time.Now) so the check is deterministic and testable. Alias
// keys are not reported separately — each model is reported once under its
// declared provider/id.
func StaleEntries(now time.Time, maxAge time.Duration) []Staleness {
	out := gatherStale(now, maxAge)
	sort.Slice(out, func(i, j int) bool { return lessStaleness(out[i], out[j]) })
	return out
}

// gatherStale collects the stale entries (canonical ids only), unsorted.
func gatherStale(now time.Time, maxAge time.Duration) []Staleness {
	var out []Staleness
	for provider, models := range index.byProvider {
		out = append(out, staleForProvider(provider, models, now, maxAge)...)
	}
	return out
}

// staleForProvider evaluates one provider's models, skipping alias keys.
func staleForProvider(provider string, models map[string]ModelPrice, now time.Time, maxAge time.Duration) []Staleness {
	var out []Staleness
	for id, p := range models {
		if isAliasKey(provider, id, p) {
			continue
		}
		if s, ok := evaluateStaleness(provider, id, p, now, maxAge); ok {
			out = append(out, s)
		}
	}
	return out
}

// lessStaleness orders staleness reports by provider then model.
func lessStaleness(a, b Staleness) bool {
	if a.Provider != b.Provider {
		return a.Provider < b.Provider
	}
	return a.Model < b.Model
}

// isAliasKey reports whether id is one of p's alias spellings rather than its
// canonical model id, so aliases aren't double-reported.
func isAliasKey(_ string, id string, p ModelPrice) bool {
	for _, a := range p.Aliases {
		if a == id {
			return true
		}
	}
	return false
}

// evaluateStaleness classifies a single entry; ok is false when it is fresh.
func evaluateStaleness(provider, id string, p ModelPrice, now time.Time, maxAge time.Duration) (Staleness, bool) {
	s := Staleness{Provider: provider, Model: id, AsOf: p.AsOf, AgeDays: -1}
	if p.Source == "" {
		s.Reason = "no-source"
		return s, true
	}
	t, err := time.Parse("2006-01-02", p.AsOf)
	if err != nil {
		s.Reason = "bad-date"
		return s, true
	}
	age := now.Sub(t)
	s.AgeDays = int(age.Hours() / 24)
	if age > maxAge {
		s.Reason = "stale"
		return s, true
	}
	return Staleness{}, false
}
