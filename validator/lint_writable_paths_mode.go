package validator

import (
	"fmt"

	"github.com/2389-research/dippin-lang/ir"
)

// The only two legal writable_paths_mode values. Matched EXACTLY — no trimming,
// no case-folding — because the runtime (tracker #648, invariant C1) fails
// closed on anything else at load time; the lint must reject what it rejects.
const (
	writablePathsModeRequire = "require"
	writablePathsModePrefer  = "prefer"
)

// lintWritablePathsMode fires the three writable_paths_mode checks (issue #307)
// on agent nodes and per-branch parallel overrides:
//   - DIP163 (error): value is not exactly "require" or "prefer".
//   - DIP164 (hint): a mode is set but no writable_paths gives it a scope.
//   - DIP165 (hint): "prefer" runs UNJAILED on hosts without Landlock ABI v3.
func lintWritablePathsMode(w *ir.Workflow) []Diagnostic {
	var diags []Diagnostic
	for _, n := range w.Nodes {
		diags = append(diags, checkNodeWritablePathsModeByKind(w, n)...)
	}
	return diags
}

func checkNodeWritablePathsModeByKind(w *ir.Workflow, n *ir.Node) []Diagnostic {
	switch cfg := n.Config.(type) {
	case ir.AgentConfig:
		hasScope := len(cfg.WritablePaths) > 0 || branchTargetsWithScope(w, n.ID)
		return checkWritablePathsModeObject(n, "", cfg.WritablePathsMode, hasScope)
	case ir.ParallelConfig:
		return checkBranchWritablePathsMode(w, n, cfg.Branches)
	default:
		return nil
	}
}

// checkBranchWritablePathsMode checks each branch override. A branch that
// declares no writable_paths of its own inherits the target agent's, so the
// DIP164 scope check consults the target too.
func checkBranchWritablePathsMode(w *ir.Workflow, n *ir.Node, branches []ir.BranchConfig) []Diagnostic {
	var diags []Diagnostic
	for _, b := range branches {
		hasScope := len(b.WritablePaths) > 0 || targetHasWritablePaths(w, b.Target)
		diags = append(diags, checkWritablePathsModeObject(n, b.Target, b.WritablePathsMode, hasScope)...)
	}
	return diags
}

// branchTargetsWithScope reports whether any block-form parallel branch targets
// the named agent, declares writable_paths of its own, and declares NO mode of
// its own. Such a branch inherits the agent's writable_paths_mode (tracker C1),
// so the agent's mode governs that branch's jail and is not inert even when the
// agent has no globs. A branch with its own mode does not use the agent's, so
// it gives it no scope. The mirror image of targetHasWritablePaths.
func branchTargetsWithScope(w *ir.Workflow, agentID string) bool {
	for _, n := range w.Nodes {
		if cfg, ok := n.Config.(ir.ParallelConfig); ok && anyBranchWithScope(cfg.Branches, agentID) {
			return true
		}
	}
	return false
}

func anyBranchWithScope(branches []ir.BranchConfig, agentID string) bool {
	for _, b := range branches {
		if b.Target == agentID && len(b.WritablePaths) > 0 && b.WritablePathsMode == "" {
			return true
		}
	}
	return false
}

// targetHasWritablePaths reports whether the named node is an agent that
// declares writable_paths. The mirror image of branchTargetsWithScope.
func targetHasWritablePaths(w *ir.Workflow, target string) bool {
	t := w.Node(target)
	if t == nil {
		return false
	}
	cfg, ok := t.Config.(ir.AgentConfig)
	return ok && len(cfg.WritablePaths) > 0
}

// checkWritablePathsModeObject runs the three checks on one declaration site
// (an agent when branch == "", else the named branch of a parallel node).
// An absent mode ("") is require by default and never fires anything.
func checkWritablePathsModeObject(n *ir.Node, branch, mode string, hasScope bool) []Diagnostic {
	if mode == "" {
		return nil
	}
	var diags []Diagnostic
	if !isLegalWritablePathsMode(mode) {
		diags = append(diags, dip163Diagnostic(n, branch, mode))
	}
	if !hasScope {
		diags = append(diags, dip164Diagnostic(n, branch))
	}
	if mode == writablePathsModePrefer {
		diags = append(diags, dip165Diagnostic(n, branch))
	}
	return diags
}

// isLegalWritablePathsMode is the exact-match check (tracker invariant C1).
func isLegalWritablePathsMode(mode string) bool {
	return mode == writablePathsModeRequire || mode == writablePathsModePrefer
}

// modeSubject names the declaration site: `node "X"` or `node "X" branch "b"`.
func modeSubject(n *ir.Node, branch string) string {
	if branch == "" {
		return fmt.Sprintf("node %q", n.ID)
	}
	return fmt.Sprintf("node %q branch %q", n.ID, branch)
}

func dip163Diagnostic(n *ir.Node, branch, mode string) Diagnostic {
	return Diagnostic{
		Code:     DIP163,
		Severity: SeverityError,
		Message:  fmt.Sprintf("%s has writable_paths_mode %q — the only legal values are exactly require and prefer (no case-folding, no surrounding whitespace); the runtime refuses to load this node", modeSubject(n, branch), mode),
		Location: n.Source,
		Help:     "write writable_paths_mode: require (refuse to start without Landlock) or writable_paths_mode: prefer (run UNJAILED without Landlock), or omit the field (absent = require).",
	}
}

func dip164Diagnostic(n *ir.Node, branch string) Diagnostic {
	return Diagnostic{
		Code:     DIP164,
		Severity: SeverityHint,
		Message:  fmt.Sprintf("%s sets writable_paths_mode but declares no writable_paths — a mode without a scope is inert (nothing to jail)", modeSubject(n, branch)),
		Location: n.Source,
		Help:     "add writable_paths: <globs> on the node, on a parallel branch that targets it, or (for a branch mode) on the branch or its target agent — so the mode has a jail to govern — or remove writable_paths_mode.",
	}
}

func dip165Diagnostic(n *ir.Node, branch string) Diagnostic {
	return Diagnostic{
		Code:     DIP165,
		Severity: SeverityHint,
		Message:  fmt.Sprintf("%s has writable_paths_mode: prefer — it runs UNJAILED on hosts without Landlock ABI v3 (macOS, Linux < 6.2); the write jail is best-effort there, not a guarantee, and operator copy must not describe this node as sandboxed", modeSubject(n, branch)),
		Location: n.Source,
		Help:     "keep prefer only if an unjailed run on those hosts is acceptable (the runtime records a jail_degraded event when it happens); use require (the default) to refuse to start instead.",
	}
}
