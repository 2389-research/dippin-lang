package ir

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
