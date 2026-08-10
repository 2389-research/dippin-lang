package formatter

import (
	"strings"
	"testing"

	"github.com/2389-research/dippin-lang/ir"
	"github.com/2389-research/dippin-lang/parser"
)

func subgraphWF(version string) *ir.Workflow {
	return &ir.Workflow{
		Name: "P", Version: version, Start: "S", Exit: "S",
		Nodes: []*ir.Node{{
			ID: "S", Kind: ir.NodeSubgraph,
			Config: ir.SubgraphConfig{Ref: "child.dip", Params: map[string]string{"topic": "hi"}},
		}},
	}
}

// TestFormatSubgraph_Dip1EmitsParams: a v1 workflow formats the binding as params:.
func TestFormatSubgraph_Dip1EmitsParams(t *testing.T) {
	out := Format(subgraphWF("1"))
	if !strings.Contains(out, "params:") || strings.Contains(out, "inputs:") {
		t.Errorf("dip 1 should emit params:, not inputs:\n%s", out)
	}
}

// TestFormatSubgraph_Dip2EmitsInputs: a dip-2 workflow formats the binding as
// inputs: and round-trips (#227).
func TestFormatSubgraph_Dip2EmitsInputs(t *testing.T) {
	out := Format(subgraphWF("2"))
	if !strings.Contains(out, "inputs:") || strings.Contains(out, "params:") {
		t.Errorf("dip 2 should emit inputs:, not params:\n%s", out)
	}
	w2, err := parser.NewParser(out, "p.dip").Parse()
	if err != nil {
		t.Fatalf("re-parse: %v\n%s", err, out)
	}
	if w2.Nodes[0].Config.(ir.SubgraphConfig).Params["topic"] != "hi" {
		t.Errorf("round-trip lost the binding:\n%s", out)
	}
}

// TestMigrateSubgraph_ParamsRelabeledToInputs: fmt --migrate (version bump +
// format) relabels a v1 subgraph params: to the dip-2 inputs: spelling (#227),
// losslessly.
func TestMigrateSubgraph_ParamsRelabeledToInputs(t *testing.T) {
	w := subgraphWF("1")
	if notes := MigrateToV2(w); len(notes) != 0 {
		t.Fatalf("lossless migration, want no notes, got %v", notes)
	}
	out := Format(w)
	if !strings.Contains(out, "inputs:") || strings.Contains(out, "params:") {
		t.Errorf("migrated dip-2 output should spell the binding inputs:\n%s", out)
	}
}
