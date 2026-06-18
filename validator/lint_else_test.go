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

// TestLintElseCountsAsSuccessPath verifies DIP105 (no forward path start→exit)
// does not fire when the success-side route to exit is the section `else`
// default. else is a forward routing path, so the success-path walk must see it.
func TestLintElseToExit(t *testing.T) {
	src := `workflow ElseToExit
  start: Work
  exit: Done
  agent Work
    prompt: "do work, loop until ready"
  agent Done
    prompt: "finish"
  edges
    Work -> Work when ctx.outcome = retry loop
    else -> Done
`
	if codes := lintCodes(t, src); hasAny(codes, validator.DIP105) {
		t.Fatalf("else route to exit should satisfy DIP105, but it fired: %v", codes)
	}
}

// TestLintElseSuppressionGatedOnTargetExisting verifies the DIP101/DIP102
// suppression does NOT apply when the `else` target is an unknown node — a
// structurally invalid `else` must not silently hide routing problems. (Lint is
// independent of Validate, mirroring how DIP144 gates on_failure on node
// existence.)
func TestLintElseSuppressionGatedOnTargetExisting(t *testing.T) {
	src := `workflow GhostElse
  start: T
  exit: D
  tool T
    command: echo hi
  agent D
    prompt: "done"
  edges
    T -> D when ctx.tool_marker = go
    else -> Ghost
`
	if codes := lintCodes(t, src); !hasAll(codes, validator.DIP101, validator.DIP102) {
		t.Fatalf("an else pointing at an unknown node must not suppress DIP101/DIP102, got: %v", codes)
	}
}

// TestLintElseDoesNotCorruptDip112 guards against the regression where adding
// the else route to the shared forward adjacency corrupted DIP112's in-degree
// topo traversal (in-degrees are computed from w.Edges, not the adjacency).
// Here T legitimately reads `data` written by its upstream A, so DIP112 must
// stay silent — the else edge must not reorder traversal so A's write is missed.
func TestLintElseDoesNotCorruptDip112(t *testing.T) {
	src := `workflow Dip112Else
  start: S
  exit: Done
  agent S
    prompt: "start"
  agent A
    prompt: "relay"
  agent B
    prompt: "produce"
    writes: data
  agent T
    prompt: "consume"
    reads: data
  agent Done
    prompt: "done"
  edges
    S -> A
    A -> B
    B -> T
    T -> Done
    else -> T
`
	if codes := lintCodes(t, src); hasAny(codes, validator.DIP112) {
		t.Fatalf("else must not corrupt DIP112 topo traversal; data is written upstream by A, got: %v", codes)
	}
}

func lintCodes(t *testing.T, src string) []string {
	t.Helper()
	p := parser.NewParser(src, "test.dip")
	w, err := p.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if diags := p.Diagnostics(); len(diags) > 0 {
		t.Fatalf("fixture produced parser diagnostics (fixtures must match real parser output): %v", diags)
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
