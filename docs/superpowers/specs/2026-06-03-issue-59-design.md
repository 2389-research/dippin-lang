# Issue #59 — DIP143: subgraph `tool_access` containment-boundary advisory

**Date:** 2026-06-03
**Closes:** [#59](https://github.com/2389-research/dippin-lang/issues/59)
**Deferred follow-up filed:** [#89](https://github.com/2389-research/dippin-lang/issues/89) (real cross-file enforcement)
**Status:** Design approved 2026-06-03 (4-expert review panel + 3 design forks decided by author)
**Arc:** #41 `tool_access` → #75 `writable_paths` → #79 TOCTOU resolver → **#59 subgraph boundary**

## Problem

A `manager_loop` node supervises a child pipeline referenced by `subgraph_ref`, and a
plain `subgraph` node embeds one via `ref`. Both point at a **separate `.dip` file**.
`tool_access` is a per-node primitive (`AgentConfig.ToolAccess`, `BranchConfig.ToolAccess`)
with no workflow-level policy. So an author who locks agents down (`tool_access: none`) in
the parent file gets **no** guarantee about the referenced child — the child's agents are
governed entirely by their own file. The restriction silently stops at the file boundary.

Per the #41 design's Non-goals §7: *"Parent's restrictive `tool_access:` does not propagate
into a child subgraph. Requires multi-workflow IR traversal the validator doesn't have today."*

## Decisive constraint

The `validator` package **must not import `parser`** (CLAUDE.md layering rule — only the
`dipx` loader tier may compose `ir + parser + simulate`). Confirmed empirically: no
production file in `validator/` imports `parser`. Therefore the validator **cannot parse
the child `.dip`** to inspect its actual `tool_access`. Real cross-file effective-access
comparison must live in `dipx`/CLI and is deferred to #89.

This is a single-workflow **advisory**, not enforcement and not propagation.

## Design — DIP143, severity `Hint`

### What it detects (intent-gated)

Fire DIP143 on a node that **references an external subgraph** —
`ManagerLoopConfig.SubgraphRef` **or** `SubgraphConfig.Ref`, non-empty — **if and only if**
the same workflow contains at least one node declaring **containment intent**: an
`AgentConfig` or any `BranchConfig` whose `tool_access` is **non-empty** (after trim).

- **Why "any non-empty" rather than `== "none"`:** the runtime fails closed — *any* non-empty
  `tool_access` value disables tools (a typo `nono` still yields no-tools per the #41 design).
  An author who typed any value intended and received restriction, so any non-empty value is
  containment intent (security review finding I5). This differs from DIP141's dead-config
  check, which correctly matches `== "none"` only.
- **Intent gate, not boundary marker:** this is the literal reading of #59 ("parent's
  *restrictive* `tool_access` does not propagate"). When the author restricted nothing, there
  is no restriction to escape, so DIP143 stays silent. The security panel argued for an
  unconditional boundary marker (fire on every subgraph ref); the author chose the intent gate
  to keep the hint targeted and the example suite clean. The broader boundary-marker idea and
  the no-lockdown population are recorded in #89.

### Severity: `Hint` (not Warning)

The referencing node has **no defect** — the concern lives in a file the linter can't see.
The repo has **no per-diagnostic suppression mechanism**, so a Warning that fires on a correct
file with no resolving edit would train authors to ignore the whole DIP139–142 safety band.
`SeverityHint` matches the established "context-dependent, may be fine" precedent (DIP125,
DIP131, DIP133). (DX review findings C1/C2.)

### Scope: both `manager_loop` and `subgraph`

Both node types have the identical cross-file gap; covering only `manager_loop` would leave a
same-shaped hole (security review M8). One hint per referencing node (each boundary is its own
advisory).

### Message & help (reframed)

The message must **not** say the child "inherits" restrictions and must **not** imply a child
declaring `tool_access: none` is fully safe — the supervisory/steering channel
(`SteerContext`, `stack.child.*`) is information-flow, a separate concern (#56). It leads with
the boundary and points the author at the actionable check (security I4, DX M1).

```text
Message: <kind> %q references subgraph %q, defined in its own file; this workflow's
         tool_access restrictions do not extend across the subgraph boundary
Help:    audit the agents in %q for their own tool_access — restrictions declared in this
         workflow do not propagate into a referenced subgraph. Cross-file enforcement is
         tracked as #89.
```

`<kind>` is `manager_loop` or `subgraph`; `%q` interpolates the node ID and `cfg.SubgraphRef`
/ `cfg.Ref` verbatim (so message and help reference the same string).

### Suppresses when

- No node in the workflow declares `tool_access` (no containment intent), **or**
- The referencing node's ref is empty (owned by DIP135 / DIP126), **or**
- The ref resolves to the node's own source file — a **direct self-reference** has no
  cross-file boundary, so the hint would be wrong (post-review, Copilot). Resolution is
  filepath-only (`refIsSelf`), mirroring `resolveRefPath`. Transitive cross-file cycles
  (A → B → A) still require multi-file traversal and remain deferred to #89.

## Implementation

New file `validator/lint_subgraph_tool_access.go` — pure IR logic with no `os`/filesystem
access, so **no `//go:build` split** (unlike `lint_manager_loop.go`, which is split only
because it `os.Stat`s the child path). `path/filepath` path-math (`refIsSelf`) is wasm-safe;
only `os` syscalls are not. Decomposed to stay within cyclomatic ≤5 / cognitive ≤7
(security/arch review: a single type-switch-with-branch-loop would blow the cognitive gate):

```go
lintSubgraphToolAccess(w)      → if !workflowDeclaresToolAccess(w) { return nil }; loop → checkSubgraphBoundary(n)
workflowDeclaresToolAccess(w)  → loop → nodeDeclaresToolAccess(n)
nodeDeclaresToolAccess(n)      → type switch: AgentConfig → toolAccessSet(cfg.ToolAccess)
                                              ParallelConfig → branchesDeclareToolAccess(cfg.Branches)
branchesDeclareToolAccess(bs)  → loop → toolAccessSet(b.ToolAccess)
checkSubgraphBoundary(n)       → kind, ref := subgraphRefOf(n); if ref == "" { return nil }; build *Diagnostic
subgraphRefOf(n)               → type switch: ManagerLoopConfig → ("manager_loop", cfg.SubgraphRef)
                                              SubgraphConfig    → ("subgraph", cfg.Ref)
toolAccessSet(s)               → strings.TrimSpace(s) != ""
```

Registered once in `validator/lint.go` `Lint()` after `lintWritablePaths(w)`.

## Edit-site checklist (derived from the #41/#75 precedent + review)

Build-breaking / test-gated:
1. `validator/lint_codes.go` — `DIP143` const + `CodeDescription[DIP143]`; update header comment range.
2. `validator/explanations.go` — `Explanations[DIP143]` in `safetyExplanations()`, all four
   fields (Summary/Trigger/Fix/Example) non-empty (gated by `TestExplanationsCoverAllCodes`/
   `TestExplanationsNoExtra`); update its doc comment `DIP138–DIP142` → `DIP138–DIP143`.
3. `validator/lint_subgraph_tool_access.go` — new file (pure, no build tag).
4. `validator/lint.go` — register; update header comment range.
5. `validator/lint_subgraph_tool_access_test.go` — parser-driven tests (see matrix).

Convention docs (no test gate, required by precedent):
6. `CLAUDE.md:85` — `51` → `52`, `DIP101-DIP142` → `DIP101-DIP143`.
7. `validator/lint_codes.go:3` + `validator/lint.go:8` — range comments.
8. `docs/validation.md` — range strings (lines 6, 15, 227, 972) + new `### DIP143` section.
9. `docs/llm-reference.md` — `51 diagnostic codes` (188) + range/description (191); **source**
   for `generated-spec.md`.
10. `site/static/skill.md` — add DIP143 to the safety section / revisit the cross-node note
    (judgment call; also a generated-spec source).
11. Regenerate: `just spec-check` refreshes `docs/generated-spec.md` + `cmd/dippin/generated-spec.md`
    (these are **assembled from** llm-reference.md + skill.md by `scripts/gen-spec.sh`, run by
    `just build`/`just check` — **not** pre-commit; never hand-edit the generated files).

Release-time (defer to tag):
12. `CHANGELOG.md` — DIP143 bullet under the next version heading.

## Tests — parser-driven (`validator/lint_subgraph_tool_access_test.go`)

Parse real `.dip` text via `parser.NewParser(src, "test").Parse()` — no hand-built IR for
fields the parser populates (the DIP101-bug lesson). Matrix:

| Case | Expect |
|---|---|
| agent `tool_access: none` + `manager_loop subgraph_ref:` | 1× DIP143 (Hint) on manager_loop |
| agent `tool_access: none` + `subgraph ref:` | 1× DIP143 on subgraph node |
| parallel branch `tool_access: none` + manager_loop | 1× DIP143 |
| agent `tool_access: nono` (any non-empty) + manager_loop | 1× DIP143 (intent = non-empty) |
| manager_loop ref, **no** `tool_access` anywhere | 0 |
| agent `tool_access: none`, **no** subgraph/manager_loop | 0 |
| `subgraph`/`manager_loop` with **empty** ref + restricted agent | 0 (DIP135/126 own empties) |
| emitted diagnostic severity | `SeverityHint` |

## No example `.dip`

DIP139/141 shipped lint-clean examples of the *safe* shape. DIP143's "safe shape" is inherently
cross-file (the child declares its own restrictions in a file we don't read), so it **cannot**
be demonstrated lint-clean in a single file. Coverage is parser-driven unit tests. (Existing
`manager_loop_demo.dip` / `api_design.dip` declare no `tool_access`, so the intent gate keeps
them silent — no example regresses.)
