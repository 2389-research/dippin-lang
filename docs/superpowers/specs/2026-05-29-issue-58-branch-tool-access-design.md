# Issue #58 — `BranchConfig.ToolAccess` per-branch override

**Date:** 2026-05-29
**Issue:** #58 (deferred from #41; unblocked by #76)
**Related roadmap:** #53 (defaults cascade), #54 (middle tier), #59 (manager-loop propagation)

## Problem

Block-form `parallel` nodes carry per-branch overrides in `ir.BranchConfig`
(`Target`, `Model`, `Provider`, `Fidelity`). `tool_access` — the v0.32.0 safety
primitive that strips an LLM agent's tool catalog — exists only on `AgentConfig`.
There is no way to scope a single fan-out branch's tool access. #76 built the
per-branch DOT round-trip infrastructure; this issue adds `tool_access` to it.

## Goal

Add `ToolAccess` to `ir.BranchConfig`, plumb it through every path the existing
branch fields travel (parser → formatter → DOT export → DOT migrate), and validate
invalid values via the existing DIP139 check. dippin **carries and lints** the
field; tracker **enforces** the override at runtime, exactly as it does for
`Model`/`Provider`/`Fidelity`.

## Semantic model (normative)

A block-form branch is a per-fan-out override of its target agent. `ToolAccess`
joins that override set with one **safety-critical resolution rule**:

> **Empty branch `tool_access` INHERITS the target agent's `tool_access`. It never
> resets to the full catalog.**

Effective value: `branch.ToolAccess if non-empty else agentNode.ToolAccess`
(which itself may be `none` or empty→full). A branch may **set or narrow** tool
access by declaring it explicitly; an omitted branch value leaves the target's
restriction in force.

Why this matters: if an omitted branch value resolved to the *full* catalog, then a
locked-down agent (`tool_access: none`) fanned out through a branch that omits the
field would silently run with all tools — an invisible loosening of the exact
safety primitive `tool_access` exists to provide. Inherit-on-empty is the safe
default and the only one that composes with the future defaults cascade (#53).

This rule is a **tracker runtime-resolution requirement**. dippin cannot enforce it
(it emits IR/`.dipx`; tracker resolves). It is documented normatively here, in the
IR field doc comment, and in `skill.md` so the tracker-side implementation
resolves *inherit, not reset*, and ships with a red-team test (agent
`tool_access: none` + branch with empty `tool_access` → zero tools).

Recognized values mirror agents: `""` (inherit) and `none` (strip tools). Any other
value is fail-closed at runtime and flagged by DIP139 at lint time.

### Why no new lint beyond DIP139

In v1 the value set is `{"", "none"}`. A branch can only **set or narrow** — it
**cannot widen** a target's restriction, because there is no `full` spelling. So no
unsafe branch/target combination is expressible, and there is nothing actionable for
an advisory to flag. DIP139 (invalid value) is therefore the only per-branch
`tool_access` diagnostic. When #54 adds a middle tier (e.g. `read_only`), "branch
widens target's restriction" becomes expressible for the first time and is where an
ordering lint would belong — noted for #54, out of scope here.

## Touch sites (every path the field must travel)

| # | Site | Change |
|---|------|--------|
| 1 | `ir/ir.go` `BranchConfig` | add `ToolAccess string` + rationale doc comment |
| 2 | `parser/parse_nodes.go` `applyBranchField` | add `case "tool_access"` (else the value is silently discarded at parse and DIP139 never sees it) |
| 3 | `formatter/format.go` `writeBranch` **guard** (`if b.Model=="" && b.Provider=="" && b.Fidelity==""`) | add `&& b.ToolAccess==""` — else a branch setting *only* `tool_access` is silently dropped on format |
| 3b | `formatter/format.go` `writeBranchFields` | emit `tool_access:` when non-empty |
| 4 | `export/dot.go` `encodeBranch` | add `appendBranchField(parts, "tool_access", b.ToolAccess)` |
| 5 | `migrate/migrate.go` `branchFieldSetters` | add `"tool_access"` entry (one line — #76 made this map table-driven for exactly this) |
| 6 | `validator/lint_tool_access.go` `lintToolAccessValues` | generalize to also scan `ParallelConfig.Branches[]` (see Validation) |
| 7 | `validator/explanations.go` DIP139 + `lint_tool_access.go` help text | broaden "agent node" wording to include branches |

**No change needed (verified by review):**
- `migrate/parity.go` — `compareParallelBranches` compares the whole struct with
  `!=`, so the new field is auto-covered. (Keep a coverage *test* as a regression guard.)
- `dipx/`, `cmd/dippin/cmd_pack.go` — `.dipx` bundles raw `.dip` source bytes;
  tracker re-parses. No per-field branch serialization exists. Not a path.
- `simulate/parallel.go` — reads only `b.Target`. Unaffected.

## Validation (DIP139 generalization)

Mirror the established DIP114 branch pattern in `validator/lint_style.go`
(`checkNodeFidelityByKind` → `checkBranchFidelities` → branch-qualified message).
Decompose so each function stays within cyclo ≤5 / cognit ≤7 (the current
`lintToolAccessValues` is cyclo 4 / cognit 5 — adding the branch loop inline would
breach the caps):

- Extract the value check into a shared helper over the existing `validToolAccess`
  map (do **not** duplicate the map — #54 then changes one place).
- Add a `ParallelConfig` branch over `n.Config` that loops `cfg.Branches` and checks
  each `b.ToolAccess`.
- Branch-qualified message (branches have no ID): `node %q branch %q has tool_access
  %q which is not recognized`, using `b.Target`. Location points at the parallel
  node's `Source` (BranchConfig carries no SourceLocation).
- Severity stays `Warning` (consistent with the agent rule and the pack-time split —
  only DIP001–009 block pack).

DIP140 is **not** extended: `BranchConfig` has no `Params` field, so there is no
params-bypass to detect. Leave a one-line tripwire comment near
`lintParamsReenablesTools` ("branches have no Params today; if added, extend this
scan") so a future `BranchConfig.Params` reopens the check deliberately.

## Testing (TDD; ~5 tests, no new reserved-char test)

#76 already proved the encoder handles arbitrary values, and `tool_access` values
(`none`) use the identical `appendBranchField`/`branchFieldSetters` path — a fresh
reserved-char test would be redundant.

1. **Parser** — `tool_access: none` inside a `branch:` block populates
   `BranchConfig.ToolAccess`; empty/absent leaves it `""`.
2. **Validator (DIP139)** — fires on an invalid branch value with a branch-qualified
   message; silent on `none`/empty. (positive + negative)
3. **Round-trip** — extend `migrate/roundtrip_test.go`'s `TestRoundtripBlockFormParallel`
   so one branch carries `tool_access: none` through `.dip → export → Migrate`.
4. **Formatter** — a branch with only `tool_access` set re-emits it (guards against the
   `writeBranch` early-return regression).
5. **Parity coverage** — a branch-only `tool_access` difference is detected by
   `compareParallelConfigs` (regression guard; production code already covers it via `!=`).

## Docs

- `docs/nodes.md` block-form section — add `tool_access` to the per-branch field list
  and the DOT-mapping token list; one sentence on the inherit-on-empty rule.
- `site/static/skill.md` — one sentence that `tool_access` is settable per-branch,
  pointing to the agent-level section for semantics + the inherit rule. Do not
  duplicate the threat-model prose.
- **Do not touch** `CHANGELOG.md` (tag-time) or hand-edit `docs/generated-spec.md`
  (pre-commit regenerates it).

## Out of scope

- **No new `examples/*.dip`** — proven by tests, per #76's precedent;
  `examples/agent_tool_access.dip` already documents the agent-level primitive.
- **No cascade/resolution logic** — effective-value resolution across
  defaults→agent→branch is #53/tracker's job. dippin carries the raw string only.
- **Branch `model`/`provider` validation** (DIP108 is agent-only) and **branch diffing**
  (`diff/diff.go` has no ParallelConfig case) are pre-existing gaps, not part of #58.
