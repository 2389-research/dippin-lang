# Changelog

All notable changes to dippin-lang are documented here. Versions follow [semver](https://semver.org/).

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

- `DIP101` / `DIP102` no longer fire on tool nodes that declare `marker_grep:`. Those nodes route via the typed `ctx.tool_marker` channel — outgoing conditional edges on them are intentional routing, not unsafe reachability. Removes the false-positive coverage hit that tracker's `TRK101` option (d) guidance triggered.
- Parser bool fields (`goal_gate`, `auto_status`, `cache_tools`, `route_required`) now accept `true/false`, `1/0`, `yes/no`, `on/off` case-insensitively via a new shared `parseBoolAttr` helper. Anything else now produces a parse diagnostic instead of silently coercing to `false`. The migrate (DOT-input) path keeps strict equality since DOT attrs are machine-emitted.

### Runtime requirement

None. All changes are dippin-internal; tracker is unaffected.

### Docs

- `docs/nodes.md` "Markers and Verbose Output" notes the DIP101/DIP102 suppression and the accepted boolean forms.
- `docs/llm-reference.md` common-mistakes table cross-references the suppression behavior.
- Hosted skill (`site/static/skill.md`) updated to match.

## [v0.28.0] — 2026-05-19

Tool-node routing fields land in the parser and IR. Authors following tracker's `TRK101` recommendation can now declare `marker_grep`, `route_required`, and `output_limit` directly in `.dip` source. Closes [#39](https://github.com/2389-research/dippin-lang/issues/39).

### Added

- `tool.marker_grep` — regex matched line-by-line against captured stdout; populates `ctx.tool_marker` at runtime.
- `tool.route_required` — boolean; when true, the node fails with `EventToolRouteMissing` if the command emits no `_TRACKER_ROUTE=<value>` sentinel line.
- `tool.output_limit` — non-negative integer (bytes); 0 uses the engine default stdout tail-window. `dippin fmt` omits the field when the value is zero.
- Reserved context variables: `ctx.tool_marker`, `ctx.tool_route`.

### Changed

- `migrate/parity.go compareToolConfigs` now compares all `ToolConfig` fields. Pre-existing `Timeout` / `Outputs` parity gaps backfilled.

### Runtime requirement

These fields pass through DOT export to tracker. Routing semantics require tracker to ship the matching `extractToolAttrs` change; see issue [#39](https://github.com/2389-research/dippin-lang/issues/39) for details.

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

- **Workflow header `requires:` keyword.** New optional workflow-header field for declaring workflow-level prerequisites (e.g., tools, MCP servers, env vars) as a comma-separated identifier list. Advisory in v1 — parsed, round-tripped by the formatter, and exposed as `ir.Workflow.Requires []string`, but not yet validated by lint. Mirrors the shape of node-level `reads:` / `writes:`. Filed from [tracker's git-preflight design](https://github.com/2389-research/tracker/blob/main/docs/superpowers/specs/2026-05-15-tracker-git-preflight-design.md) to unblock the `--git=` preflight mechanism. Canonical formatter order is `goal → requires → start → exit`. Editor support (tree-sitter, VS Code, Zed) and the hosted skill (`site/static/skill.md`) updated.

## [v0.25.0] — 2026-05-11

`.dipx` format v1.1. The spec at `docs/superpowers/specs/2026-05-06-dipx-bundle-format-design.md` is the canonical contract; this release closes ambiguities in it (Bundle 6), brings the implementation in line with the documented contract, and adds genuine cancellation through Pack/Open hot paths.

**Breaking changes for downstream consumers (Tracker, etc.):**

- `Source.Workflow` now takes `context.Context` as its first argument. Bump your `dippin-lang` import via `go install ...@v0.25.0` (or `@latest`) and update call sites.
- `dippin inspect --format=json` `status` field is now an object, not a bare `"VALID"` string. If you parse the JSON in scripts, decode `status` as an object with `valid`, `verify_skipped`, `file_count`, `byte_total`, `format_version`.

### Fixed

- **Cycle detection now covers every manifest-listed workflow.** `dipx.Open` previously DFS'd the ref graph rooted only at `m.Entry`, while `parseAllWorkflows` already parsed every manifest-listed workflow. A cycle in a manifest-listed-but-entry-unreachable workflow could slip through. `walkRefs` now iterates `detectCycles` over `m.Files`. (Bundle 6 / Phase 5 L2/L3.)
- **`dippin pack`/`unpack`/`inspect` exit code 2 (integrity failure) now matches the spec contract.** `isIntegrityErr` previously routed only 5 sentinels to exit 2; 7 others (`ErrUnsupportedFormatVersion`, `ErrFileMissing`, `ErrFileUnexpected`, `ErrEntryNotInManifest`, `ErrRefEscape`, `ErrRefCycle`, `ErrCapExceeded`, `ErrPathUnsafe`) defaulted to user-error 1. Refactored to a sentinel-slice + loop covering all 12 spec-enumerated sentinels. (Bundle 6 / Phase 8 M1.)
- **`Open` enriches manifest-decode errors with the bundle path.** `BundleError.Path` for `ErrManifestInvalid` and `ErrUnsupportedFormatVersion` was previously empty or a JSON field name (e.g., `"format_version"`); external callers now always observe the bundle file path. The original Path is preserved in Detail when non-empty. (Bundle 5 / Phase 3 manifest-decoder error-context.)
- **`Pack` subgraph parse failures attribute to `ErrSubgraphParse`.** Previously every parse failure surfaced as `ErrEntryParse` regardless of which workflow failed; subgraph failures now correctly classify as `ErrSubgraphParse` with the subgraph's filesystem path. (Bundle 5 / P10.9.)
- **`dippin inspect` emits a structured `status` object.** JSON output's `status` was previously a bare string `"VALID"`; it is now an object with `valid`, `verify_skipped`, `file_count`, `byte_total`, `format_version` per spec § "CLI / inspect command". Text footer now includes the byte total. **Breaking** for JSON consumers parsing `status` as a string. (Bundle 2 / Phase 8 M4, L1, L2.)
- **`dippin inspect --no-verify` actually skips hash verification.** Previously a no-op (warning printed, full verification still ran). Now routes through a new `dipx.OpenManifest` API that performs only structural-admission steps; tampered bundles can be inspected without integrity errors firing. (Bundle 2 / Phase 10 P10.4.)
- **`Source.Workflow` now takes `context.Context` as its first argument.** `dirSource.Workflow` checks ctx before disk I/O; `Bundle.Workflow` checks ctx at entry for interface consistency. **Breaking** for external callers (Tracker) — bump your dippin-lang import to pick up the new signature. (Bundle 1 / Phase 6 L4.)
- **`Open` and `Pack` are cancellable mid-loop.** `verifyAllHashes`, `walkSourceTree`, and `writeBundle` now check `ctx.Err()` between each entry/iteration. A long Open against a many-entry bundle or a long Pack against a deep source tree can be canceled within one entry's processing time instead of running to completion. (Bundle 1 / Phase 10 P10.2, P10.7, P10.10.)

### Spec

Seven `.dipx` bundle-format spec clarifications (no behavior change beyond the two Fixed items above). Each is described in detail in the per-commit messages on this branch.

- **Path canonicalization rule 2** narrowed to "Backslash `\` MUST be rejected" (was "Backslash `\` and any other separator…"). The implementation already rejects only backslash; the spec wording was over-broad.
- **Per-sentinel error context preamble** added to disambiguate `BundleError.Path` semantics across three real cases: bundle-relative (read-side, post-Open), JSON field name (manifest decode pre-bundle-context), source filesystem path (Pack-side). Spec now requires `Open` to enrich (b) → (a) before returning.
- **Open ordering step 5** ("Verify no extra zip entries") inserted as a normative step between manifest-shape validation and hash verification; subsequent steps renumbered. `ErrFileUnexpected` added to the error precedence list at category 4.
- **Cycle detection scope** documented: spec § "Open ordering" step 8 now specifies "every manifest-listed workflow," matching `parseAllWorkflows`.
- **Integrity-failure sentinel set** for CLI exit code 2 expanded from 5 to all 12 spec-enumerated sentinels.
- **`inspect --format=json` status object schema** documented with a concrete JSON example (`valid`, `verify_skipped`, `file_count`, `byte_total`, `format_version`).
- **Tracker integration migration example** updated: `Source.Workflow(ctx, sub.Ref, parentPath)` (was missing `ctx`). Bundle 1 will land the matching Go signature change.

## [v0.24.0] — 2026-05-08

### Added

- **`.dipx` bundle format** — deterministic, content-addressed ZIP that packages a `.dip` entry workflow plus every transitively-reachable subgraph into a single integrity-verified artifact. Bundles carry a SHA-256-per-file manifest and a workflow-tree identity hash; integrity is verified on every Open. New package `dipx/` exposes `Open`, `OpenLax`, `OpenReader`, `Pack`, `Extract`, `Validate`, and `Load`, plus the `Source` interface (`Entry`, `Workflow`) for runtime consumers like Tracker.
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
- **`WorkflowDefaults` tool-safety fields** (tracker#164 / tracker#169): `tool_commands_allow` (glob allowlist for tool-node shell commands) and `tool_denylist_add` (globs appended to tracker's default denylist). Both round-trip through parser → formatter → DOT export → migrate. Values pass through verbatim — tracker owns split and glob semantics. ([#28](https://github.com/2389-research/dippin-lang/issues/28))

### Changed
- **DOT export header format** — `ExportDOT` now emits graph-level attributes (`rankdir`, tool-safety defaults, workflow vars) as a single `graph [key=val, ...];` block instead of separate bare statements (`rankdir=TB;`). This is the form the migrate DOT parser accepts, enabling true `.dip` → DOT → `.dip` round-trips. The output remains valid DOT; consumers that only render via Graphviz are unaffected.
- **`tool_commands_allow` / `tool_denylist_add` in `vars:` no longer emitted** — before this release, these keys weren't reserved, so a workflow that smuggled them through `vars:` would have them emitted as graph attributes. They are now reserved in favor of the dedicated `defaults:` fields. Any workflow that previously set either key via `vars:` should move it into `defaults:`; otherwise the value is silently dropped from DOT output (tracker would see no allowlist). This path was never documented — issue #28 filed specifically because `defaults:` rejected the keys — so the affected population is expected to be zero, but calling it out explicitly for anyone who found the workaround.

## [v0.22.0] — 2026-04-22

### Added
- **`manager_loop` node kind** for supervising a child sub-pipeline with polling and mid-run context steering. Maps to Tracker's `stack.manager_loop` and DOT `shape=house`. Fields: `subgraph_ref`, `poll_interval`, `max_cycles`, `stop_condition`, `steer_condition`, `steer_context` (inline `k=v,k=v` or block form). Round-trips losslessly through parser → formatter → DOT export → migrate. Requires the parallel Tracker adapter update in tracker#162. ([#26](https://github.com/2389-research/dippin-lang/issues/26), [#27](https://github.com/2389-research/dippin-lang/pull/27))
- **DIP135-137** lint codes for `manager_loop` validation: missing/nonexistent `subgraph_ref` (DIP135), invalid control field — negative `poll_interval` or `max_cycles` (DIP136), unbounded supervision with no `stop_condition` and no `max_cycles` (DIP137 — the manager_loop analog of DIP104).
- **`stack.*` namespace** recognized by DIP120 so `stop_condition` and `steer_condition` can reference `stack.child.cycles`, `stack.child.outcome`, `stack.child.status` without namespace warnings.
- **`dippin scaffold manager_loop`** template emits a starter supervisor workflow.
- **Tree-sitter grammar** — `manager_loop` node rule, highlights coverage, corpus test, committed generated parser (`src/parser.c` et al.), new `just tree-sitter-generate` / `just tree-sitter-test` recipes, and CI drift check so generated files can't drift from `grammar.js` without being caught.
- **VS Code TextMate grammar** — `manager_loop` keyword, new field names, `stack.*` namespace recognition.

### Fixed
- **Parser `steer_context` block-form routing** — a single-entry block-form `steer_context` (one `k: v` line under the indent) lexes without an embedded newline; the previous newline-based heuristic mis-routed it to the inline CSV handler. Replaced with a separator-position check (`:` before `=` means block, `=` before `:` means inline).

## [v0.21.0] — 2026-04-20

### Added
- **`HumanConfig.Timeout` / `TimeoutAction`** on human nodes (tracker#112). Pairs with edge labels like `when: timeout` for auto-advance semantics. Round-trips through parser, formatter, DOT export, and migrate. ([#22](https://github.com/2389-research/dippin-lang/pull/22))
- **`WorkflowDefaults` budget fields** (tracker#67): `max_total_tokens`, `max_cost_cents`, `max_wall_time`. Allow workflows to declare global budget caps consumed by the runtime. ([#22](https://github.com/2389-research/dippin-lang/pull/22))
- **Scoped context reads** — `ctx.node.<id>.*` now validates as a legitimate read pattern in DIP121/DIP122, eliminating lint false-positives for cross-node state access (tracker#75). ([#23](https://github.com/2389-research/dippin-lang/pull/23))
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
- **DIP125 false positives on shell variable assignments** (tracker#87). Replaced regex-based binary extraction with proper shell AST parsing via `mvdan.cc/sh/v3`. Variable assignments, command substitutions, and `command -v` checks are now correctly identified. Preamble commands (`mkdir`) are skipped to find the real tool binary.
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
- Tracker team field report response written (`.tracker/field-report-response-2026-03-27.md`).

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
