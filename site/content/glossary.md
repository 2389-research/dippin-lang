---
title: "Glossary"
description: "Definitions for Dippin terms: workflow, node, edge, condition, defaults, subgraph, fan_in, parallel, .dipx, DIP codes, and more."
section_label: "Reference"
subtitle: "Terms you'll encounter authoring .dip files and using the dippin toolchain."
---

## Core concepts

**Workflow** — A directed graph of nodes and edges defined in a single `.dip` file. Has a name, goal, start node, and exit node. One workflow per file.

**Node** — A step in the pipeline. Every node has an ID and a kind. Kinds include `agent`, `human`, `tool`, `conditional`, `parallel`, `fan_in`, `subgraph`, and `manager_loop`.

**Edge** — A connection between two nodes, written `A -> B` in the `edges` block. May carry a `when` condition, a label, a weight, and a `restart:` flag.

**Condition** — A predicate attached to an edge using the `when` keyword. Syntax: `ctx.<field> <op> <value>` with operators `=`, `!=`, `contains`, and `not contains`. Complementary pairs (success/fail, contains/not-contains) are detected as exhaustive automatically.

**Defaults** — A workflow-wide block setting values inherited by all nodes unless overridden. Common keys: `model`, `provider`, `fidelity`, `retry`, `tool_commands_allow`, `tool_denylist_add`.

**Goal** — A short string at the top of a workflow stating the pipeline's purpose. Surfaces in `dippin doctor` reports and is shown to operators.

**Start node** — The single node executed first when the workflow runs. Declared via `start:` in the workflow header.

**Exit node** — The single node that terminates the workflow. Declared via `exit:` in the workflow header. All paths must reach it (lint warns otherwise).

## Node kinds

**agent** — An LLM call. Has a `prompt:` (single line or multiline indented block), optional `model:`, `provider:`, `temperature:`, `tools:`, and `outputs:`.

**human** — A gate that requires user input. Has a `mode:` (`freeform`, `choice`, `confirm`) and an optional `timeout:`.

**tool** — A shell command. Has a `command:` and optional `timeout:`. Constrained by `tool_commands_allow` / `tool_denylist_add` defaults.

**conditional** — A branching node with multiple outbound edges. Routes based on context fields without an LLM call.

**parallel** — Fans out execution to N children that run concurrently. Pairs with `fan_in`.

**fan_in** — Joins parallel branches. Has a `strategy:` (`all`, `any`, `quorum`) and an optional `quorum:` count.

**subgraph** — Embeds another `.dip` file by reference. Uses `ref:` to point to the target file. The `flatten` package resolves these into a single flat workflow for export.

**manager_loop** — A supervisory loop node that runs a child workflow repeatedly under a budget, surfaced in v0.21–v0.22.

## File and bundle formats

**.dip** — The source file extension. Plain text, indentation-based, parsed by the `dippin` CLI.

**.dipx** — A bundle format introduced in v0.24 that packs a workflow tree (root .dip + referenced subgraphs) into a single verifiable file. Created with `dippin pack`, expanded with `dippin unpack`.

**.test.json** — A scenario test file. Injects context values, asserts on execution paths, and reports edge coverage. Run with `dippin test`.

## Diagnostics

**DIP code** — A diagnostic identifier emitted by `dippin lint` and `dippin validate`. DIP001–DIP009 are structural errors (parse / shape failures). DIP101–DIP133 are semantic warnings (style, correctness hints, model catalog issues).

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

**Tracker** — The runtime that executes Dippin workflows. Consumes the JSON IR and installs `dippin` via `go install ...@latest`.

## Related

- [Language Reference]({{< relref "language" >}}) — full syntax for every construct above.
- [CLI Reference]({{< relref "cli" >}}) — every command.
- [Validation & Linting]({{< relref "validation" >}}) — full list of DIP codes.
