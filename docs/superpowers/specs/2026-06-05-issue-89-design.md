# Issue #89 — DIP146: cross-file effective `tool_access` detection across the subgraph boundary

**Date:** 2026-06-05
**Closes:** [#89](https://github.com/2389-research/dippin-lang/issues/89)
**Deferred from:** [#59](https://github.com/2389-research/dippin-lang/issues/59) (DIP143, the single-file advisory)
**Arc:** #41 `tool_access` → #75 `writable_paths` → #79 TOCTOU resolver → #59 subgraph boundary advisory (DIP143) → **#89 cross-file resolution (DIP146)**

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
lacks the restriction the parent's posture implies** — eliminating DIP143's central
false-positive class (children that *were* already audited).

## Decisive constraints (the architectural crux)

1. **`validator` cannot do this.** It may not import `parser` and is compiled to wasm
   (`GOOS=js GOARCH=wasm`), where there is no filesystem to read the child from. DIP143 lives
   in `validator` precisely because a per-file, wasm-safe lint is the most it can say.

2. **Only the loader tier / CLI may cross files.** Per CLAUDE.md, only `dipx` may compose
   `ir + parser + simulate`, and **pack-time structural validation is invoked at the CLI
   layer, not inside `dipx`** (`cmd/dippin/cmd_pack.go:validateEntryPrePack`). `dipx` itself
   MUST NOT import `validator`. The cross-file traversal therefore lives in the **CLI layer
   (`cmd/dippin`)**, which is the one place allowed to compose `parser` + `validator`, and
   which is never built for wasm.

### Decision: DIP146, a registered code emitted from the CLI

The finding surfaces as a real DIP code, **DIP146**, so it completes DIP143's story in the
same catalog (`dippin explain DIP146`, generated spec, docs) and stays consistent with all 54
existing codes. The twist:

- The **const + explanation live in `validator`** (`lint_codes.go` + `explanations.go`) — pure
  data, wasm-safe, no new imports.
- The code is **emitted from a native CLI pass** in `cmd/dippin`, never from
  `validator.Lint()`.

This is safe: `TestExplanationsCoverAllCodes` / `TestExplanationsNoExtra` only require
`CodeDescription` ↔ `Explanations` parity. **No test asserts a registered code is reachable
from `Lint()`**, so a catalog-registered, CLI-emitted code passes the harness. (Verified
empirically against `validator/explanations_test.go`.)

*Rejected alternative — a bare CLI finding with no DIP code:* avoids touching docs/count
strings, but is inconsistent with the diagnostic model (every other finding is a DIPxxx), is
not discoverable via `dippin explain`, and does not visibly close the DIP143 → #89 loop.

## Design — DIP146, severity `Hint`

### The containment rule (conservative, low-noise)

`tool_access` is **binary**: `""` = full tool catalog (the default), any non-empty value =
restricted (the runtime fails closed, so even a typo yields no-tools). Branch `""` = *inherit
the target agent's value*; the only recognized branch values are `""` and `none`, so a branch
can never grant **more** than its target agent — branch-level analysis is subsumed by the
agent check.

Because full access is the **default**, "any unrestricted child agent" would over-fire on
intentionally-tooled worker agents — the exact alert-fatigue failure mode #59 designed
DIP143's Hint severity to avoid. So DIP146 fires conservatively:

> **Fire DIP146 for a boundary edge (parent node → child file) when, and only when:**
> 1. the **entry workflow** (the file being linted) shows containment intent — at least one
>    `AgentConfig` or `BranchConfig` with non-empty `tool_access` (mirrors DIP143's
>    `workflowDeclaresToolAccess`); **and**
> 2. the boundary's resolved child workflow is **successfully parsed**; **and**
> 3. that child declares **zero** containment intent — **no** agent or branch anywhere in the
>    child restricts `tool_access` at all.

This is a strict upgrade over DIP143: by reading the child, DIP146 stays **silent** when the
child already shows tool-access awareness (any restricted agent), eliminating DIP143's main
false-positive class. The signal is *"this delegated child was never audited for tools,"* not
*"this specific agent has tools"* (often intentional).

**Why entry-gated (not per-edge):** the gate keys on the entry's intent — the file the author
is linting and reasoning about — matching DIP143's single-workflow gate. A child that itself
restricts tools but delegates to a grandchild with none is still flagged transitively
(non-propagation is transitive); we do not additionally require the *intermediate* parent to
show intent. This is the simplest defensible v1; per-edge intent gating is a possible later
refinement and is recorded here, not built.

### Transitivity & cycles

- **Full transitive traversal.** From the entry, DFS every `manager_loop`/`subgraph` boundary,
  parsing each child and recursing. Non-propagation is transitive, so a one-hop check (DIP143)
  is insufficient — this is the headline upgrade.
- **Visited-set guard (mandatory).** Keyed by canonical absolute child path
  (`filepath.Abs`+`filepath.Clean`). Terminates on self-reference (`A → A`) and cycles
  (`A → B → A`) — a child already visited is not re-traversed.
- **No separate cycle diagnostic.** Cycle *reporting* is out of scope for #89 (whose
  deliverable is tool-access detection); the visited set only needs to guarantee termination.
  `dipx` already rejects ref cycles at pack time (`ErrRefCycle`). Recorded as a possible
  follow-up.

### Superseding DIP143 in the native `lint` path

In `dippin lint`, validator.Lint() emits DIP143 per-file (it can't see across the boundary)
*and* the new cross-file pass runs. Double-reporting the same boundary is a wart. Resolution:

- The cross-file pass returns its DIP146 findings **plus the set of boundary locations it
  successfully resolved** (`map[ir.SourceLocation]bool`, keyed on the boundary node's
  `Source`).
- `CmdLint` **drops any DIP143 whose location was resolved** by the cross-file pass, then
  appends the DIP146 findings.
- Result, native `lint`: DIP146 where there's a confirmed gap; **silence** where the child was
  audited; DIP143 survives **only** for boundaries the pass could not resolve (missing or
  unparseable child — `os.Stat`/parse failed). In wasm / the playground, DIP143 stands alone,
  exactly as today. `validator` still owns DIP143 entirely (unit tests unaffected); only
  `CmdLint`'s composition changes.

### Severity: `Hint` (not Warning)

A child may be **intentionally** fully open (a trusted tool-running subgraph). The repo has
**no per-diagnostic suppression mechanism**, so a Warning that fires on such a correct file
would be corrosive — the memory note and #59's DX findings both mandate `Hint` for
"context-dependent, may be fine" advisories (precedent: DIP125/131/133, and DIP143 itself).
DIP146's value over DIP143 is **precision** (it fires far less often, and only on a confirmed
gap), not louder severity.

### Where it runs

`dippin lint` only — DIP143's home, where authors already see the boundary advisory.
`validate` stays structural-only (DIP001–DIP009) per its contract. `pack` already rejects ref
cycles and could host DIP146 later, but its pre-pack gate is structural-errors-only today;
adding a Hint-level cross-file pass there is out of scope for v1.

### Message & help (reframed, mirroring DIP143's discipline)

The message must **not** claim the child "inherits" anything, and must **not** imply that a
child setting `tool_access: none` is therefore fully safe — the supervisory/steering channel
(`SteerContext`, `stack.child.*`) is information-flow, a **separate** concern (#56). Draft:

> **DIP146** (Hint): `manager_loop "Supervise" delegates to subgraph "child.dip", which
> restricts no agent's tool_access; this workflow restricts tools, but the restriction does
> not cross the subgraph boundary.`
> **help:** `Give child.dip's agents their own tool_access (e.g. tool_access: none on
> summarizers). This bounds the child's tool catalog, not information flow across the
> supervisory boundary (see #56).`

Located at the **boundary node** in the referencing file (`node.Source`), so each hop's
finding points at an editable site in a real file.

## Scope boundary

In scope: the **tool-catalog gap** only (which tools a child's agents may call). Explicitly out
of scope: **information-flow** across the supervisory boundary (`SteerContext`, `${ctx.*}`
chaining) — that is #56 and stays a separate concern, per the #41 design's non-goal §4 and the
#59 review. DIP146 keeps the two concerns separate.

## Detection vs enforcement (tracker note)

dippin's role is **detection** — DIP146 *detects* the cross-file gap at author time. Actual
runtime enforcement of tool restrictions (and whether restrictions should cross the subgraph
boundary at runtime) is the downstream tracker runtime's responsibility. Per the
`never-gate-dippin-on-tracker` rule, we ship cross-file *detection* now and do **not** gate it
on any tracker change. If runtime cross-file enforcement is later desired, that is a separate
tracker follow-up, not a blocker here. (The issue title says "enforcement"; the deliverable is
detection.)

## WASM split

No build-tag gymnastics needed. The traversal lives entirely in `cmd/dippin`, which is never
compiled to wasm (wasm builds only `cmd/wasm` + `validator`). The `validator` additions
(DIP146 const + explanation) are pure data — no `os`/filesystem imports — so they compile for
wasm trivially. Verification gate: `just wasm` (`GOOS=js GOARCH=wasm go build ./cmd/wasm/
./validator/`) must stay green.

## Architecture summary

```
cmd/dippin/crossfile_tool_access.go   (NEW, native-only — package main)
  ├─ parses entry + children via the existing loadWorkflow (parser)
  ├─ entry-intent gate (reuses the DIP143 intent predicate's logic)
  ├─ DFS boundaries with a canonical-path visited set
  ├─ per child: "declares zero tool_access intent?" → DIP146 at boundary node
  └─ returns ([]validator.Diagnostic, resolved map[ir.SourceLocation]bool)

cmd/dippin/cmd_validate.go : CmdLint
  └─ drop DIP143 where resolved[loc]; append DIP146 findings

validator/lint_codes.go      : + DIP146 const + CodeDescription entry
validator/explanations.go    : + DIP146 explanation (Summary/Trigger/Fix/Example)
```

Complexity budget (cyclomatic ≤5 / cognitive ≤7) is met by extracting helpers:
`entryDeclaresToolAccess`, `resolveBoundaryChild`, `childDeclaresToolAccess`,
`walkBoundaries(visited, ...)`, `boundaryDiag`.

## Testing (TDD, parser-driven, real fixtures)

Multi-workflow behavior **must** use real child `.dip` fixtures parsed by the real parser
(per CLAUDE.md — don't hand-populate IR the parser doesn't set; the DIP101 bug came from
hand-built IR). Fixtures live under a test dir; the test resolves them by path like the CLI
does. Cases (each a failing test first):

1. **Fires:** entry restricts tools → child with zero `tool_access` → exactly one DIP146 at
   the boundary node; DIP143 for that boundary is suppressed.
2. **Silent (audited child):** entry restricts → child restricts ≥1 agent → no DIP146; DIP143
   suppressed (boundary resolved).
3. **Silent (no entry intent):** entry sets no `tool_access` → no DIP146 regardless of child.
4. **Transitive:** entry → child (restricts) → grandchild (zero intent) → DIP146 on the
   child→grandchild boundary.
5. **Cycle terminates:** `A → B → A` completes without hang or duplicate; visited-set proven.
6. **Self-reference:** `A → A` terminates, no spurious finding.
7. **Unresolvable child:** missing/unparseable `ref` → no DIP146, DIP143 **retained** (not
   suppressed), no crash.
8. **Catalog parity:** DIP146 in `CodeDescription` + `Explanations` (existing parity tests go
   green only when both are added).
9. **Integration:** `dippin lint` on a fixture entry prints DIP146 and not the redundant
   DIP143.

## Edit-site checklist (the test-gated + convention-only sites)

- **Atomic (build red until both present):** `validator/lint_codes.go` const +
  `CodeDescription[DIP146]`; `validator/explanations.go` `Explanations[DIP146]` (all 4 fields
  non-empty).
- **Convention-only count strings (hand-edit, no test):** `CLAUDE.md` ("54 diagnostic codes",
  "DIP101-DIP145" range), `validator/lint.go` + `lint_codes.go` comment ranges,
  `docs/validation.md` (registered total **and** "documents N" — two distinct counts),
  `docs/llm-reference.md`. 54 → 55 registered; "documents 49" → "documents 50".
- **Spec source docs (then `just spec-check`):** DIP146's prose is assembled into
  `docs/generated-spec.md` (gitignored) + `cmd/dippin/generated-spec.md` (tracked) by
  `scripts/gen-spec.sh` **from `docs/llm-reference.md` + `site/static/skill.md`** — add a
  DIP146 section to those source docs; **never** hand-edit the generated files.

## Verification gates

`just test-pkg validator` · `just test` (cmd/dippin cross-file tests) · `just complexity` ·
`just wasm` · `just fmt` · `just lint-examples` (DIP146 is a Hint — must not turn any example
red) · `just spec-check`. Pre-commit hook is the real CI gate (`just check` ends on a
tree-sitter-generate step that fails locally — see the `just-check-tree-sitter-gotcha` note).
