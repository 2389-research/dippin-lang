package main

import (
	"regexp"
	"strings"

	"github.com/2389-research/dippin-lang/pricing"
)

// nonTextMarker matches model-ID substrings for modalities the catalog does not
// price. The catalog prices *text* generation per MTok; image, speech, video,
// music and embedding models bill in a different unit, so proposing them as
// catalog adds is pure noise.
//
// Deliberately absent: "vision", "vl", "omni". Those are text models with a
// non-text *input* — they still bill text tokens and belong in the catalog.
var nonTextMarker = []string{
	"image", "-tts", "tts-", "embedding", "embed-", "-embed",
	"realtime", "-live", "live-", "asr", "-ocr", "voxtral",
	"veo-", "lyria-", "whisper", "-audio", "audio-", "speech",
	"moderation", "rerank",
}

// snapshotSuffix strips a trailing dated-snapshot marker so a model can be
// compared against its undated catalog entry: "-20251001" (YYYYMMDD),
// "-05-2026" (MM-YYYY), or a trailing "-preview"/"-latest" marker.
var snapshotSuffix = regexp.MustCompile(`(?:-\d{8}|-\d{2}-\d{4}|-preview|-latest)$`)

// filterNew reduces raw "new" candidates to the actionable signal: text models
// on providers we actually price, that are not just a redundant naming variant
// of an entry the catalog already carries.
//
// Without this, models.dev's ~170 upstream-only models drown the daily report
// and the job goes unread — which is how the catalog fell two Gemini Flash
// releases behind while the Action reported "success" every morning.
func filterNew(changes []change) []change {
	unpriced := unpricedProviders()
	var out []change
	for _, c := range changes {
		if c.Kind == "new" && actionableNew(c, unpriced) {
			out = append(out, c)
		}
	}
	return dedupe(out)
}

// dedupe collapses rows identical in provider+model+aggregator value. models.dev
// exposes some providers under several ids that map to one canonical key of
// ours (zai/zhipuai/z-ai), which otherwise reports the same model twice.
// Rows that agree on the model but disagree on price are kept — that
// disagreement is signal, not duplication.
func dedupe(changes []change) []change {
	seen := map[change]bool{}
	var out []change
	for _, c := range changes {
		if seen[c] {
			continue
		}
		seen[c] = true
		out = append(out, c)
	}
	return out
}

func actionableNew(c change, unpriced map[string]bool) bool {
	if unpriced[c.Provider] || c.Agg == "0/0" {
		return false
	}
	return isTextModel(c.Model) && !variantOfCataloged(c.Provider, c.Model)
}

// unpricedProviders lists providers whose every catalog entry is priced:false
// (e.g. Qwen, whose USD rates are console-gated and unverifiable). Derived from
// the catalog rather than hardcoded, so it self-corrects once one is priced.
func unpricedProviders() map[string]bool {
	out := map[string]bool{}
	for provider, models := range pricing.Providers() {
		if !anyPriced(models) {
			out[provider] = true
		}
	}
	return out
}

func anyPriced(models map[string]pricing.ModelPrice) bool {
	for _, p := range models {
		if p.Priced {
			return true
		}
	}
	return false
}

func isTextModel(model string) bool {
	id := strings.ToLower(model)
	for _, marker := range nonTextMarker {
		if strings.Contains(id, marker) {
			return false
		}
	}
	return true
}

// variantOfCataloged reports whether the model is a dated snapshot or
// preview/latest alias of an entry the catalog already carries — covered, not
// new (e.g. claude-haiku-4-5-20251001 vs. our claude-haiku-4-5).
func variantOfCataloged(provider, model string) bool {
	for base := model; ; {
		trimmed := snapshotSuffix.ReplaceAllString(base, "")
		if trimmed == base {
			return false
		}
		if _, found := pricing.LookupProvider(provider, trimmed); found {
			return true
		}
		base = trimmed
	}
}
