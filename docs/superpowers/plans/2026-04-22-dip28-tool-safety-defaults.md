# DIP28 — Tool safety defaults implementation plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add `tool_commands_allow` and `tool_denylist_add` as first-class fields in `.dip` `defaults:` blocks, wire them through parser, DOT export, and DOT → dip migrate so authors no longer have to drop to raw DOT to use the runtime's tool-safety features.

**Architecture:** Two string fields on `ir.WorkflowDefaults`. Parser accepts the keys in `defaults:`; DOT export emits them as graph attributes; migrate routes them back into `defaults:` on reverse conversion. No validation — strings pass through verbatim to the runtime.

**Tech Stack:** Go. Tests use the existing `parseFixture` / `readTestdata` helpers in `parser/`, plain string-contains assertions in `export/dot_test.go`, and direct `Migrate(dot)` calls in `migrate/migrate_test.go`. Build via `just`, never raw `go` commands.

**Spec:** `docs/superpowers/specs/2026-04-22-dip28-tool-safety-defaults-design.md`
**Issue:** <https://github.com/2389-research/dippin-lang/issues/28>

**Files touched:**
- Create: `parser/testdata/defaults_tool_safety.dip`
- Create: `examples/tool_safety.dip`
- Modify: `ir/ir.go` — two new `WorkflowDefaults` fields
- Modify: `parser/parse_defaults.go` — new `applyDefaultToolField` leaf + wire-in
- Modify: `formatter/format.go` — new `writeDefaultsToolSafetyFields` helper + wire-in
- Modify: `parser/parser_test.go` — new parse + round-trip tests
- Modify: `export/dot.go` — emit both keys from `writeDOTHeader`, add to `reservedGraphAttrs`
- Modify: `export/dot_test.go` — new export test
- Modify: `migrate/migrate.go` — two entries in `graphDefaultsHandlers`
- Modify: `migrate/migrate_test.go` — new migrate test
- Modify: `migrate/roundtrip_test.go` — new `.dip` → DOT → `.dip` round-trip
- Modify: `CHANGELOG.md` — unreleased entry
- Modify: `QUICK_REFERENCE.md` — `defaults:` field list
- Modify: `site/content/language.md` — tool-safety subsection
- Modify: `site/static/skill.md` — both keys in the defaults reference

---

### Task 1: Add IR fields

**Files:**
- Modify: `ir/ir.go:39-53`

- [ ] **Step 1: Add both fields to `WorkflowDefaults`**

Open `ir/ir.go` and extend the struct at lines 39–53. Append two fields at the end, before the closing brace:

```go
// WorkflowDefaults holds graph-level configuration that applies to all nodes
// unless overridden at the node level.
type WorkflowDefaults struct {
	Model             string        // Default LLM model
	Provider          string        // Default LLM provider
	RetryPolicy       string        // Default retry policy name
	MaxRetries        int           // Default max retries
	Fidelity          string        // Default fidelity level
	MaxRestarts       int           // Max loop restarts (default 5)
	RestartTarget     string        // Where to restart on loop
	CacheTools        bool          // Cache tool results
	Compaction        string        // Context compaction mode
	OnResume          string        // Fidelity behavior on resume: "preserve" or "degrade"
	MaxTotalTokens    int           // Hard ceiling on total tokens
	MaxCostCents      int           // Hard ceiling on cost in cents (USD)
	MaxWallTime       time.Duration // Hard ceiling on wall time
	ToolCommandsAllow string        // Comma-separated glob allowlist for tool shell commands
	ToolDenylistAdd   string        // Comma-separated globs appended to the runtime's default denylist
}
```

- [ ] **Step 2: Verify compile**

Run: `just build`
Expected: builds cleanly. No test target here yet — fields are unused so far.

- [ ] **Step 3: Commit**

```bash
git add ir/ir.go
git commit -m "feat(ir): add ToolCommandsAllow and ToolDenylistAdd to WorkflowDefaults

Fields for tool-safety defaults consumed by the runtime via
graph.Attrs. Refs #28."
```

---

### Task 2: Parser — fixture + failing test

**Files:**
- Create: `parser/testdata/defaults_tool_safety.dip`
- Modify: `parser/parser_test.go` (append at end)

- [ ] **Step 1: Write the fixture**

Create `parser/testdata/defaults_tool_safety.dip` exactly:

```
workflow DefaultsToolSafety
  goal: "Test tool-safety defaults parsing"
  start: A
  exit: A

  defaults
    model: claude-sonnet-4-6
    tool_commands_allow: "git *,make *,npm test"
    tool_denylist_add: "rm -rf /,dd *"

  agent A
    prompt: "Do it."

  edges
    A -> A
```

- [ ] **Step 2: Write the failing test**

Append to `parser/parser_test.go`:

```go
func TestParseDefaultsToolSafety(t *testing.T) {
	w := parseFixture(t, "defaults_tool_safety.dip")
	d := w.Defaults
	if d.ToolCommandsAllow != "git *,make *,npm test" {
		t.Errorf("tool_commands_allow = %q, want %q", d.ToolCommandsAllow, "git *,make *,npm test")
	}
	if d.ToolDenylistAdd != "rm -rf /,dd *" {
		t.Errorf("tool_denylist_add = %q, want %q", d.ToolDenylistAdd, "rm -rf /,dd *")
	}
}
```

- [ ] **Step 3: Run the test and confirm it fails**

Run: `just test-pkg parser`
Expected: `TestParseDefaultsToolSafety` FAIL. The parser currently emits `unknown defaults field "tool_commands_allow"` diagnostics and `ToolCommandsAllow` stays empty, so the first assertion fires. (The fixture may also fail to Parse if diagnostics cause an error — acceptable as long as the failure message references the unknown field.)

- [ ] **Step 4: Commit the failing test**

```bash
git add parser/testdata/defaults_tool_safety.dip parser/parser_test.go
git commit -m "test(parser): add failing test for tool-safety defaults"
```

---

### Task 3: Parser — implementation

**Files:**
- Modify: `parser/parse_defaults.go`

- [ ] **Step 1: Add the leaf handler**

In `parser/parse_defaults.go`, add a new function directly after `applyDefaultExtraField` (around line 88):

```go
// applyDefaultToolField handles tool-safety defaults: tool_commands_allow and
// tool_denylist_add. Values are stored verbatim; the runtime owns split/glob semantics.
func applyDefaultToolField(d *ir.WorkflowDefaults, key, val string) bool {
	switch key {
	case "tool_commands_allow":
		d.ToolCommandsAllow = val
	case "tool_denylist_add":
		d.ToolDenylistAdd = val
	default:
		return false
	}
	return true
}
```

- [ ] **Step 2: Wire it into `applyDefaultStringField`**

Still in `parser/parse_defaults.go`, extend `applyDefaultStringField` (lines 50–56) to call the new handler:

```go
// applyDefaultStringField handles simple string assignments for defaults.
func (p *Parser) applyDefaultStringField(key, val string) bool {
	if applyDefaultCoreField(&p.workflow.Defaults, key, val) {
		return true
	}
	if applyDefaultExtraField(&p.workflow.Defaults, key, val) {
		return true
	}
	return applyDefaultToolField(&p.workflow.Defaults, key, val)
}
```

- [ ] **Step 3: Run the test and confirm it passes**

Run: `just test-pkg parser`
Expected: `TestParseDefaultsToolSafety` PASS. Whole package green.

- [ ] **Step 4: Confirm complexity budget**

Run: `just complexity`
Expected: clean — the leaf handler is 2 cases + default (well under cyclomatic 5 / cognitive 7).

- [ ] **Step 5: Commit**

```bash
git add parser/parse_defaults.go
git commit -m "feat(parser): accept tool_commands_allow and tool_denylist_add in defaults

Refs #28."
```

---

### Task 4: Formatter — emit new fields

The formatter writes `.dip` back out from IR. Without adding the new fields here, the round-trip test in Task 5 (`parse → format → re-parse`) will not see them in the re-emitted source.

**Files:**
- Modify: `formatter/format.go`

- [ ] **Step 1: Add a new helper**

Open `formatter/format.go` and add a new helper directly after `writeDefaultsBudgetFields` (around line 197):

```go
// writeDefaultsToolSafetyFields writes tool_commands_allow and tool_denylist_add.
func writeDefaultsToolSafetyFields(wr *writer, d ir.WorkflowDefaults) {
	if d.ToolCommandsAllow != "" {
		wr.line("tool_commands_allow: %s", quoteValue(d.ToolCommandsAllow))
	}
	if d.ToolDenylistAdd != "" {
		wr.line("tool_denylist_add: %s", quoteValue(d.ToolDenylistAdd))
	}
}
```

- [ ] **Step 2: Wire it into the chain**

The defaults-writer chain ends in `writeDefaultsBudgetFields`. Append a call to the new helper at the end of it (around line 196):

```go
// writeDefaultsBudgetFields writes max_total_tokens, max_cost_cents, and max_wall_time.
func writeDefaultsBudgetFields(wr *writer, d ir.WorkflowDefaults) {
	if d.MaxTotalTokens != 0 {
		wr.line("max_total_tokens: %d", d.MaxTotalTokens)
	}
	if d.MaxCostCents != 0 {
		wr.line("max_cost_cents: %d", d.MaxCostCents)
	}
	if d.MaxWallTime != 0 {
		wr.line("max_wall_time: %s", formatDuration(d.MaxWallTime))
	}
	writeDefaultsToolSafetyFields(wr, d)
}
```

- [ ] **Step 3: Confirm build + existing formatter tests still pass**

Run: `just test-pkg formatter`
Expected: all existing tests green. No new test needed here — Task 5's round-trip is the behavioral assertion.

- [ ] **Step 4: Check complexity**

Run: `just complexity`
Expected: clean.

- [ ] **Step 5: Commit**

```bash
git add formatter/format.go
git commit -m "feat(formatter): emit tool-safety defaults

Refs #28."
```

---

### Task 5: Parser — round-trip test

**Files:**
- Modify: `parser/parser_test.go`

- [ ] **Step 1: Write the round-trip test**

Append to `parser/parser_test.go` directly below `TestParseDefaultsToolSafety`:

```go
func TestParseDefaultsToolSafetyRoundTrip(t *testing.T) {
	w1 := parseFixture(t, "defaults_tool_safety.dip")
	formatted := formatter.Format(w1)
	w2, err := NewParser(formatted, "roundtrip").Parse()
	if err != nil {
		t.Fatalf("re-parse error: %v\nformatted:\n%s", err, formatted)
	}
	d := w2.Defaults
	if d.ToolCommandsAllow != "git *,make *,npm test" {
		t.Errorf("round-trip: tool_commands_allow = %q, want %q", d.ToolCommandsAllow, "git *,make *,npm test")
	}
	if d.ToolDenylistAdd != "rm -rf /,dd *" {
		t.Errorf("round-trip: tool_denylist_add = %q, want %q", d.ToolDenylistAdd, "rm -rf /,dd *")
	}
}
```

- [ ] **Step 2: Run the test**

Run: `just test-pkg parser`
Expected: both `TestParseDefaultsToolSafety` and `TestParseDefaultsToolSafetyRoundTrip` PASS. Task 4 already wired the formatter, so round-trip holds.

- [ ] **Step 3: Commit**

```bash
git add parser/parser_test.go
git commit -m "test(parser): round-trip parse → format → reparse for tool-safety defaults"
```

---

### Task 6: DOT export — failing test

**Files:**
- Modify: `export/dot_test.go`

- [ ] **Step 1: Write the failing test**

Append to `export/dot_test.go` at the end of the file (directly after `TestExportVarsSkipsDefaultsCollision`):

```go
func TestExportToolSafetyDefaults(t *testing.T) {
	w := &ir.Workflow{
		Name:  "ToolSafety",
		Start: "A",
		Exit:  "B",
		Defaults: ir.WorkflowDefaults{
			ToolCommandsAllow: "git *,make *",
			ToolDenylistAdd:   "rm -rf /",
		},
		Nodes: []*ir.Node{
			{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "a"}},
			{ID: "B", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "b"}},
		},
		Edges: []*ir.Edge{{From: "A", To: "B"}},
	}
	out := ExportDOT(w, ExportOptions{})

	if !strings.Contains(out, `tool_commands_allow="git *,make *"`) {
		t.Errorf("expected tool_commands_allow graph attr, got:\n%s", out)
	}
	if !strings.Contains(out, `tool_denylist_add="rm -rf /"`) {
		t.Errorf("expected tool_denylist_add graph attr, got:\n%s", out)
	}
}

func TestExportToolSafetyDefaultsOmitEmpty(t *testing.T) {
	w := &ir.Workflow{
		Name:  "Empty",
		Start: "A",
		Exit:  "B",
		Nodes: []*ir.Node{
			{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "a"}},
			{ID: "B", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "b"}},
		},
		Edges: []*ir.Edge{{From: "A", To: "B"}},
	}
	out := ExportDOT(w, ExportOptions{})

	if strings.Contains(out, "tool_commands_allow") {
		t.Errorf("empty field should not emit tool_commands_allow:\n%s", out)
	}
	if strings.Contains(out, "tool_denylist_add") {
		t.Errorf("empty field should not emit tool_denylist_add:\n%s", out)
	}
}

func TestExportToolSafetyVarsCollision(t *testing.T) {
	// A user mis-specifies the keys in vars:; export sets them from Defaults and
	// must not also emit them from Vars (double-emit would be invalid DOT).
	w := &ir.Workflow{
		Name:  "Collision",
		Start: "A",
		Exit:  "B",
		Defaults: ir.WorkflowDefaults{
			ToolCommandsAllow: "git *",
		},
		Vars: map[string]string{
			"tool_commands_allow": "should-be-skipped",
			"tool_denylist_add":   "also-skipped",
		},
		Nodes: []*ir.Node{
			{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "a"}},
			{ID: "B", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "b"}},
		},
		Edges: []*ir.Edge{{From: "A", To: "B"}},
	}
	out := ExportDOT(w, ExportOptions{})

	if strings.Contains(out, `tool_commands_allow="should-be-skipped"`) {
		t.Errorf("vars 'tool_commands_allow' should be skipped, got:\n%s", out)
	}
	if strings.Contains(out, `tool_denylist_add="also-skipped"`) {
		t.Errorf("vars 'tool_denylist_add' should be skipped, got:\n%s", out)
	}
	if !strings.Contains(out, `tool_commands_allow="git *"`) {
		t.Errorf("expected tool_commands_allow from Defaults, got:\n%s", out)
	}
}
```

- [ ] **Step 2: Run and confirm failure**

Run: `just test-pkg export`
Expected: all three new tests FAIL — the emitter doesn't exist yet; `reservedGraphAttrs` doesn't include the two new keys yet.

- [ ] **Step 3: Commit the failing tests**

```bash
git add export/dot_test.go
git commit -m "test(export): add failing tests for tool-safety defaults emission"
```

---

### Task 7: DOT export — implementation

**Files:**
- Modify: `export/dot.go`

- [ ] **Step 1: Extend `reservedGraphAttrs`**

In `export/dot.go` around lines 60–65, add both keys:

```go
// reservedGraphAttrs are DOT graph attributes used by the defaults/header — skip
// them when re-emitting workflow vars so they don't collide.
var reservedGraphAttrs = map[string]bool{
	"goal": true, "rankdir": true, "model": true, "provider": true,
	"fidelity": true, "default_fidelity": true,
	"max_retries": true, "default_max_retry": true, "max_restarts": true,
	"max_total_tokens": true, "max_cost_cents": true, "max_wall_time": true,
	"tool_commands_allow": true, "tool_denylist_add": true,
}
```

- [ ] **Step 2: Add an emit helper**

Still in `export/dot.go`, directly below `writeVarsAttrs` (around line 99), add:

```go
// writeToolSafetyAttrs emits tool-safety defaults as DOT graph-level attributes.
// Empty fields are skipped.
func writeToolSafetyAttrs(b *strings.Builder, d ir.WorkflowDefaults) {
	if d.ToolCommandsAllow != "" {
		fmt.Fprintf(b, "  tool_commands_allow=%s;\n", dotQuote(d.ToolCommandsAllow))
	}
	if d.ToolDenylistAdd != "" {
		fmt.Fprintf(b, "  tool_denylist_add=%s;\n", dotQuote(d.ToolDenylistAdd))
	}
}
```

- [ ] **Step 3: Call the helper from `writeDOTHeader`**

In `writeDOTHeader` (lines 67–84), add the call after the `writeVarsAttrs` block:

```go
func writeDOTHeader(b *strings.Builder, w *ir.Workflow, opts ExportOptions) {
	rankDir := opts.RankDir
	if rankDir == "" {
		rankDir = "TB"
	}
	graphName := w.Name
	if graphName == "" {
		graphName = "workflow"
	}
	fmt.Fprintf(b, "digraph %s {\n", dotID(graphName))
	fmt.Fprintf(b, "  rankdir=%s;\n", rankDir)
	b.WriteString("  node [fontname=\"Helvetica\"];\n")
	b.WriteString("  edge [fontname=\"Helvetica\"];\n")
	if len(w.Vars) > 0 {
		writeVarsAttrs(b, w.Vars)
	}
	writeToolSafetyAttrs(b, w.Defaults)
}
```

- [ ] **Step 4: Run the tests**

Run: `just test-pkg export`
Expected: all three new tests PASS. Package green.

- [ ] **Step 5: Check complexity**

Run: `just complexity`
Expected: clean — `writeDOTHeader` gains one statement; `writeToolSafetyAttrs` is a 2-branch helper.

- [ ] **Step 6: Commit**

```bash
git add export/dot.go
git commit -m "feat(export): emit tool-safety defaults as DOT graph attrs

Refs #28."
```

---

### Task 8: Migrate — failing test

**Files:**
- Modify: `migrate/migrate_test.go`

- [ ] **Step 1: Write the failing test**

Append to `migrate/migrate_test.go` directly below `TestMigrateGraphDefaults` (around line 500):

```go
func TestMigrateToolSafetyDefaults(t *testing.T) {
	dot := `digraph G {
		graph [tool_commands_allow="git *,make *", tool_denylist_add="rm -rf /"];
		Start [shape=Mdiamond];
		Exit [shape=Msquare];
		Start -> Exit;
	}`
	w, err := Migrate(dot)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if w.Defaults.ToolCommandsAllow != "git *,make *" {
		t.Errorf("tool_commands_allow = %q, want %q", w.Defaults.ToolCommandsAllow, "git *,make *")
	}
	if w.Defaults.ToolDenylistAdd != "rm -rf /" {
		t.Errorf("tool_denylist_add = %q, want %q", w.Defaults.ToolDenylistAdd, "rm -rf /")
	}
	// Must not leak into Vars.
	if _, ok := w.Vars["tool_commands_allow"]; ok {
		t.Errorf("tool_commands_allow should route to Defaults, not Vars")
	}
	if _, ok := w.Vars["tool_denylist_add"]; ok {
		t.Errorf("tool_denylist_add should route to Defaults, not Vars")
	}
}
```

- [ ] **Step 2: Run and confirm failure**

Run: `just test-pkg migrate`
Expected: `TestMigrateToolSafetyDefaults` FAIL — both attrs currently land in `w.Vars` via the `applyUnknownGraphAttr` fallthrough.

- [ ] **Step 3: Commit the failing test**

```bash
git add migrate/migrate_test.go
git commit -m "test(migrate): add failing test for tool-safety defaults routing"
```

---

### Task 9: Migrate — implementation

**Files:**
- Modify: `migrate/migrate.go`

- [ ] **Step 1: Add entries to `graphDefaultsHandlers`**

In `migrate/migrate.go` at the map literal (lines 113–121), add two entries:

```go
// graphDefaultsHandlers maps DOT graph attribute keys to handler functions.
var graphDefaultsHandlers = map[string]func(string, *ir.Workflow){
	"goal":                func(v string, w *ir.Workflow) { w.Goal = v },
	"rankdir":             func(_ string, _ *ir.Workflow) { /* presentation-only; ignored */ },
	"default_fidelity":    func(v string, w *ir.Workflow) { w.Defaults.Fidelity = v },
	"fidelity":            func(v string, w *ir.Workflow) { w.Defaults.Fidelity = v },
	"model":               func(v string, w *ir.Workflow) { w.Defaults.Model = v },
	"provider":            func(v string, w *ir.Workflow) { w.Defaults.Provider = v },
	"tool_commands_allow": func(v string, w *ir.Workflow) { w.Defaults.ToolCommandsAllow = v },
	"tool_denylist_add":   func(v string, w *ir.Workflow) { w.Defaults.ToolDenylistAdd = v },
}
```

- [ ] **Step 2: Run the test**

Run: `just test-pkg migrate`
Expected: `TestMigrateToolSafetyDefaults` PASS. All other migrate tests still green.

- [ ] **Step 3: Commit**

```bash
git add migrate/migrate.go
git commit -m "feat(migrate): route tool-safety DOT graph attrs into Defaults

Refs #28."
```

---

### Task 10: End-to-end round-trip test (.dip → DOT → .dip)

**Files:**
- Modify: `migrate/roundtrip_test.go`

- [ ] **Step 1: Inspect existing patterns**

Run: `grep -n "func Test\|ExportDOT\|Migrate(" migrate/roundtrip_test.go | head -20`
Expected: shows how existing round-trip tests structure the parse → export → migrate chain. Follow whichever pattern matches most closely. If the file imports both `parser` and `export` already, the new test fits with no new imports.

- [ ] **Step 2: Write the round-trip test**

Append to `migrate/roundtrip_test.go` (use the exact import aliases the file already declares — if a helper like `parseDip` exists, reuse it):

```go
func TestRoundTripToolSafetyDefaults(t *testing.T) {
	src := `workflow ToolSafety
  goal: "round trip"
  start: A
  exit: A

  defaults
    tool_commands_allow: "git *,make *"
    tool_denylist_add: "rm -rf /"

  agent A
    prompt: "Do it."

  edges
    A -> A
`
	w1, err := parser.NewParser(src, "rt.dip").Parse()
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	dot := export.ExportDOT(w1, export.ExportOptions{})
	w2, err := Migrate(dot)
	if err != nil {
		t.Fatalf("migrate: %v\nDOT:\n%s", err, dot)
	}
	if w2.Defaults.ToolCommandsAllow != "git *,make *" {
		t.Errorf("tool_commands_allow after round-trip = %q, want %q; DOT:\n%s",
			w2.Defaults.ToolCommandsAllow, "git *,make *", dot)
	}
	if w2.Defaults.ToolDenylistAdd != "rm -rf /" {
		t.Errorf("tool_denylist_add after round-trip = %q, want %q; DOT:\n%s",
			w2.Defaults.ToolDenylistAdd, "rm -rf /", dot)
	}
}
```

(If `parser` / `export` aren't already imported in this file, add them to the imports.)

- [ ] **Step 3: Run the test**

Run: `just test-pkg migrate`
Expected: `TestRoundTripToolSafetyDefaults` PASS. Whole package green.

- [ ] **Step 4: Commit**

```bash
git add migrate/roundtrip_test.go
git commit -m "test(migrate): .dip → DOT → .dip round-trip for tool-safety defaults"
```

---

### Task 11: Example `.dip` file

**Files:**
- Create: `examples/tool_safety.dip`

- [ ] **Step 1: Write the example**

Create `examples/tool_safety.dip`:

```
workflow ToolSafety
  goal: "Demonstrate tool_commands_allow and tool_denylist_add defaults for constrained shell execution"
  start: Start
  exit: Done

  defaults
    model: claude-sonnet-4-6
    # Restrict tool nodes to a glob allowlist. The runtime parses the comma-separated
    # list at runtime and rejects commands that don't match any pattern.
    tool_commands_allow: "git *,make *,npm test,npm run *"
    # Patterns appended to the runtime's default denylist, in addition to its
    # built-in blocks on destructive commands.
    tool_denylist_add: "rm -rf /,dd if=*"

  agent Start
    label: Start

  tool RunTests
    label: "Run Tests"
    command:
      set -eu
      npm test

  agent Done
    label: Done
    prompt: "Summarize the results."

  edges
    Start -> RunTests
    RunTests -> Done
    Done -> Done  when done
```

- [ ] **Step 2: Validate the example**

Run: `just validate-examples`
Expected: all examples (including the new one) validate cleanly.

Run: `just lint-examples`
Expected: no errors from `tool_safety.dip`. A DIP125 (shell-AST) warning on `npm test` is fine if it appears and is consistent with other examples — investigate if unique. Warnings about `start == exit` style should not appear given the layout above.

If lint flags the file, adjust the topology until it's clean (e.g. remove the self-loop, or add a proper exit node). The goal is an example that passes the same gates as the rest of `examples/`.

- [ ] **Step 3: Commit**

```bash
git add examples/tool_safety.dip
git commit -m "docs(examples): add tool_safety.dip demonstrating tool-safety defaults

Refs #28."
```

---

### Task 12: Docs updates

**Files:**
- Modify: `CHANGELOG.md`
- Modify: `QUICK_REFERENCE.md`
- Modify: `site/content/language.md`
- Modify: `site/static/skill.md`

- [ ] **Step 1: Update `CHANGELOG.md`**

Open `CHANGELOG.md` and add a new entry at the top, above `## [v0.22.0]`:

```markdown
## [Unreleased]

### Added
- **`WorkflowDefaults` tool-safety fields**: `tool_commands_allow` (glob allowlist for tool-node shell commands) and `tool_denylist_add` (globs appended to the runtime's default denylist). Both round-trip through parser → formatter → DOT export → migrate. Values pass through verbatim — the runtime owns split and glob semantics. ([#28](https://github.com/2389-research/dippin-lang/issues/28))
```

- [ ] **Step 2: Update `QUICK_REFERENCE.md`**

Run: `grep -n "max_total_tokens\|max_cost_cents\|max_wall_time\|defaults" QUICK_REFERENCE.md | head -20`
Expected: shows the section listing `defaults:` fields. Add two rows beneath the budget fields with brief one-line descriptions:

- `tool_commands_allow: "git *,make *"` — glob allowlist for tool-node shell commands (consumed by the runtime)
- `tool_denylist_add: "rm -rf /,dd *"` — glob patterns appended to the runtime's default denylist

Match the exact row format the existing entries use (table cells, bulleted fields, whatever QUICK_REFERENCE.md is using — don't change the style).

- [ ] **Step 3: Update `site/content/language.md`**

Run: `grep -n "max_total_tokens\|defaults\|## Defaults\|### defaults" site/content/language.md | head -10`
Expected: locate the defaults reference section. Add a new subsection "Tool safety" at a sensible slot within the defaults reference:

```markdown
### Tool safety

Tool nodes that shell out can be constrained by two defaults consumed by the runtime:

- `tool_commands_allow` — comma-separated glob allowlist. When set, the runtime rejects tool-node commands that do not match any pattern.
- `tool_denylist_add` — comma-separated globs appended to the runtime's default denylist (on top of the runtime's built-in blocks).

```dippin
workflow Safe
  goal: "Constrained tool execution"
  start: A
  exit: A

  defaults
    tool_commands_allow: "git *,make *"
    tool_denylist_add: "rm -rf /,dd *"

  # ...
```

Values pass through to the runtime verbatim; dippin-lang does not validate glob syntax.
```

- [ ] **Step 4: Update `site/static/skill.md`**

Run: `grep -n "max_total_tokens\|defaults" site/static/skill.md | head -10`
Expected: locate the defaults reference within the hosted skill. Add both keys to whatever field list it uses, with one-line descriptions, matching the existing style:

- `tool_commands_allow` — glob allowlist for tool-node shell commands (string, comma-separated; optional)
- `tool_denylist_add` — glob patterns appended to the runtime's default denylist (string, comma-separated; optional)

- [ ] **Step 5: Commit**

```bash
git add CHANGELOG.md QUICK_REFERENCE.md site/content/language.md site/static/skill.md
git commit -m "docs: document tool_commands_allow and tool_denylist_add defaults

Refs #28."
```

---

### Task 13: Final verification

- [ ] **Step 1: Run full check**

Run: `just check`
Expected: all phases pass — build, vet, fmt, test-race, complexity, validate-examples.

If any phase fails, fix root cause (no `//nolint` suppressions). Re-run until green.

- [ ] **Step 2: Confirm no stale memory of the failing acceptance criteria**

Run: `grep -n "tool_commands_allow\|tool_denylist_add" $(git ls-files) | grep -v docs/superpowers`
Expected: shows references in `ir/ir.go`, `parser/parse_defaults.go`, `parser/parser_test.go`, `parser/testdata/defaults_tool_safety.dip`, `export/dot.go`, `export/dot_test.go`, `migrate/migrate.go`, `migrate/migrate_test.go`, `migrate/roundtrip_test.go`, `examples/tool_safety.dip`, `CHANGELOG.md`, `QUICK_REFERENCE.md`, `site/content/language.md`, `site/static/skill.md`.

Missing file = incomplete work.

- [ ] **Step 3: Close the loop on the issue**

When the PR merges, the issue closes automatically via the `Refs #28` commit trail plus the PR description. Reviewer to tag `v0.23.0` after merge per the project's versioning convention.

---

## Out of scope (explicit non-goals)

- Validating glob pattern syntax on either field. The runtime owns these semantics.
- Emitting any other `WorkflowDefaults` fields (`Model`, `Provider`, `MaxRetries`, …) as DOT graph attrs. Pre-existing gap.
- Lint rules (DIP codes) for the new keys. Add a follow-up issue if misuse becomes real.
- Blog post — operational plumbing, not an authoring-visible feature.
- Tagging `v0.23.0` — separate manual step after merge.
