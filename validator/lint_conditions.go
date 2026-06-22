package validator

import (
	"fmt"
	"strings"

	"github.com/2389-research/dippin-lang/ir"
)

// lintOverlappingConditions checks DIP103: multiple edges from the same node
// with conditions that compare the same variable to the same value using "=".
// This indicates contradictory or duplicated routing logic.
func lintOverlappingConditions(w *ir.Workflow) []Diagnostic {
	var diags []Diagnostic

	edgesBySource := groupConditionalEdgesBySource(w)
	for from, edges := range edgesBySource {
		diags = append(diags, findOverlaps(from, edges)...)
	}
	return diags
}

// groupConditionalEdgesBySource groups edges with conditions by their source node.
func groupConditionalEdgesBySource(w *ir.Workflow) map[string][]*ir.Edge {
	edgesBySource := make(map[string][]*ir.Edge)
	for _, e := range w.Edges {
		if e.Condition != nil {
			edgesBySource[e.From] = append(edgesBySource[e.From], e)
		}
	}
	return edgesBySource
}

// condKey identifies a unique condition comparison, including negation.
type condKey struct {
	variable string
	op       string
	value    string
	negated  bool
}

// signedComparison pairs a comparison with its negation status.
type signedComparison struct {
	cmp     ir.CondCompare
	negated bool
}

// findOverlaps detects duplicate condition comparisons among edges from the same node.
func findOverlaps(from string, edges []*ir.Edge) []Diagnostic {
	var diags []Diagnostic
	seen := make(map[condKey]*ir.Edge)
	for _, e := range edges {
		comparisons := extractSignedComparisons(e.Condition.Parsed, false)
		diags = append(diags, checkComparisonOverlaps(from, e, comparisons, seen)...)
	}
	return diags
}

// checkComparisonOverlaps checks a set of comparisons against previously seen ones.
func checkComparisonOverlaps(from string, e *ir.Edge, comparisons []signedComparison, seen map[condKey]*ir.Edge) []Diagnostic {
	var diags []Diagnostic
	for _, sc := range comparisons {
		key := condKey{variable: sc.cmp.Variable, op: sc.cmp.Op, value: sc.cmp.Value, negated: sc.negated}
		if first, ok := seen[key]; ok {
			diags = append(diags, Diagnostic{
				Code:     DIP103,
				Severity: SeverityWarning,
				Message:  fmt.Sprintf("node %q has overlapping conditions: edges to %q and %q both check %s %s %s", from, first.To, e.To, sc.cmp.Variable, sc.cmp.Op, sc.cmp.Value),
				Location: e.Source,
				Help:     "review the conditions to ensure they route to different targets for different states",
			})
		} else {
			seen[key] = e
		}
	}
	return diags
}

// extractSignedComparisons recursively extracts CondCompare nodes with
// their negation status. This tracks whether each comparison is inside a CondNot.
func extractSignedComparisons(expr ir.ConditionExpr, negated bool) []signedComparison {
	switch e := expr.(type) {
	case ir.CondCompare:
		return []signedComparison{{cmp: e, negated: negated}}
	case ir.CondAnd:
		left := extractSignedComparisons(e.Left, negated)
		return append(left, extractSignedComparisons(e.Right, negated)...)
	case ir.CondOr:
		left := extractSignedComparisons(e.Left, negated)
		return append(left, extractSignedComparisons(e.Right, negated)...)
	case ir.CondNot:
		return extractSignedComparisons(e.Inner, !negated)
	default:
		return nil
	}
}

// extractComparisons recursively extracts all CondCompare nodes from a
// condition expression tree. This flattens AND/OR/NOT to find the leaf comparisons.
func extractComparisons(expr ir.ConditionExpr) []ir.CondCompare {
	switch e := expr.(type) {
	case ir.CondCompare:
		return []ir.CondCompare{e}
	case ir.CondAnd:
		return extractBinaryComparisons(e.Left, e.Right)
	case ir.CondOr:
		return extractBinaryComparisons(e.Left, e.Right)
	case ir.CondNot:
		return extractComparisons(e.Inner)
	default:
		return nil
	}
}

// extractBinaryComparisons extracts comparisons from both sides of a binary condition.
func extractBinaryComparisons(left, right ir.ConditionExpr) []ir.CondCompare {
	return append(extractComparisons(left), extractComparisons(right)...)
}

// lintConditionNamespace checks DIP120: condition variables should use a
// namespace prefix (ctx., graph., params., stack.). Bare variables like "outcome"
// work at runtime but are inconsistent with the spec convention.
func lintConditionNamespace(w *ir.Workflow) []Diagnostic {
	var diags []Diagnostic
	for _, e := range w.Edges {
		if e.Condition == nil || e.Condition.Parsed == nil {
			continue
		}
		diags = append(diags, checkConditionVarNamespaces(e)...)
	}
	return diags
}

// checkConditionVarNamespaces checks all comparisons in a single edge condition.
func checkConditionVarNamespaces(e *ir.Edge) []Diagnostic {
	comparisons := extractComparisons(e.Condition.Parsed)
	var diags []Diagnostic
	for _, cmp := range comparisons {
		if isBareVariable(cmp.Variable) {
			diags = append(diags, Diagnostic{
				Code:     DIP120,
				Severity: SeverityWarning,
				Message:  fmt.Sprintf("edge %s → %s: condition variable %q has no namespace prefix", e.From, e.To, cmp.Variable),
				Location: e.Source,
				Help:     fmt.Sprintf("use ctx.%s instead of %s for consistency", cmp.Variable, cmp.Variable),
			})
		}
	}
	return diags
}

// isBareVariable returns true if the variable has no known namespace prefix.
func isBareVariable(v string) bool {
	parts := strings.SplitN(v, ".", 2)
	return len(parts) < 2 || !knownNamespaces[parts[0]]
}
