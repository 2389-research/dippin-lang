package validator

import (
	"fmt"

	"github.com/2389-research/dippin-lang/ir"
)

// lintUnboundedRetry checks DIP104: nodes with retry configuration that have
// no max_retries limit and no fallback target. This could cause infinite retry
// loops at runtime.
func lintUnboundedRetry(w *ir.Workflow) []Diagnostic {
	var diags []Diagnostic
	for _, n := range w.Nodes {
		if isUnboundedRetry(w, n) {
			diags = append(diags, Diagnostic{
				Code:     DIP104,
				Severity: SeverityWarning,
				Message:  fmt.Sprintf("node %q has retry configuration but no max_retries and no failure route", n.ID),
				Location: n.Source,
				Help:     "set max_retries to bound retries, or add an `on fail` edge / fallback_target for graceful degradation",
			})
		}
	}
	return diags
}

// isUnboundedRetry returns true if retry config exists with no bound: no
// max_retries and no failure route (fallback_target or `on fail` edge).
func isUnboundedRetry(w *ir.Workflow, n *ir.Node) bool {
	r := n.Retry
	hasRetryConfig := r.Policy != "" || r.RetryTarget != ""
	return hasRetryConfig && r.MaxRetries == 0 && !nodeHasFailureRoute(w, n)
}

// validRetryPolicies is the set of retry policy names recognized by the runtime engine.
var validRetryPolicies = map[string]bool{
	"standard":   true,
	"aggressive": true,
	"patient":    true,
	"linear":     true,
	"none":       true,
}

// lintRetryPolicy checks DIP113: retry_policy must be one of the known policy names.
func lintRetryPolicy(w *ir.Workflow) []Diagnostic {
	var diags []Diagnostic
	diags = append(diags, checkDefaultRetryPolicy(w)...)
	diags = append(diags, checkNodeRetryPolicies(w)...)
	return diags
}

// checkDefaultRetryPolicy checks the workflow-level default retry policy.
func checkDefaultRetryPolicy(w *ir.Workflow) []Diagnostic {
	if p := w.Defaults.RetryPolicy; p != "" && !validRetryPolicies[p] {
		return []Diagnostic{{
			Code:     DIP113,
			Severity: SeverityWarning,
			Message:  fmt.Sprintf("workflow default retry_policy %q is not a recognized policy name", p),
			Help:     "valid policies: standard, aggressive, patient, linear, none",
		}}
	}
	return nil
}

// checkNodeRetryPolicies checks per-node retry policies.
func checkNodeRetryPolicies(w *ir.Workflow) []Diagnostic {
	var diags []Diagnostic
	for _, n := range w.Nodes {
		if p := n.Retry.Policy; p != "" && !validRetryPolicies[p] {
			diags = append(diags, Diagnostic{
				Code:     DIP113,
				Severity: SeverityWarning,
				Message:  fmt.Sprintf("node %q has retry_policy %q which is not a recognized policy name", n.ID, p),
				Location: n.Source,
				Help:     "valid policies: standard, aggressive, patient, linear, none",
			})
		}
	}
	return diags
}

// lintRetryRestartConfusion checks DIP134: if the workflow has restart: true
// edges and max_retries is set in defaults but max_restarts is not, the user
// likely confused the two. max_retries controls per-node LLM retries;
// max_restarts controls the global loop restart budget.
func lintRetryRestartConfusion(w *ir.Workflow) []Diagnostic {
	if !hasRestartEdges(w) {
		return nil
	}
	if w.Defaults.MaxRetries == 0 || w.Defaults.MaxRestarts != 0 {
		return nil
	}
	return []Diagnostic{{
		Code:     DIP134,
		Severity: SeverityWarning,
		Message:  fmt.Sprintf("defaults specify max_retries: %d but no max_restarts, and the workflow has restart: true edges", w.Defaults.MaxRetries),
		Help:     "max_retries controls per-node LLM retries; max_restarts controls the loop restart budget for restart: true edges — did you mean max_restarts?",
	}}
}

// hasRestartEdges returns true if any edge has Restart: true.
func hasRestartEdges(w *ir.Workflow) bool {
	for _, e := range w.Edges {
		if e.Restart {
			return true
		}
	}
	return false
}

// lintGoalGateFallback checks DIP115: nodes with goal_gate: true should have
// a retry_target, fallback_target, or `on fail` edge so the pipeline has a
// recovery path when the gate fails.
func lintGoalGateFallback(w *ir.Workflow) []Diagnostic {
	var diags []Diagnostic
	for _, n := range w.Nodes {
		if needsGoalGateFallback(w, n) {
			diags = append(diags, Diagnostic{
				Code:     DIP115,
				Severity: SeverityWarning,
				Message:  fmt.Sprintf("node %q has goal_gate: true but no failure route", n.ID),
				Location: n.Source,
				Help:     "add an `on fail` edge, set fallback_target:, or add retry_target with max_retries so the pipeline can recover when the gate fails",
			})
		}
	}
	return diags
}

// needsGoalGateFallback returns true if a goal_gate node has no recovery path.
func needsGoalGateFallback(w *ir.Workflow, n *ir.Node) bool {
	cfg, ok := n.Config.(ir.AgentConfig)
	if !ok || !cfg.GoalGate {
		return false
	}
	return !nodeHasFailureRoute(w, n)
}
