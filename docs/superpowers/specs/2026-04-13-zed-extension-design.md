# Zed Extension for Dippin — Design Spec

## Goal

Add a Zed editor extension at `editors/zed-dippin/` providing syntax highlighting (via the existing tree-sitter grammar) and LSP integration (via `dippin lsp`).

## Prerequisite

The tree-sitter grammar at `editors/tree-sitter-dippin/` needs `conditional` added to `grammar.js`, `highlights.scm`, and the test corpus — it was added to the Go parser in v0.17.0 but the tree-sitter grammar wasn't updated.

## Architecture

The Zed extension is a thin packaging layer:

- **Syntax highlighting**: Reuses the existing tree-sitter grammar (`grammar.js` + `scanner.c`). The `highlights.scm` is copied into `languages/dippin/` since Zed expects it at that path.
- **LSP**: A minimal Rust shim (~20 lines) implements the `zed_extension_api::Extension` trait. The `language_server_command` method uses `worktree.which("dippin")` to find the binary and returns `dippin lsp` as the command.
- **Grammar source**: `extension.toml` points at the tree-sitter-dippin directory via the GitHub repo URL + commit SHA. For local dev, users can temporarily use a `file://` URL.

## Files

```text
editors/zed-dippin/
  extension.toml                    # Extension metadata + grammar + LSP declaration
  Cargo.toml                        # Rust crate for WASM compilation
  src/
    dippin.rs                       # Extension trait impl (language_server_command)
  languages/
    dippin/
      config.toml                   # Language config (name, suffixes, comments, brackets)
      highlights.scm                # Copy of tree-sitter-dippin/queries/highlights.scm
      indents.scm                   # Indentation queries for Zed
```

## File Details

### extension.toml

```toml
id = "dippin"
name = "Dippin"
version = "0.0.1"
schema_version = 1
authors = ["2389 Research"]
description = "Dippin workflow language support."
repository = "https://github.com/2389-research/dippin-lang"

[grammars.dippin]
repository = "https://github.com/2389-research/dippin-lang"
path = "editors/tree-sitter-dippin"
commit = "<current HEAD SHA>"

[language_servers.dippin-lsp]
languages = ["Dippin"]
```

### Cargo.toml

```toml
[package]
name = "zed_dippin"
version = "0.0.1"
edition = "2021"
publish = false
license = "MIT"

[lib]
path = "src/dippin.rs"
crate-type = ["cdylib"]

[dependencies]
zed_extension_api = "0.7.0"
```

### src/dippin.rs

Minimal: `new()` + `language_server_command()` that finds `dippin` on PATH and returns `Command { command: path, args: ["lsp"], env: [] }`.

### languages/dippin/config.toml

```toml
name = "Dippin"
grammar = "dippin"
path_suffixes = ["dip"]
line_comments = ["# "]
tab_size = 2
brackets = [
  { start = "(", end = ")", close = true, newline = false },
  { start = "\"", end = "\"", close = true, newline = false, not_in = ["string", "comment"] },
]
```

### languages/dippin/highlights.scm

Copy of `editors/tree-sitter-dippin/queries/highlights.scm` (after adding `conditional`).

### languages/dippin/indents.scm

Zed indent queries for the indentation-sensitive blocks. The tree-sitter grammar uses external INDENT/DEDENT tokens from `scanner.c`, so the indent queries mark which nodes get `@indent` / `@end`:

```scheme
(workflow_decl) @indent
(agent_node) @indent
(human_node) @indent
(tool_node) @indent
(subgraph_node) @indent
(conditional_node) @indent
(defaults_section) @indent
(edges_section) @indent
(stylesheet_section) @indent
(stylesheet_rule) @indent
(multiline_block) @indent
```

## Tree-sitter Grammar Updates (Prerequisite)

### grammar.js

Add `conditional_node` rule (same pattern as other node kinds):
```js
conditional_node: ($) =>
  seq("conditional", $.identifier, $._newline, $._indent, repeat1($.node_field), $._dedent),
```

Add to `node_decl` choice and add `"conditional"` keyword to extras.

### highlights.scm

Add `conditional` to the node kinds list and add identifier capture:
```scheme
(conditional_node
  (identifier) @function)
```

### test/corpus/basic.txt

Add a test case for conditional node parsing.

## Install

For development: Zed > Command Palette > "Install Dev Extension" > select `editors/zed-dippin/`. Zed compiles the Rust to WASM automatically.

For users: `dippin` must be on PATH (installed via `go install` or Homebrew).
