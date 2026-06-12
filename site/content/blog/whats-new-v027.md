---
title: "What's New in Dippin v0.27"
date: "2026-05-18"
description: "Model catalog refresh — 11+ new IDs across six providers, seven price corrections, and a retirement calendar worth pinning somewhere."
tagStyle: "release"
tagLabel: "RELEASE"
category: "Releases"
readTime: "4 min read"
related:
  - url: "whats-new-v026.html"
    title: "What's New in v0.26"
    summary: "The workflow `requires:` keyword — declare environmental dependencies so runtimes can preflight."
  - url: "whats-new-v025.html"
    title: "What's New in v0.25"
    summary: ".dipx format v1.1 — real ctx cancellation, an inspect that actually inspects, and exit code 2 that matches the spec."
---

The LLM market moves. Since the last verification pass roughly a month ago, two providers shipped new flagships, one cut prices fleet-wide, another retired its previous-generation IDs and started silently redirecting calls, and a fifth doubled the price of three older models. `v0.27.0` is the catalog refresh that catches up.

New IDs are now recognized by `dippin lint`. New prices flow through `dippin cost`. Stale IDs come out so authors get a DIP108 nudge instead of routing requests at a model that doesn't quite exist anymore.

## What's new, by provider

**OpenAI** shipped the GPT-5.5 line plus a slate of -pro variants:

| Model | Input | Output |
|---|---|---|
| `gpt-5.5` | $5.00 | $30.00 |
| `gpt-5.5-pro` | $30.00 | $180.00 |
| `gpt-5.4-pro` | $30.00 | $180.00 |
| `gpt-5.2-pro` | $21.00 | $168.00 |
| `gpt-5` | $1.25 | $10.00 |
| `gpt-5-pro` | $15.00 | $120.00 |
| `gpt-5-mini` | $0.25 | $2.00 |
| `gpt-5-nano` | $0.05 | $0.40 |

**xAI** shipped Grok 4.3 — `grok-4.3` at $1.25 / $2.50. **DeepSeek** rolled out V4 as `deepseek-v4-flash` ($0.14 / $0.28) and `deepseek-v4-pro` ($1.74 / $3.48 list; in a 75% launch discount through May 31). **Mistral** added `mistral-medium-3-5-2604` ($1.50 / $7.50, new flagship-class), `mistral-medium-3-1-2508` ($0.40 / $2.00), and the Ministral 3 generation (`ministral-3-3b-2512`, `ministral-3-8b-2512`, `ministral-3-14b-2512`). **Google** promoted `gemini-3.1-flash-lite` to GA. **Cohere** has new Command-A variants in the docs (reasoning, vision, translate) but no public per-token pricing yet — they're not in the catalog until prices surface.

## The price moves

The single most consequential change: **OpenAI doubled the price of every legacy GPT model that still has a price**. `gpt-5.2`, `gpt-5.1`, and `gpt-4.1-mini` all went up 2x on both input and output. The newer `-mini`/`-nano` tiers and the o-series held steady. If you had workflows pinned to those three models, your `dippin cost` estimates were off by 2x until you bumped to this version.

**xAI went the other direction.** The entire `grok-4.20-*-0309` family came down from $2/$6 to $1.25/$2.50, matching grok-4.3's tier. Three models, fleet-wide cut.

**DeepSeek's V4-Flash is half the price of the legacy aliases on input** ($0.14 vs $0.28) and a third less on output ($0.28 vs $0.42). The `deepseek-chat` and `deepseek-reasoner` aliases now resolve to V4-Flash and bill at V4-Flash rates — `dippin cost` reflects that.

**Anthropic repriced Haiku 3.5** to $0.80 / $4.00. The model was retired on the first-party API in February but remains available via Bedrock and Vertex AI; the new rate is what Anthropic now publishes on the pricing page.

## Retirement calendar

Pin this somewhere.

**Already retired (removed from the catalog):**

| Date | Model |
|---|---|
| 2026-04-30 | `mistral-small-3.2` |
| 2026-03-09 | `gemini-3-pro-preview` |
| 2026-02-27 | `pixtral-large` |

**Retired but still functionally callable** (kept in the catalog, priced at the rate they're actually billed):

| Date | Model |
|---|---|
| 2026-05-15 | `grok-4-1-fast-reasoning`, `grok-4-1-fast-non-reasoning` — xAI silently redirects to `grok-4.3` and bills at grok-4.3 rates ($1.25/$2.50). Requests still execute, so DIP108 would be misleading. Migrate when convenient. |

**Coming up (still in catalog with deprecation comments):**

| Date | Model |
|---|---|
| **2026-06-01** | `gemini-2.0-flash` |
| 2026-06-15 | `claude-sonnet-4-0`, `claude-opus-4-0` |
| 2026-07-24 | `deepseek-chat`, `deepseek-reasoner` aliases (sunset to V4-Flash) |
| 2026-10-23 | `gpt-4o`, `gpt-4.1-nano`, `o3-mini`, `o4-mini` |

The 2026-06-01 Gemini date is the urgent one — that's two weeks out. Workflows still pinned to `gemini-2.0-flash` should plan to migrate to `gemini-2.5-flash-lite` or `gemini-3.1-flash-lite`.

A note on the two categories. Hard-retired models (provider returns an error) come out of the catalog so DIP108 surfaces — calling them is a real bug. Soft-retired models that the provider silently redirects (like xAI's `grok-4-1-fast-*` family) stay in, priced at the rate the redirect target actually bills. Treating them like typos would be wrong twice over: the request still executes, and cost analysis needs the real billed price, not `$0`. A future, more specific "deprecated alias — prefer X" diagnostic can replace the current behavior of staying silent on these.

## Documented uncertainties

A few values we couldn't verify cleanly from canonical sources this pass, held at prior values pending re-check:

- **Mistral `nemo` and `mistral-small-2603`** — Mistral's pricing tab is JS-rendered and didn't yield to our doc fetcher. Third-party scrapers disagree on the prices. Values held at $0.02/$0.04 and $0.10/$0.30 respectively until manual verification.
- **Cohere `command-a-03-2025` and `command-r7b-12-2024`** — Cohere removed per-token pricing for these from the public page. Held at $2.50/$10.00 and $0.0375/$0.15.
- **Tiered pricing not modeled** — Gemini Pro models charge 2x for prompts >200K tokens; OpenAI's GPT-5.5 and GPT-5.4 family charge 2x input / 1.5x output for prompts >272K tokens. `dippin cost` models the base tier only.

All flagged with inline code comments for the next pass.

## Cohere ID drift

One quieter change worth noting. Cohere's documentation now lists `command-r-08-2024`, `command-r-plus-08-2024`, and `command-r7b-12-2024` as the canonical IDs; the bare `command-r`/`command-r-plus`/`command-r7b` aliases resolve to versions that were *deprecated* 2025-09-15. The catalog now accepts both the dated and bare forms (so existing `.dip` files keep validating), but the dated IDs are recommended. Bare aliases carry deprecation comments.

## The escape hatch

If you need a model we don't catalog yet — provider-specific betas, fine-tuned variants, a hosted gateway with custom IDs — pass them per lint run via scoped options (the CLI exposes this as `--extra-models`):

```go
extra := validator.ParseExtraModels("openai:my-custom-model;anthropic:claude-internal-2026")
result := validator.LintWithOptions(w, validator.Options{ExtraModels: extra})
```

The catalog is scoped to that single call — nothing mutates package-level state, so extra models never leak across lint runs. Unknown `provider:model` combos produce a DIP108 warning, not a hard error, so workflows still run — but adding the ID makes the warning go away and gets you accurate cost estimation.

## Methodology

Sources for every provider are pinned in the `validator/lint_model.go` and `cost/pricing.go` source-file comments. We verify against canonical provider docs only — never third-party aggregators, never LLM training data. (Training data is always stale by definition; the rule is in the project's CLAUDE.md.) Verification cadence is roughly monthly. If you spot a model on a provider's official docs that we don't recognize, file an issue with the canonical URL.

Full changelog: [CHANGELOG.md](https://github.com/2389-research/dippin-lang/blob/main/CHANGELOG.md).
