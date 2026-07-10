package migrate

import (
	"fmt"
	"strings"

	"github.com/2389-research/dippin-lang/ir"
)

// Difference describes a structural difference between two workflows.
type Difference struct {
	Kind string // "node_missing", "node_extra", "edge_missing", "edge_extra",
	//                "config_mismatch", "kind_mismatch", "start_mismatch",
	//                "exit_mismatch", "defaults_mismatch"
	Message string // Human-readable description
	PathA   string // Location in workflow A (e.g., "node:Validate")
	PathB   string // Location in workflow B (may be empty)
}

// CheckParity compares two workflows for structural equivalence.
// It checks:
//   - Same node IDs and kinds
//   - Same edges (from/to/conditions)
//   - Same start/exit
//   - Compatible node configurations (prompt content modulo whitespace)
//   - A subset of graph-level defaults: model, provider, max_retries,
//     max_restarts, fidelity. compareDefaults does NOT compare the budget,
//     tool, on_failure, retry-policy, or on_resume defaults, so two workflows
//     differing only in those report parity (a known gap — see issue #161).
func CheckParity(a, b *ir.Workflow) []Difference {
	var diffs []Difference
	diffs = append(diffs, compareStartExit(a, b)...)
	diffs = append(diffs, compareNodes(a, b)...)
	diffs = append(diffs, compareEdges(a, b)...)
	diffs = append(diffs, compareDefaults(a.Defaults, b.Defaults)...)
	return diffs
}

// compareStartExit checks the start and exit fields of two workflows.
func compareStartExit(a, b *ir.Workflow) []Difference {
	var diffs []Difference
	if a.Start != b.Start {
		diffs = append(diffs, Difference{
			Kind:    "start_mismatch",
			Message: fmt.Sprintf("start differs: %q vs %q", a.Start, b.Start),
			PathA:   "workflow.start",
			PathB:   "workflow.start",
		})
	}
	if a.Exit != b.Exit {
		diffs = append(diffs, Difference{
			Kind:    "exit_mismatch",
			Message: fmt.Sprintf("exit differs: %q vs %q", a.Exit, b.Exit),
			PathA:   "workflow.exit",
			PathB:   "workflow.exit",
		})
	}
	return diffs
}

// compareNodes checks nodes between two workflows for missing, extra, and mismatched.
func compareNodes(a, b *ir.Workflow) []Difference {
	aNodes := buildNodeMap(a.Nodes)
	bNodes := buildNodeMap(b.Nodes)

	var diffs []Difference
	diffs = append(diffs, diffNodesAvsB(aNodes, bNodes)...)
	diffs = append(diffs, findExtraNodes(aNodes, bNodes)...)
	return diffs
}

// buildNodeMap creates a map from node ID to node.
func buildNodeMap(nodes []*ir.Node) map[string]*ir.Node {
	m := make(map[string]*ir.Node, len(nodes))
	for _, n := range nodes {
		m[n.ID] = n
	}
	return m
}

// diffNodesAvsB finds missing nodes and config/kind mismatches from A's perspective.
func diffNodesAvsB(aNodes, bNodes map[string]*ir.Node) []Difference {
	var diffs []Difference
	for id, na := range aNodes {
		nb, ok := bNodes[id]
		if !ok {
			diffs = append(diffs, Difference{
				Kind:    "node_missing",
				Message: fmt.Sprintf("node %q present in A but missing from B", id),
				PathA:   "node:" + id,
			})
			continue
		}
		if na.Kind != nb.Kind {
			diffs = append(diffs, Difference{
				Kind:    "kind_mismatch",
				Message: fmt.Sprintf("node %q kind: %q vs %q", id, na.Kind, nb.Kind),
				PathA:   "node:" + id,
				PathB:   "node:" + id,
			})
		}
		diffs = append(diffs, compareConfigs(id, na, nb)...)
	}
	return diffs
}

// findExtraNodes finds nodes in B that are not in A.
func findExtraNodes(aNodes, bNodes map[string]*ir.Node) []Difference {
	var diffs []Difference
	for id := range bNodes {
		if _, ok := aNodes[id]; !ok {
			diffs = append(diffs, Difference{
				Kind:    "node_extra",
				Message: fmt.Sprintf("node %q present in B but missing from A", id),
				PathB:   "node:" + id,
			})
		}
	}
	return diffs
}

// compareEdges checks edges between two workflows for missing and extra.
func compareEdges(a, b *ir.Workflow) []Difference {
	aEdges := effectiveEdgeSet(a)
	bEdges := effectiveEdgeSet(b)

	var diffs []Difference
	for key := range aEdges {
		if _, ok := bEdges[key]; !ok {
			diffs = append(diffs, Difference{
				Kind:    "edge_missing",
				Message: fmt.Sprintf("edge %s present in A but missing from B", key),
				PathA:   "edge:" + key,
			})
		}
	}
	for key := range bEdges {
		if _, ok := aEdges[key]; !ok {
			diffs = append(diffs, Difference{
				Kind:    "edge_extra",
				Message: fmt.Sprintf("edge %s present in B but missing from A", key),
				PathB:   "edge:" + key,
			})
		}
	}
	return diffs
}

// edgeKey produces a canonical string key for an edge including condition.
func edgeKey(e *ir.Edge) string {
	condStr := ""
	if e.Condition != nil {
		condStr = e.Condition.Raw
	}
	return fmt.Sprintf("%s->%s[%s]", e.From, e.To, condStr)
}

// effectiveEdgeSet keys a workflow's effective edges — explicit edges plus the
// implicit parallel fan-out / fan-in edges ir.EdgesFrom synthesizes from node
// config. Comparing effective edges keeps migration parity correct when a fork
// is declared inline (parallel/fan_in) without an edges-block re-declaration
// (#136), matching a DOT that draws it explicitly.
func effectiveEdgeSet(w *ir.Workflow) map[string]*ir.Edge {
	m := make(map[string]*ir.Edge, len(w.Edges))
	for _, n := range w.Nodes {
		for _, e := range w.EdgesFrom(n.ID) {
			m[edgeKey(e)] = e
		}
	}
	return m
}

// configMismatchDiff is a helper for config type mismatch differences.
func configMismatchDiff(id, path string, aType string, bConfig interface{}) Difference {
	return Difference{
		Kind:    "config_mismatch",
		Message: fmt.Sprintf("node %q config type mismatch: %s vs %T", id, aType, bConfig),
		PathA:   path,
		PathB:   path,
	}
}

// fieldDiff creates a config_mismatch Difference for a specific field.
func fieldDiff(id, field, msg string) Difference {
	path := "node:" + id + "." + field
	return Difference{
		Kind:    "config_mismatch",
		Message: msg,
		PathA:   path,
		PathB:   path,
	}
}

// compareConfigs compares the configurations of two nodes with the same ID.
func compareConfigs(id string, a, b *ir.Node) []Difference {
	path := "node:" + id
	if diffs, handled := comparePrimaryConfigs(id, path, a.Config, b.Config); handled {
		return diffs
	}
	return compareStructuralConfigs(id, path, a.Config, b.Config)
}

// comparePrimaryConfigs handles Agent, Tool, and Human config comparisons.
func comparePrimaryConfigs(id, path string, aCfg, bCfg ir.NodeConfig) ([]Difference, bool) {
	switch ac := aCfg.(type) {
	case ir.AgentConfig:
		return compareAgentConfigs(id, path, ac, bCfg), true
	case ir.ToolConfig:
		return compareToolConfigs(id, path, ac, bCfg), true
	case ir.HumanConfig:
		return compareHumanConfigs(id, path, ac, bCfg), true
	}
	return nil, false
}

// compareStructuralConfigs handles Parallel, FanIn, and Subgraph config comparisons.
func compareStructuralConfigs(id, path string, aCfg, bCfg ir.NodeConfig) []Difference {
	switch ac := aCfg.(type) {
	case ir.ParallelConfig:
		return compareParallelConfigs(id, path, ac, bCfg)
	case ir.FanInConfig:
		return compareFanInConfigs(id, path, ac, bCfg)
	case ir.SubgraphConfig:
		return compareSubgraphConfigs(id, path, ac, bCfg)
	}
	return nil
}

func compareAgentConfigs(id, path string, ac ir.AgentConfig, bCfg interface{}) []Difference {
	bc, ok := bCfg.(ir.AgentConfig)
	if !ok {
		return []Difference{configMismatchDiff(id, path, "AgentConfig", bCfg)}
	}
	var diffs []Difference
	diffs = append(diffs, compareAgentPromptFields(id, ac, bc)...)
	diffs = append(diffs, compareAgentModelProvider(id, ac, bc)...)
	diffs = append(diffs, compareAgentBehavior(id, ac, bc)...)
	return diffs
}

// compareAgentPromptFields compares prompt and system_prompt fields including their file sources.
func compareAgentPromptFields(id string, ac, bc ir.AgentConfig) []Difference {
	var diffs []Difference
	if !promptsEqual(ac.Prompt, bc.Prompt) {
		diffs = append(diffs, fieldDiff(id, "prompt", fmt.Sprintf("node %q prompt differs", id)))
	}
	if ac.PromptFile != bc.PromptFile {
		diffs = append(diffs, fieldDiff(id, "prompt_file", fmt.Sprintf("node %q prompt_file: %q vs %q", id, ac.PromptFile, bc.PromptFile)))
	}
	if !promptsEqual(ac.SystemPrompt, bc.SystemPrompt) {
		diffs = append(diffs, fieldDiff(id, "system_prompt", fmt.Sprintf("node %q system_prompt differs", id)))
	}
	if ac.SystemPromptFile != bc.SystemPromptFile {
		diffs = append(diffs, fieldDiff(id, "system_prompt_file", fmt.Sprintf("node %q system_prompt_file: %q vs %q", id, ac.SystemPromptFile, bc.SystemPromptFile)))
	}
	return diffs
}

// compareAgentModelProvider compares model and provider fields.
func compareAgentModelProvider(id string, ac, bc ir.AgentConfig) []Difference {
	var diffs []Difference
	if ac.Model != bc.Model {
		diffs = append(diffs, fieldDiff(id, "model", fmt.Sprintf("node %q model: %q vs %q", id, ac.Model, bc.Model)))
	}
	if ac.Provider != bc.Provider {
		diffs = append(diffs, fieldDiff(id, "provider", fmt.Sprintf("node %q provider: %q vs %q", id, ac.Provider, bc.Provider)))
	}
	return diffs
}

// compareAgentBehavior compares goal_gate, auto_status, tool_access, and
// per-agent limits (writable_paths, last_response_truncate).
func compareAgentBehavior(id string, ac, bc ir.AgentConfig) []Difference {
	var diffs []Difference
	if ac.GoalGate != bc.GoalGate {
		diffs = append(diffs, fieldDiff(id, "goal_gate", fmt.Sprintf("node %q goal_gate: %v vs %v", id, ac.GoalGate, bc.GoalGate)))
	}
	if ac.AutoStatus != bc.AutoStatus {
		diffs = append(diffs, fieldDiff(id, "auto_status", fmt.Sprintf("node %q auto_status: %v vs %v", id, ac.AutoStatus, bc.AutoStatus)))
	}
	if ac.ToolAccess != bc.ToolAccess {
		diffs = append(diffs, fieldDiff(id, "tool_access", fmt.Sprintf("node %q tool_access: %q vs %q", id, ac.ToolAccess, bc.ToolAccess)))
	}
	diffs = append(diffs, compareAgentLimits(id, ac, bc)...)
	return diffs
}

// compareAgentLimits compares writable_paths and last_response_truncate.
func compareAgentLimits(id string, ac, bc ir.AgentConfig) []Difference {
	var diffs []Difference
	if strings.Join(ac.WritablePaths, ",") != strings.Join(bc.WritablePaths, ",") {
		diffs = append(diffs, fieldDiff(id, "writable_paths", fmt.Sprintf("node %q writable_paths: %v vs %v", id, ac.WritablePaths, bc.WritablePaths)))
	}
	if ac.LastResponseTruncate != bc.LastResponseTruncate {
		diffs = append(diffs, fieldDiff(id, "last_response_truncate", fmt.Sprintf("node %q last_response_truncate: %d vs %d", id, ac.LastResponseTruncate, bc.LastResponseTruncate)))
	}
	return diffs
}

func compareToolConfigs(id, path string, ac ir.ToolConfig, bCfg interface{}) []Difference {
	bc, ok := bCfg.(ir.ToolConfig)
	if !ok {
		return []Difference{configMismatchDiff(id, path, "ToolConfig", bCfg)}
	}
	var diffs []Difference
	diffs = append(diffs, compareToolScalars(id, ac, bc)...)
	diffs = append(diffs, compareToolSlices(id, ac, bc)...)
	return diffs
}

func compareToolScalars(id string, ac, bc ir.ToolConfig) []Difference {
	var diffs []Difference
	diffs = append(diffs, compareToolCommandAndTimeout(id, ac, bc)...)
	diffs = append(diffs, compareToolMarkerAndRoute(id, ac, bc)...)
	diffs = append(diffs, compareToolOutputLimit(id, ac, bc)...)
	return diffs
}

func compareToolCommandAndTimeout(id string, ac, bc ir.ToolConfig) []Difference {
	var diffs []Difference
	if !promptsEqual(ac.Command, bc.Command) {
		diffs = append(diffs, fieldDiff(id, "command", fmt.Sprintf("node %q command differs", id)))
	}
	if ac.CommandFile != bc.CommandFile {
		diffs = append(diffs, fieldDiff(id, "command_file", fmt.Sprintf("node %q command_file: %q vs %q", id, ac.CommandFile, bc.CommandFile)))
	}
	if ac.Timeout != bc.Timeout {
		diffs = append(diffs, fieldDiff(id, "timeout", fmt.Sprintf("node %q timeout: %s vs %s", id, ac.Timeout, bc.Timeout)))
	}
	return diffs
}

func compareToolMarkerAndRoute(id string, ac, bc ir.ToolConfig) []Difference {
	var diffs []Difference
	if ac.MarkerGrep != bc.MarkerGrep {
		diffs = append(diffs, fieldDiff(id, "marker_grep", fmt.Sprintf("node %q marker_grep: %q vs %q", id, ac.MarkerGrep, bc.MarkerGrep)))
	}
	if ac.RouteRequired != bc.RouteRequired {
		diffs = append(diffs, fieldDiff(id, "route_required", fmt.Sprintf("node %q route_required: %v vs %v", id, ac.RouteRequired, bc.RouteRequired)))
	}
	return diffs
}

func compareToolOutputLimit(id string, ac, bc ir.ToolConfig) []Difference {
	var diffs []Difference
	if ac.OutputLimit != bc.OutputLimit {
		diffs = append(diffs, fieldDiff(id, "output_limit", fmt.Sprintf("node %q output_limit: %d vs %d", id, ac.OutputLimit, bc.OutputLimit)))
	}
	return diffs
}

func compareToolSlices(id string, ac, bc ir.ToolConfig) []Difference {
	var diffs []Difference
	if strings.Join(ac.Outputs, ",") != strings.Join(bc.Outputs, ",") {
		diffs = append(diffs, fieldDiff(id, "outputs", fmt.Sprintf("node %q outputs: %v vs %v", id, ac.Outputs, bc.Outputs)))
	}
	return diffs
}

func compareHumanConfigs(id, path string, ac ir.HumanConfig, bCfg interface{}) []Difference {
	bc, ok := bCfg.(ir.HumanConfig)
	if !ok {
		return []Difference{configMismatchDiff(id, path, "HumanConfig", bCfg)}
	}
	if ac.Mode != bc.Mode {
		return []Difference{fieldDiff(id, "mode", fmt.Sprintf("node %q mode: %q vs %q", id, ac.Mode, bc.Mode))}
	}
	return nil
}

func compareParallelConfigs(id, path string, ac ir.ParallelConfig, bCfg interface{}) []Difference {
	bc, ok := bCfg.(ir.ParallelConfig)
	if !ok {
		return []Difference{configMismatchDiff(id, path, "ParallelConfig", bCfg)}
	}
	var diffs []Difference
	diffs = append(diffs, compareParallelTargets(id, ac, bc)...)
	diffs = append(diffs, compareParallelBranches(id, ac, bc)...)
	return diffs
}

func compareParallelTargets(id string, ac, bc ir.ParallelConfig) []Difference {
	if strings.Join(ac.Targets, ",") != strings.Join(bc.Targets, ",") {
		return []Difference{fieldDiff(id, "targets", fmt.Sprintf("node %q targets: %v vs %v", id, ac.Targets, bc.Targets))}
	}
	return nil
}

// compareParallelBranches compares branch slices position-by-position. Order is
// significant (it maps to targets). BranchConfig now carries a slice field
// (WritablePaths), so it is no longer comparable with ==; compare field-by-field.
func compareParallelBranches(id string, ac, bc ir.ParallelConfig) []Difference {
	if len(ac.Branches) != len(bc.Branches) {
		return []Difference{fieldDiff(id, "branches", fmt.Sprintf("node %q branches differ", id))}
	}
	for i := range ac.Branches {
		if !branchesEqual(ac.Branches[i], bc.Branches[i]) {
			return []Difference{fieldDiff(id, "branches", fmt.Sprintf("node %q branches differ", id))}
		}
	}
	return nil
}

// branchesEqual reports whether two branch configs are field-equal (scalar
// fields plus the comma-joined WritablePaths slice).
func branchesEqual(a, b ir.BranchConfig) bool {
	return branchScalarsEqual(a, b) &&
		strings.Join(a.WritablePaths, ",") == strings.Join(b.WritablePaths, ",")
}

func branchScalarsEqual(a, b ir.BranchConfig) bool {
	return a.Target == b.Target && a.Model == b.Model &&
		a.Provider == b.Provider && branchLimitsEqual(a, b)
}

// branchLimitsEqual reports whether the fidelity, tool_access, and
// last_response_truncate fields of two branch configs match.
func branchLimitsEqual(a, b ir.BranchConfig) bool {
	return a.Fidelity == b.Fidelity && a.ToolAccess == b.ToolAccess &&
		a.LastResponseTruncate == b.LastResponseTruncate
}

func compareFanInConfigs(id, path string, ac ir.FanInConfig, bCfg interface{}) []Difference {
	bc, ok := bCfg.(ir.FanInConfig)
	if !ok {
		return []Difference{configMismatchDiff(id, path, "FanInConfig", bCfg)}
	}
	if strings.Join(ac.Sources, ",") != strings.Join(bc.Sources, ",") {
		return []Difference{fieldDiff(id, "sources", fmt.Sprintf("node %q sources: %v vs %v", id, ac.Sources, bc.Sources))}
	}
	return nil
}

func compareSubgraphConfigs(id, path string, ac ir.SubgraphConfig, bCfg interface{}) []Difference {
	bc, ok := bCfg.(ir.SubgraphConfig)
	if !ok {
		return []Difference{configMismatchDiff(id, path, "SubgraphConfig", bCfg)}
	}
	if ac.Ref != bc.Ref {
		return []Difference{fieldDiff(id, "ref", fmt.Sprintf("node %q ref: %q vs %q", id, ac.Ref, bc.Ref))}
	}
	return nil
}

// promptsEqual compares two strings with whitespace tolerance.
// Prompts that differ only in trailing whitespace per line and leading/trailing
// whitespace overall are considered equal.
func promptsEqual(a, b string) bool {
	return normalizeWhitespace(a) == normalizeWhitespace(b)
}

// compareDefaults reports differences between workflow defaults.
func compareDefaults(a, b ir.WorkflowDefaults) []Difference {
	var diffs []Difference
	diffs = append(diffs, compareDefaultField(a.Model, b.Model, "defaults.model")...)
	diffs = append(diffs, compareDefaultField(a.Provider, b.Provider, "defaults.provider")...)
	diffs = append(diffs, compareDefaultIntField(a.MaxRetries, b.MaxRetries, "defaults.max_retries")...)
	diffs = append(diffs, compareDefaultIntField(a.MaxRestarts, b.MaxRestarts, "defaults.max_restarts")...)
	diffs = append(diffs, compareDefaultField(a.Fidelity, b.Fidelity, "defaults.fidelity")...)
	return diffs
}

// compareDefaultField compares a single string field in workflow defaults.
func compareDefaultField(a, b string, field string) []Difference {
	if a == b {
		return nil
	}
	fmtStr := field + ": %q vs %q"
	return []Difference{{
		Kind:    "defaults_mismatch",
		Message: fmt.Sprintf(fmtStr, a, b),
		PathA:   field,
		PathB:   field,
	}}
}

// compareDefaultIntField compares a single int field in workflow defaults.
func compareDefaultIntField(a, b int, field string) []Difference {
	if a == b {
		return nil
	}
	return []Difference{{
		Kind:    "defaults_mismatch",
		Message: fmt.Sprintf("%s: %d vs %d", field, a, b),
		PathA:   field,
		PathB:   field,
	}}
}
