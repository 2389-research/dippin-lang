// ABOUTME: Parses the workflow-level `inputs` block — the callee-side signature
// ABOUTME: declaring what a caller must supply. See issue #190.
package parser

import (
	"fmt"

	"github.com/2389-research/dippin-lang/ir"
)

// parseInputs parses the `inputs` section. Each entry is `name: type`, optionally
// extended by an indented attribute block (added in a later task).
func (p *Parser) parseInputs() {
	p.lexer.NextToken() // "inputs"
	p.expect(TokenNewline)
	p.expect(TokenIndent)
	p.parseInputsBody()
	p.expect(TokenOutdent)
}

// parseInputsBody scans input declarations until the block's outdent.
func (p *Parser) parseInputsBody() {
	for p.lexer.PeekToken().Type != TokenOutdent && p.lexer.PeekToken().Type != TokenEOF {
		t := p.lexer.PeekToken()
		if t.Type == TokenNewline {
			p.lexer.NextToken()
			continue
		}
		if t.Type == TokenIdentifier {
			p.appendInput(p.parseOneInput(t))
			continue
		}
		p.diagnostics = append(p.diagnostics,
			fmt.Sprintf("unexpected token in inputs block at %d:%d", t.Location.Line, t.Location.Column))
		p.lexer.NextToken()
	}
}

// appendInput adds a parsed input, diagnosing a duplicate name. The duplicate is
// still appended so the formatter round-trips the source as written.
func (p *Parser) appendInput(in *ir.Input) {
	if p.workflow.Input(in.Name) != nil {
		p.diagnostics = append(p.diagnostics,
			fmt.Sprintf("duplicate input %q at %d:%d", in.Name, in.Source.Line, in.Source.Column))
	}
	p.workflow.Inputs = append(p.workflow.Inputs, in)
}

// parseOneInput parses `name: type` and any indented attribute block.
func (p *Parser) parseOneInput(t Token) *ir.Input {
	p.lexer.NextToken() // name
	p.expect(TokenColon)
	typ := p.readFieldValue(t.Location.Line)
	return &ir.Input{Name: t.Value, Type: typ, Source: t.Location}
}
