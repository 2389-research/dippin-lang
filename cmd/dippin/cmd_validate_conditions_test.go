package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// reproDIP010 is the issue #98 repro: an edge using a tool-node field
// (marker_grep) as an operator is unparseable.
const reproDIP010 = `workflow Repro
  goal: "repro"
  start: A
  exit: Z

  tool A
    command: "echo ok"
  agent Z
    prompt: "done"

  edges
    A -> Z when marker_grep "^ok"
`

func writeRepro(t *testing.T) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "repro.dip")
	if err := os.WriteFile(p, []byte(reproDIP010), 0o644); err != nil {
		t.Fatalf("write repro: %v", err)
	}
	return p
}

// TestCmdValidate_DIP010 — validate must now FAIL (exit 1) and report DIP010 on
// stderr, where it previously passed silently.
func TestCmdValidate_DIP010(t *testing.T) {
	_, stderr, code := runCLI(t, "validate", writeRepro(t))
	if code != ExitError {
		t.Fatalf("want ExitError, got %d; stderr: %s", code, stderr)
	}
	if !strings.Contains(stderr, "error[DIP010]") {
		t.Errorf("expected error[DIP010] on stderr, got:\n%s", stderr)
	}
}

// TestCmdCheck_DIP010 — check writes diagnostics to stdout (text format) and
// exits non-zero on errors.
func TestCmdCheck_DIP010(t *testing.T) {
	stdout, _, code := runCLI(t, "check", "--format", "text", writeRepro(t))
	if code != ExitError {
		t.Fatalf("want ExitError, got %d", code)
	}
	if !strings.Contains(stdout, "error[DIP010]") {
		t.Errorf("expected error[DIP010] on stdout, got:\n%s", stdout)
	}
}

// TestCmdDoctor_DIP010 — doctor now exits non-zero when the report contains
// errors (the workflow can't execute), rather than always returning OK.
func TestCmdDoctor_DIP010(t *testing.T) {
	stdout, stderr, code := runCLI(t, "doctor", writeRepro(t))
	if code != ExitError {
		t.Fatalf("want ExitError from doctor on a workflow with errors, got %d; stderr: %s", code, stderr)
	}
	if !strings.Contains(stdout, "Health Report") {
		t.Errorf("expected the health report to still render, got:\n%s", stdout)
	}
}
