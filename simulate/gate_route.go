package simulate

import (
	"strings"

	"github.com/2389-research/dippin-lang/ir"
)

// Human-gate answer routing, mirroring tracker (github.com/2389-research/tracker):
//
//   - An answer routes by an edge's choice: key first, then its label:
//     (pipeline/engine_edges.go edgeRoutingKey prefers Choice over Label;
//     handlers/human.go mapSelectionToRoutingKey maps a picked label to its
//     edge's choice key).
//   - The comparison is an exact, trimmed, case-insensitive match
//     (handlers/human.go matchEdgeByLabelOrChoice, interviewer_console.go
//     matchConsoleChoice's strings.EqualFold). There is no substring or
//     accelerator-key matching: "n" does not pick "[N] No" or "Continue".
//   - default: answers an unattended choice gate (mode choice or unset) and a
//     freeform gate (handlers/human.go executeChoice / executeFreeform with
//     AutoApproveInterviewer). yes_no and interview gates ignore it
//     (executeYesNo passes no default; interview has no choices).

// gateDefault is the answer an unattended run gives human gate hc: its
// default: for the modes tracker honors it in, else "".
func gateDefault(hc ir.HumanConfig) string {
	if hc.Mode == "yes_no" || hc.Mode == "interview" {
		return ""
	}
	return hc.Default
}

// matchGateAnswer returns the edge answer routes the gate down: the first
// edge whose choice: key equals it, else the first whose label: does
// (trimmed, case-insensitive). Returns nil for an empty or unmatched answer.
func matchGateAnswer(edges []*ir.Edge, answer string) *ir.Edge {
	if strings.TrimSpace(answer) == "" {
		return nil
	}
	if e := findEdgeByKey(edges, answer, edgeChoice); e != nil {
		return e
	}
	return findEdgeByKey(edges, answer, edgeLabel)
}

func edgeChoice(e *ir.Edge) string { return e.Choice }
func edgeLabel(e *ir.Edge) string  { return e.Label }

// findEdgeByKey returns the first edge whose key (as read by keyOf) matches
// answer exactly, ignoring case and surrounding space.
func findEdgeByKey(edges []*ir.Edge, answer string, keyOf func(*ir.Edge) string) *ir.Edge {
	want := strings.TrimSpace(answer)
	for _, e := range edges {
		if k := keyOf(e); k != "" && strings.EqualFold(strings.TrimSpace(k), want) {
			return e
		}
	}
	return nil
}

// findEdgeByLabelSubstring is the legacy scenario convenience: a scenario's
// preferred_label also matches an edge whose label contains it
// (case-insensitive), so "yes" picks "[Y] Yes". It applies only to scenario
// values, never to a gate's default: or a typed answer, which route by
// matchGateAnswer alone.
func findEdgeByLabelSubstring(edges []*ir.Edge, target string) *ir.Edge {
	lower := strings.ToLower(target)
	for _, e := range edges {
		if strings.Contains(strings.ToLower(e.Label), lower) {
			return e
		}
	}
	return nil
}
