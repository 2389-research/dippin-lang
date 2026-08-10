# #182 — Preserve escaped quotes & reject unmatched quotes in edge conditions

**Status:** design approved 2026-07-13
**Issue:** [#182](https://github.com/2389-research/dippin-lang/issues/182) (P1 bug). Surfaced through the real CLI path by tracker#444.

## Problem (two failures, both reproduced on current `main`)

1. **Escaped quotes are corrupted.** `A -> B when ctx.tool_stdout = "say \"alpha||beta\""` is lexed to the literal value `say "alpha||beta"`, then `Condition.Raw` is reconstructed as `ctx.tool_stdout = "say "alpha||beta""` — interior quotes unescaped. Condition parsing then rejects it: `error[DIP010]: invalid condition … unexpected token "alpha||beta\"\""`.
2. **Unmatched quotes are silently accepted.** `A -> B when ctx.tool_stdout = "alpha||beta` reaches end-of-line with no closing quote, but the lexer emits an ordinary literal and reconstruction fabricates a closing quote. `dippin validate` reports "validation passed" on malformed source.

## Root cause (three sites)

- **Reconstruct (write):** `parser/parse_edges.go:formatConditionToken` wraps a `TokenLiteral` as `"` + value + `"` with **no escaping** of interior `"`/`\`.
- **Re-tokenize (read):** `simulate/condition.go:tryTokenizeQuotedCond` scans to the first `quote` byte with **no escape handling**, so it can't read an escaped literal back.
- **Lexer:** `parser/lexer.go:readQuotedContent` loops to end-of-line and returns silently when the closing `"` is missing — no error signal. (`TokenError` is declared but never produced; there is no lex-error flow today.)

## Design

### 1. Shared double-quote escape convention (round-trip)

The lexer already **unescapes** on the way in (`appendQuotedChar` turns `\"`→`"`, `\\`→`\`), so `TokenLiteral.Value` holds the literal content. The reconstruct and re-tokenize sides must use the inverse.

**`parser/parse_edges.go:formatConditionToken`** — escape a literal before wrapping:
```go
func formatConditionToken(t Token) string {
	if t.Type == TokenLiteral {
		return `"` + escapeConditionLiteral(t.Value) + `"`
	}
	return t.Value
}

// escapeConditionLiteral escapes backslashes then double quotes so the wrapped
// literal round-trips through the condition tokenizer (inverse of the lexer's
// appendQuotedChar). Order matters: backslash first.
func escapeConditionLiteral(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	return s
}
```
Result: `Condition.Raw = ctx.tool_stdout = "say \"alpha||beta\""` (correctly escaped).

**`simulate/condition.go:tryTokenizeQuotedCond`** — honor backslash escapes on the **double-quote** read path, recovering the literal content; **single-quote behavior is unchanged** (YAML-style, no backslash processing — preserves existing behavior per the AC):
```go
func tryTokenizeQuotedCond(raw string, i int) (token string, consumed int) {
	if !isQuoteChar(raw[i]) {
		return "", 0
	}
	quote := raw[i]
	if quote == '"' {
		return readDoubleQuotedCond(raw, i) // escape-aware
	}
	return readSingleQuotedCond(raw, i)     // existing scan-to-quote behavior
}
```
`readDoubleQuotedCond` scans from `i+1`, collapsing `\"`→`"` and `\\`→`\`, until an **unescaped** `"`, and returns `(content, bytesConsumed)`. `readSingleQuotedCond` keeps today's scan-to-`'` logic.

The value stored in the token is the unescaped content, matching what the lexer produced originally — so `ParseCondition(Condition.Raw)` yields the same `CondCompare.Value` a direct parse would, and DIP010 passes for valid escaped conditions.

### 2. Reject unmatched double quotes (lexer-level, everywhere)

A lex-error mechanism, since none exists:

**`parser/lexer.go`** — add `errors []string` (or `[]lexError{msg, loc}`) to `Lexer`. `readQuotedContent` returns a `terminated bool`; `tryLexQuotedString`, when a `"`-string is unterminated, records `"unterminated string literal at L:C"` (with the opening-quote location) instead of silently accepting. The unterminated literal is still tokenized (best-effort recovery to keep lexing) but the recorded error makes the parse fail. Scope: **double quotes** (the reported bug); single-quoted YAML literals keep current behavior (an unterminated `'` may be handled as a consistent follow-up, out of scope here).

**`parser/parser.go:Parse()`** — after tokenization, fold the lexer's recorded errors into `p.diagnostics` (same list the parser already returns), so an unterminated quote fails `validate`/`lint` with a source location, exactly like other parse errors.

### Data-flow summary

```text
source  --lexer(unescape)-->  TokenLiteral.Value (literal content)
        --formatConditionToken(escape)-->  Condition.Raw (escaped, quoted)
        --tryTokenizeQuotedCond(unescape)-->  CondCompare.Value (literal content)   ← round-trips
unterminated "  --readQuotedContent(terminated=false)-->  lexer.errors  --Parse()-->  diagnostic (L:C)
```

## Non-goals

- Unterminated **single**-quoted literals (YAML-style) — a consistent follow-up; the bug and AC are about double quotes.
- Changing how conditions evaluate at runtime — this is purely source fidelity (Raw preservation) + a new rejection.
- The condition value's internal semantics (`||` etc. stay literal text inside the quotes).

## Testing

- **Parser regressions** (`parser/parse_edges_test.go` or a new file): an edge condition with `\"` interior quotes reconstructs `Condition.Raw` with the escapes intact and round-trips (`Format` → reparse identical); an unterminated `"` produces a diagnostic mentioning "unterminated" with a non-zero location.
- **Condition round-trip** (`simulate/condition_test.go`): `ParseCondition` on the escaped Raw yields the expected `CondCompare{Variable, Op, Value}` with the interior quotes restored.
- **Parse-to-validation integration:** the escaped-quote `.dip` passes `Lint`/`validate` (no DIP010); the unmatched-quote `.dip` fails with a located error. Mirror the two reproductions from the issue as fixtures.
- **Single-quote preservation:** an existing single-quoted condition value still tokenizes and validates unchanged.
- `just check` (pre-commit gate) passes.

## Acceptance criteria (from the issue)

- [ ] Double-quoted condition literals preserved in `Condition.Raw`, including escape syntax.
- [ ] Unmatched double quotes in hand-authored edge conditions rejected with a source location.
- [ ] The condition AST tokenizer honors escaped quotes and backslashes.
- [ ] Existing single-quoted condition behavior preserved.
- [ ] Parser and parse-to-validation integration regressions added.
- [ ] `just check` passes.
