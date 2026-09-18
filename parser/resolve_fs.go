package parser

import (
	"fmt"
	"io/fs"
	"path"
	"strings"
)

// fsReader returns the fs.FS directiveReader backing ResolveFileDirectivesFS:
// every path is resolved as path.Join(baseDir, p) inside fsys with the FS
// safety policy (no absolute p, no ".." segment, fs.ValidPath, regular file
// only, size cap).
func fsReader(fsys fs.FS, baseDir string) directiveReader {
	return func(p string) ([]byte, error) { return loadDirectiveFileFS(fsys, baseDir, p) }
}

// loadDirectiveFileFS resolves p under baseDir within fsys and reads it. Error
// messages reference the user-written path (p), never the joined FS name, so
// diagnostics have the same shape as the disk resolver's.
func loadDirectiveFileFS(fsys fs.FS, baseDir, p string) ([]byte, error) {
	name, err := fsResolve(baseDir, p)
	if err != nil {
		return nil, err
	}
	return openCheckReadFS(fsys, p, name)
}

// fsResolve is the fs.FS counterpart of safeResolve: a purely lexical check
// that p is relative, contains no ".." segment, and joins with baseDir to an
// fs.ValidPath name. The checks run on p BEFORE joining so a ".." can never be
// canceled out by a deeper prefix (e.g. "legit/../../x"). There is no
// symlink-chain containment step here: an fs.FS has no symlink semantics of its
// own, so lexical containment is complete for it.
func fsResolve(baseDir, p string) (string, error) {
	if path.IsAbs(p) {
		return "", fmt.Errorf("absolute paths not allowed: %q", p)
	}
	if hasParentSegment(p) {
		return "", fmt.Errorf("path %q resolves outside source directory", p)
	}
	name := path.Join(baseDir, p)
	if !fs.ValidPath(name) {
		return "", fmt.Errorf("path %q is not a valid fs path", p)
	}
	return name, nil
}

// hasParentSegment reports whether the slash-separated path p contains a ".."
// element. Unlike hasParentRef it does not go through filepath.ToSlash: fs.FS
// paths are slash-separated on every platform, so a backslash is an ordinary
// character here, never a separator.
func hasParentSegment(p string) bool {
	for _, seg := range strings.Split(p, "/") {
		if seg == ".." {
			return true
		}
	}
	return false
}

// openCheckReadFS opens name in fsys exactly once and validates + reads the
// file through that single handle, mirroring openCheckRead's open→stat→read
// order. A directory or other non-regular entry is rejected up front: fsys.Open
// succeeds on a directory for embed.FS and fstest.MapFS, and reading it would
// otherwise surface as an opaque "not accessible" error.
func openCheckReadFS(fsys fs.FS, p, name string) ([]byte, error) {
	f, err := fsys.Open(name)
	if err != nil {
		return nil, pathErr(p, err, "open")
	}
	defer func() { _ = f.Close() }()
	info, err := f.Stat()
	if err != nil {
		return nil, pathErr(p, err, "stat")
	}
	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("file %q is not a regular file", p)
	}
	if err := checkFileInfo(p, info); err != nil {
		return nil, err
	}
	return readFromFD(p, f)
}
