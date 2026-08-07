// ABOUTME: `dippin inspect` prints a .dipx bundle's manifest, identity, and
// ABOUTME: file list. --no-verify routes to dipx.OpenManifest (manifest-only,
// ABOUTME: no hash verification, no parse, no ref walking).
package main

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"

	"github.com/2389-research/dippin-lang/dipx"
)

// CmdInspect is the dispatcher entry point.
func (c *CLI) CmdInspect(args []string) ExitCode {
	return ExitCode(runInspect(c.Stdout, c.Stderr, args))
}

// runInspect implements `dippin inspect <bundle.dipx> [--no-verify] [--format=text|json]`.
func runInspect(stdout, stderr io.Writer, args []string) int {
	opts, code := parseInspectArgs(stderr, args)
	if code != -1 {
		return code
	}
	if opts.noVerify {
		return runInspectNoVerify(stdout, stderr, opts)
	}
	return runInspectVerify(stdout, stderr, opts)
}

// runInspectVerify is the default path: full integrity check via dipx.Open.
func runInspectVerify(stdout, stderr io.Writer, opts inspectOpts) int {
	bundle, err := dipx.Open(context.Background(), opts.path)
	if err != nil {
		return classifyExit(stderr, err)
	}
	switch opts.format {
	case "text":
		return printInspectText(stdout, bundle)
	case "json":
		return printInspectJSON(stdout, stderr, bundle)
	default:
		fmt.Fprintf(stderr, "unknown --format value: %q (expected text or json)\n", opts.format)
		return exitDipxUserError
	}
}

// runInspectNoVerify is the --no-verify path: load only the manifest +
// identity via dipx.OpenManifest; emit the same shape with
// verify_skipped=true and byte_total=0 (we don't extract bytes here).
func runInspectNoVerify(stdout, stderr io.Writer, opts inspectOpts) int {
	manifest, identity, err := dipx.OpenManifest(context.Background(), opts.path)
	if err != nil {
		return classifyExit(stderr, err)
	}
	status := buildInspectStatus(manifest, 0, true)
	switch opts.format {
	case "text":
		return printManifestText(stdout, manifest, identity, status)
	case "json":
		return printManifestJSON(stdout, stderr, manifest, identity, status, nil)
	default:
		fmt.Fprintf(stderr, "unknown --format value: %q (expected text or json)\n", opts.format)
		return exitDipxUserError
	}
}

// inspectOpts collects parsed CLI state for runInspect.
type inspectOpts struct {
	path     string
	format   string
	noVerify bool
}

// parseInspectArgs parses inspect flags. On success returns (opts, -1);
// on failure returns (zero, non-zero exit code).
func parseInspectArgs(stderr io.Writer, args []string) (inspectOpts, int) {
	fs := flag.NewFlagSet("inspect", flag.ContinueOnError)
	fs.SetOutput(stderr)
	noVerify := fs.Bool("no-verify", false, "skip hash verification (forensic mode)")
	format := fs.String("format", "text", "output format: text or json")
	if err := fs.Parse(args); err != nil {
		return inspectOpts{}, exitDipxUserError
	}
	rest := fs.Args()
	if len(rest) != 1 {
		fmt.Fprintln(stderr, "usage: dippin inspect <bundle.dipx> [--no-verify] [--format=text|json]")
		return inspectOpts{}, exitDipxUserError
	}
	return inspectOpts{path: rest[0], format: *format, noVerify: *noVerify}, -1
}

// InspectStatus is the shared shape rendered by both text and JSON
// inspect output, per spec § "CLI / inspect command" (Bundle 6).
type InspectStatus struct {
	Valid         bool  `json:"valid"`
	VerifySkipped bool  `json:"verify_skipped"`
	FileCount     int   `json:"file_count"`
	ByteTotal     int64 `json:"byte_total"`
	FormatVersion int   `json:"format_version"`
}

// printInspectText writes a human-readable manifest summary for a fully-
// opened (hash-verified) bundle.
func printInspectText(stdout io.Writer, b *dipx.Bundle) int {
	m := b.Manifest()
	id := b.Identity()
	status := buildInspectStatus(m, b.ByteTotal(), false)
	return printManifestText(stdout, m, id, status)
}

// printManifestText is the shared text renderer. Used by both Open-side
// (printInspectText) and OpenManifest-side (--no-verify, Task 4) paths.
func printManifestText(stdout io.Writer, m dipx.Manifest, id [32]byte, status InspectStatus) int {
	fmt.Fprintf(stdout, "format: %d\n", m.FormatVersion)
	fmt.Fprintf(stdout, "entry:  %s\n", m.Entry)
	fmt.Fprintf(stdout, "identity: sha256:%s\n", hex.EncodeToString(id[:]))
	fmt.Fprintln(stdout, "files:")
	for _, e := range m.Files {
		fmt.Fprintf(stdout, "  %-50s sha256:%s\n", e.Path, e.SHA256)
	}
	label := "VALID"
	if status.VerifySkipped {
		label = "UNVERIFIED"
	}
	fmt.Fprintf(stdout, "status: %s (%d files, %d bytes, format_version %d)\n",
		label, status.FileCount, status.ByteTotal, status.FormatVersion)
	return exitDipxOK
}

// printInspectJSON writes the manifest as indented JSON with the structured
// status object. Encode failures are surfaced to stderr alongside the I/O
// exit code so an operator running `dippin inspect --format=json` in a
// script gets diagnostic context, not a bare non-zero exit.
func printInspectJSON(stdout, stderr io.Writer, b *dipx.Bundle) int {
	m := b.Manifest()
	id := b.Identity()
	status := buildInspectStatus(m, b.ByteTotal(), false)
	return printManifestJSON(stdout, stderr, m, id, status, inputsJSON(b.Entry()))
}

// printManifestJSON is the shared JSON renderer. Used by both Open-side
// and OpenManifest-side (--no-verify, Task 4) paths.
//
// inputs carries the entry workflow's declared input schema (#190) on the
// Open-side path. The --no-verify path reads the manifest only and never
// parses a workflow, so it passes nil and the key is omitted entirely —
// an "inputs": null there would read as "this bundle declares none".
func printManifestJSON(stdout, stderr io.Writer, m dipx.Manifest, id [32]byte, status InspectStatus, inputs []inputJSON) int {
	out := map[string]interface{}{
		"format_version": m.FormatVersion,
		"entry":          m.Entry,
		"identity":       "sha256:" + hex.EncodeToString(id[:]),
		"files":          m.Files,
		"status":         status,
	}
	if inputs != nil {
		out["inputs"] = inputs
	}
	enc := json.NewEncoder(stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(out); err != nil {
		fmt.Fprintln(stderr, err)
		return exitDipxIOError
	}
	return exitDipxOK
}

// buildInspectStatus packages the fields needed by both renderers.
// byteTotal is 0 when verifySkipped is true (we don't extract bytes
// in OpenManifest). valid is true unless we have an explicit reason
// to mark it false; admission-failure cases produce errors before
// reaching this code.
func buildInspectStatus(m dipx.Manifest, byteTotal int64, verifySkipped bool) InspectStatus {
	return InspectStatus{
		Valid:         true,
		VerifySkipped: verifySkipped,
		FileCount:     len(m.Files),
		ByteTotal:     byteTotal,
		FormatVersion: m.FormatVersion,
	}
}
