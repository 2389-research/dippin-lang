package validator_test

import (
	"testing"

	"github.com/2389-research/dippin-lang/validator"
)

// TestDIP125_ColonBuiltinAndSourcedFunc covers issue #315's repro: the ":"
// special builtin must not be mistaken for a missing binary, and a function
// defined by a sourced file must not fire DIP125 either.
func TestDIP125_ColonBuiltinAndSourcedFunc(t *testing.T) {
	src := `workflow Repro
  goal: "dip125 repro"
  start: A
  exit: B

  tool A
    command:
      set -eu
      : > out.log
      . ./lib.sh
      my_func
  tool B
    command: "true"

  edges
    A -> B
`
	diags := lintSrc(t, src)
	if hasCode(diags, validator.DIP125) {
		t.Errorf("expected no DIP125, got: %v", diags)
	}
}

// TestDIP125_SourcedFuncNoColon is the repro with the ":" line removed —
// the "." command still precedes my_func, so DIP125 must still be skipped.
func TestDIP125_SourcedFuncNoColon(t *testing.T) {
	src := `workflow Repro
  goal: "dip125 repro"
  start: A
  exit: B

  tool A
    command:
      set -eu
      . ./lib.sh
      my_func
  tool B
    command: "true"

  edges
    A -> B
`
	diags := lintSrc(t, src)
	if hasCode(diags, validator.DIP125) {
		t.Errorf("expected no DIP125 (source precedes my_func), got: %v", diags)
	}
}
