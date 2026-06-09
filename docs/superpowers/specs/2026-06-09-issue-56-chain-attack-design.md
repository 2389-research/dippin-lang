# Issue #56 — Chain-attack detection (`${ctx.last_response}` auto-injection)

**Date:** 2026-06-09
**Issue:** [#56](https://github.com/2389-research/dippin-lang/issues/56) — follow-up deferred from #41 (`safety-follow-up`, P2).
**Status:** Design approved; ready for TDD.

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

The flow is detected via two vectors:

- **(a) `last_response` vector (1-hop):** a direct, non-restart edge `source → sink`. The runtime
  auto-injects the source's `last_response` into the sink's prompt regardless of whether the sink
  references `${ctx.last_response}` — so the flow exists structurally on every such edge.
- **(b) explicit-key vector (multi-hop):** the source declares `writes: K`, and a tool-bearing agent that
  is **downstream** of the source (reachable via forward edges) declares `reads: K`. The tainted key
  persists in context across hops, so reachability — not adjacency — is the right relation here.

### Why this is precise (the false-positive crux)

`last_response` auto-injects into *every* downstream agent by convention, so a naive "any restricted node
near an unrestricted node" rule would be noisy. The rule above fires on **exactly** the `none → full`
topology and nothing else:

- A restricted agent feeding another restricted agent (`none → none`) does **not** fire — the taint stays
  in a tool-less context; no privilege to escalate to.
- A full agent feeding a restricted agent (`full → none`) does **not** fire — wrong direction.
- A `full → none → full` laundering chain fires on the **`none → full` hop** — precisely where laundered
  content reaches a privileged agent. This is the terror-squad chain's danger point.
- The explicit-key vector requires the sink to *actually* `reads:` the key, so it flags real consumption,
  not mere proximity.

Because a `none → full` flow can still be legitimate (e.g. a classifier over trusted input feeding a
worker), and the repo has **no per-diagnostic suppression**, severity is **Hint** — matching DIP143/146
("audit this, may be intentional"), not the DIP139–142 Warnings (which fire on definite typos / dead
config).

### Diagnostic shape

- **Code:** `DIP147`, **Severity:** `SeverityHint`.
- **Location:** the **sink** node's `Source` — the privileged consumer is where the risk materializes and
  where a future mitigation knob would live.
- **Message:** names both the restricted source and the tool-bearing sink, and which vector triggered.
- **Dedup:** at most one diagnostic per `(source, sink)` pair; if both vectors hit the same pair, the
  `last_response` message wins (it is the always-on auto-injection path).

## Implementation

New file `validator/lint_chain_attack.go`, registered in `validator/lint.go`. Reuses:

- `buildForwardAdjacency` (lint.go) — not strictly needed for vector (a); the `last_response` pass iterates
  `w.Edges` directly (non-restart explicit edges) for precision.
- `computeAvailableAndUpstream` (lint_context.go) — the `upstream` map gives, per node, the set of strictly
  upstream node IDs; reused for vector (b)'s reachability.
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

## Test matrix (failing tests first)

1. `none → full` direct edge → DIP147 fires (last_response vector).
2. same chain where the downstream agent is also `none` → DIP147 does **not** fire.
3. explicit `writes:`/`reads:` key flow `none → full` → DIP147 fires (explicit-key vector).
4. benign topology (`full → full`, `full → none`) → no false positive.
5. explanation-parity test for DIP147 stays green.
6. `lint-examples`: no existing example newly trips DIP147 (fix/annotate if so).

## Out of scope (follow-ups)

- **Carry-only mitigation attribute** (`last_response_truncate:` or a structural context-input bound):
  deferred. Needs its own design (which node it lives on, wire format, round-trip). Detection ships first
  and stands alone — mirrors the DIP143-first / DIP146-follow-up split.
- **Cross-file subgraph chains:** a restricted agent in one file feeding a tool-bearing agent across a
  `subgraph`/`manager_loop` boundary. Belongs in the `cmd/dippin` native cross-file pass (like DIP146),
  **not** the wasm-safe validator.
- **Parallel-branch / fan_in / manager_loop multi-hop vectors:** v1 is agent-node `tool_access` + direct
  edges + explicit-key reachability. Branch-level overrides and join-node `last_response` ambiguity are a
  follow-up.

## Why detection, not enforcement

Per the `never-gate-dippin-on-tracker` principle: `dippin` flags the topology now; the tracker runtime
enforces information-flow control (truncation / structural context bound). DIP147 has no runtime
dependency and is valuable standalone as an author-time advisory — the same `dippin carries + lints,
runtime enforces` pattern as the rest of the `tool_access` arc.
