package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/2389-research/dippin-lang/ir"
	"github.com/2389-research/dippin-lang/validator"
)

// --- fixtures: real .dip files parsed by the real parser ---

// entryRestrictsRefs is an entry that locks down agent Lock and delegates to a
// child via manager_loop. childRef is the child file name in the same dir.
func entryRestrictsRefs(childRef string) string {
	return `workflow Entry
  start: Lock
  exit: Sup

  agent Lock
    prompt: "x"
    tool_access: none

  manager_loop Sup
    subgraph_ref: ` + childRef + `
    max_cycles: 3

  edges
    Lock -> Sup
`
}

// entryNoIntentRefs delegates to a child but declares no tool_access itself.
func entryNoIntentRefs(childRef string) string {
	return `workflow Entry
  start: Go
  exit: Sup

  agent Go
    prompt: "x"

  manager_loop Sup
    subgraph_ref: ` + childRef + `
    max_cycles: 3

  edges
    Go -> Sup
`
}

// Children mirror the proven minimalDip shape (human start -> agent, with edges)
// so the parser accepts them; only the agents' tool_access differs.
const childZeroIntent = `workflow Child
  start: Ask
  exit: Done

  human Ask
    mode: freeform

  agent Done
    prompt: "do it"

  edges
    Ask -> Done
`

const childFullRestrict = `workflow Child
  start: Ask
  exit: Done

  human Ask
    mode: freeform

  agent Done
    prompt: "do it"
    tool_access: none

  edges
    Ask -> Done
`

const childPartial = `workflow Child
  start: Ask
  exit: B

  human Ask
    mode: freeform

  agent A
    prompt: "a"
    tool_access: none

  agent B
    prompt: "b"

  edges
    Ask -> A
    A -> B
`

const childAgentless = `workflow Child
  start: Ask
  exit: Done

  human Ask
    mode: freeform

  human Done
    mode: freeform

  edges
    Ask -> Done
`

// writeWorkflows writes name->content into a fresh temp dir and returns the dir.
func writeWorkflows(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	for name, content := range files {
		p := filepath.Join(dir, name)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", p, err)
		}
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			t.Fatalf("write %s: %v", p, err)
		}
	}
	return dir
}

// crossDiags loads dir/entry through the real loader and runs the pass.
func crossDiags(t *testing.T, dir, entry string) ([]validator.Diagnostic, map[ir.SourceLocation]childPosture) {
	t.Helper()
	entryPath := filepath.Join(dir, entry)
	w, err := loadWorkflow(entryPath)
	if err != nil {
		t.Fatalf("load %s: %v", entry, err)
	}
	return crossFileToolAccess(w, entryPath)
}

func countCode(diags []validator.Diagnostic, code string) int {
	n := 0
	for _, d := range diags {
		if d.Code == code {
			n++
		}
	}
	return n
}

func TestCrossFile_FiresOnZeroIntentChild(t *testing.T) {
	dir := writeWorkflows(t, map[string]string{
		"entry.dip": entryRestrictsRefs("child.dip"),
		"child.dip": childZeroIntent,
	})
	diags, classified := crossDiags(t, dir, "entry.dip")
	if got := countCode(diags, validator.DIP146); got != 1 {
		t.Fatalf("want 1 DIP146, got %d: %v", got, diags)
	}
	found := false
	for _, p := range classified {
		if p == postureZeroIntent {
			found = true
		}
	}
	if !found {
		t.Errorf("expected a zero-intent classification: %v", classified)
	}
}

func TestCrossFile_SilentOnFullRestrict(t *testing.T) {
	dir := writeWorkflows(t, map[string]string{
		"entry.dip": entryRestrictsRefs("child.dip"),
		"child.dip": childFullRestrict,
	})
	diags, _ := crossDiags(t, dir, "entry.dip")
	if got := countCode(diags, validator.DIP146); got != 0 {
		t.Fatalf("want 0 DIP146 (child fully restricted), got %d", got)
	}
}

func TestCrossFile_SilentOnPartialAudit(t *testing.T) {
	dir := writeWorkflows(t, map[string]string{
		"entry.dip": entryRestrictsRefs("child.dip"),
		"child.dip": childPartial,
	})
	diags, classified := crossDiags(t, dir, "entry.dip")
	if got := countCode(diags, validator.DIP146); got != 0 {
		t.Fatalf("want 0 DIP146 (partial audit keeps DIP143), got %d", got)
	}
	found := false
	for _, p := range classified {
		if p == posturePartial {
			found = true
		}
	}
	if !found {
		t.Errorf("expected a partial classification: %v", classified)
	}
}

func TestCrossFile_SilentWhenNoPathIntent(t *testing.T) {
	dir := writeWorkflows(t, map[string]string{
		"entry.dip": entryNoIntentRefs("child.dip"),
		"child.dip": childZeroIntent,
	})
	diags, _ := crossDiags(t, dir, "entry.dip")
	if got := countCode(diags, validator.DIP146); got != 0 {
		t.Fatalf("want 0 DIP146 (no intent on path), got %d", got)
	}
}

func TestCrossFile_AncestorPathIntent(t *testing.T) {
	// entry (no intent) -> mid (restricts) -> leaf (zero intent) => DIP146 on mid->leaf.
	mid := `workflow Mid
  start: Lock
  exit: Sup

  agent Lock
    prompt: "x"
    tool_access: none

  manager_loop Sup
    subgraph_ref: leaf.dip
    max_cycles: 3

  edges
    Lock -> Sup
`
	dir := writeWorkflows(t, map[string]string{
		"entry.dip": entryNoIntentRefs("mid.dip"),
		"mid.dip":   mid,
		"leaf.dip":  childZeroIntent,
	})
	diags, _ := crossDiags(t, dir, "entry.dip")
	if got := countCode(diags, validator.DIP146); got != 1 {
		t.Fatalf("want 1 DIP146 (ancestor-path intent), got %d: %v", got, diags)
	}
}

func TestCrossFile_Transitive(t *testing.T) {
	// entry (restricts) -> child (restricts) -> grandchild (zero) => DIP146 on grandchild edge.
	child := `workflow Child
  start: Lock
  exit: Sup

  agent Lock
    prompt: "x"
    tool_access: none

  manager_loop Sup
    subgraph_ref: grand.dip
    max_cycles: 3

  edges
    Lock -> Sup
`
	dir := writeWorkflows(t, map[string]string{
		"entry.dip": entryRestrictsRefs("child.dip"),
		"child.dip": child,
		"grand.dip": childZeroIntent,
	})
	diags, _ := crossDiags(t, dir, "entry.dip")
	if got := countCode(diags, validator.DIP146); got != 1 {
		t.Fatalf("want 1 DIP146 on the grandchild edge, got %d: %v", got, diags)
	}
}

func TestCrossFile_DiamondTwoFindings(t *testing.T) {
	// P1 and P2 both reference the same zero-intent child => 2 DIP146 (one per edge).
	parent := func(name, ref string) string {
		return `workflow ` + name + `
  start: Lock
  exit: Sup

  agent Lock
    prompt: "x"
    tool_access: none

  manager_loop Sup
    subgraph_ref: ` + ref + `
    max_cycles: 3

  edges
    Lock -> Sup
`
	}
	entry := `workflow Entry
  start: P1
  exit: P2

  subgraph P1
    ref: p1.dip

  subgraph P2
    ref: p2.dip

  edges
    P1 -> P2
`
	dir := writeWorkflows(t, map[string]string{
		"entry.dip": entry,
		"p1.dip":    parent("P1", "child.dip"),
		"p2.dip":    parent("P2", "child.dip"),
		"child.dip": childZeroIntent,
	})
	diags, _ := crossDiags(t, dir, "entry.dip")
	if got := countCode(diags, validator.DIP146); got != 2 {
		t.Fatalf("want 2 DIP146 (one per parent->child edge), got %d: %v", got, diags)
	}
}

func TestCrossFile_CycleTerminates(t *testing.T) {
	// A -> B -> A. Must terminate; A has intent so no DIP146 on the B->A edge.
	a := `workflow A
  start: Lock
  exit: Sup

  agent Lock
    prompt: "x"
    tool_access: none

  manager_loop Sup
    subgraph_ref: b.dip
    max_cycles: 3

  edges
    Lock -> Sup
`
	b := `workflow B
  start: Go
  exit: Sup

  human Go
    mode: freeform

  manager_loop Sup
    subgraph_ref: a.dip
    max_cycles: 3

  edges
    Go -> Sup
`
	dir := writeWorkflows(t, map[string]string{
		"a.dip": a,
		"b.dip": b,
	})
	// Completing at all proves termination (no hang/stack overflow).
	diags, _ := crossDiags(t, dir, "a.dip")
	_ = diags
}

func TestCrossFile_SelfReferenceTerminates(t *testing.T) {
	// A -> A. Must terminate; A is agent-less here so no DIP146 either way.
	a := `workflow A
  start: Go
  exit: Sup

  human Go
    mode: freeform

  manager_loop Sup
    subgraph_ref: a.dip
    max_cycles: 3

  edges
    Go -> Sup
`
	dir := writeWorkflows(t, map[string]string{"a.dip": a})
	diags, _ := crossDiags(t, dir, "a.dip")
	if got := countCode(diags, validator.DIP146); got != 0 {
		t.Fatalf("want 0 DIP146 on self-reference, got %d", got)
	}
}

func TestCrossFile_UnresolvableChildIsSilent(t *testing.T) {
	dir := writeWorkflows(t, map[string]string{
		"entry.dip": entryRestrictsRefs("missing.dip"), // no missing.dip on disk
	})
	diags, classified := crossDiags(t, dir, "entry.dip")
	if got := countCode(diags, validator.DIP146); got != 0 {
		t.Fatalf("want 0 DIP146 for unresolvable child, got %d", got)
	}
	for _, p := range classified {
		if p != postureUnresolved {
			t.Errorf("want postureUnresolved, got %v", p)
		}
	}
}

func TestCrossFile_AgentlessChildIsSilent(t *testing.T) {
	dir := writeWorkflows(t, map[string]string{
		"entry.dip": entryRestrictsRefs("child.dip"),
		"child.dip": childAgentless,
	})
	diags, _ := crossDiags(t, dir, "entry.dip")
	if got := countCode(diags, validator.DIP146); got != 0 {
		t.Fatalf("want 0 DIP146 (agent-less child has no tools to grant), got %d", got)
	}
}

// Known intentional false positive: a fully-tooled worker pool fires DIP146.
// This is mitigated by Hint severity, NOT by the gate. Do not "fix" this into a
// Warning or add a heuristic — see the design spec (#89).
func TestCrossFile_IntentionalOpenWorkerFires(t *testing.T) {
	dir := writeWorkflows(t, map[string]string{
		"entry.dip":  entryRestrictsRefs("worker.dip"),
		"worker.dip": childZeroIntent,
	})
	diags, _ := crossDiags(t, dir, "entry.dip")
	if got := countCode(diags, validator.DIP146); got != 1 {
		t.Fatalf("want 1 DIP146 (known intentional-open FP, Hint-mitigated), got %d", got)
	}
}
