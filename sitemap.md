# Dippin sitemap

A flat index of every page on this site, grouped for agent consumption. The canonical machine-readable sitemap is [sitemap.xml](https://dippin.org/sitemap.xml); this `.md` mirror exists for LLM clients that prefer Markdown.

## Documentation

- [Analysis Tools](https://dippin.org/analysis.html) — Cost estimation, health reports, edge coverage, dead branch detection, semantic diff, and optimization for AI pipelines.
- [Architecture](https://dippin.org/architecture.html) — How the Dippin toolchain is organized: IR-centric design, package dependencies, and the parser-to-execution pipeline.
- [Changelog](https://dippin.org/changelog.html) — Version history and release notes for dippin-lang.
- [CLI Reference](https://dippin.org/cli.html) — Complete command reference for the Dippin toolchain: commands for authoring, export, analysis, introspection, and bundling AI pipeline workflows — parse, validate, lint, check, fmt, simulate, cost, coverage, doctor, test, watch, inputs, pack, unpack, inspect, and more.
- [Editor Setup](https://dippin.org/editors.html) — Set up syntax highlighting, LSP diagnostics, hover docs, and go-to-definition for .dip files in VS Code, Neovim, and more.
- [Glossary](https://dippin.org/glossary.html) — Definitions for Dippin terms: workflow, node, edge, condition, defaults, subgraph, fan_in, parallel, .dipx, DIP codes, and more.
- [Language Reference](https://dippin.org/language.html) — Full syntax reference for .dip workflow files: nodes, edges, conditions, multiline prompts, parallel execution, and stylesheets.
- [Playground](https://dippin.org/playground.html) — Try Dippin in your browser. Lint, format, visualize, and health-check workflows live — nothing leaves your machine.
- [Scenario Testing](https://dippin.org/testing.html) — Write deterministic tests for AI pipelines with .test.json files. Inject context, assert on paths, check edge coverage.
- [Validation & Linting](https://dippin.org/validation.html) — 70 diagnostic codes for AI pipeline workflows. 10 structural errors and 60 semantic checks catch bugs before runtime.

## Blog

- [What's New in Dippin v0.28](https://dippin.org/blog/whats-new-v028.html) — Typed tool routing — marker_grep, route_required, and output_limit close a parser/runtime parity gap with the runtime's TRK101.
- [What's New in Dippin v0.27](https://dippin.org/blog/whats-new-v027.html) — Model catalog refresh — 11+ new IDs across six providers, seven price corrections, and a retirement calendar worth pinning somewhere.
- [What's New in Dippin v0.26](https://dippin.org/blog/whats-new-v026.html) — Workflow `requires:` keyword — declare what your workflow needs to run so runtimes can preflight, instead of crashing 20 minutes in.
- [What's New in Dippin v0.25](https://dippin.org/blog/whats-new-v025.html) — .dipx format v1.1 — real cancellation through Pack and Open, an inspect that actually inspects, and exit code 2 that actually fires when the spec says it should.
- [What's New in Dippin v0.24](https://dippin.org/blog/whats-new-v024.html) — The .dipx bundle format — pack a workflow tree into one verifiable file you can ship anywhere. Three new commands, every analysis command extended.
- [What's New in Dippin v0.21–v0.22](https://dippin.org/blog/whats-new-v021-v022.html) — Human timeouts, budget caps, and a new manager_loop node for supervising child pipelines. Two releases, five days apart.
- [What's New in Dippin v0.23](https://dippin.org/blog/whats-new-v023.html) — First-class tool_commands_allow and tool_denylist_add defaults for constraining what tool nodes can shell out. Plus a cleaner DOT header format.
- [What's New in Dippin v0.17–v0.20](https://dippin.org/blog/whats-new-v020.html) — Conditional nodes, workflow variables, shell-aware linting, per-node backend selection, and a Zed extension — four releases of toolchain improvements.
- [Conditional Edges: Routing Pipelines with when](https://dippin.org/blog/conditional-edges.html) — Build branching AI pipelines that route based on LLM output. Learn Dippin's condition syntax, operators, and exhaustive detection.
- [Cost Estimation: Know Before You Run](https://dippin.org/blog/cost-estimation.html) — Estimate per-run pipeline costs before spending real money on LLM calls. Use dippin cost and dippin optimize to find savings.
- [Multi-line Prompts Without Escaping](https://dippin.org/blog/multi-line-prompts.html) — DOT's escaped strings are unreadable. Dippin's indentation-based blocks let you write prompts with markdown, JSON, and code blocks — no escaping required.
- [CI Integration: Lint, Test, Format](https://dippin.org/blog/ci-integration.html) — Set up GitHub Actions to validate, lint, test, and format-check your Dippin workflow files on every push and pull request.
- [Editor Setup: LSP, VS Code, and Tree-sitter](https://dippin.org/blog/editor-setup.html) — Set up real-time Dippin diagnostics, hover docs, and syntax highlighting in VS Code, Neovim, Helix, or any editor with LSP support.
- [Getting Started with Dippin](https://dippin.org/blog/getting-started.html) — Install Dippin, write your first AI pipeline workflow, and validate it in under 10 minutes. A step-by-step guide from zero to a working .dip file.
- [Migrating from DOT to Dippin](https://dippin.org/blog/migrating-from-dot.html) — Convert Graphviz DOT pipeline files to Dippin with automated migration and structural parity verification. Step-by-step with real examples.
- [Scenario Testing with .test.json](https://dippin.org/blog/scenario-testing.html) — Write deterministic tests for non-deterministic AI pipelines. Inject context values, assert on execution paths, and measure edge coverage.

## Agent skills

- [Claude Code skill](https://dippin.org/skill.md) — instructions for Claude Code agents working with `.dip` files.
- [AGENTS.md](https://dippin.org/AGENTS.md) — install / configure / use the toolchain.
- [llms.txt](https://dippin.org/llms.txt) — short LLM-oriented index.
- [llms-full.txt](https://dippin.org/llms-full.txt) — full spec for deeper context.
