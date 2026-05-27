package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/2389-research/dippin-lang/dipx"
	"github.com/2389-research/dippin-lang/ir"
	"github.com/2389-research/dippin-lang/migrate"
	"github.com/2389-research/dippin-lang/parser"
	"github.com/2389-research/dippin-lang/validator"
)

// OutputFormat controls whether diagnostics are rendered as human-readable text
// or machine-readable JSON (spec §12).
type OutputFormat int

const (
	FormatText OutputFormat = iota
	FormatJSON
	FormatDOT
)

// ExitCode encodes the result of a CLI invocation.
//   - 0: success / no issues
//   - 1: errors found (validation failures, parse errors, check-mode drift)
//   - 2: usage error (bad flags, missing arguments)
type ExitCode int

const (
	ExitOK         ExitCode = 0
	ExitError      ExitCode = 1
	ExitUsageError ExitCode = 2
)

// CLI holds parsed global state and dispatches to per-command implementations.
type CLI struct {
	Stdout io.Writer
	Stderr io.Writer
	Format OutputFormat
}

// commandDispatch maps subcommand names to their handler methods.
func (c *CLI) commandDispatch() map[string]func([]string) ExitCode {
	return map[string]func([]string) ExitCode{
		"parse":              c.CmdParse,
		"validate":           c.CmdValidate,
		"lint":               c.CmdLint,
		"check":              c.CmdCheck,
		"fmt":                c.CmdFmt,
		"new":                c.CmdNew,
		"export-dip":         c.CmdExportDIP,
		"export-dot":         c.CmdExportDOT,
		"migrate":            c.CmdMigrate,
		"validate-migration": c.CmdValidateMigration,
		"simulate":           c.CmdSimulate,
		"cost":               c.CmdCost,
		"coverage":           c.CmdCoverage,
		"doctor":             c.CmdDoctor,
		"optimize":           c.CmdOptimize,
		"diff":               c.CmdDiff,
		"feedback":           c.CmdFeedback,
		"explain":            c.CmdExplain,
		"unused":             c.CmdUnused,
		"graph":              c.CmdGraph,
		"test":               c.CmdTest,
		"watch":              c.CmdWatch,
		"lsp":                c.CmdLSP,
		"spec":               c.CmdSpec,
		"pack":               c.CmdPack,
		"unpack":             c.CmdUnpack,
		"inspect":            c.CmdInspect,
	}
}

// resolveFormat maps a format string to an OutputFormat value.
// Returns false if the format string is unknown.
func resolveFormat(s string) (OutputFormat, bool) {
	formats := map[string]OutputFormat{
		"text": FormatText,
		"json": FormatJSON,
		"dot":  FormatDOT,
	}
	f, ok := formats[s]
	return f, ok
}

// parseGlobalFlags parses global flags from args and returns the CLI, remaining
// args, and an exit code. If the exit code is not -1, the caller should return it.
func parseGlobalFlags(args []string, stdout, stderr io.Writer) (*CLI, []string, ExitCode) {
	c := &CLI{
		Stdout: stdout,
		Stderr: stderr,
		Format: FormatText,
	}

	globalFlags := flag.NewFlagSet("dippin", flag.ContinueOnError)
	globalFlags.SetOutput(stderr)
	formatStr := globalFlags.String("format", "text", "output format (text|json)")
	showVersion := globalFlags.Bool("version", false, "show version info")

	if err := globalFlags.Parse(args); err != nil {
		return nil, nil, ExitUsageError
	}

	if *showVersion {
		fmt.Fprintf(stdout, "dippin %s (commit: %s, built: %s)\n", version, commit, date)
		return nil, nil, ExitOK
	}

	f, ok := resolveFormat(*formatStr)
	if !ok {
		fmt.Fprintf(stderr, "unknown format: %s\n", *formatStr)
		return nil, nil, ExitUsageError
	}
	c.Format = f

	remaining := globalFlags.Args()
	if len(remaining) == 0 {
		printGlobalUsage(stderr)
		return nil, nil, ExitUsageError
	}

	return c, remaining, ExitCode(-1)
}

// Run is the testable entry point for the entire CLI. It accepts raw args
// (without the program name), captures output into the provided writers,
// and returns a deterministic exit code.
func Run(args []string, stdout, stderr io.Writer) ExitCode {
	if len(args) == 0 {
		printGlobalUsage(stderr)
		return ExitUsageError
	}

	c, remaining, code := parseGlobalFlags(args, stdout, stderr)
	if code != ExitCode(-1) {
		return code
	}

	return c.dispatch(remaining, stdout, stderr)
}

func (c *CLI) dispatch(remaining []string, stdout, stderr io.Writer) ExitCode {
	cmd := remaining[0]
	cmdArgs := remaining[1:]

	if cmd == "help" {
		printGlobalUsage(stdout)
		return ExitOK
	}

	if cmd == "version" {
		fmt.Fprintf(stdout, "dippin %s (commit: %s, built: %s)\n", version, commit, date)
		return ExitOK
	}

	handler, ok := c.commandDispatch()[cmd]
	if !ok {
		fmt.Fprintf(stderr, "unknown command: %s\n", cmd)
		return ExitUsageError
	}
	return handler(cmdArgs)
}

func printGlobalUsage(w io.Writer) {
	fmt.Fprintln(w, "usage: dippin [--format text|json] <command> [args]")
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "commands:")
	fmt.Fprintln(w, "  parse <file>                      Parse and output IR as JSON")
	fmt.Fprintln(w, "  validate <file>                   Run structural validation (DIP001-DIP009)")
	fmt.Fprintln(w, "  lint <file>                       Run validation and semantic linting")
	fmt.Fprintln(w, "  check [--format text|json] <file> Parse+validate+lint (JSON default, for LLM tooling)")
	fmt.Fprintln(w, "  fmt [--check] [--write] <file>    Format a .dip file")
	fmt.Fprintln(w, "  new [--name N] [--write F] <template>")
	fmt.Fprintln(w, "                                    Generate a starter .dip file")
	fmt.Fprintln(w, "                                    Templates: minimal, parallel, conditional,")
	fmt.Fprintln(w, "                                               review-loop, human-gate")
	fmt.Fprintln(w, "  export-dip <file>                 Export flattened workflow as .dip")
	fmt.Fprintln(w, "  export-dot [--rankdir] [--prompts] <file>")
	fmt.Fprintln(w, "                                    Export workflow to DOT format")
	fmt.Fprintln(w, "  migrate [--output <file>] <file.dot>")
	fmt.Fprintln(w, "                                    Convert DOT to .dip")
	fmt.Fprintln(w, "  validate-migration <old.dot> <new.dip>")
	fmt.Fprintln(w, "                                    Check parity between DOT and .dip")
	fmt.Fprintln(w, "  simulate [flags] <file>           Simulate workflow execution (JSONL events)")
	fmt.Fprintln(w, "                                    --scenario key=val  Inject context values")
	fmt.Fprintln(w, "                                    --interactive       Prompt at human nodes")
	fmt.Fprintln(w, "                                    --all-paths         Enumerate all paths")
	fmt.Fprintln(w, "  cost <file>                       Estimate workflow execution cost")
	fmt.Fprintln(w, "  coverage <file>                   Analyze edge coverage and reachability")
	fmt.Fprintln(w, "  doctor <file>                     Health report card (grade A-F, suggestions)")
	fmt.Fprintln(w, "  optimize <file>                   Model cost optimization suggestions")
	fmt.Fprintln(w, "  diff <old.dip> <new.dip>          Semantic diff between two workflows")
	fmt.Fprintln(w, "  feedback <workflow> <telemetry>    Compare predicted vs actual costs")
	fmt.Fprintln(w, "  explain <DIPxxx>                   Explain a diagnostic code in detail")
	fmt.Fprintln(w, "  graph [--compact] <file>           Render ASCII DAG of the workflow")
	fmt.Fprintln(w, "  test [--verbose] [--coverage] <file>")
	fmt.Fprintln(w, "  unused <file>                     Detect dead-branch nodes and wasted cost")
	fmt.Fprintln(w, "  watch <file-or-dir> [...]          Watch .dip files and re-lint on change")
	fmt.Fprintln(w, "  lsp                               Start LSP server on stdio")
	fmt.Fprintln(w, "  spec                              Print full language specification")
	fmt.Fprintln(w, "  pack [-o out] [--dry-run] <entry.dip>")
	fmt.Fprintln(w, "                                    Build a .dipx bundle from a .dip entry")
	fmt.Fprintln(w, "  unpack [-o dir] [--force] <bundle.dipx>")
	fmt.Fprintln(w, "                                    Extract a .dipx bundle to a directory")
	fmt.Fprintln(w, "  inspect [--format text|json] <bundle.dipx>")
	fmt.Fprintln(w, "                                    Print manifest, identity, and file list of a .dipx")
	fmt.Fprintln(w, "  version                           Show version info")
	fmt.Fprintln(w, "  help                              Show this help")
}

// parseSingleFileArg is a helper for commands that take a single file argument
// with no extra flags. Returns the file path and ExitCode(-1) on success, or
// empty string and an error exit code on failure.
func parseSingleFileArg(name, usage string, args []string, stderr io.Writer) (string, ExitCode) {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(stderr)
	if err := fs.Parse(args); err != nil {
		return "", ExitUsageError
	}
	if fs.NArg() < 1 {
		fmt.Fprintln(stderr, usage)
		return "", ExitUsageError
	}
	return fs.Arg(0), ExitCode(-1)
}

// parseAndResolveDip parses .dip source text and resolves any file directives
// (e.g. command_file:) against the source file's directory. This is the shared
// CLI-side parse path — LSP/WASM intentionally skip the resolver and view the
// unresolved IR.
func parseAndResolveDip(data []byte, path string) (*ir.Workflow, error) {
	p := parser.NewParser(string(data), path)
	w, err := p.Parse()
	if err != nil {
		return nil, err
	}
	if err := parser.ResolveFileDirectives(w, filepath.Dir(path)); err != nil {
		return nil, err
	}
	return w, nil
}

// parseFile reads and parses a .dip or .dipx file, returning the entry workflow.
// Unlike loadWorkflow it does not handle .dot — that's only used by the
// migrate-related commands. .dipx bundles are integrity-verified via dipx.Load.
func parseFile(path string) (*ir.Workflow, error) {
	if strings.HasSuffix(strings.ToLower(path), ".dipx") {
		src, err := dipx.Load(context.Background(), path)
		if err != nil {
			return nil, err
		}
		return src.Entry(), nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return parseAndResolveDip(data, path)
}

// loadWorkflow reads a file and parses it to IR. It auto-detects .dot vs .dip
// vs .dipx based on the file extension. .dipx bundles are routed through
// dipx.Load, which integrity-verifies the manifest and returns the entry
// workflow; .dip files keep the existing lightweight parse path; .dot files
// flow through the migrate package.
func loadWorkflow(path string) (*ir.Workflow, error) {
	if strings.HasSuffix(strings.ToLower(path), ".dipx") {
		src, err := dipx.Load(context.Background(), path)
		if err != nil {
			return nil, err
		}
		return src.Entry(), nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if strings.HasSuffix(path, ".dot") {
		return migrate.Migrate(string(data))
	}
	return parseAndResolveDip(data, path)
}

// renderError formats a Go error as a diagnostic and writes it to stderr.
// In JSON mode it wraps the error in the spec §12 JSON diagnostic format.
func (c *CLI) renderError(err error, path string) {
	if c.Format == FormatJSON {
		diag := validator.Diagnostic{
			Code:     "DIP000",
			Severity: validator.SeverityError,
			Message:  err.Error(),
			Location: ir.SourceLocation{File: path},
		}
		c.renderDiagnostics([]validator.Diagnostic{diag})
	} else {
		fmt.Fprintf(c.Stderr, "error: %v\n", err)
	}
}

// renderDiagnostics outputs diagnostics to stderr in either text or JSON format.
// The JSON format matches spec §12 exactly.
func (c *CLI) renderDiagnostics(diags []validator.Diagnostic) {
	if len(diags) == 0 {
		return
	}

	if c.Format == FormatJSON {
		renderDiagnosticsJSON(c.Stderr, diags)
	} else {
		renderDiagnosticsText(c.Stderr, diags)
	}
}

// renderDiagnosticsText writes diagnostics in the spec §12 text format:
//
//	error[DIP003]: unknown node reference "InterpretX" in edge
//	  --> pipeline.dip:45:5
//	  = help: did you mean "Interpret"?
func renderDiagnosticsText(w io.Writer, diags []validator.Diagnostic) {
	for _, d := range diags {
		fmt.Fprintln(w, d.String())
	}
}

// renderDiagnosticsJSON writes diagnostics as a JSON array matching the spec §12
// machine-readable format.
func renderDiagnosticsJSON(w io.Writer, diags []validator.Diagnostic) {
	type jsonLoc struct {
		File      string `json:"file"`
		Line      int    `json:"line"`
		Column    int    `json:"column"`
		EndLine   int    `json:"end_line"`
		EndColumn int    `json:"end_column"`
	}
	type jsonDiag struct {
		Code     string  `json:"code"`
		Severity string  `json:"severity"`
		Message  string  `json:"message"`
		Location jsonLoc `json:"location"`
		Help     string  `json:"help,omitempty"`
		Fix      string  `json:"fix,omitempty"`
	}

	out := make([]jsonDiag, 0, len(diags))
	for _, d := range diags {
		out = append(out, jsonDiag{
			Code:     d.Code,
			Severity: d.Severity.String(),
			Message:  d.Message,
			Location: jsonLoc{
				File:      d.Location.File,
				Line:      d.Location.Line,
				Column:    d.Location.Column,
				EndLine:   d.Location.EndLine,
				EndColumn: d.Location.EndColumn,
			},
			Help: d.Help,
			Fix:  d.Fix,
		})
	}
	b, _ := json.MarshalIndent(out, "", "  ")
	fmt.Fprintln(w, string(b))
}

// renderJSON marshals any value to JSON and writes it to stdout.
func (c *CLI) renderJSON(v interface{}) ExitCode {
	enc := json.NewEncoder(c.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(v); err != nil {
		fmt.Fprintf(c.Stderr, "error: %v\n", err)
		return ExitError
	}
	return ExitOK
}

// writeOutput writes content to a file path if given, otherwise to stdout.
func writeOutput(stdout, stderr io.Writer, path, content string) ExitCode {
	if path != "" {
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			fmt.Fprintf(stderr, "error writing file: %v\n", err)
			return ExitError
		}
		return ExitOK
	}
	fmt.Fprint(stdout, content)
	return ExitOK
}
