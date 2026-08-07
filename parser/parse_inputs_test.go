package parser

import (
	"testing"

	"github.com/2389-research/dippin-lang/ir"
)

// parseSrc parses test source and returns the workflow plus any parse
// diagnostics. Diagnostics come from Parser.Diagnostics(), not from Parse().
func parseSrc(t *testing.T, src string) (*ir.Workflow, []string) {
	t.Helper()
	p := NewParser(src, "test.dip")
	w, err := p.Parse()
	if err != nil {
		t.Fatalf("parse returned error: %v", err)
	}
	return w, p.Diagnostics()
}

func TestParseInputsMinimalForm(t *testing.T) {
	src := `workflow W
  goal: "test"
  start: A
  exit: A

  inputs
    idea: text
    count: number

  agent A
    prompt:
      hi
`
	w, diags := parseSrc(t, src)
	if len(diags) != 0 {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if len(w.Inputs) != 2 {
		t.Fatalf("got %d inputs, want 2", len(w.Inputs))
	}
	if w.Inputs[0].Name != "idea" || w.Inputs[0].Type != "text" {
		t.Errorf("input 0 = %q/%q, want idea/text", w.Inputs[0].Name, w.Inputs[0].Type)
	}
	if w.Inputs[1].Name != "count" || w.Inputs[1].Type != "number" {
		t.Errorf("input 1 = %q/%q, want count/number", w.Inputs[1].Name, w.Inputs[1].Type)
	}
}

func TestParseInputsPreservesDeclarationOrder(t *testing.T) {
	src := `workflow W
  goal: "test"
  start: A
  exit: A

  inputs
    zebra: text
    apple: text
    middle: text

  agent A
    prompt:
      hi
`
	w, _ := parseSrc(t, src)
	want := []string{"zebra", "apple", "middle"}
	if len(w.Inputs) != len(want) {
		t.Fatalf("got %d inputs, want %d", len(w.Inputs), len(want))
	}
	for i, name := range want {
		if w.Inputs[i].Name != name {
			t.Errorf("input %d = %q, want %q", i, w.Inputs[i].Name, name)
		}
	}
}

func TestParseNoInputsBlockLeavesNil(t *testing.T) {
	src := `workflow W
  goal: "test"
  start: A
  exit: A

  agent A
    prompt:
      hi
`
	w, _ := parseSrc(t, src)
	if w.Inputs != nil {
		t.Errorf("Inputs = %v, want nil", w.Inputs)
	}
}

func TestParseInputsDuplicateNameDiagnoses(t *testing.T) {
	src := `workflow W
  goal: "test"
  start: A
  exit: A

  inputs
    idea: text
    idea: number

  agent A
    prompt:
      hi
`
	// Parse() itself returns a non-nil error whenever diagnostics accumulate
	// (see Parser.Parse), so this test bypasses parseSrc's fatal-on-err
	// helper and reads diagnostics directly, matching the convention used by
	// other duplicate-diagnostic tests (e.g. TestParseElseDuplicate).
	p := NewParser(src, "test.dip")
	_, _ = p.Parse()
	if len(p.Diagnostics()) == 0 {
		t.Fatal("expected a duplicate-name diagnostic, got none")
	}
}

func TestParseInputsRecordsSourceLocation(t *testing.T) {
	src := `workflow W
  goal: "test"
  start: A
  exit: A

  inputs
    idea: text

  agent A
    prompt:
      hi
`
	w, _ := parseSrc(t, src)
	if len(w.Inputs) != 1 {
		t.Fatalf("got %d inputs, want 1", len(w.Inputs))
	}
	if w.Inputs[0].Source.Line == 0 {
		t.Error("Source.Line = 0, want the declaration line")
	}
}
