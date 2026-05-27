// ABOUTME: `dippin pack` builds a deterministic .dipx bundle from an entry
// ABOUTME: .dip on disk, walking transitively-reachable subgraph refs.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/2389-research/dippin-lang/dipx"
	"github.com/2389-research/dippin-lang/validator"
)

// .dipx CLI exit codes (task spec). Distinct from the broader ExitCode space:
// ExitOK / ExitError / ExitUsageError describe the analysis pipeline's
// success/failure, whereas these codes describe pack/unpack/inspect outcomes
// where integrity vs IO vs cancellation must be distinguishable for tooling.
const (
	exitDipxOK             = 0
	exitDipxUserError      = 1
	exitDipxIntegrityError = 2
	exitDipxIOError        = 3
	exitDipxCancelled      = 4
)

// CmdPack is the dispatcher entry point. It bridges the CLI struct's testable
// I/O writers to runPack's internals.
func (c *CLI) CmdPack(args []string) ExitCode {
	return ExitCode(runPack(c.Stdout, c.Stderr, args))
}

// runPack implements `dippin pack <entry.dip> [-o output] [--dry-run]`.
// Returns one of exitDipx* per the .dipx CLI contract.
//
// Before invoking dipx.Pack on the user-supplied entry, we build a shadow
// source tree with all command_file: directives resolved to inline command:
// blocks. dipx.Pack then walks the shadow tree, so the produced bundle is
// self-contained and does not require the referenced script files to exist
// when the bundle is later opened.
func runPack(stdout, stderr io.Writer, args []string) int {
	entry, dest, dryRun, code := parsePackArgs(stderr, args)
	if code != -1 {
		return code
	}
	if code := validateEntryPrePack(stderr, entry); code != exitDipxOK {
		return code
	}
	shadowEntry, cleanup, err := prepShadowSourceTree(entry)
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return exitDipxUserError
	}
	defer cleanup()
	return dispatchPack(stdout, stderr, shadowEntry, dest, dryRun)
}

// dispatchPack routes the resolved shadow entry to the right dipx.Pack sink
// (dry-run discard, stdout, or atomic file). Extracted so runPack stays under
// the project's cyclomatic-5 cap.
func dispatchPack(stdout, stderr io.Writer, shadowEntry, dest string, dryRun bool) int {
	ctx := context.Background()
	if dryRun {
		_, err := dipx.Pack(ctx, shadowEntry, io.Discard)
		return classifyExit(stderr, err)
	}
	if dest == "-" {
		_, err := dipx.Pack(ctx, shadowEntry, stdout)
		return classifyExit(stderr, err)
	}
	return packToFile(stderr, ctx, shadowEntry, dest)
}

// validateEntryPrePack runs structural validation (DIP001-DIP009) on the entry
// workflow before invoking dipx.Pack. Per spec § "CLI", pack "runs structural
// validation first (same checks as `dippin validate`); refuses to pack invalid
// input." This lives at the CLI layer (not inside dipx) because the spec's
// loader-tier rule forbids dipx from importing validator.
func validateEntryPrePack(stderr io.Writer, entry string) int {
	w, err := loadWorkflow(entry)
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return exitDipxUserError
	}
	res := validator.Validate(w)
	if !res.HasErrors() {
		return exitDipxOK
	}
	for _, d := range res.Diagnostics {
		if d.Severity == validator.SeverityError {
			fmt.Fprintln(stderr, d.String())
		}
	}
	return exitDipxUserError
}

// parsePackArgs parses pack flags. On success returns (-1) so the caller knows
// to proceed; otherwise returns one of the exitDipx* codes.
func parsePackArgs(stderr io.Writer, args []string) (entry, dest string, dryRun bool, code int) {
	fs := flag.NewFlagSet("pack", flag.ContinueOnError)
	fs.SetOutput(stderr)
	output := fs.String("o", "", "output path (default: <entry>.dipx; '-' for stdout)")
	dry := fs.Bool("dry-run", false, "validate without writing output")
	if err := fs.Parse(args); err != nil {
		return "", "", false, exitDipxUserError
	}
	rest := fs.Args()
	if len(rest) != 1 {
		fmt.Fprintln(stderr, "usage: dippin pack <entry.dip> [-o output.dipx] [--dry-run]")
		return "", "", false, exitDipxUserError
	}
	entry = rest[0]
	if !strings.EqualFold(filepath.Ext(entry), ".dip") {
		fmt.Fprintf(stderr, "error: entry must be a .dip file (got %q)\n", entry)
		return "", "", false, exitDipxUserError
	}
	dest = *output
	if dest == "" {
		dest = strings.TrimSuffix(entry, filepath.Ext(entry)) + ".dipx"
	}
	return entry, dest, *dry, -1
}

// packToFile writes the bundle to a unique temp file alongside dest and
// atomically renames into place on success. Failures clean up the temp file.
// The unique-name pattern (os.CreateTemp) avoids clobbers between concurrent
// `dippin pack -o foo.dipx` invocations sharing the same dest.
func packToFile(stderr io.Writer, ctx context.Context, entry, dest string) int {
	f, err := os.CreateTemp(filepath.Dir(dest), filepath.Base(dest)+".*.tmp")
	if err != nil {
		fmt.Fprintln(stderr, err)
		return exitDipxIOError
	}
	tmp := f.Name()
	if _, err := dipx.Pack(ctx, entry, f); err != nil {
		_ = f.Close()
		_ = os.Remove(tmp)
		return classifyExit(stderr, err)
	}
	if err := f.Close(); err != nil {
		_ = os.Remove(tmp)
		fmt.Fprintln(stderr, err)
		return exitDipxIOError
	}
	if err := os.Rename(tmp, dest); err != nil {
		_ = os.Remove(tmp)
		fmt.Fprintln(stderr, err)
		return exitDipxIOError
	}
	return exitDipxOK
}

// classifyExit maps a dipx error to an exitDipx* code, printing it to stderr.
// Split from classifyDipxErr so each function stays under the project's
// cyclomatic-5 cap.
func classifyExit(stderr io.Writer, err error) int {
	if err == nil {
		return exitDipxOK
	}
	fmt.Fprintln(stderr, err)
	return classifyDipxErr(err)
}

// classifyDipxErr maps a non-nil dipx error to its exitDipx* code. Order
// matters: cancellation > integrity > I/O > user-error. I/O is detected via
// *os.PathError / *os.LinkError (or fs-level ErrNotExist / ErrPermission)
// so filesystem failures bubbling out of dipx.Pack/Open/Extract reach the
// documented exit-3 path instead of collapsing to user-error 1.
func classifyDipxErr(err error) int {
	if isCancelledErr(err) {
		return exitDipxCancelled
	}
	if isIntegrityErr(err) {
		return exitDipxIntegrityError
	}
	if isIOErr(err) {
		return exitDipxIOError
	}
	return exitDipxUserError
}

// isCancelledErr reports whether err is a context cancellation or deadline.
func isCancelledErr(err error) bool {
	return errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded)
}

// integritySentinels enumerates every error sentinel mapped to exit code 2
// per spec § "CLI exit codes". Update both this list and the spec table
// together; a CI test could be added in v1.2 to grep them against each
// other if drift becomes an issue.
var integritySentinels = []error{
	dipx.ErrHashMismatch,
	dipx.ErrManifestInvalid,
	dipx.ErrUnsupportedFormatVersion,
	dipx.ErrZipFeatureForbidden,
	dipx.ErrZipTruncated,
	dipx.ErrFileMissing,
	dipx.ErrFileUnexpected,
	dipx.ErrEntryNotInManifest,
	dipx.ErrRefEscape,
	dipx.ErrRefCycle,
	dipx.ErrCapExceeded,
	dipx.ErrPathUnsafe,
}

// isIntegrityErr reports whether err corresponds to any spec-enumerated
// integrity-class sentinel.
func isIntegrityErr(err error) bool {
	for _, sentinel := range integritySentinels {
		if errors.Is(err, sentinel) {
			return true
		}
	}
	return false
}

// isIOErr reports whether err is a filesystem I/O failure that should map to
// the bundle command's I/O exit code (3) rather than user-error (1).
func isIOErr(err error) bool {
	var pathErr *os.PathError
	var linkErr *os.LinkError
	if errors.As(err, &pathErr) || errors.As(err, &linkErr) {
		return true
	}
	return errors.Is(err, fs.ErrNotExist) || errors.Is(err, fs.ErrPermission)
}
