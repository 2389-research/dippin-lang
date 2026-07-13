package validator

import (
	"testing"

	"github.com/2389-research/dippin-lang/parser"
)

// validateSrc parses src through the real parser and runs structural Validate.
// Mirrors lintSrc but exercises the parse-then-validate path (the contract under
// test for DIP010) — never hand-build IR with Condition.Parsed pre-set.
func validateSrc(t *testing.T, src string) []Diagnostic {
	t.Helper()
	w, err := parser.NewParser(src, "test.dip").Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	return Validate(w).Diagnostics
}

// TestDIP010_InvalidEdgeCondition is the issue #98 repro: an edge using a
// tool-node field (marker_grep) as an operator is unparseable and must surface
// as a structural error rather than passing validation silently.
func TestDIP010_InvalidEdgeCondition(t *testing.T) {
	src := `workflow Repro
  goal: "repro"
  start: A
  exit: Z

  tool A
    command: "echo ok"
  agent Z
    prompt: "done"

  edges
    A -> Z when marker_grep "^ok"
`
	diags := validateSrc(t, src)
	if !hasCode(diags, "DIP010") {
		t.Fatalf("expected DIP010 for unparseable edge condition, got: %v", codes(diags))
	}
}

// cascadeSrc has an unparseable edge (B -> C) that appears BEFORE a valid edge
// (D -> E) which references an undefined namespace (badns) and so should trip
// DIP120. Pre-fix, EnsureConditionsParsed bails at B -> C and D -> E keeps
// Parsed == nil, silently suppressing DIP120. The fix populates every parseable
// edge regardless of earlier failures.
const cascadeSrc = `workflow W
  goal: "cascade"
  start: A
  exit: E

  agent A
    prompt: "a"
  agent B
    prompt: "b"
  agent C
    prompt: "c"
  agent D
    prompt: "d"
  agent E
    prompt: "e"

  edges
    A -> B
    B -> C when marker_grep "^ok"
    A -> D
    D -> E when badns.flavor = vanilla
`

// TestDIP010_BadEdgeDoesNotMaskLaterLint proves the cascade fix from Lint alone:
// a downstream AST-dependent lint (DIP120) still fires even though an earlier
// edge is unparseable.
func TestDIP010_BadEdgeDoesNotMaskLaterLint(t *testing.T) {
	diags := lintSrc(t, cascadeSrc)
	if !hasCode(diags, DIP120) {
		t.Fatalf("expected DIP120 on the later valid edge despite an earlier unparseable edge, got: %v", codes(diags))
	}
}

// TestDIP010_OnePerBadEdge verifies accumulate-all: every unparseable edge gets
// its own diagnostic rather than stopping at the first.
func TestDIP010_OnePerBadEdge(t *testing.T) {
	src := `workflow W
  goal: "two bad"
  start: A
  exit: C

  agent A
    prompt: "a"
  agent B
    prompt: "b"
  agent C
    prompt: "c"

  edges
    A -> B when marker_grep "^ok"
    B -> C when another_grep "^no"
`
	diags := validateSrc(t, src)
	if n := countCode(diags, "DIP010"); n != 2 {
		t.Fatalf("expected 2 DIP010 (one per bad edge), got %d: %v", n, codes(diags))
	}
}

// TestDIP010_ValidConditionsClean confirms no false positives on valid conditions.
func TestDIP010_ValidConditionsClean(t *testing.T) {
	src := `workflow W
  goal: "valid"
  start: A
  exit: C

  agent A
    prompt: "a"
  agent B
    prompt: "b"
  agent C
    prompt: "c"

  edges
    A -> B when ctx.outcome = success
    B -> C when ctx.score contains high
`
	diags := validateSrc(t, src)
	if hasCode(diags, "DIP010") {
		t.Fatalf("unexpected DIP010 on valid conditions: %v", codes(diags))
	}
}

// TestParseEdgeConditions_PopulatesPastFailure is the unit-level proof: good
// edges get Parsed populated even when an earlier edge fails, and each failure
// is reported once.
func TestParseEdgeConditions_PopulatesPastFailure(t *testing.T) {
	w, err := parser.NewParser(cascadeSrc, "test.dip").Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	failures := parseEdgeConditions(w)
	if len(failures) != 1 {
		t.Fatalf("expected 1 failure (B -> C), got %d", len(failures))
	}
	for _, e := range w.Edges {
		if e.From == "D" && e.To == "E" {
			if e.Condition == nil || e.Condition.Parsed == nil {
				t.Fatalf("expected D -> E to be parsed despite the earlier bad edge")
			}
		}
		if e.From == "B" && e.To == "C" {
			if e.Condition != nil && e.Condition.Parsed != nil {
				t.Fatalf("expected the unparseable B -> C to keep Parsed == nil")
			}
		}
	}
}

// TestDIP010_EdgesOnly confirms the scope boundary: an unparseable manager_loop
// node condition does NOT emit DIP010 (DIP010 covers edge conditions only).
func TestDIP010_EdgesOnly(t *testing.T) {
	src := `workflow M
  goal: "scope"
  start: S
  exit: D

  manager_loop S
    subgraph_ref: child.dip
    stop_condition: marker_grep "^ok"
  agent D
    prompt: "done"

  edges
    S -> D
`
	diags := validateSrc(t, src)
	if hasCode(diags, "DIP010") {
		t.Fatalf("DIP010 should not fire for manager_loop node conditions: %v", codes(diags))
	}
}

// Salvaged from PR #183 (thanks @harperreed): an escaped-quote edge condition
// parses cleanly through parse → Validate with a populated AST (#182).
func TestDIP010_EscapedQuoteConditionParsesFromSource(t *testing.T) {
	src := `workflow W
  goal: "escaped quote"
  start: A
  exit: B

  agent A
    prompt: "a"
  agent B
    prompt: "b"

  edges
    A -> B when ctx.tool_stdout = "say \"alpha||beta\""
`
	if diags := validateSrc(t, src); len(diags) != 0 {
		t.Fatalf("unexpected validation diagnostics: %v", diags)
	}
}
