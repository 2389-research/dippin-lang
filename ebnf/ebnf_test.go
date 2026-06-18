package ebnf

import (
	"os"
	"strings"
	"testing"
)

// TestValidateAcceptsWellFormed: a minimal well-formed grammar has no errors.
func TestValidateAcceptsWellFormed(t *testing.T) {
	src := `(* a comment *)
a = "x" b [ c ] { d } ;
b = "y" | 'z' ;
c = FOO ;
d = ( a | b ) - "x" ;
`
	if errs := Validate(src); len(errs) != 0 {
		t.Fatalf("well-formed grammar reported errors: %v", errs)
	}
}

func TestValidateUnterminatedComment(t *testing.T) {
	errs := Validate("a = (* oops ;\n")
	if !anyContains(errs, "comment") {
		t.Fatalf("expected an unterminated-comment error, got: %v", errs)
	}
}

func TestValidateUnclosedString(t *testing.T) {
	errs := Validate("a = \"oops ;\n")
	if !anyContains(errs, "string") && !anyContains(errs, "quote") {
		t.Fatalf("expected an unclosed-string error, got: %v", errs)
	}
}

func TestValidateUnbalancedGroup(t *testing.T) {
	errs := Validate("a = ( \"x\" ;\n")
	if !anyContains(errs, "balanc") && !anyContains(errs, "(") {
		t.Fatalf("expected an unbalanced-group error, got: %v", errs)
	}
}

func TestValidateMismatchedBracket(t *testing.T) {
	errs := Validate("a = [ \"x\" ) ;\n")
	if len(errs) == 0 {
		t.Fatalf("expected a mismatched-bracket error, got none")
	}
}

func TestValidateMissingEquals(t *testing.T) {
	errs := Validate("a \"x\" ;\n")
	if len(errs) == 0 {
		t.Fatalf("expected a missing-'=' error, got none")
	}
}

func TestValidateUndefinedNonterminal(t *testing.T) {
	// b (lowercase) is referenced but never defined.
	errs := Validate("a = b ;\n")
	if !anyContains(errs, "b") || !anyContains(errs, "defin") {
		t.Fatalf("expected an undefined-nonterminal error for b, got: %v", errs)
	}
}

func TestValidateUppercaseTerminalsNeedNoDefinition(t *testing.T) {
	// FOO is uppercase => a lexer terminal, not a nonterminal; no definition needed.
	if errs := Validate("a = FOO NEWLINE ;\n"); len(errs) != 0 {
		t.Fatalf("uppercase terminals should not require definition, got: %v", errs)
	}
}

func TestValidateTrailingTokensAfterLastRule(t *testing.T) {
	// A dangling fragment with no terminating ';'.
	errs := Validate("a = \"x\" ;\nb = \"y\"\n")
	if len(errs) == 0 {
		t.Fatalf("expected an error for an unterminated final rule, got none")
	}
}

// TestRealGrammarIsWellFormed validates the checked-in canonical grammar.
func TestRealGrammarIsWellFormed(t *testing.T) {
	src, err := os.ReadFile("../docs/GRAMMAR.ebnf")
	if err != nil {
		t.Fatalf("read GRAMMAR.ebnf: %v", err)
	}
	if errs := Validate(string(src)); len(errs) != 0 {
		t.Fatalf("docs/GRAMMAR.ebnf is not well-formed:\n%s", strings.Join(errs, "\n"))
	}
}

func anyContains(errs []string, sub string) bool {
	for _, e := range errs {
		if strings.Contains(e, sub) {
			return true
		}
	}
	return false
}
