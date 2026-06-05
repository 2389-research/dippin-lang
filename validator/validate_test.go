package validator

import (
	"strings"
	"testing"

	"github.com/2389-research/dippin-lang/ir"
)

// --- Test fixtures ---

// minimalValidWorkflow returns a valid two-node workflow.
func minimalValidWorkflow() *ir.Workflow {
	return &ir.Workflow{
		Name:  "minimal",
		Start: "Begin",
		Exit:  "End",
		Nodes: []*ir.Node{
			{ID: "Begin", Kind: ir.NodeHuman, Config: ir.HumanConfig{Mode: "freeform"}},
			{ID: "End", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "Done."}},
		},
		Edges: []*ir.Edge{
			{From: "Begin", To: "End"},
		},
	}
}

// askAndExecuteWorkflow returns the canonical example with restart edges
// and parallel/fan_in.
func askAndExecuteWorkflow() *ir.Workflow {
	return &ir.Workflow{
		Name:  "ask_and_execute",
		Start: "AskUser",
		Exit:  "Done",
		Nodes: []*ir.Node{
			{ID: "AskUser", Kind: ir.NodeHuman, Config: ir.HumanConfig{Mode: "freeform"}},
			{ID: "Interpret", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "Plan."}},
			{ID: "ImplementFanOut", Kind: ir.NodeParallel, Config: ir.ParallelConfig{Targets: []string{"ImplementClaude", "ImplementCodex"}}},
			{ID: "ImplementClaude", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "Implement."}},
			{ID: "ImplementCodex", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "Implement."}},
			{ID: "ImplementJoin", Kind: ir.NodeFanIn, Config: ir.FanInConfig{Sources: []string{"ImplementClaude", "ImplementCodex"}}},
			{ID: "Validate", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "Review."}},
			{ID: "Approve", Kind: ir.NodeHuman, Config: ir.HumanConfig{Mode: "choice"}},
			{ID: "Done", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "Ship."}},
		},
		Edges: []*ir.Edge{
			{From: "AskUser", To: "Interpret"},
			{From: "Interpret", To: "ImplementFanOut"},
			{From: "ImplementFanOut", To: "ImplementClaude"},
			{From: "ImplementFanOut", To: "ImplementCodex"},
			{From: "ImplementClaude", To: "ImplementJoin"},
			{From: "ImplementCodex", To: "ImplementJoin"},
			{From: "ImplementJoin", To: "Validate"},
			{From: "Validate", To: "Approve", Condition: &ir.Condition{
				Raw:    "ctx.outcome = success",
				Parsed: ir.CondCompare{Variable: "ctx.outcome", Op: "=", Value: "success"},
			}},
			{From: "Validate", To: "Interpret", Label: "retry", Restart: true, Condition: &ir.Condition{
				Raw:    "ctx.outcome = fail",
				Parsed: ir.CondCompare{Variable: "ctx.outcome", Op: "=", Value: "fail"},
			}},
			{From: "Approve", To: "Done"},
		},
	}
}

// --- Table-driven tests ---

func TestValidate(t *testing.T) {
	tests := []struct {
		name       string
		workflow   *ir.Workflow
		wantCodes  []string // Expected diagnostic codes (empty = no diagnostics)
		wantNoDiag bool     // If true, expect zero diagnostics
	}{
		// --- Happy path ---
		{
			name:       "valid minimal workflow",
			workflow:   minimalValidWorkflow(),
			wantNoDiag: true,
		},
		{
			name:       "valid complex workflow with restart and parallel",
			workflow:   askAndExecuteWorkflow(),
			wantNoDiag: true,
		},
		{
			name: "valid workflow with restart back-edge (no cycle)",
			workflow: &ir.Workflow{
				Name:  "restart_loop",
				Start: "A",
				Exit:  "D",
				Nodes: []*ir.Node{
					{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "a"}},
					{ID: "B", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "b"}},
					{ID: "C", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "c"}},
					{ID: "D", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "d"}},
				},
				Edges: []*ir.Edge{
					{From: "A", To: "B"},
					{From: "B", To: "C"},
					{From: "C", To: "A", Restart: true},
					{From: "C", To: "D"},
				},
			},
			wantNoDiag: true,
		},
		{
			name: "valid parallel/fan_in pair with different order",
			workflow: &ir.Workflow{
				Name:  "parallel_ok",
				Start: "Start",
				Exit:  "End",
				Nodes: []*ir.Node{
					{ID: "Start", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "go"}},
					{ID: "Fork", Kind: ir.NodeParallel, Config: ir.ParallelConfig{Targets: []string{"W1", "W2"}}},
					{ID: "W1", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "w1"}},
					{ID: "W2", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "w2"}},
					{ID: "Join", Kind: ir.NodeFanIn, Config: ir.FanInConfig{Sources: []string{"W2", "W1"}}}, // order differs
					{ID: "End", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "done"}},
				},
				Edges: []*ir.Edge{
					{From: "Start", To: "Fork"},
					{From: "Fork", To: "W1"},
					{From: "Fork", To: "W2"},
					{From: "W1", To: "Join"},
					{From: "W2", To: "Join"},
					{From: "Join", To: "End"},
				},
			},
			wantNoDiag: true,
		},

		// --- Error cases: one diagnostic each ---
		{
			name: "DIP001: start node empty",
			workflow: &ir.Workflow{
				Name:  "no_start",
				Start: "",
				Exit:  "End",
				Nodes: []*ir.Node{
					{ID: "End", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "done"}},
				},
			},
			wantCodes: []string{DIP001},
		},
		{
			name: "DIP001: start node declared but missing from nodes",
			workflow: &ir.Workflow{
				Name:  "bad_start",
				Start: "Nonexistent",
				Exit:  "End",
				Nodes: []*ir.Node{
					{ID: "End", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "done"}},
				},
			},
			wantCodes: []string{DIP001},
		},
		{
			name: "DIP002: exit node missing from nodes",
			workflow: &ir.Workflow{
				Name:  "bad_exit",
				Start: "Begin",
				Exit:  "Nonexistent",
				Nodes: []*ir.Node{
					{ID: "Begin", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "go"}},
				},
			},
			wantCodes: []string{DIP002},
		},
		{
			name: "DIP003: dangling edge target",
			workflow: &ir.Workflow{
				Name:  "dangling",
				Start: "A",
				Exit:  "A",
				Nodes: []*ir.Node{
					{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "a"}},
				},
				Edges: []*ir.Edge{
					{From: "A", To: "Nonexistent"},
				},
			},
			wantCodes: []string{DIP003},
		},
		{
			name: "DIP003: fuzzy match suggests similar node",
			workflow: &ir.Workflow{
				Name:  "fuzzy",
				Start: "Interpret",
				Exit:  "Interpret",
				Nodes: []*ir.Node{
					{ID: "Interpret", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "a"}},
				},
				Edges: []*ir.Edge{
					{From: "Interpret", To: "Interpet"}, // typo: missing 'r'
				},
			},
			wantCodes: []string{DIP003},
		},
		{
			name: "DIP004: unreachable node",
			workflow: &ir.Workflow{
				Name:  "unreachable",
				Start: "A",
				Exit:  "B",
				Nodes: []*ir.Node{
					{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "a"}},
					{ID: "B", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "b"}},
					{ID: "Island", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "island"}},
				},
				Edges: []*ir.Edge{
					{From: "A", To: "B"},
				},
			},
			wantCodes: []string{DIP004},
		},
		{
			name: "DIP005: unconditional cycle",
			workflow: &ir.Workflow{
				Name:  "cycle",
				Start: "A",
				Exit:  "D",
				Nodes: []*ir.Node{
					{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "a"}},
					{ID: "B", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "b"}},
					{ID: "C", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "c"}},
					{ID: "D", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "d"}},
				},
				Edges: []*ir.Edge{
					{From: "A", To: "B"},
					{From: "B", To: "C"},
					{From: "C", To: "A"}, // cycle: not restart
					{From: "C", To: "D"},
				},
			},
			wantCodes: []string{DIP005},
		},
		{
			name: "DIP006: exit has outgoing edge",
			workflow: &ir.Workflow{
				Name:  "exit_outgoing",
				Start: "A",
				Exit:  "B",
				Nodes: []*ir.Node{
					{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "a"}},
					{ID: "B", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "b"}},
				},
				Edges: []*ir.Edge{
					{From: "A", To: "B"},
					{From: "B", To: "A"},
				},
			},
			wantCodes: []string{DIP006},
		},
		{
			name: "DIP007: orphaned parallel node",
			workflow: &ir.Workflow{
				Name:  "orphan_parallel",
				Start: "Start",
				Exit:  "End",
				Nodes: []*ir.Node{
					{ID: "Start", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "go"}},
					{ID: "Fork", Kind: ir.NodeParallel, Config: ir.ParallelConfig{Targets: []string{"X", "Y"}}},
					{ID: "X", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "x"}},
					{ID: "Y", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "y"}},
					{ID: "End", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "done"}},
				},
				Edges: []*ir.Edge{
					{From: "Start", To: "Fork"},
					{From: "Fork", To: "X"},
					{From: "Fork", To: "Y"},
					{From: "X", To: "End"},
					{From: "Y", To: "End"},
				},
			},
			wantCodes: []string{DIP007},
		},
		{
			name: "DIP007: orphaned fan_in node",
			workflow: &ir.Workflow{
				Name:  "orphan_fanin",
				Start: "Start",
				Exit:  "End",
				Nodes: []*ir.Node{
					{ID: "Start", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "go"}},
					{ID: "X", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "x"}},
					{ID: "Y", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "y"}},
					{ID: "Join", Kind: ir.NodeFanIn, Config: ir.FanInConfig{Sources: []string{"X", "Y"}}},
					{ID: "End", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "done"}},
				},
				Edges: []*ir.Edge{
					{From: "Start", To: "X"},
					{From: "Start", To: "Y"},
					{From: "X", To: "Join"},
					{From: "Y", To: "Join"},
					{From: "Join", To: "End"},
				},
			},
			wantCodes: []string{DIP007},
		},
		{
			name: "DIP008: duplicate node ID",
			workflow: &ir.Workflow{
				Name:  "dup_node",
				Start: "A",
				Exit:  "A",
				Nodes: []*ir.Node{
					{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "first"}, Source: ir.SourceLocation{File: "test.dip", Line: 1}},
					{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "second"}, Source: ir.SourceLocation{File: "test.dip", Line: 5}},
				},
			},
			wantCodes: []string{DIP008},
		},
		{
			name: "DIP009: duplicate unconditional edge",
			workflow: &ir.Workflow{
				Name:  "dup_edge",
				Start: "A",
				Exit:  "B",
				Nodes: []*ir.Node{
					{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "a"}},
					{ID: "B", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "b"}},
				},
				Edges: []*ir.Edge{
					{From: "A", To: "B"},
					{From: "A", To: "B"},
				},
			},
			wantCodes: []string{DIP009},
		},

		// --- Edge cases ---
		{
			name: "multiple errors at once",
			workflow: &ir.Workflow{
				Name:  "multi_error",
				Start: "Missing",
				Exit:  "End",
				Nodes: []*ir.Node{
					{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "a"}, Source: ir.SourceLocation{File: "test.dip", Line: 1}},
					{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "dup"}, Source: ir.SourceLocation{File: "test.dip", Line: 5}},
					{ID: "End", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "end"}},
				},
				Edges: []*ir.Edge{
					{From: "A", To: "Ghost"},
				},
			},
			wantCodes: []string{DIP008, DIP001, DIP003},
		},
		{
			name:      "empty workflow",
			workflow:  &ir.Workflow{},
			wantCodes: []string{DIP001, DIP002},
		},
		{
			name: "DIP003: both endpoints dangling",
			workflow: &ir.Workflow{
				Name:  "both_dangling",
				Start: "Real",
				Exit:  "Real",
				Nodes: []*ir.Node{
					{ID: "Real", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "ok"}},
				},
				Edges: []*ir.Edge{
					{From: "Ghost1", To: "Ghost2"},
				},
			},
			wantCodes: []string{DIP003, DIP003},
		},
		{
			name: "DIP009: same endpoints different conditions = NOT duplicate",
			workflow: &ir.Workflow{
				Name:  "cond_branches",
				Start: "A",
				Exit:  "B",
				Nodes: []*ir.Node{
					{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "a"}},
					{ID: "B", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "b"}},
				},
				Edges: []*ir.Edge{
					{From: "A", To: "B", Condition: &ir.Condition{Raw: "ctx.x = 1", Parsed: ir.CondCompare{Variable: "ctx.x", Op: "=", Value: "1"}}},
					{From: "A", To: "B", Condition: &ir.Condition{Raw: "ctx.x = 2", Parsed: ir.CondCompare{Variable: "ctx.x", Op: "=", Value: "2"}}},
				},
			},
			wantNoDiag: true,
		},
		{
			name: "DIP005: cycle through restart edge is OK (duplicate of happy path for clarity)",
			workflow: &ir.Workflow{
				Name:  "restart_ok",
				Start: "A",
				Exit:  "D",
				Nodes: []*ir.Node{
					{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "a"}},
					{ID: "B", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "b"}},
					{ID: "C", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "c"}},
					{ID: "D", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "d"}},
				},
				Edges: []*ir.Edge{
					{From: "A", To: "B"},
					{From: "B", To: "C"},
					{From: "C", To: "A", Restart: true},
					{From: "C", To: "D"},
				},
			},
			wantNoDiag: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Validate(tt.workflow)

			if tt.wantNoDiag {
				if len(result.Diagnostics) != 0 {
					t.Errorf("expected no diagnostics, got %d:", len(result.Diagnostics))
					for _, d := range result.Diagnostics {
						t.Errorf("  %s", d.String())
					}
				}
				return
			}

			if tt.wantCodes != nil {
				gotCodes := make([]string, len(result.Diagnostics))
				for i, d := range result.Diagnostics {
					gotCodes[i] = d.Code
				}

				// Check that all expected codes are present (order-insensitive).
				wantCount := make(map[string]int)
				for _, c := range tt.wantCodes {
					wantCount[c]++
				}
				gotCount := make(map[string]int)
				for _, c := range gotCodes {
					gotCount[c]++
				}
				for code, want := range wantCount {
					if got := gotCount[code]; got < want {
						t.Errorf("expected at least %d %s diagnostic(s), got %d. All codes: %v", want, code, got, gotCodes)
					}
				}
			}
		})
	}
}

func TestDIP003FuzzyMatchHelp(t *testing.T) {
	w := &ir.Workflow{
		Name:  "fuzzy",
		Start: "Interpret",
		Exit:  "Interpret",
		Nodes: []*ir.Node{
			{ID: "Interpret", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "a"}},
		},
		Edges: []*ir.Edge{
			{From: "Interpret", To: "Interpet"}, // typo: missing 'r'
		},
	}

	result := Validate(w)

	var found bool
	for _, d := range result.Diagnostics {
		if d.Code == DIP003 {
			found = true
			if !strings.Contains(d.Help, `"Interpret"`) {
				t.Errorf("DIP003 help = %q, expected it to contain '\"Interpret\"'", d.Help)
			}
			if !strings.Contains(d.Help, "did you mean") {
				t.Errorf("DIP003 help = %q, expected 'did you mean'", d.Help)
			}
		}
	}
	if !found {
		t.Error("expected DIP003 diagnostic for fuzzy match test")
	}
}

func TestDIP004StartNodeIsReachable(t *testing.T) {
	w := &ir.Workflow{
		Name:  "start_reachable",
		Start: "A",
		Exit:  "B",
		Nodes: []*ir.Node{
			{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "a"}},
			{ID: "B", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "b"}},
			{ID: "Island", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "island"}},
		},
		Edges: []*ir.Edge{
			{From: "A", To: "B"},
		},
	}

	result := Validate(w)

	// Only "Island" should be unreachable, not A or B.
	dip004Count := 0
	for _, d := range result.Diagnostics {
		if d.Code == DIP004 {
			dip004Count++
			if !strings.Contains(d.Message, `"Island"`) {
				t.Errorf("expected DIP004 to report Island, got: %s", d.Message)
			}
		}
	}
	if dip004Count != 1 {
		t.Errorf("expected exactly 1 DIP004 diagnostic (for Island), got %d", dip004Count)
	}
}

func TestDiagnosticFormatting(t *testing.T) {
	d := Diagnostic{
		Code:     DIP003,
		Severity: SeverityError,
		Message:  `edge references unknown node "InterpretX"`,
		Location: ir.SourceLocation{File: "pipeline.dip", Line: 45, Column: 5},
		Help:     `did you mean "Interpret"?`,
	}

	s := d.String()

	if !strings.Contains(s, "error[DIP003]") {
		t.Errorf("diagnostic string missing 'error[DIP003]', got:\n%s", s)
	}
	if !strings.Contains(s, "pipeline.dip:45:5") {
		t.Errorf("diagnostic string missing location, got:\n%s", s)
	}
	if !strings.Contains(s, `= help: did you mean "Interpret"?`) {
		t.Errorf("diagnostic string missing help, got:\n%s", s)
	}
}

func TestDiagnosticFormattingNoFile(t *testing.T) {
	d := Diagnostic{
		Code:     DIP001,
		Severity: SeverityError,
		Message:  "workflow has no start node declared",
	}

	s := d.String()
	if !strings.Contains(s, "<unknown>:0:0") {
		t.Errorf("diagnostic string should show <unknown> for missing file, got:\n%s", s)
	}
}

func TestDiagnosticFormattingWithFix(t *testing.T) {
	d := Diagnostic{
		Code:     DIP003,
		Severity: SeverityError,
		Message:  `unknown node "Foo"`,
		Fix:      `rename to "Bar"`,
	}

	s := d.String()
	if !strings.Contains(s, `= fix: rename to "Bar"`) {
		t.Errorf("diagnostic string missing fix, got:\n%s", s)
	}
}

func TestResultErrors(t *testing.T) {
	r := Result{
		Diagnostics: []Diagnostic{
			{Code: DIP001, Severity: SeverityError, Message: "e1"},
			{Code: "INFO", Severity: SeverityInfo, Message: "i1"},
			{Code: DIP002, Severity: SeverityError, Message: "e2"},
		},
	}

	errs := r.Errors()
	if len(errs) != 2 {
		t.Fatalf("Errors() returned %d, want 2", len(errs))
	}
	if errs[0].Code != DIP001 || errs[1].Code != DIP002 {
		t.Errorf("Errors() = [%s, %s], want [DIP001, DIP002]", errs[0].Code, errs[1].Code)
	}
}

func TestResultHasErrors(t *testing.T) {
	tests := []struct {
		name string
		r    Result
		want bool
	}{
		{
			name: "no diagnostics",
			r:    Result{},
			want: false,
		},
		{
			name: "only info",
			r:    Result{Diagnostics: []Diagnostic{{Severity: SeverityInfo}}},
			want: false,
		},
		{
			name: "has error",
			r:    Result{Diagnostics: []Diagnostic{{Severity: SeverityError}}},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.r.HasErrors(); got != tt.want {
				t.Errorf("HasErrors() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSeverityString(t *testing.T) {
	tests := []struct {
		s    Severity
		want string
	}{
		{SeverityError, "error"},
		{SeverityWarning, "warning"},
		{SeverityInfo, "info"},
		{SeverityHint, "hint"},
		{Severity(99), "unknown"},
	}
	for _, tt := range tests {
		if got := tt.s.String(); got != tt.want {
			t.Errorf("Severity(%d).String() = %q, want %q", tt.s, got, tt.want)
		}
	}
}

func TestLevenshtein(t *testing.T) {
	tests := []struct {
		a, b string
		want int
	}{
		{"", "", 0},
		{"abc", "", 3},
		{"", "abc", 3},
		{"abc", "abc", 0},
		{"abc", "abd", 1},
		{"abc", "abcd", 1},
		{"Interpret", "Interpet", 1},   // missing 'r'
		{"Interpret", "InterpretX", 1}, // extra char
		{"Interpret", "Intepret", 1},   // transposition-ish
		{"abc", "xyz", 3},              // all different
	}
	for _, tt := range tests {
		got := levenshtein(tt.a, tt.b)
		if got != tt.want {
			t.Errorf("levenshtein(%q, %q) = %d, want %d", tt.a, tt.b, got, tt.want)
		}
	}
}

func TestCodeDescriptionCoverage(t *testing.T) {
	codes := []string{DIP001, DIP002, DIP003, DIP004, DIP005, DIP006, DIP007, DIP008, DIP009}
	for _, c := range codes {
		if desc, ok := CodeDescription[c]; !ok || desc == "" {
			t.Errorf("CodeDescription[%q] is missing or empty", c)
		}
	}
}

func TestDIP005CyclePathIncluded(t *testing.T) {
	w := &ir.Workflow{
		Name:  "cycle_path",
		Start: "A",
		Exit:  "D",
		Nodes: []*ir.Node{
			{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "a"}},
			{ID: "B", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "b"}},
			{ID: "C", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "c"}},
			{ID: "D", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "d"}},
		},
		Edges: []*ir.Edge{
			{From: "A", To: "B"},
			{From: "B", To: "C"},
			{From: "C", To: "A"}, // cycle
			{From: "C", To: "D"},
		},
	}

	result := Validate(w)

	var found bool
	for _, d := range result.Diagnostics {
		if d.Code == DIP005 {
			found = true
			if !strings.Contains(d.Message, "→") {
				t.Errorf("DIP005 message should include cycle path with arrows, got: %s", d.Message)
			}
			mentionsA := strings.Contains(d.Message, "A")
			mentionsB := strings.Contains(d.Message, "B")
			mentionsC := strings.Contains(d.Message, "C")
			if !(mentionsA && mentionsB && mentionsC) {
				t.Errorf("DIP005 message should mention cycle nodes A, B, C, got: %s", d.Message)
			}
		}
	}
	if !found {
		t.Error("expected DIP005 diagnostic for cycle test")
	}
}

func TestDIP006MultipleOutgoing(t *testing.T) {
	w := &ir.Workflow{
		Name:  "exit_multi_outgoing",
		Start: "A",
		Exit:  "B",
		Nodes: []*ir.Node{
			{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "a"}},
			{ID: "B", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "b"}},
			{ID: "C", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "c"}},
		},
		Edges: []*ir.Edge{
			{From: "A", To: "B"},
			{From: "B", To: "A"},
			{From: "B", To: "C"},
		},
	}

	result := Validate(w)

	count := 0
	for _, d := range result.Diagnostics {
		if d.Code == DIP006 {
			count++
		}
	}
	if count != 2 {
		t.Errorf("expected 2 DIP006 diagnostics for two outgoing edges from exit, got %d", count)
	}
}

func TestOnFailureUnknownNode(t *testing.T) {
	w := &ir.Workflow{
		Name: "t", Start: "A", Exit: "A",
		Defaults: ir.WorkflowDefaults{OnFailure: "Nope"},
		Nodes:    []*ir.Node{{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "go"}}},
	}
	res := Validate(w)
	if !hasCode(res.Diagnostics, DIP003) {
		t.Fatalf("expected DIP003 for unknown on_failure target, got %+v", res.Diagnostics)
	}
}

func TestOnFailureKnownNode(t *testing.T) {
	w := &ir.Workflow{
		Name: "t", Start: "A", Exit: "A",
		Defaults: ir.WorkflowDefaults{OnFailure: "A"},
		Nodes:    []*ir.Node{{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "go"}}},
	}
	res := Validate(w)
	for _, d := range res.Diagnostics {
		if d.Code == DIP003 {
			t.Fatalf("unexpected DIP003 for valid on_failure target: %+v", d)
		}
	}
}

func TestOnFailureTargetIsReachable(t *testing.T) {
	w := &ir.Workflow{
		Name: "t", Start: "A", Exit: "Done",
		Defaults: ir.WorkflowDefaults{OnFailure: "Rescue"},
		Nodes: []*ir.Node{
			{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "go"}},
			{ID: "Rescue", Kind: ir.NodeHuman, Config: ir.HumanConfig{Mode: "freeform"}},
			{ID: "Done", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "end"}},
		},
		Edges: []*ir.Edge{
			{From: "A", To: "Done"},
			{From: "Rescue", To: "Done"},
		},
	}
	res := Validate(w)
	for _, d := range res.Diagnostics {
		if d.Code == DIP004 {
			t.Fatalf("Rescue should be reachable via on_failure, got DIP004: %+v", d)
		}
	}
}
