# Runtime Upstream Issues Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Address all 4 issues from the runtime upstream dependency report (GitHub issue #3): bracket edge syntax error, model catalog extensibility, conditional node kind, and nested retry block error.

**Architecture:** Four independent changes to the parser, IR, validator, formatter, DOT exporter, migrate, scaffold, and LSP packages. Issue 1 (bracket syntax) and Issue 4 (retry blocks) emit helpful parse errors pointing to correct syntax. Issue 2 (model catalog) adds `--extra-models` CLI flag. Issue 3 (conditional node) adds a new `NodeConditional` kind with `ConditionalConfig` throughout the pipeline.

**Tech Stack:** Go, dippin-lang IR/parser/validator/formatter/export/migrate/scaffold/lsp packages

---

## File Map

| Issue | Files Created | Files Modified |
|-------|--------------|----------------|
| 1 (bracket syntax) | `parser/parse_edges_test.go` (new tests) | `parser/parse_edges.go`, `parser/lexer.go` |
| 2 (model catalog) | — | `validator/lint_model.go`, `cmd/dippin/cli.go`, `cmd/dippin/main_test.go` |
| 3 (conditional node) | `examples/conditional_gate.dip` | `ir/ir.go`, `parser/parser.go`, `parser/parse_nodes.go`, `formatter/format.go`, `formatter/format_test.go`, `export/dot.go`, `export/dot_test.go`, `migrate/migrate.go`, `migrate/migrate_test.go`, `scaffold/scaffold.go`, `lsp/symbols.go`, `validator/lint_ebnf_test.go`, `docs/GRAMMAR.ebnf`, `parser/parse_nodes_test.go` (new tests), `parser/parser_test.go` |
| 4 (retry block) | — | `parser/parse_nodes.go`, `parser/parse_nodes_test.go` (new tests) |

---

## Task 1: Bracket Edge Syntax — Emit Parse Error (Issue 1)

The parser currently silently discards `[...]` bracket annotations on edges. Rather than supporting a second edge syntax (which creates maintenance burden), emit a clear parse error with a hint to use `when`/`label:` syntax.

**Files:**
- Modify: `parser/lexer.go` — add `TokenLBracket` token type
- Modify: `parser/parse_edges.go` — detect `[` after edge target and emit diagnostic
- Test: `parser/parse_edges_test.go` (add test cases)

- [ ] **Step 1: Write the failing test**

Add a test in `parser/parse_edges_test.go` (or `parser/parser_test.go` if edge tests live there — check first):

```go
func TestBracketEdgeSyntaxError(t *testing.T) {
	input := `workflow Test
  start: A
  exit: B
  agent A
    prompt: go
  agent B
    prompt: done
  edges
    A -> B [label: "go" condition: "ctx.x = 1"]
`
	p := parser.NewParser(input, "test.dip")
	_, err := p.Parse()
	if err == nil {
		t.Fatal("expected parse error for bracket edge syntax, got nil")
	}
	if !strings.Contains(err.Error(), "bracket") {
		t.Errorf("error should mention bracket syntax, got: %s", err.Error())
	}
	if !strings.Contains(err.Error(), "when") {
		t.Errorf("error should suggest 'when' keyword syntax, got: %s", err.Error())
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `just test-pkg parser`
Expected: FAIL — currently the bracket syntax is silently consumed without error.

- [ ] **Step 3: Add `TokenLBracket` to the lexer**

In `parser/lexer.go`, add the token type after `TokenRawBlock`:

```go
const (
	// ... existing tokens ...
	TokenRawBlock   // Raw text block (multiline prompt/command content)
	TokenLBracket   // Left bracket '[' — used for error recovery
)
```

In the lexer's token scanning logic, add recognition of `[` as `TokenLBracket`. Find where the lexer scans punctuation characters (near `TokenColon`, `TokenComma`, `TokenArrow` etc.) and add:

```go
case '[':
	return Token{Type: TokenLBracket, Value: "[", Location: loc}
```

- [ ] **Step 4: Detect bracket in `parseEdgeAttributes`**

In `parser/parse_edges.go`, modify `parseEdgeAttributes` to check for `TokenLBracket`:

```go
func (p *Parser) parseEdgeAttributes(edge *ir.Edge) {
	for p.lexer.PeekToken().Type != TokenNewline && p.lexer.PeekToken().Type != TokenEOF {
		if p.lexer.PeekToken().Type == TokenLBracket {
			p.diagnostics = append(p.diagnostics, fmt.Sprintf(
				"bracket syntax [label: ...] is not supported; use keyword syntax instead (e.g., when ctx.x = 1  label: go) at %d:%d",
				p.lexer.PeekToken().Location.Line, p.lexer.PeekToken().Location.Column))
			p.consumeUntilNewline()
			return
		}
		attr := p.lexer.NextToken()
		p.applyEdgeAttribute(edge, attr.Value)
	}
}
```

- [ ] **Step 5: Run test to verify it passes**

Run: `just test-pkg parser`
Expected: PASS

- [ ] **Step 6: Run full check**

Run: `just check`
Expected: All checks pass (build, vet, fmt, test-race, complexity, validate-examples).

- [ ] **Step 7: Commit**

```bash
git add parser/lexer.go parser/parse_edges.go parser/parser_test.go
git commit -m "fix: emit parse error for bracket edge syntax [label: ...] (upstream issue #18)"
```

---

## Task 2: Model Catalog — `--extra-models` CLI Flag (Issue 2)

Add a `--extra-models` flag to extend the model catalog at runtime. Format: `provider:model1,model2;provider2:model3`.

**Files:**
- Modify: `validator/lint_model.go` — add `RegisterExtraModels` function
- Modify: `cmd/dippin/cli.go` — add `--extra-models` flag to lint/doctor commands
- Test: `validator/lint_model.go` — unit test for `RegisterExtraModels`
- Test: `cmd/dippin/main_test.go` — integration test

- [ ] **Step 1: Write the failing test for `RegisterExtraModels`**

In an appropriate test file (check if `validator/lint_model_test.go` exists, otherwise add to `validator/lint_test.go`):

```go
func TestRegisterExtraModels(t *testing.T) {
	// Verify unknown model triggers DIP108
	w := &ir.Workflow{
		Defaults: ir.WorkflowDefaults{Provider: "custom-corp"},
		Nodes: []*ir.Node{{
			ID:   "A",
			Kind: ir.NodeAgent,
			Config: ir.AgentConfig{
				Model:    "custom-llm-v1",
				Provider: "custom-corp",
				Prompt:   "test",
			},
		}},
	}
	result := Lint(w)
	if !hasDiagCode(result, DIP108) {
		t.Fatal("expected DIP108 for unknown provider before registration")
	}

	// Register extra models
	RegisterExtraModels("custom-corp:custom-llm-v1")

	result = Lint(w)
	if hasDiagCode(result, DIP108) {
		t.Fatal("DIP108 should not fire after registering extra models")
	}

	// Clean up: remove the registered models
	delete(knownModelProviders, "custom-corp")
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `just test-pkg validator`
Expected: FAIL — `RegisterExtraModels` does not exist.

- [ ] **Step 3: Implement `RegisterExtraModels`**

In `validator/lint_model.go`:

```go
// RegisterExtraModels extends the known model catalog with user-provided entries.
// Format: "provider:model1,model2;provider2:model3"
// This allows users to suppress DIP108 warnings for private or newly-released models.
func RegisterExtraModels(spec string) {
	for _, entry := range strings.Split(spec, ";") {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}
		registerOneProvider(entry)
	}
}

// registerOneProvider parses "provider:model1,model2" and adds to the catalog.
func registerOneProvider(entry string) {
	parts := strings.SplitN(entry, ":", 2)
	if len(parts) != 2 {
		return
	}
	provider := strings.TrimSpace(parts[0])
	models := strings.Split(parts[1], ",")
	if knownModelProviders[provider] == nil {
		knownModelProviders[provider] = make(map[string]bool)
	}
	for _, m := range models {
		m = strings.TrimSpace(m)
		if m != "" {
			knownModelProviders[provider][m] = true
		}
	}
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `just test-pkg validator`
Expected: PASS

- [ ] **Step 5: Wire up CLI flag**

In `cmd/dippin/cli.go`, find the `lint` and `doctor` command definitions. Add an `--extra-models` string flag. In the command's Run function, call `validator.RegisterExtraModels(extraModels)` before running lint/doctor. Read `cli.go` first to find the exact injection points.

```go
// Add flag to lint command
lintCmd.Flags().StringVar(&extraModels, "extra-models", "",
	"extend model catalog: provider:model1,model2;provider2:model3")

// In the lint command's Run, before calling Lint:
if extraModels != "" {
	validator.RegisterExtraModels(extraModels)
}
```

Apply the same pattern to the `doctor` command.

- [ ] **Step 6: Run full check**

Run: `just check`
Expected: All checks pass.

- [ ] **Step 7: Commit**

```bash
git add validator/lint_model.go cmd/dippin/cli.go validator/lint_test.go
git commit -m "feat: --extra-models flag to extend DIP108 model catalog (upstream issue #36)"
```

---

## Task 3: Conditional Node Kind (Issue 3)

Add `NodeConditional` to the IR with a minimal `ConditionalConfig` (label-only, no prompt/model). This node kind represents pure branching — evaluate outgoing edge conditions without an LLM call. Maps to `diamond` shape in DOT.

This is the largest task. It touches IR, parser, formatter, DOT export, migrate, scaffold, LSP, validator tests, and the EBNF grammar doc.

### Task 3a: IR — Add `NodeConditional` and `ConditionalConfig`

**Files:**
- Modify: `ir/ir.go:63-73` — add `NodeConditional` constant and `ConditionalConfig` type

- [ ] **Step 1: Write the failing test**

In `ir/ir_test.go`, add:

```go
func TestConditionalConfigSatisfiesInterface(t *testing.T) {
	var cfg ir.NodeConfig = ir.ConditionalConfig{}
	if cfg == nil {
		t.Fatal("ConditionalConfig should implement NodeConfig")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `just test-pkg ir`
Expected: FAIL — `ir.ConditionalConfig` undefined.

- [ ] **Step 3: Add to IR**

In `ir/ir.go`, add the constant in the `NodeKind` const block:

```go
const (
	NodeAgent       NodeKind = "agent"
	NodeHuman       NodeKind = "human"
	NodeTool        NodeKind = "tool"
	NodeParallel    NodeKind = "parallel"
	NodeFanIn       NodeKind = "fan_in"
	NodeSubgraph    NodeKind = "subgraph"
	NodeConditional NodeKind = "conditional"
)
```

After `SubgraphConfig`, add:

```go
// ConditionalConfig holds configuration for pure conditional branching nodes.
// These nodes evaluate outgoing edge conditions without making an LLM call.
type ConditionalConfig struct{}

func (ConditionalConfig) nodeConfig() {}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `just test-pkg ir`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add ir/ir.go ir/ir_test.go
git commit -m "feat(ir): add NodeConditional kind and ConditionalConfig"
```

### Task 3b: Parser — Parse `conditional` Keyword

**Files:**
- Modify: `parser/parser.go:79` — add `"conditional": true` to `workflowNodeKinds`
- Modify: `parser/parse_nodes.go:11-24` — add case for `ir.NodeConditional` in `defaultNodeConfig`

- [ ] **Step 1: Write the failing test**

In `parser/parser_test.go`, add:

```go
func TestParseConditionalNode(t *testing.T) {
	input := `workflow Test
  start: Check
  exit: Done
  conditional Check
    label: "Route by outcome"
  agent Pass
    prompt: success path
  agent Done
    prompt: wrap up
  edges
    Check -> Pass  when ctx.outcome = success
    Check -> Done
    Pass -> Done
`
	p := parser.NewParser(input, "test.dip")
	w, err := p.Parse()
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	node := w.Node("Check")
	if node == nil {
		t.Fatal("Check node not found")
	}
	if node.Kind != ir.NodeConditional {
		t.Errorf("expected NodeConditional, got %s", node.Kind)
	}
	if node.Label != "Route by outcome" {
		t.Errorf("expected label 'Route by outcome', got %q", node.Label)
	}
	if _, ok := node.Config.(ir.ConditionalConfig); !ok {
		t.Errorf("expected ConditionalConfig, got %T", node.Config)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `just test-pkg parser`
Expected: FAIL — `conditional` is not in `workflowNodeKinds`.

- [ ] **Step 3: Add parser support**

In `parser/parser.go:79`, add `"conditional"`:

```go
var workflowNodeKinds = map[string]bool{
	"agent": true, "human": true, "tool": true, "subgraph": true, "conditional": true,
}
```

In `parser/parse_nodes.go`, add the case in `defaultNodeConfig`:

```go
func defaultNodeConfig(kind ir.NodeKind) ir.NodeConfig {
	switch kind {
	case ir.NodeAgent:
		return ir.AgentConfig{Params: make(map[string]string)}
	case ir.NodeHuman:
		return ir.HumanConfig{}
	case ir.NodeTool:
		return ir.ToolConfig{}
	case ir.NodeSubgraph:
		return ir.SubgraphConfig{Params: make(map[string]string)}
	case ir.NodeConditional:
		return ir.ConditionalConfig{}
	default:
		return ir.AgentConfig{}
	}
}
```

Also in `applyConfigField`, add a case for `ConditionalConfig` (it accepts no config-specific fields, just common fields like label/class):

```go
case ir.ConditionalConfig:
	// Conditional nodes only accept common fields (label, class, reads, writes).
	// No config-specific fields to apply.
```

- [ ] **Step 4: Run test to verify it passes**

Run: `just test-pkg parser`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add parser/parser.go parser/parse_nodes.go parser/parser_test.go
git commit -m "feat(parser): support 'conditional' node kind"
```

### Task 3c: Formatter — Emit `conditional` Nodes

**Files:**
- Modify: `formatter/format.go:239-250` — add `ConditionalConfig` case in `writeNodeConfigFields`
- Test: `formatter/format_test.go`

- [ ] **Step 1: Write the failing test**

In `formatter/format_test.go`:

```go
func TestFormatConditionalNode(t *testing.T) {
	w := &ir.Workflow{
		Name:  "cond_test",
		Start: "Check",
		Exit:  "Done",
		Nodes: []*ir.Node{
			{ID: "Check", Kind: ir.NodeConditional, Label: "Route by outcome", Config: ir.ConditionalConfig{}},
			{ID: "Done", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "wrap up"}},
		},
		Edges: []*ir.Edge{
			{From: "Check", To: "Done"},
		},
	}
	got := Format(w)
	if !strings.Contains(got, "conditional Check") {
		t.Errorf("expected 'conditional Check' in output, got:\n%s", got)
	}
	if !strings.Contains(got, `label: "Route by outcome"`) {
		t.Errorf("expected label field in output, got:\n%s", got)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `just test-pkg formatter`
Expected: FAIL — `ConditionalConfig` not matched in `writeNodeConfigFields`.

- [ ] **Step 3: Add formatter support**

In `formatter/format.go`, add `ConditionalConfig` case in `writeNodeConfigFields`:

```go
func writeNodeConfigFields(wr *writer, n *ir.Node) {
	switch cfg := n.Config.(type) {
	case ir.AgentConfig:
		writeAgentFields(wr, n, cfg)
	case ir.HumanConfig:
		writeHumanFields(wr, n, cfg)
	case ir.ToolConfig:
		writeToolFields(wr, n, cfg)
	case ir.SubgraphConfig:
		writeSubgraphFields(wr, n, cfg)
	case ir.ConditionalConfig:
		writeConditionalFields(wr, n)
	}
}
```

Add the writer function:

```go
func writeConditionalFields(wr *writer, n *ir.Node) {
	writeCommonNodeFields(wr, n)
	writeIOFields(wr, n)
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `just test-pkg formatter`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add formatter/format.go formatter/format_test.go
git commit -m "feat(formatter): emit conditional node kind"
```

### Task 3d: DOT Export — Map `NodeConditional` to `diamond`

**Files:**
- Modify: `export/dot.go:86-93` — add `ir.NodeConditional: "diamond"` to `nodeShapes`
- Modify: `export/dot.go:166-197` — add `ConditionalConfig` case in `applyConfigAttrs` (no-op)
- Test: `export/dot_test.go`

- [ ] **Step 1: Write the failing test**

In `export/dot_test.go`:

```go
func TestExportConditionalNodeShape(t *testing.T) {
	w := &ir.Workflow{
		Name:  "cond",
		Start: "Gate",
		Exit:  "End",
		Nodes: []*ir.Node{
			{ID: "Gate", Kind: ir.NodeConditional, Label: "Route", Config: ir.ConditionalConfig{}},
			{ID: "End", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "done"}},
		},
		Edges: []*ir.Edge{{From: "Gate", To: "End"}},
	}
	dot := ExportDOT(w, ExportOptions{})
	// Start node overrides to Mdiamond, so test with a non-start conditional
	w2 := &ir.Workflow{
		Name:  "cond2",
		Start: "Begin",
		Exit:  "End",
		Nodes: []*ir.Node{
			{ID: "Begin", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "go"}},
			{ID: "Gate", Kind: ir.NodeConditional, Label: "Route", Config: ir.ConditionalConfig{}},
			{ID: "End", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "done"}},
		},
		Edges: []*ir.Edge{
			{From: "Begin", To: "Gate"},
			{From: "Gate", To: "End"},
		},
	}
	dot = ExportDOT(w2, ExportOptions{})
	if !strings.Contains(dot, `shape="diamond"`) && !strings.Contains(dot, `shape=diamond`) {
		t.Errorf("expected diamond shape for conditional node, got:\n%s", dot)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `just test-pkg export`
Expected: FAIL — `nodeShapes` has no entry for `NodeConditional`, falls back to `"box"`.

- [ ] **Step 3: Add DOT export support**

In `export/dot.go`, add to `nodeShapes`:

```go
var nodeShapes = map[ir.NodeKind]string{
	ir.NodeAgent:       "box",
	ir.NodeHuman:       "hexagon",
	ir.NodeTool:        "parallelogram",
	ir.NodeParallel:    "component",
	ir.NodeFanIn:       "tripleoctagon",
	ir.NodeSubgraph:    "tab",
	ir.NodeConditional: "diamond",
}
```

In `applyStructuralConfigAttrs`, add the `ConditionalConfig` case (no extra attrs needed):

```go
func applyStructuralConfigAttrs(attrs map[string]string, cfg interface{}) {
	switch c := cfg.(type) {
	case ir.SubgraphConfig:
		applySubgraphAttrs(attrs, c)
	case ir.ParallelConfig:
		applyParallelAttrs(attrs, c)
	case ir.FanInConfig:
		applyFanInAttrs(attrs, c)
	case ir.ConditionalConfig:
		// No additional attributes for conditional nodes.
	}
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `just test-pkg export`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add export/dot.go export/dot_test.go
git commit -m "feat(export): map NodeConditional to diamond shape in DOT"
```

### Task 3e: Migrate — Map `diamond` to `NodeConditional`

**Files:**
- Modify: `migrate/migrate.go:43-50` — add `"diamond": ir.NodeConditional` to `shapeToKind`
- Modify: `migrate/migrate.go:198-218` — update `resolveKind` to use `NodeConditional` for diamond
- Modify: `migrate/migrate.go:163-195` — add `ConditionalConfig` case in `applyNodeConfig`
- Test: `migrate/migrate_test.go`

- [ ] **Step 1: Write the failing test**

In `migrate/migrate_test.go`:

```go
func TestMigrateDiamondToConditional(t *testing.T) {
	dot := `digraph test {
  Gate [shape=diamond, label="Route Check"];
  Pass [shape=box, label="Pass"];
  Gate -> Pass [condition="outcome=success"];
}`
	w, err := Migrate(dot)
	if err != nil {
		t.Fatalf("migrate error: %v", err)
	}
	node := w.Node("Gate")
	if node == nil {
		t.Fatal("Gate node not found")
	}
	if node.Kind != ir.NodeConditional {
		t.Errorf("expected NodeConditional for diamond shape, got %s", node.Kind)
	}
	if _, ok := node.Config.(ir.ConditionalConfig); !ok {
		t.Errorf("expected ConditionalConfig, got %T", node.Config)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `just test-pkg migrate`
Expected: FAIL — diamond currently maps to `NodeAgent`.

- [ ] **Step 3: Update migrate**

In `migrate/migrate.go`, add to `shapeToKind`:

```go
var shapeToKind = map[string]ir.NodeKind{
	"box":           ir.NodeAgent,
	"hexagon":       ir.NodeHuman,
	"parallelogram": ir.NodeTool,
	"component":     ir.NodeParallel,
	"tripleoctagon": ir.NodeFanIn,
	"tab":           ir.NodeSubgraph,
	"diamond":       ir.NodeConditional,
}
```

Remove the special `diamond` handling in `resolveKind` — now that `diamond` is in `shapeToKind`, the lookup handles it:

```go
func resolveKind(shape string, attrs map[string]string) ir.NodeKind {
	if shape == "Mdiamond" || shape == "Msquare" {
		return ir.NodeAgent
	}
	if kind, ok := shapeToKind[shape]; ok {
		return kind
	}
	return ir.NodeAgent
}
```

Remove `resolveDiamondKind` function entirely.

In `applyNodeConfig`, add the `NodeConditional` case:

```go
case ir.NodeConditional:
	node.Config = ir.ConditionalConfig{}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `just test-pkg migrate`
Expected: PASS

**Note:** Check if existing migrate tests assume diamond → agent. If so, update those tests to expect `NodeConditional`. The `resolveDiamondKind` function had special logic for `tool_command` attribute — that case is rare (diamond + tool_command). Since `diamond` now always maps to `NodeConditional`, verify no existing tests break. If a test uses diamond + tool_command, that edge case can be dropped since the canonical DOT shape for tools is `parallelogram`, not `diamond`.

- [ ] **Step 5: Commit**

```bash
git add migrate/migrate.go migrate/migrate_test.go
git commit -m "feat(migrate): map diamond shape to NodeConditional"
```

### Task 3f: LSP, Scaffold, EBNF, and Validator Tests

**Files:**
- Modify: `lsp/symbols.go:67-80` — add `ir.NodeConditional` case
- Modify: `scaffold/scaffold.go:95-130` — update `buildConditional` to use `NodeConditional`
- Modify: `validator/lint_ebnf_test.go:73` — add `"conditional"` to kind list
- Modify: `docs/GRAMMAR.ebnf` — add `"conditional"` to node kind rule

- [ ] **Step 1: Update LSP symbols**

In `lsp/symbols.go`, add `ir.NodeConditional` case in `nodeSymbolKind`:

```go
func nodeSymbolKind(kind ir.NodeKind) protocol.SymbolKind {
	switch kind {
	case ir.NodeAgent:
		return protocol.SymbolKindFunction
	case ir.NodeHuman:
		return protocol.SymbolKindInterface
	case ir.NodeTool:
		return protocol.SymbolKindMethod
	case ir.NodeParallel, ir.NodeFanIn:
		return protocol.SymbolKindStruct
	case ir.NodeConditional:
		return protocol.SymbolKindEnum
	default:
		return protocol.SymbolKindVariable
	}
}
```

- [ ] **Step 2: Update scaffold template**

In `scaffold/scaffold.go`, update `buildConditional` to use `ir.NodeConditional` and `ir.ConditionalConfig{}` for the `Check` node:

```go
func buildConditional(name string) *ir.Workflow {
	return &ir.Workflow{
		Name:  name,
		Goal:  "Route based on status check",
		Start: "Check",
		Exit:  "Done",
		Nodes: []*ir.Node{
			{ID: "Check", Kind: ir.NodeConditional, Label: "Evaluate outcome", Config: ir.ConditionalConfig{}},
			{ID: "Pass", Kind: ir.NodeAgent, Config: ir.AgentConfig{
				Prompt: "Handle the success path.",
			}},
			{ID: "Fail", Kind: ir.NodeAgent, Config: ir.AgentConfig{
				Prompt: "Handle the failure path.",
			}},
			{ID: "Done", Kind: ir.NodeAgent, Config: ir.AgentConfig{
				Prompt: "Wrap up.",
			}},
		},
		Edges: []*ir.Edge{
			{From: "Check", To: "Pass", Condition: &ir.Condition{
				Raw:    "ctx.outcome = success",
				Parsed: ir.CondCompare{Variable: "ctx.outcome", Op: "=", Value: "success"},
			}},
			{From: "Check", To: "Fail", Condition: &ir.Condition{
				Raw:    "ctx.outcome = fail",
				Parsed: ir.CondCompare{Variable: "ctx.outcome", Op: "=", Value: "fail"},
			}},
			{From: "Check", To: "Done"},
			{From: "Pass", To: "Done"},
			{From: "Fail", To: "Done"},
		},
	}
}
```

- [ ] **Step 3: Update EBNF test**

In `validator/lint_ebnf_test.go:73`, add `"conditional"`:

```go
kinds := []string{"agent", "human", "tool", "parallel", "fan_in", "subgraph", "conditional"}
```

- [ ] **Step 4: Update GRAMMAR.ebnf**

Read `docs/GRAMMAR.ebnf` and add `"conditional"` to the node kind production rule. The exact syntax depends on the EBNF format used — add it alongside the other node kinds.

- [ ] **Step 5: Create example file**

Create `examples/conditional_gate.dip`:

```
workflow ConditionalGate
  goal: "Demonstrate conditional routing without LLM calls"
  start: Analyze
  exit: Done

  agent Analyze
    label: "Analyze Input"
    auto_status: true
    prompt:
      Evaluate the input and set STATUS: success or STATUS: fail
      based on whether the criteria are met.

  conditional Route
    label: "Route by Outcome"

  agent HandleSuccess
    prompt: Process the successful result.

  agent HandleFailure
    prompt: Handle the failure case.

  agent Done
    prompt: Summarize the final result.

  edges
    Analyze -> Route
    Route -> HandleSuccess  when ctx.outcome = success
    Route -> HandleFailure  when ctx.outcome = fail
    Route -> Done
    HandleSuccess -> Done
    HandleFailure -> Done
```

- [ ] **Step 6: Run full check**

Run: `just check`
Expected: All checks pass. The EBNF test now expects 7 node kinds. The new example validates cleanly.

- [ ] **Step 7: Commit**

```bash
git add lsp/symbols.go scaffold/scaffold.go validator/lint_ebnf_test.go docs/GRAMMAR.ebnf examples/conditional_gate.dip
git commit -m "feat: wire conditional node through LSP, scaffold, EBNF, and examples"
```

---

## Task 4: Nested Retry Block — Emit Parse Error (Issue 4)

The parser rejects nested `retry` blocks with a confusing indent error. Detect the `retry` keyword at the start of a node body line and emit a clear error pointing to the flat syntax.

**Files:**
- Modify: `parser/parse_nodes.go:44-61` — detect `retry` keyword in `parseNodeBody`
- Test: `parser/parser_test.go`

- [ ] **Step 1: Write the failing test**

In `parser/parser_test.go`:

```go
func TestNestedRetryBlockError(t *testing.T) {
	input := `workflow Test
  start: A
  exit: B
  tool A
    label: "Verify"
    command: "verify.sh"
    timeout: 30s
    retry
      policy: aggressive
      max_retries: 5
      retry_target: process
  agent B
    prompt: done
  edges
    A -> B
`
	p := parser.NewParser(input, "test.dip")
	_, err := p.Parse()
	if err == nil {
		t.Fatal("expected parse error for nested retry block")
	}
	if !strings.Contains(err.Error(), "retry") {
		t.Errorf("error should mention retry, got: %s", err.Error())
	}
	if !strings.Contains(err.Error(), "retry_policy") || !strings.Contains(err.Error(), "max_retries") {
		t.Errorf("error should suggest flat attribute names, got: %s", err.Error())
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `just test-pkg parser`
Expected: FAIL — current error is `expected 9, got 2` (confusing indent error), not a helpful message about retry blocks.

- [ ] **Step 3: Detect nested `retry` in `parseNodeBody`**

In `parser/parse_nodes.go`, modify `parseNodeBody` to detect a `retry` identifier that is NOT followed by `_` (i.e., not `retry_policy`, `retry_target`):

```go
func (p *Parser) parseNodeBody(node *ir.Node) {
	for p.lexer.PeekToken().Type != TokenOutdent && p.lexer.PeekToken().Type != TokenEOF {
		t := p.lexer.PeekToken()
		if t.Type == TokenNewline {
			p.lexer.NextToken()
			continue
		}
		if t.Type == TokenIdentifier {
			if t.Value == "retry" && !p.isRetryFieldPrefix() {
				p.diagnostics = append(p.diagnostics, fmt.Sprintf(
					"nested retry blocks are not supported; use flat attributes instead (retry_policy, max_retries, retry_target, fallback_target, base_delay) at %d:%d",
					t.Location.Line, t.Location.Column))
				p.skipNestedBlock()
				continue
			}
			key := t.Value
			p.lexer.NextToken()
			p.expect(TokenColon)
			val := p.readFieldValue(t.Location.Line)
			p.applyNodeField(node, key, val, t.Location)
		} else {
			p.lexer.NextToken()
		}
	}
}
```

Add helper methods:

```go
// isRetryFieldPrefix checks if the current "retry" token is actually "retry_policy" etc.
// by peeking at whether the next token on the same line starts with a colon (field assignment).
func (p *Parser) isRetryFieldPrefix() bool {
	// "retry" as a field would need a colon: "retry:" — but "retry" is not a valid field name.
	// "retry_policy:", "retry_target:" are valid. Since we only match the identifier "retry"
	// (the lexer breaks on '_'), check if next token is TokenColon.
	// Actually, the lexer treats "retry_policy" as a single identifier since _ is valid in identifiers.
	// So if we reach here with value == "retry", it's NOT "retry_policy" — it's a bare "retry".
	return false
}

// skipNestedBlock consumes tokens through the nested block's indent/outdent.
func (p *Parser) skipNestedBlock() {
	p.lexer.NextToken() // consume "retry"
	// Skip to end of line
	for p.lexer.PeekToken().Type != TokenNewline && p.lexer.PeekToken().Type != TokenEOF {
		p.lexer.NextToken()
	}
	if p.lexer.PeekToken().Type == TokenNewline {
		p.lexer.NextToken()
	}
	// If there's an indented block, skip it
	if p.lexer.PeekToken().Type == TokenIndent {
		p.lexer.NextToken() // consume indent
		for p.lexer.PeekToken().Type != TokenOutdent && p.lexer.PeekToken().Type != TokenEOF {
			p.lexer.NextToken()
		}
		if p.lexer.PeekToken().Type == TokenOutdent {
			p.lexer.NextToken() // consume outdent
		}
	}
}
```

**Important:** Verify how the lexer tokenizes `retry_policy` — if underscore is part of identifiers (it should be based on lexer rules), then "retry" alone will never match "retry_policy". Check `parser/lexer.go` to confirm underscores are valid in identifiers before finalizing.

- [ ] **Step 4: Run test to verify it passes**

Run: `just test-pkg parser`
Expected: PASS

- [ ] **Step 5: Run full check**

Run: `just check`
Expected: All checks pass.

- [ ] **Step 6: Commit**

```bash
git add parser/parse_nodes.go parser/parser_test.go
git commit -m "fix: clear error for nested retry blocks, suggest flat syntax (upstream issue #3)"
```

---

## Task 5: Final Validation and PR

- [ ] **Step 1: Run full check suite**

Run: `just check`
Expected: All checks pass.

- [ ] **Step 2: Run lint on all examples including the new one**

Run: `just lint-examples`
Expected: No warnings (or only pre-existing ones).

- [ ] **Step 3: Verify the example round-trips**

```bash
just build
./dippin fmt examples/conditional_gate.dip
./dippin validate examples/conditional_gate.dip
./dippin lint examples/conditional_gate.dip
./dippin doctor examples/conditional_gate.dip
```

Expected: All commands succeed cleanly.

- [ ] **Step 4: Create PR**

```bash
gh pr create --title "fix: address 4 runtime upstream issues (#3)" --body "$(cat <<'EOF'
## Summary
Addresses all 4 issues from the runtime upstream dependency report (issue #3):

- **Issue 1 (P1):** Bracket edge syntax `[label: ...]` now emits a clear parse error with hint to use `when`/`label:` keyword syntax, instead of silently discarding annotations
- **Issue 2 (P2):** New `--extra-models` CLI flag on `lint` and `doctor` commands lets users extend the DIP108 model catalog with private/new models
- **Issue 3 (P2):** New `conditional` node kind for pure branching without LLM calls — maps to `diamond` shape in DOT, wired through IR/parser/formatter/export/migrate/scaffold/LSP
- **Issue 4 (P3):** Nested `retry` block syntax now emits a clear error suggesting flat attributes (`retry_policy`, `max_retries`, etc.)

## Test plan
- [ ] `just check` passes (build, vet, fmt, test-race, complexity, validate-examples)
- [ ] `just lint-examples` passes with new `conditional_gate.dip` example
- [ ] New parser tests verify bracket syntax and nested retry errors
- [ ] New conditional node tests cover IR, parser, formatter, DOT export, migrate
- [ ] `--extra-models` flag tested in validator unit tests

Closes #3
EOF
)"
```
