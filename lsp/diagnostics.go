package lsp

import (
	"context"

	"go.lsp.dev/protocol"

	"github.com/2389-research/dippin-lang/validator"
)

// publishDiagnostics runs lint and publishes diagnostics to the client.
func (s *Server) publishDiagnostics(ctx context.Context, doc *document) {
	conn := s.getConn()
	if conn == nil {
		return
	}

	diags := collectDiagnostics(doc)
	params := protocol.PublishDiagnosticsParams{
		URI:         protocol.DocumentURI(doc.URI),
		Version:     uint32(doc.Version),
		Diagnostics: diags,
	}
	_ = conn.Notify(ctx, "textDocument/publishDiagnostics", params)
}

// collectDiagnostics gathers all diagnostics for a document. parser.Parse()
// always returns a non-nil *ir.Workflow, so doc.Parsed is never nil here.
func collectDiagnostics(doc *document) []protocol.Diagnostic {
	return lintDiagnostics(doc)
}

// lintDiagnostics runs validator and lint, converting results to LSP diagnostics.
func lintDiagnostics(doc *document) []protocol.Diagnostic {
	valResult := validator.Validate(doc.Parsed)
	lintResult := validator.Lint(doc.Parsed)

	var diags []protocol.Diagnostic
	for _, d := range valResult.Diagnostics {
		diags = append(diags, convertDiagnostic(d))
	}
	for _, d := range lintResult.Diagnostics {
		diags = append(diags, convertDiagnostic(d))
	}
	return diags
}

// convertDiagnostic converts a validator Diagnostic to an LSP Diagnostic.
func convertDiagnostic(d validator.Diagnostic) protocol.Diagnostic {
	return protocol.Diagnostic{
		Range:    sourceRange(d),
		Severity: mapSeverity(d.Severity),
		Code:     d.Code,
		Source:   "dippin",
		Message:  d.Message,
	}
}

// mapSeverity converts validator severity to LSP severity.
func mapSeverity(s validator.Severity) protocol.DiagnosticSeverity {
	switch s {
	case validator.SeverityError:
		return protocol.DiagnosticSeverityError
	case validator.SeverityWarning:
		return protocol.DiagnosticSeverityWarning
	case validator.SeverityInfo:
		return protocol.DiagnosticSeverityInformation
	default:
		return protocol.DiagnosticSeverityHint
	}
}

// sourceRange converts a validator location to an LSP range.
func sourceRange(d validator.Diagnostic) protocol.Range {
	line := d.Location.Line
	if line > 0 {
		line--
	}
	endLine := d.Location.EndLine
	if endLine > 0 {
		endLine--
	}
	return protocol.Range{
		Start: protocol.Position{Line: uint32(line), Character: uint32(d.Location.Column)},
		End:   protocol.Position{Line: uint32(endLine), Character: uint32(d.Location.EndColumn)},
	}
}

// zeroRange returns a range at position 0:0.
func zeroRange() protocol.Range {
	return protocol.Range{
		Start: protocol.Position{Line: 0, Character: 0},
		End:   protocol.Position{Line: 0, Character: 0},
	}
}
