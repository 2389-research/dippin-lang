package testrunner_test

import (
	"testing"

	"github.com/2389-research/dippin-lang/parser"
	"github.com/2389-research/dippin-lang/testrunner"
)

const gateDefaultWorkflow = `workflow GateDefault
  goal: "Route an unattended gate by its default"
  start: Start
  exit: Done

  agent Start
    prompt:
      Begin.

  human Review
    label: "Ship it?"
    mode: choice
    default: "hold"

  agent Ship
    prompt:
      Ship.

  agent Hold
    prompt:
      Hold.

  agent Done
    prompt:
      Done.

  edges
    Start -> Review
    Review -> Ship  label: "approve"
    Review -> Hold  label: "hold"
    Ship -> Done
    Hold -> Done
`

// A scenario test with no preferred_label routes the gate by its default:,
// so a suite can assert the path an unattended run takes (DB-001).
func TestRunCase_GateDefault(t *testing.T) {
	w, err := parser.NewParser(gateDefaultWorkflow, "gate.dip").Parse()
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	tc := testrunner.TestCase{
		Name: "unattended run takes the gate default",
		Expect: testrunner.Expectation{
			Status:     "success",
			Visited:    []string{"Hold"},
			NotVisited: []string{"Ship"},
		},
	}
	if cr := testrunner.RunCase(w, tc); !cr.Passed {
		t.Fatalf("case failed: %v (path %v)", cr.Errors, cr.Path)
	}
}
