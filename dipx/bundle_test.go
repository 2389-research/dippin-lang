package dipx

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/2389-research/dippin-lang/ir"
)

func newTestBundle(t *testing.T) *Bundle {
	t.Helper()
	hash := strings.Repeat("a", 64)
	m := Manifest{
		FormatVersion: 1,
		Entry:         "workflows/a.dip",
		Files: []ManifestEntry{
			{Path: "workflows/a.dip", SHA256: hash},
			{Path: "workflows/b.dip", SHA256: hash},
		},
	}
	wfA := &ir.Workflow{Name: "A", Start: "S", Exit: "E"}
	wfB := &ir.Workflow{Name: "B", Start: "S", Exit: "E"}
	return &Bundle{
		manifest:      m,
		workflows:     map[string]*ir.Workflow{"workflows/a.dip": wfA, "workflows/b.dip": wfB},
		fileBytes:     map[string][]byte{"workflows/a.dip": []byte("a-bytes"), "workflows/b.dip": []byte("b-bytes")},
		manifestBytes: []byte(`{"format_version":1}`),
	}
}

func TestBundle_Manifest_ReturnsCopy(t *testing.T) {
	b := newTestBundle(t)
	m1 := b.Manifest()
	m1.Entry = "MUTATED"
	if m1.Entry != "MUTATED" {
		t.Fatal("local mutation lost")
	}
	// Mutating files[] should also not affect the bundle.
	m1.Files[0].Path = "MUTATED-PATH"
	m2 := b.Manifest()
	if m2.Entry == "MUTATED" {
		t.Fatal("Manifest() should return a copy of Entry")
	}
	if m2.Files[0].Path == "MUTATED-PATH" {
		t.Fatal("Manifest() should return a copy of Files")
	}
}

func TestBundle_Identity(t *testing.T) {
	b := newTestBundle(t)
	id := b.Identity()
	if id == [32]byte{} {
		t.Fatal("Identity should be non-zero")
	}
}

func TestBundle_Workflow(t *testing.T) {
	b := newTestBundle(t)
	wf, err := b.Workflow(context.Background(), "b.dip", "workflows/a.dip")
	if err != nil {
		t.Fatal(err)
	}
	if wf.Name != "B" {
		t.Fatalf("Name = %q", wf.Name)
	}
}

func TestBundle_Workflow_RefMissing(t *testing.T) {
	b := newTestBundle(t)
	_, err := b.Workflow(context.Background(), "missing.dip", "workflows/a.dip")
	if !errors.Is(err, ErrFileMissing) {
		t.Fatalf("err = %v, want ErrFileMissing", err)
	}
}

func TestBundle_Resolve(t *testing.T) {
	b := newTestBundle(t)
	got, err := b.Resolve("b.dip", "workflows/a.dip")
	if err != nil {
		t.Fatal(err)
	}
	if got != "workflows/b.dip" {
		t.Fatalf("got %q", got)
	}
}

func TestBundle_ByteTotal_SumsFileBytes(t *testing.T) {
	b := &Bundle{
		fileBytes: map[string][]byte{
			"workflows/a.dip": []byte("hello"),   // 5
			"workflows/b.dip": []byte("world!!"), // 7
			"workflows/c.dip": []byte(""),        // 0
		},
	}
	if got, want := b.ByteTotal(), int64(12); got != want {
		t.Errorf("ByteTotal() = %d, want %d", got, want)
	}
}

func TestBundle_ByteTotal_EmptyMap(t *testing.T) {
	b := &Bundle{fileBytes: map[string][]byte{}}
	if got, want := b.ByteTotal(), int64(0); got != want {
		t.Errorf("ByteTotal() = %d, want %d", got, want)
	}
}

func TestBundle_ByteTotal_NilMap(t *testing.T) {
	b := &Bundle{}
	if got, want := b.ByteTotal(), int64(0); got != want {
		t.Errorf("ByteTotal() = %d, want %d", got, want)
	}
}
