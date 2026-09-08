package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/2389-research/dippin-lang/pricing"
)

// candidate is one model+price as reported by an aggregator, normalized to our
// per-1M-token convention. Source names the aggregator ("models.dev" /
// "openrouter").
type candidate struct {
	Provider   string
	Model      string
	InputPerM  float64
	OutputPerM float64
	Deprecated bool
	Source     string
}

// change is a proposed catalog edit for a human to confirm against the official source.
type change struct {
	Kind     string // "new" | "price" | "deprecated" | "disagree"
	Provider string
	Model    string
	Detail   string
	Source   string // which aggregator reported it; "both" when both agree
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
	md, or, err, orErr := fetchSources(ctx)
	if err != nil {
		fmt.Fprintf(errOut, "pricing-sync: fetch failed: %v\n", err)
		return 1
	}
	if orErr != nil {
		fmt.Fprintf(errOut, "pricing-sync: openrouter fetch failed: %v (cross-check disabled)\n", orErr)
		or = nil
	}
	changes := sortChanges(append(diff(append(md, or...), opts.tol), crossCheck(md, or, opts.tol)...))
	return reportChanges(changes, len(md), len(or), opts.mode, opts.failOnChanges, sups, time.Now())
}

// fetchSources pulls both aggregators in parallel. A models.dev failure is the
// fatal one (it is the primary source); an OpenRouter failure degrades to a
// warning so the catalog diff still runs.
func fetchSources(ctx context.Context) (md, or []candidate, err, orErr error) {
	var wg sync.WaitGroup
	wg.Add(2)
	go func() { defer wg.Done(); md, err = modelsDevFetcher{}.Fetch(ctx) }()
	go func() { defer wg.Done(); or, orErr = openRouterFetcher{}.Fetch(ctx) }()
	wg.Wait()
	return
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

// reportChanges filters the pre-computed change list (mode + suppress-list),
// prints it, and returns the exit code. Suppressed candidates are dispositioned
// drift the daily Action should not re-open an issue for (see
// drift_suppressions.json).
func reportChanges(changes []change, scannedMD, scannedOR int, mode reportMode, failOnChanges bool, sups []suppression, now time.Time) int {
	changes = applyMode(changes, mode)
	changes, suppressed := applySuppressions(changes, sups, now)
	printChanges(changes, scannedMD, scannedOR)
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
		return dedupe(dropNew(changes))
	case modeNewOnly:
		return filterNew(changes)
	default:
		return dedupe(changes)
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
	return sortChanges(out)
}

func sortChanges(out []change) []change {
	sort.Slice(out, func(i, j int) bool {
		if out[i].Provider != out[j].Provider {
			return out[i].Provider < out[j].Provider
		}
		return out[i].Model < out[j].Model
	})
	return out
}

// foldKey canonicalizes a provider or model id for cross-source matching:
// OpenRouter lowercases ids and uses dots where the catalog and models.dev use
// dashes ("anthropic/claude-fable-5.1" vs. claude-fable-5-1).
func foldKey(s string) string {
	return strings.ToLower(pricing.CanonicalModelID(s))
}

// catalogMatch resolves an aggregator (provider, model) to the catalog entry,
// tolerating case and dot/dash differences (OpenRouter lowercases ids and dots
// where the catalog uses dashes). Returns the matched catalog model id so
// change rows stay keyed on the catalog's spelling (suppressions key on it).
func catalogMatch(provider, model string) (pricing.ModelPrice, string, bool) {
	if p, ok := pricing.LookupProvider(provider, model); ok {
		return p, catalogSpelling(provider, foldKey(model), model), true
	}
	return catalogMatchFolded(provider, model)
}

// catalogMatchFolded is the last resort for aggregators that spell ids with
// different case than the catalog (OpenRouter lowercases).
func catalogMatchFolded(provider, model string) (pricing.ModelPrice, string, bool) {
	want := foldKey(model)
	for _, id := range pricing.ModelIDs(provider) {
		if foldKey(id) == want {
			p, _ := pricing.LookupProvider(provider, id)
			return p, id, true
		}
	}
	return pricing.ModelPrice{}, "", false
}

// catalogSpelling returns the catalog's own spelling for a folded id, so rows
// and suppressions key on the catalog id rather than the aggregator's variant.
func catalogSpelling(provider, folded, fallback string) string {
	for _, id := range pricing.ModelIDs(provider) {
		if foldKey(id) == folded {
			return id
		}
	}
	return fallback
}

func appendChange(out []change, c candidate, tol float64) []change {
	agg := fmt.Sprintf("%.4g/%.4g", c.InputPerM, c.OutputPerM)
	p, catID, found := catalogMatch(c.Provider, c.Model)
	if !found {
		return append(out, change{Kind: "new", Provider: c.Provider, Model: c.Model,
			Detail: fmt.Sprintf("%s per MTok (not in catalog)", agg), Agg: agg, Source: c.Source})
	}
	id := displayID(c.Model, catID)
	if !p.Priced {
		return out
	}
	if c.Deprecated {
		out = append(out, change{Kind: "deprecated", Provider: c.Provider, Model: id,
			Detail: "aggregator marks deprecated", Agg: "deprecated", Source: c.Source})
	}
	if priceDiffers(p, c, tol) {
		out = append(out, change{Kind: "price", Provider: c.Provider, Model: id,
			Detail: fmt.Sprintf("catalog %.4g/%.4g → aggregator %s",
				p.InputPerM, p.OutputPerM, agg), Agg: agg, Source: c.Source})
	}
	return out
}

// displayID prefers the catalog's spelling of a matched model (change rows and
// suppressions key on it); unmatched models keep the aggregator's id.
func displayID(model, catID string) string {
	if catID != "" {
		return catID
	}
	return model
}

// crossCheck compares the two aggregators with each other: a model both list
// at materially different prices is a "disagree" change — at least one
// aggregator is stale or wrong, so a human must check the official source
// before trusting either number. Scoped to models with a priced catalog
// entry: for models we don't price, the source consensus is already visible
// on the paired "new" rows ([both] vs. two disagreeing rows), and no
// catalog number needs defending.
func crossCheck(md, or []candidate, tol float64) []change {
	fromMD := crossIndex(md)
	var out []change
	for _, o := range or {
		m, ok := fromMD[crossKey(o)]
		if !ok {
			continue
		}
		p, catID, found := catalogMatch(o.Provider, o.Model)
		if !driftable(found, p) || !sourcesDisagree(m, o, tol) {
			continue
		}
		out = append(out, disagreeChange(m, o, catID))
	}
	return out
}

func crossIndex(md []candidate) map[string]candidate {
	out := map[string]candidate{}
	for _, c := range md {
		out[crossKey(c)] = c
	}
	return out
}

// driftable reports whether the model has a priced catalog entry (the unit a
// disagree row defends; unpriced entries like Qwen's cannot drift).
func driftable(found bool, p pricing.ModelPrice) bool {
	return found && p.Priced
}

func crossKey(c candidate) string { return foldKey(c.Provider) + "/" + foldKey(c.Model) }

// sourcesDisagree reports whether the two aggregators differ by more than tol
// on either side (symmetric; priceDiffers is anchored to the catalog value).
func sourcesDisagree(a, b candidate, tol float64) bool {
	return exceeds(a.InputPerM, b.InputPerM, tol) || exceeds(b.InputPerM, a.InputPerM, tol) ||
		exceeds(a.OutputPerM, b.OutputPerM, tol) || exceeds(b.OutputPerM, a.OutputPerM, tol)
}

func disagreeChange(m, o candidate, model string) change {
	return change{Kind: "disagree", Provider: o.Provider, Model: model, Source: "both",
		Detail: fmt.Sprintf("models.dev %.4g/%.4g vs openrouter %.4g/%.4g per MTok",
			m.InputPerM, m.OutputPerM, o.InputPerM, o.OutputPerM),
		Agg: fmt.Sprintf("md:%.4g/%.4g or:%.4g/%.4g", m.InputPerM, m.OutputPerM, o.InputPerM, o.OutputPerM)}
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
	body, err := getJSON(ctx, modelsDevURL)
	if err != nil {
		return nil, err
	}
	return parseModelsDev(body)
}

// getJSON GETs a URL with a 30s timeout and returns the body, erroring on any
// non-200. Shared by both aggregator fetchers.
func getJSON(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
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
		return nil, fmt.Errorf("%s returned %d", url, resp.StatusCode)
	}
	return io.ReadAll(resp.Body)
}
