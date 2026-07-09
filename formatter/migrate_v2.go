// ABOUTME: v1 -> dip 2 IR transform: retry_target/fallback_target node fields
// ABOUTME: become on-fail / loop edges. Used by `dippin fmt --migrate`.
package formatter

import (
	"fmt"

	"github.com/2389-research/dippin-lang/ir"
)

// MigrationNote records a case the v1->v2 transform could not express 1:1.
type MigrationNote struct {
	Node    string
	Message string
}

// MigrateToV2 rewrites a v1 workflow into dip 2 in place and returns review
// notes. Requires Condition.Parsed to be populated (caller runs
// simulate.EnsureConditionsParsed). No-op if the workflow is already v2.
func MigrateToV2(w *ir.Workflow) []MigrationNote {
	if w.Version == "2" {
		return nil
	}
	var notes []MigrationNote
	for _, n := range w.Nodes {
		notes = append(notes, migrateFallbackTarget(w, n)...)
		notes = append(notes, migrateRetryTarget(w, n)...)
	}
	w.Version = "2"
	return notes
}

// migrateFallbackTarget turns a node's fallback_target into an `on fail` edge:
// synthesize when absent, dedupe when it matches an existing fail edge, or flag
// (keep both + comment) when it diverges.
func migrateFallbackTarget(w *ir.Workflow, n *ir.Node) []MigrationNote {
	f := n.Retry.FallbackTarget
	if f == "" {
		return nil
	}
	n.Retry.FallbackTarget = ""
	existing := failEdgeTarget(w, n.ID)
	if existing == f {
		return nil // destinations agree — dedupe
	}
	if existing == "" {
		w.Edges = append(w.Edges, newFailEdge(n.ID, f, ""))
		return nil
	}
	msg := fmt.Sprintf("MIGRATION: v1 fallback_target was %q (differs from the on-fail edge -> %q) — pick one", f, existing)
	w.Edges = append(w.Edges, newFailEdge(n.ID, f, msg))
	return []MigrationNote{{Node: n.ID, Message: msg}}
}

// migrateRetryTarget drops a self retry_target (max_retries alone means re-run in
// place) and converts a non-self retry_target into a `loop` edge + review note.
func migrateRetryTarget(w *ir.Workflow, n *ir.Node) []MigrationNote {
	r := n.Retry.RetryTarget
	if r == "" {
		return nil
	}
	n.Retry.RetryTarget = ""
	if r == n.ID {
		return nil
	}
	msg := fmt.Sprintf("MIGRATION: v1 retry_target -> %q (non-self) became a loop edge — verify the loop intent", r)
	w.Edges = append(w.Edges, &ir.Edge{From: n.ID, To: r, Restart: true, Comment: msg})
	return []MigrationNote{{Node: n.ID, Message: msg}}
}

// failEdgeTarget returns the target of the first outgoing on-fail edge, or "".
func failEdgeTarget(w *ir.Workflow, nodeID string) string {
	for _, e := range w.EdgesFrom(nodeID) {
		if ir.EdgeRoutesOnFail(e) {
			return e.To
		}
	}
	return ""
}

// newFailEdge builds a `-> to on fail` edge with an optional leading comment.
func newFailEdge(from, to, comment string) *ir.Edge {
	return &ir.Edge{From: from, To: to, Comment: comment, Condition: &ir.Condition{
		Raw:    "ctx.outcome = fail",
		Parsed: ir.CondCompare{Variable: "ctx.outcome", Op: "=", Value: "fail"},
	}}
}
