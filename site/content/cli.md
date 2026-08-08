---
title: "CLI Reference"
description: "Complete command reference for the Dippin toolchain: 18 commands for authoring, export, analysis, and bundling AI pipeline workflows — parse, validate, lint, check, fmt, simulate, cost, coverage, doctor, test, watch, pack, unpack, inspect, and more."
section_label: "Reference"
subtitle: "Every command in the dippin toolchain — authoring, export, analysis, and bundles."
---

## Global Usage

```
dippin [--format text|json] <command> [args]
```

### Global Flags

| Flag | Values | Default | Description |
|------|--------|---------|-------------|
| `--format` | `text`, `json` | `text` | Output format for diagnostics. `text` produces human-readable output. `json` produces machine-readable arrays for CI/tooling integration. |

### Exit Codes

Analysis commands:

| Code | Meaning |
|------|---------|
| `0` | Success — no issues found, operation completed |
| `1` | Error — validation failures, parse errors, check-mode drift, parity mismatches |
| `2` | Usage error — bad flags, missing arguments, unknown command |

Bundle commands (`pack`, `unpack`, `inspect`) use a finer ladder so tooling can distinguish integrity failures from I/O failures:

| Code | Meaning |
|------|---------|
| `0` | Ok |
| `1` | User error (parse failure, invalid input) |
| `2` | Bundle integrity failure (hash mismatch, manifest invalid, forbidden ZIP feature, truncation, unsupported format) |
| `3` | I/O error (write failure during pack, rename failure during unpack) |
| `4` | Cancelled (`context.Canceled` / `context.DeadlineExceeded`) |

## Authoring Commands

<div class="group-badge lavender">Authoring</div>

<div class="cmd-card">
  <h3>parse</h3>
  <div class="cmd-usage">dippin parse &lt;file&gt;</div>
  <p>Parse a workflow file and output the intermediate representation (IR) as JSON. Useful for debugging, tooling integration, and inspecting how the parser interprets your workflow. Accepts <code>.dip</code> or <code>.dot</code> files (auto-detected by extension).</p>
</div>

<div class="cmd-card">
  <h3>validate</h3>
  <div class="cmd-usage">dippin validate &lt;file&gt;</div>
  <p>Run structural validation checks (DIP001-DIP010) on a workflow. Outputs "validation passed" or diagnostic messages. Exit code 1 if any errors found.</p>
</div>

<div class="cmd-card">
  <h3>lint</h3>
  <div class="cmd-usage">dippin lint [--extra-models &lt;spec&gt;] &lt;file&gt;</div>
  <p>Run both structural validation and semantic linting (DIP001-DIP010 + DIP101-DIP157). All 67 diagnostic rules. Errors cause exit code 1; warnings alone exit 0.</p>
  <dl>
    <dt><code>--extra-models "provider:model1,model2;provider2:model3"</code></dt>
    <dd>Extend the DIP108 model catalog at runtime for private or newly-released models.</dd>
  </dl>
</div>

<div class="cmd-card">
  <h3>check</h3>
  <div class="cmd-usage">dippin check [--format json|text] &lt;file&gt;</div>
  <p>Parse, validate, and lint in one shot. Designed for LLM tool-calling loops and CI. Defaults to JSON output with <code>valid</code>, <code>errors</code>, <code>warnings</code>, <code>diagnostics</code>, and <code>suggested_actions</code> fields.</p>
</div>

<div class="cmd-card">
  <h3>fmt</h3>
  <div class="cmd-usage">dippin fmt [--check] [--write] [--migrate] &lt;file&gt;</div>
  <p>Format a <code>.dip</code> file to canonical form. 2-space indentation, standard field ordering, deterministic and idempotent output. Use <code>--check</code> for CI (exit 1 if unformatted) or <code>--write</code> for in-place formatting. It also strips redundant <code>parallel</code>/<code>fan_in</code> fan edges re-declared in the <code>edges</code> block — the inline node list is authoritative (<code>DIP153</code>). <code>--migrate</code> converts a v1 file to <code>dip 2</code>: it folds <code>fallback_target</code> into an <code>on fail</code> edge and a non-self <code>retry_target</code> into a <code>loop</code> edge, flagging any case it can't express 1:1 with an inline <code># MIGRATION:</code> comment plus a stderr summary. Exit code <code>3</code> means a case needs review but the output is runtime-equivalent; <code>4</code> means the output is NOT runtime-equivalent (a non-self <code>retry_target</code>, whose loop edge the retry engine does not read — see <a href="https://github.com/2389-research/dippin-lang/issues/186">#186</a>) and must not be used as a drop-in. <code>--migrate --check</code> exits non-zero when a file is not already canonical <code>dip 2</code>.</p>
</div>

<div class="cmd-card">
  <h3>new</h3>
  <div class="cmd-usage">dippin new [--name &lt;name&gt;] [--write &lt;file&gt;] &lt;template&gt;</div>
  <p>Generate a starter <code>.dip</code> file from a built-in template. Available templates: <code>minimal</code>, <code>parallel</code>, <code>conditional</code>, <code>review-loop</code>, <code>human-gate</code>. Output always passes <code>dippin validate</code>.</p>
</div>

## Export Commands

<div class="group-badge green">Export</div>

<div class="cmd-card">
  <h3>export-dot</h3>
  <div class="cmd-usage">dippin export-dot [--rankdir=LR|TB] [--prompts] &lt;file&gt;</div>
  <p>Export a workflow to Graphviz DOT format for visualization. Maps node kinds to DOT shapes (agent=box, human=hexagon, tool=parallelogram). Goal gate nodes get red background; restart edges are dashed.</p>
</div>

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

## Analysis Commands

<div class="group-badge yellow">Analysis</div>

<div class="cmd-card">
  <h3>simulate</h3>
  <div class="cmd-usage">dippin simulate [--scenario key=val] [--interactive] [--all-paths] &lt;file&gt;</div>
  <p>Dry-run a workflow's execution graph without calling LLMs or running commands. Emits JSONL events (pipeline_start, node_enter, node_exit, edge_traverse, pipeline_end). Use <code>--scenario</code> to inject context values and <code>--all-paths</code> to enumerate all possible paths.</p>
</div>

<div class="cmd-card">
  <h3>cost</h3>
  <div class="cmd-usage">dippin cost &lt;file&gt;</div>
  <p>Estimate workflow execution cost based on model pricing tables. Per-node cost breakdown with turn and token heuristics.</p>
</div>

<div class="cmd-card">
  <h3>coverage</h3>
  <div class="cmd-usage">dippin coverage &lt;file&gt;</div>
  <p>Analyze edge coverage and reachability. Reports tool output extraction, edge condition matching, and termination analysis.</p>
</div>

<div class="cmd-card">
  <h3>doctor</h3>
  <div class="cmd-usage">dippin doctor [--extra-models &lt;spec&gt;] &lt;file&gt;</div>
  <p>Health report card aggregating lint, coverage, and cost into a letter grade (A-F). Generates actionable suggestions.</p>
  <dl>
    <dt><code>--extra-models "provider:model1,model2;provider2:model3"</code></dt>
    <dd>Extend the DIP108 model catalog at runtime for private or newly-released models.</dd>
  </dl>
</div>

<div class="cmd-card">
  <h3>test</h3>
  <div class="cmd-usage">dippin test [--verbose] [--coverage] &lt;file.dip&gt;</div>
  <p>Run scenario tests defined in <code>.test.json</code> files against a workflow. Auto-discovers the test file from the workflow path. Use <code>--verbose</code> to show execution paths. Use <code>--coverage</code> to report node and edge coverage across all test scenarios.</p>
</div>

<div class="cmd-card">
  <h3>watch</h3>
  <div class="cmd-usage">dippin watch &lt;file-or-dir&gt; [...]</div>
  <p>Watch <code>.dip</code> files or directories for changes. On each change it parses, validates, and lints the affected file. Debounces rapid saves (200ms).</p>
</div>

## Bundle Commands

<div class="group-badge lavender">Bundles</div>

<div class="cmd-card">
  <h3>pack</h3>
  <div class="cmd-usage">dippin pack [-o &lt;out&gt;] [--dry-run] [--no-inline] [--include &lt;path&gt;...] &lt;entry.dip&gt;</div>
  <p>Build a deterministic <code>.dipx</code> bundle from a <code>.dip</code> entry, walking every transitively-reachable subgraph ref. Runs structural validation (DIP001–DIP010) before packing. <code>-o -</code> writes to stdout; <code>--dry-run</code> validates and walks refs without writing. File output is atomic via <code>os.CreateTemp</code> + rename. Refuses symlinks anywhere in the source tree, including parent components.</p>
  <p>By default (inline mode) every <code>command_file:</code>, <code>prompt_file:</code>, <code>system_prompt_file:</code>, <code>prompt_include:</code>, and defaults <code>prompt_prefix_file:</code>/<code>prompt_suffix_file:</code> fragment is inlined (prompt fragments compose into the agent prompt) into the packed <code>.dip</code>, producing a self-contained <code>format_version 1</code> bundle. <code>--no-inline</code> instead ships those directive targets as separate entries under <code>workflows/</code> and keeps the <code>*_file:</code> directives, so they resolve against the extracted tree exactly as in a source-tree run (<code>format_version 2</code>). <code>--include &lt;path&gt;</code> (repeatable, requires <code>--no-inline</code>) ships extra sibling files or directories — assets referenced only from inside shell bodies — as a single file or a whole directory tree; a path that resolves to a <code>.dip</code> is an error.</p>
</div>

<div class="cmd-card">
  <h3>unpack</h3>
  <div class="cmd-usage">dippin unpack [-o &lt;dir&gt;] [--force] &lt;bundle.dipx&gt;</div>
  <p>Extract a <code>.dipx</code> bundle into a directory atomically. Uses staging dir + rename. <code>--force</code> overwrites an existing destination via a backup-aside / rename-into-place / remove-aside sequence so the original is preserved if the swap fails (e.g. cross-mount EXDEV).</p>
</div>

<div class="cmd-card">
  <h3>inspect</h3>
  <div class="cmd-usage">dippin inspect [--no-verify] [--format text|json] &lt;bundle.dipx&gt;</div>
  <p>Print a bundle's manifest, identity hash (SHA-256 over the manifest bytes-as-stored), and per-file checksums. Integrity-verifies by default; <code>--no-verify</code> skips hash verification (forensic mode).</p>
</div>
