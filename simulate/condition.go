package simulate

import (
	"fmt"
	"strings"

	"github.com/2389-research/dippin-lang/ir"
)

// ParseCondition parses a raw condition string into a ConditionExpr AST.
// This is needed because the parser stores raw condition text but does not
// always populate the Parsed field.
//
// Supported grammar:
//
//	expr   = orExpr
//	orExpr = andExpr ("or" andExpr)*
//	andExpr = unary ("and" unary)*
//	unary  = "not" unary | compare
//	compare = VARIABLE OP VALUE
//	OP     = "=" | "==" | "!=" | "contains" | "startswith" | "endswith" | "in"
func ParseCondition(raw string) (ir.ConditionExpr, error) {
	p := &condParser{tokens: tokenizeCondition(raw), pos: 0}
	expr, err := p.parseOr()
	if err != nil {
		return nil, err
	}
	if p.pos < len(p.tokens) {
		return nil, fmt.Errorf("unexpected token %q at position %d", p.tokens[p.pos], p.pos)
	}
	return expr, nil
}

type condParser struct {
	tokens []string
	pos    int
}

func (p *condParser) peek() string {
	if p.pos >= len(p.tokens) {
		return ""
	}
	return p.tokens[p.pos]
}

func (p *condParser) next() string {
	t := p.peek()
	p.pos++
	return t
}

func (p *condParser) parseOr() (ir.ConditionExpr, error) {
	left, err := p.parseAnd()
	if err != nil {
		return nil, err
	}
	for p.peek() == "or" {
		p.next() // consume "or"
		right, err := p.parseAnd()
		if err != nil {
			return nil, err
		}
		left = ir.CondOr{Left: left, Right: right}
	}
	return left, nil
}

func (p *condParser) parseAnd() (ir.ConditionExpr, error) {
	left, err := p.parseUnary()
	if err != nil {
		return nil, err
	}
	for p.peek() == "and" {
		p.next() // consume "and"
		right, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		left = ir.CondAnd{Left: left, Right: right}
	}
	return left, nil
}

func (p *condParser) parseUnary() (ir.ConditionExpr, error) {
	if p.peek() == "not" {
		p.next() // consume "not"
		inner, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		return ir.CondNot{Inner: inner}, nil
	}
	return p.parseCompare()
}

func (p *condParser) parseCompare() (ir.ConditionExpr, error) {
	variable := p.next()
	if variable == "" {
		return nil, fmt.Errorf("expected variable, got end of input")
	}

	// Handle infix negation: "variable not contains value" → CondNot{CondCompare{...}}
	negated := p.consumeInfixNot()

	op, value, err := p.parseOpValue(variable)
	if err != nil {
		return nil, err
	}

	cmp := ir.CondCompare{Variable: variable, Op: op, Value: value}
	if negated {
		return ir.CondNot{Inner: cmp}, nil
	}
	return cmp, nil
}

// consumeInfixNot consumes an infix "not" token if present.
func (p *condParser) consumeInfixNot() bool {
	if p.peek() != "not" {
		return false
	}
	p.next()
	return true
}

// parseOpValue parses the operator and value parts of a comparison.
func (p *condParser) parseOpValue(variable string) (string, string, error) {
	op := p.next()
	if op == "" {
		return "", "", fmt.Errorf("expected operator after %q, got end of input", variable)
	}
	if !isValidOperator(op) {
		return "", "", fmt.Errorf("unknown operator %q", op)
	}
	value := p.next()
	if value == "" {
		return "", "", fmt.Errorf("expected value after %q %s, got end of input", variable, op)
	}
	return op, value, nil
}

// validOperators is the set of recognized comparison operators.
// All operators do string comparison — there is no numeric coercion.
var validOperators = map[string]bool{
	"=": true, "==": true, "!=": true,
	"contains": true, "startswith": true, "endswith": true, "in": true,
}

// isValidOperator checks whether op is a recognized comparison operator.
func isValidOperator(op string) bool {
	return validOperators[op]
}

// tokenizeCondition splits a raw condition string into tokens.
// Handles quoted strings and standard whitespace splitting.
func tokenizeCondition(raw string) []string {
	var tokens []string
	raw = strings.TrimSpace(raw)
	i := 0

	for i < len(raw) {
		i = skipCondWhitespace(raw, i)
		if i >= len(raw) {
			break
		}

		token, consumed := tokenizeOne(raw, i)
		if consumed > 0 {
			tokens = append(tokens, token)
			i += consumed
		} else {
			i++ // skip unknown
		}
	}
	return tokens
}

// tokenizeOne dispatches to the appropriate tokenizer for the current position.
func tokenizeOne(raw string, i int) (string, int) {
	if token, n := tryTokenizeQuotedCond(raw, i); n > 0 {
		return token, n
	}
	if token, n := tryTokenizeOperatorCond(raw, i); n > 0 {
		return token, n
	}
	return tryTokenizeWordCond(raw, i)
}

// skipCondWhitespace advances the index past any whitespace characters.
func skipCondWhitespace(raw string, i int) int {
	for i < len(raw) && (raw[i] == ' ' || raw[i] == '\t') {
		i++
	}
	return i
}

// isQuoteChar returns true if ch is a single or double quote.
func isQuoteChar(ch byte) bool {
	return ch == '"' || ch == '\''
}

// tryTokenizeQuotedCond handles quoted strings. Double quotes honor backslash
// escapes (\" and \\); single quotes keep the existing scan-to-quote behavior
// so previously-valid single-quoted conditions are unchanged (#182).
func tryTokenizeQuotedCond(raw string, i int) (token string, consumed int) {
	if !isQuoteChar(raw[i]) {
		return "", 0
	}
	if raw[i] == '"' {
		return readDoubleQuotedCond(raw, i)
	}
	return readSingleQuotedCond(raw, i)
}

// readDoubleQuotedCond reads a double-quoted value from raw[i] (raw[i]=='"'),
// collapsing \" and \\ to their literal chars, until an unescaped closing ".
// Returns the unescaped content and the bytes consumed (both quotes included).
func readDoubleQuotedCond(raw string, i int) (string, int) {
	var b strings.Builder
	j := i + 1
	for j < len(raw) && raw[j] != '"' {
		j += appendCondEscaped(&b, raw, j)
	}
	if j < len(raw) {
		j++ // closing quote
	}
	return b.String(), j - i
}

// appendCondEscaped appends one char (collapsing \x to x) and returns bytes
// consumed — the inverse of parse_edges.escapeConditionLiteral (#182).
func appendCondEscaped(b *strings.Builder, raw string, j int) int {
	if raw[j] == '\\' && j+1 < len(raw) {
		b.WriteByte(raw[j+1])
		return 2
	}
	b.WriteByte(raw[j])
	return 1
}

// readSingleQuotedCond preserves the pre-#182 scan-to-quote behavior (no escapes).
func readSingleQuotedCond(raw string, i int) (string, int) {
	quote := raw[i]
	start := i + 1
	j := start
	for j < len(raw) && raw[j] != quote {
		j++
	}
	token := raw[start:j]
	if j < len(raw) {
		j++ // skip closing quote
	}
	return token, j - i
}

// operatorChars is the set of characters that form operators.
var operatorChars = [256]bool{
	'=': true, '!': true, '<': true, '>': true,
}

// isOperatorChar returns true if ch is an operator character.
func isOperatorChar(ch byte) bool {
	return operatorChars[ch]
}

// twoCharOperators is the set of valid two-character operators.
var twoCharOperators = map[string]bool{
	"==": true, "!=": true, "<=": true, ">=": true,
}

// tryTokenizeOperatorCond handles comparison operators.
func tryTokenizeOperatorCond(raw string, i int) (token string, consumed int) {
	if !isOperatorChar(raw[i]) {
		return "", 0
	}
	// Check for multi-char operators first.
	if i+1 < len(raw) && twoCharOperators[raw[i:i+2]] {
		return raw[i : i+2], 2
	}
	return string(raw[i]), 1
}

// tryTokenizeWordCond handles regular tokens (identifiers, keywords, values).
func tryTokenizeWordCond(raw string, i int) (token string, consumed int) {
	start := i
	for i < len(raw) && !isWordBreak(raw[i]) {
		i++
	}
	if i > start {
		return raw[start:i], i - start
	}
	return "", 0
}

// isWordBreak returns true if ch should terminate a word token.
func isWordBreak(ch byte) bool {
	return ch == ' ' || ch == '\t' || isOperatorChar(ch)
}

// EnsureConditionsParsed walks all edges and manager_loop nodes in a workflow
// and ensures that any Condition with a Raw string but nil Parsed field gets
// parsed. This is needed because the .dip parser stores Raw text only; the
// AST is lazily populated before lint checks run and before simulation.
func EnsureConditionsParsed(w *ir.Workflow) error {
	for _, e := range w.Edges {
		if err := ensureEdgeConditionParsed(e); err != nil {
			return err
		}
	}
	for _, n := range w.Nodes {
		if err := ensureNodeConditionsParsed(n); err != nil {
			return err
		}
	}
	return nil
}

// ensureEdgeConditionParsed parses a single edge's condition if needed.
func ensureEdgeConditionParsed(e *ir.Edge) error {
	if e.Condition == nil || e.Condition.Parsed != nil || e.Condition.Raw == "" {
		return nil
	}
	parsed, err := ParseCondition(e.Condition.Raw)
	if err != nil {
		return fmt.Errorf("edge %s -> %s: invalid condition %q: %w", e.From, e.To, e.Condition.Raw, err)
	}
	e.Condition.Parsed = parsed
	return nil
}

// ensureNodeConditionsParsed walks node-level conditions (currently manager_loop only).
func ensureNodeConditionsParsed(n *ir.Node) error {
	cfg, ok := n.Config.(ir.ManagerLoopConfig)
	if !ok {
		return nil
	}
	if err := ensureConditionParsed(cfg.StopCondition, n.ID, "stop_condition"); err != nil {
		return err
	}
	return ensureConditionParsed(cfg.SteerCondition, n.ID, "steer_condition")
}

// ensureConditionParsed parses a single Condition if it has Raw but no Parsed.
func ensureConditionParsed(c *ir.Condition, nodeID, field string) error {
	if c == nil || c.Parsed != nil || c.Raw == "" {
		return nil
	}
	parsed, err := ParseCondition(c.Raw)
	if err != nil {
		return fmt.Errorf("node %s %s: invalid condition %q: %w", nodeID, field, c.Raw, err)
	}
	c.Parsed = parsed
	return nil
}
