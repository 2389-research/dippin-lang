package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/2389-research/dippin-lang/export"
	"github.com/2389-research/dippin-lang/ir"
	"github.com/2389-research/dippin-lang/simulate"
	"github.com/2389-research/dippin-lang/validator"
)

// CmdSimulate runs a dry-run simulation of a workflow, emitting JSONL events.
//
//   - Default: happy path (all success), output JSONL to stdout
//   - --scenario key=val: inject context values to explore different paths
//   - --interactive: prompt at human nodes via stdin
//   - --all-paths: enumerate all possible execution paths
func (c *CLI) CmdSimulate(args []string) ExitCode {
	fs := flag.NewFlagSet("simulate", flag.ContinueOnError)
	fs.SetOutput(c.Stderr)

	var scenarios scenarioFlags
	fs.Var(&scenarios, "scenario", "inject context value (key=val), repeatable")
	interactive := fs.Bool("interactive", false, "prompt at human nodes")
	allPaths := fs.Bool("all-paths", false, "enumerate all possible execution paths")

	args = reorderSimulateArgs(args)

	if err := fs.Parse(args); err != nil {
		return ExitUsageError
	}
	if fs.NArg() < 1 {
		fmt.Fprintln(c.Stderr, "usage: dippin simulate [--scenario key=val] [--interactive] [--all-paths] <file>")
		return ExitUsageError
	}

	path := fs.Arg(0)
	w, opts, code := c.prepareSimulation(path, scenarios.values, *interactive)
	if code != ExitCode(-1) {
		return code
	}

	if *allPaths {
		return c.simulateAllPaths(w, opts)
	}
	return c.simulateSingle(w, opts)
}

// prepareSimulation loads and validates the workflow, then builds simulation options.
func (c *CLI) prepareSimulation(path string, scenario map[string]string, interactive bool) (*ir.Workflow, simulate.Options, ExitCode) {
	w, err := loadWorkflow(path)
	if err != nil {
		c.renderError(err, path)
		return nil, simulate.Options{}, ExitError
	}

	valRes := validator.Validate(w)
	if valRes.HasErrors() {
		c.renderDiagnostics(valRes.Diagnostics)
		return nil, simulate.Options{}, ExitError
	}

	opts := simulate.Options{
		Scenario:    scenario,
		Interactive: interactive,
	}
	if interactive {
		opts.Stdin = os.Stdin
		opts.Stderr = c.Stderr
	}

	return w, opts, ExitCode(-1)
}

// emitPathResult emits a single simulation path result (separator, header, JSONL).
func (c *CLI) emitPathResult(i int, res *simulate.Result) error {
	if i > 0 {
		fmt.Fprintln(c.Stdout)
	}
	if c.Format == FormatText {
		fmt.Fprintf(c.Stderr, "--- path %d: %s (%d nodes: %s) ---\n",
			i+1, res.Status, res.NodesVisited, strings.Join(res.Path, " → "))
	}
	return simulate.EmitJSONL(c.Stdout, res.Events)
}

// simulateAllPaths runs simulation in all-paths mode and renders results.
func (c *CLI) simulateAllPaths(w *ir.Workflow, opts simulate.Options) ExitCode {
	results, err := simulate.RunAllPaths(w, opts.Scenario)
	if err != nil {
		fmt.Fprintf(c.Stderr, "simulation error: %v\n", err)
		return ExitError
	}

	for i, res := range results {
		if err := c.emitPathResult(i, res); err != nil {
			fmt.Fprintf(c.Stderr, "output error: %v\n", err)
			return ExitError
		}
	}

	if c.Format == FormatText {
		fmt.Fprintf(c.Stderr, "\n%d path(s) enumerated\n", len(results))
	}
	return ExitOK
}

// simulateSingle runs a single simulation path and renders the result.
func (c *CLI) simulateSingle(w *ir.Workflow, opts simulate.Options) ExitCode {
	res, err := simulate.Run(w, opts)
	if err != nil {
		fmt.Fprintf(c.Stderr, "simulation error: %v\n", err)
		return ExitError
	}

	if c.Format == FormatDOT {
		return c.renderSimulateDOT(w, res)
	}

	if err := simulate.EmitJSONL(c.Stdout, res.Events); err != nil {
		fmt.Fprintf(c.Stderr, "output error: %v\n", err)
		return ExitError
	}

	if c.Format == FormatText {
		fmt.Fprintf(c.Stderr, "simulation complete: %s (%d nodes visited)\n", res.Status, res.NodesVisited)
		fmt.Fprintf(c.Stderr, "path: %s\n", strings.Join(res.Path, " → "))
	}

	return ExitOK
}

// renderSimulateDOT renders simulation results as a DOT graph.
func (c *CLI) renderSimulateDOT(w *ir.Workflow, res *simulate.Result) ExitCode {
	dotOpts := export.ExportOptions{
		ExecutionPath: res.Path,
	}
	dot := export.ExportDOT(w, dotOpts)
	fmt.Fprint(c.Stdout, dot)
	return ExitOK
}

// scenarioFlags implements flag.Value for repeatable --scenario key=val flags.
type scenarioFlags struct {
	values map[string]string
}

func (s *scenarioFlags) String() string {
	if s.values == nil {
		return ""
	}
	parts := make([]string, 0, len(s.values))
	for k, v := range s.values {
		parts = append(parts, k+"="+v)
	}
	return strings.Join(parts, ",")
}

func (s *scenarioFlags) Set(val string) error {
	if s.values == nil {
		s.values = make(map[string]string)
	}
	parts := strings.SplitN(val, "=", 2)
	if len(parts) != 2 {
		return fmt.Errorf("scenario must be key=value, got %q", val)
	}
	s.values[parts[0]] = parts[1]
	return nil
}

// isValueFlag returns true if the flag name requires a separate value argument.
func isValueFlag(name string) bool {
	return name == "scenario"
}

// hasFollowingValue returns true if a value-taking flag at position i has a
// non-flag value following it.
func hasFollowingValue(args []string, i int) bool {
	return i+1 < len(args) && !strings.HasPrefix(args[i+1], "-")
}

// classifyArg appends a single arg (and possibly its value) to the flags or
// positional slices. It returns the number of args consumed (1 or 2).
func classifyArg(args []string, i int, flags, positional *[]string) int {
	arg := args[i]
	if !strings.HasPrefix(arg, "-") {
		*positional = append(*positional, arg)
		return 1
	}

	if strings.Contains(arg, "=") {
		*flags = append(*flags, arg)
		return 1
	}

	flagName := strings.TrimLeft(arg, "-")
	if isValueFlag(flagName) && hasFollowingValue(args, i) {
		*flags = append(*flags, arg, args[i+1])
		return 2
	}

	*flags = append(*flags, arg)
	return 1
}

// reorderSimulateArgs moves any flag arguments that appear after a positional
// argument (the .dip file path) to before it. This allows natural CLI syntax:
//
//	dippin simulate file.dip --scenario key=val --all-paths
//
// Standard flag.FlagSet stops parsing at the first non-flag argument, so we
// partition args into flags and non-flags, then recombine as [flags... files...].
func reorderSimulateArgs(args []string) []string {
	var flags []string
	var positional []string
	for i := 0; i < len(args); {
		i += classifyArg(args, i, &flags, &positional)
	}
	return append(flags, positional...)
}
