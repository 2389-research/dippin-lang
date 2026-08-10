---
title: "Changelog"
description: "Version history and release notes for dippin-lang."
navActive: "changelog"
layout: "changelog"
---
## [v0.57.0] — 2026-08-10

### Added
- **`prices.json` now carries cache read/write rates for the verified providers** ([#210](https://github.com/2389-research/dippin-lang/issues/210)). `pricing.ModelPrice`/`pricing.Cost` have always honored cache fields, but the data left them unset, so cache traffic priced at $0 and consumers (tracker) overlaid their own cache table — two tables that could drift. Verified 2026-08-10 against official docs and populated: **Anthropic** `cache_read_mult: 0.1` + `cache_write_mult: 1.25` (0.1× read, 1.25× 5-minute write, uniform across models); **OpenAI** and **Gemini** `cache_read_mult: 0.1` (0.1× read; OpenAI has no per-token write charge, Gemini's cache cost is an hourly storage fee outside the per-token schema). The other providers (DeepSeek, Mistral, Cohere, xAI, Z.AI, Moonshot, MiniMax, Qwen) are left at 0 — their cache conventions weren't verifiable from official pages, so a consumer keeps overlaying those until they are. A downstream consumer can now drop its cache overlay for Anthropic/OpenAI/Gemini and read the rates from `pricing`.


## [v0.56.1] — 2026-08-10

### Fixed
- **`gpt-5.2-codex` now prices instead of escaping cost ceilings** ([#209](https://github.com/2389-research/dippin-lang/issues/209)). tracker's cutover to `dippin-lang/pricing` (tracker#558) surfaced that `gpt-5.2-codex` was absent from the catalog, so `pricing.Lookup` returned `found=false` and it priced at **$0**, silently escaping `--max-cost`. Verified 2026-08-10 (LiteLLM + OpenRouter agree, consistent with `gpt-5.3-codex` at its base rate on OpenAI's official page): it bills at the `gpt-5.2` rate ($1.75/$14), so it's now an **alias** of `gpt-5.2`. (`gpt-5.2-mini`, the other model in #209, does not exist in any source — official, LiteLLM, or OpenRouter — and remains a tracker-side catalog fix.)
- **DIP159 no longer false-positives a `file`/`secret` input consumed via its staged path** ([#215](https://github.com/2389-research/dippin-lang/issues/215), regression in v0.56.0). A `file` or `secret` input is consumed out-of-band by reading its staged path in a shell command — and DIP157 forbids `${inputs.x}` inside a `command:` — so it is legitimately never `${inputs.x}`-referenced. DIP159 was flagging it as a dead input, a DIP157/DIP159 pincer that made the documented staged-file pattern un-lintable (surfaced adopting v0.56.0 in tracker). `file` and `secret` inputs are now exempt from DIP159; `text`/`number`/`bool`/`enum` are still dead-input-checked.


## [v0.56.0] — 2026-08-10

### Added
- **Input constraint + dead-input lints (DIP158, DIP159) — Phase 2 of the `inputs` feature** ([#190](https://github.com/2389-research/dippin-lang/issues/190)). **DIP158** (error) validates an input's constraints against its declared type: an `enum` `default` must be one of its `options`, a `number`'s `min` must not exceed its `max`, a `pattern` must be a valid regex, and a constraint must apply to the type it's set on (`pattern`/`max_length`/`multiline` → text/secret, `min`/`max` → number, `options` → enum) — so `max_length` on a `bool` or `options` on a `text` is flagged. **DIP159** (warning) flags a declared input that no prompt, tool command, or edge condition references (a dead input, mirroring DIP107). Brings the catalog to **69 codes (DIP101–DIP159)**. Remaining Phase 3: DIP160 (cross-file subgraph arity), DIP161 (chain-attack input flow).


## [v0.55.0] — 2026-08-10

### Changed
- **`fmt --migrate` now signals non-runtime-equivalent conversions with a distinct exit code 4** ([#186](https://github.com/2389-research/dippin-lang/issues/186)). A non-self `retry_target` (retry jumps to a *different* node) migrates to a `loop` edge, but the engine dispatches the retry outcome on a channel that reads node attributes, never the edges block — so the loop edge is ignored and the retry silently becomes a self-retry. Previously this exited `3` ("review needed"), indistinguishable from equivalent-but-review cases. It now exits **`4`** (`ExitMigrateNonEquivalent`) with an explicit `NOT runtime-equivalent … do not use as a drop-in` error, so bulk migrations can separate safe conversions (0), review-but-equivalent (3), and non-equivalent (4). The output is still emitted (best-effort) with a sharper in-file marker. This is an interim safety signal; the underlying design gap — dip 2 has no representation the retry engine reads for non-self retry routing or retry-exhaustion fallback — is tracked separately for a proper fix.


## [v0.54.1] — 2026-08-08

### Fixed
- **Corrected two stale prices flagged by the daily drift check** (#201, verified against official docs 2026-08-08): `openai/gpt-4.1-nano` was $0.05/$0.20 but is **$0.10/$0.40**; `deepseek/deepseek-v4-pro` was $1.74/$3.48 (pre-discount list) but the launch discount has become the standing price at **$0.435/$0.87**. Other drift candidates were dispositioned as non-issues — `claude-sonnet-5` 2/10 is the intro rate through 2026-08-31 (durable 3/15 kept); `o3-mini`/`o4-mini`/`gpt-4.1-nano` "deprecated" were false positives from models.dev (official docs list them active). `mistral-nemo`, `mistral-small-2603`, and `cohere/command-r-08-2024` could not be verified from official pages (JS-gated) and were left unchanged pending manual verification.


## [v0.54.1] — 2026-08-08

### Added
- **Daily pricing-drift detection (Phase 3), human-gated.** A scheduled GitHub Action (`.github/workflows/pricing-sync.yml`) runs `pricing-sync` daily and, when it finds **price or deprecation drift for models already in the catalog**, opens/updates a single rolling `pricing-drift` issue with the evidence. It **never edits `prices.json` or opens a code PR** — a human confirms each candidate against its official `source` (models.dev can reflect intro/discounted rates, so this stays gated). New `sync --existing-only` flag keeps the daily signal actionable (drops the hundreds of upstream image/tts/embedding models dippin doesn't price); `sync --fail-on-changes` for CI gating.
- **`docs/pricing-integration.md`** — how a downstream consumer (tracker) pins and integrates the `pricing` package: version floor (`v0.53.0`), the `Lookup`/`Cost` API, mapping `llm.Usage → pricing.Usage`, keeping capability metadata in the consumer, and the current **cache-pricing caveat** (the `ModelPrice` cache fields exist but `prices.json` doesn't populate them yet, so `Cost` computes cache traffic at $0 until it does).

## [v0.54.0] — 2026-08-07

### Added
- **Pricing staleness checker + assistive sync tool** (`cmd/pricing-sync`, Phase 2 of the pricing-tooling design). `just check-prices` reports catalog entries whose `as_of` is older than a freshness window (default 45 days) or missing a source, and exits non-zero — usable as a CI/pre-commit re-verification reminder (it currently flags the 2026-06-10 batch). `just sync-prices` fetches the machine-readable **models.dev** aggregator (JSON, no auth, no HTML scraping), diffs it against `pricing/prices.json`, and reports candidate changes — new models, price deltas (with a `--tolerance` to suppress small ones), and upstream deprecations. It is **report-only** and never edits the catalog: detection is mechanical, but authoritative verification is not (no provider publishes price as data), so a human confirms each candidate against its official `source`. The staleness logic lives in the `pricing` package (`StaleEntries`, pure/deterministic); the diff engine is pure and fixture-tested.

## [v0.53.0] — 2026-08-07

### Changed
- **Single source of truth for model pricing + catalog: the new leaf `pricing` package.** Model prices lived in `cost/pricing.go` and were mirrored by hand in `validator/lint_model.go`'s DIP108 catalog (and again in tracker) — the exact drift trap #189 had to maintain by hand. Prices now live in one embedded data file, **`pricing/prices.json`**, loaded by a leaf `pricing` package (imports nothing from dippin) that exposes `ModelPrice`, a neutral `Usage`, one `Cost()` calc, and a total alias-resolving `Lookup()`. `cost.DefaultPricing()` projects the catalog (public API unchanged); `validator`'s DIP108 derives its known-model set from it (the hand-mirrored maps are deleted). Each entry carries a `source` URL + `as_of` date (provenance travels with the number), and a `"priced": false` flag marks known-but-unpriced models (e.g. Qwen). Purely internal — `dippin cost`/`lint` output is unchanged, verified by the existing price-assertion and `TestLintExamples` suites. Downstream consumers (tracker) can now import `github.com/2389-research/dippin-lang/pricing` directly instead of maintaining their own table. Phase 1 of the pricing-tooling design (`docs/superpowers/specs/2026-08-07-pricing-package-and-autosync-design.md`); the daily auto-sync via machine-readable aggregators follows.

## [v0.52.0] — 2026-08-07

### Added
- **Frontier model pricing + catalog refresh** ([#189](https://github.com/2389-research/dippin-lang/issues/189)). `dippin cost` reported **$0** and `dippin lint` fired spurious **DIP108** for a raft of current frontier models simply absent from the tables. Added, each verified against official provider docs on 2026-08-07 (source URLs recorded in `cost/pricing.go`): **Anthropic** `claude-opus-5` ($5/$25), `claude-sonnet-5` ($3/$15 durable; intro $2/$10 through 2026-08-31); **OpenAI** `gpt-5.6-sol` ($5/$30), `gpt-5.6-terra` ($2/$12), `gpt-5.6-luna` ($0.20/$1.20); **Gemini** `gemini-3.6-flash` ($1.50/$7.50), `gemini-3.5-flash` ($1.50/$9), `gemini-3.5-flash-lite` ($0.30/$2.50); **xAI** `grok-4.5` ($2/$6 base), `grok-build-0.1` ($1/$2). Plus three **new providers**: **Z.AI / GLM** (`zai` — glm-5.2/5.1/5/5-turbo/4.7/4.6/4.5 family), **Moonshot / Kimi** (`moonshot`+`kimi` — `kimi-k3` $3/$15), and **MiniMax** (`minimax` — MiniMax-M3/M2.7/M2.5/M2.1/M2). All base-tier list prices (large-prompt and cache tiers are out of dippin's schema).
- **Qwen recognized by the linter** (`qwen` — qwen3.7-max/plus, qwen3.6-flash). Its international per-token USD pricing is console-gated and could not be verified from an official page, so Qwen is in the model catalog (no more DIP108) but has **no cost-table row yet** — a follow-up.

### Docs
- Swept the living docs and website to current code: diagnostic-code counts and ranges corrected to **67 codes (DIP001–DIP010, DIP101–DIP157)** across `CLAUDE.md`, `AGENTS.md`, `README.md`, `docs/`, and `site/content/`; the `inputs` block and `${inputs.*}` namespace added to `docs/llm-reference.md` (feeds the embedded spec) with DIP155–DIP157; corrected the `skill.md` canonical section order (**inputs → defaults → vars**, not defaults → inputs); updated the supported-provider list.

## [v0.51.1] — 2026-08-07

### Fixed
- **Dotted Anthropic model IDs now price and lint correctly** ([#188](https://github.com/2389-research/dippin-lang/issues/188)). The pricing table and model catalog key Anthropic models with dashed versions (`claude-haiku-4-5`), but a Vercel AI Gateway–routed runtime documents the dotted spelling (`anthropic/claude-haiku-4.5`) — which a `.dip` must carry, since the first-party API rejects dotted IDs. The exact-match lookup missed the dotted form, so `dippin cost` silently reported **$0** and `dippin lint` fired a spurious **DIP108** "unknown model". Both `cost.lookupPrice` and the validator's `modelKnown` now fall back to a version-separator-insensitive match (`.` and `-` fold together) after the exact match, so a dotted ID resolves to its dashed catalog key and vice versa. Exact matches are unchanged — the legitimately dotted OpenAI/Gemini keys (`gpt-5.5`, `gemini-3.1-flash-lite`) still resolve as before, and no two keys in any provider collide under the fold, so the fallback can only turn a miss into the intended hit.

## [v0.51.0] — 2026-08-07

**Native `inputs` declaration — typed, introspectable input schema** ([#190](https://github.com/2389-research/dippin-lang/issues/190)). Phase 1 of pipeline inputs. A workflow-level `inputs` block is the callee-side signature declaring what a caller must supply — a human at the entry point, or a parent workflow via a subgraph's `params:`. One-line form (`name: type`) or an indented attribute block; six types (text, number, bool, enum, file, secret) and ten attributes (required, default, prompt, description, options, pattern, min, max, max_length, multiline). `${inputs.x}` is the first **closed** namespace in the language: unlike the open `ctx` namespace, a reference to an undeclared input is a lint error, not a maybe-valid pass-through — values are untrusted by construction. Fully additive to dip 1: a `.dip` with no `inputs` block is unchanged, and the parser accepts an unknown input type verbatim (only the lint complains), so a `.dip` using a future type still parses, formats, and packs on an older dippin.

### Added
- **`inputs` block** — declares the workflow's input signature; one-line `name: type` or an attribute block with `required`, `default`, `prompt`, `description`, `options`, `pattern`, `min`, `max`, `max_length`, `multiline`.
- **`${inputs.x}` namespace** — resolves against the declaration; referencing an undeclared input is a lint error in both prompts and edge conditions.
- **`DIP155` — unknown input type.** **`DIP156` — reference to an undeclared input** (prompts and edge conditions). **`DIP157` — `${inputs.x}` inside a tool `command:`**, which never interpolates (the runtime keeps the `inputs` namespace off its shell allowlist). All three are **error severity** — the catalog's first error-severity lint codes. Brings the catalog to 67 codes (DIP101–DIP157).
- **`dippin inputs <file> --format=json|text`** — typed JSON introspection surface for a host to collect values before a run (number/bool defaults coerced to real JSON types, declaration order preserved, empty inputs serialize to `[]` not `null`). `dippin inspect` on a `.dipx` bundle also surfaces the entry workflow's inputs.

### Changed
- **`dippin lint` exit code.** DIP155–157 are the first error-severity LINT diagnostics; `dippin lint` previously exited non-zero only on a structural `Validate` error, so an error-severity lint diagnostic would print and still exit 0. It now exits non-zero on either. This is the one non-additive behavior change in this release — CI pipelines pinned to `dippin lint`'s exit code should note it.

### Runtime pairing (requires an enforcing runtime)
- dippin carries the schema: it declares, lints, and introspects `inputs`, but does not collect, validate, or inject values at run start — that is the engine's job, paired with [tracker#553](https://github.com/2389-research/tracker/issues/553). The engine must collect input values at run start, validate them against the declared type/attributes, and inject them under `${inputs.x}`, honoring the untrusted-by-construction framing (an input value is caller-supplied and must never be trusted implicitly) and the shell-interpolation exclusion DIP157 signals (`inputs` must stay off the tool `command:` shell allowlist). dippin ships this independently and is not gated on the runtime (per `never-gate-dippin-on-tracker`).

## [v0.50.0] — 2026-07-20

Documentation-only release: brings the **embedded LLM specification** (shipped in the `dippin` binary) into line with the v0.49.0 language surface. No code or behavior change.

### Docs
- **Embedded spec / `skill.md` refresh for v0.49.0 quoted edge conditions** ([#187](https://github.com/2389-research/dippin-lang/pull/187), thanks [@lra](https://github.com/lra)). Documents lossless double-quoted condition values (`when ctx.msg = "hello world"`, with `\"`/`\\` escaping and literal `||`/`#` inside quotes); the `on <token>` shorthand's single-bare-identifier restriction (`[A-Za-z0-9][A-Za-z0-9_-]*`; quoted/multi-token values require `when`); the requirement to quote the reserved bare keyword `loop` used as a value; adds the missing **DIP010** row; corrects the lint range from DIP001–DIP152 to **DIP001–DIP154**; and fills in the CLI reference (`fmt --migrate`, `spec`, `export-dot --rankdir/--prompts`, `migrate --output`, `simulate --interactive`, `graph --compact`). Regenerated `cmd/dippin/generated-spec.md` and `site/static/llms-full.txt` to stay in sync with `skill.md`.
- **Review correction (`d1bd93b`)**: fixed a DIP010 mis-attribution in the above — an *unterminated* double quote is a lexer/parse error (`unterminated string literal at L:C`, reported at the opening quote), **not** DIP010 (which is for a condition that tokenizes but fails to parse, e.g. a bare `loop` used as a value). Independently corroborated by CodeRabbit and Copilot review.
- Added a development-practices reflection note under `docs/notes/`.

## [v0.49.0] — 2026-07-13

**Lossless quoted edge conditions** ([#182](https://github.com/2389-research/dippin-lang/issues/182)). Double-quoted edge-condition values containing escaped quotes or backslashes were corrupted while the parser reconstructed `Condition.Raw`; embedded operator-like text could then be reinterpreted, and valid workflows failed validation with DIP010. This release makes the double-quoted path escape-aware end to end, from source parsing through `Condition.Raw` to the condition parser used by simulation and validation.

### Fixed
- **Escaped double-quoted condition values now round-trip losslessly.** Interior `\"` and `\\` survive parse → `Condition.Raw` → simulate/validate with their literal values intact. Operator- and comment-like text inside the quoted value — including `||` and `#` — remains literal, while a real trailing `#` comment is still stripped. Additional boundary coverage contributed in [PR #183](https://github.com/2389-research/dippin-lang/pull/183) (thanks [@harperreed](https://github.com/harperreed)) caught the comment-boundary/backslash-parity gap and covers even/odd backslash runs, escaped quotes, literal tabs, and UTF-8.
- **Unterminated double-quoted literals are rejected.** The lexer now reports an unterminated string literal at the opening quote's source line and column instead of fabricating a closing quote and allowing validation to pass. Existing single-quoted condition behavior is unchanged, including YAML-style normalization and literal backslashes.

### Runtime pairing (requires an enforcing runtime)
- None. This is parser/validator correctness with no new runtime field or behavior to interpret. This `v0.49.0` tag is the release [tracker#444](https://github.com/2389-research/tracker/issues/444) can consume.

## [v0.48.0] — 2026-07-10

**Shared prompt fragments** ([#175](https://github.com/2389-research/dippin-lang/issues/175)). `prompt_file:` / `system_prompt_file:` load an *entire* prompt from a file — all or nothing — so control-protocol boilerplate that must be byte-identical across many agents (e.g. a STATUS / FINAL-LINE contract) gets hand-pasted and drifts (downstream: pipelines#111, the same block duplicated 65× across 11 files). This release lets a shared fragment be declared once and applied to many agents. Fully additive and non-breaking: every existing `.dip` parses, validates, formats, and packs unchanged, in v1 and `dip 2` alike.

### Added
- **`defaults` prompt cascade** ([#175](https://github.com/2389-research/dippin-lang/issues/175)). The `defaults` block gains `prompt_prefix:` / `prompt_suffix:` (inline literal) and `prompt_prefix_file:` / `prompt_suffix_file:` (fragment loaded from a file — the way to single-source across many `.dip` files). The declared fragment **cascades to every agent node** (tool/human/parallel/etc. have no prompt and are unaffected). The inline and file forms are mutually exclusive per side (setting both is a structural error).
- **Per-agent `prompt_include:`** — appends an extra fragment file after the agent's own body (for agents that need a fragment beyond the cascade).
- **Per-agent opt-out** — `prompt_prefix: none` / `prompt_suffix: none` opts an agent out of that side of the cascade. Only `none` is valid at node level (custom per-node override is a future extension).
- **`DIP154` (Hint)** — an agent sets `prompt_prefix: none` / `prompt_suffix: none` while the `defaults` block declares no cascade of that kind — a no-op opt-out (likely a leftover). Conservative, no false positives; surfaces in lint / check / watch / doctor. Brings the catalog to 64 codes (DIP101–DIP154).

### How composition works
The effective prompt is assembled at **resolve time** (the `pack` / `check` / runtime-load path; `dippin fmt` operates on unresolved IR and round-trips the directives verbatim) as **`prefix → body → prompt_include → suffix`**, joining only the non-empty parts — so the cascade **suffix is always the final content**, satisfying "the very last line MUST be exactly …" contracts. Fragment files use the **same security envelope** as `prompt_file` (relative-path containment, atomic leaf-symlink rejection, size cap, TOCTOU hardening). `dippin pack` inlines the fully-composed prompt (self-contained `format_version 1` bundle); `pack --no-inline` ships the fragment files under `workflows/` and keeps the directives, so the extracted tree composes byte-identically to a source run.

### Fixed
- **Migration-parity check now compares effective edges** (`migrate/parity.go`). Since v0.47.0's #136 stripped redundant fan edges from example `.dip` files, `dippin validate-migration` reported spurious `edge_missing` differences against the un-stripped example `.dot` siblings (a fork declared inline on a `parallel`/`fan_in` node is an *implicit* edge). `CheckParity` now keys `ir.EdgesFrom` (explicit **plus** synthesized parallel/fan_in edges) on both sides, so an inline fork matches an explicitly-drawn one. Real migration differences are still detected.

### Docs & tooling
- Documented the cascade + `prompt_include` + `none` opt-out in `docs/nodes.md`, `docs/GRAMMAR.ebnf`, `docs/{cli,integration,llm-reference}.md`, `site/content/{language,cli}.md`, and `site/static/skill.md`; added DIP154 to the validation pages; added the five keywords to the VS Code grammar (tree-sitter/Zed highlight fields generically, so no regen was needed); added `examples/shared_prompt_fragment.dip`; bumped the diagnostic catalog to 64 (DIP101–DIP154) across every hand-maintained surface; regenerated the embedded spec.

### Runtime pairing (requires an enforcing runtime)
- None. Composition is entirely resolve-time — a packed bundle carries the fully-composed prompt (inline) or the fragment files (no-inline), so the engine reads ordinary prompt content with no new field to interpret (per `never-gate-dippin-on-tracker`).

## [v0.47.0] — 2026-07-10

**Single-source `parallel`/`fan_in` — stop requiring edges-block re-declaration** ([#136](https://github.com/2389-research/dippin-lang/issues/136), Phase 1 of routing epic [#127](https://github.com/2389-research/dippin-lang/issues/127)). A `parallel Fan -> A, B` fan-out (and a `fan_in Join <- A, B` join) declares its fork inline, yet workflows routinely re-declared the same edges in the `edges` block (`Fan -> A`, `Fan -> B`, …), duplicating the fork and forcing authors to keep two declarations in sync by hand. An investigation confirmed the inline node-config list is already the single source of truth for every semantic consumer — reachability (DIP004), fan matching (DIP007), the built-in simulator, and `ir.EdgesFrom` all derive the fan edges from it; the edges-block copies are inert. This release makes that authoritative and removes the duplication.

### Added
- **`DIP153` — redundant parallel/fan_in edge** ([#136](https://github.com/2389-research/dippin-lang/issues/136)). Warns when an `edges` block declares an unconditional, attribute-free edge that merely repeats a fork already declared inline on a `parallel`/`fan_in` node. **Conservative — no false positives**: a *conditional* or *attributed* edge (`when`/`on`, `label:`, `choice:`, `weight:`, `override`, restart) between the same nodes is real routing, not a duplicate, and is never flagged; detection is a single shared predicate `ir.IsRedundantFanEdge` reused by the lint, the formatter, and the parser so the three can never disagree. Surfaces in lint / check / watch / doctor. Brings the catalog to 63 codes (DIP101–DIP153).
- **`dippin fmt` strips redundant fan edges**. Formatting removes the redundant edges-block re-declarations, leaving the inline `parallel`/`fan_in` line as the sole declaration. Idempotent and deterministic; an all-redundant edge list emits no dangling empty `edges` block. Conditional/attributed edges are preserved.

### Changed
- **Redundant fan-edge re-declaration is rejected under `dip 2`** ([#136](https://github.com/2389-research/dippin-lang/issues/136)). In a `dip 2` file, a redundant edges-block copy of an inline fork is a parse error (remediation: "the inline `parallel`/`fan_in` list is authoritative under `dip 2` — remove it, run `dippin fmt`"), consistent with `dip 2`'s rejection of `retry_target`/`fallback_target` (#134). Under v1 (the default) the re-declaration still parses and merely warns (DIP153) — fully non-breaking.
- **DOT export derives fan edges from node config**. `dippin export-dot` now synthesizes the parallel fan-out / fan-in arcs from `ParallelConfig.Targets` / `FanInConfig.Sources` (deduped against any explicit edge), so a file whose redundant edges have been stripped still renders complete fork arrows. Previously DOT drew fan arcs only from the edges block.
- **Examples cleaned** — removed 216 redundant fan-edge re-declarations across 11 example workflows (deletions only, no reformatting); a `TestLintExamples` guard now asserts zero DIP153 across the suite.

### Docs
- Documented that the inline `parallel`/`fan_in` list is the single source of truth (`docs/nodes.md`); added DIP153 to `docs/validation.md`, `site/content/validation.md`, `site/static/skill.md`; noted `fmt` strips redundant fan edges on the CLI pages; bumped the diagnostic catalog to 63 (DIP101–DIP153) across every hand-maintained doc/site surface; regenerated the embedded spec.

### Runtime pairing (requires an enforcing runtime)
- None. This is a pure authoring/analysis change — the inline list was already authoritative for the engine, so a migrated file routes fan-out/fan-in exactly as before; DIP153 and the `dip 2` rejection are enforced entirely at lint/parse time (per `never-gate-dippin-on-tracker`).

## [v0.46.0] — 2026-07-09

**`dip 2` — edges own destinations** ([#134](https://github.com/2389-research/dippin-lang/issues/134), epic [#127](https://github.com/2389-research/dippin-lang/issues/127)). The routing epic's structural payoff: under `dip 2`, a node's failure destination is no longer a node field — it is an edge, like every other routing decision. `retry_target:` and `fallback_target:` are the last two node-level routing knobs; `dip 2` removes them so that *all* control flow lives in the `edges` block and there is one place to read a graph's topology. Fully non-breaking for existing files: every valid v1 `.dip` still parses, validates, and formats unchanged, and `dip 2` is opt-in via the `dip 2` version header. A first-class migration (`dippin fmt --migrate`) mechanically rewrites v1 → `dip 2`.

### Added
- **`dippin fmt --migrate` — real v1 → `dip 2` transform** ([#134](https://github.com/2389-research/dippin-lang/issues/134)). Folds a `fallback_target:` into an `on fail` edge and a non-self `retry_target:` into a `loop` edge, then stamps the file `dip 2`. It is conservative and lossless-or-loud: a self-target `retry_target` (a no-op restart) is dropped; a target that already has an equivalent `on fail` / `loop` edge is deduped rather than duplicated (this previously could have produced a DIP009 duplicate-edge or DIP005 self-cycle in the migrated file — now caught and avoided); and any case it can't express 1:1 (e.g. a `fallback_target` that diverges from an existing on-fail edge) is preserved verbatim, both edges kept, and flagged with an inline `# MIGRATION:` comment plus a stderr summary. When any case needs human review the command exits **`3`** (`ExitMigrateReview`), distinct from the format-drift `1`. `--migrate --check` exits non-zero when a file is not already canonical `dip 2` (CI gate for a completed migration). Migration-only edge comments round-trip via a new `ir.Edge.Comment` field, emitted by the formatter as a leading `#` line.
- **`ir.EdgeRoutesOnFail(e)`** — the shared predicate for "this edge is the node's failure route" (`ctx.outcome`/`outcome` equals `fail`/`failure`), reused by the formatter, migration, and validator so the three never drift on what counts as an on-fail edge.

### Changed
- **`retry_target:` / `fallback_target:` are rejected as node fields under `dip 2`** ([#134](https://github.com/2389-research/dippin-lang/issues/134)). In a `dip 2` file the parser emits a diagnostic pointing at the offending field with the remediation "express the failure destination as an `on fail` edge (run `dippin fmt --migrate`)". Under v1 (the default when no `dip N` header is present) both fields parse exactly as before — this is a version-gated *semantic* rejection, not a grammar change, so the tree-sitter/VS Code/Zed grammars are untouched (the fields still tokenize; `dip 2` rejects them at parse-analysis time).
- **DIP104 / DIP115 / DIP144 unified on a single failure-route model** ([#134](https://github.com/2389-research/dippin-lang/issues/134)). All three "does this node have a failure destination?" lints now consult one helper, `nodeHasFailureRoute(w, n)` = *a bounded v1 `retry_target`/`fallback_target`* **or** *an `on fail` edge* — so a node that routes failure via an edge (the `dip 2` form) is recognized as covered exactly as a node that routes it via a v1 target field. DIP104 (unbounded retry), DIP115 (goal-gate fallback), and DIP144 (`on_failure` reachability) all read the same source of truth; the old per-rule `isOutcomeVar`/`isFailValue` duplication is gone. No rule fires differently on an unchanged v1 file — the model is a strict superset that additionally credits on-fail edges.

### Docs & tooling
- Documentation, site, embedded LLM spec, and editor surfaces swept for `dip 2` ([#134](https://github.com/2389-research/dippin-lang/issues/134)): `docs/{nodes,edges,cli,validation,llm-reference}.md`, `site/content/{language,cli,validation}.md`, and `site/static/skill.md` mark `retry_target`/`fallback_target` as **v1-only** (rejected under `dip 2` — use `loop`/`on fail` edges), document the real `fmt --migrate` transform (inline `# MIGRATION:` notes, exit `3`, `--migrate --check`), and reword DIP104/DIP115/DIP144 to the unified failure-route model. Embedded spec regenerated (`cmd/dippin/generated-spec.md`).

### Runtime pairing (requires an enforcing runtime)
- `dip 2` is a pure authoring/analysis change on the dippin side — the migration and the version-gated rejection are enforced entirely at parse/lint time, with no new IR the engine must interpret (a migrated `dip 2` graph routes failure through ordinary `on fail` / `loop` edges the engine already resolves). Downstream adoption is a documentation/version-support concern (recognizing the `dip 2` header), not a new runtime code path (per `never-gate-dippin-on-tracker`).

## [v0.45.0] — 2026-07-08

### Added
- **`DIP152` — coverage-aware `marker_grep` lint** ([#156](https://github.com/2389-research/dippin-lang/issues/156), [#180](https://github.com/2389-research/dippin-lang/pull/180)). Warns when a tool node's `marker_grep` enumerates a literal marker that no outgoing edge routes and that no section `else ->` default or unconditional fallback covers — that marker would be emitted at runtime with nowhere to go, previously a silent hole (any `marker_grep` blanket-exempted the source node from DIP101/DIP102). DIP152 catches it precisely, naming the uncovered markers. **Additive**: DIP101/DIP102 keep their marker exemption, so a gap gets exactly one warning. **Ultra-conservative — no false positives**: only recognizable finite literal alternations (`^(a|b|c)$` or a bare literal) are coverage-checked; any regex with a metacharacter, an empty branch, or a non-full-span group stays blanket-exempt, and any compound (`or`), negated (`!=` / `not`), other-variable, or unconditional edge makes the node safe. The exit node and edge-less dead-ends are handled correctly (the simulator dead-ends an edge-less node before any `else` routing, so `else` cannot cover it). Reuses `ir.ExtractEqualityCondition`; a `TestLintExamples` guard keeps the example suite covered. Brings the catalog to 62 codes (DIP101–DIP152).

### Changed
- **`dippin simulate` + the all-paths enumerator now honor the section `else ->` default** ([#158](https://github.com/2389-research/dippin-lang/issues/158), [#178](https://github.com/2389-research/dippin-lang/pull/178)). `ir.Workflow.ElseTarget` lives outside `Edges` and `EdgesFrom` doesn't expose it, so the built-in simulator and enumerator never traversed the `else` funnel default (they fell back to `edges[0]`, the happy-path heuristic). Now the simulator routes an unmatched node — including a concrete `--scenario` value no guard covers — to `ElseTarget`, mirroring the engine; the enumerator emits the `else` branch as a distinct path, gated on non-exhaustive guards (a declared-exhaustive success/fail set or complete partition does not enumerate an unreachable `else`). The static analyses (DIP003/004/101/102/105) were already `else`-aware since #157; this finishes the two execution-simulation consumers. Enabled by moving the edge-condition exhaustiveness helper to the shared `ir` tier as `ir.EdgesExhaustive` (validator and simulate share one source of truth — no import cycle, no drift).

### Fixed
- **Spec self-contradiction on block-form `parallel`** ([#174](https://github.com/2389-research/dippin-lang/issues/174)) — the `parallel`/`fan_in` section said "Inline syntax only", contradicting the supported block form (`parallel P` / `branch:` lines with per-branch `model`/`provider`/`fidelity`/`tool_access`/`writable_paths`/`last_response_truncate` overrides). Reworded to document both forms; the `fan_in`/DIP007 pairing applies to both.
- **`brew install` formula name** in the site + docs install instructions ([#179](https://github.com/2389-research/dippin-lang/pull/179), thanks [@lra](https://github.com/lra)) — the tap ships the formula as `dippin-lang`, so `brew install 2389-research/tap/dippin` failed; corrected all five occurrences.

## [v0.44.0] — 2026-07-01

### Added
- **Opt-in file-shipping pack mode** — `dippin pack --no-inline [--include <file-or-dir>...]` ([#73](https://github.com/2389-research/dippin-lang/issues/73), [#176](https://github.com/2389-research/dippin-lang/pull/176)). By default `dippin pack` inlines every `command_file:` / `prompt_file:` / `system_prompt_file:` body into the packed `.dip`, so a script that sources a sibling by relative path (e.g. `. "${graph.workflow_dir}/scripts/lib/bootstrap.sh"`) has no file on disk in a packed run. `--no-inline` instead **ships** those directive targets as separate bundle entries under `workflows/` and **keeps** the `*_file:` directives, so `parser.ResolveFileDirectives` resolves against the extracted tree exactly as in a source-tree run. `--include <path>` (repeatable, requires `--no-inline`) declares sibling assets the parser can't discover statically — files referenced only from inside shell bodies — as a single file or a whole directory (recursive, via a symlink-refusing `WalkDir`); an `--include` that resolves to a `.dip` is an error. Assets ship under `workflows/` as siblings of their referencing `.dip` (so relative resolution is identical source-vs-packed) and the bundle is stamped `format_version 2`. The library surface is `dipx.Pack(ctx, entry, w, dipx.PackOptions{NoInline, Include})`; the zero value reproduces the default inline behavior byte-for-byte (still `format_version 1`), and an older reader rejects a v2 bundle loudly with `ErrUnsupportedFormatVersion`. Path-safety is preserved throughout: every collected/included candidate is re-validated (containment + symlink refusal), directive targets are validated against their declaring `.dip`'s own directory (rejecting absolute paths and parent-escapes, matching source/inline behavior), assets ship `0644` (no mode carried), reserved `manifest.json`/`manifest.sig` names can't be assets, and pack-time file-count + total-size caps mirror the open-time caps. Docs: `docs/integration.md`, the site `cli` page. Motivating use-case: collapsing many byte-identical inlined bootstrap preambles into one sourced helper (pipelines#49).

### Changed
- **Website & docs freshness sweep** — updated the `cli` page's `pack` reference for `--no-inline` / `--include` and `format_version 2`; corrected the LSP hover/autocomplete capability descriptions (hover shows model/provider/command by node kind, not a prompt preview; completion offers node IDs + field names, not keywords) and the glossary's `human` mode set (`choice`, `freeform`, `interview`, `yes_no` — `confirm` was never a mode); refreshed the editor-setup blog's diagnostic range to DIP001–DIP010 + DIP101–DIP151; and added the missing agent/tool field keywords (`writable_paths`, `last_response_truncate`, `response_format`, `response_schema`, `compaction_threshold`, `backend`, `working_dir`, `cmd_timeout`) to the VS Code TextMate grammar.

### Runtime pairing (requires an enforcing runtime)
- `--no-inline` produces a bundle whose extracted tree mirrors the source tree under `workflows/`. For a packed run to resolve `${graph.workflow_dir}/<relpath>` identically to a source-tree run, the runtime must seed `graph.workflow_dir` to the unpacked-bundle root (tracker#430). dippin ships this independently and is not gated on the runtime (per `never-gate-dippin-on-tracker`); until the paired runtime lands, the dippin-side deliverable is purely that the extracted bundle byte-mirrors the source tree.

## [v0.43.0] — 2026-06-22

### Added
- Section-level `else -> <node>` edge default — the error-funnel collapse ([#154](https://github.com/2389-research/dippin-lang/issues/154), implements the decision in [#135](https://github.com/2389-research/dippin-lang/issues/135) / epic [#127](https://github.com/2389-research/dippin-lang/issues/127)). One line at the bottom of the `edges` block, `else -> Cleanup`, is the graph's **success-side default destination**: any node whose guard edges all fail to match, and which has no explicit unconditional edge of its own, routes there. It collapses the dominant line-count cost in real workflows — the per-marker failure funnel — from ~20 hand-routed edges to one. Stored as a graph-level `ir.Workflow.ElseTarget` (not a synthetic `ir.Edge`, so the edge list and the tools that read it — `EdgesFrom`, DOT export, edge-keyed analyses — are untouched); the validator seeds the else target into its reachability analysis explicitly (so DIP004 treats it as reachable and DIP105 sees it as a success route). At most one `else` per block (a second is a parse error). **Lint**: DIP102 (no default edge) and DIP101 (reachable only via conditional edges) treat a node covered by `else` as having its unconditional fallback, so they stand down. **Structural validation**: the `else` target must exist (DIP003) and is treated as reachable (DIP004), mirroring `defaults.on_failure`. **Success-side only** — `else` never intercepts a genuine node *failure* (tool exit ≠ 0, agent error), which routes via the failure cascade (`on fail` → `defaults.on_failure`); the two are distinct, non-competing channels. Formatter emits `else` last in the block; tree-sitter grammar + corpus updated (regen via `npx tree-sitter generate`). Docs: `docs/edges.md`, `docs/GRAMMAR.ebnf`, `docs/llm-reference.md`.

### Changed
- **`dippin explain` / diagnostic-text accuracy** ([#168](https://github.com/2389-research/dippin-lang/pull/168), [#172](https://github.com/2389-research/dippin-lang/pull/172)) — corrected diagnostic output that misdescribed behavior or suggested invalid values: DIP106 reworded "undefined variable" → "unrecognized variable reference" (it's a namespace/shape check, not an upstream-write check); DIP109 retitled "duplicate subgraph reference" with the real trigger (two `subgraph` nodes sharing a `ref:`) and a remediation that actually silences it (different `ref:` / consolidate, not "distinct params"); DIP113/DIP114/DIP116 fix-hints corrected to the validator's actually-valid value sets (retry-policy, fidelity, on_resume); explain examples now use valid `.dip` syntax — keyword edge form (`when ctx.outcome = success`) instead of the rejected bracket form (`[success]`), and `#` comments instead of `//`. Also fixed a stale `simulate` doc comment ("fail" is never emitted). Detection/text only — no rule-firing behavior changed.
- **Documentation, website & editor-grammar accuracy sync** ([#171](https://github.com/2389-research/dippin-lang/pull/171)) — content-only sweep bringing every hand-maintained surface in line with the current language: the `language`/`validation`/`analysis`/`architecture`/`cli`/`testing` pages, glossary, `skill.md` + `llm-reference` (+ regenerated spec), and the tree-sitter/Zed/VS Code grammars. Added `else` syntax highlighting (and refreshed the stale Zed grammar pin), corrected DIP ranges/counts to DIP101–DIP151, documented previously-missing attributes (`*_file:` directives, `tool_access`, `writable_paths`, parallel block form, the four human modes), and fixed numerous stale command/flag/diagnostic descriptions.

### Internal
- **Tech-debt cleanup batch** ([#164](https://github.com/2389-research/dippin-lang/pull/164), [#165](https://github.com/2389-research/dippin-lang/pull/165), [#166](https://github.com/2389-research/dippin-lang/pull/166), [#167](https://github.com/2389-research/dippin-lang/pull/167), [#169](https://github.com/2389-research/dippin-lang/pull/169), [#170](https://github.com/2389-research/dippin-lang/pull/170)) — removed dead code and duplication across the toolchain: test-only/zero-reference exported functions, hand-populated `Condition.Parsed` test fixtures (a DIP101 anti-pattern), unreachable defensive branches, de-facto-constant config knobs / single-use params / hand-rolled stdlib reimplementations, and duplicated simulate event-construction. No behavior change; net ~1000 lines removed.

### Runtime pairing (requires an enforcing runtime)
- `else` adds one terminal step to the engine's routing resolution: when a node's guards do not match and it has no unconditional edge of its own, route to the section `else` target (success side). The failure cascade is unchanged and runs in its own channel. Inert on the dippin side until the paired runtime reads `ir.Workflow.ElseTarget` (per `never-gate-dippin-on-tracker`).

## [v0.42.0] — 2026-06-16

Routing-syntax **Phase 0 complete** — the human-gate `choice:` routing key and the `weight:` soft-deprecation lint. All changes are dippin-side only; no behavior is gated on a paired runtime (per `never-gate-dippin-on-tracker`). Non-breaking: every valid v1 `.dip` file parses, validates, and formats unchanged — `weight:` still parses (carry-only) and merely warns now.

### Added
- `choice:` edge attribute — a dedicated human-gate routing key that disambiguates the routing intent from a display-only `label:` ([#130](https://github.com/2389-research/dippin-lang/issues/130), [#151](https://github.com/2389-research/dippin-lang/pull/151)). **Carry-only**: dippin parses, validates, formats, exports (a distinct DOT attribute), and round-trips `choice:` (`ir.Edge.Choice`) but assigns it no execution semantics — a paired runtime interprets it. Mirrors the carry-only `override:` shape (#125). `DIP150` (Hint) flags a label-routing human gate (mode `choice`, `yes_no`, or the unset default) whose outgoing edge sets a non-empty `label:` but no `choice:` — in Phase 0 that label doubles as the load-bearing routing key, so the hint suggests marking it explicitly; `freeform`/`interview` gates (which route by text/answers, not labels) are exempt. **Phase-0 source-compatibility (no `dip 2` bump)**: `label:` is NOT made display-only — `choice:` wins as the routing key when present, and when absent `label:` still serves as the routing key, so existing human-choice workflows route unchanged. Editor surfaces (tree-sitter grammar + highlights, zed, vscode tmLanguage) updated.
- `DIP151` (Warning) — soft-deprecation of the `weight:` edge attribute ([#131](https://github.com/2389-research/dippin-lang/issues/131), [#152](https://github.com/2389-research/dippin-lang/pull/152)). `weight:` was tier 4 of the 5-level routing cascade — a speculative priority hint — but the cascade never consults it and no real `.dip` workflow uses it. The lint fires one diagnostic per edge with a non-zero `weight:`, steering authors toward conditions / `on` / a single unconditional fallback. **Carry-only / no breakage**: `weight:` still parses exactly as before — dippin only now warns. Removing the keyword and the cascade tier is deferred to `dip 2` ([#134](https://github.com/2389-research/dippin-lang/issues/134)). Surfaces in lint/check/watch/doctor.

### Runtime pairing (requires an enforcing runtime)
- `choice:` unblocks the tracker's `convertEdge` follow-up: once this is tagged, the engine prefers `ir.Edge.Choice` over `Label` as the human-gate matching key (falling back to `Label` when `choice:` is absent). Inert on the dippin side until the paired runtime reads it.

### Docs
- Website, editor, and grammar surfaces resynced to the v0.41.0 language surface ([#150](https://github.com/2389-research/dippin-lang/pull/150)) — content-only, no behavior change: `site/content/{validation,language,cli}.md` counts/ranges and the DIP148/DIP149 entries, `on`/`loop`/`dip N`/single-quote notes, tree-sitter/zed/vscode keyword highlighting, and the single-quoted `STRING` alternative in `GRAMMAR.ebnf`.

## [v0.41.0] — 2026-06-16

Routing-syntax Phase 0 plus the `dip 2` version foundation. All changes are dippin-side only — no behavior is gated on a paired runtime (per `never-gate-dippin-on-tracker`). Non-breaking: every valid v1 `.dip` file parses, validates, and formats unchanged (the one intentional rejection is a previously-silently-dropped unknown edge attribute, now diagnosed — see below).

### Added
- `on <token>` edge routing shorthand — sugar for an equality test against a node's natural outcome channel ([#128](https://github.com/2389-research/dippin-lang/issues/128), [#140](https://github.com/2389-research/dippin-lang/pull/140)). `A -> B on success` desugars to `when ctx.outcome = success` for an agent source, and to `when ctx.tool_marker = <token>` for a tool source that declares `marker_grep`; nodes without a defined outcome channel (human gates, `conditional`, tool without `marker_grep`) get a located diagnostic suggesting `when`. **IR-preserving** — `on X` produces the identical `ir.Condition` as the equivalent `when`, so DIP102/DIP103, DOT export, and every other edge consumer reason over it unchanged. The channel is a single source of truth (`ir.Node.OutcomeChannel()`) shared by parser and formatter, so what the parser accepts is exactly what the formatter emits; `dippin fmt` rewrites an eligible `when` to `on` (shape-recognition, no new IR field) and is idempotent.
- `loop` keyword — a scannable, value-less synonym for `restart: true` on back-edges ([#129](https://github.com/2389-research/dippin-lang/issues/129), [#143](https://github.com/2389-research/dippin-lang/pull/143)). A back-edge is the heaviest-semantics construct in the language (restart counter, `max_restarts` failure, downstream completion-state reset), so it earns a word: `QualityGate -> Refactor on fail loop`. Sets the same `ir.Edge.Restart` field — **carried-only, no new IR, no semantics change**; `restart: true` still parses and `dippin fmt` rewrites it to the canonical `loop`. DOT export (dashed) and DIP005 (cycle exemption) work for free via the shared field. Because `loop` is value-less it terminates a preceding condition unconditionally, so a bare `loop` is reserved on a condition RHS — write `when ctx.x = "loop"` (quoted) for the literal.
- `DIP149` (Warning) — ambiguous routing: a node with ≥2 unconditional outgoing forward edges, whose winner is decided only by the cascade's lexical tiebreak on target node IDs (so renaming a target can silently change which edge fires) ([#132](https://github.com/2389-research/dippin-lang/issues/132), [#144](https://github.com/2389-research/dippin-lang/pull/144)). Conservative (favors false negatives): restart/`loop` back-edges are excluded; a guarded edge plus a single unconditional fallback is the intended pattern and is not flagged; duplicate same-var/same-value guards stay covered by DIP103. `parallel` fan-out (all edges fire) and `human` choice gates (the human picks) are exempt; `fan_in` is **not** exempt — its out-edges route by the ordinary cascade. Surfaces in lint/check/watch/doctor. Sets up the Phase 1 cascade-collapse where the lexical tier is removed.
- `dip N` format-version declaration + `fmt --migrate` scaffolding ([#133](https://github.com/2389-research/dippin-lang/issues/133), [#148](https://github.com/2389-research/dippin-lang/pull/148)). A leading line-1 `dip N` sets the parsed format version (default 1), wiring the previously-dead `ir.Workflow.Version`; the formatter emits the line **only for N>1**, so existing v1 files stay declaration-free and round-trip byte-identical. `fmt --migrate` lands as a no-op v1→v1 identity pass reusing the existing parse/format machinery. **Foundation only** — `p.version` is now reachable from `parse_edges.go`; the v1→v2 transforms (edges own their destinations) ship with [#134](https://github.com/2389-research/dippin-lang/issues/134) and are explicitly NOT in this release.
- Single-quoted strings accepted in edge `label:` attributes and `when` conditions ([#146](https://github.com/2389-research/dippin-lang/pull/146), closes [#120](https://github.com/2389-research/dippin-lang/issues/120)). Extends the v0.39.0 single-quote scalar support to the edge token-stream surface that #117 deliberately deferred. YAML-style (surrounding quotes stripped, `''` → `'`, backslashes literal); the token-stream comment stripper now treats `'` as a delimiter only at a token-start boundary, so a prose apostrophe (`it's`) still strips its trailing comment while a `#` inside a quoted token stays literal. Normalizes to double-quoted output on `fmt`.
- Edge-attribute robustness — unknown edge attributes are now diagnosed instead of silently swallowed, and keyword termination is colon-gated ([#126](https://github.com/2389-research/dippin-lang/issues/126), [#139](https://github.com/2389-research/dippin-lang/pull/139)). `A -> B bogus: true` previously parsed with `bogus` dropped on the floor (the same inert-attribute class that hid the pre-#124 `override`); it now emits exactly one located diagnostic. Colon-gating means `label`/`weight`/`restart`/`override` terminate a `when` condition only when followed by `:`, so a bare attribute word as a condition's right-hand value (`when ctx.reason = override`) is no longer truncated. Source-compatible but **not strictly non-breaking** — a file with a typo'd or runtime-only edge attribute that parsed before is now rejected; this is the intended hardening that makes the Phase 0 keyword surface safe.

### Fixed
- Extra/preview models registered via `--extra-models` are now scoped out of the global model registry ([#105](https://github.com/2389-research/dippin-lang/issues/105), [#145](https://github.com/2389-research/dippin-lang/pull/145)). `validator.RegisterExtraModels` mutated a package-level catalog, so models from one `lint`/`doctor` run leaked into every later validate/doctor/lint call in the same process (making `TestCmdLint_ExtraModels_SuppressesDIP108` non-idempotent under `-count>1`). Replaced with a pure `ParseExtraModels` threaded through a scoped `validator.Options`/`LintWithOptions` (and `doctor.DiagnoseWithOptions`); `Lint`/`Diagnose` keep their signatures and delegate with empty options.
- `startswith` edges now count as covering printf-format markers in coverage analysis ([#141](https://github.com/2389-research/dippin-lang/issues/141), [#147](https://github.com/2389-research/dippin-lang/pull/147)). `dippin coverage` (which drives `dippin doctor`'s "uncovered outputs" check) did exact-match of each tool-output literal against edge-condition values, discarding the operator — so a `when ctx.tool_stdout startswith citations-fail` edge never covered the printf-format output `citations-fail-%s`, capping doctor grades on otherwise-clean files. The operator is now carried through and an output is covered if some `startswith` term is a prefix of it (conservative — only `startswith` is extended). The public `EdgeConditions []string` report field is unchanged.
- Playground (dippin.org): corrected the lint JSON shape and initial syntax highlighting ([#142](https://github.com/2389-research/dippin-lang/pull/142)). `cmd/wasm` marshalled the raw `validator.Diagnostic` (PascalCase keys, numeric severity), so the playground JS read `undefined` for `code`/`severity`/`message` and crashed on the default workflow's DIP144 — mislabeled as "Failed to load WASM"; diagnostics now map to the camelCase/string-severity DTO matching the CLI contract. Separately, the initial highlight is deferred to `DOMContentLoaded` so the default source is no longer black-on-black until the first keystroke.

## [v0.40.0] — 2026-06-11

Carry + detection work pairing with downstream runtimes — no dippin behavior is gated on runtime readiness (per `never-gate-dippin-on-tracker`). Both attributes are carried through every `.dip` path; their semantics are owned by a paired runtime.

### Added
- `last_response_truncate` (agent node + `parallel` branch override, integer) — a character-count cap on the auto-injected `${ctx.last_response}` handoff, a **carry-only chain-attack mitigation** ([#56](https://github.com/2389-research/dippin-lang/issues/56), [#123](https://github.com/2389-research/dippin-lang/pull/123)). Carried through parse → IR → formatter → DOT export → migrate; `DIP148` (Warning) flags a negative value (`0`/unset = disabled). Carried, not interpreted — a paired runtime performs the truncation. Mitigates the `${ctx.last_response}` auto-injection vector of #56 when an enforcing runtime honors this field; `DIP147` (explicit-key handoff) is unaffected and there is no DIP147↔truncate interaction.
- `override: true` / `override: false` edge attribute → `ir.Edge.Override` (bool) — **carried, not interpreted** (the same policy as `restart:`) ([#124](https://github.com/2389-research/dippin-lang/issues/124)). Parsed in any order alongside `when` / `label` / `weight` / `restart` (no longer silently swallowed), and round-tripped through the `.dip` formatter, DOT export, and migrate. No dippin-side semantics — a paired runtime maps it onward and owns the rule that `override` edges originate from `human` nodes. Same shape as the `params:` carry shipped in v0.39.0 ([#113](https://github.com/2389-research/dippin-lang/pull/113)).

### Runtime pairing (requires an enforcing runtime)
- Edge `override` unblocks the workflow-author syntax for [2389-research/tracker#271](https://github.com/2389-research/tracker/issues/271), whose engine half (the `validation_overridden` terminal status, `pipeline.Edge.Override`, sticky checkpoint persistence, `--fail-on-override`, lint TRK102) shipped in tracker v0.35.0. Once this is tagged, tracker maps `ir.Edge.Override` → `pipeline.Edge.Override` in `convertEdge`, adds the human-origin validation (tracker spec D11), and migrates the catalogued override edges (tracker#271 finisher).
- `last_response_truncate` is carried + linted by dippin but the truncation is performed by the runtime — inert until a paired runtime reads it.

## [v0.39.0] — 2026-06-10

Detection, carry, and correctness work — no paired runtime release required (per `never-gate-dippin-on-tracker`). `DIP147` and the `params:` carry are dippin-side detection / round-trip only; the single-quote fix is a parser correctness fix; the catalog refresh is data.

### Added
- `DIP147` (Hint) — **chain-attack detection**: a restricted agent's output is laundered through an **explicitly-keyed** context handoff into a downstream tool-bearing agent, re-granting capability the upstream node was denied ([#56](https://github.com/2389-research/dippin-lang/issues/56), [#115](https://github.com/2389-research/dippin-lang/pull/115)). Covers the explicit-key handoff vector only — a **partial** #56. The other laundering vectors remain open follow-ups: `${ctx.last_response}` auto-injection (and a `last_response_truncate:` mitigation), attribute, cross-file, and branch-override. Detection, not runtime enforcement.
- `params:` on `parallel` / `fan_in` nodes are now carried through parse → IR → formatter round-trip ([#110](https://github.com/2389-research/dippin-lang/issues/110), [#113](https://github.com/2389-research/dippin-lang/pull/113)). Formatter round-trip only (not DOT export or migrate); a paired runtime owns any fan-in policy semantics.
- **Model catalog + pricing refresh** for Anthropic's 2026-05/06 releases ([#116](https://github.com/2389-research/dippin-lang/issues/116)): `claude-opus-4-8`, `claude-fable-5`, `claude-mythos-5`, and `claude-mythos-preview` are recognized models (clearing a spurious `DIP108`), with `claude-opus-4-8` ($5/$25), `claude-fable-5` ($10/$50), and `claude-mythos-5` ($10/$50) priced so `dippin cost` can estimate them. Deprecation metadata refreshed (`claude-opus-4-1` retires 2026-08-05; `claude-opus-4-0` migration target → `claude-opus-4-8`). The richer cache/batch/fast-mode pricing schema is an explicit non-goal; tracker-side runtime support is a cross-repo follow-up.

### Fixed
- **Single-quoted scalar values are no longer corrupted** ([#114](https://github.com/2389-research/dippin-lang/issues/114)). The parser stripped only double quotes, so `marker_grep: '^(a|b)$'` was stored with the literal `'` chars — and the formatter then re-wrapped it in double quotes — silently breaking the runtime regex. Single quotes are now YAML-style (surrounding quotes stripped, `''` → `'`, backslashes literal), and a `#` inside `'...'` is no longer mistaken for a trailing comment. Fixed in the parser so every path (lint / pack / format / DOT / migrate / runtime) benefits; the tree-sitter grammar admits single-quoted strings too.

## [v0.38.0] — 2026-06-09

Cross-file `tool_access` advisory completeness. Both changes are **dippin-side detection only** — they refine the `DIP146`/`DIP143` cross-file analysis shipped in v0.37.0 and require no paired runtime release (per `never-gate-dippin-on-tracker`, no dippin behavior is gated on runtime readiness).

### Added
- `DIP143` (Hint) **deep cross-file advisory** — the native `dippin lint` cross-file pass now emits `DIP143` for a **partial-audit** or **unresolvable** child found at **depth ≥ 1** (behind an already-audited intermediate), gated by path `tool_access` intent ([#102](https://github.com/2389-research/dippin-lang/issues/102)). Previously this fallback came only from the entry-file lint, so a partial/unresolvable gap more than one hop deep was silent — DIP146 already traversed transitively for *zero-intent* children, but the partial/unresolvable advisory did not. Reuses `DIP143` (no new code); the message distinguishes partial-audit (resolved — a tool-bearing agent lacks its own `tool_access`) from unresolvable (missing / unparseable / refused). Entry boundaries (depth 0) keep their existing `validator.Lint` `DIP143`. Emitted from `dippin lint`/`check`/`watch`; no wasm/validator-detection change.

### Fixed
- **Intent-aware cross-file re-walk** — a child reached **first** via a path with no `tool_access` intent and **later** via a restricting path was not re-walked under the restricting intent, so a zero-intent / partial-audit / unresolvable **grandchild** below the shared child was silently missed ([#109](https://github.com/2389-research/dippin-lang/issues/109)). The recursion guard is now intent-aware (re-walk once on a no-intent → intent upgrade; bounded and cycle-safe, with the entry file never re-walked). Pre-existing in `DIP146` since its introduction ([#89](https://github.com/2389-research/dippin-lang/issues/89)) and inherited by the #102 deep advisory; both now catch the mixed-intent shared-child shape regardless of traversal order.

### Docs
- Documented the deep `DIP143` advisory and the intent-aware cross-file traversal in the validation reference (`docs/validation.md` + the website validation page).

## [v0.37.0] — 2026-06-08

Mixed release. The cross-file `tool_access` detection + resolver hardening and the DIP010 edge-condition error are **dippin-side detection only** (no paired runtime release required). The graph-level `on_failure` route and the declarable budget guards are an **authoring surface dippin carries + lints while a paired runtime enforces** — inert until the runtime reads them (per `never-gate-dippin-on-tracker`, no dippin behavior is gated on runtime readiness).

### Added
- `DIP146` (Hint) — cross-file `tool_access` gap: a `manager_loop` (`subgraph_ref`) or `subgraph` (`ref`) boundary delegates to a child `.dip` whose agents re-grant tools a parent on the path restricted ([#89](https://github.com/2389-research/dippin-lang/issues/89)). Completes the `DIP143` (#59) arc by resolving and classifying the referenced child instead of only flagging the boundary: zero-intent child → DIP146 (DIP143 superseded); full-restrict or agent-less → silent (confirmed safe); partial-audit or unresolved → DIP143 retained ("unknown" never reads as "checked & safe"). Full transitive DFS with `EvalSymlinks`-keyed cycle termination and a depth cap. Emitted from the native CLI pass — `dippin lint`/`check`/`watch` apply it with DIP143 supersession; `validate` stays structural-only. Detection, not runtime enforcement.
- `DIP010` (Error) — an edge `when` condition that cannot be parsed (e.g. an unknown operator, or a tool-node field like `marker_grep` used in operator position) ([#98](https://github.com/2389-research/dippin-lang/issues/98)). Previously the parse error was discarded by `Lint()` and `EnsureConditionsParsed` stopped at the first bad edge, so `validate`/`lint`/`check`/`doctor` greenlit a workflow that hard-fails at `dippin simulate` — and every edge after the first bad one silently lost its AST-dependent lints (DIP103/120/121/122). DIP010 is emitted from `validator.Validate()`, so every command path catches it, and parsing now continues past failures (one diagnostic per bad edge; later edges keep getting linted). Edge conditions only; `manager_loop` node conditions are a separate follow-up.
- `on_failure: <NodeID>` — a graph-level (`defaults:`) catch-all failure route ([#92](https://github.com/2389-research/dippin-lang/issues/92)). Structurally validated (DIP003 existence with a "did you mean?" suggestion; the target is seeded into DIP004 reachability so a recovery node reachable only via the catch-all is not falsely flagged unreachable) and round-tripped through the `.dip` formatter, DOT export, and migrate.
- `DIP144` (Warning) — an agent node has no failure route ([#93](https://github.com/2389-research/dippin-lang/issues/93)). Suppressed by an explicit `ctx.outcome = fail|failure` edge, `fallback_target`, bounded retry (`retry_target` + `max_retries > 0`), or a graph `on_failure`; an unconditional/success edge does not suppress it.
- `stall_timeout` (graph `defaults:`, duration) — abort/route when no forward progress is made for a wall-clock span ([#94](https://github.com/2389-research/dippin-lang/issues/94)); `0`/unset = disabled. Round-trips through parser → formatter → DOT export → migrate (this also closed a pre-existing DOT-export bug that silently dropped every budget ceiling). `max_turns` now has defined exhaustion semantics: outcome = `fail`, routed through the failure cascade (docs-only; no new action field).
- `DIP145` (Warning) — a graph budget default is negative ([#94](https://github.com/2389-research/dippin-lang/issues/94)); `0` = unset = no warning.
- `examples/on_failure_route.dip` and `examples/budget_guards.dip` demonstrators.

### Hardened
- The DIP146 cross-file resolver now refuses symlinked and root-escaping child `.dip` refs, reaching parity with the pack walker ([#100](https://github.com/2389-research/dippin-lang/issues/100)). Fail-soft: a refused child is treated as unresolvable (DIP143 retained), so linting never hard-errors or aborts. Reuses dipx's `ReadNoFollowSymlinks` (leaf + ancestor-directory symlink refusal) and a fixed entry-file-directory containment root; absolute refs are re-rooted under that root rather than read out of tree. Lstat-based parity with the pack walker — no `O_NOFOLLOW`/build-tag complexity and no wasm impact. Detection parity, not runtime enforcement.

### Changed
- `dippin doctor` now exits non-zero when the report contains errors (e.g. DIP010 or any structural DIP001–DIP010), rather than always exiting 0. The report still renders in full; only the exit code changes, so doctor no longer greenlights a workflow that cannot execute.

### Runtime pairing (requires an enforcing runtime)
- `on_failure`, `stall_timeout`, and `max_turns` exhaustion are carried + linted by dippin but **enforced by the runtime** — inert until a paired runtime reads them. The runtime owns the failure-cascade ordering: matching fail edge → bounded node retry → node `fallback_target` → graph `on_failure` → halt.

### Docs
- Audited and refreshed all reference docs to the current language surface ([#97](https://github.com/2389-research/dippin-lang/pull/97)).

## [v0.36.0] — 2026-06-03

Dippin-side only — no paired runtime release required.

### Added
- `DIP143` (Hint) — a `manager_loop` (`subgraph_ref`) or `subgraph` (`ref`) node references a child `.dip` that does not inherit the parent workflow's `tool_access` restrictions ([#59](https://github.com/2389-research/dippin-lang/issues/59)). Fires only when the workflow declares `tool_access` intent (any non-empty value on an agent or parallel branch) **and** references an external subgraph, reminding the author to give the child's agents their own `tool_access`. `tool_access` is per-node and does not cross a file boundary. A direct self-reference (a node whose ref resolves to its own source file) is not flagged. The lint never parses the child file (the validator may not import the parser); real cross-file effective-access enforcement, plus transitive cross-file cycles, are deferred to [#89](https://github.com/2389-research/dippin-lang/issues/89).

### Hardened
- The `@file` directive resolver is hardened against a leaf TOCTOU race ([#79](https://github.com/2389-research/dippin-lang/issues/79)). The path was validated and then read via a separate `os.ReadFile`, so a concurrent symlink/rename swap of the final component could redirect the read outside `baseDir`, bypassing the [#67](https://github.com/2389-research/dippin-lang/issues/67)/[#77](https://github.com/2389-research/dippin-lang/issues/77) containment under a race. Now a single open-once path — `os.OpenFile(O_RDONLY|O_NOFOLLOW)` → `f.Stat()` (fstat the fd) → containment check → `io.ReadAll` — operates entirely on one fd, closing the check-to-read race. On Unix, `O_NOFOLLOW` makes leaf-symlink rejection atomic (`ELOOP`); the resolved path is never leaked in the error. The residual parent-directory swap race is documented as out of scope.

### Docs
- Clarified that `tool_access` is node-scoped — it constrains the executor of a single node and does not taint downstream nodes — resolving [#57](https://github.com/2389-research/dippin-lang/issues/57) via documentation rather than a graph-topology lint ([#87](https://github.com/2389-research/dippin-lang/pull/87)).

### Changed
- Internal: removed the consumer product name "tracker" from dippin sources — dippin is consumer-agnostic ([#85](https://github.com/2389-research/dippin-lang/pull/85)).

## [v0.35.0] — 2026-06-02

### Added
- `writable_paths:` — a comma-separated glob list on agent nodes and parallel branches that bounds where an agent's tools may write (e.g. `writable_paths: workspace/**, .ai/sprints/**`) ([#75](https://github.com/2389-research/dippin-lang/issues/75)). Absent = unbounded; a per-branch empty value inherits the target agent's (never resets to unbounded). dippin **carries + lints**; the runtime **enforces** an fs-level write jail on the `native` backend (covering `Write`/`Edit`/`ApplyPatch`, **`Bash` and any process Bash spawns**), resolved against an immutable session root. Distinct from the advisory `writes:` field (context keys). Carried through parse → IR → format → DOT export → migrate.
- `DIP141` — `writable_paths` nullified by `tool_access: none` (dead config; nothing left to bound).
- `DIP142` — unsafe `writable_paths` entry (absolute / `~` / Windows-drive / `..`-escape / brace mis-split). Author-clarity lint; the runtime fs-jail is the real boundary.
- `examples/agent_writable_paths.dip` demonstrating the five motivating failure/recovery-recorder shapes.

### Runtime pairing (requires an enforcing runtime)
- Enforcement ships in a coordinated runtime release: native fs-jail incl. Bash children, runtime symlink-chain resolution, immutable session-root anchor (`working_dir`/`Params` cannot relocate it), `Params`/`working_dir` bypass defense, and a single-turn red-team suite.
- **Fail-closed.** A present-but-empty or comma-only `writable_paths:` is rejected by `dippin validate`/`pack` (parse error). A malformed or **runtime-unrecognized** value → the runtime denies all writes or refuses to start, never unbounded. On `claude-code`/`acp`, the runtime **refuses to start** when `writable_paths` is set (native-only enforcement) — never a prompt-level pretend-jail.
- **Version-skew is a safety requirement, not a suggestion:** a runtime that does not enforce `writable_paths` must refuse to start rather than run unbounded. Pin an enforcing runtime, never `@latest`.

### Notes
- The primitive bounds write *location*, not network (`curl`/`cargo fetch`), reads/exfil-by-read, or *content* within an allowed path (a `workspace/**` grant can still poison `workspace/Cargo.toml`). Chain laundering is tracked in [#56](https://github.com/2389-research/dippin-lang/issues/56).
- Sequenced follow-ups: [#55](https://github.com/2389-research/dippin-lang/issues/55) (tool-name allowlists — now an orthogonal axis), [#53](https://github.com/2389-research/dippin-lang/issues/53) (defaults cascade).
- Design spec: `docs/superpowers/specs/2026-05-29-issue-75-writable-paths-design.md`.

## [v0.34.0] — 2026-05-28

### Added
- `prompt_file: <path>` and `system_prompt_file: <path>` directives on agent nodes ([#65](https://github.com/2389-research/dippin-lang/issues/65)). Symmetric extension of v0.33.0's `command_file:`; same path-relative-to-`.dip` rules and security cap. `dippin pack` inlines the content, so no runtime coordination is required.
- `examples/external_prompts.dip` demonstrating both directives.

### Fixed
- `dippin fmt --write` previously stripped inline `system_prompt:` from agent nodes on save (latent since the field was introduced). If you've been re-adding `system_prompt:` after running fmt, this fix is for you.

### Notes
- Parser stays pure (no FS I/O); resolver runs at CLI entry points only. LSP and WASM consumers see the unresolved IR view (`*File` set, content empty), which is the correct view for those contexts.
- Per-node only; `defaults agent` does not currently support `prompt` / `system_prompt` fields, so file-form-in-defaults requires a larger design — filed as [#72](https://github.com/2389-research/dippin-lang/issues/72).
- Bundled-files `.dipx` redesign considered and explicitly deferred so v0.34 stays dippin-only; filed as [#73](https://github.com/2389-research/dippin-lang/issues/73).

## [v0.33.0] — 2026-05-27

New `command_file:` directive on tool nodes replaces inline `command:` heredocs with external file references. Solves the heredoc-bloat pattern seen in long runtime workflows. Dippin-only release — no runtime coordination required (the runtime reads inlined `Command` from `.dipx` bundles unchanged).

### Added
- `command_file: <path>` directive on tool nodes. Path is relative to the `.dip` source directory.
- `parser.ResolveFileDirectives(w, baseDir)` — separate-pass file loader called by CLI entry points. Parser itself stays pure.
- Path security in the resolver: absolute-path reject, parent-tree-escape reject, symlink reject (via `Lstat`), 4 MiB size cap. Error messages reference user-written paths, not resolved absolute paths.
- Parser-time error when both `command:` and `command_file:` are set on the same tool node.
- `examples/external_files.dip` + `examples/external_files/setup.sh` demonstrate the directive.

### Notes
- LSP and `cmd/wasm` (playground) skip the resolver. They see `cfg.CommandFile != "" && cfg.Command == ""`, which is the correct unresolved-IR view.
- DOT round-trip is lossy for the directive form — pack-then-unpack rewrites `command_file:` to inline `command:`. The runtime reads from `.dipx` bundles where content is already inlined; no current consumer needs the path preserved through DOT. Deferred to a follow-up issue ([#69](https://github.com/2389-research/dippin-lang/issues/69)).

## [v0.32.0] — 2026-05-27

New agent-node safety primitive: `tool_access: none` strips an LLM's tool catalog. Coordinated runtime release — the dippin field is meaningless without runtime enforcement, so they ship together (see [#41](https://github.com/2389-research/dippin-lang/issues/41) for context, including the v0.28.2 runaway-agent incident this bounds).

### Added

- `tool_access:` field on agent nodes. One explicit value: `none` (no LLM tools). Omitted = full catalog (current behavior).
- DIP139 lint warns on invalid `tool_access` values.
- `examples/agent_tool_access.dip` demonstrates the field on a summarizer node.

### Runtime requirement

- Requires an enforcing runtime. Without runtime enforcement the `tool_access:` field is a no-op (parking decision: lint-validated runtime-no-op safety fields ship as worse-than-nothing — the coordinated release prevents this).

### Runtime-side

- Tool registry returns empty when `tool_access: none` is set.
- Anthropic translator strips the `tools` array via `tool_choice: none`.
- System prompt scrubbed of tool-naming text when tools are disabled.
- `Params` keys (`allowed_tools`, `disallowed_tools`, `tool_choice`, `permission_mode`) are not honored when `tool_access: none` is set — Params bypass defense.
- Backend-compat tests: every supported backend honors `tool_access: none` or refuses session creation with a clear error.
- Red-team test: multi-tool-call LLM response under `tool_access: none` produces zero executions (the actual v0.28.2 shape).

## [v0.31.0] — 2026-05-22

Two small dippin-internal fixes. Closes [#45](https://github.com/2389-research/dippin-lang/issues/45), [#49](https://github.com/2389-research/dippin-lang/issues/49).

### Fixed

- `mode: yes_no` on `human` nodes no longer trips DIP127. The runtime supports `yes_no` as a documented mode; dippin's validator now accepts it alongside `choice`, `freeform`, and `interview`. The four-mode list is cascaded through the validator help text, the DIP127 explanation, `docs/validation.md`, `docs/nodes.md`, `docs/llm-reference.md`, the hosted skill (`site/static/skill.md`), the LSP completion tooltip, the integration guide, and the IR field comment.
- Tool nodes used as the workflow's `start:` node no longer collapse to `AgentConfig` after a `.dip → DOT → .dip` round-trip. `migrate.resolveStartExitKind` now recovers `NodeTool` from the start-marker `Mdiamond` shape by sniffing tool-specific DOT attributes (`tool_command`, `marker_grep`, `outputs`, `route_required`, `output_limit`).

### Runtime requirement

None. Both fixes are dippin-internal; the runtime is unaffected.

## [v0.30.0] — 2026-05-21

Coverage extractor now respects shell redirection. Closes [#40](https://github.com/2389-research/dippin-lang/issues/40).

### Fixed

- `dippin coverage` no longer flags file-redirected `echo`/`printf` statements as uncovered tool outputs. The extractor switched from regex matching to an AST walker built on `mvdan.cc/sh/v3/syntax`.
- Statements that redirect to files (`>`, `>>`, `&>`, `>&`), feed into pipes, or are nested inside command substitution are now correctly skipped. Pipelines using the "log to file, printf marker on stdout" pattern flip from `partial` to `covered`.

## [v0.29.0] — 2026-05-19

Three follow-ups to v0.28.0's tool-routing surface. Closes [#42](https://github.com/2389-research/dippin-lang/issues/42), [#43](https://github.com/2389-research/dippin-lang/issues/43), [#44](https://github.com/2389-research/dippin-lang/issues/44).

### Added

- `DIP138` reserved for a future advisory: "tool node routes on stdout but declares no `marker_grep` / `outputs`". Code + description + explanation entry only; no firing logic in this release.
- `outputs:` now survives a `.dip → DOT → .dip` round-trip. DOT export emits `outputs="a,b,c"` on tool nodes that declare outputs, and `dippin migrate dot→.dip` reads it back.

### Changed

- `DIP101` / `DIP102` no longer fire on tool nodes that declare `marker_grep:`. Those nodes route via the typed `ctx.tool_marker` channel — outgoing conditional edges on them are intentional routing, not unsafe reachability. Removes the false-positive coverage hit that the runtime's `TRK101` option (d) guidance triggered.
- Parser bool fields (`goal_gate`, `auto_status`, `cache_tools`, `route_required`) now accept `true/false`, `1/0`, `yes/no`, `on/off` case-insensitively via a new shared `parseBoolAttr` helper. Anything else now produces a parse diagnostic instead of silently coercing to `false`. The migrate (DOT-input) path keeps strict equality since DOT attrs are machine-emitted.

### Runtime requirement

None. All changes are dippin-internal; the runtime is unaffected.

### Docs

- `docs/nodes.md` "Markers and Verbose Output" notes the DIP101/DIP102 suppression and the accepted boolean forms.
- `docs/llm-reference.md` common-mistakes table cross-references the suppression behavior.
- Hosted skill (`site/static/skill.md`) updated to match.

## [v0.28.0] — 2026-05-19

Tool-node routing fields land in the parser and IR. Authors following the runtime's `TRK101` recommendation can now declare `marker_grep`, `route_required`, and `output_limit` directly in `.dip` source. Closes [#39](https://github.com/2389-research/dippin-lang/issues/39).

### Added

- `tool.marker_grep` — regex matched line-by-line against captured stdout; populates `ctx.tool_marker` at runtime.
- `tool.route_required` — boolean; when true, the node fails with `EventToolRouteMissing` if the command emits no routing signal recognized by the runtime.
- `tool.output_limit` — non-negative integer (bytes); 0 uses the engine default stdout tail-window. `dippin fmt` omits the field when the value is zero.
- Reserved context variables: `ctx.tool_marker`, `ctx.tool_route`.

### Changed

- `migrate/parity.go compareToolConfigs` now compares all `ToolConfig` fields. Pre-existing `Timeout` / `Outputs` parity gaps backfilled.

### Runtime requirement

These fields pass through DOT export to the runtime. Routing semantics require the runtime to ship the matching `extractToolAttrs` change; see issue [#39](https://github.com/2389-research/dippin-lang/issues/39) for details.

### Docs

- New blog post: [`site/content/blog/whats-new-v028.md`](site/content/blog/whats-new-v028.md).
- `docs/nodes.md` gains a "Best" subsection in Markers and Verbose Output demonstrating `marker_grep`.
- Hosted skill (`site/static/skill.md`) updated with new context variables and best-practice bullet.

## [v0.27.0] — 2026-05-18

Model catalog and pricing verification pass against canonical provider docs. `Last verified: 2026-05-18` in `validator/lint_model.go` and `cost/pricing.go`. No breaking changes to public APIs — but the cost table values move and the catalog accepts new IDs, so downstream tooling that snapshots dippin's data should re-snapshot.

### Added

- **OpenAI:** `gpt-5.5`, `gpt-5.5-pro`, `gpt-5.4-pro`, `gpt-5.2-pro`, `gpt-5`, `gpt-5-pro`, `gpt-5-mini`, `gpt-5-nano`.
- **xAI:** `grok-4.3` ($1.25/$2.50, current flagship).
- **DeepSeek:** `deepseek-v4-flash` ($0.14/$0.28), `deepseek-v4-pro` ($1.74/$3.48 list — 75% launch discount through 2026-05-31).
- **Gemini:** `gemini-3.1-flash-lite` (GA promotion of the preview variant, same price).
- **Mistral:** `mistral-medium-3-5-2604` ($1.50/$7.50, new flagship-class), `mistral-medium-3-1-2508` ($0.40/$2.00), Ministral 3 generation (`ministral-3-3b-2512` $0.10/$0.10, `ministral-3-8b-2512` $0.15/$0.15, `ministral-3-14b-2512` $0.20/$0.20).
- **Cohere:** dated IDs `command-r-08-2024`, `command-r-plus-08-2024`, `command-r7b-12-2024` (the canonical recommended form; bare aliases kept callable but flagged as resolving to versions deprecated 2025-09-15).

### Changed

- **OpenAI prices doubled** for three legacy IDs: `gpt-5.2` $0.875/$7 → $1.75/$14, `gpt-5.1` $0.625/$5 → $1.25/$10, `gpt-4.1-mini` $0.20/$0.80 → $0.40/$1.60. Newer mini/nano tiers and the o-series held steady.
- **xAI fleet-wide price cut**: `grok-4.20-0309-reasoning`, `grok-4.20-0309-non-reasoning`, `grok-4.20-multi-agent-0309` all $2/$6 → $1.25/$2.50, matching grok-4.3's tier.
- **DeepSeek alias repricing**: `deepseek-chat` and `deepseek-reasoner` are compatibility aliases resolving to V4-Flash; priced at $0.14/$0.28 (down from $0.28/$0.42) to match the redirect target.
- **Anthropic `claude-haiku-3-5` repriced** to $0.80/$4.00 (Bedrock/Vertex passthrough rate; was $0.25/$1.25 in the catalog). Model was retired on the first-party API 2026-02-19; remains available via Bedrock and Vertex AI.

### Removed

Hard-retired models that the provider returns errors for — calling them is a real bug, DIP108 surfaces:

- **Mistral:** `pixtral-large` (deprecated 2026-02-27), `mistral-small-3.2` (deprecated 2026-04-30, past sunset).
- **Gemini:** `gemini-3-pro-preview` (shut down 2026-03-09).

Soft-retired models that the provider silently redirects (kept in the catalog, priced at the redirect target so cost analysis stays accurate):

- **xAI:** `grok-4-1-fast-reasoning` and `grok-4-1-fast-non-reasoning` (retired 2026-05-15; xAI redirects to grok-4.3 server-side and bills at grok-4.3 rates).

### Deprecation calendar

Still in the catalog with deprecation comments:

- **2026-06-01**: `gemini-2.0-flash`.
- **2026-06-15**: `claude-sonnet-4-0`, `claude-opus-4-0`.
- **2026-07-24**: `deepseek-chat`, `deepseek-reasoner` aliases.
- **2026-10-23**: `gpt-4o`, `gpt-4.1-nano`, `o3-mini`, `o4-mini`.

### Documented uncertainties

Inline comments in `cost/pricing.go` flag values held pending re-verification:

- Mistral `nemo` and `mistral-small-2603`: official pricing tab is JS-rendered; third-party sources conflict.
- Cohere `command-a-03-2025` and `command-r7b-12-2024`: per-token pricing removed from the public page.
- Gemini Pro >200K-tier and OpenAI gpt-5.5 / gpt-5.4 family >272K-tier: modeled at the base tier only.

### Docs

- Blog post `site/content/blog/whats-new-v027.md` covers the refresh.
- Blog post `site/content/blog/whats-new-v026.md` retrospective on the v0.26 `requires:` keyword.
- Homepage "Latest" slot and v0.25/v0.26 `related:` cross-references updated.

## [v0.26.0] — 2026-05-15

### Added

- **Workflow header `requires:` keyword.** New optional workflow-header field for declaring workflow-level prerequisites (e.g., tools, MCP servers, env vars) as a comma-separated identifier list. Advisory in v1 — parsed, round-tripped by the formatter, and exposed as `ir.Workflow.Requires []string`, but not yet validated by lint. Mirrors the shape of node-level `reads:` / `writes:`. Filed to unblock the `--git=` preflight mechanism in the runtime. Canonical formatter order is `goal → requires → start → exit`. Editor support (tree-sitter, VS Code, Zed) and the hosted skill (`site/static/skill.md`) updated.

## [v0.25.0] — 2026-05-11

`.dipx` format v1.1. The spec at `docs/superpowers/specs/2026-05-06-dipx-bundle-format-design.md` is the canonical contract; this release closes ambiguities in it (Bundle 6), brings the implementation in line with the documented contract, and adds genuine cancellation through Pack/Open hot paths.

**Breaking changes for downstream consumers:**

- `Source.Workflow` now takes `context.Context` as its first argument. Bump your `dippin-lang` import via `go install ...@v0.25.0` (or `@latest`) and update call sites.
- `dippin inspect --format=json` `status` field is now an object, not a bare `"VALID"` string. If you parse the JSON in scripts, decode `status` as an object with `valid`, `verify_skipped`, `file_count`, `byte_total`, `format_version`.

### Fixed

- **Cycle detection now covers every manifest-listed workflow.** `dipx.Open` previously DFS'd the ref graph rooted only at `m.Entry`, while `parseAllWorkflows` already parsed every manifest-listed workflow. A cycle in a manifest-listed-but-entry-unreachable workflow could slip through. `walkRefs` now iterates `detectCycles` over `m.Files`. (Bundle 6 / Phase 5 L2/L3.)
- **`dippin pack`/`unpack`/`inspect` exit code 2 (integrity failure) now matches the spec contract.** `isIntegrityErr` previously routed only 5 sentinels to exit 2; 7 others (`ErrUnsupportedFormatVersion`, `ErrFileMissing`, `ErrFileUnexpected`, `ErrEntryNotInManifest`, `ErrRefEscape`, `ErrRefCycle`, `ErrCapExceeded`, `ErrPathUnsafe`) defaulted to user-error 1. Refactored to a sentinel-slice + loop covering all 12 spec-enumerated sentinels. (Bundle 6 / Phase 8 M1.)
- **`Open` enriches manifest-decode errors with the bundle path.** `BundleError.Path` for `ErrManifestInvalid` and `ErrUnsupportedFormatVersion` was previously empty or a JSON field name (e.g., `"format_version"`); external callers now always observe the bundle file path. The original Path is preserved in Detail when non-empty. (Bundle 5 / Phase 3 manifest-decoder error-context.)
- **`Pack` subgraph parse failures attribute to `ErrSubgraphParse`.** Previously every parse failure surfaced as `ErrEntryParse` regardless of which workflow failed; subgraph failures now correctly classify as `ErrSubgraphParse` with the subgraph's filesystem path. (Bundle 5 / P10.9.)
- **`dippin inspect` emits a structured `status` object.** JSON output's `status` was previously a bare string `"VALID"`; it is now an object with `valid`, `verify_skipped`, `file_count`, `byte_total`, `format_version` per spec § "CLI / inspect command". Text footer now includes the byte total. **Breaking** for JSON consumers parsing `status` as a string. (Bundle 2 / Phase 8 M4, L1, L2.)
- **`dippin inspect --no-verify` actually skips hash verification.** Previously a no-op (warning printed, full verification still ran). Now routes through a new `dipx.OpenManifest` API that performs only structural-admission steps; tampered bundles can be inspected without integrity errors firing. (Bundle 2 / Phase 10 P10.4.)
- **`Source.Workflow` now takes `context.Context` as its first argument.** `dirSource.Workflow` checks ctx before disk I/O; `Bundle.Workflow` checks ctx at entry for interface consistency. **Breaking** for external callers (the runtime) — bump your dippin-lang import to pick up the new signature. (Bundle 1 / Phase 6 L4.)
- **`Open` and `Pack` are cancellable mid-loop.** `verifyAllHashes`, `walkSourceTree`, and `writeBundle` now check `ctx.Err()` between each entry/iteration. A long Open against a many-entry bundle or a long Pack against a deep source tree can be canceled within one entry's processing time instead of running to completion. (Bundle 1 / Phase 10 P10.2, P10.7, P10.10.)

### Spec

Seven `.dipx` bundle-format spec clarifications (no behavior change beyond the two Fixed items above). Each is described in detail in the per-commit messages on this branch.

- **Path canonicalization rule 2** narrowed to "Backslash `\` MUST be rejected" (was "Backslash `\` and any other separator…"). The implementation already rejects only backslash; the spec wording was over-broad.
- **Per-sentinel error context preamble** added to disambiguate `BundleError.Path` semantics across three real cases: bundle-relative (read-side, post-Open), JSON field name (manifest decode pre-bundle-context), source filesystem path (Pack-side). Spec now requires `Open` to enrich (b) → (a) before returning.
- **Open ordering step 5** ("Verify no extra zip entries") inserted as a normative step between manifest-shape validation and hash verification; subsequent steps renumbered. `ErrFileUnexpected` added to the error precedence list at category 4.
- **Cycle detection scope** documented: spec § "Open ordering" step 8 now specifies "every manifest-listed workflow," matching `parseAllWorkflows`.
- **Integrity-failure sentinel set** for CLI exit code 2 expanded from 5 to all 12 spec-enumerated sentinels.
- **`inspect --format=json` status object schema** documented with a concrete JSON example (`valid`, `verify_skipped`, `file_count`, `byte_total`, `format_version`).
- **Runtime integration migration example** updated: `Source.Workflow(ctx, sub.Ref, parentPath)` (was missing `ctx`). Bundle 1 will land the matching Go signature change.

## [v0.24.0] — 2026-05-08

### Added

- **`.dipx` bundle format** — deterministic, content-addressed ZIP that packages a `.dip` entry workflow plus every transitively-reachable subgraph into a single integrity-verified artifact. Bundles carry a SHA-256-per-file manifest and a workflow-tree identity hash; integrity is verified on every Open. New package `dipx/` exposes `Open`, `OpenLax`, `OpenReader`, `Pack`, `Extract`, `Validate`, and `Load`, plus the `Source` interface (`Entry`, `Workflow`) for runtime consumers.
- **New CLI commands**: `dippin pack <entry.dip>` (build a bundle, with `-o`, `--dry-run`); `dippin unpack <bundle.dipx>` (atomic extract via staging + rename, with `-o`, `--force`); `dippin inspect <bundle.dipx>` (print manifest, identity hash, file list; `--format text|json`).
- **Existing commands accept `.dipx`** — `validate`, `lint`, `doctor`, `parse`, `cost`, `coverage`, `simulate`, `optimize`, `unused`, `graph`, `diff`, `check`, `explain`, `export-dot` now transparently load a `.dipx` via `dipx.Load`, hash-verify it, and analyze the entry workflow.
- **Distinct exit codes for bundle commands**: `0` (ok), `1` (user error), `2` (integrity error), `3` (I/O error), `4` (cancelled).
- **CLAUDE.md loader-tier exemption**: `dipx` may import `ir + parser + simulate` but is forbidden from importing `validator`, `cost`, `formatter`, or any other analysis package. Pack-time structural validation runs at the CLI layer (`cmd/dippin/cmd_pack.go`).

### Fixed

- `Extract --force` no longer destroys the existing destination directory when the staging-into-place rename fails on a cross-device boundary (EXDEV). The new backup-aside / rename-into-place / remove-aside sequence preserves the original on failure.
- `Pack` rejects symlinked parent directories anywhere between the entry's source root and a leaf `.dip`, closing a host-file exfiltration vector when packing untrusted source trees (CI-runner contributor builds, mono-repo subdirs).
- `Pack`'s ref-escape check no longer false-positives on legitimate filenames whose component name begins with `..` (e.g., `..foo/bar.dip`). The check now requires the literal `..` component, not a `..` substring.
- `dippin pack -o foo.dipx` no longer races two parallel invocations against the same temp filename — uses `os.CreateTemp` for a unique staging path.

### Internal

- Upgraded `golang.org/x/text` from v0.3.3 to v0.37.0 (defensive — only `unicode/norm` is consumed; the v0.3.3 CVEs were in `x/text/language`).

## [v0.23.0] — 2026-04-22

### Added
- **`WorkflowDefaults` tool-safety fields**: `tool_commands_allow` (glob allowlist for tool-node shell commands) and `tool_denylist_add` (globs appended to the runtime's default denylist). Both round-trip through parser → formatter → DOT export → migrate. Values pass through verbatim — the runtime owns split and glob semantics. ([#28](https://github.com/2389-research/dippin-lang/issues/28))

### Changed
- **DOT export header format** — `ExportDOT` now emits graph-level attributes (`rankdir`, tool-safety defaults, workflow vars) as a single `graph [key=val, ...];` block instead of separate bare statements (`rankdir=TB;`). This is the form the migrate DOT parser accepts, enabling true `.dip` → DOT → `.dip` round-trips. The output remains valid DOT; consumers that only render via Graphviz are unaffected.
- **`tool_commands_allow` / `tool_denylist_add` in `vars:` no longer emitted** — before this release, these keys weren't reserved, so a workflow that smuggled them through `vars:` would have them emitted as graph attributes. They are now reserved in favor of the dedicated `defaults:` fields. Any workflow that previously set either key via `vars:` should move it into `defaults:`; otherwise the value is silently dropped from DOT output (the runtime would see no allowlist). This path was never documented — issue #28 filed specifically because `defaults:` rejected the keys — so the affected population is expected to be zero, but calling it out explicitly for anyone who found the workaround.

## [v0.22.0] — 2026-04-22

### Added
- **`manager_loop` node kind** for supervising a child sub-pipeline with polling and mid-run context steering. Maps to the runtime's `stack.manager_loop` and DOT `shape=house`. Fields: `subgraph_ref`, `poll_interval`, `max_cycles`, `stop_condition`, `steer_condition`, `steer_context` (inline `k=v,k=v` or block form). Round-trips losslessly through parser → formatter → DOT export → migrate. Requires the parallel runtime adapter update. ([#26](https://github.com/2389-research/dippin-lang/issues/26), [#27](https://github.com/2389-research/dippin-lang/pull/27))
- **DIP135-137** lint codes for `manager_loop` validation: missing/nonexistent `subgraph_ref` (DIP135), invalid control field — negative `poll_interval` or `max_cycles` (DIP136), unbounded supervision with no `stop_condition` and no `max_cycles` (DIP137 — the manager_loop analog of DIP104).
- **`stack.*` namespace** recognized by DIP120 so `stop_condition` and `steer_condition` can reference `stack.child.cycles`, `stack.child.outcome`, `stack.child.status` without namespace warnings.
- **`dippin scaffold manager_loop`** template emits a starter supervisor workflow.
- **Tree-sitter grammar** — `manager_loop` node rule, highlights coverage, corpus test, committed generated parser (`src/parser.c` et al.), new `just tree-sitter-generate` / `just tree-sitter-test` recipes, and CI drift check so generated files can't drift from `grammar.js` without being caught.
- **VS Code TextMate grammar** — `manager_loop` keyword, new field names, `stack.*` namespace recognition.

### Fixed
- **Parser `steer_context` block-form routing** — a single-entry block-form `steer_context` (one `k: v` line under the indent) lexes without an embedded newline; the previous newline-based heuristic mis-routed it to the inline CSV handler. Replaced with a separator-position check (`:` before `=` means block, `=` before `:` means inline).

## [v0.21.0] — 2026-04-20

### Added
- **`HumanConfig.Timeout` / `TimeoutAction`** on human nodes. Pairs with edge labels like `when: timeout` for auto-advance semantics. Round-trips through parser, formatter, DOT export, and migrate. ([#22](https://github.com/2389-research/dippin-lang/pull/22))
- **`WorkflowDefaults` budget fields**: `max_total_tokens`, `max_cost_cents`, `max_wall_time`. Allow workflows to declare global budget caps consumed by the runtime. ([#22](https://github.com/2389-research/dippin-lang/pull/22))
- **Scoped context reads** — `ctx.node.<id>.*` now validates as a legitimate read pattern in DIP121/DIP122, eliminating lint false-positives for cross-node state access. ([#23](https://github.com/2389-research/dippin-lang/pull/23))
- **Agent-readiness discovery endpoints** on the docs site: `.well-known/agent-skills/index.json`, `.well-known/mcp/server-card.json`, `.well-known/api-catalog`, `robots.txt`, and hosted `skill.md`. Lets coding agents auto-discover dippin-lang tooling. ([#24](https://github.com/2389-research/dippin-lang/pull/24))
- **`reasoning_effort` expansion** — DIP119 now accepts `none`, `minimal`, `low`, `medium`, `high`, `xhigh`, and `max` to cover Opus 4.7 and GPT-5.4.
- **Model catalog update** (verified 2026-04-17): `claude-opus-4-7` (Anthropic, $5/$25), `mistral-small-2603` (Mistral Small 4), `command-a-03-2025` (Cohere flagship, $2.50/$10).

### Fixed
- **Syntax grammars** (VS Code TextMate, language-configuration, site highlight.js) updated to cover the `conditional` node kind and `vars` section introduced in v0.20.0.
- **`claude-haiku-3-5` deprecation comment** corrected — retired 2026-02-19, not 2026-04-19.

## [v0.20.0] — 2026-04-17

### Added
- **`vars` block** at the workflow level for declaring user-defined variables. Vars export as DOT graph-level attributes and round-trip through parse → format → export → migrate.
- **DIP134 lint rule**: warns when `max_retries` is set in defaults with `restart: true` edges but no `max_restarts` — catches the common confusion between per-node LLM retries and loop restart budget.
- **Release invariant checks** (`releasecheck/`) — validates the embedded spec is tracked, current, and buildable from a source tree without `.git`.

### Fixed
- **DIP125 false positives on shell variable assignments**. Replaced regex-based binary extraction with proper shell AST parsing via `mvdan.cc/sh/v3`. Variable assignments, command substitutions, and `command -v` checks are now correctly identified. Preamble commands (`mkdir`) are skipped to find the real tool binary.
- **`go install ...@latest` broken** — `cmd/dippin/generated-spec.md` is now checked into the repo so the `go:embed` directive resolves from module proxy downloads.
- **Pre-commit hook and `just check` now mirror CI exactly** — spec freshness, release checks, complexity exclusions all aligned.

## [v0.19.1] — 2026-04-16

### Added
- **`working_dir` field** on agent nodes for per-node working directory override (e.g., `.ai/worktrees/claude`). Wired through parser, formatter, DOT export, and migrate.

## [v0.19.0] — 2026-04-16

### Added
- **`backend` field** on agent nodes for per-node backend selection (e.g., `native`, `claude-code`, `acp`). Previously this value was silently dropped by the parser.
- **`working_dir` field** on agent nodes for per-node working directory override.

### Fixed
- **Unrecognized node fields** now emit a parse diagnostic suggesting the user put the field under `params:`, instead of being silently discarded.

## [v0.18.0] — 2026-04-06

### Added
- **`flatten` package** for resolving subgraph refs into flat workflows. Recursive resolution with cycle detection and configurable depth limit (default 10). Underscore-prefixed node IDs (`Parent_Child`).
- **`dippin export-dip`** command — exports a flattened workflow as canonical `.dip` text with all subgraph refs inlined.
- **`dippin export-dot`** now automatically flattens subgraph refs before export, producing valid DOT without external references.
- **Example workflows** — `orchestrator.dip` (parent with subgraph ref) and `phases/code_review.dip` (child workflow).
- **`TestLintExamples`** now recurses into `examples/*/` subdirectories.

### Fixed
- **Start/Exit rewrite** — workflow `Start`/`Exit` fields are now correctly remapped when they point to inlined subgraph nodes.
- **Nil resolver guard** — `flatten.Flatten` returns a clear error instead of panicking when the resolver is nil but subgraph refs are present.
- **`export-dot` error rendering** — flatten errors now use `renderError` for JSON output consistency.

## [v0.17.0] — 2026-04-03

### Added
- **`conditional` node kind** for pure branching without LLM calls. Evaluates outgoing edge conditions only — no prompt, no token cost. Maps to `diamond` shape in DOT export. DOT migration auto-detects: bare `diamond` → `conditional`, `diamond` + `prompt` → `agent`, `diamond` + `tool_command` → `tool`.
- **`--extra-models` CLI flag** on `lint` and `doctor` commands. Extends the DIP108 model catalog at runtime for private or newly-released models. Format: `--extra-models "provider:model1,model2;provider2:model3"`.

### Fixed
- **Bracket edge syntax** (`[label: ...]`) now emits a clear parse error with a hint to use `when`/`label:` keyword syntax, instead of silently discarding annotations.
- **Nested `retry` blocks** now emit a clear parse error suggesting flat attributes (`retry_policy`, `max_retries`, `retry_target`, `fallback_target`, `base_delay`), instead of a confusing indent mismatch error.

## [v0.16.0] — 2026-03-31

### Added
- **Structured output support** for agent nodes. New fields:
  - `response_format`: force LLM to produce structured JSON output (`json_object` or `json_schema`)
  - `response_schema`: inline JSON Schema definition (multiline block, like `prompt:`)
  - `params`: generic key-value pass-through for runtime features (same syntax as subgraph `params`)
- **DIP130**: lint warning for invalid `response_format` value.
- **DIP131**: lint warning when `response_schema` is set without `response_format: json_schema` (schema ignored); hint when `json_schema` is set without a schema.
- **DIP132**: lint warning when `response_schema` is not valid JSON.
- **DIP133**: lint hint when agent `params` key shadows a first-class field (e.g., `model`, `provider`).
- `cmd_timeout` field now parsed and formatted on agent nodes (previously only populated by DOT migrator).

### Fixed
- **Duplicate params keys** now emit a parse diagnostic instead of silently last-write-wins.
- **Unknown defaults fields** now emit a parse diagnostic instead of being silently discarded.
- **`AgentConfig.Params`** initialized to empty map (matching `SubgraphConfig`), preventing nil-pointer issues in downstream consumers.
- **Cyclomatic/cognitive complexity** violations resolved across 6 files (lint_response.go, lint_human.go, parse_nodes.go, format.go, interactive.go).

## [v0.15.0] — 2026-03-31

### Added
- **Interview mode** for human nodes (`mode: interview`). Runtimes extract questions from upstream agent output and present each as an individual form field with optional suggested answers. New fields: `questions_key`, `answers_key`.
- **DIP127**: lint warning for invalid human node mode values.
- **DIP128**: lint warning when interview mode has a meaningless `default` value.
- **DIP129**: lint warning when interview mode has conflicting choice-style labeled edges.
- Integration guide updated with interview mode implementation guidance and recommended answer JSON schema.
- `api_design.dip` example updated to use interview mode for Q&A collection.
- **`interview_loop.dip`** example: reusable interview subgraph with iterative Q&A. Parameterized by topic and focus areas. LLM generates questions with suggested options, human answers via interview mode, assessor loops until requirements are clear. Grade A, ~$0.92/run.
- **3 blog posts**: Multi-line Prompts Without Escaping (deep dive), Conditional Edges (tutorial), Cost Estimation (tutorial). Hub-and-spokes model with cross-links.
- **Auto-deploy**: CI now deploys `site/` to GitHub Pages on successful main builds.

### Fixed
- `--version` / `-version` flags now work (previously failed with "flag provided but not defined").
- **Formatter idempotency**: subgraph param values with quotes (e.g., `"API design"`) were double-quoted on each format pass. Parser now strips surrounding quotes from param values.

## [v0.14.0] — 2026-03-27

### Added
- **`code_health_check.dip`** example: self-contained pipeline that audits a Go repo. Gathers context with shell tools, runs vet/staticcheck/tests in parallel, three-model independent review, synthesized report with quality gate and retry loop. 5 test scenarios. Grade A, ~$1/run.

## [v0.13.2] — 2026-03-27

### Changed
- **Single-source nav**: `site/_layout/nav.html` is the one source of truth. `scripts/sync-nav.sh` propagates it to all 16 pages with correct prefixes and active states. Pre-commit hook runs it automatically. No more editing nav in 16 files.
- `scripts/gen-changelog-html.sh` emits a placeholder nav that `sync-nav.sh` fills.
- `just sync-nav` recipe added.

## [v0.13.1] — 2026-03-27

### Fixed
- `just install` and `just build` now inject commit hash and build timestamp via ldflags. `dippin version` shows `dev (commit: abc1234, built: 2026-03-27T18:45:10Z)` instead of `dev (commit: none, built: unknown)`.

## [v0.13.0] — 2026-03-27

### Changed
- **Two-tier navigation** across all site pages. Top row: Docs, Playground, Blog, GitHub. Bottom row: CLI, Language, Testing, Validation, Analysis, Architecture, Editors, Changelog. Mobile collapses to hamburger with divider-separated groups.

### Fixed
- Mobile nav menu no longer renders as unstyled text on desktop (missing `display: none`).
- Blog index only shows the 5 published posts — removed 20 dead links to unwritten articles.
- Playground content no longer overlaps the nav bar (padding adjusted for two-tier height).
- Playground now has the floating dots background matching all other pages.
- Homepage "See all 25 posts" corrected to "All posts".
- Section spacing tightened (6rem → 4.5rem padding).
- Downstream consumer field report response written (`.tracker/field-report-response-2026-03-27.md`).

## [v0.12.0] — 2026-03-27

### Added
- **Blog section** with 25 planned post cards and topic filtering (Guides, Tutorials, Deep Dives, Reference).
- **5 blog posts** published: Getting Started, Scenario Testing, Migrating from DOT, CI Integration, Editor Setup. Edited for voice, clarity, and inline links.
- **Homepage "From the Blog"** section featuring 3 latest posts below the fold.
- **SEO meta tags** on all 12 site pages: Open Graph, Twitter Cards, descriptions, canonical URLs. Pages render rich previews when shared.
- **Blog ideas doc** (`docs/blog-ideas.md`) with 25 post synopses, coverage plans, and approach notes.
- Blog nav link added to all site pages.

## [v0.11.2] — 2026-03-27

### Fixed
- **Playground**: syntax-highlighted editor with transparent textarea over colored `<pre>` overlay. Tab key inserts 2 spaces.
- **Playground**: parse output shows highlighted JSON (keys, strings, booleans, numbers). Format output shows highlighted Dippin. Lint errors display with severity coloring.
- **Playground**: WASM race condition — polls for function registration before auto-linting on load. Returns `[]` not `null` for zero diagnostics.
- **Site**: syntax highlighting CSS selectors changed from `pre .hl-*` to `.hl-*` so colors work in playground output div, not just `<pre>` blocks.
- **Site**: JSON blocks inside `compare-code` divs (like gate.test.json on Testing page) now get highlighted — skip logic checks for existing `<span>` tags instead of parent class.
- **Site**: highlight.js token protection via `\x00N\x00` placeholders prevents regex passes from matching inside previously generated `<span>` class attributes.
- **Site**: JSON inside terminal output blocks (e.g. `$ dippin --format json test`) gets JSON highlighting applied to the embedded body.
- **Site**: changelog auto-generated from CHANGELOG.md via `scripts/gen-changelog-html.sh`. Pre-commit hook runs it when CHANGELOG.md is staged.

## [v0.11.1] — 2026-03-27

### Fixed
- **Playground**: WASM files (`dippin.wasm`, `wasm_exec.js`) now deployed to gh-pages so the playground actually loads.
- **Playground**: auto-runs lint on WASM load instead of showing a confusing "Ready" message while the Lint button appears active.
- **Site**: syntax highlighting (`highlight.js`) for all code blocks — Dippin, shell, terminal, and diagnostic output.
- **Site**: changelog page added at `changelog.html` with full version history.

## [v0.11.0] — 2026-03-27

### Added
- **DIP126** lint rule: subgraph `ref:` file validation — warns when referenced workflow file does not exist on disk.
- **`dippin watch`** command: file watcher that re-runs lint on `.dip` changes with 200ms debounce. Uses `fsnotify`.
- **`dippin test --coverage`** flag: edge coverage summary showing which workflow edges were/weren't traversed by test scenarios.
- **Tree-sitter grammar** scaffolding in `editors/tree-sitter-dippin/` — grammar.js, external scanner for indentation, highlight queries, and test corpus. Enables proper syntax highlighting in Neovim, Helix, and Zed.
- **WASM playground** at `site/playground.html` — browser-based editor with live parse, lint, and format via WebAssembly. Build with `just wasm`.
- `gemini-3.1-pro-preview-customtools` added to model catalog and pricing tables.
- 35 diagnostic codes total (was 34).

### Fixed
- **CI failures**: golangci-lint `errcheck` on `f.Close()`, `funlen` on `buildLintExplanations` (split into 4 functions), misspell false positive in DIP118 example.
- **Migration parity**: `consensus_task_parity.dip` and `semport_thematic.dip` model names now match DOT originals (`gemini-3.1-pro-preview-customtools`).

### Changed
- `validator/lint_tool_cmd.go` split with `//go:build !wasm` / `wasm` tags — `bash -n` syntax check and `exec.LookPath` binary check are no-ops in WASM.
- `validator/lint_subgraph.go` similarly gated for WASM (no `os.Stat`).
- Site mobile CSS improvements: table overflow handling, code word-break, install-cmd sizing.
- Site nav updated with Playground link across all pages.

### Documentation
- All references updated from 34→35 codes, DIP101–DIP125→DIP101–DIP126 across README, CLAUDE.md, docs/, and site/.
- `docs/validation.md` — full entry for DIP126.
- `docs/cli.md` — `watch` command section, `test --coverage` flag.
- `docs/editor-setup.md` — tree-sitter grammar availability.
- `dippin explain DIP126` — explanation with trigger and fix guidance.
- `mode: labeled` documented as not supported in `docs/nodes.md`.

## [v0.10.0] — 2026-03-26

### Added
- **DIP123** lint rule: tool command shell syntax errors detected via `bash -n`.
- **DIP124** lint rule: `${ctx.*}` references in tool commands that expand to empty at runtime.
- **DIP125** lint rule (hint): tool command binary not found on PATH (environment-dependent).
- **Brochure site** with 8 pages: home, CLI, Language Reference, Testing, Validation, Analysis, Architecture, Editor Setup. Hosted on GitHub Pages.
- 34 diagnostic codes total (was 31).

### Documentation
- All references updated from 31→34 codes, DIP101–DIP122→DIP101–DIP125 across README, CLAUDE.md, docs/, and site/.
- `docs/validation.md` — full entries for DIP123, DIP124, DIP125.
- `dippin explain DIP123/DIP124/DIP125` — explanations with triggers and fix guidance.

## [v0.9.0] — 2026-03-25

### Fixed
- **`preferred_label` now works on human gates** — scenario key `preferred_label` (or per-node `Gate.preferred_label`) matches against edge labels (case-insensitive substring). Previously silently ignored on freeform gates.
- **`prompt:` blocks now parse on human nodes** — `HumanConfig` gained a `Prompt` field. Multiline prompt blocks work the same as on agent nodes. Formatter round-trips correctly.
- **Tool auto-defaults no longer mask fallback edges** — empty-string scenario values (`"Node.tool_stdout": ""`) now suppress the auto-seeded `success` default, allowing unconditional fallback edges to fire.

### Added
- **`immediately_after` test assertion** — assert adjacency in the execution path: `"immediately_after": {"NodeX": "NodeY"}` checks that NodeY is the very next node after NodeX.
- **`branch` field for targeted parallel testing** — `"branch": ["WorkerA"]` limits which parallel fan-out branches are simulated. Without it, all branches are walked (new default).
- **Simulator walks all parallel branches** — parallel fan-out now visits all targets, not just the first. Matches real runtime behavior.
- **Example test suites** — `.test.json` files for `vulnerability_analyzer`, `consensus_task`, `code_quality_sweep`, and `sprint_exec` (20 tests across 4 workflows).
- **Test coverage at 95.7%** — up from 85.6%. Six packages at 100%.
- `just cover` now excludes untestable files (`main.go`, `cmd_lsp.go`) from coverage reports.

### Documentation
- `docs/testing.md` — added Caveats section (`not_visited` fragility with loop-breaking), Clearing Defaults section (empty-string technique), `immediately_after` field documentation.

## [v0.8.0] — 2026-03-25

### Fixed
- **Graph truncation on pipelines with restart edges** — `buildAdjacency()` included restart (back) edges, creating cycles that prevented Kahn's algorithm from assigning layers to downstream nodes. All nodes are now rendered. Affects both full and compact modes.
- **Simulator infinite loop on tool-gated loops** — pipelines with `when ctx.tool_stdout not contains all-done` loops would spin to the 500-step limit. New `MaxNodeVisits` option forces the loop-exit edge after N visits. The test runner sets this to 3 by default.
- **Per-node scenario injection in `dippin test`** — `NodeName.key=value` scenarios now work reliably because the loop-breaking fix allows the simulation to reach the target node.
- **Testrunner accepts empty/invalid schemas silently** — `LoadTestFile` now rejects `.test.json` files with zero tests.

### Added
- **CLI integration tests** — 32 new tests covering 10 previously untested commands (cost, coverage, doctor, optimize, unused, graph, diff, feedback, explain, test). `cmd/dippin` coverage: 44.9% → 79.3%.
- **Graph tests for parallel and restart-loop fixtures**.
- **DIP121 compound condition test** — verifies `and`/`or` conditions correctly fire per-variable.
- **Unused clean-workflow test** — verifies no false positives on linear workflows.
- `just release tag msg` recipe for tagging releases.
- DIP121/DIP122 added to README warnings table; `explain`, `unused`, `graph`, `test` added to commands table.

### Fixed (cosmetic)
- **README stale numbers** — diagnostic codes 30→34, DIP120→DIP125, examples 15→17, lint rules 21→25.
- **`appendConnector` dead branch** — identical if/else branches collapsed.

## [v0.7.0] — 2026-03-25

### Added
- **`dippin test`** — scenario test runner for workflow assertions. Define `.test.json` files alongside `.dip` workflows with expected status, visited/not-visited nodes, and path ordering. Supports `--verbose` flag for path tracing and JSON output for CI integration.
- **New package:** `testrunner/` — loads `.test.json` suites, runs each case through the simulator with injected scenario values, checks assertions against results.
- **New doc:** `docs/testing.md` — documents the `.test.json` format and test runner usage.

## [v0.6.0] — 2026-03-25

### Added
- **DIP121** lint rule: condition references variable not produced by source node's `IO.Writes`. Skips when writes are empty (advisory) or variable is a reserved runtime key (`ctx.outcome`, `ctx.status`, `ctx.internal.*`, `graph.*`, `params.*`).
- **DIP122** lint rule: condition tests value not declared in source tool's `ToolConfig.Outputs`. Only fires for tool nodes with explicitly declared outputs.
- Explanations for DIP121/DIP122 in `dippin explain`.

## [v0.5.0] — 2026-03-25

### Added
- **`dippin explain <DIPxxx>`** — rich explanations for all 34 diagnostic codes. Shows trigger conditions, fix guidance, and example snippets. Supports text and JSON output.
- **`dippin unused <file>`** — detects dead-branch nodes (reachable from start but no path to exit) and estimates wasted cost per run. Reuses `coverage.Analyze()` sink detection + `cost.Analyze()` for cost enrichment.
- **`dippin graph [--compact] <file>`** — terminal-rendered ASCII DAG visualization. Full mode renders box-drawing nodes with connectors; compact mode outputs single-line `[A] → [B] → [C]` format. JSON mode outputs layer structure.
- **New packages:** `unused/`, `graph/`, `testrunner/`
- **New files:** `validator/explanations.go` with `Explanation` struct for all DIP codes.

## [v0.4.3] — 2026-03-25

### Fixed
- **DIP101 suppressed for mixed routing** — when a source node has both unconditional and conditional outgoing edges, the conditional branches are intentional routing. DIP101 no longer fires on their destinations. Covers all four reported patterns: compound inequality conditions, exhaustive set + fallback, mixed unconditional/conditional, and labeled fallback edges.

## [v0.4.2] — 2026-03-25

### Fixed
- **DIP101/DIP102 exhaustive detection** now recognizes any complete partition — if all conditional edges from a node test the same variable with equality (2+ values), the conditions are treated as exhaustive. No longer limited to hardcoded `{success, fail}` pairs. Handles `done/more_questions`, `tasks_remain/all_done`, and any custom value set.

## [v0.4.1] — 2026-03-25

### Fixed
- **EBNF grammar** audited against parser — added infix negation, tool `outputs` field, removed undocumented numeric operators (`<`, `>`, `<=`, `>=` parsed but silently returned false)
- **Docs accuracy** — removed `state.*` namespace (not implemented), removed `ctx.preferred_label` (not in codebase), added `==` as `=` alias, added `not contains` infix syntax
- **Condition parser** — removed `<`, `>`, `<=`, `>=` from valid operators (never evaluated, silent false was a trap)

### Added
- `CHANGELOG.md` with retroactive history for all versions
- `docs/CONTRIBUTING.md` — documentation accuracy protocol with persona matrix
- `CLAUDE.md` — project conventions, gotchas, versioning policy
- Integration test (`TestLintExamples`) — lints all examples through real parser
- "Last verified" dates on model catalog and pricing table
- Tool `outputs` field documented in nodes.md and README.md

## [v0.4.0] — 2026-03-25

### Added
- **New commands:** `dippin cost`, `dippin coverage`, `dippin doctor`, `dippin optimize`, `dippin diff`, `dippin feedback`, `dippin lsp`
- **New providers:** DeepSeek, xAI (Grok), Mistral, Cohere — model validation and cost estimation
- **DIP116–DIP120** lint rules: compaction threshold, on_resume, stylesheet refs, reasoning_effort, namespace prefix
- **LSP server** with diagnostics, hover, go-to-definition, autocomplete, document symbols
- **Condition parser:** infix negation syntax (`var not contains val`)
- **Complementary pair detection:** `contains X` + `not contains X` recognized as exhaustive
- **New docs:** `docs/analysis.md`, `docs/editor-setup.md`

### Fixed
- **DIP101 false positives** — conditions were never parsed into ASTs; `Lint()` now calls `EnsureConditionsParsed()` before running checks
- **DIP101 exhaustive suppression** works for all three real-world patterns: exhaustive + fallback, exhaustive + extra variables, complementary pairs
- **DIP103** no longer flags `contains X` / `not contains X` as overlapping
- **DIP110** exempts start/exit lifecycle nodes from empty prompt warnings
- **Model pricing** corrected against official docs (claude-opus-4-6 was $15/$75, actually $5/$25; o3 was $10/$40, actually $2/$8)
- **Model IDs** corrected: `gemini-3-pro` → `gemini-3.1-pro-preview`, removed nonexistent IDs
- **Gemini pricing** added (was missing entirely from cost estimates)

### Changed
- Full documentation rewrite — all docs updated for post-v0.3.0 toolchain
- Mermaid diagrams use `<br>` instead of `<br/>`

## [v0.3.0] — 2026-03-21

### Added
- `dippin cost` and `dippin coverage` commands
- DIP119 (reasoning_effort validation), DIP120 (namespace prefix)
- DIP114 extended to parallel branch fidelity

### Fixed
- Scenario-injected values protected from node default overwrite
- errcheck and staticcheck lint errors in coverage package

### Changed
- Four largest files decomposed into focused modules

## [v0.2.0] — 2026-03-20

### Added
- `dippin check`, `dippin new` commands with 5 scaffold templates
- DIP113–DIP115 lint rules (retry policy, fidelity, goal gate)
- `base_delay` field for retry override
- Subgraph params, compaction, fidelity degradation, parallel branches, stylesheets
- `dippin version` command
- Justfile for dev workflows
- GoReleaser + GitHub Actions release pipeline
- Homebrew tap

### Changed
- All functions reduced to cyclomatic ≤ 5, cognitive ≤ 7

## [v0.1.0] — 2026-03-19

### Added
- Initial release
- Parser (indentation-aware lexer + recursive descent)
- Validator (DIP001–DIP009 structural checks)
- Linter (DIP101–DIP112 semantic warnings)
- Formatter (canonical idempotent output)
- DOT exporter with shape mapping
- DOT → Dippin migration with parity checker
- Simulator with JSONL event streaming
- 15 example workflows including 5 stress tests
- VS Code extension (syntax highlighting)

[v0.8.0]: https://github.com/2389-research/dippin-lang/compare/v0.7.0...v0.8.0
[v0.7.0]: https://github.com/2389-research/dippin-lang/compare/v0.6.0...v0.7.0
[v0.6.0]: https://github.com/2389-research/dippin-lang/compare/v0.5.0...v0.6.0
[v0.5.0]: https://github.com/2389-research/dippin-lang/compare/v0.4.3...v0.5.0
[v0.4.3]: https://github.com/2389-research/dippin-lang/compare/v0.4.2...v0.4.3
[v0.4.2]: https://github.com/2389-research/dippin-lang/compare/v0.4.1...v0.4.2
[v0.4.1]: https://github.com/2389-research/dippin-lang/compare/v0.4.0...v0.4.1
[v0.4.0]: https://github.com/2389-research/dippin-lang/compare/v0.3.0...v0.4.0
[v0.3.0]: https://github.com/2389-research/dippin-lang/compare/v0.2.0...v0.3.0
[v0.2.0]: https://github.com/2389-research/dippin-lang/compare/v0.1.0...v0.2.0
[v0.1.0]: https://github.com/2389-research/dippin-lang/releases/tag/v0.1.0
