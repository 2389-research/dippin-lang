# Issue #56 — `last_response_truncate:` carry-only mitigation attribute

**Date:** 2026-06-10
**Issue:** [#56](https://github.com/2389-research/dippin-lang/issues/56) — chain-attack mitigation, follow-up deferred from #41 (`safety-follow-up`, P2).
**Status:** Designed — the **mitigation** half of #56. DIP147 (PR #115) already ships the
explicit-key chain-attack *detection*; this adds the author-time mitigation attribute the runtime
enforces. Does **not** close #56 (the `${ctx.last_response}` auto-injection topology and cross-file
chains remain follow-ups).

## Background

`tool_access: none` bounds an agent's **tools**, not the **information flow** between agents. A
restricted summarizer that processes untrusted content can launder an injection payload into a
downstream tool-bearing agent's prompt via the auto-injected previous response (`${ctx.last_response}`)
or a named context key. DIP147 (a Hint) already *flags* the explicit-key vector. This spec adds the
**mitigation** dippin carries: `last_response_truncate:` — an author-time cap on how much of the
(potentially tainted) previous response a runtime injects into an agent's prompt.

Per the never-gate-on-runtime principle (issue #92/#93 precedent), dippin **carries + lints** the attribute; a downstream runtime
**enforces** the truncation. The attribute is **inert until a runtime reads it**. dippin's behavior is
not gated on runtime readiness.

### Explicitly out of scope (and why)

- **No `${ctx.last_response}` auto-injection *topology* lint.** A bare `none → full` edge warning is
  [#57](https://github.com/2389-research/dippin-lang/issues/57) — **rejected/deferred** ("the rejected
  in-file graph-topology lint"). It assumes a runtime-specific data-flow model. This spec does not
  resurrect it.
- **No DIP147 interaction.** A sink carrying `last_response_truncate` **still emits DIP147**.
  Truncation bounds payload *size*, not the *existence* of a laundered information flow — a tiny
  malicious payload fits within any cap. Suppressing DIP147 here would be fail-open. The attribute is
  a runtime knob dippin carries; DIP147 remains the honest advisory.
- **No graph-default form.** A workflow-wide `defaults: { last_response_truncate: N }` is a possible
  later follow-up, not this slice.
- **No "dead config" heuristic** (e.g. set on an entry node with no upstream). Out of scope; only the
  negative-value check ships.

## Semantics

`last_response_truncate: N` on an **agent node**: "cap the auto-injected previous response in *this*
agent's prompt to **N Unicode characters (runes)**." The agent being protected owns its own bound.

- **Value:** non-negative integer. Unit = **characters (runes)** — tokenizer-free and deterministic
  (dippin has no model tokenizer at author/lint time).
- **`0` / unset = no truncation** (the previous response is injected in full).
- **Parallel-branch override:** `BranchConfig.LastResponseTruncate` is a per-branch override of the
  target agent's value. Branch **`0` = inherit** the target agent's value (never resets to
  "no truncation") — mirroring how branch `tool_access`/`writable_paths` empty = inherit. The runtime
  resolves effective = branch if `> 0` else agent.

## Carry path

Mirrors `tool_access` (string) and `ToolConfig.OutputLimit` (int) exactly.

| Layer | Change |
| --- | --- |
| **IR** (`ir/ir.go`) | `AgentConfig.LastResponseTruncate int`; `BranchConfig.LastResponseTruncate int` (with the inherit-on-0 doc comment). |
| **Parser** (`parser/parse_nodes.go`) | Agent field switch: `case "last_response_truncate": cfg.LastResponseTruncate = p.parseInt(val, key, loc)`. Branch setter map: `"last_response_truncate": func(b *ir.BranchConfig, v string){ b.LastResponseTruncate = <parsed int> }`. |
| **Formatter** (`formatter/format.go`) | Agent + branch: emit `last_response_truncate: N` when `> 0`. |
| **DOT export** (`export/dot.go`) | `applyAgentRuntimeAttrs`: `if cfg.LastResponseTruncate > 0 { attrs["last_response_truncate"] = strconv.Itoa(...) }` (mirrors `output_limit`). Plus the DOT-import read-back so it round-trips. Branch attrs follow the existing branch-attr emission. |
| **Migrate** (`migrate/parity.go`) | `compareAgentBehavior`: add a `last_response_truncate` field diff mirroring the `tool_access` diff; branch parity mirrors `BranchWritablePathsDiff`. |

## DIP148 — negative `last_response_truncate` (new Warning)

Fires only when `LastResponseTruncate < 0` on an agent **or** a branch override. `0`/unset/positive =
silent. Structurally mirrors **DIP145** (negative graph budget default).

- **Severity:** Warning.
- **Message (shape):** `agent "Act" last_response_truncate is -1; cannot be negative` (and the branch
  analogue). `0` or unset = no truncation.
- **Edit sites** (per the lint-code edit-site checklist):
  - `validator/lint_codes.go`: `DIP148` const + `CodeDescription[DIP148]`.
  - `validator/explanations.go`: fully-populated `Explanations[DIP148]` (Summary/Trigger/Fix/Example,
    all non-empty) — `TestExplanationsCoverAllCodes`/`NoExtra` require this atomically with the const.
  - New file `validator/lint_last_response_truncate.go` (separate from `lint_budget.go` for clarity).
  - Describe DIP148 in `docs/validation.md` (DIP-by-DIP section) and append it to the
    `docs/llm-reference.md` warning-list prose (line ~193).
  - Bump the convention-only range strings `DIP101–DIP147` → `DIP101–DIP148` (predominantly en-dash;
    watch for hyphen too) across: `docs/validation.md` (4×), `docs/llm-reference.md`, `docs/cli.md`
    (incl. the "All 56 diagnostic rules" count → 57), `docs/integration.md`, `docs/architecture.md`,
    `docs/editor-setup.md`, plus any in-code comments in `validator/lint_codes.go` / `lint.go`.
  - `cmd/dippin/generated-spec.md` is **assembled** from `docs/llm-reference.md` + `site/static/skill.md`
    by `scripts/gen-spec.sh`; if the gen-spec source docs change, re-run it (the pre-commit hook does,
    and `releasecheck` gates freshness).
  - **`site/content/validation.md` is left for the next release's `docs(site)` pass** (per the
    #100→#107 / #102→#108 precedent — code + detection now, site surfacing at release).

## TDD build sequence (failing test first at each step)

1. **IR + parser** — `parser/*_test.go` (parser-driven): assert `last_response_truncate: 4096` lands on
   the agent's `LastResponseTruncate`; branch form lands on the branch.
2. **Formatter** — round-trip test: parse → format → re-parse preserves the value (agent + branch);
   `0`/unset emits nothing.
3. **DOT export + import** — `export/dot_test.go`: output contains `last_response_truncate="4096"`;
   import read-back round-trips.
4. **Migrate parity** — `migrate/parity_test.go`: differing `last_response_truncate` (agent + branch)
   produces a diff.
5. **DIP148** — `validator/lint_last_response_truncate_test.go` (parser-driven, reusing
   `lintSrc`/`hasCode`/`codes`): negative fires DIP148; `0`/unset/positive silent. Plus the
   explanation-parity tests stay green.
6. **DIP147 non-interaction** — a regression test asserting a sink with `last_response_truncate` set
   **still** emits DIP147 (no suppression).

## Verification

- `just test`, `just test-pkg <pkg>`, `just lint-go`, `just complexity` (cyclo ≤ 5 / cognit ≤ 7),
  `just fmt`, `just spec-check`.
- Pre-commit hook (the real CI gate; `just check` fails locally on tree-sitter-generate per
  the tree-sitter-generate gotcha).
- A code-review pass before the PR.

## Tracker follow-up

File the runtime-enforcement follow-up: a paired runtime must read `last_response_truncate` and cap the
auto-injected previous response (per-agent, branch override resolves first). Until then the attribute is
carried + linted but inert.
