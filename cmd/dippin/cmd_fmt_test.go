package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeTmp(t *testing.T, body string) string {
	t.Helper()
	dir := t.TempDir()
	p := filepath.Join(dir, "in.dip")
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

const v1Fallback = `workflow W
  goal: "t"
  start: T
  exit: Done
  agent T
    prompt: run
    fallback_target: Esc
  agent Esc
    prompt: e
  agent Done
    prompt: d
  edges
    T -> Done  on success
`

// v1 fallback_target migrates to the dip-2 spelling fallback_retry_target as a
// node attribute — the retry-exhaustion route the engine reads — NOT an on-fail
// edge (#204 Option A). Lossless, exit OK.
func TestFmtMigrate_FallbackBecomesRetryAttr(t *testing.T) {
	var stdout, stderr bytes.Buffer
	cli := &CLI{Stdout: &stdout, Stderr: &stderr, Format: FormatText}
	code := cli.CmdFmt([]string{"--migrate", writeTmp(t, v1Fallback)})
	if code != ExitOK {
		t.Fatalf("clean migrate should exit OK, got %d; stderr=%s", code, stderr.String())
	}
	out := stdout.String()
	if !strings.HasPrefix(out, "dip 2\n") {
		t.Errorf("migrated output should declare dip 2:\n%s", out)
	}
	if !strings.Contains(out, "fallback_retry_target: Esc") {
		t.Errorf("expected fallback_retry_target node attr:\n%s", out)
	}
	if strings.Contains(out, "on fail") {
		t.Errorf("must NOT synthesize an on-fail edge (retry-exhaustion is a node attr):\n%s", out)
	}
	// dip-1 spelling must be gone.
	if strings.Contains(out, "fallback_target:") {
		t.Errorf("dip-1 fallback_target spelling must not survive:\n%s", out)
	}
}

const v1Divergent = `workflow W
  goal: "t"
  start: T
  exit: Done
  agent T
    prompt: run
    fallback_target: Esc
  agent Esc
    prompt: e
  agent Fix
    prompt: f
  agent Done
    prompt: d
  edges
    T -> Done  on success
    T -> Fix   on fail
`

// A fallback_target alongside an on-fail edge to a different node is no longer a
// conflict: they are distinct channels (retry-exhaustion attr vs. genuine-failure
// edge), so both survive and migration is clean (no review note).
func TestFmtMigrate_FallbackAndFailEdgeCoexist(t *testing.T) {
	var stdout, stderr bytes.Buffer
	cli := &CLI{Stdout: &stdout, Stderr: &stderr, Format: FormatText}
	code := cli.CmdFmt([]string{"--migrate", writeTmp(t, v1Divergent)})
	if code != ExitOK {
		t.Fatalf("fallback + on-fail edge should migrate cleanly, got %d; stderr=%s", code, stderr.String())
	}
	out := stdout.String()
	if !strings.Contains(out, "fallback_retry_target: Esc") {
		t.Errorf("expected fallback_retry_target node attr:\n%s", out)
	}
	if !strings.Contains(out, "T -> Fix  on fail") {
		t.Errorf("the pre-existing on-fail edge must survive:\n%s", out)
	}
}

// `fmt --migrate --check` must exit non-zero on a v1 file (migration would change
// it) and must NOT write output to stdout.
func TestFmtMigrate_CheckDetectsUnmigratedV1(t *testing.T) {
	var stdout, stderr bytes.Buffer
	cli := &CLI{Stdout: &stdout, Stderr: &stderr, Format: FormatText}
	code := cli.CmdFmt([]string{"--migrate", "--check", writeTmp(t, v1Fallback)})
	if code != ExitError {
		t.Fatalf("--migrate --check on a v1 file should exit ExitError, got %d", code)
	}
	if stdout.Len() != 0 {
		t.Errorf("--check must not write to stdout, got:\n%s", stdout.String())
	}
}

// `fmt --migrate --check` on an already-migrated dip 2 file exits OK.
func TestFmtMigrate_CheckCleanOnAlreadyV2(t *testing.T) {
	var s1, e1 bytes.Buffer
	if code := (&CLI{Stdout: &s1, Stderr: &e1, Format: FormatText}).CmdFmt([]string{"--migrate", writeTmp(t, v1Fallback)}); code != ExitOK {
		t.Fatalf("migrate should be clean, got %d", code)
	}
	var s2, e2 bytes.Buffer
	code := (&CLI{Stdout: &s2, Stderr: &e2, Format: FormatText}).CmdFmt([]string{"--migrate", "--check", writeTmp(t, s1.String())})
	if code != ExitOK {
		t.Fatalf("--migrate --check on an already-v2 file should exit OK, got %d; stderr=%s", code, e2.String())
	}
}

const v1NonSelfRetry = `workflow W
  goal: "t"
  start: Review
  exit: Done
  agent Review
    prompt: r
    max_retries: 3
    retry_target: Implement
  agent Implement
    prompt: i
  agent Done
    prompt: d
  edges
    Review -> Done  on success
`

// A non-self retry_target migrates losslessly to a dip-2 retry_target node attr
// (#204 Option A) — the retry channel the engine reads — instead of the old
// lossy loop-edge conversion. Exit OK, no review note.
func TestFmtMigrate_NonSelfRetryKeptAsAttr(t *testing.T) {
	var stdout, stderr bytes.Buffer
	cli := &CLI{Stdout: &stdout, Stderr: &stderr, Format: FormatText}
	code := cli.CmdFmt([]string{"--migrate", writeTmp(t, v1NonSelfRetry)})
	if code != ExitOK {
		t.Fatalf("non-self retry_target should migrate cleanly, got %d; stderr=%s", code, stderr.String())
	}
	out := stdout.String()
	if !strings.Contains(out, "retry_target: Implement") {
		t.Errorf("retry_target must be kept as a node attr:\n%s", out)
	}
	if strings.Contains(out, "loop") {
		t.Errorf("must NOT synthesize a loop edge:\n%s", out)
	}
}
