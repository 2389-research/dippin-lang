// Package formatter implements canonical Dippin source formatting.
// Given an ir.Workflow, it produces deterministic .dip source text.
// The output is idempotent: Format(w) always produces the same string
// for the same IR state.
package formatter

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/2389-research/dippin-lang/ir"
)

// Format renders a workflow to canonical Dippin source text.
// The output always ends with exactly one trailing newline.
func Format(w *ir.Workflow) string {
	wr := &writer{}
	writeWorkflowHeader(wr, w)
	writeWorkflowSections(wr, w)
	return wr.String()
}

// writeWorkflowSections emits all optional top-level sections in canonical order.
func writeWorkflowSections(wr *writer, w *ir.Workflow) {
	if !isDefaultsZero(w.Defaults) {
		wr.blank()
		writeDefaults(wr, w.Defaults)
	}
	if len(w.Vars) > 0 {
		wr.blank()
		writeVars(wr, w.Vars)
	}
	for _, n := range w.Nodes {
		wr.blank()
		writeNode(wr, n)
	}
	writeWorkflowTailSections(wr, w)
}

// writeWorkflowTailSections emits stylesheet and edges after nodes.
func writeWorkflowTailSections(wr *writer, w *ir.Workflow) {
	if len(w.Stylesheet) > 0 {
		wr.blank()
		writeStylesheet(wr, w.Stylesheet)
	}
	if hasVisibleEdges(w) || w.ElseTarget != "" {
		wr.blank()
		writeEdges(wr, w)
	}
}

// hasVisibleEdges reports whether any edge survives formatting — i.e. is not a
// redundant fan edge stripped per #136. Guards the edges-block header so an
// all-redundant edge list does not emit a dangling empty block.
func hasVisibleEdges(w *ir.Workflow) bool {
	for _, e := range w.Edges {
		if !ir.IsRedundantFanEdge(w, e) {
			return true
		}
	}
	return false
}

// writer wraps a strings.Builder with indentation tracking.
type writer struct {
	buf    strings.Builder
	indent int // current indent level (each level = 2 spaces)
}

// line writes a single indented line followed by a newline.
func (wr *writer) line(format string, args ...any) {
	content := fmt.Sprintf(format, args...)
	content = strings.TrimRight(content, " \t")
	prefix := strings.Repeat("  ", wr.indent)
	wr.buf.WriteString(prefix)
	wr.buf.WriteString(content)
	wr.buf.WriteByte('\n')
}

// blank writes an empty line.
func (wr *writer) blank() {
	wr.buf.WriteByte('\n')
}

// push increases the indentation level by one.
func (wr *writer) push() {
	wr.indent++
}

// pop decreases the indentation level by one.
func (wr *writer) pop() {
	wr.indent--
}

// multilineBlock emits a multiline field in the canonical form:
//
//	key:
//	  <line1>
//	  <line2>
func (wr *writer) multilineBlock(key, content string) {
	wr.line("%s:", key)
	content = strings.TrimRight(content, " \t\n\r")
	if content == "" {
		return
	}
	wr.push()
	for _, l := range strings.Split(content, "\n") {
		l = strings.TrimRight(l, " \t\r")
		if l == "" {
			wr.blank()
		} else {
			wr.line("%s", l)
		}
	}
	wr.pop()
}

// String returns the final output with exactly one trailing newline.
func (wr *writer) String() string {
	s := wr.buf.String()
	s = strings.TrimRight(s, "\n\r \t")
	return s + "\n"
}

// --- Section emitters ---

func writeWorkflowHeader(wr *writer, w *ir.Workflow) {
	if v, err := strconv.Atoi(w.Version); err == nil && v > 1 {
		wr.line("dip %d", v)
		wr.blank()
	}
	wr.line("workflow %s", w.Name)
	wr.push()
	if w.Goal != "" {
		wr.line("goal: %s", quoteValue(w.Goal))
	}
	if len(w.Requires) > 0 {
		wr.line("requires: %s", strings.Join(w.Requires, ", "))
	}
	wr.line("start: %s", w.Start)
	wr.line("exit: %s", w.Exit)
	// We keep the indent at 1 for the rest of the top-level sections
}

func writeDefaults(wr *writer, d ir.WorkflowDefaults) {
	wr.line("defaults")
	wr.push()
	writeDefaultsModelFields(wr, d)
	writeDefaultsRestartFields(wr, d)
	wr.pop()
}

// writeDefaultsModelFields writes model/provider/fidelity fields.
func writeDefaultsModelFields(wr *writer, d ir.WorkflowDefaults) {
	if d.Model != "" {
		wr.line("model: %s", quoteValue(d.Model))
	}
	if d.Provider != "" {
		wr.line("provider: %s", quoteValue(d.Provider))
	}
	writeDefaultsRetryPolicyFields(wr, d)
}

// writeDefaultsRetryPolicyFields writes retry_policy, max_retries, and fidelity fields.
func writeDefaultsRetryPolicyFields(wr *writer, d ir.WorkflowDefaults) {
	if d.RetryPolicy != "" {
		wr.line("retry_policy: %s", quoteValue(d.RetryPolicy))
	}
	if d.MaxRetries != 0 {
		wr.line("max_retries: %d", d.MaxRetries)
	}
	if d.Fidelity != "" {
		wr.line("fidelity: %s", quoteValue(d.Fidelity))
	}
}

// writeDefaultsRestartFields writes restart, cache, and compaction fields.
func writeDefaultsRestartFields(wr *writer, d ir.WorkflowDefaults) {
	if d.MaxRestarts != 0 {
		wr.line("max_restarts: %d", d.MaxRestarts)
	}
	if d.RestartTarget != "" {
		wr.line("restart_target: %s", d.RestartTarget)
	}
	if d.OnFailure != "" {
		wr.line("on_failure: %s", d.OnFailure)
	}
	writeDefaultsCompactionFields(wr, d)
}

// writeDefaultsCompactionFields writes cache_tools, compaction, and on_resume.
func writeDefaultsCompactionFields(wr *writer, d ir.WorkflowDefaults) {
	if d.CacheTools {
		wr.line("cache_tools: true")
	}
	if d.Compaction != "" {
		wr.line("compaction: %s", quoteValue(d.Compaction))
	}
	if d.OnResume != "" {
		wr.line("on_resume: %s", quoteValue(d.OnResume))
	}
	writeDefaultsBudgetFields(wr, d)
}

// writeDefaultsBudgetFields writes max_total_tokens, max_cost_cents, max_wall_time,
// and stall_timeout, then delegates to writeDefaultsToolSafetyFields.
func writeDefaultsBudgetFields(wr *writer, d ir.WorkflowDefaults) {
	if d.MaxTotalTokens != 0 {
		wr.line("max_total_tokens: %d", d.MaxTotalTokens)
	}
	if d.MaxCostCents != 0 {
		wr.line("max_cost_cents: %d", d.MaxCostCents)
	}
	if d.MaxWallTime != 0 {
		wr.line("max_wall_time: %s", formatDuration(d.MaxWallTime))
	}
	if d.StallTimeout != 0 {
		wr.line("stall_timeout: %s", formatDuration(d.StallTimeout))
	}
	writeDefaultsToolSafetyFields(wr, d)
}

// writeDefaultsToolSafetyFields writes tool_commands_allow and tool_denylist_add.
func writeDefaultsToolSafetyFields(wr *writer, d ir.WorkflowDefaults) {
	if d.ToolCommandsAllow != "" {
		wr.line("tool_commands_allow: %s", quoteValue(d.ToolCommandsAllow))
	}
	if d.ToolDenylistAdd != "" {
		wr.line("tool_denylist_add: %s", quoteValue(d.ToolDenylistAdd))
	}
}

func writeVars(wr *writer, vars map[string]string) {
	wr.line("vars")
	wr.push()
	keys := make([]string, 0, len(vars))
	for k := range vars {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		wr.line("%s: %s", k, quoteValue(vars[k]))
	}
	wr.pop()
}

func writeNode(wr *writer, n *ir.Node) {
	if writeStructuralNode(wr, n) {
		return
	}
	wr.line("%s %s", n.Kind, n.ID)
	wr.push()
	writeNodeConfigFields(wr, n)
	wr.pop()
}

// writeStructuralNode writes parallel/fan_in nodes. Returns true if handled.
func writeStructuralNode(wr *writer, n *ir.Node) bool {
	switch cfg := n.Config.(type) {
	case ir.ParallelConfig:
		if len(cfg.Branches) > 0 {
			writeParallelBlock(wr, n.ID, cfg)
		} else {
			wr.line("parallel %s -> %s", n.ID, strings.Join(cfg.Targets, ", "))
			writeNodeParams(wr, cfg.Params)
		}
		return true
	case ir.FanInConfig:
		wr.line("fan_in %s <- %s", n.ID, strings.Join(cfg.Sources, ", "))
		writeNodeParams(wr, cfg.Params)
		return true
	}
	return false
}

// writeNodeParams emits an indented params: block beneath a structural node's
// declaration line (inline parallel / fan_in). No-op when params is empty.
func writeNodeParams(wr *writer, params map[string]string) {
	if len(params) == 0 {
		return
	}
	wr.push()
	writeSortedMapBlock(wr, "params", params)
	wr.pop()
}

// writeParallelBlock writes a parallel node in block form with per-branch config
// and an optional params block.
func writeParallelBlock(wr *writer, id string, cfg ir.ParallelConfig) {
	wr.line("parallel %s", id)
	wr.push()
	for _, b := range cfg.Branches {
		writeBranch(wr, b)
	}
	writeSortedMapBlock(wr, "params", cfg.Params)
	wr.pop()
}

// writeBranch writes a single branch entry in block-form parallel.
func writeBranch(wr *writer, b ir.BranchConfig) {
	wr.line("branch: %s", b.Target)
	if !branchHasFields(b) {
		return
	}
	wr.push()
	writeBranchFields(wr, b)
	wr.pop()
}

// branchHasFields reports whether a branch carries any optional field beyond its target.
func branchHasFields(b ir.BranchConfig) bool {
	return b.Model != "" || b.Provider != "" || b.Fidelity != "" ||
		b.ToolAccess != "" || branchHasSandboxFields(b)
}

// branchHasSandboxFields reports whether a branch carries any sandbox/limit field
// (writable_paths or last_response_truncate).
func branchHasSandboxFields(b ir.BranchConfig) bool {
	return len(b.WritablePaths) > 0 || b.LastResponseTruncate > 0
}

// writeBranchFields writes the optional fields within a branch.
func writeBranchFields(wr *writer, b ir.BranchConfig) {
	writeBranchScalarFields(wr, b)
	if len(b.WritablePaths) > 0 {
		wr.line("writable_paths: %s", strings.Join(b.WritablePaths, ", "))
	}
	if b.LastResponseTruncate > 0 {
		wr.line("last_response_truncate: %d", b.LastResponseTruncate)
	}
}

// writeBranchScalarFields writes the scalar per-branch overrides.
func writeBranchScalarFields(wr *writer, b ir.BranchConfig) {
	if b.Model != "" {
		wr.line("model: %s", quoteValue(b.Model))
	}
	if b.Provider != "" {
		wr.line("provider: %s", quoteValue(b.Provider))
	}
	if b.Fidelity != "" {
		wr.line("fidelity: %s", quoteValue(b.Fidelity))
	}
	if b.ToolAccess != "" {
		wr.line("tool_access: %s", quoteValue(b.ToolAccess))
	}
}

// writeNodeConfigFields dispatches to the appropriate field writer based on config type.
func writeNodeConfigFields(wr *writer, n *ir.Node) {
	if writePrimaryConfigFields(wr, n) {
		return
	}
	writeSecondaryConfigFields(wr, n)
}

// writePrimaryConfigFields handles agent and human config. Returns true if handled.
func writePrimaryConfigFields(wr *writer, n *ir.Node) bool {
	switch cfg := n.Config.(type) {
	case ir.AgentConfig:
		writeAgentFields(wr, n, cfg)
	case ir.HumanConfig:
		writeHumanFields(wr, n, cfg)
	default:
		return false
	}
	return true
}

// writeSecondaryConfigFields handles tool, subgraph, manager_loop, and conditional config.
func writeSecondaryConfigFields(wr *writer, n *ir.Node) {
	switch cfg := n.Config.(type) {
	case ir.ToolConfig:
		writeToolFields(wr, n, cfg)
	case ir.SubgraphConfig:
		writeSubgraphFields(wr, n, cfg)
	case ir.ManagerLoopConfig:
		writeManagerLoopFields(wr, n, cfg)
	case ir.ConditionalConfig:
		writeConditionalFields(wr, n)
	}
}

func writeAgentFields(wr *writer, n *ir.Node, cfg ir.AgentConfig) {
	writeCommonNodeFields(wr, n)
	writeAgentModelFields(wr, cfg)
	writeAgentRuntimeFields(wr, cfg)
	writeAgentResponseFields(wr, cfg)
	writeAgentBehaviorFields(wr, cfg)
	writeAgentCompactionFields(wr, cfg)
	writeRetryFields(wr, n)
	writeIOFields(wr, n)
	writeSortedMapBlock(wr, "params", cfg.Params)
	writeAgentPromptFields(wr, cfg)
}

// writeAgentPromptFields emits the four prompt-related fields: system_prompt
// and prompt, each in either *_file: directive form (if *File is set) or
// inline multiline-block form (if the content field is set). The conditional
// pairs preserve the authored form across format round-trips.
func writeAgentPromptFields(wr *writer, cfg ir.AgentConfig) {
	if cfg.SystemPromptFile != "" {
		wr.line("system_prompt_file: %s", quoteValue(cfg.SystemPromptFile))
	} else if cfg.SystemPrompt != "" {
		wr.multilineBlock("system_prompt", cfg.SystemPrompt)
	}
	if cfg.PromptFile != "" {
		wr.line("prompt_file: %s", quoteValue(cfg.PromptFile))
	} else if cfg.Prompt != "" {
		wr.multilineBlock("prompt", cfg.Prompt)
	}
}

// writeCommonNodeFields writes label and class fields common to all node types.
func writeCommonNodeFields(wr *writer, n *ir.Node) {
	if n.Label != "" {
		wr.line("label: %s", quoteValue(n.Label))
	}
	if len(n.Classes) > 0 {
		wr.line("class: %s", strings.Join(n.Classes, ", "))
	}
}

// writeAgentModelFields writes model-related fields for agent nodes.
func writeAgentModelFields(wr *writer, cfg ir.AgentConfig) {
	if cfg.Model != "" {
		wr.line("model: %s", quoteValue(cfg.Model))
	}
	if cfg.Provider != "" {
		wr.line("provider: %s", quoteValue(cfg.Provider))
	}
	if cfg.ReasoningEffort != "" {
		wr.line("reasoning_effort: %s", quoteValue(cfg.ReasoningEffort))
	}
	if cfg.Fidelity != "" {
		wr.line("fidelity: %s", quoteValue(cfg.Fidelity))
	}
}

// writeAgentRuntimeFields writes runtime behavior fields for agent nodes.
func writeAgentRuntimeFields(wr *writer, cfg ir.AgentConfig) {
	if cfg.Backend != "" {
		wr.line("backend: %s", quoteValue(cfg.Backend))
	}
	if cfg.WorkingDir != "" {
		wr.line("working_dir: %s", quoteValue(cfg.WorkingDir))
	}
	if strings.TrimSpace(cfg.ToolAccess) != "" {
		wr.line("tool_access: %s", quoteValue(cfg.ToolAccess))
	}
	writeAgentSandboxFields(wr, cfg)
}

// writeAgentSandboxFields writes sandbox/limit fields for agent nodes.
func writeAgentSandboxFields(wr *writer, cfg ir.AgentConfig) {
	if len(cfg.WritablePaths) > 0 {
		wr.line("writable_paths: %s", strings.Join(cfg.WritablePaths, ", "))
	}
	if cfg.LastResponseTruncate > 0 {
		wr.line("last_response_truncate: %d", cfg.LastResponseTruncate)
	}
}

// writeAgentResponseFields writes response format fields for agent nodes.
func writeAgentResponseFields(wr *writer, cfg ir.AgentConfig) {
	if cfg.ResponseFormat != "" {
		wr.line("response_format: %s", quoteValue(cfg.ResponseFormat))
	}
	if cfg.ResponseSchema != "" {
		wr.multilineBlock("response_schema", cfg.ResponseSchema)
	}
}

// writeAgentBehaviorFields writes behavior-related fields for agent nodes.
func writeAgentBehaviorFields(wr *writer, cfg ir.AgentConfig) {
	if cfg.GoalGate {
		wr.line("goal_gate: true")
	}
	if cfg.AutoStatus {
		wr.line("auto_status: true")
	}
	if cfg.MaxTurns != 0 {
		wr.line("max_turns: %d", cfg.MaxTurns)
	}
	if cfg.CmdTimeout != 0 {
		wr.line("cmd_timeout: %s", formatDuration(cfg.CmdTimeout))
	}
}

// writeAgentCompactionFields writes compaction-related fields for agent nodes.
func writeAgentCompactionFields(wr *writer, cfg ir.AgentConfig) {
	if cfg.CacheTools {
		wr.line("cache_tools: true")
	}
	if cfg.Compaction != "" {
		wr.line("compaction: %s", quoteValue(cfg.Compaction))
	}
	if cfg.CompactionThreshold != 0 {
		wr.line("compaction_threshold: %s", formatFloat(cfg.CompactionThreshold))
	}
}

// writeRetryFields writes retry-related fields common to all node types.
func writeRetryFields(wr *writer, n *ir.Node) {
	writeRetryPolicyFields(wr, n.Retry)
	writeRetryTargetFields(wr, n.Retry)
}

// writeRetryPolicyFields writes policy, max_retries, and base_delay fields.
func writeRetryPolicyFields(wr *writer, r ir.RetryConfig) {
	if r.Policy != "" {
		wr.line("retry_policy: %s", quoteValue(r.Policy))
	}
	if r.MaxRetries != 0 {
		wr.line("max_retries: %d", r.MaxRetries)
	}
	if r.BaseDelay != 0 {
		wr.line("base_delay: %s", formatDuration(r.BaseDelay))
	}
}

// writeRetryTargetFields writes retry_target and fallback_target fields.
func writeRetryTargetFields(wr *writer, r ir.RetryConfig) {
	if r.RetryTarget != "" {
		wr.line("retry_target: %s", r.RetryTarget)
	}
	if r.FallbackTarget != "" {
		wr.line("fallback_target: %s", r.FallbackTarget)
	}
}

// writeIOFields writes reads and writes fields common to all node types.
func writeIOFields(wr *writer, n *ir.Node) {
	if len(n.IO.Reads) > 0 {
		wr.line("reads: %s", strings.Join(n.IO.Reads, ", "))
	}
	if len(n.IO.Writes) > 0 {
		wr.line("writes: %s", strings.Join(n.IO.Writes, ", "))
	}
}

func writeHumanFields(wr *writer, n *ir.Node, cfg ir.HumanConfig) {
	writeCommonNodeFields(wr, n)
	writeHumanModeFields(wr, cfg)
	writeIOFields(wr, n)
	if cfg.Prompt != "" {
		wr.multilineBlock("prompt", cfg.Prompt)
	}
}

func writeHumanModeFields(wr *writer, cfg ir.HumanConfig) {
	if cfg.Mode != "" {
		wr.line("mode: %s", quoteValue(cfg.Mode))
	}
	if cfg.Default != "" {
		wr.line("default: %s", quoteValue(cfg.Default))
	}
	if cfg.QuestionsKey != "" {
		wr.line("questions_key: %s", quoteValue(cfg.QuestionsKey))
	}
	if cfg.AnswersKey != "" {
		wr.line("answers_key: %s", quoteValue(cfg.AnswersKey))
	}
	writeHumanTimeoutFields(wr, cfg)
}

// writeHumanTimeoutFields writes timeout and timeout_action fields for human nodes.
func writeHumanTimeoutFields(wr *writer, cfg ir.HumanConfig) {
	if cfg.Timeout != 0 {
		wr.line("timeout: %s", formatDuration(cfg.Timeout))
	}
	if cfg.TimeoutAction != "" {
		wr.line("timeout_action: %s", quoteValue(cfg.TimeoutAction))
	}
}

func writeToolRoutingFields(wr *writer, cfg ir.ToolConfig) {
	if cfg.MarkerGrep != "" {
		wr.line("marker_grep: %s", quoteValue(cfg.MarkerGrep))
	}
	if cfg.RouteRequired {
		wr.line("route_required: true")
	}
	if cfg.OutputLimit > 0 {
		wr.line("output_limit: %d", cfg.OutputLimit)
	}
}

func writeToolFields(wr *writer, n *ir.Node, cfg ir.ToolConfig) {
	writeCommonNodeFields(wr, n)
	if len(cfg.Outputs) > 0 {
		wr.line("outputs: %s", strings.Join(cfg.Outputs, ", "))
	}
	writeToolRoutingFields(wr, cfg)
	if cfg.Timeout != 0 {
		wr.line("timeout: %s", formatDuration(cfg.Timeout))
	}
	writeIOFields(wr, n)
	if cfg.CommandFile != "" {
		wr.line("command_file: %s", quoteValue(cfg.CommandFile))
	} else if cfg.Command != "" {
		wr.multilineBlock("command", cfg.Command)
	}
}

func writeConditionalFields(wr *writer, n *ir.Node) {
	writeCommonNodeFields(wr, n)
	writeIOFields(wr, n)
}

func writeSubgraphFields(wr *writer, n *ir.Node, cfg ir.SubgraphConfig) {
	writeCommonNodeFields(wr, n)
	if cfg.Ref != "" {
		wr.line("ref: %s", quoteValue(cfg.Ref))
	}
	writeSortedMapBlock(wr, "params", cfg.Params)
}

// writeSortedMapBlock emits a named block containing sorted "key: value"
// lines. Used for any map[string]string field (subgraph params, manager_loop
// steer_context) so the output is deterministic and idempotent.
func writeSortedMapBlock(wr *writer, header string, m map[string]string) {
	if len(m) == 0 {
		return
	}
	wr.line("%s:", header)
	wr.push()
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		wr.line("%s: %s", k, quoteValue(m[k]))
	}
	wr.pop()
}

// writeManagerLoopFields writes the fields of a manager_loop node in canonical order.
func writeManagerLoopFields(wr *writer, n *ir.Node, cfg ir.ManagerLoopConfig) {
	writeCommonNodeFields(wr, n)
	writeRetryFields(wr, n)
	writeIOFields(wr, n)
	if cfg.SubgraphRef != "" {
		wr.line("subgraph_ref: %s", quoteValue(cfg.SubgraphRef))
	}
	if cfg.PollInterval != 0 {
		wr.line("poll_interval: %s", formatDuration(cfg.PollInterval))
	}
	if cfg.MaxCycles != 0 {
		wr.line("max_cycles: %d", cfg.MaxCycles)
	}
	writeManagerLoopConditions(wr, cfg)
	writeSortedMapBlock(wr, "steer_context", cfg.SteerContext)
}

// writeManagerLoopConditions writes stop_condition and steer_condition if set.
// Raw already contains any quoting/escaping required by the condition-expression
// syntax (e.g., "success" as a literal), so it is emitted as-is without
// additional quoting.
func writeManagerLoopConditions(wr *writer, cfg ir.ManagerLoopConfig) {
	if s := managerLoopConditionText(cfg.StopCondition); s != "" {
		wr.line("stop_condition: %s", s)
	}
	if s := managerLoopConditionText(cfg.SteerCondition); s != "" {
		wr.line("steer_condition: %s", s)
	}
}

// managerLoopConditionText returns the best textual form of a node condition:
// prefers Raw when populated; otherwise formats Parsed via the same helper
// the edge path uses. Returns "" when the condition is nil/empty.
func managerLoopConditionText(c *ir.Condition) string {
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

// selectorSpecificity returns a numeric specificity for sorting.
func selectorSpecificity(sel ir.StyleSelector) int {
	switch sel.Kind {
	case "universal":
		return 0
	case "kind":
		return 1
	case "class":
		return 2
	case "id":
		return 3
	default:
		return 0
	}
}

// writeStylesheet emits stylesheet rules sorted by specificity.
func writeStylesheet(wr *writer, rules []ir.StylesheetRule) {
	sorted := make([]ir.StylesheetRule, len(rules))
	copy(sorted, rules)
	sort.Slice(sorted, func(i, j int) bool {
		si := selectorSpecificity(sorted[i].Selector)
		sj := selectorSpecificity(sorted[j].Selector)
		if si != sj {
			return si < sj
		}
		return sorted[i].Selector.Value < sorted[j].Selector.Value
	})

	var raw strings.Builder
	for i, rule := range sorted {
		if i > 0 {
			raw.WriteByte('\n')
		}
		raw.WriteString(formatSelectorStr(rule.Selector))
		raw.WriteByte('\n')
		writeRuleProperties(&raw, rule.Properties)
	}
	wr.multilineBlock("stylesheet", raw.String())
}

// formatSelectorStr converts a selector to its string representation.
func formatSelectorStr(sel ir.StyleSelector) string {
	switch sel.Kind {
	case "universal":
		return "*"
	case "class":
		return "." + sel.Value
	case "id":
		return "#" + sel.Value
	default:
		return sel.Value
	}
}

// writeRuleProperties writes sorted key:value pairs for a rule.
func writeRuleProperties(b *strings.Builder, props map[string]string) {
	keys := make([]string, 0, len(props))
	for k := range props {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		b.WriteString("  " + k + ": " + props[k] + "\n")
	}
}

func writeEdges(wr *writer, w *ir.Workflow) {
	wr.line("edges")
	wr.push()
	for _, e := range w.Edges {
		if ir.IsRedundantFanEdge(w, e) {
			continue // inline parallel/fan_in list is authoritative (#136 / DIP153)
		}
		writeEdge(wr, w, e)
	}
	if w.ElseTarget != "" {
		wr.line("else -> %s", w.ElseTarget)
	}
	wr.pop()
}

func writeEdge(wr *writer, w *ir.Workflow, e *ir.Edge) {
	if e.Comment != "" {
		wr.line("# %s", e.Comment)
	}
	var parts []string
	parts = append(parts, fmt.Sprintf("%s -> %s", e.From, e.To))
	parts = appendEdgeCondition(parts, w, e)
	parts = appendEdgeAttrs(parts, e)
	wr.line("%s", strings.Join(parts, "  "))
}

// appendEdgeCondition appends the condition part if one is present. When the
// source node has a natural outcome channel and the condition is exactly an
// equality against it (e.g. ctx.outcome = X), it re-emits as the `on X`
// shorthand — which also migrates hand-written `when` of that shape. Everything
// else, and any source node without an outcome channel, stays `when ...`.
func appendEdgeCondition(parts []string, w *ir.Workflow, e *ir.Edge) []string {
	if e.Condition == nil {
		return parts
	}
	text := conditionText(e.Condition)
	if text == "" {
		return parts
	}
	channel, ok := w.Node(e.From).OutcomeChannel()
	if ok {
		if tok, isOn := onShorthand(text, channel); isOn {
			return append(parts, "on "+tok)
		}
	}
	return append(parts, "when "+text)
}

// conditionText returns the canonical source text of a condition, preferring the
// parsed AST when present and falling back to the raw string.
func conditionText(c *ir.Condition) string {
	if c.Parsed != nil {
		return formatCondition(c.Parsed)
	}
	return c.Raw
}

// onShorthand returns the outcome token when text is exactly an equality test
// against the source node's outcome channel (e.g. "ctx.outcome = success", the
// shape `on success` desugars to) and the value is a bare identifier. Matching
// the node's own channel guarantees the re-emitted `on` parses back identically.
//
// ir.IsOutcomeToken is stricter than `needsQuoting`: the latter permits `:` (the
// .dip lexer splits it into a separate token) and `.`/`/` (the .dip lexer keeps
// them, but the tree-sitter grammar's identifier rule rejects them). Restricting
// to `[a-zA-Z0-9][a-zA-Z0-9_-]*` keeps fmt from emitting an `on` that fails to
// re-parse in either grammar, and mirrors the parser's accept rule exactly.
func onShorthand(text, channel string) (string, bool) {
	if tok, ok := strings.CutPrefix(text, channel+" = "); ok && ir.IsOutcomeToken(tok) {
		return tok, true
	}
	return "", false
}

// appendEdgeAttrs appends label, choice, weight, loop (back-edge), and override
// parts. A restart edge is emitted as the canonical `loop` keyword, which also
// migrates hand-written `restart: true` to the scannable form.
func appendEdgeAttrs(parts []string, e *ir.Edge) []string {
	parts = appendEdgeLabelParts(parts, e)
	if e.Weight != 0 {
		parts = append(parts, fmt.Sprintf("weight: %d", e.Weight))
	}
	if e.Restart {
		parts = append(parts, "loop")
	}
	if e.Override {
		parts = append(parts, "override: true")
	}
	return parts
}

// appendEdgeLabelParts emits the label-family attributes — `label` (display) then
// `choice` (the carried human-gate routing key) — in canonical order.
func appendEdgeLabelParts(parts []string, e *ir.Edge) []string {
	if e.Label != "" {
		parts = append(parts, fmt.Sprintf("label: %s", quoteValue(e.Label)))
	}
	if e.Choice != "" {
		parts = append(parts, fmt.Sprintf("choice: %s", quoteValue(e.Choice)))
	}
	return parts
}

// --- Condition formatting ---

const (
	precOr  = 1
	precAnd = 2
	precNot = 3
)

func formatCondition(expr ir.ConditionExpr) string {
	return formatConditionExpr(expr, 0)
}

// formatCondValue renders a comparison's right-hand value. A value that is a
// reserved value-less edge flag (e.g. `loop`) must be quoted, or re-parsing the
// emitted condition would treat it as the edge flag and truncate the condition —
// the AST path drops the quotes the raw text carried (see ir.IsReservedEdgeFlag).
func formatCondValue(v string) string {
	if ir.IsReservedEdgeFlag(v) {
		return `"` + v + `"`
	}
	return v
}

func formatConditionExpr(expr ir.ConditionExpr, parentPrec int) string {
	switch e := expr.(type) {
	case ir.CondCompare:
		return fmt.Sprintf("%s %s %s", e.Variable, e.Op, formatCondValue(e.Value))
	case ir.CondAnd:
		return fmtBinaryCondExpr(e.Left, e.Right, "and", precAnd, parentPrec)
	case ir.CondOr:
		return fmtBinaryCondExpr(e.Left, e.Right, "or", precOr, parentPrec)
	case ir.CondNot:
		return "not " + formatConditionExpr(e.Inner, precNot)
	default:
		return ""
	}
}

// fmtBinaryCondExpr formats a binary condition (and/or) with optional parenthesization.
func fmtBinaryCondExpr(left, right ir.ConditionExpr, op string, prec, parentPrec int) string {
	s := fmt.Sprintf("%s %s %s",
		formatConditionExpr(left, prec),
		op,
		formatConditionExpr(right, prec))
	if parentPrec != 0 && parentPrec != prec {
		return "(" + s + ")"
	}
	return s
}

func quoteValue(s string) string {
	if s == "" {
		return `""`
	}
	if needsQuoting(s) {
		return `"` + s + `"`
	}
	return s
}

// needsQuoting returns true if the value needs to be enclosed in double quotes.
// Simple identifiers (alphanumeric, underscore, dash, dot, slash, colon) are unquoted.
func needsQuoting(s string) bool {
	for _, ch := range s {
		if !isUnquotedChar(ch) {
			return true
		}
	}
	return false
}

// unquotedPunctuation contains punctuation characters allowed in unquoted Dippin values.
const unquotedPunctuation = "_-./:"

// isUnquotedChar returns true if ch is allowed in an unquoted Dippin value.
func isUnquotedChar(ch rune) bool {
	return isUnquotedAlphanumeric(ch) || strings.ContainsRune(unquotedPunctuation, ch)
}

// isUnquotedAlphanumeric returns true if ch is an ASCII letter or digit.
func isUnquotedAlphanumeric(ch rune) bool {
	return isASCIILetter(ch) || (ch >= '0' && ch <= '9')
}

// isASCIILetter returns true if ch is an ASCII letter.
func isASCIILetter(ch rune) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z')
}

// formatDuration renders a time.Duration as a compact human-readable string
// suitable for Dippin source: "30s", "5m", "1h30m".
func formatDuration(d time.Duration) string {
	if d == 0 {
		return "0s"
	}
	parts, d := fmtDurationParts(nil, d)
	if len(parts) == 0 {
		// Sub-second durations.
		return d.String()
	}
	return strings.Join(parts, "")
}

// fmtDurationParts appends hours, minutes, and seconds components.
func fmtDurationParts(parts []string, d time.Duration) ([]string, time.Duration) {
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

// formatFloat renders a float64 without unnecessary trailing zeros.
func formatFloat(f float64) string {
	s := strconv.FormatFloat(f, 'f', -1, 64)
	if !strings.Contains(s, ".") {
		s += ".0"
	}
	return s
}

func isDefaultsZero(d ir.WorkflowDefaults) bool {
	return d == ir.WorkflowDefaults{}
}
