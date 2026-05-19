// Package export provides DOT graph format export for Dippin workflows.
//
// The primary function ExportDOT converts an ir.Workflow into a valid DOT
// language string suitable for rendering with Graphviz. The mapping is
// documented in §15 of the Dippin design spec.
package export

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/2389-research/dippin-lang/ir"
)

// ExportOptions controls the DOT output format.
type ExportOptions struct {
	// IncludePrompts includes full prompt/command text as node attributes.
	// Default (zero value) is false; set to true for full-fidelity export.
	IncludePrompts bool

	// RankDir controls the graph layout direction: "LR" (left-to-right)
	// or "TB" (top-to-bottom). Defaults to "TB" when empty.
	RankDir string

	// HighlightGoalGates applies a distinct fill color to nodes with
	// GoalGate: true.
	HighlightGoalGates bool

	// ExecutionPath is an ordered list of node IDs visited during execution.
	// If set, nodes on the path are highlighted and numbered.
	ExecutionPath []string
}

// ExportDOT renders a workflow as a DOT language string.
// The output is a valid digraph that can be processed by Graphviz tools.
func ExportDOT(w *ir.Workflow, opts ExportOptions) string {
	var b strings.Builder

	writeDOTHeader(&b, w, opts)

	execOrder := buildExecOrder(opts.ExecutionPath)

	for _, n := range w.Nodes {
		writeNodeDOT(&b, n, w, opts, execOrder)
	}

	b.WriteByte('\n')

	for _, e := range w.Edges {
		writeEdgeDOT(&b, e)
	}

	b.WriteString("}\n")
	return b.String()
}

// reservedGraphAttrs are DOT graph attributes used by the defaults/header — skip
// them when re-emitting workflow vars so they don't collide.
var reservedGraphAttrs = map[string]bool{
	"goal": true, "rankdir": true, "model": true, "provider": true,
	"fidelity": true, "default_fidelity": true,
	"max_retries": true, "default_max_retry": true, "max_restarts": true,
	"max_total_tokens": true, "max_cost_cents": true, "max_wall_time": true,
	"tool_commands_allow": true, "tool_denylist_add": true,
}

// writeDOTHeader writes the digraph opening and global attributes.
func writeDOTHeader(b *strings.Builder, w *ir.Workflow, opts ExportOptions) {
	rankDir := opts.RankDir
	if rankDir == "" {
		rankDir = "TB"
	}
	graphName := w.Name
	if graphName == "" {
		graphName = "workflow"
	}
	fmt.Fprintf(b, "digraph %s {\n", dotID(graphName))
	b.WriteString("  node [fontname=\"Helvetica\"];\n")
	b.WriteString("  edge [fontname=\"Helvetica\"];\n")

	graphAttrs := buildGraphAttrs(w, rankDir)
	writeGraphAttrs(b, graphAttrs)
}

// buildGraphAttrs collects all graph attributes (rankdir, defaults, and vars).
func buildGraphAttrs(w *ir.Workflow, rankDir string) []string {
	attrs := []string{}

	// Add rankdir first for deterministic ordering
	attrs = append(attrs, fmt.Sprintf("rankdir=%s", rankDir))

	// Add tool-safety attributes
	if w.Defaults.ToolCommandsAllow != "" {
		attrs = append(attrs, fmt.Sprintf("tool_commands_allow=%s", dotQuote(w.Defaults.ToolCommandsAllow)))
	}
	if w.Defaults.ToolDenylistAdd != "" {
		attrs = append(attrs, fmt.Sprintf("tool_denylist_add=%s", dotQuote(w.Defaults.ToolDenylistAdd)))
	}

	// Add workflow vars (excluding reserved graph attributes)
	addGraphVarsAttrs(&attrs, w.Vars)

	return attrs
}

// addGraphVarsAttrs adds workflow vars as graph attributes, excluding reserved keys.
func addGraphVarsAttrs(attrs *[]string, vars map[string]string) {
	if len(vars) == 0 {
		return
	}
	keys := make([]string, 0, len(vars))
	for k := range vars {
		if !reservedGraphAttrs[k] {
			keys = append(keys, k)
		}
	}
	sortStrings(keys)
	for _, k := range keys {
		*attrs = append(*attrs, fmt.Sprintf("%s=%s", dotID(k), dotQuote(vars[k])))
	}
}

// writeGraphAttrs emits the graph attribute block.
func writeGraphAttrs(b *strings.Builder, attrs []string) {
	b.WriteString("  graph [")
	for i, attr := range attrs {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(attr)
	}
	b.WriteString("];\n")
}

// buildExecOrder builds a map from node ID to execution order indices.
func buildExecOrder(path []string) map[string][]int {
	execOrder := make(map[string][]int)
	for i, id := range path {
		execOrder[id] = append(execOrder[id], i+1)
	}
	return execOrder
}

// nodeShapes maps NodeKind to the corresponding DOT shape attribute.
// Per §15: agent→box, human→hexagon, tool→parallelogram,
// parallel→component, fan_in→tripleoctagon, subgraph→tab, conditional→diamond,
// manager_loop→house.
var nodeShapes = map[ir.NodeKind]string{
	ir.NodeAgent:       "box",
	ir.NodeHuman:       "hexagon",
	ir.NodeTool:        "parallelogram",
	ir.NodeParallel:    "component",
	ir.NodeFanIn:       "tripleoctagon",
	ir.NodeSubgraph:    "tab",
	ir.NodeConditional: "diamond",
	ir.NodeManagerLoop: "house",
}

// nodeShape returns the DOT shape for a given NodeKind.
// Start and exit nodes override the kind-based shape elsewhere.
func nodeShape(kind ir.NodeKind) string {
	if s, ok := nodeShapes[kind]; ok {
		return s
	}
	return "box"
}

// writeNodeDOT emits a single DOT node statement.
func writeNodeDOT(b *strings.Builder, n *ir.Node, w *ir.Workflow, opts ExportOptions, order map[string][]int) {
	attrs := buildBaseNodeAttrs(n, w, order)

	if opts.HighlightGoalGates {
		applyGoalGateHighlight(attrs, n)
	}

	applySemanticConfigAttrs(attrs, n.Config)

	if opts.IncludePrompts {
		applyPromptConfigAttrs(attrs, n.Config)
	}

	fmt.Fprintf(b, "  %s %s;\n", dotID(n.ID), formatDOTAttrs(attrs))
}

// buildBaseNodeAttrs creates the base attributes for a node (shape, label, execution order).
func buildBaseNodeAttrs(n *ir.Node, w *ir.Workflow, order map[string][]int) map[string]string {
	attrs := make(map[string]string)
	attrs["shape"] = resolveNodeShape(n, w)
	attrs["label"] = resolveNodeLabel(n, order, attrs)
	return attrs
}

// resolveNodeShape returns the DOT shape for a node, with start/exit overrides.
func resolveNodeShape(n *ir.Node, w *ir.Workflow) string {
	if n.ID == w.Start {
		return "Mdiamond"
	}
	if n.ID == w.Exit {
		return "Msquare"
	}
	return nodeShape(n.Kind)
}

// resolveNodeLabel returns the display label, annotated with execution order if applicable.
func resolveNodeLabel(n *ir.Node, order map[string][]int, attrs map[string]string) string {
	label := n.Label
	if label == "" {
		label = n.ID
	}
	ids, ok := order[n.ID]
	if !ok {
		return label
	}
	var orderStrs []string
	for _, idx := range ids {
		orderStrs = append(orderStrs, fmt.Sprintf("%d", idx))
	}
	attrs["style"] = "bold,filled"
	attrs["fillcolor"] = "#e0f0ff"
	return fmt.Sprintf("[%s] %s", strings.Join(orderStrs, ","), label)
}

// applyGoalGateHighlight highlights goal gate nodes if enabled.
func applyGoalGateHighlight(attrs map[string]string, n *ir.Node) {
	if ac, ok := n.Config.(ir.AgentConfig); ok && ac.GoalGate {
		attrs["style"] = "filled"
		attrs["fillcolor"] = "#ffcccc"
	}
}

// applySemanticConfigAttrs adds non-prompt runtime attributes unconditionally.
// These carry runtime semantics (timeout, mode, etc.) and must always be exported.
func applySemanticConfigAttrs(attrs map[string]string, cfg interface{}) {
	if !applySemanticNodeAttrs(attrs, cfg) {
		applySemanticStructuralAttrs(attrs, cfg)
	}
}

// applySemanticNodeAttrs handles human and tool semantic attrs. Returns true if matched.
func applySemanticNodeAttrs(attrs map[string]string, cfg interface{}) bool {
	switch c := cfg.(type) {
	case ir.HumanConfig:
		applyHumanSemanticAttrs(attrs, c)
	case ir.ToolConfig:
		applyToolSemanticAttrs(attrs, c)
	default:
		return false
	}
	return true
}

// applySemanticStructuralAttrs handles subgraph, parallel, fan-in, and manager_loop semantic attrs.
func applySemanticStructuralAttrs(attrs map[string]string, cfg interface{}) {
	switch c := cfg.(type) {
	case ir.SubgraphConfig:
		applySubgraphAttrs(attrs, c)
	case ir.ParallelConfig:
		applyParallelAttrs(attrs, c)
	case ir.FanInConfig:
		applyFanInAttrs(attrs, c)
	case ir.ManagerLoopConfig:
		applyManagerLoopAttrs(attrs, c)
	}
}

// applyPromptConfigAttrs adds prompt/text attributes gated behind IncludePrompts.
func applyPromptConfigAttrs(attrs map[string]string, cfg interface{}) {
	switch c := cfg.(type) {
	case ir.AgentConfig:
		applyAgentAttrs(attrs, c)
	case ir.ToolConfig:
		applyToolPromptAttrs(attrs, c)
	case ir.HumanConfig:
		applyHumanPromptAttrs(attrs, c)
	}
}

// applyAgentAttrs adds agent-specific attributes (prompt + runtime).
func applyAgentAttrs(attrs map[string]string, cfg ir.AgentConfig) {
	if cfg.Prompt != "" {
		attrs["prompt"] = escapeNewlines(cfg.Prompt)
	}
	if cfg.Model != "" {
		attrs["model"] = cfg.Model
	}
	if cfg.Provider != "" {
		attrs["provider"] = cfg.Provider
	}
	applyAgentRuntimeAttrs(attrs, cfg)
}

// applyAgentRuntimeAttrs adds backend and working_dir attributes.
func applyAgentRuntimeAttrs(attrs map[string]string, cfg ir.AgentConfig) {
	if cfg.Backend != "" {
		attrs["backend"] = cfg.Backend
	}
	if cfg.WorkingDir != "" {
		attrs["working_dir"] = cfg.WorkingDir
	}
}

// applyToolSemanticAttrs adds tool runtime attrs (always exported).
func applyToolSemanticAttrs(attrs map[string]string, cfg ir.ToolConfig) {
	if cfg.Timeout != 0 {
		attrs["timeout"] = formatDuration(cfg.Timeout)
	}
	if cfg.MarkerGrep != "" {
		attrs["marker_grep"] = cfg.MarkerGrep
	}
	if cfg.RouteRequired {
		attrs["route_required"] = "true"
	}
	applyToolOutputsAttrs(attrs, cfg)
}

// applyToolOutputsAttrs writes the output_limit and outputs DOT attrs.
// Extracted from applyToolSemanticAttrs to keep cyclomatic complexity ≤ 5.
func applyToolOutputsAttrs(attrs map[string]string, cfg ir.ToolConfig) {
	if cfg.OutputLimit > 0 {
		attrs["output_limit"] = strconv.Itoa(cfg.OutputLimit)
	}
	if len(cfg.Outputs) > 0 {
		attrs["outputs"] = strings.Join(cfg.Outputs, ",")
	}
}

// applyToolPromptAttrs adds tool command text (gated behind IncludePrompts).
func applyToolPromptAttrs(attrs map[string]string, cfg ir.ToolConfig) {
	if cfg.Command != "" {
		attrs["tool_command"] = escapeNewlines(cfg.Command)
	}
}

// applyHumanSemanticAttrs adds human mode, default, timeout, and timeout_action.
// These are all runtime-behavioral — not prompt text.
func applyHumanSemanticAttrs(attrs map[string]string, cfg ir.HumanConfig) {
	if cfg.Mode != "" {
		attrs["mode"] = cfg.Mode
	}
	if cfg.Default != "" {
		attrs["default"] = cfg.Default
	}
	if cfg.Timeout != 0 {
		attrs["timeout"] = formatDuration(cfg.Timeout)
	}
	if cfg.TimeoutAction != "" {
		attrs["timeout_action"] = cfg.TimeoutAction
	}
}

// applyHumanPromptAttrs adds human prompt text (gated behind IncludePrompts).
func applyHumanPromptAttrs(attrs map[string]string, cfg ir.HumanConfig) {
	if cfg.Prompt != "" {
		attrs["prompt"] = escapeNewlines(cfg.Prompt)
	}
}

// applySubgraphAttrs adds subgraph-specific attributes.
func applySubgraphAttrs(attrs map[string]string, cfg ir.SubgraphConfig) {
	if cfg.Ref != "" {
		attrs["ref"] = cfg.Ref
	}
}

// applyParallelAttrs adds parallel-specific attributes.
func applyParallelAttrs(attrs map[string]string, cfg ir.ParallelConfig) {
	if len(cfg.Targets) > 0 {
		attrs["targets"] = strings.Join(cfg.Targets, ",")
	}
}

// applyFanInAttrs adds fan_in-specific attributes.
func applyFanInAttrs(attrs map[string]string, cfg ir.FanInConfig) {
	if len(cfg.Sources) > 0 {
		attrs["sources"] = strings.Join(cfg.Sources, ",")
	}
}

// writeEdgeDOT emits a single DOT edge statement.
func writeEdgeDOT(b *strings.Builder, e *ir.Edge) {
	attrs := make(map[string]string)

	if e.Label != "" {
		attrs["label"] = e.Label
	}
	addEdgeConditionAttrs(attrs, e)
	addEdgeWeightAndRestart(attrs, e)

	fmt.Fprintf(b, "  %s -> %s", dotID(e.From), dotID(e.To))
	if len(attrs) > 0 {
		b.WriteString(" ")
		b.WriteString(formatDOTAttrs(attrs))
	}
	b.WriteString(";\n")
}

// addEdgeConditionAttrs resolves the condition string and adds it to attrs.
func addEdgeConditionAttrs(attrs map[string]string, e *ir.Edge) {
	if e.Condition == nil {
		return
	}
	condStr := resolveConditionStr(e.Condition)
	if condStr == "" {
		return
	}
	// Lower namespaced variables to DOT-compatible format:
	// ctx.outcome → outcome (Tracker resolves bare keys from context)
	condStr = lowerConditionNamespaces(condStr)
	// If there's no separate label, use the condition text as the edge label.
	if e.Label == "" {
		attrs["label"] = condStr
	}
	attrs["condition"] = condStr
}

// resolveConditionStr returns the formatted or raw condition string.
func resolveConditionStr(cond *ir.Condition) string {
	if cond.Parsed != nil {
		return formatCondition(cond.Parsed)
	}
	return cond.Raw
}

// addEdgeWeightAndRestart adds weight and restart attributes to attrs.
func addEdgeWeightAndRestart(attrs map[string]string, e *ir.Edge) {
	if e.Weight != 0 {
		attrs["weight"] = fmt.Sprintf("%d", e.Weight)
	}
	if e.Restart {
		attrs["restart"] = "true"
		attrs["style"] = "dashed"
	}
}

// formatDOTAttrs renders a map of DOT attributes as a bracketed list,
// with keys in sorted order for deterministic output.
func formatDOTAttrs(attrs map[string]string) string {
	if len(attrs) == 0 {
		return ""
	}

	// Sort keys for deterministic output.
	keys := make([]string, 0, len(attrs))
	for k := range attrs {
		keys = append(keys, k)
	}
	sortStrings(keys)

	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s=%s", k, dotQuote(attrs[k])))
	}
	return "[" + strings.Join(parts, ", ") + "]"
}

// dotID formats a string as a valid DOT identifier.
// If the string is a simple identifier (alphanumeric + underscore, not starting
// with a digit), it is returned as-is. Otherwise it is double-quoted.
func dotID(s string) string {
	if s == "" {
		return `""`
	}
	if isSimpleDOTID(s) {
		return s
	}
	return dotQuote(s)
}

// isSimpleDOTID returns true if s is a valid unquoted DOT identifier.
func isSimpleDOTID(s string) bool {
	if len(s) == 0 || isDigit(s[0]) {
		return false
	}
	for _, ch := range s {
		if !isValidDOTChar(ch) {
			return false
		}
	}
	return true
}

// isDigit returns true if ch is an ASCII digit.
func isDigit(ch byte) bool {
	return ch >= '0' && ch <= '9'
}

// isValidDOTChar returns true if ch is valid in an unquoted DOT identifier
// (alphanumeric or underscore).
func isValidDOTChar(ch rune) bool {
	return isAlpha(ch) || (ch >= '0' && ch <= '9') || ch == '_'
}

// isAlpha returns true if ch is an ASCII letter.
func isAlpha(ch rune) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z')
}

// dotQuote wraps a string in double quotes, escaping internal quotes and
// backslashes. Preserves DOT escape sequences like \n, \l, \r.
func dotQuote(s string) string {
	var b strings.Builder
	b.WriteByte('"')
	for i := 0; i < len(s); i++ {
		i = dotQuoteChar(&b, s, i)
	}
	b.WriteByte('"')
	return b.String()
}

// dotQuoteChar writes the character at s[i] to b, handling escaping.
// Returns the index to use for the next iteration (advanced past escape sequences).
func dotQuoteChar(b *strings.Builder, s string, i int) int {
	ch := s[i]
	if ch == '"' {
		b.WriteString(`\"`)
		return i
	}
	if ch != '\\' {
		b.WriteByte(ch)
		return i
	}
	// Check if this backslash is part of a DOT escape sequence.
	if i+1 < len(s) && isDOTEscapeChar(s[i+1]) {
		b.WriteByte('\\')
		b.WriteByte(s[i+1])
		return i + 1
	}
	b.WriteString(`\\`)
	return i
}

// isDOTEscapeChar returns true if ch follows a backslash in a DOT escape sequence.
func isDOTEscapeChar(ch byte) bool {
	return ch == 'n' || ch == 'l' || ch == 'r'
}

// escapeNewlines replaces literal newlines with the DOT \n escape for
// multi-line attribute values.
func escapeNewlines(s string) string {
	return strings.ReplaceAll(s, "\n", `\n`)
}

// lowerConditionNamespaces strips the ctx. prefix from condition variables
// for DOT-compatible output. Tracker's condition evaluator resolves bare
// variable names (e.g., "outcome") from the pipeline context, so the
// Dippin namespace prefix must be removed.
func lowerConditionNamespaces(cond string) string {
	return strings.ReplaceAll(cond, "ctx.", "")
}

// --- Condition formatting ---
// Replicates the formatter's condition serialization for DOT attribute values.

const (
	precOr  = 1
	precAnd = 2
	precNot = 3
)

func formatCondition(expr ir.ConditionExpr) string {
	return formatConditionExpr(expr, 0)
}

func formatConditionExpr(expr ir.ConditionExpr, parentPrec int) string {
	switch e := expr.(type) {
	case ir.CondCompare:
		return formatDOTCompare(e)
	case ir.CondAnd:
		return formatBinaryOp(e.Left, e.Right, "and", precAnd, parentPrec)
	case ir.CondOr:
		return formatBinaryOp(e.Left, e.Right, "or", precOr, parentPrec)
	case ir.CondNot:
		return "not " + formatConditionExpr(e.Inner, precNot)
	default:
		return ""
	}
}

// formatDOTCompare formats a compare expression, stripping the ctx. prefix.
func formatDOTCompare(e ir.CondCompare) string {
	variable := strings.TrimPrefix(e.Variable, "ctx.")
	return fmt.Sprintf("%s %s %s", variable, e.Op, e.Value)
}

// formatBinaryOp formats an and/or expression with optional parenthesization.
func formatBinaryOp(left, right ir.ConditionExpr, op string, prec, parentPrec int) string {
	s := fmt.Sprintf("%s %s %s",
		formatConditionExpr(left, prec),
		op,
		formatConditionExpr(right, prec))
	if parentPrec != 0 && parentPrec != prec {
		return "(" + s + ")"
	}
	return s
}

// --- Duration formatting ---

func formatDuration(d time.Duration) string {
	if d == 0 {
		return "0s"
	}
	parts, d := appendDurationParts(nil, d)
	if len(parts) == 0 {
		return d.String()
	}
	return strings.Join(parts, "")
}

// appendDurationParts appends hours, minutes, and seconds components.
func appendDurationParts(parts []string, d time.Duration) ([]string, time.Duration) {
	if h := int(d.Hours()); h > 0 {
		parts = append(parts, fmt.Sprintf("%dh", h))
		d -= time.Duration(h) * time.Hour
	}
	if m := int(d.Minutes()); m > 0 {
		parts = append(parts, fmt.Sprintf("%dm", m))
		d -= time.Duration(m) * time.Minute
	}
	if s := int(d.Seconds()); s > 0 {
		parts = append(parts, fmt.Sprintf("%ds", s))
	}
	return parts, d
}

// sortStrings sorts a string slice in place. Avoids importing sort for this
// single use.
func sortStrings(s []string) {
	for i := 1; i < len(s); i++ {
		for j := i; j > 0 && s[j] < s[j-1]; j-- {
			s[j], s[j-1] = s[j-1], s[j]
		}
	}
}

// applyManagerLoopScalarAttrs writes scalar manager_loop config fields as DOT node attributes.
func applyManagerLoopScalarAttrs(attrs map[string]string, cfg ir.ManagerLoopConfig) {
	if cfg.SubgraphRef != "" {
		attrs["subgraph_ref"] = cfg.SubgraphRef
	}
	if cfg.PollInterval != 0 {
		attrs["poll_interval"] = formatDuration(cfg.PollInterval)
	}
	if cfg.MaxCycles != 0 {
		attrs["max_cycles"] = fmt.Sprintf("%d", cfg.MaxCycles)
	}
	if s := flattenSteerContext(cfg.SteerContext); s != "" {
		attrs["steer_context"] = s
	}
}

// applyManagerLoopConditionAttrs writes condition manager_loop config fields as DOT node attributes.
func applyManagerLoopConditionAttrs(attrs map[string]string, cfg ir.ManagerLoopConfig) {
	if s := dotManagerLoopConditionText(cfg.StopCondition); s != "" {
		attrs["stop_condition"] = s
	}
	if s := dotManagerLoopConditionText(cfg.SteerCondition); s != "" {
		attrs["steer_condition"] = s
	}
}

// dotManagerLoopConditionText returns the best textual form of a node condition:
// prefers Raw when populated; otherwise formats Parsed via formatConditionExpr.
// Returns "" when the condition is nil/empty.
func dotManagerLoopConditionText(c *ir.Condition) string {
	if c == nil {
		return ""
	}
	if c.Raw != "" {
		return c.Raw
	}
	if c.Parsed != nil {
		return formatCondition(c.Parsed)
	}
	return ""
}

// applyManagerLoopAttrs writes manager_loop config fields as DOT node attributes.
// The steer_context map is flattened to canonical sorted "k=v,k=v" so the
// round-trip through DOT → migrate → IR is lossless.
func applyManagerLoopAttrs(attrs map[string]string, cfg ir.ManagerLoopConfig) {
	applyManagerLoopScalarAttrs(attrs, cfg)
	applyManagerLoopConditionAttrs(attrs, cfg)
}

// steerContextEncoder replaces reserved steer_context delimiter characters with
// their percent-encoded equivalents. '%' must be replaced first so it is not
// double-encoded.
var steerContextEncoder = strings.NewReplacer(
	"%", "%25",
	",", "%2C",
	"=", "%3D",
)

// encodeSteerContextToken percent-encodes the three reserved characters
// used as delimiters in the flattened steer_context representation (',', '=')
// and the escape character itself ('%'). This keeps DOT round-trip lossless
// even when keys or values contain reserved characters.
func encodeSteerContextToken(s string) string {
	if !strings.ContainsAny(s, ",=%") {
		return s
	}
	return steerContextEncoder.Replace(s)
}

// flattenSteerContext produces canonical sorted "k=v,k=v" from the map.
// Empty map returns empty string (caller suppresses the attr).
// Reserved characters (',', '=', '%') in keys and values are percent-encoded
// so the round-trip through DOT → migrate stays lossless.
func flattenSteerContext(m map[string]string) string {
	if len(m) == 0 {
		return ""
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sortStrings(keys)
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, encodeSteerContextToken(k)+"="+encodeSteerContextToken(m[k]))
	}
	return strings.Join(parts, ",")
}
