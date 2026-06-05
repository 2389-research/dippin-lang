# Documentation Coverage Audit — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use `superpowers:subagent-driven-development` to execute this plan. It is built around dispatching one small, single-purpose subagent per documentation file. Steps use checkbox (`- [ ]`) syntax for tracking. Read the **Operating Rules** before dispatching anything.

**Goal:** Bring every document in `docs/` (plus the two spec-source files outside it) fully up to date with the dippin language surface as it exists on `origin/main` today — with particular focus on the ~5 weeks of features added between 2026-05-01 and 2026-06-05.

**Architecture:** A three-phase fan-out. **Phase 0** establishes a single canonical *Feature Inventory* (the source of truth, embedded in this plan and re-verified against code). **Phase 1** dispatches one read-only audit subagent per doc file, each attacking its file from a specific angle and emitting a structured gap report — no edits. **Phase 2** dispatches one fix subagent per file-with-gaps (files are disjoint, so fixes never collide). **Phase 3** regenerates derived artifacts (`generated-spec.md`) and verifies the whole tree with `just check`.

**Tech Stack:** Markdown + EBNF docs; Go toolchain behind `just`; `dippin` CLI; bash subagents.

---

## Operating Rules (read before executing)

1. **Base must be current.** This plan was authored against `origin/main` at commit `540e129` (budget/DIP145) — which includes `on_failure`/DIP144 **and** budget/`stall_timeout`/DIP145. Before starting, run the **Phase 0** verification; if the inventory drifts from code, fix the inventory in this file first, then proceed. Do **not** audit docs against a stale checkout.

2. **Subagents are stateless and single-purpose.** Each Phase-1 / Phase-2 subagent gets exactly one file (or one tight cluster) and the *complete* checklist it needs — never "see the inventory above." The relevant inventory rows are copied into each dispatch prompt.

3. **Phase 1 is strictly read-only.** Audit subagents may run `grep`/`dippin`/`Read` but must **not** edit. They return a gap report only. This lets all Phase-1 agents run in parallel with zero conflict risk.

4. **Phase 2 fixes touch one file each.** Because every fix subagent owns a distinct file, they can run concurrently in the shared worktree without colliding. (No per-agent worktree needed — the no-conflict invariant is satisfied by file-disjointness, per the project's worktree rule.) The sole exception is the spec-source pair (`docs/llm-reference.md` + `site/static/skill.md`) feeding `generated-spec.md`; treat regeneration as Phase 3, never hand-edit the generated file.

5. **Never hand-edit derived files.** `docs/generated-spec.md`, `cmd/dippin/generated-spec.md`, and `site/static/llms-full.txt` are assembled by `scripts/gen-spec.sh` from `docs/llm-reference.md` (its `## Grammar …` section onward) + `site/static/skill.md` (its `## File Structure … ## Documentation` section). Fix those two sources, then run `just gen-spec`.

6. **Verify, don't assume.** Every claim a subagent adds to a doc must be backed by code on the current base — the `ir` struct field, the validator code registration, an `examples/*.dip`, or the CHANGELOG entry. The inventory rows below each cite their code anchor; subagents confirm the anchor still exists before documenting.

7. **Scope discipline (project CLAUDE.md).** Surgical changes only. Add/correct coverage of real features; do not rewrite prose that is already correct, do not "improve" unrelated sections, do not invent fields. If a doc is already accurate for its scope, the correct gap report is "no gaps."

---

## Feature Inventory — Source of Truth (new surface, 2026-05-01 → 2026-06-05)

Every row below is a feature that landed in the audit window and that docs must reflect. **Code anchor** is what a subagent greps to confirm the feature is real on the current base. **Owners** lists every doc file that must mention the feature (a file not listed should *not* grow a mention).

| # | Feature | Code anchor | Owner docs |
|---|---------|-------------|------------|
| A | `requires:` workflow header (env deps, e.g. `git, docker`) — v0.26.0 | `ir.go: Requires []string` | GRAMMAR.ebnf, syntax.md, nodes.md(header), llm-reference.md, skill.md |
| B | Tool-routing fields `marker_grep`, `route_required`, `output_limit` + reserved `ctx.tool_marker` / `ctx.tool_route` — v0.28/0.29 | `ir.go: MarkerGrep/RouteRequired/OutputLimit`; `validator` reserves `ctx.tool_route` | GRAMMAR.ebnf, nodes.md(tool), context.md, edges.md(routing), analysis.md(coverage), llm-reference.md, skill.md, examples/marker_routing.dip |
| C | `dippin coverage` AST-based shell parsing (respects `>`,`>>`,`&>`, pipes, cmd-subst) — v0.30.0 | `coverage` uses `mvdan.cc/sh/v3/syntax` | analysis.md, cli.md |
| D | `mode: yes_no` on `human` nodes (4th mode beside choice/freeform/interview) — v0.31.0 | `validator` DIP127 accepts `yes_no` | nodes.md(human), validation.md(DIP127), GRAMMAR.ebnf, llm-reference.md, skill.md |
| E | `tool_access: none` agent-node safety primitive + DIP139/DIP140/DIP141 — v0.32.0 | `ir.go: ToolAccess`; `validator` DIP139/140/141 | nodes.md(agent), syntax.md, validation.md, integration.md(runtime pairing), llm-reference.md, skill.md, GRAMMAR.ebnf, examples/agent_tool_access.dip |
| F | `command_file:` directive on tool nodes — v0.33.0 | `ir.go: CommandFile`; `parser.ResolveFileDirectives` | nodes.md(tool), syntax.md, integration.md(unresolved-IR view), architecture.md(resolver pass), llm-reference.md, skill.md, GRAMMAR.ebnf, examples/external_files.dip |
| G | `prompt_file:` / `system_prompt_file:` directives on agent nodes — v0.34.0 | `ir.go: PromptFile/SystemPromptFile` | nodes.md(agent), syntax.md, integration.md, llm-reference.md, skill.md, GRAMMAR.ebnf, examples/external_prompts.dip |
| H | `writable_paths:` glob write-jail on agent nodes + DIP142 (+ DIP141 nullified-by-`tool_access:none`) — v0.35.0 | `ir.go: WritablePaths`; `validator` DIP142 | nodes.md(agent), context.md(vs `writes:`), validation.md, integration.md(version-skew/fail-closed), syntax.md, llm-reference.md, skill.md, GRAMMAR.ebnf, examples/agent_writable_paths.dip |
| I | Per-branch `tool_access` / `writable_paths` override on **block-form parallel** branches — #58/#75 | `ir.go: ParallelBranch.ToolAccess/WritablePaths` | nodes.md(parallel), syntax.md, GRAMMAR.ebnf |
| J | DIP143 — subgraph/`manager_loop` child does not inherit parent `tool_access` (advisory) — v0.36.0 | `validator` DIP143 | subgraph-composition.md, validation.md, nodes.md(subgraph/manager_loop) |
| K | Graph-level `on_failure:` route — **defaults-block field ONLY** (no per-node `on_failure` exists in any NodeConfig; verified Phase 1) + DIP144 failure-route lint — #92/#93 | `ir.go: WorkflowDefaults.OnFailure`; `validator` DIP144 | edges.md, syntax.md(defaults only), nodes.md, validation.md, llm-reference.md, skill.md, GRAMMAR.ebnf, examples/on_failure_route.dip |
| L | Declarable budget/limit attrs: `stall_timeout`, `max_turns` exhaustion, `max_total_tokens`/`max_cost_cents`/`max_wall_time` + DIP145 — #94 | `ir.go: StallTimeout` + Max* fields; `validator` DIP145 | syntax.md(defaults), validation.md, cost/analysis.md, llm-reference.md, skill.md, GRAMMAR.ebnf |
| M | `@file`/`*_file` directive resolver hardened against leaf TOCTOU (`O_NOFOLLOW`, fstat-the-fd, symlink/parent-escape/4MiB caps) — #67/#79 | `parser/resolve_nofollow_*.go`; `parser.ResolveFileDirectives` | integration.md(security), architecture.md(resolver pass), syntax.md(directive security caps) |
| N | DIP138 reserved (no firing logic) + overall code count is now **54** (DIP001–009, DIP101–145) | `validator` DIP138; `grep -rhoE 'DIP[0-9]{3}' validator/*.go \| sort -u \| wc -l` → 54 | validation.md |

**Cross-cutting truth checks** (any doc may need these):
- DIP-code total is **54**; `validation.md` line 3 must read "54 … documents 49 …" (verify the second number against how many have dedicated sections).
- `tool_access` is **node-scoped** — it constrains a single node's executor and does **not** taint downstream nodes (resolves #57 by doc, not lint). Any doc describing `tool_access` propagation is wrong.
- `*_file` directives are inlined by `dippin pack`; the **parser stays pure** (no FS I/O). LSP/WASM see the *unresolved* IR view (`*File` set, content empty). `command_file:` is **lossy through DOT round-trip** (rewrites to inline `command:`).

---

## Phase 0 — Verify the inventory against code (1 subagent, ~5 min)

- [ ] **Step 1: Dispatch the ground-truth verifier**

Dispatch a single `general-purpose` subagent (read-only) with this prompt:

```
You are verifying a documentation-audit Feature Inventory against the actual dippin
codebase on the CURRENT branch. Working dir is the repo root. Do NOT edit anything.

For each of these anchors, run the command and report PRESENT/ABSENT + the matched line:
1.  grep -nE 'Requires|OnFailure|StallTimeout|WritablePaths|ToolAccess|CommandFile|PromptFile|SystemPromptFile|MarkerGrep|RouteRequired|OutputLimit' ir/ir.go
2.  grep -rhoE 'DIP[0-9]{3}' validator/*.go | sort -u   (report the full list + total count)
3.  grep -rn 'DIP138\|DIP139\|DIP140\|DIP141\|DIP142\|DIP143\|DIP144\|DIP145' validator/lint_codes.go
4.  grep -n 'yes_no' validator/*.go
5.  ls examples/ | grep -E 'tool_access|writable_paths|external_files|external_prompts|marker_routing|on_failure'
6.  head -5 docs/validation.md   (report the exact "registers N … documents M" line)
7.  grep -rn 'mvdan.cc/sh' coverage/ | head

Return a markdown table: Inventory row (A–N) | anchor found? | exact evidence line |
note any DISCREPANCY between the plan's claimed anchor and reality.
```

- [ ] **Step 2: Reconcile**

If the verifier reports any discrepancy (a code anchor moved/renamed, a DIP code missing, the count ≠ 54), **edit the Feature Inventory in this plan file** to match reality before continuing. Commit the plan correction:

```bash
git add docs/superpowers/plans/2026-06-05-docs-coverage-audit.md
git commit -m "docs(plan): reconcile audit inventory with code ground truth"
```

If no discrepancies: proceed to Phase 1.

---

## Phase 1 — Parallel read-only file audits (15 subagents)

Dispatch all of these **in one batch** (they are read-only and independent). Each subagent uses this **shared prompt skeleton**, with the per-file `TARGET`, `ANGLE`, and `CHECKLIST` substituted from the task blocks below:

```
You are auditing ONE documentation file for coverage gaps against the current dippin
codebase. Working dir is the repo root. You are READ-ONLY: do not edit any file.

TARGET FILE: <TARGET>
ANGLE: <ANGLE>

Method:
1. Read <TARGET> in full.
2. For each CHECKLIST item below, decide: is this feature CORRECTLY and COMPLETELY
   documented in <TARGET> for THIS file's scope (per ANGLE)? Confirm each feature is
   real by grepping the cited code anchor before judging it missing.
3. Also flag: stale claims (counts, dates, "future"/"deferred" items that have shipped),
   broken cross-references, and field names/syntax that don't match the parser.

CHECKLIST (features this file owns):
<CHECKLIST>

Output a gap report as markdown:
## <TARGET> — gap report
- Status: CURRENT | GAPS FOUND
- For each gap: { feature, what's wrong/missing, exact section/line to change,
  the corrected text or a precise instruction, code-anchor evidence }
Do NOT edit the file. Return only the report.
```

> **Granularity note:** keep each agent to one file so its context stays small and its report is precise. The two largest files (`nodes.md`, `validation.md`) still go to a single agent each, but their CHECKLISTs are scoped tightly below.

### Task 1.1 — `docs/GRAMMAR.ebnf`
- [ ] **Angle:** *Canonical grammar completeness* — every new keyword/field has an EBNF production and appears in the right rule.
- **CHECKLIST:** A (`requires`), B (`marker_grep`/`route_required`/`output_limit`), D (`yes_no` in human mode enum), E (`tool_access`), F (`command_file`), G (`prompt_file`/`system_prompt_file`), H (`writable_paths`), I (per-branch overrides in block-form parallel rule), K (`on_failure` in defaults + node), L (`stall_timeout`/`max_turns`/budget attrs in defaults). Confirm the header comment block still accurately describes lexer behavior.

### Task 1.2 — `docs/llm-reference.md`
- [ ] **Angle:** *Compact BNF card + field tables (feeds `generated-spec.md`)* — every new field present in the simplified BNF and any field tables; nothing runtime-only leaking in.
- **CHECKLIST:** A, B, D, E, F, G, H, K, L. Note: this file's `## Grammar …` section is copied verbatim into `generated-spec.md` by `gen-spec.sh` — gaps here propagate.

### Task 1.3 — `site/static/skill.md`
- [ ] **Angle:** *Hosted skill / agent-facing authoring guide (feeds `generated-spec.md`)* — `## File Structure … ## Documentation` span is copied into the spec; ensure new fields are demonstrated in authoring examples, not just listed.
- **CHECKLIST:** A, B, D, E, F, G, H, K, L. Also verify the 4-mode human list (choice/freeform/interview/yes_no).

### Task 1.4 — `docs/syntax.md`
- [ ] **Angle:** *Full syntax reference — top-level + defaults + per-node fields.*
- **CHECKLIST:** A (header), E, F, G, H (agent fields), I (parallel block overrides), K (`on_failure` in both defaults and node), L (budget attrs in `defaults`), M (`*_file` directive security caps & path-relative-to-`.dip` rule). Confirm the `defaults` block field list is complete.

### Task 1.5 — `docs/nodes.md`
- [ ] **Angle:** *Per-node-kind field coverage* — each node kind's section lists its current fields with correct semantics.
- **CHECKLIST:** agent → E (`tool_access`), G (`prompt_file`/`system_prompt_file`), H (`writable_paths`); tool → B (`marker_grep`/`route_required`/`output_limit`), F (`command_file`); human → D (`yes_no`); parallel → I (per-branch `tool_access`/`writable_paths`); subgraph + manager_loop → J (DIP143 non-inheritance note); header → A (`requires`), K (`on_failure`). Cross-cutting: `tool_access` is node-scoped (no downstream taint).

### Task 1.6 — `docs/edges.md`
- [ ] **Angle:** *Edge & routing semantics.*
- **CHECKLIST:** K (`on_failure` graph-level recovery route — how it interacts with explicit `when … fail` edges and `restart:`), B (stdout-based routing via `ctx.tool_route`/`ctx.tool_marker` and how edges consume them). Confirm `restart:`, `weight:`, `label:`, condition operators are all still accurate.

### Task 1.7 — `docs/context.md`
- [ ] **Angle:** *Context keys & write-scope semantics.*
- **CHECKLIST:** B (reserved `ctx.tool_marker` / `ctx.tool_route` keys — what populates them), H (`writable_paths` = filesystem write *location* jail, vs the advisory `writes:` = context keys; these are different axes and must not be conflated). Confirm the reserved-key list is complete.

### Task 1.8 — `docs/validation.md`
- [ ] **Angle:** *Diagnostic catalog completeness + correct count.*
- **CHECKLIST:** N (count = 54; verify the "registers 54 … documents M" line and that M matches reality), D (DIP127 now lists `yes_no`), E (DIP139/140/141 sections exist & correct), H (DIP142), J (DIP143), K (DIP144), L (DIP145), and DIP138 documented as reserved/non-firing. Each DIPxxx in `DIP101–145` should have a dedicated section unless intentionally internal — list any code with no section.

### Task 1.9 — `docs/architecture.md`
- [ ] **Angle:** *Internal pipeline accuracy — packages, passes, data flow.*
- **CHECKLIST:** F/G (`parser.ResolveFileDirectives` is a separate post-parse pass; parser stays pure — is it in the data-flow diagram?), M (`resolve_nofollow_*.go` hardening; O_NOFOLLOW open-once resolver), E/H (validator gained `lint_subgraph_tool_access.go`, `lint_failure_route.go`, etc. — package responsibilities current?). Confirm the package-dependency rules text matches CLAUDE.md (dipx loader-tier exemption, analysis composition).

### Task 1.10 — `docs/analysis.md`
- [ ] **Angle:** *Analysis-command behavior.*
- **CHECKLIST:** C (coverage AST: redirection/pipe/cmd-subst handling — is the `partial`→`covered` behavior documented?), B (coverage's interaction with `marker_grep`/markers-and-verbose-output convention), L (does `cost` surface the new budget ceilings `max_total_tokens`/`max_cost_cents`/`max_wall_time`/`stall_timeout`?).

### Task 1.11 — `docs/cli.md`
- [ ] **Angle:** *Command/flag surface.*
- **CHECKLIST:** confirm `dippin spec` is documented; C (coverage redirection note if user-visible); any new flags introduced by the audit-window features. Verify the command list matches `cmd/dippin/` reality (`grep -l 'func (c \*CLI) Cmd' cmd/dippin/*.go`).

### Task 1.12 — `docs/integration.md`
- [ ] **Angle:** *External-consumer contract — embedding, resolution, runtime pairing, safety.*
- **CHECKLIST:** F/G (unresolved-IR view: LSP/WASM see `*File` set + empty content; `dippin pack` inlines), E (`tool_access` requires an enforcing runtime — meaningless without it; node-scoped), H (`writable_paths` fail-closed + version-skew-is-a-safety-requirement: a non-enforcing runtime must refuse to start, pin-never-`@latest`), M (directive security caps consumers rely on). Confirm `command_file:` DOT-round-trip lossiness is noted where relevant.

### Task 1.13 — `docs/subgraph-composition.md`
- [ ] **Angle:** *Composition semantics & cross-file boundaries.*
- **CHECKLIST:** J (DIP143: child `.dip` does not inherit parent `tool_access`; fires only when workflow declares `tool_access` intent AND references an external subgraph; self-reference not flagged; cross-file effective-access enforcement deferred to #89), E (per-node `tool_access` does not cross a file boundary). Confirm `ref:` / `subgraph_ref:` and `params:` passing are accurate.

### Task 1.14 — `docs/evolution_report.md`
- [ ] **Angle:** *Stale-report reconciliation* — this file is dated **March 2025** and frames retry/tool/composition features as "v1.5+ future." Reconcile every "deferred/future" claim against what has now shipped.
- **CHECKLIST:** Walk each "future improvement" / "cannot yet be expressed" claim and mark SHIPPED (with the feature + version) vs STILL-DEFERRED. Features likely now shipped: `tool_access`, `writable_paths`, `on_failure`, budget ceilings, `requires`, file directives, subgraph composition. **Deliverable nuance:** this is a point-in-time report — the fix may be a dated "Update (2026-06): the following items have since shipped …" addendum rather than rewriting history. Flag the decision; do not rewrite silently.

### Task 1.15 — Stable-docs sweep: `docs/cli.md`-adjacent set
- [ ] **Angle:** *Confirm no stale references in lower-churn docs.* Single subagent reads `docs/CONTRIBUTING.md`, `docs/testing.md`, `docs/editor-setup.md`, `docs/DIPPIN_DESIGN_PLAN.md` and reports only whether any references a now-changed field/count/command, or claims a feature is unimplemented that has shipped. `DIPPIN_DESIGN_PLAN.md` is historical — flag stale "proposed/future" framing but recommend an addendum over a rewrite (same treatment as evolution_report).
- **CHECKLIST:** N (any DIP-count references), E/H/K/L (any "not yet supported" claims for shipped features).

- [ ] **Step (Phase 1 close): Collate reports.** Gather all 15 gap reports. Produce a single consolidated punch-list grouped by file, dropping every "Status: CURRENT" file. This punch-list drives Phase 2.

---

## Phase 2 — Apply fixes (1 subagent per file with gaps)

For each file the Phase-1 punch-list flags `GAPS FOUND`, dispatch one fix subagent. These run concurrently (disjoint files). **Do not** dispatch a fix agent for a `CURRENT` file.

- [ ] **Step 1: Dispatch fix subagents** — one per gapped file, each with this prompt:

```
Apply documentation fixes to EXACTLY ONE file. Working dir is repo root.
Touch ONLY this file: <TARGET>. Do not edit any other file.

Below is the gap report for <TARGET> from the read-only audit. For each gap, make the
minimal, surgical edit that fixes it — match the file's existing voice, heading style,
table format, and code-fence conventions. Do NOT rewrite correct prose, do NOT touch
unrelated sections, do NOT invent fields beyond the gap report. Every added claim must
match the code anchor cited in the report.

GAP REPORT:
<paste the file's Phase-1 report>

When done, output a unified diff of your changes and a one-line summary per gap closed.
```

- [ ] **Step 2: Spot-check each diff.** Reject and re-dispatch any fix that touched a second file, rewrote unrelated prose, or added an unverified claim (project rule: every changed line traces to a gap).

- [ ] **Step 3: Commit Phase 2.**

```bash
git add docs/ site/static/skill.md
git commit -m "docs: audit + refresh coverage for v0.26–v0.37 surface (requires, tool_access, writable_paths, file directives, on_failure, budget attrs, DIP138–145)"
```

---

## Phase 3 — Regenerate derived artifacts & verify

- [ ] **Step 1: Regenerate the spec** (picks up `llm-reference.md` + `skill.md` edits):

```bash
just gen-spec
```

Expected: `gen-spec: wrote docs/generated-spec.md`. This also refreshes `cmd/dippin/generated-spec.md`. Then regenerate the site mirror if `just check` expects it:

```bash
cp docs/generated-spec.md site/static/llms-full.txt
```

- [ ] **Step 2: Diff-review the generated spec.** Confirm the new fields now appear in `docs/generated-spec.md` (they should, since they came from the two sources):

```bash
grep -nE 'tool_access|writable_paths|command_file|prompt_file|on_failure|stall_timeout|yes_no|requires' docs/generated-spec.md
```

Expected: each term present at least once. If any is missing, the gap is in `llm-reference.md` or `skill.md` — fix the source, re-run `just gen-spec`.

- [ ] **Step 3: Full verification.**

```bash
just check
```

Expected: build, vet, fmt-check, lint-go, test-race, releasecheck, complexity, validate-examples, pack-examples all pass. (Note from project memory: `just check` may fail at the `tree-sitter-test` stage if the tree-sitter CLI is absent locally — that's an environment gap, not an audit regression. If it fails *only* there, run the pre-commit hook instead as the real gate, and confirm `spec-check`/`build`/`validate-examples` passed.)

- [ ] **Step 4: Commit regenerated artifacts.**

```bash
git add docs/generated-spec.md cmd/dippin/generated-spec.md site/static/llms-full.txt
git commit -m "docs: regenerate generated-spec.md from refreshed sources"
```

- [ ] **Step 5: Open the PR.** Push the branch and open a PR summarizing per-file changes and listing the inventory rows (A–N) now covered. Per project policy, never push to `main`; merge only via reviewed PR.

---

## Self-Review Checklist (run after Phase 3, before PR)

1. **Inventory coverage:** For each inventory row A–N, can you point to at least one committed doc change (or a justified "already current")? List any row with no owner-doc change and confirm it was genuinely already documented.
2. **Owner discipline:** Did any file grow a mention of a feature it does *not* own (per the inventory's Owner column)? If so, that's scope creep — revert it.
3. **Derived-file integrity:** `docs/generated-spec.md`, `cmd/dippin/generated-spec.md`, `site/static/llms-full.txt` were produced by `just gen-spec`, **not** hand-edited. Confirm `git diff` on them matches what the script emits (re-run `just gen-spec` → `git diff --exit-code docs/generated-spec.md` should be clean).
4. **Count correctness:** `validation.md` reports 54 codes; the number matches `grep -rhoE 'DIP[0-9]{3}' validator/*.go | sort -u | wc -l`.
5. **No stale "future":** `evolution_report.md` and `DIPPIN_DESIGN_PLAN.md` no longer assert shipped features are deferred (or carry a dated addendum saying so).
6. **Examples referenced:** Where a feature ships with an `examples/*.dip`, the owning doc points to it.
