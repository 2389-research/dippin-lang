package lsp

import (
	"strings"
	"testing"

	"go.lsp.dev/protocol"

	"github.com/2389-research/dippin-lang/ir"
	"github.com/2389-research/dippin-lang/validator"
)

const testDipContent = `workflow Test
  goal: "Test workflow"
  start: Ask
  exit: Done

  human Ask
    mode: freeform

  agent Done
    prompt:
      Complete the task.

  edges
    Ask -> Done
`

func TestDocumentStore_OpenAndGet(t *testing.T) {
	store := newDocumentStore()
	doc := store.open("file:///test.dip", testDipContent, 1)

	if doc.Parsed == nil {
		t.Fatalf("expected parsed workflow, got nil (err: %v)", doc.Err)
	}
	if doc.Parsed.Name != "Test" {
		t.Errorf("expected workflow name Test, got %s", doc.Parsed.Name)
	}

	got := store.get("file:///test.dip")
	if got == nil {
		t.Fatal("expected document from get, got nil")
	}
	if got.Version != 1 {
		t.Errorf("expected version 1, got %d", got.Version)
	}
}

func TestDocumentStore_Update(t *testing.T) {
	store := newDocumentStore()
	store.open("file:///test.dip", testDipContent, 1)

	updated := `workflow Updated
  goal: "Updated"
  start: A
  exit: B

  agent A
    prompt:
      Do A.

  agent B
    prompt:
      Do B.

  edges
    A -> B
`
	doc := store.update("file:///test.dip", updated, 2)
	if doc.Parsed == nil {
		t.Fatalf("expected parsed workflow after update, got nil")
	}
	if doc.Parsed.Name != "Updated" {
		t.Errorf("expected workflow name Updated, got %s", doc.Parsed.Name)
	}
}

func TestDocumentStore_Close(t *testing.T) {
	store := newDocumentStore()
	store.open("file:///test.dip", testDipContent, 1)
	store.close("file:///test.dip")

	if store.get("file:///test.dip") != nil {
		t.Error("expected nil after close")
	}
}

func TestDocumentStore_GetMissing(t *testing.T) {
	store := newDocumentStore()
	if store.get("file:///missing.dip") != nil {
		t.Error("expected nil for missing document")
	}
}

func TestCollectDiagnostics_ValidDocument(t *testing.T) {
	store := newDocumentStore()
	doc := store.open("file:///test.dip", testDipContent, 1)

	diags := collectDiagnostics(doc)
	for _, d := range diags {
		if d.Severity == protocol.DiagnosticSeverityError {
			t.Errorf("unexpected error diagnostic: %s", d.Message)
		}
	}
}

func TestCollectDiagnostics_NilParsed(t *testing.T) {
	doc := &document{URI: "file:///bad.dip", Err: nil, Parsed: nil}
	diags := collectDiagnostics(doc)
	if len(diags) != 0 {
		t.Errorf("expected no diagnostics for nil parsed without error, got %d", len(diags))
	}
}

func TestMapSeverity(t *testing.T) {
	tests := []struct {
		input validator.Severity
		want  protocol.DiagnosticSeverity
	}{
		{validator.SeverityError, protocol.DiagnosticSeverityError},
		{validator.SeverityWarning, protocol.DiagnosticSeverityWarning},
		{validator.SeverityInfo, protocol.DiagnosticSeverityInformation},
		{validator.SeverityHint, protocol.DiagnosticSeverityHint},
	}
	for _, tt := range tests {
		got := mapSeverity(tt.input)
		if got != tt.want {
			t.Errorf("mapSeverity(%v) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestSourceRange(t *testing.T) {
	d := validator.Diagnostic{
		Location: ir.SourceLocation{Line: 5, Column: 2, EndLine: 5, EndColumn: 10},
	}
	r := sourceRange(d)
	if r.Start.Line != 4 {
		t.Errorf("expected start line 4 (0-indexed), got %d", r.Start.Line)
	}
	if r.Start.Character != 2 {
		t.Errorf("expected start col 2, got %d", r.Start.Character)
	}
}

func TestSourceRange_ZeroLine(t *testing.T) {
	d := validator.Diagnostic{
		Location: ir.SourceLocation{Line: 0, Column: 0},
	}
	r := sourceRange(d)
	if r.Start.Line != 0 {
		t.Errorf("expected start line 0, got %d", r.Start.Line)
	}
}

func TestWordAtPosition(t *testing.T) {
	content := "  Ask -> Done\n  Done -> Exit\n"

	tests := []struct {
		line, col uint32
		want      string
	}{
		{0, 2, "Ask"},
		{0, 9, "Done"},
		{1, 2, "Done"},
	}
	for _, tt := range tests {
		got := wordAtPosition(content, protocol.Position{Line: tt.line, Character: tt.col})
		if got != tt.want {
			t.Errorf("wordAtPosition(%d,%d) = %q, want %q", tt.line, tt.col, got, tt.want)
		}
	}
}

func TestWordAtPosition_OutOfBounds(t *testing.T) {
	content := "hello\nworld"
	got := wordAtPosition(content, protocol.Position{Line: 99, Character: 0})
	if got != "" {
		t.Errorf("expected empty for out-of-bounds line, got %q", got)
	}

	got = wordAtPosition(content, protocol.Position{Line: 0, Character: 99})
	if got != "" {
		t.Errorf("expected empty for out-of-bounds col, got %q", got)
	}
}

func TestExtractWord(t *testing.T) {
	tests := []struct {
		line string
		col  int
		want string
	}{
		{"hello world", 0, "hello"},
		{"hello world", 6, "world"},
		{"  -> ", 2, ""},
		{"foo_bar", 3, "foo_bar"},
	}
	for _, tt := range tests {
		got := extractWord(tt.line, tt.col)
		if got != tt.want {
			t.Errorf("extractWord(%q, %d) = %q, want %q", tt.line, tt.col, got, tt.want)
		}
	}
}

func TestIsWordChar(t *testing.T) {
	for _, c := range "abcABC012_" {
		if !isWordChar(byte(c)) {
			t.Errorf("expected isWordChar(%c) = true", c)
		}
	}
	for _, c := range " ->:.\n" {
		if isWordChar(byte(c)) {
			t.Errorf("expected isWordChar(%c) = false", c)
		}
	}
}

func TestFindNodeAtPosition(t *testing.T) {
	store := newDocumentStore()
	doc := store.open("file:///test.dip", testDipContent, 1)

	if doc.Parsed == nil {
		t.Skip("parse failed")
	}

	for _, n := range doc.Parsed.Nodes {
		if n.Source.Line > 0 {
			found := findNodeAtPosition(doc, protocol.Position{
				Line:      uint32(n.Source.Line - 1),
				Character: 0,
			})
			if found == nil {
				t.Errorf("expected to find node %s at line %d", n.ID, n.Source.Line)
			}
		}
	}
}

func TestFindNodeAtPosition_NoMatch(t *testing.T) {
	store := newDocumentStore()
	doc := store.open("file:///test.dip", testDipContent, 1)

	if doc.Parsed == nil {
		t.Skip("parse failed")
	}

	found := findNodeAtPosition(doc, protocol.Position{Line: 0, Character: 0})
	if found != nil {
		t.Errorf("expected nil for line 0 (workflow header), got %s", found.ID)
	}
}

func TestNodeContainsLine(t *testing.T) {
	n := &ir.Node{Source: ir.SourceLocation{Line: 5, EndLine: 8}}
	if !nodeContainsLine(n, 5) {
		t.Error("expected line 5 to be in range")
	}
	if !nodeContainsLine(n, 7) {
		t.Error("expected line 7 to be in range")
	}
	if nodeContainsLine(n, 9) {
		t.Error("expected line 9 to be out of range")
	}
}

func TestNodeContainsLine_ZeroSource(t *testing.T) {
	n := &ir.Node{Source: ir.SourceLocation{}}
	if nodeContainsLine(n, 1) {
		t.Error("expected false for zero source location")
	}
}

func TestFormatNodeHover(t *testing.T) {
	w := &ir.Workflow{Defaults: ir.WorkflowDefaults{Model: "claude-sonnet-4-6", Provider: "anthropic"}}
	n := &ir.Node{
		ID:     "Coder",
		Kind:   ir.NodeAgent,
		Label:  "Code Writer",
		Config: ir.AgentConfig{Prompt: "Write code", Model: "claude-opus-4-6"},
	}

	result := formatNodeHover(n, w)
	if result == "" {
		t.Error("expected non-empty hover content")
	}
}

func TestFormatNodeConfig_Human(t *testing.T) {
	n := &ir.Node{Config: ir.HumanConfig{Mode: "choice"}}
	result := formatNodeConfig(n, nil)
	if result != "Mode: choice\n" {
		t.Errorf("expected 'Mode: choice\\n', got %q", result)
	}
}

func TestFormatNodeConfig_Tool(t *testing.T) {
	n := &ir.Node{Config: ir.ToolConfig{Command: "echo hello"}}
	result := formatNodeConfig(n, nil)
	if result == "" {
		t.Error("expected non-empty tool config hover")
	}
}

func TestFormatToolHoverWithRoutingFields(t *testing.T) {
	node := &ir.Node{
		Kind: ir.NodeTool,
		Config: ir.ToolConfig{
			Command:       "echo hi",
			MarkerGrep:    "^pass$",
			RouteRequired: true,
			OutputLimit:   8192,
		},
	}
	out := formatNodeConfig(node, &ir.Workflow{})
	for _, want := range []string{"marker_grep", "route_required", "output_limit"} {
		if !strings.Contains(out, want) {
			t.Errorf("hover missing %q. Output:\n%s", want, out)
		}
	}
}

func TestBuildSymbols(t *testing.T) {
	store := newDocumentStore()
	doc := store.open("file:///test.dip", testDipContent, 1)

	if doc.Parsed == nil {
		t.Skip("parse failed")
	}

	symbols := buildSymbols(doc.Parsed)
	if len(symbols) == 0 {
		t.Error("expected at least one symbol")
	}

	hasNode := false
	hasEdge := false
	for _, s := range symbols {
		if s.Kind == protocol.SymbolKindEvent {
			hasEdge = true
		} else {
			hasNode = true
		}
	}
	if !hasNode {
		t.Error("expected node symbols")
	}
	if !hasEdge {
		t.Error("expected edge symbols")
	}
}

func TestNodeSymbolKind(t *testing.T) {
	tests := []struct {
		kind ir.NodeKind
		want protocol.SymbolKind
	}{
		{ir.NodeAgent, protocol.SymbolKindFunction},
		{ir.NodeHuman, protocol.SymbolKindInterface},
		{ir.NodeTool, protocol.SymbolKindMethod},
		{ir.NodeParallel, protocol.SymbolKindStruct},
		{ir.NodeSubgraph, protocol.SymbolKindVariable},
	}
	for _, tt := range tests {
		got := nodeSymbolKind(tt.kind)
		if got != tt.want {
			t.Errorf("nodeSymbolKind(%s) = %v, want %v", tt.kind, got, tt.want)
		}
	}
}

func TestNodeIDCompletions(t *testing.T) {
	store := newDocumentStore()
	doc := store.open("file:///test.dip", testDipContent, 1)

	if doc.Parsed == nil {
		t.Skip("parse failed")
	}

	items := nodeIDCompletions(doc)
	if len(items) < 2 {
		t.Errorf("expected at least 2 completion items, got %d", len(items))
	}
}

func TestFieldCompletions(t *testing.T) {
	items := fieldCompletions()
	if len(items) == 0 {
		t.Error("expected field completion items")
	}
	found := false
	for _, item := range items {
		if item.Label == "prompt:" {
			found = true
		}
	}
	if !found {
		t.Error("expected 'prompt:' in field completions")
	}
}

func TestFieldCompletionsIncludesRoutingFields(t *testing.T) {
	items := fieldCompletions()
	want := map[string]bool{"marker_grep:": false, "route_required:": false, "output_limit:": false}
	for _, it := range items {
		if _, ok := want[it.Label]; ok {
			want[it.Label] = true
		}
	}
	for label, found := range want {
		if !found {
			t.Errorf("missing completion for %q", label)
		}
	}
}

func TestFieldCompletionsIncludesToolAccess(t *testing.T) {
	items := fieldCompletions()
	for _, it := range items {
		if it.Label == "tool_access:" {
			return
		}
	}
	t.Error("missing completion for 'tool_access:' (v0.32.0 safety primitive)")
}

func TestFieldCompletionsIncludesCommandFile(t *testing.T) {
	items := fieldCompletions()
	for _, it := range items {
		if it.Label == "command_file:" {
			return
		}
	}
	t.Error("missing completion for 'command_file:' (v0.33.0 directive)")
}

func TestConvertDiagnostic(t *testing.T) {
	d := validator.Diagnostic{
		Code:     "DIP001",
		Message:  "test error",
		Severity: validator.SeverityError,
		Location: ir.SourceLocation{Line: 3, Column: 4, EndLine: 3, EndColumn: 10},
	}
	got := convertDiagnostic(d)
	if got.Code != "DIP001" {
		t.Errorf("Code = %v, want DIP001", got.Code)
	}
	if got.Message != "test error" {
		t.Errorf("Message = %q, want %q", got.Message, "test error")
	}
	if got.Source != "dippin" {
		t.Errorf("Source = %q, want %q", got.Source, "dippin")
	}
	if got.Severity != protocol.DiagnosticSeverityError {
		t.Errorf("Severity = %v, want Error", got.Severity)
	}
	// Line is 1-indexed in source, 0-indexed in LSP
	if got.Range.Start.Line != 2 {
		t.Errorf("Range.Start.Line = %d, want 2", got.Range.Start.Line)
	}
	if got.Range.Start.Character != 4 {
		t.Errorf("Range.Start.Character = %d, want 4", got.Range.Start.Character)
	}
}

func TestConvertDiagnostic_Warning(t *testing.T) {
	d := validator.Diagnostic{
		Code:     "DIP101",
		Message:  "test warning",
		Severity: validator.SeverityWarning,
		Location: ir.SourceLocation{Line: 1, Column: 0, EndLine: 1, EndColumn: 5},
	}
	got := convertDiagnostic(d)
	if got.Severity != protocol.DiagnosticSeverityWarning {
		t.Errorf("Severity = %v, want Warning", got.Severity)
	}
}

func TestNodeLocation(t *testing.T) {
	n := &ir.Node{
		ID:   "MyNode",
		Kind: ir.NodeAgent,
		Source: ir.SourceLocation{
			Line:   5,
			Column: 2,
		},
	}
	loc := nodeLocation("file:///test.dip", n)
	if loc.URI != "file:///test.dip" {
		t.Errorf("URI = %q, want file:///test.dip", loc.URI)
	}
	// Line is 1-indexed in source, 0-indexed in LSP
	if loc.Range.Start.Line != 4 {
		t.Errorf("Range.Start.Line = %d, want 4", loc.Range.Start.Line)
	}
	if loc.Range.Start.Character != 2 {
		t.Errorf("Range.Start.Character = %d, want 2", loc.Range.Start.Character)
	}
	if loc.Range.End.Line != 4 {
		t.Errorf("Range.End.Line = %d, want 4", loc.Range.End.Line)
	}
}

func TestNodeLocation_ZeroLine(t *testing.T) {
	n := &ir.Node{
		ID:     "Zero",
		Kind:   ir.NodeAgent,
		Source: ir.SourceLocation{Line: 0, Column: 0},
	}
	loc := nodeLocation("file:///test.dip", n)
	if loc.Range.Start.Line != 0 {
		t.Errorf("Range.Start.Line = %d, want 0", loc.Range.Start.Line)
	}
}

func TestZeroRange(t *testing.T) {
	r := zeroRange()
	if r.Start.Line != 0 || r.Start.Character != 0 {
		t.Errorf("Start = (%d,%d), want (0,0)", r.Start.Line, r.Start.Character)
	}
	if r.End.Line != 0 || r.End.Character != 0 {
		t.Errorf("End = (%d,%d), want (0,0)", r.End.Line, r.End.Character)
	}
}

func TestEdgeRange(t *testing.T) {
	e := &ir.Edge{
		From: "A",
		To:   "B",
		Source: ir.SourceLocation{
			Line:   10,
			Column: 4,
		},
	}
	r := edgeRange(e)
	// Line 10 (1-indexed) -> 9 (0-indexed)
	if r.Start.Line != 9 {
		t.Errorf("Start.Line = %d, want 9", r.Start.Line)
	}
	if r.Start.Character != 4 {
		t.Errorf("Start.Character = %d, want 4", r.Start.Character)
	}
}

func TestEdgeRange_ZeroLine(t *testing.T) {
	e := &ir.Edge{
		From:   "A",
		To:     "B",
		Source: ir.SourceLocation{Line: 0, Column: 0},
	}
	r := edgeRange(e)
	if r.Start.Line != 0 {
		t.Errorf("Start.Line = %d, want 0", r.Start.Line)
	}
}

func TestEdgeSymbol_WithLabel(t *testing.T) {
	e := &ir.Edge{
		From:   "A",
		To:     "B",
		Label:  "retry",
		Source: ir.SourceLocation{Line: 5, Column: 2},
	}
	sym := edgeSymbol(e)
	if sym.Name != "A -> B (retry)" {
		t.Errorf("Name = %q, want %q", sym.Name, "A -> B (retry)")
	}
	if sym.Kind != protocol.SymbolKindEvent {
		t.Errorf("Kind = %v, want Event", sym.Kind)
	}
}

func TestIsNoOp(t *testing.T) {
	if !isNoOp("initialized") {
		t.Error("expected initialized to be no-op")
	}
	if !isNoOp("shutdown") {
		t.Error("expected shutdown to be no-op")
	}
	if isNoOp("textDocument/hover") {
		t.Error("expected hover to not be no-op")
	}
}

func TestServerCapabilities(t *testing.T) {
	caps := serverCapabilities()
	if caps.HoverProvider != true {
		t.Error("expected hover provider enabled")
	}
	if caps.DefinitionProvider != true {
		t.Error("expected definition provider enabled")
	}
	if caps.CompletionProvider == nil {
		t.Error("expected completion provider")
	}
	if caps.DocumentSymbolProvider != true {
		t.Error("expected document symbol provider enabled")
	}
}

func TestResolveDefinition_MissingDoc(t *testing.T) {
	s := NewServer()
	params := protocol.DefinitionParams{
		TextDocumentPositionParams: protocol.TextDocumentPositionParams{
			TextDocument: protocol.TextDocumentIdentifier{URI: "file:///missing.dip"},
		},
	}
	loc := s.resolveDefinition(params)
	if loc != nil {
		t.Error("expected nil for missing document")
	}
}

func TestTruncateStr(t *testing.T) {
	if got := truncateStr("hello", 10); got != "hello" {
		t.Errorf("expected hello, got %s", got)
	}
	if got := truncateStr("hello world this is long", 5); got != "hello..." {
		t.Errorf("expected hello..., got %s", got)
	}
}
