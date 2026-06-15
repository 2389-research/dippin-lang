package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/2389-research/dippin-lang/formatter"
	"github.com/2389-research/dippin-lang/parser"
)

// CmdFmt formats a .dip file to canonical form.
//   - Default: print formatted output to stdout
//   - --check: exit 1 if input is not already canonical (for CI)
//   - --write: write formatted output back to the file in-place
//   - --migrate: re-emit in the current .dip format version (v1→v1 identity)
func (c *CLI) CmdFmt(args []string) ExitCode {
	fs := flag.NewFlagSet("fmt", flag.ContinueOnError)
	fs.SetOutput(c.Stderr)
	check := fs.Bool("check", false, "exit 1 if not canonically formatted")
	write := fs.Bool("write", false, "write formatted output back to source file")
	migrate := fs.Bool("migrate", false, "re-emit in the current .dip format version (identity for already-current files)")
	if err := fs.Parse(args); err != nil {
		return ExitUsageError
	}
	if fs.NArg() < 1 {
		fmt.Fprintln(c.Stderr, "usage: dippin fmt [--check] [--write] [--migrate] <file>")
		return ExitUsageError
	}

	path := fs.Arg(0)
	data, formatted, code := c.parseAndFormat(path)
	if code != ExitCode(-1) {
		return code
	}
	return c.emitFmt(path, string(data), formatted, *check, *write, *migrate)
}

// emitFmt routes formatted output to --check, --write, or stdout.
// --migrate re-emits in the current .dip format version; it is a no-op identity
// pass for already-current files (v1→v2 transforms land with #134) and so
// suppresses the --check comparison.
func (c *CLI) emitFmt(path, original, formatted string, check, write, migrate bool) ExitCode {
	if check && !migrate {
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

// parseAndFormat reads and parses a file, returning the raw data, formatted
// output, and an exit code (ExitCode(-1) means success, continue processing).
func (c *CLI) parseAndFormat(path string) ([]byte, string, ExitCode) {
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(c.Stderr, "error: %v\n", err)
		return nil, "", ExitError
	}

	p := parser.NewParser(string(data), path)
	w, err := p.Parse()
	if err != nil {
		c.renderError(err, path)
		return nil, "", ExitError
	}

	return data, formatter.Format(w), ExitCode(-1)
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
