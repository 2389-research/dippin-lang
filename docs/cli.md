# CLI Reference

The `dippin` command-line tool provides parsing, validation, formatting, export, migration, analysis, and editor integration for `.dip` workflow files.

---

## Installation

Build from source:

```bash
go build -o dippin ./cmd/dippin
```

---

## Global Usage

```mermaid
graph LR
    CLI["dippin"] --> |"--format text\|json"| CMD{Command}
    CMD --> A["Authoring<br>parse, validate, lint,<br>check, fmt, new"]
    CMD --> B["Export / Migration<br>export-dot, export-dip,<br>migrate, validate-migration"]
    CMD --> C["Analysis<br>simulate, cost, coverage,<br>doctor, optimize, diff,<br>feedback, unused, graph"]
    CMD --> D["Editor<br>lsp"]
    CMD --> E["Info / Testing<br>version, help, explain, test, spec"]
```

```
dippin [--format text|json] <command> [args]
```

### Global Flags

| Flag | Values | Default | Description |
|------|--------|---------|-------------|
| `--format` | `text`, `json` | `text` | Output format for diagnostics. `text` produces human-readable output. `json` produces machine-readable arrays for CI/tooling integration. |

---

## Exit Codes

Analysis commands use the consistent triple:

| Code | Meaning |
|------|---------|
| `0` | Success — no issues found, operation completed |
| `1` | Error — validation failures, parse errors, check-mode drift, parity mismatches |
| `2` | Usage error — bad flags, missing arguments, unknown command |

Bundle commands (`pack`, `unpack`, `inspect`) use a finer ladder so tooling can distinguish integrity failures from I/O failures:

| Code | Meaning |
|------|---------|
| `0` | Ok |
| `1` | User error (parse failure, invalid input, missing flag arg) |
| `2` | Bundle integrity failure (`ErrHashMismatch`, `ErrManifestInvalid`, `ErrZipFeatureForbidden`, `ErrZipTruncated`, `ErrUnsupportedFormatVersion`) |
| `3` | I/O error (write failure during pack, rename failure during unpack) |
| `4` | Cancelled (`context.Canceled` / `context.DeadlineExceeded`) |

---

## Commands

### parse

Parse a workflow file and output the intermediate representation (IR) as JSON.

```bash
dippin parse <file>
```

**Input**: `.dip` or `.dot` file (auto-detected by extension)

**Output**: Indented JSON representation of the `ir.Workflow` struct, printed to stdout. This is useful for debugging, tooling integration, and inspecting how the parser interprets your workflow.

**Example**:
```bash
dippin parse pipeline.dip
```

```json
{
  "Name": "my_pipeline",
  "Version": "",
  "Goal": "Do a thing",
  "Start": "Ask",
  "Exit": "Done",
  "Nodes": [ ... ],
  "Edges": [ ... ]
}
```

---

### validate

Run structural validation checks (DIP001–DIP010) on a workflow.

```bash
dippin validate <file>
```

**Input**: `.dip` or `.dot` file

**Checks**: The 10 structural validation rules that must pass for a workflow to be executable. See [validation.md](validation.md) for details on each code.

**Output**:
- If all checks pass: `"validation passed"` (text mode) or empty JSON array
- If errors found: diagnostic messages to stderr

**Example**:
```bash
$ dippin validate pipeline.dip
validation passed

$ dippin validate broken.dip
error[DIP003]: unknown node reference "InterpretX" in edge
  --> broken.dip:45:5
  = help: did you mean "Interpret"?
```

---

### lint

Run both structural validation and semantic linting (DIP001–DIP010 + DIP101–DIP157).

```bash
dippin lint [--extra-models <spec>] <file>
```

**Input**: `.dip` or `.dot` file

**Checks**: All 67 diagnostic rules. Errors (DIP001–DIP010) cause exit code 1. Warnings (DIP101–DIP157) are reported but don't affect the exit code.

**Output**: All diagnostics (errors and warnings) to stderr.

**Flags**:

| Flag | Description |
|------|-------------|
| `--extra-models "provider:model1,model2;provider2:model3"` | Extend the DIP108 model catalog at runtime for private or newly-released models. |

**Example**:
```bash
$ dippin lint pipeline.dip
warning[DIP111]: tool command has no timeout
  --> pipeline.dip:35:3

$ echo $?
0    # warnings don't cause failure
```

---

### check

Parse, validate, and lint a workflow in one shot. Designed for LLM tool-calling loops and CI.

```bash
dippin check [--format json|text] <file>
```

**Default format**: `json` (unlike other commands which default to `text`). The `check` command parses its own `--format` flag, ignoring the global default.

**Output** (to stdout, not stderr):

```json
{
  "valid": false,
  "errors": 1,
  "warnings": 2,
  "diagnostics": [
    {"code": "DIP003", "severity": "error", "message": "unknown node reference \"Nope\" in edge", "line": 19, "fix": ""}
  ],
  "suggested_actions": []
}
```

| Field | Description |
|-------|-------------|
| `valid` | `true` if no errors (warnings allowed). Use this to decide whether to retry generation. |
| `errors` | Count of error-severity diagnostics |
| `warnings` | Count of warning-severity diagnostics |
| `diagnostics` | Array of all findings (code, severity, message, line, fix) |
| `suggested_actions` | Deduplicated non-empty `fix` strings from diagnostics |

**Text mode**: `dippin check --format text pipeline.dip` produces human-readable diagnostic output to stdout.

**Example**:
```bash
# LLM tool-calling loop:
dippin check generated.dip
# → {"valid":true,"errors":0,"warnings":0,"diagnostics":[],"suggested_actions":[]}

# Feed errors back to LLM for correction:
dippin check broken.dip | jq '.diagnostics[] | .message'
```

---

### new

Generate a starter `.dip` file from a built-in template.

```bash
dippin new [--name <name>] [--write <file>] <template>
```

**Flags**:

| Flag | Description |
|------|-------------|
| `--name <name>` | Override the workflow name (default: template name) |
| `--write <file>` | Write output to file instead of stdout |

**Available templates**:

| Template | Topology |
|----------|----------|
| `minimal` | `Start` → `Done` (two agent nodes) |
| `parallel` | `Init` → `parallel` → 2 workers → `fan_in` → `Done` |
| `conditional` | `Check` (auto_status) → `Pass`/`Fail` branching → `Done` |
| `review-loop` | `Implement` → `Review` → success: `Done` / fail: restart to `Implement` |
| `human-gate` | `Prepare` → human `Gate` (choice) → approve/reject → `Done` |

Templates are built programmatically from IR types and formatted via `formatter.Format()`, guaranteeing canonical output that always passes `dippin validate`.

**Examples**:
```bash
# Preview a template:
dippin new parallel

# Create a named workflow:
dippin new --name MyPipeline conditional --write my_pipeline.dip

# Generate and immediately validate:
dippin new review-loop --write /tmp/test.dip && dippin validate /tmp/test.dip
```

---

### fmt

Format a `.dip` file to canonical form.

```bash
dippin fmt [--check] [--write] [--migrate] <file>
```

**Flags**:

| Flag | Description |
|------|-------------|
| `--check` | Don't output anything. Exit 1 if the file is not already in canonical format. Useful for CI checks. |
| `--write` | Write the formatted output back to the source file in-place. |
| `--migrate` | Convert a v1 file to `dip 2` and emit canonical output. Folds `fallback_target` into an `on fail` edge and a non-self `retry_target` into a `loop` edge, flagging any case it can't express 1:1 with an inline `# MIGRATION:` comment plus a stderr summary. Exit codes: `3` when a case needs author review but the output is runtime-equivalent; **`4` when the output is NOT runtime-equivalent** — currently a non-self `retry_target`, whose loop edge the engine's retry channel does not read (the retry silently becomes a self-retry), so the output must not be used as a drop-in ([#186](https://github.com/2389-research/dippin-lang/issues/186)). `--migrate --check` writes nothing and exits non-zero when the file is not already canonical `dip 2`. An already-`dip 2` file is an idempotent no-op. |

**Default behavior** (no flags): Print the canonically formatted output to stdout.

**What canonical format means**:
- 2-space indentation
- Workflow header fields in standard order (goal, requires, start, exit)
- Defaults block (if present) after header
- Node definitions ordered by kind
- Edges section at end (redundant `parallel`/`fan_in` fan edges re-declared here are stripped — the inline node list is authoritative; see `DIP153`)
- Multiline fields (prompt, command) indented with `:` on the same line
- Deterministic, idempotent — formatting an already-formatted file produces identical output

**Examples**:
```bash
# Preview formatted output:
dippin fmt pipeline.dip

# Check in CI (fails if not formatted):
dippin fmt --check pipeline.dip

# Format in place:
dippin fmt --write pipeline.dip
```

---

### export-dot

Export a workflow to Graphviz DOT format for visualization.

```bash
dippin export-dot [--rankdir=LR|TB] [--prompts] <file>
```

**Flags**:

| Flag | Values | Default | Description |
|------|--------|---------|-------------|
| `--rankdir` | `LR`, `TB` | `TB` | Graph layout direction. `TB` = top-to-bottom, `LR` = left-to-right. |
| `--prompts` | — | off | Include full prompt and command text as DOT node attributes. By default, prompts are omitted for cleaner visualization. |

**Node Shape Mapping**:

| Node Kind | DOT Shape | Note |
|-----------|-----------|------|
| `agent` | `box` | |
| `human` | `hexagon` | |
| `tool` | `parallelogram` | |
| `parallel` | `component` | |
| `fan_in` | `tripleoctagon` | |
| `subgraph` | `tab` | |
| start node | `Mdiamond` | Regardless of kind |
| exit node | `Msquare` | Regardless of kind |

**Special styling**:
- Goal gate nodes get a red filled background
- Restart edges are rendered as dashed lines

**Example**:
```bash
# Generate DOT and render as PNG:
dippin export-dot pipeline.dip | dot -Tpng -o pipeline.png

# Left-to-right layout with prompts:
dippin export-dot --rankdir=LR --prompts pipeline.dip > pipeline.dot
```

---

### export-dip

Flatten all transitively-reachable subgraph refs into a single canonical `.dip` file and print to stdout.

```bash
dippin export-dip <file>
```

**Input**: `.dip` or `.dipx` file.

**What it does**: Resolves every `subgraph` node's referenced `.dip` file recursively, inlines all reachable nodes and edges into a single flat workflow, then formats the result as canonical `.dip` source. The output is suitable for inspection, archival, or passing to tools that don't support multi-file workflows.

**Output**: Canonical `.dip` text to stdout.

**Example**:
```bash
# Preview the flattened workflow:
dippin export-dip pipeline.dip

# Write to file:
dippin export-dip pipeline.dip > flat_pipeline.dip
```

---

### migrate

Convert a DOT file to `.dip` source format.

```bash
dippin migrate [--output <file>] <file.dot>
```

**Flags**:

| Flag | Description |
|------|-------------|
| `--output <file>` | Write output to the specified file instead of stdout. |

**What it does**:
1. Parses the DOT file using a custom DOT parser
2. Maps DOT shapes to Dippin node kinds
3. Extracts graph-level attributes into a `defaults` block
4. Unescapes prompt and command text from DOT string encoding
5. Prefixes bare condition variables with `ctx.` namespace
6. Identifies start/exit from `Mdiamond`/`Msquare` shapes
7. Outputs canonical `.dip` source

**Example**:
```bash
# Preview the migration:
dippin migrate old_pipeline.dot

# Write to file:
dippin migrate --output new_pipeline.dip old_pipeline.dot
```

---

### validate-migration

Check structural parity between a DOT file and a `.dip` file to verify migration correctness.

```bash
dippin validate-migration <old.dot> <new.dip>
```

**What it checks**: Compares the IR produced from both files and reports structural differences — missing nodes, different edges, changed conditions, etc.

**Output**:
- If equivalent: `"migration parity check passed"`
- If differences found: List of differences with categories

**Example**:
```bash
$ dippin validate-migration old.dot new.dip
migration parity check passed

$ dippin validate-migration old.dot broken.dip
parity check failed: 2 difference(s) found
  [node] missing node "ReviewStep" in new file
  [edge] edge Validate->Approve has different condition
```

---

### simulate

Dry-run a workflow's execution graph without calling LLMs or running commands. Emits JSONL events to stdout.

```bash
dippin simulate [--scenario key=val] [--interactive] [--all-paths] <file>
```

**Flags**:

| Flag | Description |
|------|-------------|
| `--scenario key=val` | Inject context values (repeatable). Use `NodeID.key=val` for per-node overrides. |
| `--interactive` | Prompt at human nodes instead of auto-selecting |
| `--all-paths` | Enumerate all possible paths through the graph |

**Output**: JSONL (one JSON object per line) with event types:

| Event | Fields | Description |
|-------|--------|-------------|
| `pipeline_start` | `run_id`, `workflow` | Simulation begins |
| `node_enter` | `node`, `kind`, `model`, `provider`, `prompt` | Node execution starts |
| `node_exit` | `node`, `status`, `duration_ms` | Node execution ends |
| `edge_traverse` | `from`, `to` | Edge followed |
| `pipeline_end` | `status`, `nodes_visited` | Simulation ends |

**Examples**:
```bash
# Default (all nodes succeed):
dippin simulate pipeline.dip

# Explore failure path:
dippin simulate pipeline.dip --scenario outcome=fail

# Fail only at a specific node:
dippin simulate pipeline.dip --scenario Validate.outcome=fail

# Count all possible paths:
dippin simulate pipeline.dip --all-paths
```

---

### cost

Estimate workflow execution cost based on model pricing tables.

```bash
dippin cost <file>
```

See [analysis.md](analysis.md#cost) for output format and JSON schema.

---

### coverage

Analyze edge coverage and reachability.

```bash
dippin coverage <file>
```

See [analysis.md](analysis.md#coverage) for output format and JSON schema.

---

### doctor

Health report card aggregating lint, coverage, and cost into a letter grade (A–F).

```bash
dippin doctor [--extra-models <spec>] <file>
```

**Flags**:

| Flag | Description |
|------|-------------|
| `--extra-models "provider:model1,model2;provider2:model3"` | Extend the DIP108 model catalog at runtime for private or newly-released models. |

See [analysis.md](analysis.md#doctor) for scoring and output format.

---

### optimize

Suggest cheaper model substitutions.

```bash
dippin optimize <file>
```

See [analysis.md](analysis.md#optimize) for rules applied and output format.

---

### diff

Semantic comparison between two workflow versions.

```bash
dippin diff <old.dip> <new.dip>
```

See [analysis.md](analysis.md#diff) for output format and JSON schema.

---

### feedback

Compare predicted costs against actual execution telemetry.

```bash
dippin feedback <workflow.dip> <telemetry.csv>
```

See [analysis.md](analysis.md#feedback) for input format and output schema.

---

### explain

Print a detailed explanation for any diagnostic code.

```bash
dippin explain <DIPxxx>
```

**Input**: A diagnostic code string (e.g., `DIP101`, `DIP003`).

**Output**: Rich explanation including what triggers the diagnostic, how to fix it, and an example snippet.

**Example**:
```bash
$ dippin explain DIP101
═══ DIP101 ═════════════════════════════════════════
  unreachable node after conditional branches

  Trigger: A node follows conditional branches that already cover all outcomes.
  Fix:     Route the node through a condition, add a section `else -> <node>` default that covers it, or remove it if unreachable.

  Example:
    A -> B when ctx.outcome = success
    A -> C when ctx.outcome = fail
    A -> D  # unreachable
```

In JSON mode, outputs the `Explanation` struct with `code`, `summary`, `trigger`, `fix`, and `example` fields.

---

### unused

Detect dead-branch nodes — reachable from start but with no path to exit — and estimate wasted cost.

```bash
dippin unused <file>
```

**Input**: `.dip` or `.dot` file.

**Output**: List of unused (sink) nodes with their kind, label, and estimated wasted cost per run.

**Example**:
```bash
$ dippin unused pipeline.dip
═══ Unused Nodes ══════════════════════════════════════════
  ✗ DeadEnd                      agent  (Dead End)
─── Wasted Cost ───────────────────────────────────────────
  $0.05 - $0.12 estimated wasted per run
─── Summary ───────────────────────────────────────────────
  1 unused node(s) found
```

---

### graph

Render a terminal-friendly ASCII DAG of the workflow.

```bash
dippin graph [--compact] <file>
```

**Flags**:

| Flag | Description |
|------|-------------|
| `--compact` | Single-line format: `[Start] → [Middle] → [Exit]` |

**Output**: ASCII box-drawing diagram showing topological layers. In JSON mode, outputs the layer structure.

**Example**:
```bash
$ dippin graph pipeline.dip
  ┌──────────┐
  │  Start   │
  └──────────┘
       │
       ▼
  ┌──────────┐
  │  Middle  │
  └──────────┘
       │
       ▼
  ┌──────────┐
  │   Exit   │
  └──────────┘

$ dippin graph --compact pipeline.dip
[Start] → [Middle] → [Exit]
```

---

### test

Run scenario tests defined in `.test.json` files against a workflow.

```bash
dippin test [--verbose] [--coverage] <file.dip>
```

**Flags**:

| Flag | Description |
|------|-------------|
| `--verbose` | Show the execution path for each test case |
| `--coverage` | Show edge coverage summary after test results |

**Input**: A `.dip` file. The runner auto-discovers the corresponding `.test.json` file (e.g., `pipeline.dip` → `pipeline.test.json`).

**Test file format**: See [testing.md](testing.md) for the `.test.json` schema.

**Output**:
```bash
$ dippin test pipeline.dip
═══ Test Results ═════════════════════════════════════════════
  PASS  happy path
  PASS  error path
─── Summary ───────────────────────────────────────────────────
  2 tests: 2 passed, 0 failed
```

**Exit code**: 0 if all tests pass, 1 if any fail. JSON mode outputs the full `SuiteResult`.

---

### watch

Watch `.dip` files for changes and re-run lint on save.

```bash
dippin watch <file-or-dir> [...]
```

**Input**: One or more `.dip` files or directories containing `.dip` files.

**Behavior**: Uses `fsnotify` to monitor for writes and creates. Changes are debounced (200ms) to avoid duplicate runs. On each change, the file is parsed, validated, and linted, with results printed to the terminal.

**Example**:
```bash
# Watch a directory:
dippin watch examples/

# Watch specific files:
dippin watch pipeline.dip review.dip
```

---

### lsp

Start a Language Server Protocol server on stdio.

```bash
dippin lsp
```

Provides real-time diagnostics, hover, go-to-definition, autocomplete, and document symbols. See [editor-setup.md](editor-setup.md) for editor configuration.

---

### version

Show version information.

```bash
dippin version
```

---

### help

Display global usage and command list.

```bash
dippin help
```

---

### spec

Print the full embedded language specification to stdout.

```bash
dippin spec
```

**Output**: The complete dippin language spec (Markdown) to stdout. The content is generated at build time from the source documentation and embedded in the binary, so it always matches the installed version.

**Example**:
```bash
# View the spec:
dippin spec

# Page through it:
dippin spec | less

# Save a copy:
dippin spec > dippin-spec.md
```

---

### pack

Build a deterministic `.dipx` bundle from a `.dip` entry workflow plus every transitively-reachable subgraph ref.

```bash
dippin pack [-o <output>] [--dry-run] <entry.dip>
```

**Flags**:

- `-o <path>` — output path (default: `<entry>.dipx`; `-` writes the bundle to stdout)
- `--dry-run` — validate and walk refs without writing output

**Behavior**: Runs structural validation (DIP001–DIP010) on the entry first. Walks every transitively-reachable subgraph ref under the entry's directory, computes per-file SHA-256, and writes a deterministic ZIP (fixed mtimes, sorted entries, no platform metadata, UTF-8 filename bit always set). File output is atomic — the bundle is written to a unique temp file alongside `<output>` and renamed on success. Symlinks anywhere between the entry's source root and a leaf `.dip` are refused.

**Exit codes**: bundle ladder (see *Exit Codes* above).

---

### unpack

Extract a `.dipx` bundle into a directory atomically.

```bash
dippin unpack [-o <destdir>] [--force] <bundle.dipx>
```

**Flags**:

- `-o <dir>` — destination directory (default: `<bundle>` with the `.dipx` extension stripped)
- `--force` — overwrite an existing destination

**Atomicity**: writes contents to `<destdir>.tmp`, then swaps into place. With `--force`, the existing destination is renamed to `<destdir>.bak` before the swap; on a rename failure (e.g. cross-device EXDEV) the original is restored from the backup. On any extraction failure the staging directory is removed.

**Exit codes**: bundle ladder.

---

### inputs

Print a workflow's declared `inputs` schema — the typed, introspectable contract a host uses to collect values before a run.

```bash
dippin inputs [--format text|json] <file>
```

**Flags**:

- `--format text|json` — output format (default `text`); an unrecognized value is a usage error

**Output**: the declared inputs in declaration order. `--format json` emits a stable array a host can render into a form or walk conversationally — each entry carries `name`, `type`, `required`, and any declared attributes (`default`, `prompt`, `description`, `options`, `pattern`, `min`, `max`, `max_length`, `multiline`). Defaults are **typed** (a `number` default is a JSON number, a `bool` a JSON bool); a workflow with no `inputs` block emits `[]`, never `null`. See [Context and Variables](context.md) for the `${inputs.x}` namespace and the DIP155–DIP157 lints.

```bash
dippin inputs --format json build_product.dip
```

**Exit codes**: `0` ok; `1` load/parse error; `2` usage error.

---

### inspect

Print a `.dipx` bundle's manifest, identity hash, file list, and — for the verified (default) path — the entry workflow's declared `inputs` schema, so a host can enumerate what to collect without unpacking.

```bash
dippin inspect [--format text|json] [--no-verify] <bundle.dipx>
```

**Flags**:

- `--format text|json` — output format (default `text`)
- `--no-verify` — reserved; v1 always integrity-verifies (the flag prints a stderr advisory)

The bundle's `identity` is the SHA-256 of the manifest bytes as stored on disk — a stable content-addressable id suitable for caching and runtime resolution. The `--format json` payload includes an `inputs` array (the entry workflow's schema, same shape as `dippin inputs`); the manifest-only `--no-verify` path omits the key.

**Exit codes**: bundle ladder.

---

## JSON Output Mode

All commands support `--format json` for machine-readable output. Diagnostics are emitted as a JSON array:

```bash
dippin --format json lint pipeline.dip
```

```json
[
  {
    "code": "DIP111",
    "severity": "warning",
    "message": "tool command has no timeout",
    "location": {
      "file": "pipeline.dip",
      "line": 35,
      "column": 3,
      "end_line": 0,
      "end_column": 0
    },
    "help": "add a timeout field"
  }
]
```

---

## Auto-Detection

The CLI auto-detects input format by file extension:
- `.dip` — Parsed by the Dippin parser
- `.dipx` — Loaded via `dipx.Load`: hash-verified ZIP bundle; the entry workflow is parsed and fed to the analyzer. Supported by `validate`, `lint`, `doctor`, `parse`, `cost`, `coverage`, `simulate`, `optimize`, `unused`, `graph`, `diff`, `check`, `explain`, `export-dot`.
- `.dot` — Parsed by the DOT migration parser

This applies to all commands that accept a file argument. The `migrate` command always expects a `.dot` input, `fmt` always expects `.dip`, and `pack` always expects a `.dip` entry.
