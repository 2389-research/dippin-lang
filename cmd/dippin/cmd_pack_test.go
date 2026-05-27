package main

import (
	"archive/zip"
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/2389-research/dippin-lang/dipx"
)

const minimalDip = `workflow A
  goal: "Test"
  start: Ask
  exit: Done

  human Ask
    mode: freeform

  agent Done
    prompt:
      Complete the task.

  edges
    Ask -> Done
`

// writeMinimalEntry creates a temp dir with a minimal valid .dip and returns
// the dir, entry path.
func writeMinimalEntry(t *testing.T) (string, string) {
	t.Helper()
	dir := t.TempDir()
	entry := filepath.Join(dir, "a.dip")
	if err := os.WriteFile(entry, []byte(minimalDip), 0o644); err != nil {
		t.Fatalf("write entry: %v", err)
	}
	return dir, entry
}

func TestRunPack_Happy(t *testing.T) {
	dir, entry := writeMinimalEntry(t)
	out := filepath.Join(dir, "a.dipx")
	var stdout, stderr bytes.Buffer
	code := runPack(&stdout, &stderr, []string{"-o", out, entry})
	if code != exitDipxOK {
		t.Fatalf("exit code = %d; stderr=%s", code, stderr.String())
	}
	if _, err := os.Stat(out); err != nil {
		t.Fatalf("expected output: %v", err)
	}
}

func TestRunPack_DefaultOutputName(t *testing.T) {
	dir, entry := writeMinimalEntry(t)
	var stdout, stderr bytes.Buffer
	code := runPack(&stdout, &stderr, []string{entry})
	if code != exitDipxOK {
		t.Fatalf("exit code = %d; stderr=%s", code, stderr.String())
	}
	defaultOut := filepath.Join(dir, "a.dipx")
	if _, err := os.Stat(defaultOut); err != nil {
		t.Fatalf("expected default output at %s: %v", defaultOut, err)
	}
}

func TestRunPack_MissingEntry(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := runPack(&stdout, &stderr, []string{"/nonexistent.dip"})
	if code == exitDipxOK {
		t.Fatal("expected non-zero exit")
	}
}

func TestRunPack_DryRun(t *testing.T) {
	dir, entry := writeMinimalEntry(t)
	var stdout, stderr bytes.Buffer
	code := runPack(&stdout, &stderr, []string{"--dry-run", entry})
	if code != exitDipxOK {
		t.Fatalf("exit code = %d; stderr=%s", code, stderr.String())
	}
	out := filepath.Join(dir, "a.dipx")
	if _, err := os.Stat(out); err == nil {
		t.Fatal("dry-run should not produce output")
	}
}

func TestRunPack_Stdout(t *testing.T) {
	_, entry := writeMinimalEntry(t)
	var stdout, stderr bytes.Buffer
	code := runPack(&stdout, &stderr, []string{"-o", "-", entry})
	if code != exitDipxOK {
		t.Fatalf("exit code = %d; stderr=%s", code, stderr.String())
	}
	// Bundle starts with "PK" zip magic.
	if !bytes.HasPrefix(stdout.Bytes(), []byte{'P', 'K'}) {
		t.Fatalf("expected ZIP magic on stdout, got: %x", stdout.Bytes()[:4])
	}
}

func TestRunPack_NoArgs(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := runPack(&stdout, &stderr, nil)
	if code != exitDipxUserError {
		t.Fatalf("exit code = %d, expected %d", code, exitDipxUserError)
	}
}

// TestRunPack_RejectsInvalidWorkflow confirms Fix H1: pack runs structural
// validation (DIP001-DIP009) on the entry workflow first and refuses to pack
// when validation errors are present. Here the workflow declares exit: S but
// "S" has no outgoing edges (DIP004 would fire if there were unreachable nodes,
// or DIP002 if exit references a missing node, etc.). The workflow below
// references an undeclared start node, triggering DIP001.
func TestRunPack_RejectsInvalidWorkflow(t *testing.T) {
	dir := t.TempDir()
	// Structural error: start node "Missing" doesn't exist.
	invalid := `workflow A
  goal: "Test"
  start: Missing
  exit: Done

  agent Done
    prompt:
      Complete.
`
	entry := filepath.Join(dir, "a.dip")
	if err := os.WriteFile(entry, []byte(invalid), 0o644); err != nil {
		t.Fatalf("write entry: %v", err)
	}
	out := filepath.Join(dir, "a.dipx")
	var stdout, stderr bytes.Buffer
	code := runPack(&stdout, &stderr, []string{"-o", out, entry})
	if code == exitDipxOK {
		t.Fatalf("expected non-zero exit on structural validation failure; stderr=%s", stderr.String())
	}
	if _, err := os.Stat(out); err == nil {
		t.Fatalf("invalid workflow should not produce an output file at %s", out)
	}
}

// TestIsIntegrityErr_FullSentinelSet asserts isIntegrityErr matches the
// full integrity-class set per spec § "CLI exit codes". The 5-sentinel
// subset (HashMismatch, ManifestInvalid, ZipFeatureForbidden, ZipTruncated,
// UnsupportedFormatVersion) was the original v1 set; v1.1 Bundle 6
// adds 7 more sentinels.
func TestIsIntegrityErr_FullSentinelSet(t *testing.T) {
	cases := []struct {
		name string
		err  error
	}{
		{"HashMismatch", dipx.ErrHashMismatch},
		{"ManifestInvalid", dipx.ErrManifestInvalid},
		{"ZipFeatureForbidden", dipx.ErrZipFeatureForbidden},
		{"ZipTruncated", dipx.ErrZipTruncated},
		{"UnsupportedFormatVersion", dipx.ErrUnsupportedFormatVersion},
		{"FileMissing", dipx.ErrFileMissing},
		{"FileUnexpected", dipx.ErrFileUnexpected},
		{"EntryNotInManifest", dipx.ErrEntryNotInManifest},
		{"RefEscape", dipx.ErrRefEscape},
		{"RefCycle", dipx.ErrRefCycle},
		{"CapExceeded", dipx.ErrCapExceeded},
		{"PathUnsafe", dipx.ErrPathUnsafe},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if !isIntegrityErr(tc.err) {
				t.Fatalf("isIntegrityErr(%v) = false, want true", tc.err)
			}
		})
	}
}

func TestIsIntegrityErr_RejectsNonIntegrity(t *testing.T) {
	// EntryParse and SubgraphParse are user-class errors, not integrity.
	if isIntegrityErr(dipx.ErrEntryParse) {
		t.Error("isIntegrityErr(ErrEntryParse) = true, want false")
	}
	if isIntegrityErr(dipx.ErrSubgraphParse) {
		t.Error("isIntegrityErr(ErrSubgraphParse) = true, want false")
	}
}

// TestRunPack_InlinesCommandFile confirms that a .dip referencing a script
// via `command_file:` produces a self-contained bundle: the script content
// appears inside the bundled .dip as an inline `command:` block, and the
// external script file is not bundled. This is the core fix for the PR #71
// concern that pack was leaving command_file: directives literal and
// failing to embed the script.
func TestRunPack_InlinesCommandFile(t *testing.T) {
	dir := t.TempDir()
	scriptContent := "#!/bin/sh\nset -eu\necho hello-from-script\n"
	scriptDir := filepath.Join(dir, "scripts")
	if err := os.MkdirAll(scriptDir, 0o755); err != nil {
		t.Fatalf("mkdir scripts: %v", err)
	}
	if err := os.WriteFile(filepath.Join(scriptDir, "run.sh"), []byte(scriptContent), 0o644); err != nil {
		t.Fatalf("write script: %v", err)
	}
	dipSrc := `workflow A
  goal: "Pack-time inlining smoke test"
  start: Run
  exit: Done

  tool Run
    timeout: 30s
    command_file: scripts/run.sh

  agent Done
    prompt:
      Done.

  edges
    Run -> Done
`
	entry := filepath.Join(dir, "a.dip")
	if err := os.WriteFile(entry, []byte(dipSrc), 0o644); err != nil {
		t.Fatalf("write entry: %v", err)
	}
	out := filepath.Join(dir, "a.dipx")
	var stdout, stderr bytes.Buffer
	if code := runPack(&stdout, &stderr, []string{"-o", out, entry}); code != exitDipxOK {
		t.Fatalf("pack exit = %d; stderr=%s", code, stderr.String())
	}
	bundled := readBundledDip(t, out, "workflows/a.dip")
	if strings.Contains(bundled, "command_file:") {
		t.Fatalf("bundled .dip still has command_file: directive:\n%s", bundled)
	}
	if !strings.Contains(bundled, "echo hello-from-script") {
		t.Fatalf("bundled .dip missing inlined script content:\n%s", bundled)
	}
	// Script file should NOT appear in the bundle — only the .dip with the
	// script inlined into its command: block.
	for _, name := range listBundleEntries(t, out) {
		if strings.HasSuffix(name, ".sh") {
			t.Errorf("bundle unexpectedly contains script file: %s", name)
		}
	}
}

// readBundledDip extracts the named .dip file from the bundle and returns
// its contents as a string. Fails the test on any read/zip error.
func readBundledDip(t *testing.T, bundlePath, entryName string) string {
	t.Helper()
	r, err := zip.OpenReader(bundlePath)
	if err != nil {
		t.Fatalf("open bundle: %v", err)
	}
	defer r.Close()
	for _, f := range r.File {
		if f.Name != entryName {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			t.Fatalf("open %s: %v", entryName, err)
		}
		defer rc.Close()
		body, err := io.ReadAll(rc)
		if err != nil {
			t.Fatalf("read %s: %v", entryName, err)
		}
		return string(body)
	}
	t.Fatalf("entry %s not found in bundle", entryName)
	return ""
}

// listBundleEntries returns the names of every zip entry in bundlePath.
func listBundleEntries(t *testing.T, bundlePath string) []string {
	t.Helper()
	r, err := zip.OpenReader(bundlePath)
	if err != nil {
		t.Fatalf("open bundle: %v", err)
	}
	defer r.Close()
	names := make([]string, 0, len(r.File))
	for _, f := range r.File {
		names = append(names, f.Name)
	}
	return names
}
