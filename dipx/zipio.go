package dipx

import (
	"archive/zip"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"strings"
)

// verifiedBytes wraps a byte slice produced by verifyHashes. The unexported
// type combined with the lack of any constructor besides newVerifiedBytes
// makes "parse from a non-verified source" a compile-time error: no code
// outside this package can manufacture a verifiedBytes value.
type verifiedBytes struct{ b []byte }

func newVerifiedBytes(b []byte) verifiedBytes { return verifiedBytes{b: b} }
func (v verifiedBytes) Bytes() []byte         { return v.b }

// constrainedZip is the result of openConstrainedZip: a reader plus a map of
// canonical entry name -> *zip.File that has already passed the spec's ZIP
// feature constraints.
type constrainedZip struct {
	reader  *zip.Reader
	entries map[string]*zip.File // non-directory entries only, keyed by entry name
}

// maxZipFiles is the spec's conformance limit on entry count.
const maxZipFiles = 10000

// openConstrainedZip wraps zip.NewReader and enforces the spec's v1 ZIP
// feature constraints: rejects encryption (general-purpose bit 0), non-UTF-8
// filenames (bit 11 must be set), non-Store/non-Deflate compression methods,
// symlink and other non-regular Unix mode bits, byte-equal duplicate entry
// names, ASCII-case-fold-equal duplicate entry names, and the reserved
// manifest.sig entry name (v2 signatures). Directory entries
// (names ending in `/`) are silently skipped.
//
// What this function does NOT enforce in v1 (deferred to v1.1, see
// docs/superpowers/plans/2026-05-07-dipx-followups.md):
//   - Multi-disk archives: incidentally rejected when zip.NewReader fails to
//     parse them, not via an explicit check here.
//   - Central-directory ↔ local-header agreement on filename and uncompressed
//     size (Phase 4 C2). v1 mitigates by per-file SHA-256 verification: any
//     cross-record disagreement between CD-described bytes and LH-described
//     bytes that gets decompressed will fail the manifest hash.
//   - Per-file 1000:1 compression-ratio cap (Phase 4 C3). v1 enforces only
//     the absolute 50 MB per-file uncompressed cap.
//   - Unicode case-folding for duplicate detection (Phase 9 P9.2). v1 uses
//     ASCII strings.ToLower so collisions like ß/ss may slip through.
func openConstrainedZip(r io.ReaderAt, size int64) (*constrainedZip, error) {
	zr, err := zip.NewReader(r, size)
	if err != nil {
		return nil, newError(ErrZipTruncated, "", "zip parse failed", err)
	}
	if len(zr.File) > maxZipFiles {
		return nil, newError(ErrCapExceeded, "", fmt.Sprintf("zip contains %d entries; max %d", len(zr.File), maxZipFiles), nil)
	}
	cz := &constrainedZip{reader: zr, entries: make(map[string]*zip.File, len(zr.File))}
	seenFold := make(map[string]struct{}, len(zr.File))
	for _, f := range zr.File {
		if err := admitZipEntry(cz, seenFold, f); err != nil {
			return nil, err
		}
	}
	return cz, nil
}

// admitZipEntry validates a single zip entry against feature constraints and
// records it (or silently skips it for directories) in cz/seenFold.
func admitZipEntry(cz *constrainedZip, seenFold map[string]struct{}, f *zip.File) error {
	if err := checkZipEntry(f); err != nil {
		return err
	}
	// Directory entries: silently ignored.
	if strings.HasSuffix(f.Name, "/") {
		return nil
	}
	return recordEntry(cz, seenFold, f)
}

// recordEntry inserts f into cz.entries / seenFold after duplicate checks.
func recordEntry(cz *constrainedZip, seenFold map[string]struct{}, f *zip.File) error {
	if _, dup := cz.entries[f.Name]; dup {
		return newError(ErrZipFeatureForbidden, f.Name, "duplicate entry", nil)
	}
	fold := strings.ToLower(f.Name)
	if _, dup := seenFold[fold]; dup {
		return newError(ErrZipFeatureForbidden, f.Name, "case-fold-duplicate entry", nil)
	}
	cz.entries[f.Name] = f
	seenFold[fold] = struct{}{}
	return nil
}

func checkZipEntry(f *zip.File) error {
	if f.Name == "manifest.sig" {
		return newError(ErrZipFeatureForbidden, f.Name, "manifest.sig is reserved for v2 signatures", nil)
	}
	if err := checkZipEntryFlags(f); err != nil {
		return err
	}
	if err := checkZipEntryMethod(f); err != nil {
		return err
	}
	return checkZipEntryMode(f)
}

// checkZipEntryFlags rejects encryption and non-UTF-8 filename encodings.
func checkZipEntryFlags(f *zip.File) error {
	// Encryption: bit 0 of GeneralPurposeFlag.
	if f.Flags&0x1 != 0 {
		return newError(ErrZipFeatureForbidden, f.Name, "encrypted entry", nil)
	}
	// UTF-8 filename: bit 11 is required unconditionally per spec, to defend
	// against CP437-encoded ASCII names that bypass UTF-8 enforcement.
	if f.Flags&0x800 == 0 {
		return newError(ErrZipFeatureForbidden, f.Name, "non-UTF-8 filename encoding (general-purpose bit 11 not set)", nil)
	}
	return nil
}

// checkZipEntryMethod rejects compression methods other than Store/Deflate.
func checkZipEntryMethod(f *zip.File) error {
	if f.Method != zip.Store && f.Method != zip.Deflate {
		return newError(ErrZipFeatureForbidden, f.Name, fmt.Sprintf("unsupported compression method %d", f.Method), nil)
	}
	return nil
}

// checkZipEntryMode rejects all non-regular, non-directory file modes (Unix
// creator only). Devices, sockets, named pipes, and irregular files share the
// "not a real file" risk profile that motivates rejecting symlinks.
func checkZipEntryMode(f *zip.File) error {
	// External attributes encode mode bits in the upper 16 bits when
	// CreatorVersion specifies Unix (3).
	if (f.CreatorVersion >> 8) != 3 {
		return nil
	}
	disallowed := os.ModeSymlink | os.ModeDevice | os.ModeNamedPipe | os.ModeSocket | os.ModeCharDevice | os.ModeIrregular
	if f.Mode()&disallowed != 0 {
		return newError(ErrZipFeatureForbidden, f.Name, "non-regular file mode bits set", nil)
	}
	return nil
}

// verifyAndReadEntry reads a single zip entry's decompressed bytes, enforcing
// the per-file size cap as a streaming bound (io.LimitReader), computing
// SHA-256 in tandem, and comparing against the manifest hash. Returns
// verifiedBytes — the only constructor of that type — so downstream code can
// only access bytes that have been hash-verified.
func verifyAndReadEntry(cz *constrainedZip, path, expectedHex string, perFileCap int64) (verifiedBytes, error) {
	f, ok := cz.entries[path]
	if !ok {
		return verifiedBytes{}, newError(ErrFileMissing, path, "", nil)
	}
	buf, gotHex, err := readEntryWithHash(f, path, perFileCap)
	if err != nil {
		return verifiedBytes{}, err
	}
	if gotHex != expectedHex {
		return verifiedBytes{}, newError(ErrHashMismatch, path, fmt.Sprintf("expected: %s; actual: %s", expectedHex, gotHex), nil)
	}
	return newVerifiedBytes(buf), nil
}

// readEntryWithHash opens a zip entry, reads it under a streaming size cap,
// and returns its bytes along with the hex-encoded SHA-256 digest.
func readEntryWithHash(f *zip.File, path string, perFileCap int64) ([]byte, string, error) {
	rc, err := f.Open()
	if err != nil {
		return nil, "", newError(ErrZipTruncated, path, "open failed", err)
	}
	defer rc.Close()

	limited := &io.LimitedReader{R: rc, N: perFileCap + 1}
	h := sha256.New()
	buf, err := io.ReadAll(io.TeeReader(limited, h))
	if err != nil {
		return nil, "", newError(ErrZipTruncated, path, "read failed", err)
	}
	if int64(len(buf)) > perFileCap {
		return nil, "", newError(ErrCapExceeded, path, fmt.Sprintf("file exceeds %d bytes", perFileCap), nil)
	}
	// Defense-in-depth: if the decompressed size doesn't match the central-dir
	// header, surface as ErrZipTruncated rather than letting the hash
	// comparison swallow it as ErrHashMismatch.
	if int64(len(buf)) != int64(f.UncompressedSize64) {
		return nil, "", newError(ErrZipTruncated, path, fmt.Sprintf("decompressed %d bytes, header claimed %d", len(buf), f.UncompressedSize64), nil)
	}
	return buf, hex.EncodeToString(h.Sum(nil)), nil
}
