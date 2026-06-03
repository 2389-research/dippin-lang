# Vars Block Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a `vars` block to Dippin workflows for declaring user-defined key-value variables that export as DOT graph attributes.

**Architecture:** `Workflow.Vars map[string]string` is a new top-level IR field. The parser handles `vars` as a workflow-level block (peer to `defaults`, `edges`). The formatter emits it between defaults and nodes. DOT export emits vars as graph-level attributes. DOT migrate captures unknown graph attributes into Vars.

**Tech Stack:** Go, dippin-lang IR/parser/formatter/export/migrate packages

---

## File Map

| File | Action | Responsibility |
|------|--------|----------------|
| `ir/ir.go` | Modify | Add `Vars map[string]string` to `Workflow` |
| `parser/parser.go` | Modify | Add `"vars"` case to `dispatchWorkflowField` |
| `parser/parse_vars.go` | Create | Parse `vars` block (key-value pairs) |
| `parser/parser_test.go` | Modify | Add parse tests for vars |
| `formatter/format.go` | Modify | Emit `vars` block between defaults and nodes |
| `formatter/format_test.go` | Modify | Add formatter tests for vars |
| `export/dot.go` | Modify | Emit vars as graph-level attributes |
| `export/dot_test.go` | Modify | Add DOT export test for vars |
| `migrate/migrate.go` | Modify | Catch-all for unknown graph attrs → `Workflow.Vars` |
| `migrate/migrate_test.go` | Modify | Add migrate test for vars round-trip |
| `examples/semport_thematic.dip` | Modify | Add `vars` block with existing `$` placeholders |
| `docs/GRAMMAR.ebnf` | Modify | Add `vars_section` production rule |
| `docs/nodes.md` | Modify | Document vars block |
| `site/content/language.md` | Modify | Document vars block |
| `docs/llm-reference.md` | Modify | Mention vars in workflow structure |
| `CHANGELOG.md` | Modify | Add entry |
| `site/content/changelog.md` | Modify | Add entry |

---

## Task 1: IR — Add `Vars` to `Workflow`

**Files:**
- Modify: `ir/ir.go:11-22`

- [ ] **Step 1: Add `Vars` field to `Workflow` struct**

In `ir/ir.go`, add `Vars` after `Defaults`:

```go
type Workflow struct {
	Name       string
	Version    string           // Dippin format version
	Goal       string           // Human-readable objective
	Start      string           // Explicit entry node ID (required)
	Exit       string           // Explicit exit node ID (required)
	Defaults   WorkflowDefaults // Graph-level config
	Vars       map[string]string // User-defined workflow variables
	Nodes      []*Node          // Ordered for deterministic processing
	Edges      []*Edge
	Stylesheet []StylesheetRule // Theme/styling rules
	SourceMap  *SourceMap       // File/line mapping for diagnostics
}
```

- [ ] **Step 2: Verify it compiles**

Run: `just build`
Expected: PASS (no code references `Vars` yet, zero-value `nil` is fine)

- [ ] **Step 3: Commit**

```bash
git add ir/ir.go
git commit -m "feat(ir): add Vars map to Workflow struct"
```

---

## Task 2: Parser — Parse `vars` Block

**Files:**
- Create: `parser/parse_vars.go`
- Modify: `parser/parser.go:87-100`
- Test: `parser/parser_test.go`

- [ ] **Step 1: Write the failing test**

In `parser/parser_test.go`, add:

```go
func TestParseVarsBlock(t *testing.T) {
	input := `workflow Test
  start: A
  exit: B

  vars
    source_ref: "references/sdk-python/src"
    target_name: my-crate

  agent A
    prompt: go
  agent B
    prompt: done
  edges
    A -> B
`
	p := NewParser(input, "test.dip")
	w, err := p.Parse()
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	if w.Vars == nil {
		t.Fatal("expected Vars to be populated")
	}
	if got := w.Vars["source_ref"]; got != "references/sdk-python/src" {
		t.Errorf("source_ref = %q, want %q", got, "references/sdk-python/src")
	}
	if got := w.Vars["target_name"]; got != "my-crate" {
		t.Errorf("target_name = %q, want %q", got, "my-crate")
	}
	if len(w.Vars) != 2 {
		t.Errorf("expected 2 vars, got %d", len(w.Vars))
	}
}

func TestParseVarsDuplicateKey(t *testing.T) {
	input := `workflow Test
  start: A
  exit: B

  vars
    key: first
    key: second

  agent A
    prompt: go
  agent B
    prompt: done
  edges
    A -> B
`
	p := NewParser(input, "test.dip")
	_, err := p.Parse()
	if err == nil {
		t.Fatal("expected parse error for duplicate vars key")
	}
	if !strings.Contains(err.Error(), "duplicate") {
		t.Errorf("error should mention 'duplicate', got: %s", err.Error())
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `just test-pkg parser`
Expected: FAIL — `vars` keyword is unrecognized at workflow level.

- [ ] **Step 3: Create `parser/parse_vars.go`**

```go
package parser

import (
	"fmt"

	"github.com/2389-research/dippin-lang/ir"
)

func (p *Parser) parseVars() {
	p.lexer.NextToken() // vars
	p.expect(TokenNewline)
	p.expect(TokenIndent)
	p.parseVarsBody()
	p.expect(TokenOutdent)
}

// parseVarsBody parses the indented body of a vars block.
func (p *Parser) parseVarsBody() {
	if p.workflow.Vars == nil {
		p.workflow.Vars = make(map[string]string)
	}
	for p.lexer.PeekToken().Type != TokenOutdent && p.lexer.PeekToken().Type != TokenEOF {
		t := p.lexer.PeekToken()
		if t.Type == TokenNewline {
			p.lexer.NextToken()
			continue
		}
		if t.Type == TokenIdentifier {
			p.parseVarField(t)
		} else {
			p.lexer.NextToken()
		}
	}
}

// parseVarField reads a single var field (key: value).
func (p *Parser) parseVarField(t Token) {
	key := t.Value
	p.lexer.NextToken()
	p.expect(TokenColon)
	val := p.readFieldValue(t.Location.Line)
	p.applyVarField(key, val, t.Location)
}

// applyVarField adds a var to the workflow, checking for duplicates.
func (p *Parser) applyVarField(key, val string, loc ir.SourceLocation) {
	if _, exists := p.workflow.Vars[key]; exists {
		p.diagnostics = append(p.diagnostics,
			fmt.Sprintf("duplicate vars key %q at %d:%d", key, loc.Line, loc.Column))
	}
	p.workflow.Vars[key] = val
}
```

- [ ] **Step 4: Wire `vars` into the parser dispatch**

In `parser/parser.go`, add `"vars"` case to `dispatchWorkflowField`:

```go
func (p *Parser) dispatchWorkflowField(t Token) {
	switch t.Value {
	case "goal", "start", "exit":
		p.parseWorkflowStringField(t)
	case "defaults":
		p.parseDefaults()
	case "vars":
		p.parseVars()
	case "edges":
		p.parseEdges()
	case "stylesheet":
		p.parseStylesheet()
	default:
		p.dispatchWorkflowBlock(t)
	}
}
```

- [ ] **Step 5: Run tests to verify they pass**

Run: `just test-pkg parser`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add parser/parse_vars.go parser/parser.go parser/parser_test.go
git commit -m "feat(parser): parse vars block with duplicate key detection"
```

---

## Task 3: Formatter — Emit `vars` Block

**Files:**
- Modify: `formatter/format.go:19-45`
- Test: `formatter/format_test.go`

- [ ] **Step 1: Write the failing test**

In `formatter/format_test.go`, add:

```go
func TestFormatVarsBlock(t *testing.T) {
	w := &ir.Workflow{
		Name:  "test",
		Start: "A",
		Exit:  "B",
		Vars: map[string]string{
			"source_ref":  "references/sdk-python/src",
			"target_name": "my-crate",
		},
		Nodes: []*ir.Node{
			{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "go"}},
			{ID: "B", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "done"}},
		},
		Edges: []*ir.Edge{{From: "A", To: "B"}},
	}
	got := Format(w)
	if !strings.Contains(got, "vars") {
		t.Errorf("expected 'vars' block in output, got:\n%s", got)
	}
	if !strings.Contains(got, "source_ref:") {
		t.Errorf("expected 'source_ref:' in vars block, got:\n%s", got)
	}
	if !strings.Contains(got, "target_name:") {
		t.Errorf("expected 'target_name:' in vars block, got:\n%s", got)
	}
	// source_ref should come before target_name (sorted)
	srcIdx := strings.Index(got, "source_ref:")
	tgtIdx := strings.Index(got, "target_name:")
	if srcIdx > tgtIdx {
		t.Error("vars should be sorted alphabetically")
	}
}

func TestFormatVarsEmptyOmitted(t *testing.T) {
	w := &ir.Workflow{
		Name:  "test",
		Start: "A",
		Exit:  "B",
		Nodes: []*ir.Node{
			{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "go"}},
			{ID: "B", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "done"}},
		},
		Edges: []*ir.Edge{{From: "A", To: "B"}},
	}
	got := Format(w)
	if strings.Contains(got, "vars") {
		t.Errorf("empty vars should not appear in output, got:\n%s", got)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `just test-pkg formatter`
Expected: FAIL — formatter doesn't emit `vars` block.

- [ ] **Step 3: Add `writeVars` to the formatter**

In `formatter/format.go`, add the `writeVars` function:

```go
func writeVars(wr *writer, vars map[string]string) {
	wr.line("vars")
	wr.push()
	keys := make([]string, 0, len(vars))
	for k := range vars {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		wr.line("%s: %s", k, quoteValue(vars[k]))
	}
	wr.pop()
}
```

In `Format()`, add vars emission between defaults and nodes:

```go
func Format(w *ir.Workflow) string {
	wr := &writer{}

	writeWorkflowHeader(wr, w)

	if !isDefaultsZero(w.Defaults) {
		wr.blank()
		writeDefaults(wr, w.Defaults)
	}

	if len(w.Vars) > 0 {
		wr.blank()
		writeVars(wr, w.Vars)
	}

	for _, n := range w.Nodes {
		wr.blank()
		writeNode(wr, n)
	}

	// ... rest unchanged
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `just test-pkg formatter`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add formatter/format.go formatter/format_test.go
git commit -m "feat(formatter): emit vars block with sorted keys"
```

---

## Task 4: DOT Export — Emit Vars as Graph Attributes

**Files:**
- Modify: `export/dot.go:58-72`
- Test: `export/dot_test.go`

- [ ] **Step 1: Write the failing test**

In `export/dot_test.go`, add:

```go
func TestExportVarsAsGraphAttrs(t *testing.T) {
	w := &ir.Workflow{
		Name:  "test",
		Start: "A",
		Exit:  "B",
		Vars: map[string]string{
			"source_ref":  "refs/sdk/src",
			"target_name": "my-crate",
		},
		Nodes: []*ir.Node{
			{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "go"}},
			{ID: "B", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "done"}},
		},
		Edges: []*ir.Edge{{From: "A", To: "B"}},
	}
	dot := ExportDOT(w, ExportOptions{})
	if !strings.Contains(dot, `source_ref=`) {
		t.Errorf("expected source_ref graph attribute, got:\n%s", dot)
	}
	if !strings.Contains(dot, `target_name=`) {
		t.Errorf("expected target_name graph attribute, got:\n%s", dot)
	}
}

func TestExportVarsSkipsDefaultsCollision(t *testing.T) {
	w := &ir.Workflow{
		Name:  "test",
		Start: "A",
		Exit:  "B",
		Vars: map[string]string{
			"model":    "should-be-skipped",
			"safe_var": "kept",
		},
		Nodes: []*ir.Node{
			{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "go"}},
			{ID: "B", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "done"}},
		},
		Edges: []*ir.Edge{{From: "A", To: "B"}},
	}
	dot := ExportDOT(w, ExportOptions{})
	if strings.Contains(dot, "should-be-skipped") {
		t.Errorf("vars that collide with known graph attrs should be skipped, got:\n%s", dot)
	}
	if !strings.Contains(dot, `safe_var=`) {
		t.Errorf("non-colliding vars should be emitted, got:\n%s", dot)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `just test-pkg export`
Expected: FAIL

- [ ] **Step 3: Implement vars emission in DOT header**

In `export/dot.go`, add a function to write vars and a set of reserved names:

```go
// reservedGraphAttrs are DOT graph attribute names used by defaults/header.
// Vars with these names are skipped during export to avoid collisions.
var reservedGraphAttrs = map[string]bool{
	"goal": true, "rankdir": true, "model": true, "provider": true,
	"fidelity": true, "default_fidelity": true,
	"max_retries": true, "default_max_retry": true, "max_restarts": true,
}

// writeVarsAttrs emits workflow vars as graph-level attributes.
func writeVarsAttrs(b *strings.Builder, vars map[string]string) {
	keys := make([]string, 0, len(vars))
	for k := range vars {
		if !reservedGraphAttrs[k] {
			keys = append(keys, k)
		}
	}
	sortStrings(keys)
	for _, k := range keys {
		fmt.Fprintf(b, "  %s=%s;\n", k, dotQuote(vars[k]))
	}
}
```

In `writeDOTHeader`, add a call after the existing header lines:

```go
func writeDOTHeader(b *strings.Builder, w *ir.Workflow, opts ExportOptions) {
	// ... existing code ...
	b.WriteString("  edge [fontname=\"Helvetica\"];\n")
	if len(w.Vars) > 0 {
		writeVarsAttrs(b, w.Vars)
	}
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `just test-pkg export`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add export/dot.go export/dot_test.go
git commit -m "feat(export): emit workflow vars as DOT graph attributes"
```

---

## Task 5: DOT Migrate — Capture Unknown Graph Attrs as Vars

**Files:**
- Modify: `migrate/migrate.go:121-144`
- Test: `migrate/migrate_test.go`

- [ ] **Step 1: Write the failing test**

In `migrate/migrate_test.go`, add:

```go
func TestMigrateGraphAttrsToVars(t *testing.T) {
	dot := `digraph test {
  source_ref="references/sdk/src";
  target_name="my-crate";
  model="claude-opus-4-6";
  Start [shape=Mdiamond];
  End [shape=Msquare];
  Start -> End;
}`
	w, err := Migrate(dot)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// model should go to Defaults, not Vars
	if w.Defaults.Model != "claude-opus-4-6" {
		t.Errorf("model should be in defaults, got %q", w.Defaults.Model)
	}
	// source_ref and target_name should be in Vars
	if w.Vars == nil {
		t.Fatal("expected Vars to be populated")
	}
	if got := w.Vars["source_ref"]; got != "references/sdk/src" {
		t.Errorf("source_ref = %q, want %q", got, "references/sdk/src")
	}
	if got := w.Vars["target_name"]; got != "my-crate" {
		t.Errorf("target_name = %q, want %q", got, "my-crate")
	}
	// model should NOT be in Vars
	if _, ok := w.Vars["model"]; ok {
		t.Error("model should not be in Vars (it's a known default)")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `just test-pkg migrate`
Expected: FAIL — unknown graph attrs are silently dropped.

- [ ] **Step 3: Modify `extractGraphDefaults` to capture unknown attrs**

In `migrate/migrate.go`, modify `applyIntDefault` to capture non-integer unknown attrs into `Workflow.Vars`:

Replace `applyIntDefault` with a function that tries integer parsing first, then falls back to vars:

```go
// applyUnknownGraphAttr handles graph attributes not in the handler map.
// Integer values go to known int defaults; everything else goes to Vars.
func applyUnknownGraphAttr(k, v string, w *ir.Workflow) {
	if applyIntDefault(k, v, w) {
		return
	}
	if w.Vars == nil {
		w.Vars = make(map[string]string)
	}
	w.Vars[k] = v
}

// applyIntDefault handles integer-valued graph defaults. Returns true if handled.
func applyIntDefault(k, v string, w *ir.Workflow) bool {
	n, err := strconv.Atoi(v)
	if err != nil {
		return false
	}
	switch k {
	case "default_max_retry", "max_retries":
		w.Defaults.MaxRetries = n
	case "max_restarts":
		w.Defaults.MaxRestarts = n
	default:
		return false
	}
	return true
}
```

In `extractGraphDefaults`, change the fallback call:

```go
func extractGraphDefaults(attrs map[string]string, w *ir.Workflow) {
	for k, v := range attrs {
		if handler, ok := graphDefaultsHandlers[k]; ok {
			handler(v, w)
			continue
		}
		applyUnknownGraphAttr(k, v, w)
	}
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `just test-pkg migrate`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add migrate/migrate.go migrate/migrate_test.go
git commit -m "feat(migrate): capture unknown DOT graph attributes as workflow vars"
```

---

## Task 6: Update Example — `semport_thematic.dip`

**Files:**
- Modify: `examples/semport_thematic.dip`

- [ ] **Step 1: Add `vars` block to `semport_thematic.dip`**

After the `defaults` block (line 8), add:

```dip
  vars
    source_ref: "references/claude-agent-sdk-python/src"
    target_name: "claude-agents-rs"
    target_module: "claude-agents-rs/src/"
```

- [ ] **Step 2: Verify the example validates**

Run: `just validate-examples`
Expected: PASS

- [ ] **Step 3: Verify the example lints cleanly**

Run: `just lint-examples`
Expected: No new warnings.

- [ ] **Step 4: Commit**

```bash
git add examples/semport_thematic.dip
git commit -m "docs: add vars block to semport_thematic example"
```

---

## Task 7: Documentation and Full Validation

**Files:**
- Modify: `docs/GRAMMAR.ebnf`
- Modify: `docs/nodes.md`
- Modify: `docs/llm-reference.md`
- Modify: `site/content/language.md`
- Modify: `CHANGELOG.md`
- Modify: `site/content/changelog.md`

- [ ] **Step 1: Update GRAMMAR.ebnf**

Add `vars_section` production rule and add it to `workflow_body`:

```ebnf
workflow_body   = { workflow_field | defaults_section | vars_section
                  | node_decl | edges_section
                  | stylesheet_section | NEWLINE } ;
```

Add after `defaults_section`:

```ebnf
(* Vars Section                                               *)
(* ═══════════════════════════════════════════════════════════ *)

vars_section     = "vars" NEWLINE INDENT { vars_field } OUTDENT ;

vars_field       = IDENTIFIER ":" field_value ;
```

- [ ] **Step 2: Update `docs/nodes.md`**

Add a "Vars Block" section after the "Defaults" section, documenting:
- Purpose: declare workflow-level variables for runtime expansion
- Syntax: `vars` block with key-value pairs
- DOT export: emitted as graph-level attributes
- Example showing the `vars` block

- [ ] **Step 3: Update `docs/llm-reference.md`**

Add `vars` to the workflow structure description.

- [ ] **Step 4: Update `site/content/language.md`**

Add vars block documentation to the website language reference.

- [ ] **Step 5: Update changelogs**

Add `[Unreleased]` section to `CHANGELOG.md`:

```markdown
## [Unreleased]

### Added
- **`vars` block** at the workflow level for declaring user-defined variables. Vars export as DOT graph-level attributes and round-trip through parse → format → export → migrate.
```

Mirror in `site/content/changelog.md`.

- [ ] **Step 6: Run full check**

Run: `just check`
Expected: All checks pass.

- [ ] **Step 7: Commit**

```bash
git add docs/GRAMMAR.ebnf docs/nodes.md docs/llm-reference.md site/content/language.md CHANGELOG.md site/content/changelog.md
git commit -m "docs: vars block in GRAMMAR, nodes.md, language.md, changelogs"
```

---

## Task 8: Round-Trip Test and Final PR

- [ ] **Step 1: Add a parse → format → re-parse round-trip test**

In `parser/parser_test.go`, add:

```go
func TestParseVarsRoundTrip(t *testing.T) {
	input := `workflow Test
  start: A
  exit: B

  vars
    source_ref: "references/sdk-python/src"
    target_name: my-crate

  agent A
    prompt: go

  agent B
    prompt: done

  edges
    A -> B
`
	p := NewParser(input, "test.dip")
	w, err := p.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	formatted := formatter.Format(w)
	p2 := NewParser(formatted, "roundtrip.dip")
	w2, err := p2.Parse()
	if err != nil {
		t.Fatalf("re-parse error: %v\nformatted:\n%s", err, formatted)
	}
	if len(w2.Vars) != 2 {
		t.Errorf("expected 2 vars after round-trip, got %d", len(w2.Vars))
	}
	if w2.Vars["source_ref"] != "references/sdk-python/src" {
		t.Errorf("source_ref lost in round-trip: %q", w2.Vars["source_ref"])
	}
	if w2.Vars["target_name"] != "my-crate" {
		t.Errorf("target_name lost in round-trip: %q", w2.Vars["target_name"])
	}
}
```

- [ ] **Step 2: Run full check**

Run: `just check`
Expected: All checks pass.

- [ ] **Step 3: Commit**

```bash
git add parser/parser_test.go
git commit -m "test: vars round-trip through parse → format → re-parse"
```

- [ ] **Step 4: Create PR**

```bash
gh pr create --title "feat: workflow-level vars block" --body "$(cat <<'EOF'
## Summary

Adds a `vars` block at the workflow level for declaring user-defined key-value variables:

```dip
vars
  source_ref: "references/sdk-python/src"
  target_name: my-crate
```

- Stored as `Workflow.Vars map[string]string` in the IR
- Exports as DOT graph-level attributes (what the runtime reads via `graph.Attrs`)
- DOT migration captures unknown graph attributes into Vars
- Parser detects duplicate var keys
- Formatter emits sorted keys between defaults and nodes
- Vars that collide with known defaults names (model, provider, etc.) are skipped during DOT export

## Test plan
- [ ] `just check` passes
- [ ] Parse test: vars block populates `Workflow.Vars`
- [ ] Parse test: duplicate var key emits diagnostic
- [ ] Formatter test: vars block emitted with sorted keys, omitted when empty
- [ ] DOT export test: vars as graph attributes, defaults collisions skipped
- [ ] Migrate test: unknown graph attrs → `Workflow.Vars`, known attrs → `Defaults`
- [ ] Round-trip test: parse → format → re-parse preserves vars

Closes #5
EOF
)"
```
