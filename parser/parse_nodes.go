package parser

import (
	"fmt"
	"strings"

	"github.com/2389-research/dippin-lang/ir"
)

// defaultNodeConfigs maps node kinds to their zero config constructors.
var defaultNodeConfigs = map[ir.NodeKind]func() ir.NodeConfig{
	ir.NodeAgent:       func() ir.NodeConfig { return ir.AgentConfig{Params: make(map[string]string)} },
	ir.NodeHuman:       func() ir.NodeConfig { return ir.HumanConfig{} },
	ir.NodeTool:        func() ir.NodeConfig { return ir.ToolConfig{} },
	ir.NodeSubgraph:    func() ir.NodeConfig { return ir.SubgraphConfig{Params: make(map[string]string)} },
	ir.NodeConditional: func() ir.NodeConfig { return ir.ConditionalConfig{} },
	ir.NodeManagerLoop: func() ir.NodeConfig { return ir.ManagerLoopConfig{SteerContext: make(map[string]string)} },
}

// defaultNodeConfig returns the zero config for a given node kind.
func defaultNodeConfig(kind ir.NodeKind) ir.NodeConfig {
	if fn, ok := defaultNodeConfigs[kind]; ok {
		return fn()
	}
	return ir.AgentConfig{}
}

func (p *Parser) parseNode(kind ir.NodeKind) {
	p.lexer.NextToken() // kind
	id := p.lexer.NextToken().Value
	node := &ir.Node{
		ID:     id,
		Kind:   kind,
		Source: p.lexer.PeekToken().Location,
		Config: defaultNodeConfig(kind),
	}

	p.expect(TokenNewline)
	p.expect(TokenIndent)
	p.parseNodeBody(node)
	p.expect(TokenOutdent)
	p.workflow.Nodes = append(p.workflow.Nodes, node)
}

// parseNodeBody parses the indented fields within a node declaration.
func (p *Parser) parseNodeBody(node *ir.Node) {
	for p.lexer.PeekToken().Type != TokenOutdent && p.lexer.PeekToken().Type != TokenEOF {
		p.parseNodeBodyLine(node)
	}
}

// parseNodeBodyLine processes one token (or logical line) inside a node body.
func (p *Parser) parseNodeBodyLine(node *ir.Node) {
	t := p.lexer.PeekToken()
	if t.Type == TokenNewline {
		p.lexer.NextToken()
		return
	}
	if t.Type == TokenIdentifier && t.Value == "retry" {
		p.emitNestedRetryError(t.Location)
		return
	}
	if t.Type == TokenIdentifier {
		p.parseNodeField(node, t)
		return
	}
	p.lexer.NextToken()
}

// parseNodeField parses a single key: value field in a node body.
func (p *Parser) parseNodeField(node *ir.Node, t Token) {
	key := t.Value
	p.lexer.NextToken()
	p.expect(TokenColon)
	val := p.readFieldValue(t.Location.Line)
	p.applyNodeField(node, key, val, t.Location)
}

// emitUnknownFieldHint emits a diagnostic for an unrecognized field on a node.
func (p *Parser) emitUnknownFieldHint(kind, key string, loc ir.SourceLocation) {
	hint := fmt.Sprintf("unrecognized %s field %q at %d:%d", kind, key, loc.Line, loc.Column)
	if kind == "agent" || kind == "subgraph" {
		hint += " — did you mean to put it under params:?"
	}
	p.diagnostics = append(p.diagnostics, hint)
}

func (p *Parser) emitNestedRetryError(loc ir.SourceLocation) {
	p.diagnostics = append(p.diagnostics, fmt.Sprintf(
		"nested retry blocks are not supported; use flat attributes instead (retry_policy, max_retries, retry_target, fallback_target, base_delay) at %d:%d",
		loc.Line, loc.Column))
	p.lexer.NextToken() // consume "retry"
	p.consumeUntilNewline()
	p.skipIndentedBlock()
}

// skipIndentedBlock skips over an optional indented block following the current position.
func (p *Parser) skipIndentedBlock() {
	p.skipLeadingNewline()
	if p.lexer.PeekToken().Type != TokenIndent {
		return
	}
	p.lexer.NextToken() // consume indent
	p.consumeUntilOutdent()
}

func (p *Parser) skipLeadingNewline() {
	if p.lexer.PeekToken().Type == TokenNewline {
		p.lexer.NextToken()
	}
}

func (p *Parser) consumeUntilOutdent() {
	for p.lexer.PeekToken().Type != TokenOutdent && p.lexer.PeekToken().Type != TokenEOF {
		p.lexer.NextToken()
	}
	if p.lexer.PeekToken().Type == TokenOutdent {
		p.lexer.NextToken()
	}
}

func (p *Parser) applyNodeField(n *ir.Node, key, val string, loc ir.SourceLocation) {
	if p.tryApplyCommonField(n, key, val, loc) {
		return
	}
	p.applyConfigField(n, key, val, loc)
}

// applyConfigField dispatches to config-specific field handlers.
func (p *Parser) applyConfigField(n *ir.Node, key, val string, loc ir.SourceLocation) {
	if p.applyPrimaryConfigField(n, key, val, loc) {
		return
	}
	p.applySecondaryConfigField(n, key, val, loc)
}

// applyPrimaryConfigField handles agent and human config fields. Returns true if handled.
func (p *Parser) applyPrimaryConfigField(n *ir.Node, key, val string, loc ir.SourceLocation) bool {
	switch cfg := n.Config.(type) {
	case ir.AgentConfig:
		p.applyAgentField(&cfg, key, val, loc)
		n.Config = cfg
	case ir.HumanConfig:
		p.applyHumanField(&cfg, key, val, loc)
		n.Config = cfg
	default:
		return false
	}
	return true
}

// applySecondaryConfigField handles tool, subgraph, manager_loop, and conditional config fields.
func (p *Parser) applySecondaryConfigField(n *ir.Node, key, val string, loc ir.SourceLocation) {
	switch cfg := n.Config.(type) {
	case ir.ToolConfig:
		p.applyToolField(&cfg, key, val, loc)
		n.Config = cfg
	case ir.SubgraphConfig:
		p.applySubgraphField(&cfg, key, val, loc)
		n.Config = cfg
	case ir.ManagerLoopConfig:
		p.applyManagerLoopField(&cfg, key, val, loc)
		n.Config = cfg
	case ir.ConditionalConfig:
		p.emitUnknownFieldHint("conditional", key, loc)
	}
}

// tryApplyCommonField applies fields that are common to all node types.
// Returns true if the field was handled, false otherwise.
func (p *Parser) tryApplyCommonField(n *ir.Node, key, val string, loc ir.SourceLocation) bool {
	if applyCommonStringField(n, key, val) {
		return true
	}
	return p.applyCommonComplexField(n, key, val, loc)
}

// applyCommonStringField handles simple string/slice assignments for common fields.
func applyCommonStringField(n *ir.Node, key, val string) bool {
	if applyCommonPlainField(n, key, val) {
		return true
	}
	return applyCommonRetryField(n, key, val)
}

// applyCommonPlainField handles label, class, reads, writes.
func applyCommonPlainField(n *ir.Node, key, val string) bool {
	switch key {
	case "label":
		n.Label = val
	case "class":
		n.Classes = splitComma(val)
	case "reads":
		n.IO.Reads = splitComma(val)
	case "writes":
		n.IO.Writes = splitComma(val)
	default:
		return false
	}
	return true
}

// applyCommonRetryField handles retry-related string fields.
func applyCommonRetryField(n *ir.Node, key, val string) bool {
	switch key {
	case "retry_policy":
		n.Retry.Policy = val
	case "retry_target":
		n.Retry.RetryTarget = val
	case "fallback_target":
		n.Retry.FallbackTarget = val
	default:
		return false
	}
	return true
}

// applyCommonComplexField handles fields that need parsing (int, duration).
func (p *Parser) applyCommonComplexField(n *ir.Node, key, val string, loc ir.SourceLocation) bool {
	switch key {
	case "max_retries":
		n.Retry.MaxRetries = p.parseInt(val, key, loc)
	case "base_delay":
		n.Retry.BaseDelay = p.parseDuration(val, key, loc)
	default:
		return false
	}
	return true
}

// applyAgentField applies agent-specific configuration fields.
func (p *Parser) applyAgentField(cfg *ir.AgentConfig, key, val string, loc ir.SourceLocation) {
	if applyAgentStringField(cfg, key, val) {
		return
	}
	p.applyAgentComplexField(cfg, key, val, loc)
}

// applyAgentStringField handles simple string assignments for agent config.
func applyAgentStringField(cfg *ir.AgentConfig, key, val string) bool {
	if applyAgentPromptField(cfg, key, val) {
		return true
	}
	if applyAgentModelField(cfg, key, val) {
		return true
	}
	return applyAgentRuntimeField(cfg, key, val)
}

// applyAgentPromptField handles prompt-related agent fields.
func applyAgentPromptField(cfg *ir.AgentConfig, key, val string) bool {
	switch key {
	case "prompt":
		cfg.Prompt = val
	case "system_prompt":
		cfg.SystemPrompt = val
	case "reasoning_effort":
		cfg.ReasoningEffort = val
	case "response_schema":
		cfg.ResponseSchema = val
	default:
		return false
	}
	return true
}

// applyAgentModelField handles model-related agent fields.
func applyAgentModelField(cfg *ir.AgentConfig, key, val string) bool {
	switch key {
	case "model":
		cfg.Model = val
	case "provider":
		cfg.Provider = val
	case "fidelity":
		cfg.Fidelity = val
	case "response_format":
		cfg.ResponseFormat = val
	default:
		return false
	}
	return true
}

// applyAgentRuntimeField handles runtime behavior fields.
func applyAgentRuntimeField(cfg *ir.AgentConfig, key, val string) bool {
	switch key {
	case "backend":
		cfg.Backend = val
	case "working_dir":
		cfg.WorkingDir = val
	default:
		return false
	}
	return true
}

// applyAgentComplexField handles fields needing parsing for agent config.
func (p *Parser) applyAgentComplexField(cfg *ir.AgentConfig, key, val string, loc ir.SourceLocation) {
	if p.applyAgentBoolField(cfg, key, val, loc) {
		return
	}
	if key == "params" {
		cfg.Params = p.parseParamsBlock(val)
		return
	}
	if p.applyAgentParsedField(cfg, key, val, loc) {
		return
	}
	p.emitUnknownFieldHint("agent", key, loc)
}

// applyAgentBoolField handles boolean and string agent fields.
func (p *Parser) applyAgentBoolField(cfg *ir.AgentConfig, key, val string, loc ir.SourceLocation) bool {
	switch key {
	case "goal_gate":
		cfg.GoalGate = p.parseBoolAttr(val, key, loc)
	case "auto_status":
		cfg.AutoStatus = p.parseBoolAttr(val, key, loc)
	case "cache_tools":
		cfg.CacheTools = p.parseBoolAttr(val, key, loc)
	case "compaction":
		cfg.Compaction = val
	default:
		return false
	}
	return true
}

// applyAgentParsedField handles agent fields that require parsing.
// Returns true if the field was recognized.
func (p *Parser) applyAgentParsedField(cfg *ir.AgentConfig, key, val string, loc ir.SourceLocation) bool {
	switch key {
	case "max_turns":
		cfg.MaxTurns = p.parseInt(val, key, loc)
	case "compaction_threshold":
		cfg.CompactionThreshold = p.parseFloat(val, key, loc)
	case "cmd_timeout":
		cfg.CmdTimeout = p.parseDuration(val, key, loc)
	default:
		return false
	}
	return true
}

// applyHumanField applies human-specific configuration fields.
func (p *Parser) applyHumanField(cfg *ir.HumanConfig, key, val string, loc ir.SourceLocation) {
	if applyHumanStringField(cfg, key, val) {
		return
	}
	if applyHumanInterviewField(cfg, key, val) {
		return
	}
	if p.applyHumanComplexField(cfg, key, val, loc) {
		return
	}
	p.emitUnknownFieldHint("human", key, loc)
}

// applyHumanComplexField handles fields needing parsing for human config.
func (p *Parser) applyHumanComplexField(cfg *ir.HumanConfig, key, val string, loc ir.SourceLocation) bool {
	switch key {
	case "timeout":
		cfg.Timeout = p.parseDuration(val, key, loc)
	case "timeout_action":
		switch val {
		case "", "fail", "default":
			cfg.TimeoutAction = val
		default:
			p.diagnostics = append(p.diagnostics, fmt.Sprintf(
				"invalid timeout_action %q at %d:%d (use fail, default, or empty)",
				val, loc.Line, loc.Column))
		}
	default:
		return false
	}
	return true
}

func applyHumanStringField(cfg *ir.HumanConfig, key, val string) bool {
	switch key {
	case "mode":
		cfg.Mode = val
	case "default":
		cfg.Default = val
	case "prompt":
		cfg.Prompt = val
	default:
		return false
	}
	return true
}

func applyHumanInterviewField(cfg *ir.HumanConfig, key, val string) bool {
	switch key {
	case "questions_key":
		cfg.QuestionsKey = val
	case "answers_key":
		cfg.AnswersKey = val
	default:
		return false
	}
	return true
}

// applyToolField applies tool-specific configuration fields.
func (p *Parser) applyToolField(cfg *ir.ToolConfig, key, val string, loc ir.SourceLocation) {
	if applyToolStringField(cfg, key, val) {
		return
	}
	if p.applyToolBoolField(cfg, key, val, loc) {
		return
	}
	if p.applyToolParsedField(cfg, key, val, loc) {
		return
	}
	p.emitUnknownFieldHint("tool", key, loc)
}

// applyToolStringField handles string-valued tool fields. Returns true if handled.
func applyToolStringField(cfg *ir.ToolConfig, key, val string) bool {
	switch key {
	case "command":
		cfg.Command = val
	case "outputs":
		cfg.Outputs = splitComma(val)
	case "marker_grep":
		cfg.MarkerGrep = val
	default:
		return false
	}
	return true
}

// applyToolBoolField handles boolean tool fields. Returns true if handled.
func (p *Parser) applyToolBoolField(cfg *ir.ToolConfig, key, val string, loc ir.SourceLocation) bool {
	switch key {
	case "route_required":
		cfg.RouteRequired = p.parseBoolAttr(val, key, loc)
	default:
		return false
	}
	return true
}

// applyToolParsedField handles tool fields needing parsing. Returns true if handled.
func (p *Parser) applyToolParsedField(cfg *ir.ToolConfig, key, val string, loc ir.SourceLocation) bool {
	switch key {
	case "timeout":
		cfg.Timeout = p.parseDuration(val, key, loc)
	case "output_limit":
		n := p.parseInt(val, key, loc)
		if n < 0 {
			p.diagnostics = append(p.diagnostics, fmt.Sprintf(
				"invalid output_limit %d at %d:%d (must be non-negative)", n, loc.Line, loc.Column))
			return true
		}
		cfg.OutputLimit = n
	default:
		return false
	}
	return true
}

// applySubgraphField applies subgraph-specific configuration fields.
func (p *Parser) applySubgraphField(cfg *ir.SubgraphConfig, key, val string, loc ir.SourceLocation) {
	switch key {
	case "ref":
		cfg.Ref = val
	case "params":
		cfg.Params = p.parseParamsBlock(val)
	default:
		p.emitUnknownFieldHint("subgraph", key, loc)
	}
}

// applyManagerLoopField applies manager_loop-specific configuration fields.
func (p *Parser) applyManagerLoopField(cfg *ir.ManagerLoopConfig, key, val string, loc ir.SourceLocation) {
	if p.applyManagerLoopStringField(cfg, key, val) {
		return
	}
	if p.applyManagerLoopParsedField(cfg, key, val, loc) {
		return
	}
	p.emitUnknownFieldHint("manager_loop", key, loc)
}

// applyManagerLoopStringField handles string/condition fields. Returns true if handled.
func (p *Parser) applyManagerLoopStringField(cfg *ir.ManagerLoopConfig, key, val string) bool {
	switch key {
	case "subgraph_ref":
		cfg.SubgraphRef = val
	case "stop_condition":
		cfg.StopCondition = &ir.Condition{Raw: val}
	case "steer_condition":
		cfg.SteerCondition = &ir.Condition{Raw: val}
	default:
		return false
	}
	return true
}

// applyManagerLoopParsedField handles duration/int/map fields. Returns true if handled.
func (p *Parser) applyManagerLoopParsedField(cfg *ir.ManagerLoopConfig, key, val string, loc ir.SourceLocation) bool {
	switch key {
	case "poll_interval":
		cfg.PollInterval = p.parseDuration(val, key, loc)
	case "max_cycles":
		cfg.MaxCycles = p.parseInt(val, key, loc)
	case "steer_context":
		cfg.SteerContext = p.parseSteerContext(val)
	default:
		return false
	}
	return true
}

// parseSteerContext accepts both inline CSV ("k=v,k=v") and block-form content
// (one "k: v" per line, same as parseParamsBlock). The inline form splits on
// comma without quote-awareness — values containing commas MUST use the block
// form or they will be truncated at the first comma.
//
// Disambiguates forms by looking at the first separator: ":" → block form
// (including single-entry block which has no embedded newline), "=" → inline.
//
// Post-parse: keys containing ':' are dropped with a diagnostic because the
// formatter emits block form ("key: value") and parseParamsBlock splits on the
// first colon — a colon in the key breaks the .dip-internal round-trip.
// ',' and '=' are percent-encoded at the DOT export/migrate boundary, so they
// no longer need to be rejected here.
func (p *Parser) parseSteerContext(raw string) map[string]string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return map[string]string{}
	}
	var result map[string]string
	if isSteerContextBlockForm(raw) {
		result = p.parseParamsBlock(raw)
	} else {
		result = p.parseSteerContextInline(raw)
	}
	for k := range result {
		if strings.Contains(k, ":") {
			p.diagnostics = append(p.diagnostics,
				fmt.Sprintf("steer_context key %q contains ':' which breaks block-form round-trip through the formatter", k))
			delete(result, k)
		}
	}
	return result
}

// isSteerContextBlockForm returns true when raw looks like block-form content
// (one or more "k: v" lines) rather than inline "k=v,k=v". Decides by finding
// the first ":" or "=" — whichever comes first wins.
func isSteerContextBlockForm(raw string) bool {
	colon := strings.Index(raw, ":")
	equals := strings.Index(raw, "=")
	if colon == -1 {
		return false // no colon → cannot be block form
	}
	if equals == -1 {
		return true // has colon, no equals → block form
	}
	return colon < equals
}

// parseSteerContextInline parses "k=v, k=v, k=v" into a map. Values containing
// commas are not supported — callers must use the block form for such values.
func (p *Parser) parseSteerContextInline(raw string) map[string]string {
	out := make(map[string]string)
	for _, part := range strings.Split(raw, ",") {
		p.applySteerContextEntry(out, strings.TrimSpace(part))
	}
	return out
}

// applySteerContextEntry parses a single "k=v" token into out, emitting
// a diagnostic for malformed entries or duplicate keys.
func (p *Parser) applySteerContextEntry(out map[string]string, raw string) {
	if raw == "" {
		return
	}
	kv := strings.SplitN(raw, "=", 2)
	if len(kv) != 2 {
		p.diagnostics = append(p.diagnostics,
			fmt.Sprintf("steer_context entry %q must be key=value", raw))
		return
	}
	k := strings.TrimSpace(kv[0])
	v := strings.TrimSpace(kv[1])
	if k == "" {
		return
	}
	if _, exists := out[k]; exists {
		p.diagnostics = append(p.diagnostics,
			fmt.Sprintf("duplicate steer_context key %q (last value wins)", k))
	}
	out[k] = unquoteRaw(v)
}

// parseParamsBlock parses a raw block of key: value lines into a map.
// Duplicate keys emit a diagnostic and last-write-wins.
func (p *Parser) parseParamsBlock(raw string) map[string]string {
	params := make(map[string]string)
	for _, line := range strings.Split(raw, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		p.parseParamLine(params, line)
	}
	return params
}

func (p *Parser) parseParamLine(params map[string]string, line string) {
	k, v := splitKeyValue(line)
	if k == "" {
		return
	}
	if _, exists := params[k]; exists {
		p.diagnostics = append(p.diagnostics,
			fmt.Sprintf("duplicate params key %q (last value wins)", k))
	}
	params[k] = unquoteRaw(v)
}

func (p *Parser) parseParallel() {
	p.lexer.NextToken() // parallel
	id := p.lexer.NextToken().Value

	if p.lexer.PeekToken().Type == TokenArrow {
		p.parseParallelInline(id)
		return
	}
	p.parseParallelBlock(id)
}

// parseParallelInline handles: parallel ID -> target, target
func (p *Parser) parseParallelInline(id string) {
	p.expect(TokenArrow)
	targets := p.parseCommaList()
	p.workflow.Nodes = append(p.workflow.Nodes, &ir.Node{
		ID:     id,
		Kind:   ir.NodeParallel,
		Config: ir.ParallelConfig{Targets: targets},
	})
	p.expect(TokenNewline)
}

// parseParallelBlock handles block form with per-branch config.
func (p *Parser) parseParallelBlock(id string) {
	p.expect(TokenNewline)
	p.expect(TokenIndent)
	branches := p.parseParallelBranches()
	p.expect(TokenOutdent)

	targets := branchTargets(branches)
	p.workflow.Nodes = append(p.workflow.Nodes, &ir.Node{
		ID:   id,
		Kind: ir.NodeParallel,
		Config: ir.ParallelConfig{
			Targets:  targets,
			Branches: branches,
		},
	})
}

// parseParallelBranches parses branch declarations inside a parallel block.
func (p *Parser) parseParallelBranches() []ir.BranchConfig {
	var branches []ir.BranchConfig
	for p.lexer.PeekToken().Type != TokenOutdent && p.lexer.PeekToken().Type != TokenEOF {
		t := p.lexer.PeekToken()
		if b, ok := p.tryParseBranch(t); ok {
			branches = append(branches, b)
		}
	}
	return branches
}

// tryParseBranch tries to parse a branch or skip a non-branch token.
func (p *Parser) tryParseBranch(t Token) (ir.BranchConfig, bool) {
	if t.Type == TokenNewline {
		p.lexer.NextToken()
		return ir.BranchConfig{}, false
	}
	if t.Type == TokenIdentifier && t.Value == "branch" {
		return p.parseOneBranch(), true
	}
	p.lexer.NextToken()
	return ir.BranchConfig{}, false
}

// parseOneBranch parses: branch: target\n  model: ...\n  provider: ...
func (p *Parser) parseOneBranch() ir.BranchConfig {
	p.lexer.NextToken() // "branch"
	p.expect(TokenColon)
	target := p.lexer.NextToken().Value
	bc := ir.BranchConfig{Target: target}
	p.consumeUntilNewline()

	if p.lexer.PeekToken().Type == TokenNewline {
		p.lexer.NextToken()
	}
	if p.lexer.PeekToken().Type != TokenIndent {
		return bc
	}
	p.expect(TokenIndent)
	p.parseBranchFields(&bc)
	p.expect(TokenOutdent)
	return bc
}

// parseBranchFields parses fields within a branch block.
func (p *Parser) parseBranchFields(bc *ir.BranchConfig) {
	for p.lexer.PeekToken().Type != TokenOutdent && p.lexer.PeekToken().Type != TokenEOF {
		t := p.lexer.PeekToken()
		if t.Type == TokenNewline {
			p.lexer.NextToken()
			continue
		}
		if t.Type == TokenIdentifier {
			key := t.Value
			p.lexer.NextToken()
			p.expect(TokenColon)
			val := p.readFieldValue(t.Location.Line)
			applyBranchField(bc, key, val)
		} else {
			p.lexer.NextToken()
		}
	}
}

// applyBranchField sets a field on a BranchConfig.
func applyBranchField(bc *ir.BranchConfig, key, val string) {
	switch key {
	case "model":
		bc.Model = val
	case "provider":
		bc.Provider = val
	case "fidelity":
		bc.Fidelity = val
	}
}

// branchTargets extracts target IDs from branch configs.
func branchTargets(branches []ir.BranchConfig) []string {
	targets := make([]string, len(branches))
	for i, b := range branches {
		targets[i] = b.Target
	}
	return targets
}

func (p *Parser) parseFanIn() {
	p.lexer.NextToken() // fan_in
	id := p.lexer.NextToken().Value
	p.expect(TokenBackArrow)
	sources := p.parseCommaList()
	p.workflow.Nodes = append(p.workflow.Nodes, &ir.Node{
		ID:     id,
		Kind:   ir.NodeFanIn,
		Config: ir.FanInConfig{Sources: sources},
	})
	p.expect(TokenNewline)
}
