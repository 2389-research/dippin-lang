// ABOUTME: Flattens subgraph ref nodes by inlining referenced workflows.
// ABOUTME: Produces a single flat ir.Workflow with all refs resolved.
package flatten

import (
	"fmt"
	"strings"

	"github.com/2389-research/dippin-lang/ir"
)

// Resolver loads a referenced workflow by path.
type Resolver interface {
	Resolve(refPath string, relativeTo string) (*ir.Workflow, error)
}

// Options controls flattening behavior.
type Options struct {
	MaxDepth int // default 10; catches circular refs
}

const defaultMaxDepth = 10

// Flatten resolves all subgraph ref nodes in w by inlining the referenced
// workflows. If w contains no subgraph refs, it returns a shallow copy.
func Flatten(w *ir.Workflow, resolve Resolver, opts Options) (*ir.Workflow, error) {
	if !hasSubgraphRefs(w) {
		return copyWorkflow(w), nil
	}
	if resolve == nil {
		return nil, fmt.Errorf("flatten: resolver is nil but workflow has subgraph refs")
	}
	maxDepth := opts.MaxDepth
	if maxDepth == 0 {
		maxDepth = defaultMaxDepth
	}
	return flattenRecursive(w, resolve, maxDepth, 0, nil)
}

// hasSubgraphRefs returns true if any node in w is a subgraph with a Ref.
func hasSubgraphRefs(w *ir.Workflow) bool {
	for _, n := range w.Nodes {
		cfg, ok := n.Config.(ir.SubgraphConfig)
		if ok && cfg.Ref != "" {
			return true
		}
	}
	return false
}

// copyWorkflow returns a shallow copy of w with fresh node and edge slices.
func copyWorkflow(w *ir.Workflow) *ir.Workflow {
	out := *w
	out.Nodes = make([]*ir.Node, len(w.Nodes))
	copy(out.Nodes, w.Nodes)
	out.Edges = make([]*ir.Edge, len(w.Edges))
	copy(out.Edges, w.Edges)
	return &out
}

// rewire tracks how a subgraph node's edges get remapped to inlined nodes.
type rewire struct {
	start string // prefixed start node of the inlined child
	exit  string // prefixed exit node of the inlined child
}

// flattenRecursive walks w's nodes, inlines any subgraph refs, and rewires edges.
func flattenRecursive(w *ir.Workflow, resolve Resolver, maxDepth, depth int, seen []string) (*ir.Workflow, error) {
	if depth >= maxDepth {
		return nil, fmt.Errorf("flatten: max depth %d exceeded", maxDepth)
	}

	out := newOutputWorkflow(w)
	rewires := make(map[string]rewire)

	for _, n := range w.Nodes {
		r, err := processNode(n, resolve, maxDepth, depth, seen, out)
		if err != nil {
			return nil, err
		}
		if r != nil {
			rewires[n.ID] = *r
		}
	}

	rewireParentEdges(w.Edges, rewires, out)
	rewriteStartExit(out, rewires)
	return out, nil
}

// rewriteStartExit updates Start/Exit if they pointed to inlined subgraph nodes.
func rewriteStartExit(out *ir.Workflow, rewires map[string]rewire) {
	if r, ok := rewires[out.Start]; ok {
		out.Start = r.start
	}
	if r, ok := rewires[out.Exit]; ok {
		out.Exit = r.exit
	}
}

// newOutputWorkflow creates a fresh Workflow copying the metadata fields listed
// below but no nodes/edges. It does NOT carry w.Requires, w.ElseTarget, or
// w.ElseTargetSource, so flattening a workflow with subgraph refs drops the
// requires list and the section-level `else ->` funnel default (a known gap —
// see issue #161).
func newOutputWorkflow(w *ir.Workflow) *ir.Workflow {
	return &ir.Workflow{
		Name:       w.Name,
		Version:    w.Version,
		Goal:       w.Goal,
		Start:      w.Start,
		Exit:       w.Exit,
		Defaults:   w.Defaults,
		Vars:       w.Vars,
		Stylesheet: w.Stylesheet,
		SourceMap:  w.SourceMap,
	}
}

// processNode handles a single node: either copies it or inlines its subgraph ref.
func processNode(n *ir.Node, resolve Resolver, maxDepth, depth int, seen []string, out *ir.Workflow) (*rewire, error) {
	cfg, ok := n.Config.(ir.SubgraphConfig)
	if !ok || cfg.Ref == "" {
		out.Nodes = append(out.Nodes, n)
		return nil, nil
	}
	return inlineSubgraph(n, cfg, resolve, maxDepth, depth, seen, out)
}

// inlineSubgraph resolves a subgraph ref and splices the child's nodes/edges into out.
func inlineSubgraph(n *ir.Node, cfg ir.SubgraphConfig, resolve Resolver, maxDepth, depth int, seen []string, out *ir.Workflow) (*rewire, error) {
	if containsStr(seen, cfg.Ref) {
		return nil, fmt.Errorf("flatten: cycle detected: %s", formatCycle(seen, cfg.Ref))
	}

	child, err := resolveAndValidate(n, cfg, resolve)
	if err != nil {
		return nil, err
	}

	child, err = flattenRecursive(child, resolve, maxDepth, depth+1, append(seen, cfg.Ref))
	if err != nil {
		return nil, err
	}

	prefix := n.ID + "_"
	appendPrefixed(child, prefix, out)
	return &rewire{start: prefix + child.Start, exit: prefix + child.Exit}, nil
}

// resolveAndValidate loads a child workflow and checks it has start/exit.
func resolveAndValidate(n *ir.Node, cfg ir.SubgraphConfig, resolve Resolver) (*ir.Workflow, error) {
	child, err := resolve.Resolve(cfg.Ref, n.Source.File)
	if err != nil {
		return nil, fmt.Errorf("flatten: node %q: cannot resolve ref %q: %w", n.ID, cfg.Ref, err)
	}
	if child.Start == "" {
		return nil, fmt.Errorf("flatten: node %q: resolved workflow %q has no start node", n.ID, child.Name)
	}
	if child.Exit == "" {
		return nil, fmt.Errorf("flatten: node %q: resolved workflow %q has no exit node", n.ID, child.Name)
	}
	return child, nil
}

// appendPrefixed adds child nodes and edges to out with the given ID prefix.
func appendPrefixed(child *ir.Workflow, prefix string, out *ir.Workflow) {
	for _, cn := range child.Nodes {
		prefixed := *cn
		prefixed.ID = prefix + cn.ID
		out.Nodes = append(out.Nodes, &prefixed)
	}
	for _, ce := range child.Edges {
		prefixed := *ce
		prefixed.From = prefix + ce.From
		prefixed.To = prefix + ce.To
		out.Edges = append(out.Edges, &prefixed)
	}
}

// rewireParentEdges copies parent edges into out, rewriting endpoints that
// pointed to inlined subgraph nodes to point to their start/exit instead.
func rewireParentEdges(edges []*ir.Edge, rewires map[string]rewire, out *ir.Workflow) {
	for _, e := range edges {
		ne := *e
		if r, ok := rewires[ne.To]; ok {
			ne.To = r.start
		}
		if r, ok := rewires[ne.From]; ok {
			ne.From = r.exit
		}
		out.Edges = append(out.Edges, &ne)
	}
}

// containsStr returns true if s is in the slice ss.
func containsStr(ss []string, s string) bool {
	for _, v := range ss {
		if v == s {
			return true
		}
	}
	return false
}

// formatCycle formats the seen path plus the cyclic ref for error messages.
func formatCycle(seen []string, ref string) string {
	parts := make([]string, len(seen)+1)
	copy(parts, seen)
	parts[len(seen)] = ref
	return strings.Join(parts, " → ")
}
