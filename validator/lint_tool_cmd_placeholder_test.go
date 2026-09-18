package validator_test

import (
	"testing"
)

// TestLintToolBinary_PlaceholderRegression is a real-parser regression test
// for issue #305: a tool command body containing a ${ns.key} placeholder
// must not derail DIP125's binary extraction to a shell flag or a whole
// assignment line. This is the issue's min.dip repro, parsed through the
// real parser (not hand-built IR) and linted end-to-end.
func TestLintToolBinary_PlaceholderRegression(t *testing.T) {
	src := `workflow probe
  start: Plain
  exit: WithVar

  tool Plain
    timeout: 10s
    command:
      set -eu
      echo plain

  tool WithVar
    timeout: 10s
    command:
      set -eu
      X="${graph.workflow_dir}/lib"
      echo x

  tool ParamsVar
    timeout: 10s
    command:
      set -eu
      X="${params.foo}"
      echo x

  tool NoSet
    timeout: 10s
    command:
      LIB="${graph.workflow_dir}/lib"
      echo x

  edges
    Plain -> WithVar
    Plain -> ParamsVar
    Plain -> NoSet
`
	diags := lintSrc(t, src)
	if hasCode(diags, "DIP125") {
		t.Fatalf("expected no DIP125 for placeholder-bearing tool commands, got %v", diags)
	}
}
