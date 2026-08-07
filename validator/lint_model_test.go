package validator

import (
	"fmt"
	"testing"
)

// TestLintDIP108NewAnthropicModels guards the #116 catalog refresh: the
// 2026-05/06 Anthropic IDs (incl. the invite-only Mythos line, which are real
// IDs) must be recognized so approved users don't get spurious DIP108 warnings.
func TestLintDIP108NewAnthropicModels(t *testing.T) {
	models := []string{
		"claude-opus-4-8",
		"claude-fable-5",
		"claude-mythos-5",
		"claude-mythos-preview",
	}
	for _, model := range models {
		src := fmt.Sprintf(`workflow w
  start: A
  exit: A

  agent A
    provider: anthropic
    model: %s
    prompt: go

  edges
    A -> A
`, model)
		diags := lintSrc(t, src)
		if hasCode(diags, DIP108) {
			t.Errorf("model %q tripped DIP108 (should be a known model): %v", model, codes(diags))
		}
	}
}

// TestLintDIP108DottedAnthropicID covers #188: a dotted Anthropic model ID
// (claude-haiku-4.5, the Vercel-gateway-documented spelling) must be recognized
// as the same model as the dashed catalog key (claude-haiku-4-5), so it does not
// trip DIP108. A genuinely unknown model must still be flagged.
func TestLintDIP108DottedAnthropicID(t *testing.T) {
	known := fmt.Sprintf(`workflow w
  start: A
  exit: A

  agent A
    provider: anthropic
    model: %s
    prompt: go

  edges
    A -> A
`, "claude-haiku-4.5")
	if diags := lintSrc(t, known); hasCode(diags, DIP108) {
		t.Errorf("dotted claude-haiku-4.5 tripped DIP108 (issue #188): %v", codes(diags))
	}

	unknown := fmt.Sprintf(`workflow w
  start: A
  exit: A

  agent A
    provider: anthropic
    model: %s
    prompt: go

  edges
    A -> A
`, "claude-nonexistent-9.9")
	if diags := lintSrc(t, unknown); !hasCode(diags, DIP108) {
		t.Errorf("genuinely unknown dotted model must still trip DIP108: %v", codes(diags))
	}
}
