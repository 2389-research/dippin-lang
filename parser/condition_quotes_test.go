package parser

import (
	"strings"
	"testing"
)

const condQuoteHeader = `workflow W
  start: A
  exit: B
  tool A
    command: "echo hi"
  agent B
    prompt: "x"
  edges
`

func TestConditionRaw_PreservesEscapedQuotes(t *testing.T) {
	src := condQuoteHeader + "    A -> B when ctx.tool_stdout = \"say \\\"alpha||beta\\\"\"\n"
	w, err := NewParser(src, "t.dip").Parse()
	if err != nil {
		t.Fatal(err)
	}
	got := w.Edges[0].Condition.Raw
	want := `ctx.tool_stdout = "say \"alpha||beta\""`
	if got != want {
		t.Fatalf("Raw = %q, want %q", got, want)
	}
}

func TestParse_RejectsUnterminatedDoubleQuote(t *testing.T) {
	src := condQuoteHeader + "    A -> B when ctx.tool_stdout = \"alpha||beta\n"
	_, err := NewParser(src, "t.dip").Parse()
	if err == nil || !strings.Contains(err.Error(), "unterminated string") {
		t.Fatalf("want unterminated-string rejection, got %v", err)
	}
}

func TestParse_TerminatedQuoteNotFlagged(t *testing.T) {
	src := condQuoteHeader + "    A -> B when ctx.reason = 'needs review'\n"
	if _, err := NewParser(src, "t.dip").Parse(); err != nil {
		t.Fatalf("single-quoted condition must still parse: %v", err)
	}
}
