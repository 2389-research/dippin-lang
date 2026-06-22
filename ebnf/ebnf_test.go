package ebnf

import (
	"os"
	"strings"
	"testing"
)

// TestValidateAcceptsWellFormed: a minimal well-formed W3C grammar is clean.
func TestValidateAcceptsWellFormed(t *testing.T) {
	src := `/* a comment */
a ::= "x" b c? d* e+
b ::= "y" | 'z'
c ::= FOO
d ::= ( a | b ) - "x"
e ::= [A-Za-z0-9] #xA
`
	if errs := Validate(src); len(errs) != 0 {
		t.Fatalf("well-formed grammar reported errors: %v", errs)
	}
}

func TestValidateUnterminatedComment(t *testing.T) {
	if errs := Validate("a ::= /* oops\n"); !anyContains(errs, "comment") {
		t.Fatalf("expected an unterminated-comment error, got: %v", errs)
	}
}

func TestValidateUnclosedString(t *testing.T) {
	errs := Validate("a ::= \"oops\n")
	if !anyContains(errs, "string") && !anyContains(errs, "quote") {
		t.Fatalf("expected an unclosed-string error, got: %v", errs)
	}
}

func TestValidateUnterminatedCharClass(t *testing.T) {
	if errs := Validate("a ::= [A-Z\n"); !anyContains(errs, "character class") {
		t.Fatalf("expected an unterminated-char-class error, got: %v", errs)
	}
}

func TestValidateUnbalancedGroup(t *testing.T) {
	if errs := Validate("a ::= ( \"x\"\nb ::= \"y\"\n"); len(errs) == 0 {
		t.Fatalf("expected an unbalanced-group error, got none")
	}
}

func TestValidateMissingDefEq(t *testing.T) {
	// 'a' is a production head but lacks '::='.
	if errs := Validate("a \"x\"\n"); !anyContains(errs, "::=") {
		t.Fatalf("expected a missing-'::=' error, got: %v", errs)
	}
}

func TestValidateUndefinedNonterminal(t *testing.T) {
	// b (lowercase) is referenced but never defined.
	errs := Validate("a ::= b\n")
	if !anyContains(errs, "b") || !anyContains(errs, "defin") {
		t.Fatalf("expected an undefined-nonterminal error for b, got: %v", errs)
	}
}

func TestValidateUppercaseTerminalsNeedNoDefinition(t *testing.T) {
	if errs := Validate("a ::= FOO NEWLINE\n"); len(errs) != 0 {
		t.Fatalf("uppercase terminals should not require definition, got: %v", errs)
	}
}

// TestRealGrammarIsValidW3CEBNF validates the checked-in canonical grammar.
func TestRealGrammarIsValidW3CEBNF(t *testing.T) {
	src, err := os.ReadFile("../docs/GRAMMAR.ebnf")
	if err != nil {
		t.Fatalf("read GRAMMAR.ebnf: %v", err)
	}
	if errs := Validate(string(src)); len(errs) != 0 {
		t.Fatalf("docs/GRAMMAR.ebnf is not valid W3C EBNF:\n%s", strings.Join(errs, "\n"))
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
