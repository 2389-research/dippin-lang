package parser

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/2389-research/dippin-lang/ir"
)

// TestResolve_CascadeSkipsBodylessPassthrough covers #248: a defaults prompt
// cascade must not synthesize a prompt on a body-less passthrough agent (e.g. a
// declared start:/exit: node), which would turn it into a real LLM call. A
// body-having agent still gets the cascade.
func TestResolve_CascadeSkipsBodylessPassthrough(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "frag.md"), []byte("SHARED PREFIX"), 0o644); err != nil {
		t.Fatal(err)
	}
	src := `workflow W
  start: S
  exit: E
  defaults
    prompt_prefix_file: frag.md
  agent S
    label: Start
  agent Work
    prompt: "do it"
  agent E
    label: End
  edges
    S -> Work
    Work -> E
`
	w, err := NewParser(src, filepath.Join(dir, "w.dip")).Parse()
	if err != nil {
		t.Fatal(err)
	}
	if err := ResolveFileDirectives(w, dir); err != nil {
		t.Fatal(err)
	}

	prompt := func(i int) string { return w.Nodes[i].Config.(ir.AgentConfig).Prompt }
	if got := prompt(0); got != "" { // agent S (start)
		t.Errorf("body-less start node got a synthesized prompt %q; cascade must skip it (#248)", got)
	}
	if got := prompt(2); got != "" { // agent E (exit)
		t.Errorf("body-less exit node got a synthesized prompt %q; cascade must skip it (#248)", got)
	}
	if got := prompt(1); !strings.Contains(got, "SHARED PREFIX") || !strings.Contains(got, "do it") { // agent Work
		t.Errorf("body-having node should still receive the cascade, got %q", got)
	}
}

// TestResolve_CascadeSkipsBodylessButKeepsInclude confirms a node with only a
// prompt_include (no own prompt) is NOT body-less — the cascade still applies.
func TestResolve_CascadeSkipsBodylessButKeepsInclude(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "frag.md"), []byte("PREFIX"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "inc.md"), []byte("INCLUDED"), 0o644); err != nil {
		t.Fatal(err)
	}
	src := `workflow W
  start: A
  exit: A
  defaults
    prompt_prefix_file: frag.md
  agent A
    prompt_include: inc.md
`
	w, err := NewParser(src, filepath.Join(dir, "w.dip")).Parse()
	if err != nil {
		t.Fatal(err)
	}
	if err := ResolveFileDirectives(w, dir); err != nil {
		t.Fatal(err)
	}
	got := w.Nodes[0].Config.(ir.AgentConfig).Prompt
	if !strings.Contains(got, "PREFIX") || !strings.Contains(got, "INCLUDED") {
		t.Errorf("a node with an include is not body-less; cascade should apply, got %q", got)
	}
}
