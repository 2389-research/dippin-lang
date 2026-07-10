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

func TestDip2RejectsRetryTarget(t *testing.T) {
	_, err := NewParser("dip 2\n\n"+stringsReplaceOnce(dip2Body, "%s", "retry_target"), "t.dip").Parse()
	if err == nil || !strings.Contains(err.Error(), "not a node field in dip 2") {
		t.Fatalf("want dip-2 rejection error for retry_target, got %v", err)
	}
}

func TestDip2RejectsFallbackTarget(t *testing.T) {
	_, err := NewParser("dip 2\n\n"+stringsReplaceOnce(dip2Body, "%s", "fallback_target"), "t.dip").Parse()
	if err == nil || !strings.Contains(err.Error(), "not a node field in dip 2") {
		t.Fatalf("want dip-2 rejection error for fallback_target, got %v", err)
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
