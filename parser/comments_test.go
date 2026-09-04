package parser_test

import (
	"strings"
	"testing"

	"github.com/2389-research/dippin-lang/formatter"
	"github.com/2389-research/dippin-lang/ir"
	"github.com/2389-research/dippin-lang/parser"
)

func parseMust(t *testing.T, input string) *ir.Workflow {
	t.Helper()
	w, err := parser.NewParser(input, "test.dip").Parse()
	if err != nil {
		t.Fatalf("parse failed: %v\nsource:\n%s", err, input)
	}
	return w
}

// TestAttachNodeHeaderComment verifies whole-line comments immediately above a
// node declaration become that node's HeaderComment, in source order (#259).
func TestAttachNodeHeaderComment(t *testing.T) {
	w := parseMust(t, `workflow T
  start: A
  exit: B

  # first line of the block
  # second line of the block
  agent A
    prompt: "hi"
  agent B
    prompt: "bye"
`)
	a := w.Node("A")
	b := w.Node("B")
	if a == nil || b == nil {
		t.Fatalf("missing nodes: %#v", w.Nodes)
	}
	want := []string{"# first line of the block", "# second line of the block"}
	if len(a.HeaderComment) != len(want) {
		t.Fatalf("A.HeaderComment = %#v, want %#v", a.HeaderComment, want)
	}
	for i := range want {
		if a.HeaderComment[i] != want[i] {
			t.Errorf("A.HeaderComment[%d] = %q, want %q", i, a.HeaderComment[i], want[i])
		}
	}
	if len(b.HeaderComment) != 0 {
		t.Errorf("B.HeaderComment = %#v, want empty (the block belongs to A, which follows it)", b.HeaderComment)
	}
}

// TestAttachNodeBodyComments verifies trailing inline and mid-body whole-line
// comments inside a node become its BodyComments, in source order (#259).
func TestAttachNodeBodyComments(t *testing.T) {
	w := parseMust(t, `workflow T
  start: A
  exit: A

  agent A
    label: A        # trailing on label
    # mid-body note
    prompt: "hi"  # trailing on prompt
`)
	a := w.Node("A")
	want := []string{"# trailing on label", "# mid-body note", "# trailing on prompt"}
	if len(a.BodyComments) != len(want) {
		t.Fatalf("A.BodyComments = %#v, want %#v", a.BodyComments, want)
	}
	for i := range want {
		if a.BodyComments[i] != want[i] {
			t.Errorf("A.BodyComments[%d] = %q, want %q", i, a.BodyComments[i], want[i])
		}
	}
}

// TestAttachNodeDeclLineTrailingComment verifies the trailing comment on the
// node declaration line itself is retained, on the node (#259).
func TestAttachNodeDeclLineTrailingComment(t *testing.T) {
	w := parseMust(t, `workflow T
  start: A
  exit: A

  agent A  # entry point
    prompt: "hi"
`)
	a := w.Node("A")
	want := []string{"# entry point"}
	if len(a.BodyComments) != len(want) || a.BodyComments[0] != want[0] {
		t.Fatalf("A.BodyComments = %#v, want %#v", a.BodyComments, want)
	}
	if len(a.HeaderComment) != 0 {
		t.Errorf("A.HeaderComment = %#v, want empty (the comment is on the declaration line)", a.HeaderComment)
	}
}

// TestAttachHeaderRunBreaksAtBlank verifies a blank line between the comment
// and the node breaks the header run: the comment attaches to the preceding
// node's body instead (#259).
func TestAttachHeaderRunBreaksAtBlank(t *testing.T) {
	w := parseMust(t, `workflow T
  start: A
  exit: B

  agent A
    prompt: "hi"
  # orphaned by the blank line

  agent B
    prompt: "bye"
`)
	a := w.Node("A")
	b := w.Node("B")
	if len(b.HeaderComment) != 0 {
		t.Errorf("B.HeaderComment = %#v, want empty (blank line breaks the run)", b.HeaderComment)
	}
	want := []string{"# orphaned by the blank line"}
	if len(a.BodyComments) != len(want) || a.BodyComments[0] != want[0] {
		t.Errorf("A.BodyComments = %#v, want %#v", a.BodyComments, want)
	}
}

// TestAttachEdgeComments verifies whole-line comments above an edge line become
// the edge's HeaderComment and the trailing inline comment on the edge line
// becomes its TrailingComment (#259).
func TestAttachEdgeComments(t *testing.T) {
	w := parseMust(t, `workflow T
  start: A
  exit: B

  agent A
    prompt: "hi"
  agent B
    prompt: "bye"

  edges
    # routing note
    A -> B  # success path
`)
	e := w.Edges[0]
	if e == nil {
		t.Fatal("no edges parsed")
	}
	if len(e.HeaderComment) != 1 || e.HeaderComment[0] != "# routing note" {
		t.Errorf("edge.HeaderComment = %#v, want [\"# routing note\"]", e.HeaderComment)
	}
	if e.TrailingComment != "# success path" {
		t.Errorf("edge.TrailingComment = %q, want %q", e.TrailingComment, "# success path")
	}
}

// TestCommentInsideRawBlockIsContent verifies a `#` line inside an indented raw
// prompt block is prompt content, not a retained comment (#259).
func TestCommentInsideRawBlockIsContent(t *testing.T) {
	w := parseMust(t, `workflow T
  start: A
  exit: A

  agent A
    prompt:
      line one
      # not a comment
`)
	a := w.Node("A")
	cfg, ok := a.Config.(ir.AgentConfig)
	if !ok {
		t.Fatalf("A config is %T, want ir.AgentConfig", a.Config)
	}
	if !strings.Contains(cfg.Prompt, "# not a comment") {
		t.Errorf("prompt lost the raw-block line:\n%q", cfg.Prompt)
	}
	if len(a.BodyComments) != 0 || len(a.HeaderComment) != 0 {
		t.Errorf("A retained comments from raw-block content: header=%#v body=%#v", a.HeaderComment, a.BodyComments)
	}
}

// TestHashInsideQuotedValueIsNotAComment verifies a # inside a quoted value is
// not treated as a comment and does not truncate the value (#259).
func TestHashInsideQuotedValueIsNotAComment(t *testing.T) {
	w := parseMust(t, `workflow T
  start: A
  exit: A

  agent A
    label: "A # one"
    prompt: "hi"
`)
	a := w.Node("A")
	if a.Label != "A # one" {
		t.Errorf("A.Label = %q, want %q", a.Label, "A # one")
	}
	if len(a.BodyComments) != 0 {
		t.Errorf("A.BodyComments = %#v, want empty", a.BodyComments)
	}
}

// TestFormatCommentsIdempotent runs the issue #259 repro through Format twice
// and requires both comment retention and byte-identical idempotency — the
// property that makes `fmt --check` usable as a CI gate.
func TestFormatCommentsIdempotent(t *testing.T) {
	input := `# ABOUTME: leading file comment
# second leading line
workflow CommentTest
  start: A
  exit: B

  # comment between defaults and nodes
  agent A
    label: A     # trailing inline comment
    prompt: "hi"
  agent B
    prompt: "bye"

  edges
    A -> B  # edge comment
`
	w := parseMust(t, input)
	first := formatter.Format(w)

	mustContain := []string{
		"# ABOUTME: leading file comment",
		"# second leading line",
		"# comment between defaults and nodes",
		"# trailing inline comment",
		"A -> B  # edge comment",
	}
	for _, want := range mustContain {
		if !strings.Contains(first, want) {
			t.Errorf("formatted output missing %q\ngot:\n%s", want, first)
		}
	}

	w2 := parseMust(t, first)
	second := formatter.Format(w2)
	if first != second {
		t.Errorf("formatter not idempotent with comments\nfirst:\n%s\nsecond:\n%s", first, second)
	}
}

// TestStructuralNodeComments verifies comment retention for inline parallel and
// fan_in nodes, whose declarations share a line with their targets (#259).
func TestStructuralNodeComments(t *testing.T) {
	w := parseMust(t, `workflow T
  start: Split
  exit: Join

  # fan out
  parallel Split -> A, B
    # split note

  agent A
    prompt: "a"
  agent B
    prompt: "b"
  fan_in Join <- A, B  # merge

  edges
    Join -> Join
`)
	split := w.Node("Split")
	join := w.Node("Join")
	if split == nil || join == nil {
		t.Fatalf("missing structural nodes: %#v", w.Nodes)
	}
	if len(split.HeaderComment) != 1 || split.HeaderComment[0] != "# fan out" {
		t.Errorf("Split.HeaderComment = %#v, want [\"# fan out\"]", split.HeaderComment)
	}
	if len(split.BodyComments) != 1 || split.BodyComments[0] != "# split note" {
		t.Errorf("Split.BodyComments = %#v, want [\"# split note\"]", split.BodyComments)
	}
	if len(join.BodyComments) != 1 || join.BodyComments[0] != "# merge" {
		t.Errorf("Join.BodyComments = %#v, want [\"# merge\"]", join.BodyComments)
	}

	formatted := formatter.Format(w)
	w2 := parseMust(t, formatted)
	if formatter.Format(w2) != formatted {
		t.Errorf("formatter not idempotent with structural-node comments\nfirst:\n%s\nsecond:\n%s", formatted, formatter.Format(w2))
	}
}
