# Issue #65 — `prompt_file:` / `system_prompt_file:` Implementation Plan (v0.34.0)

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Ship `prompt_file:` and `system_prompt_file:` directives on agent nodes — symmetric extension of v0.33.0's `command_file:` — and fix the latent bug where `formatter` silently drops inline `system_prompt:` on round-trip.

**Architecture:** Reuses every mechanism introduced by #52. `AgentConfig` gains two new `*File` fields. The existing `parser.ResolveFileDirectives` dispatcher grows an agent branch alongside its tool branch, calling the unchanged `loadDirectiveFile` (same 4-MiB cap, same 4 security checks, same user-path-only error messages). The pack-shadow tree's clear-directives helper is renamed and extended to walk agent nodes too. Formatter gains two conditional `*File:` / inline-content branches, replacing the existing single `Prompt`-only trailing block. No new architecture, no new abstractions. Dippin-only release — no runtime coordination required.

**Tech Stack:** Go, gofmt, golangci-lint, gocyclo (≤5), gocognit (≤7), just recipes.

**Spec:** [`docs/superpowers/specs/2026-05-27-issue-65-prompt-file-design.md`](../specs/2026-05-27-issue-65-prompt-file-design.md)

---

## Task 0: File 2 follow-up issues + add issue numbers to spec

**Purpose:** Lock the deferred non-goals into tracked work before merge, matching the #52 / #41 pattern. The spec body currently references the follow-ups as `[follow-up filed before merge]` placeholders; this task replaces those with real issue numbers.

**Files:**
- Modify: `docs/superpowers/specs/2026-05-27-issue-65-prompt-file-design.md` (replace the two `[follow-up filed before merge]` placeholders)

- [ ] **Step 1: File follow-up issue — Defaults-block support for `prompt_file:` / `system_prompt_file:`**

Run:
```bash
gh issue create --repo 2389-research/dippin-lang \
  --title "follow-up: defaults agent prompt_file/system_prompt_file support (deferred from #65)" \
  --label enhancement \
  --body "$(cat <<'EOF'
Filed per v0.34.0 spec § Non-goals #1.
See: https://github.com/2389-research/dippin-lang/blob/main/docs/superpowers/specs/2026-05-27-issue-65-prompt-file-design.md

`defaults agent` currently does not accept `prompt` or `system_prompt` at all (see `parser/parse_defaults.go` recognized-keys list). v0.34's `prompt_file:` / `system_prompt_file:` directives are per-node only, matching what `command_file:` shipped with in v0.33. Adding file-form support to `defaults agent` would require shipping inline-form support at the same time — that's a meaningfully larger design (cascade semantics, override rules, per-node-vs-default mutual exclusion).

Ship when an author wants to share one large external persona doc across agents in a workflow.
EOF
)"
```

Record the returned issue number as `DEFAULTS_ISSUE`.

- [ ] **Step 2: File follow-up issue — Bundled-files `.dipx` redesign**

Run:
```bash
gh issue create --repo 2389-research/dippin-lang \
  --title "follow-up: bundle external files as separate entries in .dipx (deferred from #65)" \
  --label enhancement \
  --body "$(cat <<'EOF'
Filed per v0.34.0 spec § Non-goals #2.
See: https://github.com/2389-research/dippin-lang/blob/main/docs/superpowers/specs/2026-05-27-issue-65-prompt-file-design.md

Currently \`dippin pack\` inlines file contents from \`command_file:\` / \`prompt_file:\` / \`system_prompt_file:\` directives into the bundled \`.dip\` text via the shadow-tree machinery (\`cmd/dippin/pack_shadow.go\`). An alternative model would package the referenced scripts/prompts as separate files inside the \`.dipx\`, preserving their relative paths, and have the runtime (or any other dipx consumer) call \`parser.ResolveFileDirectives\` against the unpacked bundle dir.

**Wins:**
- Runtime logs / error messages can reference original file paths (\`scripts/setup.sh:15\`) instead of inline-heredoc line numbers
- \`.dipx\` debug inspection shows author intent — file boundaries preserved
- Author's repo layout reflected end-to-end through the bundle

**Costs:**
- Runtime coordination required (coordinated runtime release, dipx version bump in the runtime's go.mod, new \`ResolveFileDirectives\` call after unpack)
- \`.dipx\` format version bumps — old bundles built by v0.33/v0.34 have inlined content; new format would have separate files. Need a cutover plan
- The shadow-tree machinery (\`cmd/dippin/pack_shadow.go\`) becomes obsolete and would be removed
- CLAUDE.md allows dipx to import \`parser\` (loader-tier exemption), so the architecture rule still holds — we just didn't lean on it in #52

Considered explicitly during #65 brainstorming and deferred so v0.34 stays dippin-only.
EOF
)"
```

Record the returned issue number as `BUNDLED_ISSUE`.

- [ ] **Step 3: Replace placeholders in spec with real issue numbers**

Open `docs/superpowers/specs/2026-05-27-issue-65-prompt-file-design.md`. Find each `[follow-up filed before merge]` placeholder (there are two — one in § Non-goals #1 and one in § Non-goals #2) and replace with `[#<number>](https://github.com/2389-research/dippin-lang/issues/<number>)`. Use the issue numbers from steps 1 and 2.

The § Follow-up issues section at the bottom of the spec lists the two items without explicit issue links — append `([#<DEFAULTS_ISSUE>](https://github.com/2389-research/dippin-lang/issues/<DEFAULTS_ISSUE>))` and `([#<BUNDLED_ISSUE>](https://github.com/2389-research/dippin-lang/issues/<BUNDLED_ISSUE>))` to the end of each item respectively.

- [ ] **Step 4: Commit**

```bash
git add docs/superpowers/specs/2026-05-27-issue-65-prompt-file-design.md
git commit -m "docs(spec): backlink follow-up issue numbers for #65 spec"
```

---

## Task 1: Add `PromptFile` and `SystemPromptFile` fields to `ir.AgentConfig`

**Purpose:** Two new string fields on `AgentConfig`, each clustered next to its content twin. Same primary-representation contract as `ToolConfig.CommandFile`: parser populates `*File` only; resolver fills `Prompt`/`SystemPrompt` from disk at CLI entry points; LSP/WASM consumers see `*File != "" && content == ""` and that's the correct unresolved view.

**Files:**
- Modify: `ir/ir.go` lines 90–111 (AgentConfig struct)

- [ ] **Step 1: Write failing test in `ir/ir_test.go`**

If `ir/ir_test.go` does not exist, create it. If a similar test already exists for `CommandFile`, follow the same pattern. Add:

```go
package ir

import "testing"

func TestAgentConfig_PromptFileFields(t *testing.T) {
	cfg := AgentConfig{
		PromptFile:       "prompts/task.md",
		SystemPromptFile: "prompts/persona.md",
	}
	if cfg.PromptFile != "prompts/task.md" {
		t.Errorf("PromptFile = %q, want %q", cfg.PromptFile, "prompts/task.md")
	}
	if cfg.SystemPromptFile != "prompts/persona.md" {
		t.Errorf("SystemPromptFile = %q, want %q", cfg.SystemPromptFile, "prompts/persona.md")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `just test-pkg ir`
Expected: FAIL with "unknown field PromptFile in struct literal" (or similar Go compile error).

- [ ] **Step 3: Add the fields to `AgentConfig`**

In `ir/ir.go`, modify the `AgentConfig` struct (around lines 90-111). Add `PromptFile` after `Prompt` and `SystemPromptFile` after `SystemPrompt`:

```go
// AgentConfig holds configuration for LLM agent nodes.
type AgentConfig struct {
	Prompt              string
	PromptFile          string // Source path when Prompt was loaded from prompt_file:; empty if inline. Populated by parser.ResolveFileDirectives.
	SystemPrompt        string
	SystemPromptFile    string // Source path when SystemPrompt was loaded from system_prompt_file:; empty if inline. Populated by parser.ResolveFileDirectives.
	Model               string // Per-node override
	Provider            string
	MaxTurns            int
	CmdTimeout          time.Duration
	CacheTools          bool
	Compaction          string
	CompactionThreshold float64
	ReasoningEffort     string
	Fidelity            string
	AutoStatus          bool              // Parse STATUS: from response
	GoalGate            bool              // Pipeline fails if this node fails
	ResponseFormat      string            // "json_object" or "json_schema"
	ResponseSchema      string            // JSON schema (when ResponseFormat is "json_schema")
	Backend             string            // Per-node backend override: "native", "claude-code", "acp"
	WorkingDir          string            // Per-node working directory override
	ToolAccess          string            // Raw value: "" or "none" recognized; other values lint as DIP139 and fail-closed to no-tools at runtime
	Params              map[string]string // Generic key-value pairs passed through to runtime
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `just test-pkg ir`
Expected: PASS.

- [ ] **Step 5: Run full check to ensure nothing else broke**

Run: `just check`
Expected: PASS. (If existing tests in other packages fail, they should be type-related and trivial — usually nothing breaks from a new optional field.)

- [ ] **Step 6: Commit**

```bash
git add ir/ir.go ir/ir_test.go
git commit -m "feat(ir): add PromptFile and SystemPromptFile to AgentConfig (#65)"
```

---

## Task 2: Parser — handle `prompt_file:` / `system_prompt_file:` keys + mutual-exclusion checks

**Purpose:** Extend the agent-field parser to recognize the two new keys and to emit parser-time diagnostics when an author sets both `prompt` + `prompt_file` (or both `system_prompt` + `system_prompt_file`) on the same agent. Mirrors the `checkCommandFileConflict` pattern from #52.

**Files:**
- Modify: `parser/parse_nodes.go` lines 231-265 (`applyAgentField`, `applyAgentStringField`, `applyAgentPromptField`)
- Test: `parser/parser_test.go`

- [ ] **Step 1: Write failing test for `prompt_file:` parser case**

Add to `parser/parser_test.go`:

```go
func TestParser_AgentPromptFile(t *testing.T) {
	src := `workflow W
  goal: "test"
  start: A
  exit: A

  agent A
    model: claude-sonnet-4-6
    prompt_file: prompts/task.md
`
	w, err := NewParser(src, "test.dip").Parse()
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if len(w.Nodes) != 1 {
		t.Fatalf("got %d nodes, want 1", len(w.Nodes))
	}
	cfg, ok := w.Nodes[0].Config.(ir.AgentConfig)
	if !ok {
		t.Fatalf("node 0 config is not AgentConfig: %T", w.Nodes[0].Config)
	}
	if cfg.PromptFile != "prompts/task.md" {
		t.Errorf("PromptFile = %q, want %q", cfg.PromptFile, "prompts/task.md")
	}
	if cfg.Prompt != "" {
		t.Errorf("Prompt = %q, want empty (parser must not load file)", cfg.Prompt)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `just test-pkg parser`
Expected: FAIL with `PromptFile = "", want "prompts/task.md"` (the parser silently drops the unrecognized key today).

- [ ] **Step 3: Add `prompt_file` / `system_prompt_file` cases to `applyAgentPromptField`**

In `parser/parse_nodes.go`, modify `applyAgentPromptField` (currently lines 250-265):

```go
// applyAgentPromptField handles prompt-related agent fields.
func applyAgentPromptField(cfg *ir.AgentConfig, key, val string) bool {
	switch key {
	case "prompt":
		cfg.Prompt = val
	case "prompt_file":
		cfg.PromptFile = val
	case "system_prompt":
		cfg.SystemPrompt = val
	case "system_prompt_file":
		cfg.SystemPromptFile = val
	case "reasoning_effort":
		cfg.ReasoningEffort = val
	case "response_schema":
		cfg.ResponseSchema = val
	default:
		return false
	}
	return true
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `just test-pkg parser -run TestParser_AgentPromptFile`
Expected: PASS.

- [ ] **Step 5: Write failing test for `system_prompt_file:` parser case**

Add to `parser/parser_test.go`:

```go
func TestParser_AgentSystemPromptFile(t *testing.T) {
	src := `workflow W
  goal: "test"
  start: A
  exit: A

  agent A
    model: claude-sonnet-4-6
    system_prompt_file: prompts/persona.md
`
	w, err := NewParser(src, "test.dip").Parse()
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	cfg := w.Nodes[0].Config.(ir.AgentConfig)
	if cfg.SystemPromptFile != "prompts/persona.md" {
		t.Errorf("SystemPromptFile = %q, want %q", cfg.SystemPromptFile, "prompts/persona.md")
	}
	if cfg.SystemPrompt != "" {
		t.Errorf("SystemPrompt = %q, want empty (parser must not load file)", cfg.SystemPrompt)
	}
}
```

- [ ] **Step 6: Run test to verify it passes**

Run: `just test-pkg parser -run TestParser_AgentSystemPromptFile`
Expected: PASS (Step 3's change already covers it).

- [ ] **Step 7: Write failing test for `prompt` + `prompt_file` conflict diagnostic**

Add to `parser/parser_test.go`:

```go
func TestParser_AgentPromptConflict(t *testing.T) {
	src := `workflow W
  goal: "test"
  start: A
  exit: A

  agent A
    model: claude-sonnet-4-6
    prompt: "inline"
    prompt_file: prompts/task.md
`
	_, err := NewParser(src, "test.dip").Parse()
	if err == nil {
		t.Fatal("expected diagnostic for prompt + prompt_file conflict, got nil error")
	}
	if !strings.Contains(err.Error(), "both `prompt` and `prompt_file`") {
		t.Errorf("diagnostic message missing expected phrase; got %q", err.Error())
	}
}
```

Make sure `strings` is in the imports.

- [ ] **Step 8: Run test to verify it fails**

Run: `just test-pkg parser -run TestParser_AgentPromptConflict`
Expected: FAIL (no diagnostic emitted).

- [ ] **Step 9: Add the two conflict-check helpers to `parser/parse_nodes.go`**

Add immediately after `applyAgentPromptField` (around line 265):

```go
// checkPromptFileConflict emits a diagnostic if both prompt: and prompt_file:
// are set on the same agent node. Parser-time error (not a DIP code) because
// the conflict is syntactic. Gated on the assigning key being prompt or
// prompt_file so subsequent unrelated agent string-field writes don't re-emit.
func (p *Parser) checkPromptFileConflict(cfg *ir.AgentConfig, key string, loc ir.SourceLocation) {
	if key != "prompt" && key != "prompt_file" {
		return
	}
	if cfg.Prompt != "" && cfg.PromptFile != "" {
		p.diagnostics = append(p.diagnostics, fmt.Sprintf(
			"agent node has both `prompt` and `prompt_file` set; choose one at %d:%d",
			loc.Line, loc.Column))
	}
}

// checkSystemPromptFileConflict — same shape, for system_prompt vs system_prompt_file.
func (p *Parser) checkSystemPromptFileConflict(cfg *ir.AgentConfig, key string, loc ir.SourceLocation) {
	if key != "system_prompt" && key != "system_prompt_file" {
		return
	}
	if cfg.SystemPrompt != "" && cfg.SystemPromptFile != "" {
		p.diagnostics = append(p.diagnostics, fmt.Sprintf(
			"agent node has both `system_prompt` and `system_prompt_file` set; choose one at %d:%d",
			loc.Line, loc.Column))
	}
}
```

- [ ] **Step 10: Wire the helpers into `applyAgentField`**

Modify `applyAgentField` (currently lines 231-237):

```go
// applyAgentField applies agent-specific configuration fields.
func (p *Parser) applyAgentField(cfg *ir.AgentConfig, key, val string, loc ir.SourceLocation) {
	if applyAgentStringField(cfg, key, val) {
		p.checkPromptFileConflict(cfg, key, loc)
		p.checkSystemPromptFileConflict(cfg, key, loc)
		return
	}
	p.applyAgentComplexField(cfg, key, val, loc)
}
```

(Both helpers are cheap and self-gating on key — calling both unconditionally is fine and keeps `applyAgentField` at cyclomatic 3.)

- [ ] **Step 11: Run conflict test to verify it passes**

Run: `just test-pkg parser -run TestParser_AgentPromptConflict`
Expected: PASS.

- [ ] **Step 12: Write failing test for `system_prompt` + `system_prompt_file` conflict**

Add to `parser/parser_test.go`:

```go
func TestParser_AgentSystemPromptConflict(t *testing.T) {
	src := `workflow W
  goal: "test"
  start: A
  exit: A

  agent A
    model: claude-sonnet-4-6
    system_prompt: "inline persona"
    system_prompt_file: prompts/persona.md
`
	_, err := NewParser(src, "test.dip").Parse()
	if err == nil {
		t.Fatal("expected diagnostic for system_prompt + system_prompt_file conflict, got nil error")
	}
	if !strings.Contains(err.Error(), "both `system_prompt` and `system_prompt_file`") {
		t.Errorf("diagnostic message missing expected phrase; got %q", err.Error())
	}
}
```

- [ ] **Step 13: Run test to verify it passes**

Run: `just test-pkg parser -run TestParser_AgentSystemPromptConflict`
Expected: PASS.

- [ ] **Step 14: Write failing test for cross-slot mix (should NOT error)**

Add to `parser/parser_test.go`:

```go
func TestParser_AgentCrossSlotMixOK(t *testing.T) {
	src := `workflow W
  goal: "test"
  start: A
  exit: A

  agent A
    model: claude-sonnet-4-6
    prompt_file: prompts/task.md
    system_prompt: "you are a code reviewer"
`
	w, err := NewParser(src, "test.dip").Parse()
	if err != nil {
		t.Fatalf("mixed slots should not error: %v", err)
	}
	cfg := w.Nodes[0].Config.(ir.AgentConfig)
	if cfg.PromptFile != "prompts/task.md" {
		t.Errorf("PromptFile = %q", cfg.PromptFile)
	}
	if cfg.SystemPrompt != "you are a code reviewer" {
		t.Errorf("SystemPrompt = %q", cfg.SystemPrompt)
	}
}
```

- [ ] **Step 15: Run test to verify it passes**

Run: `just test-pkg parser -run TestParser_AgentCrossSlotMixOK`
Expected: PASS.

- [ ] **Step 16: Run full parser test suite + complexity check**

Run: `just test-pkg parser` then `just complexity`
Expected: All tests PASS, no complexity violations. If `applyAgentField` trips cyclomatic >5, the cause is unrelated drift — investigate.

- [ ] **Step 17: Commit**

```bash
git add parser/parse_nodes.go parser/parser_test.go
git commit -m "feat(parser): handle prompt_file/system_prompt_file with mutual-exclusion (#65)"
```

---

## Task 3: Resolver — extend `resolveNodeDirective` with agent branch

**Purpose:** Teach the existing resolver to walk agent nodes and populate `Prompt` / `SystemPrompt` from disk when the corresponding `*File` is set. Reuse `loadDirectiveFile` and friends verbatim — security model and error-message contract carry over identically.

**Files:**
- Modify: `parser/resolve.go` (rename existing `resolveNodeDirective` body to `resolveToolDirective`, add `resolveAgentDirective`, add `loadInto` helper)
- Modify/Add: `parser/testdata/prompt_file/` (new fixture directory with `task.md` and `persona.md`)
- Test: `parser/resolve_test.go`

- [ ] **Step 1: Create the test fixture directory**

```bash
mkdir -p parser/testdata/prompt_file
```

- [ ] **Step 2: Create the prompt fixture files**

Create `parser/testdata/prompt_file/task.md`:

```
Review the diff at .ai/scratch/diff.patch.
fixture: ResolveFileDirectives prompt test
```

Create `parser/testdata/prompt_file/persona.md`:

```
You are a senior code reviewer.
fixture: ResolveFileDirectives system_prompt test
```

- [ ] **Step 3: Write failing test for agent prompt_file resolution**

Add to `parser/resolve_test.go`:

```go
func TestResolveFileDirectives_LoadsAgentPrompt(t *testing.T) {
	w := &ir.Workflow{
		Nodes: []*ir.Node{
			{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{
				PromptFile: "task.md",
			}},
		},
	}
	if err := ResolveFileDirectives(w, "testdata/prompt_file"); err != nil {
		t.Fatalf("ResolveFileDirectives: %v", err)
	}
	cfg := w.Nodes[0].Config.(ir.AgentConfig)
	if !strings.Contains(cfg.Prompt, "fixture: ResolveFileDirectives prompt test") {
		t.Errorf("Prompt not populated from file; got %q", cfg.Prompt)
	}
	if cfg.PromptFile != "task.md" {
		t.Errorf("PromptFile = %q, want %q (must be preserved post-resolve)", cfg.PromptFile, "task.md")
	}
}
```

Verify `ir.NodeAgent` is the correct constant — if `grep -n "NodeAgent\|NodeKind" ir/ir.go` shows a different name, use it instead.

- [ ] **Step 4: Run test to verify it fails**

Run: `just test-pkg parser -run TestResolveFileDirectives_LoadsAgentPrompt`
Expected: FAIL with `Prompt not populated from file; got ""` (current resolver only handles tool nodes).

- [ ] **Step 5: Restructure `resolveNodeDirective` to dispatch by node-config kind**

In `parser/resolve.go`, replace the current `resolveNodeDirective` function with the dispatcher + per-kind helpers. The current implementation (lines 38-50) becomes `resolveToolDirective`. The new structure:

```go
// resolveNodeDirective resolves any file-directive fields on a single node.
// Dispatches per node-config kind so the per-kind loader functions stay
// focused on their own field set.
func resolveNodeDirective(n *ir.Node, baseDir string) error {
	switch cfg := n.Config.(type) {
	case ir.ToolConfig:
		return resolveToolDirective(n, cfg, baseDir)
	case ir.AgentConfig:
		return resolveAgentDirective(n, cfg, baseDir)
	}
	return nil
}

// resolveToolDirective populates ToolConfig.Command from CommandFile, if set.
// Skips if Command is already inline-populated (parser's mutual-exclusion
// check should prevent both being set; defensive guard if it doesn't).
func resolveToolDirective(n *ir.Node, cfg ir.ToolConfig, baseDir string) error {
	if cfg.CommandFile == "" || cfg.Command != "" {
		return nil
	}
	contents, err := loadDirectiveFile(baseDir, cfg.CommandFile)
	if err != nil {
		return fmt.Errorf("node %q command_file: %w", n.ID, err)
	}
	cfg.Command = string(contents)
	n.Config = cfg
	return nil
}

// resolveAgentDirective populates Prompt and SystemPrompt from their *File
// twins on AgentConfig. The two slots are independent — either, both, or
// neither may be set.
func resolveAgentDirective(n *ir.Node, cfg ir.AgentConfig, baseDir string) error {
	if err := loadInto(&cfg.Prompt, cfg.PromptFile, baseDir, n.ID, "prompt_file"); err != nil {
		return err
	}
	if err := loadInto(&cfg.SystemPrompt, cfg.SystemPromptFile, baseDir, n.ID, "system_prompt_file"); err != nil {
		return err
	}
	n.Config = cfg
	return nil
}

// loadInto populates *dst from path (relative to baseDir) if path != "" and
// *dst == "". Skips if either condition fails (defensive: parser's mutual-
// exclusion check should prevent both being set, but if it happens, inline wins).
func loadInto(dst *string, path, baseDir, nodeID, directive string) error {
	if path == "" || *dst != "" {
		return nil
	}
	contents, err := loadDirectiveFile(baseDir, path)
	if err != nil {
		return fmt.Errorf("node %q %s: %w", nodeID, directive, err)
	}
	*dst = string(contents)
	return nil
}
```

- [ ] **Step 6: Run agent test to verify it passes**

Run: `just test-pkg parser -run TestResolveFileDirectives_LoadsAgentPrompt`
Expected: PASS.

- [ ] **Step 7: Verify tool tests still pass (regression check)**

Run: `just test-pkg parser -run TestResolveFileDirectives`
Expected: All existing tool-side resolver tests PASS unchanged.

- [ ] **Step 8: Write failing test for agent system_prompt_file resolution**

Add to `parser/resolve_test.go`:

```go
func TestResolveFileDirectives_LoadsAgentSystemPrompt(t *testing.T) {
	w := &ir.Workflow{
		Nodes: []*ir.Node{
			{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{
				SystemPromptFile: "persona.md",
			}},
		},
	}
	if err := ResolveFileDirectives(w, "testdata/prompt_file"); err != nil {
		t.Fatalf("ResolveFileDirectives: %v", err)
	}
	cfg := w.Nodes[0].Config.(ir.AgentConfig)
	if !strings.Contains(cfg.SystemPrompt, "fixture: ResolveFileDirectives system_prompt test") {
		t.Errorf("SystemPrompt not populated from file; got %q", cfg.SystemPrompt)
	}
	if cfg.SystemPromptFile != "persona.md" {
		t.Errorf("SystemPromptFile = %q, want %q", cfg.SystemPromptFile, "persona.md")
	}
}
```

- [ ] **Step 9: Run test to verify it passes**

Run: `just test-pkg parser -run TestResolveFileDirectives_LoadsAgentSystemPrompt`
Expected: PASS.

- [ ] **Step 10: Write failing test for both file directives on the same agent**

Add to `parser/resolve_test.go`:

```go
func TestResolveFileDirectives_LoadsBothAgentSlots(t *testing.T) {
	w := &ir.Workflow{
		Nodes: []*ir.Node{
			{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{
				PromptFile:       "task.md",
				SystemPromptFile: "persona.md",
			}},
		},
	}
	if err := ResolveFileDirectives(w, "testdata/prompt_file"); err != nil {
		t.Fatalf("ResolveFileDirectives: %v", err)
	}
	cfg := w.Nodes[0].Config.(ir.AgentConfig)
	if !strings.Contains(cfg.Prompt, "ResolveFileDirectives prompt test") {
		t.Errorf("Prompt not populated; got %q", cfg.Prompt)
	}
	if !strings.Contains(cfg.SystemPrompt, "ResolveFileDirectives system_prompt test") {
		t.Errorf("SystemPrompt not populated; got %q", cfg.SystemPrompt)
	}
}
```

- [ ] **Step 11: Run test to verify it passes**

Run: `just test-pkg parser -run TestResolveFileDirectives_LoadsBothAgentSlots`
Expected: PASS.

- [ ] **Step 12: Write failing test that error message identifies the agent directive name**

Add to `parser/resolve_test.go`:

```go
func TestResolveFileDirectives_AgentErrorIdentifiesDirective(t *testing.T) {
	w := &ir.Workflow{
		Nodes: []*ir.Node{
			{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{
				SystemPromptFile: "nonexistent.md",
			}},
		},
	}
	err := ResolveFileDirectives(w, "testdata/prompt_file")
	if err == nil {
		t.Fatal("expected missing-file error, got nil")
	}
	if !strings.Contains(err.Error(), "system_prompt_file") {
		t.Errorf("error should identify directive `system_prompt_file`; got %v", err)
	}
	if strings.Contains(err.Error(), "prompt_file:") && !strings.Contains(err.Error(), "system_prompt_file") {
		t.Errorf("error must not be ambiguous between prompt_file and system_prompt_file; got %v", err)
	}
}
```

- [ ] **Step 13: Run test to verify it passes**

Run: `just test-pkg parser -run TestResolveFileDirectives_AgentErrorIdentifiesDirective`
Expected: PASS.

- [ ] **Step 14: Run full parser suite + complexity check**

Run: `just test-pkg parser` then `just complexity`
Expected: All PASS, no complexity violations.

- [ ] **Step 15: Commit**

```bash
git add parser/resolve.go parser/resolve_test.go parser/testdata/prompt_file/
git commit -m "feat(resolver): walk agent nodes for prompt_file/system_prompt_file (#65)"
```

---

## Task 4: Pack-shadow — extend `clearCommandFileDirectives` to cover agent fields

**Purpose:** `dippin pack` runs each `.dip` through the pack-shadow preprocessor, which resolves file directives, clears the `*File` fields, and reformats so the bundled `.dip` carries inline content. Today the clear step only walks tool nodes; for #65 it must also walk agent nodes and clear `PromptFile` + `SystemPromptFile`. Rename to reflect the broader scope.

**Files:**
- Modify: `cmd/dippin/pack_shadow.go` (rename `clearCommandFileDirectives` → `clearFileDirectives`, extend body)
- Test: `cmd/dippin/pack_shadow_test.go` (if it exists; otherwise add a small integration test via the existing pack-test infrastructure)

- [ ] **Step 1: Locate any existing pack-shadow test**

Run: `ls cmd/dippin/*_test.go` and `grep -rln "prepShadowSourceTree\|clearCommandFileDirectives" cmd/dippin/`

If there is a `cmd/dippin/pack_shadow_test.go`, modify it. Otherwise, the integration coverage from Task 7 (example file packs end-to-end) will catch regressions; skip writing a dedicated unit test here.

- [ ] **Step 2: Rename the function and extend the body**

In `cmd/dippin/pack_shadow.go`, replace the existing `clearCommandFileDirectives` (currently lines 130-143) and the one call site (`inlineOne` at line 123) as follows.

Update the call site (line 123):

```go
	clearFileDirectives(wf)
```

Replace the function body:

```go
// clearFileDirectives walks every node in wf and clears any file-directive
// source-path fields (ToolConfig.CommandFile, AgentConfig.PromptFile,
// AgentConfig.SystemPromptFile). The formatter emits the *_file: directive
// when these fields are set (preserving directive form on round-trip); we
// clear them after resolving so the formatter emits the inlined content
// blocks instead. Result: the shadow .dip carries inline content only,
// and the .dipx bundle is self-contained.
func clearFileDirectives(wf *ir.Workflow) {
	for _, n := range wf.Nodes {
		clearToolFileDirective(n)
		clearAgentFileDirectives(n)
	}
}

// clearToolFileDirective clears ToolConfig.CommandFile if set.
func clearToolFileDirective(n *ir.Node) {
	tc, ok := n.Config.(ir.ToolConfig)
	if !ok || tc.CommandFile == "" {
		return
	}
	tc.CommandFile = ""
	n.Config = tc
}

// clearAgentFileDirectives clears AgentConfig.PromptFile and
// AgentConfig.SystemPromptFile if either is set.
func clearAgentFileDirectives(n *ir.Node) {
	ac, ok := n.Config.(ir.AgentConfig)
	if !ok {
		return
	}
	changed := false
	if ac.PromptFile != "" {
		ac.PromptFile = ""
		changed = true
	}
	if ac.SystemPromptFile != "" {
		ac.SystemPromptFile = ""
		changed = true
	}
	if changed {
		n.Config = ac
	}
}
```

Per-kind helpers keep each function at cyclomatic ≤4. The single call site change is the rename `clearCommandFileDirectives(wf)` → `clearFileDirectives(wf)`.

- [ ] **Step 3: Run build + complexity check**

Run: `just check`
Expected: All PASS. The integration test for pack end-to-end behavior is in Task 7; this step just confirms the rename + extension compiles, passes existing tests, and meets complexity budgets.

- [ ] **Step 4: Commit**

```bash
git add cmd/dippin/pack_shadow.go
git commit -m "feat(pack): clear agent prompt/system_prompt file directives on pack (#65)"
```

---

## Task 5: Formatter — emit `prompt_file:` / `system_prompt_file:` AND fix missing inline `system_prompt:` emission

**Purpose:** Add conditional emission for both new directives, mirroring the `command_file:` pattern. Adjacent in-scope bug fix: today's `writeAgentFields` emits `cfg.Prompt` but does **not** emit `cfg.SystemPrompt` at all — `dippin fmt` silently drops inline `system_prompt:` on round-trip. The new conditional pair requires both branches present, so the missing emission ships at the same time and gets its own regression test.

**Files:**
- Modify: `formatter/format.go` (`writeAgentFields`, currently around line 323-337)
- Test: `formatter/format_test.go`

- [ ] **Step 1: Write failing regression test for inline `system_prompt:` round-trip**

Add to `formatter/format_test.go`:

```go
func TestFormat_AgentInlineSystemPromptRoundTrip(t *testing.T) {
	// Regression: pre-#65, writeAgentFields silently dropped SystemPrompt
	// on format. This test pins the fix.
	wf := &ir.Workflow{
		Name:  "W",
		Goal:  "test",
		Start: "A",
		Exit:  "A",
		Nodes: []*ir.Node{
			{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{
				Model:        "claude-sonnet-4-6",
				SystemPrompt: "You are a code reviewer.",
			}},
		},
	}
	out := Format(wf)
	if !strings.Contains(out, "system_prompt:") {
		t.Errorf("formatter dropped inline system_prompt; output:\n%s", out)
	}
	if !strings.Contains(out, "You are a code reviewer.") {
		t.Errorf("formatter dropped system_prompt content; output:\n%s", out)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `just test-pkg formatter -run TestFormat_AgentInlineSystemPromptRoundTrip`
Expected: FAIL — "formatter dropped inline system_prompt".

- [ ] **Step 3: Update `writeAgentFields` to emit all four field forms**

In `formatter/format.go`, modify `writeAgentFields` (around lines 323-337). Replace the trailing `if cfg.Prompt != ""` block with the two conditional pairs:

```go
func writeAgentFields(wr *writer, n *ir.Node, cfg ir.AgentConfig) {
	writeCommonNodeFields(wr, n)
	writeAgentModelFields(wr, cfg)
	writeAgentRuntimeFields(wr, cfg)
	writeAgentResponseFields(wr, cfg)
	writeAgentBehaviorFields(wr, cfg)
	writeAgentCompactionFields(wr, cfg)
	writeRetryFields(wr, n)
	writeIOFields(wr, n)
	writeSortedMapBlock(wr, "params", cfg.Params)
	writeAgentPromptFields(wr, cfg)
}

// writeAgentPromptFields emits the four prompt-related fields: system_prompt
// and prompt, each in either *_file: directive form (if *File is set) or
// inline multiline-block form (if the content field is set). The conditional
// pairs preserve the authored form across format round-trips.
func writeAgentPromptFields(wr *writer, cfg ir.AgentConfig) {
	if cfg.SystemPromptFile != "" {
		wr.line("system_prompt_file: %s", quoteValue(cfg.SystemPromptFile))
	} else if cfg.SystemPrompt != "" {
		wr.multilineBlock("system_prompt", cfg.SystemPrompt)
	}
	if cfg.PromptFile != "" {
		wr.line("prompt_file: %s", quoteValue(cfg.PromptFile))
	} else if cfg.Prompt != "" {
		wr.multilineBlock("prompt", cfg.Prompt)
	}
}
```

(Extracting into `writeAgentPromptFields` keeps both functions at cyclomatic ≤5. Emitting `system_prompt` first follows the convention of "frame before instruction" and matches how authors naturally order fields.)

- [ ] **Step 4: Run regression test to verify it passes**

Run: `just test-pkg formatter -run TestFormat_AgentInlineSystemPromptRoundTrip`
Expected: PASS.

- [ ] **Step 5: Write failing test for `prompt_file:` emission**

Add to `formatter/format_test.go`:

```go
func TestFormat_AgentPromptFile(t *testing.T) {
	wf := &ir.Workflow{
		Name:  "W",
		Goal:  "test",
		Start: "A",
		Exit:  "A",
		Nodes: []*ir.Node{
			{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{
				Model:      "claude-sonnet-4-6",
				PromptFile: "prompts/task.md",
			}},
		},
	}
	out := Format(wf)
	if !strings.Contains(out, "prompt_file: prompts/task.md") {
		t.Errorf("expected `prompt_file: prompts/task.md` in output; got:\n%s", out)
	}
	if strings.Contains(out, "prompt: ") && !strings.Contains(out, "prompt_file:") {
		t.Errorf("formatter emitted inline prompt: when PromptFile was set; got:\n%s", out)
	}
}
```

- [ ] **Step 6: Run test to verify it passes**

Run: `just test-pkg formatter -run TestFormat_AgentPromptFile`
Expected: PASS.

- [ ] **Step 7: Write failing test for `system_prompt_file:` emission**

Add to `formatter/format_test.go`:

```go
func TestFormat_AgentSystemPromptFile(t *testing.T) {
	wf := &ir.Workflow{
		Name:  "W",
		Goal:  "test",
		Start: "A",
		Exit:  "A",
		Nodes: []*ir.Node{
			{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{
				Model:            "claude-sonnet-4-6",
				SystemPromptFile: "prompts/persona.md",
			}},
		},
	}
	out := Format(wf)
	if !strings.Contains(out, "system_prompt_file: prompts/persona.md") {
		t.Errorf("expected `system_prompt_file: prompts/persona.md` in output; got:\n%s", out)
	}
}
```

- [ ] **Step 8: Run test to verify it passes**

Run: `just test-pkg formatter -run TestFormat_AgentSystemPromptFile`
Expected: PASS.

- [ ] **Step 9: Write failing test for cross-slot mix preservation**

Add to `formatter/format_test.go`:

```go
func TestFormat_AgentCrossSlotMix(t *testing.T) {
	wf := &ir.Workflow{
		Name:  "W",
		Goal:  "test",
		Start: "A",
		Exit:  "A",
		Nodes: []*ir.Node{
			{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{
				Model:        "claude-sonnet-4-6",
				PromptFile:   "prompts/task.md",
				SystemPrompt: "You are a reviewer.",
			}},
		},
	}
	out := Format(wf)
	if !strings.Contains(out, "prompt_file: prompts/task.md") {
		t.Errorf("missing prompt_file: line; got:\n%s", out)
	}
	if !strings.Contains(out, "system_prompt:") {
		t.Errorf("missing inline system_prompt:; got:\n%s", out)
	}
	if !strings.Contains(out, "You are a reviewer.") {
		t.Errorf("missing system_prompt content; got:\n%s", out)
	}
}
```

- [ ] **Step 10: Run test to verify it passes**

Run: `just test-pkg formatter -run TestFormat_AgentCrossSlotMix`
Expected: PASS.

- [ ] **Step 11: Run full formatter suite + complexity check**

Run: `just test-pkg formatter` then `just complexity`
Expected: All PASS, no complexity violations.

- [ ] **Step 12: Commit**

```bash
git add formatter/format.go formatter/format_test.go
git commit -m "feat(formatter): emit prompt_file/system_prompt_file + fix inline system_prompt drop (#65)"
```

---

## Task 6: Validator — teach DIP110 about `PromptFile`

**Purpose:** `lintEmptyPrompts` (DIP110) flags agent nodes with `cfg.Prompt == ""`. With the parser-pure design, an agent that uses `prompt_file:` only has `cfg.Prompt == ""` after parse (resolver runs at CLI entry, not in the lint test harness). Without this fix, any non-start/non-exit agent using `prompt_file:` only would get a false-positive DIP110 in `TestLintExamples` and in real-world lint runs. Fix: skip the warning when either inline content OR a file directive is present.

**Files:**
- Modify: `validator/lint_style.go::checkEmptyPrompt` (around lines 50-68)
- Test: `validator/lint_style_test.go` (or wherever DIP110 is tested)

- [ ] **Step 1: Write failing test that `prompt_file:` suppresses DIP110**

Find the existing DIP110 test (likely `TestLintEmptyPrompts` or similar — run `grep -n "DIP110\|lintEmptyPrompts\|checkEmptyPrompt" validator/*_test.go`). Add:

```go
func TestLintEmptyPrompts_PromptFileSuppresses(t *testing.T) {
	// Regression: an agent with prompt_file: set but empty inline Prompt
	// must not trip DIP110 — the prompt IS authored, just not yet loaded.
	w := &ir.Workflow{
		Name:  "W",
		Start: "Start",
		Exit:  "End",
		Nodes: []*ir.Node{
			{ID: "Start", Kind: ir.NodeAgent, Config: ir.AgentConfig{}},
			{ID: "Middle", Kind: ir.NodeAgent, Config: ir.AgentConfig{
				PromptFile: "prompts/task.md",
			}},
			{ID: "End", Kind: ir.NodeAgent, Config: ir.AgentConfig{}},
		},
	}
	diags := lintEmptyPrompts(w)
	for _, d := range diags {
		if d.Code == "DIP110" && strings.Contains(d.Message, "Middle") {
			t.Errorf("DIP110 fired on agent with prompt_file: set; diag: %+v", d)
		}
	}
}
```

Match the existing test file's imports and test helpers — if `lintEmptyPrompts` is unexported and not directly callable from a `_test.go` file outside `validator`, use the same call pattern existing DIP110 tests use.

- [ ] **Step 2: Run test to verify it fails**

Run: `just test-pkg validator -run TestLintEmptyPrompts_PromptFileSuppresses`
Expected: FAIL — DIP110 fires on the Middle node.

- [ ] **Step 3: Update `checkEmptyPrompt` to recognize the file form**

In `validator/lint_style.go`, modify `checkEmptyPrompt`:

```go
// checkEmptyPrompt checks a single node for DIP110.
func checkEmptyPrompt(n *ir.Node, w *ir.Workflow) (Diagnostic, bool) {
	if n.ID == w.Start || n.ID == w.Exit {
		return Diagnostic{}, false
	}
	cfg, ok := n.Config.(ir.AgentConfig)
	if !ok {
		return Diagnostic{}, false
	}
	if cfg.PromptFile != "" {
		return Diagnostic{}, false
	}
	if strings.TrimSpace(cfg.Prompt) == "" {
		return Diagnostic{
			Code:     DIP110,
			Severity: SeverityWarning,
			Message:  fmt.Sprintf("agent node %q has an empty prompt", n.ID),
			Location: n.Source,
			Help:     "add a prompt: field with instructions for the LLM",
		}, true
	}
	return Diagnostic{}, false
}
```

The added early-return on `cfg.PromptFile != ""` keeps the function at cyclomatic 5 (was 4; one new branch). If `just complexity` complains, factor the start/exit and config-type-assert checks into a small helper.

- [ ] **Step 4: Run regression test to verify it passes**

Run: `just test-pkg validator -run TestLintEmptyPrompts_PromptFileSuppresses`
Expected: PASS.

- [ ] **Step 5: Run full validator suite + complexity check**

Run: `just test-pkg validator` then `just complexity`
Expected: All PASS, no complexity violations.

- [ ] **Step 6: Commit**

```bash
git add validator/lint_style.go validator/lint_style_test.go
git commit -m "fix(validator): DIP110 skips agents with prompt_file: set (#65)"
```

---

## Task 7: Migrate parity — add new fields to `compareAgentConfigs`

**Purpose:** The migrate parity comparator walks every field of agent/tool configs to detect DOT-roundtrip regressions. Without adding the two new fields, parity tests would silently ignore drift on `PromptFile` / `SystemPromptFile`.

**Files:**
- Modify: `migrate/parity.go` (find `compareAgentConfigs` or the agent-side parity helper)

- [ ] **Step 1: Locate the agent parity comparator**

Run: `grep -n "compareAgentConfigs\|AgentConfig" migrate/parity.go`

Identify the field-by-field comparison block. The pattern from #52 is a series of `if a.X != b.X { diffs = append(diffs, "X") }` lines.

- [ ] **Step 2: Add the two new field comparisons**

In `migrate/parity.go`, in the agent parity comparator, add (clustered with `Prompt` / `SystemPrompt`):

```go
	if a.PromptFile != b.PromptFile {
		diffs = append(diffs, "PromptFile")
	}
	if a.SystemPromptFile != b.SystemPromptFile {
		diffs = append(diffs, "SystemPromptFile")
	}
```

The exact variable names (`a` / `b`, `diffs`) and append style must match the existing convention in the file — read the surrounding lines and mirror them.

- [ ] **Step 3: Run migrate tests**

Run: `just test-pkg migrate`
Expected: PASS.

- [ ] **Step 4: Run full check**

Run: `just check`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add migrate/parity.go
git commit -m "feat(migrate): include prompt/system_prompt file fields in agent parity (#65)"
```

---

## Task 8: Example workflow + end-to-end integration

**Purpose:** Demonstrate the canonical use case (an agent with both a persona system prompt and a task user prompt, each in its own external file) and pin pack-time inlining behavior end-to-end. The example file must validate cleanly via `just validate-examples`.

**Files:**
- Create: `examples/external_prompts.dip`
- Create: `examples/external_prompts/reviewer-persona.md`
- Create: `examples/external_prompts/reviewer-task.md`
- Test: `parser/resolve_test.go` (CLI-style integration coverage)

- [ ] **Step 1: Create the example workflow**

Create `examples/external_prompts.dip`:

```
workflow ExternalPrompts
  goal: "Demonstrate prompt_file:/system_prompt_file: directives (issue #65)"
  start: Reviewer
  exit: Reviewer

  agent Reviewer
    model: claude-sonnet-4-6
    system_prompt_file: external_prompts/reviewer-persona.md
    prompt_file: external_prompts/reviewer-task.md
    auto_status: true
```

- [ ] **Step 2: Create the system prompt fixture**

Create `examples/external_prompts/reviewer-persona.md`:

```
You are a senior code reviewer. Be concise, direct, and constructive.
Focus on correctness, security, and maintainability over style.
```

- [ ] **Step 3: Create the user prompt fixture**

Create `examples/external_prompts/reviewer-task.md`:

```
Review the diff at .ai/scratch/diff.patch.
Report findings ranked by severity (critical, important, minor).
End with STATUS: success or STATUS: fail.
```

- [ ] **Step 4: Verify the example validates**

Run: `just validate-examples`
Expected: PASS — `examples/external_prompts.dip` validates cleanly. (`dippin validate` uses the CLI parse path which invokes `ResolveFileDirectives`; the prompts resolve and the workflow validates as a normal agent node.)

- [ ] **Step 5: Verify the example lints clean**

Run: `just lint-examples`
Expected: PASS — no DIP warnings.

- [ ] **Step 6: Write end-to-end test that exercises the example through the resolver**

Add to `parser/resolve_test.go`:

```go
func TestResolveFileDirectives_ExternalPromptsExample(t *testing.T) {
	// Pins end-to-end resolution of examples/external_prompts.dip.
	// Mirrors the equivalent integration test for examples/external_files.dip
	// that #52 added.
	srcAbs, err := filepath.Abs("../examples/external_prompts.dip")
	if err != nil {
		t.Fatalf("Abs: %v", err)
	}
	data, err := os.ReadFile(srcAbs)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	wf, err := NewParser(string(data), srcAbs).Parse()
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if err := ResolveFileDirectives(wf, filepath.Dir(srcAbs)); err != nil {
		t.Fatalf("ResolveFileDirectives: %v", err)
	}
	var reviewer ir.AgentConfig
	for _, n := range wf.Nodes {
		if n.ID == "Reviewer" {
			reviewer = n.Config.(ir.AgentConfig)
		}
	}
	if !strings.Contains(reviewer.SystemPrompt, "senior code reviewer") {
		t.Errorf("SystemPrompt not loaded from file; got %q", reviewer.SystemPrompt)
	}
	if !strings.Contains(reviewer.Prompt, "STATUS: success") {
		t.Errorf("Prompt not loaded from file; got %q", reviewer.Prompt)
	}
}
```

- [ ] **Step 7: Run integration test**

Run: `just test-pkg parser -run TestResolveFileDirectives_ExternalPromptsExample`
Expected: PASS.

- [ ] **Step 8: Run full check**

Run: `just check`
Expected: PASS.

- [ ] **Step 9: Commit**

```bash
git add examples/external_prompts.dip examples/external_prompts/ parser/resolve_test.go
git commit -m "feat(examples): external_prompts.dip demonstrating both prompt directives (#65)"
```

---

## Task 9: Documentation updates

**Purpose:** Update every doc surface that enumerates agent fields or directives.

**Files:**
- Modify: `CHANGELOG.md`
- Modify: `docs/nodes.md`
- Modify: `docs/llm-reference.md`
- Modify: `docs/GRAMMAR.ebnf`
- Modify: `site/static/skill.md`

- [ ] **Step 1: Add CHANGELOG entry**

Open `CHANGELOG.md`. Add a new `## [v0.34.0] — UNRELEASED` section above the most recent `## [v0.33.0]` heading. Mirror the v0.33.0 section's shape:

```markdown
## [v0.34.0] — UNRELEASED

### Added
- `prompt_file:` and `system_prompt_file:` directives on agent nodes ([#65](https://github.com/2389-research/dippin-lang/issues/65)). Symmetric extension of v0.33.0's `command_file:`. Reuses the parser-pure resolver pass — pack-time inlining via the shadow tree carries the resolved content into `.dipx` bundles, so no runtime coordination is required.
- `parser.ResolveFileDirectives` now walks agent nodes in addition to tool nodes. `loadDirectiveFile` and the 4-layer security model (absolute reject, parent-tree escape reject, symlink reject, 4-MiB cap) are shared across all three directives.
- `examples/external_prompts.dip` demonstrating both prompt directives.

### Fixed
- `dippin fmt` previously silently dropped inline `system_prompt:` on agent nodes (pre-existing bug, latent since the field was introduced). Inline `system_prompt:` now round-trips correctly.

### Notes
- Parser stays pure (no FS I/O); resolver runs at CLI entry points only. LSP and WASM consumers see the unresolved IR view (`*File` set, content empty), which is the correct view for those contexts.
- Per-node only; `defaults agent` does not currently support `prompt` / `system_prompt` fields, so file-form-in-defaults requires a larger design — filed as a follow-up.
```

- [ ] **Step 2: Add nodes.md rows**

Open `docs/nodes.md`. Find the agent-node fields table. Add two rows (clustered near the existing `prompt` / `system_prompt` rows):

```markdown
| `prompt_file` | string | Path (relative to .dip dir) to an external file whose contents become the agent's `prompt`. Mutually exclusive with `prompt:`. |
| `system_prompt_file` | string | Path (relative to .dip dir) to an external file whose contents become the agent's `system_prompt`. Mutually exclusive with `system_prompt:`. |
```

If the table column structure differs from this 3-column shape, mirror the existing rows' column layout (read the surrounding rows and match the format).

- [ ] **Step 3: Add llm-reference.md entries**

Open `docs/llm-reference.md`. Find the agent optional-fields list. Add two entries near `prompt` / `system_prompt`:

```markdown
- `prompt_file: <relative-path>` — load `prompt` from external file
- `system_prompt_file: <relative-path>` — load `system_prompt` from external file
```

Match the exact list-item style in the surrounding section.

- [ ] **Step 4: Add GRAMMAR.ebnf productions**

Open `docs/GRAMMAR.ebnf`. Find the productions for `prompt` and `system_prompt`. Add two new productions immediately adjacent:

```ebnf
prompt_file ::= "prompt_file" ":" field_value
system_prompt_file ::= "system_prompt_file" ":" field_value
```

If `prompt` / `system_prompt` productions are inlined into a larger alternation (`agent_field ::= prompt | system_prompt | ...`), extend the alternation list to include the two new productions.

- [ ] **Step 5: Add skill.md subsection**

Open `site/static/skill.md`. Find the `command_file:` subsection introduced by v0.33. Add a new sibling subsection immediately after it:

```markdown
### `prompt_file:` and `system_prompt_file:`

Reference external prompt files from agent nodes:

```dip
agent Reviewer
  model: claude-sonnet-4-6
  system_prompt_file: prompts/persona.md
  prompt_file: prompts/task.md
```

Path resolution and security are identical to `command_file:` above:
- Paths resolved relative to the `.dip` source directory
- Absolute paths rejected
- Parent-tree escapes (`..`) rejected
- Symlinks rejected
- 4 MiB size cap

The two slots are independent — an agent may use any combination of inline `prompt:`, `prompt_file:`, inline `system_prompt:`, `system_prompt_file:`. Only same-slot conflict (`prompt:` + `prompt_file:`) is a parser-time error.

**Pack-time loading:** `dippin pack` inlines the prompt content into the bundled `.dip` so the `.dipx` is self-contained. The runtime reads inline prompts; no separate file lookup at runtime.

**Non-goals:** defaults-block file-form support and bundled-files `.dipx` redesign are tracked as follow-up issues.
```

Also add `prompt_file` and `system_prompt_file` rows to the skill.md agent-fields table (if one exists — match the same shape as `command_file:`).

- [ ] **Step 6: Verify generated docs regenerate cleanly**

Run: `just check` (the pre-commit hook regenerates `docs/generated-spec.md` if applicable; verify nothing else complains).
Expected: PASS.

- [ ] **Step 7: Commit**

```bash
git add CHANGELOG.md docs/nodes.md docs/llm-reference.md docs/GRAMMAR.ebnf site/static/skill.md docs/generated-spec.md
git commit -m "docs: prompt_file/system_prompt_file directives (#65)"
```

(`docs/generated-spec.md` may or may not have changed; if `git status` shows it unmodified, drop it from the add list.)

---

## Task 10: LSP completion + VSCode TextMate grammar

**Purpose:** Surface the two new field names in completion and syntax highlighting.

**Files:**
- Modify: `lsp/completion.go`
- Modify: `editors/vscode/syntaxes/dippin.tmLanguage.json` (around line 148)

- [ ] **Step 1: Add LSP completion entries**

Open `lsp/completion.go`. Find the agent-completion list (look for entries like `prompt`, `system_prompt`, `model`, or for `command_file` (which v0.33 added)). Add:

```go
	{Label: "prompt_file", Kind: protocol.CompletionItemKindField, Detail: "Load prompt from external file (relative to .dip dir)"},
	{Label: "system_prompt_file", Kind: protocol.CompletionItemKindField, Detail: "Load system_prompt from external file (relative to .dip dir)"},
```

If the existing completion entries use a different shape (e.g., snippet form or different field names on the struct), mirror the existing `command_file` entry's exact shape. Locate it with: `grep -n "command_file" lsp/completion.go`.

- [ ] **Step 2: Add VSCode field-name regex entries**

Open `editors/vscode/syntaxes/dippin.tmLanguage.json`. Around line 148 (the agent-field alternation regex that includes `command_file`), add `prompt_file` and `system_prompt_file` to the alternation. The exact form depends on the regex shape — find the line containing `command_file` and extend the alternation in place:

```
\\b(model|provider|prompt|system_prompt|prompt_file|system_prompt_file|command_file|...)\\b
```

(The actual line is much longer; do not rewrite it from scratch. Read the existing line, locate the `command_file` token in the alternation, and add `prompt_file|system_prompt_file|` immediately before it.)

- [ ] **Step 3: Run LSP tests**

Run: `just test-pkg lsp`
Expected: PASS.

- [ ] **Step 4: Run full check**

Run: `just check`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add lsp/completion.go editors/vscode/syntaxes/dippin.tmLanguage.json
git commit -m "feat(lsp, vscode): prompt_file/system_prompt_file completion + highlighting (#65)"
```

---

## Task 11: Open PR

**Purpose:** Open the PR for review. Tag-and-release runs separately after merge (see Task 12).

**Files:** None — pure git/gh operations.

- [ ] **Step 1: Push branch and open PR**

Run:
```bash
git push -u origin <feature-branch-name>

gh pr create --repo 2389-research/dippin-lang \
  --title "v0.34.0: prompt_file:/system_prompt_file: directives (#65) + fix inline system_prompt drop" \
  --body "$(cat <<'EOF'
## Summary

Symmetric extension of v0.33.0's `command_file:` to agent-node prompts. Two new directives on agent nodes: `prompt_file:` and `system_prompt_file:`. Reuses the parser-pure resolver, the 4-layer security model, and the pack-time shadow-tree inlining mechanism introduced in #52.

**Adjacent bug fix in scope:** `dippin fmt` previously silently dropped inline `system_prompt:` on agent nodes (pre-existing bug). Fixed.

**No runtime coordination required** — `.dipx` bundles carry inlined prompt content, same as commands.

## Test plan

- [ ] `just check` passes
- [ ] `just validate-examples` includes the new `examples/external_prompts.dip`
- [ ] `dippin fmt` round-trips a file with `prompt_file:` + inline `system_prompt:` losslessly
- [ ] `dippin pack` of the new example produces a `.dipx` whose unpacked `.dip` has inline `prompt:` / `system_prompt:` content
- [ ] Parser emits the expected conflict diagnostic for `prompt:` + `prompt_file:` on the same agent
- [ ] Parser emits the expected conflict diagnostic for `system_prompt:` + `system_prompt_file:` on the same agent

## Spec

[`docs/superpowers/specs/2026-05-27-issue-65-prompt-file-design.md`](docs/superpowers/specs/2026-05-27-issue-65-prompt-file-design.md)

## Follow-ups filed before merge

- `defaults agent prompt_file/system_prompt_file` support (#<DEFAULTS_ISSUE>)
- Bundled-files `.dipx` redesign (#<BUNDLED_ISSUE>)

Closes #65.
EOF
)"
```

Use the issue numbers from Task 0 in place of `<DEFAULTS_ISSUE>` and `<BUNDLED_ISSUE>`.

- [ ] **Step 2: Watch CI**

```bash
gh pr checks --watch
```

Expected: all checks pass within ~5 minutes. If any fail, diagnose and push fixes as separate commits (do not amend). Common failure: pre-commit hook reformatted a doc file; rerun `git add` + commit.

- [ ] **Step 3: Address review feedback as it lands**

For each reviewer comment: read carefully, decide if valid, fix if so. Use `gh pr-review` (the prbuddy extension) to triage if there are many comments at once.

---

## Task 12: Tag v0.34.0 (post-merge)

**Purpose:** Cut the v0.34.0 release once the PR is merged to main.

**Files:**
- Modify: `CHANGELOG.md` (replace `UNRELEASED` with today's date)

**Precondition:** PR from Task 11 is merged to main. You are on `main`, fully synced (`git pull origin main`).

- [ ] **Step 1: Update CHANGELOG date**

In `CHANGELOG.md`, find `## [v0.34.0] — UNRELEASED` and replace `UNRELEASED` with today's date in `YYYY-MM-DD` form. Use:

```bash
TODAY=$(date -u +%F)
```

and edit by hand or via sed:

```bash
sed -i.bak "s/## \\[v0.34.0\\] — UNRELEASED/## [v0.34.0] — ${TODAY}/" CHANGELOG.md && rm CHANGELOG.md.bak
```

- [ ] **Step 2: Commit the release-prep change**

```bash
git -C /home/clint/code/2389/dippin-lang add CHANGELOG.md
git -C /home/clint/code/2389/dippin-lang commit -m "release: prep v0.34.0"
```

(Use `git -C <path>` to avoid the shell-reset-after-checkout pattern that misrouted v0.32 prep onto the wrong branch.)

- [ ] **Step 3: Push the release-prep commit**

```bash
git -C /home/clint/code/2389/dippin-lang push origin main
```

- [ ] **Step 4: Tag v0.34.0**

```bash
git -C /home/clint/code/2389/dippin-lang tag -a v0.34.0 -m "v0.34.0: prompt_file/system_prompt_file directives (#65)"
git -C /home/clint/code/2389/dippin-lang push origin v0.34.0
```

- [ ] **Step 5: Watch GoReleaser**

```bash
gh run watch
```

Expected: GoReleaser builds cross-platform binaries, publishes the GitHub release, and updates the Homebrew tap within ~5 minutes.

- [ ] **Step 6: Verify release**

```bash
gh release view v0.34.0
```

Expected: release page shows the v0.34.0 binaries and the CHANGELOG-style notes.

- [ ] **Step 7: Close issue #65**

```bash
gh issue close 65 --repo 2389-research/dippin-lang --comment "Shipped in v0.34.0: https://github.com/2389-research/dippin-lang/releases/tag/v0.34.0"
```
