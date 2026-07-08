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
`(nil, false)`. It must be a small hand-parser, **not** one loose regex like
`^\^?\(([^)]+)\)\$?$` (which false-accepts `^(a.b|c)$` and would extract wrong
markers):

1. `stripAnchors`: remove exactly one leading `^` and one trailing `$` if present.
2. If the remainder starts with `(` **and ends with** `)` (a group spanning the
   *entire* remainder), the branches are `split(inner, "|")`. Otherwise the whole
   remainder is a single branch. (The full-span requirement rejects `(a|b)|c`,
   `(a|b)?`, `(?i)(a|b)`, `(a)(b)`.)
3. Each branch must be a non-empty `isLiteralToken`:
   - **empty branch → non-enumerable** (`(a|b|)`, `(|a)`, `()`); otherwise the
     empty string becomes a phantom marker no edge can spell → false DIP152.
   - a branch containing any regex metacharacter — `. * + ? [ ] { } ( ) | \ ^ $`
     (including an escaped `\.`) → non-enumerable. (`-`, `_`, `#`, apostrophes,
     unicode are **not** metacharacters, so `setup-failed`, `tests_green` are
     literals.)
4. **Do not trim intra-branch whitespace** — regex spaces are literal, so `(a | b)`
   is branches `"a "`/`" b"`; trimming would mismatch a verbatim `on a` route.
   Dedup identical branches for clean output.

Non-enumerable (`ok == false`): emit no DIP152 and preserve the current blanket
DIP101/DIP102 exemption. This never false-positives on a regex we can't fully
understand and matches 100% of real `marker_grep` values in the repo.

**Anchoring assumption:** enumeration treats the anchored group `^(a|b)$` and a
bare literal `tests_pass` as equivalent marker sets. This rests on the runtime
storing `ctx.tool_marker` as the matched *token* equal to a branch literal (which
holds for every anchored value in the corpus). Stated as an explicit assumption;
if it can't be confirmed against the runtime, restrict enumeration to fully
`^...$`-anchored forms.

## Coverage computation (ultra-conservative — no false positives)

The condition grammar is richer than equality: an edge `when` can carry
`CondOr` / `CondAnd` / `CondNot` and ops `= == != contains startswith endswith
in` (`simulate/condition.go`, `ir/edge.go`). The `on <marker>` sugar can only
express a single equality, so authors routing markers with `or`, a `!=`
catch-all, or `not ... = x` will use `when`. Computing the routed set from bare
equality alone would therefore **false-positive** on those valid workflows. Since
a lint that cries wolf is worse than none, the rule is deliberately conservative:
we only warn when routing is *simple enough to be certain* there is a gap.

For a tool node `n` with an enumerable marker set `M`, classify each outgoing
edge:

- **Simple marker route**: `ir.ExtractEqualityCondition(e)` succeeds *and* its
  variable is exactly `"ctx.tool_marker"` — add the value to the routed set `R`.
  (`ExtractEqualityCondition` only matches a bare `CondCompare` with `=`/`==`;
  it needs `Condition.Parsed`, which `Lint` guarantees by calling
  `EnsureConditionsParsed` first — `validator/lint.go`.)
- **Unconditional edge** (`e.Condition == nil`) → set `hasUnconditional`.
- **Anything else** (compound `or`/`and`, `not`, `!=`/`contains`/…, or an
  equality on a different variable) → set `hasComplexRoute`. We cannot cheaply
  prove what it covers, so its mere presence makes the node safe.

**Node is marker-safe** iff any of: a valid `else` default
(`hasValidElseDefault(w)`), `hasUnconditional`, `hasComplexRoute`, or `M ⊆ R`.
Otherwise DIP152 fires for `n`, listing the uncovered markers `M − R` (sorted,
deterministic). One diagnostic per node listing all uncovered markers (precedent:
`lint_model.go` joins a list into one diagnostic), not one per marker.

This catches the exact silent hole the issue describes — every edge a simple
`on <marker>` / `when ctx.tool_marker = <marker>` and some marker unrouted with no
`else`/fallback — while going quietly silent on any workflow whose routing we
can't fully reason about (false negatives are acceptable; false positives are
not). Reuses `ir.ExtractEqualityCondition` (moved to `ir` in #158) and the
existing `hasUnconditionalEdge` / `hasValidElseDefault` helpers.

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
  - `lintMarkerCoverage(w *ir.Workflow) []Diagnostic` — iterate nodes.
  - `checkMarkerCoverage(w, n) (Diagnostic, bool)` — per-node (mirrors
    `checkConditionalReachability`) to keep the loop under the caps.
  - `enumerateMarkers` + `stripAnchors`, `isLiteralToken`, `splitAlternation`
    (the metachar/paren handling busts ≤7 cognitive if inlined).
  - `classifyMarkerEdges(edges) (routed set, hasUnconditional, hasComplexRoute)`.
  - `uncoveredMarkers(M, R) []string` (sorted).
  - All kept in `validator` (only the validator needs them — no `ir` move, per
    YAGNI; the routed-set reuses the already-`ir` `ExtractEqualityCondition`).
- Register the pass in `LintWithOptions` (`validator/lint.go`, append after the
  last pass) and bump the `Lint` doc range comment there. Ordering is
  irrelevant — each pass is a pure function of `w`. `EnsureConditionsParsed`
  already runs first at the top of `LintWithOptions`.
- **DIP code number:** the spec assumes `DIP152`, but an unmerged sibling
  (#136) also targets DIP152. **Reconfirm the next free code against `main` at
  implementation time** (`git grep 'DIP15' -- validator/lint_codes.go`) and
  renumber to DIP153 if #136 landed first.
- DIP152 code registration + docs (the full new-code checklist — see the
  "Docs & count sweep" section; the earlier 3-item list was incomplete):
  - `DIP152 = "DIP152"` constant + `linterCodeDescriptions` entry + range
    comment in `validator/lint_codes.go`.
  - Explanation in `validator/explanations.go` `reachabilityExplanations()`
    (`Code`/`Summary`/`Trigger`/`Example`/`Fix` all non-empty) so
    `TestExplanationsCoverAllCodes` and `TestExplanationsNoExtra` pass.

Keep every function under the complexity caps (cyclomatic ≤5, cognitive ≤7);
the helper names above are the pre-planned extractions.

## Diagnostic shape

```
warning[DIP152]: node "RunTests": marker_grep enumerates markers that no edge
  routes and no else default covers: tests-failed
  = help: route each marker (e.g. `RunTests -> <node> on tests-failed`), add an
    unconditional fallback edge, or add a section `else -> <node>` default
```

Location: the node's source (`n.Source`).

## Blast radius (and the mandatory CI guard)

Only two examples declare `marker_grep`, and both are already covered:
`error_funnel.dip` (the `*-failed` markers fall to `else -> Cleanup`) and
`marker_routing.dip` (both markers routed, `M ⊆ R`). So DIP152 fires on no
current example. (An earlier draft cited `all_features.dip` — that file lives in
`parser/testdata/`, not `examples/`; the real rationale is else + routed.)

**This coverage is not enforced by any existing CI path** and must be made so:
`just lint-examples` ends in `|| true` (never fails), `just validate-examples`
runs `validator.Validate` (structural only — never `Lint`), and
`TestLintExamples` (`validator/lint_examples_test.go`) previously zero-asserted
only `DIP108`/`DIP147`. **`DIP152` is now included in that zero-assertion set** —
otherwise the "examples stay covered" guarantee would be a one-time manual
observation that silently rots the moment someone adds a marker tool to an
example. This was a required deliverable, and it shipped in the implementation.

## Docs & count sweep

Adding DIP152 changes the counts (61 → 62 codes, 51 → 52 semantic warnings,
range DIP101–DIP151 → DIP101–DIP152). Update every hardcoded occurrence (both
ASCII `-` and en-dash `–`; `git grep` for the counts/ranges):
`CLAUDE.md`, `README.md`, `AGENTS.md`, `docs/architecture.md`,
`docs/integration.md`, `docs/llm-reference.md`, `docs/validation.md` (+ a new
`### DIP152` section), `site/content/{validation,glossary,architecture,editors,
cli}.md`, `site/content/blog/editor-setup.md`, `site/static/skill.md` (diagnostics
table row + range), and the `validator/lint.go` / `lint_codes.go` range comments.
Then **regenerate the embedded spec**: `bash scripts/gen-spec.sh` and commit
`cmd/dippin/generated-spec.md` (its currency is a hard CI gate). Do **not**
hand-edit `CHANGELOG.md` / `site/content/changelog.md` (generated).

## Testing

`validator/lint_marker_coverage_test.go` — **parser-driven** (parse real `.dip`
text; never hand-set `Condition.Parsed`, per the DIP101 fixture-drift lesson):
- **Gap warns** — `^(go|stop)$` with only `on go`, no else/unconditional →
  DIP152 naming `stop`.
- **Fully routed** — both markers routed → no DIP152.
- **else-covered** — gap but a valid `else` default → no DIP152.
- **Unconditional-edge covered** — gap but an unconditional out-edge → no DIP152.
- **Multi-marker gap** — `^(a|b|c)$` with only `on a` → DIP152 lists `b, c`.
- **Bare-literal marker_grep** — `tests_pass` routed / unrouted.
- False-positive guards (the squad blockers — these must stay green):
  - **OR routing** — `when ctx.tool_marker = go or ctx.tool_marker = stop` →
    `hasComplexRoute` → **no DIP152**.
  - **`!=` catch-all** — `on go` + `when ctx.tool_marker != go` → no DIP152.
  - **`not` catch-all** — `when not ctx.tool_marker = go` → no DIP152.
  - **Non-enumerable regex** — `.*fail.*`, `^\d+$` → no DIP152.
  - **Empty branch** — `^(a|b|)$` is non-enumerable → no DIP152 (no phantom `""`).
- `enumerateMarkers` unit table: anchored/unanchored, single literal, group,
  metachar branch → reject, empty branch → reject, non-full-span group
  (`(a|b)|c`, `(a|b)?`) → reject, intra-branch whitespace preserved.
- Explanation-parity (`TestExplanationsCoverAllCodes`) — automatic.
- `TestLintExamples` extended to zero-assert DIP152 (the required guard above).
- Integration: a real `.dip` parsed through the pipeline confirms DIP152 fires.

## Known limitations (stated, not fixed)

- **Non-enumerable `marker_grep`** (complex regex) keeps the blanket exemption,
  so a genuine hole behind `.*fail.*` stays silent. This is the status quo — the
  additive framing doesn't regress it.
- **Marker routed via a non-`tool_marker` condition** (e.g.
  `when ctx.tool_stdout contains tests-failed`) is not recognized, so DIP152
  could flag `tests-failed` as unrouted. Unusual/discouraged (the point of
  `marker_grep` is `tool_marker`); the `else`/unconditional escape hatches cover
  it, and `hasComplexRoute` silences the node if that condition is the only
  routing. Documented, not fixed.
- **Expected first-release friction:** an author who routes the happy marker and
  relies on a runtime default / `route_required` (no `else`, no fallback) will
  see DIP152. This is intended — the fix is one `else`, one fallback edge, or
  routing the marker — but the issue thread should note it.
