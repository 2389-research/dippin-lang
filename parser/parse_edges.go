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
		p.applyEdgeAttribute(edge, attr.Value)
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
func (p *Parser) applyEdgeAttribute(edge *ir.Edge, attrName string) {
	switch attrName {
	case "when":
		edge.Condition = &ir.Condition{Raw: p.readConditionRaw()}
	case "label":
		p.expect(TokenColon)
		edge.Label = p.lexer.NextToken().Value
	case "weight":
		p.expect(TokenColon)
		wt := p.lexer.NextToken()
		edge.Weight = p.parseInt(wt.Value, "weight", wt.Location)
	default:
		p.applyEdgeBoolAttribute(edge, attrName)
	}
}

// applyEdgeBoolAttribute applies the boolean edge attributes (restart, override),
// each of the form "<name>: true|false". Carried, not interpreted.
func (p *Parser) applyEdgeBoolAttribute(edge *ir.Edge, attrName string) {
	switch attrName {
	case "restart":
		p.expect(TokenColon)
		edge.Restart = (p.lexer.NextToken().Value == "true")
	case "override":
		p.expect(TokenColon)
		edge.Override = (p.lexer.NextToken().Value == "true")
	}
}

// readConditionRaw reads tokens until a newline/EOF or a known edge attribute keyword.
func (p *Parser) readConditionRaw() string {
	var parts []string
	for p.lexer.PeekToken().Type != TokenNewline && p.lexer.PeekToken().Type != TokenEOF {
		pk := p.lexer.PeekToken()
		if edgeAttrKeywords[pk.Value] {
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
