package parser

import (
	"strings"
	"testing"

	"github.com/2389-research/dippin-lang/formatter"
)

const versionBody = `workflow Test
  start: A
  exit: B

  agent A
    prompt: "Do A."

  agent B
    prompt: "Do B."

  edges
    A -> B
`

func TestParseDefaultVersionIsOne(t *testing.T) {
	w, err := NewParser(versionBody, "test.dip").Parse()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if w.Version != "1" {
		t.Errorf("version = %q, want %q", w.Version, "1")
	}
}

func TestParseDipDeclaration(t *testing.T) {
	w, err := NewParser("dip 2\n\n"+versionBody, "test.dip").Parse()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if w.Version != "2" {
		t.Errorf("version = %q, want %q", w.Version, "2")
	}
	if w.Name != "Test" {
		t.Errorf("name = %q, want %q", w.Name, "Test")
	}
}

func TestParseDipDeclarationInvalidVersion(t *testing.T) {
	_, err := NewParser("dip x\n\n"+versionBody, "test.dip").Parse()
	if err == nil {
		t.Fatal("expected error for non-integer version, got nil")
	}
	if !strings.Contains(err.Error(), "invalid version declaration") {
		t.Errorf("error = %q, want it to mention invalid version declaration", err.Error())
	}
}

func TestParseDipDeclarationRejectsZero(t *testing.T) {
	// `dip 0` is below the v1 floor; the formatter only emits `dip N` for N > 1,
	// so accepting 0 would let formatting silently drop the line. Note `dip -2`
	// is not a *negative* operand here: the lexer skips a bare leading `-`
	// (general tokenization, not version-specific), so it collapses to `dip 2`.
	// Below-floor rejection therefore only concerns 0.
	_, err := NewParser("dip 0\n\n"+versionBody, "test.dip").Parse()
	if err == nil {
		t.Fatal("expected error for version below 1, got nil")
	}
	if !strings.Contains(err.Error(), "invalid version declaration") {
		t.Errorf("error = %q, want it to mention invalid version declaration", err.Error())
	}
}

func TestParseDipDeclarationMissingOperandSingleDiagnostic(t *testing.T) {
	// `dip` with no operand must not consume the newline and then mis-report the
	// following token — exactly one diagnostic should result.
	p := NewParser("dip\n\n"+versionBody, "test.dip")
	_, _ = p.Parse()
	if got := len(p.Diagnostics()); got != 1 {
		t.Fatalf("missing operand produced %d diagnostics, want 1: %v", got, p.Diagnostics())
	}
	if !strings.Contains(p.Diagnostics()[0], "expected integer after 'dip'") {
		t.Errorf("diagnostic = %q, want missing-operand message", p.Diagnostics()[0])
	}
}

func TestParserVersionReachableFromEdges(t *testing.T) {
	// The acceptance criterion for #134: p.version is set before edge parsing.
	p := NewParser("dip 2\n\n"+versionBody, "test.dip")
	if _, err := p.Parse(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.version != 2 {
		t.Errorf("p.version = %d, want 2", p.version)
	}
}

func TestVersionRoundTrip(t *testing.T) {
	w, err := NewParser("dip 2\n\n"+versionBody, "test.dip").Parse()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := formatter.Format(w)
	if !strings.HasPrefix(out, "dip 2\n") {
		t.Errorf("formatted output does not start with dip 2 declaration:\n%s", out)
	}
	// Reparse the formatted output: version must survive.
	w2, err := NewParser(out, "test.dip").Parse()
	if err != nil {
		t.Fatalf("reparse error: %v", err)
	}
	if w2.Version != "2" {
		t.Errorf("round-tripped version = %q, want %q", w2.Version, "2")
	}
	if formatter.Format(w2) != out {
		t.Errorf("formatter not idempotent for versioned file")
	}
}

func TestV1FileGainsNoDeclaration(t *testing.T) {
	w, err := NewParser(versionBody, "test.dip").Parse()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := formatter.Format(w)
	if strings.Contains(out, "dip ") {
		t.Errorf("v1 file gained a dip declaration:\n%s", out)
	}
}
