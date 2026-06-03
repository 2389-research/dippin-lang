---
title: "What's New in Dippin v0.17–v0.20"
date: "2026-04-17"
description: "Conditional nodes, workflow variables, shell-aware linting, per-node backend selection, and a Zed extension — four releases of toolchain improvements."
tagStyle: "release"
tagLabel: "RELEASE"
category: "Releases"
readTime: "6 min read"
related:
  - url: "conditional-edges.html"
    title: "Conditional Edges"
    summary: "Learn how to route between nodes based on runtime context."
  - url: "cost-estimation.html"
    title: "Cost Estimation"
    summary: "See how conditional nodes affect your pipeline cost estimates."
---

Four releases shipped in two weeks. Here's what changed and why.

## Conditional nodes: pure routing without LLM calls

**v0.17.0** &mdash; The biggest addition. Before this, routing decisions required an `agent` node with `auto_status: true` — which meant burning an LLM API call just to check a condition. Now there's a dedicated node kind:

```dippin
conditional CheckOutcome
  label: "Route by Result"
```

Conditional nodes have no prompt, no model, no provider. They evaluate outgoing edge conditions and route accordingly — zero token cost. In DOT export they render as diamond shapes, matching the visual convention for decision points.

The scaffold template (`dippin new conditional`) now uses the real `conditional` kind instead of the `agent` workaround.

## Workflow variables

**v0.20.0** &mdash; Parameterized workflows were always possible, but the variables lived outside the `.dip` file — you'd patch them into the DOT export manually. Now there's a `vars` block:

```dippin
workflow SemportRust
  goal: "Port $source_ref into $target_name"
  start: Start
  exit: Exit

  vars
    source_ref: "references/claude-agent-sdk-python/src"
    target_name: claude-agents-rs
    target_module: "claude-agents-rs/src/"
```

Vars export as DOT graph-level attributes (read by the runtime via `graph.Attrs`), round-trip through parse/format/export/migrate, and the parser catches duplicate keys. DOT migration now captures unknown graph attributes into `Workflow.Vars` instead of silently dropping them.

## Shell-aware tool linting

**v0.20.0** &mdash; DIP125 ("tool command binary not found on PATH") used to parse shell commands with regex and first-token extraction. This produced false positives on common patterns:

```
hint[DIP125]: tool command binary "COUNTER='.ai/count.txt'" not found on PATH
hint[DIP125]: tool command binary "count=$(cat" not found on PATH
```

Now the linter uses `mvdan.cc/sh/v3` (the parser behind `shfmt`) to walk the actual shell AST. Variable assignments, command substitutions, `command -v` checks, and preamble commands are all handled correctly. The exact script from the bug report now produces zero false positives.

## Per-node backend and working directory

**v0.19.0 / v0.19.1** &mdash; Two new fields on agent nodes:

```dippin
agent Worker
  backend: claude-code
  working_dir: .ai/worktrees/feature-branch
  prompt: Implement the feature.
```

`backend` selects the execution backend (native, claude-code, acp). `working_dir` sets an isolated working directory — essential for git-worktree-based parallel execution. Both were previously silently dropped by the parser, causing hours of debugging for users.

As part of this fix, **unrecognized fields on any node type now emit a parse diagnostic** instead of being silently discarded. The error message suggests `params:` for agent/subgraph nodes that support it.

## Model catalog extensibility

**v0.17.0** &mdash; The `--extra-models` flag on `lint` and `doctor` lets you extend the DIP108 model catalog at runtime:

```bash
dippin lint --extra-models "custom-corp:my-model-v1,my-model-v2" pipeline.dip
```

No more false DIP108 warnings for private or newly-released models.

## DIP134: max_retries vs max_restarts

**v0.20.0** &mdash; A new lint rule catches a common confusion. When you set `max_retries` in defaults and your workflow has `restart: true` edges but no `max_restarts`, the linter now warns you. `max_retries` controls per-node LLM retries; `max_restarts` controls the global loop restart budget. Setting the wrong one silently does nothing useful.

## Zed extension

The tree-sitter grammar now works with tree-sitter 0.24+, and there's a new Zed editor extension at `editors/zed-dippin/` with syntax highlighting and LSP integration. Install via Zed's "Install Dev Extension" command.

## Upgrading

```bash
go install github.com/2389-research/dippin-lang/cmd/dippin@v0.20.0
```

Or via Homebrew if you're on the tap.
