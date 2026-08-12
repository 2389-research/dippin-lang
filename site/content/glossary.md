---
title: "Glossary"
description: "Definitions for Dippin terms: workflow, node, edge, condition, defaults, subgraph, fan_in, parallel, .dipx, DIP codes, and more."
section_label: "Reference"
subtitle: "Terms you'll encounter authoring .dip files and using the dippin toolchain."
---

## Core concepts

**Workflow** — A directed graph of nodes and edges defined in a single `.dip` file. Has a name, goal, start node, and exit node. One workflow per file.

**Node** — A step in the pipeline. Every node has an ID and a kind. Kinds include `agent`, `human`, `tool`, `conditional`, `parallel`, `fan_in`, `subgraph`, and `manager_loop`.

**Edge** — A connection between two nodes, written `A -> B` in the `edges` block. May carry a `when` condition (or the `on <token>` shorthand), a `loop` keyword for back-edges, a `label:`, a `choice:` routing key, a `restart:` flag, and an `override:` flag. (`weight:` is still parsed but soft-deprecated — DIP151.)

**Condition** — A predicate attached to an edge using the `when` keyword. Syntax: `ctx.<field> <op> <value>` with operators `=`, `==` (synonym for `=`), `!=`, `contains`, `not contains`, `startswith`, `endswith`, and `in`. Complementary pairs (success/fail, contains/not-contains) are detected as exhaustive automatically.

**on / loop** — Edge shorthand keywords. `on <token>` desugars into an equality test against the source node's outcome channel (`ctx.outcome` for agents, `ctx.tool_marker` for tools with `marker_grep`) — equivalent to the matching `when`. `loop` marks a back-edge (a deliberate cycle) so reachability lint treats it as intentional.

**else** — A section-level funnel default in the `edges` block, written `else -> <NodeID>`. Provides the success-side default route when no other outbound edge from the current section matches. At most one `else` per edges block; it takes no source node and no attributes (#154 / #157).

**choice:** — An edge attribute carrying a human-gate routing key, written `choice: <key>`. The value is carried, not interpreted — it marks the load-bearing routing key so a `label:` can stay display-only (DIP150). Distinct from a `human` node's `mode: choice`, which selects the gate's interaction style.

**marker_grep / ctx.tool_marker** — `marker_grep` is a regex on a `tool` node, matched line-by-line against the command's stdout; the matched value populates `ctx.tool_marker`, the tool's outcome channel for routing (e.g. via `on` or `when`).

**inputs** — A workflow-level block declaring what a caller must supply (the callee-side signature), whether a human at the entry point or a parent via a subgraph's `params:`. Each entry is `name: type` (`text`, `number`, `bool`, `enum`, `file`, `secret`) with optional attributes (`required`, `default`, `prompt`, `description`, `options`, `pattern`, `min`, `max`, `max_length`, `multiline`). Read in prompts/conditions as `${inputs.name}` — a **closed** namespace: a reference to an undeclared input is a lint error (DIP156), unlike the open `ctx`. Introspect with `dippin inputs`.

**Defaults** — A workflow-wide block setting values inherited by all nodes unless overridden. Common keys: `model`, `provider`, `fidelity`, `retry`, `tool_commands_allow`, `tool_denylist_add`.

**Goal** — A short string at the top of a workflow stating the pipeline's purpose. Surfaces in `dippin doctor` reports and is shown to operators.

**Start node** — The single node executed first when the workflow runs. Declared via `start:` in the workflow header.

**Exit node** — The single node that terminates the workflow. Declared via `exit:` in the workflow header. All paths must reach it (lint warns otherwise).

## Node kinds

**agent** — An LLM call. Has a `prompt:` (single line or multiline indented block), optional `model:`, `provider:`, `temperature:`, `tools:`, and `outputs:`.

**human** — A gate that requires user input. Has a `mode:` (`choice`, `freeform`, `interview`, `yes_no`) and an optional `timeout:`.

**tool** — A shell command. Has a `command:` and optional `timeout:`. Constrained by `tool_commands_allow` / `tool_denylist_add` defaults.

**conditional** — A branching node with multiple outbound edges. Routes based on context fields without an LLM call.

**parallel** — Fans out execution to N children that run concurrently. Pairs with `fan_in`.

**fan_in** — Joins parallel branches. Has a `strategy:` (`all`, `any`, `quorum`) and an optional `quorum:` count.

**subgraph** — Embeds another `.dip` file by reference. Uses `ref:` to point to the target file. The `flatten` package resolves these into a single flat workflow for export.

**manager_loop** — A supervisory loop node that runs a child workflow repeatedly under a budget, surfaced in v0.21–v0.22.

## File and bundle formats

**.dip** — The source file extension. Plain text, indentation-based, parsed by the `dippin` CLI.

**dip N** — An optional leading format-version declaration (e.g. `dip 1`) at the top of a `.dip` file. Defaults to version 1 when absent.

**.dipx** — A bundle format introduced in v0.24 that packs a workflow tree (root .dip + referenced subgraphs) into a single verifiable file. Created with `dippin pack`, expanded with `dippin unpack`.

**.test.json** — A scenario test file. Injects context values, asserts on execution paths, and reports edge coverage. Run with `dippin test`.

## Diagnostics

**DIP code** — A diagnostic identifier emitted by `dippin lint` and `dippin validate`. DIP001–DIP010 are structural errors (parse / shape failures). DIP101–DIP162 are semantic warnings (style, correctness hints, model catalog issues).

**Exhaustive conditions** — A pair of edge conditions detected as covering all outcomes (e.g., `when ctx.status = success` + `when ctx.status = fail`). Suppresses DIP101 / DIP102 warnings.

**Lint** — `dippin lint` runs semantic checks and emits DIP1xx codes. Does not exit non-zero by default; CI typically pipes through `--fail-on warning`.

**Validate** — `dippin validate` runs structural checks and emits DIP0xx codes. Exits non-zero on any error.

## Commands

**parse** — Parse a `.dip` file into the JSON IR for downstream tooling.

**format** — Rewrite a `.dip` file in canonical form (2-space indent, sorted keys where order is insignificant).

**doctor** — Grade a workflow A–F. Combines lint, coverage, unused detection, and cost into a health report.

**cost** — Estimate per-run cost based on the model catalog and prompt sizes. Pricing is verified against provider docs and dated in source.

**optimize** — Suggest cheaper model alternatives that preserve the workflow's structural shape.

**simulate** — Walk every reachable path of a workflow without making real LLM calls. Used by `test` under the hood.

**migrate** — Convert Graphviz DOT pipeline files to Dippin. Pairs with `validate-migration` to verify structural parity.

**pack / unpack / inspect** — Produce, expand, and introspect `.dipx` bundles.

## Toolchain

**IR** — The intermediate representation. A set of Go structs in the `ir` package that all downstream consumers program against.

**LSP** — The Dippin Language Server Protocol implementation under `lsp/`. Provides hover docs, go-to-definition, completions, and live diagnostics in any LSP-aware editor.

**WASM playground** — The browser-side parse / lint / format experience at `/playground.html`. The `dippin` core compiled to WebAssembly under `cmd/wasm/`.

**Runtime** — A consuming system that executes Dippin workflows. The runtime consumes the JSON IR and is responsible for actually running nodes, managing state, and enforcing runtime-side constraints such as `writable_paths`.

## Related

- [Language Reference]({{< relref "language" >}}) — full syntax for every construct above.
- [CLI Reference]({{< relref "cli" >}}) — every command.
- [Validation & Linting]({{< relref "validation" >}}) — full list of DIP codes.
