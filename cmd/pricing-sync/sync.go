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
	// Agg is the canonical aggregator-side value used to match a suppression:
	// "%.4g/%.4g" input/output for new/price changes, "deprecated" for a
	// deprecation flag. A suppression only applies while this value is unchanged.
	Agg string
}

// syncOptions is the parsed command line for `sync`.
type syncOptions struct {
	tol           float64
	mode          reportMode
	failOnChanges bool
}

// parseSyncFlags parses and validates the flag set. Split out of runSync to
// keep both functions within the repo's complexity ceiling.
func parseSyncFlags(args []string) (syncOptions, int) {
	fs := flag.NewFlagSet("sync", flag.ContinueOnError)
	tol := fs.Float64("tolerance", 0.0, "ignore price deltas at or below this fraction (e.g. 0.05 = 5%)")
	existingOnly := fs.Bool("existing-only", false, "report only price/deprecation drift for models already in the catalog (drop 'new'); the low-noise daily signal")
	newOnly := fs.Bool("new-only", false, "report only new upstream models, filtered to actionable text-model adds on priced providers; the complement of --existing-only")
	failOnChanges := fs.Bool("fail-on-changes", false, "exit non-zero when any candidate is reported (for CI gating)")
	if err := fs.Parse(args); err != nil {
		return syncOptions{}, 2
	}
	if *existingOnly && *newOnly {
		fmt.Fprintln(errOut, "pricing-sync: --existing-only and --new-only are mutually exclusive")
		return syncOptions{}, 2
	}
	return syncOptions{tol: *tol, mode: selectMode(*existingOnly, *newOnly), failOnChanges: *failOnChanges}, 0
}

// runSync fetches machine-readable aggregators, diffs against the embedded
// catalog, and prints candidate changes. It never writes prices.json — the
// aggregators drive detection, not authority.
func runSync(ctx context.Context, args []string) int {
	opts, code := parseSyncFlags(args)
	if code != 0 {
		return code
	}
	sups, err := loadSuppressions()
	if err != nil {
		fmt.Fprintf(errOut, "pricing-sync: bad drift_suppressions.json: %v\n", err)
		return 1
	}
	cands, err := modelsDevFetcher{}.Fetch(ctx)
	if err != nil {
		fmt.Fprintf(errOut, "pricing-sync: fetch failed: %v\n", err)
		return 1
	}
	return reportChanges(cands, opts.tol, opts.mode, opts.failOnChanges, sups, time.Now())
}

// reportMode picks which slice of the diff a run reports. The daily Action runs
// both halves separately so each gets its own section in the drift issue.
type reportMode int

const (
	modeAll reportMode = iota
	modeExistingOnly
	modeNewOnly
)

func selectMode(existingOnly, newOnly bool) reportMode {
	switch {
	case existingOnly:
		return modeExistingOnly
	case newOnly:
		return modeNewOnly
	default:
		return modeAll
	}
}

// reportChanges diffs, filters (drop-new + suppress-list), prints, and returns
// the exit code. Suppressed candidates are dispositioned drift the daily Action
// should not re-open an issue for (see drift_suppressions.json).
func reportChanges(cands []candidate, tol float64, mode reportMode, failOnChanges bool, sups []suppression, now time.Time) int {
	changes := applyMode(diff(cands, tol), mode)
	changes, suppressed := applySuppressions(changes, sups, now)
	printChanges(changes, len(cands))
	if suppressed > 0 {
		printfOut("pricing-sync: %d dispositioned candidate(s) suppressed via drift_suppressions.json\n", suppressed)
	}
	if failOnChanges && len(changes) > 0 {
		return 1
	}
	return 0
}

func applyMode(changes []change, mode reportMode) []change {
	switch mode {
	case modeExistingOnly:
		return dropNew(changes)
	case modeNewOnly:
		return filterNew(changes)
	default:
		return changes
	}
}

// dropNew filters out "new" (upstream-only) candidates, leaving price and
// deprecation drift for models already in our catalog — the actionable,
// low-noise signal (models.dev lists hundreds of image/tts/embedding models we
// deliberately don't price).
func dropNew(changes []change) []change {
	out := changes[:0]
	for _, c := range changes {
		if c.Kind != "new" {
			out = append(out, c)
		}
	}
	return out
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
	agg := fmt.Sprintf("%.4g/%.4g", c.InputPerM, c.OutputPerM)
	p, found := pricing.LookupProvider(c.Provider, c.Model)
	if !found {
		return append(out, change{Kind: "new", Provider: c.Provider, Model: c.Model,
			Detail: fmt.Sprintf("%s per MTok (not in catalog)", agg), Agg: agg})
	}
	if !p.Priced {
		return out
	}
	if c.Deprecated {
		out = append(out, change{Kind: "deprecated", Provider: c.Provider, Model: c.Model,
			Detail: "aggregator marks deprecated", Agg: "deprecated"})
	}
	if priceDiffers(p, c, tol) {
		out = append(out, change{Kind: "price", Provider: c.Provider, Model: c.Model,
			Detail: fmt.Sprintf("catalog %.4g/%.4g → aggregator %s",
				p.InputPerM, p.OutputPerM, agg), Agg: agg})
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
