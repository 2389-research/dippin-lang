package main

import (
	_ "embed"
	"encoding/json"
	"strings"
	"time"
)

//go:embed drift_suppressions.json
var suppressionsJSON []byte

// suppression records a drift candidate a human has already dispositioned so the
// daily sync stops re-flagging it — until its review_by date passes, or until the
// aggregator's reported value changes from the one dispositioned here (either
// re-surfaces the candidate for fresh review). Keyed on provider/model/kind plus
// the canonical aggregator value (see change.Agg).
//
// Aggregator may contain `*` wildcards (each matches any run of characters) for
// a value that churns without meaning anything — OpenRouter's reseller rates
// move daily, so an exact value re-surfaces the same dispositioned row every
// morning. Keep the stable part literal: "md:1.4/4.4 or:*" still re-surfaces
// the moment models.dev (closer to the official price) changes. Source, when
// set, restricts the suppression to rows reported by that aggregator
// ("openrouter", "models.dev" or "both"), so an "openrouter"-scoped wildcard
// never hides a models.dev price change for the same model.
type suppression struct {
	Provider   string `json:"provider"`
	Model      string `json:"model"`
	Kind       string `json:"kind"` // "price" | "deprecated" | "new" | "disagree"
	Aggregator string `json:"aggregator_value"`
	Source     string `json:"source,omitempty"`
	Reason     string `json:"reason"`
	ReviewBy   string `json:"review_by"` // YYYY-MM-DD
}

// loadSuppressions parses the embedded checked-in suppress-list.
func loadSuppressions() ([]suppression, error) {
	var out []suppression
	if err := json.Unmarshal(suppressionsJSON, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// active reports whether the suppression is still in effect at now (before its
// review_by date). A missing or malformed review_by is treated as expired, so a
// bad entry re-surfaces the candidate rather than hiding it indefinitely.
func (s suppression) active(now time.Time) bool {
	t, err := time.Parse("2006-01-02", s.ReviewBy)
	if err != nil {
		return false
	}
	return now.Before(t)
}

// matches reports whether the suppression covers this change: same identity,
// same reporting aggregator (when Source is set), and an aggregator value
// matching the dispositioned one (exactly, or per its `*` wildcards). A changed
// value outside the pattern won't match, so the candidate re-surfaces.
func (s suppression) matches(c change) bool {
	if !s.sameCandidate(c) {
		return false
	}
	return wildcardMatch(s.Aggregator, c.Agg)
}

// sameCandidate reports whether the suppression names this change's model and
// kind, and (when Source is set) the aggregator that reported it.
func (s suppression) sameCandidate(c change) bool {
	if s.Provider != c.Provider || s.Model != c.Model || s.Kind != c.Kind {
		return false
	}
	return s.Source == "" || s.Source == c.Source
}

// wildcardMatch reports whether value matches pattern, where each `*` in the
// pattern matches any run of characters (including none) and everything else
// is literal. Unlike path.Match, `*` crosses the `/` in "in/out" values.
func wildcardMatch(pattern, value string) bool {
	parts := strings.Split(pattern, "*")
	if len(parts) == 1 {
		return pattern == value
	}
	if !strings.HasPrefix(value, parts[0]) {
		return false
	}
	rest := value[len(parts[0]):]
	for _, mid := range parts[1 : len(parts)-1] {
		i := strings.Index(rest, mid)
		if i < 0 {
			return false
		}
		rest = rest[i+len(mid):]
	}
	return strings.HasSuffix(rest, parts[len(parts)-1])
}

// applySuppressions drops changes covered by an active suppression, returning the
// survivors and the count suppressed.
func applySuppressions(changes []change, sups []suppression, now time.Time) ([]change, int) {
	kept := changes[:0]
	suppressed := 0
	for _, c := range changes {
		if suppressedBy(c, sups, now) {
			suppressed++
			continue
		}
		kept = append(kept, c)
	}
	return kept, suppressed
}

func suppressedBy(c change, sups []suppression, now time.Time) bool {
	for _, s := range sups {
		if s.active(now) && s.matches(c) {
			return true
		}
	}
	return false
}
