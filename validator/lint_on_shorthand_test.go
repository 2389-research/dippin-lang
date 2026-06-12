package validator_test

import (
	"testing"

	"github.com/2389-research/dippin-lang/parser"
	"github.com/2389-research/dippin-lang/validator"
)

// TestLintReasonsOverOnShorthand proves the `on` edge shorthand feeds the
// condition lints identically to the equivalent `when`: it desugars to a real
// ir.Condition, so the AST-dependent and presence-dependent checks reason over
// it. Two `on success` edges from one node overlap (DIP103); a lone conditional
// `on` edge with no fallback is a routing node without a default (DIP102).
//
// Tests parse real .dip text through the production parser, per CLAUDE.md.
func TestLintReasonsOverOnShorthand(t *testing.T) {
	dup := `workflow Dup
  start: A
  exit: C

  agent A
    prompt: "a"

  agent B
    prompt: "b"

  agent C
    prompt: "c"

  edges
    A -> B  on success
    A -> C  on success
`
	if !lintHasCode(t, dup, validator.DIP103) {
		t.Errorf("expected DIP103 for duplicate `on success` edges")
	}

	noFallback := `workflow NoFallback
  start: A
  exit: C

  agent A
    prompt: "a"

  agent B
    prompt: "b"

  agent C
    prompt: "c"

  edges
    A -> B  on success
    B -> C
`
	if !lintHasCode(t, noFallback, validator.DIP102) {
		t.Errorf("expected DIP102: a conditional `on` edge with no fallback is a routing node without a default")
	}
}

func lintHasCode(t *testing.T, src, code string) bool {
	t.Helper()
	w, err := parser.NewParser(src, "test.dip").Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	for _, d := range validator.Lint(w).Diagnostics {
		if d.Code == code {
			return true
		}
	}
	return false
}
