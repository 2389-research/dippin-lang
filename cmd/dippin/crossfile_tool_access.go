package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/2389-research/dippin-lang/ir"
	"github.com/2389-research/dippin-lang/validator"
)

// crossFileMaxDepth bounds subgraph-ref recursion (mirrors dipx's
// maxManifestDepth). Deeper chains stop silently: the boundary stays
// unresolved, so its DIP143 advisory is retained.
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
	walkBoundaries(entry, intentSeen, 0, visited, &diags, classified)
	return diags, classified
}

// walkBoundaries inspects each subgraph/manager_loop boundary in w.
func walkBoundaries(w *ir.Workflow, intentSeen bool, depth int, visited map[string]bool, diags *[]validator.Diagnostic, classified map[ir.SourceLocation]childPosture) {
	for _, n := range w.Nodes {
		_, ref := boundaryKindRef(n)
		if ref == "" || n.Source.File == "" {
			continue
		}
		visitBoundary(n, ref, intentSeen, depth, visited, diags, classified)
	}
}

// visitBoundary resolves one boundary's child, records its posture, emits DIP146
// when the path shows intent and the child is zero-intent, then recurses.
func visitBoundary(n *ir.Node, ref string, intentSeen bool, depth int, visited map[string]bool, diags *[]validator.Diagnostic, classified map[ir.SourceLocation]childPosture) {
	child, childPath := resolveBoundaryChild(n, ref)
	if child == nil {
		classified[n.Source] = postureUnresolved
		return
	}
	posture := classifyChild(child)
	classified[n.Source] = posture
	if intentSeen && posture == postureZeroIntent {
		*diags = append(*diags, boundaryDiag(n, ref))
	}
	maybeRecurse(child, childPath, intentSeen, depth, visited, diags, classified)
}

// maybeRecurse descends into a child's own boundaries unless it was already
// visited (cycle guard) or the depth cap is reached. The child is marked visited
// before recursing (pre-order) so cycles terminate.
func maybeRecurse(child *ir.Workflow, childPath string, intentSeen bool, depth int, visited map[string]bool, diags *[]validator.Diagnostic, classified map[ir.SourceLocation]childPosture) {
	key := canonicalKey(childPath)
	if key == "" || visited[key] || depth+1 >= crossFileMaxDepth {
		return
	}
	visited[key] = true
	childIntent := intentSeen || validator.WorkflowDeclaresToolAccess(child)
	walkBoundaries(child, childIntent, depth+1, visited, diags, classified)
}

// resolveBoundaryChild resolves ref relative to the boundary node's source file
// and parses the child workflow. Fail-soft: any error yields (nil, "").
// Callers (walkBoundaries) guarantee ref != "" and n.Source.File != "".
func resolveBoundaryChild(n *ir.Node, ref string) (*ir.Workflow, string) {
	path := resolveBoundaryRefPath(ref, n.Source.File)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, ""
	}
	w, err := parseAndResolveDip(data, path)
	if err != nil {
		return nil, ""
	}
	return w, path
}

// resolveBoundaryRefPath resolves a (possibly relative) ref against the parent's
// source file directory, mirroring DIP135's resolveRefPath.
func resolveBoundaryRefPath(ref, sourceFile string) string {
	if filepath.IsAbs(ref) {
		return ref
	}
	return filepath.Join(filepath.Dir(sourceFile), ref)
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
// file reached via different paths (or a symlink cycle) maps to one key.
func canonicalKey(path string) string {
	if path == "" {
		return ""
	}
	if resolved, err := filepath.EvalSymlinks(path); err == nil {
		return resolved
	}
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
