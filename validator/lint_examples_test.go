package validator_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/2389-research/dippin-lang/dipx"
	"github.com/2389-research/dippin-lang/parser"
	"github.com/2389-research/dippin-lang/validator"
)

// TestLintExamples parses every .dip file in examples/ through the real
// parser and lints it, asserting zero DIP108 (unknown model) warnings, zero
// DIP147 (chain-attack topology), zero DIP152 (unrouted marker_grep markers),
// and zero DIP153 (redundant parallel/fan_in edges). This catches model catalog
// staleness, invalid model IDs, and guards
// against examples demonstrating restricted->tool-bearing laundering or
// unreachable marker routing.
func TestLintExamples(t *testing.T) {
	examples, err := filepath.Glob("../examples/*.dip")
	if err != nil {
		t.Fatalf("glob examples: %v", err)
	}
	subExamples, err := filepath.Glob("../examples/*/*.dip")
	if err != nil {
		t.Fatalf("glob sub-examples: %v", err)
	}
	examples = append(examples, subExamples...)
	if len(examples) == 0 {
		t.Fatal("no .dip files found in examples/")
	}

	for _, path := range examples {
		name := filepath.Base(path)
		t.Run(name, func(t *testing.T) {
			src, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("read %s: %v", path, err)
			}

			p := parser.NewParser(string(src), path)
			w, err := p.Parse()
			if err != nil {
				t.Fatalf("parse error in %s: %v", name, err)
			}

			result := validator.Lint(w)
			for _, d := range result.Diagnostics {
				if d.Code == validator.DIP108 || d.Code == validator.DIP147 || d.Code == validator.DIP152 || d.Code == validator.DIP153 {
					t.Errorf("%s: %s", name, d.Message)
				}
			}
		})
	}
}

// TestPlaygroundExamplesLintClean guards every example embedded in the website
// playground (the EXAMPLES object in site/layouts/_default/playground.html): each
// one a visitor can load MUST lint completely clean — zero diagnostics of any
// severity — through the same passes the WASM playground runs (Validate + Lint).
// Without this, a lint that newly fires (or an example edit) silently ships a
// warning-producing demo, which is how the DIP144 "Greeter has no failure route"
// regression reached the live playground.
func TestPlaygroundExamplesLintClean(t *testing.T) {
	const playgroundHTML = "../site/layouts/_default/playground.html"
	raw, err := os.ReadFile(playgroundHTML)
	if err != nil {
		t.Fatalf("read %s: %v", playgroundHTML, err)
	}
	examples := extractPlaygroundExamples(t, string(raw))
	if len(examples) < 6 {
		t.Fatalf("expected >= 6 embedded playground examples, found %d", len(examples))
	}
	for _, src := range examples {
		name := strings.TrimSpace(strings.SplitN(src, "\n", 2)[0])
		t.Run(name, func(t *testing.T) {
			w, err := parser.NewParser(src, "playground.dip").Parse()
			if err != nil {
				t.Fatalf("playground example does not parse: %v\n---\n%s", err, src)
			}
			var diags []validator.Diagnostic
			diags = append(diags, validator.Validate(w).Diagnostics...)
			diags = append(diags, validator.Lint(w).Diagnostics...)
			for _, d := range diags {
				t.Errorf("playground example not clean: %s[%s] %s (fix the EXAMPLES entry in %s)",
					d.Severity, d.Code, d.Message, playgroundHTML)
			}
		})
	}
}

// extractPlaygroundExamples pulls every backtick-delimited source that begins
// with "workflow" out of the playground's `const EXAMPLES = { ... }` object.
func extractPlaygroundExamples(t *testing.T, doc string) []string {
	t.Helper()
	i := strings.Index(doc, "const EXAMPLES")
	if i < 0 {
		t.Fatal("playground.html: no `const EXAMPLES` object")
	}
	body := doc[i:]
	var out []string
	for {
		a := strings.IndexByte(body, '`')
		if a < 0 {
			break
		}
		rel := strings.IndexByte(body[a+1:], '`')
		if rel < 0 {
			break
		}
		content := body[a+1 : a+1+rel]
		if strings.HasPrefix(strings.TrimSpace(content), "workflow ") {
			out = append(out, content)
		}
		body = body[a+1+rel+1:]
	}
	return out
}

// TestPackExamples round-trips every example .dip through dipx.Pack →
// dipx.Open, asserting that each example bundles cleanly and reopens
// without error. This catches packer regressions and bundle-shape drift.
func TestPackExamples(t *testing.T) {
	matches, err := filepath.Glob("../examples/*.dip")
	if err != nil {
		t.Fatal(err)
	}
	subdirMatches, err := filepath.Glob("../examples/*/*.dip")
	if err != nil {
		t.Fatal(err)
	}
	matches = append(matches, subdirMatches...)
	if len(matches) == 0 {
		t.Fatal("no .dip files found in examples/")
	}

	for _, path := range matches {
		path := path
		t.Run(filepath.Base(path), func(t *testing.T) {
			var buf bytes.Buffer
			if _, err := dipx.Pack(context.Background(), path, &buf, dipx.PackOptions{}); err != nil {
				t.Fatalf("Pack failed: %v", err)
			}
			bundlePath := filepath.Join(t.TempDir(), "bundle.dipx")
			if err := os.WriteFile(bundlePath, buf.Bytes(), 0o644); err != nil {
				t.Fatalf("write bundle: %v", err)
			}
			if _, err := dipx.Open(context.Background(), bundlePath); err != nil {
				t.Fatalf("Open failed: %v", err)
			}
		})
	}
}
