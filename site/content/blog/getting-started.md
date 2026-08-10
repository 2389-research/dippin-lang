---
title: "Getting Started with Dippin"
date: "2026-03-27"
description: "Install Dippin, write your first AI pipeline workflow, and validate it in under 10 minutes. A step-by-step guide from zero to a working .dip file."
tagStyle: "guide"
tagLabel: "GUIDE"
category: "Authoring"
readTime: "8 min read"
related:
  - url: "scenario-testing.html"
    title: "Scenario Testing with .test.json"
    summary: "Write deterministic tests for non-deterministic pipelines. Assert on execution paths, not LLM output."
  - url: "ci-integration.html"
    title: "CI Integration"
    summary: "Set up GitHub Actions to lint, test, and format-check your workflows on every push."
  - url: "migrating-from-dot.html"
    title: "Migrating from DOT"
    summary: "Already have Graphviz DOT pipelines? Migrate them automatically with parity verification."
  - url: "editor-setup.html"
    title: "Editor Setup"
    summary: "Get real-time diagnostics, hover docs, and syntax highlighting in VS Code, Neovim, or any LSP editor."
  - url: "conditional-edges.html"
    title: "Conditional Edges"
    summary: "Route pipelines based on LLM output with the when keyword. Build branching workflows step by step."
  - url: "cost-estimation.html"
    title: "Cost Estimation"
    summary: "Estimate per-run pipeline costs before spending real money on LLM calls."
---

Dippin is a domain-specific language for authoring AI pipeline workflows. It replaces
[Graphviz DOT](https://graphviz.org/doc/info/lang.html) as the format for
defining multi-step LLM pipelines, with first-class support for multi-line prompts,
conditional routing, parallel execution, and a full
[validation toolchain](../validation.html). You'll go from zero to a
validated, linted, and visualized workflow in under 10 minutes.

## <span class="step-num">1</span> Install the Toolchain

Dippin is a single [Go](https://go.dev/) binary. Install it with `go install`:

<pre>
<span class="hl-dim">$</span> <span class="hl-shcmd">go install github.com/2389-research/dippin-lang/cmd/dippin@latest</span>
</pre>

Or via Homebrew (macOS and Linux):

<pre>
<span class="hl-dim">$</span> <span class="hl-shcmd">brew install 2389-research/tap/dippin-lang</span>
</pre>

Verify it works:

<pre>
<span class="hl-dim">$</span> <span class="hl-shcmd">dippin --version</span>
dippin v0.10.0
</pre>

One binary, everything included: parser, validator, linter, formatter, simulator,
cost estimator, LSP server, and DOT exporter. No plugins or runtime dependencies.

## <span class="step-num">2</span> Write Your First Workflow

Create a file called `pipeline.dip`. A Dippin workflow has four parts:
a header with metadata, optional defaults, node definitions, and edges connecting them.

<pre>
<span class="hl-kw">workflow</span> <span class="hl-node">MyFirstPipeline</span>
  <span class="hl-field">goal</span>: <span class="hl-str">"Summarize a document and then translate the summary to Spanish"</span>
  <span class="hl-field">start</span>: Start
  <span class="hl-field">exit</span>: Exit

  <span class="hl-kw">defaults</span>
    <span class="hl-field">provider</span>: anthropic
    <span class="hl-field">model</span>: claude-sonnet-4-6

  <span class="hl-kw">agent</span> <span class="hl-node">Start</span>
    <span class="hl-field">label</span>: Start

  <span class="hl-kw">agent</span> <span class="hl-node">Exit</span>
    <span class="hl-field">label</span>: Exit

  <span class="hl-kw">agent</span> <span class="hl-node">Summarize</span>
    <span class="hl-field">label</span>: <span class="hl-str">"Summarize Document"</span>
    <span class="hl-field">prompt</span>:
      Read the provided document carefully. Produce a concise summary
      that captures the key points, main arguments, and conclusions.
      Keep the summary under 200 words.

  <span class="hl-kw">agent</span> <span class="hl-node">Translate</span>
    <span class="hl-field">label</span>: <span class="hl-str">"Translate to Spanish"</span>
    <span class="hl-field">prompt</span>:
      Translate the summary from the previous step into Spanish.
      Preserve the meaning and tone. Use formal register.

  <span class="hl-kw">edges</span>
    Start -> Summarize
    Summarize -> Translate
    Translate -> Exit
</pre>

A few things to notice. Prompts are multi-line blocks introduced by `prompt:`
with indented content below -- no quoting, no escaping, no `\n` sequences.
That's the main reason Dippin exists.<sup>1</sup> Node types (`agent`,
`tool`, `human`) declare what kind of step each node is.
And edges live in a dedicated `edges` block at the bottom, separate from
the node definitions. You can try this yourself in the
[playground](../playground.html).

## <span class="step-num">3</span> Validate the Workflow

Run `dippin validate` to check structural correctness:

<pre>
<span class="hl-dim">$</span> <span class="hl-shcmd">dippin validate pipeline.dip</span>
<span class="hl-pass">PASS</span>  pipeline.dip
</pre>

Validation catches structural errors: missing start/exit nodes, references to
undefined nodes in edges, duplicate node names, unreachable nodes, and cycles.
These are the [DIP001-DIP010 error codes](../validation.html). If any
fire, the file is invalid.

Try breaking it -- rename `Exit` to `Done` in the header
but not in the node list:

<pre>
<span class="hl-dim">$</span> <span class="hl-shcmd">dippin validate pipeline.dip</span>
<span class="hl-fail">FAIL</span>  pipeline.dip
  <span class="hl-fail">DIP002</span>: exit node "Done" is not defined
</pre>

## <span class="step-num">4</span> Lint for Best Practices

Validation checks structure. Linting checks semantics -- missing timeouts, unknown
models, unreachable branches, and 22 other potential problems:

<pre>
<span class="hl-dim">$</span> <span class="hl-shcmd">dippin lint pipeline.dip</span>
<span class="hl-pass">PASS</span>  pipeline.dip  <span class="hl-dim">(0 errors, 0 warnings)</span>
</pre>

Clean. If you'd used a nonexistent model name, you'd see:

<pre>
<span class="hl-dim">$</span> <span class="hl-shcmd">dippin lint pipeline.dip</span>
<span class="hl-warn">WARN</span>  pipeline.dip
  <span class="hl-warn">DIP108</span>: node "Summarize": unknown model "claude-5-opus" for provider "anthropic"
</pre>

70 diagnostic codes in total: 10 structural errors (DIP001-DIP010) and
60 semantic warnings (DIP101-DIP160). Run `dippin explain DIP108`
to see what any code means.

## <span class="step-num">5</span> Format for Consistency

The formatter produces canonical output with deterministic field ordering. It's
idempotent: formatting already-formatted code produces identical output.

<pre>
<span class="hl-dim">$</span> <span class="hl-shcmd">dippin fmt pipeline.dip</span>
</pre>

Prints the formatted version to stdout. Use `--write` to format
in place, or `--check` to verify formatting without changing the file
(handy for [CI](ci-integration.html)):

<pre>
<span class="hl-dim">$</span> <span class="hl-shcmd">dippin fmt --check pipeline.dip</span>
<span class="hl-pass">OK</span>  pipeline.dip
</pre>

## <span class="step-num">6</span> Visualize with DOT Export

Export to [Graphviz](https://graphviz.org/) DOT format and render a diagram:

<pre>
<span class="hl-dim">$</span> <span class="hl-shcmd">dippin export-dot pipeline.dip | dot -Tpng -o pipeline.png</span>
</pre>

For left-to-right layout (better for wide pipelines):

<pre>
<span class="hl-dim">$</span> <span class="hl-shcmd">dippin export-dot --rankdir=LR pipeline.dip | dot -Tpng -o pipeline.png</span>
</pre>

Or view the topology right in your terminal:

<pre>
<span class="hl-dim">$</span> <span class="hl-shcmd">dippin graph pipeline.dip</span>

  &#x250c;&#x2500;&#x2500;&#x2500;&#x2500;&#x2500;&#x2500;&#x2500;&#x2510;   &#x250c;&#x2500;&#x2500;&#x2500;&#x2500;&#x2500;&#x2500;&#x2500;&#x2500;&#x2500;&#x2500;&#x2500;&#x2510;   &#x250c;&#x2500;&#x2500;&#x2500;&#x2500;&#x2500;&#x2500;&#x2500;&#x2500;&#x2500;&#x2500;&#x2500;&#x2510;   &#x250c;&#x2500;&#x2500;&#x2500;&#x2500;&#x2500;&#x2500;&#x2510;
  &#x2502; Start &#x2502;&#x2500;&#x2500;&#x25b6;&#x2502; Summarize &#x2502;&#x2500;&#x2500;&#x25b6;&#x2502; Translate  &#x2502;&#x2500;&#x2500;&#x25b6;&#x2502; Exit &#x2502;
  &#x2514;&#x2500;&#x2500;&#x2500;&#x2500;&#x2500;&#x2500;&#x2500;&#x2518;   &#x2514;&#x2500;&#x2500;&#x2500;&#x2500;&#x2500;&#x2500;&#x2500;&#x2500;&#x2500;&#x2500;&#x2500;&#x2518;   &#x2514;&#x2500;&#x2500;&#x2500;&#x2500;&#x2500;&#x2500;&#x2500;&#x2500;&#x2500;&#x2500;&#x2500;&#x2518;   &#x2514;&#x2500;&#x2500;&#x2500;&#x2500;&#x2500;&#x2500;&#x2518;
</pre>

## <span class="step-num">7</span> Add a Tool Node

Dippin supports three node kinds: `agent` (LLM call),
`tool` (shell command), and `human` (user input).
Here's a tool node that prepares the workspace:

<pre>
<span class="hl-kw">tool</span> <span class="hl-node">SetupWorkspace</span>
  <span class="hl-field">label</span>: <span class="hl-str">"Setup Workspace"</span>
  <span class="hl-field">timeout</span>: 30s
  <span class="hl-field">command</span>:
    set -eu
    mkdir -p .ai/output
    printf 'workspace-ready'
</pre>

Tool nodes run shell commands directly. The `timeout` field matters --
without it, `dippin lint` emits a DIP111 warning.
The `command:` block works like `prompt:`: indent
your script and write any shell syntax you need. See the
[language reference](../language.html) for the full field list.

## <span class="step-num">8</span> Add Conditional Routing

Edges can carry conditions that control which path is taken at runtime:

<pre>
<span class="hl-kw">edges</span>
  Start -> SetupWorkspace
  SetupWorkspace -> Summarize  <span class="hl-cond">when</span> ctx.outcome = success
  SetupWorkspace -> Exit       <span class="hl-cond">when</span> ctx.outcome = fail
  Summarize -> Translate
  Translate -> Exit
</pre>

The `when` keyword introduces a condition expression. Conditions
check context variables set at runtime. When a node's outgoing edges have
complementary conditions (like `success`/`fail`), Dippin
recognizes them as exhaustive and suppresses the DIP101 "unconditional fallback
missing" warning.<sup>2</sup>

## <span class="step-num">9</span> Run the Full Check Suite

`dippin check` runs validate + lint together and can output JSON for
programmatic consumption:

<pre>
<span class="hl-dim">$</span> <span class="hl-shcmd">dippin check pipeline.dip</span>
<span class="hl-pass">PASS</span>  pipeline.dip  <span class="hl-dim">valid=true errors=0 warnings=0</span>
</pre>

<pre>
<span class="hl-dim">$</span> <span class="hl-shcmd">dippin --format json check pipeline.dip</span>
{
  "file": "pipeline.dip",
  "valid": true,
  "errors": 0,
  "warnings": 0,
  "diagnostics": []
}
</pre>

## <span class="step-num">10</span> Explore the Examples

The repository ships with real-world examples in the
[`examples/`](https://github.com/2389-research/dippin-lang/tree/main/examples)
directory. A few worth studying:

| File | What it demonstrates |
|------|---------------------|
| `ask_and_execute.dip` | Human input, parallel implementation across 3 providers, cross-review |
| `code_quality_sweep.dip` | 3-provider parallel analysis, fan-out/fan-in, goal gates, retry loops |
| `consensus_task.dip` | Multi-model consensus with postmortem and bounded restarts |
| `complexity_cleanup.dip` | Tool nodes with shell scripts, conditional routing, iterative refinement |
| `human_gate_showcase.dip` | Human choice and freeform gates with preferred_label for testing |

Validate them all at once:

<pre>
<span class="hl-dim">$</span> <span class="hl-shcmd">for f in examples/*.dip; do dippin validate "$f"; done</span>
<span class="hl-pass">PASS</span>  examples/ask_and_execute.dip
<span class="hl-pass">PASS</span>  examples/code_quality_sweep.dip
<span class="hl-pass">PASS</span>  examples/complexity_cleanup.dip
<span class="hl-pass">PASS</span>  examples/consensus_task.dip
<span class="hl-pass">PASS</span>  examples/human_gate_showcase.dip
<span class="hl-dim">...</span>
</pre>

## Quick Reference: Essential Commands

| Command | What it does |
|---------|-------------|
| `dippin validate file.dip` | Structural correctness (DIP001-DIP010) |
| `dippin lint file.dip` | Semantic warnings (DIP101-DIP160) |
| `dippin check file.dip` | Validate + lint combined |
| `dippin fmt file.dip` | Canonical formatting to stdout |
| `dippin fmt --write file.dip` | Format in place |
| `dippin export-dot file.dip` | Convert to Graphviz DOT |
| `dippin graph file.dip` | ASCII topology in terminal |
| `dippin test file.dip` | Run scenario tests |
| `dippin cost file.dip` | Estimate per-run cost |
| `dippin doctor file.dip` | Health report with letter grade |
| `dippin lsp` | Start LSP server for [editor integration](editor-setup.html) |
| `dippin explain DIPxxx` | Explain a diagnostic code |

## What's Next?

You've got a working toolchain and your first workflow. Pick whichever fits
where you're headed:

<div class="footnotes">
  <h3>Notes</h3>
  <ol>
    <li id="fn1">DOT requires prompts to be a single quoted string with <code>\n</code> for newlines and <code>\"</code> for inner quotes. For a 20-line system prompt, that's unreadable. Dippin's indentation-based blocks were the original motivation for building the language. See <a href="migrating-from-dot.html">Migrating from DOT</a> for a side-by-side comparison.</li>
    <li id="fn2">Exhaustive condition detection also handles <code>contains</code>/<code>not-contains</code> complementary pairs, not just <code>success</code>/<code>fail</code>. The implementation lives in <a href="https://github.com/2389-research/dippin-lang/tree/main/validator">validator/lint.go</a>.</li>
  </ol>
</div>
