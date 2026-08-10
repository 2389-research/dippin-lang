# dip 2 retry channel (Option A for #186/#204) — Design

**Status:** Proposal, needs maintainer + tracker-engine sign-off before build
**Date:** 2026-08-10
**Issues:** #186 (migration symptom + shipped exit-4 stopgap, v0.55.0), #204 (the A-vs-B design fork), #134 (the dip-2 routing redesign this partially reopens)

---

## Problem (recap)

dip 2 has two runtime channels but only models one. **Structural routing** (success/fail/branch) lives in the edges block (#134, correct). The **retry channel** (`OutcomeRetry` — fired by cost/stall/infra/empty-response conditions) is dispatched by the engine's `processRetryOutcome`, which reads **node attributes** and never consults the edges block: the retry destination from `retry_target` (else self), and retry-exhaustion from a fallback attribute (else terminal halt).

#134 made the parser **reject** `retry_target`/`fallback_target` in dip 2, so dip 2 has no representation the retry engine reads. `fmt --migrate` converts them to `loop`/`on fail` edges the retry channel ignores → silent non-equivalence (the retry becomes a self-retry; exhaustion terminal-halts). v0.55.0 flags this with exit 4 but does not fix it.

## Decision: Option A — dip 2 re-admits the retry channel as node attributes

Retry is a node-level policy triggered by runtime conditions, not graph structure — and dip 2 **already** keeps the rest of that policy on the node (`max_retries`, `retry_policy`, `base_delay`). #134's error was deleting only the retry *destination* and *exhaustion fallback* while keeping the budget/policy. A restores them:

```dip
agent Review
  max_retries: 3
  retry_target: Implement            # re-admitted in dip 2 (non-self retry destination)
  fallback_retry_target: MarkFailed  # new dip-2 spelling for the retry-exhaustion route
  prompt: …
edges
  Review -> Done  on success         # structural routing stays in edges
```

The engine already reads these attrs, so **A requires no engine change** — only that tracker confirms the field names it reads. Migration becomes lossless and exit-4 goes away for these cases.

## The two sub-decisions this needs sign-off on

### Sub-decision 1 — reopen #134 for the retry channel

A re-admits `retry_target` (and adds `fallback_retry_target`) as node fields in dip 2, which the parser currently rejects (`isV2RejectedNodeField`). This is **purely additive**: no existing dip-2 file uses these fields (they don't parse today), so nothing breaks. It does walk back #134's "all routing is an edge" for the retry channel specifically — the deliberate, justified scope of this proposal. **Recommendation: yes.**

### Sub-decision 2 — how does v1 `fallback_target` migrate? (the load-bearing, potentially-breaking one)

Today dippin migrates v1 `fallback_target` → an **`on fail` edge** (treating it as an on-*failure* route). The tracker engine treats `fallback_target` as **retry-exhaustion** (a different channel). These can't both be right, and this is exactly the disagreement #204 names.

- **If the engine's model is authoritative** (fallback_target = retry exhaustion), the correct migration is v1 `fallback_target` → dip-2 **`fallback_retry_target`** (a node attr), NOT an on-fail edge. That makes dippin's *current* `migrateFallbackTarget` (→ on-fail edge) a **latent bug** producing non-equivalent output today — and A would change that migration's behavior.
- **If both meanings legitimately exist** (some authors want on-failure routing, some want retry-exhaustion), we may need to keep the on-fail edge migration AND add `fallback_retry_target`, and the migrator can't tell which the author meant → it must flag for review.

**This is the one call I can't make from dippin alone** — it depends on what the tracker engine actually reads and what v1 authors meant. Recommendation pending tracker confirmation: treat `fallback_target` as retry-exhaustion (migrate → `fallback_retry_target`), since that's what the engine dispatches and what #186 reports as the real-world breakage. But it is potentially breaking for any workflow that currently relies on the on-fail-edge migration, so it needs an explicit yes.

## Implementation surface (once ratified)

- **Grammar** (`docs/GRAMMAR.ebnf`): document `retry_target` and `fallback_retry_target` as valid dip-2 `common_field`s (or agent/tool-scoped); note they are the retry channel, distinct from edges.
- **IR** (`ir/ir.go`): `RetryConfig` already has `RetryTarget`/`FallbackTarget`; add `FallbackRetryTarget` (or repurpose `FallbackTarget` as the retry-exhaustion field and drop the on-fail conflation). Additive, omitempty.
- **Parser** (`parser/parse_nodes.go`): remove `retry_target` (and the new field) from `isV2RejectedNodeField`; parse them under dip 2. Keep the dip-1 parse path unchanged.
- **Migration** (`formatter/migrate_v2.go`): `migrateRetryTarget` preserves a non-self `retry_target` verbatim instead of synthesizing a loop edge (removes the exit-4 case); `migrateFallbackTarget` resolves per sub-decision 2. Update/retire the exit-4 stopgap accordingly.
- **Formatter** (`formatter/format.go`): emit the retry fields under dip 2 (currently `writeRetryTargetFields` is v1-shaped).
- **Sweep:** tree-sitter grammar, VS Code/Zed/site highlighters, `docs/llm-reference.md`, `skill.md`, `docs/cli.md` (the exit-4 note), examples, changelog.
- **Tests:** round-trip a dip-2 file with the retry attrs; migration of the #186 repro is now lossless (exit 0, attrs preserved); the fallback_target behavior per sub-decision 2.

## Why not B (engine reads edges)

Recorded in #204: B is a larger cross-repo runtime change that overloads each edge with two meanings (`loop` = success-loop AND retry-destination; `on fail` = genuine-failure AND retry-exhaustion), forcing two distinct triggers to share an edge — and retry fires on infra conditions that don't map onto graph edges at all. A is minimal and engine-aligned.

## Blocking questions for the tracker maintainer

1. What exact node field(s) does `processRetryOutcome` read for (a) the retry destination and (b) retry exhaustion? (`retry_target` / `fallback_target` / `fallback_retry_target`?)
2. Is v1 `fallback_target` retry-exhaustion, on-failure, or both? (Determines sub-decision 2 and whether dippin's current on-fail-edge migration is a bug.)
3. Any objection to A re-admitting these as dip-2 node attributes (vs B teaching the engine to read edges)?
