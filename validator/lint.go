package validator

import (
	"github.com/2389-research/dippin-lang/ir"
	"github.com/2389-research/dippin-lang/simulate"
)

// Lint runs all semantic quality checks (DIP101–DIP157, except DIP138 —
// reserved, no firing logic — and DIP146 — which the CLI's cross-file pass
// emits, not this function) on the workflow
// and returns all diagnostics found. These are warnings, not errors —
// the workflow can still execute, but the findings indicate likely bugs
// or quality issues.
//
// Lint is independent of Validate. Callers should run both:
//
//	structureResult := validator.Validate(w)
//	lintResult := validator.Lint(w)
func Lint(w *ir.Workflow) Result {
	return LintWithOptions(w, Options{})
}

// Options carries per-invocation lint configuration. It is passed by value so
// each call is fully scoped — nothing here mutates package-level state, which
// keeps lint runs from leaking into one another (e.g. across validate/doctor/lint).
type Options struct {
	// ExtraModels extends the known model catalog for the DIP108 check only.
	// nil is treated as empty.
	ExtraModels ExtraModels
}

// LintWithOptions runs all semantic quality checks like Lint, but with the given
// per-invocation Options (currently a scoped extra-models catalog for DIP108).
func LintWithOptions(w *ir.Workflow, opts Options) Result {
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
	for _, pass := range lintPasses(opts) {
		diags = append(diags, pass(w)...)
	}
	return Result{Diagnostics: diags}
}

// lintPasses returns the ordered semantic-lint passes. Every pass is a pure
// func(*ir.Workflow) []Diagnostic; lintModelProvider is wrapped to close over
// opts.ExtraModels. Order is irrelevant to correctness (diagnostics are
// independent), so new passes append to the end.
func lintPasses(opts Options) []func(*ir.Workflow) []Diagnostic {
	return []func(*ir.Workflow) []Diagnostic{
		lintConditionalReachability,
		lintDefaultEdge,
		lintOverlappingConditions,
		lintUnboundedRetry,
		lintSuccessPath,
		lintUndefinedVariables,
		lintUnusedWrites,
		func(w *ir.Workflow) []Diagnostic { return lintModelProvider(w, opts.ExtraModels) },
		lintNamespaceCollisions,
		lintEmptyPrompts,
		lintToolTimeout,
		lintReadsWithoutUpstreamWrites,
		lintRetryPolicy,
		lintRetryRestartConfusion,
		lintFidelity,
		lintGoalGateFallback,
		lintCompactionThreshold,
		lintOnResume,
		lintReasoningEffort,
		lintConditionNamespace,
		lintStylesheetRefs,
		lintConditionUndefinedOutput,
		lintConditionUndeclaredValue,
		lintToolSyntax,
		lintToolCtxVars,
		lintToolBinary,
		lintSubgraphRef,
		lintManagerLoop,
		lintHumanMode,
		lintInterviewDefault,
		lintInterviewLabeledEdges,
		lintHumanChoiceKey,
		lintResponseFormat,
		lintResponseSchemaMismatch,
		lintResponseSchemaJSON,
		lintAgentParamsShadow,
		lintToolAccessValues,
		lintParamsReenablesTools,
		lintWritablePaths,
		lintSubgraphToolAccess,
		lintAgentFailureRoute,
		lintBudgetRanges,
		lintChainAttack,
		lintLastResponseTruncate,
		lintAmbiguousRouting,
		lintUnusedWeight,
		lintMarkerCoverage,
		lintRedundantFanEdge,
		lintPromptOptOut,
		lintUnknownInputType,
		lintUndeclaredInputRef,
		lintInputInToolCommand,
	}
}

// knownNamespaces lists the valid namespace prefixes for variable references.
// Per §8.2 of the Dippin spec: ctx. (runtime context), graph. (workflow-level
// attributes), params. (module parameters for composition), stack. (supervisor
// state exposed by manager_loop, e.g., stack.child.cycles).
// node.* is intentionally excluded: it requires structural validation
// (node ID must exist, ref must have exactly 3 parts) handled in isVarRefValid.
// Used by lint_context.go (DIP106) and lint_conditions.go (DIP120).
//
// inputs. (declared caller-supplied values) is unlike the others: membership
// here only stops DIP106/DIP120 flagging the prefix. It is a *closed*
// namespace — every key is additionally resolved against Workflow.Inputs by
// DIP156 in lint_inputs.go.
var knownNamespaces = map[string]bool{
	"ctx":    true,
	"graph":  true,
	"params": true,
	"stack":  true,
	"inputs": true,
}

// buildForwardAdjacency builds a forward adjacency map for non-restart edges,
// including implicit edges from parallel and fan_in nodes.
// Used by lint_reachability.go (DIP105) and lint_context.go (DIP112).
func buildForwardAdjacency(w *ir.Workflow) map[string][]string {
	adj := buildNonRestartAdjacency(w)
	addParallelFanInEdges(adj, w)
	return adj
}
