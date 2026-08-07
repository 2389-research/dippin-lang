---
title: "Validation & Linting"
description: "67 diagnostic codes for AI pipeline workflows. 10 structural errors and 57 semantic checks catch bugs before runtime."
section_label: "Diagnostics"
subtitle: "67 diagnostic codes — 10 structural errors and 57 semantic checks — to catch problems before runtime."
---

## Overview

Dippin provides two levels of analysis:

**Structural validation** (DIP001-DIP010): Errors that must be fixed. A workflow with any of these cannot execute. Run with `dippin validate`.

**Semantic linting** (DIP101-DIP157): Warnings that flag likely bugs or questionable patterns. They don't block execution but should be reviewed. Run with `dippin lint` for both levels.

### Diagnostic Format

Diagnostics are displayed in a rustc-inspired format:

```
error[DIP003]: unknown node reference "InterpretX" in edge
  --> pipeline.dip:45:5
  = help: did you mean "Interpret"?
```

## Structural Errors (DIP001-DIP010)

These must be fixed for a workflow to be valid. Each causes exit code 1.

<div class="diag-card error">
  <span class="diag-code">DIP001</span> — Start Node Missing
  <p>The workflow must declare a <code>start:</code> field pointing to an existing node.</p>
  <pre>error[DIP001]: start node does not exist
  --&gt; pipeline.dip:1:1
  = help: add "start: &lt;NodeID&gt;" to the workflow header</pre>
</div>

<div class="diag-card error">
  <span class="diag-code">DIP002</span> — Exit Node Missing
  <p>The workflow must declare an <code>exit:</code> field pointing to an existing node.</p>
  <pre>error[DIP002]: exit node does not exist
  --&gt; pipeline.dip:1:1
  = help: add "exit: &lt;NodeID&gt;" to the workflow header</pre>
</div>

<div class="diag-card error">
  <span class="diag-code">DIP003</span> — Unknown Node Reference in Edge
  <p>Every edge's From and To must reference existing node IDs. The validator uses Levenshtein distance to suggest corrections for typos.</p>
  <pre>error[DIP003]: unknown node reference "InterpretX" in edge
  --&gt; pipeline.dip:45:5
  = help: did you mean "Interpret"?</pre>
</div>

<div class="diag-card error">
  <span class="diag-code">DIP004</span> — Unreachable Node from Start
  <p>Every node must be reachable from the start node via some path of edges. BFS from the start node cannot reach this node.</p>
  <pre>error[DIP004]: node unreachable from start
  --&gt; pipeline.dip:20:3
  = help: add an edge leading to this node, or remove it</pre>
</div>

<div class="diag-card error">
  <span class="diag-code">DIP005</span> — Unconditional Cycle Detected
  <p>The workflow graph must be a DAG, with the exception of restart edges. A back-edge not marked <code>restart: true</code> would loop forever.</p>
  <pre>error[DIP005]: unconditional cycle detected
  --&gt; pipeline.dip:50:5
  = help: remove an edge in this cycle or mark it "restart: true"</pre>
</div>

<div class="diag-card error">
  <span class="diag-code">DIP006</span> — Exit Node Has Outgoing Edges
  <p>The exit node is the terminal — it must have zero outgoing edges.</p>
  <pre>error[DIP006]: exit node has outgoing edges
  --&gt; pipeline.dip:55:5
  = help: remove outgoing edges from the exit node</pre>
</div>

<div class="diag-card error">
  <span class="diag-code">DIP007</span> — Parallel/Fan-In Mismatch
  <p>Every <code>parallel</code> node must have a matching <code>fan_in</code> node with the same set of branch nodes.</p>
  <pre>error[DIP007]: parallel fan-out/fan-in mismatch
  --&gt; pipeline.dip:15:3
  = help: add a matching fan_in node</pre>
</div>

<div class="diag-card error">
  <span class="diag-code">DIP008</span> — Duplicate Node ID
  <p>Node IDs must be globally unique within a workflow.</p>
  <pre>error[DIP008]: duplicate node ID
  --&gt; pipeline.dip:30:3
  = help: rename this node or remove the duplicate</pre>
</div>

<div class="diag-card error">
  <span class="diag-code">DIP009</span> — Duplicate Edge
  <p>No two edges may have the same (from, to, condition) combination. Edges with different conditions on the same pair are not duplicates.</p>
  <pre>error[DIP009]: duplicate edge
  --&gt; pipeline.dip:60:5
  = help: remove the duplicate edge</pre>
</div>

<div class="diag-card error">
  <span class="diag-code">DIP010</span> — Unparseable Edge Condition
  <p>Every edge <code>when</code> condition must parse into a valid expression. An unparseable condition — an unknown operator, or a tool-node field like <code>marker_grep</code> used in operator position — leaves the edge's routing undefined, so the workflow cannot execute. One diagnostic fires per bad edge; every parseable edge is still checked.</p>
  <pre>error[DIP010]: edge A -&gt; Z: invalid condition "marker_grep \"^ok\"": unknown operator "^ok"
  --&gt; pipeline.dip:14:5
  = help: valid operators: = == != contains startswith endswith in</pre>
</div>

## Semantic Warnings (DIP101-DIP157)

These flag likely bugs or questionable patterns. Warnings alone exit 0.

<div class="diag-card warning">
  <span class="diag-code">DIP101</span> — Node Only Reachable via Conditional Edges
  <p>All incoming edges have conditions — if none match, execution can never reach this node. Automatically suppressed when source nodes have exhaustive conditions (e.g., success/fail pairs).</p>
  <pre>warning[DIP101]: node "NextPhase" is only reachable through conditional edges
  --&gt; pipeline.dip:25:3
  = help: add an unconditional edge, or verify conditions are exhaustive</pre>
</div>

<div class="diag-card warning">
  <span class="diag-code">DIP105</span> — No Success Path to Exit
  <p>There is no guaranteed path from start to exit through unconditional edges alone. If conditions don't match, execution may never reach the exit.</p>
  <pre>warning[DIP105]: no success path from start to exit
  --&gt; pipeline.dip:1:1</pre>
</div>

<div class="diag-card warning">
  <span class="diag-code">DIP108</span> — Unknown Model/Provider
  <p>The model or provider isn't in the recognized list. Verified against official provider documentation.</p>
  <pre>warning[DIP108]: unknown model/provider combination
  --&gt; pipeline.dip:15:5</pre>
</div>

<div class="diag-card warning">
  <span class="diag-code">DIP110</span> — Empty Prompt on Agent
  <p>An agent node has no prompt text. An agent without a prompt won't produce meaningful output.</p>
  <pre>warning[DIP110]: empty prompt on agent node
  --&gt; pipeline.dip:12:3</pre>
</div>

<div class="diag-card warning">
  <span class="diag-code">DIP111</span> — Tool Without Timeout
  <p>A tool node has no <code>timeout</code> field. Without a timeout, a hanging command blocks the pipeline indefinitely.</p>
  <pre>warning[DIP111]: tool command has no timeout
  --&gt; pipeline.dip:35:3</pre>
</div>

<div class="diag-card warning">
  <span class="diag-code">DIP115</span> — Goal Gate Without Recovery Path
  <p>A node has <code>goal_gate: true</code> but no failure route — no <code>on fail</code> edge, and (in v1) no <code>retry_target</code>/<code>fallback_target</code> — so the pipeline has no recovery path if the gate fails.</p>
  <pre>warning[DIP115]: node "validate_tests" has goal_gate but no recovery path
  --&gt; pipeline.dip:18:3
  = help: add an on fail edge (dip 2), or retry_target / fallback_target (v1)</pre>
</div>

<div class="diag-card warning">
  <span class="diag-code">DIP119</span> — Invalid reasoning_effort Value
  <p>The <code>reasoning_effort</code> field on an agent node has an unrecognized value. Valid levels: <code>none</code>, <code>minimal</code>, <code>low</code>, <code>medium</code>, <code>high</code>, <code>xhigh</code>, <code>max</code>. Not all providers support all levels.</p>
  <pre>warning[DIP119]: node "Analyze" has reasoning_effort "extreme" which is not a recognized level
  --&gt; pipeline.dip:12:5
  = help: valid levels: none, minimal, low, medium, high, xhigh, max</pre>
</div>

<div class="diag-card warning">
  <span class="diag-code">DIP123</span> — Tool Command Syntax Error
  <p>The tool command block has a shell syntax error detectable by <code>bash -n</code> — unclosed quotes, bad redirects, missing <code>fi</code>/<code>done</code>.</p>
  <pre>warning[DIP123]: tool command has shell syntax error:
  unexpected EOF while looking for matching `"'
  --&gt; pipeline.dip:45:5</pre>
</div>

<div class="diag-card warning">
  <span class="diag-code">DIP124</span> — Runtime Variable in Tool Command
  <p>A tool command contains <code>${ctx.*}</code> interpolation. These are Dippin runtime variables that expand to empty strings in the shell.</p>
  <pre>warning[DIP124]: tool command references ${ctx.api_url}
  which expands to empty at runtime
  --&gt; pipeline.dip:50:5</pre>
</div>

<div class="diag-card warning">
  <span class="diag-code">DIP125</span> — Binary Not Found
  <p>The first command in the tool block references a binary not on the current PATH. This is a hint — the deployment environment may differ.</p>
  <pre>hint[DIP125]: tool command binary "npx" not found on PATH
  --&gt; pipeline.dip:55:5</pre>
</div>

<div class="diag-card warning">
  <span class="diag-code">DIP126</span> — Subgraph ref file does not exist
  <p>A subgraph node's <code>ref:</code> path does not resolve to an existing file on disk.</p>
</div>

<div class="diag-card warning">
  <span class="diag-code">DIP127</span> — Invalid human node mode
  <p>The <code>mode:</code> value on a human node must be <code>choice</code>, <code>freeform</code>, <code>interview</code>, or <code>yes_no</code>.</p>
  <pre>warning[DIP127]: invalid human node mode "dialog"
  --&gt; pipeline.dip:22:5
  = help: use one of: choice, freeform, interview, yes_no</pre>
</div>

<div class="diag-card warning">
  <span class="diag-code">DIP128</span> — Interview mode with meaningless default
  <p>A human node with <code>mode: interview</code> has a <code>default:</code> value, which has no effect in interview mode — answers come from the extracted question list.</p>
  <pre>warning[DIP128]: human node "GatherRequirements" has mode: interview with a default value
  --&gt; pipeline.dip:28:5
  = help: remove the default field — it is ignored in interview mode</pre>
</div>

<div class="diag-card warning">
  <span class="diag-code">DIP129</span> — Interview mode with choice-style labeled edges
  <p>A human node with <code>mode: interview</code> has outgoing edges with choice labels, which conflict — interview mode collects freeform answers, not discrete choices.</p>
  <pre>warning[DIP129]: human node "GatherRequirements" uses interview mode but has choice-style edges
  --&gt; pipeline.dip:30:5
  = help: remove edge labels or switch to mode: choice</pre>
</div>

<div class="diag-card warning">
  <span class="diag-code">DIP130</span> — Invalid response_format value
  <p>The <code>response_format:</code> field on an agent node must be <code>json_object</code> or <code>json_schema</code>. Any other value is unrecognised and will be rejected at runtime.</p>
  <pre>warning[DIP130]: invalid response_format "xml" on agent "Parse"
  --&gt; pipeline.dip:18:5
  = help: use "json_object" or "json_schema"</pre>
</div>

<div class="diag-card warning">
  <span class="diag-code">DIP131</span> — response_schema / response_format mismatch
  <p>Two related hints: (1) if <code>response_schema:</code> is set but <code>response_format:</code> is not <code>json_schema</code>, the schema will be ignored; (2) if <code>response_format: json_schema</code> is set but no <code>response_schema:</code> is provided, the model receives no schema to enforce.</p>
  <pre>warning[DIP131]: response_schema set but response_format is not json_schema
  --&gt; pipeline.dip:20:5
  = help: set response_format: json_schema or remove response_schema</pre>
</div>

<div class="diag-card warning">
  <span class="diag-code">DIP132</span> — response_schema is not valid JSON
  <p>The value of <code>response_schema:</code> must be a valid JSON object. Malformed JSON will cause a runtime error when the model attempts to apply structured output.</p>
  <pre>warning[DIP132]: response_schema on agent "Extract" is not valid JSON
  --&gt; pipeline.dip:22:5
  = help: fix the JSON syntax in response_schema</pre>
</div>

<div class="diag-card warning">
  <span class="diag-code">DIP133</span> — params key shadows a first-class field
  <p>A key inside the agent's <code>params:</code> block (e.g. <code>model</code> or <code>provider</code>) duplicates a first-class field on the same node. The first-class field takes precedence; the params entry is silently ignored.</p>
  <pre>hint[DIP133]: params key "model" shadows the first-class model field on agent "Analyze"
  --&gt; pipeline.dip:35:7
  = help: remove the params key and set the field directly</pre>
</div>

### DIP134 — max_retries set with restart edges but no max_restarts

**Severity:** Warning

Fires when `max_retries` is set in defaults and the workflow has `restart: true` edges, but `max_restarts` is not set. These are commonly confused: `max_retries` controls per-node LLM retries, while `max_restarts` controls the global loop restart budget.

**Fix:** Set `max_restarts` in defaults to control loop iterations, or add it alongside `max_retries` if both behaviors are intended.

### DIP143 — Referenced subgraph does not inherit `tool_access`

**Severity:** Hint

A `manager_loop` (`subgraph_ref`) or `subgraph` (`ref`) node references a child `.dip` while this workflow declares a `tool_access` restriction — and `tool_access` does not cross the file boundary. The wasm/playground lint cannot read the child, so it reminds you to give the child's agents their own `tool_access`. (For the resolved cross-file check that *reads* the child, see DIP146.)

```text
hint[DIP143]: manager_loop "Supervise" references subgraph "child.dip", defined in its own file; this workflow's tool_access restrictions do not extend across the subgraph boundary
```

The native `dippin lint` cross-file pass also emits DIP143 for a **deep** (depth ≥ 1) child that is either **partial-audit** (resolved — some agents restricted, at least one tool-bearing agent without a `tool_access` of its own) or **unresolvable** (missing, unparseable, or refused) — boundaries the entry-only lint never sees. Like DIP146, it fires only when a workflow on the path declares `tool_access`.

### DIP144 — Agent node has no failure route

**Severity:** Warning

An agent node has no way to handle failure — no `ctx.outcome = fail` edge, no `fallback_target`, no bounded retry (`retry_target` + `max_retries`), and no graph-level `on_failure`. Any one of those suppresses it; an unconditional/success edge does not.

```text
warning[DIP144]: agent node "Build" has no failure route (no fail edge, no fallback_target, no bounded retry, no graph on_failure)
  = help: add `-> <node> when ctx.outcome = fail`, set fallback_target:, add retry_target with max_retries, or declare a workflow-level on_failure:
```

### DIP145 — Negative graph budget default

**Severity:** Warning

A graph budget default (`max_total_tokens`, `max_cost_cents`, `max_wall_time`, or `stall_timeout`) is set to a negative value. Budgets are non-negative; `0` (or unset) means **no limit**.

```text
warning[DIP145]: workflow budget default max_cost_cents is -5; budgets cannot be negative
```

### DIP146 — Child subgraph re-grants restricted tools (cross-file)

**Severity:** Hint

`dippin lint` resolves a `manager_loop` (`subgraph_ref`) or `subgraph` (`ref`) child across the file boundary and finds it declares **no** `tool_access` restriction on any agent, while a workflow on the path from the linted entry restricts tools. Unlike DIP143 (which cannot read the child), DIP146 reads and confirms the child and traverses transitively. The traversal is intent-aware: a child reached through both a no-intent path and a restricting path is still checked under the restricting path. Detection only — dippin flags the gap; runtime enforcement is separate.

```text
hint[DIP146]: manager_loop "Supervise" delegates to subgraph "worker.dip", which declares no tool_access restriction on any agent; a workflow on this path restricts tools, but the restriction does not cross the subgraph boundary
```

### DIP147 — Restricted agent's output reaches a privileged prompt (chain attack)

**Severity:** Hint

A `tool_access: none` agent declares a context key in `writes:`, and a downstream tool-bearing agent (one that omits `tool_access`, so it holds the full catalog) declares that same key in `reads:`. `tool_access` bounds an agent's *tools*, not the *information* its output carries — so the restricted agent's output (potentially a laundered injection payload) reaches a privileged agent's prompt, re-granting capability the upstream node was denied. dippin flags the explicitly-keyed handoff the author wired by hand; the runtime enforces the bound. Detection only.

```text
hint[DIP147]: restricted agent "Summarize" (tool_access: none) writes context key "summary" that tool-bearing agent "Act" reads — its output reaches a privileged prompt
```

### DIP148 — Negative `last_response_truncate`

**Severity:** Warning

An agent node (or a parallel-branch override) sets `last_response_truncate` to a negative value. The field is a character cap on the prior node's response before it is auto-injected into this agent's prompt (a mitigation for the `${ctx.last_response}` flow); a negative cap is meaningless. On an agent, `0` (or unset) means no truncation; on a parallel-branch override, `0` (or unset) inherits the target agent's cap. Detection only — dippin carries and lints the field; a runtime enforces the truncation.

```text
warning[DIP148]: agent "Writer" last_response_truncate is -1; cannot be negative
```

### DIP149 — Ambiguous routing

**Severity:** Warning

A node has two or more **unconditional** (no `when`) outgoing forward edges. Both are equally eligible, so which one fires is decided only by the routing cascade's lexical tiebreak — the alphabetical order of the target node ID. That is silent action-at-a-distance: renaming a target node can change which edge fires with no other change. Restart/`loop` back-edges are excluded, and `parallel` fan-out and `human` choice gates are exempt. Keep one unconditional edge as the default fallback and guard the others with a `when` condition.

```text
warning[DIP149]: node "Route" has multiple unconditional outgoing edges; which one fires is decided only by the lexical tiebreak
```

### DIP150 — Human gate routes by label

**Severity:** Hint

An outgoing edge from a **label-routing** `human` node (mode `choice`, `yes_no`, or the unset default) sets a non-empty `label:` but no `choice:`. In Phase 0 the label is overloaded — besides being display text, it is the routing key the runtime matches the user's selection against, so nothing in the syntax signals that the label is load-bearing. Add `choice: "<key>"` to mark the routing key explicitly, leaving `label:` for display. `choice:` wins when present; `label:` still routes when it is absent, so existing workflows route unchanged. Does not fire when `choice:` is already set, for non-`human` source nodes, or for `freeform`/`interview` gates (which route by text/answers, not labels).

```text
hint[DIP150]: human gate "Approve" routes by label "yes"; use choice: "yes" to mark the routing key (label: stays for display)
```

### DIP151 — Edge weight is unused by routing

**Severity:** Warning

An edge carries a `weight:` attribute. `weight:` was tier 4 of the routing cascade — a speculative priority hint — but the cascade never consults it and no real workflow uses it. It still **parses** (carry-only, no breakage in Phase 0), but both the keyword and the cascade tier are slated for removal under `dip 2`. Remove `weight:` and express priority with conditions instead — guard edges with `when` / `on`, or rely on a single unconditional fallback.

```text
warning[DIP151]: edge "Route" -> "A" sets weight: 5, which routing does not use; guard edges with when / on or rely on a single unconditional fallback instead
```

### DIP152 — marker_grep enumerates an unrouted marker

**Severity:** Warning

A tool node's `marker_grep` enumerates a literal marker that no outgoing edge routes and that no section `else ->` default or unconditional fallback edge covers — so that marker would be emitted at runtime with nowhere to go. Only checked when `marker_grep` is a recognizable literal alternation (e.g. `^(a|b|c)$` or a bare literal); complex regexes are left unflagged, and any compound (`or`), negated (`!=` / `not`), or other-variable edge makes the node safe. Fix by routing the marker with an edge (`on <marker>`), adding an unconditional fallback edge, or adding a section `else -> <node>` default.

```text
warning[DIP152]: tool node "RunTests" emits markers that no edge routes and no else default covers: tests-failed
```

### DIP153 — redundant parallel/fan_in edge

**Severity:** Warning

An `edges` block declares an unconditional, attribute-free edge that merely repeats a fork already declared inline on a `parallel Fan -> A, B` or `fan_in Join <- A, B` node. The inline list is the single source of truth — validation, simulation, and DOT export all derive the fan edges from it — so the re-declaration conveys nothing new and must be kept in sync by hand. A *conditional* or *attributed* edge (`when`, `label:`, `choice:`, `weight:`) between the same nodes is real routing, not a duplicate, and is left untouched. Fix by removing the redundant edge — `dippin fmt` strips it automatically; under a `dip 2` header the re-declaration is rejected outright.

```text
warning[DIP153]: edges-block edge 'Fan -> A' redundantly repeats the inline parallel/fan_in fork; the inline list is authoritative — run 'dippin fmt' to remove it (rejected under 'dip 2')
```

### DIP154 — prompt-cascade opt-out is a no-op

**Severity:** Hint

An agent sets `prompt_prefix: none` or `prompt_suffix: none` to opt out of the defaults prompt cascade, but the `defaults` block declares no cascade of that kind, so the opt-out does nothing. Fix by removing the unnecessary opt-out or adding the intended cascade to `defaults`.

```text
hint[DIP154]: agent "A" sets prompt_suffix: none but no defaults prompt_suffix cascade is declared — the opt-out is a no-op
```

> **Full catalog:** This page highlights the most common diagnostics. For every code (DIP001–DIP010, DIP101–DIP157) with full descriptions, run `dippin explain <code>` or see the [generated language spec](https://github.com/2389-research/dippin-lang/blob/main/cmd/dippin/generated-spec.md). Codes DIP135–DIP142 are documented there.
