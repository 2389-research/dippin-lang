# Editor Setup

Dippin provides a Language Server Protocol (LSP) server and a VS Code extension for editor integration.

---

## LSP Server

The built-in LSP server provides rich editing features for any LSP-compatible editor.

```sh
dippin lsp
```

This starts the server on stdio (stdin/stdout), speaking JSON-RPC 2.0 per the LSP specification.

### Features

| Feature | Description |
|---------|-------------|
| **Diagnostics** | Parse errors and lint warnings published on every change (DIP001–DIP010 structural, DIP101–DIP152 semantic) |
| **Hover** | Tooltip showing node kind, model, provider, prompt preview, and field summary |
| **Go-to-definition** | Jump from a node reference in an edge to the node's declaration |
| **Autocomplete** | Node IDs in edges, field names within node blocks, keywords |
| **Document symbols** | Outline view showing all nodes and the edges section |

Diagnostics are the most valuable feature — you get real-time feedback as you type, identical to running `dippin lint` on every save.

---

## VS Code

### Extension Install

The VS Code extension provides syntax highlighting independently of the LSP server. For the best experience, install both.

**Syntax highlighting** (no build step needed):

```sh
# Symlink into VS Code extensions directory
ln -s "$(pwd)/editors/vscode" ~/.vscode/extensions/dippin-lang

# Restart VS Code
```

Or copy:

```sh
cp -r editors/vscode ~/.vscode/extensions/dippin-lang
```

Open any `.dip` file — the status bar should show "Dippin" as the language mode.

### LSP Configuration

To connect the LSP server in VS Code, add to your `settings.json`:

```json
{
  "dippin.lsp.enabled": true,
  "dippin.lsp.path": "dippin"
}
```

Or use a generic LSP client extension (like vscode-languageclient) with this configuration:

```json
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

| Element | Example | Color category |
|---------|---------|---------------|
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

---

## Neovim

### With nvim-lspconfig

Add to your Neovim config:

```lua
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

Register the filetype:

```lua
vim.filetype.add({
  extension = {
    dip = 'dippin',
  },
})
```

### Tree-sitter Grammar

A tree-sitter grammar is available in [`editors/tree-sitter-dippin/`](../editors/tree-sitter-dippin/). It includes an external scanner for Dippin's indentation-sensitive syntax and highlight queries for full semantic highlighting in Neovim, Helix, and Zed.

To use with `nvim-treesitter`, register the parser in your config. See the grammar directory for installation instructions.

### Basic Syntax Highlighting (Fallback)

For basic keyword highlighting without tree-sitter, create `~/.config/nvim/syntax/dippin.vim`:

```vim
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

---

## Other Editors

Any editor with LSP support can use `dippin lsp`. The server communicates over stdio, so the configuration is always:

- **Command:** `dippin lsp`
- **Transport:** stdio
- **File types:** `.dip`

Editors known to work with generic LSP clients:
- Sublime Text (via LSP package)
- Emacs (via lsp-mode or eglot)
- Helix (native LSP support)
- Zed (native LSP support)

---

## Graphviz Integration

For visual workflow diagrams, pipe `export-dot` output to Graphviz:

```sh
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
