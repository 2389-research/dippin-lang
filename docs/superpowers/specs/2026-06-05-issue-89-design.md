# Issue #89 — DIP146: cross-file effective `tool_access` detection across the subgraph boundary

**Date:** 2026-06-05
**Closes:** [#89](https://github.com/2389-research/dippin-lang/issues/89)
**Deferred from:** [#59](https://github.com/2389-research/dippin-lang/issues/59) (DIP143, the single-file advisory)
**Arc:** #41 `tool_access` → #75 `writable_paths` → #79 TOCTOU resolver → #59 subgraph boundary advisory (DIP143) → **#89 cross-file resolution (DIP146)**
**Review:** Design forks decided by author; spec hardened by a 5-expert review panel (security, Go-architecture, static-analysis DX, correctness/algorithms, conventions/prior-art) — finding tags inline below.

## Problem

A `manager_loop` node supervises a child pipeline via `subgraph_ref`, and a plain
`subgraph` node embeds one via `ref`. Both point at a **separate `.dip` file**.
`tool_access` is a per-node primitive (`AgentConfig.ToolAccess`, `BranchConfig.ToolAccess`)
with no workflow-level policy and no field on the boundary node itself. So an author who
locks agents down (`tool_access: none`) in the parent file gets **no** guarantee about the
referenced child — the child's agents are governed entirely by their own file. The
restriction silently stops at the file boundary.

DIP143 (#59) flags this boundary but **deliberately does not open the child** — the
`validator` package may not import `parser` (CLAUDE.md layering rule) and compiles to wasm
(no filesystem). DIP143 is therefore a per-file **advisory Hint**: it fires on every boundary
where the parent shows containment intent, *guessing* that the child might be under-restricted
without ever confirming it.

**This issue is the residual:** actually resolve the referenced child, parse it, inspect its
agents' effective `tool_access`, and emit a precise finding **only when the child genuinely
lacks any tool restriction** the parent's posture implies — while never letting "I couldn't
fully check" masquerade as "checked and safe."

## Decisive constraints (the architectural crux)

1. **`validator` cannot do this.** It may not import `parser` and is compiled to wasm
   (`GOOS=js GOARCH=wasm`), where there is no filesystem to read the child from. DIP143 lives
   in `validator` precisely because a per-file, wasm-safe lint is the most it can say.

2. **Only the loader tier / CLI may cross files.** Per CLAUDE.md, only the **CLI may import
   *both* `parser` and `validator`** (`dipx` may import `parser` but is forbidden from
   importing `validator`). Pack-time structural validation is already **invoked at the CLI
   layer, not inside `dipx`** (`cmd/dippin/cmd_pack.go:validateEntryPrePack`). The cross-file
   traversal therefore lives in the **CLI layer (`cmd/dippin`)** — the one place allowed to
   compose `parser` + `validator` + `ir`, and which is never built for wasm.

### Decision: DIP146, a registered code emitted from the CLI

The finding surfaces as a real DIP code, **DIP146**, so it completes DIP143's story in the
same catalog (`dippin explain DIP146`, generated spec, docs) and stays consistent with all 54
existing codes. The twist:

- The **const + explanation live in `validator`** (`lint_codes.go` + `explanations.go`) — pure
  data, wasm-safe, no new imports. A code comment at the `DIP146` const records that it is
  **intentionally CLI-emitted, not reachable from `validator.Lint()`** (so a future maintainer
  doesn't "fix" it). *(Arch I2, Security M4)*
- The code is **emitted from a native CLI pass** in `cmd/dippin`, never from `validator.Lint()`.

This is safe: `TestExplanationsCoverAllCodes` / `TestExplanationsNoExtra` only require
`CodeDescription` ↔ `Explanations` parity. **No test asserts a registered code is reachable
from `Lint()`** (verified: `explanations_test.go`, `lint_test.go`, `validate_test.go` all
assert per-case expected codes, never "all registered codes are emitted"). `dippin explain`
reads solely from `validator.Explanations` (`cmd_explain.go`), so `dippin explain DIP146`
works regardless of Lint-reachability. *(Arch I2, Conventions I3)*

*Rejected alternative — a bare CLI finding with no DIP code:* avoids touching docs/count
strings, but is inconsistent with the diagnostic model (every other finding is a DIPxxx), is
not discoverable via `dippin explain`, and does not visibly close the DIP143 → #89 loop.

## Design — DIP146, severity `Hint`

### The `tool_access` model (verified)

Binary: `""` = full tool catalog (the **default**), any non-empty value = restricted (the
runtime fails closed — even a typo `nono` yields no-tools, per the #41 design and the
`AgentConfig.ToolAccess` doc comment, `ir/ir.go:113`). Branch `""` = *inherit the target
agent's value*; the only recognized branch values are `""` and `none`, so a branch can never
grant **more** than its target agent. Therefore **agent-or-branch containment intent** is the
right unit: "the child restricts nothing" means no agent and no branch declares a non-empty
`tool_access`. *(Security M1, Correctness M3)*

This is a fact about the **documented runtime contract**, not something dippin enforces;
DIP146 detection presumes the tracker honors fail-closed (consistent with the
`never-gate-dippin-on-tracker` rule).

### Per-child classification

For each resolved child workflow, classify its tool-access posture (reusing the exported
predicate, below):

| Child posture | Definition | Outcome |
| --- | --- | --- |
| **zero-intent** | child has ≥1 agent **and** no agent/branch restricts | **DIP146 fires** (the confirmed gap) |
| **full-restrict** | every agent is restricted (no tool-bearing agent remains) | **silent** — confirmed safe |
| **partial-audit** | some agents restricted, ≥1 tool-bearing agent open | **no DIP146; DIP143 retained** (honest "audit this") |
| **agent-less** | child has zero agents (all tool/human nodes) | **silent** — no tools to grant *(DX M2)* |
| **unresolved** | child missing / unparseable / `.dipx` / depth-capped | **no DIP146; DIP143 retained** *(Security C1/C2)* |

The crux decision (security panel C1/C2): the conservative DIP146 rule must **not** be paired
with blanket DIP143 suppression. A *partial-audit* child — restricts agent A but leaves
tool-bearing agent B open — is the single most dangerous shape; it must never go silent. So
DIP143 is suppressed **only** for `zero-intent` (replaced by the precise DIP146) and
`full-restrict`/`agent-less` (confirmed safe). For `partial-audit` and `unresolved`, the
honest DIP143 boundary Hint is **retained**. "Unknown/partial" never reads as "checked & safe."

The CLI only **filters and appends**: it drops `validator`'s DIP143 where the pass classified
the boundary's child as zero-intent/full-restrict/agent-less, and appends DIP146 findings. It
never *constructs* a DIP143 (no message duplication). *(Arch I3)*

### The intent gate — ancestor-path

DIP146 fires at a boundary edge `parent → child` when:

1. **containment intent exists somewhere on the path** from the linted entry down to this
   boundary — at least one workflow on that path has an agent/branch with non-empty
   `tool_access`; **and**
2. the child is **zero-intent** (table above); **and**
3. the boundary's child file **resolved and parsed**.

Implementation: thread an `intentSeen` boolean down the DFS, OR-ing in each visited workflow's
intent as you descend. This **subsumes** entry-only gating (the entry's intent is on every
path) and per-edge gating (the immediate parent's intent is on the path), and closes the
security hole where an *intermediate* parent declares a real restriction the entry does not.
Same complexity as entry-only. *(Security I1, DX I2)*

When **no** workflow on the path restricts anything, the author has shown no tool-restriction
posture, so DIP146 stays silent — the gate that keeps un-restricting workflows noise-free.

### Traversal, transitivity, cycles & termination

- **Full transitive DFS** over every `manager_loop`/`subgraph` boundary, parsing each child and
  recursing. Non-propagation is transitive, so a one-hop check (DIP143) is insufficient — this
  is the headline upgrade.
- **Per-edge emission, per-node recursion (decoupled).** The zero-intent check and DIP146
  emission happen for **every boundary edge** at the edge's own boundary node. The visited-set
  gates only **recursion** (whether to descend into a child's *own* boundaries). So a diamond
  (`P1 → C` and `P2 → C`, both un-restricted on the path) yields **two** DIP146 findings — one
  at each editable boundary site — while `C`'s subtree is walked once. Conflating the two would
  under-report. *(Correctness I4)*
- **Symlink-resolving visited set.** Keyed by `filepath.EvalSymlinks` (with `filepath.Abs`+
  `Clean` fallback when EvalSymlinks errors). A child is only recursed into after it parsed, so
  the file exists and `EvalSymlinks` is safe. Lexical `Abs+Clean` alone misses symlink cycles
  (`a.dip → link-to-a.dip → a.dip`) → infinite recursion. Do **not** reuse `dipx.Canonicalize`
  — it is a *bundle-path* validator (rejects absolute paths, requires a `workflows/` prefix),
  semantically wrong for arbitrary CLI file arguments; reuse the *pattern*, not the function.
  *(Correctness C1)*
- **Recursion depth cap** (mirror `dipx`'s `maxDepth`, `resolve.go:215`). The visited set bounds
  *distinct files* but not depth on a legitimately deep acyclic chain. On exceeding the cap,
  stop descending **silently** (the child becomes `unresolved` → DIP143 retained at depth-1);
  #89's deliverable is tool-access detection, not a new depth/cycle diagnostic. *(Correctness C2)*
- **Pre-order, pre-seeded.** The entry's canonical path is seeded into the visited set before
  traversal, and each child is marked visited **before** recursing into it (pre-order). This is
  the load-bearing termination invariant — post-order marking would recurse `A → A` forever.
  Tested directly. *(Correctness I1/M4)*
- **No separate cycle diagnostic.** `dipx` already rejects ref cycles at pack time
  (`ErrRefCycle`, in a separate pass — its loose-file walker itself only has a visited set,
  same as here). Cycle *reporting* is out of scope.

### `.dipx` bundle inputs — out of scope for v1

`loadWorkflow` accepts `.dipx` and returns `dipx.Load(...).Entry()`, whose `subgraph_ref`/`ref`
are **in-bundle paths**, not files resolvable on disk from `node.Source.File`. The disk-based
traversal would either no-op (every child "unresolved") or parse a stale same-named file. So
the cross-file pass is **skipped entirely for `.dipx` entries** (DIP143 stands, consistent with
wasm). Resolving children through the bundle's verified-bytes map is a possible follow-up.
*(DX C1)*

### Superseding DIP143 in the native `lint` path

`validator.Lint()` emits DIP143 per-file (it can't see across the boundary) for the entry's
direct boundaries. The cross-file pass reuses the **same `*ir.Workflow`** `CmdLint` already
parsed once (`cmd_validate.go:63`) — never a re-parse — so the boundary node's
`ir.SourceLocation` recorded by the pass is **byte-identical** to DIP143's `Location`. The pass
returns `(diags []validator.Diagnostic, classified map[ir.SourceLocation]childPosture)`. Then:

- `CmdLint` drops a DIP143 iff `classified[d.Location]` ∈ {zero-intent, full-restrict,
  agent-less}; otherwise (partial-audit / unresolved / not-classified) it is **retained**.
- Append the DIP146 findings.

Critical invariants the pass must honor *(Correctness I2/I3, Arch I3, DX C2)*:
- Supersession keys come from the **same `w.Nodes`** Lint saw — never a re-parse, never a
  path-normalized `Source` (else `entry.dip` vs `/abs/entry.dip` mismatch → every DIP143
  leaks). A regression test lints via a **relative** path and asserts DIP143 is dropped.
- **Skip boundaries with `node.Source.File == ""`** (stdin/in-memory) — they can't resolve a
  child and would insert a zero-value `SourceLocation` key that collides with every other
  zero-located DIP143.

Result, native `lint`: DIP146 where there's a confirmed gap; **silence** where the child is
fully restricted or agent-less; **DIP143 retained** for partial-audit and unresolvable
boundaries (at the entry's direct depth). In wasm / the playground, DIP143 stands alone,
exactly as today. `validator` still owns DIP143 entirely (unit tests unaffected); only
`CmdLint`'s composition changes — an accepted, test-pinned coupling (the integration test
guards the location join against a future DIP143 location change).

**Known limitation (documented):** DIP143 is a per-file lint that runs only on the entry, so a
**partial-audit or unparseable child at depth > 1** (behind an already-audited intermediate)
is **not** separately flagged — the retained-DIP143 backstop covers the entry's direct
boundaries only. Deeper boundaries get DIP146 (for zero-intent gaps) but no partial/unresolved
advisory. Recorded as a follow-up. *(Security I2)*

### Severity: `Hint` (not Warning)

A child may be **intentionally** fully open (a trusted tool-running worker pool the parent just
orchestrates) — and the conservative gate fires *precisely* on that pattern (a fully-tooled
worker sets no restriction → zero-intent). Reading the child raises confidence in the **fact**
("child restricts nothing"), but **not** in the **defect** ("this is a problem") — a true
fact is frequently a non-defect intent. The repo has **no per-diagnostic suppression
mechanism**, so a Warning that fires on such a correct-but-intentional child would be corrosive
— it would train authors to ignore the whole DIP139–146 safety band. `Hint` matches the
established "context-dependent, may be fine" precedent (DIP125/131/133, DIP143). DIP146's value
over DIP143 is **precision** (it fires only on a confirmed zero-intent gap, and silences
audited children), not louder severity. If a suppression mechanism ever lands,
`DIP146 → Warning` is the natural upgrade. *(DX I1/I3, Security M3)*

### Where it runs

`dippin lint` only — DIP143's home, where authors already see the boundary advisory.
`validate` stays structural-only (DIP001–DIP009) per its contract. `pack` already rejects ref
cycles and could host DIP146 later, but its pre-pack gate is structural-errors-only today;
adding a Hint-level cross-file pass there is out of scope for v1.

### Message, help & `explain` discipline

The message must **not** claim the child "inherits" anything, and must **not** imply that a
child setting `tool_access: none` is therefore fully safe — the supervisory/steering channel
(`SteerContext`, `stack.child.*`) is information-flow, a **separate** concern (#56). Draft:

> **DIP146** (Hint): `manager_loop "Supervise" delegates to subgraph "child.dip", which
> declares no tool_access restriction on any agent; a workflow on this path restricts tools,
> but the restriction does not cross the subgraph boundary.`
> **help:** `Give child.dip's agents their own tool_access (e.g. tool_access: none on
> summarizers). This bounds the child's tool catalog, not information flow across the
> supervisory boundary (see #56). Multiple boundaries referencing the same child each get a
> Hint; one tool_access edit in the child clears them all.`

Located at the **boundary node** in the referencing file (`node.Source`) — each hop's finding
points at an editable site in a real file. Per-boundary-edge (not per-child) emission is
deliberate: each finding has its own editable site, preserving DIP143's discipline; no dedup.
*(DX M1)*

The DIP146 **`Explanation` `Trigger`** must state, so the firing is self-explanatory: (a) it
fired because a workflow **on the path** declares tool_access intent (the ancestor-path gate —
explains the "grandchild surprise"); (b) it traversed **transitively**; (c) it **read and
confirmed** the child has zero restriction (the precision win over DIP143). DIP143's and
DIP146's `explain` text **cross-reference** each other and state the native-vs-wasm split
("DIP143 is the filesystem-free advisory; native `lint` resolves boundaries and upgrades them
to DIP146 or silence"). DIP143's help line "Cross-file enforcement is tracked as #89" is
updated to reflect #89 shipped as DIP146. *(DX I4/I5, M3)*

## Scope boundary

In scope: the **tool-catalog gap** only (which tools a child's agents may call). Explicitly out
of scope: **information-flow** across the supervisory boundary (`SteerContext`, `${ctx.*}`
chaining) — that is #56 and stays a separate concern, per the #41 design's non-goal §4 and the
#59 review. The `docs/validation.md` DIP146 section carries a standing **"what DIP146 does NOT
check"** list (partial-audit depth>1, info-flow #56, runtime enforcement) so a green result is
never mistaken for a guarantee. *(Security M2/M5)*

## Detection vs enforcement (tracker note)

dippin's role is **detection** — DIP146 *detects* the cross-file gap at author time. Actual
runtime enforcement of tool restrictions is the downstream tracker runtime's responsibility.
Per the `never-gate-dippin-on-tracker` rule, we ship cross-file *detection* now and do **not**
gate it on any tracker change. (The issue title says "enforcement"; the deliverable is
detection.) A clean DIP146 result means *"every delegated child this lint could resolve is
either fully locked down or had no restricting intent to escape"* — **not** that the child's
tools are restricted at runtime; the docs say so explicitly. *(Security C2)*

## Path-safety posture (v1: termination-only)

The pass resolves refs lexically (relative to `node.Source.File`, mirroring DIP135's
`resolveManagerLoopRef` / `resolveRefPath`) — inheriting dippin's existing "resolve whatever
the author wrote, no sandbox" lint posture. v1 adds **only** the termination hardening above
(EvalSymlinks-keyed visited set + depth cap), which fixes real infinite-recursion/stack-overflow
bugs. **Symlink-read refusal and root-escape** (reusing `dipx`'s hardened
`readNoFollowSymlinks` + `ErrRefEscape`, currently unexported, hardened in #79/#85) are a
**documented follow-up**, not v1: a local linter's only output is DIP codes (no content
exfiltration), and non-`.dip` targets fail to parse, so the marginal risk is low while the
code/coupling cost (defining a "root" for loose-file lint, factoring an unexported `dipx`
primitive) is real. This is an explicit, recorded scope choice. *(Security I3, Correctness M2;
panel preferred full hardening — deferred deliberately.)*

`SubgraphConfig.Ref` is documented "Workflow name **or** path"; a name-form ref that doesn't
resolve to a file is treated as `unresolved` (DIP143 retained) — acknowledged, not specially
handled. *(Security I4)*

## WASM split

No build-tag gymnastics needed. The traversal lives entirely in `cmd/dippin`, which is never
compiled to wasm (`just wasm` builds only `./cmd/wasm/`; `validator` compiles transitively as a
dependency). The `validator` additions (DIP146 const + explanation, and the exported intent
predicate — all pure IR/data, no `os` imports) compile for wasm trivially. Verification gate:
`just wasm` must stay green. *(Arch M1)*

## Architecture summary

```
validator/lint_subgraph_tool_access.go
  └─ EXPORT WorkflowDeclaresToolAccess(*ir.Workflow) bool  (+ a node/branch predicate)
       so DIP146's gate + child-classification reuse DIP143's EXACT intent logic
       (wasm-safe; prevents the two checks from drifting — the bug class this arc closes)

cmd/dippin/crossfile_tool_access.go   (NEW, native-only — package main)
  ├─ skip if entry is .dipx
  ├─ DFS the SAME *ir.Workflow CmdLint parsed (no re-parse); intentSeen threaded down
  ├─ per boundary EDGE: classify child → DIP146 (zero-intent) at boundary node
  ├─ visited set (EvalSymlinks key, pre-seeded, pre-order) gates RECURSION only
  ├─ depth cap; skip Source.File=="" boundaries
  └─ returns ([]validator.Diagnostic, map[ir.SourceLocation]childPosture)

cmd/dippin/cmd_validate.go : CmdLint
  └─ drop DIP143 where classified ∈ {zero-intent, full-restrict, agent-less}; append DIP146

validator/lint_codes.go      : + DIP146 const (+ "CLI-emitted" comment) + CodeDescription
validator/explanations.go    : + DIP146 explanation (4 fields) + group-comment range bump
```

**Helper decomposition (complexity budget cyclo ≤5 / cognitive ≤7, no `//nolint`).** `dipx`'s
own ref DFS needed a 3-way split (`detectCycles`/`dfsVisit`/`dfsVisitEdge`) to fit; mirror it.
The recursive walker is split into a **driver** (`walkBoundaries`: iterate nodes, gate
recursion on visited/depth) + a **per-edge worker** (`visitBoundary`: extract ref, resolve,
parse, classify child, emit, recurse). Other helpers: `resolveBoundaryChild` (ref → path →
parse, fail-soft), `classifyChild` (reuses the exported predicate), `boundaryDiag`. Plan for
**6** helpers, not 5 — a single recursive walker will bust cognitive-7. *(Arch M2, Conventions
C1, Correctness M3)*

## Prior art not reused (and why)

`dipx/helpers.go:walkSourceTree` is *already* a transitive ref-walk over loose on-disk `.dip`
files (not just archives) with a `filepath.Abs` visited set and both ref kinds via `refFromNode`
— closer to reusable than "pack/.dipx-oriented" suggests. It is **not** reused because: (a) it
is **fail-hard** (missing/unparseable/escaping child aborts the whole walk via `ErrSubgraphParse`
/ `ErrRefEscape`), whereas DIP146 needs **fail-soft per-edge** (skip → retain DIP143); (b) it
enforces the bundle `workflows/`-root escape guard that doesn't apply to ad-hoc loose-file lint;
(c) it is unexported and lives in `dipx`, which MUST NOT import `validator`, so it can't return
`validator.Diagnostic`. The hand-rolled traversal reuses its *canonicalization pattern*, not its
code. *(Arch I1)*

## Testing (TDD, parser-driven, real fixtures)

Multi-workflow behavior **must** use real child `.dip` fixtures parsed by the real parser
(per CLAUDE.md — don't hand-populate IR the parser doesn't set; the DIP101 bug came from
hand-built IR). Fixtures live in **`cmd/dippin/testdata/crossfile/`** (entry + child files,
referenced by relative path so `Source.File` is populated). The pass is exposed as a plain
function returning `([]validator.Diagnostic, map[ir.SourceLocation]childPosture)`, unit-tested
like `runPack`. Each case is a failing test first:

1. **Fires (zero-intent):** entry restricts → child restricts nothing → one DIP146 at the
   boundary; DIP143 for that boundary suppressed. Use a **relative** entry path (guards the
   supersession-key normalization bug).
2. **Silent (full-restrict):** entry restricts → child restricts every agent → no DIP146;
   DIP143 suppressed.
3. **Retains DIP143 (partial-audit):** entry restricts → child restricts A, leaves
   tool-bearing B open → no DIP146; **DIP143 retained** (the security-critical case).
4. **Silent (no path intent):** no workflow on the path restricts → no DIP146 regardless of
   child.
5. **Ancestor-path gate:** entry has *no* intent → intermediate parent restricts → child
   zero-intent → DIP146 fires (entry-only gating would miss this).
6. **Transitive:** entry → child (restricts) → grandchild (zero-intent) → DIP146 on the
   child→grandchild boundary.
7. **Diamond (per-edge):** `P1, P2 → C` (zero-intent) → **two** DIP146s, one per boundary;
   `C` walked once.
8. **Cycle terminates:** `A → B → A` completes without hang/duplicate; **symlink cycle**
   variant terminates (EvalSymlinks key); deep acyclic chain hits the depth cap and stops.
9. **Self-reference:** `A → A` terminates, no spurious finding.
10. **Unresolvable child:** missing/unparseable `ref`, and `.dipx` entry → no DIP146, DIP143
    **retained**, no crash.
11. **Agent-less child:** child with zero agents → no DIP146 (no tools to grant).
12. **Intentional-open worker (known FP):** fully-tooled child → DIP146 **does** fire, with a
    test comment: *known intentional-open false positive, mitigated by Hint severity, not by
    the gate — do not "fix" into a Warning or add a heuristic.*
13. **Catalog parity:** DIP146 in `CodeDescription` + `Explanations` (parity tests).
14. **Integration:** `dippin lint` on a fixture entry prints DIP146 and not the redundant
    DIP143; surviving DIP143 (case 3/10) renders.

Like DIP143, DIP146's *safe* shape is inherently cross-file, so it has **no lint-clean example**
in `examples/` (mirrors #59). `just lint-examples` runs the CLI (now including the cross-file
pass) — DIP146 is a Hint and must not turn any example red; confirm no example unexpectedly
triggers it.

## Edit-site checklist (verified against the tree)

- **Atomic (build red until both present):** `validator/lint_codes.go` const +
  `CodeDescription[DIP146]`; `validator/explanations.go` `Explanations[DIP146]` (all 4 fields
  non-empty, in `safetyExplanations()` next to DIP143).
- **Group-comment bump (easy to miss):** `validator/explanations.go:413`
  (`(DIP138–DIP143)` → `…DIP146`) — #59's analogous bump; **this spec's earlier draft omitted
  it.** *(Conventions C2)*
- **Convention-only count strings (hand-edit, no test):** `CLAUDE.md:85` ("54 diagnostic
  codes" → 55; "DIP101-DIP145" → "DIP101-DIP146"); `validator/lint.go:8` +
  `validator/lint_codes.go:3` comment ranges (`…DIP145` → `…DIP146`); `docs/validation.md`
  **four** range strings (lines 6, 15, 227, 1070) **and** line 3's two counts ("registers 54"
  → 55, "documents 49" → 50); `docs/llm-reference.md:188` ("54…" → 55) + line 191 range.
- **New documented section:** append a **`### DIP146`** section to `docs/validation.md` (after
  the DIP145 block) — without it, "documents 50" is false (there are exactly 49 `### DIPxxx`
  sections today). *(Conventions C2)*
- **Spec source docs (then the pre-commit hook / `just spec-check` regenerates):** DIP146 prose
  is assembled into `docs/generated-spec.md` (gitignored) + `cmd/dippin/generated-spec.md`
  (tracked) by `scripts/gen-spec.sh` from `docs/llm-reference.md` (`## Grammar`→EOF) +
  `site/static/skill.md` (`## File Structure`→`## Documentation`). Edit the existing
  **"Subgraph boundary" paragraph at `site/static/skill.md:106`** (it already documents DIP143
  and references #89) — not a new section. The **pre-commit hook auto-regenerates** and
  `git add`s the tracked spec (`.git/hooks/pre-commit`); **never** hand-edit the generated
  files. *(Conventions I1 — note: #59's spec wrongly claimed gen-spec is not pre-commit.)*

## Verification gates

`just test-pkg validator` · `just test` (cmd/dippin cross-file tests) · `just complexity` ·
`just wasm` · `just fmt` · `just lint-examples` (Hint — must not turn any example red) ·
`just spec-check`. Pre-commit hook is the real CI gate (`just check` ends on a
tree-sitter-generate step that fails locally — see the `just-check-tree-sitter-gotcha` note).
