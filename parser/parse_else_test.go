package parser

import (
	"strings"
	"testing"
)

// buildElseDip produces a workflow whose edges block contains the given lines
// (each already indented relative to the block), so a section-level `else`
// default can be exercised.
func buildElseDip(edgeLines string) string {
	return "workflow X\n" +
		"  goal: \"Test else\"\n" +
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
		"  edges\n" + edgeLines
}

// TestParseElseTarget: a section-level `else -> X` populates Workflow.ElseTarget
// and does not create a normal edge.
func TestParseElseTarget(t *testing.T) {
	src := buildElseDip("    A -> B when ctx.outcome = success\n    else -> C\n")
	p := NewParser(src, "test.dip")
	w, err := p.Parse()
	if err != nil {
		t.Fatalf("unexpected parse error: %v (%v)", err, p.Diagnostics())
	}
	if diags := p.Diagnostics(); len(diags) > 0 {
		t.Fatalf("valid else workflow produced parser diagnostics: %v", diags)
	}
	if w.ElseTarget != "C" {
		t.Fatalf("ElseTarget = %q, want %q", w.ElseTarget, "C")
	}
	for _, e := range w.Edges {
		if e.From == "else" || e.To == "else" {
			t.Fatalf("else produced a spurious edge: %+v", e)
		}
	}
	if len(w.Edges) != 1 {
		t.Fatalf("got %d edges, want 1 (the A -> B edge only)", len(w.Edges))
	}
}

// TestParseElseDuplicate: a second `else` in the same edges block is a diagnostic.
func TestParseElseDuplicate(t *testing.T) {
	src := buildElseDip("    A -> B when ctx.outcome = success\n    else -> C\n    else -> B\n")
	p := NewParser(src, "test.dip")
	_, _ = p.Parse()
	diags := strings.Join(p.Diagnostics(), "\n")
	if !strings.Contains(diags, "else") {
		t.Fatalf("expected a duplicate-else diagnostic, got: %q", diags)
	}
}

// TestParseElseMissingTarget: `else ->` with no destination is a clear parse
// diagnostic, and must not desync the parser by swallowing the newline as the
// target (which would leak a confusing downstream DIP003).
func TestParseElseMissingTarget(t *testing.T) {
	src := buildElseDip("    A -> B when ctx.outcome = success\n    else ->\n    B -> C\n")
	p := NewParser(src, "test.dip")
	w, _ := p.Parse()
	diags := strings.Join(p.Diagnostics(), "\n")
	if !strings.Contains(diags, "else") || !strings.Contains(diags, "target") {
		t.Fatalf("expected an else-missing-target diagnostic, got: %q", diags)
	}
	if w.ElseTarget != "" {
		t.Fatalf("ElseTarget should be empty on missing target, got %q", w.ElseTarget)
	}
	// Parser must have recovered: the following B -> C edge still parses.
	foundBC := false
	for _, e := range w.Edges {
		if e.From == "B" && e.To == "C" {
			foundBC = true
		}
	}
	if !foundBC {
		t.Fatalf("parser desynced after `else ->`; B -> C edge missing. Edges: %+v", w.Edges)
	}
}

// TestParseElseMissingArrow: `else <node>` (missing the arrow) gives one clear
// diagnostic about the missing `->`, does NOT consume the target and then
// misreport it as a missing target, leaves ElseTarget empty, and recovers so
// the following edge still parses.
func TestParseElseMissingArrow(t *testing.T) {
	src := buildElseDip("    A -> B when ctx.outcome = success\n    else C\n    B -> C\n")
	p := NewParser(src, "test.dip")
	w, _ := p.Parse()
	diags := strings.Join(p.Diagnostics(), "\n")
	if !strings.Contains(diags, "else") || !strings.Contains(diags, "->") {
		t.Fatalf("expected an else-missing-arrow diagnostic mentioning `->`, got: %q", diags)
	}
	// The target WAS present, so the misleading "requires a target node" message
	// must not appear.
	if strings.Contains(diags, "requires a target node") {
		t.Fatalf("misleading missing-target diagnostic on a missing-arrow line: %q", diags)
	}
	if w.ElseTarget != "" {
		t.Fatalf("ElseTarget should be empty on malformed else, got %q", w.ElseTarget)
	}
	foundBC := false
	for _, e := range w.Edges {
		if e.From == "B" && e.To == "C" {
			foundBC = true
		}
	}
	if !foundBC {
		t.Fatalf("parser desynced after `else C`; B -> C edge missing. Edges: %+v", w.Edges)
	}
}
