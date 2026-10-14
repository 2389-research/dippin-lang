package simulate

import (
	"bytes"
	"strings"
	"testing"
)

// reviewGateSrc builds a workflow whose Review gate offers "approve" (first
// edge) and "hold" (second edge). defaultField is spliced into the gate body.
func reviewGateSrc(mode, defaultField string) string {
	return `workflow ReviewGate
  goal: "Interactive gate routing"
  start: Start
  exit: Done

  agent Start
    prompt:
      Begin.

  human Review
    label: "Ship it?"
    mode: ` + mode + `
` + defaultField + `
  agent Ship
    prompt:
      Ship.

  agent Hold
    prompt:
      Hold.

  agent Done
    prompt:
      Done.

  edges
    Start -> Review
    Review -> Ship  label: "approve"
    Review -> Hold  label: "hold"
    Ship -> Done
    Hold -> Done
`
}

func runInteractive(t *testing.T, src, input string) []string {
	t.Helper()
	var stderr bytes.Buffer
	res, err := Run(mustParseElseWorkflow(t, src), Options{
		Interactive: true,
		Stdin:       strings.NewReader(input),
		Stderr:      &stderr,
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	return res.Path
}

// The answer typed at an --interactive choice gate picks the edge, as
// tracker routes the answer it collects.
func TestInteractive_TypedChoiceRoutesTheGate(t *testing.T) {
	path := runInteractive(t, reviewGateSrc("choice", ""), "hold\n")
	assertPathContains(t, path, "Hold")
	assertPathNotContains(t, path, "Ship")
}

func TestInteractive_TypedFreeformAnswerRoutesTheGate(t *testing.T) {
	path := runInteractive(t, reviewGateSrc("freeform", ""), "hold\n")
	assertPathContains(t, path, "Hold")
	assertPathNotContains(t, path, "Ship")
}

// At end of input the prompt answers with the gate's default:, and that
// answer routes the gate too.
func TestInteractive_EndOfInputTakesDefault(t *testing.T) {
	path := runInteractive(t, reviewGateSrc("choice", `    default: "hold"`), "")
	assertPathContains(t, path, "Hold")
	assertPathNotContains(t, path, "Ship")
}

// A blank answer with no default leaves the gate on its first edge.
func TestInteractive_BlankAnswerTakesFirstEdge(t *testing.T) {
	path := runInteractive(t, reviewGateSrc("choice", ""), "\n")
	assertPathContains(t, path, "Ship")
	assertPathNotContains(t, path, "Hold")
}

// gateEdgesSrc builds a workflow whose Gate (in mode) has the given edges, in
// order, each leading to its own agent; an edge is "<Target>  <attrs>".
func gateEdgesSrc(mode string, edges ...string) string {
	var agents, lines strings.Builder
	for _, e := range edges {
		target := strings.Fields(e)[0]
		agents.WriteString("  agent " + target + "\n    prompt:\n      Go.\n\n")
		lines.WriteString("    Gate -> " + e + "\n    " + target + " -> Done\n")
	}
	return `workflow Gate
  goal: "Gate answers"
  start: Start
  exit: Done

  agent Start
    prompt:
      Begin.

  human Gate
    label: "Pick"
    mode: ` + mode + `

` + agents.String() + `  agent Done
    prompt:
      Done.

  edges
    Start -> Gate
` + lines.String()
}

// A typed answer is matched exactly (case-insensitive), as tracker's console
// and freeform matching do — never as a substring. "n" names neither
// "Continue" nor "[N] No", so the gate falls back to its first edge.
func TestInteractive_NoSubstringMatch(t *testing.T) {
	for _, mode := range []string{"choice", "freeform"} {
		src := gateEdgesSrc(mode, `Stop  label: "Stop"`, `Cont  label: "Continue"`, `No  label: "[N] No"`)
		path := runInteractive(t, src, "n\n")
		assertPathContains(t, path, "Stop")
		assertPathNotContains(t, path, "Cont")
		assertPathNotContains(t, path, "No")
	}
}

// "a" at a hold/approve/abort gate is no answer at all, not "approve".
func TestInteractive_SingleLetterIsNotAPrefix(t *testing.T) {
	src := gateEdgesSrc("choice", `Hold  label: "hold"`, `Approve  label: "approve"`, `Abort  label: "abort"`)
	path := runInteractive(t, src, "a\n")
	assertPathContains(t, path, "Hold")
	assertPathNotContains(t, path, "Approve")
	assertPathNotContains(t, path, "Abort")
}

// A choice gate also takes an option by its 1-based number, as tracker's
// console does.
func TestInteractive_ChoiceByNumber(t *testing.T) {
	src := gateEdgesSrc("choice", `Hold  label: "hold"`, `Approve  label: "approve"`)
	path := runInteractive(t, src, "2\n")
	assertPathContains(t, path, "Approve")
	assertPathNotContains(t, path, "Hold")
}

// An answer routes to the edge it names, whether it names the display label
// (choice gate) or the choice: key (freeform gate).
func TestInteractive_ChoiceKeyEdges(t *testing.T) {
	edges := []string{`Hold  label: "Put on hold"  choice: "hold"`, `Ship  label: "Ship now"  choice: "approve"`}
	for mode, answer := range map[string]string{"choice": "ship now", "freeform": "approve"} {
		path := runInteractive(t, gateEdgesSrc(mode, edges...), answer+"\n")
		assertPathContains(t, path, "Ship")
		assertPathNotContains(t, path, "Hold")
	}
}

// An answer typed at a human node with a single edge does not route a later
// gate: it is only recorded where the answer is consulted.
func TestInteractive_SingleEdgeAnswerDoesNotLeak(t *testing.T) {
	src := `workflow Leak
  goal: "No leak"
  start: Notes
  exit: Done

  human Notes
    label: "Any notes?"
    mode: freeform

  human Review
    label: "Ship it?"
    mode: choice

  agent Ship
    prompt:
      Ship.

  agent Hold
    prompt:
      Hold.

  agent Done
    prompt:
      Done.

  edges
    Notes -> Review
    Review -> Ship  label: "approve"
    Review -> Hold  label: "hold"
    Ship -> Done
    Hold -> Done
`
	path := runInteractive(t, src, "hold\n\n")
	assertPathContains(t, path, "Ship")
	assertPathNotContains(t, path, "Hold")
}

// Each gate reads its own line: answers typed for successive gates are not
// swallowed by the first prompt's read-ahead.
func TestInteractive_SuccessiveGatesReadSuccessiveLines(t *testing.T) {
	src := `workflow Two
  goal: "Two gates"
  start: First
  exit: Done

  human First
    label: "First?"
    mode: choice

  human Second
    label: "Second?"
    mode: choice

  agent A
    prompt:
      A.

  agent B
    prompt:
      B.

  agent Done
    prompt:
      Done.

  edges
    First -> Second  label: "go"
    First -> Done    label: "stop"
    Second -> A  label: "a"
    Second -> B  label: "b"
    A -> Done
    B -> Done
`
	path := runInteractive(t, src, "go\nb\n")
	assertPathContains(t, path, "B")
	assertPathNotContains(t, path, "A")
}
