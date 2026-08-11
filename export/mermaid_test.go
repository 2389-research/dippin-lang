package export_test

import (
	"strings"
	"testing"

	"github.com/2389-research/dippin-lang/export"
	"github.com/2389-research/dippin-lang/parser"
)

func mermaidOf(t *testing.T, src string) string {
	t.Helper()
	w, err := parser.NewParser(src, "t.dip").Parse()
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	return export.ExportMermaid(w)
}

func TestExportMermaid_ShapesAndClassesByKind(t *testing.T) {
	out := mermaidOf(t, `workflow W
  start: A
  exit: Done
  agent A
    prompt: a
  human H
    mode: freeform
  tool T
    command: echo
  agent Done
    prompt: d
  edges
    A -> H
    H -> T
    T -> Done`)
	for _, want := range []string{
		"flowchart TD",
		`A(["A"]):::agent`,    // agent → stadium
		`H[/"H"/]:::human`,    // human → parallelogram
		`T["T"]:::tool`,       // tool → rectangle
		"class A startNode",   // start emphasis
		"class Done exitNode", // exit emphasis
	} {
		if !strings.Contains(out, want) {
			t.Errorf("mermaid missing %q:\n%s", want, out)
		}
	}
}

func TestExportMermaid_EdgeLabels(t *testing.T) {
	out := mermaidOf(t, `workflow W
  start: Gate
  exit: Done
  defaults
    on_failure: Done
  agent Gate
    prompt: g
  agent Done
    prompt: d
  edges
    Gate -> Done  when ctx.outcome = success
    Gate -> Gate  when ctx.outcome = fail`)
	// Outcome test compacts to just the value.
	if !strings.Contains(out, `Gate -->|"success"| Done`) {
		t.Errorf("expected compacted success label:\n%s", out)
	}
	if !strings.Contains(out, `Gate -->|"fail"| Gate`) {
		t.Errorf("expected compacted fail label:\n%s", out)
	}
}

func TestExportMermaid_IncludesFanEdges(t *testing.T) {
	// A parallel node's inline targets and a fan_in node's inline sources are not
	// in w.Edges, but the graph must still connect them (#254 review).
	out := mermaidOf(t, `workflow W
  start: Init
  exit: Done
  agent Init
    prompt: i
  parallel Fan -> WorkerA, WorkerB
  agent WorkerA
    prompt: a
  agent WorkerB
    prompt: b
  fan_in Join <- WorkerA, WorkerB
  agent Done
    prompt: d
  edges
    Init -> Fan
    Join -> Done`)
	for _, want := range []string{"Fan --> WorkerA", "Fan --> WorkerB", "WorkerA --> Join", "WorkerB --> Join"} {
		if !strings.Contains(out, want) {
			t.Errorf("mermaid missing fan edge %q:\n%s", want, out)
		}
	}
}

func TestExportMermaid_UnconditionalEdgeHasNoLabel(t *testing.T) {
	out := mermaidOf(t, `workflow W
  start: A
  exit: B
  agent A
    prompt: a
  agent B
    prompt: b
  edges
    A -> B`)
	if !strings.Contains(out, "A --> B\n") {
		t.Errorf("unconditional edge should be unlabeled:\n%s", out)
	}
}

func TestExportMermaid_DistinctIDsDoNotCollide(t *testing.T) {
	// `a.b` and `a-b` both sanitize to `a_b`; they must NOT collapse to one node.
	out := mermaidOf(t, `workflow W
  start: "a.b"
  exit: Done
  agent "a.b"
    prompt: x
  agent "a-b"
    prompt: y
  agent Done
    prompt: d
  edges
    "a.b" -> "a-b"
    "a-b" -> Done`)
	if !strings.Contains(out, `a_b(["a.b"])`) || !strings.Contains(out, `a_b_2(["a-b"])`) {
		t.Errorf("colliding IDs must be disambiguated:\n%s", out)
	}
	// The edge between them must reference the two distinct ids.
	if !strings.Contains(out, "a_b --> a_b_2") {
		t.Errorf("edge should connect the two disambiguated nodes:\n%s", out)
	}
}

func TestExportMermaid_UsesNodeLabel(t *testing.T) {
	// A node with a Label shows the label as display text, not the bare ID.
	out := mermaidOf(t, `workflow W
  start: Check
  exit: Done
  conditional Check
    label: "Evaluate outcome"
  agent Done
    prompt: d
  edges
    Check -> Done`)
	if !strings.Contains(out, `Check{"Evaluate outcome"}`) {
		t.Errorf("labelled node should display its label:\n%s", out)
	}
}

func TestExportMermaid_SanitizesNodeIDsAndText(t *testing.T) {
	// A dotted/hyphenated ID must sanitize the mermaid node id but keep the
	// original text as the display label.
	out := mermaidOf(t, `workflow W
  start: "step-1.a"
  exit: "step-1.a"
  agent "step-1.a"
    prompt: x`)
	if !strings.Contains(out, `step_1_a(["step-1.a"])`) {
		t.Errorf("expected sanitized id with original display text:\n%s", out)
	}
}
