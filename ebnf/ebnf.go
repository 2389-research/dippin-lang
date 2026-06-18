// Package ebnf provides a structural well-formedness validator for the
// dippin grammar specification (docs/GRAMMAR.ebnf).
//
// It validates the grammar's actual dialect — relaxed/W3C-style EBNF with
// space concatenation (not ISO/IEC 14977 comma concatenation, which the spec
// deliberately omits for readability): `(* *)` comments, `"..."`/`'...'`
// terminals, UPPERCASE lexer terminals, lowercase nonterminals, the metasymbols
// `= ; | ( ) [ ] { } -`, and `;`-terminated rules.
//
// Validate reports: unterminated comments, unclosed string terminals, malformed
// rule heads (`name = ...`), unbalanced/mismatched grouping brackets, an
// unterminated final rule, and nonterminals referenced but never defined.
package ebnf

import (
	"fmt"
	"unicode"
)

// Validate returns a list of well-formedness errors in the EBNF source.
// An empty slice means the grammar is well-formed.
func Validate(src string) []string {
	toks, errs := scan(src)
	if len(errs) > 0 {
		return errs
	}
	return checkRules(toks)
}

// token is a lexical unit. kind is 'n' (name), 's' (string), or the literal
// metasymbol byte for '=', ';', '|', '-', '(', ')', '[', ']', '{', '}'.
type token struct {
	kind byte
	text string
}

// --- scanner ---

type scanner struct {
	r   []rune
	i   int
	out []token
	err []string
}

func scan(src string) ([]token, []string) {
	s := &scanner{r: []rune(src)}
	for s.i < len(s.r) {
		s.step()
	}
	return s.out, s.err
}

func (s *scanner) peek(n int) rune {
	if s.i+n < len(s.r) {
		return s.r[s.i+n]
	}
	return 0
}

func (s *scanner) step() {
	c := s.r[s.i]
	if s.tryDelimited(c) {
		return
	}
	switch {
	case unicode.IsSpace(c):
		s.i++
	case isNameStart(c):
		s.scanName()
	default:
		s.scanSymbol(c)
	}
}

// tryDelimited consumes a comment `(* *)`, an ISO special sequence `? ?`, or a
// string terminal — the three constructs whose bodies are skipped wholesale so
// arbitrary text inside them (apostrophes, brackets) does not perturb the scan.
// Returns false if the current rune starts none of them.
func (s *scanner) tryDelimited(c rune) bool {
	switch {
	case s.atComment():
		s.skipComment()
	case c == '?':
		s.skipSpecial()
	case isQuote(c):
		s.scanString(c)
	default:
		return false
	}
	return true
}

func (s *scanner) atComment() bool {
	return s.r[s.i] == '(' && s.peek(1) == '*'
}

// skipSpecial consumes an ISO/IEC 14977 special sequence: ? ... ? (its content
// is implementation-defined prose and is ignored).
func (s *scanner) skipSpecial() {
	s.i++ // opening '?'
	for s.i < len(s.r) {
		if s.r[s.i] == '?' {
			s.i++
			return
		}
		s.i++
	}
	s.err = append(s.err, "unterminated special sequence (missing closing `?`)")
}

func (s *scanner) skipComment() {
	s.i += 2 // consume "(*"
	for s.i < len(s.r) {
		if s.r[s.i] == '*' && s.peek(1) == ')' {
			s.i += 2
			return
		}
		s.i++
	}
	s.err = append(s.err, "unterminated comment (missing `*)`)")
}

func (s *scanner) scanString(q rune) {
	start := s.i
	s.i++ // opening quote
	for s.i < len(s.r) {
		switch s.r[s.i] {
		case q:
			s.i++
			s.out = append(s.out, token{'s', string(s.r[start:s.i])})
			return
		case '\n':
			s.err = append(s.err, "unclosed string literal (newline before closing quote)")
			return
		}
		s.i++
	}
	s.err = append(s.err, "unclosed string literal (missing closing quote at end of file)")
}

func (s *scanner) scanName() {
	start := s.i
	for s.i < len(s.r) && isNamePart(s.r[s.i]) {
		s.i++
	}
	s.out = append(s.out, token{'n', string(s.r[start:s.i])})
}

func (s *scanner) scanSymbol(c rune) {
	s.i++
	switch c {
	case '=', ';', '|', '-', '(', ')', '[', ']', '{', '}':
		s.out = append(s.out, token{byte(c), string(c)})
	}
	// Any other punctuation (e.g. an ISO concatenation comma) is ignored — it
	// does not affect rule shape or bracket balance.
}

// --- structural checker ---

type checker struct {
	defined map[string]bool
	refs    []string
	errs    []string
}

func checkRules(toks []token) []string {
	c := &checker{defined: map[string]bool{}}
	var rule []token
	for _, t := range toks {
		if t.kind == ';' {
			c.rule(rule)
			rule = nil
			continue
		}
		rule = append(rule, t)
	}
	if len(rule) > 0 {
		c.errs = append(c.errs, fmt.Sprintf("unterminated rule (missing `;`): starts with %q", rule[0].text))
	}
	c.checkUndefined()
	return c.errs
}

func (c *checker) rule(toks []token) {
	if len(toks) < 2 || toks[0].kind != 'n' || toks[1].kind != '=' {
		c.errs = append(c.errs, malformedHead(toks))
		return
	}
	c.defined[toks[0].text] = true
	c.checkBody(toks[0].text, toks[2:])
}

func malformedHead(toks []token) string {
	if len(toks) == 0 {
		return "empty rule (stray `;`)"
	}
	return fmt.Sprintf("rule %q is malformed: expected `name = ...`", toks[0].text)
}

func (c *checker) checkBody(name string, body []token) {
	var stack []byte
	for _, t := range body {
		c.bodyToken(name, t, &stack)
	}
	if len(stack) > 0 {
		c.errs = append(c.errs, fmt.Sprintf("rule %q: unclosed group (missing closing bracket)", name))
	}
}

func (c *checker) bodyToken(name string, t token, stack *[]byte) {
	switch t.kind {
	case '(', '[', '{':
		*stack = append(*stack, t.kind)
	case ')', ']', '}':
		if !popMatches(stack, t.kind) {
			c.errs = append(c.errs, fmt.Sprintf("rule %q: unbalanced %q", name, t.text))
		}
	case 'n':
		c.refs = append(c.refs, t.text)
	}
}

func popMatches(stack *[]byte, closing byte) bool {
	n := len(*stack)
	if n == 0 {
		return false
	}
	open := (*stack)[n-1]
	*stack = (*stack)[:n-1]
	return bracketsMatch(open, closing)
}

func bracketsMatch(open, closing byte) bool {
	switch open {
	case '(':
		return closing == ')'
	case '[':
		return closing == ']'
	case '{':
		return closing == '}'
	}
	return false
}

func (c *checker) checkUndefined() {
	seen := map[string]bool{}
	for _, r := range c.refs {
		if seen[r] || !isNonterminal(r) || c.defined[r] {
			continue
		}
		seen[r] = true
		c.errs = append(c.errs, fmt.Sprintf("undefined nonterminal %q (referenced but no rule defines it)", r))
	}
}

// --- character classes ---

// isNonterminal reports whether a name is a grammar nonterminal (lowercase
// first letter). UPPERCASE names are lexer terminals and need no definition.
func isNonterminal(name string) bool {
	return name != "" && name[0] >= 'a' && name[0] <= 'z'
}

func isQuote(c rune) bool {
	return c == '"' || c == '\''
}

func isNameStart(c rune) bool {
	return c == '_' || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

func isNamePart(c rune) bool {
	return isNameStart(c) || (c >= '0' && c <= '9')
}
