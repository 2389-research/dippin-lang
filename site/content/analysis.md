---
title: "Analysis Tools"
description: "Cost estimation, health reports, edge coverage, dead branch detection, semantic diff, and optimization for AI pipelines."
section_label: "Analysis"
subtitle: "Cost, coverage, health, optimization, and change tracking."
---

## Overview

Dippin includes eight analysis commands that inspect workflows for cost, coverage, health, optimization opportunities, dead code, change impact, calibration, and dry-run behavior. `doctor` aggregates cost + coverage + lint into a single grade. Run it first for an overview, then drill into specific commands for details.

<div class="flow-diagram">
  <div class="flow-step">dippin doctor</div>
  <div class="flow-arrow">&rarr;</div>
  <div class="flow-step">dippin lint</div>
  <div class="flow-arrow">+</div>
  <div class="flow-step">dippin coverage</div>
  <div class="flow-arrow">+</div>
  <div class="flow-step">dippin cost</div>
</div>

A typical workflow: run `doctor` first, then drill into `lint`, `coverage`, or `cost` if the grade is below B. Use `optimize` after `cost` shows high costs. Use `diff` to review changes, and `feedback` to calibrate after production runs.

## cost

Estimate execution cost based on model pricing tables. Prompt length estimates input tokens, output tokens are estimated heuristically per turn, and `max_turns` determines the turn range. Tool and human nodes cost $0. Unknown models are costed at $0 with an assumption note.

```
$ dippin cost pipeline.dip
```

<div class="terminal">
  <div class="terminal-bar">
    <span class="terminal-dot"></span><span class="terminal-dot"></span><span class="terminal-dot"></span>
    <span class="terminal-title">dippin cost</span>
  </div>
  <pre><span class="prompt">$</span> dippin cost pipeline.dip
<span class="dim">&#9552;&#9552;&#9552; Cost Estimate &#9552;&#9552;&#9552;&#9552;&#9552;&#9552;&#9552;&#9552;&#9552;&#9552;&#9552;&#9552;&#9552;&#9552;&#9552;&#9552;&#9552;&#9552;&#9552;</span>
                            Min Expected      Max
  <span class="dim">&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;</span> <span class="dim">&#9472;&#9472;&#9472;&#9472;&#9472;</span> <span class="dim">&#9472;&#9472;&#9472;&#9472;&#9472;</span> <span class="dim">&#9472;&#9472;&#9472;&#9472;&#9472;</span>
  TOTAL                       $3.21    $3.59   $14.10

<span class="dim">&#9472;&#9472;&#9472; By Provider &#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;</span>
  openai                      $0.38    $0.57    $2.96
  anthropic                   $2.83    $3.02   $11.13

<span class="dim">&#9472;&#9472;&#9472; Top Cost Drivers &#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;</span>
  CommitWork                  $2.12 <span class="dim">(max)</span>  openai/gpt-5.2
  ImplementClaude             $2.12 <span class="dim">(max)</span>  anthropic/claude-sonnet-4-6
  InterpretRequest            $1.44 <span class="dim">(max)</span>  anthropic/claude-opus-4-6

<span class="dim">&#9472;&#9472;&#9472; Assumptions &#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;</span>
  <span class="warn">&#8226;</span> <span class="dim">unknown model "gemini-3-flash" (provider "gemini"): cost set to $0</span></pre>
</div>

**When to use:** Before deploying a pipeline with expensive models. Compare providers. Identify cost drivers to optimize.

## coverage

Analyze edge coverage and reachability. For tool nodes, extracts possible outputs from `printf`/`echo` patterns in the command, then checks whether outgoing edge conditions cover those outputs.

<div class="terminal">
  <div class="terminal-bar">
    <span class="terminal-dot"></span><span class="terminal-dot"></span><span class="terminal-dot"></span>
    <span class="terminal-title">dippin coverage</span>
  </div>
  <pre><span class="prompt">$</span> dippin coverage pipeline.dip
<span class="dim">&#9552;&#9552;&#9552; Coverage Analysis &#9552;&#9552;&#9552;&#9552;&#9552;&#9552;&#9552;&#9552;&#9552;&#9552;&#9552;&#9552;&#9552;&#9552;&#9552;</span>
<span class="dim">&#9472;&#9472;&#9472; Edge Coverage &#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;</span>
  <span class="pass">&#10003;</span> SetupWorkspace               no_conditions
  <span class="fail">&#10007;</span> ValidateBuild                partial
      <span class="dim">missing: validation-pass-go</span>
      <span class="dim">missing: validation-pass-swift</span>

<span class="dim">&#9472;&#9472;&#9472; Reachability &#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;</span>
  <span class="pass">&#10003;</span> 30/30 nodes reachable

<span class="dim">&#9472;&#9472;&#9472; Termination &#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;</span>
  <span class="pass">&#10003;</span> all paths reach exit: true</pre>
</div>

**When to use:** After writing conditional routing to verify all tool outputs have matching edges. The `missing` entries tell you exactly which edges to add.

## doctor

Health report card — a single grade (A-F) aggregating lint, coverage, and cost into one score.

<div class="terminal">
  <div class="terminal-bar">
    <span class="terminal-dot"></span><span class="terminal-dot"></span><span class="terminal-dot"></span>
    <span class="terminal-title">dippin doctor</span>
  </div>
  <pre><span class="prompt">$</span> dippin doctor pipeline.dip
<span class="dim">&#9552;&#9552;&#9552; Health Report Card &#9552;&#9552;&#9552;&#9552;&#9552;&#9552;&#9552;&#9552;&#9552;&#9552;&#9552;&#9552;&#9552;&#9552;</span>
  Grade: <span class="pass">A</span>  Score: 95/100

<span class="dim">&#9472;&#9472;&#9472; Lint &#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;</span>
  Errors: <span class="pass">0</span>  Warnings: <span class="warn">1</span>  Hints: 0

<span class="dim">&#9472;&#9472;&#9472; Coverage &#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;</span>
  Reachable: <span class="pass">21/21 nodes</span>
  <span class="pass">&#10003;</span> All paths terminate

<span class="dim">&#9472;&#9472;&#9472; Cost &#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;</span>
  Expected: $2.10  <span class="dim">(range: $1.50 - $8.40)</span>

<span class="dim">&#9472;&#9472;&#9472; Suggestions &#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;</span>
  <span class="warn">&#8226;</span> <span class="dim">[lint] review lint warnings - run `dippin lint` for details</span></pre>
</div>

### Scoring Breakdown

Starts at 100 points, with deductions for issues:

| Issue | Deduction |
|-------|-----------|
| Each lint error | -20 points |
| Each lint warning | -5 points |
| Unreachable node | -15 per node |
| Non-terminating paths | -15 |
| Uncovered tool outputs | -10 per tool |

### Grades

| Grade | Score Range |
|-------|-------------|
| **A** | 90-100 |
| **B** | 80-89 |
| **C** | 70-79 |
| **D** | 60-69 |
| **F** | <60 |

## optimize

Suggest cheaper model substitutions without sacrificing quality. Rules include: simple prompts can use cheaper models, nodes in retry loops can use cheaper models for mechanical iterations, and bookkeeping tasks (summary, cleanup, commit) can use cheaper models.

<div class="terminal">
  <div class="terminal-bar">
    <span class="terminal-dot"></span><span class="terminal-dot"></span><span class="terminal-dot"></span>
    <span class="terminal-title">dippin optimize</span>
  </div>
  <pre><span class="prompt">$</span> dippin optimize pipeline.dip
<span class="dim">&#9552;&#9552;&#9552; Optimization Report &#9552;&#9552;&#9552;&#9552;&#9552;&#9552;&#9552;&#9552;&#9552;&#9552;&#9552;&#9552;&#9552;</span>
<span class="dim">&#9472;&#9472;&#9472; Cost Summary &#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;</span>
  Current:   $3.59 <span class="dim">(expected)</span>
  Optimized: $0.00 <span class="dim">(expected)</span>
  Savings:   <span class="pass">$3.59</span> <span class="dim">(expected)</span>

<span class="dim">&#9472;&#9472;&#9472; Suggestions &#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;</span>
  <span class="warn">&#8226;</span> [InterpretRequest] simple prompt does not need an expensive model
    claude-opus-4-6 <span class="dim">&#8594;</span> claude-haiku-4-5  <span class="pass">(saves ~$0.41)</span>
  <span class="warn">&#8226;</span> [CommitWork] bookkeeping task can use a cheaper model
    gpt-5.2 <span class="dim">&#8594;</span> gpt-4o-mini  <span class="pass">(saves ~$0.30)</span></pre>
</div>

**When to use:** After `dippin cost` shows high costs. Review each suggestion — some "simple" prompts may actually need a capable model.

## simulate

Dry-run a workflow without real side effects, emitting JSONL events to stdout. It walks the IR graph from start to exit, logging each node visited — no LLM calls and no shell commands run. Use `--scenario key=val` to inject context and explore conditional branches, `--all-paths` to enumerate every reachable path, and `--interactive` to prompt at human nodes.

```
$ dippin simulate pipeline.dip --scenario outcome=fail
```

**When to use:** Verify routing logic before running the real pipeline — confirm a scenario reaches the expected node sequence, or audit every branch with `--all-paths`.

## diff

Semantic comparison between two workflow versions. Unlike text-based `diff`, this compares graph structure: nodes added/removed, edges changed, field-level modifications, and cost impact.

<div class="terminal">
  <div class="terminal-bar">
    <span class="terminal-dot"></span><span class="terminal-dot"></span><span class="terminal-dot"></span>
    <span class="terminal-title">dippin diff</span>
  </div>
  <pre><span class="prompt">$</span> dippin diff v1.dip v2.dip
<span class="dim">&#9552;&#9552;&#9552; Semantic Diff &#9552;&#9552;&#9552;&#9552;&#9552;&#9552;&#9552;&#9552;&#9552;&#9552;&#9552;&#9552;&#9552;&#9552;&#9552;&#9552;&#9552;&#9552;&#9552;</span>
<span class="dim">&#9472;&#9472;&#9472; Nodes &#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;</span>
  <span class="pass">+</span> FinalQualityGate

<span class="dim">&#9472;&#9472;&#9472; Edges &#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;</span>
  <span class="pass">+</span> FinalQualityGate -&gt; Exit <span class="dim">[ctx.outcome = fail]</span>
  <span class="pass">+</span> FinalQualityGate -&gt; PersistSprint <span class="dim">[ctx.outcome = success]</span>
  <span class="fail">-</span> WriteFinalSprint -&gt; PersistSprint

<span class="dim">&#9472;&#9472;&#9472; Cost Delta &#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;&#9472;</span>
  Old: $5.35 <span class="dim">(expected)</span>  New: $5.78 <span class="dim">(expected)</span>
  Delta: <span class="warn">+$0.43</span> <span class="dim">(expected)</span></pre>
</div>

**When to use:** Code review for workflow changes. See exactly what graph structure changed and how it affects cost, rather than parsing indentation diffs.

## feedback

Compare predicted costs against actual execution telemetry to calibrate estimates. Takes the workflow file (for predicted costs) and a JSONL telemetry file — one JSON object per line with fields: `event`, `node`, `model`, `provider`, `tokens_in`, `tokens_out`, `actual_cost`, `turns`, `timestamp`.

```
$ dippin feedback pipeline.dip telemetry.jsonl
```

After running a pipeline in production, export telemetry and feed it back to see how accurate the cost predictions were. Outliers (>2x or <0.5x ratio) are flagged for investigation.
