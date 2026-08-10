---
title: "Language Reference"
description: "Full syntax reference for .dip workflow files: nodes, edges, conditions, multiline prompts, parallel execution, and stylesheets."
section_label: "Syntax"
subtitle: "Complete syntax for .dip workflow files."
---

## File Structure

Every `.dip` file contains exactly one workflow. The top-level structure has up to five sections, in this order:

<div class="flow-diagram">
  <div class="flow-step">workflow &lt;name&gt;</div>
  <div class="flow-arrow">&rarr;</div>
  <div class="flow-step">Header<br>goal, requires, start, exit</div>
  <div class="flow-arrow">&rarr;</div>
  <div class="flow-step">Defaults (optional)<br>model, provider, ...</div>
  <div class="flow-arrow">&rarr;</div>
  <div class="flow-step">Vars (optional)<br>key: value, ...</div>
  <div class="flow-arrow">&rarr;</div>
  <div class="flow-step">Node Definitions<br>agent, human, tool, ...</div>
  <div class="flow-arrow">&rarr;</div>
  <div class="flow-step">Edges (optional)<br>A -&gt; B when ...</div>
</div>

Dippin uses indentation-sensitive syntax (like Python). Use 2 spaces or tabs consistently. The canonical formatter always outputs 2-space indentation.

A `.dip` file may optionally declare its **format version** on the first line, before the `workflow` declaration — e.g. `dip 2`. With no declaration the version defaults to **1**, and the formatter only emits the `dip N` line for versions greater than 1 (a v1 file never gains one). The version is parsed before the workflow body so a format version can change edge syntax wholesale; `dippin fmt --migrate` converts a v1 file to `dip 2` — a lossless version bump that keeps the retry channel on the node, relabeling `fallback_target` to its dip-2 spelling `fallback_retry_target`.

## Workflow Header

The workflow declaration is the first line, followed by required and optional header fields:

```
workflow my_pipeline
  goal: "Ask user for a task, implement it, review, ship"
  start: AskUser
  exit: Done
```

| Field | Required | Description |
|-------|----------|-------------|
| `workflow <name>` | Yes | Declares the workflow and its identifier |
| `goal: <text>` | No | Human-readable objective for this pipeline |
| `requires: <id>[, <id>...]` | No | Workflow-level declared prerequisites (advisory; comma-separated identifiers — tools, MCP servers, env vars). Mirrors node-level `reads:` / `writes:` for shape. |
| `start: <NodeID>` | Yes | Entry point node — execution begins here |
| `exit: <NodeID>` | Yes | Terminal node — execution ends here |

## Defaults Block

The optional `defaults` block sets graph-level configuration that applies to all nodes unless overridden at the node level.

```
  defaults
    model: claude-opus-4-6
    provider: anthropic
    retry_policy: standard
    max_retries: 3
    fidelity: high
    max_restarts: 5
    cache_tools: true
    compaction: summary
    stall_timeout: 5m
    on_failure: Escalate
    prompt_suffix_file: protocols/status-contract.md
```

The `prompt_prefix`/`prompt_suffix` (inline) and `prompt_prefix_file`/`prompt_suffix_file` (fragment file) defaults **cascade a shared prompt fragment to every agent** (#175) — the effective prompt is composed as `prefix → body → prompt_include → suffix` at resolve time, with the suffix always last. An agent opts out with `prompt_suffix: none` / `prompt_prefix: none`. The `system_prompt_file` default (#72) is different in kind: it is a **fallback** shared system prompt (persona) — used only by agents that declare no `system_prompt`/`system_prompt_file` of their own, which fully override it.

| Field | Type | Description |
|-------|------|-------------|
| `model` | String | Default LLM model for all agent nodes |
| `prompt_prefix` / `prompt_suffix` | String | Inline prompt fragment cascaded to every agent as a prefix/suffix (#175) |
| `prompt_prefix_file` / `prompt_suffix_file` | String | Prompt fragment loaded from a file and cascaded to every agent (mutually exclusive with the inline form) |
| `system_prompt_file` | String | Shared system prompt (persona) loaded from a file; a fallback default for agents that set none of their own — a node's own `system_prompt`/`system_prompt_file` overrides it (#72). File form only |
| `provider` | String | Default LLM provider (e.g., "openai", "anthropic") |
| `retry_policy` | String | Default retry strategy name |
| `max_retries` | Integer | Default max retry attempts per node |
| `fidelity` | String | Default checkpoint fidelity level |
| `max_restarts` | Integer | Max loop restarts before pipeline failure (default: 5) |
| `restart_target` | String | Node ID to jump to on restart loops |
| `cache_tools` | Boolean | Whether to cache tool call results |
| `compaction` | String | Context compaction mode for long pipelines |
| `on_resume` | String | Fidelity behavior when a run resumes: `preserve` (keep the checkpoint fidelity level) or `degrade` (downgrade on resume). Only meaningful when `fidelity` is also set. |
| `max_total_tokens` | Integer | Hard ceiling on total tokens across the run. `0`/unset = no limit. |
| `max_cost_cents` | Integer | Hard ceiling on total cost, in **US cents** (e.g. `1000` = $10.00). `0`/unset = no limit. |
| `max_wall_time` | Duration | Hard ceiling on **wall-clock** run time (e.g. `30m`, `2h`). `0`/unset = no limit. |
| `stall_timeout` | Duration | Abort/route when no forward progress is made for a wall-clock span (e.g. `30s`, `5m`); `0`/unset = no limit. Enforced by the runtime. |
| `on_failure` | NodeID | Graph-level catch-all failure route — the runtime sends a failing node here when no more specific route (fail edge → bounded retry → `fallback_target`) matches. Carried + linted by dippin; enforced by the runtime. |

### Tool safety

Tool nodes that shell out can be constrained by two defaults consumed by the runtime:

- `tool_commands_allow` — comma-separated glob allowlist. When set, the runtime rejects tool-node commands that do not match any pattern.
- `tool_denylist_add` — comma-separated globs appended to the runtime's default denylist (on top of the runtime's built-in blocks).

```dippin
workflow Safe
  goal: "Constrained tool execution"
  start: A
  exit: A

  defaults
    tool_commands_allow: "git *,make *"
    tool_denylist_add: "rm -rf /,dd *"

  # ...
```

Values pass through to the runtime verbatim; dippin-lang does not validate glob syntax.

## Vars Block

The optional `vars` block declares user-defined variables that are substituted wherever `$key` placeholders appear in prompts and commands.

```
  vars
    source_ref: "references/claude-agent-sdk-python/src"
    target_name: claude-agents-rs
    target_module: "claude-agents-rs/src/"
```

Values can be quoted strings or bare identifiers. Keys must be unique — duplicate keys cause a parse error.

Vars are exported as graph-level DOT attributes so they round-trip through `dippin export-dot` and `dippin migrate`.

## Inputs Block

The optional `inputs` block declares the workflow's callee-side signature — the named values a caller (a human at the entry point, or a parent workflow via a `subgraph` node's `params:`) must or may supply at run start:

```
  inputs
    idea: text
      required: true
      prompt: "What do you want built?"
      description: "One or two sentences describing the change."
      max_length: 4000
      multiline: true
    target_branch: text
      default: main
      pattern: "^[A-Za-z0-9._/-]+$"
    risk: enum
      default: medium
      options: low, medium, high
```

Each entry is `name: type` with an optional indented block of attributes (`required`, `prompt`, `description`, `default`, `options`, `pattern`, `min`, `max`, `max_length`, `multiline`). The six v1 types are `text`, `number`, `bool`, `enum`, `file`, `secret`. Declaration order is significant — a host renders inputs as an ordered form — so the formatter never reorders entries.

Reference a declared input as `${inputs.name}` in prompts. The `inputs` namespace is closed: an unrecognized type is DIP155, a reference to an undeclared input is DIP156, and a reference inside a tool node's `command:` is DIP157 (inputs never interpolate into a shell).

## Node Kinds

There are 8 node kinds, each with its own syntax and configuration:

<div class="flow-diagram">
  <div class="pipeline-box lavender">agent<br>LLM interaction</div>
  <div class="pipeline-box green">human<br>Decision gate</div>
  <div class="pipeline-box yellow">tool<br>Shell command</div>
  <div class="pipeline-box lavender">parallel<br>Fan-out</div>
  <div class="pipeline-box green">fan_in<br>Join</div>
  <div class="pipeline-box yellow">subgraph<br>Sub-pipeline</div>
  <div class="pipeline-box cream">conditional<br>Pure routing</div>
  <div class="pipeline-box lavender">manager_loop<br>Child supervisor</div>
</div>

### Common Fields

These fields (`label`, `class`, `reads`, `writes`, and the retry fields) are accepted by **all** block-style node kinds — agent, human, tool, subgraph, `conditional`, and `manager_loop` — since they share the node-field parser:

| Field | Type | Description |
|-------|------|-------------|
| `label` | String | Human-readable display name. Defaults to the node ID if omitted. |
| `class` | CSV | Comma-separated stylesheet class names for theming (reserved for post-v1). |
| `reads` | CSV | Context keys this node expects to read. Advisory — used for linting (DIP112), not enforced at runtime. |
| `writes` | CSV | Context keys this node will produce. Advisory — used for linting (DIP107), not enforced at runtime. |
| `retry_policy` | String | Named retry strategy: `standard`, `aggressive`, `patient`, `linear`, `none`. Overrides the workflow default. |
| `max_retries` | Integer | Maximum retry attempts before giving up. Overrides the workflow default. |
| `base_delay` | Duration | Override the retry policy's default base delay (e.g. `500ms`, `2s`, `1m`). |
| `retry_target` | String | Node ID to jump to when retrying — the engine's retry channel, read from the node (not an edge). Same spelling in `dip 1` and `dip 2`. |
| `fallback_target` / `fallback_retry_target` | String | Node ID to route to when all retries are exhausted (read from the node, not an edge). Spelled `fallback_target` in `dip 1`, `fallback_retry_target` in `dip 2`; `dippin fmt --migrate` relabels it. |

### agent

Agent nodes invoke an LLM. They are the most configurable node kind. Key fields include `model`, `provider`, `prompt`, `system_prompt`, `max_turns`, `auto_status`, and `goal_gate`.

```
  agent Analyze
    label: "Analyze the request"
    model: claude-opus-4-6
    provider: anthropic
    goal_gate: true
    auto_status: true
    reads: human_response
    writes: analysis
    prompt:
      You are a senior software architect.
      Analyze the following request carefully.
```

| Field | Type | Description |
|-------|------|-------------|
| `model` | String | LLM model to use (overrides defaults) |
| `provider` | String | LLM provider (e.g., "anthropic", "openai") |
| `backend` | String | Per-node backend override (e.g., `native`, `claude-code`, `acp`) |
| `working_dir` | String | Per-node working directory override for isolated execution. |
| `prompt` | Block | Multiline prompt text sent to the model |
| `prompt_file` | String | Path (relative to the `.dip` dir) to an external file whose contents become the prompt. Mutually exclusive with `prompt:` — setting both is a parse error. |
| `system_prompt` | Block | System-level instructions prepended before the prompt |
| `system_prompt_file` | String | Path (relative to the `.dip` dir) to an external file whose contents become the system prompt. Mutually exclusive with `system_prompt:` — setting both is a parse error. |
| `prompt_include` | String | Path to a fragment file appended after the body, before the defaults cascade suffix (#175). |
| `prompt_prefix` / `prompt_suffix` | `none` | Set to `none` to opt this agent out of the corresponding `defaults` prompt cascade (#175). |
| `tool_access` | String | LLM tool-catalog gate. Set to `none` to strip the model's tool registry on this agent. DIP139 warns on unknown values; the runtime fail-closes. |
| `writable_paths` | CSV (globs) | Comma-separated glob list bounding where this agent's tools may write (e.g. `workspace/**, .ai/sprints/**`). Absent = unbounded; a present-but-empty value is rejected by `dippin validate`/`pack`. Enforced by the runtime. |
| `last_response_truncate` | Integer | Caps how much of the prior node's response is carried into this agent's context, in characters. `0`/unset = no truncation (a negative value raises DIP148). |
| `reasoning_effort` | String | Extended thinking effort level: `none`, `minimal`, `low`, `medium`, `high`, `xhigh`, `max`. Controls how much reasoning budget the LLM spends. |
| `fidelity` | String | Checkpoint fidelity level for state persistence. |
| `max_turns` | Integer | Maximum conversation turns before the node exits |
| `auto_status` | Boolean | Automatically extract `STATUS: success/fail` from model output into `ctx.outcome` |
| `goal_gate` | Boolean | Marks this node as a goal gate — requires a failure route (an `on fail` edge; or `retry_target`/`fallback_target` in v1) for recovery |
| `cache_tools` | Boolean | Whether to cache tool call results for this agent. Overrides the workflow default. |
| `compaction` | String | Context compaction mode for managing long context windows. Overrides the workflow default. |
| `compaction_threshold` | Float | Threshold value that triggers compaction (provider-specific semantics). |
| `reads` | CSV | Context keys this node reads as input (advisory metadata) |
| `writes` | CSV | Context keys this node writes as output (advisory metadata) |
| `response_format` | String | Structured output mode: `json_object` or `json_schema`. Instructs the model to return valid JSON. |
| `response_schema` | Block | JSON Schema definition enforced when `response_format: json_schema` is set. Must be valid JSON. |
| `params` | Block | Arbitrary key-value pairs forwarded to the provider API. Keys must not duplicate first-class fields (see DIP133). |
| `cmd_timeout` | Duration | Maximum wall-clock time for this agent call before the runtime cancels and errors (e.g., `30s`, `2m`). |

### human

Human nodes pause execution and wait for human input. Four modes: `choice` (predefined options from edge labels), `freeform` (open text input), `interview` (structured Q&A from upstream agent output), and `yes_no` (binary Y/N prompt).

```
  human Approve
    label: "Ship it?"
    mode: choice
    default: "yes"
```

| Field | Type | Description |
|-------|------|-------------|
| `mode` | String | Interaction mode: `choice`, `freeform`, `interview`, or `yes_no`. |
| `default` | String | Default selection if no input. Only meaningful for `choice` mode. |
| `prompt` | Block | Prompt text shown to the human (also the interview fallback when no questions are detected). |
| `questions_key` | String | Context key to read questions from. Interview mode only (default `interview_questions`). |
| `answers_key` | String | Context key to write answers to. Interview mode only (default `interview_answers`). |
| `timeout` | Duration | How long to wait for input before `timeout_action` fires (e.g. `5m`). `0`/unset = wait indefinitely. |
| `timeout_action` | String | What to do when `timeout` elapses: `fail` (the node fails), `default` (use the `default` selection), or empty. Empty falls back to the node's `default` answer if one is set, otherwise fails. Any other value is a parse error. |

### tool

Tool nodes execute shell commands. The command's stdout is captured as `ctx.tool_stdout` and stderr as `ctx.tool_stderr`. Always include a `timeout`.

```dippin
  tool RunTests
    label: "Run test suite"
    outputs: tests_pass, tests_fail
    marker_grep: "^(tests_pass|tests_fail)$"
    timeout: 60s
    command:
      pytest --tb=short
```

Declare `marker_grep` for typed routing (populates `ctx.tool_marker`); `route_required: true` makes the node fail if the command emits no routing signal recognized by the runtime; `output_limit` overrides the captured-stdout byte cap. Use `command_file` (a path relative to the `.dip` dir) instead of inline `command:` to load the script from an external file — the two are mutually exclusive.

### parallel

Parallel nodes fan execution out to multiple branches that run concurrently. Every `parallel` must have a matching `fan_in`.

```
  parallel FanOut -> TaskA, TaskB, TaskC
```

Use **block form** when branches need different models, providers, fidelity levels, or tool access. Each `branch:` entry declares a fan-out target (equivalent to an inline `->` target) and attaches per-branch overrides for `model`, `provider`, `fidelity`, `tool_access`, `writable_paths`, and `last_response_truncate`. The fan-in node must still list the same target IDs. An omitted branch `tool_access` or `writable_paths` inherits the target agent's setting — it never re-grants the full catalog or resets to unbounded.

```
  parallel split
    branch: fast
      model: claude-haiku-4-5
      provider: anthropic
      fidelity: summary
    branch: accurate
      model: claude-opus-4-7
      provider: anthropic
      fidelity: full
```

### fan_in

Fan-in nodes join concurrent branches back together. Sources must match the targets of a corresponding `parallel` node.

```
  fan_in Join <- TaskA, TaskB, TaskC
```

### subgraph

Subgraph nodes embed another workflow as a single step. Parameters are passed via the `params.*` namespace.

```
  subgraph ReviewProcess
    ref: review_pipeline
    params:
      strict: true
      model: gpt-5.4
```

### manager_loop

Manager loop nodes supervise a child sub-pipeline: they spawn it, poll it on a configurable cadence, and can steer it by injecting additional context during execution. They map to `stack.manager_loop` in the runtime and export as DOT shape `house`.

```dippin
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

| Field | Type | Description |
|-------|------|-------------|
| `subgraph_ref` | String | **Required.** Path to the child `.dip` file (DIP135 if missing or not found) |
| `poll_interval` | Duration | How often to poll the child (e.g. `30s`, `5m`). `0` means event-driven |
| `max_cycles` | Integer | Maximum poll cycles before the node exits. `0` = unbounded — triggers DIP137 |
| `stop_condition` | Condition | Expression over `stack.child.*` evaluated each cycle; when true the loop exits |
| `steer_condition` | Condition | When true, inject `steer_context` into the running child |
| `steer_context` | map[string]string | Key-value pairs injected on steer. Inline `k=v, k=v` or indented block form. Inline values may not contain commas |

**Runtime state** exposed as context variables: `stack.child.cycles`, `stack.child.outcome`, `stack.child.status`.

**Lint codes:** DIP135 (subgraph_ref missing or file not found), DIP136 (invalid control field value), DIP137 (unbounded loop — max_cycles: 0).

See [docs/nodes.md](https://github.com/2389-research/dippin-lang/blob/main/docs/nodes.md) for the complete field reference.

### Conditional Nodes

Conditional nodes evaluate outgoing edge conditions without making an LLM call — pure routing with zero token cost.

```dippin
conditional CheckOutcome
  label: "Route by Result"
```

Conditional nodes accept only common fields (`label`, `class`, `reads`, `writes`). No `prompt`, `model`, or `provider`. Maps to `diamond` shape in DOT export.

## Edges

The `edges` block defines connections between nodes. Each edge is a single line:

```
<FromID> -> <ToID> [on <token> | when <condition>] [label: <text>] [choice: <key>] [weight: <int>] [loop] [override: true]
```

### Basic and Conditional Edges

```
  edges
    AskUser -> Interpret
    Interpret -> Validate
    Validate -> Approve   when ctx.outcome = success
    Validate -> Retry     when ctx.outcome = fail
```

### Edge Attributes

| Attribute | Type | Description |
|-----------|------|-------------|
| `on <token>` | Token | Shorthand for an equality guard against the source node's outcome channel — agent (`ctx.outcome`) or tool + `marker_grep` (`ctx.tool_marker`) |
| `when <expr>` | Condition | Boolean guard — edge only traversed if true |
| `label: <text>` | String | Human-readable label (used for human gate choices when no `choice:` is set) |
| `choice: <key>` | String | Carried, not interpreted by dippin — explicit human-gate routing key the paired runtime matches the user's selection against, leaving `label:` for display. Wins over `label:` when present; runtime falls back to `label:` when absent |
| `weight: <int>` | Integer | **Soft-deprecated** (raises DIP151). Parsed but **unused by routing**; slated for removal in `dip 2`. Historically a priority hint, but the cascade never consults it — guard edges with `when` / `on` instead |
| `loop` | Flag | Bare keyword marking a back-edge (loop restart). Legacy `restart: true` is an accepted synonym that `dippin fmt` rewrites to `loop` |
| `override: true` | Boolean | Carried, not interpreted by dippin — marks a human-authored validation override for a paired runtime to act on |

### Outcome Shorthand (`on`)

The most common condition is an equality test against the source node's *natural outcome channel*. `on <token>` is shorthand for exactly that:

```
    Review -> Approve  on success      # = when ctx.outcome = success
    Review -> Reject   on fail         # = when ctx.outcome = fail
    Tests  -> Ship     on tests_green  # = when ctx.tool_marker = tests_green
```

The channel comes from the source node's kind: **agent** nodes use `ctx.outcome`; **tool** nodes that declare `marker_grep` use `ctx.tool_marker`. `on` is pure sugar — it produces the identical condition, is non-breaking, and `dippin fmt` rewrites eligible `when` edges into `on` form. Sources without an outcome channel — human gates, `conditional` nodes, and tools without `marker_grep` — have no `on` channel; use `when` there, as well as for anything that isn't a single equality (`and`/`or`, `contains`, `!=`, etc.).

### Loop Edges

A `loop` edge creates a controlled back-edge. When followed, the engine increments a restart counter (max controlled by `max_restarts`), clears downstream nodes, resets retry budgets, and resumes from the target.

```
    Validate -> Implement when ctx.outcome = fail loop
```

`loop` is a bare keyword; the legacy `restart: true` still parses and `dippin fmt` rewrites it to `loop`. Because `loop` is a reserved bare keyword, write `when ctx.x = "loop"` (quoted) for the literal value on a condition's right-hand side.

### Section-Level Default (`else`)

A single `else -> <node>` line, written at the bottom of the `edges` block, is the graph's **success-side default destination**: any node whose guard edges all fail to match, and which has no explicit unconditional edge of its own, routes there.

```
  edges
    SetupRun -> FetchIssues  on setup-ok
    RunTests -> Package      on tests-ok
    else -> Cleanup
```

At most one `else` per edges block (a second is a parse error), and `else ->` requires a target node. It has no source node. `else` is success-side only — it never intercepts a genuine node *failure*, which routes via the failure cascade (`on fail` edge → `defaults.on_failure`). A node covered by `else` is not flagged DIP101 or DIP102.

> **Note:** `dippin simulate` / `dippin test` do not yet traverse the `else` default (tracked in [#158](https://github.com/2389-research/dippin-lang/issues/158)); a paired runtime resolves it.

## Conditions

Conditions appear on edges after the `when` keyword. All operators perform string comparison.

String values in conditions and `label:` may be **unquoted** (`success`), **double-quoted** (`"needs review"`, supporting `\"` and `\\` escapes), or **single-quoted** (YAML-style literal, where `''` is the only escape — handy for regex-like values: `'^(green|red)$'`). A single-quoted edge value normalizes to double quotes on `dippin fmt`.

### Comparison Operators

| Operator | Meaning | Example |
|----------|---------|---------|
| `=`, `==` | String equality | `ctx.outcome = success` |
| `!=` | String inequality | `ctx.outcome != fail` |
| `contains` | Substring match | `ctx.response contains "approved"` |
| `not contains` | Negated substring | `ctx.tool_stdout not contains all-done` |
| `startswith` | Prefix match | `ctx.response startswith "yes"` |
| `endswith` | Suffix match | `ctx.response endswith "done"` |
| `in` | Value in list | `ctx.status in "pass,fail,skip"` |

### Logical Operators

| Operator | Meaning | Precedence |
|----------|---------|------------|
| `not` | Logical negation | Highest |
| `and` | Logical AND | Medium |
| `or` | Logical OR | Lowest |

Parentheses control precedence:

```
    A -> B when ctx.outcome = success and ctx.score = high
    A -> C when ctx.outcome = fail or ctx.status = blocked
    A -> D when (ctx.x = 1 or ctx.y = 2) and ctx.z = 3
```

## Multiline Blocks

Fields like `prompt` and `command` support multiline content. Write the key followed by `:`, then indent the content on subsequent lines:

```
  agent MyAgent
    prompt:
      You are a code reviewer.

      ## Rules
      - Check for bugs
      - Check for security issues

      ## Context
      ${ctx.last_response}
```

The first content line's indentation sets the baseline. All content is de-indented by that amount. Empty lines are preserved. The block ends when indentation returns to or above the field's level. No quoting or escaping needed.

```
  tool RunTests
    timeout: 60s
    command:
      #!/bin/sh
      set -eu
      if pytest --tb=short 2>&1; then
        printf 'pass'
      else
        printf 'fail'
        exit 1
      fi
```
