---
title: "What's New in Dippin v0.23"
date: "2026-04-22"
description: "First-class tool_commands_allow and tool_denylist_add defaults for constraining what tool nodes can shell out. Plus a cleaner DOT header format."
tagStyle: "release"
tagLabel: "RELEASE"
category: "Releases"
readTime: "3 min read"
related:
  - url: "whats-new-v021-v022.html"
    title: "What's New in v0.21–v0.22"
    summary: "Human timeouts, budget caps, and the manager_loop node kind."
  - url: "whats-new-v020.html"
    title: "What's New in v0.17–v0.20"
    summary: "Conditional nodes, vars, shell-aware linting."
---

`v0.23.0` is a small, security-flavored release. Two new `defaults:` fields give you first-class control over what tool nodes are allowed to shell out — something you previously could only reach by dropping to raw DOT or patching the `Workflow.Vars` map in the runtime directly.

## The problem

Dippin `tool` nodes execute real shell commands. The runtime reads an allowlist and denylist at start time to decide which commands are permitted. Until this release, those lists lived in `graph.Attrs` — DOT-level attributes. If you wanted to set them from a `.dip` file, your options were:

1. Export to DOT, hand-edit the graph attributes, and run the runtime from the DOT file.
2. Patch `Workflow.Vars` programmatically via the runtime's library API.

Both defeat the purpose of `.dip` as the authoring format. The feature requested by [#28](https://github.com/2389-research/dippin-lang/issues/28) was simple: accept these two keys in the `defaults:` block, same as every other workflow-level setting.

## The new fields

```dippin
workflow ToolSafety
  goal: "Constrained shell execution"
  start: Start
  exit: Done

  defaults
    model: claude-sonnet-4-6
    # Glob allowlist — commands matching ANY pattern are permitted
    tool_commands_allow: "git *,make *,npm test,npm run *"
    # Patterns appended to the runtime's built-in default denylist
    tool_denylist_add: "rm -rf /,dd if=*"

  tool RunTests
    timeout: 60s
    command:
      set -eu
      npm test

  # ...
```

Both values are comma-separated glob patterns. Dippin-lang itself doesn't validate the glob syntax and stores the string verbatim — the runtime owns splitting and matching semantics, so there's no parse-time normalization that could drift from the runtime's parser.

Both fields round-trip through parse → format → DOT export → migrate, so DOT-authored workflows and migrations from legacy DOT files carry them through unchanged. See `examples/tool_safety.dip` for a working example.

## DOT header cleanup

Landing this feature required tightening how `ExportDOT` emits graph-level attributes. Previously the header wrote each attribute as a bare statement:

```
digraph ToolSafety {
  rankdir=TB;
  node [fontname="Helvetica"];
  tool_commands_allow="git *,make *";
  ...
}
```

Valid DOT, but the migrate parser only accepts attributes inside a `graph [...]` block. Round-tripping a `.dip` through export → Migrate → reparse would silently drop the attributes. The header now emits a single block:

```
digraph ToolSafety {
  node [fontname="Helvetica"];
  edge [fontname="Helvetica"];
  graph [rankdir=TB, tool_commands_allow="git *,make *"];
  ...
}
```

Still valid DOT. Graphviz consumers are unaffected. If you have tooling that grep's the old bare-statement form — that's where the breakage lives; switch to parsing the `graph [...]` block.

## Migration note

If you previously smuggled `tool_commands_allow` or `tool_denylist_add` through `vars:` (because `defaults:` rejected them), move them to `defaults:` now. These keys are reserved in the exporter starting with this release, so values in `vars:` are silently dropped from DOT output. The workaround was never documented — but if you found it, the fix is a one-line move.

## What's next

Full notes in [CHANGELOG.md](https://github.com/2389-research/dippin-lang/blob/main/CHANGELOG.md). The runtime-side allowlist is shipped; the denylist runtime is pending. Once that lands, `tool_denylist_add` gets teeth — until then, the field parses and round-trips but the runtime only enforces its built-in defaults.
