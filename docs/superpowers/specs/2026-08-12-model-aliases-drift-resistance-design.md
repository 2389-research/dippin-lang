# Drift-Resistant Model References (Model Aliases) — Design

**Status:** Draft for approval (research + design only; no implementation)
**Date:** 2026-08-12
**Origin:** GitHub issue #264 — pinned model IDs silently rot when a provider deprecates or renames a model.
**Related:** #188 (dotted-ID fold), #189 (frontier catalog), `pricing/prices.json` (single source of truth), DIP108 (unknown-model lint), the 2026-08-07 pricing-package design.

---

## 1. Problem

A `.dip` file pins a concrete model ID on every agent node:

```dip
agent reviewer {
  provider: anthropic
  model: claude-opus-4-6
}
```

That ID is a point-in-time fact. Providers churn: models are renamed, superseded, and eventually **retired** (Anthropic's own lifecycle: Active → Legacy → Deprecated → Retired; retired requests *fail outright*). When `claude-opus-4-6` is retired, this file does not warn — it keeps validating (the ID is still in our catalog with `deprecated: true`, priced for Bedrock/Vertex passthrough) right up until the downstream runtime calls a dead endpoint. The author's *intent* was "use the current best Opus," but the file encodes a frozen snapshot that quietly drifts out from under them.

### Concrete rot example

Author writes, in Q2, a workflow pinned to `claude-opus-4-6` and `gpt-5.2`. Six months later:

- Anthropic retires `claude-opus-4-6`; the current flagship is `claude-opus-4-8`.
- OpenAI's `gpt-5.2` is superseded by `gpt-5.6-*`.

The `.dip` file is unchanged, still lints clean (DIP108 only flags *unknown* IDs, not deprecated ones by default), and the runtime either errors on a retired model or silently runs an old one the author didn't mean to keep using. Every pinned workflow across the fleet has to be hand-edited to recover. **We want authors to be able to say what they mean — "the current top Opus" — and have that survive provider churn.**

### The tension to resolve

Two goals pull in opposite directions:

- **Drift-resistance** wants a *mutable* reference ("latest Opus") that follows the provider.
- **Reproducibility** wants an *immutable* reference (an exact ID) so a workflow runs the same model today and next year.

The maintainer's proposed spellings (`anthropic-opus-latest`, `anthropic-sonnet-4`, `anthropic-opus-sota`, "second best in family") are all *mutable* references. The core design question is whether mutable aliases and reproducibility can coexist. **They can — but only if we separate the reference an author writes from the resolution a build records.** That is the throughline of this spec.

---

## 2. Prior art and the lessons we take

Surveyed both LLM providers and general versioning systems. Full source list at the end; the load-bearing lessons:

| System | Mechanism | Lesson taken |
|---|---|---|
| **OpenAI** | `gpt-4o` (rolling default) vs `gpt-4o-2024-08-06` (dated snapshot) vs `chatgpt-4o-latest` (bleeding, *not for production*) | Make the rolling/experimental pointer **syntactically distinct** from the managed default; don't overload one alias for both audiences. |
| **Anthropic** | `claude-opus-4-5` alias → latest snapshot; formal **Active→Legacy→Deprecated→Retired** lifecycle with ≥60-day retirement notice | The alias isn't the safety mechanism — the **named lifecycle + notice window** is. Tooling should *warn before things break*, keyed on lifecycle state. |
| **AWS Bedrock / Azure OpenAI** | Inference profiles / deployment names: a **user-controlled indirection handle** decoupled from the concrete model; Azure attaches an explicit **upgrade policy** (`auto` / `pinned` / `upgrade-when-retired`) | Cleanest DSL shape in the set: **name the reference once; make its drift behavior a separate, explicit field** — not a property baked into the string. |
| **Google Gemini** | `-latest` can resolve to *preview/experimental*, not just stable | Never let a rolling alias silently cross **maturity tiers**. Scope "latest" to stable, or make maturity explicit. Directly relevant: our catalog has `-preview` and `[unpriced]` entries. |
| **OpenRouter** | `model:intent` suffix grammar (`:free`, `:nitro`, `:floor`), `openrouter/auto` router | A compact composable suffix grammar is good *syntax*; intent-based references are inherently **non-reproducible** and must be opt-in. |
| **Docker** | mutable `latest` vs immutable `@sha256:…`; best practice is `tag@digest` | Offer a **`readable-name@pin` dual form**: the name is for humans/diffs, the pin is the reproducibility guarantee that defeats silent substitution. |
| **semver** `^`/`~`/`x` | one-operator drift-tolerance vocabulary; `0.x` special case is a footgun | A tiny operator vocabulary is powerful, but semantics must be **dead obvious** — no context-dependent rules. |
| **npm / Terraform / Go MVS** | declared *intent* (range/tag) + generated, committed *lock* (exact versions + hashes); Go is reproducible-by-default, drift-by-explicit-action | **Separate declaration from resolution.** Readable *and* reproducible is achievable as two layers, not one string. Terraform's partial lock (providers but not modules) warns: **lock everything, or the lock lies.** |
| **Ubuntu `stable` vs `bookworm`; browser channels** | pin-the-rolling-channel vs pin-the-frozen-release; named risk tiers (`stable`/`beta`/`canary`) | Same-looking references with opposite drift behavior must be **visually unambiguous**. Named tiers ("give me the tested one") are a very human way to express drift tolerance when reproducibility isn't the goal. |

**Synthesis.** The strongest cross-industry pattern (npm, Terraform, Go) is *declaration + resolution as two layers*. The cleanest single-reference ergonomics (Azure, Bedrock) is *the alias and its drift policy are separate fields*. Docker's `tag@digest` shows the two can even live in one token. This spec proposes an alias scheme that is drift-resistant to write and reproducible to run — by resolving aliases to concrete IDs at author time (`dippin fmt`), the way `npm install` writes a lockfile.

---

## 3. Proposed alias scheme

### 3.1 Syntax

An alias is a value an author may write **anywhere a `model:` is accepted** (per-node `model:` and top-level `defaults.model`). Grammatically it is just a string; the meaning is assigned by the resolver against the catalog. Two spellings, chosen to match the existing `provider` / `model` split rather than the maintainer's single-token `anthropic-opus-latest` (see rationale below):

```dip
# Form A — provider stays explicit, model carries the alias (RECOMMENDED)
agent reviewer {
  provider: anthropic
  model: opus@latest        # family "opus", selector "latest"
}

# Form B — self-describing single token, provider omitted/inferred
agent reviewer {
  model: anthropic/opus@latest
}
```

**Alias grammar:** `[<provider>/]<family>@<selector>`

- `<provider>` — optional; a catalog provider or provider-alias key (`anthropic`, `google`→`gemini`, …). If omitted, taken from `provider:` / `defaults.provider`.
- `<family>` — a **catalog-defined family tag** (`opus`, `sonnet`, `haiku`, `gpt-flagship`, `gpt-mini`, `gemini-pro`, `gemini-flash`, …). *Not* string-parsed from the model ID (see §3.3 — the catalog is too irregular for that).
- `<selector>` — one of a small, fixed vocabulary:
  - `@latest` — newest **stable** member of the family (highest `rank`, excluding `preview`/unpriced and `deprecated`).
  - `@stable` — one rank below `@latest`; the "deliberate second-best" for teams that want a soak period before adopting the newest. Falls back to `@latest` if the family has only one member.
  - `@sota` — "best in family." **Distinct from `@latest`**: `latest` = newest by release, `sota` = highest *capability rank* (they usually coincide, but a newer cost-optimized mini could be "latest" without being "best"). See open question 2.
  - `@<major>` — newest patch within a declared major line (`opus@4` → newest `claude-opus-4-*`). The direct analogue of Docker's `1.2` tag and semver `~`.

**Why `family@selector`, not the maintainer's `anthropic-opus-latest`.** Three reasons: (1) `@` is a hard, unambiguous boundary between "what family" and "how to pick" — `anthropic-opus-latest` collides visually with real IDs like `claude-opus-4-5` and can't be distinguished from a provider that literally ships a model called `opus-latest`; (2) it keeps `provider:` doing its existing job instead of folding the provider into the model token; (3) it mirrors Docker `repo:tag` / `image@digest`, which authors already know. The maintainer's four intents map cleanly onto the selector vocabulary: `-latest`→`@latest`, `-sonnet-4`→`sonnet@4`, `-sota`→`@sota`, "second best"→`@stable`.

### 3.2 Resolution semantics

Resolution is a pure function `resolve(provider, family, selector) → (modelID, ok)` over the catalog. For a family's member set (all catalog entries tagged with that `family`, for that canonical provider):

1. **Filter** to *eligible* members: `priced != false`, `deprecated != true`, `maturity == stable` (i.e. not `preview`). `@sota`/`@latest` never resolve to a preview or a retired model.
2. **Order** the eligible set by the catalog's `rank` (integer; higher = newer/better — see §3.3).
3. **Select:**
   - `@latest` → `max(rank)`.
   - `@stable` → second-highest `rank` (or `max` if only one).
   - `@sota` → `max(capability_rank)` if that field exists, else `max(rank)` (they coincide until a family needs to distinguish them).
   - `@<major>` → `max(rank)` among members whose ID is in that major line. Requires a `major` tag or a parseable major on the member (Anthropic-clean families support this; irregular families may only support `@latest`/`@stable`).
4. **Map** the selected member back to its concrete `model` ID.

Selectors are the *entire* vocabulary — no ranges, no operators, no wildcards. This is deliberate (semver `0.x` lesson: keep it dead-obvious). Everything else is an exact ID, exactly as today.

### 3.3 The hard part — how do we know "newest"/"best"/"second-best"?

**We cannot reliably derive ordering by parsing model IDs.** The catalog dump proves it. Anthropic is clean (`claude-<tier>-<major>-<minor>`), but:

- OpenAI ships **codenames**: `gpt-5.6-luna`, `gpt-5.6-sol`, `gpt-5.6-terra` — no ordering in the string, and `gpt-5.5` vs `gpt-5.6-luna` vs `gpt-5-pro` don't sort by capability.
- Grok encodes **variants**: `grok-4.20-0309-reasoning` vs `-non-reasoning` vs `-multi-agent`.
- Gemini mixes stable, `-preview`, and `-customtools`; MiniMax has `M2.7-highspeed`; Mistral has date-stamped and named lines side by side.

Any string-parsing heuristic will mis-rank one of these within a release. **Therefore ordering must be explicit catalog metadata, not inferred.** Proposal — three new *optional* fields on a catalog entry (`pricing/prices.json`), all backward-compatible (absent = not alias-addressable):

```json
{
  "provider": "anthropic",
  "model": "claude-opus-4-8",
  "family": "opus",          // NEW: the alias family tag
  "rank": 480,               // NEW: ordering within family (higher = newer); e.g. major*100+minor
  "maturity": "stable",      // NEW (optional): "stable" | "preview" (absent ⇒ stable)
  "input_per_m": 15, "output_per_m": 75,
  "source": "…", "as_of": "2026-08-10"
}
```

- **`family`** groups members an alias can select among. A model with no `family` is simply not reachable by an alias (only by exact ID) — so the feature rolls out family-by-family, no big-bang migration.
- **`rank`** gives a total order *within a family*. It does the job the ID string can't. A simple, reviewable convention (e.g. `major*100 + minor`, so `claude-opus-4-8` = 480, `claude-opus-5` = 500) makes diffs obvious and `@stable` = "second-highest rank" well-defined even across the 4→5 major boundary.
- **`maturity`** keeps `@latest`/`@sota` from ever picking a preview (the Gemini lesson). Defaults to stable so existing entries need no change.
- **`@sota` vs `@latest`:** for the common case they're the same field (`rank`). Only if a family ever has a newest-but-not-best member do we add a separate `capability_rank`; deferred as an open question rather than built speculatively.

**Who maintains the ordering?** The same humans and the same daily auto-sync PR flow that already maintain `prices.json` (per the 2026-08-07 pricing design). `rank`/`family` are reviewable data, set when an entry is added. The auto-sync job can *propose* `family`/`rank` for a new model (highest existing rank + increment) in its PR, but a human confirms — ranking is a judgment call (is `gpt-5.6-luna` a flagship or a specialized variant?) and must not be auto-merged. A `prices_test.go` invariant enforces integrity: within a `family`, ranks are unique and every alias-addressable family has at least one eligible (non-deprecated, non-preview, priced) member — so `@latest` can never resolve to nothing silently.

---

## 4. Resolution timing — the reproducibility decision

Three options; this is the central call.

**Option 1 — pure runtime alias (mutable forever).** The `.dip` file keeps `model: opus@latest`; the downstream runtime resolves at execution. Maximum drift-resistance, **zero reproducibility** — two runs a month apart may use different models, and the `.dip` doesn't record which. This is `chatgpt-4o-latest` / `openrouter/auto`: explicitly not-for-production per those very vendors. Rejected as the *default*.

**Option 2 — author-time pin-resolution (RECOMMENDED).** Aliases are **sugar that `dippin fmt` resolves in place**, exactly like `npm install` turning a range into a locked version. The author writes intent; a deterministic tool writes the fact:

```dip
# author writes:
model: opus@latest
# `dippin fmt` (or `dippin pin`) rewrites, against today's catalog, to:
model: claude-opus-4-8   # opus@latest, pinned 2026-08-12
```

The committed file is a **concrete, reproducible ID** (a trailing comment records the alias it came from, so intent survives and re-pinning is mechanical). Re-running `dippin fmt` after a catalog bump re-resolves — a *reviewable diff* that shows exactly which node moved from `4-8` to `5`, gated by a human, keyed to a catalog version. This is the Docker `tag@digest` idea and the npm/Terraform declaration-vs-resolution split, adapted to a formatter we already ship. Drift becomes an **explicit, diffable action** (Go MVS's "frozen unless bumped"), not a silent runtime surprise.

**Option 3 — a lockfile.** Keep the alias in the `.dip` and emit a sibling `dippin.lock` mapping `alias → resolved ID @ catalog date`. More machinery (a second artifact to commit, a `dipx` pack question), and dippin has no lockfile concept today. The in-file pin (Option 2) gets the same reproducibility with zero new artifacts because the `.dip` *is* the thing that ships. Rejected as heavier than needed; revisit only if authors demand the alias stay literally in the source.

### Recommendation

**Adopt Option 2, with two knobs:**

1. **`dippin fmt` resolves aliases to pinned IDs by default** (with the origin comment). The reproducible artifact is what gets committed and packed. `dippin lint`/`check` emit an advisory (new code, §5) on any *unresolved* alias still in a file that's about to be packed — "this alias hasn't been pinned; run `dippin fmt`."
2. **An opt-out for teams that genuinely want runtime resolution** (evergreen fleets that accept drift): a top-level `resolve_models: runtime` directive tells `fmt` to leave aliases literal and signals the runtime to resolve them. This is the Azure "upgrade policy is a separate explicit field" lesson — the *alias* says which family; a *separate directive* says when it resolves. Default is `pin` (author-time). Building this knob is deferred to Phase 3; Phases 1–2 ship pin-only.

This gives both properties the problem statement demanded: authors write drift-resistant intent; the shipped/packed artifact is reproducible; moving forward is a reviewed diff, not silent rot.

---

## 5. Catalog interaction (DIP108, cost, flags, new fields)

**An alias MUST price.** Cost estimation (`dippin cost`) and DIP108 both run on the *resolved* ID. Under the recommended Option 2 this is automatic: by the time `cost`/`lint` see the file, `fmt` has already pinned the alias to a concrete catalog ID that prices normally. For the runtime-resolution opt-out, `cost` resolves the alias through the same `resolve()` function before pricing, so an alias never estimates at $0 — if it did, we'd have reintroduced the unpriced-model hole.

**DIP108 (unknown model/provider).** Today it flags IDs absent from the catalog. Extensions:

- An alias whose `family@selector` **resolves** is *known* — no DIP108.
- An alias referencing an **unknown family** (`opus2@latest` where no entry tags `family: opus2`) → DIP108-style error, help text listing known families for the provider (mirrors the existing "known models for X" help).
- No change to exact-ID behavior.

**`deprecated` / `priced` flags** are precisely what make resolution safe:

- Resolution's eligibility filter (`§3.2 step 1`) *excludes* `deprecated` and `priced:false` members. So `@latest`/`@sota` structurally **cannot** land on a retired or unpriced model — the exact failure the whole feature exists to prevent.
- This also motivates a **new, separate lint on exact deprecated IDs** (§6, DIP-D): today pinning `claude-opus-4-6` (deprecated) lints clean. That's the rot. A new warning — "node uses deprecated model `claude-opus-4-6`; consider `opus@latest`" — is the nudge from pinned-and-rotting toward alias-and-fresh, and is arguably worth shipping *first*, independently.

**New `prices.json` fields (all optional, backward-compatible):** `family` (string), `rank` (int), `maturity` (`"stable"|"preview"`, absent ⇒ stable), and *deferred* `capability_rank` (int, only if `@sota` must ever diverge from `@latest`). No existing field changes; no price row changes. An entry without `family` is simply not alias-addressable. **This spec proposes these fields; it does not modify `prices.json`.**

---

## 6. Failure modes and exact lint/error behavior

| Situation | Detection | Behavior |
|---|---|---|
| **Alias resolves to nothing** — family exists but every member filtered out (all deprecated/preview), or family has zero eligible members | `resolve()` returns `ok=false` at `fmt`/lint time; also caught statically by the `prices_test.go` invariant so it can't ship | **Error** (fails `lint`/`check`, error-severity like DIP155–158). Message: `alias "opus@latest" resolves to no eligible model (all members deprecated/preview)`. `fmt` refuses to pin. |
| **Unknown family** — `opus2@latest`, no catalog entry tags that family | Family tag not present for provider | **Error**, DIP108 family variant. Help: `known families for anthropic: opus, sonnet, haiku`. |
| **Alias resolves to a `deprecated` model** | Cannot happen for `@latest`/`@stable`/`@sota` (filtered). Only possible if a pinned `@<major>` line has *only* deprecated members | **Error** if no eligible member in that major (`opus@3` when all opus-3 are retired): `no active model in family "opus" major 3`. |
| **Ambiguous family** — a model tags a `family` that could match two provider groupings, or duplicate `rank` within a family | `prices_test.go` invariant (unique `(family, rank)`; family belongs to one canonical provider) | **Build-time test failure** in the pricing package — never reaches a user. Catalog integrity is enforced at our CI, not the author's lint. |
| **Selector typo** — `opus@newest` | Selector not in the fixed vocabulary | **Error** at parse/lint: `unknown selector "newest"; valid: latest, stable, sota, or a major number`. |
| **Exact deprecated ID pinned** (`claude-opus-4-6`) | Catalog `deprecated: true` on the resolved entry | **New warning (DIP-D, next free code)**: `node "reviewer" uses deprecated model "claude-opus-4-6"; consider alias opus@latest`. Advisory, does not fail `check`. This is the drift *smoke detector* for the millions of already-pinned files. |
| **Unresolved alias reaches pack** (runtime-resolution off, author forgot `fmt`) | Alias still literal at pack time in `pin` mode | **Warning**: `unpinned alias "opus@latest"; run dippin fmt to pin`. |

Every "resolves to nothing / to a bad model" case is caught **either** by a build-time catalog invariant (our problem, our CI) **or** by an author-facing lint (their problem, their diff) — never silently at the runtime call, which is the status quo we're eliminating.

---

## 7. Phased implementation plan

- **Phase 0 — deprecation smoke detector (ship first, independently).** Add the DIP-D warning on exact `deprecated` model IDs. Pure detection over existing catalog data, no new fields, no syntax. Immediately surfaces the existing rot in real files and de-risks the rest by validating that "deprecated" flows to a lint. *No grammar change.*
- **Phase 1 — catalog metadata.** Add optional `family`/`rank`/`maturity` to `pricing/prices.json` schema + `ModelPrice`/`fileEntry`, backfill Anthropic families first (cleanest), add `prices_test.go` invariants (unique `(family,rank)`, ≥1 eligible member per family). Add `pricing.ResolveAlias(provider, family, selector)`. *No `.dip` syntax yet — pure pricing-package work, fully testable in isolation.*
- **Phase 2 — alias resolution in the toolchain.** Accept `family@selector` in `model:`/`defaults.model`. Wire DIP108 + cost through `ResolveAlias`. Implement `dippin fmt` pin-resolution (rewrite alias → ID + origin comment) and the failure-mode lints (§6). This is the user-visible feature in `pin` mode. Sweep docs/site/skill/editor surfaces (per the standing docs rule).
- **Phase 3 — runtime-resolution opt-out (only if wanted).** `resolve_models: runtime` directive + the pack-time "unpinned alias" warning + downstream runtime contract. Gated on a real ask; `pin` mode covers the stated problem alone.

Backward-compatible throughout: absent `family` ⇒ no alias-addressability; every current `.dip` and every current catalog entry keep working untouched.

---

## 8. Open questions for the maintainer

1. **`@stable` = "second-best" semantics.** Is "second-highest rank in family" the right definition of the deliberate-stability tier? Edge cases: a family with one member (fall back to latest — proposed), and whether `@stable` should instead mean "latest that's been out ≥N days" (needs a release-date field, not just a rank). Rank-based is simpler; date-based is truer to "soak period." Which?
2. **Do we need `@sota` distinct from `@latest` at all today?** Every current family's newest *is* its best, so one `rank` field suffices now. Adding `@sota` + `capability_rank` speculatively violates "no abstraction for single-use." Recommend shipping `@latest`/`@stable`/`@major` and treating `@sota` as an alias for `@latest` until a family actually needs the split. Agree?
3. **Pin-by-default vs runtime-by-default.** The recommendation is author-time pinning (`fmt` resolves; committed file is concrete) for reproducibility, with runtime resolution as an opt-in directive. Is that the right default for your downstream consumers — or do the evergreen fleets want the alias to stay literal in the source and resolve at run time (Phase 3 becomes Phase 1)?

Secondary: (4) exact alias *token* — `opus@latest` (this spec) vs the maintainer's `anthropic-opus-latest` vs a suffix form; (5) the `rank` numbering convention (`major*100+minor`) and who owns assigning it when auto-sync proposes a new model; (6) should `dippin fmt` *always* re-resolve aliases (a `fmt` can change which model runs) or only a dedicated `dippin pin` command, keeping `fmt` purely cosmetic?

---

## Sources

- OpenAI models & aliases: platform.openai.com/docs/models, developers.openai.com/api/docs/models/gpt-4o
- Anthropic deprecations/lifecycle: platform.claude.com/docs/en/about-claude/model-deprecations, anthropic.com/research/deprecation-commitments
- AWS Bedrock inference profiles: docs.aws.amazon.com/bedrock/latest/userguide/inference-profiles.html, /cross-region-inference.html
- Google Gemini/Vertex versions: ai.google.dev/gemini-api/docs/models, developers.googleblog.com (Gemini 1.5 update)
- Azure OpenAI model versioning: learn.microsoft.com/azure/foundry/foundry-models/concepts/model-versions, /openai/how-to/working-with-models
- OpenRouter variants & routing: openrouter.ai/announcements/introducing-nitro-and-floor-price-shortcuts, /docs/faq
- models.dev catalog: models.dev, opencode.ai/v2/docs/models
- Docker digests: docs.docker.com/dhi/core-concepts/digests
- semver: github.com/npm/node-semver, docs.npmjs.com/cli/v6/using-npm/semver
- npm dist-tags: docs.npmjs.com/cli/dist-tag
- Debian/Ubuntu apt pinning: manpages.debian.org/apt_preferences.5, wiki.debian.org/AptConfiguration
- Go MVS: research.swtch.com/vgo-mvs
- Terraform lock file: developer.hashicorp.com/terraform/language/files/dependency-lock
- Chrome release channels: chromium.org/getting-involved/chrome-release-channels
