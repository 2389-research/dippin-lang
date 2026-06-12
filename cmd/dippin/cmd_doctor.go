package main

import (
	"fmt"
	"io"

	"github.com/2389-research/dippin-lang/cost"
	"github.com/2389-research/dippin-lang/doctor"
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

	// Note: doctor's lint summary does NOT apply the cross-file DIP146 pass /
	// DIP143 supersession. doctor.Diagnose composes validator.Lint inside the
	// doctor package and surfaces only hint COUNTS, not per-line diagnostics, so
	// the CLI-layer cross-file pass can't reach it without a layering change. The
	// effect is minor (a DIP143 vs DIP146 hint is counted the same; only a
	// fully-restricted child differs by one hint). `dippin lint`/`check`/`watch`
	// carry the precise DIP146 behavior. (#89)
	report := doctor.DiagnoseWithOptions(w, cost.DefaultPricing(), lintOptions(extraModels))
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
