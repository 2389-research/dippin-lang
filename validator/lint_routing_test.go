package validator

import "testing"

// TestLint_DIP149_TwoUnconditionalEdges: a node with two outgoing edges that
// both have no condition is ambiguous — only the lexical tiebreak decides which
// fires. DIP149 must warn.
func TestLint_DIP149_TwoUnconditionalEdges(t *testing.T) {
	src := `workflow X
  start: A
  exit: Done

  agent A
    prompt: "x"

  agent B
    prompt: "x"

  agent C
    prompt: "x"

  agent Done
    prompt: "x"

  edges
    A -> B
    A -> C
    B -> Done
    C -> Done
`
	if !hasCode(lintSrc(t, src), DIP149) {
		t.Errorf("expected DIP149 for two unconditional edges, got: %v", codes(lintSrc(t, src)))
	}
}

// TestLint_DIP149_GuardedPlusSingleFallback: a guarded edge plus a single
// unconditional fallback is the intended routing pattern. DIP149 must NOT warn.
func TestLint_DIP149_GuardedPlusSingleFallback(t *testing.T) {
	src := `workflow X
  start: A
  exit: Done

  agent A
    prompt: "x"

  agent B
    prompt: "x"

  agent Done
    prompt: "x"

  edges
    A -> B     when ctx.outcome = success
    A -> Done
    B -> Done
`
	if hasCode(lintSrc(t, src), DIP149) {
		t.Errorf("DIP149 should not fire for a guarded edge plus single fallback, got: %v", codes(lintSrc(t, src)))
	}
}

// TestLint_DIP149_SingleUnconditionalEdge: a node with exactly one outgoing
// edge is never ambiguous. DIP149 must NOT warn.
func TestLint_DIP149_SingleUnconditionalEdge(t *testing.T) {
	src := `workflow X
  start: A
  exit: Done

  agent A
    prompt: "x"

  agent Done
    prompt: "x"

  edges
    A -> Done
`
	if hasCode(lintSrc(t, src), DIP149) {
		t.Errorf("DIP149 should not fire for a single edge, got: %v", codes(lintSrc(t, src)))
	}
}

// TestLint_DIP149_RestartBackEdgeNotCounted: an unconditional forward edge
// alongside a restart back-edge is not a forward-routing ambiguity — the
// restart edge is a distinct re-execution channel. DIP149 must NOT warn.
func TestLint_DIP149_RestartBackEdgeNotCounted(t *testing.T) {
	src := `workflow X
  start: A
  exit: Done

  agent A
    prompt: "x"

  agent Done
    prompt: "x"

  edges
    A -> Done
    A -> A      restart: true
`
	if hasCode(lintSrc(t, src), DIP149) {
		t.Errorf("DIP149 should not fire when the second edge is a restart back-edge, got: %v", codes(lintSrc(t, src)))
	}
}

// TestLint_DIP149_ParallelFanOutNotFlagged: a parallel node's multiple
// unconditional outgoing edges are structural fan-out (all targets run), not a
// tiebreak-resolved choice. ir.EdgesFrom synthesizes them. DIP149 must NOT warn.
func TestLint_DIP149_ParallelFanOutNotFlagged(t *testing.T) {
	src := `workflow X
  start: Fan
  exit: Done

  parallel Fan -> B, C

  agent B
    prompt: "x"

  agent C
    prompt: "x"

  agent Done
    prompt: "x"

  edges
    B -> Done
    C -> Done
`
	if hasCode(lintSrc(t, src), DIP149) {
		t.Errorf("DIP149 should not fire for a parallel fan-out node, got: %v", codes(lintSrc(t, src)))
	}
}

// TestLint_DIP149_HumanChoiceGateNotFlagged: a human node (esp. mode: choice)
// routes on the human's label selection, not the cascade tiebreak, so its
// multiple unconditional outgoing edges are not ambiguous. DIP149 must NOT warn.
func TestLint_DIP149_HumanChoiceGateNotFlagged(t *testing.T) {
	src := `workflow X
  start: Gate
  exit: Done

  human Gate
    mode: choice
    default: approve

  agent B
    prompt: "x"

  agent C
    prompt: "x"

  agent Done
    prompt: "x"

  edges
    Gate -> B
    Gate -> C
    B -> Done
    C -> Done
`
	if hasCode(lintSrc(t, src), DIP149) {
		t.Errorf("DIP149 should not fire for a human choice gate, got: %v", codes(lintSrc(t, src)))
	}
}

// TestLint_DIP149_MessageNamesNode: the diagnostic must name the source node so
// the author can find it, and carry a real source location.
func TestLint_DIP149_MessageNamesNode(t *testing.T) {
	src := `workflow X
  start: A
  exit: Done

  agent A
    prompt: "x"

  agent B
    prompt: "x"

  agent C
    prompt: "x"

  agent Done
    prompt: "x"

  edges
    A -> B
    A -> C
    B -> Done
    C -> Done
`
	var found bool
	for _, d := range lintSrc(t, src) {
		if d.Code == DIP149 {
			found = true
			if d.Location.Line == 0 {
				t.Errorf("DIP149 must carry a real source location, got Line=0")
			}
		}
	}
	if !found {
		t.Fatalf("expected DIP149 diagnostic, got: %v", codes(lintSrc(t, src)))
	}
}
