# on_failure route (#92) + DIP144 lint (#93) Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a graph-level `on_failure: <NodeID>` default failure route (carried, validated, formatted, DOT-round-tripped) and a new DIP144 lint warning for agent nodes that have no failure route.

**Architecture:** `on_failure` lands in `ir.WorkflowDefaults.OnFailure`, parsed in the `defaults:` block, validated structurally (reusing DIP003), seeded into DIP004 reachability so catch-all-only recovery nodes aren't falsely flagged, and round-tripped through both the `.dip` formatter and DOT export/migrate. DIP144 flags agent nodes lacking a failure route, suppressed by an explicit fail-outcome edge, `fallback_target`, bounded retry, or graph `on_failure`. dippin carries/validates; the tracker runtime enforces (not gated on tracker readiness).

**Tech Stack:** Go; `just` for all build/test (never raw `go`); pre-commit hook mirrors CI (cyclomatic ≤5 / cognitive ≤7).

**Spec:** `docs/superpowers/specs/2026-06-04-issue-92-93-design.md`

**Worktree (use ABSOLUTE paths; cwd may reset between commands):**
`/home/clint/code/2389/dippin-lang/.claude/worktrees/feat+92-on-failure-route`

**Build order constraint:** `on_failure` (Tasks 1–5) must land before DIP144 (Task 6) so suppressor (d) compiles.

**Test-run convention:** `just test-pkg <pkg>` runs the whole package verbose — find your test name in the output. Use `just test` for the full suite and `just complexity` before each commit-worthy step. Do NOT run raw `go test`.

---

### Task 1: Parse `on_failure` into `WorkflowDefaults.OnFailure`

**Files:**
- Modify: `ir/ir.go` (WorkflowDefaults struct, ~line 47)
- Modify: `parser/parse_defaults.go:77-91` (`applyDefaultExtraField`)
- Modify: `parser/testdata/defaults_complex.dip`
- Test: `parser/parser_test.go` (`TestParseDefaultsComplex`, ~line 859)

- [ ] **Step 1: Add the failing test assertion**

In `parser/testdata/defaults_complex.dip`, add one line inside the `defaults` block (after `restart_target: A`):

```
    on_failure: A
```

In `parser/parser_test.go`, inside `TestParseDefaultsComplex`, add after the `RestartTarget` assertion:

```go
	if d.OnFailure != "A" {
		t.Errorf("on_failure = %q, want A", d.OnFailure)
	}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `just test-pkg parser`
Expected: compile error — `d.OnFailure undefined (type ir.WorkflowDefaults has no field or method OnFailure)`.

- [ ] **Step 3: Add the IR field**

In `ir/ir.go`, in `WorkflowDefaults`, directly after the `RestartTarget` line:

```go
	RestartTarget     string        // Where to restart on loop
	OnFailure         string        // Graph-level default failure route (node ID); runtime catch-all when no node-level fail route applies
```

- [ ] **Step 4: Add the parser case**

In `parser/parse_defaults.go`, in `applyDefaultExtraField`, add a case to the switch:

```go
	case "on_resume":
		d.OnResume = val
	case "on_failure":
		d.OnFailure = val
```

- [ ] **Step 5: Run test to verify it passes**

Run: `just test-pkg parser`
Expected: PASS (TestParseDefaultsComplex green).

- [ ] **Step 6: Commit**

```bash
cd /home/clint/code/2389/dippin-lang/.claude/worktrees/feat+92-on-failure-route
git add ir/ir.go parser/parse_defaults.go parser/testdata/defaults_complex.dip parser/parser_test.go
git commit -m "feat: parse graph-level on_failure into WorkflowDefaults (#92)

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

### Task 2: Structural validation of `on_failure` (reuse DIP003)

**Files:**
- Modify: `validator/validate.go` (`Validate`, ~line 14; add `checkOnFailureTarget`)
- Modify: `validator/codes.go:7` (DIP003 comment)
- Test: `validator/validate_test.go`

- [ ] **Step 1: Write the failing tests**

Add to `validator/validate_test.go`:

```go
func TestOnFailureUnknownNode(t *testing.T) {
	w := &ir.Workflow{
		Name: "t", Start: "A", Exit: "A",
		Defaults: ir.WorkflowDefaults{OnFailure: "Nope"},
		Nodes:    []*ir.Node{{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "go"}}},
	}
	res := Validate(w)
	if !hasCode(res.Diagnostics, DIP003) {
		t.Fatalf("expected DIP003 for unknown on_failure target, got %+v", res.Diagnostics)
	}
}

func TestOnFailureKnownNode(t *testing.T) {
	w := &ir.Workflow{
		Name: "t", Start: "A", Exit: "A",
		Defaults: ir.WorkflowDefaults{OnFailure: "A"},
		Nodes:    []*ir.Node{{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "go"}}},
	}
	res := Validate(w)
	for _, d := range res.Diagnostics {
		if d.Code == DIP003 {
			t.Fatalf("unexpected DIP003 for valid on_failure target: %+v", d)
		}
	}
}
```

If `hasCode` does not already exist in the test package, add this helper to `validator/validate_test.go`:

```go
func hasCode(diags []Diagnostic, code string) bool {
	for _, d := range diags {
		if d.Code == code {
			return true
		}
	}
	return false
}
```

(First grep `validator/*_test.go` for an existing `hasCode`/`hasDiagnostic` helper and reuse it instead of redefining — redefinition is a compile error.)

- [ ] **Step 2: Run tests to verify they fail**

Run: `just test-pkg validator`
Expected: `TestOnFailureUnknownNode` FAILS ("expected DIP003 ... got []").

- [ ] **Step 3: Implement the check**

In `validator/validate.go`, add the call inside `Validate` after `checkEdgeEndpoints(w)`:

```go
	diags = append(diags, checkEdgeEndpoints(w)...)
	diags = append(diags, checkOnFailureTarget(w)...)
```

Add the function (place near `checkEdgeEndpoints`):

```go
// checkOnFailureTarget verifies DIP003: the graph-level on_failure route, if set,
// references an existing node. on_failure is a node reference like an edge endpoint,
// so it reuses the DIP003 code.
func checkOnFailureTarget(w *ir.Workflow) []Diagnostic {
	target := w.Defaults.OnFailure
	if target == "" || w.Node(target) != nil {
		return nil
	}
	d := Diagnostic{
		Code:     DIP003,
		Severity: SeverityError,
		Message:  fmt.Sprintf("on_failure references unknown node %q", target),
	}
	if suggestion := closestNodeID(w, target); suggestion != "" {
		d.Help = fmt.Sprintf("did you mean %q?", suggestion)
	} else {
		d.Help = fmt.Sprintf("declare a node with ID %q or fix the on_failure target", target)
	}
	return []Diagnostic{d}
}
```

- [ ] **Step 4: Update the DIP003 comment for honesty**

In `validator/codes.go`, change line 7:

```go
	DIP003 = "DIP003" // unknown node reference (edge endpoint or on_failure target)
```

- [ ] **Step 5: Run tests to verify they pass**

Run: `just test-pkg validator`
Expected: both new tests PASS.

- [ ] **Step 6: Check complexity & commit**

```bash
cd /home/clint/code/2389/dippin-lang/.claude/worktrees/feat+92-on-failure-route
just complexity
git add validator/validate.go validator/codes.go validator/validate_test.go
git commit -m "feat: validate on_failure references an existing node via DIP003 (#92)

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

### Task 3: DIP004 reachability — seed the `on_failure` target

**Files:**
- Modify: `validator/validate.go` (`buildAllEdgeAdjacency`, ~line 111)
- Test: `validator/validate_test.go`

- [ ] **Step 1: Write the failing test**

A node reachable ONLY via `on_failure` must not be flagged DIP004:

```go
func TestOnFailureTargetIsReachable(t *testing.T) {
	w := &ir.Workflow{
		Name: "t", Start: "A", Exit: "Done",
		Defaults: ir.WorkflowDefaults{OnFailure: "Rescue"},
		Nodes: []*ir.Node{
			{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "go"}},
			{ID: "Rescue", Kind: ir.NodeHuman, Config: ir.HumanConfig{Mode: "freeform"}},
			{ID: "Done", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "end"}},
		},
		Edges: []*ir.Edge{
			{From: "A", To: "Done"},
			{From: "Rescue", To: "Done"},
		},
	}
	res := Validate(w)
	for _, d := range res.Diagnostics {
		if d.Code == DIP004 {
			t.Fatalf("Rescue should be reachable via on_failure, got DIP004: %+v", d)
		}
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `just test-pkg validator`
Expected: FAIL — `Rescue should be reachable via on_failure, got DIP004` (Rescue has no incoming edge).

- [ ] **Step 3: Seed the on_failure edge into reachability**

In `validator/validate.go`, modify `buildAllEdgeAdjacency`:

```go
// buildAllEdgeAdjacency builds an adjacency map including all edges (including restart).
// It also includes implicit parallel/fan_in edges and the graph-level on_failure route,
// so a recovery node reachable only via on_failure is not falsely flagged DIP004.
func buildAllEdgeAdjacency(w *ir.Workflow) map[string][]string {
	adj := make(map[string][]string)
	for _, e := range w.Edges {
		adj[e.From] = append(adj[e.From], e.To)
	}
	addParallelFanInEdges(adj, w)
	addOnFailureEdge(adj, w)
	return adj
}

// addOnFailureEdge adds the graph-level on_failure target as reachable from start.
// One path suffices for reachability; on_failure is conceptually reachable from any
// failing node, and start is always reachable.
func addOnFailureEdge(adj map[string][]string, w *ir.Workflow) {
	if w.Defaults.OnFailure != "" {
		adj[w.Start] = append(adj[w.Start], w.Defaults.OnFailure)
	}
}
```

Note: `buildAllEdgeAdjacency` feeds only DIP004 reachability — NOT `buildNonRestartAdjacency` (DIP005 cycles) or `buildForwardAdjacency` (DIP105 success path), so this adds no false cycle and no false success path.

- [ ] **Step 4: Run test to verify it passes**

Run: `just test-pkg validator`
Expected: PASS. Also confirm no DIP005/DIP105 regressions in the package output.

- [ ] **Step 5: Commit**

```bash
cd /home/clint/code/2389/dippin-lang/.claude/worktrees/feat+92-on-failure-route
just complexity
git add validator/validate.go validator/validate_test.go
git commit -m "fix: seed on_failure target into DIP004 reachability (#92)

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

### Task 4: Formatter emit + `.dip` round-trip

**Files:**
- Modify: `formatter/format.go:164-173` (`writeDefaultsRestartFields`)
- Test: `formatter/format_test.go`
- Test: `parser/parser_test.go`

- [ ] **Step 1: Write the failing formatter test**

Add to `formatter/format_test.go` (mirror `TestFormatDefaultsRestartTarget`):

```go
func TestFormatDefaultsOnFailure(t *testing.T) {
	w := &ir.Workflow{
		Name:  "test",
		Start: "A",
		Exit:  "A",
		Defaults: ir.WorkflowDefaults{
			OnFailure: "Escalate",
		},
		Nodes: []*ir.Node{
			{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "go."}},
		},
	}
	output := Format(w)
	assertContains(t, output, "    on_failure: Escalate")
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `just test-pkg formatter`
Expected: FAIL — output does not contain `on_failure: Escalate`.

- [ ] **Step 3: Emit on_failure in the formatter**

In `formatter/format.go`, in `writeDefaultsRestartFields`, add after the `restart_target` block:

```go
	if d.RestartTarget != "" {
		wr.line("restart_target: %s", d.RestartTarget)
	}
	if d.OnFailure != "" {
		wr.line("on_failure: %s", d.OnFailure)
	}
	writeDefaultsCompactionFields(wr, d)
```

- [ ] **Step 4: Run formatter test to verify it passes**

Run: `just test-pkg formatter`
Expected: PASS. Run `just complexity` — `writeDefaultsRestartFields` now has 3 `if`s (cyclomatic 4), within budget.

- [ ] **Step 5: Add the `.dip` round-trip test**

Add to `parser/parser_test.go` (mirror `TestParseDefaultsBudgetRoundTrip`):

```go
func TestParseDefaultsOnFailureRoundTrip(t *testing.T) {
	w1 := parseFixture(t, "defaults_complex.dip")
	if w1.Defaults.OnFailure != "A" {
		t.Fatalf("precondition: on_failure = %q, want A", w1.Defaults.OnFailure)
	}
	formatted := formatter.Format(w1)
	w2, err := NewParser(formatted, "roundtrip").Parse()
	if err != nil {
		t.Fatalf("re-parse error: %v\nformatted:\n%s", err, formatted)
	}
	if w2.Defaults.OnFailure != "A" {
		t.Errorf("round-trip: on_failure = %q, want A", w2.Defaults.OnFailure)
	}
}
```

(`parseFixture` and the `formatter` import already exist in this test file — confirm via the existing `TestParseDefaultsBudgetRoundTrip`.)

- [ ] **Step 6: Run round-trip test to verify it passes**

Run: `just test-pkg parser`
Expected: PASS.

- [ ] **Step 7: Commit**

```bash
cd /home/clint/code/2389/dippin-lang/.claude/worktrees/feat+92-on-failure-route
git add formatter/format.go formatter/format_test.go parser/parser_test.go
git commit -m "feat: format on_failure default + .dip round-trip (#92)

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

### Task 5: DOT export + migrate round-trip

**Files:**
- Modify: `export/dot.go` (`buildGraphAttrs` ~line 88, `reservedGraphAttrs` ~line 61)
- Modify: `migrate/migrate.go:114-123` (`graphDefaultsHandlers`)
- Test: `export/dot_test.go`
- Test: `migrate/roundtrip_test.go`

- [ ] **Step 1: Write the failing round-trip test**

Add to `migrate/roundtrip_test.go` (mirror `TestRoundTripToolSafetyDefaults`):

```go
// TestRoundTripOnFailureDefault verifies on_failure round-trips losslessly
// through parse → ExportDOT → Migrate.
func TestRoundTripOnFailureDefault(t *testing.T) {
	src := `workflow OnFailure
  goal: "round trip"
  start: A
  exit: A

  defaults
    on_failure: A

  agent A
    prompt: "Do it."

  edges
    A -> A
`
	w1, err := dipparser.NewParser(src, "rt.dip").Parse()
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	dot := export.ExportDOT(w1, export.ExportOptions{})
	w2, err := Migrate(dot)
	if err != nil {
		t.Fatalf("migrate: %v\nDOT:\n%s", err, dot)
	}
	if w2.Defaults.OnFailure != "A" {
		t.Errorf("on_failure after round-trip = %q, want %q; DOT:\n%s",
			w2.Defaults.OnFailure, "A", dot)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `just test-pkg migrate`
Expected: FAIL — `on_failure after round-trip = "" , want "A"` (DOT never emitted it).

- [ ] **Step 3: Emit on_failure in DOT export + reserve the attr name**

In `export/dot.go`, add to `reservedGraphAttrs` (append to the tool-safety line):

```go
	"tool_commands_allow": true, "tool_denylist_add": true,
	"on_failure": true,
```

In `buildGraphAttrs`, add after the tool-safety attribute block (before `addGraphVarsAttrs`):

```go
	if w.Defaults.OnFailure != "" {
		attrs = append(attrs, fmt.Sprintf("on_failure=%s", dotQuote(w.Defaults.OnFailure)))
	}
```

- [ ] **Step 4: Map on_failure back on migrate**

In `migrate/migrate.go`, add an entry to `graphDefaultsHandlers`:

```go
	"tool_denylist_add":   func(v string, w *ir.Workflow) { w.Defaults.ToolDenylistAdd = v },
	"on_failure":          func(v string, w *ir.Workflow) { w.Defaults.OnFailure = v },
```

- [ ] **Step 5: Run round-trip test to verify it passes**

Run: `just test-pkg migrate`
Expected: PASS.

- [ ] **Step 6: Add the DOT export assertion**

Add to `export/dot_test.go` (mirror `TestExportToolSafetyDefaults`):

```go
func TestExportOnFailureDefault(t *testing.T) {
	w := &ir.Workflow{
		Name: "t", Start: "A", Exit: "A",
		Defaults: ir.WorkflowDefaults{OnFailure: "Escalate"},
		Nodes:    []*ir.Node{{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "go"}}},
	}
	out := ExportDOT(w, ExportOptions{})
	if !strings.Contains(out, `on_failure="Escalate"`) {
		t.Errorf("expected on_failure graph attr, got:\n%s", out)
	}
}
```

- [ ] **Step 7: Run export test, complexity, commit**

Run: `just test-pkg export` → PASS.

```bash
cd /home/clint/code/2389/dippin-lang/.claude/worktrees/feat+92-on-failure-route
just complexity
git add export/dot.go export/dot_test.go migrate/migrate.go migrate/roundtrip_test.go
git commit -m "feat: round-trip on_failure through DOT export/migrate (#92)

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

### Task 6: DIP144 lint — agent node without a failure route (#93)

**Files:**
- Create: `validator/lint_failure_route.go`
- Create: `validator/lint_failure_route_test.go`
- Modify: `validator/lint.go` (append chain ~line 62; doc comment line 8)
- Modify: `validator/lint_codes.go` (const block ~line 47; CodeDescription ~line 94; header comment line 3)
- Modify: `validator/explanations.go` (`reachabilityExplanations`, ~line 105)

- [ ] **Step 1: Write the failing lint tests**

Create `validator/lint_failure_route_test.go`:

```go
package validator

import (
	"testing"

	"github.com/2389-research/dippin-lang/parser"
)

func lintSrc(t *testing.T, src string) []Diagnostic {
	t.Helper()
	w, err := parser.NewParser(src, "t.dip").Parse()
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	return Lint(w).Diagnostics
}

func TestDIP144FiresOnRoutelessAgent(t *testing.T) {
	src := `workflow W
  start: A
  exit: Done
  agent A
    prompt: "go"
  agent Done
    prompt: "end"
  edges
    A -> Done
`
	if !hasCode(lintSrc(t, src), DIP144) {
		t.Fatal("expected DIP144 on routeless agent A")
	}
}

func TestDIP144SuppressedByFailEdge(t *testing.T) {
	src := `workflow W
  start: A
  exit: Done
  agent A
    prompt: "go"
  human Rescue
    mode: freeform
  agent Done
    prompt: "end"
  edges
    A -> Done when ctx.outcome = success
    A -> Rescue when ctx.outcome = fail
    Rescue -> Done
`
	for _, d := range lintSrc(t, src) {
		if d.Code == DIP144 && d.Message != "" && containsNode(d.Message, "A") {
			t.Fatalf("DIP144 should be suppressed by fail edge on A: %+v", d)
		}
	}
}

func TestDIP144SuppressedByFallbackTarget(t *testing.T) {
	src := `workflow W
  start: A
  exit: Done
  agent A
    prompt: "go"
    fallback_target: Done
  agent Done
    prompt: "end"
  edges
    A -> Done
`
	if hasCode(lintSrc(t, src), DIP144) {
		t.Fatal("DIP144 should be suppressed by fallback_target")
	}
}

func TestDIP144SuppressedByGraphOnFailure(t *testing.T) {
	src := `workflow W
  start: A
  exit: Done
  defaults
    on_failure: Rescue
  agent A
    prompt: "go"
  human Rescue
    mode: freeform
  agent Done
    prompt: "end"
  edges
    A -> Done
    Rescue -> Done
`
	if hasCode(lintSrc(t, src), DIP144) {
		t.Fatal("DIP144 should be suppressed by graph on_failure")
	}
}

func TestDIP144FiresWithDIP115OnRoutelessGoalGate(t *testing.T) {
	src := `workflow W
  start: G
  exit: Done
  agent G
    prompt: "gate"
    goal_gate: true
  agent Done
    prompt: "end"
  edges
    G -> Done
`
	diags := lintSrc(t, src)
	if !hasCode(diags, DIP144) {
		t.Fatal("expected DIP144 on routeless goal_gate agent")
	}
	if !hasCode(diags, DIP115) {
		t.Fatal("expected DIP115 to still fire alongside DIP144 (independent codes)")
	}
}

func TestDIP144NotOnHumanOrToolNodes(t *testing.T) {
	src := `workflow W
  start: H
  exit: Done
  human H
    mode: freeform
  agent Done
    prompt: "end"
    fallback_target: Done
  edges
    H -> Done
`
	for _, d := range lintSrc(t, src) {
		if d.Code == DIP144 && containsNode(d.Message, "H") {
			t.Fatalf("DIP144 must not fire on human node H: %+v", d)
		}
	}
}
```

Add a tiny helper at the bottom of the new test file (or reuse an existing substring helper if present):

```go
func containsNode(msg, id string) bool {
	return msg != "" && (msg == id || stringContains(msg, "\""+id+"\""))
}

func stringContains(s, sub string) bool { return strings.Contains(s, sub) }
```

…and add `"strings"` to the test imports. (If a `strings`-based substring helper already exists in the package's tests, reuse it and drop these.)

- [ ] **Step 2: Run tests to verify they fail**

Run: `just test-pkg validator`
Expected: compile error — `undefined: DIP144` and `undefined: lintAgentFailureRoute` references will surface once added; first failure is `undefined: DIP144`.

- [ ] **Step 3: Add the DIP144 constant + description**

In `validator/lint_codes.go`, in the const block after `DIP143`:

```go
	DIP143 = "DIP143" // referenced subgraph does not inherit this workflow's tool_access restrictions
	DIP144 = "DIP144" // agent node has no failure route
```

And in the `CodeDescription` init after the `DIP143` line:

```go
	CodeDescription[DIP144] = "agent node has no failure route"
```

Update the header comment (line 3) from `DIP101–DIP143` to `DIP101–DIP144`.

- [ ] **Step 4: Add the Explanation entry**

In `validator/explanations.go`, inside `reachabilityExplanations()`'s returned map (after the `DIP105` entry):

```go
		DIP144: {
			Code:    DIP144,
			Summary: "agent node has no failure route",
			Trigger: "An agent node can fail at runtime but has no fail edge, fallback_target, bounded retry, or graph-level on_failure.",
			Fix:     "Add a `-> <node> when ctx.outcome = fail` edge, set fallback_target, add retry_target with max_retries, or declare a workflow-level on_failure.",
			Example: "agent Implement\n  prompt: \"build it\"\nImplement -> Test  // DIP144: no route if Implement fails",
		},
```

- [ ] **Step 5: Create the lint implementation**

Create `validator/lint_failure_route.go`:

```go
package validator

import (
	"fmt"

	"github.com/2389-research/dippin-lang/ir"
)

// lintAgentFailureRoute checks DIP144: an agent node that can fail at runtime
// but has no way to route that failure — no fail edge, no fallback_target, no
// bounded retry target, and no graph-level on_failure. Such a node is a silent
// dead-end. Scoped to agent nodes; the exit node and nodes with no outgoing
// edges are skipped (failure there ends the run, not a routing gap).
func lintAgentFailureRoute(w *ir.Workflow) []Diagnostic {
	var diags []Diagnostic
	for _, n := range w.Nodes {
		if needsFailureRoute(w, n) {
			diags = append(diags, Diagnostic{
				Code:     DIP144,
				Severity: SeverityWarning,
				Message:  fmt.Sprintf("agent node %q has no failure route (no fail edge, no fallback_target, no graph on_failure)", n.ID),
				Location: n.Source,
				Help:     "add `-> <node> when ctx.outcome = fail`, set fallback_target:, or declare a workflow-level on_failure:",
			})
		}
	}
	return diags
}

// needsFailureRoute reports whether an agent node lacks any failure route.
func needsFailureRoute(w *ir.Workflow, n *ir.Node) bool {
	if _, ok := n.Config.(ir.AgentConfig); !ok {
		return false
	}
	if n.ID == w.Exit {
		return false
	}
	outgoing := w.EdgesFrom(n.ID)
	if len(outgoing) == 0 {
		return false
	}
	return !hasFailureRoute(w, n, outgoing)
}

// hasFailureRoute reports whether the node has any failure-handling route.
func hasFailureRoute(w *ir.Workflow, n *ir.Node, outgoing []*ir.Edge) bool {
	if w.Defaults.OnFailure != "" {
		return true
	}
	if hasBoundedFailureTarget(n.Retry) {
		return true
	}
	return hasFailEdge(outgoing)
}

// hasBoundedFailureTarget reports whether retry config supplies a recovery target:
// a fallback_target, or a retry_target bounded by max_retries.
func hasBoundedFailureTarget(r ir.RetryConfig) bool {
	if r.FallbackTarget != "" {
		return true
	}
	return r.RetryTarget != "" && r.MaxRetries > 0
}

// hasFailEdge reports whether any outgoing edge routes on outcome = fail/failure.
// An unconditional/success edge does NOT count — a hard failure does not traverse it.
func hasFailEdge(edges []*ir.Edge) bool {
	for _, e := range edges {
		cmp, ok := extractEqualityCondition(e)
		if ok && isOutcomeVar(cmp.Variable) && isFailValue(cmp.Value) {
			return true
		}
	}
	return false
}

// isOutcomeVar matches the outcome variable in both namespaced and bare forms.
func isOutcomeVar(v string) bool { return v == "ctx.outcome" || v == "outcome" }

// isFailValue matches the failure outcome values the engine recognizes.
func isFailValue(v string) bool { return v == "fail" || v == "failure" }
```

- [ ] **Step 6: Register the lint**

In `validator/lint.go`, add to the `Lint` append chain (after `lintSubgraphToolAccess(w)`):

```go
	diags = append(diags, lintSubgraphToolAccess(w)...)
	diags = append(diags, lintAgentFailureRoute(w)...)
```

Update the `Lint` doc comment (line 8) from `DIP101–DIP143` to `DIP101–DIP144`.

- [ ] **Step 7: Run tests to verify they pass**

Run: `just test-pkg validator`
Expected: all DIP144 tests PASS; `TestExplanationsCoverAllCodes` / `TestExplanationsNoExtra` PASS (DIP144 in both maps).

- [ ] **Step 8: Complexity + commit**

Run: `just complexity` (all new helpers ≤5 cyclomatic / ≤7 cognitive).

```bash
cd /home/clint/code/2389/dippin-lang/.claude/worktrees/feat+92-on-failure-route
git add validator/lint_failure_route.go validator/lint_failure_route_test.go validator/lint.go validator/lint_codes.go validator/explanations.go
git commit -m "feat: DIP144 lint — agent node without a failure route (#93)

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

### Task 7: Example, docs, range bumps, spec regen, wasm

**Files:**
- Create: `examples/on_failure_route.dip`
- Modify: docs (`docs/validation.md`, `docs/llm-reference.md`, `docs/edges.md`, `docs/nodes.md`, `docs/syntax.md`)
- Modify: `CLAUDE.md` (count + range)
- Regenerate: `docs/generated-spec.md` + `cmd/dippin/generated-spec.md` (via pre-commit/`just spec-check` — do NOT hand-edit)

- [ ] **Step 1: Measure the DIP144 blast radius**

```bash
cd /home/clint/code/2389/dippin-lang/.claude/worktrees/feat+92-on-failure-route
just build
for f in examples/*.dip; do n=$(./dippin lint "$f" 2>/dev/null | grep -c DIP144); echo "$n  $f"; done | sort -rn
```

Record the counts. Decision rule: leave `stress_*`/adversarial fixtures warning; only consider editing a couple of flagship examples (e.g. `orchestrator.dip`, `api_design.dip`) if a one-line `on_failure` is natural. Do NOT mass-edit. Lint warnings do not fail CI.

- [ ] **Step 2: Create the lint-clean demonstrator example**

Create `examples/on_failure_route.dip`:

```
workflow OnFailureDemo
  goal: "Demonstrate a graph-level on_failure recovery route"
  start: Plan
  exit: Done

  defaults
    on_failure: Escalate

  agent Plan
    prompt: "Draft a plan."
    model: claude-sonnet-4-6
    provider: anthropic

  agent Build
    prompt: "Execute the plan."
    model: claude-sonnet-4-6
    provider: anthropic

  human Escalate
    mode: freeform
    prompt: "An automated step failed — please intervene."

  agent Done
    prompt: "Summarize the result."
    model: claude-sonnet-4-6
    provider: anthropic

  edges
    Plan -> Build
    Build -> Done
    Escalate -> Done
```

- [ ] **Step 3: Verify the example is structurally valid and DIP144-clean**

```bash
cd /home/clint/code/2389/dippin-lang/.claude/worktrees/feat+92-on-failure-route
./dippin validate examples/on_failure_route.dip   # expect: validation passed
./dippin lint examples/on_failure_route.dip        # expect: no DIP144, no DIP108
```

Expected: `validate` passes (Escalate reachable via the on_failure seed — exercises Task 3 end-to-end); `lint` shows no DIP144 (on_failure suppresses it) and no DIP108 (valid model). If DIP108 fires, swap the model for one verified in `validator/lint_model.go`.

- [ ] **Step 4: Document the attribute, the lint, and precedence**

- `docs/validation.md`: add a `DIP144` row/section alongside DIP101–DIP143; bump the "DIP101–DIP143" range strings and the code count to include DIP144.
- `docs/llm-reference.md`: bump the "DIP101–DIP143" range and the "52 ... codes" count by one.
- `docs/edges.md` (Routing Priority cascade, ~line 110): document the 5-level failure precedence (fail edge > bounded retry > fallback_target > on_failure > halt), stating that dippin validates existence/shape and the runtime owns ordering, and that a `restart:true` fail edge carries its own `max_restarts` budget.
- `docs/nodes.md:55-56`: cross-reference `fallback_target` to the graph-level `on_failure` catch-all.
- `docs/syntax.md` (defaults table): add an `on_failure` row.
- `CLAUDE.md:85`: bump the code count and the "DIP101-DIP143" range to DIP144.

- [ ] **Step 5: Regenerate the spec (never hand-edit generated files)**

```bash
cd /home/clint/code/2389/dippin-lang/.claude/worktrees/feat+92-on-failure-route
just spec-check
git status --short   # docs/generated-spec.md + cmd/dippin/generated-spec.md may change
```

- [ ] **Step 6: Verify the wasm build (Netlify preview gate)**

```bash
cd /home/clint/code/2389/dippin-lang/.claude/worktrees/feat+92-on-failure-route
GOOS=js GOARCH=wasm go build ./cmd/wasm/ ./validator/
```

Expected: builds cleanly, no output. (This is a build-only verification the justfile's `wasm` recipe also covers via `just wasm`.)

- [ ] **Step 7: Full suite + commit**

Run: `just test` then `just complexity` then `just validate-examples`.
Expected: all green. (`just check` additionally runs tree-sitter-generate which FAILS in this env with no tree-sitter CLI — treat clean-everything-before-tree-sitter as green; the pre-commit hook is the real gate.)

```bash
cd /home/clint/code/2389/dippin-lang/.claude/worktrees/feat+92-on-failure-route
git add examples/on_failure_route.dip docs/ cmd/dippin/generated-spec.md CLAUDE.md
git commit -m "docs: document on_failure + DIP144, add demonstrator example, bump code range (#92, #93)

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

### Task 8: Open the PR

- [ ] **Step 1: Push and open the PR**

```bash
cd /home/clint/code/2389/dippin-lang/.claude/worktrees/feat+92-on-failure-route
git push -u origin feat/92-on-failure-route
```

- [ ] **Step 2: Create the PR** with body covering: the `on_failure` attribute + precedence semantics; the DIP144 lint + its suppression set; the DIP003 node-existence validation + DIP004 reachability seeding; the DOT round-trip decision (tool-safety precedent); the DIP102-suppression descope rationale; the examples blast-radius decision; and the cross-repo split (dippin carries/validates, tracker enforces — note the tracker `upstream` follow-up to be filed, blocking-linked to #295, and the `tracker validate` warning-fatality question). Title: `feat: graph-level on_failure route + DIP144 failure-route lint (Closes #92, Closes #93)`.

- [ ] **Step 3: Do NOT tag a release.** Per the maintainer, tagging (~v0.37.0) requires explicit go-ahead.
```
