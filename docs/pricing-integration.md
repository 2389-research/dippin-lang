# Integrating the `pricing` package (for tracker and other consumers)

`github.com/2389-research/dippin-lang/pricing` is the single source of truth for
LLM model prices and the model catalog. It is a **leaf package** — it imports
nothing from the rest of dippin — so a consumer can depend on it without pulling
in the parser, validator, or analysis machinery.

This guide is written for **tracker** (which today maintains its own pricing
table, tracker #518) but applies to any consumer.

## Version to pin

- **Floor:** `v0.53.0` — the release that introduced the `pricing` package.
- **Recommended:** the latest tag (currently **`v0.64.0`**). Everything since the
  floor is additive to the catalog data and the `ModelPrice` shape; the
  `Lookup`/`Cost` API a consumer calls is unchanged. Notable additions: cache
  read/write rates (`v0.57.0`+, see below), `ModelPrice.Deprecated` (`v0.59.0`),
  per-family OpenAI cache multipliers (`v0.59.1`), tail-provider cache rates
  (`v0.61.0`) and Mistral/Cohere verified-no-discount (`v0.62.1`), and the drift
  suppress-list tooling.

```sh
go get github.com/2389-research/dippin-lang@v0.64.0
```

Consumers **pin a specific tag** — never `@latest` — per dippin's release model.

## The API

```go
import "github.com/2389-research/dippin-lang/pricing"

type ModelPrice struct {
    InputPerM, OutputPerM float64 // USD per 1M tokens
    CachedInputPerM       float64 // OpenAI-style absolute cached-input price (0 = use mult)
    CacheReadMult         float64 // cache-read/input ratio (0 = unverified; 1 = verified no discount)
    CacheWriteMult        float64
    Aliases               []string
    Priced                bool    // false = in catalog but no established price (e.g. Qwen)
    Deprecated            bool    // true = retired on the first-party provider API, still priced for Bedrock/Vertex passthrough (#224)
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

## Cache pricing

Cache read rates **are** populated (as of `v0.57.0`+), so `pricing.Cost` bills
cache traffic for most providers directly — no consumer overlay needed. Read the
cache fields per entry rather than assuming a global convention:

- **Populated with a verified rate:** Anthropic (`0.1×` read / `1.25×` write),
  OpenAI (per-family — gpt-4o `0.5×`, gpt-4.1 `0.25×`, GPT-5 `0.1×`), Gemini
  (`0.1×` read), and DeepSeek, Z.AI/GLM, xAI/Grok, Moonshot/Kimi (absolute
  `CachedInputPerM` from their published cache-hit prices).
- **Verified no discount:** Mistral and Cohere carry `CacheReadMult: 1` — their
  official pages publish no cached-input rate, so a cache read bills at the full
  input rate. `1` is deliberately distinct from `0`.
- **Still unverified (`0`):** MiniMax (official page is audio-only) and Qwen
  (unpriced/console-gated). For these two only, `Cost` prices cache reads at `$0`,
  so a consumer that bills their cache traffic should overlay its own default
  until dippin verifies them.

So the rule for a consumer: **overlay a default cache rate only where both
`CacheReadMult` and `CachedInputPerM` are `0`** (i.e. MiniMax/Qwen today) — every
other provider is authoritative in the catalog.

## Keeping prices current

dippin owns the freshness process (you consume a pinned tag and adopt new
releases):

- `just check-prices` — flags catalog entries overdue for re-verification.
- `just sync-prices` — diffs `prices.json` against models.dev and reports
  candidates (report-only; a human confirms against each official `source`).
- `just new-prices` — reports only upstream models missing from the catalog,
  filtered to actionable text-model adds on priced providers.
- A daily GitHub Action opens a rolling issue when it detects price/deprecation
  drift for existing models **or** new upstream models the catalog is missing.
  Persistent non-candidates (models we deliberately don't carry) are silenced
  via `cmd/pricing-sync/drift_suppressions.json`.

When dippin cuts a new tag with refreshed prices, tracker bumps its pin to adopt
it — the same flow tracker already uses for every other dippin feature.
