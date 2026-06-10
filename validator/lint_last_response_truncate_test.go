package validator

import (
	"strings"
	"testing"
)

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

func TestLint_DIP148_CarriesSourceLocation(t *testing.T) {
	src := `workflow X
  start: A
  exit: A

  agent A
    prompt: "x"
    last_response_truncate: -1
`
	diags := lintSrc(t, src)
	var found bool
	for _, d := range diags {
		if d.Code == DIP148 {
			found = true
			if d.Location.Line <= 0 {
				t.Errorf("DIP148 diagnostic must carry a real source location, got Line=%d", d.Location.Line)
			}
			break
		}
	}
	if !found {
		t.Fatalf("expected DIP148 diagnostic, got: %v", codes(diags))
	}
}

func TestLint_DIP148_MessageNamesNodeAndValue(t *testing.T) {
	src := `workflow X
  start: A
  exit: A

  agent A
    prompt: "x"
    last_response_truncate: -1
`
	diags := lintSrc(t, src)
	var msg string
	for _, d := range diags {
		if d.Code == DIP148 {
			msg = d.Message
			break
		}
	}
	if msg == "" {
		t.Fatalf("expected DIP148 diagnostic, got: %v", codes(diags))
	}
	if !strings.Contains(msg, "A") || !strings.Contains(msg, "-1") {
		t.Errorf("message must name node and value, got: %q", msg)
	}
}
