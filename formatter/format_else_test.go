package formatter

import (
	"strings"
	"testing"

	"github.com/2389-research/dippin-lang/parser"
)

// TestFormatElseDefault verifies the formatter emits a section-level `else -> X`
// as the last entry of the edges block, and that the result round-trips (parse →
// format → parse preserves ElseTarget).
func TestFormatElseDefault(t *testing.T) {
	src := `workflow ElseFmt
  goal: "fmt else"
  start: T
  exit: D
  agent T
    prompt: "t"
  agent D
    prompt: "d"
  agent Cleanup
    prompt: "c"
  edges
    T -> D when ctx.outcome = success
    else -> Cleanup
`
	p := parser.NewParser(src, "t.dip")
	w, err := p.Parse()
	if err != nil {
		t.Fatalf("parse: %v (%v)", err, p.Diagnostics())
	}
	out := Format(w)
	if !strings.Contains(out, "else -> Cleanup") {
		t.Fatalf("formatted output missing `else -> Cleanup`:\n%s", out)
	}
	// else must be the last edges-block line (after the guarded edge).
	if idx, last := strings.Index(out, "T -> D"), strings.Index(out, "else -> Cleanup"); idx == -1 || last < idx {
		t.Fatalf("else not emitted after the regular edge:\n%s", out)
	}
	// Round-trip: re-parse the formatted text, ElseTarget preserved.
	w2, err := parser.NewParser(out, "t2.dip").Parse()
	if err != nil {
		t.Fatalf("reparse: %v", err)
	}
	if w2.ElseTarget != "Cleanup" {
		t.Fatalf("round-trip ElseTarget = %q, want Cleanup", w2.ElseTarget)
	}
}
