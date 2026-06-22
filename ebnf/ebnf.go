// Package ebnf validates the dippin grammar specification (docs/GRAMMAR.ebnf)
// against the W3C EBNF notation (the XML-spec / bottlecaps dialect).
//
// It is, in effect, a parser for the W3C "EBNF of EBNF":
//
//	Grammar    ::= Production*
//	Production ::= Name '::=' Choice
//	Choice     ::= Sequence ('|' Sequence)*
//	Sequence   ::= Difference*
//	Difference ::= Item ('-' Item)?
//	Item       ::= Primary ('?' | '*' | '+')*
//	Primary    ::= Name | Terminal | '(' Choice ')'
//	Terminal   ::= StringLiteral | CharCode | CharClass
//
// where a production whose RHS reaches a `Name '::='` lookahead has ended.
// Validate reports lexical errors (unterminated comment / string / char class),
// structural errors (missing `::=`, unbalanced `()`, stray tokens), and
// nonterminals referenced but never defined (UPPERCASE names are lexer
// terminals and need no production).
package ebnf

import (
	"fmt"
	"strings"
	"unicode"
)

// Validate returns a list of well-formedness errors. Empty means well-formed.
func Validate(src string) []string {
	toks, errs := scan(src)
	if len(errs) > 0 {
		return errs
	}
	p := &parser{toks: toks, defined: map[string]bool{}}
	p.grammar()
	p.checkUndefined()
	return p.errs
}

// token kinds: 'n' name, 't' terminal (string/charcode/charclass),
// 'd' '::=', and the literal bytes '|', '-', '?', '*', '+', '(', ')'.
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
	if s.tryDelimited() {
		return
	}
	c := s.r[s.i]
	switch {
	case unicode.IsSpace(c):
		s.i++
	case s.atCharCode():
		s.scanCharCode()
	case isNameStart(c):
		s.scanName()
	default:
		s.scanSymbol(c)
	}
}

// tryDelimited consumes a comment, string terminal, or char class — the
// constructs whose bodies are skipped/captured wholesale. Returns false if the
// current rune starts none of them.
func (s *scanner) tryDelimited() bool {
	c := s.r[s.i]
	switch {
	case s.atComment():
		s.skipComment()
	case isQuote(c):
		s.scanString(c)
	case c == '[':
		s.scanCharClass()
	default:
		return false
	}
	return true
}

func (s *scanner) atComment() bool {
	return s.r[s.i] == '/' && s.peek(1) == '*'
}

func (s *scanner) atCharCode() bool {
	return s.r[s.i] == '#' && s.peek(1) == 'x'
}

func (s *scanner) skipComment() {
	s.i += 2 // "/*"
	for s.i < len(s.r) {
		if s.r[s.i] == '*' && s.peek(1) == '/' {
			s.i += 2
			return
		}
		s.i++
	}
	s.err = append(s.err, "unterminated comment (missing `*/`)")
}

func (s *scanner) scanString(q rune) {
	start := s.i
	s.i++ // opening quote
	for s.i < len(s.r) {
		switch s.r[s.i] {
		case q:
			s.i++
			s.out = append(s.out, token{'t', string(s.r[start:s.i])})
			return
		case '\n':
			s.err = append(s.err, "unclosed string literal (newline before closing quote)")
			return
		}
		s.i++
	}
	s.err = append(s.err, "unclosed string literal (missing closing quote at end of file)")
}

func (s *scanner) scanCharClass() {
	start := s.i
	s.i++ // '['
	for s.i < len(s.r) {
		if s.r[s.i] == ']' {
			s.i++
			s.out = append(s.out, token{'t', string(s.r[start:s.i])})
			return
		}
		if s.r[s.i] == '\n' {
			break
		}
		s.i++
	}
	s.err = append(s.err, "unterminated character class (missing `]`)")
}

func (s *scanner) scanCharCode() {
	start := s.i
	s.i += 2 // "#x"
	for s.i < len(s.r) && isHex(s.r[s.i]) {
		s.i++
	}
	s.out = append(s.out, token{'t', string(s.r[start:s.i])})
}

func (s *scanner) scanName() {
	start := s.i
	for s.i < len(s.r) && isNamePart(s.r[s.i]) {
		s.i++
	}
	s.out = append(s.out, token{'n', string(s.r[start:s.i])})
}

func (s *scanner) scanSymbol(c rune) {
	if c == ':' && s.peek(1) == ':' && s.peek(2) == '=' {
		s.i += 3
		s.out = append(s.out, token{'d', "::="})
		return
	}
	s.i++
	switch c {
	case '|', '-', '?', '*', '+', '(', ')':
		s.out = append(s.out, token{byte(c), string(c)})
	default:
		s.err = append(s.err, fmt.Sprintf("unexpected character %q", string(c)))
	}
}

// --- parser ---

type parser struct {
	toks    []token
	i       int
	defined map[string]bool
	refs    []string
	errs    []string
}

var eofToken = token{kind: 0}

func (p *parser) peekN(n int) token {
	if p.i+n < len(p.toks) {
		return p.toks[p.i+n]
	}
	return eofToken
}

func (p *parser) cur() token  { return p.peekN(0) }
func (p *parser) next() token { t := p.cur(); p.i++; return t }
func (p *parser) at(k byte) bool {
	return p.cur().kind == k
}

func (p *parser) grammar() {
	for p.i < len(p.toks) {
		p.production()
	}
}

func (p *parser) production() {
	head := p.next()
	if head.kind != 'n' {
		p.errs = append(p.errs, fmt.Sprintf("expected a rule name to start a production, got %q", head.text))
		return
	}
	if !p.at('d') {
		p.errs = append(p.errs, fmt.Sprintf("rule %q: expected `::=` after the name", head.text))
		return
	}
	p.next() // '::='
	p.defined[head.text] = true
	p.choice()
}

func (p *parser) choice() {
	p.sequence()
	for p.at('|') {
		p.next()
		p.sequence()
	}
}

func (p *parser) sequence() {
	for p.startsItem() {
		p.difference()
	}
}

// startsItem reports whether the current token begins another item in the
// current sequence. A name that is the head of the next production (Name '::=')
// ends the sequence instead.
func (p *parser) startsItem() bool {
	switch p.cur().kind {
	case 'n':
		return !(p.peekN(1).kind == 'd')
	case 't', '(':
		return true
	}
	return false
}

func (p *parser) difference() {
	p.item()
	if p.at('-') {
		p.next()
		p.item()
	}
}

func (p *parser) item() {
	p.primary()
	for p.atPostfix() {
		p.next()
	}
}

func (p *parser) atPostfix() bool {
	switch p.cur().kind {
	case '?', '*', '+':
		return true
	}
	return false
}

func (p *parser) primary() {
	t := p.next()
	switch t.kind {
	case 'n':
		p.refs = append(p.refs, t.text)
	case 't':
		// terminal — needs no definition
	case '(':
		p.group()
	default:
		p.errs = append(p.errs, fmt.Sprintf("unexpected token %q in expression", t.text))
	}
}

func (p *parser) group() {
	p.choice()
	if !p.at(')') {
		p.errs = append(p.errs, "unbalanced group (missing `)`)")
		return
	}
	p.next() // ')'
}

func (p *parser) checkUndefined() {
	seen := map[string]bool{}
	for _, r := range p.refs {
		if seen[r] || !isNonterminal(r) || p.defined[r] {
			continue
		}
		seen[r] = true
		p.errs = append(p.errs, fmt.Sprintf("undefined nonterminal %q (referenced but no rule defines it)", r))
	}
}

// --- character classes ---

// isNonterminal reports whether a name is a grammar nonterminal (lowercase
// first letter). UPPERCASE names are lexer terminals and need no production.
func isNonterminal(name string) bool {
	return name != "" && name[0] >= 'a' && name[0] <= 'z'
}

func isNameStart(c rune) bool {
	return c == '_' || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

func isNamePart(c rune) bool {
	return isNameStart(c) || (c >= '0' && c <= '9')
}

func isQuote(c rune) bool {
	return c == '"' || c == '\''
}

func isHex(c rune) bool {
	return strings.ContainsRune("0123456789abcdefABCDEF", c)
}
