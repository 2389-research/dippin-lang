# Coverage extraction respects shell redirection (issue #40) — v0.30.0

**Date:** 2026-05-21
**Closes:** #40
**Release target:** v0.30.0
**Predecessor:** v0.28.0 shipped declared `outputs:` as the explicit escape hatch (`coverage/coverage.go:99-102` prefers `cfg.Outputs` over regex extraction). This batch fixes the auto-extraction so authors who don't declare `outputs:` get correct coverage on log-write patterns.

## Motivation

`coverage/coverage.go:48-53` matches every `printf '...'` and `echo '...'` string literal in a tool command body, regardless of whether that statement redirects to a file. Tool nodes that log to disk and emit only a final routing marker to stdout — a common, healthy pattern — drop from `A/100` to `A/90` in `dippin doctor` because the log lines are reported as missing edge targets.

Field report (from issue body, `pipelines` PR #13): `build-and-ship/refactor-express.dip`'s `SnapshotTests` node is clean post-TRK101 but flagged partial because file-redirected echoes are counted as stdout outputs they never actually become.

The escape hatch (declare `outputs:` explicitly) works, but authors of pre-existing `.dip` files shouldn't have to retrofit it for the extractor to do the right thing.

## Scope

Replace the regex-based `extractToolOutputs` in `coverage/coverage.go` with an AST walker built on `mvdan.cc/sh/v3/syntax` — already a project dependency, already used by `validator/lint_tool_helpers.go:extractBinary` with the same walker pattern. The walker visits `echo` and `printf` `CallExpr` nodes, rejects those whose statement redirects or feeds a pipe, and extracts literal-string args.

No new lint codes. DIP138 (reserved in v0.29.0 for "tool routes on stdout but declares no marker_grep/outputs") is a different concept — declaration coverage, not extraction accuracy — and stays reserved with no firing logic.

Out of scope:

- DIP138 firing logic (separate concern).
- Sharing the AST walker between `coverage/` and `validator/` packages — wait for the second consumer before extracting a `shellast` helper package.
- Heredocs, command substitution analysis beyond "skip if the echo/printf is nested inside one," and other extraction enhancements not described in #40.
- Improving the `printf '%s\n' 'value'` extraction for arbitrary format specifiers. The existing pattern in `outputPatterns` only handles a literal `%s`; the new walker mirrors that scope.

## Design

### Detection algorithm

In `coverage/coverage.go`, replace the regex-based extractor with an AST-driven one. Pattern follows `validator/lint_tool_helpers.go:extractBinary` line-for-line:

```go
func extractToolOutputs(command string) []string {
    parser := syntax.NewParser(syntax.KeepComments(false))
    prog, err := parser.Parse(strings.NewReader(command), "")
    if err != nil {
        return nil
    }
    var outputs []string
    seen := make(map[string]bool)
    syntax.Walk(prog, walkForOutputs(&outputs, seen))
    return outputs
}
```

The walker visits AST nodes top-down. At each node it decides:

1. **`*syntax.Stmt`** — if the statement has any `Redirs` entries (`>`, `>>`, `&>`, `>&`, etc.), or its `Cmd` is a `*syntax.BinaryCmd` with a pipe operator, do not descend. The statement does not write to the engine's stdout.
2. **`*syntax.CmdSubst`** — do not descend. `$(...)` and backtick contents are captured by the enclosing word, not emitted to stdout.
3. **`*syntax.CallExpr`** — if the first arg is a literal `echo` or `printf`, extract the literal-string args (see below). Continue walking to find further `echo`/`printf` calls inside compound commands (`if`, `while`, `for`, `{ … }`, subshells, etc.).

For `echo` calls: skip flag args (`-n`, `-e`, `-E`, `--`); extract every remaining `extractWordLiteral`-resolvable arg.

For `printf` calls: if there are exactly two args and the first is a literal `%s` (or `%s` followed by `\n`, matching the existing third regex pattern), extract the second arg. Otherwise extract the first arg (the format string itself, when used as a plain literal — matches the original first/second regex patterns). Skip pure format-specifier values via the existing `isFormatSpecifier` helper.

Dedup via the `seen` map, preserving first-occurrence order.

### Why AST over regex post-filter

The issue body proposes both. The regex post-filter (scan for `>`, `>>`, `|` in the same statement) is ~5 lines but inherits the original regex's blind spots:

- `printf '%s > %s' 'foo' 'bar'` triggers the heuristic and under-extracts.
- `echo -n 'foo'` is missed entirely by the original regexes (the `-n` breaks the `echo\s+'...'` match) and stays missed under the post-filter.
- Command substitution (`echo $(printf 'inner')`) over-extracts under the regex approach because both calls match.

The AST approach is the established pattern in this codebase (`extractBinary` already does it for the same reason), and `mvdan.cc/sh/v3` is already a dependency — the marginal cost is zero.

### Edge cases (frozen decisions)

| Snippet | Walker decision | Rationale |
|---|---|---|
| `echo 'foo' > log` | Skip | Stmt has `Redirs` |
| `echo 'foo' >> log` | Skip | Append redirect |
| `echo 'foo' \| tee log` | Skip | Pipe — conservative; `tee` does echo back to stdout, but authors who need the marker recognized declare `outputs:` |
| `echo 'foo' >&2` | Skip | stderr fd-redirect (still in `Redirs`) |
| `echo 'foo' &` | Extract | Backgrounded; stdout still reaches engine |
| `echo $(printf 'inner')` | Extract `'foo'` from outer `echo`; skip inner `printf` (inside `CmdSubst`) | Inner captured by `$()`, not stdout |
| `echo -n 'foo'` | Extract `'foo'` | Skip flag args |
| `printf '%s\n' 'inline'` | Extract `'inline'` | Existing third-regex case, preserved |
| `printf 'literal'` | Extract `'literal'` | Existing first-regex case |
| `( echo 'foo' )` | Extract | Subshell still writes to stdout |
| Malformed shell | Return `nil` | Parser fails; DIP123 catches syntax errors upstream |

### Fallback on parse error

`mvdan.cc/sh/v3` is strict — `extractBinary` defends with `extractBinaryFallback`. For coverage, we deliberately do **not** fall back to the regex extractor. Rationale:

- DIP123 (`bash -n` syntax check) already lint-warns on unparseable commands. By the time `dippin coverage` runs on a workflow the author cares about, syntax errors are surfaced upstream.
- A coverage extractor that produces output for syntax-broken commands gives false confidence; better to return `nil` (status becomes `"unknown"` per `determineStatus` at `coverage.go:115-117`) and let the linter surface the actual problem.
- `extractBinary`'s fallback exists because DIP125 (`bin not on PATH`) needs to fire even on syntactically-questionable commands. Coverage has no such failure mode to preserve.

If a user reports that coverage now reports "unknown" where it used to extract from a broken command, the fix is to fix the command, not to revive the regex fallback.

### Walker placement

The walker logic lives in `coverage/coverage.go` next to `extractToolOutputs`. We do **not** extract a shared `shellast` helper package between `coverage/` and `validator/` in this PR. Reasons:

- Two consumers is the "rule of three minus one" — extract on the third.
- The validator's `walkForBinary` and coverage's `walkForOutputs` traverse the same AST but answer different questions. Their walk functions can share helper primitives (`extractWordLiteral`) but their decision logic should stay in their own packages.
- `validator` is an `analysis` package per CLAUDE.md's import-graph rules; pulling `coverage` into `validator` via a shared helper would tangle the layering.

`extractWordLiteral` (`validator/lint_tool_helpers.go:131-140`) is duplicated into `coverage/coverage.go` rather than imported. Three lines; not worth a shared package.

## Tests

In `coverage/coverage_test.go`, extend `TestExtractToolOutputs` with the following table rows:

1. **`redirect_overwrite`**: `echo 'foo' > log.txt; printf 'green'` → `["green"]`. The current bug. Must NOT contain `"foo"`.
2. **`redirect_append_and_pipe`**: `echo 'foo' >> log; echo 'bar' \| tee /tmp/x; printf 'green'` → `["green"]`.
3. **`printf_format_two_arg`**: `printf '%s\n' 'inline'; printf 'tail'` → `["inline", "tail"]`. Regression for the existing third-regex pattern.
4. **`echo_unredirected`**: `echo 'foo'` → `["foo"]`. Regression for the normal case.
5. **`echo_with_flags`**: `echo -n 'foo'` → `["foo"]`. The original regex missed this case entirely; the AST walker now handles it.
6. **`echo_command_sub`**: `echo $(printf 'inner'); printf 'outer'` → `["outer"]`. Inner `printf` is inside `CmdSubst` and must be skipped; outer `echo` extracts nothing (its only arg is the substitution, not a literal). Verify both behaviors.
7. **`stderr_redirect`**: `echo 'oops' >&2; printf 'real'` → `["real"]`.
8. **`subshell_unredirected`**: `( echo 'inside' ); printf 'outside'` → `["inside", "outside"]`.
9. **`malformed_shell`**: `echo 'unclosed` → `nil`. Parser error path.
10. **`dedup`**: `printf 'green'; printf 'green'` → `["green"]`. Existing `seen` map behavior.

Per CLAUDE.md, fixtures must run through the real parser. `TestExtractToolOutputs` is already a string-input unit test (no parser involvement); no changes to that pattern.

Integration coverage: add `examples/coverage_redirection.dip` exercising the bug pattern from the issue body (the `RunTests` snippet — file-redirected echoes + final routing printf). `TestLintExamples` and `just validate-examples` then guard against regressions. The example must lint clean — declare appropriate edge conditions for `validation_pass` / `validation_fail` so coverage reports `covered`, not `partial`.

## Docs

- `CHANGELOG.md` — new `## [v0.30.0] — 2026-05-21` entry. Sections: **Fixed** (coverage now correctly skips file-redirected echo/printf — issue #40). One entry; no `Added` or `Changed` sections for this release.
- `docs/llm-reference.md` — coverage-extraction section, if one exists. Note that file-redirected output statements are no longer counted as stdout. If no such section, skip.
- `site/static/skill.md` — same. Update only the coverage-related guidance, not the lint codes section (no DIP changes).
- `cmd/dippin/generated-spec.md` — regenerates on commit via the existing hook.

## Non-goals

- DIP138 firing logic.
- Heredoc / `cat <<EOF` support — extractor remains echo/printf only.
- Variable expansion analysis (`echo "$VAR"` → no extraction; current behavior preserved by the literal-only `extractWordLiteral`).
- Improving `printf` format-specifier coverage beyond the existing `%s` / `%s\n` cases.
- Runtime-side changes. No runtime coupling; this is a pure dippin-internal lint-correctness fix.

## Backwards compatibility

- The set of extracted outputs is strictly **smaller or equal** to today's set for any given command. Workflows that were `covered` stay `covered` (their edge conditions still cover the now-smaller output set). Workflows that were `partial` because of false-positive redirected lines flip to `covered` — the intended fix.
- Workflows that were `partial` for *real* coverage gaps stay `partial`. No false negatives are introduced.
- One observable change: `echo -n 'foo'` now extracts; previously it was silently missed. Workflows using `echo -n` for routing markers will start being analyzed correctly. Edge cases relying on the old behavior (no extraction → status `unknown` → no missing-edges report) lose their accidental free pass. This is a correctness improvement, not a regression.
- The escape hatch (`outputs: …` declared explicitly) is unchanged. Anyone who retrofitted it during v0.28.0 keeps it; the auto-extraction improvement just means fewer authors need to.

## Acceptance criteria

- [ ] `coverage/coverage.go` no longer imports `regexp` for output extraction (verify via grep).
- [ ] `extractToolOutputs` uses `mvdan.cc/sh/v3/syntax` with the walker pattern.
- [ ] All ten table rows in `TestExtractToolOutputs` pass.
- [ ] `examples/coverage_redirection.dip` lints clean and `dippin coverage` reports `covered`.
- [ ] CHANGELOG.md has the v0.30.0 entry.
- [ ] `just check` green (build, vet, fmt, test-race, complexity, validate-examples).
- [ ] Tagged as `v0.30.0` after merge.
