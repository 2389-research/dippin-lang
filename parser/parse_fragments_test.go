package parser

import (
	"strings"
	"testing"

	"github.com/2389-research/dippin-lang/ir"
)

func TestParseDefaults_PromptCascade(t *testing.T) {
	src := `workflow W
  start: A
  exit: A
  defaults
    prompt_suffix_file: frag/status.md
    prompt_prefix: "hello"
  agent A
    prompt: "x"
`
	w, err := NewParser(src, "t.dip").Parse()
	if err != nil {
		t.Fatal(err)
	}
	if w.Defaults.PromptSuffixFile != "frag/status.md" {
		t.Errorf("suffix_file=%q", w.Defaults.PromptSuffixFile)
	}
	if w.Defaults.PromptPrefix != "hello" {
		t.Errorf("prefix=%q", w.Defaults.PromptPrefix)
	}
}

func TestParseDefaults_PromptSuffixBothFormsError(t *testing.T) {
	src := `workflow W
  start: A
  exit: A
  defaults
    prompt_suffix: "x"
    prompt_suffix_file: f.md
  agent A
    prompt: "y"
`
	_, err := NewParser(src, "t.dip").Parse()
	if err == nil || !strings.Contains(err.Error(), "prompt_suffix") {
		t.Fatalf("want both-forms error, got %v", err)
	}
}

func TestParseAgent_PromptIncludeAndOptOut(t *testing.T) {
	src := `workflow W
  start: A
  exit: A
  agent A
    prompt: "x"
    prompt_include: frag/extra.md
    prompt_suffix: none
`
	w, err := NewParser(src, "t.dip").Parse()
	if err != nil {
		t.Fatal(err)
	}
	cfg := w.Nodes[0].Config.(ir.AgentConfig)
	if cfg.PromptInclude != "frag/extra.md" {
		t.Errorf("include=%q", cfg.PromptInclude)
	}
	if cfg.PromptSuffix != "none" {
		t.Errorf("suffix=%q", cfg.PromptSuffix)
	}
}
