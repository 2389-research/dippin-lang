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

A `.dip` file may optionally declare its **format version** on the first line, before the `workflow` declaration — e.g. `dip 2`. With no declaration the version defaults to **1**, and the formatter only emits the `dip N` line for versions greater than 1 (a v1 file never gains one). The version is parsed before the workflow body so future format versions can change edge syntax wholesale; `dippin fmt --migrate` re-emits a file in its current format version, still formatting it to canonical form — today the v1→v1 migration itself is an identity pass with no version transform.

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
```

| Field | Type | Description |
|-------|------|-------------|
| `model` | String | Default LLM model for all agent nodes |
| `provider` | String | Default LLM provider (e.g., "openai", "anthropic") |
| `retry_policy` | String | Default retry strategy name |
| `max_retries` | Integer | Default max retry attempts per node |
| `fidelity` | String | Default checkpoint fidelity level |
| `max_restarts` | Integer | Max loop restarts before pipeline failure (default: 5) |
| `cache_tools` | Boolean | Whether to cache tool call results |
| `compaction` | String | Context compaction mode for long pipelines |
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
| `system_prompt` | Block | System-level instructions prepended before the prompt |
| `reasoning_effort` | String | Extended thinking effort level: `none`, `minimal`, `low`, `medium`, `high`, `xhigh`, `max`. Controls how much reasoning budget the LLM spends. |
| `fidelity` | String | Checkpoint fidelity level for state persistence. |
| `max_turns` | Integer | Maximum conversation turns before the node exits |
| `auto_status` | Boolean | Automatically extract `STATUS: success/fail` from model output into `ctx.outcome` |
| `goal_gate` | Boolean | Marks this node as a goal gate — requires `retry_target` or `fallback_target` for recovery |
| `reads` | String | Context key this node reads as input (advisory metadata) |
| `writes` | String | Context key this node writes as output (advisory metadata) |
| `response_format` | String | Structured output mode: `json_object` or `json_schema`. Instructs the model to return valid JSON. |
| `response_schema` | Block | JSON Schema definition enforced when `response_format: json_schema` is set. Must be valid JSON. |
| `params` | Block | Arbitrary key-value pairs forwarded to the provider API. Keys must not duplicate first-class fields (see DIP133). |
| `cmd_timeout` | Duration | Maximum wall-clock time for this agent call before the runtime cancels and errors (e.g., `30s`, `2m`). |

### human

Human nodes pause execution and wait for human input. Two modes: `choice` (predefined options from edge labels) and `freeform` (open text input).

```
  human Approve
    label: "Ship it?"
    mode: choice
    default: "yes"
```

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

Declare `marker_grep` for typed routing (populates `ctx.tool_marker`); `route_required: true` makes the node fail if the command emits no routing signal recognized by the runtime; `output_limit` overrides the captured-stdout byte cap.

### parallel

Parallel nodes fan execution out to multiple branches that run concurrently. Every `parallel` must have a matching `fan_in`.

```
  parallel FanOut -> TaskA, TaskB, TaskC
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
<FromID> -> <ToID> [on <token> | when <condition>] [label: <text>] [choice: <key>] [weight: <int>] [loop]
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
| `weight: <int>` | Integer | Priority hint — higher wins among competing edges |
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
