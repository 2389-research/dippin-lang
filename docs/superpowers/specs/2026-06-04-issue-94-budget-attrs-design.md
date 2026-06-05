# Design: declarable budget/limit attrs — `stall_timeout`, `max_turns` exhaustion, DIP145 (#94)

**Date:** 2026-06-04
**Issue:** Closes #94 (P2)
**Status:** Approved design (expert-panel reviewed) — ready for implementation plan

## Summary

Make runtime budget guards declarable in `.dip` and give `max_turns` defined
exhaustion semantics. dippin **carries + validates + lints** the authoring
surface; the tracker runtime **enforces** it (`pipeline.BudgetGuard` +
`pipeline/dippin_adapter.go` `extractWorkflowDefaults`). Every new field is
**inert** until the tracker reads it — but per the project's
`never-gate-dippin-on-tracker` principle we ship the surface now and do **not**
gate any dippin behavior on tracker readiness.

Four changes, one PR:

1. **New graph default `stall_timeout`** (`time.Duration`) — abort/route when no
   forward progress is made for a wall-clock span. Sibling of the three existing
   budget ceilings.
2. **`max_turns` exhaustion semantics** — hitting `max_turns` ends the node with
   outcome `fail`, routed through the existing #92/#93 failure cascade. **No new
   declarable action field** (declining the issue's "ideally configurable"
   `fail|truncate|fallback` enum as YAGNI). Documentation-only.
3. **Close a pre-existing DOT export data-loss bug.** The three existing budget
   fields (`max_total_tokens`/`max_cost_cents`/`max_wall_time`) are silently
   dropped on `.dip → DOT` export today. Fix all three plus emit the new
   `stall_timeout`, with round-trip tests.
4. **New lint DIP145** (warning) — a graph budget default set to a **negative**
   value. Mirrors the existing DIP136 (negative manager_loop control field).

This is dippin-side only. A tracker `upstream` follow-up (extending the existing
`on-failure-tracker-followup`) is filed once this lands.

## Naming decision (ratified)

The new field is **`stall_timeout`**, a `time.Duration`. Considered
`no_progress_window` (the issue's spelling) and `no_progress_timeout`. Rejected
`no_progress_window` because `_window` reads as a *count of iterations*, which
mismatches the duration type we chose. `_timeout` / `stall_timeout` unambiguously
signal wall-clock (consistent with the repo's existing `cmd_timeout` duration
field). `stall_timeout` is the chosen spelling across **every** surface — IR field,
parser key, formatter, DOT/Attrs key, docs, and the tracker follow-up. This is a
**new** cross-repo key the tracker has never seen; we get the spelling right once.

## #94.1 — `stall_timeout` graph-level default

### IR
Add `StallTimeout time.Duration` to `ir.WorkflowDefaults`, next to `MaxWallTime`
(`ir/ir.go:52-54`). Zero value = unset (matches every sibling). No in-repo package
reads it — carried metadata for the downstream engine (verified: nothing reads the
existing budget fields either).

### Surface / parser
`defaults:` block field only (**not** the workflow header). Verified by the #92
design: the header parser accepts only `goal`/`start`/`exit`/`requires` + block
keywords (`parser/parser.go:101-128`); a header field emits "unexpected top-level
identifier". This is consistent with every other default-semantics field and with
the three existing budget fields. Parsed in `parser/parse_defaults.go`
`applyDefaultBudgetField` (the duration-parsed switch, `:134-146`) via one
`case "stall_timeout": d.StallTimeout = p.parseDuration(...)`.

### Semantics (documented; tracker enforces)
Abort/route the run when no forward progress is made for the configured wall-clock
span. Like `max_turns` exhaustion, hitting `stall_timeout` is a **failure** that
routes through the graph `on_failure` cascade (below). `0`/absent = disabled.
dippin validates shape/range only; the runtime owns "what counts as progress".

## #94.2 — `max_turns` exhaustion semantics (docs-only)

When an agent node reaches `max_turns` without completing, the engine treats it as
a **failure** (`ctx.outcome = fail`) — **not** a successful stop — which flows
through the cascade shipped in #92/#93:

1. matching fail edge (`when ctx.outcome = fail|failure`)
2. bounded node retry (`retry_target` + `max_retries`)
3. node `fallback_target`
4. graph `on_failure` (catch-all)
5. halt

**No new field.** The issue asks for the action to be "**ideally** configurable" —
"ideally," not "must." An explicit `fail|truncate|fallback` enum would be **inert**
(nothing reads it), require its own validation + DOT round-trip + lint, and is the
speculative configurability the working-style guide says to cut. Reusing the
cascade gives authors a real, already-validated routing story today. If a concrete
`truncate` need surfaces later, it is an additive field.

### Docs (load-bearing, not a footnote)
`docs/nodes.md:116` currently documents `max_turns` as a pure cost lever
("Set higher for multi-step tool-using agents"). This is the exact ambiguity #94
calls out. Rewrite the table row to state the exhaustion outcome **inline** (the
row is what authors scan), and add a short `### max_turns exhaustion` subsection:

> | `max_turns` | Integer | 1 | Maximum request-response cycles in the agent's
> tool-using loop. **Reaching this limit ends the node with outcome `fail`** — a
> hard cap, not a soft budget. The failure routes through the standard failure
> cascade (fail edge → bounded retry → `fallback_target` → graph `on_failure` →
> halt). Ensure a failure route exists (see DIP144) or the run halts on
> exhaustion. |

Reciprocally, the Failure Handling / Routing Priority cascade in
`docs/edges.md:110+` lists `max_turns` exhaustion (and `stall_timeout`) as failure
*sources* feeding the cascade, alongside `goal_gate` failure and erroring.

### DIP144 interaction (docs-only linkage)
`max_turns` → `on_failure` makes DIP144 ("agent node has no failure route") the
guardrail that catches the footgun: a `max_turns`-capped node with no failure route
is a latent dead-end. **No new lint logic** — DIP144 already fires on routeless
agents regardless of *why* they fail; do not add `max_turns`-specific logic. Add a
bidirectional docs cross-link (DIP144 section in `docs/validation.md:987+` ↔ the
`max_turns` subsection).

## #94.3 — close the pre-existing DOT export data-loss bug

**Framing (corrected by review):** this is not merely "add export for the new
field." The three **existing** budget fields are silently dropped on `.dip → DOT`
export *today*: `buildGraphAttrs` (`export/dot.go:89-110`) emits only
`tool_commands_allow`, `tool_denylist_add`, `on_failure` — never the budget fields
— yet migrate-IN reads all three (`migrate/migrate.go:185-196`) and they sit in
`reservedGraphAttrs` (`export/dot.go:61-67`). The round-trip is one-directional and
lossy: a workflow with budgets exported to DOT loses them. Budget ceilings are
safety-relevant; the #92/#93 design ratified the precedent that safety-critical
graph defaults round-trip through DOT *with* a regression test (the tool-safety
pair). #94 closes this for all four budget fields.

### Wire-format contract (ratified)
Durations cross the repo boundary as a **Go `time.ParseDuration` literal**
(canonical compact output of `formatDuration`, e.g. `5m`, `1h30m`). This is the
**existing** behavior of `max_wall_time` (confirmed: `migrate/migrate_test.go:825`
asserts `max_wall_time="30m"`; round-trip parses via `time.ParseDuration`). We keep
it and apply the same to `stall_timeout` — changing `max_wall_time` to an
integer-seconds wire form would be a breaking, out-of-scope change to a shipped
field. The spec states this as an explicit contract clause so the tracker adapter
knows to parse a Go duration literal; the int fields (`max_total_tokens`,
`max_cost_cents`) remain plain integers. The tracker follow-up names this.

### Export (`export/dot.go`)
- Emit all four budget fields in `buildGraphAttrs`, each guarded by `!= 0` so an
  unset field is never written (mirrors the formatter). Durations emit via the same
  `formatDuration` path the formatter uses — **not** raw `time.Duration.String()` —
  for round-trip symmetry.
- **Complexity (measured, mandatory):** `buildGraphAttrs` is cyclo 4 today; adding
  four inline `if`-emit branches → cyclo 8, **busting ≤5**. Extract
  `appendBudgetGraphAttrs(w, &attrs)` covering all four fields. This is required to
  pass the gate, not gold-plating.
- Add `"stall_timeout": true` to `reservedGraphAttrs` **in lockstep with the emit**
  — otherwise a `vars` entry literally named `stall_timeout` (migrated into
  `Workflow.Vars`) would emit twice in the `graph [...]` block. (The three existing
  budget keys are already reserved.)

### Migrate (`migrate/migrate.go`)
Add `case "stall_timeout": return tryApplyDurationDefault(v, &w.Defaults.StallTimeout)`
to `applyIntBudgetDefault` (`:185`). Switch case — near-zero cognit cost.

### Formatter (`formatter/format.go`)
Add `stall_timeout` to `writeDefaultsBudgetFields` (`:193`), guarded `!= 0`, via
`formatDuration`. Adds one `if` → cyclo 5 (at the cap, passes). No extraction
needed; don't over-extract.

## #94.4 — DIP145 lint (warning): negative graph budget default

Fires when any of the four graph budget defaults is **negative** (`< 0`):
`max_total_tokens`, `max_cost_cents`, `max_wall_time`, `stall_timeout`.

- **`0` = unset = no warning** for all four (the formatter emits with `!= 0`; `0` is
  structurally indistinguishable from unset, and there is no coherent "zero-length"
  semantic). Condition is `v < 0`, **not** `<= 0`. Getting this wrong false-positives
  on every workflow that omits a budget.
- **Reachability confirmed empirically:** `strconv.Atoi("-5")` and
  `time.ParseDuration("-5m")` both succeed with no error
  (`parser/parse_helpers.go:45,61`), so a negative lands in the IR with no
  structural diagnostic. DIP145 is reachable and **disjoint** from any DIP00x
  (no double-firing — verified there is no input triggering both).
- **Precedent:** DIP136 already lints negative manager_loop control fields
  (`poll_interval`/`max_cycles`) with "0 means unset" framing. DIP145 mirrors its
  voice. One unified DIP145 covers all four budget fields (DIP136 is
  manager_loop-scoped and does not cover the int ceilings, so no merge).
- **Structure (complexity-safe):** `time.Duration` is an `int64`, so table-drive on
  normalized `int64` — cyclo ~2, trivially testable:
  ```go
  checks := []struct{ name string; v int64 }{
      {"max_total_tokens", int64(d.MaxTotalTokens)},
      {"max_cost_cents",   int64(d.MaxCostCents)},
      {"max_wall_time",    int64(d.MaxWallTime)},
      {"stall_timeout",    int64(d.StallTimeout)},
  }
  for _, c := range checks { if c.v < 0 { /* append DIP145 */ } }
  ```
  New file `validator/lint_budget.go`; register `lintBudgetRanges` in
  `validator/lint.go` (near `lintCompactionThreshold`).

### Message + Help (the `0 = no limit` footgun fix)
The message **names the field and the value**; the Help **states the 0-convention**
so an author doesn't "fix" a negative by setting `0` (which means *unlimited*, the
opposite of intent — the single highest-leverage clarity fix in this issue):

> **Message:** `workflow budget default max_cost_cents is -5; budgets cannot be negative`
> **Help:** `use a positive cap (e.g. max_cost_cents: 1000 for $10.00), or omit it / set 0 for no limit`

### Explanations entry (4-field, test-gated)
```
DIP145: {
    Code:    DIP145,
    Summary: "graph budget default is negative",
    Trigger: "A workflow budget default (max_total_tokens, max_cost_cents, max_wall_time, or stall_timeout) is set to a negative value.",
    Fix:     "Use a positive cap, or omit the field / set 0 to mean no limit.",
    Example: "defaults\n  max_cost_cents: -5    // DIP145: negative; use a positive cap or 0 for no limit",
}
```
`TestExplanationsCoverAllCodes`/`NoExtra` enforce this atomically against the
`CodeDescription[DIP145]` const.

## Per-node budgets — deferred (Out of scope)

`max_turns` stays the only per-node knob (already there; we just define its
exhaustion). **No** per-node `max_total_tokens`/`max_cost_cents`/`max_wall_time`/
`stall_timeout`. Tracker #67's contract is graph-level (`graph.Attrs`); no concrete
per-node need surfaced; adding per-node ceilings invents a node-vs-defaults
precedence story with no runtime to honor it. Adding them later is purely additive
(new `AgentConfig` fields) and breaks nothing shipped here. **Named explicitly as a
deferred decision:** the most likely-missed one is per-node `max_wall_time` (the IR
already has per-node `CmdTimeout` and `MaxTurns`, so authors may expect it).

## Docs

- **`docs/syntax.md` defaults table (`:113-124`)** currently omits all three budget
  fields entirely. Add all **four** with units and the 0-convention:
  - `max_total_tokens` (Integer) — hard ceiling on total tokens; `0`/unset = no limit.
  - `max_cost_cents` (Integer) — hard ceiling on cost in **US cents** (`1000` = $10.00); `0`/unset = no limit.
  - `max_wall_time` (Duration) — hard ceiling on **wall-clock** run time (`30m`, `2h`); `0`/unset = no limit.
  - `stall_timeout` (Duration) — **wall-clock** span with no forward progress before the run aborts and routes through `on_failure` (`5m`, `90s`); elapsed time, **not** a turn count; `0`/unset = disabled.
- State the **`0 = no limit`** convention once, prominently, in the budget
  subsection: *"`0` (or unset) means **no limit** — it does not mean 'zero budget'."*
- Distinguish the family: the three `max_*` fields bound **totals** (monotonic
  ceilings); `stall_timeout` bounds **inactivity** (a sliding timer). One sentence
  prevents the author mental-model error.
- `docs/nodes.md:116` — the `max_turns` row + exhaustion subsection (above).
- `docs/edges.md:110+` — list `max_turns`/`stall_timeout` exhaustion as cascade
  failure sources.

## Examples blast radius

Measured: no example sets any budget default; only `stress_edge_cases.dip` uses
`max_turns: 10` (valid). **DIP145 will not fire on any existing example.** Add
**one new lint-clean example** that teaches the feature end-to-end: a cost ceiling +
`stall_timeout` + `max_turns` on an agent, all layered on a graph `on_failure` so
the budget abort has somewhere to go (and DIP144 stays suppressed). Inline comments
state units, the 0-convention, and "wall-clock" for `stall_timeout`. Verify
lint-clean with `just lint-examples`.

## Lint plumbing + count strings (test-gated)

- const + `CodeDescription[DIP145]` (`validator/lint_codes.go`); 4-field
  `Explanations[DIP145]` (`validator/explanations.go`); register `lintBudgetRanges`
  in `validator/lint.go`.
- Bump every hardcoded range / count string (prose only, no machine contract):
  - `"53 diagnostic codes"` → **54**: `CLAUDE.md:85`, `docs/llm-reference.md:188`,
    `docs/validation.md:3`.
  - `"documents 48 of them"` → **49**: `docs/validation.md:3` (DIP145 *gets* a
    dedicated section, so the documented count moves too — write the section before
    bumping the count, or the 49 is a lie).
  - `DIP101–DIP144` → `DIP101–DIP145`: grep **both** the en-dash (`–`) and hyphen
    (`-`) forms across `CLAUDE.md`, `validator/lint.go`, `validator/lint_codes.go:3`,
    `docs/validation.md` (`:3,:6,:15,:227,:1049`), `docs/llm-reference.md`
    (`:188,:191`).
- Regenerate `docs/generated-spec.md` + `cmd/dippin/generated-spec.md` via
  `scripts/gen-spec.sh` (never hand-edit the generated files).
- No automated test guards the prose counts — add a one-line count check to the PR
  description.

## Cross-repo follow-up (tracker, separate repo)

Extends the existing `on-failure-tracker-followup`. File a tracker `upstream` issue
(`area/dippin-lang`), specifying:
- Adapter (`pipeline/dippin_adapter.go` `extractWorkflowDefaults`) copies the now-
  exported budget keys + `stall_timeout` into `graph.Attrs`.
- **Wire-format contract:** `max_wall_time`/`stall_timeout` arrive as Go
  `time.ParseDuration` literals (compact form); `max_total_tokens`/`max_cost_cents`
  as plain integers. `stall_timeout` is a **new** key not in tracker #67.
- Budget / stall abort routes through the graph `on_failure` cascade (same as
  `max_turns` exhaustion and `on_failure` itself). Empty/zero = feature-off.

## Testing (TDD — failing test first each step)

- **parser:** `stall_timeout` → `WorkflowDefaults.StallTimeout`; negative
  (`stall_timeout: -5m`, `max_cost_cents: -5`) parses with **no** structural
  diagnostic (proves the DIP145 premise in-repo).
- **formatter:** round-trip mirroring `TestParseDefaultsBudgetRoundTrip`; assert the
  typed `Duration`, never the source string (`90s` → `1m30s` normalizes). Lock the
  `formatDuration(-5*time.Minute) == "-5m0s"` quirk with a parse→format→parse
  semantic-equality test (it is not a bug; don't "fix" it).
- **DOT export/migrate:** export emits all four budget fields; `.dip → DOT → .dip`
  round-trip preserves them (assert `Duration` equality). Include a `90s` case (the
  one input whose string changes) and a `vars`-named-`stall_timeout` collision case
  (proves the `reservedGraphAttrs` fix — no double-emit). A `1h30m` DOT-quoted
  duration migrates in correctly.
- **DIP145:** fires on each negative field; **silent on 0**; silent on positive;
  message names field + value; Help states the 0-convention.
- `TestExplanationsCoverAllCodes`/`NoExtra` for DIP145.
- `TestLintExamples` still zero-DIP108; new example lint-clean (no DIP144).
- `GOOS=js GOARCH=wasm go build ./cmd/wasm/ ./validator/` (Netlify preview gate —
  `lint_budget.go` imports only `ir`+`fmt`, wasm-safe).

## Build order

1. `stall_timeout`: IR field → parser → formatter → DOT export (+ close the 3-field
   export gap, extract `appendBudgetGraphAttrs`) → migrate → round-trip tests.
2. DIP145 lint + plumbing (`lint_budget.go`, codes, explanations, register).
3. New lint-clean example + docs (`syntax.md`, `nodes.md`, `edges.md`,
   `validation.md` DIP145 section + DIP144 cross-link) + count/range bumps + spec
   regen.
4. wasm build verification.

## Out of scope

Per-node budget fields; tracker runtime enforcement; any `error`/strict-flag
severity for DIP145; a `max_turns` action enum (`fail|truncate|fallback`); changing
`max_wall_time`'s wire format to integer seconds; node-existence validation for
`restart_target`/`retry_target`/`fallback_target` (a pre-existing #92 follow-up).
