package parser_test

import (
	"testing"

	"github.com/2389-research/dippin-lang/formatter"
	"github.com/2389-research/dippin-lang/parser"
	"github.com/2389-research/dippin-lang/simulate"
)

func TestEscapedQuoteCondition_ValidatesAndRoundTrips(t *testing.T) {
	src := "workflow W\n  start: A\n  exit: B\n  tool A\n    command: \"echo hi\"\n  agent B\n    prompt: \"x\"\n  edges\n    A -> B when ctx.tool_stdout = \"say \\\"alpha||beta\\\"\"\n"
	w, err := parser.NewParser(src, "t.dip").Parse()
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if err := simulate.EnsureConditionsParsed(w); err != nil {
		t.Fatalf("condition parse (DIP010 path): %v", err)
	}
	out := formatter.Format(w)
	w2, err := parser.NewParser(out, "t.dip").Parse()
	if err != nil {
		t.Fatalf("reparse:\n%s\nerr: %v", out, err)
	}
	if formatter.Format(w2) != out {
		t.Fatalf("not idempotent:\n--- first ---\n%s\n--- second ---\n%s", out, formatter.Format(w2))
	}
}
