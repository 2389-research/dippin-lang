package validator_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/2389-research/dippin-lang/dipx"
	"github.com/2389-research/dippin-lang/parser"
	"github.com/2389-research/dippin-lang/validator"
)

// TestLintExamples parses every .dip file in examples/ through the real
// parser and lints it, asserting zero DIP108 (unknown model) warnings and zero
// DIP147 (chain-attack topology). This catches model catalog staleness and
// invalid model IDs, and guards against an example silently demonstrating the
// restricted->tool-bearing laundering topology.
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
				if d.Code == validator.DIP108 || d.Code == validator.DIP147 {
					t.Errorf("%s: %s", name, d.Message)
				}
			}
		})
	}
}

// TestPackExamples round-trips every example .dip through dipx.Pack →
// dipx.OpenReader, asserting that each example bundles cleanly and reopens
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
			if _, err := dipx.Pack(context.Background(), path, &buf); err != nil {
				t.Fatalf("Pack failed: %v", err)
			}
			if _, err := dipx.OpenReader(context.Background(), bytes.NewReader(buf.Bytes()), int64(buf.Len())); err != nil {
				t.Fatalf("Open failed: %v", err)
			}
		})
	}
}
