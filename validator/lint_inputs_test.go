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

// --- DIP158: invalid or inapplicable input constraint ---

func TestDIP158EnumDefaultNotInOptions(t *testing.T) {
	src := `workflow W
  goal: t
  start: A
  exit: A
  inputs
    risk: enum
      options: low, high
      default: medium
  agent A
    prompt:
      hi
`
	diags := lintSrc(t, src)
	if !hasCode(diags, "DIP158") {
		t.Fatalf("enum default not in options should fire DIP158, got %v", diags)
	}
	for _, d := range diags {
		if d.Code == "DIP158" && d.Severity != validator.SeverityError {
			t.Errorf("DIP158 severity = %v, want Error", d.Severity)
		}
	}
}

func TestDIP158EnumDefaultInOptionsClean(t *testing.T) {
	src := `workflow W
  goal: t
  start: A
  exit: A
  inputs
    risk: enum
      options: low, medium, high
      default: medium
  agent A
    prompt:
      hi
`
	if hasCode(lintSrc(t, src), "DIP158") {
		t.Error("a valid enum default must not fire DIP158")
	}
}

func TestDIP158MinGreaterThanMax(t *testing.T) {
	src := `workflow W
  goal: t
  start: A
  exit: A
  inputs
    n: number
      min: 10
      max: 5
  agent A
    prompt:
      hi
`
	if !hasCode(lintSrc(t, src), "DIP158") {
		t.Error("min > max should fire DIP158")
	}
}

func TestDIP158MalformedPattern(t *testing.T) {
	src := `workflow W
  goal: t
  start: A
  exit: A
  inputs
    s: text
      pattern: "[unterminated"
  agent A
    prompt:
      hi
`
	if !hasCode(lintSrc(t, src), "DIP158") {
		t.Error("a malformed pattern regex should fire DIP158")
	}
}

func TestDIP158ConstraintOnWrongType(t *testing.T) {
	// max_length is a text constraint; on a bool it is inapplicable.
	src := `workflow W
  goal: t
  start: A
  exit: A
  inputs
    flag: bool
      max_length: 4
  agent A
    prompt:
      hi
`
	if !hasCode(lintSrc(t, src), "DIP158") {
		t.Error("max_length on a bool should fire DIP158")
	}
}

func TestDIP158OptionsOnNonEnum(t *testing.T) {
	src := `workflow W
  goal: t
  start: A
  exit: A
  inputs
    name: text
      options: a, b
  agent A
    prompt:
      hi
`
	if !hasCode(lintSrc(t, src), "DIP158") {
		t.Error("options on a non-enum should fire DIP158")
	}
}

func TestDIP158ValidConstraintsClean(t *testing.T) {
	src := `workflow W
  goal: t
  start: A
  exit: A
  inputs
    idea: text
      pattern: "^[a-z]+$"
      max_length: 100
      multiline: true
    n: number
      min: 1
      max: 10
  agent A
    prompt:
      Build ${inputs.idea} x${inputs.n}.
`
	if hasCode(lintSrc(t, src), "DIP158") {
		t.Errorf("valid constraints must not fire DIP158: %v", lintSrc(t, src))
	}
}

// --- DIP159: declared input never referenced (dead input) ---

func TestDIP159DeadInput(t *testing.T) {
	src := `workflow W
  goal: t
  start: A
  exit: A
  inputs
    used: text
    unused: text
  agent A
    prompt:
      Only ${inputs.used} here.
`
	diags := lintSrc(t, src)
	if !hasCode(diags, "DIP159") {
		t.Fatalf("an unreferenced input should fire DIP159, got %v", diags)
	}
	for _, d := range diags {
		if d.Code == "DIP159" && d.Severity != validator.SeverityWarning {
			t.Errorf("DIP159 severity = %v, want Warning", d.Severity)
		}
		if d.Code == "DIP159" && !contains(d.Message, "unused") {
			t.Errorf("DIP159 should name the dead input; got %q", d.Message)
		}
	}
}

func TestDIP159ReferencedInputsClean(t *testing.T) {
	// referenced in a prompt, an edge condition, and a tool command respectively.
	src := `workflow W
  goal: t
  start: A
  exit: B
  inputs
    p: text
    c: enum
      options: x, y
    t: text
  agent A
    prompt:
      ${inputs.p}
  tool B
    command:
      echo ${inputs.t}
  edges
    A -> B when inputs.c = x
`
	if hasCode(lintSrc(t, src), "DIP159") {
		t.Errorf("inputs referenced in prompt/condition/command must not fire DIP159: %v", lintSrc(t, src))
	}
}

func contains(s, sub string) bool { return len(s) >= len(sub) && (indexOf(s, sub) >= 0) }
func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}

// TestDIP159FileSecretExempt covers #215: a file/secret input is consumed by
// reading its staged path in a command (DIP157 forbids ${inputs.x} there), so
// it is legitimately never ${inputs.x}-referenced and must NOT trip DIP159.
func TestDIP159FileSecretExempt(t *testing.T) {
	src := `workflow W
  goal: t
  start: Setup
  exit: Setup
  inputs
    spec: file
    token: secret
  tool Setup
    command:
      if [ -f .tracker/inputs/spec ]; then cp .tracker/inputs/spec SPEC.md; fi
`
	if hasCode(lintSrc(t, src), "DIP159") {
		t.Errorf("file/secret inputs consumed via staged path must not trip DIP159: %v", lintSrc(t, src))
	}
}

// TestDIP159StillFiresForValueTypes: text/number/bool/enum are still checked.
func TestDIP159StillFiresForValueTypes(t *testing.T) {
	src := `workflow W
  goal: t
  start: A
  exit: A
  inputs
    unused: text
  agent A
    prompt:
      nothing referenced
`
	if !hasCode(lintSrc(t, src), "DIP159") {
		t.Error("a dead text input should still fire DIP159")
	}
}
