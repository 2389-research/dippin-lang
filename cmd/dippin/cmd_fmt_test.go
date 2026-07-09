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

func TestFmtMigrate_SynthesizesFailEdge(t *testing.T) {
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
	if strings.Contains(out, "fallback_target") {
		t.Errorf("migrated output must not keep fallback_target:\n%s", out)
	}
	if !strings.Contains(out, "T -> Esc  on fail") {
		t.Errorf("expected synthesized on-fail edge:\n%s", out)
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

func TestFmtMigrate_DivergentFlagsReviewExit(t *testing.T) {
	var stdout, stderr bytes.Buffer
	cli := &CLI{Stdout: &stdout, Stderr: &stderr, Format: FormatText}
	code := cli.CmdFmt([]string{"--migrate", writeTmp(t, v1Divergent)})
	if code != ExitMigrateReview {
		t.Fatalf("divergent migrate should exit ExitMigrateReview, got %d", code)
	}
	if !strings.Contains(stdout.String(), "# MIGRATION:") {
		t.Errorf("expected inline MIGRATION comment:\n%s", stdout.String())
	}
	if !strings.Contains(stderr.String(), "need review") {
		t.Errorf("expected stderr review summary, got:\n%s", stderr.String())
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
