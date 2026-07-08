---
title: "Editor Setup: LSP, VS Code, and Tree-sitter"
date: "2026-03-27"
description: "Set up real-time Dippin diagnostics, hover docs, and syntax highlighting in VS Code, Neovim, Helix, or any editor with LSP support."
tagStyle: "guide"
tagLabel: "GUIDE"
category: "Tooling"
readTime: "10 min read"
related:
  - url: "getting-started.html"
    title: "Getting Started"
    summary: "Write your first workflow from scratch with the toolchain you just set up."
  - url: "ci-integration.html"
    title: "CI Integration"
    summary: "Enforce the same checks in CI that your editor now shows you locally."
---

Dippin ships with a built-in
[Language Server Protocol](https://microsoft.github.io/language-server-protocol/)
(LSP) server and a VS Code extension. Together they give you real-time diagnostics,
hover documentation, go-to-definition, autocomplete, and syntax highlighting -- the
same checks you'd get from `dippin lint`, but live as you type.

## The LSP Server

The LSP server is built into the `dippin` binary. No separate
install needed:

<pre>
<span class="hl-dim">$</span> <span class="hl-shcmd">dippin lsp</span>
</pre>

Starts the server on stdio (stdin/stdout), speaking JSON-RPC 2.0 per the
[LSP spec](https://microsoft.github.io/language-server-protocol/specifications/lsp/3.17/specification/).
Any editor with LSP support can connect.<sup>1</sup>

### What the LSP Provides

<div class="feature-grid">
  <div class="feature-item">
    <h4>Diagnostics</h4>
    <p>Parse errors and every lint diagnostic on every keystroke — DIP001-DIP010 structural, DIP101-DIP152 semantic (62 codes). Same checks as <code>dippin lint</code>.</p>
  </div>
  <div class="feature-item">
    <h4>Hover</h4>
    <p>Hover over a node name to see its kind, model, provider, prompt preview, and field summary.</p>
  </div>
  <div class="feature-item">
    <h4>Go-to-definition</h4>
    <p>Click a node reference in an edge declaration to jump to the node's definition.</p>
  </div>
  <div class="feature-item">
    <h4>Autocomplete</h4>
    <p>Node IDs in edge declarations, field names within node blocks, and Dippin keywords.</p>
  </div>
  <div class="feature-item">
    <h4>Document Symbols</h4>
    <p>Outline view showing all nodes and the edges section for quick navigation.</p>
  </div>
  <div class="feature-item">
    <h4>Real-time Feedback</h4>
    <p>Diagnostics update as you type -- same as running <code>dippin lint</code> on every save.</p>
  </div>
</div>

<div class="callout">
  <h4>Universal configuration</h4>
  <p>
    For any editor, the LSP config is always the same: command is
    <code>dippin lsp</code>, transport is stdio, file type is <code>.dip</code>.
    See the <a href="../editors.html">editors reference</a> for more detail.
  </p>
</div>

## VS Code

<div class="editor-section vscode">
  <span class="editor-badge vscode">VS Code</span>

  <h3>Install the Extension</h3>

  The VS Code extension lives in the repository at
  [`editors/vscode/`](https://github.com/2389-research/dippin-lang/tree/main/editors/vscode).
  It provides syntax highlighting independently of the LSP server. For the best
  experience, install both.

  Symlink into your VS Code extensions directory:

  <pre>
<span class="hl-dim">$</span> <span class="hl-shcmd">ln -s "$(pwd)/editors/vscode" ~/.vscode/extensions/dippin-lang</span>
  </pre>

  Or copy it:

  <pre>
<span class="hl-dim">$</span> <span class="hl-shcmd">cp -r editors/vscode ~/.vscode/extensions/dippin-lang</span>
  </pre>

  Restart VS Code. Open any `.dip` file -- the status bar should show
  "Dippin" as the language mode.

  <h3>What Gets Highlighted</h3>

  The TextMate grammar covers all Dippin syntax elements:

  <table>
    <thead>
      <tr><th>Element</th><th>Example</th><th>Category</th></tr>
    </thead>
    <tbody>
      <tr><td>Keywords</td><td><code>workflow</code>, <code>agent</code>, <code>tool</code>, <code>human</code>, <code>edges</code></td><td>keyword</td></tr>
      <tr><td>Node names</td><td><code>agent MyNode</code></td><td>function</td></tr>
      <tr><td>Field keys</td><td><code>model:</code>, <code>prompt:</code>, <code>timeout:</code></td><td>tag</td></tr>
      <tr><td>Strings</td><td><code>"quoted value"</code></td><td>string</td></tr>
      <tr><td>Comments</td><td><code># comment</code></td><td>comment</td></tr>
      <tr><td>Arrows</td><td><code>-></code>, <code><-</code></td><td>operator</td></tr>
      <tr><td>Conditions</td><td><code>when</code>, <code>and</code>, <code>or</code>, <code>not</code></td><td>keyword</td></tr>
      <tr><td>Variables</td><td><code>ctx.outcome</code>, <code>${ctx.var}</code></td><td>variable</td></tr>
      <tr><td>Booleans</td><td><code>true</code>, <code>false</code></td><td>constant</td></tr>
      <tr><td>Numbers</td><td><code>3</code>, <code>60s</code>, <code>5m</code></td><td>number</td></tr>
    </tbody>
  </table>

  <h3>Connect the LSP</h3>

  The extension includes a built-in LSP client. Add to your VS Code
  `settings.json`:

  <pre>
{
  "dippin.lsp.enabled": true,
  "dippin.lsp.path": "dippin"
}
  </pre>

  `dippin.lsp.path` defaults to `"dippin"`,
  which assumes the binary is on your `$PATH`. If you installed
  somewhere else, provide the full path.

  Once connected, you'll see squiggly underlines for diagnostics, hover
  tooltips on node names, and autocomplete suggestions when you type edge
  declarations.

  <h3>Extension Details</h3>

  The extension is minimal by design:

  <table>
    <thead>
      <tr><th>Field</th><th>Value</th></tr>
    </thead>
    <tbody>
      <tr><td>Name</td><td><code>dippin-lang</code></td></tr>
      <tr><td>Display name</td><td>Dippin</td></tr>
      <tr><td>Version</td><td>0.2.0</td></tr>
      <tr><td>VS Code engine</td><td>^1.75.0</td></tr>
      <tr><td>Activation</td><td><code>onLanguage:dippin</code></td></tr>
      <tr><td>LSP client</td><td><code>vscode-languageclient ^9.0.1</code></td></tr>
    </tbody>
  </table>

</div>

## Neovim

<div class="editor-section neovim">
  <span class="editor-badge neovim">Neovim</span>

  <h3>With nvim-lspconfig</h3>

  Add to your Neovim configuration:

  <pre>
<span class="hl-shkw">local</span> lspconfig = <span class="hl-shcmd">require</span>('lspconfig')
<span class="hl-shkw">local</span> configs = <span class="hl-shcmd">require</span>('lspconfig.configs')

<span class="hl-shkw">if not</span> configs.dippin <span class="hl-shkw">then</span>
  configs.dippin = {
    default_config = {
      cmd = { 'dippin', 'lsp' },
      filetypes = { 'dippin' },
      root_dir = lspconfig.util.root_pattern('.git'),
      settings = {},
    },
  }
<span class="hl-shkw">end</span>

lspconfig.dippin.setup({})
  </pre>

  <h3>Register the Filetype</h3>

  Tell Neovim that `.dip` files are Dippin files:

  <pre>
vim.filetype.add({
  extension = {
    dip = 'dippin',
  },
})
  </pre>

  <h3>Tree-sitter Grammar</h3>

  A full [tree-sitter](https://tree-sitter.github.io/tree-sitter/) grammar
  is available in
  [`editors/tree-sitter-dippin/`](https://github.com/2389-research/dippin-lang/tree/main/editors/tree-sitter-dippin).
  It includes an external scanner for Dippin's indentation-sensitive syntax and
  highlight queries for semantic highlighting. To use with `nvim-treesitter`,
  register the parser in your config.

  The tree-sitter grammar gives more precise highlighting than the fallback
  syntax file, correctly handling multi-line prompt blocks, nested conditions,
  and indentation-based scope boundaries.<sup>2</sup>

  <h3>Basic Syntax Highlighting (Fallback)</h3>

  For keyword highlighting without tree-sitter, create
  `~/.config/nvim/syntax/dippin.vim`:

  <pre>
syn keyword dippinKeyword workflow agent human tool parallel fan_in subgraph edges defaults
syn keyword dippinCondKeyword when and or not
syn keyword dippinBoolean true false
syn match dippinField /^\s*\w\+:/
syn match dippinArrow /->/
syn match dippinArrow /&lt;-/
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
  </pre>

</div>

## Other Editors

<div class="editor-section other">
  <span class="editor-badge other">Any LSP Editor</span>

  Any editor with LSP support can use `dippin lsp`. The config
  is always the same:

  <table>
    <thead>
      <tr><th>Setting</th><th>Value</th></tr>
    </thead>
    <tbody>
      <tr><td>Command</td><td><code>dippin lsp</code></td></tr>
      <tr><td>Transport</td><td>stdio</td></tr>
      <tr><td>File types</td><td><code>.dip</code></td></tr>
    </tbody>
  </table>

  <h3>Editors Known to Work</h3>

  <table>
    <thead>
      <tr><th>Editor</th><th>LSP client</th><th>Tree-sitter support</th></tr>
    </thead>
    <tbody>
      <tr><td><a href="https://www.sublimetext.com/">Sublime Text</a></td><td>LSP package</td><td>No</td></tr>
      <tr><td><a href="https://www.gnu.org/software/emacs/">Emacs</a></td><td>lsp-mode or eglot</td><td>No</td></tr>
      <tr><td><a href="https://helix-editor.com/">Helix</a></td><td>Native LSP</td><td>Yes (register grammar)</td></tr>
      <tr><td><a href="https://zed.dev/">Zed</a></td><td>Native LSP</td><td>Yes (register grammar)</td></tr>
    </tbody>
  </table>

  For Helix and Zed, register the tree-sitter grammar from
  `editors/tree-sitter-dippin/` for full semantic highlighting
  alongside LSP diagnostics.

</div>

## dippin watch as an Alternative

If your editor doesn't support LSP, or you just want something lightweight,
`dippin watch` gives live feedback in a terminal pane:

<pre>
<span class="hl-dim">$</span> <span class="hl-shcmd">dippin watch pipeline.dip</span>
<span class="hl-dim">Watching pipeline.dip for changes...</span>

<span class="hl-dim">[14:32:05]</span> <span class="hl-pass">PASS</span>  pipeline.dip  <span class="hl-dim">(0 errors, 0 warnings)</span>
<span class="hl-dim">[14:32:18]</span> <span class="hl-warn">WARN</span>  pipeline.dip
  <span class="hl-warn">DIP108</span>: node "Analyze": unknown model "gpt-4-turbo" for provider "openai"
<span class="hl-dim">[14:32:24]</span> <span class="hl-pass">PASS</span>  pipeline.dip  <span class="hl-dim">(0 errors, 0 warnings)</span>
</pre>

Run it in a split terminal next to your editor. On every save, you get
instant lint results. The watcher debounces at 200ms to avoid duplicate
runs when editors write multiple times on save.

Watch a whole directory:

<pre>
<span class="hl-dim">$</span> <span class="hl-shcmd">dippin watch pipelines/</span>
<span class="hl-dim">Watching 8 .dip files in pipelines/...</span>
</pre>

## Graphviz Integration

For visual workflow diagrams alongside your editor, pipe `export-dot`
output to [Graphviz](https://graphviz.org/):

<pre>
<span class="hl-dim"># PNG output</span>
<span class="hl-dim">$</span> <span class="hl-shcmd">dippin export-dot pipeline.dip | dot -Tpng -o pipeline.png</span>

<span class="hl-dim"># SVG for web</span>
<span class="hl-dim">$</span> <span class="hl-shcmd">dippin export-dot pipeline.dip | dot -Tsvg -o pipeline.svg</span>

<span class="hl-dim"># Left-to-right layout</span>
<span class="hl-dim">$</span> <span class="hl-shcmd">dippin export-dot --rankdir=LR pipeline.dip | dot -Tpng -o pipeline.png</span>

<span class="hl-dim"># Include prompt text in nodes</span>
<span class="hl-dim">$</span> <span class="hl-shcmd">dippin export-dot --prompts pipeline.dip | dot -Tpng -o pipeline.png</span>
</pre>

Install Graphviz: `brew install graphviz` on macOS,
`apt install graphviz` on Debian/Ubuntu, or visit
[graphviz.org](https://graphviz.org/download/).

## Choosing Your Setup

A quick decision guide:

| Situation | Recommendation |
|-----------|---------------|
| VS Code, want the full experience | Install extension + enable LSP in settings |
| Neovim, want the full experience | nvim-lspconfig + tree-sitter grammar |
| Neovim, quick setup | nvim-lspconfig + fallback syntax file |
| Helix or Zed | Native LSP config + tree-sitter grammar |
| Editor without LSP | `dippin watch` in a terminal pane |
| Highlighting only, no diagnostics | VS Code extension or Vim syntax file alone |

## What's Next?

Your editor now has live diagnostics and syntax highlighting.

<div class="footnotes">
  <h3>Notes</h3>
  <ol>
    <li id="fn1">The LSP server reuses the same parse and lint code paths as the CLI. When you see a diagnostic in your editor, it's the exact same check that <code>dippin lint</code> and <a href="ci-integration.html">CI</a> run. No separate analysis engine, no divergence risk. The implementation is in <a href="https://github.com/2389-research/dippin-lang/tree/main/lsp"><code>lsp/</code></a>.</li>
    <li id="fn2">Dippin's indentation-based syntax (similar to Python or YAML) means a regex-based TextMate grammar can't accurately determine scope boundaries. Tree-sitter's incremental parsing handles this correctly, which is why the tree-sitter grammar gives better results for multi-line <code>prompt:</code> and <code>command:</code> blocks.</li>
  </ol>
</div>
