---
title: "Scenario Testing with .test.json"
date: "2026-03-27"
description: "Write deterministic tests for non-deterministic AI pipelines. Inject context values, assert on execution paths, and measure edge coverage."
tagStyle: "guide"
tagLabel: "GUIDE"
category: "Testing"
readTime: "10 min read"
related:
  - url: "ci-integration.html"
    title: "CI Integration"
    summary: "Run your scenario tests automatically in GitHub Actions alongside lint and format checks."
  - url: "getting-started.html"
    title: "Getting Started"
    summary: "New to Dippin? Start here to install the toolchain and write your first workflow."
---

AI pipelines are non-deterministic -- LLM outputs vary between runs. But the
*structure* of your pipeline is deterministic: given a particular outcome at each
node, the same path should always be followed. Dippin's scenario testing lets you inject
context values and assert on execution paths, giving you deterministic tests for
non-deterministic systems.

## The Core Idea

A `.test.json` file sits next to your `.dip` file. It holds
an array of test scenarios, each declaring what context values to inject and what
execution behavior to expect (visited nodes, path ordering, status).

The [simulator](https://github.com/2389-research/dippin-lang/tree/main/simulate)
walks the workflow graph, uses your injected values to evaluate edge conditions,
and reports which nodes were visited and in what order. Your test asserts
the result matches expectations.

<div class="callout">
  <h4>Auto-discovery</h4>
  <p>
    <code>dippin test pipeline.dip</code> automatically looks for <code>pipeline.test.json</code>
    in the same directory. No configuration needed.
  </p>
</div>

## A Workflow to Test

Take a real example from the Dippin repository:
[`code_quality_sweep.dip`](https://github.com/2389-research/dippin-lang/blob/main/examples/code_quality_sweep.dip).
This workflow runs three LLM providers in parallel to analyze a codebase, synthesizes
findings, fans out into three work streams (fix bugs, write docs, write tests),
then finishes with a quality gate that can restart the whole process.

The key structural elements:

<pre>
<span class="hl-kw">workflow</span> <span class="hl-node">CodeQualitySweep</span>
  <span class="hl-field">goal</span>: <span class="hl-str">"Analyze the dippin-lang codebase with three LLM providers in parallel..."</span>
  <span class="hl-field">start</span>: ScanCodebase
  <span class="hl-field">exit</span>: Done

  <span class="hl-cmt"># Phase 1: Scan</span>
  <span class="hl-kw">agent</span> <span class="hl-node">ScanCodebase</span>
    <span class="hl-field">label</span>: <span class="hl-str">"Map the codebase"</span>
    <span class="hl-cmt">...</span>

  <span class="hl-cmt"># Phase 2: Three-provider parallel analysis</span>
  <span class="hl-kw">parallel</span> <span class="hl-node">AnalysisFan</span> -> AnalyzeAnthropic, AnalyzeGemini, AnalyzeOpenAI
  <span class="hl-kw">fan_in</span> <span class="hl-node">AnalysisJoin</span> <- AnalyzeAnthropic, AnalyzeGemini, AnalyzeOpenAI

  <span class="hl-cmt"># Phase 3: Synthesize findings</span>
  <span class="hl-kw">agent</span> <span class="hl-node">Synthesize</span>
    <span class="hl-cmt">...</span>

  <span class="hl-cmt"># Phase 4: Three parallel work streams</span>
  <span class="hl-kw">parallel</span> <span class="hl-node">WorkFan</span> -> FixBugs, WriteDocs, WriteTests
  <span class="hl-kw">fan_in</span> <span class="hl-node">WorkJoin</span> <- FixBugs, WriteDocs, WriteTests

  <span class="hl-cmt"># Phase 5: Quality gate with retry</span>
  <span class="hl-kw">agent</span> <span class="hl-node">QualityGate</span>
    <span class="hl-field">goal_gate</span>: <span class="hl-bool">true</span>
    <span class="hl-cmt">...</span>

  <span class="hl-kw">edges</span>
    <span class="hl-cmt">...</span>
    QualityGate -> Done         <span class="hl-cond">when</span> ctx.outcome = success
    QualityGate -> Synthesize   <span class="hl-cond">when</span> ctx.outcome = fail  <span class="hl-field">restart</span>: <span class="hl-bool">true</span>
    QualityGate -> Done
</pre>

The quality gate has three outgoing edges: success goes to Done, failure restarts from
Synthesize, and an unconditional fallback also goes to Done. Three distinct
execution paths, three test scenarios.

## The Test File

Here's the
[`code_quality_sweep.test.json`](https://github.com/2389-research/dippin-lang/blob/main/examples/code_quality_sweep.test.json)
from the repository:

<pre>
{
  "tests": [
    {
      "name": "quality gate passes -- all branches traversed",
      "scenario": {"outcome": "success"},
      "expect": {
        "status": "success",
        "visited": [
          "ScanCodebase",
          "AnalyzeAnthropic", "AnalyzeGemini", "AnalyzeOpenAI",
          "Synthesize",
          "FixBugs", "WriteDocs", "WriteTests",
          "QualityGate", "Done"
        ],
        "path_contains": ["ScanCodebase", "Synthesize", "QualityGate", "Done"]
      }
    },
    {
      "name": "quality gate fails -- restarts from Synthesize",
      "scenario": {"outcome": "fail"},
      "expect": {
        "visited": ["QualityGate", "Synthesize"],
        "path_contains": ["QualityGate", "Synthesize"]
      }
    },
    {
      "name": "all three analysis providers run",
      "scenario": {"outcome": "success"},
      "expect": {
        "path_contains": [
          "AnalyzeAnthropic", "AnalyzeGemini",
          "AnalyzeOpenAI", "AnalysisJoin"
        ]
      }
    },
    {
      "name": "no outcome -- unconditional fallback to Done",
      "scenario": {},
      "expect": {
        "status": "success",
        "visited": ["QualityGate", "Done"]
      }
    },
    {
      "name": "branch filter -- only Gemini analysis",
      "scenario": {"outcome": "success"},
      "branch": ["AnalyzeGemini"],
      "expect": {
        "status": "success",
        "visited": ["AnalyzeGemini"],
        "not_visited": ["AnalyzeAnthropic", "AnalyzeOpenAI"]
      }
    }
  ]
}
</pre>

## Anatomy of a Test Case

Each test case has three parts:

<table class="field-table">
  <thead>
    <tr><th>Field</th><th>Purpose</th></tr>
  </thead>
  <tbody>
    <tr><td><code>name</code></td><td>Human-readable description shown in test output</td></tr>
    <tr><td><code>scenario</code></td><td>Context values to inject. <code>{"outcome": "success"}</code> sets <code>ctx.outcome</code> to "success" at every node.</td></tr>
    <tr><td><code>expect</code></td><td>Assertions about the simulation result</td></tr>
    <tr><td><code>branch</code></td><td>Optional. Filters parallel fan-out to only these branches.</td></tr>
  </tbody>
</table>

### Expectation Fields

The `expect` object supports five assertion types:

<table class="field-table">
  <thead>
    <tr><th>Field</th><th>What it checks</th></tr>
  </thead>
  <tbody>
    <tr><td><code>status</code></td><td>Overall simulation status: <code>"success"</code> or <code>"fail"</code></td></tr>
    <tr><td><code>visited</code></td><td>Node names that must appear in the execution path</td></tr>
    <tr><td><code>not_visited</code></td><td>Node names that must NOT appear in the execution path</td></tr>
    <tr><td><code>path_contains</code></td><td>Node names that must appear in order (not necessarily adjacent)</td></tr>
    <tr><td><code>immediately_after</code></td><td>Object mapping node names: <code>{"A": "B"}</code> asserts B appears right after A</td></tr>
  </tbody>
</table>

## Running Tests

<pre>
<span class="hl-dim">$</span> <span class="hl-shcmd">dippin test examples/code_quality_sweep.dip</span>
<span class="hl-pass">PASS</span>  quality gate passes -- all branches traversed
<span class="hl-pass">PASS</span>  quality gate fails -- restarts from Synthesize
<span class="hl-pass">PASS</span>  all three analysis providers run
<span class="hl-pass">PASS</span>  all three work streams run
<span class="hl-pass">PASS</span>  no outcome -- unconditional fallback to Done
<span class="hl-pass">PASS</span>  branch filter -- only Gemini analysis

<span class="hl-pass">6/6 passed</span>  <span class="hl-dim">examples/code_quality_sweep.dip</span>
</pre>

### Verbose Mode

Add `--verbose` to see the full execution path for each scenario.
Invaluable when debugging a failing test:

<pre>
<span class="hl-dim">$</span> <span class="hl-shcmd">dippin test --verbose examples/code_quality_sweep.dip</span>
<span class="hl-pass">PASS</span>  quality gate passes -- all branches traversed
  <span class="hl-dim">path: ScanCodebase -> AnalysisFan -> AnalyzeAnthropic -> AnalysisJoin
        -> AnalyzeGemini -> AnalysisJoin -> AnalyzeOpenAI -> AnalysisJoin
        -> Synthesize -> WorkFan -> FixBugs -> WorkJoin -> WriteDocs
        -> WorkJoin -> WriteTests -> WorkJoin -> QualityGate -> Done</span>
<span class="hl-dim">...</span>
</pre>

## Writing Your Own Tests

### Step 1: Identify the branches

Look at your edge conditions. Each `when` clause creates a branch.
You need at least one test scenario per branch. Think about:

- The success path (all conditions evaluate to "success")
- The failure path (what happens when a node fails)
- The fallback path (what happens when no condition matches)
- Parallel branch isolation (individual fan-out branches)

### Step 2: Write the scenario injection

The `scenario` object sets context values the simulator uses when
evaluating edge conditions. Keys correspond to variable names in
`when` clauses, without the `ctx.` prefix:

<pre>
<span class="hl-cmt">// Edge condition in .dip file:</span>
QualityGate -> Done  <span class="hl-cond">when</span> ctx.outcome = success

<span class="hl-cmt">// Corresponding scenario injection in .test.json:</span>
"scenario": {"outcome": "success"}
</pre>

### Step 3: Assert on structure, not content

The key insight: you assert on *which nodes were visited*
and *in what order*, never on LLM response content. The tests
stay deterministic because you're testing the graph's routing logic, not
the models' output.<sup>1</sup>

<div class="callout-warn callout">
  <h4>Pitfall: not_visited fragility</h4>
  <p>
    Be careful with <code>not_visited</code>. If your workflow has retry loops with
    <code>restart: true</code> edges, the simulator's loop-breaking may visit nodes
    you don't expect. Prefer positive assertions (<code>visited</code>,
    <code>path_contains</code>) when possible. Use <code>not_visited</code> sparingly.<sup>2</sup>
  </p>
</div>

## The branch Filter

For workflows with `parallel` fan-out, you can test individual branches
in isolation with the `branch` field:

<pre>
{
  "name": "branch filter -- only Gemini analysis",
  "scenario": {"outcome": "success"},
  "branch": ["AnalyzeGemini"],
  "expect": {
    "status": "success",
    "visited": ["AnalyzeGemini"],
    "not_visited": ["AnalyzeAnthropic", "AnalyzeOpenAI"]
  }
}
</pre>

The `branch` array lists the parallel targets to include. All other
fan-out branches get skipped, letting you test branch-specific behavior without
noise from other parallel paths.

## Edge Coverage

Add `--coverage` to see which edges your test suite covers:

<pre>
<span class="hl-dim">$</span> <span class="hl-shcmd">dippin test --coverage examples/code_quality_sweep.dip</span>
<span class="hl-pass">6/6 passed</span>  <span class="hl-dim">examples/code_quality_sweep.dip</span>

<span class="hl-dim">Edge coverage:</span> <span class="hl-pass">24/25 edges covered (96.0%)</span>
<span class="hl-dim">Uncovered edges:</span>
  <span class="hl-warn">QualityGate -> Synthesize  when ctx.outcome = fail  restart: true</span>
</pre>

The uncovered edge tells you exactly what test case to add. Write a scenario that
triggers that condition to reach 100%.

## JSON Output for CI

For [CI integration](ci-integration.html), use `--format json`
to get machine-readable results:

<pre>
<span class="hl-dim">$</span> <span class="hl-shcmd">dippin test --format json examples/code_quality_sweep.dip</span>
{
  "file": "examples/code_quality_sweep.dip",
  "total": 6,
  "passed": 6,
  "failed": 0,
  "results": [
    {"name": "quality gate passes -- all branches traversed", "status": "pass"},
    {"name": "quality gate fails -- restarts from Synthesize", "status": "pass"},
    <span class="hl-dim">...</span>
  ]
}
</pre>

## What's Next?

You know how to write scenario tests that verify your pipeline's routing
logic deterministically. Related topics:

<div class="footnotes">
  <h3>Notes</h3>
  <ol>
    <li id="fn1">The simulator doesn't call any LLMs. It walks the graph using your injected context values to pick edges, which is why tests run in milliseconds and cost nothing. The simulation implementation is in <a href="https://github.com/2389-research/dippin-lang/tree/main/simulate"><code>simulate/</code></a>.</li>
    <li id="fn2">The <code>not_visited</code> fragility was discovered during field testing. Retry loops with <code>restart: true</code> create cycles, and the simulator breaks them after a bounded number of iterations -- but the nodes visited during those iterations can surprise you. See the <a href="../testing.html">testing reference</a> for details on loop-breaking behavior.</li>
  </ol>
</div>
