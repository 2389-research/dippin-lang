# Issue #134 — `dip 2`: edges own destinations

**Date:** 2026-07-09
**Status:** Design approved
**Issue:** dippin-lang#134 (epic #127, Phase 1; foundation #133 shipped; proposal `docs/proposals/2026-06-11-edge-syntax-evolution.md`)

## Problem

Routing **destinations** are split between the `edges` block and node-level
fields. `retry_target` / `fallback_target` are destinations masquerading as node
config: to answer "where does this node go when it fails?" a reader reconciles up
to four places (a `when fail` edge, an unconditional catch-all, `fallback_target`,
and graph `defaults.on_failure`). In `dev_loop.dip` nearly every node carries
`fallback_target: CleanupWorktree` **and** an explicit `-> CleanupWorktree` fail
edge — the same destination asserted twice.

The honest distinction: `max_retries` / `base_delay` / `retry_policy` are
**budgets** (rightly on the node); `retry_target` / `fallback_target` are
**destinations** — edges in disguise.

## Goal

Under language format version `dip 2`, make the `edges` block the single source of
truth for destinations. `dip 2` becomes the first feature to give the version
declaration (shipped in #133) behavioral meaning; **v1 is completely unchanged**.

- Node keeps only budgets (`max_retries`, `base_delay`, `retry_policy`).
- `max_retries: N` means "re-run *me* up to N times"; the post-exhaustion
  destination is the node's `on fail` edge (a normal edge; destination = `To`).
- `retry_target` / `fallback_target` are **rejected as node fields** under v2.
- `defaults.on_failure` stays — a graph-wide default is the opposite of
  duplication.
- `fmt --migrate` converts a v1 file into an equivalent v2 file, flagging any
  case it can't express 1:1.

```dippin
# dip 2
  tool Test
    max_retries: 2          # budget only — re-run in place, then take the fail edge
  edges
    Test -> Verify  on success
    Test -> Fix     on fail   # the fallback destination
    # defaults.on_failure: EscalateReview catches anything unrouted
```

## Non-goals

- **No new syntax.** Failure destinations are expressed with the *existing*
  `on fail` edge and (for retry-to-elsewhere) the *existing* `loop` edge. We do
  NOT add `retry`/`fallback` edge attributes — that would grow the edge-attribute
  surface the epic is reducing.
- **Not** `weight:` / the cascade tiebreak / the broader cascade collapse
  (separate issue), **not** the `else ->` funnel default (done, #154), **not**
  `parallel` / `fan_in` single-sourcing (#136, the next epic item).
- **No runtime/engine change.** dippin ships the syntax + IR + spec delta;
  execution ordering is the runtime's (see "Cascade-ordering delta").

## IR — no schema change

`RetryConfig.RetryTarget` / `FallbackTarget` (`ir/ir.go:240`) **stay in the
struct**. A v1 workflow populates them (unchanged). A v2 workflow leaves them
empty: a fallback is a normal `on fail` edge (`Condition = ctx.outcome = fail`),
and a retry-to-elsewhere is a `loop` edge (`ir.Edge.Restart = true`). No new
`ir.Edge` destination fields. The one small, migration-only addition is an
optional `Comment string` on `ir.Edge` (see Migration); it is empty for all
normal parse/format paths.

The language version lives in `parser.version` (mirrored to
`ir.Workflow.Version`, set by `parseVersionDeclaration`, `parser/parser.go:44`).

## Parser — version-gated rejection

`applyCommonRetryField` (`parser/parse_nodes.go:203`) becomes version-aware (it
needs access to `p.version` and `p.diagnostics`, so the version/diagnostic sink is
threaded in, or the check is lifted to a caller that has `p`). When
`p.version >= 2` and the key is `retry_target` or `fallback_target`:

- emit a diagnostic located at the offending line: *"`retry_target` is not a node
  field in dip 2 — express the failure destination as an `on fail` edge (see
  `fmt --migrate`)"* (and likewise for `fallback_target`);
- do **not** populate the field.

The v1 path (`p.version == 1`) is byte-for-byte unchanged. `retry_policy`,
`max_retries`, `base_delay` remain node fields in both versions.

Grammar: the tree-sitter grammar is version-agnostic (these parse as generic node
fields; `on fail` / `loop` edges already exist), so **no grammar change**.

## Formatter — emission falls out for free

A v2 IR has empty `RetryTarget`/`FallbackTarget`, so `writeRetryTargetFields`
(`formatter/format.go:513`) emits nothing, and the `dip 2` line is already emitted
for `version > 1`. **No formatter change is needed to emit v2** — only the
migration transform is new (plus rendering `Edge.Comment`, below).

## `fmt --migrate` — the real work

Today `--migrate` is a v1→v1 identity pass (`cmd/dippin/cmd_fmt.go`). Add a
pre-format transform `migrateToV2(w *ir.Workflow) (*ir.Workflow, []MigrationNote)`
invoked when `--migrate` is set and the input is v1. Per node carrying a v1
target:

- **`fallback_target F`:**
  - node has no `on fail` edge → synthesize `Node -> F on fail`; clear the field.
  - node has an `on fail` edge to `F` → drop the field (dedupe; destinations agree).
  - node has an `on fail` edge to a **different** target → keep both edges, attach
    a `# MIGRATION:` note to the synthesized `-> F on fail` edge, and record a
    review note. (Ambiguous: v1 could route guard-fail and retry-exhaustion to
    different nodes; dip 2 has one fail edge — the author picks.)
- **`retry_target R`:**
  - `R == node.ID` (self) → drop (budget already means "re-run in place").
  - `R != node.ID` → synthesize `Node -> R loop` (`Restart = true`) + a review
    note (unusual); clear the field.
- Clear `RetryTarget`/`FallbackTarget`; set `w.Version = "2"`.

**Surfacing review notes (both channels):**
- **stderr summary** + a distinct exit code for "migrated, but N case(s) need
  review" (so scripts/CI can detect it), separate from clean success.
- **inline** `# MIGRATION: …` comments via the optional `ir.Edge.Comment` field,
  which the formatter renders as a leading `#` line before that edge. If per-edge
  comment rendering proves invasive, the fallback is a single header comment block
  listing the review cases — decided during implementation, but the design target
  is inline.

Never silently drop a destination. Clean files (the common case — self
`retry_target` + a matching fallback edge) migrate with no notes.

## Validator — retarget to the unified model

Extract a shared helper `nodeHasFailureRoute(w, n)` = *has a non-empty
`FallbackTarget` (bounded, per existing logic) **or** an outgoing `on fail`
(`ctx.outcome = fail`) edge*. Rewire:

- **DIP104** (unbounded retry, `lint_retry.go:12`) — currently
  `hasRetryConfig && MaxRetries==0 && FallbackTarget==""`; under v2
  `FallbackTarget` is always empty, so it must also treat an `on fail` edge as a
  bound. Use the shared helper.
- **DIP115** (goal-gate fallback, `lint_retry.go:113`) — currently reads only the
  two target fields; use the shared helper so a v2 `on fail` edge satisfies it.
- **DIP144** (agent failure route, `lint_failure_route.go`) — already reads both;
  refactor onto the shared helper for consistency (behavior-preserving).
- **DIP134** (retry/restart confusion) — unaffected.

Net effect: the failure-route lints give identical verdicts whether a failure
destination is expressed the v1 way (`fallback_target`) or the v2 way (`on fail`
edge).

## Cascade-ordering delta — documented, not modeled

dip 2's "retry in place up to N, *then* take the `on fail` edge" differs from
today's cascade, where a tier-1 `fail` edge preempts node retry. dippin's
simulator does **not** model retry execution (it resolves routing edges only), so
there is **no dippin code change** for the ordering. Document the delta in
`docs/edges.md` (and the spec regen) so the runtime can converge, per the standing
"don't gate dippin on the tracker" policy.

## Edit sites (summary)

- `ir/edge.go` — add optional `Comment string` to `Edge` (migration-only).
- `parser/parse_nodes.go` (+ `parser/parser.go` to thread `version`) — v2
  rejection of `retry_target`/`fallback_target`.
- `cmd/dippin/cmd_fmt.go` — wire `--migrate` to `migrateToV2` for v1 input;
  migrated-with-notes exit code.
- New `migrate`-side (or `formatter`-side) `migrateToV2` transform + `MigrationNote`.
- `formatter/format.go` — render `Edge.Comment` as a leading `#` line.
- `validator/lint_retry.go`, `validator/lint_failure_route.go` — `nodeHasFailureRoute`
  helper; retarget DIP104/DIP115/DIP144.
- `docs/edges.md` + spec regen; count/catalog docs unchanged (no new DIP code).

## Testing

- **Parser:** `dip 2` file with `retry_target`/`fallback_target` node fields →
  diagnostic; v1 file with the same fields → accepted, byte-identical to today.
- **Migration golden tests** (all five cases):
  1. self `retry_target` + matching `fallback` edge → clean v2, no notes.
  2. `fallback_target`, no `on fail` edge → synthesized `on fail` edge.
  3. `fallback_target` matching an existing `on fail` edge → deduped.
  4. `fallback_target` divergent from the `on fail` edge → both edges + inline
     `# MIGRATION:` note + review exit code.
  5. non-self `retry_target` → `loop` edge + note.
- **Validator:** DIP104/DIP115/DIP144 fire identically for a v1 `fallback_target`
  node and its migrated v2 (`on fail` edge) equivalent.
- **Round-trip:** run `fmt --migrate` on every `examples/*.dip` that uses these
  fields (`code_quality_sweep.dip`, `api_design.dip`, `complexity_cleanup.dip`,
  `stress_edge_cases.dip`, `ask_and_execute.dip`, `code_health_check.dip`,
  `consensus_task.dip`, `consensus_task_parity.dip`, `megaplan_quality.dip`,
  `sprint_exec.dip`); the migrated v2 output validates cleanly (or with expected
  review notes), and re-formatting it is idempotent.
- **Formatter:** a v2 workflow formats with a `dip 2` header and no
  `retry_target`/`fallback_target` lines.

## Downstream follow-up (not this issue)

- tracker `convertEdge` / engine: the retry-then-fail-edge ordering delta and the
  removal of `retry_target`/`fallback_target` from the node schema (documented
  here; coordinated separately, not gated).
