package validator

import (
	"testing"

	"github.com/2389-research/dippin-lang/parser"
)

func lintDiags(t *testing.T, src string) []Diagnostic {
	t.Helper()
	w, err := parser.NewParser(src, "test.dip").Parse()
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	return Lint(w).Diagnostics
}

func TestLintRedundantFanEdge(t *testing.T) {
	src := `workflow W
  start: Fan
  exit: Done
  parallel Fan -> A, B
  agent A
    prompt: "a"
  agent B
    prompt: "b"
  fan_in Join <- A, B
  agent Done
    prompt: "d"
  edges
    Fan -> A
    Fan -> B
    A -> Join
    B -> Join
    Join -> Done
`
	if got := countCode(lintDiags(t, src), DIP153); got != 4 {
		t.Fatalf("want 4 DIP153 (Fan->A, Fan->B, A->Join, B->Join), got %d", got)
	}
}

func TestLintRedundantFanEdge_ConditionalExempt(t *testing.T) {
	src := `workflow W
  start: Fan
  exit: A
  parallel Fan -> A, B
  agent A
    prompt: "a"
  agent B
    prompt: "b"
  edges
    Fan -> A when ctx.x = 1
    Fan -> B
`
	// Only the plain Fan->B is redundant; the conditional Fan->A is kept.
	if got := countCode(lintDiags(t, src), DIP153); got != 1 {
		t.Fatalf("want 1 DIP153, got %d", got)
	}
}

func TestLintRedundantFanEdge_NoFanNoWarn(t *testing.T) {
	src := `workflow W
  start: A
  exit: B
  agent A
    prompt: "a"
  agent B
    prompt: "b"
  edges
    A -> B
`
	if got := countCode(lintDiags(t, src), DIP153); got != 0 {
		t.Fatalf("want 0 DIP153, got %d", got)
	}
}
