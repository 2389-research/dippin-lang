package validator_test

import (
	"testing"

	"github.com/2389-research/dippin-lang/parser"
	"github.com/2389-research/dippin-lang/validator"
)

// TestLintMarkerGrepSuppression verifies that a tool node declaring
// marker_grep: routes via the typed ctx.tool_marker channel — outgoing
// conditional edges on such nodes are intentional and should NOT trigger
// DIP101 (unreachable via conditional) or DIP102 (no default edge).
//
// The companion negative case asserts that without marker_grep the same
// shape still fires both warnings, guarding against accidental over-
// suppression.
//
// Tests parse real .dip text through the production parser so they exercise
// the same code path as user workflows (per CLAUDE.md guidance).
func TestLintMarkerGrepSuppression(t *testing.T) {
	tests := []struct {
		name      string
		src       string
		wantCodes []string // codes that MUST appear; empty = none of DIP101/DIP102
	}{
		{
			name: "DIP101/DIP102: marker_grep on tool node suppresses both",
			src: `workflow MarkerSafe
  start: T
  exit: D
  tool T
    command: echo hi
    marker_grep: "^(go|stop)$"
  agent D
    prompt: "done"
  edges
    T -> D when ctx.tool_marker = go
`,
			wantCodes: []string{}, // no DIP101, no DIP102
		},
		{
			name: "DIP101/DIP102: without marker_grep, same shape fires both",
			src: `workflow MarkerUnsafe
  start: T
  exit: D
  tool T
    command: echo hi
  agent D
    prompt: "done"
  edges
    T -> D when ctx.tool_marker = go
`,
			wantCodes: []string{validator.DIP101, validator.DIP102},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			p := parser.NewParser(tt.src, tt.name+".dip")
			w, err := p.Parse()
			if err != nil {
				t.Fatalf("parse error: %v", err)
			}

			result := validator.Lint(w)

			gotCodes := make([]string, len(result.Diagnostics))
			for i, d := range result.Diagnostics {
				gotCodes[i] = d.Code
			}

			// Every expected code must appear at least once.
			for _, want := range tt.wantCodes {
				found := false
				for _, got := range gotCodes {
					if got == want {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected %s diagnostic, got none. All codes: %v", want, gotCodes)
				}
			}

			// When wantCodes is empty, DIP101 and DIP102 specifically must NOT fire.
			if len(tt.wantCodes) == 0 {
				for _, got := range gotCodes {
					if got == validator.DIP101 || got == validator.DIP102 {
						t.Errorf("unexpected %s diagnostic (expected suppression). All codes: %v", got, gotCodes)
					}
				}
			}
		})
	}
}
