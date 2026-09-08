package main

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// OpenRouter is the second aggregator, cross-referenced against models.dev so
// a stale or wrong entry on one side is caught by the other (see crossCheck).
// It is assistive like models.dev — never authoritative.
const openRouterURL = "https://openrouter.ai/api/v1/models"

// openRouterProvider maps an OpenRouter model-id prefix to our canonical
// provider key. Unmapped prefixes are skipped (not proposed as "new" under an
// unrecognized provider), mirroring aggregatorProvider.
var openRouterProvider = map[string]string{
	"anthropic":  "anthropic",
	"openai":     "openai",
	"google":     "gemini",
	"x-ai":       "grok",
	"z-ai":       "zai",
	"moonshotai": "moonshot",
	"meta":       "meta",
	"meta-llama": "meta",
	"minimax":    "minimax",
	"qwen":       "qwen",
	"mistralai":  "mistral",
	"deepseek":   "deepseek",
	"cohere":     "cohere",
}

// openRouterModel is the per-model shape from OpenRouter's /api/v1/models.
// Pricing is per-token USD *as a string*; extra fields are ignored so a schema
// drift degrades to fewer candidates rather than a corrupt catalog.
type openRouterModel struct {
	ID      string `json:"id"` // "provider/model[:variant]"
	Pricing struct {
		Prompt     string `json:"prompt"`
		Completion string `json:"completion"`
	} `json:"pricing"`
}

type openRouterBody struct {
	Data []openRouterModel `json:"data"`
}

type openRouterFetcher struct{}

func (openRouterFetcher) Fetch(ctx context.Context) ([]candidate, error) {
	body, err := getJSON(ctx, openRouterURL)
	if err != nil {
		return nil, err
	}
	return parseOpenRouter(body)
}

// parseOpenRouter turns OpenRouter's model list into normalized candidates for
// the providers we recognize.
func parseOpenRouter(body []byte) ([]candidate, error) {
	var raw openRouterBody
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, err
	}
	var out []candidate
	for _, m := range raw.Data {
		if c, ok := openRouterCandidate(m); ok {
			out = append(out, c)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Provider != out[j].Provider {
			return out[i].Provider < out[j].Provider
		}
		return out[i].Model < out[j].Model
	})
	return out, nil
}

// openRouterCandidate normalizes one entry. Skipped: unmapped prefixes,
// "~"-prefixed routing variants (OpenRouter's own relay routes), and
// ":free"/":batch"/":name"-suffixed model variants (the catalog prices the
// base model only).
func openRouterCandidate(m openRouterModel) (candidate, bool) {
	prefix, rest, ok := openRouterParts(m.ID)
	if !ok {
		return candidate{}, false
	}
	prov, mapped := openRouterProvider[prefix]
	if !mapped {
		return candidate{}, false
	}
	in, errIn := perMTok(m.Pricing.Prompt)
	out, errOut := perMTok(m.Pricing.Completion)
	if errIn != nil || errOut != nil {
		return candidate{}, false
	}
	return candidate{Provider: prov, Model: rest, InputPerM: in, OutputPerM: out,
		Source: "openrouter"}, true
}

// openRouterParts splits a model id into (provider prefix, model) and rejects
// "~"-prefixed routing variants and ":variant"-suffixed model variants.
func openRouterParts(id string) (prefix, rest string, ok bool) {
	prefix, rest, cut := strings.Cut(id, "/")
	if !cut || strings.HasPrefix(prefix, "~") {
		return "", "", false
	}
	if strings.Contains(rest, ":") {
		return "", "", false
	}
	return prefix, rest, true
}

// perMTok converts an OpenRouter per-token USD price (string) to per-1M-tokens.
func perMTok(perToken string) (float64, error) {
	v, err := strconv.ParseFloat(perToken, 64)
	if err != nil {
		return 0, fmt.Errorf("openrouter price %q: %w", perToken, err)
	}
	return v * 1_000_000, nil
}
