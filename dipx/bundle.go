package dipx

import (
	"context"
	"crypto/sha256"

	"github.com/2389-research/dippin-lang/ir"
)

// Bundle is an opened .dipx. All workflows are parsed and normalized eagerly
// on Open; no file handles are held after Open returns. Bundle implements
// Source and is immutable post-Open.
type Bundle struct {
	manifest      Manifest
	manifestBytes []byte                  // for Identity()
	workflows     map[string]*ir.Workflow // canonical bundle path -> parsed workflow
	fileBytes     map[string][]byte       // canonical bundle path -> raw bytes
}

// Manifest returns a defensive copy of the parsed manifest. Callers may mutate
// the returned value without affecting the bundle. Cost is O(len(Files)).
func (b *Bundle) Manifest() Manifest {
	out := Manifest{
		FormatVersion: b.manifest.FormatVersion,
		Entry:         b.manifest.Entry,
		Files:         make([]ManifestEntry, len(b.manifest.Files)),
	}
	copy(out.Files, b.manifest.Files)
	return out
}

// Identity returns SHA-256(manifest.json bytes-as-stored). This is the
// authoritative bundle identity for provenance tracking.
func (b *Bundle) Identity() [32]byte {
	return sha256.Sum256(b.manifestBytes)
}

// ByteTotal returns the sum of uncompressed file sizes across every entry
// in the bundle. Only valid when the bundle was opened with full hash
// verification (i.e., via Open); a manifest-only inspection path
// (OpenManifest, Bundle 2 / Task 3) does not populate fileBytes and
// would return 0.
func (b *Bundle) ByteTotal() int64 {
	var total int64
	for _, raw := range b.fileBytes {
		total += int64(len(raw))
	}
	return total
}

// Entry returns the entry workflow.
func (b *Bundle) Entry() *ir.Workflow {
	return b.workflows[b.manifest.Entry]
}

// Resolve takes a parent's bundle-relative path and a ref string, and returns
// the bundle-relative path of the referenced workflow. Errors on path traversal
// or escape from workflows/.
func (b *Bundle) Resolve(refPath, relativeTo string) (string, error) {
	resolved, err := resolveLexically(refPath, relativeTo)
	if err != nil {
		return "", err
	}
	if _, ok := b.workflows[resolved]; !ok {
		return "", newError(ErrFileMissing, resolved, "ref resolves to path not in manifest", nil)
	}
	return resolved, nil
}

// Workflow resolves refPath relative to relativeTo and returns the parsed
// child workflow. Argument order matches flatten.Resolver.Resolve.
//
// ctx is checked at entry to satisfy the Source interface contract; the
// Bundle's workflows map is already in memory, so no actual I/O happens
// inside this method. Callers passing context.Background() are fine.
func (b *Bundle) Workflow(ctx context.Context, refPath, relativeTo string) (*ir.Workflow, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	resolved, err := b.Resolve(refPath, relativeTo)
	if err != nil {
		return nil, err
	}
	return b.workflows[resolved], nil
}
