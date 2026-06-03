//go:build windows

package parser

// oNoFollow is 0 on Windows: there is no atomic O_NOFOLLOW equivalent for
// os.OpenFile. Windows is NOT a supported build/release target for dippin
// (.goreleaser.yml ships darwin + linux only; CI is Linux-only) — this file
// exists solely so the package compiles under GOOS=windows.
//
// With oNoFollow == 0, os.OpenFile FOLLOWS a leaf symlink and opens its target,
// so the fstat in checkFileInfo sees the target (never ModeSymlink) and cannot
// reject a leaf symlink. The fd-based fstat→read still closes the fstat-to-read
// race, but atomic leaf-symlink rejection is a Unix-only guarantee (#79).
const oNoFollow = 0
