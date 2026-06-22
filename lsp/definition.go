package lsp

import (
	"context"
	"encoding/json"
	"strings"

	"go.lsp.dev/jsonrpc2"
	"go.lsp.dev/protocol"

	"github.com/2389-research/dippin-lang/ir"
)

// handleDefinition supports go-to-definition for node IDs in edge declarations.
func (s *Server) handleDefinition(ctx context.Context, reply jsonrpc2.Replier, req jsonrpc2.Request) error {
	var params protocol.DefinitionParams
	if err := json.Unmarshal(req.Params(), &params); err != nil {
		return reply(ctx, nil, err)
	}

	loc := s.resolveDefinition(params)
	return reply(ctx, loc, nil)
}

// resolveDefinition finds the definition location for the word at the cursor.
func (s *Server) resolveDefinition(params protocol.DefinitionParams) *protocol.Location {
	doc := s.store.get(string(params.TextDocument.URI))
	if doc == nil {
		return nil
	}
	word := wordAtPosition(doc.Content, params.Position)
	if word == "" {
		return nil
	}
	node := doc.Parsed.Node(word)
	if node == nil {
		return nil
	}
	loc := nodeLocation(string(params.TextDocument.URI), node)
	return &loc
}

// wordAtPosition extracts the word (alphanumeric + underscore) at the given position.
func wordAtPosition(content string, pos protocol.Position) string {
	lines := strings.Split(content, "\n")
	lineIdx := int(pos.Line)
	if lineIdx >= len(lines) {
		return ""
	}
	line := lines[lineIdx]
	col := int(pos.Character)
	if col >= len(line) {
		return ""
	}
	return extractWord(line, col)
}

// extractWord finds the word boundaries around a column position.
func extractWord(line string, col int) string {
	start := wordStart(line, col)
	end := wordEnd(line, col)
	if start == end {
		return ""
	}
	return line[start:end]
}

// wordStart scans backward from col to find the word's start.
func wordStart(line string, col int) int {
	i := col
	for i > 0 && isWordChar(line[i-1]) {
		i--
	}
	return i
}

// wordEnd scans forward from col to find the word's end.
func wordEnd(line string, col int) int {
	i := col
	for i < len(line) && isWordChar(line[i]) {
		i++
	}
	return i
}

// isWordChar returns true for characters that form identifiers.
func isWordChar(c byte) bool {
	return isLetter(c) || isDigit(c) || c == '_'
}

func isLetter(c byte) bool { return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') }
func isDigit(c byte) bool  { return c >= '0' && c <= '9' }

// nodeLocation converts a node's source location to an LSP Location.
func nodeLocation(uri string, n *ir.Node) protocol.Location {
	line := n.Source.Line
	if line > 0 {
		line--
	}
	return protocol.Location{
		URI: protocol.DocumentURI(uri),
		Range: protocol.Range{
			Start: protocol.Position{Line: uint32(line), Character: uint32(n.Source.Column)},
			End:   protocol.Position{Line: uint32(line), Character: uint32(n.Source.Column)},
		},
	}
}
