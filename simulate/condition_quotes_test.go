package simulate

import (
	"testing"

	"github.com/2389-research/dippin-lang/ir"
)

func TestParseCondition_EscapedQuotes(t *testing.T) {
	expr, err := ParseCondition(`ctx.tool_stdout = "say \"alpha||beta\""`)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	cmp, ok := expr.(ir.CondCompare)
	if !ok {
		t.Fatalf("want CondCompare, got %T", expr)
	}
	if cmp.Value != `say "alpha||beta"` {
		t.Fatalf("Value = %q, want %q", cmp.Value, `say "alpha||beta"`)
	}
}

func TestParseCondition_SingleQuotePreserved(t *testing.T) {
	expr, err := ParseCondition(`ctx.reason = 'needs review'`)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	cmp := expr.(ir.CondCompare)
	if cmp.Value != "needs review" {
		t.Fatalf("Value = %q", cmp.Value)
	}
}
