// Package releasecheck validates release invariants for the dippin binary.
//
// These tests verify that cmd/dippin/generated-spec.md is checked in,
// current with scripts/gen-spec.sh, and that the binary builds from a
// source tree without .git. They shell out to external tools (cp, sh,
// git, go build) and should not be run in untrusted environments.
package releasecheck

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

func repoRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("failed to resolve test file path")
	}
	return filepath.Dir(filepath.Dir(file))
}

func TestGeneratedSpecSourceIsTrackedInGitCheckout(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping release check in short mode")
	}
	requireBinary(t, "git")
	root := repoRoot(t)
	if _, err := os.Stat(filepath.Join(root, ".git")); err != nil {
		t.Skip("git metadata is unavailable in this source tree")
	}
	cmd := exec.Command("git", "ls-files", "--error-unmatch", "cmd/dippin/generated-spec.md")
	cmd.Dir = root
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("cmd/dippin/generated-spec.md is not tracked in git: %v\n%s", err, out)
	}
}

func TestGeneratedSpecIsCurrentWithGenerator(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping release check in short mode")
	}
	if runtime.GOOS == "windows" {
		t.Skip("gen-spec.sh requires a POSIX shell")
	}
	requireBinary(t, "sh")

	root := repoRoot(t)
	specPath := filepath.Join(root, "cmd", "dippin", "generated-spec.md")

	before, err := os.ReadFile(specPath)
	if err != nil {
		t.Fatalf("read spec before regeneration: %v", err)
	}

	// Run gen-spec.sh in a temp copy so we don't mutate the working tree.
	copyDir := copyTree(t, root)
	runCmd(t, copyDir, "sh", "scripts/gen-spec.sh")

	after, err := os.ReadFile(filepath.Join(copyDir, "cmd", "dippin", "generated-spec.md"))
	if err != nil {
		t.Fatalf("read spec after regeneration: %v", err)
	}

	if !bytes.Equal(before, after) {
		t.Fatal("cmd/dippin/generated-spec.md is stale; run ./scripts/gen-spec.sh and commit the result")
	}
}

func TestCLIBuildsFromCopiedSourceTree(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping release check in short mode")
	}

	copyDir := copyTree(t, repoRoot(t))
	runCmd(t, copyDir, "go", "build", "./cmd/dippin")
}

// copyTree copies the repo tree to a temp directory using cp -a, excluding
// .git to simulate a module proxy download. Works on Linux and macOS.
func copyTree(t *testing.T, root string) string {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("copyTree requires cp -a (POSIX)")
	}
	requireBinary(t, "cp")
	dst := t.TempDir()
	runCmd(t, root, "cp", "-a", ".", dst)
	if err := os.RemoveAll(filepath.Join(dst, ".git")); err != nil {
		t.Fatalf("remove copied .git directory: %v", err)
	}
	return dst
}

func requireBinary(t *testing.T, name string) {
	t.Helper()
	if _, err := exec.LookPath(name); err != nil {
		t.Skipf("%s not found on PATH, skipping", name)
	}
}

func runCmd(t *testing.T, dir, name string, args ...string) {
	t.Helper()
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("%s failed: %v\n%s", cmd.String(), err, out)
	}
}
