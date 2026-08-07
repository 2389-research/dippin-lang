# Integrating the `pricing` package (for tracker and other consumers)

`github.com/2389-research/dippin-lang/pricing` is the single source of truth for
LLM model prices and the model catalog. It is a **leaf package** — it imports
nothing from the rest of dippin — so a consumer can depend on it without pulling
in the parser, validator, or analysis machinery.

This guide is written for **tracker** (which today maintains its own pricing
table, tracker #518) but applies to any consumer.

## Version to pin

- **Floor:** `v0.53.0` — the release that introduced the `pricing` package.
- **Recommended:** the latest tag (currently **`v0.54.0`**), which adds
  `pricing.StaleEntries` and the sync tooling. Nothing in `v0.54.0` changes the
  `Lookup`/`Cost` API a consumer uses.

```sh
go get github.com/2389-research/dippin-lang@v0.54.0
```

Consumers **pin a specific tag** — never `@latest` — per dippin's release model.

## The API

```go
import "github.com/2389-research/dippin-lang/pricing"

type ModelPrice struct {
    InputPerM, OutputPerM float64 // USD per 1M tokens
    CachedInputPerM       float64 // OpenAI-style absolute cached-input price (0 = use mult)
    CacheReadMult         float64 // Anthropic/Gemini "0.1x" convention (0 = default)
    CacheWriteMult        float64
    Aliases               []string
    Priced                bool    // false = in catalog but no established price (e.g. Qwen)
    Source, AsOf          string  // provenance: published-price URL + verification date
}

type Usage struct{ Input, Output, CacheRead, CacheWrite, Reasoning int }

func Cost(u Usage, p ModelPrice) float64            // the one cost calc — call this
func Lookup(model string) (ModelPrice, bool)         // alias- and version-fold-resolving
func LookupProvider(provider, model string) (ModelPrice, bool)
```

- **Version-separator fold:** `Lookup`/`LookupProvider` treat `claude-haiku-4.5`
  and `claude-haiku-4-5` as the same model (issue #188).
- **Provider aliases:** `LookupProvider` resolves `google→gemini`, `xai→grok`,
  `kimi→moonshot`.
- **Policy is the caller's.** `found=false` means unknown — you decide what that
  means (tracker's budget path: treat as `$0` + warning, never a hard fail). A
  `found=true` with `Priced=false` means recognized-but-unpriced (e.g. Qwen):
  price it as `$0`, but don't warn that the model is unknown.

## How tracker should integrate

1. **Delete tracker's own price table.** Replace every price lookup with
   `pricing.LookupProvider(provider, model)`.

2. **Map `llm.Usage` → `pricing.Usage` and call `pricing.Cost`.** Don't
   reimplement the arithmetic — there is one implementation, so there is nothing
   to drift.

   ```go
   func costUSD(provider, model string, u llm.Usage) (float64, bool) {
       p, ok := pricing.LookupProvider(provider, model)
       if !ok {
           return 0, false // caller policy: $0 + warning, never a hard fail
       }
       return pricing.Cost(pricing.Usage{
           Input:      u.PromptTokens,
           Output:     u.CompletionTokens,
           CacheRead:  u.CacheReadTokens,
           CacheWrite: u.CacheWriteTokens,
           Reasoning:  u.ReasoningTokens,
       }, p), true
   }
   ```

3. **Keep capability metadata in tracker.** Context window, tool/vision/reasoning
   support, and max-output are runtime concerns — they stay in tracker's
   `ModelInfo`. Only the money fields + aliases come from `pricing`. Have
   `ModelInfo` call `pricing.LookupProvider` for the price part.

4. **Move the #518 drift test down.** The "published price is the source of
   truth" guarantee now lives in `pricing/prices_test.go` (every entry has a
   `Source` + well-formed `AsOf`, no duplicate keys). Delete tracker's copy.

## Important caveat: cache pricing is not yet populated

`ModelPrice` carries the cache fields (matching tracker #518's shape), **but
dippin's `prices.json` does not populate them yet** — dippin's own estimator
models base input/output only. So today `CachedInputPerM`, `CacheReadMult`, and
`CacheWriteMult` are all `0`, and `pricing.Cost` therefore computes **cache
traffic at $0**.

If tracker prices cache reads/writes, you have two choices until dippin's
`prices.json` carries cache rates:

- **Keep tracker's cache-rate table** for the cache terms only, and use
  `pricing` for base input/output; or
- **Accept base-only costing** for now.

Populating cache rates in `prices.json` (and teaching the sync tool to pull them
from the aggregators) is planned — track it so the day it lands, tracker can
drop its remaining cache table.

## Keeping prices current

dippin owns the freshness process (you consume a pinned tag and adopt new
releases):

- `just check-prices` — flags catalog entries overdue for re-verification.
- `just sync-prices` — diffs `prices.json` against models.dev and reports
  candidates (report-only; a human confirms against each official `source`).
- A daily GitHub Action opens a rolling issue when it detects price/deprecation
  drift for existing models.

When dippin cuts a new tag with refreshed prices, tracker bumps its pin to adopt
it — the same flow tracker already uses for every other dippin feature.
