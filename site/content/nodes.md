---
title: "Nodes"
description: "The eight .dip node kinds — agent, human, tool, parallel, fan_in, subgraph, manager_loop, conditional — with their shared and per-kind fields."
section_label: "Language"
subtitle: "The eight node kinds and their fields."
---

## The 8 Node Kinds

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

## Common Fields

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

The retry fields (`retry_policy`, `max_retries`, `base_delay`, `retry_target`, `fallback_target`/`fallback_retry_target`) are summarized here only — see Configuration → Retry & Recovery for the full retry configuration.

## agent

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

## human

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

## tool

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

## parallel

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

## fan_in

Fan-in nodes join concurrent branches back together. Sources must match the targets of a corresponding `parallel` node.

```
  fan_in Join <- TaskA, TaskB, TaskC
```

## subgraph

Subgraph nodes embed another workflow as a single step. Parameters are passed via the `params.*` namespace.

```
  subgraph ReviewProcess
    ref: review_pipeline
    params:
      strict: true
      model: gpt-5.4
```

## manager_loop

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

## conditional

Conditional nodes evaluate outgoing edge conditions without making an LLM call — pure routing with zero token cost.

```dippin
conditional CheckOutcome
  label: "Route by Result"
```

Conditional nodes accept only common fields (`label`, `class`, `reads`, `writes`). No `prompt`, `model`, or `provider`. Maps to `diamond` shape in DOT export.
