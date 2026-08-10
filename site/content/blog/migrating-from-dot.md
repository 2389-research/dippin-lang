---
title: "Migrating from DOT to Dippin"
date: "2026-03-27"
description: "Convert Graphviz DOT pipeline files to Dippin with automated migration and structural parity verification. Step-by-step with real examples."
tagStyle: "guide"
tagLabel: "GUIDE"
category: "Migration"
readTime: "12 min read"
related:
  - url: "scenario-testing.html"
    title: "Scenario Testing"
    summary: "Write deterministic tests for your newly migrated workflows."
  - url: "ci-integration.html"
    title: "CI Integration"
    summary: "Add validation, linting, and formatting checks to your CI pipeline."
---

If your team has existing pipelines authored in
[Graphviz DOT](https://graphviz.org/doc/info/lang.html), Dippin provides
an automated migration path. `dippin migrate` converts DOT files
to Dippin syntax, and `dippin validate-migration` verifies structural parity
between the original and the converted file.<sup>1</sup>

## Why Migrate?

DOT was never designed for AI pipelines. The pain points stack up:

| Problem in DOT | How Dippin fixes it |
|----------------|---------------------|
| Prompts need `\n` escaping and `\"` quoting | Indentation-based multi-line blocks, no escaping |
| No validation beyond syntax | [70 diagnostic codes](../validation.html): structural errors + semantic warnings |
| No testing framework | [Scenario testing](scenario-testing.html) with `.test.json` files |
| No formatting standard | Idempotent `dippin fmt` with canonical field ordering |
| No cost estimation | `dippin cost` and `dippin optimize` |
| Node types inferred from shape attribute | Explicit `agent`, `tool`, `human` keywords |

## The Migration Process

We'll migrate `consensus_task.dot`, a 17-node multi-model consensus
workflow with conditional edges and restart loops.

<h3 class="migration-step">Run the Converter</h3>

<pre>
<span class="hl-dim">$</span> <span class="hl-shcmd">dippin migrate --output consensus_task.dip consensus_task.dot</span>
<span class="hl-pass">OK</span>  Migrated consensus_task.dot -> consensus_task.dip
  <span class="hl-dim">17 nodes, 18 edges converted</span>
  <span class="hl-warn">2 warnings:</span>
  <span class="hl-warn">  - condition prefix "context." converted to "ctx."</span>
  <span class="hl-warn">  - condition prefix "context.internal." converted to "ctx.internal."</span>
</pre>

The converter handles the mechanical translation: DOT shape attributes map to
Dippin node kinds, escaped strings become multi-line blocks,
edge attributes become `when` clauses, and graph-level attributes
become workflow header fields.

<h3 class="migration-step">Understand the Mapping</h3>

How DOT constructs map to Dippin:

| DOT | Dippin |
|-----|--------|
| `shape=box` | `agent` |
| `shape=hexagon` | `human` |
| `shape=component` | `tool` |
| `shape=Mdiamond` | `agent` (start node) |
| `shape=Msquare` | `agent` (exit node) |
| `llm_provider="anthropic"` | `provider: anthropic` |
| `llm_model="claude-opus-4-6"` | `model: claude-opus-4-6` |
| `prompt="text\nmore text"` | `prompt:` + indented block |
| `condition="outcome=success"` | `when ctx.outcome = success` |
| `loop_restart=true` | `restart: true` |

<h3 class="migration-step">Compare Side by Side</h3>

A key section of the consensus workflow in both formats:

<div class="side-by-side">
  <div>
    <span class="side-label dot">DOT</span>
    <pre>
ReviewConsensus [
  shape=box,
  label="Review Consensus",
  llm_provider="anthropic",
  llm_model="claude-opus-4-6",
  reasoning_effort="high",
  max_retries=1,
  retry_target="Implement",
  prompt="Produce final consensus
verdict. Use success when ready
to exit, retry when rework is
required, fail when blocked."
];

ReviewConsensus -> Exit
  [condition="outcome=success",
   label="pass"];
ReviewConsensus -> Postmortem
  [condition="outcome=retry",
   label="retry"];
ReviewConsensus -> Exit;</pre>
  </div>
  <div>
    <span class="side-label dip">Dippin</span>
    <pre>
<span class="hl-kw">agent</span> <span class="hl-node">ReviewConsensus</span>
  <span class="hl-field">label</span>: <span class="hl-str">"Review Consensus"</span>
  <span class="hl-field">model</span>: claude-opus-4-6
  <span class="hl-field">provider</span>: anthropic
  <span class="hl-field">reasoning_effort</span>: high
  <span class="hl-field">max_retries</span>: <span class="hl-num">1</span>
  <span class="hl-field">retry_target</span>: Implement
  <span class="hl-field">prompt</span>:
    Produce final consensus
    verdict. Use success when
    ready to exit, retry when
    rework is required, fail
    when blocked.

<span class="hl-kw">edges</span>
  ReviewConsensus -> Exit
    <span class="hl-cond">when</span> ctx.outcome = success
    <span class="hl-field">label</span>: pass
  ReviewConsensus -> Postmortem
    <span class="hl-cond">when</span> ctx.outcome = retry
    <span class="hl-field">label</span>: retry
  ReviewConsensus -> Exit</pre>
  </div>
</div>

The Dippin version reads more clearly, and you can validate, lint,
[test](scenario-testing.html), and format it with the toolchain.

<h3 class="migration-step">Verify Structural Parity</h3>

`validate-migration` compares the DOT original against
the Dippin conversion to verify nothing was lost:

<pre>
<span class="hl-dim">$</span> <span class="hl-shcmd">dippin validate-migration consensus_task.dot consensus_task.dip</span>
<span class="hl-pass">PASS</span>  Migration parity verified
  <span class="hl-dim">Nodes: 17/17 matched</span>
  <span class="hl-dim">Edges: 18/18 matched</span>
  <span class="hl-dim">Conditions: 4/4 matched</span>
</pre>

Parity checking compares node count, node names, edge connections, condition
expressions, and key attributes (model, provider, labels). It catches
dropped nodes and mangled conditions.

<h3 class="migration-step">Validate and Lint the Result</h3>

<pre>
<span class="hl-dim">$</span> <span class="hl-shcmd">dippin validate consensus_task.dip</span>
<span class="hl-pass">PASS</span>  consensus_task.dip
</pre>

<pre>
<span class="hl-dim">$</span> <span class="hl-shcmd">dippin lint consensus_task.dip</span>
<span class="hl-pass">PASS</span>  consensus_task.dip  <span class="hl-dim">(0 errors, 0 warnings)</span>
</pre>

<h3 class="migration-step">Format for Canonical Style</h3>

<pre>
<span class="hl-dim">$</span> <span class="hl-shcmd">dippin fmt --write consensus_task.dip</span>
</pre>

The formatter normalizes field ordering, indentation, and spacing, so the
migrated file follows the same conventions as hand-written Dippin code.

## Common Parity Issues

### Condition Prefix Differences

DOT conditions often use `context.` as the variable prefix. Dippin uses
`ctx.`. The migrator handles the translation automatically:

<div class="side-by-side">
  <div>
    <span class="side-label dot">DOT</span>
    <pre>
[condition="context.internal.
loop_restart_count=0",
 loop_restart=true]</pre>
  </div>
  <div>
    <span class="side-label dip">Dippin</span>
    <pre>
<span class="hl-cond">when</span> ctx.internal.loop_restart_count = 0
  <span class="hl-field">restart</span>: <span class="hl-bool">true</span></pre>
  </div>
</div>

### Missing Fields

DOT files often omit fields that Dippin encourages. After migration, run
`dippin lint` to find the gaps:

<pre>
<span class="hl-dim">$</span> <span class="hl-shcmd">dippin lint migrated.dip</span>
<span class="hl-warn">WARN</span>  migrated.dip
  <span class="hl-warn">DIP111</span>: tool node "RunScript" has no timeout
  <span class="hl-warn">DIP103</span>: node "Process" has no prompt
</pre>

These warnings are actionable -- add the missing fields to strengthen the pipeline.

### Graph-Level Defaults

DOT graph attributes like `default_fidelity` and `default_max_retry`
become Dippin [defaults block](../language.html) fields:

<pre>
<span class="hl-cmt"># DOT: graph [ default_fidelity="truncate", default_max_retry=3 ]</span>
<span class="hl-cmt"># Becomes:</span>

<span class="hl-kw">defaults</span>
  <span class="hl-field">fidelity</span>: truncate
  <span class="hl-field">max_retries</span>: <span class="hl-num">3</span>
</pre>

## Batch Migration

Migrate an entire directory of DOT files:

<pre>
<span class="hl-dim">$</span> <span class="hl-shcmd">for f in pipelines/*.dot; do</span>
  <span class="hl-shcmd">dippin migrate --output "${f%.dot}.dip" "$f"</span>
<span class="hl-shcmd">done</span>
</pre>

Then verify parity for all of them:

<pre>
<span class="hl-dim">$</span> <span class="hl-shcmd">for f in pipelines/*.dot; do</span>
  <span class="hl-shcmd">dip="${f%.dot}.dip"</span>
  <span class="hl-shcmd">[ -f "$dip" ] && dippin validate-migration "$f" "$dip"</span>
<span class="hl-shcmd">done</span>
<span class="hl-pass">PASS</span>  pipelines/consensus_task.dot <-> pipelines/consensus_task.dip
<span class="hl-pass">PASS</span>  pipelines/code_review.dot <-> pipelines/code_review.dip
<span class="hl-pass">PASS</span>  pipelines/deploy_check.dot <-> pipelines/deploy_check.dip
</pre>

## The CI Safety Net

The Dippin [CI workflow](https://docs.github.com/en/actions) includes a
migration parity step on every push. It finds all `.dot` files with a
corresponding `.dip` file and verifies they remain structurally equivalent:

<pre>
<span class="hl-cmt"># From .github/workflows/ci.yml</span>
- name: Validate migration parity
  run: |
    failed=0
    for f in examples/*.dot; do
      base=$(basename "$f" .dot)
      dip="examples/${base}.dip"
      if [ -f "$dip" ]; then
        if ! ./dippin validate-migration "$f" "$dip" 2>&amp;1; then
          failed=1
        fi
      fi
    done
    exit $failed
</pre>

As you improve the Dippin version, the CI check confirms it stays
structurally compatible with the DOT original. Once you're confident,
delete the DOT file and the check stops running for that pair.

## Post-Migration Improvements

After migration, you can take advantage of features DOT doesn't support:

| Feature | What to do |
|---------|-----------|
| Scenario tests | Create a [`.test.json`](scenario-testing.html) file to test execution paths |
| Goal gates | Add `goal_gate: true` to critical quality check nodes |
| Data flow declarations | Add `reads`/`writes` to enable DIP106/DIP107/DIP112 checks |
| Cost estimation | Run `dippin cost` and `dippin optimize` |
| Health reports | Run `dippin doctor` for a letter-grade assessment |
| Watch mode | Run `dippin watch pipeline.dip` for live feedback |

## What's Next?

Your DOT pipelines are now Dippin files with full toolchain support.

<div class="footnotes">
  <h3>Notes</h3>
  <ol>
    <li id="fn1">DOT was chosen as the original authoring format because Graphviz is ubiquitous and every developer already has <code>dot</code> installed. But DOT's string-quoting requirements made multi-line prompts painful to write and impossible to diff-review in PRs. The tipping point was a 40-line prompt that required 80+ characters of escape sequences -- at that point, building a purpose-built language was cheaper than the ongoing authoring tax.</li>
  </ol>
</div>
