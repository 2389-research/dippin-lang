# #136 — Single-source `parallel`/`fan_in` (stop requiring edges-block re-declaration)

**Status:** design approved 2026-07-09
**Issue:** [#136](https://github.com/2389-research/dippin-lang/issues/136), Phase 1 of routing epic [#127](https://github.com/2389-research/dippin-lang/issues/127)

## Problem

A `parallel` node declares its fan-out targets inline:

```dippin
  parallel DesignFan -> DesignClaude, DesignGPT, DesignGemini
```

…and workflows routinely **re-declare the same edges** in the `edges` block:

```dippin
  edges
    DesignFan -> DesignClaude
    DesignFan -> DesignGPT
    DesignFan -> DesignGemini
```

Same for `fan_in X <- a, b, c`. The fork is spelled twice and must be kept in sync by hand.

## Investigation verdict (Step 1 — the actual ask, confirmed)

The inline node config is **authoritative** for every semantic consumer; the edges-block re-declaration is functionally redundant:

| Consumer | Authoritative source | Edges-block re-declaration |
|---|---|---|
| Validator DIP004 (reachability) — `addParallelFanInEdges` | inline config | ignored |
| Validator DIP007 (fan matching) — `collectParallelFanIn` | inline config | ignored |
| Simulator fan traversal — `resolveTargets`/`findJoinNode` | inline config | ignored |
| `ir.EdgesFrom()` — `parallelEdgesFrom`/`fanInEdgesFrom` | **both, deduped**: an explicit *unconditional* edge suppresses the synthesized implicit one | redundant |
| Formatter node line | inline config | emitted verbatim (no synthesis) |
| **DOT export** — `export/dot.go:52` iterates raw `w.Edges` | **edges block only; never synthesizes fan edges from config** | **load-bearing here** |

**Conclusions:**

1. Stripping the redundant edges-block lines does **not** change reachability, fan matching, simulation, or `EdgesFrom()` output (`EdgesFrom` already synthesizes+dedupes the fan edges from config).
2. The **one** consumer that would visibly regress is **DOT export**: it draws fan-out arrows only because they appear in `w.Edges`. It must be taught to derive fan edges from node config, or stripped `.dip` files render without fork arrows. This is a prerequisite fix.
3. A re-declared edge **with a condition or routing attribute is not redundant** — `EdgesFrom` keeps it alongside the implicit unconditional edge, and it carries author intent. Redundancy detection must be limited to **unconditional, attribute-free** re-declarations.

## Design

### Shared predicate (DRY)

Add to `ir` (mirroring #134's `ir.EdgeRoutesOnFail`):

```go
// IsRedundantFanEdge reports whether e merely repeats a parallel/fan_in fork
// already declared inline on a node's config, carrying no extra information:
// it is unconditional and attribute-free, and either
//   - From is a parallel node and To is one of its Targets, or
//   - To is a fan_in node and From is one of its Sources.
func IsRedundantFanEdge(w *Workflow, e *Edge) bool
```

"Attribute-free" = `e.Condition == nil` **and** no `Label`, `Choice`, `Weight`, `Override`, or any other routing/display attribute set. If the edge carries *any* such attribute it is NOT redundant (stripping would lose information).

All three consumers below call this one predicate — no drift.

### 1. DOT export fix (prerequisite — must land first)

`export/dot.go` gains a second emission pass: after emitting the explicit `w.Edges`, synthesize the fan edges from `ParallelConfig.Targets` / `FanInConfig.Sources` that are **not already present** as an explicit edge (same dedup rule as `ir.parallelEdgesFrom`/`fanInEdgesFrom`: skip a target/source already covered by an unconditional explicit edge). Result: fork arrows render whether or not the author re-declared them. Idempotent and additive — no double-draw when both forms are present today.

### 2. New lint — `DIP153` (Warning): "redundant parallel/fan_in edge"

Fires once per edges-block edge for which `ir.IsRedundantFanEdge(w, e)` is true. Message names the node and target/source and points at `fmt`:
> `DIP153: edges-block edge 'DesignFan -> DesignClaude' redundantly repeats the inline parallel fan-out; the inline list is authoritative — run 'dippin fmt' to remove it (rejected under 'dip 2')`

Surfaces in lint / check / watch / doctor (the standard four paths). Advisory only under v1 — never an error.

### 3. `fmt` strips redundant fan edges

The formatter's edge-emission pass skips any edge for which `ir.IsRedundantFanEdge(w, e)` holds. `dippin fmt --write` therefore removes the redundant lines; the inline `parallel`/`fan_in` line is the sole surviving declaration. Idempotent (a second `fmt` is a no-op) and deterministic. Conditional/attributed re-declarations are preserved.

### 4. `dip 2` rejection

Under a `dip 2` header, a redundant fan-edge re-declaration is an **error**, consistent with #134's version-gated rejection of `retry_target`/`fallback_target`. Emitted as a parser post-parse diagnostic (the parser has the full workflow after node+edge parsing), pointing at the offending edge with remediation "the inline `parallel`/`fan_in` list is authoritative under `dip 2` — remove the redundant edge (run `dippin fmt`)". Reuses `ir.IsRedundantFanEdge`.

### 5. Examples

Run `dippin fmt --write` over the examples that re-declare fan edges (`fanin_policy.dip`, `consensus_task_parity.dip`, and any others) so they carry no DIP153 warning. They remain v1. `TestLintExamples` gains a guard asserting zero DIP153 across the example suite (mirrors the DIP108 guard).

## Non-goals

- No change to fan-out/fan-in **semantics** or the `params:`/`branch:` block forms (per issue).
- No grammar change — the edges-block syntax is unchanged; redundancy is a semantic property, so tree-sitter/VS Code/Zed grammars are untouched.
- Not touching conditional or attributed edges between a parallel/fan_in node and its targets/sources — those stay.

## Documentation & tooling sweep (standing directive)

- `docs/nodes.md`, `docs/edges.md` — document that the inline `parallel`/`fan_in` list is the single source of truth; edges-block re-declaration is redundant (DIP153) and rejected under `dip 2`.
- `docs/validation.md` + `validator` explanation prose — add DIP153. (`site/content/validation.md` is hand-maintained; surface DIP153 there at release time per `release-process`.)
- `site/content/cli.md` + `docs/cli.md` — note `fmt` strips redundant fan edges; update the DIP catalog count (62 → 63).
- Regenerate embedded spec (`cmd/dippin/generated-spec.md` via `scripts/gen-spec.sh`).
- No editor-grammar change (semantic-only).

## Acceptance criteria

- [x] Written confirmation of the authoritative declaration (this document).
- [ ] `export/dot.go` synthesizes fan edges from node config; DOT of a stripped file still draws fork arrows (regression test).
- [ ] `ir.IsRedundantFanEdge` shared predicate with unit tests (unconditional-match = true; conditional/attributed/non-fan = false).
- [ ] DIP153 fires on redundant fan edges, is silent on conditional/attributed ones, surfaces in lint/check/watch/doctor.
- [ ] `fmt` strips redundant fan edges; idempotent; preserves conditional/attributed edges.
- [ ] `dip 2` rejects redundant fan-edge re-declaration with a clear diagnostic.
- [ ] Examples stripped; `TestLintExamples` asserts zero DIP153.
- [ ] Docs/site/spec swept; catalog count updated.

## Test fixtures (from issue)

Use `dev_loop.dip` squad fan-out / `build_product.dip` review fan-out if present; otherwise `fanin_policy.dip` and `consensus_task_parity.dip` (confirmed to carry the redundant pattern).
