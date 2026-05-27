# v0.33.0 — `command_file:` directive on tool nodes (issue #52)

**Date:** 2026-05-27
**Closes:** [#52](https://github.com/2389-research/dippin-lang/issues/52)
**Status:** Design approved 2026-05-27
**Predecessor:** v0.32.0 (`tool_access:`) shipped 2026-05-27 as a joint dippin + tracker release. Issue #52 was filed during v0.32 work; the four-round squad simplification pattern from #41 informs this spec's tight scope.
**Research:** Three reviewer transcripts (YAGNI / pragmatism, parser+security, future-maintainer + integration) consulted while writing this spec — recorded the converged design below.

## Problem

`.dip` tool nodes inline shell scripts via multiline `command:` blocks. There is no way to reference an external file. The result on real workflows is heredoc bloat — tracker's `examples/build_product.dip` is 1212 lines, of which ~200 are heredoc-write helpers (`Setup` writes shell scripts to disk for downstream nodes to source). The "Setup writes helper, downstream `source` it" pattern is a workaround that bloats `Setup` and forces a runtime write step for content that could live in the repo.

## Design

One field. One scope. Parser-pure.

```dip
tool Setup
  command_file: scripts/setup.sh
```

`command_file:` references an external file relative to the `.dip` source directory. A separate resolver pass loads the file at workflow-load time (CLI entry points), populating the existing `Command` field. The parser itself stays pure — no FS I/O — so LSP and WASM contexts work unchanged.

## Non-goals (v1)

These are tracked as numbered follow-up issues, filed before merge (§ Release coordination):

1. **`prompt_file:` and `system_prompt_file:` directives.** The cited production pain (tracker's `build_product.dip` heredoc bloat) is entirely shell scripts in tool nodes. Prompt variants ship when a workflow with the analogous prompt-bloat pain surfaces.
2. **Configurable size cap.** v1 hardcodes 4MB. When someone has a legitimate larger file, file an issue with the use case and pick a real default; don't pre-emptively configure.
3. **Symlink full-chain resolution.** v1 uses `os.Lstat` to reject symlinks at the directive's path. Full `EvalSymlinks` resolution (catching symlinks deeper in the resolved path) is deferred.
4. **Glob support** (`command_file: scripts/*.sh`). Composition semantics undefined; no current need.
5. **DOT round-trip preservation of `command_file:`.** DOT export emits inlined `command` only; pack-then-unpack loses the directive form. Spec documents the trade-off explicitly. Tracker reads from `.dipx` bundles where content is already inlined; no current consumer needs the path.
6. **Graceful LSP/WASM "file directive seen; not loaded in this context" message.** Authors using the playground or LSP see `cfg.Command == ""` for tool nodes with `command_file:` set. v1 lets this silently work for lint (which doesn't dereference `Command`); a future polish surfaces a clearer signal.

## Dippin-side design

### IR (`ir/ir.go`)

One new field on `ToolConfig`, clustered near `Command`:

```go
type ToolConfig struct {
    Command       string
    CommandFile   string  // Source path when Command was loaded from command_file:; empty if inline. See parser.ResolveFileDirectives.
    Timeout       time.Duration
    Outputs       []string
    MarkerGrep    string
    RouteRequired bool
    OutputLimit   int
}
```

`CommandFile` is the **primary** representation when an author uses the directive form. The parser populates `CommandFile` only; `Command` stays empty until the resolver runs. Downstream consumers that need execution-ready content call the resolver (CLI entry points do this automatically). Downstream consumers that lint or analyze structure (LSP, WASM playground) see `CommandFile != "" && Command == ""` and act accordingly — lint code that doesn't dereference `Command` works unchanged; lint code that does dereference needs a defensive null-check (none today need this).

This is intentionally not the `Condition.Parsed` gotcha pattern: `CommandFile` and `Command` are parallel fields representing different source forms, not the same field with two populations. Lint over `CommandFile` for "file form authored" semantics; lint over `Command` for "inline content present" semantics. A post-resolver invariant check could assert `CommandFile != "" → Command != ""` if needed in v2.

### Parser (`parser/parse_nodes.go`)

One new case in the tool-field handler (the function that today dispatches `command`, `outputs`, `marker_grep`, etc.):

```go
case "command_file":
    if cfg.Command != "" {
        p.diagnostics = append(p.diagnostics, diagPos(loc,
            "tool node %q has both `command` and `command_file` set; choose one", n.ID))
        return
    }
    cfg.CommandFile = val
```

And the symmetric check on the existing `command` case: if `cfg.CommandFile != ""`, emit the same diagnostic. Either ordering of the two directives in the source triggers the same error message.

**Parser stays pure** — no FS I/O. The new case just stores the path verbatim. Path security and file reading happen entirely in the resolver.

This is a parser-time error (untyped diagnostic, not a DIP code). The error fires immediately when the source has both directives — there's no semantic ambiguity to defer to lint, and we avoid a DIP catalog entry for a syntactic mistake.

### Resolver (`parser/resolve.go`, new file)

```go
package parser

import (
    "fmt"
    "os"
    "path/filepath"
    "github.com/2389-research/dippin-lang/ir"
)

const maxDirectiveFileSize = 4 << 20 // 4 MiB

// ResolveFileDirectives loads file contents for every tool node with
// CommandFile set, populating Command from the file's bytes. Paths are
// resolved relative to baseDir (typically the directory of the .dip
// source file). Returns the first error encountered.
//
// Called by CLI entry points after parse. LSP and WASM playground contexts
// skip this call — the IR retains CommandFile-set / Command-empty state,
// which is the correct unresolved view for those consumers.
func ResolveFileDirectives(w *ir.Workflow, baseDir string) error {
    for _, n := range w.Nodes {
        tc, ok := n.Config.(ir.ToolConfig)
        if !ok || tc.CommandFile == "" {
            continue
        }
        contents, err := loadDirectiveFile(baseDir, tc.CommandFile)
        if err != nil {
            return fmt.Errorf("node %q command_file: %w", n.ID, err)
        }
        tc.Command = string(contents)
        n.Config = tc
    }
    return nil
}

// loadDirectiveFile resolves p relative to baseDir, applies path security,
// and reads the file. Returns the file's bytes or a clean error.
// Error messages reference the user-written path (p), not the resolved
// absolute path, to avoid leaking CI directory structure into diagnostics.
func loadDirectiveFile(baseDir, p string) ([]byte, error) {
    if filepath.IsAbs(p) {
        return nil, fmt.Errorf("absolute paths not allowed: %q", p)
    }
    resolved := filepath.Join(baseDir, p)
    rel, err := filepath.Rel(baseDir, resolved)
    if err != nil || rel == "." || filepath.IsAbs(rel) || hasParentRef(rel) {
        return nil, fmt.Errorf("path %q resolves outside source directory", p)
    }
    info, err := os.Lstat(resolved)
    if err != nil {
        return nil, fmt.Errorf("path %q: %w", p, err)
    }
    if info.Mode()&os.ModeSymlink != 0 {
        return nil, fmt.Errorf("symlinks not allowed: %q", p)
    }
    if info.Size() > maxDirectiveFileSize {
        return nil, fmt.Errorf("file %q exceeds %d byte limit (size %d)", p, maxDirectiveFileSize, info.Size())
    }
    return os.ReadFile(resolved)
}

func hasParentRef(rel string) bool {
    for _, seg := range filepath.SplitList(filepath.ToSlash(rel)) {
        if seg == ".." {
            return true
        }
    }
    return false
}
```

Four security checks:

1. **Absolute paths rejected** — `filepath.IsAbs`
2. **Path stays under baseDir** — `filepath.Rel` returns a path that doesn't start with `..` and isn't absolute. The `hasParentRef` helper catches `..` segments after `Rel` (defensive belt; `Rel` should already prevent this on a non-symlink filesystem)
3. **Symlinks rejected** — `os.Lstat` mode check. Closes the realistic CI-untrusted-PR exposure (committed symlink → resolver reads /etc/passwd into IR → leaks via doctor JSON or error logs)
4. **Size cap 4 MiB** — `os.Stat` before read. Generous (prompt files can hit ~1MB legitimately; 4MB gives headroom for shell-script-with-vendored-data cases without enabling 50MB OOM accidents)

Error messages reference the **user-written path** (`scripts/setup.sh`), never the resolved absolute path. Prevents CI directory structure from leaking into PR review comments.

### CLI integration

Every CLI entry point that parses a `.dip` file calls `ResolveFileDirectives` after parse, with `baseDir = filepath.Dir(absoluteSourcePath)`. Touchpoints:

- `cmd/dippin/cli.go` (the main parse helper used by most subcommands)
- `cmd/dippin/cmd_fmt.go`, `cmd_migrate.go` if they have their own parse paths

The pattern:

```go
w, err := p.Parse()
if err != nil { return err }
if err := parser.ResolveFileDirectives(w, filepath.Dir(path)); err != nil {
    return err
}
```

LSP (`lsp/document.go`) and WASM (`cmd/wasm/main.go`) **do not** call the resolver. Their consumers don't need execution-ready content. This is the key architectural win — parser stays pure, so the contexts that can't do FS I/O don't have to error out.

### Formatter (`formatter/format.go`)

Conditional emission in the tool-node formatter:

```go
if cfg.CommandFile != "" {
    wr.line("command_file: %s", quoteValue(cfg.CommandFile))
} else if cfg.Command != "" {
    wr.multilineBlock("command", cfg.Command)
}
```

`dippin fmt` preserves the authored directive form. A `.dip` file with `command_file: scripts/setup.sh` round-trips to itself; a `.dip` file with inline `command:` round-trips to itself.

### DOT export + migrate

**DOT export does NOT emit `command_file`.** The DOT exporter only ever emits the inlined `command` attribute. Tracker reads from `.dipx` bundles where content has already been inlined by the CLI's resolver pass during `dippin pack`.

**Migrate does NOT extract `command_file` from DOT attrs.** A `.dip → DOT → .dip` round-trip is lossy for the directive form (becomes inline). This is documented in skill.md as an explicit non-goal. The use case for round-trip preservation is debugging tracker workflows; not justified for v1 (filed as follow-up).

`migrate/parity.go` adds `CommandFile` to `compareToolConfigs`'s field-list to catch parity regressions. The new field-equality check is a single line; mirrors the v0.32 pattern.

### Pack-time validation

`dippin pack` follows the same flow as other CLI commands: parse → resolve → pack. If a referenced file is missing, the resolver returns a clean error before pack writes the bundle. Authors get the error at pack time, not at runtime.

## Tracker-side design

**None.** This is a dippin-internal feature. Tracker reads from `.dipx` bundles where `Command` is already populated by the CLI's resolver pass during pack. Tracker sees the same IR shape it sees today.

No tracker coordination required. v0.33.0 ships dippin-only.

## Tests

### Parser (`parser/parse_nodes_test.go`)

| Case | Input | Expected `cfg.CommandFile` / `cfg.Command` | Expected diagnostic |
|---|---|---|---|
| explicit file | `command_file: scripts/setup.sh` | `CommandFile = "scripts/setup.sh"`, `Command = ""` | none |
| explicit inline | `command: echo hi` | `CommandFile = ""`, `Command = "echo hi"` | none |
| both set (inline first) | `command: x` then `command_file: y` | parsed up to error | parser diagnostic: "tool node %q has both `command` and `command_file`" |
| both set (file first) | `command_file: y` then `command: x` | parsed up to error | same diagnostic |
| neither set | (no command at all) | `CommandFile = ""`, `Command = ""` | none (separate concern: existing DIP for missing-command on tool nodes if any) |

### Resolver (`parser/resolve_test.go`, new file)

Uses real fixture files under `parser/testdata/`:

| Case | Setup | Expected |
|---|---|---|
| valid resolve | fixture file at `parser/testdata/setup.sh` | `Command` populated with file bytes; no error |
| absolute path rejected | `CommandFile = "/etc/passwd"` | error containing "absolute paths not allowed" |
| parent escape rejected | `CommandFile = "../escape.sh"` | error containing "resolves outside source directory" |
| post-Clean escape rejected | `CommandFile = "legit/../../../etc/passwd"` | same error |
| symlink rejected | `parser/testdata/link.sh` is symlink → `/etc/passwd` | error containing "symlinks not allowed" |
| oversize rejected | fixture file > 4MB | error containing "exceeds 4 byte limit" pattern |
| missing file | `CommandFile = "nonexistent.sh"` | error containing "no such file" (wrapped from os.Lstat) |
| user-path in errors | any error case | error message contains the user-written path (`"nonexistent.sh"`), NOT the resolved absolute path |

### Integration (`validator/lint_examples_test.go::TestLintExamples`)

The new `examples/external_files.dip` workflow must:
1. Parse cleanly via the standard test harness (which uses the parser without calling the resolver — so `cfg.Command == ""` is expected for the tool node using `command_file:`)
2. The companion test that DOES call the resolver (e.g., a new test in `parser/resolve_test.go` or a CLI integration test) loads `examples/external_files.dip` + its sibling `examples/external_files/setup.sh` and asserts `cfg.Command` is populated

### Formatter round-trip (`formatter/format_test.go`)

- `tool X\n  command_file: scripts/setup.sh` → format → contains `command_file: scripts/setup.sh`
- `tool X\n  command: echo hi` → format → contains `command: echo hi` (inline)

### Parity (`migrate/parity.go` + existing parity tests)

Existing parity tests pass after `CommandFile` is added to `compareToolConfigs`'s field-list.

## Example (`examples/external_files.dip` + `examples/external_files/setup.sh`)

```
workflow ExternalFiles
  goal: "Demonstrate command_file: directive (issue #52)"
  start: Setup
  exit: Verify

  tool Setup
    command_file: external_files/setup.sh

  agent Verify
    model: claude-sonnet-4-6
    prompt: "Confirm setup ran successfully. STATUS: success or STATUS: fail."
    auto_status: true

  edges
    Setup -> Verify
```

```sh
# examples/external_files/setup.sh
set -eu
echo "Running external setup script..."
mkdir -p .ai/scratch
echo "ready" > .ai/scratch/setup.done
```

The example demonstrates the canonical use case: an external shell script invoked by a tool node, replacing what would otherwise be a 5-10 line heredoc `command:` block.

## Release coordination

### PR sequence

1. Dippin PR opens against `main`. Pre-merge: file 6 follow-up issues with numbered titles (see § Follow-ups) so they're referenced from the spec text.
2. PR reviewed + approved.
3. Bump CHANGELOG date + tag dippin v0.33.0.

No tracker coordination needed. Tracker continues to read `.dipx` bundles where content is already inlined.

### Version

Dippin: v0.32.0 → **v0.33.0**. Minor bump (new authoring-surface directive).

### Doc updates

- **`CHANGELOG.md`** — `## [v0.33.0]` entry. Sections:
  - **Added**: `command_file:` directive on tool nodes; `parser.ResolveFileDirectives`; `examples/external_files.dip`.
  - **Notes**: parser stays pure; resolver invoked by CLI entry points; LSP and WASM see the unresolved IR view.
- **`docs/nodes.md`** — add `command_file` row to the tool-node fields table.
- **`docs/llm-reference.md`** — add `command_file` to the tool optional-fields list.
- **`docs/GRAMMAR.ebnf`** — add `"command_file" ":" field_value` production alongside `command`.
- **`site/static/skill.md`** — new `command_file:` field row + a short subsection covering:
  - Purpose: replace inline `command:` heredoc with external script ref
  - Path resolution: relative to `.dip` source directory; absolute paths rejected
  - Security: symlinks rejected, 4MB cap, parent-tree escape rejected
  - Non-goals: glob, prompt_file (link follow-up), DOT round-trip preservation (link follow-up)
  - Pack-time loading: `dippin pack` bundles the inlined content into `.dipx`
- **No `docs/validation.md` change** — there's no new DIP code (mutual-exclusion is a parser-time error, not a lint).
- **`lsp/completion.go`** — +1 entry (`command_file:`).
- **`editors/vscode/syntaxes/dippin.tmLanguage.json`** — add `command_file` to the field-name regex alternation (line 148).

Auto-handled (no change needed): tree-sitter grammar, Zed highlights, site `highlight.js`, WASM playground (all use generic field-name patterns or don't enumerate fields).

## Follow-up issues to file BEFORE merge

(Mirrors the #41 "task #0" pattern that broke the DIP28 anti-pattern of unfilled "follow-up if needed" promises.)

1. **Add `prompt_file:` and `system_prompt_file:` directives.** When workflow authors hit prompt-bloat analogous to the shell-script pain, ship the symmetric directives. Estimated v0.34 or v0.35.
2. **Configurable size cap for `*_file:` directives.** v1 hardcodes 4MB. If a legitimate use case exceeds it, expose a workflow-level or CLI-level config.
3. **Symlink full-chain resolution via `EvalSymlinks`.** v1 only rejects symlinks at the directive's path. A deeper symlink (in a parent directory that resolves elsewhere) could still escape. Tighten if the threat model requires.
4. **Glob support for `command_file:` patterns.** `command_file: scripts/*.sh` expansion with defined composition semantics.
5. **Preserve `command_file:` path through DOT round-trip.** Required if tracker (or any other consumer) wants source-path attribution in debugging output. Currently no consumer needs this.
6. **Graceful LSP/WASM "file directive seen; not loaded in this context" signal.** v1 lets these contexts silently render the unresolved IR. A future polish surfaces a clearer note in the LSP diagnostic stream and the playground UI.

## Complexity budgets

All new functions ≤ 5 cyclomatic, ≤ 7 cognitive per CLAUDE.md.

| Function | Estimated cyclomatic | Notes |
|---|---|---|
| `ResolveFileDirectives` | 4 | iterate nodes + type-assert + check empty + call helper + assign |
| `loadDirectiveFile` | 5 | 4 security checks + read; right at the cap |
| `hasParentRef` | 3 | iterate segments + check + return |
| Parser `command_file:` case | 3 | check Command, emit diagnostic, assign |
| Parser `command:` case (extended) | +1 | check CommandFile, emit if set |
| Formatter conditional | 3 | branch + emit |

If `loadDirectiveFile` trips `just complexity` after implementation, extract each security check to its own predicate helper (`isAbsolute`, `escapesBase`, `isSymlink`, `tooLarge`) — each becomes a 1-cyc helper, main function drops to 4.

## Design journey

Issue #52 was authored 2026-05-26 with a thorough body (proposed syntax, IR/parser touchpoints, security considerations, alternatives). The pre-spec design absorbed the issue body's recommendations and added the same kinds of rings the v0.32 spec went through four times — three IR fields for formatter preservation, four security layers stacked, three directives "for symmetry," DOT round-trip lossiness documented in a dedicated section.

A focused three-persona squad review (YAGNI/pragmatism, parser+security, future-maintainer + integration) cut scope aggressively:

- **v1 surface: 3 directives → 1 directive.** Cited pain is shell scripts, not prompts. Prompt variants ship when prompt-bloat surfaces.
- **IR fields: 3 → 1.** Parser pure (reviewer 2's architectural win) means `CommandFile` is the primary representation, not metadata. Other two prompt fields deferred with the directive deferral.
- **Security layers: 4 stacked → 4 distinct checks with one productive addition.** Symlink reject via `Lstat` closes a real CI-PR exposure that the original four layers missed.
- **DIP140 lint code: dropped.** Parser-time error for mutual exclusion is simpler and matches the syntactic-mistake category.
- **Parser stays pure.** Resolver pass runs at CLI entry points; LSP and WASM contexts skip it without special handling.
- **Editor/site integration surface: most auto-handle.** Only VSCode TextMate regex (line 148) and LSP completion need explicit updates. Tree-sitter, Zed, highlight.js, WASM all use generic field-name patterns.

The simplification was the design. Same lesson as #41.
