package validator

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/2389-research/dippin-lang/parser"
)

func TestEnumerateMarkers(t *testing.T) {
	cases := []struct {
		in   string
		want []string
		ok   bool
	}{
		{"^(tests-ok|tests-failed)$", []string{"tests-failed", "tests-ok"}, true},
		{"^(a|b|c)$", []string{"a", "b", "c"}, true},
		{"(a|b)", []string{"a", "b"}, true},
		{"tests_pass", []string{"tests_pass"}, true},
		{"^pass$", []string{"pass"}, true},
		{"^(a|a|b)$", []string{"a", "b"}, true}, // dedup
		// non-enumerable → (nil,false):
		{"^(a|b|)$", nil, false},  // empty branch
		{"(a|b)|c", nil, false},   // group not full-span
		{"(a|b)?", nil, false},    // quantifier after group
		{"^(a.b|c)$", nil, false}, // metachar in branch
		{"^\\d+$", nil, false},    // metachars
		{".*fail.*", nil, false},  // metachars
		{"(?i)(a|b)", nil, false}, // flags: inner group has metachars, not full-span
	}
	for _, tc := range cases {
		got, ok := enumerateMarkers(tc.in)
		if ok != tc.ok {
			t.Errorf("%q: ok = %v, want %v", tc.in, ok, tc.ok)
			continue
		}
		if ok {
			sort.Strings(got)
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("%q: markers = %v, want %v", tc.in, got, tc.want)
			}
		}
	}
}

// lintFor parses .dip source and returns the diagnostics for `code`.
func markerDiagsFor(t *testing.T, src string) []Diagnostic {
	t.Helper()
	w, err := parser.NewParser(src, "test.dip").Parse()
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	var out []Diagnostic
	for _, d := range Lint(w).Diagnostics {
		if d.Code == DIP152 {
			out = append(out, d)
		}
	}
	return out
}

const markerBase = `workflow W
  goal: "t"
  start: RunTests
  exit: Done

  tool RunTests
    command: run
    marker_grep: "%s"

  agent Done
    prompt: done

  edges
%s`

func mk(grep, edges string) string { return fmt.Sprintf(markerBase, grep, edges) }

func TestDIP152_GapWarns(t *testing.T) {
	diags := markerDiagsFor(t, mk("^(go|stop)$", "    RunTests -> Done on go\n"))
	if len(diags) != 1 {
		t.Fatalf("want 1 DIP152, got %d: %v", len(diags), diags)
	}
	if !strings.Contains(diags[0].Message, "stop") {
		t.Errorf("message should name unrouted marker stop: %q", diags[0].Message)
	}
}

func TestDIP152_FullyRoutedClean(t *testing.T) {
	diags := markerDiagsFor(t, mk("^(go|stop)$",
		"    RunTests -> Done on go\n    RunTests -> Done on stop\n"))
	if len(diags) != 0 {
		t.Fatalf("want 0 DIP152, got %v", diags)
	}
}

func TestDIP152_ElseCovered(t *testing.T) {
	diags := markerDiagsFor(t, mk("^(go|stop)$",
		"    RunTests -> Done on go\n    else -> Done\n"))
	if len(diags) != 0 {
		t.Fatalf("else default should suppress DIP152, got %v", diags)
	}
}

func TestDIP152_UnconditionalCovered(t *testing.T) {
	diags := markerDiagsFor(t, mk("^(go|stop)$",
		"    RunTests -> Done on go\n    RunTests -> Done\n"))
	if len(diags) != 0 {
		t.Fatalf("unconditional edge should suppress DIP152, got %v", diags)
	}
}

func TestDIP152_MultiMarkerGap(t *testing.T) {
	diags := markerDiagsFor(t, mk("^(a|b|c)$", "    RunTests -> Done on a\n"))
	if len(diags) != 1 || !strings.Contains(diags[0].Message, "b, c") {
		t.Fatalf("want DIP152 listing b, c: %v", diags)
	}
}

// --- false-positive guards (squad blockers) ---

func TestDIP152_OrRoutingClean(t *testing.T) {
	diags := markerDiagsFor(t, mk("^(go|stop)$",
		"    RunTests -> Done when ctx.tool_marker = go or ctx.tool_marker = stop\n"))
	if len(diags) != 0 {
		t.Fatalf("or-routing must not warn (hasComplexRoute), got %v", diags)
	}
}

func TestDIP152_NotEqualCatchAllClean(t *testing.T) {
	diags := markerDiagsFor(t, mk("^(go|stop)$",
		"    RunTests -> Done on go\n    RunTests -> Done when ctx.tool_marker != go\n"))
	if len(diags) != 0 {
		t.Fatalf("!= catch-all must not warn, got %v", diags)
	}
}

func TestDIP152_NonEnumerableClean(t *testing.T) {
	diags := markerDiagsFor(t, mk(`.*fail.*`, "    RunTests -> Done on x\n"))
	if len(diags) != 0 {
		t.Fatalf("non-enumerable regex must not warn, got %v", diags)
	}
}

func TestDIP152_EmptyBranchClean(t *testing.T) {
	diags := markerDiagsFor(t, mk("^(go|stop|)$", "    RunTests -> Done on go\n"))
	if len(diags) != 0 {
		t.Fatalf("empty-branch regex is non-enumerable, must not warn, got %v", diags)
	}
}
