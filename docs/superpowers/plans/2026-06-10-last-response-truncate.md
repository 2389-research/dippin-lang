# `last_response_truncate:` Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a carry-only `last_response_truncate:` integer attribute (per-agent, with parallel-branch override) that dippin parses, carries through every representation, and lints — while a downstream runtime enforces the actual truncation.

**Architecture:** `last_response_truncate: N` caps (at the runtime) the auto-injected previous response in an agent's prompt to N Unicode characters. It is a per-agent `AgentConfig` field plus a `BranchConfig` override (branch `0` = inherit), mirroring the existing `tool_access` / `output_limit` carry pattern across parser → IR → formatter → DOT export → migrate import/parity. A new lint code **DIP148** (Warning) flags a negative value, mirroring DIP145. It deliberately does **not** interact with DIP147 (a mitigated sink still emits the chain-attack Hint).

**Tech Stack:** Go; `just` task runner (never raw `go`); parser-driven tests; pre-commit hook is the real CI gate.

**Spec:** `docs/superpowers/specs/2026-06-10-issue-56-last-response-truncate-design.md`

**Worktree / branch:** `.claude/worktrees/56-last-response-truncate` on `feat/56-last-response-truncate` (already created off `origin/main` @ 1cee9ce; the spec is already committed there as 45c2ee7).

---

## Conventions for every task

- Run all builds/tests via `just`, never raw `go`. Key commands: `just test-pkg <pkg>` (one package), `just test` (all), `just lint-go`, `just complexity` (cyclo ≤ 5 / cognit ≤ 7), `just fmt`, `just spec-check`.
- Work entirely inside the worktree `.claude/worktrees/56-last-response-truncate`. Stage **explicit paths** (never `git add -A`).
- Commit trailer on every commit:
  ```text
  Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>
  ```
- `0` / unset always means "no truncation" for an agent; `0` on a branch means "inherit the agent's value". The formatter/exporter emit the field only when `> 0`, so a negative value never round-trips — it is a lint concern only.

---

## Task 1: IR fields + parser (agent and branch)

**Files:**
- Modify: `ir/ir.go` (add field to `AgentConfig` ~line 122 and `BranchConfig` ~line 180)
- Modify: `parser/parse_nodes.go` (`applyAgentParsedField` ~line 408; `applyBranchFieldChecked` ~line 945)
- Test: `parser/parser_test.go`

- [ ] **Step 1: Write the failing parser tests**

Add to `parser/parser_test.go`:

```go
func TestParse_AgentLastResponseTruncate(t *testing.T) {
	src := `workflow X
  start: A
  exit: A

  agent A
    prompt: "x"
    last_response_truncate: 4096
`
	w, err := NewParser(src, "test.dip").Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	acfg := w.Nodes[0].Config.(ir.AgentConfig)
	if acfg.LastResponseTruncate != 4096 {
		t.Errorf("last_response_truncate = %d, want 4096", acfg.LastResponseTruncate)
	}
}

func TestParse_BranchLastResponseTruncate(t *testing.T) {
	src := `workflow X
  start: P
  exit: W

  agent W
    prompt: "w"

  parallel P
    branch
      target: W
      last_response_truncate: 2048
`
	w, err := NewParser(src, "test.dip").Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	pcfg := w.Nodes[0].Config.(ir.ParallelConfig)
	if len(pcfg.Branches) != 1 || pcfg.Branches[0].LastResponseTruncate != 2048 {
		t.Errorf("branch last_response_truncate = %+v, want 2048", pcfg.Branches)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `just test-pkg parser`
Expected: FAIL — `acfg.LastResponseTruncate` / `pcfg.Branches[0].LastResponseTruncate` undefined (field does not exist yet).

- [ ] **Step 3: Add the IR fields**

In `ir/ir.go`, in `AgentConfig` (immediately after the `WritablePaths []string` field, before `Params`):

```go
	// LastResponseTruncate caps, at the runtime, the number of Unicode
	// characters of the auto-injected previous response ("last response") that
	// this agent receives in its prompt. 0 / unset = no truncation (full
	// response injected). A chain-attack mitigation (issue #56): it bounds how
	// much potentially-tainted upstream output reaches a privileged prompt.
	// dippin carries + lints (DIP148 flags a negative value); the runtime
	// enforces the truncation. Inert until a runtime reads it.
	LastResponseTruncate int
```

In `ir/ir.go`, in `BranchConfig` (after the `WritablePaths []string` override field):

```go
	// LastResponseTruncate is a per-branch override of the target agent's
	// last_response_truncate. 0 INHERITS the target agent's value (never resets
	// to "no truncation") — the runtime resolves effective = branch if > 0 else
	// agent, mirroring ToolAccess / WritablePaths inheritance. See issue #56.
	LastResponseTruncate int
```

- [ ] **Step 4: Add the agent parser case**

In `parser/parse_nodes.go`, in `applyAgentParsedField`'s switch, add a case alongside `max_turns`:

```go
	case "last_response_truncate":
		cfg.LastResponseTruncate = p.parseInt(val, key, loc)
```

- [ ] **Step 5: Add the branch parser handling**

In `parser/parse_nodes.go`, in `applyBranchFieldChecked`, intercept the int field before the string setter map (it needs `p.parseInt`, which the `func(*ir.BranchConfig, string)` setter map cannot call):

```go
func (p *Parser) applyBranchFieldChecked(bc *ir.BranchConfig, key, val string, loc ir.SourceLocation) {
	if p.rejectEmptyWritablePaths(key, val, loc) {
		return
	}
	if key == "last_response_truncate" {
		bc.LastResponseTruncate = p.parseInt(val, key, loc)
		return
	}
	if !applyBranchField(bc, key, val) {
		p.emitUnknownFieldHint("branch", key, loc)
	}
}
```

- [ ] **Step 6: Run tests to verify they pass**

Run: `just test-pkg parser`
Expected: PASS (both new tests green; existing parser tests still green).

- [ ] **Step 7: Commit**

```bash
git add ir/ir.go parser/parse_nodes.go parser/parser_test.go
git commit -m "feat(#56): parse last_response_truncate on agents and branches

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

## Task 2: Formatter (agent and branch round-trip)

**Files:**
- Modify: `formatter/format.go` (`writeAgentRuntimeFields` ~line 419; `branchHasFields` ~line 298; `writeBranchFields` ~line 304)
- Test: `formatter/format_test.go`

- [ ] **Step 1: Write the failing round-trip test**

Add to `formatter/format_test.go`:

```go
func TestFormat_LastResponseTruncate_RoundTrip(t *testing.T) {
	src := `workflow X
  start: P
  exit: W

  agent W
    prompt: "w"
    last_response_truncate: 4096

  parallel P
    branch
      target: W
      last_response_truncate: 2048
`
	w, err := parser.NewParser(src, "test.dip").Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	out := Format(w)
	if !strings.Contains(out, "last_response_truncate: 4096") {
		t.Errorf("agent last_response_truncate not emitted:\n%s", out)
	}
	if !strings.Contains(out, "last_response_truncate: 2048") {
		t.Errorf("branch last_response_truncate not emitted:\n%s", out)
	}
	// Re-parse the formatted output: the values must survive.
	w2, err := parser.NewParser(out, "test.dip").Parse()
	if err != nil {
		t.Fatalf("re-parse error: %v\n%s", err, out)
	}
	if w2.Nodes[0].Config.(ir.AgentConfig).LastResponseTruncate != 4096 {
		t.Errorf("agent value lost on round-trip")
	}
}

func TestFormat_LastResponseTruncate_OmittedWhenZero(t *testing.T) {
	w := &ir.Workflow{
		Name: "X", Start: "A", Exit: "A",
		Nodes: []*ir.Node{{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "x"}}},
		Edges: []*ir.Edge{{From: "A", To: "A"}},
	}
	if strings.Contains(Format(w), "last_response_truncate") {
		t.Errorf("last_response_truncate emitted for zero value")
	}
}
```

(Confirm the existing import block of `format_test.go` already imports `strings`, `ir`, and `parser`; the file's other tests use all three, so no import change is expected.)

- [ ] **Step 2: Run test to verify it fails**

Run: `just test-pkg formatter`
Expected: FAIL — formatted output does not contain `last_response_truncate`.

- [ ] **Step 3: Emit the agent field**

In `formatter/format.go`, in `writeAgentRuntimeFields`, after the `writable_paths` block:

```go
	if cfg.LastResponseTruncate > 0 {
		wr.line("last_response_truncate: %d", cfg.LastResponseTruncate)
	}
```

- [ ] **Step 4: Emit the branch field**

In `formatter/format.go`, extend `branchHasFields` to include the new field:

```go
func branchHasFields(b ir.BranchConfig) bool {
	return b.Model != "" || b.Provider != "" || b.Fidelity != "" ||
		b.ToolAccess != "" || len(b.WritablePaths) > 0 || b.LastResponseTruncate > 0
}
```

and in `writeBranchFields`, after the `writable_paths` block:

```go
	if b.LastResponseTruncate > 0 {
		wr.line("last_response_truncate: %d", b.LastResponseTruncate)
	}
```

- [ ] **Step 5: Run test to verify it passes**

Run: `just test-pkg formatter`
Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add formatter/format.go formatter/format_test.go
git commit -m "feat(#56): format last_response_truncate on agents and branches

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

## Task 3: DOT export + migrate import (round-trip)

**Files:**
- Modify: `export/dot.go` (`applyAgentRuntimeAttrs` ~line 326; `encodeBranch` ~line 455; add `appendBranchIntField` helper)
- Modify: `migrate/migrate.go` (`applyRuntimeSafetyAttrs` ~line 435; `branchFieldSetters` ~line 673)
- Test: `export/dot_test.go`, `migrate/migrate_test.go`

- [ ] **Step 1: Write the failing DOT export test**

Add to `export/dot_test.go` (mirror `TestExportDOT_AgentToolAccess` ~line 350):

```go
func TestExportDOT_AgentLastResponseTruncate(t *testing.T) {
	w := &ir.Workflow{
		Name: "X", Start: "A", Exit: "A",
		Nodes: []*ir.Node{{
			ID: "A", Kind: ir.NodeAgent,
			Config: ir.AgentConfig{Prompt: "x", LastResponseTruncate: 4096},
		}},
		Edges: []*ir.Edge{{From: "A", To: "A"}},
	}
	out := ExportDOT(w, Options{})
	if !strings.Contains(out, `last_response_truncate="4096"`) &&
		!strings.Contains(out, `last_response_truncate=4096`) {
		t.Errorf("DOT output missing last_response_truncate attribute:\n%s", out)
	}
}
```

(Confirm the exact exporter entry point and `Options` struct by reading the top of `export/dot_test.go` — use whatever signature the sibling `TestExportDOT_AgentToolAccess` uses; the snippet above assumes `ExportDOT(w, Options{})`.)

- [ ] **Step 2: Run test to verify it fails**

Run: `just test-pkg export`
Expected: FAIL — attribute absent from DOT output.

- [ ] **Step 3: Emit the agent DOT attribute**

In `export/dot.go`, in `applyAgentRuntimeAttrs`, after the `writable_paths` block (mirrors `output_limit` at `applyToolOutputsAttrs`):

```go
	if cfg.LastResponseTruncate > 0 {
		attrs["last_response_truncate"] = strconv.Itoa(cfg.LastResponseTruncate)
	}
```

(`strconv` is already imported in `dot.go` — it is used by `applyToolOutputsAttrs`.)

- [ ] **Step 4: Emit the branch DOT token**

In `export/dot.go`, add an int-valued sibling to `appendBranchField` (which only handles non-empty strings):

```go
// appendBranchIntField appends key=value only when val > 0.
func appendBranchIntField(parts []string, key string, val int) []string {
	if val <= 0 {
		return parts
	}
	return append(parts, key+"="+encodeBranchToken(strconv.Itoa(val)))
}
```

and call it at the end of `encodeBranch`, after the `writable_paths` line:

```go
	parts = appendBranchIntField(parts, "last_response_truncate", b.LastResponseTruncate)
```

- [ ] **Step 5: Write the failing migrate read-back test**

Add to `migrate/migrate_test.go` a DOT→IR round-trip test. First read an existing migrate test (e.g. one exercising `tool_access` or `output_limit` read-back) to copy its exact harness (function names like `dotToWorkflow` / `Migrate` differ — use the real one). The assertion to add:

```go
func TestMigrate_AgentLastResponseTruncate_RoundTrip(t *testing.T) {
	w := &ir.Workflow{
		Name: "X", Start: "A", Exit: "A",
		Nodes: []*ir.Node{{
			ID: "A", Kind: ir.NodeAgent,
			Config: ir.AgentConfig{Prompt: "x", LastResponseTruncate: 4096},
		}},
		Edges: []*ir.Edge{{From: "A", To: "A"}},
	}
	// <export to DOT, then import back to IR using the harness the sibling
	//  output_limit/tool_access round-trip test uses>
	got := /* imported */ .Nodes[0].Config.(ir.AgentConfig).LastResponseTruncate
	if got != 4096 {
		t.Errorf("last_response_truncate lost on DOT round-trip: got %d, want 4096", got)
	}
}
```

(Implementation note: locate the existing `tool_access` or `output_limit` round-trip test in `migrate/*_test.go` and clone its export+import calls verbatim, swapping the asserted field. Do not invent a harness.)

- [ ] **Step 6: Run tests to verify they fail**

Run: `just test-pkg export` then `just test-pkg migrate`
Expected: export PASS after Step 3-4; migrate FAIL (read-back not implemented).

- [ ] **Step 7: Implement the migrate read-back (agent)**

In `migrate/migrate.go`, in `applyRuntimeSafetyAttrs`, after the `writable_paths` block (mirrors `applyMaxTurns`):

```go
	if v, ok := attrs["last_response_truncate"]; ok {
		if n, err := strconv.Atoi(v); err == nil {
			cfg.LastResponseTruncate = n
		}
	}
```

(`strconv` is already imported in `migrate.go` — used by `applyMaxTurns`.)

- [ ] **Step 8: Implement the migrate read-back (branch)**

In `migrate/migrate.go`, add to the `branchFieldSetters` map (~line 673). The setter signature is `func(*ir.BranchConfig, string)`; migrate has no diagnostic channel, so parse with `Atoi` and ignore the error (mirrors `applyMaxTurns`):

```go
	"last_response_truncate": func(b *ir.BranchConfig, v string) {
		if n, err := strconv.Atoi(v); err == nil {
			b.LastResponseTruncate = n
		}
	},
```

- [ ] **Step 9: Run tests to verify they pass**

Run: `just test-pkg export` and `just test-pkg migrate`
Expected: PASS.

- [ ] **Step 10: Commit**

```bash
git add export/dot.go export/dot_test.go migrate/migrate.go migrate/migrate_test.go
git commit -m "feat(#56): carry last_response_truncate through DOT export + migrate import

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

## Task 4: Migrate parity diff (agent and branch)

**Files:**
- Modify: `migrate/parity.go` (`compareAgentBehavior` ~line 262; branch comparison via `compareParallelBranches` ~line 371)
- Test: `migrate/parity_test.go`

- [ ] **Step 1: Write the failing parity tests**

Add to `migrate/parity_test.go` (mirror `TestParity_AgentWritablePathsDiff` ~line 9):

```go
func TestParity_AgentLastResponseTruncateDiff(t *testing.T) {
	a := ir.AgentConfig{LastResponseTruncate: 4096}
	b := ir.AgentConfig{LastResponseTruncate: 2048}
	diffs := compareAgentBehavior("A", a, b)
	if len(diffs) == 0 {
		t.Error("expected a last_response_truncate diff, got none")
	}
}

func TestParity_AgentLastResponseTruncateEqual(t *testing.T) {
	a := ir.AgentConfig{LastResponseTruncate: 4096}
	b := ir.AgentConfig{LastResponseTruncate: 4096}
	if diffs := compareAgentBehavior("A", a, b); len(diffs) != 0 {
		t.Errorf("expected no diff for equal values, got %v", diffs)
	}
}
```

(For the branch parity, first read `compareParallelBranches` and the per-branch comparator it calls to find the exact function name, then add a `TestParity_BranchLastResponseTruncateDiff` mirroring `TestParity_BranchWritablePathsDiff` ~line 18.)

- [ ] **Step 2: Run tests to verify they fail**

Run: `just test-pkg migrate`
Expected: FAIL — no diff produced for differing `last_response_truncate`.

- [ ] **Step 3: Implement the agent parity diff**

In `migrate/parity.go`, in `compareAgentBehavior`, after the `writable_paths` diff (mirrors the `tool_access` diff):

```go
	if ac.LastResponseTruncate != bc.LastResponseTruncate {
		diffs = append(diffs, fieldDiff(id, "last_response_truncate",
			fmt.Sprintf("node %q last_response_truncate: %d vs %d", id, ac.LastResponseTruncate, bc.LastResponseTruncate)))
	}
```

- [ ] **Step 4: Implement the branch parity diff**

Read the per-branch comparator invoked by `compareParallelBranches` (~line 371). Add a `LastResponseTruncate` comparison mirroring how that comparator already diffs `ToolAccess` / `WritablePaths`. If branch fields are compared inline, add:

```go
	if a.LastResponseTruncate != b.LastResponseTruncate {
		diffs = append(diffs, fieldDiff(id, "branch.last_response_truncate",
			fmt.Sprintf("node %q branch last_response_truncate: %d vs %d", id, a.LastResponseTruncate, b.LastResponseTruncate)))
	}
```

(Match the surrounding variable names and `fieldDiff` signature exactly as the existing branch comparisons use them.)

- [ ] **Step 5: Run tests to verify they pass**

Run: `just test-pkg migrate`
Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add migrate/parity.go migrate/parity_test.go
git commit -m "feat(#56): migrate parity diff for last_response_truncate

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

## Task 5: DIP148 — negative `last_response_truncate` lint

**Files:**
- Create: `validator/lint_last_response_truncate.go`
- Create: `validator/lint_last_response_truncate_test.go`
- Modify: `validator/lint_codes.go` (const block ~line 54; `init()` CodeDescription ~line 105; range comment ~line 3)
- Modify: `validator/explanations.go` (`safetyExplanations` map ~line 478, after DIP147)
- Modify: `validator/lint.go` (dispatch ~line 79)

- [ ] **Step 1: Write the failing lint tests**

Create `validator/lint_last_response_truncate_test.go`:

```go
package validator

import "testing"

func TestLint_DIP148_FiresOnNegativeAgentValue(t *testing.T) {
	src := `workflow X
  start: A
  exit: A

  agent A
    prompt: "x"
    last_response_truncate: -1
`
	if !hasCode(lintSrc(t, src), DIP148) {
		t.Errorf("expected DIP148, got: %v", codes(lintSrc(t, src)))
	}
}

func TestLint_DIP148_FiresOnNegativeBranchValue(t *testing.T) {
	src := `workflow X
  start: P
  exit: W

  agent W
    prompt: "w"

  parallel P
    branch
      target: W
      last_response_truncate: -5
`
	if !hasCode(lintSrc(t, src), DIP148) {
		t.Errorf("expected DIP148 for negative branch value, got: %v", codes(lintSrc(t, src)))
	}
}

func TestLint_DIP148_SilentOnZeroAndPositive(t *testing.T) {
	for _, v := range []string{"0", "4096"} {
		src := `workflow X
  start: A
  exit: A

  agent A
    prompt: "x"
    last_response_truncate: ` + v + "\n"
		if hasCode(lintSrc(t, src), DIP148) {
			t.Errorf("DIP148 should not fire for value %q", v)
		}
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `just test-pkg validator`
Expected: FAIL — `DIP148` undefined (compile error).

- [ ] **Step 3: Register the DIP148 code**

In `validator/lint_codes.go`: add the const after `DIP147` (line ~54):

```go
	DIP148 = "DIP148" // last_response_truncate is negative
```

add to `init()` after the `DIP147` CodeDescription (line ~105):

```go
	CodeDescription[DIP148] = "last_response_truncate is negative"
```

and bump the range comment at the top of the file (line 3): `(DIP101–DIP147)` → `(DIP101–DIP148)`.

- [ ] **Step 4: Add the explanation (required — `TestExplanationsCoverAllCodes` enforces it)**

In `validator/explanations.go`, in `safetyExplanations`'s returned map, after the `DIP147` entry (before the closing `}` at ~line 478):

```go
		DIP148: {
			Code:    DIP148,
			Summary: "last_response_truncate is negative",
			Trigger: "An agent node or a per-branch override sets last_response_truncate to a negative value. The field caps how many Unicode characters of the auto-injected previous response the runtime injects into the agent's prompt — a chain-attack mitigation (issue #56). A negative cap is meaningless; 0 (or unset) means no truncation.",
			Fix:     "Use a non-negative character count (e.g. last_response_truncate: 4096), or omit the field / set 0 for no truncation. dippin carries + lints this value; a runtime enforces the truncation.",
			Example: "agent Writer\n  prompt: \"write the report\"\n  last_response_truncate: -1   // DIP148: negative; use a non-negative count or omit",
		},
```

Also update the `safetyExplanations` doc-comment range (`DIP138–DIP147` → `DIP138–DIP148`) at ~line 420.

- [ ] **Step 5: Implement the lint**

Create `validator/lint_last_response_truncate.go`:

```go
package validator

import (
	"fmt"

	"github.com/2389-research/dippin-lang/ir"
)

// lintLastResponseTruncate checks DIP148: last_response_truncate must not be
// negative, on an agent node or a parallel-branch override. 0 / unset means "no
// truncation" (the formatter emits only positive values), so only values < 0 are
// flagged. dippin carries + lints; a runtime enforces the truncation. This does
// NOT interact with DIP147 — a sink carrying last_response_truncate still emits
// the chain-attack Hint (truncation bounds size, not the laundered flow).
func lintLastResponseTruncate(w *ir.Workflow) []Diagnostic {
	var diags []Diagnostic
	for _, n := range w.Nodes {
		switch cfg := n.Config.(type) {
		case ir.AgentConfig:
			if cfg.LastResponseTruncate < 0 {
				diags = append(diags, lastResponseTruncateDiag(
					fmt.Sprintf("agent %q last_response_truncate is %d; cannot be negative", n.ID, cfg.LastResponseTruncate)))
			}
		case ir.ParallelConfig:
			for _, b := range cfg.Branches {
				if b.LastResponseTruncate < 0 {
					diags = append(diags, lastResponseTruncateDiag(
						fmt.Sprintf("parallel %q branch -> %q last_response_truncate is %d; cannot be negative", n.ID, b.Target, b.LastResponseTruncate)))
				}
			}
		}
	}
	return diags
}

func lastResponseTruncateDiag(msg string) Diagnostic {
	return Diagnostic{
		Code:     DIP148,
		Severity: SeverityWarning,
		Message:  msg,
		Help:     "use a non-negative character count (e.g. last_response_truncate: 4096), or omit it / set 0 for no truncation",
	}
}
```

(Verify `SeverityWarning`, `Diagnostic`, and the field names `Code`/`Severity`/`Message`/`Help` against `validator/lint_budget.go` — they match DIP145's construction.)

- [ ] **Step 6: Wire it into the dispatch**

In `validator/lint.go`, after `lintChainAttack(w)` (~line 79):

```go
	diags = append(diags, lintLastResponseTruncate(w)...)
```

- [ ] **Step 7: Run tests to verify they pass**

Run: `just test-pkg validator`
Expected: PASS — DIP148 tests green; `TestExplanationsCoverAllCodes` / `TestExplanationsNoExtra` still green.

- [ ] **Step 8: Commit**

```bash
git add validator/lint_last_response_truncate.go validator/lint_last_response_truncate_test.go validator/lint_codes.go validator/explanations.go validator/lint.go
git commit -m "feat(#56): DIP148 lints negative last_response_truncate

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

## Task 6: DIP147 non-interaction regression test

**Files:**
- Test: `validator/lint_chain_attack_test.go`

This locks the design decision that a mitigated sink **still** emits DIP147 (no suppression / fail-open).

- [ ] **Step 1: Write the regression test**

Add to `validator/lint_chain_attack_test.go` (read the file first to reuse its existing `lintSrc`/`hasCode` usage and a known DIP147-firing fixture):

```go
func TestLint_DIP147_StillFiresWhenSinkHasLastResponseTruncate(t *testing.T) {
	src := `workflow X
  start: S
  exit: W

  agent S
    prompt: "summarize untrusted input"
    tool_access: none
    writes: tainted

  agent W
    prompt: "write the report"
    last_response_truncate: 100
    reads: tainted

  edges
    S -> W
`
	if !hasCode(lintSrc(t, src), DIP147) {
		t.Errorf("DIP147 must still fire when sink sets last_response_truncate (truncation is not a full fix); got: %v", codes(lintSrc(t, src)))
	}
}
```

(Adjust the `edges` / `reads:` / `writes:` syntax to match the exact form the existing passing DIP147 tests in this file use — copy a known-firing fixture and add only `last_response_truncate: 100` to the sink.)

- [ ] **Step 2: Run test to verify it passes immediately**

Run: `just test-pkg validator`
Expected: PASS (no production change — `lintLastResponseTruncate` and `lintChainAttack` are independent). This test is a guard: if a future change couples them, it fails.

- [ ] **Step 3: Commit**

```bash
git add validator/lint_chain_attack_test.go
git commit -m "test(#56): DIP147 still fires when sink sets last_response_truncate

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

## Task 7: Docs + generated-spec refresh

**Files:**
- Modify: `docs/validation.md` (add a DIP148 section; bump range strings)
- Modify: `docs/llm-reference.md` (append DIP148 to the warning-list prose ~line 193; bump the range string)
- Modify: `docs/cli.md`, `docs/integration.md`, `docs/architecture.md`, `docs/editor-setup.md` (bump `DIP101–DIP147` → `DIP101–DIP148`; in `cli.md` also `All 56 diagnostic rules` → `57`)
- Regenerate: `cmd/dippin/generated-spec.md` (via the pre-commit hook or `bash scripts/gen-spec.sh`)
- **Do NOT touch** `site/content/validation.md` (hand-maintained; surfaced in the next release's docs(site) pass per the #100→#107 precedent).

- [ ] **Step 1: Find every range string**

Run:
```bash
grep -rnE "DIP101[–-]DIP147" docs/ validator/lint.go validator/lint_codes.go
grep -rn "56 diagnostic" docs/
```
Record each hit.

- [ ] **Step 2: Add the DIP148 description to `docs/validation.md`**

In the DIP-by-DIP section, after the DIP147 entry, add a DIP148 entry matching the surrounding format (Severity: Warning; what triggers it; the fix). Mirror the DIP145 entry's structure.

- [ ] **Step 3: Append DIP148 to the `docs/llm-reference.md` warning list**

At the end of the `**DIP101–DIP147** (warnings): ...` prose list (~line 193), append `, negative last_response_truncate` and change the range to `**DIP101–DIP148**`.

- [ ] **Step 4: Bump all range strings**

Replace each `DIP101–DIP147` (en-dash) and any `DIP101-DIP147` (hyphen) hit from Step 1 with `…DIP148`. In `docs/cli.md`, also change `All 56 diagnostic rules` → `All 57 diagnostic rules`. Use the precise count: 10 structural (DIP001–DIP010) + 47 semantic = 57.

- [ ] **Step 5: Regenerate the spec and verify freshness**

Run:
```bash
bash scripts/gen-spec.sh
just spec-check
```
Expected: `spec-check` passes (the tracked `cmd/dippin/generated-spec.md` is current). If `gen-spec.sh` reports it only writes `docs/generated-spec.md`, the pre-commit hook copies/refreshes the tracked `cmd/dippin/generated-spec.md` — commit whatever it regenerates.

- [ ] **Step 6: Commit**

```bash
git add docs/validation.md docs/llm-reference.md docs/cli.md docs/integration.md docs/architecture.md docs/editor-setup.md cmd/dippin/generated-spec.md
git commit -m "docs(#56): document DIP148 + last_response_truncate; bump code range to DIP148

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

## Task 8: Full verification, example, and PR

**Files:**
- Optional: `examples/` (a small demonstrator — only if the maintainer wants one; skip otherwise to stay minimal)

- [ ] **Step 1: Run the full gate locally**

Run, in order:
```bash
just fmt
just lint-go
just complexity
just test
just spec-check
```
Expected: all pass. (If `just check` is run, it may fail only at the final `tree-sitter-generate` step — that is the known environment gotcha, not a regression. The grammar was NOT changed in this work, so no tree-sitter regen is needed.)

- [ ] **Step 2: Run the pre-commit hook as the real CI mirror**

The previous task commits already trigger `scripts/pre-commit`. Confirm the last commit printed `=== pre-commit: all checks passed ===`. If any commit was made with `--no-verify`, re-run by amending or via `bash scripts/pre-commit`.

- [ ] **Step 3: Self-review the diff**

Run: `git diff origin/main --stat` and skim `git diff origin/main`.
Confirm: every changed line traces to last_response_truncate / DIP148; no stray edits; no non-ASCII introduced (`grep -rnP "[^\x00-\x7F]"` on any file where a literal `''` or quote was typed — none expected here).

- [ ] **Step 4: Run a code-review pass**

Invoke the `/code-review` skill (effort `high`) on the branch diff. Triage findings; fix real issues; re-run the affected `just test-pkg <pkg>`.

- [ ] **Step 5: Push and open the PR**

```bash
git push -u origin feat/56-last-response-truncate
```
Open the PR with a retry loop (GraphQL flakes here) — base `main`. Title: `last_response_truncate: carry-only chain-attack mitigation + DIP148 (#56)`. Body must state:
- This **partially addresses #56** (the mitigation attribute), does NOT `Closes` it (the `${ctx.last_response}` auto-injection topology + cross-file chains remain follow-ups).
- Attribute is **carried + linted by dippin, enforced by a runtime** — inert until a runtime reads it (the never-gate-on-runtime split).
- No DIP147 interaction (a mitigated sink still shows the Hint).
- `site/content/validation.md` deliberately untouched (next release's docs(site) pass).

- [ ] **Step 6: File the runtime-enforcement follow-up issue**

Open a tracker issue: a paired runtime must read `last_response_truncate` and cap the auto-injected previous response to N characters (per-agent; branch override resolves first, branch 0 = inherit). Link it from the PR body.

- [ ] **Step 7: Watch CI to green and address review bots**

Poll `gh pr checks`; address CodeRabbit/Copilot/Codex comments (a stale-count or range-string nit is the likely class). Merge only with maintainer go-ahead.

---

## Self-review (completed by plan author)

- **Spec coverage:** placement (Task 1 — sink agent + branch), char units (Task 1 parser + comments), carry path parse→IR→formatter→DOT→migrate (Tasks 1–4), DIP148 negative-only (Task 5), no-DIP147-interaction (Task 6, with regression guard), docs incl. `site/content/validation.md` deferral (Task 7), tracker follow-up (Task 8 Step 6). All spec sections map to a task.
- **Placeholder scan:** the three "read the sibling test harness and clone it" notes (migrate read-back test, migrate branch parity comparator, DIP147 fixture form) are deliberate — the exact harness/fixture differs by file and must be copied verbatim rather than guessed. The executor reads the named function and mirrors it; the asserted field and expected value are fully specified.
- **Type consistency:** `LastResponseTruncate int` used identically across `AgentConfig`, `BranchConfig`, parser, formatter, DOT, migrate, validator. Lint code `DIP148`, function `lintLastResponseTruncate`, helper `lastResponseTruncateDiag`, branch DOT helper `appendBranchIntField` — each defined once and referenced consistently.
