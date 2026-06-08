package validator

import (
	"fmt"

	"github.com/2389-research/dippin-lang/ir"
	"github.com/2389-research/dippin-lang/simulate"
)

// This file holds the one structural check that needs to parse edge conditions
// (DIP010). It is separate from validate.go so the dependency on the simulate
// package — the only cross-package import the validator's structural checks
// have — stays isolated to a single file. simulate.ParseCondition is pure
// string→AST (no filesystem), so this remains wasm-safe.

// edgeNeedsParse reports whether an edge's condition still needs parsing.
// Mirrors the guard in simulate.ensureEdgeConditionParsed: skip conditions that
// are absent, already parsed, or have no raw text.
func edgeNeedsParse(e *ir.Edge) bool {
	return e.Condition != nil && e.Condition.Parsed == nil && e.Condition.Raw != ""
}

// edgeParseFailure pairs an edge with the error from parsing its condition.
// It is not named *Error because it does not implement the error interface.
type edgeParseFailure struct {
	edge *ir.Edge
	err  error
}

// parseEdgeConditions parses every edge condition that needs it, populating
// Condition.Parsed for each that succeeds. Unlike simulate.EnsureConditionsParsed
// it does NOT stop at the first failure — every parseable edge is populated, so
// one bad condition cannot mask the AST-dependent lints (DIP103/120/121/122) on
// later edges. It returns one failure per unparseable edge.
//
// This mutates w by lazily populating Condition.Parsed, consistent with how the
// codebase treats Parsed as lazily-populated shared state. The population is
// idempotent (guarded by Parsed != nil).
func parseEdgeConditions(w *ir.Workflow) []edgeParseFailure {
	var failures []edgeParseFailure
	for _, e := range w.Edges {
		if f := parseOneEdge(e); f != nil {
			failures = append(failures, *f)
		}
	}
	return failures
}

// parseOneEdge parses a single edge's condition, populating Parsed on success or
// returning a failure on error. Split out to keep parseEdgeConditions under the
// cognitive-complexity ceiling (mirrors the checkEndpoint extraction idiom).
func parseOneEdge(e *ir.Edge) *edgeParseFailure {
	if !edgeNeedsParse(e) {
		return nil
	}
	parsed, err := simulate.ParseCondition(e.Condition.Raw)
	if err != nil {
		return &edgeParseFailure{edge: e, err: err}
	}
	e.Condition.Parsed = parsed
	return nil
}

// checkEdgeConditions verifies DIP010: every edge condition can be parsed.
// An unparseable condition leaves the edge's routing undefined — the workflow
// cannot execute deterministically — so it is a structural error, emitted here
// (from Validate) rather than as a lint warning, so every command path catches it.
func checkEdgeConditions(w *ir.Workflow) []Diagnostic {
	var diags []Diagnostic
	for _, f := range parseEdgeConditions(w) {
		diags = append(diags, Diagnostic{
			Code:     DIP010,
			Severity: SeverityError,
			Message:  fmt.Sprintf("edge %s -> %s: invalid condition %q: %v", f.edge.From, f.edge.To, f.edge.Condition.Raw, f.err),
			Location: f.edge.Source,
			Help:     "valid operators: = == != contains startswith endswith in (combine with and/or/not)",
		})
	}
	return diags
}
