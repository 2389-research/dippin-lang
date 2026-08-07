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

func TestParseInputsFullAttributeBlock(t *testing.T) {
	src := `workflow W
  goal: "test"
  start: A
  exit: A

  inputs
    idea: text
      required: true
      prompt: "What do you want built?"
      description: "One or two sentences."
      multiline: true
      max_length: 4000
    risk: enum
      options: low, medium, high
      default: medium
    retries: number
      min: 1
      max: 10
    branch: text
      pattern: "^[a-z]+$"

  agent A
    prompt:
      hi
`
	w, diags := parseSrc(t, src)
	if len(diags) != 0 {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if len(w.Inputs) != 4 {
		t.Fatalf("got %d inputs, want 4", len(w.Inputs))
	}

	idea := w.Inputs[0]
	if !idea.Required {
		t.Error("idea.Required = false, want true")
	}
	if idea.Prompt != "What do you want built?" {
		t.Errorf("idea.Prompt = %q", idea.Prompt)
	}
	if idea.Description != "One or two sentences." {
		t.Errorf("idea.Description = %q", idea.Description)
	}
	if !idea.Multiline {
		t.Error("idea.Multiline = false, want true")
	}
	if idea.MaxLength != 4000 {
		t.Errorf("idea.MaxLength = %d, want 4000", idea.MaxLength)
	}

	risk := w.Inputs[1]
	if len(risk.Options) != 3 || risk.Options[0] != "low" || risk.Options[2] != "high" {
		t.Errorf("risk.Options = %v, want [low medium high]", risk.Options)
	}
	if risk.Default != "medium" || !risk.HasDefault {
		t.Errorf("risk.Default = %q (has=%v), want medium/true", risk.Default, risk.HasDefault)
	}

	retries := w.Inputs[2]
	if retries.Min != "1" || retries.Max != "10" {
		t.Errorf("retries min/max = %q/%q, want 1/10", retries.Min, retries.Max)
	}

	if w.Inputs[3].Pattern != "^[a-z]+$" {
		t.Errorf("branch.Pattern = %q", w.Inputs[3].Pattern)
	}
}

func TestParseInputsEmptyDefaultIsDistinctFromAbsent(t *testing.T) {
	src := `workflow W
  goal: "test"
  start: A
  exit: A

  inputs
    withEmpty: text
      default: ""
    withNone: text

  agent A
    prompt:
      hi
`
	w, _ := parseSrc(t, src)
	if !w.Inputs[0].HasDefault {
		t.Error("withEmpty.HasDefault = false, want true")
	}
	if w.Inputs[0].Default != "" {
		t.Errorf("withEmpty.Default = %q, want empty", w.Inputs[0].Default)
	}
	if w.Inputs[1].HasDefault {
		t.Error("withNone.HasDefault = true, want false")
	}
}

func TestParseInputsUnknownTypeCarriedVerbatim(t *testing.T) {
	src := `workflow W
  goal: "test"
  start: A
  exit: A

  inputs
    when: duration
      required: true

  agent A
    prompt:
      hi
`
	w, diags := parseSrc(t, src)
	if len(diags) != 0 {
		t.Fatalf("parser must not diagnose an unknown type (that is DIP155's job): %v", diags)
	}
	if w.Inputs[0].Type != "duration" {
		t.Errorf("Type = %q, want duration carried verbatim", w.Inputs[0].Type)
	}
	if !w.Inputs[0].Required {
		t.Error("attributes must still parse under an unknown type")
	}
}

func TestParseInputsUnknownAttributeHints(t *testing.T) {
	src := `workflow W
  goal: "test"
  start: A
  exit: A

  inputs
    idea: text
      nonsense: 42
      required: true

  agent A
    prompt:
      hi
`
	// Parser.Parse() returns a non-nil error whenever ANY diagnostic
	// accumulates, so a test that EXPECTS diagnostics cannot use parseSrc
	// (which fatals on error). Call Parse() directly and inspect
	// Diagnostics(), matching the existing TestParseElseDuplicate convention.
	p := NewParser(src, "test.dip")
	w, _ := p.Parse()
	diags := p.Diagnostics()
	if len(diags) == 0 {
		t.Fatal("expected an unknown-attribute hint")
	}
	if !w.Inputs[0].Required {
		t.Error("a stray attribute must not desync the scan — required: true was lost")
	}
}
