package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/2389-research/dippin-lang/pricing"
)

// runCheck prints the staleness report and exits non-zero if any entry is
// overdue — so it can gate CI / pre-commit as a re-verification reminder.
func runCheck(args []string) int {
	fs := flag.NewFlagSet("check", flag.ContinueOnError)
	days := fs.Int("max-age-days", maxAgeDays, "flag entries whose as_of is older than this")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	stale := pricing.StaleEntries(nowUTC(), time.Duration(*days)*24*time.Hour)
	if len(stale) == 0 {
		fmt.Printf("pricing: all entries verified within %d days\n", *days)
		return 0
	}
	fmt.Printf("pricing: %d entr%s need re-verification (window %d days):\n",
		len(stale), plural(len(stale)), *days)
	for _, s := range stale {
		fmt.Fprintf(os.Stdout, "  %-10s %-28s %s\n", s.Provider, s.Model, describeStale(s))
	}
	return 1
}

func describeStale(s pricing.Staleness) string {
	switch s.Reason {
	case "no-source":
		return "no source URL"
	case "bad-date":
		return fmt.Sprintf("unparseable as_of %q", s.AsOf)
	default:
		return fmt.Sprintf("as_of %s (%d days old)", s.AsOf, s.AgeDays)
	}
}

func plural(n int) string {
	if n == 1 {
		return "y"
	}
	return "ies"
}
