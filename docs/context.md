# Context and Variables Reference

Context is the shared state that flows between nodes during workflow execution. It's how nodes communicate — one node writes data, another reads it.

---

## How Context Works

At runtime, context is a `map[string]string` — a flat key-value store threaded through all nodes. When a node executes, it:

1. **Reads** values from context (e.g., the previous node's response)
2. **Processes** something (LLM call, human input, shell command)
3. **Writes** new values back to context (e.g., its output)

The next node in the pipeline sees the updated context.

```mermaid
graph LR
    subgraph Context
        KV["map[string]string"]
    end
    A["Node A"] -- "writes: outcome, response" --> KV
    KV -- "reads: outcome, response" --> B["Node B"]
    B -- "writes: analysis" --> KV
    KV -- "reads: analysis" --> C["Node C"]
```

---

## Variable Namespaces

In `.dip` source files, variables use explicit namespaces for clarity and validation. There are four namespaces:

### ctx — Runtime Context

The primary namespace. Contains handler outputs and reserved keys.

| Variable | Set By | Description |
|----------|--------|-------------|
| `ctx.outcome` | Engine / auto_status | Last node's execution status: `"success"`, `"fail"`, or `"retry"` |
| `ctx.status` | Engine | Alias/complement to `ctx.outcome`; reserved by the validator so references are never flagged undefined |
| `ctx.last_response` | Agent nodes | The LLM's most recent response text |
| `ctx.human_response` | Human nodes | The human's input text |
| `ctx.tool_stdout` | Tool nodes | Standard output from the shell command |
| `ctx.tool_stderr` | Tool nodes | Standard error from the shell command |
| `ctx.tool_marker` | Tool nodes | Tool stdout regex match (when `marker_grep` is declared on the source tool node) |
| `ctx.tool_route` | Tool nodes | A routing value the runtime extracts from the tool's stdout — populated when the tool emits a routing sentinel the runtime recognizes (format defined by the runtime); `route_required: true` additionally fails the node if none is emitted |

These reserved keys are **always available** — the validator knows about them at parse time and can flag typos.

The `ctx.internal.*` prefix is reserved for engine-internal use. Keys under it are always treated as defined by the validator and should not be authored in workflows.

Custom context keys written by nodes (via the engine's `ContextUpdates`) also live in the `ctx` namespace.

### graph — Workflow Attributes

Read-only attributes from the workflow definition:

| Variable | Source | Description |
|----------|--------|-------------|
| `graph.goal` | Workflow header | The workflow's goal string |
| `graph.name` | Workflow header | The workflow's name |
| `graph.start` | Workflow header | The start node ID |
| `graph.exit` | Workflow header | The exit node ID |

Graph attributes are auto-injected into context with the `graph.` prefix.

### params — Subgraph Parameters

Available inside subgraph workflows. These are the parameters passed from the parent workflow:

```dippin
# Parent workflow:
  subgraph SecurityScan
    ref: security/scan_pipeline
    params:
      severity: critical
      model: gpt-5.4

# Inside security/scan_pipeline.dip:
  agent Scanner
    model: ${params.model}
    prompt:
      Scan for ${params.severity} vulnerabilities.
```

Parameters are substituted at expansion time — they don't persist in runtime context.

### inputs — Workflow Inputs

Values supplied by the caller — a human at the entry point, or a parent workflow via a `subgraph` node's `params:` — and bound once at run start:

```dippin
  inputs
    idea: text
      required: true
      prompt: "What do you want built?"

  agent Plan
    prompt:
      Request: ${inputs.idea}
```

Unlike `ctx`, the `inputs` namespace is **closed**: it contains exactly the names declared in the workflow's `inputs` block, and a reference to an undeclared input is a lint error (**DIP156**), not silently treated as defined.

Input values are **untrusted by construction** — they come from outside the workflow author's control, so a host should frame them as data to reason about rather than instructions to follow. For the same reason, `${inputs.x}` is never interpolated inside a tool node's `command:` (**DIP157**): a shell command built from caller-supplied text is an injection vector, so `inputs` references belong in `agent`/`human` prompts, not in a shell command line.

---

## Using Variables in Prompts

Reference context variables in prompts using `${namespace.key}` syntax:

```dippin
  agent Summarize
    prompt:
      The user asked: ${ctx.human_response}

      Previous analysis: ${ctx.last_response}

      Our goal is: ${graph.goal}
```

---

## Using Variables in Conditions

Edge conditions reference the same namespaced variables:

```dippin
  edges
    Check -> Pass when ctx.outcome = success
    Check -> Fail when ctx.outcome = fail
    Route -> A    when ctx.tool_stdout contains "ready"
    Route -> B    when graph.goal contains "review"
```

---

## I/O Declarations (reads/writes)

Nodes can declare which context keys they expect and produce. These are **advisory** — they're used by the linter, not enforced at runtime.

```dippin
  agent Interpret
    reads: human_response
    writes: plan, summary
    prompt:
      Based on the user input, create a plan.
```

**Important**: Use bare key names in `reads`/`writes`, not namespaced:
- Correct: `reads: human_response`
- Incorrect: `reads: ctx.human_response`

**`writes` vs `writable_paths`**: `writes` is an advisory declaration of which `ctx` keys a node produces — used by the linter (DIP107/DIP112) but not enforced at runtime. It is entirely distinct from `writable_paths`, which is a filesystem write-location jail applied to agent nodes at the engine level (see [nodes.md](nodes.md) for `writable_paths`). Do not conflate them.

The linter uses these declarations for:
- **DIP107**: Detecting writes that no downstream node reads
- **DIP112**: Detecting reads that no upstream node writes
- **DIP106**: Detecting undefined variable references in prompts

---

## Namespace Lowering

At the IR-to-engine boundary, namespaces are translated to flat keys:

| Dippin syntax | Engine context key |
|---------------|-------------------|
| `ctx.outcome` | `outcome` |
| `ctx.last_response` | `last_response` |
| `ctx.custom_key` | `custom_key` |
| `graph.goal` | `graph.goal` (already prefixed) |
| `params.model` | Substituted at expansion time |

This is transparent to workflow authors — you always use namespaced syntax in `.dip` files.

---

## Context Preservation Across Restarts

```mermaid
graph TD
    subgraph "Restart Cycle"
        Implement --> Review
        Review -- "ctx.outcome = fail (restart)" --> Implement
        Review -- "ctx.outcome = success" --> Ship
    end
    Context["Context (preserved across restarts)"] -.-> Implement
    Context -.-> Review
```

When a restart edge is followed, context is **fully preserved**. All key-values survive across restarts. This is intentional — it enables iterative refinement patterns:

```dippin
  # First iteration: Implement writes code, Review writes feedback
  # Restart: Implement sees the feedback, writes better code
  edges
    Implement -> Review
    Review -> Implement when ctx.outcome = fail restart: true
    Review -> Ship      when ctx.outcome = success
```

What **is** cleared on restart:
- Completed node status (downstream nodes re-execute)
- Retry counts for cleared nodes (fresh budgets)
- Node-local `SessionStats` (fresh stats per re-execution)

What **survives**:
- All context key-values
- The global restart counter (increments toward `max_restarts`)

---

## Validation Tiers

The validator treats variables differently based on what it can verify at parse time:

| Tier | Variables | Validation |
|------|-----------|------------|
| **Always known** | `ctx.outcome`, `ctx.status`, `ctx.tool_stdout`, `ctx.tool_stderr`, `ctx.tool_marker`, `ctx.tool_route`, `ctx.internal.*`, `graph.*`, `params.*` | Never flagged undefined — the validator recognizes these at parse time (`ctx.tool_marker` is populated when `marker_grep` is declared; `ctx.tool_route` is populated when the tool emits a routing sentinel the runtime recognizes — `route_required: true` additionally fails the node if none is emitted) |
| **Declared outputs** | Keys from upstream `writes` declarations, including well-known-by-convention keys such as `ctx.last_response` (agent nodes) and `ctx.human_response` (human nodes) | Warning if referenced but not declared by any upstream node (DIP112) — a misspelling yields a warning, not an error |
| **Dynamic** | Everything else | Warning only — never an error (runtime context is open) |

This tiered approach catches typos in common variables while allowing flexibility for custom context keys.
