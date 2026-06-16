package validator

import (
	"fmt"

	"github.com/2389-research/dippin-lang/ir"
)

var validHumanModes = map[string]bool{
	"choice":    true,
	"freeform":  true,
	"interview": true,
	"yes_no":    true,
}

func lintHumanMode(w *ir.Workflow) []Diagnostic {
	var diags []Diagnostic
	for _, n := range w.Nodes {
		cfg, ok := n.Config.(ir.HumanConfig)
		if !ok || cfg.Mode == "" {
			continue
		}
		if !validHumanModes[cfg.Mode] {
			diags = append(diags, Diagnostic{
				Code:     DIP127,
				Severity: SeverityWarning,
				Message:  fmt.Sprintf("node %q has mode %q which is not a recognized human mode", n.ID, cfg.Mode),
				Location: n.Source,
				Help:     "valid modes: choice, freeform, interview, yes_no",
			})
		}
	}
	return diags
}

func lintInterviewDefault(w *ir.Workflow) []Diagnostic {
	var diags []Diagnostic
	for _, n := range w.Nodes {
		cfg, ok := n.Config.(ir.HumanConfig)
		if !ok || cfg.Mode != "interview" {
			continue
		}
		if cfg.Default != "" {
			diags = append(diags, Diagnostic{
				Code:     DIP128,
				Severity: SeverityWarning,
				Message:  fmt.Sprintf("node %q is mode interview but has default %q which is ignored", n.ID, cfg.Default),
				Location: n.Source,
				Help:     "default is only meaningful for choice mode; remove it",
			})
		}
	}
	return diags
}

func lintInterviewLabeledEdges(w *ir.Workflow) []Diagnostic {
	var diags []Diagnostic
	for _, n := range w.Nodes {
		cfg, ok := n.Config.(ir.HumanConfig)
		if !ok || cfg.Mode != "interview" {
			continue
		}
		if d := checkInterviewLabels(n, w.Edges); d != nil {
			diags = append(diags, *d)
		}
	}
	return diags
}

// lintHumanChoiceKey flags outgoing edges from human nodes that route by a
// display label: with no explicit choice: key. In Phase 0 such a label is the
// routing key the runtime matches the user's selection against — load-bearing
// and even order-sensitive — yet nothing in the syntax says so. A Hint (not a
// Warning): these files route correctly today, so choice: is a clarity upgrade,
// not a defect.
func lintHumanChoiceKey(w *ir.Workflow) []Diagnostic {
	var diags []Diagnostic
	for _, e := range w.Edges {
		if !humanLabelRoutesWithoutChoice(w, e) {
			continue
		}
		diags = append(diags, Diagnostic{
			Code:     DIP150,
			Severity: SeverityHint,
			Message:  fmt.Sprintf("human gate %q routes by label %q; use choice: %q to mark the routing key (label: stays for display)", e.From, e.Label, e.Label),
			Location: e.Source,
			Help:     "add choice: \"<key>\" to mark the routing key explicitly; choice: wins when present, label: still routes when it is absent",
		})
	}
	return diags
}

// nonRoutingHumanModes are the human gate modes whose outgoing edge labels are
// display text, not the routing key the runtime matches a selection against:
// freeform takes open text, and interview collects answers (the same reason
// DIP129 says "interview does not route by label"). DIP150 does not flag their
// labeled edges. Every other mode — choice, yes_no, and the unset default (a
// mode-less human gate with labeled edges is the canonical multi-choice gate) —
// routes by label/selection in Phase 0.
var nonRoutingHumanModes = map[string]bool{
	"freeform":  true,
	"interview": true,
}

// humanLabelRoutesWithoutChoice reports whether edge e routes a human gate by a
// display label: with no explicit choice: key — the condition DIP150 flags.
func humanLabelRoutesWithoutChoice(w *ir.Workflow, e *ir.Edge) bool {
	if e.Label == "" || e.Choice != "" {
		return false
	}
	return sourceIsLabelRoutingHuman(w.Node(e.From))
}

// sourceIsLabelRoutingHuman reports whether n is a human gate whose mode routes
// by edge label (see nonRoutingHumanModes for the exclusions).
func sourceIsLabelRoutingHuman(n *ir.Node) bool {
	if n == nil || n.Kind != ir.NodeHuman {
		return false
	}
	cfg, ok := n.Config.(ir.HumanConfig)
	return ok && !nonRoutingHumanModes[cfg.Mode]
}

func checkInterviewLabels(n *ir.Node, edges []*ir.Edge) *Diagnostic {
	labelCount := 0
	for _, e := range edges {
		if e.From == n.ID && e.Label != "" {
			labelCount++
		}
	}
	if labelCount <= 1 {
		return nil
	}
	return &Diagnostic{
		Code:     DIP129,
		Severity: SeverityWarning,
		Message:  fmt.Sprintf("node %q is mode interview but has %d labeled edges (interview does not route by label)", n.ID, labelCount),
		Location: n.Source,
		Help:     "interview mode collects answers, not choices; use mode choice for label-based routing",
	}
}
