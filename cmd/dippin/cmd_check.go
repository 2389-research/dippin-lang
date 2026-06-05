package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"

	"github.com/2389-research/dippin-lang/ir"
	"github.com/2389-research/dippin-lang/validator"
)

// checkReportError reports an error in the appropriate format for the check command.
func checkReportError(stdout io.Writer, stderr io.Writer, formatStr string, errMsg string) {
	if formatStr == "json" {
		renderCheckJSON(stdout, false, nil, errMsg)
	} else {
		fmt.Fprintf(stderr, "error: %v\n", errMsg)
	}
}

// renderCheckTextResult renders check results in text format.
func renderCheckTextResult(w io.Writer, diags []validator.Diagnostic) {
	if len(diags) > 0 {
		renderDiagnosticsText(w, diags)
	} else {
		fmt.Fprintln(w, "check passed")
	}
}

// CmdCheck runs parse + validate + lint in one shot and outputs a compact
// JSON summary to stdout. Designed for LLM tool-calling loops.
// CmdCheck parses its own --format flag, defaulting to JSON (unlike other commands).
func (c *CLI) CmdCheck(args []string) ExitCode {
	fs := flag.NewFlagSet("check", flag.ContinueOnError)
	fs.SetOutput(c.Stderr)
	formatStr := fs.String("format", "json", "output format (json|text)")
	if err := fs.Parse(args); err != nil {
		return ExitUsageError
	}
	if fs.NArg() < 1 {
		fmt.Fprintln(c.Stderr, "usage: dippin check [--format json|text] <file>")
		return ExitUsageError
	}

	path := fs.Arg(0)
	w, parseErr := parseFile(path)
	if parseErr != nil {
		checkReportError(c.Stdout, c.Stderr, *formatStr, parseErr.Error())
		return ExitError
	}

	return runCheckPipeline(c.Stdout, *formatStr, w, path)
}

// runCheckPipeline validates, lints, and renders the check result.
func runCheckPipeline(stdout io.Writer, formatStr string, w *ir.Workflow, path string) ExitCode {
	valRes := validator.Validate(w)
	lintRes := validator.Lint(w)
	allDiags := append(valRes.Diagnostics, lintRes.Diagnostics...)
	allDiags = applyCrossFileToolAccess(allDiags, w, path)

	if formatStr == "json" {
		renderCheckJSON(stdout, !valRes.HasErrors(), allDiags, "")
	} else {
		renderCheckTextResult(stdout, allDiags)
	}

	if valRes.HasErrors() {
		return ExitError
	}
	return ExitOK
}

// checkDiag is the JSON representation of a single diagnostic in check output.
type checkDiag struct {
	Code     string `json:"code"`
	Severity string `json:"severity"`
	Message  string `json:"message"`
	Line     int    `json:"line,omitempty"`
	Fix      string `json:"fix,omitempty"`
}

// checkResult is the JSON representation of the check command output.
type checkResult struct {
	Valid            bool        `json:"valid"`
	Errors           int         `json:"errors"`
	Warnings         int         `json:"warnings"`
	Diagnostics      []checkDiag `json:"diagnostics"`
	SuggestedActions []string    `json:"suggested_actions"`
}

// toCheckDiag converts a validator diagnostic to a checkDiag.
func toCheckDiag(d validator.Diagnostic) checkDiag {
	return checkDiag{
		Code:     d.Code,
		Severity: d.Severity.String(),
		Message:  d.Message,
		Line:     d.Location.Line,
		Fix:      d.Fix,
	}
}

// countSeverities counts errors and warnings in a slice of diagnostics.
func countSeverities(diags []validator.Diagnostic) (int, int) {
	var errors, warnings int
	for _, d := range diags {
		switch d.Severity {
		case validator.SeverityError:
			errors++
		case validator.SeverityWarning:
			warnings++
		}
	}
	return errors, warnings
}

// collectUniqueFixes returns unique fix suggestions from diagnostics.
func collectUniqueFixes(diags []validator.Diagnostic) []string {
	var actions []string
	seen := make(map[string]bool)
	for _, d := range diags {
		if d.Fix != "" && !seen[d.Fix] {
			seen[d.Fix] = true
			actions = append(actions, d.Fix)
		}
	}
	return actions
}

// aggregateCheckDiags converts validator diagnostics into check output fields,
// returning the converted diagnostics, error/warning counts, and unique suggested actions.
func aggregateCheckDiags(diags []validator.Diagnostic) ([]checkDiag, int, int, []string) {
	cds := make([]checkDiag, 0, len(diags))
	for _, d := range diags {
		cds = append(cds, toCheckDiag(d))
	}
	errors, warnings := countSeverities(diags)
	actions := collectUniqueFixes(diags)
	return cds, errors, warnings, actions
}

// renderCheckJSON writes the compact check output to w.
func renderCheckJSON(w io.Writer, valid bool, diags []validator.Diagnostic, parseErr string) {
	res := checkResult{
		Valid:       valid,
		Diagnostics: make([]checkDiag, 0, len(diags)),
	}

	if parseErr != "" {
		res.Errors = 1
		res.Diagnostics = append(res.Diagnostics, checkDiag{
			Code:     "DIP000",
			Severity: "error",
			Message:  parseErr,
		})
	}

	cds, errors, warnings, actions := aggregateCheckDiags(diags)
	res.Diagnostics = append(res.Diagnostics, cds...)
	res.Errors += errors
	res.Warnings += warnings
	res.SuggestedActions = actions

	if res.SuggestedActions == nil {
		res.SuggestedActions = []string{}
	}

	b, _ := json.Marshal(res)
	fmt.Fprintln(w, string(b))
}
