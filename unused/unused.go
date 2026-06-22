// Package unused detects dead-branch nodes in a workflow.
//
// A node is "unused" if it is reachable from start but has no path to exit
// (a sink node). The package computes the wasted cost of these nodes by
// combining coverage analysis (which finds sinks) with cost analysis.
package unused

import (
	"github.com/2389-research/dippin-lang/cost"
	"github.com/2389-research/dippin-lang/coverage"
	"github.com/2389-research/dippin-lang/ir"
)

// Report is the result of unused-node analysis.
type Report struct {
	UnusedNodes []UnusedNode   `json:"unused_nodes"`
	TotalWasted cost.CostRange `json:"total_wasted"`
}

// UnusedNode describes a single dead-branch node and its wasted cost.
type UnusedNode struct {
	NodeID     string         `json:"node_id"`
	Label      string         `json:"label"`
	Kind       string         `json:"kind"`
	WastedCost cost.CostRange `json:"wasted_cost"`
}

// Analyze detects unused (sink) nodes and estimates wasted cost.
func Analyze(w *ir.Workflow) *Report {
	sinks := coverage.Analyze(w).Termination.SinkNodes
	costs := cost.Analyze(w, cost.DefaultPricing())
	return buildReport(w, sinks, costs)
}

// buildReport constructs the report from sink nodes and cost data.
func buildReport(w *ir.Workflow, sinks []string, costs *cost.Report) *Report {
	r := &Report{}
	for _, id := range sinks {
		node := buildUnusedNode(w, id, costs)
		r.UnusedNodes = append(r.UnusedNodes, node)
		r.TotalWasted = cost.AddCostRange(r.TotalWasted, node.WastedCost)
	}
	return r
}

// buildUnusedNode creates an UnusedNode entry for the given node ID.
func buildUnusedNode(w *ir.Workflow, id string, costs *cost.Report) UnusedNode {
	un := UnusedNode{NodeID: id}
	if n := w.Node(id); n != nil {
		un.Label = n.Label
		un.Kind = string(n.Kind)
	}
	if nc, ok := costs.Nodes[id]; ok {
		un.WastedCost = nc.Cost
	}
	return un
}
