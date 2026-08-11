package simulate

import (
	"regexp"
	"strings"

	"github.com/2389-research/dippin-lang/ir"
)

// applyNodeDefaults seeds default context values for agent and tool nodes.
// Per-node scenario values (e.g., --scenario NodeName.outcome=fail) are
// applied first, then global defaults fill in any remaining keys.
func (s *simulator) applyNodeDefaults(node *ir.Node) {
	// Apply per-node scenario overrides for this node.
	s.applyNodeScenario(node.ID)

	if ac, ok := node.Config.(ir.AgentConfig); ok && ac.AutoStatus {
		s.setContextDefaultForNode("outcome", "success", node.ID)
	}
	if tc, ok := node.Config.(ir.ToolConfig); ok {
		s.setContextDefaultForNode("tool_stdout", "success", node.ID)
		s.setContextDefaultForNode("outcome", "success", node.ID)
		s.deriveToolMarker(tc, node.ID)
	}
}

// applyNodeScenario applies per-node scenario values matching "NodeID.key".
// For example, --scenario "Planner.outcome=fail" sets ctx["outcome"]="fail"
// only when processing the Planner node.
// An empty value ("") explicitly clears the key, preventing auto-defaults
// from filling it. This lets unconditional fallback edges fire.
func (s *simulator) applyNodeScenario(nodeID string) {
	prefix := nodeID + "."
	for k, v := range s.opts.Scenario {
		if strings.HasPrefix(k, prefix) {
			key := strings.TrimPrefix(k, prefix)
			if v == "" {
				delete(s.ctx, key)
			} else {
				s.updateContext(key, v)
			}
		}
	}
}

// deriveToolMarker mirrors the runtime: a tool that declares marker_grep has its
// ctx.tool_marker populated from the first stdout line the regex matches. Simulate
// otherwise only sets ctx.tool_stdout, so `on <marker>` edges (which route on
// ctx.tool_marker) never match and silently fall through to the first edge — the
// cause of spurious "infinite loop" test failures on marker-routed workflows. An
// explicit scenario `NodeID.tool_marker=...` wins; otherwise it's derived from
// tool_stdout so a scenario can drive marker routing by supplying realistic output.
func (s *simulator) deriveToolMarker(tc ir.ToolConfig, nodeID string) {
	if tc.MarkerGrep == "" || s.scenarioHasKey("tool_marker", nodeID) {
		return
	}
	if m := firstMarkerMatch(tc.MarkerGrep, s.ctx["tool_stdout"]); m != "" {
		s.updateContext("tool_marker", m)
	}
}

// firstMarkerMatch returns the regex's match against the first matching line of
// stdout, or "" if the pattern is invalid, stdout is empty, or nothing matches.
func firstMarkerMatch(pattern, stdout string) string {
	re, err := regexp.Compile(pattern)
	if err != nil || stdout == "" {
		return ""
	}
	for _, line := range strings.Split(stdout, "\n") {
		if m := re.FindString(line); m != "" {
			return m
		}
	}
	return ""
}

// setContextDefaultForNode sets a context key only if it wasn't injected
// via the --scenario flag (global or per-node). Scenario values always
// take precedence over node defaults. An empty scenario value explicitly
// suppresses the default (see applyNodeScenario).
func (s *simulator) setContextDefaultForNode(key, value, nodeID string) {
	if s.scenarioHasKey(key, nodeID) {
		return
	}
	s.updateContext(key, value)
}

// scenarioHasKey returns true if the scenario provides a value for this key
// (globally or per-node), meaning the auto-default should not be applied.
func (s *simulator) scenarioHasKey(key, nodeID string) bool {
	if _, ok := s.opts.Scenario[key]; ok {
		return true
	}
	_, ok := s.opts.Scenario[nodeID+"."+key]
	return ok
}

// operatorFuncs maps condition operators to their evaluation functions.
var operatorFuncs = map[string]func(ctxVal, value string) bool{
	"=":          func(a, b string) bool { return a == b },
	"==":         func(a, b string) bool { return a == b },
	"!=":         func(a, b string) bool { return a != b },
	"contains":   func(a, b string) bool { return strings.Contains(a, b) },
	"startswith": func(a, b string) bool { return strings.HasPrefix(a, b) },
	"endswith":   func(a, b string) bool { return strings.HasSuffix(a, b) },
}

func (s *simulator) evalCondition(expr ir.ConditionExpr) bool {
	switch e := expr.(type) {
	case ir.CondCompare:
		return s.evalCompare(e)
	case ir.CondAnd:
		return s.evalCondAnd(e)
	case ir.CondOr:
		return s.evalCondOr(e)
	case ir.CondNot:
		return !s.evalCondition(e.Inner)
	}
	return false
}

func (s *simulator) evalCondAnd(e ir.CondAnd) bool {
	return s.evalCondition(e.Left) && s.evalCondition(e.Right)
}

func (s *simulator) evalCondOr(e ir.CondOr) bool {
	return s.evalCondition(e.Left) || s.evalCondition(e.Right)
}

// evalCompare evaluates a single comparison condition.
func (s *simulator) evalCompare(e ir.CondCompare) bool {
	ctxVal := s.resolveVariable(e.Variable)

	if fn, ok := operatorFuncs[e.Op]; ok {
		return fn(ctxVal, e.Value)
	}
	if e.Op == "in" {
		return evalIn(ctxVal, e.Value)
	}
	return false
}

// evalIn checks if ctxVal matches any comma-separated item in value.
func evalIn(ctxVal, value string) bool {
	for _, p := range strings.Split(value, ",") {
		if ctxVal == strings.TrimSpace(p) {
			return true
		}
	}
	return false
}

func (s *simulator) resolveVariable(variable string) string {
	// Variables use namespaced access: "ctx.outcome", "ctx.tool_stdout",
	// "graph.goal", "ctx.internal.loop_restart_count", etc.
	// Strip "ctx." prefix for simple context lookups.
	key := variable
	if strings.HasPrefix(key, "ctx.") {
		key = strings.TrimPrefix(key, "ctx.")
	}

	// Check scenario/context map (bare key first, then full variable name).
	if v, ok := s.ctx[key]; ok {
		return v
	}
	if v, ok := s.ctx[variable]; ok {
		return v
	}

	// Handle graph.* references.
	if variable == "graph.goal" {
		return s.workflow.Goal
	}

	return ""
}
