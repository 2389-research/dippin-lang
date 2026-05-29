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

// TestParseAgentWritablePathsEmptyIsError replaces the old EmptyIsNil test.
// A present-but-empty writable_paths: is now rejected at parse time (fail-closed)
// because pack_shadow re-formats through IR and drops nil WritablePaths, so a
// present-but-empty value would silently become absent (unbounded) after pack.
func TestParseAgentWritablePathsEmptyIsError(t *testing.T) {
	src := `workflow X
  start: A
  exit: A

  agent A
    prompt: "x"
    writable_paths:
`
	_, err := NewParser(src, "test.dip").Parse()
	if err == nil {
		t.Error("expected parse error for empty writable_paths:, got nil")
	}
}

func TestParseAgentWritablePathsWhitespaceIsError(t *testing.T) {
	src := `workflow X
  start: A
  exit: A

  agent A
    prompt: "x"
    writable_paths:
`
	_, err := NewParser(src, "test.dip").Parse()
	if err == nil {
		t.Error("expected parse error for whitespace-only writable_paths:, got nil")
	}
}

func TestParseAgentWritablePathsOmittedIsOK(t *testing.T) {
	src := `workflow X
  start: A
  exit: A

  agent A
    prompt: "x"
`
	w, err := NewParser(src, "test.dip").Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	cfg := w.Node("A").Config.(ir.AgentConfig)
	if cfg.WritablePaths != nil {
		t.Errorf("omitted writable_paths should be nil, got %v", cfg.WritablePaths)
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

func TestParseBranchWritablePathsEmptyIsError(t *testing.T) {
	src := `workflow X
  start: split
  exit: join

  agent a
    prompt: "a"

  parallel split
    branch: a
      writable_paths:

  fan_in join <- a

  edges
    split -> a
    a -> join
`
	_, err := NewParser(src, "test.dip").Parse()
	if err == nil {
		t.Error("expected parse error for empty branch writable_paths:, got nil")
	}
}

func TestParseBranchWritablePathsOmittedIsOK(t *testing.T) {
	src := `workflow X
  start: split
  exit: join

  agent a
    prompt: "a"

  parallel split
    branch: a
      model: claude-sonnet-4-6

  fan_in join <- a

  edges
    split -> a
    a -> join
`
	w, err := NewParser(src, "test.dip").Parse()
	if err != nil {
		t.Fatalf("parse error (branch omitting writable_paths should be fine): %v", err)
	}
	cfg := w.Node("split").Config.(ir.ParallelConfig)
	if len(cfg.Branches) != 1 {
		t.Fatalf("branches = %d, want 1", len(cfg.Branches))
	}
	if cfg.Branches[0].WritablePaths != nil {
		t.Errorf("omitted branch writable_paths should be nil (inherit agent's), got %v", cfg.Branches[0].WritablePaths)
	}
}

// TestParseBranchUnknownFieldIsError verifies that a branch-block typo like
// writable_path (missing s) is rejected via the unknown-field hint path (FIX B).
func TestParseBranchUnknownFieldIsError(t *testing.T) {
	src := `workflow X
  start: split
  exit: join

  agent a
    prompt: "a"

  parallel split
    branch: a
      writable_path: workspace/**

  fan_in join <- a

  edges
    split -> a
    a -> join
`
	_, err := NewParser(src, "test.dip").Parse()
	if err == nil {
		t.Error("expected parse error for unknown branch field 'writable_path', got nil")
	}
}

// TestParseAgentWritablePathsCommaOnlyIsError verifies that a comma-only value
// (e.g. "writable_paths: ,") is rejected at parse time.  The old guard only
// checked TrimSpace == "", so "," slipped through and silently produced nil
// WritablePaths — the same fail-open hole as a blank value.
func TestParseAgentWritablePathsCommaOnlyIsError(t *testing.T) {
	src := `workflow X
  start: A
  exit: A

  agent A
    prompt: "x"
    writable_paths: ,
`
	_, err := NewParser(src, "test.dip").Parse()
	if err == nil {
		t.Error("expected parse error for comma-only writable_paths: ,, got nil")
	}
}

// TestParseAgentWritablePathsMultiCommaOnlyIsError verifies that ", ," (multiple
// commas, all empty entries) is also rejected.
func TestParseAgentWritablePathsMultiCommaOnlyIsError(t *testing.T) {
	src := `workflow X
  start: A
  exit: A

  agent A
    prompt: "x"
    writable_paths: , ,
`
	_, err := NewParser(src, "test.dip").Parse()
	if err == nil {
		t.Error("expected parse error for comma-only writable_paths: , ,, got nil")
	}
}

// TestParseBranchWritablePathsCommaOnlyIsError verifies the same fail-closed fix
// applies to branch-level writable_paths.
func TestParseBranchWritablePathsCommaOnlyIsError(t *testing.T) {
	src := `workflow X
  start: split
  exit: join

  agent a
    prompt: "a"

  parallel split
    branch: a
      writable_paths: ,

  fan_in join <- a

  edges
    split -> a
    a -> join
`
	_, err := NewParser(src, "test.dip").Parse()
	if err == nil {
		t.Error("expected parse error for comma-only branch writable_paths: ,, got nil")
	}
}

// TestParseBranchWritablePathsSpacedCommaOnlyIsError verifies "  ,  ," is rejected.
func TestParseBranchWritablePathsSpacedCommaOnlyIsError(t *testing.T) {
	src := `workflow X
  start: split
  exit: join

  agent a
    prompt: "a"

  parallel split
    branch: a
      writable_paths: , ,

  fan_in join <- a

  edges
    split -> a
    a -> join
`
	_, err := NewParser(src, "test.dip").Parse()
	if err == nil {
		t.Error("expected parse error for comma-only branch writable_paths: , ,, got nil")
	}
}

// TestParseBranchValidFieldsParseClean confirms known branch fields still parse correctly.
func TestParseBranchValidFieldsParseClean(t *testing.T) {
	src := `workflow X
  start: split
  exit: join

  agent a
    prompt: "a"

  parallel split
    branch: a
      model: claude-sonnet-4-6
      provider: anthropic
      fidelity: high
      tool_access: none
      writable_paths: workspace/**

  fan_in join <- a

  edges
    split -> a
    a -> join
`
	w, err := NewParser(src, "test.dip").Parse()
	if err != nil {
		t.Fatalf("parse error for valid branch fields: %v", err)
	}
	cfg := w.Node("split").Config.(ir.ParallelConfig)
	if len(cfg.Branches) != 1 {
		t.Fatalf("branches = %d, want 1", len(cfg.Branches))
	}
	b := cfg.Branches[0]
	if b.Model != "claude-sonnet-4-6" || b.Provider != "anthropic" || b.Fidelity != "high" || b.ToolAccess != "none" {
		t.Errorf("branch fields not parsed correctly: %+v", b)
	}
	if len(b.WritablePaths) != 1 || b.WritablePaths[0] != "workspace/**" {
		t.Errorf("branch WritablePaths = %v, want [workspace/**]", b.WritablePaths)
	}
}
