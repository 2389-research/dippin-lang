# #175 Shared Prompt Fragments Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Single-source shared prompt fragments — a `defaults` cascade (`prompt_prefix`/`prompt_suffix` + `_file` variants) applied to every agent, per-node `prompt_include:`, and per-node `prompt_prefix: none`/`prompt_suffix: none` opt-out — composed at resolve time into the effective prompt, round-tripped by the formatter, and honored by pack.

**Architecture:** Authored directive fields on `ir.WorkflowDefaults` + `ir.AgentConfig`; the formatter (unresolved IR) emits directives verbatim; `parser.ResolveFileDirectives` loads fragment files through the existing `loadDirectiveInto` security envelope and composes `prefix → body → include → suffix` into `AgentConfig.Prompt`. Additive and non-breaking.

**Tech Stack:** Go; `just`; pre-commit gate (`export PATH="/usr/local/go/bin:$PATH"` first). Tree-sitter regen via `npx tree-sitter generate`.

**Spec:** `docs/superpowers/specs/2026-07-10-issue-175-shared-prompt-fragments-design.md`

---

### Task 1: IR fields

**Files:** Modify `ir/ir.go` (`WorkflowDefaults`, `AgentConfig`).

- [ ] **Step 1: Add fields (no behavior yet).** To `WorkflowDefaults`:
```go
	PromptPrefix     string // defaults cascade: inline prefix applied to every agent's prompt
	PromptSuffix     string // defaults cascade: inline suffix applied to every agent's prompt
	PromptPrefixFile string // defaults cascade: prefix fragment loaded from a file
	PromptSuffixFile string // defaults cascade: suffix fragment loaded from a file
```
To `AgentConfig`:
```go
	PromptInclude string // fragment file appended after the body (before the cascade suffix)
	PromptPrefix  string // node-level: "none" opts out of the defaults prefix cascade ("" = inherit)
	PromptSuffix  string // node-level: "none" opts out of the defaults suffix cascade ("" = inherit)
```
- [ ] **Step 2: Build.** `export PATH="/usr/local/go/bin:$PATH"; go build ./...` → PASS (fields unused so far is fine; they're struct fields).
- [ ] **Step 3: Commit.** `git add ir/ir.go && git commit -m "feat(ir): prompt-fragment fields on defaults + agent (#175)"`

---

### Task 2: Composition helper + tests (the heart)

**Files:** Create `ir/prompt_compose.go`; Test `ir/prompt_compose_test.go`.

This helper is pure (no file IO) — it takes already-resolved fragment strings and assembles the effective prompt. Resolve (Task 4) loads the files then calls it.

- [ ] **Step 1: Write the failing test.**
```go
package ir

import "testing"

func TestComposePrompt(t *testing.T) {
	cases := []struct {
		name                            string
		prefix, body, include, suffix   string
		want                            string
	}{
		{"body only", "", "do X", "", "", "do X"},
		{"prefix+body+suffix", "P", "B", "", "S", "P\n\nB\n\nS"},
		{"include after body before suffix", "", "B", "I", "S", "B\n\nI\n\nS"},
		{"empty parts no blank lines", "", "B", "", "", "B"},
		{"suffix always last", "P", "B", "I", "S", "P\n\nB\n\nI\n\nS"},
		{"all empty", "", "", "", "", ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := ComposePrompt(c.prefix, c.body, c.include, c.suffix); got != c.want {
				t.Errorf("ComposePrompt = %q, want %q", got, c.want)
			}
		})
	}
}
```
- [ ] **Step 2: Run → FAIL** (`undefined: ComposePrompt`). `go test ./ir/ -run TestComposePrompt`
- [ ] **Step 3: Implement `ir/prompt_compose.go`.**
```go
package ir

import "strings"

// ComposePrompt assembles the effective prompt from already-resolved parts,
// in order prefix → body → include → suffix, joining only non-empty parts with
// a blank line. The suffix is always last (satisfies "final line must be …").
func ComposePrompt(prefix, body, include, suffix string) string {
	parts := make([]string, 0, 4)
	for _, p := range []string{prefix, body, include, suffix} {
		if p != "" {
			parts = append(parts, p)
		}
	}
	return strings.Join(parts, "\n\n")
}
```
- [ ] **Step 4: Run → PASS.** `go test ./ir/ -run TestComposePrompt`
- [ ] **Step 5: Commit.** `git add ir/prompt_compose.go ir/prompt_compose_test.go && git commit -m "feat(ir): ComposePrompt fragment assembly (#175)"`

---

### Task 3: Parser — defaults + agent directives

**Files:** Modify `parser/parse_defaults.go` (new defaults keywords + mutual-exclusion), `parser/parse_nodes.go` (agent `prompt_include`, `prompt_prefix`, `prompt_suffix`). Test: `parser/parse_defaults_test.go`, `parser/parse_nodes_test.go` (use existing test files/idiom).

**Pattern to mirror:** how `prompt_file`/`system_prompt_file` agent fields and existing string defaults are parsed (`applyDefaultStringField` in parse_defaults.go; the agent `*_file` field handling in parse_nodes.go `tryApplyCommonField`/agent field dispatch).

- [ ] **Step 1: Failing test — defaults cascade parse.**
```go
func TestParseDefaults_PromptCascade(t *testing.T) {
	src := `workflow W
  start: A
  exit: A
  defaults
    prompt_suffix_file: frag/status.md
    prompt_prefix: "hello"
  agent A
    prompt: "x"
`
	w, err := NewParser(src, "t.dip").Parse()
	if err != nil { t.Fatal(err) }
	if w.Defaults.PromptSuffixFile != "frag/status.md" { t.Errorf("suffix_file=%q", w.Defaults.PromptSuffixFile) }
	if w.Defaults.PromptPrefix != "hello" { t.Errorf("prefix=%q", w.Defaults.PromptPrefix) }
}

func TestParseDefaults_PromptSuffixBothFormsError(t *testing.T) {
	src := `workflow W
  start: A
  exit: A
  defaults
    prompt_suffix: "x"
    prompt_suffix_file: f.md
  agent A
    prompt: "y"
`
	_, err := NewParser(src, "t.dip").Parse()
	if err == nil || !strings.Contains(err.Error(), "prompt_suffix") {
		t.Fatalf("want both-forms error, got %v", err)
	}
}

func TestParseAgent_PromptIncludeAndOptOut(t *testing.T) {
	src := `workflow W
  start: A
  exit: A
  agent A
    prompt: "x"
    prompt_include: frag/extra.md
    prompt_suffix: none
`
	w, err := NewParser(src, "t.dip").Parse()
	if err != nil { t.Fatal(err) }
	cfg := w.Nodes[0].Config.(AgentConfig)
	if cfg.PromptInclude != "frag/extra.md" { t.Errorf("include=%q", cfg.PromptInclude) }
	if cfg.PromptSuffix != "none" { t.Errorf("suffix=%q", cfg.PromptSuffix) }
}
```
(Adjust `w.Defaults` accessor name — grep ir for the field on Workflow holding `*WorkflowDefaults`.)
- [ ] **Step 2: Run → FAIL.** `go test ./parser/ -run 'PromptCascade|PromptSuffixBoth|PromptIncludeAndOptOut'`
- [ ] **Step 3: Implement.**
  - `parse_defaults.go`: register `prompt_prefix`, `prompt_suffix`, `prompt_prefix_file`, `prompt_suffix_file` as string defaults (extend `applyDefaultStringField` or the dispatch it delegates to). After assignment, emit a structural diagnostic if both `PromptSuffix` and `PromptSuffixFile` (or both prefix fields) end up non-empty: `fmt.Sprintf("defaults sets both prompt_suffix and prompt_suffix_file at %d:%d — use one", ...)`.
  - `parse_nodes.go`: register agent fields `prompt_include`, `prompt_prefix`, `prompt_suffix` (string assignment into `AgentConfig`). Follow the exact dispatch used for `prompt_file`.
- [ ] **Step 4: Run → PASS + full parser suite.** `go test ./parser/`
- [ ] **Step 5: Commit.** `git add parser/ && git commit -m "feat(parser): prompt-fragment directives on defaults + agent (#175)"`

---

### Task 4: Resolve-time composition

**Files:** Modify `parser/resolve.go` (`resolveAgentDirective`). Test: `parser/resolve_test.go` + a `parser/testdata/` fragment fixture dir.

- [ ] **Step 1: Failing test.** Create `parser/testdata/frag_status.md` containing `STATUS: success or STATUS: fail`. Then:
```go
func TestResolve_CascadeAndInclude(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir,"suffix.md"), []byte("END WITH STATUS"), 0o644)
	os.WriteFile(filepath.Join(dir,"extra.md"), []byte("EXTRA"), 0o644)
	src := `workflow W
  start: A
  exit: B
  defaults
    prompt_suffix_file: suffix.md
  agent A
    prompt: "body A"
    prompt_include: extra.md
  agent B
    prompt: "body B"
    prompt_suffix: none
`
	w, err := NewParser(src, filepath.Join(dir,"w.dip")).Parse()
	if err != nil { t.Fatal(err) }
	if err := ResolveFileDirectives(w, dir); err != nil { t.Fatal(err) }
	a := w.Nodes[0].Config.(ir.AgentConfig).Prompt
	if a != "body A\n\nEXTRA\n\nEND WITH STATUS" { t.Errorf("A prompt=%q", a) }
	b := w.Nodes[1].Config.(ir.AgentConfig).Prompt
	if b != "body B" { t.Errorf("B (opted out) prompt=%q", b) }
}
```
(This test lives in `package parser` — adjust `ir.AgentConfig` references to the local import.)
- [ ] **Step 2: Run → FAIL.**
- [ ] **Step 3: Implement.** In `resolveAgentDirective`, after loading `Prompt`/`SystemPrompt` from their `*File` twins, load the fragment files and compose. Keep functions under the complexity caps by extracting a helper `composeAgentPrompt(cfg, defaults, baseDir, nodeID) (string, error)` that:
  1. loads `prompt_include` into a local `include` var via `loadDirectiveInto`;
  2. resolves the effective prefix: `"" if cfg.PromptPrefix == "none" else (defaults.PromptPrefix or loaded defaults.PromptPrefixFile)`;
  3. same for suffix;
  4. returns `ir.ComposePrompt(prefix, cfg.Prompt, include, suffix)`.
  Load each defaults fragment file **once per workflow** (cache the loaded prefix/suffix strings before iterating nodes) so N agents don't re-read the file N times — resolve currently iterates nodes; hoist the defaults-fragment load to `ResolveFileDirectives` and pass the strings down. Assign the composed result to `cfg.Prompt`, then `n.Config = cfg`.
  Reuse `loadDirectiveInto` for every file read (security envelope). A `none` sentinel is never treated as a path.
- [ ] **Step 4: Run → PASS + full parser suite.** `go test ./parser/`
- [ ] **Step 5: Security regression test.** Mirror an existing `prompt_file` symlink/oob/oversize test for `prompt_suffix_file` (grep `resolve_test.go` for the existing symlink test and clone it with a fragment directive). Confirm hard error.
- [ ] **Step 6: Commit.** `git add parser/ && git commit -m "feat(parser): compose prompt fragments at resolve time (#175)"`

---

### Task 5: Formatter round-trip

**Files:** Modify `formatter/format.go` (`writeAgentPromptFields`, defaults writer). Test: `formatter/format_test.go`.

- [ ] **Step 1: Failing test.** Format a workflow with defaults `prompt_suffix_file` + agent `prompt_include` + agent `prompt_suffix: none`; assert each directive is emitted, and that `Format(Parse(Format(x))) == Format(x)` (idempotent). Fragment files need not exist — the formatter uses **unresolved** IR (do NOT call ResolveFileDirectives in this test).
- [ ] **Step 2: Run → FAIL.**
- [ ] **Step 3: Implement.**
  - Defaults writer: emit `prompt_prefix`/`prompt_suffix` (quoted) and `prompt_prefix_file`/`prompt_suffix_file` when set (grep for where OnFailure/other defaults are written; add alongside).
  - `writeAgentPromptFields`: after the prompt/prompt_file emission, emit `prompt_include:` when set, and `prompt_prefix:`/`prompt_suffix:` when set (the `none` sentinel prints as `prompt_suffix: none`). Choose a deterministic field order.
- [ ] **Step 4: Run → PASS + full formatter suite.** `go test ./formatter/`
- [ ] **Step 5: Commit.** `git add formatter/ && git commit -m "feat(formatter): round-trip prompt-fragment directives (#175)"`

---

### Task 6: DIP154 lint

**Files:** `validator/lint_codes.go`, `validator/explanations.go`, new `validator/lint_prompt_optout.go`, register in `validator/lint.go`. Test: `validator/lint_prompt_optout_test.go`.

**Pattern to mirror:** DIP153 (`validator/lint_redundant_fan_edge.go`) end-to-end.

- [ ] **Step 1: Failing test.** An agent with `prompt_suffix: none` while `w.Defaults.PromptSuffix`/`PromptSuffixFile` are both empty → exactly 1 DIP154 (Hint). With a cascade declared → 0. Use `countCode(Lint(w).Diagnostics, DIP154)`.
- [ ] **Step 2: Run → FAIL** (`undefined: DIP154`).
- [ ] **Step 3: Implement.** Add `DIP154` const + description; explanation entry (Code/Summary/Trigger/Fix/Example); `lintPromptOptOut(w)` iterates agent nodes, flags `cfg.PromptPrefix == "none"` when no prefix cascade, and `cfg.PromptSuffix == "none"` when no suffix cascade; `SeverityHint`; register `lintPromptOptOut` in lint.go. Update the `DIP101–DIP153` range comment in lint.go/lint_codes.go to `DIP101–DIP154`.
- [ ] **Step 4: Run → PASS + full validator suite** (explanation-parity test must pass). `go test ./validator/`
- [ ] **Step 5: Commit.** `git add validator/ && git commit -m "feat(validator): DIP154 — prompt opt-out without cascade (#175)"`

---

### Task 7: pack --no-inline ships fragment files

**Files:** Modify `dipx/helpers.go` (the `*_file` collector near line 609). Test: `dipx/*_test.go` (mirror the existing prompt_file --no-inline test).

- [ ] **Step 1: Failing test.** Pack a workflow with `defaults.prompt_suffix_file` + agent `prompt_include` under `--no-inline`; assert the fragment files appear as bundle entries under `workflows/` and the directives are retained. (Clone the existing prompt_file no-inline test.)
- [ ] **Step 2: Run → FAIL.**
- [ ] **Step 3: Implement.** In the agent-config collector (helpers.go:609) add `cfg.PromptInclude`. Add a defaults collector for `w.Defaults.PromptPrefixFile` / `PromptSuffixFile` (find where the walk enumerates node configs; extend it to also yield the two defaults fragment paths once). Each path flows through the same containment/symlink validation as `PromptFile`.
- [ ] **Step 4: Run → PASS + full dipx suite.** `go test ./dipx/ ./cmd/dippin/`
- [ ] **Step 5: Commit.** `git add dipx/ && git commit -m "feat(dipx): ship prompt-fragment files under pack --no-inline (#175)"`

---

### Task 8: Editor grammars

**Files:** `editors/` tree-sitter grammar + corpus, VS Code tmLanguage, Zed.

- [ ] **Step 1:** Add the new keywords (`prompt_prefix`, `prompt_suffix`, `prompt_prefix_file`, `prompt_suffix_file`, `prompt_include`) to the tree-sitter grammar field list and highlights; regenerate with `npx tree-sitter generate` (run in the tree-sitter dir); add a corpus case exercising a defaults cascade + agent include.
- [ ] **Step 2:** Add the keywords to the VS Code TextMate grammar (mirror where `prompt_file` is listed) and the Zed grammar.
- [ ] **Step 3: Verify** the tree-sitter corpus test passes (`npx tree-sitter test` in that dir) and the generated parser is committed.
- [ ] **Step 4: Commit.** `git add editors/ && git commit -m "feat(editors): highlight prompt-fragment keywords (#175)"`

---

### Task 9: Docs / site / spec sweep + example

**Files:** `docs/nodes.md`, `docs/cli.md`, `docs/GRAMMAR.ebnf` (W3C EBNF, `/* */` comments), `docs/llm-reference.md`, `docs/validation.md`; `site/content/{language,cli}.md`, `site/static/skill.md`; a new example under `examples/`; regenerate `cmd/dippin/generated-spec.md`.

- [ ] **Step 1:** Document the defaults cascade, per-node `prompt_include`, and `none` opt-out in `docs/nodes.md`; note in `docs/cli.md`/`site/content/cli.md` that pack ships fragment files under `--no-inline`. Add the five keywords to `GRAMMAR.ebnf`. Add DIP154 to `docs/validation.md` (+ `site/content/validation.md` per the always-current directive) and bump the catalog count **63 → 64** / `DIP101–DIP153` → `DIP101–DIP154` across every hand-maintained doc/site surface (grep for `63`, `DIP101-DIP153`, `DIP101–DIP153`).
- [ ] **Step 2:** Add `examples/shared_prompt_fragment.dip` (+ a sibling fragment file) demonstrating a defaults `prompt_suffix_file` cascade across ≥2 agents with one opt-out. Confirm `dippin validate` + `dippin lint` clean and `just validate-examples` passes.
- [ ] **Step 3:** Regenerate spec (`bash scripts/gen-spec.sh`); `go test ./releasecheck/`.
- [ ] **Step 4: Commit.** `git add docs/ site/ cmd/dippin/generated-spec.md examples/ && git commit -m "docs: document shared prompt fragments + DIP154 (#175)"`

---

## Final verification

```bash
export PATH="/usr/local/go/bin:$PATH"
.git/hooks/pre-commit   # full CI-equivalent gate
```
Then squad review (correctness of composition/ordering/opt-out; security-envelope reuse; formatter round-trip; pack inline+no-inline; the #134-class check: does a packed+unpacked tree compose byte-identically to a source run?).

## Self-review notes
- Spec coverage: T1 IR, T2 compose, T3 parse, T4 resolve/compose+security, T5 formatter, T6 DIP154, T7 pack, T8 grammars, T9 docs+example. All acceptance criteria mapped.
- Type consistency: `ComposePrompt(prefix, body, include, suffix)` signature fixed in T2, consumed in T4. Field names (`PromptPrefix/Suffix[File]`, `PromptInclude`) consistent T1→T3→T4→T5→T7.
- Complexity: resolve composition extracted into `composeAgentPrompt` + hoisted defaults-fragment load to keep functions ≤5 cyclomatic / ≤7 cognitive (the pre-commit hook enforces).
- Known adjustment points flagged inline (real `w.Defaults` accessor, existing test helpers, exact formatter/dispatch call sites) — implementer greps for actual names.
