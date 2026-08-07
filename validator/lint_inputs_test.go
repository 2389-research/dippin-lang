package validator_test

import (
	"testing"

	"github.com/2389-research/dippin-lang/parser"
	"github.com/2389-research/dippin-lang/validator"
)

// lintSrc parses source and returns its lint diagnostics. Parse diagnostics
// come from Parser.Diagnostics(), not from Parse() — there is no
// parser.Parse(src) package function.
func lintSrc(t *testing.T, src string) []validator.Diagnostic {
	t.Helper()
	p := parser.NewParser(src, "test.dip")
	w, err := p.Parse()
	if err != nil {
		t.Fatalf("parse returned error: %v", err)
	}
	if diags := p.Diagnostics(); len(diags) != 0 {
		t.Fatalf("parse diagnostics: %v", diags)
	}
	return validator.Lint(w).Diagnostics
}

// hasCode reports whether any diagnostic carries the given code.
func hasCode(diags []validator.Diagnostic, code string) bool {
	for _, d := range diags {
		if d.Code == code {
			return true
		}
	}
	return false
}

func TestDIP155UnknownInputType(t *testing.T) {
	src := `workflow W
  goal: "test"
  start: A
  exit: A

  inputs
    when: duration

  agent A
    prompt:
      hi
`
	diags := lintSrc(t, src)
	if !hasCode(diags, "DIP155") {
		t.Fatalf("want DIP155 for unknown type, got %v", diags)
	}
	for _, d := range diags {
		if d.Code == "DIP155" && d.Severity != validator.SeverityError {
			t.Errorf("DIP155 severity = %v, want Error", d.Severity)
		}
	}
}

func TestDIP155AcceptsEveryKnownType(t *testing.T) {
	src := `workflow W
  goal: "test"
  start: A
  exit: A

  inputs
    a: text
    b: number
    c: bool
    d: enum
    e: file
    f: secret

  agent A
    prompt:
      hi
`
	if diags := lintSrc(t, src); hasCode(diags, "DIP155") {
		t.Errorf("DIP155 fired on a known type: %v", diags)
	}
}
