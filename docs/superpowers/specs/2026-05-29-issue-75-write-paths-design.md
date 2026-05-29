# Issue #75 — `write_paths:` path-bounded write-scope primitive

**Date:** 2026-05-29
**Issue:** [#75](https://github.com/2389-research/dippin-lang/issues/75) (write-scope tier, deferred from #41 follow-ups)
**Related roadmap:** #55 (tool-name allowlists), #53 (defaults cascade), #54 (read-only tier), #58 (per-branch override), #56 (chain-attack mitigation)
**Status:** Design — under review
**Tracker dependency:** Joint release; pinned tracker SHA recorded in the implementation plan before dippin merges (mirrors #41).

## Problem

`tool_access: none` (v0.32.0, #41) is binary: an agent has either no LLM tools or
the full file-mutation catalog (Read, Write, Edit, ApplyPatch, Glob, GrepSearch,
Bash). Five real verdict/failure-recording agent sites in
[`2389-research/pipelines`](https://github.com/2389-research/pipelines) need to
**write** a small, prompt-known set of files — and only those — but sit at the wrong
end of the binary:

- With `tool_access: none`: their legitimate single-file write breaks at runtime.
- Without `tool_access:`: they keep the full mutation catalog, which empirically
  leaks via the scope-creep mode the v0.28.2 incident demonstrated.

This is distinct from the sibling follow-ups: #54 (`read_only`) does not help sites
that must *write*; #55 (`allowed_tools`) scopes tool *names*, not *paths* — granting
`Write` still lets the agent write anywhere.

### Real incident — prompt-only constraints failed (PR#25)

`2389-research/pipelines` [PR #25](https://github.com/2389-research/pipelines/pull/25)
bug 4: `redecompose_single` is an agent whose prompt restricts it to writing two YAML
files. In one production run (Rust target, `reasoning_effort: high`) it instead wrote
arbitrary Rust source, ran `cargo build` / `cargo test`, and fabricated ledger state.
The prompt-based path restriction did not hold — the agent had `Write`/`Edit`/`Bash`
and used them. A structural write-scope primitive enforced at the API boundary would
have prevented the incident; prompt discipline did not.

### Motivating sites (acceptance fixtures)

From PR #23's Phase-2 reclassification table (the five sites tagged **D**/**C+D**):

| Site | File:line | Writes |
|---|---|---|
| `L6Failed` | `greenfield/greenfield_review.dip:195` | `workspace/.review-failed` |
| `L7Failed` | `greenfield/greenfield_review.dip:328` | `workspace/.review-failed` |
| `Gate1Failed` | `greenfield/greenfield_synthesis.dip:455` | `workspace/.synthesis-failed` (+ reads `gate1-findings.md`) |
| `Gate2Failed` | `greenfield/greenfield_validation.dip:202` | `workspace/.validation-failed` |
| `RecoveryManager` | `sprint/sprint_exec_yaml_v2.dip:741` | `.ai/sprints/SPRINT-<id>-recovery-analysis.md` + `.ai/sprints/SPRINT-<id>-redecompose-request.yaml` + `.ai/managers/recovery-journal.md` |

`RecoveryManager`'s `.ai/` paths are why the simplest "named workspace tier" shape is
insufficient — the primitive must express arbitrary author-chosen globs.

## Design

One new field. A glob list. Same "dippin carries + lints; tracker enforces" contract
as #41/#58.

```dip
agent RecoveryManager
  prompt: "Record the recovery analysis and journal entry."
  write_paths: .ai/sprints/**, .ai/managers/recovery-journal.md
```

`write_paths` is a comma-separated glob list on `AgentConfig` and `BranchConfig`.

- **Absent** → unbounded writes (current behavior).
- **Non-empty** → tracker bounds all file mutations to the listed globs.

### Shape rationale (why a separate field, not overloaded `tool_access`)

Three shapes were considered (issue #75 sketches them):

1. **Named tier** (`tool_access: workspace_only`, tracker-defined glob). Simplest, no
   new field — but a fixed `workspace/**` cannot express `RecoveryManager`'s `.ai/`
   paths. **Rejected** (fails acceptance).
2. **Overloaded `tool_access`** (`tool_access: { write_paths: [...] }`). This is the
   structured-`tool_access` design the #41 spec explicitly rejected after three review
   rounds, and a breaking type change to a field tracker already consumes as a string.
   **Rejected** (reopens solved complexity).
3. **Separate `write_paths` list field**, orthogonal to the `tool_access` scalar.
   Rides the proven `reads`/`writes`/`outputs` list rails and the `tool_access` carry
   pattern; composition handled by lints, not field entanglement. **Chosen.**

`tool_access` stays a dumb scalar. `write_paths` is a dumb glob list. Neither the
parser nor the validator resolves, coerces, or canonicalizes — tracker owns runtime
semantics, exactly as #41 established.

### Semantic model (normative)

- **Effective resolution (tracker-side):** `write_paths` bounds the set of paths any
  file-mutating tool may write. It is a runtime-resolution concern; dippin carries the
  raw glob list only.
- **Branch inherit-on-empty** (mirrors #58's `ToolAccess` rule): an empty branch
  `write_paths` **inherits the target agent's `write_paths`**; it never resets to
  unbounded. Effective = `branch.WritePaths if non-empty else agentNode.WritePaths`.
  A branch may set or narrow scope by declaring globs; an omitted branch value leaves
  the target's bound in force. If an omitted branch resolved to *unbounded*, a
  path-locked agent fanned out through a branch that omits the field would silently
  run unbounded — an invisible loosening of the exact primitive `write_paths` exists
  to provide. Inherit-on-empty is the safe default and the only one that composes with
  the future defaults cascade (#53).
- **Empty-but-present** (`write_paths:` with no non-empty entries): the list-field
  pipeline (`len>0` formatter/export guards) collapses this to "absent" → unbounded.
  Rather than add machinery to preserve a present-vs-absent distinction the rest of the
  list-field surface does not preserve, this case is surfaced by **DIP142** at lint
  time. Authors who skip lint and ship an empty list get unbounded writes; the lint is
  the guard.

## Threat model & dippin/tracker split

**dippin CARRIES + LINTS; tracker ENFORCES.** dippin stores the raw glob list,
round-trips it (parse → IR → format → DOT export → migrate), and lints it. dippin does
not resolve, coerce, or canonicalize globs, and cannot enforce path bounding — it emits
IR/`.dipx`; tracker resolves and enforces.

**Enforcement contract (normative, tracker-side).** Writes are bounded to the declared
globs as a **filesystem-level jail that covers every tool the agent can invoke,
including `Bash`.** This is the structural property the PR#25 incident required: a
`reasoning_effort: high` agent that reaches for `bash("echo … > /outside/path")` or
`cargo build` is blocked at the fs boundary, not asked nicely in the prompt. Path
bounding at the tool-argument layer alone (validating `Write`/`Edit` arguments) is
**insufficient** — `Bash` is a write primitive and must be inside the same jail.

This contract is a tracker responsibility, co-designed and **joint-released** with this
dippin change (as #41 was). dippin allocates no lint code for enforcement; the runtime
is the source of truth.

**Per-backend requirement (mirrors #41).** The jail MUST hold on every backend dippin
supports (`native`, `claude-code`, `acp`). If a backend cannot enforce the fs jail when
`write_paths` is set, session-creation **refuses to start** with an error pointing to
this issue — authors get a hard failure, never a silent unbounded run. The
implementation plan records the exact per-backend enforcement mechanism with cited docs
(URL + retrieval date), mirroring CLAUDE.md's model-catalog discipline, and ships a
tracker-side test that exercises runtime behavior, not just documented behavior.

**Bypass defense (tracker-side).** When `write_paths` is set, tracker MUST NOT honor a
`Params` key that would widen the write surface (e.g. a working-dir or permission
override). Mirrors #41's `Params` bypass defense for `tool_access`.

**Chain caveat (per #56).** `write_paths` bounds a single agent's *immediate* writes,
not information flow. A `write_paths` agent can still launder data through an allowed
file into a downstream unbounded agent. Out of scope; tracked in #56.

## Dippin-side design

### IR (`ir/ir.go`)

`AgentConfig` gains `WritePaths []string`, clustered with `ToolAccess`:

```go
// WritePaths bounds the file paths this agent's tools may write, as author-chosen
// globs (e.g. "workspace/**", ".ai/sprints/**"). Empty = unbounded (full catalog can
// write anywhere). dippin carries + lints; tracker enforces an fs-level write jail
// covering every tool (Bash included). See issue #75.
WritePaths []string
```

`BranchConfig` gains `WritePaths []string` with the inherit-on-empty rationale doc
comment (mirroring the existing `ToolAccess` branch comment): empty INHERITS the target
agent's `write_paths`, never resets to unbounded.

### Parser (`parser/parse_nodes.go`)

- Agent runtime-field handler: `case "write_paths": cfg.WritePaths = splitComma(val)`
  (same helper `reads`/`writes`/`outputs` use).
- `applyBranchField`: `case "write_paths": b.WritePaths = splitComma(val)` (else a
  branch value is silently discarded and DIP142 never sees it).

Stores verbatim. No normalization, no glob validation — that is the validator's job.

### Formatter (`formatter/format.go`)

- Agent: emit in the agent runtime/IO cluster when `len(cfg.WritePaths) > 0`:
  `wr.line("write_paths: %s", strings.Join(cfg.WritePaths, ", "))`.
- Branch: `writeBranchFields` emits `write_paths` when non-empty, and the `writeBranch`
  early-return **guard** gains `&& len(b.WritePaths) == 0` — else a branch setting only
  `write_paths` is silently dropped on format.

### DOT export (`export/dot.go`)

- Agent: `attrs["write_paths"] = strings.Join(cfg.WritePaths, ",")` via
  `applyAgentRuntimeAttrs` when non-empty. Add `write_paths` to `reservedGraphAttrs` so
  it can't collide with author `vars:`.
- Branch: encode via `encodeBranch`'s field mechanism. **Risk to verify in planning:**
  the per-branch DOT encoding (#76's `appendBranchField`) is scalar-oriented; a
  comma-joined list value must not collide with the branch field separator. If it does,
  the branch encoder needs a list-aware path or an alternate join char for branch
  `write_paths`. A dedicated round-trip test covers this (see Tests).

### Migrate (`migrate/migrate.go`)

- Agent: `if v, ok := attrs["write_paths"]; ok { cfg.WritePaths = splitComma(v) }`.
- Branch: add `"write_paths"` to `branchFieldSetters` (the table #76 made for exactly
  this) using `splitComma`.
- Parity (`migrate/parity.go`): join-compare both agent and branch `WritePaths`
  (`strings.Join(a, ",") != strings.Join(b, ",")`), mirroring the `Outputs` comparator.
  Do not switch to `reflect.DeepEqual` — the comparator deliberately ignores other
  fields with separate round-trip behavior.

### Pack-time validation (`cmd/dippin/cmd_pack.go`)

DIP141/DIP142 are warning-severity at pack time (only DIP001–009 block `dippin pack`),
consistent with DIP139/DIP140.

## Lint rules (2 new codes)

### DIP141 — `write_paths` nullified by `tool_access: none`

Fires when a single `AgentConfig` or `BranchConfig` has non-empty `write_paths` **and**
`tool_access` normalizing to `none`. `none` strips the entire tool catalog, so there is
no `Write`/`Edit`/`Bash` left to bound — `write_paths` is dead config. Same dead-config
flavor as DIP140. Warning severity. Scans agent + branch (branch-qualified message via
`b.Target`, like DIP139).

> Note: only the *same config object* declaring both fields is flagged. A branch that
> sets `tool_access: none` while *inheriting* an agent's `write_paths` is a legitimate
> narrowing (the branch chose no-tools), not author error — not flagged.

### DIP142 — ineffective or unsafe `write_paths` value

Fires on:
- **Empty-but-present** list (→ silent unbounded; author intent was to restrict).
- **Absolute path** entry (`/etc/**`, `/tmp/x`) — escapes any workspace-relative jail.
- **Parent escape** — an entry whose normalized form escapes its base via `..`
  (`../../etc/**`). On-brand with the #67/#77 full-chain symlink-escape rejection.

Warning severity. Scans agent + branch. The empty-list and escaping-path sub-cases
share one code because both describe "this `write_paths` value will not bound writes the
way the author expects"; the message names the specific sub-case.

### Existing-code reuse

DIP140's `lintParamsReenablesTools` is not extended: `write_paths` re-enabling via
`Params` is a tracker-side bypass defense, not a dippin lint. Leave the existing
`BranchConfig`-has-no-Params tripwire comment intact.

## Tests (TDD)

1. **Parser** (`parser/parse_write_paths_test.go`, new): `write_paths: a, b` on an agent
   populates `WritePaths = ["a","b"]`; inside a `branch:` block populates
   `BranchConfig.WritePaths`; empty/absent leaves `nil`. Real parser, no hand-built IR.
2. **Validator DIP141** (`validator/lint_tool_access_test.go` or new file): positive
   (`write_paths` + `tool_access: none` on agent, and on branch) + negative (each alone).
3. **Validator DIP142**: positive (empty-present, absolute path, `..` escape; agent +
   branch, branch-qualified message) + negative (ordinary relative globs).
4. **Round-trip** (`migrate/roundtrip_test.go`): a multi-glob `write_paths` on an agent
   survives `.dip → export → Migrate`; extend `TestRoundtripBlockFormParallel` so one
   branch carries `write_paths` (covers the branch-encoder separator risk).
5. **Formatter**: a branch with only `write_paths` set re-emits it (guards the
   `writeBranch` early-return regression).
6. **Parity coverage**: a branch-only `write_paths` difference is detected.
7. **Integration** (`validator/lint_examples_test.go::TestLintExamples`):
   `examples/agent_write_paths.dip` produces zero warnings; existing examples don't
   regress.

### Tracker-side tests (specified for cross-repo verification)

- **Red-team (the PR#25 vector):** agent with `write_paths: workspace/**`; mock LLM
  emits `bash("echo pwned > /etc/x")`, `write("../escape.rs", …)`, `bash("cargo build")`.
  Assert zero writes outside `workspace/**` and the build is blocked.
- **Bypass:** `write_paths` + a `Params` working-dir/permission override → override
  ignored.
- **Branch inherit:** agent `write_paths: workspace/**` fanned out through a branch with
  empty `write_paths` → branch still bounded to `workspace/**` (never unbounded).
- **Backend compat:** each backend enforces the jail or refuses session creation.

## Example (`examples/agent_write_paths.dip`)

One lint-clean file exercising the five acceptance shapes (review/synthesis/validation
failure recorders bounded to `workspace/**`-style sentinels, plus a recovery-manager
agent bounded to `.ai/sprints/**` + `.ai/managers/recovery-journal.md`). Mirrors
`examples/agent_tool_access.dip`'s style.

## Decomposition & sequencing

This spec ships **#75 only**, end-to-end. Recorded order for the related follow-ups:

1. **#75 (now)** — the `write_paths` primitive (this work). Agent + per-branch, joint
   tracker release, DIP141/DIP142.
2. **#55 (next)** — `allowed_tools` / `disallowed_tools`. With enforcement at the fs
   layer, tool-name scoping is now a genuinely *orthogonal* axis (which tool names
   exist) rather than a prerequisite. Composes with #75 — dropping `Bash` shrinks the
   jail's attack surface — but neither blocks the other. Still gated on its own blocker:
   a `KnownAgentTools` registry + case-normalization (typo-footgun finding from #41
   round 1).
3. **#53 (last)** — `defaults:` cascade over both `tool_access` and `write_paths`
   (`defaults → agent → branch`). #58's inherit-on-empty already lets a cascade layer
   cleanly. Gated on the opt-out-spelling / live-catalog-semantics decision.

Each is its own spec → plan → PR. This spec records the order; it does not implement #55
or #53.

## Docs

- `docs/nodes.md` — agent-node field list + per-branch field list gain `write_paths`;
  one sentence on the inherit-on-empty rule and the comma-list/no-brace-expansion
  limitation.
- `site/static/skill.md` — `write_paths` in the agent-node section: the glob-list shape,
  the fs-jail threat model it bounds (PR#25 vector), the chain caveat and cross-node
  non-goals with links to #55/#53/#56, and the tracker version requirement. One sentence
  that it is settable per-branch, pointing at the agent-level semantics.
- `docs/validation.md` — DIP141 + DIP142 entries (Trigger / Fix / Example).
- **Do not touch** `CHANGELOG.md` (tag-time) or hand-edit `docs/generated-spec.md`
  (pre-commit regenerates it).

## Known limitations

- **No brace-expansion globs.** `write_paths` is comma-split (consistent with
  `reads`/`writes`/`outputs`), so `*.{md,yaml}` is not expressible — its comma is a list
  separator. The five sites don't need it; authors enumerate entries instead. Documented
  in `docs/nodes.md`.
- **Glob dialect is tracker's.** dippin does not interpret glob semantics (`**` depth,
  character classes); it lints only for obvious unsafe shapes (absolute, `..` escape).
  The exact match semantics are the tracker contract.

## Release coordination

1. Tracker PR opens first: fs-jail enforcement, per-backend behavior, `Params` bypass
   defense, red-team + branch-inherit + backend-compat tests. Lands on tracker `main`,
   untagged.
2. Dippin PR opens in parallel (commit-pinned `go.mod` to the tracker SHA during the
   window so integration tests run against the merged-but-untagged tracker).
3. Spec records the pinned tracker SHA before dippin merges.
4. Tracker tag, then dippin tag, `go.mod` bumped, CHANGELOG updated at tag time.

PR body: **Fixes #75**; references #55/#53 as sequenced follow-ups.
