# `writable_paths` Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a `writable_paths` glob-list field to dippin agent nodes and parallel branches that carries + lints path-bounded write scope (the runtime enforces the fs-jail at runtime).

**Architecture:** `writable_paths` is a `[]string` on `AgentConfig` and `BranchConfig`, orthogonal to the `tool_access` scalar. It rides the established list-field rails (`reads`/`writes`/`outputs`) through parse → IR → format → DOT export → migrate, plus two new lints (DIP141 dead-config, DIP142 unsafe entry). dippin does no resolution/coercion; the parser uses `splitCommaNoEmpty` so empty fails closed at the runtime. This plan is **dippin-side only** — the runtime fs-jail enforcement is a separate, joint-released PR (see spec § Release coordination).

**Tech Stack:** Go; `just` for all build/test/lint; pre-commit enforces cyclo ≤ 5 / cognit ≤ 7, gofmt, golangci-lint, race tests, example validation.

**Spec:** `docs/superpowers/specs/2026-05-29-issue-75-writable-paths-design.md`

**Conventions (CLAUDE.md):**
- All ops via `just`, never raw `go`. Gate: `just check`. Single pkg: `just test-pkg <pkg>`. Complexity: `just complexity`.
- TDD: failing test first, watch it fail, then implement. Never `//nolint` — extract helpers.
- Commit trailer: `Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>`
- Work is on branch `feat/75-write-paths` in worktree `/home/clint/code/2389/dippin-lang-75-write-paths`. Never commit to `main`.
- Do not hand-edit `docs/generated-spec.md` (pre-commit regenerates it). Do not touch `CHANGELOG.md`.

**Cross-cutting facts (verified against current source — do not re-discover):**
- `parser.splitComma("")` returns `[""]` (len 1); `parser.splitCommaNoEmpty("")` returns `nil`. We use **`splitCommaNoEmpty`** for `writable_paths` (fail-closed). `migrate.splitComma` already drops empties, so both round-trip paths agree.
- Adding a slice field to `BranchConfig` makes it **non-comparable** → breaks `migrate/parity.go:compareParallelBranches` (`!=`) AND `migrate/roundtrip_test.go:assertBranchesEqual` (`!=`). Both must be fixed in Task 1 or nothing compiles.
- These functions are already at cyclo **5** (the cap) and need helper extraction before a new branch can be added: `parser.applyBranchField`, `formatter.writeBranch`, `formatter.writeBranchFields`, `migrate.applyRuntimeAttrs`. `parser.applyAgentRuntimeField` is cyclo 4 (one case of headroom — no refactor needed).
- `parser.hasParentRef` is unexported; the `validator` package must keep its own copy (codebase rule: packages import `ir` only).
- Recognized backends: `native`, `claude-code`, `acp` (no validation exists; do not add one here).

---

## Task 0: Confirm follow-up issues + sequence (admin, no code)

Restores the #41 "task #0 files the follow-ups" discipline. #55/#53/#56 already exist as issues; this task only verifies and records.

- [ ] **Step 1: Verify the sequenced follow-up issues exist**

Run:
```bash
gh issue view 55 --json number,title,state -q '.number,.title,.state'
gh issue view 53 --json number,title,state -q '.number,.title,.state'
gh issue view 56 --json number,title,state -q '.number,.title,.state'
```
Expected: all three resolve (titles for tool-name allowlists / defaults cascade / chain-attack). If any is missing, file it before proceeding (mirror the existing issue bodies). No commit for this step.

---

## Task 1: IR fields + parity/test-helper fixes (compile-green foundation)

**Files:**
- Modify: `ir/ir.go` (AgentConfig ~line 91-113, BranchConfig ~line 151-165)
- Modify: `migrate/parity.go` (`compareParallelBranches` 365-379, `compareAgentBehavior` 261-274)
- Modify: `migrate/roundtrip_test.go` (`assertBranchesEqual`)
- Test: `migrate/parity_test.go` (new test for the slice compare)

- [ ] **Step 1: Write the failing parity test**

Add to `migrate/parity_test.go` (create the file if it does not exist; package `migrate`):

```go
package migrate

import (
	"testing"

	"github.com/2389-research/dippin-lang/ir"
)

func TestParity_AgentWritablePathsDiff(t *testing.T) {
	a := ir.AgentConfig{WritablePaths: []string{"workspace/**"}}
	b := ir.AgentConfig{WritablePaths: []string{"workspace/**", ".ai/**"}}
	diffs := compareAgentConfigs("N", "", a, b)
	if len(diffs) == 0 {
		t.Fatalf("expected a writable_paths difference, got none")
	}
}

func TestParity_BranchWritablePathsDiff(t *testing.T) {
	a := ir.ParallelConfig{Branches: []ir.BranchConfig{{Target: "x", WritablePaths: []string{"workspace/**"}}}}
	b := ir.ParallelConfig{Branches: []ir.BranchConfig{{Target: "x", WritablePaths: []string{".ai/**"}}}}
	diffs := compareParallelBranches("N", a, b)
	if len(diffs) == 0 {
		t.Fatalf("expected a branch writable_paths difference, got none")
	}
}
```

- [ ] **Step 2: Run the test — expect a COMPILE failure**

Run: `just test-pkg migrate`
Expected: build error — `ir.AgentConfig` / `ir.BranchConfig` have no field `WritablePaths`. (This is the red state.)

- [ ] **Step 3: Add the IR fields**

In `ir/ir.go`, in `AgentConfig`, immediately after the `ToolAccess` field (line 111):

```go
	// WritablePaths bounds the file paths this agent's tools may write, as
	// author-chosen globs (e.g. "workspace/**", ".ai/sprints/**") resolved against
	// the session root. Empty/absent = unbounded. A present-but-empty or malformed
	// value fails CLOSED at the runtime (deny-all / refuse-to-start), never
	// unbounded. dippin carries + lints; the runtime enforces an fs-level write jail on
	// the native backend (Bash + its children included); claude-code/acp refuse to
	// start. See issue #75.
	WritablePaths []string
```

In `BranchConfig`, after the `ToolAccess` field (line 164):

```go
	// WritablePaths is a per-branch override of the target agent's writable_paths.
	// Empty INHERITS the target agent's writable_paths (never resets to unbounded) —
	// the runtime resolves effective = branch if non-empty else agent. dippin carries +
	// lints; the runtime enforces. See issue #75.
	WritablePaths []string
```

- [ ] **Step 4: Fix `compareParallelBranches` (the `!=` no longer compiles)**

In `migrate/parity.go`, replace `compareParallelBranches` (365-379) with a field-by-field comparison and update the doc comment:

```go
// compareParallelBranches compares branch slices position-by-position. Order is
// significant (it maps to targets). BranchConfig now carries a slice field
// (WritablePaths), so it is no longer comparable with ==; compare field-by-field.
func compareParallelBranches(id string, ac, bc ir.ParallelConfig) []Difference {
	if len(ac.Branches) != len(bc.Branches) {
		return []Difference{fieldDiff(id, "branches", fmt.Sprintf("node %q branches differ", id))}
	}
	for i := range ac.Branches {
		if !branchesEqual(ac.Branches[i], bc.Branches[i]) {
			return []Difference{fieldDiff(id, "branches", fmt.Sprintf("node %q branches differ", id))}
		}
	}
	return nil
}

// branchesEqual reports whether two branch configs are field-equal (scalar
// fields plus the comma-joined WritablePaths slice).
func branchesEqual(a, b ir.BranchConfig) bool {
	return a.Target == b.Target &&
		a.Model == b.Model &&
		a.Provider == b.Provider &&
		a.Fidelity == b.Fidelity &&
		a.ToolAccess == b.ToolAccess &&
		strings.Join(a.WritablePaths, ",") == strings.Join(b.WritablePaths, ",")
}
```

(`strings` is already imported in `parity.go`.) `compareParallelBranches` cyclo: 2 ifs + 1 for = 4. `branchesEqual` cyclo: 5 `&&` → 6 — **over cap**. Reduce by grouping the scalar compare:

```go
func branchesEqual(a, b ir.BranchConfig) bool {
	scalarsEqual := a.Target == b.Target && a.Model == b.Model &&
		a.Provider == b.Provider && a.Fidelity == b.Fidelity && a.ToolAccess == b.ToolAccess
	return scalarsEqual && strings.Join(a.WritablePaths, ",") == strings.Join(b.WritablePaths, ",")
}
```

That is still 5 `&&` total in one function (cyclo 6). Split the scalar compare into its own helper:

```go
func branchesEqual(a, b ir.BranchConfig) bool {
	return branchScalarsEqual(a, b) &&
		strings.Join(a.WritablePaths, ",") == strings.Join(b.WritablePaths, ",")
}

func branchScalarsEqual(a, b ir.BranchConfig) bool {
	return a.Target == b.Target && a.Model == b.Model &&
		a.Provider == b.Provider && a.Fidelity == b.Fidelity && a.ToolAccess == b.ToolAccess
}
```

`branchesEqual` cyclo 2; `branchScalarsEqual` cyclo 5 (4 `&&`). Both ≤ 5.

- [ ] **Step 5: Add the agent `WritablePaths` compare**

In `migrate/parity.go`, in `compareAgentBehavior` (261-274), after the `ToolAccess` block and before `return diffs`:

```go
	if strings.Join(ac.WritablePaths, ",") != strings.Join(bc.WritablePaths, ",") {
		diffs = append(diffs, fieldDiff(id, "writable_paths", fmt.Sprintf("node %q writable_paths: %v vs %v", id, ac.WritablePaths, bc.WritablePaths)))
	}
```

`compareAgentBehavior` goes from cyclo 4 → 5 (at cap, OK).

- [ ] **Step 6: Fix the `assertBranchesEqual` test helper (also uses `!=`)**

In `migrate/roundtrip_test.go`, replace the comparison line inside `assertBranchesEqual`:

```go
	for i := range want {
		if !branchesEqual(got[i], want[i]) {
			t.Errorf("branch[%d]: got %+v want %+v", i, got[i], want[i])
		}
	}
```

(Reuses the production `branchesEqual` helper from Step 4 — same package `migrate`.)

- [ ] **Step 7: Run the tests — expect PASS**

Run: `just test-pkg migrate`
Expected: PASS (both new parity tests green, existing round-trip tests still green).

- [ ] **Step 8: Verify complexity, then commit**

Run: `just complexity`
Expected: `Complexity OK.`

```bash
git add ir/ir.go migrate/parity.go migrate/parity_test.go migrate/roundtrip_test.go
git commit -m "feat(ir): add writable_paths field to AgentConfig/BranchConfig + parity

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

## Task 2: Parser — agent `writable_paths`

**Files:**
- Modify: `parser/parse_nodes.go` (`applyAgentRuntimeField` 332-345)
- Test: `parser/parse_writable_paths_test.go` (new)

- [ ] **Step 1: Write the failing test**

Create `parser/parse_writable_paths_test.go` (package `parser`):

```go
package parser

import (
	"testing"

	"github.com/2389-research/dippin-lang/ir"
)

func TestParseAgentWritablePaths(t *testing.T) {
	src := `workflow X
  start: A
  exit: A

  agent A
    prompt: "x"
    writable_paths: workspace/**, .ai/sprints/**
`
	w, err := NewParser(src, "test.dip").Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	cfg := w.Node("A").Config.(ir.AgentConfig)
	want := []string{"workspace/**", ".ai/sprints/**"}
	if len(cfg.WritablePaths) != 2 || cfg.WritablePaths[0] != want[0] || cfg.WritablePaths[1] != want[1] {
		t.Errorf("WritablePaths = %v, want %v", cfg.WritablePaths, want)
	}
}

func TestParseAgentWritablePathsEmptyIsNil(t *testing.T) {
	src := `workflow X
  start: A
  exit: A

  agent A
    prompt: "x"
    writable_paths:
`
	w, err := NewParser(src, "test.dip").Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	cfg := w.Node("A").Config.(ir.AgentConfig)
	if cfg.WritablePaths != nil {
		t.Errorf("bare writable_paths should be nil (fail-closed at the runtime), got %v", cfg.WritablePaths)
	}
}
```

- [ ] **Step 2: Run the test — expect FAIL**

Run: `just test-pkg parser`
Expected: FAIL — `WritablePaths = []` want `[workspace/** .ai/sprints/**]` (the field is never set; parser drops the unknown key).

- [ ] **Step 3: Implement the parser case**

In `parser/parse_nodes.go`, in `applyAgentRuntimeField` (332-345), add a case after `tool_access`:

```go
	case "writable_paths":
		cfg.WritablePaths = splitCommaNoEmpty(val)
```

(`applyAgentRuntimeField` goes cyclo 4 → 5, at cap, OK.)

- [ ] **Step 4: Run the test — expect PASS**

Run: `just test-pkg parser`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add parser/parse_nodes.go parser/parse_writable_paths_test.go
git commit -m "feat(parser): parse agent writable_paths (splitCommaNoEmpty)

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

## Task 3: Parser — branch `writable_paths` (refactor `applyBranchField` to table-driven)

**Files:**
- Modify: `parser/parse_nodes.go` (`applyBranchField` 801-813)
- Test: `parser/parse_writable_paths_test.go`

- [ ] **Step 1: Write the failing test**

Append to `parser/parse_writable_paths_test.go`:

```go
func TestParseBranchWritablePaths(t *testing.T) {
	src := `workflow X
  start: split
  exit: join

  agent a
    prompt: "a"

  parallel split
    branch: a
      writable_paths: workspace/**, .ai/**

  fan_in join <- a

  edges
    split -> a
    a -> join
`
	w, err := NewParser(src, "test.dip").Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	cfg := w.Node("split").Config.(ir.ParallelConfig)
	if len(cfg.Branches) != 1 {
		t.Fatalf("branches = %d, want 1", len(cfg.Branches))
	}
	got := cfg.Branches[0].WritablePaths
	if len(got) != 2 || got[0] != "workspace/**" || got[1] != ".ai/**" {
		t.Errorf("branch WritablePaths = %v, want [workspace/** .ai/**]", got)
	}
}
```

- [ ] **Step 2: Run the test — expect FAIL**

Run: `just test-pkg parser`
Expected: FAIL — branch `WritablePaths` is empty (the `applyBranchField` switch drops `writable_paths`).

- [ ] **Step 3: Refactor `applyBranchField` to table-driven and add the case**

`applyBranchField` is at cyclo 5; a 5th switch case would breach the cap. Replace the whole function (801-813) with a table-driven dispatch (mirrors migrate's `branchFieldSetters`):

```go
// branchFieldSetters maps a branch field key to the BranchConfig field it sets.
// Table-driven keeps applyBranchField under the cyclo≤5 cap as fields are added.
var branchFieldSetters = map[string]func(*ir.BranchConfig, string){
	"model":          func(b *ir.BranchConfig, v string) { b.Model = v },
	"provider":       func(b *ir.BranchConfig, v string) { b.Provider = v },
	"fidelity":       func(b *ir.BranchConfig, v string) { b.Fidelity = v },
	"tool_access":    func(b *ir.BranchConfig, v string) { b.ToolAccess = v },
	"writable_paths": func(b *ir.BranchConfig, v string) { b.WritablePaths = splitCommaNoEmpty(v) },
}

// applyBranchField sets a field on a BranchConfig.
func applyBranchField(bc *ir.BranchConfig, key, val string) {
	if set, ok := branchFieldSetters[key]; ok {
		set(bc, val)
	}
}
```

`applyBranchField` is now cyclo 2. (The map literal is package-level, not a function body, so it does not count against complexity.)

- [ ] **Step 4: Run the test — expect PASS**

Run: `just test-pkg parser`
Expected: PASS (new branch test + existing `TestParseBranchToolAccess` at line 2426 still green).

- [ ] **Step 5: Verify complexity, then commit**

Run: `just complexity`
Expected: `Complexity OK.`

```bash
git add parser/parse_nodes.go parser/parse_writable_paths_test.go
git commit -m "feat(parser): parse per-branch writable_paths (table-driven applyBranchField)

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

## Task 4: Formatter — agent emit

**Files:**
- Modify: `formatter/format.go` (`writeAgentRuntimeFields` 382-393)
- Test: `formatter/format_test.go`

- [ ] **Step 1: Write the failing test**

Append to `formatter/format_test.go` (package `formatter`):

```go
func TestFormatAgentWritablePaths(t *testing.T) {
	w := &ir.Workflow{
		Name: "T", Start: "A", Exit: "A",
		Nodes: []*ir.Node{
			{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{
				Prompt:        "x",
				WritablePaths: []string{"workspace/**", ".ai/sprints/**"},
			}},
		},
	}
	out := Format(w)
	if !strings.Contains(out, "writable_paths: workspace/**, .ai/sprints/**") {
		t.Errorf("formatted output missing writable_paths; got:\n%s", out)
	}
}
```

- [ ] **Step 2: Run the test — expect FAIL**

Run: `just test-pkg formatter`
Expected: FAIL — output does not contain the `writable_paths:` line.

- [ ] **Step 3: Implement the emit**

In `formatter/format.go`, in `writeAgentRuntimeFields` (382-393), after the `ToolAccess` block:

```go
	if len(cfg.WritablePaths) > 0 {
		wr.line("writable_paths: %s", strings.Join(cfg.WritablePaths, ", "))
	}
```

(`writeAgentRuntimeFields` goes cyclo 4 → 5, at cap, OK. List fields are emitted unquoted via `strings.Join`, matching `writeIOFields`.)

- [ ] **Step 4: Run the test — expect PASS**

Run: `just test-pkg formatter`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add formatter/format.go formatter/format_test.go
git commit -m "feat(formatter): emit agent writable_paths

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

## Task 5: Formatter — branch emit (refactor `writeBranch` + `writeBranchFields`)

**Files:**
- Modify: `formatter/format.go` (`writeBranch` 264-273, `writeBranchFields` 275-289)
- Test: `formatter/format_test.go`

- [ ] **Step 1: Write the failing test**

Append to `formatter/format_test.go` (mirrors `TestFormatBranchToolAccessOnly` at line 2138):

```go
func TestFormatBranchWritablePathsOnly(t *testing.T) {
	w := &ir.Workflow{
		Name: "T", Start: "split", Exit: "join",
		Nodes: []*ir.Node{
			{ID: "a", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "a"}},
			{ID: "split", Kind: ir.NodeParallel, Config: ir.ParallelConfig{
				Targets:  []string{"a"},
				Branches: []ir.BranchConfig{{Target: "a", WritablePaths: []string{"workspace/**"}}},
			}},
			{ID: "join", Kind: ir.NodeFanIn, Config: ir.FanInConfig{Sources: []string{"a"}}},
		},
	}
	out := Format(w)
	if !strings.Contains(out, "writable_paths: workspace/**") {
		t.Errorf("formatted output missing per-branch writable_paths; got:\n%s", out)
	}
}
```

- [ ] **Step 2: Run the test — expect FAIL**

Run: `just test-pkg formatter`
Expected: FAIL — branch with only `writable_paths` set is dropped by the `writeBranch` early-return guard (which doesn't know about the new field) and never emitted.

- [ ] **Step 3: Refactor the guard and the field emit**

In `formatter/format.go`, replace `writeBranch` (264-273) so the early-return uses a helper that includes the new field:

```go
// writeBranch writes a single branch entry in block-form parallel.
func writeBranch(wr *writer, b ir.BranchConfig) {
	wr.line("branch: %s", b.Target)
	if !branchHasFields(b) {
		return
	}
	wr.push()
	writeBranchFields(wr, b)
	wr.pop()
}

// branchHasFields reports whether a branch carries any optional field beyond its target.
func branchHasFields(b ir.BranchConfig) bool {
	return b.Model != "" || b.Provider != "" || b.Fidelity != "" ||
		b.ToolAccess != "" || len(b.WritablePaths) > 0
}
```

`writeBranch` cyclo 2; `branchHasFields` cyclo 5 (4 `||`). Both ≤ 5.

Replace `writeBranchFields` (275-289) by extracting the scalar emits, then adding the list emit:

```go
// writeBranchFields writes the optional fields within a branch.
func writeBranchFields(wr *writer, b ir.BranchConfig) {
	writeBranchScalarFields(wr, b)
	if len(b.WritablePaths) > 0 {
		wr.line("writable_paths: %s", strings.Join(b.WritablePaths, ", "))
	}
}

// writeBranchScalarFields writes the scalar per-branch overrides.
func writeBranchScalarFields(wr *writer, b ir.BranchConfig) {
	if b.Model != "" {
		wr.line("model: %s", quoteValue(b.Model))
	}
	if b.Provider != "" {
		wr.line("provider: %s", quoteValue(b.Provider))
	}
	if b.Fidelity != "" {
		wr.line("fidelity: %s", quoteValue(b.Fidelity))
	}
	if b.ToolAccess != "" {
		wr.line("tool_access: %s", quoteValue(b.ToolAccess))
	}
}
```

`writeBranchFields` cyclo 2; `writeBranchScalarFields` cyclo 5 (4 ifs). Both ≤ 5.

- [ ] **Step 4: Run the test — expect PASS**

Run: `just test-pkg formatter`
Expected: PASS (new test + `TestFormatBranchToolAccessOnly` + `TestFormatParallelBlock` still green).

- [ ] **Step 5: Verify complexity, then commit**

Run: `just complexity`
Expected: `Complexity OK.`

```bash
git add formatter/format.go formatter/format_test.go
git commit -m "feat(formatter): emit per-branch writable_paths (extract branchHasFields/scalar helpers)

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

## Task 6: DOT export — agent attr

**Files:**
- Modify: `export/dot.go` (`applyAgentRuntimeAttrs` 300-311)
- Test: `export/dot_test.go`

- [ ] **Step 1: Write the failing test**

Append to `export/dot_test.go` (package `export`; confirm the package name at the top of that file and match it):

```go
func TestExportAgentWritablePaths(t *testing.T) {
	w := &ir.Workflow{
		Name: "T", Start: "A", Exit: "A",
		Nodes: []*ir.Node{
			{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{
				Prompt:        "x",
				WritablePaths: []string{"workspace/**", ".ai/**"},
			}},
		},
	}
	dot := ExportDOT(w, ExportOptions{})
	if !strings.Contains(dot, `writable_paths="workspace/**,.ai/**"`) {
		t.Errorf("DOT missing writable_paths attr; got:\n%s", dot)
	}
}
```

- [ ] **Step 2: Run the test — expect FAIL**

Run: `just test-pkg export`
Expected: FAIL — DOT output has no `writable_paths` attr.

- [ ] **Step 3: Implement the attr**

In `export/dot.go`, in `applyAgentRuntimeAttrs` (300-311), after the `ToolAccess` block:

```go
	if len(cfg.WritablePaths) > 0 {
		attrs["writable_paths"] = strings.Join(cfg.WritablePaths, ",")
	}
```

(`applyAgentRuntimeAttrs` goes cyclo 4 → 5, at cap, OK. Bare-comma join, matching `applyToolOutputsAttrs`. Do NOT touch `reservedGraphAttrs` — that map is graph-level only; `tool_access` isn't in it either.)

- [ ] **Step 4: Run the test — expect PASS**

Run: `just test-pkg export`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add export/dot.go export/dot_test.go
git commit -m "feat(export): emit agent writable_paths DOT attr

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

## Task 7: DOT export — branch encode

**Files:**
- Modify: `export/dot.go` (`encodeBranch` 425-434)
- Test: covered by the round-trip test in Task 10 (the branch encode is verified end-to-end there; no standalone test needed — adding one would duplicate Task 10).

- [ ] **Step 1: Implement the branch field encode**

In `export/dot.go`, in `encodeBranch` (425-434), after the `tool_access` line:

```go
	parts = appendBranchField(parts, "writable_paths", strings.Join(b.WritablePaths, ","))
```

`appendBranchField` omits the field when the joined value is `""` (empty/nil slice), and `encodeBranchToken` percent-encodes the internal commas (`,`→`%2C`) so the list survives the branch (`,`) and field (`;`) separators losslessly. `encodeBranch` is straight-line (cyclo unchanged).

- [ ] **Step 2: Build to confirm it compiles**

Run: `just build`
Expected: success.

- [ ] **Step 3: Commit**

```bash
git add export/dot.go
git commit -m "feat(export): encode per-branch writable_paths in DOT branch token

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

## Task 8: Migrate — agent attr read (refactor `applyRuntimeAttrs`)

**Files:**
- Modify: `migrate/migrate.go` (`applyRuntimeAttrs` 419-430)
- Test: covered by Task 10 round-trip; plus a direct unit test below.

- [ ] **Step 1: Write the failing test**

Append to `migrate/roundtrip_test.go` (package `migrate`):

```go
func TestMigrateAgentWritablePaths(t *testing.T) {
	w := &ir.Workflow{
		Name: "T", Start: "A", Exit: "A",
		Nodes: []*ir.Node{
			{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{
				Prompt:        "x",
				WritablePaths: []string{"workspace/**", ".ai/**"},
			}},
		},
	}
	dot := export.ExportDOT(w, export.ExportOptions{})
	got, err := Migrate(dot)
	if err != nil {
		t.Fatalf("migrate: %v", err)
	}
	cfg := got.Node("A").Config.(ir.AgentConfig)
	if len(cfg.WritablePaths) != 2 || cfg.WritablePaths[0] != "workspace/**" || cfg.WritablePaths[1] != ".ai/**" {
		t.Errorf("WritablePaths after migrate = %v, want [workspace/** .ai/**]", cfg.WritablePaths)
	}
}
```

- [ ] **Step 2: Run the test — expect FAIL**

Run: `just test-pkg migrate`
Expected: FAIL — `WritablePaths` is empty after migrate (the attr is never read back).

- [ ] **Step 3: Refactor `applyRuntimeAttrs` and read the attr**

`applyRuntimeAttrs` is at cyclo 5; adding a block breaches the cap. Replace it (419-430) by extracting the conditional branches into a helper:

```go
// applyRuntimeAttrs applies runtime-cluster attributes (backend, working_dir,
// tool_access, writable_paths).
func applyRuntimeAttrs(cfg *ir.AgentConfig, attrs map[string]string) {
	if v, ok := attrs["backend"]; ok {
		cfg.Backend = v
	}
	if v, ok := attrs["working_dir"]; ok {
		cfg.WorkingDir = v
	}
	applyRuntimeSafetyAttrs(cfg, attrs)
}

// applyRuntimeSafetyAttrs applies the tool_access + writable_paths safety attrs.
func applyRuntimeSafetyAttrs(cfg *ir.AgentConfig, attrs map[string]string) {
	if v, ok := attrs["tool_access"]; ok && strings.TrimSpace(v) != "" {
		cfg.ToolAccess = v
	}
	if v, ok := attrs["writable_paths"]; ok {
		cfg.WritablePaths = splitComma(v)
	}
}
```

`applyRuntimeAttrs` cyclo 3; `applyRuntimeSafetyAttrs` cyclo 4 (the `ok && trim` is +2, the second `if ok` is +1, base 1 = 4). Both ≤ 5. (migrate's `splitComma` drops empties, matching the parser's `splitCommaNoEmpty`.)

- [ ] **Step 4: Run the test — expect PASS**

Run: `just test-pkg migrate`
Expected: PASS.

- [ ] **Step 5: Verify complexity, then commit**

Run: `just complexity`
Expected: `Complexity OK.`

```bash
git add migrate/migrate.go migrate/roundtrip_test.go
git commit -m "feat(migrate): read agent writable_paths DOT attr (extract safety-attr helper)

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

## Task 9: Migrate — branch setter

**Files:**
- Modify: `migrate/migrate.go` (`branchFieldSetters` 657-667)
- Test: covered by Task 10.

- [ ] **Step 1: Add the branch setter**

In `migrate/migrate.go`, in the `branchFieldSetters` map (657-667), add after the `tool_access` entry:

```go
	"writable_paths": func(b *ir.BranchConfig, v string) { b.WritablePaths = splitComma(v) },
```

The setter receives the already-`decodeBranchToken`'d value (commas restored from `%2C`), so `splitComma` recovers the list. Map literal — no complexity impact.

- [ ] **Step 2: Build to confirm it compiles**

Run: `just build`
Expected: success.

- [ ] **Step 3: Commit**

```bash
git add migrate/migrate.go
git commit -m "feat(migrate): decode per-branch writable_paths from DOT branch token

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

## Task 10: Round-trip tests (agent format + DOT; branch DOT)

**Files:**
- Modify: `migrate/roundtrip_test.go`
- Test: `formatter/format_test.go` (format-idempotence for agent)

- [ ] **Step 1: Write the failing branch round-trip test**

Extend `TestRoundtripBlockFormParallel` in `migrate/roundtrip_test.go`: add a `writable_paths` line to the `fast` branch in the `src` string, and add `WritablePaths` to the expected branch. Change the `branch: fast` block to:

```dippin
    branch: fast
      model: claude-haiku-4-5
      provider: anthropic
      fidelity: summary
      tool_access: none
      writable_paths: workspace/**, .ai/sprints/**
```

and update the `assertBranchesEqual` expectation's first branch:

```go
		{Target: "fast", Model: "claude-haiku-4-5", Provider: "anthropic", Fidelity: "summary", ToolAccess: "none", WritablePaths: []string{"workspace/**", ".ai/sprints/**"}},
```

- [ ] **Step 2: Write the failing agent format-idempotence test**

Append to `formatter/format_test.go`:

```go
func TestFormatAgentWritablePathsRoundTrips(t *testing.T) {
	src := `workflow X
  start: A
  exit: A

  agent A
    prompt: "x"
    writable_paths: workspace/**, .ai/sprints/**
`
	w1, err := parser.NewParser(src, "rt.dip").Parse()
	if err != nil {
		t.Fatalf("parse1: %v", err)
	}
	w2, err := parser.NewParser(Format(w1), "rt.dip").Parse()
	if err != nil {
		t.Fatalf("parse2: %v", err)
	}
	got := w2.Node("A").Config.(ir.AgentConfig).WritablePaths
	if len(got) != 2 || got[0] != "workspace/**" || got[1] != ".ai/sprints/**" {
		t.Errorf("WritablePaths after format round-trip = %v, want [workspace/** .ai/sprints/**]", got)
	}
}
```

(Confirm `parser` is imported in `format_test.go`; if not, add `"github.com/2389-research/dippin-lang/parser"`.)

- [ ] **Step 3: Run the tests — expect PASS (Tasks 2-9 already implement the field)**

Run: `just test-pkg migrate && just test-pkg formatter`
Expected: PASS. If the branch round-trip fails, the encode/decode (Tasks 7/9) or parity (Task 1) is wrong — debug there.

- [ ] **Step 4: Commit**

```bash
git add migrate/roundtrip_test.go formatter/format_test.go
git commit -m "test: round-trip writable_paths through DOT (branch) and format (agent)

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

## Task 11: Validator — DIP141 (dead-config) + registration + new lint file

**Files:**
- Create: `validator/lint_writable_paths.go`
- Modify: `validator/lint_codes.go` (const block ~42-45, CodeDescription ~87-88)
- Modify: `validator/explanations.go` (before line 414)
- Modify: `validator/lint.go` (header comment 8-11, dispatch ~60)
- Test: `validator/lint_writable_paths_test.go` (new)

- [ ] **Step 1: Write the failing DIP141 test**

Create `validator/lint_writable_paths_test.go` (package `validator`; reuses `hasCode`/`codes` from `lint_tool_access_test.go`):

```go
package validator

import (
	"testing"

	"github.com/2389-research/dippin-lang/parser"
)

func lintSrc(t *testing.T, src string) []Diagnostic {
	t.Helper()
	w, err := parser.NewParser(src, "test.dip").Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	return Lint(w).Diagnostics
}

func TestLint_DIP141_WritablePathsWithToolAccessNone(t *testing.T) {
	src := `workflow X
  start: A
  exit: A

  agent A
    prompt: "x"
    tool_access: none
    writable_paths: workspace/**
`
	if !hasCode(lintSrc(t, src), DIP141) {
		t.Errorf("expected DIP141, got: %v", codes(lintSrc(t, src)))
	}
}

func TestLint_DIP141_NotFiredWhenAlone(t *testing.T) {
	src := `workflow X
  start: A
  exit: A

  agent A
    prompt: "x"
    writable_paths: workspace/**
`
	if hasCode(lintSrc(t, src), DIP141) {
		t.Errorf("DIP141 should not fire without tool_access: none; got: %v", codes(lintSrc(t, src)))
	}
}
```

- [ ] **Step 2: Run the test — expect FAIL (undefined: DIP141)**

Run: `just test-pkg validator`
Expected: build error — `undefined: DIP141`.

- [ ] **Step 3: Register the new codes**

In `validator/lint_codes.go`, inside the `const` block after `DIP140` (line 44):

```go
	DIP141 = "DIP141" // writable_paths nullified by tool_access: none (dead config)
	DIP142 = "DIP142" // unsafe writable_paths entry (absolute / ~ / parent-escape / brace)
```

In the `init()` body after the `DIP140` description (line 88):

```go
	CodeDescription[DIP141] = "writable_paths nullified by tool_access: none (dead config)"
	CodeDescription[DIP142] = "unsafe writable_paths entry (absolute / ~ / parent-escape / brace)"
```

- [ ] **Step 4: Bump the `lint.go` header range and wire the dispatch**

In `validator/lint.go`, change the header comment (line 8) `DIP101–DIP140` → `DIP101–DIP142`. In the dispatch (after line 60, before `return Result{...}`):

```go
	diags = append(diags, lintWritablePaths(w)...)
```

- [ ] **Step 5: Create `validator/lint_writable_paths.go` with the DIP141 ladder**

```go
package validator

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/2389-research/dippin-lang/ir"
)

// lintWritablePaths fires DIP141 (writable_paths nullified by tool_access: none)
// and DIP142 (unsafe writable_paths entry) on agent nodes and per-branch overrides.
// dippin carries + lints the field; the runtime enforces the fs-level write jail.
func lintWritablePaths(w *ir.Workflow) []Diagnostic {
	var diags []Diagnostic
	for _, n := range w.Nodes {
		diags = append(diags, checkNodeWritablePathsByKind(n)...)
	}
	return diags
}

func checkNodeWritablePathsByKind(n *ir.Node) []Diagnostic {
	switch cfg := n.Config.(type) {
	case ir.AgentConfig:
		return checkWritablePathsObject(n, cfg.WritablePaths, cfg.ToolAccess, "")
	case ir.ParallelConfig:
		return checkBranchWritablePaths(n, cfg.Branches)
	default:
		return nil
	}
}

func checkBranchWritablePaths(n *ir.Node, branches []ir.BranchConfig) []Diagnostic {
	var diags []Diagnostic
	for _, b := range branches {
		diags = append(diags, checkWritablePathsObject(n, b.WritablePaths, b.ToolAccess, b.Target)...)
	}
	return diags
}

func checkWritablePathsObject(n *ir.Node, paths []string, toolAccess, branch string) []Diagnostic {
	if len(paths) == 0 {
		return nil
	}
	var diags []Diagnostic
	if strings.ToLower(strings.TrimSpace(toolAccess)) == "none" {
		diags = append(diags, dip141Diagnostic(n, branch))
	}
	return diags
}

func dip141Diagnostic(n *ir.Node, branch string) Diagnostic {
	msg := fmt.Sprintf("node %q has writable_paths but tool_access \"none\" — none strips all tools, so there is nothing to bound (dead config)", n.ID)
	if branch != "" {
		msg = fmt.Sprintf("node %q branch %q has writable_paths but tool_access \"none\" — none strips all tools, so there is nothing to bound (dead config)", n.ID, branch)
	}
	return Diagnostic{
		Code:     DIP141,
		Severity: SeverityWarning,
		Message:  msg,
		Location: n.Source,
		Help:     "remove writable_paths (no tools to bound) or drop tool_access: none to grant a bounded tool catalog.",
	}
}
```

(The `path/filepath` import is unused until Task 12 — add it now and Task 12 uses it, OR omit it here and add in Task 12. To keep this step compiling, **omit `"path/filepath"` for now** and add it in Task 12 Step 3.)

`checkWritablePathsObject` cyclo 3; all others ≤ 3.

- [ ] **Step 6: Add the DIP141 explanation**

In `validator/explanations.go`, before the closing `}` of `nodeValidationExplanations()` (line 414), after the `DIP140` entry:

```go
		DIP141: {
			Code:    DIP141,
			Summary: "writable_paths nullified by tool_access: none (dead config)",
			Trigger: "An agent node or per-branch override sets writable_paths together with tool_access: none on the same object. tool_access: none strips the entire tool catalog, so there is no Write/Edit/Bash left for writable_paths to bound — the field is dead config.",
			Fix:     "Remove writable_paths (there are no tools to bound) or drop tool_access: none to grant a bounded tool catalog. A branch that inherits writable_paths while setting tool_access: none is legitimate narrowing and is not flagged.",
			Example: "agent Summarize\n  prompt: \"Summarize\"\n  tool_access: none\n  writable_paths: workspace/**   // DIP141: none strips all tools — nothing to bound",
		},
```

- [ ] **Step 7: Run the test — expect PASS**

Run: `just test-pkg validator`
Expected: PASS.

- [ ] **Step 8: Verify complexity, then commit**

Run: `just complexity`
Expected: `Complexity OK.`

```bash
git add validator/lint_writable_paths.go validator/lint_writable_paths_test.go validator/lint_codes.go validator/explanations.go validator/lint.go
git commit -m "feat(validator): DIP141 — writable_paths nullified by tool_access: none

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

## Task 12: Validator — DIP142 (unsafe entry)

**Files:**
- Modify: `validator/lint_writable_paths.go`
- Modify: `validator/explanations.go`
- Test: `validator/lint_writable_paths_test.go`

- [ ] **Step 1: Write the failing DIP142 tests**

Append to `validator/lint_writable_paths_test.go`:

```go
func TestLint_DIP142_UnsafeEntries(t *testing.T) {
	cases := []struct {
		name, entry string
	}{
		{"absolute", "/etc/**"},
		{"home", "~/secrets/**"},
		{"windows drive", `C:\Users\x`},
		{"parent escape", "../../etc/**"},
		{"brace mis-split", "*.{md"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			src := "workflow X\n  start: A\n  exit: A\n\n  agent A\n    prompt: \"x\"\n    writable_paths: " + tc.entry + "\n"
			if !hasCode(lintSrc(t, src), DIP142) {
				t.Errorf("expected DIP142 for %q; got: %v", tc.entry, codes(lintSrc(t, src)))
			}
		})
	}
}

func TestLint_DIP142_SafeEntries(t *testing.T) {
	src := `workflow X
  start: A
  exit: A

  agent A
    prompt: "x"
    writable_paths: workspace/**, .ai/sprints/**, .ai/managers/recovery-journal.md
`
	if hasCode(lintSrc(t, src), DIP142) {
		t.Errorf("DIP142 should not fire for relative globs; got: %v", codes(lintSrc(t, src)))
	}
}

func TestLint_DIP142_Branch(t *testing.T) {
	src := `workflow X
  start: split
  exit: join

  agent a
    prompt: "a"

  parallel split
    branch: a
      writable_paths: /etc/**

  fan_in join <- a

  edges
    split -> a
    a -> join
`
	if !hasCode(lintSrc(t, src), DIP142) {
		t.Errorf("expected DIP142 on branch; got: %v", codes(lintSrc(t, src)))
	}
}
```

- [ ] **Step 2: Run the tests — expect FAIL**

Run: `just test-pkg validator`
Expected: FAIL — no DIP142 emitted (the unsafe-entry check doesn't exist yet).

- [ ] **Step 3: Implement the DIP142 check**

In `validator/lint_writable_paths.go`, add `"path/filepath"` to the import block. In `checkWritablePathsObject`, after the DIP141 block and before `return diags`:

```go
	for _, entry := range paths {
		if kind := unsafeEntryKind(entry); kind != "" {
			diags = append(diags, dip142Diagnostic(n, branch, entry, kind))
		}
	}
```

(`checkWritablePathsObject` goes cyclo 3 → 5: +for +if. At cap, OK.)

Add the predicate + diagnostic helpers at the end of the file:

```go
// unsafeEntryKind classifies a writable_paths entry that will not bound writes as
// the author expects. Returns "" for a safe relative glob. This is a lexical
// clarity check, not the security boundary — the runtime fs-jail is authoritative.
func unsafeEntryKind(entry string) string {
	e := strings.TrimSpace(entry)
	if e == "" {
		return ""
	}
	if isAbsoluteEntry(e) {
		return "absolute"
	}
	if isParentEscape(e) {
		return "escape"
	}
	if strings.Count(e, "{") != strings.Count(e, "}") {
		return "brace"
	}
	return ""
}

func isAbsoluteEntry(e string) bool {
	return strings.HasPrefix(e, "/") || strings.HasPrefix(e, "~") ||
		strings.HasPrefix(e, "\\") || hasWindowsDrive(e)
}

func hasWindowsDrive(e string) bool {
	return len(e) >= 2 && e[1] == ':' &&
		((e[0] >= 'A' && e[0] <= 'Z') || (e[0] >= 'a' && e[0] <= 'z'))
}

// isParentEscape reports whether the cleaned entry contains a `..` segment that
// would escape its base. Local copy of parser.hasParentRef (validator imports ir only).
func isParentEscape(e string) bool {
	cleaned := filepath.ToSlash(filepath.Clean(e))
	for _, seg := range strings.Split(cleaned, "/") {
		if seg == ".." {
			return true
		}
	}
	return false
}

func dip142Diagnostic(n *ir.Node, branch, entry, kind string) Diagnostic {
	reason := writablePathReason(kind)
	msg := fmt.Sprintf("node %q writable_paths entry %q %s", n.ID, entry, reason)
	if branch != "" {
		msg = fmt.Sprintf("node %q branch %q writable_paths entry %q %s", n.ID, branch, entry, reason)
	}
	return Diagnostic{
		Code:     DIP142,
		Severity: SeverityWarning,
		Message:  msg,
		Location: n.Source,
		Help:     "use workspace-relative globs (e.g. .ai/sprints/**). Absolute, ~, and ..-escaping entries are rejected by the fs jail (it bounds writes to the session root); this lint catches obvious lexical cases only — the runtime jail is the real boundary. See #67/#77.",
	}
}

func writablePathReason(kind string) string {
	if kind == "brace" {
		return "is malformed (brace expansion is split on commas — enumerate entries instead)"
	}
	return "escapes the workspace (absolute / ~ / parent path) — the runtime write-jail will not honor it"
}
```

Complexity: `unsafeEntryKind` cyclo 5; `isAbsoluteEntry` cyclo 4; `hasWindowsDrive` cyclo 4; `isParentEscape` cyclo 3; `dip142Diagnostic` cyclo 2; `writablePathReason` cyclo 2. All ≤ 5.

- [ ] **Step 4: Add the DIP142 explanation**

In `validator/explanations.go`, after the `DIP141` entry (before line 414's closing `}`):

```go
		DIP142: {
			Code:    DIP142,
			Summary: "unsafe writable_paths entry (absolute / ~ / parent-escape / brace)",
			Trigger: "A writable_paths entry is an absolute path, starts with ~ or a Windows drive, escapes its base via .., or is a brace-expansion fragment torn apart by comma-splitting. Such an entry will not bound writes to the workspace the way the author expects.",
			Fix:     "Use workspace-relative globs (e.g. .ai/sprints/**). Absolute, ~, Windows-drive, and ..-escaping entries are rejected by the runtime fs jail; brace expansion (*.{md,yaml}) is split on commas — enumerate entries instead. This lint catches obvious lexical cases only; the runtime jail is the real boundary.",
			Example: "agent Recorder\n  prompt: \"record\"\n  writable_paths: /etc/**   // DIP142: absolute path escapes the workspace jail",
		},
```

- [ ] **Step 5: Run the tests — expect PASS**

Run: `just test-pkg validator`
Expected: PASS.

- [ ] **Step 6: Verify complexity, then commit**

Run: `just complexity`
Expected: `Complexity OK.`

```bash
git add validator/lint_writable_paths.go validator/lint_writable_paths_test.go validator/explanations.go
git commit -m "feat(validator): DIP142 — unsafe writable_paths entry (absolute/~/escape/brace)

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

## Task 13: Example file + dedicated lint-clean test

**Files:**
- Create: `examples/agent_writable_paths.dip`
- Test: `validator/lint_writable_paths_test.go`

- [ ] **Step 1: Write the failing lint-clean test**

Append to `validator/lint_writable_paths_test.go`:

```go
func TestExampleAgentWritablePathsLintsClean(t *testing.T) {
	data, err := os.ReadFile("../examples/agent_writable_paths.dip")
	if err != nil {
		t.Fatalf("read example: %v", err)
	}
	diags := lintSrc(t, string(data))
	if hasCode(diags, DIP141) || hasCode(diags, DIP142) {
		t.Errorf("example should be DIP141/DIP142-clean; got: %v", codes(diags))
	}
}
```

(Add `"os"` to the test file's import block.)

- [ ] **Step 2: Run the test — expect FAIL (file missing)**

Run: `just test-pkg validator`
Expected: FAIL — `read example: open ../examples/agent_writable_paths.dip: no such file or directory`.

- [ ] **Step 3: Create the example (exercises all 5 acceptance shapes)**

Create `examples/agent_writable_paths.dip`:

```dippin
workflow AgentWritablePaths
  goal: "Demonstrate path-bounded write scope on failure-recorder agents (issue #75)"
  start: Review
  exit: RecoveryManager

  agent Review
    model: claude-sonnet-4-6
    prompt: "Review the work. Emit STATUS: success or STATUS: fail."
    auto_status: true

  agent L6Failed
    model: claude-sonnet-4-6
    prompt: "Record the review failure to the sentinel file."
    writable_paths: workspace/.review-failed

  agent Gate1Failed
    model: claude-sonnet-4-6
    prompt: "Record the synthesis failure."
    reads: gate1-findings.md
    writable_paths: workspace/.synthesis-failed

  agent Gate2Failed
    model: claude-sonnet-4-6
    prompt: "Record the validation failure."
    writable_paths: workspace/.validation-failed

  agent RecoveryManager
    model: claude-sonnet-4-6
    prompt: "Write the recovery analysis, the redecompose request, and the journal entry."
    writable_paths: .ai/sprints/**, .ai/managers/recovery-journal.md

  edges
    Review -> L6Failed
    Review -> Gate1Failed
    L6Failed -> Gate2Failed
    Gate1Failed -> Gate2Failed
    Gate2Failed -> RecoveryManager
```

- [ ] **Step 4: Run the dedicated test + the example validator — expect PASS**

Run: `just test-pkg validator`
Expected: PASS.

Run: `just validate-examples`
Expected: all examples validate (DIP001–009 clean). If the edge `condition:` or graph shape trips a structural error, simplify the edges (e.g. use plain `Review -> L6Failed`) until `validate` is clean — the field, not the graph, is what this example demonstrates.

Run: `just lint-examples`
Expected: no DIP141/DIP142 on the new file (other advisory warnings are acceptable; `lint-examples` is non-blocking).

- [ ] **Step 5: Commit**

```bash
git add examples/agent_writable_paths.dip validator/lint_writable_paths_test.go
git commit -m "docs(examples): add agent_writable_paths.dip exercising the 5 motivating shapes

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

## Task 14: Docs, LSP completion, and code-count updates

**Files:**
- Modify: `lsp/completion.go` (~line 57)
- Modify: `docs/nodes.md`
- Modify: `site/static/skill.md`
- Modify: `docs/validation.md`
- Modify: `CLAUDE.md` (~line 85)
- Modify: `docs/llm-reference.md` (~line 188)

- [ ] **Step 1: Add the LSP completion entry**

In `lsp/completion.go`, in the `fields` slice, after the `tool_access:` entry (line 57):

```go
		{"writable_paths:", "Glob list bounding where this agent's tools may write (native backend; runtime-enforced fs jail)"},
```

- [ ] **Step 2: Update `docs/nodes.md`**

Add `writable_paths` to the agent-node field list and the per-branch field list. Include: a comma-separated glob list bounding writes (e.g. `workspace/**, .ai/sprints/**`); the inherit-on-empty rule for branches (empty inherits the target agent's, never resets to unbounded); the no-brace-expansion limitation (comma-split, so `*.{md,yaml}` is not expressible — enumerate entries); and one sentence on the `writes:` vs `writable_paths:` distinction (`writes:` = advisory context keys produced; `writable_paths:` = enforced file globs).

- [ ] **Step 3: Update `site/static/skill.md`**

Add a `writable_paths` entry to the agent-node field section documenting: the glob-list shape; native-backend enforcement, fail-closed (empty/malformed → deny-all/refuse), and the immutable session-root anchor; the residual-escape scope (network / content-within-path / reads are out of scope); the chain caveat and cross-node non-goals with links to #55/#53/#56; the **safety requirement that an enforcing runtime is required** (an older runtime does not enforce it — refuse/pin, do not run unbounded); and one sentence that it is settable per-branch (pointing to the agent-level semantics).

- [ ] **Step 4: Update `docs/validation.md`**

Add DIP141 and DIP142 entries (Trigger / Fix / Example), copying the text from the `explanations.go` entries written in Tasks 11/12.

- [ ] **Step 5: Update the diagnostic-code counts**

In `CLAUDE.md` (~line 85): change `49 diagnostic codes` → `51 diagnostic codes` and `DIP101-DIP140` → `DIP101-DIP142`.
In `docs/llm-reference.md` (~line 188): change `49 diagnostic codes` → `51 diagnostic codes` (and any `DIP101-DIP140` range string → `DIP101-DIP142`).

- [ ] **Step 6: Build to confirm `lsp` compiles, then commit**

Run: `just build`
Expected: success.

```bash
git add lsp/completion.go docs/nodes.md site/static/skill.md docs/validation.md CLAUDE.md docs/llm-reference.md
git commit -m "docs: document writable_paths (nodes/skill/validation), LSP completion, code counts

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

## Task 15: Full gate + PR

- [ ] **Step 1: Run the full suite**

Run: `just check`
Expected: all green (build, vet, fmt, test-race, complexity, validate-examples). Fix anything that fails before proceeding — do not open the PR on a red gate.

- [ ] **Step 2: Push the branch**

```bash
git push -u origin feat/75-write-paths
```

- [ ] **Step 3: Open the PR**

```bash
gh pr create --title "feat: writable_paths path-bounded write-scope primitive (#75)" --body "$(cat <<'EOF'
Fixes #75.

Adds `writable_paths` — a comma-separated glob list on agent nodes and parallel
branches that bounds where an agent's tools may write. dippin **carries + lints**
the field (DIP141 dead-config, DIP142 unsafe entry); **the runtime enforces** an
fs-level write jail (Bash + children) on the native backend. Orthogonal to the
`tool_access` scalar; fails **closed** on empty/malformed (runtime deny-all/refuse).

Design: `docs/superpowers/specs/2026-05-29-issue-75-writable-paths-design.md`
Plan: `docs/superpowers/plans/2026-05-29-issue-75-writable-paths.md`

**Coordinated runtime release.** This is the dippin side. The runtime's fs-jail enforcement
(native-only; claude-code/acp refuse-to-start; immutable session-root anchor;
fail-closed on empty/unrecognized; symlink-chain resolution; `Params`/`working_dir`
bypass defense; red-team suite) ships in a paired runtime PR. **A runtime older than
the paired tag does not enforce `writable_paths` and must refuse rather than run
unbounded** — requiring an enforcing runtime is a safety requirement, recorded at tag time.

Sequenced follow-ups (not in this PR): #55 (tool-name allowlists, orthogonal axis),
#53 (defaults cascade), #56 (chain-attack mitigation).

🤖 Generated with [Claude Code](https://claude.com/claude-code)
EOF
)"
```

Expected: PR created on `feat/75-write-paths`. Downstream consumers see this only after a release tag is cut (v0.35.0).

---

## Self-review checklist (run before declaring the plan done)

- **Spec coverage:** IR (T1), parser agent+branch (T2/T3), formatter agent+branch (T4/T5), DOT export agent+branch (T6/T7), migrate agent+branch (T8/T9), round-trip (T10), DIP141 (T11), DIP142 (T12), example + dedicated lint test (T13), docs/skill/validation/counts/LSP (T14), follow-up gate (T0), release/PR (T15). Fail-closed (`splitCommaNoEmpty`) in T2/T3; native-only + runtime contract documented in T14 skill.md; parity non-comparable fix in T1.
- **Runtime-side** work (fs-jail, backends, red-team) is intentionally OUT of this plan — separate repo/PR per the spec.
- **Complexity:** every function touched that was at cyclo 5 (`applyBranchField`, `writeBranch`, `writeBranchFields`, `applyRuntimeAttrs`) is refactored with helpers in its task; new lint helpers are each ≤ 5.
- **Type consistency:** field is `WritablePaths []string` everywhere; helpers `branchesEqual`/`branchScalarsEqual`/`branchHasFields`/`writeBranchScalarFields`/`applyRuntimeSafetyAttrs`/`unsafeEntryKind`/`isAbsoluteEntry`/`hasWindowsDrive`/`isParentEscape` are each defined once.
