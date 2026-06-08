# DIP010 — unparseable edge condition (issue #98)

**Status:** design approved (post expert-panel review)
**Branch:** `fix/98-unparseable-edge-condition`
**Date:** 2026-06-05

## Problem

`validator.Lint()` discards the error from `simulate.EnsureConditionsParsed()`
(`validator/lint.go:20`, `_ = simulate.EnsureConditionsParsed(w)`), and that
function early-returns on the **first** unparseable edge condition
(`simulate/condition.go:269-281`). Consequences:

1. An invalid `when …` edge condition (e.g. `A -> Z when marker_grep "^ok"`,
   easy to write because `marker_grep` is a real tool-node field) produces **no
   diagnostic at all**. `dippin validate`/`lint`/`check`/`doctor` greenlight a
   file that hard-fails at `dippin simulate` and at runtime
   (`edge A -> Z: invalid condition "marker_grep \"^ok\"": unknown operator "^ok"`).
2. **Cascade masking:** the early-return leaves every edge *after* the bad one
   with `Condition.Parsed == nil`. Lints that read the parsed AST then silently
   skip those edges.

### Corrected premise (from empirical panel review)

The issue claims DIP101/102/103/120 are masked. That is **wrong for DIP101/102**:
`hasUnconditionalEdge`/`allEdgesConditional`/`hasMissingDefault`
(`lint_reachability.go:92,113,156`) test only `Condition == nil` / `!= nil` —
they fire on *presence*, never reading `.Parsed`, so they are **not** masked.
The lints actually masked are the `.Parsed`-dependent ones:

- **DIP103** `lint_conditions.go:53`
- **DIP120** `lint_conditions.go:126`
- **DIP121** `lint_condition_types.go:49`
- **DIP122** `lint_condition_types.go:115`
- exhaustiveness/auto-suppression (`extractEqualityCondition`
  `lint_reachability.go:312`, `hasComplementaryPair` `:357`)

Empirically verified: a valid graph emits DIP120; inserting an earlier bad edge
makes **DIP120 vanish**. This is the discriminating behavior for the cascade test.

## Decisions (with rationale)

| Decision | Choice | Why |
|---|---|---|
| Code | **DIP010**, structural, `SeverityError` | Execution-blocking (undefined routing). DIP010 is unused; structural range is DIP001–DIP009. Mirrors DIP003 (unknown node ref is structural though "about" an edge). |
| Where emitted | **`validator.Validate()`** | Run by every command path (validate/lint/check/watch/simulate/pack/wasm/doctor/test/lsp). One check, caught everywhere. Sidesteps the four-CLI-paths gotcha. |
| Accumulate-all | **Validator parses edges itself** via `simulate.ParseCondition` (option c) | Only approach yielding one diagnostic *per* bad edge. Leaves `EnsureConditionsParsed` + its 6 callers **untouched**. No `simulate` API change; wasm-safe (`ParseCondition` is pure string→AST). |
| Scope | **Edges only** | Matches issue. `manager_loop` stop/steer conditions have the same swallowed-error class but no lint reads their `.Parsed` (manager-loop lints use Raw-based `conditionPresent`) — separate follow-up. |
| Doctor | **Exit non-zero on `SeverityError`** | User decision. `doctor` becomes a gate, not just a report. Note: this now makes doctor exit non-zero for *any* error (incl. existing DIP001–009), so `unhealthy.dip`'s DIP004 flips `TestCmdDoctor_WithSuggestions` to `ExitError`. |

## Implementation

### 1. `validator/validate_conditions.go` (new file)

Isolates the single `simulate` import to one file (keeps `validate.go` `ir`-only).
File-level doc comment states *why* it's separate.

```go
package validator

import (
	"fmt"
	"github.com/2389-research/dippin-lang/ir"
	"github.com/2389-research/dippin-lang/simulate"
)

// edgeNeedsParse mirrors simulate.ensureEdgeConditionParsed's guard
// (simulate/condition.go:285): skip nil / already-parsed / empty conditions.
func edgeNeedsParse(e *ir.Edge) bool {
	return e.Condition != nil && e.Condition.Parsed == nil && e.Condition.Raw != ""
}

// edgeParseFailure pairs an edge with its condition parse error.
// (Not named *Error — it does not implement the error interface.)
type edgeParseFailure struct {
	edge *ir.Edge
	err  error
}

// parseEdgeConditions parses every edge condition that needs it, populating
// Condition.Parsed for each that succeeds (idempotent; guarded by Parsed!=nil).
// Unlike simulate.EnsureConditionsParsed it does NOT stop at the first failure —
// every parseable edge is populated, so one bad condition cannot mask the
// AST-dependent lints (DIP103/120/121/122) on later edges. Returns one failure
// per unparseable edge.
//
// NOTE: this mutates w (lazily populates Condition.Parsed), consistent with the
// codebase's lazy-population pattern (see CLAUDE.md "Key gotcha"). Idempotent.
func parseEdgeConditions(w *ir.Workflow) []edgeParseFailure {
	var failures []edgeParseFailure
	for _, e := range w.Edges {
		f := parseOneEdge(e)
		if f != nil {
			failures = append(failures, *f)
		}
	}
	return failures
}

// parseOneEdge parses a single edge's condition (mirrors checkEndpoint split
// idiom to stay under the cognitive-complexity ceiling).
func parseOneEdge(e *ir.Edge) *edgeParseFailure {
	if !edgeNeedsParse(e) {
		return nil
	}
	parsed, err := simulate.ParseCondition(e.Condition.Raw)
	if err != nil {
		return &edgeParseFailure{edge: e, err: err}
	}
	e.Condition.Parsed = parsed
	return nil
}

// checkEdgeConditions verifies DIP010: every edge condition is parseable.
func checkEdgeConditions(w *ir.Workflow) []Diagnostic {
	var diags []Diagnostic
	for _, f := range parseEdgeConditions(w) {
		diags = append(diags, Diagnostic{
			Code:     DIP010,
			Severity: SeverityError,
			Message:  fmt.Sprintf("edge %s -> %s: invalid condition %q: %v", f.edge.From, f.edge.To, f.edge.Condition.Raw, f.err),
			Location: f.edge.Source,
			Help:     "valid operators: = == != contains startswith endswith in (combine with and/or/not)",
		})
	}
	return diags
}
```

Wire into `Validate()` (`validate.go:28`-ish), after the existing structural checks:
```go
diags = append(diags, checkEdgeConditions(w)...)
```

### 2. `validator/lint.go` — cascade fix for standalone `Lint()`

Replace `_ = simulate.EnsureConditionsParsed(w)` with:
```go
// Populate edge condition ASTs accumulate-all (one bad edge must not mask the
// AST-dependent lints DIP103/120/121/122 on later edges). DIP010 (from Validate)
// reports the failures; here we only need the population side effect.
parseEdgeConditions(w)
// manager_loop node conditions: still populated by EnsureConditionsParsed in the
// common (no-bad-edge) case; behavior-preserving.
_ = simulate.EnsureConditionsParsed(w)
```

### 3. `cmd/dippin/cmd_doctor.go` — doctor exits non-zero on errors

```go
func (c *CLI) renderDoctorReport(r *doctor.Report) ExitCode {
	if c.Format == FormatJSON {
		if code := c.renderJSON(r); code != ExitOK {
			return code
		}
	} else {
		renderDoctorText(c.Stdout, r)
	}
	if r.Lint.Errors > 0 {
		return ExitError
	}
	return ExitOK
}
```

### 4. Registration (build gates — all mandatory unless noted)

- `validator/codes.go`: add `DIP010 = "DIP010"` const (after DIP009) + `CodeDescription[DIP010] = "edge condition cannot be parsed"` to the map literal (structural codes live here, NOT `lint_codes.go`'s `init()`).
- `validator/explanations.go`: add full 4-field `Explanations[DIP010]` (Summary/Trigger/Fix/Example) in the structural block.
- `validator/validate_test.go:655`: add `DIP010` to the hardcoded `TestCodeDescriptionCoverage` slice (consistency; parity tests already cover it).
- Update in-code range comments: `codes.go:3`, `diagnostic.go:3`, `validate.go:11` → DIP001–DIP010.

Parity is auto-enforced: `TestExplanationsCoverAllCodes`/`NoExtra`
(`explanations_test.go:5,19`) fail the build if DIP010 is in only one map.

### 5. Docs + generated spec

Bump count `55 → 56` codes / `54 → 55` documented sections and range
`DIP001–DIP009 → DIP001–DIP010` in hand-authored docs: `CLAUDE.md:85`,
`docs/validation.md` (`:3` counts, `:5,14,47,1191,1199` ranges + **new DIP010
subsection** after DIP009), `docs/llm-reference.md:190,192`, `docs/cli.md`,
`docs/architecture.md`, `docs/editor-setup.md`, `docs/integration.md`,
`README.md`, `site/content/*`, `site/static/skill.md`.

**Then regenerate:** `docs/llm-reference.md` + `site/static/skill.md` are gen-spec
*inputs* — run `./scripts/gen-spec.sh` and commit outputs `docs/generated-spec.md`
+ `cmd/dippin/generated-spec.md` (+ `site/static/llms-full.txt` if CI copies it).
`releasecheck_test.go:44` + CI fail on a stale-spec diff.

**Do NOT hand-edit:** `docs/generated-spec.md`, `cmd/dippin/generated-spec.md`,
`site/static/llms-full.txt` (generated); `docs/superpowers/**`,
`docs/build_dippin.dot`, `docs/evolution_report.md` (frozen/legacy records).

Update `CHANGELOG.md`.

## Tests (TDD — failing first; real parser-driven fixtures, never hand-set `.Parsed`)

Helpers: `lintSrc(t, src)` (`lint_writable_paths_test.go:10`), `hasCode`/`codes`
(`lint_tool_access_test.go:70,79`), `countCode` (`lint_subgraph_tool_access_test.go:12`),
`runCLI` (`main_test.go:15`), CLI template `cmd_validate_crossfile_test.go:11-33`.
Add a new **`validateSrc(t, src)`** helper mirroring `lintSrc` but calling
`Validate` (none exists; `validate_test.go` hand-builds IR — do not copy that).

1. **Repro** — `A -> Z when marker_grep "^ok"` → exactly one DIP010, `SeverityError`,
   `Location` set to the edge source.
2. **Cascade** (corrected) — verified fixture: an earlier unparseable edge
   (`B -> C when marker_grep "^ok"` → DIP010) and a LATER valid edge referencing
   an undefined namespace (`D -> E when badns.flavor = vanilla` → DIP120). Assert
   **both** DIP010 (Validate) and DIP120 (Lint) fire. Pre-fix DIP120 is suppressed;
   this is the discriminating assertion. (Asserting DIP101/102 here proves nothing.)
3. **Multiple bad edges** → `countCode(diags, DIP010) == 2`.
4. **All-valid** → no DIP010; existing behavior unchanged.
5. **`parseEdgeConditions` unit test** → good edges get `.Parsed != nil` past a bad
   edge; bad edges reported.
6. **CLI integration** — streams matter:
   - `dippin validate <repro>` → `ExitError` + `error[DIP010]` on **stderr**.
   - `dippin check --format text <repro>` → `ExitError` + DIP010 on **stdout**.
   - `dippin doctor <repro>` → `ExitError` (new); assert via `--format json`
     `lint.errors >= 1` and/or grade < A.
7. **Negative / scope** — an unparseable `manager_loop stop_condition` does **NOT**
   emit DIP010 (edges-only boundary).
8. **Parity** — DIP010 present in `CodeDescription` + `Explanations`.

Existing test to update: `TestCmdDoctor_WithSuggestions` (`main_test.go:2008`) →
expect `ExitError` (unhealthy.dip has DIP004). Watch `just validate-examples` —
empirically no example flips red, but re-confirm.

## Out of scope (noted follow-ups)

- `manager_loop` node-condition swallowed-error (same class; no lint reads node `.Parsed`).
- Adding a condition-parse pass to `just check`/CI (issue option 3).

## Risk notes

- **No new strictness:** DIP010 surfaces the *same* `ParseCondition` gate
  `dippin simulate` already enforces — any condition failing it already fails
  simulate today.
- **`Validate()` non-pure:** now lazily populates `Condition.Parsed` (idempotent,
  guarded) — documented, consistent with existing pattern.
- **DIP009 co-fire:** a duplicated bad edge emits both DIP009 (Raw-keyed) and
  DIP010 — both true, benign.
- **WASM green:** only adds `simulate.ParseCondition` (already imported).
