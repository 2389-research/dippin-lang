---
name: dippin-lang
description: Use when working with .dip workflow files, the dippin CLI, AI pipeline authoring, or debugging DIP diagnostic codes
---

# dippin-lang

DSL and toolchain for AI pipeline workflows. `.dip` files define directed graphs of agent steps, human gates, tool calls, and control-flow constructs.

## Quick Start

```
go install github.com/2389-research/dippin-lang/cmd/dippin@latest
```

Generate from template: `dippin new minimal`, `dippin new parallel`, `dippin new conditional`, `dippin new review-loop`, `dippin new human-gate`

Write with `--name` and `--write`: `dippin new --name MyPipeline --write pipeline.dip review-loop`

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

  <node declarations>

  edges
    <edge declarations>
```

Four sections in order: **header** (workflow name, goal, optional `requires`, start, exit) → **defaults** (optional) → **nodes** (any order) → **edges** (optional). `defaults` and `edges` are bare keywords — no colon after them. `requires:` is a comma-separated list of environmental dependencies (e.g. `git, docker, jq`); semantics live in downstream consumers and unknown entries are accepted without a parser diagnostic.

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
| `system_prompt` | multiline | System message |
| `model` | string | Must be valid model ID (DIP108) |
| `provider` | string | anthropic, openai, google, deepseek, xai, mistral, cohere |
| `backend` | string | Per-node backend override (e.g., `native`, `claude-code`, `acp`) |
| `working_dir` | string | Per-node working directory override for isolated execution. |
| `max_turns` | int | Max conversation turns |
| `cmd_timeout` | duration | e.g. `30s`, `5m` |
| `auto_status` | bool | Parses `STATUS: success/fail` → `ctx.outcome` |
| `goal_gate` | bool | Pipeline fails if gate fails. Add `retry_target` or `fallback_target` (DIP115) |
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

### human — user decision gate

```
  human Approve
    mode: choice
```

| Field | Type | Notes |
|-------|------|-------|
| `mode` | string | **Required.** `choice`, `freeform`, or `interview` (DIP127) |
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
| `command` | multiline | Shell command. Supports pipes, here-docs, case/esac. |
| `timeout` | duration | **Required** (DIP111). e.g. `30s`, `5m` |
| `outputs` | CSV | Possible stdout values for condition checks |
| `marker_grep` | string | Regex matched against stdout; sets `ctx.tool_marker`. Tracker validates at runtime. |
| `route_required` | bool | When true, fails the node if no `_TRACKER_ROUTE=<value>` sentinel line is emitted. |
| `output_limit` | int | Per-node stdout byte cap (non-negative integer); 0 (or omitted) uses the engine default. |
| `reads` | CSV | Context keys read |
| `writes` | CSV | Context keys written |

Do NOT use `${ctx.*}` in commands — they expand to empty at parse time (DIP124). Output is captured as `ctx.tool_stdout` and `ctx.tool_stderr`.

### parallel / fan_in — concurrent execution

```
  parallel FanOut -> WorkerA, WorkerB, WorkerC
  fan_in Merge <- WorkerA, WorkerB, WorkerC
```

Inline syntax only. Every `parallel` must have a matching `fan_in` with identical target/source sets (DIP007). Wire edges from each target to the `fan_in` node in the `edges` block. All targets execute concurrently with independent context copies.

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

Spawns a child `.dip` pipeline, polls it on a cadence, and can steer it by injecting context. Maps to `stack.manager_loop` in Tracker; DOT shape `house`. Full reference: [docs/nodes.md](https://github.com/2389-research/dippin-lang/blob/main/docs/nodes.md).

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
| `retry_target` | Node ID to jump to on retry |
| `fallback_target` | Node ID if retries exhausted |

**Every node must have at least one field.** An empty node body causes a parse error.

## Edges

```
  edges
    Start -> Analyze
    Analyze -> Decide
    Decide -> Merge when ctx.outcome = success
    Decide -> Revise when ctx.outcome = fail
    Revise -> Analyze restart: true
```

| Attribute | Syntax | Notes |
|-----------|--------|-------|
| condition | `when <expr>` | Guard expression |
| label | `label: <text>` | Display text / human choice button |
| weight | `weight: <int>` | Priority (higher wins) |
| restart | `restart: true` | **Required on back-edges** to avoid DIP005 (unconditional cycle) |

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
    tool_commands_allow: "git *,make *"
    tool_denylist_add: "rm -rf /,dd *"
```

All defaults are inherited by nodes unless overridden at the node level.

| Default Field | Type | Notes |
|---------------|------|-------|
| `max_total_tokens` | int | Budget cap on total tokens consumed. |
| `max_cost_cents` | int | Budget cap in cents (e.g. 1000 = $10.00). |
| `max_wall_time` | duration | Maximum wall-clock time for the workflow (e.g. `30m`, `2h`). |
| `tool_commands_allow` | string | Glob allowlist for tool-node shell commands (comma-separated; optional). |
| `tool_denylist_add` | string | Glob patterns appended to tracker's default denylist (comma-separated; optional). |

## CLI Reference

**Global flag:** `--format text|json` — must come **before** the subcommand (e.g. `dippin --format json check file.dip`).

Use `dippin help` (not `--help`) to see all commands.

### Authoring

| Command | Purpose |
|---------|---------|
| `dippin parse <file>` | Output IR as JSON |
| `dippin validate <file>` | Structural checks only (DIP001-DIP009) |
| `dippin lint <file>` | Full validation + semantic warnings (DIP001-DIP137) |
| `dippin check <file>` | All-in-one. JSON output by default — **use this for automated workflows** |
| `dippin fmt <file>` | Print canonical format to stdout |
| `dippin fmt --check <file>` | Exit 1 if not formatted |
| `dippin fmt --write <file>` | Rewrite file in place |
| `dippin new <template>` | Generate from template: `minimal`, `parallel`, `conditional`, `review-loop`, `human-gate` |

### Export

| Command | Purpose |
|---------|---------|
| `dippin export-dot <file>` | Export to Graphviz DOT |
| `dippin export-dip <file>` | Export flattened .dip (resolves subgraph refs) |
| `dippin migrate <file.dot>` | Convert DOT to .dip |
| `dippin validate-migration <old.dot> <new.dip>` | Verify migration parity |

### Analysis

| Command | Purpose |
|---------|---------|
| `dippin simulate <file>` | Dry-run (JSONL events). `--scenario key=val` to inject context. `--all-paths` for exhaustive |
| `dippin cost <file>` | Estimate execution cost by model/provider. Requires model/provider on nodes or in defaults |
| `dippin coverage <file>` | Edge coverage and reachability |
| `dippin doctor <file>` | Health report card (grade A-F) |
| `dippin optimize <file>` | Suggest cheaper model substitutions |
| `dippin diff <file1> <file2>` | Semantic diff between two workflows |
| `dippin feedback <file>` | Compare predicted vs actual costs |
| `dippin explain <code>` | Explain a diagnostic code (e.g. `dippin explain DIP005`) |
| `dippin unused <file>` | Detect dead-branch nodes and wasted cost |
| `dippin graph <file>` | Render ASCII DAG of the workflow |
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

**Recommended workflow:** author and lint as `.dip`, package with `dippin pack` for distribution to runtimes (e.g. Tracker).

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
| DIP005 | Unconditional cycle | Add `restart: true` to the back-edge |
| DIP006 | Exit has outgoing edges | Remove edges from exit node or change exit to a different node |
| DIP007 | Parallel/fan_in mismatch | Add matching `fan_in` node with identical target set, and wire edges from each target to the fan_in node |
| DIP008 | Duplicate node ID | Rename one of the duplicate nodes |
| DIP009 | Duplicate edge | Remove the duplicate. Uniqueness is determined by `(source, target)` pair — two edges to the same target with different labels are still duplicates |

### Semantic Warnings (should fix)

| Code | Issue | Fix |
|------|-------|-----|
| DIP101 | Node only reachable via conditionals | Add unconditional fallback edge or make conditions exhaustive (`success`/`fail`) |
| DIP102 | No default edge from routing node | Add unconditional edge or exhaustive conditions |
| DIP103 | Overlapping conditions | Disambiguate condition expressions |
| DIP104 | Unbounded retry loop | Add `max_retries` or `fallback_target` |
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
| DIP115 | Goal gate without recovery | Add `retry_target` or `fallback_target` |
| DIP116 | Invalid compaction threshold | Use float 0.0-1.0 |
| DIP117 | Stylesheet class references undefined class | Fix class name in stylesheet block |
| DIP118 | Stylesheet ID references unknown node | Fix node ID in stylesheet block |
| DIP119 | Invalid reasoning_effort | Use: `none`, `minimal`, `low`, `medium`, `high`, `xhigh`, `max` |
| DIP120 | Condition var missing namespace | Prefix with `ctx.`, `params.`, or `graph.` |
| DIP123 | Shell syntax error in command | Fix the shell command (checked via `bash -n`) |
| DIP124 | `${ctx.*}` in tool command | Remove — runtime variables expand to empty at parse time |
| DIP125 | Command binary not on PATH | Install the binary or fix the command |
| DIP126 | Subgraph ref file missing | Check `ref:` path |
| DIP127 | Invalid human mode | Use: `choice`, `freeform`, `interview` |
| DIP130 | Invalid response_format | Use: `json_object`, `json_schema` |
| DIP131 | Schema/format mismatch | `response_schema` requires `response_format: json_schema`, and vice versa — both must be present together |
| DIP132 | Invalid JSON in response_schema | Fix the JSON syntax |
| DIP133 | Params key shadows field | Rename the params key |

## Best Practices

- **Always set `timeout`** on tool nodes — no timeout means infinite hang
- **Prefer `marker_grep:`** over regexing `ctx.tool_stdout` in edges when the runtime supports it. Typed routing leaves stdout free for diagnostic output and avoids truncation foot-guns. Declaring `marker_grep:` also suppresses DIP101/DIP102 on the source node — the validator treats it as a safe typed-routing channel.
- **Boolean fields** (`goal_gate`, `auto_status`, `cache_tools`, `route_required`) accept `true/false`, `1/0`, `yes/no`, `on/off` case-insensitively. Anything else is a parse diagnostic.
- **Use `auto_status: true`** on agent nodes that drive conditional routing via `ctx.outcome`
- **Use `success`/`fail`** as condition values — the linter recognizes these as exhaustive
- **Mark back-edges `restart: true`** — loops without it trigger DIP005
- **Declare `reads`/`writes`** on nodes to document data flow (enables DIP107/DIP112 checks)
- **Add `retry_target` or `fallback_target`** to `goal_gate: true` nodes
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
    Review -> Implement when ctx.outcome = fail restart: true
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
| `ctx.tool_route` | `_TRACKER_ROUTE=<value>` sentinel (when `route_required: true`) |
| `ctx.preferred_label` | Human choice selection (maps to edge label) |
| `ctx.interview_answers` | Interview mode answers (via `answers_key`) |
| `params.<key>` | Parent subgraph params |
| `graph.<field>` | Workflow-level metadata |

## Documentation

- [Language Reference](https://2389-research.github.io/dippin-lang/language.html)
- [CLI Reference](https://2389-research.github.io/dippin-lang/cli.html)
- [Validation & Linting](https://2389-research.github.io/dippin-lang/validation.html)
- [Scenario Testing](https://2389-research.github.io/dippin-lang/testing.html)
- [Analysis Tools](https://2389-research.github.io/dippin-lang/analysis.html)
- [Playground](https://2389-research.github.io/dippin-lang/playground.html)
- [GitHub](https://github.com/2389-research/dippin-lang)
