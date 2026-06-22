package testrunner

import (
	"fmt"

	"github.com/2389-research/dippin-lang/ir"
	"github.com/2389-research/dippin-lang/simulate"
)

// defaultMaxNodeVisits is the per-node visit limit for test scenarios.
// When a node is visited more times than this, the simulator forces the
// loop-exit edge. This prevents tool-gated loops from spinning to the
// global maxSteps limit while still allowing reasonable loop iterations.
const defaultMaxNodeVisits = 3

// RunSuite runs all test cases in a suite against the given workflow.
func RunSuite(w *ir.Workflow, suite *TestSuite) *SuiteResult {
	sr := &SuiteResult{Total: len(suite.Tests)}
	for _, tc := range suite.Tests {
		cr := RunCase(w, tc)
		sr.Results = append(sr.Results, cr)
		if cr.Passed {
			sr.Passed++
		} else {
			sr.Failed++
		}
	}
	return sr
}

// RunCase runs a single test case and returns the result.
func RunCase(w *ir.Workflow, tc TestCase) CaseResult {
	cr := CaseResult{Name: tc.Name}

	result, err := simulate.Run(w, simulate.Options{
		Scenario:      tc.Scenario,
		MaxNodeVisits: defaultMaxNodeVisits,
		Branch:        toSet(tc.Branch),
	})
	if err != nil {
		cr.Errors = append(cr.Errors, fmt.Sprintf("simulation error: %v", err))
		return cr
	}

	cr.Path = result.Path
	cr.Events = result.Events
	cr.Errors = checkExpectations(result, tc.Expect)
	cr.Passed = len(cr.Errors) == 0
	return cr
}

// checkExpectations validates all assertions against a simulation result.
func checkExpectations(result *simulate.Result, expect Expectation) []string {
	var errs []string
	errs = append(errs, checkStatus(result, expect)...)
	errs = append(errs, checkVisited(result, expect)...)
	errs = append(errs, checkNotVisited(result, expect)...)
	errs = append(errs, checkPathContains(result, expect)...)
	errs = append(errs, checkImmediatelyAfter(result, expect)...)
	return errs
}

// checkStatus verifies the simulation status matches expectations.
func checkStatus(result *simulate.Result, expect Expectation) []string {
	if expect.Status != "" && result.Status != expect.Status {
		return []string{fmt.Sprintf("expected status %q, got %q", expect.Status, result.Status)}
	}
	return nil
}

// checkVisited verifies that all expected nodes were visited.
func checkVisited(result *simulate.Result, expect Expectation) []string {
	if len(expect.Visited) == 0 {
		return nil
	}
	visited := toSet(result.Path)
	var errs []string
	for _, node := range expect.Visited {
		if !visited[node] {
			errs = append(errs, fmt.Sprintf("expected node %q to be visited", node))
		}
	}
	return errs
}

// checkNotVisited verifies that excluded nodes were not visited.
func checkNotVisited(result *simulate.Result, expect Expectation) []string {
	if len(expect.NotVisited) == 0 {
		return nil
	}
	visited := toSet(result.Path)
	var errs []string
	for _, node := range expect.NotVisited {
		if visited[node] {
			errs = append(errs, fmt.Sprintf("expected node %q to NOT be visited", node))
		}
	}
	return errs
}

// checkPathContains verifies that expected nodes appear in order in the path.
func checkPathContains(result *simulate.Result, expect Expectation) []string {
	if len(expect.PathContains) == 0 {
		return nil
	}
	var errs []string
	pathIdx := 0
	for _, expected := range expect.PathContains {
		idx := findInPath(result.Path, expected, pathIdx)
		if idx < 0 {
			errs = append(errs, fmt.Sprintf("expected %q in path order, not found after previous matches", expected))
		} else {
			pathIdx = idx + 1
		}
	}
	return errs
}

// checkImmediatelyAfter verifies that for each (from, to) pair,
// the node immediately following from in the path is to.
func checkImmediatelyAfter(result *simulate.Result, expect Expectation) []string {
	if len(expect.ImmediatelyAfter) == 0 {
		return nil
	}
	var errs []string
	for from, to := range expect.ImmediatelyAfter {
		if err := checkAdjacent(result.Path, from, to); err != "" {
			errs = append(errs, err)
		}
	}
	return errs
}

// checkAdjacent returns an error string if from is not immediately followed by to in path.
func checkAdjacent(path []string, from, to string) string {
	idx := findInPath(path, from, 0)
	if idx < 0 {
		return fmt.Sprintf("immediately_after: node %q not found in path", from)
	}
	if idx == len(path)-1 {
		return fmt.Sprintf("immediately_after: node %q is the last element in path, no next node", from)
	}
	if path[idx+1] != to {
		return fmt.Sprintf("immediately_after: expected %q after %q, got %q", to, from, path[idx+1])
	}
	return ""
}

// findInPath returns the index of target in path starting from offset, or -1.
func findInPath(path []string, target string, offset int) int {
	for i := offset; i < len(path); i++ {
		if path[i] == target {
			return i
		}
	}
	return -1
}

func toSet(items []string) map[string]bool {
	s := make(map[string]bool, len(items))
	for _, item := range items {
		s[item] = true
	}
	return s
}
