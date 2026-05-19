# Dippin LLM Reference Card

Compact reference for LLMs generating `.dip` workflow files. Paste into system prompts or tool descriptions.

---

## Grammar (simplified BNF)

```
workflow <Name>
  goal: "<text>"
  [requires: <dep1>, <dep2>, ...]
  start: <NodeID>
  exit: <NodeID>

  [defaults
    model: <string>
    provider: <string>
    max_total_tokens: <int>
    max_cost_cents: <int>
    max_wall_time: <duration>
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
    <From> -> <To> [when <condition>] [label: <text>] [weight: <int>] [restart: true]
```

---

## Node Kinds

| Kind | Required Fields | Optional Fields |
|------|----------------|-----------------|
| `agent` | `prompt` | `model`, `provider`, `backend`, `working_dir`, `auto_status`, `goal_gate`, `reasoning_effort`, `fidelity`, `max_turns`, `system_prompt` |
| `human` | `mode` (freeform\|choice\|interview) | `default`, `timeout` (duration, e.g. 5m), `timeout_action` (string: fail\|default) |
| `tool` | `command` | `timeout` (e.g. 30s, 5m), `outputs` (CSV), `marker_grep` (regex), `route_required` (bool), `output_limit` (bytes) |
| `parallel` | `-> Target1, Target2` (inline) | — |
| `fan_in` | `<- Source1, Source2` (inline) | — |
| `subgraph` | `ref` | `params` |

All kinds also accept: `label`, `reads`, `writes`, `retry_policy`, `max_retries`, `base_delay`, `retry_target`, `fallback_target`.

---

## Edge Conditions

```
when <variable> <op> <value>
when <expr> and <expr>
when <expr> or <expr>
when not <expr>
```

**Comparison operators:** `=`, `==`, `!=`, `contains`, `not contains`, `startswith`, `endswith`, `in` (all string comparison, no numeric ops)

**Variables:** Always namespace-qualified: `ctx.outcome`, `ctx.status`, `graph.goal`

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
| 9 | Hand-parsing tool stdout for routing | Use `marker_grep: "<regex>"` (and optionally `route_required: true`) instead of regexing `ctx.tool_stdout` in edge conditions. Mirrors tracker's TRK101 recommendation; populates `ctx.tool_marker` directly. |
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

**Contextual keywords** (not reserved — usable as node IDs): `workflow`, `agent`, `human`, `tool`, `subgraph`, `parallel`, `fan_in`, `edges`, `defaults`, `when`, `and`, `or`, `not`, `true`, `false`, `restart`, `label`, `weight`.

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

Workflow: author and lint as `.dip`; package with `dippin pack` for distribution to runtimes (e.g., Tracker). `dippin check pipeline.dipx` validates the bundled entry workflow exactly as if it were on disk. Bundle commands return distinct exit codes (`0` ok, `1` user error, `2` integrity error, `3` I/O error, `4` cancelled) so tooling can disambiguate failures that the analysis-command `0/1/2` ladder collapses.

---

## Diagnostic Code Summary

39 diagnostic codes across two categories:

- **DIP001–DIP009** (errors): start/exit missing, unknown refs, unreachable nodes, cycles, duplicates, parallel/fan_in mismatch
- **DIP101–DIP133** (warnings): conditional reachability, missing defaults, overlapping conditions, unbounded retries, undefined variables, unknown models, empty prompts, missing timeouts, invalid policy/fidelity/reasoning_effort, stylesheet refs, namespace prefixes, condition type checking, structured output validation
