---
title: "Structural Validation"
description: "10 structural error codes (DIP001–DIP010) a workflow must satisfy to execute. Caught by dippin validate before runtime."
section_label: "Diagnostics"
subtitle: "Structural errors (DIP001–DIP010) — the 10 checks a workflow must pass to run."
---

## Overview

Dippin provides two levels of analysis:

**Structural validation** (DIP001-DIP010): Errors that must be fixed. A workflow with any of these cannot execute. Run with `dippin validate`.

**Semantic linting** (DIP101-DIP160): Warnings that flag likely bugs or questionable patterns. They don't block execution but should be reviewed. Run with `dippin lint` for both levels.

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
