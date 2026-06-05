package validator

import (
	"fmt"

	"github.com/2389-research/dippin-lang/ir"
)

// lintAgentFailureRoute checks DIP144: an agent node that can fail at runtime
// but has no way to route that failure — no fail edge, no fallback_target, no
// bounded retry target, and no graph-level on_failure. Such a node is a silent
// dead-end. Scoped to agent nodes; the exit node and nodes with no outgoing
// edges are skipped (failure there ends the run, not a routing gap).
func lintAgentFailureRoute(w *ir.Workflow) []Diagnostic {
	var diags []Diagnostic
	for _, n := range w.Nodes {
		if needsFailureRoute(w, n) {
			diags = append(diags, Diagnostic{
				Code:     DIP144,
				Severity: SeverityWarning,
				Message:  fmt.Sprintf("agent node %q has no failure route (no fail edge, no fallback_target, no bounded retry, no graph on_failure)", n.ID),
				Location: n.Source,
				Help:     "add `-> <node> when ctx.outcome = fail`, set fallback_target:, add retry_target with max_retries, or declare a workflow-level on_failure:",
			})
		}
	}
	return diags
}

// needsFailureRoute reports whether an agent node lacks any failure route.
func needsFailureRoute(w *ir.Workflow, n *ir.Node) bool {
	if _, ok := n.Config.(ir.AgentConfig); !ok {
		return false
	}
	if n.ID == w.Exit {
		return false
	}
	outgoing := w.EdgesFrom(n.ID)
	if len(outgoing) == 0 {
		return false
	}
	return !hasFailureRoute(w, n, outgoing)
}

// hasFailureRoute reports whether the node has any failure-handling route.
func hasFailureRoute(w *ir.Workflow, n *ir.Node, outgoing []*ir.Edge) bool {
	if w.Defaults.OnFailure != "" && w.Node(w.Defaults.OnFailure) != nil {
		return true
	}
	if hasBoundedFailureTarget(n.Retry) {
		return true
	}
	return hasFailEdge(outgoing)
}

// hasBoundedFailureTarget reports whether retry config supplies a recovery target:
// a fallback_target, or a retry_target bounded by max_retries.
func hasBoundedFailureTarget(r ir.RetryConfig) bool {
	if r.FallbackTarget != "" {
		return true
	}
	return r.RetryTarget != "" && r.MaxRetries > 0
}

// hasFailEdge reports whether any outgoing edge routes on outcome = fail/failure.
// An unconditional/success edge does NOT count — a hard failure does not traverse it.
func hasFailEdge(edges []*ir.Edge) bool {
	for _, e := range edges {
		cmp, ok := extractEqualityCondition(e)
		if ok && isOutcomeVar(cmp.Variable) && isFailValue(cmp.Value) {
			return true
		}
	}
	return false
}

// isOutcomeVar matches the outcome variable in both namespaced and bare forms.
func isOutcomeVar(v string) bool { return v == "ctx.outcome" || v == "outcome" }

// isFailValue matches the failure outcome values the engine recognizes.
func isFailValue(v string) bool { return v == "fail" || v == "failure" }
