package validator

import (
	"github.com/2389-research/dippin-lang/ir"
	"github.com/2389-research/dippin-lang/simulate"
)

// Lint runs all semantic quality checks (DIP101–DIP145) on the workflow
// and returns all diagnostics found. These are warnings, not errors —
// the workflow can still execute, but the findings indicate likely bugs
// or quality issues.
//
// Lint is independent of Validate. Callers should run both:
//
//	structureResult := validator.Validate(w)
//	lintResult := validator.Lint(w)
func Lint(w *ir.Workflow) Result {
	// Ensure condition ASTs are populated — the AST-dependent lint checks
	// (DIP103/120/121/122) read edge Condition.Parsed.
	//
	// parseEdgeConditions populates every parseable edge accumulate-all, so one
	// unparseable edge does not mask those lints on later edges (the bug behind
	// DIP010 / issue #98). Edge parse failures are reported as DIP010 from
	// Validate; here we only need the population side effect.
	//
	// EnsureConditionsParsed then populates manager_loop node conditions in the
	// common case. If an edge is unparseable it returns early before reaching the
	// node loop, leaving node conditions unparsed — but that changes no lint
	// result: the manager_loop checks gate on condition *presence* via
	// conditionPresent (Raw != "" || Parsed != nil), which is Raw-dominated for
	// any declared condition, and no lint inspects a node condition's parsed AST.
	parseEdgeConditions(w)
	_ = simulate.EnsureConditionsParsed(w)

	var diags []Diagnostic

	diags = append(diags, lintConditionalReachability(w)...)
	diags = append(diags, lintDefaultEdge(w)...)
	diags = append(diags, lintOverlappingConditions(w)...)
	diags = append(diags, lintUnboundedRetry(w)...)
	diags = append(diags, lintSuccessPath(w)...)
	diags = append(diags, lintUndefinedVariables(w)...)
	diags = append(diags, lintUnusedWrites(w)...)
	diags = append(diags, lintModelProvider(w)...)
	diags = append(diags, lintNamespaceCollisions(w)...)
	diags = append(diags, lintEmptyPrompts(w)...)
	diags = append(diags, lintToolTimeout(w)...)
	diags = append(diags, lintReadsWithoutUpstreamWrites(w)...)
	diags = append(diags, lintRetryPolicy(w)...)
	diags = append(diags, lintRetryRestartConfusion(w)...)
	diags = append(diags, lintFidelity(w)...)
	diags = append(diags, lintGoalGateFallback(w)...)
	diags = append(diags, lintCompactionThreshold(w)...)
	diags = append(diags, lintOnResume(w)...)
	diags = append(diags, lintReasoningEffort(w)...)
	diags = append(diags, lintConditionNamespace(w)...)
	diags = append(diags, lintStylesheetRefs(w)...)
	diags = append(diags, lintConditionUndefinedOutput(w)...)
	diags = append(diags, lintConditionUndeclaredValue(w)...)
	diags = append(diags, lintToolSyntax(w)...)
	diags = append(diags, lintToolCtxVars(w)...)
	diags = append(diags, lintToolBinary(w)...)
	diags = append(diags, lintSubgraphRef(w)...)
	diags = append(diags, lintManagerLoop(w)...)
	diags = append(diags, lintHumanMode(w)...)
	diags = append(diags, lintInterviewDefault(w)...)
	diags = append(diags, lintInterviewLabeledEdges(w)...)
	diags = append(diags, lintResponseFormat(w)...)
	diags = append(diags, lintResponseSchemaMismatch(w)...)
	diags = append(diags, lintResponseSchemaJSON(w)...)
	diags = append(diags, lintAgentParamsShadow(w)...)
	diags = append(diags, lintToolAccessValues(w)...)
	diags = append(diags, lintParamsReenablesTools(w)...)
	diags = append(diags, lintWritablePaths(w)...)
	diags = append(diags, lintSubgraphToolAccess(w)...)
	diags = append(diags, lintAgentFailureRoute(w)...)
	diags = append(diags, lintBudgetRanges(w)...)

	return Result{Diagnostics: diags}
}

// knownNamespaces lists the valid namespace prefixes for variable references.
// Per §8.2 of the Dippin spec: ctx. (runtime context), graph. (workflow-level
// attributes), params. (module parameters for composition), stack. (supervisor
// state exposed by manager_loop, e.g., stack.child.cycles).
// node.* is intentionally excluded: it requires structural validation
// (node ID must exist, ref must have exactly 3 parts) handled in isVarRefValid.
// Used by lint_context.go (DIP106) and lint_conditions.go (DIP120).
var knownNamespaces = map[string]bool{
	"ctx":    true,
	"graph":  true,
	"params": true,
	"stack":  true,
}

// buildForwardAdjacency builds a forward adjacency map for non-restart edges,
// including implicit edges from parallel and fan_in nodes.
// Used by lint_reachability.go (DIP105) and lint_context.go (DIP112).
func buildForwardAdjacency(w *ir.Workflow) map[string][]string {
	adj := buildNonRestartAdjacency(w)
	addParallelFanInEdges(adj, w)
	return adj
}
