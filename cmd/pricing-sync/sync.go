package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"net/http"
	"sort"
	"time"

	"github.com/2389-research/dippin-lang/pricing"
)

// candidate is one model+price as reported by an aggregator, normalized to our
// per-1M-token convention.
type candidate struct {
	Provider   string
	Model      string
	InputPerM  float64
	OutputPerM float64
	Deprecated bool
}

// change is a proposed catalog edit for a human to confirm against the official source.
type change struct {
	Kind     string // "new" | "price" | "deprecated"
	Provider string
	Model    string
	Detail   string
}

// runSync fetches machine-readable aggregators, diffs against the embedded
// catalog, and prints candidate changes. It never writes prices.json — the
// aggregators drive detection, not authority.
func runSync(ctx context.Context, args []string) int {
	fs := flag.NewFlagSet("sync", flag.ContinueOnError)
	tol := fs.Float64("tolerance", 0.0, "ignore price deltas at or below this fraction (e.g. 0.05 = 5%)")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	cands, err := modelsDevFetcher{}.Fetch(ctx)
	if err != nil {
		fmt.Fprintf(errOut, "pricing-sync: fetch failed: %v\n", err)
		return 1
	}
	changes := diff(cands, *tol)
	printChanges(changes, len(cands))
	return 0
}

// diff compares aggregator candidates against the embedded catalog. It is pure
// (no I/O) so it is unit-tested with fixtures. tol suppresses small price
// deltas. Unknown-to-us models are proposed as adds; known models with a
// materially different price are proposed as updates; models the aggregator
// marks deprecated are surfaced.
func diff(cands []candidate, tol float64) []change {
	var out []change
	for _, c := range cands {
		out = appendChange(out, c, tol)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Provider != out[j].Provider {
			return out[i].Provider < out[j].Provider
		}
		return out[i].Model < out[j].Model
	})
	return out
}

func appendChange(out []change, c candidate, tol float64) []change {
	p, found := pricing.LookupProvider(c.Provider, c.Model)
	if !found {
		return append(out, change{Kind: "new", Provider: c.Provider, Model: c.Model,
			Detail: fmt.Sprintf("%.4g/%.4g per MTok (not in catalog)", c.InputPerM, c.OutputPerM)})
	}
	if !p.Priced {
		return out
	}
	if c.Deprecated {
		out = append(out, change{Kind: "deprecated", Provider: c.Provider, Model: c.Model,
			Detail: "aggregator marks deprecated"})
	}
	if priceDiffers(p, c, tol) {
		out = append(out, change{Kind: "price", Provider: c.Provider, Model: c.Model,
			Detail: fmt.Sprintf("catalog %.4g/%.4g → aggregator %.4g/%.4g",
				p.InputPerM, p.OutputPerM, c.InputPerM, c.OutputPerM)})
	}
	return out
}

// priceDiffers reports whether input or output differs by more than tol (a
// fraction of the catalog value; tol=0 means any difference).
func priceDiffers(p pricing.ModelPrice, c candidate, tol float64) bool {
	return exceeds(p.InputPerM, c.InputPerM, tol) || exceeds(p.OutputPerM, c.OutputPerM, tol)
}

func exceeds(have, got, tol float64) bool {
	if have == got {
		return false
	}
	if have == 0 {
		return true
	}
	d := (got - have) / have
	if d < 0 {
		d = -d
	}
	return d > tol
}

// --- models.dev fetcher ---

const modelsDevURL = "https://models.dev/api.json"

type modelsDevFetcher struct{}

func (modelsDevFetcher) Fetch(ctx context.Context) ([]candidate, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, modelsDevURL, nil)
	if err != nil {
		return nil, err
	}
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("models.dev returned %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return parseModelsDev(body)
}
