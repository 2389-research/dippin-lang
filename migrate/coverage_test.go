package migrate

import (
	"strings"
	"testing"
	"time"

	"github.com/2389-research/dippin-lang/ir"
)

// --- addOrUpdateNode ---

func TestAddOrUpdateNode_NewNode(t *testing.T) {
	p := &parser{
		graph: &dotGraph{},
	}
	attrs := map[string]string{"shape": "box", "label": "Test"}
	p.addOrUpdateNode("A", attrs)
	if len(p.graph.Nodes) != 1 {
		t.Fatalf("nodes = %d, want 1", len(p.graph.Nodes))
	}
	if p.graph.Nodes[0].ID != "A" {
		t.Errorf("ID = %q, want A", p.graph.Nodes[0].ID)
	}
	if p.graph.Nodes[0].Attrs["shape"] != "box" {
		t.Errorf("shape = %q, want box", p.graph.Nodes[0].Attrs["shape"])
	}
}

func TestAddOrUpdateNode_MergeAttrs(t *testing.T) {
	p := &parser{
		graph: &dotGraph{
			Nodes: []dotNode{
				{ID: "A", Attrs: map[string]string{"shape": "box"}},
			},
		},
	}
	p.addOrUpdateNode("A", map[string]string{"label": "Updated"})
	if len(p.graph.Nodes) != 1 {
		t.Fatalf("nodes = %d, want 1", len(p.graph.Nodes))
	}
	if p.graph.Nodes[0].Attrs["shape"] != "box" {
		t.Error("original shape attr should be preserved")
	}
	if p.graph.Nodes[0].Attrs["label"] != "Updated" {
		t.Errorf("label = %q, want Updated", p.graph.Nodes[0].Attrs["label"])
	}
}

// --- updateParenDepth ---

func TestUpdateParenDepth(t *testing.T) {
	tests := []struct {
		ch    byte
		depth int
		want  int
	}{
		{'(', 0, 1},
		{'(', 2, 3},
		{')', 1, 0},
		{')', 3, 2},
		{'a', 0, 0},
		{'a', 5, 5},
	}
	for _, tt := range tests {
		got := updateParenDepth(tt.ch, tt.depth)
		if got != tt.want {
			t.Errorf("updateParenDepth(%q, %d) = %d, want %d", tt.ch, tt.depth, got, tt.want)
		}
	}
}

// --- trySplitAt edge cases ---

func TestTrySplitAt_EmptyLeft(t *testing.T) {
	// Split at position 0 → empty left side.
	_, ok := trySplitAt("&&right", 0, 2)
	if ok {
		t.Error("expected false for empty left")
	}
}

func TestTrySplitAt_EmptyRight(t *testing.T) {
	_, ok := trySplitAt("left&&", 4, 2)
	if ok {
		t.Error("expected false for empty right")
	}
}

func TestTrySplitAt_Valid(t *testing.T) {
	parts, ok := trySplitAt("a && b", 2, 2)
	if !ok {
		t.Fatal("expected true for valid split")
	}
	if parts[0] != "a" || parts[1] != "b" {
		t.Errorf("parts = %v, want [a, b]", parts)
	}
}

// --- splitLogicalOp with parens ---

func TestSplitLogicalOp_RespectParens(t *testing.T) {
	// The || is inside parens, so it shouldn't split.
	_, ok := splitLogicalOp("(a || b)", "||")
	if ok {
		t.Error("expected no split when || is inside parens")
	}
}

func TestSplitLogicalOp_OutsideParens(t *testing.T) {
	parts, ok := splitLogicalOp("a || b", "||")
	if !ok {
		t.Fatal("expected split for top-level ||")
	}
	if parts[0] != "a" || parts[1] != "b" {
		t.Errorf("parts = %v", parts)
	}
}

// --- parseCondExpr edge cases ---

func TestParseCondExpr_Empty(t *testing.T) {
	_, err := parseCondExpr("")
	if err == nil {
		t.Fatal("expected error for empty expression")
	}
}

func TestParseCondExpr_NotPrefix(t *testing.T) {
	expr, err := parseCondExpr("not outcome=success")
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	_, ok := expr.(ir.CondNot)
	if !ok {
		t.Errorf("expected CondNot, got %T", expr)
	}
}

func TestParseCondExpr_ExclamationNot(t *testing.T) {
	expr, err := parseCondExpr("!outcome=success")
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	_, ok := expr.(ir.CondNot)
	if !ok {
		t.Errorf("expected CondNot, got %T", expr)
	}
}

func TestParseCondExpr_OrExpression(t *testing.T) {
	expr, err := parseCondExpr("a=1 || b=2")
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	_, ok := expr.(ir.CondOr)
	if !ok {
		t.Errorf("expected CondOr, got %T", expr)
	}
}

func TestParseCondExpr_AndExpression(t *testing.T) {
	expr, err := parseCondExpr("a=1 && b=2")
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	_, ok := expr.(ir.CondAnd)
	if !ok {
		t.Errorf("expected CondAnd, got %T", expr)
	}
}

// --- parseComparison edge cases ---

func TestParseComparison_NotEqual(t *testing.T) {
	expr, err := parseComparison("outcome!=fail")
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	cmp := expr.(ir.CondCompare)
	if cmp.Op != "!=" {
		t.Errorf("Op = %q, want !=", cmp.Op)
	}
}

func TestParseComparison_Contains(t *testing.T) {
	expr, err := parseComparison("output contains error")
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	cmp := expr.(ir.CondCompare)
	if cmp.Op != "contains" {
		t.Errorf("Op = %q, want contains", cmp.Op)
	}
}

func TestParseComparison_StartsWith(t *testing.T) {
	expr, err := parseComparison("output startswith pass")
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	cmp := expr.(ir.CondCompare)
	if cmp.Op != "startswith" {
		t.Errorf("Op = %q, want startswith", cmp.Op)
	}
}

func TestParseComparison_EndsWith(t *testing.T) {
	expr, err := parseComparison("output endswith done")
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	cmp := expr.(ir.CondCompare)
	if cmp.Op != "endswith" {
		t.Errorf("Op = %q, want endswith", cmp.Op)
	}
}

func TestParseComparison_InOperator(t *testing.T) {
	expr, err := parseComparison("status in success,fail")
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	cmp := expr.(ir.CondCompare)
	if cmp.Op != "in" {
		t.Errorf("Op = %q, want in", cmp.Op)
	}
}

func TestParseComparison_Unparseable(t *testing.T) {
	_, err := parseComparison("just_a_word")
	if err == nil {
		t.Fatal("expected error for unparseable comparison")
	}
}

// --- addNamespacePrefix ---

func TestAddNamespacePrefix_Bare(t *testing.T) {
	if got := addNamespacePrefix("outcome"); got != "ctx.outcome" {
		t.Errorf("got %q, want ctx.outcome", got)
	}
}

func TestAddNamespacePrefix_AlreadyNamespaced(t *testing.T) {
	if got := addNamespacePrefix("ctx.outcome"); got != "ctx.outcome" {
		t.Errorf("got %q, want ctx.outcome", got)
	}
}

func TestAddNamespacePrefix_LegacyContext(t *testing.T) {
	if got := addNamespacePrefix("context.outcome"); got != "ctx.outcome" {
		t.Errorf("got %q, want ctx.outcome", got)
	}
}

// --- formatCondExpr edge cases ---

func TestFormatCondExpr_And(t *testing.T) {
	expr := ir.CondAnd{
		Left:  ir.CondCompare{Variable: "ctx.a", Op: "=", Value: "1"},
		Right: ir.CondCompare{Variable: "ctx.b", Op: "=", Value: "2"},
	}
	s := formatCondExpr(expr)
	if !strings.Contains(s, "and") {
		t.Errorf("expected 'and', got %q", s)
	}
}

func TestFormatCondExpr_Or(t *testing.T) {
	expr := ir.CondOr{
		Left:  ir.CondCompare{Variable: "ctx.a", Op: "=", Value: "1"},
		Right: ir.CondCompare{Variable: "ctx.b", Op: "=", Value: "2"},
	}
	s := formatCondExpr(expr)
	if !strings.Contains(s, "or") {
		t.Errorf("expected 'or', got %q", s)
	}
}

func TestFormatCondExpr_Not(t *testing.T) {
	expr := ir.CondNot{Inner: ir.CondCompare{Variable: "ctx.a", Op: "=", Value: "1"}}
	s := formatCondExpr(expr)
	if !strings.HasPrefix(s, "not ") {
		t.Errorf("expected 'not ' prefix, got %q", s)
	}
}

func TestFormatCondExpr_Nil(t *testing.T) {
	s := formatCondExpr(nil)
	if s != "" {
		t.Errorf("expected empty, got %q", s)
	}
}

// --- formatBinaryOp with parentheses ---

func TestFormatBinaryOp_WithParens(t *testing.T) {
	// When parentPrec != 0 and != prec, parens should be added.
	left := ir.CondCompare{Variable: "ctx.a", Op: "=", Value: "1"}
	right := ir.CondCompare{Variable: "ctx.b", Op: "=", Value: "2"}
	s := formatBinaryOp(left, right, "or", condPrecOr, condPrecAnd)
	if !strings.HasPrefix(s, "(") {
		t.Errorf("expected parens, got %q", s)
	}
}

// --- compareAgentConfigs / compareToolConfigs / compareHumanConfigs type mismatch ---

func TestCompareAgentConfigs_TypeMismatch(t *testing.T) {
	ac := ir.AgentConfig{Prompt: "test", Model: "o1"}
	diffs := compareAgentConfigs("A", "node:A", ac, ir.ToolConfig{Command: "echo"})
	if len(diffs) != 1 {
		t.Fatalf("expected 1 diff, got %d", len(diffs))
	}
	if diffs[0].Kind != "config_mismatch" {
		t.Errorf("kind = %q, want config_mismatch", diffs[0].Kind)
	}
}

func TestCompareToolConfigs_TypeMismatch(t *testing.T) {
	tc := ir.ToolConfig{Command: "echo"}
	diffs := compareToolConfigs("A", "node:A", tc, ir.AgentConfig{Prompt: "test"})
	if len(diffs) != 1 {
		t.Fatalf("expected 1 diff, got %d", len(diffs))
	}
}

func TestCompareHumanConfigs_TypeMismatch(t *testing.T) {
	hc := ir.HumanConfig{Mode: "choice"}
	diffs := compareHumanConfigs("A", "node:A", hc, ir.AgentConfig{Prompt: "test"})
	if len(diffs) != 1 {
		t.Fatalf("expected 1 diff, got %d", len(diffs))
	}
}

func TestCompareParallelConfigs_TypeMismatch(t *testing.T) {
	pc := ir.ParallelConfig{Targets: []string{"A"}}
	diffs := compareParallelConfigs("P", "node:P", pc, ir.AgentConfig{})
	if len(diffs) != 1 {
		t.Fatalf("expected 1 diff, got %d", len(diffs))
	}
}

func TestCompareFanInConfigs_TypeMismatch(t *testing.T) {
	fc := ir.FanInConfig{Sources: []string{"A"}}
	diffs := compareFanInConfigs("J", "node:J", fc, ir.AgentConfig{})
	if len(diffs) != 1 {
		t.Fatalf("expected 1 diff, got %d", len(diffs))
	}
}

func TestCompareSubgraphConfigs_TypeMismatch(t *testing.T) {
	sc := ir.SubgraphConfig{Ref: "test.dip"}
	diffs := compareSubgraphConfigs("S", "node:S", sc, ir.AgentConfig{})
	if len(diffs) != 1 {
		t.Fatalf("expected 1 diff, got %d", len(diffs))
	}
}

// --- Matching configs (no diffs) ---

func TestCompareAgentConfigs_MatchingPrompts(t *testing.T) {
	ac := ir.AgentConfig{Prompt: "test", Model: "o1", Provider: "openai"}
	diffs := compareAgentConfigs("A", "node:A", ac, ac)
	if len(diffs) != 0 {
		t.Errorf("expected 0 diffs for matching configs, got %d", len(diffs))
	}
}

func TestCompareToolConfigs_MatchingCommand(t *testing.T) {
	tc := ir.ToolConfig{Command: "echo hello"}
	diffs := compareToolConfigs("T", "node:T", tc, tc)
	if len(diffs) != 0 {
		t.Errorf("expected 0 diffs, got %d", len(diffs))
	}
}

func TestCompareToolConfigs_DifferentCommand(t *testing.T) {
	tc1 := ir.ToolConfig{Command: "echo hello"}
	tc2 := ir.ToolConfig{Command: "echo goodbye"}
	diffs := compareToolConfigs("T", "node:T", tc1, tc2)
	if len(diffs) != 1 {
		t.Fatalf("expected 1 diff, got %d", len(diffs))
	}
}

func TestCompareToolConfigsDifferentMarkerGrep(t *testing.T) {
	a := ir.ToolConfig{MarkerGrep: "^pass$"}
	b := ir.ToolConfig{MarkerGrep: "^fail$"}
	diffs := compareToolConfigs("T", "", a, b)
	if len(diffs) == 0 {
		t.Error("expected difference for MarkerGrep, got none")
	}
}

func TestCompareToolConfigsDifferentRouteRequired(t *testing.T) {
	a := ir.ToolConfig{RouteRequired: true}
	b := ir.ToolConfig{RouteRequired: false}
	diffs := compareToolConfigs("T", "", a, b)
	if len(diffs) == 0 {
		t.Error("expected difference for RouteRequired, got none")
	}
}

func TestCompareToolConfigsDifferentOutputLimit(t *testing.T) {
	a := ir.ToolConfig{OutputLimit: 1024}
	b := ir.ToolConfig{OutputLimit: 2048}
	diffs := compareToolConfigs("T", "", a, b)
	if len(diffs) == 0 {
		t.Error("expected difference for OutputLimit, got none")
	}
}

func TestCompareToolConfigsDifferentTimeout(t *testing.T) {
	a := ir.ToolConfig{Timeout: 30 * time.Second}
	b := ir.ToolConfig{Timeout: 60 * time.Second}
	diffs := compareToolConfigs("T", "", a, b)
	if len(diffs) == 0 {
		t.Error("expected difference for Timeout, got none")
	}
}

func TestCompareToolConfigsDifferentOutputs(t *testing.T) {
	a := ir.ToolConfig{Outputs: []string{"pass", "fail"}}
	b := ir.ToolConfig{Outputs: []string{"green", "red"}}
	diffs := compareToolConfigs("T", "", a, b)
	if len(diffs) == 0 {
		t.Error("expected difference for Outputs, got none")
	}
}

func TestCompareHumanConfigs_Matching(t *testing.T) {
	hc := ir.HumanConfig{Mode: "choice"}
	diffs := compareHumanConfigs("H", "node:H", hc, hc)
	if len(diffs) != 0 {
		t.Errorf("expected 0 diffs, got %d", len(diffs))
	}
}

func TestCompareHumanConfigs_DifferentMode(t *testing.T) {
	hc1 := ir.HumanConfig{Mode: "choice"}
	hc2 := ir.HumanConfig{Mode: "freeform"}
	diffs := compareHumanConfigs("H", "node:H", hc1, hc2)
	if len(diffs) != 1 {
		t.Fatalf("expected 1 diff, got %d", len(diffs))
	}
}

func TestCompareParallelConfigs_DifferentTargets(t *testing.T) {
	pc1 := ir.ParallelConfig{Targets: []string{"A", "B"}}
	pc2 := ir.ParallelConfig{Targets: []string{"A", "C"}}
	diffs := compareParallelConfigs("P", "node:P", pc1, pc2)
	if len(diffs) != 1 {
		t.Fatalf("expected 1 diff, got %d", len(diffs))
	}
}

func TestCompareFanInConfigs_DifferentSources(t *testing.T) {
	fc1 := ir.FanInConfig{Sources: []string{"A"}}
	fc2 := ir.FanInConfig{Sources: []string{"B"}}
	diffs := compareFanInConfigs("J", "node:J", fc1, fc2)
	if len(diffs) != 1 {
		t.Fatalf("expected 1 diff, got %d", len(diffs))
	}
}

func TestCompareSubgraphConfigs_DifferentRef(t *testing.T) {
	sc1 := ir.SubgraphConfig{Ref: "a.dip"}
	sc2 := ir.SubgraphConfig{Ref: "b.dip"}
	diffs := compareSubgraphConfigs("S", "node:S", sc1, sc2)
	if len(diffs) != 1 {
		t.Fatalf("expected 1 diff, got %d", len(diffs))
	}
}

// --- compareStructuralConfigs ---

func TestCompareStructuralConfigs_Parallel(t *testing.T) {
	diffs := compareStructuralConfigs("P", "node:P",
		ir.ParallelConfig{Targets: []string{"A"}},
		ir.ParallelConfig{Targets: []string{"A"}})
	if len(diffs) != 0 {
		t.Errorf("expected 0 diffs, got %d", len(diffs))
	}
}

func TestCompareStructuralConfigs_FanIn(t *testing.T) {
	diffs := compareStructuralConfigs("J", "node:J",
		ir.FanInConfig{Sources: []string{"A"}},
		ir.FanInConfig{Sources: []string{"A"}})
	if len(diffs) != 0 {
		t.Errorf("expected 0 diffs, got %d", len(diffs))
	}
}

func TestCompareStructuralConfigs_Subgraph(t *testing.T) {
	diffs := compareStructuralConfigs("S", "node:S",
		ir.SubgraphConfig{Ref: "a.dip"},
		ir.SubgraphConfig{Ref: "a.dip"})
	if len(diffs) != 0 {
		t.Errorf("expected 0 diffs, got %d", len(diffs))
	}
}

func TestCompareStructuralConfigs_UnknownType(t *testing.T) {
	// AgentConfig is not a structural config, should return nil.
	diffs := compareStructuralConfigs("A", "node:A",
		ir.AgentConfig{Prompt: "test"},
		ir.AgentConfig{Prompt: "test"})
	if len(diffs) != 0 {
		t.Errorf("expected 0 diffs for unknown type, got %d", len(diffs))
	}
}

// --- DOT lexer edge cases ---

func TestLexer_Peek_AtEOF(t *testing.T) {
	l := newLexer("")
	ch := l.peek()
	if ch != 0 {
		t.Errorf("expected 0, got %d", ch)
	}
}

func TestTokenKindName_Unknown(t *testing.T) {
	name := tokenKindName(tokenKind(999))
	if name != "unknown" {
		t.Errorf("expected 'unknown', got %q", name)
	}
}

func TestLexDashOrArrow_Arrow(t *testing.T) {
	// An arrow -> should be lexed as tokArrow.
	input := `digraph G { A -> B; }`
	g, err := parseDOT(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(g.Edges) != 1 {
		t.Errorf("edges = %d, want 1", len(g.Edges))
	}
}

// --- DOT writeEscapeChar edge cases ---

func TestDOTParseUnknownEscape(t *testing.T) {
	// \z is not a known escape → should be written as \z.
	input := `digraph G { A [label="test\zvalue"]; }`
	g, err := parseDOT(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var nodeA *dotNode
	for i := range g.Nodes {
		if g.Nodes[i].ID == "A" {
			nodeA = &g.Nodes[i]
		}
	}
	if nodeA == nil {
		t.Fatal("node A not found")
	}
	if !strings.Contains(nodeA.Attrs["label"], "\\z") {
		t.Errorf("label = %q, expected \\z escape preserved", nodeA.Attrs["label"])
	}
}

func TestDOTParseEscapeR(t *testing.T) {
	// \r should be ignored (mapped to 0).
	input := `digraph G { A [label="line\rend"]; }`
	g, err := parseDOT(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	label := g.Nodes[0].Attrs["label"]
	if strings.Contains(label, "\r") {
		t.Errorf("label should not contain \\r, got %q", label)
	}
}

func TestDOTParseEscapeL(t *testing.T) {
	// \l is a DOT left-justified newline, mapped to \n.
	input := `digraph G { A [label="line1\lline2"]; }`
	g, err := parseDOT(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	label := g.Nodes[0].Attrs["label"]
	if !strings.Contains(label, "\n") {
		t.Errorf("label should contain \\n from \\l, got %q", label)
	}
}

// --- DOT parser: parseStatement edge cases ---

func TestDOTParseStatement_DefaultEdge(t *testing.T) {
	// "edge" default statement.
	input := `digraph G { edge [style=bold]; A -> B; }`
	_, err := parseDOT(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDOTParseStatement_Subgraph(t *testing.T) {
	// Subgraph keyword should be skipped without error.
	input := `digraph G { subgraph cluster_0 { A; } }`
	_, err := parseDOT(input)
	// May or may not error — just check no panic.
	_ = err
}

// --- applyModelField / applyProviderField legacy aliases ---

func TestApplyModelField_LlmModel(t *testing.T) {
	cfg := &ir.AgentConfig{}
	attrs := map[string]string{"llm_model": "gpt-4"}
	applyModelField(cfg, attrs)
	if cfg.Model != "gpt-4" {
		t.Errorf("model = %q, want gpt-4", cfg.Model)
	}
}

func TestApplyProviderField_LlmProvider(t *testing.T) {
	cfg := &ir.AgentConfig{}
	attrs := map[string]string{"llm_provider": "openai"}
	applyProviderField(cfg, attrs)
	if cfg.Provider != "openai" {
		t.Errorf("provider = %q, want openai", cfg.Provider)
	}
}

// --- buildOtherConfig ---

func TestBuildOtherConfig_Fallback(t *testing.T) {
	// Unknown kind should get a fallback AgentConfig.
	cfg := buildOtherConfig(ir.NodeKind("unknown"), map[string]string{})
	if _, ok := cfg.(ir.AgentConfig); !ok {
		t.Errorf("expected AgentConfig fallback, got %T", cfg)
	}
}

// --- convertDOTGraph error in edges ---

func TestMigrate_EmptyCondition(t *testing.T) {
	input := `digraph G {
		A -> B [condition=""];
	}`
	// Empty condition should not error (nil condition).
	_, err := Migrate(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// --- readKeyValuePair edge: DOT attrs with no value ---

func TestDOTParse_AttrWithNoValue(t *testing.T) {
	// An attr key followed by a closing bracket should be handled.
	input := `digraph G { A [shape]; }`
	_, err := parseDOT(input)
	// Should produce an error since shape has no = value.
	if err == nil {
		t.Fatal("expected error for attr without value")
	}
}

// --- compareAgentConfigs with field diffs ---

func TestCompareAgentConfigs_FieldDiffs(t *testing.T) {
	ac := ir.AgentConfig{Prompt: "a", Model: "o1", Provider: "p", GoalGate: true, AutoStatus: true}
	bc := ir.AgentConfig{Prompt: "b", Model: "o2", Provider: "q", GoalGate: false, AutoStatus: false}
	diffs := compareAgentConfigs("A", "node:A", ac, bc)
	// Should see: prompt, model, provider, goal_gate, auto_status.
	if len(diffs) < 5 {
		t.Errorf("expected at least 5 diffs, got %d", len(diffs))
	}
}

func TestCompareParallelConfigs_DifferentBranches(t *testing.T) {
	a := ir.ParallelConfig{
		Targets:  []string{"A"},
		Branches: []ir.BranchConfig{{Target: "A", Model: "x"}},
	}
	b := ir.ParallelConfig{
		Targets:  []string{"A"},
		Branches: []ir.BranchConfig{{Target: "A", Model: "y"}},
	}
	diffs := compareParallelConfigs("P", "nodes.P", a, b)
	if len(diffs) == 0 {
		t.Fatal("expected a branches difference, got none")
	}
}
