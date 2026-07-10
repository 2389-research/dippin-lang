package ir

import "slices"

// Edge represents a connection between nodes in the workflow graph.
type Edge struct {
	From      string
	To        string
	Label     string     // Display label / human choice text
	Choice    string     // Carried, not interpreted: explicit human-gate routing key (preferred over Label when set; #130)
	Condition *Condition // Edge guard; parser sets only Raw — Parsed (AST) stays nil until simulate.EnsureConditionsParsed()
	Weight    int        // Priority hint for edge selection
	Restart   bool       // Back-edge: triggers downstream clear + re-execution
	Override  bool       // Carried, not interpreted: human-authored validation override (tracker#271)
	Comment   string     // Optional leading `# ` line the formatter emits before this edge (migration review notes)
	Source    SourceLocation
}

// EdgeRoutesOnFail reports whether an edge's guard routes the failure outcome
// (ctx.outcome / outcome = fail / failure). Requires Condition.Parsed (populated
// by simulate.EnsureConditionsParsed); an edge whose AST is not yet parsed
// returns false.
func EdgeRoutesOnFail(e *Edge) bool {
	cmp, ok := ExtractEqualityCondition(e)
	return ok && isOutcomeVariable(cmp.Variable) && isFailOutcome(cmp.Value)
}

func isOutcomeVariable(v string) bool { return v == "ctx.outcome" || v == "outcome" }

func isFailOutcome(v string) bool { return v == "fail" || v == "failure" }

// IsRedundantFanEdge reports whether e merely repeats a parallel/fan_in fork
// already declared inline on a node's config, carrying no extra information —
// it is unconditional and attribute-free, and either From is a parallel node
// listing To as a target, or To is a fan_in node listing From as a source.
// Such an edge can be stripped without losing information; a conditional or
// attributed edge between the same nodes is NOT redundant.
func IsRedundantFanEdge(w *Workflow, e *Edge) bool {
	return edgeIsPlain(e) && (fromIsParallelTarget(w, e) || toIsFanInSource(w, e))
}

func fromIsParallelTarget(w *Workflow, e *Edge) bool {
	from := w.Node(e.From)
	if from == nil {
		return false
	}
	cfg, ok := from.Config.(ParallelConfig)
	return ok && slices.Contains(cfg.Targets, e.To)
}

func toIsFanInSource(w *Workflow, e *Edge) bool {
	to := w.Node(e.To)
	if to == nil {
		return false
	}
	cfg, ok := to.Config.(FanInConfig)
	return ok && slices.Contains(cfg.Sources, e.From)
}

// edgeIsPlain reports whether an edge carries no guard, label, or routing
// attribute — i.e. it conveys only "From connects to To".
func edgeIsPlain(e *Edge) bool {
	return e.Condition == nil && !edgeHasAttrs(e)
}

func edgeHasAttrs(e *Edge) bool {
	return e.Label != "" || e.Choice != "" || e.Comment != "" || edgeHasRoutingAttrs(e)
}

func edgeHasRoutingAttrs(e *Edge) bool {
	return e.Weight != 0 || e.Restart || e.Override
}

// Condition is a parsed, validated boolean expression attached to an edge.
type Condition struct {
	Raw    string        // Original source text
	Parsed ConditionExpr // AST for evaluation
}

// ConditionExpr is the AST for edge conditions.
type ConditionExpr interface {
	conditionExpr()
}

// CondAnd represents a logical AND of two conditions.
type CondAnd struct {
	Left, Right ConditionExpr
}

func (CondAnd) conditionExpr() {}

// CondOr represents a logical OR of two conditions.
type CondOr struct {
	Left, Right ConditionExpr
}

func (CondOr) conditionExpr() {}

// CondNot represents a logical negation.
type CondNot struct {
	Inner ConditionExpr
}

func (CondNot) conditionExpr() {}

// CondCompare represents a comparison between a context variable and a value.
// Variables use namespaced access: "ctx.outcome", "graph.goal", etc.
type CondCompare struct {
	Variable string // Namespaced: "ctx.outcome", "graph.goal"
	Op       string // "=", "==", "!=", "contains", "startswith", "endswith", "in"
	Value    string
}

func (CondCompare) conditionExpr() {}
