//go:build !windows

package parser

import "syscall"

// oNoFollow makes os.OpenFile reject a symlink in the final path component
// atomically: open() fails with ELOOP if the leaf is a symlink. This closes the
// leaf check-to-read TOCTOU race (#79) without a separate Lstat. It affects only
// the final component — contained parent symlinks are still followed and remain
// validated by checkContainment.
const oNoFollow = syscall.O_NOFOLLOW
