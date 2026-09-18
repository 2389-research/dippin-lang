package formatter

import (
	"strings"
	"testing"

	"github.com/2389-research/dippin-lang/ir"
	"github.com/2389-research/dippin-lang/parser"
)

// TestFormatAgentWritablePathsModeSlot pins the canonical slot: the mode line
// is emitted directly after writable_paths.
func TestFormatAgentWritablePathsModeSlot(t *testing.T) {
	w := &ir.Workflow{
		Name: "T", Start: "A", Exit: "A",
		Nodes: []*ir.Node{
			{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{
				Prompt:               "x",
				WritablePaths:        []string{"workspace/**"},
				WritablePathsMode:    "prefer",
				LastResponseTruncate: 100,
			}},
		},
	}
	out := Format(w)
	want := "    writable_paths: workspace/**\n    writable_paths_mode: prefer\n    last_response_truncate: 100\n"
	if !strings.Contains(out, want) {
		t.Errorf("formatted output missing canonical writable_paths_mode slot; want\n%s\ngot:\n%s", want, out)
	}
}

// TestFormatAgentWritablePathsModeAbsentOmitted: fmt never invents the attribute.
func TestFormatAgentWritablePathsModeAbsentOmitted(t *testing.T) {
	w := &ir.Workflow{
		Name: "T", Start: "A", Exit: "A",
		Nodes: []*ir.Node{
			{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "x", WritablePaths: []string{"workspace/**"}}},
		},
	}
	if out := Format(w); strings.Contains(out, "writable_paths_mode") {
		t.Errorf("fmt must not emit writable_paths_mode when absent; got:\n%s", out)
	}
}

// TestFormatAgentWritablePathsModeVerbatimQuoted: a bad spelling with trailing
// space is re-emitted quoted so it survives a second parse (fmt does not
// launder a value DIP163 must still be able to see).
func TestFormatAgentWritablePathsModeVerbatimQuoted(t *testing.T) {
	w := &ir.Workflow{
		Name: "T", Start: "A", Exit: "A",
		Nodes: []*ir.Node{
			{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "x", WritablePathsMode: "prefer "}},
		},
	}
	out := Format(w)
	if !strings.Contains(out, `writable_paths_mode: "prefer "`) {
		t.Errorf("expected quoted verbatim value; got:\n%s", out)
	}
	w2, err := parser.NewParser(out, "rt.dip").Parse()
	if err != nil {
		t.Fatalf("reparse: %v", err)
	}
	if got := w2.Node("A").Config.(ir.AgentConfig).WritablePathsMode; got != "prefer " {
		t.Errorf("round-trip = %q, want %q", got, "prefer ")
	}
}

func TestFormatBranchWritablePathsModeOnly(t *testing.T) {
	w := &ir.Workflow{
		Name: "T", Start: "split", Exit: "join",
		Nodes: []*ir.Node{
			{ID: "a", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "a"}},
			{ID: "split", Kind: ir.NodeParallel, Config: ir.ParallelConfig{
				Targets:  []string{"a"},
				Branches: []ir.BranchConfig{{Target: "a", WritablePathsMode: "prefer"}},
			}},
			{ID: "join", Kind: ir.NodeFanIn, Config: ir.FanInConfig{Sources: []string{"a"}}},
		},
	}
	out := Format(w)
	if !strings.Contains(out, "writable_paths_mode: prefer") {
		t.Errorf("formatted output missing per-branch writable_paths_mode; got:\n%s", out)
	}
}

func TestFormatBranchWritablePathsModeSlot(t *testing.T) {
	w := &ir.Workflow{
		Name: "T", Start: "split", Exit: "join",
		Nodes: []*ir.Node{
			{ID: "a", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "a"}},
			{ID: "split", Kind: ir.NodeParallel, Config: ir.ParallelConfig{
				Targets: []string{"a"},
				Branches: []ir.BranchConfig{{
					Target: "a", WritablePaths: []string{"workspace/**"}, WritablePathsMode: "require", LastResponseTruncate: 5,
				}},
			}},
			{ID: "join", Kind: ir.NodeFanIn, Config: ir.FanInConfig{Sources: []string{"a"}}},
		},
	}
	out := Format(w)
	want := "writable_paths: workspace/**\n      writable_paths_mode: require\n      last_response_truncate: 5\n"
	if !strings.Contains(out, want) {
		t.Errorf("branch slot wrong; want\n%s\ngot:\n%s", want, out)
	}
}

const writablePathsModeRoundTripSrc = `workflow X
  start: split
  exit: join

  agent A
    prompt: "x"
    writable_paths: workspace/**, .ai/sprints/**
    writable_paths_mode: prefer

  parallel split
    branch: A
      writable_paths: workspace/**
      writable_paths_mode: require

  fan_in join <- A

  edges
    split -> A
    A -> join
`

func TestFormatWritablePathsModeRoundTripsAndIsIdempotent(t *testing.T) {
	w1, err := parser.NewParser(writablePathsModeRoundTripSrc, "rt.dip").Parse()
	if err != nil {
		t.Fatalf("parse1: %v", err)
	}
	once := Format(w1)
	w2, err := parser.NewParser(once, "rt.dip").Parse()
	if err != nil {
		t.Fatalf("parse2: %v", err)
	}
	if got := w2.Node("A").Config.(ir.AgentConfig).WritablePathsMode; got != "prefer" {
		t.Errorf("agent WritablePathsMode after round-trip = %q, want prefer", got)
	}
	if got := w2.Node("split").Config.(ir.ParallelConfig).Branches[0].WritablePathsMode; got != "require" {
		t.Errorf("branch WritablePathsMode after round-trip = %q, want require", got)
	}
	if twice := Format(w2); twice != once {
		t.Errorf("format not idempotent:\n--- once ---\n%s\n--- twice ---\n%s", once, twice)
	}
}
