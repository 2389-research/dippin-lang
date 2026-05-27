package lsp

import (
	"context"
	"encoding/json"

	"go.lsp.dev/jsonrpc2"
	"go.lsp.dev/protocol"
)

// handleCompletion provides completion items for node IDs and field names.
func (s *Server) handleCompletion(ctx context.Context, reply jsonrpc2.Replier, req jsonrpc2.Request) error {
	var params protocol.CompletionParams
	if err := json.Unmarshal(req.Params(), &params); err != nil {
		return reply(ctx, nil, err)
	}

	doc := s.store.get(string(params.TextDocument.URI))
	if doc == nil || doc.Parsed == nil {
		return reply(ctx, nil, nil)
	}

	items := nodeIDCompletions(doc)
	items = append(items, fieldCompletions()...)
	return reply(ctx, items, nil)
}

// nodeIDCompletions returns completion items for all node IDs in the workflow.
func nodeIDCompletions(doc *document) []protocol.CompletionItem {
	var items []protocol.CompletionItem
	for _, n := range doc.Parsed.Nodes {
		items = append(items, protocol.CompletionItem{
			Label:  n.ID,
			Kind:   protocol.CompletionItemKindReference,
			Detail: string(n.Kind) + " node",
		})
	}
	return items
}

// fieldCompletions returns completion items for common .dip field names.
func fieldCompletions() []protocol.CompletionItem {
	fields := []struct{ name, detail string }{
		{"prompt:", "Agent prompt text"},
		{"model:", "LLM model name"},
		{"provider:", "LLM provider name"},
		{"max_turns:", "Maximum agentic turns"},
		{"label:", "Display label"},
		{"mode:", "Human node mode (choice|freeform|interview|yes_no)"},
		{"command:", "Tool node shell command"},
		{"goal_gate:", "Fail pipeline if node fails"},
		{"reasoning_effort:", "Reasoning effort (none|minimal|low|medium|high|xhigh|max)"},
		{"fidelity:", "Fidelity level"},
		{"tool_access:", "LLM tool catalog gate on agent (none = no tools; omit = full)"},
		{"marker_grep:", "Regex matched against tool stdout; sets ctx.tool_marker"},
		{"command_file:", "External script reference for tool node (path relative to .dip dir)"},
		{"route_required:", "Require _TRACKER_ROUTE= sentinel line from tool stdout"},
		{"output_limit:", "Per-node stdout byte cap (non-negative int; 0 = engine default)"},
	}
	var items []protocol.CompletionItem
	for _, f := range fields {
		items = append(items, protocol.CompletionItem{
			Label:  f.name,
			Kind:   protocol.CompletionItemKindProperty,
			Detail: f.detail,
		})
	}
	return items
}
