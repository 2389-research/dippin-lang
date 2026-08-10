# v0.32.0 — `tool_access:` agent-node safety primitive (issue #41)

**Date:** 2026-05-26
**Closes:** [#41](https://github.com/2389-research/lang/issues/41)
**Status:** Design approved 2026-05-26
**Runtime enforcement:** Pinned runtime PR/commit recorded in implementation plan
**Research:** [`docs/superpowers/research/2026-05-19-issue-41-terror-squad.md`](../research/2026-05-19-issue-41-terror-squad.md) (round-1 findings, parked v0.30.0); rounds 2–4 reviewer transcripts consolidated below

## Problem

`.dip` agent nodes always ship the full file-mutation tool catalog (Read, Write, Edit, ApplyPatch, Glob, GrepSearch, Bash). The v0.28.2 runaway-agent incident — a 10-minute, 39k-token run where an agent implemented an entire Go project from a `SPEC.md` it found on disk during what should have been a passthrough "acknowledge" node — remains structurally easy to hit. Existing mitigations are inadequate:

- `max_turns: 1` does not bound damage within a single turn (engine executes all tool calls inside turn N=1 before checking the cap; the runtime's session layer handles this check).
- `allowed_tools` / `disallowed_tools` work only with `backend: claude-code`; the native backend silently drops them.
- HARD CONSTRAINT prompt text is empirically unreliable (the v0.28.2 incident occurred with such text in place).

Dippin needs a per-agent runtime-enforced primitive that strips the LLM's tool catalog.

## Design

One field. One value. Coordinated runtime release.

```dip
agent ReportFinalStatus
  prompt: "Summarize what was implemented."
  tool_access: none      # no LLM tools — bounded summarizer
```

`tool_access: none` removes the LLM's tool registry on this agent's session and sets the equivalent of `tool_choice: none` on the underlying API request. The field is omitted on agents that should retain full tool access (current behavior).

## Non-goals (v1)

These are tracked as numbered follow-up issues, filed before merge (see § Release coordination):

1. **`defaults:`-block cascade.** Workflow-level `defaults: tool_access: none` with per-node opt-out. Requires an explicit opt-out spelling (`tool_access: full` or similar) and shipping that spelling commits to a live-catalog semantic that will tax future evolution. Defer until incident data shows per-node annotation is insufficient. ([#53](https://github.com/2389-research/dippin-lang/issues/53))
2. **Middle tier (`read_filesystem` / `read_only`).** A "Read-only summarizer" tier has murky safety semantics — A→disk→B→C laundering chain still works under read-only, and `.env` exfiltration is unbounded. Ships when a richer surface lands with explicit threat-model framing. ([#54](https://github.com/2389-research/dippin-lang/issues/54))
3. **Companion list fields (`disallowed_tools`, `allowed_tools`).** Round-1 finding #2 demonstrated typo footguns (lowercase silently no-ops against the runtime's CamelCase catalog). Defer until a `KnownAgentTools` registry and case-normalization are locked. ([#55](https://github.com/2389-research/dippin-lang/issues/55))
4. **Chain-attack mitigation (`${ctx.last_response}` auto-injection).** `tool_access: none` bounds tools available to THIS agent's LLM call. It does NOT bound information flow between agents. A `tool_access: none` summarizer feeding a `tool_access: full` writer via `${ctx.last_response}` remains a viable chain. Tracked as a separate issue for `last_response_truncate:` or structural context-threading change. ([#56](https://github.com/2389-research/dippin-lang/issues/56))
5. **Cross-node lint (`tool_access: none` source → `full` target edge warning).** Useful once cascade exists; less so when authors annotate per-node. Defer. ([#57](https://github.com/2389-research/dippin-lang/issues/57))
6. **`BranchConfig.ToolAccess` per-branch override.** Block-form parallel nodes' per-branch fields (`Model`/`Provider`/`Fidelity`) already lose values through DOT round-trip today — per-branch DOT emission infrastructure doesn't exist. Fix that first; then add safety field overrides. ([#58](https://github.com/2389-research/dippin-lang/issues/58))
7. **`ManagerLoopConfig.SubgraphRef` cross-workflow propagation.** Parent's restrictive `tool_access:` does not propagate into a child subgraph. Requires multi-workflow IR traversal the validator doesn't have today. ([#59](https://github.com/2389-research/dippin-lang/issues/59))
8. **`Params` bypass lint.** Author writes `params: { allowed_tools: Bash }` to re-enable tools. v1 defends against this on the runtime side (see Runtime-side design below); a dippin lint for the params surface is deferred. ([#60](https://github.com/2389-research/dippin-lang/issues/60))
9. **Tool nodes under `tool_access:`.** Tool nodes execute arbitrary shell unconditionally; DIP28's `tool_commands_allow` / `tool_denylist_add` already cover that surface. skill.md cross-references this distinction. ([#61](https://github.com/2389-research/dippin-lang/issues/61))

## Dippin-side design

### IR (`ir/ir.go`)

Add one field to `AgentConfig`, clustered with the runtime/backend block:

```go
type AgentConfig struct {
    // ... existing fields ...
    Backend     string
    WorkingDir  string
    ToolAccess  string  // "" (default = full catalog) or "none"
    Params      map[string]string
}
```

Plain `string`. Matches `Backend`, `Compaction`, `ReasoningEffort`, etc. The same-shape decision as DIP28's `ToolCommandsAllow` (`WorkflowDefaults.ToolCommandsAllow string`): the runtime owns runtime semantics, IR stays authoring-format-agnostic, no parse-time normalization that could drift from the runtime.

No named type, no constants, no accessor methods, no helper functions. Round-2/3/4 reviewer pressure to add `ToolAccessPolicy` + `EffectiveToolAccess()` + canonical-accessor enforcement was traced to round-2's "fail-closed coercion at IR-consumption layer" framing, which scattered safety logic across 18+ consumers. With per-node-only scope (no cascade), there is exactly one runtime consumer, which does its own normalization. The complexity is gone, not solved.

### Parser (`parser/parse_nodes.go`)

One case in the agent runtime-field handler, in the `backend` / `working_dir` cluster:

```go
case "tool_access":
    cfg.ToolAccess = val
```

Stores verbatim. No normalization, no validation. Matches the established DIP127 / DIP130 pattern (parser dumb; validator owns enum check).

### Validator (`validator/lint_tool_access.go`, new file)

One lint code: **DIP139**, error severity. Fires when `cfg.ToolAccess` is set but does not normalize to a recognized value.

```go
var validToolAccess = map[string]bool{
    "":     true,  // default
    "none": true,  // explicit no-tools
}

func lintToolAccessValues(res *Result, w *ir.Workflow) {
    for _, n := range w.Nodes {
        cfg, ok := n.Config.(ir.AgentConfig)
        if !ok {
            continue
        }
        canonical := strings.ToLower(strings.TrimSpace(cfg.ToolAccess))
        if validToolAccess[canonical] {
            continue
        }
        res.emit(n, DIP139, fmt.Sprintf(
            "node %q has tool_access %q which is not recognized; valid: none (or omit for full catalog)",
            n.ID, cfg.ToolAccess))
    }
}
```

Diagnostic Help (`validator/explanations.go`):

```text
DIP139 Trigger: tool_access value is not "none" (case-insensitive) or empty
DIP139 Fix:     Use `tool_access: none` to disable LLM tools, or omit the field for the full catalog
DIP139 Example: agent X
                  tool_access: nono   // DIP139: typo — runtime falls back to no-tools
```

Authors who skip `dippin lint` and ship an invalid value get fail-closed runtime behavior (see Runtime-side design). DIP139 surfaces the typo when lint runs. The "falls back to none at runtime" sentence is load-bearing — without it, authors who hit the lint might treat it as informational.

Complexity: `lintToolAccessValues` is a single loop with a continue + an emit. Cyclomatic ≤ 3, cognitive ≤ 3.

### Formatter (`formatter/format.go`)

Emit in `writeAgentRuntimeFields` (`format.go:366-373`), after `Backend` and `WorkingDir`. Conditional on non-empty:

```go
if cfg.ToolAccess != "" {
    wr.line("tool_access: %s", quoteValue(cfg.ToolAccess))
}
```

Verbatim. Invalid values survive formatting → re-parsing still trips DIP139.

### DOT export (`export/dot.go`)

Per-node `tool_access` attribute on agent nodes via `applyAgentRuntimeAttrs` (`export/dot.go:280-302`). Verbatim. `reservedGraphAttrs` (`export/dot.go:60`) gains `tool_access` so it can't collide with author `vars:`.

No canonicalization. The runtime handles normalization on its end; dippin preserves authoring fidelity.

### Migrate (`migrate/migrate.go`)

`extractAgentAttrs` reads `tool_access` from DOT attrs into `cfg.ToolAccess`. Unconditional setter. Mirrors the existing `Backend` / `WorkingDir` extraction.

`compareAgentConfigs` (`migrate/parity.go:219-228`) — add `ToolAccess` to the existing field-list. Do NOT bundle a broader `reflect.DeepEqual` upgrade — the existing comparator deliberately ignores ~11 fields that have separate untested round-trip behavior; switching to DeepEqual would surface every one as a new failure.

No new `resolveStartExitKind` heuristic. `tool_access` lives on agent shape's normal path.

### Pack-time validation (`cmd/dippin/cmd_pack.go`)

DIP139 is warning-severity at pack time (matches existing pack-time / lint-time split — only DIP001–009 errors block `dippin pack`). The workflow runs as if `tool_access: none` via the runtime's fail-closed handling; `dippin lint` reports the typo. Spec calls this out explicitly so reviewers don't "helpfully" add error-severity rejection.

## Runtime-side design

The dippin field is meaningless without runtime enforcement. The parking decision (v0.30.0 → v0.32.0) explicitly rejected "lint-validated runtime-no-op safety fields." The runtime ships enforcement in the SAME release.

### Session layer

`SessionConfig` gains `ToolAccess string` populated from the dippin adapter.

### Profile / tool catalog

```go
canonical := strings.ToLower(strings.TrimSpace(cfg.ToolAccess))
if canonical != "" {
    // v1: any non-empty value disables tools (fail-closed for invalid).
    // The only recognized spelling is "none"; invalid values trigger the same
    // restriction so a lint-skipped typo can't ship full tools.
    return []Tool{}
}
// existing full-catalog return
```

One conditional. No enum, no map, no helper. When v2 adds `read_filesystem`, this becomes a switch — but v1 is `is-empty / is-not-empty`.

### Session execution layer

When `cfg.ToolAccess` is non-empty (case-normalized):

1. Set `request.ToolChoice = llm.ToolChoiceNone()` for the Anthropic backend so the translator strips the `tools` array from the API request.
2. **System-prompt scrub.** Omit every prompt element that names a tool. At minimum: the "File tool arguments (read, write, edit, glob, grep_search) MUST use paths relative to the working directory." prefix line. Runtime-side test asserts: when `ToolAccess` is non-empty, the assembled system prompt contains no occurrence of `read`, `write`, `edit`, `glob`, `grep_search`, `bash`, `apply_patch` as standalone case-insensitive words.

### Dippin adapter

`extractAgentAttrs`: read `tool_access` from `graph.Attrs` into `cfg.ToolAccess`. Mirrors the existing `Backend`/`WorkingDir` pattern.

### `Params` bypass defense (runtime-side)

When `cfg.ToolAccess` is non-empty, the runtime MUST NOT honor any `Params` key that re-enables tools: `allowed_tools`, `disallowed_tools`, `tool_choice`, `permission_mode`. Implementation: in the codepaths that translate Params to runtime settings, check `cfg.ToolAccess` first and skip the Params override.

Test (runtime-side): construct a session with `ToolAccess: "none"` and `Params: {"allowed_tools": "Bash"}`. Assert zero tools registered.

This addresses round-1 finding #3 at the runtime layer rather than via a dippin lint contract. Authors who write the bypass get the correct (restrictive) behavior; DIP133 (existing hint-severity same-name-shadow lint) catches the redundant `params: { tool_access: ... }` case.

### Backend compatibility

`tool_access: none` MUST be honored on every backend dippin supports.

- **`backend: native`**: implemented as described above (empty registry + `tool_choice: none`).
- **`backend: claude-code`**: the runtime translates to claude-code's equivalent deny mechanism. The implementation plan documents the exact spelling, cites Claude Code CLI docs (URL + retrieval date, mirroring `CLAUDE.md`'s model-catalog pattern), and ships a runtime-side test that exercises the runtime behavior (not just the documented behavior).
- **`backend: acp`**: same requirement.

If the implementer cannot verify the deny-equivalent spelling for a given backend, that backend's session-creation refuses to start when `cfg.ToolAccess` is non-empty, with an error pointing to the relevant issue. Authors get a runtime error, not a silent no-op.

This is enforced in the runtime, not in dippin's validator. The dippin spec does not allocate a lint code for backend compatibility — the runtime error is the source of truth. If runtime-level rejection proves too late (e.g., long pipelines that take minutes to reach the bad agent), the implementation plan can add a dippin pre-flight lint at that time. v1 ships without.

## Tests

### Dippin

**Parser** (`parser/parse_tool_access_test.go`, new file): parse real `.dip` text via `parser.NewParser(src, "test").Parse()`. No hand-built IR.

| Case | Input | `cfg.ToolAccess` | Diagnostic |
|---|---|---|---|
| valid | `tool_access: none` | `"none"` | none |
| case variant | `tool_access: None` | `"None"` | none from parser; validator recognizes after normalization |
| invalid | `tool_access: foo` | `"foo"` | DIP139 from validator |
| quoted | `tool_access: "none"` | `"none"` | none |
| empty | `tool_access:` | `""` | none from parser; same as absent (acceptable v1 — author typed too little, gets default) |

**Validator** (`validator/lint_tool_access_test.go`, new file): DIP139 positive (invalid value) + negative (each recognized case).

**Integration** (`validator/lint_examples_test.go::TestLintExamples`): `examples/agent_tool_access.dip` must produce zero DIP139. Existing examples must not regress.

**Round-trip** (`migrate/roundtrip_test.go`): extend `compareAgentConfigs` to include `ToolAccess` (the field only — broader parity audit is a separate follow-up). New `TestRoundtripPreservesToolAccess` parses `.dip` with `tool_access: none`, round-trips through DOT export → `Migrate`, asserts the field survives.

### Runtime (cross-repo, specified here so dippin-side reviewer can verify the runtime PR meets the contract)

**Unit:**
- `builtInToolsForConfig` returns empty slice when `cfg.ToolAccess` is non-empty.
- `session_run` sets `ToolChoiceNone` and omits tool-naming system-prompt prefix when `cfg.ToolAccess` is non-empty.

**Integration:**
- End-to-end `.dip` → session → mocked LLM with `tool_access: none` produces a tool-free request.

**Red-team (the v0.28.2 vector):**
- Construct an agent with `cfg.ToolAccess = "none"`.
- Mock the LLM to emit a single response containing **multiple tool calls**: `[bash("rm -rf data/"), write("payload.py", "..."), bash("./payload.py")]`. This is the actual v0.28.2 shape — single-turn multi-tool-call smash-and-grab.
- Assert: zero tool-call executions, no file write, no shell execution. Response returned as plain text.

**Bypass attempts:**
- `tool_access: "none"` + `params: {"allowed_tools": "Bash"}` → zero tools registered.
- `tool_access: "noen"` (typo) → zero tools registered (fail-closed).
- `tool_access: "None"` (case variant) → zero tools registered.

**System-prompt audit:**
- Construct session with `cfg.ToolAccess = "none"`. Assemble the system prompt. Assert no occurrence of `read`, `write`, `edit`, `glob`, `grep_search`, `bash`, `apply_patch` as standalone case-insensitive words.

**Backend compatibility:**
- For each backend dippin supports (`native`, `claude-code`, `acp`): integration test with `tool_access: none` + a mocked LLM emitting tool calls. Assert zero executions OR session-creation refusal with a clear error.

## Example (`examples/agent_tool_access.dip`)

One file. Lint-clean. Mirrors `examples/tool_safety.dip` syntactic style. Demonstrates the field on a summarizer agent:

```dippin
workflow AgentToolAccess
  goal: "Demonstrate the tool_access agent-node safety primitive (issue #41)"
  start: Plan
  exit: ReportFinalStatus

  agent Plan
    model: claude-sonnet-4-6
    prompt: "Plan the work. Output a numbered task list."

  agent Implement
    model: claude-sonnet-4-6
    prompt: "Execute the plan."

  agent ReportFinalStatus
    model: claude-sonnet-4-6
    prompt: "Summarize what was implemented. Emit STATUS: success or STATUS: failure."
    tool_access: none      # bounded summarizer — no LLM tools available
    auto_status: true

  edges
    Plan -> Implement
    Implement -> ReportFinalStatus
```

The example demonstrates the load-bearing case: a closing summarizer that should never make file mutations. `Plan` and `Implement` retain full tools by omission.

## Release coordination

### PR sequence

1. **Runtime PR opens first.** Implements `SessionConfig.ToolAccess`, profile/session_run changes, Params bypass defense, system-prompt scrub, red-team test, backend-compat tests. Lands on the runtime's main branch but is not tagged.
2. **Dippin PR opens** (parallel, not blocked). Uses a commit-pinned go.mod during the PR window so integration tests run against the merged-but-untagged runtime SHA.
3. **Implementation plan task #0** files all follow-up issues (§ Non-goals 1–9) as numbered GitHub issues in `2389-research/dippin-lang`, records the numbers in this spec, and updates skill.md cross-references. DIP28's spec made an equivalent "file follow-up if needed" promise that was never kept — task #0 breaks the pattern.
4. **Spec text records the pinned runtime SHA** before dippin merges.
5. **Runtime tag** cut after dippin PR is approved.
6. **Dippin tag (v0.32.0)** cut immediately after the runtime tag.

### Version bumps

- Dippin: v0.31.0 → **v0.32.0**. Minor bump (new authoring-surface field).
- Runtime: minor bump (new `SessionConfig` field with runtime semantics).

### Doc updates

- **`CHANGELOG.md`** — `## [v0.32.0] — <date>` entry. Sections:
  - **Added**: `tool_access: none` agent-node field. DIP139 lint. `examples/agent_tool_access.dip`.
  - **Runtime-side** (linked to runtime tag): tool-registry filter, `tool_choice: none` for Anthropic, system-prompt scrub, Params bypass defense, red-team test.
- **`docs/validation.md`** — DIP139 entry with Trigger / Fix / Example.
- **`site/static/skill.md`** — new `tool_access:` field in the agent-node section. Document:
  - The single explicit value (`tool_access: none`); omission = full catalog.
  - The v0.28.2 threat model it bounds (single-agent multi-tool-call vector).
  - Non-goals: chain attack between agents, cross-node propagation, cascade. Explicit links to follow-up issues.
  - Tool-access scope ≠ tool-node safety (cross-reference DIP28's `tool_commands_allow`).
  - Runtime version requirement (requires an enforcing runtime).
- **Terror-squad doc closing note** — `docs/superpowers/research/2026-05-19-issue-41-terror-squad.md` Status line: "parked" → "shipped in v0.32.0 — see this spec."

## Design journey

This spec went through brainstorming + round-1 terror squad (parked v0.30.0) + rounds 2/3/4 of five-reviewer expert review. Each round absorbed valid findings and grew the spec, eventually reaching ~770 lines and four DIP codes with three versioned cross-repo contracts.

Round 4 surfaced that the absorbed complexity was itself the problem: an unexported function the spec required to be called across packages (impossible), a `forbidigo` lint that was sold as compile-time enforcement but is regex-based, an invented `# dippin:allow-DIP141` suppression syntax that doesn't exist in dippin, a four-state enum (`""`, `none`, `full`, invalid) whose author guidance was contradictory.

The simplification: drop the defaults cascade. Cascade requires an explicit `full` opt-out spelling. The opt-out spelling commits to a live-catalog semantic. The live-catalog semantic + cascade interaction creates cross-node leak questions (DIP141). Backend compatibility for the leak questions creates DIP142. Each ring expanded the spec by 100+ lines.

Per-node only — same shape as `goal_gate`, `max_turns`, `auto_status` — addresses the v0.28.2 vector for any agent the author annotates. Authors who want workflow-wide policy use a code-review convention; if incidents accumulate, cascade ships in v2 (follow-up #1) with proper design.

The full review transcripts are preserved at `/tmp/claude-1000/.../tasks/*.output` and were consulted while writing this spec. Findings that survived simplification (system-prompt scrub completeness, red-team test shape covering multi-tool-call response, Params bypass defense, coordinated runtime release, backend-compat tests, fail-closed runtime behavior, runtime version pin) are in the spec body. Findings that became moot under the simpler design are documented as non-goals.
