// ABOUTME: Exhaustiveness analysis over a node's outgoing edge conditions.
// ABOUTME: Shared by validator (DIP101/DIP102 suppression) and simulate (else routing).
package ir

// knownExhaustiveSets lists value sets that are known to be mutually exhaustive
// for a given variable. If a node's outgoing edges collectively cover all values
// in any set for a variable, the conditions are exhaustive.
var knownExhaustiveSets = map[string][][]string{
	"ctx.outcome": {
		{"success", "fail"},
		{"success", "failure"},
	},
	"outcome": {
		{"success", "fail"},
		{"success", "failure"},
	},
}

// EdgesExhaustive returns true if a set of sibling edges forms an exhaustive
// condition set — i.e. every runtime case is guaranteed to match some edge.
// Three detection strategies (any match = exhaustive):
//
//  1. Known value sets — e.g., outcome = success + outcome = fail
//  2. Complete partition — all conditional edges test the same variable
//     with equality and there are 2+ values (author declares "these are
//     the only cases")
//  3. Complementary pair — e.g., "contains X" + "not contains X"
func EdgesExhaustive(edges []*Edge) bool {
	byVar := collectConditionValues(edges)
	if matchesExhaustiveSet(byVar) {
		return true
	}
	if isCompletePartition(edges, byVar) {
		return true
	}
	return hasComplementaryPair(edges)
}

// ExtractEqualityCondition returns the CondCompare if the edge has a simple
// equality condition (= or ==), and false otherwise.
func ExtractEqualityCondition(e *Edge) (CondCompare, bool) {
	if !edgeHasParsedCondition(e) {
		return CondCompare{}, false
	}
	cmp, ok := e.Condition.Parsed.(CondCompare)
	if !ok || !isEqualityOp(cmp.Op) {
		return CondCompare{}, false
	}
	return cmp, true
}

// isCompletePartition returns true if every conditional edge tests the same
// variable with equality and there are 2+ distinct values. This means the
// author has partitioned all routing on a single variable — the conditions
// cover all intended cases by construction.
func isCompletePartition(edges []*Edge, byVar map[string]map[string]bool) bool {
	if len(byVar) != 1 {
		return false // conditions span multiple variables
	}
	conditionalCount := countConditionalEdges(edges)
	if conditionalCount < 2 {
		return false // need at least 2 branches to form a partition
	}
	equalityCount := countEqualityEdges(edges)
	return equalityCount == conditionalCount
}

// countConditionalEdges returns the number of edges with a condition.
func countConditionalEdges(edges []*Edge) int {
	n := 0
	for _, e := range edges {
		if e.Condition != nil {
			n++
		}
	}
	return n
}

// countEqualityEdges returns the number of edges with simple equality conditions.
func countEqualityEdges(edges []*Edge) int {
	n := 0
	for _, e := range edges {
		if _, ok := ExtractEqualityCondition(e); ok {
			n++
		}
	}
	return n
}

// collectConditionValues groups equality condition values by variable name.
func collectConditionValues(edges []*Edge) map[string]map[string]bool {
	byVar := make(map[string]map[string]bool)
	for _, e := range edges {
		cmp, ok := ExtractEqualityCondition(e)
		if !ok {
			continue
		}
		if byVar[cmp.Variable] == nil {
			byVar[cmp.Variable] = make(map[string]bool)
		}
		byVar[cmp.Variable][cmp.Value] = true
	}
	return byVar
}

// edgeHasParsedCondition returns true if the edge has a parsed condition.
func edgeHasParsedCondition(e *Edge) bool {
	return e.Condition != nil && e.Condition.Parsed != nil
}

// isEqualityOp returns true for "=" and "==" operators.
func isEqualityOp(op string) bool {
	return op == "=" || op == "=="
}

// matchesExhaustiveSet returns true if any variable's values cover a known exhaustive set.
func matchesExhaustiveSet(byVar map[string]map[string]bool) bool {
	for variable, values := range byVar {
		if variableIsExhaustive(variable, values) {
			return true
		}
	}
	return false
}

// variableIsExhaustive returns true if the given values for a variable
// cover at least one known exhaustive set.
func variableIsExhaustive(variable string, values map[string]bool) bool {
	sets, known := knownExhaustiveSets[variable]
	if !known {
		return false
	}
	for _, set := range sets {
		if coversAll(values, set) {
			return true
		}
	}
	return false
}

// hasComplementaryPair returns true if any two edges form a complementary pair:
// one asserts a condition and another negates the same condition.
func hasComplementaryPair(edges []*Edge) bool {
	positives, negatives := classifyConditions(edges)
	for key := range positives {
		if negatives[key] {
			return true
		}
	}
	return false
}

// classifyConditions splits edge conditions into positive keys and negated keys.
func classifyConditions(edges []*Edge) (pos, neg map[string]bool) {
	pos = make(map[string]bool)
	neg = make(map[string]bool)
	for _, e := range edges {
		if !edgeHasParsedCondition(e) {
			continue
		}
		if key, ok := conditionKey(e.Condition.Parsed); ok {
			pos[key] = true
		}
		if key, ok := negatedConditionKey(e.Condition.Parsed); ok {
			neg[key] = true
		}
	}
	return pos, neg
}

// conditionKey returns a stable key for a simple comparison expression.
func conditionKey(expr ConditionExpr) (string, bool) {
	cmp, ok := expr.(CondCompare)
	if !ok {
		return "", false
	}
	return cmp.Variable + "|" + cmp.Op + "|" + cmp.Value, true
}

// negatedConditionKey returns the key for the inner comparison of a CondNot.
func negatedConditionKey(expr ConditionExpr) (string, bool) {
	neg, ok := expr.(CondNot)
	if !ok {
		return "", false
	}
	return conditionKey(neg.Inner)
}

// coversAll returns true if values contains every element in required.
func coversAll(values map[string]bool, required []string) bool {
	for _, r := range required {
		if !values[r] {
			return false
		}
	}
	return true
}
