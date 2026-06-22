package lsp

import (
	"context"
	"encoding/json"

	"go.lsp.dev/jsonrpc2"
	"go.lsp.dev/protocol"

	"github.com/2389-research/dippin-lang/ir"
)

// handleDocumentSymbol returns document symbols for nodes and edges.
func (s *Server) handleDocumentSymbol(ctx context.Context, reply jsonrpc2.Replier, req jsonrpc2.Request) error {
	var params protocol.DocumentSymbolParams
	if err := json.Unmarshal(req.Params(), &params); err != nil {
		return reply(ctx, nil, err)
	}

	doc := s.store.get(string(params.TextDocument.URI))
	if doc == nil {
		return reply(ctx, nil, nil)
	}

	symbols := buildSymbols(doc.Parsed)
	return reply(ctx, symbols, nil)
}

// buildSymbols creates document symbols for all nodes and edges.
func buildSymbols(w *ir.Workflow) []protocol.DocumentSymbol {
	var symbols []protocol.DocumentSymbol
	for _, n := range w.Nodes {
		symbols = append(symbols, nodeSymbol(n))
	}
	for _, e := range w.Edges {
		symbols = append(symbols, edgeSymbol(e))
	}
	return symbols
}

// nodeSymbol creates a DocumentSymbol for a node.
func nodeSymbol(n *ir.Node) protocol.DocumentSymbol {
	return protocol.DocumentSymbol{
		Name:           n.ID,
		Detail:         string(n.Kind),
		Kind:           nodeSymbolKind(n.Kind),
		Range:          nodeRange(n),
		SelectionRange: nodeRange(n),
	}
}

// edgeSymbol creates a DocumentSymbol for an edge.
func edgeSymbol(e *ir.Edge) protocol.DocumentSymbol {
	name := e.From + " -> " + e.To
	if e.Label != "" {
		name += " (" + e.Label + ")"
	}
	return protocol.DocumentSymbol{
		Name:           name,
		Kind:           protocol.SymbolKindEvent,
		Range:          edgeRange(e),
		SelectionRange: edgeRange(e),
	}
}

// nodeSymbolKinds maps IR node kinds to their LSP symbol kind.
var nodeSymbolKinds = map[ir.NodeKind]protocol.SymbolKind{
	ir.NodeAgent:       protocol.SymbolKindFunction,
	ir.NodeHuman:       protocol.SymbolKindInterface,
	ir.NodeTool:        protocol.SymbolKindMethod,
	ir.NodeParallel:    protocol.SymbolKindStruct,
	ir.NodeFanIn:       protocol.SymbolKindStruct,
	ir.NodeConditional: protocol.SymbolKindEnum,
	ir.NodeManagerLoop: protocol.SymbolKindClass,
}

// nodeSymbolKind maps a node kind to its LSP symbol kind.
func nodeSymbolKind(kind ir.NodeKind) protocol.SymbolKind {
	if sk, ok := nodeSymbolKinds[kind]; ok {
		return sk
	}
	return protocol.SymbolKindVariable
}

// nodeRange converts a node's source location to an LSP range.
func nodeRange(n *ir.Node) protocol.Range {
	line := n.Source.Line
	if line > 0 {
		line--
	}
	endLine := n.Source.EndLine
	if endLine > 0 {
		endLine--
	}
	return protocol.Range{
		Start: protocol.Position{Line: uint32(line), Character: uint32(n.Source.Column)},
		End:   protocol.Position{Line: uint32(endLine), Character: uint32(n.Source.EndColumn)},
	}
}

// edgeRange converts an edge's source location to an LSP range.
func edgeRange(e *ir.Edge) protocol.Range {
	line := e.Source.Line
	if line > 0 {
		line--
	}
	return protocol.Range{
		Start: protocol.Position{Line: uint32(line), Character: uint32(e.Source.Column)},
		End:   protocol.Position{Line: uint32(line), Character: uint32(e.Source.Column)},
	}
}
