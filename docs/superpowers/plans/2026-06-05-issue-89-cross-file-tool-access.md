# DIP146 Cross-File `tool_access` Detection — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Detect, at `dippin lint` time, when a tool-restricting workflow delegates across a `subgraph_ref`/`ref` file boundary to a child workflow that restricts no agent's `tool_access` — emitting `DIP146` (Hint) and superseding the per-file `DIP143` advisory where the child was conclusively resolved.

**Architecture:** The cross-file traversal lives in `cmd/dippin` (the CLI layer — the only tier allowed to compose `parser` + `validator` + `ir`, never built for wasm). `validator` only gains the `DIP146` const + explanation (pure wasm-safe data) and exports its existing `tool_access` intent predicates so the new pass reuses them (no logic drift). `CmdLint` runs the pass after `validator.Lint()`, filters `DIP143` for resolved boundaries, and appends `DIP146`.

**Tech Stack:** Go; `just` recipes for all build/test; parser-driven `.dip` fixtures; complexity caps cyclo ≤5 / cognitive ≤7 (no `//nolint`).

**Design spec:** `docs/superpowers/specs/2026-06-05-issue-89-design.md` (read it first — it carries the decided forks and the per-child posture table).

**Worktree:** `/home/clint/code/2389/dippin-lang/.claude/worktrees/feat+89-cross-file-tool-access` (branch `feat/89-cross-file-tool-access`). Run every command from there with absolute paths; the shell cwd resets between commands. Stage explicit paths — never `git add -A` (it would sweep gitignored `./dippin`, `./wasm`, `site/static/*.wasm`).

---

## File map

| File | Responsibility | Action |
| --- | --- | --- |
| `validator/lint_subgraph_tool_access.go` | DIP143 + the shared intent predicates | **Modify**: export `WorkflowDeclaresToolAccess` / `NodeDeclaresToolAccess`; refresh DIP143 help text |
| `validator/lint_codes.go` | DIP code consts + `CodeDescription` registry | **Modify**: add `DIP146`; bump range comment |
| `validator/explanations.go` | `dippin explain` text | **Modify**: add `DIP146` entry; bump group comment; refresh DIP143 `#89` line |
| `cmd/dippin/crossfile_tool_access.go` | Native cross-file traversal + classification + DIP146 emission | **Create** |
| `cmd/dippin/crossfile_tool_access_test.go` | Unit tests (parser-driven temp-dir fixtures) | **Create** |
| `cmd/dippin/cmd_validate.go` | `CmdLint` wiring (supersession + `.dipx` skip) | **Modify** |
| `cmd/dippin/cmd_validate_crossfile_test.go` | `CmdLint` integration test | **Create** |
| `CLAUDE.md`, `docs/validation.md`, `docs/llm-reference.md`, `site/static/skill.md` | Counts + DIP146 docs | **Modify** |

---

## Task 1: Export the `tool_access` intent predicates from `validator`

So the new CLI pass reuses DIP143's *exact* intent logic (no second implementation that can drift). Pure rename + call-site update; no behavior change.

**Files:**
- Modify: `validator/lint_subgraph_tool_access.go`

- [ ] **Step 1: Run the existing DIP143 tests (green baseline)**

Run: `cd /home/clint/code/2389/dippin-lang/.claude/worktrees/feat+89-cross-file-tool-access && just test-pkg validator 2>&1 | tail -5`
Expected: `ok` (the package passes before any change).

- [ ] **Step 2: Export `workflowDeclaresToolAccess`**

In `validator/lint_subgraph_tool_access.go`, replace the function (note the call inside it to the node predicate is also renamed):

```go
// WorkflowDeclaresToolAccess reports whether any node in w expresses tool_access
// containment intent (an agent or parallel branch with a non-empty tool_access).
// Exported so the CLI's cross-file pass (DIP146) reuses DIP143's exact intent
// logic — wasm-safe, pure IR.
func WorkflowDeclaresToolAccess(w *ir.Workflow) bool {
	for _, n := range w.Nodes {
		if NodeDeclaresToolAccess(n) {
			return true
		}
	}
	return false
}

// NodeDeclaresToolAccess reports whether an agent (or any parallel branch)
// declares a non-empty tool_access value.
func NodeDeclaresToolAccess(n *ir.Node) bool {
	switch cfg := n.Config.(type) {
	case ir.AgentConfig:
		return toolAccessSet(cfg.ToolAccess)
	case ir.ParallelConfig:
		return branchesDeclareToolAccess(cfg.Branches)
	default:
		return false
	}
}
```

- [ ] **Step 3: Update the internal caller**

In the same file, `lintSubgraphToolAccess` calls the predicate. Change:

```go
	if !workflowDeclaresToolAccess(w) {
		return nil
	}
```

to:

```go
	if !WorkflowDeclaresToolAccess(w) {
		return nil
	}
```

(`branchesDeclareToolAccess` and `toolAccessSet` stay unexported — they are internal helpers.)

- [ ] **Step 4: Verify it compiles and the package still passes**

Run: `cd /home/clint/code/2389/dippin-lang/.claude/worktrees/feat+89-cross-file-tool-access && just test-pkg validator 2>&1 | tail -5`
Expected: `ok` — identical behavior, renamed symbols.

- [ ] **Step 5: Commit**

```bash
cd /home/clint/code/2389/dippin-lang/.claude/worktrees/feat+89-cross-file-tool-access
git add validator/lint_subgraph_tool_access.go
git commit -m "refactor: export tool_access intent predicates for cross-file reuse (#89)

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

## Task 2: Register `DIP146` (const + description + explanation)

The const + a fully-populated explanation are **atomic** — `TestExplanationsCoverAllCodes`/`NoExtra` go red if either is missing. DIP146 is registered here but **emitted by the CLI**, never by `validator.Lint()` (the parity tests don't require Lint-reachability — verified).

**Files:**
- Modify: `validator/lint_codes.go`
- Modify: `validator/explanations.go`

- [ ] **Step 1: Add the const + registry entry**

In `validator/lint_codes.go`:

1. Line 3 comment — bump the registry range:

```go
// Diagnostic codes for semantic quality warnings (DIP101–DIP146).
```

2. In the `const (...)` block, after the `DIP145` line, add (the comment records the CLI-emitted nature so a future maintainer doesn't "fix" it):

```go
	// DIP146 is emitted by the CLI's native cross-file pass (cmd/dippin), NOT by
	// validator.Lint() — the validator cannot read the child .dip. Registered here
	// only so it appears in the catalog / `dippin explain` / docs.
	DIP146 = "DIP146" // child subgraph re-grants tools the parent restricted (cross-file)
```

3. After the `CodeDescription[DIP145] = ...` line, add:

```go
	CodeDescription[DIP146] = "child subgraph re-grants tools the parent restricted (cross-file)"
```

- [ ] **Step 2: Add the explanation entry**

In `validator/explanations.go`, in `safetyExplanations()`:

1. Bump the function's group comment (line ~413):

```go
// safetyExplanations returns explanations for tool-safety lint rules (DIP138–DIP146).
```

2. Add a `DIP146` entry immediately after the `DIP143` entry (before the closing `}` of the returned map):

```go
		DIP146: {
			Code:    DIP146,
			Summary: "child subgraph re-grants tools the parent restricted (cross-file)",
			Trigger: "Surfaced by `dippin lint` (native, cross-file) — not by validator.Lint(), which cannot read the child. A workflow on the path from the linted entry down to a manager_loop (subgraph_ref) or subgraph (ref) boundary declares tool_access on some agent or branch (containment intent), and the resolved child workflow declares NO tool_access restriction on any of its agents. The pass reads and confirms the child (the precision win over DIP143) and traverses transitively, so a zero-intent grandchild is flagged too. A child that restricts every agent is silent; one that restricts some but leaves a tool-bearing agent open keeps the DIP143 advisory instead. Hint, not Warning: a fully-open child may be intentional (a trusted tool worker), and there is no per-diagnostic suppression.",
			Fix:     "Give the referenced .dip's agents their own tool_access (e.g. tool_access: none on summarizers). Restrictions declared in a parent do not flow into a referenced subgraph. This bounds the child's tool catalog, not information flow across the supervisory boundary (steer_context / stack.child.* — see #56). A clean result means every resolvable child is fully locked down or had no restriction to escape — not that the child's tools are enforced at runtime.",
			Example: "agent Lock\n  tool_access: none\nmanager_loop Supervise\n  subgraph_ref: worker.dip   // DIP146 if worker.dip restricts no agent's tool_access",
		},
```

- [ ] **Step 3: Refresh DIP143's `#89`-deferred wording (now that #89 ships)**

In `validator/explanations.go`, the `DIP143` entry's `Trigger` ends with `Cross-file effective-access enforcement is deferred (#89).` Replace that final sentence with:

```
Native `dippin lint` resolves the child and may upgrade this to DIP146 (gap) or silence; DIP143 is the filesystem-free advisory (e.g. the wasm playground) or the fallback when the child cannot be resolved.
```

In `validator/lint_subgraph_tool_access.go`, `checkSubgraphBoundary`'s `Help` string currently ends `Cross-file enforcement is tracked as #89.` Replace that trailing sentence with:

```
Native `dippin lint` resolves the child (DIP146); this advisory is the filesystem-free check or the fallback when the child cannot be resolved.
```

(Do **not** bump `validator/lint.go:8`'s comment — that sentence describes the range `Lint()` itself emits, and `Lint()` does not emit DIP146.)

- [ ] **Step 4: Run validator tests (parity must pass) + wasm build**

Run: `cd /home/clint/code/2389/dippin-lang/.claude/worktrees/feat+89-cross-file-tool-access && just test-pkg validator 2>&1 | tail -8 && just wasm && echo WASM_OK`
Expected: `ok` for validator (parity tests green); `WASM_OK` (the pure-data additions compile for wasm).

- [ ] **Step 5: Commit**

```bash
cd /home/clint/code/2389/dippin-lang/.claude/worktrees/feat+89-cross-file-tool-access
git add validator/lint_codes.go validator/explanations.go validator/lint_subgraph_tool_access.go
git commit -m "feat: register DIP146 cross-file tool_access code (catalog + explanation) (#89)

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

## Task 3: The cross-file traversal (`crossFileToolAccess`)

TDD: write the fixture-driven test first, watch it fail (undefined function), then implement the whole pass. The pass is a plain function — unit-tested directly with real temp-dir `.dip` files parsed by the real parser.

**Files:**
- Create: `cmd/dippin/crossfile_tool_access_test.go`
- Create: `cmd/dippin/crossfile_tool_access.go`

- [ ] **Step 1: Write the failing test file**

Create `cmd/dippin/crossfile_tool_access_test.go`:

```go
package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/2389-research/dippin-lang/ir"
	"github.com/2389-research/dippin-lang/validator"
)

// --- fixtures: real .dip files parsed by the real parser ---

// entryRestrictsRefs is an entry that locks down agent Lock and delegates to a
// child via manager_loop. %s is the child ref (a file name in the same dir).
func entryRestrictsRefs(childRef string) string {
	return `workflow Entry
  start: Lock
  exit: Sup

  agent Lock
    prompt: "x"
    tool_access: none

  manager_loop Sup
    subgraph_ref: ` + childRef + `
    max_cycles: 3

  edges
    Lock -> Sup
`
}

// entryNoIntentRefs delegates to a child but declares no tool_access itself.
func entryNoIntentRefs(childRef string) string {
	return `workflow Entry
  start: Go
  exit: Sup

  agent Go
    prompt: "x"

  manager_loop Sup
    subgraph_ref: ` + childRef + `
    max_cycles: 3

  edges
    Go -> Sup
`
}

// Children mirror the proven minimalDip shape (human start -> agent, with edges)
// so the parser accepts them; only the agents' tool_access differs.
const childZeroIntent = `workflow Child
  start: Ask
  exit: Done

  human Ask
    mode: freeform

  agent Done
    prompt: "do it"

  edges
    Ask -> Done
`

const childFullRestrict = `workflow Child
  start: Ask
  exit: Done

  human Ask
    mode: freeform

  agent Done
    prompt: "do it"
    tool_access: none

  edges
    Ask -> Done
`

const childPartial = `workflow Child
  start: Ask
  exit: B

  human Ask
    mode: freeform

  agent A
    prompt: "a"
    tool_access: none

  agent B
    prompt: "b"

  edges
    Ask -> A
    A -> B
`

const childAgentless = `workflow Child
  start: Ask
  exit: Done

  human Ask
    mode: freeform

  human Done
    mode: freeform

  edges
    Ask -> Done
`

// writeWorkflows writes name->content into a fresh temp dir and returns the dir.
func writeWorkflows(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	for name, content := range files {
		p := filepath.Join(dir, name)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", p, err)
		}
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			t.Fatalf("write %s: %v", p, err)
		}
	}
	return dir
}

// crossDiags loads dir/entry through the real loader and runs the pass.
func crossDiags(t *testing.T, dir, entry string) ([]validator.Diagnostic, map[ir.SourceLocation]childPosture) {
	t.Helper()
	entryPath := filepath.Join(dir, entry)
	w, err := loadWorkflow(entryPath)
	if err != nil {
		t.Fatalf("load %s: %v", entry, err)
	}
	return crossFileToolAccess(w, entryPath)
}

func countCode(diags []validator.Diagnostic, code string) int {
	n := 0
	for _, d := range diags {
		if d.Code == code {
			n++
		}
	}
	return n
}

func TestCrossFile_FiresOnZeroIntentChild(t *testing.T) {
	dir := writeWorkflows(t, map[string]string{
		"entry.dip": entryRestrictsRefs("child.dip"),
		"child.dip": childZeroIntent,
	})
	diags, classified := crossDiags(t, dir, "entry.dip")
	if got := countCode(diags, validator.DIP146); got != 1 {
		t.Fatalf("want 1 DIP146, got %d: %v", got, diags)
	}
	found := false
	for _, p := range classified {
		if p == postureZeroIntent {
			found = true
		}
	}
	if !found {
		t.Errorf("expected a zero-intent classification: %v", classified)
	}
}

func TestCrossFile_SilentOnFullRestrict(t *testing.T) {
	dir := writeWorkflows(t, map[string]string{
		"entry.dip": entryRestrictsRefs("child.dip"),
		"child.dip": childFullRestrict,
	})
	diags, _ := crossDiags(t, dir, "entry.dip")
	if got := countCode(diags, validator.DIP146); got != 0 {
		t.Fatalf("want 0 DIP146 (child fully restricted), got %d", got)
	}
}

func TestCrossFile_SilentOnPartialAudit(t *testing.T) {
	dir := writeWorkflows(t, map[string]string{
		"entry.dip": entryRestrictsRefs("child.dip"),
		"child.dip": childPartial,
	})
	diags, classified := crossDiags(t, dir, "entry.dip")
	if got := countCode(diags, validator.DIP146); got != 0 {
		t.Fatalf("want 0 DIP146 (partial audit keeps DIP143), got %d", got)
	}
	found := false
	for _, p := range classified {
		if p == posturePartial {
			found = true
		}
	}
	if !found {
		t.Errorf("expected a partial classification: %v", classified)
	}
}

func TestCrossFile_SilentWhenNoPathIntent(t *testing.T) {
	dir := writeWorkflows(t, map[string]string{
		"entry.dip": entryNoIntentRefs("child.dip"),
		"child.dip": childZeroIntent,
	})
	diags, _ := crossDiags(t, dir, "entry.dip")
	if got := countCode(diags, validator.DIP146); got != 0 {
		t.Fatalf("want 0 DIP146 (no intent on path), got %d", got)
	}
}

func TestCrossFile_AncestorPathIntent(t *testing.T) {
	// entry (no intent) -> mid (restricts) -> leaf (zero intent) => DIP146 on mid->leaf.
	mid := `workflow Mid
  start: Lock
  exit: Sup

  agent Lock
    prompt: "x"
    tool_access: none

  manager_loop Sup
    subgraph_ref: leaf.dip
    max_cycles: 3

  edges
    Lock -> Sup
`
	dir := writeWorkflows(t, map[string]string{
		"entry.dip": entryNoIntentRefs("mid.dip"),
		"mid.dip":   mid,
		"leaf.dip":  childZeroIntent,
	})
	diags, _ := crossDiags(t, dir, "entry.dip")
	if got := countCode(diags, validator.DIP146); got != 1 {
		t.Fatalf("want 1 DIP146 (ancestor-path intent), got %d: %v", got, diags)
	}
}

func TestCrossFile_Transitive(t *testing.T) {
	// entry (restricts) -> child (restricts) -> grandchild (zero) => DIP146 on grandchild edge.
	child := `workflow Child
  start: Lock
  exit: Sup

  agent Lock
    prompt: "x"
    tool_access: none

  manager_loop Sup
    subgraph_ref: grand.dip
    max_cycles: 3

  edges
    Lock -> Sup
`
	dir := writeWorkflows(t, map[string]string{
		"entry.dip": entryRestrictsRefs("child.dip"),
		"child.dip": child,
		"grand.dip": childZeroIntent,
	})
	diags, _ := crossDiags(t, dir, "entry.dip")
	if got := countCode(diags, validator.DIP146); got != 1 {
		t.Fatalf("want 1 DIP146 on the grandchild edge, got %d: %v", got, diags)
	}
}

func TestCrossFile_DiamondTwoFindings(t *testing.T) {
	// P1 and P2 both reference the same zero-intent child => 2 DIP146 (one per edge).
	parent := func(name, ref string) string {
		return `workflow ` + name + `
  start: Lock
  exit: Sup

  agent Lock
    prompt: "x"
    tool_access: none

  manager_loop Sup
    subgraph_ref: ` + ref + `
    max_cycles: 3

  edges
    Lock -> Sup
`
	}
	entry := `workflow Entry
  start: P1
  exit: P2

  subgraph P1
    ref: p1.dip

  subgraph P2
    ref: p2.dip

  edges
    P1 -> P2
`
	dir := writeWorkflows(t, map[string]string{
		"entry.dip": entry,
		"p1.dip":    parent("P1", "child.dip"),
		"p2.dip":    parent("P2", "child.dip"),
		"child.dip": childZeroIntent,
	})
	diags, _ := crossDiags(t, dir, "entry.dip")
	if got := countCode(diags, validator.DIP146); got != 2 {
		t.Fatalf("want 2 DIP146 (one per parent->child edge), got %d: %v", got, diags)
	}
}

func TestCrossFile_CycleTerminates(t *testing.T) {
	// A -> B -> A. Must terminate; A has intent so no DIP146 on the B->A edge.
	a := `workflow A
  start: Lock
  exit: Sup

  agent Lock
    prompt: "x"
    tool_access: none

  manager_loop Sup
    subgraph_ref: b.dip
    max_cycles: 3

  edges
    Lock -> Sup
`
	b := `workflow B
  start: Go
  exit: Sup

  human Go
    mode: freeform

  manager_loop Sup
    subgraph_ref: a.dip
    max_cycles: 3

  edges
    Go -> Sup
`
	dir := writeWorkflows(t, map[string]string{
		"a.dip": a,
		"b.dip": b,
	})
	// Completing at all proves termination (no hang/stack overflow).
	diags, _ := crossDiags(t, dir, "a.dip")
	_ = diags
}

func TestCrossFile_SelfReferenceTerminates(t *testing.T) {
	// A -> A. Must terminate; A is agent-less here so no DIP146 either way.
	a := `workflow A
  start: Go
  exit: Sup

  human Go
    mode: freeform

  manager_loop Sup
    subgraph_ref: a.dip
    max_cycles: 3

  edges
    Go -> Sup
`
	dir := writeWorkflows(t, map[string]string{"a.dip": a})
	diags, _ := crossDiags(t, dir, "a.dip")
	if got := countCode(diags, validator.DIP146); got != 0 {
		t.Fatalf("want 0 DIP146 on self-reference, got %d", got)
	}
}

func TestCrossFile_UnresolvableChildIsSilent(t *testing.T) {
	dir := writeWorkflows(t, map[string]string{
		"entry.dip": entryRestrictsRefs("missing.dip"), // no missing.dip on disk
	})
	diags, classified := crossDiags(t, dir, "entry.dip")
	if got := countCode(diags, validator.DIP146); got != 0 {
		t.Fatalf("want 0 DIP146 for unresolvable child, got %d", got)
	}
	for _, p := range classified {
		if p != postureUnresolved {
			t.Errorf("want postureUnresolved, got %v", p)
		}
	}
}

func TestCrossFile_AgentlessChildIsSilent(t *testing.T) {
	dir := writeWorkflows(t, map[string]string{
		"entry.dip": entryRestrictsRefs("child.dip"),
		"child.dip": childAgentless,
	})
	diags, _ := crossDiags(t, dir, "entry.dip")
	if got := countCode(diags, validator.DIP146); got != 0 {
		t.Fatalf("want 0 DIP146 (agent-less child has no tools to grant), got %d", got)
	}
}

// Known intentional false positive: a fully-tooled worker pool fires DIP146.
// This is mitigated by Hint severity, NOT by the gate. Do not "fix" this into a
// Warning or add a heuristic — see the design spec (#89).
func TestCrossFile_IntentionalOpenWorkerFires(t *testing.T) {
	dir := writeWorkflows(t, map[string]string{
		"entry.dip": entryRestrictsRefs("worker.dip"),
		"worker.dip": childZeroIntent,
	})
	diags, _ := crossDiags(t, dir, "entry.dip")
	if got := countCode(diags, validator.DIP146); got != 1 {
		t.Fatalf("want 1 DIP146 (known intentional-open FP, Hint-mitigated), got %d", got)
	}
}
```

- [ ] **Step 2: Run the test to verify it fails to compile**

Run: `cd /home/clint/code/2389/dippin-lang/.claude/worktrees/feat+89-cross-file-tool-access && go test ./cmd/dippin/ -run TestCrossFile 2>&1 | tail -8`
Expected: FAIL — `undefined: crossFileToolAccess`, `undefined: childPosture`, `undefined: postureZeroIntent`, etc.

- [ ] **Step 3: Implement the pass**

Create `cmd/dippin/crossfile_tool_access.go`:

```go
package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/2389-research/dippin-lang/ir"
	"github.com/2389-research/dippin-lang/validator"
)

// crossFileMaxDepth bounds subgraph-ref recursion (mirrors dipx's
// maxManifestDepth). Deeper chains stop silently: the boundary stays
// unresolved, so its DIP143 advisory is retained.
const crossFileMaxDepth = 32

// childPosture classifies a referenced child workflow's tool_access stance.
// The zero value is postureUnresolved, so a missing classification never
// supersedes a DIP143 advisory.
type childPosture int

const (
	postureUnresolved   childPosture = iota // missing / unparseable / depth-capped
	postureZeroIntent                       // >=1 agent, nothing restricts tools
	postureFullRestrict                     // >=1 agent, every agent restricted
	posturePartial                          // some restricted, >=1 agent still open
	postureAgentless                        // no agent nodes (no tools to grant)
)

// crossFileToolAccess walks every subgraph/manager_loop boundary reachable from
// entry, emitting DIP146 (Hint) when a workflow on the path restricts tools and a
// resolved child restricts none. It also returns, per boundary-node location, the
// child's posture, so CmdLint can supersede the per-file DIP143 advisory. entryPath
// is the on-disk path entry was parsed from (seeds the visited set).
func crossFileToolAccess(entry *ir.Workflow, entryPath string) ([]validator.Diagnostic, map[ir.SourceLocation]childPosture) {
	var diags []validator.Diagnostic
	classified := map[ir.SourceLocation]childPosture{}
	visited := map[string]bool{}
	if key := canonicalKey(entryPath); key != "" {
		visited[key] = true
	}
	intentSeen := validator.WorkflowDeclaresToolAccess(entry)
	walkBoundaries(entry, intentSeen, 0, visited, &diags, classified)
	return diags, classified
}

// walkBoundaries inspects each subgraph/manager_loop boundary in w.
func walkBoundaries(w *ir.Workflow, intentSeen bool, depth int, visited map[string]bool, diags *[]validator.Diagnostic, classified map[ir.SourceLocation]childPosture) {
	for _, n := range w.Nodes {
		_, ref := boundaryKindRef(n)
		if ref == "" || n.Source.File == "" {
			continue
		}
		visitBoundary(n, ref, intentSeen, depth, visited, diags, classified)
	}
}

// visitBoundary resolves one boundary's child, records its posture, emits DIP146
// when the path shows intent and the child is zero-intent, then recurses.
func visitBoundary(n *ir.Node, ref string, intentSeen bool, depth int, visited map[string]bool, diags *[]validator.Diagnostic, classified map[ir.SourceLocation]childPosture) {
	child, childPath := resolveBoundaryChild(n, ref)
	if child == nil {
		classified[n.Source] = postureUnresolved
		return
	}
	posture := classifyChild(child)
	classified[n.Source] = posture
	if intentSeen && posture == postureZeroIntent {
		*diags = append(*diags, boundaryDiag(n, ref))
	}
	maybeRecurse(child, childPath, intentSeen, depth, visited, diags, classified)
}

// maybeRecurse descends into a child's own boundaries unless it was already
// visited (cycle guard) or the depth cap is reached. The child is marked visited
// before recursing (pre-order) so cycles terminate.
func maybeRecurse(child *ir.Workflow, childPath string, intentSeen bool, depth int, visited map[string]bool, diags *[]validator.Diagnostic, classified map[ir.SourceLocation]childPosture) {
	key := canonicalKey(childPath)
	if key == "" || visited[key] || depth+1 >= crossFileMaxDepth {
		return
	}
	visited[key] = true
	childIntent := intentSeen || validator.WorkflowDeclaresToolAccess(child)
	walkBoundaries(child, childIntent, depth+1, visited, diags, classified)
}

// resolveBoundaryChild resolves ref relative to the boundary node's source file
// and parses the child workflow. Fail-soft: any error yields (nil, "").
// Callers (walkBoundaries) guarantee ref != "" and n.Source.File != "".
func resolveBoundaryChild(n *ir.Node, ref string) (*ir.Workflow, string) {
	path := resolveBoundaryRefPath(ref, n.Source.File)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, ""
	}
	w, err := parseAndResolveDip(data, path)
	if err != nil {
		return nil, ""
	}
	return w, path
}

// resolveBoundaryRefPath resolves a (possibly relative) ref against the parent's
// source file directory, mirroring DIP135's resolveRefPath.
func resolveBoundaryRefPath(ref, sourceFile string) string {
	if filepath.IsAbs(ref) {
		return ref
	}
	return filepath.Join(filepath.Dir(sourceFile), ref)
}

// classifyChild determines a child's tool_access posture from its agent nodes.
func classifyChild(child *ir.Workflow) childPosture {
	agents, restricted := countAgents(child)
	switch {
	case agents == 0:
		return postureAgentless
	case !validator.WorkflowDeclaresToolAccess(child):
		return postureZeroIntent
	case restricted == agents:
		return postureFullRestrict
	default:
		return posturePartial
	}
}

// countAgents returns the number of agent nodes and how many declare a tool_access
// restriction. Branches can only restrict (never grant more than their target
// agent), so the agent census is the right basis for full-restrict detection.
func countAgents(child *ir.Workflow) (agents, restricted int) {
	for _, n := range child.Nodes {
		if _, ok := n.Config.(ir.AgentConfig); !ok {
			continue
		}
		agents++
		if validator.NodeDeclaresToolAccess(n) {
			restricted++
		}
	}
	return agents, restricted
}

// canonicalKey returns a symlink-resolved absolute key for the visited set, so a
// file reached via different paths (or a symlink cycle) maps to one key.
func canonicalKey(path string) string {
	if path == "" {
		return ""
	}
	if resolved, err := filepath.EvalSymlinks(path); err == nil {
		return resolved
	}
	if abs, err := filepath.Abs(path); err == nil {
		return abs
	}
	return filepath.Clean(path)
}

// boundaryKindRef returns the node-kind label ("manager_loop"/"subgraph") and the
// external ref for boundary nodes, or ("","") otherwise.
func boundaryKindRef(n *ir.Node) (string, string) {
	switch cfg := n.Config.(type) {
	case ir.ManagerLoopConfig:
		return "manager_loop", cfg.SubgraphRef
	case ir.SubgraphConfig:
		return "subgraph", cfg.Ref
	default:
		return "", ""
	}
}

// boundaryDiag builds the DIP146 hint for a zero-intent child boundary.
func boundaryDiag(n *ir.Node, ref string) validator.Diagnostic {
	kind, _ := boundaryKindRef(n)
	return validator.Diagnostic{
		Code:     validator.DIP146,
		Severity: validator.SeverityHint,
		Message: fmt.Sprintf(
			"%s %q delegates to subgraph %q, which declares no tool_access restriction on any agent; a workflow on this path restricts tools, but the restriction does not cross the subgraph boundary",
			kind, n.ID, ref),
		Location: n.Source,
		Help:     "give the referenced .dip's agents their own tool_access (e.g. tool_access: none); this bounds the child's tool catalog, not information flow across the supervisory boundary (see #56). Multiple boundaries referencing the same child each get a hint; one tool_access edit clears them all.",
	}
}
```

(`applyCrossFileToolAccess` and `supersedes` — the `CmdLint` glue — are added in Task 4, alongside their caller, so this file has no unused functions at this commit.)

- [ ] **Step 4: Run the pass tests**

Run: `cd /home/clint/code/2389/dippin-lang/.claude/worktrees/feat+89-cross-file-tool-access && go test ./cmd/dippin/ -run TestCrossFile -v 2>&1 | tail -30`
Expected: all `TestCrossFile_*` PASS.

- [ ] **Step 5: Check complexity (helpers must fit cyclo ≤5 / cognitive ≤7)**

Run: `cd /home/clint/code/2389/dippin-lang/.claude/worktrees/feat+89-cross-file-tool-access && just complexity 2>&1 | tail -5`
Expected: no violations. If any helper is over, split it (e.g. extract a guard) — do **not** add `//nolint`.

- [ ] **Step 6: Commit**

```bash
cd /home/clint/code/2389/dippin-lang/.claude/worktrees/feat+89-cross-file-tool-access
git add cmd/dippin/crossfile_tool_access.go cmd/dippin/crossfile_tool_access_test.go
git commit -m "feat: cross-file tool_access traversal + DIP146 classification (#89)

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

## Task 4: Wire the pass into `CmdLint` (supersession + integration test)

**Files:**
- Modify: `cmd/dippin/cmd_validate.go`
- Create: `cmd/dippin/cmd_validate_crossfile_test.go`

- [ ] **Step 1: Write the failing integration test**

Create `cmd/dippin/cmd_validate_crossfile_test.go`:

```go
package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCmdLint_DIP146SupersedesDIP143(t *testing.T) {
	dir := t.TempDir()
	write := func(name, content string) string {
		p := filepath.Join(dir, name)
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
		return p
	}
	entry := write("entry.dip", entryRestrictsRefs("child.dip"))
	write("child.dip", childZeroIntent)

	stdout, _, code := runCLI(t, "lint", entry)
	if code != ExitOK {
		t.Fatalf("want exit 0 (hints don't fail), got %d", code)
	}
	if !strings.Contains(stdout, "DIP146") {
		t.Errorf("expected DIP146 in output:\n%s", stdout)
	}
	if strings.Contains(stdout, "DIP143") {
		t.Errorf("DIP143 should be superseded for a resolved zero-intent child:\n%s", stdout)
	}
}

func TestCmdLint_RetainsDIP143OnPartialChild(t *testing.T) {
	dir := t.TempDir()
	entry := filepath.Join(dir, "entry.dip")
	if err := os.WriteFile(entry, []byte(entryRestrictsRefs("child.dip")), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "child.dip"), []byte(childPartial), 0o644); err != nil {
		t.Fatal(err)
	}
	stdout, _, _ := runCLI(t, "lint", entry)
	if strings.Contains(stdout, "DIP146") {
		t.Errorf("no DIP146 expected for partial-audit child:\n%s", stdout)
	}
	if !strings.Contains(stdout, "DIP143") {
		t.Errorf("DIP143 must be RETAINED for partial-audit child:\n%s", stdout)
	}
}
```

(`entryRestrictsRefs`, `childZeroIntent`, `childPartial` come from `crossfile_tool_access_test.go` — same `package main` test binary.)

- [ ] **Step 2: Run it — fails because the pass isn't wired in**

Run: `cd /home/clint/code/2389/dippin-lang/.claude/worktrees/feat+89-cross-file-tool-access && go test ./cmd/dippin/ -run TestCmdLint_DIP146 2>&1 | tail -12`
Expected: FAIL — output still shows `DIP143`, no `DIP146` (the pass runs but `CmdLint` doesn't call it yet).

- [ ] **Step 3: Add the supersession glue, then wire it into `CmdLint`**

First, append these two functions to `cmd/dippin/crossfile_tool_access.go` and add `"strings"` to its import block (it is now used):

```go
// applyCrossFileToolAccess runs the native cross-file pass, drops the per-file
// DIP143 advisory for boundaries it conclusively resolved (zero-intent,
// full-restrict, or agent-less), and appends the DIP146 findings. Skipped for
// .dipx bundles, whose child refs are in-bundle paths, not on disk. The map
// lookup yields postureUnresolved (the zero value) for any DIP143 not classified
// here, which supersedes() treats as "retain".
func applyCrossFileToolAccess(diags []validator.Diagnostic, w *ir.Workflow, path string) []validator.Diagnostic {
	if strings.HasSuffix(strings.ToLower(path), ".dipx") {
		return diags
	}
	cross, classified := crossFileToolAccess(w, path)
	var kept []validator.Diagnostic
	for _, d := range diags {
		if d.Code == validator.DIP143 && supersedes(classified[d.Location]) {
			continue
		}
		kept = append(kept, d)
	}
	return append(kept, cross...)
}

// supersedes reports whether a resolved child posture makes the per-file DIP143
// advisory redundant: zero-intent is replaced by DIP146; full-restrict and
// agent-less are confirmed safe. Partial-audit and unresolved retain DIP143.
func supersedes(p childPosture) bool {
	return p == postureZeroIntent || p == postureFullRestrict || p == postureAgentless
}
```

Then, in `cmd/dippin/cmd_validate.go`, inside `CmdLint`, replace:

```go
	// Merge all diagnostics.
	allDiags := append(valRes.Diagnostics, lintRes.Diagnostics...)
	c.renderDiagnostics(allDiags)
```

with:

```go
	// Merge all diagnostics, then apply the native cross-file tool_access pass
	// (DIP146) — it supersedes DIP143 for boundaries whose child it resolved.
	allDiags := append(valRes.Diagnostics, lintRes.Diagnostics...)
	allDiags = applyCrossFileToolAccess(allDiags, w, path)
	c.renderDiagnostics(allDiags)
```

- [ ] **Step 4: Run the integration tests**

Run: `cd /home/clint/code/2389/dippin-lang/.claude/worktrees/feat+89-cross-file-tool-access && go test ./cmd/dippin/ -run TestCmdLint -v 2>&1 | tail -20`
Expected: `TestCmdLint_DIP146SupersedesDIP143` and `TestCmdLint_RetainsDIP143OnPartialChild` PASS (plus existing CmdLint tests still green).

- [ ] **Step 5: Full cmd/dippin package + complexity**

Run: `cd /home/clint/code/2389/dippin-lang/.claude/worktrees/feat+89-cross-file-tool-access && go test ./cmd/dippin/ 2>&1 | tail -3 && just complexity 2>&1 | tail -3`
Expected: `ok` and no complexity violations.

- [ ] **Step 6: Commit**

```bash
cd /home/clint/code/2389/dippin-lang/.claude/worktrees/feat+89-cross-file-tool-access
git add cmd/dippin/cmd_validate.go cmd/dippin/crossfile_tool_access.go cmd/dippin/cmd_validate_crossfile_test.go
git commit -m "feat: wire DIP146 cross-file pass into dippin lint with DIP143 supersession (#89)

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

## Task 5: Documentation + count strings + spec generation

Convention-only counts (no test guards them) plus the `dippin explain`/spec source docs. Skipping any leaves the docs lying.

**Files:**
- Modify: `CLAUDE.md`, `docs/validation.md`, `docs/llm-reference.md`, `site/static/skill.md`

- [ ] **Step 1: `CLAUDE.md` counts**

In `CLAUDE.md` line ~85, change `54 diagnostic codes` → `55 diagnostic codes` and `DIP101-DIP145 (semantic warnings)` → `DIP101-DIP146 (semantic warnings)`.

- [ ] **Step 2: `docs/validation.md` counts + ranges**

- Line 3: `registers 54 diagnostic codes` → `registers 55 diagnostic codes`; `documents 49 of them` → `documents 50 of them`.
- Lines 6, 15, 227, 1070: change each `DIP101–DIP145` → `DIP101–DIP146` (note the en-dash `–`, not a hyphen).

- [ ] **Step 3: `docs/validation.md` — new DIP146 section**

Immediately after the `### DIP145: Negative Graph Budget Default` block (the `---` that ends it, before `## Running Validation`), insert:

```markdown
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

**What DIP146 does NOT check:** partial-audit or unparseable children more than
one hop deep (the DIP143 fallback covers only the entry's direct boundaries);
information flow across the supervisory boundary (`steer_context` /
`stack.child.*` — see [#56](https://github.com/2389-research/dippin-lang/issues/56));
runtime enforcement (the tracker runtime enforces; dippin detects). A clean
result means every resolvable child is fully locked down or had no restriction to
escape — not that the child's tools are restricted at runtime.

---
```

- [ ] **Step 4: `docs/llm-reference.md`**

- Line 188: `54 diagnostic codes across two categories:` → `55 diagnostic codes across two categories:`.
- Line 191: change `DIP101–DIP145` → `DIP101–DIP146`, and append to the end of that description sentence: `, cross-file subgraph tool_access`.

- [ ] **Step 5: `site/static/skill.md` — extend the Subgraph boundary paragraph**

In the `**Subgraph boundary:**` paragraph (line ~106), the final sentence ends `...cross-file enforcement is tracked as [#89](...).` Replace that trailing clause with:

```
...and native `dippin lint` now resolves the child across the file boundary: [DIP146](https://2389-research.github.io/dippin-lang/validation.html#dip146) (Hint) fires when a resolved child restricts no agent's `tool_access` while a workflow on the path does, superseding DIP143 for boundaries it can resolve ([#89](https://github.com/2389-research/dippin-lang/issues/89)). DIP143 remains the filesystem-free advisory (e.g. the wasm playground) and the fallback when the child can't be resolved.
```

- [ ] **Step 6: Regenerate the spec and verify**

Run: `cd /home/clint/code/2389/dippin-lang/.claude/worktrees/feat+89-cross-file-tool-access && just spec-check 2>&1 | tail -5`
Expected: passes (or regenerates `cmd/dippin/generated-spec.md` cleanly). If it reports drift, run the generator it names, then re-run `just spec-check`.

- [ ] **Step 7: Confirm `dippin explain DIP146` works**

Run: `cd /home/clint/code/2389/dippin-lang/.claude/worktrees/feat+89-cross-file-tool-access && just build 2>&1 | tail -2 && ./dippin explain DIP146 2>&1 | head -20`
Expected: prints the DIP146 summary/trigger/fix/example (proves catalog registration + `explain` wiring).

- [ ] **Step 8: Commit**

```bash
cd /home/clint/code/2389/dippin-lang/.claude/worktrees/feat+89-cross-file-tool-access
git add CLAUDE.md docs/validation.md docs/llm-reference.md site/static/skill.md cmd/dippin/generated-spec.md
git commit -m "docs: document DIP146 + bump diagnostic-code counts to 55 (#89)

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

## Task 6: Full verification + PR

**Files:** none (verification + PR).

- [ ] **Step 1: WASM build stays green (the load-bearing split)**

Run: `cd /home/clint/code/2389/dippin-lang/.claude/worktrees/feat+89-cross-file-tool-access && just wasm && echo WASM_OK`
Expected: `WASM_OK` — validator's DIP146 additions are pure data; the filesystem traversal lives only in `cmd/dippin` (never a wasm target).

- [ ] **Step 2: Full test suite (race) + complexity + lint-examples**

Run: `cd /home/clint/code/2389/dippin-lang/.claude/worktrees/feat+89-cross-file-tool-access && just test-race 2>&1 | tail -6 && just complexity 2>&1 | tail -3 && just lint-examples 2>&1 | tail -5`
Expected: all `ok`; no complexity violations; `lint-examples` exit 0 (DIP146 is a Hint — confirm it didn't turn any example red; if an example now emits an *unexpected* DIP146, investigate before proceeding).

- [ ] **Step 3: `gofmt`**

Run: `cd /home/clint/code/2389/dippin-lang/.claude/worktrees/feat+89-cross-file-tool-access && just fmt && git diff --stat`
Expected: no formatting changes (or commit them if any: `git add -p` the touched files only).

- [ ] **Step 4: Push the branch**

```bash
cd /home/clint/code/2389/dippin-lang/.claude/worktrees/feat+89-cross-file-tool-access
git push -u origin feat/89-cross-file-tool-access
```

(The pre-commit hook runs the real CI gate on each commit; `just check` ends on a tree-sitter-generate step that fails locally — that's expected, see the `just-check-tree-sitter-gotcha` note. The hook is authoritative.)

- [ ] **Step 5: Open the PR**

```bash
cd /home/clint/code/2389/dippin-lang/.claude/worktrees/feat+89-cross-file-tool-access
gh pr create --title "feat: cross-file tool_access detection (DIP146) — Closes #89" --body "$(cat <<'EOF'
Closes #89. Completes the DIP143 (#59) arc: actually resolves referenced child workflows and detects cross-file tool_access gaps.

## Where the analysis lives
The cross-file traversal lives in `cmd/dippin` (the CLI layer — the only tier that may compose `parser` + `validator` + `ir`, and never built for wasm), mirroring `cmd_pack.go`'s pattern. `validator` cannot do this (no `parser` import; compiles to wasm with no filesystem).

## Diagnostic surface
`DIP146` — const + explanation registered in the validator catalog (so `dippin explain DIP146`, the spec, and docs work) but **emitted from the native CLI pass**, never from `validator.Lint()`. Safe: the parity tests require only `CodeDescription` ↔ `Explanations`, not Lint-reachability.

## Containment rule (conservative)
Fires only when a workflow on the path from the linted entry restricts tools (**ancestor-path** intent gate) **and** a transitively-resolved child restricts **no** agent. A fully-restricted child is silent; a **partial-audit** child (restricts some, leaves a tool-bearing agent open) or an unresolvable child **retains** the DIP143 advisory — "unknown/partial" never reads as "checked & safe."

## Transitivity & cycles
Full transitive DFS; per-edge emission decoupled from per-node recursion (a diamond yields one finding per boundary). Termination: `EvalSymlinks`-keyed visited set (catches symlink cycles) + pre-order marking + a depth cap. No separate cycle diagnostic.

## DIP143 supersession
Native `lint` drops DIP143 for boundaries the pass conclusively resolved (zero-intent → replaced by DIP146; full-restrict/agent-less → confirmed safe) and retains it otherwise. In the wasm playground DIP143 stands alone, as before.

## Scope boundary
Tool-catalog gap only. Information flow (`steer_context` / `stack.child.*`) stays #56. Detection, not runtime enforcement (not gated on the tracker).

## WASM
`validator` additions are pure data; traversal is CLI-only. `just wasm` stays green.

Design + plan: `docs/superpowers/specs/2026-06-05-issue-89-design.md`, `docs/superpowers/plans/2026-06-05-issue-89-cross-file-tool-access.md`. Hardened by a 5-expert review panel (security / architecture / DX / correctness / conventions).

🤖 Generated with [Claude Code](https://claude.com/claude-code)
EOF
)"
```

---

## Post-implementation

Before requesting human review, run the expert squad code-review pass (`review-squad:experts` and/or `/code-review`) on the diff — this is a P2 safety feature with cross-file recursion and false-positive risk. Do **not** tag a release; after merge this batches into a later minor bump per the maintainer's go-ahead.
