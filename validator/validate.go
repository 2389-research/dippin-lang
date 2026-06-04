package validator

import (
	"fmt"
	"sort"
	"strings"

	"github.com/2389-research/dippin-lang/ir"
)

// Validate runs all graph-structure checks (DIP001–DIP009) on the workflow
// and returns all diagnostics found. It always runs all checks — never
// short-circuits — so that a single pass reports everything.
func Validate(w *ir.Workflow) Result {
	var diags []Diagnostic

	// Run all checks unconditionally. DIP008 first because duplicate detection
	// is logically prior to other graph checks.
	diags = append(diags, checkNoDuplicateNodes(w)...)
	diags = append(diags, checkStartExists(w)...)
	diags = append(diags, checkExitExists(w)...)
	diags = append(diags, checkEdgeEndpoints(w)...)
	diags = append(diags, checkOnFailureTarget(w)...)
	diags = append(diags, checkExitNoOutgoing(w)...)
	diags = append(diags, checkNoDuplicateEdges(w)...)
	diags = append(diags, checkReachability(w)...)
	diags = append(diags, checkNoCycles(w)...)
	diags = append(diags, checkParallelFanIn(w)...)

	return Result{Diagnostics: diags}
}

// checkStartExists verifies DIP001: the start node ID is set and references
// an existing node.
func checkStartExists(w *ir.Workflow) []Diagnostic {
	if w.Start == "" {
		return []Diagnostic{{
			Code:     DIP001,
			Severity: SeverityError,
			Message:  "workflow has no start node declared",
			Help:     "add a start: field to the workflow",
		}}
	}
	if w.Node(w.Start) == nil {
		return []Diagnostic{{
			Code:     DIP001,
			Severity: SeverityError,
			Message:  fmt.Sprintf("start node %q is declared but does not exist in the node list", w.Start),
			Help:     fmt.Sprintf("add a node with ID %q to the workflow", w.Start),
		}}
	}
	return nil
}

// checkExitExists verifies DIP002: the exit node ID is set and references
// an existing node.
func checkExitExists(w *ir.Workflow) []Diagnostic {
	if w.Exit == "" {
		return []Diagnostic{{
			Code:     DIP002,
			Severity: SeverityError,
			Message:  "workflow has no exit node declared",
			Help:     "add an exit: field to the workflow",
		}}
	}
	if w.Node(w.Exit) == nil {
		return []Diagnostic{{
			Code:     DIP002,
			Severity: SeverityError,
			Message:  fmt.Sprintf("exit node %q is declared but does not exist in the node list", w.Exit),
			Help:     fmt.Sprintf("add a node with ID %q to the workflow", w.Exit),
		}}
	}
	return nil
}

// checkEdgeEndpoints verifies DIP003: every edge endpoint references an existing node.
// If a dangling reference is close to an existing node ID (Levenshtein ≤ 2),
// a "did you mean?" suggestion is included.
func checkEdgeEndpoints(w *ir.Workflow) []Diagnostic {
	var diags []Diagnostic
	nodeSet := buildNodeSet(w)

	for _, e := range w.Edges {
		diags = append(diags, checkEndpoint(w, nodeSet, e, e.From, "source")...)
		diags = append(diags, checkEndpoint(w, nodeSet, e, e.To, "target")...)
	}
	return diags
}

// checkEndpoint verifies a single edge endpoint exists, producing a diagnostic if not.
func checkEndpoint(w *ir.Workflow, nodeSet map[string]bool, e *ir.Edge, endpoint, role string) []Diagnostic {
	if endpoint == "" || nodeSet[endpoint] {
		return nil
	}
	d := Diagnostic{
		Code:     DIP003,
		Severity: SeverityError,
		Message:  fmt.Sprintf("edge references unknown %s node %q", role, endpoint),
		Location: e.Source,
	}
	if suggestion := closestNodeID(w, endpoint); suggestion != "" {
		d.Help = fmt.Sprintf("did you mean %q?", suggestion)
	} else {
		d.Help = fmt.Sprintf("declare a node with ID %q or fix the edge %s", endpoint, role)
	}
	return []Diagnostic{d}
}

// checkOnFailureTarget verifies DIP003: the graph-level on_failure route, if set,
// references an existing node. on_failure is a node reference like an edge endpoint,
// so it reuses the DIP003 code.
func checkOnFailureTarget(w *ir.Workflow) []Diagnostic {
	target := w.Defaults.OnFailure
	if target == "" || w.Node(target) != nil {
		return nil
	}
	d := Diagnostic{
		Code:     DIP003,
		Severity: SeverityError,
		Message:  fmt.Sprintf("on_failure references unknown node %q", target),
	}
	if suggestion := closestNodeID(w, target); suggestion != "" {
		d.Help = fmt.Sprintf("did you mean %q?", suggestion)
	} else {
		d.Help = fmt.Sprintf("declare a node with ID %q or fix the on_failure target", target)
	}
	return []Diagnostic{d}
}

// buildAllEdgeAdjacency builds an adjacency map including all edges (including restart).
// It also includes implicit parallel/fan_in edges.
func buildAllEdgeAdjacency(w *ir.Workflow) map[string][]string {
	adj := make(map[string][]string)
	for _, e := range w.Edges {
		adj[e.From] = append(adj[e.From], e.To)
	}
	addParallelFanInEdges(adj, w)
	return adj
}

// buildNonRestartAdjacency builds an adjacency map excluding restart edges.
// Does NOT include implicit parallel/fan_in edges (used for cycle detection).
func buildNonRestartAdjacency(w *ir.Workflow) map[string][]string {
	adj := make(map[string][]string)
	for _, e := range w.Edges {
		if !e.Restart {
			adj[e.From] = append(adj[e.From], e.To)
		}
	}
	return adj
}

// addParallelFanInEdges adds implicit edges from parallel targets and fan_in sources.
func addParallelFanInEdges(adj map[string][]string, w *ir.Workflow) {
	for _, n := range w.Nodes {
		switch cfg := n.Config.(type) {
		case ir.ParallelConfig:
			adj[n.ID] = append(adj[n.ID], cfg.Targets...)
		case ir.FanInConfig:
			for _, src := range cfg.Sources {
				adj[src] = append(adj[src], n.ID)
			}
		}
	}
}

// bfsReachable returns the set of nodes reachable from start via BFS.
func bfsReachable(start string, adj map[string][]string) map[string]bool {
	visited := make(map[string]bool)
	queue := []string{start}
	visited[start] = true

	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]
		for _, next := range adj[curr] {
			if !visited[next] {
				visited[next] = true
				queue = append(queue, next)
			}
		}
	}
	return visited
}

// checkReachability verifies DIP004: all nodes are reachable from the start node.
// Uses BFS traversal including restart edges (restart edges are valid paths for
// reachability purposes).
func checkReachability(w *ir.Workflow) []Diagnostic {
	// If start doesn't exist, we can't do reachability.
	if w.Start == "" || w.Node(w.Start) == nil {
		return nil
	}

	adj := buildAllEdgeAdjacency(w)
	visited := bfsReachable(w.Start, adj)
	return findUnreachableNodes(w, visited)
}

// findUnreachableNodes returns diagnostics for nodes not in the visited set.
func findUnreachableNodes(w *ir.Workflow, visited map[string]bool) []Diagnostic {
	var diags []Diagnostic
	for _, n := range w.Nodes {
		if !visited[n.ID] {
			diags = append(diags, Diagnostic{
				Code:     DIP004,
				Severity: SeverityError,
				Message:  fmt.Sprintf("node %q is unreachable from start node %q", n.ID, w.Start),
				Location: n.Source,
				Help:     fmt.Sprintf("add an edge leading to %q or remove it", n.ID),
			})
		}
	}
	return diags
}

// checkNoCycles verifies DIP005: no unconditional cycles exist after excluding
// edges marked restart: true. Per ADR 1, restart edges are back-edges that
// trigger downstream clear + re-execution and are not considered cycles.
func checkNoCycles(w *ir.Workflow) []Diagnostic {
	if w.Start == "" || w.Node(w.Start) == nil {
		return nil
	}

	adj := buildNonRestartAdjacency(w)
	return detectCyclesDFS(w, adj)
}

// cycleDetector holds state for DFS-based cycle detection.
type cycleDetector struct {
	adj      map[string][]string
	color    map[string]int // 0=white, 1=gray, 2=black
	parent   map[string]string
	reported map[string]bool
	diags    []Diagnostic
}

// detectCyclesDFS uses DFS with white/gray/black coloring to find cycles.
func detectCyclesDFS(w *ir.Workflow, adj map[string][]string) []Diagnostic {
	cd := &cycleDetector{
		adj:      adj,
		color:    make(map[string]int),
		parent:   make(map[string]string),
		reported: make(map[string]bool),
	}
	for _, n := range w.Nodes {
		if cd.color[n.ID] == 0 {
			cd.visit(n.ID)
		}
	}
	return cd.diags
}

// visit performs DFS from node, recording cycles found.
func (cd *cycleDetector) visit(node string) {
	cd.color[node] = 1 // gray
	for _, next := range cd.adj[node] {
		cd.processNeighbor(node, next)
	}
	cd.color[node] = 2 // black
}

// processNeighbor handles a single neighbor during DFS traversal.
func (cd *cycleDetector) processNeighbor(node, next string) {
	if cd.color[next] == 1 && !cd.reported[next] {
		cd.reported[next] = true
		cycle := reconstructCycle(cd.parent, node, next)
		cd.diags = append(cd.diags, Diagnostic{
			Code:     DIP005,
			Severity: SeverityError,
			Message:  fmt.Sprintf("unconditional cycle detected: %s", strings.Join(cycle, " → ")),
			Help:     "break the cycle by removing an edge or marking it restart: true",
		})
	}
	if cd.color[next] == 0 {
		cd.parent[next] = node
		cd.visit(next)
	}
}

// reconstructCycle builds the cycle path from the DFS parent map.
// Given that we found an edge from → to where to is already gray (in the
// current path), we walk parent pointers from "from" back to "to" to
// reconstruct: to → ... → from → to.
//
// The loop is bounded by the total number of nodes to prevent infinite
// iteration when the parent chain is incomplete (e.g. due to implicit
// edges added for parallel/fan-in nodes that were never DFS-visited).
func reconstructCycle(parent map[string]string, from, to string) []string {
	path := []string{to}
	curr := from
	seen := make(map[string]bool)
	seen[to] = true
	maxSteps := len(parent) + 2
	for canWalk(curr, to, seen, maxSteps) {
		seen[curr] = true
		path = append(path, curr)
		curr = parent[curr]
		maxSteps--
	}
	path = append(path, to)
	reversePath(path)
	return path
}

// canWalk checks whether the cycle reconstruction loop should continue.
func canWalk(curr, target string, seen map[string]bool, maxSteps int) bool {
	return curr != target && curr != "" && !seen[curr] && maxSteps > 0
}

// reversePath reverses a string slice in place.
func reversePath(path []string) {
	for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
		path[i], path[j] = path[j], path[i]
	}
}

// checkExitNoOutgoing verifies DIP006: the exit node has no outgoing edges.
func checkExitNoOutgoing(w *ir.Workflow) []Diagnostic {
	if w.Exit == "" || w.Node(w.Exit) == nil {
		return nil
	}

	outgoing := w.EdgesFrom(w.Exit)
	if len(outgoing) == 0 {
		return nil
	}

	var diags []Diagnostic
	for _, e := range outgoing {
		diags = append(diags, Diagnostic{
			Code:     DIP006,
			Severity: SeverityError,
			Message:  fmt.Sprintf("exit node %q has outgoing edge to %q", w.Exit, e.To),
			Location: e.Source,
			Help:     "remove the outgoing edge or change the exit node",
		})
	}
	return diags
}

// checkParallelFanIn verifies DIP007: every parallel fan-out node has a matching
// fan-in node, and vice versa. Matching means the ParallelConfig.Targets set
// equals the FanInConfig.Sources set (order-insensitive).
func checkParallelFanIn(w *ir.Workflow) []Diagnostic {
	parallels, fanIns := collectParallelFanIn(w)
	var diags []Diagnostic
	diags = append(diags, checkUnmatchedParallels(parallels, fanIns)...)
	diags = append(diags, checkUnmatchedFanIns(fanIns, parallels)...)
	return diags
}

// nodeTargets pairs a node with its sorted target/source list for matching.
type nodeTargets struct {
	node   *ir.Node
	sorted []string
}

// collectParallelFanIn collects and sorts targets/sources from parallel and fan_in nodes.
func collectParallelFanIn(w *ir.Workflow) (parallels, fanIns []nodeTargets) {
	for _, n := range w.Nodes {
		switch cfg := n.Config.(type) {
		case ir.ParallelConfig:
			sorted := make([]string, len(cfg.Targets))
			copy(sorted, cfg.Targets)
			sort.Strings(sorted)
			parallels = append(parallels, nodeTargets{node: n, sorted: sorted})
		case ir.FanInConfig:
			sorted := make([]string, len(cfg.Sources))
			copy(sorted, cfg.Sources)
			sort.Strings(sorted)
			fanIns = append(fanIns, nodeTargets{node: n, sorted: sorted})
		}
	}
	return
}

// hasMatch checks if any entry in candidates has sorted list equal to target.
func hasMatch(target []string, candidates []nodeTargets) bool {
	for _, c := range candidates {
		if slicesEqual(target, c.sorted) {
			return true
		}
	}
	return false
}

// checkUnmatchedParallels returns diagnostics for parallel nodes without matching fan_in.
func checkUnmatchedParallels(parallels, fanIns []nodeTargets) []Diagnostic {
	var diags []Diagnostic
	for _, p := range parallels {
		if !hasMatch(p.sorted, fanIns) {
			diags = append(diags, Diagnostic{
				Code:     DIP007,
				Severity: SeverityError,
				Message:  fmt.Sprintf("parallel node %q has targets %v but no matching fan_in node", p.node.ID, p.sorted),
				Location: p.node.Source,
				Help:     fmt.Sprintf("add a fan_in node with sources: %v", p.sorted),
			})
		}
	}
	return diags
}

// checkUnmatchedFanIns returns diagnostics for fan_in nodes without matching parallel.
func checkUnmatchedFanIns(fanIns, parallels []nodeTargets) []Diagnostic {
	var diags []Diagnostic
	for _, f := range fanIns {
		if !hasMatch(f.sorted, parallels) {
			diags = append(diags, Diagnostic{
				Code:     DIP007,
				Severity: SeverityError,
				Message:  fmt.Sprintf("fan_in node %q has sources %v but no matching parallel node", f.node.ID, f.sorted),
				Location: f.node.Source,
				Help:     fmt.Sprintf("add a parallel node with targets: %v", f.sorted),
			})
		}
	}
	return diags
}

// checkNoDuplicateNodes verifies DIP008: no two nodes share the same ID.
func checkNoDuplicateNodes(w *ir.Workflow) []Diagnostic {
	var diags []Diagnostic
	seen := make(map[string]*ir.Node)

	for _, n := range w.Nodes {
		if first, ok := seen[n.ID]; ok {
			diags = append(diags, Diagnostic{
				Code:     DIP008,
				Severity: SeverityError,
				Message:  fmt.Sprintf("duplicate node ID %q", n.ID),
				Location: n.Source,
				Help:     fmt.Sprintf("first declaration at %s:%d:%d", locFile(first.Source), first.Source.Line, first.Source.Column),
			})
		} else {
			seen[n.ID] = n
		}
	}
	return diags
}

// checkNoDuplicateEdges verifies DIP009: no two edges are identical.
// The dedup key is (From, To, Condition.Raw). Two edges with the same
// endpoints but different conditions are conditional branches, not duplicates.
func checkNoDuplicateEdges(w *ir.Workflow) []Diagnostic {
	type edgeKey struct {
		from, to, condRaw string
	}

	var diags []Diagnostic
	seen := make(map[edgeKey]*ir.Edge)

	for _, e := range w.Edges {
		condRaw := ""
		if e.Condition != nil {
			condRaw = e.Condition.Raw
		}
		key := edgeKey{from: e.From, to: e.To, condRaw: condRaw}
		if first, ok := seen[key]; ok {
			diags = append(diags, Diagnostic{
				Code:     DIP009,
				Severity: SeverityError,
				Message:  fmt.Sprintf("duplicate edge from %q to %q", e.From, e.To),
				Location: e.Source,
				Help:     fmt.Sprintf("first declaration at %s:%d:%d", locFile(first.Source), first.Source.Line, first.Source.Column),
			})
		} else {
			seen[key] = e
		}
	}
	return diags
}

// --- Helpers ---

// buildNodeSet returns a set of all node IDs in the workflow.
func buildNodeSet(w *ir.Workflow) map[string]bool {
	set := make(map[string]bool, len(w.Nodes))
	for _, n := range w.Nodes {
		set[n.ID] = true
	}
	return set
}

// closestNodeID returns the node ID most similar to the given name,
// or "" if no node is within edit distance 2.
func closestNodeID(w *ir.Workflow, name string) string {
	bestDist := 3 // threshold: Levenshtein ≤ 2
	bestID := ""
	for _, n := range w.Nodes {
		d := levenshtein(name, n.ID)
		if d < bestDist {
			bestDist = d
			bestID = n.ID
		}
	}
	return bestID
}

// levenshtein computes the Levenshtein edit distance between two strings.
func levenshtein(a, b string) int {
	la, lb := len(a), len(b)
	if la == 0 {
		return lb
	}
	if lb == 0 {
		return la
	}

	prev := make([]int, lb+1)
	curr := make([]int, lb+1)

	for j := 0; j <= lb; j++ {
		prev[j] = j
	}

	for i := 1; i <= la; i++ {
		curr[0] = i
		fillLevenshteinRow(a, b, i, prev, curr)
		prev, curr = curr, prev
	}
	return prev[lb]
}

// fillLevenshteinRow computes one row of the Levenshtein distance matrix.
func fillLevenshteinRow(a, b string, i int, prev, curr []int) {
	for j := 1; j <= len(b); j++ {
		cost := 1
		if a[i-1] == b[j-1] {
			cost = 0
		}
		curr[j] = min3(curr[j-1]+1, prev[j]+1, prev[j-1]+cost)
	}
}

// min3 returns the minimum of three integers.
func min3(a, b, c int) int {
	m := a
	if b < m {
		m = b
	}
	if c < m {
		m = c
	}
	return m
}

// slicesEqual returns true if two sorted string slices are element-wise equal.
func slicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// locFile returns the file from a SourceLocation, defaulting to "<unknown>".
func locFile(loc ir.SourceLocation) string {
	if loc.File == "" {
		return "<unknown>"
	}
	return loc.File
}
