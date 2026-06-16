package validator

import "testing"

// DIP150 fires when a human gate's outgoing edge routes by label: with no choice:.
func TestLint_DIP150_FiresOnHumanLabelRoute(t *testing.T) {
	src := `workflow X
  start: Begin
  exit: Done

  human Begin
    mode: choice

  agent Done
    prompt: "done"

  edges
    Begin -> Done  label: "yes"
`
	if !hasCode(lintSrc(t, src), DIP150) {
		t.Errorf("expected DIP150, got: %v", codes(lintSrc(t, src)))
	}
}

// DIP150 is silent once an explicit choice: routing key is present.
func TestLint_DIP150_SilentWhenChoicePresent(t *testing.T) {
	src := `workflow X
  start: Begin
  exit: Done

  human Begin
    mode: choice

  agent Done
    prompt: "done"

  edges
    Begin -> Done  label: "yes"  choice: "yes"
`
	if hasCode(lintSrc(t, src), DIP150) {
		t.Errorf("expected no DIP150 when choice: present, got: %v", codes(lintSrc(t, src)))
	}
}

// DIP150 does not fire for freeform/interview human modes, whose edge labels are
// display text rather than a routing key.
func TestLint_DIP150_SilentForNonRoutingModes(t *testing.T) {
	for _, mode := range []string{"freeform", "interview"} {
		src := `workflow X
  start: Begin
  exit: Done

  human Begin
    mode: ` + mode + `

  agent Done
    prompt: "done"

  edges
    Begin -> Done  label: "yes"
`
		if hasCode(lintSrc(t, src), DIP150) {
			t.Errorf("mode %q: expected no DIP150, got: %v", mode, codes(lintSrc(t, src)))
		}
	}
}

// DIP150 still fires for a mode-less human gate — the canonical multi-choice gate
// routes by its edge labels.
func TestLint_DIP150_FiresOnModelessHumanGate(t *testing.T) {
	src := `workflow X
  start: Begin
  exit: Done

  human Begin
    prompt: "pick one"

  agent Done
    prompt: "done"

  edges
    Begin -> Done  label: "yes"
`
	if !hasCode(lintSrc(t, src), DIP150) {
		t.Errorf("expected DIP150 on mode-less human gate, got: %v", codes(lintSrc(t, src)))
	}
}

// DIP150 does not fire for a labeled edge whose source is a non-human node.
func TestLint_DIP150_SilentForNonHumanSource(t *testing.T) {
	src := `workflow X
  start: Route
  exit: Done

  agent Route
    prompt: "route"

  agent Done
    prompt: "done"

  edges
    Route -> Done  label: "yes"
`
	if hasCode(lintSrc(t, src), DIP150) {
		t.Errorf("expected no DIP150 for non-human source, got: %v", codes(lintSrc(t, src)))
	}
}
