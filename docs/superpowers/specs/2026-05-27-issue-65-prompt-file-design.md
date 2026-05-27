# v0.34.0 — `prompt_file:` and `system_prompt_file:` directives on agent nodes (issue #65)

**Date:** 2026-05-27
**Closes:** [#65](https://github.com/2389-research/dippin-lang/issues/65)
**Status:** Design approved 2026-05-27
**Predecessor:** v0.33.0 (`command_file:`) shipped 2026-05-27 (PR #71). This spec is a symmetric extension — the resolver, security model, formatter pattern, pack-time shadow-tree inlining, and doc layout are all reused. Net delta is roughly 60% smaller than #52 because no new architecture is introduced.

## Problem

`.dip` agent nodes inline prompts via multiline `prompt:` and `system_prompt:` blocks. There is no way to reference an external file. The cited production pain in #52 was shell-script heredoc bloat; the same pain applies to prompts when an agent's system prompt becomes long enough to dominate the `.dip` source. #52 deferred the symmetric prompt directives until the analogous pain surfaced; we ship them now to close the directive-family asymmetry before authors hit the bloat in production and re-invent workarounds.

## Design

Two fields. One scope (per-node). Parser-pure. Reuse #52's resolver.

```dip
agent Reviewer
  model: claude-sonnet-4-6
  system_prompt_file: prompts/reviewer-persona.md
  prompt_file: prompts/reviewer-task.md
```

Each directive references an external file relative to the `.dip` source directory. The existing `parser.ResolveFileDirectives` pass (extended to walk agent nodes) loads the files at workflow-load time, populating the existing `Prompt` and `SystemPrompt` fields. The parser itself stays pure — no FS I/O — so LSP and WASM contexts work unchanged. The same `loadDirectiveFile` helper (4-layer security: absolute reject, parent-tree escape reject, symlink reject via `os.Lstat`, 4-MiB size cap) is the single source of truth for file loading across all three directives.

## Non-goals (v1)

These are tracked as numbered follow-up issues, filed before merge (§ Release coordination):

1. **Defaults-block support for `prompt_file:` / `system_prompt_file:`.** Per-node only in v1, matching what `command_file:` shipped with. Today's `defaults agent` block does not support `prompt` or `system_prompt` at all (see `parser/parse_defaults.go` recognized-keys list), so file-form support would have to ship alongside inline-form support — that's a meaningfully larger design (cascade semantics, override rules, per-node-vs-default mutual exclusion) and not justified by current pain. File-form-in-defaults ships when an author wants to share one large external persona doc across agents. ([follow-up filed before merge])
2. **Bundled-files `.dipx` redesign.** Currently `dippin pack` inlines file contents into the bundled `.dip` text via the shadow-tree machinery (`cmd/dippin/pack_shadow.go`). An alternative model would package the referenced scripts/prompts as separate files inside the `.dipx` and have tracker (and any other dipx consumer) call `parser.ResolveFileDirectives` against the unpacked bundle dir. Wins: tracker logs can reference original file paths; `.dipx` debug inspection shows author intent. Costs: tracker coordination required (the `.dipx` format version would bump); the shadow-tree machinery is obsoleted; old bundles need a cutover plan. Considered for this spec and explicitly deferred so v0.34 stays dippin-only. ([follow-up filed before merge])
3. **Glob support** (`prompt_file: prompts/*.md`). Composition semantics undefined; no current need. Mirrors the same follow-up filed against `command_file:` ([#68](https://github.com/2389-research/dippin-lang/issues/68)).

## Dippin-side design

### IR (`ir/ir.go`)

Two new fields on `AgentConfig`, each clustered next to its content twin:

```go
type AgentConfig struct {
    Prompt           string
    PromptFile       string  // Source path when Prompt was loaded from prompt_file:; empty if inline. Populated by parser.ResolveFileDirectives.
    SystemPrompt     string
    SystemPromptFile string  // Source path when SystemPrompt was loaded from system_prompt_file:; empty if inline. Populated by parser.ResolveFileDirectives.
    Model            string
    // ... rest unchanged
}
```

Same primary-representation contract as `ToolConfig.CommandFile`: parser populates `*File` only; resolver fills `Prompt`/`SystemPrompt` from disk at CLI entry points; LSP and WASM consumers see `*File != "" && content == ""` and that's the correct unresolved view for them. A post-resolver invariant would be `PromptFile != "" → Prompt != ""` (and same for system) but is not enforced in v1.

The two fields are independent. An agent may use any combination: inline prompt + file system_prompt, file prompt + inline system_prompt, both inline, both file. Only same-slot conflict (`prompt:` + `prompt_file:`) is an error; cross-slot mixing (`prompt:` + `system_prompt_file:`) is fine.

### Parser (`parser/parse_nodes.go`)

Two new cases on the agent-field dispatcher, each with the symmetric mutual-exclusion check, following the `command_file:` pattern from #52:

```go
case "prompt_file":
    cfg.PromptFile = val
    p.checkPromptFileConflict(cfg, key, loc)
case "system_prompt_file":
    cfg.SystemPromptFile = val
    p.checkSystemPromptFileConflict(cfg, key, loc)
```

And symmetric guards on the existing `prompt` and `system_prompt` cases — each calls the corresponding `checkPromptFileConflict` / `checkSystemPromptFileConflict` helper so either ordering triggers the diagnostic. Two helpers rather than one generic to keep cyclomatic ≤5 per function (matches the `checkCommandFileConflict` refactor pattern from #52).

```go
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
```

(`checkSystemPromptFileConflict` is identical with `SystemPrompt`/`SystemPromptFile` and `system_prompt`/`system_prompt_file`.)

**Parser stays pure** — no FS I/O. The two new cases store the path verbatim. Path security and file reading happen entirely in the resolver. These are parser-time errors (untyped diagnostics, not DIP codes), same rationale as #52.

### Resolver (`parser/resolve.go`)

The existing `resolveNodeDirective` grows an agent branch alongside the tool branch:

```go
func resolveNodeDirective(n *ir.Node, baseDir string) error {
    switch cfg := n.Config.(type) {
    case ir.ToolConfig:
        return resolveToolDirective(n, cfg, baseDir)
    case ir.AgentConfig:
        return resolveAgentDirective(n, cfg, baseDir)
    }
    return nil
}

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

// loadInto populates *dst from path if path != "" and *dst == "". Skips if
// either condition fails (defensive: parser mutual-exclusion check should
// prevent both being set, but if it happens, inline wins).
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

`loadDirectiveFile`, `safeResolve`, `statDirectiveFile`, `readDirectiveFile`, `checkFileInfo`, `hasParentRef` — all unchanged from #52. Same 4-MiB cap, same 4 security checks, same user-path-only error messages.

`resolveToolDirective` is extracted from the existing inline body of `resolveNodeDirective` — no behavior change for tool nodes, just a rename for symmetry.

If `loadInto` trips cyclomatic on `just complexity` (unlikely at 3), the early-return chain can be inlined into `resolveAgentDirective` and each `*File` field handled with its own 3-line block.

### CLI integration

No new touchpoints. The existing `parseAndResolveDip` helper in `cmd/dippin/cli.go` already calls `parser.ResolveFileDirectives` after parse; the resolver's new agent branch is picked up automatically by every CLI entry point that uses the helper.

`cmd_fmt.go` continues to use the bare parser path (no resolve) to preserve the directive form across `dippin fmt` — same trade-off as #52.

### Pack-time shadow tree (`cmd/dippin/pack_shadow.go`)

The shadow-tree machinery from #52 needs two surgical updates:

1. `parser.ResolveFileDirectives` (already called by `inlineOne`) now also populates `Prompt` and `SystemPrompt` for agent nodes with the corresponding `*File` set. No code change needed in pack_shadow — it inherits the new behavior from the resolver extension.

2. `clearCommandFileDirectives` is renamed `clearFileDirectives` and extended to walk agent nodes too:

```go
func clearFileDirectives(wf *ir.Workflow) {
    for _, n := range wf.Nodes {
        switch cfg := n.Config.(type) {
        case ir.ToolConfig:
            if cfg.CommandFile != "" {
                cfg.CommandFile = ""
                n.Config = cfg
            }
        case ir.AgentConfig:
            changed := false
            if cfg.PromptFile != "" {
                cfg.PromptFile = ""
                changed = true
            }
            if cfg.SystemPromptFile != "" {
                cfg.SystemPromptFile = ""
                changed = true
            }
            if changed {
                n.Config = cfg
            }
        }
    }
}
```

If this trips cyclomatic ≤5, extract the agent branch to its own helper (`clearAgentFileDirectives`) — keeps the parent function at 3, the helper at 4.

Result: pack-shadow walks each `.dip`, resolves all three directive families into IR content fields, clears all three `*File` fields, formats with inline `command:` / `prompt:` / `system_prompt:` blocks, writes to the shadow tree. `dipx.Pack` then bundles the fully-inlined `.dip` — tracker sees inline content, same as today. **No tracker coordination required for v0.34.**

### Formatter (`formatter/format.go`)

**Adjacent bug fix in scope:** today's `writeAgentFields` emits `cfg.Prompt` (line 334-336) but does **not** emit `cfg.SystemPrompt` at all. `dippin fmt` of a workflow with inline `system_prompt:` silently drops it on round-trip — a pre-existing data-loss bug. Since #65 adds `system_prompt_file:` emission via the conditional-pair pattern, the missing inline `system_prompt:` emission must ship at the same time (the conditional pair requires both branches). A focused regression test for inline `system_prompt:` round-trip is added alongside the directive tests.

Two new conditional blocks in `writeAgentFields`, replacing the existing trailing `Prompt` block:

```go
if cfg.PromptFile != "" {
    wr.line("prompt_file: %s", quoteValue(cfg.PromptFile))
} else if cfg.Prompt != "" {
    wr.multilineBlock("prompt", cfg.Prompt)
}
if cfg.SystemPromptFile != "" {
    wr.line("system_prompt_file: %s", quoteValue(cfg.SystemPromptFile))
} else if cfg.SystemPrompt != "" {
    wr.multilineBlock("system_prompt", cfg.SystemPrompt)
}
```

`dippin fmt` preserves the authored directive form for each field independently. A `.dip` file with `prompt_file:` + inline `system_prompt:` round-trips to itself; either directive can be present without the other; inline `system_prompt:` now round-trips correctly for the first time.

### DOT export + migrate

**DOT export does NOT emit `prompt_file` or `system_prompt_file`.** Mirrors #52: DOT emits inlined `prompt` / `system_prompt` only. Tracker reads from `.dipx` bundles where content has already been inlined by the pack-shadow pass.

**Migrate does NOT extract `prompt_file` or `system_prompt_file` from DOT attrs.** A `.dip → DOT → .dip` round-trip is lossy for the directive form (becomes inline). Documented in skill.md as an explicit non-goal, mirroring the `command_file:` round-trip trade-off.

`migrate/parity.go` adds `PromptFile` and `SystemPromptFile` to `compareAgentConfigs`'s field list to catch parity regressions. Two single-line additions; mirrors the v0.33 pattern.

### Pack-time validation

`dippin pack` follows the same flow as other CLI commands: parse → resolve → pack. If a referenced prompt file is missing, the resolver returns a clean error before pack writes the bundle. Authors get the error at pack time, not at runtime. Same UX as #52.

## Tracker-side design

**None.** This is a dippin-internal feature. Tracker reads from `.dipx` bundles where `Prompt` and `SystemPrompt` are already populated by the pack-shadow pass. Tracker sees the same IR shape it sees today.

No tracker coordination required. v0.34.0 ships dippin-only.

## Tests

### Parser (`parser/parse_nodes_test.go`)

| Case | Input | Expected `cfg.PromptFile` / `cfg.Prompt` | Expected diagnostic |
|---|---|---|---|
| explicit prompt_file | `prompt_file: prompts/p.md` | `PromptFile = "prompts/p.md"`, `Prompt = ""` | none |
| explicit inline prompt | `prompt: "hi"` | `PromptFile = ""`, `Prompt = "hi"` | none |
| both prompt set (inline first) | `prompt: x` then `prompt_file: y` | parsed up to error | parser diagnostic: "agent node has both `prompt` and `prompt_file`" |
| both prompt set (file first) | `prompt_file: y` then `prompt: x` | parsed up to error | same diagnostic |
| explicit system_prompt_file | `system_prompt_file: prompts/s.md` | `SystemPromptFile = "prompts/s.md"`, `SystemPrompt = ""` | none |
| explicit inline system_prompt | `system_prompt: "you are a..."` | `SystemPromptFile = ""`, `SystemPrompt = "you are a..."` | none |
| both system set (inline first) | `system_prompt: x` then `system_prompt_file: y` | parsed up to error | parser diagnostic: "agent node has both `system_prompt` and `system_prompt_file`" |
| both system set (file first) | `system_prompt_file: y` then `system_prompt: x` | parsed up to error | same diagnostic |
| cross-slot mix is fine | `prompt_file: y` + `system_prompt: x` | both populated independently | none |
| neither set | (no prompt at all) | `PromptFile = ""`, `Prompt = ""` | none |

### Resolver (`parser/resolve_test.go`, extending the existing file)

Uses real fixture files under `parser/testdata/`:

| Case | Setup | Expected |
|---|---|---|
| valid prompt resolve | fixture file at `parser/testdata/prompt.md` | `Prompt` populated with file bytes; no error |
| valid system_prompt resolve | fixture file at `parser/testdata/system.md` | `SystemPrompt` populated with file bytes; no error |
| both file directives resolve | both fixtures + agent uses both | both fields populated; no error |
| absolute path rejected (prompt) | `PromptFile = "/etc/passwd"` | error containing "absolute paths not allowed" |
| parent escape rejected (prompt) | `PromptFile = "../escape.md"` | error containing "resolves outside source directory" |
| symlink rejected (prompt) | `parser/testdata/link.md` is symlink → `/etc/passwd` | error containing "symlinks not allowed" |
| oversize rejected (prompt) | fixture file > 4MB | error containing "exceeds 4 byte limit" pattern |
| missing file (prompt) | `PromptFile = "nonexistent.md"` | error containing "not found" |
| user-path in errors (prompt) | any error case | error message contains the user-written path, NOT the resolved absolute path |
| error message includes directive name | resolver error on `system_prompt_file:` | error mentions `system_prompt_file`, not `prompt_file` |

The existing #52 resolver tests (for tool/`command_file`) remain unchanged and continue to pass — the security model is reused, so we don't need to duplicate the per-check coverage matrix. The agent-side tests cover the agent-specific branching (two fields per node, error-message directive identification) but trust the shared `loadDirectiveFile` helper.

### Integration (`validator/lint_examples_test.go::TestLintExamples`)

The new `examples/external_prompts.dip` workflow must parse cleanly via the standard test harness (which uses the parser without calling the resolver — so `cfg.Prompt == ""` and `cfg.SystemPrompt == ""` are expected for agent nodes using the file directives).

A companion CLI integration test loads `examples/external_prompts.dip` + its sibling files via a CLI entry point (which DOES call the resolver) and asserts that both `cfg.Prompt` and `cfg.SystemPrompt` are populated.

### Formatter round-trip (`formatter/format_test.go`)

- `agent X\n  prompt_file: prompts/p.md` → format → contains `prompt_file: prompts/p.md`
- `agent X\n  system_prompt_file: prompts/s.md` → format → contains `system_prompt_file: prompts/s.md`
- `agent X\n  prompt_file: y\n  system_prompt: "hi"` → format → both lines present, neither converted
- `agent X\n  prompt: "hi"` → format → contains `prompt: "hi"` (inline unchanged)
- **Regression (pre-existing bug):** `agent X\n  system_prompt: "you are a..."` → format → contains `system_prompt: "you are a..."` (inline preserved — was previously silently dropped)

### Parity (`migrate/parity.go` + existing parity tests)

Existing parity tests pass after `PromptFile` and `SystemPromptFile` are added to `compareAgentConfigs`'s field list.

### Pack-shadow (`cmd/dippin/pack_shadow_test.go` or equivalent)

- Pack a workflow with `prompt_file:` → inspect shadow `.dip` → assert inline `prompt:` block contains the file's contents and the `prompt_file:` line is absent.
- Pack a workflow with `system_prompt_file:` → same assertion for `system_prompt:`.
- Pack a workflow with all three directive families → assert all inlined, all `*File` lines absent.

## Example (`examples/external_prompts.dip` + sibling files)

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

```md
<!-- examples/external_prompts/reviewer-persona.md -->
You are a senior code reviewer. Be concise, direct, and constructive.
Focus on correctness, security, and maintainability over style.
```

```md
<!-- examples/external_prompts/reviewer-task.md -->
Review the diff at .ai/scratch/diff.patch.
Report findings ranked by severity (critical, important, minor).
End with STATUS: success or STATUS: fail.
```

The example demonstrates the canonical use case: an agent with both a persona-style system prompt and a task-specific user prompt, each in its own external file. Mirrors the `examples/external_files.dip` shape from #52.

## Release coordination

### PR sequence

1. Dippin PR opens against `main`. Pre-merge: file the deferred follow-up issues (defaults-block support, bundled-files redesign) with numbered titles so they're referenced from this spec text.
2. PR reviewed + approved.
3. Bump CHANGELOG date + tag dippin v0.34.0.

No tracker coordination needed. Tracker continues to read `.dipx` bundles where content is already inlined by the pack-shadow pass.

### Version

Dippin: v0.33.0 → **v0.34.0**. Minor bump (new authoring-surface directives).

### Doc updates

- **`CHANGELOG.md`** — `## [v0.34.0]` entry. Sections:
  - **Added**: `prompt_file:` and `system_prompt_file:` directives on agent nodes; resolver and pack-shadow extended to walk agent nodes; `examples/external_prompts.dip`.
  - **Notes**: parser stays pure; resolver and pack-shadow share machinery with `command_file:`; LSP and WASM see the unresolved IR view.
- **`docs/nodes.md`** — add `prompt_file` and `system_prompt_file` rows to the agent-node fields table.
- **`docs/llm-reference.md`** — add both fields to the agent optional-fields list.
- **`docs/GRAMMAR.ebnf`** — add `"prompt_file" ":" field_value` and `"system_prompt_file" ":" field_value` productions alongside `prompt` / `system_prompt`.
- **`site/static/skill.md`** — new `prompt_file:` and `system_prompt_file:` field rows. New subsection covers:
  - Purpose: replace inline `prompt:` / `system_prompt:` heredocs with external file refs
  - Path resolution and security: same model as `command_file:` (cross-link to that subsection)
  - Pack-time loading: `dippin pack` bundles inlined content into `.dipx`
  - Non-goals: defaults-block support, bundled-files `.dipx` (link follow-ups)
- **No `docs/validation.md` change** — no new DIP code (mutual-exclusion is a parser-time error, not a lint).
- **`lsp/completion.go`** — +2 entries (`prompt_file:`, `system_prompt_file:`).
- **`editors/vscode/syntaxes/dippin.tmLanguage.json`** — add `prompt_file` and `system_prompt_file` to the field-name regex alternation (same line as `command_file`).

Auto-handled (no change needed): tree-sitter grammar, Zed highlights, site `highlight.js`, WASM playground (all use generic field-name patterns).

## Follow-up issues to file BEFORE merge

(Mirrors the #52 "task #0" pattern that broke the DIP28 anti-pattern of unfilled "follow-up if needed" promises.)

1. **Defaults-block support for `prompt_file:` / `system_prompt_file:`.** v1 is per-node only, matching `command_file:`. When a workflow needs a shared external persona doc across agents, expose the file-form in `defaults agent`.
2. **Bundled-files `.dipx` redesign.** Replace the pack-shadow inline-into-`.dip` mechanism with bundled-as-files semantics: pack copies referenced scripts/prompts into the `.dipx` preserving relative paths; tracker (or any dipx consumer) calls `parser.ResolveFileDirectives` against the unpacked bundle dir. Trade-offs documented in § Non-goals #2. This is the more elegant end state but requires tracker coordination and a `.dipx` format version bump; deferred so v0.34 stays dippin-only.

## Complexity budgets

All new functions ≤ 5 cyclomatic, ≤ 7 cognitive per CLAUDE.md.

| Function | Estimated cyclomatic | Notes |
|---|---|---|
| `resolveNodeDirective` (extended switch) | 3 | type-switch on two cases + default-nil |
| `resolveAgentDirective` | 4 | two `loadInto` calls + assign + return |
| `loadInto` | 3 | early-return on empty/conflict + load + assign |
| `clearFileDirectives` (extended) | 5 | type-switch + per-case clears; at the cap |
| Parser `prompt_file:` case | 3 | assign + call helper |
| Parser `system_prompt_file:` case | 3 | assign + call helper |
| Parser `prompt:` case (extended) | +1 | call helper after existing assign |
| Parser `system_prompt:` case (extended) | +1 | call helper after existing assign |
| `checkPromptFileConflict` | 3 | early-return + check + emit |
| `checkSystemPromptFileConflict` | 3 | same shape |
| Formatter `prompt_file:` branch | 3 | branch + emit (two of these) |

If `clearFileDirectives` trips `just complexity`, extract the agent branch to its own `clearAgentFileDirectives` helper — parent drops to 3, helper to 4.

## Design journey

#52 deferred prompt directives explicitly: "Prompt variants ship when a workflow with the analogous prompt-bloat pain surfaces." That moment is now — the directive-family asymmetry is on the radar and authors are starting to refactor large prompts. Shipping the symmetric pair closes the loop.

Brainstorming surfaced two non-trivial decisions:

1. **Defaults-block scope:** per-node only, matching #52. The inline form (`defaults agent system_prompt: ...`) already covers the shared-config use case; the file-form-in-defaults is a v2 polish for a use case that hasn't surfaced.
2. **Bundled-files `.dipx` redesign:** considered and explicitly deferred. The current inlining model (introduced by #52 because dipx couldn't import formatter) is structurally less elegant than bundling referenced files alongside `workflow.dip` and resolving at unpack time. But the bundled-files model requires tracker coordination and a format version bump, and the user-visible behavior is identical (referenced content reaches tracker either way). The redesign is filed as a follow-up so v0.34 stays dippin-only and ships fast.

Everything else mechanical: copy the #52 model field-by-field to two new agent fields, extend the existing resolver dispatcher with an agent branch, extend `clearCommandFileDirectives` → `clearFileDirectives` with an agent branch. The simplification is that there is no simplification needed — the work shipped in #52 was already the right shape for this extension.
