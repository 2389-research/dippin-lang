package formatter

import (
	"testing"

	"github.com/2389-research/dippin-lang/ir"
	"github.com/2389-research/dippin-lang/simulate"
)

// v1wf builds a v1 workflow with the given retry config on node "T" and the
// given extra edges from T; conditions are parsed so EdgeRoutesOnFail works.
func v1wf(t *testing.T, retry ir.RetryConfig, extra []*ir.Edge) *ir.Workflow {
	t.Helper()
	nodes := []*ir.Node{
		{ID: "T", Kind: ir.NodeTool, Config: ir.ToolConfig{Command: "run"}, Retry: retry},
		{ID: "Done", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "d"}},
		{ID: "Esc", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "e"}},
	}
	edges := append([]*ir.Edge{{From: "T", To: "Done", Condition: &ir.Condition{Raw: "ctx.outcome = success"}}}, extra...)
	w := &ir.Workflow{Name: "W", Version: "1", Start: "T", Exit: "Done", Nodes: nodes, Edges: edges}
	_ = simulate.EnsureConditionsParsed(w)
	return w
}

func failTo(to string) *ir.Edge {
	return &ir.Edge{From: "T", To: to, Condition: &ir.Condition{Raw: "ctx.outcome = fail"}}
}

func edgeTo(w *ir.Workflow, to string) *ir.Edge {
	for _, e := range w.Edges {
		if e.From == "T" && e.To == to {
			return e
		}
	}
	return nil
}

func TestMigrate_FallbackNoFailEdge_Synthesizes(t *testing.T) {
	w := v1wf(t, ir.RetryConfig{FallbackTarget: "Esc"}, nil)
	notes := MigrateToV2(w)
	if len(notes) != 0 {
		t.Fatalf("clean synthesize should have no notes, got %v", notes)
	}
	if w.Version != "2" || w.Nodes[0].Retry.FallbackTarget != "" {
		t.Fatalf("version/field not cleared: v=%s fb=%q", w.Version, w.Nodes[0].Retry.FallbackTarget)
	}
	e := edgeTo(w, "Esc")
	if e == nil || !ir.EdgeRoutesOnFail(e) {
		t.Fatalf("expected synthesized on-fail edge T->Esc, edges=%v", w.Edges)
	}
}

func TestMigrate_FallbackMatchesFailEdge_Dedupes(t *testing.T) {
	w := v1wf(t, ir.RetryConfig{FallbackTarget: "Esc"}, []*ir.Edge{failTo("Esc")})
	notes := MigrateToV2(w)
	if len(notes) != 0 {
		t.Fatalf("matching target should dedupe with no notes, got %v", notes)
	}
	count := 0
	for _, e := range w.Edges {
		if e.To == "Esc" {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("expected exactly one T->Esc edge after dedupe, got %d", count)
	}
}

func TestMigrate_FallbackDivergesFromFailEdge_FlagsAndKeepsBoth(t *testing.T) {
	w := v1wf(t, ir.RetryConfig{FallbackTarget: "Esc"}, []*ir.Edge{failTo("Done")})
	notes := MigrateToV2(w)
	if len(notes) != 1 {
		t.Fatalf("divergent fallback should produce exactly one review note, got %v", notes)
	}
	e := edgeTo(w, "Esc")
	if e == nil || e.Comment == "" {
		t.Fatalf("expected a flagged T->Esc on-fail edge with a comment, edges=%v", w.Edges)
	}
}

func TestMigrate_SelfRetryTarget_Dropped(t *testing.T) {
	w := v1wf(t, ir.RetryConfig{RetryTarget: "T", MaxRetries: 2}, nil)
	notes := MigrateToV2(w)
	if len(notes) != 0 || w.Nodes[0].Retry.RetryTarget != "" {
		t.Fatalf("self retry_target should drop silently, notes=%v rt=%q", notes, w.Nodes[0].Retry.RetryTarget)
	}
	if edgeTo(w, "T") != nil {
		t.Fatalf("self retry_target must not synthesize a loop edge")
	}
}

func TestMigrate_NonSelfRetryTarget_LoopEdgeAndNote(t *testing.T) {
	w := v1wf(t, ir.RetryConfig{RetryTarget: "Esc", MaxRetries: 2}, nil)
	notes := MigrateToV2(w)
	if len(notes) != 1 {
		t.Fatalf("non-self retry_target should produce one note, got %v", notes)
	}
	e := edgeTo(w, "Esc")
	if e == nil || !e.Restart {
		t.Fatalf("expected a loop edge T->Esc, edges=%v", w.Edges)
	}
}
