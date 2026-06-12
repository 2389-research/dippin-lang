package parser

import (
	"fmt"
	"strings"

	"github.com/2389-research/dippin-lang/ir"
)

func (p *Parser) parseEdges() {
	p.lexer.NextToken() // edges
	p.expect(TokenNewline)
	p.expect(TokenIndent)
	p.parseEdgesBody()
	p.expect(TokenOutdent)
}

// parseEdgesBody parses the indented body of an edges block.
func (p *Parser) parseEdgesBody() {
	for p.lexer.PeekToken().Type != TokenOutdent && p.lexer.PeekToken().Type != TokenEOF {
		t := p.lexer.PeekToken()
		if t.Type == TokenNewline {
			p.lexer.NextToken()
			continue
		}
		p.parseSingleEdge()
	}
}

// parseSingleEdge parses a single edge declaration: "from -> to [attributes...]"
func (p *Parser) parseSingleEdge() {
	from := p.lexer.NextToken().Value
	p.expect(TokenArrow)
	to := p.lexer.NextToken().Value
	edge := &ir.Edge{From: from, To: to}
	p.parseEdgeAttributes(edge)
	p.workflow.Edges = append(p.workflow.Edges, edge)
	p.expect(TokenNewline)
}

// parseEdgeAttributes parses optional attributes (when, label, weight, restart) on an edge.
func (p *Parser) parseEdgeAttributes(edge *ir.Edge) {
	if p.lexer.PeekToken().Type == TokenLBracket {
		p.emitBracketSyntaxError()
		return
	}
	for p.lexer.PeekToken().Type != TokenNewline && p.lexer.PeekToken().Type != TokenEOF {
		attr := p.lexer.NextToken()
		p.applyEdgeAttribute(edge, attr)
	}
}

// emitBracketSyntaxError emits a diagnostic for unsupported bracket syntax and skips to EOL.
func (p *Parser) emitBracketSyntaxError() {
	tok := p.lexer.NextToken() // consume '['
	p.diagnostics = append(p.diagnostics, fmt.Sprintf(
		"bracket syntax [label: ...] is not supported at %d:%d; use keyword syntax instead (e.g., when ctx.x = 1  label: go)",
		tok.Location.Line, tok.Location.Column,
	))
	p.consumeUntilNewline()
}

// edgeAttrKeywords contains the set of edge attribute keywords that terminate condition parsing.
var edgeAttrKeywords = map[string]bool{
	"label": true, "weight": true, "restart": true, "override": true,
}

// applyEdgeAttribute applies a single edge attribute.
func (p *Parser) applyEdgeAttribute(edge *ir.Edge, attr Token) {
	switch attr.Value {
	case "when":
		edge.Condition = &ir.Condition{Raw: p.readConditionRaw()}
	case "on":
		p.applyOnAttribute(edge, attr)
	case "label":
		p.expect(TokenColon)
		edge.Label = p.lexer.NextToken().Value
	case "weight":
		p.expect(TokenColon)
		wt := p.lexer.NextToken()
		edge.Weight = p.parseInt(wt.Value, "weight", wt.Location)
	default:
		p.applyEdgeBoolAttribute(edge, attr)
	}
}

// applyOnAttribute desugars the `on <token>` shorthand into an equality test
// against the source node's natural outcome channel: ctx.outcome for agent
// nodes, ctx.tool_marker for tool nodes that declare marker_grep. It produces
// the same ir.Condition as the equivalent `when`. A source node with no defined
// outcome channel (human gate, conditional, marker-less tool) is a located
// diagnostic suggesting `when`.
func (p *Parser) applyOnAttribute(edge *ir.Edge, attr Token) {
	channel, ok := p.workflow.Node(edge.From).OutcomeChannel()
	if !ok {
		p.diagnostics = append(p.diagnostics, fmt.Sprintf(
			"`on` shorthand at %d:%d requires a source node with a defined outcome "+
				"channel (agent, or tool with marker_grep); use `when` instead",
			attr.Location.Line, attr.Location.Column,
		))
		p.consumeOptionalValue()
		return
	}
	if t := p.lexer.PeekToken().Type; t == TokenNewline || t == TokenEOF {
		p.diagnostics = append(p.diagnostics, fmt.Sprintf(
			"`on` at %d:%d requires an outcome token (e.g. `on success`)",
			attr.Location.Line, attr.Location.Column,
		))
		return
	}
	edge.Condition = &ir.Condition{Raw: channel + " = " + p.lexer.NextToken().Value}
}

// consumeOptionalValue consumes a single value token following an attribute when
// one is present, so the per-token attribute loop stays aligned after an error.
func (p *Parser) consumeOptionalValue() {
	if t := p.lexer.PeekToken().Type; t != TokenNewline && t != TokenEOF {
		p.lexer.NextToken()
	}
}

// applyEdgeBoolAttribute applies the boolean edge attributes (restart, override),
// each of the form "<name>: true|false". Carried, not interpreted. An
// unrecognized attribute is diagnosed once rather than silently swallowed.
func (p *Parser) applyEdgeBoolAttribute(edge *ir.Edge, attr Token) {
	switch attr.Value {
	case "restart":
		p.expect(TokenColon)
		edge.Restart = (p.lexer.NextToken().Value == "true")
	case "override":
		p.expect(TokenColon)
		edge.Override = (p.lexer.NextToken().Value == "true")
	default:
		p.emitUnknownEdgeAttribute(attr)
	}
}

// emitUnknownEdgeAttribute records a located diagnostic for an unrecognized edge
// attribute and consumes its optional ": value" payload, so the per-token
// attribute loop diagnoses each unknown attribute exactly once.
func (p *Parser) emitUnknownEdgeAttribute(attr Token) {
	p.diagnostics = append(p.diagnostics, fmt.Sprintf(
		"unknown edge attribute %q at %d:%d",
		attr.Value, attr.Location.Line, attr.Location.Column,
	))
	if p.lexer.PeekToken().Type != TokenColon {
		return
	}
	p.lexer.NextToken() // consume ':'
	if t := p.lexer.PeekToken().Type; t != TokenNewline && t != TokenEOF {
		p.lexer.NextToken() // consume value (only when present)
	}
}

// readConditionRaw reads tokens until a newline/EOF or a known edge attribute
// keyword. An attribute keyword only terminates the condition when it is
// followed by ':' (the shape of an attribute); a bare keyword on the right-hand
// side of a condition (e.g. "when ctx.reason = override") is part of the value.
func (p *Parser) readConditionRaw() string {
	var parts []string
	for p.lexer.PeekToken().Type != TokenNewline && p.lexer.PeekToken().Type != TokenEOF {
		pk := p.lexer.PeekToken()
		if edgeAttrKeywords[pk.Value] && p.lexer.PeekTokenN(1).Type == TokenColon {
			break
		}
		t := p.lexer.NextToken()
		parts = append(parts, formatConditionToken(t))
	}
	return strings.TrimSpace(strings.Join(parts, " "))
}

// formatConditionToken formats a single token for raw condition text.
func formatConditionToken(t Token) string {
	if t.Type == TokenLiteral {
		return "\"" + t.Value + "\""
	}
	return t.Value
}
