---
title: "Export & Visualization"
description: "Turn a Dippin workflow into a diagram — export to Mermaid or Graphviz DOT, migrate existing DOT pipelines to .dip, and pipe DOT output through Graphviz for PNG/SVG images."
section_label: "Tooling"
subtitle: "Turn a workflow into a diagram — Mermaid, DOT, and migration."
---

A workflow is a graph, and sometimes you want to see it. Dippin exports to two diagram formats — [Mermaid](https://mermaid.js.org/) and Graphviz DOT — and can migrate existing DOT pipelines back into `.dip` source.

## export-mermaid

<div class="cmd-card">
  <h3>export-mermaid</h3>
  <div class="cmd-usage">dippin export-mermaid &lt;file&gt;</div>
  <p>Export a workflow to a <a href="https://mermaid.js.org/">Mermaid</a> flowchart — shapes and colors by node kind, edges labeled by routing condition, start/exit emphasized. Renders natively on GitHub and in the <a href="/playground/">playground</a>; the quickest way to drop a live workflow diagram into a README. Subgraph refs are flattened first.</p>
</div>

Because Mermaid renders natively on GitHub, piping a workflow into a fenced ```` ```mermaid ```` block in your README gives you a live, always-current diagram with no build step:

```
dippin export-mermaid pipeline.dip
```

## export-dot

<div class="cmd-card">
  <h3>export-dot</h3>
  <div class="cmd-usage">dippin export-dot [--rankdir=LR|TB] [--prompts] &lt;file&gt;</div>
  <p>Export a workflow to Graphviz DOT format for visualization. Maps node kinds to DOT shapes (agent=box, human=hexagon, tool=parallelogram). Goal gate nodes get red background; restart edges are dashed.</p>
  <dl>
    <dt><code>--rankdir=LR|TB</code></dt>
    <dd>Set graph layout direction — left-to-right or top-to-bottom.</dd>
    <dt><code>--prompts</code></dt>
    <dd>Include prompt text inside the rendered nodes.</dd>
  </dl>
</div>

## migrate / validate-migration

Coming from an existing Graphviz DOT pipeline? Convert it to `.dip` source and verify the conversion preserved the graph.

<div class="cmd-card">
  <h3>migrate</h3>
  <div class="cmd-usage">dippin migrate [--output &lt;file&gt;] &lt;file.dot&gt;</div>
  <p>Convert a DOT file to <code>.dip</code> source format. Maps DOT shapes to Dippin node kinds, extracts graph attributes, unescapes prompts, and prefixes bare condition variables with <code>ctx.</code>.</p>
</div>

<div class="cmd-card">
  <h3>validate-migration</h3>
  <div class="cmd-usage">dippin validate-migration &lt;old.dot&gt; &lt;new.dip&gt;</div>
  <p>Check structural parity between a DOT file and a <code>.dip</code> file to verify migration correctness. Reports missing nodes, different edges, and changed conditions.</p>
</div>

## Graphviz Integration

For visual workflow diagrams, pipe `export-dot` output to Graphviz:

```
# PNG output
dippin export-dot pipeline.dip | dot -Tpng -o pipeline.png

# SVG for web
dippin export-dot pipeline.dip | dot -Tsvg -o pipeline.svg

# Left-to-right layout (better for wide pipelines)
dippin export-dot --rankdir=LR pipeline.dip | dot -Tpng -o pipeline.png

# Include prompt text in nodes
dippin export-dot --prompts pipeline.dip | dot -Tpng -o pipeline.png
```

Install Graphviz: `brew install graphviz` (macOS), `apt install graphviz` (Debian/Ubuntu), or [graphviz.org](https://graphviz.org/download/).
