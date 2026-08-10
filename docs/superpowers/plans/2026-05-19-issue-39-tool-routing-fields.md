# Tool Routing Fields (issue #39) Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add `marker_grep`, `route_required`, `output_limit` to `tool` nodes so authors following the runtime's TRK101 advice stop hitting `unrecognized tool field` parse errors. Strings pass through verbatim to the runtime via DOT.

**Architecture:** Extend `ir.ToolConfig` with three exported fields. Mirror the existing `apply{Agent,Human}{String,Bool,Parsed}Field` split in the tool parser to stay under the cyclomatic-5 budget. Formatter and DOT exporter use the same "emit when non-default" contract so `.dip ⇄ DOT` round-trips are idempotent. Migrate reads attrs back. Tests, docs, hosted skill, blog, CHANGELOG. Single PR; tag v0.28.0; downstream consumers bump their dep separately.

**Tech Stack:** Go 1.25, `just` task runner, no new deps. Hugo for site. Tree-sitter / Zed unaffected (generic `field_name` rule already).

**Spec:** [`docs/superpowers/specs/2026-05-19-issue-39-tool-routing-fields-design.md`](../specs/2026-05-19-issue-39-tool-routing-fields-design.md)

---

## Prerequisite reading for the implementing engineer

Skim these before starting — they establish the patterns this plan mirrors:

- `ir/ir.go:127-134` — current `ToolConfig` shape and adjacent configs.
- `parser/parse_nodes.go:406-417` — current `applyToolField` (4 cases, at the cyclo-5 budget).
- `parser/parse_nodes.go:232-343` — `applyAgentField` family — the split pattern this plan copies for tool.
- `parser/parse_nodes.go:346-403` — `applyHumanField` family — secondary reference for the split pattern.
- `parser/parse_helpers.go:45-67` — `parseInt`, `parseDuration` (existing helpers; no new ones needed).
- `formatter/format.go:488-500` — current `writeToolFields`.
- `export/dot.go:303-315` — `applyToolSemanticAttrs` / `applyToolPromptAttrs`.
- `migrate/migrate.go:467-509` — `buildHumanConfig` / `buildToolConfig` (human config is the split pattern reference).
- `migrate/parity.go:257-266` — current `compareToolConfigs` (only compares `Command`).
- `lsp/completion.go:42-65` — `fieldCompletions` shape.
- `lsp/hover.go:77-106` — `formatNodeConfig`, `formatAgentHover` (the multi-line helper this plan copies for tool).
- `CLAUDE.md` — complexity budgets, "Test fixtures should match real parser output" (test design).
- `docs/superpowers/plans/2026-04-22-dip28-tool-safety-defaults.md` — closest past plan; same shape of work, smaller scope.

**CLAUDE.md non-negotiables:**
- Cyclomatic complexity ≤ 5 per function, cognitive complexity ≤ 7. Extract helpers; don't `//nolint`.
- Never run raw `go test`; always use `just <recipe>`.
- Parser tests parse real `.dip` text. Never hand-construct `ir.ToolConfig{...}` to assert field values.
- After changes, `just check` must pass (build, vet, fmt, test-race, releasecheck, complexity, validate-examples, pack-examples, tree-sitter-test).

---

## Design decisions (locked in; don't re-litigate)

| Decision | Value | Why |
|---|---|---|
| Field names | `marker_grep`, `route_required`, `output_limit` | Match the runtime's attr keys verbatim — the runtime reads `n.Attrs["marker_grep"]` etc. |
| Bool parsing | Strict `val == "true"` | Matches existing `goal_gate`/`auto_status`/`cache_tools`; #43 normalizes later. |
| `output_limit` units | Raw bytes (int) | Spec defers SI suffixes to a future issue. |
| Zero/empty handling | Parse stores; formatter / DOT omit when default | Matches existing convention. Authors writing `output_limit: 0` see it disappear on `dippin fmt`; documented in `docs/nodes.md`. |
| Regex validation | None at parse time | Pass-through — the runtime owns these semantics (DIP28 precedent). |
| Field ordering in formatter | `outputs → marker_grep → route_required → output_limit → timeout → reads → writes → command` | Routing-shape fields cluster, then execution, then IO, then command. |
| Migrate parity | Extend to all `ToolConfig` fields | Backfills pre-existing `Timeout`/`Outputs` gap while in the area. |
| Runtime repo | Out of scope | Separate follow-up PR pins `dippin-lang@v0.28.0` and updates `extractToolAttrs`. |

---

## File map

**Modify:**
- `ir/ir.go` (Task 1)
- `parser/parse_nodes.go` (Tasks 2–5)
- `parser/testdata/all_features.dip` (Task 6)
- `formatter/format.go` (Task 7)
- `export/dot.go` (Task 8)
- `migrate/migrate.go` (Task 9)
- `migrate/parity.go` (Task 10)
- `lsp/completion.go` (Task 11)
- `lsp/hover.go` (Task 12)
- `editors/vscode/syntaxes/dippin.tmLanguage.json` (Task 13)
- `docs/nodes.md`, `docs/llm-reference.md`, `docs/context.md`, `docs/edges.md`, `docs/validation.md` (Tasks 15–16)
- `site/static/skill.md`, `site/content/language.md` (Task 17)
- `CHANGELOG.md` (Task 19)

**Create:**
- `parser/parse_tool_test.go` (new — keeps tool-specific tests separate; Tasks 3–5)
- `examples/marker_routing.dip` (Task 14)
- `site/content/blog/whats-new-v028.md` (Task 18)

---

## Task 1: IR — extend `ToolConfig`

**Files:**
- Modify: `ir/ir.go:127-134`

- [ ] **Step 1: Extend ToolConfig**

```go
// ToolConfig holds configuration for shell command nodes.
type ToolConfig struct {
	Command       string // Shell command (multiline OK)
	Timeout       time.Duration
	Outputs       []string // Declared possible stdout values for coverage analysis
	MarkerGrep    string   // Regex matched line-by-line against stdout; populates ctx.tool_marker
	RouteRequired bool     // True → node fails if the command emits no routing signal recognized by the runtime
	OutputLimit   int      // Bytes; > 0 = override engine default
}
```

- [ ] **Step 2: Verify build**

Run: `just build`
Expected: clean build, no errors.

- [ ] **Step 3: Run full test suite (sanity)**

Run: `just test`
Expected: all pass.

- [ ] **Step 4: Commit**

```bash
git add ir/ir.go
git commit -m "ir: add MarkerGrep/RouteRequired/OutputLimit to ToolConfig

Wires three new exported fields onto tool-node config to support
the runtime's stdout-routing primitives. Parser/formatter/DOT updates
land in follow-on commits.

Refs #39"
```

---

## Task 2: Parser — split `applyToolField` into string/bool/parsed helpers

This refactor is required *before* adding cases — current `applyToolField` is at cyclomatic 5; the pre-commit hook will fail otherwise. Pure refactor: behavior unchanged.

**Files:**
- Modify: `parser/parse_nodes.go:405-417`

- [ ] **Step 1: Run existing parser tests to establish baseline**

Run: `just test-pkg parser`
Expected: all pass (will re-run after refactor; output count should be identical).

- [ ] **Step 2: Replace `applyToolField` with split helpers**

Replace lines 405-417 with:

```go
// applyToolField applies tool-specific configuration fields.
func (p *Parser) applyToolField(cfg *ir.ToolConfig, key, val string, loc ir.SourceLocation) {
	if applyToolStringField(cfg, key, val) {
		return
	}
	if applyToolBoolField(cfg, key, val) {
		return
	}
	if p.applyToolParsedField(cfg, key, val, loc) {
		return
	}
	p.emitUnknownFieldHint("tool", key, loc)
}

// applyToolStringField handles string-valued tool fields. Returns true if handled.
func applyToolStringField(cfg *ir.ToolConfig, key, val string) bool {
	switch key {
	case "command":
		cfg.Command = val
	case "outputs":
		cfg.Outputs = splitComma(val)
	default:
		return false
	}
	return true
}

// applyToolBoolField handles boolean tool fields. Returns true if handled.
func applyToolBoolField(cfg *ir.ToolConfig, key, val string) bool {
	switch key {
	default:
		_ = val // no bool fields yet; route_required lands in a follow-on task
		return false
	}
}

// applyToolParsedField handles tool fields needing parsing. Returns true if handled.
func (p *Parser) applyToolParsedField(cfg *ir.ToolConfig, key, val string, loc ir.SourceLocation) bool {
	switch key {
	case "timeout":
		cfg.Timeout = p.parseDuration(val, key, loc)
	default:
		return false
	}
	return true
}
```

- [ ] **Step 3: Run parser tests (must still pass — refactor is behavior-preserving)**

Run: `just test-pkg parser`
Expected: all pass, same count as Step 1.

- [ ] **Step 4: Run complexity check**

Run: `just complexity`
Expected: pass (each new helper has cyclo ≤ 2).

- [ ] **Step 5: Commit**

```bash
git add parser/parse_nodes.go
git commit -m "parser: split applyToolField into string/bool/parsed helpers

Mirrors the existing applyAgent{String,Bool,Parsed}Field shape.
Behavior unchanged; this prepares the function for the three new
routing fields without breaching the cyclomatic-5 budget.

Refs #39"
```

---

## Task 3: Parser — `marker_grep`

**Files:**
- Modify: `parser/parse_nodes.go` (`applyToolStringField`)
- Create: `parser/parse_tool_test.go`

- [ ] **Step 1: Write the failing test**

Create `parser/parse_tool_test.go`:

```go
package parser

import (
	"strings"
	"testing"

	"github.com/2389-research/dippin-lang/ir"
)

// parseToolFixture parses a minimal workflow with a single tool node and returns
// the parsed ToolConfig plus any diagnostics. Tests must parse real .dip text —
// hand-building ir.ToolConfig bypasses the parser and masks bugs (per CLAUDE.md).
func parseToolFixture(t *testing.T, body string) (ir.ToolConfig, []string) {
	t.Helper()
	src := "workflow W\n  goal: \"x\"\n  start: T\n  exit: T\n\n  tool T\n" + body + "\n\n  edges\n"
	p := NewParser(src, "test.dip")
	wf, _ := p.Parse()
	if len(wf.Nodes) == 0 {
		t.Fatalf("no nodes parsed; diagnostics: %v", p.Diagnostics())
	}
	cfg, ok := wf.Nodes[0].Config.(ir.ToolConfig)
	if !ok {
		t.Fatalf("node 0 is not a tool: %T", wf.Nodes[0].Config)
	}
	return cfg, p.Diagnostics()
}

func TestParseToolMarkerGrep(t *testing.T) {
	cfg, diags := parseToolFixture(t, "    marker_grep: '^(green|red)$'")
	if len(diags) != 0 {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if cfg.MarkerGrep != "^(green|red)$" {
		t.Errorf("MarkerGrep = %q, want %q", cfg.MarkerGrep, "^(green|red)$")
	}
}
```

- [ ] **Step 2: Run test (must fail with unrecognized field)**

Run: `just test-pkg parser`
Expected: FAIL — diagnostic "unrecognized tool field \"marker_grep\""

- [ ] **Step 3: Add marker_grep to `applyToolStringField`**

In `parser/parse_nodes.go`, extend `applyToolStringField`:

```go
func applyToolStringField(cfg *ir.ToolConfig, key, val string) bool {
	switch key {
	case "command":
		cfg.Command = val
	case "outputs":
		cfg.Outputs = splitComma(val)
	case "marker_grep":
		cfg.MarkerGrep = val
	default:
		return false
	}
	return true
}
```

- [ ] **Step 4: Run test (must pass)**

Run: `just test-pkg parser`
Expected: PASS.

- [ ] **Step 5: Add tests for unquoted, empty, and regex special chars**

Append to `parser/parse_tool_test.go`:

```go
func TestParseToolMarkerGrepUnquoted(t *testing.T) {
	cfg, diags := parseToolFixture(t, "    marker_grep: tests_pass")
	if len(diags) != 0 {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if cfg.MarkerGrep != "tests_pass" {
		t.Errorf("MarkerGrep = %q, want %q", cfg.MarkerGrep, "tests_pass")
	}
}

func TestParseToolMarkerGrepRegexMetachars(t *testing.T) {
	// Stored verbatim — dippin does not validate the regex (the runtime does).
	cfg, diags := parseToolFixture(t, `    marker_grep: '^\[(PASS|FAIL)\]\s+\d+$'`)
	if len(diags) != 0 {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if !strings.Contains(cfg.MarkerGrep, `\[`) {
		t.Errorf("MarkerGrep lost metachars: %q", cfg.MarkerGrep)
	}
}
```

- [ ] **Step 6: Run tests**

Run: `just test-pkg parser`
Expected: all pass.

- [ ] **Step 7: Commit**

```bash
git add parser/parse_nodes.go parser/parse_tool_test.go
git commit -m "parser: accept marker_grep on tool nodes

Stores the regex string verbatim; the runtime validates and applies it
at runtime. Tests cover quoted/unquoted forms and regex metachars.

Refs #39"
```

---

## Task 4: Parser — `route_required`

**Files:**
- Modify: `parser/parse_nodes.go` (`applyToolBoolField`)
- Modify: `parser/parse_tool_test.go`

- [ ] **Step 1: Write failing test**

Append to `parser/parse_tool_test.go`:

```go
func TestParseToolRouteRequired(t *testing.T) {
	cfg, diags := parseToolFixture(t, "    route_required: true")
	if len(diags) != 0 {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if !cfg.RouteRequired {
		t.Error("RouteRequired = false, want true")
	}
}

func TestParseToolRouteRequiredExplicitFalse(t *testing.T) {
	cfg, diags := parseToolFixture(t, "    route_required: false")
	if len(diags) != 0 {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if cfg.RouteRequired {
		t.Error("RouteRequired = true, want false")
	}
}
```

- [ ] **Step 2: Run tests (must fail)**

Run: `just test-pkg parser`
Expected: FAIL — unrecognized field.

- [ ] **Step 3: Add route_required to `applyToolBoolField`**

In `parser/parse_nodes.go`:

```go
func applyToolBoolField(cfg *ir.ToolConfig, key, val string) bool {
	switch key {
	case "route_required":
		cfg.RouteRequired = (val == "true")
	default:
		return false
	}
	return true
}
```

- [ ] **Step 4: Run tests**

Run: `just test-pkg parser`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add parser/parse_nodes.go parser/parse_tool_test.go
git commit -m "parser: accept route_required on tool nodes

Strict bool parsing matches existing goal_gate/auto_status/cache_tools
convention (#43 will normalize all four together later).

Refs #39"
```

---

## Task 5: Parser — `output_limit`

**Files:**
- Modify: `parser/parse_nodes.go` (`applyToolParsedField`)
- Modify: `parser/parse_tool_test.go`

- [ ] **Step 1: Write failing tests**

Append to `parser/parse_tool_test.go`:

```go
func TestParseToolOutputLimit(t *testing.T) {
	cfg, diags := parseToolFixture(t, "    output_limit: 1048576")
	if len(diags) != 0 {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if cfg.OutputLimit != 1048576 {
		t.Errorf("OutputLimit = %d, want 1048576", cfg.OutputLimit)
	}
}

func TestParseToolOutputLimitNegative(t *testing.T) {
	_, diags := parseToolFixture(t, "    output_limit: -1")
	if len(diags) == 0 {
		t.Fatal("expected diagnostic for negative output_limit, got none")
	}
	found := false
	for _, d := range diags {
		if strings.Contains(d, "output_limit") && strings.Contains(d, "non-negative") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected 'non-negative' diagnostic, got: %v", diags)
	}
}

func TestParseToolOutputLimitNonNumeric(t *testing.T) {
	cfg, diags := parseToolFixture(t, "    output_limit: abc")
	if len(diags) == 0 {
		t.Fatal("expected diagnostic from parseInt, got none")
	}
	if cfg.OutputLimit != 0 {
		t.Errorf("OutputLimit should default to 0 on parse error, got %d", cfg.OutputLimit)
	}
}

func TestParseToolAllRoutingFields(t *testing.T) {
	body := "    marker_grep: pass\n    route_required: true\n    output_limit: 65536"
	cfg, diags := parseToolFixture(t, body)
	if len(diags) != 0 {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if cfg.MarkerGrep != "pass" || !cfg.RouteRequired || cfg.OutputLimit != 65536 {
		t.Errorf("got %+v", cfg)
	}
}
```

- [ ] **Step 2: Run tests (must fail)**

Run: `just test-pkg parser`
Expected: FAIL — unrecognized field.

- [ ] **Step 3: Add output_limit to `applyToolParsedField`**

In `parser/parse_nodes.go`:

```go
func (p *Parser) applyToolParsedField(cfg *ir.ToolConfig, key, val string, loc ir.SourceLocation) bool {
	switch key {
	case "timeout":
		cfg.Timeout = p.parseDuration(val, key, loc)
	case "output_limit":
		n := p.parseInt(val, key, loc)
		if n < 0 {
			p.diagnostics = append(p.diagnostics, fmt.Sprintf(
				"invalid output_limit %d at %d:%d (must be non-negative)", n, loc.Line, loc.Column))
			return true
		}
		cfg.OutputLimit = n
	default:
		return false
	}
	return true
}
```

- [ ] **Step 4: Run tests**

Run: `just test-pkg parser`
Expected: PASS.

- [ ] **Step 5: Run complexity check**

Run: `just complexity`
Expected: pass.

- [ ] **Step 6: Commit**

```bash
git add parser/parse_nodes.go parser/parse_tool_test.go
git commit -m "parser: accept output_limit on tool nodes

Non-negative integer (bytes); 0 is engine default. Negative emits
a diagnostic. Combined-fields test confirms the three new fields
parse together.

Refs #39"
```

---

## Task 6: Extend `all_features.dip` for round-trip coverage

This drives the formatter round-trip test in the next task.

**Files:**
- Modify: `parser/testdata/all_features.dip`

- [ ] **Step 1: Locate the tool node in the fixture**

Run: `grep -n '^  tool ' parser/testdata/all_features.dip`
Read the tool node's current shape.

- [ ] **Step 2: Add the three new fields to one tool node**

Edit the existing tool node block to include:

```dippin
    marker_grep: '^(pass|fail)$'
    route_required: true
    output_limit: 65536
```

(Insert after `timeout:`, before `command:` per the canonical formatter order — Task 7 will reorder this naturally.)

- [ ] **Step 3: Run round-trip test (will fail until formatter is updated in Task 7)**

Run: `just test-pkg parser`
Expected: any round-trip-style test currently using this fixture may fail; that's expected and resolved in Task 7. Do not commit yet.

If no parser-side test uses this fixture for round-trip, just confirm parse succeeds with no diagnostics:

Run: `just test-pkg parser`
Expected: pass.

- [ ] **Step 4: Do not commit yet** — Task 7 lands the formatter change in the same commit so the fixture stays consistent.

---

## Task 7: Formatter — emit new fields, fix ordering

**Files:**
- Modify: `formatter/format.go:488-500`

- [ ] **Step 1: Write failing round-trip test (or use existing one if it covers all_features.dip)**

Check: `grep -n 'all_features' formatter/format_test.go`
If there's a round-trip test that loads `all_features.dip`, it will already fail after Task 6. Otherwise, add:

```go
func TestFormatToolRoutingFieldsRoundTrip(t *testing.T) {
	src := `workflow W
  goal: "x"
  start: T
  exit: T

  tool T
    outputs: pass, fail
    marker_grep: '^(pass|fail)$'
    route_required: true
    output_limit: 65536
    timeout: 30s
    command:
      echo pass

  edges
`
	wf1, _ := parser.NewParser(src, "test.dip").Parse()
	out := Format(wf1)
	wf2, _ := parser.NewParser(out, "test.dip").Parse()
	out2 := Format(wf2)
	if out != out2 {
		t.Errorf("not idempotent:\nfirst:\n%s\nsecond:\n%s", out, out2)
	}
}
```

- [ ] **Step 2: Run test (must fail)**

Run: `just test-pkg formatter`
Expected: FAIL — fields not emitted.

- [ ] **Step 3: Update `writeToolFields`**

Replace:

```go
func writeToolFields(wr *writer, n *ir.Node, cfg ir.ToolConfig) {
	writeCommonNodeFields(wr, n)
	if len(cfg.Outputs) > 0 {
		wr.line("outputs: %s", strings.Join(cfg.Outputs, ", "))
	}
	if cfg.MarkerGrep != "" {
		wr.line("marker_grep: %s", quoteValue(cfg.MarkerGrep))
	}
	if cfg.RouteRequired {
		wr.line("route_required: true")
	}
	if cfg.OutputLimit > 0 {
		wr.line("output_limit: %d", cfg.OutputLimit)
	}
	if cfg.Timeout != 0 {
		wr.line("timeout: %s", formatDuration(cfg.Timeout))
	}
	writeIOFields(wr, n)
	if cfg.Command != "" {
		wr.multilineBlock("command", cfg.Command)
	}
}
```

- [ ] **Step 4: Run tests**

Run: `just test-pkg formatter`
Expected: PASS.

- [ ] **Step 5: Run complexity check**

Run: `just complexity`
Expected: pass (`writeToolFields` is now ~cyclo 7-ish; if it fails, extract a `writeToolRoutingFields` helper).

If complexity fails, extract:

```go
func writeToolRoutingFields(wr *writer, cfg ir.ToolConfig) {
	if cfg.MarkerGrep != "" {
		wr.line("marker_grep: %s", quoteValue(cfg.MarkerGrep))
	}
	if cfg.RouteRequired {
		wr.line("route_required: true")
	}
	if cfg.OutputLimit > 0 {
		wr.line("output_limit: %d", cfg.OutputLimit)
	}
}
```

Then call it from `writeToolFields` after the `outputs:` line.

- [ ] **Step 6: Run full test suite (including parser tests that depend on all_features.dip)**

Run: `just test`
Expected: all pass.

- [ ] **Step 7: Commit**

```bash
git add formatter/format.go formatter/format_test.go parser/testdata/all_features.dip
git commit -m "formatter: emit marker_grep/route_required/output_limit

Emits fields when non-default in canonical order: outputs →
marker_grep → route_required → output_limit → timeout → reads →
writes → command. Zero/empty/false omitted; round-trip stable.
all_features.dip fixture extended to drive idempotency check.

Refs #39"
```

---

## Task 8: DOT export — emit semantic attrs

**Files:**
- Modify: `export/dot.go:303-308`

- [ ] **Step 1: Write failing test**

Append to `export/dot_test.go`:

```go
func TestExportDOTToolRoutingFields(t *testing.T) {
	src := `workflow W
  goal: "x"
  start: T
  exit: T

  tool T
    marker_grep: '^pass$'
    route_required: true
    output_limit: 8192
    timeout: 30s
    command:
      echo pass

  edges
`
	wf, _ := parser.NewParser(src, "test.dip").Parse()
	out, err := ExportDOT(wf, ExportOptions{})
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		`marker_grep="^pass$"`,
		`route_required="true"`,
		`output_limit="8192"`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("DOT output missing %q. Full output:\n%s", want, out)
		}
	}
}

func TestExportDOTToolRoutingOmitWhenZero(t *testing.T) {
	src := `workflow W
  goal: "x"
  start: T
  exit: T

  tool T
    timeout: 30s
    command:
      echo pass

  edges
`
	wf, _ := parser.NewParser(src, "test.dip").Parse()
	out, _ := ExportDOT(wf, ExportOptions{})
	for _, unwanted := range []string{"marker_grep", "route_required", "output_limit"} {
		if strings.Contains(out, unwanted) {
			t.Errorf("DOT output unexpectedly contains %q. Full output:\n%s", unwanted, out)
		}
	}
}
```

- [ ] **Step 2: Run tests (must fail)**

Run: `just test-pkg export`
Expected: FAIL.

- [ ] **Step 3: Update `applyToolSemanticAttrs`**

Replace lines 303-308:

```go
// applyToolSemanticAttrs adds tool runtime attrs (always exported).
func applyToolSemanticAttrs(attrs map[string]string, cfg ir.ToolConfig) {
	if cfg.Timeout != 0 {
		attrs["timeout"] = formatDuration(cfg.Timeout)
	}
	if cfg.MarkerGrep != "" {
		attrs["marker_grep"] = cfg.MarkerGrep
	}
	if cfg.RouteRequired {
		attrs["route_required"] = "true"
	}
	if cfg.OutputLimit > 0 {
		attrs["output_limit"] = strconv.Itoa(cfg.OutputLimit)
	}
}
```

Add `"strconv"` to imports if not already present.

- [ ] **Step 4: Run tests**

Run: `just test-pkg export`
Expected: PASS.

- [ ] **Step 5: Run complexity check**

Run: `just complexity`
Expected: pass (cyclo ~5; if it fails, extract).

- [ ] **Step 6: Commit**

```bash
git add export/dot.go export/dot_test.go
git commit -m "export: emit tool routing attrs in DOT semantic export

Adds marker_grep / route_required / output_limit to applyToolSemanticAttrs
with the same emission contract as the formatter (non-default only).
The runtime reads these via n.Attrs[...].

Refs #39"
```

---

## Task 9: Migrate — read attrs back

**Files:**
- Modify: `migrate/migrate.go:496-509`

- [ ] **Step 1: Write failing test**

Append to `migrate/migrate_test.go`:

```go
func TestBuildToolConfigRoutingAttrs(t *testing.T) {
	attrs := map[string]string{
		"tool_command":   "echo hi",
		"timeout":        "30s",
		"marker_grep":    "^pass$",
		"route_required": "true",
		"output_limit":   "8192",
	}
	cfg, err := buildToolConfig(attrs)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.MarkerGrep != "^pass$" {
		t.Errorf("MarkerGrep = %q", cfg.MarkerGrep)
	}
	if !cfg.RouteRequired {
		t.Error("RouteRequired = false")
	}
	if cfg.OutputLimit != 8192 {
		t.Errorf("OutputLimit = %d", cfg.OutputLimit)
	}
}

func TestBuildToolConfigOutputLimitInvalid(t *testing.T) {
	attrs := map[string]string{"output_limit": "not_a_number"}
	_, err := buildToolConfig(attrs)
	if err == nil {
		t.Error("expected error for non-numeric output_limit, got nil")
	}
}
```

- [ ] **Step 2: Run test (must fail)**

Run: `just test-pkg migrate`
Expected: FAIL.

- [ ] **Step 3: Refactor `buildToolConfig` and add helpers**

Replace lines 496-509:

```go
func buildToolConfig(attrs map[string]string) (ir.ToolConfig, error) {
	cfg := ir.ToolConfig{}
	if v, ok := attrs["tool_command"]; ok {
		cfg.Command = v
	}
	if v, ok := attrs["marker_grep"]; ok {
		cfg.MarkerGrep = v
	}
	if v, ok := attrs["route_required"]; ok {
		cfg.RouteRequired = (v == "true")
	}
	if err := applyToolTimeoutAttr(&cfg, attrs); err != nil {
		return cfg, err
	}
	if err := applyToolOutputLimitAttr(&cfg, attrs); err != nil {
		return cfg, err
	}
	return cfg, nil
}

func applyToolTimeoutAttr(cfg *ir.ToolConfig, attrs map[string]string) error {
	v, ok := attrs["timeout"]
	if !ok {
		return nil
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return fmt.Errorf("invalid timeout %q: %w", v, err)
	}
	cfg.Timeout = d
	return nil
}

func applyToolOutputLimitAttr(cfg *ir.ToolConfig, attrs map[string]string) error {
	v, ok := attrs["output_limit"]
	if !ok {
		return nil
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fmt.Errorf("invalid output_limit %q: %w", v, err)
	}
	cfg.OutputLimit = n
	return nil
}
```

Add `"strconv"` to imports.

- [ ] **Step 4: Run tests**

Run: `just test-pkg migrate`
Expected: PASS.

- [ ] **Step 5: Complexity check**

Run: `just complexity`
Expected: pass.

- [ ] **Step 6: Commit**

```bash
git add migrate/migrate.go migrate/migrate_test.go
git commit -m "migrate: read marker_grep/route_required/output_limit from DOT

Splits buildToolConfig to keep cyclo budget, extracts timeout and
output_limit attr application into helpers.

Refs #39"
```

---

## Task 10: Migrate parity — extend `compareToolConfigs`

**Files:**
- Modify: `migrate/parity.go:257-266`

- [ ] **Step 1: Write failing tests**

Append to `migrate/parity_test.go` (or wherever `TestCompareToolConfigs_DifferentCommand` lives — `grep -rn 'compareToolConfigs\|TestCompareToolConfigs' migrate/`):

```go
func TestCompareToolConfigsDifferentMarkerGrep(t *testing.T) {
	a := ir.ToolConfig{MarkerGrep: "^pass$"}
	b := ir.ToolConfig{MarkerGrep: "^fail$"}
	diffs := compareToolConfigs("T", "", a, b)
	if len(diffs) == 0 {
		t.Error("expected difference for MarkerGrep, got none")
	}
}

func TestCompareToolConfigsDifferentRouteRequired(t *testing.T) {
	a := ir.ToolConfig{RouteRequired: true}
	b := ir.ToolConfig{RouteRequired: false}
	diffs := compareToolConfigs("T", "", a, b)
	if len(diffs) == 0 {
		t.Error("expected difference for RouteRequired, got none")
	}
}

func TestCompareToolConfigsDifferentOutputLimit(t *testing.T) {
	a := ir.ToolConfig{OutputLimit: 1024}
	b := ir.ToolConfig{OutputLimit: 2048}
	diffs := compareToolConfigs("T", "", a, b)
	if len(diffs) == 0 {
		t.Error("expected difference for OutputLimit, got none")
	}
}

func TestCompareToolConfigsDifferentTimeout(t *testing.T) {
	a := ir.ToolConfig{Timeout: 30 * time.Second}
	b := ir.ToolConfig{Timeout: 60 * time.Second}
	diffs := compareToolConfigs("T", "", a, b)
	if len(diffs) == 0 {
		t.Error("expected difference for Timeout, got none")
	}
}

func TestCompareToolConfigsDifferentOutputs(t *testing.T) {
	a := ir.ToolConfig{Outputs: []string{"pass", "fail"}}
	b := ir.ToolConfig{Outputs: []string{"green", "red"}}
	diffs := compareToolConfigs("T", "", a, b)
	if len(diffs) == 0 {
		t.Error("expected difference for Outputs, got none")
	}
}
```

- [ ] **Step 2: Run tests (must fail)**

Run: `just test-pkg migrate`
Expected: most fail (only `MarkerGrep`/`RouteRequired`/`OutputLimit` if Command happens to be empty in both? — check; expect at minimum 3-5 failures).

- [ ] **Step 3: Replace `compareToolConfigs` with helper-extracted version**

```go
func compareToolConfigs(id, path string, ac ir.ToolConfig, bCfg interface{}) []Difference {
	bc, ok := bCfg.(ir.ToolConfig)
	if !ok {
		return []Difference{configMismatchDiff(id, path, "ToolConfig", bCfg)}
	}
	var diffs []Difference
	diffs = append(diffs, compareToolScalars(id, ac, bc)...)
	diffs = append(diffs, compareToolSlices(id, ac, bc)...)
	return diffs
}

func compareToolScalars(id string, ac, bc ir.ToolConfig) []Difference {
	var diffs []Difference
	if !promptsEqual(ac.Command, bc.Command) {
		diffs = append(diffs, fieldDiff(id, "command", fmt.Sprintf("node %q command differs", id)))
	}
	if ac.Timeout != bc.Timeout {
		diffs = append(diffs, fieldDiff(id, "timeout", fmt.Sprintf("node %q timeout: %s vs %s", id, ac.Timeout, bc.Timeout)))
	}
	if ac.MarkerGrep != bc.MarkerGrep {
		diffs = append(diffs, fieldDiff(id, "marker_grep", fmt.Sprintf("node %q marker_grep: %q vs %q", id, ac.MarkerGrep, bc.MarkerGrep)))
	}
	if ac.RouteRequired != bc.RouteRequired {
		diffs = append(diffs, fieldDiff(id, "route_required", fmt.Sprintf("node %q route_required: %v vs %v", id, ac.RouteRequired, bc.RouteRequired)))
	}
	if ac.OutputLimit != bc.OutputLimit {
		diffs = append(diffs, fieldDiff(id, "output_limit", fmt.Sprintf("node %q output_limit: %d vs %d", id, ac.OutputLimit, bc.OutputLimit)))
	}
	return diffs
}

func compareToolSlices(id string, ac, bc ir.ToolConfig) []Difference {
	var diffs []Difference
	if strings.Join(ac.Outputs, ",") != strings.Join(bc.Outputs, ",") {
		diffs = append(diffs, fieldDiff(id, "outputs", fmt.Sprintf("node %q outputs: %v vs %v", id, ac.Outputs, bc.Outputs)))
	}
	return diffs
}
```

Add `"strings"` to imports if not already present.

- [ ] **Step 4: Run tests**

Run: `just test-pkg migrate`
Expected: PASS (existing `_DifferentCommand` test still passes; new tests pass).

- [ ] **Step 5: Complexity check**

Run: `just complexity`
Expected: pass.

- [ ] **Step 6: Commit**

```bash
git add migrate/parity.go migrate/parity_test.go
git commit -m "migrate: parity compares all ToolConfig fields

Extends compareToolConfigs to detect drift in Timeout, Outputs,
MarkerGrep, RouteRequired, OutputLimit. Previously only Command
was compared; Timeout/Outputs gaps backfilled while in the area.

Refs #39"
```

---

## Task 11: LSP completion — three new entries

**Files:**
- Modify: `lsp/completion.go:42-65`

- [ ] **Step 1: Write failing test (if there's a completion test fixture)**

Check: `grep -rn 'fieldCompletions\|TestCompletion' lsp/`

If a test asserts the completion list, add to it:

```go
func TestFieldCompletionsIncludesRoutingFields(t *testing.T) {
	items := fieldCompletions()
	want := map[string]bool{"marker_grep:": false, "route_required:": false, "output_limit:": false}
	for _, it := range items {
		if _, ok := want[it.Label]; ok {
			want[it.Label] = true
		}
	}
	for label, found := range want {
		if !found {
			t.Errorf("missing completion for %q", label)
		}
	}
}
```

- [ ] **Step 2: Run test (must fail)**

Run: `just test-pkg lsp`
Expected: FAIL.

- [ ] **Step 3: Add three entries**

In `lsp/completion.go`, extend the `fields` slice in `fieldCompletions`:

```go
{"marker_grep:", "Regex matched against tool stdout; sets ctx.tool_marker"},
{"route_required:", "Fail the node if the command emits no runtime-recognized routing signal on stdout"},
{"output_limit:", "Per-node stdout byte cap (positive int)"},
```

- [ ] **Step 4: Run tests**

Run: `just test-pkg lsp`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add lsp/completion.go lsp/completion_test.go
git commit -m "lsp: complete marker_grep/route_required/output_limit

Refs #39"
```

---

## Task 12: LSP hover — extract `formatToolHover`

**Files:**
- Modify: `lsp/hover.go:77-106`

- [ ] **Step 1: Write failing test**

Append to wherever hover is tested (`grep -rn 'formatNodeConfig\|TestHover' lsp/`):

```go
func TestFormatToolHoverWithRoutingFields(t *testing.T) {
	node := &ir.Node{
		Kind: ir.NodeTool,
		Config: ir.ToolConfig{
			Command:       "echo hi",
			MarkerGrep:    "^pass$",
			RouteRequired: true,
			OutputLimit:   8192,
		},
	}
	out := formatNodeConfig(node, &ir.Workflow{})
	for _, want := range []string{"marker_grep", "route_required", "output_limit"} {
		if !strings.Contains(out, want) {
			t.Errorf("hover missing %q. Output:\n%s", want, out)
		}
	}
}
```

- [ ] **Step 2: Run test (must fail)**

Run: `just test-pkg lsp`
Expected: FAIL.

- [ ] **Step 3: Replace the `ToolConfig` branch with a helper**

In `lsp/hover.go`, replace the case:

```go
case ir.ToolConfig:
    return formatToolHover(cfg)
```

Then add `formatToolHover` mirroring `formatAgentHover`:

```go
// formatToolHover formats tool-specific hover info.
func formatToolHover(cfg ir.ToolConfig) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Command: `%s`\n", truncateStr(cfg.Command, 80))
	if cfg.MarkerGrep != "" {
		fmt.Fprintf(&b, "marker_grep: `%s`\n", cfg.MarkerGrep)
	}
	if cfg.RouteRequired {
		b.WriteString("route_required: true\n")
	}
	if cfg.OutputLimit > 0 {
		fmt.Fprintf(&b, "output_limit: %d bytes\n", cfg.OutputLimit)
	}
	return b.String()
}
```

- [ ] **Step 4: Run tests**

Run: `just test-pkg lsp`
Expected: PASS.

- [ ] **Step 5: Complexity check**

Run: `just complexity`
Expected: pass.

- [ ] **Step 6: Commit**

```bash
git add lsp/hover.go lsp/hover_test.go
git commit -m "lsp: surface routing fields in tool hover

Extracts formatToolHover mirroring formatAgentHover and adds
conditional lines for MarkerGrep, RouteRequired, OutputLimit.

Refs #39"
```

---

## Task 13: VSCode tmLanguage keyword

**Files:**
- Modify: `editors/vscode/syntaxes/dippin.tmLanguage.json:148`

- [ ] **Step 1: Locate the field-name regex**

Run: `grep -n 'goal_gate\|auto_status' editors/vscode/syntaxes/dippin.tmLanguage.json`
Find the alternation list at line 148.

- [ ] **Step 2: Append three keywords**

Insert `marker_grep|route_required|output_limit` into the existing alternation (alphabetical position is fine but match neighboring style).

- [ ] **Step 3: Run tree-sitter check (also exercises the grammar repo)**

Run: `just check`
Expected: pass (the tmLanguage is JSON, not a grammar; no separate test, but `just check` validates examples and full build).

- [ ] **Step 4: Commit**

```bash
git add editors/vscode/syntaxes/dippin.tmLanguage.json
git commit -m "editors(vscode): highlight new tool routing field names

Refs #39"
```

---

## Task 14: Create `examples/marker_routing.dip`

**Files:**
- Create: `examples/marker_routing.dip`

- [ ] **Step 1: Create the example**

Model after `examples/tool_safety.dip`. Must be lint-clean (DIP111 timeout, no model warnings, etc.):

```dippin
workflow MarkerRouting
  goal: "Show typed marker-driven routing on tool nodes"
  start: RunTests
  exit: Done

  defaults
    model: claude-sonnet-4-6
    provider: anthropic

  tool RunTests
    label: "Run test suite, route by marker"
    outputs: tests_green, tests_red
    marker_grep: '^(tests_green|tests_red)$'
    route_required: true
    output_limit: 65536
    timeout: 2m
    command:
      #!/bin/sh
      set -eu
      if go test ./... > /tmp/test.log 2>&1; then
        printf 'tests_green\n'
      else
        printf 'tests_red\n'
      fi

  agent Done
    prompt: "Summarize the routing outcome."

  edges
    RunTests -> Done when ctx.tool_marker = tests_green
    RunTests -> Done when ctx.tool_marker = tests_red
```

- [ ] **Step 2: Validate**

Run: `./dippin validate examples/marker_routing.dip`
(Or after `just build`: `dippin validate examples/marker_routing.dip`)
Expected: clean.

Run: `./dippin lint examples/marker_routing.dip`
Expected: zero warnings.

- [ ] **Step 3: Run integration test**

Run: `just test-pkg validator`
Expected: `TestLintExamples` and `TestPackExamples` pass with the new example.

- [ ] **Step 4: Commit**

```bash
git add examples/marker_routing.dip
git commit -m "examples: add marker_routing.dip demoing tool routing fields

Shows marker_grep + route_required + output_limit together with
a representative go-test runner pattern. Lint-clean.

Refs #39"
```

---

## Task 15: Docs — `docs/nodes.md`

**Files:**
- Modify: `docs/nodes.md`

- [ ] **Step 1: Add three rows to the Tool-Specific Fields table (line ~261)**

After the `outputs` row:

```markdown
| `marker_grep` | String | — | Regex matched line-by-line against captured stdout. The last match populates `ctx.tool_marker`. The runtime validates and applies the regex at runtime. |
| `route_required` | Boolean | false | When true, the node fails if the command emits no routing signal recognized by the runtime (the runtime defines the routing-signal format). The routing value populates `ctx.tool_route`. |
| `output_limit` | Integer | — | Per-node override for the engine's captured-stdout byte cap. Positive integer; omit (or set 0) to use the engine default. |
```

- [ ] **Step 2: Extend the "Markers and Verbose Output" section (line ~286)**

After the existing prose, append:

```markdown
**Best (when the runtime supports typed markers):** declare `marker_grep` and let the runtime parse the routing signal directly, freeing stdout for diagnostic output:

```dippin
  tool RunTests
    marker_grep: '^(tests_pass|tests_fail)$'
    timeout: 60s
    command:
      pytest 2>&1
      [ $? -eq 0 ] && printf 'tests_pass\n' || printf 'tests_fail\n'
```

When `marker_grep` is declared, the runtime populates `ctx.tool_marker` and routing edges can reference it instead of `ctx.tool_stdout`. `route_required: true` makes the absence of any match a hard failure. `output_limit` overrides the per-node stdout cap when the command genuinely needs a larger window.
```markdown

- [ ] **Step 3: Commit**

```bash
git add docs/nodes.md
git commit -m "docs(nodes): document marker_grep / route_required / output_limit

Refs #39"
```

---

## Task 16: Docs — `docs/llm-reference.md`, reserved ctx vars

**Files:**
- Modify: `docs/llm-reference.md`
- Modify: `docs/context.md`
- Modify: `docs/edges.md`
- Modify: `docs/validation.md`

- [ ] **Step 1: Update llm-reference.md tool row (line ~50)**

Change:
```markdown
| `tool` | `command` | `timeout` (e.g. 30s, 5m) |
```
to:
```markdown
| `tool` | `command` | `timeout` (e.g. 30s, 5m), `outputs` (CSV), `marker_grep` (regex), `route_required` (bool), `output_limit` (bytes) |
```

- [ ] **Step 2: Add a row to the common-mistakes table (after row 8)**

```markdown
| 9 | Hand-parsing tool stdout for routing | Use `marker_grep: '<regex>'` (and optionally `route_required: true`) instead of regexing `ctx.tool_stdout` in edge conditions. Mirrors the runtime's TRK101 recommendation; populates `ctx.tool_marker` directly. |
```

- [ ] **Step 3: Add `ctx.tool_marker` and `ctx.tool_route` to context.md, edges.md, validation.md**

In each file: locate the existing reserved `ctx.*` table or list and add two rows describing the new variables:

```markdown
| `ctx.tool_marker` | Tool stdout regex match (when `marker_grep` is declared) |
| `ctx.tool_route` | A routing value the runtime extracts from the tool's stdout (format defined by the runtime; set when `route_required` is honored) |
```

Match the surrounding style — `grep -n 'ctx\.tool_stdout' docs/context.md docs/edges.md docs/validation.md` to find the exact insertion point in each.

- [ ] **Step 4: Regenerate the embedded spec**

Run: `just gen-spec`
Expected: refreshes `docs/generated-spec.md` and `cmd/dippin/generated-spec.md`.

- [ ] **Step 5: Commit**

```bash
git add docs/ cmd/dippin/generated-spec.md
git commit -m "docs: add tool routing fields and ctx.tool_marker/route to references

Updates llm-reference.md tool row + common-mistakes table; adds
ctx.tool_marker / ctx.tool_route to the reserved variable lists
in context.md, edges.md, validation.md. Regenerates embedded spec.

Refs #39"
```

---

## Task 17: Website — `skill.md` + `language.md`

**Files:**
- Modify: `site/static/skill.md`
- Modify: `site/content/language.md`

- [ ] **Step 1: Update `skill.md` tool field table (line ~117)**

Insert three rows mirroring the docs/nodes.md edits.

- [ ] **Step 2: Update `skill.md` Best Practices section (line ~411)**

After the "Always set timeout" bullet, add:

```markdown
- **Prefer `marker_grep:`** over regexing `ctx.tool_stdout` in edges when the runtime supports it. Typed routing leaves stdout free for diagnostic output and avoids truncation foot-guns.
```

- [ ] **Step 3: Update `skill.md` Context Variables table (line ~462)**

Add:

```markdown
| `ctx.tool_marker` | Tool stdout regex match (when `marker_grep` declared) |
| `ctx.tool_route` | A routing value the runtime extracts from the tool's stdout (format defined by the runtime; set when `route_required` is honored) |
```

- [ ] **Step 4: Update `site/content/language.md` `### tool` section (line ~175)**

Replace the existing tool example with one that shows the new fields too:

```dippin
  tool RunTests
    label: "Run test suite"
    outputs: tests_pass, tests_fail
    marker_grep: '^(tests_pass|tests_fail)$'
    timeout: 60s
    command:
      pytest --tb=short
```

Add a short prose line below the example: "Declare `marker_grep` for typed routing (populates `ctx.tool_marker`); `route_required: true` makes the node fail if the command emits no routing signal recognized by the runtime; `output_limit` overrides the captured-stdout byte cap."

- [ ] **Step 5: Commit**

```bash
git add site/static/skill.md site/content/language.md
git commit -m "site: document tool routing fields in skill and language pages

Adds field table rows, Best Practices bullet, Context Variables
entries to hosted skill.md; updates language.md tool example.

Refs #39"
```

---

## Task 18: Blog post — `whats-new-v028.md`

**Files:**
- Create: `site/content/blog/whats-new-v028.md`

- [ ] **Step 1: Determine the release date**

Pick today (or planned tag day). Use it as both `date` and `publishDate`.

- [ ] **Step 2: Create the post**

Mirror `site/content/blog/whats-new-v027.md` frontmatter shape:

```markdown
---
title: "What's new in v0.28: typed tool routing"
date: "2026-MM-DD"
publishDate: "2026-MM-DD"
draft: false
description: "Tool nodes can now declare marker_grep, route_required, and output_limit — typed routing primitives that close the parity gap with the runtime."
---

## Three new tool-node fields

Issue #39 closes a parser gap: the runtime already supported `marker_grep`, `route_required`, and `output_limit`, but `.dip` source couldn't reach them. Authors following the runtime's `TRK101` lint recommendation hit `unrecognized tool field "marker_grep"` from dippin instead of cleaner routing.

```dippin
  tool RunTests
    marker_grep: '^(tests_pass|tests_fail)$'
    route_required: true
    output_limit: 65536
    timeout: 2m
    command:
      go test ./... 2>&1
```

### What each field does

- **`marker_grep`** — regex matched against stdout. Last match populates `ctx.tool_marker`. Routing edges can reference it instead of regexing `ctx.tool_stdout`.
- **`route_required`** — when true, the node fails (`EventToolRouteMissing`) if the command emits no routing signal recognized by the runtime. The routing value populates `ctx.tool_route`.
- **`output_limit`** — per-node override of the engine's captured-stdout byte cap, for tools that need a larger or smaller window than the global default.

### Why it matters

Without these, authors hit the runtime's TRK101 truncation foot-gun and worked around it by enumerating every possible stdout marker as a conditional edge — which made `dippin coverage` flag false-positive non-exhaustive routing. Typed `marker_grep` is the canonical fix; now dippin accepts it.

### Runtime requirement

These fields forward to the runtime via DOT export. Routing semantics require an enforcing runtime (the version that ships the matching `extractToolAttrs` change). Earlier runtime versions silently ignore the new attrs.

### Coming next

- **#42** — DIP138, lint warning when a tool node parses stdout for routing without declaring `marker_grep` / `outputs`.
- **#43** — normalize boolean parsing (`true` / `yes` / `1` accepted consistently).
- **#44** — close the existing `outputs` DOT round-trip gap.
```markdown

- [ ] **Step 3: Verify Hugo build**

Run: `just site-build` (or whatever the site build command is — `grep -A2 'site-build:' Justfile`)
Expected: clean build; the new post is included.

- [ ] **Step 4: Commit**

```bash
git add site/content/blog/whats-new-v028.md
git commit -m "site(blog): add v0.28 release post for tool routing fields

Refs #39"
```

---

## Task 19: CHANGELOG + version bump

**Files:**
- Modify: `CHANGELOG.md`

- [ ] **Step 1: Add v0.28.0 entry at the top**

Insert before the `## [v0.27.0]` heading:

```markdown
## [v0.28.0] — 2026-MM-DD

Tool-node routing fields land in the parser and IR. Authors following the runtime's `TRK101` recommendation can now declare `marker_grep`, `route_required`, and `output_limit` directly in `.dip` source. Closes #39.

### Added

- `tool.marker_grep` — regex matched line-by-line against captured stdout; populates `ctx.tool_marker` at runtime.
- `tool.route_required` — boolean; when true, the node fails with `EventToolRouteMissing` if the command emits no routing signal recognized by the runtime.
- `tool.output_limit` — positive integer (bytes) overriding the engine default stdout tail-window for this node.
- Reserved context variables: `ctx.tool_marker`, `ctx.tool_route`.

### Changed

- `migrate/parity.go compareToolConfigs` now compares all `ToolConfig` fields. Pre-existing `Timeout` / `Outputs` parity gaps backfilled.

### Runtime requirement

These fields pass through DOT export to the runtime. Routing semantics require the runtime to ship the matching `extractToolAttrs` change; see issue #39 for details.

### Docs

- New blog post: ["What's new in v0.28: typed tool routing"](blog/whats-new-v028.md).
- `docs/nodes.md` gains a "Best (when the runtime supports typed markers)" note in the Markers and Verbose Output section.
- Hosted skill (`site/static/skill.md`) updated with new ctx vars and best-practice bullet.
```

- [ ] **Step 2: Commit**

```bash
git add CHANGELOG.md
git commit -m "changelog: v0.28.0 — tool routing fields

Refs #39"
```

---

## Task 20: Full check, push, tag

- [ ] **Step 1: Run the full gate**

Run: `just check`
Expected: pass. If anything fails, fix it before tagging.

- [ ] **Step 2: Push branch and open PR**

(Per `commit-commands:commit-push-pr` skill, or manually with `gh pr create`.)

- [ ] **Step 3: Merge PR, pull main**

```bash
git checkout main
git pull origin main
```

- [ ] **Step 4: Tag**

```bash
just release v0.28.0 "tool routing fields: marker_grep, route_required, output_limit (#39)"
```

Or manually:
```bash
git tag -a v0.28.0 -m "tool routing fields: marker_grep, route_required, output_limit (#39)"
git push origin v0.28.0
```

- [ ] **Step 5: Verify GoReleaser**

Check `gh run list --workflow=release.yml --limit 3` for the tag-triggered run; confirm it completes and the Homebrew tap updates.

- [ ] **Step 6: Open runtime follow-up PR**

Separately in the runtime repo: bump `dippin-lang` dep to `v0.28.0`, update the runtime's pipeline adapter layer's `extractToolAttrs` to forward the three new fields (sketch in spec out-of-scope section).

---

## Self-review checklist (for the spec author, before handoff)

- [x] Every spec section maps to a task: IR (1), parser (2-5), formatter (7), DOT (8), migrate (9-10), LSP (11-12), VSCode (13), example (14), docs (15-16), website (17), blog (18), changelog (19), release (20).
- [x] No TODO / TBD / "appropriate error handling" placeholders.
- [x] Function names consistent across tasks (`applyToolStringField` used in Task 2 and Task 3; `formatToolHover` used in Task 12).
- [x] Test fixtures parse real `.dip` text per CLAUDE.md.
- [x] Each task commits before the next starts.
- [x] `just check` gate at end before tag.
