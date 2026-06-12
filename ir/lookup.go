package ir

// Node returns the node with the given ID, or nil if not found.
func (w *Workflow) Node(id string) *Node {
	for _, n := range w.Nodes {
		if n.ID == id {
			return n
		}
	}
	return nil
}

// OutcomeChannel returns the node's natural outcome channel — the context
// variable the `on <token>` edge shorthand routes against — and whether the
// node has one. Agent nodes route on ctx.outcome; tool nodes that declare
// marker_grep route on ctx.tool_marker. Any other node has none.
//
// Human gates are deliberately excluded: they route on the human's choice
// (preferred_label / edge labels), and nothing populates ctx.outcome for them,
// so `on` would desugar to a condition that never matches. Human-gate routing
// keys are the domain of the `choice:` work (#130); use `when` meanwhile.
//
// This is the single source of truth shared by the parser (which desugars `on`)
// and the formatter (which re-emits it), so the two never diverge.
func (n *Node) OutcomeChannel() (string, bool) {
	if n == nil {
		return "", false
	}
	switch n.Kind {
	case NodeAgent:
		return "ctx.outcome", true
	case NodeTool:
		return n.toolMarkerChannel()
	}
	return "", false
}

// toolMarkerChannel returns ctx.tool_marker only when the tool declares marker_grep.
func (n *Node) toolMarkerChannel() (string, bool) {
	if cfg, ok := n.Config.(ToolConfig); ok && cfg.MarkerGrep != "" {
		return "ctx.tool_marker", true
	}
	return "", false
}

// IsOutcomeToken reports whether s is a valid `on <token>` value: a single bare
// identifier matching the shorthand grammar `[a-zA-Z0-9][a-zA-Z0-9_-]*`. This is
// the single source of truth shared by the parser (which rejects non-conforming
// `on` values) and the formatter (which only rewrites `when` to `on` for tokens
// that satisfy it), guaranteeing the two agree on what re-parses as a shorthand.
func IsOutcomeToken(s string) bool {
	for i, ch := range s {
		if !isOutcomeTokenChar(ch, i == 0) {
			return false
		}
	}
	return s != ""
}

// isOutcomeTokenChar reports whether ch is allowed at this position in an outcome
// token: an ASCII letter or digit anywhere, or `_`/`-` in any non-leading spot.
func isOutcomeTokenChar(ch rune, first bool) bool {
	if isASCIILetterOrDigit(ch) {
		return true
	}
	return !first && (ch == '_' || ch == '-')
}

func isASCIILetterOrDigit(ch rune) bool {
	return isASCIILetter(ch) || (ch >= '0' && ch <= '9')
}

func isASCIILetter(ch rune) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z')
}

// EdgesFrom returns all edges originating from the given node ID.
// This includes explicit edges from the workflow's Edges slice, as well as
// implicit edges defined by parallel fan-outs and fan-in joins.
// Redundant unconditional edges are de-duplicated.
func (w *Workflow) EdgesFrom(id string) []*Edge {
	var out []*Edge
	seenTargets := make(map[string]bool)

	out = w.explicitEdgesFrom(id, out, seenTargets)
	out = w.parallelEdgesFrom(id, out, seenTargets)
	out = w.fanInEdgesFrom(id, out, seenTargets)

	return out
}

// explicitEdgesFrom collects edges from the Edges slice where From == id.
func (w *Workflow) explicitEdgesFrom(id string, out []*Edge, seen map[string]bool) []*Edge {
	for _, e := range w.Edges {
		if e.From == id {
			out = append(out, e)
			if e.Condition == nil {
				seen[e.To] = true
			}
		}
	}
	return out
}

// parallelEdgesFrom adds implicit fan-out edges from a parallel node.
func (w *Workflow) parallelEdgesFrom(id string, out []*Edge, seen map[string]bool) []*Edge {
	node := w.Node(id)
	if node == nil {
		return out
	}
	cfg, ok := node.Config.(ParallelConfig)
	if !ok {
		return out
	}
	for _, target := range cfg.Targets {
		if !seen[target] {
			out = append(out, &Edge{From: id, To: target})
			seen[target] = true
		}
	}
	return out
}

// fanInEdgesFrom adds implicit edges from id to any fan-in node that lists id as a source.
func (w *Workflow) fanInEdgesFrom(id string, out []*Edge, seen map[string]bool) []*Edge {
	for _, n := range w.Nodes {
		cfg, ok := n.Config.(FanInConfig)
		if !ok {
			continue
		}
		out = w.addFanInSourceEdges(id, n.ID, cfg.Sources, out, seen)
	}
	return out
}

// addFanInSourceEdges adds an edge from srcID to fanInID if srcID appears in sources and is not already seen.
func (w *Workflow) addFanInSourceEdges(srcID, fanInID string, sources []string, out []*Edge, seen map[string]bool) []*Edge {
	for _, src := range sources {
		if src == srcID && !seen[fanInID] {
			out = append(out, &Edge{From: srcID, To: fanInID})
			seen[fanInID] = true
		}
	}
	return out
}

// EdgesTo returns all edges targeting the given node ID.
// This includes explicit edges from the workflow's Edges slice, as well as
// implicit edges defined by parallel fan-outs and fan-in joins.
// Redundant unconditional edges are de-duplicated.
func (w *Workflow) EdgesTo(id string) []*Edge {
	var out []*Edge
	seenSources := make(map[string]bool)

	out = w.explicitEdgesTo(id, out, seenSources)
	out = w.parallelEdgesTo(id, out, seenSources)
	out = w.fanInEdgesTo(id, out, seenSources)

	return out
}

// explicitEdgesTo collects edges from the Edges slice where To == id.
func (w *Workflow) explicitEdgesTo(id string, out []*Edge, seen map[string]bool) []*Edge {
	for _, e := range w.Edges {
		if e.To == id {
			out = append(out, e)
			if e.Condition == nil {
				seen[e.From] = true
			}
		}
	}
	return out
}

// parallelEdgesTo adds implicit edges from parallel nodes that target id.
func (w *Workflow) parallelEdgesTo(id string, out []*Edge, seen map[string]bool) []*Edge {
	for _, n := range w.Nodes {
		cfg, ok := n.Config.(ParallelConfig)
		if !ok {
			continue
		}
		if w.parallelTargetsNode(cfg.Targets, id) && !seen[n.ID] {
			out = append(out, &Edge{From: n.ID, To: id})
			seen[n.ID] = true
		}
	}
	return out
}

// parallelTargetsNode returns true if targets contains id.
func (w *Workflow) parallelTargetsNode(targets []string, id string) bool {
	for _, t := range targets {
		if t == id {
			return true
		}
	}
	return false
}

// fanInEdgesTo adds implicit edges from fan-in sources to id, if id is a fan-in node.
func (w *Workflow) fanInEdgesTo(id string, out []*Edge, seen map[string]bool) []*Edge {
	node := w.Node(id)
	if node == nil {
		return out
	}
	cfg, ok := node.Config.(FanInConfig)
	if !ok {
		return out
	}
	for _, src := range cfg.Sources {
		if !seen[src] {
			out = append(out, &Edge{From: src, To: id})
			seen[src] = true
		}
	}
	return out
}

// AllEdges returns every edge in the workflow, including implicit edges from
// parallel fan-outs and fan-in joins. This is the complete graph for consumers
// that need to traverse all connections (TUI, BFS, validation).
func (w *Workflow) AllEdges() []*Edge {
	seen := make(map[string]bool)
	var all []*Edge
	for _, n := range w.Nodes {
		for _, e := range w.EdgesFrom(n.ID) {
			key := e.From + "->" + e.To
			if !seen[key] {
				seen[key] = true
				all = append(all, e)
			}
		}
	}
	return all
}

// NodeIDs returns all node IDs in declaration order.
func (w *Workflow) NodeIDs() []string {
	ids := make([]string, len(w.Nodes))
	for i, n := range w.Nodes {
		ids[i] = n.ID
	}
	return ids
}
