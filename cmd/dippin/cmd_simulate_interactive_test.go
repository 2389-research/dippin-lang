package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const interactiveGateDip = `workflow ReviewGate
  goal: "Interactive gate routing"
  start: Start
  exit: Done

  agent Start
    prompt:
      Begin.

  human Review
    label: "Ship it?"
    mode: choice

  agent Ship
    prompt:
      Ship.

  agent Hold
    prompt:
      Hold.

  agent Done
    prompt:
      Done.

  edges
    Start -> Review
    Review -> Ship  label: "approve"
    Review -> Hold  label: "hold"
    Ship -> Done
    Hold -> Done
`

// withStdin points os.Stdin, which simulate --interactive reads, at input
// for the rest of the test.
func withStdin(t *testing.T, input string) {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := w.WriteString(input); err != nil {
		t.Fatal(err)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	old := os.Stdin
	os.Stdin = r
	t.Cleanup(func() {
		os.Stdin = old
		_ = r.Close()
	})
}

// pathLine returns the "path: ..." summary simulate prints to stderr.
func pathLine(stderr string) string {
	for _, line := range strings.Split(stderr, "\n") {
		if strings.HasPrefix(line, "path: ") {
			return line
		}
	}
	return ""
}

// `dippin simulate --interactive` follows the edge the user types.
func TestCmdSimulate_InteractiveChoiceRoutes(t *testing.T) {
	dipFile := filepath.Join(t.TempDir(), "review.dip")
	if err := os.WriteFile(dipFile, []byte(interactiveGateDip), 0644); err != nil {
		t.Fatal(err)
	}
	withStdin(t, "hold\n")

	_, stderr, code := runCLI(t, "simulate", "--interactive", dipFile)
	if code != ExitOK {
		t.Fatalf("expected exit 0, got %d; stderr: %s", code, stderr)
	}
	path := pathLine(stderr)
	if !strings.Contains(path, "Hold") || strings.Contains(path, "Ship") {
		t.Errorf("expected the typed choice to route to Hold, got %q", path)
	}
}

// Typed answers and --max-node-visits combine (#323 + #324): past the bound,
// an answer that already leaves the loop is honored, and only an answer that
// takes the `loop` edge is overridden — via default: "hold" — so the run
// never re-enters the loop.
func TestCmdSimulate_InteractiveLoopBound(t *testing.T) {
	cases := []struct {
		name, input, want, notWant string
	}{
		{"exit answer past bound is honored", "revise\nrevise\nrevise\napprove\n", "Done", "Hold"},
		{"loop answer past bound leaves via default", "revise\nrevise\nrevise\nrevise\nrevise\n", "Hold", ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			withStdin(t, c.input)
			_, stderr, code := runCLI(t, "simulate", "--interactive", "--max-node-visits", "3", testdata("gate_loop.dip"))
			if code != ExitOK {
				t.Fatalf("expected exit 0, got %d; stderr: %s", code, stderr)
			}
			path := pathLine(stderr)
			if !strings.Contains(path, c.want) {
				t.Errorf("expected path to reach %q, got %q", c.want, path)
			}
			if c.notWant != "" && strings.Contains(path, c.notWant) {
				t.Errorf("path must not visit %q, got %q", c.notWant, path)
			}
		})
	}
}
