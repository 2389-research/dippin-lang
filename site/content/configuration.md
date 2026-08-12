---
title: "Configuration"
description: "Reference for the defaults block and graph-level configuration: models, retry and recovery, budgets, context management, and tool safety."
section_label: "Configuration"
subtitle: "The defaults block and graph-level configuration."
---

## The Defaults Block

The optional `defaults` block sets graph-level configuration that applies to all nodes unless overridden at the node level. It comes after the workflow header and before the `vars` block.

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

**Inheritance model.** Every field in `defaults` is a fallback: a node that does not declare the corresponding field inherits the graph-level value. A node that *does* declare it overrides the default for that node only. See [Node-Level Overrides](#node-level-overrides) for the precedence rules.

The full set of defaults fields:

| Field | Type | Description |
|-------|------|-------------|
| `model` | String | Default LLM model for all agent nodes |
| `provider` | String | Default LLM provider (e.g., "openai", "anthropic") |
| `prompt_prefix` / `prompt_suffix` | String | Inline prompt fragment cascaded to every agent as a prefix/suffix (#175) |
| `prompt_prefix_file` / `prompt_suffix_file` | String | Prompt fragment loaded from a file and cascaded to every agent (mutually exclusive with the inline form) |
| `system_prompt_file` | String | Shared system prompt (persona) loaded from a file; a fallback default for agents that set none of their own — a node's own `system_prompt`/`system_prompt_file` overrides it (#72). File form only |
| `retry_policy` | String | Default retry strategy name |
| `max_retries` | Integer | Default max retry attempts per node |
| `fidelity` | String | Default checkpoint fidelity level |
| `max_restarts` | Integer | Max loop restarts before pipeline failure (default: 5) |
| `restart_target` | String | Node ID to jump to on restart loops |
| `cache_tools` | Boolean | Whether to cache tool call results |
| `compaction` | String | Context compaction mode for long pipelines |
| `on_resume` | String | Fidelity behavior when a run resumes: `preserve` or `degrade`. Only meaningful when `fidelity` is also set. |
| `max_total_tokens` | Integer | Hard ceiling on total tokens across the run. `0`/unset = no limit. |
| `max_cost_cents` | Integer | Hard ceiling on total cost, in US cents (e.g. `1000` = $10.00). `0`/unset = no limit. |
| `max_wall_time` | Duration | Hard ceiling on wall-clock run time (e.g. `30m`, `2h`). `0`/unset = no limit. |
| `stall_timeout` | Duration | Abort/route when no forward progress is made for a wall-clock span (e.g. `30s`, `5m`); `0`/unset = no limit. |
| `on_failure` | NodeID | Graph-level catch-all failure route. |
| `tool_commands_allow` | CSV (globs) | Comma-separated glob allowlist for tool-node commands. |
| `tool_denylist_add` | CSV (globs) | Comma-separated globs appended to the runtime's default denylist. |

## Model & Provider Defaults

`model` sets the default LLM model for all agent nodes, and `provider` sets the default LLM provider (e.g. `anthropic`, `openai`). Any agent node can override either with its own `model` / `provider` field.

For the catalog of recognized model IDs, per-provider pricing, and the DIP108 staleness check, see **Models & Pricing**.

The prompt-cascade defaults are also model-adjacent. `prompt_prefix`/`prompt_suffix` (inline) and `prompt_prefix_file`/`prompt_suffix_file` (fragment file) cascade a shared prompt fragment to every agent (#175) — the effective prompt is composed as `prefix → body → prompt_include → suffix` at resolve time, with the suffix always last. An agent opts out with `prompt_suffix: none` / `prompt_prefix: none`. Parts are joined with a fixed blank line (`\n\n`), and a body-less passthrough agent (no own prompt or include) is skipped so the cascade never synthesizes a prompt on it (#248, #249). The `system_prompt_file` default (#72) is different in kind: it is a *fallback* shared system prompt (persona) — used only by agents that declare no `system_prompt`/`system_prompt_file` of their own, which fully override it.

## Retry & Recovery

These defaults control what happens when a node fails, and how the pipeline recovers.

| Field | Type | Description |
|-------|------|-------------|
| `retry_policy` | String | Named retry strategy: `standard`, `aggressive`, `patient`, `linear`, `none`. |
| `max_retries` | Integer | Maximum retry attempts before giving up. |
| `restart_target` | String | Node ID to jump to on restart loops. |
| `max_restarts` | Integer | Max loop restarts before pipeline failure (default: 5). |
| `on_failure` | NodeID | Graph-level catch-all failure route — the runtime sends a failing node here when no more specific route matches. |

The retry and restart *targets* are read from the node, not the defaults, but they participate in the same recovery cascade:

| Field | Type | Description |
|-------|------|-------------|
| `retry_target` | String | Node ID to jump to when retrying — the engine's retry channel, read from the node (not an edge). Same spelling in `dip 1` and `dip 2`. |
| `fallback_target` / `fallback_retry_target` | String | Node ID to route to when all retries are exhausted (read from the node, not an edge). Spelled `fallback_target` in `dip 1`, `fallback_retry_target` in `dip 2`; `dippin fmt --migrate` relabels it. |

**The failure cascade.** When a node fails, the runtime resolves the route in this order: a `ctx.outcome = fail` edge → bounded retry (`retry_target` + `max_retries`) → `fallback_target` → the graph-level `on_failure`. Any one of these gives a node a failure route; an unconditional or success-side edge does not. An agent node with none of them is flagged **DIP144**.

## Budgets & Limits

These defaults cap total resource consumption across the whole run. Each is a hard ceiling; `0` or unset means no limit.

| Field | Type | Description |
|-------|------|-------------|
| `max_total_tokens` | Integer | Hard ceiling on total tokens across the run. |
| `max_cost_cents` | Integer | Hard ceiling on total cost, in US cents (e.g. `1000` = $10.00). |
| `max_wall_time` | Duration | Hard ceiling on wall-clock run time (e.g. `30m`, `2h`). |
| `stall_timeout` | Duration | Abort/route when no forward progress is made for a wall-clock span (e.g. `30s`, `5m`). |

Durations use the same suffix grammar throughout (`30s`, `5m`, `2h`). Budgets are carried and linted by dippin and enforced by the runtime. A negative value on any of these four is flagged **DIP145** — budgets are non-negative, and `0` (or unset) is the way to say "no limit".

## Context Management

These defaults tune how context is persisted and compacted over a long run.

| Field | Type | Description |
|-------|------|-------------|
| `fidelity` | String | Default checkpoint fidelity level for state persistence. |
| `on_resume` | String | Fidelity behavior when a run resumes: `preserve` (keep the checkpoint fidelity level) or `degrade` (downgrade on resume). Only meaningful when `fidelity` is also set. |
| `cache_tools` | Boolean | Whether to cache tool call results. |
| `compaction` | String | Context compaction mode for long pipelines. |

`fidelity`, `cache_tools`, and `compaction` are all overridable per agent node. Agent nodes additionally accept a `compaction_threshold` (Float) that triggers compaction with provider-specific semantics.

## Tool Safety

Tool nodes that shell out can be constrained by two graph-level defaults consumed by the runtime:

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

Agent nodes carry their own tool-safety fields (these are per-node, not defaults):

| Field | Type | Description |
|-------|------|-------------|
| `tool_access` | String | LLM tool-catalog gate. Set to `none` to strip the model's tool registry on this agent. DIP139 warns on unknown values; the runtime fail-closes. |
| `writable_paths` | CSV (globs) | Comma-separated glob list bounding where this agent's tools may write (e.g. `workspace/**, .ai/sprints/**`). Absent = unbounded; a present-but-empty value is rejected by `dippin validate`/`pack`. |

<div class="caveat-card">
  <h4>tool_access does not cross file or context boundaries</h4>
  <p>A <code>tool_access</code> restriction does not follow a <code>subgraph</code>/<code>manager_loop</code> call into a child <code>.dip</code> — a child whose agents re-grant tools is flagged <strong>DIP146</strong> (and its detection-only entry-lint counterpart DIP143). And <code>tool_access</code> bounds an agent's <em>tools</em>, not the <em>information</em> its output carries: a restricted agent that writes a context key a downstream tool-bearing agent reads is flagged <strong>DIP147</strong> (chain attack). Both are detection-only; the runtime enforces the bound.</p>
</div>

## Node-Level Overrides

Every `defaults` field is a fallback. A node inherits the graph-level value unless it declares the same field, in which case the node's value wins for that node only. Fields overridable per node include `model`, `provider`, `retry_policy`, `max_retries`, `fidelity`, `cache_tools`, and `compaction`.

The precedence, from most to least specific:

1. **Node field** — a value set directly on the node.
2. **Defaults block** — the graph-level `defaults` value.
3. **Built-in default** — e.g. `max_restarts` defaults to 5, format version defaults to 1.

Two caveats on where a value is read from:

- The retry and restart *targets* (`retry_target`, `restart_target`, `fallback_target`/`fallback_retry_target`) are read from the node, not composed from defaults — they are the node's own recovery channel.
- The prompt cascade is additive, not an override: a node's `prompt` body is wrapped by the `defaults` prefix/suffix rather than replacing them. A node opts out of a cascade side with `prompt_prefix: none` / `prompt_suffix: none`, and a node's own `system_prompt`/`system_prompt_file` fully replaces the `system_prompt_file` fallback.
