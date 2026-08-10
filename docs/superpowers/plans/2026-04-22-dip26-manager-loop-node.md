# manager_loop node kind (issue #26) Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a new `manager_loop` node kind to dippin-lang so `.dip` files can natively express the runtime's `stack.manager_loop` supervisor — a node that runs a child subgraph, polls on an interval, and can steer the running child by injecting context.

**Architecture:** Add `NodeManagerLoop` + `ManagerLoopConfig` to the IR, mirroring the existing `SubgraphConfig` shape with duration/int/condition/map fields. Parser dispatch follows the existing per-kind handler pattern in `parser/parse_nodes.go`. `stop_condition` and `steer_condition` reuse `ir.Condition{Raw, Parsed}` and are parsed on demand via an extended `simulate.EnsureConditionsParsed`. `steer_context` uses `map[string]string` and supports both inline CSV (`key=val,key=val`) and a multi-line block form mirroring `subgraph.params`. DOT exports as `shape=house` (matching the runtime's existing DOT parser) with a lossless flat-attr round-trip. Linter allocates DIP135/136/137 and extends DIP120's namespace list with `stack.child.*`. Tree-sitter grammar, VS Code TextMate, LSP symbol map, DOT migrator, scaffold, docs, site, and changelog all updated in the same PR. Tree-sitter generated files are committed (consumers compile `src/parser.c` directly); a CI drift check prevents them going stale.

**Tech Stack:** Go 1.25, `just` task runner, `simulate.EnsureConditionsParsed` for condition AST, tree-sitter CLI (`tree-sitter-cli ^0.24.0`), Hugo for site.

**Scope note:** This plan also folds in a trivial pre-existing bug fix — the VS Code TextMate `node-declaration` regex at line 30 omits `parallel` and `fan_in`. That fix is in Task 16 and is called out in the CHANGELOG under "Fixed". Per recent repo precedent (commit `bfa4219`), grammar-sync fixes batch into one commit.

---

## Prerequisite reading for the implementing engineer

Skim these before starting — they establish the patterns this plan mirrors:

- `ir/ir.go` — how `NodeKind` constants, `NodeConfig` interface, and per-kind structs are defined.
- `parser/parse_nodes.go:11-17` (default configs), `:152-163` (secondary field dispatch), `:415-451` (subgraph + params block parsing).
- `validator/lint_codes.go` — DIP code catalog and description init pattern.
- `validator/explanations.go` — **every DIP code must have an entry here; CI enforces it via `explanations_test.go`.**
- `validator/lint_subgraph.go` — the closest analog (validates `SubgraphConfig.Ref` against filesystem via DIP126).
- `validator/lint.go:17-60` — where linters get registered; note `simulate.EnsureConditionsParsed(w)` at line 20.
- `export/dot.go:113-121` (`nodeShapes`) and `:217-226` (`applySemanticStructuralAttrs`).
- `migrate/migrate.go:44-51` (`shapeToKind`) and `:256-270` (`buildOtherConfig`).
- `formatter/format.go:295-305` (`writeSecondaryConfigFields`).
- `scaffold/scaffold.go` — full file; the `buildConditional` function is the closest model.
- `simulate/condition.go:265-288` — `EnsureConditionsParsed` walks edges; we need to extend it.

**CLAUDE.md non-negotiables:**
- Cyclomatic complexity ≤ 5 per function, cognitive complexity ≤ 7. Extract helpers instead of adding `//nolint`.
- Never run raw `go test` etc. — always use `just <recipe>`.
- After any changes: run `just check` (build, vet, fmt, test-race, releasecheck, complexity, validate-examples) and it must pass.
- Pricing/model catalogs are not touched in this plan.

---

## Design decisions (locked in; don't re-litigate)

| Decision | Value | Why |
|---|---|---|
| Keyword | `manager_loop` | Matches the runtime's `stack.manager_loop`; `fan_in` precedent for underscored multi-word kinds; `manager` alone collides with "manager agent" in AI-pipeline vocabulary. |
| IR constant | `NodeManagerLoop NodeKind = "manager_loop"` | Underscores match string value. |
| Config struct | `ManagerLoopConfig` | Matches naming of `ParallelConfig`, `SubgraphConfig`. |
| Required field | `subgraph_ref` (not `ref`) | A manager_loop is not itself a subgraph — `ref` would mislead. |
| Condition fields | `StopCondition *ir.Condition`, `SteerCondition *ir.Condition` | Reuse existing Condition struct; plug into `simulate.EnsureConditionsParsed` so `Parsed` AST is guaranteed before lint/simulate (per the DIP101 bug pattern CLAUDE.md warns about). |
| `steer_context` | `map[string]string`, supports inline CSV AND block form | Mirrors `AgentConfig.Params` / `SubgraphConfig.Params`. Parse-time structuring keeps column-precise diagnostics. |
| DOT shape | `house` | Matches the runtime's existing DOT parser mapping for `stack.manager_loop`. |
| DIP codes | DIP135, DIP136, DIP137 (3 codes, not 5) | `poll_interval`/`max_cycles`/`steer_context` shape errors collapse into DIP136 per existing DIP116/DIP130 precedent. DIP137 covers "unbounded manager" as a DIP104 analog. |
| Tree-sitter generated files | Commit them | Zed pins a commit SHA and compiles `src/parser.c` directly; canonical grammars (tree-sitter-go/rust/python) all commit. Add `just tree-sitter-generate` + CI drift check to prevent staleness. |
| Scaffold template | Add `dippin scaffold manager_loop` | Precedent: `conditional` is a one-kind-showcase template. |
| VS Code regex | Fix `parallel`/`fan_in` omission in same PR | One-line regex edit, trivial revert, repo precedent (commit `bfa4219`). |
| Runtime adapter | Separate runtime issue filed first, cross-linked in PR body; do NOT block merge | Downstream consumers pin versions — they can pin to `<v0.22.0` until their adapter ships. |
| Version | v0.22.0 | Pre-1.0 additive; no breaking IR change. |

---

## File Structure

**New files:**
- `validator/lint_manager_loop.go` — DIP135/136/137 linter
- `validator/lint_manager_loop_test.go` — unit tests per DIP code
- `scaffold/manager_loop.go` *(or inline in scaffold.go if small)* — template builder
- `examples/manager_loop_demo.dip` — integration example; exercises all fields
- `docs/superpowers/plans/2026-04-22-dip26-manager-loop-node.md` — this file

**Modified files (by area):**
- **IR:** `ir/ir.go`
- **Parser:** `parser/parser.go`, `parser/parse_nodes.go`
- **Condition parsing:** `simulate/condition.go`
- **Formatter:** `formatter/format.go`
- **DOT export:** `export/dot.go`
- **DOT migrate:** `migrate/migrate.go`
- **Validator:** `validator/lint_codes.go`, `validator/explanations.go`, `validator/lint.go` (registration), `validator/lint_conditions.go` (extend knownNamespaces for `stack.child.*`)
- **LSP:** `lsp/symbols.go`
- **Scaffold:** `scaffold/scaffold.go`
- **Tree-sitter:** `editors/tree-sitter-dippin/grammar.js`, `queries/highlights.scm`, `test/corpus/basic.txt`, plus generated `src/parser.c`, `src/grammar.json`, `src/node-types.json`, `src/tree_sitter/*.h`
- **VS Code:** `editors/vscode/syntaxes/dippin.tmLanguage.json`
- **Docs:** `docs/nodes.md`, `docs/GRAMMAR.ebnf`, `README.md`
- **Site:** `site/content/language.md`, `site/static/skill.md` (and `site/content/changelog.md` regenerates from CHANGELOG.md via `just changelog-md`)
- **Build:** `Justfile`, `.github/workflows/ci.yml`, `.gitignore`
- **Changelog:** `CHANGELOG.md`

---

## Task 1: Add NodeManagerLoop and ManagerLoopConfig to the IR

**Files:**
- Modify: `ir/ir.go` (enum at lines 70-78; config types section ending at line 167)

- [ ] **Step 1.1: Write the failing test**

Add to `ir/ir_test.go` (if the file doesn't have a node-kind enumeration test, add a new `TestNodeKinds` — otherwise extend it):

```go
func TestManagerLoopConfig_ImplementsNodeConfig(t *testing.T) {
    var _ NodeConfig = ManagerLoopConfig{}

    cfg := ManagerLoopConfig{
        SubgraphRef:    "quality_loop",
        PollInterval:   30 * time.Second,
        MaxCycles:      12,
        StopCondition:  &Condition{Raw: "stack.child.cycles = 10"},
        SteerCondition: &Condition{Raw: "stack.child.cycles = 5"},
        SteerContext:   map[string]string{"hint": "speed_up", "priority": "high"},
    }

    if cfg.SubgraphRef != "quality_loop" {
        t.Errorf("SubgraphRef = %q, want %q", cfg.SubgraphRef, "quality_loop")
    }
    if cfg.PollInterval != 30*time.Second {
        t.Errorf("PollInterval = %v, want 30s", cfg.PollInterval)
    }
    if got := cfg.SteerContext["hint"]; got != "speed_up" {
        t.Errorf("SteerContext[hint] = %q, want %q", got, "speed_up")
    }
}

func TestNodeManagerLoop_IsNodeKind(t *testing.T) {
    if NodeManagerLoop != "manager_loop" {
        t.Errorf("NodeManagerLoop = %q, want %q", NodeManagerLoop, "manager_loop")
    }
}
```

- [ ] **Step 1.2: Run the test — expect it to fail**

```shell
just test-pkg ir
```

Expected: FAIL with `undefined: NodeManagerLoop` and `undefined: ManagerLoopConfig`.

- [ ] **Step 1.3: Add the constant**

In `ir/ir.go`, extend the const block at line 70-78:

```go
const (
    NodeAgent       NodeKind = "agent"
    NodeHuman       NodeKind = "human"
    NodeTool        NodeKind = "tool"
    NodeParallel    NodeKind = "parallel"
    NodeFanIn       NodeKind = "fan_in"
    NodeSubgraph    NodeKind = "subgraph"
    NodeConditional NodeKind = "conditional"
    NodeManagerLoop NodeKind = "manager_loop"
)
```

- [ ] **Step 1.4: Add the config struct**

Append after the `ConditionalConfig` definition (around line 167):

```go
// ManagerLoopConfig holds configuration for manager_loop supervisor nodes.
// A manager_loop runs a child subgraph, polls at PollInterval, and may
// steer the child by injecting SteerContext when SteerCondition evaluates
// true against ctx.stack.child.* variables exposed by the runtime.
type ManagerLoopConfig struct {
    SubgraphRef    string            // Child subgraph to supervise (required)
    PollInterval   time.Duration     // Polling cadence; 0 = event-driven
    MaxCycles      int               // Hard cap on child cycles; 0 = unbounded
    StopCondition  *Condition        // Terminate supervision when true
    SteerCondition *Condition        // Inject SteerContext when true
    SteerContext   map[string]string // Key-value hints injected into child
}

func (ManagerLoopConfig) nodeConfig() {}
```

- [ ] **Step 1.5: Run the tests to confirm they pass**

```shell
just test-pkg ir
```

Expected: PASS.

- [ ] **Step 1.6: Commit**

```bash
git add ir/ir.go ir/ir_test.go
git commit -m "feat(ir): add NodeManagerLoop kind and ManagerLoopConfig

Supports the runtime's stack.manager_loop supervisor nodes — a child subgraph
ref, poll interval, max cycles, stop/steer conditions (reusing ir.Condition
for AST parsing consistency), and a key-value steer_context map."
```

---

## Task 2: Register default config and add parser dispatch for `manager_loop`

**Files:**
- Modify: `parser/parser.go:79-80` (workflowNodeKinds)
- Modify: `parser/parse_nodes.go:11-17` (defaultNodeConfigs), `:151-163` (applySecondaryConfigField)

- [ ] **Step 2.1: Write the failing test**

Add to `parser/parser_test.go`:

```go
func TestParseManagerLoopNode(t *testing.T) {
    src := `workflow W
  start: M
  exit: M

  manager_loop M
    label: "Quality Gate Supervisor"
    subgraph_ref: quality_loop
    poll_interval: 30s
    max_cycles: 12

  edges
    M -> M
`
    w, diags, err := Parse(src)
    if err != nil {
        t.Fatalf("Parse error: %v", err)
    }
    if len(diags) != 0 {
        t.Fatalf("unexpected diagnostics: %v", diags)
    }
    if len(w.Nodes) != 1 {
        t.Fatalf("got %d nodes, want 1", len(w.Nodes))
    }
    n := w.Nodes[0]
    if n.Kind != ir.NodeManagerLoop {
        t.Errorf("Kind = %q, want %q", n.Kind, ir.NodeManagerLoop)
    }
    cfg, ok := n.Config.(ir.ManagerLoopConfig)
    if !ok {
        t.Fatalf("Config = %T, want ManagerLoopConfig", n.Config)
    }
    if cfg.SubgraphRef != "quality_loop" {
        t.Errorf("SubgraphRef = %q, want %q", cfg.SubgraphRef, "quality_loop")
    }
    if cfg.PollInterval != 30*time.Second {
        t.Errorf("PollInterval = %v, want 30s", cfg.PollInterval)
    }
    if cfg.MaxCycles != 12 {
        t.Errorf("MaxCycles = %d, want 12", cfg.MaxCycles)
    }
    if n.Label != "Quality Gate Supervisor" {
        t.Errorf("Label = %q", n.Label)
    }
}
```

- [ ] **Step 2.2: Run it to confirm failure**

```shell
just test-pkg parser
```

Expected: FAIL — either `unexpected top-level identifier: manager_loop` or a nil Config type assertion.

- [ ] **Step 2.3: Register keyword dispatch**

In `parser/parser.go:79-80`, extend `workflowNodeKinds`:

```go
var workflowNodeKinds = map[string]bool{
    "agent": true, "human": true, "tool": true,
    "subgraph": true, "conditional": true, "manager_loop": true,
}
```

- [ ] **Step 2.4: Register default config constructor**

In `parser/parse_nodes.go:11-17`, extend `defaultNodeConfigs`:

```go
var defaultNodeConfigs = map[ir.NodeKind]func() ir.NodeConfig{
    ir.NodeAgent:       func() ir.NodeConfig { return ir.AgentConfig{Params: make(map[string]string)} },
    ir.NodeHuman:       func() ir.NodeConfig { return ir.HumanConfig{} },
    ir.NodeTool:        func() ir.NodeConfig { return ir.ToolConfig{} },
    ir.NodeSubgraph:    func() ir.NodeConfig { return ir.SubgraphConfig{Params: make(map[string]string)} },
    ir.NodeConditional: func() ir.NodeConfig { return ir.ConditionalConfig{} },
    ir.NodeManagerLoop: func() ir.NodeConfig { return ir.ManagerLoopConfig{SteerContext: make(map[string]string)} },
}
```

- [ ] **Step 2.5: Add the field handler dispatch arm**

In `parser/parse_nodes.go`, extend `applySecondaryConfigField` (lines 151-163):

```go
func (p *Parser) applySecondaryConfigField(n *ir.Node, key, val string, loc ir.SourceLocation) {
    switch cfg := n.Config.(type) {
    case ir.ToolConfig:
        p.applyToolField(&cfg, key, val, loc)
        n.Config = cfg
    case ir.SubgraphConfig:
        p.applySubgraphField(&cfg, key, val, loc)
        n.Config = cfg
    case ir.ManagerLoopConfig:
        p.applyManagerLoopField(&cfg, key, val, loc)
        n.Config = cfg
    case ir.ConditionalConfig:
        p.emitUnknownFieldHint("conditional", key, loc)
    }
}
```

- [ ] **Step 2.6: Implement `applyManagerLoopField`**

Append to `parser/parse_nodes.go` (after `applySubgraphField` at line 425):

```go
// applyManagerLoopField applies manager_loop-specific configuration fields.
func (p *Parser) applyManagerLoopField(cfg *ir.ManagerLoopConfig, key, val string, loc ir.SourceLocation) {
    if p.applyManagerLoopStringField(cfg, key, val) {
        return
    }
    if p.applyManagerLoopParsedField(cfg, key, val, loc) {
        return
    }
    p.emitUnknownFieldHint("manager_loop", key, loc)
}

// applyManagerLoopStringField handles string/condition fields. Returns true if handled.
func (p *Parser) applyManagerLoopStringField(cfg *ir.ManagerLoopConfig, key, val string) bool {
    switch key {
    case "subgraph_ref":
        cfg.SubgraphRef = val
    case "stop_condition":
        cfg.StopCondition = &ir.Condition{Raw: val}
    case "steer_condition":
        cfg.SteerCondition = &ir.Condition{Raw: val}
    default:
        return false
    }
    return true
}

// applyManagerLoopParsedField handles duration/int/map fields. Returns true if handled.
func (p *Parser) applyManagerLoopParsedField(cfg *ir.ManagerLoopConfig, key, val string, loc ir.SourceLocation) bool {
    switch key {
    case "poll_interval":
        cfg.PollInterval = p.parseDuration(val, key, loc)
    case "max_cycles":
        cfg.MaxCycles = p.parseInt(val, key, loc)
    case "steer_context":
        cfg.SteerContext = p.parseSteerContext(val)
    default:
        return false
    }
    return true
}

// parseSteerContext accepts both inline CSV ("k=v,k=v") and block-form content
// (one "k: v" per line, same as parseParamsBlock). Returns an empty map on empty input.
func (p *Parser) parseSteerContext(raw string) map[string]string {
    raw = strings.TrimSpace(raw)
    if raw == "" {
        return map[string]string{}
    }
    if strings.Contains(raw, "\n") {
        return p.parseParamsBlock(raw)
    }
    return p.parseSteerContextInline(raw)
}

// parseSteerContextInline parses "k=v, k=v, k=v" into a map.
func (p *Parser) parseSteerContextInline(raw string) map[string]string {
    out := make(map[string]string)
    for _, part := range strings.Split(raw, ",") {
        kv := strings.SplitN(strings.TrimSpace(part), "=", 2)
        if len(kv) != 2 {
            p.diagnostics = append(p.diagnostics,
                fmt.Sprintf("steer_context entry %q must be key=value", strings.TrimSpace(part)))
            continue
        }
        k := strings.TrimSpace(kv[0])
        v := strings.TrimSpace(kv[1])
        if k == "" {
            continue
        }
        if _, exists := out[k]; exists {
            p.diagnostics = append(p.diagnostics,
                fmt.Sprintf("duplicate steer_context key %q (last value wins)", k))
        }
        out[k] = unquoteRaw(v)
    }
    return out
}
```

Complexity note: each helper stays under the 5/7 cyclomatic/cognitive limit. If `parseSteerContextInline` trips the cap, extract the per-entry body into `applySteerContextEntry(out map[string]string, raw string)`.

- [ ] **Step 2.7: Run the parser tests**

```shell
just test-pkg parser
```

Expected: PASS.

- [ ] **Step 2.8: Add inline-CSV and block-form parser tests**

Extend `parser/parser_test.go`:

```go
func TestParseManagerLoop_SteerContextInline(t *testing.T) {
    src := `workflow W
  start: M
  exit: M
  manager_loop M
    subgraph_ref: inner
    steer_context: hint=speed_up, priority=high
  edges
    M -> M
`
    w, diags, err := Parse(src)
    if err != nil || len(diags) != 0 {
        t.Fatalf("Parse: err=%v diags=%v", err, diags)
    }
    cfg := w.Nodes[0].Config.(ir.ManagerLoopConfig)
    if cfg.SteerContext["hint"] != "speed_up" || cfg.SteerContext["priority"] != "high" {
        t.Errorf("SteerContext = %v", cfg.SteerContext)
    }
}

func TestParseManagerLoop_SteerContextBlock(t *testing.T) {
    src := `workflow W
  start: M
  exit: M
  manager_loop M
    subgraph_ref: inner
    steer_context:
      hint: speed_up
      priority: high
  edges
    M -> M
`
    w, diags, err := Parse(src)
    if err != nil || len(diags) != 0 {
        t.Fatalf("Parse: err=%v diags=%v", err, diags)
    }
    cfg := w.Nodes[0].Config.(ir.ManagerLoopConfig)
    if cfg.SteerContext["hint"] != "speed_up" || cfg.SteerContext["priority"] != "high" {
        t.Errorf("SteerContext = %v", cfg.SteerContext)
    }
}

func TestParseManagerLoop_UnknownFieldHint(t *testing.T) {
    src := `workflow W
  start: M
  exit: M
  manager_loop M
    subgraph_ref: inner
    bogus_field: value
  edges
    M -> M
`
    _, diags, err := Parse(src)
    if err != nil {
        t.Fatalf("Parse error: %v", err)
    }
    found := false
    for _, d := range diags {
        if strings.Contains(d, "manager_loop") && strings.Contains(d, "bogus_field") {
            found = true
        }
    }
    if !found {
        t.Errorf("expected manager_loop unknown-field diagnostic; got %v", diags)
    }
}
```

- [ ] **Step 2.9: Run tests**

```shell
just test-pkg parser
```

Expected: PASS (all four tests).

- [ ] **Step 2.10: Commit**

```bash
git add parser/parser.go parser/parse_nodes.go parser/parser_test.go
git commit -m "feat(parser): parse manager_loop node declarations

Adds keyword dispatch, default ManagerLoopConfig constructor, per-field
handlers for subgraph_ref/poll_interval/max_cycles/stop_condition/
steer_condition/steer_context. steer_context accepts both inline CSV
('k=v,k=v') and the block form used by subgraph.params."
```

---

## Task 3: Extend `simulate.EnsureConditionsParsed` to walk manager_loop nodes

Per CLAUDE.md: "Any code reading `Condition.Parsed` must ensure it's been called first." The linter calls `EnsureConditionsParsed(w)` at `validator/lint.go:20` before running checks. Today that function only walks edges. We need it to also visit `ManagerLoopConfig.StopCondition` and `.SteerCondition`.

**Files:**
- Modify: `simulate/condition.go:265-288`

- [ ] **Step 3.1: Write the failing test**

Append to `simulate/condition_test.go`:

```go
func TestEnsureConditionsParsed_ManagerLoop(t *testing.T) {
    stop := &ir.Condition{Raw: "stack.child.cycles = 10"}
    steer := &ir.Condition{Raw: "stack.child.cycles = 5"}
    w := &ir.Workflow{
        Name:  "W",
        Start: "M",
        Exit:  "M",
        Nodes: []*ir.Node{
            {ID: "M", Kind: ir.NodeManagerLoop, Config: ir.ManagerLoopConfig{
                SubgraphRef:    "inner",
                StopCondition:  stop,
                SteerCondition: steer,
            }},
        },
    }
    if err := EnsureConditionsParsed(w); err != nil {
        t.Fatalf("EnsureConditionsParsed error: %v", err)
    }
    if stop.Parsed == nil {
        t.Errorf("StopCondition.Parsed still nil after EnsureConditionsParsed")
    }
    if steer.Parsed == nil {
        t.Errorf("SteerCondition.Parsed still nil after EnsureConditionsParsed")
    }
}
```

- [ ] **Step 3.2: Run it — expect nil Parsed**

```shell
just test-pkg simulate
```

Expected: FAIL with "StopCondition.Parsed still nil".

- [ ] **Step 3.3: Extend `EnsureConditionsParsed`**

Replace lines 268-275 in `simulate/condition.go`:

```go
// EnsureConditionsParsed walks all edges and manager_loop nodes in a workflow
// and ensures that any Condition with a Raw string but nil Parsed field
// gets parsed. This is needed because the .dip parser stores Raw text only;
// the AST is lazily populated before lint checks run and before simulation.
func EnsureConditionsParsed(w *ir.Workflow) error {
    for _, e := range w.Edges {
        if err := ensureEdgeConditionParsed(e); err != nil {
            return err
        }
    }
    for _, n := range w.Nodes {
        if err := ensureNodeConditionsParsed(n); err != nil {
            return err
        }
    }
    return nil
}

// ensureNodeConditionsParsed walks node-level conditions (currently manager_loop only).
func ensureNodeConditionsParsed(n *ir.Node) error {
    cfg, ok := n.Config.(ir.ManagerLoopConfig)
    if !ok {
        return nil
    }
    if err := ensureConditionParsed(cfg.StopCondition, n.ID, "stop_condition"); err != nil {
        return err
    }
    return ensureConditionParsed(cfg.SteerCondition, n.ID, "steer_condition")
}

// ensureConditionParsed parses a single Condition if it has Raw but no Parsed.
func ensureConditionParsed(c *ir.Condition, nodeID, field string) error {
    if c == nil || c.Parsed != nil || c.Raw == "" {
        return nil
    }
    parsed, err := ParseCondition(c.Raw)
    if err != nil {
        return fmt.Errorf("node %s %s: invalid condition %q: %w", nodeID, field, c.Raw, err)
    }
    c.Parsed = parsed
    return nil
}
```

- [ ] **Step 3.4: Run tests**

```shell
just test-pkg simulate
```

Expected: PASS.

- [ ] **Step 3.5: Commit**

```bash
git add simulate/condition.go simulate/condition_test.go
git commit -m "feat(simulate): parse manager_loop stop/steer conditions on demand

Extends EnsureConditionsParsed to walk ManagerLoopConfig so StopCondition
and SteerCondition get their Parsed AST populated before lint and
simulation run. Matches the pattern used for edge conditions."
```

---

## Task 4: Format manager_loop nodes

**Files:**
- Modify: `formatter/format.go` (add a `ManagerLoopConfig` arm to `writeSecondaryConfigFields` around line 295)

- [ ] **Step 4.1: Write the failing test**

Append to `formatter/format_test.go`:

```go
func TestFormatManagerLoop_RoundTrip(t *testing.T) {
    src := `workflow W
  start: M
  exit: M

  manager_loop M
    label: "Quality Gate"
    subgraph_ref: quality_loop
    poll_interval: 30s
    max_cycles: 12
    stop_condition: stack.child.cycles = 10
    steer_condition: stack.child.cycles = 5
    steer_context: hint=speed_up, priority=high

  edges
    M -> M
`
    w, _, err := parser.Parse(src)
    if err != nil {
        t.Fatalf("Parse: %v", err)
    }
    out := Format(w)

    // Round-trip: reparse the formatted output.
    w2, diags, err := parser.Parse(out)
    if err != nil || len(diags) != 0 {
        t.Fatalf("reparse of formatted output: err=%v diags=%v\n%s", err, diags, out)
    }
    cfg := w2.Nodes[0].Config.(ir.ManagerLoopConfig)
    if cfg.SubgraphRef != "quality_loop" || cfg.MaxCycles != 12 || cfg.PollInterval != 30*time.Second {
        t.Errorf("round-trip lost fields: %+v", cfg)
    }
    if cfg.SteerContext["hint"] != "speed_up" {
        t.Errorf("round-trip lost steer_context: %v", cfg.SteerContext)
    }
    // Idempotency: formatting the formatted output returns the same string.
    if Format(w2) != out {
        t.Errorf("format is not idempotent")
    }
}
```

- [ ] **Step 4.2: Run — expect failure**

```shell
just test-pkg formatter
```

Expected: FAIL — output missing manager_loop block or round-trip loses fields.

- [ ] **Step 4.3: Add the formatter arm**

Extend `writeSecondaryConfigFields` in `formatter/format.go:295-305`:

```go
func writeSecondaryConfigFields(wr *writer, n *ir.Node) {
    switch cfg := n.Config.(type) {
    case ir.ToolConfig:
        writeToolFields(wr, n, cfg)
    case ir.SubgraphConfig:
        writeSubgraphFields(wr, n, cfg)
    case ir.ManagerLoopConfig:
        writeManagerLoopFields(wr, n, cfg)
    case ir.ConditionalConfig:
        writeConditionalFields(wr, n)
    }
}
```

Add these helpers near the subgraph writer (follow the style of `writeSubgraphFields`):

```go
func writeManagerLoopFields(wr *writer, n *ir.Node, cfg ir.ManagerLoopConfig) {
    writeCommonNodeFields(wr, n)
    writeRetryFields(wr, n)
    writeIOFields(wr, n)
    if cfg.SubgraphRef != "" {
        wr.line("subgraph_ref: %s", quoteValue(cfg.SubgraphRef))
    }
    if cfg.PollInterval != 0 {
        wr.line("poll_interval: %s", cfg.PollInterval.String())
    }
    if cfg.MaxCycles != 0 {
        wr.line("max_cycles: %d", cfg.MaxCycles)
    }
    writeManagerLoopConditions(wr, cfg)
    writeSteerContext(wr, cfg.SteerContext)
}

func writeManagerLoopConditions(wr *writer, cfg ir.ManagerLoopConfig) {
    if cfg.StopCondition != nil && cfg.StopCondition.Raw != "" {
        wr.line("stop_condition: %s", cfg.StopCondition.Raw)
    }
    if cfg.SteerCondition != nil && cfg.SteerCondition.Raw != "" {
        wr.line("steer_condition: %s", cfg.SteerCondition.Raw)
    }
}

// writeSteerContext emits steer_context as a block (one key per line) so
// round-trips through the block form stay canonical. Inline form is input-only.
func writeSteerContext(wr *writer, m map[string]string) {
    if len(m) == 0 {
        return
    }
    wr.line("steer_context:")
    wr.push()
    for _, k := range sortedKeys(m) {
        wr.line("%s: %s", k, quoteValue(m[k]))
    }
    wr.pop()
}
```

If `sortedKeys` doesn't already exist in `formatter/`, check for an equivalent helper first (`writeSubgraphParams` likely sorts; reuse whatever it uses).

- [ ] **Step 4.4: Run tests**

```shell
just test-pkg formatter
```

Expected: PASS.

- [ ] **Step 4.5: Commit**

```bash
git add formatter/format.go formatter/format_test.go
git commit -m "feat(formatter): write manager_loop nodes canonically

Emits subgraph_ref, poll_interval, max_cycles, stop_condition,
steer_condition as individual fields and steer_context as a block
form (sorted keys) so round-trip through parser→formatter→parser is
idempotent."
```

---

## Task 5: DOT export — shape=house plus flat attrs

**Files:**
- Modify: `export/dot.go:113-121` (nodeShapes), `:217-226` (applySemanticStructuralAttrs)

- [ ] **Step 5.1: Write the failing test**

Append to `export/dot_test.go`:

```go
func TestExportDOT_ManagerLoop(t *testing.T) {
    w := &ir.Workflow{
        Name:  "W",
        Start: "M",
        Exit:  "M",
        Nodes: []*ir.Node{
            {ID: "M", Kind: ir.NodeManagerLoop, Label: "Supervisor", Config: ir.ManagerLoopConfig{
                SubgraphRef:    "inner",
                PollInterval:   30 * time.Second,
                MaxCycles:      12,
                StopCondition:  &ir.Condition{Raw: "stack.child.cycles = 10"},
                SteerCondition: &ir.Condition{Raw: "stack.child.cycles = 5"},
                SteerContext:   map[string]string{"hint": "speed_up", "priority": "high"},
            }},
        },
        Edges: []*ir.Edge{{From: "M", To: "M"}},
    }
    out := Export(w, ExportOptions{})
    assertContains(t, out, `shape=house`)
    assertContains(t, out, `subgraph_ref="inner"`)
    assertContains(t, out, `poll_interval="30s"`)
    assertContains(t, out, `max_cycles="12"`)
    assertContains(t, out, `stop_condition="stack.child.cycles = 10"`)
    assertContains(t, out, `steer_condition="stack.child.cycles = 5"`)
    // steer_context flattens to canonical "k=v,k=v" (sorted).
    assertContains(t, out, `steer_context="hint=speed_up,priority=high"`)
}
```

(If `assertContains` doesn't exist, inline `if !strings.Contains(out, sub) { t.Errorf(...) }`.)

- [ ] **Step 5.2: Run — expect failure**

```shell
just test-pkg export
```

Expected: FAIL — `shape=box` (the fallback) instead of `shape=house`.

- [ ] **Step 5.3: Add shape mapping**

Extend `nodeShapes` in `export/dot.go:113-121`:

```go
var nodeShapes = map[ir.NodeKind]string{
    ir.NodeAgent:       "box",
    ir.NodeHuman:       "hexagon",
    ir.NodeTool:        "parallelogram",
    ir.NodeParallel:    "component",
    ir.NodeFanIn:       "tripleoctagon",
    ir.NodeSubgraph:    "tab",
    ir.NodeConditional: "diamond",
    ir.NodeManagerLoop: "house",
}
```

- [ ] **Step 5.4: Add config attr handler**

Extend `applySemanticStructuralAttrs` (around line 217) with a `ManagerLoopConfig` case, and add a new helper:

```go
func applySemanticStructuralAttrs(attrs map[string]string, cfg interface{}) {
    switch c := cfg.(type) {
    case ir.SubgraphConfig:
        applySubgraphAttrs(attrs, c)
    case ir.ParallelConfig:
        applyParallelAttrs(attrs, c)
    case ir.FanInConfig:
        applyFanInAttrs(attrs, c)
    case ir.ManagerLoopConfig:
        applyManagerLoopAttrs(attrs, c)
    }
}

func applyManagerLoopAttrs(attrs map[string]string, cfg ir.ManagerLoopConfig) {
    if cfg.SubgraphRef != "" {
        attrs["subgraph_ref"] = cfg.SubgraphRef
    }
    if cfg.PollInterval != 0 {
        attrs["poll_interval"] = cfg.PollInterval.String()
    }
    if cfg.MaxCycles != 0 {
        attrs["max_cycles"] = fmt.Sprintf("%d", cfg.MaxCycles)
    }
    if cfg.StopCondition != nil && cfg.StopCondition.Raw != "" {
        attrs["stop_condition"] = cfg.StopCondition.Raw
    }
    if cfg.SteerCondition != nil && cfg.SteerCondition.Raw != "" {
        attrs["steer_condition"] = cfg.SteerCondition.Raw
    }
    if s := flattenSteerContext(cfg.SteerContext); s != "" {
        attrs["steer_context"] = s
    }
}

// flattenSteerContext produces canonical sorted "k=v,k=v" from the map.
func flattenSteerContext(m map[string]string) string {
    if len(m) == 0 {
        return ""
    }
    keys := make([]string, 0, len(m))
    for k := range m {
        keys = append(keys, k)
    }
    sort.Strings(keys)
    parts := make([]string, 0, len(keys))
    for _, k := range keys {
        parts = append(parts, k+"="+m[k])
    }
    return strings.Join(parts, ",")
}
```

Add `"sort"` to the import block if not already present.

- [ ] **Step 5.5: Run tests**

```shell
just test-pkg export
```

Expected: PASS.

- [ ] **Step 5.6: Commit**

```bash
git add export/dot.go export/dot_test.go
git commit -m "feat(export): emit manager_loop nodes as DOT shape=house

Flat attrs: subgraph_ref, poll_interval, max_cycles, stop_condition,
steer_condition, and steer_context (flattened to canonical sorted
k=v,k=v so round-trip through the runtime's DOT adapter is lossless)."
```

---

## Task 6: DOT migrate — shape=house → NodeManagerLoop

**Files:**
- Modify: `migrate/migrate.go:44-51` (shapeToKind), `:256-270` (buildOtherConfig)

- [ ] **Step 6.1: Write the failing test**

Append to `migrate/migrate_test.go` (or the nearest existing test file):

```go
func TestMigrate_ManagerLoop(t *testing.T) {
    dot := `digraph W {
  M [shape=house, label="Supervisor", subgraph_ref="inner", poll_interval="30s", max_cycles="12", stop_condition="stack.child.cycles = 10", steer_condition="stack.child.cycles = 5", steer_context="hint=speed_up,priority=high"];
  M -> M;
}`
    w, err := Migrate(dot)
    if err != nil {
        t.Fatalf("Migrate: %v", err)
    }
    var n *ir.Node
    for _, node := range w.Nodes {
        if node.ID == "M" {
            n = node
        }
    }
    if n == nil || n.Kind != ir.NodeManagerLoop {
        t.Fatalf("got kind %v, want NodeManagerLoop", n)
    }
    cfg, ok := n.Config.(ir.ManagerLoopConfig)
    if !ok {
        t.Fatalf("Config = %T", n.Config)
    }
    if cfg.SubgraphRef != "inner" || cfg.MaxCycles != 12 || cfg.PollInterval != 30*time.Second {
        t.Errorf("lost fields: %+v", cfg)
    }
    if cfg.SteerContext["hint"] != "speed_up" {
        t.Errorf("SteerContext = %v", cfg.SteerContext)
    }
    if cfg.StopCondition == nil || cfg.StopCondition.Raw != "stack.child.cycles = 10" {
        t.Errorf("StopCondition = %+v", cfg.StopCondition)
    }
}
```

- [ ] **Step 6.2: Run — expect failure**

```shell
just test-pkg migrate
```

Expected: FAIL — `house` shape defaults to `NodeAgent`.

- [ ] **Step 6.3: Add shape mapping**

Extend `shapeToKind` in `migrate/migrate.go:44-51`:

```go
var shapeToKind = map[string]ir.NodeKind{
    "box":           ir.NodeAgent,
    "hexagon":       ir.NodeHuman,
    "parallelogram": ir.NodeTool,
    "component":     ir.NodeParallel,
    "tripleoctagon": ir.NodeFanIn,
    "tab":           ir.NodeSubgraph,
    "house":         ir.NodeManagerLoop,
}
```

- [ ] **Step 6.4: Add config builder**

Extend `buildOtherConfig` at line 256-270:

```go
func buildOtherConfig(kind ir.NodeKind, attrs map[string]string) ir.NodeConfig {
    switch kind {
    case ir.NodeParallel:
        return buildParallelConfig(attrs)
    case ir.NodeFanIn:
        return buildFanInConfig(attrs)
    case ir.NodeSubgraph:
        return buildSubgraphConfig(attrs)
    case ir.NodeConditional:
        return ir.ConditionalConfig{}
    case ir.NodeManagerLoop:
        return buildManagerLoopConfig(attrs)
    default:
        return ir.AgentConfig{}
    }
}
```

Append the builder (near `buildSubgraphConfig` at line 481):

```go
func buildManagerLoopConfig(attrs map[string]string) ir.ManagerLoopConfig {
    cfg := ir.ManagerLoopConfig{SteerContext: map[string]string{}}
    if v, ok := attrs["subgraph_ref"]; ok {
        cfg.SubgraphRef = v
    }
    if v, ok := attrs["poll_interval"]; ok {
        if d, err := time.ParseDuration(v); err == nil {
            cfg.PollInterval = d
        }
    }
    if v, ok := attrs["max_cycles"]; ok {
        if n, err := strconv.Atoi(v); err == nil {
            cfg.MaxCycles = n
        }
    }
    if v, ok := attrs["stop_condition"]; ok && v != "" {
        cfg.StopCondition = &ir.Condition{Raw: v}
    }
    if v, ok := attrs["steer_condition"]; ok && v != "" {
        cfg.SteerCondition = &ir.Condition{Raw: v}
    }
    if v, ok := attrs["steer_context"]; ok && v != "" {
        cfg.SteerContext = parseFlattenedSteerContext(v)
    }
    return cfg
}

// parseFlattenedSteerContext splits "k=v,k=v" into a map.
func parseFlattenedSteerContext(s string) map[string]string {
    out := map[string]string{}
    for _, part := range strings.Split(s, ",") {
        kv := strings.SplitN(strings.TrimSpace(part), "=", 2)
        if len(kv) != 2 {
            continue
        }
        out[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
    }
    return out
}
```

Check cyclomatic: `buildManagerLoopConfig` has 6 branches. If gocyclo complains, split the parsing into `applyManagerLoopScalarAttrs` + `applyManagerLoopConditionAttrs` helpers.

- [ ] **Step 6.5: Run tests**

```shell
just test-pkg migrate
```

Expected: PASS.

- [ ] **Step 6.6: Commit**

```bash
git add migrate/migrate.go migrate/migrate_test.go
git commit -m "feat(migrate): convert DOT shape=house to NodeManagerLoop

Reverses the export direction so 'dippin migrate' can convert runtime
DOT pipelines that use stack.manager_loop. Handles all six manager
attrs plus the flattened steer_context back into a map."
```

---

## Task 7: Register lint codes DIP135, DIP136, DIP137

**Files:**
- Modify: `validator/lint_codes.go`
- Modify: `validator/explanations.go` (**CI fails if any DIP code in `CodeDescription` lacks an entry in `Explanations`**)

- [ ] **Step 7.1: Write the failing test**

Add to `validator/explanations_test.go` (it probably already has a "every code has an explanation" check — this will catch missing entries automatically).

For now, add a direct assertion at the bottom:

```go
func TestManagerLoopCodesRegistered(t *testing.T) {
    for _, code := range []string{DIP135, DIP136, DIP137} {
        if _, ok := CodeDescription[code]; !ok {
            t.Errorf("%s missing from CodeDescription", code)
        }
        if _, ok := Explanations[code]; !ok {
            t.Errorf("%s missing from Explanations", code)
        }
    }
}
```

- [ ] **Step 7.2: Run — expect failure**

```shell
just test-pkg validator
```

Expected: FAIL — `undefined: DIP135`.

- [ ] **Step 7.3: Allocate codes**

Extend `validator/lint_codes.go`:

```go
const (
    // ... existing DIP101–DIP134 ...
    DIP135 = "DIP135" // manager_loop subgraph_ref missing or file does not exist
    DIP136 = "DIP136" // manager_loop control field has invalid value (poll_interval/max_cycles/steer_context)
    DIP137 = "DIP137" // unbounded manager_loop: no stop_condition and no max_cycles
)

func init() {
    // ... existing descriptions ...
    CodeDescription[DIP135] = "manager_loop subgraph_ref missing or file does not exist"
    CodeDescription[DIP136] = "manager_loop control field has invalid value"
    CodeDescription[DIP137] = "unbounded manager_loop: no stop_condition and no max_cycles"
}
```

- [ ] **Step 7.4: Add Explanations entries**

In `validator/explanations.go`, append inside `addLintExplanations()` (find the end of the existing entries and add these):

```go
Explanations[DIP135] = Explanation{
    Code:    DIP135,
    Summary: CodeDescription[DIP135],
    Trigger: "A manager_loop node either has no subgraph_ref field, or the referenced file does not exist on disk.",
    Fix:     "Set subgraph_ref to the path of an existing .dip file defining the child pipeline.",
    Example: "manager_loop M\n  subgraph_ref: quality_loop.dip  # file must exist relative to this workflow",
}
Explanations[DIP136] = Explanation{
    Code:    DIP136,
    Summary: CodeDescription[DIP136],
    Trigger: "A manager_loop field value failed to parse: poll_interval is not a valid duration, max_cycles is not a positive integer, or steer_context entries are not key=value pairs.",
    Fix:     "Use Go duration format for poll_interval (e.g., 30s, 1m30s), a positive integer for max_cycles, and comma-separated key=value pairs (or a block) for steer_context.",
    Example: "manager_loop M\n  subgraph_ref: inner\n  poll_interval: 30s\n  max_cycles: 12\n  steer_context: hint=speed_up, priority=high",
}
Explanations[DIP137] = Explanation{
    Code:    DIP137,
    Summary: CodeDescription[DIP137],
    Trigger: "A manager_loop node has neither a stop_condition nor a max_cycles cap, so supervision can run forever.",
    Fix:     "Set stop_condition to a predicate that terminates supervision (e.g., stack.child.cycles = 10), or set max_cycles to a positive integer bound.",
    Example: "manager_loop M\n  subgraph_ref: inner\n  stop_condition: stack.child.outcome = success\n  # or: max_cycles: 20",
}
```

- [ ] **Step 7.5: Run tests**

```shell
just test-pkg validator
```

Expected: PASS.

- [ ] **Step 7.6: Commit**

```bash
git add validator/lint_codes.go validator/explanations.go validator/explanations_test.go
git commit -m "feat(validator): allocate DIP135-137 for manager_loop nodes

DIP135: subgraph_ref missing or file does not exist (analogous to DIP126).
DIP136: control field has invalid value (poll_interval/max_cycles/steer_context).
DIP137: unbounded manager_loop (no stop_condition and no max_cycles) — DIP104 analog.

Splits per existing granularity precedent: DIP116/DIP130 bundle format
checks; per-field codes are reserved for enumerated value sets."
```

---

## Task 8: Implement the manager_loop linter

**Files:**
- Create: `validator/lint_manager_loop.go`
- Create: `validator/lint_manager_loop_test.go`
- Modify: `validator/lint.go` (register the new linter)

- [ ] **Step 8.1: Write the failing tests**

Create `validator/lint_manager_loop_test.go`:

```go
package validator

import (
    "strings"
    "testing"

    "github.com/2389-research/dippin-lang/ir"
)

func managerLoopWorkflow(cfg ir.ManagerLoopConfig) *ir.Workflow {
    return &ir.Workflow{
        Name:  "W",
        Start: "M",
        Exit:  "M",
        Nodes: []*ir.Node{{ID: "M", Kind: ir.NodeManagerLoop, Config: cfg}},
        Edges: []*ir.Edge{{From: "M", To: "M"}},
    }
}

func TestLintManagerLoop_DIP135_MissingRef(t *testing.T) {
    w := managerLoopWorkflow(ir.ManagerLoopConfig{MaxCycles: 5})
    diags := lintManagerLoop(w)
    if !hasCode(diags, DIP135) {
        t.Errorf("expected DIP135, got %v", diags)
    }
}

func TestLintManagerLoop_DIP135_RefFileDoesNotExist(t *testing.T) {
    w := managerLoopWorkflow(ir.ManagerLoopConfig{SubgraphRef: "does_not_exist.dip", MaxCycles: 5})
    w.Nodes[0].Source = ir.SourceLocation{File: "/tmp/fakeworkflow.dip"}
    diags := lintManagerLoop(w)
    if !hasCode(diags, DIP135) {
        t.Errorf("expected DIP135, got %v", diags)
    }
}

func TestLintManagerLoop_DIP136_BadDuration(t *testing.T) {
    // PollInterval is a time.Duration, so parse errors are caught at parse time.
    // DIP136 fires on negative durations or max_cycles < 0.
    w := managerLoopWorkflow(ir.ManagerLoopConfig{
        SubgraphRef: "inner.dip",
        PollInterval: -5,
        MaxCycles:   5,
    })
    diags := lintManagerLoop(w)
    if !hasCode(diags, DIP136) {
        t.Errorf("expected DIP136 for negative poll_interval, got %v", diags)
    }
}

func TestLintManagerLoop_DIP136_NegativeMaxCycles(t *testing.T) {
    w := managerLoopWorkflow(ir.ManagerLoopConfig{
        SubgraphRef: "inner.dip",
        MaxCycles:  -1,
    })
    diags := lintManagerLoop(w)
    if !hasCode(diags, DIP136) {
        t.Errorf("expected DIP136 for negative max_cycles, got %v", diags)
    }
}

func TestLintManagerLoop_DIP137_Unbounded(t *testing.T) {
    w := managerLoopWorkflow(ir.ManagerLoopConfig{SubgraphRef: "inner.dip"})
    // Avoid DIP135 for nonexistent ref by skipping file check (no Source).
    diags := lintManagerLoop(w)
    if !hasCode(diags, DIP137) {
        t.Errorf("expected DIP137 for no stop_condition and no max_cycles, got %v", diags)
    }
}

func TestLintManagerLoop_Clean(t *testing.T) {
    w := managerLoopWorkflow(ir.ManagerLoopConfig{
        SubgraphRef:   "inner.dip",
        MaxCycles:     5,
        StopCondition: &ir.Condition{Raw: "stack.child.cycles = 10"},
    })
    diags := lintManagerLoop(w)
    // No DIP136/DIP137. DIP135 may fire if the test worker has a Source set;
    // for this unit test, no Source => DIP135 skipped.
    for _, d := range diags {
        if d.Code == DIP136 || d.Code == DIP137 {
            t.Errorf("unexpected %s: %s", d.Code, d.Message)
        }
    }
}

// hasCode reports whether diags contains any diagnostic with the given code.
func hasCode(diags []Diagnostic, code string) bool {
    for _, d := range diags {
        if d.Code == code {
            return true
        }
    }
    return false
}

// (If hasCode already exists in another test file, remove the redefinition.)
```

Check whether `hasCode` already exists — `grep "func hasCode" validator/*_test.go`. If it does, drop the helper from this file.

- [ ] **Step 8.2: Run — expect failure**

```shell
just test-pkg validator
```

Expected: FAIL — `undefined: lintManagerLoop`.

- [ ] **Step 8.3: Implement the linter**

Create `validator/lint_manager_loop.go`:

```go
//go:build !wasm

package validator

import (
    "fmt"
    "os"
    "path/filepath"

    "github.com/2389-research/dippin-lang/ir"
)

// lintManagerLoop emits DIP135 (ref missing/missing file), DIP136 (invalid
// control field), and DIP137 (unbounded supervision).
func lintManagerLoop(w *ir.Workflow) []Diagnostic {
    var diags []Diagnostic
    for _, n := range w.Nodes {
        cfg, ok := n.Config.(ir.ManagerLoopConfig)
        if !ok {
            continue
        }
        diags = appendManagerLoopDiags(diags, n, cfg)
    }
    return diags
}

// appendManagerLoopDiags appends all DIP135/136/137 diagnostics for a node.
func appendManagerLoopDiags(diags []Diagnostic, n *ir.Node, cfg ir.ManagerLoopConfig) []Diagnostic {
    if d := checkManagerLoopRef(n, cfg); d != nil {
        diags = append(diags, *d)
    }
    diags = append(diags, checkManagerLoopControl(n, cfg)...)
    if d := checkManagerLoopUnbounded(n, cfg); d != nil {
        diags = append(diags, *d)
    }
    return diags
}

// checkManagerLoopRef emits DIP135 if subgraph_ref is empty or points to a
// non-existent file (when the file path can be resolved).
func checkManagerLoopRef(n *ir.Node, cfg ir.ManagerLoopConfig) *Diagnostic {
    if cfg.SubgraphRef == "" {
        return &Diagnostic{
            Code:     DIP135,
            Severity: SeverityWarning,
            Message:  fmt.Sprintf("manager_loop %q has no subgraph_ref", n.ID),
            Location: n.Source,
            Help:     "set subgraph_ref to the path of a .dip file defining the child pipeline",
        }
    }
    if !filepath.IsAbs(cfg.SubgraphRef) && n.Source.File == "" {
        return nil // cannot resolve
    }
    resolved := cfg.SubgraphRef
    if !filepath.IsAbs(resolved) && n.Source.File != "" {
        resolved = filepath.Join(filepath.Dir(n.Source.File), resolved)
    }
    if _, err := os.Stat(resolved); err != nil {
        return &Diagnostic{
            Code:     DIP135,
            Severity: SeverityWarning,
            Message:  fmt.Sprintf("manager_loop %q references %q which does not exist", n.ID, cfg.SubgraphRef),
            Location: n.Source,
            Help:     fmt.Sprintf("resolved path: %s", resolved),
        }
    }
    return nil
}

// checkManagerLoopControl emits DIP136 for each invalid control field.
func checkManagerLoopControl(n *ir.Node, cfg ir.ManagerLoopConfig) []Diagnostic {
    var out []Diagnostic
    if cfg.PollInterval < 0 {
        out = append(out, Diagnostic{
            Code:     DIP136,
            Severity: SeverityWarning,
            Message:  fmt.Sprintf("manager_loop %q poll_interval is negative (%s)", n.ID, cfg.PollInterval),
            Location: n.Source,
            Help:     "use a non-negative duration such as 30s; 0 means event-driven",
        })
    }
    if cfg.MaxCycles < 0 {
        out = append(out, Diagnostic{
            Code:     DIP136,
            Severity: SeverityWarning,
            Message:  fmt.Sprintf("manager_loop %q max_cycles is negative (%d)", n.ID, cfg.MaxCycles),
            Location: n.Source,
            Help:     "use a positive integer; 0 means unbounded",
        })
    }
    return out
}

// checkManagerLoopUnbounded emits DIP137 when both the stop_condition and
// max_cycles are unset — supervision will run forever.
func checkManagerLoopUnbounded(n *ir.Node, cfg ir.ManagerLoopConfig) *Diagnostic {
    hasStop := cfg.StopCondition != nil && cfg.StopCondition.Raw != ""
    hasMax := cfg.MaxCycles > 0
    if hasStop || hasMax {
        return nil
    }
    return &Diagnostic{
        Code:     DIP137,
        Severity: SeverityWarning,
        Message:  fmt.Sprintf("manager_loop %q is unbounded: no stop_condition and no max_cycles", n.ID),
        Location: n.Source,
        Help:     "set stop_condition (e.g., stack.child.outcome = success) or max_cycles to bound supervision",
    }
}
```

Wasm build tag matches `lint_subgraph.go`'s split — if CI compiles a wasm target, also create `lint_manager_loop_wasm.go` with a stub `lintManagerLoop` that skips the `os.Stat` check. Look at `lint_subgraph_wasm.go` as the template.

- [ ] **Step 8.4: Register the linter**

In `validator/lint.go`, add to the `diags = append(...)` chain (around line 50, next to `lintSubgraphRef`):

```go
diags = append(diags, lintManagerLoop(w)...)
```

- [ ] **Step 8.5: Run tests**

```shell
just test-pkg validator
```

Expected: PASS (all 6 manager_loop tests).

- [ ] **Step 8.6: Commit**

```bash
git add validator/lint_manager_loop.go validator/lint_manager_loop_test.go validator/lint.go
git commit -m "feat(validator): lint manager_loop nodes (DIP135-137)

DIP135 checks subgraph_ref presence and file existence.
DIP136 flags negative poll_interval or max_cycles.
DIP137 flags unbounded supervision (no stop_condition and no max_cycles)."
```

---

## Task 9: Extend DIP120 namespace check for `stack.child.*`

`stop_condition`/`steer_condition` reference `stack.child.cycles`, `stack.child.outcome`, etc. Today DIP120's `knownNamespaces` map only knows about `ctx`, `graph`, `params`. Without extending it, manager_loop conditions will fire spurious DIP120 "variable missing namespace prefix" warnings.

**Files:**
- Modify: `validator/lint.go:62-72` (`knownNamespaces` map)
- Check: `validator/lint_conditions.go` for how it consumes `knownNamespaces`

- [ ] **Step 9.1: Write the failing test**

Add to `validator/lint_conditions_test.go` (or nearest analog):

```go
func TestLintConditions_StackChildNamespace(t *testing.T) {
    w := &ir.Workflow{
        Name:  "W",
        Start: "M",
        Exit:  "M",
        Nodes: []*ir.Node{
            {ID: "M", Kind: ir.NodeManagerLoop, Config: ir.ManagerLoopConfig{
                SubgraphRef:   "inner.dip",
                MaxCycles:     5,
                StopCondition: &ir.Condition{Raw: "stack.child.cycles = 10"},
            }},
        },
        Edges: []*ir.Edge{{From: "M", To: "M"}},
    }
    _ = simulate.EnsureConditionsParsed(w)
    diags := Lint(w).Diagnostics
    for _, d := range diags {
        if d.Code == DIP120 {
            t.Errorf("stack.child.* should be a recognized namespace, got DIP120: %s", d.Message)
        }
    }
}
```

- [ ] **Step 9.2: Run — expect failure**

```shell
just test-pkg validator
```

Expected: FAIL — `stack` is not in `knownNamespaces`.

- [ ] **Step 9.3: Extend knownNamespaces**

In `validator/lint.go:62-72`:

```go
// knownNamespaces lists the valid namespace prefixes for variable references.
// Per §8.2 of the Dippin spec: ctx. (runtime context), graph. (workflow-level
// attributes), params. (module parameters for composition), stack. (supervisor
// state exposed by manager_loop, e.g., stack.child.cycles).
var knownNamespaces = map[string]bool{
    "ctx":    true,
    "graph":  true,
    "params": true,
    "stack":  true,
}
```

- [ ] **Step 9.4: Run the test**

```shell
just test-pkg validator
```

Expected: PASS.

- [ ] **Step 9.5: Commit**

```bash
git add validator/lint.go validator/lint_conditions_test.go
git commit -m "feat(validator): recognize stack.* namespace for manager_loop conditions

stop_condition and steer_condition reference stack.child.cycles,
stack.child.outcome, etc. Extends DIP120's knownNamespaces so these
don't fire spurious 'missing namespace prefix' warnings."
```

---

## Task 10: Add scaffold template `manager_loop`

**Files:**
- Modify: `scaffold/scaffold.go` (TemplateNames, templateBuilders, new builder function)

- [ ] **Step 10.1: Write the failing test**

Append to `scaffold/scaffold_test.go` (or the nearest file):

```go
func TestBuildManagerLoop(t *testing.T) {
    w, err := Build("manager_loop", "Supervisor")
    if err != nil {
        t.Fatalf("Build: %v", err)
    }
    if w.Name != "Supervisor" {
        t.Errorf("Name = %q", w.Name)
    }
    hasManager := false
    for _, n := range w.Nodes {
        if n.Kind == ir.NodeManagerLoop {
            hasManager = true
            cfg := n.Config.(ir.ManagerLoopConfig)
            if cfg.SubgraphRef == "" || cfg.MaxCycles == 0 {
                t.Errorf("manager_loop config missing required fields: %+v", cfg)
            }
        }
    }
    if !hasManager {
        t.Errorf("no NodeManagerLoop node in template output")
    }
}

func TestTemplateNames_IncludesManagerLoop(t *testing.T) {
    names := TemplateNames()
    for _, n := range names {
        if n == "manager_loop" {
            return
        }
    }
    t.Errorf("manager_loop missing from TemplateNames: %v", names)
}
```

- [ ] **Step 10.2: Run — expect failure**

```shell
just test-pkg scaffold
```

Expected: FAIL — unknown template.

- [ ] **Step 10.3: Add template**

Update `scaffold/scaffold.go`:

```go
func TemplateNames() []string {
    return []string{"conditional", "human-gate", "manager_loop", "minimal", "parallel", "review-loop"}
}

var templateBuilders = map[string]func(string) *ir.Workflow{
    "minimal":      buildMinimal,
    "parallel":     buildParallel,
    "conditional":  buildConditional,
    "review-loop":  buildReviewLoop,
    "human-gate":   buildHumanGate,
    "manager_loop": buildManagerLoop,
}
```

Append the builder (after `buildHumanGate`):

```go
func buildManagerLoop(name string) *ir.Workflow {
    return &ir.Workflow{
        Name:  name,
        Goal:  "Supervise a child pipeline with periodic steering",
        Start: "Supervise",
        Exit:  "Supervise",
        Nodes: []*ir.Node{
            {ID: "Supervise", Kind: ir.NodeManagerLoop, Label: "Quality Gate Supervisor", Config: ir.ManagerLoopConfig{
                SubgraphRef:    "child_pipeline.dip",
                PollInterval:   30 * time.Second,
                MaxCycles:      10,
                StopCondition:  &ir.Condition{Raw: "stack.child.outcome = success"},
                SteerCondition: &ir.Condition{Raw: "stack.child.cycles = 5"},
                SteerContext:   map[string]string{"hint": "halfway_through"},
            }},
        },
        Edges: []*ir.Edge{
            {From: "Supervise", To: "Supervise"},
        },
    }
}
```

Add `"time"` to imports if not already present.

- [ ] **Step 10.4: Run tests**

```shell
just test-pkg scaffold
```

Expected: PASS.

- [ ] **Step 10.5: Commit**

```bash
git add scaffold/scaffold.go scaffold/scaffold_test.go
git commit -m "feat(scaffold): add manager_loop template

'dippin scaffold manager_loop' emits a supervisor that polls a child
pipeline every 30s, caps at 10 cycles, and injects a steer_context
halfway through. Matches the one-kind-showcase pattern of the
conditional template."
```

---

## Task 11: Add LSP symbol mapping

**Files:**
- Modify: `lsp/symbols.go` (nodeSymbolKinds map)

- [ ] **Step 11.1: Look at the current mapping**

Read `lsp/symbols.go:67-74` (approx). The map is `ir.NodeKind → protocol.SymbolKind`.

- [ ] **Step 11.2: Write the failing test**

If there's a `symbols_test.go`, add:

```go
func TestNodeSymbolKind_ManagerLoop(t *testing.T) {
    got, ok := nodeSymbolKinds[ir.NodeManagerLoop]
    if !ok {
        t.Fatalf("ir.NodeManagerLoop missing from nodeSymbolKinds")
    }
    // Expect a structural kind, e.g., protocol.SymbolKindClass or SymbolKindInterface.
    if got == 0 {
        t.Errorf("nodeSymbolKinds[NodeManagerLoop] = 0 (invalid)")
    }
}
```

If there's no test file, skip the test — the compiler will catch unregistered kinds if they're used.

- [ ] **Step 11.3: Add the mapping**

Mirror whatever kind `NodeSubgraph` uses — manager_loop is structurally similar (it embeds a sub-pipeline). If subgraph maps to `protocol.SymbolKindClass`, do the same:

```go
var nodeSymbolKinds = map[ir.NodeKind]protocol.SymbolKind{
    ir.NodeAgent:       protocol.SymbolKindFunction,
    ir.NodeHuman:       protocol.SymbolKindEvent,
    ir.NodeTool:        protocol.SymbolKindMethod,
    ir.NodeParallel:    protocol.SymbolKindOperator,
    ir.NodeFanIn:       protocol.SymbolKindOperator,
    ir.NodeSubgraph:    protocol.SymbolKindClass,
    ir.NodeConditional: protocol.SymbolKindBoolean,
    ir.NodeManagerLoop: protocol.SymbolKindClass, // supervisor embedding a sub-pipeline
}
```

Use whichever values the existing map already uses; don't change existing entries.

- [ ] **Step 11.4: Run**

```shell
just test-pkg lsp
```

Expected: PASS or (if no test) the build still succeeds.

- [ ] **Step 11.5: Commit**

```bash
git add lsp/symbols.go lsp/symbols_test.go
git commit -m "feat(lsp): register NodeManagerLoop in symbol outline"
```

---

## Task 12: Add example and integration test

**Files:**
- Create: `examples/manager_loop_demo.dip`
- Create: `examples/child_pipeline.dip` (referenced by subgraph_ref so DIP135 doesn't fire)
- No test-code changes — `TestLintExamples` at `validator/lint_examples_test.go:15-51` auto-picks up new `.dip` files.

- [ ] **Step 12.1: Create the child pipeline**

`examples/child_pipeline.dip`:

```dippin
workflow ChildPipeline
  goal: "A minimal child pipeline supervised by manager_loop_demo."
  start: DoWork
  exit: Done

  agent DoWork
    prompt:
      Perform one cycle of supervised work.
      End with STATUS: success or STATUS: fail.
    auto_status: true

  agent Done
    prompt: Finalize this cycle's result.

  edges
    DoWork -> Done
```

- [ ] **Step 12.2: Create the demo workflow**

`examples/manager_loop_demo.dip`:

```dippin
workflow ManagerLoopDemo
  goal: "Supervise a child pipeline with periodic steering."
  start: Supervise
  exit: Supervise

  manager_loop Supervise
    label: "Quality Gate Supervisor"
    subgraph_ref: child_pipeline.dip
    poll_interval: 30s
    max_cycles: 12
    stop_condition: stack.child.outcome = success
    steer_condition: stack.child.cycles = 5
    steer_context:
      hint: halfway_through
      priority: high

  edges
    Supervise -> Supervise
```

- [ ] **Step 12.3: Run the full check**

```shell
just build && ./dippin validate examples/manager_loop_demo.dip && ./dippin lint examples/manager_loop_demo.dip && ./dippin fmt examples/manager_loop_demo.dip
```

Expected: all succeed, lint clean.

- [ ] **Step 12.4: Run the integration suite**

```shell
just test-pkg validator
```

`TestLintExamples` should parse and lint every `.dip` file including the new demo. Expected: PASS.

- [ ] **Step 12.5: Run validate-examples**

```shell
just validate-examples
```

Expected: "All examples valid."

- [ ] **Step 12.6: Commit**

```bash
git add examples/manager_loop_demo.dip examples/child_pipeline.dip
git commit -m "feat(examples): add manager_loop_demo.dip and child_pipeline.dip

Exercises every manager_loop field (subgraph_ref, poll_interval,
max_cycles, stop/steer conditions, steer_context block form). Picked
up automatically by TestLintExamples."
```

---

## Task 13: Update the EBNF grammar

**Files:**
- Modify: `docs/GRAMMAR.ebnf:65-67` (node_decl choice) and new production rule

- [ ] **Step 13.1: Add to `node_decl`**

Replace lines 65-67:

```ebnf
node_decl       = agent_node | human_node | tool_node
                | subgraph_node | conditional_node
                | parallel_node | fan_in_node
                | manager_loop_node ;
```

- [ ] **Step 13.2: Add the production rule**

Insert after the conditional_node definition (around line 123):

```ebnf
(* --- Manager Loop Node ------------------------------------ *)

manager_loop_node = "manager_loop" IDENTIFIER NEWLINE
                    INDENT { manager_loop_field | common_field } OUTDENT ;

manager_loop_field = "subgraph_ref"   ":" IDENTIFIER
                   | "poll_interval"  ":" DURATION
                   | "max_cycles"     ":" INTEGER
                   | "stop_condition" ":" field_value   (* condition expression, stored raw *)
                   | "steer_condition" ":" field_value
                   | "steer_context"  ":" steer_context_value ;

steer_context_value = field_value                      (* inline "k=v, k=v" *)
                    | NEWLINE INDENT { IDENTIFIER ":" field_value } OUTDENT ;  (* block form *)
```

- [ ] **Step 13.3: Update the canonical field-ordering comment**

Extend the block at the bottom (around line 305):

```ebnf
(* Manager loop: label, subgraph_ref, poll_interval, max_cycles,      *)
(*   stop_condition, steer_condition, steer_context                   *)
```

- [ ] **Step 13.4: Update the keywords list**

Line 321 lists contextual keywords. Add `manager_loop`:

```ebnf
(* workflow, agent, human, tool, subgraph, parallel, fan_in,           *)
(* manager_loop, edges, defaults, vars, when, and, or, not, contains,   *)
(* startswith, endswith, in, true, false, restart, label, weight        *)
```

- [ ] **Step 13.5: Commit**

```bash
git add docs/GRAMMAR.ebnf
git commit -m "docs(grammar): add manager_loop_node production rule"
```

---

## Task 14: Update tree-sitter grammar + commit generated files

**Files:**
- Modify: `editors/tree-sitter-dippin/grammar.js`
- Modify: `editors/tree-sitter-dippin/queries/highlights.scm`
- Modify: `editors/tree-sitter-dippin/test/corpus/basic.txt`
- Commit (first time): `editors/tree-sitter-dippin/src/parser.c`, `src/grammar.json`, `src/node-types.json`, `src/tree_sitter/*.h`
- Modify: `.gitignore` (ensure `editors/tree-sitter-dippin/node_modules/` is ignored)
- Modify: `Justfile` (add `tree-sitter-generate`, `tree-sitter-test` recipes)
- Modify: `.github/workflows/ci.yml` (add drift check)

- [ ] **Step 14.1: Extend `grammar.js`**

In `editors/tree-sitter-dippin/grammar.js:49-57`, add `$.manager_loop_node`:

```js
node_decl: ($) =>
  choice(
    $.agent_node,
    $.human_node,
    $.tool_node,
    $.subgraph_node,
    $.conditional_node,
    $.parallel_node,
    $.fan_in_node,
    $.manager_loop_node
  ),
```

Add the rule after `conditional_node` (around line 72):

```js
manager_loop_node: ($) =>
  seq("manager_loop", $.identifier, $._indent, repeat1(choice($.node_field, $._newline)), $._dedent),
```

- [ ] **Step 14.2: Update highlights**

In `editors/tree-sitter-dippin/queries/highlights.scm:10-18`, add `"manager_loop"`:

```scm
; Node kinds
[
  "agent"
  "human"
  "tool"
  "subgraph"
  "conditional"
  "parallel"
  "fan_in"
  "manager_loop"
] @type
```

And extend the identifier-in-declaration list (around line 73):

```scm
(manager_loop_node
  (identifier) @function)
```

- [ ] **Step 14.3: Add corpus test**

Append to `editors/tree-sitter-dippin/test/corpus/basic.txt`:

```text
================================================================================
Manager loop node
================================================================================

workflow Sup
  start: M
  exit: M

  manager_loop M
    subgraph_ref: inner
    poll_interval: 30s
    max_cycles: 12

  edges
    M -> M

--------------------------------------------------------------------------------

(source_file
  (workflow_decl
    (identifier)
    (workflow_body
      (workflow_field (field_value (raw_inline)))
      (workflow_field (field_value (raw_inline)))
      (node_decl
        (manager_loop_node
          (identifier)
          (node_field (field_name (identifier)) (field_value (raw_inline)))
          (node_field (field_name (identifier)) (field_value (raw_inline)))
          (node_field (field_name (identifier)) (field_value (raw_inline)))))
      (edges_section
        (edge_entry (identifier) (identifier))))))
```

- [ ] **Step 14.4: Add `just` recipes**

Append to `Justfile`:

```make
# Regenerate tree-sitter parser from grammar.js
tree-sitter-generate:
    cd editors/tree-sitter-dippin && npx tree-sitter generate

# Run tree-sitter corpus tests
tree-sitter-test: tree-sitter-generate
    cd editors/tree-sitter-dippin && npx tree-sitter test
```

- [ ] **Step 14.5: Generate + run tests**

```shell
cd editors/tree-sitter-dippin && npm install  # only if node_modules is missing
just tree-sitter-generate
just tree-sitter-test
```

Expected: all corpus cases pass. If the generator isn't installed globally, `npx tree-sitter-cli` is pulled from `devDependencies`.

- [ ] **Step 14.6: Update .gitignore**

Ensure `editors/tree-sitter-dippin/node_modules/` is ignored. Append to `.gitignore` if not:

```text
# Tree-sitter build deps
editors/tree-sitter-dippin/node_modules/
```

- [ ] **Step 14.7: Add CI drift check**

In `.github/workflows/ci.yml`, add a new job (copy the pattern from the `Verify generated-spec.md is current` step at lines 41-48):

```yaml
  tree-sitter:
    name: Tree-sitter parser
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with:
          node-version: '20'
      - name: Install tree-sitter-cli
        run: npm install -g tree-sitter-cli@^0.24.0
      - name: Regenerate
        working-directory: editors/tree-sitter-dippin
        run: tree-sitter generate
      - name: Check for drift
        run: |
          if ! git diff --exit-code editors/tree-sitter-dippin/src; then
            echo "::error::tree-sitter generated files are stale — run 'just tree-sitter-generate' and commit"
            exit 1
          fi
      - name: Run corpus tests
        working-directory: editors/tree-sitter-dippin
        run: tree-sitter test
```

- [ ] **Step 14.8: Stage the generated files**

```shell
git add editors/tree-sitter-dippin/src/parser.c editors/tree-sitter-dippin/src/grammar.json editors/tree-sitter-dippin/src/node-types.json editors/tree-sitter-dippin/src/tree_sitter/
git add editors/tree-sitter-dippin/tree-sitter.json  # if untracked
git add editors/tree-sitter-dippin/grammar.js editors/tree-sitter-dippin/queries/highlights.scm editors/tree-sitter-dippin/test/corpus/basic.txt
git add Justfile .gitignore .github/workflows/ci.yml
```

Verify before committing: `git status` shows no stray file in `node_modules/`. If `package-lock.json` is new (not previously tracked), decide: commit it (reproducible installs) or extend `.gitignore` for consistency with the existing setup. The git status at plan time shows `package-lock.json` as untracked — **commit it**.

```shell
git add editors/tree-sitter-dippin/package-lock.json
```

- [ ] **Step 14.9: Commit**

```bash
git commit -m "feat(tree-sitter): add manager_loop rule and commit generated parser

- grammar.js: new manager_loop_node rule in node_decl choice
- highlights.scm: add 'manager_loop' to @type list and identifier @function
- test/corpus/basic.txt: new case covering a manager_loop declaration
- src/parser.c etc. now committed (consumers compile directly; Zed pins
  commit SHA and expects src/ to exist); gitignore node_modules
- Justfile: tree-sitter-generate + tree-sitter-test recipes
- ci.yml: new job checks generated files don't drift from grammar.js"
```

- [ ] **Step 14.10: Bump Zed extension commit SHA**

After merging this PR, the Zed extension needs to point at the new SHA. Don't do this in the PR — do it in a follow-up:

```shell
# Post-merge, on main:
git rev-parse HEAD   # copy the merge SHA
# Edit editors/zed-dippin/extension.toml, update the commit = "..." line
# Commit + tag or push
```

Add a TODO to the PR body so this isn't forgotten.

---

## Task 15: Update VS Code TextMate grammar

**Files:**
- Modify: `editors/vscode/syntaxes/dippin.tmLanguage.json`

- [ ] **Step 15.1: Extend the node-declaration regex**

Line 30:

```json
"node-declaration": {
  "match": "^\\s*(agent|tool|human|subgraph|conditional|parallel|fan_in|manager_loop)\\s+(\\w+)",
  "captures": {
    "1": { "name": "keyword.control.node.dippin" },
    "2": { "name": "entity.name.function.node.dippin" }
  }
},
```

Note: `parallel` and `fan_in` are added here as the pre-existing bug fix (they're still handled separately by `#parallel-fanin` for arrow forms, but plain block-form declarations were unhighlighted).

- [ ] **Step 15.2: Extend the field-name regex**

Line 148 — add `subgraph_ref`, `poll_interval`, `max_cycles`, `stop_condition`, `steer_condition`, `steer_context` to the field alternation:

```json
{
  "match": "^\\s*(label|model|provider|mode|default|ref|subgraph_ref|timeout|poll_interval|max_cycles|fidelity|reasoning_effort|goal_gate|auto_status|max_turns|max_retries|base_delay|retry_policy|retry_target|fallback_target|max_restarts|restart_target|cache_tools|compaction|stop_condition|steer_condition|steer_context|class|reads|writes)\\s*(:)",
  "captures": {
    "1": { "name": "entity.name.tag.field.dippin" },
    "2": { "name": "punctuation.separator.key-value.dippin" }
  }
},
```

- [ ] **Step 15.3: Extend the condition namespace regex**

Line 121 — add `stack` to the namespace alternation:

```json
{
  "match": "\\b(ctx|graph|params|stack)(\\.)([\\w.-]+)\\b",
  "captures": {
    "1": { "name": "variable.language.namespace.dippin" },
    "2": { "name": "punctuation.accessor.dippin" },
    "3": { "name": "variable.other.member.dippin" }
  }
}
```

- [ ] **Step 15.4: Commit**

```bash
git add editors/vscode/syntaxes/dippin.tmLanguage.json
git commit -m "feat(vscode): highlight manager_loop plus fix parallel/fan_in

- Adds manager_loop keyword to node-declaration regex
- Adds subgraph_ref/poll_interval/max_cycles/stop_condition/steer_condition/
  steer_context to field regex
- Adds stack.* to the recognized condition namespaces
- Fixes a pre-existing bug: parallel/fan_in were missing from the
  node-declaration alternation (block-form declarations went unhighlighted)"
```

---

## Task 16: Update `docs/nodes.md`

**Files:**
- Modify: `docs/nodes.md`

- [ ] **Step 16.1: Update node-kind count**

Line 9: `There are 6 node kinds` — count the kinds currently documented. With conditional (already in grammar but missing from this doc per exploration) plus manager_loop, verify the correct final number. Read the file to confirm present state before editing.

If conditional is missing, also add it. If this doc already has conditional and just needs manager_loop, change "6" to "7" or "8" as appropriate.

- [ ] **Step 16.2: Extend the mermaid diagram**

Around lines 11-25, add manager_loop to Composition Nodes:

```text
    subgraph Composition Nodes
        subgraph_node["subgraph<br>Sub-pipeline"]
        manager_loop["manager_loop<br>Supervisor"]
    end
```

- [ ] **Step 16.3: Extend the kind table**

Around lines 27-34, add a row:

```text
| `manager_loop` | Supervise a child subgraph with polling and steering | Block with subgraph_ref |
```

- [ ] **Step 16.4: Append a "Manager Loop Nodes" section**

Add after the "Subgraph Nodes" section (around line 420):

````markdown
---

## Manager Loop Nodes

Manager loop nodes supervise a child sub-pipeline, polling it on a cadence and optionally steering it by injecting context during execution. They map to the runtime's `stack.manager_loop` and DOT's `shape=house`.

```dippin
  manager_loop QualityGate
    label: "Quality Gate Supervisor"
    subgraph_ref: quality_loop.dip
    poll_interval: 30s
    max_cycles: 12
    stop_condition: stack.child.outcome = success
    steer_condition: stack.child.cycles = 5
    steer_context: hint=halfway_through, priority=high
```

### Manager-Loop-Specific Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `subgraph_ref` | String | **Yes** | Path to the `.dip` file of the child pipeline to supervise. Lints DIP135 if missing or file does not exist. |
| `poll_interval` | Duration | No | How often the supervisor wakes to check child state (e.g., `30s`, `5m`). `0` means event-driven (no polling). |
| `max_cycles` | Integer | No | Hard cap on child cycles. `0` means unbounded — unless `stop_condition` is set, this triggers DIP137. |
| `stop_condition` | Condition | No | Terminates supervision when true. Uses `stack.child.*` runtime variables. |
| `steer_condition` | Condition | No | Injects `steer_context` into the child when true. |
| `steer_context` | Map | No | Key-value hints. Inline form (`key=val,key=val`) or block form (one `key: val` per line). |

### Supervisor State

While supervision is running, the runtime exposes the child's state under the `stack.child.*` namespace:

- `stack.child.cycles` — how many iterations the child has run
- `stack.child.outcome` — the child's last reported outcome
- `stack.child.status` — `running`, `stopped`, `failed`

Use these in `stop_condition` and `steer_condition` expressions.

### Lint Checks

- **DIP135**: `subgraph_ref` missing or points to a nonexistent file
- **DIP136**: invalid control field (negative `poll_interval` or `max_cycles`)
- **DIP137**: unbounded supervisor (no `stop_condition` and no `max_cycles`)

See `examples/manager_loop_demo.dip` for a complete working example.
````

- [ ] **Step 16.5: Commit**

```bash
git add docs/nodes.md
git commit -m "docs: add Manager Loop Nodes section

Documents subgraph_ref, poll_interval, max_cycles, stop_condition,
steer_condition, steer_context, and stack.child.* runtime variables.
Extends node-kind count, kind table, and mermaid composition diagram."
```

---

## Task 17: Update README, site language reference, and skill file

**Files:**
- Modify: `README.md`
- Modify: `site/content/language.md`
- Modify: `site/static/skill.md`

- [ ] **Step 17.1: README.md**

Find the Node Types section (the exploration located it around line 197; verify). Add a short manager_loop entry matching the existing style. Example:

```markdown
### manager_loop — supervise a child pipeline

```dippin
  manager_loop Supervisor
    subgraph_ref: child.dip
    poll_interval: 30s
    max_cycles: 12
    stop_condition: stack.child.outcome = success
```

DOT shape: `house`. See [docs/nodes.md#manager-loop-nodes](docs/nodes.md#manager-loop-nodes).
```markdown

- [ ] **Step 17.2: site/content/language.md**

Mirror the `docs/nodes.md` change — same kind count, same manager_loop entry in the kind grid.

- [ ] **Step 17.3: site/static/skill.md**

Find the subgraph section (Step 1's exploration showed it around line 120-145) and add a `manager_loop` section after it:

````markdown
### manager_loop — supervise a child pipeline

```
  manager_loop Supervisor
    subgraph_ref: child.dip
    poll_interval: 30s
    max_cycles: 12
    stop_condition: stack.child.outcome = success
    steer_condition: stack.child.cycles = 5
    steer_context: hint=halfway_through
```markdown

| Field | Type | Notes |
|-------|------|-------|
| `subgraph_ref` | string | Path to the child `.dip` file (DIP135 if missing or nonexistent) |
| `poll_interval` | duration | `30s`, `5m`, etc. `0` = event-driven |
| `max_cycles` | int | Hard cap; `0` = unbounded (DIP137 if no stop_condition either) |
| `stop_condition` | expression | Terminates when true; uses `stack.child.*` |
| `steer_condition` | expression | Injects `steer_context` when true |
| `steer_context` | map | `key=val,key=val` inline, or block form |

Maps to DOT `shape=house`. Runtime exposes `stack.child.cycles`, `stack.child.outcome`, `stack.child.status`.
````

- [ ] **Step 17.4: Commit**

```bash
git add README.md site/content/language.md site/static/skill.md
git commit -m "docs: document manager_loop in README, site language ref, and skill"
```

---

## Task 18: Update CHANGELOG

**Files:**
- Modify: `CHANGELOG.md`

- [ ] **Step 18.1: Prepend a new v0.22.0 block**

Add above the v0.21.0 entry:

```markdown
## [v0.22.0] — YYYY-MM-DD

### Added
- **`manager_loop` node kind** for supervising a child sub-pipeline with polling and mid-run context steering. Maps to the runtime's `stack.manager_loop` and DOT `shape=house`. Fields: `subgraph_ref`, `poll_interval`, `max_cycles`, `stop_condition`, `steer_condition`, `steer_context` (inline CSV or block form). Round-trips through parser → formatter → DOT export → migrate. Requires a coordinated runtime release. ([#26](https://github.com/2389-research/dippin-lang/issues/26))
- **DIP135-137** lint codes for manager_loop validation: missing/nonexistent `subgraph_ref` (DIP135), invalid control field (DIP136), unbounded supervision (DIP137 — the manager_loop analog of DIP104).
- **`stack.*` namespace** recognized by DIP120; `stop_condition`/`steer_condition` can reference `stack.child.cycles`, `stack.child.outcome`, etc., without namespace warnings.
- **`dippin scaffold manager_loop`** template emits a starter supervisor workflow.
- **Tree-sitter grammar**: `manager_loop` node rule, committed generated parser (`src/parser.c` et al.), new `just tree-sitter-generate` / `just tree-sitter-test` recipes, and CI drift check.

### Fixed
- **VS Code TextMate grammar**: `parallel` and `fan_in` are now highlighted in their plain block-form declaration regex (previously only the arrow forms were covered). Bundled with the `manager_loop` grammar update.
```

Set the release date when tagging (don't date-stamp prematurely).

- [ ] **Step 18.2: Regenerate site changelog**

```shell
just changelog-md
```

This overwrites `site/content/changelog.md` from `CHANGELOG.md`. Commit the regenerated file together.

- [ ] **Step 18.3: Commit**

```bash
git add CHANGELOG.md site/content/changelog.md
git commit -m "docs: changelog for manager_loop feature (unreleased)"
```

---

## Task 19: File the runtime follow-up issue

**Pre-merge task — do this BEFORE the dippin-lang PR opens.**

- [ ] **Step 19.1: Check for cross-repo coordination style**

Review past CHANGELOG entries for cross-repo coordination style.

- [ ] **Step 19.2: Draft the runtime adapter issue body**

Title: `feat: add ir.NodeManagerLoop support to the runtime's pipeline adapter layer`

Body:
```text
dippin-lang v0.22.0 (see 2389-research/dippin-lang#26) adds a new IR node
kind `NodeManagerLoop` that expresses `stack.manager_loop` supervisors.
The runtime's DOT parser already handles the node via `shape=house`, but the
dippin → runtime adapter does not yet map `ir.NodeManagerLoop` → flat
adapter attrs.

**Until this lands**, pin the runtime to `dippin-lang@v0.21.0`. Bumping to
v0.22.0 without this adapter update will cause `stack.manager_loop` nodes
authored in `.dip` to silently lose their config on the runtime side.

**What needs to happen in this repo**:

1. The pipeline adapter's `nodeKindToShapeMap`: add
   `ir.NodeManagerLoop: "house"`.
2. In the flat-attr extractor for manager nodes, read these fields from
   `ir.ManagerLoopConfig` and write them as DOT-style flat attrs that
   the runtime's pipeline handler for manager_loop already consumes: `subgraph_ref`,
   `poll_interval` (as string duration), `max_cycles` (as stringified int),
   `stop_condition`, `steer_condition`, `steer_context` (flattened to
   canonical sorted `k=v,k=v`).
3. Add a round-trip test: `.dip` with manager_loop → adapter → handler.

Reference (dippin side): `export/dot.go applyManagerLoopAttrs` has the
same flattening logic; mirror it.
```

- [ ] **Step 19.3: File the issue in the runtime repo**

File the issue in the runtime repository. Copy the issue number and update the dippin PR body to reference it and update the CHANGELOG entry with the real number.

---

## Task 20: Final verification and PR prep

- [ ] **Step 20.1: Run the full check**

```shell
just check
```

Expected: "All checks passed." If any step fails:
- **complexity violations**: extract helpers; never add `//nolint`.
- **validate-examples**: the new `.dip` file must parse and lint clean (zero errors, DIP108 zero).
- **fmt-check**: run `just fmt` to auto-fix.
- **test-race**: fix data-race issues (unlikely for this change).

- [ ] **Step 20.2: Run tree-sitter check**

```shell
just tree-sitter-test
```

Expected: all corpus cases pass.

- [ ] **Step 20.3: Manual round-trip smoke test**

```shell
./dippin fmt examples/manager_loop_demo.dip > /tmp/m1.dip
./dippin fmt /tmp/m1.dip > /tmp/m2.dip
diff /tmp/m1.dip /tmp/m2.dip
```

Expected: no diff (idempotent).

```shell
./dippin export-dot examples/manager_loop_demo.dip > /tmp/m.dot
grep 'shape=house' /tmp/m.dot
grep 'subgraph_ref' /tmp/m.dot
grep 'steer_context' /tmp/m.dot
```

Expected: all three greps succeed.

```shell
./dippin migrate /tmp/m.dot > /tmp/back.dip
./dippin validate /tmp/back.dip
```

Expected: `/tmp/back.dip` is a valid `.dip` with the manager_loop node preserved.

- [ ] **Step 20.4: Self-review the diff**

```shell
git log --oneline main..HEAD
git diff main..HEAD --stat
```

Check:
- No unintended changes to unrelated files.
- No `//nolint` directives added.
- No hand-edits to `editors/tree-sitter-dippin/src/*` (they should be regenerated).
- CHANGELOG has the release date placeholder `YYYY-MM-DD` (don't date-stamp yet).

- [ ] **Step 20.5: Push and open PR**

```shell
git push -u origin feat/manager-loop-node
gh pr create --title "feat: add manager_loop node kind (issue #26)" --body "$(cat <<'EOF'
## Summary

Adds a new `manager_loop` node kind to the dippin-lang IR so `.dip` files can natively express the runtime's `stack.manager_loop` supervisors — a node that runs a child sub-pipeline, polls on an interval, and can steer the running child via context injection. Closes #26.

Without this change, pipelines using manager_loop must be authored in DOT format end-to-end.

## What's in this PR

- **IR**: `NodeManagerLoop` kind + `ManagerLoopConfig` struct (fields: `SubgraphRef`, `PollInterval`, `MaxCycles`, `StopCondition`, `SteerCondition`, `SteerContext`)
- **Parser**: `manager_loop` keyword dispatch; `steer_context` parses both inline CSV and block form
- **Conditions**: `stop_condition`/`steer_condition` reuse `ir.Condition{Raw, Parsed}`; `simulate.EnsureConditionsParsed` extended to walk manager_loop nodes so lint rules work
- **Formatter**: canonical output with idempotent round-trip
- **DOT export**: `shape=house` + lossless flat-attr round-trip (steer_context flattens to sorted `k=v,k=v`)
- **DOT migrate**: reverse mapping `shape=house` → `NodeManagerLoop`
- **Linter**: DIP135 (missing/nonexistent ref), DIP136 (invalid control field), DIP137 (unbounded supervisor — DIP104 analog); extends DIP120 namespaces with `stack.*`
- **Scaffold**: `dippin scaffold manager_loop` template
- **LSP**: symbol mapping for manager_loop nodes
- **Tree-sitter**: grammar.js rule + highlights + corpus test + committed generated parser + new `just tree-sitter-generate` / `just tree-sitter-test` recipes + CI drift check
- **VS Code**: grammar updated for `manager_loop` keyword, new fields, `stack.*` namespace
- **Docs**: `docs/nodes.md` full section, README node-type entry, EBNF grammar rule, published site language ref, hosted `skill.md` section
- **Example**: `examples/manager_loop_demo.dip` (and `examples/child_pipeline.dip` for the ref)
- **CHANGELOG**: v0.22.0 Added section

Also folds in a pre-existing bug fix: the VS Code TextMate node-declaration regex was missing `parallel` and `fan_in` — noted under "Fixed" in the CHANGELOG.

## Cross-repo coordination

**Runtime side** — the runtime's pipeline adapter layer needs a parallel update to map `ir.NodeManagerLoop` → `shape=house` with flat attrs. **Until that lands, downstream consumers should pin to `dippin-lang@v0.21.0`.** Bumping to v0.22.0 without the adapter change will silently drop manager_loop config on the runtime side.

## Post-merge TODO

- Bump `editors/zed-dippin/extension.toml` `commit = "..."` SHA to point at the merge commit so Zed picks up the new grammar.

## Test plan

- [x] `just check` passes (build, vet, fmt, test-race, releasecheck, complexity ≤5/≤7, validate-examples)
- [x] `just tree-sitter-test` passes
- [x] Round-trip: `.dip` → `fmt` → `fmt` is idempotent
- [x] Round-trip: `.dip` → `export-dot` → `migrate` → validate preserves all manager_loop config
- [x] `dippin scaffold manager_loop` produces a valid workflow
- [x] `dippin lint examples/manager_loop_demo.dip` is clean

🤖 Generated with [Claude Code](https://claude.com/claude-code)
EOF
)"
```

- [ ] **Step 20.6: Confirm post-merge steps are tracked**

Create a reminder (issue, TODO, or calendar note):

1. Update `editors/zed-dippin/extension.toml` `commit = ...` to the merge commit SHA.
2. Tag `v0.22.0` when ready; GoReleaser handles the rest.
3. Coordinate with downstream consumers to land the matching adapter update + bump their `dippin-lang` pin.

---

## Self-review notes

- **Spec coverage**: every field from issue #26 (`subgraph_ref`, `poll_interval`, `max_cycles`, `stop_condition`, `steer_condition`, `steer_context`) has a parser handler, a formatter emitter, a DOT export attr, a migrate builder, and test coverage. The 5 "What needs to happen" items in the issue map to Tasks 1, 2, 8, 5, and 19 respectively.
- **Cross-cutting invariant**: any new code reading `Condition.Parsed` goes through `simulate.EnsureConditionsParsed` first — Task 3 extends that function to cover manager_loop, and Task 8's linter uses only `Raw` plus `Parsed` via the extended walk.
- **Complexity budget**: each new function has ≤5 branches. Functions with more (e.g., `buildManagerLoopConfig`) are already pre-split into scalar/condition helpers.
- **DIP code permanence**: DIP135/136/137 are contiguous with DIP134 (the last allocated) and follow the existing pattern (DIP116/DIP130 bundle format checks; DIP104 is the DIP137 precedent).
- **No placeholders**: every step has exact paths, code, and commands. Expected test output stated where relevant.
