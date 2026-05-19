package parser

import (
	"strings"
	"testing"

	"github.com/2389-research/dippin-lang/ir"
)

// TestParseBoolAttrFields verifies that all four bool fields
// (goal_gate, auto_status, cache_tools, route_required) accept
// the canonical truthy/falsy forms case-insensitively, and that
// any other value produces a parse diagnostic.
func TestParseBoolAttrFields(t *testing.T) {
	cases := []struct {
		name        string
		field       string // "goal_gate" | "auto_status" | "cache_tools" | "route_required"
		val         string
		wantBool    bool
		wantDiagSub string // substring expected in diagnostics; "" means no diag
	}{
		// agent fields — goal_gate
		{"goal_gate=true", "goal_gate", "true", true, ""},
		{"goal_gate=FALSE", "goal_gate", "FALSE", false, ""},
		{"goal_gate=yes", "goal_gate", "yes", true, ""},
		{"goal_gate=no", "goal_gate", "no", false, ""},
		{"goal_gate=1", "goal_gate", "1", true, ""},
		{"goal_gate=0", "goal_gate", "0", false, ""},
		{"goal_gate=on", "goal_gate", "on", true, ""},
		{"goal_gate=Off", "goal_gate", "Off", false, ""},
		{"goal_gate=maybe", "goal_gate", "maybe", false, "invalid boolean"},
		{"goal_gate=2", "goal_gate", "2", false, "invalid boolean"},

		// agent fields — auto_status
		{"auto_status=yes", "auto_status", "yes", true, ""},
		{"auto_status=garbage", "auto_status", "garbage", false, "invalid boolean"},

		// agent fields — cache_tools
		{"cache_tools=on", "cache_tools", "on", true, ""},
		{"cache_tools=nope", "cache_tools", "nope", false, "invalid boolean"},

		// tool field — route_required
		{"route_required=YES", "route_required", "YES", true, ""},
		{"route_required=true", "route_required", "true", true, ""},
		{"route_required=False", "route_required", "False", false, ""},
		{"route_required=maybe", "route_required", "maybe", false, "invalid boolean"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			src := buildBoolFieldDip(tc.field, tc.val)
			p := NewParser(src, "test.dip")
			w, _ := p.Parse()
			diags := p.Diagnostics()
			got := readBoolField(t, w, tc.field)
			if got != tc.wantBool {
				t.Errorf("field %s = %q: got %v, want %v", tc.field, tc.val, got, tc.wantBool)
			}
			joined := strings.Join(diags, "\n")
			if tc.wantDiagSub == "" {
				if joined != "" {
					t.Errorf("expected no diagnostics, got: %s", joined)
				}
				return
			}
			if !strings.Contains(joined, tc.wantDiagSub) {
				t.Errorf("expected diagnostic containing %q, got: %s", tc.wantDiagSub, joined)
			}
		})
	}
}

// buildBoolFieldDip produces a minimal valid .dip workflow that exercises
// one of the four bool fields. The agent / tool node is the start node so
// the workflow always has at least one node.
func buildBoolFieldDip(field, val string) string {
	switch field {
	case "goal_gate", "auto_status", "cache_tools":
		return "workflow X\n" +
			"  start: A\n" +
			"  exit: A\n" +
			"  agent A\n" +
			"    " + field + ": " + val + "\n"
	case "route_required":
		return "workflow X\n" +
			"  start: T\n" +
			"  exit: T\n" +
			"  tool T\n" +
			"    command: echo hi\n" +
			"    " + field + ": " + val + "\n"
	}
	panic("unknown field: " + field)
}

// readBoolField extracts the value of the bool field under test from the
// parsed workflow.
func readBoolField(t *testing.T, w *ir.Workflow, field string) bool {
	t.Helper()
	if len(w.Nodes) == 0 {
		t.Fatal("workflow has no nodes")
	}
	n := w.Nodes[0]
	switch field {
	case "goal_gate":
		cfg, ok := n.Config.(ir.AgentConfig)
		if !ok {
			t.Fatalf("expected AgentConfig, got %T", n.Config)
		}
		return cfg.GoalGate
	case "auto_status":
		cfg, ok := n.Config.(ir.AgentConfig)
		if !ok {
			t.Fatalf("expected AgentConfig, got %T", n.Config)
		}
		return cfg.AutoStatus
	case "cache_tools":
		cfg, ok := n.Config.(ir.AgentConfig)
		if !ok {
			t.Fatalf("expected AgentConfig, got %T", n.Config)
		}
		return cfg.CacheTools
	case "route_required":
		cfg, ok := n.Config.(ir.ToolConfig)
		if !ok {
			t.Fatalf("expected ToolConfig, got %T", n.Config)
		}
		return cfg.RouteRequired
	}
	t.Fatalf("unknown field: %s", field)
	return false
}
