package main

import (
	"fmt"
	"io"

	"github.com/2389-research/dippin-lang/doctor"
	"github.com/2389-research/dippin-lang/validator"
)

// CmdDoctor produces a health report card for a workflow.
func (c *CLI) CmdDoctor(args []string) ExitCode {
	path, extraModels, code := parseLintArgs("doctor", "usage: dippin doctor [--extra-models spec] <file>", args, c)
	if code != ExitCode(-1) {
		return code
	}

	w, err := loadWorkflow(path)
	if err != nil {
		c.renderError(err, path)
		return ExitError
	}

	// doctor's hint count now reflects the cross-file DIP146 pass / DIP143
	// supersession: the native pass lives here in cmd/dippin (which doctor must
	// not import), so we run it and hand doctor the corrected diagnostics via
	// DiagnoseFromDiagnostics. Only the tool_access pass is applied — DIP160
	// cross-file input arity is a Warning and folding it in would change doctor's
	// score, out of scope for this hint-count fix. (#101, follows #89)
	diags := append(validator.Validate(w).Diagnostics,
		validator.LintWithOptions(w, lintOptions(extraModels)).Diagnostics...)
	diags = applyCrossFileToolAccess(diags, w, path)
	report := doctor.DiagnoseFromDiagnostics(w, diags)
	return c.renderDoctorReport(report)
}

// renderDoctorReport outputs the doctor report in the selected format and
// exits non-zero when the report contains errors — a workflow that fails
// structural validation (e.g. DIP010) cannot execute, so doctor must not
// greenlight it. Warnings/hints alone still exit OK.
func (c *CLI) renderDoctorReport(r *doctor.Report) ExitCode {
	if c.Format == FormatJSON {
		if code := c.renderJSON(r); code != ExitOK {
			return code
		}
	} else {
		renderDoctorText(c.Stdout, r)
	}
	if r.Lint.Errors > 0 {
		return ExitError
	}
	return ExitOK
}

// renderDoctorText writes a human-readable doctor report.
func renderDoctorText(w io.Writer, r *doctor.Report) {
	fmt.Fprintln(w, "═══ Health Report Card ════════════════════════════════════")
	fmt.Fprintf(w, "  Grade: %s  Score: %d/100\n", r.Grade, r.Score)
	fmt.Fprintln(w)
	renderDoctorLint(w, r)
	renderDoctorCoverage(w, r)
	renderDoctorCost(w, r)
	renderDoctorSuggestions(w, r)
}

func renderDoctorLint(w io.Writer, r *doctor.Report) {
	fmt.Fprintln(w, "─── Lint ──────────────────────────────────────────────────")
	fmt.Fprintf(w, "  Errors: %d  Warnings: %d  Hints: %d\n",
		r.Lint.Errors, r.Lint.Warnings, r.Lint.Hints)
	fmt.Fprintln(w)
}

func renderDoctorCoverage(w io.Writer, r *doctor.Report) {
	fmt.Fprintln(w, "─── Coverage ──────────────────────────────────────────────")
	fmt.Fprintf(w, "  Reachable: %d/%d nodes\n",
		r.Coverage.ReachableNodes, r.Coverage.TotalNodes)
	icon := "✓"
	if !r.Coverage.AllTerminate {
		icon = "✗"
	}
	fmt.Fprintf(w, "  %s All paths terminate\n", icon)
	if r.Coverage.UncoveredTools > 0 {
		fmt.Fprintf(w, "  ✗ %d tool node(s) with uncovered outputs\n", r.Coverage.UncoveredTools)
	}
	fmt.Fprintln(w)
}

func renderDoctorCost(w io.Writer, r *doctor.Report) {
	fmt.Fprintln(w, "─── Cost ──────────────────────────────────────────────────")
	fmt.Fprintf(w, "  Expected: %s  (range: %s – %s)\n",
		formatUSD(r.Cost.Total.Expected),
		formatUSD(r.Cost.Total.Min),
		formatUSD(r.Cost.Total.Max))
	fmt.Fprintln(w)
}

func renderDoctorSuggestions(w io.Writer, r *doctor.Report) {
	if len(r.Suggestions) == 0 {
		fmt.Fprintln(w, "─── No suggestions — workflow is healthy! ─────────────────")
		return
	}
	fmt.Fprintln(w, "─── Suggestions ───────────────────────────────────────────")
	for _, s := range r.Suggestions {
		fmt.Fprintf(w, "  [%s] %s\n", s.Category, s.Message)
	}
}
