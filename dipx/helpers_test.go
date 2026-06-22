package dipx

import (
	"archive/zip"
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/2389-research/dippin-lang/ir"
)

func TestReadManifestEntry_Happy(t *testing.T) {
	src := `{"format_version":1,"entry":"workflows/a.dip","files":[{"path":"workflows/a.dip","sha256":"` + strings.Repeat("a", 64) + `"}]}`
	cz := buildSingleEntryZip(t, "manifest.json", []byte(src))
	raw, err := readManifestEntry(cz)
	if err != nil {
		t.Fatal(err)
	}
	if string(raw) != src {
		t.Fatalf("got %q", raw)
	}
}

func TestReadManifestEntry_Missing(t *testing.T) {
	cz := buildSingleEntryZip(t, "workflows/a.dip", []byte("x"))
	_, err := readManifestEntry(cz)
	if !errors.Is(err, ErrManifestMissing) {
		t.Fatalf("err = %v, want ErrManifestMissing", err)
	}
}

func TestReadManifestEntry_OversizedRejected(t *testing.T) {
	big := bytes.Repeat([]byte("a"), maxManifestSize+10)
	cz := buildSingleEntryZip(t, "manifest.json", big)
	_, err := readManifestEntry(cz)
	if !errors.Is(err, ErrManifestInvalid) {
		t.Fatalf("err = %v, want ErrManifestInvalid", err)
	}
}

func TestVerifyAllHashes_Happy(t *testing.T) {
	contentA := []byte("a")
	contentB := []byte("b")
	manifest := Manifest{
		FormatVersion: 1,
		Entry:         "workflows/a.dip",
		Files: []ManifestEntry{
			{Path: "workflows/a.dip", SHA256: hashOf(contentA)},
			{Path: "workflows/b.dip", SHA256: hashOf(contentB)},
		},
	}
	cz := buildMultiEntryZip(t, map[string][]byte{
		"workflows/a.dip": contentA,
		"workflows/b.dip": contentB,
		"manifest.json":   []byte("{}"),
	})
	verified, totalBytes, err := verifyAllHashesCtx(context.Background(), cz, manifest, 100<<20)
	if err != nil {
		t.Fatal(err)
	}
	if len(verified) != 2 {
		t.Fatalf("verified count = %d", len(verified))
	}
	if totalBytes != int64(len(contentA)+len(contentB)) {
		t.Fatalf("totalBytes = %d", totalBytes)
	}
}

func TestVerifyAllHashes_TotalCap(t *testing.T) {
	content := bytes.Repeat([]byte("a"), 10)
	manifest := Manifest{
		FormatVersion: 1,
		Entry:         "workflows/a.dip",
		Files: []ManifestEntry{
			{Path: "workflows/a.dip", SHA256: hashOf(content)},
			{Path: "workflows/b.dip", SHA256: hashOf(content)},
		},
	}
	cz := buildMultiEntryZip(t, map[string][]byte{
		"workflows/a.dip": content,
		"workflows/b.dip": content,
		"manifest.json":   []byte("{}"),
	})
	_, _, err := verifyAllHashesCtx(context.Background(), cz, manifest, 15) // total cap below sum
	if !errors.Is(err, ErrCapExceeded) {
		t.Fatalf("err = %v, want ErrCapExceeded", err)
	}
}

func TestParseAllWorkflows_Happy(t *testing.T) {
	src := `workflow A
  goal: x
  start: S
  exit: S
  agent S
    prompt: hi
`
	verified := map[string]verifiedBytes{
		"workflows/a.dip": newVerifiedBytes([]byte(src)),
	}
	parsed, err := parseAllWorkflows(verified, "workflows/a.dip")
	if err != nil {
		t.Fatal(err)
	}
	if parsed["workflows/a.dip"].Name != "A" {
		t.Fatalf("name = %q", parsed["workflows/a.dip"].Name)
	}
}

func TestParseAllWorkflows_EntryParseError(t *testing.T) {
	// "workflow" with no name token triggers parser diagnostics and a non-nil
	// error from Parse(). The parser is permissive about totally unrelated
	// junk like "garbage" (it just produces an empty workflow), so we use a
	// well-formed-prefix-but-wrong shape instead.
	verified := map[string]verifiedBytes{
		"workflows/a.dip": newVerifiedBytes([]byte("workflow")),
	}
	_, err := parseAllWorkflows(verified, "workflows/a.dip")
	if !errors.Is(err, ErrEntryParse) {
		t.Fatalf("err = %v, want ErrEntryParse", err)
	}
}

func TestWalkRefs_AcceptsValid(t *testing.T) {
	parent := &ir.Workflow{
		Name: "P", Start: "X", Exit: "X",
		Nodes: []*ir.Node{
			{ID: "X", Kind: ir.NodeSubgraph, Config: ir.SubgraphConfig{Ref: "child.dip"}},
		},
	}
	child := &ir.Workflow{Name: "C", Start: "Y", Exit: "Y"}
	parsed := map[string]*ir.Workflow{
		"workflows/parent.dip": parent,
		"workflows/child.dip":  child,
	}
	manifest := Manifest{Entry: "workflows/parent.dip", Files: []ManifestEntry{
		{Path: "workflows/parent.dip"}, {Path: "workflows/child.dip"},
	}}
	if err := walkRefs(parsed, manifest); err != nil {
		t.Fatal(err)
	}
}

func TestWalkRefs_RejectsEscape(t *testing.T) {
	parent := &ir.Workflow{
		Name: "P", Start: "X", Exit: "X",
		Nodes: []*ir.Node{
			{ID: "X", Kind: ir.NodeSubgraph, Config: ir.SubgraphConfig{Ref: "../escape.dip"}},
		},
	}
	parsed := map[string]*ir.Workflow{"workflows/parent.dip": parent}
	manifest := Manifest{Entry: "workflows/parent.dip", Files: []ManifestEntry{{Path: "workflows/parent.dip"}}}
	err := walkRefs(parsed, manifest)
	if !errors.Is(err, ErrPathUnsafe) && !errors.Is(err, ErrRefEscape) {
		t.Fatalf("err = %v, want ErrPathUnsafe or ErrRefEscape", err)
	}
}

// TestWalkRefs_DetectsCycleInUnreachableWorkflow asserts that detectCycles
// runs for every manifest-listed workflow, not just m.Entry. This catches
// the historical asymmetry where parseAllWorkflows parsed every workflow
// (so it would parse a cyclic-but-unreachable one) but detectCycles only
// rooted at m.Entry. After Task 4 + Task 8 of v1.1 Bundle 6, the cycle
// surfaces as ErrRefCycle.
func TestWalkRefs_DetectsCycleInUnreachableWorkflow(t *testing.T) {
	// Entry has no refs (no cycle reachable from entry).
	entry := &ir.Workflow{Name: "E", Start: "X", Exit: "X",
		Nodes: []*ir.Node{{ID: "X"}},
	}
	// Two manifest-listed workflows form a cycle, neither reachable from entry.
	a := &ir.Workflow{Name: "A", Start: "Y", Exit: "Y",
		Nodes: []*ir.Node{
			{ID: "Y", Kind: ir.NodeSubgraph, Config: ir.SubgraphConfig{Ref: "b.dip"}},
		},
	}
	b := &ir.Workflow{Name: "B", Start: "Z", Exit: "Z",
		Nodes: []*ir.Node{
			{ID: "Z", Kind: ir.NodeSubgraph, Config: ir.SubgraphConfig{Ref: "a.dip"}},
		},
	}
	parsed := map[string]*ir.Workflow{
		"workflows/entry.dip": entry,
		"workflows/a.dip":     a,
		"workflows/b.dip":     b,
	}
	manifest := Manifest{Entry: "workflows/entry.dip", Files: []ManifestEntry{
		{Path: "workflows/entry.dip"},
		{Path: "workflows/a.dip"},
		{Path: "workflows/b.dip"},
	}}
	err := walkRefs(parsed, manifest)
	if !errors.Is(err, ErrRefCycle) {
		t.Fatalf("walkRefs err = %v, want ErrRefCycle", err)
	}
}

func TestNormalizeConditions_PopulatesParsedAST(t *testing.T) {
	wf := &ir.Workflow{
		Name: "W", Start: "A", Exit: "B",
		Nodes: []*ir.Node{{ID: "A"}, {ID: "B"}},
		Edges: []*ir.Edge{
			{From: "A", To: "B", Condition: &ir.Condition{Raw: "ctx.x = y"}},
		},
	}
	parsed := map[string]*ir.Workflow{"workflows/w.dip": wf}
	if err := normalizeConditions(parsed); err != nil {
		t.Fatal(err)
	}
	if wf.Edges[0].Condition.Parsed == nil {
		t.Fatal("Parsed AST not populated")
	}
}

// TestVerifyAllHashesCtx_PreCancelledReturnsContextErr asserts that a
// pre-cancelled ctx causes verifyAllHashesCtx to return context.Canceled
// without doing the full pass. P10.10.
func TestVerifyAllHashesCtx_PreCancelledReturnsContextErr(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	m := Manifest{Files: []ManifestEntry{{Path: "workflows/a.dip"}}}
	_, _, err := verifyAllHashesCtx(ctx, nil, m, 100<<20)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("err = %v, want context.Canceled", err)
	}
}

func buildMultiEntryZip(t *testing.T, files map[string][]byte) *constrainedZip {
	t.Helper()
	var buf bytes.Buffer
	w := zip.NewWriter(&buf)
	for name, content := range files {
		// Use the writeUTF8Entry helper from zipio_test.go to set bit 11.
		writeUTF8Entry(t, w, name, content)
	}
	w.Close()
	cz, err := openConstrainedZip(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		t.Fatal(err)
	}
	return cz
}
