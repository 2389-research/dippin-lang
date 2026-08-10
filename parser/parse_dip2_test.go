package parser

import "strings"
import "testing"

const dip2Body = `workflow W
  goal: "t"
  start: A
  exit: B

  agent A
    prompt: a
    %s: B

  agent B
    prompt: b

  edges
    A -> B
`

// TestDip2AcceptsRetryTarget: dip 2 re-admits retry_target as a node attr (the
// retry channel is distinct from the edges block; #204 Option A).
func TestDip2AcceptsRetryTarget(t *testing.T) {
	w, err := NewParser("dip 2\n\n"+stringsReplaceOnce(dip2Body, "%s", "retry_target"), "t.dip").Parse()
	if err != nil {
		t.Fatalf("dip 2 must accept retry_target: %v", err)
	}
	if w.Nodes[0].Retry.RetryTarget != "B" {
		t.Errorf("dip 2 RetryTarget = %q, want B", w.Nodes[0].Retry.RetryTarget)
	}
}

// TestDip2AcceptsFallbackRetryTarget: dip 2 uses fallback_retry_target for the
// retry-exhaustion route.
func TestDip2AcceptsFallbackRetryTarget(t *testing.T) {
	w, err := NewParser("dip 2\n\n"+stringsReplaceOnce(dip2Body, "%s", "fallback_retry_target"), "t.dip").Parse()
	if err != nil {
		t.Fatalf("dip 2 must accept fallback_retry_target: %v", err)
	}
	if w.Nodes[0].Retry.FallbackTarget != "B" {
		t.Errorf("dip 2 FallbackTarget = %q, want B", w.Nodes[0].Retry.FallbackTarget)
	}
}

// TestDip2RejectsFallbackTarget: the dip-1 spelling fallback_target is rejected
// in dip 2, steering to fallback_retry_target.
func TestDip2RejectsFallbackTarget(t *testing.T) {
	_, err := NewParser("dip 2\n\n"+stringsReplaceOnce(dip2Body, "%s", "fallback_target"), "t.dip").Parse()
	if err == nil || !strings.Contains(err.Error(), "fallback_retry_target") {
		t.Fatalf("want dip-2 rejection pointing to fallback_retry_target, got %v", err)
	}
}

func TestV1AcceptsRetryTarget(t *testing.T) {
	w, err := NewParser(stringsReplaceOnce(dip2Body, "%s", "retry_target"), "t.dip").Parse()
	if err != nil {
		t.Fatalf("v1 must still accept retry_target: %v", err)
	}
	if w.Nodes[0].Retry.RetryTarget != "B" {
		t.Errorf("v1 RetryTarget = %q, want B", w.Nodes[0].Retry.RetryTarget)
	}
}

func stringsReplaceOnce(s, old, new string) string { return strings.Replace(s, old, new, 1) }

const dip2FanBody = `workflow W
  start: Fan
  exit: Join

  parallel Fan -> A, B

  agent A
    prompt: a

  agent B
    prompt: b

  fan_in Join <- A, B

  edges
    Fan -> A
    Fan -> B
    A -> Join
    B -> Join
`

func TestDip2RejectsRedundantFanEdge(t *testing.T) {
	_, err := NewParser("dip 2\n\n"+dip2FanBody, "t.dip").Parse()
	if err == nil || !strings.Contains(err.Error(), "authoritative under dip 2") {
		t.Fatalf("want dip-2 rejection for redundant fan edges, got %v", err)
	}
}

func TestV1AcceptsRedundantFanEdge(t *testing.T) {
	w, err := NewParser(dip2FanBody, "t.dip").Parse()
	if err != nil {
		t.Fatalf("v1 must still accept redundant fan edges (DIP153 warns): %v", err)
	}
	if len(w.Edges) != 4 {
		t.Errorf("v1 must keep all 4 edges, got %d", len(w.Edges))
	}
}
