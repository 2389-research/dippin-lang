# Dippin Language Syntax Reference

This document describes the complete syntax of the `.dip` file format — the human-friendly alternative to DOT for authoring AI-driven workflow pipelines.

---

## File Structure

Every `.dip` file contains exactly one workflow. The top-level structure is:

```mermaid
graph TD
    WF["workflow &lt;name&gt;"]
    WF --> H["Header<br>goal, requires, start, exit"]
    WF --> D["Defaults (optional)<br>model, provider, retry_policy, ..."]
    WF --> N["Node Definitions<br>agent, human, tool, parallel, fan_in, subgraph"]
    WF --> E["Edges Section (optional)<br>A -> B when condition"]
    H ~~~ D ~~~ N ~~~ E
```

```dippin
workflow <name>
  <header fields>

  [defaults]
    <key: value pairs>

  <node definitions>

  [edges]
    <edge definitions>
```

Sections must appear in this order: header, defaults, nodes, edges.

---

## Indentation

Dippin uses **indentation-sensitive syntax** (like Python). Indentation defines scope — child elements are indented relative to their parent.

- Use **2 spaces** or **tabs** consistently (do not mix)
- The canonical formatter always outputs 2-space indentation
- Indentation level changes produce `INDENT` / `OUTDENT` tokens internally

```dippin
workflow Example         # level 0
  goal: "Do a thing"    # level 1 (child of workflow)
  start: A

  agent A                # level 1 (child of workflow)
    prompt:              # level 2 (child of agent A)
      Hello world        # level 3 (multiline content)
```

---

## Comments

Line comments start with `#` and extend to the end of the line. They are ignored by the parser.

```dippin
# This is a comment
workflow Example   # Inline comment
  start: A
```

Section headers like `# ── Phase 1 ──────────` are purely decorative — the parser treats them as regular comments.

---

## Workflow Header

The workflow declaration is the first line, followed by required and optional header fields:

```dippin
workflow my_pipeline
  goal: "Ask user for a task, implement it, review, ship"
  requires: git, docker
  start: AskUser
  exit: Done
```

| Field | Required | Description |
|-------|----------|-------------|
| `workflow <name>` | Yes | Declares the workflow and its identifier |
| `goal: <text>` | No | Human-readable objective for this pipeline |
| `requires: <list>` | No | Comma-separated environmental dependencies (e.g. `git, docker, jq`). Surfaced as `Workflow.Requires` in the IR; semantics live in downstream consumers (preflight checks, etc.). Unknown entries are accepted without diagnostic. |
| `start: <NodeID>` | Yes | Entry point node — execution begins here |
| `exit: <NodeID>` | Yes | Terminal node — execution ends here |

The name on the `workflow` line is an identifier (alphanumeric, underscores, dashes). Goal is a quoted or unquoted string.

---

## Defaults Block

The optional `defaults` block sets graph-level configuration that applies to all nodes unless overridden at the node level.

```dippin
  defaults
    model: claude-opus-4-6
    provider: anthropic
    retry_policy: standard
    max_retries: 3
    fidelity: high
    max_restarts: 5
    restart_target: Start
    cache_tools: true
    compaction: summary
```

| Field | Type | Description |
|-------|------|-------------|
| `model` | String | Default LLM model for all agent nodes |
| `provider` | String | Default LLM provider (e.g., "openai", "anthropic") |
| `retry_policy` | String | Default retry strategy name |
| `max_retries` | Integer | Default max retry attempts per node |
| `fidelity` | String | Default checkpoint fidelity level |
| `max_restarts` | Integer | Max loop restarts before pipeline failure (default: 5) |
| `restart_target` | String | Node ID to jump to on restart loops |
| `cache_tools` | Boolean | Whether to cache tool call results |
| `compaction` | String | Context compaction mode for long pipelines |
| `on_failure` | String | Graph-level default failure route — node to jump to when an agent has no other failure route (see [edges.md](edges.md) for the full precedence cascade) |
| `max_total_tokens` | Integer | Hard ceiling on total tokens across the run. `0`/unset = no limit. |
| `max_cost_cents` | Integer | Hard ceiling on total cost, in **US cents** (e.g. `1000` = $10.00). `0`/unset = no limit. |
| `max_wall_time` | Duration | Hard ceiling on **wall-clock** run time (e.g. `30m`, `2h`). `0`/unset = no limit. |
| `stall_timeout` | Duration | **Wall-clock** span with no forward progress before the run aborts and routes through `on_failure` (e.g. `5m`, `90s`). Elapsed time, **not** a turn count. `0`/unset = disabled. |
| `on_resume` | String | Fidelity behavior when a run resumes: `preserve` (keep the checkpoint fidelity level) or `degrade` (downgrade on resume). Only meaningful when `fidelity` is also set. |
| `tool_commands_allow` | String | Comma-separated glob allowlist for `tool` node shell commands (e.g. `git *,make *`). The runtime rejects any tool command not matched by at least one glob. |
| `tool_denylist_add` | String | Comma-separated globs appended to the runtime's default denylist (e.g. `rm -rf *`). Matched commands are refused regardless of `tool_commands_allow`. |

All budget fields use `0` (or unset) to mean **no limit** — `0` does not mean
"zero budget." The three `max_*` fields bound *totals* (monotonic ceilings);
`stall_timeout` bounds *inactivity* (a sliding timer).

---

## Node Definitions

Nodes are defined with `<kind> <ID>` followed by an indented block of fields.

```dippin
  agent Analyze
    label: "Analyze the request"
    prompt:
      You are a senior software architect.
      Review the request carefully.
```

There are **8 node kinds**: `agent`, `human`, `tool`, `parallel`, `fan_in`, `conditional`, `subgraph`, `manager_loop`. Each has its own set of valid fields. See [nodes.md](nodes.md) for full details.

### Agent node: tool_access and writable_paths

Two security-scoped fields are available on `agent` nodes:

- **`tool_access: none`** — strips the LLM's tool catalog for that node. The agent sees no tools and cannot make tool calls. Scoped to the node only; no downstream taint. Other values are linted as DIP139 and fail closed at the runtime.

- **`writable_paths: <glob,glob>`** — bounds where the agent's tools may write, as a comma-separated list of globs resolved against the session root (e.g. `workspace/**,.ai/sprints/**`). Absent means unbounded. A present-but-empty value is a parse error. Malformed values fail closed at the runtime (deny-all / refuse-to-start).

The two fields address different axes — `tool_access` controls *whether* the agent has tools, `writable_paths` controls *where* its tools may write. Setting both `tool_access: none` and `writable_paths` on the same agent (or branch) is dead config and lints as DIP141: with no tools, there is nothing left to bound.

```dippin
  agent Coder
    writable_paths: workspace/**,tmp/**
    prompt:
      Implement the feature, writing only under workspace/ and tmp/.
```

See [nodes.md](nodes.md) for full field details including DIP codes, backend notes, and brace-expansion caveats.

### Parallel and Fan-In shorthand

Parallel and fan-in nodes use a compact inline syntax instead of a block:

```dippin
  parallel FanOut -> TaskA, TaskB, TaskC
  fan_in Join <- TaskA, TaskB, TaskC
```

The `->` and `<-` operators define the fan-out targets and fan-in sources respectively.

### Block form parallel with per-branch overrides

When branches need different configuration, use the block form. Each `branch:` line names a target node and opens an indented block of per-branch overrides:

```dippin
  parallel FanOut
    branch: fast
      model: claude-haiku-4-5
      tool_access: none
    branch: accurate
      model: claude-opus-4-7
      writable_paths: workspace/**
```

Here the `fast` branch overrides `tool_access` to strip tools, while the `accurate` branch keeps its tools but bounds their writes via `writable_paths`. (Don't set both on one branch — see the DIP141 note above.)

The fan-in node still lists the same target IDs as usual:

```dippin
  fan_in Join <- fast, accurate
```

Per-branch overridable fields:

| Field | Description |
|-------|-------------|
| `model` | LLM model for this branch's target agent |
| `provider` | LLM provider for this branch's target agent |
| `fidelity` | Checkpoint fidelity for this branch |
| `tool_access` | Tool-catalog gate (`none` to strip; omit to inherit the target agent's setting) |
| `writable_paths` | Write-scope globs; omit to inherit the target agent's setting — empty never resets to unbounded |

An omitted field inherits the target agent's value. An inline-form parallel (`->`) and a block-form parallel are mutually exclusive on the same node.

---

## Edges Section

The `edges` block defines connections between nodes. Every edge is a line of the form:

```
<FromID> -> <ToID> [when <condition>] [label: <text>] [weight: <int>] [restart: true]
```

### Basic edges

```dippin
  edges
    AskUser -> Interpret
    Interpret -> Validate
    Validate -> Done
```

### Conditional edges

Add `when <expression>` to gate an edge on a runtime condition:

```dippin
  edges
    Validate -> Approve   when ctx.outcome = success
    Validate -> Retry     when ctx.outcome = fail
```

### Edge attributes

| Attribute | Type | Description |
|-----------|------|-------------|
| `when <expr>` | Condition | Boolean guard — edge only traversed if true |
| `label: <text>` | String | Human-readable label (used for human gate choices) |
| `weight: <int>` | Integer | Priority hint — higher wins among competing edges |
| `restart: true` | Boolean | Marks this as a back-edge (loop restart) |

Attributes can be combined on a single line:

```dippin
    Task -> Start when ctx.outcome = fail label: "retry" restart: true
```

See [edges.md](edges.md) for full details on conditions, routing, and restart semantics.

---

## Strings

String values can be:
- **Unquoted**: Simple values without spaces or special characters — `start: MyNode`
- **Double-quoted**: `label: "My Node Label"`
- **Single-quoted** (YAML-style, literal): `marker_grep: '^(green|red)$'`

Double-quoted strings support escape sequences:
- `\"` — literal double quote
- `\\` — literal backslash

Single-quoted strings are literal — backslashes are not escapes. The only
escape is `''`, which produces a single `'`. This makes single quotes
convenient for regular expressions and other values full of backslashes:
`marker_grep: 'it''s a \d+ match'` stores `it's a \d+ match`.

---

## Multiline Blocks

Fields like `prompt` and `command` support multiline content. Write the key followed by `:`, then indent the content on subsequent lines:

```dippin
  agent MyAgent
    prompt:
      You are a code reviewer.

      ## Rules
      - Check for bugs
      - Check for security issues
      - Run `pytest` to validate

      ## Context
      ${ctx.last_response}
```

**How it works:**
- The first content line's indentation sets the baseline
- All content is de-indented by that amount
- Empty lines within the block are preserved
- The block ends when indentation returns to or above the field's level
- No quoting or escaping needed — content is literal text

Tool commands work the same way:

```dippin
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

### File directives (`*_file`)

Instead of an inline multiline block, you can reference an external file. The parser itself stays pure and does **not** read the file; the contents are loaded in a separate post-parse step (`parser.ResolveFileDirectives`) that CLI entry points run after parsing. LSP and WASM consumers skip that step and retain the unresolved directive (the `*_file` field set, content empty).

| Directive | Node kind | Replaces |
|-----------|-----------|---------|
| `prompt_file: <path>` | `agent` | inline `prompt:` block |
| `system_prompt_file: <path>` | `agent` | inline `system_prompt:` block |
| `command_file: <path>` | `tool` | inline `command:` block |

```dippin
  agent Analyze
    prompt_file: prompts/analyze.txt
    system_prompt_file: prompts/system.txt

  tool RunTests
    command_file: scripts/run_tests.sh
```

**Security rules** (enforced by the parser at load time):

- Path is relative to the `.dip` source directory — resolved against its parent directory.
- Absolute paths are rejected (`/etc/passwd` → parse error).
- Parent-tree escape via `..` is rejected lexically (`../secrets` → parse error).
- Symlinks at the final path component are rejected (`O_NOFOLLOW`; atomic on Unix, advisory on other platforms).
- Files larger than **4 MiB** are rejected.
- Specifying both an inline block and a `*_file` directive on the same node is a parse error.

---

## Condition Expressions

Conditions appear on edges after the `when` keyword. They are boolean expressions comparing context variables to values.

### Comparison operators

All operators perform string comparison — there is no numeric coercion.

| Operator | Meaning | Example |
|----------|---------|---------|
| `=`, `==` | String equality | `ctx.outcome = success` |
| `!=` | String inequality | `ctx.outcome != fail` |
| `contains` | Substring match | `ctx.response contains "approved"` |
| `not contains` | Negated substring | `ctx.tool_stdout not contains all-done` |
| `startswith` | Prefix match | `ctx.response startswith "yes"` |
| `endswith` | Suffix match | `ctx.response endswith "done"` |
| `in` | Value in list | `ctx.status in "pass,fail,skip"` |

Infix `not` can negate any operator: `ctx.outcome not contains "error"` is equivalent to `not ctx.outcome contains "error"`.

### Logical operators

Combine comparisons with `and`, `or`, and `not`:

```dippin
    A -> B when ctx.outcome = success and ctx.score = high
    A -> C when ctx.outcome = fail or ctx.status = blocked
    A -> D when not ctx.outcome = success
```

Parentheses control precedence:

```dippin
    A -> B when (ctx.x = 1 or ctx.y = 2) and ctx.z = 3
```

### Context variable namespaces

All variables in conditions use explicit namespaces:

| Namespace | Contents | Examples |
|-----------|----------|----------|
| `ctx.*` | Runtime context (handler outputs) | `ctx.outcome`, `ctx.last_response`, `ctx.tool_stdout` |
| `graph.*` | Workflow-level attributes | `graph.goal`, `graph.name` |
| `params.*` | Subgraph parameters | `params.model`, `params.severity` |

See [context.md](context.md) for the full variable reference.

---

## Identifiers

Identifiers (workflow names, node IDs, field names) may contain:
- Letters (a-z, A-Z)
- Digits (0-9)
- Underscores (`_`)
- Dashes (`-`)
- Dots (`.`)
- Forward slashes (`/`)
- Colons (`:`)

Examples: `AskUser`, `my-task`, `sub.workflow`, `tools/check`

---

## Keywords

The following words have special meaning and cannot be used as identifiers in certain contexts:

`workflow`, `goal`, `start`, `exit`, `defaults`, `agent`, `human`, `tool`, `parallel`, `fan_in`, `subgraph`, `edges`, `when`, `and`, `or`, `not`

---

## Complete Example

```dippin
workflow ask_and_execute
  goal: "Ask user for a task, implement it, review, ship"
  start: AskUser
  exit: Done

  defaults
    model: claude-opus-4-6
    provider: anthropic
    retry_policy: standard

  # ── Phase 1: Gather ──────────────────────────

  human AskUser
    label: "What would you like to build?"
    mode: freeform

  agent Interpret
    label: "Interpret the request"
    reads: human_response
    writes: plan
    prompt:
      You are a senior software architect.
      Read the user's request and produce a plan.

  # ── Phase 2: Implement (parallel) ────────────

  parallel FanOut -> ImplA, ImplB

  agent ImplA
    label: "Implement (Claude)"
    model: claude-opus-4-6
    reads: last_response
    prompt:
      Implement the plan from the previous step.

  agent ImplB
    label: "Implement (GPT)"
    model: gpt-5.4
    provider: openai
    reads: last_response
    prompt:
      Implement the plan from the previous step.

  fan_in Join <- ImplA, ImplB

  # ── Phase 3: Review ──────────────────────────

  agent Validate
    label: "Validate implementation"
    goal_gate: true
    auto_status: true
    max_retries: 2
    prompt:
      Review the implementations. Run tests.
      Respond with STATUS: success or STATUS: fail.

  human Approve
    label: "Ship it?"
    mode: choice
    default: "Yes"

  agent Done
    prompt:
      Pipeline complete.

  # ── Routing ──────────────────────────────────

  edges
    AskUser -> Interpret
    Interpret -> FanOut
    Join -> Validate
    Validate -> Approve      when ctx.outcome = success
    Validate -> Interpret    when ctx.outcome = fail    restart: true
    Approve -> Done
```
