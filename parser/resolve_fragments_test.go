package parser

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/2389-research/dippin-lang/ir"
)

func TestResolve_CascadeAndInclude(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "suffix.md"), []byte("END WITH STATUS"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "extra.md"), []byte("EXTRA"), 0o644); err != nil {
		t.Fatal(err)
	}
	src := `workflow W
  start: A
  exit: B
  defaults
    prompt_suffix_file: suffix.md
  agent A
    prompt: "body A"
    prompt_include: extra.md
  agent B
    prompt: "body B"
    prompt_suffix: none
`
	w, err := NewParser(src, filepath.Join(dir, "w.dip")).Parse()
	if err != nil {
		t.Fatal(err)
	}
	if err := ResolveFileDirectives(w, dir); err != nil {
		t.Fatal(err)
	}
	if got := w.Nodes[0].Config.(ir.AgentConfig).Prompt; got != "body A\n\nEXTRA\n\nEND WITH STATUS" {
		t.Errorf("A prompt=%q", got)
	}
	if got := w.Nodes[1].Config.(ir.AgentConfig).Prompt; got != "body B" {
		t.Errorf("B (opted out) prompt=%q", got)
	}
}

func TestResolve_FragmentRejectsSymlink(t *testing.T) {
	tmp := t.TempDir()
	external := filepath.Join(t.TempDir(), "outside.md")
	if err := os.WriteFile(external, []byte("external"), 0o644); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(tmp, "frag.md")
	if err := os.Symlink(external, link); err != nil {
		t.Skipf("symlink unsupported: %v", err)
	}
	w := &ir.Workflow{
		Defaults: ir.WorkflowDefaults{PromptSuffixFile: "frag.md"},
		Nodes:    []*ir.Node{{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "x"}}},
	}
	err := ResolveFileDirectives(w, tmp)
	if err == nil || !strings.Contains(err.Error(), "symlinks not allowed") {
		t.Errorf("expected symlink rejection for prompt_suffix_file; got %v", err)
	}
}
