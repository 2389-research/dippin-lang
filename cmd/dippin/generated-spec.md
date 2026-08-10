# Dippin Language Specification

Complete reference for AI agents generating `.dip` workflow files. This document is the canonical, self-contained spec for the dippin DSL and CLI toolchain.

Install: `go install github.com/2389-research/dippin-lang/cmd/dippin@latest`

Generate from template: `dippin new minimal`, `dippin new parallel`, `dippin new conditional`, `dippin new review-loop`, `dippin new human-gate`

## Grammar (simplified BNF)

```
workflow <Name>
  goal: "<text>"
  [requires: <dep1>, <dep2>, ...]
  start: <NodeID>
  exit: <NodeID>

  [inputs
    <name>: <text|number|bool|enum|file|secret>
      [required: true] [default: <value>] [prompt: "<text>"] [description: "<text>"]
      [options: <a>, <b>] [pattern: "<regex>"] [min: <n>] [max: <n>] [max_length: <int>] [multiline: true]
    ...]

  [defaults
    model: <string>
    provider: <string>
    max_total_tokens: <int>
    max_cost_cents: <int>
    max_wall_time: <duration>
    stall_timeout: <duration>
    on_failure: <NodeID>
    tool_commands_allow: "<glob>,<glob>"
    tool_denylist_add: "<glob>,<glob>"
    ...]

  [vars
    <key>: <value>
    ...]

  <kind> <NodeID>
    <field>: <value>
    <multiline_field>:
      <indented content>

  parallel <ID> -> <Target1>, <Target2>[, ...]
  fan_in <ID> <- <Source1>, <Source2>[, ...]

  edges
    <From> -> <To> [on <token> | when <condition>] [label: <text>] [choice: <key>] [weight: <int>] [loop] [override: true]
    [else -> <NodeID>]   # success-side default for any node with no matching guard / unconditional edge; at most one per block
```

`else` is reserved as the first token of an edges-block line, so a node cannot be used as an edge *source* under the ID `else` (the other contextual keywords `on`/`when`/`loop` are only special after `->` and remain usable as node IDs).

---

## Node Kinds

| Kind | Required Fields | Optional Fields |
|------|----------------|-----------------|
| `agent` | `prompt` (or `prompt_file`) | `model`, `provider`, `backend`, `working_dir`, `tool_access` (`none` disables LLM tools; DIP139 warns on unknown), `auto_status`, `goal_gate`, `reasoning_effort`, `fidelity`, `max_turns`, `prompt_file` (external file for `prompt`; mutually exclusive with `prompt:`), `system_prompt`, `system_prompt_file` (external file for `system_prompt`; mutually exclusive with `system_prompt:`), `writable_paths` (CSV glob list bounding where this agent's tools may write, e.g. `workspace/**, .ai/**`; absent = unbounded; present-but-empty is a parse error; a malformed or runtime-unrecognized value fails closed at the runtime = deny-all) |
| `human` | `mode` (freeform\|choice\|interview\|yes_no) | `default`, `timeout` (duration, e.g. 5m), `timeout_action` (string: fail\|default) |
| `tool` | `command` (or `command_file`) | `timeout` (e.g. 30s, 5m), `outputs` (CSV), `marker_grep` (regex), `route_required` (bool), `output_limit` (bytes), `command_file` (path to external script, relative to .dip dir) |
| `parallel` | `-> Target1, Target2` (inline) | — |
| `fan_in` | `<- Source1, Source2` (inline) | — |
| `subgraph` | `ref` | `params` |

All kinds also accept: `label`, `reads`, `writes`, `retry_policy`, `max_retries`, `base_delay`. **v1-only (rejected under `dip 2`):** `retry_target`, `fallback_target` — in `dip 2` use a `loop` edge and an `on fail` edge instead (`dippin fmt --migrate` converts).

---

## Edge Conditions

```
when <variable> <op> <value>
when <expr> and <expr>
when <expr> or <expr>
when not <expr>
on <token>                       # sugar: equality vs the source node's outcome channel
```

**Comparison operators:** `=`, `==`, `!=`, `contains`, `not contains`, `startswith`, `endswith`, `in` (all string comparison, no numeric ops)

**`on <token>` shorthand:** desugars to `when <channel> = <token>`, where the channel is the source node's natural outcome channel — `ctx.outcome` for agent nodes, `ctx.tool_marker` for tool nodes with `marker_grep`. IR-identical to the equivalent `when`; `dippin fmt` rewrites eligible `when` edges to `on`. The value must be a single bare identifier (`[A-Za-z0-9][A-Za-z0-9_-]*`); quoted, multi-token, or any other values require an explicit `when <channel> = ...`. Source nodes with no outcome channel must use `when`: human gates (which route on the choice/label, not `ctx.outcome`), `conditional` nodes, and tools without `marker_grep`.

**Variables:** Always namespace-qualified: `ctx.outcome`, `ctx.status`, `graph.goal`, `params.*` (subgraph params), `inputs.*` (declared workflow inputs — a closed namespace; an undeclared reference is DIP156)

---

## Common LLM Mistakes

| # | Mistake | Fix |
|---|---------|-----|
| 1 | Missing `start:` or `exit:` field | Every workflow needs both. They reference node IDs declared below. |
| 2 | Edge references undeclared node | Every node in an edge must be declared as `agent`, `human`, `tool`, etc. |
| 3 | `parallel` targets without matching `fan_in` sources | `parallel P -> A, B` requires `fan_in J <- A, B` with the same set. |
| 4 | Bare variable names in conditions | Use `ctx.outcome`, not `outcome`. All variables need a namespace prefix. |
| 5 | Agent node with empty prompt | Every `agent` node should have a `prompt:` field with content (except start/exit lifecycle nodes). |
| 6 | Missing tool timeout | Add `timeout: 60s` (or appropriate duration) to every `tool` node. |
| 7 | Exhaustive conditions flagged | `ctx.outcome = success` + `ctx.outcome = fail` is exhaustive — DIP101/DIP102 are auto-suppressed. No need to add a fallback edge. |
| 8 | Verbose output sharing stdout with routing marker | When a tool's stdout drives routing, redirect verbose output to a sibling file and `printf` only the marker. Otherwise large output (test logs, stack traces) can crowd out the marker under runtime stdout caps. See `nodes.md` → Tool Nodes → Markers and Verbose Output. |
| 9 | Hand-parsing tool stdout for routing | Use `marker_grep: "<regex>"` (and optionally `route_required: true`) instead of regexing `ctx.tool_stdout` in edge conditions. Populates `ctx.tool_marker` directly — typed routing is more reliable than substring matching on raw stdout. |
| 10 | DIP101/DIP102 flagged on marker-routed tool node | If the tool already declares `marker_grep:`, the validator treats it as a safe routing source and suppresses both warnings. If you're still seeing them, the source node isn't a `tool`, or `marker_grep` is empty. |
| 11 | Boolean field rejected as invalid | Boolean fields (`goal_gate`, `auto_status`, `cache_tools`, `route_required`) accept `true/false`, `1/0`, `yes/no`, `on/off`, case-insensitive. Anything else is a parse error — pre-v0.29 silently coerced unknown values to `false`. |

---

## Exhaustive Conditions

When outgoing edges from a node cover all possible values, DIP101 and DIP102 warnings are automatically suppressed. Known exhaustive sets:

- `ctx.outcome`: `{success, fail}` or `{success, failure}`
- `outcome`: `{success, fail}` or `{success, failure}`

Tool nodes that declare `marker_grep:` are also treated as exhaustive (typed routing via `ctx.tool_marker`).

This means the common pattern below is valid with zero warnings:

```
Gate -> Fix when ctx.outcome = fail
Gate -> Next when ctx.outcome = success
```

---

## Example: Conditional Routing

```
workflow ReviewPipeline
  goal: "Review code and route by outcome"
  start: Analyze
  exit: Done

  agent Analyze
    auto_status: true
    prompt:
      Analyze the code changes.
      Set STATUS: success if approved, STATUS: fail if changes needed.

  agent Approve
    prompt:
      Finalize the approved changes.

  agent RequestChanges
    prompt:
      Describe what changes are needed.

  agent Done
    prompt:
      Summarize the review outcome.

  edges
    Analyze -> Approve  when ctx.outcome = success
    Analyze -> RequestChanges  when ctx.outcome = fail
    Analyze -> Done
    Approve -> Done
    RequestChanges -> Done
```

---

## Identifiers & Reserved Words

**Identifiers:** `[a-zA-Z0-9][a-zA-Z0-9_\-./]*` — letters, digits, underscore, dash, dot, slash.

**Contextual keywords** (not reserved — usable as node IDs): `workflow`, `agent`, `human`, `tool`, `subgraph`, `parallel`, `fan_in`, `edges`, `defaults`, `inputs`, `vars`, `when`, `on`, `and`, `or`, `not`, `true`, `false`, `restart`, `loop`, `override`, `label`, `weight`.

**Position-reserved keyword:** `else` is the one exception — it is reserved *only* as the first token of an `edges`-block line (where it introduces the section default), so it cannot be an edge *source* node ID there. Everywhere else `else` is an ordinary identifier and may be used as a node ID.

---

## Validation with `dippin check`

Use `dippin check` in tool-calling loops to validate generated `.dip` files. It runs parse + validate + lint in one shot and outputs JSON to stdout:

```bash
dippin check my_workflow.dip
```

```json
{"valid":true,"errors":0,"warnings":0,"diagnostics":[],"suggested_actions":[]}
```

```json
{"valid":false,"errors":1,"warnings":2,"diagnostics":[{"code":"DIP003","severity":"error","message":"unknown node reference \"Nope\" in edge","line":19,"fix":""}],"suggested_actions":[]}
```

Use `valid` to decide whether to retry generation. Use `diagnostics` to feed error details back to the LLM for correction. Use `suggested_actions` for actionable fixes when available.

---

## Bundles (`.dipx`)

A `.dipx` is a deterministic ZIP that packages a `.dip` entry plus every transitively-reachable subgraph as one integrity-verified artifact. Every analysis command (`validate`, `lint`, `doctor`, `check`, `parse`, `cost`, `coverage`, `simulate`, `optimize`, `unused`, `graph`, `diff`, `explain`, `export-dot`) accepts either `.dip` or `.dipx` as input.

- **Build a bundle**: `dippin pack pipeline.dip` → `pipeline.dipx`
- **Inspect**: `dippin inspect pipeline.dipx` (prints manifest, sha256 identity, file list)
- **Extract**: `dippin unpack pipeline.dipx -o ./out` (atomic via staging dir + rename)

Workflow: author and lint as `.dip`; package with `dippin pack` for distribution to the runtime. `dippin check pipeline.dipx` validates the bundled entry workflow exactly as if it were on disk. Bundle commands return distinct exit codes (`0` ok, `1` user error, `2` integrity error, `3` I/O error, `4` cancelled) so tooling can disambiguate failures that the analysis-command `0/1/2` ladder collapses.

---

## Diagnostic Code Summary

67 diagnostic codes across two categories:

- **DIP001–DIP010** (errors): start/exit missing, unknown refs, unreachable nodes, cycles, duplicates, parallel/fan_in mismatch, unparseable edge conditions
- **DIP101–DIP159** (warnings): conditional reachability, missing defaults, overlapping conditions, unbounded retries, undefined variables, unknown models, empty prompts, missing timeouts, invalid policy/fidelity/reasoning_effort, stylesheet refs, namespace prefixes, condition type checking, structured output validation, manager_loop checks, tool-access safety, writable-paths safety, subgraph tool_access boundary, agent failure route, negative budget defaults, cross-file subgraph tool_access, restricted→tool-bearing info-flow (chain-attack), negative last_response_truncate, ambiguous routing (multiple unconditional edges), human-gate choice key (label routes without explicit choice), edge weight (unused by routing), marker coverage (marker_grep enumerates a marker no edge routes), redundant parallel/fan_in edge (edges-block re-declaration of an inline fork; the inline list is authoritative), prompt-cascade opt-out no-op (prompt_prefix/suffix: none with no defaults cascade), unknown input type (DIP155), reference to an undeclared input in a prompt or edge condition (DIP156), `${inputs.x}` inside a tool `command:` which never interpolates (DIP157), invalid or inapplicable input constraint — enum default ∉ options, min > max, bad pattern regex, or a constraint on a type that lacks it (DIP158), declared-but-unreferenced input / dead input (DIP159). DIP155–DIP158 are error-severity — they cause `dippin lint` and `dippin check` to exit non-zero (DIP159 is a warning).

---

## File Structure (strict order)

```
workflow <Name>
  goal: "<description>"
  requires: <dep1>, <dep2>   # optional; environmental deps surfaced to runtimes
  start: <NodeID>
  exit: <NodeID>

  defaults
    model: claude-sonnet-4-6
    provider: anthropic

  inputs
    idea: text
      required: true
      prompt: "What do you want built?"

  <node declarations>

  edges
    <edge declarations>
```

Sections in canonical order: **header** (workflow name, goal, optional `requires`, start, exit) → **inputs** (optional) → **defaults** (optional) → **vars** (optional) → **nodes** (any order) → **stylesheet** (optional) → **edges** (optional). `inputs`, `defaults`, `vars`, and `edges` are bare keywords — no colon after them. `requires:` is a comma-separated list of environmental dependencies (e.g. `git, docker, jq`); semantics live in downstream consumers and unknown entries are accepted without a parser diagnostic.

`inputs` declares the workflow's callee-side signature: entries are `name: type` (types: `text`, `number`, `bool`, `enum`, `file`, `secret`) with an optional indented block of attributes (`required`, `prompt`, `description`, `default`, `options`, `pattern`, `min`, `max`, `max_length`, `multiline`). Declaration order is significant and never reordered by the formatter. Reference a declared input as `${inputs.name}` in a prompt — never inside a tool `command:`, which never interpolates it (DIP157); an undeclared reference is DIP156, and an unrecognized type is DIP155.

Indentation: 2 spaces. Comments: `#` line comments (literal inside multiline blocks).

## Node Types

### agent — LLM call

```
  agent Review
    prompt:
      Analyze the code and produce a structured review.
      Rate quality from 1-10.
    model: claude-sonnet-4-6
    provider: anthropic
    auto_status: true
    goal_gate: true
    retry_policy: standard
    max_retries: 3
```

| Field | Type | Notes |
|-------|------|-------|
| `prompt` | multiline | Required (DIP110 if empty, start/exit exempt) |
| `prompt_file` | string | Path (relative to `.dip` source directory) to an external file whose contents become the agent's `prompt`. Mutually exclusive with `prompt:`. See "Prompt File Directives" below. |
| `system_prompt` | multiline | System message |
| `system_prompt_file` | string | Path (relative to `.dip` source directory) to an external file whose contents become the agent's `system_prompt`. Mutually exclusive with `system_prompt:`. See "Prompt File Directives" below. |
| `model` | string | Must be valid model ID (DIP108) |
| `provider` | string | anthropic, openai, google, deepseek, xai, mistral, cohere |
| `backend` | string | Per-node backend override (e.g., `native`, `claude-code`, `acp`) |
| `working_dir` | string | Per-node working directory override for isolated execution. |
| `tool_access` | string | LLM tool-catalog gate. Only one explicit value: `none` (no tools). Omitted = full catalog. Invalid values are fail-closed at runtime and warned by DIP139. An enforcing runtime is required. See "Agent Tool Access" below. |
| `writable_paths` | CSV (globs) | Comma-separated glob list bounding where this agent's tools may write (e.g. `workspace/**, .ai/sprints/**`). Absent = unbounded. An enforcing runtime is required. See "Writable Paths" below. |
| `max_turns` | int | Max conversation turns |
| `cmd_timeout` | duration | e.g. `30s`, `5m` |
| `auto_status` | bool | Parses `STATUS: success/fail` → `ctx.outcome` |
| `goal_gate` | bool | Pipeline fails if gate fails. Add a failure route — an `on fail` edge (dip 2), or `retry_target`/`fallback_target` (v1). See DIP115 |
| `response_format` | string | `json_object` or `json_schema` (DIP130) |
| `response_schema` | multiline JSON | Must be valid JSON (DIP132). Requires `response_format: json_schema` (DIP131) |
| `reasoning_effort` | string | `none`, `minimal`, `low`, `medium`, `high`, `xhigh`, `max` (DIP119) |
| `fidelity` | string | `full`, `summary:high`, `summary:medium`, `summary:low`, `compact`, `truncate` (DIP114) |
| `cache_tools` | bool | Cache tool results |
| `compaction` | string | Context compaction strategy |
| `compaction_threshold` | float | 0.0-1.0 (DIP116) |
| `params` | key: value | Custom parameters. Keys must not shadow field names (DIP133) |
| `reads` | CSV | Context keys read (advisory) |
| `writes` | CSV | Context keys written (advisory) |

**Agent Tool Access (`tool_access:`)** — *added v0.32.0; an enforcing runtime is required.*

A node-level gate on the LLM tool catalog. One explicit value:

- `tool_access: none` — The runtime returns an empty tool registry to the model, strips the `tools` array from the request (e.g., Anthropic `tool_choice: none`), and scrubs tool-naming text from the system prompt. `Params` keys (`allowed_tools`, `disallowed_tools`, `tool_choice`, `permission_mode`) are ignored when the gate is set — Params cannot reopen it.
- *omitted* — Full catalog (current behavior, unchanged).

Invalid values fall back to no-tools at runtime (fail-closed) and are flagged by [DIP139](https://2389-research.github.io/dippin-lang/validation.html#dip139). Bad spelling reduces capability, never expands it.

**Threat model bounded:** the v0.28.2 single-agent, multi-tool-call vector — an LLM emitting multiple tool calls in a single response to bypass per-call gating. Set `tool_access: none` on summarizer / reporter / status-only agents that should never execute tools.

**Scope is per node, not per pipeline.** `tool_access: none` constrains the executor of *this one node* — it does not taint or restrict what downstream nodes do with this node's output. A bounded summarizer feeding a tool-capable agent is a normal, intended pattern: the restriction is fully intact (this node still cannot call tools), it simply does not extend across edges. Pipeline-level information-flow control is a separate, runtime-layer concern (see [#56](https://github.com/2389-research/dippin-lang/issues/56)).

**Non-goals (deferred):** cross-node propagation / cascade ([#53](https://github.com/2389-research/dippin-lang/issues/53)). Chain attacks between agents ([#56](https://github.com/2389-research/dippin-lang/issues/56)) are *partially* addressed by [DIP147](https://2389-research.github.io/dippin-lang/validation.html#dip147) (Hint): a `tool_access: none` agent that declares a context key in `writes:` flowing into a downstream tool-bearing agent's `reads:`. DIP147 keys off dippin's existing declared-IO flow model (the same `reads:`/`writes:` analysis as DIP107/DIP112), **not** a runtime-specific data-flow model — which is why the broader cross-node *edge* lint ([#57](https://github.com/2389-research/dippin-lang/issues/57), the bare `none → full` / `${ctx.last_response}` auto-injection edge) remains rejected: it would have to assume the runtime's auto-injection semantics. The `${ctx.last_response}` vector and a `last_response_truncate:` mitigation remain #56 follow-ups; the v1 field bounds a single-agent vector.

**Scope vs. tool-node safety:** `tool_access` gates *LLM-driven* tool calls on agent nodes. It is unrelated to `tool` nodes (shell commands authored directly in `.dip`), whose allowlist/denylist is controlled by the v0.28.x defaults `tool_commands_allow` and `tool_denylist_add`.

`tool_access` may also be set per-branch on a block-form `parallel` node; an omitted branch value inherits the target agent's setting.

**Subgraph boundary:** `tool_access` does not cross a `subgraph_ref` / `ref` file boundary — a referenced child `.dip` (`manager_loop` or `subgraph` node) is governed entirely by its own file. When a workflow declares `tool_access` and also references a subgraph, [DIP143](https://2389-research.github.io/dippin-lang/validation.html#dip143) (Hint) reminds you to give the child's agents their own `tool_access`. This is distinct from the in-file flow lints (DIP147, and the rejected #57 edge warning): it concerns the cross-*file* boundary. Native `dippin lint` now resolves the child across that boundary: [DIP146](https://2389-research.github.io/dippin-lang/validation.html#dip146) (Hint) fires when a resolved child restricts no agent's `tool_access` while a workflow on the path does, superseding DIP143 for boundaries it can resolve ([#89](https://github.com/2389-research/dippin-lang/issues/89)). DIP143 remains the filesystem-free advisory (e.g. the wasm playground) and the fallback when the child can't be resolved.

**Writable Paths (`writable_paths:`)** — *added v0.35.0; an enforcing runtime is required.*

A node-level glob list bounding where the agent's tools may write. Shape: comma-separated globs (e.g. `workspace/**, .ai/sprints/**`).

- `writable_paths: workspace/**, .ai/sprints/**` — the runtime confines all file mutations (Write, Edit, ApplyPatch, Bash, and any process Bash spawns) to paths matching these globs, resolved against an **immutable session root**. `working_dir` and `Params` keys cannot relocate the anchor.
- *omitted* — unbounded writes (current behavior, unchanged).
- **Fail-closed:** A present-but-empty `writable_paths:` is rejected by `dippin validate`/`pack` (parse error — list at least one glob or omit the field). A `writable_paths` that is malformed or **unrecognized by a runtime that does not enforce this field** → the runtime must deny all writes or refuse to start. Never falls through to unbounded. **A runtime that does not enforce `writable_paths` must refuse to start rather than run unbounded — this is a safety requirement, not a suggestion.**

**Enforcement scope (native backend only):** `writable_paths` is enforced on the `native` backend. On `claude-code` and `acp`, session creation **refuses to start** when `writable_paths` is set — fail-closed, never a silent no-op.

**Residual escape classes (out of scope):** `writable_paths` bounds *where writes land*; it does **not** bound network (e.g. `curl`, `cargo fetch`), reads / read-based exfiltration, or *content* within an allowed path (an agent with `writable_paths: workspace/**` can still poison `workspace/Cargo.toml`). Chain laundering (writing an allowed file that a downstream unbounded agent reads) is tracked in [#56](https://github.com/2389-research/dippin-lang/issues/56).

**Non-goals (deferred):** cross-node propagation / defaults cascade ([#53](https://github.com/2389-research/dippin-lang/issues/53)), tool-name allowlists ([#55](https://github.com/2389-research/dippin-lang/issues/55)), chain-attack mitigation ([#56](https://github.com/2389-research/dippin-lang/issues/56)).

**Lint:** DIP141 fires when `writable_paths` is set alongside `tool_access: none` on the same object (dead config — no tools to bound). DIP142 fires on unsafe entries: absolute paths, `~`, Windows drive letters, `..` escapes, or brace-expansion fragments (`*.{md` from `*.{md,yaml}` being comma-split). Use workspace-relative globs (e.g. `.ai/sprints/**`).

`writable_paths` may also be set per-branch on a block-form `parallel` node; an omitted branch value **inherits the target agent's** setting — it never resets to unbounded.

### human — user decision gate

```
  human Approve
    mode: choice
```

| Field | Type | Notes |
|-------|------|-------|
| `mode` | string | **Required.** `choice`, `freeform`, `interview`, or `yes_no` (DIP127) |
| `default` | string | Default choice (meaningless in interview mode — DIP128) |
| `prompt` | multiline | Prompt text |
| `questions_key` | string | Context key for interview questions |
| `answers_key` | string | Context key for interview answers |
| `timeout` | duration | e.g. `5m`, `1h`. How long to wait for human response. |
| `timeout_action` | string | `fail` or `default`. Action on timeout (default: `fail`). |
| `reads` | CSV | Context keys read |
| `writes` | CSV | Context keys written |

**Modes:**
- `choice`: Outgoing edge labels become buttons. Human selects one.
- `freeform`: Open text input → `ctx.human_response`
- `interview`: Structured Q&A from upstream agent output. Don't combine with choice-style edges (DIP129).
- `yes_no`: Binary Y/N prompt — two outgoing edges labeled `[Y]` and `[N]`.

### tool — shell command

```
  tool RunTests
    command:
      npm test -- --coverage
    timeout: 60s
    outputs: pass, fail
```

| Field | Type | Notes |
|-------|------|-------|
| `command` | multiline | Shell command. Supports pipes, here-docs, case/esac. Required unless `command_file` is set. |
| `command_file` | string | Path (relative to the `.dip` source directory) to an external script whose contents replace inline `command:`. Mutually exclusive with `command`. See "Tool Command File" below. |
| `timeout` | duration | **Required** (DIP111). e.g. `30s`, `5m` |
| `outputs` | CSV | Possible stdout values for condition checks |
| `marker_grep` | string | Regex matched against stdout; sets `ctx.tool_marker`. The runtime validates and applies the regex. |
| `route_required` | bool | When true, fails the node if the command emits no routing signal recognized by the runtime (the runtime defines the routing-signal format). |
| `output_limit` | int | Per-node stdout byte cap (non-negative integer); 0 (or omitted) uses the engine default. |
| `reads` | CSV | Context keys read |
| `writes` | CSV | Context keys written |

Do NOT use `${ctx.*}` in commands — they expand to empty at parse time (DIP124). Output is captured as `ctx.tool_stdout` and `ctx.tool_stderr`.

**Tool Command File (`command_file:`)** — *added v0.33.0.*

Reference an external file for a tool node's command instead of inlining a heredoc:

```dip
tool Setup
  command_file: scripts/setup.sh
```

Path resolution: relative to the `.dip` source directory. Absolute paths rejected. Symlinks rejected. Parent-tree escape (`../../etc/passwd`) rejected. 4 MiB size cap.

Mutually exclusive with `command:` — specifying both is a parse error.

Loading: CLI entry points (`dippin lint`, `dippin pack`, `dippin validate`, `dippin doctor`) load the file contents into the IR after parse. The LSP and the playground skip loading; they show the path unresolved. The runtime reads `.dipx` bundles where content is already inlined, so it sees no difference from inline `command:`.

Non-goals (deferred): configurable size cap ([#66](https://github.com/2389-research/dippin-lang/issues/66)), full-chain symlink resolution ([#67](https://github.com/2389-research/dippin-lang/issues/67)), glob expansion ([#68](https://github.com/2389-research/dippin-lang/issues/68)), DOT round-trip preservation of the directive form ([#69](https://github.com/2389-research/dippin-lang/issues/69)), graceful LSP/WASM not-loaded signal ([#70](https://github.com/2389-research/dippin-lang/issues/70)). See issue [#52](https://github.com/2389-research/dippin-lang/issues/52).

**Prompt File Directives (`prompt_file:` and `system_prompt_file:`)** — *added v0.34.0.*

Reference external prompt files from agent nodes:

```dip
agent Reviewer
  model: claude-sonnet-4-6
  system_prompt_file: prompts/persona.md
  prompt_file: prompts/task.md
```

Convention: keep prompt files in a `prompts/` directory alongside your `.dip`.

Path resolution and security are identical to `command_file:` above:
- Paths resolved relative to the `.dip` source directory
- Absolute paths rejected
- Parent-tree escapes (`..`) rejected
- Symlinks rejected
- 4 MiB size cap

The two slots are independent — an agent may use any combination of inline `prompt:`, `prompt_file:`, inline `system_prompt:`, `system_prompt_file:`. Only same-slot conflicts (`prompt:` + `prompt_file:`, or `system_prompt:` + `system_prompt_file:`) are parser-time errors. Cross-slot mixes (e.g. `prompt:` + `system_prompt_file:`) are fine.

**Pack-time loading:** `dippin pack` inlines the prompt content into the bundled `.dip` so the `.dipx` is self-contained. The runtime reads inline prompts from the bundle; no separate file lookup at runtime.

**Shared prompt fragments (#175)** — single-source boilerplate shared across agents. In the `defaults` block, `prompt_prefix:`/`prompt_suffix:` (inline) or `prompt_prefix_file:`/`prompt_suffix_file:` (fragment file) cascade to **every agent**; a per-agent `prompt_include: <file>` appends an extra fragment. The effective prompt is composed at resolve time as `prefix → body → include → suffix` (suffix always last — satisfies "final line must be …"). An agent opts out with `prompt_suffix: none` / `prompt_prefix: none` (`DIP154` hints on an opt-out with no matching cascade). Fragment files use the same security envelope as `prompt_file`; `pack` inlines the composed prompt, or ships the fragment files under `--no-inline`.

### parallel / fan_in — concurrent execution

```
  parallel FanOut -> WorkerA, WorkerB, WorkerC
  fan_in Merge <- WorkerA, WorkerB, WorkerC
```

Both inline (`parallel P -> A, B`) and block form (`parallel P` with `branch:` lines) are supported; block form additionally allows per-branch `model` / `provider` / `fidelity` / `tool_access` / `writable_paths` / `last_response_truncate` overrides (an omitted per-branch value inherits the target agent's setting). Every `parallel` must have a matching `fan_in` with identical target/source sets (DIP007) — this applies to both forms. Wire edges from each target to the `fan_in` node in the `edges` block. All targets execute concurrently with independent context copies.

### subgraph — embed another workflow

```
  subgraph CodeReview
    ref: phases/code_review.dip
    params:
      repo: myproject
      branch: main
    reads: analysis
    writes: review_result
```

| Field | Type | Notes |
|-------|------|-------|
| `ref` | string | Path to .dip file (DIP126 if missing) |
| `params` | key: value | Passed to child via `${params.key}` |
| `reads` | CSV | Context keys read |
| `writes` | CSV | Context keys written |

### manager_loop — supervised child pipeline

Spawns a child `.dip` pipeline, polls it on a cadence, and can steer it by injecting context. Maps to `stack.manager_loop` in the runtime; DOT shape `house`. Full reference: [docs/nodes.md](https://github.com/2389-research/dippin-lang/blob/main/docs/nodes.md).

```dip
  manager_loop QualityGate
    label: "Quality Gate Supervisor"
    subgraph_ref: quality_loop.dip
    poll_interval: 30s
    max_cycles: 12
    stop_condition: stack.child.outcome = success
    steer_condition: stack.child.cycles = 5
    steer_context:
      hint: halfway_through
      priority: high
```

| Field | Type | Notes |
|-------|------|-------|
| `subgraph_ref` | string | **Required.** Path to child .dip file (DIP135 if missing/not found) |
| `poll_interval` | duration | Poll cadence (e.g. `30s`). `0` = event-driven |
| `max_cycles` | int | Max poll cycles. `0` = unbounded → DIP137 |
| `stop_condition` | condition | Over `stack.child.*`; when true the loop exits |
| `steer_condition` | condition | When true, inject `steer_context` into child |
| `steer_context` | map[string]string | Inline `k=v, k=v` or block form. No commas in inline values |

Runtime state: `stack.child.cycles`, `stack.child.outcome`, `stack.child.status`. Lint: DIP135 (bad ref), DIP136 (invalid field), DIP137 (unbounded).

## Common Fields (all block nodes)

| Field | Notes |
|-------|-------|
| `label` | Display name (defaults to node ID) |
| `class` | CSS class names (reserved) |
| `retry_policy` | `standard`, `aggressive`, `patient`, `linear`, `none` (DIP113 if invalid) |
| `max_retries` | Max retry attempts |
| `base_delay` | Override base delay, e.g. `500ms`, `2s` |
| `retry_target` | **v1 only** (rejected under `dip 2`) — node to jump to on retry; use a `loop` edge in dip 2 |
| `fallback_target` | **v1 only** (rejected under `dip 2`) — node if retries exhausted; use an `on fail` edge in dip 2 |

**Every node must have at least one field.** An empty node body causes a parse error.

## Edges

```
  edges
    Start -> Analyze
    Analyze -> Decide
    Decide -> Merge when ctx.outcome = success
    Decide -> Revise when ctx.outcome = fail
    Revise -> Analyze loop
```

| Attribute | Syntax | Notes |
|-----------|--------|-------|
| condition | `when <expr>` | Guard expression |
| outcome shorthand | `on <token>` | Sugar for `when ctx.outcome = <token>` (agent) or `when ctx.tool_marker = <token>` (tool + `marker_grep`); `fmt` rewrites eligible `when` to `on`. `<token>` must be a single bare identifier (`[A-Za-z0-9][A-Za-z0-9_-]*`) — quoted or other values need `when`. Not for human gates (route on choice/label) or marker-less tools — use `when` |
| label | `label: <text>` | Display text / human choice button |
| choice | `choice: <key>` | Human-gate routing key; carried, not interpreted — `choice:` is preferred when present, and `label:` remains the fallback routing key when `choice:` is absent (DIP150) |
| weight | `weight: <int>` | Soft-deprecated (DIP151) — parsed but ignored by routing; removal slated for dip 2 |
| override | `override: true` | Carried, not interpreted by the parser |
| else default | `else -> <NodeID>` | Section-level success-side default route; at most one per `edges` block; no source node and no attributes (#157) |
| loop | `loop` | Bare keyword marking a back-edge; **required on back-edges** to avoid DIP005 (unconditional cycle). Legacy `restart: true` still parses; `fmt` rewrites it to `loop` |

### Conditions

Variables must have namespace prefix: `ctx.`, `params.`, or `graph.` (DIP120 if missing).

| Operator | Example |
|----------|---------|
| `=` or `==` | `ctx.outcome = success` |
| `!=` | `ctx.outcome != fail` |
| `contains` | `ctx.response contains error` |
| `not contains` | `ctx.response not contains error` |
| `startswith` | `ctx.type startswith urgent` |
| `endswith` | `ctx.name endswith _review` |
| `in` | `ctx.tier in gold,silver,bronze` |
| `and` / `or` | `ctx.outcome = success and ctx.score = high` |
| `not` | `not ctx.flagged = true` |

Parentheses control precedence. Operator priority: `not` > `and` > `or`.

**Quoted values** *(lossless since v0.49.0)*: condition values may be double-quoted — `when ctx.msg = "hello world"`. Inside double quotes, escaped `\"` and `\\` are preserved losslessly, and operator- or comment-like text (`||`, `#`) is literal — only a real trailing `#` comment is stripped. An unterminated double quote is a parse error, rejected before validation (reported at the opening quote — not a DIP010, which is for conditions that tokenize but fail to parse). Quoting is required when the value is the reserved bare keyword `loop`: `when ctx.x = "loop"` (unquoted `loop` is taken as the back-edge flag).

**Exhaustive detection:** The linter auto-detects exhaustive condition pairs (`success`/`fail`, complementary `contains`/`not contains`). Using `success`/`fail` as condition values suppresses DIP101/DIP102 warnings.

## Multiline Blocks

Fields `prompt:`, `system_prompt:`, `command:`, `response_schema:` support indented content:

```
  agent MyAgent
    prompt:
      First line sets the indentation baseline.
      All subsequent lines are de-indented by that amount.

      Empty lines are preserved.
      # This is literal content, not a comment.
```

## Defaults Block

```
  defaults
    model: claude-sonnet-4-6
    provider: anthropic
    retry_policy: standard
    max_retries: 2
    fidelity: medium
    max_restarts: 3
    cache_tools: true
    compaction: auto
    max_total_tokens: 500000
    max_cost_cents: 1000
    max_wall_time: 30m
    on_failure: Escalate
    stall_timeout: 5m
    tool_commands_allow: "git *,make *"
    tool_denylist_add: "rm -rf /,dd *"
```

All defaults are inherited by nodes unless overridden at the node level.

| Default Field | Type | Notes |
|---------------|------|-------|
| `max_total_tokens` | int | Budget cap on total tokens consumed. |
| `max_cost_cents` | int | Budget cap in cents (e.g. 1000 = $10.00). |
| `max_wall_time` | duration | Maximum wall-clock time for the workflow (e.g. `30m`, `2h`). |
| `on_failure` | string | Graph-level default failure route — node to jump to when an agent has no explicit failure edge, no fallback_target, and no bounded retry. |
| `stall_timeout` | duration | Wall-clock span with no forward progress before the run aborts/routes through on_failure (e.g. `5m`, `90s`). 0/unset = disabled. |
| `tool_commands_allow` | string | Glob allowlist for tool-node shell commands (comma-separated; optional). |
| `tool_denylist_add` | string | Glob patterns appended to the runtime's default denylist (comma-separated; optional). |

## CLI Reference

**Global flag:** `--format text|json` — must come **before** the subcommand (e.g. `dippin --format json check file.dip`).

Use `dippin help` (not `--help`) to see all commands.

### Authoring

| Command | Purpose |
|---------|---------|
| `dippin parse <file>` | Output IR as JSON |
| `dippin validate <file>` | Structural checks only (DIP001-DIP010) |
| `dippin lint <file>` | Full validation + semantic warnings (DIP001–DIP154) |
| `dippin check <file>` | All-in-one. JSON output by default — **use this for automated workflows** |
| `dippin fmt <file>` | Print canonical format to stdout |
| `dippin fmt --check <file>` | Exit 1 if not formatted |
| `dippin fmt --write <file>` | Rewrite file in place |
| `dippin fmt --migrate <file>` | Convert a v1 file to `dip 2` (edges own destinations). Combines with `--check`/`--write`. Exit 3 when the migration flags cases needing author review |
| `dippin new <template>` | Generate from template: `minimal`, `parallel`, `conditional`, `review-loop`, `human-gate` |
| `dippin spec` | Print the full embedded language specification |

### Export

| Command | Purpose |
|---------|---------|
| `dippin export-dot <file>` | Export to Graphviz DOT. `--rankdir` to set layout direction, `--prompts` to include prompts |
| `dippin export-dip <file>` | Export flattened .dip (resolves subgraph refs) |
| `dippin migrate <file.dot>` | Convert DOT to .dip. `--output <file>` to write instead of stdout |
| `dippin validate-migration <old.dot> <new.dip>` | Verify migration parity |

### Analysis

| Command | Purpose |
|---------|---------|
| `dippin simulate <file>` | Dry-run (JSONL events). `--scenario key=val` to inject context. `--all-paths` for exhaustive. `--interactive` to prompt at human nodes |
| `dippin cost <file>` | Estimate execution cost by model/provider. Requires model/provider on nodes or in defaults |
| `dippin coverage <file>` | Edge coverage and reachability |
| `dippin doctor <file>` | Health report card (grade A-F) |
| `dippin optimize <file>` | Suggest cheaper model substitutions |
| `dippin diff <file1> <file2>` | Semantic diff between two workflows |
| `dippin feedback <file>` | Compare predicted vs actual costs |
| `dippin explain <code>` | Explain a diagnostic code (e.g. `dippin explain DIP005`) |
| `dippin unused <file>` | Detect dead-branch nodes and wasted cost |
| `dippin graph <file>` | Render ASCII DAG of the workflow. `--compact` for a denser layout |
| `dippin test <file>` | Run `.test.json` scenario tests. `--verbose --coverage` for details |
| `dippin watch <file>` | Watch for changes, re-validate on save |
| `dippin lsp` | Start LSP server on stdio (for editor integration) |

### Bundles

| Command | Purpose |
|---------|---------|
| `dippin pack <entry.dip>` | Build a deterministic `.dipx` bundle from a `.dip` entry. `-o <out>` (default: `<entry>.dipx`; `-` for stdout). `--dry-run` validates without writing. |
| `dippin unpack <bundle.dipx>` | Atomic extract. `-o <destdir>` (default: bundle name without `.dipx`). `--force` overwrites with rollback-safe backup-aside swap. |
| `dippin inspect <bundle.dipx>` | Print manifest, identity (sha256 over manifest bytes), and per-file checksums. `--format text\|json`. |

## Bundle Workflow (`.dipx`)

A `.dipx` is a deterministic, content-addressed ZIP packaging a `.dip` entry plus every transitively-reachable subgraph ref. **Every analysis command (validate, lint, doctor, check, parse, cost, coverage, simulate, optimize, unused, graph, diff, explain, export-dot) accepts a `.dipx` argument** — the bundle is opened via `dipx.Load`, hash-verified, and the entry workflow is fed to the analyzer just like a `.dip` would be.

**Recommended workflow:** author and lint as `.dip`, package with `dippin pack` for distribution to the runtime.

**Exit codes for bundle commands** (`pack` / `unpack` / `inspect`) are `0` ok, `1` user error, `2` integrity error, `3` I/O error, `4` cancelled — distinct from the analysis-command standard `0` / `1` / `2` set.

## Validation Workflow

The primary loop for authoring .dip files:

```
1. Write or edit the .dip file
2. Run: dippin check --format json <file>
3. Parse the JSON output:
   {
     "valid": true/false,
     "errors": 0,
     "warnings": 0,
     "diagnostics": [
       {"code": "DIP111", "severity": "warning", "message": "...", "line": 35, "fix": "..."}
     ],
     "suggested_actions": ["add timeout to tool node"]
   }
4. Fix each diagnostic using the code reference below
5. Repeat until valid: true with 0 errors and 0 warnings
```

## Diagnostic Codes

### Structural Errors (must fix)

| Code | Issue | Fix |
|------|-------|-----|
| DIP001 | Start node missing | Add `start: <NodeID>` to workflow header |
| DIP002 | Exit node missing | Add `exit: <NodeID>` to workflow header |
| DIP003 | Unknown node in edge | Check spelling of node IDs in edges block |
| DIP004 | Unreachable node | Add an edge path from start to the node |
| DIP005 | Unconditional cycle | Add `loop` to the back-edge |
| DIP006 | Exit has outgoing edges | Remove edges from exit node or change exit to a different node |
| DIP007 | Parallel/fan_in mismatch | Add matching `fan_in` node with identical target set, and wire edges from each target to the fan_in node |
| DIP008 | Duplicate node ID | Rename one of the duplicate nodes |
| DIP009 | Duplicate edge | Remove the duplicate. Uniqueness is determined by `(source, target)` pair — two edges to the same target with different labels are still duplicates |
| DIP010 | Edge condition cannot be parsed | Fix the `when` expression syntax — check operators, quoting (a bare reserved word like `loop` as a value is taken as a flag — quote it), and namespace prefixes. (An *unterminated* `"` is a separate parse-time error, not DIP010.) |

### Semantic Warnings (should fix)

| Code | Issue | Fix |
|------|-------|-----|
| DIP101 | Node only reachable via conditionals | Add unconditional fallback edge or make conditions exhaustive (`success`/`fail`) |
| DIP102 | No default edge from routing node | Add unconditional edge or exhaustive conditions |
| DIP103 | Overlapping conditions | Disambiguate condition expressions |
| DIP104 | Unbounded retry loop | Add `max_retries`, an `on fail` edge (dip 2), or `fallback_target` (v1) |
| DIP105 | No success path start→exit | Ensure at least one unconditional path exists |
| DIP106 | Undefined variable in prompt | Check `${var}` references |
| DIP107 | Written key never read downstream | Remove unused `writes` or add consumer node |
| DIP108 | Unknown model/provider | Use valid model ID (see provider docs) |
| DIP109 | Namespace collision in imports | Rename conflicting subgraph namespaces |
| DIP110 | Empty agent prompt | Add `prompt:` content (start/exit nodes exempt) |
| DIP111 | Tool without timeout | Add `timeout: 30s` (or appropriate duration) |
| DIP112 | Reads key not written upstream | Add `writes:` to producing node |
| DIP113 | Invalid retry policy | Use: `standard`, `aggressive`, `patient`, `linear`, `none` |
| DIP114 | Invalid fidelity | Use: `full`, `summary:high`, `summary:medium`, `summary:low`, `compact`, `truncate` |
| DIP115 | Goal gate without recovery | Add an `on fail` edge (dip 2), or `retry_target`/`fallback_target` (v1) |
| DIP116 | Invalid compaction threshold | Use float 0.0-1.0 |
| DIP117 | Stylesheet class references undefined class | Fix class name in stylesheet block |
| DIP118 | Stylesheet ID references unknown node | Fix node ID in stylesheet block |
| DIP119 | Invalid reasoning_effort | Use: `none`, `minimal`, `low`, `medium`, `high`, `xhigh`, `max` |
| DIP120 | Condition var missing namespace | Prefix with `ctx.`, `params.`, or `graph.` |
| DIP123 | Shell syntax error in command | Fix the shell command (checked via `bash -n`) |
| DIP124 | `${ctx.*}` in tool command | Remove — runtime variables expand to empty at parse time |
| DIP125 | Command binary not on PATH | Install the binary or fix the command |
| DIP126 | Subgraph ref file missing | Check `ref:` path |
| DIP127 | Invalid human mode | Use: `choice`, `freeform`, `interview`, `yes_no` |
| DIP130 | Invalid response_format | Use: `json_object`, `json_schema` |
| DIP131 | Schema/format mismatch | `response_schema` requires `response_format: json_schema`, and vice versa — both must be present together |
| DIP132 | Invalid JSON in response_schema | Fix the JSON syntax |
| DIP133 | Params key shadows field | Rename the params key |
| DIP139 | Invalid `tool_access` value | Use `tool_access: none` or omit the field; runtime fail-closes on unknown values |
| DIP140 | `params` re-enables tools that `tool_access` strips | Remove the `params` key (`allowed_tools`, `disallowed_tools`, `tool_choice`, `permission_mode`); `tool_access` governs the catalog |
| DIP141 | `writable_paths` set with `tool_access: none` (dead config) | Remove `writable_paths` — `tool_access: none` already strips all tools, leaving nothing for `writable_paths` to bound |
| DIP142 | Unsafe `writable_paths` entry (absolute, `..` escape, `~`, brace fragment) | Use workspace-relative globs (e.g. `.ai/sprints/**`); absolute, `~`, Windows-drive, and `..`-escaping entries are rejected by the runtime fs jail |
| DIP143 | Workflow uses `tool_access` but references a subgraph — child agents unguarded | Open the referenced `.dip` and set `tool_access` on its agents directly; parent restrictions do not cross the file boundary |
| DIP144 | Agent node has no failure route | Add `-> <node> when ctx.outcome = fail` (dip 2), or set `fallback_target`/`retry_target`+`max_retries` (v1), or declare `on_failure:` in defaults |
| DIP145 | Graph budget default is negative | Use a positive cap, or omit the field / set `0` to mean no limit |
| DIP146 | Referenced subgraph child restricts no agent's `tool_access` while a workflow on the path does (cross-file) | Set `tool_access` on the child's agents; native `dippin lint` resolves the child and supersedes DIP143 |
| DIP147 | A `tool_access: none` agent `writes:` a context key that a downstream tool-bearing agent `reads:` (chain-attack / info-flow) | Confirm the restricted agent's input is trusted; if not, give the consumer `tool_access: none` too or insert a sanitizing step. Detection only — the runtime enforces |
| DIP149 | A node has 2+ unconditional outgoing edges — which one fires is decided only by the lexical (alphabetical) tiebreak on target node ID | Keep at most one unconditional edge as the default fallback; guard the others with `when`. Restart/`loop` back-edges are exempt |
| DIP150 | A label-routing `human` gate (mode `choice`, `yes_no`, or unset default) routes an outgoing edge by `label:` with no explicit `choice:` — the label doubles as the routing key, so nothing signals it is load-bearing (Hint). `freeform`/`interview` gates are exempt | Add `choice: "<key>"` to mark the routing key, leaving `label:` for display. `choice:` wins when present; `label:` still routes when it is absent |
| DIP151 | An edge carries a `weight:` attribute — speculative tier-4 routing priority that the cascade never consults and no real workflow uses. It still parses (carry-only), but the keyword and cascade tier are slated for removal in `dip 2` (Warning) | Remove `weight:`; express priority with conditions — guard edges with `when` / `on`, or rely on a single unconditional fallback |
| DIP152 | A tool node's `marker_grep` enumerates a literal marker that no edge routes and that no section `else ->` default or unconditional edge covers — the marker would be emitted at runtime with nowhere to go. Only checked for recognizable literal-alternation greps; any compound/negated/other-variable edge makes the node safe (Warning) | Route the marker with an edge (`on <marker>`), add an unconditional fallback edge, or add a section `else -> <node>` default |
| DIP153 | An `edges` block re-declares an unconditional, attribute-free edge that already exists as an inline `parallel`/`fan_in` fork — the inline list is authoritative (validation, simulation, DOT all derive fan edges from it), so the re-declaration is redundant (Warning). A conditional/attributed edge between the same nodes is kept | Remove the redundant edge (`dippin fmt` strips it); rejected outright under `dip 2` |
| DIP154 | An agent sets `prompt_prefix: none` / `prompt_suffix: none` to opt out of the defaults prompt cascade, but no cascade of that kind is declared — the opt-out is a no-op (Hint) | Remove the unnecessary opt-out, or add the intended `prompt_prefix`/`prompt_suffix`(`_file`) cascade to `defaults` |

## Best Practices

- **Always set `timeout`** on tool nodes — no timeout means infinite hang
- **Prefer `marker_grep:`** over regexing `ctx.tool_stdout` in edges when the runtime supports it. Typed routing leaves stdout free for diagnostic output and avoids truncation foot-guns. Declaring `marker_grep:` also suppresses DIP101/DIP102 on the source node — the validator treats it as a safe typed-routing channel — but for a recognizable literal-alternation grep, DIP152 still flags any enumerated marker that no edge routes and no `else`/unconditional fallback covers.
- **Boolean fields** (`goal_gate`, `auto_status`, `cache_tools`, `route_required`) accept `true/false`, `1/0`, `yes/no`, `on/off` case-insensitively. Anything else is a parse diagnostic.
- **Use `auto_status: true`** on agent nodes that drive conditional routing via `ctx.outcome`
- **Use `success`/`fail`** as condition values — the linter recognizes these as exhaustive
- **Mark back-edges `loop`** — loops without it trigger DIP005
- **Declare `reads`/`writes`** on nodes to document data flow (enables DIP107/DIP112 checks)
- **Add a failure route** (an `on fail` edge, or `retry_target`/`fallback_target` in v1) to `goal_gate: true` nodes
- **Run `dippin check`** after every edit — it catches issues the formatter won't
- **Use `dippin doctor`** for a health grade and actionable improvement suggestions

## Common Patterns

### Review Loop
```
  agent Implement
    prompt: Build the feature.
    auto_status: true
  agent Review
    prompt: Review the implementation.
    auto_status: true
  edges
    Implement -> Review
    Review -> Done when ctx.outcome = success
    Review -> Implement when ctx.outcome = fail loop
```

### Human Gate
```
  human Approve
    mode: choice
  edges
    Review -> Approve
    Approve -> Deploy label: Approve
    Approve -> Revise label: Request Changes
```

### Parallel Fan-Out
```
  parallel Split -> Claude, GPT, Gemini
  fan_in Merge <- Claude, GPT, Gemini
  edges
    Analyze -> Split
    Claude -> Merge
    GPT -> Merge
    Gemini -> Merge
    Merge -> Consensus
```

## Context Variables

| Variable | Source |
|----------|--------|
| `ctx.outcome` | `auto_status: true` on agent, or tool exit status |
| `ctx.human_response` | Freeform human input |
| `ctx.tool_stdout` | Tool command stdout |
| `ctx.tool_marker` | Tool stdout regex match (when `marker_grep` declared) |
| `ctx.tool_route` | A routing value the runtime extracts from the tool's stdout — populated when the tool emits a routing sentinel the runtime recognizes (format defined by the runtime); `route_required: true` additionally fails the node if none is emitted |
| `ctx.preferred_label` | Human choice selection (maps to edge label) |
| `ctx.interview_answers` | Interview mode answers (via `answers_key`) |
| `params.<key>` | Parent subgraph params |
| `graph.<field>` | Workflow-level metadata |


---

## Documentation

- [Language Reference](https://dippin.org/language/)
- [CLI Reference](https://dippin.org/cli/)
- [Validation & Linting](https://dippin.org/validation/)
- [Scenario Testing](https://dippin.org/testing/)
- [Analysis Tools](https://dippin.org/analysis/)
- [Playground](https://dippin.org/playground/)
- [GitHub](https://github.com/2389-research/dippin-lang)
