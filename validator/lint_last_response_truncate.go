package validator

import (
	"fmt"

	"github.com/2389-research/dippin-lang/ir"
)

// lintLastResponseTruncate checks DIP148: last_response_truncate must not be
// negative, on an agent node or a parallel-branch override. 0 / unset means "no
// truncation" (the formatter emits only positive values), so only values < 0 are
// flagged. dippin carries + lints; a runtime enforces the truncation. This does
// NOT interact with DIP147 — a sink carrying last_response_truncate still emits
// the chain-attack Hint (truncation bounds size, not the laundered flow).
func lintLastResponseTruncate(w *ir.Workflow) []Diagnostic {
	var diags []Diagnostic
	for _, n := range w.Nodes {
		switch cfg := n.Config.(type) {
		case ir.AgentConfig:
			if cfg.LastResponseTruncate < 0 {
				diags = append(diags, lastResponseTruncateDiag(
					fmt.Sprintf("agent %q last_response_truncate is %d; cannot be negative", n.ID, cfg.LastResponseTruncate)))
			}
		case ir.ParallelConfig:
			diags = append(diags, lintBranchTruncate(n.ID, cfg.Branches)...)
		}
	}
	return diags
}

// lintBranchTruncate flags negative per-branch last_response_truncate overrides
// for one parallel node.
func lintBranchTruncate(id string, branches []ir.BranchConfig) []Diagnostic {
	var diags []Diagnostic
	for _, b := range branches {
		if b.LastResponseTruncate < 0 {
			diags = append(diags, lastResponseTruncateDiag(
				fmt.Sprintf("parallel %q branch -> %q last_response_truncate is %d; cannot be negative", id, b.Target, b.LastResponseTruncate)))
		}
	}
	return diags
}

func lastResponseTruncateDiag(msg string) Diagnostic {
	return Diagnostic{
		Code:     DIP148,
		Severity: SeverityWarning,
		Message:  msg,
		Help:     "use a non-negative character count (e.g. last_response_truncate: 4096), or omit it / set 0 for no truncation",
	}
}
