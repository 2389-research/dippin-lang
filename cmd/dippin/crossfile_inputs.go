// ABOUTME: DIP160 — cross-file check that a subgraph node's params: provides
// ABOUTME: every input the referenced child workflow declares as required.
package main

import (
	"fmt"
	"path/filepath"

	"github.com/2389-research/dippin-lang/ir"
	"github.com/2389-research/dippin-lang/validator"
)

// applyCrossFileChecks runs every on-disk cross-file pass (DIP146 tool-access
// supersession and DIP160 input arity) over the entry's diagnostics. It is the
// single seam the CLI's validate/check/watch paths call. Skipped for .dipx
// bundles, whose child refs are in-bundle paths, not on disk.
func applyCrossFileChecks(diags []validator.Diagnostic, w *ir.Workflow, path string) []validator.Diagnostic {
	diags = applyCrossFileToolAccess(diags, w, path)
	if filepath.Ext(path) == ".dipx" {
		return diags
	}
	return append(diags, crossFileInputArity(w, path)...)
}

// crossFileInputArity walks every subgraph node in the entry workflow, resolves
// its referenced child, and emits DIP160 for each input the child declares
// required: true that the parent's params: does not provide. Like the DIP146
// cross-file pass it reads child files on disk (so it is skipped for .dipx and
// when a child cannot be read/parsed), reusing the same root-containment
// resolution.
func crossFileInputArity(entry *ir.Workflow, entryPath string) []validator.Diagnostic {
	root := filepath.Dir(absOrClean(entryPath))
	var diags []validator.Diagnostic
	for _, n := range entry.Nodes {
		diags = append(diags, subgraphInputArity(n, root)...)
	}
	return diags
}

// subgraphInputArity checks one node: if it is a subgraph with a resolvable
// child, flag each required child input missing from the node's params:.
func subgraphInputArity(n *ir.Node, root string) []validator.Diagnostic {
	cfg, ok := n.Config.(ir.SubgraphConfig)
	if !ok || cfg.Ref == "" || n.Source.File == "" {
		return nil
	}
	child, _ := resolveBoundaryChild(n, cfg.Ref, root)
	if child == nil {
		return nil // unreadable/unparseable child — nothing to check (matches DIP146)
	}
	return missingRequiredInputs(n, cfg, child)
}

// missingRequiredInputs flags each required child input the parent's params omits.
func missingRequiredInputs(n *ir.Node, cfg ir.SubgraphConfig, child *ir.Workflow) []validator.Diagnostic {
	var diags []validator.Diagnostic
	for _, in := range child.Inputs {
		if _, provided := cfg.Params[in.Name]; in.Required && !provided {
			diags = append(diags, missingRequiredInputDiag(n, cfg.Ref, in.Name))
		}
	}
	return diags
}

// missingRequiredInputDiag builds a DIP160 warning for one omitted required input.
func missingRequiredInputDiag(n *ir.Node, ref, name string) validator.Diagnostic {
	return validator.Diagnostic{
		Code:     validator.DIP160,
		Severity: validator.SeverityWarning,
		Message: fmt.Sprintf("subgraph %q omits required input %q of %q — the child starts with it unset",
			n.ID, name, ref),
		Location: n.Source,
		Help:     fmt.Sprintf("add %q to this subgraph's params:, or give the child input a default", name),
	}
}
