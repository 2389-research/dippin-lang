# Pipeline Inputs — Phase 1 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a first-class `inputs` block to the dippin grammar, IR, parser, formatter, and validator, plus a typed JSON introspection surface, so a host can enumerate a pipeline's declared inputs before running it.

**Architecture:** `inputs` is the callee-side signature of a `.dip` file — it declares what a caller must supply, whether that caller is a human at the entry point or a parent workflow via `subgraph … params:`. Declared inputs are read as `${inputs.name}`, a *closed* namespace: references are checked against the declaration, unlike every other namespace in the language. The parser accepts unknown types and attributes without failing; the validator diagnoses them. The IR stores raw source text plus the declared type, and the JSON projection coerces to real JSON types.

**Tech Stack:** Go, `just` for all build/test operations, tree-sitter (JS grammar), VS Code TextMate JSON, Zed `.scm` queries.

**Spec:** `docs/superpowers/specs/2026-08-06-pipeline-inputs-design.md`
**Issue:** dippin-lang #190 (paired: tracker #553, tracker-runner #210)
**Branch:** `feat/190-pipeline-inputs` (worktree at `.claude/worktrees/190-pipeline-inputs`)

## Global Constraints

- **Never run raw `go build` / `go test` / `gocyclo`.** Every operation goes through `just` (`just test`, `just test-pkg <pkg>`, `just check`, `just fmt`). If a needed command has no recipe, add one first.
- **Cyclomatic complexity ≤ 5, cognitive complexity ≤ 7 per function**, enforced by the pre-commit hook. In practice a `switch` may have **at most 4 cases plus `default`**. When a function exceeds budget, extract helpers — never add `//nolint`.
- **Never commit to `main`.** All work stays on `feat/190-pipeline-inputs`.
- **Commit message trailer** on every commit: `Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>`
- **`just check` fails locally on tree-sitter-generate** (the CLI is not installed here). The pre-commit hook is the real gate — it runs build, vet, golangci-lint, gofmt, full tests, complexity, and example validation.
- **Test fixtures are built by parsing real `.dip` source**, never by hand-assembling IR structs. Hand-populated fixtures caused the DIP101 bug by masking that production code never set a field.
- **IR additions are additive and zero-value-safe.** A `.dip` with no `inputs` block must produce `Inputs == nil` and byte-identical formatter output.
- **Attribute names are snake_case** (`max_length`, not `maxLength`).
- **There is no `parser.Parse(src)` function.** The API is `parser.NewParser(input, filename)` → `.Parse() (*ir.Workflow, error)` → `.Diagnostics() []string`. Parse diagnostics are returned by the separate `Diagnostics()` call, **not** by `Parse()`. Each test file below defines a small `parseSrc` helper wrapping this; use it rather than inventing another entry point.
- **`parser/parse_inputs_test.go` is `package parser`** (internal), matching `parser/parser_test.go`. Validator and formatter tests are external (`package validator_test`, `package formatter_test`).

## Canonical orders (referenced by several tasks)

**Workflow section order** (formatter): `header (goal, requires, start, exit)` → `inputs` → `defaults` → `vars` → nodes → `stylesheet` → `edges`

**Input attribute order** (formatter): `required`, `prompt`, `description`, `default`, `options`, `pattern`, `min`, `max`, `max_length`, `multiline`

**Input declaration order**: preserved exactly as authored. The formatter must **not** sort `Inputs` — a host renders them as an ordered form. This deliberately differs from `vars`, which the formatter sorts alphabetically.

---

### Task 1: IR types and minimal-form parsing

Adds `ir.Input`, `Workflow.Inputs`, the lookup helper, and a parser that handles the one-line form (`idea: text`) with no attribute block.

**Files:**
- Modify: `ir/ir.go` (add `Inputs` field to `Workflow`, add `Input` type)
- Modify: `ir/lookup.go` (add `Workflow.Input` method)
- Create: `parser/parse_inputs.go`
- Modify: `parser/parser.go:162-176` (`dispatchWorkflowSimpleField` — add the `inputs` case)
- Test: `parser/parse_inputs_test.go`

**Interfaces:**
- Produces: `ir.Input` struct (all fields below), `ir.Workflow.Inputs []*ir.Input`, `func (w *Workflow) Input(name string) *Input`, `func (p *Parser) parseInputs()`

- [ ] **Step 1: Write the failing test**

Create `parser/parse_inputs_test.go`:

```go
package parser

import (
	"testing"

	"github.com/2389-research/dippin-lang/ir"
)

// parseSrc parses test source and returns the workflow plus any parse
// diagnostics. Diagnostics come from Parser.Diagnostics(), not from Parse().
func parseSrc(t *testing.T, src string) (*ir.Workflow, []string) {
	t.Helper()
	p := NewParser(src, "test.dip")
	w, err := p.Parse()
	if err != nil {
		t.Fatalf("parse returned error: %v", err)
	}
	return w, p.Diagnostics()
}

func TestParseInputsMinimalForm(t *testing.T) {
	src := `workflow W
  goal: "test"
  start: A
  exit: A

  inputs
    idea: text
    count: number

  agent A
    prompt:
      hi
`
	w, diags := parseSrc(t, src)
	if len(diags) != 0 {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if len(w.Inputs) != 2 {
		t.Fatalf("got %d inputs, want 2", len(w.Inputs))
	}
	if w.Inputs[0].Name != "idea" || w.Inputs[0].Type != "text" {
		t.Errorf("input 0 = %q/%q, want idea/text", w.Inputs[0].Name, w.Inputs[0].Type)
	}
	if w.Inputs[1].Name != "count" || w.Inputs[1].Type != "number" {
		t.Errorf("input 1 = %q/%q, want count/number", w.Inputs[1].Name, w.Inputs[1].Type)
	}
}

func TestParseInputsPreservesDeclarationOrder(t *testing.T) {
	src := `workflow W
  goal: "test"
  start: A
  exit: A

  inputs
    zebra: text
    apple: text
    middle: text

  agent A
    prompt:
      hi
`
	w, _ := parseSrc(t, src)
	want := []string{"zebra", "apple", "middle"}
	if len(w.Inputs) != len(want) {
		t.Fatalf("got %d inputs, want %d", len(w.Inputs), len(want))
	}
	for i, name := range want {
		if w.Inputs[i].Name != name {
			t.Errorf("input %d = %q, want %q", i, w.Inputs[i].Name, name)
		}
	}
}

func TestParseNoInputsBlockLeavesNil(t *testing.T) {
	src := `workflow W
  goal: "test"
  start: A
  exit: A

  agent A
    prompt:
      hi
`
	w, _ := parseSrc(t, src)
	if w.Inputs != nil {
		t.Errorf("Inputs = %v, want nil", w.Inputs)
	}
}

func TestParseInputsDuplicateNameDiagnoses(t *testing.T) {
	src := `workflow W
  goal: "test"
  start: A
  exit: A

  inputs
    idea: text
    idea: number

  agent A
    prompt:
      hi
`
	_, diags := parseSrc(t, src)
	if len(diags) == 0 {
		t.Fatal("expected a duplicate-name diagnostic, got none")
	}
}

func TestParseInputsRecordsSourceLocation(t *testing.T) {
	src := `workflow W
  goal: "test"
  start: A
  exit: A

  inputs
    idea: text

  agent A
    prompt:
      hi
`
	w, _ := parseSrc(t, src)
	if len(w.Inputs) != 1 {
		t.Fatalf("got %d inputs, want 1", len(w.Inputs))
	}
	if w.Inputs[0].Source.Line == 0 {
		t.Error("Source.Line = 0, want the declaration line")
	}
}
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `just test-pkg parser`
Expected: FAIL — `w.Inputs` undefined (compile error).

- [ ] **Step 3: Add the IR types**

In `ir/ir.go`, add to the `Workflow` struct, immediately after the `Requires` field:

```go
	// Inputs declares the values a caller must supply — a human at the entry
	// point, or a parent workflow via a subgraph node's params:. This is the
	// callee-side signature. Declaration order is significant: a host renders
	// these as an ordered form, so the formatter must not sort them. Values are
	// untrusted by construction and are read as ${inputs.name}. See issue #190.
	Inputs []*Input
```

Then add the `Input` type after the `WorkflowDefaults` struct:

```go
// Input declares one caller-supplied value bound at run start.
//
// Default, Min and Max are stored as raw source text rather than typed values so
// the formatter can round-trip a file byte-for-byte; the CLI's JSON projection
// coerces them per Type. Unknown Type values are carried verbatim and diagnosed
// by the validator (DIP155), never rejected by the parser — that keeps a .dip
// using a future type parseable, formattable, and packable on an older dippin.
type Input struct {
	Name        string
	Type        string   // v1: text | number | bool | enum | file | secret
	Required    bool     // Host must obtain a value even when Default is set
	Default     string   // Raw source text; a form prefill, not a substitute for Required
	HasDefault  bool     // Distinguishes an absent default from an empty-string default
	Prompt      string   // What a host asks the caller
	Description string   // Help text
	Options     []string // enum choices
	Pattern     string   // text: regex the host enforces
	Min         string   // number: inclusive lower bound, raw text
	Max         string   // number: inclusive upper bound, raw text
	MaxLength   int      // text: character cap
	Multiline   bool     // text: host renders a textarea
	Source      SourceLocation
}
```

- [ ] **Step 4: Add the lookup helper**

In `ir/lookup.go`, add:

```go
// Input returns the declared input with the given name, or nil if the workflow
// declares no such input. Used by the validator to resolve ${inputs.x} against
// the declaration — inputs is the only closed namespace in the language.
func (w *Workflow) Input(name string) *Input {
	for _, in := range w.Inputs {
		if in.Name == name {
			return in
		}
	}
	return nil
}
```

- [ ] **Step 5: Create the parser**

Create `parser/parse_inputs.go`:

```go
// ABOUTME: Parses the workflow-level `inputs` block — the callee-side signature
// ABOUTME: declaring what a caller must supply. See issue #190.
package parser

import (
	"fmt"

	"github.com/2389-research/dippin-lang/ir"
)

// parseInputs parses the `inputs` section. Each entry is `name: type`, optionally
// extended by an indented attribute block (added in a later task).
func (p *Parser) parseInputs() {
	p.lexer.NextToken() // "inputs"
	p.expect(TokenNewline)
	p.expect(TokenIndent)
	p.parseInputsBody()
	p.expect(TokenOutdent)
}

// parseInputsBody scans input declarations until the block's outdent.
func (p *Parser) parseInputsBody() {
	for p.lexer.PeekToken().Type != TokenOutdent && p.lexer.PeekToken().Type != TokenEOF {
		t := p.lexer.PeekToken()
		if t.Type == TokenNewline {
			p.lexer.NextToken()
			continue
		}
		if t.Type == TokenIdentifier {
			p.appendInput(p.parseOneInput(t))
			continue
		}
		p.diagnostics = append(p.diagnostics,
			fmt.Sprintf("unexpected token in inputs block at %d:%d", t.Location.Line, t.Location.Column))
		p.lexer.NextToken()
	}
}

// appendInput adds a parsed input, diagnosing a duplicate name. The duplicate is
// still appended so the formatter round-trips the source as written.
func (p *Parser) appendInput(in *ir.Input) {
	if p.workflow.Input(in.Name) != nil {
		p.diagnostics = append(p.diagnostics,
			fmt.Sprintf("duplicate input %q at %d:%d", in.Name, in.Source.Line, in.Source.Column))
	}
	p.workflow.Inputs = append(p.workflow.Inputs, in)
}

// parseOneInput parses `name: type` and any indented attribute block.
func (p *Parser) parseOneInput(t Token) *ir.Input {
	p.lexer.NextToken() // name
	p.expect(TokenColon)
	typ := p.readFieldValue(t.Location.Line)
	return &ir.Input{Name: t.Value, Type: typ, Source: t.Location}
}
```

- [ ] **Step 6: Wire the dispatcher**

In `parser/parser.go`, in `dispatchWorkflowSimpleField`, add the `inputs` case alongside `vars`:

```go
	case "vars":
		p.parseVars()
	case "inputs":
		p.parseInputs()
```

Update that function's doc comment to read: `// dispatchWorkflowSimpleField handles header fields and config blocks (defaults, vars, inputs). Returns true if handled.`

- [ ] **Step 7: Run the tests**

Run: `just test-pkg parser`
Expected: PASS — all five tests.

- [ ] **Step 8: Run the full suite**

Run: `just test`
Expected: PASS. If a formatter round-trip test fails, that is expected and Task 3 fixes it — note which test and continue.

- [ ] **Step 9: Commit**

```bash
git add ir/ir.go ir/lookup.go parser/parse_inputs.go parser/parser.go parser/parse_inputs_test.go
git commit -m "$(cat <<'EOF'
feat(ir,parser): inputs block — IR types and minimal-form parsing (#190)

Adds ir.Input and Workflow.Inputs, plus a parser for the one-line
`name: type` declaration form. Declaration order is preserved; a
duplicate name is diagnosed but still appended so the source
round-trips.

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

### Task 2: Input attribute blocks

Adds the indented attribute body: ten attributes, an unknown-attribute hint, and verbatim carriage of unknown types.

**Files:**
- Modify: `parser/parse_inputs.go`
- Test: `parser/parse_inputs_test.go`

**Interfaces:**
- Consumes: `ir.Input`, `p.parseOneInput` from Task 1
- Produces: `func (p *Parser) parseInputFields(in *ir.Input)`, `func (p *Parser) applyInputField(in *ir.Input, key, val string, loc ir.SourceLocation)`

- [ ] **Step 1: Write the failing test**

Append to `parser/parse_inputs_test.go`:

```go
func TestParseInputsFullAttributeBlock(t *testing.T) {
	src := `workflow W
  goal: "test"
  start: A
  exit: A

  inputs
    idea: text
      required: true
      prompt: "What do you want built?"
      description: "One or two sentences."
      multiline: true
      max_length: 4000
    risk: enum
      options: low, medium, high
      default: medium
    retries: number
      min: 1
      max: 10
    branch: text
      pattern: "^[a-z]+$"

  agent A
    prompt:
      hi
`
	w, diags := parseSrc(t, src)
	if len(diags) != 0 {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if len(w.Inputs) != 4 {
		t.Fatalf("got %d inputs, want 4", len(w.Inputs))
	}

	idea := w.Inputs[0]
	if !idea.Required {
		t.Error("idea.Required = false, want true")
	}
	if idea.Prompt != "What do you want built?" {
		t.Errorf("idea.Prompt = %q", idea.Prompt)
	}
	if idea.Description != "One or two sentences." {
		t.Errorf("idea.Description = %q", idea.Description)
	}
	if !idea.Multiline {
		t.Error("idea.Multiline = false, want true")
	}
	if idea.MaxLength != 4000 {
		t.Errorf("idea.MaxLength = %d, want 4000", idea.MaxLength)
	}

	risk := w.Inputs[1]
	if len(risk.Options) != 3 || risk.Options[0] != "low" || risk.Options[2] != "high" {
		t.Errorf("risk.Options = %v, want [low medium high]", risk.Options)
	}
	if risk.Default != "medium" || !risk.HasDefault {
		t.Errorf("risk.Default = %q (has=%v), want medium/true", risk.Default, risk.HasDefault)
	}

	retries := w.Inputs[2]
	if retries.Min != "1" || retries.Max != "10" {
		t.Errorf("retries min/max = %q/%q, want 1/10", retries.Min, retries.Max)
	}

	if w.Inputs[3].Pattern != "^[a-z]+$" {
		t.Errorf("branch.Pattern = %q", w.Inputs[3].Pattern)
	}
}

func TestParseInputsEmptyDefaultIsDistinctFromAbsent(t *testing.T) {
	src := `workflow W
  goal: "test"
  start: A
  exit: A

  inputs
    withEmpty: text
      default: ""
    withNone: text

  agent A
    prompt:
      hi
`
	w, _ := parseSrc(t, src)
	if !w.Inputs[0].HasDefault {
		t.Error("withEmpty.HasDefault = false, want true")
	}
	if w.Inputs[0].Default != "" {
		t.Errorf("withEmpty.Default = %q, want empty", w.Inputs[0].Default)
	}
	if w.Inputs[1].HasDefault {
		t.Error("withNone.HasDefault = true, want false")
	}
}

func TestParseInputsUnknownTypeCarriedVerbatim(t *testing.T) {
	src := `workflow W
  goal: "test"
  start: A
  exit: A

  inputs
    when: duration
      required: true

  agent A
    prompt:
      hi
`
	w, diags := parseSrc(t, src)
	if len(diags) != 0 {
		t.Fatalf("parser must not diagnose an unknown type (that is DIP155's job): %v", diags)
	}
	if w.Inputs[0].Type != "duration" {
		t.Errorf("Type = %q, want duration carried verbatim", w.Inputs[0].Type)
	}
	if !w.Inputs[0].Required {
		t.Error("attributes must still parse under an unknown type")
	}
}

func TestParseInputsUnknownAttributeHints(t *testing.T) {
	src := `workflow W
  goal: "test"
  start: A
  exit: A

  inputs
    idea: text
      nonsense: 42
      required: true

  agent A
    prompt:
      hi
`
	w, diags := parseSrc(t, src)
	if len(diags) == 0 {
		t.Fatal("expected an unknown-attribute hint")
	}
	if !w.Inputs[0].Required {
		t.Error("a stray attribute must not desync the scan — required: true was lost")
	}
}
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `just test-pkg parser`
Expected: FAIL — attributes are not parsed; `idea.Required` is false.

- [ ] **Step 3: Parse the attribute block**

In `parser/parse_inputs.go`, replace `parseOneInput` with:

```go
// parseOneInput parses `name: type` plus any indented attribute block.
func (p *Parser) parseOneInput(t Token) *ir.Input {
	p.lexer.NextToken() // name
	p.expect(TokenColon)
	typ := p.readFieldValue(t.Location.Line)
	in := &ir.Input{Name: t.Value, Type: typ, Source: t.Location}

	if p.lexer.PeekToken().Type == TokenNewline {
		p.lexer.NextToken()
	}
	if p.lexer.PeekToken().Type != TokenIndent {
		return in
	}
	p.expect(TokenIndent)
	p.parseInputFields(in)
	p.expect(TokenOutdent)
	return in
}

// parseInputFields parses the attributes inside one input's indented block.
func (p *Parser) parseInputFields(in *ir.Input) {
	for p.lexer.PeekToken().Type != TokenOutdent && p.lexer.PeekToken().Type != TokenEOF {
		t := p.lexer.PeekToken()
		if t.Type == TokenNewline {
			p.lexer.NextToken()
			continue
		}
		if t.Type != TokenIdentifier {
			p.lexer.NextToken()
			continue
		}
		p.lexer.NextToken() // key
		p.expect(TokenColon)
		val := p.readFieldValue(t.Location.Line)
		p.applyInputField(in, t.Value, val, t.Location)
	}
}
```

- [ ] **Step 4: Add the attribute appliers**

Append to `parser/parse_inputs.go`. Each switch stays within the 4-case complexity budget:

```go
// applyInputField dispatches one attribute, hinting on an unrecognized key.
// Unknown attributes are a hint, never a parse failure — see issue #190's
// forward-compatibility requirement.
func (p *Parser) applyInputField(in *ir.Input, key, val string, loc ir.SourceLocation) {
	if applyInputTextField(in, key, val) {
		return
	}
	if applyInputValueField(in, key, val) {
		return
	}
	if p.applyInputParsedField(in, key, val, loc) {
		return
	}
	p.emitUnknownFieldHint("input", key, loc)
}

// applyInputTextField handles plain string attributes.
func applyInputTextField(in *ir.Input, key, val string) bool {
	switch key {
	case "prompt":
		in.Prompt = val
	case "description":
		in.Description = val
	case "pattern":
		in.Pattern = val
	default:
		return false
	}
	return true
}

// applyInputValueField handles the default and the constraints kept as raw text.
func applyInputValueField(in *ir.Input, key, val string) bool {
	switch key {
	case "default":
		in.Default = val
		in.HasDefault = true
	case "options":
		in.Options = splitCommaNoEmpty(val)
	case "min":
		in.Min = val
	case "max":
		in.Max = val
	default:
		return false
	}
	return true
}

// applyInputParsedField handles attributes needing conversion.
func (p *Parser) applyInputParsedField(in *ir.Input, key, val string, loc ir.SourceLocation) bool {
	switch key {
	case "required":
		in.Required = p.parseBoolAttr(val, key, loc)
	case "multiline":
		in.Multiline = p.parseBoolAttr(val, key, loc)
	case "max_length":
		in.MaxLength = p.parseInt(val, key, loc)
	default:
		return false
	}
	return true
}
```

- [ ] **Step 5: Run the tests**

Run: `just test-pkg parser`
Expected: PASS.

- [ ] **Step 6: Verify the complexity budget**

Run: `just complexity`
Expected: no output (clean). If any new function is flagged, split its switch further — do not add `//nolint`.

- [ ] **Step 7: Commit**

```bash
git add parser/parse_inputs.go parser/parse_inputs_test.go
git commit -m "$(cat <<'EOF'
feat(parser): inputs attribute blocks (#190)

Ten attributes across three appliers, each within the 4-case complexity
budget. Unknown types are carried verbatim and unknown attributes emit a
hint — neither is a parse failure, so a .dip using a future type stays
parseable on an older dippin.

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

### Task 3: Formatter emission and round-trip

Emits the `inputs` section in canonical position and attribute order, preserving declaration order.

**Files:**
- Modify: `formatter/format.go:31-42` (`writeWorkflowSections`), plus new writers
- Test: `formatter/format_inputs_test.go`

**Interfaces:**
- Consumes: `ir.Workflow.Inputs`, `ir.Input` from Task 1
- Produces: `func writeInputs(wr *writer, inputs []*ir.Input)`

- [ ] **Step 1: Write the failing test**

Create `formatter/format_inputs_test.go`:

```go
package formatter_test

import (
	"strings"
	"testing"

	"github.com/2389-research/dippin-lang/formatter"
	"github.com/2389-research/dippin-lang/ir"
	"github.com/2389-research/dippin-lang/parser"
)

// parseSrc parses test source and returns the workflow plus any parse
// diagnostics. Diagnostics come from Parser.Diagnostics(), not from Parse().
func parseSrc(t *testing.T, src string) (*ir.Workflow, []string) {
	t.Helper()
	p := parser.NewParser(src, "test.dip")
	w, err := p.Parse()
	if err != nil {
		t.Fatalf("parse returned error: %v", err)
	}
	return w, p.Diagnostics()
}

const inputsSource = `workflow W
  goal: "test"
  start: A
  exit: A

  inputs
    idea: text
      required: true
      prompt: "What do you want built?"
      multiline: true
      max_length: 4000
    risk: enum
      options: low, medium, high
      default: medium
    plain: text

  agent A
    prompt:
      hi
`

func TestFormatInputsRoundTrip(t *testing.T) {
	w, diags := parseSrc(t, inputsSource)
	if len(diags) != 0 {
		t.Fatalf("parse diagnostics: %v", diags)
	}
	out := formatter.Format(w)

	w2, diags2 := parseSrc(t, out)
	if len(diags2) != 0 {
		t.Fatalf("reparse diagnostics: %v\n%s", diags2, out)
	}
	if got := formatter.Format(w2); got != out {
		t.Errorf("format is not idempotent:\n--- first ---\n%s\n--- second ---\n%s", out, got)
	}
	if len(w2.Inputs) != 3 {
		t.Fatalf("got %d inputs after round-trip, want 3", len(w2.Inputs))
	}
}

func TestFormatInputsPreservesDeclarationOrder(t *testing.T) {
	src := `workflow W
  goal: "test"
  start: A
  exit: A

  inputs
    zebra: text
    apple: text

  agent A
    prompt:
      hi
`
	w, _ := parseSrc(t, src)
	out := formatter.Format(w)
	zi, ai := strings.Index(out, "zebra"), strings.Index(out, "apple")
	if zi == -1 || ai == -1 {
		t.Fatalf("both inputs must be emitted:\n%s", out)
	}
	if zi > ai {
		t.Errorf("declaration order not preserved — inputs must not be sorted:\n%s", out)
	}
}

func TestFormatInputsSectionPosition(t *testing.T) {
	src := `workflow W
  goal: "test"
  start: A
  exit: A

  defaults
    max_retries: 3

  vars
    k: v

  inputs
    idea: text

  agent A
    prompt:
      hi
`
	w, _ := parseSrc(t, src)
	out := formatter.Format(w)
	ii := strings.Index(out, "\n  inputs\n")
	di := strings.Index(out, "\n  defaults\n")
	vi := strings.Index(out, "\n  vars\n")
	if ii == -1 || di == -1 || vi == -1 {
		t.Fatalf("all three sections must be emitted:\n%s", out)
	}
	if !(ii < di && di < vi) {
		t.Errorf("want inputs < defaults < vars:\n%s", out)
	}
}

func TestFormatNoInputsEmitsNoBlock(t *testing.T) {
	src := `workflow W
  goal: "test"
  start: A
  exit: A

  agent A
    prompt:
      hi
`
	w, _ := parseSrc(t, src)
	if out := formatter.Format(w); strings.Contains(out, "inputs") {
		t.Errorf("emitted an inputs block for a workflow with none:\n%s", out)
	}
}

func TestFormatInputsOmitsZeroValues(t *testing.T) {
	src := `workflow W
  goal: "test"
  start: A
  exit: A

  inputs
    plain: text

  agent A
    prompt:
      hi
`
	w, _ := parseSrc(t, src)
	out := formatter.Format(w)
	for _, unwanted := range []string{"required:", "multiline:", "max_length:", "default:"} {
		if strings.Contains(out, unwanted) {
			t.Errorf("emitted zero-value attribute %q:\n%s", unwanted, out)
		}
	}
}
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `just test-pkg formatter`
Expected: FAIL — no `inputs` block is emitted.

- [ ] **Step 3: Wire the section into the canonical order**

In `formatter/format.go`, in `writeWorkflowSections`, insert **before** the `defaults` block so the order is inputs → defaults → vars:

```go
func writeWorkflowSections(wr *writer, w *ir.Workflow) {
	if len(w.Inputs) > 0 {
		wr.blank()
		writeInputs(wr, w.Inputs)
	}
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
	writeWorkflowTailSections(wr, w)
}
```

- [ ] **Step 4: Add the writers**

Add to `formatter/format.go`, next to `writeVars`:

```go
// writeInputs emits the inputs section. Declaration order is significant —
// a host renders these as an ordered form — so unlike vars, it is not sorted.
func writeInputs(wr *writer, inputs []*ir.Input) {
	wr.line("inputs")
	wr.push()
	for _, in := range inputs {
		writeOneInput(wr, in)
	}
	wr.pop()
}

// writeOneInput emits `name: type` plus any non-zero attributes, in canonical
// order: required, prompt, description, default, options, pattern, min, max,
// max_length, multiline.
func writeOneInput(wr *writer, in *ir.Input) {
	wr.line("%s: %s", in.Name, in.Type)
	wr.push()
	writeInputTextAttrs(wr, in)
	writeInputValueAttrs(wr, in)
	writeInputNumericAttrs(wr, in)
	wr.pop()
}

// writeInputTextAttrs emits required, prompt, and description.
func writeInputTextAttrs(wr *writer, in *ir.Input) {
	if in.Required {
		wr.line("required: true")
	}
	if in.Prompt != "" {
		wr.line("prompt: %s", quoteValue(in.Prompt))
	}
	if in.Description != "" {
		wr.line("description: %s", quoteValue(in.Description))
	}
}

// writeInputValueAttrs emits default, options, and pattern.
func writeInputValueAttrs(wr *writer, in *ir.Input) {
	if in.HasDefault {
		wr.line("default: %s", quoteValue(in.Default))
	}
	if len(in.Options) > 0 {
		wr.line("options: %s", strings.Join(in.Options, ", "))
	}
	if in.Pattern != "" {
		wr.line("pattern: %s", quoteValue(in.Pattern))
	}
}

// writeInputNumericAttrs emits min, max, max_length, and multiline.
func writeInputNumericAttrs(wr *writer, in *ir.Input) {
	if in.Min != "" {
		wr.line("min: %s", in.Min)
	}
	if in.Max != "" {
		wr.line("max: %s", in.Max)
	}
	if in.MaxLength > 0 {
		wr.line("max_length: %d", in.MaxLength)
	}
	if in.Multiline {
		wr.line("multiline: true")
	}
}
```

- [ ] **Step 5: Run the tests**

Run: `just test-pkg formatter`
Expected: PASS.

If `TestFormatInputsRoundTrip` fails on idempotency because an empty attribute block emits a stray blank indent for an input with no attributes, guard `writeOneInput`'s push/pop — only push when at least one attribute will be emitted.

- [ ] **Step 6: Run the full suite and complexity**

Run: `just test && just complexity`
Expected: PASS, clean.

- [ ] **Step 7: Commit**

```bash
git add formatter/format.go formatter/format_inputs_test.go
git commit -m "$(cat <<'EOF'
feat(formatter): emit the inputs section (#190)

Canonical position is directly after the header — inputs is the file's
contract, ahead of defaults and vars, which are configuration.
Declaration order is preserved rather than sorted: a host renders these
as an ordered form, so the author's order is the intended order.

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

### Task 4: DIP155 — unknown input type

The first **error-severity lint** in the codebase. This task also fixes `dippin lint`'s exit code, which today ignores lint-severity errors entirely.

**Files:**
- Modify: `validator/lint_codes.go` (add `DIP155` const and description)
- Create: `validator/lint_inputs.go`
- Modify: `validator/lint.go:97-114` (`lintPasses` — register the pass)
- Modify: `cmd/dippin/cmd_validate.go:85-89` (`CmdLint` exit code)
- Test: `validator/lint_inputs_test.go`

**Interfaces:**
- Consumes: `ir.Workflow.Inputs`, `ir.Input`
- Produces: `func lintUnknownInputType(w *ir.Workflow) []Diagnostic`, `var knownInputTypes map[string]bool`

- [ ] **Step 1: Write the failing test**

Create `validator/lint_inputs_test.go`:

```go
package validator_test

import (
	"testing"

	"github.com/2389-research/dippin-lang/parser"
	"github.com/2389-research/dippin-lang/validator"
)

// lintSrc parses source and returns its lint diagnostics. Parse diagnostics
// come from Parser.Diagnostics(), not from Parse() — there is no
// parser.Parse(src) package function.
func lintSrc(t *testing.T, src string) []validator.Diagnostic {
	t.Helper()
	p := parser.NewParser(src, "test.dip")
	w, err := p.Parse()
	if err != nil {
		t.Fatalf("parse returned error: %v", err)
	}
	if diags := p.Diagnostics(); len(diags) != 0 {
		t.Fatalf("parse diagnostics: %v", diags)
	}
	return validator.Lint(w).Diagnostics
}

// hasCode reports whether any diagnostic carries the given code.
func hasCode(diags []validator.Diagnostic, code string) bool {
	for _, d := range diags {
		if d.Code == code {
			return true
		}
	}
	return false
}

func TestDIP155UnknownInputType(t *testing.T) {
	src := `workflow W
  goal: "test"
  start: A
  exit: A

  inputs
    when: duration

  agent A
    prompt:
      hi
`
	diags := lintSrc(t, src)
	if !hasCode(diags, "DIP155") {
		t.Fatalf("want DIP155 for unknown type, got %v", diags)
	}
	for _, d := range diags {
		if d.Code == "DIP155" && d.Severity != validator.SeverityError {
			t.Errorf("DIP155 severity = %v, want Error", d.Severity)
		}
	}
}

func TestDIP155AcceptsEveryKnownType(t *testing.T) {
	src := `workflow W
  goal: "test"
  start: A
  exit: A

  inputs
    a: text
    b: number
    c: bool
    d: enum
    e: file
    f: secret

  agent A
    prompt:
      hi
`
	if diags := lintSrc(t, src); hasCode(diags, "DIP155") {
		t.Errorf("DIP155 fired on a known type: %v", diags)
	}
}
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `just test-pkg validator`
Expected: FAIL — no DIP155 diagnostic.

- [ ] **Step 3: Add the code**

In `validator/lint_codes.go`, after the `DIP154` const:

```go
	DIP155 = "DIP155" // input declares a type this dippin does not recognize
```

And in the same file's description map, after the `DIP154` entry:

```go
	DIP155: "input declares an unrecognized type",
```

- [ ] **Step 4: Write the lint pass**

Create `validator/lint_inputs.go`:

```go
// ABOUTME: Lint passes for the workflow-level inputs block (DIP155-DIP157).
// ABOUTME: dippin lints the declaration; the engine validates supplied values.
package validator

import (
	"fmt"

	"github.com/2389-research/dippin-lang/ir"
)

// knownInputTypes is the v1 closed set of input types. A type outside this set
// is carried through the parser and IR verbatim and diagnosed here — that keeps
// a .dip using a future type parseable, formattable and packable on an older
// dippin, so only the lint complains. See issue #190.
var knownInputTypes = map[string]bool{
	"text":   true,
	"number": true,
	"bool":   true,
	"enum":   true,
	"file":   true,
	"secret": true,
}

// lintUnknownInputType checks DIP155: an input's declared type must be one this
// dippin recognizes.
func lintUnknownInputType(w *ir.Workflow) []Diagnostic {
	var diags []Diagnostic
	for _, in := range w.Inputs {
		if knownInputTypes[in.Type] {
			continue
		}
		diags = append(diags, Diagnostic{
			Code:     DIP155,
			Severity: SeverityError,
			Message:  fmt.Sprintf("input %q declares unrecognized type %q", in.Name, in.Type),
			Location: in.Source,
			Help:     "valid types are text, number, bool, enum, file, secret — or upgrade dippin if this type is newer than this build",
		})
	}
	return diags
}
```

- [ ] **Step 5: Register the pass**

In `validator/lint.go`, in the `lintPasses` slice, add after `lintPromptOptOut`:

```go
		lintUnknownInputType,
```

Also update `Lint`'s doc comment range from `DIP101–DIP154` to `DIP101–DIP155`.

- [ ] **Step 6: Run the tests**

Run: `just test-pkg validator`
Expected: PASS.

- [ ] **Step 7: Write the exit-code test**

DIP155 is the **first** error-severity lint in the codebase. `CmdLint` currently exits non-zero only on `valRes.HasErrors()` — structural errors — so an error-severity *lint* would print and exit 0. Add to `cmd/dippin/main_test.go`:

```go
func TestLintExitsNonZeroOnErrorSeverityLint(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "unknown_type.dip")
	src := `workflow W
  goal: "test"
  start: A
  exit: A

  inputs
    when: duration

  agent A
    prompt:
      hi
`
	if err := os.WriteFile(path, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	cli := &CLI{Stdout: &stdout, Stderr: &stderr}
	if got := cli.CmdLint([]string{path}); got != ExitError {
		t.Errorf("CmdLint exit = %v, want ExitError — an error-severity lint must fail the command", got)
	}
}
```

Confirm the imports (`bytes`, `os`, `path/filepath`) are present in that file; add any that are missing.

- [ ] **Step 8: Run it to verify it fails**

Run: `just test-pkg cmd/dippin`
Expected: FAIL — exit is `ExitOK` because lint errors are not counted.

- [ ] **Step 9: Fix the exit code**

In `cmd/dippin/cmd_validate.go`, in `CmdLint`, replace the final exit block:

```go
	// Exit 1 on any error-severity diagnostic, whether structural (Validate) or
	// semantic (Lint). Lint gained its first error-severity codes with the
	// inputs block (DIP155-DIP157, #190); before that every lint diagnostic was
	// a warning, so checking valRes alone was sufficient.
	if valRes.HasErrors() || hasErrorSeverity(lintRes.Diagnostics) {
		return ExitError
	}
	return ExitOK
}

// hasErrorSeverity reports whether any diagnostic is error-severity.
func hasErrorSeverity(diags []validator.Diagnostic) bool {
	for _, d := range diags {
		if d.Severity == validator.SeverityError {
			return true
		}
	}
	return false
}
```

Also update `CmdLint`'s doc comment: `// Errors — structural or error-severity lint — cause exit 1; warnings alone exit 0.`

- [ ] **Step 10: Run the tests**

Run: `just test-pkg cmd/dippin && just test`
Expected: PASS.

- [ ] **Step 11: Commit**

```bash
git add validator/lint_codes.go validator/lint_inputs.go validator/lint.go cmd/dippin/cmd_validate.go validator/lint_inputs_test.go cmd/dippin/main_test.go
git commit -m "$(cat <<'EOF'
feat(validator): DIP155 unknown input type (#190)

First error-severity lint in the codebase, which exposed that CmdLint
exited non-zero only on structural Validate errors — an error-severity
lint printed and exited 0. CmdLint now fails on either.

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

### Task 5: DIP156 — reference to an undeclared input

`inputs.` is the first *closed* namespace in the language: a reference is checked against the declaration, not just for a valid prefix. Two scan paths — prompts use `${inputs.x}`, edge conditions use bare `inputs.x`.

**Files:**
- Modify: `validator/lint.go:124-129` (`knownNamespaces` — add `inputs`)
- Modify: `validator/lint_inputs.go`
- Modify: `validator/lint_codes.go`
- Modify: `validator/lint.go` (`lintPasses`)
- Test: `validator/lint_inputs_test.go`

**Interfaces:**
- Consumes: `ir.Workflow.Input(name)` from Task 1, `varRefPattern` and `nodePrompt` from `validator/lint_context.go`, `extractComparisons` from `validator/lint_conditions.go`
- Produces: `func lintUndeclaredInputRef(w *ir.Workflow) []Diagnostic`

- [ ] **Step 1: Write the failing test**

Append to `validator/lint_inputs_test.go`:

```go
func TestDIP156UndeclaredRefInPrompt(t *testing.T) {
	src := `workflow W
  goal: "test"
  start: A
  exit: A

  inputs
    idea: text

  agent A
    prompt:
      Build ${inputs.idae} for me.
`
	diags := lintSrc(t, src)
	if !hasCode(diags, "DIP156") {
		t.Fatalf("want DIP156 for a typo'd input ref, got %v", diags)
	}
}

func TestDIP156DeclaredRefInPromptIsClean(t *testing.T) {
	src := `workflow W
  goal: "test"
  start: A
  exit: A

  inputs
    idea: text

  agent A
    prompt:
      Build ${inputs.idea} for me.
`
	diags := lintSrc(t, src)
	if hasCode(diags, "DIP156") {
		t.Errorf("DIP156 fired on a declared input: %v", diags)
	}
	if hasCode(diags, "DIP106") {
		t.Errorf("DIP106 fired on the inputs namespace — it must be in knownNamespaces: %v", diags)
	}
}

func TestDIP156UndeclaredRefInEdgeCondition(t *testing.T) {
	src := `workflow W
  goal: "test"
  start: A
  exit: B

  inputs
    risk: enum
      options: low, high

  agent A
    prompt:
      hi

  agent B
    prompt:
      bye

  edges
    A -> B when inputs.rsk = high
`
	diags := lintSrc(t, src)
	if !hasCode(diags, "DIP156") {
		t.Fatalf("want DIP156 for a typo'd input ref in a condition, got %v", diags)
	}
}

func TestDIP156DeclaredRefInEdgeConditionIsClean(t *testing.T) {
	src := `workflow W
  goal: "test"
  start: A
  exit: B

  inputs
    risk: enum
      options: low, high

  agent A
    prompt:
      hi

  agent B
    prompt:
      bye

  edges
    A -> B when inputs.risk = high
`
	diags := lintSrc(t, src)
	if hasCode(diags, "DIP156") {
		t.Errorf("DIP156 fired on a declared input: %v", diags)
	}
	if hasCode(diags, "DIP120") {
		t.Errorf("DIP120 fired on the inputs namespace: %v", diags)
	}
}
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `just test-pkg validator`
Expected: FAIL — no DIP156; also DIP106/DIP120 fire because `inputs` is not a known namespace.

- [ ] **Step 3: Register the namespace**

In `validator/lint.go`, add to `knownNamespaces`:

```go
	"inputs": true,
```

Extend that variable's doc comment with:

```go
// inputs. (declared caller-supplied values) is unlike the others: membership
// here only stops DIP106/DIP120 flagging the prefix. It is a *closed*
// namespace — every key is additionally resolved against Workflow.Inputs by
// DIP156 in lint_inputs.go.
```

- [ ] **Step 4: Add the code constant**

In `validator/lint_codes.go`:

```go
	DIP156 = "DIP156" // reference to an input the workflow does not declare
```

and the description:

```go
	DIP156: "reference to an undeclared input",
```

- [ ] **Step 5: Write the lint pass**

Append to `validator/lint_inputs.go` (add `strings` to the imports):

```go
// inputsPrefix is the namespace prefix for declared-input references.
const inputsPrefix = "inputs."

// lintUndeclaredInputRef checks DIP156: every ${inputs.x} in a prompt and every
// bare inputs.x in an edge condition must resolve to a declared input. inputs is
// the only closed namespace in the language — ctx is open, so a typo there is
// undetectable, which is precisely why caller input does not live in ctx.
func lintUndeclaredInputRef(w *ir.Workflow) []Diagnostic {
	var diags []Diagnostic
	diags = append(diags, undeclaredInputRefsInPrompts(w)...)
	diags = append(diags, undeclaredInputRefsInConditions(w)...)
	return diags
}

// undeclaredInputRefsInPrompts scans ${inputs.x} references in node prompts.
func undeclaredInputRefsInPrompts(w *ir.Workflow) []Diagnostic {
	var diags []Diagnostic
	for _, n := range w.Nodes {
		prompt := nodePrompt(n)
		if prompt == "" {
			continue
		}
		for _, m := range varRefPattern.FindAllStringSubmatch(prompt, -1) {
			name, ok := undeclaredInputName(w, m[1])
			if !ok {
				continue
			}
			diags = append(diags, Diagnostic{
				Code:     DIP156,
				Severity: SeverityError,
				Message:  fmt.Sprintf("node %q references undeclared input ${inputs.%s}", n.ID, name),
				Location: n.Source,
				Help:     "declare it in the workflow's inputs block, or correct the name",
			})
		}
	}
	return diags
}

// undeclaredInputRefsInConditions scans bare inputs.x variables in edge
// conditions. Conditions reference variables without ${} — see docs/context.md.
func undeclaredInputRefsInConditions(w *ir.Workflow) []Diagnostic {
	var diags []Diagnostic
	for _, e := range w.Edges {
		if e.Condition == nil || e.Condition.Parsed == nil {
			continue
		}
		for _, cmp := range extractComparisons(e.Condition.Parsed) {
			name, ok := undeclaredInputName(w, cmp.Variable)
			if !ok {
				continue
			}
			diags = append(diags, Diagnostic{
				Code:     DIP156,
				Severity: SeverityError,
				Message:  fmt.Sprintf("edge %s → %s references undeclared input %q", e.From, e.To, "inputs."+name),
				Location: e.Source,
				Help:     "declare it in the workflow's inputs block, or correct the name",
			})
		}
	}
	return diags
}

// undeclaredInputName returns the input name from an inputs.-prefixed reference
// when that input is not declared. ok is false for any other reference.
func undeclaredInputName(w *ir.Workflow, ref string) (string, bool) {
	if !strings.HasPrefix(ref, inputsPrefix) {
		return "", false
	}
	name := strings.TrimPrefix(ref, inputsPrefix)
	if name == "" || w.Input(name) != nil {
		return "", false
	}
	return name, true
}
```

- [ ] **Step 6: Register the pass**

In `validator/lint.go`, add to `lintPasses` after `lintUnknownInputType`:

```go
		lintUndeclaredInputRef,
```

Update `Lint`'s doc comment range to `DIP101–DIP156`.

- [ ] **Step 7: Run the tests**

Run: `just test-pkg validator && just complexity`
Expected: PASS, clean.

- [ ] **Step 8: Commit**

```bash
git add validator/lint.go validator/lint_codes.go validator/lint_inputs.go validator/lint_inputs_test.go
git commit -m "$(cat <<'EOF'
feat(validator): DIP156 undeclared input reference (#190)

inputs is the first closed namespace in the language: membership in
knownNamespaces only silences the DIP106/DIP120 prefix checks, and every
key is additionally resolved against Workflow.Inputs. Covers both scan
paths — ${inputs.x} in prompts and bare inputs.x in edge conditions.

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

### Task 6: DIP157 — input reference inside a tool command

The engine keeps the entire `inputs.` namespace off its shell-interpolation allowlist, so `${inputs.x}` in a `command:` is dead text that expands to nothing, silently.

**Files:**
- Modify: `validator/lint_inputs.go`
- Modify: `validator/lint_codes.go`
- Modify: `validator/lint.go` (`lintPasses`)
- Test: `validator/lint_inputs_test.go`

**Interfaces:**
- Consumes: `ir.ToolConfig` (`Command` field), `varRefPattern`
- Produces: `func lintInputInToolCommand(w *ir.Workflow) []Diagnostic`

- [ ] **Step 1: Write the failing test**

Append to `validator/lint_inputs_test.go`:

```go
func TestDIP157InputRefInToolCommand(t *testing.T) {
	src := `workflow W
  goal: "test"
  start: T
  exit: T

  inputs
    idea: text

  tool T
    command:
      echo ${inputs.idea}
`
	diags := lintSrc(t, src)
	if !hasCode(diags, "DIP157") {
		t.Fatalf("want DIP157 for an input ref in a tool command, got %v", diags)
	}
	for _, d := range diags {
		if d.Code == "DIP157" && d.Severity != validator.SeverityError {
			t.Errorf("DIP157 severity = %v, want Error", d.Severity)
		}
	}
}

func TestDIP157SecretGetsSharperHelp(t *testing.T) {
	src := `workflow W
  goal: "test"
  start: T
  exit: T

  inputs
    token: secret

  tool T
    command:
      curl -H "Authorization: ${inputs.token}" https://example.com
`
	diags := lintSrc(t, src)
	var help string
	for _, d := range diags {
		if d.Code == "DIP157" {
			help = d.Help
		}
	}
	if help == "" {
		t.Fatalf("want DIP157, got %v", diags)
	}
	if !strings.Contains(help, "secret") {
		t.Errorf("help for a secret should mention it, got %q", help)
	}
}

func TestDIP157CleanWhenInputRefIsInAnAgentPrompt(t *testing.T) {
	src := `workflow W
  goal: "test"
  start: A
  exit: A

  inputs
    idea: text

  agent A
    prompt:
      Build ${inputs.idea}.
`
	if diags := lintSrc(t, src); hasCode(diags, "DIP157") {
		t.Errorf("DIP157 fired on an agent prompt, which interpolates fine: %v", diags)
	}
}
```

Add `"strings"` to the test file's imports.

- [ ] **Step 2: Run the test to verify it fails**

Run: `just test-pkg validator`
Expected: FAIL — no DIP157.

- [ ] **Step 3: Add the code constant**

In `validator/lint_codes.go`:

```go
	DIP157 = "DIP157" // input reference inside a tool command never interpolates
```

and the description:

```go
	DIP157: "input reference in a tool command is never interpolated",
```

- [ ] **Step 4: Write the lint pass**

Append to `validator/lint_inputs.go`:

```go
// lintInputInToolCommand checks DIP157: an ${inputs.x} reference inside a tool
// node's command body never interpolates. The runtime keeps the whole inputs.
// namespace off its shell-interpolation allowlist (the same mechanism that
// blocks LLM-origin ctx.* keys from reaching a shell), so the reference is dead
// text that expands to nothing — silently, for every input type. See #190.
func lintInputInToolCommand(w *ir.Workflow) []Diagnostic {
	var diags []Diagnostic
	for _, n := range w.Nodes {
		cfg, ok := n.Config.(ir.ToolConfig)
		if !ok || cfg.Command == "" {
			continue
		}
		diags = append(diags, inputRefsInCommand(w, n, cfg.Command)...)
	}
	return diags
}

// inputRefsInCommand reports every inputs.-namespaced reference in one command body.
func inputRefsInCommand(w *ir.Workflow, n *ir.Node, command string) []Diagnostic {
	var diags []Diagnostic
	for _, m := range varRefPattern.FindAllStringSubmatch(command, -1) {
		if !strings.HasPrefix(m[1], inputsPrefix) {
			continue
		}
		name := strings.TrimPrefix(m[1], inputsPrefix)
		diags = append(diags, Diagnostic{
			Code:     DIP157,
			Severity: SeverityError,
			Message:  fmt.Sprintf("tool %q references ${inputs.%s}, which never interpolates in a command", n.ID, name),
			Location: n.Source,
			Help:     inputInCommandHelp(w, name),
		})
	}
	return diags
}

// inputInCommandHelp tailors the fix hint, sharpening it for a secret.
func inputInCommandHelp(w *ir.Workflow, name string) string {
	if in := w.Input(name); in != nil && in.Type == "secret" {
		return "the runtime never expands inputs into a shell — a secret least of all; pass it through the runtime's credential mechanism instead"
	}
	return "the runtime keeps the inputs namespace off its shell allowlist; route the value through an agent or a declared context key instead"
}
```

- [ ] **Step 5: Register the pass**

In `validator/lint.go`, add to `lintPasses` after `lintUndeclaredInputRef`:

```go
		lintInputInToolCommand,
```

Update `Lint`'s doc comment range to `DIP101–DIP157`.

- [ ] **Step 6: Run the tests**

Run: `just test-pkg validator && just complexity`
Expected: PASS, clean.

- [ ] **Step 7: Commit**

```bash
git add validator/lint.go validator/lint_codes.go validator/lint_inputs.go validator/lint_inputs_test.go
git commit -m "$(cat <<'EOF'
feat(validator): DIP157 input reference in a tool command (#190)

The runtime keeps the whole inputs. namespace off its shell-interpolation
allowlist, so ${inputs.x} in a command: is dead text that expands to
nothing silently. Error severity because the failure mode is an empty
shell variable rather than a crash. Subsumes the narrower
secret-in-command rule as a sharper help message on the same code.

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

### Task 7: `dippin inputs` and the typed JSON projection

The single most important export for the downstream host: a schema document it can render into a form or walk conversationally.

**Files:**
- Create: `cmd/dippin/cmd_inputs.go`
- Modify: `cmd/dippin/cli.go:55-71` (command table)
- Test: `cmd/dippin/cmd_inputs_test.go`

**Interfaces:**
- Consumes: `ir.Workflow.Inputs`, `loadWorkflow(path)` from `cmd/dippin`
- Produces: `func (c *CLI) CmdInputs(args []string) ExitCode`, `func inputsJSON(w *ir.Workflow) []inputJSON`, `func coerceInputValue(typ, raw string) any`

- [ ] **Step 1: Write the failing test**

Create `cmd/dippin/cmd_inputs_test.go`:

```go
package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

const inputsFixture = `workflow W
  goal: "test"
  start: A
  exit: A

  inputs
    idea: text
      required: true
      prompt: "What do you want built?"
      max_length: 4000
    retries: number
      default: 3
    verbose: bool
      default: false
    risk: enum
      options: low, high
      default: low

  agent A
    prompt:
      Build ${inputs.idea}.
`

func writeInputsFixture(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "w.dip")
	if err := os.WriteFile(path, []byte(inputsFixture), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestCmdInputsJSONTypedDefaults(t *testing.T) {
	path := writeInputsFixture(t)
	var stdout, stderr bytes.Buffer
	cli := &CLI{Stdout: &stdout, Stderr: &stderr}
	if got := cli.CmdInputs([]string{"--format=json", path}); got != ExitOK {
		t.Fatalf("exit = %v, stderr = %s", got, stderr.String())
	}

	var got []map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &got); err != nil {
		t.Fatalf("output is not JSON: %v\n%s", err, stdout.String())
	}
	if len(got) != 4 {
		t.Fatalf("got %d inputs, want 4", len(got))
	}

	// Declaration order is preserved.
	if got[0]["name"] != "idea" || got[3]["name"] != "risk" {
		t.Errorf("declaration order not preserved: %v", got)
	}
	// A number default is a JSON number, not a string.
	if _, ok := got[1]["default"].(float64); !ok {
		t.Errorf("retries default = %T(%v), want a JSON number", got[1]["default"], got[1]["default"])
	}
	// A bool default is a JSON bool.
	if v, ok := got[2]["default"].(bool); !ok || v != false {
		t.Errorf("verbose default = %T(%v), want JSON false", got[2]["default"], got[2]["default"])
	}
	// A text/enum default stays a string.
	if _, ok := got[3]["default"].(string); !ok {
		t.Errorf("risk default = %T, want a JSON string", got[3]["default"])
	}
	if got[0]["required"] != true {
		t.Errorf("idea.required = %v, want true", got[0]["required"])
	}
	// An input with no declared default omits the key entirely, rather than
	// emitting a zero value a host would mistake for a real default.
	if _, present := got[0]["default"]; present {
		t.Errorf("idea has no declared default but the key was emitted: %v", got[0])
	}
}

func TestCmdInputsNoInputsEmitsEmptyArray(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "none.dip")
	src := `workflow W
  goal: "test"
  start: A
  exit: A

  agent A
    prompt:
      hi
`
	if err := os.WriteFile(path, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	cli := &CLI{Stdout: &stdout, Stderr: &stderr}
	if got := cli.CmdInputs([]string{"--format=json", path}); got != ExitOK {
		t.Fatalf("exit = %v, stderr = %s", got, stderr.String())
	}
	var got []map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &got); err != nil {
		t.Fatalf("output is not JSON: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("got %d inputs, want an empty array", len(got))
	}
}

func TestCmdInputsTextFormat(t *testing.T) {
	path := writeInputsFixture(t)
	var stdout, stderr bytes.Buffer
	cli := &CLI{Stdout: &stdout, Stderr: &stderr}
	if got := cli.CmdInputs([]string{path}); got != ExitOK {
		t.Fatalf("exit = %v, stderr = %s", got, stderr.String())
	}
	out := stdout.String()
	for _, want := range []string{"idea", "text", "required", "retries"} {
		if !bytes.Contains([]byte(out), []byte(want)) {
			t.Errorf("text output missing %q:\n%s", want, out)
		}
	}
}
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `just test-pkg cmd/dippin`
Expected: FAIL — `CmdInputs` undefined.

- [ ] **Step 3: Write the command**

Create `cmd/dippin/cmd_inputs.go`:

```go
// ABOUTME: `dippin inputs` prints a workflow's declared input schema — the
// ABOUTME: introspection surface a host uses to collect values before a run.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"strconv"

	"github.com/2389-research/dippin-lang/ir"
)

// inputJSON is the stable wire shape of one declared input. Defaults and bounds
// are coerced to real JSON types per the declared type, so a host can inject
// them typed instead of re-flattening everything to string.
type inputJSON struct {
	Name        string   `json:"name"`
	Type        string   `json:"type"`
	Required    bool     `json:"required"`
	Default     any      `json:"default,omitempty"`
	Prompt      string   `json:"prompt,omitempty"`
	Description string   `json:"description,omitempty"`
	Options     []string `json:"options,omitempty"`
	Pattern     string   `json:"pattern,omitempty"`
	Min         any      `json:"min,omitempty"`
	Max         any      `json:"max,omitempty"`
	MaxLength   int      `json:"max_length,omitempty"`
	Multiline   bool     `json:"multiline,omitempty"`
}

// CmdInputs is the dispatcher entry point.
func (c *CLI) CmdInputs(args []string) ExitCode {
	fs := flag.NewFlagSet("inputs", flag.ContinueOnError)
	fs.SetOutput(c.Stderr)
	format := fs.String("format", "text", "output format: text|json")
	if err := fs.Parse(args); err != nil {
		return ExitError
	}
	if fs.NArg() != 1 {
		fmt.Fprintln(c.Stderr, "usage: dippin inputs [--format=text|json] <file>")
		return ExitError
	}

	w, err := loadWorkflow(fs.Arg(0))
	if err != nil {
		c.renderError(err, fs.Arg(0))
		return ExitError
	}
	if *format == "json" {
		return writeInputsJSON(c.Stdout, w)
	}
	return writeInputsText(c.Stdout, w)
}

// writeInputsJSON emits the schema array. A workflow with no inputs emits [],
// never null — a host iterating the result should not have to nil-check.
func writeInputsJSON(out io.Writer, w *ir.Workflow) ExitCode {
	enc := json.NewEncoder(out)
	enc.SetIndent("", "  ")
	if err := enc.Encode(inputsJSON(w)); err != nil {
		return ExitError
	}
	return ExitOK
}

// inputsJSON projects the IR into the wire shape, preserving declaration order.
func inputsJSON(w *ir.Workflow) []inputJSON {
	out := make([]inputJSON, 0, len(w.Inputs))
	for _, in := range w.Inputs {
		out = append(out, oneInputJSON(in))
	}
	return out
}

// oneInputJSON projects a single input.
func oneInputJSON(in *ir.Input) inputJSON {
	j := inputJSON{
		Name:        in.Name,
		Type:        in.Type,
		Required:    in.Required,
		Prompt:      in.Prompt,
		Description: in.Description,
		Options:     in.Options,
		Pattern:     in.Pattern,
		MaxLength:   in.MaxLength,
		Multiline:   in.Multiline,
	}
	if in.HasDefault {
		j.Default = coerceInputValue(in.Type, in.Default)
	}
	if in.Min != "" {
		j.Min = coerceInputValue(in.Type, in.Min)
	}
	if in.Max != "" {
		j.Max = coerceInputValue(in.Type, in.Max)
	}
	return j
}

// coerceInputValue converts raw declaration text to a JSON-native value per the
// declared type. The IR keeps raw text so the formatter round-trips a file
// byte-for-byte; typing happens here. An uncoercible value falls back to the
// raw string — DIP155/DIP158 report the declaration defect, not this projection.
func coerceInputValue(typ, raw string) any {
	switch typ {
	case "number":
		if n, err := strconv.ParseFloat(raw, 64); err == nil {
			return n
		}
	case "bool":
		if b, err := strconv.ParseBool(raw); err == nil {
			return b
		}
	}
	return raw
}

// writeInputsText emits a human-readable listing.
func writeInputsText(out io.Writer, w *ir.Workflow) ExitCode {
	if len(w.Inputs) == 0 {
		fmt.Fprintln(out, "no declared inputs")
		return ExitOK
	}
	for _, in := range w.Inputs {
		req := "optional"
		if in.Required {
			req = "required"
		}
		fmt.Fprintf(out, "%s: %s (%s)\n", in.Name, in.Type, req)
		writeOneInputTextDetail(out, in)
	}
	return ExitOK
}

// writeOneInputTextDetail emits the indented detail lines for one input.
func writeOneInputTextDetail(out io.Writer, in *ir.Input) {
	if in.Prompt != "" {
		fmt.Fprintf(out, "    prompt: %s\n", in.Prompt)
	}
	if in.Description != "" {
		fmt.Fprintf(out, "    description: %s\n", in.Description)
	}
	if in.HasDefault {
		fmt.Fprintf(out, "    default: %s\n", in.Default)
	}
	if len(in.Options) > 0 {
		fmt.Fprintf(out, "    options: %v\n", in.Options)
	}
}
```

- [ ] **Step 4: Register the command**

In `cmd/dippin/cli.go`, add to the command table, keeping alphabetical placement near `"inspect"`:

```go
		"inputs":             c.CmdInputs,
```

- [ ] **Step 5: Run the tests**

Run: `just test-pkg cmd/dippin`
Expected: PASS.

- [ ] **Step 6: Verify the help text lists the command**

Run: `just build && ./dippin help 2>&1 | grep inputs`
Expected: the `inputs` command appears. If the help text is a hand-maintained list in `cmd/dippin/cli.go`, add a one-line entry: `inputs      Print a workflow's declared input schema`.

- [ ] **Step 7: Run the full suite and complexity**

Run: `just test && just complexity`
Expected: PASS, clean.

- [ ] **Step 8: Commit**

```bash
git add cmd/dippin/cmd_inputs.go cmd/dippin/cli.go cmd/dippin/cmd_inputs_test.go
git commit -m "$(cat <<'EOF'
feat(cli): dippin inputs — typed JSON introspection surface (#190)

The schema export a host uses to collect values before a run. The IR
keeps raw declaration text so the formatter round-trips byte-for-byte;
this projection coerces per declared type, so number and bool defaults
arrive as JSON numbers and bools rather than strings. Declaration order
is preserved; a workflow with no inputs emits [] rather than null.

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

### Task 8: Surface a bundle's entry inputs in `dippin inspect`

A host should be able to enumerate what to collect from a `.dipx` without unpacking it.

**Files:**
- Modify: `cmd/dippin/cmd_inspect.go` (`printInspectJSON`, `printManifestJSON`)
- Test: `cmd/dippin/cmd_inspect_test.go`

**Interfaces:**
- Consumes: `(*dipx.Bundle).Entry() *ir.Workflow`, `inputsJSON(w) []inputJSON` from Task 7
- Existing helpers in this package (do **not** invent new names): `packForTest(t) string` (in `cmd_unpack_test.go`) packs the fixed `minimalDip` source and returns a bundle path; `writeMinimalEntry(t) (dir, entry string)` (in `cmd_pack_test.go`) writes that source; `runPack(stdout, stderr, args) int`; `runInspect(stdout, stderr, args) int`; exit codes are `exitDipxOK` / `exitDipxUserError` / `exitDipxIOError`, **not** `ExitOK`.

**Key structural constraint:** `printManifestJSON` is shared by both the verify path (`printInspectJSON`, which has a parsed `*dipx.Bundle`) and the `--no-verify` path (`runInspectNoVerify`, which reads the manifest only and has **no** parsed workflow). Inputs exist only on the verify path. Pass them down as a parameter that the no-verify path sets to `nil`, and omit the JSON key entirely when nil — the `--no-verify` contract must not gain an `"inputs": null`.

- [ ] **Step 1: Write the failing test**

`packForTest` packs a fixed minimal workflow with no inputs, so this task needs its own packer for a source with inputs. Append to `cmd/dippin/cmd_inspect_test.go`, matching that file's existing style:

```go
// packInputsBundleForTest packs a workflow that declares inputs and returns the
// bundle path. packForTest packs the fixed minimalDip source, which has none.
func packInputsBundleForTest(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	entry := filepath.Join(dir, "a.dip")
	if err := os.WriteFile(entry, []byte(inputsFixture), 0o644); err != nil {
		t.Fatalf("write entry: %v", err)
	}
	out := filepath.Join(dir, "a.dipx")
	var so, se bytes.Buffer
	if code := runPack(&so, &se, []string{"-o", out, entry}); code != exitDipxOK {
		t.Fatalf("pack failed: %d; %s", code, se.String())
	}
	return out
}

func TestRunInspect_JSONSurfacesEntryInputs(t *testing.T) {
	bundle := packInputsBundleForTest(t)
	var stdout, stderr bytes.Buffer
	if code := runInspect(&stdout, &stderr, []string{"--format=json", bundle}); code != exitDipxOK {
		t.Fatalf("exit code = %d; stderr=%s", code, stderr.String())
	}
	var payload struct {
		Inputs []struct {
			Name     string `json:"name"`
			Type     string `json:"type"`
			Required bool   `json:"required"`
		} `json:"inputs"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &payload); err != nil {
		t.Fatalf("output is not JSON: %v\n%s", err, stdout.String())
	}
	if len(payload.Inputs) != 4 {
		t.Fatalf("got %d inputs, want 4", len(payload.Inputs))
	}
	if payload.Inputs[0].Name != "idea" || payload.Inputs[0].Type != "text" {
		t.Errorf("first input = %q/%q, want idea/text", payload.Inputs[0].Name, payload.Inputs[0].Type)
	}
	if !payload.Inputs[0].Required {
		t.Error("idea.required = false, want true")
	}
}

func TestRunInspect_NoVerifyOmitsInputsKey(t *testing.T) {
	bundle := packInputsBundleForTest(t)
	var stdout, stderr bytes.Buffer
	if code := runInspect(&stdout, &stderr, []string{"--format=json", "--no-verify", bundle}); code != exitDipxOK {
		t.Fatalf("exit code = %d; stderr=%s", code, stderr.String())
	}
	var payload map[string]interface{}
	if err := json.Unmarshal(stdout.Bytes(), &payload); err != nil {
		t.Fatalf("output is not JSON: %v", err)
	}
	if _, present := payload["inputs"]; present {
		t.Error("--no-verify emitted an inputs key; that path never parses a workflow")
	}
}

func TestRunInspect_JSONInputsEmptyArrayWhenNoneDeclared(t *testing.T) {
	bundle := packForTest(t) // the fixed minimalDip source declares no inputs
	var stdout, stderr bytes.Buffer
	if code := runInspect(&stdout, &stderr, []string{"--format=json", bundle}); code != exitDipxOK {
		t.Fatalf("exit code = %d; stderr=%s", code, stderr.String())
	}
	var payload struct {
		Inputs []interface{} `json:"inputs"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &payload); err != nil {
		t.Fatalf("output is not JSON: %v", err)
	}
	if payload.Inputs == nil {
		t.Error("inputs = null, want an empty array")
	}
	if len(payload.Inputs) != 0 {
		t.Errorf("got %d inputs, want 0", len(payload.Inputs))
	}
}
```

Confirm `bytes`, `encoding/json`, `os`, and `path/filepath` are imported in that file; add any that are missing.

- [ ] **Step 2: Run the test to verify it fails**

Run: `just test-pkg cmd/dippin`
Expected: FAIL — the JSON payload has no `inputs` key.

- [ ] **Step 3: Thread inputs through the JSON renderer**

In `cmd/dippin/cmd_inspect.go`, add the parameter to the shared renderer and emit the key only when non-nil:

```go
// printManifestJSON is the shared JSON renderer. Used by both Open-side
// and OpenManifest-side (--no-verify, Task 4) paths.
//
// inputs carries the entry workflow's declared input schema (#190) on the
// Open-side path. The --no-verify path reads the manifest only and never
// parses a workflow, so it passes nil and the key is omitted entirely —
// an "inputs": null there would read as "this bundle declares none".
func printManifestJSON(stdout, stderr io.Writer, m dipx.Manifest, id [32]byte, status InspectStatus, inputs []inputJSON) int {
	out := map[string]interface{}{
		"format_version": m.FormatVersion,
		"entry":          m.Entry,
		"identity":       "sha256:" + hex.EncodeToString(id[:]),
		"files":          m.Files,
		"status":         status,
	}
	if inputs != nil {
		out["inputs"] = inputs
	}
	enc := json.NewEncoder(stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(out); err != nil {
		fmt.Fprintln(stderr, err)
		return exitDipxIOError
	}
	return exitDipxOK
}
```

Update the two call sites:

```go
func printInspectJSON(stdout, stderr io.Writer, b *dipx.Bundle) int {
	m := b.Manifest()
	id := b.Identity()
	status := buildInspectStatus(m, b.ByteTotal(), false)
	return printManifestJSON(stdout, stderr, m, id, status, inputsJSON(b.Entry()))
}
```

and in `runInspectNoVerify`, pass `nil` as the final argument to `printManifestJSON`.

`inputsJSON` returns a non-nil empty slice for a workflow with no inputs, so the verify path always emits `"inputs": []` rather than null.

Leave `printInspectText` and `printManifestText` unchanged — this task adds the machine-readable surface only.

- [ ] **Step 4: Run the tests**

Run: `just test-pkg cmd/dippin`
Expected: PASS, including the pre-existing `TestRunInspect_JSON`, `TestRunInspect_NoVerifyEmitsVerifySkippedTrue`, and `TestRunInspect_JSONIsParseable`.

- [ ] **Step 5: Run the full suite and complexity**

Run: `just test && just complexity`
Expected: PASS, clean.

- [ ] **Step 6: Commit**

```bash
git add cmd/dippin/cmd_inspect.go cmd/dippin/cmd_inspect_test.go
git commit -m "$(cat <<'EOF'
feat(cli): surface entry-workflow inputs in dippin inspect (#190)

A host can enumerate what to collect from a .dipx without unpacking it.
The shared JSON renderer takes the schema as a parameter; the --no-verify
path passes nil and the key is omitted, since that path reads the
manifest only and never parses a workflow.

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

### Task 9: Grammar, docs, editors, site, skill, and a worked example

The standing rule in this repo: every language change lands with its documentation, site, skill, and editor surfaces swept in the same batch.

**Files:**
- Modify: `docs/GRAMMAR.ebnf`
- Modify: `docs/syntax.md`
- Modify: `docs/context.md`
- Modify: `docs/validation.md`
- Modify: `site/static/skill.md`
- Modify: `site/content/language.md`
- Modify: `site/static/highlight.js:42,120`
- Modify: `editors/tree-sitter-dippin/grammar.js:33-48`
- Modify: `editors/vscode/syntaxes/dippin.tmLanguage.json:48`
- Modify: `editors/zed-dippin/languages/dippin/highlights.scm:4-5`
- Create: `examples/pipeline_inputs.dip`
- Regenerated (by the pre-commit hook, do not hand-edit): `docs/generated-spec.md`, `cmd/dippin/generated-spec.md`

**Interfaces:**
- Consumes: everything from Tasks 1–8

- [ ] **Step 1: Write the worked example**

Create `examples/pipeline_inputs.dip`. Attribute order must match the formatter's canonical order so `dippin fmt` leaves it unchanged:

```dip
workflow IdeaToPR
  goal: "Turn a user's idea into a shipped PR."
  start: Plan
  exit: Done

  inputs
    idea: text
      required: true
      prompt: "What do you want built?"
      description: "One or two sentences describing the change."
      max_length: 4000
      multiline: true
    target_branch: text
      default: main
      pattern: "^[A-Za-z0-9._/-]+$"
    spec: file
      default: SPEC.md
    risk: enum
      default: medium
      options: low, medium, high

  agent Plan
    label: "Plan the change"
    prompt:
      Read ${inputs.spec} and plan how to implement this request.

      Request: ${inputs.idea}

      Target branch: ${inputs.target_branch}
      Risk tolerance: ${inputs.risk}

  agent Done
    label: "Summarize"
    prompt:
      Summarize what was planned.

  edges
    Plan -> Done
```

- [ ] **Step 2: Verify the example validates, lints clean, and is already formatted**

```bash
just build
./dippin validate examples/pipeline_inputs.dip
./dippin lint examples/pipeline_inputs.dip
./dippin fmt --check examples/pipeline_inputs.dip
./dippin inputs --format=json examples/pipeline_inputs.dip
```

Expected: validate and lint clean (exit 0, zero warnings), `fmt --check` reports no change, and the JSON lists four inputs in declaration order. If `fmt --check` reports a diff, reorder the example's attributes to the canonical order rather than changing the formatter.

- [ ] **Step 3: Update the grammar**

In `docs/GRAMMAR.ebnf`, add `inputs_section` to `workflow_body`:

```ebnf
workflow_body ::= ( workflow_field | defaults_section | vars_section | inputs_section
                  | node_decl | edges_section | stylesheet_section | NEWLINE )*
```

Then add a new section after the Vars Section block:

```ebnf
/* ============================================================ */
/* Inputs Section                                               */
/* ============================================================ */

/* The callee-side signature: what a caller must supply, whether a  */
/* human at the entry point or a parent workflow via a subgraph     */
/* node's params:. Declaration order is significant (a host renders */
/* an ordered form) so the formatter does not sort it. Values are   */
/* untrusted by construction and are read as ${inputs.name}.        */

inputs_section ::= "inputs" NEWLINE INDENT input_decl* OUTDENT

input_decl ::= IDENTIFIER ":" input_type NEWLINE ( INDENT input_field* OUTDENT )?

/* The named types are the v1 closed set. IDENTIFIER makes the      */
/* production forward-compatible: the parser accepts any type token */
/* and an unrecognized one is diagnosed by the validator (DIP155),  */
/* never by the parser.                                             */
input_type ::= "text" | "number" | "bool" | "enum" | "file" | "secret" | IDENTIFIER

input_field ::= "required" ":" BOOLEAN
              | "default" ":" field_value       /* form prefill; does not satisfy required */
              | "prompt" ":" field_value        /* what a host asks the caller */
              | "description" ":" field_value   /* help text */
              | "options" ":" field_value       /* enum: comma-separated choices */
              | "pattern" ":" field_value       /* text: regex the host enforces */
              | "min" ":" field_value           /* number: inclusive lower bound */
              | "max" ":" field_value           /* number: inclusive upper bound */
              | "max_length" ":" INTEGER        /* text: character cap */
              | "multiline" ":" BOOLEAN         /* text: host renders a textarea */
```

- [ ] **Step 4: Update the tree-sitter grammar**

In `editors/tree-sitter-dippin/grammar.js`, add `$.inputs_section` to the `workflow_body` choice list, then add the rule after `defaults_field`:

```js
    // ── Inputs ────────────────────────────────────────────────
    inputs_section: ($) =>
      seq("inputs", $._indent, repeat1(choice($.input_decl, $._newline)), $._dedent),

    input_decl: ($) =>
      seq(
        $.identifier,
        ":",
        $.field_value,
        optional(seq($._indent, repeat1(choice($.input_field, $._newline)), $._dedent))
      ),

    input_field: ($) => seq($.field_name, ":", $.field_value),
```

Note: `vars_section` is absent from this grammar today — a pre-existing gap. Do **not** fix it here; that is unrelated to this change.

- [ ] **Step 5: Update the editor highlighters**

`editors/vscode/syntaxes/dippin.tmLanguage.json` — add `inputs` to the section keyword alternation:

```json
      "match": "^\\s*(defaults|vars|inputs|edges)\\b",
```

`editors/zed-dippin/languages/dippin/highlights.scm` — add `"inputs"` to the keyword list alongside `"defaults"` and `"edges"`.

`site/static/highlight.js` — add `inputs` to both alternations, on line 42 and line 120:

```js
    h = h.replace(/\b(edges|defaults|vars|inputs|stylesheet)\b/g, function (_, k) { return s("kw", k); });
```

```js
    return /\b(workflow|agent|human|tool|subgraph|conditional|manager_loop|parallel|fan_in|edges|defaults|vars|inputs|stylesheet)\b/.test(t);
```

- [ ] **Step 6: Update the prose docs**

`docs/syntax.md` — add an `inputs` section modeled on the existing `vars` section, documenting the `name: type` form, the ten attributes, the six types, that declaration order is significant, and that `required` plus `default` means "prefill, still must be supplied."

`docs/context.md` — add an `inputs` entry to the Variable Namespaces section after `params`, stating: values are supplied by the caller and bound at run start; the namespace is closed, so a reference to an undeclared input is a lint error (DIP156) — unlike `ctx`; values are untrusted by construction and a host frames them as data rather than instructions; and `${inputs.x}` never interpolates inside a tool `command:` (DIP157).

`docs/validation.md` — add DIP155, DIP156, and DIP157 rows to the diagnostic table, all error severity.

`site/content/language.md` — add an `inputs` subsection near the existing `vars` coverage at line 114.

`site/static/skill.md` — add `inputs` to the File Structure section. **This file is a gen-spec source**: `docs/generated-spec.md` and `cmd/dippin/generated-spec.md` are assembled from `docs/llm-reference.md` plus this file by `scripts/gen-spec.sh`, which the pre-commit hook runs automatically. Never hand-edit the generated files.

- [ ] **Step 7: Confirm the LSP needs no change**

The spec's sweep list names the LSP, but `lsp/completion.go` offers only **node-field** completions (`prompt:`, `model:`, `command:` …) and has no section-keyword completion at all — `defaults`, `vars`, and `edges` are absent too. Input attributes are workflow-section-scoped, so adding them to `fieldCompletions()` would surface `required:` and `max_length:` inside every agent block, which is wrong.

Make no LSP change. Confirm the situation still holds by running `grep -n "vars\|defaults\|edges" lsp/completion.go` and verifying it returns nothing; if that grep now matches, section-keyword completion has since been added and `inputs` belongs in it.

- [ ] **Step 8: Verify every example still validates**

Run: `just validate-examples && just lint-examples`
Expected: all clean, including the new example.

- [ ] **Step 9: Run the full suite**

Run: `just test`
Expected: PASS. `TestLintExamples` parses every example through parse → lint and asserts zero DIP108 warnings; `TestEBNFOperatorsMatchParser` reads `docs/GRAMMAR.ebnf`.

- [ ] **Step 10: Commit**

The pre-commit hook regenerates the spec files; add them to the commit if it modifies them.

```bash
git add docs/ site/ editors/ examples/pipeline_inputs.dip cmd/dippin/generated-spec.md
git commit -m "$(cat <<'EOF'
docs(inputs): grammar, docs, site, editors, and worked example (#190)

Sweeps every language surface for the inputs block: GRAMMAR.ebnf,
syntax/context/validation docs, tree-sitter, VS Code, Zed, site
highlighting, the hosted skill, and an examples/ workflow that
validates and lints clean.

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

### Task 10: Changelog and PR

**Files:**
- Modify: `CHANGELOG.md`

- [ ] **Step 1: Add the changelog entry**

Read the top of `CHANGELOG.md` and match the established entry structure exactly (this repo distinguishes detection-only from runtime-paired entries). Add an Unreleased entry covering: the `inputs` block, the `${inputs.x}` closed namespace, DIP155/156/157, `dippin inputs`, entry-input surfacing in `dippin inspect`, and the `dippin lint` exit-code change for error-severity lints. Note that `site/content/changelog.md` is generated from this file — do not hand-edit it.

- [ ] **Step 2: Run the full gate**

Run: `just test && just validate-examples && just lint-examples && just complexity`
Expected: all PASS.

- [ ] **Step 3: Commit and push**

```bash
git add CHANGELOG.md
git commit -m "$(cat <<'EOF'
docs(changelog): pipeline inputs Phase 1 (#190)

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>
EOF
)"
git push -u origin feat/190-pipeline-inputs
```

- [ ] **Step 4: Open the PR**

```bash
gh pr create --title "feat: native inputs declaration — typed, introspectable input schema (#190)" --body "$(cat <<'EOF'
Phase 1 of pipeline inputs. Closes the grammar/IR/validate half of #190; pairs with tracker#553 (engine collect/validate/inject).

## What this adds

- **`inputs` block** — the callee-side signature declaring what a caller must supply, whether a human at the entry point or a parent workflow via `subgraph … params:`.
- **`${inputs.x}`** — the first *closed* namespace in the language. References resolve against the declaration, not just a valid prefix.
- **DIP155/156/157** — unknown type, undeclared reference (prompts and edge conditions), and input reference inside a tool `command:` (which never interpolates).
- **`dippin inputs <file> --format=json`** — the typed schema export a host uses to collect values before a run. `dippin inspect` surfaces a bundle's entry inputs too.

## Notable

- **`dippin lint` exit code changed.** DIP155–157 are the first error-severity lints; `CmdLint` previously exited non-zero only on structural `Validate` errors, so an error-severity lint printed and exited 0. It now fails on either.
- **Declaration order is preserved, not sorted** — a host renders these as an ordered form. This deliberately differs from `vars`.
- **Forward compatible by layering** — the parser never rejects an unknown type or attribute, so a `.dip` using a future type still parses, formats, and packs on an older dippin. Only the lint complains.

Design: `docs/superpowers/specs/2026-08-06-pipeline-inputs-design.md`
Plan: `docs/superpowers/plans/2026-08-06-pipeline-inputs-phase1.md`

Deferred to Phase 2/3: DIP158 (constraint consistency), DIP159 (dead input), DIP160 (cross-file subgraph arity), DIP161 (chain-attack integration).

🤖 Generated with [Claude Code](https://claude.com/claude-code)

https://claude.ai/code/session_01UU7HSXYSm6n1moA3RzCeUR
EOF
)"
```

Do **not** merge the PR. Merging and tagging are always explicit human decisions in this repo.

---

## Deferred to later phases

Not in scope for this plan, tracked in the spec:

- **DIP158** — invalid or inapplicable constraint (enum `default` not in `options`, `min` > `max`, malformed `pattern`, constraint on a type that has none).
- **DIP159** — declared input never referenced (dead input).
- **DIP160** — subgraph `params:` omits a required input of the referenced child (cross-file; belongs at the CLI layer with `crossfile_tool_access.go`, not inside `validator`).
- **DIP161** — untrusted input reaches a tool-bearing agent (DIP147 family).
