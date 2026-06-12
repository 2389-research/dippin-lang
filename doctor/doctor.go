// Package doctor produces a health report card for Dippin workflows.
//
// It aggregates results from validator, coverage, and cost analysis
// into a single scored report with a letter grade (A-F) and actionable
// suggestions. No new analysis logic — purely aggregation.
package doctor

import (
	"sort"

	"github.com/2389-research/dippin-lang/cost"
	"github.com/2389-research/dippin-lang/coverage"
	"github.com/2389-research/dippin-lang/ir"
	"github.com/2389-research/dippin-lang/validator"
)

// Report is the full health report card.
type Report struct {
	Grade       string       `json:"grade"`
	Score       int          `json:"score"`
	Lint        LintSummary  `json:"lint"`
	Coverage    CovSummary   `json:"coverage"`
	Cost        CostSummary  `json:"cost"`
	Suggestions []Suggestion `json:"suggestions"`
}

// LintSummary summarizes validation and lint results.
type LintSummary struct {
	Errors   int `json:"errors"`
	Warnings int `json:"warnings"`
	Hints    int `json:"hints"`
}

// CovSummary summarizes coverage analysis results.
type CovSummary struct {
	ReachableNodes   int  `json:"reachable_nodes"`
	TotalNodes       int  `json:"total_nodes"`
	UnreachableCount int  `json:"unreachable_count"`
	AllTerminate     bool `json:"all_terminate"`
	UncoveredTools   int  `json:"uncovered_tools"`
}

// CostSummary summarizes cost analysis results.
type CostSummary struct {
	Total cost.CostRange `json:"total"`
}

// Suggestion is a single actionable recommendation.
type Suggestion struct {
	Category string `json:"category"`
	Message  string `json:"message"`
}

// Diagnose produces a health Report for the given workflow using default lint
// options.
func Diagnose(w *ir.Workflow, pricing cost.PricingTable) *Report {
	return DiagnoseWithOptions(w, pricing, validator.Options{})
}

// DiagnoseWithOptions produces a health Report, threading per-invocation lint
// options (e.g. a scoped extra-models catalog) into the lint pass so doctor's
// DIP108 summary matches `dippin lint --extra-models`.
func DiagnoseWithOptions(w *ir.Workflow, pricing cost.PricingTable, opts validator.Options) *Report {
	valResult := validator.Validate(w)
	lintResult := validator.LintWithOptions(w, opts)
	covReport := coverage.Analyze(w)
	costReport := cost.Analyze(w, pricing)

	r := &Report{}
	r.Lint = buildLintSummary(valResult, lintResult)
	r.Coverage = buildCovSummary(covReport)
	r.Cost = CostSummary{Total: costReport.Total}
	r.Score = computeScore(r)
	r.Grade = scoreToGrade(r.Score)
	r.Suggestions = buildSuggestions(r, covReport)
	return r
}

// buildLintSummary counts errors, warnings, and hints from both passes.
func buildLintSummary(val, lint validator.Result) LintSummary {
	var s LintSummary
	for _, d := range val.Diagnostics {
		classifyDiagnostic(&s, d)
	}
	for _, d := range lint.Diagnostics {
		classifyDiagnostic(&s, d)
	}
	return s
}

// classifyDiagnostic increments the appropriate counter in the summary.
func classifyDiagnostic(s *LintSummary, d validator.Diagnostic) {
	switch d.Severity {
	case validator.SeverityError:
		s.Errors++
	case validator.SeverityWarning:
		s.Warnings++
	default:
		s.Hints++
	}
}

// buildCovSummary extracts key coverage metrics.
func buildCovSummary(r *coverage.Report) CovSummary {
	uncovered := countUncoveredTools(r)
	return CovSummary{
		ReachableNodes:   r.Reachability.ReachableNodes,
		TotalNodes:       r.Reachability.TotalNodes,
		UnreachableCount: len(r.Reachability.UnreachableNodes),
		AllTerminate:     r.Termination.AllPathsTerminate,
		UncoveredTools:   uncovered,
	}
}

// countUncoveredTools counts tool nodes with partial coverage.
func countUncoveredTools(r *coverage.Report) int {
	count := 0
	for _, nc := range r.Nodes {
		if nc.Status == "partial" {
			count++
		}
	}
	return count
}

// computeScore starts at 100 and applies deductions.
func computeScore(r *Report) int {
	score := 100
	score -= r.Lint.Errors * 20
	score -= r.Lint.Warnings * 5
	score -= r.Coverage.UnreachableCount * 15
	if !r.Coverage.AllTerminate {
		score -= 15
	}
	score -= r.Coverage.UncoveredTools * 10
	if score < 0 {
		score = 0
	}
	return score
}

// scoreToGrade converts a numeric score to a letter grade.
func scoreToGrade(score int) string {
	switch {
	case score >= 90:
		return "A"
	case score >= 80:
		return "B"
	case score >= 70:
		return "C"
	case score >= 60:
		return "D"
	default:
		return "F"
	}
}

// buildSuggestions generates actionable recommendations.
func buildSuggestions(r *Report, cov *coverage.Report) []Suggestion {
	var s []Suggestion
	s = appendLintSuggestions(s, r)
	s = appendCovSuggestions(s, r, cov)
	sort.Slice(s, func(i, j int) bool {
		return s[i].Category < s[j].Category
	})
	return s
}

// appendLintSuggestions adds suggestions for lint issues.
func appendLintSuggestions(s []Suggestion, r *Report) []Suggestion {
	if r.Lint.Errors > 0 {
		s = append(s, Suggestion{
			Category: "lint",
			Message:  "fix validation errors before deployment — run `dippin check` for details",
		})
	}
	if r.Lint.Warnings > 0 {
		s = append(s, Suggestion{
			Category: "lint",
			Message:  "review lint warnings — run `dippin lint` for details",
		})
	}
	return s
}

// appendCovSuggestions adds suggestions for coverage issues.
func appendCovSuggestions(s []Suggestion, r *Report, cov *coverage.Report) []Suggestion {
	if r.Coverage.UnreachableCount > 0 {
		s = append(s, Suggestion{
			Category: "coverage",
			Message:  "remove or connect unreachable nodes — run `dippin coverage` for details",
		})
	}
	if !r.Coverage.AllTerminate {
		s = append(s, Suggestion{
			Category: "coverage",
			Message:  "ensure all paths reach exit — some nodes are dead ends",
		})
	}
	if r.Coverage.UncoveredTools > 0 {
		s = append(s, Suggestion{
			Category: "coverage",
			Message:  "add edges for uncovered tool outputs — run `dippin coverage` for details",
		})
	}
	return s
}
