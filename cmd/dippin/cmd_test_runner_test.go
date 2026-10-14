package main

import (
	"strings"
	"testing"

	"github.com/2389-research/dippin-lang/simulate"
)

// `dippin test` on a suite that asserts a gate's default: path passes: an
// unattended case takes the default, and a preferred_label case overrides it
// (DB-001).
func TestCmdTest_GateDefault(t *testing.T) {
	simulate.ResetRunCounter()
	stdout, stderr, code := runCLI(t, "test", testdata("gate_default.dip"))
	if code != ExitOK {
		t.Fatalf("expected exit 0, got %d\nstdout: %s\nstderr: %s", code, stdout, stderr)
	}
	if strings.Contains(stdout, "FAIL") {
		t.Errorf("expected every case to pass, got: %s", stdout)
	}
}
