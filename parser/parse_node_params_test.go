package parser

import (
	"testing"

	"github.com/2389-research/dippin-lang/ir"
)

func TestParseParallelBlockParams(t *testing.T) {
	input := `workflow Test
  start: P
  exit: J

  agent A
    prompt: "A"
  agent B
    prompt: "B"

  parallel P
    branch: A
    branch: B
    params:
      fan_in_policy: all
      quorum: 2

  fan_in J <- A, B

  edges
    J -> J
`
	w, err := NewParser(input, "test.dip").Parse()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	cfg := findNode(t, w, "P").Config.(ir.ParallelConfig)
	if len(cfg.Branches) != 2 {
		t.Fatalf("branches = %d, want 2", len(cfg.Branches))
	}
	if cfg.Params["fan_in_policy"] != "all" {
		t.Errorf("params[fan_in_policy] = %q, want all", cfg.Params["fan_in_policy"])
	}
	if cfg.Params["quorum"] != "2" {
		t.Errorf("params[quorum] = %q, want 2", cfg.Params["quorum"])
	}
}

func TestParseParallelInlineParams(t *testing.T) {
	input := `workflow Test
  start: P
  exit: J

  agent A
    prompt: "A"
  agent B
    prompt: "B"

  parallel P -> A, B
    params:
      fan_in_policy: all

  fan_in J <- A, B

  edges
    J -> J
`
	w, err := NewParser(input, "test.dip").Parse()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	cfg := findNode(t, w, "P").Config.(ir.ParallelConfig)
	if len(cfg.Targets) != 2 {
		t.Errorf("targets = %d, want 2", len(cfg.Targets))
	}
	if len(cfg.Branches) != 0 {
		t.Errorf("branches = %d, want 0 (inline form)", len(cfg.Branches))
	}
	if cfg.Params["fan_in_policy"] != "all" {
		t.Errorf("params[fan_in_policy] = %q, want all", cfg.Params["fan_in_policy"])
	}
}

func TestParseFanInParams(t *testing.T) {
	input := `workflow Test
  start: P
  exit: J

  agent A
    prompt: "A"
  agent B
    prompt: "B"

  parallel P -> A, B

  fan_in J <- A, B
    params:
      fan_in_policy: quorum
      quorum: 2

  edges
    J -> J
`
	w, err := NewParser(input, "test.dip").Parse()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	cfg := findNode(t, w, "J").Config.(ir.FanInConfig)
	if len(cfg.Sources) != 2 {
		t.Errorf("sources = %d, want 2", len(cfg.Sources))
	}
	if cfg.Params["fan_in_policy"] != "quorum" {
		t.Errorf("params[fan_in_policy] = %q, want quorum", cfg.Params["fan_in_policy"])
	}
	if cfg.Params["quorum"] != "2" {
		t.Errorf("params[quorum] = %q, want 2", cfg.Params["quorum"])
	}
}

// TestParseInlineParallelNoParamsStillWorks guards that the optional block does
// not break the bare inline form (no trailing params).
func TestParseInlineParallelNoParamsStillWorks(t *testing.T) {
	input := `workflow Test
  start: P
  exit: J

  agent A
    prompt: "A"
  agent B
    prompt: "B"

  parallel P -> A, B
  fan_in J <- A, B

  edges
    J -> J
`
	w, err := NewParser(input, "test.dip").Parse()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	pcfg := findNode(t, w, "P").Config.(ir.ParallelConfig)
	if len(pcfg.Targets) != 2 {
		t.Errorf("targets = %d, want 2", len(pcfg.Targets))
	}
	if len(pcfg.Params) != 0 {
		t.Errorf("parallel params = %d, want 0", len(pcfg.Params))
	}
	fcfg := findNode(t, w, "J").Config.(ir.FanInConfig)
	if len(fcfg.Params) != 0 {
		t.Errorf("fan_in params = %d, want 0", len(fcfg.Params))
	}
}

// TestParseParallelParamsAlwaysNonNil guards the agent/subgraph convention that
// Params is non-nil even when no params block is present.
func TestParseParallelParamsAlwaysNonNil(t *testing.T) {
	input := `workflow Test
  start: P
  exit: J

  agent A
    prompt: "A"
  agent B
    prompt: "B"

  parallel P -> A, B
  fan_in J <- A, B

  edges
    J -> J
`
	w, err := NewParser(input, "test.dip").Parse()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if findNode(t, w, "P").Config.(ir.ParallelConfig).Params == nil {
		t.Error("inline parallel Params is nil, want non-nil empty map")
	}
	if findNode(t, w, "J").Config.(ir.FanInConfig).Params == nil {
		t.Error("fan_in Params is nil, want non-nil empty map")
	}
}

// TestParseParallelBlockParamsAlwaysNonNil guards non-nil Params for block form.
func TestParseParallelBlockParamsAlwaysNonNil(t *testing.T) {
	input := `workflow Test
  start: P
  exit: J

  agent A
    prompt: "A"
  agent B
    prompt: "B"

  parallel P
    branch: A
    branch: B

  fan_in J <- A, B

  edges
    J -> J
`
	w, err := NewParser(input, "test.dip").Parse()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if findNode(t, w, "P").Config.(ir.ParallelConfig).Params == nil {
		t.Error("block parallel Params is nil, want non-nil empty map")
	}
}

// TestParseParallelMalformedNestedBlockNoDesync guards that a malformed nested
// block beneath an inline parallel does not swallow the rest of the workflow.
func TestParseParallelMalformedNestedBlockNoDesync(t *testing.T) {
	input := `workflow Test
  start: P
  exit: J

  agent A
    prompt: "A"

  parallel P -> A, A
    foo: bar
      baz: qux

  fan_in J <- A, A

  edges
    J -> J
`
	w, err := NewParser(input, "test.dip").Parse()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// The fan_in node after the malformed block must still be parsed.
	var sawFanIn bool
	for _, n := range w.Nodes {
		if n.Kind == ir.NodeFanIn {
			sawFanIn = true
		}
	}
	if !sawFanIn {
		t.Error("fan_in node after malformed nested block was dropped (parse desync)")
	}
}
