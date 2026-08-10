# Response: Structured Output Support in Dippin-Lang IR

**From:** Dippin-lang team
**To:** Downstream consumers
**Date:** 2026-03-31
**Re:** Feature Request — Structured Output Support (2026-03-31)
**Status:** Implemented, ready for integration

---

## Summary

All requested changes are implemented and passing CI. Downstream consumers can begin integration immediately.

**Version:** v0.16.0 (not yet tagged — pending your confirmation)
**Commits:** 13 commits on `main` since v0.15.0

---

## What Was Implemented

### 1. `response_format` and `response_schema` on `AgentConfig`

```go
type AgentConfig struct {
    // ... existing fields ...
    ResponseFormat string            // "json_object" or "json_schema"
    ResponseSchema string            // JSON schema (when ResponseFormat is "json_schema")
    Params         map[string]string  // Generic key-value pairs passed through to runtime
}
```

Both fields are typed strings, exactly as requested. `ResponseSchema` stores the raw JSON content verbatim — no escaping, no transformation. It arrives in the IR ready to pass directly to the LLM API.

### 2. `.dip` syntax — matches your requested format exactly

```dip
agent GenerateQuestions
  label: "Generate Interview Questions"
  model: claude-opus-4-6
  provider: anthropic
  response_format: json_object
  prompt:
    Output a JSON object with interview questions...

agent StrictQuestions
  label: "Generate Typed Questions"
  model: gpt-5.4
  provider: openai
  response_format: json_schema
  response_schema:
    {
      "type": "object",
      "properties": {
        "questions": {
          "type": "array",
          "items": {
            "type": "object",
            "properties": {
              "text": {"type": "string"},
              "context": {"type": "string"},
              "options": {"type": "array", "items": {"type": "string"}}
            },
            "required": ["text"]
          }
        }
      },
      "required": ["questions"]
    }
  prompt:
    Generate interview questions for the developer...
```

`response_schema` uses the same multiline block pattern as `prompt:` and `command:` — indent after the colon, write raw JSON. Internal indentation is preserved through parse/format round-trips. The formatter emits it as a multiline block, identical to how it was authored.

### 3. Generic `Params` map on `AgentConfig`

```dip
agent MyNode
  label: "My Node"
  model: claude-sonnet-4-6
  params:
    backend: claude-code
    permission_mode: auto
    mcp_server: github
  prompt:
    Do something...
```

Same syntax and semantics as the existing `SubgraphConfig.Params`. Keys are sorted alphabetically in formatted output. Values are unquoted when possible (same rules as all other Dippin values).

### 4. Validation rules (DIP130-DIP133)

| Code | Severity | What it catches |
|------|----------|-----------------|
| DIP130 | Warning | `response_format` is not `json_object` or `json_schema` |
| DIP131 | Warning | `response_schema` without `response_format: json_schema` (schema will be ignored by API) |
| DIP131 | Hint | `response_format: json_schema` without `response_schema` (valid but advisory) |
| DIP132 | Warning | `response_schema` content is not valid JSON |
| DIP133 | Hint | `params` key shadows a first-class field (e.g., putting `model` in params instead of using the typed field) |

---

## Runtime Integration Notes

### What the adapter needs to do

Your `pipeline/dippin_adapter.go` can now read directly from the IR:

```go
cfg := node.Config.(ir.AgentConfig)

// Typed fields — check these first
if cfg.ResponseFormat != "" {
    attrs["response_format"] = cfg.ResponseFormat
}
if cfg.ResponseSchema != "" {
    attrs["response_schema"] = cfg.ResponseSchema
}

// Generic params — fall back to these for keys without typed fields
for k, v := range cfg.Params {
    if _, exists := attrs[k]; !exists {
        attrs[k] = v
    }
}
```

### Field types and nil safety

- `ResponseFormat`: `string`, zero value is `""` (not set)
- `ResponseSchema`: `string`, zero value is `""` (not set). Content is raw JSON, ready for API consumption.
- `Params`: `map[string]string`, initialized to empty map (never nil). Safe to range over without nil check. `len(cfg.Params) == 0` when no params block is declared.

### Canonical field ordering in formatted output

```text
label → model → provider → reasoning_effort → fidelity →
response_format → response_schema →
goal_gate → auto_status → max_turns → cmd_timeout →
cache_tools → compaction → compaction_threshold →
retry fields → reads → writes →
params →
prompt (always last)
```

### Schema content fidelity

The JSON schema content round-trips exactly. We verified this with:
- `TestParseResponseFormat`: parses the fixture and validates `json.Valid()` on the schema
- `TestRoundtripResponseFormatFields`: parse → format → parse, verifies field values survive and `json.Valid()` holds
- `TestRoundtripTestdata/response_format.dip`: generic round-trip (format idempotency)

### Quoted param values

Param values are unquoted by the parser (same fix we applied for subgraph params). `backend: "claude-code"` stores as `claude-code`. `backend: claude-code` also stores as `claude-code`. The adapter receives clean, unquoted values.

---

## Secondary Request: Generic `Params` Map

Implemented as requested. The `Params` map is available on `AgentConfig` for any future runtime features. DIP133 provides the soft guardrail — if a user puts a first-class field like `model` or `response_format` into params instead of using the typed field, they'll see a hint (not an error, not blocking).

Current first-class field allowlist for DIP133: `model`, `provider`, `prompt`, `system_prompt`, `max_turns`, `response_format`, `response_schema`, `reasoning_effort`, `fidelity`, `auto_status`, `goal_gate`, `cache_tools`, `compaction`, `compaction_threshold`.

---

## Additional Fixes (Found During Implementation)

These pre-existing issues were discovered and fixed during code review:

1. **`cmd_timeout` round-trip**: Agent `cmd_timeout` was populated by the DOT migrator but never emitted by the formatter or parsed from `.dip` files. Now fully supported — parse, format, round-trip.
2. **Duplicate params keys**: Previously silently last-write-wins. Now emits a parse diagnostic.
3. **Unknown defaults fields**: Previously silently discarded (e.g., `response_format` in a `defaults` block). Now emits a parse diagnostic.
4. **Formatter idempotency** (from v0.15.0): Subgraph param values with quotes were double-quoted on each format pass. Fixed.

---

## Testing Your Integration

Once you update the adapter:

```bash
# Verify the IR fields exist
go doc github.com/2389-research/dippin-lang/ir AgentConfig

# Parse a .dip file with response_format and inspect the IR
# (your existing test harness should work)

# The example file already uses it:
dippin lint examples/api_design.dip
# → clean (no DIP130-133 warnings)

# Verify structured output validation:
dippin explain DIP130
dippin explain DIP131
dippin explain DIP132
dippin explain DIP133
```

---

## Documentation

All documentation updated:
- EBNF grammar (`docs/GRAMMAR.ebnf`)
- Node field reference (`docs/nodes.md`, `site/language.html`)
- Validation reference (`docs/validation.md`, `site/validation.html`)
- CLI reference (`docs/cli.md`, `site/cli.html`)
- Integration guide (`docs/integration.md`)
- LLM reference (`docs/llm-reference.md`)
- Architecture (`site/architecture.html`)
- CHANGELOG.md (v0.16.0 entry)
- README.md, CLAUDE.md

---

## Next Steps

1. Tag v0.16.0 and push (waiting for your go-ahead)
2. Runtime adapter update (your side — should be minimal given the plumbing is already merged on `feat/interview-mode`)
3. Integration test: `.dip` → IR → adapter → node attrs → codergen → agent session → LLM request
