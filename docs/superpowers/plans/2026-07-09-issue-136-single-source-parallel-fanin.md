# #136 Single-source `parallel`/`fan_in` Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make the inline `parallel X -> a,b,c` / `fan_in X <- a,b,c` list the single source of truth for fork edges — warn (DIP153) + `fmt`-strip redundant edges-block re-declarations under v1, reject them under `dip 2`, and teach DOT export to draw fan edges from node config so stripped files still render.

**Architecture:** One shared predicate `ir.IsRedundantFanEdge(w, e)` is consumed by three sites (validator lint, formatter strip, parser dip-2 rejection). DOT export gains a synthesis pass so it no longer depends on the redundant edges. Every consumer already reads the inline config for semantics (confirmed in the design spec), so this is additive — no reachability/simulation behavior changes.

**Tech Stack:** Go; `just` recipes; pre-commit hook gate (run `export PATH="/usr/local/go/bin:$PATH"` first so the hook finds `go`).

**Spec:** `docs/superpowers/specs/2026-07-09-issue-136-single-source-parallel-fanin-design.md`

---

### Task 1: `ir.IsRedundantFanEdge` shared predicate

**Files:**
- Modify: `ir/edge.go` (add function near `EdgeRoutesOnFail`)
- Test: `ir/edge_test.go` (or a new `ir/redundant_fan_edge_test.go`)

- [ ] **Step 1: Write the failing test**

Add to the `ir` package tests. Build a workflow with a parallel node `Fan -> A, B` and a fan_in node `Join <- A, B`.

```go
func TestIsRedundantFanEdge(t *testing.T) {
	w := &Workflow{
		Nodes: []*Node{
			{ID: "Fan", Config: ParallelConfig{Targets: []string{"A", "B"}}},
			{ID: "A"}, {ID: "B"},
			{ID: "Join", Config: FanInConfig{Sources: []string{"A", "B"}}},
		},
	}
	cases := []struct {
		name string
		e    *Edge
		want bool
	}{
		{"parallel fork match", &Edge{From: "Fan", To: "A"}, true},
		{"fan_in source match", &Edge{From: "A", To: "Join"}, true},
		{"not a fan edge", &Edge{From: "A", To: "B"}, false},
		{"conditional not redundant", &Edge{From: "Fan", To: "A", Condition: &Condition{Raw: "ctx.x = 1"}}, false},
		{"labeled not redundant", &Edge{From: "Fan", To: "A", Label: "left"}, false},
		{"choice not redundant", &Edge{From: "Fan", To: "A", Choice: "left"}, false},
		{"weighted not redundant", &Edge{From: "Fan", To: "A", Weight: 2}, false},
		{"override not redundant", &Edge{From: "Fan", To: "A", Override: true}, false},
		{"restart not redundant", &Edge{From: "Fan", To: "A", Restart: true}, false},
		{"parallel target not listed", &Edge{From: "Fan", To: "C"}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := IsRedundantFanEdge(w, tc.e); got != tc.want {
				t.Errorf("IsRedundantFanEdge = %v, want %v", got, tc.want)
			}
		})
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `export PATH="/usr/local/go/bin:$PATH"; go test ./ir/ -run TestIsRedundantFanEdge`
Expected: FAIL — `undefined: IsRedundantFanEdge`.

- [ ] **Step 3: Write the implementation**

Add to `ir/edge.go`:

```go
// IsRedundantFanEdge reports whether e merely repeats a parallel/fan_in fork
// already declared inline on a node's config, carrying no extra information —
// it is unconditional and attribute-free, and either From is a parallel node
// listing To as a target, or To is a fan_in node listing From as a source.
// Such an edge can be stripped without losing information; a conditional or
// attributed edge between the same nodes is NOT redundant.
func IsRedundantFanEdge(w *Workflow, e *Edge) bool {
	if !edgeIsPlain(e) {
		return false
	}
	if from := w.Node(e.From); from != nil {
		if cfg, ok := from.Config.(ParallelConfig); ok && contains(cfg.Targets, e.To) {
			return true
		}
	}
	if to := w.Node(e.To); to != nil {
		if cfg, ok := to.Config.(FanInConfig); ok && contains(cfg.Sources, e.From) {
			return true
		}
	}
	return false
}

// edgeIsPlain reports whether an edge carries no guard, label, or routing
// attribute — i.e. it conveys only "From connects to To".
func edgeIsPlain(e *Edge) bool {
	return e.Condition == nil && e.Label == "" && e.Choice == "" &&
		e.Weight == 0 && !e.Restart && !e.Override && e.Comment == ""
}

func contains(xs []string, s string) bool {
	for _, x := range xs {
		if x == s {
			return true
		}
	}
	return false
}
```

Note: if a `contains` helper already exists in the `ir` package, reuse it and drop the local copy. Check with `grep -rn "func contains" ir/` first; keep the build free of duplicate-symbol errors.

- [ ] **Step 4: Run test to verify it passes**

Run: `export PATH="/usr/local/go/bin:$PATH"; go test ./ir/ -run TestIsRedundantFanEdge`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add ir/edge.go ir/*_test.go
git commit -m "feat(ir): IsRedundantFanEdge — shared predicate for redundant fan edges"
```

---

### Task 2: DOT export synthesizes fan edges from node config

**Files:**
- Modify: `export/dot.go` (after the `w.Edges` loop at line ~52)
- Test: `export/dot_test.go`

**Why first among consumers:** DOT is the only consumer that reads raw `w.Edges` and never synthesizes fan edges. It must draw forks from config *before* we strip any edges, so a stripped `.dip` still renders arrows.

- [ ] **Step 1: Write the failing test**

A workflow whose parallel/fan_in forks are declared ONLY inline (no edges-block entries) must still emit those edges in DOT.

```go
func TestExportDOT_SynthesizesFanEdges(t *testing.T) {
	w := &ir.Workflow{
		Name: "W",
		Nodes: []*ir.Node{
			{ID: "Fan", Config: ir.ParallelConfig{Targets: []string{"A", "B"}}},
			{ID: "A"}, {ID: "B"},
			{ID: "Join", Config: ir.FanInConfig{Sources: []string{"A", "B"}}},
		},
		// No Edges at all — forks live only in node config.
	}
	dot := ExportDOT(w, ExportOptions{})
	for _, want := range []string{`"Fan" -> "A"`, `"Fan" -> "B"`, `"A" -> "Join"`, `"B" -> "Join"`} {
		if !strings.Contains(dot, want) {
			t.Errorf("DOT missing synthesized fan edge %q\n%s", want, dot)
		}
	}
}

func TestExportDOT_NoDoubleDrawWhenExplicit(t *testing.T) {
	w := &ir.Workflow{
		Name:  "W",
		Nodes: []*ir.Node{{ID: "Fan", Config: ir.ParallelConfig{Targets: []string{"A"}}}, {ID: "A"}},
		Edges: []*ir.Edge{{From: "Fan", To: "A"}},
	}
	dot := ExportDOT(w, ExportOptions{})
	if strings.Count(dot, `"Fan" -> "A"`) != 1 {
		t.Errorf("expected exactly one Fan->A edge, got:\n%s", dot)
	}
}
```

- [ ] **Step 2: Run to verify the first test fails**

Run: `export PATH="/usr/local/go/bin:$PATH"; go test ./export/ -run TestExportDOT_SynthesizesFanEdges`
Expected: FAIL — synthesized edges absent.
(The no-double-draw test should already pass today; it guards the fix.)

- [ ] **Step 3: Implement synthesis pass**

In `export/dot.go`, replace the explicit-edge loop body with a version that also emits synthesized fan edges. After line 54 (`}` closing the `for _, e := range w.Edges` loop), add:

```go
	writeSynthesizedFanEdges(&b, w)
```

Then add the function (dedup against unconditional explicit edges, mirroring `ir.parallelEdgesFrom`/`fanInEdgesFrom`):

```go
// writeSynthesizedFanEdges emits parallel fan-out and fan-in edges that are
// declared inline on node config but absent from w.Edges as an unconditional
// explicit edge. This lets DOT render forks even when the redundant edges-block
// re-declaration has been stripped (issue #136).
func writeSynthesizedFanEdges(b *strings.Builder, w *ir.Workflow) {
	seen := make(map[[2]string]bool)
	for _, e := range w.Edges {
		if e.Condition == nil {
			seen[[2]string{e.From, e.To}] = true
		}
	}
	emit := func(from, to string) {
		key := [2]string{from, to}
		if seen[key] {
			return
		}
		seen[key] = true
		writeEdgeDOT(b, &ir.Edge{From: from, To: to})
	}
	for _, n := range w.Nodes {
		switch cfg := n.Config.(type) {
		case ir.ParallelConfig:
			for _, t := range cfg.Targets {
				emit(n.ID, t)
			}
		case ir.FanInConfig:
			for _, s := range cfg.Sources {
				emit(s, n.ID)
			}
		}
	}
}
```

- [ ] **Step 4: Run to verify both tests pass**

Run: `export PATH="/usr/local/go/bin:$PATH"; go test ./export/`
Expected: PASS (all existing DOT tests unchanged — files that still carry explicit fork edges hit the `seen` dedup, so output is byte-identical).

- [ ] **Step 5: Commit**

```bash
git add export/dot.go export/dot_test.go
git commit -m "feat(export): synthesize parallel/fan_in edges in DOT from node config (#136)"
```

---

### Task 3: DIP153 lint — redundant fan edge

**Files:**
- Modify: `validator/lint_codes.go` (add `DIP153` const + description entry)
- Modify: `validator/explanations.go` (add `DIP153` explanation)
- Create: `validator/lint_redundant_fan_edge.go`
- Modify: `validator/lint.go` (register `lintRedundantFanEdge` in the lint-func list near line 111)
- Test: `validator/lint_redundant_fan_edge_test.go`

**Pattern to mirror:** DIP152 (`validator/lint_marker_coverage.go`, its entries in `lint_codes.go`, `explanations.go`, and its registration in `lint.go`). Follow that file's structure exactly for signature, diagnostic construction, and registration.

- [ ] **Step 1: Write the failing test**

```go
func TestLintRedundantFanEdge(t *testing.T) {
	src := `workflow W
  parallel Fan -> A, B
  agent A
    prompt: "a"
  agent B
    prompt: "b"
  fan_in Join <- A, B
  agent Done
    prompt: "d"
edges
  Fan -> A
  Fan -> B
  A -> Join
  B -> Join
  Join -> Done
`
	w := mustParse(t, src) // use the package's existing parse test helper
	res := Lint(w)
	got := countCode(res, DIP153) // helper: number of diagnostics with this code
	if got != 4 {
		t.Fatalf("want 4 DIP153 (Fan->A, Fan->B, A->Join, B->Join), got %d: %+v", got, res.Diagnostics)
	}
}

func TestLintRedundantFanEdge_ConditionalExempt(t *testing.T) {
	// A conditional edge between the same nodes is NOT redundant.
	src := `workflow W
  parallel Fan -> A, B
  agent A
    prompt: "a"
  agent B
    prompt: "b"
edges
  Fan -> A when ctx.x = 1
  Fan -> B
`
	w := mustParse(t, src)
	res := Lint(w)
	if countCode(res, DIP153) != 1 { // only Fan->B is redundant
		t.Fatalf("want 1 DIP153, got: %+v", res.Diagnostics)
	}
}
```

If `mustParse`/`countCode` helpers don't exist under those names, use whatever the validator test suite already uses (grep the `_test.go` files for the established parse helper and diagnostic-counting idiom — DIP152's test file shows the current convention).

- [ ] **Step 2: Run to verify it fails**

Run: `export PATH="/usr/local/go/bin:$PATH"; go test ./validator/ -run TestLintRedundantFanEdge`
Expected: FAIL — `undefined: DIP153` / `lintRedundantFanEdge`.

- [ ] **Step 3: Add the code + description + explanation + lint function + registration**

`validator/lint_codes.go` — add after DIP152:
```go
	DIP153 = "DIP153" // edges-block edge redundantly repeats an inline parallel/fan_in fork
```
and in the description map:
```go
	DIP153: "edges-block edge redundantly repeats an inline parallel/fan_in fork (the inline list is authoritative)",
```

`validator/explanations.go` — add a `DIP153` entry mirroring DIP152's shape (Code, Summary, Detail, Example, Fix). Example:
```go
		DIP153: {
			Code:    DIP153,
			// Summary/Detail: the inline parallel/fan_in list is the single source of
			// truth; an unconditional edges-block edge repeating it is redundant. Run
			// `dippin fmt` to strip it. Rejected under `dip 2`.
			Example: "parallel Fan -> A, B\nedges\n  Fan -> A  # DIP153: redundant — remove; the inline list is authoritative",
		},
```
(Match the exact field set the other explanation entries use — the explanation-parity test enforces every code has an entry.)

`validator/lint_redundant_fan_edge.go`:
```go
package validator

import (
	"fmt"

	"github.com/2389-research/dippin-lang/ir"
)

// lintRedundantFanEdge flags edges-block edges that redundantly repeat an inline
// parallel fan-out or fan_in join (DIP153). The inline node-config list is the
// single source of truth; such an unconditional, attribute-free edge conveys
// nothing new and is stripped by `dippin fmt` (rejected under `dip 2`).
func lintRedundantFanEdge(w *ir.Workflow) []Diagnostic {
	var out []Diagnostic
	for _, e := range w.Edges {
		if !ir.IsRedundantFanEdge(w, e) {
			continue
		}
		out = append(out, Diagnostic{
			Code:     DIP153,
			Severity: SeverityWarning,
			Message: fmt.Sprintf(
				"edges-block edge '%s -> %s' redundantly repeats the inline parallel/fan_in fork; the inline list is authoritative — run 'dippin fmt' to remove it (rejected under 'dip 2')",
				e.From, e.To),
			Location: e.Source,
		})
	}
	return out
}
```
Match the real `Diagnostic` struct field names + severity constant + location field to the DIP152 lint file (adjust `Severity`/`Location`/constructor to whatever that file actually uses).

`validator/lint.go` — register in the lint-func slice near line 111 (next to `lintMarkerCoverage`):
```go
		lintRedundantFanEdge,
```

- [ ] **Step 4: Run to verify it passes + full validator suite**

Run: `export PATH="/usr/local/go/bin:$PATH"; go test ./validator/`
Expected: PASS (including the explanation-parity test now that DIP153 has an entry).

- [ ] **Step 5: Commit**

```bash
git add validator/
git commit -m "feat(validator): DIP153 — redundant parallel/fan_in edge lint (#136)"
```

---

### Task 4: `fmt` strips redundant fan edges

**Files:**
- Modify: `formatter/format.go` (`writeEdges`, line ~744)
- Test: `formatter/format_test.go`

- [ ] **Step 1: Write the failing test**

```go
func TestFormat_StripsRedundantFanEdges(t *testing.T) {
	src := `workflow W
  parallel Fan -> A, B
  agent A
    prompt: "a"
  agent B
    prompt: "b"
edges
  Fan -> A
  Fan -> B
`
	w := mustParse(t, src)
	out := Format(w) // use the package's real entrypoint name
	if strings.Contains(out, "Fan -> A") || strings.Contains(out, "Fan -> B") {
		t.Fatalf("redundant fan edges should be stripped:\n%s", out)
	}
	if !strings.Contains(out, "parallel Fan -> A, B") {
		t.Fatalf("inline parallel line must survive:\n%s", out)
	}
	// Idempotent: formatting the stripped output again is a no-op.
	w2 := mustParse(t, out)
	if Format(w2) != out {
		t.Fatalf("format not idempotent")
	}
}

func TestFormat_KeepsConditionalFanEdge(t *testing.T) {
	src := `workflow W
  parallel Fan -> A, B
  agent A
    prompt: "a"
  agent B
    prompt: "b"
edges
  Fan -> A when ctx.x = 1
  Fan -> B
`
	w := mustParse(t, src)
	out := Format(w)
	if !strings.Contains(out, "Fan -> A") { // conditional edge preserved
		t.Fatalf("conditional fan edge must be preserved:\n%s", out)
	}
	if strings.Contains(out, "Fan -> B\n") && !strings.Contains(out, "when") {
		// Fan -> B (plain) should be stripped
	}
}
```
Use the formatter package's real public entrypoint (grep for `func Format` / `func FormatWorkflow` in `formatter/`); adjust the call name accordingly.

- [ ] **Step 2: Run to verify it fails**

Run: `export PATH="/usr/local/go/bin:$PATH"; go test ./formatter/ -run TestFormat_StripsRedundantFanEdges`
Expected: FAIL — redundant edges still emitted.

- [ ] **Step 3: Implement the skip**

In `formatter/format.go`, `writeEdges`:
```go
	for _, e := range w.Edges {
		if ir.IsRedundantFanEdge(w, e) {
			continue
		}
		writeEdge(wr, w, e)
	}
```
Ensure `ir` is imported (it already is in this file).

- [ ] **Step 4: Run to verify it passes + full formatter suite**

Run: `export PATH="/usr/local/go/bin:$PATH"; go test ./formatter/`
Expected: PASS. If any existing golden/round-trip test now fails because its fixture had redundant fan edges, that's expected — regenerate/adjust that fixture to the stripped form (the stripped form is the new canonical output) and confirm the change is only the removed redundant lines.

- [ ] **Step 5: Commit**

```bash
git add formatter/
git commit -m "feat(formatter): strip redundant parallel/fan_in edges (#136)"
```

---

### Task 5: `dip 2` rejects redundant fan-edge re-declaration

**Files:**
- Modify: `parser/parser.go` (`Parse()`, add a post-parse pass)
- Create/modify: `parser/parse_edges.go` or a small new `parser/reject_v2.go` for the pass
- Test: `parser/parser_test.go` (or the file holding the existing dip-2 rejection tests for `retry_target`)

**Pattern to mirror:** the existing dip-2 rejection of `retry_target`/`fallback_target` (`parser/parse_nodes.go:172`, `isV2RejectedNodeField`). This one is a *post-parse* pass because redundancy is cross-referential (edge vs node config).

- [ ] **Step 1: Write the failing test**

```go
func TestParse_V2RejectsRedundantFanEdge(t *testing.T) {
	src := `dip 2
workflow W
  parallel Fan -> A, B
  agent A
    prompt: "a"
  agent B
    prompt: "b"
edges
  Fan -> A
  Fan -> B
`
	_, err := Parse(src) // use the package's real parse entrypoint
	if err == nil {
		t.Fatal("expected dip 2 to reject redundant fan edges")
	}
	if !strings.Contains(err.Error(), "inline") {
		t.Fatalf("diagnostic should mention the inline list is authoritative, got: %v", err)
	}
}

func TestParse_V1AllowsRedundantFanEdge(t *testing.T) {
	src := strings.Replace(/* same source without the `dip 2` header */ ..., "dip 2\n", "", 1)
	if _, err := Parse(src); err != nil {
		t.Fatalf("v1 must still accept redundant fan edges (warn only), got: %v", err)
	}
}
```
Adjust to the real parse entrypoint and test helper the parser suite uses.

- [ ] **Step 2: Run to verify it fails**

Run: `export PATH="/usr/local/go/bin:$PATH"; go test ./parser/ -run TestParse_V2RejectsRedundantFanEdge`
Expected: FAIL — no error emitted under dip 2.

- [ ] **Step 3: Implement the post-parse pass**

In `parser/parser.go`, `Parse()`, insert before the diagnostics check:
```go
func (p *Parser) Parse() (*ir.Workflow, error) {
	p.parseVersionDeclaration()
	p.parseTopLevel()
	p.rejectRedundantFanEdgesUnderV2()
	if len(p.diagnostics) > 0 {
		return p.workflow, fmt.Errorf("parsing errors: %s", strings.Join(p.diagnostics, "; "))
	}
	return p.workflow, nil
}
```

Add the pass (in `parser.go` or a new `parser/reject_v2.go`):
```go
// rejectRedundantFanEdgesUnderV2 reports, under dip 2, any edges-block edge that
// merely repeats an inline parallel/fan_in fork — the inline list is the single
// source of truth under dip 2 (issue #136). No-op under v1 (DIP153 warns there).
func (p *Parser) rejectRedundantFanEdgesUnderV2() {
	if p.version < 2 {
		return
	}
	for _, e := range p.workflow.Edges {
		if ir.IsRedundantFanEdge(p.workflow, e) {
			p.diagnostics = append(p.diagnostics, fmt.Sprintf(
				"redundant edge %q -> %q: the inline parallel/fan_in list is authoritative under dip 2 — remove it (run `dippin fmt`) at %d:%d",
				e.From, e.To, e.Source.Line, e.Source.Column))
		}
	}
}
```
Confirm `e.Source` exposes `.Line`/`.Column` (it's an `ir.SourceLocation`); adjust field access to the real shape.

- [ ] **Step 4: Run to verify it passes + full parser suite**

Run: `export PATH="/usr/local/go/bin:$PATH"; go test ./parser/`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add parser/
git commit -m "feat(parser): reject redundant parallel/fan_in edges under dip 2 (#136)"
```

---

### Task 6: Strip examples + `TestLintExamples` DIP153 guard

**Files:**
- Modify: example `.dip` files that redundantly re-declare fan edges (`examples/fanin_policy.dip`, `examples/consensus_task_parity.dip`, and any others surfaced by the check below — do NOT hand-edit; use `dippin fmt --write`)
- Modify: `validator/lint_examples_test.go` (add DIP153 zero-assertion)

- [ ] **Step 1: Build the binary and find affected examples**

```bash
export PATH="/usr/local/go/bin:$PATH"
just build
for f in examples/*.dip; do
  n=$(./dippin lint "$f" 2>&1 | grep -c DIP153 || true)
  [ "$n" != "0" ] && echo "$f: $n"
done
```
Record the list — these are the files Task 6 strips.

- [ ] **Step 2: Strip via fmt --write and verify only fan edges were removed**

```bash
for f in <affected files>; do ./dippin fmt --write "$f"; done
git diff --stat examples/
git diff examples/  # every removed line must be a redundant fan edge; inline parallel/fan_in lines intact
```
Re-run the loop from Step 1 — expect zero DIP153 across all examples. Run `just validate-examples` — expect all pass.

- [ ] **Step 3: Add the guard test**

In `validator/lint_examples_test.go`, alongside the existing DIP108 zero-assertion, assert zero DIP153 across the example suite (mirror the exact idiom already there for DIP108).

- [ ] **Step 4: Run the integration test**

Run: `export PATH="/usr/local/go/bin:$PATH"; go test ./validator/ -run TestLintExamples`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add examples/ validator/lint_examples_test.go
git commit -m "chore(examples): strip redundant fan edges + DIP153 example guard (#136)"
```

---

### Task 7: Docs / site / embedded-spec sweep

**Files:**
- Modify: `docs/nodes.md`, `docs/edges.md`, `docs/validation.md`, `docs/cli.md`
- Modify: `site/content/cli.md` (DIP-count 62 → 63; note fmt strips redundant fan edges)
- Leave `site/content/validation.md` for the release docs(site) pass (hand-maintained; per `release-process`)
- Regenerate: `cmd/dippin/generated-spec.md` (via the pre-commit hook / `scripts/gen-spec.sh`)

- [ ] **Step 1: Edit docs**

- `docs/nodes.md` / `docs/edges.md`: state that the inline `parallel X -> a,b,c` / `fan_in X <- a,b,c` list is the single source of truth for fork edges; re-declaring them in the `edges` block is redundant (DIP153) and rejected under `dip 2`. Note DOT still renders forks from config.
- `docs/validation.md`: add the DIP153 row/section (mirror the DIP152 entry's format).
- `docs/cli.md` + `site/content/cli.md`: under `fmt`, note it strips redundant `parallel`/`fan_in` edges; bump the lint-count from 62 to 63 wherever it appears (grep both files for `62`).

- [ ] **Step 2: Regenerate spec + verify currency**

```bash
export PATH="/usr/local/go/bin:$PATH"
bash scripts/gen-spec.sh    # or let the pre-commit hook do it
go test ./releasecheck/     # currency gate must pass
```

- [ ] **Step 3: Commit**

```bash
git add docs/ site/content/cli.md cmd/dippin/generated-spec.md
git commit -m "docs: single-source parallel/fan_in + DIP153 (#136)"
```

---

## Final verification (after all tasks)

```bash
export PATH="/usr/local/go/bin:$PATH"
.git/hooks/pre-commit   # full CI-equivalent gate: build, vet, lint, test, complexity, validate-examples
```
Expected: all checks pass. Then run the squad review before landing.

## Self-review notes
- Spec coverage: Task 1 = predicate; Task 2 = DOT fix; Task 3 = DIP153; Task 4 = fmt strip; Task 5 = dip 2 reject; Task 6 = examples + guard; Task 7 = docs. All acceptance criteria mapped.
- Type consistency: `IsRedundantFanEdge(w, e)` signature identical across Tasks 1/3/4/5. `edgeIsPlain` covers every `ir.Edge` routing field enumerated from `ir/edge.go`.
- Known adjustment points flagged inline (real entrypoint/helper names in each package) — implementers must grep for the actual `Format`/`Parse`/`Diagnostic`/test-helper names rather than assume.
