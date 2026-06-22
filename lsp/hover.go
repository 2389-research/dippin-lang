package lsp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"go.lsp.dev/jsonrpc2"
	"go.lsp.dev/protocol"

	"github.com/2389-research/dippin-lang/ir"
)

// handleHover returns node information when hovering over a node ID.
func (s *Server) handleHover(ctx context.Context, reply jsonrpc2.Replier, req jsonrpc2.Request) error {
	var params protocol.HoverParams
	if err := json.Unmarshal(req.Params(), &params); err != nil {
		return reply(ctx, nil, err)
	}

	doc := s.store.get(string(params.TextDocument.URI))
	if doc == nil || doc.Parsed == nil {
		return reply(ctx, nil, nil)
	}

	node := findNodeAtPosition(doc, params.Position)
	if node == nil {
		return reply(ctx, nil, nil)
	}

	content := formatNodeHover(node, doc.Parsed)
	result := protocol.Hover{
		Contents: protocol.MarkupContent{
			Kind:  protocol.Markdown,
			Value: content,
		},
	}
	return reply(ctx, result, nil)
}

// findNodeAtPosition finds the node whose source location contains the position.
func findNodeAtPosition(doc *document, pos protocol.Position) *ir.Node {
	line := int(pos.Line) + 1
	for _, n := range doc.Parsed.Nodes {
		if nodeContainsLine(n, line) {
			return n
		}
	}
	return nil
}

// nodeContainsLine checks if a node's source range includes the given line.
func nodeContainsLine(n *ir.Node, line int) bool {
	if n.Source.Line == 0 {
		return false
	}
	end := n.Source.EndLine
	if end == 0 {
		end = n.Source.Line
	}
	return line >= n.Source.Line && line <= end
}

// formatNodeHover builds a markdown summary for a node.
func formatNodeHover(n *ir.Node, w *ir.Workflow) string {
	var b strings.Builder
	fmt.Fprintf(&b, "**%s** (`%s`)\n\n", n.ID, n.Kind)
	if n.Label != "" {
		fmt.Fprintf(&b, "Label: %s\n\n", n.Label)
	}
	b.WriteString(formatNodeConfig(n, w))
	return b.String()
}

// formatNodeConfig formats kind-specific configuration for hover.
func formatNodeConfig(n *ir.Node, w *ir.Workflow) string {
	switch cfg := n.Config.(type) {
	case ir.AgentConfig:
		return formatAgentHover(cfg, w)
	case ir.HumanConfig:
		return fmt.Sprintf("Mode: %s\n", cfg.Mode)
	case ir.ToolConfig:
		return formatToolHover(cfg)
	default:
		return ""
	}
}

// formatToolHover formats tool-specific hover info.
func formatToolHover(cfg ir.ToolConfig) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Command: `%s`\n", truncateStr(cfg.Command))
	if cfg.MarkerGrep != "" {
		fmt.Fprintf(&b, "marker_grep: `%s`\n", cfg.MarkerGrep)
	}
	if cfg.RouteRequired {
		b.WriteString("route_required: true\n")
	}
	if cfg.OutputLimit > 0 {
		fmt.Fprintf(&b, "output_limit: %d bytes\n", cfg.OutputLimit)
	}
	return b.String()
}

// formatAgentHover formats agent-specific hover info.
func formatAgentHover(cfg ir.AgentConfig, w *ir.Workflow) string {
	var b strings.Builder
	model := cfg.Model
	if model == "" {
		model = w.Defaults.Model
	}
	provider := cfg.Provider
	if provider == "" {
		provider = w.Defaults.Provider
	}
	fmt.Fprintf(&b, "Model: %s/%s\n", provider, model)
	if cfg.MaxTurns > 0 {
		fmt.Fprintf(&b, "Max Turns: %d\n", cfg.MaxTurns)
	}
	return b.String()
}

// hoverCommandMaxLen bounds the command preview shown in tool hovers.
const hoverCommandMaxLen = 80

// truncateStr shortens a string for display.
func truncateStr(s string) string {
	if len(s) <= hoverCommandMaxLen {
		return s
	}
	return s[:hoverCommandMaxLen] + "..."
}
