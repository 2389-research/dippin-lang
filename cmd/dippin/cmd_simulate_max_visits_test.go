package main

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// worklistDip loops Pick -> Work -> Pick while the tool prints "more". A
// scenario that fixes Pick's output at "more" never lets the loop end.
const worklistDip = `workflow Worklist
  goal: "Work through a list until it is empty"
  start: Pick
  exit: Done

  tool Pick
    command:
      printf 'more'

  agent Work
    prompt:
      Work on the item.

  agent Done
    prompt:
      Done.

  edges
    Pick -> Work  when ctx.tool_stdout = more
    Pick -> Done  when ctx.tool_stdout = empty
    Work -> Pick  loop
`

func writeWorklist(t *testing.T) string {
	t.Helper()
	dipFile := filepath.Join(t.TempDir(), "worklist.dip")
	if err := os.WriteFile(dipFile, []byte(worklistDip), 0644); err != nil {
		t.Fatal(err)
	}
	return dipFile
}

// Without a bound, a loop whose marker the scenario fixes runs into the
// step limit.
func TestCmdSimulate_WorklistLoopWithoutBound(t *testing.T) {
	_, stderr, code := runCLI(t, "simulate", "--scenario", "Pick.tool_stdout=more", writeWorklist(t))
	if code == ExitOK {
		t.Fatalf("expected the unbounded loop to fail; stderr: %s", stderr)
	}
	if !strings.Contains(stderr, "exceeded 500 steps") {
		t.Errorf("expected the step-limit error, got: %s", stderr)
	}
}

// --max-node-visits forces the loop's exit edge once Pick has been visited
// more than N times, wherever the flag sits on the command line (DB-005).
func TestCmdSimulate_MaxNodeVisitsBoundsLoop(t *testing.T) {
	dipFile := writeWorklist(t)
	scenario := []string{"--scenario", "Pick.tool_stdout=more"}
	cases := map[string][]string{
		"flag before file": append(append([]string{"simulate", "--max-node-visits", "3"}, scenario...), dipFile),
		"flag after file":  append(append([]string{"simulate"}, scenario...), dipFile, "--max-node-visits", "3"),
		"flag=value form":  append(append([]string{"simulate", "--max-node-visits=3"}, scenario...), dipFile),
	}
	for name, args := range cases {
		t.Run(name, func(t *testing.T) {
			_, stderr, code := runCLI(t, args...)
			if code != ExitOK {
				t.Fatalf("expected exit 0, got %d; stderr: %s", code, stderr)
			}
			if !strings.Contains(stderr, "simulation complete: success") {
				t.Errorf("expected the bounded loop to reach the exit node, got: %s", stderr)
			}
		})
	}
}

// reorderSimulateArgs keeps --max-node-visits next to its value when the flag
// follows the file argument.
func TestReorderSimulateArgs_MaxNodeVisitsValue(t *testing.T) {
	got := reorderSimulateArgs([]string{"wf.dip", "--max-node-visits", "3"})
	want := []string{"--max-node-visits", "3", "wf.dip"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("reorderSimulateArgs = %v, want %v", got, want)
	}
}

// A negative --max-node-visits is a usage error, not a silent "no limit".
func TestCmdSimulate_MaxNodeVisitsNegative(t *testing.T) {
	dipFile := writeWorklist(t)
	for name, args := range map[string][]string{
		"flag=value form": {"simulate", "--max-node-visits=-1", dipFile},
		"separate value":  {"simulate", "--max-node-visits", "-2", dipFile},
	} {
		t.Run(name, func(t *testing.T) {
			_, stderr, code := runCLI(t, args...)
			if code != ExitUsageError {
				t.Fatalf("expected exit %d, got %d; stderr: %s", ExitUsageError, code, stderr)
			}
			if !strings.Contains(stderr, "--max-node-visits must be >= 0") {
				t.Errorf("expected a negative-value error, got: %s", stderr)
			}
		})
	}
}

// --all-paths does not honor --max-node-visits, so combining them is a usage
// error rather than a silently ignored flag.
func TestCmdSimulate_MaxNodeVisitsWithAllPaths(t *testing.T) {
	_, stderr, code := runCLI(t, "simulate", "--all-paths", "--max-node-visits", "3", writeWorklist(t))
	if code != ExitUsageError {
		t.Fatalf("expected exit %d, got %d; stderr: %s", ExitUsageError, code, stderr)
	}
	if !strings.Contains(stderr, "cannot be combined with --all-paths") {
		t.Errorf("expected a flag-combination error, got: %s", stderr)
	}
}
