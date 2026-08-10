# `pricing` Package + Daily Auto-Sync — Design

**Status:** Draft for approval (build gated on sign-off)
**Date:** 2026-08-07
**Origin:** User design spec (7 points, informed by tracker #518) + the "daily auto-update mechanism" ask.
**Related:** #188 (dotted-ID lookup, shipped v0.51.1), #189 (frontier catalog, shipped v0.52.0), tracker #518 (published-price modeling + provenance + drift test).

---

## Problem

Two problems, one root cause.

1. **Drift by construction.** Model pricing lives in `cost/pricing.go` (prices) and is *mirrored by hand* in `validator/lint_model.go` (the DIP108 known-model catalog). I maintained both by hand in #189 — exactly the drift trap. Separately, tracker maintains its *own* pricing table (#518), so the same numbers live in three places across two repos.
2. **Price churn is coupled to grammar releases.** Prices change weekly; the language changes rarely. Editing a Go literal + cutting a release to update one number is too heavy, and it makes a daily automated refresh awkward.

## Goals

- One source of truth for prices, consumed by dippin's `cost` estimator, dippin's `validator` catalog, **and** tracker's runtime budget path — no cross-repo drift because there is nothing to drift.
- Price edits are reviewable **data diffs**, not code changes — scriptable and scrapable.
- Price churn decoupled from grammar releases.
- A daily mechanism that detects provider price/model changes and lands them with as little manual work as is *safe*.

## Non-goals

- Auto-scraping arbitrary JS-rendered pages and trusting the result blindly (the #189 experience: pages redirect, 404, render client-side, price in CNY). Fetching is assistive, never authoritative.
- Migrating tracker in this repo. This repo ships the leaf `pricing` package designed *so tracker can import it*; tracker's cutover (delete its table, import `dippin-lang/pricing`, move #518's drift test down) is a separate tracker PR.
- Capability metadata (context window, tool/vision/reasoning support, max output). Per the spec, those stay in tracker's `ModelInfo`; only price fields + aliases migrate.

## Architecture

### 1. A leaf `pricing` package — pure data + pure math

New top-level package `pricing/` that imports **nothing** from dippin (not `ir`, not `cost`). Both `cost` and `validator` import it; tracker can too. Because it is a leaf with no analysis deps, `validator` importing it does **not** violate the "packages import `ir`, not each other" rule (it's a shared leaf like the stdlib).

```text
pricing/
  prices.json        # the source of truth (embedded)
  pricing.go         # //go:embed, types, Lookup, Cost
  prices_test.go     # schema/consistency + the relocated drift test
```

### 2. `ModelPrice` — carry prices the way providers publish them

Adopt tracker #518's cache modeling verbatim so the two sides are byte-compatible and tracker's migration is a straight import:

```go
type ModelPrice struct {
    InputPerM, OutputPerM float64
    CachedInputPerM       float64  // OpenAI: absolute cached-input price (0 = use mult)
    CacheReadMult         float64  // Anthropic/Gemini "0.1x" convention (0 = default)
    CacheWriteMult        float64
    Aliases               []string
    Source                string   // published-price URL (provenance)
    AsOf                  string    // date string; scripts can't call time.Now at build
}
```

"Zero means use the other/default" is preserved, so a new entry prices cache traffic sanely without stating both conventions.

**Refinement — known-but-unpriced models (the Qwen/Gemma case).** #189 left Qwen recognized-but-unpriced and Gemma free. The data must distinguish *"in catalog, price not established"* from *"not in catalog"*, so the DIP108 catalog can derive from this data without falsely pricing Qwen at a real number or hiding it. Proposal: an entry may set `"priced": false` (or omit input/output and carry a `Note`), and `Cost` treats an unpriced entry as `0` while `Lookup` still returns `found=true`. This is what lets the validator catalog unify onto `prices.json` (see §5).

### 3. One calc, over a neutral `Usage`

```go
type Usage struct{ Input, Output, CacheRead, CacheWrite, Reasoning int }

func Cost(u Usage, p ModelPrice) float64   // the single implementation both sides call
```

tracker maps its `llm.Usage → pricing.Usage` at runtime; dippin's estimator feeds projected counts. dippin does **not** depend on tracker's `llm.Usage` — the neutral struct is defined here. Zero drift because there is one function.

dippin's existing `computeCostRange` (min/expected/max over turn scenarios) becomes a thin wrapper that calls `pricing.Cost` per scenario.

### 4. Data-as-data via `//go:embed`

`prices.json` is checked in and embedded. Shape per entry:

```json
{
  "provider": "zai",
  "model": "glm-5.2",
  "input_per_m": 1.40,
  "output_per_m": 4.40,
  "aliases": [],
  "source": "https://docs.z.ai/guides/overview/pricing",
  "as_of": "2026-08-07"
}
```

A price edit is a JSON diff — reviewable, scriptable, and decoupled from grammar releases. The `AsOf` per entry replaces the file-level "Last verified" comment and is exactly what the staleness checker and daily job read.

### 5. Total, policy-free `Lookup` — and catalog unification

```go
func Lookup(model string) (ModelPrice, bool)          // alias-resolving, version-separator-insensitive (#188 fold)
func LookupProvider(provider, model string) (ModelPrice, bool)
```

- Alias resolution + the `.`↔`-` version fold from #188 live here (one place, both consumers).
- Unknown → `found=false`; the **caller** sets policy: tracker's budget path → `$0` + warning (never hard-fail); dippin's estimator → louder flag / the DIP108 lint.
- **Catalog unification (kills the #189 drift):** `validator`'s DIP108 known-model check becomes "is this in `pricing`?" via `Lookup`, instead of the hand-mirrored `knownModelProviders` map. A known-but-unpriced entry (Qwen) is `found=true`, so no DIP108, no fabricated price. `ExtraModels` (user-supplied `--extra-models`) still layers on top in the validator.
  - Open question: the validator catalog currently carries a few IDs that may not want price rows (deprecated aliases, invite-only Mythos). These become unpriced-or-priced entries in `prices.json`. Net simpler, but it means `prices.json` is the *model catalog*, not just the *price list*. I think that's correct and is the whole point — flagging it as a deliberate scope call.

### 6. Relocate the drift test

tracker #518's test (catalog constants vs. `PublishedPrice`) moves down into `pricing/prices_test.go`, so the "published price is the source of truth" guarantee travels with the data. In dippin terms this becomes: every entry has a non-empty `Source` and a well-formed `AsOf`; no duplicate (provider, model); aliases don't collide; and the existing `TestLintExamples`/model-catalog assertions keep passing against the derived catalog.

## The daily auto-sync mechanism

**Fetch-source decision (researched 2026-08-07).** No first-party provider exposes price as a public API — model-list APIs (OpenAI/Anthropic/Google/xAI/…) are auth-gated, active-only, and price-free; pricing is HTML (often JS-rendered); deprecation is prose. So the pipeline does **not** scrape provider HTML. Instead it pulls **machine-readable third-party aggregators**, all confirmed live as anonymous JSON:

| Source | Why | Caveat |
|---|---|---|
| **models.dev** `api.json` | primary — per-Mtok in/out + cache, dated metadata, **explicit `status: deprecated/beta`**, covers Qwen/Moonshot/GLM/MiniMax | community-maintained (not authoritative) |
| **LiteLLM** `model_prices_and_context_window.json` | cross-check — hundreds of models + cache costs | per-token; messy IDs; no deprecation flags |
| **OpenRouter** `/api/v1/models` | cross-check — pricing.prompt/completion across proxied providers | only what OpenRouter offers; alias redirects |

The honest split: **detection is fully mechanical; authoritative verification is not** (nobody official publishes price as data). So the pipeline automates up to a reviewable diff, and auto-ships only a narrow, high-confidence subset defined by **cross-source agreement**.

```text
GitHub Action (cron, daily) — no API keys, no HTML scraping
  └─ fetch models.dev + LiteLLM + OpenRouter (JSON)
  └─ normalize IDs (apply the #188 .↔- fold + alias map) and reconcile the 3 sources
  └─ diff the reconciled view against prices.json:
       • new model                    → propose add
       • price changed                → propose update (old→new, per-source values)
       • marked deprecated upstream    → propose status flag (never auto-delete)
       • entry AsOf > N days           → staleness flag
  └─ classify each change by confidence:
       HIGH = ≥2 sources agree on the number (or on the add)
       LOW  = single-source, sources disagree, or a price delta beyond tolerance
  └─ if any diff:
       open a PR with the patch + per-change evidence (which sources, what values)
       CI runs prices_test.go + full suite
       (auto-merge/auto-patch only the HIGH-confidence adds/small-deltas — see policy)
```

Our `prices.json` keeps each entry's `Source` pointing at the **official** provider page (provenance/authority); the aggregators drive *detection*, never final authority. A human reviews anything LOW-confidence, and periodically confirms HIGH-confidence numbers against the official `Source`.

**Auto-release policy (the key decision — recommend PR-gated).**
- **Recommended:** the daily job opens a PR; a human approves; merge triggers the existing release flow. For the *safe subset* — a provider exposing a **stable machine-readable pricing endpoint** (JSON/CSV, not scraped HTML), change limited to **adds or within-tolerance deltas** — allow auto-merge + auto-patch-release. Everything else waits for review.
- **Not recommended:** auto-merge + auto-release of HTML-scraped price changes. A misparse becomes a wrong cost shipped to Homebrew and downstream consumers with no human in the loop. The #189 fetch experience (redirects, 404s, CNY, client-render) is the evidence.

A `just sync-prices` recipe runs the same fetch+diff locally and prints the proposed patch, so the mechanism is usable by hand and in CI from day one.

**Staleness checker** (`just check-prices`, also a CI/pre-commit reminder): flags entries whose `AsOf` is older than N days or whose `Source` is empty — so DeepSeek/Mistral/Cohere (last verified 2026-06-10, never re-checked in #189) surface as due.

## Phasing

- **Phase 1 — the leaf package (this PR).** Create `pricing/` with `prices.json` (migrate every current `cost/pricing.go` entry + provenance), `ModelPrice`/`Usage`/`Cost`/`Lookup`, and `prices_test.go` (schema + drift + no-collision). Repoint `cost` to consume it (thin wrappers; public `cost` API unchanged). Repoint `validator` DIP108 onto `Lookup` (+ keep `ExtraModels`). Delete the hand-mirrored maps. Full sweep of docs referencing the pricing/catalog internals. **No behavior change for users** — `dippin cost`/`lint` output identical; this is purely the data/architecture refactor, verified by the existing tests.
- **Phase 2 — staleness + local sync tool.** `just check-prices`, `just sync-prices` (fetchers for providers with clean sources; assistive diff output). No scheduling yet.
- **Phase 3 — the daily Action.** Cron workflow that runs fetch+diff and opens a PR with evidence; wire the safe-subset auto-merge/auto-release policy agreed in Phase 2 review.

Phase 1 is the foundation and is unambiguous; Phases 2–3 carry the automation-safety decisions and get their own review.

## Open decisions (need sign-off before/with build)

1. **Cross-repo boundary.** This repo builds the leaf `pricing` package + migrates dippin's `cost`/`validator` onto it. tracker's cutover is a separate tracker PR (not done here). Confirm.
2. **Catalog unification.** Fold the validator's DIP108 known-model set into `prices.json` (via `Lookup`), making `prices.json` the single model *catalog*, with an explicit `priced:false`/unpriced representation for Qwen/Gemma. Confirm (vs. keeping the catalog separate and only sharing prices).
3. **`ModelPrice` shape.** Adopt tracker #518's exact struct (cache mults + absolute cached-input + `Aliases`/`Source`/`AsOf`) so tracker imports without translation. Confirm the field set — especially whether `Reasoning` in `Usage` needs its own price field now or later.
4. **Auto-release policy.** PR-gated with a narrow machine-readable-source auto-merge subset (recommended), vs. broader auto-release. This is the safety-critical call.
5. **Package name/location.** Top-level `pricing/` in dippin-lang. Confirm that's the import path both repos want (`github.com/2389-research/dippin-lang/pricing`).
