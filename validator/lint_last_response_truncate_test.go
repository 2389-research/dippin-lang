package validator

import "testing"

func TestLint_DIP148_FiresOnNegativeAgentValue(t *testing.T) {
	src := `workflow X
  start: A
  exit: A

  agent A
    prompt: "x"
    last_response_truncate: -1
`
	if !hasCode(lintSrc(t, src), DIP148) {
		t.Errorf("expected DIP148, got: %v", codes(lintSrc(t, src)))
	}
}

func TestLint_DIP148_FiresOnNegativeBranchValue(t *testing.T) {
	src := `workflow X
  start: P
  exit: W

  agent W
    prompt: "w"

  parallel P
    branch: W
      last_response_truncate: -5
`
	if !hasCode(lintSrc(t, src), DIP148) {
		t.Errorf("expected DIP148 for negative branch value, got: %v", codes(lintSrc(t, src)))
	}
}

func TestLint_DIP148_SilentOnZeroAndPositive(t *testing.T) {
	for _, v := range []string{"0", "4096"} {
		src := `workflow X
  start: A
  exit: A

  agent A
    prompt: "x"
    last_response_truncate: ` + v + "\n"
		if hasCode(lintSrc(t, src), DIP148) {
			t.Errorf("DIP148 should not fire for value %q", v)
		}
	}
}
