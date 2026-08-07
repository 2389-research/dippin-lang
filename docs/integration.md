# Integration Guide

How to consume Dippin workflows from external projects — as a Go library, via CLI tooling, or both.

---

## Overview

Dippin provides two integration paths:

```mermaid
graph LR
    subgraph "Go Consumers"
        LIB["Import as Go library<br><code>github.com/2389-research/dippin-lang</code>"]
        LIB --> PARSE["parser.NewParser().Parse()"]
        PARSE --> IR["*ir.Workflow"]
        IR --> ADAPT["Your adapter<br>ir.Workflow → your types"]
    end

    subgraph "Non-Go Consumers"
        CLI["dippin parse file.dip"]
        CLI --> JSON["JSON on stdout"]
        JSON --> YOUR["Your parser<br>JSON → your types"]
    end
```

| Path | When to use | Pros | Cons |
|------|-------------|------|------|
| **Go library** | Your project is in Go | Type-safe, no subprocess overhead, access to all IR types | Adds a Go module dependency |
| **CLI (JSON)** | Any language, CI pipelines, scripting | Language-agnostic, no build dependency | Subprocess overhead, JSON parsing required |

For Go projects, **use both**: the library for production integration, the CLI for development tooling (formatting, linting, DOT export).

---

## Go Library Integration

### Installation

```bash
go get github.com/2389-research/dippin-lang
```

### Package Overview

| Package | Import | Purpose |
|---------|--------|---------|
| `ir` | `github.com/2389-research/dippin-lang/ir` | All IR types (Workflow, Node, Edge, conditions, configs) |
| `parser` | `github.com/2389-research/dippin-lang/parser` | Parse `.dip` source → `*ir.Workflow` |
| `validator` | `github.com/2389-research/dippin-lang/validator` | Structural validation + semantic linting |
| `formatter` | `github.com/2389-research/dippin-lang/formatter` | IR → canonical `.dip` source text |
| `export` | `github.com/2389-research/dippin-lang/export` | IR → Graphviz DOT format |
| `migrate` | `github.com/2389-research/dippin-lang/migrate` | DOT → IR conversion (for migration) |

### Parsing a .dip File

```go
import (
    "os"
    "github.com/2389-research/dippin-lang/parser"
)

data, err := os.ReadFile("pipeline.dip")
if err != nil {
    return err
}

p := parser.NewParser(string(data), "pipeline.dip")
workflow, err := p.Parse()
if err != nil {
    return err // parse error with source location
}

// workflow is *ir.Workflow — the canonical IR
fmt.Println(workflow.Name)   // "my_pipeline"
fmt.Println(workflow.Start)  // "AskUser"
fmt.Println(workflow.Exit)   // "Done"
```

**Unresolved-IR view and `ResolveFileDirectives`.** `parser.NewParser().Parse()` is intentionally pure — it does **not** read files referenced by `command_file:`, `prompt_file:`, `system_prompt_file:`, `prompt_include:`, or the defaults `prompt_prefix_file:`/`prompt_suffix_file:` cascade. Nodes that use these directives arrive with the `*File` field set (e.g., `AgentConfig.PromptFile`) and the content field (`Prompt`, `SystemPrompt`, `Command`) empty. This is the correct *unresolved view* for LSP, WASM, and most consumers.

A Go consumer that needs inlined content (e.g., a runtime that will execute the node) must call `parser.ResolveFileDirectives` after `Parse()`:

```go
import "path/filepath"

if err := parser.ResolveFileDirectives(workflow, filepath.Dir(path)); err != nil {
    return err // names only the user-written path, never the resolved absolute path
}
```

The CLI commands `lint`, `validate`, and `pack` call this automatically. Direct `Parse()` callers — including adapters, LSP servers, and WASM hosts — do not need to call it unless they require the inlined content.

**`*_file` security constraints.** `ResolveFileDirectives` enforces: relative paths only (absolute paths are rejected), no `..`-based or symlink-based parent escape, leaf symlinks rejected atomically on Unix via `O_NOFOLLOW`, and a 4 MiB per-file cap. Error messages name only the user-written path, never the resolved absolute path.

**DOT round-trip lossiness.** `export.ExportDOT` inlines the resolved command text as `tool_command=` in DOT output (`applyToolPromptAttrs` in `export/dot.go`). As a result, export-to-DOT then `dippin migrate` back yields `command:` (inline form), not `command_file:`. The `*_file` directive is not preserved through a DOT round-trip.

### Validating a Workflow

```go
import "github.com/2389-research/dippin-lang/validator"

// Structural checks (DIP001–DIP010) — must pass
result := validator.Validate(workflow)
if result.HasErrors() {
    for _, d := range result.Diagnostics {
        fmt.Println(d.String())
        // error[DIP003]: unknown node reference "InterpretX" in edge
        //   --> pipeline.dip:45:5
        //   = help: did you mean "Interpret"?
    }
}

// Semantic lint (DIP101–DIP157; DIP146 is CLI/cross-file only) — warnings
lintResult := validator.Lint(workflow)
for _, d := range lintResult.Diagnostics {
    fmt.Println(d.String())
}
```

### Formatting and Exporting

```go
import (
    "github.com/2389-research/dippin-lang/formatter"
    "github.com/2389-research/dippin-lang/export"
)

// Round-trip: IR → canonical .dip source
canonical := formatter.Format(workflow)

// IR → Graphviz DOT
dot := export.ExportDOT(workflow, export.ExportOptions{
    RankDir:        "LR",
    IncludePrompts: true,
})
```

---

### Loading `.dipx` Bundles

For projects shipping pipelines as content-addressed artifacts, the `dipx` package replaces ad-hoc directory walking with a single integrity-verified entry point:

```go
import "github.com/2389-research/dippin-lang/dipx"

src, err := dipx.Load(ctx, "pipeline.dipx")  // also accepts pipeline.dip
if err != nil { return err }

entry := src.Entry()                                          // *ir.Workflow
child, err := src.Workflow("phases/review.dip", entry.SourceLocation.File)
```

`Source` is the interface runtimes program against. Both `.dip`-on-disk and `.dipx`-bundle sources satisfy it; the `Workflow(refPath, relativeTo string)` argument order matches `flatten.Resolver.Resolve`. `dipx.Open` returns a `*Bundle` that additionally exposes:

- `Manifest()` — the parsed manifest (format version, entry path, files[])
- `Identity()` — 32-byte SHA-256 over the manifest bytes-as-stored (a stable content-addressed id)

`dipx.Pack(ctx, entry, w, opts)` writes a deterministic ZIP from a `.dip` entry; `dipx.Extract(ctx, src, dest, allowOverwrite)` atomically unpacks one. Both refuse symlinks anywhere in the source tree (Pack) or staging tree (Extract) and enforce the spec's per-file (50 MB) and total-bundle (100 MB) caps.

`PackOptions{}` (the zero value) is the default: every `command_file:` / `prompt_file:` / `system_prompt_file:` body — plus `prompt_include:` and the defaults `prompt_prefix_file:`/`prompt_suffix_file:` cascade, composed into the agent prompt — is inlined into the bundled `.dip`, producing a self-contained `format_version 1` bundle. `PackOptions{NoInline: true, Include: []string{...}}` instead ships those directive targets — plus any extra sibling files or directories listed in `Include` (relative to the entry's directory) — as separate entries under `workflows/`, keeps the `*_file:` directives, and produces `format_version 2`. `Include` requires `NoInline`.

**Note:** `dipx` does not run `validate`/`lint`/`cost` on the bundled workflow — call `validator.Validate(src.Entry())` from your code if needed. `dippin pack` does this at the CLI layer because `dipx` is bound by the loader-tier rule and cannot import `validator`.

---

## The IR Type System

The `ir.Workflow` is the central type all integration code works with. Understanding it is critical for writing adapters.

### Workflow Structure

```mermaid
graph TD
    W["Workflow"] --> Name["Name, Goal, Start, Exit"]
    W --> Defaults["WorkflowDefaults<br>model, provider, retry_policy, ..."]
    W --> Nodes["[]*Node"]
    W --> Edges["[]*Edge"]
    Nodes --> N["Node"]
    N --> Kind["Kind: agent | human | tool | parallel | fan_in | subgraph"]
    N --> Config["Config (NodeConfig interface)"]
    N --> Retry["RetryConfig"]
    N --> IO["NodeIO (reads/writes)"]
    Config --> AC["AgentConfig"]
    Config --> HC["HumanConfig"]
    Config --> TC["ToolConfig"]
    Config --> PC["ParallelConfig"]
    Config --> FC["FanInConfig"]
    Config --> SC["SubgraphConfig"]
    Edges --> E["Edge"]
    E --> Cond["Condition (optional)"]
    Cond --> AST["CondCompare | CondAnd | CondOr | CondNot"]
```

### Node Kind → Config Type Mapping

Each node kind has exactly one config type. Use a type switch to access kind-specific fields:

```go
switch cfg := node.Config.(type) {
case ir.AgentConfig:
    fmt.Println(cfg.Prompt)
    fmt.Println(cfg.Model)
    fmt.Println(cfg.GoalGate)
case ir.HumanConfig:
    fmt.Println(cfg.Mode)         // "choice", "freeform", "interview", or "yes_no"
    fmt.Println(cfg.Default)
    fmt.Println(cfg.QuestionsKey) // interview mode: context key for questions
    fmt.Println(cfg.AnswersKey)   // interview mode: context key for answers
case ir.ToolConfig:
    fmt.Println(cfg.Command)
    fmt.Println(cfg.Timeout)
case ir.ParallelConfig:
    fmt.Println(cfg.Targets) // []string of node IDs
case ir.FanInConfig:
    fmt.Println(cfg.Sources) // []string of node IDs
case ir.SubgraphConfig:
    fmt.Println(cfg.Ref)
    fmt.Println(cfg.Params)
}
```

### Lookup Helpers

The `ir.Workflow` type provides convenience methods for graph traversal:

```go
node := workflow.Node("Analyze")         // find by ID
outgoing := workflow.EdgesFrom("Analyze") // all edges leaving this node
incoming := workflow.EdgesTo("Analyze")   // all edges entering this node
ids := workflow.NodeIDs()                 // all IDs in declaration order
```

### Interview Mode (Runtime Integration)

When `cfg.Mode == "interview"`, the runtime is expected to:

1. **Read questions** from `ctx[cfg.QuestionsKey]` (default: `"interview_questions"`).
2. **Parse questions** from the upstream agent's markdown output — numbered/bulleted lines ending in `?`, imperative prompts ("Describe...", "List..."). Skip content inside fenced code blocks.
3. **Extract inline options** from trailing parentheticals: `"Auth model? (API key, OAuth, JWT)"` becomes a Select field with 3 options + "Other" freeform.
4. **Present each question** as an individual form field. Recommended mapping:
   - Questions with options → Select + "Other" text input
   - Yes/no questions → Confirm + optional elaboration
   - All other questions → TextArea
5. **Store answers** in `ctx[cfg.AnswersKey]` (default: `"interview_answers"`) as JSON and in `ctx["human_response"]` as markdown.
6. **Handle edge cases:**
   - 0 questions: show confirmation + optional freeform note
   - Malformed output: fall back to single TextArea with `cfg.Prompt` as instructions
   - Ctrl-C: write partial answers with `canceled: true` in JSON
   - Retry loop: pre-fill with previous answers from `ctx[cfg.AnswersKey]`

Recommended answer JSON schema:

```json
{
  "questions": [
    {"id": "q1", "text": "Auth model?", "options": ["API key", "OAuth", "JWT"], "answer": "OAuth", "elaboration": "Google + GitHub"},
    {"id": "q2", "text": "Describe integrations", "answer": "Salesforce nightly sync..."}
  ],
  "incomplete": false,
  "canceled": false
}
```

### Reusable Interview Loop

The `examples/interview_loop.dip` workflow encapsulates the full interview pattern (ask → answer → assess → loop) as a standalone workflow that can be embedded via `subgraph`:

```dippin
subgraph Requirements
  ref: interview_loop.dip
  params:
    topic: "API design"
    focus: "resources, auth, scale, integrations"
```

The loop handles question generation with inline options, structured answer collection via interview mode, and iterative refinement until the assessor is satisfied. It writes `requirements_summary` to context when complete.

---

## Writing an Adapter

The recommended pattern: your project imports dippin-lang and converts `*ir.Workflow` into your own types. This keeps dippin-lang generic and your project in control of the mapping.

### Example Consumer Types

A typical runtime might use `pipeline.Graph`, `pipeline.Node`, and `pipeline.Edge` with a shape-to-handler mapping:

```go
// Example consumer types (simplified)
type Graph struct {
    Name, StartNode, ExitNode string
    Nodes map[string]*Node
    Edges []*Edge
    Attrs map[string]string
}

type Node struct {
    ID, Shape, Label, Handler string
    Attrs map[string]string
}

type Edge struct {
    From, To, Label, Condition string
    Attrs map[string]string
}
```

### Adapter: ir.Workflow → pipeline.Graph

```go
package pipeline

import (
    "fmt"
    "github.com/2389-research/dippin-lang/ir"
)

// kindToShape maps Dippin node kinds back to DOT shapes
// that the runtime's handler registry understands.
var kindToShape = map[ir.NodeKind]string{
    ir.NodeAgent:    "box",
    ir.NodeHuman:    "hexagon",
    ir.NodeTool:     "parallelogram",
    ir.NodeParallel: "component",
    ir.NodeFanIn:    "tripleoctagon",
    ir.NodeSubgraph: "tab",
}

// FromDippinIR converts a Dippin IR workflow into a runtime Graph.
// The returned Graph is ready for use with NewEngine and the handler registry.
func FromDippinIR(w *ir.Workflow) *Graph {
    g := NewGraph(w.Name)

    // Map workflow-level defaults to graph attributes.
    if w.Goal != "" {
        g.Attrs["goal"] = w.Goal
    }
    if w.Defaults.Model != "" {
        g.Attrs["llm_model"] = w.Defaults.Model
    }
    if w.Defaults.Provider != "" {
        g.Attrs["llm_provider"] = w.Defaults.Provider
    }
    if w.Defaults.RetryPolicy != "" {
        g.Attrs["default_retry_policy"] = w.Defaults.RetryPolicy
    }
    if w.Defaults.MaxRetries > 0 {
        g.Attrs["default_max_retry"] = fmt.Sprintf("%d", w.Defaults.MaxRetries)
    }
    if w.Defaults.MaxRestarts > 0 {
        g.Attrs["max_restarts"] = fmt.Sprintf("%d", w.Defaults.MaxRestarts)
    }

    // Convert nodes.
    for _, n := range w.Nodes {
        node := &Node{
            ID:    n.ID,
            Label: n.Label,
            Shape: kindToShape[n.Kind],
            Attrs: make(map[string]string),
        }

        // Start/exit get special shapes.
        if n.ID == w.Start {
            node.Shape = "Mdiamond"
        }
        if n.ID == w.Exit {
            node.Shape = "Msquare"
        }

        // Extract kind-specific attributes.
        switch cfg := n.Config.(type) {
        case ir.AgentConfig:
            if cfg.Prompt != "" {
                node.Attrs["prompt"] = cfg.Prompt
            }
            if cfg.SystemPrompt != "" {
                node.Attrs["system_prompt"] = cfg.SystemPrompt
            }
            if cfg.Model != "" {
                node.Attrs["llm_model"] = cfg.Model
            }
            if cfg.Provider != "" {
                node.Attrs["llm_provider"] = cfg.Provider
            }
            if cfg.MaxTurns > 0 {
                node.Attrs["max_turns"] = fmt.Sprintf("%d", cfg.MaxTurns)
            }
            if cfg.GoalGate {
                node.Attrs["goal_gate"] = "true"
            }
            if cfg.AutoStatus {
                node.Attrs["auto_status"] = "true"
            }
            if cfg.ReasoningEffort != "" {
                node.Attrs["reasoning_effort"] = cfg.ReasoningEffort
            }
            if cfg.Fidelity != "" {
                node.Attrs["fidelity"] = cfg.Fidelity
            }
            // tool_access: enforce fail-closed at the runtime.
            // "" means no restriction; "none" strips all tools.
            // Any unrecognized value must also fail closed (no tools) —
            // dippin lints DIP139 for unrecognized values but does not enforce.
            // tool_access is node-scoped; per-branch parallel overrides (BranchConfig.ToolAccess)
            // inherit the agent's value when empty and never re-grant tools.
            if cfg.ToolAccess != "" {
                node.Attrs["tool_access"] = cfg.ToolAccess
            }
            // writable_paths: the runtime MUST enforce as a write-jail (fs-level).
            // Fail-closed rules (safety requirement):
            //   - present-but-empty or comma-only is a parse error → deny all writes
            //   - malformed or runtime-unrecognized → deny all writes or refuse to start
            //   - native backend: enforce fs-level write jail (Bash and children included)
            //   - claude-code/acp backends: refuse to start when writable_paths is set
            //   - VERSION SKEW: a runtime that does not enforce writable_paths MUST refuse
            //     to start; consumers must pin an enforcing runtime, never @latest
            // dippin carries and lints (DIP141, DIP142); the runtime enforces.
            if len(cfg.WritablePaths) > 0 {
                node.Attrs["writable_paths"] = strings.Join(cfg.WritablePaths, ",")
            }
        case ir.HumanConfig:
            if cfg.Mode != "" {
                node.Attrs["mode"] = cfg.Mode
            }
            if cfg.Default != "" {
                node.Attrs["default"] = cfg.Default
            }
        case ir.ToolConfig:
            if cfg.Command != "" {
                node.Attrs["tool_command"] = cfg.Command
            }
            if cfg.Timeout > 0 {
                node.Attrs["timeout"] = cfg.Timeout.String()
            }
        case ir.ParallelConfig:
            // Targets are expressed via edges, not attributes.
        case ir.FanInConfig:
            // Sources are expressed via edges, not attributes.
        case ir.SubgraphConfig:
            if cfg.Ref != "" {
                node.Attrs["subgraph_ref"] = cfg.Ref
            }
        }

        // Retry config.
        if n.Retry.Policy != "" {
            node.Attrs["retry_policy"] = n.Retry.Policy
        }
        if n.Retry.MaxRetries > 0 {
            node.Attrs["max_retries"] = fmt.Sprintf("%d", n.Retry.MaxRetries)
        }

        g.AddNode(node)
    }

    // Convert edges.
    for _, e := range w.Edges {
        edge := &Edge{
            From:  e.From,
            To:    e.To,
            Label: e.Label,
            Attrs: make(map[string]string),
        }

        // Flatten condition AST back to string for the runtime's engine.
        if e.Condition != nil {
            edge.Condition = e.Condition.Raw
        }

        if e.Restart {
            edge.Attrs["restart"] = "true"
        }
        if e.Weight > 0 {
            edge.Attrs["weight"] = fmt.Sprintf("%d", e.Weight)
        }

        g.AddEdge(edge)
    }

    return g
}
```

### Using the Adapter

Replace the file-loading code to auto-detect format:

```go
import (
    "path/filepath"
    "strings"

    "github.com/2389-research/dippin-lang/parser"
    "github.com/2389-research/dippin-lang/validator"
)

func LoadPipeline(path string) (*Graph, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, err
    }

    ext := strings.ToLower(filepath.Ext(path))
    switch ext {
    case ".dip":
        p := parser.NewParser(string(data), path)
        w, err := p.Parse()
        if err != nil {
            return nil, fmt.Errorf("dippin parse error: %w", err)
        }

        // Validate before converting.
        result := validator.Validate(w)
        if result.HasErrors() {
            return nil, fmt.Errorf("validation failed: %s", result.Diagnostics[0].String())
        }

        return FromDippinIR(w), nil

    case ".dot":
        return ParseDOT(string(data))

    default:
        return nil, fmt.Errorf("unsupported file extension: %s", ext)
    }
}
```

This is a **zero-disruption migration**: existing `.dot` files keep working, `.dip` files are supported alongside, and the engine doesn't know the difference.

---

## CLI Integration (Non-Go)

For projects not written in Go, the `dippin` CLI provides JSON output suitable for any language.

### Parse to JSON

```bash
dippin parse pipeline.dip
```

Outputs the full IR as JSON to stdout:

```json
{
  "Name": "my_pipeline",
  "Goal": "Ask user for a task",
  "Start": "AskUser",
  "Exit": "Done",
  "Defaults": {
    "Model": "claude-opus-4-6",
    "Provider": "anthropic"
  },
  "Nodes": [
    {
      "ID": "AskUser",
      "Kind": "human",
      "Label": "What would you like to build?",
      "Config": {
        "Mode": "freeform"
      }
    }
  ],
  "Edges": [
    {
      "From": "AskUser",
      "To": "Interpret",
      "Condition": null
    }
  ]
}
```

### Validate Before Use

```bash
# Structural validation (exit code 0 = pass, 1 = fail)
dippin validate pipeline.dip

# Full lint with JSON diagnostics for machine parsing
dippin --format json lint pipeline.dip
```

### Format and Export

```bash
# Canonical formatting (for CI checks)
dippin fmt --check pipeline.dip

# Export to DOT for visualization
dippin export-dot pipeline.dip | dot -Tpng -o pipeline.png
```

### Example: Python Consumer

```python
import json
import subprocess

def load_dippin(path: str) -> dict:
    result = subprocess.run(
        ["dippin", "parse", path],
        capture_output=True, text=True
    )
    if result.returncode != 0:
        raise ValueError(f"Parse failed: {result.stderr}")
    return json.loads(result.stdout)

workflow = load_dippin("pipeline.dip")
for node in workflow["Nodes"]:
    print(f"{node['Kind']}: {node['ID']}")
```

### Example: TypeScript/Node Consumer

```typescript
import { execFileSync } from "child_process";

function loadDippin(path: string): Workflow {
  const stdout = execFileSync("dippin", ["parse", path], {
    encoding: "utf-8",
  });
  return JSON.parse(stdout);
}
```

---

## Migration Path

For projects currently consuming DOT files, Dippin provides a smooth migration:

```mermaid
graph TD
    A["1. Install dippin CLI"] --> B["2. Migrate existing .dot files"]
    B --> C["3. Verify parity"]
    C --> D["4. Add .dip loading to your project"]
    D --> E["5. Run both formats in parallel"]
    E --> F["6. Switch to .dip as default"]
```

### Step 1: Install the CLI

```bash
go install github.com/2389-research/dippin-lang/cmd/dippin@latest
```

### Step 2: Migrate existing DOT files

```bash
dippin migrate --output pipeline.dip pipeline.dot
```

### Step 3: Verify structural parity

```bash
dippin validate-migration pipeline.dot pipeline.dip
# Output: "migration parity check passed"
```

### Step 4: Add .dip loading

Add the adapter code (see above) or CLI integration to your project. Support both `.dot` and `.dip` during transition.

### Step 5: Run both formats

Keep both files during the transition period. Author new pipelines in `.dip`. Validate that existing workflows behave identically.

### Step 6: Switch to .dip as default

Once confident, make `.dip` the default format. Keep DOT as an export target for visualization:

```bash
dippin export-dot pipeline.dip | dot -Tpng -o pipeline.png
```

---

## CI Integration

### Format Check

Enforce canonical formatting in CI:

```yaml
- name: Check Dippin formatting
  run: dippin fmt --check pipeline.dip
```

### Lint Check

Catch structural errors and warnings:

```yaml
- name: Lint Dippin workflows
  run: dippin lint pipeline.dip
```

### JSON Diagnostics for Annotations

Use JSON output for GitHub Actions annotations or similar:

```yaml
- name: Lint with JSON output
  run: dippin --format json lint pipeline.dip > diagnostics.json
```

---

## Condition String Compatibility

Runtimes that store conditions as raw strings (`edge.Condition`) can use Dippin's `Condition.Raw` field directly — Dippin parses conditions into an AST but always preserves the original raw text. When adapting back to such runtime types, use the raw string:

```go
if e.Condition != nil {
    edge.Condition = e.Condition.Raw
}
```

**Namespace difference**: Dippin uses namespaced variables (`ctx.outcome`), while some runtimes use bare names (`outcome`) or prefixed names (`context.tool_stdout`). During migration, the `migrate` command handles the translation. When writing new `.dip` files, always use the `ctx.` namespace — the adapter should strip the prefix when converting to the runtime's format if needed.

---

## LLM Tool-Calling Integration

When using LLMs to generate `.dip` files, `dippin check` provides a single-call validation endpoint optimized for tool-calling loops.

### Tool Definition

Define `dippin check` as a tool that the LLM can call to validate its generated output:

```json
{
  "name": "validate_workflow",
  "description": "Validate a .dip workflow file. Returns JSON with valid (bool), error/warning counts, diagnostics, and suggested fixes.",
  "parameters": {
    "file_path": {"type": "string", "description": "Path to the .dip file to validate"}
  }
}
```

### Generate → Check → Fix Loop

```python
import json, subprocess, tempfile

def generate_and_validate(llm, prompt, max_retries=3):
    source = llm.generate(prompt)

    for attempt in range(max_retries):
        with tempfile.NamedTemporaryFile(suffix=".dip", mode="w", delete=False) as f:
            f.write(source)
            path = f.name

        result = subprocess.run(
            ["dippin", "check", path],
            capture_output=True, text=True
        )
        check = json.loads(result.stdout)

        if check["valid"]:
            return source

        # Feed diagnostics back to the LLM for correction.
        errors = [d["message"] for d in check["diagnostics"] if d["severity"] == "error"]
        fixes = check["suggested_actions"]
        source = llm.generate(
            f"Fix these errors in the .dip file:\n"
            + "\n".join(errors)
            + ("\nSuggested fixes: " + ", ".join(fixes) if fixes else "")
            + f"\n\nOriginal file:\n{source}"
        )

    raise ValueError("Failed to generate valid workflow after retries")
```

### Scaffold + Modify Pattern

For LLMs that struggle with .dip syntax from scratch, use `dippin new` to generate a valid starting point, then modify:

```bash
# Generate a template the LLM can modify
dippin new conditional --name MyWorkflow --write /tmp/base.dip

# LLM modifies the file...

# Validate the modifications
dippin check /tmp/base.dip
```

### Reference Card

The [`docs/llm-reference.md`](llm-reference.md) file (~2000 tokens) is designed for injection into system prompts. It covers grammar, node kinds, edge syntax, common mistakes, and a complete example.

---

## What Dippin Does NOT Provide

Dippin is a **language and toolchain**, not a runtime. It does not:

- Execute workflows
- Manage pipeline state or checkpoints
- Call LLM APIs
- Handle human interaction UI
- Run shell commands
- Enforce `tool_access` restrictions — `ToolAccess` is a hint that the runtime must enforce; without an enforcing runtime, `tool_access: none` is a no-op
- Enforce `writable_paths` as a write-jail — dippin carries and lints these fields (DIP141, DIP142); the runtime is responsible for the fs-level jail; a runtime that does not enforce `writable_paths` **must refuse to start** (version skew is a safety requirement)

These responsibilities stay in the consuming project (e.g., the runtime's engine, handler registry, and UI). Dippin's job is to parse, validate, format, and export — the consuming project does everything else.
