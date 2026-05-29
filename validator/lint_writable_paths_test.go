package validator

import (
	"testing"

	"github.com/2389-research/dippin-lang/parser"
)

func lintSrc(t *testing.T, src string) []Diagnostic {
	t.Helper()
	w, err := parser.NewParser(src, "test.dip").Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	return Lint(w).Diagnostics
}

func TestLint_DIP141_WritablePathsWithToolAccessNone(t *testing.T) {
	src := `workflow X
  start: A
  exit: A

  agent A
    prompt: "x"
    tool_access: none
    writable_paths: workspace/**
`
	if !hasCode(lintSrc(t, src), DIP141) {
		t.Errorf("expected DIP141, got: %v", codes(lintSrc(t, src)))
	}
}

func TestLint_DIP141_NotFiredWhenAlone(t *testing.T) {
	src := `workflow X
  start: A
  exit: A

  agent A
    prompt: "x"
    writable_paths: workspace/**
`
	if hasCode(lintSrc(t, src), DIP141) {
		t.Errorf("DIP141 should not fire without tool_access: none; got: %v", codes(lintSrc(t, src)))
	}
}
