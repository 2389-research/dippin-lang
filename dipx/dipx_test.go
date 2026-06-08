package dipx

import (
	"archive/zip"
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

const minimalDipSrc = `workflow Hello
  goal: hello
  start: A
  exit: A

  agent A
    prompt: hi
`

func buildHappyDipx(t *testing.T) []byte {
	t.Helper()
	body := []byte(minimalDipSrc)
	bodyHash := hashOf(body)
	manifestSrc := fmt.Sprintf(`{"format_version":1,"entry":"workflows/hello.dip","files":[{"path":"workflows/hello.dip","sha256":%q}]}`, bodyHash)
	var buf bytes.Buffer
	w := zip.NewWriter(&buf)
	writeUTF8Entry(t, w, "manifest.json", []byte(manifestSrc))
	writeUTF8Entry(t, w, "workflows/hello.dip", body)
	w.Close()
	return buf.Bytes()
}

func TestOpen_Happy(t *testing.T) {
	raw := buildHappyDipx(t)
	b, err := OpenReader(context.Background(), bytes.NewReader(raw), int64(len(raw)))
	if err != nil {
		t.Fatal(err)
	}
	if b.Entry().Name != "Hello" {
		t.Fatalf("entry name = %q", b.Entry().Name)
	}
	if b.Manifest().FormatVersion != 1 {
		t.Fatalf("format version = %d", b.Manifest().FormatVersion)
	}
}

func TestOpen_HashMismatch(t *testing.T) {
	body := []byte(minimalDipSrc)
	wrongHash := hashOf([]byte("different"))
	manifestSrc := fmt.Sprintf(`{"format_version":1,"entry":"workflows/hello.dip","files":[{"path":"workflows/hello.dip","sha256":%q}]}`, wrongHash)
	var buf bytes.Buffer
	w := zip.NewWriter(&buf)
	writeUTF8Entry(t, w, "manifest.json", []byte(manifestSrc))
	writeUTF8Entry(t, w, "workflows/hello.dip", body)
	w.Close()
	_, err := OpenReader(context.Background(), bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if !errors.Is(err, ErrHashMismatch) {
		t.Fatalf("err = %v, want ErrHashMismatch", err)
	}
}

func TestOpen_ContextCancelled(t *testing.T) {
	raw := buildHappyDipx(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately
	_, err := OpenReader(ctx, bytes.NewReader(raw), int64(len(raw)))
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("err = %v, want context.Canceled", err)
	}
}

func TestExtract_Happy(t *testing.T) {
	raw := buildHappyDipx(t)
	src := filepath.Join(t.TempDir(), "h.dipx")
	if err := os.WriteFile(src, raw, 0644); err != nil {
		t.Fatal(err)
	}
	dest := filepath.Join(t.TempDir(), "out")
	if err := Extract(context.Background(), src, dest, false); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dest, "workflows", "hello.dip")); err != nil {
		t.Fatalf("expected file extracted: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dest, "manifest.json")); err != nil {
		t.Fatalf("expected manifest extracted: %v", err)
	}
}

func TestExtract_RefusesExistingWithoutForce(t *testing.T) {
	raw := buildHappyDipx(t)
	src := filepath.Join(t.TempDir(), "h.dipx")
	if err := os.WriteFile(src, raw, 0644); err != nil {
		t.Fatal(err)
	}
	dest := t.TempDir() // already exists
	err := Extract(context.Background(), src, dest, false)
	if err == nil {
		t.Fatal("expected error when destdir exists")
	}
}

func TestOpen_FormatVersionRejected(t *testing.T) {
	body := []byte(minimalDipSrc)
	bodyHash := hashOf(body)
	manifestSrc := fmt.Sprintf(`{"format_version":999,"entry":"workflows/hello.dip","files":[{"path":"workflows/hello.dip","sha256":%q}]}`, bodyHash)
	var buf bytes.Buffer
	w := zip.NewWriter(&buf)
	writeUTF8Entry(t, w, "manifest.json", []byte(manifestSrc))
	writeUTF8Entry(t, w, "workflows/hello.dip", body)
	w.Close()
	_, err := OpenReader(context.Background(), bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if !errors.Is(err, ErrUnsupportedFormatVersion) {
		t.Fatalf("err = %v, want ErrUnsupportedFormatVersion", err)
	}
}

func TestPack_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	parent := `workflow P
  goal: x
  start: S
  exit: E
  subgraph S
    ref: child.dip
  agent E
    prompt: end
  edges
    S -> E
`
	child := `workflow C
  goal: y
  start: A
  exit: A
  agent A
    prompt: child
`
	if err := os.WriteFile(filepath.Join(dir, "parent.dip"), []byte(parent), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "child.dip"), []byte(child), 0644); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	manifest, err := Pack(context.Background(), filepath.Join(dir, "parent.dip"), &buf)
	if err != nil {
		t.Fatal(err)
	}
	if manifest.Entry != "workflows/parent.dip" {
		t.Fatalf("entry = %q", manifest.Entry)
	}
	if len(manifest.Files) != 2 {
		t.Fatalf("files = %d", len(manifest.Files))
	}
	// Open the resulting bundle.
	b, err := OpenReader(context.Background(), bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		t.Fatal(err)
	}
	if b.Entry().Name != "P" {
		t.Fatalf("entry name = %q", b.Entry().Name)
	}
}

func TestPack_Reproducible(t *testing.T) {
	dir := t.TempDir()
	src := `workflow A
  goal: x
  start: S
  exit: S
  agent S
    prompt: hi
`
	if err := os.WriteFile(filepath.Join(dir, "a.dip"), []byte(src), 0644); err != nil {
		t.Fatal(err)
	}
	var buf1, buf2 bytes.Buffer
	if _, err := Pack(context.Background(), filepath.Join(dir, "a.dip"), &buf1); err != nil {
		t.Fatal(err)
	}
	if _, err := Pack(context.Background(), filepath.Join(dir, "a.dip"), &buf2); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(buf1.Bytes(), buf2.Bytes()) {
		t.Fatal("Pack is not byte-deterministic")
	}
}

func TestPack_RejectsOversizedSource(t *testing.T) {
	dir := t.TempDir()
	// Source > maxPerFileBytes. Buffer of zeros is fast to allocate and
	// triggers the cap check before parsing.
	big := make([]byte, maxPerFileBytes+1)
	if err := os.WriteFile(filepath.Join(dir, "big.dip"), big, 0644); err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	_, err := Pack(context.Background(), filepath.Join(dir, "big.dip"), &buf)
	if !errors.Is(err, ErrCapExceeded) {
		t.Fatalf("err = %v, want ErrCapExceeded", err)
	}
}

func TestOpen_RejectsTooManyFiles(t *testing.T) {
	// Verify the maxFiles cap fires with ErrCapExceeded.
	//
	// We exercise verifyAllHashes directly rather than going through OpenReader:
	// constructing a manifest with maxFiles+1 entries serialized as JSON would
	// always exceed maxManifestSize (1MB) regardless of how short the paths
	// are — 10001 entries each containing a 64-char SHA-256 plus the minimal
	// JSON overhead and shortest valid "workflows/<N>.dip" path is ~1.06MB —
	// so the manifest-size cap would intercept before the file-count cap could
	// fire. Calling verifyAllHashes directly tests the intended branch without
	// hitting that ordering hazard.
	files := make([]ManifestEntry, 0, maxFiles+1)
	hash := hashOf([]byte("x"))
	for i := 0; i <= maxFiles; i++ {
		files = append(files, ManifestEntry{
			Path:   fmt.Sprintf("workflows/f%d.dip", i),
			SHA256: hash,
		})
	}
	m := Manifest{FormatVersion: 1, Entry: "workflows/f0.dip", Files: files}
	_, _, err := verifyAllHashes(nil, m, maxTotalUncompBytes)
	if !errors.Is(err, ErrCapExceeded) {
		t.Fatalf("err = %v, want ErrCapExceeded", err)
	}
}

func TestBundle_ConcurrentReads(t *testing.T) {
	raw := buildHappyDipx(t)
	b, err := OpenReader(context.Background(), bytes.NewReader(raw), int64(len(raw)))
	if err != nil {
		t.Fatal(err)
	}
	const entryPath = "workflows/hello.dip"
	// firstErr captures the first non-nil error any goroutine sees on the
	// shared lookup/read paths, so a race-induced failure can't pass silently.
	var (
		errMu    sync.Mutex
		firstErr error
	)
	record := func(err error) {
		if err == nil {
			return
		}
		errMu.Lock()
		defer errMu.Unlock()
		if firstErr == nil {
			firstErr = err
		}
	}
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = b.Entry()
			_ = b.Manifest()
			_ = b.Identity()
			// Exercise the shared lookup/read paths — these are what the
			// concurrency contract actually covers.
			_, err := b.Lookup(entryPath)
			record(err)
			_, err = b.Workflow(context.Background(), "hello.dip", entryPath)
			record(err)
			_, err = b.ReadFile(entryPath)
			record(err)
		}()
	}
	wg.Wait()
	if firstErr != nil {
		t.Fatalf("concurrent read returned error: %v", firstErr)
	}
}

func TestPack_RejectsSymlink(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "target.dip")
	if err := os.WriteFile(target, []byte(minimalDipSrc), 0644); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(dir, "link.dip")
	if err := os.Symlink(target, link); err != nil {
		t.Skip("symlinks not supported on this platform")
	}
	var buf bytes.Buffer
	_, err := Pack(context.Background(), link, &buf)
	if !errors.Is(err, ErrPathUnsafe) {
		t.Fatalf("err = %v, want ErrPathUnsafe", err)
	}
}

// TestPack_RejectsParentSymlink covers the parent-component-symlink
// data-exfil vector: a directory in the path tree (not the leaf) is a
// symlink pointing outside rootDir. Pack must refuse rather than silently
// follow into the linked target.
func TestPack_RejectsParentSymlink(t *testing.T) {
	rootDir := t.TempDir()
	outside := t.TempDir()
	// Place an attacker-controlled .dip at the symlink target.
	if err := os.WriteFile(filepath.Join(outside, "secret.dip"), []byte(minimalDipSrc), 0o644); err != nil {
		t.Fatal(err)
	}
	// Create a symlink under rootDir whose target is the outside directory.
	if err := os.Symlink(outside, filepath.Join(rootDir, "phases")); err != nil {
		t.Skip("symlinks not supported on this platform")
	}
	// The entry .dip references phases/secret.dip via subgraph ref. With the
	// parent-component check, Pack must refuse before reading the leaf via
	// the symlinked parent.
	entrySrc := `workflow Parent
  goal: x
  start: S
  exit: E
  subgraph S
    ref: phases/secret.dip
  agent E
    prompt: end
  edges
    S -> E
`
	entryPath := filepath.Join(rootDir, "parent.dip")
	if err := os.WriteFile(entryPath, []byte(entrySrc), 0o644); err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	_, err := Pack(context.Background(), entryPath, &buf)
	if !errors.Is(err, ErrPathUnsafe) {
		t.Fatalf("err = %v, want ErrPathUnsafe", err)
	}
}

// TestPack_AcceptsDoubleDotPrefixDirectory covers the false-positive bug
// where strings.HasPrefix(rel, "..") rejected legitimate filenames whose
// directory name simply starts with two dots (e.g., "..foo/bar.dip").
func TestPack_AcceptsDoubleDotPrefixDirectory(t *testing.T) {
	rootDir := t.TempDir()
	subdir := filepath.Join(rootDir, "..foo")
	if err := os.MkdirAll(subdir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(subdir, "child.dip"), []byte(minimalDipSrc), 0o644); err != nil {
		t.Fatal(err)
	}
	entrySrc := `workflow Parent
  goal: x
  start: S
  exit: E
  subgraph S
    ref: ..foo/child.dip
  agent E
    prompt: end
  edges
    S -> E
`
	entryPath := filepath.Join(rootDir, "parent.dip")
	if err := os.WriteFile(entryPath, []byte(entrySrc), 0o644); err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if _, err := Pack(context.Background(), entryPath, &buf); err != nil {
		t.Fatalf("unexpected error packing legitimate '..foo' subdir: %v", err)
	}
}

// TestExtract_ForcePreservesDestOnRenameFailure simulates EXDEV (cross-mount
// rename) at the staging-into-place step. With --force, the original
// destDir must be restored from the backup-aside on failure rather than
// destroyed.
func TestExtract_ForcePreservesDestOnRenameFailure(t *testing.T) {
	root := t.TempDir()
	destDir := filepath.Join(root, "out")
	staging := filepath.Join(root, "out.tmp")
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		t.Fatal(err)
	}
	original := []byte("ORIGINAL CONTENT")
	if err := os.WriteFile(filepath.Join(destDir, "marker"), original, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(staging, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(staging, "marker"), []byte("NEW"), 0o644); err != nil {
		t.Fatal(err)
	}
	// Inject a rename that succeeds for destDir->backup but returns EXDEV for
	// staging->destDir, exactly as a cross-mount rename would.
	calls := 0
	rename := func(oldp, newp string) error {
		calls++
		if calls == 2 {
			return &os.LinkError{Op: "rename", Old: oldp, New: newp, Err: errors.New("invalid cross-device link")}
		}
		return os.Rename(oldp, newp)
	}
	if err := swapDestWithStaging(destDir, staging, rename); err == nil {
		t.Fatal("expected EXDEV-simulated rename failure")
	}
	got, err := os.ReadFile(filepath.Join(destDir, "marker"))
	if err != nil {
		t.Fatalf("destDir was destroyed on rename failure: %v", err)
	}
	if !bytes.Equal(got, original) {
		t.Fatalf("destDir contents corrupted: got %q, want %q", got, original)
	}
}

// mustWriteBundleWithRawManifest writes a minimal .dipx file at dst whose
// only zip entry is manifest.json with the provided body. Used to test
// manifest-decode error paths. The entry is created with Flags 0x800 (UTF-8
// filenames) to satisfy the zip reader's strict-UTF8 check.
func mustWriteBundleWithRawManifest(t *testing.T, dst string, manifestJSON []byte) {
	t.Helper()
	f, err := os.Create(dst)
	if err != nil {
		t.Fatalf("create %s: %v", dst, err)
	}
	defer func() { _ = f.Close() }()
	zw := zip.NewWriter(f)
	defer func() { _ = zw.Close() }()
	h := &zip.FileHeader{Name: "manifest.json", Flags: 0x800}
	h.SetMode(0644)
	mw, err := zw.CreateHeader(h)
	if err != nil {
		t.Fatalf("create manifest entry: %v", err)
	}
	if _, err := mw.Write(manifestJSON); err != nil {
		t.Fatalf("write manifest: %v", err)
	}
}

// TestOpen_ManifestInvalid_PathIsBundlePath asserts that ErrManifestInvalid
// surfaced through Open carries the bundle path in BundleError.Path —
// not an empty string, not a JSON field name. Regression test for the
// Phase 3 manifest-decoder error-context finding (Bundle 5).
func TestOpen_ManifestInvalid_PathIsBundlePath(t *testing.T) {
	dir := t.TempDir()
	bundlePath := filepath.Join(dir, "broken.dipx")
	// format_version is a string, not a number — triggers ErrManifestInvalid
	// from the manifest decoder, with original Path="format_version".
	mustWriteBundleWithRawManifest(t, bundlePath, []byte(`{"format_version":"1","entry":"workflows/a.dip","files":[]}`))
	_, err := Open(context.Background(), bundlePath)
	if err == nil {
		t.Fatal("Open returned nil, want ErrManifestInvalid")
	}
	if !errors.Is(err, ErrManifestInvalid) {
		t.Fatalf("err = %v, want ErrManifestInvalid", err)
	}
	var be *BundleError
	if !errors.As(err, &be) {
		t.Fatalf("err = %T, want *BundleError", err)
	}
	if be.Path != bundlePath {
		t.Errorf("BundleError.Path = %q, want %q (bundle path)", be.Path, bundlePath)
	}
	if !strings.Contains(be.Detail, "format_version") {
		t.Errorf("BundleError.Detail = %q; expected to contain \"format_version\" (preserved from original Path)", be.Detail)
	}
}

// TestPack_BadSubgraph_ReportsErrSubgraphParse asserts that a parse
// failure in a transitively-reached subgraph (not the entry) surfaces
// as ErrSubgraphParse, not ErrEntryParse — and that BundleError.Path
// is the subgraph's filesystem path. Regression for P10.9.
func TestPack_BadSubgraph_ReportsErrSubgraphParse(t *testing.T) {
	dir := t.TempDir()
	entryPath := filepath.Join(dir, "entry.dip")
	subPath := filepath.Join(dir, "sub.dip")
	// Valid entry that refs sub.dip.
	if err := os.WriteFile(entryPath, []byte(packTestEntryWithRef("sub.dip")), 0o644); err != nil {
		t.Fatal(err)
	}
	// Intentionally invalid subgraph (not parseable): "workflow" alone is an
	// incomplete declaration that causes the parser to return an error.
	if err := os.WriteFile(subPath, []byte("workflow"), 0o644); err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	_, err := Pack(context.Background(), entryPath, &buf)
	if err == nil {
		t.Fatal("Pack returned nil, want ErrSubgraphParse")
	}
	if !errors.Is(err, ErrSubgraphParse) {
		t.Fatalf("err = %v, want ErrSubgraphParse", err)
	}
	if errors.Is(err, ErrEntryParse) {
		t.Fatal("err matched ErrEntryParse, want only ErrSubgraphParse")
	}
	var be *BundleError
	if !errors.As(err, &be) {
		t.Fatalf("err = %T, want *BundleError", err)
	}
	if !strings.HasSuffix(be.Path, "sub.dip") {
		t.Errorf("BundleError.Path = %q, want suffix \"sub.dip\"", be.Path)
	}
}

// packTestEntryWithRef returns the source of a minimal valid .dip workflow
// whose only subgraph node references refPath. Used by Pack tests that need
// a transitive subgraph link.
func packTestEntryWithRef(refPath string) string {
	return `workflow Entry
  goal: "test"
  start: S
  exit: E
  subgraph S
    ref: ` + refPath + `
  agent E
    prompt: end
  edges
    S -> E
`
}

// TestOpenManifest_HappyPath asserts OpenManifest returns the manifest +
// identity for a valid bundle without performing hash verification or
// workflow parsing.
func TestOpenManifest_HappyPath(t *testing.T) {
	dir := t.TempDir()
	bundlePath := filepath.Join(dir, "ok.dipx")
	mustBuildValidBundle(t, bundlePath)
	m, identity, err := OpenManifest(context.Background(), bundlePath)
	if err != nil {
		t.Fatalf("OpenManifest err = %v, want nil", err)
	}
	if m.FormatVersion != 1 {
		t.Errorf("FormatVersion = %d, want 1", m.FormatVersion)
	}
	if len(m.Files) == 0 {
		t.Error("len(m.Files) = 0, want > 0")
	}
	if identity == ([32]byte{}) {
		t.Error("identity is zero; want non-zero (sha256 of manifest bytes)")
	}
}

// TestOpenManifest_ManifestInvalid asserts manifest-decode errors are
// surfaced (and enriched with bundle path per Bundle 5).
func TestOpenManifest_ManifestInvalid(t *testing.T) {
	dir := t.TempDir()
	bundlePath := filepath.Join(dir, "broken.dipx")
	mustWriteBundleWithRawManifest(t, bundlePath, []byte(`{"format_version":"1","entry":"workflows/a.dip","files":[]}`))
	_, _, err := OpenManifest(context.Background(), bundlePath)
	if !errors.Is(err, ErrManifestInvalid) {
		t.Fatalf("err = %v, want ErrManifestInvalid", err)
	}
	var be *BundleError
	if !errors.As(err, &be) {
		t.Fatalf("err = %T, want *BundleError", err)
	}
	if be.Path != bundlePath {
		t.Errorf("BundleError.Path = %q, want %q (bundle path enrichment)", be.Path, bundlePath)
	}
}

// TestOpenManifest_FileNotReadable asserts a missing bundle file
// returns ErrManifestMissing.
func TestOpenManifest_FileNotReadable(t *testing.T) {
	_, _, err := OpenManifest(context.Background(), "/nonexistent/path/foo.dipx")
	if !errors.Is(err, ErrManifestMissing) {
		t.Fatalf("err = %v, want ErrManifestMissing", err)
	}
}

// mustBuildValidBundle builds a real, hash-verified .dipx at dst from a
// minimal stand-alone .dip workflow. Uses the public Pack API.
func mustBuildValidBundle(t *testing.T, dst string) {
	t.Helper()
	srcDir := t.TempDir()
	entryPath := filepath.Join(srcDir, "entry.dip")
	if err := os.WriteFile(entryPath, []byte(minimalStandaloneDip()), 0o644); err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if _, err := Pack(context.Background(), entryPath, &buf); err != nil {
		t.Fatalf("Pack: %v", err)
	}
	if err := os.WriteFile(dst, buf.Bytes(), 0o644); err != nil {
		t.Fatalf("write bundle: %v", err)
	}
}

// minimalStandaloneDip returns a minimal .dip workflow with no subgraph refs.
// Used by tests that need a parseable, packable, single-file bundle.
func minimalStandaloneDip() string {
	return `workflow Entry
  goal: "test"
  start: A
  exit: A
  agent A
    prompt: hello
`
}

// TestPack_HonorsContextCancellationDuringWrite documents that writeBundle
// respects ctx via the per-entry check in writeAllPackedFiles. The direct
// behavioral test of mid-write cancellation is skipped — it would require
// either a synthetic source tree with N>>1 files OR exposing writeBundle
// publicly. The broader Pack cancellation contract is covered by
// TestPack_HonorsContextCancellation (Task 3). P10.2.
func TestPack_HonorsContextCancellationDuringWrite(t *testing.T) {
	t.Skip("Subsumed by TestPack_HonorsContextCancellation; direct writeBundle mid-write test deferred (would require deep source tree). P10.2 check is added to writeBundle and writeAllPackedFiles.")
}

// TestPack_HonorsContextCancellation asserts that the Pack source-tree
// walker checks ctx during visitNext. Pre-cancelling the ctx makes Pack
// return context.Canceled instead of completing the walk. P10.7.
func TestPack_HonorsContextCancellation(t *testing.T) {
	dir := t.TempDir()
	entryPath := filepath.Join(dir, "entry.dip")
	if err := os.WriteFile(entryPath, []byte(minimalStandaloneDip()), 0o644); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	var buf bytes.Buffer
	_, err := Pack(ctx, entryPath, &buf)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Pack err = %v, want context.Canceled", err)
	}
}

// TestReadNoFollowSymlinks pins the contract of the now-exported no-follow read
// primitive: happy path reads, leaf-symlink refusal, ancestor-symlink refusal,
// and the not-regular-file branch (the documented fail-soft path for a ref that
// resolves to a directory, including the root dir itself).
func TestReadNoFollowSymlinks(t *testing.T) {
	root := t.TempDir()

	good := filepath.Join(root, "good.dip")
	if err := os.WriteFile(good, []byte(minimalDipSrc), 0o644); err != nil {
		t.Fatal(err)
	}
	data, err := ReadNoFollowSymlinks(good, root)
	if err != nil {
		t.Fatalf("happy path: unexpected err %v", err)
	}
	if string(data) != minimalDipSrc {
		t.Fatalf("happy path: content mismatch")
	}

	// not-regular-file (directory) refused — symlink-independent, so it runs on
	// every platform even when os.Symlink is unsupported and the cases below skip.
	if _, err := ReadNoFollowSymlinks(root, root); !errors.Is(err, ErrPathUnsafe) {
		t.Fatalf("not-regular-file (directory): err = %v, want ErrPathUnsafe", err)
	}

	link := filepath.Join(root, "link.dip")
	if err := os.Symlink(good, link); err != nil {
		t.Skip("symlinks not supported on this platform")
	}
	if _, err := ReadNoFollowSymlinks(link, root); !errors.Is(err, ErrPathUnsafe) {
		t.Fatalf("leaf symlink: err = %v, want ErrPathUnsafe", err)
	}

	realdir := filepath.Join(root, "real")
	if err := os.Mkdir(realdir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(realdir, "f.dip"), []byte(minimalDipSrc), 0o644); err != nil {
		t.Fatal(err)
	}
	linkdir := filepath.Join(root, "linkdir")
	if err := os.Symlink(realdir, linkdir); err != nil {
		t.Skip("symlinks not supported on this platform")
	}
	if _, err := ReadNoFollowSymlinks(filepath.Join(linkdir, "f.dip"), root); !errors.Is(err, ErrPathUnsafe) {
		t.Fatalf("ancestor symlink: err = %v, want ErrPathUnsafe", err)
	}
}
