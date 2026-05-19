# AGENTS.md — dippin-lang

dippin-lang is a domain-specific language and toolchain for authoring AI pipeline workflows. It replaces Graphviz DOT as the authoring format for the Tracker runtime.

This file describes how AI coding agents (Claude Code, Cursor, Aider, etc.) can install, configure, and use the toolchain. For human-targeted docs see [`README.md`](README.md) and [`https://dippin.org`](https://dippin.org).

## Installation

```sh
go install github.com/2389-research/dippin-lang/cmd/dippin@latest
```

Or via Homebrew:

```sh
brew install 2389-research/tap/dippin
```

Verify:

```sh
dippin version
```

The CLI is a single static Go binary. No runtime dependencies beyond Go itself for `go install`.

## Configuration

Workflows live in `.dip` files. The toolchain has no global configuration file — every setting is per-workflow via the `defaults` block:

```dippin
workflow Example
  goal: "Demonstrate defaults"
  start: First
  exit: Done

  defaults
    model: claude-sonnet-4-6
    provider: anthropic
    fidelity: high
    retry: 2

  agent First
    prompt: Greet the user.

  agent Done
    prompt: Wrap up.

  edges
    First -> Done
```

Editor / agent integrations:

- **LSP**: `dippin lsp` starts the language server. Configure your editor to launch it for `*.dip` files.
- **Claude Code skill**: add `@https://dippin.org/skill.md` to a project's `CLAUDE.md` to teach Claude Code the syntax and DIP diagnostic codes.
- **VS Code / Zed extensions**: see [`editors/`](editors/) in this repo.

## Usage

```sh
dippin parse pipeline.dip          # JSON IR
dippin validate pipeline.dip       # structural errors (DIP001-DIP009)
dippin lint pipeline.dip           # semantic warnings (DIP101-DIP133)
dippin format pipeline.dip         # canonical formatting
dippin doctor pipeline.dip         # A-F health report
dippin cost pipeline.dip           # per-run cost estimate
dippin test pipeline.dip           # scenario tests from .test.json files
dippin simulate pipeline.dip       # walk every reachable path
dippin pack pipeline.dip           # produce a .dipx bundle
dippin unpack pipeline.dipx        # expand a bundle
```

Run `dippin --help` for the full command list (28+ commands).

## Building this site

The `https://dippin.org` Hugo site lives under [`site/`](site/). Build:

```sh
cd site && hugo --minify
```

The Netlify build (see [`netlify.toml`](netlify.toml)) additionally compiles the WASM playground and generates the changelog + LLM spec.

## a14y configuration

- Target URL: https://dippin.org/
- Scorecard: 0.2.0
- Mode: site
- Last runs:
  - 2026-05-19 — 72 (scorecard 0.2.0)
