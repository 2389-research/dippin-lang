---
title: "What's New in Dippin v0.21–v0.22"
date: "2026-04-22"
description: "Human timeouts, budget caps, and a new manager_loop node for supervising child pipelines. Two releases, five days apart."
tagStyle: "release"
tagLabel: "RELEASE"
category: "Releases"
readTime: "5 min read"
related:
  - url: "whats-new-v020.html"
    title: "What's New in v0.17–v0.20"
    summary: "Conditional nodes, vars, shell-aware linting, and per-node backend selection."
  - url: "cost-estimation.html"
    title: "Cost Estimation"
    summary: "Estimate pipeline cost before running — now with hard budget caps."
---

Two releases, five days apart. `v0.21.0` shipped budget caps for workflows that get chatty with the API, plus proper timeout support on human nodes. `v0.22.0` added a new node kind — `manager_loop` — for supervising a child sub-pipeline without writing bespoke orchestrator glue.

Here's what changed.

## Human nodes that don't block forever

**v0.21.0** &mdash; Human-gate workflows have always had a hole: if nobody ever answers, the pipeline just sits there. You could hack around it by adding a separate timeout branch in your orchestrator, but the `human` node itself had no concept of a deadline. Now it does:

```dippin
human ReviewGate
  label: "Approve deployment?"
  mode: choice
  timeout: 5m
  timeout_action: default
```

`timeout` accepts any Go duration (`30s`, `5m`, `1h`), and `timeout_action` chooses what happens when it fires: `default` takes the default choice, `fail` fails the node, and `advance` routes through a dedicated edge. Pair it with an edge label to take the timeout branch explicitly:

```dippin
edges
  ReviewGate -> Deploy   when: approved
  ReviewGate -> Rollback when: timeout
```

Round-trips through parse, format, DOT export, and migrate, so DOT-authored workflows pick it up automatically. The runtime consumes the semantics on its side.

## Hard budget caps

**v0.21.0** &mdash; Three new `defaults:` fields let a workflow declare an upper bound on what it's allowed to spend before the runtime pulls the plug:

```dippin
workflow ExpensiveThing
  goal: "Deep reasoning over a large corpus"
  start: Start
  exit: Done

  defaults
    max_total_tokens: 500000
    max_cost_cents: 1000       # $10.00
    max_wall_time: 30m
```

These are ceilings, not estimates — `dippin cost` still tells you what the workflow *should* spend; these fields tell the runtime what it's *allowed* to spend. Useful for pipelines that might loop unexpectedly, or for keeping a staging environment from draining the API budget on one bad run.

If you haven't used `dippin cost` yet, the [cost estimation post](cost-estimation.html) walks through how to get per-run projections from the `.dip` file alone.

## Manager loops: supervisors without the glue

**v0.22.0** &mdash; The biggest addition of the two releases. Before this, "parent pipeline that spawns a child and steers it" was pattern-by-convention: you'd write an `agent` node that called out to a nested workflow, add a bunch of conditional edges to keep re-running it, and hope the retry logic held. Now it's a first-class node kind that maps directly onto the runtime's `stack.manager_loop` construct:

```dippin
manager_loop QualityGate
  subgraph_ref: quality_loop.dip
  poll_interval: 30s
  max_cycles: 5
  stop_condition: stack.child.outcome == "pass"
  steer_condition: stack.child.cycles >= 3
  steer_context:
    focus: "prioritize security issues"
    budget_remaining: "tight"
```

The fields, briefly:

- `subgraph_ref` — path to the child `.dip` file.
- `poll_interval` — how often the supervisor checks the child's state.
- `max_cycles` — maximum iterations before the supervisor gives up.
- `stop_condition` — expression that, when true, tells the supervisor to exit. Usually derived from `stack.child.*`.
- `steer_condition` — when true, the supervisor injects the `steer_context` values into the child's running context mid-run. This is the "nudge" mechanism.
- `steer_context` — `k: v` pairs (inline `k=v,k=v` or block form) pushed into the child when `steer_condition` fires.

DOT export uses shape `house` for these nodes. The `stack.*` namespace is now recognized by `DIP120`, so referencing `stack.child.cycles` / `stack.child.outcome` / `stack.child.status` in conditions doesn't flag as unknown.

### New lint codes

Three lints fire on common manager-loop mistakes:

- **DIP135** — `subgraph_ref` missing or the file doesn't exist.
- **DIP136** — invalid control field (negative `poll_interval`, negative `max_cycles`, etc.).
- **DIP137** — unbounded supervision (`max_cycles: 0` and no `stop_condition`). The manager-loop analog of `DIP104`.

### Scaffolding

`dippin scaffold manager_loop` emits a starter supervisor workflow wired up with sensible defaults — useful if you want a working example to edit rather than hand-rolling the structure.

## Quieter wins

A few smaller items worth calling out:

- **Scoped context reads.** `ctx.node.<id>.*` now validates as a legitimate pattern in `DIP121` / `DIP122`, fixing false-positive warnings when one node reads another node's state.
- **`reasoning_effort` expansion.** `DIP119` accepts `none`, `minimal`, `low`, `medium`, `high`, `xhigh`, and `max` — enough to cover Opus 4.7 and GPT-5.4.
- **Model catalog refresh.** Added `claude-opus-4-7` (Anthropic, $5/$25), `mistral-small-2603` (Mistral Small 4), and `command-a-03-2025` (Cohere's flagship, $2.50/$10). Pricing verified 2026-04-17.
- **Agent-readiness endpoints** on the docs site: `.well-known/agent-skills/index.json`, `.well-known/mcp/server-card.json`, `.well-known/api-catalog`, `robots.txt`, and a hosted `skill.md`. Coding agents can auto-discover dippin-lang tooling without hard-coded config.
- **Tree-sitter grammar** for `manager_loop` (and a CI drift check so generated files can't silently diverge from `grammar.js`). VS Code TextMate grammar picked up the new keyword and `stack.*` namespace.

## What's next

`v0.23.0` is queued up: first-class `defaults:` fields for the runtime's tool-safety allowlist (`tool_commands_allow`) and denylist additions (`tool_denylist_add`). No more smuggling those through `vars:`.

Full details in [CHANGELOG.md](https://github.com/2389-research/dippin-lang/blob/main/CHANGELOG.md).
