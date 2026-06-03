//go:build unix

package parser

import "syscall"

// oNoFollow makes os.OpenFile reject a symlink in the final path component
// atomically: open() fails with ELOOP if the leaf is a symlink. This closes the
// leaf check-to-read TOCTOU race (#79) without a separate Lstat. It affects only
// the final component — contained parent symlinks are still followed and remain
// validated by checkContainment.
//
// Scoped to the `unix` meta-constraint (darwin, linux, *bsd, …) rather than
// !windows: O_NOFOLLOW is only defined on those targets. Non-unix builds —
// notably js/wasm (the docs-site playground builds GOOS=js) and Windows — take
// the oNoFollow=0 fallback in resolve_nofollow_other.go.
const oNoFollow = syscall.O_NOFOLLOW
