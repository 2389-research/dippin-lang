// Package diff performs semantic comparison between two Dippin workflows.
//
// Unlike text-based diff, this compares the workflow graph structure:
// nodes added/removed/modified, edges added/removed, and field-level
// changes on agent node configurations. It also computes cost delta.
package diff

import (
	"sort"
	"strconv"

	"github.com/2389-research/dippin-lang/cost"
	"github.com/2389-research/dippin-lang/ir"
)

// Report is the full semantic diff result.
type Report struct {
	NodesAdded    []string      `json:"nodes_added"`
	NodesRemoved  []string      `json:"nodes_removed"`
	NodesModified []NodeDiff    `json:"nodes_modified"`
	EdgesAdded    []EdgeSummary `json:"edges_added"`
	EdgesRemoved  []EdgeSummary `json:"edges_removed"`
	CostDelta     CostDelta     `json:"cost_delta"`
}

// NodeDiff describes field-level changes to a modified node.
type NodeDiff struct {
	NodeID  string        `json:"node_id"`
	Changes []FieldChange `json:"changes"`
}

// FieldChange records a single field that changed.
type FieldChange struct {
	Field    string `json:"field"`
	OldValue string `json:"old_value"`
	NewValue string `json:"new_value"`
}

// EdgeSummary identifies an edge by its endpoints and optional condition.
type EdgeSummary struct {
	From      string `json:"from"`
	To        string `json:"to"`
	Condition string `json:"condition,omitempty"`
}

// CostDelta shows old vs new cost ranges and the difference.
type CostDelta struct {
	OldCost cost.CostRange `json:"old_cost"`
	NewCost cost.CostRange `json:"new_cost"`
	Delta   cost.CostRange `json:"delta"`
}

// Compare produces a semantic diff between old and new workflows.
func Compare(old, new *ir.Workflow) *Report {
	r := &Report{}
	compareNodes(old, new, r)
	compareEdges(old, new, r)
	r.CostDelta = compareCosts(old, new)
	return r
}

// compareNodes finds added, removed, and modified nodes.
func compareNodes(old, new *ir.Workflow, r *Report) {
	oldNodes := nodeMap(old)
	newNodes := nodeMap(new)

	r.NodesAdded = findAddedIDs(oldNodes, newNodes)
	r.NodesRemoved = findAddedIDs(newNodes, oldNodes) // reversed args
	r.NodesModified = findModifiedNodes(oldNodes, newNodes)
}

// nodeMap builds a map from node ID to node.
func nodeMap(w *ir.Workflow) map[string]*ir.Node {
	m := make(map[string]*ir.Node, len(w.Nodes))
	for _, n := range w.Nodes {
		m[n.ID] = n
	}
	return m
}

// findAddedIDs returns sorted IDs present in b but not a.
func findAddedIDs(a, b map[string]*ir.Node) []string {
	var added []string
	for id := range b {
		if _, ok := a[id]; !ok {
			added = append(added, id)
		}
	}
	sort.Strings(added)
	return added
}

// findModifiedNodes compares nodes present in both workflows.
func findModifiedNodes(oldNodes, newNodes map[string]*ir.Node) []NodeDiff {
	var diffs []NodeDiff
	for id, oldNode := range oldNodes {
		newNode, ok := newNodes[id]
		if !ok {
			continue
		}
		changes := compareNodeFields(oldNode, newNode)
		if len(changes) > 0 {
			diffs = append(diffs, NodeDiff{NodeID: id, Changes: changes})
		}
	}
	sort.Slice(diffs, func(i, j int) bool {
		return diffs[i].NodeID < diffs[j].NodeID
	})
	return diffs
}

// compareNodeFields compares label, kind, and agent-config fields of two nodes
// (other config kinds are not field-diffed).
func compareNodeFields(old, new *ir.Node) []FieldChange {
	var changes []FieldChange
	if old.Label != new.Label {
		changes = append(changes, FieldChange{"label", old.Label, new.Label})
	}
	if string(old.Kind) != string(new.Kind) {
		changes = append(changes, FieldChange{"kind", string(old.Kind), string(new.Kind)})
	}
	changes = append(changes, compareConfigFields(old, new)...)
	return changes
}

// compareConfigFields compares agent config fields between two nodes.
func compareConfigFields(old, new *ir.Node) []FieldChange {
	oldAC, oldOK := old.Config.(ir.AgentConfig)
	newAC, newOK := new.Config.(ir.AgentConfig)
	if !oldOK || !newOK {
		return nil
	}
	return diffAgentConfig(oldAC, newAC)
}

// agentField defines a comparison between old and new AgentConfig values.
type agentField struct {
	name   string
	oldVal string
	newVal string
}

// diffAgentConfig compares specific AgentConfig fields using a table.
func diffAgentConfig(old, new ir.AgentConfig) []FieldChange {
	fields := agentFieldTable(old, new)
	var changes []FieldChange
	for _, f := range fields {
		if f.oldVal != f.newVal {
			changes = append(changes, FieldChange{f.name, f.oldVal, f.newVal})
		}
	}
	return changes
}

// agentFieldTable builds the comparison table for agent config fields.
func agentFieldTable(old, new ir.AgentConfig) []agentField {
	return []agentField{
		{"model", old.Model, new.Model},
		{"provider", old.Provider, new.Provider},
		{"prompt", truncate(old.Prompt), truncate(new.Prompt)},
		{"max_turns", strconv.Itoa(old.MaxTurns), strconv.Itoa(new.MaxTurns)},
		{"reasoning_effort", old.ReasoningEffort, new.ReasoningEffort},
		{"fidelity", old.Fidelity, new.Fidelity},
	}
}

// compareEdges finds added and removed edges.
func compareEdges(old, new *ir.Workflow, r *Report) {
	oldEdges := edgeSet(old)
	newEdges := edgeSet(new)
	r.EdgesAdded = findAddedEdges(oldEdges, newEdges)
	r.EdgesRemoved = findAddedEdges(newEdges, oldEdges)
}

// edgeKey produces a comparable string key for an edge.
func edgeKey(e *ir.Edge) string {
	cond := ""
	if e.Condition != nil {
		cond = e.Condition.Raw
	}
	return e.From + "->" + e.To + ":" + cond
}

// edgeSet builds a map from edge key to EdgeSummary.
func edgeSet(w *ir.Workflow) map[string]EdgeSummary {
	m := make(map[string]EdgeSummary, len(w.Edges))
	for _, e := range w.Edges {
		cond := ""
		if e.Condition != nil {
			cond = e.Condition.Raw
		}
		m[edgeKey(e)] = EdgeSummary{From: e.From, To: e.To, Condition: cond}
	}
	return m
}

// findAddedEdges returns edges in b but not a.
func findAddedEdges(a, b map[string]EdgeSummary) []EdgeSummary {
	var added []EdgeSummary
	for key, es := range b {
		if _, ok := a[key]; !ok {
			added = append(added, es)
		}
	}
	sort.Slice(added, func(i, j int) bool {
		return added[i].From+added[i].To < added[j].From+added[j].To
	})
	return added
}

// compareCosts computes cost delta between workflows.
func compareCosts(old, new *ir.Workflow) CostDelta {
	pricing := cost.DefaultPricing()
	oldCost := cost.Analyze(old, pricing).Total
	newCost := cost.Analyze(new, pricing).Total
	return CostDelta{
		OldCost: oldCost,
		NewCost: newCost,
		Delta: cost.CostRange{
			Min:      newCost.Min - oldCost.Min,
			Expected: newCost.Expected - oldCost.Expected,
			Max:      newCost.Max - oldCost.Max,
		},
	}
}

// promptDiffMaxLen bounds the prompt preview shown in field-change tables.
const promptDiffMaxLen = 50

// truncate shortens a string to promptDiffMaxLen, adding "..." if truncated.
func truncate(s string) string {
	if len(s) <= promptDiffMaxLen {
		return s
	}
	return s[:promptDiffMaxLen] + "..."
}
