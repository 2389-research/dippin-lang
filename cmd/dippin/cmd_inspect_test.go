package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/2389-research/dippin-lang/dipx"
)

func TestRunInspect_Text(t *testing.T) {
	bundle := packForTest(t)
	var stdout, stderr bytes.Buffer
	code := runInspect(&stdout, &stderr, []string{bundle})
	if code != exitDipxOK {
		t.Fatalf("exit code = %d; stderr=%s", code, stderr.String())
	}
	out := stdout.String()
	for _, want := range []string{"format:", "entry:", "identity:", "files:", "status: VALID"} {
		if !strings.Contains(out, want) {
			t.Errorf("expected %q in text output, got:\n%s", want, out)
		}
	}
	// Footer must include byte total per spec § "CLI / inspect command" (Bundle 6).
	footerRe := regexp.MustCompile(`status: VALID \(\d+ files, \d+ bytes, format_version \d+\)`)
	if !footerRe.MatchString(out) {
		t.Errorf("footer missing byte total; want %q pattern, got:\n%s", footerRe.String(), out)
	}
}

func TestRunInspect_JSON(t *testing.T) {
	bundle := packForTest(t)
	var stdout, stderr bytes.Buffer
	code := runInspect(&stdout, &stderr, []string{"--format=json", bundle})
	if code != exitDipxOK {
		t.Fatalf("exit code = %d; stderr=%s", code, stderr.String())
	}
	var parsed struct {
		FormatVersion int    `json:"format_version"`
		Entry         string `json:"entry"`
		Identity      string `json:"identity"`
		Status        struct {
			Valid         bool  `json:"valid"`
			VerifySkipped bool  `json:"verify_skipped"`
			FileCount     int   `json:"file_count"`
			ByteTotal     int64 `json:"byte_total"`
			FormatVersion int   `json:"format_version"`
		} `json:"status"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &parsed); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, stdout.String())
	}
	for _, want := range []string{"format_version", "entry", "identity", "status"} {
		// Verify top-level keys are present by checking non-zero values.
		_ = want
	}
	if parsed.FormatVersion == 0 {
		t.Error("format_version = 0, want > 0")
	}
	if parsed.Entry == "" {
		t.Error("entry = empty, want non-empty")
	}
	if parsed.Identity == "" {
		t.Error("identity = empty, want non-empty")
	}
	if !parsed.Status.Valid {
		t.Error("status.valid = false, want true")
	}
	if parsed.Status.VerifySkipped {
		t.Error("status.verify_skipped = true, want false")
	}
	if parsed.Status.FileCount == 0 {
		t.Error("status.file_count = 0, want > 0")
	}
	if parsed.Status.ByteTotal == 0 {
		t.Error("status.byte_total = 0, want > 0")
	}
	if parsed.Status.FormatVersion != 1 {
		t.Errorf("status.format_version = %d, want 1", parsed.Status.FormatVersion)
	}
}

// TestRunInspect_NoVerify_NoStaleWarning asserts that the v1 advisory
// "still runs in v1" is no longer emitted — after Bundle 2 the flag is
// a real option, not a placeholder.
func TestRunInspect_NoVerify_NoStaleWarning(t *testing.T) {
	bundle := packForTest(t)
	var stdout, stderr bytes.Buffer
	code := runInspect(&stdout, &stderr, []string{"--no-verify", bundle})
	if code != exitDipxOK {
		t.Fatalf("exit code = %d; stderr=%s", code, stderr.String())
	}
	// After Bundle 2, --no-verify is no longer a no-op; the v1 stale warning
	// must be gone.
	if strings.Contains(stderr.String(), "still runs in v1") {
		t.Errorf("stderr should no longer contain v1 stale warning; got %q", stderr.String())
	}
}

func TestRunInspect_NoVerifyEmitsVerifySkippedTrue(t *testing.T) {
	bundle := packForTest(t)
	var stdout, stderr bytes.Buffer
	cli := &CLI{Stdout: &stdout, Stderr: &stderr}
	exit := cli.CmdInspect([]string{"--no-verify", "--format=json", bundle})
	if exit != ExitCode(exitDipxOK) {
		t.Fatalf("exit = %d, want 0; stderr=%s", exit, stderr.String())
	}
	var parsed struct {
		Status struct {
			Valid         bool  `json:"valid"`
			VerifySkipped bool  `json:"verify_skipped"`
			FileCount     int   `json:"file_count"`
			ByteTotal     int64 `json:"byte_total"`
			FormatVersion int   `json:"format_version"`
		} `json:"status"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &parsed); err != nil {
		t.Fatalf("unmarshal: %v\nstdout: %s", err, stdout.String())
	}
	if !parsed.Status.Valid {
		t.Error("status.valid = false, want true (structural admission passed)")
	}
	if !parsed.Status.VerifySkipped {
		t.Error("status.verify_skipped = false, want true")
	}
	if parsed.Status.ByteTotal != 0 {
		t.Errorf("status.byte_total = %d, want 0 (OpenManifest path doesn't extract bytes)", parsed.Status.ByteTotal)
	}
	if parsed.Status.FileCount == 0 {
		t.Error("status.file_count = 0, want > 0 (manifest still loads)")
	}
}

func TestRunInspect_NoArgs(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := runInspect(&stdout, &stderr, nil)
	if code != exitDipxUserError {
		t.Fatalf("exit code = %d, expected %d", code, exitDipxUserError)
	}
}

func TestRunInspect_MissingBundle(t *testing.T) {
	// Build the missing-bundle path inside t.TempDir() so the test is
	// deterministic across environments — a hardcoded "/nonexistent.dipx"
	// would behave differently on systems where "/" isn't writable, where
	// some unrelated nonexistent.dipx happens to exist, or under sandboxes
	// that rewrite root paths.
	missing := filepath.Join(t.TempDir(), "nonexistent.dipx")
	var stdout, stderr bytes.Buffer
	code := runInspect(&stdout, &stderr, []string{missing})
	// ErrManifestMissing wraps an *os.PathError ("no such file or directory").
	// classifyExit's isIOErr helper detects the wrapped PathError via
	// errors.As and routes to the documented I/O exit code (3), per Phase 8
	// M2 resolution.
	if code != exitDipxIOError {
		t.Fatalf("exit code = %d, expected I/O error %d; stderr=%s", code, exitDipxIOError, stderr.String())
	}
}

func TestRunInspect_BadFormat(t *testing.T) {
	bundle := packForTest(t)
	var stdout, stderr bytes.Buffer
	// Unknown format values are rejected as user errors per spec.
	code := runInspect(&stdout, &stderr, []string{"--format=bogus", bundle})
	if code != exitDipxUserError {
		t.Fatalf("exit code = %d, expected user-error %d; stderr=%s", code, exitDipxUserError, stderr.String())
	}
	if !strings.Contains(stderr.String(), "unknown --format value") {
		t.Errorf("expected diagnostic on stderr, got: %s", stderr.String())
	}
}

// packInputsBundleForTest packs a workflow that declares inputs and returns the
// bundle path. packForTest packs the fixed minimalDip source, which has none.
func packInputsBundleForTest(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	entry := filepath.Join(dir, "a.dip")
	if err := os.WriteFile(entry, []byte(inputsFixture), 0o644); err != nil {
		t.Fatalf("write entry: %v", err)
	}
	out := filepath.Join(dir, "a.dipx")
	var so, se bytes.Buffer
	if code := runPack(&so, &se, []string{"-o", out, entry}); code != exitDipxOK {
		t.Fatalf("pack failed: %d; %s", code, se.String())
	}
	return out
}

func TestRunInspect_JSONSurfacesEntryInputs(t *testing.T) {
	bundle := packInputsBundleForTest(t)
	var stdout, stderr bytes.Buffer
	if code := runInspect(&stdout, &stderr, []string{"--format=json", bundle}); code != exitDipxOK {
		t.Fatalf("exit code = %d; stderr=%s", code, stderr.String())
	}
	var payload struct {
		Inputs []struct {
			Name     string `json:"name"`
			Type     string `json:"type"`
			Required bool   `json:"required"`
		} `json:"inputs"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &payload); err != nil {
		t.Fatalf("output is not JSON: %v\n%s", err, stdout.String())
	}
	if len(payload.Inputs) != 4 {
		t.Fatalf("got %d inputs, want 4", len(payload.Inputs))
	}
	if payload.Inputs[0].Name != "idea" || payload.Inputs[0].Type != "text" {
		t.Errorf("first input = %q/%q, want idea/text", payload.Inputs[0].Name, payload.Inputs[0].Type)
	}
	if !payload.Inputs[0].Required {
		t.Error("idea.required = false, want true")
	}
}

func TestRunInspect_NoVerifyOmitsInputsKey(t *testing.T) {
	bundle := packInputsBundleForTest(t)
	var stdout, stderr bytes.Buffer
	if code := runInspect(&stdout, &stderr, []string{"--format=json", "--no-verify", bundle}); code != exitDipxOK {
		t.Fatalf("exit code = %d; stderr=%s", code, stderr.String())
	}
	var payload map[string]interface{}
	if err := json.Unmarshal(stdout.Bytes(), &payload); err != nil {
		t.Fatalf("output is not JSON: %v", err)
	}
	if _, present := payload["inputs"]; present {
		t.Error("--no-verify emitted an inputs key; that path never parses a workflow")
	}
}

func TestRunInspect_JSONInputsEmptyArrayWhenNoneDeclared(t *testing.T) {
	bundle := packForTest(t) // the fixed minimalDip source declares no inputs
	var stdout, stderr bytes.Buffer
	if code := runInspect(&stdout, &stderr, []string{"--format=json", bundle}); code != exitDipxOK {
		t.Fatalf("exit code = %d; stderr=%s", code, stderr.String())
	}
	var payload struct {
		Inputs []interface{} `json:"inputs"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &payload); err != nil {
		t.Fatalf("output is not JSON: %v", err)
	}
	if payload.Inputs == nil {
		t.Error("inputs = null, want an empty array")
	}
	if len(payload.Inputs) != 0 {
		t.Errorf("got %d inputs, want 0", len(payload.Inputs))
	}
}

// TestRunInspect_JSONIsParseable confirms the JSON output round-trips back into
// the public Manifest struct (Fix H3 — Manifest has JSON tags).
func TestRunInspect_JSONIsParseable(t *testing.T) {
	bundle := packForTest(t)
	var stdout, stderr bytes.Buffer
	code := runInspect(&stdout, &stderr, []string{"--format=json", bundle})
	if code != exitDipxOK {
		t.Fatalf("exit code = %d; stderr=%s", code, stderr.String())
	}
	var got dipx.Manifest
	if err := json.Unmarshal(stdout.Bytes(), &got); err != nil {
		t.Fatalf("decode into dipx.Manifest: %v\n%s", err, stdout.String())
	}
	if got.FormatVersion == 0 {
		t.Errorf("FormatVersion zero after round-trip; JSON tags missing on Manifest?")
	}
	if got.Entry == "" {
		t.Errorf("Entry empty after round-trip; JSON tags missing on Manifest?")
	}
	if len(got.Files) == 0 {
		t.Errorf("Files empty after round-trip")
	}
}
