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
