# Issue #76 — DOT export silently drops block-form parallel branches

**Date:** 2026-05-28
**Issue:** #76
**Unblocks:** #58 (per-branch `tool_access` override)

## Problem

A `parallel` node can be authored two ways:

- **inline:** `parallel P -> A, B` → populates `ParallelConfig.Targets`
- **block:** `parallel P` + indented `branch: A` / `model:` / `provider:` / `fidelity:`
  → populates `ParallelConfig.Branches []ir.BranchConfig`

`ir.BranchConfig` carries `Target, Model, Provider, Fidelity`. The parser always
populates **both** `Targets` and `Branches` in block form (`parser/parse_nodes.go`
`parseParallelBlock` → `branchTargets`), with `Targets[i] == Branches[i].Target`
in the same order.

The DOT path ignores `Branches`:

- `export/dot.go` `applyParallelAttrs` writes only `targets=`, never the branches.
- `migrate/migrate.go` `buildParallelConfig` reads only `targets=`.

Consequences through `dippin export` / `dippin simulate` DOT output:

- per-branch `model`/`provider`/`fidelity` are silently lost; and
- if the IR has `Branches` but empty `Targets` (e.g. hand-built IR or a future
  caller), the exported node loses its fan-out entirely — a structural break.

The `.dip → .dip` path is fine (formatter `writeParallelBlock` + simulate read
`Branches`). The gap is specific to the DOT path.

## Goal / acceptance criteria

1. A block-form parallel node survives `dippin export` → DOT → (migrate) → `.dip`
   without losing its fan-out targets.
2. Per-branch `model`/`provider`/`fidelity` round-trip through DOT (full
   round-trip, not warn-on-loss — confirmed with the user).
3. A test covers a block-form-only parallel node through the DOT round-trip.

## Decision: full round-trip via an encoded `branches=` node attribute

DOT here is dippin's own dialect (the runtime consumes inlined `.dipx`, not DOT), so
the migrate parser can read back any encoding we choose. We mirror the existing
`steer_context` precedent (`manager_loop`), which flattens a structured value into
a percent-encoded string for a lossless DOT round-trip.

### Encoding

`applyParallelAttrs` emits, for a parallel node:

- **`targets=`** — always, from `cfg.Targets`; if `Targets` is empty but
  `Branches` is present, derived from each `branch.Target`. Co-emitted for
  backward compatibility (old binaries and tools that read only `targets=`).
- **`branches=`** — whenever `cfg.Branches` is non-empty (even if every branch is
  target-only — required for `.dip→DOT→.dip` idempotence; see Idempotence below).

Each branch is encoded as uniform `k=v` tokens joined by `;`:

- `target` is always present (first).
- `model` / `provider` / `fidelity` only when non-empty.

Branches are joined by `,`, in slice order (order is semantically meaningful —
it maps positionally to targets). Example node attribute:

```dot
split [shape=component,
  targets="fast,accurate",
  branches="target=fast;model=claude-haiku-4-5;provider=anthropic;fidelity=summary,target=accurate;model=claude-opus-4-7;provider=anthropic;fidelity=full"];
```

### Reserved-character safety

Five characters are percent-encoded inside keys and values (`%`→`%25`, `,`→`%2C`,
`;`→`%3B`, `=`→`%3D`, `\`→`%5C`). This is a **superset** of the `steer_context`
encoder (which reserves only `% , =`): `;` is the new within-branch field
separator, and `\` is encoded so a value containing a literal backslash followed
by `n`/`l`/`r` cannot survive the outer DOT-quote layer as a DOT escape sequence
(`\n`/`\l`/`\r`) and get decoded to a newline by the migrate lexer. Encode order
is irrelevant: `strings.NewReplacer` is single-pass and the percent-encoded forms
are mutually exclusive (none is a prefix of another), so an already-substituted
token can never recombine into another reserved char.

We do **not** reuse `steerContextEncoder` — adding `;`/`\` to it would change the
`steer_context` wire format and break its round-trip tests. Instead each package
gets a dedicated `branchEncoder`/`branchDecoder` `strings.Replacer` pair, mirroring
the steer pattern (encoder in `export`, decoder in `migrate`; the two packages
cannot share private helpers and already duplicate the steer encoder/decoder
across the boundary — this is the established, accepted norm).

The outer DOT-quote layer (`dotQuote` / lexer `readString`) still handles `"` and
newlines unambiguously, so those are intentionally **not** in the percent set
(only `\` is, to defeat DOT escape-sequence interpretation). A code comment notes
this so a future editor doesn't assume the percent set is self-sufficient.

### Decoding

`buildParallelConfig`:

- Parse `branches=` into `[]ir.BranchConfig` (split on `,`, then `;`, then the
  first `=`; percent-decode each token).
- `branches=` is the **source of truth for target order**: when present, set
  `cfg.Targets` from the branch order (reproducing the parser invariant
  `Targets[i] == Branches[i].Target`); `targets=` is advisory and not trusted to
  re-derive order when `branches=` exists.
- When `branches=` is absent, behave as today: read `targets=`, and let
  `inferParallelFanIn` backfill `Targets` from edges if absent.

**No `TrimSpace` on branch values.** Targets are node IDs that must match edge
endpoints exactly; `steer_context`'s `TrimSpace` (it handles advisory hints) would
corrupt them. Trim only around the top-level `;`/`,` structural boundaries.

**Malformed-token policy:** skip unknown *keys* leniently (mirrors steer's
non-erroring contract; supports hand-edited DOT), but a branch token lacking a
non-empty `target` is dropped as a whole branch rather than producing a
`BranchConfig{Target:""}` (which would corrupt the edge mapping). `buildParallelConfig`
stays non-erroring (returns a value, not an error).

## Critical correctness requirements (from design review)

These were surfaced by the expert design review and are mandatory:

### Targets always populated; inference must not clobber Branches

Every consumer reads `ParallelConfig.Targets`, never `Branches` (validator
reachability `validate.go:137`, DIP007 `validate.go:343`, in-degree
`lint_context.go:208`, simulate). Only DIP114 `checkBranchFidelities`
(`lint_style.go:139`) reads `Branches`.

Therefore migrate must populate `Targets` from branch order whenever `branches=`
is present (matching the parser). Additionally, `inferParallelTargets`
(`migrate/migrate.go`) currently rebuilds the config as
`ir.ParallelConfig{Targets: targets}` when `Targets` is empty — **dropping any
`Branches`**. Harden it to preserve `cfg.Branches` on rebuild:
`ir.ParallelConfig{Targets: targets, Branches: cfg.Branches}`. With migrate always
setting `Targets` for branch nodes, this path won't fire for them, but the rebuild
is a latent trap given the new field and must be closed.

### Parity check must compare Branches

`migrate/parity.go` `compareParallelConfigs` compares only
`strings.Join(ac.Targets, ",")`. This is the exact blind spot that allowed the
bug: `dippin migrate --check` reports PASS even when per-branch config is dropped.
Extend it to compare `Branches` field-by-field (Target/Model/Provider/Fidelity),
following the `compareToolConfigs` → scalar/slice decomposition. Without this the
fix is unverifiable through the parity tool and can silently regress.

## Idempotence

The formatter chooses block vs inline form on `len(cfg.Branches) > 0`
(`writeStructuralNode`). For `.dip → DOT → .dip` to be idempotent:

- A block-form node must come back with `Branches` populated **and** `Targets`
  derived in branch order — matching a fresh block-form parse.
- An inline node (no branches) must come back with `Branches == nil` (not an empty
  slice) so the formatter keeps inline form. `buildParallelConfig` only sets
  `Branches` when the `branches=` attr is present.
- A block-form node whose branches are all target-only still emits `branches=`
  (no "skip if no overrides" optimization), so it round-trips as block form rather
  than being downgraded to inline.

## Backward / forward compatibility

- **Old dippin reading new DOT** (`branches=` present): the DOT parser stores
  unknown attrs in a generic map and never rejects them; old `buildParallelConfig`
  reads only `targets=` (always co-emitted). Graceful degradation — loses
  per-branch config, which is the pre-fix status quo. No parse error.
- **New dippin reading old DOT** (no `branches=`): `Branches` stays nil, existing
  edge inference reconstructs `Targets`. Works.

## Complexity decomposition (cyclo ≤ 5, cog ≤ 7; no `//nolint`)

Verified against the repo's gocyclo/gocognit baselines for the steer_context
functions. Branches add one nesting level over steer_context, so helper extraction
is mandatory.

### Export (`export/dot.go`)

| Function | Responsibility |
|---|---|
| `applyParallelAttrs(attrs, cfg)` | dispatch to the two helpers below |
| `applyParallelTargetsAttr(attrs, cfg)` | emit `targets=` from `cfg.Targets`, else derive via `parallelBranchTargets(cfg.Branches)` |
| `applyParallelBranchesAttr(attrs, cfg)` | guard `len(Branches)>0`; set `attrs["branches"]=encodeBranches(...)` |
| `encodeBranches([]BranchConfig) string` | map each branch → `encodeBranch`, join with `,` |
| `encodeBranch(BranchConfig) string` | build `;`-joined tokens; `target` always, others via `appendBranchField` |
| `appendBranchField(parts, key, val) []string` | append `key=encodeBranchToken(val)` only when `val != ""` |
| `encodeBranchToken(s) string` | percent-encode `% , ; =` via the `branchEncoder` Replacer |
| `parallelBranchTargets([]BranchConfig) []string` | local copy of parser's `branchTargets` |

### Migrate (`migrate/migrate.go`)

| Function | Responsibility |
|---|---|
| `buildParallelConfig(attrs)` | set `Targets`; if `branches=` present, `Branches = parseBranches(v)` and derive `Targets` from branch order |
| `parseBranches(string) []BranchConfig` | split on `,` → `parseBranchToken`; drop branches with empty target |
| `parseBranchToken(string) BranchConfig` | split on `;` → `applyBranchToken` |
| `applyBranchToken(*BranchConfig, token)` | split first `=`, skip guard, switch on key, `decodeBranchToken` value. **4-case switch is at cyclo 5 — zero headroom**; comment it, go table-driven if a 5th field is added (#58). |
| `decodeBranchToken(s) string` | percent-decode via the `branchDecoder` Replacer |

`compareParallelConfigs` (parity.go) split into `compareParallelTargets` +
`compareParallelBranches` if needed to stay under caps.

## Testing (TDD — failing test first)

1. **Export unit test** — block-form `ParallelConfig` → DOT contains correct
   `targets=` and `branches=` (assert exact byte layout, incl. the empty-Targets
   derivation path).
2. **Migrate unit test** — `branches=` attr → `[]BranchConfig`, asserting both
   `Targets` (order-preserved) and each branch's fields.
3. **Round-trip (acceptance)** — block-form-only `.dip` → parse → `ExportDOT` →
   `Migrate` → assert `Branches` (Target/Model/Provider/Fidelity) and `Targets`
   preserved. Assert field-level equality, not the encoded string.
4. **Reserved-char edge case** — each of `% , ; =` (and a value already containing
   `%`) in target/model/provider/fidelity round-trips. Mirror the steer_context
   reserved-char test in `migrate/roundtrip_test.go`.
5. **Edge cases** — target-only branch; targets-only (no branches) stays inline
   (`Branches == nil`); malformed `branches=` token tolerated (skipped, no error).
6. **Parity test** — extend the parallel parity tests so a per-branch-field
   difference is detected (guards C2 against regression).

Run via `just test-pkg export`, `just test-pkg migrate`, `just complexity`, then
`just check` before commit.

## Documentation

Extend `docs/nodes.md` (currently inline-only ~line 344) with the block-form
`parallel` section and a `branches=` DOT-attribute note mirroring the
`steer_context` percent-encoding row. `docs/GRAMMAR.ebnf` already specifies the
block form — no change. **Do not touch `CHANGELOG.md`** — that is a tag-time step.

## Parallel start/exit recovery (in scope — added during execution)

A parallel node that is also the workflow start/exit has its shape overridden to
Mdiamond/Msquare on export, so on import `resolveStartExitKind` must recover
`NodeParallel` (it already recovers `NodeManagerLoop`/`NodeTool` via
`hasManagerLoopAttrs`/`hasToolConfigAttrs`). Add a `hasParallelAttrs` check
(detecting the `branches`/`targets` keys, which are unique to parallel nodes) so a
parallel start/exit node round-trips its config instead of degrading to
`NodeAgent`.

This was originally scoped out as a pre-existing gap, but the canonical block-form
example (`parser/testdata/parallel_branches.dip`) uses `start: split` — the
parallel node *is* the entry point — so without this, the most realistic block-form
pattern would not satisfy acceptance criterion #1. The fix is small and mirrors the
existing kind-recovery helpers.

## Out of scope

- **No new `examples/*.dip`.** Not in the acceptance criteria; the round-trip is
  proven by the tests above. (If one were added it would need top-level placement
  and DIP108-valid per-branch models — deliberately omitted.)
- **Fan-in start/exit recovery.** A `fan_in` node that is also start/exit is not
  recovered (`sources` is not checked). Pre-existing, not triggered by this change,
  and not part of #76.
