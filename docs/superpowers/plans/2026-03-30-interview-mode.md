# Interview Mode Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add `mode: interview` to human nodes so the runtime can parse upstream agent questions into individual huh form fields with suggested options + freeform escape hatch.

**Architecture:** Minimal DSL change — add `"interview"` as a valid mode plus two optional context key fields (`questions_key`, `answers_key`) to `HumanConfig`. Three new lint rules (DIP127-DIP129) catch misconfiguration. Simulator handles interview like multi-line freeform. All heuristic/rendering logic stays in the runtime.

**Tech Stack:** Go, dippin-lang IR/parser/validator/formatter/simulator

---

## File Map

| File | Action | Responsibility |
|------|--------|----------------|
| `ir/ir.go:100-107` | Modify | Add `QuestionsKey`, `AnswersKey` to `HumanConfig` |
| `parser/parse_nodes.go:230-240` | Modify | Parse `questions_key` and `answers_key` fields |
| `validator/lint_codes.go:4-31,33-61` | Modify | Register DIP127, DIP128, DIP129 |
| `validator/lint_human.go` | Create | Three lint functions for interview mode |
| `validator/explanations.go:296-303` | Modify | Add DIP127-DIP129 explanations |
| `validator/lint.go:17-52` | Modify | Wire new lint functions into `Lint()` |
| `validator/lint_test.go` | Modify | Tests for DIP127-DIP129 |
| `formatter/format.go:356-368` | Modify | Emit `questions_key` and `answers_key` fields |
| `simulate/interactive.go:11-54` | Modify | Handle interview mode prompting |
| `simulate/simulate_test.go` | Modify | Tests for interview mode simulation |
| `parser/testdata/human_interview.dip` | Create | Parser round-trip test fixture |
| `examples/api_design.dip:49-51` | Modify | Switch `AnswerQuestions` to `mode: interview` |
| `docs/nodes.md:149-199` | Modify | Document interview mode |
| `docs/integration.md:155-175` | Modify | Update type switch example and adapter guidance |
| `docs/validation.md` | Modify | Document DIP127-DIP129 |
| `CHANGELOG.md` | Modify | Add v0.15.0 entry |
| `site/changelog.html` | Auto-generated | Via `scripts/gen-changelog-html.sh` |

---

### Task 1: IR — Add interview fields to HumanConfig

**Files:**
- Modify: `ir/ir.go:100-107`

- [ ] **Step 1: Write the failing test**

No standalone IR test needed — the IR is a plain struct. The parser and formatter tests in later tasks cover serialization. Proceed to implementation.

- [ ] **Step 2: Add fields to HumanConfig**

```go
// HumanConfig holds configuration for human gate nodes.
type HumanConfig struct {
	Mode         string // "choice" | "freeform" | "interview"
	Default      string // Default choice
	Prompt       string // Instructions shown to the human
	QuestionsKey string // Context key to read questions from (interview mode)
	AnswersKey   string // Context key to write answers to (interview mode)
}
```

- [ ] **Step 3: Run tests to verify nothing breaks**

Run: `just test`
Expected: All existing tests pass (no code references the new fields yet).

- [ ] **Step 4: Commit**

```bash
git add ir/ir.go
git commit -m "feat(ir): add QuestionsKey and AnswersKey to HumanConfig for interview mode"
```

---

### Task 2: Parser — Accept new human fields

**Files:**
- Modify: `parser/parse_nodes.go:230-240`
- Create: `parser/testdata/human_interview.dip`

- [ ] **Step 1: Create the test fixture**

Create `parser/testdata/human_interview.dip`:
```dip
workflow HumanInterview
  goal: "Test interview mode parsing"
  start: Agent
  exit: Done

  agent Agent
    writes: interview_questions
    prompt: Ask questions.

  human Gate
    label: "Answer the questions"
    mode: interview
    questions_key: interview_questions
    answers_key: interview_answers
    reads: interview_questions
    writes: interview_answers, human_response
    prompt:
      If no questions were detected, describe your
      requirements here instead.

  agent Done
    reads: interview_answers
    prompt: Proceed.

  edges
    Agent -> Gate
    Gate -> Done
```

- [ ] **Step 2: Write the parser round-trip test**

Add to `parser/parser_test.go` (follow the existing `TestParseHumanNode` pattern):

```go
func TestParseHumanInterview(t *testing.T) {
	w := parseFixture(t, "human_interview.dip")
	gate := findNode(t, w, "Gate")
	cfg, ok := gate.Config.(ir.HumanConfig)
	if !ok {
		t.Fatal("Gate is not HumanConfig")
	}
	if cfg.Mode != "interview" {
		t.Errorf("Mode = %q, want interview", cfg.Mode)
	}
	if cfg.QuestionsKey != "interview_questions" {
		t.Errorf("QuestionsKey = %q, want interview_questions", cfg.QuestionsKey)
	}
	if cfg.AnswersKey != "interview_answers" {
		t.Errorf("AnswersKey = %q, want interview_answers", cfg.AnswersKey)
	}
	if cfg.Prompt == "" {
		t.Error("expected non-empty prompt")
	}
}
```

- [ ] **Step 3: Run test to verify it fails**

Run: `just test-pkg parser`
Expected: FAIL — `questions_key` and `answers_key` not handled yet.

- [ ] **Step 4: Add field handling in applyHumanField**

In `parser/parse_nodes.go`, update the `applyHumanField` function:

```go
func (p *Parser) applyHumanField(cfg *ir.HumanConfig, key, val string, loc ir.SourceLocation) {
	switch key {
	case "mode":
		cfg.Mode = val
	case "default":
		cfg.Default = val
	case "prompt":
		cfg.Prompt = val
	case "questions_key":
		cfg.QuestionsKey = val
	case "answers_key":
		cfg.AnswersKey = val
	}
}
```

- [ ] **Step 5: Run test to verify it passes**

Run: `just test-pkg parser`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add parser/parse_nodes.go parser/testdata/human_interview.dip parser/parser_test.go
git commit -m "feat(parser): parse questions_key and answers_key on human nodes"
```

---

### Task 3: Formatter — Emit new fields

**Files:**
- Modify: `formatter/format.go:356-368`

- [ ] **Step 1: Write the round-trip test**

Add to `formatter/format_test.go` (follow existing pattern — parse a fixture, format it, re-parse, compare):

```go
func TestFormatHumanInterview(t *testing.T) {
	w := parseFixture(t, "human_interview.dip")
	got := Format(w)
	w2 := parseString(t, got)
	gate := findNode(t, w2, "Gate")
	cfg := gate.Config.(ir.HumanConfig)
	if cfg.QuestionsKey != "interview_questions" {
		t.Errorf("round-trip QuestionsKey = %q", cfg.QuestionsKey)
	}
	if cfg.AnswersKey != "interview_answers" {
		t.Errorf("round-trip AnswersKey = %q", cfg.AnswersKey)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `just test-pkg formatter`
Expected: FAIL — new fields not emitted.

- [ ] **Step 3: Update writeHumanFields**

In `formatter/format.go`, update `writeHumanFields`:

```go
func writeHumanFields(wr *writer, n *ir.Node, cfg ir.HumanConfig) {
	writeCommonNodeFields(wr, n)
	if cfg.Mode != "" {
		wr.line("mode: %s", quoteValue(cfg.Mode))
	}
	if cfg.Default != "" {
		wr.line("default: %s", quoteValue(cfg.Default))
	}
	if cfg.QuestionsKey != "" {
		wr.line("questions_key: %s", quoteValue(cfg.QuestionsKey))
	}
	if cfg.AnswersKey != "" {
		wr.line("answers_key: %s", quoteValue(cfg.AnswersKey))
	}
	writeIOFields(wr, n)
	if cfg.Prompt != "" {
		wr.multilineBlock("prompt", cfg.Prompt)
	}
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `just test-pkg formatter`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add formatter/format.go formatter/format_test.go
git commit -m "feat(formatter): emit questions_key and answers_key for human nodes"
```

---

### Task 4: Lint rules — DIP127, DIP128, DIP129

**Files:**
- Modify: `validator/lint_codes.go`
- Create: `validator/lint_human.go`
- Modify: `validator/lint.go:17-52`

#### DIP127: invalid human node mode
#### DIP128: interview mode with `default` set (meaningless)
#### DIP129: interview mode with multiple labeled outgoing edges (conflicting semantics)

- [ ] **Step 1: Register the new codes**

In `validator/lint_codes.go`, add after line 31 (after DIP126):

```go
DIP127 = "DIP127" // invalid human node mode
DIP128 = "DIP128" // interview mode with meaningless default
DIP129 = "DIP129" // interview mode with conflicting choice-style edges
```

In the `init()` function, add after line 60 (after DIP126 registration):

```go
CodeDescription[DIP127] = "invalid human node mode"
CodeDescription[DIP128] = "interview mode with meaningless default value"
CodeDescription[DIP129] = "interview mode with conflicting choice-style edges"
```

- [ ] **Step 2: Write the failing tests**

Add to `validator/lint_test.go`:

```go
func TestLint_DIP127_InvalidHumanMode(t *testing.T) {
	w := cleanMinimalWorkflow()
	w.Nodes[0].Kind = ir.NodeHuman
	w.Nodes[0].Config = ir.HumanConfig{Mode: "invalid"}
	res := Lint(w)
	assertHasCode(t, res, DIP127)
}

func TestLint_DIP127_ValidModes(t *testing.T) {
	for _, mode := range []string{"choice", "freeform", "interview", ""} {
		w := cleanMinimalWorkflow()
		w.Nodes[0].Kind = ir.NodeHuman
		w.Nodes[0].Config = ir.HumanConfig{Mode: mode}
		res := Lint(w)
		assertNoCode(t, res, DIP127)
	}
}

func TestLint_DIP128_InterviewWithDefault(t *testing.T) {
	w := cleanMinimalWorkflow()
	w.Nodes[0].Kind = ir.NodeHuman
	w.Nodes[0].Config = ir.HumanConfig{Mode: "interview", Default: "yes"}
	res := Lint(w)
	assertHasCode(t, res, DIP128)
}

func TestLint_DIP128_ChoiceWithDefault_NoDiag(t *testing.T) {
	w := cleanMinimalWorkflow()
	w.Nodes[0].Kind = ir.NodeHuman
	w.Nodes[0].Config = ir.HumanConfig{Mode: "choice", Default: "yes"}
	res := Lint(w)
	assertNoCode(t, res, DIP128)
}

func TestLint_DIP129_InterviewWithLabeledEdges(t *testing.T) {
	w := cleanMinimalWorkflow()
	w.Nodes[0].Kind = ir.NodeHuman
	w.Nodes[0].Config = ir.HumanConfig{Mode: "interview"}
	// Add a second exit node and labeled edges.
	w.Nodes = append(w.Nodes, &ir.Node{
		ID: "Alt", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "Alt."},
	})
	w.Edges = []*ir.Edge{
		{From: w.Nodes[0].ID, To: w.Exit, Label: "approve"},
		{From: w.Nodes[0].ID, To: "Alt", Label: "reject"},
		{From: "Alt", To: w.Exit},
	}
	res := Lint(w)
	assertHasCode(t, res, DIP129)
}

func TestLint_DIP129_InterviewSingleEdge_NoDiag(t *testing.T) {
	w := cleanMinimalWorkflow()
	w.Nodes[0].Kind = ir.NodeHuman
	w.Nodes[0].Config = ir.HumanConfig{Mode: "interview"}
	res := Lint(w)
	assertNoCode(t, res, DIP129)
}
```

- [ ] **Step 3: Run tests to verify they fail**

Run: `just test-pkg validator`
Expected: FAIL — DIP127/DIP128/DIP129 not defined yet.

- [ ] **Step 4: Create lint_human.go**

Create `validator/lint_human.go`:

```go
package validator

import (
	"fmt"

	"github.com/2389-research/dippin-lang/ir"
)

var validHumanModes = map[string]bool{
	"choice":    true,
	"freeform":  true,
	"interview": true,
}

func lintHumanMode(w *ir.Workflow) []Diagnostic {
	var diags []Diagnostic
	for _, n := range w.Nodes {
		cfg, ok := n.Config.(ir.HumanConfig)
		if !ok || cfg.Mode == "" {
			continue
		}
		if !validHumanModes[cfg.Mode] {
			diags = append(diags, Diagnostic{
				Code:     DIP127,
				Severity: SeverityWarning,
				Message:  fmt.Sprintf("node %q has mode %q which is not a recognized human mode", n.ID, cfg.Mode),
				Location: n.Source,
				Help:     "valid modes: choice, freeform, interview",
			})
		}
	}
	return diags
}

func lintInterviewDefault(w *ir.Workflow) []Diagnostic {
	var diags []Diagnostic
	for _, n := range w.Nodes {
		cfg, ok := n.Config.(ir.HumanConfig)
		if !ok || cfg.Mode != "interview" {
			continue
		}
		if cfg.Default != "" {
			diags = append(diags, Diagnostic{
				Code:     DIP128,
				Severity: SeverityWarning,
				Message:  fmt.Sprintf("node %q is mode interview but has default %q which is ignored", n.ID, cfg.Default),
				Location: n.Source,
				Help:     "default is only meaningful for choice mode; remove it",
			})
		}
	}
	return diags
}

func lintInterviewLabeledEdges(w *ir.Workflow) []Diagnostic {
	var diags []Diagnostic
	for _, n := range w.Nodes {
		cfg, ok := n.Config.(ir.HumanConfig)
		if !ok || cfg.Mode != "interview" {
			continue
		}
		labelCount := 0
		for _, e := range w.Edges {
			if e.From == n.ID && e.Label != "" {
				labelCount++
			}
		}
		if labelCount > 1 {
			diags = append(diags, Diagnostic{
				Code:     DIP129,
				Severity: SeverityWarning,
				Message:  fmt.Sprintf("node %q is mode interview but has %d labeled edges (interview does not route by label)", n.ID, labelCount),
				Location: n.Source,
				Help:     "interview mode collects answers, not choices; use mode choice for label-based routing",
			})
		}
	}
	return diags
}
```

- [ ] **Step 5: Wire into Lint()**

In `validator/lint.go`, add after line 49 (after `lintSubgraphRef`):

```go
diags = append(diags, lintHumanMode(w)...)
diags = append(diags, lintInterviewDefault(w)...)
diags = append(diags, lintInterviewLabeledEdges(w)...)
```

- [ ] **Step 6: Run tests to verify they pass**

Run: `just test-pkg validator`
Expected: PASS

- [ ] **Step 7: Check complexity**

Run: `just complexity`
Expected: All three functions are simple loops — should be well under the cyclomatic 5 / cognitive 7 limits.

- [ ] **Step 8: Commit**

```bash
git add validator/lint_codes.go validator/lint_human.go validator/lint.go validator/lint_test.go
git commit -m "feat(lint): DIP127-DIP129 for human mode validation and interview edge cases"
```

---

### Task 5: Explanations — dippin explain DIP127/DIP128/DIP129

**Files:**
- Modify: `validator/explanations.go:296-303`

- [ ] **Step 1: Add explanations**

In `validator/explanations.go`, add after the DIP126 entry (before the closing `}`):

```go
DIP127: {
	Code:    DIP127,
	Summary: "invalid human node mode",
	Trigger: "A human node has a mode value other than choice, freeform, or interview.",
	Fix:     "Change mode to one of: choice, freeform, interview.",
	Example: "human Gate\n  mode: interactive  // invalid — did you mean interview?",
},
DIP128: {
	Code:    DIP128,
	Summary: "interview mode with meaningless default value",
	Trigger: "A human node with mode interview also has a default value. Interview mode collects answers to questions — it has no predefined choices to default to.",
	Fix:     "Remove the default field, or change mode to choice if you want label-based routing.",
	Example: "human Ask\n  mode: interview\n  default: yes  // meaningless",
},
DIP129: {
	Code:    DIP129,
	Summary: "interview mode with conflicting choice-style edges",
	Trigger: "A human node with mode interview has multiple labeled outgoing edges. Interview mode does not route by label — it collects structured answers and follows a single edge.",
	Fix:     "Remove edge labels, or change mode to choice if routing is intended.",
	Example: "human Ask\n  mode: interview\n\nedges\n  Ask -> A label: yes\n  Ask -> B label: no  // conflicting",
},
```

- [ ] **Step 2: Run tests**

Run: `just test-pkg validator`
Expected: PASS (existing `TestCmdExplain` covers the map lookup).

- [ ] **Step 3: Verify explain command works**

Run: `./dippin explain DIP127`
Expected: Shows code summary, trigger, fix, and example.

Run: `./dippin explain DIP128`
Run: `./dippin explain DIP129`

- [ ] **Step 4: Commit**

```bash
git add validator/explanations.go
git commit -m "docs: add DIP127-DIP129 explanations for dippin explain"
```

---

### Task 6: Simulator — Handle interview mode

**Files:**
- Modify: `simulate/interactive.go:11-54`
- Modify: `simulate/simulate_test.go`

- [ ] **Step 1: Write the failing tests**

Add to `simulate/simulate_test.go`:

```go
func interviewWorkflow() *ir.Workflow {
	return &ir.Workflow{
		Name:  "InterviewTest",
		Start: "Start",
		Exit:  "Done",
		Nodes: []*ir.Node{
			{ID: "Start", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "Ask questions."}},
			{ID: "Ask", Kind: ir.NodeHuman, Label: "Answer questions", Config: ir.HumanConfig{
				Mode:         "interview",
				QuestionsKey: "questions",
				AnswersKey:   "answers",
			}},
			{ID: "Done", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "Done."}},
		},
		Edges: []*ir.Edge{
			{From: "Start", To: "Ask"},
			{From: "Ask", To: "Done"},
		},
	}
}

func TestRunHumanInteractive_InterviewMode(t *testing.T) {
	ResetRunCounter()
	w := interviewWorkflow()

	input := strings.NewReader("answer one\nanswer two\n\n")
	var stderr bytes.Buffer

	res, err := Run(w, Options{
		Interactive: true,
		Stdin:       input,
		Stderr:      &stderr,
	})
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}
	if res.Status != "success" {
		t.Errorf("Status = %q, want success", res.Status)
	}

	// Verify prompt shows interview header.
	if !strings.Contains(stderr.String(), "[HUMAN]") {
		t.Error("expected [HUMAN] prompt on stderr")
	}
	if !strings.Contains(stderr.String(), "interview") {
		t.Error("expected 'interview' in stderr prompt")
	}

	// Verify answers stored in context.
	var foundAnswers bool
	for _, ev := range res.Events {
		if cu, ok := ev.(event.ContextUpdate); ok && cu.Key == "answers" {
			foundAnswers = true
			if !strings.Contains(cu.Value, "answer one") {
				t.Errorf("answers missing 'answer one': %s", cu.Value)
			}
		}
	}
	if !foundAnswers {
		t.Error("expected context update for answers key")
	}
}

func TestRunHumanAutoSuccess_InterviewMode(t *testing.T) {
	w := interviewWorkflow()
	res, err := Run(w, Options{})
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}
	if res.Status != "success" {
		t.Errorf("non-interactive interview should auto-succeed, got %q", res.Status)
	}
}

func TestRunHumanInteractive_InterviewEOF(t *testing.T) {
	w := interviewWorkflow()
	input := strings.NewReader("")
	res, err := Run(w, Options{
		Interactive: true,
		Stdin:       input,
	})
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}
	if res.Status != "success" {
		t.Errorf("EOF on interview should succeed, got %q", res.Status)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `just test-pkg simulate`
Expected: FAIL — interview mode not handled distinctly yet.

- [ ] **Step 3: Update interactive.go**

Replace `simulate/interactive.go` with:

```go
package simulate

import (
	"bufio"
	"fmt"
	"strings"

	"github.com/2389-research/dippin-lang/ir"
)

// handleHumanInteraction prompts the user when the node is a human node in interactive mode.
func (s *simulator) handleHumanInteraction(node *ir.Node) error {
	hc, ok := node.Config.(ir.HumanConfig)
	if !ok || !s.opts.Interactive || s.opts.Stdin == nil {
		return nil
	}
	if hc.Mode == "interview" {
		return s.handleInterviewMode(node, hc)
	}
	response, err := s.promptInteractive(node, hc)
	if err != nil {
		return fmt.Errorf("interactive prompt at %q: %w", node.ID, err)
	}
	s.updateContext("human_response", response)
	return nil
}

func (s *simulator) handleInterviewMode(node *ir.Node, hc ir.HumanConfig) error {
	s.writeInterviewPrompt(node, hc)
	scanner := bufio.NewScanner(s.opts.Stdin)
	var lines []string
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			break
		}
		lines = append(lines, line)
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("interactive interview at %q: %w", node.ID, err)
	}
	combined := strings.Join(lines, "\n")
	answersKey := hc.AnswersKey
	if answersKey == "" {
		answersKey = "interview_answers"
	}
	s.updateContext(answersKey, combined)
	s.updateContext("human_response", combined)
	return nil
}

func (s *simulator) promptInteractive(node *ir.Node, hc ir.HumanConfig) (string, error) {
	s.writeInteractivePrompt(node, hc)
	scanner := bufio.NewScanner(s.opts.Stdin)
	if scanner.Scan() {
		return strings.TrimSpace(scanner.Text()), nil
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}
	return hc.Default, nil
}

// writeInteractivePrompt writes the human prompt to stderr if available.
func (s *simulator) writeInteractivePrompt(node *ir.Node, hc ir.HumanConfig) {
	if s.opts.Stderr == nil {
		return
	}
	label := node.Label
	if label == "" {
		label = node.ID
	}
	fmt.Fprintf(s.opts.Stderr, "\n[HUMAN] %s\n", label)
	if hc.Mode == "freeform" {
		fmt.Fprintf(s.opts.Stderr, "  Enter response: ")
	} else {
		fmt.Fprintf(s.opts.Stderr, "  Enter choice: ")
	}
}

// writeInterviewPrompt writes interview-specific prompt to stderr.
func (s *simulator) writeInterviewPrompt(node *ir.Node, hc ir.HumanConfig) {
	if s.opts.Stderr == nil {
		return
	}
	label := node.Label
	if label == "" {
		label = node.ID
	}
	fmt.Fprintf(s.opts.Stderr, "\n[HUMAN] %s (interview mode)\n", label)
	fmt.Fprintf(s.opts.Stderr, "  Enter answers line by line (blank line to finish):\n")
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `just test-pkg simulate`
Expected: PASS

- [ ] **Step 5: Check complexity**

Run: `just complexity`
Expected: All functions under limits.

- [ ] **Step 6: Commit**

```bash
git add simulate/interactive.go simulate/simulate_test.go
git commit -m "feat(simulate): handle interview mode with multi-line stdin collection"
```

---

### Task 7: Update api_design.dip example

**Files:**
- Modify: `examples/api_design.dip:49-51`

- [ ] **Step 1: Update AnswerQuestions node**

Change the `AnswerQuestions` human node from:

```dip
  human AnswerQuestions
    label: "Answer the interviewer's questions."
    mode: freeform
```

To:

```dip
  human AnswerQuestions
    label: "Answer the interviewer's questions."
    mode: interview
    questions_key: interview_questions
    answers_key: interview_answers
    reads: interview_questions
    writes: interview_answers, human_response
```

Also update the `Interviewer` agent to add `writes: interview_questions` if not already present.

- [ ] **Step 2: Validate**

Run: `just validate-examples`
Expected: All examples valid.

- [ ] **Step 3: Lint**

Run: `just lint-examples`
Expected: No new warnings (DIP127-DIP129 should be clean).

- [ ] **Step 4: Run doctor**

Run: `./dippin doctor examples/api_design.dip`
Expected: Grade A or B.

- [ ] **Step 5: Update the test file if it exists**

If `examples/api_design.test.json` has scenarios referencing `human_response`, update them to also set `interview_answers`.

- [ ] **Step 6: Commit**

```bash
git add examples/api_design.dip examples/api_design.test.json
git commit -m "feat: api_design.dip uses interview mode for Q&A collection"
```

---

### Task 8: Documentation — nodes.md

**Files:**
- Modify: `docs/nodes.md:149-199`

- [ ] **Step 1: Add interview mode section**

After the existing "Freeform Mode" subsection (line 197), add:

```markdown
### Interview Mode

In `interview` mode, the runtime extracts questions from the upstream agent's output and presents each as an individual form field. Questions with inline options (e.g., "Auth model? (API key, OAuth, JWT)") are shown as selection lists with an "Other (freeform)" escape hatch. Pure text questions become text areas.

```dippin
  human AnswerQuestions
    label: "Answer the interviewer's questions."
    mode: interview
    questions_key: interview_questions
    answers_key: interview_answers
    reads: interview_questions
    writes: interview_answers, human_response
    prompt:
      If no questions were detected, describe your
      requirements here instead.
```

#### Interview-Specific Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `questions_key` | String | `interview_questions` | Context key to read extracted questions from. Must be written by an upstream agent or tool. |
| `answers_key` | String | `interview_answers` | Context key to write structured answers to. Also writes `human_response` as markdown for backwards compatibility. |

#### How It Works

1. The upstream agent writes its output (containing questions) to the context key specified by `questions_key`.
2. The runtime parses the output for questions — numbered lists, lines ending in `?`, and imperative prompts ("Describe...", "List...").
3. Inline options in parentheses (e.g., `(API key, OAuth, JWT)`) become selection choices with an additional "Other" freeform option.
4. Questions without options become text areas.
5. Answers are stored in `answers_key` as structured JSON and in `human_response` as markdown.

#### Fallback Behavior

If the upstream output contains no parseable questions (e.g., the agent said "No further questions needed"), the runtime falls back to showing the `prompt` field as a single text area. This makes `prompt` serve as fallback instructions for interview mode.

#### Lint Checks

- **DIP127**: Fires if `mode` is not one of `choice`, `freeform`, `interview`.
- **DIP128**: Fires if `mode: interview` has a `default` value (meaningless — interview doesn't route by selection).
- **DIP129**: Fires if `mode: interview` has multiple labeled outgoing edges (interview collects answers, not choices).
```markdown

- [ ] **Step 2: Update the human-specific fields table**

Update the table at lines 160-165 to add interview-specific fields:

```markdown
| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `mode` | String | — | Interaction mode: `"choice"` (select from edge labels), `"freeform"` (open text), or `"interview"` (structured Q&A from upstream agent output). |
| `default` | String | — | Default selection if no input. Only meaningful for `"choice"` mode. |
| `questions_key` | String | `interview_questions` | Context key to read questions from. Interview mode only. |
| `answers_key` | String | `interview_answers` | Context key to write answers to. Interview mode only. |
```

- [ ] **Step 3: Commit**

```bash
git add docs/nodes.md
git commit -m "docs: document interview mode in nodes.md"
```

---

### Task 9: Documentation — integration.md

**Files:**
- Modify: `docs/integration.md:155-175`

- [ ] **Step 1: Update the type switch example**

Update the HumanConfig case in the code example at line 161-163:

```go
case ir.HumanConfig:
    fmt.Println(cfg.Mode)         // "choice", "freeform", or "interview"
    fmt.Println(cfg.Default)
    fmt.Println(cfg.QuestionsKey) // interview mode: context key for questions
    fmt.Println(cfg.AnswersKey)   // interview mode: context key for answers
```

- [ ] **Step 2: Add interview mode integrator guidance**

After the type switch section, add a subsection:

```markdown
### Interview Mode (Runtime Integration)

When `cfg.Mode == "interview"`, the runtime is expected to:

1. **Read questions** from `ctx[cfg.QuestionsKey]` (default: `"interview_questions"`).
2. **Parse questions** from the upstream agent's markdown output:
   - Numbered/bulleted lines ending in `?`
   - Imperative prompts ("Describe...", "List...", "Provide...")
   - Skip content inside fenced code blocks
3. **Extract inline options** from trailing parentheticals: `"Auth model? (API key, OAuth, JWT)"` → Select field with 3 options + "Other" freeform.
4. **Present each question** as an individual form field. Recommended mapping:
   - Questions with options → `huh.Select` + "Other" TextInput
   - Yes/no questions → `huh.Confirm` + optional elaboration TextArea
   - All other questions → `huh.TextArea`
5. **Store answers** in `ctx[cfg.AnswersKey]` (default: `"interview_answers"`) as JSON and in `ctx["human_response"]` as markdown for downstream agent consumption.
6. **Handle edge cases:**
   - 0 questions: show confirmation + optional freeform note; write empty answers
   - Malformed output: fall back to single TextArea with `cfg.Prompt` as instructions
   - Ctrl-C: write partial answers with `canceled: true` in JSON
   - Retry loop: pre-fill with previous answers from `ctx[cfg.AnswersKey]`

#### Recommended Answer JSON Schema

```json
{
  "questions": [
    {"id": "q1", "text": "Auth model?", "options": ["API key", "OAuth", "JWT"], "answer": "OAuth", "elaboration": "Google + GitHub"},
    {"id": "q2", "text": "Describe integrations", "answer": "Salesforce nightly sync..."}
  ],
  "incomplete": false,
  "canceled": false
}
```
```markdown

- [ ] **Step 3: Commit**

```bash
git add docs/integration.md
git commit -m "docs: interview mode integration guidance for the runtime"
```

---

### Task 10: Documentation — validation.md

**Files:**
- Modify: `docs/validation.md`

- [ ] **Step 1: Add DIP127-DIP129 entries**

Find the section listing lint codes and add entries following the existing pattern:

```markdown
### DIP127 — Invalid human node mode

A human node has a `mode` value that is not one of the recognized modes.

**Valid modes:** `choice`, `freeform`, `interview`

**Fix:** Change the mode to a valid value.

### DIP128 — Interview mode with meaningless default

A human node with `mode: interview` also sets a `default` value. Interview mode collects structured answers — it has no predefined choices to default to.

**Fix:** Remove the `default` field, or change the mode to `choice` if you want label-based routing.

### DIP129 — Interview mode with conflicting choice-style edges

A human node with `mode: interview` has multiple labeled outgoing edges. Interview mode does not route by label — it collects answers and follows a single unconditional edge.

**Fix:** Remove edge labels, or change the mode to `choice` if routing by selection is intended.
```

- [ ] **Step 2: Commit**

```bash
git add docs/validation.md
git commit -m "docs: DIP127-DIP129 in validation reference"
```

---

### Task 11: CHANGELOG and site

**Files:**
- Modify: `CHANGELOG.md`
- Auto-generated: `site/changelog.html`

- [ ] **Step 1: Add v0.15.0 entry**

Prepend to `CHANGELOG.md` after the header:

```markdown
## [v0.15.0] — 2026-03-30

### Added
- **Interview mode** for human nodes (`mode: interview`). Runtimes extract questions from upstream agent output and present each as an individual form field with optional suggested answers. New fields: `questions_key`, `answers_key`. See [nodes.md](docs/nodes.md) for details.
- **DIP127**: lint warning for invalid human node mode values.
- **DIP128**: lint warning when interview mode has a meaningless `default` value.
- **DIP129**: lint warning when interview mode has conflicting choice-style labeled edges.
- **Integration guide** updated with interview mode implementation guidance and recommended answer JSON schema for downstream consumers.
- `api_design.dip` example updated to use interview mode for Q&A collection.

### Changed
- `--version` / `-version` flags now work (previously failed with "flag provided but not defined"). `dippin version` subcommand still works too.
```

- [ ] **Step 2: Regenerate changelog HTML**

Run: `scripts/gen-changelog-html.sh`
Expected: `site/changelog.html` updated with new v0.15.0 card.

- [ ] **Step 3: Sync nav**

Run: `just sync-nav` (or let pre-commit handle it).

- [ ] **Step 4: Commit**

```bash
git add CHANGELOG.md site/changelog.html
git commit -m "docs: v0.15.0 changelog — interview mode, DIP127-DIP129, --version fix"
```

---

### Task 12: Final verification

- [ ] **Step 1: Full check suite**

Run: `just check`
Expected: All checks pass — build, vet, fmt, test-race, complexity, validate-examples.

- [ ] **Step 2: Lint all examples**

Run: `just lint-examples`
Expected: Zero warnings.

- [ ] **Step 3: Test the full parse-lint-format round-trip**

Run: `./dippin fmt examples/api_design.dip | diff - examples/api_design.dip`
Expected: No diff (already canonical).

- [ ] **Step 4: Simulate the updated example**

Run: `./dippin simulate examples/api_design.dip`
Expected: Successful simulation with interview node in path.

Run: `./dippin simulate --all-paths examples/api_design.dip`
Expected: Multiple paths, all successful.

- [ ] **Step 5: Doctor check**

Run: `./dippin doctor examples/api_design.dip`
Expected: Grade A.

- [ ] **Step 6: Tag release (after merge)**

```bash
git tag -a v0.15.0 -m "feat: interview mode for human nodes"
git push origin v0.15.0
```
