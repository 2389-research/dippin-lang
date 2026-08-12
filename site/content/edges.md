---
title: "Edges & Routing"
description: "How .dip workflows connect nodes: the edges block, edge attributes, outcome shorthand, loops, section-level defaults, conditions, and the routing cascade."
section_label: "Language"
subtitle: "Connecting nodes: edges, conditions, and the routing cascade."
---

## The Edges Block

The `edges` block defines connections between nodes. Each edge is a single line:

```
<FromID> -> <ToID> [on <token> | when <condition>] [label: <text>] [choice: <key>] [weight: <int>] [loop] [override: true]
```

Basic edges connect nodes unconditionally; conditional edges add a guard with `on` or `when`:

```
  edges
    AskUser -> Interpret
    Interpret -> Validate
    Validate -> Approve   when ctx.outcome = success
    Validate -> Retry     when ctx.outcome = fail
```

## Edge Attributes

| Attribute | Type | Description |
|-----------|------|-------------|
| `on <token>` | Token | Shorthand for an equality guard against the source node's outcome channel — agent (`ctx.outcome`) or tool + `marker_grep` (`ctx.tool_marker`) |
| `when <expr>` | Condition | Boolean guard — edge only traversed if true |
| `label: <text>` | String | Human-readable label (used for human gate choices when no `choice:` is set) |
| `choice: <key>` | String | Carried, not interpreted by dippin — explicit human-gate routing key the paired runtime matches the user's selection against, leaving `label:` for display. Wins over `label:` when present; runtime falls back to `label:` when absent |
| `weight: <int>` | Integer | **Soft-deprecated** (raises DIP151). Parsed but **unused by routing**; slated for removal in `dip 2`. Historically a priority hint, but the cascade never consults it — guard edges with `when` / `on` instead |
| `loop` | Flag | Bare keyword marking a back-edge (loop restart). Legacy `restart: true` is an accepted synonym that `dippin fmt` rewrites to `loop` |
| `override: true` | Boolean | Carried, not interpreted by dippin — marks a human-authored validation override for a paired runtime to act on |

## Outcome Shorthand (`on`)

The most common condition is an equality test against the source node's *natural outcome channel*. `on <token>` is shorthand for exactly that:

```
    Review -> Approve  on success      # = when ctx.outcome = success
    Review -> Reject   on fail         # = when ctx.outcome = fail
    Tests  -> Ship     on tests_green  # = when ctx.tool_marker = tests_green
```

The channel comes from the source node's kind: **agent** nodes use `ctx.outcome`; **tool** nodes that declare `marker_grep` use `ctx.tool_marker`. `on` is pure sugar — it produces the identical condition, is non-breaking, and `dippin fmt` rewrites eligible `when` edges into `on` form. Sources without an outcome channel — human gates, `conditional` nodes, and tools without `marker_grep` — have no `on` channel; use `when` there, as well as for anything that isn't a single equality (`and`/`or`, `contains`, `!=`, etc.).

## Loop Edges

A `loop` edge creates a controlled back-edge. When followed, the engine increments a restart counter (max controlled by `max_restarts`), clears downstream nodes, resets retry budgets, and resumes from the target.

```
    Validate -> Implement when ctx.outcome = fail loop
```

`loop` is a bare keyword; the legacy `restart: true` still parses and `dippin fmt` rewrites it to `loop`. Because `loop` is a reserved bare keyword, write `when ctx.x = "loop"` (quoted) for the literal value on a condition's right-hand side.

## Section-Level Default (`else`)

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

## Routing Cascade

Guards compose into a single routing model. On the **success side**, an edge's guard is either `on <token>` (equality against the source's outcome channel) or a full `when <expr>`; a node whose guard edges all fail to match, and which has no unconditional edge, falls through to the section-level `else` default. On the **failure side**, a genuine node failure never routes through `else` — it follows the failure cascade: an `on fail` edge, then the graph-level `defaults.on_failure` catch-all.

The linter reasons about this cascade: **DIP101** flags a source node whose outgoing conditions leave a case unrouted (and suppresses when the conditions are exhaustive or the node is covered by `else`). **DIP149** and **DIP152** are the related routing diagnostics — see the lint reference for the full behavior of each code.
