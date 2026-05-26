# Issue #41 — `tool_access:` Implementation Plan (v0.32.0)

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Ship `tool_access: none` on agent nodes (joint dippin v0.32.0 + tracker release) so authors can bound the v0.28.2 runaway-agent vector per-node.

**Architecture:** One `string` field on `ir.AgentConfig`. One lint code (DIP139). Verbatim through parser/formatter/DOT/migrate. Tracker reads it and short-circuits the tool registry when non-empty. No defaults cascade, no named type, no canonical accessor — the simplification is the design.

**Tech Stack:** Go, dippin's recursive-descent parser, validator's `lintX(w) []Diagnostic` pattern, tracker's `tracker/agent/session.go` enforcement layer.

**Spec:** `docs/superpowers/specs/2026-05-26-issue-41-design.md`

**Build/test commands:** Always use `just` recipes (see `CLAUDE.md`). Never raw `go test`.

---

## Task 0: File follow-up issues + spec backlinks

**Purpose:** Break the DIP28 anti-pattern where "file follow-up if needed" promises rotted. Issues are numbered first so the spec can reference them and skill.md can link them.

**Files:**
- Modify: `docs/superpowers/specs/2026-05-26-issue-41-design.md` (record issue numbers in § Non-goals)

- [ ] **Step 1: File 9 follow-up issues against `2389-research/dippin-lang`**

For each non-goal in the spec (§ Non-goals 1–9), file one GitHub issue with title + body. Use `gh issue create`. Title pattern: `follow-up: <non-goal title> (deferred from #41)`. Body template:

```
Filed per v0.32.0 spec § Follow-up issues #N.
See: https://github.com/2389-research/dippin-lang/blob/main/docs/superpowers/specs/2026-05-26-issue-41-design.md#non-goals-v1

<one-paragraph context from the spec's non-goal text>
```

Label each with `safety-follow-up`. Use `gh label create safety-follow-up --color FBCA04 --description "Deferred from v0.32.0 issue #41"` first if the label doesn't exist.

The nine issues, in spec order:
1. `defaults:`-block cascade for `tool_access`
2. Middle tier (`read_filesystem` / `read_only`) for tool_access
3. Companion list fields (`disallowed_tools`, `allowed_tools`)
4. Chain-attack mitigation (`${ctx.last_response}` auto-injection)
5. Cross-node lint for tool_access leak (DIP141 candidate)
6. `BranchConfig.ToolAccess` per-branch override (depends on per-branch DOT round-trip fix)
7. `ManagerLoopConfig.SubgraphRef` cross-workflow tool_access propagation
8. `Params` bypass lint (dippin-side companion to tracker runtime defense)
9. Tool nodes under tool_access — skill.md cross-reference work

Record the assigned issue numbers (e.g., `#56, #57, #58, ...`) for Step 2.

- [ ] **Step 2: Update spec § Non-goals to reference the filed issue numbers**

Edit `docs/superpowers/specs/2026-05-26-issue-41-design.md`. For each non-goal bullet, append ` ([#N](https://github.com/2389-research/dippin-lang/issues/N))` to the bullet text. Example:

```
1. **`defaults:`-block cascade.** Workflow-level ... Defer until incident data shows per-node annotation is insufficient. ([#56](https://github.com/2389-research/dippin-lang/issues/56))
```

- [ ] **Step 3: Commit**

```bash
git add docs/superpowers/specs/2026-05-26-issue-41-design.md
git commit -m "docs(spec): link v0.32.0 follow-up issues into spec non-goals"
```

---

## Task 1: Add `ToolAccess` field to `ir.AgentConfig`

**Files:**
- Modify: `ir/ir.go` (extend `AgentConfig` struct)

- [ ] **Step 1: Add the field**

Open `ir/ir.go`. Find `type AgentConfig struct` (around line 90). Add `ToolAccess string` immediately after `WorkingDir string` and before `Params map[string]string`:

```go
type AgentConfig struct {
	Prompt              string
	SystemPrompt        string
	Model               string // Per-node override
	Provider            string
	MaxTurns            int
	CmdTimeout          time.Duration
	CacheTools          bool
	Compaction          string
	CompactionThreshold float64
	ReasoningEffort     string
	Fidelity            string
	AutoStatus          bool              // Parse STATUS: from response
	GoalGate            bool              // Pipeline fails if this node fails
	ResponseFormat      string            // "json_object" or "json_schema"
	ResponseSchema      string            // JSON schema (when ResponseFormat is "json_schema")
	Backend             string            // Per-node backend override: "native", "claude-code", "acp"
	WorkingDir          string            // Per-node working directory override
	ToolAccess          string            // "" (default = full catalog) or "none" (no LLM tools)
	Params              map[string]string // Generic key-value pairs passed through to runtime
}
```

- [ ] **Step 2: Verify build**

Run: `just build`
Expected: success (no test changes; just a new field that nothing reads yet).

- [ ] **Step 3: Commit**

```bash
git add ir/ir.go
git commit -m "feat(ir): add AgentConfig.ToolAccess field"
```

---

## Task 2: Parser handler for `tool_access:`

**Files:**
- Test: `parser/parser_test.go` (extend with new test cases)
- Modify: `parser/parse_nodes.go` (one new case in `applyAgentRuntimeField`)

- [ ] **Step 1: Write the failing test**

Append to `parser/parser_test.go` (use the existing real-source-text pattern — search for an existing `TestParseAgent...` for convention):

```go
func TestParseAgent_ToolAccess(t *testing.T) {
	cases := []struct {
		name string
		src  string
		want string
	}{
		{
			name: "explicit none",
			src: `workflow X
  start: A
  exit: A

  agent A
    prompt: "x"
    tool_access: none
`,
			want: "none",
		},
		{
			name: "case variant",
			src: `workflow X
  start: A
  exit: A

  agent A
    prompt: "x"
    tool_access: None
`,
			want: "None",
		},
		{
			name: "invalid value stored verbatim",
			src: `workflow X
  start: A
  exit: A

  agent A
    prompt: "x"
    tool_access: foo
`,
			want: "foo",
		},
		{
			name: "quoted",
			src: `workflow X
  start: A
  exit: A

  agent A
    prompt: "x"
    tool_access: "none"
`,
			want: "none",
		},
		{
			name: "omitted (default)",
			src: `workflow X
  start: A
  exit: A

  agent A
    prompt: "x"
`,
			want: "",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p := parser.NewParser(tc.src, "test.dip")
			w, err := p.Parse()
			if err != nil {
				t.Fatalf("parse error: %v", err)
			}
			node := w.Node("A")
			if node == nil {
				t.Fatalf("node A not found")
			}
			cfg, ok := node.Config.(ir.AgentConfig)
			if !ok {
				t.Fatalf("expected AgentConfig, got %T", node.Config)
			}
			if cfg.ToolAccess != tc.want {
				t.Errorf("ToolAccess = %q, want %q", cfg.ToolAccess, tc.want)
			}
		})
	}
}
```

If the existing test file uses package-local helpers instead of fully-qualified `parser.NewParser`, mirror that style. Run `grep -n "func TestParseAgent" parser/parser_test.go` to find the convention and align imports.

- [ ] **Step 2: Run test to verify it fails**

Run: `just test-pkg parser`
Expected: FAIL — `TestParseAgent_ToolAccess` failures because all cases get `ToolAccess = ""` (parser doesn't know the key yet).

- [ ] **Step 3: Add the parser case**

Open `parser/parse_nodes.go`. Find `applyAgentRuntimeField` (around line 285). Add a `case "tool_access":` arm:

```go
// applyAgentRuntimeField handles runtime behavior fields.
func applyAgentRuntimeField(cfg *ir.AgentConfig, key, val string) bool {
	switch key {
	case "backend":
		cfg.Backend = val
	case "working_dir":
		cfg.WorkingDir = val
	case "tool_access":
		cfg.ToolAccess = val
	default:
		return false
	}
	return true
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `just test-pkg parser`
Expected: PASS (all five cases — including the quoted-string case, since `unquoteRaw` strips matched outer quotes before `val` reaches `applyAgentRuntimeField`).

- [ ] **Step 5: Commit**

```bash
git add parser/parse_nodes.go parser/parser_test.go
git commit -m "feat(parser): accept tool_access: field on agent nodes"
```

---

## Task 3: DIP139 validator lint

**Files:**
- Modify: `validator/lint_codes.go` (add DIP139 constant + CodeDescription entry)
- Create: `validator/lint_tool_access.go` (new file with `lintToolAccessValues`)
- Modify: `validator/lint.go` (register the lint in the `Lint(w)` pipeline)
- Modify: `validator/explanations.go` (add DIP139 explanation entry)
- Test: `validator/lint_tool_access_test.go` (new file)

- [ ] **Step 1: Write the failing test**

Create `validator/lint_tool_access_test.go`:

```go
package validator

import (
	"strings"
	"testing"

	"github.com/2389-research/dippin-lang/parser"
)

func TestLint_DIP139_InvalidToolAccess(t *testing.T) {
	src := `workflow X
  start: A
  exit: A

  agent A
    prompt: "x"
    tool_access: foo
`
	p := parser.NewParser(src, "test.dip")
	w, err := p.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	res := Lint(w)
	if !hasCode(res.Diagnostics, DIP139) {
		t.Errorf("expected DIP139, got: %v", codes(res.Diagnostics))
	}
}

func TestLint_DIP139_ValidValues(t *testing.T) {
	cases := []struct {
		name string
		val  string
	}{
		{"none lowercase", "none"},
		{"none uppercase", "NONE"},
		{"none mixed", "None"},
		{"none with surrounding whitespace", "  none  "},
		{"empty omitted", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var line string
			if tc.val != "" {
				// Wrap in quotes if it contains leading/trailing whitespace
				// so the parser preserves it (then strings.TrimSpace in the
				// validator does its job).
				if strings.TrimSpace(tc.val) != tc.val {
					line = "    tool_access: \"" + tc.val + "\"\n"
				} else {
					line = "    tool_access: " + tc.val + "\n"
				}
			}
			src := "workflow X\n  start: A\n  exit: A\n\n  agent A\n    prompt: \"x\"\n" + line
			p := parser.NewParser(src, "test.dip")
			w, err := p.Parse()
			if err != nil {
				t.Fatalf("parse error: %v", err)
			}
			res := Lint(w)
			if hasCode(res.Diagnostics, DIP139) {
				t.Errorf("DIP139 should not fire for %q; got: %v", tc.val, codes(res.Diagnostics))
			}
		})
	}
}

// hasCode and codes are test helpers — if they don't exist in another
// _test.go in this package, add them. Check first via:
//   grep -n "func hasCode\|func codes" validator/*_test.go
func hasCode(diags []Diagnostic, code string) bool {
	for _, d := range diags {
		if d.Code == code {
			return true
		}
	}
	return false
}

func codes(diags []Diagnostic) []string {
	out := make([]string, 0, len(diags))
	for _, d := range diags {
		out = append(out, d.Code)
	}
	return out
}
```

If `hasCode` / `codes` already exist in the validator test package (grep first), delete the local definitions to avoid `redeclared` build errors.

- [ ] **Step 2: Run test to verify it fails**

Run: `just test-pkg validator`
Expected: FAIL — DIP139 is undefined (compile error).

- [ ] **Step 3: Add the DIP139 constant + description**

Edit `validator/lint_codes.go`. After the `DIP138` line in the `const` block, add `DIP139`:

```go
	DIP137 = "DIP137" // unbounded manager_loop: no stop_condition and no max_cycles
	DIP138 = "DIP138" // tool node routes on stdout but declares no marker_grep / outputs (reserved)
	DIP139 = "DIP139" // invalid tool_access value on agent node
)
```

Also update the file's top-of-file range comment from `(DIP101–DIP138)` to `(DIP101–DIP139)`. Search `grep -n "DIP138" validator/lint_codes.go` to locate.

In the `func init()` block at the bottom of the file, add the `CodeDescription` entry:

```go
	CodeDescription[DIP138] = "tool node routes on stdout but declares no marker_grep / outputs"
	CodeDescription[DIP139] = "invalid tool_access value on agent node"
}
```

- [ ] **Step 4: Create `validator/lint_tool_access.go`**

```go
package validator

import (
	"fmt"
	"strings"

	"github.com/2389-research/dippin-lang/ir"
)

// validToolAccess lists recognized values for AgentConfig.ToolAccess after
// case-insensitive trim. Empty = inherit default (full catalog); "none" =
// disable LLM tools. v1 ships only these two values; middle tiers are
// tracked as a follow-up issue.
var validToolAccess = map[string]bool{
	"":     true,
	"none": true,
}

// lintToolAccessValues fires DIP139 when an agent's tool_access is set to
// a value other than "" or "none" (case-insensitive). Authors who skip
// lint and ship an invalid value get fail-closed tracker behavior — DIP139
// surfaces the typo at lint time.
func lintToolAccessValues(w *ir.Workflow) []Diagnostic {
	var diags []Diagnostic
	for _, n := range w.Nodes {
		cfg, ok := n.Config.(ir.AgentConfig)
		if !ok {
			continue
		}
		canonical := strings.ToLower(strings.TrimSpace(cfg.ToolAccess))
		if validToolAccess[canonical] {
			continue
		}
		diags = append(diags, Diagnostic{
			Code:     DIP139,
			Severity: SeverityWarning,
			Message:  fmt.Sprintf("node %q has tool_access %q which is not recognized", n.ID, cfg.ToolAccess),
			Location: n.Source,
			Help:     "valid value: none (omit the field for the full catalog). Invalid values fall back to no-tools at runtime — fix the typo or remove the field.",
		})
	}
	return diags
}
```

Severity: `SeverityWarning` to match DIP127/DIP130 (other enum-validation lints). The spec section "DIP139, error severity" was hyperbole — the consistent project convention is warning for enum mistakes; the runtime fail-closes regardless.

- [ ] **Step 5: Register the lint in `Lint(w)`**

Edit `validator/lint.go`. Find the chain of `diags = append(diags, lintX(w)...)` calls (around line 24–58). Add one line at the end, matching the existing pattern:

```go
	diags = append(diags, lintAgentParamsShadow(w)...)
	diags = append(diags, lintToolAccessValues(w)...)

	return Result{Diagnostics: diags}
}
```

- [ ] **Step 6: Add the DIP139 explanation**

Edit `validator/explanations.go`. After the `DIP138` block (around line 393–399), add:

```go
		DIP138: {
			Code:    DIP138,
			Summary: "tool node routes on stdout but declares no marker_grep / outputs",
			Trigger: "A tool node has outgoing conditional edges that test ctx.tool_stdout but the tool itself declares neither marker_grep nor outputs. The workflow is using untyped stdout-text routing where the typed marker_grep channel (ctx.tool_marker) is clearer. (Reserved: no firing logic in v0.29.0.)",
			Fix:     "Add marker_grep: \"<regex>\" to the tool node and switch edges to test ctx.tool_marker, or declare outputs: <values> so coverage analysis can see the routing set.",
			Example: "tool Check\n  command: echo done\n  marker_grep: \"^(done|more)$\"\nedges\n  Check -> Next when ctx.tool_marker = done",
		},
		DIP139: {
			Code:    DIP139,
			Summary: "invalid tool_access value on agent node",
			Trigger: "An agent node has tool_access set to a value other than 'none' (case-insensitive) or empty. The field is the v0.32.0 safety primitive that strips an LLM's tool catalog; v1 recognizes only one explicit value.",
			Fix:     "Use `tool_access: none` to disable LLM tools, or omit the field for the full catalog. Invalid values fall back to no-tools at runtime (fail-closed) — the diagnostic surfaces the typo so author intent matches runtime behavior.",
			Example: "agent ReportFinalStatus\n  prompt: \"Summarize\"\n  tool_access: nono   // DIP139: typo — runtime treats as 'none'",
		},
```

- [ ] **Step 7: Run test to verify it passes**

Run: `just test-pkg validator`
Expected: PASS (both `TestLint_DIP139_*` tests).

- [ ] **Step 8: Commit**

```bash
git add validator/lint_codes.go validator/lint_tool_access.go validator/lint.go validator/explanations.go validator/lint_tool_access_test.go
git commit -m "feat(validator): add DIP139 — invalid tool_access value"
```

---

## Task 4: Formatter emit

**Files:**
- Test: `formatter/format_test.go` (extend with a test case)
- Modify: `formatter/format.go` (extend `writeAgentRuntimeFields`)

- [ ] **Step 1: Write the failing test**

Open `formatter/format_test.go`. Find the existing agent-formatting test (search `grep -n "TestFormat.*Agent\|writeAgent" formatter/format_test.go`). Add a test:

```go
func TestFormat_AgentToolAccess(t *testing.T) {
	src := `workflow X
  start: A
  exit: A

  agent A
    prompt: "x"
    tool_access: none
`
	p := parser.NewParser(src, "test.dip")
	w, err := p.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	out, err := Format(w)
	if err != nil {
		t.Fatalf("format error: %v", err)
	}
	if !strings.Contains(out, "tool_access: none") {
		t.Errorf("formatted output missing tool_access line:\n%s", out)
	}
}

func TestFormat_AgentToolAccessOmitted(t *testing.T) {
	src := `workflow X
  start: A
  exit: A

  agent A
    prompt: "x"
`
	p := parser.NewParser(src, "test.dip")
	w, _ := p.Parse()
	out, _ := Format(w)
	if strings.Contains(out, "tool_access") {
		t.Errorf("omitted tool_access should not appear in formatted output:\n%s", out)
	}
}
```

If `Format` is exported differently in this package (check via `grep -n "^func Format" formatter/`), match the actual signature. Imports needed: `"strings"`, `"testing"`, `"github.com/2389-research/dippin-lang/parser"`.

- [ ] **Step 2: Run test to verify it fails**

Run: `just test-pkg formatter`
Expected: FAIL — formatted output omits the line (formatter doesn't know about the field yet).

- [ ] **Step 3: Add the emit**

Edit `formatter/format.go`. Find `writeAgentRuntimeFields` (around line 366):

```go
// writeAgentRuntimeFields writes runtime behavior fields for agent nodes.
func writeAgentRuntimeFields(wr *writer, cfg ir.AgentConfig) {
	if cfg.Backend != "" {
		wr.line("backend: %s", quoteValue(cfg.Backend))
	}
	if cfg.WorkingDir != "" {
		wr.line("working_dir: %s", quoteValue(cfg.WorkingDir))
	}
	if cfg.ToolAccess != "" {
		wr.line("tool_access: %s", quoteValue(cfg.ToolAccess))
	}
}
```

Verbatim emission — invalid values survive formatting → re-parsing still trips DIP139.

- [ ] **Step 4: Run test to verify it passes**

Run: `just test-pkg formatter`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add formatter/format.go formatter/format_test.go
git commit -m "feat(formatter): emit tool_access on agent nodes"
```

---

## Task 5: DOT export emit + reserved-attr registration

**Files:**
- Test: `export/dot_test.go` (extend)
- Modify: `export/dot.go` (extend `applyAgentRuntimeAttrs` + `reservedGraphAttrs`)

- [ ] **Step 1: Write the failing test**

Open `export/dot_test.go`. Add:

```go
func TestExportDOT_AgentToolAccess(t *testing.T) {
	src := `workflow X
  start: A
  exit: A

  agent A
    prompt: "x"
    tool_access: none
`
	p := parser.NewParser(src, "test.dip")
	w, err := p.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	out, err := ExportDOT(w, ExportOptions{})
	if err != nil {
		t.Fatalf("export error: %v", err)
	}
	if !strings.Contains(out, `tool_access="none"`) && !strings.Contains(out, `tool_access=none`) {
		t.Errorf("DOT output missing tool_access attribute:\n%s", out)
	}
}
```

If the convention here uses a different helper (search `grep -n "func TestExport" export/dot_test.go`), match style. Imports: `"strings"`, `"testing"`, parser.

- [ ] **Step 2: Run test to verify it fails**

Run: `just test-pkg export`
Expected: FAIL — DOT output omits the attr.

- [ ] **Step 3: Add the emit + reserve the attr**

Edit `export/dot.go`. Update `reservedGraphAttrs` (line 61) to include `tool_access`:

```go
var reservedGraphAttrs = map[string]bool{
	"goal": true, "rankdir": true, "model": true, "provider": true,
	"fidelity": true, "default_fidelity": true,
	"max_retries": true, "default_max_retry": true, "max_restarts": true,
	"max_total_tokens": true, "max_cost_cents": true, "max_wall_time": true,
	"tool_commands_allow": true, "tool_denylist_add": true,
	"tool_access": true,
}
```

Update `applyAgentRuntimeAttrs` (line 295) to emit `tool_access`:

```go
// applyAgentRuntimeAttrs adds backend, working_dir, and tool_access attributes.
func applyAgentRuntimeAttrs(attrs map[string]string, cfg ir.AgentConfig) {
	if cfg.Backend != "" {
		attrs["backend"] = cfg.Backend
	}
	if cfg.WorkingDir != "" {
		attrs["working_dir"] = cfg.WorkingDir
	}
	if cfg.ToolAccess != "" {
		attrs["tool_access"] = cfg.ToolAccess
	}
}
```

(Update the doc comment.)

- [ ] **Step 4: Run test to verify it passes**

Run: `just test-pkg export`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add export/dot.go export/dot_test.go
git commit -m "feat(export): emit tool_access attribute in DOT export"
```

---

## Task 6: Migrate extraction + parity comparator

**Files:**
- Test: `migrate/migrate_test.go` (extend with extraction test)
- Modify: `migrate/migrate.go` (one new attr extraction)
- Modify: `migrate/parity.go` (extend `compareAgentBehavior`)

- [ ] **Step 1: Write the failing tests**

Open `migrate/migrate_test.go`. Add:

```go
func TestMigrate_ToolAccessAttr(t *testing.T) {
	dot := `digraph X {
  rankdir=TB;
  A [shape=box, label=A, prompt="x", tool_access="none"];
}`
	w, err := Migrate(dot)
	if err != nil {
		t.Fatalf("migrate error: %v", err)
	}
	node := w.Node("A")
	if node == nil {
		t.Fatalf("node A not found")
	}
	cfg, ok := node.Config.(ir.AgentConfig)
	if !ok {
		t.Fatalf("expected AgentConfig, got %T", node.Config)
	}
	if cfg.ToolAccess != "none" {
		t.Errorf("ToolAccess = %q, want %q", cfg.ToolAccess, "none")
	}
}
```

If the DOT-input shape used by other `migrate_test.go` tests differs (check existing `func TestMigrate_*`), match the shape so the node is recognized as `NodeAgent`.

- [ ] **Step 2: Run test to verify it fails**

Run: `just test-pkg migrate`
Expected: FAIL — `cfg.ToolAccess` is empty (migrate doesn't extract the attr yet).

- [ ] **Step 3: Add the extraction**

Edit `migrate/migrate.go`. Find the runtime-attr extraction block (around line 399–404, the function that includes `cfg.Backend = v` and `cfg.WorkingDir = v`):

```go
	if v, ok := attrs["backend"]; ok {
		cfg.Backend = v
	}
	if v, ok := attrs["working_dir"]; ok {
		cfg.WorkingDir = v
	}
	if v, ok := attrs["tool_access"]; ok {
		cfg.ToolAccess = v
	}
}
```

- [ ] **Step 4: Extend the parity comparator**

Edit `migrate/parity.go`. Find `compareAgentBehavior` (around line 246). Add a `ToolAccess` comparison:

```go
// compareAgentBehavior compares goal_gate, auto_status, and tool_access fields.
func compareAgentBehavior(id string, ac, bc ir.AgentConfig) []Difference {
	var diffs []Difference
	if ac.GoalGate != bc.GoalGate {
		diffs = append(diffs, fieldDiff(id, "goal_gate", fmt.Sprintf("node %q goal_gate: %v vs %v", id, ac.GoalGate, bc.GoalGate)))
	}
	if ac.AutoStatus != bc.AutoStatus {
		diffs = append(diffs, fieldDiff(id, "auto_status", fmt.Sprintf("node %q auto_status: %v vs %v", id, ac.AutoStatus, bc.AutoStatus)))
	}
	if ac.ToolAccess != bc.ToolAccess {
		diffs = append(diffs, fieldDiff(id, "tool_access", fmt.Sprintf("node %q tool_access: %q vs %q", id, ac.ToolAccess, bc.ToolAccess)))
	}
	return diffs
}
```

Update the doc comment accordingly. Do NOT switch to `reflect.DeepEqual` on the full config — the existing comparator's 11-field gap is tracked separately (spec § Migrate).

- [ ] **Step 5: Run tests to verify they pass**

Run: `just test-pkg migrate`
Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add migrate/migrate.go migrate/parity.go migrate/migrate_test.go
git commit -m "feat(migrate): preserve tool_access through DOT round-trip + parity compare"
```

---

## Task 7: Round-trip integration test

**Files:**
- Test: `migrate/roundtrip_test.go` (new test function)

- [ ] **Step 1: Write the test**

Append to `migrate/roundtrip_test.go`:

```go
// TestRoundtripPreservesToolAccess verifies that AgentConfig.ToolAccess
// survives the .dip → DOT → migrate cycle.
//
// Bug shape: an author writes tool_access: none on a summarizer node, ships
// it to tracker via dipx bundle; the value must arrive intact at tracker
// for the safety primitive to bind. This test exercises the dippin half of
// that chain (parser → DOT export → migrate); the tracker half is covered
// in the tracker-side red-team test.
func TestRoundtripPreservesToolAccess(t *testing.T) {
	src := `workflow ToolAccessRT
  start: A
  exit: A

  agent A
    prompt: "x"
    tool_access: none
`
	p := parser.NewParser(src, "test.dip")
	original, err := p.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	dot, err := export.ExportDOT(original, export.ExportOptions{})
	if err != nil {
		t.Fatalf("export error: %v", err)
	}

	migrated, err := Migrate(dot)
	if err != nil {
		t.Fatalf("migrate error: %v", err)
	}

	diffs := CompareWorkflows(original, migrated)
	for _, d := range diffs {
		t.Errorf("round-trip diff: %s", d)
	}

	node := migrated.Node("A")
	if node == nil {
		t.Fatalf("node A missing after round-trip")
	}
	cfg, ok := node.Config.(ir.AgentConfig)
	if !ok {
		t.Fatalf("expected AgentConfig, got %T", node.Config)
	}
	if cfg.ToolAccess != "none" {
		t.Errorf("ToolAccess after round-trip = %q, want %q", cfg.ToolAccess, "none")
	}
}
```

`CompareWorkflows` (or the actual exported parity entry point) is what calls `compareAgentConfigs` under the hood — check the existing roundtrip tests for the right helper name. If it's `migrate.CompareWorkflows`, use that; if it's another name, match it.

- [ ] **Step 2: Run test**

Run: `just test-pkg migrate`
Expected: PASS (Tasks 1–6 supplied all the plumbing).

- [ ] **Step 3: Commit**

```bash
git add migrate/roundtrip_test.go
git commit -m "test(migrate): round-trip preserves tool_access"
```

---

## Task 8: Example file + integration test

**Files:**
- Create: `examples/agent_tool_access.dip`

- [ ] **Step 1: Create the example**

Write `examples/agent_tool_access.dip`:

```
workflow AgentToolAccess
  goal: "Demonstrate the tool_access agent-node safety primitive (issue #41)"
  start: Plan
  exit: ReportFinalStatus

  agent Plan
    model: claude-sonnet-4-6
    prompt: "Plan the work. Output a numbered task list."

  agent Implement
    model: claude-sonnet-4-6
    prompt: "Execute the plan."

  agent ReportFinalStatus
    model: claude-sonnet-4-6
    prompt: "Summarize what was implemented. Emit STATUS: success or STATUS: failure."
    tool_access: none
    auto_status: true

  edges
    Plan -> Implement
    Implement -> ReportFinalStatus
```

- [ ] **Step 2: Verify it validates**

Run: `just validate-examples`
Expected: success (the new example must pass `dippin validate`, like every other example).

- [ ] **Step 3: Verify lint integration**

Run: `just test-pkg validator`
Expected: PASS — specifically the existing `TestLintExamples` in `validator/lint_examples_test.go` parses every `examples/*.dip` and runs the full lint pipeline; this implicitly confirms the new file is DIP139-clean.

- [ ] **Step 4: Run `just check` (full suite)**

Run: `just check`
Expected: PASS — build, vet, fmt, race tests, complexity, validate-examples all green. If anything fails, fix before committing.

- [ ] **Step 5: Commit**

```bash
git add examples/agent_tool_access.dip
git commit -m "examples: add agent_tool_access.dip demonstrating issue #41 primitive"
```

---

## Task 9: Documentation updates

**Files:**
- Modify: `docs/validation.md` (add DIP139 section)
- Modify: `site/static/skill.md` (document `tool_access:` field)
- Modify: `CHANGELOG.md` (v0.32.0 entry)

- [ ] **Step 1: Add DIP139 to `docs/validation.md`**

Find an existing `### DIPNNN` heading in `docs/validation.md` (e.g., DIP127). Mirror its structure. Add a new section after DIP138:

```markdown
### DIP139 — invalid tool_access value on agent node

**Severity:** warning

**Triggers when:** an agent node has `tool_access:` set to a value other than `none` (case-insensitive) or empty.

**Fix:** use `tool_access: none` to disable LLM tools, or omit the field for the full catalog. Invalid values fall back to no-tools at runtime (fail-closed), so the diagnostic exists to surface the typo — runtime behavior is safe regardless.

**Example:**
```dip
agent ReportFinalStatus
  prompt: "Summarize the results"
  tool_access: nono   // DIP139: not recognized; runtime treats as 'none'
```

(Match the formatting convention of the surrounding DIP entries — they may use different heading levels or fenced-block flavors.)

- [ ] **Step 2: Document `tool_access:` in `site/static/skill.md`**

Locate the agent-node field table or section in `site/static/skill.md` (search: `grep -n "backend:" site/static/skill.md` to find the runtime-fields neighborhood). Add a new entry for `tool_access:` immediately after `backend:`:

````markdown
- `tool_access: none` — bound the LLM's tool catalog. When set to `none`, the agent's LLM session has no file-mutation tools (`Read`, `Write`, `Edit`, `ApplyPatch`, `Glob`, `GrepSearch`, `Bash`) and the API request sets `tool_choice: none`. Omit the field for the default behavior (full tool catalog). Single valid value in v1; richer policies tracked as follow-ups.

  **Threat model bounded:** the v0.28.2 single-agent multi-tool-call vector (an agent emits multiple tool calls in one LLM response before `max_turns` checks the cap). Use `tool_access: none` on summarizer / acknowledge / report nodes — any agent that should never make file mutations.

  **Non-goals (v1):** does NOT propagate across edges, parallel branches, or `manager_loop` child subgraphs. Does NOT prevent chain attacks where a `tool_access: none` agent's response is laundered into a downstream `tool_access: full` agent's prompt via `${ctx.last_response}`. See [follow-up issues](...) for those.

  **Not the same as `tool_commands_allow:`** (DIP28). `tool_access:` bounds LLM-emitted tool calls. `tool_commands_allow:` / `tool_denylist_add:` bound `tool` node shell execution. They address distinct threats.

  **Requires tracker ≥ vX.Y.Z** (will be filled in at release time once tracker tag is cut — see CHANGELOG).
````

After the tracker is tagged in Task 12, return here and replace `vX.Y.Z` with the actual tracker tag.

- [ ] **Step 3: Add CHANGELOG entry**

Edit `CHANGELOG.md`. Add at the top:

```markdown
## [v0.32.0] — <date-at-release-time>

New agent-node safety primitive: `tool_access: none` strips an LLM's tool catalog. Joint release with tracker `<tracker-tag>` — the dippin field is meaningless without tracker enforcement, so they ship together (see issue #41 for context, including the v0.28.2 runaway-agent incident this bounds).

### Added
- `tool_access:` field on agent nodes. One explicit value: `none` (no LLM tools). Omitted = full catalog (current behavior).
- DIP139 lint warns on invalid `tool_access` values.
- `examples/agent_tool_access.dip` demonstrates the field on a summarizer node.

### Tracker-side (linked to tracker tag `<tracker-tag>`)
- Tool registry returns empty when `tool_access: none` is set.
- Anthropic translator strips the `tools` array via `tool_choice: none`.
- System prompt scrubbed of tool-naming text when tools are disabled.
- `Params` keys (`allowed_tools`, `disallowed_tools`, `tool_choice`, `permission_mode`) are not honored when `tool_access: none` is set — Params bypass defense.
- Backend-compat tests: every supported backend honors `tool_access: none` or refuses session creation with a clear error.
- Red-team test: multi-tool-call LLM response under `tool_access: none` produces zero executions (the actual v0.28.2 shape).
```

The `<date-at-release-time>` and `<tracker-tag>` placeholders are filled in at Task 12.

- [ ] **Step 4: Run `just check`**

Run: `just check`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add docs/validation.md site/static/skill.md CHANGELOG.md
git commit -m "docs: document tool_access field, DIP139, and v0.32.0 changelog stub"
```

---

## Task 10: Terror-squad doc closing note

**Files:**
- Modify: `docs/superpowers/research/2026-05-19-issue-41-terror-squad.md` (update Status line)

- [ ] **Step 1: Update the Status line**

Edit `docs/superpowers/research/2026-05-19-issue-41-terror-squad.md`. Find the Status line near the top (line 4: `**Status:** Issue #41 parked at end of brainstorming. Returns in v0.31.0 as a **joint dippin + tracker** release.`). Replace with:

```markdown
**Status:** Shipped in v0.32.0 — see [`docs/superpowers/specs/2026-05-26-issue-41-design.md`](../specs/2026-05-26-issue-41-design.md). The simplification process is documented in the spec's "Design journey" section.
```

- [ ] **Step 2: Commit**

```bash
git add docs/superpowers/research/2026-05-19-issue-41-terror-squad.md
git commit -m "docs(research): mark issue #41 terror-squad doc as shipped"
```

---

## Task 11: Tracker-side PR (cross-repo umbrella)

**Repo:** `2389-research/tracker` (sibling repo; clone alongside `dippin-lang` if not present)

**Files** (tracker repo paths; exact code informed by the spec § Tracker-side design):
- Modify: `tracker/agent/session.go` (add `SessionConfig.ToolAccess string`)
- Modify: `tracker/agent/profile.go` (`builtInToolsForConfig` short-circuit)
- Modify: `tracker/agent/session_run.go` (set `ToolChoiceNone`, scrub system prompt)
- Modify: `tracker/pipeline/dippin_adapter.go` (`extractAgentAttrs` reads `tool_access`)
- Modify: `tracker/pipeline/handlers/codergen.go` and other Params-honoring codepaths (skip dangerous Params keys when `ToolAccess` is non-empty)
- Create: tracker-side test files for unit, integration, red-team, system-prompt audit, backend-compat (see spec § Tests / Tracker)

This task is **a single umbrella** — the tracker repo gets its own dedicated PR. The dippin spec describes the contract; the tracker engineer implements following the dippin spec § Tracker-side design.

- [ ] **Step 1: Clone tracker if not present**

If `../tracker` doesn't exist:

```bash
cd ..
gh repo clone 2389-research/tracker
cd dippin-lang
```

If access is restricted, escalate to the user before continuing.

- [ ] **Step 2: Verify tracker file paths cited in the spec**

```bash
ls ../tracker/tracker/agent/{session.go,profile.go,session_run.go}
ls ../tracker/tracker/pipeline/dippin_adapter.go
```

If any path doesn't exist, the spec's file references need updating — escalate to the user to reconcile before opening the tracker PR.

- [ ] **Step 3: Implement tracker changes per spec § Tracker-side design**

Follow the spec's tracker section (`docs/superpowers/specs/2026-05-26-issue-41-design.md`, the entire "Tracker-side design" subsection) as the source of truth. The implementer:

1. Adds `SessionConfig.ToolAccess string` to `tracker/agent/session.go`.
2. Inserts the case-normalized short-circuit at the top of `builtInToolsForConfig` in `tracker/agent/profile.go`:
   ```go
   canonical := strings.ToLower(strings.TrimSpace(cfg.ToolAccess))
   if canonical != "" {
       return []Tool{}
   }
   // existing return
   ```
3. In `session_run.go`, when `ToolAccess` is non-empty: set `request.ToolChoice = llm.ToolChoiceNone()` for Anthropic, and skip the "File tool arguments..." system-prompt prefix (line 24-31 of the existing file).
4. In `dippin_adapter.go::extractAgentAttrs`, read `tool_access` from `graph.Attrs` into `cfg.ToolAccess` (mirror the existing `Backend`/`WorkingDir` extraction).
5. In every codepath that translates `cfg.Params[...]` to runtime settings (audit: grep for `cfg.Params[` in `tracker/pipeline/`), check `cfg.ToolAccess` first; if non-empty, skip the bypass-eligible keys (`allowed_tools`, `disallowed_tools`, `tool_choice`, `permission_mode`).
6. Backend-compat: for `claude-code` and `acp`, verify the deny-equivalent spelling (cite docs URL + date in code comment) and ship a runtime test. If a backend can't be verified, that backend rejects session creation when `cfg.ToolAccess` is non-empty, with an error message pointing to the relevant follow-up issue.

- [ ] **Step 4: Write tracker-side tests per spec § Tests / Tracker**

Required test set (all live in the tracker repo, under `tracker/agent/` or wherever existing session tests live):

- **Unit:** `builtInToolsForConfig` returns empty slice when `cfg.ToolAccess = "none"`; non-empty `cfg.ToolAccess` of any value behaves the same (fail-closed). `session_run` sets `ToolChoiceNone` + omits tool-naming prefix.
- **Integration:** mocked-LLM test producing a single response with `[bash("rm -rf data/"), write("payload.py", "..."), bash("./payload.py")]` (the actual v0.28.2 multi-tool-call shape); assert zero tool-call executions. **This is the red-team test.**
- **Bypass tests:** `tool_access: "none"` + `params: {"allowed_tools": "Bash"}` → zero tools registered. `tool_access: "noen"` (typo) → zero tools (fail-closed). `tool_access: "None"` (case variant) → zero tools.
- **System-prompt audit:** assemble the system prompt under `tool_access: "none"`; assert no standalone case-insensitive occurrences of `read`, `write`, `edit`, `glob`, `grep_search`, `bash`, `apply_patch`.
- **Backend-compat:** for each backend (`native`, `claude-code`, `acp`): integration test with `tool_access: none` + mocked LLM emitting tool calls; assert zero executions OR session-creation refusal with a clear error.

- [ ] **Step 5: Open the tracker PR**

```bash
cd ../tracker
git checkout -b feat/issue-41-tool-access
# (after all changes + tests)
git push -u origin feat/issue-41-tool-access
gh pr create --title "feat: honor tool_access: none from dippin (issue #41 joint release)" --body "$(cat <<'EOF'
## Summary

Joint release with dippin-lang v0.32.0. Implements runtime enforcement for the new `tool_access:` agent-node field — see dippin spec at https://github.com/2389-research/dippin-lang/blob/main/docs/superpowers/specs/2026-05-26-issue-41-design.md for the full design.

When dippin sets `tool_access: none` on an agent node, this PR:
- Returns an empty tool registry from `builtInToolsForConfig`.
- Sets `tool_choice: none` on the Anthropic API request.
- Strips tool-naming text from the system prompt.
- Refuses to honor `params` bypass keys (`allowed_tools`, `disallowed_tools`, `tool_choice`, `permission_mode`).
- Verifies the deny-equivalent spelling for each backend or refuses session creation.

## Test plan
- [ ] Unit tests for `builtInToolsForConfig` short-circuit
- [ ] System-prompt audit test (no tool words when tool_access non-empty)
- [ ] Red-team multi-tool-call test (the v0.28.2 vector)
- [ ] Bypass tests (params keys, case variants, typos)
- [ ] Backend-compat tests (native, claude-code, acp)

🤖 Generated with [Claude Code](https://claude.com/claude-code)
EOF
)"
cd ../dippin-lang
```

- [ ] **Step 6: Record the tracker PR number + target commit SHA in this plan**

After the tracker PR is opened, record: **Tracker PR: #_____ — target SHA (once merged): _______**

Return to the dippin spec (`docs/superpowers/specs/2026-05-26-issue-41-design.md`) and update the `**Tracker dependency:**` line at the top with the PR URL.

- [ ] **Step 7: Commit dippin spec backlink**

```bash
git add docs/superpowers/specs/2026-05-26-issue-41-design.md
git commit -m "docs(spec): link tracker PR for v0.32.0 joint release"
```

---

## Task 12: Joint-release tagging

**Sequence (DO NOT REORDER):**

- [ ] **Step 1: Tracker PR review + merge**

The tracker PR (from Task 11) goes through normal review. **Do not tag tracker yet.** Wait for dippin PR (Task 13 below) to be approved first; tag tracker just before tagging dippin to minimize the window where the tracker version is published without dippin pointing at it.

- [ ] **Step 2: Open dippin PR**

From a feature branch in dippin (recommended: name it `feat/issue-41-tool-access` parallel to tracker's):

```bash
git checkout -b feat/issue-41-tool-access  # or rebase onto main if already named
git push -u origin feat/issue-41-tool-access
gh pr create --title "v0.32.0: tool_access: agent-node safety primitive (#41)" --body "$(cat <<'EOF'
## Summary

Closes #41. Joint release with tracker [PR link from Task 11].

- New `tool_access: none` field on agent nodes — bounds the v0.28.2 runaway-agent vector for any agent the author annotates.
- DIP139 lint warns on invalid values.
- Example, docs, CHANGELOG updated.
- 9 numbered follow-up issues filed and linked in the spec's § Non-goals.

## Spec
docs/superpowers/specs/2026-05-26-issue-41-design.md

## Test plan
- [ ] `just check` passes locally
- [ ] All examples lint-clean
- [ ] Round-trip test preserves `tool_access`
- [ ] DIP139 fires on invalid value, doesn't fire on valid

🤖 Generated with [Claude Code](https://claude.com/claude-code)
EOF
)"
```

- [ ] **Step 3: Use go.mod replace during PR review (optional)**

If integration tests need to run against the unmerged tracker PR, add a temporary `replace` directive to `go.mod`:

```
replace github.com/2389-research/tracker => ../tracker
```

Do NOT commit this `replace` directive. It's strictly for local validation during review. Remove before the final commit on the dippin PR.

- [ ] **Step 4: Final tag sequence (after both PRs approved)**

```bash
# In tracker repo (assumes you're authorized to tag and push):
cd ../tracker
git checkout main && git pull
git tag -a v<tracker-version> -m "tool_access runtime enforcement for dippin #41"
git push origin v<tracker-version>

# In dippin repo:
cd ../dippin-lang
git checkout main && git pull
# Bump tracker dep in go.mod to the newly-tagged version:
go get github.com/2389-research/tracker@v<tracker-version>
go mod tidy
just check  # verify
# Update CHANGELOG.md: replace <date-at-release-time> with today's date,
# and <tracker-tag> with the actual tracker tag. Same for skill.md's
# "Requires tracker ≥ vX.Y.Z" line.
git add CHANGELOG.md site/static/skill.md go.mod go.sum
git commit -m "release: prep v0.32.0 (tracker v<tracker-version>)"
git tag -a v0.32.0 -m "tool_access: agent-node safety primitive (#41)"
git push origin main
git push origin v0.32.0
```

GoReleaser handles cross-platform build + Homebrew tap publication on tag push.

- [ ] **Step 5: Verify GoReleaser ran**

Check `gh run list --workflow=release.yml --limit=1` (or whatever the release workflow is named — see `.github/workflows/`). If it failed, debug and re-run.

- [ ] **Step 6: Announce in CHANGELOG comments / PR thread**

Reply on issue #41 with: "Shipped in v0.32.0 (joint with tracker v<version>). See <CHANGELOG link>."

Close issue #41.

---

## Notes for the executing engineer

- **`just check` is your friend.** Run it after every task. The pre-commit hook runs it too; better to catch issues before pushing.
- **Cyclomatic complexity ≤ 5 / cognitive ≤ 7 per function.** If a function trips the limit, extract a helper. Never add `//nolint`. The new `lintToolAccessValues` is well under the limit (cyclomatic 3).
- **No hand-built IR in tests.** Every test parses real `.dip` text via `parser.NewParser(src, "test.dip").Parse()`. The DIP101 incident (root cause: tests pre-populated `Condition.Parsed` by hand, masking that production code never set it) is exactly the failure mode this rule prevents.
- **Tracker file paths in Task 11 are best-effort.** If `tracker/agent/session.go:213` doesn't match what you find when you clone, the spec's line references are stale. Don't fight the spec — re-read the actual code, escalate to the user if the spec's described logic doesn't match the current tracker.
- **Don't merge with bot review state in play.** Branch protection on dippin `main` is not enforced, so `gh pr merge` will go through even with a CodeRabbit `CHANGES_REQUESTED`. The project owner explicitly flagged this — always confirm with the user before merging if any review state is non-APPROVED, even from bots.
- **Follow-up issues stay filed.** Task 0 is non-optional. The DIP28 anti-pattern is that follow-ups promised in a spec never get filed; we break it by filing them BEFORE merge with `gh issue create`.
