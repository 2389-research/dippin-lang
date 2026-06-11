# Proposal: Evolving Edge / Routing Syntax in Dippin

**Status:** Draft for discussion
**Date:** 2026-06-11
**Scope:** Analysis + proposal. Breaking changes are on the table, gated behind a versioned `.dip` format (`dip 2`). No implementation here.

---

## TL;DR

Routing is the part of Dippin that has accreted the most overloading. Today a reader
answering *"what happens after this node?"* may have to reconcile **four** places:
the `when` edge condition, an unconditional catch-all edge, the node's
`retry_target`/`fallback_target` fields, and the graph's `defaults.on_failure` — plus a
**five-level resolution cascade** with an invisible alphabetical tiebreak. The two
mechanisms for loops (`restart: true` edges vs `retry_target`+`max_retries`) overlap.
`label:` means three different things. And the single most common edge —
`when ctx.outcome = success/fail` — is retyped on nearly every gate.

This proposal:

1. **Diagnoses** the overloading against two *real* production workflows from downstream
   repos (`pipelines/dev_loop.dip`, 653 lines; `tracker/build_product.dip`, 2061 lines),
   not toy examples.
2. Proposes a **source-compatible Phase 0** (additive sugar + new diagnostics, no format
   version bump) that already removes most of the daily pain — `on <outcome>`, `loop`,
   `choice:`, deprecate `weight`, reject unknown attributes. "Source-compatible" rather
   than strictly non-breaking: the new diagnostics (unknown-attr rejection via #126) can
   newly *reject* a file that parses today, and a few items carry a small compatibility
   rule (below) so no currently-working routing silently changes.
3. Proposes a **versioned Phase 1 (`dip 2`)** that fixes the structural problems:
   make the `edges` block the single source of truth for *destinations*, keep only
   *budgets* on nodes, collapse the cascade to "first match in source order + explicit
   default," and make ambiguity a lint error instead of an alphabetical coin-flip.
4. Confirms this is **feasible and contained**: the architecture is IR-centric, so a
   text-syntax change barely touches validator/simulate/export/cost/diff/lsp/tracker;
   `dippin fmt` is a ready-made auto-migration vector.

The guiding values throughout: **DRY** (one source of truth per routing fact),
**YAGNI** (the cascade is richer than any real workflow uses), and **least surprise**
(the heaviest semantics should not wear the lightest syntax).

---

## 1. Evidence: routing in real workflows

Toy examples hide the problem. Two production workflows expose it. Both live in
*downstream* repos, not in dippin-lang itself —
[`2389-research/pipelines`](https://github.com/2389-research/pipelines/blob/main/dev_loop/dev_loop.dip)
and
[`2389-research/tracker`](https://github.com/2389-research/tracker/blob/main/examples/build_product.dip)
— and consume dippin as a parsed-IR dependency.

### 1a. `pipelines/dev_loop.dip` (downstream repo) — 653 lines, ~70-line edges block

The dominant pattern is the **error funnel**: nearly every node routes its failure to a
single `CleanupWorktree` handler. `CleanupWorktree` appears as an edge *target* **more
than 20 times**:

```dippin
    SetupRun              -> FetchOpenIssues   when ctx.tool_marker = setup-ok
    SetupRun              -> Exit              when ctx.tool_marker = setup-resume-required   label: resume_required
    SetupRun              -> CleanupWorktree   when ctx.tool_marker = setup-failed            label: setup_failed
    SetupRun              -> Exit              when ctx.tool_marker = setup-lock-held         label: lock_held
    FetchOpenIssues       -> PreFilter         when ctx.tool_marker = fetched-ok
    FetchOpenIssues       -> CleanupWorktree   when ctx.tool_marker = fetch-failed            label: fetch_failed
    PreFilter             -> SelectNextIssue   when ctx.tool_marker = filter-ok
    PreFilter             -> CleanupWorktree   when ctx.tool_marker = filter-empty            label: no_candidates
    PreFilter             -> CleanupWorktree   when ctx.tool_marker = filter-failed           label: filter_failed
    ...
```

Three compounding smells, all visible above:

- **The marker enum is declared twice.** The node says
  `marker_grep: "^(setup-ok|setup-resume-required|setup-failed|setup-lock-held)$"`, then
  the edges block re-lists every one of those four values as a separate
  `when ctx.tool_marker = …` edge. If the regex and the edges drift, routing breaks
  silently. (DIP103 only flags overlapping/duplicate *conditions between edges* — it does
  **not** reconcile a node's `marker_grep` enum against the set of markers the edges
  handle, so a marker the regex emits but no edge handles goes uncaught today.)
- **The same fail destination is asserted twice per node.** Most of these nodes *also*
  carry `fallback_target: CleanupWorktree` as a node field (lines 53, 79, 91, 130,
  157, 183, 223, …). So "on failure go to CleanupWorktree" is written once as a node
  attribute and again as an explicit edge. Pure DRY violation, at ~20× scale.
- **Manual column alignment.** The author hand-aligned `->`, `when`, and `label:` into
  columns with runs of spaces. When a syntax needs hand-built ASCII tables to stay
  readable, the syntax is too noisy.

### 1b. `tracker/build_product.dip` (downstream repo) — 2061 lines

Here the split-brain reaches its worst. `TestMilestone`'s failure routing is expressed
in **four** places:

```dippin
  tool TestMilestone
    retry_target: TestMilestone          # node field (1)
    fallback_target: EscalateMilestone   # node field (2)
  ...
  edges
    TestMilestone -> VerifyMilestone   when ctx.outcome = success
    TestMilestone -> EscalateMilestone when ctx.tool_stdout contains escalate
    TestMilestone -> FixMilestone      when ctx.outcome = fail     # explicit fail edge (3)
    TestMilestone -> EscalateMilestone                             # unconditional catch-all (4)
```

…and the graph sets `defaults.on_failure: EscalateReview` as a fifth, global layer.
To know where `TestMilestone` goes when it fails, a reader must hold all of these *plus*
the 5-level failure cascade in their head simultaneously.

The file is littered with **paragraph-length comments whose only job is to explain
cascade interactions** — for example:

> *"The trailing unconditional `-> EscalateMilestone` is the strict-failure catch-all
> (mirrors TestMilestone): conditional edges are tried first, so it only ever fires on
> an unexpected outcome — and it keeps the node's printed markers covered."*

and

> *"stop/abandon are listed BEFORE continue so the freeform labels[0] fallback (if
> `default` were ever dropped) is safe."*

When authors must write a paragraph to justify *edge ordering* and *dead catch-all
edges added only to satisfy a lint*, the model is too subtle. The unconditional
catch-all edges are semantically dead in the common case but **required** to silence
DIP102 and keep marker coverage green — i.e. syntax-driven cargo-culting.

`label:` is simultaneously doing all of its jobs in this one file: display
(`label: revise`), human-gate choice key (`ApprovePlan`/`OperatorDecision`/
`EscalateMilestone`/`EscalateReview` all use `label: "approve"`, `"adjust"`,
`"stop"`, `"abandon"`, …), and failure-reason audit tag (`label: setup_failed`).

---

## 2. Diagnosis — three root causes

Stripping the symptoms down, there are three structural causes.

### Cause A — Destinations masquerade as node config (split-brain)

Routing **destinations** live in two places: edges (`when … -> X`) *and* node fields
(`retry_target`, `fallback_target`, `defaults.on_failure`). The honest distinction is:

- `max_retries`, `base_delay`, `retry_policy`, `max_restarts` are **budgets** — how
  many times / how long. These legitimately belong on the node.
- `retry_target`, `fallback_target`, `on_failure` are **destinations** — they are edges
  in disguise.

Destinations-as-node-fields is the category error that creates the split-brain. The fix
is not "move everything onto nodes" or "move everything into edges" wholesale — it is
**one locus per concern**: destinations are edges; budgets are node config.

### Cause B — Two overlapping loop/retry mechanisms

`restart: true` edges (bounded by `max_restarts`) and `retry_target`+`max_retries`+
`fallback_target` both express "on failure, go somewhere up to N times, then give up."
`build_product.dip` uses both; `dev_loop.dip` mixes `restart: true` back-edges with
`fallback_target`. Two mechanisms, one job (DRY).

The clean separation by *intent*:
- **Node-local retry** = re-run the *same* node (budget: `max_retries`). The author
  shouldn't name a target at all — the target is "this node."
- **Workflow loop** = jump back to an *earlier stage* (budget: `max_restarts`). This is
  a genuine back-edge and deserves visible syntax.

### Cause C — The resolution cascade is richer than any workflow uses

`docs/edges.md` documents a 5-level routing cascade *and* a separate 5-level failure
cascade:

| # | Routing cascade | In real use? |
|---|---|---|
| 1 | First matching `when` condition | **Yes — the only tier anyone relies on** |
| 2 | Handler `PreferredLabel` match | Invisible at author time; not exercised in examples |
| 3 | Handler `SuggestedNextNodes` | Invisible at author time; not exercised |
| 4 | Highest `weight` | **`weight:` appears in no `examples/*.dip` and neither cited production workflow** (only in parser test fixtures + docs) |
| 5 | Alphabetically-first target ID | **Footgun** — renaming a node silently re-routes |

Of five tiers, authors can predict exactly one (conditions). Two are runtime handler
outputs invisible in the source. One (`weight`) is unused. One (lexical) is silent
action-at-a-distance: two equal edges resolve by *spelling*. This is textbook YAGNI
debt — speculative machinery the language doesn't need, plus a determinism rule that
produces 2am debugging.

`label:` overloading (display / human-choice key / `PreferredLabel` routing signal) is a
fourth, smaller conflation that rides on top of cause C.

---

## 3. Design principles

Applied directly from our working style:

1. **One source of truth per routing fact** (DRY). A node's fate after success and
   after failure should be readable in *one* place.
2. **Sugar the 80% case, keep the escape hatch** (pragmatism). `when <expr>` stays for
   arbitrary conditions; the pervasive `ctx.outcome`/`ctx.tool_marker` equality gets a
   shorthand.
3. **Heavy semantics get heavy syntax** (least surprise). A back-edge that clears
   downstream state and can fail the whole run should not look like `weight: 5`.
4. **Delete speculative tiers** (YAGNI). If no workflow uses `weight` or the handler
   tiers, they are liabilities, not features.
5. **Ambiguity is an error, not a coin-flip.** Replace the alphabetical tiebreak with a
   lint diagnostic.
6. **Don't throw away what works.** The `edges` block's single-place-to-see-topology
   property is a genuine strength (DOT export, reachability analysis, diffing all read
   it naturally). Keep it; fix what's broken around it.

---

## 4. Proposal

Deliberately phased. Phase 0 is non-breaking and delivers most of the daily-readability
win cheaply. Phase 1 is the versioned structural fix.

### Phase 0 — Source-compatible sugar + new diagnostics (no version bump)

Old syntax keeps parsing and `dippin fmt` can rewrite to the new forms. Two caveats keep
this honest rather than "strictly non-breaking": (1) the unknown-attr diagnostic (#126)
can newly reject files with typo'd attributes; (2) where a Phase 0 item could otherwise
change the meaning of a working file (`choice:`, `weight`), it carries an explicit
compatibility rule so routing behavior is preserved until `dip 2`.

**0.1 `on <token>` — outcome/marker shorthand.**
`on X` desugars to a comparison against the node's *natural outcome channel*:
`ctx.outcome` for agent/human nodes, `ctx.tool_marker` for tool nodes that declare
`marker_grep`. This single sugar transforms both real files:

```dippin
# before
SetupRun -> FetchOpenIssues   when ctx.tool_marker = setup-ok
SetupRun -> CleanupWorktree   when ctx.tool_marker = setup-failed   label: setup_failed
QualityGate -> WriteReport    when ctx.outcome = success
QualityGate -> Synthesize     when ctx.outcome = fail   restart: true

# after
SetupRun -> FetchOpenIssues   on setup-ok
SetupRun -> CleanupWorktree   on setup-failed   label: setup_failed
QualityGate -> WriteReport    on success
QualityGate -> Synthesize     on fail   loop
```

`when <expr>` remains for everything non-equality (`ctx.tool_stdout contains escalate`,
`and`/`or`, etc.). Sugar for the common case, full power retained.

**0.2 `loop` keyword — replaces `restart: true`.**
A back-edge is the heaviest construct in the language; give it a word, not a boolean
flag buried among attributes. `loop` reads as what it is. (`restart: true` still parses;
`fmt` rewrites it.)

**0.3 `choice:` — split human-gate keys from display labels.**
Human-gate matching gets a dedicated `choice:` key so the routing key is no longer
conflated with display text. **Compatibility rule (Phase 0):** a v1 edge with only
`label:` keeps routing exactly as today — `label` still populates the choice key when
`choice:` is absent, so existing `Approve -> Ship label: "yes"` workflows are unaffected.
The clean split — `label:` becomes display-*only* — lands in `dip 2`, where `fmt --migrate`
rewrites routing-bearing `label:` into `choice:`. After that split, deleting a `label:`
is provably safe and deleting a `choice:` is provably load-bearing.

```dippin
ApprovePlan -> PickNextMilestone  choice: "approve"
ApprovePlan -> Decompose          choice: "adjust"   loop
ApprovePlan -> Done               choice: "reject"
```

**0.4 Soft-deprecate `weight:`.** Emit a lint (`DIPxxx: weight is unused by routing`) but
**keep parsing and preserving it** — `weight` is part of the documented v1 routing
priority, so a Phase 0 `fmt` must not strip it (an external workflow could rely on
weighted tie-breaking even though no `examples/*.dip` or cited production workflow does).
Actual removal — and shrinking the cascade by a tier — is deferred to `dip 2`, where
`fmt --migrate` drops it.

**0.5 Reject unknown edge attributes (already in flight via #124 pt 4).**
The parser's `applyEdgeAttribute` switch has no `default`, so unknown attributes are
silently swallowed (the `override` bug #124 documents). A `default` that emits a
diagnostic closes the whole class of silent-drop bugs and is a precondition for evolving
the attribute set safely.

> **Phase 0 alone** rewrites essentially every edge in both production workflows into a
> shorter, single-meaning form, with no format version bump and `fmt`-driven migration.

### Phase 1 — `dip 2`: structural fixes (versioned, breaking)

Gated behind a format declaration so old files keep working and `fmt --migrate` upgrades
them mechanically (see §5).

**1.1 Edges are the single source of truth for destinations.**
Remove `retry_target` and `fallback_target` as node fields. A node keeps only its
*budget* (`max_retries`, `base_delay`, `retry_policy`). The destinations move to edges:

- node-local retry needs **no destination** — `max_retries: N` on the node means
  "re-run me up to N times"; the runtime re-executes in place.
- the post-exhaustion destination is just the node's `on fail` edge.
- `defaults.on_failure` **stays** — it is genuinely a graph-wide *default*, and it
  belongs in `defaults`. It is the one acceptable "second place," because a default that
  every node inherits is the opposite of duplication.

```dippin
# dip 1 (today): destination split across node + edges + defaults
tool TestMilestone
  retry_target: TestMilestone
  fallback_target: EscalateMilestone
edges
  TestMilestone -> VerifyMilestone   when ctx.outcome = success
  TestMilestone -> FixMilestone      when ctx.outcome = fail
  TestMilestone -> EscalateMilestone                            # dead catch-all

# dip 2: budget on node, destinations as edges, no duplication
tool TestMilestone
  max_retries: 2          # re-run in place up to twice, then take the fail edge
edges
  TestMilestone -> VerifyMilestone  on success
  TestMilestone -> FixMilestone     on fail
  # graph defaults.on_failure: EscalateReview catches anything unrouted
```

**1.2 Collapse the cascade to two predictable tiers + explicit default.**

New routing resolution, in full:

1. First edge whose `on`/`when` guard matches, **in source order** (top-to-bottom, the
   way humans read).
2. The unconditional edge, if present (the explicit default/fallback).

That's it. Removed: `weight` (gone in 0.4), the two invisible handler tiers, and the
alphabetical tiebreak. **Two outgoing edges that can both match with no source-order
intent become a lint error** (`DIPxxx: ambiguous routing`), not an alphabetical
silent pick. Failure routing folds into the same order: an `on fail` edge is just a
guard match; node `max_retries` is consulted before the fail edge fires; `on_failure` is
the graph default; then halt.

> Note: tiers 2–3 are *runtime handler* behaviors, so their removal is a contract change
> the **engine** must honor, not something dippin enforces alone (per our standing rule:
> ship the dippin syntax/IR now, don't gate on the tracker runtime — but flag the
> semantics delta in the changelog/spec so the engine can converge).

**1.3 The error-funnel default (the biggest single win — needs the most design care).**

Both real files funnel ~20–25 edges into one handler. The mechanical cause: a tool node
*succeeds* (exit 0) but emits a `…-failed` marker, so the engine does **not** treat it as
a node failure — the author must route every failure marker by hand, and each carries a
near-mechanical audit `label:` (`setup_failed` = node + `_failed`).

Two candidate directions (I am **not** picking silently — this is the open design
question):

- **(a) A section-level `else ->` default.** One line at the bottom of the `edges`
  block: `else -> CleanupWorktree` means "any node whose guards all fail, and which has
  no explicit unconditional edge, routes here." Collapses ~25 edges to one. Cost: it is a
  second implicit-routing mechanism (mild tension with 1.1's "one source of truth"), and
  it interacts with DIP102/marker-coverage lint, which today *forces* the explicit
  catch-alls.
- **(b) Marker-classified routing.** Let `marker_grep` declare which markers are
  terminal-failure (e.g. a `fail_markers:` set), so the engine treats them as node
  failures and they flow through `on fail` / `on_failure` like everything else — no
  per-marker edge needed. Cost: more node config; couples the tool node to routing.

My lean is **(a)** for its simplicity and because it generalizes beyond tools, but this
is the part to prototype on `dev_loop.dip` before committing.

**1.4 Single-source `parallel`/`fan_in`.**
Today `parallel P -> a, b, c` declares the fan-out, and the `edges` block frequently
**re-declares** `P -> a`, `P -> b`, `P -> c`. The inline list should be authoritative;
re-declaration becomes redundant (lint) or forbidden. *(Confirm against the engine
whether the edges-block entries are load-bearing or purely echoed before changing.)*

---

## 5. Migration & blast radius (feasibility)

A versioned breaking change is **feasible and contained** because Dippin is IR-centric:
almost every consumer programs against `ir.Edge`, not `.dip` text.

**Versioning.** `ir.Workflow.Version` already exists (`ir/ir.go:13`) but is *dead* —
never read or emitted. Reuse it: add a `version: 2` workflow-header field (or a `dip 2`
line-1 declaration if we want a true bootstrap switch), wired in `parser/parser.go`
(~line 102) + `formatter/format.go` `writeWorkflowHeader`. The lexer needs no change.

**`fmt` is the migration vector.** `dippin fmt` already parses to IR and re-emits
canonical text (`cmd/dippin/cmd_fmt.go` → `parser.Parse()` → `formatter.Format()`). A
`fmt --migrate` mode = parse v1 → emit v2. Phase 0 sugar and the 1.1 field→edge reshuffle
are largely **IR-preserving** (the destinations already exist in IR as
`RetryConfig.RetryTarget/FallbackTarget`), so migration is mostly an emit-side change.
The genuinely new IR work is small: moving destinations from `RetryConfig` onto `Edge`.

**Blast radius per component:**

| Component | Coupled to edge *text*? | Size |
|---|---|---|
| `parser/parse_edges.go`, `lexer.go` | Yes — concrete syntax | **M** |
| `editors/tree-sitter-dippin/grammar.js` | Yes — *independent reimplementation*, must stay in lockstep; regen via `npx tree-sitter generate` | **M** |
| `formatter/format.go` | Yes — the migration emitter | **S–M** |
| `simulate/condition.go` | Only if the *condition expression* grammar changes (it doesn't here; `on` is edge-framing sugar) | **S** |
| `ir/edge.go` | Only the destination-field move | **S** |
| `validator/` (DIP003/004/005/009/101/102/103/105/144 + new ambiguity/weight lints) | Reads IR, not text; new codes need explanations + parity test | **S** |
| `export/`, `cost/`, `coverage/`, `graph/`, `diff/`, `lsp/`, `dipx/` loader | Read IR — **no logic change**; dipx bundles embed raw `.dip` so need re-`fmt` | **S/none** |
| `migrate/` (DOT→dip) parity | Produces IR; only golden `.dip` fixtures change | **S** |
| **`tracker` (downstream)** | Consumes parsed IR via its adapter `convertEdge`, **never `.dip` text** | **none** unless an IR field moves — then a one-line adapter change |

The two real costs are the **hand-written parser** and the **duplicated tree-sitter
grammar** (treat parser + `grammar.js` as one atomic change). Everything downstream of IR
is fixture/bundle re-formatting at most. Spec is doc-assembled
(`scripts/gen-spec.sh`), so grammar updates are markdown edits + a script run.

---

## 6. Considered and rejected

**Fully colocated routing tables (XState / LangGraph style):** put each node's outgoing
routing on the node as an `on: { success: X, fail: Y }` table; drop the `edges` block.
This is the strongest cross-system pattern and tempting. **Rejected** because it
sacrifices the `edges` block's single-place-to-read-topology property, which DOT export,
reachability/coverage analysis, and `diff` all rely on — and which is genuinely valuable
on a 2061-line graph. We get ~80% of the colocation benefit (locality of a node's fate)
from Phase 1.1 (budgets on node, one fail edge) without losing the global view. Pragmatism
over purity.

**Numeric conditions / typed values:** conditions are string-equality only
(`count = 0` is a string compare), and `consensus_task.dip` leans on
`ctx.internal.loop_restart_count != 0`. Tightening this is real but **orthogonal** to
edge *structure* and out of scope here; worth its own proposal.

---

## 7. Recommendation

- **Do Phase 0 now.** It is non-breaking, `fmt`-migratable, and removes the bulk of the
  daily readability tax (`on`, `loop`, `choice:`, deprecate `weight`, reject-unknown
  already landing via #124). Sequence it right after #124 merges.
- **Prototype Phase 1.3 (`else ->`)** against `dev_loop.dip` before anything else in
  Phase 1 — the error funnel is the single biggest line-count win and the least
  settled design.
- **Then commit Phase 1** behind `dip 2` with `fmt --migrate`, coordinating the cascade
  semantics change with the engine team (don't gate dippin on it; document the delta).

## Open questions

1. Error funnel: `else ->` default route (6/1.3a) vs marker-classified failure (1.3b)?
2. Does the engine actually consume the duplicated `parallel`/`fan_in` edges-block
   entries, or are they echoes safe to forbid (1.4)?
3. `on <token>` for tool nodes keys off `ctx.tool_marker` — what is the fallback when a
   tool node has *no* `marker_grep`? (Probably: `on` is only legal where the node has a
   defined outcome channel; otherwise require `when`.)
4. Cascade change is a runtime contract: what is the rollout order between dippin's spec
   and the tracker engine so a `dip 2` file never runs under `dip 1` semantics?
