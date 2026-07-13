# #182 Edge-condition quote handling — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development or executing-plans. Steps use checkbox (`- [ ]`) syntax.

**Goal:** Preserve escaped double quotes in edge-condition `Condition.Raw` (round-tripping through reconstruct + re-tokenize) and reject unmatched double quotes with a source location.

**Architecture:** Shared double-quote escape convention across the reconstruct side (`parser/parse_edges.go`) and the condition re-tokenizer (`simulate/condition.go`), plus a new lexer error list (`parser/lexer.go`) surfaced by `Parse()`.

**Tech Stack:** Go; `just`; pre-commit gate (`export PATH="/usr/local/go/bin:$PATH"`).

**Spec:** `docs/superpowers/specs/2026-07-13-issue-182-edge-condition-quotes-design.md`

---

### Task 1: Escape literals on reconstruct

**Files:** Modify `parser/parse_edges.go` (`formatConditionToken`); Test `parser/parse_edges_test.go`.

- [ ] **Step 1: Failing test.** An edge condition with interior escaped quotes reconstructs `Condition.Raw` with the escapes preserved.
```go
func TestConditionRaw_PreservesEscapedQuotes(t *testing.T) {
	src := `workflow W
  start: A
  exit: B
  tool A
    command: "echo hi"
  agent B
    prompt: "x"
  edges
    A -> B when ctx.tool_stdout = "say \"alpha||beta\""
`
	w, err := NewParser(src, "t.dip").Parse()
	if err != nil { t.Fatal(err) }
	got := w.Edges[0].Condition.Raw
	want := `ctx.tool_stdout = "say \"alpha||beta\""`
	if got != want { t.Fatalf("Raw = %q, want %q", got, want) }
}
```
- [ ] **Step 2: Run → FAIL** (`Raw` has unescaped interior quotes). `go test ./parser/ -run TestConditionRaw_PreservesEscapedQuotes`
- [ ] **Step 3: Implement.** In `parser/parse_edges.go`:
```go
func formatConditionToken(t Token) string {
	if t.Type == TokenLiteral {
		return `"` + escapeConditionLiteral(t.Value) + `"`
	}
	return t.Value
}

// escapeConditionLiteral is the inverse of the lexer's appendQuotedChar:
// backslash first, then double-quote, so the wrapped literal round-trips.
func escapeConditionLiteral(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	return s
}
```
(Confirm `strings` is imported in parse_edges.go — it is, used by `readConditionRaw`.)
- [ ] **Step 4: Run → PASS.**
- [ ] **Step 5: Commit.** `git commit -m "fix(parser): escape interior quotes when reconstructing condition raw (#182)"`

---

### Task 2: Honor escapes on condition re-tokenize

**Files:** Modify `simulate/condition.go` (`tryTokenizeQuotedCond`); Test `simulate/condition_test.go`.

- [ ] **Step 1: Failing test.** `ParseCondition` on the escaped Raw recovers the literal value with interior quotes restored.
```go
func TestParseCondition_EscapedQuotes(t *testing.T) {
	expr, err := ParseCondition(`ctx.tool_stdout = "say \"alpha||beta\""`)
	if err != nil { t.Fatalf("parse: %v", err) }
	cmp, ok := expr.(ir.CondCompare)
	if !ok { t.Fatalf("want CondCompare, got %T", expr) }
	if cmp.Value != `say "alpha||beta"` { t.Fatalf("Value = %q", cmp.Value) }
}
```
- [ ] **Step 2: Run → FAIL** (tokenizer stops at the first interior quote). `go test ./simulate/ -run TestParseCondition_EscapedQuotes`
- [ ] **Step 3: Implement.** Split `tryTokenizeQuotedCond` by quote kind; double quotes become escape-aware, single quotes keep existing behavior:
```go
func tryTokenizeQuotedCond(raw string, i int) (token string, consumed int) {
	if !isQuoteChar(raw[i]) {
		return "", 0
	}
	if raw[i] == '"' {
		return readDoubleQuotedCond(raw, i)
	}
	return readSingleQuotedCond(raw, i)
}

// readDoubleQuotedCond reads a double-quoted value from raw[i] (raw[i]=='"'),
// collapsing \" and \\ to their literal chars, until an unescaped closing ".
// Returns the unescaped content and total bytes consumed (incl. both quotes).
func readDoubleQuotedCond(raw string, i int) (string, int) {
	var b strings.Builder
	j := i + 1
	for j < len(raw) && raw[j] != '"' {
		if raw[j] == '\\' && j+1 < len(raw) {
			b.WriteByte(raw[j+1])
			j += 2
			continue
		}
		b.WriteByte(raw[j])
		j++
	}
	if j < len(raw) {
		j++ // closing quote
	}
	return b.String(), j - i
}

// readSingleQuotedCond preserves the existing scan-to-quote behavior (no escapes).
func readSingleQuotedCond(raw string, i int) (string, int) {
	quote := raw[i]
	start := i + 1
	j := start
	for j < len(raw) && raw[j] != quote {
		j++
	}
	token := raw[start:j]
	if j < len(raw) {
		j++
	}
	return token, j - i
}
```
Keep the existing `consumed` accounting equivalent to the old `i - (start - 1)` (i.e. bytes from the opening quote through the closing quote). Verify `strings` is imported in simulate/condition.go.
- [ ] **Step 4: Run → PASS + full simulate suite.** `go test ./simulate/`
- [ ] **Step 5: Commit.** `git commit -m "fix(simulate): honor backslash escapes in double-quoted condition values (#182)"`

---

### Task 3: End-to-end escaped-quote round-trip

**Files:** Test only — `parser/parse_edges_test.go` (or a new integration test).

- [ ] **Step 1: Failing/guard test.** The escaped-quote `.dip` validates (no DIP010) and `Format` round-trips.
```go
func TestEscapedQuoteCondition_ValidatesAndRoundTrips(t *testing.T) {
	src := `workflow W
  start: A
  exit: B
  tool A
    command: "echo hi"
  agent B
    prompt: "x"
  edges
    A -> B when ctx.tool_stdout = "say \"alpha||beta\""
`
	w, err := NewParser(src, "t.dip").Parse()
	if err != nil { t.Fatalf("parse: %v", err) }
	// EnsureConditionsParsed must succeed (the DIP010 path).
	if err := simulate.EnsureConditionsParsed(w); err != nil {
		t.Fatalf("condition parse: %v", err)
	}
	// Format → reparse identical.
	out := formatter.Format(w)
	w2, err := NewParser(out, "t.dip").Parse()
	if err != nil { t.Fatalf("reparse: %v\n%s", err, out) }
	if formatter.Format(w2) != out { t.Fatal("not idempotent") }
}
```
(Import cycle check: if `parser` test importing `simulate`+`formatter` creates a cycle, put this test in `package parser_test` or under `cmd/dippin`. Use whatever the existing cross-package parser tests use — `roundtrip_test.go` is `package parser_test` and imports both, so follow that.)
- [ ] **Step 2: Run → PASS** (Tasks 1+2 make it pass; if it fails, the round-trip convention is off — fix before proceeding). `go test ./parser/ -run TestEscapedQuoteCondition`
- [ ] **Step 3: Commit.** `git commit -m "test(parser): escaped-quote condition validates + round-trips (#182)"`

---

### Task 4: Reject unmatched double quotes (lexer + surface)

**Files:** Modify `parser/lexer.go` (Lexer struct, `readQuotedContent`, `tryLexQuotedString`), `parser/parser.go` (`Parse`); Test `parser/lexer_test.go` + `parser/parser_test.go`.

- [ ] **Step 1: Failing test.** An unterminated double quote is rejected with a location.
```go
func TestParse_RejectsUnterminatedDoubleQuote(t *testing.T) {
	src := `workflow W
  start: A
  exit: B
  tool A
    command: "echo hi"
  agent B
    prompt: "x"
  edges
    A -> B when ctx.tool_stdout = "alpha||beta
`
	_, err := NewParser(src, "t.dip").Parse()
	if err == nil || !strings.Contains(err.Error(), "unterminated string") {
		t.Fatalf("want unterminated-string rejection, got %v", err)
	}
}
```
- [ ] **Step 2: Run → FAIL** (currently "validation passed"). `go test ./parser/ -run TestParse_RejectsUnterminatedDoubleQuote`
- [ ] **Step 3: Implement.**
  - `parser/lexer.go`: add to `Lexer` struct a field `errs []string` (or a `[]lexError{Msg string; Loc ir.SourceLocation}` if you want structured output — string with `L:C` embedded is sufficient and matches other diagnostics). Change `readQuotedContent` to also return `terminated bool`:
    ```go
    func readQuotedContent(line string, start int) (string, int, bool) {
        ...
        terminated := i < len(line) // true iff we stopped on a closing quote
        if terminated { i++ }
        return content.String(), i, terminated
    }
    ```
  - In `tryLexQuotedString`, for the `'"'` case, capture `terminated` and record an error when false:
    ```go
    content, newI, terminated := readQuotedContent(line, i+1)
    if !terminated {
        l.errs = append(l.errs, fmt.Sprintf("unterminated string literal at %d:%d", loc.Line, loc.Column))
    }
    l.tokens = append(l.tokens, Token{Type: TokenLiteral, Value: content, Location: loc})
    return newI, true
    ```
    (Add a `Errors() []string` accessor. Leave the single-quote branch unchanged.)
  - `parser/parser.go:Parse()`: fold lexer errors in before returning. Lexing happens at `NewLexer` time, so the errors already exist:
    ```go
    func (p *Parser) Parse() (*ir.Workflow, error) {
        p.diagnostics = append(p.diagnostics, p.lexer.Errors()...)
        p.parseVersionDeclaration()
        p.parseTopLevel()
        p.rejectRedundantFanEdgesUnderV2()
        if len(p.diagnostics) > 0 {
            return p.workflow, fmt.Errorf("parsing errors: %s", strings.Join(p.diagnostics, "; "))
        }
        return p.workflow, nil
    }
    ```
    Confirm `fmt` is imported in lexer.go (it is, used for other formatting).
- [ ] **Step 4: Run → PASS + full parser suite** (ensure no existing valid file now errors — a properly-terminated quote sets `terminated=true`, no error). `go test ./parser/ ./...`
- [ ] **Step 5: Commit.** `git commit -m "fix(lexer): reject unterminated double-quoted string with a source location (#182)"`

---

### Task 5: Single-quote preservation + integration guards

**Files:** Test — `parser/parse_edges_test.go` / `simulate/condition_test.go`.

- [ ] **Step 1: Add guard tests.**
  - A single-quoted condition value (e.g. `A -> B when ctx.reason = 'needs review'`) still tokenizes and validates unchanged (`Condition.Raw` and `CondCompare.Value` as before this change).
  - A single-quoted value containing a double quote (`'say "hi"'`) is unaffected by the double-quote escape logic.
- [ ] **Step 2: Run → PASS.** `go test ./parser/ ./simulate/`
- [ ] **Step 3: Full gate.** `export PATH="/usr/local/go/bin:$PATH"; .git/hooks/pre-commit`
- [ ] **Step 4: Commit.** `git commit -m "test: single-quote condition behavior preserved (#182)"`

---

## No docs/grammar/catalog changes
This is a correctness fix — no new syntax, no new DIP code, no grammar change. The CHANGELOG entry is added at release time. (If any user-facing behavior note is warranted, it's "unterminated quotes are now a parse error" — mention in the release notes.)

## Final verification + squad review
```bash
export PATH="/usr/local/go/bin:$PATH"; .git/hooks/pre-commit
```
Then squad review: escape round-trip correctness (does every `TokenLiteral` value survive reconstruct→re-tokenize byte-identically, incl. backslashes, `||`, empty string, a value that is just `"`); unterminated detection false-positive check (no valid terminated string is flagged); single-quote regression; and re-run the two issue reproductions end-to-end via the built binary.

## Self-review notes
- Convention symmetry: `escapeConditionLiteral` (backslash then quote) is the exact inverse of `appendQuotedChar` / `readDoubleQuotedCond` (which collapse `\x`→`x`). Order verified.
- `consumed` accounting in Task 2 returns bytes from opening quote through closing quote — same span the old `i - (start - 1)` produced.
- Blast radius of Task 4 is bounded: only an *unterminated* `"` newly errors; terminated strings are byte-for-byte unchanged.
