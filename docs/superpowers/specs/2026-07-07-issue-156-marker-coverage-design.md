# Issue #156 — Coverage-aware `marker_grep` lint (DIP152)

**Date:** 2026-07-07
**Status:** Design approved
**Issue:** dippin-lang#156 (split from #154; refs epic #127, #135)

## Problem

DIP101 (`lintConditionalReachability`) and DIP102 (`lintDefaultEdge`) blanket-exempt
any tool node with a non-empty `marker_grep` via `toolHasMarkerRouting`
(`validator/lint_reachability.go:87`). The exemption is too loose: it suppresses
the warnings whether or not every marker the regex can emit is actually routed by
an edge. A marker the `marker_grep` enumerates but that no edge handles — and that
`else` does not cover — is a silent routing hole today. There is currently **no
marker-coverage lint at all**.

Example of the hole (silent today):

```
tool RunTests
  marker_grep: ^(tests-ok|tests-failed)$
edges
  RunTests -> Done on tests-ok
  # tests-failed is emitted by the tool but routed nowhere, and no else default
```

## Goal

Add **DIP152** (semantic Warning): a tool node whose `marker_grep` enumerates a
literal marker that is neither routed by an outgoing edge nor covered by a
section `else ->` default (or an unconditional edge) gets a warning naming the
uncovered marker(s). Detection only — no behavior change to routing.

## Non-goals

- Enumerating markers from regexes that are not a recognizable finite literal
  alternation (see "Marker enumeration"). Those keep today's blanket exemption.
- Tightening the DIP101/DIP102 exemption itself (DIP152 is additive — see
  "Relationship to DIP101/DIP102").
- Any runtime/engine change. `marker_grep` matching and `tool_marker` population
  are the runtime's; this is a static authoring lint.
- `route_required` interaction — orthogonal (it governs the no-marker-emitted
  case at runtime; DIP152 governs enumerated-but-unrouted markers statically).

## Marker enumeration (the conservative gate)

`MarkerGrep` is stored unquoted (e.g. `^(tests-ok|tests-failed)$`, `^(a|b)$`, or a
bare literal `tests_pass`; parser `parse_nodes.go:524`).

A helper `enumerateMarkers(markerGrep string) ([]string, bool)` recognizes **only**
the finite literal-alternation shape and returns `(markers, true)`; otherwise
`(nil, false)`:

1. Strip an optional leading `^` and trailing `$`.
2. The remainder is either:
   - a single **literal** token (bare `tests_pass`), or
   - a single parenthesized group `(x|y|z)` whose branches are all **literals**
     split on `|`.
3. A "literal" branch contains no regex metacharacters
   (`. * + ? [ ] { } ( ) | \ ^ $` beyond the outer anchors/group). Any branch
   with a metacharacter → the whole value is **non-enumerable**.

Non-enumerable (`ok == false`): emit no DIP152 and preserve the current blanket
DIP101/DIP102 exemption. This never false-positives on a regex we can't fully
understand and matches 100% of real `marker_grep` values in the repo.

## Coverage computation

For a tool node `n` with an enumerable marker set `M`:

- **Routed set** `R` = `{ v : some outgoing edge of n has condition
  ir.ExtractEqualityCondition == (variable "ctx.tool_marker", value v) }`. The
  `on <marker>` sugar already desugars to `ctx.tool_marker = <marker>`
  (`parse_edges.go:155`), so both edge spellings are captured. Reuses
  `ir.ExtractEqualityCondition` (moved to `ir` in #158).
- **Node is marker-safe** iff any of:
  - a valid `else` default exists (`hasValidElseDefault(w)`), **or**
  - `n` has an unconditional outgoing edge (`hasUnconditionalEdge`), **or**
  - every marker in `M` is in `R` (`M ⊆ R`).
- Otherwise DIP152 fires for `n`, listing the uncovered markers `M − R`
  (sorted for deterministic output).

One diagnostic per node (listing all uncovered markers), not one per marker.

## Relationship to DIP101 / DIP102 (additive, no double-warnings)

DIP101/DIP102 keep their existing blanket marker exemption
(`toolHasMarkerRouting` unchanged). Marker nodes legitimately route via the typed
`tool_marker` channel, so they stay exempt from the *generic* conditional-routing
lints. The coverage gap — previously silent — is now caught *precisely* by
DIP152. A node with a gap therefore gets exactly one warning (DIP152), never both
a DIP102 and a DIP152.

Rejected alternative: making the DIP101/DIP102 exemption coverage-aware *and*
adding DIP152 — that double-warns the same node.

## Placement & wiring

- New file `validator/lint_marker_coverage.go`:
  - `lintMarkerCoverage(w *ir.Workflow) []Diagnostic` — iterate tool nodes,
    enumerate, compute coverage, emit DIP152.
  - `enumerateMarkers` + small helpers (kept in `validator`; only the validator
    needs them — no `ir` move, per YAGNI).
- Register the pass wherever the other lint passes are aggregated (alongside
  `lintDefaultEdge` etc.).
- DIP152 code registration (the standard new-code checklist):
  - `DIP152 = "DIP152"` constant + `linterCodeDescriptions` entry in
    `validator/lint_codes.go`.
  - Explanation in `validator/explanations.go` (`Code`/`Summary`/`Trigger`/
    `Example`/`Fix` all non-empty) so `TestExplanationsCoverAllCodes` passes.

Keep every function under the complexity caps (cyclomatic ≤5, cognitive ≤7);
extract helpers (`enumerateMarkers`, `routedMarkerSet`, `uncoveredMarkers`)
rather than inlining.

## Diagnostic shape

```
warning[DIP152]: node "RunTests": marker_grep enumerates markers that no edge
  routes and no else default covers: tests-failed
  = help: route each marker (e.g. `RunTests -> <node> on tests-failed`), add an
    unconditional fallback edge, or add a section `else -> <node>` default
```

Location: the node's source (`n.Source`).

## Blast radius

All `examples/*.dip` are already covered (`error_funnel.dip` via `else`;
`marker_routing.dip` and `all_features.dip` route every marker), so no example
reconciliation is expected. Verify with `just lint-examples`.

## Testing

`validator/lint_marker_coverage_test.go`:
- **Gap warns** — `^(go|stop)$` with only `on go` and no else/unconditional →
  DIP152 naming `stop`.
- **Fully routed** — both markers routed → no DIP152.
- **else-covered** — gap but a valid `else` default → no DIP152.
- **Unconditional-edge covered** — gap but an unconditional out-edge → no DIP152.
- **Non-enumerable regex** — e.g. `.*fail.*` → no DIP152 (no false positive).
- **Bare-literal marker_grep** — `tests_pass` routed / unrouted.
- **Multi-marker gap** — `^(a|b|c)$` with only `on a` → DIP152 lists `b, c`.
- **enumerateMarkers** unit table (anchored/unanchored, single literal, group,
  metacharacter → non-enumerable).
- Explanation-parity (`TestExplanationsCoverAllCodes`) passes automatically.
- Integration: a real `.dip` parsed through the pipeline confirms DIP152 fires
  (mirrors `TestLintExamples`).
