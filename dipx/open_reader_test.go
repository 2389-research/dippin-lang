package dipx

import (
	"bytes"
	"context"
	"path/filepath"
	"testing"
)

// TestOpenReader_InMemoryBundle covers #247: an embedded consumer can load a
// .dipx from bytes (no temp file) via the exported OpenReader, with the same
// strict admission as Open.
func TestOpenReader_InMemoryBundle(t *testing.T) {
	dir := t.TempDir()
	writeTree(t, dir, map[string]string{
		"a.dip": `workflow A
  goal: "in-memory bundle"
  start: R
  exit: R

  agent R
    model: claude-sonnet-4-6
    prompt: "hi"
`,
	})
	raw := packBytes(t, filepath.Join(dir, "a.dip"), PackOptions{})

	b, err := OpenReader(context.Background(), bytes.NewReader(raw), int64(len(raw)))
	if err != nil {
		t.Fatalf("OpenReader: %v", err)
	}
	if b == nil {
		t.Fatal("OpenReader returned a nil bundle")
	}
	if b.Manifest().Entry == "" {
		t.Fatal("OpenReader bundle has no entry workflow")
	}
}
