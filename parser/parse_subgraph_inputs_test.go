package parser

import (
	"strings"
	"testing"

	"github.com/2389-research/dippin-lang/ir"
)

// subgraphSrc builds a workflow with a subgraph node whose call-site binding
// uses the given keyword (`params` or `inputs`), optionally under a dip-N header.
func subgraphSrc(header, keyword string) string {
	return header + `workflow P
  start: S
  exit: S
  subgraph S
    ref: child.dip
    ` + keyword + `:
      topic: hi
`
}

func subgraphParams(t *testing.T, w *ir.Workflow) map[string]string {
	t.Helper()
	return w.Nodes[0].Config.(ir.SubgraphConfig).Params
}

// TestSubgraph_Dip1AcceptsParams: dip 1 keeps the `params:` call-site binding.
func TestSubgraph_Dip1AcceptsParams(t *testing.T) {
	w, err := NewParser(subgraphSrc("", "params"), "p.dip").Parse()
	if err != nil {
		t.Fatalf("dip 1 must accept subgraph params: %v", err)
	}
	if subgraphParams(t, w)["topic"] != "hi" {
		t.Errorf("params[topic] = %q, want hi", subgraphParams(t, w)["topic"])
	}
}

// TestSubgraph_Dip2AcceptsInputs: dip 2 spells the binding `inputs:` (#227).
func TestSubgraph_Dip2AcceptsInputs(t *testing.T) {
	w, err := NewParser(subgraphSrc("dip 2\n\n", "inputs"), "p.dip").Parse()
	if err != nil {
		t.Fatalf("dip 2 must accept subgraph inputs: %v", err)
	}
	if subgraphParams(t, w)["topic"] != "hi" {
		t.Errorf("inputs[topic] = %q, want hi", subgraphParams(t, w)["topic"])
	}
}

// TestSubgraph_Dip2RejectsParams: dip 2 rejects `params:` on a subgraph and
// points at `inputs:`.
func TestSubgraph_Dip2RejectsParams(t *testing.T) {
	_, err := NewParser(subgraphSrc("dip 2\n\n", "params"), "p.dip").Parse()
	if err == nil {
		t.Fatal("dip 2 must reject subgraph params:")
	}
	if !strings.Contains(err.Error(), "inputs:") {
		t.Errorf("rejection should point to inputs:, got %v", err)
	}
}
