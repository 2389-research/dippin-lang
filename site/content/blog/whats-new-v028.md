---
title: "What's New in Dippin v0.28"
date: "2026-05-19"
description: "Typed tool routing — marker_grep, route_required, and output_limit close a parser/runtime parity gap with the runtime's TRK101."
tagStyle: "release"
tagLabel: "RELEASE"
category: "Releases"
readTime: "3 min read"
related:
  - url: "whats-new-v027.html"
    title: "What's New in Dippin v0.27"
    summary: "Model catalog refresh — 11+ new IDs across six providers, seven price corrections, and a retirement calendar worth pinning somewhere."
  - url: "whats-new-v026.html"
    title: "What's New in v0.26"
    summary: "The workflow `requires:` keyword — declare environmental dependencies so runtimes can preflight."
---

Issue [#39](https://github.com/2389-research/dippin-lang/issues/39) closes a parser gap. The runtime already supported `marker_grep`, `route_required`, and `output_limit` for stdout-driven routing, but `.dip` source couldn't reach them — authors following the runtime's TRK101 lint recommendation hit `unrecognized tool field "marker_grep"` from dippin instead of a cleaner pipeline.

## Three new tool-node fields

```dippin
tool RunTests
  marker_grep: "^(tests_pass|tests_fail)$"
  route_required: true
  output_limit: 65536
  timeout: 2m
  command:
    go test ./... 2>&1
```

- **`marker_grep`** — regex matched against captured stdout. The last match populates `ctx.tool_marker`, so routing edges can reference it instead of parsing raw `ctx.tool_stdout`.
- **`route_required`** — when true, the node fails (`EventToolRouteMissing`) if the command emits no routing signal recognized by the runtime. The routing value populates `ctx.tool_route`.
- **`output_limit`** — per-node override of the engine's captured-stdout byte cap, for tools that need a larger or smaller window than the global default.

## Why it matters

Without these fields, authors working around the runtime's TRK101 truncation foot-gun had to enumerate every possible stdout marker as a conditional edge. That made `dippin coverage` flag false-positive non-exhaustive routing on nodes that were actually fully covered. Typed `marker_grep` is the canonical fix; now dippin accepts it.

## Runtime requirement

These fields forward through DOT export to the runtime. Routing semantics require the matching `extractToolAttrs` change in the runtime; older runtime versions silently ignore the new attrs.

## Coming next

- **[#42](https://github.com/2389-research/dippin-lang/issues/42)** — DIP138 lint, recommending `marker_grep` when a tool node parses stdout for routing without declaring it.
- **[#43](https://github.com/2389-research/dippin-lang/issues/43)** — normalize boolean parsing (`true` / `yes` / `1` accepted consistently).
- **[#44](https://github.com/2389-research/dippin-lang/issues/44)** — close the existing `outputs` DOT round-trip gap.

Full changelog: [CHANGELOG.md](https://github.com/2389-research/dippin-lang/blob/main/CHANGELOG.md).
