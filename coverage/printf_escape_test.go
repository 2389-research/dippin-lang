package coverage

import (
	"testing"

	"github.com/2389-research/dippin-lang/parser"
)

const printfNewlineDip = `workflow PrintfNewline
  goal: "Route on printf markers that end in a newline"
  start: RunTool
  exit: Done

  tool RunTool
    command:
      if run-checks; then
        printf 'pass\n'
      else
        printf 'fail\n'
      fi

  agent Fixup
    prompt:
      Fix.

  agent Done
    prompt:
      Done.

  edges
    RunTool -> Done   when ctx.tool_stdout = pass
    RunTool -> Fixup  when ctx.tool_stdout = fail
    Fixup -> Done
`

// Edges that test the bare markers cover a tool whose printf markers end in
// \n, as they do at runtime, where tool stdout is trimmed (DB-003).
func TestAnalyze_PrintfNewlineMarkersCovered(t *testing.T) {
	w, err := parser.NewParser(printfNewlineDip, "t.dip").Parse()
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	cov := Analyze(w).Nodes["RunTool"]
	if cov.Status != "covered" || len(cov.MissingEdges) != 0 {
		t.Errorf("RunTool: status %q, missing %v, extracted %v; want covered with nothing missing",
			cov.Status, cov.MissingEdges, cov.ExtractedOutputs)
	}
}
