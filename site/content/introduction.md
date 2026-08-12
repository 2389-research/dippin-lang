---
title: "Introduction"
description: "What Dippin is and why it exists — a DSL and toolchain for authoring AI pipeline workflows that replaces Graphviz DOT as the authoring format."
section_label: "Get Started"
subtitle: "What Dippin is and why it exists."
---

## What is Dippin

Dippin is a DSL and toolchain for authoring AI pipeline workflows. It replaces Graphviz DOT as the authoring format consumed by a downstream pipeline runtime.

You write workflows as `.dip` files — multi-model agent workflows with first-class syntax for prompts, tools, conditions, and parallel execution. Instead of packing everything into escaped DOT strings, you get real code:

```
workflow CodeReview
  goal: "Plan, implement, and review"
  start: Planner
  exit: Done

  agent Planner
    model: claude-opus-4-6
    prompt:
      You are a senior architect.
      Analyze the request and
      produce an implementation plan.

  agent Coder
    model: claude-sonnet-4-6
    prompt:
      Implement the plan.
      Use best practices.

  agent Review
    auto_status: true
    prompt:
      Review the code.
      Set STATUS: success or fail.

  agent Done
    prompt: Ship it.

  edges
    Planner -> Coder
    Coder -> Review
    Review -> Done  when ctx.outcome = success
    Review -> Coder  when ctx.outcome = fail  restart: true
```

## Why not just DOT

Graphviz DOT is great for graph visualization. But authoring AI pipelines with multi-line prompts, typed nodes, and conditional edges? It falls apart.

Dippin is built around the things that matter when authoring pipelines — not string attributes on a graph node:

- **Multi-line prompts** — indented blocks with zero escaping. Write real prompts, preserve blank lines, embed variables like `${ctx.input}`.
- **Typed node kinds** — `agent`, `tool`, `human`, `conditional`, `parallel`, `fan_in`, `subgraph`. Each with typed, validated config fields.
- **Diagnostics** — structural validation and semantic lint. Dead edges, unreachable nodes, missing prompts, invalid models. Things DOT silently ignores.
- **Parallel execution** — native `parallel` fan-out and `fan_in` join with per-branch model overrides. Multi-provider consensus in a few lines.
- **Conditional edges** — route pipelines based on LLM output with the `when` keyword.

DOT wasn't built for any of this. Dippin was.

## The Toolchain at a Glance

Dippin is more than a syntax — it's a full toolchain that lets you catch problems before runtime. Every command works offline against the workflow source, without deploying or calling any LLMs.

<div class="flow-diagram">
  <div class="flow-step">Author<br>.dip files</div>
  <div class="flow-arrow">&rarr;</div>
  <div class="flow-step">Validate<br>structural checks</div>
  <div class="flow-arrow">&rarr;</div>
  <div class="flow-step">Lint<br>semantic diagnostics</div>
  <div class="flow-arrow">&rarr;</div>
  <div class="flow-step">Test<br>scenario runs</div>
  <div class="flow-arrow">&rarr;</div>
  <div class="flow-step">Analyze<br>cost, coverage, doctor</div>
  <div class="flow-arrow">&rarr;</div>
  <div class="flow-step">Export<br>Mermaid, DOT</div>
</div>

- **Author** — write `.dip` files, or scaffold one with `dippin new`.
- **Validate & Lint** — 72 diagnostic rules (DIP001-DIP010 structural, DIP101-DIP162 semantic) catch dead edges, unreachable nodes, missing prompts, and invalid models. See the [CLI Reference](/cli/).
- **Test** — `dippin test` injects context, simulates every conditional branch, and checks assertions, with CI-ready output. See [Testing](/testing/).
- **Analyze** — `dippin cost`, `dippin coverage`, and `dippin doctor` estimate spend, check reachability, and grade a workflow A&ndash;F. See [Analysis](/analysis/).
- **Export** — turn a workflow into a live diagram with `export-mermaid` or `export-dot`. See [Export & Visualization](/export/).

## Next Steps

<div class="cmd-card">
  <h3><a href="/language/">Language Reference</a></h3>
  <p>The full syntax for <code>.dip</code> files — file structure, nodes, edges, conditions, multiline prompts, and stylesheets.</p>
</div>

<div class="cmd-card">
  <h3><a href="/nodes/">Nodes</a></h3>
  <p>The typed node kinds — <code>agent</code>, <code>tool</code>, <code>human</code>, <code>conditional</code>, <code>parallel</code>, <code>fan_in</code>, <code>subgraph</code> — and their fields.</p>
</div>

<div class="cmd-card">
  <h3><a href="/playground/">Playground</a></h3>
  <p>Write and validate <code>.dip</code> workflows in the browser, with live Mermaid diagrams rendered as you type.</p>
</div>
