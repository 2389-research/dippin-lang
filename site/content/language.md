---
title: "Language Overview"
description: "The shape of a .dip workflow file: file structure, format version, the workflow header, and where each language feature is documented."
section_label: "Language"
subtitle: "The shape of a .dip file — structure, header, and where each piece is documented."
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

## Format Version (dip N)

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

## Where to go next

Each part of the language has its own reference page:

<div class="flow-diagram">
  <div class="flow-step"><strong><a href="/nodes">Nodes</a></strong><br>The 8 node kinds (agent, human, tool, parallel, fan_in, subgraph, conditional, manager_loop) and their common fields</div>
  <div class="flow-step"><strong><a href="/edges">Edges &amp; Routing</a></strong><br>The edges block, conditions, <code>on</code>/<code>when</code> guards, loop back-edges, and the <code>else</code> default</div>
  <div class="flow-step"><strong><a href="/prompts">Prompts &amp; Context</a></strong><br>Multiline blocks, <code>prompt</code>/<code>system_prompt</code>, prompt files, and the defaults cascade</div>
  <div class="flow-step"><strong><a href="/inputs">Inputs</a></strong><br>The <code>inputs</code> block — the workflow's callee-side signature</div>
  <div class="flow-step"><strong><a href="/configuration">Configuration</a></strong><br>The <code>defaults</code> block: models, retries, budgets, and tool safety</div>
  <div class="flow-step"><strong><a href="/models">Models &amp; Pricing</a></strong><br>Supported providers, the model catalog, and cost estimation</div>
</div>
