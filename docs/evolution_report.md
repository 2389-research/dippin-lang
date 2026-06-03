# Dippin Language Evolution Report
## Analysis of Runtime Integration and Future Improvements

**Date**: March 20, 2025  
**Status**: Post-Migration Analysis  
**Context**: Following successful verification of Dippin migration tooling against the runtime's vulnerability_analyzer.dot pipeline

---

## Executive Summary

This report analyzes the integration between Dippin (the declarative pipeline language) and the runtime (the execution engine), identifying what works well and proposing concrete evolution paths. The analysis reveals that **Dippin's core design is sound**, but several runtime features cannot yet be expressed in Dippin IR. The priority recommendations focus on bridging these gaps while preserving the clean separation between language (Dippin) and the runtime.

**Key Finding**: The runtime has evolved sophisticated runtime behaviors (stylesheets, fidelity modes, retry policies, event models, composition) that should inform Dippin v1.5+, but the v1 design correctly deferred these to avoid complexity. The migration tooling proves that topology and basic node configuration can be expressed cleanly.

---

## 1. What Dippin Gets Right

### 1.1 Explicit Start/Exit Nodes
**Runtime Reality**: The engine uses `StartNode` and `ExitNode` fields explicitly. The old DOT convention of inferring start from `Mdiamond` shape was fragile.

**Dippin Design**: `start: node_id` and `exit: node_id` are required top-level fields.

**Verdict**: ✅ **Perfect alignment**. This removes ambiguity and makes validation trivial.

---

### 1.2 Multiline Content Without Escaping
**Runtime Reality**: DOT files use `\n` escaping for prompts and shell commands, which breaks syntax highlighting and readability. The migration tool correctly un-escapes these.

**Dippin Design**: Indentation-based multiline blocks preserve literal content.

**Verdict**: ✅ **Major usability win**. LLMs and humans can author/repair prompts without fighting escape sequences.

---

### 1.3 Typed Node Configurations
**Runtime Reality**: The engine dispatches to different handlers based on `node.Handler` (string), then type-casts `node.Attrs` (map[string]string) to extract typed config.

**Dippin Design**: `NodeConfig` is a sealed union of typed structs (`AgentConfig`, `ToolConfig`, etc.). Each node kind has its own field set.

**Verdict**: ✅ **Structurally superior**. This makes invalid states unrepresentable (e.g., a tool node with a `prompt` field) and enables parse-time schema validation. The runtime's type-checking becomes unnecessary.

---

### 1.4 Condition AST in IR
**Runtime Reality**: Conditions are stored as raw strings (`edge.Condition`) and parsed at evaluation time in `condition.go`. Syntax errors only surface during execution.

**Dippin Design**: Conditions are parsed into an AST (`ConditionExpr`) during Dippin parsing. The IR stores both raw source and parsed AST.

**Verdict**: ✅ **Correctness improvement**. Parse-time validation catches syntax errors before execution. The AST makes optimization and analysis possible (e.g., dead edge detection).

---

### 1.5 Validation Code Registry (DIP001–DIP112)
**Runtime Reality**: Validation errors in `validate.go` return ad-hoc strings. Error recovery is limited.

**Dippin Design**: Every validation rule has a unique code (DIP001 = missing start, DIP002 = missing exit, etc.). Diagnostics include severity, location, message, and help text.

**Verdict**: ✅ **Developer experience win**. Structured diagnostics enable IDE integration, error documentation, and multi-error recovery.

---

### 1.6 IR as Engine Input (Future)
**Runtime Reality**: The engine operates on `pipeline.Graph`, which is DOT-shaped.

**Dippin Design**: The IR (`ir.Workflow`) is the canonical representation. An adapter (`IRToGraph()`) bridges the transition, but the long-term plan is for the engine to consume IR directly.

**Verdict**: ✅ **Correct architecture**. Decoupling language syntax from execution model allows Dippin to evolve independently and supports future alternative frontends (YAML, GUI, visual editor).

---

## 2. Gaps Between Dippin IR and the Runtime

### 2.1 **Missing: Retry Policy Names and Backoff Strategies**

**Runtime Has**:
- Named retry policies (`standard`, `aggressive`, `patient`, `linear`, `none`) defined in `retry_policy.go`
- Each policy specifies `MaxRetries`, `BaseDelay`, and a `BackoffFn` (exponential or linear)
- Resolution hierarchy: node attr `retry_policy` → graph attr `default_retry_policy` → fallback to `standard`
- Per-node `max_retries` override that replaces the policy's default attempt count

**Dippin IR Has**:
```go
type RetryConfig struct {
    Policy         string // Named policy: "standard", "aggressive", "patient", "linear", "none"
    MaxRetries     int    // Override default
    RetryTarget    string // Node to jump to on retry
    FallbackTarget string // Fallback if retries exhausted
}
```

**Gap**: Dippin IR has `Policy` (string) but doesn't define what valid policy names are, what each policy's backoff function is, or how base delays are configured. The IR is silent on the backoff mechanics.

**Proposed Solution**:

**Option A (Recommended)**: Document valid policy names in the Dippin spec and make the engine responsible for policy resolution. Add an optional `base_delay` field to `RetryConfig` for per-node override:

```go
type RetryConfig struct {
    Policy         string        // "standard" | "aggressive" | "patient" | "linear" | "none"
    MaxRetries     int           // Override policy's default max
    BaseDelay      time.Duration // Override policy's default base delay (optional)
    RetryTarget    string
    FallbackTarget string
}
```

**Dippin Syntax**:
```
agent analyze_code:
  prompt: "Analyze this code"
  retry_policy: aggressive
  max_retries: 7
  base_delay: 500ms
  retry_target: validate_analysis
  fallback_target: manual_review
```

**Option B**: Make backoff strategies first-class in IR with their own type:
```go
type BackoffStrategy interface {
    backoffStrategy()
}
type ExponentialBackoff struct {
    Base time.Duration
}
type LinearBackoff struct {
    Base time.Duration
}
```
This adds complexity but gives Dippin full control over retry mechanics. **Not recommended** for v1 — defer to engine.

---

### 2.2 **Missing: Fidelity Modes and Context Compaction**

**Runtime Has** (`fidelity.go`):
- Six fidelity levels controlling context verbosity:
  - `full`: complete context (default first execution)
  - `summary:high`: all keys + trimmed artifacts (2000 chars)
  - `summary:medium`: key decisions only (outcome, last_response, human_response)
  - `summary:low`: one-line per completed node
  - `compact`: only graph.goal + outcome
  - `truncate`: medium keys capped at 500 chars each
- Degradation chain: on checkpoint resume, fidelity degrades one level (e.g., `full` → `summary:high`)
- Resolution: node attr `fidelity` → graph attr `default_fidelity` → `full`
- `CompactContext()` applies fidelity-specific compaction logic

**Dippin IR Has**:
```go
type AgentConfig struct {
    // ...
    Fidelity            string  // Present but undocumented
    Compaction          string  // "auto" | "none" (from validate_semantic.go)
    CompactionThreshold float64 // Not used by engine yet
}
```

**Gap**: Dippin IR has `Fidelity` and `Compaction` strings but doesn't enumerate valid values, specify degradation behavior, or explain how compaction interacts with checkpoint resume.

**Proposed Solution**:

**1. Document valid fidelity levels in the spec**:
```
fidelity: full | summary:high | summary:medium | summary:low | compact | truncate
```

**2. Add `Compaction` to `WorkflowDefaults`** (it's currently per-node only):
```go
type WorkflowDefaults struct {
    // ...
    Fidelity   string // Default fidelity level
    Compaction string // "auto" | "none"
}
```

**3. Clarify in spec**:
- `compaction: auto` → engine applies fidelity-based compaction on resume
- `compaction: none` → context is preserved verbatim across resume
- `compaction_threshold` (per-node) is a future extension for token-budget-based compaction (not implemented yet in engine)

**4. Consider adding degradation policy to IR** (post-v1):
```go
type FidelityConfig struct {
    Level    string // "full", "summary:high", etc.
    OnResume string // "degrade" | "preserve" | "reset"
}
```

---

### 2.3 **Missing: Stylesheet System**

**Runtime Has** (`stylesheet.go`):
- CSS-like syntax for per-node LLM config via selectors:
  ```
  * { model: gpt-4; provider: openai; }
  .coder { model: o1; reasoning_effort: high; }
  #critical_gate { max_retries: 5; }
  ```
- Specificity ordering: universal (`*`) < shape (`agent`) < class (`.coder`) < ID (`#critical_gate`)
- Explicit node attributes override all stylesheet rules
- Stored in graph attribute `model_stylesheet` (string)

**Dippin IR Has**: Nothing. Stylesheets were deferred to post-v1.

**Gap**: Runtime pipelines can use stylesheets to avoid repeating `model: o1` on every agent node. Migrated Dippin files must inline all stylesheet-resolved values.

**Proposed Solution** (v1.5):

Add a top-level `stylesheet` block to Dippin syntax:

```
workflow analyze_codebase:
  start: scan
  exit: report
  
  defaults:
    model: gpt-4o-mini
    provider: openai
  
  stylesheet:
    * { temperature: 0.7 }
    .coder { model: o1; reasoning_effort: medium }
    #critical_gate { max_retries: 5; retry_policy: aggressive }
  
  agent scan:
    class: coder
    prompt: "Scan the codebase"
    # Inherits model: o1, reasoning_effort: medium from .coder
```

**IR Addition**:
```go
type Stylesheet struct {
    Rules []StyleRule
}

type StyleRule struct {
    Selector   string            // "*" | ".class" | "#id" | "kind"
    Properties map[string]string // Resolved to typed config during parsing
}

type Workflow struct {
    // ...
    Stylesheet *Stylesheet // Optional
}
```

**Implementation Notes**:
- Parse stylesheet at Dippin parse time (not a string-in-IR)
- Resolve during formatting/export (expand to explicit fields)
- Validation warning if a selector matches no nodes
- Migration tool extracts `model_stylesheet` and generates inline fields (lossy for now)

---

### 2.4 **Missing: Parallel Branch Configuration**

**Runtime Has**:
- `parallel` handler fans out to multiple child pipelines, each with an isolated context snapshot
- Branch results are JSON-encoded and stored in `parallel.results` context key
- `fan_in` handler waits for all branches to complete

**Dippin IR Has**:
```go
type ParallelConfig struct {
    Targets []string // Fan-out target node IDs
}
```

**Gap**: No way to specify branch-specific parameters or how results are aggregated. The IR assumes targets are node IDs in the same graph, but the runtime's `parallel` handler can execute independent subgraphs.

**Proposed Solution** (v1.5):

Add optional per-branch configuration:
```go
type ParallelConfig struct {
    Branches []ParallelBranch
}

type ParallelBranch struct {
    Target string            // Node ID or subgraph ref
    Params map[string]string // Branch-specific context overrides
}
```

**Dippin Syntax**:
```
parallel analyze:
  branch:
    target: lint
    params:
      tool: eslint
  branch:
    target: test
    params:
      suite: unit
  branch:
    target: security_scan
```

---

### 2.5 **Missing: Subgraph Parameters**

**Runtime Has** (`subgraph.go`):
- `SubgraphHandler` passes parent context to child via `WithInitialContext(pctx.Snapshot())`
- No explicit parameter declaration or namespace isolation

**Dippin IR Has**:
```go
type SubgraphConfig struct {
    Ref    string            // Workflow name or path
    Params map[string]string // Parameter overrides
}
```

**Gap**: Params are in IR but no mechanism to declare what parameters a workflow accepts or how they map to internal context keys.

**Proposed Solution** (v1.5):

Add top-level `params` block to Dippin:
```
workflow security_scan:
  params:
    repo_path: string  # Required
    severity: string = "high"  # Optional with default
  
  start: clone_repo
  exit: report
  
  tool clone_repo:
    command: git clone ${params.repo_path}
```

**IR Addition**:
```go
type WorkflowParam struct {
    Name    string
    Type    string // "string" | "int" | "bool" for v1
    Default string // Empty = required
}

type Workflow struct {
    // ...
    Params []WorkflowParam
}
```

**Validation**:
- DIP-new: Subgraph reference passes undefined param
- DIP-new: Required param not provided
- DIP-new: Param type mismatch (post-v1 when typed params exist)

---

### 2.6 **Missing: Goal Gate Semantics**

**Runtime Has** (`engine.go:goalGateRetryTarget`):
- Nodes with `goal_gate: true` force pipeline failure if they complete with `outcome != success` or `partial_success`
- At exit, engine checks all completed goal gates and triggers retry/fallback if any failed
- Used for critical validation nodes that must succeed for the pipeline to be valid

**Dippin IR Has**:
```go
type AgentConfig struct {
    // ...
    GoalGate bool // Present but semantics not documented
}
```

**Gap**: Dippin IR has the field but spec doesn't explain the exit-time validation behavior.

**Proposed Solution**:

**1. Document in spec**:
> `goal_gate: true` marks a node as critical. If the pipeline reaches the exit node but any completed goal gate has `outcome != success`, the pipeline fails even if the exit node itself succeeded. Goal gates enable declarative invariant checking (e.g., "all tests must pass").

**2. Add validation** (DIP-new):
- Warn if a node has `goal_gate: true` but no retry_target or fallback_target
- Error if a non-agent node has `goal_gate: true` (only agents produce outcomes)

---

### 2.7 **Missing: Restart Semantics**

**Runtime Has** (`engine.go`):
- When an edge targets an already-completed node, triggers "restart loop" handling:
  - Increments global `RestartCount`
  - Checks `max_restarts` (default 5, from graph attr)
  - Clears downstream completion state via BFS
  - Resets retry counters for cleared nodes
  - Jumps to `restart_target` (graph attr) if specified, else the re-entered node
- Used for iterative refinement loops

**Dippin IR Has**:
```go
type Edge struct {
    // ...
    Restart bool // Back-edge: triggers downstream clear + re-execution
}
```

**Gap**: IR has `Restart` flag but doesn't specify the restart target resolution logic or max restart count.

**Proposed Solution**:

**1. Add `max_restarts` and `restart_target` to `WorkflowDefaults`**:
```go
type WorkflowDefaults struct {
    // ...
    MaxRestarts   int    // Default 5
    RestartTarget string // Node to jump to on loop detection (optional)
}
```

**2. Clarify in spec**:
> When an edge with `restart: true` is selected, the engine:
> 1. Increments the global restart counter
> 2. If counter exceeds `max_restarts`, pipeline fails
> 3. Determines restart target: `workflow.restart_target` if set, else `edge.To`
> 4. Clears all nodes downstream of restart target (BFS)
> 5. Resets retry counters for cleared nodes
> 6. Jumps to restart target

**3. Validation** (DIP-new):
- Warn if a workflow has `restart: true` edges but no `max_restarts` (uses default)
- Error if `restart_target` references a non-existent node

---

## 3. Condition Language Alignment

### The Runtime's Condition Evaluator (`condition.go`)

**Operators Supported**:
- Comparison: `=`, `!=`
- String ops: `contains`, `startswith`, `endswith`, `in` (comma-separated list)
- Logical: `&&`, `||`, `not` (prefix)
- Precedence: `||` (lowest) → `&&` → clause-level operators

**Variable Resolution**:
- Direct: `outcome` → looks up `ctx.Get("outcome")`
- Prefixed: `context.outcome` → strips prefix, looks up `outcome`
- No namespacing enforcement (any key can be accessed)

**Examples**:
```
outcome=success
outcome!=fail && tool_stdout contains PASSED
not outcome=retry || human_response startswith yes
status in completed,partial_success,success
```

---

### Dippin's Condition AST (`ir/edge.go`)

**AST Types**:
- `CondAnd`, `CondOr`, `CondNot` (logical combinators)
- `CondCompare` with:
  - `Variable`: **namespace-prefixed** (`ctx.outcome`, `graph.goal`)
  - `Op`: `"="`, `"=="`, `"!="`, `"contains"`, `"startswith"`, `"endswith"`, `"in"`
  - `Value`: string literal

**Examples** (after parsing):
```
CondCompare{Variable: "ctx.outcome", Op: "=", Value: "success"}
CondAnd{
  Left: CondCompare{Variable: "ctx.outcome", Op: "!=", Value: "fail"},
  Right: CondCompare{Variable: "ctx.tool_stdout", Op: "contains", Value: "PASSED"}
}
```

---

### Alignment Analysis

| Feature | Runtime | Dippin IR | Status |
|---------|---------|-----------|--------|
| `=` / `==` | Both work as equality | Both supported | ✅ Aligned |
| `!=` | Supported | Supported | ✅ Aligned |
| `contains` | Word operator | String operator | ✅ Aligned |
| `startswith` | Word operator | String operator | ✅ Aligned |
| `endswith` | Word operator | String operator | ✅ Aligned |
| `in` (list membership) | Comma-separated | String operator | ✅ Aligned |
| `&&` | Supported | `CondAnd` | ✅ Aligned |
| `||` | Supported | `CondOr` | ✅ Aligned |
| `not` (prefix) | Supported | `CondNot` | ✅ Aligned |
| Variable namespacing | Optional `context.` prefix | **Required** `ctx.` or `graph.` | ⚠️ **Breaking** |

---

### **Issue: Namespace Requirement Mismatch**

**Runtime**: Accepts `outcome=success` (no prefix)  
**Dippin IR**: Requires `ctx.outcome=success` (mandatory namespace)

**Impact**: Migrated DOT files would have conditions like `outcome=success` which don't parse in strict Dippin.

**Options**:

**A. Migration Tool Auto-Prefixes** (Recommended):
- During `dippin migrate`, prepend `ctx.` to all non-prefixed variables in conditions
- Document this transformation in migration output
- Example: `outcome=success` → `ctx.outcome=success`

**B. Parser Adds Default Namespace**:
- If a variable has no `.` delimiter, assume `ctx.` prefix
- Disadvantage: Hides namespace semantics, harder to teach

**C. Dippin Spec Allows Bare Variables**:
- Revert to the runtime's behavior: bare variables default to `ctx.`
- Add validation warning if variable doesn't exist in known context keys
- Disadvantage: Less explicit, harder to distinguish `ctx` vs `graph` keys

**Recommendation**: **Option A**. Make namespacing explicit in Dippin source but automatic in migration. This teaches best practices without breaking existing pipelines.

---

### Operator Syntax Recommendation

**Keep the runtime's word-based operators** (`contains`, `startswith`, `endswith`, `in`) in Dippin syntax. They're more readable than symbol alternatives (`~=`, `^=`, `$=`, `∈`) and already widely used in existing pipelines.

**No changes needed** — Dippin's operator set is a perfect match.

---

## 4. Stylesheet / Theming

### The Runtime's Implementation (`stylesheet.go`)

**Syntax**: CSS-like rules in `model_stylesheet` graph attribute (string):
```
* { model: gpt-4o; provider: openai; }
.coder { model: o1; reasoning_effort: high; }
#critical_gate { max_retries: 5; retry_policy: aggressive; }
agent { temperature: 0.7; }
```

**Selectors**:
- `*`: Universal (all nodes)
- `.class`: Class selector (matches `node.Attrs["class"]`)
- `#id`: ID selector (matches `node.ID`)
- `shape`: Shape selector (matches `node.Shape`, e.g., `agent`)

**Specificity**: `*` < `shape` < `.class` < `#id`  
**Resolution**: Apply rules in specificity order (low→high). Explicit node attributes override all stylesheet rules.

**Current Status in Dippin**: Deferred to post-v1 per `docs/DIPPIN_DESIGN_PLAN.md`.

---

### Integration Proposal (v1.5)

**Design Goals**:
1. **Not a string-in-a-string**: Make stylesheet a first-class top-level block, not a DOT attribute
2. **Parse-time validation**: Catch typos in selectors, property names, and values during parsing
3. **Explicit expansion**: Formatter/exporter can inline stylesheet-resolved values for clarity
4. **Simple selectors only**: No combinators, pseudo-classes, or cascading complexity

---

### Proposed Dippin Syntax

```dippin
workflow analyze_codebase:
  start: scan
  exit: report
  
  defaults:
    model: gpt-4o-mini
    provider: openai
  
  stylesheet:
    * {
      temperature: 0.7
      max_tokens: 2000
    }
    
    .coder {
      model: o1
      reasoning_effort: medium
    }
    
    #critical_gate {
      max_retries: 5
      retry_policy: aggressive
    }
  
  agent scan:
    class: coder  # Inherits model, reasoning_effort from .coder
    prompt: "Scan the codebase"
  
  agent critical_gate:
    prompt: "Validate all tests pass"
    # Inherits max_retries, retry_policy from #critical_gate
```

---

### IR Addition

```go
type Stylesheet struct {
    Rules []StyleRule
}

type StyleRule struct {
    Selector   string            // "*" | ".class" | "#id" | kind ("agent", "tool", etc.)
    Properties map[string]string // Key-value pairs to merge into node config
}

type Workflow struct {
    Name       string
    Stylesheet *Stylesheet // Optional, nil if no stylesheet block present
    // ... rest of fields
}
```

---

### Resolution Behavior

**At parse time**:
1. Parse stylesheet block into `StyleRule` list
2. Validate selector syntax and property names
3. Store in IR

**At validation time** (optional warning):
- Check if any selector matches zero nodes (unused rule)

**At runtime or export**:
- For each node, collect matching rules by specificity
- Merge properties in specificity order (low→high)
- Apply explicit node config last (overrides everything)

**Example resolution**:
```
Node: agent scan, class: [coder]

Matches:
  * { temperature: 0.7, max_tokens: 2000 } [specificity: 0]
  agent { model: gpt-4o } [specificity: 1, from defaults]
  .coder { model: o1, reasoning_effort: medium } [specificity: 2]

Resolved config:
  temperature: 0.7        (from *)
  max_tokens: 2000        (from *)
  model: o1               (.coder overrides agent)
  reasoning_effort: medium (from .coder)
```

---

### Migration Strategy

**For v1 (current)**:
- Migration tool reads `model_stylesheet` from DOT graph attrs
- Parses stylesheet and resolves each node's effective config
- Inlines resolved values into each node's explicit fields
- **Lossy**: Original stylesheet intent is lost, nodes become verbose

**For v1.5**:
- Migration tool generates a `stylesheet:` block
- Nodes get `class:` attributes instead of repeated config fields
- **Lossless**: Stylesheet intent is preserved

---

### Implementation Checklist (v1.5)

- [ ] Add `Stylesheet` to IR
- [ ] Add stylesheet parser (CSS-like subset)
- [ ] Add `class:` field to `Node` IR
- [ ] Add DIP-new validation codes for stylesheet errors
- [ ] Update formatter to emit stylesheet block
- [ ] Update DOT exporter to flatten stylesheet to attributes
- [ ] Update migration tool to preserve stylesheets (not inline)

---

## 5. Fidelity and Compaction

### Current State

**Runtime** (`fidelity.go`):
- Six fidelity levels: `full`, `summary:high`, `summary:medium`, `summary:low`, `compact`, `truncate`
- Degradation chain on resume: `full` → `summary:high` → `summary:medium` → `summary:low` → `compact` → `truncate`
- Compaction logic:
  - `full`: No compaction, restore all checkpoint context
  - `summary:high`: All keys + trimmed artifacts (2000 chars per node)
  - `summary:medium`: Only `mediumKeys` (outcome, last_response, human_response, tool_stdout, tool_stderr, goal)
  - `summary:low`: One-line summary per completed node
  - `compact`: Only `graph.goal` + `outcome`
  - `truncate`: Medium keys capped at 500 chars each

**Dippin IR** (`ir.go`):
```go
type AgentConfig struct {
    // ...
    Fidelity            string  // e.g., "full", "summary:high"
    Compaction          string  // "auto" | "none"
    CompactionThreshold float64 // Not yet used by engine
}
```

---

### What's Missing

1. **No fidelity enum**: Valid fidelity values are undocumented in IR
2. **No degradation policy**: IR doesn't specify what happens on resume
3. **No graph-level default**: Fidelity is per-node only (should also be in `WorkflowDefaults`)
4. **Compaction threshold unused**: The field exists but engine doesn't implement token-budget-based compaction yet

---

### Proposed Improvements

#### 5.1 Document Fidelity Levels in Spec

Add to Dippin spec (`docs/spec.md`):

```markdown
### Fidelity Levels

Controls how much prior node context is injected into agent prompts:

- `full`: Complete context from checkpoint (default for first execution)
- `summary:high`: All context keys + trimmed artifact responses (2000 chars per node)
- `summary:medium`: Key decisions only (outcome, last_response, human_response, tool_stdout, tool_stderr, goal)
- `summary:low`: One-line summary per completed node
- `compact`: Only workflow goal + current outcome
- `truncate`: Medium keys capped at 500 chars each

**Degradation on Resume**: When a pipeline resumes from checkpoint, in-memory state is lost, so fidelity degrades one level (e.g., `full` → `summary:high`) and compaction is re-applied.

**Resolution Hierarchy**:
1. Node-level `fidelity:` attribute
2. Workflow-level `defaults.fidelity`
3. Default to `full`
```

#### 5.2 Add Fidelity to WorkflowDefaults

```go
type WorkflowDefaults struct {
    // ...
    Fidelity   string // Default fidelity level for all nodes
    Compaction string // "auto" | "none"
}
```

**Dippin Syntax**:
```
workflow analyze:
  defaults:
    fidelity: summary:high
    compaction: auto
```

#### 5.3 Add Validation for Fidelity Values

**DIP-new: Invalid Fidelity Level**
```
Code: DIP113
Severity: Error
Message: node "analyze" has invalid fidelity "sumary:high" (typo)
Help: valid values are: full, summary:high, summary:medium, summary:low, compact, truncate
```

#### 5.4 Defer Compaction Threshold to v2

The `compaction_threshold` field (intended for token-budget-based compaction) is not yet implemented in the runtime. Mark it as **experimental** in Dippin spec and ignore it in v1 validation.

**Future semantics** (v2+):
```
agent analyze:
  compaction_threshold: 0.75
  # When context size exceeds 75% of model's context window,
  # automatically compact to next-lower fidelity level
```

---

## 6. Event Model

### The Runtime's Event System (`events.go`)

**Event Types**:
- Pipeline lifecycle: `pipeline_started`, `pipeline_completed`, `pipeline_failed`
- Node lifecycle: `stage_started`, `stage_completed`, `stage_failed`, `stage_retrying`
- Control flow: `loop_restart`, `checkpoint_saved`
- Composition: `interview_started`, `interview_completed`, `parallel_started`, `parallel_completed`

**PipelineEvent Structure**:
```go
type PipelineEvent struct {
    Type      PipelineEventType
    Timestamp time.Time
    RunID     string
    NodeID    string
    Message   string
    Err       error
}
```

**Usage**:
- TUI subscribes to events to update UI (green checkmarks, error logs)
- Trace builder collects events for execution logs
- Custom event handlers can emit metrics, logs, webhooks

---

### Should Dippin Be Aware of Events?

**Arguments FOR**:
1. **Observability hooks**: Allow workflows to declare event subscriptions (e.g., "send webhook on `stage_failed`")
2. **Conditional logic**: Enable edges like `on pipeline_failed -> restart`
3. **Declarative monitoring**: Specify SLOs inline (e.g., "alert if `stage_retrying` count > 5")

**Arguments AGAINST**:
1. **Runtime concern**: Events are execution-time artifacts, not workflow structure
2. **Separation of concerns**: Observability should be configured at the engine level, not per-workflow
3. **Complexity**: Adding event semantics to IR bloats the language without clear authoring value
4. **The runtime already handles it**: The engine emits events; the TUI/handlers consume them; workflows don't need to know

---

### Recommendation: **No Change for v1**

Events are a **runtime observability layer**, not a workflow modeling primitive. Keep them in the engine, not the IR.

**Post-v1 Extension** (if needed):
Add an optional `on_event:` hook for external integrations:
```
workflow deploy:
  on_event:
    stage_failed:
      webhook: https://alerts.example.com/pipeline-failed
      retry: true
```

This would be syntactic sugar for configuring the engine's event handler, not a first-class IR concept.

---

## 7. Validation Convergence

### Current State

**Runtime** (`validate.go`, `validate_semantic.go`):
- `Validate(g *Graph)`: Structural checks (start/exit exist, no cycles, all edges reference real nodes, reachability)
- `ValidateSemantic(g *Graph, registry *HandlerRegistry)`: Handler registration, condition syntax, attribute types
- Returns `ValidationError` with separate `Errors` and `Warnings` lists
- Ad-hoc error messages (no codes)

**Dippin** (`validator/validate.go`):
- Structured checks with error codes (DIP001–DIP009+)
- Multi-diagnostic collection (reports all errors, not just first)
- Severity levels (Error, Warning, Info, Hint)
- Location tracking (file, line, column)
- Help text with suggested fixes

---

### Validation Coverage Comparison

| Check | Runtime | Dippin | Code | Notes |
|-------|---------|--------|------|-------|
| Start node exists | ✅ | ✅ | DIP001 | Aligned |
| Exit node exists | ✅ | ✅ | DIP002 | Aligned |
| Edge endpoints valid | ✅ | ✅ | DIP003 | Dippin adds "did you mean?" |
| All nodes reachable | ✅ | ✅ | DIP004 | Aligned |
| No unconditional cycles | ✅ | ✅ | DIP005 | The runtime excludes conditional edges; Dippin excludes `restart: true` |
| No duplicate edges | ✅ | ✅ | DIP006 | Aligned |
| Exit has no outgoing edges | ✅ | ✅ | DIP007 | Aligned |
| No duplicate node IDs | ❌ | ✅ | DIP008 | **Runtime missing** |
| Parallel/fan_in pairing | ❌ | ✅ | DIP009 | **Runtime missing** |
| Handler registration | ✅ | ❌ | — | **Dippin missing** (needs registry input) |
| Condition syntax | ✅ | ❌ | — | **Dippin missing** (parser validates during parse, not in validator) |
| Attribute types | ✅ | ❌ | — | **Dippin missing** (should be DIP-new codes) |
| Conditional node fail edges | ⚠️ Warning | ❌ | — | **Dippin missing** |
| Edge label consistency | ⚠️ Warning | ❌ | — | **Dippin missing** |

---

### Convergence Recommendations

#### 7.1 **The Runtime Should Delegate Structural Validation to Dippin**

**Why**: Dippin's validator is more rigorous (error codes, location tracking, multi-diagnostic) and catches issues the runtime misses (duplicate nodes, parallel/fan_in pairing).

**How**:
1. Add `dippin/validator.Validate(ir.Workflow)` call in the runtime CLI before converting IR to Graph
2. If validation fails, print diagnostics and exit
3. Remove redundant checks from `pipeline/validate.go` (keep only Graph-specific runtime checks if any)

**Benefits**:
- Single source of truth for validation rules
- Better error messages for users
- No drift between Dippin and runtime validation

---

#### 7.2 **Dippin Validator Should Add Semantic Checks**

**Missing checks to add**:

**DIP110: Handler Registration** (semantic)
```
Code: DIP110
Severity: Error
Message: node "analyze" references unregistered handler "codergen"
Help: ensure the handler is registered in the engine's HandlerRegistry before execution
```
**Implementation**: `ValidateSemantic(w *Workflow, registry *HandlerRegistry)` function (mirrors the runtime's)

**DIP111: Invalid Attribute Types** (semantic)
```
Code: DIP111
Severity: Error
Message: node "retry_node" has invalid max_retries "abc": must be a non-negative integer
Help: change max_retries to a valid integer (e.g., 3)
```

**DIP112: Missing Fail Edge on Conditional Node** (warning)
```
Code: DIP112
Severity: Warning
Message: node "validate" is a conditional but has no fail edge
Help: add an edge with condition "ctx.outcome != success" or "ctx.outcome = fail"
```

**DIP113: Inconsistent Edge Label Usage** (warning)
```
Code: DIP113
Severity: Warning
Message: node "validate" has 3 outgoing edges but only 2 are labeled
Help: either label all edges or remove all labels for consistency
```

**DIP114: Invalid Fidelity Level** (error)
```
Code: DIP114
Severity: Error
Message: node "analyze" has invalid fidelity "sumary:high"
Help: valid fidelity levels are: full, summary:high, summary:medium, summary:low, compact, truncate
```

---

#### 7.3 **Validation Layering**

Organize validation into tiers:

**Tier 1: Parse-Time** (parser responsibility)
- Syntax errors (missing colons, bad indentation)
- Required fields (agent without prompt)
- Condition AST parsing

**Tier 2: Structural** (validator, no external input)
- DIP001–DIP009 (start/exit, reachability, cycles, etc.)
- Graph topology correctness

**Tier 3: Semantic** (validator with registry)
- DIP110–DIP114 (handler registration, attribute types, fidelity values)
- Requires engine context (handler registry, known context keys)

**Tier 4: Runtime** (engine)
- Edge condition evaluation (dynamic context values)
- Timeout enforcement
- Resource limits

---

#### 7.4 **Autofix Recommendation**

The runtime has `AutoFix(g *Graph)` which adds self-loop retry edges to conditional nodes missing fail edges.

**Dippin should add similar capability**:
```bash
$ dippin validate --fix workflow.dip
```

**Fixes to apply**:
1. DIP112: Add fail edge to conditional nodes
2. DIP003: Suggest node ID correction for typos (Levenshtein distance ≤ 2)
3. DIP113: Normalize edge labels (all or none)

**Output**:
```
Applied 3 fixes:
  - Added fail edge validate -> validate (condition: ctx.outcome = fail)
  - Renamed edge target "anlyze" to "analyze"
  - Removed inconsistent edge labels on node "router"
```

---

## 8. Composition Model

### The Runtime's Subgraph Handling (`subgraph.go`)

**Current Implementation**:
- `SubgraphHandler` looks up referenced graph by name from a pre-loaded map
- Parent context is passed to child via `WithInitialContext(pctx.Snapshot())`
- Child result context is merged back into parent
- No parameter declaration, namespace isolation, or import resolution

**Limitations**:
1. **No file-based imports**: Subgraphs must be pre-loaded into the engine's graph map
2. **No parameter contracts**: No way to declare what parameters a subgraph expects
3. **No namespace isolation**: Child context keys can collide with parent
4. **No composition validation**: Can't validate a subgraph reference at parse time

---

### Dippin's SubgraphConfig (`ir.go`)

```go
type SubgraphConfig struct {
    Ref    string            // Workflow name or path
    Params map[string]string // Parameter overrides
}
```

**Improvements over the runtime**:
- `Ref` can be a file path (e.g., `./security_scan.dip`)
- `Params` allows passing values to child workflow

**Still Missing**:
- Parameter declaration in child workflow
- Import resolution (how does engine load `./security_scan.dip`?)
- Namespace prefixing for context keys
- Inline expansion vs runtime dispatch

---

### Proposed Improvements (v1.5)

#### 8.1 **Add Parameter Declarations**

Child workflows declare parameters:
```
workflow security_scan:
  params:
    repo_path: string          # Required
    severity: string = "high"  # Optional with default
  
  start: scan
  exit: report
  
  tool scan:
    command: ./scan.sh ${params.repo_path} --severity ${params.severity}
```

Parent workflows pass parameters:
```
subgraph run_security_scan:
  ref: ./security_scan.dip
  params:
    repo_path: /tmp/repo
    severity: critical
```

**IR Addition**:
```go
type WorkflowParam struct {
    Name    string
    Type    string // "string" | "int" | "bool"
    Default string // Empty = required
}

type Workflow struct {
    // ...
    Params []WorkflowParam
}
```

---

#### 8.2 **File-Based Import Resolution**

**Dippin Syntax**:
```
subgraph run_scan:
  ref: ./security_scan.dip
  params:
    repo_path: /tmp/repo
```

**Resolution Logic** (in parser/validator):
1. Resolve `ref` relative to parent file's directory
2. Parse referenced `.dip` file
3. Validate parameter compatibility:
   - All required params are provided
   - Provided params match declared names
4. Store resolved `ir.Workflow` in IR (inline expansion) or store just the path (runtime dispatch)

---

#### 8.3 **Namespace Isolation**

**Problem**: Child workflow's context keys collide with parent's.

**Example**:
- Parent sets `ctx.outcome = "parent_success"`
- Child sets `ctx.outcome = "child_success"`
- After child completes, parent's `ctx.outcome` is overwritten

**Solution**: Prefix child context keys with subgraph node ID:
```
Parent context before:
  ctx.outcome = "parent_success"

Child context (isolated):
  run_scan.ctx.outcome = "child_success"
  run_scan.ctx.vulnerabilities = "CVE-2024-1234"

Parent context after merge:
  ctx.outcome = "parent_success"  # Preserved
  run_scan.ctx.outcome = "child_success"
  run_scan.ctx.vulnerabilities = "CVE-2024-1234"
```

**Dippin Syntax** for accessing child context:
```
agent analyze_results:
  prompt: |
    Review the scan results:
    Vulnerabilities: ${run_scan.ctx.vulnerabilities}
```

---

#### 8.4 **Inline Expansion vs Runtime Dispatch**

**Two strategies for subgraph execution**:

**A. Inline Expansion** (compile-time):
- Parser loads child workflow and expands it into parent graph
- All nodes/edges are flattened into one IR
- Validation happens at parse time
- Pro: No runtime file I/O, easier debugging
- Con: Large workflows blow up IR size

**B. Runtime Dispatch** (current runtime behavior):
- `SubgraphConfig.Ref` is resolved at runtime
- Engine loads child graph when node executes
- Pro: Smaller IR, supports dynamic subgraphs
- Con: Harder to validate at parse time, runtime errors

**Recommendation**: **Support both**
- Default to **inline expansion** for static refs (`.dip` files)
- Support **runtime dispatch** for dynamic refs (e.g., `${workflow_name}.dip`)
- Add a `mode:` option:
  ```
  subgraph run_scan:
    ref: ./security_scan.dip
    mode: inline  # or "runtime"
  ```

---

#### 8.5 **Validation for Subgraphs**

**DIP-new codes**:

**DIP201: Subgraph Not Found**
```
Code: DIP201
Severity: Error
Message: subgraph reference "./security_scan.dip" not found
Help: ensure the file exists relative to the current workflow directory
```

**DIP202: Missing Required Parameter**
```
Code: DIP202
Severity: Error
Message: subgraph "security_scan" requires parameter "repo_path" but it was not provided
Help: add repo_path: <value> to the params block
```

**DIP203: Unknown Parameter**
```
Code: DIP203
Severity: Warning
Message: subgraph "security_scan" does not accept parameter "unknown_param"
Help: remove the parameter or check the subgraph's params declaration
```

**DIP204: Parameter Type Mismatch**
```
Code: DIP204
Severity: Error
Message: parameter "severity" expects type int but got "critical" (string)
Help: change the value to a valid integer
```

---

## 9. Priority Recommendations

### Top 5 Changes by Impact and Effort

| # | Recommendation | Impact | Effort | Priority | Target |
|---|----------------|--------|--------|----------|--------|
| **1** | **Add Retry Policy Documentation** | High | Low | **P0** | v1.1 |
| **2** | **Converge Validation (Runtime → Dippin)** | High | Medium | **P0** | v1.2 |
| **3** | **Add Fidelity Enum + Validation** | Medium | Low | **P1** | v1.2 |
| **4** | **Implement Stylesheet System** | High | High | **P1** | v1.5 |
| **5** | **Add Subgraph Parameter System** | High | High | **P1** | v1.5 |

---

### 1. Add Retry Policy Documentation
**Impact**: Closes a critical spec gap. The runtime relies on five named policies but Dippin doesn't document them.

**Effort**: Low — just documentation + validation.

**Changes**:
- Update Dippin spec to enumerate valid `retry_policy` values
- Add `base_delay` field to `RetryConfig` (optional)
- Add DIP-new code for invalid policy names

**Why P0**: Blocks correct migration of pipelines using `retry_policy: aggressive` etc. Without this, users don't know what values are valid.

---

### 2. Converge Validation (Runtime → Dippin)
**Impact**: Eliminates dual maintenance, improves error messages, catches bugs earlier.

**Effort**: Medium — requires refactoring the runtime CLI to call Dippin validator first.

**Changes**:
- Add `ValidateSemantic(w *Workflow, registry *HandlerRegistry)` to Dippin
- Add DIP110–DIP114 codes (handler registration, attribute types, warnings)
- Wire Dippin validator into the runtime CLI before `IRToGraph()`
- Deprecate `pipeline/validate.go` checks that overlap with Dippin

**Why P0**: Validation is currently split between two codebases. Consolidating it in Dippin reduces bugs and improves UX.

---

### 3. Add Fidelity Enum + Validation
**Impact**: Prevents typos in fidelity values, documents degradation behavior.

**Effort**: Low — add validation + spec docs.

**Changes**:
- Document six fidelity levels in spec
- Add `Fidelity` to `WorkflowDefaults`
- Add DIP114 code for invalid fidelity values
- Clarify degradation chain and resume behavior in spec

**Why P1**: Important for correctness but not blocking migration (fidelity is optional, defaults to `full`).

---

### 4. Implement Stylesheet System
**Impact**: Unlocks cleaner, more maintainable workflows. Avoids repeating `model: o1` on every agent.

**Effort**: High — new parser, IR types, resolution logic, formatter changes.

**Changes**:
- Add `stylesheet:` top-level block to Dippin syntax
- Add `Stylesheet` to IR
- Add `class:` field to nodes
- Implement CSS-like parser with specificity resolution
- Update formatter and DOT exporter
- Update migration tool to preserve stylesheets (not inline)

**Why P1**: High impact on readability but not critical for v1 (inlining works fine for now).

---

### 5. Add Subgraph Parameter System
**Impact**: Enables true composition with parameter contracts and namespace isolation.

**Effort**: High — parser changes, IR additions, validation codes, namespace prefixing logic.

**Changes**:
- Add `params:` block to workflow declarations
- Add `WorkflowParam` to IR
- Implement parameter validation (DIP201–DIP204)
- Add namespace prefixing for child context keys
- Support inline expansion mode

**Why P1**: Critical for multi-file workflows but can be deferred until more pipelines need composition.

---

### Other Recommendations (Deferred to v2+)

- **Goal Gate Semantics Documentation** (Low effort, medium impact) — just spec clarification
- **Restart Target Configuration** (Low effort, low impact) — works fine with defaults
- **Event Model Hooks** (High effort, low impact) — observability is already handled externally
- **Compaction Threshold Implementation** (High effort, low impact) — token budgets not needed yet
- **Parallel Branch Configuration** (Medium effort, medium impact) — wait for more parallel use cases

---

## Conclusion

### What Works

Dippin's v1 design is **structurally sound**:
- Explicit start/exit removes ambiguity ✅
- Typed node configs prevent invalid states ✅
- Multiline blocks without escaping improve readability ✅
- Condition AST enables parse-time validation ✅
- Error code registry enables structured diagnostics ✅

### What's Missing

The runtime has evolved **runtime features** that should inform Dippin v1.5+:
- **Retry policies**: Need documented enum + `base_delay` field
- **Fidelity modes**: Need spec docs + validation + degradation policy
- **Stylesheets**: Need first-class syntax block (not string-in-attribute)
- **Subgraph params**: Need contracts, namespace isolation, import resolution
- **Validation convergence**: Need to consolidate runtime and Dippin validators

### Next Steps

**Immediate** (v1.1–v1.2):
1. Document retry policies and fidelity levels in Dippin spec
2. Add validation codes DIP110–DIP114 (semantic checks)
3. Wire Dippin validator into the runtime CLI
4. Add namespace auto-prefixing in migration tool for conditions

**Near-term** (v1.5):
1. Implement stylesheet system as first-class syntax
2. Add subgraph parameter declarations + validation
3. Implement namespace isolation for composed workflows

**Long-term** (v2+):
1. Token-budget-based context compaction
2. Event model hooks for observability
3. Advanced composition (dynamic refs, conditional imports)

---

**Report compiled on**: March 20, 2025  
**Dippin version analyzed**: v1.0 (post-migration verification)  
**Runtime version analyzed**: Latest (includes stylesheets, fidelity, retry policies)
