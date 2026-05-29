package parser

import (
	"testing"

	"github.com/2389-research/dippin-lang/ir"
)

func TestParseAgentWritablePaths(t *testing.T) {
	src := `workflow X
  start: A
  exit: A

  agent A
    prompt: "x"
    writable_paths: workspace/**, .ai/sprints/**
`
	w, err := NewParser(src, "test.dip").Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	cfg := w.Node("A").Config.(ir.AgentConfig)
	want := []string{"workspace/**", ".ai/sprints/**"}
	if len(cfg.WritablePaths) != 2 || cfg.WritablePaths[0] != want[0] || cfg.WritablePaths[1] != want[1] {
		t.Errorf("WritablePaths = %v, want %v", cfg.WritablePaths, want)
	}
}

func TestParseAgentWritablePathsEmptyIsNil(t *testing.T) {
	src := `workflow X
  start: A
  exit: A

  agent A
    prompt: "x"
    writable_paths:
`
	w, err := NewParser(src, "test.dip").Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	cfg := w.Node("A").Config.(ir.AgentConfig)
	if cfg.WritablePaths != nil {
		t.Errorf("bare writable_paths should be nil (fail-closed at tracker), got %v", cfg.WritablePaths)
	}
}

func TestParseBranchWritablePaths(t *testing.T) {
	src := `workflow X
  start: split
  exit: join

  agent a
    prompt: "a"

  parallel split
    branch: a
      writable_paths: workspace/**, .ai/**

  fan_in join <- a

  edges
    split -> a
    a -> join
`
	w, err := NewParser(src, "test.dip").Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	cfg := w.Node("split").Config.(ir.ParallelConfig)
	if len(cfg.Branches) != 1 {
		t.Fatalf("branches = %d, want 1", len(cfg.Branches))
	}
	got := cfg.Branches[0].WritablePaths
	if len(got) != 2 || got[0] != "workspace/**" || got[1] != ".ai/**" {
		t.Errorf("branch WritablePaths = %v, want [workspace/** .ai/**]", got)
	}
}
