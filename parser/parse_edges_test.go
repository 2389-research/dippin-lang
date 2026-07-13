package parser

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// buildEdgeDip produces a minimal three-node workflow whose single edge line
// is the given text (e.g. "A -> B bogus: true").
func buildEdgeDip(edgeLine string) string {
	return "workflow X\n" +
		"  goal: \"Test edges\"\n" +
		"  start: A\n" +
		"  exit: C\n" +
		"\n" +
		"  agent A\n" +
		"    prompt: \"Do A.\"\n" +
		"\n" +
		"  agent B\n" +
		"    prompt: \"Do B.\"\n" +
		"\n" +
		"  agent C\n" +
		"    prompt: \"Do C.\"\n" +
		"\n" +
		"  edges\n" +
		"    " + edgeLine + "\n" +
		"    B -> C\n"
}

// buildOnDip produces a workflow whose first edge originates from a node of the
// given kind, so `on` desugaring can be exercised per source-node channel.
// kindBlock is the node declaration for A (e.g. an agent, tool, or parallel).
func buildOnDip(kindBlock, edgeLine string) string {
	return "workflow X\n" +
		"  goal: \"Test on\"\n" +
		"  start: A\n" +
		"  exit: C\n" +
		"\n" +
		kindBlock + "\n" +
		"\n" +
		"  agent B\n" +
		"    prompt: \"Do B.\"\n" +
		"\n" +
		"  agent C\n" +
		"    prompt: \"Do C.\"\n" +
		"\n" +
		"  edges\n" +
		"    " + edgeLine + "\n" +
		"    B -> C\n"
}

const onAgentA = "  agent A\n    prompt: \"Do A.\""
const onHumanA = "  human A\n    mode: freeform"
const onToolMarkerA = "  tool A\n    command: \"echo hi\"\n    marker_grep: \"^(ok|bad)$\""
const onToolNoMarkerA = "  tool A\n    command: \"echo hi\""
const onParallelA = "  parallel A\n    targets: B, C"

// TestParseOnAgentSource: `on X` from an agent source desugars to the outcome channel.
func TestParseOnAgentSource(t *testing.T) {
	p := NewParser(buildOnDip(onAgentA, "A -> B on success"), "test.dip")
	w, err := p.Parse()
	if err != nil {
		t.Fatalf("unexpected parse error: %v (%v)", err, p.Diagnostics())
	}
	got := w.Edges[0].Condition
	if got == nil || got.Raw != "ctx.outcome = success" {
		t.Fatalf("Condition = %+v, want Raw %q", got, "ctx.outcome = success")
	}
}

// TestParseOnToolMarkerSource: a tool with marker_grep routes on ctx.tool_marker.
func TestParseOnToolMarkerSource(t *testing.T) {
	p := NewParser(buildOnDip(onToolMarkerA, "A -> B on tests_green"), "test.dip")
	w, err := p.Parse()
	if err != nil {
		t.Fatalf("unexpected parse error: %v (%v)", err, p.Diagnostics())
	}
	if got := w.Edges[0].Condition; got == nil || got.Raw != "ctx.tool_marker = tests_green" {
		t.Fatalf("Condition = %+v, want Raw %q", got, "ctx.tool_marker = tests_green")
	}
}

// TestParseOnHyphenatedToken: a hyphenated marker is one identifier token.
func TestParseOnHyphenatedToken(t *testing.T) {
	p := NewParser(buildOnDip(onToolMarkerA, "A -> B on setup-failed"), "test.dip")
	w, err := p.Parse()
	if err != nil {
		t.Fatalf("unexpected parse error: %v (%v)", err, p.Diagnostics())
	}
	if got := w.Edges[0].Condition; got == nil || got.Raw != "ctx.tool_marker = setup-failed" {
		t.Fatalf("Condition = %+v, want Raw %q", got, "ctx.tool_marker = setup-failed")
	}
}

// TestParseOnNonBareTokenDiagnoses: an `on` value that is not a bare identifier
// must be diagnosed rather than silently desugared, keeping the parser's accept
// rule symmetric with what the formatter emits. This covers single-token cases
// (`a.b`/`a/b` keep `.`/`/`; a quoted literal) and the lexer-split case
// (`a:b` → a · : · b), which must yield one clear diagnostic pointing at `when`
// — not the value `a` plus stray "unknown edge attribute" errors for `:`/`b`.
func TestParseOnNonBareTokenDiagnoses(t *testing.T) {
	for _, tok := range []string{"a.b", "a/b", `"quoted"`, "a:b"} {
		t.Run(tok, func(t *testing.T) {
			p := NewParser(buildOnDip(onAgentA, "A -> B on "+tok), "test.dip")
			w, err := p.Parse()
			if err == nil {
				t.Fatalf("expected parse error for non-bare `on` value %q", tok)
			}
			joined := strings.Join(p.Diagnostics(), "\n")
			if !strings.Contains(joined, "`on`") || !strings.Contains(joined, "when") {
				t.Errorf("expected diagnostic mentioning `on` and `when`, got: %v", p.Diagnostics())
			}
			// The split value must be reported as a whole, not leak its tail.
			if strings.Contains(joined, "unknown edge attribute") {
				t.Errorf("split `on` value leaked to the attribute loop: %v", p.Diagnostics())
			}
			// No usable condition should have been attached from a bad value.
			if c := w.Edges[0].Condition; c != nil {
				t.Errorf("expected no condition for invalid `on` value, got %+v", c)
			}
		})
	}
}

// TestParseOnIdenticalToWhen: `on X` must produce the same Condition as the
// equivalent `when`, and coexist with trailing attributes.
func TestParseOnIdenticalToWhen(t *testing.T) {
	pOn := NewParser(buildOnDip(onAgentA, "A -> B on fail  label: retry"), "test.dip")
	wOn, err := pOn.Parse()
	if err != nil {
		t.Fatalf("on parse error: %v (%v)", err, pOn.Diagnostics())
	}
	pWhen := NewParser(buildOnDip(onAgentA, "A -> B when ctx.outcome = fail  label: retry"), "test.dip")
	wWhen, err := pWhen.Parse()
	if err != nil {
		t.Fatalf("when parse error: %v (%v)", err, pWhen.Diagnostics())
	}
	if wOn.Edges[0].Condition.Raw != wWhen.Edges[0].Condition.Raw {
		t.Errorf("on Raw %q != when Raw %q", wOn.Edges[0].Condition.Raw, wWhen.Edges[0].Condition.Raw)
	}
	if wOn.Edges[0].Label != "retry" {
		t.Errorf("label = %q, want retry", wOn.Edges[0].Label)
	}
}

// TestParseOnNoChannelDiagnoses: `on` on a node with no defined outcome channel
// (parallel, or tool without marker_grep) is a located diagnostic suggesting when.
func TestParseOnNoChannelDiagnoses(t *testing.T) {
	for _, tc := range []struct{ name, block string }{
		{"parallel", onParallelA},
		{"tool-no-marker", onToolNoMarkerA},
		{"human", onHumanA}, // human gates route on labels, not ctx.outcome (#130)
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := NewParser(buildOnDip(tc.block, "A -> B on whatever"), "test.dip")
			_, err := p.Parse()
			if err == nil {
				t.Fatal("expected parse error for `on` without a channel")
			}
			joined := strings.Join(p.Diagnostics(), "\n")
			if !strings.Contains(joined, "`on`") || !strings.Contains(joined, "when") {
				t.Errorf("expected diagnostic mentioning `on` and `when`, got: %v", p.Diagnostics())
			}
			if !strings.Contains(joined, "16:") {
				t.Errorf("expected located diagnostic (line 16), got: %v", p.Diagnostics())
			}
		})
	}
}

// TestParseOnUnknownSourceDefersToValidator: an `on` edge from an undeclared
// source node must parse cleanly (no parse diagnostic), exactly as the
// equivalent `when` edge does, so the validator's DIP003 owns the unknown-node
// report rather than a misleading channel diagnostic emitted at parse time.
// (The conditions differ by design: `when` carries its explicit condition,
// while `on` attaches none since the channel can't be resolved for an unknown
// node — but neither should produce a parse diagnostic.)
func TestParseOnUnknownSourceDefersToValidator(t *testing.T) {
	for _, line := range []string{"Nope -> B on success", "Nope -> B when ctx.outcome = success"} {
		t.Run(line, func(t *testing.T) {
			p := NewParser(buildOnDip(onAgentA, line), "test.dip")
			if _, err := p.Parse(); err != nil {
				t.Fatalf("expected clean parse (DIP003 is the validator's job), got: %v", p.Diagnostics())
			}
			if len(p.Diagnostics()) != 0 {
				t.Errorf("expected no parse diagnostics, got: %v", p.Diagnostics())
			}
		})
	}
	// The `on` form attaches no condition when the source is unknown.
	p := NewParser(buildOnDip(onAgentA, "Nope -> B on success"), "test.dip")
	w, _ := p.Parse()
	if c := w.Edges[0].Condition; c != nil {
		t.Errorf("expected no condition on `on` edge from unknown source, got %+v", c)
	}
}

// TestParseOnNoChannelGluedValueNoLeak guards the no-channel error path against
// the same lexer-split leak as the value path: `on a:b` on a node without an
// outcome channel must emit only the no-channel diagnostic and consume the whole
// glued value, not leave `:`/`b` to cascade as "unknown edge attribute" errors.
func TestParseOnNoChannelGluedValueNoLeak(t *testing.T) {
	p := NewParser(buildOnDip(onHumanA, "A -> B on a:b"), "test.dip")
	_, err := p.Parse()
	if err == nil {
		t.Fatal("expected parse error for `on` without a channel")
	}
	joined := strings.Join(p.Diagnostics(), "\n")
	if !strings.Contains(joined, "`on`") {
		t.Errorf("expected `on` no-channel diagnostic, got: %v", p.Diagnostics())
	}
	if strings.Contains(joined, "unknown edge attribute") {
		t.Errorf("glued `on` value leaked to the attribute loop: %v", p.Diagnostics())
	}
}

// TestParseUnknownEdgeAttributeDiagnoses covers #126(a): an unrecognized edge
// attribute must become a single located parse diagnostic, not be silently
// swallowed, and must not fire once per token (name / ':' / value).
func TestParseUnknownEdgeAttributeDiagnoses(t *testing.T) {
	p := NewParser(buildEdgeDip("A -> B bogus: true"), "test.dip")
	w, err := p.Parse()
	if err == nil {
		t.Fatal("expected parse error for unknown edge attribute, got nil")
	}
	diags := p.Diagnostics()

	count := 0
	for _, d := range diags {
		if strings.Contains(d, "bogus") {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("expected exactly one diagnostic mentioning 'bogus', got %d: %v", count, diags)
	}
	joined := strings.Join(diags, "\n")
	if !strings.Contains(joined, "unknown edge attribute") {
		t.Errorf("expected an 'unknown edge attribute' diagnostic, got: %v", diags)
	}
	// Located: the diagnostic carries a line:column.
	if !strings.Contains(joined, "16:") {
		t.Errorf("expected diagnostic to carry the attribute's line (16), got: %v", diags)
	}

	// Subsequent edges still parse after the diagnostic.
	found := false
	for _, e := range w.Edges {
		if e.From == "B" && e.To == "C" {
			found = true
		}
	}
	if !found {
		t.Error("expected edge B -> C to still parse after unknown-attr diagnostic")
	}
}

// TestParseUnknownBareEdgeAttributeDiagnosesOnce covers an unknown attribute
// with no ": value" payload (a bare word), which must still fire exactly once.
func TestParseUnknownBareEdgeAttributeDiagnosesOnce(t *testing.T) {
	p := NewParser(buildEdgeDip("A -> B bogus"), "test.dip")
	_, err := p.Parse()
	if err == nil {
		t.Fatal("expected parse error for unknown bare edge attribute, got nil")
	}
	count := 0
	for _, d := range p.Diagnostics() {
		if strings.Contains(d, "bogus") {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("expected exactly one diagnostic mentioning 'bogus', got %d: %v", count, p.Diagnostics())
	}
}

// TestParseUnknownEdgeAttributeMissingValue covers the no-value form
// ("bogus:" with nothing after the colon): the diagnostic must fire once and
// the parser must not consume the newline, so the following edge still parses.
func TestParseUnknownEdgeAttributeMissingValue(t *testing.T) {
	p := NewParser(buildEdgeDip("A -> B bogus:"), "test.dip")
	w, err := p.Parse()
	if err == nil {
		t.Fatal("expected parse error for unknown edge attribute, got nil")
	}
	count := 0
	for _, d := range p.Diagnostics() {
		if strings.Contains(d, "bogus") {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("expected exactly one diagnostic mentioning 'bogus', got %d: %v", count, p.Diagnostics())
	}
	found := false
	for _, e := range w.Edges {
		if e.From == "B" && e.To == "C" {
			found = true
		}
	}
	if !found {
		t.Error("expected edge B -> C to still parse after a value-less unknown attr")
	}
}

// TestParseConditionWithBareKeywordRHS covers #126(b): a condition whose bare
// unquoted right-hand value is an attribute keyword must NOT truncate the
// condition. The keyword only terminates a condition when followed by ':'.
func TestParseConditionWithBareKeywordRHS(t *testing.T) {
	for _, kw := range []string{"override", "restart", "label", "weight"} {
		t.Run(kw, func(t *testing.T) {
			p := NewParser(buildEdgeDip("A -> B when ctx.reason = "+kw), "test.dip")
			w, err := p.Parse()
			if err != nil {
				t.Fatalf("unexpected parse error: %v", err)
			}
			var edge = w.Edges[0]
			if edge.Condition == nil {
				t.Fatalf("expected a condition on edge A->B, got nil")
			}
			want := "ctx.reason = " + kw
			if edge.Condition.Raw != want {
				t.Errorf("condition Raw = %q, want %q", edge.Condition.Raw, want)
			}
		})
	}
}

// TestParseKeywordAttributesStillTerminateCondition guards the other side of
// #126(b): a genuine attribute (keyword followed by ':') still terminates the
// condition rather than being absorbed into it.
func TestParseKeywordAttributesStillTerminateCondition(t *testing.T) {
	p := NewParser(buildEdgeDip(`A -> B when ctx.x == "ok" label: approved weight: 5`), "test.dip")
	w, err := p.Parse()
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	edge := w.Edges[0]
	if edge.Condition == nil || edge.Condition.Raw != `ctx.x == "ok"` {
		t.Errorf("condition = %+v, want Raw %q", edge.Condition, `ctx.x == "ok"`)
	}
	if edge.Label != "approved" {
		t.Errorf("label = %q, want %q", edge.Label, "approved")
	}
	if edge.Weight != 5 {
		t.Errorf("weight = %d, want 5", edge.Weight)
	}
}

// TestParseLoopKeyword: the bare `loop` keyword sets ir.Edge.Restart, consuming
// no colon or value, and emits no diagnostics.
func TestParseLoopKeyword(t *testing.T) {
	p := NewParser(buildEdgeDip("A -> B loop"), "test.dip")
	w, err := p.Parse()
	if err != nil {
		t.Fatalf("unexpected parse error: %v (%v)", err, p.Diagnostics())
	}
	if !w.Edges[0].Restart {
		t.Errorf("Restart = false, want true for `A -> B loop`")
	}
}

// TestParseRestartTrueBackCompat: the legacy `restart: true` form still parses to
// the same Restart field, so existing files keep working.
func TestParseRestartTrueBackCompat(t *testing.T) {
	p := NewParser(buildEdgeDip("A -> B restart: true"), "test.dip")
	w, err := p.Parse()
	if err != nil {
		t.Fatalf("unexpected parse error: %v (%v)", err, p.Diagnostics())
	}
	if !w.Edges[0].Restart {
		t.Errorf("Restart = false, want true for `A -> B restart: true`")
	}
}

// TestParseLoopWithValueDiagnoses: writing `loop` in the legacy boolean shape
// `loop: true` still sets Restart, but the stray `: value` is reported once with a
// clear hint — not leaked as spurious "unknown edge attribute" errors for `:`/`true`.
func TestParseLoopWithValueDiagnoses(t *testing.T) {
	p := NewParser(buildEdgeDip("A -> B loop: true"), "test.dip")
	w, _ := p.Parse()
	if !w.Edges[0].Restart {
		t.Errorf("Restart = false, want true even for the mistaken `loop: true`")
	}
	joined := strings.Join(p.Diagnostics(), "\n")
	if !strings.Contains(joined, "`loop` takes no value") {
		t.Errorf("expected a clear `loop takes no value` diagnostic, got: %v", p.Diagnostics())
	}
	if strings.Contains(joined, "unknown edge attribute") {
		t.Errorf("`loop: true` leaked stray tokens as unknown-attribute errors: %v", p.Diagnostics())
	}
}

// TestParseLoopTerminatesCondition guards the design trap: a value-less `loop`
// keyword must terminate a preceding condition unconditionally (it has no ':' to
// gate on), so it is recorded as Restart rather than absorbed into the condition
// Raw. Both the `on` shorthand and a full `when` must behave identically.
func TestParseLoopTerminatesCondition(t *testing.T) {
	cases := map[string]string{
		"on":   "A -> B on fail loop",
		"when": "A -> B when ctx.outcome = fail loop",
	}
	for name, line := range cases {
		t.Run(name, func(t *testing.T) {
			p := NewParser(buildOnDip(onAgentA, line), "test.dip")
			w, err := p.Parse()
			if err != nil {
				t.Fatalf("unexpected parse error: %v (%v)", err, p.Diagnostics())
			}
			edge := w.Edges[0]
			if !edge.Restart {
				t.Errorf("Restart = false, want true for %q", line)
			}
			if edge.Condition == nil || edge.Condition.Raw != "ctx.outcome = fail" {
				t.Fatalf("condition Raw = %+v, want %q (loop must not leak in)",
					edge.Condition, "ctx.outcome = fail")
			}
			if strings.Contains(edge.Condition.Raw, "loop") {
				t.Errorf("condition Raw %q absorbed the `loop` terminator", edge.Condition.Raw)
			}
		})
	}
}

// TestParseLoopIsReservedOnConditionRHS pins the deliberate trade-off: because
// `loop` terminates a condition unconditionally, it cannot be a bare unquoted
// right-hand value — it is consumed as the back-edge flag. The condition's RHS
// must be quoted to mean the literal string "loop".
func TestParseLoopIsReservedOnConditionRHS(t *testing.T) {
	// Bare `loop` on the RHS is taken as the flag, truncating the condition.
	p := NewParser(buildEdgeDip("A -> B when ctx.x = loop"), "test.dip")
	w, err := p.Parse()
	if err != nil {
		t.Fatalf("unexpected parse error: %v (%v)", err, p.Diagnostics())
	}
	edge := w.Edges[0]
	if !edge.Restart {
		t.Errorf("Restart = false, want true: bare `loop` RHS is the flag")
	}
	if edge.Condition == nil || edge.Condition.Raw != "ctx.x =" {
		t.Errorf("condition Raw = %+v, want %q", edge.Condition, "ctx.x =")
	}

	// Quoting keeps `loop` as the literal value, with no back-edge.
	pq := NewParser(buildEdgeDip(`A -> B when ctx.x = "loop"`), "test.dip")
	wq, err := pq.Parse()
	if err != nil {
		t.Fatalf("unexpected parse error: %v (%v)", err, pq.Diagnostics())
	}
	edgeq := wq.Edges[0]
	if edgeq.Restart {
		t.Errorf("Restart = true, want false for quoted `\"loop\"` RHS")
	}
	if edgeq.Condition == nil || edgeq.Condition.Raw != `ctx.x = "loop"` {
		t.Errorf("condition Raw = %+v, want %q", edgeq.Condition, `ctx.x = "loop"`)
	}
}

// TestExistingDipFilesStillParse covers #126(c): the new diagnostics must not
// regress any currently-valid .dip file. Every examples/*.dip and
// parser/testdata/*.dip must still parse without error.
func TestExistingDipFilesStillParse(t *testing.T) {
	dirs := []string{"../examples", "testdata"}
	var files []string
	for _, d := range dirs {
		matches, err := filepath.Glob(filepath.Join(d, "*.dip"))
		if err != nil {
			t.Fatalf("glob %s: %v", d, err)
		}
		files = append(files, matches...)
	}
	if len(files) == 0 {
		t.Fatal("no .dip files found to guard against regression")
	}
	for _, f := range files {
		f := f
		t.Run(filepath.Base(f), func(t *testing.T) {
			src, err := os.ReadFile(f)
			if err != nil {
				t.Fatalf("read %s: %v", f, err)
			}
			p := NewParser(string(src), f)
			if _, err := p.Parse(); err != nil {
				t.Errorf("%s no longer parses: %v", f, err)
			}
		})
	}
}

// The following edge-condition quote cases are salvaged from PR #183 (thanks
// @harperreed) — its coverage caught a real comment-stripping/backslash-parity
// gap in the #182 fix (advanceInDoubleQuote was not escape-aware).

func TestParseConditionPreservesHashInsideEscapedQuotes(t *testing.T) {
	w, err := NewParser(buildEdgeDip(`A -> B when ctx.tool_stdout = "say \"alpha # beta\""`), "test.dip").Parse()
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if got, want := w.Edges[0].Condition.Raw, `ctx.tool_stdout = "say \"alpha # beta\""`; got != want {
		t.Fatalf("Raw = %q, want %q (hash inside escaped quotes must survive)", got, want)
	}
}

func TestParseConditionStripsCommentAfterEscapedQuoteLiteral(t *testing.T) {
	w, err := NewParser(buildEdgeDip(`A -> B when ctx.x = "say \"alpha\"" # trailing`), "test.dip").Parse()
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if got, want := w.Edges[0].Condition.Raw, `ctx.x = "say \"alpha\""`; got != want {
		t.Fatalf("Raw = %q, want comment-free %q", got, want)
	}
}

func TestParseConditionBackslashParityControlsQuoteClosure(t *testing.T) {
	cases := []struct{ name, edge, want string }{
		{"even run closes quote", `A -> B when ctx.x = "path\\" # trailing`, `ctx.x = "path\\"`},
		{"odd run escapes quote", `A -> B when ctx.x = "path\\\"quoted" # trailing`, `ctx.x = "path\\\"quoted"`},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			w, err := NewParser(buildEdgeDip(c.edge), "test.dip").Parse()
			if err != nil {
				t.Fatalf("parse: %v", err)
			}
			if got := w.Edges[0].Condition.Raw; got != c.want {
				t.Fatalf("Raw = %q, want %q", got, c.want)
			}
		})
	}
}

func TestParseConditionKeepsSingleQuoteNormalization(t *testing.T) {
	w, err := NewParser(buildEdgeDip(`A -> B when ctx.tool_stdout = 'say alpha||beta'`), "test.dip").Parse()
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if got, want := w.Edges[0].Condition.Raw, `ctx.tool_stdout = "say alpha||beta"`; got != want {
		t.Fatalf("Raw = %q, want normalized single-quoted literal %q", got, want)
	}
}

func TestParseConditionEscapesNormalizedSingleQuoteValue(t *testing.T) {
	w, err := NewParser(buildEdgeDip(`A -> B when ctx.x = 'it''s \d+ here'`), "test.dip").Parse()
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if got, want := w.Edges[0].Condition.Raw, `ctx.x = "it's \\d+ here"`; got != want {
		t.Fatalf("Raw = %q, want escaped double-quoted normalization %q", got, want)
	}
}

func TestParseConditionPreservesLiteralTabWhenNormalizing(t *testing.T) {
	w, err := NewParser(buildEdgeDip("A -> B when ctx.x = 'a\tcafé'"), "test.dip").Parse()
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if got, want := w.Edges[0].Condition.Raw, "ctx.x = \"a\tcafé\""; got != want {
		t.Fatalf("Raw = %q, want literal tab + non-ASCII %q", got, want)
	}
}
