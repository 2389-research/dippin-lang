package formatter

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/2389-research/dippin-lang/ir"
	"github.com/2389-research/dippin-lang/parser"
	"github.com/2389-research/dippin-lang/simulate"
)

// v1wf builds a v1 workflow with the given retry config on node "T".
func v1wf(t *testing.T, retry ir.RetryConfig, extra []*ir.Edge) *ir.Workflow {
	t.Helper()
	nodes := []*ir.Node{
		{ID: "T", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "run"}, Retry: retry},
		{ID: "Done", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "d"}},
		{ID: "Esc", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "e"}},
	}
	edges := append([]*ir.Edge{{From: "T", To: "Done", Condition: &ir.Condition{Raw: "ctx.outcome = success"}}}, extra...)
	w := &ir.Workflow{Name: "W", Version: "1", Start: "T", Exit: "Done", Nodes: nodes, Edges: edges}
	_ = simulate.EnsureConditionsParsed(w)
	return w
}

func failTo(to string) *ir.Edge {
	return &ir.Edge{From: "T", To: to, Condition: &ir.Condition{Raw: "ctx.outcome = fail"}}
}

func edgeCount(w *ir.Workflow, to string) int {
	n := 0
	for _, e := range w.Edges {
		if e.From == "T" && e.To == to {
			n++
		}
	}
	return n
}

// Under #204 Option A the retry channel stays on the node — migration is a
// lossless version bump. fallback_target is kept as FallbackTarget (emitted as
// the dip-2 spelling fallback_retry_target by the formatter), NOT converted to
// an on-fail edge.
func TestMigrate_FallbackKeptAsAttr(t *testing.T) {
	w := v1wf(t, ir.RetryConfig{FallbackTarget: "Esc"}, nil)
	notes := MigrateToV2(w)
	if len(notes) != 0 {
		t.Fatalf("lossless migration should have no notes, got %v", notes)
	}
	if w.Version != "2" {
		t.Fatalf("version = %q, want 2", w.Version)
	}
	if w.Nodes[0].Retry.FallbackTarget != "Esc" {
		t.Errorf("FallbackTarget must be kept, got %q", w.Nodes[0].Retry.FallbackTarget)
	}
	if edgeCount(w, "Esc") != 0 {
		t.Errorf("must NOT synthesize an edge to Esc")
	}
	if !strings.Contains(Format(w), "fallback_retry_target: Esc") {
		t.Errorf("formatter must emit the dip-2 spelling:\n%s", Format(w))
	}
}

// retry_target (self or non-self) is kept as a node attr — no loop edge.
func TestMigrate_RetryTargetKeptAsAttr(t *testing.T) {
	for _, target := range []string{"Esc", "T"} {
		w := v1wf(t, ir.RetryConfig{RetryTarget: target, MaxRetries: 2}, nil)
		if notes := MigrateToV2(w); len(notes) != 0 {
			t.Fatalf("retry_target=%s: lossless, want no notes, got %v", target, notes)
		}
		if w.Nodes[0].Retry.RetryTarget != target {
			t.Errorf("retry_target=%s: RetryTarget must be kept, got %q", target, w.Nodes[0].Retry.RetryTarget)
		}
		if edgeCount(w, target) != 0 {
			t.Errorf("retry_target=%s: must NOT synthesize a loop edge", target)
		}
	}
}

// A fallback_target and a genuine on-fail edge to a different node are distinct
// channels and both survive migration (no conversion, no conflict).
func TestMigrate_FallbackAndFailEdgeCoexist(t *testing.T) {
	w := v1wf(t, ir.RetryConfig{FallbackTarget: "Esc"}, []*ir.Edge{failTo("Done")})
	if notes := MigrateToV2(w); len(notes) != 0 {
		t.Fatalf("want no notes, got %v", notes)
	}
	if w.Nodes[0].Retry.FallbackTarget != "Esc" {
		t.Errorf("FallbackTarget must survive, got %q", w.Nodes[0].Retry.FallbackTarget)
	}
	if edgeCount(w, "Done") != 2 { // the success edge + the pre-existing on-fail edge
		t.Errorf("the pre-existing on-fail edge to Done must survive, edges=%v", w.Edges)
	}
}

// dip 2 round-trips the retry attrs: parse a dip-2 file with the attrs, format,
// re-parse — the attrs are preserved.
func TestMigrate_Dip2RetryAttrsRoundTrip(t *testing.T) {
	src := `dip 2

workflow W
  goal: "t"
  start: A
  exit: B

  agent A
    prompt: a
    max_retries: 3
    retry_target: B
    fallback_retry_target: B

  agent B
    prompt: b

  edges
    A -> B
`
	w, err := parser.NewParser(src, "t.dip").Parse()
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if w.Nodes[0].Retry.RetryTarget != "B" || w.Nodes[0].Retry.FallbackTarget != "B" {
		t.Fatalf("parsed retry attrs = %q/%q, want B/B", w.Nodes[0].Retry.RetryTarget, w.Nodes[0].Retry.FallbackTarget)
	}
	out := Format(w)
	w2, err := parser.NewParser(out, "t.dip").Parse()
	if err != nil {
		t.Fatalf("re-parse: %v\n%s", err, out)
	}
	if w2.Nodes[0].Retry.RetryTarget != "B" || w2.Nodes[0].Retry.FallbackTarget != "B" {
		t.Errorf("round-trip lost retry attrs: %q/%q\n%s", w2.Nodes[0].Retry.RetryTarget, w2.Nodes[0].Retry.FallbackTarget, out)
	}
}

func TestMigrate_ExamplesRoundTripToValidV2(t *testing.T) {
	matches, _ := filepath.Glob("../examples/*.dip")
	for _, path := range matches {
		src, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		w, err := parser.NewParser(string(src), path).Parse()
		if err != nil {
			t.Fatalf("%s: parse: %v", path, err)
		}
		_ = simulate.EnsureConditionsParsed(w)
		MigrateToV2(w)
		out := Format(w)
		if _, err := parser.NewParser(out, path).Parse(); err != nil {
			t.Errorf("%s: migrated v2 does not re-parse: %v\n%s", path, err, out)
		}
	}
}
