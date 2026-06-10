package validator

import (
	"strings"
	"testing"
)

// TestLint_DIP147_ExplicitKeyNoneToFull: a tool_access:none agent writes a
// context key that a downstream tool-bearing agent reads — laundering an
// author-declared key into a privileged prompt. DIP147 fires.
func TestLint_DIP147_ExplicitKeyNoneToFull(t *testing.T) {
	src := `workflow X
  start: Summarize
  exit: Writer

  agent Summarize
    prompt: "summarize the untrusted input"
    tool_access: none
    writes: tainted

  agent Writer
    prompt: "write the report"
    reads: tainted

  edges
    Summarize -> Writer
`
	diags := lintSrc(t, src)
	if !hasCode(diags, DIP147) {
		t.Fatalf("expected DIP147, got: %v", codes(diags))
	}
	var msg string
	for _, d := range diags {
		if d.Code == DIP147 {
			msg = d.Message
		}
	}
	// The diagnostic names both nodes and the laundered key.
	for _, want := range []string{"Summarize", "Writer", "tainted"} {
		if !strings.Contains(msg, want) {
			t.Errorf("DIP147 message should mention %q; got %q", want, msg)
		}
	}
}

// TestLint_DIP147_PerKeyDiagnostics: a restricted agent laundering two distinct
// keys into one tool-bearing sink yields one diagnostic per (source, key, sink)
// flow, not a single collapsed hint.
func TestLint_DIP147_PerKeyDiagnostics(t *testing.T) {
	src := `workflow X
  start: Summarize
  exit: Writer

  agent Summarize
    prompt: "summarize"
    tool_access: none
    writes: tainted_a, tainted_b

  agent Writer
    prompt: "write"
    reads: tainted_a, tainted_b

  edges
    Summarize -> Writer
`
	var n int
	for _, d := range lintSrc(t, src) {
		if d.Code == DIP147 {
			n++
		}
	}
	if n != 2 {
		t.Errorf("expected 2 DIP147 (one per laundered key), got %d", n)
	}
}

// TestLint_DIP147_DuplicateWriteKeyDedup: a duplicated key in writes: yields one
// diagnostic for that (source, key, sink) flow, not one per duplicate.
func TestLint_DIP147_DuplicateWriteKeyDedup(t *testing.T) {
	src := `workflow X
  start: Summarize
  exit: Writer

  agent Summarize
    prompt: "summarize"
    tool_access: none
    writes: tainted, tainted

  agent Writer
    prompt: "write"
    reads: tainted

  edges
    Summarize -> Writer
`
	var n int
	for _, d := range lintSrc(t, src) {
		if d.Code == DIP147 {
			n++
		}
	}
	if n != 1 {
		t.Errorf("expected 1 DIP147 (deduped), got %d", n)
	}
}

// TestLint_DIP147_LastResponseEdgeNotFlagged: a bare none -> full edge with no
// declared key dependency (the ${ctx.last_response} auto-injection topology) is
// intentionally OUT OF SCOPE — that is issue #57 (closed/deferred), and its
// mitigation is #56's last_response_truncate: attribute. DIP147 must not fire on
// adjacency alone.
func TestLint_DIP147_LastResponseEdgeNotFlagged(t *testing.T) {
	src := `workflow X
  start: Summarize
  exit: Writer

  agent Summarize
    prompt: "summarize"
    tool_access: none

  agent Writer
    prompt: "act on ${ctx.last_response}"

  edges
    Summarize -> Writer
`
	diags := lintSrc(t, src)
	if hasCode(diags, DIP147) {
		t.Errorf("DIP147 should not fire on a bare none->full edge (no declared key); got: %v", codes(diags))
	}
}

// TestLint_DIP147_NoneToNoneDoesNotFire: a restricted agent's key flowing into
// another restricted agent keeps the taint in a tool-less context — no privilege
// to escalate to, so DIP147 must not fire.
func TestLint_DIP147_NoneToNoneDoesNotFire(t *testing.T) {
	src := `workflow X
  start: Summarize
  exit: Downstream

  agent Summarize
    prompt: "summarize"
    tool_access: none
    writes: tainted

  agent Downstream
    prompt: "still restricted"
    tool_access: none
    reads: tainted

  edges
    Summarize -> Downstream
`
	diags := lintSrc(t, src)
	if hasCode(diags, DIP147) {
		t.Errorf("DIP147 should not fire for none->none; got: %v", codes(diags))
	}
}

// TestLint_DIP147_MultiHopKeyFlow: the tainted key persists in context across a
// non-agent intermediate (a tool node). The restricted source and the
// tool-bearing sink are not adjacent, but reachability + the shared key still
// flags the laundering.
func TestLint_DIP147_MultiHopKeyFlow(t *testing.T) {
	src := `workflow X
  start: Summarize
  exit: Writer

  agent Summarize
    prompt: "summarize"
    tool_access: none
    writes: tainted

  tool Passthrough
    command: echo hi
    timeout: 5s

  agent Writer
    prompt: "write"
    reads: tainted

  edges
    Summarize -> Passthrough
    Passthrough -> Writer
`
	diags := lintSrc(t, src)
	if !hasCode(diags, DIP147) {
		t.Fatalf("expected DIP147 for multi-hop key flow, got: %v", codes(diags))
	}
}

// TestLint_DIP147_KeyFlowThroughFanIn: the explicit-key flow follows declared-IO
// reachability through any intermediate node, including parallel/fan_in (same
// adjacency as DIP112). A restricted writer routed through a fan_in into a
// downstream tool-bearing reader is a real laundering path and fires — only the
// branch-level tool_access OVERRIDE classification is out of scope, not paths
// that traverse a fan_in.
func TestLint_DIP147_KeyFlowThroughFanIn(t *testing.T) {
	src := `workflow X
  start: Seed
  exit: Writer

  agent Seed
    prompt: "seed"

  agent Risky
    prompt: "handle untrusted"
    tool_access: none
    writes: tainted

  agent Other
    prompt: "other work"

  agent Writer
    prompt: "write"
    reads: tainted

  parallel P -> Risky, Other
  fan_in F <- Risky, Other

  edges
    Seed -> P
    F -> Writer
`
	diags := lintSrc(t, src)
	if !hasCode(diags, DIP147) {
		t.Fatalf("expected DIP147 for key flow through fan_in, got: %v", codes(diags))
	}
}

// TestLint_DIP147_BenignNoFalsePositive: full->full and full->none key flows
// carry no restricted->privileged escalation, so DIP147 must stay silent.
func TestLint_DIP147_BenignNoFalsePositive(t *testing.T) {
	src := `workflow X
  start: A
  exit: C

  agent A
    prompt: "a"
    writes: shared

  agent B
    prompt: "b"
    reads: shared

  agent C
    prompt: "c"
    tool_access: none
    reads: shared

  edges
    A -> B
    B -> C
`
	diags := lintSrc(t, src)
	if hasCode(diags, DIP147) {
		t.Errorf("DIP147 false positive on benign topology; got: %v", codes(diags))
	}
}

// TestLint_DIP147_StillFiresWhenSinkHasLastResponseTruncate: truncation bounds
// payload SIZE, not the EXISTENCE of a laundered information flow — a tiny
// malicious payload fits within any cap. Suppressing DIP147 when a sink sets
// last_response_truncate would be fail-open. This guards that DIP148
// (lintLastResponseTruncate) and DIP147 (lintChainAttack) stay independent: the
// same none->full key-laundering fixture must still fire DIP147 even when the
// sink mitigates with last_response_truncate.
func TestLint_DIP147_StillFiresWhenSinkHasLastResponseTruncate(t *testing.T) {
	src := `workflow X
  start: Summarize
  exit: Writer

  agent Summarize
    prompt: "summarize the untrusted input"
    tool_access: none
    writes: tainted

  agent Writer
    prompt: "write the report"
    last_response_truncate: 100
    reads: tainted

  edges
    Summarize -> Writer
`
	diags := lintSrc(t, src)
	if !hasCode(diags, DIP147) {
		t.Errorf("DIP147 must still fire when sink sets last_response_truncate (truncation is not a full fix); got: %v", codes(diags))
	}
}
