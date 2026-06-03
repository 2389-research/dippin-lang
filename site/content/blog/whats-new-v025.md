---
title: "What's New in Dippin v0.25"
date: "2026-05-11"
description: ".dipx format v1.1 — real cancellation through Pack and Open, an inspect that actually inspects, and exit code 2 that actually fires when the spec says it should."
tagStyle: "release"
tagLabel: "RELEASE"
category: "Releases"
readTime: "5 min read"
related:
  - url: "whats-new-v026.html"
    title: "What's New in v0.26"
    summary: "The workflow `requires:` keyword — declare environmental dependencies so runtimes can preflight."
  - url: "whats-new-v024.html"
    title: "What's New in v0.24"
    summary: "The `.dipx` bundle format — pack a workflow tree into one verifiable file you can ship anywhere."
---

`v0.24.0` introduced `.dipx` — a deterministic, hash-verified bundle format for shipping workflows. The shape was right. The implementation was mostly right. But shipping a format means it leaves your machine and starts getting opened, packed, and inspected by *other* code — runtimes, CI scripts, forensic tools — and that's where the gaps showed up.

`v0.25.0` is the follow-up release that closes them. The format itself (v1.1 in the spec) is unchanged in any way that affects bundles you've already packed. What changes is the contract between the library and its callers: cancellation works end-to-end, `inspect` actually does what its flags claim, error sentinels match the spec, and the exit codes you script against are the ones the spec promises.

If you're a downstream consumer (anything importing `github.com/2389-research/dippin-lang/dipx`), there are two breaking changes — both small, both worth taking. Migration is at the bottom.

## Cancellation that actually cancels

`Pack` walks a source tree and hashes every file. `Open` verifies every hash against the manifest. Both are O(files × bytes). For a small pipeline that's nothing; for a many-file workflow with large heredocs, it can take long enough that you'd like to be able to cancel.

Before v0.25, you could pass a `context.Context` into the entry points, but it was decorative for most of the work. `verifyAllHashes` opened and hashed every entry in a single loop with no ctx check. `walkSourceTree` did the same on the producer side. `writeBundle` happily streamed entries to disk regardless. So you'd cancel a Pack, the calling goroutine would unblock, and the work would continue running to completion in the background.

v0.25 fixes the hot paths:

- `verifyAllHashes` checks `ctx.Err()` between each manifest entry (P10.10).
- `walkSourceTree` checks `ctx.Err()` between each iteration of the ref walk (P10.7).
- `writeBundle` checks `ctx.Err()` between each entry it writes (P10.2).

The practical effect: a long Open against a many-entry bundle, or a long Pack against a deep source tree, can be canceled within one entry's worth of processing time instead of one bundle's worth. Whatever timeout you set actually means something now.

Part of making cancellation honest required a small but breaking signature change to `Source.Workflow`. It now takes `context.Context` as its first argument:

```go
// before (v0.24)
wf, err := source.Workflow(ref, parent)

// after (v0.25)
wf, err := source.Workflow(ctx, ref, parent)
```

Both `dirSource.Workflow` (on-disk) and `Bundle.Workflow` (bundled) check the context at entry, so subgraph resolution can be canceled even mid-walk through a deep ref graph. If you're a runtime integrator or otherwise call `Source.Workflow` directly, this is the one place you'll need to touch.

## Inspect actually inspects

`dippin inspect` had two issues worth complaining about.

**`--no-verify` did nothing.** The flag existed. It printed a warning. Then `inspect` ran the full hash-verification pipeline anyway. The use case the flag was supposed to serve — forensic inspection of a *tampered* bundle, where you want to see what's inside without integrity errors firing first — was unreachable. You'd run `inspect --no-verify badbundle.dipx`, get back `ErrHashMismatch`, and have no way to find out what was actually in there.

v0.25 routes `--no-verify` through a new `dipx.OpenManifest` API that performs only structural-admission steps: open the ZIP, read the manifest, validate its shape, list the entries. No hash verification. Tampered bundles open. You can see what's inside.

```sh
$ dippin inspect --no-verify tampered.dipx
format: 1
entry:    workflows/pipeline.dip
identity: sha256:0a7d9f... (manifest)
files:
  workflows/pipeline.dip          sha256:abc123…
  workflows/phases/review.dip     sha256:def456…
status:
  valid:          false
  verify_skipped: true
  file_count:     2
  byte_total:     8421
  format_version: 1
```

**`--format=json` returned `status: "VALID"` as a bare string.** No room to express "valid but verification was skipped," no count of files, no byte total. The spec called for a structured object; v0.24 shipped a string. v0.25 lands the spec-compliant shape:

```json
{
  "status": {
    "valid": true,
    "verify_skipped": false,
    "file_count": 2,
    "byte_total": 8421,
    "format_version": 1
  }
}
```

This is a **breaking change for anyone parsing `dippin inspect --format=json` output in scripts.** If you decoded `status` as a string, you need to decode it as an object now. Migration is one struct change.

## Errors point to the right file

Two attribution bugs got fixed.

**Pack errors no longer lie about which file failed.** When `dippin pack` parsed the workflow tree and a subgraph file had a syntax error, every parse failure surfaced as `ErrEntryParse` with the entry workflow's path — regardless of which file actually failed. So you'd get `entry workflow parse error: pipeline.dip` when the broken file was actually `phases/review.dip`. v0.25 classifies subgraph parse failures as `ErrSubgraphParse` with the offending subgraph's filesystem path. Errors now identify their source.

**Manifest-decode errors carry the bundle path.** Previously, if `Open` failed to decode a bundle's `manifest.json`, `BundleError.Path` would either be empty (`ErrUnsupportedFormatVersion`) or a JSON field name (`"format_version"` for `ErrManifestInvalid`). External callers had no way to log "which bundle did this come from?" without threading the path back through themselves. v0.25's `Open` enriches the error with the bundle path before returning, preserving any original Path in the `Detail` field for diagnostic purposes.

## Exit code 2 actually means integrity failure

The `.dipx` spec enumerates 12 integrity-failure sentinels — errors that mean "the bundle is structurally wrong or tampered," distinct from user errors like "you typed the wrong filename." The CLI commands `pack`, `unpack`, and `inspect` are supposed to exit with code 2 for any of those, and code 1 for everything else.

v0.24 shipped a check that matched only 5 of the 12 sentinels. The other 7 — `ErrUnsupportedFormatVersion`, `ErrFileMissing`, `ErrFileUnexpected`, `ErrEntryNotInManifest`, `ErrRefEscape`, `ErrRefCycle`, `ErrCapExceeded`, `ErrPathUnsafe` — defaulted to user-error 1. A script trying to distinguish "this bundle is corrupt" from "the user typo'd the filename" couldn't reliably do so.

v0.25 refactors `isIntegrityErr` to a sentinel-slice + loop covering all 12 spec-enumerated sentinels. The exit-code contract now actually matches what the spec promised:

| Exit code | Meaning |
|---|---|
| 0 | OK |
| 1 | User error (bad CLI args, file not found, etc.) |
| 2 | Integrity error (bundle is malformed, tampered, or violates spec invariants) |
| 3 | I/O error |
| 4 | Cancelled |

## Cycle detection covers every workflow

`dipx.Open` walks the workflow ref graph and rejects cycles — but it was only walking from `m.Entry` (the manifest's entry workflow), not from every workflow listed in the manifest. Meanwhile, `parseAllWorkflows` parses *every* manifest-listed workflow, including ones unreachable from entry.

The gap: a cycle between two workflows that were both in the manifest but not reachable from entry would slip through the cycle check, get parsed successfully, and sit in the bundle waiting to misbehave if anything ever resolved into it.

v0.25 has `walkRefs` iterate `detectCycles` over every entry in `m.Files`. Every manifest-listed workflow now participates in cycle detection, matching what `parseAllWorkflows` does next door.

## Seven spec clarifications

Bundle 6 closed seven ambiguities in the `.dipx` spec. None change observable behavior beyond what's already covered above; all of them tighten the contract:

- **Path canonicalization rule 2** narrowed to "Backslash `\` MUST be rejected." The implementation already rejected only backslash; the spec wording was over-broad.
- **Per-sentinel error-context preamble** documenting `BundleError.Path` semantics across three real cases: bundle-relative (read-side, post-Open), JSON field name (manifest decode pre-bundle-context), source filesystem path (Pack-side).
- **Open ordering step 5** ("Verify no extra zip entries") inserted as a normative step between manifest-shape validation and hash verification. `ErrFileUnexpected` added to the error-precedence list at category 4.
- **Cycle detection scope** documented: "every manifest-listed workflow," matching `parseAllWorkflows`.
- **Integrity-failure sentinel set** for CLI exit code 2 expanded from 5 to all 12 spec-enumerated sentinels.
- **`inspect --format=json` status object schema** documented with a concrete example.
- **Runtime integration migration example** updated to include `ctx` in `Source.Workflow` calls.

Specs are easier to read when there's one obvious answer for each question. These commits make seven previously-ambiguous questions obvious.

## Migration

If you don't import `github.com/2389-research/dippin-lang/dipx` directly and you don't parse `dippin inspect --format=json` output in scripts, you don't need to do anything — bundles you packed under v0.24 continue to open under v0.25 (the on-the-wire format is unchanged; only the library contract moved). Just upgrade:

```sh
go install github.com/2389-research/dippin-lang/cmd/dippin@latest
```

If you *do* import the library — custom analyzers, runtimes, anything that goes deeper than the CLI:

1. `go get github.com/2389-research/dippin-lang@v0.25.0` (or `@latest`).
2. Update any call to `source.Workflow(ref, parent)` to `source.Workflow(ctx, ref, parent)`. The compiler will tell you exactly where.
3. If you parse `dippin inspect --format=json` in scripts, decode `status` as an object (`valid`, `verify_skipped`, `file_count`, `byte_total`, `format_version`) instead of a string.

That's the whole list.

## What's next

Full notes in [CHANGELOG.md](https://github.com/2389-research/dippin-lang/blob/main/CHANGELOG.md). The deferred items from v1 — per-file compression-ratio cap, Unicode case-folding for duplicate detection, central-directory/local-header agreement, `DIPX_DEBUG=1` step trace — are still on the v1.x roadmap; none of them block production use of the format today.

`v0.24` said the `.dipx` format existed. `v0.25` is the release that makes it the thing you actually want to deploy from.
