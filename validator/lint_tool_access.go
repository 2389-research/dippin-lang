package validator

import (
	"fmt"
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

// lintToolAccessValues fires DIP139 when an agent's tool_access is set to
// a value other than "" or "none" (case-insensitive). Authors who skip
// lint and ship an invalid value get fail-closed tracker behavior — DIP139
// surfaces the typo at lint time.
func lintToolAccessValues(w *ir.Workflow) []Diagnostic {
	var diags []Diagnostic
	for _, n := range w.Nodes {
		cfg, ok := n.Config.(ir.AgentConfig)
		if !ok {
			continue
		}
		canonical := strings.ToLower(strings.TrimSpace(cfg.ToolAccess))
		if validToolAccess[canonical] {
			continue
		}
		diags = append(diags, Diagnostic{
			Code:     DIP139,
			Severity: SeverityWarning,
			Message:  fmt.Sprintf("node %q has tool_access %q which is not recognized", n.ID, cfg.ToolAccess),
			Location: n.Source,
			Help:     "valid value: none (omit the field for the full catalog). Invalid values fall back to no-tools at runtime — fix the typo or remove the field.",
		})
	}
	return diags
}
