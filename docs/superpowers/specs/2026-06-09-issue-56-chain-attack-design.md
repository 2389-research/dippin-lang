# Issue #56 — Chain-attack detection (`${ctx.last_response}` auto-injection)

**Date:** 2026-06-09
**Issue:** [#56](https://github.com/2389-research/dippin-lang/issues/56) — follow-up deferred from #41 (`safety-follow-up`, P2).
**Status:** Implemented. Scope narrowed during build — see § Scope decision (#57).

## Scope decision (#57)

Issue [#57](https://github.com/2389-research/dippin-lang/issues/57) ("Cross-node lint:
`tool_access: none` source → `full` target edge warning") was filed as a sibling non-goal of #56 and
**closed/deferred** ("useful once cascade exists; less so when authors annotate per-node"); `skill.md`
calls it "the rejected in-file graph-topology lint." A bare `none → full` *edge* warning (the
`${ctx.last_response}` auto-injection topology) is therefore **out of scope** for this PR.

DIP147 instead ships only the **explicit-key vector** — a flow the author wired by hand
(`writes: K` on a restricted agent → `reads: K` on a downstream tool-bearing agent). This is genuinely
new (#57 never covered named-key laundering), maximally precise (no edge-proximity heuristic — it fires
only on a declared data dependency), and does not resurrect the rejected edge warning. The
`${ctx.last_response}` auto-injection vector and the `last_response_truncate:` mitigation attribute remain
#56 follow-ups. Consequently this PR **partially addresses** #56 (it does not `Closes` it).

## The gap

`tool_access: none` bounds the **tools** available to an agent's LLM call. It does **not** bound the
**information flow** between agents. A `tool_access: none` summarizer that processes untrusted content,
then feeds a `tool_access: full` writer via `${ctx.last_response}` auto-injection (or via a named context
key), remains a viable injection-laundering chain: the restricted agent's output carries a payload that a
privileged downstream agent then acts on with full tools.

This closes the conceptual gap the #41 safety epic leaves open. `tool_access` neutralizes a compromised
agent's *tools*; its *output* can still launder a payload into a privileged context. The existing DIP143
and DIP146 explanations already point here: *"this bounds the child's tool catalog, not information flow
across the supervisory boundary ... — see #56."*

Root-cause analysis: `docs/superpowers/research/2026-05-19-issue-41-terror-squad.md` finding #5 (the
`full → read_only → full` laundering chain). Deferred as a non-goal in
`docs/superpowers/specs/2026-05-26-issue-41-design.md`.

## What ships: DIP147 (detection lint, in-file)

A **detection/advisory** lint — `dippin` flags the dangerous topology; the runtime (tracker) enforces the
actual information-flow control (truncation / context bound). **No tracker dependency**: this is pure
`ir.Workflow` + graph analysis, wasm-safe, immediately valuable, and ships alone.

### The rule

DIP147 fires when a **restricted source** agent's output reaches a **tool-bearing sink** agent, where:

- **source** = an agent node with canonical `tool_access == "none"` (after trim + lowercase). A deliberate,
  recognized restriction. Invalid/typo values are out of scope — DIP139 owns those, and they fail closed
  to no-tools rather than expressing deliberate `none` intent.
- **sink** = an agent node with canonical `tool_access == ""` (the full catalog / default). Only agents
  have a tool catalog and an auto-injected prompt.

The flow is detected via the **explicit-key vector** (multi-hop): the source declares `writes: K`, and a
tool-bearing agent that is **downstream** of the source (reachable via forward edges) declares `reads: K`.
The tainted key persists in context across hops, so reachability — not adjacency — is the right relation,
and a non-agent node (e.g. a tool node) between source and sink does not hide the flow.

### Why this is precise (the false-positive crux)

The rule fires on an **author-declared data dependency**, never on graph adjacency, so it cannot be noisy
the way a bare-edge heuristic would be:

- A restricted agent's key flowing into another restricted agent (`none → none`) does **not** fire — the
  taint stays in a tool-less context; no privilege to escalate to.
- A `full → … → full` or `full → none` key flow does **not** fire — no restricted-source escalation.
- The sink must *actually* declare the key in `reads:`, so it flags real, hand-wired consumption.

Because a restricted-source key flow can still be legitimate (a classifier over trusted input feeding a
worker), and the repo has **no per-diagnostic suppression**, severity is **Hint** — matching DIP143/146
("audit this, may be intentional"), not the DIP139–142 Warnings (which fire on definite typos / dead
config).

### Diagnostic shape

- **Code:** `DIP147`, **Severity:** `SeverityHint`.
- **Location:** the **sink** node's `Source` — the privileged consumer is where the risk materializes and
  where a future mitigation knob would live.
- **Message:** names the restricted source, the laundered key, and the tool-bearing sink.
- One diagnostic per `(source, key, sink)` flow (each distinct restricted source upstream of a sink that
  shares a written/read key).

## Implementation

New file `validator/lint_chain_attack.go`, registered in `validator/lint.go`. Reuses:

- `buildForwardAdjacency` (lint.go) + `computeAvailableAndUpstream` (lint_context.go) — the `upstream` map
  gives, per node, the set of strictly upstream node IDs; used for the explicit-key reachability relation.
- The canonicalization idiom from `lint_tool_access.go` (`strings.ToLower(strings.TrimSpace(...))`).

The validator must **not** import the parser or cmd-dippin and must compile to wasm — DIP147 touches only
`ir.Workflow` + graph, satisfying this.

### Edit-site checklist (per `adding-a-dip-lint-code` memory)

1. `validator/lint_codes.go` — `DIP147` const + `CodeDescription[DIP147]`.
2. `validator/explanations.go` — `Explanations[DIP147]` in `safetyExplanations()`, all four fields
   (Summary / Trigger / Fix / Example) non-empty (gated by `TestExplanationsCoverAllCodes`).
3. `validator/lint.go` — register `lintChainAttack(w)` in `Lint()`.
4. `validator/lint_chain_attack.go` + `validator/lint_chain_attack_test.go` (table-driven, parser-driven
   via `lintSrc`/`hasCode`).
5. Docs: `docs/validation.md` (DIP147 explanation, next to DIP143–146); the gen-spec source docs
   (`docs/llm-reference.md`, `site/static/skill.md`) then regen `cmd/dippin/generated-spec.md` via
   `scripts/gen-spec.sh` (freshness-gated by `releasecheck`). Sweep the hardcoded `DIP14x` range strings.

## Test matrix

1. explicit `writes:`/`reads:` key flow `none → full` → DIP147 fires; message names both nodes + key.
2. bare `none → full` edge with no declared key (last_response topology) → DIP147 does **not** fire
   (documents the #57 scope boundary).
3. `none → none` key flow → DIP147 does **not** fire (sink restricted; no escalation).
4. multi-hop key flow across a non-agent (tool) node → DIP147 fires (reachability, not adjacency).
5. benign topology (`full → full`, `full → none` key flow) → no false positive.
6. explanation-parity test for DIP147 stays green.
7. `lint-examples`: no existing example newly trips DIP147.

## Out of scope (follow-ups)

- **`${ctx.last_response}` auto-injection edge (bare `none → full`)** and the **`last_response_truncate:`
  mitigation attribute:** the edge-warning topology is issue #57 (closed/deferred); the attribute is #56's
  literally-named mitigation. Both remain follow-ups; this PR partially addresses #56.
- **Cross-file subgraph chains:** a restricted agent in one file feeding a tool-bearing agent across a
  `subgraph`/`manager_loop` boundary. Belongs in the `cmd/dippin` native cross-file pass (like DIP146),
  **not** the wasm-safe validator.
- **Parallel-branch / fan_in / manager_loop vectors:** v1 is agent-node `tool_access` + explicit-key
  reachability. Branch-level overrides are a follow-up.

## Why detection, not enforcement

Per the `never-gate-dippin-on-tracker` principle: `dippin` flags the topology now; the tracker runtime
enforces information-flow control (truncation / structural context bound). DIP147 has no runtime
dependency and is valuable standalone as an author-time advisory — the same `dippin carries + lints,
runtime enforces` pattern as the rest of the `tool_access` arc.
