// Command pricing-sync is the assistive tooling for keeping pricing/prices.json
// current. It never edits the catalog itself — it reports.
//
//	pricing-sync check              staleness report (AsOf age / missing source)
//	pricing-sync sync               diff the catalog against machine-readable
//	                                aggregators (models.dev) and report candidates
//
// Detection is mechanical; authoritative verification is not (no provider
// publishes price as data), so this prints proposals for a human to confirm
// against each entry's official Source. See
// docs/superpowers/specs/2026-08-07-pricing-package-and-autosync-design.md.
package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"
)

// stdOut is the report destination; a var so tests can capture output.
var stdOut io.Writer = os.Stdout

func printfOut(format string, a ...any) { fmt.Fprintf(stdOut, format, a...) }

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "usage: pricing-sync <check|sync> [--max-age-days N]")
		return 2
	}
	switch args[0] {
	case "check":
		return runCheck(args[1:])
	case "sync":
		return runSync(context.Background(), args[1:])
	default:
		fmt.Fprintf(os.Stderr, "unknown subcommand %q (want check|sync)\n", args[0])
		return 2
	}
}

// maxAgeDays is the default freshness window before an entry is flagged stale.
const maxAgeDays = 45

func nowUTC() time.Time { return time.Now().UTC() }
