# Coverage redirection fix (issue #40) Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Ship the v0.30.0 fix for issue #40 — `dippin coverage` stops false-flagging file-redirected `echo`/`printf` as uncovered tool outputs by replacing the regex-based extractor with an AST walker built on `mvdan.cc/sh/v3/syntax`.

**Architecture:** Six sequential tasks. Task 1 creates the worktree. Task 2 implements the AST walker via TDD on `coverage/coverage.go` and `coverage/coverage_test.go`. Task 3 adds the example `.dip` file. Task 4 writes the CHANGELOG entry and runs `just check`. Task 5 opens the PR and monitors CI through merge. Task 6 tags `v0.30.0` and verifies the release.

**Tech Stack:** Go 1.x, `mvdan.cc/sh/v3/syntax` (already a project dependency), `just` for build automation, `gh` for PR/release operations. Pre-commit hook enforces golangci-lint + gocyclo ≤ 5 + gocognit ≤ 7 + gofmt + test-race + validate-examples.

---

## Conventions

- **Always use `just`.** `just check`, `just test-pkg coverage`, `just fmt`, etc. Never raw `go test`. (See CLAUDE.md.)
- **Tests must parse real `.dip` text** through the real parser. Hand-built IR is forbidden. `TestExtractToolOutputs` is an in-package unit test on raw shell strings — no parser involvement — and stays that way.
- **Complexity gates:** cyclomatic ≤ 5, cognitive ≤ 7 per function (excludes `_test.go`). Extract helpers; never add `//nolint`.
- **All work happens in the worktree from Task 1.** Subagent prompts must include the worktree path verbatim, and run `git log --oneline -3` to verify before committing.
- **Spec reference:** `docs/superpowers/specs/2026-05-21-issue-40-coverage-redirection-design.md`.

---

## Task 1: Worktree setup + baseline

**Files:** N/A (environment setup)

- [ ] **Step 1: Create the worktree via EnterWorktree**

Use the EnterWorktree tool with `name: "feat-issue-40-coverage-redirection"`. This branches from `origin/main` (default `worktree.baseRef = fresh`) and switches the session CWD into the worktree. Capture the path returned — every subsequent subagent prompt must include it.

- [ ] **Step 2: Verify worktree state**

Run: `git log --oneline -3 && git status`
Expected: tip is `8723909 docs(spec): add v0.30.0 coverage-redirection design for #40` (or whatever the latest origin/main commit is after the spec lands), working tree clean, branch `feat-issue-40-coverage-redirection`.

- [ ] **Step 3: Establish baseline**

Run: `just check`
Expected: all checks pass. If anything fails on clean origin/main, stop and investigate before touching code.

---

## Task 2: Replace regex extractor with AST walker

**Files:**
- Modify: `coverage/coverage.go` (replace `outputPatterns` + `extractToolOutputs`; add helpers)
- Modify: `coverage/coverage_test.go` (extend `TestExtractToolOutputs` table)

### Step 2.1: Write the failing test for the core bug

- [ ] Open `coverage/coverage_test.go`, locate `TestExtractToolOutputs` (around line 202), and replace the existing test-cases slice with the expanded version. Keep the existing rows; add ten new ones.

```go
func TestExtractToolOutputs(t *testing.T) {
	tests := []struct {
		name    string
		command string
		want    []string
	}{
		// Existing cases (regression coverage)
		{"single_printf", "printf 'pass'", []string{"pass"}},
		{"double_quote_printf", `printf "fail"`, []string{"fail"}},
		{"echo", "echo 'done'", []string{"done"}},
		{"printf_format", "printf '%s' 'value'", []string{"value"}},
		{"multiple", "printf 'a'\nprintf 'b'", []string{"a", "b"}},
		{"no_match", "run-binary", nil},
		{"dedup", "printf 'x'\nprintf 'x'", []string{"x"}},

		// Issue #40 — file-redirected statements must not extract
		{"redirect_overwrite", "echo 'foo' > log.txt\nprintf 'green'", []string{"green"}},
		{"redirect_append_and_pipe", "echo 'foo' >> log\necho 'bar' | tee /tmp/x\nprintf 'green'", []string{"green"}},
		{"stderr_redirect", "echo 'oops' >&2\nprintf 'real'", []string{"real"}},

		// Issue #40 — preserve format-two-arg pattern with newline format
		{"printf_format_newline", "printf '%s\\n' 'inline'\nprintf 'tail'", []string{"inline", "tail"}},

		// Issue #40 — echo with flags now extracts (regex missed it)
		{"echo_with_flags", "echo -n 'foo'", []string{"foo"}},

		// Issue #40 — command substitution: inner skipped, outer's only arg is non-literal
		{"echo_command_sub", "echo $(printf 'inner')\nprintf 'outer'", []string{"outer"}},

		// Issue #40 — subshell without redirect still writes to stdout
		{"subshell_unredirected", "( echo 'inside' )\nprintf 'outside'", []string{"inside", "outside"}},

		// Issue #40 — parse errors return nil (no regex fallback)
		{"malformed_shell", "echo 'unclosed", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractToolOutputs(tt.command)
			if !slicesEqual(got, tt.want) {
				t.Errorf("extractToolOutputs(%q) = %v, want %v", tt.command, got, tt.want)
			}
		})
	}
}
```

### Step 2.2: Run the test to verify the new cases fail

Run: `just test-pkg coverage`
Expected: the existing seven rows still pass; **at least** `redirect_overwrite`, `redirect_append_and_pipe`, `stderr_redirect`, and `echo_with_flags` FAIL. `malformed_shell` probably also fails (the regex extractor doesn't care about syntax). Confirm the failure output before proceeding.

### Step 2.3: Replace the regex extractor with the AST walker

- [ ] Open `coverage/coverage.go`. Make the following edits as one save:

1. Update the imports: remove `"regexp"`, add `"mvdan.cc/sh/v3/syntax"`. The file should now import: `sort`, `strings`, `github.com/2389-research/dippin-lang/ir`, `github.com/2389-research/dippin-lang/simulate`, `mvdan.cc/sh/v3/syntax`.

2. Delete the `outputPatterns` block (lines 48-54 — the `var outputPatterns = []*regexp.Regexp{ … }` declaration).

3. Replace the existing `extractToolOutputs` function (lines 145-159) and add helpers. The full replacement section (place where `outputPatterns` was, plus where `extractToolOutputs` was):

```go
// echoFlags are echo invocations that should be skipped when extracting
// literal output values from an echo call's arguments.
var echoFlags = map[string]bool{
	"-n": true, "-e": true, "-E": true, "--": true,
}

// extractToolOutputs finds printf/echo string literals in a shell command
// whose stdout reaches the engine. Statements that redirect to files or
// feed into pipes are skipped, as are echo/printf calls nested inside
// command substitution. Returns nil on shell parse errors — DIP123 catches
// syntax problems upstream.
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

// walkForOutputs returns a syntax.Walk callback that collects literal
// echo/printf args from statements whose stdout reaches the engine.
func walkForOutputs(outputs *[]string, seen map[string]bool) func(syntax.Node) bool {
	return func(node syntax.Node) bool {
		switch n := node.(type) {
		case *syntax.Stmt:
			return !stmtIsRedirected(n)
		case *syntax.CmdSubst:
			return false
		case *syntax.CallExpr:
			collectFromCall(n, outputs, seen)
		}
		return true
	}
}

// stmtIsRedirected returns true if a statement's stdout does not reach
// the engine — either via a file/fd redirect or a pipe.
func stmtIsRedirected(s *syntax.Stmt) bool {
	if len(s.Redirs) > 0 {
		return true
	}
	if bin, ok := s.Cmd.(*syntax.BinaryCmd); ok {
		return bin.Op == syntax.Pipe || bin.Op == syntax.PipeAll
	}
	return false
}

// collectFromCall extracts literal output args from an echo or printf call.
func collectFromCall(call *syntax.CallExpr, outputs *[]string, seen map[string]bool) {
	if len(call.Args) == 0 {
		return
	}
	name := extractWordLiteralLocal(call.Args[0])
	switch name {
	case "echo":
		collectEchoArgs(call.Args[1:], outputs, seen)
	case "printf":
		collectPrintfArgs(call.Args[1:], outputs, seen)
	}
}

// collectEchoArgs extracts literal args from `echo`, skipping flag args.
func collectEchoArgs(args []*syntax.Word, outputs *[]string, seen map[string]bool) {
	for _, w := range args {
		lit := extractWordLiteralLocal(w)
		if lit == "" || echoFlags[lit] {
			continue
		}
		addOutput(lit, outputs, seen)
	}
}

// collectPrintfArgs extracts literal args from `printf`. The two-arg
// `printf '%s' 'value'` and `printf '%s\n' 'value'` patterns extract the
// value; otherwise the format-string itself is the output literal.
func collectPrintfArgs(args []*syntax.Word, outputs *[]string, seen map[string]bool) {
	if len(args) == 0 {
		return
	}
	first := extractWordLiteralLocal(args[0])
	if first == "" {
		return
	}
	if isFormatTwoArg(first, args) {
		if val := extractWordLiteralLocal(args[1]); val != "" {
			addOutput(val, outputs, seen)
		}
		return
	}
	addOutput(first, outputs, seen)
}

// isFormatTwoArg returns true if printf was invoked with a bare `%s` or
// `%s\n` format string and exactly one value arg.
func isFormatTwoArg(first string, args []*syntax.Word) bool {
	return (first == "%s" || first == "%s\\n") && len(args) == 2
}

// addOutput appends a literal output value if it hasn't been seen and
// isn't itself a format specifier.
func addOutput(val string, outputs *[]string, seen map[string]bool) {
	if seen[val] || isFormatSpecifier(val) {
		return
	}
	seen[val] = true
	*outputs = append(*outputs, val)
}

// extractWordLiteralLocal returns the literal string content of a simple
// Word arg, handling bare, single-quoted, and double-quoted forms.
// Returns "" if the word contains expansions or substitutions.
func extractWordLiteralLocal(w *syntax.Word) string {
	if w == nil || len(w.Parts) != 1 {
		return ""
	}
	switch p := w.Parts[0].(type) {
	case *syntax.Lit:
		return p.Value
	case *syntax.SglQuoted:
		return p.Value
	case *syntax.DblQuoted:
		return extractDblQuotedLit(p)
	}
	return ""
}

// extractDblQuotedLit returns the literal content of a double-quoted
// word part, or "" if it contains expansions.
func extractDblQuotedLit(d *syntax.DblQuoted) string {
	if len(d.Parts) != 1 {
		return ""
	}
	lit, ok := d.Parts[0].(*syntax.Lit)
	if !ok {
		return ""
	}
	return lit.Value
}
```

4. Leave `isFormatSpecifier` unchanged.

### Step 2.4: Run tests, verify all green

Run: `just test-pkg coverage`
Expected: all rows in `TestExtractToolOutputs` pass, and the rest of the coverage tests still pass. If any row fails, examine the output and adjust — do not skip ahead.

### Step 2.5: Run complexity check

Run: `just complexity`
Expected: no cyclomatic ≥ 6 or cognitive ≥ 8 findings. Every new function above is designed to fit ≤ 5 / ≤ 7; if any exceeds, extract a helper rather than adding `//nolint`.

### Step 2.6: Run the full check suite

Run: `just check`
Expected: build, vet, golangci-lint, fmt, test-race, complexity, validate-examples all pass.

### Step 2.7: Commit the walker change

```bash
git add coverage/coverage.go coverage/coverage_test.go
git commit -m "$(cat <<'EOF'
fix(coverage): replace regex extractor with AST walker (#40)

Statements that redirect to files (>, >>, &>, >&) or feed into pipes
no longer have their echo/printf args counted as stdout outputs.
echo/printf calls nested inside command substitution are skipped.
Parse errors return nil (DIP123 catches syntax problems upstream).

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>
EOF
)"
```

Verify with `git log --oneline -3`.

---

## Task 3: Add `examples/coverage_redirection.dip`

**Files:**
- Create: `examples/coverage_redirection.dip`

### Step 3.1: Write the example

- [ ] Create `examples/coverage_redirection.dip` with the bug pattern from the issue body, wired to a complete workflow so `validate-examples` and `lint-examples` succeed:

```
workflow CoverageRedirection
  goal: "Demonstrate that file-redirected echoes no longer trip coverage"
  start: RunTests
  exit: Done

  tool RunTests
    timeout: 2m
    command: |
      set -eu
      echo 'capturing test baseline' > /tmp/coverage_redirection.log
      echo 'build_system=python' >> /tmp/coverage_redirection.log
      if true; then printf 'validation_pass'; else printf 'validation_fail'; fi

  agent Pass
    prompt: Summarize the successful run.

  agent Fail
    prompt: Summarize the failed run.

  agent Done
    prompt: Final report.

  edges
    RunTests -> Pass  when ctx.tool_stdout = validation_pass
    RunTests -> Fail  when ctx.tool_stdout = validation_fail
    Pass -> Done
    Fail -> Done
```

### Step 3.2: Verify the example lints clean

Run: `just validate-examples && just lint-examples`
Expected: both pass with no errors or warnings against `examples/coverage_redirection.dip`. The `RunTests` node should report `covered` (status), not `partial`, because the two routing markers (`validation_pass`, `validation_fail`) are the only stdout-reaching outputs after the redirect-aware extractor.

### Step 3.3: Spot-check the coverage result directly

Run: `go run ./cmd/dippin coverage examples/coverage_redirection.dip`
Expected: `RunTests` listed with status `covered`, no missing edges. The two file-redirected `echo` statements MUST NOT appear in any extracted-outputs section.

### Step 3.4: Commit the example

```bash
git add examples/coverage_redirection.dip
git commit -m "$(cat <<'EOF'
test(examples): add coverage_redirection.dip exercising #40 fix

Demonstrates that file-redirected echoes are no longer counted as
stdout outputs, so a tool node emitting only routing markers reports
'covered' instead of 'partial'.

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

## Task 4: CHANGELOG entry + final check

**Files:**
- Modify: `CHANGELOG.md` (add v0.30.0 section at the top)

### Step 4.1: Add the CHANGELOG entry

- [ ] Open `CHANGELOG.md`. Insert the following block after the file header / table-of-contents (if any) and before the existing `## [v0.29.0]` entry:

```markdown
## [v0.30.0] — 2026-05-21

### Fixed

- `dippin coverage` no longer flags file-redirected `echo`/`printf` statements as uncovered tool outputs. The extractor now uses an AST walker (`mvdan.cc/sh/v3/syntax`) and skips statements that redirect to files (`>`, `>>`, `&>`, `>&`), feed into pipes, or are nested inside command substitution. Pipelines using the "log to file, printf marker on stdout" pattern correctly report `covered` instead of `partial`. (#40)

### Closed

- #40 — coverage flags file-redirected echo/printf as uncovered outputs
```

Match the punctuation and section style of the v0.29.0 entry exactly.

### Step 4.2: Run the full check

Run: `just check`
Expected: every step passes. The pre-commit hook will regenerate `docs/generated-spec.md` from the updated CHANGELOG and source; if it does, stage that file with the CHANGELOG commit.

### Step 4.3: Commit the CHANGELOG

```bash
git add CHANGELOG.md docs/generated-spec.md
git commit -m "$(cat <<'EOF'
docs(changelog): v0.30.0 — coverage redirection fix (#40)

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>
EOF
)"
```

If `docs/generated-spec.md` didn't change, drop it from the `git add` and commit only `CHANGELOG.md`.

### Step 4.4: Sanity-check the branch state

Run: `git log --oneline origin/main..HEAD`
Expected: three commits — the AST walker, the example, the CHANGELOG.

---

## Task 5: Open PR + drive to merge

**Files:** N/A (GitHub operations)

### Step 5.1: Push the branch

```bash
git push -u origin feat-issue-40-coverage-redirection
```

### Step 5.2: Open the PR

```bash
gh pr create --title "fix(coverage): respect shell redirection in extractor (#40)" --body "$(cat <<'EOF'
## Summary

- Replace the regex-based `extractToolOutputs` in `coverage/coverage.go` with an AST walker built on `mvdan.cc/sh/v3/syntax` (already a project dependency, already used by `validator/lint_tool_helpers.go:extractBinary`).
- Statements that redirect to files (`>`, `>>`, `&>`, `>&`), feed into pipes, or nest `echo`/`printf` inside command substitution are skipped — they don't reach the engine's stdout.
- Add `examples/coverage_redirection.dip` exercising the bug pattern from issue #40 and verifying that the `RunTests` node reports `covered`.
- Tag v0.30.0 after merge.

## Test plan

- [ ] `just check` green locally
- [ ] CI green
- [ ] `examples/coverage_redirection.dip` lints clean and reports `covered`
- [ ] Spot-check on `2389-research/pipelines` PR #13 (the field-report source) to confirm `build-and-ship/refactor-express.dip` flips from A/90 to A/100

Closes #40.

🤖 Generated with [Claude Code](https://claude.com/claude-code)
EOF
)"
```

Capture the PR URL from the output.

### Step 5.3: Monitor CI to green

Run: `gh pr view --json state,statusCheckRollup,reviewDecision`
Expected: poll every 60–90s until `statusCheckRollup` shows all checks `SUCCESS`. If any check fails, fetch the logs (`gh run view --log-failed`), fix, push, re-run this step.

### Step 5.4: Handle review feedback

If CodeRabbit (or another reviewer) flags issues:
1. **Verify the claim independently** before acting. Per CLAUDE.md / project lore: CodeRabbit on the v0.29.0 PR falsely claimed the generated-spec was stale when it had actually been regenerated. Don't burn cycles re-running generators on false claims.
2. If the claim is real, fix in a new commit on the same branch, push, re-run Step 5.3.
3. If the claim is false, refute on-thread with a citation.

### Step 5.5: Merge

Once CI is green and any reviews are resolved:
```bash
gh pr merge --squash --delete-branch
```

Verify with `gh pr view --json state` — `state` should flip to `MERGED` (not just `reviewDecision: APPROVED`).

### Step 5.6: Update local main

```bash
git checkout main
git pull --ff-only origin main
git log --oneline -3
```
Expected: the squash commit from the PR is the new HEAD on `main`.

---

## Task 6: Tag v0.30.0 + verify release

**Files:** N/A (release operations)

### Step 6.1: Tag the release

```bash
git tag -a v0.30.0 -m "v0.30.0: coverage redirection fix (#40)"
git push origin v0.30.0
```

### Step 6.2: Wait for GoReleaser to publish

Run: `gh run watch` (or `gh run list --workflow=release.yml --limit=1`) and wait for the GoReleaser-triggered workflow to complete.

### Step 6.3: Verify the release

Run: `gh release view v0.30.0`
Expected: release page exists with cross-platform binaries listed. The Homebrew tap should also have updated (`brew info 2389-research/tap/dippin-lang` or the tap repo's commit log).

### Step 6.4: Exit the worktree

Use the ExitWorktree tool with `action: "remove"`. The branch was already deleted in Step 5.5; ExitWorktree removes the worktree directory and returns the session to the original CWD.

### Step 6.5: Confirm with the user

Report:
- PR URL (from Step 5.2)
- Tag URL (`gh release view v0.30.0 --web --json url` or construct from `https://github.com/2389-research/dippin-lang/releases/tag/v0.30.0`)
- Brief confirmation that `just check` was green and CI passed

---

## Self-review

**Spec coverage check:**
- Spec §"Detection algorithm" → Task 2 (Steps 2.3, lines covering `extractToolOutputs`, `walkForOutputs`, `stmtIsRedirected`, `collectFromCall`, `collectEchoArgs`, `collectPrintfArgs`).
- Spec §"Edge cases (frozen decisions)" → Task 2 Step 2.1 (test table rows: `redirect_overwrite`, `redirect_append_and_pipe`, `stderr_redirect`, `echo_command_sub`, `subshell_unredirected`, `printf_format_newline`, `echo_with_flags`, `malformed_shell`). One spec row not explicit-tested: `( echo 'foo' ) > file` (subshell with outer redirect) — implicitly covered because `stmtIsRedirected` returns true at the outer Stmt level; if a future bug suggests otherwise, add a row.
- Spec §"Fallback on parse error" → Task 2 (extractor returns nil on `parser.Parse` error; `malformed_shell` test row verifies).
- Spec §"Walker placement" → Task 2 (helpers added in `coverage/coverage.go`; no shared package).
- Spec §"Tests" → Task 2 Step 2.1 (table) + Task 3 (integration via example).
- Spec §"Docs" → Task 4 (CHANGELOG); skill.md and llm-reference.md not modified — they contain no coverage-extraction guidance to update.
- Spec §"Acceptance criteria" → each item maps to a verification step (regex import removal: Step 2.3 #1; AST walker: Step 2.3 #3; ten test rows: Step 2.1; example lints: Step 3.2; CHANGELOG: Step 4.1; `just check`: Step 4.2; v0.30.0 tag: Step 6.1).

**Placeholder scan:** No "TBD", "TODO", "implement later", or "similar to Task N" references. All code shown in full. All commands include expected output.

**Type consistency:** `extractWordLiteralLocal` defined once (Step 2.3), called from `collectFromCall`, `collectEchoArgs`, `collectPrintfArgs`. `addOutput` signature consistent across callers. `walkForOutputs` returns `func(syntax.Node) bool` matching `syntax.Walk`'s expectation. No mismatches.

**Spec drift check:** The spec's edge-case table mentions `( echo 'foo' )` (subshell, no redirect) extracting. Task 2 Step 2.1 includes `subshell_unredirected` (asserts `["inside", "outside"]`). ✓
