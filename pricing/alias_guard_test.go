package pricing

import (
	"encoding/json"
	"strings"
	"testing"
)

// retiredSlugs maps each retired (or folded-in) model slug to the canonical
// catalog entry that must own it: the slug itself, or — for an alternate
// spelling of the same retired model — that model's entry. An alias inherits
// its owner's price, Deprecated flag and metadata, so a retired slug must never
// hang off a live entry or an unrelated deprecated one; DIP161 and pricing
// would silently describe the wrong model. Found in review on #334 and #337;
// add new ones as models retire.
var retiredSlugs = map[string]string{
	"openai/gpt-5.2-codex":             "gpt-5.2-codex",
	"grok/grok-code-fast-1":            "grok-code-fast-1",
	"grok/grok-code-fast":              "grok-code-fast-1",
	"grok/grok-code-fast-1-0825":       "grok-code-fast-1",
	"grok/grok-4-1-fast-reasoning":     "grok-4-1-fast-reasoning",
	"grok/grok-4-1-fast-non-reasoning": "grok-4-1-fast-non-reasoning",
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
	checkRetiredSlugs(t, loadEntriesForTest(t))
}

// checkRetiredSlugs requires each retired slug to resolve, in the raw catalog
// data, to exactly its designated deprecated owner entry.
func checkRetiredSlugs(t *testing.T, entries []fileEntry) {
	t.Helper()
	owner := map[string]fileEntry{}
	for _, e := range entries {
		owner[e.Provider+"/"+e.Model] = e
		for _, a := range e.Aliases {
			owner[e.Provider+"/"+a] = e
		}
	}
	for key, want := range retiredSlugs {
		e, ok := owner[key]
		switch {
		case !ok:
			t.Errorf("%s: retired slug must still resolve (as entry %s)", key, want)
		case e.Model != want:
			t.Errorf("%s: retired slug is owned by %s, want %s; an alias inherits its owner's price and flags", key, e.Model, want)
		case !e.Deprecated:
			t.Errorf("%s: owner %s must be deprecated", key, e.Model)
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
