package pricing

import (
	"encoding/json"
	"strings"
	"testing"
)

// retiredSlugs are model ids a provider has retired (or folded into another
// model) that must never be modeled as an alias of a live entry: an alias
// inherits its target's Deprecated flag, so DIP161 could never warn on them.
// Each must be its own deprecated entry instead. Found in review twice in one
// week (#334: gpt-5.2-codex, grok-code-fast-1); add new ones as they retire.
var retiredSlugs = map[string][]string{
	"openai": {"gpt-5.2-codex"},
	"grok":   {"grok-code-fast-1", "grok-code-fast", "grok-code-fast-1-0825", "grok-4-1-fast-reasoning", "grok-4-1-fast-non-reasoning"},
}

// TestAliasesInheritCorrectFlags guards the alias pitfall: an alias is the same
// model under another spelling and shares every flag with its target. A
// preview snapshot or a retired slug is not that, so it needs its own entry.
func TestAliasesInheritCorrectFlags(t *testing.T) {
	for _, m := range loadEntriesForTest(t) {
		for _, a := range m.Aliases {
			if strings.HasSuffix(a, "-preview") && m.Maturity != "preview" {
				t.Errorf("%s/%s: alias %q looks like a preview but its target isn't maturity=preview; give it its own entry", m.Provider, m.Model, a)
			}
		}
	}
	for provider, slugs := range retiredSlugs {
		for _, slug := range slugs {
			p, ok := LookupProvider(provider, slug)
			if !ok {
				t.Errorf("%s/%s: retired slug must still resolve (as a deprecated entry)", provider, slug)
				continue
			}
			if !p.Deprecated {
				t.Errorf("%s/%s: retired slug resolves to a live entry; it must be its own deprecated entry, not an alias", provider, slug)
			}
		}
	}
}

// loadEntriesForTest parses the embedded catalog's raw entries (canonical ids
// with their declared aliases and maturity).
func loadEntriesForTest(t *testing.T) []fileEntry {
	t.Helper()
	var pf priceFile
	if err := json.Unmarshal(pricesJSON, &pf); err != nil {
		t.Fatalf("embedded prices.json does not parse: %v", err)
	}
	return pf.Models
}
