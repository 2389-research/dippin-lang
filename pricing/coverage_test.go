package pricing

import (
	"reflect"
	"sort"
	"testing"
)

// knownCacheGaps is the allowlist of priced models that carry no cache rate,
// each for a documented reason. TestCacheGapsAllowlist fails if this set drifts
// — so a newly added priced model cannot ship without either a verified cache
// rate in prices.json or a deliberate, reasoned entry here.
var knownCacheGaps = []string{
	// glm-4-32b-0414-128k: Z.AI's pricing page lists no cached-input rate for it
	// (every other GLM entry has one). Unknown, not "no discount".
	"zai/glm-4-32b-0414-128k",
	// MiniMax M2.x: retiring 2026-11-18 and no longer on the pay-as-you-go
	// page, so their cache rates can't be confirmed. M3 / M3.1-Flash carry the
	// rates printed in the page's "Prompt caching Read" column, confirmed by a
	// human on 2026-10-19 (#292).
	"minimax/MiniMax-M2.5",
	"minimax/MiniMax-M2.5-highspeed",
	"minimax/MiniMax-M2.7",
	"minimax/MiniMax-M2.7-highspeed",
	// qwen3.8-*: Alibaba's Context Cache page lists both under implicit and
	// explicit caching but says their hit price is not the standard 20%/10% of
	// input and defers to the login-gated console. No public rate to commit.
	"qwen/qwen3.8-flash",
	"qwen/qwen3.8-max",
}

func TestCacheGapsAllowlist(t *testing.T) {
	got := CacheGaps()
	want := append([]string(nil), knownCacheGaps...)
	sort.Strings(want)
	if !reflect.DeepEqual(got, want) {
		t.Errorf("cache-coverage gap set changed.\n got:  %v\n want: %v\n"+
			"A priced model has no cache rate. Add a verified rate to prices.json, "+
			"or — if the provider publishes none — add it to knownCacheGaps with a reason.",
			got, want)
	}
}
