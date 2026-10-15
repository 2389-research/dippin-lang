# CLAUDE.md

## Project

dippin-lang is a DSL and toolchain for authoring AI pipeline workflows. It replaces Graphviz DOT as the authoring format consumed by a downstream pipeline runtime.

## Claude Code Skill

A hosted skill teaches Claude Code to author, validate, and debug .dip files. Add to any project's CLAUDE.md:

```
@https://2389-research.github.io/dippin-lang/skill.md
```

## Build & Test — always use `just`

All common operations go through the justfile. Never run raw `go build`, `go test`, `gocyclo`, etc. directly — use the corresponding `just` recipe. If you find yourself running a command repeatedly that isn't in the justfile, add a recipe for it first.

```sh
just check              # full suite: build, vet, fmt, test-race, complexity, validate-examples
just test               # go test ./... -count=1
just test-race          # go test ./... -count=1 -race
just test-pkg validator # test a single package with -v
just build              # build the dippin binary
just install            # go install to $GOBIN
just vet                # go vet ./...
just fmt                # gofmt -w .
just fmt-check          # check formatting (CI-style, exit 1 if unformatted)
just complexity         # cyclomatic ≤ 5 + cognitive ≤ 7 (excludes tests)
just validate-examples  # run dippin validate on all examples/*.dip
just lint-examples      # run dippin lint on all examples/*.dip
just cover              # generate test coverage report
just cover-html         # open coverage in browser
just docs-sync-check   # docs/X.md changes must reach site/content/X.md (also in check/CI)
just release-check v0.X.0  # pre-tag gate: CHANGELOG entry, docs/site sync, no stale versions
just setup-hooks        # install pre-commit hook (required for first checkout)
just clean              # remove build artifacts
```

## Code Quality

Pre-commit hook enforces (mirrors CI exactly):
- `golangci-lint` (includes staticcheck, errcheck, etc.)
- Cyclomatic complexity ≤ 5 per function (`gocyclo`)
- Cognitive complexity ≤ 7 per function (`gocognit`)
- `gofmt` canonical formatting
- All tests pass with race detector
- All example `.dip` files validate

When a function exceeds complexity: extract helpers, don't add `//nolint` directives.

## Architecture

Everything flows through `ir.Workflow`. Packages import `ir` but not each other (except analysis packages that compose: doctor → validator + coverage + cost, unused → coverage + cost). One additional exemption: **`dipx`** is a "loader tier" package and may compose `ir + parser + simulate`. The exemption is bounded — `dipx` MUST NOT import `validator`, `cost`, `formatter`, or any other analysis package. Pack-time structural validation is invoked at the CLI layer (`cmd/dippin/cmd_pack.go`), not inside `dipx`.

Key gotcha: The parser stores edge conditions as `Condition.Raw` (plain text). `Condition.Parsed` (AST) is only populated by `simulate.EnsureConditionsParsed()`. Any code reading `Condition.Parsed` must ensure it's been called first — `Lint()` does this automatically.

## Git Workflow

- **Never commit or push to `main`.** All work happens on a unique feature branch (e.g. `fix/<issue>-<slug>`, `feat/<slug>`). Branch first, then commit.
- **Never push directly to `main`.** Changes reach `main` only through a reviewed Pull Request. `main` is delivered to downstream consumers via tags (see Versioning), so it must stay green and intentional.
- **Work off an isolated worktree per unit of work**, not the shared checkout — so an in-progress branch never collides with other work. Use `git worktree add` (or the `superpowers:using-git-worktrees` skill).
- **Parallel subagents each get their own worktree.** When dispatching subagents to work on independent/parallel branches of work simultaneously, give each its own git worktree (`isolation: "worktree"` for the Agent tool / `opts.isolation: "worktree"` in workflows) so concurrent edits can't conflict. Sequential subagents editing the same branch share the one worktree.
- Leave `.claude/settings.local.json` uncommitted.
- Commit messages end with the trailer: `Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>`.

## Versioning

Tag semver releases after batches of meaningful changes. Downstream consumers **always pin to a specific dippin version, never `@latest`** — so merged work on `main` is invisible to them until it's tagged. Cutting a tag is what delivers a feature to consumers; they then bump their pin to adopt it.

**Every release updates the website and documentation — no exceptions.** The site (dippin.org, Hugo in `site/`) has two kinds of pages:
- *Generated at deploy* (nothing to do by hand): `/changelog` from CHANGELOG.md, `llms-full.txt` from `docs/` via gen-spec, `/prices.json` + `/models.json` from `pricing/prices.json`.
- *Hand-written* (must be updated in the same PR as the change): `site/content/*.md` are rewrites of `docs/*.md` with their own structure — when a PR changes `docs/X.md`, it also changes `site/content/X.md` (enforced by `scripts/docs-sync-check.sh` in `just check`, pre-commit and CI; for a docs change with genuinely nothing user-facing, add a `Site-Sync: none` commit trailer). Also keep `site/static/skill.md` (the hosted Claude Code skill), README.md and AGENTS.md true. Never hard-code "the current/recommended version is vX" — link to the releases page instead.

Release checklist:
1. Changelog PR: add `## [v0.X.0] — YYYY-MM-DD` to CHANGELOG.md; review that every user-facing change in the release is reflected in `docs/`, its `site/content/` twin, and `site/static/skill.md` where relevant.
2. After it merges, on an up-to-date `main`: `just release-check v0.X.0` (fails on a missing CHANGELOG entry, docs/site drift since the last tag, or stale version claims).
3. Tag and push:
   ```sh
   git tag -a v0.X.0 -m "description" && git push origin v0.X.0
   ```
4. Verify: the GoReleaser run succeeds (`.goreleaser.yml` builds cross-platform binaries and publishes to the Homebrew tap), the main-branch CI "Deploy Site" job succeeded, and https://dippin.org/changelog/ shows the new version.
5. Open an adoption issue on downstream consumers (e.g. tracker) to bump their pin.

## Model Catalog & Pricing

The model catalog + pricing is a single source of truth: **`pricing/prices.json`**, embedded (`//go:embed`) by the leaf **`pricing`** package. `cost` (estimator) and `validator` (DIP108) both derive from it â never hand-maintain a second table. Each entry carries its published-price `source` URL and an `as_of` date; verify prices against official provider docs before committing and set `as_of`. Never use training data for pricing â it goes stale. An entry with `"priced": false` is a known-but-unpriced model (recognized by DIP108, priced $0) â used when a provider's USD rate isn't verifiable from an official page (e.g. `qwen3.7-plus`, whose only published rate is an open-ended limited-time discount). The same state applies at provider level via `custom_providers` in `prices.json` (currently `openai-compat`, aliased from `openai-compatible`, #297): a BYOK / custom-gateway provider whose model ids the catalog cannot enumerate — `pricing.CustomProvider(p)` is true, `LookupProvider` finds *any* id under it as `Priced=false`, DIP108 never fires, and a `@` in the id is opaque (not a family alias). It is scoped to the declared name so DIP108 stays strict for real providers. Cache read/write rates (cache_read_mult/cache_write_mult, or an absolute cached_input_per_m) are populated only for providers whose cache convention has been verified against official docs (Anthropic 0.1x read/1.25x write — Opus/Sonnet 5.5 are 0.05x, Fable/Mythos 5.1 0.025x; Gemini 0.1x read — gemini-4-argon 0.05x; Mistral 0.1x read per its prompt-caching doc; Qwen 0.2x implicit / 0.1x+1.25x explicit; OpenAI per-family read — gpt-4o 0.5x, gpt-4.1/o3/o4-mini 0.25x, o3-mini 0.5x, GPT-5 0.1x, gpt-6.1-sol 0.05x, `-pro` models 1x (no cached discount); DeepSeek, Z.AI/GLM, xAI/Grok, and Moonshot/Kimi carry absolute `cached_input_per_m` from their published cache-hit prices). Cohere carries `cache_read_mult: 1` — verified against its official page, which publishes **no** cached-input discount, so a cache read bills at the full input rate (1.0 is distinct from an unverified 0, which prices cache reads at $0 and means "consumer may overlay"). The still-unverified models leave the cache fields 0 and a consumer overlays until verified: MiniMax (the pricing page does show prompt-cache read/write columns, but automated reads of it disagree on the read rate and whether it is tiered by input length — see #292; no rate is committed until a human confirms) and qwen3.8-max/flash (Alibaba's cache page says their hit price is non-standard and defers to the login-gated console). An entry with `"deprecated": true` is retired on the first-party provider API (already shut down — an *announced* future shutdown does not set it) but still priced (it bills on Bedrock/Vertex passthrough) â a consumer treating the catalog as a first-party allowlist should filter on `ModelPrice.Deprecated`; catalog membership ≠ first-party callability. The `pricing` package is a leaf (imports nothing from dippin), so both consumers importing it does not violate the "packages import `ir`, not each other" rule. Optional per-entry metadata (all absent = unknown, never a claim of zero): drift-resistance `family`/`rank`/`maturity` — a consumer resolves a family reference (newest / one-back) via `pricing.ResolveAlias(provider, family, selector)`, which excludes deprecated/unpriced/preview — and capability `context_window`/`max_output`/`capabilities`. `pricing.CacheGaps()` lists priced models still lacking a cache rate; `TestCacheGapsAllowlist` fails if that set grows, so a new priced model can't ship without a rate (or a reasoned allowlist entry). **DIP161** warns when an agent pins a `deprecated` catalog model.

Supported providers: Anthropic, OpenAI, Google/Gemini, DeepSeek, xAI/Grok, Mistral, Cohere, Z.AI/GLM, Moonshot/Kimi, MiniMax, Qwen.

`TestLintExamples` in `validator/lint_examples_test.go` parses all example .dip files through the real parser and asserts zero DIP108 warnings — this catches model catalog staleness and invalid model IDs.

## Lint Rules

75 diagnostic codes: DIP001-DIP010 (structural errors), DIP101-DIP165 (semantic lint — mostly warnings; DIP155-DIP158 and DIP163 are error-severity and fail `lint`/`check`; DIP161 warns on agents pinned to a `deprecated` catalog model; DIP162 warns on an unresolvable `family@selector` model alias; DIP163-DIP165 cover `writable_paths_mode` — DIP163 rejects any value other than exactly `require`/`prefer`, DIP164 hints on a mode with no `writable_paths`, DIP165 hints that a `prefer` node runs UNJAILED on hosts without Landlock). DIP101/DIP102 suppress automatically when source node conditions are exhaustive (success/fail pairs, contains/not-contains complementary pairs). DIP121/DIP122 only fire when source nodes declare writes/outputs (advisory metadata).

## Testing

Test fixtures should match real parser output. If the parser doesn't populate a field, tests shouldn't either. The DIP101 bug was caused by tests pre-populating `Condition.Parsed` by hand, masking that production code never set it.

Integration test `TestLintExamples` runs every example through parse → lint to catch regressions that unit tests with hand-built IR would miss.
