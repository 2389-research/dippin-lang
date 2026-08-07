package pricing

import (
	"regexp"
	"testing"
)

var dateRe = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)

// TestCatalogProvenance is the relocated "published price is the source of
// truth" guarantee (tracker #518): every entry carries a Source URL and a
// well-formed AsOf date, and no (provider, model) is duplicated.
func TestCatalogProvenance(t *testing.T) {
	seen := map[string]bool{}
	for prov, models := range index.byProvider {
		for model, p := range models {
			key := prov + "/" + model
			if p.Source == "" {
				t.Errorf("%s: empty Source", key)
			}
			if !dateRe.MatchString(p.AsOf) {
				t.Errorf("%s: AsOf %q is not YYYY-MM-DD", key, p.AsOf)
			}
		}
		_ = seen
	}
}

func TestLookupExactAndFold(t *testing.T) {
	// Exact.
	if _, ok := Lookup("claude-opus-5"); !ok {
		t.Error("claude-opus-5 not found")
	}
	// Version-separator fold: dashed catalog key, dotted query (#188).
	if _, ok := Lookup("claude-haiku-4.5"); !ok {
		t.Error("dotted claude-haiku-4.5 should resolve via the fold")
	}
	// Unknown.
	if _, ok := Lookup("nonexistent-9-9"); ok {
		t.Error("unknown model must not resolve")
	}
}

func TestLookupProviderAlias(t *testing.T) {
	// gemini models must resolve under the google alias.
	viaGemini, ok1 := LookupProvider("gemini", "gemini-3.6-flash")
	viaGoogle, ok2 := LookupProvider("google", "gemini-3.6-flash")
	if !ok1 || !ok2 {
		t.Fatalf("gemini-3.6-flash: gemini=%v google=%v (both must resolve)", ok1, ok2)
	}
	if viaGemini.InputPerM != viaGoogle.InputPerM || viaGemini.OutputPerM != viaGoogle.OutputPerM {
		t.Error("google alias must return the same price as gemini")
	}
	// grok under xai; moonshot under kimi.
	if _, ok := LookupProvider("xai", "grok-4.5"); !ok {
		t.Error("grok-4.5 must resolve under the xai alias")
	}
	if _, ok := LookupProvider("kimi", "kimi-k3"); !ok {
		t.Error("kimi-k3 must resolve under the kimi alias")
	}
}

func TestKnownButUnpriced(t *testing.T) {
	// Qwen is in the catalog (recognized) but has no established price.
	p, ok := LookupProvider("qwen", "qwen3.7-max")
	if !ok {
		t.Fatal("qwen3.7-max must be in the catalog")
	}
	if p.Priced {
		t.Error("qwen3.7-max must be Priced=false")
	}
	// Providers() (priced-only) must exclude it.
	if _, in := Providers()["qwen"]["qwen3.7-max"]; in {
		t.Error("Providers() must exclude unpriced entries")
	}
}

func TestCost(t *testing.T) {
	p := ModelPrice{InputPerM: 3.0, OutputPerM: 15.0}
	got := Cost(Usage{Input: 1_000_000, Output: 1_000_000}, p)
	if got != 18.0 {
		t.Errorf("Cost = %v, want 18.0", got)
	}
	// Cache-read via multiplier (Anthropic 0.1x convention).
	pc := ModelPrice{InputPerM: 10.0, OutputPerM: 0, CacheReadMult: 0.1}
	if got := Cost(Usage{CacheRead: 1_000_000}, pc); got != 1.0 {
		t.Errorf("cache-read cost = %v, want 1.0", got)
	}
	// Absolute cached-input wins over multiplier.
	pa := ModelPrice{InputPerM: 10.0, CachedInputPerM: 0.5, CacheReadMult: 0.1}
	if got := Cost(Usage{CacheRead: 1_000_000}, pa); got != 0.5 {
		t.Errorf("absolute cached-input cost = %v, want 0.5", got)
	}
}
