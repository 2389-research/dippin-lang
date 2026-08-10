# #72 — Shared `system_prompt_file` default

**Goal:** Let a workflow declare one shared system prompt (persona/role) once in the
`defaults` block and have it apply to every agent that does not set its own —
so a large external persona doc can be authored once, not pasted into each agent.

**Status:** design approved 2026-08-10. Small, self-contained follow-up to #175
(the `prompt_prefix`/`prompt_suffix` cascade, which covered the prompt *body*).
The system prompt is a distinct field from the prompt body, so #175 did not
address it; this closes the remaining half of #72.

## Background

- Agent nodes already carry `system_prompt` (inline) and `system_prompt_file`
  (path), mutually exclusive, resolved to `AgentConfig.SystemPrompt` by
  `parser.ResolveFileDirectives` → `resolveAgentDirective`.
- The `defaults` block already carries a prompt cascade from #175:
  `prompt_prefix`/`prompt_suffix` (+ `_file` forms) that **wrap** each agent's
  prompt body, loaded once in `loadPromptCascade` and applied in
  `resolveAgentDirective`, with a per-node `none` opt-out.
- The `defaults` block does **not** carry any system-prompt field today.

## Decisions

1. **Fallback default (node wins).** `defaults.system_prompt_file` supplies a
   system prompt only to agents that declare none of their own. A node's own
   `system_prompt` or `system_prompt_file` fully overrides it — never merged.
   This mirrors how `model`/`provider` defaults already resolve, not the
   wrap/compose model of `prompt_prefix`.
2. **File form only.** The `defaults` block accepts `system_prompt_file` (a
   path) and not an inline `defaults.system_prompt`. The use case is sharing one
   large external persona doc; inline sharing can be added later if asked. An
   inline `system_prompt:` under `defaults` remains an unknown-field error.
3. **No opt-out sentinel.** A node overrides the default by setting its own
   system prompt. There is no `system_prompt: none` to run an agent with *no*
   system prompt despite a default — unlikely for a shared-persona workflow, and
   omitting it keeps the surface minimal.

## Design

### IR
Add one field to `ir.WorkflowDefaults`:
```go
SystemPromptFile string // defaults-block shared system prompt (path); #72
```

### Parser (`parser/parse_defaults.go`)
Recognize `system_prompt_file` in the defaults field handlers, assigning
`d.SystemPromptFile = val`. No inline counterpart, so no
inline-XOR-file check is added (unlike `checkDefaultsPromptExclusive` for
prefix/suffix). An inline `system_prompt` under `defaults` still falls through to
the existing `unknown defaults field` diagnostic.

### Resolve (`parser/resolve.go`)
- `loadPromptCascade` loads the defaults file once into the cascade struct
  (new field `systemPrompt string`), using the same `loadDirectiveInto`
  containment/size/symlink checks every other file directive uses. A missing or
  out-of-tree default file errors at resolve time, exactly like a node's own
  `system_prompt_file`.
- `resolveAgentDirective` resolves the node's own `system_prompt`/`_file` into
  `cfg.SystemPrompt` as it does today; then, **only if `cfg.SystemPrompt == ""`**,
  fills it from `cascade.systemPrompt`. The node's own value (inline or file)
  always wins because it is non-empty by that point.

### Formatter (`formatter/format.go`)
Emit `system_prompt_file: <path>` in the `defaults` block when
`Defaults.SystemPromptFile != ""`, alongside the existing defaults fields.

### Surfaces to sweep (same batch)
`docs` (defaults reference + nodes/prompt docs), `docs/GRAMMAR.ebnf`
(`defaults_field`), `docs/llm-reference.md`, `site/static/skill.md`,
`site/static/llms-full.txt`, `site/content` defaults/language docs, regenerated
`generated-spec.md`. Tree-sitter needs no change (generic `field_name`).

## Testing

- **Parser:** `defaults` accepts `system_prompt_file`; an inline
  `system_prompt` under `defaults` is rejected as unknown.
- **Resolve:** (a) an agent with no system prompt inherits the default file's
  content; (b) an agent with its own inline `system_prompt` keeps it, default
  unused; (c) an agent with its own `system_prompt_file` keeps it, default
  unused; (d) a missing default file errors.
- **Formatter:** a workflow with `defaults.system_prompt_file` round-trips.
- **Example:** one `examples/*.dip` demonstrating the shared persona (optional,
  if it doesn't bloat the example set).

## Out of scope (YAGNI)

Inline `defaults.system_prompt`; a `system_prompt: none` opt-out sentinel; a
"defaults system prompt is set but every agent overrides it" dead-default lint.
