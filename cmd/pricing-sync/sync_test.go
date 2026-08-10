package main

import (
	"testing"
	"time"
)

var testNow = time.Date(2026, 8, 10, 0, 0, 0, 0, time.UTC)

func TestApplySuppressions_FiltersActiveDispositioned(t *testing.T) {
	changes := []change{
		{Kind: "price", Provider: "anthropic", Model: "claude-sonnet-5", Agg: "2/10"},
		{Kind: "deprecated", Provider: "openai", Model: "o3-mini", Agg: "deprecated"},
		{Kind: "price", Provider: "openai", Model: "gpt-9-real", Agg: "5/20"}, // not suppressed
	}
	sups := []suppression{
		{Provider: "anthropic", Model: "claude-sonnet-5", Kind: "price", Aggregator: "2/10", ReviewBy: "2026-11-08"},
		{Provider: "openai", Model: "o3-mini", Kind: "deprecated", Aggregator: "deprecated", ReviewBy: "2026-11-08"},
	}
	kept, n := applySuppressions(changes, sups, testNow)
	if n != 2 {
		t.Fatalf("suppressed = %d, want 2", n)
	}
	if len(kept) != 1 || kept[0].Model != "gpt-9-real" {
		t.Errorf("kept = %+v, want only gpt-9-real", kept)
	}
}

func TestApplySuppressions_ReSurfacesOnChangedAggValue(t *testing.T) {
	// The aggregator now reports 1/8 instead of the dispositioned 2/10 — the
	// disposition no longer applies, so it must re-surface.
	changes := []change{{Kind: "price", Provider: "anthropic", Model: "claude-sonnet-5", Agg: "1/8"}}
	sups := []suppression{{Provider: "anthropic", Model: "claude-sonnet-5", Kind: "price", Aggregator: "2/10", ReviewBy: "2026-11-08"}}
	kept, n := applySuppressions(changes, sups, testNow)
	if n != 0 || len(kept) != 1 {
		t.Errorf("changed aggregator value must re-surface; kept=%+v suppressed=%d", kept, n)
	}
}

func TestApplySuppressions_ExpiredReviewByReSurfaces(t *testing.T) {
	changes := []change{{Kind: "price", Provider: "anthropic", Model: "claude-sonnet-5", Agg: "2/10"}}
	sups := []suppression{{Provider: "anthropic", Model: "claude-sonnet-5", Kind: "price", Aggregator: "2/10", ReviewBy: "2026-01-01"}}
	kept, n := applySuppressions(changes, sups, testNow)
	if n != 0 || len(kept) != 1 {
		t.Errorf("expired suppression (past review_by) must re-surface; kept=%+v suppressed=%d", kept, n)
	}
}

// TestEmbeddedSuppressionsWellFormed guards the checked-in drift_suppressions.json:
// it parses, every entry has the required fields, a valid kind, and a parseable
// review_by date (a malformed date silently disables a suppression, so fail loud).
func TestEmbeddedSuppressionsWellFormed(t *testing.T) {
	sups, err := loadSuppressions()
	if err != nil {
		t.Fatalf("embedded drift_suppressions.json does not parse: %v", err)
	}
	validKind := map[string]bool{"price": true, "deprecated": true, "new": true}
	for _, s := range sups {
		if s.Provider == "" || s.Model == "" || s.Aggregator == "" || s.Reason == "" {
			t.Errorf("incomplete suppression: %+v", s)
		}
		if !validKind[s.Kind] {
			t.Errorf("%s/%s: invalid kind %q", s.Provider, s.Model, s.Kind)
		}
		if _, err := time.Parse("2006-01-02", s.ReviewBy); err != nil {
			t.Errorf("%s/%s: review_by %q is not YYYY-MM-DD", s.Provider, s.Model, s.ReviewBy)
		}
	}
}

// findChange returns the change for a model, or nil.
func findChange(changes []change, model string) *change {
	for i := range changes {
		if changes[i].Model == model {
			return &changes[i]
		}
	}
	return nil
}

func TestDiffClassifiesChanges(t *testing.T) {
	cands := []candidate{
		// Known + same price as the catalog → no change.
		{Provider: "anthropic", Model: "claude-opus-5", InputPerM: 5, OutputPerM: 25},
		// Known + different price → "price".
		{Provider: "anthropic", Model: "claude-sonnet-5", InputPerM: 4, OutputPerM: 20},
		// Not in catalog → "new".
		{Provider: "openai", Model: "gpt-6-imaginary", InputPerM: 1, OutputPerM: 2},
		// Known + aggregator says deprecated → "deprecated".
		{Provider: "anthropic", Model: "claude-opus-5", InputPerM: 5, OutputPerM: 25, Deprecated: true},
	}
	changes := diff(cands, 0)

	if c := findChange(changes, "gpt-6-imaginary"); c == nil || c.Kind != "new" {
		t.Errorf("gpt-6-imaginary should be a 'new' change, got %+v", c)
	}
	if c := findChange(changes, "claude-sonnet-5"); c == nil || c.Kind != "price" {
		t.Errorf("claude-sonnet-5 price delta should be a 'price' change, got %+v", c)
	}
	// The same-price opus-5 candidate must not produce a price change; the
	// deprecated one must produce a 'deprecated' change.
	var sawDeprecated bool
	for _, c := range changes {
		if c.Model == "claude-opus-5" && c.Kind == "price" {
			t.Error("claude-opus-5 at catalog price must not yield a price change")
		}
		if c.Model == "claude-opus-5" && c.Kind == "deprecated" {
			sawDeprecated = true
		}
	}
	if !sawDeprecated {
		t.Error("expected a 'deprecated' change for the deprecated opus-5 candidate")
	}
}

func TestDiffToleranceSuppressesSmallDeltas(t *testing.T) {
	// sonnet-5 catalog is 3/15; a 4% output bump under a 5% tolerance is ignored.
	cands := []candidate{{Provider: "anthropic", Model: "claude-sonnet-5", InputPerM: 3, OutputPerM: 15.6}}
	if got := diff(cands, 0.05); len(got) != 0 {
		t.Errorf("4%% delta under 5%% tolerance should be suppressed, got %+v", got)
	}
	if got := diff(cands, 0.0); len(got) != 1 {
		t.Errorf("with zero tolerance the delta should surface, got %+v", got)
	}
}

func TestParseModelsDevMapsProvidersAndSkipsUnknown(t *testing.T) {
	body := []byte(`{
	  "anthropic": {"models": {"claude-opus-5": {"id":"claude-opus-5","cost":{"input":5,"output":25}}}},
	  "google":    {"models": {"gemini-3.6-flash": {"id":"gemini-3.6-flash","cost":{"input":1.5,"output":7.5}}}},
	  "some-unknown-provider": {"models": {"whatever": {"id":"whatever","cost":{"input":1,"output":1}}}}
	}`)
	cands, err := parseModelsDev(body)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	var haveGemini, haveUnknown bool
	for _, c := range cands {
		if c.Provider == "gemini" && c.Model == "gemini-3.6-flash" {
			haveGemini = true // google → gemini mapping
		}
		if c.Model == "whatever" {
			haveUnknown = true
		}
	}
	if !haveGemini {
		t.Error("google provider must map to gemini")
	}
	if haveUnknown {
		t.Error("unknown provider must be skipped")
	}
}

func TestDropNewKeepsOnlyExistingModelDrift(t *testing.T) {
	changes := []change{
		{Kind: "new", Provider: "openai", Model: "gpt-6-imaginary"},
		{Kind: "price", Provider: "anthropic", Model: "claude-sonnet-5"},
		{Kind: "deprecated", Provider: "gemini", Model: "gemini-2.0-flash"},
	}
	got := dropNew(changes)
	if len(got) != 2 {
		t.Fatalf("dropNew kept %d, want 2 (price+deprecated)", len(got))
	}
	for _, c := range got {
		if c.Kind == "new" {
			t.Errorf("dropNew leaked a 'new' change: %+v", c)
		}
	}
}
