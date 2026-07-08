// ABOUTME: DIP152 — a tool node's marker_grep enumerates a literal marker that
// ABOUTME: no edge routes and no else/unconditional edge covers.
package validator

import (
	"fmt"
	"sort"
	"strings"

	"github.com/2389-research/dippin-lang/ir"
)

// markerMetachars are the regex metacharacters that make a branch non-literal.
const markerMetachars = ".*+?[]{}()|\\^$"

// enumerateMarkers returns the literal marker set of a marker_grep value when it
// is a recognizable finite literal alternation (optional ^...$ anchors around a
// single literal token or a full-span (a|b|c) group of literal branches).
// Returns (nil, false) for anything else — those keep the blanket DIP101/DIP102
// exemption and get no DIP152.
func enumerateMarkers(markerGrep string) ([]string, bool) {
	branches, ok := splitAlternation(stripAnchors(markerGrep))
	if !ok {
		return nil, false
	}
	return collectUniqueLiterals(branches)
}

// collectUniqueLiterals converts a slice of regex branches into a deduplicated
// literal marker list, returning (nil, false) if any branch is empty or contains
// a metacharacter.
func collectUniqueLiterals(branches []string) ([]string, bool) {
	seen := make(map[string]struct{})
	var markers []string
	for _, b := range branches {
		if b == "" || !isLiteralToken(b) {
			return nil, false
		}
		if _, dup := seen[b]; !dup {
			seen[b] = struct{}{}
			markers = append(markers, b)
		}
	}
	return markers, true
}

// stripAnchors removes one leading ^ and one trailing $ if present.
func stripAnchors(s string) string {
	return strings.TrimSuffix(strings.TrimPrefix(s, "^"), "$")
}

// splitAlternation returns the branches of a full-span (a|b|c) group, or the
// whole string as a single branch when it is not a full-span group. Empty input
// is non-enumerable.
func splitAlternation(s string) ([]string, bool) {
	if s == "" {
		return nil, false
	}
	if strings.HasPrefix(s, "(") && strings.HasSuffix(s, ")") {
		return strings.Split(s[1:len(s)-1], "|"), true
	}
	return []string{s}, true
}

// isLiteralToken reports whether s contains no regex metacharacter.
func isLiteralToken(s string) bool {
	return !strings.ContainsAny(s, markerMetachars)
}

// lintMarkerCoverage checks DIP152 across all tool nodes.
func lintMarkerCoverage(w *ir.Workflow) []Diagnostic {
	var diags []Diagnostic
	elseValid := hasValidElseDefault(w)
	for _, n := range w.Nodes {
		if n.ID == w.Exit {
			continue // the exit node legitimately has no outgoing edges
		}
		if d, ok := checkMarkerCoverage(w, n, elseValid); ok {
			diags = append(diags, d)
		}
	}
	return diags
}

// checkMarkerCoverage returns a DIP152 diagnostic for one node if its enumerable
// marker_grep has markers neither routed nor covered.
func checkMarkerCoverage(w *ir.Workflow, n *ir.Node, elseValid bool) (Diagnostic, bool) {
	markers, ok := nodeEnumerableMarkers(n)
	if !ok {
		return Diagnostic{}, false
	}
	channel, _ := n.OutcomeChannel() // "ctx.tool_marker" for a marker tool
	edges := w.EdgesFrom(n.ID)
	routed, hasUncond, hasComplex := classifyMarkerEdges(edges, channel)
	if markerNodeCovered(len(edges) > 0, elseValid, hasUncond, hasComplex) {
		return Diagnostic{}, false
	}
	missing := uncoveredMarkers(markers, routed)
	if len(missing) == 0 {
		return Diagnostic{}, false
	}
	return markerCoverageDiag(n, missing), true
}

// nodeEnumerableMarkers returns the enumerable marker set for a tool node, or
// (nil,false) if the node is not a marker tool or its grep is non-enumerable.
func nodeEnumerableMarkers(n *ir.Node) ([]string, bool) {
	cfg, ok := n.Config.(ir.ToolConfig)
	if !ok || cfg.MarkerGrep == "" {
		return nil, false
	}
	return enumerateMarkers(cfg.MarkerGrep)
}

// classifyMarkerEdges splits a node's outgoing edges into the simple-equality
// routed marker set plus flags for an unconditional edge and any "complex" edge
// (compound/negated/other-variable) whose mere presence makes the node safe.
func classifyMarkerEdges(edges []*ir.Edge, channel string) (routed map[string]struct{}, hasUncond, hasComplex bool) {
	routed = make(map[string]struct{})
	for _, e := range edges {
		if e.Condition == nil {
			hasUncond = true
			continue
		}
		if cmp, ok := ir.ExtractEqualityCondition(e); ok && cmp.Variable == channel {
			routed[cmp.Value] = struct{}{}
			continue
		}
		hasComplex = true
	}
	return routed, hasUncond, hasComplex
}

// markerNodeCovered reports whether the node is safe regardless of the routed
// set. A node with no outgoing edges is never covered: `else` and unconditional
// fallbacks only route a node whose guard edges fail to match (the simulator
// dead-ends an edge-less node before any else routing, and DIP102 skips it too),
// so an edge-less marker tool strands its markers.
func markerNodeCovered(hasEdges, elseValid, hasUncond, hasComplex bool) bool {
	return hasEdges && (elseValid || hasUncond || hasComplex)
}

// uncoveredMarkers returns the sorted markers not in the routed set.
func uncoveredMarkers(markers []string, routed map[string]struct{}) []string {
	var missing []string
	for _, m := range markers {
		if _, ok := routed[m]; !ok {
			missing = append(missing, m)
		}
	}
	sort.Strings(missing)
	return missing
}

// markerCoverageDiag builds the DIP152 diagnostic for a node.
func markerCoverageDiag(n *ir.Node, missing []string) Diagnostic {
	return Diagnostic{
		Code:     DIP152,
		Severity: SeverityWarning,
		Message:  fmt.Sprintf("tool node %q emits markers that no edge routes and no else default covers: %s", n.ID, quoteMarkers(missing)),
		Location: n.Source,
		Help:     "route each marker with an edge (e.g. `on <marker>`), add an unconditional fallback edge, or add a section `else -> <node>` default",
	}
}

// quoteMarkers renders markers as a comma-separated list of quoted literals, so
// values containing whitespace stay legible in the diagnostic.
func quoteMarkers(markers []string) string {
	quoted := make([]string, len(markers))
	for i, m := range markers {
		quoted[i] = fmt.Sprintf("%q", m)
	}
	return strings.Join(quoted, ", ")
}
