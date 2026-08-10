package main

import (
	_ "embed"
	"encoding/json"
	"time"
)

//go:embed drift_suppressions.json
var suppressionsJSON []byte

// suppression records a drift candidate a human has already dispositioned so the
// daily sync stops re-flagging it — until its review_by date passes, or until the
// aggregator's reported value changes from the one dispositioned here (either
// re-surfaces the candidate for fresh review). Keyed on provider/model/kind plus
// the canonical aggregator value (see change.Agg).
type suppression struct {
	Provider   string `json:"provider"`
	Model      string `json:"model"`
	Kind       string `json:"kind"` // "price" | "deprecated" | "new"
	Aggregator string `json:"aggregator_value"`
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

// matches reports whether the suppression covers this change: same identity and
// the same aggregator value that was dispositioned. A changed aggregator value
// won't match, so the candidate re-surfaces.
func (s suppression) matches(c change) bool {
	return s.Provider == c.Provider && s.Model == c.Model &&
		s.Kind == c.Kind && s.Aggregator == c.Agg
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
