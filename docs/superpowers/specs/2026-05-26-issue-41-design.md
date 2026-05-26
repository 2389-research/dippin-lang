# v0.32.0 — `tool_access:` agent-node safety primitive (issue #41)

**Date:** 2026-05-26
**Closes:** [#41](https://github.com/2389-research/dippin-lang/issues/41)
**Status:** Design approved 2026-05-26, pending implementation
**Predecessor:** v0.30.0 brainstorming parked #41 in favor of #40; v0.31.0 shipped #45 + #49. Issue #41 returns here as a joint dippin + tracker release per the parking decision.
**Tracker dependency:** Pinned tracker PR/commit TBD at implementation start (see § Release coordination)
**Research:** [`docs/superpowers/research/2026-05-19-issue-41-terror-squad.md`](../research/2026-05-19-issue-41-terror-squad.md) (round 1 findings, parked) + round-2 reviewer transcripts (consolidated into this spec)

## Problem

`.dip` agent nodes on the native backend always ship the full file-mutation tool catalog (Read, Write, Edit, ApplyPatch, Glob, GrepSearch, Bash). The v0.28.2 runaway-agent incident — a 10-minute, 39k-output-token run where an agent implemented an entire Go project from a `SPEC.md` it found on disk during what should have been a passthrough "acknowledge" node — remains structurally easy to hit. Existing mitigations are inadequate:

- `max_turns: 1` does NOT bound damage within the single turn (engine executes all tool calls inside turn N=1 before checking the cap; `tracker/agent/session.go:213`).
- `allowed_tools` / `disallowed_tools` attrs work only with `backend: claude-code`; the native backend silently drops them.
- HARD CONSTRAINT prompt text is empirically unreliable (the v0.28.2 incident occurred with such text in place).

Dippin needs a language-level primitive that bounds an agent's tool catalog. The primitive must:

1. Bound the specific v0.28.2 attack shape (multi-tool-call in a single LLM response).
2. Be enforced at runtime by tracker, not just lint-validated by dippin — shipping lint-validated, runtime-no-op safety fields is materially worse than the status quo because authors ship believing they're sandboxed when they aren't.
3. Default to "safe by author intent, opt-in opt-out" via a workflow-level cascade — per-node-only requires the same author judgment that the v0.28.2 incident failed.

## Non-goals (v1)

These are explicitly out of scope and are tracked as follow-up issues:

1. **Middle tier (`read_filesystem` / `read_only`).** Round 1 finding #5 demonstrated middle-tier safety semantics are murky (A→disk→B→C laundering chain still works under `read_only`; `.env` exfiltration is unbounded). v1 ships `none` and `full` only. The middle tier returns when a richer surface ships, with explicit threat-model framing.
2. **`disallowed_tools` / `allowed_tools` list fields.** Round 1 finding #2 demonstrated typo footguns (lowercase silently no-ops against tracker's CamelCase catalog). Defer until a `KnownAgentTools` registry exists and case-normalization is locked.
3. **Chain-attack mitigation (`${ctx.last_response}` auto-injection).** `tool_access: none` bounds tools available to THIS agent's LLM call. It does NOT bound information flow between agents. A `tool_access: none` summarizer feeding a `tool_access: full` writer via `${ctx.last_response}` remains a viable chain attack. Tracked as a separate issue for `last_response_truncate:` or structural context-threading change.
4. **`ManagerLoopConfig.SubgraphRef` cross-workflow safety propagation.** A parent's restrictive `tool_access:` does not propagate into a child subgraph's agents (child has its own `defaults:`). v1 lints only intra-workflow edges; cross-workflow propagation is a separate design question requiring multi-workflow IR-traversal infrastructure that dippin's validator does not have today.
5. **Prompt-shape heuristic lint** (warn when `tool_access: none` is set on an agent whose prompt mentions `Read`, `Write`, `bash`, etc.). Heuristic — high false-positive risk. Filed as DIP143 candidate; not in v1.
6. **Tool nodes (shell command nodes) under `tool_access:`.** `tool_access:` bounds LLM-emitted tool calls. Tool nodes execute arbitrary shell unconditionally. DIP28's `tool_commands_allow` / `tool_denylist_add` handle that surface. Spec must make this distinction loud in skill.md and validation.md.

## Design decisions

These were locked in brainstorming with user approval and survived round-2 expert review. Each is justified once here; downstream sections implement them.

### D1. Single enum field; no companion list fields in v1

Source surface: `tool_access:` on agent nodes; same field accepted in `defaults:`. No `disallowed_tools`, no `allowed_tools`, no `tool_choice` in v1. Precedent reset: round 1 finding #4 said precedence rules between fields must be locked even if companion fields defer. With single-field surface there is no precedence to lock.

### D2. Field name `tool_access:` (source); IR named type `ir.ToolAccessPolicy`

Avoids the `tools`/`tool` collision flagged by three round-1 reviewers. Reserves the bare name `tools:` for a future map-shaped field. IR uses a **named string type** (mirrors `ir.NodeKind` precedent) with constants for valid values, not a bare `string`. Round-2 IR designer (C3): "A safety field should not be the weakest-typed field in the struct."

```go
// ir/ir.go
type ToolAccessPolicy string

const (
    ToolAccessDefault ToolAccessPolicy = ""     // unset; resolves via defaults cascade
    ToolAccessNone    ToolAccessPolicy = "none" // no LLM tools
    ToolAccessFull    ToolAccessPolicy = "full" // explicit full catalog (opt-out from cascade)
)
```

### D3. Enum values: `none` and `full` (plus default empty)

`tool_access: none` — no LLM tools.
`tool_access: full` — full catalog. Explicit opt-out from `defaults: tool_access: none` cascade.
Omitted — inherits from `defaults:` if set, else full.

**Why `full` ships as an explicit value (departure from brainstorming preference).** Round-2 language critic (C2): in the recommended `defaults: tool_access: none` cascade pattern, the only spelling for "this agent is intentionally trusted with tools" was to delete the field — invisible in source diffs. Reviewers see `defaults: tool_access: none` and assume every agent is sandboxed; in fact some are not, but the safety-bearing absence is unobservable. `tool_access: full` makes the opt-out PR-reviewable. Cost: one extra enum value. Benefit: visible source-diff signal that an author has consciously taken on the v0.28.2 risk surface for a specific node.

No middle tier (`read_only`, `read_filesystem`) in v1 — see non-goal #1.

### D4. Fail-closed coercion lives in ONE place: `ir.AgentConfig.EffectiveToolAccess()`

Round-2 security (C1), parser (P4), IR (C2), future-maintainer (F1) converged: "coerce at the IR-consumption layer" scatters fail-closed logic across 18+ direct consumers of `ir.AgentConfig`, drift inevitable.

Single canonical accessor:

```go
// ir/ir.go
func (c AgentConfig) EffectiveToolAccess() ToolAccessPolicy {
    switch normalizeToolAccess(string(c.ToolAccess)) {
    case "":
        return ToolAccessDefault
    case "none":
        return ToolAccessNone
    case "full":
        return ToolAccessFull
    default:
        return ToolAccessNone // fail closed
    }
}

func normalizeToolAccess(v string) string {
    return strings.ToLower(strings.TrimSpace(v))
}
```

**All runtime consumers go through `EffectiveToolAccess()`.** The raw `c.ToolAccess` field stays writable for the parser and readable for the validator (which lints authoring mistakes against the raw form via DIP139). No consumer except the validator reads `c.ToolAccess` directly.

Workflow-level resolution helper, also in `ir/`:

```go
func (w *Workflow) EffectiveToolAccess(n *Node) ToolAccessPolicy {
    if cfg, ok := n.Config.(AgentConfig); ok {
        if eff := cfg.EffectiveToolAccess(); eff != ToolAccessDefault {
            return eff
        }
        return w.Defaults.ToolAccess.normalize() // workflow default; same coercion rules
    }
    return ToolAccessDefault
}
```

The existing `Model`/`Provider` resolution duplication (5 separate sites — `cost/cost.go:159`, `optimize/optimize.go:62`, `validator/lint_model.go:231`, `simulate/events.go:81-88`, `lsp/hover.go:111-115`) is pre-existing tech debt; a follow-up issue tracks consolidating those into the same pattern. The safety field gets the right pattern from day one.

### D5. Parser stores verbatim; validator owns enum check

Parser stores `cfg.ToolAccess = ir.ToolAccessPolicy(val)` exactly as written. No `validToolAccess` map in parser (matches DIP127 / DIP130 pattern; round 1 finding #8). Validator owns enum validation via DIP139.

Defaults-block (`parser/parse_defaults.go`) uses the SAME normalization helper as the per-node case — both call into `applyToolAccessField(val)` to avoid case-handling drift between the two parse sites (round-2 parser P2).

### D6. Defaults cascade: per-node fully replaces

`defaults: tool_access: none` cascades to every agent in the workflow. Per-node `tool_access:` (any value, including `full`) fully replaces the default (no merge; matches `Model`/`Provider`, mirrors DIP28 `ToolCommandsAllow`). Round 1 finding #1: cascade is the only thing that catches "safe by default, opt out per node."

### D7. Validator: DIP139, DIP140, DIP141, DIP142 (no DIP133 extension)

Four new diagnostic codes. **DIP133 is NOT extended** for `tool_access` (round-2 IR I2): DIP133 is `SeverityHint`, which is wrong for a safety field. Same-name shadow (`params: { tool_access: full }` shadowing typed `tool_access:`) folds into DIP140 at warning severity. DIP142 (backend-compat) is described in D9.

| Code | Catches | Severity | File |
|---|---|---|---|
| DIP139 | `tool_access:` value not in `{"", "none", "full"}` (case-normalized) | error | `validator/lint_tool_access.go` |
| DIP140 | `params:` contains `{allowed_tools, disallowed_tools, tool_choice, permission_mode, backend, tool_access}` (case-normalized keys). Fires unconditionally — does NOT gate on `tool_access:` state. | warning | `validator/lint_tool_access.go` |
| DIP141 | Agent with effective `tool_access: none` has outgoing edge to one or more agent nodes whose effective `tool_access` is `full` or default | warning (advisory) | `validator/lint_tool_access.go` |
| DIP142 | `tool_access: none` combined with a backend whose deny-equivalent spelling is unverified (set fixed by implementation plan; see D9) | error | `validator/lint_tool_access.go` |

**DIP140 fires unconditionally** (round-2 LC C3, sec C2, future F11). The dangerous case — author writes `params: { permission_mode: bypassPermissions }` with no `tool_access:` set — must trip the lint. Gating only on `tool_access: none` would silence the dangerous case and warn the safe case (already-claimed-safety + redundant params).

DIP140's key list is a **versioned contract** with tracker (round-2 sec C2). The list lives in `validator/known_tracker_params.go` with a comment naming the tracker commit SHA last verified against. Tracker-side CI test (see § Tests) fails the build when an unlisted Params key gains runtime effect.

DIP141 stays advisory (warning) per brainstorming decision. The diagnostic Help text must explicitly reference the chain-attack interaction: "If the downstream agent receives `${ctx.last_response}` from this restricted agent, the chain-attack vector is active — see follow-up issue for `last_response_truncate:`."

### D8. `BranchConfig` and `ManagerLoopConfig`

`BranchConfig` gains a `ToolAccess ir.ToolAccessPolicy` field (round-2 IR C4). Matches the existing pattern of `BranchConfig` overriding `Model`/`Provider`/`Fidelity` per branch. Per-branch override fully replaces per-target effective tool_access (same semantic as defaults cascade per node). Cost: one IR field + one parse case + one round-trip case. Benefit: no breaking change when v2 adds richer surface; PR-reviewable per-branch safety override.

`ManagerLoopConfig` does NOT gain a `ToolAccess` field in v1 — the supervised child is a separate workflow with its own `defaults:` block. Cross-workflow propagation is non-goal #4.

### D9. Backend coupling: uniform — with one combination explicitly blocked pending verification

`tool_access: none` means "no LLM tools" regardless of backend. Three implementations:

- **`backend: native`**: register empty tool set + set `tool_choice: none` on the Anthropic request (translator already supports — `tracker/llm/anthropic/translate.go:182-184`). OpenAI equivalent.
- **`backend: claude-code`**: PENDING verification. Round-2 security (C3): Claude Code's `allowed_tools: []` semantics are not documented as deny-all; if empty array means "default permissive," the safety guarantee inverts on this backend.
- **`backend: acp`**: PENDING verification.

**v1 ships DIP142** as the backend-compatibility lint. **Error** severity. Fires when `tool_access: none` is combined with a backend whose deny-equivalent spelling has not been verified.

The set of unverified backends is fixed by the implementation plan based on documentation evidence (URL + retrieval date, mirroring `CLAUDE.md`'s model-catalog pricing-verification pattern). At spec authorship time, both `backend: claude-code` and `backend: acp` are unverified. The implementation plan documents:

- For each backend, the documented spelling for "no LLM tools" (e.g., `--allowedTools '' '' '' ''` for Claude Code CLI, or whatever the verified spelling is).
- The corresponding tracker translation code path.
- The verification date and source URL.

If both backends are verified by merge time, DIP142 ships with an empty unverified set (effectively disabled but reserved for future backends). If one is verified and one is not, DIP142 fires only on the unverified one. If neither is verified, DIP142 blocks both combinations.

Authors who need `claude-code` and hit DIP142 use `backend: claude-code`'s own `allowed_tools` attribute (silently dropped by native backend, but supported on claude-code) until verification lands.

### D10. Chain-attack vector: explicit non-goal

`tool_access: none` bounds tools available to THIS agent's LLM call. It does NOT bound information flow between agents (`${ctx.last_response}` auto-injection per `pipeline/transforms.go:55-82`). Spec, skill.md, and DIP141 Help text all call this out. Follow-up issue tracks the structural fix.

## Dippin-side design

### IR (`ir/ir.go`)

```go
// ToolAccessPolicy describes an agent's LLM-tool catalog policy.
// Used by tracker's session layer to bound the v0.28.2 runaway-agent vector.
// Values resolve via AgentConfig.EffectiveToolAccess() — never read the raw field
// from non-validator code.
type ToolAccessPolicy string

const (
    ToolAccessDefault ToolAccessPolicy = ""
    ToolAccessNone    ToolAccessPolicy = "none"
    ToolAccessFull    ToolAccessPolicy = "full"
)

// AgentConfig: add ToolAccess in the runtime/backend cluster (after Backend/WorkingDir,
// per round-2 IR I1 — semantically a runtime-safety concern, not model config).
type AgentConfig struct {
    // ... existing fields up to and including Backend, WorkingDir ...
    ToolAccess ToolAccessPolicy
    Params     map[string]string
}

// EffectiveToolAccess returns the runtime-effective policy.
// Invalid values fail closed to ToolAccessNone.
// Empty (= ToolAccessDefault) means "consult workflow defaults".
func (c AgentConfig) EffectiveToolAccess() ToolAccessPolicy {
    return normalizeToolAccessPolicy(string(c.ToolAccess))
}

// WorkflowDefaults: cluster with existing tool-safety fields per round-2 IR M3.
type WorkflowDefaults struct {
    // ... existing fields ...
    // Tool-safety cluster (DIP28 + this spec)
    ToolAccess        ToolAccessPolicy // none | full | "" (full)
    ToolCommandsAllow string
    ToolDenylistAdd   string
}

// BranchConfig: per-branch override matching the existing Model/Provider/Fidelity pattern.
type BranchConfig struct {
    Target     string
    Model      string
    Provider   string
    Fidelity   string
    ToolAccess ToolAccessPolicy
}

// Workflow-level resolver.
func (w *Workflow) EffectiveToolAccess(n *Node) ToolAccessPolicy {
    cfg, ok := n.Config.(AgentConfig)
    if !ok {
        return ToolAccessDefault
    }
    if eff := cfg.EffectiveToolAccess(); eff != ToolAccessDefault {
        return eff
    }
    return normalizeToolAccessPolicy(string(w.Defaults.ToolAccess))
}
```

`normalizeToolAccessPolicy(string) ToolAccessPolicy` is the shared helper used by `EffectiveToolAccess()`, the validator, and the formatter. It lowercases + trims + maps to a constant; unknown values map to `ToolAccessNone`.

### Parser (`parser/parse_nodes.go` and `parser/parse_defaults.go`)

Per-node case in `applyAgentRuntimeField` (cluster with `backend`, `working_dir`):

```go
case "tool_access":
    cfg.ToolAccess = ir.ToolAccessPolicy(val) // verbatim; validator lints
```

Defaults-block case in `applyDefaultToolField` (sibling to `tool_commands_allow`):

```go
case "tool_access":
    d.ToolAccess = ir.ToolAccessPolicy(val) // verbatim; validator lints
```

Both call sites store verbatim — no parser-side normalization. Per round-2 parser P2, both sites are co-located with comments cross-referencing the other to prevent drift when one is edited.

Branch-block case in `applyBranchField` (`parser/parse_nodes.go:733-742`):

```go
case "tool_access":
    branch.ToolAccess = ir.ToolAccessPolicy(val)
```

### Validator (`validator/lint_tool_access.go`, new file)

`validToolAccess = map[string]bool{"": true, "none": true, "full": true}`. Lookup uses `normalizeToolAccessPolicy` so `None`, `NONE`, `Full`, ` none ` all hit the right case.

**`lintToolAccessValues`** — DIP139, error. For every agent node + defaults block + branch config: if raw `ToolAccess` non-empty and `normalizeToolAccessPolicy(raw)` is not a recognized constant, emit:

```
node "X" has tool_access "Nono" which is not recognized
Help: Valid values: none, full. Omit the field to inherit from defaults (or full if no default set).
      Invalid values fall back to 'none' at runtime — fix the typo or remove the field.
```

The "falls back to none at runtime" sentence is load-bearing (round-2 parser P11). DIP139's diagnostic must surface this so authors who hit the lint don't treat it as informational.

**`lintParamsBypass`** — DIP140, warning. For every agent node: scan `cfg.Params` keys (case-normalized via `strings.ToLower(strings.TrimSpace(k))`). If any key matches the danger set, emit one diagnostic per match. **Unconditional** — does not gate on `tool_access:` state.

Danger key set lives in `validator/known_tracker_params.go`:

```go
// knownDangerousTrackerParams — Params keys that affect tracker's runtime tool
// registry or session config. Adding a typed dippin field for any of these
// requires removing it from this map AND a tracker-side change.
//
// Last verified against tracker commit: <SHA> (filled in at implementation time)
var knownDangerousTrackerParams = map[string]bool{
    "allowed_tools":     true,
    "disallowed_tools":  true,
    "tool_choice":       true,
    "permission_mode":   true,
    "backend":           true, // typed field exists; params override is bypass
    "tool_access":       true, // typed field exists; params override is bypass
    // NOTE: extend with verification against tracker's Params consumers when adding.
}
```

The tracker-side CI test (see § Tests) enumerates the keys tracker reads from `cfg.Params` and asserts they're a subset of `knownDangerousTrackerParams` ∪ a known-safe set. New tracker reads fail the build until dippin's list is updated.

**`lintCrossNodeToolAccess`** — DIP141, warning. For every agent node with `w.EffectiveToolAccess(n) == ToolAccessNone`: walk outgoing edges (and `BranchConfig.Target` references for parallel-fan-out nodes whose `Branches:` reference agents). For each target that resolves to an agent node with `w.EffectiveToolAccess(target) == ToolAccessFull` (or unset → full), emit:

```
agent "Summary" has tool_access: none; outgoing edge to agent "Writer" with tool_access: full
Help: The downstream agent has access to the full tool catalog. If "Summary"'s response
      feeds into "Writer"'s prompt via ${ctx.last_response}, the chain-attack vector is
      active — see the chain-attack follow-up issue (filed at merge time). Suppress this lint with `# dippin:allow-DIP141`
      on the edge if intentional.
```

Function structure to fit ≤5 cyclomatic / ≤7 cognitive per CLAUDE.md (round-2 parser P6):

- `lintCrossNodeToolAccess(res *Result, w *ir.Workflow)` — outer; iterates nodes
- `agentTargets(n *ir.Node, edges []*ir.Edge, w *ir.Workflow) []*ir.Node` — collect direct + branch-target agent nodes
- `checkRestrictiveLeak(res *Result, w *ir.Workflow, src, dst *ir.Node)` — emit one diag per leak
- `effectiveToolAccess` lives on `ir.Workflow` (D4), not in the validator

**`lintParamsSameNameShadow`** — folded into DIP140 (round-2 IR I2). When `params:` contains a key whose name matches a typed first-class agent field (specifically `tool_access`, but the danger set already covers it), DIP140 fires at warning severity. **DIP133 is NOT extended for `tool_access`** — its hint severity would silently degrade the safety claim.

### Explanations (`validator/explanations.go`)

DIP139, DIP140, DIP141, DIP142 entries with `Code`, `Summary`, `Trigger`, `Fix`, `Example` fields following the existing DIP127–138 pattern (`explanations.go:316-394`).

### Formatter (`formatter/format.go`)

Per-node `tool_access:` emitted by `writeAgentRuntimeFields` (`format.go:366-373`), conditional on `cfg.ToolAccess != ""`. Position: in the runtime/backend cluster (NOT the model cluster — round-2 IR I1 + I4):

```go
if cfg.ToolAccess != "" {
    wr.line("tool_access: %s", quoteValue(string(cfg.ToolAccess)))
}
```

Defaults-block `tool_access:` emitted by `writeDefaultsToolSafetyFields` (extend the existing `format.go:204-211` block to include `tool_access` first, then the existing `tool_commands_allow` and `tool_denylist_add`).

`BranchConfig.ToolAccess` emitted by the branch-block formatter analogously to `Model`/`Provider`/`Fidelity`.

Formatter emits **verbatim** invalid values too — so round-tripping `tool_access: nono` through `format` preserves the value and the validator still emits DIP139 after migrate (round-2 parser P7).

### DOT export (`export/dot.go`)

Per-node `tool_access` attribute on agent nodes via `applyAgentRuntimeAttrs` (`export/dot.go:280-302`). Defaults-block `tool_access` emitted in the defaults cluster.

`reservedGraphAttrs` (`export/dot.go:60`) gains `tool_access` so it can't collide with author `vars:`.

`BranchConfig.ToolAccess` emitted as a per-branch DOT attribute on the branch label/edge (matching the existing per-branch `model`/`provider`/`fidelity` pattern).

### Migrate (`migrate/migrate.go`)

`extractAgentAttrs` reads `tool_access` from DOT attrs into `cfg.ToolAccess`. Unconditional setter — invalid values flow through (validator catches; round-2 parser P7 confirms).

Defaults extraction reads `tool_access` analogously.

`BranchConfig` migration adds `ToolAccess` extraction.

**No new heuristic in `resolveStartExitKind`** — `tool_access` lives on agent shape's normal path (round-2 IR I5). The v0.31.0-shipped `hasToolConfigAttrs` is unaffected.

`compareAgentConfigs` (`migrate/parity.go:219-228`) — extend to compare `ToolAccess`. The existing comparator is incomplete (omits `Backend`, `WorkingDir`, `Params`); follow-up issue tracks the broader cleanup. For this spec, we extend with `ToolAccess` and add a separate `TestRoundtripPreservesToolAccess` that uses `reflect.DeepEqual` on the full config (round-2 IR I4) to catch any other quietly-broken parity.

### Pack-time validation (`cmd/dippin/cmd_pack.go`)

DIP139 is warning-severity, not error — so `dippin pack` does NOT refuse to pack a workflow with `tool_access: foo`. The workflow runs as if `tool_access: none` (fail-closed via `EffectiveToolAccess`); `dippin lint` reports the typo. This matches the existing pack-time / lint-time split (round-2 IR I6). Spec calls it out explicitly so reviewers don't "helpfully" add error-severity rejection.

### Dipx (`dipx/`)

**No changes to dipx for fail-closed semantics.** Coercion lives on `ir.AgentConfig.EffectiveToolAccess()` per D4 — dipx is the wrong layer (round-2 sec C1: dipx doesn't read `AgentConfig` fields). dipx's structural-loader job is unchanged.

## Tracker-side design

Tracker repo: `2389-research/tracker`. Files referenced below are based on the issue body's facts (`tracker/agent/session.go:213`, `tracker/agent/profile.go:9-31`, `tracker/agent/session_run.go:24-31, 120`) and the dippin adapter (`tracker/pipeline/dippin_adapter.go`).

### `tracker/agent/session.go`

`SessionConfig` gains `ToolAccess string` (or import dippin's `ir.ToolAccessPolicy` if a shared module is created — see Release coordination follow-up). Constructor reads from dippin adapter.

### `tracker/agent/profile.go`

`builtInToolsForConfig(cfg SessionConfig)`:

```go
if cfg.ToolAccess == "none" {
    return []Tool{} // no tools registered
}
// existing full-catalog return for "", "full"
```

### `tracker/agent/session_run.go`

When `cfg.ToolAccess == "none"`:

1. Set `request.ToolChoice = llm.ToolChoiceNone()` so Anthropic translator strips the `tools` array from the API request (`tracker/llm/anthropic/translate.go:182-184`).
2. **System-prompt audit pass** (round-2 sec I6, not just the one line from the issue body). Every prompt component that names a tool must be omitted when `ToolAccess == "none"`. This includes:
   - The "File tool arguments (read, write, edit, glob, grep_search) MUST use paths relative to the working directory." prefix (`session_run.go:24-31`).
   - Any per-tool short descriptions in the Anthropic `tools` array (auto-stripped when array is empty, but verify).
   - Reasoning/role text that enumerates tool names.
   - Any other prefix added by `agent/profile.go` or downstream prompt-assembly code.

   Tracker-side test asserts: when `ToolAccess == "none"`, the assembled system prompt contains no occurrence of `read`, `write`, `edit`, `glob`, `grep_search`, `bash`, `apply_patch` as standalone words.

### `tracker/pipeline/dippin_adapter.go`

`extractAgentAttrs`: read `tool_access` from `graph.Attrs` (set by dippin's DOT exporter) into `cfg.ToolAccess`. Resolve workflow-defaults cascade by reading `w.Defaults.ToolAccess`.

### Backend-specific translation

- **`backend: native`** — empty registry + `tool_choice: none`. Implemented per above.
- **`backend: claude-code`** — `tracker/pipeline/handlers/codergen.go`. Translation pending verification (D9). Block the combination via DIP139-adjacent validator error in v1 if not verified by merge.
- **`backend: acp`** — analogous to claude-code. Same blocking treatment if unverified.

The implementation plan must include the documentation citation (Claude Code CLI docs URL + retrieval date) for the deny-equivalent spelling, or the v1 spec ships with the combination blocked.

## Tests

### Dippin parser (`parser/parse_tool_access_test.go`, new file)

Parse real `.dip` text via `parser.NewParser(src, "test").Parse()`. No hand-built IR.

| Case | Input | Expected `cfg.ToolAccess` | Expected diagnostic |
|---|---|---|---|
| valid none | `tool_access: none` | `"none"` | none |
| valid full | `tool_access: full` | `"full"` | none |
| case variant | `tool_access: None` | `"None"` (verbatim) | none from parser; DIP139 from validator after normalization → recognized |
| invalid | `tool_access: foo` | `"foo"` | none from parser; DIP139 from validator |
| quoted | `tool_access: "none"` | `"none"` | none |
| quoted + comment | `tool_access: "none" # secured` | `"none"` | none |
| comment-no-space | `tool_access: none# rationale` | `"none# rationale"` | DIP139 (this is a pre-existing parser edge case per round-2 P3; spec acknowledges) |
| defaults-block | `defaults\n  tool_access: none` | `w.Defaults.ToolAccess = "none"` | none |
| branch-config | parallel block with `branch: X\n  tool_access: none` | `branch.ToolAccess = "none"` | none |

### Validator (`validator/lint_tool_access_test.go`, new file)

Each lint code: positive case (lint fires with expected message and position) + negative case (lint does NOT fire).

- DIP139 positive: invalid value; negative: each of `""`, `"none"`, `"full"`, `"None"`, `"NONE"`.
- DIP140 positive: each danger key in `params:` (case variants); negative: safe keys (`temperature`, `max_tokens`, etc.).
- DIP141 positive: `tool_access: none` source agent with edge to `tool_access: full` (or default) agent; negative: chain of `tool_access: none` agents; negative: `tool_access: full` source agent.
- DIP141 with cascade: `defaults: tool_access: none` + per-node `tool_access: full` correctly inherits opposite from default and lint fires.
- DIP141 across `BranchConfig.Target`: parent has `tool_access: none`, parallel block branches to mixed `tool_access` agents.
- DIP142 positive: `tool_access: none` + unverified backend (set per implementation plan); negative: `tool_access: none` + verified backend; negative: `tool_access: full` + unverified backend.

### Integration (`validator/lint_examples_test.go::TestLintExamples`)

Runs all `examples/*.dip` through real parse → lint. The new `examples/agent_tool_access.dip` must produce zero DIP139/140/141 (lint-clean — round 1 finding "examples must lint-clean"). Existing examples must not regress.

### Round-trip (`migrate/roundtrip_test.go`)

Extend `compareAgentConfigs` to assert `ToolAccess`. Add new `TestRoundtripPreservesToolAccess`:

- Parses `.dip` with `defaults: tool_access: none`, per-node `tool_access: full`, branch with `tool_access: none`.
- Round-trips through DOT export → `Migrate`.
- Asserts: every `ToolAccess` survives exactly (including case-preserved invalid values — DIP139 still fires after round-trip).

Use `reflect.DeepEqual` on the full `AgentConfig` to catch any other quietly-broken parity.

### Tracker side (cross-repo)

Specified in tracker PR; spec calls out the required shape so the dippin-side reviewer can verify the tracker PR meets the contract.

**Unit:**
- `builtInToolsForConfig` returns empty slice for `ToolAccess == "none"`.
- Session-run sets `ToolChoiceNone` and skips the tool-naming system-prompt prefix.

**Integration:**
- End-to-end `.dip` → session → mocked LLM with `tool_access: none` produces tool-free request.

**Red-team (the v0.28.2 vector):**
- Constructs an agent with `ToolAccess == "none"`.
- Mocks the LLM to emit a single response containing **multiple tool calls**: `[bash("rm -rf data/"), write("payload.py", "..."), bash("./payload.py")]` (round-2 sec I5 — the actual v0.28.2 shape, NOT a single tool call).
- Asserts: zero tool-call executions, no file write, no shell execution. Response returned as plain text.
- Additional red-team scenarios:
  1. Bypass via `params: { allowed_tools: Bash }` with `tool_access: none` — assert dippin lint catches (DIP140) AND tracker runtime rejects the bypass attempt.
  2. Bypass via `params: { backend: claude-code }` — assert lint + runtime both honor the typed `tool_access: none` over the params-shadowed `backend`.
  3. Invalid `tool_access: full_catalog_please` — assert fail-closed coercion lands as `none` (via `EffectiveToolAccess()`).

**System-prompt audit:**
- Construct session with `ToolAccess == "none"`. Assemble the system prompt. Assert no occurrence of `read`, `write`, `edit`, `glob`, `grep_search`, `bash`, `apply_patch` as standalone case-insensitive words.

**Cross-repo contract test (round-2 IR I3):**
- Tracker CI test imports dippin's `dipx`, opens a `.dipx` bundle with `tool_access: none`, asserts the resulting `SessionConfig.ToolAccess == "none"`. Round-trip preservation across the dippin→dipx→tracker pipeline.
- Tracker CI test enumerates keys read from `cfg.Params` (via codepath audit or runtime assertion). Asserts every key is in dippin's `knownDangerousTrackerParams` ∪ a known-safe set. Adding a new tracker Params read without updating dippin breaks the build.

## Examples (`examples/agent_tool_access.dip`)

Lint-clean. Mirrors the issue body's `ReportFinalStatus` framing:

```
workflow ReportFinalStatusDemo:
  start: Plan

  defaults:
    tool_access: none      # workflow-wide safe-by-default

  agent Plan:
    prompt: "Plan the work; output a numbered task list."
    tool_access: full      # explicit opt-out — Plan needs tools to investigate
    edges:
      - on: success → Implement

  agent Implement:
    prompt: "Execute the plan."
    tool_access: full
    edges:
      - on: success → ReportFinalStatus

  agent ReportFinalStatus:
    prompt: "Summarize what was implemented. Emit STATUS: line."
    # tool_access inherited from defaults → none. Bounded summarizer.
    auto_status: true
```

The example demonstrates: defaults cascade, explicit per-node `full` opt-out, default-inherited `none` for the summarizer. It does NOT demonstrate the cross-node chain attack (which is the non-goal); the comment in the example file calls that out and links to the follow-up issue.

## Release coordination

### PR sequence

1. **Tracker PR opens first.** Implements `SessionConfig.ToolAccess`, profile/session_run changes, system-prompt audit, red-team test, cross-repo contract test. Lands on tracker `main` but is **not tagged**.
2. **Dippin PR opens after tracker PR is open** (parallel, not blocked). Uses `replace github.com/2389-research/tracker => ../tracker` (or a commit-pinned go.mod) during the PR window so integration tests run against the merged-but-untagged tracker SHA.
3. **Spec pins the tracker SHA.** The implementation plan starts by recording the target tracker PR number; spec text updates with the SHA before dippin merges.
4. **Tracker tag** (e.g., `vX.Y.Z`) cut after dippin PR is approved.
5. **Dippin tag (v0.32.0)** cut immediately after tracker tag, with go.mod bumped to the new tracker version.

**Hard mechanism (round-2 future F5):** dippin's CI gains a check that fails if the go.mod tracker version is older than the SHA pinned in the spec. Without this, the joint-release discipline relies on PR-reviewer attention.

### Version bumps

- Dippin: v0.31.0 → **v0.32.0**. Minor bump (new authoring-surface field).
- Tracker: minor bump (new `SessionConfig` field with runtime semantics).

### Doc updates

- **`CHANGELOG.md`** — `## [v0.32.0] — <date>` entry. Lead sentence: "New `tool_access:` field on agent nodes bounds the LLM tool catalog at runtime. Joint release with tracker `<tag>`." Sections: Added (the field + lint codes + examples), Tracker-side (system-prompt scrub, tool-registry filter, red-team test).
- **`docs/validation.md`** — DIP139, DIP140, DIP141 documented with examples. DIP133 documentation amended to note `tool_access` is intentionally excluded from `agentFirstClassFields` (DIP140 covers it instead).
- **`site/static/skill.md`** — new `tool_access:` field in agent-node section. Document:
  - The cascade pattern (`defaults: tool_access: none` + explicit per-node `tool_access: full`).
  - Tool-access scope ≠ tool-node safety (cross-reference DIP28's `tool_commands_allow`).
  - The chain-attack non-goal + follow-up issue link.
  - Tracker version requirement (`requires tracker >= vX.Y`).
- **`cmd/dippin/generated-spec.md`** — regenerates via existing hook; no manual edits.
- **Terror-squad doc closing note** — update `docs/superpowers/research/2026-05-19-issue-41-terror-squad.md` Status line from "parked" to "shipped in v0.32.0 — see docs/superpowers/specs/2026-05-26-issue-41-design.md."

## Findings matrix (round 1 + round 2)

### Round 1 (terror-squad) — every critical addressed or knowingly deferred

| # | Finding | Resolution |
|---|---|---|
| 1 | Per-node-only misses bug class | D6 defaults cascade + D3 explicit `full` opt-out |
| 2 | DIP141 reserved typo footgun | N/A — no `disallowed_tools` in v1 (D1); when it returns, `KnownAgentTools` registry + case-normalization required |
| 3 | `Params` parallel attack surface | D7 DIP140 unconditional + versioned tracker contract |
| 4 | Precedence rules undefined | N/A — single field in v1 (D1) |
| 5 | `read_only` doesn't bound v0.28.2 vector | D3 no middle tier; non-goal #3 chain attack with follow-up |
| 6 | `Tools` name collision | D2 `tool_access` |
| 7 | Fail-open on parse error | D4 fail-closed via `EffectiveToolAccess()` |
| 8 | DIP139-in-parser violates pattern | D5 validator owns enum check |
| 9 | `DisallowedTools []string` vs DIP28 string | N/A — no list field in v1 (D1) |
| 10 | Cross-node coverage holes | D7 DIP141 + D8 `BranchConfig.ToolAccess` (intra-workflow); cross-workflow non-goal #4 |

### Round 2 critical findings — resolutions in spec

| Ref | Finding | Resolution |
|---|---|---|
| LC-C1 | Cross-node lint misses `BranchConfig`/`SubgraphRef` | DIP141 walks `BranchConfig.Target`; `SubgraphRef` deferred (non-goal #4) with explicit doc |
| LC-C2 | Opt-out invisible in source diffs | D3 ships `tool_access: full` as explicit opt-out |
| LC-C3 | DIP140 gating inverted | D7 DIP140 unconditional |
| Sec-C1 | Fail-closed has no canonical home | D4 single `EffectiveToolAccess()` accessor on `ir.AgentConfig` |
| Sec-C2 | DIP140 danger-list incomplete | D7 versioned contract + tracker-side CI test |
| Sec-C3 | Claude Code `allowed_tools: []` unverified | D9 block `claude-code`+`none` combo until verified-with-citation |
| Parser-P1 | Case normalization missing | D4 `normalizeToolAccessPolicy` helper used by validator + runtime |
| Parser-P2 | Defaults-block parses separately | D5 shared parse helper + cross-reference comments |
| Parser-P4 | Fail-closed split across consumers | D4 single accessor |
| IR-C1 | Resolution helper duplicated | D4 placed on `ir.Workflow` (the right layer for safety primitives) |
| IR-C2 | "IR-consumption layer" undefined | D4 collapses to one accessor; all consumers go through it |
| IR-C3 | `string` field type concedes correctness | D2 named type `ir.ToolAccessPolicy` |
| IR-C4 | BranchConfig override asymmetry | D8 ship `BranchConfig.ToolAccess` now |
| Future-F1 | L4 invents new infrastructure | D4 — yes, intentionally. Safety field gets canonical accessor that DIP28 lacks. Follow-up issue tracks retroactive DIP28 cleanup. |
| Future-F2 | Name lock-in unchanged | Accepted — `tool_access` is a policy name; v2 growth to middle tier is expected and documented |
| Future-F4 | Backend uniformity asserted not enforced | D9 + cross-repo contract test |
| Future-F8 | DIP141 false-negatives | D8 `BranchConfig` traversal; `SubgraphRef` non-goal #4 |

### Important findings absorbed without separate resolution

- LC-I1 / Future-F2 — field-name policy framing accepted
- Sec-I1 — DIP140 unconditional firing also resolves the "intentional opt-out with custom tool selection" case
- Sec-I2 — DIP141 stays warning; Help text references chain attack
- Sec-I4 — DIP142 prompt-heuristic deferred (non-goal #5)
- Sec-I5 — red-team test shape spec'd above
- Sec-I6 — system-prompt audit spec'd above
- Parser-P3 — quote+comment edge case acknowledged in test matrix; pre-existing parser behavior, not regressed
- Parser-P6 — DIP141 helper decomposition pre-planned to fit complexity caps
- Parser-P7 — formatter/migrate emit verbatim spec'd above
- Parser-P11 — DIP139 Help text "falls back to none at runtime" spec'd above
- IR-I1 — `ToolAccess` placed in runtime cluster (after `Backend`/`WorkingDir`)
- IR-I2 — DIP133 NOT extended; same-name shadow folds into DIP140
- IR-I3 — cross-repo contract test spec'd above
- IR-I4 — formatter position corrected; round-trip `reflect.DeepEqual` upgrade
- Future-F5 — go.mod tracker-version CI guard spec'd above
- Future-F6 — DIP139/140/141 contiguous allocation OK; no block reservation in v1
- Future-F7 — `lint_tool_access.go` owns "tool-access concerns"; v2 `disallowed_tools` extends same file
- Future-F9 — skill.md gains `requires tracker >= vX.Y` directive
- Future-F10 — `examples/agent_tool_access.dip` evolves in place when middle tier lands; no `_v2` suffix sprawl

## Follow-up issues to file at merge time

1. **Chain-attack mitigation.** `${ctx.last_response}` auto-injection allows A→B→C laundering even with `tool_access: none` on B. Candidate fix: `last_response_truncate:` field or structural context-threading change. (References non-goal #3.)
2. **Cross-workflow safety propagation.** `ManagerLoopConfig.SubgraphRef` does not inherit parent's `tool_access`. Requires dipx-aware validator pass to traverse child workflows. (References non-goal #4.)
3. **DIP143 prompt-shape heuristic.** Warn when `tool_access: none` on an agent whose prompt mentions tool-shaped verbs. Heuristic; high false-positive risk. (References non-goal #5.)
4. **Model/Provider resolution helper consolidation.** Five duplicated resolution sites in `cost`, `optimize`, `validator/lint_model`, `simulate/events`, `lsp/hover`. Move to `ir.Workflow.EffectiveAgentConfig` once the v0.32 `EffectiveToolAccess` pattern is validated.
5. **Shared `dippin-contract` Go module.** Round-2 IR I3 — type-share between dippin and tracker via a small shared module (currently coordinated by string-typed `graph.Attrs`). Long-term answer to the joint-release contract problem.
6. **DIP28 retroactive `EffectiveToolCommandsAllow` accessor.** DIP28's `ToolCommandsAllow` lacks the canonical-accessor treatment this spec gives `ToolAccess`. Round-2 future F1 noted no precedent exists; this spec creates the precedent — retroactive cleanup of DIP28 is the natural follow-up.
7. **Middle-tier surface (`read_filesystem`).** v2 surface that supports a read-only catalog with explicit threat-model framing (terror finding #5).
8. **Companion list fields (`disallowed_tools`, `allowed_tools`).** v2 surface that ships after `KnownAgentTools` registry and case-normalization are locked.

## Complexity budgets

All new functions ≤ 5 cyclomatic, ≤ 7 cognitive per CLAUDE.md.

| Function | Estimated cyclomatic | Notes |
|---|---|---|
| `normalizeToolAccessPolicy` | 2 | one switch |
| `AgentConfig.EffectiveToolAccess` | 1 | delegates |
| `Workflow.EffectiveToolAccess` | 3 | type assert + cfg-eff + defaults |
| `lintToolAccessValues` | 4 | iterate nodes + defaults + branches; per-iteration check |
| `lintParamsBypass` | 4 | iterate nodes + per-node iterate params + danger-set check |
| `lintCrossNodeToolAccess` | 3 | outer iterate; delegates to helpers |
| `agentTargets` | 4 | direct edges + branch targets |
| `checkRestrictiveLeak` | 3 | predicate + emit |
| `lintBackendToolAccessCompat` | 3 | iterate nodes + per-node check backend against unverified set |

If any function trips `just complexity`, extract helpers per CLAUDE.md (no `//nolint`).
