package main

import "testing"

func newChange(provider, model string) change {
	return change{Kind: "new", Provider: provider, Model: model, Agg: "1/5"}
}

func models(cs []change) []string {
	out := make([]string, 0, len(cs))
	for _, c := range cs {
		out = append(out, c.Model)
	}
	return out
}

func has(cs []change, model string) bool {
	for _, c := range cs {
		if c.Model == model {
			return true
		}
	}
	return false
}

// A genuinely new text model on a priced provider is the whole point of the
// report: it must survive every filter.
func TestFilterNew_KeepsGenuineTextModel(t *testing.T) {
	got := filterNew([]change{
		newChange("gemini", "gemini-3.8-flash"),
		newChange("openai", "gpt-5.6"),
		newChange("anthropic", "claude-fable-5-1"),
		newChange("moonshot", "kimi-k2.6"),
	})
	if len(got) != 4 {
		t.Fatalf("kept %v, want all 4 text models", models(got))
	}
}

// Non-"new" kinds are price/deprecation drift, reported by the existing path.
func TestFilterNew_DropsNonNewKinds(t *testing.T) {
	got := filterNew([]change{
		{Kind: "price", Provider: "openai", Model: "gpt-5.3"},
		{Kind: "deprecated", Provider: "openai", Model: "o3-mini"},
		newChange("openai", "gpt-5.6"),
	})
	if len(got) != 1 || got[0].Model != "gpt-5.6" {
		t.Errorf("kept %v, want only gpt-5.6", models(got))
	}
}

// Qwen carries only priced:false entries (console-gated, no verifiable USD
// rate), so proposing 50 more Qwen adds is noise we can never action.
func TestFilterNew_DropsUnpricedProviders(t *testing.T) {
	got := filterNew([]change{
		newChange("qwen", "qwen3.8-max"),
		newChange("gemini", "gemini-3.8-flash"),
	})
	if len(got) != 1 || got[0].Provider != "gemini" {
		t.Errorf("kept %v, want only the gemini entry", models(got))
	}
}

// The catalog prices text generation. Image/tts/embedding/video/asr models bill
// in a different unit and are deliberately out of scope.
func TestFilterNew_DropsNonTextModalities(t *testing.T) {
	got := filterNew([]change{
		newChange("gemini", "gemini-3.1-flash-image"),
		newChange("gemini", "gemini-3.1-flash-tts-preview"),
		newChange("gemini", "gemini-embedding-2"),
		newChange("gemini", "veo-3.1-generate-preview"),
		newChange("gemini", "lyria-3-pro-preview"),
		newChange("openai", "text-embedding-3-large"),
		newChange("openai", "gpt-realtime-2.1"),
		newChange("qwen", "qwen3-asr-flash"),
		newChange("mistral", "voxtral-mini-tts-latest"),
		newChange("gemini", "gemini-3.5-live-translate-preview"),
	})
	if len(got) != 0 {
		t.Errorf("kept %v, want none (all non-text modalities)", models(got))
	}
}

// A vision- or reasoning-capable *text* model still bills text tokens and
// belongs in the catalog — the modality filter must not over-reach.
func TestFilterNew_KeepsVisionCapableTextModels(t *testing.T) {
	got := filterNew([]change{
		newChange("deepseek", "deepseek-v4-flash-vision-exp"),
		newChange("cohere", "command-a-vision-07-2025"),
	})
	if len(got) != 2 {
		t.Errorf("kept %v, want both vision-input text models", models(got))
	}
}

// claude-haiku-4-5-20251001 is the dated snapshot of claude-haiku-4-5, which
// the catalog already carries — covered, not new.
func TestFilterNew_DropsDatedSnapshotOfCatalogedModel(t *testing.T) {
	got := filterNew([]change{
		newChange("anthropic", "claude-haiku-4-5-20251001"),
		newChange("anthropic", "claude-opus-4-5-20251101"),
		newChange("anthropic", "claude-sonnet-4-5-20250929"),
	})
	if len(got) != 0 {
		t.Errorf("kept %v, want none (all dated snapshots of cataloged models)", models(got))
	}
}

// gemini-2.5-computer-use-preview-10-2025 style: a MM-YYYY suffix is the same
// snapshot idea. Only drop when the undated base is actually cataloged.
func TestFilterNew_KeepsDatedSnapshotWhenBaseNotCataloged(t *testing.T) {
	got := filterNew([]change{newChange("cohere", "command-a-plus-05-2026")})
	if len(got) != 1 {
		t.Errorf("kept %v, want the entry (base command-a-plus is not cataloged)", models(got))
	}
}

// The catalog carries gemini-3-flash-preview; the aggregator's bare
// gemini-3-flash would be the same model without the preview marker.
func TestFilterNew_DropsPreviewVariantOfCatalogedModel(t *testing.T) {
	got := filterNew([]change{newChange("gemini", "gemini-3.1-flash-lite-preview")})
	if len(got) != 0 {
		t.Errorf("kept %v, want none (already cataloged)", models(got))
	}
}

// Regression guard against the live catalog: the daily job must surface the
// Gemini flash releases we are currently two versions behind on.
func TestFilterNew_SurfacesGeminiFlashGap(t *testing.T) {
	got := filterNew([]change{
		newChange("gemini", "gemini-3.7-flash"),
		newChange("gemini", "gemini-3.8-flash"),
	})
	if !has(got, "gemini-3.8-flash") || !has(got, "gemini-3.7-flash") {
		t.Errorf("kept %v, want both uncataloged flash releases", models(got))
	}
}

// The aggregator reports 0/0 when it has no USD rate (open-weight or
// console-gated models). There is nothing to verify against an official source,
// so these can never be actioned from this report.
func TestFilterNew_DropsZeroPricedCandidates(t *testing.T) {
	got := filterNew([]change{
		{Kind: "new", Provider: "cohere", Model: "c4ai-aya-expanse-32b", Agg: "0/0"},
		{Kind: "new", Provider: "gemini", Model: "gemma-4-31b-it", Agg: "0/0"},
		{Kind: "new", Provider: "mistral", Model: "labs-devstral-small-2512", Agg: "0/0"},
		newChange("openai", "gpt-5.6"),
	})
	if len(got) != 1 || got[0].Model != "gpt-5.6" {
		t.Errorf("kept %v, want only gpt-5.6", models(got))
	}
}

// models.dev exposes Z.AI under both "zai" and "zhipuai"; both map to our "zai"
// key, so an identical model arrives twice and was reported twice.
func TestDedupe_CollapsesIdenticalRows(t *testing.T) {
	got := dedupe([]change{
		{Kind: "new", Provider: "zai", Model: "glm-5.3", Agg: "1.4/4.4"},
		{Kind: "new", Provider: "zai", Model: "glm-5.3", Agg: "1.4/4.4"},
	})
	if len(got) != 1 {
		t.Errorf("got %d rows, want 1 after collapsing the duplicate alias", len(got))
	}
}

// Conflicting prices for the same model are real signal (the two upstream
// provider entries disagree) and must NOT be collapsed away.
func TestDedupe_KeepsConflictingPrices(t *testing.T) {
	got := dedupe([]change{
		{Kind: "new", Provider: "zai", Model: "glm-5v-turbo", Agg: "5/22"},
		{Kind: "new", Provider: "zai", Model: "glm-5v-turbo", Agg: "1.2/4"},
	})
	if len(got) != 2 {
		t.Errorf("got %d rows, want both conflicting prices preserved", len(got))
	}
}
