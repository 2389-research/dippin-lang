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

// TestGPT52CodexAlias covers #209: gpt-5.2-codex bills at the gpt-5.2 rate
// (confirmed 2026-08-10 by LiteLLM + OpenRouter; consistent with gpt-5.3-codex
// pricing at its base rate on the official page), modeled as an alias.
func TestGPT52CodexAlias(t *testing.T) {
	base, ok := LookupProvider("openai", "gpt-5.2")
	if !ok {
		t.Fatal("gpt-5.2 must be priced")
	}
	codex, ok := LookupProvider("openai", "gpt-5.2-codex")
	if !ok {
		t.Fatal("gpt-5.2-codex must resolve (issue #209 budget escape)")
	}
	if codex.InputPerM != base.InputPerM || codex.OutputPerM != base.OutputPerM {
		t.Errorf("gpt-5.2-codex %v/%v != gpt-5.2 %v/%v",
			codex.InputPerM, codex.OutputPerM, base.InputPerM, base.OutputPerM)
	}
}

// TestCacheRatesVerifiedProviders covers #210: Anthropic/OpenAI/Gemini carry
// verified cache read multipliers (0.1x), Anthropic also a 1.25x write; Cost
// honors them. Unverified providers stay at 0 (consumer overlay).
func TestCacheRatesVerifiedProviders(t *testing.T) {
	opus, _ := Lookup("claude-opus-5")
	if opus.CacheReadMult != 0.1 || opus.CacheWriteMult != 1.25 {
		t.Errorf("claude-opus-5 cache mults = %v/%v, want 0.1/1.25", opus.CacheReadMult, opus.CacheWriteMult)
	}
	// cache read = 0.1 x input: 1M cache-read tokens at $5 input → $0.50.
	if got := Cost(Usage{CacheRead: 1_000_000}, opus); got != 0.50 {
		t.Errorf("opus cache-read cost = %v, want 0.50", got)
	}
	gpt, _ := Lookup("gpt-5.2")
	if gpt.CacheReadMult != 0.1 || gpt.CacheWriteMult != 0 {
		t.Errorf("gpt-5.2 cache mults = %v/%v, want 0.1/0", gpt.CacheReadMult, gpt.CacheWriteMult)
	}
	// A provider we did NOT verify keeps zero cache rates.
	glm, _ := Lookup("glm-5.2")
	if glm.CacheReadMult != 0 {
		t.Errorf("glm-5.2 (unverified) cache read = %v, want 0", glm.CacheReadMult)
	}
}

// TestDeprecatedModelsMarked covers #224: models retired on the first-party
// provider API are still priced (they bill on Bedrock/Vertex passthrough) but
// carry Deprecated=true so a consumer treating the catalog as a first-party
// allowlist can filter them; current models are Deprecated=false.
func TestDeprecatedModelsMarked(t *testing.T) {
	retired := []string{"claude-opus-4-1", "claude-opus-4-0", "claude-sonnet-4-0", "claude-haiku-3-5"}
	for _, m := range retired {
		p, ok := Lookup(m)
		if !ok {
			t.Errorf("%s should still be in the catalog (priced for passthrough)", m)
			continue
		}
		if !p.Deprecated {
			t.Errorf("%s should be Deprecated=true (retired on the first-party API)", m)
		}
		if p.InputPerM <= 0 {
			t.Errorf("%s should still carry a price (Bedrock/Vertex passthrough)", m)
		}
	}
	// A current model is not deprecated.
	if p, _ := Lookup("claude-opus-5"); p.Deprecated {
		t.Error("claude-opus-5 must not be Deprecated")
	}
}
