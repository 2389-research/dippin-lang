// ABOUTME: DIP154 — an agent sets prompt_prefix/prompt_suffix: none to opt out of
// a defaults prompt cascade, but no cascade of that kind is declared, so the
// opt-out is a no-op (likely a leftover or mistake). Hint severity (#175).
package validator

import (
	"fmt"

	"github.com/2389-research/dippin-lang/ir"
)

// lintPromptOptOut flags no-op cascade opt-outs (DIP154).
func lintPromptOptOut(w *ir.Workflow) []Diagnostic {
	var diags []Diagnostic
	hasPrefix := w.Defaults.PromptPrefix != "" || w.Defaults.PromptPrefixFile != ""
	hasSuffix := w.Defaults.PromptSuffix != "" || w.Defaults.PromptSuffixFile != ""
	for _, n := range w.Nodes {
		cfg, ok := n.Config.(ir.AgentConfig)
		if !ok {
			continue
		}
		diags = appendOptOutDiag(diags, n.ID, "prompt_prefix", cfg.PromptPrefix, hasPrefix, n.Source)
		diags = appendOptOutDiag(diags, n.ID, "prompt_suffix", cfg.PromptSuffix, hasSuffix, n.Source)
	}
	return diags
}

func appendOptOutDiag(diags []Diagnostic, nodeID, field, val string, hasCascade bool, loc ir.SourceLocation) []Diagnostic {
	if val != "none" || hasCascade {
		return diags
	}
	return append(diags, Diagnostic{
		Code:     DIP154,
		Severity: SeverityHint,
		Message: fmt.Sprintf(
			"agent %q sets %s: none but no defaults %s cascade is declared — the opt-out is a no-op",
			nodeID, field, field),
		Location: loc,
	})
}
