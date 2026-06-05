# Validation and Linting Reference

Dippin registers 55 diagnostic codes split into two categories; this page gives a dedicated section to every code except `DIP138`, which is reserved and has no firing logic (54 documented sections):

- **Structural validation** (DIP001–DIP009): Errors that **must** be fixed. A workflow with any of these cannot execute.
- **Semantic linting** (DIP101–DIP146): Warnings that flag likely bugs or questionable patterns. They don't block execution but should be reviewed.

Run `dippin validate <file>` for structural checks only, or `dippin lint <file>` for both.

```mermaid
graph LR
    SRC[".dip file"] --> PARSE["Parser"]
    PARSE --> IR["IR"]
    IR --> VAL["Structural Validation<br>DIP001–DIP009<br>(errors)"]
    IR --> LINT["Semantic Linting<br>DIP101–DIP146<br>(warnings)"]
    VAL --> DIAG["Diagnostics"]
    LINT --> DIAG
```

---

## Diagnostic Format

Diagnostics are displayed in a rustc-inspired format:

```
error[DIP003]: unknown node reference "InterpretX" in edge
  --> pipeline.dip:45:5
  = help: did you mean "Interpret"?
```

Each diagnostic has:

| Field | Description |
|-------|-------------|
| `Code` | Unique identifier (e.g., DIP003) |
| `Severity` | `error`, `warning`, `info`, or `hint` |
| `Message` | Human-readable explanation of the issue |
| `Location` | File, line, and column where the issue was detected |
| `Help` | Optional suggestion for fixing the issue |
| `Fix` | Optional replacement text |

In JSON output mode (`--format json`), diagnostics are emitted as an array of objects with these exact fields.

---

## Structural Validation Errors (DIP001–DIP009)

### DIP001: Start Node Missing

**Severity**: Error

The workflow must declare a `start:` field pointing to an existing node.

```
error[DIP001]: start node does not exist
  --> pipeline.dip:1:1
  = help: add "start: <NodeID>" to the workflow header
```

**What triggers it**:
- The `start:` field is missing entirely
- The `start:` field references a node ID that doesn't exist

**How to fix**: Add or correct the `start:` field in the workflow header:
```dippin
workflow MyPipeline
  start: FirstNode    # Must match an actual node ID
  exit: LastNode
```

---

### DIP002: Exit Node Missing

**Severity**: Error

The workflow must declare an `exit:` field pointing to an existing node.

```
error[DIP002]: exit node does not exist
  --> pipeline.dip:1:1
  = help: add "exit: <NodeID>" to the workflow header
```

**What triggers it**:
- The `exit:` field is missing entirely
- The `exit:` field references a node ID that doesn't exist

**How to fix**: Add or correct the `exit:` field in the workflow header.

---

### DIP003: Unknown Node Reference in Edge

**Severity**: Error

Every edge's `From` and `To` must reference existing node IDs.

```
error[DIP003]: unknown node reference "InterpretX" in edge
  --> pipeline.dip:45:5
  = help: did you mean "Interpret"?
```

**What triggers it**: An edge references a node that hasn't been declared in the workflow.

**Smart suggestions**: The validator uses Levenshtein distance (edit distance ≤ 2) to suggest corrections for typos.

**How to fix**: Either correct the node name in the edge, or declare the missing node.

---

### DIP004: Unreachable Node from Start

**Severity**: Error

Every node must be reachable from the start node via some path of edges.

```
error[DIP004]: node unreachable from start
  --> pipeline.dip:20:3
  = help: add an edge leading to this node, or remove it
```

**What triggers it**: BFS from the start node cannot reach this node. The node is an island — connected to nothing or only connected to other unreachable nodes.

**How to fix**: Either add edges that connect this node to the main graph, or remove the orphaned node.

---

### DIP005: Unconditional Cycle Detected

**Severity**: Error

The workflow graph must be a DAG (directed acyclic graph), with the exception of restart edges.

```
error[DIP005]: unconditional cycle detected
  --> pipeline.dip:50:5
  = help: remove an edge in this cycle or mark it "restart: true"
```

**What triggers it**: DFS finds a back-edge that is not marked `restart: true`. This means the pipeline would loop forever.

**Important**: Restart edges (`restart: true`) are excluded from cycle detection. They are the intentional mechanism for controlled iteration.

**How to fix**: Either remove an edge to break the cycle, or mark the back-edge as `restart: true` to make it a controlled loop.

---

### DIP006: Exit Node Has Outgoing Edges

**Severity**: Error

The exit node is the terminal — it must have zero outgoing edges.

```
error[DIP006]: exit node has outgoing edges
  --> pipeline.dip:55:5
  = help: remove outgoing edges from the exit node
```

**What triggers it**: An edge has the exit node as its `From` field.

**How to fix**: Remove the outgoing edge. If you need processing after the current exit node, make a new exit node.

---

### DIP007: Parallel/Fan-In Mismatch

**Severity**: Error

Every `parallel` node must have a matching `fan_in` node with the same set of branch nodes.

```
error[DIP007]: parallel fan-out/fan-in mismatch
  --> pipeline.dip:15:3
  = help: add a matching fan_in node
```

**What triggers it**:
- A `parallel` node exists without a corresponding `fan_in`
- A `fan_in` node exists without a corresponding `parallel`
- The target/source sets don't match between the pair

**How to fix**: Ensure every `parallel P -> A, B, C` has a matching `fan_in J <- A, B, C` with the same nodes (order doesn't matter).

---

### DIP008: Duplicate Node ID

**Severity**: Error

Node IDs must be globally unique within a workflow.

```
error[DIP008]: duplicate node ID
  --> pipeline.dip:30:3
  = help: rename this node or remove the duplicate
```

**What triggers it**: Two nodes share the same ID.

**How to fix**: Rename one of the duplicate nodes.

---

### DIP009: Duplicate Edge

**Severity**: Error

No two edges may have the same (from, to, condition) combination.

```
error[DIP009]: duplicate edge
  --> pipeline.dip:60:5
  = help: remove the duplicate edge
```

**What triggers it**: Two edges with identical source, target, and condition raw text.

**Note**: Edges with different conditions on the same (from, to) pair are **not** duplicates — that's intentional conditional branching.

---

## Semantic Lint Warnings (DIP101–DIP146)

### DIP101: Node Only Reachable via Conditional Edges

**Severity**: Warning

A node where **all** incoming edges have conditions may be unreachable at runtime if no condition matches.

```
warning[DIP101]: node "NextPhase" is only reachable through conditional edges and may be skipped at runtime
  --> pipeline.dip:25:3
  = help: add an unconditional edge to this node, or verify all conditions are exhaustive
```

**What triggers it**: Every edge leading to this node has a `when` clause. If none of those conditions evaluate to true, execution can never reach this node.

**Exhaustive suppression**: DIP101 is automatically suppressed when **all** source nodes feeding into this node have exhaustive outgoing conditions. For example, if node Gate has edges `Gate -> A when ctx.outcome = success` and `Gate -> B when ctx.outcome = fail`, both A and B are guaranteed reachable because `success` + `fail` covers all outcomes.

Known exhaustive sets: `ctx.outcome` / `outcome` with values `{success, fail}` or `{success, failure}`.

**How to fix**: Add an unconditional incoming edge, or ensure the source node's conditions are exhaustive.

---

### DIP102: Routing Node Missing Default Edge

**Severity**: Warning

A node with conditional outgoing edges but no unconditional fallback.

```
warning[DIP102]: node "Gate" has conditional outgoing edges but no unconditional default edge
  --> pipeline.dip:35:3
  = help: add an unconditional edge as a fallback, or ensure conditions are exhaustive
```

**What triggers it**: A node has one or more outgoing edges with `when` conditions, but no outgoing edge without a condition.

**Exhaustive suppression**: DIP102 is automatically suppressed when the node's outgoing conditions form an exhaustive set. For example, `when ctx.outcome = success` + `when ctx.outcome = fail` covers all cases — no default edge needed.

**Why it matters**: If no condition matches at runtime and conditions are not exhaustive, execution gets stuck — there's no default path to follow.

**How to fix**: Add an unconditional fallback edge, or ensure conditions are exhaustive:
```dippin
  edges
    Check -> Pass when ctx.outcome = success
    Check -> Retry when ctx.outcome = retry
    Check -> Fail    # unconditional fallback
```

---

### DIP103: Overlapping Conditions

**Severity**: Warning

Multiple edges from the same node test the same variable for the same value.

```
warning[DIP103]: overlapping or contradictory conditions
  --> pipeline.dip:45:5
```

**What triggers it**: Two edges from node A both check `ctx.outcome = success`. One will shadow the other.

**How to fix**: Remove or merge the duplicate condition.

---

### DIP104: Unbounded Retry Loop

**Severity**: Warning

A retry path has no `max_retries` or `fallback_target` to bound it.

```
warning[DIP104]: unbounded retry loop (no max_retries or fallback)
  --> pipeline.dip:40:3
```

**What triggers it**: A node has a `retry_target` set but no `max_retries` or `fallback_target`. This could cause infinite retries.

**How to fix**: Set `max_retries` and/or `fallback_target`:
```dippin
  agent Validate
    max_retries: 3
    retry_target: Implement
    fallback_target: ManualReview
```

---

### DIP105: No Success Path to Exit

**Severity**: Warning

There is no guaranteed path from start to exit through unconditional edges alone.

```
warning[DIP105]: no success path from start to exit
  --> pipeline.dip:1:1
```

**What triggers it**: Every path from start to exit goes through at least one conditional edge. If conditions don't match, execution may never reach the exit.

**How to fix**: Ensure at least one complete path from start to exit uses only unconditional edges or has unconditional fallbacks at every decision point.

---

### DIP106: Undefined Variable in Prompt

**Severity**: Warning

A prompt references a variable that no upstream node is known to produce.

```
warning[DIP106]: undefined variable reference in prompt
  --> pipeline.dip:22:5
```

**What triggers it**: The prompt contains `${ctx.something}` or `${params.key}` where `something` is not a reserved key and no upstream node declares it in `writes`.

**How to fix**: Either add the key to an upstream node's `writes`, or verify the variable name is correct.

---

### DIP107: Unused Context Write

**Severity**: Warning

A node produces a context key that no downstream node reads.

```
warning[DIP107]: unused context key (written but never read)
  --> pipeline.dip:18:5
```

**What triggers it**: A node declares `writes: summary` but no downstream node has `reads: summary` or references `${ctx.summary}` in its prompt.

**How to fix**: Remove the unused `writes` declaration, or add a downstream consumer.

---

### DIP108: Unknown Model/Provider

**Severity**: Warning

The model or provider isn't in the engine's recognized list.

```
warning[DIP108]: unknown model/provider combination
  --> pipeline.dip:15:5
```

**What triggers it**: A model or provider string doesn't match any known LLM provider.

**How to fix**: Check for typos. Use recognized model/provider combinations.

---

### DIP109: Namespace Collision in Imports

**Severity**: Warning

A subgraph parameter name conflicts with an existing context key.

```
warning[DIP109]: namespace collision in imports
  --> pipeline.dip:28:5
```

**What triggers it**: A `subgraph` node's `params` map contains a key that shadows an existing context variable.

**How to fix**: Rename the parameter to avoid the collision.

---

### DIP110: Empty Prompt on Agent

**Severity**: Warning

An agent node has no prompt text.

```
warning[DIP110]: empty prompt on agent node
  --> pipeline.dip:12:3
```

**What triggers it**: An agent node is defined without a `prompt` field or with an empty prompt.

**Why it matters**: An agent without a prompt won't produce meaningful output.

**How to fix**: Add a prompt, or change the node kind if it doesn't need one.

---

### DIP111: Tool Without Timeout

**Severity**: Warning

A tool node has no `timeout` field.

```
warning[DIP111]: tool command has no timeout
  --> pipeline.dip:35:3
```

**What triggers it**: A tool node defines a `command` but no `timeout`.

**Why it matters**: Without a timeout, a hanging command blocks the pipeline indefinitely.

**How to fix**: Add a timeout:
```dippin
  tool RunTests
    timeout: 60s
    command:
      pytest
```

---

### DIP112: Reads Key Not Produced Upstream

**Severity**: Warning

A node declares a `reads` key that no upstream node produces.

```
warning[DIP112]: reads key not produced by any upstream writes
  --> pipeline.dip:25:3
```

**What triggers it**: A node has `reads: plan` but no node upstream (reachable via incoming edges from start) declares `writes: plan`.

**How to fix**: Either add the key to an upstream node's `writes` or remove it from `reads`.

---

### DIP113: Invalid Retry Policy Name

**Severity**: Warning

A node or workflow default specifies a `retry_policy` value that is not a recognized policy name.

```
warning[DIP113]: node "analyze" has retry_policy "agressive" which is not a recognized policy name
  --> pipeline.dip:15:3
  = help: valid policies: standard, aggressive, patient, linear, none
```

**Valid policies**:

| Policy | Backoff | Description |
|--------|---------|-------------|
| `standard` | Exponential | Default. 3 attempts, exponential backoff from base delay |
| `aggressive` | Exponential | More attempts, shorter initial delay |
| `patient` | Exponential | Fewer attempts, longer delays between retries |
| `linear` | Linear | Fixed delay between attempts |
| `none` | — | No retries (node fails immediately on error) |

**How to fix**: Check for typos. Use one of the five recognized policy names.

---

### DIP114: Invalid Fidelity Level

**Severity**: Warning

A node or workflow default specifies a `fidelity` value that is not a recognized level.

```
warning[DIP114]: node "analyze" has fidelity "sumary:high" which is not a recognized level
  --> pipeline.dip:12:3
  = help: valid levels: full, summary:high, summary:medium, summary:low, compact, truncate
```

**Valid fidelity levels**:

| Level | Context Injected | Use Case |
|-------|-----------------|----------|
| `full` | Complete context from all prior nodes | Default for first execution |
| `summary:high` | All keys + trimmed artifacts (2000 chars/node) | Reduce context for large pipelines |
| `summary:medium` | Key decisions only (outcome, last_response, human_response) | Moderate context reduction |
| `summary:low` | One-line summary per completed node | Minimal context |
| `compact` | Only workflow goal + current outcome | Near-zero context |
| `truncate` | Medium keys capped at 500 chars each | Hard size limit |

**Degradation on resume**: When a pipeline resumes from checkpoint, fidelity degrades one level (e.g., `full` → `summary:high`).

**How to fix**: Check for typos. Use one of the six recognized levels.

---

### DIP115: Goal Gate Without Recovery Path

**Severity**: Warning

A node has `goal_gate: true` but no `retry_target` or `fallback_target`, meaning the pipeline has no recovery path if the gate fails.

```
warning[DIP115]: node "validate_tests" has goal_gate: true but no retry_target or fallback_target
  --> pipeline.dip:18:3
  = help: add retry_target or fallback_target so the pipeline can recover when the gate fails
```

**What `goal_gate` means**: When a node with `goal_gate: true` completes with `outcome != success`, the pipeline fails at exit — even if the exit node itself succeeded. Goal gates enforce invariants (e.g., "all tests must pass").

**How to fix**: Add `retry_target: <node>` to retry from an earlier point, or `fallback_target: <node>` to route to a recovery path.

---

### DIP116: Invalid Compaction Threshold or On-Resume Value

**Severity**: Warning

Configuration values for `compaction_threshold` or `on_resume` are outside valid ranges.

```
warning[DIP116]: node "Analyze" has compaction_threshold 1.50 outside valid range [0.0, 1.0]
  --> pipeline.dip:15:3
```

**What triggers it**:
- `compaction_threshold` is set to a value outside `[0.0, 1.0]`
- `on_resume` is set to a value other than `"preserve"` or `"degrade"`
- `on_resume` is set without `fidelity` being configured

**How to fix**: Use a threshold between 0.0 and 1.0, and set on_resume to `"preserve"` or `"degrade"`.

---

### DIP117: Stylesheet References Undefined Class

**Severity**: Warning

A stylesheet rule targets a class that no node declares.

```
warning[DIP117]: stylesheet references class "critical" which is not declared on any node
  --> pipeline.dip:80:5
  = help: add class: critical to a node declaration
```

**How to fix**: Add `class: critical` to the relevant node, or fix the class name in the stylesheet.

---

### DIP118: Stylesheet References Unknown Node ID

**Severity**: Warning

A stylesheet rule targets a node ID that doesn't exist.

```
warning[DIP118]: stylesheet references node ID "Analize" which does not exist
  --> pipeline.dip:82:5
```

**How to fix**: Fix the node ID spelling in the stylesheet selector.

---

### DIP119: Invalid Reasoning Effort

**Severity**: Warning

A node specifies a `reasoning_effort` value that isn't recognized.

```
warning[DIP119]: node "Analyze" has reasoning_effort "max" which is not a recognized level
  --> pipeline.dip:12:3
  = help: valid levels: none, minimal, low, medium, high, xhigh, max
```

**Valid levels**: `none`, `minimal`, `low`, `medium`, `high`, `xhigh`, `max`

**How to fix**: Use one of the seven recognized levels.

---

### DIP120: Condition Variable Missing Namespace Prefix

**Severity**: Warning

A condition references a variable without a namespace prefix.

```
warning[DIP120]: condition variable "outcome" should use a namespace prefix (e.g., ctx.outcome)
  --> pipeline.dip:45:5
```

**What triggers it**: An edge condition uses a bare variable name like `outcome` instead of `ctx.outcome`.

**How to fix**: Add the appropriate namespace prefix (`ctx.`, `graph.`, `params.`).

---

### DIP121: Condition References Variable Not in Source Writes

**Severity**: Warning

An edge condition references a variable that the source node doesn't declare in its `writes`.

```
warning[DIP121]: edge Gate → Pass: condition references "ctx.score" but node "Gate" does not declare it in writes
  --> pipeline.dip:45:5
  = help: add writes: score to node "Gate", or use a reserved variable
```

**What triggers it**: An edge from node A has a condition like `ctx.score = high`, but node A's `IO.Writes` doesn't include `score`. Only fires when the source node has non-empty `writes` declarations.

**Skipped for**: Reserved runtime variables (`ctx.outcome`, `ctx.status`, `ctx.tool_stdout`, `ctx.tool_stderr`, `ctx.tool_marker`, `ctx.tool_route`, `ctx.internal.*`, `graph.*`, `params.*`).

**How to fix**: Add the variable to the source node's `writes` list, or verify the condition references the correct variable.

---

### DIP122: Condition Tests Value Not in Tool Outputs

**Severity**: Warning

An edge condition tests a value that the source tool node doesn't declare in its `outputs`.

```
warning[DIP122]: edge RunTest → Pass: condition tests value "retry" but tool "RunTest" does not declare it in outputs
  --> pipeline.dip:50:5
  = help: add "retry" to tool "RunTest" outputs, or check for typos
```

**What triggers it**: A tool node declares `outputs: ["success", "fail"]` but an outgoing edge condition checks `ctx.outcome = retry`. Only fires for tool nodes with explicitly declared outputs.

**How to fix**: Add the missing value to the tool's `outputs` list, or fix the typo in the condition.

---

### DIP123: Tool Command Shell Syntax Error

**Severity**: Warning

The tool command block has a shell syntax error detectable by `bash -n`.

```
warning[DIP123]: tool command has shell syntax error: unexpected EOF while looking for matching `"'
  --> pipeline.dip:45:5
```

**What triggers it**: Running `bash -n` on the command block reports an error — unclosed quotes, bad redirects, missing `fi`/`done`, etc.

**How to fix**: Fix the syntax error. Test your command in a terminal: `echo 'your command' | bash -n`

---

### DIP124: Tool Command References Runtime Variable

**Severity**: Warning

A tool command contains `${ctx.*}` interpolation that won't resolve at shell execution time.

```
warning[DIP124]: tool command references ${ctx.api_url} which expands to empty at runtime
  --> pipeline.dip:50:5
```

**What triggers it**: The command block contains patterns like `${ctx.outcome}`, `${ctx.tool_stdout}`, etc. These are Dippin runtime variables — the shell sees them as undefined variables that expand to empty strings.

**How to fix**: Pass context values through environment variables set by the pipeline runner, or use file-based IPC (write to a shared `.ai/` directory).

---

### DIP125: Tool Command Binary Not Found

**Severity**: Hint

The first non-preamble command in the tool block references a binary not found on the current PATH.

```
hint[DIP125]: tool command binary "npx" not found on PATH
  --> pipeline.dip:55:5
```

**What triggers it**: The linter extracts the first real command (skipping `set -eu`, `cd`, `export`, `mkdir -p` preamble lines) and checks if the binary exists via `exec.LookPath`.

**How to fix**: Install the missing binary, or use a full path. Note: this check runs on the developer's machine — the deployment environment may have different binaries available.

**Caveat**: This is a hint (not a warning) because PATH is environment-dependent. A binary missing on your laptop may exist in the pipeline runner's container.

---

### DIP126: Subgraph Ref File Not Found

**Severity**: Warning

A subgraph node references a file that does not exist at the declared path.

```
warning[DIP126]: subgraph node "Review" references "review_pipeline.dip" which does not exist
  --> pipeline.dip:28:5
  = help: resolved path: /home/user/project/review_pipeline.dip
```

**What triggers it**: A `subgraph` node declares `ref: path/to/workflow.dip` but the file cannot be found on disk. The path is resolved relative to the workflow source file's directory.

**How to fix**: Verify the file path is correct, or create the missing workflow file. Absolute paths are also supported.

**Note**: This check is skipped when source file context is unavailable (e.g., hand-constructed IR in tests) and in WASM environments where filesystem access is not available.

### DIP127: Invalid Human Node Mode

**Severity**: Warning

A human node has a `mode` value that is not one of the recognized modes.

```
warning[DIP127]: node "Gate" has mode "interactive" which is not a recognized human mode
  --> pipeline.dip:12:3
  = help: valid modes: choice, freeform, interview, yes_no
```

**What triggers it**: A human node declares a mode other than `choice`, `freeform`, `interview`, or `yes_no`.

**How to fix**: Change the mode to a valid value.

### DIP128: Interview Mode with Meaningless Default

**Severity**: Warning

A human node with `mode: interview` also sets a `default` value. Interview mode collects structured answers — it has no predefined choices to default to.

```
warning[DIP128]: node "Ask" is mode interview but has default "yes" which is ignored
  --> pipeline.dip:15:3
  = help: default is only meaningful for choice mode; remove it
```

**What triggers it**: `mode: interview` combined with a non-empty `default` field.

**How to fix**: Remove the `default` field, or change the mode to `choice` if you want label-based routing.

### DIP129: Interview Mode with Conflicting Choice-Style Edges

**Severity**: Warning

A human node with `mode: interview` has multiple labeled outgoing edges. Interview mode does not route by label — it collects answers and follows a single unconditional edge.

```
warning[DIP129]: node "Ask" is mode interview but has 2 labeled edges (interview does not route by label)
  --> pipeline.dip:15:3
  = help: interview mode collects answers, not choices; use mode choice for label-based routing
```

**What triggers it**: A `mode: interview` node has 2 or more outgoing edges with labels.

**How to fix**: Remove edge labels, or change the mode to `choice` if routing by selection is intended.

### DIP130: Invalid `response_format` Value

**Severity**: Warning

An agent node (or any node) specifies a `response_format` value that is not one of the recognized values.

```
warning[DIP130]: node "Analyze" has response_format "json" which is not a recognized value
  --> pipeline.dip:12:3
  = help: valid values: json_object, json_schema
```

**What triggers it**: A `response_format` field is set to anything other than `json_object` or `json_schema`. Also fires when `response_format` is used on non-agent nodes, where it has no effect.

**How to fix**: Use one of the two recognized values:
- `json_object` — Forces the LLM to return valid JSON (any shape).
- `json_schema` — Forces the LLM to return JSON matching the schema in `response_schema`.

---

### DIP131: `response_schema` / `response_format` Mismatch

**Severity**: Warning (or Hint)

There is a mismatch between `response_schema` and `response_format: json_schema`.

```
warning[DIP131]: node "Analyze" has response_schema but response_format is not json_schema
  --> pipeline.dip:12:3
  = help: set response_format: json_schema to enforce the schema

hint[DIP131]: node "Analyze" has response_format: json_schema but no response_schema
  --> pipeline.dip:12:3
  = help: add a response_schema block to define the expected JSON shape
```

**What triggers it**:
- `response_schema` is set but `response_format` is not `json_schema` — the schema will be ignored (warning).
- `response_format: json_schema` is set but `response_schema` is absent — schema enforcement has no definition (hint).

**How to fix**: Ensure both fields are set together when using schema-constrained output.

---

### DIP132: `response_schema` Must Be Valid JSON

**Severity**: Warning

The content of the `response_schema` block is not valid JSON.

```
warning[DIP132]: node "Analyze" has response_schema that is not valid JSON
  --> pipeline.dip:12:3
  = help: response_schema must be a valid JSON Schema object
```

**What triggers it**: The `response_schema` multiline block cannot be parsed as valid JSON.

**How to fix**: Ensure the schema block contains valid JSON. The content must be a JSON object (typically a JSON Schema definition).

---

### DIP133: Agent `params` Key Shadows First-Class Field

**Severity**: Hint

An agent node's `params` block contains a key that matches a first-class agent field (e.g., `model`, `provider`).

```
hint[DIP133]: node "Analyze" params key "model" shadows the first-class field model
  --> pipeline.dip:12:3
  = help: use the top-level model: field instead of params: model to avoid ambiguity
```

**What triggers it**: A key in the agent's `params` block has the same name as a recognized first-class field like `model`, `provider`, `prompt`, `system_prompt`, etc.

**How to fix**: Move the value to the appropriate first-class field, or rename the params key if the intent is to pass it through as a custom parameter.

---

### DIP134: max_retries Set With Restart Edges but No max_restarts

**Severity**: Warning

`max_retries` is set in defaults and the workflow has `restart: true` edges, but `max_restarts` is not set. These are commonly confused: `max_retries` controls per-node LLM retries, while `max_restarts` controls the global loop restart budget.

**What triggers it**: `defaults.max_retries` is set and at least one edge is marked `restart: true`, but `defaults.max_restarts` is absent.

**How to fix**: Set `max_restarts` in defaults to control loop iterations, or add it alongside `max_retries` if both behaviors are intended:
```dippin
  defaults
    max_retries: 3       # per-node LLM retry attempts
    max_restarts: 5      # global loop restart budget
```

---

### DIP135: manager_loop subgraph_ref Missing or File Does Not Exist

**Severity**: Warning

A `manager_loop` node either has no `subgraph_ref` field set, or the referenced file cannot be found on disk.

```text
warning[DIP135]: manager_loop node "Supervise" has no subgraph_ref or the file does not exist
  --> pipeline.dip:12:3
  = help: set subgraph_ref to the path of an existing .dip file that defines the child pipeline
```

**What triggers it**: A `manager_loop` node is declared without `subgraph_ref`, or the path in `subgraph_ref` does not resolve to an existing file (resolved relative to the workflow source file's directory).

**How to fix**: Set `subgraph_ref` to the path of an existing `.dip` file that defines the child pipeline:
```dippin
  manager_loop Supervise
    subgraph_ref: quality_loop.dip  # file must exist relative to this workflow
```

---

### DIP136: manager_loop Control Field Has Invalid Value

**Severity**: Warning

A `manager_loop` node has a `poll_interval` or `max_cycles` value that is negative.

```text
warning[DIP136]: manager_loop node "Supervise" poll_interval is negative
  --> pipeline.dip:12:3
  = help: use non-negative values for poll_interval and max_cycles
```

**What triggers it**: `poll_interval` or `max_cycles` is set to a negative value.

**How to fix**: Use non-negative values for both fields:
```dippin
  manager_loop Supervise
    subgraph_ref: inner.dip
    poll_interval: 30s    # non-negative duration
    max_cycles: 10        # non-negative integer
```

---

### DIP137: Unbounded manager_loop

**Severity**: Warning

A `manager_loop` node has neither a `stop_condition` nor a `max_cycles` cap, so supervision can run forever.

```text
warning[DIP137]: manager_loop node "Supervise" is unbounded: no stop_condition and no max_cycles
  --> pipeline.dip:12:3
  = help: set stop_condition (e.g., stack.child.outcome = success) or max_cycles to bound supervision
```

**What triggers it**: A `manager_loop` node declares neither `stop_condition` nor `max_cycles`.

**How to fix**: Set `stop_condition` or `max_cycles` (or both) to bound the supervision loop:
```dippin
  manager_loop Supervise
    subgraph_ref: inner.dip
    stop_condition: stack.child.outcome = success  # or: max_cycles: 20
```

---

### DIP139: Invalid tool_access Value on Agent Node or Parallel Branch

**Severity**: Warning

An agent node has `tool_access:` set to a value other than `none` (case-insensitive) or empty. The field is the v0.32.0 safety primitive that strips an LLM's tool catalog; v1 recognizes only one explicit value.

```text
warning[DIP139]: node "ReportFinalStatus" has tool_access "nono" which is not recognized
  --> pipeline.dip:12:3
  = help: use `tool_access: none` to disable LLM tools, or omit the field for the full catalog
```

**What triggers it**: An agent node declares `tool_access:` with a value other than `none` (case-insensitive) or empty. Invalid values fall back to no-tools at runtime (fail-closed) — the diagnostic surfaces the typo so author intent matches runtime behavior. Also fires on a per-branch override within a parallel node that sets `tool_access` to an unrecognized value.

**How to fix**: Use `tool_access: none` to disable LLM tools, or omit the field for the full catalog:
```dippin
  agent ReportFinalStatus
    prompt: "Summarize the results"
    tool_access: none    # explicit: no LLM tools
```

---

### DIP140: `params` Re-enables Tools That `tool_access` Strips

**Severity**: Warning

An agent node sets `tool_access` (any non-empty value) and also sets a `params` key that would re-grant tools: `allowed_tools`, `disallowed_tools`, `tool_choice`, or `permission_mode`. When `tool_access` is set, the runtime ignores these `params` keys (fail-closed), so the override is silently neutralized — a likely bypass attempt or dead config.

```text
warning[DIP140]: node "Summarize" sets tool_access but params key "allowed_tools" re-enables tools — tool_access wins (fail-closed)
  --> pipeline.dip:12:3
  = help: remove the params key; tool_access governs the tool catalog. To grant tools instead, omit tool_access.
```

**What triggers it**: An agent node has `tool_access` set (any non-empty value) and its `params` block contains any of: `allowed_tools`, `disallowed_tools`, `tool_choice`, or `permission_mode`.

**How to fix**: Remove the conflicting `params` key. If you want tools available, omit `tool_access` instead:
```dippin
  agent Summarize
    prompt: "Summarize"
    tool_access: none
    params:
      allowed_tools: Bash   # DIP140: the runtime strips this — tool_access wins
  # Fix: remove the params key, or remove tool_access to grant the tool
```

---

### DIP141: `writable_paths` Nullified by `tool_access: none`

**Severity**: Warning

An agent node or per-branch override sets `writable_paths` together with `tool_access: none` on the same object. `tool_access: none` strips the entire tool catalog, so there is no Write/Edit/Bash left for `writable_paths` to bound — the field is dead config.

```text
warning[DIP141]: node "Summarize" has writable_paths but tool_access "none" — none strips all tools, so there is nothing to bound (dead config)
  --> pipeline.dip:12:3
  = help: remove writable_paths (no tools to bound) or drop tool_access: none to grant a bounded tool catalog.
```

**What triggers it**: An agent node or a per-branch override declares both a non-empty `writable_paths` and `tool_access: none` (case-insensitive) on the same object. Only the *same config object* declaring both is flagged — a branch that sets `tool_access: none` while *inheriting* an agent's `writable_paths` is legitimate narrowing and is not flagged.

**How to fix**: Remove `writable_paths` (there are no tools to bound) or drop `tool_access: none` to grant a bounded tool catalog:

```dippin
  agent Summarize
    prompt: "Summarize the results"
    tool_access: none
    writable_paths: workspace/**   # DIP141: none strips all tools — nothing to bound
```

---

### DIP142: Unsafe `writable_paths` Entry

**Severity**: Warning

A `writable_paths` entry is an absolute path, starts with `~` or a Windows drive, escapes its base via `..`, or is a brace-expansion fragment torn apart by comma-splitting. Such an entry will not bound writes to the workspace the way the author expects.

```text
warning[DIP142]: node "Recorder" writable_paths entry "/etc/**" escapes the workspace (absolute / ~ / parent path) — the runtime write-jail will not honor it
  --> pipeline.dip:12:3
  = help: use workspace-relative globs (e.g. .ai/sprints/**). Absolute, ~, and ..-escaping entries are rejected by the fs jail (it bounds writes to the session root); this lint catches obvious lexical cases only — the runtime jail is the real boundary. See #67/#77.
```

**What triggers it**: A `writable_paths` entry matches any of:
- **Absolute** — leading `/`, `~`, `\`, or Windows drive letter (`C:\`)
- **Parent escape** — contains a `..` segment that escapes the base (e.g. `../../etc/**`, `foo/../../bar`)
- **Brace mis-split** — unbalanced `{` or `}` (a glob like `*.{md,yaml}` is torn apart by comma-splitting into `*.{md` and `yaml}` — enumerate entries instead)

This is a lexical clarity check, not the security boundary — the runtime fs-jail is authoritative.

**How to fix**: Use workspace-relative globs. Enumerate brace-expansion alternatives:

```dippin
  agent Recorder
    prompt: "record"
    writable_paths: /etc/**   # DIP142: absolute path escapes the workspace jail
    # Fix: writable_paths: workspace/output.md
```

---

### DIP143: Referenced Subgraph Does Not Inherit `tool_access`

**Severity**: Hint

A `manager_loop` (`subgraph_ref`) or `subgraph` (`ref`) node references a child `.dip` file, and this workflow declares `tool_access` on at least one agent or parallel branch. `tool_access` is a per-node primitive and does not cross a file boundary — the child subgraph's agents are governed entirely by their own file, so the parent's restriction does not propagate into it.

```text
hint[DIP143]: manager_loop "Supervise" references subgraph "child.dip", defined in its own file; this workflow's tool_access restrictions do not extend across the subgraph boundary
  --> pipeline.dip:9:3
  = help: audit the agents in "child.dip" for their own tool_access — restrictions declared in this workflow do not propagate into a referenced subgraph. Cross-file enforcement is tracked as #89.
```

**What triggers it**: All of the following hold:
- A node references an external subgraph — `manager_loop` via `subgraph_ref`, or `subgraph` via `ref` (non-empty).
- The workflow declares `tool_access` (any non-empty value — a fail-closed typo still expresses restriction) on some agent or parallel branch.

A node whose ref resolves to its **own source file** (a direct self-reference) is not flagged — there is no cross-file boundary. Transitive cross-file cycles (A → B → A) are not detected and are deferred to [#89](https://github.com/2389-research/dippin-lang/issues/89).

This is a Hint, not a Warning: the referencing node has no defect. The check is **file-level** — it does not verify that the restricted node and the subgraph node are related, and it does **not** read the child file (the validator cannot parse it). It bounds the child's *tool catalog* concern, not information flow across the supervisory boundary (`steer_context` / `stack.child.*` — see [#56](https://github.com/2389-research/dippin-lang/issues/56)). Real cross-file effective-access enforcement is deferred to [#89](https://github.com/2389-research/dippin-lang/issues/89).

**How to fix**: Open the referenced `.dip` and give its agents their own `tool_access`:

```dippin
  agent Summarize
    tool_access: none
  manager_loop Supervise
    subgraph_ref: child.dip   # DIP143: child.dip's agents must set their own tool_access
```

---

### DIP144: Agent Node Has No Failure Route

**Severity**: Warning

An agent node has no declared failure route at any level. If the node fails at runtime, the pipeline has nowhere to go and will halt.

```
warning[DIP144]: agent node "Build" has no failure route (no fail edge, no fallback_target, no bounded retry, no graph on_failure)
  --> pipeline.dip:12:3
  = help: add `-> <node> when ctx.outcome = fail`, set fallback_target:, add retry_target with max_retries, or declare a workflow-level on_failure:
```

This warning is especially important for agents with a bounded `max_turns`: turn
exhaustion ends the node as `fail`, so without a failure route the run halts on
exhaustion.

**What triggers it**: An `agent` node that has none of:
- An outgoing edge with `when ctx.outcome = fail` (or `failure`)
- A `fallback_target` field
- A `retry_target` + `max_retries` bounded retry
- A `defaults.on_failure` declared on the workflow

**Does not fire on**: `human`, `tool`, `parallel`, `fan_in`, `conditional`, `subgraph`, or `manager_loop` nodes.

**How to fix**: Add any one of the four suppression routes:

```dippin
  # Option 1: explicit fail edge
  edges
    Build -> Escalate when ctx.outcome = fail

  # Option 2: node fallback_target
  agent Build
    fallback_target: Escalate

  # Option 3: bounded retry
  agent Build
    retry_target: Build
    max_retries: 3
    fallback_target: Escalate

  # Option 4: graph-level catch-all
  defaults
    on_failure: Escalate
```

See [edges.md](edges.md) — Failure Handling for the full five-level precedence cascade.

---

### DIP145: Negative Graph Budget Default

A workflow budget default is set to a negative value. Budgets are non-negative;
`0` (or unset) means **no limit**.

```text
warning[DIP145]: workflow budget default max_cost_cents is -5; budgets cannot be negative
```

**Trigger:** `max_total_tokens`, `max_cost_cents`, `max_wall_time`, or
`stall_timeout` in the `defaults` block is negative.

**Fix:** Use a positive cap, or omit the field / set `0` for no limit. Note `0`
means *unlimited*, not "zero budget."

---

### DIP146: Child Subgraph Re-Grants Restricted Tools (cross-file)

`dippin lint` resolves a `manager_loop` (`subgraph_ref`) or `subgraph` (`ref`)
child across the file boundary and finds it declares **no** `tool_access`
restriction on any agent, while a workflow on the path from the linted entry
restricts tools. Unlike DIP143 (which cannot read the child), DIP146 reads and
confirms the child, and traverses transitively.

```text
hint[DIP146]: manager_loop "Supervise" delegates to subgraph "worker.dip", which declares no tool_access restriction on any agent; a workflow on this path restricts tools, but the restriction does not cross the subgraph boundary
```

**Trigger:** A workflow on the path declares `tool_access` on some agent/branch,
and a resolved child restricts none of its agents. A child that restricts every
agent is silent; one that restricts some but leaves a tool-bearing agent open
keeps the DIP143 advisory instead. Emitted by the CLI only (native), not by the
wasm/playground linter.

**Fix:** Give the child's agents their own `tool_access` (e.g. `tool_access:
none`). Restrictions in a parent do not flow into a referenced subgraph.

**What DIP146 does NOT check:** DIP146 *does* traverse transitively — a
zero-intent grandchild is flagged. The limitation is narrower: a **partial-audit
or unparseable** child deeper than the entry's direct boundaries gets no
*additional* advisory, because the DIP143 fallback is only emitted for the linted
file itself. Also out of scope:
information flow across the supervisory boundary (`steer_context` /
`stack.child.*` — see [#56](https://github.com/2389-research/dippin-lang/issues/56));
runtime enforcement (the tracker runtime enforces; dippin detects). A clean
result means every resolvable child is fully locked down or had no restriction to
escape — not that the child's tools are restricted at runtime.

---

## Running Validation

### Structural validation only

```bash
dippin validate pipeline.dip
```

Runs DIP001–DIP009. Exit code 0 if all pass, 1 if any errors.

### Full lint (validation + semantic)

```bash
dippin lint pipeline.dip
```

Runs all DIP001–DIP009 errors and DIP101–DIP146 warnings. Exit code 1 only for errors; warnings alone exit 0.

### JSON output for CI

```bash
dippin --format json lint pipeline.dip
```

Emits diagnostics as a JSON array for machine consumption.
