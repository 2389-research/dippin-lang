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
	if err == nil || !strings.Contains(err.Error(), "exceeds") {
		t.Errorf("expected oversize rejection; got %v", err)
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
