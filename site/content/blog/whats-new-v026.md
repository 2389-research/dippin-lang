---
title: "What's New in Dippin v0.26"
date: "2026-05-15"
description: "Workflow `requires:` keyword — declare what your workflow needs to run so runtimes can preflight, instead of crashing 20 minutes in."
tagStyle: "release"
tagLabel: "RELEASE"
category: "Releases"
readTime: "3 min read"
related:
  - url: "whats-new-v027.html"
    title: "What's New in v0.27"
    summary: "Model catalog refresh — 11+ new IDs, seven price corrections, and a retirement calendar."
  - url: "whats-new-v025.html"
    title: "What's New in v0.25"
    summary: ".dipx format v1.1 — real ctx cancellation, an inspect that actually inspects, and exit code 2 that matches the spec."
---

`v0.26.0` is a small release. One new keyword — `requires:` — in the workflow header. The shape is intentionally simple: a comma-separated list of identifiers the workflow needs in order to run. The motivation is less simple, and it shows up most visibly the first time a workflow crashes 20 minutes in because `git` wasn't on the runtime's PATH.

## The problem

Workflows quietly assume things about the environment they execute in. A pipeline that ends with `git commit && gh pr create` assumes `git` and `gh` exist. A workflow that calls a `jq` filter inside a tool node assumes `jq` is installed. An agent that uses an MCP filesystem server assumes the runtime has that server configured. None of those assumptions live in the `.dip` file — they live in the author's head, until they don't.

The only signal that a missing dependency is about to break a workflow has historically been the workflow itself: it runs, it burns through several agent turns, it generates the first `git commit` tool call, and that's the moment everything falls over. By then you've spent real money and real wall-clock time on a run that was doomed from the first agent.

## The shape

`requires:` is a workflow-header field. Comma-separated list. Same author-facing shape as `reads:` and `writes:` on node bodies:

```dippin
workflow BuildProduct
  goal: "Build a feature end-to-end."
  requires: git, docker, jq
  start: Plan
  exit: Done
```

Canonical order is `goal → requires → start → exit`. The formatter knows this; round-tripping a workflow through `dippin fmt` preserves your declaration and puts it in the right spot. The parser stores the list as `ir.Workflow.Requires []string` — entries trimmed, empties dropped, `nil` when the field is absent (matching the IR's nil-vs-empty conventions).

## The design decision

Dippin-lang parses and formats `requires:`. It does not interpret it. Whether `git` means "a binary on PATH," "an MCP server named git," or "an environment variable named `GIT_AUTHOR`" depends entirely on the runtime executing the workflow. The DSL stays runtime-agnostic; consumers decide what each token resolves to.

This is the same pattern as `tool_commands_allow` / `tool_denylist_add` in v0.23: dippin-lang owns the *shape* of the declaration, runtimes own the *meaning*. It's deliberate. Encoding "git means a binary on PATH" into the linter would couple a runtime-agnostic toolchain to one specific executor.

## What v1 doesn't do (and why)

Three things you might expect that aren't here yet:

1. **No lint validation.** Authors can declare any identifier. The linter doesn't warn about unknown ones, doesn't check whether you've declared a dep you don't actually use, doesn't flag mismatches between `requires:` and what tool nodes invoke. Adding any of that requires a vocabulary of recognized identifiers that doesn't exist yet — and writing one before there's usage signal would mean guessing wrong.

2. **No typed categories.** `requires: bin:git, env:GITHUB_TOKEN, mcp:filesystem` would be nice. It's also a commitment to a category system. The current bare-identifier form is a forward-compatible subset — typed entries can land later as a strict extension without rewriting anyone's `.dip` files.

3. **No enforcement.** A runtime that doesn't recognize a declared entry should warn-and-continue, not fail. This is explicit guidance for runtime authors: forward declarations against newer runtime versions should be a friction, not a wall.

v1 is the smallest thing that lets the next layer get built. Validation, typing, and enforcement come from the consumer side first — when there's enough real-world usage to know what shapes those features should take.

## The motivating use case

The motivating use case was a git-preflight check: a `--git=` preflight that reads `Workflow.Requires`, checks `git` exists on PATH before the first agent fires, and hard-fails in seconds when it doesn't — instead of burning $20–$100 of LLM spend before falling over at the first `git commit` tool call. Future preflights (`--docker=`, `--gh=`, …) can follow the same shape against the same field.

Any other runtime can read the same field and do the same kind of preflight. The contract is in the `.dip` file; the implementation is in whichever executor's hands.

## Editor support

Tree-sitter, VS Code, and Zed grammars all recognize `requires:` as a workflow header keyword. The hosted skill at `https://2389-research.github.io/dippin-lang/skill.md` is updated for LLM authors. If you author `.dip` files in an editor with LSP, `requires:` autocompletes after `goal:` and gets the same syntax highlighting as the other header fields.

## What's next

Full notes in [CHANGELOG.md](https://github.com/2389-research/dippin-lang/blob/main/CHANGELOG.md). When usage shapes settle, lint validation lands — and possibly typed categories. Until then, declare what your workflow needs, ship the `.dip` to whatever runtime, and let the runtime tell you in seconds whether the environment can actually run it.
