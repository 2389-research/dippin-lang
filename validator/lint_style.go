package validator

import (
	"fmt"
	"strings"

	"github.com/2389-research/dippin-lang/ir"
)

// lintNamespaceCollisions checks DIP109: multiple subgraph nodes referencing
// the same file without different parameters could cause namespace collisions
// when expanded.
func lintNamespaceCollisions(w *ir.Workflow) []Diagnostic {
	var diags []Diagnostic

	seen := make(map[string]*ir.Node)
	for _, n := range w.Nodes {
		cfg, ok := n.Config.(ir.SubgraphConfig)
		if !ok {
			continue
		}
		if first, exists := seen[cfg.Ref]; exists {
			diags = append(diags, Diagnostic{
				Code:     DIP109,
				Severity: SeverityWarning,
				Message:  fmt.Sprintf("nodes %q and %q both reference subgraph %q, which may cause namespace collisions", first.ID, n.ID, cfg.Ref),
				Location: n.Source,
				Help:     "use distinct node IDs and ensure imported names do not collide after expansion",
			})
		} else {
			seen[cfg.Ref] = n
		}
	}
	return diags
}

// lintEmptyPrompts checks DIP110: agent nodes should have a non-empty prompt.
// An agent without a prompt has nothing to send to the LLM.
func lintEmptyPrompts(w *ir.Workflow) []Diagnostic {
	var diags []Diagnostic
	for _, n := range w.Nodes {
		if d, ok := checkEmptyPrompt(n, w); ok {
			diags = append(diags, d)
		}
	}
	return diags
}

// hasAuthoredPrompt returns true if the agent config has a prompt source (inline or file).
func hasAuthoredPrompt(cfg ir.AgentConfig) bool {
	return cfg.PromptFile != "" || strings.TrimSpace(cfg.Prompt) != ""
}

// checkEmptyPrompt checks a single node for DIP110.
func checkEmptyPrompt(n *ir.Node, w *ir.Workflow) (Diagnostic, bool) {
	if n.ID == w.Start || n.ID == w.Exit {
		return Diagnostic{}, false
	}
	cfg, ok := n.Config.(ir.AgentConfig)
	if !ok {
		return Diagnostic{}, false
	}
	if !hasAuthoredPrompt(cfg) {
		return Diagnostic{
			Code:     DIP110,
			Severity: SeverityWarning,
			Message:  fmt.Sprintf("agent node %q has an empty prompt", n.ID),
			Location: n.Source,
			Help:     "add a prompt: field with instructions for the LLM",
		}, true
	}
	return Diagnostic{}, false
}

// lintToolTimeout checks DIP111: tool nodes should have a timeout configured.
// Without a timeout, a hanging tool command could block the entire pipeline.
func lintToolTimeout(w *ir.Workflow) []Diagnostic {
	var diags []Diagnostic
	for _, n := range w.Nodes {
		cfg, ok := n.Config.(ir.ToolConfig)
		if !ok {
			continue
		}
		if cfg.Timeout == 0 {
			diags = append(diags, Diagnostic{
				Code:     DIP111,
				Severity: SeverityWarning,
				Message:  fmt.Sprintf("tool node %q has no timeout configured", n.ID),
				Location: n.Source,
				Help:     "add a timeout (e.g., timeout: 60s) to prevent hanging commands",
			})
		}
	}
	return diags
}

// validFidelityLevels is the set of fidelity levels recognized by Tracker's engine.
var validFidelityLevels = map[string]bool{
	"full":           true,
	"summary:high":   true,
	"summary:medium": true,
	"summary:low":    true,
	"compact":        true,
	"truncate":       true,
}

// lintFidelity checks DIP114: fidelity must be one of the known levels.
func lintFidelity(w *ir.Workflow) []Diagnostic {
	var diags []Diagnostic
	diags = append(diags, checkDefaultFidelity(w)...)
	diags = append(diags, checkNodeFidelity(w)...)
	return diags
}

// checkDefaultFidelity checks the workflow-level default fidelity.
func checkDefaultFidelity(w *ir.Workflow) []Diagnostic {
	if f := w.Defaults.Fidelity; f != "" && !validFidelityLevels[f] {
		return []Diagnostic{{
			Code:     DIP114,
			Severity: SeverityWarning,
			Message:  fmt.Sprintf("workflow default fidelity %q is not a recognized level", f),
			Help:     "valid levels: full, summary:high, summary:medium, summary:low, compact, truncate",
		}}
	}
	return nil
}

// checkNodeFidelity checks per-node fidelity levels on agent and parallel branch configs.
func checkNodeFidelity(w *ir.Workflow) []Diagnostic {
	var diags []Diagnostic
	for _, n := range w.Nodes {
		diags = append(diags, checkNodeFidelityByKind(n)...)
	}
	return diags
}

// checkNodeFidelityByKind checks fidelity for a single node based on its config type.
func checkNodeFidelityByKind(n *ir.Node) []Diagnostic {
	switch cfg := n.Config.(type) {
	case ir.AgentConfig:
		return checkFidelityValue(n, cfg.Fidelity, "")
	case ir.ParallelConfig:
		return checkBranchFidelities(n, cfg.Branches)
	default:
		return nil
	}
}

// checkBranchFidelities checks fidelity on each branch of a parallel node.
func checkBranchFidelities(n *ir.Node, branches []ir.BranchConfig) []Diagnostic {
	var diags []Diagnostic
	for _, b := range branches {
		diags = append(diags, checkFidelityValue(n, b.Fidelity, b.Target)...)
	}
	return diags
}

// checkFidelityValue validates a single fidelity string, returning a diagnostic if invalid.
func checkFidelityValue(n *ir.Node, fidelity, branch string) []Diagnostic {
	if fidelity == "" || validFidelityLevels[fidelity] {
		return nil
	}
	msg := fmt.Sprintf("node %q has fidelity %q which is not a recognized level", n.ID, fidelity)
	if branch != "" {
		msg = fmt.Sprintf("node %q branch %q has fidelity %q which is not a recognized level", n.ID, branch, fidelity)
	}
	return []Diagnostic{{
		Code:     DIP114,
		Severity: SeverityWarning,
		Message:  msg,
		Location: n.Source,
		Help:     "valid levels: full, summary:high, summary:medium, summary:low, compact, truncate",
	}}
}

// lintCompactionThreshold checks DIP116: compaction_threshold must be in [0.0, 1.0].
func lintCompactionThreshold(w *ir.Workflow) []Diagnostic {
	var diags []Diagnostic
	for _, n := range w.Nodes {
		cfg, ok := n.Config.(ir.AgentConfig)
		if !ok {
			continue
		}
		if isInvalidThreshold(cfg.CompactionThreshold) {
			diags = append(diags, Diagnostic{
				Code:     DIP116,
				Severity: SeverityWarning,
				Message:  fmt.Sprintf("node %q has compaction_threshold %.2f outside valid range [0.0, 1.0]", n.ID, cfg.CompactionThreshold),
				Location: n.Source,
				Help:     "compaction_threshold should be between 0.0 and 1.0",
			})
		}
	}
	return diags
}

// isInvalidThreshold returns true if threshold is set and outside [0, 1].
func isInvalidThreshold(t float64) bool {
	return t != 0 && (t < 0 || t > 1)
}

// validOnResumeValues is the set of valid on_resume values.
var validOnResumeValues = map[string]bool{
	"preserve": true,
	"degrade":  true,
}

// lintOnResume checks DIP116: on_resume must be "preserve" or "degrade",
// and should not be set without fidelity.
func lintOnResume(w *ir.Workflow) []Diagnostic {
	var diags []Diagnostic
	or := w.Defaults.OnResume
	if or == "" {
		return nil
	}
	if !validOnResumeValues[or] {
		diags = append(diags, Diagnostic{
			Code:     DIP116,
			Severity: SeverityWarning,
			Message:  fmt.Sprintf("workflow default on_resume %q is not a recognized value", or),
			Help:     "valid values: preserve, degrade",
		})
	}
	if w.Defaults.Fidelity == "" {
		diags = append(diags, Diagnostic{
			Code:     DIP116,
			Severity: SeverityWarning,
			Message:  "on_resume is set but fidelity is not configured",
			Help:     "set fidelity before configuring on_resume behavior",
		})
	}
	return diags
}

// lintStylesheetRefs checks DIP117/DIP118: stylesheet selectors must reference
// existing classes and node IDs.
func lintStylesheetRefs(w *ir.Workflow) []Diagnostic {
	var diags []Diagnostic
	classes, nodeIDs := collectClassesAndIDs(w)
	for _, rule := range w.Stylesheet {
		diags = append(diags, checkSelectorRef(rule, classes, nodeIDs)...)
	}
	return diags
}

// checkSelectorRef validates a single stylesheet rule's selector.
func checkSelectorRef(rule ir.StylesheetRule, classes, nodeIDs map[string]bool) []Diagnostic {
	switch rule.Selector.Kind {
	case "class":
		if !classes[rule.Selector.Value] {
			return []Diagnostic{{
				Code:     DIP117,
				Severity: SeverityWarning,
				Message:  fmt.Sprintf("stylesheet references class %q which is not declared on any node", rule.Selector.Value),
				Help:     "add class: " + rule.Selector.Value + " to a node declaration",
			}}
		}
	case "id":
		if !nodeIDs[rule.Selector.Value] {
			return []Diagnostic{{
				Code:     DIP118,
				Severity: SeverityWarning,
				Message:  fmt.Sprintf("stylesheet references node ID %q which does not exist", rule.Selector.Value),
				Help:     "check the node ID spelling or add the node",
			}}
		}
	}
	return nil
}

// collectClassesAndIDs builds sets of all declared classes and node IDs.
func collectClassesAndIDs(w *ir.Workflow) (map[string]bool, map[string]bool) {
	classes := make(map[string]bool)
	nodeIDs := make(map[string]bool)
	for _, n := range w.Nodes {
		nodeIDs[n.ID] = true
		for _, c := range n.Classes {
			classes[c] = true
		}
	}
	return classes, nodeIDs
}
