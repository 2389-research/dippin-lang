# dip 2 — Edges Own Destinations Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Under `dip 2`, reject `retry_target`/`fallback_target` as node fields and express failure destinations via the `on fail` edge, with `fmt --migrate` converting v1 files (flagging non-1:1 cases).

**Architecture:** Version-gated rejection in the parser; a v1→v2 IR transform (`formatter.MigrateToV2`) wired into `fmt --migrate`; the validator's failure-route lints retargeted to a unified helper backed by a new `ir.EdgeRoutesOnFail`. No new syntax, no grammar change; the simulator is untouched (cascade-ordering delta is documented only).

**Tech Stack:** Go; `parser`, `ir`, `formatter`, `validator`, `cmd/dippin`; `just` recipes; pre-commit hook is the authoritative gate.

**Spec:** `docs/superpowers/specs/2026-07-09-issue-134-dip2-edges-own-destinations-design.md`

**Environment:** every shell step assumes `export PATH=/usr/local/go/bin:$HOME/go/bin:$PATH`. Commit via `git commit` (the pre-commit hook runs the full suite — never `--no-verify`). Avoid `just check` (fails on tree-sitter generate here); use individual recipes. Complexity caps are hard: cyclomatic ≤5, cognitive ≤7 — extract helpers, never `//nolint`.

---

## File Structure

- **Modify** `ir/edge.go` — add `Edge.Comment` field; add `ir.EdgeRoutesOnFail`.
- **Modify** `formatter/format.go` — render `Edge.Comment` before the edge line.
- **Modify** `parser/parse_nodes.go` — v2-reject `retry_target`/`fallback_target`.
- **Modify** `validator/lint_retry.go`, `validator/lint_failure_route.go` — unified `nodeHasFailureRoute`; retarget DIP104/DIP115/DIP144.
- **Create** `formatter/migrate_v2.go` (+ test) — `MigrateToV2` transform + `MigrationNote`.
- **Modify** `cmd/dippin/cmd_fmt.go`, `cmd/dippin/cli.go` — wire `--migrate`; `ExitMigrateReview` code.
- **Modify** `docs/edges.md` + regen embedded spec.

---

### Task 1: `ir.EdgeRoutesOnFail` + `Edge.Comment`

**Files:**
- Modify: `ir/edge.go`
- Test: `ir/edge_fail_test.go` (create)

- [ ] **Step 1: Write the failing test.** Create `ir/edge_fail_test.go`:

```go
package ir

import "testing"

func failEdge(variable, value string) *Edge {
	return &Edge{From: "A", To: "B", Condition: &Condition{
		Parsed: CondCompare{Variable: variable, Op: "=", Value: value},
	}}
}

func TestEdgeRoutesOnFail(t *testing.T) {
	cases := []struct {
		name string
		edge *Edge
		want bool
	}{
		{"ctx.outcome = fail", failEdge("ctx.outcome", "fail"), true},
		{"ctx.outcome = failure", failEdge("ctx.outcome", "failure"), true},
		{"bare outcome = fail", failEdge("outcome", "fail"), true},
		{"ctx.outcome = success", failEdge("ctx.outcome", "success"), false},
		{"other variable = fail", failEdge("ctx.tool_marker", "fail"), false},
		{"unconditional edge", &Edge{From: "A", To: "B"}, false},
		{"unparsed condition", &Edge{From: "A", To: "B", Condition: &Condition{Raw: "ctx.outcome = fail"}}, false},
	}
	for _, tc := range cases {
		if got := EdgeRoutesOnFail(tc.edge); got != tc.want {
			t.Errorf("%s: EdgeRoutesOnFail = %v, want %v", tc.name, got, tc.want)
		}
	}
}
```

- [ ] **Step 2: Run — expect FAIL** (`undefined: EdgeRoutesOnFail`).

Run: `go test ./ir/ -run TestEdgeRoutesOnFail`
Expected: build error `undefined: EdgeRoutesOnFail`.

- [ ] **Step 3: Implement.** In `ir/edge.go`, add `Comment` to the `Edge` struct (after `Override`):

```go
	Override  bool // Carried, not interpreted: human-authored validation override (tracker#271)
	Comment   string // Optional leading `# ` line the formatter emits before this edge (migration review notes)
	Source    SourceLocation
```

Then append to `ir/edge.go`:

```go
// EdgeRoutesOnFail reports whether an edge's guard routes the failure outcome
// (ctx.outcome / outcome = fail / failure). Requires Condition.Parsed (populated
// by simulate.EnsureConditionsParsed); an edge whose AST is not yet parsed
// returns false.
func EdgeRoutesOnFail(e *Edge) bool {
	cmp, ok := ExtractEqualityCondition(e)
	return ok && isOutcomeVariable(cmp.Variable) && isFailOutcome(cmp.Value)
}

func isOutcomeVariable(v string) bool { return v == "ctx.outcome" || v == "outcome" }

func isFailOutcome(v string) bool { return v == "fail" || v == "failure" }
```

- [ ] **Step 4: Run — expect PASS.**

Run: `go test ./ir/ -run TestEdgeRoutesOnFail -v`
Expected: PASS.

- [ ] **Step 5: Build all + complexity.**

Run: `go build ./... && just complexity`
Expected: builds; `Complexity OK.`

- [ ] **Step 6: Commit.**

```bash
git add ir/edge.go ir/edge_fail_test.go
git commit -m "feat(ir): add EdgeRoutesOnFail helper + Edge.Comment for dip 2 migration"
```

---

### Task 2: formatter renders `Edge.Comment`

**Files:**
- Modify: `formatter/format.go` (`writeEdge`, ~line 756)
- Test: `formatter/format_test.go` (add)

- [ ] **Step 1: Write the failing test.** Add to `formatter/format_test.go`:

```go
func TestFormatEdgeComment(t *testing.T) {
	w := &ir.Workflow{
		Name: "W", Version: "2", Start: "A", Exit: "B",
		Nodes: []*ir.Node{
			{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "a"}},
			{ID: "B", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "b"}},
		},
		Edges: []*ir.Edge{
			{From: "A", To: "B", Comment: "MIGRATION: review this"},
		},
	}
	out := formatter.Format(w)
	if !strings.Contains(out, "# MIGRATION: review this\n") {
		t.Errorf("expected leading migration comment line, got:\n%s", out)
	}
	// The comment must appear immediately before the edge.
	if !strings.Contains(out, "# MIGRATION: review this\n    A -> B") {
		t.Errorf("comment not positioned before its edge:\n%s", out)
	}
}
```

(Ensure `formatter/format_test.go` imports `strings`, `ir`, `formatter` — match the existing test file's imports.)

- [ ] **Step 2: Run — expect FAIL** (no comment emitted).

Run: `go test ./formatter/ -run TestFormatEdgeComment`
Expected: FAIL (comment absent).

- [ ] **Step 3: Implement.** In `formatter/format.go`, change `writeEdge` (line ~756):

```go
func writeEdge(wr *writer, w *ir.Workflow, e *ir.Edge) {
	if e.Comment != "" {
		wr.line("# %s", e.Comment)
	}
	var parts []string
	parts = append(parts, fmt.Sprintf("%s -> %s", e.From, e.To))
	parts = appendEdgeCondition(parts, w, e)
	parts = appendEdgeAttrs(parts, e)
	wr.line("%s", strings.Join(parts, "  "))
}
```

- [ ] **Step 4: Run — expect PASS.**

Run: `go test ./formatter/ -run TestFormatEdgeComment -v`
Expected: PASS.

- [ ] **Step 5: Full formatter suite (no regression on existing round-trips).**

Run: `go test ./formatter/ -count=1`
Expected: `ok`.

- [ ] **Step 6: Commit.**

```bash
git add formatter/format.go formatter/format_test.go
git commit -m "feat(formatter): emit Edge.Comment as a leading # line"
```

---

### Task 3: parser v2-rejects `retry_target`/`fallback_target`

**Files:**
- Modify: `parser/parse_nodes.go` (`tryApplyCommonField`, ~line 171)
- Test: `parser/parse_dip2_test.go` (create)

- [ ] **Step 1: Write the failing test.** Create `parser/parse_dip2_test.go`:

```go
package parser

import "strings"
import "testing"

const dip2Body = `workflow W
  goal: "t"
  start: A
  exit: B

  agent A
    prompt: a
    %s: B

  agent B
    prompt: b

  edges
    A -> B
`

func TestDip2RejectsRetryTarget(t *testing.T) {
	_, err := NewParser("dip 2\n\n"+strings_ReplaceOnce(dip2Body, "%s", "retry_target"), "t.dip").Parse()
	if err == nil || !strings.Contains(err.Error(), "not a node field in dip 2") {
		t.Fatalf("want dip-2 rejection error for retry_target, got %v", err)
	}
}

func TestDip2RejectsFallbackTarget(t *testing.T) {
	_, err := NewParser("dip 2\n\n"+strings_ReplaceOnce(dip2Body, "%s", "fallback_target"), "t.dip").Parse()
	if err == nil || !strings.Contains(err.Error(), "not a node field in dip 2") {
		t.Fatalf("want dip-2 rejection error for fallback_target, got %v", err)
	}
}

func TestV1AcceptsRetryTarget(t *testing.T) {
	w, err := NewParser(strings_ReplaceOnce(dip2Body, "%s", "retry_target"), "t.dip").Parse()
	if err != nil {
		t.Fatalf("v1 must still accept retry_target: %v", err)
	}
	if w.Nodes[0].Retry.RetryTarget != "B" {
		t.Errorf("v1 RetryTarget = %q, want B", w.Nodes[0].Retry.RetryTarget)
	}
}

func strings_ReplaceOnce(s, old, new string) string { return strings.Replace(s, old, new, 1) }
```

- [ ] **Step 2: Run — expect FAIL** (v2 currently accepts these fields).

Run: `go test ./parser/ -run 'TestDip2Rejects|TestV1Accepts'`
Expected: the two `Rejects` tests FAIL (no error today); `V1Accepts` passes.

- [ ] **Step 3: Implement.** In `parser/parse_nodes.go`, change `tryApplyCommonField` (line 171):

```go
func (p *Parser) tryApplyCommonField(n *ir.Node, key, val string, loc ir.SourceLocation) bool {
	if p.version >= 2 && isV2RejectedNodeField(key) {
		p.diagnostics = append(p.diagnostics, fmt.Sprintf(
			"%q is not a node field in dip 2 — express the failure destination as an `on fail` edge (run `dippin fmt --migrate`) at %d:%d",
			key, loc.Line, loc.Column))
		return true // handled (rejected) — do not fall through to unknown-field hint
	}
	if applyCommonStringField(n, key, val) {
		return true
	}
	return p.applyCommonComplexField(n, key, val, loc)
}

// isV2RejectedNodeField lists node fields removed under dip 2 (their destinations
// move to the edges block; see #134).
func isV2RejectedNodeField(key string) bool {
	return key == "retry_target" || key == "fallback_target"
}
```

(Confirm `fmt` is imported in `parser/parse_nodes.go` — it is, used by existing diagnostics.)

- [ ] **Step 4: Run — expect PASS.**

Run: `go test ./parser/ -run 'TestDip2Rejects|TestV1Accepts' -v`
Expected: all PASS.

- [ ] **Step 5: Full parser suite (v1 unchanged).**

Run: `go test ./parser/ -count=1 && just complexity`
Expected: `ok`; `Complexity OK.`

- [ ] **Step 6: Commit.**

```bash
git add parser/parse_nodes.go parser/parse_dip2_test.go
git commit -m "feat(parser): reject retry_target/fallback_target as node fields under dip 2"
```

---

### Task 4: unify the failure-route lints (DIP104 / DIP115 / DIP144)

**Files:**
- Modify: `validator/lint_failure_route.go` (add `nodeHasFailureRoute`; `hasFailEdge` → `ir.EdgeRoutesOnFail`)
- Modify: `validator/lint_retry.go` (DIP104 `isUnboundedRetry`, DIP115 `needsGoalGateFallback`)
- Test: `validator/lint_failure_route_test.go` (add)

- [ ] **Step 1: Write the failing test.** Add to `validator/lint_failure_route_test.go` (create if absent; `package validator`, import `strings`/`testing`/`parser`):

```go
func diagsForCode(t *testing.T, src, code string) int {
	t.Helper()
	w, err := parser.NewParser(src, "t.dip").Parse()
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	n := 0
	for _, d := range Lint(w).Diagnostics {
		if d.Code == code {
			n++
		}
	}
	return n
}

const goalGateV1 = `workflow W
  goal: "t"
  start: G
  exit: D
  agent G
    prompt: g
    goal_gate: true
    fallback_target: D
  agent D
    prompt: d
  edges
    G -> D
`

const goalGateV2Edge = `workflow W
  goal: "t"
  start: G
  exit: D
  agent G
    prompt: g
    goal_gate: true
  agent D
    prompt: d
  edges
    G -> D  on fail
`

func TestDIP115_SatisfiedByFallbackTargetAndByFailEdge(t *testing.T) {
	if got := diagsForCode(t, goalGateV1, "DIP115"); got != 0 {
		t.Errorf("v1 fallback_target should satisfy DIP115, got %d", got)
	}
	if got := diagsForCode(t, goalGateV2Edge, "DIP115"); got != 0 {
		t.Errorf("v2 on-fail edge should satisfy DIP115, got %d", got)
	}
}
```

- [ ] **Step 2: Run — expect FAIL** (the v2-edge case currently warns because DIP115 reads only the target fields).

Run: `go test ./validator/ -run TestDIP115_SatisfiedByFallbackTargetAndByFailEdge`
Expected: FAIL (`v2 on-fail edge should satisfy DIP115, got 1`).

- [ ] **Step 3: Implement.** In `validator/lint_failure_route.go`, replace the `hasFailEdge` / `isOutcomeVar` / `isFailValue` block with a delegation to `ir`, and add the shared helper:

```go
// nodeHasFailureRoute reports whether a node has a recovery path for failure —
// a bounded retry target (v1 fallback_target / retry_target+max_retries) or an
// `on fail` edge (the dip 2 form). Shared by DIP104, DIP115, DIP144.
func nodeHasFailureRoute(w *ir.Workflow, n *ir.Node) bool {
	return hasBoundedFailureTarget(n.Retry) || hasFailEdge(w.EdgesFrom(n.ID))
}

// hasFailEdge reports whether any outgoing edge routes on outcome = fail/failure.
func hasFailEdge(edges []*ir.Edge) bool {
	for _, e := range edges {
		if ir.EdgeRoutesOnFail(e) {
			return true
		}
	}
	return false
}
```

Delete the old `isOutcomeVar` / `isFailValue` functions (now in `ir`). Update `hasFailureRoute` (DIP144) to reuse the shared helper:

```go
func hasFailureRoute(w *ir.Workflow, n *ir.Node, outgoing []*ir.Edge) bool {
	if w.Defaults.OnFailure != "" && w.Node(w.Defaults.OnFailure) != nil {
		return true
	}
	return nodeHasFailureRoute(w, n)
}
```

(`hasFailEdge`'s single caller `hasFailureRoute` now routes through `nodeHasFailureRoute`, but keep `hasFailEdge` — `nodeHasFailureRoute` uses it. If `lint_failure_route_test.go` referenced the removed `extractEqualityCondition` earlier it already uses `ir.ExtractEqualityCondition`.)

In `validator/lint_retry.go`, retarget DIP104 and DIP115 to the shared helper.

DIP104 — change `lintUnboundedRetry` to pass the workflow and node:

```go
func lintUnboundedRetry(w *ir.Workflow) []Diagnostic {
	var diags []Diagnostic
	for _, n := range w.Nodes {
		if isUnboundedRetry(w, n) {
			diags = append(diags, Diagnostic{
				Code:     DIP104,
				Severity: SeverityWarning,
				Message:  fmt.Sprintf("node %q has retry configuration but no max_retries and no failure route", n.ID),
				Location: n.Source,
				Help:     "set max_retries to bound retries, or add an `on fail` edge / fallback_target for graceful degradation",
			})
		}
	}
	return diags
}

// isUnboundedRetry returns true if retry config exists with no bound: no
// max_retries and no failure route (fallback_target or `on fail` edge).
func isUnboundedRetry(w *ir.Workflow, n *ir.Node) bool {
	r := n.Retry
	hasRetryConfig := r.Policy != "" || r.RetryTarget != ""
	return hasRetryConfig && r.MaxRetries == 0 && !nodeHasFailureRoute(w, n)
}
```

DIP115 — change `needsGoalGateFallback` to take the workflow:

```go
func lintGoalGateFallback(w *ir.Workflow) []Diagnostic {
	var diags []Diagnostic
	for _, n := range w.Nodes {
		if needsGoalGateFallback(w, n) {
			diags = append(diags, Diagnostic{
				Code:     DIP115,
				Severity: SeverityWarning,
				Message:  fmt.Sprintf("node %q has goal_gate: true but no failure route", n.ID),
				Location: n.Source,
				Help:     "add an `on fail` edge, set fallback_target:, or add retry_target with max_retries so the pipeline can recover when the gate fails",
			})
		}
	}
	return diags
}

// needsGoalGateFallback returns true if a goal_gate node has no recovery path.
func needsGoalGateFallback(w *ir.Workflow, n *ir.Node) bool {
	cfg, ok := n.Config.(ir.AgentConfig)
	if !ok || !cfg.GoalGate {
		return false
	}
	return !nodeHasFailureRoute(w, n)
}
```

- [ ] **Step 4: Run — expect PASS** (new + full validator suite; DIP144's existing tests must still pass, and `Lint` populates `Condition.Parsed` before these run so `ir.EdgeRoutesOnFail` sees the AST).

Run: `go test ./validator/ -run 'TestDIP115|TestDIP104|TestDIP144|FailureRoute' -v && go test ./validator/ -count=1`
Expected: PASS; `ok`.

- [ ] **Step 5: Complexity + lint.**

Run: `just complexity && golangci-lint run ./validator/... ./ir/...`
Expected: `Complexity OK.`; `0 issues`.

- [ ] **Step 6: Commit.**

```bash
git add validator/lint_failure_route.go validator/lint_retry.go validator/lint_failure_route_test.go
git commit -m "feat(validator): unify DIP104/DIP115/DIP144 on nodeHasFailureRoute (v1 target or on-fail edge)"
```

---

### Task 5: `formatter.MigrateToV2` transform

**Files:**
- Create: `formatter/migrate_v2.go`
- Test: `formatter/migrate_v2_test.go`

- [ ] **Step 1: Write the failing test.** Create `formatter/migrate_v2_test.go`:

```go
package formatter

import (
	"testing"

	"github.com/2389-research/dippin-lang/ir"
	"github.com/2389-research/dippin-lang/simulate"
)

// v1wf builds a v1 workflow with the given retry config on node "T" and the
// given extra edges from T; conditions are parsed so EdgeRoutesOnFail works.
func v1wf(t *testing.T, retry ir.RetryConfig, extra []*ir.Edge) *ir.Workflow {
	t.Helper()
	nodes := []*ir.Node{
		{ID: "T", Kind: ir.NodeTool, Config: ir.ToolConfig{Command: "run"}, Retry: retry},
		{ID: "Done", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "d"}},
		{ID: "Esc", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "e"}},
	}
	edges := append([]*ir.Edge{{From: "T", To: "Done", Condition: &ir.Condition{Raw: "ctx.outcome = success"}}}, extra...)
	w := &ir.Workflow{Name: "W", Version: "1", Start: "T", Exit: "Done", Nodes: nodes, Edges: edges}
	_ = simulate.EnsureConditionsParsed(w)
	return w
}

func failTo(to string) *ir.Edge {
	return &ir.Edge{From: "T", To: to, Condition: &ir.Condition{Raw: "ctx.outcome = fail"}}
}

func edgeTo(w *ir.Workflow, to string) *ir.Edge {
	for _, e := range w.Edges {
		if e.From == "T" && e.To == to {
			return e
		}
	}
	return nil
}

func TestMigrate_FallbackNoFailEdge_Synthesizes(t *testing.T) {
	w := v1wf(t, ir.RetryConfig{FallbackTarget: "Esc"}, nil)
	notes := MigrateToV2(w)
	if len(notes) != 0 {
		t.Fatalf("clean synthesize should have no notes, got %v", notes)
	}
	if w.Version != "2" || w.Nodes[0].Retry.FallbackTarget != "" {
		t.Fatalf("version/field not cleared: v=%s fb=%q", w.Version, w.Nodes[0].Retry.FallbackTarget)
	}
	e := edgeTo(w, "Esc")
	if e == nil || !ir.EdgeRoutesOnFail(e) {
		t.Fatalf("expected synthesized on-fail edge T->Esc, edges=%v", w.Edges)
	}
}

func TestMigrate_FallbackMatchesFailEdge_Dedupes(t *testing.T) {
	w := v1wf(t, ir.RetryConfig{FallbackTarget: "Esc"}, []*ir.Edge{failTo("Esc")})
	notes := MigrateToV2(w)
	if len(notes) != 0 {
		t.Fatalf("matching target should dedupe with no notes, got %v", notes)
	}
	count := 0
	for _, e := range w.Edges {
		if e.To == "Esc" {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("expected exactly one T->Esc edge after dedupe, got %d", count)
	}
}

func TestMigrate_FallbackDivergesFromFailEdge_FlagsAndKeepsBoth(t *testing.T) {
	w := v1wf(t, ir.RetryConfig{FallbackTarget: "Esc"}, []*ir.Edge{failTo("Done")})
	notes := MigrateToV2(w)
	if len(notes) != 1 {
		t.Fatalf("divergent fallback should produce exactly one review note, got %v", notes)
	}
	e := edgeTo(w, "Esc")
	if e == nil || e.Comment == "" {
		t.Fatalf("expected a flagged T->Esc on-fail edge with a comment, edges=%v", w.Edges)
	}
}

func TestMigrate_SelfRetryTarget_Dropped(t *testing.T) {
	w := v1wf(t, ir.RetryConfig{RetryTarget: "T", MaxRetries: 2}, nil)
	notes := MigrateToV2(w)
	if len(notes) != 0 || w.Nodes[0].Retry.RetryTarget != "" {
		t.Fatalf("self retry_target should drop silently, notes=%v rt=%q", notes, w.Nodes[0].Retry.RetryTarget)
	}
	if edgeTo(w, "T") != nil {
		t.Fatalf("self retry_target must not synthesize a loop edge")
	}
}

func TestMigrate_NonSelfRetryTarget_LoopEdgeAndNote(t *testing.T) {
	w := v1wf(t, ir.RetryConfig{RetryTarget: "Esc", MaxRetries: 2}, nil)
	notes := MigrateToV2(w)
	if len(notes) != 1 {
		t.Fatalf("non-self retry_target should produce one note, got %v", notes)
	}
	e := edgeTo(w, "Esc")
	if e == nil || !e.Restart {
		t.Fatalf("expected a loop edge T->Esc, edges=%v", w.Edges)
	}
}
```

- [ ] **Step 2: Run — expect FAIL** (`undefined: MigrateToV2`).

Run: `go test ./formatter/ -run TestMigrate_`
Expected: build error `undefined: MigrateToV2`.

- [ ] **Step 3: Implement.** Create `formatter/migrate_v2.go`:

```go
// ABOUTME: v1 -> dip 2 IR transform: retry_target/fallback_target node fields
// ABOUTME: become on-fail / loop edges. Used by `dippin fmt --migrate`.
package formatter

import (
	"fmt"

	"github.com/2389-research/dippin-lang/ir"
)

// MigrationNote records a case the v1->v2 transform could not express 1:1.
type MigrationNote struct {
	Node    string
	Message string
}

// MigrateToV2 rewrites a v1 workflow into dip 2 in place and returns review
// notes. Requires Condition.Parsed to be populated (caller runs
// simulate.EnsureConditionsParsed). No-op if the workflow is already v2.
func MigrateToV2(w *ir.Workflow) []MigrationNote {
	if w.Version == "2" {
		return nil
	}
	var notes []MigrationNote
	for _, n := range w.Nodes {
		notes = append(notes, migrateFallbackTarget(w, n)...)
		notes = append(notes, migrateRetryTarget(w, n)...)
	}
	w.Version = "2"
	return notes
}

// migrateFallbackTarget turns a node's fallback_target into an `on fail` edge:
// synthesize when absent, dedupe when it matches an existing fail edge, or flag
// (keep both + comment) when it diverges.
func migrateFallbackTarget(w *ir.Workflow, n *ir.Node) []MigrationNote {
	f := n.Retry.FallbackTarget
	if f == "" {
		return nil
	}
	n.Retry.FallbackTarget = ""
	existing := failEdgeTarget(w, n.ID)
	if existing == f {
		return nil // destinations agree — dedupe
	}
	if existing == "" {
		w.Edges = append(w.Edges, newFailEdge(n.ID, f, ""))
		return nil
	}
	msg := fmt.Sprintf("MIGRATION: v1 fallback_target was %q (differs from the on-fail edge -> %q) — pick one", f, existing)
	w.Edges = append(w.Edges, newFailEdge(n.ID, f, msg))
	return []MigrationNote{{Node: n.ID, Message: msg}}
}

// migrateRetryTarget drops a self retry_target (max_retries alone means re-run in
// place) and converts a non-self retry_target into a `loop` edge + review note.
func migrateRetryTarget(w *ir.Workflow, n *ir.Node) []MigrationNote {
	r := n.Retry.RetryTarget
	if r == "" {
		return nil
	}
	n.Retry.RetryTarget = ""
	if r == n.ID {
		return nil
	}
	msg := fmt.Sprintf("MIGRATION: v1 retry_target -> %q (non-self) became a loop edge — verify the loop intent", r)
	w.Edges = append(w.Edges, &ir.Edge{From: n.ID, To: r, Restart: true, Comment: msg})
	return []MigrationNote{{Node: n.ID, Message: msg}}
}

// failEdgeTarget returns the target of the first outgoing on-fail edge, or "".
func failEdgeTarget(w *ir.Workflow, nodeID string) string {
	for _, e := range w.EdgesFrom(nodeID) {
		if ir.EdgeRoutesOnFail(e) {
			return e.To
		}
	}
	return ""
}

// newFailEdge builds a `-> to on fail` edge with an optional leading comment.
func newFailEdge(from, to, comment string) *ir.Edge {
	return &ir.Edge{From: from, To: to, Comment: comment, Condition: &ir.Condition{
		Raw:    "ctx.outcome = fail",
		Parsed: ir.CondCompare{Variable: "ctx.outcome", Op: "=", Value: "fail"},
	}}
}
```

- [ ] **Step 4: Run — expect PASS.**

Run: `go test ./formatter/ -run TestMigrate_ -v`
Expected: all PASS.

- [ ] **Step 5: Complexity.** Each helper is small; confirm the caps.

Run: `just complexity && golangci-lint run ./formatter/...`
Expected: `Complexity OK.`; `0 issues`.

- [ ] **Step 6: Commit.**

```bash
git add formatter/migrate_v2.go formatter/migrate_v2_test.go
git commit -m "feat(formatter): MigrateToV2 — fold retry_target/fallback_target into edges"
```

---

### Task 6: wire `fmt --migrate`

**Files:**
- Modify: `cmd/dippin/cli.go` (add `ExitMigrateReview` code)
- Modify: `cmd/dippin/cmd_fmt.go`
- Test: `cmd/dippin/cmd_fmt_test.go` (add)

- [ ] **Step 1: Write the failing test.** Add to `cmd/dippin/cmd_fmt_test.go`:

```go
func writeTmp(t *testing.T, body string) string {
	t.Helper()
	dir := t.TempDir()
	p := filepath.Join(dir, "in.dip")
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

const v1Fallback = `workflow W
  goal: "t"
  start: T
  exit: Done
  tool T
    command: run
    fallback_target: Esc
  agent Esc
    prompt: e
  agent Done
    prompt: d
  edges
    T -> Done  on success
`

func TestFmtMigrate_SynthesizesFailEdge(t *testing.T) {
	var stdout, stderr bytes.Buffer
	cli := &CLI{Stdout: &stdout, Stderr: &stderr, Format: FormatText}
	code := cli.CmdFmt([]string{"--migrate", writeTmp(t, v1Fallback)})
	if code != ExitOK {
		t.Fatalf("clean migrate should exit OK, got %d; stderr=%s", code, stderr.String())
	}
	out := stdout.String()
	if !strings.HasPrefix(out, "dip 2\n") {
		t.Errorf("migrated output should declare dip 2:\n%s", out)
	}
	if strings.Contains(out, "fallback_target") {
		t.Errorf("migrated output must not keep fallback_target:\n%s", out)
	}
	if !strings.Contains(out, "T -> Esc  on fail") {
		t.Errorf("expected synthesized on-fail edge:\n%s", out)
	}
}

const v1Divergent = `workflow W
  goal: "t"
  start: T
  exit: Done
  tool T
    command: run
    fallback_target: Esc
  agent Esc
    prompt: e
  agent Fix
    prompt: f
  agent Done
    prompt: d
  edges
    T -> Done  on success
    T -> Fix   on fail
`

func TestFmtMigrate_DivergentFlagsReviewExit(t *testing.T) {
	var stdout, stderr bytes.Buffer
	cli := &CLI{Stdout: &stdout, Stderr: &stderr, Format: FormatText}
	code := cli.CmdFmt([]string{"--migrate", writeTmp(t, v1Divergent)})
	if code != ExitMigrateReview {
		t.Fatalf("divergent migrate should exit ExitMigrateReview, got %d", code)
	}
	if !strings.Contains(stdout.String(), "# MIGRATION:") {
		t.Errorf("expected inline MIGRATION comment:\n%s", stdout.String())
	}
	if !strings.Contains(stderr.String(), "need review") {
		t.Errorf("expected stderr review summary, got:\n%s", stderr.String())
	}
}
```

(Match the test file's existing imports: `bytes`, `os`, `path/filepath`, `strings`, `testing`.)

- [ ] **Step 2: Run — expect FAIL** (`--migrate` is still an identity pass; `ExitMigrateReview` undefined).

Run: `go test ./cmd/dippin/ -run TestFmtMigrate`
Expected: build error / FAIL.

- [ ] **Step 3: Implement.** In `cmd/dippin/cli.go`, add the exit code after `ExitUsageError`:

```go
	ExitUsageError ExitCode = 2
	ExitMigrateReview ExitCode = 3 // `fmt --migrate` succeeded but flagged cases needing author review
```

Rewrite `cmd/dippin/cmd_fmt.go`'s `CmdFmt` to run the migration transform on v1 input. Replace `parseAndFormat`'s use inside `CmdFmt` with a parse-then-(migrate)-then-format flow:

```go
import (
	"flag"
	"fmt"
	"os"

	"github.com/2389-research/dippin-lang/formatter"
	"github.com/2389-research/dippin-lang/ir"
	"github.com/2389-research/dippin-lang/parser"
	"github.com/2389-research/dippin-lang/simulate"
)

func (c *CLI) CmdFmt(args []string) ExitCode {
	fs := flag.NewFlagSet("fmt", flag.ContinueOnError)
	fs.SetOutput(c.Stderr)
	check := fs.Bool("check", false, "exit 1 if not canonically formatted")
	write := fs.Bool("write", false, "write formatted output back to source file")
	migrate := fs.Bool("migrate", false, "convert a v1 file to dip 2 (edges own destinations)")
	if err := fs.Parse(args); err != nil {
		return ExitUsageError
	}
	if fs.NArg() < 1 {
		fmt.Fprintln(c.Stderr, "usage: dippin fmt [--check] [--write] [--migrate] <file>")
		return ExitUsageError
	}
	path := fs.Arg(0)
	w, data, code := c.parseFile(path)
	if code != ExitCode(-1) {
		return code
	}
	var notes []formatter.MigrationNote
	if *migrate {
		notes = c.migrateWorkflow(w)
	}
	formatted := formatter.Format(w)
	if ec := c.emitFmt(path, string(data), formatted, *check, *write, *migrate); ec != ExitOK {
		return ec
	}
	return c.reportMigrationNotes(notes)
}

// parseFile reads and parses a file. Returns (workflow, raw, -1) on success or
// (nil, nil, code) on failure.
func (c *CLI) parseFile(path string) (*ir.Workflow, []byte, ExitCode) {
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(c.Stderr, "error: %v\n", err)
		return nil, nil, ExitError
	}
	w, err := parser.NewParser(string(data), path).Parse()
	if err != nil {
		c.renderError(err, path)
		return nil, nil, ExitError
	}
	return w, data, ExitCode(-1)
}

// migrateWorkflow parses conditions then runs the v1->v2 transform.
func (c *CLI) migrateWorkflow(w *ir.Workflow) []formatter.MigrationNote {
	_ = simulate.EnsureConditionsParsed(w)
	return formatter.MigrateToV2(w)
}

// reportMigrationNotes prints a stderr summary of review cases and returns the
// review exit code when any exist.
func (c *CLI) reportMigrationNotes(notes []formatter.MigrationNote) ExitCode {
	if len(notes) == 0 {
		return ExitOK
	}
	fmt.Fprintf(c.Stderr, "migrated with %d case(s) that need review:\n", len(notes))
	for _, n := range notes {
		fmt.Fprintf(c.Stderr, "  node %q: %s\n", n.Node, n.Message)
	}
	return ExitMigrateReview
}
```

Delete the now-unused `parseAndFormat` **only if** nothing else calls it (`grep -rn parseAndFormat cmd/dippin/`); if `CmdFmt` was its sole caller, remove it, otherwise leave it. Keep `emitFmt`, `fmtCheck`, `boolToPath`, `writeOutput`, `renderError` unchanged. Update the `emitFmt` doc comment: `--migrate` now performs the real v1→v2 transform.

- [ ] **Step 4: Run — expect PASS** (new + full cmd suite; the non-migrate `fmt` path is unchanged since `parseFile`+`Format` reproduce the old behavior).

Run: `go test ./cmd/dippin/ -run 'TestFmtMigrate|TestFmt|TestRunFmt' -v && go test ./cmd/dippin/ -count=1`
Expected: PASS; `ok`.

- [ ] **Step 5: Complexity + lint.**

Run: `just complexity && golangci-lint run ./cmd/dippin/...`
Expected: `Complexity OK.`; `0 issues`.

- [ ] **Step 6: Commit.**

```bash
git add cmd/dippin/cli.go cmd/dippin/cmd_fmt.go cmd/dippin/cmd_fmt_test.go
git commit -m "feat(cmd): fmt --migrate converts v1 to dip 2 with review notes + exit code"
```

---

### Task 7: docs, cascade-ordering delta, example round-trip

**Files:**
- Modify: `docs/edges.md`
- Test: `formatter/migrate_v2_test.go` (add an examples round-trip test)

- [ ] **Step 1: Write the round-trip guard test.** Add to `formatter/migrate_v2_test.go`:

```go
func TestMigrate_ExamplesRoundTripToValidV2(t *testing.T) {
	matches, _ := filepath.Glob("../examples/*.dip")
	for _, path := range matches {
		src, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		w, err := parser.NewParser(string(src), path).Parse()
		if err != nil {
			t.Fatalf("%s: parse: %v", path, err)
		}
		_ = simulate.EnsureConditionsParsed(w)
		MigrateToV2(w)
		out := Format(w)
		// The migrated text must re-parse as a valid dip 2 file.
		if _, err := parser.NewParser(out, path).Parse(); err != nil {
			t.Errorf("%s: migrated v2 does not re-parse: %v\n%s", path, err, out)
		}
	}
}
```

(Add imports `os`, `path/filepath`, and `github.com/2389-research/dippin-lang/parser` to the test file.)

- [ ] **Step 2: Run — expect PASS** (the transform + formatter produce re-parseable v2; if any example fails, fix the transform, not the example).

Run: `go test ./formatter/ -run TestMigrate_ExamplesRoundTripToValidV2 -v`
Expected: PASS.

- [ ] **Step 3: Document the cascade-ordering delta.** In `docs/edges.md`, under the format-version / routing section, add a subsection:

```markdown
### dip 2: retry budgets vs failure destinations

Under `dip 2`, a node carries only retry **budgets** — `max_retries`, `base_delay`,
`retry_policy`. The `retry_target` and `fallback_target` node fields are removed:
`max_retries: N` means "re-run this node up to N times," and the post-exhaustion
destination is the node's `on fail` edge. Convert a v1 file with `dippin fmt
--migrate` (it rewrites the fields into edges and flags any case it cannot express
1:1).

**Runtime ordering note (spec delta):** in `dip 2` a node exhausts its
`max_retries` in place *before* control follows the `on fail` edge — the retry
budget is consumed first, then the failure edge is taken. This differs from the
v1 cascade where a matching `fail` edge preempts node retry. dippin ships the
syntax and IR; the enforcing runtime converges on this ordering.
```

- [ ] **Step 4: Regenerate the embedded spec + verify.**

Run: `bash scripts/gen-spec.sh && go test ./releasecheck/ -count=1`
Expected: `ok` (embedded spec current). (`docs/edges.md` feeds the spec only if referenced by `gen-spec.sh`'s sources — if `releasecheck` shows no change, that's fine; commit any regenerated `cmd/dippin/generated-spec.md`.)

- [ ] **Step 5: Commit.**

```bash
git add docs/edges.md formatter/migrate_v2_test.go cmd/dippin/generated-spec.md
git commit -m "docs(edges): document dip 2 retry-budget vs on-fail-edge model + ordering delta"
```

---

### Task 8: Final verification

- [ ] **Step 1: Full race suite + all gates.**

Run:
```bash
go test ./... -count=1 -race && just vet && just complexity && golangci-lint run && just validate-examples
```
Expected: all green.

- [ ] **Step 2: Manual smoke test — reject under v2, migrate a v1 file.**

Run:
```bash
printf 'dip 2\n\nworkflow W\n  goal: "t"\n  start: T\n  exit: D\n  tool T\n    command: run\n    fallback_target: D\n  agent D\n    prompt: d\n  edges\n    T -> D on success\n' > /tmp/v2reject.dip
go run ./cmd/dippin validate /tmp/v2reject.dip   # expect: rejection diagnostic
printf 'workflow W\n  goal: "t"\n  start: T\n  exit: D\n  tool T\n    command: run\n    fallback_target: Esc\n  agent Esc\n    prompt: e\n  agent D\n    prompt: d\n  edges\n    T -> D on success\n' > /tmp/v1.dip
go run ./cmd/dippin fmt --migrate /tmp/v1.dip    # expect: dip 2 output with `T -> Esc  on fail`
```
Expected: v2 file errors on `fallback_target`; the v1 file migrates to a `dip 2` file with a synthesized `on fail` edge.

- [ ] **Step 3: Push + PR** (only when the user asks — do not auto-merge).

```bash
git push -u origin feat/134-dip2-edges-own-destinations
gh pr create --base main --title "feat: dip 2 — edges own destinations (#134)" --body "Implements #134 per docs/superpowers/specs/2026-07-09-issue-134-dip2-edges-own-destinations-design.md."
```
