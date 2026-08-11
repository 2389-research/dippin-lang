// ABOUTME: CLI command for exporting workflows to DOT graph format.
// ABOUTME: Flattens subgraph refs before export so output is always valid DOT.
package main

import (
	"flag"
	"fmt"

	"github.com/2389-research/dippin-lang/export"
	"github.com/2389-research/dippin-lang/flatten"
)

// CmdExportDOT exports a workflow to DOT graph format.
//   - --rankdir=LR|TB (default TB)
//   - --prompts includes prompt text in DOT node attributes
func (c *CLI) CmdExportDOT(args []string) ExitCode {
	fs := flag.NewFlagSet("export-dot", flag.ContinueOnError)
	fs.SetOutput(c.Stderr)
	rankdir := fs.String("rankdir", "TB", "graph layout direction (LR|TB)")
	prompts := fs.Bool("prompts", false, "include prompt text in DOT attributes")
	if err := fs.Parse(args); err != nil {
		return ExitUsageError
	}
	if fs.NArg() < 1 {
		fmt.Fprintln(c.Stderr, "usage: dippin export-dot [--rankdir=LR|TB] [--prompts] <file>")
		return ExitUsageError
	}

	path := fs.Arg(0)
	w, err := loadWorkflow(path)
	if err != nil {
		c.renderError(err, path)
		return ExitError
	}

	w, err = flatten.Flatten(w, &flatten.DiskResolver{}, flatten.Options{})
	if err != nil {
		c.renderError(err, path)
		return ExitError
	}

	opts := export.ExportOptions{
		IncludePrompts: *prompts,
		RankDir:        *rankdir,
	}

	dot := export.ExportDOT(w, opts)
	fmt.Fprint(c.Stdout, dot)
	return ExitOK
}

// CmdExportMermaid exports a workflow to a Mermaid flowchart (renders on GitHub,
// in docs, and in the playground). Subgraph refs are flattened first so the graph
// is complete.
func (c *CLI) CmdExportMermaid(args []string) ExitCode {
	fs := flag.NewFlagSet("export-mermaid", flag.ContinueOnError)
	fs.SetOutput(c.Stderr)
	if err := fs.Parse(args); err != nil {
		return ExitUsageError
	}
	if fs.NArg() < 1 {
		fmt.Fprintln(c.Stderr, "usage: dippin export-mermaid <file>")
		return ExitUsageError
	}

	path := fs.Arg(0)
	w, err := loadWorkflow(path)
	if err != nil {
		c.renderError(err, path)
		return ExitError
	}

	w, err = flatten.Flatten(w, &flatten.DiskResolver{}, flatten.Options{})
	if err != nil {
		c.renderError(err, path)
		return ExitError
	}

	fmt.Fprint(c.Stdout, export.ExportMermaid(w))
	return ExitOK
}
