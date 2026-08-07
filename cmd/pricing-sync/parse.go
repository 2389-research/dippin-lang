package main

import (
	"encoding/json"
	"os"
	"sort"
)

// errOut is the destination for diagnostics; a var so tests can capture it.
var errOut = os.Stderr

// aggregatorProvider maps a models.dev provider id to our canonical provider
// key. Only mapped providers yield candidates; unknown providers are skipped
// (rather than proposing every model under an unrecognized provider as "new").
var aggregatorProvider = map[string]string{
	"anthropic":  "anthropic",
	"openai":     "openai",
	"google":     "gemini",
	"google-ai":  "gemini",
	"xai":        "grok",
	"deepseek":   "deepseek",
	"mistral":    "mistral",
	"cohere":     "cohere",
	"zhipuai":    "zai",
	"zai":        "zai",
	"z-ai":       "zai",
	"moonshotai": "moonshot",
	"moonshot":   "moonshot",
	"minimax":    "minimax",
	"alibaba":    "qwen",
	"qwen":       "qwen",
}

// modelsDevModel is the per-model shape we consume from models.dev's api.json.
// Extra fields are ignored; a schema drift yields fewer/no candidates (safe for
// an assistive, report-only tool) rather than a corrupt catalog.
type modelsDevModel struct {
	ID   string `json:"id"`
	Cost struct {
		Input  float64 `json:"input"`
		Output float64 `json:"output"`
	} `json:"cost"`
	Status string `json:"status"`
}

type modelsDevProvider struct {
	Models map[string]modelsDevModel `json:"models"`
}

// parseModelsDev turns models.dev's api.json into normalized candidates for the
// providers we recognize.
func parseModelsDev(body []byte) ([]candidate, error) {
	var raw map[string]modelsDevProvider
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, err
	}
	var out []candidate
	for aggProv, node := range raw {
		ours, ok := aggregatorProvider[aggProv]
		if !ok {
			continue
		}
		out = append(out, candidatesForProvider(ours, node)...)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Provider != out[j].Provider {
			return out[i].Provider < out[j].Provider
		}
		return out[i].Model < out[j].Model
	})
	return out, nil
}

func candidatesForProvider(provider string, node modelsDevProvider) []candidate {
	var out []candidate
	for id, m := range node.Models {
		modelID := m.ID
		if modelID == "" {
			modelID = id
		}
		out = append(out, candidate{
			Provider:   provider,
			Model:      modelID,
			InputPerM:  m.Cost.Input,
			OutputPerM: m.Cost.Output,
			Deprecated: m.Status == "deprecated",
		})
	}
	return out
}

// printChanges renders the sync report.
func printChanges(changes []change, scanned int) {
	if len(changes) == 0 {
		printfOut("pricing-sync: catalog agrees with models.dev across %d scanned models\n", scanned)
		return
	}
	printfOut("pricing-sync: %d candidate change(s) from models.dev (%d models scanned) — confirm against each official source before editing prices.json:\n", len(changes), scanned)
	for _, c := range changes {
		printfOut("  [%-10s] %-10s %-28s %s\n", c.Kind, c.Provider, c.Model, c.Detail)
	}
}
