package migrate

import (
	"fmt"
	"sort"
	"strings"
	"testing"

	"github.com/2389-research/dippin-lang/export"
	"github.com/2389-research/dippin-lang/ir"
	dipparser "github.com/2389-research/dippin-lang/parser"
)

// TestExportDOT_ManagerLoop_SteerContextReservedCharsRoundTrip verifies that
// steer_context values with reserved delimiters (',', '=', '%') round-trip
// losslessly through ExportDOT → parseFlattenedSteerContext. The test extracts
// the steer_context attr value directly from the DOT string and passes it
// through parseFlattenedSteerContext to isolate the delimiter-escaping logic
// from the surrounding graph structure.
func TestExportDOT_ManagerLoop_SteerContextReservedCharsRoundTrip(t *testing.T) {
	// Values with reserved delimiters (',', '=', '%') must round-trip losslessly
	// through flattenSteerContext → parseFlattenedSteerContext.
	original := map[string]string{
		"hint":     "be concise, technical",
		"priority": "high=medium",
		"meta":     "100%",
	}
	w := &ir.Workflow{
		Name:  "W",
		Start: "S",
		Exit:  "E",
		Nodes: []*ir.Node{
			{ID: "S", Kind: ir.NodeAgent, Config: ir.AgentConfig{}},
			{ID: "M", Kind: ir.NodeManagerLoop, Config: ir.ManagerLoopConfig{
				SubgraphRef:  "inner",
				MaxCycles:    5,
				SteerContext: original,
			}},
			{ID: "E", Kind: ir.NodeAgent, Config: ir.AgentConfig{}},
		},
		Edges: []*ir.Edge{{From: "S", To: "M"}, {From: "M", To: "E"}},
	}
	dot := export.ExportDOT(w, export.ExportOptions{})

	// Extract the steer_context attribute value from the DOT string.
	// It appears as: steer_context="<value>"
	attrVal := extractDOTAttr(dot, "steer_context")
	if attrVal == "" {
		t.Fatalf("steer_context attr not found in DOT output:\n%s", dot)
	}

	// Decode via parseFlattenedSteerContext (the migrate path).
	got := parseFlattenedSteerContext(attrVal)

	for k, want := range original {
		if got[k] != want {
			t.Errorf("round-trip lost %q: got %q want %q", k, got[k], want)
		}
	}
}

// TestExportDOT_ManagerLoop_SteerContextNoReservedCharsPassThrough verifies
// that plain keys/values (no reserved chars) are not mangled by the encoding.
func TestExportDOT_ManagerLoop_SteerContextNoReservedCharsPassThrough(t *testing.T) {
	original := map[string]string{
		"hint":     "speed_up",
		"priority": "high",
	}
	w := &ir.Workflow{
		Name:  "W",
		Start: "S",
		Exit:  "E",
		Nodes: []*ir.Node{
			{ID: "S", Kind: ir.NodeAgent, Config: ir.AgentConfig{}},
			{ID: "M", Kind: ir.NodeManagerLoop, Config: ir.ManagerLoopConfig{
				SubgraphRef:  "inner",
				MaxCycles:    3,
				SteerContext: original,
			}},
			{ID: "E", Kind: ir.NodeAgent, Config: ir.AgentConfig{}},
		},
		Edges: []*ir.Edge{{From: "S", To: "M"}, {From: "M", To: "E"}},
	}
	dot := export.ExportDOT(w, export.ExportOptions{})

	attrVal := extractDOTAttr(dot, "steer_context")
	if attrVal == "" {
		t.Fatalf("steer_context attr not found in DOT output:\n%s", dot)
	}

	// Plain values must not be mangled — the encoded form should equal "k=v,k=v".
	keys := []string{"hint", "priority"}
	sort.Strings(keys)
	var parts []string
	for _, k := range keys {
		parts = append(parts, k+"="+original[k])
	}
	wantFlat := strings.Join(parts, ",")
	if attrVal != wantFlat {
		t.Errorf("plain steer_context encoded as %q, want %q", attrVal, wantFlat)
	}

	got := parseFlattenedSteerContext(attrVal)
	for k, want := range original {
		if got[k] != want {
			t.Errorf("round-trip changed %q: got %q want %q", k, got[k], want)
		}
	}
}

// extractDOTAttr extracts the value of a DOT attribute from a line like:
//
//	key="value"
//
// It handles the quoted format produced by ExportDOT. Returns "" if not found.
func extractDOTAttr(dot, key string) string {
	needle := fmt.Sprintf(`%s="`, key)
	idx := strings.Index(dot, needle)
	if idx == -1 {
		return ""
	}
	start := idx + len(needle)
	end := strings.Index(dot[start:], `"`)
	if end == -1 {
		return ""
	}
	return dot[start : start+end]
}

// TestRoundtripPreservesToolOutputs verifies that ToolConfig.Outputs
// survives a .dip → DOT → .dip migration. Regression test for the
// silent-drop bug in v0.28.0 where applyToolSemanticAttrs never
// emitted the outputs DOT attr.
func TestRoundtripPreservesToolOutputs(t *testing.T) {
	src := `workflow MarkerRouting
  start: S
  exit: D
  agent S
    prompt: "start"
  tool T
    command: echo hi
    outputs: tests_green, tests_red
    marker_grep: "^(tests_green|tests_red)$"
  agent D
    prompt: "done"
  edges
    S -> T
    T -> D when ctx.tool_marker = tests_green
    T -> D when ctx.tool_marker = tests_red
`
	w1, err := dipparser.NewParser(src, "rt.dip").Parse()
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	dot := export.ExportDOT(w1, export.ExportOptions{IncludePrompts: true})

	// DOT must contain outputs attr with comma-joined values.
	if !strings.Contains(dot, `outputs="tests_green,tests_red"`) {
		t.Errorf("DOT output missing outputs attr; got:\n%s", dot)
	}

	w2, err := Migrate(dot)
	if err != nil {
		t.Fatalf("migrate: %v\nDOT:\n%s", err, dot)
	}

	var got []string
	for _, n := range w2.Nodes {
		if n.ID != "T" {
			continue
		}
		cfg, ok := n.Config.(ir.ToolConfig)
		if !ok {
			t.Fatalf("node T config is %T, want ir.ToolConfig", n.Config)
		}
		got = cfg.Outputs
	}
	want := []string{"tests_green", "tests_red"}
	if len(got) != len(want) {
		t.Fatalf("Outputs after round-trip = %v, want %v; DOT:\n%s", got, want, dot)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("Outputs[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

// TestRoundTripToolSafetyDefaults verifies that tool-safety defaults
// round-trip losslessly through parse → ExportDOT → Migrate.
func TestRoundTripToolSafetyDefaults(t *testing.T) {
	src := `workflow ToolSafety
  goal: "round trip"
  start: A
  exit: A

  defaults
    tool_commands_allow: "git *,make *"
    tool_denylist_add: "rm -rf /"

  agent A
    prompt: "Do it."

  edges
    A -> A
`
	w1, err := dipparser.NewParser(src, "rt.dip").Parse()
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	dot := export.ExportDOT(w1, export.ExportOptions{})
	w2, err := Migrate(dot)
	if err != nil {
		t.Fatalf("migrate: %v\nDOT:\n%s", err, dot)
	}
	if w2.Defaults.ToolCommandsAllow != "git *,make *" {
		t.Errorf("tool_commands_allow after round-trip = %q, want %q; DOT:\n%s",
			w2.Defaults.ToolCommandsAllow, "git *,make *", dot)
	}
	if w2.Defaults.ToolDenylistAdd != "rm -rf /" {
		t.Errorf("tool_denylist_add after round-trip = %q, want %q; DOT:\n%s",
			w2.Defaults.ToolDenylistAdd, "rm -rf /", dot)
	}
}
