package formatter_test

import (
	"strings"
	"testing"

	"github.com/2389-research/dippin-lang/formatter"
	"github.com/2389-research/dippin-lang/ir"
	"github.com/2389-research/dippin-lang/parser"
)

// parseSrc parses test source and returns the workflow plus any parse
// diagnostics. Diagnostics come from Parser.Diagnostics(), not from Parse().
func parseSrc(t *testing.T, src string) (*ir.Workflow, []string) {
	t.Helper()
	p := parser.NewParser(src, "test.dip")
	w, err := p.Parse()
	if err != nil {
		t.Fatalf("parse returned error: %v", err)
	}
	return w, p.Diagnostics()
}

const inputsSource = `workflow W
  goal: "test"
  start: A
  exit: A

  inputs
    idea: text
      required: true
      prompt: "What do you want built?"
      multiline: true
      max_length: 4000
    risk: enum
      options: low, medium, high
      default: medium
    plain: text

  agent A
    prompt:
      hi
`

func TestFormatInputsRoundTrip(t *testing.T) {
	w, diags := parseSrc(t, inputsSource)
	if len(diags) != 0 {
		t.Fatalf("parse diagnostics: %v", diags)
	}
	out := formatter.Format(w)

	w2, diags2 := parseSrc(t, out)
	if len(diags2) != 0 {
		t.Fatalf("reparse diagnostics: %v\n%s", diags2, out)
	}
	if got := formatter.Format(w2); got != out {
		t.Errorf("format is not idempotent:\n--- first ---\n%s\n--- second ---\n%s", out, got)
	}
	if len(w2.Inputs) != 3 {
		t.Fatalf("got %d inputs after round-trip, want 3", len(w2.Inputs))
	}
}

func TestFormatInputsPreservesDeclarationOrder(t *testing.T) {
	src := `workflow W
  goal: "test"
  start: A
  exit: A

  inputs
    zebra: text
    apple: text

  agent A
    prompt:
      hi
`
	w, _ := parseSrc(t, src)
	out := formatter.Format(w)
	zi, ai := strings.Index(out, "zebra"), strings.Index(out, "apple")
	if zi == -1 || ai == -1 {
		t.Fatalf("both inputs must be emitted:\n%s", out)
	}
	if zi > ai {
		t.Errorf("declaration order not preserved — inputs must not be sorted:\n%s", out)
	}
}

func TestFormatInputsSectionPosition(t *testing.T) {
	src := `workflow W
  goal: "test"
  start: A
  exit: A

  defaults
    max_retries: 3

  vars
    k: v

  inputs
    idea: text

  agent A
    prompt:
      hi
`
	w, _ := parseSrc(t, src)
	out := formatter.Format(w)
	ii := strings.Index(out, "\n  inputs\n")
	di := strings.Index(out, "\n  defaults\n")
	vi := strings.Index(out, "\n  vars\n")
	if ii == -1 || di == -1 || vi == -1 {
		t.Fatalf("all three sections must be emitted:\n%s", out)
	}
	if !(ii < di && di < vi) {
		t.Errorf("want inputs < defaults < vars:\n%s", out)
	}
}

func TestFormatNoInputsEmitsNoBlock(t *testing.T) {
	src := `workflow W
  goal: "test"
  start: A
  exit: A

  agent A
    prompt:
      hi
`
	w, _ := parseSrc(t, src)
	if out := formatter.Format(w); strings.Contains(out, "inputs") {
		t.Errorf("emitted an inputs block for a workflow with none:\n%s", out)
	}
}

func TestFormatInputsOmitsZeroValues(t *testing.T) {
	src := `workflow W
  goal: "test"
  start: A
  exit: A

  inputs
    plain: text

  agent A
    prompt:
      hi
`
	w, _ := parseSrc(t, src)
	out := formatter.Format(w)
	for _, unwanted := range []string{"required:", "multiline:", "max_length:", "default:"} {
		if strings.Contains(out, unwanted) {
			t.Errorf("emitted zero-value attribute %q:\n%s", unwanted, out)
		}
	}
}
