# Structured Output Support — Design Spec

**From:** Downstream consumer feature request (2026-03-31)
**Priority:** High — blocking production use of interview mode

## Summary

Add `response_format`, `response_schema`, and a generic `Params` map to `AgentConfig` in the dippin-lang IR. This lets `.dip` files express structured JSON output requirements that the runtime threads through to LLM APIs (Anthropic, OpenAI, Gemini). The typed fields give immediate validation and discoverability; the generic params map prevents this same blocking pattern for future runtime features.

## Approach: Typed Fields with Params Fallthrough

- `response_format` and `response_schema` are first-class `AgentConfig` fields with dedicated lint rules
- `Params map[string]string` is a generic pass-through for future runtime features
- Lint hint (DIP133) warns when a params key shadows a first-class field
- The runtime adapter reads typed fields first, falls back to params — dippin-lang stores both independently, no merge/override logic

---

## IR Changes

**File:** `ir/ir.go`

Add three fields to `AgentConfig`:

```go
type AgentConfig struct {
    // ... existing fields ...
    ResponseFormat string            // "json_object" or "json_schema"
    ResponseSchema string            // JSON schema string (when ResponseFormat is "json_schema")
    Params         map[string]string // Generic key-value pairs passed through to runtime
}
```

No changes to `ToolConfig`, `HumanConfig`, or `SubgraphConfig`.

---

## Parser Changes

**File:** `parser/parse_nodes.go`

**`response_format`** — simple string field in `applyAgentModelField()` (LLM-related config):
```go
case "response_format":
    cfg.ResponseFormat = val
```

**`response_schema`** — multiline block in `applyAgentPromptField()` (multiline content field). Uses existing `readFieldValue()` which preserves content verbatim — base indentation stripped, internal indentation preserved. JSON indentation survives round-trip.
```go
case "response_schema":
    cfg.ResponseSchema = val
```

**`params`** — reuses existing `parseParamsBlock()` (same as `SubgraphConfig`). Added to `applyAgentComplexField()`:
```go
case "params":
    cfg.Params = parseParamsBlock(val)
```

---

## Formatter Changes

**File:** `formatter/format.go`

New helper `writeAgentResponseFields()` slots between model and behavior fields:

```
1. Common fields (label, class)
2. Model fields (model, provider, reasoning_effort, fidelity)
3. Response fields (response_format, response_schema)  <- NEW
4. Behavior fields (goal_gate, auto_status, max_turns)
5. Compaction fields (cache_tools, compaction, compaction_threshold)
6. Retry fields
7. IO fields (reads, writes)
8. Params  <- NEW (key-sorted, reuses writeSubgraphParams)
9. Prompt (always last, multiline block)
```

- `response_format` emits as a simple value: `wr.line("response_format: %s", quoteValue(cfg.ResponseFormat))`
- `response_schema` emits as a multiline block: `wr.multilineBlock("response_schema", cfg.ResponseSchema)`
- `params` emits via existing `writeSubgraphParams()` function (key-sorted)

---

## Lint Rules

**Four new DIP codes (DIP130-DIP133):**

| Code | Severity | Rule |
|------|----------|------|
| DIP130 | Warning | `response_format` must be `json_object` or `json_schema` |
| DIP131 | Warning/Hint | `response_schema` without `response_format: json_schema` (warning); `response_format: json_schema` without `response_schema` (hint) |
| DIP132 | Warning | `response_schema` is not valid JSON |
| DIP133 | Hint | Agent `params` key shadows a first-class field |

**Additional validation:** `response_format` and `response_schema` on non-agent nodes (tool/human) — handled by DIP130 (extend it to check node kind before checking value).

### DIP130: Invalid response_format value

Fires when `response_format` is set but isn't `json_object` or `json_schema`.

### DIP131: response_schema / response_format mismatch

Two sub-cases with different severities:
- `response_schema` set without `response_format: json_schema` → **warning** (schema will be ignored by the API)
- `response_format: json_schema` set without `response_schema` → **hint** (valid use case where model infers schema from prompt, just advisory)

### DIP132: Invalid JSON in response_schema

Fires when `response_schema` content doesn't pass `json.Valid()`. Validates at lint time, not parse time.

### DIP133: Params key shadows first-class field

Fires as a **hint** when any agent `params` key matches a first-class `AgentConfig` field name. Allowlist includes: `model`, `provider`, `prompt`, `system_prompt`, `max_turns`, `response_format`, `response_schema`, `reasoning_effort`, `fidelity`, `auto_status`, `goal_gate`, `cache_tools`, `compaction`, `compaction_threshold`.

Message: `params key "X" shadows a first-class field — use the typed field instead for validation`

### Non-agent node validation

Handled by DIP130 — extend the check to first verify the node is an agent. If `response_format` or `response_schema` appears on a tool or human node, DIP130 fires with a message like: `node "X": response_format is only valid on agent nodes`.

---

## .dip Syntax Examples

### json_object mode (force valid JSON output):
```dip
agent GenerateQuestions
  label: "Generate Interview Questions"
  model: claude-opus-4-6
  provider: anthropic
  response_format: json_object
  prompt:
    Output a JSON object with interview questions...
```

### json_schema mode (force JSON matching a schema):
```dip
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

### Generic params (future runtime features):
```dip
agent MyNode
  label: "My Node"
  model: claude-sonnet-4-6
  params:
    backend: claude-code
    permission_mode: auto
  prompt:
    Do something...
```

---

## Testing

### Parser tests
- New fixture `parser/testdata/response_format.dip` with all three patterns (json_object, json_schema + schema, params)
- Assert parsed IR has correct `ResponseFormat`, `ResponseSchema`, and `Params` values

### Round-trip test
- Parse → format → parse the fixture
- Specifically verify JSON schema internal indentation survives the round-trip

### Formatter test
- Verify field ordering: response_format after model fields, before behavior fields
- Verify response_schema emits as multiline block
- Verify params emits sorted, after IO fields, before prompt

### Lint tests
- DIP130: `response_format: xml` → warning
- DIP131 (warning): `response_schema` without `response_format: json_schema`
- DIP131 (hint): `response_format: json_schema` without `response_schema`
- DIP132: invalid JSON in `response_schema` → warning
- DIP132: valid JSON → no diagnostic
- DIP133: params key `model` → hint
- Non-agent: `response_format` on tool node → warning

### Integration test
- Add an example `.dip` file using `response_format` (or update an existing one) and ensure `TestLintExamples` passes
