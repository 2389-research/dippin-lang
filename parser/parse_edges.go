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
		p.parseEdgeOrElse(t)
	}
}

// parseEdgeOrElse dispatches a non-blank edges-block entry: the section-level
// `else -> <node>` default, or an ordinary `from -> to` edge.
func (p *Parser) parseEdgeOrElse(t Token) {
	if t.Type == TokenIdentifier && t.Value == "else" {
		p.parseElseDefault()
		return
	}
	p.parseSingleEdge()
}

// parseElseDefault parses a section-level `else -> <node>` entry — the graph's
// success-side default destination. It populates Workflow.ElseTarget rather than
// appending an ir.Edge (the funnel default has no source node, so it is not a
// normal edge). A second `else` in the same block is diagnosed and ignored.
func (p *Parser) parseElseDefault() {
	tok := p.lexer.NextToken() // 'else'
	p.expect(TokenArrow)
	to := p.lexer.NextToken().Value
	p.consumeUntilNewline() // tolerate (and discard) any trailing tokens; else takes no attributes
	p.expect(TokenNewline)
	if p.workflow.ElseTarget != "" {
		p.diagnostics = append(p.diagnostics, fmt.Sprintf(
			"duplicate `else` default at %d:%d; an edges block may declare at most one `else`",
			tok.Location.Line, tok.Location.Column,
		))
		return
	}
	p.workflow.ElseTarget = to
	p.workflow.ElseTargetSource = tok.Location
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

// edgeAttrKeywords contains the value-bearing edge attribute keywords (each of the
// form "<name>: value"). They terminate condition parsing only when followed by ':'.
var edgeAttrKeywords = map[string]bool{
	"label": true, "choice": true, "weight": true, "restart": true, "override": true,
}

// applyEdgeAttribute dispatches a single edge attribute by its structural form:
// a `when` condition, an `on` shorthand, a value-less `loop` flag, or — by
// default — a uniform "<keyword>: value" attribute.
func (p *Parser) applyEdgeAttribute(edge *ir.Edge, attr Token) {
	switch attr.Value {
	case "when":
		edge.Condition = &ir.Condition{Raw: p.readConditionRaw()}
	case "on":
		p.applyOnAttribute(edge, attr)
	case "loop":
		p.applyLoopAttribute(edge)
	default:
		p.applyEdgeValueAttribute(edge, attr)
	}
}

// applyLoopAttribute applies the value-less `loop` flag — a synonym for `restart:
// true` that sets the same ir.Edge.Restart field and consumes no colon or value.
// A `loop: <value>` written in the legacy boolean shape is a mistake: the flag is
// still set, but the stray `: value` is diagnosed once with a clear hint rather
// than leaking as spurious "unknown edge attribute" errors for `:` and the value.
func (p *Parser) applyLoopAttribute(edge *ir.Edge) {
	edge.Restart = true
	if p.lexer.PeekToken().Type != TokenColon {
		return
	}
	colon := p.lexer.NextToken() // consume ':'
	p.diagnostics = append(p.diagnostics, fmt.Sprintf(
		"`loop` takes no value at %d:%d; write `loop` (no colon)",
		colon.Location.Line, colon.Location.Column,
	))
	if t := p.lexer.PeekToken().Type; t != TokenNewline && t != TokenEOF {
		p.lexer.NextToken() // consume the stray value
	}
}

// applyOnAttribute desugars the `on <token>` shorthand into an equality test
// against the source node's natural outcome channel: ctx.outcome for agent
// nodes, ctx.tool_marker for tool nodes that declare marker_grep. It produces
// the same ir.Condition as the equivalent `when`. A known source node with no
// defined outcome channel (human gate, conditional, marker-less tool), or a
// value that is not a bare identifier, is a located diagnostic suggesting `when`.
func (p *Parser) applyOnAttribute(edge *ir.Edge, attr Token) {
	node := p.workflow.Node(edge.From)
	if node == nil {
		// Unknown source node: stay silent here and let the validator's DIP003
		// report it clearly, exactly as a `when` edge does — emitting a channel
		// diagnostic would be misleading and would fail the parse prematurely.
		p.discardOnValue()
		return
	}
	channel, ok := node.OutcomeChannel()
	if !ok {
		p.diagnostics = append(p.diagnostics, fmt.Sprintf(
			"`on` shorthand at %d:%d requires a source node with a defined outcome "+
				"channel (agent, or tool with marker_grep); use `when` instead",
			attr.Location.Line, attr.Location.Column,
		))
		p.discardOnValue()
		return
	}
	if raw, ok := p.readOnValue(channel, attr); ok {
		edge.Condition = &ir.Condition{Raw: channel + " = " + raw}
	}
}

// readOnValue consumes and validates the `on` shorthand's value token, returning
// the bare identifier and true on success. The value must be a single bare
// identifier so the shorthand re-parses; a quoted literal, or one the lexer
// splits (e.g. `a:b` → a · : · b), belongs in a full `when`. Glued fragments are
// joined first so a split value is diagnosed once as a whole rather than leaking
// its tail to the attribute loop. On any malformed value it emits a located
// diagnostic suggesting `when` and returns false.
func (p *Parser) readOnValue(channel string, attr Token) (string, bool) {
	if t := p.lexer.PeekToken().Type; t == TokenNewline || t == TokenEOF {
		p.diagnostics = append(p.diagnostics, fmt.Sprintf(
			"`on` at %d:%d requires an outcome token (e.g. `on success`)",
			attr.Location.Line, attr.Location.Column,
		))
		return "", false
	}
	val := p.lexer.NextToken()
	raw := p.joinGluedTokens(val)
	if val.Type != TokenIdentifier || !ir.IsOutcomeToken(raw) {
		p.diagnostics = append(p.diagnostics, fmt.Sprintf(
			"`on` value %q at %d:%d must be a bare identifier ([A-Za-z0-9][A-Za-z0-9_-]*); "+
				"use `when %s = ...` for other values",
			raw, val.Location.Line, val.Location.Column, channel,
		))
		return "", false
	}
	return raw, true
}

// joinGluedTokens returns val's text concatenated with any immediately following
// tokens that carry no intervening whitespace — the fragments a lexer-split
// value (e.g. `a:b` → a · : · b) decomposes into — consuming them so a malformed
// `on` value is reported once as the whole intended token.
func (p *Parser) joinGluedTokens(val Token) string {
	raw := val.Value
	end := val.Location.Column + len(val.Value)
	for p.tokenGluedAt(val.Location.Line, end) {
		t := p.lexer.NextToken()
		raw += t.Value
		end += len(t.Value)
	}
	return raw
}

// tokenGluedAt reports whether the next token abuts (line, col) with no
// whitespace gap. Newline/EOF never count as glued.
func (p *Parser) tokenGluedAt(line, col int) bool {
	next := p.lexer.PeekToken()
	return next.Type != TokenNewline && next.Type != TokenEOF &&
		next.Location.Line == line && next.Location.Column == col
}

// discardOnValue consumes an `on` value — including any glued fragments of a
// lexer-split value (e.g. `a:b` → a · : · b) — when one is present, so an `on`
// used where it isn't valid doesn't leave token fragments behind to cascade as
// spurious "unknown edge attribute" diagnostics.
func (p *Parser) discardOnValue() {
	if t := p.lexer.PeekToken().Type; t == TokenNewline || t == TokenEOF {
		return
	}
	p.joinGluedTokens(p.lexer.NextToken())
}

// applyEdgeValueAttribute applies the uniform "<keyword>: value" edge attributes:
// the strings `label` and `choice` (the latter a carried, not interpreted
// human-gate routing key), the integer `weight`, and the booleans `restart` and
// `override` (carried, not interpreted). `restart` is the legacy spelling of the
// `loop` flag. An unrecognized attribute is diagnosed once rather than silently
// swallowed.
func (p *Parser) applyEdgeValueAttribute(edge *ir.Edge, attr Token) {
	switch attr.Value {
	case "label", "choice":
		p.applyEdgeStringAttribute(edge, attr.Value)
	case "weight":
		p.expect(TokenColon)
		wt := p.lexer.NextToken()
		edge.Weight = p.parseInt(wt.Value, "weight", wt.Location)
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

// applyEdgeStringAttribute reads a "<name>: value" string edge attribute and
// stores it on the matching field: `label` (display text) or `choice` (the
// carried, not interpreted human-gate routing key).
func (p *Parser) applyEdgeStringAttribute(edge *ir.Edge, name string) {
	p.expect(TokenColon)
	v := p.lexer.NextToken().Value
	if name == "choice" {
		edge.Choice = v
		return
	}
	edge.Label = v
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

// readConditionRaw reads tokens until a newline/EOF or a terminating edge
// attribute keyword (see conditionTerminates): a value-bearing keyword followed
// by ':', or a value-less keyword unconditionally. A bare value-bearing keyword
// on the right-hand side (e.g. "when ctx.reason = override") is part of the value.
func (p *Parser) readConditionRaw() string {
	var parts []string
	for p.lexer.PeekToken().Type != TokenNewline && p.lexer.PeekToken().Type != TokenEOF {
		if p.conditionTerminates(p.lexer.PeekToken()) {
			break
		}
		t := p.lexer.NextToken()
		parts = append(parts, formatConditionToken(t))
	}
	return strings.TrimSpace(strings.Join(parts, " "))
}

// conditionTerminates reports whether the upcoming token ends the current
// condition: a value-less edge keyword (e.g. `loop`) terminates unconditionally,
// while a value-bearing keyword (e.g. `restart`) terminates only when followed by
// ':' — a bare such keyword on a condition's RHS is part of the value (#126).
func (p *Parser) conditionTerminates(pk Token) bool {
	if pk.Type == TokenIdentifier && ir.IsReservedEdgeFlag(pk.Value) {
		return true
	}
	return edgeAttrKeywords[pk.Value] && p.lexer.PeekTokenN(1).Type == TokenColon
}

// formatConditionToken formats a single token for raw condition text.
func formatConditionToken(t Token) string {
	if t.Type == TokenLiteral {
		return "\"" + t.Value + "\""
	}
	return t.Value
}
