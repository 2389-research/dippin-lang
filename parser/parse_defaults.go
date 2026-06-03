package parser

import (
	"fmt"

	"github.com/2389-research/dippin-lang/ir"
)

func (p *Parser) parseDefaults() {
	p.lexer.NextToken() // defaults
	p.expect(TokenNewline)
	p.expect(TokenIndent)
	p.parseDefaultsBody()
	p.expect(TokenOutdent)
}

// parseDefaultsBody parses the indented body of a defaults block.
func (p *Parser) parseDefaultsBody() {
	for p.lexer.PeekToken().Type != TokenOutdent && p.lexer.PeekToken().Type != TokenEOF {
		t := p.lexer.PeekToken()
		if t.Type == TokenNewline {
			p.lexer.NextToken()
			continue
		}
		if t.Type == TokenIdentifier {
			p.parseDefaultField(t)
		} else {
			p.lexer.NextToken()
		}
	}
}

// parseDefaultField reads a single default field (key: value) and applies it.
func (p *Parser) parseDefaultField(t Token) {
	key := t.Value
	p.lexer.NextToken()
	p.expect(TokenColon)
	val := p.readFieldValue(t.Location.Line)
	p.applyDefaultField(key, val, t.Location)
}

// applyDefaultField applies a single default field value to the workflow defaults.
func (p *Parser) applyDefaultField(key, val string, loc ir.SourceLocation) {
	if p.applyDefaultStringField(key, val) {
		return
	}
	p.applyDefaultComplexField(key, val, loc)
}

// applyDefaultStringField handles simple string assignments for defaults.
func (p *Parser) applyDefaultStringField(key, val string) bool {
	if applyDefaultCoreField(&p.workflow.Defaults, key, val) {
		return true
	}
	if applyDefaultExtraField(&p.workflow.Defaults, key, val) {
		return true
	}
	return applyDefaultToolField(&p.workflow.Defaults, key, val)
}

// applyDefaultCoreField handles model, provider, retry_policy defaults.
func applyDefaultCoreField(d *ir.WorkflowDefaults, key, val string) bool {
	switch key {
	case "model":
		d.Model = val
	case "provider":
		d.Provider = val
	case "retry_policy":
		d.RetryPolicy = val
	default:
		return false
	}
	return true
}

// applyDefaultExtraField handles fidelity, restart_target, compaction, on_resume defaults.
func applyDefaultExtraField(d *ir.WorkflowDefaults, key, val string) bool {
	switch key {
	case "fidelity":
		d.Fidelity = val
	case "restart_target":
		d.RestartTarget = val
	case "compaction":
		d.Compaction = val
	case "on_resume":
		d.OnResume = val
	default:
		return false
	}
	return true
}

// applyDefaultToolField handles tool-safety defaults: tool_commands_allow and
// tool_denylist_add. Values are stored verbatim; the runtime owns split/glob semantics.
func applyDefaultToolField(d *ir.WorkflowDefaults, key, val string) bool {
	switch key {
	case "tool_commands_allow":
		d.ToolCommandsAllow = val
	case "tool_denylist_add":
		d.ToolDenylistAdd = val
	default:
		return false
	}
	return true
}

// applyDefaultComplexField handles fields needing parsing for defaults.
func (p *Parser) applyDefaultComplexField(key, val string, loc ir.SourceLocation) {
	switch key {
	case "max_retries":
		p.workflow.Defaults.MaxRetries = p.parseInt(val, key, loc)
	case "max_restarts":
		p.workflow.Defaults.MaxRestarts = p.parseInt(val, key, loc)
	case "cache_tools":
		p.workflow.Defaults.CacheTools = (val == "true")
	default:
		p.applyDefaultBudgetField(key, val, loc)
	}
}

// applyDefaultBudgetField handles budget-related default fields.
func (p *Parser) applyDefaultBudgetField(key, val string, loc ir.SourceLocation) {
	switch key {
	case "max_total_tokens":
		p.workflow.Defaults.MaxTotalTokens = p.parseInt(val, key, loc)
	case "max_cost_cents":
		p.workflow.Defaults.MaxCostCents = p.parseInt(val, key, loc)
	case "max_wall_time":
		p.workflow.Defaults.MaxWallTime = p.parseDuration(val, key, loc)
	default:
		p.diagnostics = append(p.diagnostics,
			fmt.Sprintf("unknown defaults field %q at %d:%d", key, loc.Line, loc.Column))
	}
}
