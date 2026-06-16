package validator

import (
	"fmt"

	"github.com/2389-research/dippin-lang/ir"
)

// lintAmbiguousRouting checks DIP149: a node with two or more unconditional
// (no `when`) outgoing forward edges. Such edges are equally eligible, so which
// one fires is decided only by the cascade's lexical tiebreak on target node
// IDs — renaming a target can silently change routing. The author should keep
// at most one unconditional edge (the default fallback) and guard the rest.
//
// Conservative by design (favors false negatives):
//   - Restart back-edges are excluded; they are a distinct re-execution channel,
//     not a forward-routing competitor.
//   - A guarded edge plus a single unconditional fallback is the intended
//     pattern and is not flagged.
//   - Duplicate same-variable/same-value guards are already covered by DIP103.
//   - parallel source nodes are excluded: their multiple unconditional outgoing
//     edges are structural fan-out (synthesized by ir.EdgesFrom), all fired
//     together rather than chosen by a tiebreak. (fan_in nodes are NOT excluded:
//     ir.EdgesFrom synthesizes edges *into* a fan_in, never out of it, so a
//     fan_in's own outgoing edges route by the ordinary cascade.)
//   - human source nodes are excluded: a choice/yes_no gate routes on the
//     human's label selection, not the cascade tiebreak (see ir.OutcomeChannel's
//     human-gate note). This is conservative for non-choice modes (interview /
//     freeform), which do route by the cascade — a deliberate false-negative.
func lintAmbiguousRouting(w *ir.Workflow) []Diagnostic {
	var diags []Diagnostic
	for _, n := range w.Nodes {
		if isStructuralOrChoiceRouter(n.Kind) {
			continue
		}
		if d, ok := checkAmbiguousRouting(n, w.EdgesFrom(n.ID)); ok {
			diags = append(diags, d)
		}
	}
	return diags
}

// isStructuralOrChoiceRouter reports whether a source node's outgoing edges are
// resolved by something other than the cascade's unconditional-edge tiebreak:
// parallel fan-out (all edges fire) or a human choice gate (the human picks).
// DIP149 does not apply to these. fan_in is intentionally absent — its outgoing
// edges route by the ordinary cascade (ir.EdgesFrom only synthesizes fan-out
// edges for parallel nodes, never for fan_in).
func isStructuralOrChoiceRouter(k ir.NodeKind) bool {
	switch k {
	case ir.NodeParallel, ir.NodeHuman:
		return true
	}
	return false
}

// checkAmbiguousRouting flags a node whose outgoing forward edges include two
// or more unconditional ones.
func checkAmbiguousRouting(n *ir.Node, edges []*ir.Edge) (Diagnostic, bool) {
	if countUnconditionalForwardEdges(edges) < 2 {
		return Diagnostic{}, false
	}
	return Diagnostic{
		Code:     DIP149,
		Severity: SeverityWarning,
		Message:  fmt.Sprintf("node %q has multiple unconditional outgoing edges; which one fires is decided only by the lexical tiebreak", n.ID),
		Location: n.Source,
		Help:     "keep at most one unconditional edge as the default fallback and guard the others with a `when` condition",
	}, true
}

// lintUnusedWeight checks DIP151: an edge carrying a `weight:` attribute. The
// routing cascade does not consult weights, so the keyword is dead machinery —
// speculative tier-4 priority that no real workflow uses. weight: still parses
// (carry-only, no breakage in Phase 0), but the keyword and the cascade tier are
// slated for removal under `dip 2`. The lint steers authors toward conditions,
// `on`, or a single unconditional fallback before that removal lands.
func lintUnusedWeight(w *ir.Workflow) []Diagnostic {
	var diags []Diagnostic
	for _, e := range w.Edges {
		if e.Weight == 0 {
			continue
		}
		diags = append(diags, Diagnostic{
			Code:     DIP151,
			Severity: SeverityWarning,
			Message:  fmt.Sprintf("edge %q -> %q sets weight: %d, which routing does not use; guard edges with when / on or rely on a single unconditional fallback instead", e.From, e.To, e.Weight),
			Location: e.Source,
			Help:     "weight: is parsed but unused by routing and is slated for removal in dip 2; remove it or express priority with conditions",
		})
	}
	return diags
}

// countUnconditionalForwardEdges counts edges with no condition that are not
// restart back-edges.
func countUnconditionalForwardEdges(edges []*ir.Edge) int {
	n := 0
	for _, e := range edges {
		if e.Condition == nil && !e.Restart {
			n++
		}
	}
	return n
}
