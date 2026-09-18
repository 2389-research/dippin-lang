package validator

import (
	"testing"

	"github.com/2389-research/dippin-lang/ir"
)

// --- DIP140: params re-enables tools that tool_access strips ---

func TestLint_DIP140_ParamsReenablesToolsUnderToolAccess(t *testing.T) {
	for _, key := range []string{"allowed_tools", "disallowed_tools", "tool_choice", "permission_mode"} {
		t.Run(key, func(t *testing.T) {
			w := &ir.Workflow{
				Name:  "dip140",
				Start: "A",
				Exit:  "A",
				Nodes: []*ir.Node{
					{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{
						Prompt:     "Summarize.",
						ToolAccess: "none",
						Params:     map[string]string{key: "Bash"},
					}},
				},
			}
			res := Lint(w)
			if !hasCode(res.Diagnostics, DIP140) {
				t.Errorf("expected DIP140 for params key %q under tool_access; got: %v", key, codes(res.Diagnostics))
			}
		})
	}
}

func TestLint_DIP140_Severity(t *testing.T) {
	w := &ir.Workflow{
		Name:  "dip140_sev",
		Start: "A",
		Exit:  "A",
		Nodes: []*ir.Node{
			{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{
				Prompt:     "Summarize.",
				ToolAccess: "none",
				Params:     map[string]string{"allowed_tools": "Bash"},
			}},
		},
	}
	res := Lint(w)
	for _, d := range res.Diagnostics {
		if d.Code == DIP140 && d.Severity != SeverityWarning {
			t.Errorf("DIP140 should be warning, got %s", d.Severity)
		}
	}
}

func TestLint_DIP140_NoToolAccess_NoFire(t *testing.T) {
	// Without tool_access set, allowed_tools is a legitimate backend:
	// claude-code knob — the runtime honors it, so no bypass and no warning.
	w := &ir.Workflow{
		Name:  "dip140_no_access",
		Start: "A",
		Exit:  "A",
		Nodes: []*ir.Node{
			{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{
				Prompt: "Work.",
				Params: map[string]string{"allowed_tools": "Bash"},
			}},
		},
	}
	res := Lint(w)
	if hasCode(res.Diagnostics, DIP140) {
		t.Errorf("DIP140 should not fire without tool_access; got: %v", codes(res.Diagnostics))
	}
}

func TestLint_DIP140_NonToolParams_NoFire(t *testing.T) {
	// tool_access set, but params has no tool-re-enabling key.
	w := &ir.Workflow{
		Name:  "dip140_other_params",
		Start: "A",
		Exit:  "A",
		Nodes: []*ir.Node{
			{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{
				Prompt:     "Summarize.",
				ToolAccess: "none",
				Params:     map[string]string{"temperature": "0.2"},
			}},
		},
	}
	res := Lint(w)
	if hasCode(res.Diagnostics, DIP140) {
		t.Errorf("DIP140 should not fire for non-tool params; got: %v", codes(res.Diagnostics))
	}
}

// --- DIP133 now also covers params: { tool_access: ... } (shadow of the
// first-class tool_access field) ---

func TestLint_DIP133_ToolAccessShadow(t *testing.T) {
	w := &ir.Workflow{
		Name:  "dip133_tool_access_shadow",
		Start: "A",
		Exit:  "A",
		Nodes: []*ir.Node{
			{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{
				Prompt: "Hello.",
				Params: map[string]string{"tool_access": "none"},
			}},
		},
	}
	res := Lint(w)
	if !hasCode(res.Diagnostics, DIP133) {
		t.Errorf("expected DIP133 for params tool_access shadow; got: %v", codes(res.Diagnostics))
	}
}

// params: { writable_paths_mode: ... } shadows the typed field (#307). Before the
// typed field existed the key rode params: silently; now the params spelling
// must surface as DIP133 so a stale `params: writable_paths_mode: Prefer` can
// never lint clean while bypassing DIP163's exact-match check.
func TestLint_DIP133_WritablePathsModeShadow(t *testing.T) {
	src := `workflow X
  start: A
  exit: A

  agent A
    prompt: "x"
    writable_paths: workspace/**
    params:
      writable_paths_mode: Prefer
`
	diags := lintSrc(t, src)
	if !hasCodeMentioning(diags, DIP133, "writable_paths_mode") {
		t.Errorf("expected DIP133 naming writable_paths_mode for the params shadow; got: %v", codes(diags))
	}
}

func TestLint_DIP133_WritablePathsShadow(t *testing.T) {
	src := `workflow X
  start: A
  exit: A

  agent A
    prompt: "x"
    params:
      writable_paths: workspace/**
`
	if !hasCodeMentioning(lintSrc(t, src), DIP133, "writable_paths") {
		t.Errorf("expected DIP133 for params writable_paths shadow; got: %v", codes(lintSrc(t, src)))
	}
}
