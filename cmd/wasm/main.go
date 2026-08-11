//go:build wasm

package main

import (
	"encoding/json"
	"syscall/js"

	"github.com/2389-research/dippin-lang/doctor"
	"github.com/2389-research/dippin-lang/export"
	"github.com/2389-research/dippin-lang/formatter"
	"github.com/2389-research/dippin-lang/parser"
	"github.com/2389-research/dippin-lang/validator"
)

func main() {
	js.Global().Set("dippinParse", js.FuncOf(jsParse))
	js.Global().Set("dippinLint", js.FuncOf(jsLint))
	js.Global().Set("dippinFormat", js.FuncOf(jsFormat))
	js.Global().Set("dippinMermaid", js.FuncOf(jsMermaid))
	js.Global().Set("dippinDoctor", js.FuncOf(jsDoctor))

	// Block forever — WASM modules run as long-lived processes.
	select {}
}

// jsMermaid renders the workflow as a Mermaid flowchart (the playground Graph
// view). Subgraph refs are NOT flattened — the browser has no filesystem — so a
// subgraph node renders as a single boundary node.
func jsMermaid(_ js.Value, args []js.Value) interface{} {
	src := args[0].String()
	w, err := parser.NewParser(src, "playground.dip").Parse()
	if err != nil {
		return toJSON(map[string]string{"error": err.Error()})
	}
	return export.ExportMermaid(w)
}

// jsDoctor returns the health report card (grade, score, lint/coverage/cost
// summary, suggestions) as JSON.
func jsDoctor(_ js.Value, args []js.Value) interface{} {
	src := args[0].String()
	w, err := parser.NewParser(src, "playground.dip").Parse()
	if err != nil {
		return toJSON(map[string]string{"error": err.Error()})
	}
	return toJSON(doctor.DiagnoseWithOptions(w, validator.Options{}))
}

func jsParse(_ js.Value, args []js.Value) interface{} {
	src := args[0].String()
	p := parser.NewParser(src, "playground.dip")
	w, err := p.Parse()
	if err != nil {
		return toJSON(map[string]string{"error": err.Error()})
	}
	return toJSON(w)
}

func jsLint(_ js.Value, args []js.Value) interface{} {
	src := args[0].String()
	p := parser.NewParser(src, "playground.dip")
	w, err := p.Parse()
	if err != nil {
		return toJSON(map[string]string{"error": err.Error()})
	}
	valRes := validator.Validate(w)
	lintRes := validator.Lint(w)
	all := append(valRes.Diagnostics, lintRes.Diagnostics...)
	return toJSON(toLintDiags(all))
}

// lintDiag is the JSON shape of a diagnostic returned to the playground. It
// matches the CLI's camelCase contract with a string severity — see checkDiag
// in cmd/dippin/cmd_check.go — rather than the raw validator.Diagnostic, whose
// PascalCase fields and numeric Severity the playground JS cannot read.
type lintDiag struct {
	Code     string `json:"code"`
	Severity string `json:"severity"`
	Message  string `json:"message"`
	Line     int    `json:"line,omitempty"`
	Fix      string `json:"fix,omitempty"`
}

// toLintDiags converts validator diagnostics to the playground JSON shape. It
// always returns a non-nil slice so a clean workflow marshals to "[]".
func toLintDiags(diags []validator.Diagnostic) []lintDiag {
	out := make([]lintDiag, 0, len(diags))
	for _, d := range diags {
		out = append(out, lintDiag{
			Code:     d.Code,
			Severity: d.Severity.String(),
			Message:  d.Message,
			Line:     d.Location.Line,
			Fix:      d.Fix,
		})
	}
	return out
}

func jsFormat(_ js.Value, args []js.Value) interface{} {
	src := args[0].String()
	p := parser.NewParser(src, "playground.dip")
	w, err := p.Parse()
	if err != nil {
		return toJSON(map[string]string{"error": err.Error()})
	}
	return formatter.Format(w)
}

func toJSON(v interface{}) string {
	b, _ := json.Marshal(v)
	return string(b)
}
