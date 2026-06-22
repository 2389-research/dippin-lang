package validator

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/2389-research/dippin-lang/ir"
)

// varRefPattern matches ${...} variable references in prompt text.
var varRefPattern = regexp.MustCompile(`\$\{([^}]+)\}`)

// lintUndefinedVariables checks DIP106: ${variable} references in prompts
// must use a known namespace prefix (ctx., graph., params., stack.) or be a
// node-scoped ref (node.<id>.<key> or ctx.node.<id>.<key>). References without
// a recognized namespace or valid node ID are flagged.
func lintUndefinedVariables(w *ir.Workflow) []Diagnostic {
	var diags []Diagnostic
	for _, n := range w.Nodes {
		diags = append(diags, checkNodeVarRefs(n, w)...)
	}
	return diags
}

// checkNodeVarRefs checks variable references in a single node's prompt.
func checkNodeVarRefs(n *ir.Node, w *ir.Workflow) []Diagnostic {
	prompt := nodePrompt(n)
	if prompt == "" {
		return nil
	}
	var diags []Diagnostic
	matches := varRefPattern.FindAllStringSubmatch(prompt, -1)
	for _, m := range matches {
		varRef := m[1]
		if !isVarRefValid(varRef, w) {
			diags = append(diags, Diagnostic{
				Code:     DIP106,
				Severity: SeverityWarning,
				Message:  fmt.Sprintf("node %q has unrecognized variable reference ${%s}", n.ID, varRef),
				Location: n.Source,
				Help:     varRefHelp(varRef),
			})
		}
	}
	return diags
}

// varRefHelp returns an appropriate help message for an invalid variable reference.
// Node-scoped refs get a specific message; everything else gets the generic namespace hint.
func varRefHelp(varRef string) string {
	bare := strings.TrimPrefix(varRef, "ctx.")
	if strings.HasPrefix(bare, "node.") {
		parts := strings.SplitN(bare, ".", 3) // ["node", "<id>", ...]
		if len(parts) >= 2 && parts[1] != "" {
			return fmt.Sprintf("node %q does not exist in the workflow", parts[1])
		}
		return "node-scoped refs must be in the form node.<id>.<key>"
	}
	return fmt.Sprintf("use a namespaced variable like ${ctx.%s}, ${graph.%s}, or ${params.%s}", varRef, varRef, varRef)
}

// isVarRefValid returns true if the variable reference is valid: either a known
// namespace prefix or a node-scoped reference (node.<existingNode>.<key> or
// ctx.node.<existingNode>.<key>) with a real node ID and exactly 3 parts.
func isVarRefValid(varRef string, w *ir.Workflow) bool {
	bare := strings.TrimPrefix(varRef, "ctx.")
	if strings.HasPrefix(bare, "node.") {
		return isNodeScopedKey(bare, w)
	}
	parts := strings.SplitN(varRef, ".", 2)
	return len(parts) >= 2 && knownNamespaces[parts[0]]
}

// isNodeScopedKey returns true if the key matches node.<nodeID>.<key> where nodeID
// is a real node in the workflow. The key must have exactly 3 dot-separated parts.
// Callers must strip any "ctx." prefix before calling this function.
func isNodeScopedKey(key string, w *ir.Workflow) bool {
	if !strings.HasPrefix(key, "node.") {
		return false
	}
	parts := strings.SplitN(key, ".", 3) // ["node", "<nodeID>", "<key>"]
	if len(parts) < 3 {
		return false
	}
	return w.Node(parts[1]) != nil
}

// lintUnusedWrites checks DIP107: context keys declared in a node's writes:
// that are not referenced in any other node's reads:. These are dead outputs
// that may indicate unused work.
func lintUnusedWrites(w *ir.Workflow) []Diagnostic {
	allReads := collectAllReads(w)

	var diags []Diagnostic
	for _, n := range w.Nodes {
		for _, key := range n.IO.Writes {
			if !allReads[key] {
				diags = append(diags, Diagnostic{
					Code:     DIP107,
					Severity: SeverityWarning,
					Message:  fmt.Sprintf("node %q writes context key %q but no node declares it in reads", n.ID, key),
					Location: n.Source,
					Help:     fmt.Sprintf("add reads: %s to a downstream node, or remove this write if unused", key),
				})
			}
		}
	}
	return diags
}

// collectAllReads gathers all read keys across all nodes.
func collectAllReads(w *ir.Workflow) map[string]bool {
	allReads := make(map[string]bool)
	for _, n := range w.Nodes {
		for _, key := range n.IO.Reads {
			allReads[key] = true
		}
	}
	return allReads
}

// lintReadsWithoutUpstreamWrites checks DIP112: keys in a node's reads: list
// should appear in the writes: list of at least one upstream node (reachable
// via forward edges from start). This uses a simple flow analysis — for each
// node, compute the set of keys written by upstream nodes, then check reads.
func lintReadsWithoutUpstreamWrites(w *ir.Workflow) []Diagnostic {
	if w.Start == "" || w.Node(w.Start) == nil {
		return nil
	}

	adj := buildForwardAdjacency(w)
	available, upstream := computeAvailableAndUpstream(w, adj)
	return checkUnprovidedReads(w, available, upstream)
}

// computeAvailableAndUpstream performs a topological traversal and computes:
//   - available: context keys written by upstream nodes (available at each node)
//   - upstream: set of node IDs that are strictly upstream of each node
func computeAvailableAndUpstream(w *ir.Workflow, adj map[string][]string) (map[string]map[string]bool, map[string]map[string]bool) {
	inDegree := computeInDegrees(w, adj)
	queue := findRootNodes(w, inDegree)
	available := initializeAvailable(w)
	upstream := initializeAvailable(w)

	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]
		addNodeWrites(w, available, curr)
		queue = propagateAndEnqueue(adj, available, upstream, inDegree, curr, queue)
	}

	return available, upstream
}

// addNodeWrites adds a node's write keys to its available set.
func addNodeWrites(w *ir.Workflow, available map[string]map[string]bool, nodeID string) {
	if n := w.Node(nodeID); n != nil {
		for _, key := range n.IO.Writes {
			available[nodeID][key] = true
		}
	}
}

// propagateAndEnqueue propagates available keys and upstream node IDs to successors,
// then enqueues successors whose in-degree reaches zero.
func propagateAndEnqueue(adj map[string][]string, available, upstream map[string]map[string]bool, inDegree map[string]int, curr string, queue []string) []string {
	for _, next := range adj[curr] {
		if upstream[next] == nil {
			upstream[next] = make(map[string]bool)
		}
		propagateKeys(available, curr, next)
		propagateKeys(upstream, curr, next)
		upstream[next][curr] = true
		inDegree[next]--
		if inDegree[next] == 0 {
			queue = append(queue, next)
		}
	}
	return queue
}

// computeInDegrees calculates the in-degree for each node considering
// both explicit edges and implicit parallel/fan_in edges.
func computeInDegrees(w *ir.Workflow, adj map[string][]string) map[string]int {
	inDegree := make(map[string]int)
	for _, n := range w.Nodes {
		inDegree[n.ID] = 0
	}
	countExplicitInDegrees(w, inDegree)
	countImplicitInDegrees(w, inDegree)
	return inDegree
}

// countExplicitInDegrees counts in-degrees from non-restart edges.
func countExplicitInDegrees(w *ir.Workflow, inDegree map[string]int) {
	for _, e := range w.Edges {
		if !e.Restart {
			inDegree[e.To]++
		}
	}
}

// countImplicitInDegrees counts in-degrees from parallel/fan_in implicit edges.
func countImplicitInDegrees(w *ir.Workflow, inDegree map[string]int) {
	for _, n := range w.Nodes {
		switch cfg := n.Config.(type) {
		case ir.ParallelConfig:
			for _, t := range cfg.Targets {
				inDegree[t]++
			}
		case ir.FanInConfig:
			inDegree[n.ID] += len(cfg.Sources)
		}
	}
}

// findRootNodes returns all nodes with in-degree 0 (entry points for traversal).
func findRootNodes(w *ir.Workflow, inDegree map[string]int) []string {
	var roots []string
	for _, n := range w.Nodes {
		if inDegree[n.ID] == 0 {
			roots = append(roots, n.ID)
		}
	}
	return roots
}

// initializeAvailable creates the availability map with empty sets for each node.
func initializeAvailable(w *ir.Workflow) map[string]map[string]bool {
	available := make(map[string]map[string]bool)
	for _, n := range w.Nodes {
		available[n.ID] = make(map[string]bool)
	}
	return available
}

// propagateKeys copies all available keys from source to destination node.
func propagateKeys(available map[string]map[string]bool, from, to string) {
	for key := range available[from] {
		available[to][key] = true
	}
}

// checkUnprovidedReads generates diagnostics for reads that have no upstream write.
func checkUnprovidedReads(w *ir.Workflow, available, upstream map[string]map[string]bool) []Diagnostic {
	var diags []Diagnostic
	for _, n := range w.Nodes {
		for _, key := range n.IO.Reads {
			diags = append(diags, checkReadKey(n, key, w, available[n.ID], upstream[n.ID])...)
		}
	}
	return diags
}

// checkReadKey validates a single read key for a node, emitting DIP112 if needed.
func checkReadKey(n *ir.Node, key string, w *ir.Workflow, avail, upstream map[string]bool) []Diagnostic {
	if avail[key] {
		return nil
	}
	if msg := nodeReadDIP112(n.ID, key, w, upstream); msg != "" {
		return []Diagnostic{{
			Code:     DIP112,
			Severity: SeverityWarning,
			Message:  msg,
			Location: n.Source,
			Help:     fmt.Sprintf("add writes: %s to an upstream node, or the key may be auto-injected at runtime", key),
		}}
	}
	return nil
}

// nodeReadDIP112 returns a non-empty diagnostic message when a read key is
// problematic, or empty string if the read is valid.
func nodeReadDIP112(nodeID, key string, w *ir.Workflow, upstream map[string]bool) string {
	parts := strings.SplitN(strings.TrimPrefix(key, "ctx."), ".", 3)
	if len(parts) == 3 && parts[0] == "node" {
		return nodeScopedReadMsg(nodeID, key, parts[1], w, upstream)
	}
	if isNodeScopedKey(key, w) {
		return ""
	}
	return fmt.Sprintf("node %q reads context key %q but no upstream node declares it in writes", nodeID, key)
}

// nodeScopedReadMsg validates a node.<id>.* read and returns an error message or "".
func nodeScopedReadMsg(readerID, key, refNodeID string, w *ir.Workflow, upstream map[string]bool) string {
	if w.Node(refNodeID) == nil {
		return fmt.Sprintf("node %q reads %q but node %q does not exist in the workflow", readerID, key, refNodeID)
	}
	if !upstream[refNodeID] {
		return fmt.Sprintf("node %q reads %q but node %q is not upstream", readerID, key, refNodeID)
	}
	return ""
}

// nodePrompt extracts the prompt text from a node if it has one.
func nodePrompt(n *ir.Node) string {
	switch cfg := n.Config.(type) {
	case ir.AgentConfig:
		return cfg.Prompt
	default:
		return ""
	}
}
