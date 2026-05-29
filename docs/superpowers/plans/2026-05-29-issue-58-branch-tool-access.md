# BranchConfig.ToolAccess per-branch override (#58) — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add `tool_access` as a per-branch override on block-form `parallel` nodes — carried faithfully through parse → format → DOT round-trip and validated via DIP139 — mirroring how `Model`/`Provider`/`Fidelity` already work.

**Architecture:** Add `ToolAccess` to `ir.BranchConfig` and plug it into the table-driven extension points #76 left in place (one line each in parser, formatter, DOT export, DOT migrate). Generalize the DIP139 lint to scan branches, mirroring DIP114's `checkBranchFidelities`. dippin carries + lints; tracker enforces the override (inherit-on-empty) at runtime.

**Tech Stack:** Go. Packages `ir`, `parser`, `formatter`, `export`, `migrate`, `validator`. Spec: `docs/superpowers/specs/2026-05-29-issue-58-branch-tool-access-design.md`.

**Conventions (CLAUDE.md):**
- All ops via `just`. NEVER raw `go build`/`go test`/`gocyclo`. Use `just test-pkg <pkg>`, `just complexity`, `just check`.
- Complexity caps (pre-commit + CI): cyclomatic ≤ 5, cognitive ≤ 7 per non-test function. No `//nolint` — extract helpers.
- TDD: failing test first, watch it fail, then implement.
- Commit trailer: `Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>`
- Leave `.claude/settings.local.json` uncommitted; stage only the files named in each commit.
- Already on branch `feat/58-branch-tool-access`.

---

## File map

| File | Change |
|---|---|
| `ir/ir.go` | add `ToolAccess string` to `BranchConfig` + rationale comment |
| `parser/parse_nodes.go` | `applyBranchField`: add `case "tool_access"` |
| `parser/parser_test.go` | parser test |
| `formatter/format.go` | `writeBranch` guard + `writeBranchFields` emit |
| `formatter/format_test.go` | formatter test |
| `export/dot.go` | `encodeBranch`: one `appendBranchField` line |
| `migrate/migrate.go` | `branchFieldSetters`: one map entry |
| `migrate/migrate_test.go` | migrate decode test |
| `validator/lint_tool_access.go` | generalize DIP139 to branches (helper decomposition) + DIP140 tripwire comment |
| `validator/explanations.go` | broaden DIP139 wording to include branches |
| `validator/lint_test.go` | DIP139 branch tests |
| `migrate/roundtrip_test.go` | extend acceptance round-trip |
| `docs/nodes.md`, `skill.md` | docs |

---

## Task 1: IR field + parser

**Files:**
- Modify: `ir/ir.go` (`BranchConfig`), `parser/parse_nodes.go` (`applyBranchField`)
- Test: `parser/parser_test.go`

- [ ] **Step 1: Write the failing test.** Add to `parser/parser_test.go`:

```go
func TestParseBranchToolAccess(t *testing.T) {
	input := `workflow Test
  start: split
  exit: join

  agent worker_a
    prompt: "A"

  agent worker_b
    prompt: "B"

  parallel split
    branch: worker_a
      tool_access: none
    branch: worker_b
      model: claude-haiku-4-5

  fan_in join <- worker_a, worker_b

  edges
    join -> join
`
	p := NewParser(input, "test.dip")
	w, err := p.Parse()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var pNode *ir.Node
	for _, n := range w.Nodes {
		if n.Kind == ir.NodeParallel {
			pNode = n
			break
		}
	}
	if pNode == nil {
		t.Fatal("parallel node not found")
	}
	cfg := pNode.Config.(ir.ParallelConfig)
	if len(cfg.Branches) != 2 {
		t.Fatalf("branches = %d, want 2", len(cfg.Branches))
	}
	if cfg.Branches[0].ToolAccess != "none" {
		t.Errorf("branch[0].ToolAccess = %q, want none", cfg.Branches[0].ToolAccess)
	}
	if cfg.Branches[1].ToolAccess != "" {
		t.Errorf("branch[1].ToolAccess = %q, want empty", cfg.Branches[1].ToolAccess)
	}
}
```

- [ ] **Step 2: Run to verify it fails.** `just test-pkg parser`
Expected: FAIL — `branch[0].ToolAccess = "" want none` (field doesn't exist / not parsed). It may not compile until the IR field is added — that's an expected failing state; proceed to Step 3.

- [ ] **Step 3: Add the IR field.** In `ir/ir.go`, change `BranchConfig`:

```go
// BranchConfig holds per-branch configuration for block-form parallel nodes.
type BranchConfig struct {
	Target   string
	Model    string
	Provider string
	Fidelity string
	// ToolAccess is a per-branch override of the target agent's tool_access.
	// Recognized values mirror AgentConfig.ToolAccess: "" (inherit) and "none"
	// (strip tools); other values lint as DIP139 and fail closed at runtime.
	// Empty INHERITS the target agent's tool_access (never resets to the full
	// catalog) — tracker resolves effective = branch if non-empty else agent.
	// dippin carries + lints this field; tracker enforces the override, exactly
	// as it does for Model/Provider/Fidelity.
	ToolAccess string
}
```

- [ ] **Step 4: Add the parser case.** In `parser/parse_nodes.go`, `applyBranchField`:

```go
// applyBranchField sets a field on a BranchConfig.
func applyBranchField(bc *ir.BranchConfig, key, val string) {
	switch key {
	case "model":
		bc.Model = val
	case "provider":
		bc.Provider = val
	case "fidelity":
		bc.Fidelity = val
	case "tool_access":
		bc.ToolAccess = val
	}
}
```

- [ ] **Step 5: Run to verify it passes.** `just test-pkg parser`
Expected: PASS.

- [ ] **Step 6: Commit.**
```bash
git add ir/ir.go parser/parse_nodes.go parser/parser_test.go
git commit -m "feat: parse per-branch tool_access into BranchConfig (#58)

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

## Task 2: Formatter (guard + emit)

**Files:**
- Modify: `formatter/format.go` (`writeBranch`, `writeBranchFields`)
- Test: `formatter/format_test.go`

- [ ] **Step 1: Write the failing test.** Add to `formatter/format_test.go`:

```go
// A branch that sets ONLY tool_access must still be emitted (regression guard
// for the writeBranch early-return that previously checked only model/provider/fidelity).
func TestFormatBranchToolAccessOnly(t *testing.T) {
	w := &ir.Workflow{
		Name:  "T",
		Start: "split",
		Exit:  "join",
		Nodes: []*ir.Node{
			{ID: "a", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "a"}},
			{ID: "split", Kind: ir.NodeParallel, Config: ir.ParallelConfig{
				Targets: []string{"a"},
				Branches: []ir.BranchConfig{{Target: "a", ToolAccess: "none"}},
			}},
			{ID: "join", Kind: ir.NodeFanIn, Config: ir.FanInConfig{Sources: []string{"a"}}},
		},
	}
	out := Format(w)
	if !strings.Contains(out, "tool_access: none") {
		t.Errorf("formatted output missing per-branch tool_access; got:\n%s", out)
	}
}
```

> Note: confirm `Format` is the package entrypoint and `strings` is imported in `format_test.go` (it is used widely there). If the helper signature differs, match the existing format tests in the file.

- [ ] **Step 2: Run to verify it fails.** `just test-pkg formatter`
Expected: FAIL — the branch is skipped by the early-return guard, so `tool_access: none` is absent.

- [ ] **Step 3: Update the guard and the writer.** In `formatter/format.go`:

```go
// writeBranch writes a single branch entry in block-form parallel.
func writeBranch(wr *writer, b ir.BranchConfig) {
	wr.line("branch: %s", b.Target)
	if b.Model == "" && b.Provider == "" && b.Fidelity == "" && b.ToolAccess == "" {
		return
	}
	wr.push()
	writeBranchFields(wr, b)
	wr.pop()
}

// writeBranchFields writes the optional fields within a branch.
func writeBranchFields(wr *writer, b ir.BranchConfig) {
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

> `writeBranchFields` now has 4 `if`s — cyclo 5, at the cap. That is within limits (≤5). Do not split unless `just complexity` reports a violation.

- [ ] **Step 4: Run to verify it passes.** `just test-pkg formatter`
Expected: PASS.

- [ ] **Step 5: Check complexity.** `just complexity`
Expected: no output.

- [ ] **Step 6: Commit.**
```bash
git add formatter/format.go formatter/format_test.go
git commit -m "feat: emit per-branch tool_access in formatter (#58)

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

## Task 3: DOT export + migrate (round-trip halves)

**Files:**
- Modify: `export/dot.go` (`encodeBranch`), `migrate/migrate.go` (`branchFieldSetters`)
- Test: `migrate/migrate_test.go`

- [ ] **Step 1: Write the failing test.** Add to `migrate/migrate_test.go`:

```go
func TestMigrateBranchToolAccess(t *testing.T) {
	dot := `digraph G {
		Start [shape=Mdiamond];
		P [shape=component, targets="a,b", branches="target=a;tool_access=none,target=b;model=claude-haiku-4-5"];
		a [shape=box];
		b [shape=box];
		Exit [shape=Msquare];
		Start -> P;
		P -> a;
		P -> b;
		a -> Exit;
		b -> Exit;
	}`
	w, err := Migrate(dot)
	if err != nil {
		t.Fatalf("migrate: %v", err)
	}
	cfg, ok := w.Node("P").Config.(ir.ParallelConfig)
	if !ok {
		t.Fatalf("config type = %T, want ParallelConfig", w.Node("P").Config)
	}
	if len(cfg.Branches) != 2 {
		t.Fatalf("branches = %d, want 2", len(cfg.Branches))
	}
	if cfg.Branches[0].ToolAccess != "none" {
		t.Errorf("branch[0].ToolAccess = %q, want none", cfg.Branches[0].ToolAccess)
	}
	if cfg.Branches[1].ToolAccess != "" {
		t.Errorf("branch[1].ToolAccess = %q, want empty", cfg.Branches[1].ToolAccess)
	}
}
```

- [ ] **Step 2: Run to verify it fails.** `just test-pkg migrate`
Expected: FAIL — `branchFieldSetters` has no `tool_access` key, so `ToolAccess` stays empty.

- [ ] **Step 3a: Add the export token.** In `export/dot.go`, `encodeBranch`:

```go
// encodeBranch encodes one branch as ';'-joined k=v tokens. target is always
// first; model/provider/fidelity/tool_access only when non-empty.
func encodeBranch(b ir.BranchConfig) string {
	parts := []string{"target=" + encodeBranchToken(b.Target)}
	parts = appendBranchField(parts, "model", b.Model)
	parts = appendBranchField(parts, "provider", b.Provider)
	parts = appendBranchField(parts, "fidelity", b.Fidelity)
	parts = appendBranchField(parts, "tool_access", b.ToolAccess)
	return strings.Join(parts, ";")
}
```

- [ ] **Step 3b: Add the migrate setter.** In `migrate/migrate.go`, `branchFieldSetters`:

```go
var branchFieldSetters = map[string]func(*ir.BranchConfig, string){
	"target":      func(b *ir.BranchConfig, v string) { b.Target = v },
	"model":       func(b *ir.BranchConfig, v string) { b.Model = v },
	"provider":    func(b *ir.BranchConfig, v string) { b.Provider = v },
	"fidelity":    func(b *ir.BranchConfig, v string) { b.Fidelity = v },
	"tool_access": func(b *ir.BranchConfig, v string) { b.ToolAccess = v },
}
```

- [ ] **Step 4: Run to verify it passes.** `just test-pkg migrate`
Expected: PASS.

- [ ] **Step 5: Commit.**
```bash
git add export/dot.go migrate/migrate.go migrate/migrate_test.go
git commit -m "feat: round-trip per-branch tool_access through DOT (#58)

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

## Task 4: DIP139 validation for branches

**Files:**
- Modify: `validator/lint_tool_access.go` (generalize `lintToolAccessValues` + DIP140 tripwire comment), `validator/explanations.go` (DIP139 wording)
- Test: `validator/lint_test.go`

- [ ] **Step 1: Write the failing tests.** Add to `validator/lint_test.go`:

```go
func TestDIP139_BranchInvalidToolAccess(t *testing.T) {
	w := &ir.Workflow{
		Nodes: []*ir.Node{
			{ID: "split", Kind: ir.NodeParallel, Config: ir.ParallelConfig{
				Targets: []string{"a", "b"},
				Branches: []ir.BranchConfig{
					{Target: "a", ToolAccess: "none"},  // valid
					{Target: "b", ToolAccess: "nono"},  // invalid -> DIP139
				},
			}},
		},
	}
	diags := lintToolAccessValues(w)
	if len(diags) != 1 {
		t.Fatalf("got %d diagnostics, want 1; %+v", len(diags), diags)
	}
	if diags[0].Code != DIP139 {
		t.Errorf("code = %s, want DIP139", diags[0].Code)
	}
	if !strings.Contains(diags[0].Message, `branch "b"`) {
		t.Errorf("message %q should name branch \"b\"", diags[0].Message)
	}
}

func TestDIP139_BranchValidToolAccessSilent(t *testing.T) {
	w := &ir.Workflow{
		Nodes: []*ir.Node{
			{ID: "split", Kind: ir.NodeParallel, Config: ir.ParallelConfig{
				Targets: []string{"a", "b"},
				Branches: []ir.BranchConfig{
					{Target: "a", ToolAccess: "none"},
					{Target: "b"}, // empty -> inherit, silent
				},
			}},
		},
	}
	if diags := lintToolAccessValues(w); len(diags) != 0 {
		t.Errorf("got %d diagnostics, want 0; %+v", len(diags), diags)
	}
}
```

> Confirm `strings` is imported in `lint_test.go`; if not, add it. If the test file is `package validator` (white-box) it can call `lintToolAccessValues` directly — verify by checking existing tests in the file (they reference unexported lint funcs, so it is white-box).

- [ ] **Step 2: Run to verify it fails.** `just test-pkg validator`
Expected: FAIL — `lintToolAccessValues` only scans `AgentConfig`, so it returns 0 diagnostics for the branch case.

- [ ] **Step 3: Generalize the DIP139 lint.** In `validator/lint_tool_access.go`, replace `lintToolAccessValues` with the decomposition below (mirrors `lint_style.go`'s DIP114 `checkNodeFidelityByKind`/`checkBranchFidelities`/`checkFidelityValue`):

```go
// lintToolAccessValues fires DIP139 when an agent node — or a per-branch override
// on a parallel node — sets tool_access to a value other than "" or "none"
// (case-insensitive). Authors who skip lint and ship an invalid value get
// fail-closed tracker behavior; DIP139 surfaces the typo at lint time.
func lintToolAccessValues(w *ir.Workflow) []Diagnostic {
	var diags []Diagnostic
	for _, n := range w.Nodes {
		diags = append(diags, checkNodeToolAccessByKind(n)...)
	}
	return diags
}

// checkNodeToolAccessByKind checks tool_access for a single node by config type.
func checkNodeToolAccessByKind(n *ir.Node) []Diagnostic {
	switch cfg := n.Config.(type) {
	case ir.AgentConfig:
		return checkToolAccessValue(n, cfg.ToolAccess, "")
	case ir.ParallelConfig:
		return checkBranchToolAccess(n, cfg.Branches)
	default:
		return nil
	}
}

// checkBranchToolAccess checks tool_access on each branch of a parallel node.
func checkBranchToolAccess(n *ir.Node, branches []ir.BranchConfig) []Diagnostic {
	var diags []Diagnostic
	for _, b := range branches {
		diags = append(diags, checkToolAccessValue(n, b.ToolAccess, b.Target)...)
	}
	return diags
}

// checkToolAccessValue validates a single tool_access string, returning a
// DIP139 diagnostic if unrecognized. A non-empty branch arg produces a
// branch-qualified message.
func checkToolAccessValue(n *ir.Node, toolAccess, branch string) []Diagnostic {
	canonical := strings.ToLower(strings.TrimSpace(toolAccess))
	if validToolAccess[canonical] {
		return nil
	}
	msg := fmt.Sprintf("node %q has tool_access %q which is not recognized", n.ID, toolAccess)
	if branch != "" {
		msg = fmt.Sprintf("node %q branch %q has tool_access %q which is not recognized", n.ID, branch, toolAccess)
	}
	return []Diagnostic{{
		Code:     DIP139,
		Severity: SeverityWarning,
		Message:  msg,
		Location: n.Source,
		Help:     "valid value: none (omit the field for the full catalog). Invalid values fall back to no-tools at runtime — fix the typo or remove the field.",
	}}
}
```

- [ ] **Step 4: Add the DIP140 tripwire comment.** In `validator/lint_tool_access.go`, immediately above `func lintParamsReenablesTools`, add:

```go
// NOTE: DIP140 is agent-only because BranchConfig has no Params field today.
// If BranchConfig.Params is ever added, extend this scan to branches — the
// tool-re-enabling bypass would reopen at the branch level.
```

- [ ] **Step 5: Broaden the DIP139 explanation wording.** In `validator/explanations.go`, the `DIP139` entry: update `Summary`/`Trigger` so they say "an agent node or a parallel branch" rather than "agent node" only. Keep `Fix`/`Example` as-is (the agent example still illustrates the rule). Minimal edit — just acknowledge branches in the trigger text.

- [ ] **Step 6: Run to verify it passes + complexity.** `just test-pkg validator` then `just complexity`
Expected: PASS; no complexity output. (`checkNodeToolAccessByKind` cyclo 3, `checkToolAccessValue` cyclo 3 — both under cap.)

- [ ] **Step 7: Commit.**
```bash
git add validator/lint_tool_access.go validator/explanations.go validator/lint_test.go
git commit -m "feat: DIP139 validates per-branch tool_access (#58)

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

## Task 5: Acceptance round-trip + parity coverage

**Files:**
- Test: `migrate/roundtrip_test.go`

- [ ] **Step 1: Extend the acceptance round-trip.** In `migrate/roundtrip_test.go`, in `TestRoundtripBlockFormParallel`, add `tool_access: none` to the `fast` branch in the `src` workflow and to the expected `want` slice.

Edit the `parallel split` block in `src` so the `fast` branch reads:

```
  parallel split
    branch: fast
      model: claude-haiku-4-5
      provider: anthropic
      fidelity: summary
      tool_access: none
    branch: accurate
      model: claude-opus-4-7
      provider: anthropic
      fidelity: full
```

And update the `want` slice's first element to include `ToolAccess: "none"`:

```go
	assertBranchesEqual(t, cfg.Branches, []ir.BranchConfig{
		{Target: "fast", Model: "claude-haiku-4-5", Provider: "anthropic", Fidelity: "summary", ToolAccess: "none"},
		{Target: "accurate", Model: "claude-opus-4-7", Provider: "anthropic", Fidelity: "full"},
	})
```

- [ ] **Step 2: Add a parity coverage test.** Add to `migrate/roundtrip_test.go` (or `migrate/coverage_test.go` if that is where parity tests live — check `TestCompareParallelConfigs_DifferentBranches` and colocate):

```go
func TestCompareParallelConfigs_DifferentBranchToolAccess(t *testing.T) {
	a := ir.ParallelConfig{
		Targets:  []string{"A"},
		Branches: []ir.BranchConfig{{Target: "A", ToolAccess: "none"}},
	}
	b := ir.ParallelConfig{
		Targets:  []string{"A"},
		Branches: []ir.BranchConfig{{Target: "A"}},
	}
	diffs := compareParallelConfigs("P", "nodes.P", a, b)
	if len(diffs) == 0 {
		t.Fatal("expected a branches difference for differing tool_access, got none")
	}
}
```

> This guards that `--check` parity detects a per-branch `tool_access` change. The production code already covers it (`compareParallelBranches` uses `!=`), so this test should PASS immediately — it is a regression guard, not a red test. If it is colocated in `migrate/coverage_test.go`, place it next to `TestCompareParallelConfigs_DifferentBranches`.

- [ ] **Step 3: Run.** `just test-pkg migrate`
Expected: PASS (both tests).

- [ ] **Step 4: Commit.**
```bash
git add migrate/roundtrip_test.go migrate/coverage_test.go
git commit -m "test: acceptance round-trip + parity coverage for branch tool_access (#58)

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```
(If you did not touch `coverage_test.go`, drop it from the `git add`.)

---

## Task 6: Docs

**Files:**
- Modify: `docs/nodes.md`, `skill.md` (find it: `git ls-files | grep skill.md`)

- [ ] **Step 1: Update `docs/nodes.md`.** In the block-form parallel section (under `## Parallel Nodes`), add `tool_access` to the per-branch field list and to the `branches=` DOT token list. Add one sentence on the inherit rule. Concretely:

- In the "Each `branch:` entry … per-branch overrides for `model`, `provider`, and `fidelity`." sentence, change to: "… per-branch overrides for `model`, `provider`, `fidelity`, and `tool_access`."
- After that sentence add: "A branch's `tool_access` follows the same rules as an agent's (`none` to strip tools, omit to inherit). An omitted branch `tool_access` inherits the target agent's setting — it never re-grants the full catalog."
- In the `#### DOT mapping` paragraph, change "(`target` plus any of `model`/`provider`/`fidelity`)" to "(`target` plus any of `model`/`provider`/`fidelity`/`tool_access`)".

- [ ] **Step 2: Update `skill.md`.** Locate the file (`git ls-files | grep skill.md`) and its `tool_access` section and block-form branch example. Add ONE sentence near the `tool_access` documentation: "`tool_access` may also be set per-branch on a block-form `parallel` node; an omitted branch value inherits the target agent's setting." Do not duplicate the threat-model prose.

- [ ] **Step 3: Verify docs build / examples still pass.** `just validate-examples`
Expected: all examples validate (docs-only change). `docs/generated-spec.md` may be auto-regenerated by the pre-commit hook — expected; include it if the hook stages it.

- [ ] **Step 4: Commit.**
```bash
git add docs/nodes.md skill.md
git commit -m "docs: document per-branch tool_access (#58)

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```
(Adjust the `skill.md` path to the actual location from Step 2.)

---

## Task 7: Full gate + PR

- [ ] **Step 1: Full gate.** `just check`
Expected: all checks pass (build, vet, fmt, test-race, complexity, validate-examples, tree-sitter).

- [ ] **Step 2: Open the PR.**
```bash
git push -u origin feat/58-branch-tool-access
gh pr create --title "feat: per-branch tool_access override for block-form parallel (#58)" --body "$(cat <<'EOF'
## Summary
Adds `tool_access` as a per-branch override on block-form `parallel` nodes, carried through parse → format → DOT round-trip and validated via DIP139. Fixes #58 (deferred from #41, unblocked by #76).

- `ir.BranchConfig` gains `ToolAccess`, plugged into the table-driven extension points #76 added (one line each: parser, formatter, DOT export, DOT migrate).
- DIP139 now validates per-branch `tool_access` (mirrors DIP114's branch handling); invalid values fail closed at runtime, branch-qualified lint message.
- Parity (`--check`) auto-covers the new field via struct `!=`; covered by a regression test.

## Safety semantics (for the tracker side)
Empty branch `tool_access` **inherits** the target agent's value — it never resets to the full catalog. `effective = branch if non-empty else agent`. dippin carries + lints; tracker enforces. See the spec for the normative rule and the rationale for why no extra lint is needed in v1 (the value set can't express a widening).

## Testing
- Parser, formatter (tool_access-only branch), DOT round-trip, DIP139 (branch-qualified, pos+neg), acceptance round-trip, parity coverage.
- `just check` green.

Note: CHANGELOG untouched (tag-time per CLAUDE.md). Tracker consumption of per-branch tool_access is out of scope — this carries the field through the toolchain, mirroring how Model/Provider/Fidelity are carried.

🤖 Generated with [Claude Code](https://claude.com/claude-code)
EOF
)"
```

---

## Self-review notes (author)

- **Spec coverage:** IR+parser (T1), formatter guard+emit (T2), DOT round-trip (T3), DIP139 generalization + DIP140 tripwire + wording (T4), acceptance + parity tests (T5), docs (T6). All spec touch-sites mapped. Parity needs no production change (confirmed — `branchesKey` was removed in #76). No example file (per spec out-of-scope).
- **Complexity:** `writeBranchFields` reaches cyclo 5 (4 ifs) — at the cap, OK. Validator decomposition keeps each function ≤ cyclo 3. `just complexity` run in T2, T4, and the T7 gate.
- **Type consistency:** field `ToolAccess`; DOT/parse/format key `tool_access` (matches the agent field key at `parse_nodes.go`). New validator helpers: `checkNodeToolAccessByKind`, `checkBranchToolAccess`, `checkToolAccessValue` (parallel to the DIP114 trio). Reuses the existing `validToolAccess` map — not duplicated.
