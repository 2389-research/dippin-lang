package validator

import (
	"os"
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

func TestLint_DIP142_UnsafeEntries(t *testing.T) {
	cases := []struct {
		name, entry string
	}{
		{"absolute", "/etc/**"},
		{"home", "~/secrets/**"},
		{"windows drive", `C:\Users\x`},
		{"parent escape", "../../etc/**"},
		{"brace mis-split", "*.{md"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			src := "workflow X\n  start: A\n  exit: A\n\n  agent A\n    prompt: \"x\"\n    writable_paths: " + tc.entry + "\n"
			if !hasCode(lintSrc(t, src), DIP142) {
				t.Errorf("expected DIP142 for %q; got: %v", tc.entry, codes(lintSrc(t, src)))
			}
		})
	}
}

func TestLint_DIP142_SafeEntries(t *testing.T) {
	src := `workflow X
  start: A
  exit: A

  agent A
    prompt: "x"
    writable_paths: workspace/**, .ai/sprints/**, .ai/managers/recovery-journal.md
`
	if hasCode(lintSrc(t, src), DIP142) {
		t.Errorf("DIP142 should not fire for relative globs; got: %v", codes(lintSrc(t, src)))
	}
}

func TestLint_DIP141_Branch(t *testing.T) {
	src := `workflow X
  start: split
  exit: join

  agent a
    prompt: "a"

  parallel split
    branch: a
      writable_paths: workspace/**
      tool_access: none

  fan_in join <- a

  edges
    split -> a
    a -> join
`
	if !hasCode(lintSrc(t, src), DIP141) {
		t.Errorf("expected DIP141 on branch; got: %v", codes(lintSrc(t, src)))
	}
}

func TestLint_DIP142_Branch(t *testing.T) {
	src := `workflow X
  start: split
  exit: join

  agent a
    prompt: "a"

  parallel split
    branch: a
      writable_paths: /etc/**

  fan_in join <- a

  edges
    split -> a
    a -> join
`
	if !hasCode(lintSrc(t, src), DIP142) {
		t.Errorf("expected DIP142 on branch; got: %v", codes(lintSrc(t, src)))
	}
}

func TestExampleAgentWritablePathsLintsClean(t *testing.T) {
	data, err := os.ReadFile("../examples/agent_writable_paths.dip")
	if err != nil {
		t.Fatalf("read example: %v", err)
	}
	diags := lintSrc(t, string(data))
	if hasCode(diags, DIP141) || hasCode(diags, DIP142) {
		t.Errorf("example should be DIP141/DIP142-clean; got: %v", codes(diags))
	}
}
