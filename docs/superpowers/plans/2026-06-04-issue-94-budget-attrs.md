# Budget/Limit Attrs (#94) Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make `stall_timeout` declarable in `.dip`, close a pre-existing DOT export data-loss bug for all four budget fields, add DIP145 (negative-budget lint), and document `max_turns` exhaustion semantics.

**Architecture:** A new `time.Duration` graph default `stall_timeout` lands in `ir.WorkflowDefaults` and flows through the four round-trip surfaces (parser → formatter → DOT export → migrate), exactly mirroring the existing `max_wall_time`. The DOT exporter currently emits *none* of the budget fields (a confirmed data-loss bug); we add a single `appendBudgetGraphAttrs` helper that emits all four. DIP145 is a table-driven range lint firing on any negative budget default. `max_turns` exhaustion is documentation-only (it reuses the #92/#93 failure cascade).

**Tech Stack:** Go; `just` for all build/test (never raw `go`); pre-commit hook runs the full gate (build, vet, golangci-lint, all tests w/ race, gocyclo ≤5, gocognit ≤7, example validation) on every commit.

**Spec:** `docs/superpowers/specs/2026-06-04-issue-94-budget-attrs-design.md`

**Conventions (read before starting):**
- All testing via `just test-pkg <pkg>` (e.g. `just test-pkg parser`) for fast iteration; `just complexity` for the cyclo/cognit gate; `just fmt` before committing. The `git commit` itself runs the full pre-commit gate — that is the real CI mirror.
- TDD: write the failing test, run it, watch it fail, then implement.
- Parser tests parse real `.dip` source — never hand-populate IR fields the parser doesn't set.
- Stage explicit paths in `git add` (never `git add -A` — `./dippin`/`./wasm` are gitignored build artifacts).
- Field name is `stall_timeout` everywhere (IR `StallTimeout`, parser/DOT key `stall_timeout`).

---

## Task 1: `stall_timeout` IR field + parser

**Files:**
- Modify: `ir/ir.go:52-54` (add field to `WorkflowDefaults`)
- Modify: `parser/parse_defaults.go:134-146` (`applyDefaultBudgetField`)
- Test: `parser/parser_test.go` (new test), `parser/testdata/defaults_budget.dip` (fixture)

- [ ] **Step 1: Add `stall_timeout` to the budget fixture**

Edit `parser/testdata/defaults_budget.dip`, adding one line to the `defaults` block (after `max_wall_time: 30m`):

```dippin
  defaults
    model: claude-sonnet-4-6
    max_total_tokens: 500000
    max_cost_cents: 1000
    max_wall_time: 30m
    stall_timeout: 5m
```

- [ ] **Step 2: Write the failing test**

Add to `parser/parser_test.go` (near the existing `TestParseDefaultsBudget` around line 1560):

```go
func TestParseDefaultsStallTimeout(t *testing.T) {
	w := parseFixture(t, "defaults_budget.dip")
	if w.Defaults.StallTimeout != 5*time.Minute {
		t.Errorf("stall_timeout = %v, want 5m0s", w.Defaults.StallTimeout)
	}
}

func TestParseDefaultsNegativeBudgetNoStructuralError(t *testing.T) {
	src := `workflow Neg
  goal: "g"
  start: A
  exit: A

  defaults
    max_cost_cents: -5
    stall_timeout: -5m

  agent A
    prompt: "Do it."

  edges
    A -> A
`
	w, err := NewParser(src, "neg.dip").Parse()
	if err != nil {
		t.Fatalf("negative budget must parse without a structural error, got: %v", err)
	}
	if w.Defaults.MaxCostCents != -5 {
		t.Errorf("max_cost_cents = %d, want -5", w.Defaults.MaxCostCents)
	}
	if w.Defaults.StallTimeout != -5*time.Minute {
		t.Errorf("stall_timeout = %v, want -5m0s", w.Defaults.StallTimeout)
	}
}
```

- [ ] **Step 3: Run the test to verify it fails**

Run: `just test-pkg parser`
Expected: FAIL — `w.Defaults.StallTimeout` undefined (field doesn't exist yet).

- [ ] **Step 4: Add the IR field**

In `ir/ir.go`, inside `WorkflowDefaults`, add `StallTimeout` immediately after `MaxWallTime` (line 54):

```go
	MaxWallTime       time.Duration // Hard ceiling on wall time
	StallTimeout      time.Duration // Abort/route when no progress for this wall-clock span (0 = disabled)
```

- [ ] **Step 5: Parse the field**

In `parser/parse_defaults.go`, add a case to `applyDefaultBudgetField` (before the `default:`):

```go
	case "max_wall_time":
		p.workflow.Defaults.MaxWallTime = p.parseDuration(val, key, loc)
	case "stall_timeout":
		p.workflow.Defaults.StallTimeout = p.parseDuration(val, key, loc)
```

- [ ] **Step 6: Run the test to verify it passes**

Run: `just test-pkg parser`
Expected: PASS.

- [ ] **Step 7: Check complexity**

Run: `just complexity`
Expected: passes. (`applyDefaultBudgetField` goes from cyclo 4 → 5, at the cap.)

- [ ] **Step 8: Commit**

```bash
just fmt
git add ir/ir.go parser/parse_defaults.go parser/parser_test.go parser/testdata/defaults_budget.dip
git commit -m "feat: parse stall_timeout graph default into ir.WorkflowDefaults (#94)"
```

---

## Task 2: Formatter emit + `.dip` round-trip + negative-duration quirk lock

**Files:**
- Modify: `formatter/format.go:192-204` (`writeDefaultsBudgetFields`)
- Test: `parser/parser_test.go` (extend `TestParseDefaultsBudgetRoundTrip`), `formatter/format_test.go` (new quirk test)

- [ ] **Step 1: Extend the round-trip test**

In `parser/parser_test.go`, add to `TestParseDefaultsBudgetRoundTrip` (after the `MaxWallTime` assertion at line 1644):

```go
	if d.StallTimeout != 5*time.Minute {
		t.Errorf("round-trip: stall_timeout = %v, want 5m0s", d.StallTimeout)
	}
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `just test-pkg parser`
Expected: FAIL — formatter drops `stall_timeout`, so the re-parsed value is 0.

- [ ] **Step 3: Emit `stall_timeout` in the formatter**

In `formatter/format.go`, `writeDefaultsBudgetFields`, add after the `MaxWallTime` block (line 201):

```go
	if d.MaxWallTime != 0 {
		wr.line("max_wall_time: %s", formatDuration(d.MaxWallTime))
	}
	if d.StallTimeout != 0 {
		wr.line("stall_timeout: %s", formatDuration(d.StallTimeout))
	}
```

- [ ] **Step 4: Run the test to verify it passes**

Run: `just test-pkg parser`
Expected: PASS.

- [ ] **Step 5: Write the negative-duration quirk-lock test**

`formatDuration(-5m)` returns `"-5m0s"` (the per-component `> 0` guards fall through to `d.String()`). This is NOT a bug — it re-parses losslessly. Lock it so nobody "fixes" it. Add to `formatter/format_test.go`:

```go
func TestFormatDurationNegativeRoundTrips(t *testing.T) {
	for _, d := range []time.Duration{-5 * time.Minute, -30 * time.Second} {
		s := formatDuration(d)
		got, err := time.ParseDuration(s)
		if err != nil {
			t.Fatalf("formatDuration(%v) = %q, not re-parseable: %v", d, s, err)
		}
		if got != d {
			t.Errorf("round-trip: formatDuration(%v) = %q -> %v, want %v", d, s, got, d)
		}
	}
}
```

(If `format_test.go` lacks a `time` import, add it.)

- [ ] **Step 6: Run and verify it passes**

Run: `just test-pkg formatter`
Expected: PASS.

- [ ] **Step 7: Check complexity + commit**

```bash
just complexity   # writeDefaultsBudgetFields: cyclo 4 -> 5, at cap
just fmt
git add formatter/format.go formatter/format_test.go parser/parser_test.go
git commit -m "feat: emit stall_timeout in formatter; lock negative-duration round-trip (#94)"
```

---

## Task 3: DOT export — close the gap for all four budget fields

**Files:**
- Modify: `export/dot.go:61-67` (`reservedGraphAttrs`), `:88-110` (`buildGraphAttrs` + new helper)
- Test: `export/dot_test.go`

- [ ] **Step 1: Write the failing export test**

In `export/dot_test.go`, add (mirroring `TestExportOnFailureDefault` at line 1445 and `TestExportToolSafetyVarsCollision` at 1412):

```go
func TestExportBudgetDefaults(t *testing.T) {
	w := &ir.Workflow{
		Name: "B", Start: "A", Exit: "A",
		Defaults: ir.WorkflowDefaults{
			MaxTotalTokens: 500000,
			MaxCostCents:   1000,
			MaxWallTime:    30 * time.Minute,
			StallTimeout:   5 * time.Minute,
		},
		Nodes: []*ir.Node{{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "x"}}},
		Edges: []*ir.Edge{{From: "A", To: "A"}},
	}
	out := ExportDOT(w, ExportOptions{})
	for _, want := range []string{
		`max_total_tokens="500000"`,
		`max_cost_cents="1000"`,
		`max_wall_time="30m"`,
		`stall_timeout="5m"`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("expected %s in DOT, got:\n%s", want, out)
		}
	}
}

func TestExportBudgetDefaultsOmitEmpty(t *testing.T) {
	w := &ir.Workflow{
		Name: "B", Start: "A", Exit: "A",
		Nodes: []*ir.Node{{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "x"}}},
		Edges: []*ir.Edge{{From: "A", To: "A"}},
	}
	out := ExportDOT(w, ExportOptions{})
	if strings.Contains(out, "stall_timeout") || strings.Contains(out, "max_total_tokens") {
		t.Errorf("unset budget fields must not be emitted, got:\n%s", out)
	}
}
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `just test-pkg export`
Expected: FAIL — `buildGraphAttrs` never emits budget fields.

- [ ] **Step 3: Add the budget-emit helper**

In `export/dot.go`, add a new helper (place it just after `buildGraphAttrs`, before `addGraphVarsAttrs` at line 112):

```go
// appendBudgetGraphAttrs emits the workflow budget defaults as DOT graph
// attributes. Each is omitted when unset (0) so absence round-trips. Durations
// use formatDuration (the compact Go-duration literal the migrate path parses
// back with time.ParseDuration), NOT time.Duration.String().
func appendBudgetGraphAttrs(attrs *[]string, d ir.WorkflowDefaults) {
	if d.MaxTotalTokens != 0 {
		*attrs = append(*attrs, fmt.Sprintf("max_total_tokens=%s", dotQuote(fmt.Sprintf("%d", d.MaxTotalTokens))))
	}
	if d.MaxCostCents != 0 {
		*attrs = append(*attrs, fmt.Sprintf("max_cost_cents=%s", dotQuote(fmt.Sprintf("%d", d.MaxCostCents))))
	}
	if d.MaxWallTime != 0 {
		*attrs = append(*attrs, fmt.Sprintf("max_wall_time=%s", dotQuote(formatDuration(d.MaxWallTime))))
	}
	if d.StallTimeout != 0 {
		*attrs = append(*attrs, fmt.Sprintf("stall_timeout=%s", dotQuote(formatDuration(d.StallTimeout))))
	}
}
```

- [ ] **Step 4: Call the helper from `buildGraphAttrs`**

In `export/dot.go` `buildGraphAttrs`, add the call after the `on_failure` block (line 104), before `addGraphVarsAttrs`:

```go
	if w.Defaults.OnFailure != "" {
		attrs = append(attrs, fmt.Sprintf("on_failure=%s", dotQuote(w.Defaults.OnFailure)))
	}
	appendBudgetGraphAttrs(&attrs, w.Defaults)

	// Add workflow vars (excluding reserved graph attributes)
	addGraphVarsAttrs(&attrs, w.Vars)
```

- [ ] **Step 5: Reserve `stall_timeout`**

In `export/dot.go` `reservedGraphAttrs` (line 65, where the three budget keys already are), add `stall_timeout`:

```go
	"max_total_tokens": true, "max_cost_cents": true, "max_wall_time": true,
	"stall_timeout": true,
```

- [ ] **Step 6: Run the test to verify it passes**

Run: `just test-pkg export`
Expected: PASS.

- [ ] **Step 7: Check complexity (the critical gate)**

Run: `just complexity`
Expected: PASS. `buildGraphAttrs` stays cyclo 5 (the four budget `if`s now live in `appendBudgetGraphAttrs`, which is cyclo 5). Had we inlined them, `buildGraphAttrs` would be cyclo 8 — over the limit.

- [ ] **Step 8: Commit**

```bash
just fmt
git add export/dot.go export/dot_test.go
git commit -m "fix: emit all four budget defaults in DOT export (close data-loss gap) (#94)"
```

---

## Task 4: Migrate `stall_timeout` + full `.dip → DOT → .dip` round-trip

**Files:**
- Modify: `migrate/migrate.go:185-196` (`applyIntBudgetDefault`)
- Test: `migrate/roundtrip_test.go`

- [ ] **Step 1: Write the failing round-trip test**

In `migrate/roundtrip_test.go`, add (mirroring `TestRoundTripToolSafetyDefaults` at line 220). The `90s` case is deliberate — it is the one input whose source string changes (`90s` → `1m30s`), so assert the typed `Duration`, never a string:

```go
// TestRoundTripBudgetDefaults verifies all four budget defaults round-trip
// losslessly through parse -> ExportDOT -> Migrate. Asserts Duration equality,
// never source-string equality (90s normalizes to 1m30s).
func TestRoundTripBudgetDefaults(t *testing.T) {
	src := `workflow Budget
  goal: "round trip"
  start: A
  exit: A

  defaults
    max_total_tokens: 500000
    max_cost_cents: 1000
    max_wall_time: 30m
    stall_timeout: 90s

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
	d := w2.Defaults
	if d.MaxTotalTokens != 500000 {
		t.Errorf("max_total_tokens = %d, want 500000; DOT:\n%s", d.MaxTotalTokens, dot)
	}
	if d.MaxCostCents != 1000 {
		t.Errorf("max_cost_cents = %d, want 1000; DOT:\n%s", d.MaxCostCents, dot)
	}
	if d.MaxWallTime != 30*time.Minute {
		t.Errorf("max_wall_time = %v, want 30m0s; DOT:\n%s", d.MaxWallTime, dot)
	}
	if d.StallTimeout != 90*time.Second {
		t.Errorf("stall_timeout = %v, want 1m30s; DOT:\n%s", d.StallTimeout, dot)
	}
}

// TestRoundTripStallTimeoutVarsCollision verifies a vars entry named
// stall_timeout is treated as reserved (not double-emitted / not migrated to Vars).
func TestRoundTripStallTimeoutVarsCollision(t *testing.T) {
	src := `workflow Collide
  goal: "round trip"
  start: A
  exit: A

  defaults
    stall_timeout: 5m

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
	if strings.Count(dot, "stall_timeout") != 1 {
		t.Errorf("stall_timeout should appear exactly once in DOT, got:\n%s", dot)
	}
	w2, err := Migrate(dot)
	if err != nil {
		t.Fatalf("migrate: %v", err)
	}
	if w2.Defaults.StallTimeout != 5*time.Minute {
		t.Errorf("stall_timeout = %v, want 5m0s", w2.Defaults.StallTimeout)
	}
	if _, ok := w2.Vars["stall_timeout"]; ok {
		t.Errorf("stall_timeout leaked into Vars: %v", w2.Vars)
	}
}
```

(Ensure `strings` and `time` are imported in `roundtrip_test.go`.)

- [ ] **Step 2: Run the test to verify it fails**

Run: `just test-pkg migrate`
Expected: FAIL — `stall_timeout` migrates into `Vars` (unknown attr) instead of `StallTimeout`, so `w2.Defaults.StallTimeout` is 0.

- [ ] **Step 3: Migrate `stall_timeout`**

In `migrate/migrate.go` `applyIntBudgetDefault`, add a case before `default:`:

```go
	case "max_wall_time":
		return tryApplyDurationDefault(v, &w.Defaults.MaxWallTime)
	case "stall_timeout":
		return tryApplyDurationDefault(v, &w.Defaults.StallTimeout)
```

- [ ] **Step 4: Run the test to verify it passes**

Run: `just test-pkg migrate`
Expected: PASS.

- [ ] **Step 5: Check complexity + commit**

```bash
just complexity   # applyIntBudgetDefault: cyclo 4 -> 5, at cap (switch)
just fmt
git add migrate/migrate.go migrate/roundtrip_test.go
git commit -m "feat: migrate stall_timeout from DOT; full budget round-trip test (#94)"
```

---

## Task 5: DIP145 lint — negative graph budget default

**Files:**
- Modify: `validator/lint_codes.go:3` (comment), `:48` (const), `:96` (CodeDescription)
- Modify: `validator/explanations.go:205-256` (`configExplanations`)
- Create: `validator/lint_budget.go`
- Modify: `validator/lint.go:63` (register)
- Test: `validator/lint_budget_test.go` (new)

- [ ] **Step 1: Write the failing lint test**

Create `validator/lint_budget_test.go`:

```go
package validator

import (
	"testing"
	"time"

	"github.com/2389-research/dippin-lang/ir"
)

func budgetWorkflow(d ir.WorkflowDefaults) *ir.Workflow {
	return &ir.Workflow{
		Name: "B", Start: "A", Exit: "A", Defaults: d,
		Nodes: []*ir.Node{{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "x"}}},
		Edges: []*ir.Edge{{From: "A", To: "A"}},
	}
}

func countCode(diags []Diagnostic, code string) int {
	n := 0
	for _, d := range diags {
		if d.Code == code {
			n++
		}
	}
	return n
}

func TestDIP145FiresOnNegativeBudgets(t *testing.T) {
	w := budgetWorkflow(ir.WorkflowDefaults{
		MaxTotalTokens: -1,
		MaxCostCents:   -5,
		MaxWallTime:    -30 * time.Second,
		StallTimeout:   -5 * time.Minute,
	})
	got := countCode(lintBudgetRanges(w), DIP145)
	if got != 4 {
		t.Errorf("DIP145 count = %d, want 4 (one per negative field)", got)
	}
}

func TestDIP145SilentOnUnsetZero(t *testing.T) {
	w := budgetWorkflow(ir.WorkflowDefaults{}) // all zero = unset
	if got := countCode(lintBudgetRanges(w), DIP145); got != 0 {
		t.Errorf("DIP145 fired on unset (0) budgets: %d, want 0", got)
	}
}

func TestDIP145SilentOnValidPositive(t *testing.T) {
	w := budgetWorkflow(ir.WorkflowDefaults{
		MaxTotalTokens: 500000,
		MaxCostCents:   1000,
		MaxWallTime:    30 * time.Minute,
		StallTimeout:   5 * time.Minute,
	})
	if got := countCode(lintBudgetRanges(w), DIP145); got != 0 {
		t.Errorf("DIP145 fired on valid positive budgets: %d, want 0", got)
	}
}

func TestDIP145MessageNamesFieldAndValue(t *testing.T) {
	w := budgetWorkflow(ir.WorkflowDefaults{MaxCostCents: -5})
	diags := lintBudgetRanges(w)
	if len(diags) != 1 {
		t.Fatalf("want 1 diag, got %d", len(diags))
	}
	msg := diags[0].Message
	if !strings.Contains(msg, "max_cost_cents") || !strings.Contains(msg, "-5") {
		t.Errorf("message must name field and value, got: %q", msg)
	}
}
```

(Add `"strings"` to the import block.)

- [ ] **Step 2: Run the test to verify it fails**

Run: `just test-pkg validator`
Expected: FAIL — `lintBudgetRanges` and `DIP145` undefined.

- [ ] **Step 3: Add the DIP145 const + description**

In `validator/lint_codes.go`:
- Line 3 comment: change `(DIP101–DIP144)` to `(DIP101–DIP145)`.
- After `DIP144` const (line 48), add:

```go
	DIP144 = "DIP144" // agent node has no failure route
	DIP145 = "DIP145" // graph budget default is negative
```

- After the `CodeDescription[DIP144]` line (96), add:

```go
	CodeDescription[DIP144] = "agent node has no failure route"
	CodeDescription[DIP145] = "graph budget default is negative"
```

- [ ] **Step 4: Add the Explanations entry**

In `validator/explanations.go`, inside `configExplanations()` (the map returned at line 205, alongside DIP116), add an entry before the closing brace:

```go
		DIP145: {
			Code:    DIP145,
			Summary: "graph budget default is negative",
			Trigger: "A workflow budget default (max_total_tokens, max_cost_cents, max_wall_time, or stall_timeout) is set to a negative value.",
			Fix:     "Use a positive cap, or omit the field / set 0 to mean no limit.",
			Example: "defaults\n  max_cost_cents: -5    // DIP145: negative; use a positive cap or 0 for no limit",
		},
```

- [ ] **Step 5: Implement the lint**

Create `validator/lint_budget.go`:

```go
package validator

import (
	"fmt"

	"github.com/2389-research/dippin-lang/ir"
)

// lintBudgetRanges checks DIP145: a graph-level budget default must not be
// negative. Zero means "unset / no limit" (the formatter emits only non-zero
// values), so only values < 0 are flagged. time.Duration is an int64, so all
// four fields normalize to int64 for a single uniform check.
func lintBudgetRanges(w *ir.Workflow) []Diagnostic {
	d := w.Defaults
	checks := []struct {
		name string
		v    int64
	}{
		{"max_total_tokens", int64(d.MaxTotalTokens)},
		{"max_cost_cents", int64(d.MaxCostCents)},
		{"max_wall_time", int64(d.MaxWallTime)},
		{"stall_timeout", int64(d.StallTimeout)},
	}
	var diags []Diagnostic
	for _, c := range checks {
		if c.v < 0 {
			diags = append(diags, Diagnostic{
				Code:     DIP145,
				Severity: SeverityWarning,
				Message:  fmt.Sprintf("workflow budget default %s is %d; budgets cannot be negative", c.name, c.v),
				Help:     "use a positive cap (e.g. max_cost_cents: 1000 for $10.00), or omit it / set 0 for no limit",
			})
		}
	}
	return diags
}
```

- [ ] **Step 6: Register the lint**

In `validator/lint.go`, after `lintAgentFailureRoute` (line 63):

```go
	diags = append(diags, lintAgentFailureRoute(w)...)
	diags = append(diags, lintBudgetRanges(w)...)
```

- [ ] **Step 7: Run the tests to verify they pass**

Run: `just test-pkg validator`
Expected: PASS — including `TestExplanationsCoverAllCodes` / `TestExplanationsNoExtra` (the new const + Explanations entry now agree).

- [ ] **Step 8: Check complexity + commit**

```bash
just complexity   # lintBudgetRanges: cyclo 2, well under
just fmt
git add validator/lint_codes.go validator/explanations.go validator/lint_budget.go validator/lint.go validator/lint_budget_test.go
git commit -m "feat: DIP145 lint for negative graph budget defaults (#94)"
```

---

## Task 6: Docs — `max_turns` exhaustion, budget fields, DIP145 section, count bumps

**Files:**
- Modify: `docs/nodes.md:116` (max_turns row + subsection)
- Modify: `docs/edges.md` (cascade failure sources)
- Modify: `docs/syntax.md` (defaults table — add four budget rows)
- Modify: `docs/validation.md` (DIP145 section + DIP144 cross-link + counts)
- Modify: `docs/llm-reference.md` (counts + range bullet)
- Modify: `CLAUDE.md:85` (count)

- [ ] **Step 1: Rewrite the `max_turns` row in `docs/nodes.md:116`**

Replace the existing `max_turns` table row with:

```text
| `max_turns` | Integer | 1 | Maximum request-response cycles in the agent's tool-using loop. **Reaching this limit ends the node with outcome `fail`** — it is a hard cap, not a soft budget. The failure routes through the standard failure cascade (fail edge → bounded retry → `fallback_target` → graph `on_failure` → halt). Ensure a failure route exists (see DIP144) or the run halts on exhaustion. |
```

- [ ] **Step 2: Add a `max_turns exhaustion` subsection**

After the agent-config table in `docs/nodes.md` (near the existing `goal_gate`/`auto_status` prose subsections), add:

```markdown
### max_turns exhaustion

When an agent reaches `max_turns` without completing, the engine treats it as a
failure (`ctx.outcome = fail`), **not** a successful stop. `max_turns` is therefore
a routing event, not just a cost control. An agent with `max_turns` set but no
failure route is a latent dead-end — [DIP144](validation.md#dip144) warns you. Pair
every bounded `max_turns` with one of: a `when ctx.outcome = fail` edge,
`fallback_target`, a bounded `retry_target`, or a graph `on_failure`.
```

- [ ] **Step 3: Add cascade failure sources in `docs/edges.md`**

In the Routing Priority / Failure Handling section (around line 110), add a sentence listing the failure sources that feed the cascade:

```markdown
A node enters the failure cascade when it errors, fails a `goal_gate`, **exhausts
`max_turns`**, or trips the graph **`stall_timeout`**. All route through the same
priority order below.
```

- [ ] **Step 4: Add the four budget rows to the `docs/syntax.md` defaults table**

In the `defaults` table (around line 113-124), add (the three existing budget fields are currently undocumented here — this fixes that too):

```text
| `max_total_tokens` | Integer | Hard ceiling on total tokens across the run. `0`/unset = no limit. |
| `max_cost_cents` | Integer | Hard ceiling on total cost, in **US cents** (e.g. `1000` = $10.00). `0`/unset = no limit. |
| `max_wall_time` | Duration | Hard ceiling on **wall-clock** run time (e.g. `30m`, `2h`). `0`/unset = no limit. |
| `stall_timeout` | Duration | **Wall-clock** span with no forward progress before the run aborts and routes through `on_failure` (e.g. `5m`, `90s`). Elapsed time, **not** a turn count. `0`/unset = disabled. |
```

Add one prose line below the table:

```markdown
All budget fields use `0` (or unset) to mean **no limit** — `0` does not mean
"zero budget." The three `max_*` fields bound *totals* (monotonic ceilings);
`stall_timeout` bounds *inactivity* (a sliding timer).
```

- [ ] **Step 5: Add the DIP145 section to `docs/validation.md`**

After the DIP144 section (ends ~line 1029), add a new section mirroring the DIP144/DIP116 structure:

```markdown
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
```markdown

- [ ] **Step 6: Cross-link DIP144 → max_turns**

In the DIP144 section of `docs/validation.md` (~line 987-1029), add to the trigger prose:

```markdown
This warning is especially important for agents with a bounded `max_turns`: turn
exhaustion ends the node as `fail`, so without a failure route the run halts on
exhaustion.
```

- [ ] **Step 7: Bump all count + range strings**

Update each (grep both en-dash `–` and hyphen `-` forms of `DIP101-DIP144`):
- `CLAUDE.md:85`: `53 diagnostic codes` → `54 diagnostic codes`; `DIP101-DIP144` → `DIP101-DIP145`.
- `docs/llm-reference.md:188`: `53 diagnostic codes` → `54`; line 191 bullet `DIP101–DIP144` → `DIP101–DIP145`, and append `, negative budget defaults` to the warning list.
- `docs/validation.md`: line 3 `registers 53` → `54`, `documents 48` → `49`; lines 6, 15, 227, 1049 `DIP101–DIP144` → `DIP101–DIP145`.

Run to find any missed occurrences:

```bash
grep -rn "DIP101.DIP144\|53 diagnostic\|registers 53\|documents 48" CLAUDE.md docs/ validator/
```

Expected after edits: no stale `DIP144`-range or `53`/`48` count strings remain (except historical references in `docs/superpowers/` specs, which are dated snapshots — leave those).

- [ ] **Step 8: Regenerate the assembled spec**

Run: `just spec-check` (or the generation step it wraps). If it reports the generated files are stale, run the generator (`scripts/gen-spec.sh`) and stage `docs/generated-spec.md` + `cmd/dippin/generated-spec.md`. NEVER hand-edit the generated files.

Note: the `git commit` pre-commit hook auto-runs gen-spec and will fail if the generated files are stale — so if unsure, just attempt the commit and let the hook regenerate.

- [ ] **Step 9: Commit**

```bash
just fmt
git add CLAUDE.md docs/nodes.md docs/edges.md docs/syntax.md docs/validation.md docs/llm-reference.md docs/generated-spec.md cmd/dippin/generated-spec.md
git commit -m "docs: max_turns exhaustion semantics, budget attrs, DIP145, count bumps (#94)"
```

---

## Task 7: New lint-clean example exercising the budget attrs

**Files:**
- Create: `examples/budget_guards.dip`

- [ ] **Step 1: Write the example**

Create `examples/budget_guards.dip` — a cost-ceilinged research+build flow with a stall window and a graph `on_failure` escalation, so the budget abort has somewhere to route (and DIP144 stays suppressed). Adjust node/edge syntax to match a known-good existing example (copy the header/shape from `examples/on_failure_route.dip`):

```dippin
workflow BudgetGuards
  goal: "Research and draft with a cost ceiling and a stall guard"
  start: Research
  exit: Done

  defaults
    model: claude-sonnet-4-6
    max_total_tokens: 500000   # hard ceiling: total tokens across the run; 0 = no limit
    max_cost_cents: 2000       # hard ceiling: $20.00 total spend; 0 = no limit
    max_wall_time: 30m         # hard ceiling: wall-clock; 0 = no limit
    stall_timeout: 5m          # wall-clock; abort if no progress for 5 minutes
    on_failure: Escalate       # budget / stall / max_turns failures route here

  agent Research
    prompt: "Research the topic and gather sources."
    max_turns: 8               # exhaustion ends the node as fail -> routes to on_failure

  agent Draft
    prompt: "Draft the document from the research."
    max_turns: 12

  human Escalate
    mode: freeform
    prompt: "The run hit a budget/stall/turn limit. How should we proceed?"

  agent Done
    prompt: "Finalize."

  edges
    Research -> Draft
    Draft -> Done
    Escalate -> Done
```

- [ ] **Step 2: Validate it parses + is lint-clean**

Run:
```bash
just build
./dippin validate examples/budget_guards.dip
./dippin lint examples/budget_guards.dip
```
Expected: validate passes; lint emits **no DIP144** (the graph `on_failure` suppresses it) and **no DIP145** (all budgets positive). If the example node/edge syntax errors, fix it against `examples/on_failure_route.dip` until clean. Minor warnings on unrelated codes are acceptable but prefer zero.

- [ ] **Step 3: Confirm the example suite still passes**

Run: `just validate-examples` and `just lint-examples`
Expected: all examples validate; no new failures.

- [ ] **Step 4: Commit**

```bash
git add examples/budget_guards.dip
git commit -m "docs: add lint-clean budget_guards example (#94)"
```

---

## Task 8: Full verification + wasm gate

**Files:** none (verification only)

- [ ] **Step 1: Run the test suite with race**

Run: `just test-race`
Expected: all packages PASS.

- [ ] **Step 2: Run the complexity gate**

Run: `just complexity`
Expected: PASS (no function over cyclo 5 / cognit 7).

- [ ] **Step 3: Verify the wasm build (Netlify preview gate)**

Run: `GOOS=js GOARCH=wasm go build ./cmd/wasm/ ./validator/`
Expected: builds with no error. (`lint_budget.go` imports only `ir`+`fmt`, wasm-safe.)

(This raw `go build` is the one exception the CI/Netlify gate uses; there is no `just` recipe for the cross-compile check.)

- [ ] **Step 4: Confirm no stale count strings**

Run: `grep -rn "DIP101.DIP144\|53 diagnostic\|registers 53\|documents 48" CLAUDE.md docs/ validator/ cmd/`
Expected: only dated `docs/superpowers/` snapshots remain (if any); no live docs/code.

- [ ] **Step 5: Final clean status check**

Run: `git status --short`
Expected: clean (all work committed; no stray `./dippin`/`./wasm` artifacts staged).

---

## Self-Review (completed against the spec)

- **stall_timeout field** → Task 1 (IR+parser), Task 2 (formatter), Task 3 (DOT export), Task 4 (migrate). ✅
- **max_turns exhaustion semantics** → Task 6 (docs-only, nodes.md + edges.md + DIP144 cross-link). ✅
- **DOT round-trip gap (all 4 fields)** → Task 3 (export) + Task 4 (migrate round-trip test). ✅
- **DIP145** → Task 5 (const, description, explanation, lint, register, tests). ✅
- **0=unset boundary** → Task 5 (`v < 0`) + tests `TestDIP145SilentOnUnsetZero`. ✅
- **wire-format contract (Go duration literal)** → Task 3 helper uses `formatDuration`; Task 4 asserts Duration equality incl. the `90s` case. ✅
- **per-node deferred** → not implemented (correct; documented in spec Out of scope). ✅
- **docs (syntax.md/nodes.md/edges.md/validation.md/llm-reference.md) + count bumps** → Task 6. ✅
- **new lint-clean example** → Task 7. ✅
- **wasm gate** → Task 8. ✅
- **var-collision reserved guard** → Task 3 (reserved) + Task 4 (`TestRoundTripStallTimeoutVarsCollision`). ✅
- **negative-duration format quirk** → Task 2 (`TestFormatDurationNegativeRoundTrips`). ✅
