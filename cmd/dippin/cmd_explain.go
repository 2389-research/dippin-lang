package main

import (
	"fmt"
	"io"
	"regexp"
	"sort"
	"strings"

	"github.com/2389-research/dippin-lang/validator"
)

var dipCodeRe = regexp.MustCompile(`^DIP\d+$`)

// CmdExplain displays detailed documentation for a DIP diagnostic code.
func (c *CLI) CmdExplain(args []string) ExitCode {
	if len(args) != 1 {
		fmt.Fprintln(c.Stderr, "usage: dippin explain <DIPxxx>")
		return ExitUsageError
	}

	code := strings.ToUpper(args[0])
	if !dipCodeRe.MatchString(code) {
		fmt.Fprintf(c.Stderr, "invalid code format: %s (expected DIPxxx)\n", args[0])
		return ExitUsageError
	}

	exp, ok := validator.Explanations[code]
	if !ok {
		fmt.Fprintf(c.Stderr, "unknown code: %s\n\n", code)
		printValidCodes(c.Stderr)
		return ExitError
	}

	return c.renderExplanation(exp)
}

func (c *CLI) renderExplanation(exp validator.Explanation) ExitCode {
	if c.Format == FormatJSON {
		return c.renderJSON(exp)
	}
	renderExplanationText(c.Stdout, exp)
	return ExitOK
}

func renderExplanationText(w io.Writer, exp validator.Explanation) {
	header := fmt.Sprintf("═══ %s ", exp.Code)
	header += strings.Repeat("═", max(0, 58-len(header)))
	fmt.Fprintln(w, header)
	fmt.Fprintf(w, "  %s\n", exp.Summary)
	fmt.Fprintln(w)
	fmt.Fprintf(w, "  Trigger: %s\n", exp.Trigger)
	fmt.Fprintf(w, "  Fix:     %s\n", exp.Fix)
	fmt.Fprintln(w)
	fmt.Fprintln(w, "  Example:")
	for _, line := range strings.Split(exp.Example, "\n") {
		fmt.Fprintf(w, "    %s\n", line)
	}
}

func printValidCodes(w io.Writer) {
	codes := make([]string, 0, len(validator.Explanations))
	for code := range validator.Explanations {
		codes = append(codes, code)
	}
	sort.Strings(codes)
	fmt.Fprintf(w, "valid codes: %s\n", strings.Join(codes, ", "))
}
