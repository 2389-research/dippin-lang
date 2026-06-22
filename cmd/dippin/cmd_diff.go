package main

import (
	"flag"
	"fmt"
	"io"

	"github.com/2389-research/dippin-lang/diff"
)

// CmdDiff performs a semantic diff between two workflow files.
func (c *CLI) CmdDiff(args []string) ExitCode {
	fs := flag.NewFlagSet("diff", flag.ContinueOnError)
	fs.SetOutput(c.Stderr)
	if err := fs.Parse(args); err != nil {
		return ExitUsageError
	}
	if fs.NArg() < 2 {
		fmt.Fprintln(c.Stderr, "usage: dippin diff <old.dip> <new.dip>")
		return ExitUsageError
	}

	oldW, err := loadWorkflow(fs.Arg(0))
	if err != nil {
		c.renderError(err, fs.Arg(0))
		return ExitError
	}

	newW, err := loadWorkflow(fs.Arg(1))
	if err != nil {
		c.renderError(err, fs.Arg(1))
		return ExitError
	}

	report := diff.Compare(oldW, newW)
	return c.renderDiffReport(report)
}

// renderDiffReport outputs the diff report in the selected format.
func (c *CLI) renderDiffReport(r *diff.Report) ExitCode {
	if c.Format == FormatJSON {
		return c.renderJSON(r)
	}
	renderDiffText(c.Stdout, r)
	return ExitOK
}

// renderDiffText writes a human-readable diff report.
func renderDiffText(w io.Writer, r *diff.Report) {
	fmt.Fprintln(w, "═══ Semantic Diff ═════════════════════════════════════════")
	renderDiffNodes(w, r)
	renderDiffEdges(w, r)
	renderDiffCost(w, r)
}

func renderDiffNodes(w io.Writer, r *diff.Report) {
	fmt.Fprintln(w, "─── Nodes ─────────────────────────────────────────────────")
	renderDiffNodeList(w, "+", r.NodesAdded)
	renderDiffNodeList(w, "-", r.NodesRemoved)
	renderModifiedNodes(w, r.NodesModified)
	if len(r.NodesAdded)+len(r.NodesRemoved)+len(r.NodesModified) == 0 {
		fmt.Fprintln(w, "  (no changes)")
	}
	fmt.Fprintln(w)
}

func renderDiffNodeList(w io.Writer, prefix string, ids []string) {
	for _, id := range ids {
		fmt.Fprintf(w, "  %s %s\n", prefix, id)
	}
}

func renderModifiedNodes(w io.Writer, mods []diff.NodeDiff) {
	for _, nd := range mods {
		fmt.Fprintf(w, "  ~ %s\n", nd.NodeID)
		for _, c := range nd.Changes {
			fmt.Fprintf(w, "      %s: %q → %q\n", c.Field, c.OldValue, c.NewValue)
		}
	}
}

func renderDiffEdges(w io.Writer, r *diff.Report) {
	fmt.Fprintln(w, "─── Edges ─────────────────────────────────────────────────")
	renderEdgeList(w, "+", r.EdgesAdded)
	renderEdgeList(w, "-", r.EdgesRemoved)
	if len(r.EdgesAdded)+len(r.EdgesRemoved) == 0 {
		fmt.Fprintln(w, "  (no changes)")
	}
	fmt.Fprintln(w)
}

func renderEdgeList(w io.Writer, prefix string, edges []diff.EdgeSummary) {
	for _, es := range edges {
		fmt.Fprintf(w, "  %s %s -> %s", prefix, es.From, es.To)
		if es.Condition != "" {
			fmt.Fprintf(w, " [%s]", es.Condition)
		}
		fmt.Fprintln(w)
	}
}

func renderDiffCost(w io.Writer, r *diff.Report) {
	fmt.Fprintln(w, "─── Cost Delta ────────────────────────────────────────────")
	fmt.Fprintf(w, "  Old: %s (expected)  New: %s (expected)\n",
		formatUSD(r.CostDelta.OldCost.Expected),
		formatUSD(r.CostDelta.NewCost.Expected))
	sign := "+"
	if r.CostDelta.Delta.Expected < 0 {
		sign = ""
	}
	fmt.Fprintf(w, "  Delta: %s%s (expected)\n",
		sign, formatUSD(r.CostDelta.Delta.Expected))
}
