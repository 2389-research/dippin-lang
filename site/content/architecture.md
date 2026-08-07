---
title: "Architecture"
description: "How the Dippin toolchain is organized: IR-centric design, package dependencies, and the parser-to-execution pipeline."
section_label: "Internals"
subtitle: "How the dippin toolchain is organized — packages, data flow, and design decisions."
---

## Compiler Pipeline

Dippin is a multi-stage compiler pipeline. All downstream consumers program against the canonical IR — a set of Go structs defined in the `ir` package.

<div class="flow-diagram">
  <div class="pipeline-box lavender">Source File<br>(.dip or .dot)</div>
  <div class="flow-arrow">&rarr;</div>
  <div class="pipeline-box green">Dippin Parser<br><code>parser</code></div>
  <div class="flow-arrow">&rarr;</div>
  <div class="pipeline-box yellow">Canonical IR<br><code>ir</code></div>
  <div class="flow-arrow">&rarr;</div>
  <div class="pipeline-box lavender">Outputs</div>
</div>

From the canonical IR, multiple output stages fan out:

<div class="flow-diagram">
  <div class="pipeline-box yellow">IR</div>
  <div class="flow-arrow">&rarr;</div>
  <div class="pipeline-box lavender">Validator / Linter</div>
</div>
<div class="flow-diagram">
  <div class="pipeline-box yellow">IR</div>
  <div class="flow-arrow">&rarr;</div>
  <div class="pipeline-box green">Formatter</div>
</div>
<div class="flow-diagram">
  <div class="pipeline-box yellow">IR</div>
  <div class="flow-arrow">&rarr;</div>
  <div class="pipeline-box lavender">DOT Exporter</div>
</div>
<div class="flow-diagram">
  <div class="pipeline-box yellow">IR</div>
  <div class="flow-arrow">&rarr;</div>
  <div class="pipeline-box green">Simulator</div>
</div>
<div class="flow-diagram">
  <div class="pipeline-box yellow">IR</div>
  <div class="flow-arrow">&rarr;</div>
  <div class="pipeline-box yellow">Cost / Coverage / Doctor</div>
</div>

## Package Map

```
dippin-lang/
├── ir/                 # Canonical intermediate representation (types only)
│   ├── ir.go           # Workflow, Node, NodeConfig, RetryConfig, NodeIO
│   ├── edge.go         # Edge, Condition, ConditionExpr
│   ├── source.go       # SourceLocation, SourceMap
│   └── lookup.go       # Helper methods (Node, EdgesFrom, EdgesTo, NodeIDs)
│
├── parser/             # Lexer + recursive descent parser
│   ├── lexer.go        # Indentation-aware tokenizer
│   ├── parser.go       # Produces ir.Workflow from tokens
│   ├── parse_defaults.go  # Defaults block parsing
│   ├── parse_edges.go     # Edge and condition parsing
│   ├── parse_nodes.go     # Node declaration parsing
│   ├── parse_stylesheet.go # Stylesheet section parsing
│   └── parse_helpers.go   # Shared utilities
│
├── validator/          # Graph validation + semantic linting
│   ├── validate.go     # 10 structural checks (DIP001-DIP010)
│   ├── lint.go         # Lint orchestration (DIP101-DIP157)
│   └── diagnostic.go   # Diagnostic type, Result, Severity
│
├── formatter/          # Canonical .dip source formatter
├── export/             # DOT graph exporter
├── migrate/            # DOT ↔ Dippin migration + parity checker
├── simulate/           # Reference graph executor (JSONL events)
├── cost/               # Cost estimation engine + pricing tables
├── coverage/           # Edge coverage analysis
├── doctor/             # Health report card (grade A-F)
├── optimize/           # Model optimization suggestions
├── diff/               # Semantic workflow comparison
├── feedback/           # Cost calibration from telemetry
├── unused/             # Dead-branch detection
├── graph/              # ASCII DAG rendering
├── testrunner/         # Scenario test runner
├── lsp/                # Language Server Protocol server
├── dipx/               # .dipx bundle format (loader tier)
├── scaffold/           # Template scaffolding for dippin new
└── cmd/dippin/         # CLI entry point + command handlers
```

## Key Design Decisions

<div class="decision-grid">
  <div class="decision-card">
    <h4>IR as the universal interface</h4>
    <p>Everything flows through <code>ir.Workflow</code>. Packages import <code>ir</code> but not each other (except analysis packages that compose). This decouples parsing from all downstream consumers.</p>
  </div>
  <div class="decision-card">
    <h4>Sealed interfaces</h4>
    <p>Both <code>NodeConfig</code> and <code>ConditionExpr</code> use Go's sealed interface pattern with unexported methods. Only types within the <code>ir</code> package can implement them, preventing invalid configurations.</p>
  </div>
  <div class="decision-card">
    <h4>No short-circuiting validation</h4>
    <p>All 10 structural checks and all 57 semantic checks run unconditionally. A single call reports every issue, not just the first. This enables batch fixing.</p>
  </div>
  <div class="decision-card">
    <h4>Zero external dependencies</h4>
    <p>Core packages depend only on the Go standard library. The <code>ir</code> package imports only <code>time</code>. Only the <code>lsp</code> package uses external libraries (go.lsp.dev for JSON-RPC 2.0).</p>
  </div>
  <div class="decision-card">
    <h4>Idempotent formatter</h4>
    <p>The formatter is the authority on canonical form. <code>Format(Parse(Format(Parse(source))))</code> always equals <code>Format(Parse(source))</code>. The <code>fmt --check</code> command uses this property.</p>
  </div>
  <div class="decision-card">
    <h4>Testable CLI</h4>
    <p>The <code>Run</code> function accepts <code>args []string</code> and <code>io.Writer</code>, making it fully testable without touching <code>os.Args</code> or <code>os.Stdout</code>. All commands are exercised via this interface.</p>
  </div>
  <div class="decision-card">
    <h4>Loader-tier exemption</h4>
    <p>The <code>dipx</code> package is a bounded exception to the IR-only rule: it may import <code>parser</code> and <code>simulate</code> to materialize a parsed workflow tree from a <code>.dipx</code> bundle, but MUST NOT import <code>validator</code>, <code>cost</code>, <code>formatter</code>, or any analysis package. Pack-time structural validation runs at the CLI layer.</p>
  </div>
</div>

## Dependency Graph

The `ir` package is a leaf dependency. Most packages import only `ir`. Analysis packages compose each other.

### Layer 0: Foundation

<div class="flow-diagram">
  <div class="pipeline-box yellow">ir</div>
  <div class="pipeline-box yellow">event</div>
</div>

### Layer 1: Direct IR Consumers

<div class="flow-diagram">
  <div class="pipeline-box lavender">parser</div>
  <div class="pipeline-box lavender">formatter</div>
  <div class="pipeline-box lavender">validator</div>
  <div class="pipeline-box lavender">export</div>
  <div class="pipeline-box lavender">migrate</div>
  <div class="pipeline-box lavender">scaffold</div>
</div>
<div class="flow-diagram">
  <div class="pipeline-box green">simulate</div>
  <div class="pipeline-box green">cost</div>
  <div class="pipeline-box green">coverage</div>
  <div class="pipeline-box green">graph</div>
  <div class="pipeline-box green">diff</div>
</div>

### Layer 2: Composition

<div class="flow-diagram">
  <div class="pipeline-box yellow">doctor</div>
  <div class="flow-arrow">&larr;</div>
  <div class="pipeline-box green">validator + coverage + cost</div>
</div>
<div class="flow-diagram">
  <div class="pipeline-box yellow">optimize</div>
  <div class="flow-arrow">&larr;</div>
  <div class="pipeline-box green">cost</div>
</div>
<div class="flow-diagram">
  <div class="pipeline-box yellow">unused</div>
  <div class="flow-arrow">&larr;</div>
  <div class="pipeline-box green">coverage + cost</div>
</div>
<div class="flow-diagram">
  <div class="pipeline-box yellow">testrunner</div>
  <div class="flow-arrow">&larr;</div>
  <div class="pipeline-box green">simulate</div>
</div>
<div class="flow-diagram">
  <div class="pipeline-box yellow">lsp</div>
  <div class="flow-arrow">&larr;</div>
  <div class="pipeline-box green">parser + validator</div>
</div>

### Layer 3: CLI

<div class="flow-diagram">
  <div class="pipeline-box lavender">cmd/dippin</div>
  <div class="flow-arrow">&larr;</div>
  <div class="pipeline-box green">all packages</div>
</div>
