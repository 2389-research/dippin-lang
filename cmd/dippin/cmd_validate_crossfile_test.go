package main

import (
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
