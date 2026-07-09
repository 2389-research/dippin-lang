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
// dropped when it targets the node itself (no destination beyond the budget),
// deduped when an on-fail edge to it already exists, synthesized when absent, or
// flagged (keep both + comment) when it diverges from a different on-fail edge.
func migrateFallbackTarget(w *ir.Workflow, n *ir.Node) []MigrationNote {
	f := n.Retry.FallbackTarget
	if f == "" {
		return nil
	}
	n.Retry.FallbackTarget = ""
	if f == n.ID {
		return nil // fallback to self carries no destination beyond the retry budget
	}
	if hasFailEdgeTo(w, n.ID, f) {
		return nil // an on-fail edge to f already exists — dedupe
	}
	other, diverges := anyFailEdgeTarget(w, n.ID)
	if !diverges {
		w.Edges = append(w.Edges, newFailEdge(n.ID, f, ""))
		return nil
	}
	msg := fmt.Sprintf("MIGRATION: v1 fallback_target was %q (differs from the on-fail edge -> %q) — pick one", f, other)
	w.Edges = append(w.Edges, newFailEdge(n.ID, f, msg))
	return []MigrationNote{{Node: n.ID, Message: msg}}
}

// migrateRetryTarget drops a self retry_target (max_retries alone means re-run in
// place), dedupes against an existing loop edge, and otherwise converts a
// non-self retry_target into a `loop` edge + review note.
func migrateRetryTarget(w *ir.Workflow, n *ir.Node) []MigrationNote {
	r := n.Retry.RetryTarget
	if r == "" {
		return nil
	}
	n.Retry.RetryTarget = ""
	if r == n.ID || hasLoopEdgeTo(w, n.ID, r) {
		return nil
	}
	msg := fmt.Sprintf("MIGRATION: v1 retry_target -> %q (non-self) became a loop edge — verify the loop intent", r)
	w.Edges = append(w.Edges, &ir.Edge{From: n.ID, To: r, Restart: true, Comment: msg})
	return []MigrationNote{{Node: n.ID, Message: msg}}
}

// hasFailEdgeTo reports whether nodeID already has an on-fail edge to target.
func hasFailEdgeTo(w *ir.Workflow, nodeID, target string) bool {
	for _, e := range w.EdgesFrom(nodeID) {
		if e.To == target && ir.EdgeRoutesOnFail(e) {
			return true
		}
	}
	return false
}

// hasLoopEdgeTo reports whether nodeID already has a loop (restart) edge to target.
func hasLoopEdgeTo(w *ir.Workflow, nodeID, target string) bool {
	for _, e := range w.EdgesFrom(nodeID) {
		if e.To == target && e.Restart {
			return true
		}
	}
	return false
}

// anyFailEdgeTarget returns a representative on-fail edge target for nodeID, and
// whether one exists (used to detect a divergent fallback).
func anyFailEdgeTarget(w *ir.Workflow, nodeID string) (string, bool) {
	for _, e := range w.EdgesFrom(nodeID) {
		if ir.EdgeRoutesOnFail(e) {
			return e.To, true
		}
	}
	return "", false
}

// newFailEdge builds a `-> to on fail` edge with an optional leading comment.
func newFailEdge(from, to, comment string) *ir.Edge {
	return &ir.Edge{From: from, To: to, Comment: comment, Condition: &ir.Condition{
		Raw:    "ctx.outcome = fail",
		Parsed: ir.CondCompare{Variable: "ctx.outcome", Op: "=", Value: "fail"},
	}}
}
