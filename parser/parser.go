package parser

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/2389-research/dippin-lang/ir"
)

type Parser struct {
	lexer       *Lexer
	filename    string
	diagnostics []string // Simple for now
	workflow    *ir.Workflow
	version     int // .dip format version; defaults to 1 when no `dip N` declaration is present
}

func NewParser(input string, filename string) *Parser {
	return &Parser{
		lexer:    NewLexer(input, filename),
		filename: filename,
		version:  1,
		workflow: &ir.Workflow{
			SourceMap: &ir.SourceMap{},
		},
	}
}

func (p *Parser) Parse() (*ir.Workflow, error) {
	p.parseVersionDeclaration()
	p.parseTopLevel()
	p.rejectRedundantFanEdgesUnderV2()
	if len(p.diagnostics) > 0 {
		return p.workflow, fmt.Errorf("parsing errors: %s", strings.Join(p.diagnostics, "; "))
	}
	return p.workflow, nil
}

// rejectRedundantFanEdgesUnderV2 reports, under dip 2, any edges-block edge that
// merely repeats an inline parallel/fan_in fork — the inline list is the single
// source of truth under dip 2 (#136). No-op under v1, where DIP153 warns instead.
func (p *Parser) rejectRedundantFanEdgesUnderV2() {
	if p.version < 2 {
		return
	}
	for _, e := range p.workflow.Edges {
		if ir.IsRedundantFanEdge(p.workflow, e) {
			p.diagnostics = append(p.diagnostics, fmt.Sprintf(
				"redundant edge %q -> %q: the inline parallel/fan_in list is authoritative under dip 2 — remove it (run `dippin fmt`) at %d:%d",
				e.From, e.To, e.Source.Line, e.Source.Column))
		}
	}
}

// Diagnostics returns the accumulated diagnostic messages.
func (p *Parser) Diagnostics() []string {
	return p.diagnostics
}

// parseVersionDeclaration consumes an optional leading `dip N` format-version
// line. It runs before any body parsing so later edge-parsing routines can
// branch on p.version. Absent a declaration the version stays 1. The parsed
// value is mirrored into Workflow.Version as its string form.
func (p *Parser) parseVersionDeclaration() {
	for p.lexer.PeekToken().Type == TokenNewline {
		p.lexer.NextToken()
	}
	t := p.lexer.PeekToken()
	if t.Type != TokenIdentifier || t.Value != "dip" {
		p.workflow.Version = strconv.Itoa(p.version)
		return
	}
	p.lexer.NextToken() // dip
	p.consumeVersionNumber()
	p.expect(TokenNewline)
	p.workflow.Version = strconv.Itoa(p.version)
}

// consumeVersionNumber reads and validates the numeric operand of a `dip N`
// declaration, leaving the trailing newline for the caller. A missing operand
// (`dip` then end-of-line) is reported without consuming the newline, so the
// caller's newline expectation still holds and no spurious second diagnostic
// follows. Versions below 1 are rejected: the formatter only emits `dip N` for
// N > 1, so accepting `dip 0` would let formatting silently drop the line.
func (p *Parser) consumeVersionNumber() {
	num := p.lexer.PeekToken()
	if num.Type == TokenNewline {
		p.diagnostics = append(p.diagnostics, fmt.Sprintf("invalid version declaration: expected integer after 'dip' at %d:%d", num.Location.Line, num.Location.Column))
		return
	}
	p.lexer.NextToken() // operand
	n, err := strconv.Atoi(num.Value)
	if err != nil || n < 1 {
		p.diagnostics = append(p.diagnostics, fmt.Sprintf("invalid version declaration: expected integer >= 1 after 'dip', got %q at %d:%d", num.Value, num.Location.Line, num.Location.Column))
		return
	}
	p.version = n
}

// parseTopLevel consumes top-level tokens looking for workflow declarations.
func (p *Parser) parseTopLevel() {
	for p.lexer.PeekToken().Type != TokenEOF {
		t := p.lexer.PeekToken()
		if t.Type == TokenNewline {
			p.lexer.NextToken()
			continue
		}
		if t.Type == TokenIdentifier && t.Value == "workflow" {
			p.parseWorkflow()
		} else {
			p.lexer.NextToken()
		}
	}
}

func (p *Parser) parseWorkflow() {
	p.lexer.NextToken() // workflow
	name := p.lexer.NextToken().Value
	p.workflow.Name = name
	p.expect(TokenNewline)

	p.expect(TokenIndent)
	p.parseWorkflowBody()
	p.expect(TokenOutdent)
}

// parseWorkflowBody parses the indented body of a workflow declaration.
func (p *Parser) parseWorkflowBody() {
	for p.lexer.PeekToken().Type != TokenOutdent && p.lexer.PeekToken().Type != TokenEOF {
		t := p.lexer.PeekToken()
		if t.Type == TokenNewline {
			p.lexer.NextToken()
			continue
		}
		if t.Type == TokenIdentifier {
			p.dispatchWorkflowField(t)
		} else {
			p.lexer.NextToken()
		}
	}
}

// workflowNodeKinds maps identifiers to their node kinds for dispatch.
var workflowNodeKinds = map[string]bool{
	"agent": true, "human": true, "tool": true,
	"subgraph": true, "conditional": true, "manager_loop": true,
}

// workflowSimpleBlocks maps workflow block keywords to their parser methods.
// Populated lazily to avoid init-order issues; see dispatchWorkflowBlock.

// dispatchWorkflowField routes a workflow-level identifier to the right handler.
func (p *Parser) dispatchWorkflowField(t Token) {
	if dispatchWorkflowSimpleField(p, t) {
		return
	}
	p.dispatchWorkflowBlock(t)
}

// dispatchWorkflowSimpleField handles header fields and config blocks (defaults, vars). Returns true if handled.
func dispatchWorkflowSimpleField(p *Parser, t Token) bool {
	switch t.Value {
	case "goal", "start", "exit":
		p.parseWorkflowStringField(t)
	case "defaults":
		p.parseDefaults()
	case "vars":
		p.parseVars()
	default:
		return dispatchWorkflowTailField(p, t)
	}
	return true
}

// dispatchWorkflowTailField handles edges, stylesheet, and requires. Returns true if handled.
func dispatchWorkflowTailField(p *Parser, t Token) bool {
	switch t.Value {
	case "edges":
		p.parseEdges()
	case "stylesheet":
		p.parseStylesheet()
	case "requires":
		p.parseWorkflowRequiresField(t)
	default:
		return false
	}
	return true
}

// dispatchWorkflowBlock handles parallel, fan_in, node kinds, and unknown identifiers.
func (p *Parser) dispatchWorkflowBlock(t Token) {
	switch t.Value {
	case "parallel":
		p.parseParallel()
	case "fan_in":
		p.parseFanIn()
	default:
		p.dispatchWorkflowDefault(t)
	}
}

// dispatchWorkflowDefault handles node kinds and unknown identifiers.
func (p *Parser) dispatchWorkflowDefault(t Token) {
	if workflowNodeKinds[t.Value] {
		p.parseNode(ir.NodeKind(t.Value))
		return
	}
	p.diagnostics = append(p.diagnostics, fmt.Sprintf("unexpected top-level identifier: %s at %d:%d", t.Value, t.Location.Line, t.Location.Column))
	p.lexer.NextToken()
}

// parseWorkflowRequiresField parses "requires: a, b, c" into Workflow.Requires.
// Whitespace is trimmed and empty entries are dropped. A missing or empty list
// leaves Workflow.Requires nil (matches IR nil-vs-empty conventions).
func (p *Parser) parseWorkflowRequiresField(t Token) {
	p.lexer.NextToken() // requires
	p.expect(TokenColon)
	val := p.readFieldValue(t.Location.Line)
	p.workflow.Requires = splitCommaNoEmpty(val)
}

// parseWorkflowStringField parses a simple "key: value" field on the workflow.
func (p *Parser) parseWorkflowStringField(t Token) {
	p.lexer.NextToken()
	p.expect(TokenColon)
	val := p.readFieldValue(t.Location.Line)
	switch t.Value {
	case "goal":
		p.workflow.Goal = val
	case "start":
		p.workflow.Start = val
	case "exit":
		p.workflow.Exit = val
	}
}

func (p *Parser) expect(t TokenType) {
	tok := p.lexer.NextToken()
	if tok.Type != t {
		p.diagnostics = append(p.diagnostics, fmt.Sprintf("expected %v, got %v at %d:%d", t, tok.Type, tok.Location.Line, tok.Location.Column))
	}
}

func (p *Parser) parseCommaList() []string {
	var list []string
	for {
		list = append(list, p.lexer.NextToken().Value)
		if p.lexer.PeekToken().Type != TokenComma {
			break
		}
		p.lexer.NextToken() // comma
	}
	return list
}

// readFieldValue reads a field value, which may be:
// - A raw block (multiline content detected by the lexer)
// - A single-line value on the same line as the key
// - A newline followed by a raw block (key: \n <indented block>)
func (p *Parser) readFieldValue(lineNum int) string {
	if p.lexer.PeekToken().Type == TokenRawBlock {
		return p.lexer.NextToken().Value
	}
	if p.lexer.PeekToken().Type == TokenNewline {
		return p.readBlockAfterNewline()
	}
	return p.readSingleLineValue(lineNum)
}

// readBlockAfterNewline consumes a newline and checks for a raw block after it.
func (p *Parser) readBlockAfterNewline() string {
	p.lexer.NextToken() // consume newline
	if p.lexer.PeekToken().Type == TokenRawBlock {
		return p.lexer.NextToken().Value
	}
	return ""
}

// readSingleLineValue reads a single-line value using raw extraction.
func (p *Parser) readSingleLineValue(lineNum int) string {
	raw := p.lexer.RawValueText(lineNum)
	p.consumeUntilNewline()
	return unquoteRaw(raw)
}

// consumeUntilNewline consumes tokens until a newline or EOF is reached.
func (p *Parser) consumeUntilNewline() {
	for p.lexer.PeekToken().Type != TokenNewline && p.lexer.PeekToken().Type != TokenEOF {
		p.lexer.NextToken()
	}
}

// unquoteRaw unquotes a quoted string. Double quotes process \" and \\ escapes;
// single quotes are YAML-style literals where a doubled single-quote escapes (no
// backslash processing). A value that is not wrapped in matching quotes is
// returned unchanged.
func unquoteRaw(raw string) string {
	if len(raw) < 2 || raw[0] != raw[len(raw)-1] {
		return raw
	}
	inner := raw[1 : len(raw)-1]
	switch raw[0] {
	case '"':
		inner = strings.ReplaceAll(inner, `\"`, `"`)
		return strings.ReplaceAll(inner, `\\`, `\`)
	case '\'':
		return strings.ReplaceAll(inner, `''`, `'`)
	}
	return raw
}
