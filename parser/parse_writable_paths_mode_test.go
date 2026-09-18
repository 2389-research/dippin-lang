package parser

import (
	"strings"
	"testing"

	"github.com/2389-research/dippin-lang/ir"
)

// agentModeSrc builds a minimal workflow with one agent whose
// writable_paths_mode line is exactly modeLine (so callers control quoting
// and trailing whitespace precisely).
func agentModeSrc(modeLine string) string {
	return "workflow X\n  start: A\n  exit: A\n\n  agent A\n    prompt: \"x\"\n    writable_paths: workspace/**\n    " + modeLine + "\n"
}

// TestParseAgentWritablePathsModeVerbatim asserts the value reaches the IR
// verbatim — no case-folding, no trimming of what the tokenizer preserves.
// The single-line tokenizer (RawValueText) trims trailing whitespace on
// unquoted values, so a trailing space only survives inside quotes; the lint
// must see the exact spelling that survives.
func TestParseAgentWritablePathsModeVerbatim(t *testing.T) {
	cases := []struct {
		name, line, want string
	}{
		{"require", "writable_paths_mode: require", "require"},
		{"prefer", "writable_paths_mode: prefer", "prefer"},
		{"capitalized is kept", "writable_paths_mode: Prefer", "Prefer"},
		{"misspelling is kept", "writable_paths_mode: preferred", "preferred"},
		{"unquoted trailing space is trimmed by the tokenizer", "writable_paths_mode: prefer   ", "prefer"},
		{"quoted trailing space survives", `writable_paths_mode: "prefer "`, "prefer "},
		{"inline comment stripped", "writable_paths_mode: prefer  # degrade on macOS", "prefer"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			w, err := NewParser(agentModeSrc(tc.line), "test.dip").Parse()
			if err != nil {
				t.Fatalf("parse error: %v", err)
			}
			got := w.Node("A").Config.(ir.AgentConfig).WritablePathsMode
			if got != tc.want {
				t.Errorf("WritablePathsMode = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestParseAgentWritablePathsModeOmittedIsEmpty(t *testing.T) {
	src := `workflow X
  start: A
  exit: A

  agent A
    prompt: "x"
    writable_paths: workspace/**
`
	w, err := NewParser(src, "test.dip").Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if got := w.Node("A").Config.(ir.AgentConfig).WritablePathsMode; got != "" {
		t.Errorf("WritablePathsMode = %q, want empty (absent = require, not default-filled)", got)
	}
}

// TestParseAgentWritablePathsModeEmptyIsError: a present-but-empty
// writable_paths_mode: cannot be represented in the string IR field (it would
// collapse to absent), so the parser rejects it fail-closed, mirroring
// writable_paths.
func TestParseAgentWritablePathsModeEmptyIsError(t *testing.T) {
	cases := []struct{ name, line string }{
		{"bare", "writable_paths_mode:"},
		{"whitespace only", "writable_paths_mode:   "},
		{"empty quotes", `writable_paths_mode: ""`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := NewParser(agentModeSrc(tc.line), "test.dip").Parse()
			if err == nil {
				t.Fatal("expected parse error for empty writable_paths_mode:, got nil")
			}
			if !strings.Contains(err.Error(), "writable_paths_mode") {
				t.Errorf("error should name writable_paths_mode; got: %v", err)
			}
		})
	}
}

func branchModeSrc(modeLine string) string {
	return `workflow X
  start: split
  exit: join

  agent a
    prompt: "a"

  parallel split
    branch: a
      writable_paths: workspace/**
      ` + modeLine + `

  fan_in join <- a

  edges
    split -> a
    a -> join
`
}

func TestParseBranchWritablePathsModeVerbatim(t *testing.T) {
	cases := []struct {
		name, line, want string
	}{
		{"prefer", "writable_paths_mode: prefer", "prefer"},
		{"require", "writable_paths_mode: require", "require"},
		{"capitalized is kept", "writable_paths_mode: Prefer", "Prefer"},
		{"quoted trailing space survives", `writable_paths_mode: "prefer "`, "prefer "},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			w, err := NewParser(branchModeSrc(tc.line), "test.dip").Parse()
			if err != nil {
				t.Fatalf("parse error: %v", err)
			}
			cfg := w.Node("split").Config.(ir.ParallelConfig)
			if len(cfg.Branches) != 1 {
				t.Fatalf("branches = %d, want 1", len(cfg.Branches))
			}
			if got := cfg.Branches[0].WritablePathsMode; got != tc.want {
				t.Errorf("branch WritablePathsMode = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestParseBranchWritablePathsModeEmptyIsError(t *testing.T) {
	_, err := NewParser(branchModeSrc("writable_paths_mode:"), "test.dip").Parse()
	if err == nil {
		t.Fatal("expected parse error for empty branch writable_paths_mode:, got nil")
	}
	if !strings.Contains(err.Error(), "writable_paths_mode") {
		t.Errorf("error should name writable_paths_mode; got: %v", err)
	}
}

func TestParseBranchWritablePathsModeOmittedIsEmpty(t *testing.T) {
	w, err := NewParser(branchModeSrc("model: claude-sonnet-4-6"), "test.dip").Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	cfg := w.Node("split").Config.(ir.ParallelConfig)
	if got := cfg.Branches[0].WritablePathsMode; got != "" {
		t.Errorf("branch WritablePathsMode = %q, want empty (inherits target)", got)
	}
}
