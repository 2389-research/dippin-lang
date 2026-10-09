package simulate

import "testing"

// planLoopSrc builds a workflow whose PlanReview gate offers "approve" (to
// Done), "hold" (to Hold) and "revise" (back to Plan). defaultField is spliced
// into the gate body.
func planLoopSrc(defaultField string) string {
	return `workflow PlanLoop
  goal: "Gate-driven revise loop"
  start: Plan
  exit: Done

  agent Plan
    prompt:
      Plan.

  human PlanReview
    label: "Approve the plan?"
    mode: choice
` + defaultField + `
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
}

// reviseEveryVisit is re-applied on each visit to PlanReview, so without a
// loop bound the gate sends the run back to Plan forever.
var reviseEveryVisit = map[string]string{"PlanReview.preferred_label": "revise"}

func runGateLoop(t *testing.T, src string, scenario map[string]string) *Result {
	t.Helper()
	res, err := Run(mustParseElseWorkflow(t, src), Options{Scenario: scenario, MaxNodeVisits: 3})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if res.Status != "success" {
		t.Errorf("Status = %q, want success (path %v)", res.Status, res.Path)
	}
	return res
}

func countVisits(path []string, nodeID string) int {
	n := 0
	for _, id := range path {
		if id == nodeID {
			n++
		}
	}
	return n
}

// MaxNodeVisits bounds a gate routed by label: past the limit the gate leaves
// through its default: edge (DB-006).
func TestGateLoop_BoundExitsThroughDefault(t *testing.T) {
	res := runGateLoop(t, planLoopSrc(`    default: "hold"`), reviseEveryVisit)
	assertPathContains(t, res.Path, "Hold")
	if got := countVisits(res.Path, "PlanReview"); got != 4 {
		t.Errorf("PlanReview visited %d times, want 4 (MaxNodeVisits+1); path %v", got, res.Path)
	}
}

// With no default:, the bound leaves through the first edge that is not the
// loop.
func TestGateLoop_BoundExitsThroughFirstOtherEdge(t *testing.T) {
	res := runGateLoop(t, planLoopSrc(""), reviseEveryVisit)
	assertPathContains(t, res.Path, "Done")
	assertPathNotContains(t, res.Path, "Hold")
}

// A default: that names the loop edge keeps an unattended run looping; the
// bound leaves through the first other edge instead.
func TestGateLoop_DefaultThatLoopsStillExits(t *testing.T) {
	res := runGateLoop(t, planLoopSrc(`    default: "revise"`), nil)
	assertPathContains(t, res.Path, "Done")
	assertPathNotContains(t, res.Path, "Hold")
}

// A gate that loops on a condition rather than a label keeps the ordinary
// loop exit: the first conditional edge whose guard does not match.
func TestGateLoop_ConditionalGateKeepsConditionalExit(t *testing.T) {
	src := `workflow ConditionalGate
  goal: "Gate that loops on outcome"
  start: Plan
  exit: Done

  agent Plan
    prompt:
      Plan.

  human Approve
    label: "Approve?"
    mode: yes_no

  agent Done
    prompt:
      Done.

  edges
    Plan -> Approve
    Approve -> Plan  when ctx.outcome = fail  loop
    Approve -> Done  when ctx.outcome = success
`
	res := runGateLoop(t, src, map[string]string{"outcome": "fail"})
	assertPathContains(t, res.Path, "Done")
}
