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
