package validator

import (
	"strings"
	"testing"
)

// modeAgentSrc builds a one-agent workflow. extra lines are appended verbatim
// under the agent (callers control quoting / spelling of the mode value).
func modeAgentSrc(extra ...string) string {
	src := "workflow X\n  start: A\n  exit: A\n\n  agent A\n    prompt: \"x\"\n"
	for _, l := range extra {
		src += "    " + l + "\n"
	}
	return src
}

// modeBranchSrc builds a parallel workflow whose single branch carries the
// given branch lines and whose target agent carries agentExtra lines.
func modeBranchSrc(agentExtra []string, branchLines ...string) string {
	src := "workflow X\n  start: split\n  exit: join\n\n  agent a\n    prompt: \"a\"\n"
	for _, l := range agentExtra {
		src += "    " + l + "\n"
	}
	src += "\n  parallel split\n    branch: a\n"
	for _, l := range branchLines {
		src += "      " + l + "\n"
	}
	src += "\n  fan_in join <- a\n\n  edges\n    split -> a\n    a -> join\n"
	return src
}

func onlyCode(diags []Diagnostic, code string) []Diagnostic {
	var out []Diagnostic
	for _, d := range diags {
		if d.Code == code {
			out = append(out, d)
		}
	}
	return out
}

// --- DIP163: value must be exactly require or prefer -----------------------

func TestLint_DIP163_CleanValues(t *testing.T) {
	for _, mode := range []string{"require", "prefer"} {
		t.Run(mode, func(t *testing.T) {
			diags := lintSrc(t, modeAgentSrc("writable_paths: workspace/**", "writable_paths_mode: "+mode))
			if hasCode(diags, DIP163) {
				t.Errorf("DIP163 must not fire on %q; got %v", mode, codes(diags))
			}
		})
	}
}

func TestLint_DIP163_AbsentIsClean(t *testing.T) {
	diags := lintSrc(t, modeAgentSrc("writable_paths: workspace/**"))
	if hasCode(diags, DIP163) {
		t.Errorf("DIP163 must not fire when the field is absent; got %v", codes(diags))
	}
}

func TestLint_DIP163_BadValues(t *testing.T) {
	cases := []struct {
		name, line, wantQuoted string
	}{
		{"capitalized", "writable_paths_mode: Prefer", `"Prefer"`},
		{"misspelling", "writable_paths_mode: preferred", `"preferred"`},
		{"quoted trailing space", `writable_paths_mode: "prefer "`, `"prefer "`},
		{"uppercase require", "writable_paths_mode: REQUIRE", `"REQUIRE"`},
		{"leading space quoted", `writable_paths_mode: " require"`, `" require"`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			diags := onlyCode(lintSrc(t, modeAgentSrc("writable_paths: workspace/**", tc.line)), DIP163)
			if len(diags) != 1 {
				t.Fatalf("want exactly one DIP163, got %d: %v", len(diags), diags)
			}
			d := diags[0]
			if d.Severity != SeverityError {
				t.Errorf("DIP163 severity = %v, want error", d.Severity)
			}
			if !strings.Contains(d.Message, `node "A"`) {
				t.Errorf("message should name the node; got %q", d.Message)
			}
			if !strings.Contains(d.Message, tc.wantQuoted) {
				t.Errorf("message should quote the offending value %s; got %q", tc.wantQuoted, d.Message)
			}
		})
	}
}

func TestLint_DIP163_BranchOverride(t *testing.T) {
	diags := onlyCode(lintSrc(t, modeBranchSrc(nil, "writable_paths: workspace/**", "writable_paths_mode: Prefer")), DIP163)
	if len(diags) != 1 {
		t.Fatalf("want exactly one DIP163 on the branch, got %d: %v", len(diags), diags)
	}
	msg := diags[0].Message
	if !strings.Contains(msg, `node "split"`) || !strings.Contains(msg, `branch "a"`) || !strings.Contains(msg, `"Prefer"`) {
		t.Errorf("branch DIP163 should name node, branch and quoted value; got %q", msg)
	}
	if diags[0].Severity != SeverityError {
		t.Errorf("branch DIP163 severity = %v, want error", diags[0].Severity)
	}
}

func TestLint_DIP163_BranchCleanValues(t *testing.T) {
	for _, mode := range []string{"require", "prefer"} {
		diags := lintSrc(t, modeBranchSrc(nil, "writable_paths: workspace/**", "writable_paths_mode: "+mode))
		if hasCode(diags, DIP163) {
			t.Errorf("branch DIP163 must not fire on %q; got %v", mode, codes(diags))
		}
	}
}

// --- DIP164: mode without a scope is inert ---------------------------------

func TestLint_DIP164_Agent(t *testing.T) {
	cases := []struct {
		name  string
		lines []string
		want  bool
	}{
		{"mode without paths", []string{"writable_paths_mode: prefer"}, true},
		{"require without paths", []string{"writable_paths_mode: require"}, true},
		{"mode with paths", []string{"writable_paths: workspace/**", "writable_paths_mode: prefer"}, false},
		{"paths only", []string{"writable_paths: workspace/**"}, false},
		{"neither", nil, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			diags := onlyCode(lintSrc(t, modeAgentSrc(tc.lines...)), DIP164)
			if got := len(diags) > 0; got != tc.want {
				t.Fatalf("DIP164 fired=%v, want %v; diags=%v", got, tc.want, diags)
			}
			if tc.want && diags[0].Severity != SeverityHint {
				t.Errorf("DIP164 severity = %v, want hint", diags[0].Severity)
			}
			if tc.want && !strings.Contains(diags[0].Message, `node "A"`) {
				t.Errorf("DIP164 should name the node; got %q", diags[0].Message)
			}
		})
	}
}

// For a branch override, "set without scope" means the branch declares a mode
// and NEITHER the branch NOR the target agent declares writable_paths (the
// branch inherits the target's paths when it declares none of its own).
func TestLint_DIP164_Branch(t *testing.T) {
	cases := []struct {
		name       string
		agentExtra []string
		branch     []string
		want       bool
	}{
		{"branch mode, no paths anywhere", nil, []string{"writable_paths_mode: prefer"}, true},
		{"branch mode, branch paths", nil, []string{"writable_paths: workspace/**", "writable_paths_mode: prefer"}, false},
		{"branch mode, target agent paths", []string{"writable_paths: workspace/**"}, []string{"writable_paths_mode: prefer"}, false},
		{"branch paths only", nil, []string{"writable_paths: workspace/**"}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			diags := onlyCode(lintSrc(t, modeBranchSrc(tc.agentExtra, tc.branch...)), DIP164)
			if got := len(diags) > 0; got != tc.want {
				t.Fatalf("DIP164 fired=%v, want %v; diags=%v", got, tc.want, diags)
			}
			if tc.want && (!strings.Contains(diags[0].Message, `node "split"`) || !strings.Contains(diags[0].Message, `branch "a"`)) {
				t.Errorf("branch DIP164 should name node and branch; got %q", diags[0].Message)
			}
		})
	}
}

// DIP164 must still fire on an invalid mode value (the author set *something*
// without a scope) — it is orthogonal to DIP163.
func TestLint_DIP164_FiresAlongsideDIP163(t *testing.T) {
	diags := lintSrc(t, modeAgentSrc("writable_paths_mode: Prefer"))
	if !hasCode(diags, DIP163) || !hasCode(diags, DIP164) {
		t.Errorf("want both DIP163 and DIP164; got %v", codes(diags))
	}
}

// --- DIP165: prefer runs UNJAILED without Landlock ---------------------------

func TestLint_DIP165_Agent(t *testing.T) {
	cases := []struct {
		name  string
		lines []string
		want  bool
	}{
		{"prefer", []string{"writable_paths: workspace/**", "writable_paths_mode: prefer"}, true},
		{"require", []string{"writable_paths: workspace/**", "writable_paths_mode: require"}, false},
		{"absent", []string{"writable_paths: workspace/**"}, false},
		{"Prefer is not prefer", []string{"writable_paths: workspace/**", "writable_paths_mode: Prefer"}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			diags := onlyCode(lintSrc(t, modeAgentSrc(tc.lines...)), DIP165)
			if got := len(diags) > 0; got != tc.want {
				t.Fatalf("DIP165 fired=%v, want %v; diags=%v", got, tc.want, diags)
			}
			if !tc.want {
				return
			}
			if len(diags) != 1 {
				t.Errorf("DIP165 should fire once per prefer declaration; got %d", len(diags))
			}
			if diags[0].Severity != SeverityHint {
				t.Errorf("DIP165 severity = %v, want hint", diags[0].Severity)
			}
			if !strings.Contains(diags[0].Message, "UNJAILED") || !strings.Contains(diags[0].Message, `node "A"`) {
				t.Errorf("DIP165 message must name the node and say UNJAILED; got %q", diags[0].Message)
			}
			if strings.Contains(strings.ToLower(diags[0].Help), "sandboxed") && !strings.Contains(diags[0].Help, "not") {
				t.Errorf("operator copy must not call a prefer node sandboxed; got %q", diags[0].Help)
			}
		})
	}
}

func TestLint_DIP165_Branch(t *testing.T) {
	diags := onlyCode(lintSrc(t, modeBranchSrc(nil, "writable_paths: workspace/**", "writable_paths_mode: prefer")), DIP165)
	if len(diags) != 1 {
		t.Fatalf("want exactly one branch DIP165, got %d: %v", len(diags), diags)
	}
	msg := diags[0].Message
	if !strings.Contains(msg, `node "split"`) || !strings.Contains(msg, `branch "a"`) || !strings.Contains(msg, "UNJAILED") {
		t.Errorf("branch DIP165 should name node, branch and say UNJAILED; got %q", msg)
	}
}

func TestLint_DIP165_BranchRequireIsClean(t *testing.T) {
	diags := lintSrc(t, modeBranchSrc(nil, "writable_paths: workspace/**", "writable_paths_mode: require"))
	if hasCode(diags, DIP165) {
		t.Errorf("branch require must not fire DIP165; got %v", codes(diags))
	}
}

// Agent prefer + branch prefer on the same target: one hint per declaration.
func TestLint_DIP165_OncePerDeclaration(t *testing.T) {
	src := modeBranchSrc([]string{"writable_paths: workspace/**", "writable_paths_mode: prefer"}, "writable_paths_mode: prefer")
	if n := len(onlyCode(lintSrc(t, src), DIP165)); n != 2 {
		t.Errorf("want 2 DIP165 (agent + branch), got %d", n)
	}
}

// Prefer is a hint, so the error-severity gate (which DIP163 shares with
// DIP155–DIP158) is not tripped by a well-formed prefer declaration.
func TestLint_WritablePathsMode_PreferHasNoErrors(t *testing.T) {
	for _, d := range lintSrc(t, modeAgentSrc("writable_paths: workspace/**", "writable_paths_mode: prefer")) {
		if d.Severity == SeverityError {
			t.Errorf("unexpected error-severity diagnostic on a clean prefer node: %v", d)
		}
	}
}

// Agent-level DIP164 must respect inherited scope: a block-form parallel branch
// that targets the agent and declares its own writable_paths inherits the
// agent's mode (tracker C1), so the agent's mode is not inert.
func TestLint_DIP164_AgentScopedByBranch(t *testing.T) {
	cases := []struct {
		name   string
		branch []string
		want   bool
	}{
		{"branch targeting agent declares paths", []string{"writable_paths: workspace/**"}, false},
		{"branch targeting agent declares no paths", []string{"model: claude-sonnet-4-6"}, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			diags := lintSrc(t, modeBranchSrc([]string{"writable_paths_mode: prefer"}, tc.branch...))
			if got := hasCode(diags, DIP164); got != tc.want {
				t.Errorf("agent DIP164 fired=%v, want %v; diags=%v", got, tc.want, codes(diags))
			}
			if n := len(onlyCode(diags, DIP165)); n != 1 {
				t.Errorf("DIP165 should be unaffected (want 1 for the agent prefer), got %d", n)
			}
		})
	}
}
