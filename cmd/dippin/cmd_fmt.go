package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/2389-research/dippin-lang/formatter"
	"github.com/2389-research/dippin-lang/ir"
	"github.com/2389-research/dippin-lang/parser"
	"github.com/2389-research/dippin-lang/simulate"
)

// fmtFlags holds the parsed flags for `dippin fmt`.
type fmtFlags struct {
	check, write, migrate bool
	path                  string
}

// parseFmtFlags parses `dippin fmt` arguments.
// Returns (flags, -1) on success or (zero, code) on error.
func parseFmtFlags(args []string, stderr io.Writer) (fmtFlags, ExitCode) {
	fs := flag.NewFlagSet("fmt", flag.ContinueOnError)
	fs.SetOutput(stderr)
	check := fs.Bool("check", false, "exit 1 if not canonically formatted")
	write := fs.Bool("write", false, "write formatted output back to source file")
	migrate := fs.Bool("migrate", false, "convert a v1 file to dip 2 (edges own destinations)")
	if err := fs.Parse(args); err != nil {
		return fmtFlags{}, ExitUsageError
	}
	if fs.NArg() < 1 {
		fmt.Fprintln(stderr, "usage: dippin fmt [--check] [--write] [--migrate] <file>")
		return fmtFlags{}, ExitUsageError
	}
	return fmtFlags{check: *check, write: *write, migrate: *migrate, path: fs.Arg(0)}, ExitCode(-1)
}

// CmdFmt formats a .dip file to canonical form.
//   - Default: print formatted output to stdout
//   - --check: exit 1 if input is not already canonical (for CI)
//   - --write: write formatted output back to the file in-place
//   - --migrate: convert a v1 file to dip 2 (edges own destinations)
func (c *CLI) CmdFmt(args []string) ExitCode {
	flags, code := parseFmtFlags(args, c.Stderr)
	if code != ExitCode(-1) {
		return code
	}
	return c.runFmt(flags)
}

// runFmt executes the parse → (migrate) → format → emit pipeline.
func (c *CLI) runFmt(flags fmtFlags) ExitCode {
	w, data, code := c.readAndParseFile(flags.path)
	if code != ExitCode(-1) {
		return code
	}
	notes := c.conditionalMigrate(w, flags.migrate)
	formatted := formatter.Format(w)
	if ec := c.emitFmt(flags.path, string(data), formatted, flags.check, flags.write); ec != ExitOK {
		return ec
	}
	return c.reportMigrationNotes(notes)
}

// conditionalMigrate runs the v1→v2 transform when migrate is true.
func (c *CLI) conditionalMigrate(w *ir.Workflow, migrate bool) []formatter.MigrationNote {
	if !migrate {
		return nil
	}
	return c.migrateWorkflow(w)
}

// readAndParseFile reads and parses a file. Returns (workflow, raw, -1) on success or
// (nil, nil, code) on failure.
func (c *CLI) readAndParseFile(path string) (*ir.Workflow, []byte, ExitCode) {
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(c.Stderr, "error: %v\n", err)
		return nil, nil, ExitError
	}
	w, err := parser.NewParser(string(data), path).Parse()
	if err != nil {
		c.renderError(err, path)
		return nil, nil, ExitError
	}
	return w, data, ExitCode(-1)
}

// migrateWorkflow parses conditions then runs the v1->v2 transform.
func (c *CLI) migrateWorkflow(w *ir.Workflow) []formatter.MigrationNote {
	_ = simulate.EnsureConditionsParsed(w)
	return formatter.MigrateToV2(w)
}

// reportMigrationNotes prints a stderr summary of review cases and returns the
// review exit code when any exist.
func (c *CLI) reportMigrationNotes(notes []formatter.MigrationNote) ExitCode {
	if len(notes) == 0 {
		return ExitOK
	}
	fmt.Fprintf(c.Stderr, "migrated with %d case(s) that need review:\n", len(notes))
	for _, n := range notes {
		fmt.Fprintf(c.Stderr, "  node %q: %s\n", n.Node, n.Message)
	}
	return ExitMigrateReview
}

// emitFmt routes formatted output to --check, --write, or stdout. With --check,
// the (already migrated, if --migrate) formatted text is compared to the
// original — so `fmt --migrate --check` exits non-zero exactly when the file is
// not already canonical dip 2 (i.e. migration would change it), never writing.
func (c *CLI) emitFmt(path, original, formatted string, check, write bool) ExitCode {
	if check {
		return c.fmtCheck(path, original, formatted)
	}
	return writeOutput(c.Stdout, c.Stderr, boolToPath(write, path), formatted)
}

// boolToPath returns path if cond is true, otherwise empty string.
func boolToPath(cond bool, path string) string {
	if cond {
		return path
	}
	return ""
}

// fmtCheck compares formatted output against original data and returns
// the appropriate exit code for --check mode.
func (c *CLI) fmtCheck(path, original, formatted string) ExitCode {
	if formatted != original {
		if c.Format == FormatText {
			fmt.Fprintf(c.Stderr, "%s: not canonically formatted\n", path)
		}
		return ExitError
	}
	return ExitOK
}
