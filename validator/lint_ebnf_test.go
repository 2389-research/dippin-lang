package validator_test

import (
	"os"
	"strings"
	"testing"

	"github.com/2389-research/dippin-lang/simulate"
)

// TestEBNFOperatorsMatchParser verifies that every operator accepted by the
// condition parser is documented in GRAMMAR.ebnf, and vice versa.
// This catches operator drift between implementation and spec.
func TestEBNFOperatorsMatchParser(t *testing.T) {
	ebnf, err := os.ReadFile("../docs/GRAMMAR.ebnf")
	if err != nil {
		t.Fatalf("read GRAMMAR.ebnf: %v", err)
	}
	content := string(ebnf)

	// Operators the parser accepts (from simulate.ParseCondition).
	// Test each by parsing a condition that uses it.
	parserOps := []string{"=", "==", "!=", "contains", "startswith", "endswith", "in"}
	for _, op := range parserOps {
		raw := "ctx.x " + op + " val"
		if _, err := simulate.ParseCondition(raw); err != nil {
			t.Errorf("parser rejects operator %q: %v", op, err)
			continue
		}
		if !strings.Contains(content, `"`+op+`"`) {
			t.Errorf("operator %q accepted by parser but not in GRAMMAR.ebnf", op)
		}
	}

	// Operators that should NOT be accepted.
	rejectedOps := []string{"<", ">", "<=", ">=", "like", "matches"}
	for _, op := range rejectedOps {
		raw := "ctx.x " + op + " val"
		if _, err := simulate.ParseCondition(raw); err == nil {
			t.Errorf("parser accepts operator %q which should be rejected", op)
		}
	}
}

// TestEBNFInfixNegationDocumented verifies that infix negation syntax
// (e.g., "var not contains val") is documented in the EBNF.
func TestEBNFInfixNegationDocumented(t *testing.T) {
	ebnf, err := os.ReadFile("../docs/GRAMMAR.ebnf")
	if err != nil {
		t.Fatalf("read GRAMMAR.ebnf: %v", err)
	}
	content := string(ebnf)

	// The EBNF should document infix negation (W3C optional: `"not"?`).
	if !strings.Contains(content, `"not"?`) {
		t.Error("GRAMMAR.ebnf does not document infix negation syntax")
	}

	// The parser should accept infix negation.
	if _, err := simulate.ParseCondition("ctx.x not contains val"); err != nil {
		t.Errorf("parser rejects infix negation: %v", err)
	}
}

// TestEBNFNodeKindsDocumented verifies that all 7 node kinds appear in the EBNF.
func TestEBNFNodeKindsDocumented(t *testing.T) {
	ebnf, err := os.ReadFile("../docs/GRAMMAR.ebnf")
	if err != nil {
		t.Fatalf("read GRAMMAR.ebnf: %v", err)
	}
	content := string(ebnf)

	kinds := []string{"agent", "human", "tool", "parallel", "fan_in", "subgraph", "conditional"}
	for _, k := range kinds {
		if !strings.Contains(content, `"`+k+`"`) {
			t.Errorf("node kind %q not found in GRAMMAR.ebnf", k)
		}
	}
}

// TestEBNFToolOutputsDocumented verifies the tool outputs field is in EBNF.
func TestEBNFToolOutputsDocumented(t *testing.T) {
	ebnf, err := os.ReadFile("../docs/GRAMMAR.ebnf")
	if err != nil {
		t.Fatalf("read GRAMMAR.ebnf: %v", err)
	}
	if !strings.Contains(string(ebnf), `"outputs"`) {
		t.Error("GRAMMAR.ebnf does not document tool outputs field")
	}
}
