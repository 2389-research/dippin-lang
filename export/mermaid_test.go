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
