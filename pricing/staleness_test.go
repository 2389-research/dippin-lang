package pricing

import (
	"testing"
	"time"
)

func TestStaleEntriesFlagsOldAsOf(t *testing.T) {
	// A generous window: nothing is stale far enough back.
	early := time.Date(2026, 6, 15, 0, 0, 0, 0, time.UTC)
	if got := StaleEntries(early, 365*24*time.Hour); len(got) != 0 {
		t.Errorf("with a 1-year window in mid-2026, want 0 stale, got %d: %+v", len(got), got)
	}

	// A 60-day window well after the 2026-06-10 batch: those entries are stale,
	// the 2026-08-07 additions are not.
	now := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	stale := StaleEntries(now, 60*24*time.Hour)
	if len(stale) == 0 {
		t.Fatal("expected the 2026-06-10 entries to be stale under a 60-day window")
	}
	for _, s := range stale {
		if s.Reason != "stale" {
			t.Errorf("%s/%s: reason %q, want stale", s.Provider, s.Model, s.Reason)
		}
		if s.AsOf == "2026-08-07" {
			t.Errorf("%s/%s dated 2026-08-07 must not be stale under a 60-day window as of 2026-09-01", s.Provider, s.Model)
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
