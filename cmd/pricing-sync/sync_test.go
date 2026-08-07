package main

import "testing"

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
