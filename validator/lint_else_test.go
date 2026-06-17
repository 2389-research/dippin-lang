package validator_test

import (
	"testing"

	"github.com/2389-research/dippin-lang/parser"
	"github.com/2389-research/dippin-lang/validator"
)

// TestLintElseSuppressesDefaultAndReachability verifies that a section-level
// `else -> X` default makes DIP102 (no unconditional default) and DIP101
// (reachable only via conditional edges) stand down: the else default covers
// any node whose guards do not match, so those nodes are not stuck and their
// conditional-only targets are intentional.
//
// The same shape WITHOUT an else still fires both, guarding against
// over-suppression.
func TestLintElseSuppressesDefaultAndReachability(t *testing.T) {
	withElse := `workflow ElseDefault
  start: T
  exit: D
  tool T
    command: echo hi
  agent D
    prompt: "done"
  agent Cleanup
    prompt: "clean"
  edges
    T -> D when ctx.tool_marker = go
    else -> Cleanup
`
	withoutElse := `workflow NoElse
  start: T
  exit: D
  tool T
    command: echo hi
  agent D
    prompt: "done"
  edges
    T -> D when ctx.tool_marker = go
`
	if codes := lintCodes(t, withElse); hasAny(codes, validator.DIP101, validator.DIP102) {
		t.Fatalf("else should suppress DIP101/DIP102, but got: %v", codes)
	}
	if codes := lintCodes(t, withoutElse); !hasAll(codes, validator.DIP101, validator.DIP102) {
		t.Fatalf("without else, expected DIP101 and DIP102, got: %v", codes)
	}
}

func lintCodes(t *testing.T, src string) []string {
	t.Helper()
	p := parser.NewParser(src, "test.dip")
	w, err := p.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	var codes []string
	for _, d := range validator.Lint(w).Diagnostics {
		codes = append(codes, d.Code)
	}
	return codes
}

func hasAny(codes []string, want ...string) bool {
	for _, c := range codes {
		for _, w := range want {
			if c == w {
				return true
			}
		}
	}
	return false
}

func hasAll(codes []string, want ...string) bool {
	for _, w := range want {
		found := false
		for _, c := range codes {
			if c == w {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}
