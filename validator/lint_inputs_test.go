package validator_test

import (
	"strings"
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

func TestDIP156UndeclaredRefInPrompt(t *testing.T) {
	src := `workflow W
  goal: "test"
  start: A
  exit: A

  inputs
    idea: text

  agent A
    prompt:
      Build ${inputs.idae} for me.
`
	diags := lintSrc(t, src)
	if !hasCode(diags, "DIP156") {
		t.Fatalf("want DIP156 for a typo'd input ref, got %v", diags)
	}
}

func TestDIP156DeclaredRefInPromptIsClean(t *testing.T) {
	src := `workflow W
  goal: "test"
  start: A
  exit: A

  inputs
    idea: text

  agent A
    prompt:
      Build ${inputs.idea} for me.
`
	diags := lintSrc(t, src)
	if hasCode(diags, "DIP156") {
		t.Errorf("DIP156 fired on a declared input: %v", diags)
	}
	if hasCode(diags, "DIP106") {
		t.Errorf("DIP106 fired on the inputs namespace — it must be in knownNamespaces: %v", diags)
	}
}

func TestDIP156UndeclaredRefInEdgeCondition(t *testing.T) {
	src := `workflow W
  goal: "test"
  start: A
  exit: B

  inputs
    risk: enum
      options: low, high

  agent A
    prompt:
      hi

  agent B
    prompt:
      bye

  edges
    A -> B when inputs.rsk = high
`
	diags := lintSrc(t, src)
	if !hasCode(diags, "DIP156") {
		t.Fatalf("want DIP156 for a typo'd input ref in a condition, got %v", diags)
	}
}

func TestDIP156DeclaredRefInEdgeConditionIsClean(t *testing.T) {
	src := `workflow W
  goal: "test"
  start: A
  exit: B

  inputs
    risk: enum
      options: low, high

  agent A
    prompt:
      hi

  agent B
    prompt:
      bye

  edges
    A -> B when inputs.risk = high
`
	diags := lintSrc(t, src)
	if hasCode(diags, "DIP156") {
		t.Errorf("DIP156 fired on a declared input: %v", diags)
	}
	if hasCode(diags, "DIP120") {
		t.Errorf("DIP120 fired on the inputs namespace: %v", diags)
	}
}

func TestDIP157InputRefInToolCommand(t *testing.T) {
	src := `workflow W
  goal: "test"
  start: T
  exit: T

  inputs
    idea: text

  tool T
    command:
      echo ${inputs.idea}
`
	diags := lintSrc(t, src)
	if !hasCode(diags, "DIP157") {
		t.Fatalf("want DIP157 for an input ref in a tool command, got %v", diags)
	}
	for _, d := range diags {
		if d.Code == "DIP157" && d.Severity != validator.SeverityError {
			t.Errorf("DIP157 severity = %v, want Error", d.Severity)
		}
	}
}

func TestDIP157SecretGetsSharperHelp(t *testing.T) {
	src := `workflow W
  goal: "test"
  start: T
  exit: T

  inputs
    token: secret

  tool T
    command:
      curl -H "Authorization: ${inputs.token}" https://example.com
`
	diags := lintSrc(t, src)
	var help string
	for _, d := range diags {
		if d.Code == "DIP157" {
			help = d.Help
		}
	}
	if help == "" {
		t.Fatalf("want DIP157, got %v", diags)
	}
	if !strings.Contains(help, "secret") {
		t.Errorf("help for a secret should mention it, got %q", help)
	}
}

func TestDIP157CleanWhenInputRefIsInAnAgentPrompt(t *testing.T) {
	src := `workflow W
  goal: "test"
  start: A
  exit: A

  inputs
    idea: text

  agent A
    prompt:
      Build ${inputs.idea}.
`
	if diags := lintSrc(t, src); hasCode(diags, "DIP157") {
		t.Errorf("DIP157 fired on an agent prompt, which interpolates fine: %v", diags)
	}
}
