package dipx

import (
	"context"
	"crypto/sha256"
	"io"
	"os"
	"path/filepath"

	"github.com/2389-research/dippin-lang/ir"
)

// openMode selects strict vs lax behavior on extra zip entries.
type openMode int

const (
	modeStrict openMode = iota
	modeLax
)

// Open reads a .dipx from disk in strict mode (the default).
func Open(ctx context.Context, path string) (*Bundle, error) {
	return openFile(ctx, path, modeStrict)
}

// OpenLax is Open with extra zip file entries silently tolerated. For
// hand-edited bundles only. NEVER call OpenLax on bytes obtained from any
// non-local source.
func OpenLax(ctx context.Context, path string) (*Bundle, error) {
	return openFile(ctx, path, modeLax)
}

// OpenReader is Open from any io.ReaderAt of known size.
func OpenReader(ctx context.Context, r io.ReaderAt, size int64) (*Bundle, error) {
	return openFromReader(ctx, r, size, modeStrict)
}

// OpenManifest performs only Open's structural-admission steps (zip open,
// manifest read, manifest decode, manifest shape verification) and returns
// the parsed manifest plus the bundle identity hash. Hash verification,
// extras check, parsing, ref walking, and conditional normalization are
// all skipped — this is the entry point for forensic-mode `dippin inspect
// --no-verify` against a bundle whose hashes may not match.
//
// Bundle path enrichment (Bundle 5) is applied to manifest-decode errors
// so external callers observe case (a) Path semantics consistently with
// Open.
func OpenManifest(ctx context.Context, path string) (Manifest, [32]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return Manifest{}, [32]byte{}, newError(ErrManifestMissing, path, "file not readable", err)
	}
	defer func() { _ = f.Close() }()
	stat, err := f.Stat()
	if err != nil {
		return Manifest{}, [32]byte{}, newError(ErrManifestMissing, path, "stat failed", err)
	}
	_, manifest, manifestBytes, err := openAndReadManifest(ctx, f, stat.Size())
	if err != nil {
		return Manifest{}, [32]byte{}, enrichBundlePath(err, path)
	}
	return manifest, sha256.Sum256(manifestBytes), nil
}

// Validate is Open-and-discard.
func Validate(ctx context.Context, path string) error {
	_, err := Open(ctx, path)
	return err
}

// openFile opens path on disk and delegates to openFromReader. Errors
// returned by openFromReader are enriched with the bundle path so external
// callers observe case (a) Path semantics per spec § "Per-sentinel error
// context": ErrManifestInvalid and ErrUnsupportedFormatVersion errors
// originating in the manifest decoder (which has no bundle path in scope)
// are rewritten to carry the bundle path here.
func openFile(ctx context.Context, path string, mode openMode) (*Bundle, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, newError(ErrManifestMissing, path, "file not readable", err)
	}
	defer func() { _ = f.Close() }()
	stat, err := f.Stat()
	if err != nil {
		return nil, newError(ErrManifestMissing, path, "stat failed", err)
	}
	bundle, err := openFromReader(ctx, f, stat.Size(), mode)
	if err != nil {
		return nil, enrichBundlePath(err, path)
	}
	return bundle, nil
}

// openFromReader runs the 10-step Open ordering: zip → manifest read → manifest
// decode → manifest shape → strict-extras check → hash verify → parse →
// walkRefs → normalize → build. ctx is checked before each disk-/CPU-bound
// stage.
func openFromReader(ctx context.Context, r io.ReaderAt, size int64, mode openMode) (*Bundle, error) {
	cz, manifest, manifestBytes, err := openAndReadManifest(ctx, r, size)
	if err != nil {
		return nil, err
	}
	if err := checkExtraEntriesCtx(ctx, cz, manifest, mode); err != nil {
		return nil, err
	}
	verified, err := verifyHashesCtx(ctx, cz, manifest)
	if err != nil {
		return nil, err
	}
	parsed, err := parseAndLink(ctx, verified, manifest)
	if err != nil {
		return nil, err
	}
	return buildBundle(manifest, manifestBytes, parsed, verified), nil
}

// openAndReadManifest performs steps 1-4 of the Open ordering: zip open,
// manifest read, manifest decode, manifest shape verification. Returns the
// constrained zip plus the parsed manifest and its raw bytes (for Identity).
func openAndReadManifest(ctx context.Context, r io.ReaderAt, size int64) (*constrainedZip, Manifest, []byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, Manifest{}, nil, err
	}
	cz, err := openConstrainedZip(r, size)
	if err != nil {
		return nil, Manifest{}, nil, err
	}
	manifestBytes, manifest, err := readAndDecodeManifest(cz)
	if err != nil {
		return nil, Manifest{}, nil, err
	}
	return cz, manifest, manifestBytes, nil
}

// readAndDecodeManifest reads manifest.json from cz, decodes it, and verifies
// its shape. Returns the raw bytes (for Identity) and the parsed Manifest.
func readAndDecodeManifest(cz *constrainedZip) ([]byte, Manifest, error) {
	manifestBytes, err := readManifestEntry(cz)
	if err != nil {
		return nil, Manifest{}, err
	}
	manifest, err := decodeManifest(manifestBytes)
	if err != nil {
		return nil, Manifest{}, err
	}
	if err := verifyManifestShape(manifest); err != nil {
		return nil, Manifest{}, err
	}
	return manifestBytes, manifest, nil
}

// checkExtraEntriesCtx checks ctx then runs the strict-extras check.
func checkExtraEntriesCtx(ctx context.Context, cz *constrainedZip, m Manifest, mode openMode) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return checkExtraEntries(cz, m, mode)
}

// verifyHashesCtx verifies every file's SHA-256, with per-entry ctx
// checks inside the loop (P10.10).
func verifyHashesCtx(ctx context.Context, cz *constrainedZip, m Manifest) (map[string]verifiedBytes, error) {
	verified, _, err := verifyAllHashesCtx(ctx, cz, m, maxTotalUncompBytes)
	return verified, err
}

// parseAndLink parses every verified workflow, walks refs to confirm closure
// and acyclicity, and normalizes parsed conditions on every workflow. ctx is
// checked between each CPU-bound stage so a canceled context aborts before
// the next pass starts.
func parseAndLink(ctx context.Context, verified map[string]verifiedBytes, m Manifest) (map[string]*ir.Workflow, error) {
	parsed, err := parseAllWorkflowsCtx(ctx, verified, m.Entry)
	if err != nil {
		return nil, err
	}
	if err := walkRefsCtx(ctx, parsed, m); err != nil {
		return nil, err
	}
	return normalizeConditionsCtx(ctx, parsed)
}

// parseAllWorkflowsCtx checks ctx then runs parseAllWorkflows.
func parseAllWorkflowsCtx(ctx context.Context, verified map[string]verifiedBytes, entry string) (map[string]*ir.Workflow, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return parseAllWorkflows(verified, entry)
}

// walkRefsCtx checks ctx then runs walkRefs.
func walkRefsCtx(ctx context.Context, parsed map[string]*ir.Workflow, m Manifest) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return walkRefs(parsed, m)
}

// normalizeConditionsCtx checks ctx then runs normalizeConditions, returning
// the unchanged parsed map (for chaining inside parseAndLink).
func normalizeConditionsCtx(ctx context.Context, parsed map[string]*ir.Workflow) (map[string]*ir.Workflow, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if err := normalizeConditions(parsed); err != nil {
		return nil, err
	}
	return parsed, nil
}

// buildBundle constructs the final Bundle, copying verified bytes into a
// path-keyed map for ReadFile.
func buildBundle(m Manifest, manifestBytes []byte, parsed map[string]*ir.Workflow, verified map[string]verifiedBytes) *Bundle {
	fileBytes := make(map[string][]byte, len(verified))
	for path, vb := range verified {
		fileBytes[path] = vb.Bytes()
	}
	return &Bundle{
		manifest:      m,
		manifestBytes: manifestBytes,
		workflows:     parsed,
		fileBytes:     fileBytes,
	}
}

// Pack builds a .dipx from an entry .dip on disk and writes it to w. Walks
// every transitively-reachable subgraph ref. Validates structurally, applies
// all path-safety and ZIP-feature constraints, and produces a deterministic
// byte stream. Returns the resulting Manifest.
func Pack(ctx context.Context, entryPath string, w io.Writer) (Manifest, error) {
	manifest, all, err := preparePackManifest(ctx, entryPath)
	if err != nil {
		return Manifest{}, err
	}
	if err := ctx.Err(); err != nil {
		return Manifest{}, err
	}
	if err := writeBundle(ctx, w, manifest, all); err != nil {
		return Manifest{}, err
	}
	return manifest, nil
}

// preparePackManifest walks the source tree and assembles a verified Manifest
// alongside the packed-file slice ready to write. Split out from Pack so each
// half stays under the project's complexity caps.
func preparePackManifest(ctx context.Context, entryPath string) (Manifest, []packedFile, error) {
	if err := ctx.Err(); err != nil {
		return Manifest{}, nil, err
	}
	entry, all, err := walkSourceTree(ctx, entryPath)
	if err != nil {
		return Manifest{}, nil, err
	}
	manifest := buildManifestForPack(entry, all)
	if err := verifyManifestShape(manifest); err != nil {
		return Manifest{}, nil, err
	}
	return manifest, all, nil
}

// Extract unpacks a .dipx into destDir atomically. Writes to destDir+".tmp"
// and renames on success. On failure the staging directory is removed and
// any pre-existing destDir is preserved.
func Extract(ctx context.Context, path, destDir string, allowOverwrite bool) error {
	if err := checkDestExists(destDir, allowOverwrite); err != nil {
		return err
	}
	bundle, err := Open(ctx, path)
	if err != nil {
		return err
	}
	staging := destDir + ".tmp"
	if err := stageBundle(ctx, bundle, staging); err != nil {
		return err
	}
	if err := swapDestWithStaging(destDir, staging, os.Rename); err != nil {
		_ = os.RemoveAll(staging)
		return err
	}
	return nil
}

// swapDestWithStaging atomically replaces destDir with staging. If destDir
// does not exist this collapses to a single rename. If destDir does exist,
// the rename-old-aside / rename-new-into-place / remove-aside sequence
// preserves the original on rename failure (notably EXDEV when staging is on
// a different mount). The rename function is injected so tests can simulate
// EXDEV.
func swapDestWithStaging(destDir, staging string, rename func(string, string) error) error {
	backup := destDir + ".bak"
	_ = os.RemoveAll(backup)
	_, err := os.Lstat(destDir)
	if os.IsNotExist(err) {
		return rename(staging, destDir)
	}
	if err != nil {
		return err
	}
	return swapWithBackup(destDir, staging, backup, rename)
}

// swapWithBackup performs the three-step swap when destDir is known to
// exist: rename destDir to backup, rename staging to destDir, then remove
// the backup on success. On the second rename's failure, restore destDir
// from backup before returning the error.
func swapWithBackup(destDir, staging, backup string, rename func(string, string) error) error {
	if err := rename(destDir, backup); err != nil {
		return err
	}
	if err := rename(staging, destDir); err != nil {
		_ = rename(backup, destDir)
		return err
	}
	_ = os.RemoveAll(backup)
	return nil
}

// checkDestExists returns ErrPathUnsafe when destDir exists and overwrite is
// disallowed. IsNotExist means the path is safe to use. Any other stat error
// (e.g. permission denied because the parent directory is unreadable) is
// surfaced to the caller as the underlying *os.PathError so classifyExit's
// isIOErr can route to the documented I/O exit code (3) — failing later in
// stageBundle's MkdirAll would produce a less clear diagnostic.
func checkDestExists(destDir string, allowOverwrite bool) error {
	if allowOverwrite {
		return nil
	}
	_, err := os.Stat(destDir)
	if err == nil {
		return newError(ErrPathUnsafe, destDir, "destination exists; use --force", nil)
	}
	if !os.IsNotExist(err) {
		return err
	}
	return nil
}

// stageBundle creates a fresh staging directory and writes the bundle into it.
// On any failure the staging directory is removed before returning.
func stageBundle(ctx context.Context, bundle *Bundle, staging string) error {
	if err := os.RemoveAll(staging); err != nil {
		return err
	}
	if err := os.MkdirAll(staging, 0o755); err != nil {
		return err
	}
	if err := writeBundleToDir(ctx, bundle, staging); err != nil {
		_ = os.RemoveAll(staging)
		return err
	}
	return nil
}

// writeBundleToDir writes every bundle file plus manifest.json under root.
func writeBundleToDir(ctx context.Context, b *Bundle, root string) error {
	for path, raw := range b.fileBytes {
		if err := writeOneFile(ctx, root, path, raw); err != nil {
			return err
		}
	}
	manifestPath := filepath.Join(root, "manifest.json")
	return os.WriteFile(manifestPath, b.manifestBytes, 0o644)
}

// writeOneFile writes a single bundle-relative file under root, creating any
// missing parent directories. ctx is checked first so a canceled extract
// aborts before any further disk work.
func writeOneFile(ctx context.Context, root, path string, raw []byte) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	full := filepath.Join(root, filepath.FromSlash(path))
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		return err
	}
	return os.WriteFile(full, raw, 0o644)
}

// checkExtraEntries enforces strict mode: any non-directory zip entry not
// listed in files[] is rejected. Directory entries are always ignored
// (already filtered out at constrainedZip construction).
func checkExtraEntries(cz *constrainedZip, m Manifest, mode openMode) error {
	if mode == modeLax {
		return nil
	}
	listed := make(map[string]struct{}, len(m.Files)+1)
	listed["manifest.json"] = struct{}{}
	for _, e := range m.Files {
		listed[e.Path] = struct{}{}
	}
	for name := range cz.entries {
		if _, ok := listed[name]; !ok {
			return newError(ErrFileUnexpected, name, "", nil)
		}
	}
	return nil
}
