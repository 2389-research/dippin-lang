# DIP152 Marker-Coverage Lint — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add DIP152 — warn when a tool node's `marker_grep` enumerates a literal marker that no edge routes and no `else`/unconditional edge covers.

**Architecture:** A new additive lint pass (`validator/lint_marker_coverage.go`). It enumerates markers only from recognizable literal-alternation regexes (else stays quiet), computes routing ultra-conservatively (any non-simple edge makes the node safe → no false positives), and reuses the existing `ir.ExtractEqualityCondition` / `hasValidElseDefault` / `hasUnconditionalEdge` helpers. DIP101/DIP102 keep their blanket marker exemption (no double-warnings).

**Tech Stack:** Go; the `validator` package; `just` recipes; pre-commit hook is the authoritative gate.

**Spec:** `docs/superpowers/specs/2026-07-07-issue-156-marker-coverage-design.md`

**Environment:** every shell step assumes `export PATH=/usr/local/go/bin:$HOME/go/bin:$PATH`. Commit via `git commit` (the pre-commit hook runs the full suite — do NOT use `--no-verify`). Avoid `just check` (fails on tree-sitter generate here); use the individual recipes.

---

## File Structure

- **Create** `validator/lint_marker_coverage.go` — the pass + all helpers.
- **Create** `validator/lint_marker_coverage_test.go` — parser-driven tests.
- **Modify** `validator/lint_codes.go` — DIP152 const + description + range comment.
- **Modify** `validator/explanations.go` — DIP152 explanation entry.
- **Modify** `validator/lint.go` — register the pass + bump the range comment.
- **Modify** `validator/lint_examples_test.go` — zero-assert DIP152 (the required guard).
- **Modify** docs/site count+range surfaces (Task 6) + regenerate embedded spec.

---

### Task 0: Reconfirm the code number

- [ ] **Step 1: Confirm DIP152 is still free on `main`**

Run: `git grep -hoE 'DIP15[0-9]' -- validator/lint_codes.go | sort -u`
Expected: highest is `DIP151`. (Confirmed free at implementation time — `DIP152` was allocated and is used throughout this branch. Had the sibling issue #136 landed first and taken `DIP152`, this would have been renumbered.)

---

### Task 1: Register DIP152 (const, description, explanation)

**Files:**
- Modify: `validator/lint_codes.go` (const block ends at DIP151 ~line 58; `linterCodeDescriptions` map; header range comment ~line 3)
- Modify: `validator/explanations.go` (`Explanations` map, after the DIP151 entry ~line 175)
- Test: `validator/explanations_test.go` (existing `TestExplanationsCoverAllCodes` / `TestExplanationsNoExtra`)

- [ ] **Step 1: Add the constant.** In `validator/lint_codes.go`, after the `DIP151 = ...` line inside the `const (...)` block, add:

```go
	DIP152 = "DIP152" // marker_grep enumerates a marker not routed by any edge or else default
```

- [ ] **Step 2: Add the description.** In the `linterCodeDescriptions` map, after the `DIP151:` entry, add:

```go
	DIP152: "marker_grep enumerates a marker no edge routes and no else default covers",
```

- [ ] **Step 3: Add the explanation.** In `validator/explanations.go`, inside `var Explanations = map[string]Explanation{ ... }`, after the `DIP151: { ... },` block, add:

```go
		DIP152: {
			Code:    DIP152,
			Summary: CodeDescription[DIP152],
			Trigger: "A tool node's marker_grep enumerates a literal marker (e.g. \"tests-failed\" in ^(tests-ok|tests-failed)$) that no outgoing edge routes and that no section `else ->` default or unconditional fallback edge covers, so that marker would be emitted at runtime with nowhere to go. Only checked when marker_grep is a recognizable literal alternation; complex regexes are left unflagged.",
			Fix:     "Route the marker with an edge (`Node -> Target on <marker>`), add an unconditional fallback edge, or add a section `else -> <node>` default.",
			Example: "tool RunTests\n  marker_grep: ^(tests-ok|tests-failed)$\nedges\n  RunTests -> Done on tests-ok  # DIP152: tests-failed is emitted but routed nowhere",
		},
```

- [ ] **Step 4: Run the parity tests — expect PASS** (a registered code with a full explanation and no emitter yet is fine).

Run: `go test ./validator/ -run 'TestExplanations' -v`
Expected: PASS (`TestExplanationsCoverAllCodes`, `TestExplanationsNoExtra`).

- [ ] **Step 5: Confirm `dippin explain DIP152` works.**

Run: `go run ./cmd/dippin explain DIP152`
Expected: prints the summary/trigger/fix/example above.

- [ ] **Step 6: Commit.**

```bash
git add validator/lint_codes.go validator/explanations.go
git commit -m "feat(validator): register DIP152 marker-coverage code + explanation"
```

---

### Task 2: `enumerateMarkers` + helpers (marker set extraction)

**Files:**
- Create: `validator/lint_marker_coverage.go`
- Test: `validator/lint_marker_coverage_test.go`

- [ ] **Step 1: Write the failing unit test.** Create `validator/lint_marker_coverage_test.go`:

```go
package validator

import (
	"reflect"
	"sort"
	"testing"
)

func TestEnumerateMarkers(t *testing.T) {
	cases := []struct {
		in    string
		want  []string
		ok    bool
	}{
		{"^(tests-ok|tests-failed)$", []string{"tests-failed", "tests-ok"}, true},
		{"^(a|b|c)$", []string{"a", "b", "c"}, true},
		{"(a|b)", []string{"a", "b"}, true},
		{"tests_pass", []string{"tests_pass"}, true},
		{"^pass$", []string{"pass"}, true},
		{"^(a|a|b)$", []string{"a", "b"}, true}, // dedup
		// non-enumerable → (nil,false):
		{"^(a|b|)$", nil, false},   // empty branch
		{"(a|b)|c", nil, false},    // group not full-span
		{"(a|b)?", nil, false},     // quantifier after group
		{"^(a.b|c)$", nil, false},  // metachar in branch
		{"^\\d+$", nil, false},     // metachars
		{".*fail.*", nil, false},   // metachars
		{"(?i)(a|b)", nil, false},  // flags: inner group has metachars, not full-span
	}
	for _, tc := range cases {
		got, ok := enumerateMarkers(tc.in)
		if ok != tc.ok {
			t.Errorf("%q: ok = %v, want %v", tc.in, ok, tc.ok)
			continue
		}
		if ok {
			sort.Strings(got)
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("%q: markers = %v, want %v", tc.in, got, tc.want)
			}
		}
	}
}
```

- [ ] **Step 2: Run it — expect FAIL** (undefined `enumerateMarkers`).

Run: `go test ./validator/ -run TestEnumerateMarkers`
Expected: FAIL / build error `undefined: enumerateMarkers`.

- [ ] **Step 3: Implement.** Create `validator/lint_marker_coverage.go`:

```go
// ABOUTME: DIP152 — a tool node's marker_grep enumerates a literal marker that
// ABOUTME: no edge routes and no else/unconditional edge covers.
package validator

import (
	"fmt"
	"sort"
	"strings"

	"github.com/2389-research/dippin-lang/ir"
)

// markerMetachars are the regex metacharacters that make a branch non-literal.
const markerMetachars = ".*+?[]{}()|\\^$"

// enumerateMarkers returns the literal marker set of a marker_grep value when it
// is a recognizable finite literal alternation (optional ^...$ anchors around a
// single literal token or a full-span (a|b|c) group of literal branches).
// Returns (nil, false) for anything else — those keep the blanket DIP101/DIP102
// exemption and get no DIP152.
func enumerateMarkers(markerGrep string) ([]string, bool) {
	branches, ok := splitAlternation(stripAnchors(markerGrep))
	if !ok {
		return nil, false
	}
	seen := make(map[string]struct{})
	var markers []string
	for _, b := range branches {
		if b == "" || !isLiteralToken(b) {
			return nil, false
		}
		if _, dup := seen[b]; dup {
			continue
		}
		seen[b] = struct{}{}
		markers = append(markers, b)
	}
	return markers, true
}

// stripAnchors removes one leading ^ and one trailing $ if present.
func stripAnchors(s string) string {
	return strings.TrimSuffix(strings.TrimPrefix(s, "^"), "$")
}

// splitAlternation returns the branches of a full-span (a|b|c) group, or the
// whole string as a single branch when it is not a full-span group. Empty input
// is non-enumerable.
func splitAlternation(s string) ([]string, bool) {
	if s == "" {
		return nil, false
	}
	if strings.HasPrefix(s, "(") && strings.HasSuffix(s, ")") {
		return strings.Split(s[1:len(s)-1], "|"), true
	}
	return []string{s}, true
}

// isLiteralToken reports whether s contains no regex metacharacter.
func isLiteralToken(s string) bool {
	return !strings.ContainsAny(s, markerMetachars)
}
```

- [ ] **Step 4: Run — expect PASS.**

Run: `go test ./validator/ -run TestEnumerateMarkers -v`
Expected: PASS.

- [ ] **Step 5: Complexity check.**

Run: `just complexity`
Expected: `Complexity OK.` (If a function trips it, the extractions above are already the minimal split — recheck the `&&`/`||` counts.)

- [ ] **Step 6: Commit.**

```bash
git add validator/lint_marker_coverage.go validator/lint_marker_coverage_test.go
git commit -m "feat(validator): enumerateMarkers for DIP152 (literal-alternation only)"
```

---

### Task 3: Coverage logic + register the pass

**Files:**
- Modify: `validator/lint_marker_coverage.go`
- Modify: `validator/lint.go` (pass list ~lines 54-60; `Lint` doc range comment ~line 8)
- Test: `validator/lint_marker_coverage_test.go`

- [ ] **Step 1: Write the failing lint test.** Append to `validator/lint_marker_coverage_test.go`:

```go
// lintFor parses .dip source and returns the diagnostics for `code`.
func markerDiagsFor(t *testing.T, src string) []Diagnostic {
	t.Helper()
	w, err := parser.NewParser(src, "test.dip").Parse()
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	var out []Diagnostic
	for _, d := range Lint(w).Diagnostics {
		if d.Code == DIP152 {
			out = append(out, d)
		}
	}
	return out
}

const markerBase = `workflow W
  goal: "t"
  start: RunTests
  exit: Done

  tool RunTests
    command: run
    marker_grep: "%s"

  agent Done
    prompt: done

  edges
%s`

func mk(grep, edges string) string { return fmt.Sprintf(markerBase, grep, edges) }

func TestDIP152_GapWarns(t *testing.T) {
	diags := markerDiagsFor(t, mk("^(go|stop)$", "    RunTests -> Done on go\n"))
	if len(diags) != 1 {
		t.Fatalf("want 1 DIP152, got %d: %v", len(diags), diags)
	}
	if !strings.Contains(diags[0].Message, "stop") {
		t.Errorf("message should name unrouted marker stop: %q", diags[0].Message)
	}
}

func TestDIP152_FullyRoutedClean(t *testing.T) {
	diags := markerDiagsFor(t, mk("^(go|stop)$",
		"    RunTests -> Done on go\n    RunTests -> Done on stop\n"))
	if len(diags) != 0 {
		t.Fatalf("want 0 DIP152, got %v", diags)
	}
}

func TestDIP152_ElseCovered(t *testing.T) {
	diags := markerDiagsFor(t, mk("^(go|stop)$",
		"    RunTests -> Done on go\n    else -> Done\n"))
	if len(diags) != 0 {
		t.Fatalf("else default should suppress DIP152, got %v", diags)
	}
}

func TestDIP152_UnconditionalCovered(t *testing.T) {
	diags := markerDiagsFor(t, mk("^(go|stop)$",
		"    RunTests -> Done on go\n    RunTests -> Done\n"))
	if len(diags) != 0 {
		t.Fatalf("unconditional edge should suppress DIP152, got %v", diags)
	}
}

func TestDIP152_MultiMarkerGap(t *testing.T) {
	diags := markerDiagsFor(t, mk("^(a|b|c)$", "    RunTests -> Done on a\n"))
	if len(diags) != 1 || !strings.Contains(diags[0].Message, "b, c") {
		t.Fatalf("want DIP152 listing b, c: %v", diags)
	}
}

// --- false-positive guards (squad blockers) ---

func TestDIP152_OrRoutingClean(t *testing.T) {
	diags := markerDiagsFor(t, mk("^(go|stop)$",
		"    RunTests -> Done when ctx.tool_marker = go or ctx.tool_marker = stop\n"))
	if len(diags) != 0 {
		t.Fatalf("or-routing must not warn (hasComplexRoute), got %v", diags)
	}
}

func TestDIP152_NotEqualCatchAllClean(t *testing.T) {
	diags := markerDiagsFor(t, mk("^(go|stop)$",
		"    RunTests -> Done on go\n    RunTests -> Done when ctx.tool_marker != go\n"))
	if len(diags) != 0 {
		t.Fatalf("!= catch-all must not warn, got %v", diags)
	}
}

func TestDIP152_NonEnumerableClean(t *testing.T) {
	diags := markerDiagsFor(t, mk(`.*fail.*`, "    RunTests -> Done on x\n"))
	if len(diags) != 0 {
		t.Fatalf("non-enumerable regex must not warn, got %v", diags)
	}
}

func TestDIP152_EmptyBranchClean(t *testing.T) {
	diags := markerDiagsFor(t, mk("^(go|stop|)$", "    RunTests -> Done on go\n"))
	if len(diags) != 0 {
		t.Fatalf("empty-branch regex is non-enumerable, must not warn, got %v", diags)
	}
}
```

Add the imports `"fmt"`, `"strings"`, and `"github.com/2389-research/dippin-lang/parser"` to the test file's import block.

- [ ] **Step 2: Run — expect FAIL** (`Lint` doesn't emit DIP152 yet; `undefined` for the new funcs).

Run: `go test ./validator/ -run TestDIP152`
Expected: FAIL.

- [ ] **Step 3: Implement the coverage logic.** Append to `validator/lint_marker_coverage.go`:

```go
// lintMarkerCoverage checks DIP152 across all tool nodes.
func lintMarkerCoverage(w *ir.Workflow) []Diagnostic {
	var diags []Diagnostic
	elseValid := hasValidElseDefault(w)
	for _, n := range w.Nodes {
		if d, ok := checkMarkerCoverage(w, n, elseValid); ok {
			diags = append(diags, d)
		}
	}
	return diags
}

// checkMarkerCoverage returns a DIP152 diagnostic for one node if its enumerable
// marker_grep has markers neither routed nor covered.
func checkMarkerCoverage(w *ir.Workflow, n *ir.Node, elseValid bool) (Diagnostic, bool) {
	markers, ok := nodeEnumerableMarkers(n)
	if !ok {
		return Diagnostic{}, false
	}
	routed, hasUncond, hasComplex := classifyMarkerEdges(w.EdgesFrom(n.ID))
	if markerNodeCovered(elseValid, hasUncond, hasComplex) {
		return Diagnostic{}, false
	}
	missing := uncoveredMarkers(markers, routed)
	if len(missing) == 0 {
		return Diagnostic{}, false
	}
	return markerCoverageDiag(n, missing), true
}

// nodeEnumerableMarkers returns the enumerable marker set for a tool node, or
// (nil,false) if the node is not a marker tool or its grep is non-enumerable.
func nodeEnumerableMarkers(n *ir.Node) ([]string, bool) {
	cfg, ok := n.Config.(ir.ToolConfig)
	if !ok || cfg.MarkerGrep == "" {
		return nil, false
	}
	return enumerateMarkers(cfg.MarkerGrep)
}

// classifyMarkerEdges splits a node's outgoing edges into the simple-equality
// routed marker set plus flags for an unconditional edge and any "complex" edge
// (compound/negated/other-variable) whose mere presence makes the node safe.
func classifyMarkerEdges(edges []*ir.Edge) (routed map[string]struct{}, hasUncond, hasComplex bool) {
	routed = make(map[string]struct{})
	for _, e := range edges {
		if e.Condition == nil {
			hasUncond = true
			continue
		}
		if cmp, ok := ir.ExtractEqualityCondition(e); ok && cmp.Variable == "ctx.tool_marker" {
			routed[cmp.Value] = struct{}{}
			continue
		}
		hasComplex = true
	}
	return routed, hasUncond, hasComplex
}

// markerNodeCovered reports whether the node is safe regardless of the routed set.
func markerNodeCovered(elseValid, hasUncond, hasComplex bool) bool {
	return elseValid || hasUncond || hasComplex
}

// uncoveredMarkers returns the sorted markers not in the routed set.
func uncoveredMarkers(markers []string, routed map[string]struct{}) []string {
	var missing []string
	for _, m := range markers {
		if _, ok := routed[m]; !ok {
			missing = append(missing, m)
		}
	}
	sort.Strings(missing)
	return missing
}

// markerCoverageDiag builds the DIP152 diagnostic for a node.
func markerCoverageDiag(n *ir.Node, missing []string) Diagnostic {
	return Diagnostic{
		Code:     DIP152,
		Severity: SeverityWarning,
		Message:  fmt.Sprintf("tool node %q emits markers that no edge routes and no else default covers: %s", n.ID, strings.Join(missing, ", ")),
		Location: n.Source,
		Help:     "route each marker with an edge (e.g. `on <marker>`), add an unconditional fallback edge, or add a section `else -> <node>` default",
	}
}
```

- [ ] **Step 4: Register the pass.** In `validator/lint.go`, after the last `diags = append(diags, lint...(w)...)` line in `LintWithOptions`, add:

```go
	diags = append(diags, lintMarkerCoverage(w)...)
```

Also update the `Lint` doc-comment code range near line 8 to include DIP152.

- [ ] **Step 5: Run — expect PASS.**

Run: `go test ./validator/ -run TestDIP152 -v`
Expected: all PASS.

- [ ] **Step 6: Full validator suite + complexity (no regression on DIP101/102).**

Run: `go test ./validator/ -count=1 && just complexity`
Expected: `ok` + `Complexity OK.`

- [ ] **Step 7: Commit.**

```bash
git add validator/lint_marker_coverage.go validator/lint_marker_coverage_test.go validator/lint.go
git commit -m "feat(validator): DIP152 coverage lint + register pass"
```

---

### Task 4: CI guard — examples stay covered

**Files:**
- Modify: `validator/lint_examples_test.go` (zero-assertion ~line 50)

- [ ] **Step 1: Confirm examples are clean today.**

Run: `just lint-examples 2>&1 | grep -i DIP152 || echo "no DIP152 in examples"`
Expected: `no DIP152 in examples` (error_funnel covered by `else`; marker_routing fully routed).

- [ ] **Step 2: Extend the zero-assertion.** In `validator/lint_examples_test.go`, change the condition:

```go
			if d.Code == validator.DIP108 || d.Code == validator.DIP147 {
```
to:
```go
			if d.Code == validator.DIP108 || d.Code == validator.DIP147 || d.Code == validator.DIP152 {
```
Also update the test's doc comment to mention DIP152.

- [ ] **Step 3: Run — expect PASS** (no example emits DIP152).

Run: `go test ./validator/ -run TestLintExamples -v`
Expected: PASS.

- [ ] **Step 4: Commit.**

```bash
git add validator/lint_examples_test.go
git commit -m "test(validator): TestLintExamples zero-asserts DIP152 (guard example coverage)"
```

---

### Task 5: Docs & count sweep + regenerate embedded spec

**Files:** count/range hardcodes across docs + site + the generated spec.

- [ ] **Step 1: Find every count/range hardcode.**

Run:
```bash
git grep -nE '61 diagnostic|61 codes|51 semantic|DIP101[-–]DIP151|60 documented' -- \
  CLAUDE.md README.md AGENTS.md docs/ site/ | grep -v CHANGELOG
```
This lists the occurrences to bump: 61→62 codes, 51→52 semantic, `DIP101-DIP151`→`DIP101-DIP152` (watch BOTH ASCII `-` and en-dash `–`).

- [ ] **Step 2: Update each non-generated file** from Step 1: `CLAUDE.md`, `README.md`, `AGENTS.md`, `docs/architecture.md`, `docs/integration.md`, `docs/llm-reference.md`, `docs/validation.md`, `site/content/{validation,glossary,architecture,editors,cli}.md`, `site/content/blog/editor-setup.md`, `site/static/skill.md`, and the `validator/lint.go` + `validator/lint_codes.go` range comments. Do NOT edit `CHANGELOG.md` / `site/content/changelog.md` / `cmd/dippin/generated-spec.md` (generated).

- [ ] **Step 3: Add the DIP152 doc entry.** In `docs/validation.md` and `site/content/validation.md`, add a `### DIP152 — marker_grep enumerates an unrouted marker` section modeled on the neighbouring DIP102 section, with the trigger, the `example.dip` snippet, and the fix (route/fallback/else). In `site/static/skill.md`, add the DIP152 row to the diagnostics table.

- [ ] **Step 4: Regenerate the embedded spec.**

Run: `bash scripts/gen-spec.sh`
This rewrites `cmd/dippin/generated-spec.md` from `docs/llm-reference.md` + `site/static/skill.md`.

- [ ] **Step 5: Verify the spec-currency gate.**

Run: `go test ./releasecheck/ -count=1`
Expected: PASS (the embedded spec matches the generator).

- [ ] **Step 6: Commit.**

```bash
git add -A
git commit -m "docs: document DIP152 + bump diagnostic counts (61->62, DIP101-DIP152)"
```

---

### Task 6: Final verification

- [ ] **Step 1: Full race suite + gates.**

Run:
```bash
go test ./... -count=1 -race && just vet && just complexity && golangci-lint run && just validate-examples && just lint-examples 2>&1 | grep -c DIP152
```
Expected: all green; the final `grep -c DIP152` prints `0` (no example warns).

- [ ] **Step 2: Manual smoke test of the warning + explain.**

Run:
```bash
printf 'workflow W\n  goal: "t"\n  start: T\n  exit: D\n  tool T\n    command: run\n    marker_grep: ^(go|stop)$\n  agent D\n    prompt: d\n  edges\n    T -> D on go\n' > /tmp/dip152.dip
go run ./cmd/dippin lint /tmp/dip152.dip
go run ./cmd/dippin explain DIP152
```
Expected: `lint` reports DIP152 naming `stop`; `explain` prints the explanation.

- [ ] **Step 3: Push and open the PR** (only when the user asks — do not auto-merge).

```bash
git push -u origin feat/156-marker-coverage
gh pr create --base main --title "feat(validator): coverage-aware marker_grep lint (DIP152) (#156)" --body "Implements #156 per docs/superpowers/specs/2026-07-07-issue-156-marker-coverage-design.md."
```
