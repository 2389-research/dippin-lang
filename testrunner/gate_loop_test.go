package testrunner_test

import (
	"testing"

	"github.com/2389-research/dippin-lang/parser"
	"github.com/2389-research/dippin-lang/testrunner"
)

const gateLoopWorkflow = `workflow PlanLoop
  goal: "Gate-driven revise loop"
  start: Plan
  exit: Done

  agent Plan
    prompt:
      Plan.

  human PlanReview
    label: "Approve the plan?"
    mode: choice
    default: "hold"

  agent Hold
    prompt:
      Hold.

  agent Done
    prompt:
      Done.

  edges
    Plan -> PlanReview
    PlanReview -> Done  label: "approve"
    PlanReview -> Hold  label: "hold"
    PlanReview -> Plan  label: "revise"  loop
    Hold -> Done
`

// A scenario that answers "revise" at every visit no longer spins to the
// 500-step limit: the runner's visit bound sends the gate out through its
// default: (DB-006).
func TestRunCase_GateLoopIsBounded(t *testing.T) {
	w, err := parser.NewParser(gateLoopWorkflow, "loop.dip").Parse()
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	tc := testrunner.TestCase{
		Name:     "revise on every visit exits through the default",
		Scenario: map[string]string{"PlanReview.preferred_label": "revise"},
		Expect: testrunner.Expectation{
			Status:  "success",
			Visited: []string{"Hold", "Done"},
		},
	}
	if cr := testrunner.RunCase(w, tc); !cr.Passed {
		t.Fatalf("case failed: %v (path %v)", cr.Errors, cr.Path)
	}
}
