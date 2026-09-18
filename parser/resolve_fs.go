package parser

import (
	"fmt"
	"io/fs"
	"path"
	"strings"
)

// fsReader returns the fs.FS directiveReader backing ResolveFileDirectivesFS:
// every path is resolved as path.Join(baseDir, p) inside fsys with the FS
// safety policy (no absolute p, no ".." segment, fs.ValidPath, no symlink
// component when the FS is a ReadLinkFS, regular file only, size cap).
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
	if err := rejectSymlinkComponents(fsys, baseDir, p); err != nil {
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

// rejectSymlinkComponents mirrors the disk resolver's symlink policy on an FS
// that can report symlinks (fs.ReadLinkFS — os.DirFS, fstest.MapFS): every
// component of p below baseDir is Lstat'ed and a symlink anywhere in the chain
// is rejected — a leaf symlink with the disk path's "symlinks not allowed"
// message (O_NOFOLLOW parity), a symlinked parent with the containment message.
// Without this, an os.DirFS Open would follow the link and Stat would report
// the target as a regular file, so a directive could read host content from
// outside the FS root. Components of baseDir itself are not checked, matching
// the disk path, which resolves and accepts the base as-is. An FS that does not
// implement ReadLinkFS cannot express symlinks and is trusted as self-contained.
func rejectSymlinkComponents(fsys fs.FS, baseDir, p string) error {
	rl, ok := fsys.(fs.ReadLinkFS)
	if !ok {
		return nil
	}
	segs := strings.Split(p, "/")
	for i := range segs {
		info, err := rl.Lstat(path.Join(baseDir, strings.Join(segs[:i+1], "/")))
		if err != nil {
			return pathErr(p, err, "stat")
		}
		if info.Mode()&fs.ModeSymlink != 0 {
			return symlinkErr(p, i == len(segs)-1)
		}
	}
	return nil
}

// symlinkErr names the user path only, with the same message the disk
// resolver emits for the matching case (leaf vs. parent component).
func symlinkErr(p string, leaf bool) error {
	if leaf {
		return fmt.Errorf("symlinks not allowed: %q", p)
	}
	return fmt.Errorf("path %q resolves outside source directory", p)
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
