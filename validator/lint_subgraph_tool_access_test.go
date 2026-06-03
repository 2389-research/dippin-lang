package validator

import "testing"

// DIP143 fires (Hint) when a workflow declares tool_access containment intent
// AND references an external subgraph whose own file does not inherit it.
// Intent gate: any non-empty tool_access on an agent or parallel branch.
// Scope: both manager_loop (subgraph_ref) and subgraph (ref) nodes.

const dip143 = "DIP143"

func countCode(diags []Diagnostic, code string) int {
	n := 0
	for _, d := range diags {
		if d.Code == code {
			n++
		}
	}
	return n
}

func TestLint_DIP143_ManagerLoopWithRestrictedAgent(t *testing.T) {
	src := `workflow X
  start: Lock
  exit: Sup

  agent Lock
    prompt: "x"
    tool_access: none

  manager_loop Sup
    subgraph_ref: child.dip
    max_cycles: 3

  edges
    Lock -> Sup
`
	if !hasCode(lintSrc(t, src), dip143) {
		t.Errorf("expected DIP143; got: %v", codes(lintSrc(t, src)))
	}
}

func TestLint_DIP143_SubgraphNodeWithRestrictedAgent(t *testing.T) {
	src := `workflow X
  start: Lock
  exit: Sub

  agent Lock
    prompt: "x"
    tool_access: none

  subgraph Sub
    ref: child.dip

  edges
    Lock -> Sub
`
	if !hasCode(lintSrc(t, src), dip143) {
		t.Errorf("expected DIP143 on subgraph node; got: %v", codes(lintSrc(t, src)))
	}
}

func TestLint_DIP143_ParallelBranchIntent(t *testing.T) {
	src := `workflow X
  start: split
  exit: Sup

  agent a
    prompt: "a"

  parallel split
    branch: a
      tool_access: none

  fan_in join <- a

  manager_loop Sup
    subgraph_ref: child.dip
    max_cycles: 3

  edges
    split -> a
    a -> join
    join -> Sup
`
	if !hasCode(lintSrc(t, src), dip143) {
		t.Errorf("expected DIP143 from branch intent; got: %v", codes(lintSrc(t, src)))
	}
}

func TestLint_DIP143_AnyNonEmptyToolAccessIsIntent(t *testing.T) {
	// A fail-closed typo still expresses (and receives) restriction at runtime,
	// so any non-empty tool_access counts as containment intent.
	src := `workflow X
  start: Lock
  exit: Sup

  agent Lock
    prompt: "x"
    tool_access: nono

  manager_loop Sup
    subgraph_ref: child.dip
    max_cycles: 3

  edges
    Lock -> Sup
`
	if !hasCode(lintSrc(t, src), dip143) {
		t.Errorf("expected DIP143 (non-empty tool_access = intent); got: %v", codes(lintSrc(t, src)))
	}
}

func TestLint_DIP143_NoIntentNoFire(t *testing.T) {
	src := `workflow X
  start: A
  exit: Sup

  agent A
    prompt: "x"

  manager_loop Sup
    subgraph_ref: child.dip
    max_cycles: 3

  edges
    A -> Sup
`
	if hasCode(lintSrc(t, src), dip143) {
		t.Errorf("DIP143 should not fire without tool_access intent; got: %v", codes(lintSrc(t, src)))
	}
}

func TestLint_DIP143_RestrictedAgentNoSubgraphNoFire(t *testing.T) {
	src := `workflow X
  start: A
  exit: A

  agent A
    prompt: "x"
    tool_access: none
`
	if hasCode(lintSrc(t, src), dip143) {
		t.Errorf("DIP143 should not fire without a subgraph reference; got: %v", codes(lintSrc(t, src)))
	}
}

func TestLint_DIP143_EmptyRefNoFire(t *testing.T) {
	// manager_loop with no subgraph_ref is DIP135's concern, not DIP143's.
	src := `workflow X
  start: Lock
  exit: Sup

  agent Lock
    prompt: "x"
    tool_access: none

  manager_loop Sup
    max_cycles: 3

  edges
    Lock -> Sup
`
	if hasCode(lintSrc(t, src), dip143) {
		t.Errorf("DIP143 should not fire on an empty ref; got: %v", codes(lintSrc(t, src)))
	}
}

func TestLint_DIP143_SelfReferenceNoFire(t *testing.T) {
	// A node referencing its own source file (test.dip, the lintSrc filename) is
	// not a cross-file boundary — there is nothing the parent's restriction fails
	// to reach. Transitive cross-file cycles remain out of scope (#89).
	src := `workflow X
  start: Lock
  exit: Sup

  agent Lock
    prompt: "x"
    tool_access: none

  manager_loop Sup
    subgraph_ref: test.dip
    max_cycles: 3

  edges
    Lock -> Sup
`
	if hasCode(lintSrc(t, src), dip143) {
		t.Errorf("DIP143 should not fire on a self-reference; got: %v", codes(lintSrc(t, src)))
	}
}

func TestLint_DIP143_OneHintPerReferencingNode(t *testing.T) {
	src := `workflow X
  start: Lock
  exit: Sup

  agent Lock
    prompt: "x"
    tool_access: none

  manager_loop Sup
    subgraph_ref: child.dip
    max_cycles: 3

  edges
    Lock -> Sup
`
	if got := countCode(lintSrc(t, src), dip143); got != 1 {
		t.Errorf("expected exactly 1 DIP143, got %d", got)
	}
}

func TestLint_DIP143_SeverityIsHint(t *testing.T) {
	src := `workflow X
  start: Lock
  exit: Sup

  agent Lock
    prompt: "x"
    tool_access: none

  manager_loop Sup
    subgraph_ref: child.dip
    max_cycles: 3

  edges
    Lock -> Sup
`
	var found bool
	for _, d := range lintSrc(t, src) {
		if d.Code == dip143 {
			found = true
			if d.Severity != SeverityHint {
				t.Errorf("DIP143 severity = %v, want SeverityHint", d.Severity)
			}
		}
	}
	if !found {
		t.Fatal("DIP143 not emitted")
	}
}
