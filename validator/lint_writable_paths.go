package validator

import (
	"fmt"
	"strings"

	"github.com/2389-research/dippin-lang/ir"
)

// lintWritablePaths fires DIP141 (writable_paths nullified by tool_access: none)
// and DIP142 (unsafe writable_paths entry) on agent nodes and per-branch overrides.
// dippin carries + lints the field; the tracker enforces the fs-level write jail.
func lintWritablePaths(w *ir.Workflow) []Diagnostic {
	var diags []Diagnostic
	for _, n := range w.Nodes {
		diags = append(diags, checkNodeWritablePathsByKind(n)...)
	}
	return diags
}

func checkNodeWritablePathsByKind(n *ir.Node) []Diagnostic {
	switch cfg := n.Config.(type) {
	case ir.AgentConfig:
		return checkWritablePathsObject(n, cfg.WritablePaths, cfg.ToolAccess, "")
	case ir.ParallelConfig:
		return checkBranchWritablePaths(n, cfg.Branches)
	default:
		return nil
	}
}

func checkBranchWritablePaths(n *ir.Node, branches []ir.BranchConfig) []Diagnostic {
	var diags []Diagnostic
	for _, b := range branches {
		diags = append(diags, checkWritablePathsObject(n, b.WritablePaths, b.ToolAccess, b.Target)...)
	}
	return diags
}

func checkWritablePathsObject(n *ir.Node, paths []string, toolAccess, branch string) []Diagnostic {
	if len(paths) == 0 {
		return nil
	}
	var diags []Diagnostic
	if strings.ToLower(strings.TrimSpace(toolAccess)) == "none" {
		diags = append(diags, dip141Diagnostic(n, branch))
	}
	return diags
}

func dip141Diagnostic(n *ir.Node, branch string) Diagnostic {
	msg := fmt.Sprintf("node %q has writable_paths but tool_access \"none\" — none strips all tools, so there is nothing to bound (dead config)", n.ID)
	if branch != "" {
		msg = fmt.Sprintf("node %q branch %q has writable_paths but tool_access \"none\" — none strips all tools, so there is nothing to bound (dead config)", n.ID, branch)
	}
	return Diagnostic{
		Code:     DIP141,
		Severity: SeverityWarning,
		Message:  msg,
		Location: n.Source,
		Help:     "remove writable_paths (no tools to bound) or drop tool_access: none to grant a bounded tool catalog.",
	}
}
