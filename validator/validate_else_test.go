package validator

import (
	"strings"
	"testing"

	"github.com/2389-research/dippin-lang/ir"
)

// TestElseTargetUnknownNode: a section `else` pointing at a non-existent node is
// a DIP003 unknown-node error, mirroring on_failure.
func TestElseTargetUnknownNode(t *testing.T) {
	w := minimalValidWorkflow()
	w.ElseTarget = "Ghost"
	diags := Validate(w).Diagnostics
	if !hasCodeMentioning(diags, DIP003, "else", "Ghost") {
		t.Fatalf("expected DIP003 mentioning else and Ghost, got: %v", diags)
	}
}

// TestElseTargetKnownNodeOK: an else pointing at an existing node raises no
// unknown-node error.
func TestElseTargetKnownNodeOK(t *testing.T) {
	w := minimalValidWorkflow()
	w.ElseTarget = "End"
	for _, d := range Validate(w).Diagnostics {
		if d.Code == DIP003 {
			t.Fatalf("unexpected DIP003 for valid else target: %v", d)
		}
	}
}

// TestElseTargetReachable: a node reachable ONLY as the else target must not be
// flagged DIP004 (unreachable), mirroring on_failure reachability.
func TestElseTargetReachable(t *testing.T) {
	w := &ir.Workflow{
		Name:  "elsereach",
		Start: "A",
		Exit:  "B",
		Nodes: []*ir.Node{
			{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "a"}},
			{ID: "B", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "b"}},
			{ID: "Cleanup", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "clean"}},
		},
		Edges:      []*ir.Edge{{From: "A", To: "B"}},
		ElseTarget: "Cleanup",
	}
	for _, d := range Validate(w).Diagnostics {
		if d.Code == DIP004 && strings.Contains(d.Message, "Cleanup") {
			t.Fatalf("Cleanup is reachable via else but was flagged DIP004: %v", d)
		}
	}
}

// hasCodeMentioning reports whether any diagnostic has the given code and its
// message contains all the given substrings.
func hasCodeMentioning(diags []Diagnostic, code string, subs ...string) bool {
	for _, d := range diags {
		if d.Code != code {
			continue
		}
		ok := true
		for _, s := range subs {
			if !strings.Contains(d.Message, s) {
				ok = false
				break
			}
		}
		if ok {
			return true
		}
	}
	return false
}
