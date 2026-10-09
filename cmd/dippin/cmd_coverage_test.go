package main

import (
	"strings"
	"testing"
)

// `dippin coverage` counts printf 'pass\n' and printf 'fail\n' as covered by
// edges that test pass and fail, instead of reporting both outputs missing
// (DB-003).
func TestCmdCoverage_PrintfNewlineMarkersCovered(t *testing.T) {
	stdout, stderr, code := runCLI(t, "coverage", testdata("printf_newline_coverage.dip"))
	if code != ExitOK {
		t.Fatalf("expected exit 0, got %d; stderr: %s", code, stderr)
	}
	if strings.Contains(stdout, "partial") || strings.Contains(stdout, "missing:") {
		t.Errorf("expected RunTool fully covered, got: %s", stdout)
	}
}
