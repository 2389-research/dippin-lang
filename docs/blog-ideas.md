# Blog Post Ideas

25 planned posts covering every discrete feature of the Dippin toolchain. Each post should be self-contained, practical, and include runnable examples.

## Status

### Published (8)
- [x] 1. Getting Started with Dippin — `site/blog/getting-started.html`
- [x] 2. Scenario Testing with .test.json — `site/blog/scenario-testing.html`
- [x] 3. Migrating from DOT to Dippin — `site/blog/migrating-from-dot.html`
- [x] 4. CI Integration: Lint, Test, Format — `site/blog/ci-integration.html`
- [x] 5. Editor Setup: LSP, VS Code, Tree-sitter — `site/blog/editor-setup.html`

### Remaining (17)
- [x] 6. Conditional Edges: Routing with `when` — `site/blog/conditional-edges.html`
- [ ] 7. Parallel Execution: Fan-Out and Fan-In
- [ ] 8. Human Gates: Choice vs Freeform
- [ ] 9. Goal Gates: Pipeline-Critical Quality Checks
- [ ] 10. Retry Policies and Fallback Targets
- [ ] 11. Edge Coverage: Are Your Tests Complete?
- [x] 12. Cost Estimation: Know Before You Run — `site/blog/cost-estimation.html`
- [ ] 13. The Doctor Will See You Now: Health Reports
- [ ] 14. Dead Branch Detection with `dippin unused`
- [ ] 15. `dippin watch`: Live Feedback While You Author
- [ ] 16. ASCII DAGs with `dippin graph`
- [x] 17. Multi-line Prompts Without Escaping — `site/blog/multi-line-prompts.html`
- [ ] 18. Shell Scripts in Tool Nodes
- [ ] 19. Context Variables and Data Flow
- [ ] 20. Subgraph Composition
- [ ] 21. Semantic Diff: Comparing Workflow Versions
- [ ] 22. The WASM Playground: Dippin in Your Browser
- [ ] 23. The Formatter: Canonical Style and Idempotency
- [ ] 24. The 35 Diagnostics: What They Catch and Why
- [ ] 25. Model Catalog: Supported Providers and Pricing

---

## Detailed Outlines

### Guides

### 1. Getting Started with Dippin
**Synopsis**: Zero-to-running in 5 minutes. The reader leaves with a working installation and a validated workflow.
**Cover**:
- Install via `go install` and Homebrew
- Create a minimal 2-node workflow from scratch (not from a template)
- `dippin validate` and `dippin lint` — what the output means
- `dippin export-dot | dot -Tpng` for visualization
- `dippin new` templates for bootstrapping bigger workflows
- Link to the playground for instant experimentation
**Approach**: Hands-on walkthrough with copy-pasteable commands. Show terminal output at each step. End with "where to go next" links to other posts.

### 2. Scenario Testing with .test.json
**Synopsis**: How to write deterministic tests for non-deterministic AI pipelines by injecting context values and asserting on execution paths.
**Cover**:
- The `.test.json` file format and auto-discovery (`pipeline.dip` → `pipeline.test.json`)
- Scenario injection: `"scenario": {"outcome": "fail"}` and per-node `"Node.key": "val"`
- All 5 expectation fields: `status`, `visited`, `not_visited`, `path_contains`, `immediately_after`
- The `branch` field for targeted parallel testing
- `--verbose` for path tracing, `--coverage` for edge coverage
- JSON output for CI integration
- Common pitfall: `not_visited` fragility with loop-breaking
**Approach**: Build a conditional workflow, write 3 test cases that exercise success/failure/retry paths. Show output at each step.

### 3. Migrating from DOT to Dippin
**Synopsis**: Step-by-step migration of an existing Graphviz DOT pipeline file to Dippin with structural parity verification.
**Cover**:
- `dippin migrate --output pipeline.dip pipeline.dot` — what the converter does
- Shape-to-kind mapping (box→agent, hexagon→human, etc.)
- How prompts/commands are unescaped from DOT string encoding
- `dippin validate-migration pipeline.dot pipeline.dip` — parity checking
- Common parity failures: model name differences, condition prefix (`context.` → `ctx.`)
- Manual cleanup: adding missing fields, fixing formatting
- Side-by-side comparison of DOT vs Dippin for the same workflow
**Approach**: Use a real example from the `examples/` directory that has both `.dot` and `.dip` files. Show the migration end-to-end.

### 4. CI Integration: Lint, Test, Format in Your Pipeline
**Synopsis**: Set up automated workflow validation in GitHub Actions (or any CI). The reader leaves with a working CI config.
**Cover**:
- Exit codes: 0=ok, 1=errors, 2=usage — which commands return what
- `dippin lint` in CI: warnings don't fail, errors do
- `dippin fmt --check` for formatting enforcement
- `dippin test` for scenario regression tests with JSON output
- `dippin check --format json` for LLM tool-calling loops
- The dogfooding CI job: how dippin-lang tests itself
- Sample GitHub Actions workflow YAML
**Approach**: Build a `.github/workflows/dippin.yml` from scratch, explain each step, show passing/failing output.

### 5. Editor Setup: LSP, VS Code, and Tree-sitter
**Synopsis**: Get real-time diagnostics, hover docs, and syntax highlighting in your editor.
**Cover**:
- VS Code extension: install from `editors/vscode/`, features (highlighting, comment toggling, folding)
- LSP server: `dippin lsp` on stdio — what capabilities it provides
- Neovim with nvim-lspconfig: minimal config to connect
- Neovim with tree-sitter: the `editors/tree-sitter-dippin/` grammar
- Helix/Zed: tree-sitter grammar registration
- What the LSP provides: diagnostics, hover (node kind/model/prompt preview), go-to-definition, autocomplete, document symbols
- `dippin watch` as a lightweight alternative when LSP isn't available
**Approach**: Editor-specific setup sections with exact config snippets. Screenshots or terminal recordings of each feature.

---

## Tutorials

### 6. Conditional Edges: Routing Pipelines with `when`
**Synopsis**: Build branching workflows where the execution path depends on LLM output or tool results.
**Cover**:
- Basic conditions: `when ctx.outcome = success`
- All operators: `=`, `==`, `!=`, `contains`, `not contains`, `startswith`, `endswith`, `in`
- Compound conditions: `and`, `or`, `not`, parentheses for precedence
- `auto_status: true` for self-assessment routing
- Exhaustive detection: why `success`/`fail` pairs suppress DIP101/DIP102
- Complementary pairs: `contains X` / `not contains X`
- Common mistake: forgetting the `ctx.` namespace prefix (DIP120)
**Approach**: Build a review-loop workflow step by step, adding conditions progressively. Show lint output at each stage.

### 7. Parallel Execution: Fan-Out and Fan-In
**Synopsis**: Run multiple LLM calls concurrently and merge the results.
**Cover**:
- `parallel FanOut -> A, B, C` and `fan_in Join <- A, B, C` syntax
- How the simulator walks all branches
- Per-branch model overrides for multi-provider consensus
- The DIP007 parity check: targets must match sources
- Edges: you need explicit `FanOut -> A`, `A -> Join` edges
- Pattern: 3-model consensus with cross-review (like `consensus_task.dip`)
- Testing parallel workflows with the `branch` field
**Approach**: Build a 3-model consensus workflow from scratch. Test it with `dippin test`.

### 8. Human Gates: Choice vs Freeform
**Synopsis**: Pause pipeline execution for human input, then route based on the response.
**Cover**:
- `mode: choice` — buttons from edge labels, `default` for auto-select
- `mode: freeform` — open text, stored in `ctx.human_response`
- `preferred_label` for scenario-driven testing of choice gates
- Routing: edge labels match human selection
- Testing human nodes: scenario injection bypasses the prompt
- The `human_gate_showcase.dip` example
**Approach**: Build a human approval gate, test both paths, show how `preferred_label` works.

### 9. Goal Gates: Pipeline-Critical Quality Checks
**Synopsis**: Mark nodes that MUST pass for the pipeline to succeed, even if execution reaches exit.
**Cover**:
- `goal_gate: true` — what it means semantically
- Pairing with `auto_status: true` for automatic pass/fail detection
- `retry_target` and `fallback_target` for recovery (DIP115 warns if missing)
- How goal gates appear in DOT export (red filled background)
- Testing goal gate failures
**Approach**: Add a security review gate to a pipeline, show how it fails the pipeline even when exit is reached.

### 10. Retry Policies and Fallback Targets
**Synopsis**: Build resilient pipelines that recover from failures instead of crashing.
**Cover**:
- Retry policies: `standard`, `aggressive`, `patient`, `linear`, `none`
- `max_retries` and `base_delay` overrides
- `retry_target`: jump to a different node on retry
- `fallback_target`: last resort when all retries exhausted
- Restart edges: `restart: true` with `max_restarts` control
- DIP104: unbounded retry warning, DIP115: goal gate without recovery
- Pattern: implement → review → retry loop with bounded restarts
**Approach**: Build a review-loop with escalating retry behavior. Show simulation with `--scenario` to explore paths.

### 11. Edge Coverage: Are Your Tests Complete?
**Synopsis**: Use `dippin test --coverage` to find blind spots in your test suite.
**Cover**:
- What edge coverage measures: traversed edges vs total edges
- Reading the output: `14/16 edges covered (87.5%)` + uncovered list
- Why parallel fan-out edges may show as uncovered (implicit traversal)
- Strategy: write test cases that target uncovered edges
- Pairing with `dippin coverage` for static reachability analysis
**Approach**: Start with a partial test suite, run `--coverage`, write missing cases, achieve 100%.

### 12. Cost Estimation: Know Before You Run
**Synopsis**: Estimate per-run pipeline cost before spending real money on LLM calls.
**Cover**:
- `dippin cost` — per-node and total cost breakdown
- How token estimation works (prompt length → input tokens, heuristic for output)
- `max_turns` multiplier for agentic loops
- Tool and human nodes cost $0
- `dippin optimize` — cheaper model substitution suggestions
- `dippin feedback` — calibrating estimates with real telemetry
- The pricing table: 7 providers, verification policy
**Approach**: Cost a real workflow, show the breakdown, run optimize, compare suggestions.

### 13. The Doctor Will See You Now: Health Reports
**Synopsis**: Get a letter grade and actionable suggestions for any pipeline.
**Cover**:
- `dippin doctor` — the grade scale (A–F) and what each means
- Scoring components: lint issues, edge coverage, cost efficiency, reachability
- How to go from C to A: fix warnings, add tests, add timeouts
- JSON output for dashboards
- Pairing with `dippin optimize` for the cost component
**Approach**: Doctor a messy workflow (intentionally flawed), fix issues one by one, show grade improving.

### 14. Dead Branch Detection with `dippin unused`
**Synopsis**: Find nodes you're paying for that can never contribute to pipeline success.
**Cover**:
- What "unused" means: reachable from start but no path to exit
- How it differs from DIP004 (unreachable = no path FROM start)
- Cost estimation for wasted LLM calls
- Common causes: leftover nodes from refactoring, misrouted edges
- JSON output for automated cleanup
**Approach**: Show a workflow with dead branches, run `unused`, fix the routing.

### 15. `dippin watch`: Live Feedback While You Author
**Synopsis**: Get instant lint results on every save without leaving your editor.
**Cover**:
- Basic usage: `dippin watch examples/` or `dippin watch pipeline.dip`
- How debouncing works (200ms — avoids duplicate runs on multi-write saves)
- Pairing with split-pane terminal: editor on left, watch on right
- When to use watch vs LSP
- Directory watching: monitors all `.dip` files in the directory
**Approach**: Show a terminal recording of editing a file and seeing instant feedback.

### 16. ASCII DAGs with `dippin graph`
**Synopsis**: Visualize your workflow topology right in the terminal.
**Cover**:
- Full mode: box-drawing characters, topological layers
- Compact mode: `[Start] → [Middle] → [Exit]` one-liner
- How parallel branches are rendered
- JSON output for programmatic use
- When to use `graph` vs `export-dot`
**Approach**: Show the same workflow in full, compact, and DOT-rendered modes.

---

## Deep Dives

### 17. Multi-line Prompts Without Escaping
**Synopsis**: The #1 reason Dippin exists. How indentation-based blocks solve DOT's escaping nightmare.
**Cover**:
- The DOT problem: `prompt="line1\nline2\n\"quoted\""` — real examples from production pipelines
- The Dippin solution: indent after `:`, write anything
- What's preserved: blank lines, markdown, code blocks, shell scripts, JSON
- How the lexer handles indentation: indent stack, base-level stripping
- Edge cases: Dippin syntax inside prompts (it's all literal), `#` comments are literal too
- Formatter round-tripping: format preserves multiline content exactly
**Approach**: Show the same prompt in DOT and Dippin side by side. Then show increasingly complex content (markdown with code blocks, JSON, here-docs).

### 18. Shell Scripts in Tool Nodes
**Synopsis**: Run real shell commands with full syntax support — no escaping, no limitations.
**Cover**:
- The `command:` multiline block — any shell syntax works
- `timeout` — why DIP111 warns without one
- `outputs` — declared stdout values for coverage analysis
- `ctx.tool_stdout` and `ctx.tool_stderr` — accessing command results
- DIP123: shell syntax checking via `bash -n`
- DIP124: why `${ctx.*}` in commands doesn't work (runtime vs parse time)
- DIP125: binary existence checking on PATH
- Stress test: the `stress_shell_hell.dip` example (here-docs, case/esac, traps, arrays)
**Approach**: Build a tool node with a non-trivial script, show the lint checks catching real issues.

### 19. Context Variables and Data Flow
**Synopsis**: How data flows between pipeline nodes through the context system.
**Cover**:
- Three namespaces: `ctx.*` (runtime), `graph.*` (workflow-level), `params.*` (subgraph parameters)
- `reads` and `writes` declarations — advisory metadata for linting
- `${ctx.key}` interpolation in prompts
- Reserved runtime variables: `ctx.outcome`, `ctx.status`, `ctx.tool_stdout`, `ctx.human_response`
- DIP106: undefined variable in prompt
- DIP107: unused writes (written but never read)
- DIP112: reads without upstream writes
- DIP120: missing namespace prefix in conditions
- DIP121/DIP122: condition variable/value not declared in source node
**Approach**: Build a data pipeline where nodes pass information. Show how lint catches broken data flow.

### 20. Subgraph Composition
**Synopsis**: Embed reusable workflows inside other workflows.
**Cover**:
- `subgraph` node syntax: `ref` and `params`
- Parameter passing via `params.*` namespace
- DIP126: ref file validation
- When to decompose: signs a workflow should be split
- Namespace isolation between parent and child workflows
- DIP109: namespace collision detection
**Approach**: Extract a review sub-pipeline from a larger workflow, parameterize it, embed it in two different parent workflows.

### 21. Semantic Diff: Comparing Workflow Versions
**Synopsis**: Go beyond text diffs to understand what actually changed between workflow versions.
**Cover**:
- `dippin diff old.dip new.dip` — what it compares
- Diff categories: added/removed nodes, changed edges, modified conditions, updated config
- Field-level precision: "node X prompt changed" vs "node X model changed from A to B"
- JSON output for automated change tracking
- Use case: PR reviews for pipeline changes
**Approach**: Make incremental changes to a workflow, run diff at each step, show how it captures the semantic meaning.

### 22. The WASM Playground: Dippin in Your Browser
**Synopsis**: How the entire Go toolchain compiles to WebAssembly and runs client-side.
**Cover**:
- The build: `GOOS=js GOARCH=wasm go build`
- Build tag split: `//go:build !wasm` for platform-dependent code (exec, os.Stat)
- What works in WASM: parse, validate, lint (most rules), format
- What doesn't: DIP123 (bash -n), DIP125 (exec.LookPath), DIP126 (os.Stat)
- The JS bridge: `wasm_exec.js`, `syscall/js` API
- The transparent textarea trick for syntax-highlighted editing
- Binary size considerations (4MB)
**Approach**: Walk through the architecture, show the build process, explain each design decision.

### 23. The Formatter: Canonical Style and Idempotency
**Synopsis**: How `dippin fmt` produces deterministic output and why idempotency matters.
**Cover**:
- The idempotency guarantee: `format(format(x)) == format(x)` — tested in CI
- Canonical field ordering per node type (label, class, model, ... prompt last)
- Multiline block handling: preserving content while normalizing indentation
- `--check` for CI: exit 1 if not formatted
- `--write` for in-place formatting
- How the formatter is implemented: IR → canonical text (not text → text)
**Approach**: Show a messy workflow, format it, format the output again, prove they match.

---

## Reference

### 24. The 35 Diagnostics: What They Catch and Why
**Synopsis**: A complete walkthrough of every diagnostic code with examples and fix guidance.
**Cover**:
- Design philosophy: errors (DIP001–009) vs warnings (DIP101–126) vs hints
- Each code: what triggers it, example code that triggers it, how to fix it
- Suppression rules: when DIP101/DIP102 auto-suppress (exhaustive conditions)
- Advisory checks: DIP121/DIP122 only fire when writes/outputs are declared
- Environment-dependent: DIP125 (PATH), DIP126 (file existence)
- `dippin explain DIPxxx` for quick reference
**Approach**: Group by theme (structure, reachability, data flow, config, tool commands). Each code gets a before/after example.

### 25. Model Catalog: Supported Providers and Pricing
**Synopsis**: Every supported model across all 7 providers, with pricing and verification policy.
**Cover**:
- Providers: Anthropic, OpenAI, Google/Gemini, DeepSeek, xAI/Grok, Mistral, Cohere
- Model IDs: exact strings to use in `model:` fields
- Pricing: per-1M-token input/output rates
- The verification policy: never use training data, check official docs, maintain "Last verified" dates
- How DIP108 catches unknown model/provider combos
- Adding new models: where to edit (`lint_model.go`, `pricing.go`), what to verify
- `TestLintExamples` as a staleness check
**Approach**: Tables organized by provider, with links to official pricing pages.
