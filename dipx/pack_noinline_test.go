package dipx

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// dipWithCommandFile is a minimal workflow referencing a script via
// command_file:, used across the no-inline pack tests.
const dipWithCommandFile = `workflow A
  goal: "no-inline test"
  start: Run
  exit: Done

  tool Run
    timeout: 30s
    command_file: scripts/run.sh

  agent Done
    prompt:
      Done.

  edges
    Run -> Done
`

// writeTree writes files (path->contents) under dir, creating parent dirs.
func writeTree(t *testing.T, dir string, files map[string]string) {
	t.Helper()
	for rel, body := range files {
		full := filepath.Join(dir, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

// packBytes packs entryPath with opts and returns the raw bundle bytes.
func packBytes(t *testing.T, entryPath string, opts PackOptions) []byte {
	t.Helper()
	var buf bytes.Buffer
	if _, err := Pack(context.Background(), entryPath, &buf, opts); err != nil {
		t.Fatalf("Pack: %v", err)
	}
	return buf.Bytes()
}

// manifestPaths returns the set of files[] paths in a packed bundle.
func manifestPaths(m Manifest) map[string]struct{} {
	out := make(map[string]struct{}, len(m.Files))
	for _, e := range m.Files {
		out[e.Path] = struct{}{}
	}
	return out
}

// openBundle opens raw bundle bytes via the internal reader path.
func openBundle(t *testing.T, raw []byte) *Bundle {
	t.Helper()
	b, err := openFromReader(context.Background(), bytes.NewReader(raw), int64(len(raw)))
	if err != nil {
		t.Fatalf("openFromReader: %v", err)
	}
	return b
}

func TestPack_NoInlineRoundTrip(t *testing.T) {
	dir := t.TempDir()
	writeTree(t, dir, map[string]string{
		"a.dip":          dipWithCommandFile,
		"scripts/run.sh": "#!/bin/sh\necho hello-from-script\n",
	})
	raw := packBytes(t, filepath.Join(dir, "a.dip"), PackOptions{NoInline: true})

	b := openBundle(t, raw)
	m := b.Manifest()
	if m.FormatVersion != 2 {
		t.Fatalf("FormatVersion = %d, want 2", m.FormatVersion)
	}
	paths := manifestPaths(m)
	if _, ok := paths["workflows/a.dip"]; !ok {
		t.Errorf("manifest missing workflows/a.dip; files=%v", m.Files)
	}
	if _, ok := paths["workflows/scripts/run.sh"]; !ok {
		t.Errorf("manifest missing workflows/scripts/run.sh; files=%v", m.Files)
	}
	if len(m.Files) != 2 {
		t.Errorf("len(Files) = %d, want 2: %v", len(m.Files), m.Files)
	}
	// The bundled .dip must RETAIN the command_file: directive (not inlined).
	dipText := string(b.fileBytes["workflows/a.dip"])
	if !strings.Contains(dipText, "command_file:") {
		t.Errorf("bundled .dip lost command_file: directive:\n%s", dipText)
	}
	if strings.Contains(dipText, "hello-from-script") {
		t.Errorf("bundled .dip unexpectedly inlined script body:\n%s", dipText)
	}
	// Asset bytes are present and hash-verified (open succeeded).
	if got := string(b.fileBytes["workflows/scripts/run.sh"]); !strings.Contains(got, "hello-from-script") {
		t.Errorf("asset bytes wrong: %q", got)
	}
}

func TestPack_V1DefaultUnchanged(t *testing.T) {
	dir := t.TempDir()
	writeTree(t, dir, map[string]string{"a.dip": minimalStandaloneDip()})
	entry := filepath.Join(dir, "a.dip")
	first := packBytes(t, entry, PackOptions{})
	second := packBytes(t, entry, PackOptions{})
	if !bytes.Equal(first, second) {
		t.Fatal("zero-opts Pack is not byte-deterministic")
	}
	b := openBundle(t, first)
	if b.Manifest().FormatVersion != 1 {
		t.Fatalf("FormatVersion = %d, want 1", b.Manifest().FormatVersion)
	}
}

func TestPack_NoInlinePromptFiles(t *testing.T) {
	dir := t.TempDir()
	writeTree(t, dir, map[string]string{
		"a.dip": `workflow A
  goal: "prompt-file no-inline"
  start: R
  exit: R

  agent R
    model: claude-sonnet-4-6
    system_prompt_file: prompts/persona.md
    prompt_file: prompts/task.md
`,
		"prompts/persona.md": "PERSONA\n",
		"prompts/task.md":    "TASK\n",
	})
	raw := packBytes(t, filepath.Join(dir, "a.dip"), PackOptions{NoInline: true})
	m := openBundle(t, raw).Manifest()
	paths := manifestPaths(m)
	for _, want := range []string{"workflows/prompts/persona.md", "workflows/prompts/task.md"} {
		if _, ok := paths[want]; !ok {
			t.Errorf("manifest missing %s; files=%v", want, m.Files)
		}
	}
}

func TestPack_IncludeDir(t *testing.T) {
	dir := t.TempDir()
	writeTree(t, dir, map[string]string{
		"a.dip":                 minimalStandaloneDip(),
		"scripts/lib/boot.sh":   "boot\n",
		"scripts/lib/helper.sh": "help\n",
	})
	raw := packBytes(t, filepath.Join(dir, "a.dip"), PackOptions{NoInline: true, Include: []string{"scripts/lib"}})
	paths := manifestPaths(openBundle(t, raw).Manifest())
	for _, want := range []string{"workflows/scripts/lib/boot.sh", "workflows/scripts/lib/helper.sh"} {
		if _, ok := paths[want]; !ok {
			t.Errorf("manifest missing %s; paths=%v", want, paths)
		}
	}
}

func TestPack_IncludeFile(t *testing.T) {
	dir := t.TempDir()
	writeTree(t, dir, map[string]string{
		"a.dip":        minimalStandaloneDip(),
		"scripts/x.sh": "x\n",
	})
	raw := packBytes(t, filepath.Join(dir, "a.dip"), PackOptions{NoInline: true, Include: []string{"scripts/x.sh"}})
	m := openBundle(t, raw).Manifest()
	if m.FormatVersion != 2 {
		t.Fatalf("FormatVersion = %d, want 2", m.FormatVersion)
	}
	if len(m.Files) != 2 {
		t.Fatalf("len(Files) = %d, want 2: %v", len(m.Files), m.Files)
	}
}

func TestPack_Dedup(t *testing.T) {
	dir := t.TempDir()
	writeTree(t, dir, map[string]string{
		"a.dip":          dipWithCommandFile,
		"scripts/run.sh": "run\n",
	})
	// scripts/run.sh referenced both by command_file: and --include.
	raw := packBytes(t, filepath.Join(dir, "a.dip"), PackOptions{NoInline: true, Include: []string{"scripts/run.sh"}})
	m := openBundle(t, raw).Manifest()
	count := 0
	for _, e := range m.Files {
		if e.Path == "workflows/scripts/run.sh" {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("scripts/run.sh appears %d times, want 1: %v", count, m.Files)
	}
}

func TestPack_NoInlineReproducible(t *testing.T) {
	dir := t.TempDir()
	writeTree(t, dir, map[string]string{
		"a.dip":          dipWithCommandFile,
		"scripts/run.sh": "run\n",
		"lib/b.sh":       "b\n",
		"lib/a.sh":       "a\n",
	})
	entry := filepath.Join(dir, "a.dip")
	opts := PackOptions{NoInline: true, Include: []string{"lib"}}
	if !bytes.Equal(packBytes(t, entry, opts), packBytes(t, entry, opts)) {
		t.Fatal("no-inline Pack is not byte-deterministic")
	}
}

func TestPack_IncludeDipFileRejected(t *testing.T) {
	dir := t.TempDir()
	writeTree(t, dir, map[string]string{
		"a.dip":      minimalStandaloneDip(),
		"helper.dip": minimalStandaloneDip(),
	})
	var buf bytes.Buffer
	_, err := Pack(context.Background(), filepath.Join(dir, "a.dip"), &buf, PackOptions{NoInline: true, Include: []string{"helper.dip"}})
	if !errors.Is(err, ErrPathUnsafe) {
		t.Fatalf("err = %v, want ErrPathUnsafe", err)
	}
}

func TestPack_IncludeEscapeRejected(t *testing.T) {
	dir := t.TempDir()
	writeTree(t, dir, map[string]string{"sub/a.dip": minimalStandaloneDip()})
	writeTree(t, dir, map[string]string{"outside.sh": "x\n"})
	var buf bytes.Buffer
	_, err := Pack(context.Background(), filepath.Join(dir, "sub", "a.dip"), &buf, PackOptions{NoInline: true, Include: []string{"../outside.sh"}})
	if !errors.Is(err, ErrPathUnsafe) {
		t.Fatalf("err = %v, want ErrPathUnsafe", err)
	}
}

func TestPack_IncludeEmptyRejected(t *testing.T) {
	dir := t.TempDir()
	writeTree(t, dir, map[string]string{"a.dip": minimalStandaloneDip()})
	var buf bytes.Buffer
	_, err := Pack(context.Background(), filepath.Join(dir, "a.dip"), &buf, PackOptions{NoInline: true, Include: []string{""}})
	if !errors.Is(err, ErrPathUnsafe) {
		t.Fatalf("err = %v, want ErrPathUnsafe", err)
	}
}

func TestPack_IncludeSymlinkRejected(t *testing.T) {
	dir := t.TempDir()
	writeTree(t, dir, map[string]string{
		"a.dip":       minimalStandaloneDip(),
		"real/tgt.sh": "x\n",
	})
	link := filepath.Join(dir, "link.sh")
	if err := os.Symlink(filepath.Join(dir, "real", "tgt.sh"), link); err != nil {
		t.Skipf("symlink unsupported: %v", err)
	}
	var buf bytes.Buffer
	_, err := Pack(context.Background(), filepath.Join(dir, "a.dip"), &buf, PackOptions{NoInline: true, Include: []string{"link.sh"}})
	if !errors.Is(err, ErrPathUnsafe) {
		t.Fatalf("err = %v, want ErrPathUnsafe", err)
	}
}

func TestPack_IncludeDirWithDipRejected(t *testing.T) {
	dir := t.TempDir()
	writeTree(t, dir, map[string]string{
		"a.dip":            minimalStandaloneDip(),
		"tools/helper.dip": minimalStandaloneDip(),
	})
	var buf bytes.Buffer
	_, err := Pack(context.Background(), filepath.Join(dir, "a.dip"), &buf, PackOptions{NoInline: true, Include: []string{"tools"}})
	if !errors.Is(err, ErrPathUnsafe) {
		t.Fatalf("err = %v, want ErrPathUnsafe", err)
	}
}

func TestCheckPackCaps_FileCount(t *testing.T) {
	all := make([]packedFile, maxFiles+1)
	if err := checkPackCaps(all); !errors.Is(err, ErrCapExceeded) {
		t.Fatalf("err = %v, want ErrCapExceeded", err)
	}
}

func TestCheckPackCaps_TotalBytes(t *testing.T) {
	big := make([]byte, maxPerFileBytes)
	all := make([]packedFile, 0, 3)
	for i := 0; i < 3; i++ { // 3 * 50MB = 150MB > 100MB total cap
		all = append(all, packedFile{bytes: big})
	}
	if err := checkPackCaps(all); !errors.Is(err, ErrCapExceeded) {
		t.Fatalf("err = %v, want ErrCapExceeded", err)
	}
}

func TestBundleAssetPathFor(t *testing.T) {
	root := t.TempDir()
	got, err := bundlePathFor(filepath.Join(root, "scripts", "run.sh"), root)
	if err != nil {
		t.Fatal(err)
	}
	if got != "workflows/scripts/run.sh" {
		t.Fatalf("got %q, want workflows/scripts/run.sh", got)
	}
}

func TestEnsureAssetUnderRoot_EscapeRejected(t *testing.T) {
	root := t.TempDir()
	if err := ensureAssetUnderRoot(filepath.Dir(root)+"/evil", root); !errors.Is(err, ErrPathUnsafe) {
		t.Fatalf("err = %v, want ErrPathUnsafe", err)
	}
}

func TestParseAllWorkflows_SkipsAssets(t *testing.T) {
	verified := map[string]verifiedBytes{
		"workflows/entry.dip":      newVerifiedBytes([]byte(minimalStandaloneDip())),
		"workflows/scripts/run.sh": newVerifiedBytes([]byte("#!/bin/sh\nnot a workflow\n")),
	}
	out, err := parseAllWorkflows(verified, "workflows/entry.dip")
	if err != nil {
		t.Fatalf("parseAllWorkflows: %v", err)
	}
	if _, ok := out["workflows/entry.dip"]; !ok {
		t.Error("missing workflow entry")
	}
	if _, ok := out["workflows/scripts/run.sh"]; ok {
		t.Error("asset was parsed as a workflow")
	}
}

// TestPack_NoInlineExtractRoundTrip is the structural acceptance test: a source
// tree whose command_file: script sources a sibling via --include, packed with
// --no-inline, extracted, must be byte-identical under extracted/workflows/<rel>
// with directives preserved.
func TestPack_NoInlineExtractRoundTrip(t *testing.T) {
	src := t.TempDir()
	tree := map[string]string{
		"dev_loop.dip":             dipWithCommandFile,
		"scripts/run.sh":           "#!/bin/sh\n. \"${graph.workflow_dir}/scripts/lib/bootstrap.sh\"\n",
		"scripts/lib/bootstrap.sh": "#!/bin/sh\necho bootstrapped\n",
	}
	writeTree(t, src, tree)
	// Point command_file: at scripts/run.sh (dipWithCommandFile already does).
	raw := packBytes(t, filepath.Join(src, "dev_loop.dip"), PackOptions{NoInline: true, Include: []string{"scripts/lib"}})

	bundleFile := filepath.Join(t.TempDir(), "out.dipx")
	if err := os.WriteFile(bundleFile, raw, 0o644); err != nil {
		t.Fatal(err)
	}
	dest := filepath.Join(t.TempDir(), "extracted")
	if err := Extract(context.Background(), bundleFile, dest, false); err != nil {
		t.Fatalf("Extract: %v", err)
	}
	for rel, want := range tree {
		got, err := os.ReadFile(filepath.Join(dest, "workflows", filepath.FromSlash(rel)))
		if err != nil {
			t.Fatalf("read extracted %s: %v", rel, err)
		}
		if rel == "dev_loop.dip" {
			if !strings.Contains(string(got), "command_file:") {
				t.Errorf("extracted entry lost command_file: directive")
			}
			continue // .dip is reformatted; assets are byte-identical
		}
		if string(got) != want {
			t.Errorf("extracted %s not byte-identical:\n got %q\nwant %q", rel, got, want)
		}
	}
}
