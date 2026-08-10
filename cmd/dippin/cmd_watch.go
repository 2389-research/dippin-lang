package main

import (
	"flag"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"

	"github.com/2389-research/dippin-lang/validator"
)

const debounceDelay = 200 * time.Millisecond

// CmdWatch watches .dip files and re-runs lint on changes.
func (c *CLI) CmdWatch(args []string) ExitCode {
	fs := flag.NewFlagSet("watch", flag.ContinueOnError)
	fs.SetOutput(c.Stderr)

	if err := fs.Parse(args); err != nil {
		return ExitUsageError
	}
	if fs.NArg() < 1 {
		fmt.Fprintln(c.Stderr, "usage: dippin watch <file-or-dir> [...]")
		return ExitUsageError
	}

	return c.runWatch(fs.Args())
}

// runWatch sets up the file watcher and enters the event loop.
func (c *CLI) runWatch(targets []string) ExitCode {
	watcher, err := setupWatcher(targets)
	if err != nil {
		fmt.Fprintf(c.Stderr, "error: %v\n", err)
		return ExitError
	}
	defer func() { _ = watcher.Close() }()

	fmt.Fprintf(c.Stdout, "watching %d target(s) for .dip changes...\n", len(targets))
	c.watchLoop(watcher)
	return ExitOK
}

// setupWatcher creates an fsnotify watcher and adds all target paths.
func setupWatcher(targets []string) (*fsnotify.Watcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("create watcher: %w", err)
	}
	for _, target := range targets {
		if err := watcher.Add(target); err != nil {
			_ = watcher.Close()
			return nil, fmt.Errorf("watch %q: %w", target, err)
		}
	}
	return watcher, nil
}

// watchState tracks debounce state for the file watcher event loop.
type watchState struct {
	timer   *time.Timer
	pending string
}

func newWatchState() *watchState {
	timer := time.NewTimer(0)
	if !timer.Stop() {
		<-timer.C
	}
	return &watchState{timer: timer}
}

// watchLoop processes file system events with debouncing.
func (c *CLI) watchLoop(watcher *fsnotify.Watcher) {
	ws := newWatchState()
	for {
		if !ws.waitForEvent(watcher, c) {
			return
		}
	}
}

// waitForEvent blocks until a watcher event, error, or debounce timer fires.
// Returns false when the watcher is closed.
func (ws *watchState) waitForEvent(watcher *fsnotify.Watcher, c *CLI) bool {
	select {
	case ev := <-watcher.Events:
		ws.handleFsEvent(ev)
	case err := <-watcher.Errors:
		fmt.Fprintf(c.Stderr, "watcher error: %v\n", err)
	case <-ws.timer.C:
		ws.drain(c)
	}
	return true
}

func (ws *watchState) handleFsEvent(ev fsnotify.Event) {
	if f := filterDipEvent(ev); f != "" {
		ws.pending = f
		ws.timer.Reset(debounceDelay)
	}
}

func (ws *watchState) drain(c *CLI) {
	if ws.pending != "" {
		c.lintAndPrint(ws.pending)
		ws.pending = ""
	}
}

// filterDipEvent returns the .dip file path if the event is a write/create.
func filterDipEvent(ev fsnotify.Event) string {
	if !ev.Has(fsnotify.Write) && !ev.Has(fsnotify.Create) {
		return ""
	}
	if strings.HasSuffix(ev.Name, ".dip") {
		return ev.Name
	}
	return ""
}

// lintAndPrint parses and lints a single .dip file, printing results.
func (c *CLI) lintAndPrint(path string) {
	fmt.Fprintf(c.Stdout, "\n\u2500\u2500\u2500 %s \u2500\u2500\u2500\n", filepath.Base(path))

	w, err := parseFile(path)
	if err != nil {
		fmt.Fprintf(c.Stderr, "parse error: %v\n", err)
		return
	}

	valRes := validator.Validate(w)
	lintRes := validator.Lint(w)
	all := append(valRes.Diagnostics, lintRes.Diagnostics...)
	all = applyCrossFileChecks(all, w, path)

	if len(all) == 0 {
		fmt.Fprintln(c.Stdout, "  \u2714 no issues")
		return
	}
	c.renderDiagnostics(all)
}
