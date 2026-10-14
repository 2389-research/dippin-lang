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

// visitGate resolves PlanReview's next edge as the simulator would on its
// visit-th visit, answering with answer (as a typed --interactive answer or
// a per-visit preferred_label does). s carries visits across calls.
func visitGate(t *testing.T, s *simulator, visit int, answer string) string {
	t.Helper()
	s.nodeVisits["PlanReview"] = visit
	s.ctx["preferred_label"] = answer
	node := s.workflow.Node("PlanReview")
	e := s.resolveGate(node, s.workflow.EdgesFrom("PlanReview"))
	if e == nil {
		t.Fatalf("visit %d answer %q: no edge", visit, answer)
	}
	return e.To
}

func newGateLoopSim(t *testing.T, defaultField string) *simulator {
	t.Helper()
	return &simulator{
		workflow:   mustParseElseWorkflow(t, planLoopSrc(defaultField)),
		opts:       Options{MaxNodeVisits: 3},
		ctx:        map[string]string{},
		visited:    map[string]bool{},
		nodeVisits: map[string]int{},
	}
}

// Past the visit bound, only an answer that takes the `loop` edge is
// overridden. "revise" three times then "approve" must leave through
// approve — even when default: names the loop edge (which used to send the
// run back into the loop until the 500-step limit) or another exit (which
// used to silently reroute approve to Hold).
func TestGateLoop_BoundKeepsAnAnswerThatLeaves(t *testing.T) {
	for _, def := range []string{`    default: "revise"`, `    default: "hold"`, ""} {
		s := newGateLoopSim(t, def)
		for visit := 1; visit <= 3; visit++ {
			if to := visitGate(t, s, visit, "revise"); to != "Plan" {
				t.Fatalf("%q: visit %d revise -> %s, want Plan", def, visit, to)
			}
		}
		if to := visitGate(t, s, 4, "approve"); to != "Done" {
			t.Errorf("%q: past the bound, approve -> %s, want Done", def, to)
		}
	}
}

// Past the bound, a loop answer leaves through default: unless default:
// is itself the loop edge, then through the first edge without `loop`.
func TestGateLoop_BoundOverridesOnlyTheLoopEdge(t *testing.T) {
	cases := map[string]string{`    default: "hold"`: "Hold", `    default: "revise"`: "Done", "": "Done"}
	for def, want := range cases {
		s := newGateLoopSim(t, def)
		if to := visitGate(t, s, 4, "revise"); to != want {
			t.Errorf("%q: past the bound, revise -> %s, want %s", def, to, want)
		}
	}
}
