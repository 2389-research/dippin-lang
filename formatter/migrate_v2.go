// ABOUTME: v1 -> dip 2 IR transform. The retry channel (retry_target /
// ABOUTME: fallback_target) stays on the node — dip 2 re-admits it — so
// ABOUTME: migration is lossless; the formatter relabels the dip-2 spellings.
package formatter

import "github.com/2389-research/dippin-lang/ir"

// MigrationNote records a case the v1->v2 transform could not express 1:1.
// Retained for future migration passes; the retry-channel migration is now
// lossless and produces none.
type MigrationNote struct {
	Node    string
	Message string
}

// MigrateToV2 rewrites a v1 workflow into dip 2 in place and returns review
// notes. No-op if the workflow is already v2.
//
// dip 2 re-admits retry_target and adds fallback_retry_target as node
// attributes (the retry channel the engine reads — distinct from the edges
// block; see #204). So the v1 retry attributes are already dip-2-compatible:
// migration keeps them on the node and only bumps the version. The formatter
// emits the dip-2 spelling (fallback_target -> fallback_retry_target). This is
// lossless — it replaces the earlier lossy conversion of these attrs into
// loop / on-fail edges the retry engine never read (#186).
func MigrateToV2(w *ir.Workflow) []MigrationNote {
	if w.Version == "2" {
		return nil
	}
	w.Version = "2"
	return nil
}
