package pricing

import "testing"

func TestResolveAlias(t *testing.T) {
	cases := []struct {
		provider, family, selector, want string
		ok                               bool
	}{
		{"anthropic", "opus", "latest", "claude-opus-5", true},
		{"anthropic", "opus", "sota", "claude-opus-5", true},
		{"anthropic", "opus", "stable", "claude-opus-4-8", true},
		{"anthropic", "sonnet", "latest", "claude-sonnet-5", true},
		{"anthropic", "sonnet", "stable", "claude-sonnet-4-6", true},
		{"anthropic", "haiku", "latest", "claude-haiku-4-5", true}, // haiku-3-5 is deprecated → excluded
		{"anthropic", "haiku", "stable", "claude-haiku-4-5", true}, // only one eligible → falls back to latest
		{"anthropic", "nope", "latest", "", false},
		{"anthropic", "opus", "bogus", "", false},
	}
	for _, c := range cases {
		got, ok := ResolveAlias(c.provider, c.family, c.selector)
		if got != c.want || ok != c.ok {
			t.Errorf("ResolveAlias(%q,%q,%q) = (%q,%v), want (%q,%v)",
				c.provider, c.family, c.selector, got, ok, c.want, c.ok)
		}
	}
}

func TestResolveAliasNeverReturnsDeprecated(t *testing.T) {
	for _, sel := range []string{"latest", "stable", "sota"} {
		if got, _ := ResolveAlias("anthropic", "opus", sel); got == "claude-opus-4-0" || got == "claude-opus-4-1" {
			t.Errorf("selector %q surfaced deprecated model %q", sel, got)
		}
	}
}
