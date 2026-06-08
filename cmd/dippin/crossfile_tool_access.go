package main

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/2389-research/dippin-lang/dipx"
	"github.com/2389-research/dippin-lang/ir"
	"github.com/2389-research/dippin-lang/validator"
)

// crossFileMaxDepth bounds subgraph-ref recursion (mirrors dipx's
// maxManifestDepth). On hitting the cap the pass stops recursing further;
// boundaries deeper than the cap are never visited (neither classified nor
// flagged). A backstop against pathological depth, not a real-workflow limit.
const crossFileMaxDepth = 32

// childPosture classifies a referenced child workflow's tool_access stance.
// The zero value is postureUnresolved, so a missing classification never
// supersedes a DIP143 advisory.
type childPosture int

const (
	postureUnresolved   childPosture = iota // missing / unparseable / depth-capped
	postureZeroIntent                       // >=1 agent, nothing restricts tools
	postureFullRestrict                     // >=1 agent, every agent restricted
	posturePartial                          // some restricted, >=1 agent still open
	postureAgentless                        // no agent nodes (no tools to grant)
)

// crossFileToolAccess walks every subgraph/manager_loop boundary reachable from
// entry, emitting DIP146 (Hint) when a workflow on the path restricts tools and a
// resolved child restricts none. It also returns, per boundary-node location, the
// child's posture, so CmdLint can supersede the per-file DIP143 advisory. entryPath
// is the on-disk path entry was parsed from (seeds the visited set).
func crossFileToolAccess(entry *ir.Workflow, entryPath string) ([]validator.Diagnostic, map[ir.SourceLocation]childPosture) {
	var diags []validator.Diagnostic
	classified := map[ir.SourceLocation]childPosture{}
	visited := map[string]bool{}
	if key := canonicalKey(entryPath); key != "" {
		visited[key] = true
	}
	intentSeen := validator.WorkflowDeclaresToolAccess(entry)
	// root is the containment anchor for every child ref, captured ONCE from the
	// entry's directory and threaded unchanged (never recomputed per-parent — see
	// spec D2/C2). absOrClean makes it absolute when possible (it falls back to a
	// lexical Clean only if filepath.Abs fails, e.g. an unavailable cwd) so
	// ensureUnderRoot's filepath.Rel can compare it against absolute child targets;
	// in the fallback case a relative root simply fails the Rel check and refuses,
	// which is the safe (fail-soft) direction.
	root := filepath.Dir(absOrClean(entryPath))
	walkBoundaries(entry, intentSeen, 0, root, visited, &diags, classified)
	return diags, classified
}

// walkBoundaries inspects each subgraph/manager_loop boundary in w.
func walkBoundaries(w *ir.Workflow, intentSeen bool, depth int, root string, visited map[string]bool, diags *[]validator.Diagnostic, classified map[ir.SourceLocation]childPosture) {
	for _, n := range w.Nodes {
		_, ref := boundaryKindRef(n)
		if ref == "" || n.Source.File == "" {
			continue
		}
		visitBoundary(n, ref, intentSeen, depth, root, visited, diags, classified)
	}
}

// visitBoundary resolves one boundary's child, records its posture, emits DIP146
// when the path shows intent and the child is zero-intent, then recurses.
func visitBoundary(n *ir.Node, ref string, intentSeen bool, depth int, root string, visited map[string]bool, diags *[]validator.Diagnostic, classified map[ir.SourceLocation]childPosture) {
	child, childPath := resolveBoundaryChild(n, ref, root)
	if child == nil {
		classified[n.Source] = postureUnresolved
		return
	}
	posture := classifyChild(child)
	classified[n.Source] = posture
	if intentSeen && posture == postureZeroIntent {
		*diags = append(*diags, boundaryDiag(n, ref))
	}
	maybeRecurse(child, childPath, intentSeen, depth, root, visited, diags, classified)
}

// maybeRecurse descends into a child's own boundaries unless it was already
// visited (cycle guard) or the depth cap is reached. The child is marked visited
// before recursing (pre-order) so cycles terminate.
func maybeRecurse(child *ir.Workflow, childPath string, intentSeen bool, depth int, root string, visited map[string]bool, diags *[]validator.Diagnostic, classified map[ir.SourceLocation]childPosture) {
	key := canonicalKey(childPath)
	if key == "" || visited[key] || depth+1 > crossFileMaxDepth {
		return
	}
	visited[key] = true
	childIntent := intentSeen || validator.WorkflowDeclaresToolAccess(child)
	walkBoundaries(child, childIntent, depth+1, root, visited, diags, classified)
}

// resolveBoundaryChild resolves ref relative to the boundary node's source file,
// refusing root-escaping or symlinked targets, then parses the child workflow.
// Fail-soft: ANY error (escape, symlink, read, parse) yields (nil, ""), routing
// the boundary to postureUnresolved so the per-file DIP143 advisory is retained.
// This is detection parity with the pack walker (Lstat-based, not the parser's
// stronger O_NOFOLLOW open-once — the residual leaf TOCTOU matches pack's; a lint
// emits only DIP codes, never file content). Callers (walkBoundaries) guarantee
// ref != "" and n.Source.File != "".
func resolveBoundaryChild(n *ir.Node, ref, root string) (*ir.Workflow, string) {
	path, err := resolveChildPath(n.Source.File, ref, root)
	if err != nil {
		return nil, ""
	}
	data, err := dipx.ReadNoFollowSymlinks(path, root)
	if err != nil {
		return nil, ""
	}
	w, err := parseAndResolveDip(data, path)
	if err != nil {
		return nil, ""
	}
	return w, path
}

// resolveChildPath joins ref against parentFile's directory and refuses any result
// that escapes root. Pack-identical (mirrors dipx's resolveRefOnDisk): filepath.Join
// swallows the leading slash of an absolute ref, so abs refs are RE-ROOTED under
// root, not read out-of-root. The filepath.Abs is load-bearing — at the entry level
// parentFile (n.Source.File) may be relative while root is absolute, and
// ensureUnderRoot's filepath.Rel needs both absolute. Do not simplify it away.
func resolveChildPath(parentFile, ref, root string) (string, error) {
	target := filepath.Clean(filepath.Join(filepath.Dir(parentFile), ref))
	abs, err := filepath.Abs(target)
	if err != nil {
		return "", err
	}
	if err := ensureUnderRoot(abs, root); err != nil {
		return "", err
	}
	return abs, nil
}

// classifyChild determines a child's tool_access posture from its agent nodes.
func classifyChild(child *ir.Workflow) childPosture {
	agents, restricted := countAgents(child)
	switch {
	case agents == 0:
		return postureAgentless
	case !validator.WorkflowDeclaresToolAccess(child):
		return postureZeroIntent
	case restricted == agents:
		return postureFullRestrict
	default:
		return posturePartial
	}
}

// countAgents returns the number of agent nodes and how many declare a tool_access
// restriction. The census is intentionally AgentConfig-only. A parallel branch's
// tool_access is a PATH-LOCAL override (effective = branch if non-empty else agent),
// so a branch "none" does NOT guarantee its target agent never runs unrestricted via
// a direct edge. Counting a branch as restricting its agent would let classifyChild
// return postureFullRestrict and drop DIP143 for a child whose agent can still run
// with full tools — the false assurance this feature exists to prevent. Branch intent
// IS still honored by validator.WorkflowDeclaresToolAccess, which keeps such a child
// out of postureZeroIntent — so it classifies as posturePartial (DIP143 retained, no
// DIP146), the correct conservative outcome. Do not "simplify" this to count branches.
func countAgents(child *ir.Workflow) (agents, restricted int) {
	for _, n := range child.Nodes {
		if _, ok := n.Config.(ir.AgentConfig); !ok {
			continue
		}
		agents++
		if validator.NodeDeclaresToolAccess(n) {
			restricted++
		}
	}
	return agents, restricted
}

// canonicalKey returns a symlink-resolved absolute key for the visited set, so a
// file reached via different (possibly relative) spellings — or a symlink cycle —
// maps to one key. EvalSymlinks may return a relative path for relative input, so
// the result is always run through absOrClean to keep keys comparable.
func canonicalKey(path string) string {
	if path == "" {
		return ""
	}
	if resolved, err := filepath.EvalSymlinks(path); err == nil {
		return absOrClean(resolved)
	}
	return absOrClean(path)
}

// absOrClean returns the absolute form of path, falling back to a lexical clean
// when the working directory is unavailable.
func absOrClean(path string) string {
	if abs, err := filepath.Abs(path); err == nil {
		return abs
	}
	return filepath.Clean(path)
}

// boundaryKindRef returns the node-kind label ("manager_loop"/"subgraph") and the
// external ref for boundary nodes, or ("","") otherwise.
func boundaryKindRef(n *ir.Node) (string, string) {
	switch cfg := n.Config.(type) {
	case ir.ManagerLoopConfig:
		return "manager_loop", cfg.SubgraphRef
	case ir.SubgraphConfig:
		return "subgraph", cfg.Ref
	default:
		return "", ""
	}
}

// boundaryDiag builds the DIP146 hint for a zero-intent child boundary.
func boundaryDiag(n *ir.Node, ref string) validator.Diagnostic {
	kind, _ := boundaryKindRef(n)
	return validator.Diagnostic{
		Code:     validator.DIP146,
		Severity: validator.SeverityHint,
		Message: fmt.Sprintf(
			"%s %q delegates to subgraph %q, which declares no tool_access restriction on any agent; a workflow on this path restricts tools, but the restriction does not cross the subgraph boundary",
			kind, n.ID, ref),
		Location: n.Source,
		Help:     "give the referenced .dip's agents their own tool_access (e.g. tool_access: none); this bounds the child's tool catalog, not information flow across the supervisory boundary (see #56). Multiple boundaries referencing the same child each get a hint; one tool_access edit clears them all.",
	}
}

// applyCrossFileToolAccess runs the native cross-file pass, drops the per-file
// DIP143 advisory for boundaries it conclusively resolved (zero-intent,
// full-restrict, or agent-less), and appends the DIP146 findings. Skipped for
// .dipx bundles, whose child refs are in-bundle paths, not on disk. The map
// lookup yields postureUnresolved (the zero value) for any DIP143 not classified
// here, which supersedes() treats as "retain".
func applyCrossFileToolAccess(diags []validator.Diagnostic, w *ir.Workflow, path string) []validator.Diagnostic {
	if strings.HasSuffix(strings.ToLower(path), ".dipx") {
		return diags
	}
	cross, classified := crossFileToolAccess(w, path)
	var kept []validator.Diagnostic
	for _, d := range diags {
		if d.Code == validator.DIP143 && supersedes(classified[d.Location]) {
			continue
		}
		kept = append(kept, d)
	}
	return append(kept, cross...)
}

// supersedes reports whether a resolved child posture makes the per-file DIP143
// advisory redundant: zero-intent is replaced by DIP146; full-restrict and
// agent-less are confirmed safe. Partial-audit and unresolved retain DIP143.
func supersedes(p childPosture) bool {
	return p == postureZeroIntent || p == postureFullRestrict || p == postureAgentless
}
