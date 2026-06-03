package validator

import (
	"fmt"
	"sort"
	"strings"

	"github.com/2389-research/dippin-lang/ir"
)

// validToolAccess lists recognized values for AgentConfig.ToolAccess after
// case-insensitive trim. Empty = inherit default (full catalog); "none" =
// disable LLM tools. v1 ships only these two values; middle tiers are
// tracked as a follow-up issue.
var validToolAccess = map[string]bool{
	"":     true,
	"none": true,
}

// lintToolAccessValues fires DIP139 when an agent node — or a per-branch override
// on a parallel node — sets tool_access to a value other than "" or "none"
// (case-insensitive). Authors who skip lint and ship an invalid value get
// fail-closed runtime behavior; DIP139 surfaces the typo at lint time.
func lintToolAccessValues(w *ir.Workflow) []Diagnostic {
	var diags []Diagnostic
	for _, n := range w.Nodes {
		diags = append(diags, checkNodeToolAccessByKind(n)...)
	}
	return diags
}

// checkNodeToolAccessByKind checks tool_access for a single node by config type.
func checkNodeToolAccessByKind(n *ir.Node) []Diagnostic {
	switch cfg := n.Config.(type) {
	case ir.AgentConfig:
		return checkToolAccessValue(n, cfg.ToolAccess, "")
	case ir.ParallelConfig:
		return checkBranchToolAccess(n, cfg.Branches)
	default:
		return nil
	}
}

// checkBranchToolAccess checks tool_access on each branch of a parallel node.
func checkBranchToolAccess(n *ir.Node, branches []ir.BranchConfig) []Diagnostic {
	var diags []Diagnostic
	for _, b := range branches {
		diags = append(diags, checkToolAccessValue(n, b.ToolAccess, b.Target)...)
	}
	return diags
}

// checkToolAccessValue validates a single tool_access string, returning a
// DIP139 diagnostic if unrecognized. A non-empty branch arg produces a
// branch-qualified message.
func checkToolAccessValue(n *ir.Node, toolAccess, branch string) []Diagnostic {
	canonical := strings.ToLower(strings.TrimSpace(toolAccess))
	if validToolAccess[canonical] {
		return nil
	}
	msg := fmt.Sprintf("node %q has tool_access %q which is not recognized", n.ID, toolAccess)
	// Omission semantics differ: an agent omitting tool_access gets the full
	// catalog; a branch omitting it inherits the target agent's setting (never
	// re-grants the full catalog). Tailor the fix hint so it isn't misleading.
	help := "valid value: none (omit the field for the full catalog). Invalid values fall back to no-tools at runtime — fix the typo or remove the field."
	if branch != "" {
		msg = fmt.Sprintf("node %q branch %q has tool_access %q which is not recognized", n.ID, branch, toolAccess)
		help = "valid value: none (omit to inherit the target agent's tool_access). Invalid values fall back to no-tools at runtime — fix the typo or remove the field."
	}
	return []Diagnostic{{
		Code:     DIP139,
		Severity: SeverityWarning,
		Message:  msg,
		Location: n.Source,
		Help:     help,
	}}
}

// toolReenablingParamsKeys lists Params keys that re-grant LLM tools. When an
// agent sets tool_access, the runtime strips these (fail-closed); the lint surfaces
// the neutralized override at author time. Source: v0.32.0 issue-41 design
// § Params bypass defense.
var toolReenablingParamsKeys = map[string]bool{
	"allowed_tools":    true,
	"disallowed_tools": true,
	"tool_choice":      true,
	"permission_mode":  true,
}

// NOTE: DIP140 is agent-only because BranchConfig has no Params field today.
// If BranchConfig.Params is ever added, extend this scan to branches — the
// tool-re-enabling bypass would reopen at the branch level.

// lintParamsReenablesTools fires DIP140 when an agent declares tool_access
// (non-empty) yet also sets a Params key that would re-enable tools. The runtime
// ignores such keys when tool_access is set, so the override is silently
// neutralized — a likely bypass attempt or dead config.
func lintParamsReenablesTools(w *ir.Workflow) []Diagnostic {
	var diags []Diagnostic
	for _, n := range w.Nodes {
		cfg, ok := n.Config.(ir.AgentConfig)
		if !ok || strings.TrimSpace(cfg.ToolAccess) == "" || len(cfg.Params) == 0 {
			continue
		}
		diags = append(diags, checkParamsReenable(n, cfg.Params)...)
	}
	return diags
}

func checkParamsReenable(n *ir.Node, params map[string]string) []Diagnostic {
	keys := make([]string, 0, len(params))
	for k := range params {
		if toolReenablingParamsKeys[k] {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)
	var diags []Diagnostic
	for _, k := range keys {
		diags = append(diags, Diagnostic{
			Code:     DIP140,
			Severity: SeverityWarning,
			Message:  fmt.Sprintf("node %q params key %q re-enables tools that tool_access strips — the runtime ignores it", n.ID, k),
			Location: n.Source,
			Help:     "remove the params key; tool_access governs the tool catalog. To grant tools instead, omit tool_access.",
		})
	}
	return diags
}
