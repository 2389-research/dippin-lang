# Design: node `params:` on `parallel` / `fan_in` (issue #110)

## Goal

Let `parallel` and `fan_in` nodes carry a generic `Params map[string]string`,
mirroring `AgentConfig` / `SubgraphConfig`, so a workflow can express a fan-in
aggregation policy (e.g. `fan_in_policy: all`, `quorum: 2`, `required_branches: ...`)
in `.dip`.

**Carry-only.** dippin stores + round-trips the keys; the engine (tracker #313)
reads the keys it needs. No engine semantics, no policy-key validation, no new DIP
code, no tracker dependency (detection/carry, not enforcement).

## Scope decision

Node-level `Params` is a `.dip`-native concept. Empirically, `AgentConfig.Params`
and `SubgraphConfig.Params` round-trip **only through the formatter** (`.dip` ↔
`.dip`): they are not serialized to DOT (`export/dot.go`), not read by `migrate`,
and not compared by `migrate/parity.go`. (The #94 "silent DOT drop" was graph-level
**budgets** in `WorkflowDefaults`, which genuinely export to DOT — a different
layer.)

Decision (confirmed with maintainer): **mirror agent/subgraph exactly.** params
round-trips through the formatter only; DOT export and migrate do not carry node
params. Parity is unchanged. This keeps parallel/fan_in consistent with the
existing node-params precedent and avoids scope creep.

## Layers

1. **IR** (`ir/ir.go`): add `Params map[string]string` to `ParallelConfig` and
   `FanInConfig`. Initialize to `make(map[string]string)` at construction sites
   (parser node creation) the same way agent/subgraph do.
2. **Parser** (`parser/parse_nodes.go`) — three paths, all reusing
   `parseParamsBlock`:
   - **block parallel** (`parseParallelBranches`): recognize a `params:` line
     (today non-`branch:` lines are silently skipped).
   - **inline parallel** (`parallel P -> A, B`): after the target list, accept an
     OPTIONAL indented block that may contain `params:` (today `expect(TokenNewline)`
     terminates the node).
   - **fan_in** (`parseFanIn`): give it the same OPTIONAL indented `params:` block
     (it has no block body today).
3. **Formatter** (`formatter/format.go`): emit `params:` via the existing
   `writeSortedMapBlock` for `ParallelConfig` (both inline and block form) and
   `FanInConfig`. Inline parallel with params must emit the indented block beneath
   the inline line.
4. **tree-sitter** (`editors/tree-sitter-dippin/grammar.js`): extend
   `parallel_node` / `fan_in_node` to accept the optional params block; regenerate
   + corpus tests. (Local `tree-sitter generate` is unavailable — rely on the
   pre-commit hook + CI tree-sitter job.)

## Out of scope

- DOT export / migrate / parity changes (mirror agent/subgraph = no node params there).
- Policy-key validation / new DIP code (carry-only, like agent `params:`).
- Any engine or tracker behavior.

## Verification (TDD, red first)

1. parser: block `parallel` with `params:` → `ParallelConfig.Params` populated.
2. parser: inline `parallel` with trailing `params:` block → Params populated
   (currently a fatal parse error).
3. parser: `fan_in` with `params:` block → `FanInConfig.Params` populated
   (currently fatal).
4. round-trip: parse → format → re-parse preserves Params for both kinds (formatter).
5. tree-sitter corpus tests for the new grammar.
6. an example `.dip` exercising `fan_in` params; `just validate-examples` /
   `just lint-examples` stay green.
7. DIP007 (parallel/fan_in mismatch) regression — still fires/doesn't-fire with the
   new block body.

Downstream consumer: tracker #313 (engine fan-in aggregation policy).
