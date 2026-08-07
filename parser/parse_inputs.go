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

// parseOneInput parses `name: type` plus any indented attribute block.
func (p *Parser) parseOneInput(t Token) *ir.Input {
	p.lexer.NextToken() // name
	p.expect(TokenColon)
	typ := p.readFieldValue(t.Location.Line)
	in := &ir.Input{Name: t.Value, Type: typ, Source: t.Location}

	if p.lexer.PeekToken().Type == TokenNewline {
		p.lexer.NextToken()
	}
	if p.lexer.PeekToken().Type != TokenIndent {
		return in
	}
	p.expect(TokenIndent)
	p.parseInputFields(in)
	p.expect(TokenOutdent)
	return in
}

// parseInputFields parses the attributes inside one input's indented block.
func (p *Parser) parseInputFields(in *ir.Input) {
	for p.lexer.PeekToken().Type != TokenOutdent && p.lexer.PeekToken().Type != TokenEOF {
		t := p.lexer.PeekToken()
		if t.Type == TokenNewline {
			p.lexer.NextToken()
			continue
		}
		if t.Type != TokenIdentifier {
			p.lexer.NextToken()
			continue
		}
		p.lexer.NextToken() // key
		p.expect(TokenColon)
		val := p.readFieldValue(t.Location.Line)
		p.applyInputField(in, t.Value, val, t.Location)
	}
}

// applyInputField dispatches one attribute, hinting on an unrecognized key.
// Unknown attributes are a hint, never a parse failure — see issue #190's
// forward-compatibility requirement.
func (p *Parser) applyInputField(in *ir.Input, key, val string, loc ir.SourceLocation) {
	if applyInputTextField(in, key, val) {
		return
	}
	if applyInputValueField(in, key, val) {
		return
	}
	if p.applyInputParsedField(in, key, val, loc) {
		return
	}
	p.emitUnknownFieldHint("input", key, loc)
}

// applyInputTextField handles plain string attributes.
func applyInputTextField(in *ir.Input, key, val string) bool {
	switch key {
	case "prompt":
		in.Prompt = val
	case "description":
		in.Description = val
	case "pattern":
		in.Pattern = val
	default:
		return false
	}
	return true
}

// applyInputValueField handles the default and the constraints kept as raw text.
func applyInputValueField(in *ir.Input, key, val string) bool {
	switch key {
	case "default":
		in.Default = val
		in.HasDefault = true
	case "options":
		in.Options = splitCommaNoEmpty(val)
	case "min":
		in.Min = val
	case "max":
		in.Max = val
	default:
		return false
	}
	return true
}

// applyInputParsedField handles attributes needing conversion.
func (p *Parser) applyInputParsedField(in *ir.Input, key, val string, loc ir.SourceLocation) bool {
	switch key {
	case "required":
		in.Required = p.parseBoolAttr(val, key, loc)
	case "multiline":
		in.Multiline = p.parseBoolAttr(val, key, loc)
	case "max_length":
		in.MaxLength = p.parseInt(val, key, loc)
	default:
		return false
	}
	return true
}
