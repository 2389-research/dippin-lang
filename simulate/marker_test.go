package simulate

import (
	"testing"

	"github.com/2389-research/dippin-lang/parser"
)

// TestRun_ToolMarkerDerivedFromStdout guards the simulate fidelity fix: a tool
// declaring marker_grep has ctx.tool_marker derived from ctx.tool_stdout, so an
// `on <marker>` edge routes correctly in a scenario that supplies realistic tool
// output — rather than silently falling through to the first edge (which caused
// spurious "infinite loop" failures on marker-routed workflows).
func TestRun_ToolMarkerDerivedFromStdout(t *testing.T) {
	ResetRunCounter()
	src := `workflow W
  start: Pick
  exit: Done

  tool Pick
    marker_grep: "^(go_impl|all_done)$"
    outputs: go_impl, all_done
    command: echo placeholder

  agent Impl
    prompt: "implement"

  agent Done
    prompt: "done"

  edges
    Pick -> Impl  on go_impl
    Pick -> Done  on all_done
    Impl -> Done
`
	w, err := parser.NewParser(src, "t.dip").Parse()
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	_ = EnsureConditionsParsed(w)

	// tool_stdout "all_done" must derive ctx.tool_marker=all_done and route to the
	// all_done edge (→ Done), skipping Impl — NOT fall through to the first edge.
	res, err := Run(w, Options{Scenario: map[string]string{"Pick.tool_stdout": "all_done"}})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	assertPathContains(t, res.Path, "Done")
	assertPathNotContains(t, res.Path, "Impl")

	// The other marker still routes to Impl.
	res2, err := Run(w, Options{Scenario: map[string]string{"Pick.tool_stdout": "go_impl"}})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	assertPathContains(t, res2.Path, "Impl")

	// An explicit tool_marker scenario value still wins over derivation.
	res3, err := Run(w, Options{Scenario: map[string]string{"Pick.tool_stdout": "go_impl", "Pick.tool_marker": "all_done"}})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	assertPathNotContains(t, res3.Path, "Impl")
}
