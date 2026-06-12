package main

import (
	"flag"
	"fmt"

	"github.com/2389-research/dippin-lang/validator"
)

// CmdValidate runs structural validation (DIP001-DIP010) on a workflow.
func (c *CLI) CmdValidate(args []string) ExitCode {
	path, code := parseSingleFileArg("validate", "usage: dippin validate <file>", args, c.Stderr)
	if code != ExitCode(-1) {
		return code
	}

	w, err := loadWorkflow(path)
	if err != nil {
		c.renderError(err, path)
		return ExitError
	}

	res := validator.Validate(w)
	c.renderDiagnostics(res.Diagnostics)

	if res.HasErrors() {
		return ExitError
	}
	if c.Format == FormatText {
		fmt.Fprintln(c.Stdout, "validation passed")
	}
	return ExitOK
}

// parseLintArgs parses lint/doctor flags and returns (path, extraModels, ok).
// It writes usage to c.Stderr and returns ExitUsageError on failure.
func parseLintArgs(name, usage string, args []string, c *CLI) (path, extraModels string, code ExitCode) {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(c.Stderr)
	em := fs.String("extra-models", "", "extend model catalog: provider:model1,model2;provider2:model3")
	if err := fs.Parse(args); err != nil {
		return "", "", ExitUsageError
	}
	if fs.NArg() < 1 {
		fmt.Fprintln(c.Stderr, usage)
		return "", "", ExitUsageError
	}
	return fs.Arg(0), *em, ExitCode(-1)
}

// lintOptions builds scoped lint options from an --extra-models spec. An empty
// spec yields default options (nil extra catalog).
func lintOptions(extraModels string) validator.Options {
	if extraModels == "" {
		return validator.Options{}
	}
	return validator.Options{ExtraModels: validator.ParseExtraModels(extraModels)}
}

// CmdLint runs both structural validation and semantic linting.
// Errors cause exit 1; warnings alone exit 0.
func (c *CLI) CmdLint(args []string) ExitCode {
	path, extraModels, code := parseLintArgs("lint", "usage: dippin lint [--extra-models spec] <file>", args, c)
	if code != ExitCode(-1) {
		return code
	}

	w, err := loadWorkflow(path)
	if err != nil {
		c.renderError(err, path)
		return ExitError
	}

	// Run both validation and linting per spec — lint includes all checks.
	// Extra models are passed as scoped options, never mutating global state.
	valRes := validator.Validate(w)
	lintRes := validator.LintWithOptions(w, lintOptions(extraModels))

	// Merge all diagnostics, then apply the native cross-file tool_access pass
	// (DIP146) — it supersedes DIP143 for boundaries whose child it resolved.
	allDiags := append(valRes.Diagnostics, lintRes.Diagnostics...)
	allDiags = applyCrossFileToolAccess(allDiags, w, path)
	c.renderDiagnostics(allDiags)

	// Exit 1 only if there are errors; warnings alone pass.
	if valRes.HasErrors() {
		return ExitError
	}
	return ExitOK
}
