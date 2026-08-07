package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

const inputsFixture = `workflow W
  goal: "test"
  start: A
  exit: A

  inputs
    idea: text
      required: true
      prompt: "What do you want built?"
      max_length: 4000
      multiline: true
      pattern: "^[a-z]+$"
    retries: number
      default: 3
      min: 1
      max: 10
    verbose: bool
      default: false
    risk: enum
      options: low, high
      default: low

  agent A
    prompt:
      Build ${inputs.idea}.
`

func writeInputsFixture(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "w.dip")
	if err := os.WriteFile(path, []byte(inputsFixture), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestCmdInputsJSONTypedDefaults(t *testing.T) {
	path := writeInputsFixture(t)
	var stdout, stderr bytes.Buffer
	cli := &CLI{Stdout: &stdout, Stderr: &stderr}
	if got := cli.CmdInputs([]string{"--format=json", path}); got != ExitOK {
		t.Fatalf("exit = %v, stderr = %s", got, stderr.String())
	}

	var got []map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &got); err != nil {
		t.Fatalf("output is not JSON: %v\n%s", err, stdout.String())
	}
	if len(got) != 4 {
		t.Fatalf("got %d inputs, want 4", len(got))
	}

	// Declaration order is preserved.
	if got[0]["name"] != "idea" || got[3]["name"] != "risk" {
		t.Errorf("declaration order not preserved: %v", got)
	}
	// A number default is a JSON number, not a string.
	if _, ok := got[1]["default"].(float64); !ok {
		t.Errorf("retries default = %T(%v), want a JSON number", got[1]["default"], got[1]["default"])
	}
	// A bool default is a JSON bool.
	if v, ok := got[2]["default"].(bool); !ok || v != false {
		t.Errorf("verbose default = %T(%v), want JSON false", got[2]["default"], got[2]["default"])
	}
	// A text/enum default stays a string.
	if _, ok := got[3]["default"].(string); !ok {
		t.Errorf("risk default = %T, want a JSON string", got[3]["default"])
	}
	if got[0]["required"] != true {
		t.Errorf("idea.required = %v, want true", got[0]["required"])
	}
	// An input with no declared default omits the key entirely, rather than
	// emitting a zero value a host would mistake for a real default.
	if _, present := got[0]["default"]; present {
		t.Errorf("idea has no declared default but the key was emitted: %v", got[0])
	}
}

func TestCmdInputsNoInputsEmitsEmptyArray(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "none.dip")
	src := `workflow W
  goal: "test"
  start: A
  exit: A

  agent A
    prompt:
      hi
`
	if err := os.WriteFile(path, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	cli := &CLI{Stdout: &stdout, Stderr: &stderr}
	if got := cli.CmdInputs([]string{"--format=json", path}); got != ExitOK {
		t.Fatalf("exit = %v, stderr = %s", got, stderr.String())
	}
	var got []map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &got); err != nil {
		t.Fatalf("output is not JSON: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("got %d inputs, want an empty array", len(got))
	}
}

func TestCmdInputsTextFormat(t *testing.T) {
	path := writeInputsFixture(t)
	var stdout, stderr bytes.Buffer
	cli := &CLI{Stdout: &stdout, Stderr: &stderr}
	if got := cli.CmdInputs([]string{path}); got != ExitOK {
		t.Fatalf("exit = %v, stderr = %s", got, stderr.String())
	}
	out := stdout.String()
	for _, want := range []string{
		"idea", "text", "required", "retries",
		"pattern: ^[a-z]+$", "multiline: true", "max_length: 4000",
		"min: 1", "max: 10",
	} {
		if !bytes.Contains([]byte(out), []byte(want)) {
			t.Errorf("text output missing %q:\n%s", want, out)
		}
	}
}

func TestCmdInputsUnknownFormatIsAnError(t *testing.T) {
	path := writeInputsFixture(t)
	var stdout, stderr bytes.Buffer
	cli := &CLI{Stdout: &stdout, Stderr: &stderr}
	if got := cli.CmdInputs([]string{"--format=yaml", path}); got != ExitError {
		t.Fatalf("exit = %v, want ExitError", got)
	}
	if stdout.Len() != 0 {
		t.Errorf("stdout should be empty on an unknown --format, got: %s", stdout.String())
	}
	if stderr.Len() == 0 {
		t.Errorf("expected an error message on stderr")
	}
}
