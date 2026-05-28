package parser

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/2389-research/dippin-lang/ir"
)

func TestResolveFileDirectives_LoadsContent(t *testing.T) {
	w := &ir.Workflow{
		Nodes: []*ir.Node{
			{ID: "A", Kind: ir.NodeTool, Config: ir.ToolConfig{
				CommandFile: "setup.sh",
			}},
		},
	}
	baseDir := "testdata/command_file"
	if err := ResolveFileDirectives(w, baseDir); err != nil {
		t.Fatalf("ResolveFileDirectives: %v", err)
	}
	cfg := w.Nodes[0].Config.(ir.ToolConfig)
	if !strings.Contains(cfg.Command, "fixture: ResolveFileDirectives test") {
		t.Errorf("Command not populated from file; got %q", cfg.Command)
	}
	if cfg.CommandFile != "setup.sh" {
		t.Errorf("CommandFile = %q, want %q", cfg.CommandFile, "setup.sh")
	}
}

func TestResolveFileDirectives_SkipsInline(t *testing.T) {
	w := &ir.Workflow{
		Nodes: []*ir.Node{
			{ID: "A", Kind: ir.NodeTool, Config: ir.ToolConfig{
				Command: "inline content",
			}},
		},
	}
	if err := ResolveFileDirectives(w, "testdata/command_file"); err != nil {
		t.Fatalf("ResolveFileDirectives: %v", err)
	}
	cfg := w.Nodes[0].Config.(ir.ToolConfig)
	if cfg.Command != "inline content" {
		t.Errorf("inline Command modified: %q", cfg.Command)
	}
}

func TestResolveFileDirectives_RejectsAbsolutePath(t *testing.T) {
	w := &ir.Workflow{
		Nodes: []*ir.Node{
			{ID: "A", Kind: ir.NodeTool, Config: ir.ToolConfig{
				CommandFile: "/etc/passwd",
			}},
		},
	}
	err := ResolveFileDirectives(w, "testdata/command_file")
	if err == nil || !strings.Contains(err.Error(), "absolute paths not allowed") {
		t.Errorf("expected absolute-path rejection; got %v", err)
	}
	// Error must reference the user-written path
	if !strings.Contains(err.Error(), "/etc/passwd") {
		t.Errorf("error should reference user-written path /etc/passwd; got %v", err)
	}
}

func TestResolveFileDirectives_RejectsParentEscape(t *testing.T) {
	w := &ir.Workflow{
		Nodes: []*ir.Node{
			{ID: "A", Kind: ir.NodeTool, Config: ir.ToolConfig{
				CommandFile: "../../../etc/passwd",
			}},
		},
	}
	err := ResolveFileDirectives(w, "testdata/command_file")
	if err == nil || !strings.Contains(err.Error(), "resolves outside source directory") {
		t.Errorf("expected parent-escape rejection; got %v", err)
	}
}

func TestResolveFileDirectives_RejectsPostCleanEscape(t *testing.T) {
	w := &ir.Workflow{
		Nodes: []*ir.Node{
			{ID: "A", Kind: ir.NodeTool, Config: ir.ToolConfig{
				CommandFile: "legit/../../../etc/passwd",
			}},
		},
	}
	err := ResolveFileDirectives(w, "testdata/command_file")
	if err == nil || !strings.Contains(err.Error(), "resolves outside source directory") {
		t.Errorf("expected post-Clean escape rejection; got %v", err)
	}
}

func TestResolveFileDirectives_RejectsSymlink(t *testing.T) {
	tmp := t.TempDir()
	external := filepath.Join(t.TempDir(), "outside.txt")
	if err := os.WriteFile(external, []byte("external"), 0o644); err != nil {
		t.Fatalf("write external: %v", err)
	}
	linkPath := filepath.Join(tmp, "link.sh")
	if err := os.Symlink(external, linkPath); err != nil {
		t.Skipf("symlink not supported on this platform: %v", err)
	}
	w := &ir.Workflow{
		Nodes: []*ir.Node{
			{ID: "A", Kind: ir.NodeTool, Config: ir.ToolConfig{
				CommandFile: "link.sh",
			}},
		},
	}
	err := ResolveFileDirectives(w, tmp)
	if err == nil || !strings.Contains(err.Error(), "symlinks not allowed") {
		t.Errorf("expected symlink rejection; got %v", err)
	}
}

func TestResolveFileDirectives_RejectsOversize(t *testing.T) {
	tmp := t.TempDir()
	bigPath := filepath.Join(tmp, "big.sh")
	// Write 4 MiB + 1 byte
	big := make([]byte, (4<<20)+1)
	if err := os.WriteFile(bigPath, big, 0o644); err != nil {
		t.Fatalf("write big: %v", err)
	}
	w := &ir.Workflow{
		Nodes: []*ir.Node{
			{ID: "A", Kind: ir.NodeTool, Config: ir.ToolConfig{
				CommandFile: "big.sh",
			}},
		},
	}
	err := ResolveFileDirectives(w, tmp)
	if err == nil || !strings.Contains(err.Error(), "too large") {
		t.Errorf("expected oversize rejection; got %v", err)
	}
	if err != nil && !strings.Contains(err.Error(), "max 4 MiB") {
		t.Errorf("expected size cap rendered in MiB; got %v", err)
	}
}

func TestResolveFileDirectives_MissingFile(t *testing.T) {
	// Use an absolute baseDir so we can detect leaks of the resolved
	// absolute path in error messages. The CLI typically resolves to
	// absolute paths before invoking ResolveFileDirectives.
	absBase, err := filepath.Abs("testdata/command_file")
	if err != nil {
		t.Fatalf("filepath.Abs: %v", err)
	}
	w := &ir.Workflow{
		Nodes: []*ir.Node{
			{ID: "A", Kind: ir.NodeTool, Config: ir.ToolConfig{
				CommandFile: "nonexistent.sh",
			}},
		},
	}
	err = ResolveFileDirectives(w, absBase)
	if err == nil {
		t.Fatalf("expected missing-file error; got nil")
	}
	// Error must reference the user-written path
	if !strings.Contains(err.Error(), "nonexistent.sh") {
		t.Errorf("error should reference user-written path; got %v", err)
	}
	// Error must NOT contain the resolved absolute path (information leak).
	if strings.Contains(err.Error(), absBase) {
		t.Errorf("error leaked resolved absolute path %q in message: %v", absBase, err)
	}
}

func TestResolveFileDirectives_LoadsAgentPrompt(t *testing.T) {
	w := &ir.Workflow{
		Nodes: []*ir.Node{
			{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{
				PromptFile: "task.md",
			}},
		},
	}
	if err := ResolveFileDirectives(w, "testdata/prompt_file"); err != nil {
		t.Fatalf("ResolveFileDirectives: %v", err)
	}
	cfg := w.Nodes[0].Config.(ir.AgentConfig)
	if !strings.Contains(cfg.Prompt, "fixture: ResolveFileDirectives prompt test") {
		t.Errorf("Prompt not populated from file; got %q", cfg.Prompt)
	}
	if cfg.PromptFile != "task.md" {
		t.Errorf("PromptFile = %q, want %q (must be preserved post-resolve)", cfg.PromptFile, "task.md")
	}
}

func TestResolveFileDirectives_LoadsAgentSystemPrompt(t *testing.T) {
	w := &ir.Workflow{
		Nodes: []*ir.Node{
			{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{
				SystemPromptFile: "persona.md",
			}},
		},
	}
	if err := ResolveFileDirectives(w, "testdata/prompt_file"); err != nil {
		t.Fatalf("ResolveFileDirectives: %v", err)
	}
	cfg := w.Nodes[0].Config.(ir.AgentConfig)
	if !strings.Contains(cfg.SystemPrompt, "fixture: ResolveFileDirectives system_prompt test") {
		t.Errorf("SystemPrompt not populated from file; got %q", cfg.SystemPrompt)
	}
	if cfg.SystemPromptFile != "persona.md" {
		t.Errorf("SystemPromptFile = %q, want %q", cfg.SystemPromptFile, "persona.md")
	}
}

func TestResolveFileDirectives_LoadsBothAgentSlots(t *testing.T) {
	w := &ir.Workflow{
		Nodes: []*ir.Node{
			{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{
				PromptFile:       "task.md",
				SystemPromptFile: "persona.md",
			}},
		},
	}
	if err := ResolveFileDirectives(w, "testdata/prompt_file"); err != nil {
		t.Fatalf("ResolveFileDirectives: %v", err)
	}
	cfg := w.Nodes[0].Config.(ir.AgentConfig)
	if !strings.Contains(cfg.Prompt, "ResolveFileDirectives prompt test") {
		t.Errorf("Prompt not populated; got %q", cfg.Prompt)
	}
	if !strings.Contains(cfg.SystemPrompt, "ResolveFileDirectives system_prompt test") {
		t.Errorf("SystemPrompt not populated; got %q", cfg.SystemPrompt)
	}
}

func TestResolveFileDirectives_AgentErrorIdentifiesDirective(t *testing.T) {
	w := &ir.Workflow{
		Nodes: []*ir.Node{
			{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{
				SystemPromptFile: "nonexistent.md",
			}},
		},
	}
	err := ResolveFileDirectives(w, "testdata/prompt_file")
	if err == nil {
		t.Fatal("expected missing-file error, got nil")
	}
	if !strings.Contains(err.Error(), "system_prompt_file") {
		t.Errorf("error should identify directive `system_prompt_file`; got %v", err)
	}
	// Substring `prompt_file:` matches both directives; the bare directive token
	// is preceded by a space in the error format, so we use that as the anchor.
	if strings.Contains(err.Error(), " prompt_file:") {
		t.Errorf("error must not be ambiguous between prompt_file and system_prompt_file; got %v", err)
	}
}

func TestResolveFileDirectives_ExternalPromptsExample(t *testing.T) {
	// Pins end-to-end resolution of examples/external_prompts.dip.
	// Mirrors the equivalent integration test for examples/external_files.dip
	// that #52 added.
	srcAbs, err := filepath.Abs("../examples/external_prompts.dip")
	if err != nil {
		t.Fatalf("Abs: %v", err)
	}
	data, err := os.ReadFile(srcAbs)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	wf, err := NewParser(string(data), srcAbs).Parse()
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if err := ResolveFileDirectives(wf, filepath.Dir(srcAbs)); err != nil {
		t.Fatalf("ResolveFileDirectives: %v", err)
	}
	var reviewer ir.AgentConfig
	for _, n := range wf.Nodes {
		if n.ID == "Reviewer" {
			reviewer = n.Config.(ir.AgentConfig)
		}
	}
	if !strings.Contains(reviewer.SystemPrompt, "senior code reviewer") {
		t.Errorf("SystemPrompt not loaded from file; got %q", reviewer.SystemPrompt)
	}
	if !strings.Contains(reviewer.Prompt, "STATUS: success") {
		t.Errorf("Prompt not loaded from file; got %q", reviewer.Prompt)
	}
}
