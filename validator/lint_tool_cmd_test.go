package validator

import (
	"testing"

	"github.com/2389-research/dippin-lang/ir"
)

func TestLintToolSyntax_Valid(t *testing.T) {
	w := &ir.Workflow{
		Name: "test", Start: "T", Exit: "T",
		Nodes: []*ir.Node{
			{ID: "T", Kind: ir.NodeTool, Config: ir.ToolConfig{
				Command: "set -eu\necho hello",
			}},
		},
	}
	diags := lintToolSyntax(w)
	if len(diags) != 0 {
		t.Errorf("expected no DIP123, got %d: %v", len(diags), diags)
	}
}

func TestLintToolSyntax_Error(t *testing.T) {
	w := &ir.Workflow{
		Name: "test", Start: "T", Exit: "T",
		Nodes: []*ir.Node{
			{ID: "T", Kind: ir.NodeTool, Config: ir.ToolConfig{
				Command: "echo \"unclosed",
			}},
		},
	}
	diags := lintToolSyntax(w)
	if len(diags) != 1 {
		t.Fatalf("expected 1 DIP123, got %d", len(diags))
	}
	if diags[0].Code != DIP123 {
		t.Errorf("expected DIP123, got %s", diags[0].Code)
	}
}

func TestLintToolSyntax_EmptyCommand(t *testing.T) {
	w := &ir.Workflow{
		Name: "test", Start: "T", Exit: "T",
		Nodes: []*ir.Node{
			{ID: "T", Kind: ir.NodeTool, Config: ir.ToolConfig{Command: ""}},
		},
	}
	diags := lintToolSyntax(w)
	if len(diags) != 0 {
		t.Errorf("expected no diagnostics for empty command, got %d", len(diags))
	}
}

func TestLintToolCtxVars_Found(t *testing.T) {
	w := &ir.Workflow{
		Name: "test", Start: "T", Exit: "T",
		Nodes: []*ir.Node{
			{ID: "T", Kind: ir.NodeTool, Config: ir.ToolConfig{
				Command: "curl ${ctx.api_url}/endpoint",
			}},
		},
	}
	diags := lintToolCtxVars(w)
	if len(diags) != 1 {
		t.Fatalf("expected 1 DIP124, got %d", len(diags))
	}
	if diags[0].Code != DIP124 {
		t.Errorf("expected DIP124, got %s", diags[0].Code)
	}
}

func TestLintToolCtxVars_Multiple(t *testing.T) {
	w := &ir.Workflow{
		Name: "test", Start: "T", Exit: "T",
		Nodes: []*ir.Node{
			{ID: "T", Kind: ir.NodeTool, Config: ir.ToolConfig{
				Command: "echo ${ctx.a} ${ctx.b}",
			}},
		},
	}
	diags := lintToolCtxVars(w)
	if len(diags) != 2 {
		t.Errorf("expected 2 DIP124, got %d", len(diags))
	}
}

func TestLintToolCtxVars_None(t *testing.T) {
	w := &ir.Workflow{
		Name: "test", Start: "T", Exit: "T",
		Nodes: []*ir.Node{
			{ID: "T", Kind: ir.NodeTool, Config: ir.ToolConfig{
				Command: "echo hello $HOME",
			}},
		},
	}
	diags := lintToolCtxVars(w)
	if len(diags) != 0 {
		t.Errorf("expected no DIP124, got %d", len(diags))
	}
}

func TestLintToolBinary_Found(t *testing.T) {
	w := &ir.Workflow{
		Name: "test", Start: "T", Exit: "T",
		Nodes: []*ir.Node{
			{ID: "T", Kind: ir.NodeTool, Config: ir.ToolConfig{
				Command: "set -eu\nls -la",
			}},
		},
	}
	diags := lintToolBinary(w)
	if len(diags) != 0 {
		t.Errorf("expected no DIP125 for ls, got %d", len(diags))
	}
}

func TestLintToolBinary_NotFound(t *testing.T) {
	w := &ir.Workflow{
		Name: "test", Start: "T", Exit: "T",
		Nodes: []*ir.Node{
			{ID: "T", Kind: ir.NodeTool, Config: ir.ToolConfig{
				Command: "set -eu\nnonexistent_binary_xyz --flag",
			}},
		},
	}
	diags := lintToolBinary(w)
	if len(diags) != 1 {
		t.Fatalf("expected 1 DIP125, got %d", len(diags))
	}
	if diags[0].Code != DIP125 {
		t.Errorf("expected DIP125, got %s", diags[0].Code)
	}
}

func TestLintToolBinary_Builtin(t *testing.T) {
	w := &ir.Workflow{
		Name: "test", Start: "T", Exit: "T",
		Nodes: []*ir.Node{
			{ID: "T", Kind: ir.NodeTool, Config: ir.ToolConfig{
				Command: "echo hello",
			}},
		},
	}
	diags := lintToolBinary(w)
	if len(diags) != 0 {
		t.Errorf("expected no DIP125 for shell builtin, got %d", len(diags))
	}
}

func TestLintToolBinary_SkipsPreamble(t *testing.T) {
	w := &ir.Workflow{
		Name: "test", Start: "T", Exit: "T",
		Nodes: []*ir.Node{
			{ID: "T", Kind: ir.NodeTool, Config: ir.ToolConfig{
				Command: "set -eu\ncd /tmp\nls -la",
			}},
		},
	}
	diags := lintToolBinary(w)
	if len(diags) != 0 {
		t.Errorf("expected no DIP125, preamble should be skipped, got %d", len(diags))
	}
}

func TestExtractBinary(t *testing.T) {
	tests := []struct {
		name string
		cmd  string
		want string
	}{
		{"builtin_echo", "echo hello", ""},
		{"preamble_then_ls", "set -eu\nls -la", "ls"},
		{"cd_then_git", "set -eu\ncd /tmp\ngit status", "git"},
		{"comment_then_curl", "# comment\nset -eu\ncurl http://x", "curl"},
		{"empty", "", ""},
		{"var_assign_then_echo", "COUNTER='.ai/count.txt'\necho done", ""},
		{"var_assign_then_printf", "count=0\nprintf '%s' $count", ""},
		{"pure_assignment", "FOO=bar", ""},
		{"multiple_assignments", "FOO=bar\nBAZ=qux", ""},
		{"inline_assign_with_cmd", "FOO=bar ls -la", "ls"},
		{"cmd_subst_in_assign", "count=$(cat file)\necho $count", "cat"},
		{"pipe", "cat file | grep pattern", "cat"},
		{"if_builtins_only", "if true; then echo yes; fi", ""},
		{"mkdir_preamble", "mkdir -p .ai/cache\nshellcheck script.sh", "shellcheck"},
		{"mkdir_then_touch", "mkdir -p /tmp/out\ntouch /tmp/out/file", "touch"},
		{"command_v_query", "command -v shellcheck", ""},
		{"command_v_check", "if command -v shellcheck >/dev/null 2>&1; then shellcheck script.sh; fi", "shellcheck"},
		{"command_v_and", "command -v git && git status", "git"},
		{"command_exec", "command git status", "git"},
		{"command_p_exec", "command -p git status", "git"},
		{"heredoc", "cat <<'EOF'\nhello\nEOF", "cat"},
		{"arithmetic", "count=$((count + 1))\nprintf '%s' $count", ""},
		// issue #315: colon builtin must not be treated as a binary.
		{"colon_builtin", ": > out.log", ""},
		{"colon_then_ls", ": > out.log\nls -la", "ls"},
		// issue #315: "." / "source" before the first real command
		// makes the symbol space unknowable - skip DIP125.
		{"source_then_ls", ". ./lib.sh\nls -la", ""},
		{"dot_source_then_func", "set -eu\n: > out.log\n. ./lib.sh\nmy_func", ""},
		{"source_keyword_then_realbin", "source ./lib.sh\nrealbin", ""},
		// "."/"source" AFTER the first real command does not suppress it.
		{"ls_then_source", "ls x\n. ./lib.sh", "ls"},
		// issue #315: a name matching a FuncDecl in the body is a
		// shell function, not a PATH binary.
		{"func_decl_then_call", "my_func() { echo hi; }\nmy_func", ""},
		// FuncDecl detection is order-independent: bodyDefinesFunc walks
		// the whole file, so a call before its declaration is still caught.
		{"call_then_func_decl", "my_func\nmy_func() { echo; }", ""},
		// A "." inside an as-yet-uncalled function body still counts as
		// "before" in source order once that function is reached by the
		// walk, per extractBinary's lexical-order semantics.
		{"source_inside_func_then_realbin", "foo() { . lib.sh; }\nrealbin", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractBinary(tt.cmd)
			if got != tt.want {
				t.Errorf("extractBinary(%q) = %q, want %q", tt.cmd, got, tt.want)
			}
		})
	}
}

// TestExtractBinary_NewBuiltins exercises every builtin added by issue #315
// as a script's first command, confirming each is skipped in favor of the
// real binary that follows.
func TestExtractBinary_NewBuiltins(t *testing.T) {
	builtins := []string{
		":", "pwd", "umask", "type", "readonly", "alias", "getopts",
		"times", "ulimit", "hash", "kill", "jobs", "fg", "bg", "let",
		"typeset", "[[",
	}
	for _, b := range builtins {
		t.Run(b, func(t *testing.T) {
			cmd := b + " x\nrealbin"
			if b == "[[" {
				// [[ ... ]] is a compound test, not a simple command.
				cmd = "[[ -n x ]]\nrealbin"
			}
			got := extractBinary(cmd)
			if got != "realbin" {
				t.Errorf("extractBinary(%q) = %q, want %q", cmd, got, "realbin")
			}
		})
	}
}

func TestLintToolBinary_VariableAssignment(t *testing.T) {
	w := &ir.Workflow{
		Name: "test", Start: "T", Exit: "T",
		Nodes: []*ir.Node{
			{ID: "T", Kind: ir.NodeTool, Config: ir.ToolConfig{
				Command: "set -eu\nCOUNTER='.ai/ralph/iteration-count.txt'\ncount=0\n[ -f \"$COUNTER\" ] && count=$(cat \"$COUNTER\")\ncount=$((count + 1))\nprintf '%s' \"$count\" > \"$COUNTER\"\nprintf '%s' \"$count\"",
			}},
		},
	}
	diags := lintToolBinary(w)
	if len(diags) != 0 {
		t.Errorf("expected no DIP125 for variable assignments, got %d: %v", len(diags), diags)
	}
}

// TestLintToolBinary_Placeholder covers issue #305: a ${ns.key} placeholder
// anywhere in the command body must not derail extraction to the "set"
// argument or a whole assignment line. Each node here should extract the
// same binary the placeholder-free path would (echo, a builtin -> no hint).
func TestLintToolBinary_Placeholder(t *testing.T) {
	w := &ir.Workflow{
		Name: "probe", Start: "Plain", Exit: "NoSet",
		Nodes: []*ir.Node{
			{ID: "Plain", Kind: ir.NodeTool, Config: ir.ToolConfig{
				Command: "set -eu\necho plain",
			}},
			{ID: "WithVar", Kind: ir.NodeTool, Config: ir.ToolConfig{
				Command: "set -eu\nX=\"${graph.workflow_dir}/lib\"\necho x",
			}},
			{ID: "ParamsVar", Kind: ir.NodeTool, Config: ir.ToolConfig{
				Command: "set -eu\nX=\"${params.foo}\"\necho x",
			}},
			{ID: "NoSet", Kind: ir.NodeTool, Config: ir.ToolConfig{
				Command: "LIB=\"${graph.workflow_dir}/lib\"\necho x",
			}},
		},
	}
	diags := lintToolBinary(w)
	if len(diags) != 0 {
		t.Errorf("expected no DIP125 hints for placeholder-bearing commands, got %d: %v", len(diags), diags)
	}
}

// TestExtractBinary_Placeholder exercises extractBinary directly for the
// placeholder cases, plus a placeholder-as-binary-name case that must yield
// "" since the real command can't be known before expansion.
func TestExtractBinary_Placeholder(t *testing.T) {
	tests := []struct {
		name string
		cmd  string
		want string
	}{
		{"graph_var_then_echo", "set -eu\nX=\"${graph.workflow_dir}/lib\"\necho x", ""},
		{"params_var_then_echo", "set -eu\nX=\"${params.foo}\"\necho x", ""},
		{"no_set_assign_then_echo", "LIB=\"${graph.workflow_dir}/lib\"\necho x", ""},
		{"placeholder_is_binary", "${params.bin} --flag", ""},
		{"placeholder_concat_with_text", "${params.prefix}bin --flag", ""},
		{"placeholder_var_then_nonbuiltin", "set -eu\nX=\"${params.foo}\"\nls x", "ls"},
		{"plain_shell_var_then_realbin", "${TOOL} --flag\nrealbin x", "realbin"},
		{"plain_shell_var_default_with_cmd_subst", "value=${CACHE:-$(missing-helper)}\necho \"$value\"", "missing-helper"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractBinary(tt.cmd)
			if got != tt.want {
				t.Errorf("extractBinary(%q) = %q, want %q", tt.cmd, got, tt.want)
			}
		})
	}
}

func TestLintToolBinary_AgentNodeIgnored(t *testing.T) {
	w := &ir.Workflow{
		Name: "test", Start: "A", Exit: "A",
		Nodes: []*ir.Node{
			{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "go."}},
		},
	}
	diags := lintToolBinary(w)
	if len(diags) != 0 {
		t.Errorf("expected no DIP125 for agent node, got %d", len(diags))
	}
}
