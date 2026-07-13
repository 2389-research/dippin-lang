// ABOUTME: Verifies simulator conditions against Raw text produced by the real .dip parser.
// ABOUTME: Covers quoting compatibility that crosses the parser and simulator boundary.
package simulate_test

import (
	"testing"

	"github.com/2389-research/dippin-lang/ir"
	"github.com/2389-research/dippin-lang/parser"
	"github.com/2389-research/dippin-lang/simulate"
)

func TestParseConditionPreservesNormalizedLiteralTab(t *testing.T) {
	source := "workflow X\n" +
		"  goal: \"Test\"\n" +
		"  start: A\n" +
		"  exit: B\n" +
		"\n" +
		"  agent A\n" +
		"    prompt: \"Do A.\"\n" +
		"\n" +
		"  agent B\n" +
		"    prompt: \"Do B.\"\n" +
		"\n" +
		"  edges\n" +
		"    A -> B when ctx.x = 'a\tcafé'\n"
	w, err := parser.NewParser(source, "test.dip").Parse()
	if err != nil {
		t.Fatalf("parser error: %v", err)
	}
	expr, err := simulate.ParseCondition(w.Edges[0].Condition.Raw)
	if err != nil {
		t.Fatalf("ParseCondition error: %v", err)
	}
	cmp, ok := expr.(ir.CondCompare)
	if !ok {
		t.Fatalf("expected CondCompare, got %T", expr)
	}
	if got, want := cmp.Value, "a\tcafé"; got != want {
		t.Fatalf("Value = %q, want literal tab and non-ASCII content %q", got, want)
	}
}
