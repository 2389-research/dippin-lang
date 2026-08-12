---
title: "Models & Pricing"
description: "The supported model providers, the embedded model catalog that backs DIP108 and cost estimation, and how per-token and cached-input pricing is derived."
section_label: "Configuration"
subtitle: "Supported providers, the model catalog, and how cost is priced."
---

## Supported Providers

Dippin recognizes models from eleven providers. Any agent node whose `provider` / `model` pair resolves to a catalog entry is considered known; anything else raises [DIP108](/lint/).

- Anthropic
- OpenAI
- Google/Gemini
- DeepSeek
- xAI/Grok
- Mistral
- Cohere
- Z.AI/GLM
- Moonshot/Kimi
- MiniMax
- Qwen

Several providers are known by more than one name. The catalog carries a small alias map so that, for example, `google` resolves to `gemini`, `xai` to `grok`, and `kimi` to `moonshot`.

## The Model Catalog

The catalog is a **single source of truth**: `pricing/prices.json`. It is embedded into the binary at build time (`//go:embed`) by the leaf `pricing` package, so there is no external file to ship or keep in sync at runtime. Both consumers derive from it — the `cost` estimator and the `validator` (DIP108) — so there is never a second, hand-maintained table to drift out of alignment.

The file has two top-level keys:

- `provider_aliases` — a map of alternate provider spellings to their canonical name.
- `models` — an array of catalog entries.

Each entry describes one model and its published price:

| Field | Meaning |
| --- | --- |
| `provider` | Canonical provider name |
| `model` | Model identifier |
| `input_per_m` | Input price per million tokens |
| `output_per_m` | Output price per million tokens |
| `source` | URL of the official provider page the price was read from |
| `as_of` | Date the price was last verified against that page |
| `cache_read_mult` / `cache_write_mult` | Cache multipliers, when the provider's convention is verified |
| `cached_input_per_m` | Absolute cached-input price per million tokens, when published that way |
| `priced` | `false` marks a known-but-unpriced model |
| `deprecated` | `true` marks a model retired first-party but still billed on passthrough |
| `family` / `rank` / `maturity` | Optional drift-resistance metadata (see [below](#drift-resistance--capability-metadata)) — the family a model belongs to, its ordering within that family (higher `rank` = newer), and `stable`/`preview` |
| `context_window` / `max_output` / `capabilities` | Optional capability metadata — max input/output tokens and capability tags (`tools`, `vision`, `reasoning`) |

Every priced entry carries its `source` URL and an `as_of` date. Prices are verified against official provider documentation before they are committed — never from model training data, which goes stale.

## Model ID Formats

A model identifier may be written with dots or with dashes, and the two forms resolve to the **same** catalog entry. For example, `claude-haiku-4.5` and `claude-haiku-4-5` are equivalent. This means a `.dip` file can use whichever spelling matches a provider's marketing name without tripping DIP108.

## Unknown Models (DIP108)

When an agent node names a `provider` / `model` pair that isn't in the catalog, the linter emits **DIP108 — Unknown Model/Provider**. It is a warning, not an error: the workflow still runs, but an unknown model can't be priced and may be a typo.

To teach the tools about private or in-house models, extend the catalog for a single run with `--extra-models`. The flag is available on `dippin lint` and `dippin doctor` (it tunes the DIP108 catalog check, which `dippin validate` does not run), and takes a `provider:model1,model2;provider2:model3` spec:

```sh
dippin lint --extra-models "custom-corp:custom-llm-v1" pipeline.dip
```

Models added this way suppress DIP108 for the run but are unpriced (they carry no rate), so they contribute nothing to a cost estimate.

## Cache & Cached-Input Pricing

Providers discount tokens served from a prompt cache, and they publish that discount in one of two shapes. The catalog records whichever shape the provider uses:

- **Multipliers** — `cache_read_mult` and `cache_write_mult` express the cached rate as a fraction of the base input rate (for example, a read multiplier of `0.1` means a cache read bills at one tenth of the input rate; a write multiplier above `1` means priming the cache costs more than a plain input token).
- **Absolute** — `cached_input_per_m` records a flat cached-input price per million tokens, used when a provider publishes the cache-hit price directly rather than as a ratio.

A multiplier is only populated once a provider's cache convention has been **verified against official docs**. There is a meaningful difference between a verified `cache_read_mult` of `1` — the provider publishes *no* cached-input discount, so a cache read bills at the full input rate — and an unverified `0`, which prices cache reads at `$0` and signals that a downstream consumer may overlay its own rate until the value is verified. A couple of providers deliberately leave the cache fields unset for now because no official token cache price is available to verify against.

<div class="caveat-card">
  <h4>Verified vs. consumer-overlay</h4>
  <p>An absent or <code>0</code> cache field is not a claim that cache reads are free — it means the rate has not been verified from an official page, and a consumer of the catalog is expected to overlay its own value. A verified full-rate cache read is recorded explicitly as a multiplier of <code>1</code>, distinct from the unverified <code>0</code>.</p>
</div>

## Priced, Unpriced, and Deprecated

Two flags mark entries that behave differently from an ordinary priced model:

- **`priced: false`** — a *known-but-unpriced* model. It is recognized by DIP108 (so it never warns), but it is priced at `$0`. This is used when a provider's USD rate can't be verified from an official page.
- **`deprecated: true`** — a model retired on the first-party provider API but still priced, because it continues to bill on passthrough platforms. Catalog membership therefore does **not** imply first-party callability: a consumer treating the catalog as a first-party allowlist should filter on the deprecated flag, because being in the catalog and being callable first-party are two different things.

## Drift Resistance & Capability Metadata

Beyond price, an entry can carry optional metadata so a consumer can derive its **whole** model catalog from dippin instead of hand-maintaining a parallel one:

- **`family` / `rank` / `maturity`** — the family a model belongs to (e.g. `opus`), its ordering within that family (`rank`, higher = newer), and `stable`/`preview`. This lets a consumer resolve a *family reference* to a concrete model without string-parsing irregular IDs — newest-in-family, one release back, and so on. Resolution deliberately excludes deprecated, unpriced, and preview models, so it can never land on a retired or unpriced model. Go consumers can call `pricing.ResolveAlias(provider, family, selector)` (`latest`/`sota` = newest eligible, `stable` = one release back); a JSON consumer resolves from these fields directly.
- **`context_window` / `max_output` / `capabilities`** — max input context tokens, max output tokens, and capability tags (`tools`, `vision`, `reasoning`). Absent means **unknown**, never a claim of zero. Populated in verified per-provider batches.

### Model families / aliases

An `agent`'s `model:` may be a **family alias** instead of a concrete id:
`[<provider>/]<family>@<selector>`, where the selector is `latest`, `stable`, or
`sota` — for example `opus@latest` (provider taken from the node's `provider:`)
or the self-contained `anthropic/opus@stable`. It lets an author say "the current
top Opus" rather than pinning an id that silently rots when the provider retires a
model. `dippin fmt` **pins** a resolvable alias in place to its concrete catalog
id (author-time, one-way — e.g. `opus@latest` → `claude-opus-5`); after that the
value is a normal id and prices/lints like any other. Resolution excludes
deprecated, preview, and unpriced members, so an alias can never pin to a retired
or unpriced model. An alias that resolves to nothing is left untouched by `fmt`
and flagged **[DIP162](/lint/)**. Aliases resolve only for families that carry
`family`/`rank` metadata (currently the Anthropic families: `opus`, `sonnet`,
`haiku`).

### Cache coverage

Every priced model should carry a cache rate so a consumer never has to guess one. dippin's own tests fail if a new priced model ships without a cache rate (or a documented, reasoned exception), so the set of models lacking a verified rate only shrinks over time. Go consumers can read that set via `pricing.CacheGaps()`.

### Deprecation detection (DIP161)

Because the catalog records `deprecated`, dippin warns when a workflow pins a retired model: **[DIP161](/lint/)** fires when an `agent` names a model that is in the catalog but flagged deprecated — a drift smoke-detector that catches a pipeline still pinned to a model that's been retired first-party (but is still billed on passthrough platforms).

## Estimating Cost

Because pricing lives in the same embedded catalog the validator uses, `dippin cost` can estimate a workflow's per-run token spend directly from the model each agent names — no separate price table to configure. For the cost command, its assumptions, and the related coverage and optimization tools, see [Analysis Tools](/analysis/).

## Use the catalog in your own tools

The full catalog is published as consumable JSON at **[`/prices.json`](/prices.json)** (also aliased as **[`/models.json`](/models.json)** — the catalog carries more than prices) — the exact file dippin embeds, refreshed on every deploy, and served with `Access-Control-Allow-Origin: *` so you can fetch it straight from a browser. A JSON Schema for the shape is published at **[`/models.schema.json`](/models.schema.json)**. Each entry carries `provider`, `model`, `input_per_m` / `output_per_m` (USD per million tokens), optional cache fields (`cache_read_mult` / `cache_write_mult`, or an absolute `cached_input_per_m`), the `priced` / `deprecated` flags, optional drift metadata (`family` / `rank` / `maturity`) and capability metadata (`context_window` / `max_output` / `capabilities`), and provenance (`source` URL + `as_of` date). A top-level `provider_aliases` map resolves shorthand names (e.g. `xai` → `grok`) to canonical providers.

```sh
curl -s https://dippin.org/prices.json | jq '.models[] | select(.provider == "anthropic")'
```

## Report a correction

Every price is verified against the provider's **official** pricing page — each entry records the `source` URL and the `as_of` date it was last checked. Prices are never taken from model memory or third-party aggregators as ground truth, so a correction needs a first-party citation. If you spot a stale or wrong number, open an issue or a PR against [`pricing/prices.json`](https://github.com/2389-research/dippin-lang/blob/main/pricing/prices.json), **including the official source URL and the date you checked**.
