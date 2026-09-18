package simulate

import (
	"testing"

	"github.com/2389-research/dippin-lang/ir"
	"github.com/2389-research/dippin-lang/parser"
)

// mustParseElseWorkflow parses .dip source through the real parser and
// ensures conditions are AST-parsed, matching what Run/RunAllPaths expect
// (CLAUDE.md: "Test fixtures should match real parser output").
func mustParseElseWorkflow(t *testing.T, src string) *ir.Workflow {
	t.Helper()
	w, err := parser.NewParser(src, "t.dip").Parse()
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if err := EnsureConditionsParsed(w); err != nil {
		t.Fatalf("EnsureConditionsParsed: %v", err)
	}
	return w
}

// nonExhaustiveElseWorkflow: Route has a single guard (ctx.flag = yes) that
// never matches (flag is unset), no unconditional edge, and a section
// `else -> Cleanup`. The guard set is non-exhaustive, so the no-match case
// routes to Cleanup.
func nonExhaustiveElseWorkflow() *ir.Workflow {
	return &ir.Workflow{
		Name: "NonExhaustiveElse", Version: "1", Start: "Route", Exit: "Done",
		ElseTarget: "Cleanup",
		Nodes: []*ir.Node{
			{ID: "Route", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "route"}},
			{ID: "Win", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "win"}},
			{ID: "Cleanup", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "cleanup"}},
			{ID: "Done", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "done"}},
		},
		Edges: []*ir.Edge{
			{From: "Route", To: "Win", Condition: &ir.Condition{Raw: "ctx.flag = yes"}},
			{From: "Win", To: "Done"},
			{From: "Cleanup", To: "Done"},
		},
	}
}

// partitionElseWorkflow: Route guards on a gold/silver complete partition
// (which EdgesExhaustive treats as exhaustive) plus `else -> Cleanup`. A
// concrete scenario value the guards don't cover (tier=bronze) must still route
// to Cleanup — the engine falls to else on any unmatched outcome.
func partitionElseWorkflow() *ir.Workflow {
	return &ir.Workflow{
		Name: "PartitionElse", Version: "1", Start: "Route", Exit: "Done",
		ElseTarget: "Cleanup",
		Nodes: []*ir.Node{
			{ID: "Route", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "route"}},
			{ID: "Gold", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "gold"}},
			{ID: "Silver", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "silver"}},
			{ID: "Cleanup", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "cleanup"}},
			{ID: "Done", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "done"}},
		},
		Edges: []*ir.Edge{
			{From: "Route", To: "Gold", Condition: &ir.Condition{Raw: "ctx.tier = gold"}},
			{From: "Route", To: "Silver", Condition: &ir.Condition{Raw: "ctx.tier = silver"}},
			{From: "Gold", To: "Done"},
			{From: "Silver", To: "Done"},
			{From: "Cleanup", To: "Done"},
		},
	}
}

// exhaustiveElseWorkflow: Route guards on the exhaustive ctx.outcome
// success/fail set with `else -> Cleanup`. Used by the path enumerator, which
// trusts the declared partition and must not enumerate an unreachable else path.
func exhaustiveElseWorkflow() *ir.Workflow {
	return &ir.Workflow{
		Name: "ExhaustiveElse", Version: "1", Start: "Route", Exit: "Done",
		ElseTarget: "Cleanup",
		Nodes: []*ir.Node{
			{ID: "Route", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "route"}},
			{ID: "Win", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "win"}},
			{ID: "Lose", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "lose"}},
			{ID: "Cleanup", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "cleanup"}},
			{ID: "Done", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "done"}},
		},
		Edges: []*ir.Edge{
			{From: "Route", To: "Win", Condition: &ir.Condition{Raw: "ctx.outcome = success"}},
			{From: "Route", To: "Lose", Condition: &ir.Condition{Raw: "ctx.outcome = fail"}},
			{From: "Win", To: "Done"},
			{From: "Lose", To: "Done"},
			{From: "Cleanup", To: "Done"},
		},
	}
}

// successOnlyElseWorkflow: Build guards only on ctx.outcome = success (no
// explicit `on fail` edge, no unconditional edge of its own), plus a section
// `else -> Cleanup`. Per docs/edges.md's Failure Handling contract, `else` is
// success-side only — a genuine fail outcome must NOT route to Cleanup via
// else. Used for issue #306. Parsed through the real parser per CLAUDE.md's
// Testing guidance (also exercises `on fail` edge-guard sugar elsewhere).
func successOnlyElseWorkflow(t *testing.T) *ir.Workflow {
	return mustParseElseWorkflow(t, `workflow SuccessOnlyElse
  start: Setup
  exit: Done

  agent Setup
    prompt: "setup"

  agent Build
    prompt: "build"

  agent Cleanup
    prompt: "cleanup"

  agent Done
    prompt: "done"

  edges
    Setup -> Build
    Build -> Done when ctx.outcome = success
    Cleanup -> Done
    else -> Cleanup
`)
}

// explicitFailEdgeElseWorkflow: like successOnlyElseWorkflow, but Build also
// declares its own explicit `on fail` edge to Escalate (using the `on fail`
// sugar spelling for `when ctx.outcome = fail`). A fail outcome must still
// route through that explicit fail edge (the failure cascade), not through
// else, and not through the generic edges[0] fallback either.
func explicitFailEdgeElseWorkflow(t *testing.T) *ir.Workflow {
	return mustParseElseWorkflow(t, `workflow ExplicitFailEdgeElse
  start: Setup
  exit: Done

  agent Setup
    prompt: "setup"

  agent Build
    prompt: "build"

  agent Escalate
    prompt: "escalate"

  agent Cleanup
    prompt: "cleanup"

  agent Done
    prompt: "done"

  edges
    Setup -> Build
    Build -> Done      when ctx.outcome = success
    Build -> Escalate  on fail
    Escalate -> Done
    Cleanup -> Done
    else -> Cleanup
`)
}

// failThenHumanElseWorkflow: Build routes to a human Review node on fail
// (the failure cascade), and Review — which never touches ctx.outcome itself
// — has its own unmatched-guard case (ctx.answer unset) that must still fall
// to the section `else -> Cleanup`. Regression fixture for #306 follow-up:
// the fail-gate must be scoped to the CURRENT node's own outcome, not a
// stale ctx.outcome="fail" left in the flat context map by Build.
func failThenHumanElseWorkflow(t *testing.T) *ir.Workflow {
	return mustParseElseWorkflow(t, `workflow FailThenHumanElse
  start: Build
  exit: Done

  agent Build
    prompt: "build"

  human Review
    label: "Review the failure"

  agent Cleanup
    prompt: "cleanup"

  agent Done
    prompt: "done"

  edges
    Build -> Review  on fail
    Review -> Done   when ctx.answer = yes
    Cleanup -> Done
    else -> Cleanup
`)
}

// failOnlyElseWorkflow: Route's single guard is ctx.outcome = failure (the
// alternate failure spelling) — no success edge, no unconditional edge of
// its own. The guard's unmatched complement is the SUCCESS side (or any
// other non-failure value), which is exactly what else exists to catch, so
// else must still apply here despite the guard mentioning an outcome value.
func failOnlyElseWorkflow(t *testing.T) *ir.Workflow {
	return mustParseElseWorkflow(t, `workflow FailOnlyElse
  start: Route
  exit: Done

  agent Route
    prompt: "route"

  agent Escalate
    prompt: "escalate"

  agent Cleanup
    prompt: "cleanup"

  agent Done
    prompt: "done"

  edges
    Route -> Escalate  when ctx.outcome = failure
    Escalate -> Done
    Cleanup -> Done
    else -> Cleanup
`)
}

// toolFailFanInElseWorkflow: a tool T fails, routing (via its own explicit
// fail guard) to a fan_in node Join. Join's only conditional edge doesn't
// match and it has no unconditional edge of its own, so it must fall to
// `else -> Cleanup` — a fan_in/parallel node never owns ctx.outcome itself
// (visitNode returns before applyNodeDefaults for these kinds), so it must
// not inherit T's true nodeOwnsOutcome / ctx.outcome="fail" and lose its own
// else default. Regression fixture for #306 follow-up round 2.
func toolFailFanInElseWorkflow(t *testing.T) *ir.Workflow {
	return mustParseElseWorkflow(t, `workflow ToolFailFanInElse
  start: T
  exit: Done

  tool T
    command: "run"

  fan_in Join <- T

  agent Z1
    prompt: "z1"

  agent Cleanup
    prompt: "cleanup"

  agent Done
    prompt: "done"

  edges
    T -> Join    when ctx.outcome = fail
    Join -> Z1   when ctx.x = y
    Z1 -> Done
    Cleanup -> Done
    else -> Cleanup
`)
}

func pathContains(path []string, id string) bool {
	for _, p := range path {
		if p == id {
			return true
		}
	}
	return false
}

func anyPathContains(results []*Result, id string) bool {
	for _, r := range results {
		if pathContains(r.Path, id) {
			return true
		}
	}
	return false
}

func TestSimulate_RoutesNoMatchToElse(t *testing.T) {
	res, err := Run(nonExhaustiveElseWorkflow(), Options{})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !pathContains(res.Path, "Cleanup") {
		t.Errorf("unmatched guard did not route to else target Cleanup; path=%v", res.Path)
	}
	if pathContains(res.Path, "Win") {
		t.Errorf("guard should not have matched; path=%v", res.Path)
	}
}

// A concrete scenario value that matches no guard (bronze, against a gold/silver
// partition) must route to the else default — even though the partition looks
// statically exhaustive — because that is what the engine does.
func TestSimulate_UnmatchedScenarioValueRoutesToElse(t *testing.T) {
	res, err := Run(partitionElseWorkflow(), Options{Scenario: map[string]string{"tier": "bronze"}})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !pathContains(res.Path, "Cleanup") {
		t.Errorf("unmatched scenario value did not route to else target Cleanup; path=%v", res.Path)
	}
	if pathContains(res.Path, "Gold") || pathContains(res.Path, "Silver") {
		t.Errorf("no partition guard should have matched tier=bronze; path=%v", res.Path)
	}
}

func TestRunAllPaths_EnumeratesElseBranch(t *testing.T) {
	results, err := RunAllPaths(nonExhaustiveElseWorkflow(), nil)
	if err != nil {
		t.Fatalf("RunAllPaths: %v", err)
	}
	if !anyPathContains(results, "Cleanup") {
		t.Errorf("path enumerator did not emit the else branch (Cleanup); %d paths", len(results))
	}
	if !anyPathContains(results, "Win") {
		t.Errorf("path enumerator dropped the guard branch (Win); %d paths", len(results))
	}
}

func TestRunAllPaths_NoElseBranchForExhaustive(t *testing.T) {
	results, err := RunAllPaths(exhaustiveElseWorkflow(), nil)
	if err != nil {
		t.Fatalf("RunAllPaths: %v", err)
	}
	if anyPathContains(results, "Cleanup") {
		t.Errorf("exhaustive node should not enumerate an else branch; %d paths reached Cleanup", len(results))
	}
}

func TestRunAllPaths_NoElseBranchForPartition(t *testing.T) {
	results, err := RunAllPaths(partitionElseWorkflow(), nil)
	if err != nil {
		t.Fatalf("RunAllPaths: %v", err)
	}
	if anyPathContains(results, "Cleanup") {
		t.Errorf("complete-partition node should not enumerate an else branch; %d paths reached Cleanup", len(results))
	}
}

// Issue #306(a): a fail outcome against a success-only guard must NOT route
// to the else target — else is success-side only per docs/edges.md. Pins the
// exact fallback path (Minor 3, review round 1): with no `on fail` edge and
// no else, resolveConditionalNext falls back to edges[0] — Build's own
// `when ctx.outcome = success` edge — so the run still completes Setup ->
// Build -> Done with status "success" (not Cleanup, not dead_end).
func TestSimulate_FailOutcomeDoesNotRouteToElse(t *testing.T) {
	res, err := Run(successOnlyElseWorkflow(t), Options{Scenario: map[string]string{"Build.outcome": "fail"}})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	wantPath := []string{"Setup", "Build", "Done"}
	if !equalPaths(res.Path, wantPath) {
		t.Errorf("path = %v, want %v", res.Path, wantPath)
	}
	if res.Status != "success" {
		t.Errorf("status = %q, want success", res.Status)
	}
}

// Issue #306(b): a fail outcome must still route via an explicit `on fail`
// edge when one is declared — only the else shortcut is removed.
func TestSimulate_FailOutcomeStillRoutesViaExplicitFailEdge(t *testing.T) {
	res, err := Run(explicitFailEdgeElseWorkflow(t), Options{Scenario: map[string]string{"Build.outcome": "fail"}})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !pathContains(res.Path, "Escalate") {
		t.Errorf("fail outcome did not route via explicit on-fail edge to Escalate; path=%v", res.Path)
	}
	if pathContains(res.Path, "Cleanup") {
		t.Errorf("fail outcome must not also route to else target Cleanup; path=%v", res.Path)
	}
}

// Issue #306(c): a non-fail unmatched outcome must still route to else — the
// fix is scoped to a failure outcome only, not "unmatched" in general.
func TestSimulate_NonFailUnmatchedOutcomeStillRoutesToElse(t *testing.T) {
	res, err := Run(successOnlyElseWorkflow(t), Options{Scenario: map[string]string{"Build.outcome": "timeout"}})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !pathContains(res.Path, "Cleanup") {
		t.Errorf("non-fail unmatched outcome did not route to else target Cleanup; path=%v", res.Path)
	}
}

// Issue #306(d): the all-paths enumerator must not emit the else branch for
// the outcome=fail complement of a success-only guard.
func TestRunAllPaths_NoElseBranchOnFailComplement(t *testing.T) {
	results, err := RunAllPaths(successOnlyElseWorkflow(t), nil)
	if err != nil {
		t.Fatalf("RunAllPaths: %v", err)
	}
	if anyPathContains(results, "Cleanup") {
		t.Errorf("success-only guard should not enumerate an else-on-fail branch; %d paths reached Cleanup", len(results))
	}
}

// Review round 1, Important 1: a fail outcome injected on Build must route
// Build's own `on fail` edge to the human Review node (the failure cascade),
// but Review — which never sets ctx.outcome itself — must still get its OWN
// else default when its guard (ctx.answer = yes) doesn't match, rather than
// inheriting Build's stale ctx.outcome="fail" from the flat context map and
// losing its else. Base/correct behavior: Build -> Review -> Cleanup -> Done.
func TestSimulate_OnFailThenHumanNodeStillGetsOwnElse(t *testing.T) {
	res, err := Run(failThenHumanElseWorkflow(t), Options{Scenario: map[string]string{"Build.outcome": "fail"}})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	wantPath := []string{"Build", "Review", "Cleanup", "Done"}
	if !equalPaths(res.Path, wantPath) {
		t.Errorf("path = %v, want %v", res.Path, wantPath)
	}
}

// Review round 2: a fan_in node reached after a tool failure must get its
// own else default, not inherit the failing tool's nodeOwnsOutcome/ctx.outcome
// via visitNode's early return for parallel/fan-in kinds (which skips
// applyNodeDefaults — the only place nodeOwnsOutcome is normally refreshed).
func TestSimulate_FanInAfterFailStillGetsOwnElse(t *testing.T) {
	res, err := Run(toolFailFanInElseWorkflow(t), Options{Scenario: map[string]string{"T.outcome": "fail"}})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	wantPath := []string{"T", "Join", "Cleanup", "Done"}
	if !equalPaths(res.Path, wantPath) {
		t.Errorf("path = %v, want %v", res.Path, wantPath)
	}
}

// Review round 1, Important 2: the "failure" spelling of a fail outcome must
// be recognized identically to "fail" — a success-only guard must not route
// to else for outcome=failure either.
func TestSimulate_FailureSpellingDoesNotRouteToElse(t *testing.T) {
	res, err := Run(successOnlyElseWorkflow(t), Options{Scenario: map[string]string{"Build.outcome": "failure"}})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if pathContains(res.Path, "Cleanup") {
		t.Errorf("outcome=failure must not route to else target Cleanup; path=%v", res.Path)
	}
}

// Review round 1, Important 2: the all-paths enumerator must not treat a
// lone `ctx.outcome = failure` guard as omitting fail — its complement is
// the success side (or any other non-failure value), which else legitimately
// catches, so the else branch must still be enumerated.
func TestRunAllPaths_FailureOnlyGuardStillEnumeratesElse(t *testing.T) {
	results, err := RunAllPaths(failOnlyElseWorkflow(t), nil)
	if err != nil {
		t.Fatalf("RunAllPaths: %v", err)
	}
	if !anyPathContains(results, "Cleanup") {
		t.Errorf("a fail-only guard's non-failure complement should still enumerate the else branch; %d paths", len(results))
	}
}

// equalPaths reports whether two node-ID paths are identical, in order.
func equalPaths(got, want []string) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if got[i] != want[i] {
			return false
		}
	}
	return true
}
