package pricing

import (
	"testing"
	"time"
)

func TestStaleEntriesFlagsOldAsOf(t *testing.T) {
	// A generous window: nothing is stale far enough back.
	early := time.Date(2026, 8, 8, 0, 0, 0, 0, time.UTC)
	if got := StaleEntries(early, 365*24*time.Hour); len(got) != 0 {
		t.Errorf("with a 1-year window in early Aug 2026, want 0 stale, got %d: %+v", len(got), got)
	}

	// Far enough ahead that every entry has aged out of a 60-day window: all of
	// them are reported as stale except the explicit price_unlisted ones.
	now := time.Date(2028, 1, 1, 0, 0, 0, 0, time.UTC)
	stale := StaleEntries(now, 60*24*time.Hour)
	if len(stale) == 0 {
		t.Fatal("expected live entries to be stale under a 60-day window in 2028")
	}
	for _, s := range stale {
		if s.Reason != "stale" {
			t.Errorf("%s/%s: reason %q, want stale", s.Provider, s.Model, s.Reason)
		}
		if p, _ := LookupProvider(s.Provider, s.Model); p.PriceUnlisted {
			t.Errorf("%s/%s is price_unlisted and must not be reported stale", s.Provider, s.Model)
		}
	}
}

func TestStaleEntriesDeterministicOrder(t *testing.T) {
	now := time.Date(2026, 12, 1, 0, 0, 0, 0, time.UTC)
	a := StaleEntries(now, 24*time.Hour)
	b := StaleEntries(now, 24*time.Hour)
	if len(a) != len(b) {
		t.Fatal("nondeterministic length")
	}
	for i := range a {
		if a[i] != b[i] {
			t.Fatalf("order differs at %d: %+v vs %+v", i, a[i], b[i])
		}
	}
	// Sorted by provider then model.
	for i := 1; i < len(a); i++ {
		if a[i-1].Provider > a[i].Provider {
			t.Errorf("not sorted by provider at %d", i)
		}
	}
}

// Only an explicit PriceUnlisted entry is exempt from the age check: a
// deprecated model with a still-published price keeps being re-verified, and a
// missing source or bad date is always flagged.
func TestStaleEntriesSkipsOnlyPriceUnlistedByAge(t *testing.T) {
	now := time.Date(2027, 6, 1, 0, 0, 0, 0, time.UTC)
	window := 45 * 24 * time.Hour
	old := ModelPrice{Source: "https://example.com/pricing", AsOf: "2026-01-01"}
	if _, ok := evaluateStaleness("p", "live", old, now, window); !ok {
		t.Error("an old live entry must be stale")
	}
	old.Deprecated = true
	if _, ok := evaluateStaleness("p", "legacy", old, now, window); !ok {
		t.Error("an old deprecated entry whose price is still published must stay stale-checked")
	}
	old.PriceUnlisted = true
	if s, ok := evaluateStaleness("p", "retired", old, now, window); ok {
		t.Errorf("an old price_unlisted entry must not be flagged by age, got %+v", s)
	}
	noSource := ModelPrice{PriceUnlisted: true, AsOf: "2026-01-01"}
	if s, ok := evaluateStaleness("p", "retired", noSource, now, window); !ok || s.Reason != "no-source" {
		t.Errorf("a deprecated entry without a source must still be flagged no-source, got %+v ok=%v", s, ok)
	}
}
