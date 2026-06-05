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

Tag semver releases after batches of meaningful changes. Downstream consumers **always pin to a specific dippin version, never `@latest`** — so merged work on `main` is invisible to them until it's tagged. Cutting a tag is what delivers a feature to consumers; they then bump their pin to adopt it. Update CHANGELOG.md when tagging.

```sh
git tag -a v0.X.0 -m "description" && git push origin v0.X.0
```

GoReleaser is configured (`.goreleaser.yml`) — pushing a tag triggers GitHub Actions to build cross-platform binaries and publish to Homebrew tap.

## Model Catalog & Pricing

Model names and pricing in `validator/lint_model.go` and `cost/pricing.go` must be verified against official provider documentation before committing. Source URLs and "Last verified" dates are maintained as code comments. Never use training data for pricing — it goes stale.

Supported providers: Anthropic, OpenAI, Google/Gemini, DeepSeek, xAI/Grok, Mistral, Cohere.

`TestLintExamples` in `validator/lint_examples_test.go` parses all example .dip files through the real parser and asserts zero DIP108 warnings — this catches model catalog staleness and invalid model IDs.

## Lint Rules

54 diagnostic codes: DIP001-DIP009 (structural errors), DIP101-DIP145 (semantic warnings). DIP101/DIP102 suppress automatically when source node conditions are exhaustive (success/fail pairs, contains/not-contains complementary pairs). DIP121/DIP122 only fire when source nodes declare writes/outputs (advisory metadata).

## Testing

Test fixtures should match real parser output. If the parser doesn't populate a field, tests shouldn't either. The DIP101 bug was caused by tests pre-populating `Condition.Parsed` by hand, masking that production code never set it.

Integration test `TestLintExamples` runs every example through parse → lint to catch regressions that unit tests with hand-built IR would miss.
