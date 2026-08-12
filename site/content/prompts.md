---
title: "Prompts & Context"
description: "How agent prompts are authored in .dip files: multiline text blocks, prompt fields, the defaults prompt cascade, context truncation, and interpolation."
section_label: "Language"
subtitle: "Authoring prompts, the defaults cascade, and context control."
---

## Multiline Blocks

Fields like `prompt` and `command` support multiline content. Write the key followed by `:`, then indent the content on subsequent lines:

```
  agent MyAgent
    prompt:
      You are a code reviewer.

      ## Rules
      - Check for bugs
      - Check for security issues

      ## Context
      ${ctx.last_response}
```

The first content line's indentation sets the baseline. All content is de-indented by that amount. Empty lines are preserved. The block ends when indentation returns to or above the field's level. No quoting or escaping needed.

```
  tool RunTests
    timeout: 60s
    command:
      #!/bin/sh
      set -eu
      if pytest --tb=short 2>&1; then
        printf 'pass'
      else
        printf 'fail'
        exit 1
      fi
```

## Agent Prompt Fields

Agent nodes carry the prompt content sent to the model. A prompt can be written inline as a multiline block or loaded from an external file, and a system prompt (persona / role instructions) can be supplied the same way.

| Field | Type | Description |
|-------|------|-------------|
| `prompt` | Block | Multiline prompt text sent to the model |
| `prompt_file` | String | Path (relative to the `.dip` dir) to an external file whose contents become the prompt. Mutually exclusive with `prompt:` — setting both is a parse error. |
| `system_prompt` | Block | System-level instructions prepended before the prompt |
| `system_prompt_file` | String | Path (relative to the `.dip` dir) to an external file whose contents become the system prompt. Mutually exclusive with `system_prompt:` — setting both is a parse error. |
| `prompt_include` | String | Path to a fragment file appended after the body, before the defaults cascade suffix (#175). |

## The Defaults Prompt Cascade

The `prompt_prefix`/`prompt_suffix` (inline) and `prompt_prefix_file`/`prompt_suffix_file` (fragment file) defaults **cascade a shared prompt fragment to every agent** (#175). This lets you author a shared fragment once — a status contract, a house style, a protocol — and have it wrap every agent's own prompt.

```
  defaults
    model: claude-opus-4-6
    provider: anthropic
    prompt_suffix_file: protocols/status-contract.md
```

| Field | Type | Description |
|-------|------|-------------|
| `prompt_prefix` / `prompt_suffix` | String | Inline prompt fragment cascaded to every agent as a prefix/suffix (#175) |
| `prompt_prefix_file` / `prompt_suffix_file` | String | Prompt fragment loaded from a file and cascaded to every agent (mutually exclusive with the inline form) |
| `system_prompt_file` | String | Shared system prompt (persona) loaded from a file; a fallback default for agents that set none of their own — a node's own `system_prompt`/`system_prompt_file` overrides it (#72). File form only |

### Composition order

The effective prompt is composed as `prefix → body → prompt_include → suffix` at resolve time, with the suffix always last. Parts are joined with a fixed blank line (`\n\n`), and only non-empty parts are joined — so the cascade suffix is always the final content, satisfying "the very last line MUST be exactly …" contracts.

A body-less passthrough agent (no own prompt or include — e.g. a declared `start:`/`exit:` node) is skipped so the cascade never synthesizes a prompt on it (#248, #249).

### Opting out

An agent opts out with `prompt_suffix: none` / `prompt_prefix: none`.

| Field | Type | Description |
|-------|------|-------------|
| `prompt_prefix` / `prompt_suffix` | `none` | Set to `none` to opt this agent out of the corresponding `defaults` prompt cascade (#175). |

An agent that sets `prompt_prefix: none` / `prompt_suffix: none` while the `defaults` block declares no cascade of that kind raises DIP154 — the opt-out is a no-op.

### The system-prompt fallback

The `system_prompt_file` default (#72) is different in kind from the prefix/suffix cascade: it is a **fallback** shared system prompt (persona) — used only by agents that declare no `system_prompt`/`system_prompt_file` of their own, which fully override it. It is file form only.

## Context Truncation

The `last_response_truncate` field caps how much of the prior node's response is carried into an agent's context, in characters.

| Field | Type | Description |
|-------|------|-------------|
| `last_response_truncate` | Integer | Caps how much of the prior node's response is carried into this agent's context, in characters. `0`/unset = no truncation (a negative value raises DIP148). |

It is also available as a `parallel` branch override, where `0`/unset inherits the target agent's cap. The field is carried and linted by dippin but the truncation is performed by a paired runtime. A negative value raises **DIP148**.

## Interpolation

Prompts can reference workflow context and declared inputs with `${…}` placeholders:

- `${ctx.*}` — values from the runtime context channel, e.g. `${ctx.last_response}` (the prior node's response) or `${ctx.outcome}`.
- `${inputs.*}` — a declared workflow input, e.g. `${inputs.idea}`. The `inputs` namespace is closed — see [Inputs](inputs.md).

```
  agent Analyze
    prompt:
      Review the request below.

      ## Task
      ${inputs.idea}

      ## Prior response
      ${ctx.last_response}
```
