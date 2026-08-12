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

// TestDIP108FrontierCatalog covers #189: current frontier models across every
// provider (incl. the new zai/moonshot/minimax/qwen providers) must be
// recognized so they don't trip a spurious DIP108.
func TestDIP108FrontierCatalog(t *testing.T) {
	cases := []struct{ provider, model string }{
		{"anthropic", "claude-opus-5"},
		{"anthropic", "claude-sonnet-5"},
		{"openai", "gpt-5.6-sol"},
		{"gemini", "gemini-3.6-flash"},
		{"xai", "grok-4.5"},
		{"zai", "glm-5.2"},
		{"moonshot", "kimi-k3"},
		{"minimax", "MiniMax-M3"},
		{"qwen", "qwen3.7-max"},
	}
	for _, c := range cases {
		src := fmt.Sprintf(`workflow w
  start: A
  exit: A

  agent A
    provider: %s
    model: %s
    prompt: go

  edges
    A -> A
`, c.provider, c.model)
		if diags := lintSrc(t, src); hasCode(diags, DIP108) {
			t.Errorf("%s/%s tripped DIP108 (should be known): %v", c.provider, c.model, codes(diags))
		}
	}
}

// TestLintDIP161DeprecatedModel covers #264 Phase 0: an agent pinned to a
// catalog model flagged deprecated must trip DIP161 (a warning), while a
// current model must not. Deprecated IDs live in pricing/prices.json
// ("deprecated": true) — e.g. claude-opus-4-1, claude-sonnet-4-0.
func TestLintDIP161DeprecatedModel(t *testing.T) {
	dip161Src := func(model string) string {
		return fmt.Sprintf(`workflow w
  start: A
  exit: A

  agent A
    provider: anthropic
    model: %s
    prompt: go

  edges
    A -> A
`, model)
	}

	for _, model := range []string{"claude-opus-4-1", "claude-sonnet-4-0"} {
		diags := lintSrc(t, dip161Src(model))
		if !hasCode(diags, DIP161) {
			t.Errorf("deprecated model %q did not trip DIP161: %v", model, codes(diags))
		}
		if hasCode(diags, DIP108) {
			t.Errorf("deprecated model %q wrongly tripped DIP108 (it is in the catalog): %v", model, codes(diags))
		}
	}

	// A current, non-deprecated model must not trip DIP161.
	if diags := lintSrc(t, dip161Src("claude-opus-4-8")); hasCode(diags, DIP161) {
		t.Errorf("current model claude-opus-4-8 wrongly tripped DIP161: %v", codes(diags))
	}
}
