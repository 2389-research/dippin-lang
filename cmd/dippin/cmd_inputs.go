// ABOUTME: `dippin inputs` prints a workflow's declared input schema — the
// ABOUTME: introspection surface a host uses to collect values before a run.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"strconv"

	"github.com/2389-research/dippin-lang/ir"
)

// inputJSON is the stable wire shape of one declared input. Defaults and bounds
// are coerced to real JSON types per the declared type, so a host can inject
// them typed instead of re-flattening everything to string.
type inputJSON struct {
	Name        string   `json:"name"`
	Type        string   `json:"type"`
	Required    bool     `json:"required"`
	Default     any      `json:"default,omitempty"`
	Prompt      string   `json:"prompt,omitempty"`
	Description string   `json:"description,omitempty"`
	Options     []string `json:"options,omitempty"`
	Pattern     string   `json:"pattern,omitempty"`
	Min         any      `json:"min,omitempty"`
	Max         any      `json:"max,omitempty"`
	MaxLength   int      `json:"max_length,omitempty"`
	Multiline   bool     `json:"multiline,omitempty"`
}

// CmdInputs is the dispatcher entry point.
func (c *CLI) CmdInputs(args []string) ExitCode {
	fs := flag.NewFlagSet("inputs", flag.ContinueOnError)
	fs.SetOutput(c.Stderr)
	format := fs.String("format", "text", "output format: text|json")
	if err := fs.Parse(args); err != nil {
		return ExitError
	}
	if fs.NArg() != 1 {
		fmt.Fprintln(c.Stderr, "usage: dippin inputs [--format=text|json] <file>")
		return ExitError
	}

	w, err := loadWorkflow(fs.Arg(0))
	if err != nil {
		c.renderError(err, fs.Arg(0))
		return ExitError
	}
	return writeInputsInFormat(c.Stdout, c.Stderr, w, *format)
}

// writeInputsInFormat dispatches to the requested output format, rejecting
// anything other than "text" or "json" so a typo'd --format fails loudly
// instead of silently falling back to human text.
func writeInputsInFormat(out, stderr io.Writer, w *ir.Workflow, format string) ExitCode {
	switch format {
	case "json":
		return writeInputsJSON(out, w)
	case "text":
		return writeInputsText(out, w)
	default:
		fmt.Fprintf(stderr, "unknown --format value: %q (expected text or json)\n", format)
		return ExitError
	}
}

// writeInputsJSON emits the schema array. A workflow with no inputs emits [],
// never null — a host iterating the result should not have to nil-check.
func writeInputsJSON(out io.Writer, w *ir.Workflow) ExitCode {
	enc := json.NewEncoder(out)
	enc.SetIndent("", "  ")
	if err := enc.Encode(inputsJSON(w)); err != nil {
		return ExitError
	}
	return ExitOK
}

// inputsJSON projects the IR into the wire shape, preserving declaration order.
func inputsJSON(w *ir.Workflow) []inputJSON {
	out := make([]inputJSON, 0, len(w.Inputs))
	for _, in := range w.Inputs {
		out = append(out, oneInputJSON(in))
	}
	return out
}

// oneInputJSON projects a single input.
func oneInputJSON(in *ir.Input) inputJSON {
	j := inputJSON{
		Name:        in.Name,
		Type:        in.Type,
		Required:    in.Required,
		Prompt:      in.Prompt,
		Description: in.Description,
		Options:     in.Options,
		Pattern:     in.Pattern,
		MaxLength:   in.MaxLength,
		Multiline:   in.Multiline,
	}
	if in.HasDefault {
		j.Default = coerceInputValue(in.Type, in.Default)
	}
	if in.Min != "" {
		j.Min = coerceInputValue(in.Type, in.Min)
	}
	if in.Max != "" {
		j.Max = coerceInputValue(in.Type, in.Max)
	}
	return j
}

// coerceInputValue converts raw declaration text to a JSON-native value per the
// declared type. The IR keeps raw text so the formatter round-trips a file
// byte-for-byte; typing happens here. An uncoercible value falls back to the
// raw string — DIP155 reports the declaration defect, not this projection.
func coerceInputValue(typ, raw string) any {
	switch typ {
	case "number":
		if n, err := strconv.ParseFloat(raw, 64); err == nil {
			return n
		}
	case "bool":
		if b, err := strconv.ParseBool(raw); err == nil {
			return b
		}
	}
	return raw
}

// writeInputsText emits a human-readable listing.
func writeInputsText(out io.Writer, w *ir.Workflow) ExitCode {
	if len(w.Inputs) == 0 {
		fmt.Fprintln(out, "no declared inputs")
		return ExitOK
	}
	for _, in := range w.Inputs {
		req := "optional"
		if in.Required {
			req = "required"
		}
		fmt.Fprintf(out, "%s: %s (%s)\n", in.Name, in.Type, req)
		writeOneInputTextDetail(out, in)
	}
	return ExitOK
}

// writeOneInputTextDetail emits the indented detail lines for one input.
func writeOneInputTextDetail(out io.Writer, in *ir.Input) {
	writeInputTextBasics(out, in)
	writeInputRangeConstraints(out, in)
	writeInputLengthConstraints(out, in)
}

// writeInputTextBasics emits prompt/description/default/options.
func writeInputTextBasics(out io.Writer, in *ir.Input) {
	if in.Prompt != "" {
		fmt.Fprintf(out, "    prompt: %s\n", in.Prompt)
	}
	if in.Description != "" {
		fmt.Fprintf(out, "    description: %s\n", in.Description)
	}
	if in.HasDefault {
		fmt.Fprintf(out, "    default: %s\n", in.Default)
	}
	if len(in.Options) > 0 {
		fmt.Fprintf(out, "    options: %v\n", in.Options)
	}
}

// writeInputRangeConstraints emits pattern/min/max.
func writeInputRangeConstraints(out io.Writer, in *ir.Input) {
	if in.Pattern != "" {
		fmt.Fprintf(out, "    pattern: %s\n", in.Pattern)
	}
	if in.Min != "" {
		fmt.Fprintf(out, "    min: %s\n", in.Min)
	}
	if in.Max != "" {
		fmt.Fprintf(out, "    max: %s\n", in.Max)
	}
}

// writeInputLengthConstraints emits max_length/multiline.
func writeInputLengthConstraints(out io.Writer, in *ir.Input) {
	if in.MaxLength > 0 {
		fmt.Fprintf(out, "    max_length: %d\n", in.MaxLength)
	}
	if in.Multiline {
		fmt.Fprintf(out, "    multiline: %v\n", in.Multiline)
	}
}
