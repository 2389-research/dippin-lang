package dipx

import (
	"errors"
	"strings"
	"testing"
)

func TestCanonicalize_Valid(t *testing.T) {
	cases := []string{
		"workflows/foo.dip",
		"workflows/sub/bar.dip",
		"workflows/a/b/c/d/e/f/g/h/i/j/k/l/m/n/o.dip", // 16 components, just at cap
	}
	for _, in := range cases {
		got, err := Canonicalize(in)
		if err != nil {
			t.Errorf("Canonicalize(%q): unexpected error: %v", in, err)
			continue
		}
		if got != in {
			t.Errorf("Canonicalize(%q) = %q, want unchanged", in, got)
		}
	}
}

func TestCanonicalize_Rejects(t *testing.T) {
	cases := []struct {
		name string
		in   string
	}{
		{"absolute", "/workflows/foo.dip"},
		{"backslash", "workflows\\foo.dip"},
		{"dot-dot", "workflows/../etc/passwd"},
		{"leading-dot", "./workflows/foo.dip"},
		{"empty-component", "workflows//foo.dip"},
		{"nul", "workflows/foo\x00.dip"},
		{"control", "workflows/foo\x01.dip"},
		{"del", "workflows/foo\x7f.dip"},
		{"trailing-space", "workflows/foo .dip"},
		{"leading-space", "workflows/ foo.dip"},
		{"trailing-dot", "workflows/foo.dip.."},
		{"win-reserved-con", "workflows/CON.dip"},
		{"win-reserved-com1", "workflows/COM1.dip"},
		{"win-reserved-con-multi-ext", "workflows/CON.tar.dip"},
		{"win-reserved-com1-multi-ext", "workflows/COM1.foo.dip"},
		{"win-reserved-aux-with-prefix", "workflows/AUX.something.dip"},
		{"missing-extension", "workflows/foo"},
		{"wrong-extension", "workflows/foo.txt"},
		{"uppercase-extension", "workflows/foo.DIP"},
		{"too-many-components", "workflows/a/b/c/d/e/f/g/h/i/j/k/l/m/n/o/p.dip"}, // 17
		{"too-long", "workflows/" + strings.Repeat("a", 1020) + ".dip"},
		{"not-under-workflows", "other/foo.dip"},
		{"empty", ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := Canonicalize(c.in)
			if err == nil {
				t.Fatalf("Canonicalize(%q) succeeded, expected error", c.in)
			}
			if !errors.Is(err, ErrPathUnsafe) {
				t.Fatalf("error = %v, want ErrPathUnsafe", err)
			}
		})
	}
}

func TestCanonicalize_ErrorContext(t *testing.T) {
	_, err := Canonicalize("workflows/CON.dip")
	if err == nil {
		t.Fatal("expected error")
	}
	var be *BundleError
	if !errors.As(err, &be) {
		t.Fatalf("expected *BundleError, got %T", err)
	}
	if be.Path != "workflows/CON.dip" {
		t.Errorf("Path = %q, want full path", be.Path)
	}
	if be.Detail == "" {
		t.Errorf("Detail empty; expected mention of the offending component")
	}
}

func TestCanonicalize_RejectsNFD(t *testing.T) {
	// é as NFD: e + U+0301 (combining acute)
	in := "workflows/café.dip"
	_, err := Canonicalize(in)
	if err == nil {
		t.Fatal("expected error for NFD-encoded path")
	}
	if !errors.Is(err, ErrPathUnsafe) {
		t.Fatalf("err = %v, want ErrPathUnsafe", err)
	}
}

func TestResolve_Sibling(t *testing.T) {
	got, err := resolveLexically("foo.dip", "workflows/parent.dip")
	if err != nil {
		t.Fatal(err)
	}
	if got != "workflows/foo.dip" {
		t.Errorf("got %q, want workflows/foo.dip", got)
	}
}

func TestResolve_Subdir(t *testing.T) {
	got, err := resolveLexically("phases/code_review.dip", "workflows/parent.dip")
	if err != nil {
		t.Fatal(err)
	}
	if got != "workflows/phases/code_review.dip" {
		t.Errorf("got %q, want workflows/phases/code_review.dip", got)
	}
}

func TestResolve_DotDotInRefAllowed(t *testing.T) {
	// .. in ref is OK as long as resolved path stays in workflows/
	got, err := resolveLexically("../sibling/foo.dip", "workflows/sub/parent.dip")
	if err != nil {
		t.Fatal(err)
	}
	if got != "workflows/sibling/foo.dip" {
		t.Errorf("got %q, want workflows/sibling/foo.dip", got)
	}
}

func TestResolve_DotDotEscapeRejected(t *testing.T) {
	_, err := resolveLexically("../../etc/passwd", "workflows/parent.dip")
	if err == nil {
		t.Fatal("expected error escaping workflows/")
	}
	if !errors.Is(err, ErrPathUnsafe) {
		t.Fatalf("err = %v, want ErrPathUnsafe", err)
	}
}

func TestDetectCycle_Acyclic(t *testing.T) {
	graph := map[string][]string{
		"a": {"b", "c"},
		"b": {"d"},
		"c": {"d"},
		"d": {},
	}
	if err := detectCycles(graph, "a"); err != nil {
		t.Fatalf("expected acyclic, got %v", err)
	}
}

func TestDetectCycle_SelfLoop(t *testing.T) {
	graph := map[string][]string{"a": {"a"}}
	err := detectCycles(graph, "a")
	if !errors.Is(err, ErrRefCycle) {
		t.Fatalf("err = %v, want ErrRefCycle", err)
	}
}

func TestDetectCycle_ThreeCycle(t *testing.T) {
	graph := map[string][]string{
		"a": {"b"},
		"b": {"c"},
		"c": {"a"},
	}
	err := detectCycles(graph, "a")
	if !errors.Is(err, ErrRefCycle) {
		t.Fatalf("err = %v, want ErrRefCycle", err)
	}
}

func TestDetectCycle_DepthCap(t *testing.T) {
	// Linear chain a0 -> a1 -> ... -> a65
	graph := map[string][]string{}
	for i := 0; i <= 65; i++ {
		next := []string{}
		if i < 65 {
			next = []string{key(i + 1)}
		}
		graph[key(i)] = next
	}
	err := detectCycles(graph, key(0))
	if !errors.Is(err, ErrCapExceeded) {
		t.Fatalf("err = %v, want ErrCapExceeded", err)
	}
}

func key(i int) string { return "node" + string(rune('0'+i%10)) + string(rune('0'+i/10)) }

func TestDetectCycle_FullCyclePath(t *testing.T) {
	graph := map[string][]string{
		"a": {"b"},
		"b": {"c"},
		"c": {"a"},
	}
	err := detectCycles(graph, "a")
	if err == nil {
		t.Fatal("expected ErrRefCycle")
	}
	var be *BundleError
	if !errors.As(err, &be) {
		t.Fatalf("expected *BundleError, got %T", err)
	}
	if be.Path != "a" {
		t.Errorf("Path = %q, want a (cycle entry node)", be.Path)
	}
	want := "a -> b -> c -> a"
	if be.Detail != want {
		t.Errorf("Detail = %q, want %q", be.Detail, want)
	}
}

func TestDetectCycle_AtCapDepthSucceeds(t *testing.T) {
	// Linear chain a0 -> a1 -> ... -> a64 (depth 64, exactly at cap).
	graph := map[string][]string{}
	for i := 0; i <= 64; i++ {
		next := []string{}
		if i < 64 {
			next = []string{key(i + 1)}
		}
		graph[key(i)] = next
	}
	if err := detectCycles(graph, key(0)); err != nil {
		t.Fatalf("expected at-cap chain to succeed, got %v", err)
	}
}

func TestDetectCycle_EmptyGraph(t *testing.T) {
	if err := detectCycles(map[string][]string{}, "any"); err != nil {
		t.Fatalf("expected empty graph to succeed, got %v", err)
	}
}

func TestResolveCanonicalize_AdversarialComposition(t *testing.T) {
	// All of these should fail: the resolved path either escapes workflows/,
	// has empty components after cleaning, or violates suffix/component rules.
	cases := []struct {
		name       string
		refPath    string
		relativeTo string
	}{
		{"escape-via-many-dotdots", "../../../etc/passwd.dip", "workflows/sub/parent.dip"},
		{"backslash-in-ref", "sub\\foo.dip", "workflows/parent.dip"},
		{"ref-cleans-to-just-workflows", "..", "workflows/parent.dip"},
		{"ref-with-nul", "foo\x00.dip", "workflows/parent.dip"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := resolveLexically(c.refPath, c.relativeTo)
			if err == nil {
				t.Fatalf("expected error")
			}
			if !errors.Is(err, ErrPathUnsafe) {
				t.Fatalf("err = %v, want ErrPathUnsafe", err)
			}
		})
	}
}
