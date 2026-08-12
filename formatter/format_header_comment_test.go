package formatter

import (
	"strings"
	"testing"

	"github.com/2389-research/dippin-lang/parser"
)

// TestFormatHeaderCommentRoundtrip verifies the leading file-level comment block
// (e.g. an `# ABOUTME:` header) survives parse -> Format verbatim at the top (#259).
func TestFormatHeaderCommentRoundtrip(t *testing.T) {
	input := `# ABOUTME: This workflow ships a feature end to end.
# ABOUTME: Requires the anthropic + openai providers.
#
# Provider deps: anthropic, openai

workflow WithHeader
  start: Begin
  exit: Begin

  agent Begin
    prompt: "Go."
`
	w, err := parser.NewParser(input, "test.dip").Parse()
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	want := []string{
		"# ABOUTME: This workflow ships a feature end to end.",
		"# ABOUTME: Requires the anthropic + openai providers.",
		"#",
		"# Provider deps: anthropic, openai",
	}
	if len(w.HeaderComment) != len(want) {
		t.Fatalf("HeaderComment = %#v, want %#v", w.HeaderComment, want)
	}
	for i := range want {
		if w.HeaderComment[i] != want[i] {
			t.Errorf("HeaderComment[%d] = %q, want %q", i, w.HeaderComment[i], want[i])
		}
	}

	formatted := Format(w)
	header := strings.Join(want, "\n") + "\n\n"
	if !strings.HasPrefix(formatted, header) {
		t.Errorf("formatted output does not begin with the verbatim header block\ngot:\n%s", formatted)
	}
	if !strings.Contains(formatted, "workflow WithHeader") {
		t.Errorf("formatted output missing workflow line\ngot:\n%s", formatted)
	}

	// Round-trips: reparsing and reformatting is stable.
	w2, err := parser.NewParser(formatted, "test.dip").Parse()
	if err != nil {
		t.Fatalf("second parse failed: %v\nformatted:\n%s", err, formatted)
	}
	if reformatted := Format(w2); formatted != reformatted {
		t.Errorf("formatter not idempotent with header comment\nfirst:\n%s\nsecond:\n%s", formatted, reformatted)
	}
}

// TestFormatNoHeaderCommentByteIdentical guards against a spurious leading blank
// line when the file has no leading comment (#259).
func TestFormatNoHeaderCommentByteIdentical(t *testing.T) {
	input := `workflow NoHeader
  start: Begin
  exit: Begin

  agent Begin
    prompt: "Go."
`
	w, err := parser.NewParser(input, "test.dip").Parse()
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if len(w.HeaderComment) != 0 {
		t.Fatalf("HeaderComment = %#v, want empty", w.HeaderComment)
	}

	formatted := Format(w)
	if strings.HasPrefix(formatted, "\n") {
		t.Errorf("formatted output starts with a spurious blank line\ngot:\n%q", formatted)
	}
	if !strings.HasPrefix(formatted, "workflow NoHeader") {
		t.Errorf("formatted output should start with the workflow line\ngot:\n%s", formatted)
	}
}
