---
title: "Scenario Testing"
description: "Write deterministic tests for AI pipelines with .test.json files. Inject context, assert on paths, check edge coverage."
section_label: "Testing"
subtitle: "Inject context values, simulate execution paths, and assert outcomes — all without calling LLMs."
---

## How It Works

The `dippin test` command runs scenario-based assertions against workflow files. Test cases inject context values into the simulator and verify the execution path.

<div class="flow-diagram">
  <div class="flow-step">.test.json</div>
  <div class="flow-arrow">&rarr;</div>
  <div class="flow-step">Load &amp; Parse</div>
  <div class="flow-arrow">&rarr;</div>
  <div class="flow-step">Inject Scenario</div>
  <div class="flow-arrow">&rarr;</div>
  <div class="flow-step">Simulate Workflow</div>
  <div class="flow-arrow">&rarr;</div>
  <div class="flow-step">Check Assertions</div>
  <div class="flow-arrow">&rarr;</div>
  <div class="flow-step">PASS / FAIL</div>
</div>

## Test File Format

Test files use `.test.json` extension and are auto-discovered from the workflow path:

```
pipeline.dip       → pipeline.test.json
src/flow.dip       → src/flow.test.json
```

### Schema

```
{
  "tests": [
    {
      "name": "descriptive test name",
      "scenario": {
        "key": "value"
      },
      "expect": {
        "status": "success",
        "visited": ["NodeA", "NodeB"],
        "not_visited": ["NodeC"],
        "path_contains": ["NodeA", "NodeB"],
        "immediately_after": {"NodeA": "NodeB"}
      }
    }
  ]
}
```

### Expectation Fields

All expectation fields are optional. Only specified fields are checked.

| Field | Type | Description |
|-------|------|-------------|
| `status` | string | Expected simulation status: `"success"` (reached exit) or `"dead_end"` |
| `visited` | string[] | Node IDs that must appear in the execution path |
| `not_visited` | string[] | Node IDs that must NOT appear in the execution path |
| `path_contains` | string[] | Node IDs that must appear in order (non-contiguous matches allowed) |
| `immediately_after` | object | Map of `{"NodeA": "NodeB"}` pairs asserting NodeB is the immediate next node after NodeA |

## Example

A workflow that routes based on outcome, with matching test scenarios:

<div class="compare-grid">
  <div class="compare-panel">
    <span class="compare-label good">gate.dip</span>
    <div class="compare-code"><pre><span class="kw">workflow</span> <span class="node">Gate</span>
  <span class="kw">goal:</span> <span class="str">"Route based on outcome"</span>
  <span class="kw">start:</span> Start
  <span class="kw">exit:</span> Exit

  <span class="kw">agent</span> <span class="node">Start</span>
    <span class="kw">label:</span> Start

  <span class="kw">agent</span> <span class="node">Pass</span>
    <span class="kw">model:</span> claude-sonnet-4-6
    <span class="kw">prompt:</span> Handle success.

  <span class="kw">agent</span> <span class="node">Fix</span>
    <span class="kw">model:</span> claude-sonnet-4-6
    <span class="kw">prompt:</span> Handle failure.

  <span class="kw">agent</span> <span class="node">Exit</span>
    <span class="kw">label:</span> Exit

  <span class="kw">edges</span>
    Start <span class="op">-&gt;</span> Pass  <span class="kw">when</span> ctx.outcome = success
    Start <span class="op">-&gt;</span> Fix   <span class="kw">when</span> ctx.outcome = fail
    Pass <span class="op">-&gt;</span> Exit
    Fix <span class="op">-&gt;</span> Exit</pre></div>
  </div>

  <div class="compare-panel">
    <span class="compare-label good">gate.test.json</span>
    <div class="compare-code"><pre>{
  "tests": [
    {
      "name": "success path",
      "scenario": {"outcome": "success"},
      "expect": {
        "status": "success",
        "visited": ["Start", "Pass", "Exit"],
        "not_visited": ["Fix"]
      }
    },
    {
      "name": "failure path",
      "scenario": {"outcome": "fail"},
      "expect": {
        "status": "success",
        "visited": ["Start", "Fix", "Exit"],
        "not_visited": ["Pass"],
        "immediately_after": {"Start": "Fix"}
      }
    }
  ]
}</pre></div>
  </div>
</div>

## Test Output

<div class="terminal">
  <div class="terminal-bar">
    <span class="terminal-dot"></span>
    <span class="terminal-dot"></span>
    <span class="terminal-dot"></span>
    <span class="terminal-title">dippin test</span>
  </div>
  <pre><span class="prompt">$</span> dippin test gate.dip
<span class="dim">&#9552;&#9552;&#9552; Test Results &#9552;&#9552;&#9552;&#9552;&#9552;&#9552;&#9552;&#9552;&#9552;&#9552;&#9552;&#9552;&#9552;&#9552;&#9552;&#9552;&#9552;&#9552;&#9552;&#9552;</span>
  <span class="pass">PASS</span>  success path
  <span class="pass">PASS</span>  failure path
<span class="dim">&#9472;&#9472;&#9472; Summary &#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;</span>
  <span class="pass">2 tests: 2 passed, 0 failed</span>

<span class="prompt">$</span> dippin test --verbose gate.dip
<span class="dim">&#9552;&#9552;&#9552; Test Results &#9552;&#9552;&#9552;&#9552;&#9552;&#9552;&#9552;&#9552;&#9552;&#9552;&#9552;&#9552;&#9552;&#9552;&#9552;&#9552;&#9552;&#9552;&#9552;&#9552;</span>
  <span class="pass">PASS</span>  success path
        <span class="dim">path: Start &rarr; Pass &rarr; Exit</span>
  <span class="pass">PASS</span>  failure path
        <span class="dim">path: Start &rarr; Fix &rarr; Exit</span>
<span class="dim">&#9472;&#9472;&#9472; Summary &#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;</span>
  <span class="pass">2 tests: 2 passed, 0 failed</span></pre>
</div>

## Scenario Keys

The `scenario` object maps context keys to values. The simulator resolves conditions by looking up `ctx.<key>` in the scenario context.

| Key | Description |
|-----|-------------|
| `outcome` | Maps to `ctx.outcome` — the most common routing variable |
| `status` | Maps to `ctx.status` |
| `tool_stdout` | Maps to `ctx.tool_stdout` — tool command output |

## Caveats

<div class="caveat-card">
  <h4>not_visited and loop breaking</h4>
  <p>The test runner limits per-node visits to 3. When a loop exceeds this limit, the simulator forces the loop-exit edge and continues execution rather than stopping. Nodes downstream of the loop-exit can be visited even though the loop was broken. For edge-routing assertions in workflows with loops, prefer <code>path_contains</code> over <code>not_visited</code>.</p>
</div>

<div class="caveat-card">
  <h4>immediately_after for edge routing</h4>
  <p>When testing which specific edge a node takes, <code>immediately_after</code> is more precise than <code>path_contains</code>. Use it to verify that a conditional edge routes to the expected next node.</p>
</div>

<div class="caveat-card">
  <h4>Clearing tool defaults</h4>
  <p>Tool nodes auto-seed <code>ctx.tool_stdout</code> and <code>ctx.outcome</code> to <code>"success"</code>. To test unconditional fallback edges after a tool node, set the key to an empty string: <code>"ToolNode.tool_stdout": ""</code>.</p>
</div>

## Coverage Flag

Use `--coverage` to report edge coverage across all test scenarios:

```
$ dippin test --coverage gate.dip
  PASS  success path
  PASS  failure path
─── Edge Coverage ───
  4/4 edges covered (100.0%)
```

This helps identify edges that no test scenario exercises. (Coverage is edge-only; there is no node-coverage report.)

## CI Integration

Use `--format json` for machine-readable output. Exit code is 0 if all tests pass, 1 if any fail.

```
$ dippin --format json test pipeline.dip
{
  "results": [
    {"name": "happy path", "passed": true, "path": ["Start", "Gate", "Pass", "Exit"]},
    {"name": "error path", "passed": false, "errors": ["expected status \"dead_end\", got \"success\""]}
  ],
  "passed": 1,
  "failed": 1,
  "total": 2
}
```
