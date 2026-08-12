---
title: "Editor Setup"
description: "Set up syntax highlighting, LSP diagnostics, hover docs, and go-to-definition for .dip files in VS Code, Neovim, and more."
section_label: "Editor Integration"
subtitle: "LSP server, VS Code syntax, Neovim, and more."
---

## LSP Features

The built-in LSP server (`dippin lsp`) provides rich editing features for any LSP-compatible editor. It communicates over stdio using JSON-RPC 2.0.

| Feature | Description |
|---------|-------------|
| **Diagnostics** | Parse errors and lint warnings published on every change (DIP001-DIP010 structural, DIP101-DIP162 semantic) |
| **Hover** | Tooltip showing node kind and a per-kind field summary — model/provider for agents, command for tools, mode for human gates |
| **Go-to-definition** | Jump from a node reference in an edge to the node's declaration |
| **Autocomplete** | Node IDs in edges and field names within node blocks |
| **Document symbols** | Outline view showing all nodes and the edges section |

Diagnostics are the most valuable feature — you get real-time feedback as you type, identical to running `dippin lint` on every save.

## VS Code

### Extension Install

The VS Code extension provides syntax highlighting independently of the LSP server. For the best experience, install both.

```
# Symlink into VS Code extensions directory
ln -s "$(pwd)/editors/vscode" ~/.vscode/extensions/dippin-lang

# Or copy
cp -r editors/vscode ~/.vscode/extensions/dippin-lang

# Restart VS Code — .dip files should show "Dippin" as the language mode
```

### LSP Configuration

Add to your `settings.json` to connect the LSP server:

```
{
  "dippin.lsp.enabled": true,
  "dippin.lsp.path": "dippin"
}
```

Or use a generic LSP client extension with this configuration:

```
{
  "languageServerExample.trace.server": "verbose",
  "languageserver": {
    "dippin": {
      "command": "dippin",
      "args": ["lsp"],
      "filetypes": ["dippin"],
      "rootPatterns": [".git"]
    }
  }
}
```

### Highlighted Elements

| Element | Example | Color Category |
|---------|---------|----------------|
| Keywords | `workflow`, `agent`, `tool`, `human`, `edges` | keyword |
| Node names | `agent MyNode` | function |
| Field keys | `model:`, `prompt:`, `timeout:` | tag |
| Strings | `"quoted value"`, `'quoted value'` | string |
| Comments | `# comment` | comment |
| Arrows | `->`, `<-` | operator |
| Conditions | `when`, `and`, `or`, `not` | keyword |
| Edge keywords | `on`, `loop`, `choice:`, `weight:`, `override:`, `restart:`, `else` | keyword |
| Variables | `ctx.outcome`, `graph.goal` | variable |
| Interpolation | `${ctx.var}` | variable |
| Booleans | `true`, `false` | constant |
| Numbers | `3`, `60s`, `5m` | number |

## Neovim

### With nvim-lspconfig

Add to your Neovim config:

```
local lspconfig = require('lspconfig')
local configs = require('lspconfig.configs')

if not configs.dippin then
  configs.dippin = {
    default_config = {
      cmd = { 'dippin', 'lsp' },
      filetypes = { 'dippin' },
      root_dir = lspconfig.util.root_pattern('.git'),
      settings = {},
    },
  }
end

lspconfig.dippin.setup({})
```

### Filetype Registration

Register the `.dip` extension:

```
vim.filetype.add({
  extension = {
    dip = 'dippin',
  },
})
```

### Tree-sitter

For full syntax highlighting with tree-sitter, install the `tree-sitter-dippin` grammar. This provides more accurate highlighting than regex-based syntax files, including proper scoping of multiline blocks and condition expressions.

### Basic Syntax Highlighting

For keyword highlighting without tree-sitter, create `~/.config/nvim/syntax/dippin.vim`:

```
syn keyword dippinKeyword workflow agent human tool parallel fan_in subgraph edges defaults
syn keyword dippinCondKeyword when and or not
syn keyword dippinBoolean true false
syn match dippinField /^\s*\w\+:/
syn match dippinArrow /->/
syn match dippinArrow /<-/
syn match dippinComment /#.*/
syn match dippinVariable /\${[^}]*}/
syn region dippinString start=/"/ end=/"/

hi link dippinKeyword Keyword
hi link dippinCondKeyword Conditional
hi link dippinBoolean Boolean
hi link dippinField Tag
hi link dippinArrow Operator
hi link dippinComment Comment
hi link dippinVariable Identifier
hi link dippinString String
```

## Other Editors

Any editor with LSP support can use `dippin lsp`. The common configuration pattern:

| Setting | Value |
|---------|-------|
| Command | `dippin lsp` |
| Transport | stdio |
| File types | `.dip` |

Editors known to work with generic LSP clients:

- **Sublime Text** — via LSP package
- **Emacs** — via lsp-mode or eglot
- **Helix** — native LSP support
- **Zed** — native LSP support

## Graphviz Integration

For visual workflow diagrams, pipe `export-dot` output to Graphviz:

```
# PNG output
dippin export-dot pipeline.dip | dot -Tpng -o pipeline.png

# SVG for web
dippin export-dot pipeline.dip | dot -Tsvg -o pipeline.svg

# Left-to-right layout (better for wide pipelines)
dippin export-dot --rankdir=LR pipeline.dip | dot -Tpng -o pipeline.png

# Include prompt text in nodes
dippin export-dot --prompts pipeline.dip | dot -Tpng -o pipeline.png
```

Install Graphviz: `brew install graphviz` (macOS), `apt install graphviz` (Debian/Ubuntu), or [graphviz.org](https://graphviz.org/download/).
