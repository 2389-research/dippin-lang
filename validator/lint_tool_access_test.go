package validator

import (
	"strings"
	"testing"

	"github.com/2389-research/dippin-lang/parser"
)

func TestLint_DIP139_InvalidToolAccess(t *testing.T) {
	src := `workflow X
  start: A
  exit: A

  agent A
    prompt: "x"
    tool_access: foo
`
	p := parser.NewParser(src, "test.dip")
	w, err := p.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	res := Lint(w)
	if !hasCode(res.Diagnostics, DIP139) {
		t.Errorf("expected DIP139, got: %v", codes(res.Diagnostics))
	}
}

func TestLint_DIP139_ValidValues(t *testing.T) {
	cases := []struct {
		name string
		val  string
	}{
		{"none lowercase", "none"},
		{"none uppercase", "NONE"},
		{"none mixed", "None"},
		{"none with surrounding whitespace", "  none  "},
		{"empty omitted", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var line string
			if tc.val != "" {
				// Wrap in quotes if it contains leading/trailing whitespace
				// so the parser preserves it (then strings.TrimSpace in the
				// validator does its job).
				if strings.TrimSpace(tc.val) != tc.val {
					line = "    tool_access: \"" + tc.val + "\"\n"
				} else {
					line = "    tool_access: " + tc.val + "\n"
				}
			}
			src := "workflow X\n  start: A\n  exit: A\n\n  agent A\n    prompt: \"x\"\n" + line
			p := parser.NewParser(src, "test.dip")
			w, err := p.Parse()
			if err != nil {
				t.Fatalf("parse error: %v", err)
			}
			res := Lint(w)
			if hasCode(res.Diagnostics, DIP139) {
				t.Errorf("DIP139 should not fire for %q; got: %v", tc.val, codes(res.Diagnostics))
			}
		})
	}
}

// hasCode and codes are test helpers used by lint tests.
func hasCode(diags []Diagnostic, code string) bool {
	for _, d := range diags {
		if d.Code == code {
			return true
		}
	}
	return false
}

func codes(diags []Diagnostic) []string {
	out := make([]string, 0, len(diags))
	for _, d := range diags {
		out = append(out, d.Code)
	}
	return out
}
