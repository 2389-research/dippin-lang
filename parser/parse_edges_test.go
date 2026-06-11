package parser

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// buildEdgeDip produces a minimal three-node workflow whose single edge line
// is the given text (e.g. "A -> B bogus: true").
func buildEdgeDip(edgeLine string) string {
	return "workflow X\n" +
		"  goal: \"Test edges\"\n" +
		"  start: A\n" +
		"  exit: C\n" +
		"\n" +
		"  agent A\n" +
		"    prompt: \"Do A.\"\n" +
		"\n" +
		"  agent B\n" +
		"    prompt: \"Do B.\"\n" +
		"\n" +
		"  agent C\n" +
		"    prompt: \"Do C.\"\n" +
		"\n" +
		"  edges\n" +
		"    " + edgeLine + "\n" +
		"    B -> C\n"
}

// TestParseUnknownEdgeAttributeDiagnoses covers #126(a): an unrecognized edge
// attribute must become a single located parse diagnostic, not be silently
// swallowed, and must not fire once per token (name / ':' / value).
func TestParseUnknownEdgeAttributeDiagnoses(t *testing.T) {
	p := NewParser(buildEdgeDip("A -> B bogus: true"), "test.dip")
	w, err := p.Parse()
	if err == nil {
		t.Fatal("expected parse error for unknown edge attribute, got nil")
	}
	diags := p.Diagnostics()

	count := 0
	for _, d := range diags {
		if strings.Contains(d, "bogus") {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("expected exactly one diagnostic mentioning 'bogus', got %d: %v", count, diags)
	}
	joined := strings.Join(diags, "\n")
	if !strings.Contains(joined, "unknown edge attribute") {
		t.Errorf("expected an 'unknown edge attribute' diagnostic, got: %v", diags)
	}
	// Located: the diagnostic carries a line:column.
	if !strings.Contains(joined, "16:") {
		t.Errorf("expected diagnostic to carry the attribute's line (16), got: %v", diags)
	}

	// Subsequent edges still parse after the diagnostic.
	found := false
	for _, e := range w.Edges {
		if e.From == "B" && e.To == "C" {
			found = true
		}
	}
	if !found {
		t.Error("expected edge B -> C to still parse after unknown-attr diagnostic")
	}
}

// TestParseUnknownBareEdgeAttributeDiagnosesOnce covers an unknown attribute
// with no ": value" payload (a bare word), which must still fire exactly once.
func TestParseUnknownBareEdgeAttributeDiagnosesOnce(t *testing.T) {
	p := NewParser(buildEdgeDip("A -> B bogus"), "test.dip")
	_, err := p.Parse()
	if err == nil {
		t.Fatal("expected parse error for unknown bare edge attribute, got nil")
	}
	count := 0
	for _, d := range p.Diagnostics() {
		if strings.Contains(d, "bogus") {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("expected exactly one diagnostic mentioning 'bogus', got %d: %v", count, p.Diagnostics())
	}
}

// TestParseUnknownEdgeAttributeMissingValue covers the no-value form
// ("bogus:" with nothing after the colon): the diagnostic must fire once and
// the parser must not consume the newline, so the following edge still parses.
func TestParseUnknownEdgeAttributeMissingValue(t *testing.T) {
	p := NewParser(buildEdgeDip("A -> B bogus:"), "test.dip")
	w, err := p.Parse()
	if err == nil {
		t.Fatal("expected parse error for unknown edge attribute, got nil")
	}
	count := 0
	for _, d := range p.Diagnostics() {
		if strings.Contains(d, "bogus") {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("expected exactly one diagnostic mentioning 'bogus', got %d: %v", count, p.Diagnostics())
	}
	found := false
	for _, e := range w.Edges {
		if e.From == "B" && e.To == "C" {
			found = true
		}
	}
	if !found {
		t.Error("expected edge B -> C to still parse after a value-less unknown attr")
	}
}

// TestParseConditionWithBareKeywordRHS covers #126(b): a condition whose bare
// unquoted right-hand value is an attribute keyword must NOT truncate the
// condition. The keyword only terminates a condition when followed by ':'.
func TestParseConditionWithBareKeywordRHS(t *testing.T) {
	for _, kw := range []string{"override", "restart", "label", "weight"} {
		t.Run(kw, func(t *testing.T) {
			p := NewParser(buildEdgeDip("A -> B when ctx.reason = "+kw), "test.dip")
			w, err := p.Parse()
			if err != nil {
				t.Fatalf("unexpected parse error: %v", err)
			}
			var edge = w.Edges[0]
			if edge.Condition == nil {
				t.Fatalf("expected a condition on edge A->B, got nil")
			}
			want := "ctx.reason = " + kw
			if edge.Condition.Raw != want {
				t.Errorf("condition Raw = %q, want %q", edge.Condition.Raw, want)
			}
		})
	}
}

// TestParseKeywordAttributesStillTerminateCondition guards the other side of
// #126(b): a genuine attribute (keyword followed by ':') still terminates the
// condition rather than being absorbed into it.
func TestParseKeywordAttributesStillTerminateCondition(t *testing.T) {
	p := NewParser(buildEdgeDip(`A -> B when ctx.x == "ok" label: approved weight: 5`), "test.dip")
	w, err := p.Parse()
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	edge := w.Edges[0]
	if edge.Condition == nil || edge.Condition.Raw != `ctx.x == "ok"` {
		t.Errorf("condition = %+v, want Raw %q", edge.Condition, `ctx.x == "ok"`)
	}
	if edge.Label != "approved" {
		t.Errorf("label = %q, want %q", edge.Label, "approved")
	}
	if edge.Weight != 5 {
		t.Errorf("weight = %d, want 5", edge.Weight)
	}
}

// TestExistingDipFilesStillParse covers #126(c): the new diagnostics must not
// regress any currently-valid .dip file. Every examples/*.dip and
// parser/testdata/*.dip must still parse without error.
func TestExistingDipFilesStillParse(t *testing.T) {
	dirs := []string{"../examples", "testdata"}
	var files []string
	for _, d := range dirs {
		matches, err := filepath.Glob(filepath.Join(d, "*.dip"))
		if err != nil {
			t.Fatalf("glob %s: %v", d, err)
		}
		files = append(files, matches...)
	}
	if len(files) == 0 {
		t.Fatal("no .dip files found to guard against regression")
	}
	for _, f := range files {
		f := f
		t.Run(filepath.Base(f), func(t *testing.T) {
			src, err := os.ReadFile(f)
			if err != nil {
				t.Fatalf("read %s: %v", f, err)
			}
			p := NewParser(string(src), f)
			if _, err := p.Parse(); err != nil {
				t.Errorf("%s no longer parses: %v", f, err)
			}
		})
	}
}
