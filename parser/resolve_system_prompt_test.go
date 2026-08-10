package parser

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/2389-research/dippin-lang/ir"
)

// agentSystemPrompt returns node i's resolved SystemPrompt.
func agentSystemPrompt(t *testing.T, w *ir.Workflow, i int) string {
	t.Helper()
	return w.Nodes[i].Config.(ir.AgentConfig).SystemPrompt
}

// TestResolve_DefaultsSystemPromptCascade covers #72: a defaults-block
// system_prompt_file is a fallback default. An agent with no system prompt of
// its own inherits it; an agent with its own inline system_prompt, or its own
// system_prompt_file, keeps that and never sees the default.
func TestResolve_DefaultsSystemPromptCascade(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "persona.md"), []byte("SHARED PERSONA"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "own.md"), []byte("OWN PERSONA"), 0o644); err != nil {
		t.Fatal(err)
	}
	src := `workflow W
  start: Inherit
  exit: OwnFile
  defaults
    system_prompt_file: persona.md
  agent Inherit
    prompt: "a"
  agent OwnInline
    prompt: "b"
    system_prompt: "OWN INLINE"
  agent OwnFile
    prompt: "c"
    system_prompt_file: own.md
`
	w, err := NewParser(src, filepath.Join(dir, "w.dip")).Parse()
	if err != nil {
		t.Fatal(err)
	}
	if w.Defaults.SystemPromptFile != "persona.md" {
		t.Fatalf("defaults.SystemPromptFile = %q, want persona.md", w.Defaults.SystemPromptFile)
	}
	if err := ResolveFileDirectives(w, dir); err != nil {
		t.Fatal(err)
	}
	if got := agentSystemPrompt(t, w, 0); got != "SHARED PERSONA" {
		t.Errorf("Inherit system prompt = %q, want the shared default", got)
	}
	if got := agentSystemPrompt(t, w, 1); got != "OWN INLINE" {
		t.Errorf("OwnInline system prompt = %q, want its own inline value (default not applied)", got)
	}
	if got := agentSystemPrompt(t, w, 2); got != "OWN PERSONA" {
		t.Errorf("OwnFile system prompt = %q, want its own file (default not applied)", got)
	}
}

// TestResolve_DefaultsSystemPromptMissingFileErrors confirms a missing default
// file errors at resolve time, like any other file directive.
func TestResolve_DefaultsSystemPromptMissingFileErrors(t *testing.T) {
	w := &ir.Workflow{
		Defaults: ir.WorkflowDefaults{SystemPromptFile: "nope.md"},
		Nodes:    []*ir.Node{{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "x"}}},
	}
	err := ResolveFileDirectives(w, t.TempDir())
	if err == nil || !strings.Contains(err.Error(), "system_prompt_file") {
		t.Errorf("expected a system_prompt_file resolve error, got %v", err)
	}
}

// TestParse_DefaultsInlineSystemPromptRejected confirms only the file form is
// accepted under defaults; an inline system_prompt is an unknown field.
func TestParse_DefaultsInlineSystemPromptRejected(t *testing.T) {
	src := `workflow W
  start: A
  exit: A
  defaults
    system_prompt: "inline persona"
  agent A
    prompt: "a"
`
	_, err := NewParser(src, "w.dip").Parse()
	if err == nil || !strings.Contains(err.Error(), "unknown defaults field") {
		t.Errorf("inline system_prompt under defaults should be an unknown-field error, got %v", err)
	}
}
