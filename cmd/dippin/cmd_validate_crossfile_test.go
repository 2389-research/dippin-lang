package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCmdLint_DIP146SupersedesDIP143(t *testing.T) {
	dir := t.TempDir()
	write := func(name, content string) string {
		p := filepath.Join(dir, name)
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
		return p
	}
	entry := write("entry.dip", entryRestrictsRefs("child.dip"))
	write("child.dip", childZeroIntent)

	_, stderr, code := runCLI(t, "lint", entry)
	if code != ExitOK {
		t.Fatalf("want exit 0 (hints don't fail), got %d", code)
	}
	if !strings.Contains(stderr, "hint[DIP146]") {
		t.Errorf("expected DIP146 in output:\n%s", stderr)
	}
	if strings.Contains(stderr, "hint[DIP143]") {
		t.Errorf("DIP143 should be superseded for a resolved zero-intent child:\n%s", stderr)
	}
}

func TestCmdLint_RetainsDIP143OnPartialChild(t *testing.T) {
	dir := t.TempDir()
	entry := filepath.Join(dir, "entry.dip")
	if err := os.WriteFile(entry, []byte(entryRestrictsRefs("child.dip")), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "child.dip"), []byte(childPartial), 0o644); err != nil {
		t.Fatal(err)
	}
	_, stderr, _ := runCLI(t, "lint", entry)
	if strings.Contains(stderr, "hint[DIP146]") {
		t.Errorf("no DIP146 expected for partial-audit child:\n%s", stderr)
	}
	if !strings.Contains(stderr, "hint[DIP143]") {
		t.Errorf("DIP143 must be RETAINED for partial-audit child:\n%s", stderr)
	}
}

func TestCmdCheck_AppliesDIP146Supersession(t *testing.T) {
	dir := t.TempDir()
	entry := filepath.Join(dir, "entry.dip")
	if err := os.WriteFile(entry, []byte(entryRestrictsRefs("child.dip")), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "child.dip"), []byte(childZeroIntent), 0o644); err != nil {
		t.Fatal(err)
	}
	// check writes diagnostics to stdout; use text format for substring matching.
	stdout, _, _ := runCLI(t, "check", "--format", "text", entry)
	if !strings.Contains(stdout, "hint[DIP146]") {
		t.Errorf("dippin check should emit DIP146 for a zero-intent child:\n%s", stdout)
	}
	if strings.Contains(stdout, "hint[DIP143]") {
		t.Errorf("dippin check should supersede DIP143 for a resolved zero-intent child:\n%s", stdout)
	}
}

func TestCmdWatch_LintAndPrintAppliesDIP146(t *testing.T) {
	dir := t.TempDir()
	entry := filepath.Join(dir, "entry.dip")
	if err := os.WriteFile(entry, []byte(entryRestrictsRefs("child.dip")), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "child.dip"), []byte(childFullRestrict), 0o644); err != nil {
		t.Fatal(err)
	}
	var out, errBuf bytes.Buffer
	c := &CLI{Stdout: &out, Stderr: &errBuf}
	c.lintAndPrint(entry) // exercises watch's render path directly (no fsnotify loop)
	combined := out.String() + errBuf.String()
	// child is fully restricted => DIP143 superseded (confirmed safe), no DIP146.
	// Use bracketed form to avoid matching the test name in temp-dir paths.
	if strings.Contains(combined, "hint[DIP143]") {
		t.Errorf("watch should supersede DIP143 for a fully-restricted child:\n%s", combined)
	}
	if strings.Contains(combined, "hint[DIP146]") {
		t.Errorf("no DIP146 expected for a fully-restricted child:\n%s", combined)
	}
}
