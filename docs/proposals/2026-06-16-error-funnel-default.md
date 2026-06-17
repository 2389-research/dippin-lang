# Decision: The error-funnel default — `else ->` over marker-classified failure

**Status:** Decided (design spike). Recommends **(a) section-level `else ->`**; opens the
implementation issue.
**Date:** 2026-06-16
**Scope:** Design spike for [#135](https://github.com/2389-research/dippin-lang/issues/135)
(epic [#127](https://github.com/2389-research/dippin-lang/issues/127), Phase 1.3). Produces a
**decision**, not an implementation. No production Go, no new lint, no parser/formatter change
ships here — those land under the implementation issue this doc opens.
**Parent proposal:** [`2026-06-11-edge-syntax-evolution.md`](./2026-06-11-edge-syntax-evolution.md)
§4 "Phase 1.3".

---

## TL;DR — the recommendation

**Adopt (a): a single section-level `else -> <handler>` default in the `edges` block.**
**Reject (b) `fail_markers:`** as the funnel mechanism (keep it on the table only as an
*optional, engine-side* root-cause fix that needs no dippin syntax).

On a faithful 6-node funnel fixture, (a) collapses the failure funnel from **7 hand-routed
`-> Cleanup` edges (each with an audit `label:`) to 1 line**, taking the whole `edges` block
from **14 edges → 8** and deleting **all 7 audit labels**. It generalizes across tool, agent,
and human nodes; it lives entirely inside the `edges` block, so it stays consistent with
[#134](https://github.com/2389-research/dippin-lang/issues/134) ("edges own destinations")
instead of fighting it; and it has the smaller, cleaner runtime contract of the two.

(b) is genuinely appealing because it fixes the *semantic root cause* (a tool exits 0 yet emits
a `…-failed` marker). But as **dippin syntax** it is the worse choice: it scatters per-node
`fail_markers:` config across the file, re-declares part of the `marker_grep` enum, only covers
the *failure* subset of unmatched markers, and pushes routing intent back onto the node just as
#134 is trying to pull destinations off nodes. The root-cause concern it raises is real and is
captured below as a runtime-contract note — the engine *may* reclassify markers as failures on
its own, with no `.dip` surface.

This unblocks #134 by settling failure flow into **two non-competing channels** (see §6): a
*genuine node failure* takes the retry budget → `on fail` guard → `defaults.on_failure`, while a
*non-failure unmatched outcome* (the funnel case) takes the node's unconditional edge → section
`else`. `else` never intercepts a hard failure — exactly the contract `on fail` edges and
`defaults.on_failure` must agree with.

---

## 1. The problem (one paragraph — full framing is in the parent proposal §1, §4)

The dominant line-count cost in the two production workflows
([`dev_loop.dip`](https://github.com/2389-research/pipelines/blob/main/dev_loop/dev_loop.dip),
~653 lines; [`build_product.dip`](https://github.com/2389-research/tracker/blob/main/examples/build_product.dip),
~2061 lines) is the **error funnel**: nearly every node routes its failure to one shared handler
(`CleanupWorktree`, `EscalateMilestone`), written one edge at a time — `-> CleanupWorktree`
appears 20+ times in `dev_loop.dip`. The mechanical cause is verified: a **tool node succeeds
(exit 0) but emits a `…-failed` marker**, so the engine does *not* treat it as a node failure.
`defaults.on_failure` never fires (the node "succeeded"), so the author hand-routes every failure
marker as its own edge, each carrying a near-mechanical audit `label:` (`setup_failed` =
node-name + `_failed`).

> **Fixture note.** The two real workflows live in downstream repos and are **not present in
> this repo**, so the prototypes below run against a **condensed in-repo fixture** that exhibits
> the exact pattern. The baseline is embedded as a fenced `dippin` block (below), **not** added
> to `examples/` — it is illustrative, so it never enters the `just validate-examples` /
> `TestLintExamples` set and needs no validation. The (a)/(b) rewrites use syntax that does not
> exist yet (`else ->`, `fail_markers:`) and therefore *cannot* be real `.dip` files; they are
> illustrative blocks only. The edge-count comparison is **baseline (real pattern, counted) vs
> proposed (illustrative)**.

---

## 2. Baseline fixture — "today, verbose"

A faithful, condensed funnel: six `tool` nodes, each `marker_grep`-ing an `…-ok` / `…-failed`
pair (some with an extra non-failure outcome), each succeeding on exit 0, each hand-routing its
`…-failed` marker to a single shared `Cleanup` handler with a per-marker audit label. Marker
mechanics mirror the in-repo canonical [`examples/marker_routing.dip`](../../examples/marker_routing.dip).

```dippin
workflow ErrorFunnelBaseline
  goal: "Condensed dev-loop funnel: every tool failure marker hand-routed to Cleanup"
  start: SetupRun
  exit: Done

  defaults
    model: claude-sonnet-4-6
    provider: anthropic

  tool SetupRun
    marker_grep: "^(setup-ok|setup-resume-required|setup-failed)$"
    command: ./setup.sh

  tool FetchIssues
    marker_grep: "^(fetch-ok|fetch-failed)$"
    command: ./fetch.sh

  tool PreFilter
    marker_grep: "^(filter-ok|filter-empty|filter-failed)$"
    command: ./prefilter.sh

  tool RunBuild
    marker_grep: "^(build-ok|build-failed)$"
    command: ./build.sh

  tool RunTests
    marker_grep: "^(tests-ok|tests-failed)$"
    command: ./test.sh

  tool Package
    marker_grep: "^(pkg-ok|pkg-failed)$"
    command: ./package.sh

  tool Cleanup
    marker_grep: "^(cleanup-done)$"
    command: ./cleanup.sh

  agent Done
    prompt: "Summarize the run."

  edges
    # --- happy path + distinct non-failure outcomes (stay explicit) ---
    SetupRun   -> FetchIssues  when ctx.tool_marker = setup-ok
    SetupRun   -> Done         when ctx.tool_marker = setup-resume-required   label: resume_required
    FetchIssues-> PreFilter    when ctx.tool_marker = fetch-ok
    PreFilter  -> RunBuild     when ctx.tool_marker = filter-ok
    RunBuild   -> RunTests     when ctx.tool_marker = build-ok
    RunTests   -> Package      when ctx.tool_marker = tests-ok
    Package    -> Done         when ctx.tool_marker = pkg-ok

    # --- the error funnel: 7 near-identical hand-routed failure edges ---
    SetupRun   -> Cleanup      when ctx.tool_marker = setup-failed            label: setup_failed
    FetchIssues-> Cleanup      when ctx.tool_marker = fetch-failed            label: fetch_failed
    PreFilter  -> Cleanup      when ctx.tool_marker = filter-empty            label: no_candidates
    PreFilter  -> Cleanup      when ctx.tool_marker = filter-failed           label: filter_failed
    RunBuild   -> Cleanup      when ctx.tool_marker = build-failed            label: build_failed
    RunTests   -> Cleanup      when ctx.tool_marker = tests-failed            label: tests_failed
    Package    -> Cleanup      when ctx.tool_marker = pkg-failed              label: pkg_failed

    Cleanup    -> Done         when ctx.tool_marker = cleanup-done
```

**Baseline counts:** the six funnel **source** nodes emit **14 outgoing edges** (7 happy/outcome +
7 failure) — of which **7 are funnel `-> Cleanup` edges** carrying **7 audit `label:` tags**. (The
counts throughout this doc are edges *out of the funnel nodes*; they exclude the terminal
`Cleanup -> Done` edge, which is the handler's own exit and is unchanged by either design.) This is
the DRY tax the spike targets.

> **A finding that corrects the proposal's framing.** The parent proposal (§1b, §4 Phase 1.3)
> says the explicit catch-alls are "required to silence DIP102 and keep marker coverage green."
> Reading the real lint logic, that is **not** what forces these specific edges. `lintDefaultEdge`
> (DIP102) and `lintConditionalReachability` (DIP101) **blanket-exempt any tool node with a
> non-empty `marker_grep`** via `toolHasMarkerRouting` (`validator/lint_reachability.go:79`,
> reached through `nodeIsSafeRouter:151` and `sourceIsSafe:65`). So none of the funnel tool
> nodes above are flagged by DIP102/DIP101 today, with or without a catch-all — and **no
> marker-coverage lint exists at all** (the proposal §1a itself notes a marker the regex emits
> but no edge handles "goes uncaught today"). What actually forces the 7 edges is the **runtime
> semantics**: an exit-0 `…-failed` marker matches no guard, so without an explicit edge the node
> has no route and stalls (or, with a stray unconditional edge, falls to the lexical tiebreak
> that DIP149 warns about). The lint interaction is therefore not "stop DIP102 false-positives"
> but "**tighten the blanket marker exemption** so it does not hide genuinely unrouted markers,
> and teach it that `else` is a valid catch-all." Details in §5.

---

## 3. Prototype (a) — section-level `else -> Cleanup`

One line at the bottom of the `edges` block. Semantics: *any node whose guards all fail to match,
and which has no explicit unconditional edge of its own, routes here.*

```dippin
  edges
    # --- happy path + distinct non-failure outcomes (unchanged) ---
    SetupRun   -> FetchIssues  on setup-ok
    SetupRun   -> Done         on setup-resume-required
    FetchIssues-> PreFilter    on fetch-ok
    PreFilter  -> RunBuild     on filter-ok
    RunBuild   -> RunTests     on build-ok
    RunTests   -> Package      on tests-ok
    Package    -> Done         on pkg-ok
    Cleanup    -> Done         on cleanup-done

    # --- the entire error funnel, collapsed ---
    else -> Cleanup
```

(Shown with the Phase-0 `on <marker>` sugar from the parent proposal §4.0.1, which is orthogonal
but compounds the win.)

**Before/after:**

| Metric | Baseline | (a) `else ->` | Δ |
|---|---|---|---|
| Funnel failure edges | 7 | **1** | −6 |
| Audit `label:` tags | 7 | **0** | −7 |
| Edges out of funnel nodes (excl. terminal `Cleanup -> Done`) | 14 | **8** | −6 |

**Readability.** The 7 mechanically-named labels (`setup_failed`, `fetch_failed`, …) vanish —
they were only ever node-name + `_failed`, pure noise that `else` makes structural. Each tool node
now shows *only its meaningful branches* (the happy path and any genuine alternate outcome like
`setup-resume-required`); "everything else is a failure → Cleanup" is stated **once**, where a
reader looks for it. `PreFilter`'s `filter-empty` (a non-error "no work" outcome that the baseline
hand-routed to Cleanup) is simply dropped — it now falls through `else` to the same place, which
is the intended behavior, expressed by absence rather than a labeled edge. At the scale of the real
`dev_loop.dip` (20+ funnel edges), this is the single largest line-count win in the epic.

**Generality.** `else` is not tool-specific. An agent node whose `on success` / `on fail` guards
don't cover an unexpected outcome, or a human gate with an unhandled choice, all fall through the
same `else`. (b) cannot do this — it only reclassifies *tool markers*.

---

## 4. Prototype (b) — `fail_markers:` on the tool node

Each tool declares which markers are terminal failures; the engine then treats them as node
failures that flow through the existing failure path (`on fail` edge / `defaults.on_failure`).

```dippin
  defaults
    model: claude-sonnet-4-6
    provider: anthropic
    on_failure: Cleanup        # all reclassified failures funnel here

  tool SetupRun
    marker_grep: "^(setup-ok|setup-resume-required|setup-failed)$"
    fail_markers: setup-failed          # ← reclassify as node failure
    command: ./setup.sh

  tool FetchIssues
    marker_grep: "^(fetch-ok|fetch-failed)$"
    fail_markers: fetch-failed
    command: ./fetch.sh

  tool PreFilter
    marker_grep: "^(filter-ok|filter-empty|filter-failed)$"
    fail_markers: filter-failed         # but NOT filter-empty (not an error)
    command: ./prefilter.sh

  # …RunBuild, RunTests, Package each gain a fail_markers: line…

  edges
    SetupRun   -> FetchIssues  on setup-ok
    SetupRun   -> Done         on setup-resume-required
    FetchIssues-> PreFilter    on fetch-ok
    PreFilter  -> RunBuild     on filter-ok
    PreFilter  -> Cleanup      on filter-empty        # STILL hand-routed: not a failure
    RunBuild   -> RunTests     on build-ok
    RunTests   -> Package      on tests-ok
    Package    -> Done         on pkg-ok
    Cleanup    -> Done         on cleanup-done
    # the 6 *-failed edges are gone; defaults.on_failure: Cleanup catches them
```

**Before/after:**

| Metric | Baseline | (b) `fail_markers:` | Δ |
|---|---|---|---|
| Funnel failure edges | 7 | **1** (`filter-empty`, still not a failure) | −6 |
| Audit `label:` tags | 7 | **0** | −7 |
| New per-node `fail_markers:` lines | 0 | **6** | +6 |
| New `defaults.on_failure` line | 0 | 1 | +1 |
| Net new/removed source lines | — | ≈ **break-even** | ~0 |

**Readability.** (b) removes the failure *edges* but spends the savings re-declaring failure
markers on each node — and those declarations **partially duplicate `marker_grep`** (the regex
already enumerates `setup-failed`; `fail_markers:` lists it again). Routing intent is now split
across three places (the `marker_grep` regex, `fail_markers:`, and `defaults.on_failure`), the
exact split-brain the epic is trying to *remove*. Worse, `fail_markers:` only covers the *failure*
subset: `PreFilter`'s `filter-empty` is a non-error "no candidates" outcome, so it is **not** a
fail marker and **still needs a hand-routed edge** — (b) leaves a ragged half-funnel. The
line-count is roughly break-even, versus a clean −6 for (a).

**The one thing (b) does better:** it fixes the *root cause*. After (b), an exit-0-`…-failed`
marker really is a node failure, so it composes with retry budgets and `on fail` uniformly with a
genuine exit-≠0 crash. That is a real semantic win — but it is a **runtime** property, and the
engine can adopt it without any `.dip` syntax (see §6). dippin does not need `fail_markers:` to get
the funnel collapse; `else` already delivers it with less surface.

---

## 5. DIP102 + marker-coverage interaction (precise rule changes)

Grounded in the real logic in `validator/lint_reachability.go` and
`validator/lint_condition_types.go`.

**Today (verified):**

- **DIP102** (`lintDefaultEdge`, `lint_reachability.go:126`) flags a node with conditional
  outgoing edges but no unconditional default — **unless** `nodeIsSafeRouter` (`:151`) returns
  true, which it does when conditions are exhaustive *or* `toolHasMarkerRouting` (`:79`) is true,
  i.e. the tool has any non-empty `marker_grep`.
- **DIP101** (`lintConditionalReachability`, `:16`) flags a node reachable only by conditional
  edges — suppressed by the same `sourceIsSafe` / `toolHasMarkerRouting` path (`:65`).
- **There is no marker-coverage lint.** Nothing reconciles a node's `marker_grep` enum against the
  set of markers its edges handle (the parent proposal §1a confirms this gap).

The consequence: for the funnel tool nodes, `marker_grep` presence is a **blanket exemption** —
DIP101/DIP102 stay silent whether or not every marker is routed. That is *too loose*: it already
hides unrouted markers today.

**What each lint must learn for (a):**

1. **DIP102 — treat a node as having a default when a section `else ->` covers it.** Add the
   section default to the "does this node have a fallback?" test. Concretely, in `nodeIsSafeRouter`
   (or its caller in `lintDefaultEdge`), a node is safe if `w` declares a section `else` target and
   the node has no competing unconditional edge — the `else` *is* its unconditional default. This is
   the direct analog of the existing `hasUnconditionalEdge` check (`:92`), lifted from per-node to
   per-graph, mirroring how `defaults.on_failure` is already a graph-wide default.

2. **Tighten the blanket marker exemption into a coverage-aware one (the marker-coverage rule the
   proposal wants).** Replace "tool has any `marker_grep` ⇒ safe" with "tool's `marker_grep`
   markers are **all routed** (each has a guard edge) **or** an `else`/unconditional edge covers the
   remainder." When `else` exists, every unrouted marker is covered by construction, so the node is
   safe *and correctly so* — closing the existing false-negative instead of widening it. This is new
   logic (a small `markerSet(node) ⊆ routedMarkers(node) ∪ {else}` check), and it belongs to the
   **implementation issue**, not this spike.

3. **DIP101** needs the symmetric change in `sourceIsSafe` (`:65`): a source node is safe if its
   unmatched markers route through `else`, so a handler reachable only via the funnel (e.g.
   `Cleanup`) is not falsely flagged once its incoming per-marker edges are replaced by `else`.

**For (b):** no DIP102/coverage *rule* change is needed for the failure markers themselves — once a
`fail_marker` is a node failure, the node's failure route is governed by the **DIP144**
(`lintAgentFailureRoute`, `lint_failure_route.go:14`) family / `defaults.on_failure`, not DIP102.
But (b) introduces a **new** validation obligation: `fail_markers:` values must be a subset of the
`marker_grep` enum (a typo'd `fail_marker` that the regex can never emit is dead config) — a new
lint of its own. And the blanket `toolHasMarkerRouting` exemption still needs tightening
independently, because (b) does nothing for non-failure unrouted markers like `filter-empty`.

Net: **(a)'s lint change is smaller and strictly corrective** (it tightens an exemption that is
already too loose, and reuses the existing unconditional-default machinery), whereas **(b) adds a
brand-new lint and still leaves the looseness to fix separately.**

---

## 6. Runtime-contract implications

Per standing policy, dippin ships the syntax/IR + spec delta now and **never gates on the tracker
runtime** (see `never-gate-dippin-on-tracker`); this section documents the engine delta, it does
not make dippin wait.

**(a) `else ->`.** The engine's routing resolution gains one terminal step, and `else` is the
**non-failure** default — it must **not** intercept a genuine node failure on its way to
`defaults.on_failure`. This is not a free choice: the existing failure-route lint already encodes
the rule that *"an unconditional/success edge does NOT count — a hard failure does not traverse
it"* (`hasFailEdge`, `validator/lint_failure_route.go:65`). `else` is exactly such an
unconditional success-side edge, so a hard failure must skip it. The two channels are therefore
resolved **separately, by node result**:

- **Genuine node failure** (tool exit ≠ 0, agent/human error): retry budget → first matching
  `on fail` guard → **`defaults.on_failure`** → halt. `else` is **not** in this path.
- **Non-failure outcome with no matching guard** (the funnel case — incl. an exit-0 `…-failed`
  marker, which the engine sees as *success*): node's own unconditional edge → section **`else`**
  → halt.

So `else` (fires on any unmatched *non-failure* outcome) and `defaults.on_failure` (fires only on a
genuine node failure) have **distinct triggers and never compete** — a failure can never be
swallowed by `else`. This is a small, local addition to the resolver — one lookup of a
graph-level field on the success side — with no change to how a marker is classified.

> *(Earlier drafts of this section listed a single linear cascade with `else` ahead of
> `defaults.on_failure`; that ordering would have let a hard failure with no `on fail` guard hit
> `else` first and bypass the dedicated failure handler — the bug flagged by Codex's P2 review.
> The two-channel resolution above is the corrected contract, and it is also the
> failure-routing contract #134 must adopt.)*

**(b) `fail_markers:`.** The engine must, after a tool exits 0, check whether the emitted marker is
in the node's `fail_markers` set and, if so, **reclassify the node result as a failure** before
routing — feeding the existing failure path (retry budget → `on fail` edge → `defaults.on_failure`).
This is a deeper change (it touches result *classification*, not just routing) and it must stay in
lockstep with retry/budget semantics. It is the more invasive of the two contracts.

**Independent of the syntax choice:** the root-cause observation behind (b) — that an exit-0
`…-failed` marker is morally a failure — is something the **engine may adopt on its own** (treat a
configured marker convention as a failure) with **no `.dip` surface at all**. Recommending (a) does
not foreclose that; it just declines to spend dippin syntax on it.

---

## 7. Interaction with #134 ("edges own destinations")

[#134](https://github.com/2389-research/dippin-lang/issues/134) makes the `edges` block the single
source of truth for *destinations* and pulls `retry_target`/`fallback_target` off nodes. The funnel
decision must not reintroduce node-side routing.

- **(a) is on the right side of #134.** `else -> Cleanup` is *an edge* — it lives in the `edges`
  block, names a destination there, and is the block's graph-scoped default, exactly as
  `defaults.on_failure` is the graph-scoped failure default that the parent proposal §4.1.1 already
  blesses as "the one acceptable second place." The parent proposal flagged a "second implicit
  routing mechanism" tension for (a); the prototype dissolves it: `else` is *explicit* (written in
  the block) and merely graph-scoped rather than node-scoped, so it adds **no new locus on the
  node** and reads as "the edges block's success-side default arm." It composes cleanly with
  `on fail` (a guard edge, tried first) and `defaults.on_failure` (the failure-side default, in a
  separate channel that `else` never intercepts — see §6).

- **(b) pushes against #134.** It puts new routing-relevant config (`fail_markers:`) **back onto the
  node**, partially re-coupling node and routing at the same moment #134 is decoupling them. Not
  fatal (it is classification, not a destination), but it is swimming upstream.

This is why settling on (a) **unblocks #134**: failure flow has a single, edge-resident answer that
`on fail` + `defaults.on_failure` + the `else` default already agree on.

---

## 8. Prior art (confirming, not re-researching — see parent proposal §4)

Both reference points favor (a) for an indentation DSL:

- **XState** uses ordered guarded transitions with an explicit unguarded default arm — first match
  wins, last unguarded entry is the fallback. `else ->` is precisely this: source-ordered guards +
  one explicit default.
- **LangGraph** conditional edges take an exhaustive `key → node` map; the framework expects every
  key handled or a default supplied. The corresponding discipline here is the marker-coverage
  tightening in §5 (every marker routed, or covered by `else`).

Neither idiom reaches for "reclassify an output value as an error" (b) as the routing primitive;
both model the funnel as **explicit default + exhaustiveness check**, which is (a).

---

## 9. Recommendation

**Adopt (a) `else ->`.** It wins on every axis the spike measured: largest line-count collapse
(7 funnel edges + 7 labels → 1 line; 14 → 8 edges on the fixture, scaling to 20+ → 1 on the real
`dev_loop.dip`), generality across node kinds, a smaller and strictly-corrective lint change, the
lighter runtime contract, and full consistency with #134. The parent proposal's lean toward (a) is
**confirmed** by the prototype.

**Reject (b) `fail_markers:` as dippin syntax.** Its only advantage is a root-cause semantic fix
that (i) the engine can adopt with no `.dip` surface and (ii) (a) does not need. As syntax it is
break-even on lines, re-duplicates the marker enum, covers only the failure subset, adds a new lint,
and re-couples node to routing against #134's grain.

This is a single, actionable decision: implement the section-level `else ->` default and the
paired DIP102/marker-coverage tightening.

---

## 10. Implementation plan → issue

Tracked in **[#154](https://github.com/2389-research/dippin-lang/issues/154)** (opened from this
spike). Edit-site sketch for the implementer:

- **Parser** (`parser/parse_edges.go`): recognize a section-level `else -> <node>` entry in the
  `edges` block; store as a graph-level default edge on the IR (new `ir.Workflow.ElseTarget` or an
  `Edge` with an `Else` flag — implementer's call). Reject more than one `else` per block.
- **tree-sitter** (`editors/tree-sitter-dippin/grammar.js`): mirror the `else` edge production;
  regen via `npx tree-sitter generate` (kept in lockstep with the hand parser).
- **Formatter** (`formatter/format.go`): emit the `else ->` line last in the `edges` block.
- **Lint** (`validator/lint_reachability.go`): per §5 — (1) `else` counts as a node's default in
  DIP102 (`nodeIsSafeRouter`/`lintDefaultEdge`); (2) replace the blanket `toolHasMarkerRouting`
  exemption with a coverage-aware check (markers all routed, or covered by `else`); (3) symmetric
  DIP101 (`sourceIsSafe`) update. New explanations + parity test per the project's lint-code
  checklist.
- **Docs/spec** (`docs/edges.md` + `scripts/gen-spec.sh` source docs): document `else ->`, the
  updated resolution cascade (§6), and the coexistence of `else` vs `defaults.on_failure`.
- **Runtime contract**: changelog/spec delta only — the engine adds the `else` resolution step
  (§6). Do **not** gate dippin on it.

Out of scope for that issue: `fail_markers:` (rejected), `dip 2` edges-own-destinations (#134),
single-source `parallel`/`fan_in` (#136).
