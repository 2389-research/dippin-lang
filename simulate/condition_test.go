// ABOUTME: Verifies condition tokenization, parsing, and workflow AST population.
// ABOUTME: Covers comparison operators, boolean composition, quoting, and errors.
package simulate

import (
	"strings"
	"testing"

	"github.com/2389-research/dippin-lang/ir"
)

func TestParseCondition_SimpleCompare(t *testing.T) {
	expr, err := ParseCondition("ctx.outcome = success")
	if err != nil {
		t.Fatalf("ParseCondition error: %v", err)
	}

	cmp, ok := expr.(ir.CondCompare)
	if !ok {
		t.Fatalf("expected CondCompare, got %T", expr)
	}
	if cmp.Variable != "ctx.outcome" {
		t.Errorf("Variable = %q, want ctx.outcome", cmp.Variable)
	}
	if cmp.Op != "=" {
		t.Errorf("Op = %q, want =", cmp.Op)
	}
	if cmp.Value != "success" {
		t.Errorf("Value = %q, want success", cmp.Value)
	}
}

func TestParseCondition_DecodesEscapedDoubleQuotedValue(t *testing.T) {
	expr, err := ParseCondition(`ctx.tool_stdout = "say \"alpha\\beta\""`)
	if err != nil {
		t.Fatalf("ParseCondition error: %v", err)
	}
	cmp, ok := expr.(ir.CondCompare)
	if !ok {
		t.Fatalf("expected CondCompare, got %T", expr)
	}
	if cmp.Value != `say "alpha\beta"` {
		t.Fatalf("Value = %q, want decoded quote and backslash", cmp.Value)
	}
}

func TestParseCondition_NotEqual(t *testing.T) {
	expr, err := ParseCondition("ctx.x != empty")
	if err != nil {
		t.Fatalf("ParseCondition error: %v", err)
	}

	cmp, ok := expr.(ir.CondCompare)
	if !ok {
		t.Fatalf("expected CondCompare, got %T", expr)
	}
	if cmp.Op != "!=" {
		t.Errorf("Op = %q, want !=", cmp.Op)
	}
}

func TestParseCondition_And(t *testing.T) {
	expr, err := ParseCondition("ctx.outcome = success and ctx.tool_stdout != empty")
	if err != nil {
		t.Fatalf("ParseCondition error: %v", err)
	}

	and, ok := expr.(ir.CondAnd)
	if !ok {
		t.Fatalf("expected CondAnd, got %T", expr)
	}
	if _, ok := and.Left.(ir.CondCompare); !ok {
		t.Errorf("Left = %T, want CondCompare", and.Left)
	}
	if _, ok := and.Right.(ir.CondCompare); !ok {
		t.Errorf("Right = %T, want CondCompare", and.Right)
	}
}

func TestParseCondition_Or(t *testing.T) {
	expr, err := ParseCondition("ctx.x = a or ctx.x = b")
	if err != nil {
		t.Fatalf("ParseCondition error: %v", err)
	}

	or, ok := expr.(ir.CondOr)
	if !ok {
		t.Fatalf("expected CondOr, got %T", expr)
	}
	if _, ok := or.Left.(ir.CondCompare); !ok {
		t.Errorf("Left = %T, want CondCompare", or.Left)
	}
	if _, ok := or.Right.(ir.CondCompare); !ok {
		t.Errorf("Right = %T, want CondCompare", or.Right)
	}
}

func TestParseCondition_Not(t *testing.T) {
	expr, err := ParseCondition("not ctx.outcome = success")
	if err != nil {
		t.Fatalf("ParseCondition error: %v", err)
	}

	not, ok := expr.(ir.CondNot)
	if !ok {
		t.Fatalf("expected CondNot, got %T", expr)
	}
	if _, ok := not.Inner.(ir.CondCompare); !ok {
		t.Errorf("Inner = %T, want CondCompare", not.Inner)
	}
}

func TestParseCondition_ComplexAndOr(t *testing.T) {
	// "and" binds tighter than "or": a or b and c = a or (b and c)
	expr, err := ParseCondition("ctx.x = a or ctx.y = b and ctx.z = c")
	if err != nil {
		t.Fatalf("ParseCondition error: %v", err)
	}

	or, ok := expr.(ir.CondOr)
	if !ok {
		t.Fatalf("expected CondOr at top, got %T", expr)
	}
	if _, ok := or.Left.(ir.CondCompare); !ok {
		t.Errorf("Left = %T, want CondCompare", or.Left)
	}
	if _, ok := or.Right.(ir.CondAnd); !ok {
		t.Errorf("Right = %T, want CondAnd", or.Right)
	}
}

func TestParseCondition_Contains(t *testing.T) {
	expr, err := ParseCondition("ctx.output contains error")
	if err != nil {
		t.Fatalf("ParseCondition error: %v", err)
	}

	cmp, ok := expr.(ir.CondCompare)
	if !ok {
		t.Fatalf("expected CondCompare, got %T", expr)
	}
	if cmp.Op != "contains" {
		t.Errorf("Op = %q, want contains", cmp.Op)
	}
}

func TestParseCondition_In(t *testing.T) {
	expr, err := ParseCondition("ctx.x in 'a, b, c'")
	if err != nil {
		t.Fatalf("ParseCondition error: %v", err)
	}

	cmp, ok := expr.(ir.CondCompare)
	if !ok {
		t.Fatalf("expected CondCompare, got %T", expr)
	}
	if cmp.Op != "in" {
		t.Errorf("Op = %q, want in", cmp.Op)
	}
	if cmp.Value != "a, b, c" {
		t.Errorf("Value = %q, want a, b, c", cmp.Value)
	}
}

func TestParseCondition_EmptyInput(t *testing.T) {
	_, err := ParseCondition("")
	if err == nil {
		t.Fatal("expected error for empty input")
	}
}

func TestTokenizeCondition(t *testing.T) {
	tests := []struct {
		input string
		want  []string
	}{
		{"ctx.x = y", []string{"ctx.x", "=", "y"}},
		{"ctx.x != y", []string{"ctx.x", "!=", "y"}},
		{"not ctx.x = y", []string{"not", "ctx.x", "=", "y"}},
		{"a = b and c = d", []string{"a", "=", "b", "and", "c", "=", "d"}},
	}

	for _, tt := range tests {
		tokens := tokenizeCondition(tt.input)
		if len(tokens) != len(tt.want) {
			t.Errorf("tokenize(%q) = %v, want %v", tt.input, tokens, tt.want)
			continue
		}
		for i, tok := range tokens {
			if tok != tt.want[i] {
				t.Errorf("tokenize(%q)[%d] = %q, want %q", tt.input, i, tok, tt.want[i])
			}
		}
	}
}

func TestEnsureConditionsParsed(t *testing.T) {
	// Pre-parse the third edge via the real parser so the fixture matches
	// parser output, then assert EnsureConditionsParsed leaves it untouched
	// (idempotency / "already parsed" branch).
	preParsed, err := ParseCondition("ctx.z = w")
	if err != nil {
		t.Fatalf("ParseCondition error: %v", err)
	}

	w := &ir.Workflow{
		Edges: []*ir.Edge{
			{From: "A", To: "B", Condition: &ir.Condition{Raw: "ctx.x = y"}},
			{From: "B", To: "C"}, // no condition
			{From: "C", To: "D", Condition: &ir.Condition{
				Raw:    "ctx.z = w",
				Parsed: preParsed,
			}}, // already parsed
		},
	}

	if err := EnsureConditionsParsed(w); err != nil {
		t.Fatalf("EnsureConditionsParsed error: %v", err)
	}

	// First edge should now have Parsed.
	if w.Edges[0].Condition.Parsed == nil {
		t.Error("edge A->B condition should be parsed")
	}
	cmp, ok := w.Edges[0].Condition.Parsed.(ir.CondCompare)
	if !ok {
		t.Fatalf("expected CondCompare, got %T", w.Edges[0].Condition.Parsed)
	}
	if cmp.Variable != "ctx.x" || cmp.Op != "=" || cmp.Value != "y" {
		t.Errorf("parsed condition = %+v", cmp)
	}

	// Second edge should remain nil.
	if w.Edges[1].Condition != nil {
		t.Error("edge B->C should have no condition")
	}

	// Third edge should remain as-is (already parsed, idempotent).
	if w.Edges[2].Condition.Parsed != preParsed {
		t.Errorf("edge C->D condition Parsed changed: got %+v, want %+v", w.Edges[2].Condition.Parsed, preParsed)
	}
}

func TestEnsureConditionsParsed_ManagerLoop(t *testing.T) {
	stop := &ir.Condition{Raw: "stack.child.cycles = 10"}
	steer := &ir.Condition{Raw: "stack.child.cycles = 5"}
	w := &ir.Workflow{
		Name:  "W",
		Start: "M",
		Exit:  "M",
		Nodes: []*ir.Node{
			{ID: "M", Kind: ir.NodeManagerLoop, Config: ir.ManagerLoopConfig{
				SubgraphRef:    "inner",
				StopCondition:  stop,
				SteerCondition: steer,
			}},
		},
	}
	if err := EnsureConditionsParsed(w); err != nil {
		t.Fatalf("EnsureConditionsParsed error: %v", err)
	}
	if stop.Parsed == nil {
		t.Errorf("StopCondition.Parsed still nil after EnsureConditionsParsed")
	}
	if steer.Parsed == nil {
		t.Errorf("SteerCondition.Parsed still nil after EnsureConditionsParsed")
	}
}

func TestEnsureConditionsParsed_ManagerLoop_InvalidCondition(t *testing.T) {
	w := &ir.Workflow{
		Name:  "W",
		Start: "M",
		Exit:  "M",
		Nodes: []*ir.Node{
			{ID: "Supervisor", Kind: ir.NodeManagerLoop, Config: ir.ManagerLoopConfig{
				SubgraphRef:   "inner",
				StopCondition: &ir.Condition{Raw: "stack.child.cycles ="},
			}},
		},
	}
	err := EnsureConditionsParsed(w)
	if err == nil {
		t.Fatal("expected error for invalid condition, got nil")
	}
	if !strings.Contains(err.Error(), "node Supervisor stop_condition") {
		t.Errorf("error %q does not contain %q", err.Error(), "node Supervisor stop_condition")
	}
	if !strings.Contains(err.Error(), `"stack.child.cycles ="`) {
		t.Errorf("error %q does not include the offending raw condition", err.Error())
	}
}
