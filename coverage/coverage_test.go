package coverage

import (
	"testing"

	"github.com/2389-research/dippin-lang/ir"
)

// makeEdge creates a simple unconditional edge.
func makeEdge(from, to string) *ir.Edge {
	return &ir.Edge{From: from, To: to}
}

// makeCondEdge creates an edge with a simple equality condition.
func makeCondEdge(from, to, variable, value string) *ir.Edge {
	return &ir.Edge{
		From: from,
		To:   to,
		Condition: &ir.Condition{
			Raw:    variable + " = " + value,
			Parsed: ir.CondCompare{Variable: variable, Op: "=", Value: value},
		},
	}
}

func TestCoveredToolNode(t *testing.T) {
	w := &ir.Workflow{
		Start: "start",
		Exit:  "end",
		Nodes: []*ir.Node{
			{ID: "start", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "go"}},
			{ID: "check", Kind: ir.NodeTool, Config: ir.ToolConfig{Command: "printf 'pass'"}},
			{ID: "end", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "done"}},
		},
		Edges: []*ir.Edge{
			makeEdge("start", "check"),
			makeCondEdge("check", "end", "ctx.result", "pass"),
		},
	}

	r := Analyze(w)
	cov := r.Nodes["check"]
	if cov.Status != "covered" {
		t.Errorf("expected status covered, got %s", cov.Status)
	}
	if len(cov.MissingEdges) != 0 {
		t.Errorf("expected no missing edges, got %v", cov.MissingEdges)
	}
}

func TestPartialToolNode(t *testing.T) {
	w := &ir.Workflow{
		Start: "start",
		Exit:  "end",
		Nodes: []*ir.Node{
			{ID: "start", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "go"}},
			{ID: "check", Kind: ir.NodeTool, Config: ir.ToolConfig{
				Command: "printf 'pass'\nprintf 'fail'",
			}},
			{ID: "end", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "done"}},
		},
		Edges: []*ir.Edge{
			makeEdge("start", "check"),
			makeCondEdge("check", "end", "ctx.result", "pass"),
		},
	}

	r := Analyze(w)
	cov := r.Nodes["check"]
	if cov.Status != "partial" {
		t.Errorf("expected status partial, got %s", cov.Status)
	}
	if len(cov.MissingEdges) != 1 || cov.MissingEdges[0] != "fail" {
		t.Errorf("expected missing edge [fail], got %v", cov.MissingEdges)
	}
}

func TestFallbackMakesCovered(t *testing.T) {
	w := &ir.Workflow{
		Start: "start",
		Exit:  "end",
		Nodes: []*ir.Node{
			{ID: "start", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "go"}},
			{ID: "check", Kind: ir.NodeTool, Config: ir.ToolConfig{
				Command: "printf 'pass'\nprintf 'fail'",
			}},
			{ID: "end", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "done"}},
		},
		Edges: []*ir.Edge{
			makeEdge("start", "check"),
			makeCondEdge("check", "end", "ctx.result", "pass"),
			makeEdge("check", "end"), // fallback
		},
	}

	r := Analyze(w)
	cov := r.Nodes["check"]
	if cov.Status != "covered" {
		t.Errorf("expected status covered with fallback, got %s", cov.Status)
	}
	if !cov.HasFallback {
		t.Error("expected HasFallback to be true")
	}
}

func TestUnknownOutputs(t *testing.T) {
	w := &ir.Workflow{
		Start: "start",
		Exit:  "end",
		Nodes: []*ir.Node{
			{ID: "start", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "go"}},
			{ID: "check", Kind: ir.NodeTool, Config: ir.ToolConfig{
				Command: "run-custom-binary --flag",
			}},
			{ID: "end", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "done"}},
		},
		Edges: []*ir.Edge{
			makeEdge("start", "check"),
			makeCondEdge("check", "end", "ctx.result", "ok"),
		},
	}

	r := Analyze(w)
	cov := r.Nodes["check"]
	if cov.Status != "unknown" {
		t.Errorf("expected status unknown, got %s", cov.Status)
	}
}

func TestNoConditionsStatus(t *testing.T) {
	w := &ir.Workflow{
		Start: "start",
		Exit:  "end",
		Nodes: []*ir.Node{
			{ID: "start", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "go"}},
			{ID: "check", Kind: ir.NodeTool, Config: ir.ToolConfig{
				Command: "printf 'done'",
			}},
			{ID: "end", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "done"}},
		},
		Edges: []*ir.Edge{
			makeEdge("start", "check"),
			makeEdge("check", "end"),
		},
	}

	r := Analyze(w)
	cov := r.Nodes["check"]
	if cov.Status != "no_conditions" {
		t.Errorf("expected status no_conditions, got %s", cov.Status)
	}
}

func TestUnreachableNode(t *testing.T) {
	w := &ir.Workflow{
		Start: "start",
		Exit:  "end",
		Nodes: []*ir.Node{
			{ID: "start", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "go"}},
			{ID: "end", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "done"}},
			{ID: "orphan", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "lost"}},
		},
		Edges: []*ir.Edge{
			makeEdge("start", "end"),
		},
	}

	r := Analyze(w)
	if r.Reachability.ReachableNodes != 2 {
		t.Errorf("expected 2 reachable nodes, got %d", r.Reachability.ReachableNodes)
	}
	if len(r.Reachability.UnreachableNodes) != 1 || r.Reachability.UnreachableNodes[0] != "orphan" {
		t.Errorf("expected unreachable [orphan], got %v", r.Reachability.UnreachableNodes)
	}
}

func TestSinkNodeDetection(t *testing.T) {
	w := &ir.Workflow{
		Start: "start",
		Exit:  "end",
		Nodes: []*ir.Node{
			{ID: "start", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "go"}},
			{ID: "sink", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "stuck"}},
			{ID: "end", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "done"}},
		},
		Edges: []*ir.Edge{
			makeEdge("start", "sink"),
			makeEdge("start", "end"),
			// sink has no path to end
		},
	}

	r := Analyze(w)
	if r.Termination.AllPathsTerminate {
		t.Error("expected AllPathsTerminate to be false")
	}
	if len(r.Termination.SinkNodes) != 1 || r.Termination.SinkNodes[0] != "sink" {
		t.Errorf("expected sink nodes [sink], got %v", r.Termination.SinkNodes)
	}
}

func TestExtractToolOutputs(t *testing.T) {
	tests := []struct {
		name    string
		command string
		want    []string
	}{
		// Existing cases (regression coverage)
		{"single_printf", "printf 'pass'", []string{"pass"}},
		{"double_quote_printf", `printf "fail"`, []string{"fail"}},
		{"echo", "echo 'done'", []string{"done"}},
		{"printf_format", "printf '%s' 'value'", []string{"value"}},
		{"multiple", "printf 'a'\nprintf 'b'", []string{"a", "b"}},
		{"no_match", "run-binary", nil},
		{"dedup", "printf 'x'\nprintf 'x'", []string{"x"}},

		// Issue #40 — file-redirected statements must not extract
		{"redirect_overwrite", "echo 'foo' > log.txt\nprintf 'green'", []string{"green"}},
		{"redirect_append_and_pipe", "echo 'foo' >> log\necho 'bar' | tee /tmp/x\nprintf 'green'", []string{"green"}},
		{"stderr_redirect", "echo 'oops' >&2\nprintf 'real'", []string{"real"}},

		// Issue #40 — preserve format-two-arg pattern with newline format
		{"printf_format_newline", "printf '%s\\n' 'inline'\nprintf 'tail'", []string{"inline", "tail"}},

		// Issue #40 — echo with flags now extracts (regex missed it)
		{"echo_with_flags", "echo -n 'foo'", []string{"foo"}},

		// Issue #40 — command substitution: inner skipped, outer's only arg is non-literal
		{"echo_command_sub", "echo $(printf 'inner')\nprintf 'outer'", []string{"outer"}},

		// Issue #40 — subshell without redirect still writes to stdout
		{"subshell_unredirected", "( echo 'inside' )\nprintf 'outside'", []string{"inside", "outside"}},

		// Issue #40 — parse errors return nil (no regex fallback)
		{"malformed_shell", "echo 'unclosed", nil},

		// Issue #40 — && / || (non-pipe BinaryCmd) descends into both sides
		{"echo_and", "echo 'a' && echo 'b'", []string{"a", "b"}},

		// PR review — only stdout-affecting redirects should skip the statement
		{"stdin_redirect_keeps_stdout", "printf 'ok' < input.txt", []string{"ok"}},
		{"stderr_only_redirect", "printf 'ok' 2>err.log", []string{"ok"}},
		{"stderr_dup_to_stdout", "printf 'ok' 2>&1", []string{"ok"}},
		{"fd1_explicit_redirect", "printf 'gone' 1>out.log\nprintf 'kept'", []string{"kept"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractToolOutputs(tt.command)
			if !slicesEqual(got, tt.want) {
				t.Errorf("extractToolOutputs(%q) = %v, want %v", tt.command, got, tt.want)
			}
		})
	}
}

// --- appendConditionValues tests ---

func TestAppendConditionValues_CondOr(t *testing.T) {
	expr := ir.CondOr{
		Left:  ir.CondCompare{Variable: "ctx.x", Op: "=", Value: "a"},
		Right: ir.CondCompare{Variable: "ctx.x", Op: "=", Value: "b"},
	}
	got := appendConditionValues(nil, expr)
	want := []string{"a", "b"}
	if !slicesEqual(got, want) {
		t.Errorf("appendConditionValues(CondOr) = %v, want %v", got, want)
	}
}

func TestAppendConditionValues_CondNot(t *testing.T) {
	expr := ir.CondNot{
		Inner: ir.CondCompare{Variable: "ctx.x", Op: "=", Value: "bad"},
	}
	got := appendConditionValues(nil, expr)
	want := []string{"bad"}
	if !slicesEqual(got, want) {
		t.Errorf("appendConditionValues(CondNot) = %v, want %v", got, want)
	}
}

func TestAppendConditionValues_CondAnd(t *testing.T) {
	expr := ir.CondAnd{
		Left:  ir.CondCompare{Variable: "ctx.x", Op: "=", Value: "yes"},
		Right: ir.CondCompare{Variable: "ctx.y", Op: "=", Value: "no"},
	}
	got := appendConditionValues(nil, expr)
	want := []string{"yes", "no"}
	if !slicesEqual(got, want) {
		t.Errorf("appendConditionValues(CondAnd) = %v, want %v", got, want)
	}
}

func TestAppendConditionValues_NilExpr(t *testing.T) {
	got := appendConditionValues(nil, nil)
	if len(got) != 0 {
		t.Errorf("appendConditionValues(nil) = %v, want empty", got)
	}
}

func TestAppendConditionValues_Nested(t *testing.T) {
	// (a OR b) AND NOT(c)
	expr := ir.CondAnd{
		Left: ir.CondOr{
			Left:  ir.CondCompare{Variable: "ctx.x", Op: "=", Value: "a"},
			Right: ir.CondCompare{Variable: "ctx.x", Op: "=", Value: "b"},
		},
		Right: ir.CondNot{
			Inner: ir.CondCompare{Variable: "ctx.y", Op: "=", Value: "c"},
		},
	}
	got := appendConditionValues(nil, expr)
	want := []string{"a", "b", "c"}
	if !slicesEqual(got, want) {
		t.Errorf("appendConditionValues(nested) = %v, want %v", got, want)
	}
}

// --- addReverseNodeEdges tests ---

func TestAddReverseNodeEdges_Parallel(t *testing.T) {
	adj := make(map[string][]string)
	n := &ir.Node{
		ID:     "FanOut",
		Kind:   ir.NodeParallel,
		Config: ir.ParallelConfig{Targets: []string{"W1", "W2", "W3"}},
	}
	addReverseNodeEdges(adj, n)

	// Each target should have a reverse edge to FanOut.
	for _, target := range []string{"W1", "W2", "W3"} {
		found := false
		for _, src := range adj[target] {
			if src == "FanOut" {
				found = true
			}
		}
		if !found {
			t.Errorf("expected reverse edge from %s to FanOut", target)
		}
	}
}

func TestAddReverseNodeEdges_FanIn(t *testing.T) {
	adj := make(map[string][]string)
	n := &ir.Node{
		ID:     "Join",
		Kind:   ir.NodeFanIn,
		Config: ir.FanInConfig{Sources: []string{"W1", "W2"}},
	}
	addReverseNodeEdges(adj, n)

	// Join should have reverse edges to its sources.
	if !slicesEqual(adj["Join"], []string{"W1", "W2"}) {
		t.Errorf("expected reverse edges [W1 W2] for Join, got %v", adj["Join"])
	}
}

func TestAddReverseNodeEdges_AgentNode(t *testing.T) {
	adj := make(map[string][]string)
	n := &ir.Node{
		ID:     "Agent1",
		Kind:   ir.NodeAgent,
		Config: ir.AgentConfig{Prompt: "test"},
	}
	addReverseNodeEdges(adj, n)

	// Agent node should not add any reverse edges.
	if len(adj) != 0 {
		t.Errorf("expected no reverse edges for agent node, got %v", adj)
	}
}

// --- addImplicitEdges tests ---

func TestAddImplicitEdges_Parallel(t *testing.T) {
	adj := make(map[string][]string)
	w := &ir.Workflow{
		Nodes: []*ir.Node{
			{ID: "FanOut", Kind: ir.NodeParallel, Config: ir.ParallelConfig{
				Targets: []string{"W1", "W2"},
			}},
		},
	}
	addImplicitEdges(adj, w)

	if !slicesEqual(adj["FanOut"], []string{"W1", "W2"}) {
		t.Errorf("expected implicit edges [W1 W2] from FanOut, got %v", adj["FanOut"])
	}
}

func TestAddImplicitEdges_FanIn(t *testing.T) {
	adj := make(map[string][]string)
	w := &ir.Workflow{
		Nodes: []*ir.Node{
			{ID: "Join", Kind: ir.NodeFanIn, Config: ir.FanInConfig{
				Sources: []string{"W1", "W2"},
			}},
		},
	}
	addImplicitEdges(adj, w)

	// Each source should have an implicit edge to Join.
	for _, src := range []string{"W1", "W2"} {
		found := false
		for _, dst := range adj[src] {
			if dst == "Join" {
				found = true
			}
		}
		if !found {
			t.Errorf("expected implicit edge from %s to Join", src)
		}
	}
}

// --- Reachability with parallel/fan_in ---

func TestReachabilityWithParallelFanIn(t *testing.T) {
	w := &ir.Workflow{
		Start: "start",
		Exit:  "end",
		Nodes: []*ir.Node{
			{ID: "start", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "go"}},
			{ID: "fan_out", Kind: ir.NodeParallel, Config: ir.ParallelConfig{
				Targets: []string{"w1", "w2"},
			}},
			{ID: "w1", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "w1"}},
			{ID: "w2", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "w2"}},
			{ID: "join", Kind: ir.NodeFanIn, Config: ir.FanInConfig{
				Sources: []string{"w1", "w2"},
			}},
			{ID: "end", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "done"}},
		},
		Edges: []*ir.Edge{
			makeEdge("start", "fan_out"),
			makeEdge("join", "end"),
		},
	}

	r := Analyze(w)

	// All nodes should be reachable due to implicit edges.
	if r.Reachability.ReachableNodes != 6 {
		t.Errorf("expected 6 reachable nodes, got %d (unreachable: %v)",
			r.Reachability.ReachableNodes, r.Reachability.UnreachableNodes)
	}

	// All paths should terminate.
	if !r.Termination.AllPathsTerminate {
		t.Errorf("expected all paths to terminate, sink nodes: %v", r.Termination.SinkNodes)
	}
}

// --- Termination with parallel ---

func TestTerminationWithParallelSink(t *testing.T) {
	w := &ir.Workflow{
		Start: "start",
		Exit:  "end",
		Nodes: []*ir.Node{
			{ID: "start", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "go"}},
			{ID: "fan_out", Kind: ir.NodeParallel, Config: ir.ParallelConfig{
				Targets: []string{"w1", "w2"},
			}},
			{ID: "w1", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "w1"}},
			{ID: "w2", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "w2"}},
			// No fan_in — w1 and w2 are sinks.
			{ID: "end", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "done"}},
		},
		Edges: []*ir.Edge{
			makeEdge("start", "fan_out"),
			makeEdge("start", "end"),
		},
	}

	r := Analyze(w)
	if r.Termination.AllPathsTerminate {
		t.Error("expected some sink nodes when fan-out targets have no path to exit")
	}
}

// --- Tool node with declared outputs ---

func TestDeclaredOutputsCoverage(t *testing.T) {
	w := &ir.Workflow{
		Start: "start",
		Exit:  "end",
		Nodes: []*ir.Node{
			{ID: "start", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "go"}},
			{ID: "check", Kind: ir.NodeTool, Config: ir.ToolConfig{
				Command: "custom-binary",
				Outputs: []string{"pass", "fail", "error"},
			}},
			{ID: "end", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "done"}},
		},
		Edges: []*ir.Edge{
			makeEdge("start", "check"),
			makeCondEdge("check", "end", "ctx.result", "pass"),
			makeCondEdge("check", "end", "ctx.result", "fail"),
		},
	}

	r := Analyze(w)
	cov := r.Nodes["check"]
	// determineStatus returns "unknown" when ExtractedOutputs is empty,
	// even with DeclaredOutputs — status only considers extracted outputs.
	if cov.Status != "unknown" {
		t.Errorf("expected unknown (declared outputs not in extracted), got %s", cov.Status)
	}
	if len(cov.MissingEdges) != 1 || cov.MissingEdges[0] != "error" {
		t.Errorf("expected missing [error], got %v", cov.MissingEdges)
	}
	// DeclaredOutputs should be used, not extracted.
	if len(cov.ExtractedOutputs) != 0 {
		t.Errorf("expected no extracted outputs when declared, got %v", cov.ExtractedOutputs)
	}
	if len(cov.DeclaredOutputs) != 3 {
		t.Errorf("expected 3 declared outputs, got %v", cov.DeclaredOutputs)
	}
}

// --- Condition edge with CondOr in edge for collectEdgeConditions ---

func TestCollectEdgeConditions_OrCondition(t *testing.T) {
	w := &ir.Workflow{
		Start: "start",
		Exit:  "end",
		Nodes: []*ir.Node{
			{ID: "start", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "go"}},
			{ID: "check", Kind: ir.NodeTool, Config: ir.ToolConfig{
				Command: "printf 'a'\nprintf 'b'\nprintf 'c'",
			}},
			{ID: "end", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "done"}},
		},
		Edges: []*ir.Edge{
			makeEdge("start", "check"),
			{From: "check", To: "end", Condition: &ir.Condition{
				Raw: "ctx.result = a or ctx.result = b",
				Parsed: ir.CondOr{
					Left:  ir.CondCompare{Variable: "ctx.result", Op: "=", Value: "a"},
					Right: ir.CondCompare{Variable: "ctx.result", Op: "=", Value: "b"},
				},
			}},
		},
	}

	r := Analyze(w)
	cov := r.Nodes["check"]
	// OR extracts both "a" and "b" values.
	if len(cov.EdgeConditions) != 2 {
		t.Errorf("expected 2 edge conditions, got %v", cov.EdgeConditions)
	}
}

// --- Condition edge with CondNot for collectEdgeConditions ---

func TestCollectEdgeConditions_NotCondition(t *testing.T) {
	w := &ir.Workflow{
		Start: "start",
		Exit:  "end",
		Nodes: []*ir.Node{
			{ID: "start", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "go"}},
			{ID: "check", Kind: ir.NodeTool, Config: ir.ToolConfig{
				Command: "printf 'pass'\nprintf 'fail'",
			}},
			{ID: "end", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "done"}},
		},
		Edges: []*ir.Edge{
			makeEdge("start", "check"),
			{From: "check", To: "end", Condition: &ir.Condition{
				Raw:    "not ctx.result = fail",
				Parsed: ir.CondNot{Inner: ir.CondCompare{Variable: "ctx.result", Op: "=", Value: "fail"}},
			}},
		},
	}

	r := Analyze(w)
	cov := r.Nodes["check"]
	// NOT extracts the inner value "fail".
	if len(cov.EdgeConditions) != 1 || cov.EdgeConditions[0] != "fail" {
		t.Errorf("expected edge conditions [fail], got %v", cov.EdgeConditions)
	}
}

// slicesEqual compares two string slices for equality.
func slicesEqual(a, b []string) bool {
	if len(a) == 0 && len(b) == 0 {
		return true
	}
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
