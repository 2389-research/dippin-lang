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
	// deepseek-chat/-reasoner: rolling aliases, not on the current DeepSeek
	// pricing page (which lists deepseek-v4-flash/-pro), so no cache-hit price
	// to verify against.
	"deepseek/deepseek-chat",
	"deepseek/deepseek-reasoner",
	// grok-4-1-fast-*: superseded, absent from the current x.ai models page.
	"grok/grok-4-1-fast-non-reasoning",
	"grok/grok-4-1-fast-reasoning",
	// MiniMax: the pricing page does now show "Prompt caching Read"/"Write"
	// columns for these models, so the earlier "publishes no token cache price"
	// note is stale. It is still not reliably machine-readable: two automated
	// reads of the same page disagreed on the read rate ($0.03/M vs. $0.06/M
	// for <=512k with $0.12/M above it), and M3 shows no write price at all.
	// Left as a gap rather than committing an unconfirmed number — a human must
	// read the table and set the rates. Tracked in #292.
	"minimax/MiniMax-M2",
	"minimax/MiniMax-M2.1",
	"minimax/MiniMax-M2.5",
	"minimax/MiniMax-M2.5-highspeed",
	"minimax/MiniMax-M2.7",
	"minimax/MiniMax-M2.7-highspeed",
	"minimax/MiniMax-M3",
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
