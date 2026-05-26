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
