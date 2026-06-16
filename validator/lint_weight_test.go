package validator

import "testing"

func TestLint_DIP151_WeightFires(t *testing.T) {
	src := `workflow X
  start: A
  exit: B

  agent A
    prompt: "decide"

  agent B
    prompt: "done"

  edges
    A -> B weight: 5
`
	if !hasCode(lintSrc(t, src), DIP151) {
		t.Errorf("expected DIP151, got: %v", codes(lintSrc(t, src)))
	}
}

func TestLint_DIP151_NoWeightSilent(t *testing.T) {
	src := `workflow X
  start: A
  exit: B

  agent A
    prompt: "decide"

  agent B
    prompt: "done"

  edges
    A -> B
`
	if hasCode(lintSrc(t, src), DIP151) {
		t.Errorf("did not expect DIP151, got: %v", codes(lintSrc(t, src)))
	}
}

func TestLint_DIP151_TwoWeightedEdgesTwoDiagnostics(t *testing.T) {
	src := `workflow X
  start: A
  exit: C

  agent A
    prompt: "decide"

  agent B
    prompt: "b"

  agent C
    prompt: "c"

  edges
    A -> B weight: 10
    A -> C weight: 5
    B -> C
`
	n := 0
	for _, d := range lintSrc(t, src) {
		if d.Code == DIP151 {
			n++
		}
	}
	if n != 2 {
		t.Errorf("expected 2 DIP151 diagnostics, got %d: %v", n, codes(lintSrc(t, src)))
	}
}
