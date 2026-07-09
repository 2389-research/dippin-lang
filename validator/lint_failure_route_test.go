package validator

import (
	"strings"
	"testing"

	"github.com/2389-research/dippin-lang/parser"
)

func dip144Fires(diags []Diagnostic, nodeID string) bool {
	for _, d := range diags {
		if d.Code == DIP144 && strings.Contains(d.Message, "\""+nodeID+"\"") {
			return true
		}
	}
	return false
}

func TestDIP144FiresOnRoutelessAgent(t *testing.T) {
	src := `workflow W
  start: A
  exit: Done
  agent A
    prompt: "go"
  agent Done
    prompt: "end"
  edges
    A -> Done
`
	if !hasCode(lintSrc(t, src), DIP144) {
		t.Fatal("expected DIP144 on routeless agent A")
	}
}

func TestDIP144SuppressedByFailEdge(t *testing.T) {
	src := `workflow W
  start: A
  exit: Done
  agent A
    prompt: "go"
  human Rescue
    mode: freeform
  agent Done
    prompt: "end"
  edges
    A -> Done when ctx.outcome = success
    A -> Rescue when ctx.outcome = fail
    Rescue -> Done
`
	if dip144Fires(lintSrc(t, src), "A") {
		t.Fatal("DIP144 should be suppressed by fail edge on A")
	}
}

func TestDIP144SuppressedByFallbackTarget(t *testing.T) {
	src := `workflow W
  start: A
  exit: Done
  agent A
    prompt: "go"
    fallback_target: Done
  agent Done
    prompt: "end"
  edges
    A -> Done
`
	if dip144Fires(lintSrc(t, src), "A") {
		t.Fatal("DIP144 should be suppressed by fallback_target")
	}
}

func TestDIP144SuppressedByBoundedRetry(t *testing.T) {
	src := `workflow W
  start: A
  exit: Done
  agent A
    prompt: "go"
    retry_target: A
    max_retries: 3
  agent Done
    prompt: "end"
  edges
    A -> Done
`
	if dip144Fires(lintSrc(t, src), "A") {
		t.Fatal("DIP144 should be suppressed by bounded retry_target + max_retries")
	}
}

func TestDIP144SuppressedByFailureEdgeSpelling(t *testing.T) {
	src := `workflow W
  start: A
  exit: Done
  agent A
    prompt: "go"
  human Rescue
    mode: freeform
  agent Done
    prompt: "end"
  edges
    A -> Done when ctx.outcome = success
    A -> Rescue when ctx.outcome = failure
    Rescue -> Done
`
	if dip144Fires(lintSrc(t, src), "A") {
		t.Fatal("DIP144 should be suppressed by a ctx.outcome = failure edge (alternate spelling)")
	}
}

func TestDIP144SkipsExitAndTerminalNodes(t *testing.T) {
	// Mid is a routeless agent (DIP144 fires — sanity check the lint is active);
	// Done is the exit node (skipped); Leaf is a terminal agent with no outgoing
	// edges (skipped). This pins the exit-node and zero-outgoing guards.
	src := `workflow W
  start: Mid
  exit: Done
  agent Mid
    prompt: "go"
  agent Leaf
    prompt: "leaf"
  agent Done
    prompt: "end"
  edges
    Mid -> Done
    Mid -> Leaf
`
	diags := lintSrc(t, src)
	if !dip144Fires(diags, "Mid") {
		t.Fatal("DIP144 should fire on routeless agent Mid (sanity check)")
	}
	if dip144Fires(diags, "Done") {
		t.Fatal("DIP144 must not fire on the exit node")
	}
	if dip144Fires(diags, "Leaf") {
		t.Fatal("DIP144 must not fire on a terminal node with no outgoing edges")
	}
}

func TestDIP144SuppressedByGraphOnFailure(t *testing.T) {
	src := `workflow W
  start: A
  exit: Done
  defaults
    on_failure: Rescue
  agent A
    prompt: "go"
  human Rescue
    mode: freeform
  agent Done
    prompt: "end"
  edges
    A -> Done
    Rescue -> Done
`
	if dip144Fires(lintSrc(t, src), "A") {
		t.Fatal("DIP144 should be suppressed by graph on_failure")
	}
}

func TestDIP144FiresWhenOnFailureTargetMissing(t *testing.T) {
	src := `workflow W
  start: A
  exit: Done
  defaults
    on_failure: Ghost
  agent A
    prompt: "go"
  agent Done
    prompt: "end"
  edges
    A -> Done
`
	if !hasCode(lintSrc(t, src), DIP144) {
		t.Fatal("DIP144 should fire when on_failure points to a nonexistent node (not a real route)")
	}
}

func TestDIP144FiresWithDIP115OnRoutelessGoalGate(t *testing.T) {
	src := `workflow W
  start: G
  exit: Done
  agent G
    prompt: "gate"
    goal_gate: true
  agent Done
    prompt: "end"
  edges
    G -> Done
`
	diags := lintSrc(t, src)
	if !hasCode(diags, DIP144) {
		t.Fatal("expected DIP144 on routeless goal_gate agent")
	}
	if !hasCode(diags, DIP115) {
		t.Fatal("expected DIP115 to still fire alongside DIP144 (independent codes)")
	}
}

func TestDIP144NotOnHumanOrToolNodes(t *testing.T) {
	src := `workflow W
  start: H
  exit: Done
  human H
    mode: freeform
  tool T
    command: "echo hi"
  agent Done
    prompt: "end"
    fallback_target: Done
  edges
    H -> T
    T -> Done
`
	diags := lintSrc(t, src)
	if dip144Fires(diags, "H") {
		t.Fatal("DIP144 must not fire on human node H")
	}
	if dip144Fires(diags, "T") {
		t.Fatal("DIP144 must not fire on tool node T")
	}
}

func diagsForCode(t *testing.T, src, code string) int {
	t.Helper()
	w, err := parser.NewParser(src, "t.dip").Parse()
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	n := 0
	for _, d := range Lint(w).Diagnostics {
		if d.Code == code {
			n++
		}
	}
	return n
}

const goalGateV1 = `workflow W
  goal: "t"
  start: G
  exit: D
  agent G
    prompt: g
    goal_gate: true
    fallback_target: D
  agent D
    prompt: d
  edges
    G -> D
`

const goalGateV2Edge = `workflow W
  goal: "t"
  start: G
  exit: D
  agent G
    prompt: g
    goal_gate: true
  agent D
    prompt: d
  edges
    G -> D  on fail
`

func TestDIP115_SatisfiedByFallbackTargetAndByFailEdge(t *testing.T) {
	if got := diagsForCode(t, goalGateV1, "DIP115"); got != 0 {
		t.Errorf("v1 fallback_target should satisfy DIP115, got %d", got)
	}
	if got := diagsForCode(t, goalGateV2Edge, "DIP115"); got != 0 {
		t.Errorf("v2 on-fail edge should satisfy DIP115, got %d", got)
	}
}
