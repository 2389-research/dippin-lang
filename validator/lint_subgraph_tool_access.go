package validator

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/2389-research/dippin-lang/ir"
)

// lintSubgraphToolAccess emits DIP143 (Hint): a node references an external
// subgraph (manager_loop subgraph_ref or subgraph ref) whose own .dip file does
// not inherit this workflow's tool_access restrictions. It fires only when the
// workflow shows containment intent (some node declares tool_access), and never
// parses the child file — the validator may not import the parser (layering
// rule). Real cross-file effective-access enforcement is tracked as #89.
func lintSubgraphToolAccess(w *ir.Workflow) []Diagnostic {
	if !WorkflowDeclaresToolAccess(w) {
		return nil
	}
	var diags []Diagnostic
	for _, n := range w.Nodes {
		if d := checkSubgraphBoundary(n); d != nil {
			diags = append(diags, *d)
		}
	}
	return diags
}

// WorkflowDeclaresToolAccess reports whether any node in w expresses tool_access
// containment intent (an agent or parallel branch with a non-empty tool_access).
// Exported so the CLI's cross-file pass (DIP146) reuses DIP143's exact intent
// logic — wasm-safe, pure IR.
func WorkflowDeclaresToolAccess(w *ir.Workflow) bool {
	for _, n := range w.Nodes {
		if NodeDeclaresToolAccess(n) {
			return true
		}
	}
	return false
}

// NodeDeclaresToolAccess reports whether an agent (or any parallel branch)
// declares a non-empty tool_access value.
func NodeDeclaresToolAccess(n *ir.Node) bool {
	switch cfg := n.Config.(type) {
	case ir.AgentConfig:
		return toolAccessSet(cfg.ToolAccess)
	case ir.ParallelConfig:
		return branchesDeclareToolAccess(cfg.Branches)
	default:
		return false
	}
}

func branchesDeclareToolAccess(branches []ir.BranchConfig) bool {
	for _, b := range branches {
		if toolAccessSet(b.ToolAccess) {
			return true
		}
	}
	return false
}

// toolAccessSet reports containment intent: any non-empty tool_access value.
// The runtime fails closed, so even an invalid value (a fail-closed typo)
// expresses and receives restriction — unlike DIP141's dead-config check,
// which deliberately matches only the canonical "none".
func toolAccessSet(s string) bool {
	return strings.TrimSpace(s) != ""
}

// checkSubgraphBoundary emits DIP143 for a node referencing an external subgraph.
func checkSubgraphBoundary(n *ir.Node) *Diagnostic {
	kind, ref := subgraphRefOf(n)
	if ref == "" || refIsSelf(ref, n.Source.File) {
		return nil
	}
	return &Diagnostic{
		Code:     DIP143,
		Severity: SeverityHint,
		Message: fmt.Sprintf(
			"%s %q references subgraph %q, defined in its own file; this workflow's tool_access restrictions do not extend across the subgraph boundary",
			kind, n.ID, ref),
		Location: n.Source,
		Help: fmt.Sprintf(
			"audit the agents in %q for their own tool_access — restrictions declared in this workflow do not propagate into a referenced subgraph. Native `dippin lint` resolves the child (DIP146); this advisory is the filesystem-free check or the fallback when the child cannot be resolved.",
			ref),
	}
}

// refIsSelf reports whether ref resolves to the node's own source file — a direct
// self-reference, where there is no cross-file boundary to warn about. Resolution
// mirrors resolveRefPath (relative to the source file's directory) but stays
// filepath-only so this lint compiles for wasm. Transitive cross-file cycles
// (A → B → A) require multi-file traversal and remain out of scope (#89).
func refIsSelf(ref, sourceFile string) bool {
	if sourceFile == "" {
		return false
	}
	resolved := ref
	if !filepath.IsAbs(ref) {
		resolved = filepath.Join(filepath.Dir(sourceFile), ref)
	}
	return filepath.Clean(resolved) == filepath.Clean(sourceFile)
}

// subgraphRefOf returns the node-kind label and the external subgraph reference
// for nodes that point at a separate .dip file, or ("", "") for other nodes.
func subgraphRefOf(n *ir.Node) (string, string) {
	switch cfg := n.Config.(type) {
	case ir.ManagerLoopConfig:
		return "manager_loop", cfg.SubgraphRef
	case ir.SubgraphConfig:
		return "subgraph", cfg.Ref
	default:
		return "", ""
	}
}
