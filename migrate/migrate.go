package migrate

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/2389-research/dippin-lang/formatter"
	"github.com/2389-research/dippin-lang/ir"
)

// Migrate parses a DOT digraph string and produces a Dippin IR workflow.
// It applies all post-migration cleanup automatically:
//   - Shape → node kind mapping
//   - Un-escaping of prompts and commands (handled by the DOT lexer)
//   - Namespace prefixing for condition variables (bare "outcome" → "ctx.outcome")
//   - Start/exit identification from Mdiamond/Msquare shapes
//   - Graph-level attribute extraction to WorkflowDefaults
func Migrate(dotSource string) (*ir.Workflow, error) {
	dg, err := parseDOT(dotSource)
	if err != nil {
		return nil, fmt.Errorf("migrate: %w", err)
	}
	return convertDOTGraph(dg)
}

// MigrateToSource parses DOT and returns canonical .dip source text.
// Convenience: equivalent to formatter.Format(Migrate(dotSource)).
func MigrateToSource(dotSource string) (string, error) {
	w, err := Migrate(dotSource)
	if err != nil {
		return "", err
	}
	return formatter.Format(w), nil
}

// --- Shape → Kind Mapping ---

// shapeToKind maps DOT shape attributes to IR node kinds.
// Mdiamond and Msquare are handled specially (start/exit markers).
// diamond is handled with attribute-based disambiguation in resolveDiamondKind.
var shapeToKind = map[string]ir.NodeKind{
	"box":           ir.NodeAgent,
	"hexagon":       ir.NodeHuman,
	"parallelogram": ir.NodeTool,
	"component":     ir.NodeParallel,
	"tripleoctagon": ir.NodeFanIn,
	"tab":           ir.NodeSubgraph,
	"house":         ir.NodeManagerLoop,
}

// convertDOTGraph transforms a parsed DOT graph into an IR workflow.
func convertDOTGraph(dg *dotGraph) (*ir.Workflow, error) {
	w := &ir.Workflow{
		Name:    dg.Name,
		Version: "1",
	}

	extractGraphDefaults(dg.GraphAttrs, w)
	w.Start, w.Exit = identifyStartExit(dg.Nodes)

	if err := buildIRNodes(dg, w); err != nil {
		return nil, err
	}
	if err := buildIREdges(dg, w); err != nil {
		return nil, err
	}

	inferParallelFanIn(w)
	return w, nil
}

// identifyStartExit finds start and exit node IDs from DOT shapes.
func identifyStartExit(nodes []dotNode) (start, exit string) {
	for _, dn := range nodes {
		shape := dn.Attrs["shape"]
		if shape == "Mdiamond" {
			start = dn.ID
		}
		if shape == "Msquare" {
			exit = dn.ID
		}
	}
	return start, exit
}

// buildIRNodes converts all DOT nodes to IR nodes on the workflow.
func buildIRNodes(dg *dotGraph, w *ir.Workflow) error {
	for _, dn := range dg.Nodes {
		node, err := convertNode(dn, dg.Edges)
		if err != nil {
			return fmt.Errorf("migrate: node %q: %w", dn.ID, err)
		}
		w.Nodes = append(w.Nodes, node)
	}
	return nil
}

// buildIREdges converts all DOT edges to IR edges on the workflow.
func buildIREdges(dg *dotGraph, w *ir.Workflow) error {
	for _, de := range dg.Edges {
		edge, err := convertEdge(de)
		if err != nil {
			return fmt.Errorf("migrate: edge %s->%s: %w", de.From, de.To, err)
		}
		w.Edges = append(w.Edges, edge)
	}
	return nil
}

// graphDefaultsHandlers maps DOT graph attribute keys to handler functions.
var graphDefaultsHandlers = map[string]func(string, *ir.Workflow){
	"goal":                func(v string, w *ir.Workflow) { w.Goal = v },
	"rankdir":             func(_ string, _ *ir.Workflow) { /* presentation-only; ignored */ },
	"default_fidelity":    func(v string, w *ir.Workflow) { w.Defaults.Fidelity = v },
	"fidelity":            func(v string, w *ir.Workflow) { w.Defaults.Fidelity = v },
	"model":               func(v string, w *ir.Workflow) { w.Defaults.Model = v },
	"provider":            func(v string, w *ir.Workflow) { w.Defaults.Provider = v },
	"tool_commands_allow": func(v string, w *ir.Workflow) { w.Defaults.ToolCommandsAllow = v },
	"tool_denylist_add":   func(v string, w *ir.Workflow) { w.Defaults.ToolDenylistAdd = v },
}

// extractGraphDefaults populates workflow-level fields from DOT graph attributes.
func extractGraphDefaults(attrs map[string]string, w *ir.Workflow) {
	for k, v := range attrs {
		if handler, ok := graphDefaultsHandlers[k]; ok {
			handler(v, w)
			continue
		}
		applyUnknownGraphAttr(k, v, w)
	}
}

// applyUnknownGraphAttr routes unknown graph attributes: only recognized
// integer-backed default keys (max_retries, max_restarts) go to Defaults;
// all other attrs, including unrecognized integer-valued ones, are captured
// in Workflow.Vars. Keys that aren't valid Dippin identifiers are skipped.
func applyUnknownGraphAttr(k, v string, w *ir.Workflow) {
	if applyIntDefault(k, v, w) {
		return
	}
	if !isDippinIdentifier(k) {
		return
	}
	if w.Vars == nil {
		w.Vars = make(map[string]string)
	}
	w.Vars[k] = v
}

// dippinIdentPattern matches valid Dippin identifiers:
// alphanumeric start, followed by alphanumeric, underscore, dash, dot, or slash.
var dippinIdentPattern = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_\-./]*$`)

// isDippinIdentifier returns true if s is a valid Dippin identifier.
func isDippinIdentifier(s string) bool {
	return dippinIdentPattern.MatchString(s)
}

// applyIntDefault handles integer-valued and duration-valued graph defaults.
// Returns true if the key was recognized and applied.
func applyIntDefault(k, v string, w *ir.Workflow) bool {
	if applyIntBudgetDefault(k, v, w) {
		return true
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return false
	}
	switch k {
	case "default_max_retry", "max_retries":
		w.Defaults.MaxRetries = n
	case "max_restarts":
		w.Defaults.MaxRestarts = n
	default:
		return false
	}
	return true
}

// applyIntBudgetDefault handles budget-related integer and duration defaults.
func applyIntBudgetDefault(k, v string, w *ir.Workflow) bool {
	switch k {
	case "max_total_tokens":
		return tryApplyIntDefault(v, &w.Defaults.MaxTotalTokens)
	case "max_cost_cents":
		return tryApplyIntDefault(v, &w.Defaults.MaxCostCents)
	case "max_wall_time":
		return tryApplyDurationDefault(v, &w.Defaults.MaxWallTime)
	default:
		return false
	}
}

// tryApplyIntDefault parses an integer and sets it on target. Returns true on success.
func tryApplyIntDefault(v string, target *int) bool {
	n, err := strconv.Atoi(v)
	if err != nil {
		return false
	}
	*target = n
	return true
}

// tryApplyDurationDefault parses a duration and sets it on target. Returns true on success.
func tryApplyDurationDefault(v string, target *time.Duration) bool {
	d, err := time.ParseDuration(v)
	if err != nil {
		return false
	}
	*target = d
	return true
}

// convertNode converts a DOT node to an IR node.
func convertNode(dn dotNode, edges []dotEdge) (*ir.Node, error) {
	shape := dn.Attrs["shape"]
	kind := resolveKind(shape, dn.Attrs)

	node := &ir.Node{
		ID:   dn.ID,
		Kind: kind,
	}

	if label, ok := dn.Attrs["label"]; ok {
		node.Label = label
	}

	return applyNodeConfig(node, kind, dn.Attrs)
}

// applyNodeConfig sets kind-specific config on the node.
func applyNodeConfig(node *ir.Node, kind ir.NodeKind, attrs map[string]string) (*ir.Node, error) {
	cfg, err := buildNodeConfig(node, kind, attrs)
	if err != nil {
		return nil, err
	}
	node.Config = cfg
	return node, nil
}

// buildNodeConfig returns the config for a node given its kind and DOT attributes.
func buildNodeConfig(node *ir.Node, kind ir.NodeKind, attrs map[string]string) (ir.NodeConfig, error) {
	switch kind {
	case ir.NodeAgent:
		node.Retry = buildRetryConfig(attrs)
		return buildAgentConfig(attrs), nil
	case ir.NodeHuman:
		return buildHumanConfig(attrs)
	case ir.NodeTool:
		return buildToolConfig(attrs)
	default:
		return buildOtherConfig(kind, attrs), nil
	}
}

// buildOtherConfig builds config for parallel, fan_in, subgraph, conditional, manager_loop, or fallback.
func buildOtherConfig(kind ir.NodeKind, attrs map[string]string) ir.NodeConfig {
	switch kind {
	case ir.NodeParallel:
		return buildParallelConfig(attrs)
	case ir.NodeFanIn:
		return buildFanInConfig(attrs)
	case ir.NodeSubgraph:
		return buildSubgraphConfig(attrs)
	default:
		return buildFlowControlConfig(kind, attrs)
	}
}

// buildFlowControlConfig handles conditional, manager_loop, and the agent fallback.
func buildFlowControlConfig(kind ir.NodeKind, attrs map[string]string) ir.NodeConfig {
	switch kind {
	case ir.NodeConditional:
		return ir.ConditionalConfig{}
	case ir.NodeManagerLoop:
		return buildManagerLoopConfig(attrs)
	default:
		return ir.AgentConfig{}
	}
}

// resolveKind determines the IR node kind from the DOT shape and attributes.
func resolveKind(shape string, attrs map[string]string) ir.NodeKind {
	if shape == "Mdiamond" || shape == "Msquare" {
		return resolveStartExitKind(attrs)
	}
	if shape == "diamond" {
		return resolveDiamondKind(attrs)
	}
	if kind, ok := shapeToKind[shape]; ok {
		return kind
	}
	return ir.NodeAgent
}

// hasManagerLoopAttrs reports whether attrs contain any manager_loop-specific
// key. Used to detect manager_loop nodes when their kind-based shape was
// overridden to Mdiamond/Msquare by start/exit marker export.
func hasManagerLoopAttrs(attrs map[string]string) bool {
	managerLoopKeys := []string{
		"subgraph_ref",
		"poll_interval",
		"max_cycles",
		"stop_condition",
		"steer_condition",
		"steer_context",
	}
	for _, key := range managerLoopKeys {
		if _, ok := attrs[key]; ok {
			return true
		}
	}
	return false
}

// resolveStartExitKind disambiguates start/exit-marker shapes (Mdiamond, Msquare)
// by checking for any manager_loop-specific attribute. Start/exit override the
// kind-based shape in export, so migrate has to recover the original kind from
// attrs. Checking any manager_loop key (not just subgraph_ref) ensures partial
// configurations (e.g., poll_interval set but subgraph_ref missing) are not
// silently downgraded to NodeAgent, losing all configured fields.
func resolveStartExitKind(attrs map[string]string) ir.NodeKind {
	if hasManagerLoopAttrs(attrs) {
		return ir.NodeManagerLoop
	}
	return ir.NodeAgent
}

// resolveDiamondKind handles diamond shape disambiguation.
// diamond + tool_command → NodeTool, diamond + prompt → NodeAgent,
// bare diamond → NodeConditional (pure routing).
func resolveDiamondKind(attrs map[string]string) ir.NodeKind {
	if _, hasTool := attrs["tool_command"]; hasTool {
		return ir.NodeTool
	}
	if _, hasPrompt := attrs["prompt"]; hasPrompt {
		return ir.NodeAgent
	}
	return ir.NodeConditional
}

// --- Config builders ---

func buildAgentConfig(attrs map[string]string) ir.AgentConfig {
	cfg := ir.AgentConfig{}
	applyPromptAttrs(&cfg, attrs)
	applyModelAttrs(&cfg, attrs)
	applyBehaviorAttrs(&cfg, attrs)
	applyPerformanceAttrs(&cfg, attrs)
	return cfg
}

// applyPromptAttrs applies prompt-related attributes to agent config.
func applyPromptAttrs(cfg *ir.AgentConfig, attrs map[string]string) {
	if v, ok := attrs["prompt"]; ok {
		cfg.Prompt = v
	}
	if v, ok := attrs["system_prompt"]; ok {
		cfg.SystemPrompt = v
	}
}

// applyModelAttrs applies model and provider attributes to agent config.
func applyModelAttrs(cfg *ir.AgentConfig, attrs map[string]string) {
	applyModelField(cfg, attrs)
	applyProviderField(cfg, attrs)
	if v, ok := attrs["reasoning_effort"]; ok {
		cfg.ReasoningEffort = v
	}
	if v, ok := attrs["fidelity"]; ok {
		cfg.Fidelity = v
	}
	if v, ok := attrs["backend"]; ok {
		cfg.Backend = v
	}
	if v, ok := attrs["working_dir"]; ok {
		cfg.WorkingDir = v
	}
}

// applyModelField sets the model field from model or llm_model attrs.
func applyModelField(cfg *ir.AgentConfig, attrs map[string]string) {
	if v, ok := attrs["model"]; ok {
		cfg.Model = v
	}
	if v, ok := attrs["llm_model"]; ok {
		cfg.Model = v
	}
}

// applyProviderField sets the provider field from provider or llm_provider attrs.
func applyProviderField(cfg *ir.AgentConfig, attrs map[string]string) {
	if v, ok := attrs["provider"]; ok {
		cfg.Provider = v
	}
	if v, ok := attrs["llm_provider"]; ok {
		cfg.Provider = v
	}
}

// applyBehaviorAttrs applies behavior-related attributes to agent config.
func applyBehaviorAttrs(cfg *ir.AgentConfig, attrs map[string]string) {
	if v, ok := attrs["goal_gate"]; ok && isTruthy(v) {
		cfg.GoalGate = true
	}
	if v, ok := attrs["auto_status"]; ok && isTruthy(v) {
		cfg.AutoStatus = true
	}
	applyMaxTurns(cfg, attrs)
}

// applyMaxTurns parses and sets the max_turns field.
func applyMaxTurns(cfg *ir.AgentConfig, attrs map[string]string) {
	if v, ok := attrs["max_turns"]; ok {
		if n, err := strconv.Atoi(v); err == nil {
			cfg.MaxTurns = n
		}
	}
}

// applyPerformanceAttrs applies performance-related attributes to agent config.
func applyPerformanceAttrs(cfg *ir.AgentConfig, attrs map[string]string) {
	applyCmdTimeout(cfg, attrs)
	if v, ok := attrs["cache_tools"]; ok && isTruthy(v) {
		cfg.CacheTools = true
	}
	if v, ok := attrs["compaction"]; ok {
		cfg.Compaction = v
	}
}

// applyCmdTimeout parses and sets the cmd_timeout duration.
func applyCmdTimeout(cfg *ir.AgentConfig, attrs map[string]string) {
	if v, ok := attrs["cmd_timeout"]; ok {
		if d, err := time.ParseDuration(v); err == nil {
			cfg.CmdTimeout = d
		}
	}
}

func buildHumanConfig(attrs map[string]string) (ir.HumanConfig, error) {
	cfg := ir.HumanConfig{}
	if v, ok := attrs["mode"]; ok {
		cfg.Mode = v
	}
	if v, ok := attrs["default"]; ok {
		cfg.Default = v
	}
	if err := applyHumanTimeout(&cfg, attrs); err != nil {
		return cfg, err
	}
	return cfg, nil
}

// applyHumanTimeout parses and sets timeout fields on HumanConfig.
// Returns an error if timeout is unparseable or timeout_action is invalid.
func applyHumanTimeout(cfg *ir.HumanConfig, attrs map[string]string) error {
	if err := applyHumanTimeoutDuration(cfg, attrs); err != nil {
		return err
	}
	return applyHumanTimeoutAction(cfg, attrs)
}

// applyHumanTimeoutDuration parses and sets the timeout duration field.
func applyHumanTimeoutDuration(cfg *ir.HumanConfig, attrs map[string]string) error {
	v, ok := attrs["timeout"]
	if !ok {
		return nil
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return fmt.Errorf("invalid timeout %q: %w", v, err)
	}
	cfg.Timeout = d
	return nil
}

// applyHumanTimeoutAction validates and sets the timeout_action field.
func applyHumanTimeoutAction(cfg *ir.HumanConfig, attrs map[string]string) error {
	v, ok := attrs["timeout_action"]
	if !ok {
		return nil
	}
	switch v {
	case "", "fail", "default":
		cfg.TimeoutAction = v
		return nil
	default:
		return fmt.Errorf("invalid timeout_action %q (use fail, default, or empty)", v)
	}
}

func buildToolConfig(attrs map[string]string) (ir.ToolConfig, error) {
	cfg := ir.ToolConfig{}
	applyToolStringAttrs(&cfg, attrs)
	if err := applyToolTimeoutAttr(&cfg, attrs); err != nil {
		return cfg, err
	}
	if err := applyToolOutputLimitAttr(&cfg, attrs); err != nil {
		return cfg, err
	}
	return cfg, nil
}

func applyToolStringAttrs(cfg *ir.ToolConfig, attrs map[string]string) {
	if v, ok := attrs["tool_command"]; ok {
		cfg.Command = v
	}
	if v, ok := attrs["marker_grep"]; ok {
		cfg.MarkerGrep = v
	}
	if v, ok := attrs["route_required"]; ok {
		cfg.RouteRequired = (v == "true")
	}
	if v, ok := attrs["outputs"]; ok {
		cfg.Outputs = splitComma(v)
	}
}

func applyToolTimeoutAttr(cfg *ir.ToolConfig, attrs map[string]string) error {
	v, ok := attrs["timeout"]
	if !ok {
		return nil
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return fmt.Errorf("invalid timeout %q: %w", v, err)
	}
	cfg.Timeout = d
	return nil
}

func applyToolOutputLimitAttr(cfg *ir.ToolConfig, attrs map[string]string) error {
	v, ok := attrs["output_limit"]
	if !ok {
		return nil
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fmt.Errorf("invalid output_limit %q: %w", v, err)
	}
	if n < 0 {
		return fmt.Errorf("invalid output_limit %d: must be non-negative", n)
	}
	cfg.OutputLimit = n
	return nil
}

func buildParallelConfig(attrs map[string]string) ir.ParallelConfig {
	cfg := ir.ParallelConfig{}
	if v, ok := attrs["targets"]; ok {
		cfg.Targets = splitComma(v)
	}
	return cfg
}

func buildFanInConfig(attrs map[string]string) ir.FanInConfig {
	cfg := ir.FanInConfig{}
	if v, ok := attrs["sources"]; ok {
		cfg.Sources = splitComma(v)
	}
	return cfg
}

func buildSubgraphConfig(attrs map[string]string) ir.SubgraphConfig {
	cfg := ir.SubgraphConfig{}
	if v, ok := attrs["ref"]; ok {
		cfg.Ref = v
	}
	return cfg
}

// buildManagerLoopConfig constructs a ManagerLoopConfig from DOT node attrs.
// Mirrors the flat-attr format produced by export.applyManagerLoopAttrs.
func buildManagerLoopConfig(attrs map[string]string) ir.ManagerLoopConfig {
	cfg := ir.ManagerLoopConfig{SteerContext: map[string]string{}}
	applyManagerLoopScalarMigrate(&cfg, attrs)
	applyManagerLoopConditionMigrate(&cfg, attrs)
	return cfg
}

// applyManagerLoopScalarMigrate populates subgraph_ref, poll_interval,
// max_cycles, and steer_context (flattened form) from DOT attrs.
func applyManagerLoopScalarMigrate(cfg *ir.ManagerLoopConfig, attrs map[string]string) {
	if v, ok := attrs["subgraph_ref"]; ok {
		cfg.SubgraphRef = v
	}
	applyScalarDuration(&cfg.PollInterval, attrs, "poll_interval")
	applyScalarInt(&cfg.MaxCycles, attrs, "max_cycles")
	if v, ok := attrs["steer_context"]; ok && v != "" {
		cfg.SteerContext = parseFlattenedSteerContext(v)
	}
}

// applyScalarDuration reads attrs[key] and parses into *target if present and valid.
// Thin wrapper over tryApplyDurationDefault that also handles the attr-lookup.
func applyScalarDuration(target *time.Duration, attrs map[string]string, key string) {
	if v, ok := attrs[key]; ok {
		_ = tryApplyDurationDefault(v, target)
	}
}

// applyScalarInt reads attrs[key] and parses into *target if present and valid.
// Thin wrapper over tryApplyIntDefault that also handles the attr-lookup.
func applyScalarInt(target *int, attrs map[string]string, key string) {
	if v, ok := attrs[key]; ok {
		_ = tryApplyIntDefault(v, target)
	}
}

// applyManagerLoopConditionMigrate populates StopCondition and SteerCondition
// as *ir.Condition with Raw set; Parsed remains nil until simulate.EnsureConditionsParsed runs.
func applyManagerLoopConditionMigrate(cfg *ir.ManagerLoopConfig, attrs map[string]string) {
	if v, ok := attrs["stop_condition"]; ok && v != "" {
		cfg.StopCondition = &ir.Condition{Raw: v}
	}
	if v, ok := attrs["steer_condition"]; ok && v != "" {
		cfg.SteerCondition = &ir.Condition{Raw: v}
	}
}

// steerContextDecoder reverses steerContextEncoder from export/dot.go.
// Sequences are listed longest-first so the replacer matches greedily.
var steerContextDecoder = strings.NewReplacer(
	"%25", "%",
	"%2C", ",",
	"%3D", "=",
)

// decodeSteerContextToken reverses encodeSteerContextToken from export.
// Returns the input unchanged when it contains no '%' escapes.
func decodeSteerContextToken(s string) string {
	if !strings.Contains(s, "%") {
		return s
	}
	return steerContextDecoder.Replace(s)
}

// parseFlattenedSteerContext splits "k=v,k=v" (canonical form from DOT export)
// into a map. Malformed entries are silently skipped — they cannot occur from
// our own exporter and strict rejection would break migrations of manually
// edited DOT files. Percent-encoded tokens (produced by encodeSteerContextToken
// in export) are decoded back to their original values.
func parseFlattenedSteerContext(s string) map[string]string {
	out := map[string]string{}
	for _, part := range strings.Split(s, ",") {
		kv := strings.SplitN(strings.TrimSpace(part), "=", 2)
		if len(kv) != 2 {
			continue
		}
		k := decodeSteerContextToken(strings.TrimSpace(kv[0]))
		v := decodeSteerContextToken(strings.TrimSpace(kv[1]))
		out[k] = v
	}
	return out
}

func buildRetryConfig(attrs map[string]string) ir.RetryConfig {
	rc := ir.RetryConfig{}
	if v, ok := attrs["retry_policy"]; ok {
		rc.Policy = v
	}
	applyRetryMaxRetries(&rc, attrs)
	if v, ok := attrs["retry_target"]; ok {
		rc.RetryTarget = v
	}
	if v, ok := attrs["fallback_target"]; ok {
		rc.FallbackTarget = v
	}
	return rc
}

// applyRetryMaxRetries parses and sets the max_retries field on RetryConfig.
func applyRetryMaxRetries(rc *ir.RetryConfig, attrs map[string]string) {
	if v, ok := attrs["max_retries"]; ok {
		if n, err := strconv.Atoi(v); err == nil {
			rc.MaxRetries = n
		}
	}
}

// --- Edge conversion ---

func convertEdge(de dotEdge) (*ir.Edge, error) {
	e := &ir.Edge{
		From: de.From,
		To:   de.To,
	}

	if v, ok := de.Attrs["label"]; ok {
		e.Label = v
	}
	if err := applyEdgeCondition(e, de.Attrs); err != nil {
		return nil, err
	}
	applyEdgeWeight(e, de.Attrs)
	applyEdgeRestart(e, de.Attrs)
	return e, nil
}

// applyEdgeCondition parses and sets the condition on an edge.
func applyEdgeCondition(e *ir.Edge, attrs map[string]string) error {
	v, ok := attrs["condition"]
	if !ok {
		return nil
	}
	cond, err := parseCondition(v)
	if err != nil {
		return fmt.Errorf("condition %q: %w", v, err)
	}
	e.Condition = cond
	return nil
}

// applyEdgeWeight parses and sets the weight on an edge.
func applyEdgeWeight(e *ir.Edge, attrs map[string]string) {
	if v, ok := attrs["weight"]; ok {
		if n, err := strconv.Atoi(v); err == nil {
			e.Weight = n
		}
	}
}

// applyEdgeRestart checks for restart/loop_restart attrs.
func applyEdgeRestart(e *ir.Edge, attrs map[string]string) {
	if v, ok := attrs["restart"]; ok && isTruthy(v) {
		e.Restart = true
	}
	if v, ok := attrs["loop_restart"]; ok && isTruthy(v) {
		e.Restart = true
	}
}

// --- Condition parsing ---

// parseCondition parses a DOT condition string into an ir.Condition.
// It handles:
//   - Simple comparisons: outcome=success, tool_stdout contains pass
//   - AND/OR: outcome=success && tool_stdout contains done
//   - NOT: not outcome=fail, !outcome=fail
//   - Namespace prefixing: bare "outcome" → "ctx.outcome"
func parseCondition(raw string) (*ir.Condition, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}

	expr, err := parseCondExpr(raw)
	if err != nil {
		return nil, err
	}

	canonRaw := formatCondExpr(expr)
	return &ir.Condition{
		Raw:    canonRaw,
		Parsed: expr,
	}, nil
}

// condExprParser is a function that attempts to parse a condition expression.
// Returns (nil, nil) if this parser does not apply.
type condExprParser func(string) (ir.ConditionExpr, error)

// parseCondExpr parses a condition expression string into an AST.
func parseCondExpr(s string) (ir.ConditionExpr, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, fmt.Errorf("empty condition expression")
	}
	return tryCondParsers(s)
}

// tryCondParsers tries each condition parser in precedence order.
func tryCondParsers(s string) (ir.ConditionExpr, error) {
	parsers := []condExprParser{tryParseOr, tryParseAnd, tryParseNot}
	for _, p := range parsers {
		if expr, err := p(s); expr != nil || err != nil {
			return expr, err
		}
	}
	return parseComparison(s)
}

// tryParseOr attempts to parse an OR expression.
func tryParseOr(s string) (ir.ConditionExpr, error) {
	parts, ok := splitLogicalOp(s, "||")
	if !ok {
		return nil, nil
	}
	left, err := parseCondExpr(parts[0])
	if err != nil {
		return nil, err
	}
	right, err := parseCondExpr(parts[1])
	if err != nil {
		return nil, err
	}
	return ir.CondOr{Left: left, Right: right}, nil
}

// tryParseAnd attempts to parse an AND expression.
func tryParseAnd(s string) (ir.ConditionExpr, error) {
	parts, ok := splitLogicalOp(s, "&&")
	if !ok {
		return nil, nil
	}
	left, err := parseCondExpr(parts[0])
	if err != nil {
		return nil, err
	}
	right, err := parseCondExpr(parts[1])
	if err != nil {
		return nil, err
	}
	return ir.CondAnd{Left: left, Right: right}, nil
}

// tryParseNot attempts to parse a NOT expression.
func tryParseNot(s string) (ir.ConditionExpr, error) {
	rest := ""
	if strings.HasPrefix(s, "not ") {
		rest = s[4:]
	} else if strings.HasPrefix(s, "!") {
		rest = s[1:]
	}
	if rest == "" {
		return nil, nil
	}
	inner, err := parseCondExpr(rest)
	if err != nil {
		return nil, err
	}
	return ir.CondNot{Inner: inner}, nil
}

// splitLogicalOp splits a condition string on the given logical operator (&&, ||).
// Returns the two parts if the operator is found at the top level (not inside parens).
func splitLogicalOp(s, op string) ([]string, bool) {
	depth := 0
	for i := 0; i <= len(s)-len(op); i++ {
		depth = updateParenDepth(s[i], depth)
		if depth == 0 && s[i:i+len(op)] == op {
			return trySplitAt(s, i, len(op))
		}
	}
	return nil, false
}

// updateParenDepth adjusts the parenthesis nesting depth.
func updateParenDepth(ch byte, depth int) int {
	if ch == '(' {
		return depth + 1
	}
	if ch == ')' {
		return depth - 1
	}
	return depth
}

// trySplitAt splits a string at position i with operator length opLen.
func trySplitAt(s string, i, opLen int) ([]string, bool) {
	left := strings.TrimSpace(s[:i])
	right := strings.TrimSpace(s[i+opLen:])
	if left != "" && right != "" {
		return []string{left, right}, true
	}
	return nil, false
}

// parseComparison parses a single condition comparison like "outcome=success"
// or "tool_stdout contains pass".
func parseComparison(s string) (ir.ConditionExpr, error) {
	s = strings.TrimSpace(s)

	// Try != first (before =) to avoid matching the = in !=.
	if idx := strings.Index(s, "!="); idx > 0 {
		return buildCompare(s, idx, 2, "!="), nil
	}

	// Try = (equality).
	if idx := strings.Index(s, "="); idx > 0 {
		return buildCompare(s, idx, 1, "="), nil
	}

	// Try word-based operators: contains, startswith, endswith, in.
	for _, op := range []string{" contains ", " startswith ", " endswith ", " in "} {
		if idx := strings.Index(s, op); idx > 0 {
			return buildCompare(s, idx, len(op), strings.TrimSpace(op)), nil
		}
	}

	return nil, fmt.Errorf("cannot parse condition comparison: %q", s)
}

// buildCompare creates a CondCompare from a string, split position, and operator.
func buildCompare(s string, idx, opLen int, op string) ir.CondCompare {
	variable := strings.TrimSpace(s[:idx])
	value := strings.TrimSpace(s[idx+opLen:])
	return ir.CondCompare{
		Variable: addNamespacePrefix(variable),
		Op:       op,
		Value:    value,
	}
}

// addNamespacePrefix adds the ctx. namespace to bare condition variable names.
// Variables that already contain a dot are left as-is (graph.*, ctx.*).
// The legacy "context." prefix is normalized to "ctx.".
func addNamespacePrefix(variable string) string {
	variable = strings.TrimSpace(variable)

	// Already namespaced with "context." → normalize to "ctx."
	if strings.HasPrefix(variable, "context.") {
		return "ctx." + variable[len("context."):]
	}

	// Already namespaced (contains a dot).
	if strings.Contains(variable, ".") {
		return variable
	}

	// Bare variable name → add ctx. prefix.
	return "ctx." + variable
}

// formatCondExpr renders a condition AST back to a canonical string.
func formatCondExpr(expr ir.ConditionExpr) string {
	return formatCondExprPrec(expr, 0)
}

const (
	condPrecOr  = 1
	condPrecAnd = 2
	condPrecNot = 3
)

func formatCondExprPrec(expr ir.ConditionExpr, parentPrec int) string {
	switch e := expr.(type) {
	case ir.CondCompare:
		return fmt.Sprintf("%s %s %s", e.Variable, e.Op, e.Value)
	case ir.CondAnd:
		return formatBinaryOp(e.Left, e.Right, "and", condPrecAnd, parentPrec)
	case ir.CondOr:
		return formatBinaryOp(e.Left, e.Right, "or", condPrecOr, parentPrec)
	case ir.CondNot:
		return "not " + formatCondExprPrec(e.Inner, condPrecNot)
	default:
		return ""
	}
}

// formatBinaryOp formats a binary AND/OR expression with optional parens.
func formatBinaryOp(left, right ir.ConditionExpr, keyword string, prec, parentPrec int) string {
	s := fmt.Sprintf("%s %s %s",
		formatCondExprPrec(left, prec),
		keyword,
		formatCondExprPrec(right, prec))
	if parentPrec != 0 && parentPrec != prec {
		return "(" + s + ")"
	}
	return s
}

// --- Parallel/Fan-in inference ---

// inferParallelFanIn fills in Targets and Sources from edges when not
// explicitly set in DOT attributes.
func inferParallelFanIn(w *ir.Workflow) {
	for _, n := range w.Nodes {
		switch cfg := n.Config.(type) {
		case ir.ParallelConfig:
			inferParallelTargets(w, n, cfg)
		case ir.FanInConfig:
			inferFanInSources(w, n, cfg)
		}
	}
}

// inferParallelTargets sets targets from outgoing edges if not already set.
func inferParallelTargets(w *ir.Workflow, n *ir.Node, cfg ir.ParallelConfig) {
	if len(cfg.Targets) != 0 {
		return
	}
	edges := w.EdgesFrom(n.ID)
	targets := make([]string, 0, len(edges))
	for _, e := range edges {
		targets = append(targets, e.To)
	}
	n.Config = ir.ParallelConfig{Targets: targets}
}

// inferFanInSources sets sources from incoming edges if not already set.
func inferFanInSources(w *ir.Workflow, n *ir.Node, cfg ir.FanInConfig) {
	if len(cfg.Sources) != 0 {
		return
	}
	edges := w.EdgesTo(n.ID)
	sources := make([]string, 0, len(edges))
	for _, e := range edges {
		sources = append(sources, e.From)
	}
	n.Config = ir.FanInConfig{Sources: sources}
}

// --- Helpers ---

func isTruthy(s string) bool {
	s = strings.ToLower(strings.TrimSpace(s))
	return s == "true" || s == "1" || s == "yes"
}

func splitComma(s string) []string {
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}
