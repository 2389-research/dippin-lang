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

// TestRankFamilySkipsAliases guards the alias double-count fix: byProvider holds
// a model under its canonical id AND each alias, so a family-tagged model with an
// alias must not be counted twice (which would let "stable" resolve to a second
// spelling of the newest model instead of the true one-release-back model).
func TestRankFamilySkipsAliases(t *testing.T) {
	models := map[string]ModelPrice{
		"opus-new":       {Family: "opus", Rank: 50, Priced: true, Aliases: []string{"opus-new-alias"}},
		"opus-new-alias": {Family: "opus", Rank: 50, Priced: true, Aliases: []string{"opus-new-alias"}},
		"opus-old":       {Family: "opus", Rank: 40, Priced: true},
	}
	got := rankFamily(models, "opus")
	if len(got) != 2 {
		t.Fatalf("rankFamily counted aliases: got %d candidates, want 2 (%v)", len(got), got)
	}
	if got[0].id != "opus-new" || got[1].id != "opus-old" {
		t.Errorf("rankFamily order wrong: got %v, want [opus-new opus-old] (latest=canonical, stable=true second)", got)
	}
}

func TestResolveModelRef(t *testing.T) {
	cases := []struct {
		nodeProvider, modelValue string
		wantConcrete             string
		wantResolved, wantAlias  bool
	}{
		{"anthropic", "opus@latest", "claude-opus-5", true, true},
		{"anthropic", "opus@stable", "claude-opus-4-8", true, true},
		{"anthropic", "sonnet@latest", "claude-sonnet-5", true, true},
		{"", "anthropic/opus@stable", "claude-opus-4-8", true, true},     // provider from prefix
		{"openai", "anthropic/opus@latest", "claude-opus-5", true, true}, // prefix overrides node
		{"anthropic", "bogus@latest", "", false, true},                   // alias, unresolvable
		{"anthropic", "opus@newest", "", false, false},                   // bad selector → not an alias
		{"anthropic", "claude-opus-5", "", false, false},                 // concrete id
		{"anthropic", "", "", false, false},                              // empty
	}
	for _, c := range cases {
		concrete, resolved, isAlias := ResolveModelRef(c.nodeProvider, c.modelValue)
		if concrete != c.wantConcrete || resolved != c.wantResolved || isAlias != c.wantAlias {
			t.Errorf("ResolveModelRef(%q,%q) = (%q,%v,%v), want (%q,%v,%v)",
				c.nodeProvider, c.modelValue, concrete, resolved, isAlias,
				c.wantConcrete, c.wantResolved, c.wantAlias)
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
