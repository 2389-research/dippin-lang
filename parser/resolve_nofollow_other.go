//go:build !unix

package parser

// oNoFollow is 0 on every non-unix target — Windows, js/wasm (the docs-site
// playground builds GOOS=js), plan9, etc. — where os.OpenFile has no atomic
// O_NOFOLLOW equivalent. dippin ships only darwin + linux (.goreleaser.yml; CI
// is Linux-only); this file exists so the parser package still compiles for
// those non-unix builds.
//
// With oNoFollow == 0, os.OpenFile FOLLOWS a leaf symlink and opens its target,
// so the fstat in checkFileInfo sees the target (never ModeSymlink) and cannot
// reject a leaf symlink. The fd-based fstat→read still closes the fstat-to-read
// race, but atomic leaf-symlink rejection is a unix-only guarantee (#79).
const oNoFollow = 0
