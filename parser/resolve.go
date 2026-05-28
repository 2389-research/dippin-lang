package parser

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/2389-research/dippin-lang/ir"
)

const maxDirectiveFileSize = 4 << 20 // 4 MiB

// ResolveFileDirectives loads file contents for every tool node with
// CommandFile set, populating Command from the file's bytes. Paths are
// resolved relative to baseDir (typically the directory of the .dip
// source file). Returns the first error encountered.
//
// Parser entry points (NewParser/Parse) do NOT call this — they stay pure.
// CLI entry points call it after parsing. LSP and WASM contexts skip it;
// the IR retains the CommandFile-set / Command-empty state, which is the
// correct unresolved view for those consumers.
func ResolveFileDirectives(w *ir.Workflow, baseDir string) error {
	for _, n := range w.Nodes {
		if err := resolveNodeDirective(n, baseDir); err != nil {
			return err
		}
	}
	return nil
}

// resolveNodeDirective resolves any file-directive fields on a single node.
// Dispatches per node-config kind so the per-kind loader functions stay
// focused on their own field set.
func resolveNodeDirective(n *ir.Node, baseDir string) error {
	switch cfg := n.Config.(type) {
	case ir.ToolConfig:
		return resolveToolDirective(n, cfg, baseDir)
	case ir.AgentConfig:
		return resolveAgentDirective(n, cfg, baseDir)
	}
	return nil
}

// resolveToolDirective populates ToolConfig.Command from CommandFile, if set.
// Skips if Command is already inline-populated (parser's mutual-exclusion
// check should prevent both being set; defensive guard if it doesn't).
func resolveToolDirective(n *ir.Node, cfg ir.ToolConfig, baseDir string) error {
	if cfg.CommandFile == "" || cfg.Command != "" {
		return nil
	}
	contents, err := loadDirectiveFile(baseDir, cfg.CommandFile)
	if err != nil {
		return fmt.Errorf("node %q command_file: %w", n.ID, err)
	}
	cfg.Command = string(contents)
	n.Config = cfg
	return nil
}

// resolveAgentDirective populates Prompt and SystemPrompt from their *File
// twins on AgentConfig. The two slots are independent — either, both, or
// neither may be set.
func resolveAgentDirective(n *ir.Node, cfg ir.AgentConfig, baseDir string) error {
	if err := loadInto(&cfg.Prompt, cfg.PromptFile, baseDir, n.ID, "prompt_file"); err != nil {
		return err
	}
	if err := loadInto(&cfg.SystemPrompt, cfg.SystemPromptFile, baseDir, n.ID, "system_prompt_file"); err != nil {
		return err
	}
	n.Config = cfg
	return nil
}

// loadInto populates *dst from path (relative to baseDir) if path != "" and
// *dst == "". Skips if either condition fails (defensive: parser's mutual-
// exclusion check should prevent both being set, but if it happens, inline wins).
func loadInto(dst *string, path, baseDir, nodeID, directive string) error {
	if path == "" || *dst != "" {
		return nil
	}
	contents, err := loadDirectiveFile(baseDir, path)
	if err != nil {
		return fmt.Errorf("node %q %s: %w", nodeID, directive, err)
	}
	*dst = string(contents)
	return nil
}

// loadDirectiveFile resolves p relative to baseDir, applies path security
// checks, and reads the file. Error messages reference the user-written
// path (p), never the resolved absolute path, to avoid leaking directory
// structure into diagnostics. In particular, *fs.PathError values from
// os.Lstat / os.ReadFile carry the resolved absolute path in their
// stringification, so we never wrap them with %w — instead we branch on
// the error kind and emit a user-path-only message.
func loadDirectiveFile(baseDir, p string) ([]byte, error) {
	resolved, err := safeResolve(baseDir, p)
	if err != nil {
		return nil, err
	}
	info, err := statDirectiveFile(p, resolved)
	if err != nil {
		return nil, err
	}
	if err := checkFileInfo(p, info); err != nil {
		return nil, err
	}
	return readDirectiveFile(p, resolved)
}

// statDirectiveFile lstats the resolved path, rewriting any error so it
// only mentions the user-written path p (never the absolute resolved one).
func statDirectiveFile(p, resolved string) (os.FileInfo, error) {
	info, err := os.Lstat(resolved)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, fmt.Errorf("file %q not found", p)
		}
		return nil, fmt.Errorf("cannot stat file %q: not accessible", p)
	}
	return info, nil
}

// readDirectiveFile reads the resolved path, rewriting any error so it
// only mentions the user-written path p (never the absolute resolved one).
func readDirectiveFile(p, resolved string) ([]byte, error) {
	contents, err := os.ReadFile(resolved)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, fmt.Errorf("file %q not found", p)
		}
		return nil, fmt.Errorf("cannot read file %q: not accessible", p)
	}
	return contents, nil
}

// safeResolve joins baseDir/p and ensures the result stays under baseDir.
// Rejects absolute paths and any path that escapes via `..`.
func safeResolve(baseDir, p string) (string, error) {
	if filepath.IsAbs(p) {
		return "", fmt.Errorf("absolute paths not allowed: %q", p)
	}
	resolved := filepath.Join(baseDir, p)
	rel, err := filepath.Rel(baseDir, resolved)
	if err != nil || hasParentRef(rel) || filepath.IsAbs(rel) {
		return "", fmt.Errorf("path %q resolves outside source directory", p)
	}
	return resolved, nil
}

// checkFileInfo enforces symlink and size policies on the resolved file.
func checkFileInfo(p string, info os.FileInfo) error {
	if info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("symlinks not allowed: %q", p)
	}
	if info.Size() > maxDirectiveFileSize {
		return fmt.Errorf("file %q is too large (size %s, max %s)",
			p, formatMiB(info.Size()), formatMiB(maxDirectiveFileSize))
	}
	return nil
}

// formatMiB renders a byte count as a human-readable MiB string. The limit
// (4 MiB) is a whole-MiB value, so an integer cap renders cleanly; user-file
// sizes get one decimal place to distinguish, e.g., 5.0 MiB from 4.9 MiB.
func formatMiB(n int64) string {
	const mib = 1 << 20
	if n%mib == 0 {
		return fmt.Sprintf("%d MiB", n/mib)
	}
	return fmt.Sprintf("%.1f MiB", float64(n)/float64(mib))
}

// hasParentRef returns true if rel contains a `..` path segment.
// Used as defensive belt after filepath.Rel; should not fire on a
// non-symlink filesystem since Rel canonicalizes already.
func hasParentRef(rel string) bool {
	for _, seg := range strings.Split(filepath.ToSlash(rel), "/") {
		if seg == ".." {
			return true
		}
	}
	return false
}
