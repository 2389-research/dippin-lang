package validator

import (
	"fmt"
	"strings"

	"github.com/2389-research/dippin-lang/ir"
)

// lintChainAttack emits DIP147 (Hint): a restricted (tool_access: none) agent
// declares a context key in writes: that a downstream tool-bearing
// (tool_access: full) agent declares in reads:, laundering whatever the
// restricted agent processed into a privileged context.
//
// tool_access bounds an agent's TOOLS, not its INFORMATION FLOW. An agent marked
// tool_access: none is the author's declaration that its inputs/outputs are
// untrusted; when that output reaches an agent that still holds the full tool
// catalog, an injection payload the restricted agent summarized can drive the
// privileged agent's tools. This is the chain-attack gap left open by the #41
// tool_access arc (issue #56; see also the DIP143/DIP146 "not information flow"
// notes).
//
// Scope (v1): the EXPLICIT-KEY vector only — an author-declared writes:/reads:
// dependency, multi-hop via forward-edge reachability. The implicit
// ${ctx.last_response} auto-injection edge (a bare none -> full edge with no
// declared key) is deliberately NOT flagged here: that topology is issue #57
// (cross-node edge warning, closed/deferred), and its mitigation is #56's
// last_response_truncate: attribute (a follow-up). Firing only on an explicit
// key dependency keeps DIP147 precise — it flags a flow the author wired by hand,
// not mere graph adjacency.
//
// Detection only — dippin flags the topology; the runtime enforces info-flow
// control (key sanitization / context bound). Cross-file chains and
// parallel-branch / fan_in / manager_loop vectors are out of scope (follow-ups).
// Wasm-safe: ir.Workflow + graph only.
func lintChainAttack(w *ir.Workflow) []Diagnostic {
	if w.Start == "" || w.Node(w.Start) == nil {
		return nil
	}
	adj := buildForwardAdjacency(w)
	_, upstream := computeAvailableAndUpstream(w, adj)

	var diags []Diagnostic
	for _, sink := range w.Nodes {
		diags = append(diags, chainAttacksIntoSink(w, sink, upstream[sink.ID])...)
	}
	return diags
}

// chainAttacksIntoSink emits DIP147 for every restricted agent upstream of sink
// that writes a context key sink reads. Returns nil unless sink is a tool-bearing
// agent that declares reads:.
func chainAttacksIntoSink(w *ir.Workflow, sink *ir.Node, upstreamIDs map[string]bool) []Diagnostic {
	if !isToolBearingAgent(sink) {
		return nil
	}
	reads := readSet(sink)
	if len(reads) == 0 {
		return nil
	}
	var diags []Diagnostic
	for srcID := range upstreamIDs {
		diags = append(diags, chainAttacksFromSource(srcID, w.Node(srcID), reads, sink)...)
	}
	return diags
}

// chainAttacksFromSource emits one DIP147 per context key that src (when it is a
// restricted agent) declares in writes: and sink declares in reads: — a distinct
// laundered (source, key, sink) flow each.
func chainAttacksFromSource(srcID string, src *ir.Node, reads map[string]bool, sink *ir.Node) []Diagnostic {
	if !isRestrictedAgent(src) {
		return nil
	}
	var diags []Diagnostic
	for _, key := range src.IO.Writes {
		if reads[key] {
			diags = append(diags, chainAttackDiagnostic(srcID, key, sink))
		}
	}
	return diags
}

// chainAttackDiagnostic builds the DIP147 hint located at the privileged sink.
func chainAttackDiagnostic(srcID, key string, sink *ir.Node) Diagnostic {
	return Diagnostic{
		Code:     DIP147,
		Severity: SeverityHint,
		Message: fmt.Sprintf(
			"restricted agent %q (tool_access: none) writes context key %q that tool-bearing agent %q reads — its output reaches a privileged prompt",
			srcID, key, sink.ID),
		Location: sink.Source,
		Help:     chainAttackHelp,
	}
}

const chainAttackHelp = "tool_access bounds tools, not information flow: a restricted agent's output can launder an injection payload into a privileged agent's prompt. Audit whether the restricted agent's input is trusted; if not, drop tool_access on the consumer or insert a sanitizing / validating step between them. The runtime enforces the bound (key sanitization / context limit); this is an author-time advisory."

// isRestrictedAgent reports whether n is an agent node with a deliberate,
// recognized tool_access: none restriction. Invalid/typo values are DIP139's
// domain (and fail closed), not a deliberate "none" — they are not treated as a
// restricted source here.
func isRestrictedAgent(n *ir.Node) bool {
	if n == nil {
		return false
	}
	cfg, ok := n.Config.(ir.AgentConfig)
	return ok && strings.ToLower(strings.TrimSpace(cfg.ToolAccess)) == "none"
}

// isToolBearingAgent reports whether n is an agent node that holds the full tool
// catalog (tool_access omitted / empty after trim). Invalid values fail closed
// to no-tools at runtime, so they are not tool-bearing.
func isToolBearingAgent(n *ir.Node) bool {
	if n == nil {
		return false
	}
	cfg, ok := n.Config.(ir.AgentConfig)
	return ok && strings.TrimSpace(cfg.ToolAccess) == ""
}

// readSet returns the set of context keys a node declares in reads:.
func readSet(n *ir.Node) map[string]bool {
	set := make(map[string]bool, len(n.IO.Reads))
	for _, k := range n.IO.Reads {
		set[k] = true
	}
	return set
}
