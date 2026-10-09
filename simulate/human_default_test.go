package simulate

import "testing"

// gateDefaultSrc builds a workflow whose Review gate offers "approve" (first
// edge) and "hold" (second edge). defaultField is spliced into the gate body,
// so callers choose whether the gate declares a default: and which one.
func gateDefaultSrc(mode, defaultField string) string {
	return `workflow GateDefault
  goal: "Gate default"
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

func runGateDefault(t *testing.T, src string, scenario map[string]string) []string {
	t.Helper()
	res, err := Run(mustParseElseWorkflow(t, src), Options{Scenario: scenario})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	return res.Path
}

// An unattended run takes the gate's default:, as tracker's --auto-approve
// does, not the first declared edge (DB-001).
func TestGateDefault_UnattendedRunTakesDefault(t *testing.T) {
	path := runGateDefault(t, gateDefaultSrc("choice", `    default: "hold"`), nil)
	assertPathContains(t, path, "Hold")
	assertPathNotContains(t, path, "Ship")
}

// Tracker's auto-approve also answers a freeform gate with its default:.
func TestGateDefault_FreeformGateTakesDefault(t *testing.T) {
	path := runGateDefault(t, gateDefaultSrc("freeform", `    default: "hold"`), nil)
	assertPathContains(t, path, "Hold")
	assertPathNotContains(t, path, "Ship")
}

func TestGateDefault_PreferredLabelOverridesDefault(t *testing.T) {
	scenario := map[string]string{"Review.preferred_label": "approve"}
	path := runGateDefault(t, gateDefaultSrc("choice", `    default: "hold"`), scenario)
	assertPathContains(t, path, "Ship")
	assertPathNotContains(t, path, "Hold")
}

func TestGateDefault_NoDefaultTakesFirstEdge(t *testing.T) {
	path := runGateDefault(t, gateDefaultSrc("choice", ""), nil)
	assertPathContains(t, path, "Ship")
	assertPathNotContains(t, path, "Hold")
}

func TestGateDefault_UnmatchedDefaultTakesFirstEdge(t *testing.T) {
	path := runGateDefault(t, gateDefaultSrc("choice", `    default: "defer"`), nil)
	assertPathContains(t, path, "Ship")
	assertPathNotContains(t, path, "Hold")
}
