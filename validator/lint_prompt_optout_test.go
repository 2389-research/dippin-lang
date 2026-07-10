package validator

import (
	"testing"

	"github.com/2389-research/dippin-lang/parser"
)

func TestLintPromptOptOut_NoCascade(t *testing.T) {
	src := `workflow W
  start: A
  exit: A
  agent A
    prompt: "x"
    prompt_suffix: none
`
	w, _ := parser.NewParser(src, "t.dip").Parse()
	if got := countCode(Lint(w).Diagnostics, DIP154); got != 1 {
		t.Fatalf("want 1 DIP154, got %d", got)
	}
}

func TestLintPromptOptOut_WithCascadeSilent(t *testing.T) {
	src := `workflow W
  start: A
  exit: A
  defaults
    prompt_suffix: "END"
  agent A
    prompt: "x"
    prompt_suffix: none
`
	w, _ := parser.NewParser(src, "t.dip").Parse()
	if got := countCode(Lint(w).Diagnostics, DIP154); got != 0 {
		t.Fatalf("want 0 DIP154 (cascade declared), got %d", got)
	}
}
