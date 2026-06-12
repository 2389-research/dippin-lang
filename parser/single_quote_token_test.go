package parser

import (
	"strings"
	"testing"
)

// --- Lexer: single-quoted strings become one TokenLiteral (#120) ---

func TestLexSingleQuotedString(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []TokenType
	}{
		{
			name:  "single quoted with spaces",
			input: `'hello world'`,
			want:  []TokenType{TokenLiteral, TokenNewline, TokenEOF},
		},
		{
			name:  "single quoted glued after identifier and colon",
			input: `label: 'My Label'`,
			want:  []TokenType{TokenIdentifier, TokenColon, TokenLiteral, TokenNewline, TokenEOF},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := NewLexer(tt.input, "test.dip")
			var got []TokenType
			for {
				tok := l.NextToken()
				got = append(got, tok.Type)
				if tok.Type == TokenEOF {
					break
				}
			}
			if len(got) != len(tt.want) {
				t.Fatalf("token count = %d (%v), want %d (%v)", len(got), got, len(tt.want), tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Fatalf("token[%d] = %v, want %v (full %v)", i, got[i], tt.want[i], got)
				}
			}
		})
	}
}

// TestLexSingleQuotedContent verifies the stored literal value: surrounding
// quotes stripped, a doubled single quote collapsed to one, no backslash processing.
func TestLexSingleQuotedContent(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"plain", `'hello world'`, "hello world"},
		{"doubled escape", `'it''s here'`, "it's here"},
		{"empty", `''`, ""},
		{"literal backslash", `'a \d+ b'`, `a \d+ b`},
		{"internal hash", `'a # b'`, "a # b"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := NewLexer(tt.input, "test.dip")
			tok := l.NextToken()
			if tok.Type != TokenLiteral {
				t.Fatalf("first token = %v, want TokenLiteral", tok.Type)
			}
			if tok.Value != tt.want {
				t.Fatalf("value = %q, want %q", tok.Value, tt.want)
			}
		})
	}
}

// --- Comment stripping: # inside a single-quoted token is literal,
//     but a prose apostrophe must NOT protect a trailing comment (#120). ---

func TestStripCommentSingleQuoteToken(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		// A # inside a single-quoted TOKEN is content, not a comment.
		{"hash in single-quoted label", `    A -> B label: 'a # b'`, `    A -> B label: 'a # b'`},
		{"hash in single-quoted condition", `    A -> B when ctx.x = 'a # b'`, `    A -> B when ctx.x = 'a # b'`},
		{"trailing comment after closing quote", `    A -> B label: 'a b' # note`, `    A -> B label: 'a b' `},
		// Colon-glued single quote (no space after `:`) still opens a token.
		{"hash in colon-glued label", `    A -> B label:'a # b'`, `    A -> B label:'a # b'`},
		// REGRESSION GUARD: a prose apostrophe must not protect the comment.
		{"prose apostrophe still strips comment", `  goal: it's great # note`, `  goal: it's great `},
		{"prose apostrophe no comment", `  goal: it's great`, `  goal: it's great`},
		// '' doubled-quote escape inside a token keeps the # literal.
		{"doubled quote then hash", `    A -> B label: 'it''s # x'`, `    A -> B label: 'it''s # x'`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := stripComment(tt.in); got != tt.want {
				t.Fatalf("stripComment(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

// --- End-to-end: edge label and when condition accept single-quoted values. ---

func TestParseEdgeSingleQuotedLabel(t *testing.T) {
	p := NewParser(buildEdgeDip("A -> B label: 'a # b'"), "test.dip")
	w, err := p.Parse()
	if err != nil {
		t.Fatalf("parse error: %v (%v)", err, p.Diagnostics())
	}
	if got := w.Edges[0].Label; got != "a # b" {
		t.Fatalf("label = %q, want %q", got, "a # b")
	}
}

func TestParseEdgeSingleQuotedCondition(t *testing.T) {
	p := NewParser(buildEdgeDip("A -> B when ctx.reason = 'gave up'"), "test.dip")
	w, err := p.Parse()
	if err != nil {
		t.Fatalf("parse error: %v (%v)", err, p.Diagnostics())
	}
	got := w.Edges[0].Condition
	// A single-quoted condition operand normalizes to a double-quoted literal
	// in the raw condition text (matches existing TokenLiteral handling).
	if got == nil || got.Raw != `ctx.reason = "gave up"` {
		t.Fatalf("condition = %+v, want Raw %q", got, `ctx.reason = "gave up"`)
	}
}

func TestParseEdgeSingleQuotedDoubledEscape(t *testing.T) {
	p := NewParser(buildEdgeDip("A -> B label: 'it''s done'"), "test.dip")
	w, err := p.Parse()
	if err != nil {
		t.Fatalf("parse error: %v (%v)", err, p.Diagnostics())
	}
	if got := w.Edges[0].Label; got != "it's done" {
		t.Fatalf("label = %q, want %q", got, "it's done")
	}
}

// TestParseEdgeSingleQuotedHelpersAgree guards that prose values with
// apostrophes elsewhere in the same workflow still strip trailing comments.
func TestParseProseApostropheStillStripsComment(t *testing.T) {
	src := "workflow X\n" +
		"  goal: it's great # trailing note\n" +
		"  start: A\n" +
		"  exit: A\n" +
		"\n" +
		"  agent A\n" +
		"    prompt: \"Do A.\"\n"
	p := NewParser(src, "test.dip")
	w, err := p.Parse()
	if err != nil {
		t.Fatalf("parse error: %v (%v)", err, p.Diagnostics())
	}
	if !strings.HasPrefix(w.Goal, "it's great") || strings.Contains(w.Goal, "#") {
		t.Fatalf("goal = %q, want it's great with the comment stripped", w.Goal)
	}
}
