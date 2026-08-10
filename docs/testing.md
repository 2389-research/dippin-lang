# Scenario Testing

The `dippin test` command runs scenario-based assertions against workflow files. Test cases inject context values into the simulator and verify the execution path.

```mermaid
flowchart LR
    A[".test.json"] --> B["Load & Parse"]
    B --> C["Inject Scenario"]
    C --> D["Simulate Workflow"]
    D --> E["Check Assertions"]
    E --> F{"Pass?"}
    F -->|Yes| G["PASS"]
    F -->|No| H["FAIL + errors"]
```

---

## Test File Format

Test files use `.test.json` extension and are auto-discovered from the workflow path:

```text
pipeline.dip       → pipeline.test.json
src/flow.dip       → src/flow.test.json
```

### Schema

```json
{
  "tests": [
    {
      "name": "descriptive test name",
      "scenario": {
        "key": "value"
      },
      "expect": {
        "status": "success",
        "visited": ["NodeA", "NodeB"],
        "not_visited": ["NodeC"],
        "path_contains": ["NodeA", "NodeB"],
        "immediately_after": {"NodeA": "NodeB"}
      }
    }
  ]
}
```

### Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `tests` | array | yes | List of test cases |
| `tests[].name` | string | yes | Human-readable test name (shown in output) |
| `tests[].scenario` | object | no | Context key/value pairs injected into the simulator. These determine which conditional edges are taken. |
| `tests[].branch` | string[] | no | Restricts which parallel-branch targets the simulator walks. When set, only the listed branch targets are explored; omit to walk all branches. |
| `tests[].expect` | object | yes | Assertions to check against the simulation result |

### Expectation Fields

All expectation fields are optional. Only specified fields are checked.

| Field | Type | Description |
|-------|------|-------------|
| `status` | string | Expected simulation status: `"success"` (reached exit) or `"dead_end"` |
| `visited` | string[] | Node IDs that must appear in the execution path |
| `not_visited` | string[] | Node IDs that must NOT appear in the execution path |
| `path_contains` | string[] | Node IDs that must appear **in order** in the execution path. Non-contiguous matches are allowed (other nodes may appear between them). |
| `immediately_after` | object | Map of `{"NodeA": "NodeB"}` pairs asserting that NodeB is the immediate next node visited after NodeA. Useful for verifying specific edge routing. |

### Caveats

**`not_visited` and loop breaking**: The test runner limits per-node visits to 3. When a loop exceeds this limit, the simulator forces the loop-exit edge and **continues execution** rather than stopping. This means nodes downstream of the loop-exit can be visited even though the loop was broken. For edge-routing assertions in workflows with loops, prefer `path_contains` (which checks ordering) over `not_visited` (which can be fragile when the simulator continues past a forced loop exit).

**`immediately_after` for edge routing**: When testing which specific edge a node takes, `immediately_after` is more precise than `path_contains`. Use it to verify that a conditional edge routes to the expected next node.

### Clearing Defaults

Tool nodes auto-seed `ctx.tool_stdout` and `ctx.outcome` to `"success"`. To test unconditional fallback edges after a tool node, set the key to an empty string to suppress the default:

```json
{
  "scenario": {"ToolNode.tool_stdout": ""},
  "expect": {"visited": ["FallbackNode"]}
}
```

This prevents the auto-default from firing, so no conditional edge matches and the unconditional fallback edge is taken.

---

## Example

Given a workflow `gate.dip`:

```dippin
workflow Gate
  goal: "Route based on outcome"
  start: Start
  exit: Exit

  agent Start
    label: Start

  agent Exit
    label: Exit

  agent Pass
    label: Pass
    model: claude-sonnet-4-6
    provider: anthropic
    prompt: Handle success.

  agent Fix
    label: Fix
    model: claude-sonnet-4-6
    provider: anthropic
    prompt: Handle failure.

  edges
    Start -> Pass  when ctx.outcome = success
    Start -> Fix  when ctx.outcome = fail
    Pass -> Exit
    Fix -> Exit
```

Create `gate.test.json`:

```json
{
  "tests": [
    {
      "name": "success path",
      "scenario": {"outcome": "success"},
      "expect": {
        "status": "success",
        "visited": ["Start", "Pass", "Exit"],
        "not_visited": ["Fix"]
      }
    },
    {
      "name": "failure path",
      "scenario": {"outcome": "fail"},
      "expect": {
        "status": "success",
        "visited": ["Start", "Fix", "Exit"],
        "not_visited": ["Pass"],
        "path_contains": ["Start", "Fix", "Exit"],
        "immediately_after": {"Start": "Fix"}
      }
    }
  ]
}
```

Run:

```bash
$ dippin test gate.dip
═══ Test Results ═════════════════════════════════════════════
  PASS  success path
  PASS  failure path
─── Summary ───────────────────────────────────────────────────
  2 tests: 2 passed, 0 failed

$ dippin test --verbose gate.dip
═══ Test Results ═════════════════════════════════════════════
  PASS  success path
        path: Start → Pass → Exit
  PASS  failure path
        path: Start → Fix → Exit
─── Summary ───────────────────────────────────────────────────
  2 tests: 2 passed, 0 failed
```

---

## Scenario Keys

The `scenario` object maps context keys to values. These are injected into the simulator's context before execution begins. The simulator uses these values to resolve conditional edges.

Common scenario keys:

| Key | Description |
|-----|-------------|
| `outcome` | Maps to `ctx.outcome` — the most common routing variable |
| `status` | Maps to `ctx.status` |
| `tool_stdout` | Maps to `ctx.tool_stdout` — tool command output |

The simulator resolves conditions by looking up `ctx.<key>` in the scenario context. For example, `scenario: {"outcome": "fail"}` causes `when ctx.outcome = fail` edges to match.

---

## Edge Coverage

Pass `--coverage` to report which workflow edges were exercised across all test scenarios:

```bash
$ dippin test --coverage gate.dip
═══ Test Results ═════════════════════════════════════════════
  PASS  success path
  PASS  failure path
─── Summary ───────────────────────────────────────────────────
  2 tests: 2 passed, 0 failed

─── Edge Coverage ─────────────────────────────────────────────
  4/4 edges covered (100.0%)
```

This is **edge** coverage only — there is no node-coverage report. Any uncovered edges are listed beneath the summary so you can add scenarios that exercise them.

---

## CI Integration

Use `--format json` for machine-readable output:

```bash
dippin --format json test pipeline.dip
```

```json
{
  "results": [
    {"name": "happy path", "passed": true, "path": ["Start", "Gate", "Pass", "Exit"]},
    {"name": "error path", "passed": false, "errors": ["expected status \"dead_end\", got \"success\""]}
  ],
  "passed": 1,
  "failed": 1,
  "total": 2
}
```

Exit code is 0 if all tests pass, 1 if any fail.
