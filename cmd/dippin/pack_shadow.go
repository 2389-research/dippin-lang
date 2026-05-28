// ABOUTME: Shadow-tree preprocessing for `dippin pack` — inlines command_file:
// ABOUTME: directives at pack time so bundles are self-contained.
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/2389-research/dippin-lang/formatter"
	"github.com/2389-research/dippin-lang/ir"
	"github.com/2389-research/dippin-lang/parser"
)

// ensureUnderRoot rejects any absolute path that resolves outside rootDir.
// Guards the shadow-tree writer against a malicious subgraph ref like
// `subgraph: ../../../escape.dip`, which would otherwise cause
// writeShadowFile to write outside the temp shadow dir (since
// filepath.Join(shadowDir, "../../../escape.dip") escapes). dipx.Pack does
// its own escape check, but it runs AFTER we've built the shadow tree —
// too late to prevent the out-of-tree write.
func ensureUnderRoot(absPath, rootDir string) error {
	rel, err := filepath.Rel(rootDir, absPath)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return fmt.Errorf("pack-inline: ref escapes source root: %s", absPath)
	}
	return nil
}

// prepShadowSourceTree walks the entry workflow and every transitively-
// reachable subgraph .dip, resolves any command_file: directives by reading
// the referenced scripts from disk, then writes the inlined .dip sources
// into a fresh temp directory preserving the original relative layout.
//
// Returns the shadow path of the entry workflow and a cleanup func the
// caller must invoke (even on error from downstream Pack). The shadow
// tree is what dipx.Pack should walk: scripts referenced by command_file:
// are now embedded in the .dip text so the resulting bundle is self-
// contained and does not need the external script files to exist when
// the bundle is later opened.
//
// Architecture note: this preprocessing lives in the CLI layer because
// dipx must not import formatter or parser/resolve (CLAUDE.md loader-
// tier rule). dipx still walks the shadow tree using its own parser
// path; we only adjust the .dip text before handing it off.
func prepShadowSourceTree(entry string) (string, func(), error) {
	entryAbs, err := filepath.Abs(entry)
	if err != nil {
		return "", noop, err
	}
	rootDir := filepath.Dir(entryAbs)
	shadow, cleanup, err := makeShadowDir()
	if err != nil {
		return "", noop, err
	}
	if err := walkAndInline(entryAbs, rootDir, shadow); err != nil {
		cleanup()
		return "", noop, err
	}
	rel, err := filepath.Rel(rootDir, entryAbs)
	if err != nil {
		cleanup()
		return "", noop, err
	}
	return filepath.Join(shadow, rel), cleanup, nil
}

// noop is the cleanup returned when prep fails before any temp dir was made.
func noop() {}

// makeShadowDir creates a fresh temp dir for the shadow source tree and
// returns a cleanup func that removes it.
func makeShadowDir() (string, func(), error) {
	dir, err := os.MkdirTemp("", "dippin-pack-shadow-*")
	if err != nil {
		return "", noop, err
	}
	cleanup := func() { _ = os.RemoveAll(dir) }
	return dir, cleanup, nil
}

// walkAndInline performs a BFS over the entry workflow's subgraph ref graph,
// rooted at entryAbs. Each visited .dip is parsed, command_file: directives
// are resolved against the source file's directory, every tool node has its
// CommandFile field cleared (so the formatter emits inline command: blocks),
// the result is formatted, and written to the shadow tree at the same
// relative path as the original source.
func walkAndInline(entryAbs, rootDir, shadowDir string) error {
	visited := map[string]struct{}{}
	queue := []string{entryAbs}
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		if _, ok := visited[cur]; ok {
			continue
		}
		visited[cur] = struct{}{}
		refs, err := inlineOne(cur, rootDir, shadowDir)
		if err != nil {
			return err
		}
		queue = append(queue, refs...)
	}
	return nil
}

// inlineOne reads, parses, inlines, formats, and writes a single .dip into
// the shadow tree. Returns absolute paths to any subgraph refs the workflow
// declares so the caller can enqueue them for further processing.
func inlineOne(srcAbs, rootDir, shadowDir string) ([]string, error) {
	data, err := os.ReadFile(srcAbs)
	if err != nil {
		return nil, fmt.Errorf("pack-inline: read %s: %w", srcAbs, err)
	}
	wf, err := parser.NewParser(string(data), srcAbs).Parse()
	if err != nil {
		return nil, fmt.Errorf("pack-inline: parse %s: %w", srcAbs, err)
	}
	if err := parser.ResolveFileDirectives(wf, filepath.Dir(srcAbs)); err != nil {
		return nil, fmt.Errorf("pack-inline: %w", err)
	}
	clearFileDirectives(wf)
	if err := writeShadowFile(srcAbs, rootDir, shadowDir, wf); err != nil {
		return nil, err
	}
	return collectRefPaths(wf, filepath.Dir(srcAbs), rootDir)
}

// clearFileDirectives walks every node in wf and clears any file-directive
// source-path fields (ToolConfig.CommandFile, AgentConfig.PromptFile,
// AgentConfig.SystemPromptFile). The formatter emits the *_file: directive
// when these fields are set (preserving directive form on round-trip); we
// clear them after resolving so the formatter emits the inlined content
// blocks instead. Result: the shadow .dip carries inline content only,
// and the .dipx bundle is self-contained.
func clearFileDirectives(wf *ir.Workflow) {
	for _, n := range wf.Nodes {
		clearToolFileDirective(n)
		clearAgentFileDirectives(n)
	}
}

// clearToolFileDirective clears ToolConfig.CommandFile if set.
func clearToolFileDirective(n *ir.Node) {
	tc, ok := n.Config.(ir.ToolConfig)
	if !ok || tc.CommandFile == "" {
		return
	}
	tc.CommandFile = ""
	n.Config = tc
}

// clearAgentFileDirectives clears AgentConfig.PromptFile and
// AgentConfig.SystemPromptFile.
func clearAgentFileDirectives(n *ir.Node) {
	ac, ok := n.Config.(ir.AgentConfig)
	if !ok {
		return
	}
	ac.PromptFile = ""
	ac.SystemPromptFile = ""
	n.Config = ac
}

// writeShadowFile mirrors srcAbs's path under shadowDir (relative to rootDir)
// and writes the formatted workflow text to the resulting location.
func writeShadowFile(srcAbs, rootDir, shadowDir string, wf *ir.Workflow) error {
	rel, err := filepath.Rel(rootDir, srcAbs)
	if err != nil {
		return err
	}
	dst := filepath.Join(shadowDir, rel)
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	return os.WriteFile(dst, []byte(formatter.Format(wf)), 0o644)
}

// collectRefPaths returns absolute disk paths for every subgraph ref in wf,
// resolved against srcDir (the directory of the workflow that declared them).
// Each resolved target is validated to be under rootDir — refs that escape
// fail fast here (before BFS enqueues them) so the parent workflow's context
// is in the error message.
func collectRefPaths(wf *ir.Workflow, srcDir, rootDir string) ([]string, error) {
	var out []string
	for _, n := range wf.Nodes {
		ref := refFromNodeForShadow(n)
		if ref == "" {
			continue
		}
		target, err := filepath.Abs(filepath.Join(srcDir, ref))
		if err != nil {
			return nil, err
		}
		if err := ensureUnderRoot(target, rootDir); err != nil {
			return nil, err
		}
		out = append(out, target)
	}
	return out, nil
}

// refFromNodeForShadow mirrors dipx/helpers.go::refFromNode. Duplicated here
// because dipx exports the type-switch logic only through Pack's internal
// walker. Keep the two in sync: the set of ref-bearing node kinds must match
// so dipx.Pack's subgraph walk sees the same set we shadow.
func refFromNodeForShadow(n *ir.Node) string {
	switch cfg := n.Config.(type) {
	case ir.SubgraphConfig:
		return cfg.Ref
	case ir.ManagerLoopConfig:
		return cfg.SubgraphRef
	}
	return ""
}
