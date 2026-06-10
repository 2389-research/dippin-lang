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
