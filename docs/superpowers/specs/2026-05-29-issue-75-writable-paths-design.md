# Issue #75 — `writable_paths:` path-bounded write-scope primitive

**Date:** 2026-05-29
**Issue:** [#75](https://github.com/2389-research/dippin-lang/issues/75) (write-scope tier, deferred from #41 follow-ups)
**Related roadmap:** #55 (tool-name allowlists), #53 (defaults cascade), #54 (read-only tier), #58 (per-branch override), #56 (chain-attack mitigation), #67/#77 (symlink-escape rejection)
**Target version:** dippin **v0.35.0** (repo is at v0.34.0); joint with a new tracker tag.
**Status:** Design — revised after 5-reviewer expert panel (DSL, security, round-trip, lint, release). See § Design journey.
**Tracker dependency:** Joint release. The pinned tracker SHA is recorded here before dippin merges (mirrors #41).

> **Revision note (rev 2).** The first draft used the name `write_paths`, claimed an
> "fs-jail covering Bash" as a delivered contract, failed *open* on empty values, and
> asserted a `splitComma`-based empty-collapse that the code does not actually do. The
> expert panel found these (and a non-compiling parity change). This revision: renames
> to **`writable_paths`** (the existing advisory `writes:` field made `write_paths:` a
> silent-no-op footgun), makes empty/malformed **fail closed**, scopes enforcement to
> the **native backend** (others refuse-to-start), and reframes the jail as the
> *required tracker contract with enumerated residual-escape classes* rather than a
> delivered guarantee.

## Problem

`tool_access: none` (v0.32.0, #41) is binary: an agent has either no LLM tools or the
full file-mutation catalog (Read, Write, Edit, ApplyPatch, Glob, GrepSearch, Bash).
Five real verdict/failure-recording agent sites in
[`2389-research/pipelines`](https://github.com/2389-research/pipelines) need to
**write** a small, prompt-known set of files — and only those — but sit at the wrong end
of the binary:

- With `tool_access: none`: their legitimate single-file write breaks at runtime.
- Without `tool_access:`: they keep the full mutation catalog, which empirically leaks
  via the scope-creep mode the v0.28.2 incident demonstrated.

Distinct from the sibling follow-ups: #54 (`read_only`) does not help sites that must
*write*; #55 (`allowed_tools`) scopes tool *names*, not *paths* — granting `Write` still
lets the agent write anywhere.

### Real incident — prompt-only constraints failed (PR#25)

`2389-research/pipelines` [PR #25](https://github.com/2389-research/pipelines/pull/25)
bug 4: `redecompose_single`, an agent whose prompt restricted it to writing two YAML
files, instead wrote arbitrary Rust source, ran `cargo build` / `cargo test`, and
fabricated ledger state (Rust target, `reasoning_effort: high`). Prompt discipline did
not hold — the agent had `Write`/`Edit`/`Bash` and used them.

A structural write-scope primitive enforced at the runtime boundary would have bounded
*where* the agent could write. **It would not, by itself, have stopped the fabricated
ledger state** if the ledger lived under an allowed path, nor the network access `cargo`
needs — see § Threat model for the honest scope. The primitive's value is bounding the
*location* of writes, which addresses the "wrote arbitrary files all over the tree" half
of the incident.

### Motivating sites (acceptance fixtures)

From PR #23's Phase-2 reclassification table (the five sites tagged **D**/**C+D**):

| Site | File:line | Writes |
|---|---|---|
| `L6Failed` | `greenfield/greenfield_review.dip:195` | `workspace/.review-failed` |
| `L7Failed` | `greenfield/greenfield_review.dip:328` | `workspace/.review-failed` |
| `Gate1Failed` | `greenfield/greenfield_synthesis.dip:455` | `workspace/.synthesis-failed` (+ reads `gate1-findings.md`) |
| `Gate2Failed` | `greenfield/greenfield_validation.dip:202` | `workspace/.validation-failed` |
| `RecoveryManager` | `sprint/sprint_exec_yaml_v2.dip:741` | `.ai/sprints/SPRINT-<id>-recovery-analysis.md` + `.ai/sprints/SPRINT-<id>-redecompose-request.yaml` + `.ai/managers/recovery-journal.md` |

`RecoveryManager`'s `.ai/` paths are why a fixed "named workspace tier" is insufficient —
the primitive must express arbitrary author-chosen globs.

**Adoption note:** because v1 enforces only on the **native** backend (§ Backends), a
site can adopt `writable_paths` only if it runs (or moves to) the native backend. Sites
on `claude-code`/`acp` get a refuse-to-start until those backends gain jail support;
that is a deliberate fail-closed posture, not a silent no-op.

## Design

One new field. A glob list. Same "dippin carries + lints; tracker enforces" contract as
#41/#58.

```dip
agent RecoveryManager
  prompt: "Record the recovery analysis and journal entry."
  writable_paths: .ai/sprints/**, .ai/managers/recovery-journal.md
```

`writable_paths` is a comma-separated glob list on `AgentConfig` and `BranchConfig`.

- **Absent** → unbounded writes (current behavior).
- **Non-empty** → tracker bounds all file mutations to the listed globs (native backend).
- **Present-but-empty / malformed / unrecognized** → **fail closed**: tracker denies all
  writes (or refuses to start). Never falls through to unbounded. See § Fail-closed.

### Why `writable_paths`, not `write_paths`

dippin already has an advisory `writes:` field (the IO metadata listing *context keys*
a node produces, DIP107). `write_paths:` sits one fat-finger from `writes:`, and the
failure is silent: an author who types `writes: workspace/out.md` intending to bound a
file gets advisory context-key linting and **no jail** — reproducing the PR#25
"restriction that silently isn't one" failure mode. `writable_paths` is visually and
semantically distinct ("the paths this agent may write"), parallels the noun-phrase
style of `working_dir`, and removes the prefix collision.

### Why a separate field, not overloaded `tool_access`

Three shapes were considered (issue #75 sketches them):

1. **Named tier** (`tool_access: workspace_only`, tracker-defined glob). Cannot express
   `RecoveryManager`'s `.ai/` paths. **Rejected** (fails acceptance).
2. **Overloaded `tool_access`** (`tool_access: { ... }`). The structured-`tool_access`
   design #41 explicitly rejected after three review rounds; a breaking type change to a
   field tracker already consumes as a string. **Rejected.**
3. **Separate `writable_paths` list field**, orthogonal to the `tool_access` scalar.
   Composition handled by lints, not field entanglement. **Chosen.**

`tool_access` stays a dumb scalar. `writable_paths` is a dumb glob list. dippin does not
resolve, coerce, or canonicalize globs (lint-time escape *detection* is read-only and
not written back to IR — § Lints). Tracker owns runtime semantics, as #41 established.

### Semantic model (normative)

- **Effective resolution (tracker-side):** `writable_paths` bounds the set of paths any
  file-mutating tool may write, **resolved against an immutable session root** (§ Anchor).
- **Branch inherit-on-empty** (mirrors #58's `ToolAccess` rule): an empty branch
  `writable_paths` **inherits the target agent's**; it never resets to unbounded.
  Effective = `branch.WritablePaths if non-empty else agentNode.WritablePaths`. A branch
  may set or narrow scope by declaring globs; an omitted branch value leaves the target's
  bound in force. If an omitted branch resolved to *unbounded*, a path-locked agent
  fanned out through a branch that omits the field would silently run unbounded — an
  invisible loosening of the exact primitive `writable_paths` exists to provide.
  Inherit-on-empty is the safe default and the only one that composes with the future
  defaults cascade (#53).

### Fail-closed (empty / malformed / unrecognized)

This is a **safety polarity decision**: `writable_paths` fails *closed*, matching #41's
`tool_access` ("any non-empty value disables tools; a typo still disables them"), not the
advisory list fields.

- **dippin side:** the parser uses `splitCommaNoEmpty` (not `splitComma`) for
  `writable_paths`, so stray/blank entries are dropped and a bare `writable_paths:`
  becomes `nil` (== absent in dippin's IR). This is a **deliberate deviation** from
  `reads`/`writes`/`outputs` (which use the empty-keeping `splitComma`); it is called out
  here because it makes both round-trip paths agree (§ Round-trip) and avoids re-emitting
  a malformed `writable_paths:` line. *Consequence:* dippin's IR cannot distinguish "bare
  `writable_paths:`" from "field absent," so dippin does **not** emit a lint for the
  empty case — the fail-closed backstop is tracker-side.
- **Tracker side (normative):** tracker re-parses the raw `.dip` bytes (via `.dipx`), so
  it *can* see a present-but-empty `writable_paths:` in the source text. Tracker MUST
  treat a `writable_paths` that is present-but-empty, all-blank, malformed (unparseable
  glob, brace-expansion mis-split — § Known limitations), or **unrecognized by an older
  tracker that predates this field** as **deny-all-writes or refuse-to-start** — never as
  unbounded. The version-skew case (§ Release coordination) is the most dangerous: an old
  tracker that silently ignores the field is the PR#25 incident in disguise.

## Threat model & dippin/tracker split

**dippin CARRIES + LINTS; tracker ENFORCES.** dippin stores the raw glob list,
round-trips it (parse → IR → format → DOT export → migrate), and lints it for obvious
unsafe *shapes*. dippin cannot enforce path bounding — it emits IR/`.dipx`; tracker
resolves and enforces.

### Required tracker enforcement contract (normative)

This section is the contract the joint tracker PR must satisfy. It is written as
*requirements*, not as a delivered guarantee — the dippin reviewer verifies the tracker
PR against it.

1. **Write containment.** On the native backend, every file mutation performed by any
   tool the agent can invoke — `Write`/`Edit`/`ApplyPatch` **and `Bash` and any process
   `Bash` spawns** — must be confined to paths matching `writable_paths`, resolved
   against the session root. Tool-argument validation alone is **insufficient**: `Bash`
   is a write primitive and its children inherit the confinement.
2. **Symlink-chain resolution at write time.** A write whose *resolved real path* (full
   parent symlink chain, `filepath.EvalSymlinks`-equivalent) escapes the allowed set must
   be denied — including a symlink the agent **forges at runtime inside an allowed
   directory** pointing out of the jail. This is the actual lesson of #67/#77
   (`parser/resolve.go`'s `checkContainment` resolves the full chain); lexical lint
   cannot see a runtime-forged symlink. TOCTOU between glob-check and write must be closed
   (check at the moment of write, on the resolved path).
3. **Immutable anchor (§ Anchor).** Globs resolve against a fixed session root.
   `working_dir` and any `Params` key MUST NOT relocate the anchor.
4. **Fail-closed (§ Fail-closed).** Empty/malformed/unrecognized → deny-all or refuse.
5. **Bypass defense.** When `writable_paths` is set, tracker MUST NOT honor a `Params`
   key (or the first-class `working_dir` field) that widens or relocates the write
   surface. Mirrors #41's `Params` bypass defense, extended to the `working_dir` field.

#### Anchor

`writable_paths` globs resolve against an **immutable session root** chosen by tracker at
session creation. `working_dir` is a first-class `AgentConfig` field (`ir/ir.go`), not a
`Params` key; if globs resolved relative to `working_dir`, then `working_dir` would be a
jail-relocation primitive. The contract: the anchor is fixed at session start and neither
`working_dir` nor `Params` can move it. Tested by the `working_dir` relocation red-team
case below.

### Residual escape classes (explicitly in/out of scope)

The primitive bounds *where file writes land*. It does **not** bound:

- **Network.** `writable_paths` says nothing about network; `bash("curl -d @secret …")`
  and `cargo`'s crate fetches remain permitted. **Out of scope** (orthogonal control).
- **Reads / exfil-by-read.** Bounds writes only; reading `.env` and laundering it
  (network, or via an allowed write consumed downstream) is the chain caveat (#56).
- **Content within an allowed path (granularity).** An agent with
  `writable_paths: workspace/**` can still poison `workspace/Cargo.toml`, CI configs,
  lockfiles, or fabricate state in a ledger that lives under `workspace/` (the PR#25
  fabricated-ledger element). `writable_paths` bounds *location*, not *trustworthiness of
  content*; it is only as safe as the blast radius of the allowed location. Narrow globs
  (single sentinel files, as in the motivating sites) are where it is strongest.
- **Hardlinks / inherited FDs / `/proc` re-entry.** A perfect path jail must also resist
  hardlinks to outside inodes, writes through file descriptors inherited from the parent
  process (no path syscall to intercept), and `/proc/self/...` re-entry. These are
  tracker enforcement-mechanism concerns; the contract requires them to be addressed (or
  the residual risk documented) in the tracker PR. The dippin spec names them so the
  reviewer checks the tracker red-team covers them, rather than letting "fs jail" imply
  they are handled.

### Backends

v1 enforces on the **native** backend only. On `claude-code` and `acp`, where tracker
does not control the child process's sandbox, an fs-level jail over `Bash` may be
infeasible — and reaching for the backend's *prompt-level* `--allowedTools`/permission
flags is **not** enforcement (that is the prompt-discipline failure PR#25 demonstrates).
Therefore, when `writable_paths` is set on an agent whose backend cannot enforce the
jail, **session creation refuses to start** with an error pointing to this issue —
authors get a hard failure, never a silent unbounded run or a prompt-level pretend-jail.

**Late-failure caveat (carried from #41).** Runtime refusal can land deep in a long
pipeline — the motivating sites are terminal nodes (`:195`/`:328`/`:455`/`:741`). The
implementation plan records, against the pinned tracker SHA, which backends are
jail-capable, so the dippin reviewer can sanity-check coverage. A dippin pre-flight lint
is *not* added in v1 (dippin doesn't know tracker's per-backend capability); if late
failures prove painful, a follow-up adds one.

### Chain caveat (#56)

`writable_paths` bounds a single agent's *immediate* writes, not information flow. A
`writable_paths` agent can launder data through an allowed file into a downstream
unbounded agent — or into a later **tool node** / CI step that executes the allowed file
(linking the chain caveat to the granularity caveat above). Out of scope; tracked in #56.

## Dippin-side design

### IR (`ir/ir.go`)

`AgentConfig` gains `WritablePaths []string`, clustered with `ToolAccess`:

```go
// WritablePaths bounds the file paths this agent's tools may write, as author-chosen
// globs (e.g. "workspace/**", ".ai/sprints/**") resolved against the session root.
// Empty/absent = unbounded. A present-but-empty or malformed value fails CLOSED at
// the tracker (deny-all / refuse-to-start), never unbounded. dippin carries + lints;
// tracker enforces an fs-level write jail on the native backend (Bash + its children
// included); claude-code/acp refuse to start. See issue #75.
WritablePaths []string
```

`BranchConfig` gains `WritablePaths []string` with the inherit-on-empty rationale doc
comment (mirroring the existing `ToolAccess` branch comment): empty INHERITS the target
agent's `writable_paths`, never resets to unbounded.

> **Comparability note:** adding a slice field makes `BranchConfig` non-comparable, which
> breaks `migrate/parity.go`'s struct-`!=` comparison. The parity change (below) is
> mandatory, not optional.

### Parser (`parser/parse_nodes.go`)

- Agent runtime-field handler: `case "writable_paths": cfg.WritablePaths = splitCommaNoEmpty(val)`.
- `applyBranchField`: `case "writable_paths": b.WritablePaths = splitCommaNoEmpty(val)`
  (else a branch value is silently discarded and the lint never sees it).

`splitCommaNoEmpty` (not `splitComma`) — the deliberate deviation justified in
§ Fail-closed. Stores verbatim otherwise; no glob validation in the parser.

> **Complexity:** `applyBranchField` is at cyclo 5 today (4-case switch); a 5th case
> breaches the cap. Extract the branch-field dispatch (e.g. a map or a split helper) so
> it stays ≤5. Check `applyAgentRuntimeField` similarly.

### Formatter (`formatter/format.go`)

- Agent: emit in the agent runtime cluster when `len(cfg.WritablePaths) > 0`:
  `wr.line("writable_paths: %s", strings.Join(cfg.WritablePaths, ", "))`.
- Branch: `writeBranchFields` emits `writable_paths` when non-empty; the `writeBranch`
  early-return **guard** gains `&& len(b.WritablePaths) == 0`.

> **Complexity:** `writeBranch` and `writeBranchFields` are both at cyclo 5 today. The
> guard addition and the new emit each breach the cap — extract helpers (e.g. a
> `branchHasFields(b)` predicate and a slice-field emit helper). Budget this.

### DOT export (`export/dot.go`)

- Agent: `attrs["writable_paths"] = strings.Join(cfg.WritablePaths, ",")` via
  `applyAgentRuntimeAttrs` when non-empty.
- Branch: encode via `appendBranchField(parts, "writable_paths", strings.Join(b.WritablePaths, ","))`.
  The joined value passes through `encodeBranchToken`, which percent-encodes `,`→`%2C`
  (and `;`/`=`/`%`/`\`), so the comma-joined list survives both the branch separator (`,`)
  and the field separator (`;`). **Verified safe** by three reviewers against
  `export/dot.go` — no list-aware path or alternate join char is needed. The round-trip
  test (below) is the regression guard.
- **Do NOT add `writable_paths` to `reservedGraphAttrs`.** That map is graph-level only
  (it filters `Workflow.Vars` on re-export); `writable_paths` is a node attr, like
  `tool_access` (which is also not in that map). The first draft's instruction was wrong.

### Migrate (`migrate/migrate.go`)

- Agent: `if v, ok := attrs["writable_paths"]; ok { cfg.WritablePaths = splitComma(v) }`
  using migrate's own `splitComma` (which already drops empties — consistent with the
  parser's `splitCommaNoEmpty`, so both round-trip paths produce identical IR).
- Branch: add `"writable_paths"` to `branchFieldSetters`:
  `func(b *ir.BranchConfig, v string) { b.WritablePaths = splitComma(v) }`. The setter
  value is **already** `decodeBranchToken`'d by the caller — split only, do not decode
  again. (Not the literal one-liner the scalar entries are — it needs the split wrapper.)
- Parity (`migrate/parity.go`):
  - **Rewrite `compareParallelBranches`** to compare field-by-field instead of struct
    `!=` (which no longer compiles): scalar fields + `strings.Join(WritablePaths, ",")`.
    Extract a `branchesEqual(a, b)` helper to stay under the complexity cap.
  - Add `WritablePaths` join-compare to the agent comparator (`compareAgentBehavior` —
    `AgentConfig` has no slice field today, so this is a new compare, not a reuse of the
    `Outputs` comparator which lives on `ToolConfig`).

### Editor support & misc touch sites

- `lsp/completion.go` — add `{"writable_paths:", "<help>"}` to the curated agent-field
  completion table (authoring UX; no round-trip impact).
- `validator/lint.go` header comment range bump (`DIP101–DIP140` → `DIP142`).
- **No change** to `diff/diff.go` (its `agentFieldTable` is intentionally a partial
  subset that already omits `tool_access`/`working_dir`), `simulate/*`, `flatten/*`,
  `coverage/*`, `doctor/*`, `dipx/*` (all read only specific behavioral fields or raw
  bytes; `writable_paths` is invisible to them) — confirmed by reviewers.

### Pack-time validation (`cmd/dippin/cmd_pack.go`)

DIP141/DIP142 are warning-severity at pack time (only DIP001–009 block `dippin pack`),
consistent with DIP139/DIP140. The fail-closed runtime backstop (not a pack error) is the
safety guarantee for malformed values.

## Lint rules (2 new codes)

New file `validator/lint_writable_paths.go` (keeps DIP139's already-tight functions in
`lint_tool_access.go` undisturbed). Helper ladder (each ≤ cyclo 5 / cognit 7), modeled on
`lint_style.go`'s `checkNodeFidelityByKind → checkBranchFidelities → checkFidelityValue`:

```
lintWritablePaths(w)                      // range nodes -> checkNodeWritablePathsByKind
checkNodeWritablePathsByKind(n)           // switch AgentConfig / ParallelConfig
checkBranchWritablePaths(n, branches)     // range branches -> checkWritablePathsObject
checkWritablePathsObject(n, paths, ta, branch)
                                          //   DIP141 if dip141Triggers(paths, ta)
                                          //   + range entries -> unsafeEntryKind
dip141Triggers(paths, toolAccess) bool    // len(paths)>0 && normalize(ta)=="none"
unsafeEntryKind(entry) string             // "" | "absolute" | "escape" | "brace"
isAbsoluteEntry(e) / isEscapeEntry(e)     // split out if unsafeEntryKind exceeds cyclo 5
```

### DIP141 — `writable_paths` nullified by `tool_access: none`

Fires when a single `AgentConfig` or `BranchConfig` has non-empty `writable_paths` **and**
`tool_access` normalizing to `none` (reuse DIP139's `strings.ToLower(strings.TrimSpace())`
so `None`/`NONE ` trip it). `none` strips the entire catalog, so there is nothing to
bound — dead config. Warning. Scans agent + branch (branch-qualified message via
`b.Target`).

Only the *same config object* declaring both is flagged. A branch that sets
`tool_access: none` while *inheriting* an agent's `writable_paths` is legitimate narrowing
(the branch chose no-tools), not author error — not flagged.

> When #55 lands, this dead-config check extends to `writable_paths` set alongside a
> tool-name allowlist that excludes all write tools — same flavor, deferred to #55's spec.

Message: `node %q has writable_paths but tool_access "none" — none strips all tools, so there is nothing to bound (dead config)` (branch form adds `branch %q`).
Help: `remove writable_paths (no tools to bound) or drop tool_access: none to grant a bounded tool catalog.`

### DIP142 — unsafe `writable_paths` entry

Fires per-entry when an entry will not bound writes the way the jail expects. **This is an
author-clarity lint, not an escape control** — the fs-jail is the real boundary; the help
text says so. The empty-list case is **not** a dippin lint (it is unobservable post-
`splitCommaNoEmpty`; the tracker fail-closes on it — § Fail-closed). Predicate set
(reuse `parser/resolve.go` helpers for the escape check rather than reimplementing):

- **Absolute** — leading `/`, OR Windows shape `^[A-Za-z]:` / `^\\` (checked
  build-OS-independently, since tracker may run a different OS), OR leading `~` (home
  expansion escapes the workspace too).
- **Parent escape** — `hasParentRef(filepath.Clean(e))` (the `parser/resolve.go`
  predicate), catching `../x` and `foo/../../bar`; NOT a naive `..` substring (which
  false-positives on `..bar`). Detection only — the cleaned path is not written back to
  IR, preserving authoring fidelity and the "no canonicalization" boundary.
- **Brace mis-split** — an entry with an unbalanced `{` or `}` (a `*.{md,yaml}` glob the
  comma-split tore apart — § Known limitations).

Warning. Scans agent + branch.
Message: `node %q writable_paths entry %q escapes the workspace (absolute / ~ / parent path) — the tracker write-jail will not honor it` (brace variant: `… is malformed (brace expansion is split on commas)`).
Help: `use workspace-relative globs (e.g. .ai/sprints/**). Absolute, ~, and ..-escaping entries are rejected by the fs jail (it bounds writes to the session root); this lint catches obvious lexical cases only — the runtime jail is the real boundary. See #67/#77.`

> DIP141 and DIP142 may both fire on one object (e.g. `writable_paths: /etc/**` +
> `tool_access: none`) — they describe different defects; emitting both is intentional
> and consistent with DIP139/DIP140's non-deduplication.

### Registration touch-points (all four)

1. `validator/lint_codes.go` — `const` block (DIP141, DIP142) **and** the
   `CodeDescription` init map.
2. `validator/explanations.go` — `Explanations` map entries with Trigger/Fix/Example at
   DIP139's depth (Example annotated with the code; Trigger names the PR#25 *why*).
3. `validator/lint.go` — header range comment bump.
4. Wire `lintWritablePaths` into the validator's lint dispatch.

## Tests (TDD)

1. **Parser** (`parser/parse_writable_paths_test.go`, new): `writable_paths: a, b` on an
   agent → `["a","b"]`; inside a `branch:` block → `BranchConfig.WritablePaths`; bare
   `writable_paths:` and `writable_paths: ,` → `nil` (via `splitCommaNoEmpty`); stray
   `a,,b` → `["a","b"]`. Real parser, no hand-built IR.
2. **Validator DIP141**: positive (`writable_paths` + `tool_access: none` on agent, and on
   branch; case-variant `None`) + negative (each alone; branch inheriting under
   `tool_access: none`).
3. **Validator DIP142**: positive (absolute `/etc/**`, `~/x`, Windows `C:\x`, `../../x`,
   brace `*.{md`; agent + branch, branch-qualified) + negative (`workspace/**`,
   `.ai/sprints/**`).
4. **Round-trip** (`migrate/roundtrip_test.go`): a multi-glob `writable_paths` on an agent
   survives `.dip → export → Migrate` identically via **both** the format and the DOT
   paths (guards the parser/migrate `splitComma*` agreement); extend
   `TestRoundtripBlockFormParallel` so one branch carries `writable_paths` (covers the
   percent-encode branch path).
5. **Formatter**: a branch with only `writable_paths` set re-emits it (guards the
   `writeBranch` early-return regression).
6. **Parity coverage**: a branch-only and an agent-only `writable_paths` difference are
   each detected by the rewritten comparators (regression guard for the non-comparable-
   struct fix).
7. **Example lint assertion** (dedicated test — **not** `TestLintExamples`, which only
   checks DIP108): lint `examples/agent_writable_paths.dip` and assert zero DIP141/DIP142.

### Tracker-side tests (specified for cross-repo verification; restore #41 rigor)

- **Red-team (single-turn, PR#25/v0.28.2 shape):** agent with
  `writable_paths: workspace/**`; mock LLM emits, **in one response/turn**, multiple tool
  calls: `write("../escape.rs", …)`, `bash("echo pwned > /etc/x")`, `bash("cargo build")`.
  Assert zero writes outside `workspace/**` and that a **child process spawned by Bash**
  attempting an outside write is also blocked.
- **Runtime symlink escape:** `bash("ln -s /etc workspace/link && echo x > workspace/link/y")`
  → denied (resolved real path escapes).
- **`working_dir` relocation:** `writable_paths: workspace/**` + `working_dir: /` (or `..`)
  → still bounded to the original session-root anchor.
- **Fail-closed:** present-but-empty `writable_paths:`, a malformed/uncompilable glob, and
  (version-skew) a tracker that doesn't recognize the field → deny-all / refuse, never
  unbounded.
- **Bypass:** `Params` working-dir / permission override ignored.
- **Branch inherit:** agent `writable_paths: workspace/**` fanned out through a branch with
  empty `writable_paths` → branch still bounded (never unbounded).
- **Backend:** native enforces; `claude-code`/`acp` refuse to start.

## Example (`examples/agent_writable_paths.dip`)

One lint-clean file exercising the five acceptance shapes (review/synthesis/validation
failure recorders bounded to `workspace/**`-style sentinels, plus a recovery-manager
agent bounded to `.ai/sprints/**` + `.ai/managers/recovery-journal.md`). Mirrors
`examples/agent_tool_access.dip`'s style; the dedicated test above asserts it lints clean.

## Known limitations

- **No brace-expansion globs.** `writable_paths` is comma-split, so `*.{md,yaml}` is torn
  into `*.{md` and `yaml}` — not merely "not expressible" but actively mis-split. DIP142's
  brace check flags the unbalanced fragments at lint time; at runtime the fragments match
  nothing and fail closed. Authors enumerate entries instead. Documented in `docs/nodes.md`.
- **Glob dialect is tracker's.** dippin does not interpret glob match semantics (`**`
  depth, character classes); it lints only obvious unsafe lexical shapes (§ DIP142). The
  exact match semantics are the tracker contract.
- **Bounds location, not content or network** — see § Residual escape classes.

## Decomposition & sequencing

This spec ships **#75 only**, end-to-end. Recorded order for the related follow-ups (the
implementation plan's **task #0** files/records #55/#53/#56 follow-ups as numbered issues
and updates skill.md cross-references — restoring the #41 discipline DIP28 dropped):

1. **#75 (now)** — the `writable_paths` primitive. Agent + per-branch, native-backend
   enforcement, joint tracker release, DIP141/DIP142.
2. **#55 (next)** — `allowed_tools` / `disallowed_tools`. With enforcement at the fs layer,
   tool-name scoping is a genuinely *orthogonal* axis (which tool names exist), not a
   prerequisite. Composes with #75 (dropping `Bash` shrinks the jail's attack surface and
   the network/granularity residuals) but neither blocks the other. Gated on its own
   blocker: a `KnownAgentTools` registry + case-normalization (#41 round-1 typo footgun).
3. **#53 (last)** — `defaults:` cascade over both `tool_access` and `writable_paths`
   (`defaults → agent → branch`). #58's inherit-on-empty already lets a cascade layer
   cleanly. Gated on the opt-out-spelling / live-catalog decision. The inherit-on-empty
   composition is asserted, not proven, until #53 is designed.

Each is its own spec → plan → PR.

## Docs

- `docs/nodes.md` — agent + per-branch field lists gain `writable_paths`; the inherit-on-
  empty rule; the comma-list / no-brace-expansion limitation; the `writes:` vs
  `writable_paths:` distinction (one explicit sentence — advisory context keys vs enforced
  file globs).
- `site/static/skill.md` — `writable_paths` in the agent-node section: the glob-list shape,
  the **native-backend** enforcement + fail-closed + immutable-anchor contract, the
  residual-escape scope (network/content/reads out of scope), the chain caveat and
  cross-node non-goals with links to #55/#53/#56, and the **`requires tracker ≥ <tag>`
  safety requirement** (§ Release coordination). One sentence that it is settable per-branch.
- `docs/validation.md` — DIP141 + DIP142 entries (Trigger / Fix / Example).
- **Count updates:** `CLAUDE.md` (line ~85, "49 … DIP101-DIP140") and
  `docs/llm-reference.md` (line ~188, "49 diagnostic codes") → 51 / DIP101-DIP142.
- **Do not touch** `CHANGELOG.md` (tag-time) or hand-edit `docs/generated-spec.md`
  (pre-commit regenerates it).

## Release coordination

**Module-dependency direction (corrected from the first draft):** dippin has **no** Go
dependency on tracker — tracker imports dippin's `ir`/`dipx`. So there is no dippin-side
`go.mod` pin to tracker. Cross-repo integration tests live in the **tracker** repo, which
pins a dippin commit SHA in *its* `go.mod` during the window. The dippin PR carries the
field + lints + round-trip tests, which need no tracker dependency.

1. **Tracker PR opens first:** native fs-jail enforcement (Bash + children + symlink-chain
   + immutable anchor), `claude-code`/`acp` refuse-to-start, fail-closed on
   empty/malformed/unrecognized, `Params`/`working_dir` bypass defense, and the full
   red-team suite above. Lands on tracker `main`, untagged; tracker `go.mod` pins the
   dippin PR SHA for its integration tests.
2. **Dippin PR opens in parallel** (no tracker dep): field, lints, round-trip/parity
   tests, example, docs.
3. **Spec records the pinned tracker SHA** here before dippin merges.
4. **Version-skew safety statement** goes in skill.md and (at tag time) CHANGELOG, naming
   the failure mode explicitly: *without a tracker ≥ the paired tag, `writable_paths` is
   not enforced; the paired tracker fail-closes on the field, so an unpinned/older tracker
   must refuse rather than run unbounded.* This mirrors v0.32.0's "lint-validated runtime-
   no-op safety fields ship as worse-than-nothing" language.
5. **Tracker tag first, then dippin tag (v0.35.0)** immediately after, `go.mod` bumped on
   the tracker side, CHANGELOG updated at tag time. Tracker takes a minor bump (new
   `SessionConfig` semantics).

PR body: **Fixes #75**; references #55/#53/#56 as sequenced follow-ups.

## Design journey

Brainstormed (shape, sequencing, threat model, dippin/tracker split) → drafted →
5-reviewer expert panel (DSL-design, security, round-trip, lint, release). The panel
found, and this revision folds in: the `splitComma("")→[""]` empty-handling bug (4
reviewers) → `splitCommaNoEmpty` + tracker fail-closed; the non-compiling
`BranchConfig`-slice parity change → field-by-field comparator; the over-claimed
"fs-jail covering Bash" → required-contract framing with enumerated residual escapes and
native-only scope; the `working_dir` relocation bypass → immutable anchor; the
version-skew silent-unbounded → fail-closed + `requires tracker ≥ tag`; the backwards
`go.mod` pin direction → tracker-pins-dippin; the `writes:`/`write_paths:` collision →
`writable_paths`; plus complexity budgeting (4 near-cap functions), the lint helper
ladder, restored #41 red-team rigor, and the DIP-count/touch-site corrections. The
branch-encoder comma "risk" the draft flagged was verified a non-issue (percent-encoding).
