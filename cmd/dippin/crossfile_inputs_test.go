package main

import (
	"path/filepath"
	"testing"

	"github.com/2389-research/dippin-lang/validator"
)

// arityDiags loads dir/entry and runs the DIP160 cross-file input-arity pass.
func arityDiags(t *testing.T, dir, entry string) []validator.Diagnostic {
	t.Helper()
	entryPath := filepath.Join(dir, entry)
	w, err := loadWorkflow(entryPath)
	if err != nil {
		t.Fatalf("load %s: %v", entry, err)
	}
	return crossFileInputArity(w, entryPath)
}

const childRequiresTopic = `workflow Child
  goal: c
  start: A
  exit: A
  inputs
    topic: text
      required: true
    tone: text
  agent A
    prompt:
      ${inputs.topic}
`

func TestDIP160_FiresWhenRequiredInputOmitted(t *testing.T) {
	dir := writeWorkflows(t, map[string]string{
		"child.dip": childRequiresTopic,
		"parent.dip": `workflow Parent
  goal: p
  start: Sub
  exit: Sub
  subgraph Sub
    ref: child.dip
    params:
      tone: friendly
`,
	})
	diags := arityDiags(t, dir, "parent.dip")
	if countCode(diags, validator.DIP160) != 1 {
		t.Fatalf("want 1 DIP160 (required 'topic' omitted), got %d: %v", countCode(diags, validator.DIP160), diags)
	}
}

func TestDIP160_SilentWhenRequiredInputProvided(t *testing.T) {
	dir := writeWorkflows(t, map[string]string{
		"child.dip": childRequiresTopic,
		"parent.dip": `workflow Parent
  goal: p
  start: Sub
  exit: Sub
  subgraph Sub
    ref: child.dip
    params:
      topic: sprint planning
`,
	})
	if n := countCode(arityDiags(t, dir, "parent.dip"), validator.DIP160); n != 0 {
		t.Errorf("required input provided — want 0 DIP160, got %d", n)
	}
}

func TestDIP160_SilentWhenNoRequiredInputs(t *testing.T) {
	dir := writeWorkflows(t, map[string]string{
		"child.dip": `workflow Child
  goal: c
  start: A
  exit: A
  inputs
    tone: text
  agent A
    prompt:
      hi
`,
		"parent.dip": `workflow Parent
  goal: p
  start: Sub
  exit: Sub
  subgraph Sub
    ref: child.dip
`,
	})
	if n := countCode(arityDiags(t, dir, "parent.dip"), validator.DIP160); n != 0 {
		t.Errorf("child has no required inputs — want 0 DIP160, got %d", n)
	}
}

func TestDIP160_SilentWhenChildUnreadable(t *testing.T) {
	dir := writeWorkflows(t, map[string]string{
		"parent.dip": `workflow Parent
  goal: p
  start: Sub
  exit: Sub
  subgraph Sub
    ref: missing.dip
`,
	})
	if n := countCode(arityDiags(t, dir, "parent.dip"), validator.DIP160); n != 0 {
		t.Errorf("unreadable child — want 0 DIP160 (matches DIP146), got %d", n)
	}
}
