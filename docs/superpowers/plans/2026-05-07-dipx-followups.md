# `.dipx` v1 Follow-ups (deferred from gate reviews)

This file collects findings from the per-phase gate reviews that are NOT blocking v1 but should be addressed in v1.1 or later.

## Spec gaps in path canonicalization (Phase 2 gate)

Path canonicalization rules 1 and 4 (per spec § "Path canonicalization") cover NFC normalization and ASCII control-character rejection respectively, but do not address:

- **NFKC compatibility equivalents**: e.g., `ﬁ` (U+FB01) and `fi` (U+0066 U+0069) are visually identical but byte-distinct, both NFC-stable.
- **Format-class Unicode characters**: ZWJ (U+200D), ZWNJ (U+200C), ZWNBSP/BOM (U+FEFF) are mid-string-invisible.
- **Bidi controls**: U+202A–U+202E, U+2066–U+2069 reorder display.
- **IDN homographs**: Cyrillic `а` (U+0430) vs Latin `a` (U+0061) for visual spoofing.

These pass current Canonicalize rules and could enable visual-spoofing attacks at the operator-UI layer.

**Disposition**: defer to v1.1. Investigate adding NFKC normalization rejection or a "no format-class characters" rule. Coordinate with downstream consumers (which present bundle paths to operators).

## Wider "any other separator" interpretation (Phase 2 gate)

Spec rule 2 says "Backslash `\` and any other separator MUST be rejected." Current implementation only rejects `\`. Other Unicode separator-like glyphs (U+2044 FRACTION SLASH, U+FF0F FULLWIDTH SOLIDUS, U+29F8 BIG SOLIDUS, U+FF3C FULLWIDTH BACKSLASH) are not blocked. None are real path separators on supported OSes, but the spec wording is broader than the implementation.

**Disposition**: narrow the spec wording in v1.1 (recommended), or broaden the rune check in `checkRune`. Practical risk is low.

## Test coverage gaps (Phase 2 gate)

Adversarial input categories not yet tested:
- Bidi marks (U+200E/U+200F)
- ZWNBSP/BOM (U+FEFF)
- IDN homograph pairs
- Percent-encoded `..` (e.g., `%2e%2e`) — spec confirms this is treated literally; tests would lock in the behavior.

**Disposition**: add as v1.1 test additions if NFKC normalization is added.

## Manifest decoder error-context (Phase 3 gate)

The crypto-discipline gate flagged that `BundleError.Path` for `ErrManifestInvalid` and `ErrUnsupportedFormatVersion` is sometimes empty (preflight, JSON pre-pass) or carries the field name (`format_version`) rather than the bundle path. The spec § "Per-sentinel error context" table calls for the bundle path. The manifest decoder runs before the bundle path is in scope; the right architectural fix is for `Open` (Task 13) to enrich errors with the bundle path before returning. **Disposition**: revisit in Phase 5 when Open is wired up; spec wording may also need tightening to clarify that `Path` is "where the error originated" (which can be a field name) vs "the bundle file path".

## Non-canonical `entry` returns `ErrPathUnsafe` (Phase 3 gate)

`verifyEntryInFiles` returns `ErrPathUnsafe` (via `Canonicalize`) for a non-canonical `entry`. Spec is ambiguous between this and `ErrManifestInvalid`. **Disposition**: defensible interpretation; document and move on.

## Manifest case-fold uses `strings.ToLower` (Phase 3 gate)

Same Unicode case-fold concern raised in Phase 2. `verifyOneFile` uses `strings.ToLower` for case-fold-duplicate detection; full Unicode case-folding (e.g. `cases.Fold()` from `golang.org/x/text/cases`) would catch German `ß ↔ ss`, Turkish dotless `I ↔ i`, Greek final sigma. **Disposition**: bundled with the existing v1.1 case-folding upgrade.

## Cosmetic minors (Phase 3 gate, deferred)

- `assertCanonicalIntLiteral` is partly dead code (Int64 + JSON grammar already reject most non-canonical literals). Defense in depth; leave as-is.
- Missing `sha256` `Detail` message could distinguish "missing" from "malformed"; current "sha256 not 64-char lowercase hex" is acceptable.
- `assertNoTrailingTokens` `dec.More()` check is partly redundant with the EOF Token check. Belt-and-suspenders; leave.

## Phase 4 ZIP I/O findings (deferred)

### Central-directory ↔ local-header mismatch detection (Phase 4 security gate, C2)

The spec mandates that the central directory and local file headers must agree on filename and uncompressed size to defend against ZIP parser-confusion (smuggling/cloak attacks). Go's `archive/zip` reader exposes only central-directory metadata; checking the local header requires either a custom zip reader or wrapping `archive/zip`'s internals.

**Disposition:** v1.1. Practical risk in v1 is mitigated by hash verification: whatever bytes get decompressed are hashed, and if they don't match the manifest, `ErrHashMismatch` fires. The defense gap is for cross-implementation interop (a non-Go reader could disagree with our reader), but v1 is Go-only per spec § "Conventions".

### Compression-ratio cap (1000:1) not enforced (Phase 4 security gate, C3)

Spec § "Streaming cap enforcement" requires per-file compression-ratio cap of 1000:1. Currently only the absolute uncompressed-size cap (50 MB per file) is enforced. A maximally-compressed entry could expand within the absolute cap.

**Disposition:** v1.1. Implementation requires `f.OpenRaw()` plus manual deflate chain to count compressed input bytes alongside decompressed output. Significant restructure. The 50 MB absolute cap provides DoS protection in the meantime.

### Hardlink / repeated central-dir pointer detection (Phase 4 security gate, I5)

Two central-directory records pointing at the same local-header offset under different names slip through current duplicate-name detection. Defense-in-depth gap related to C2.

**Disposition:** v1.1. Bundled with C2 fix.

### `ErrZipTruncated` over-broadens (Phase 4 gates I6, L3)

Mid-stream EOF, corrupt deflate, malformed central directory, and `ErrInsecurePath` from `archive/zip` all map to `ErrZipTruncated`. Operationally distinct conditions are conflated.

**Disposition:** v1.1 polish. Distinguish via dedicated sentinels (`ErrZipFormat`, `ErrZipPathInsecure`).

### `archive/zip.ErrInsecurePath` handling (Phase 4 security gate, I7)

`zip.NewReader` may return both a reader AND `ErrInsecurePath` (Go version dependent on `GODEBUG=zipinsecurepath`). Current code treats this as fatal-zip; should map to `ErrPathUnsafe`.

**Disposition:** v1.1. Test by setting `GODEBUG=zipinsecurepath=0` in environment.

### CI grep for `verifiedBytes{` literals (Phase 4 crypto gate, M1)

The `verifiedBytes` zero-value `verifiedBytes{}` and one-arg constructor `newVerifiedBytes(buf)` are the only legitimate constructions. A future helper inside `dipx` could accidentally write `verifiedBytes{b: rawBytes}` and bypass `verifyAndReadEntry`. Spec calls for a CI grep at Task 26 final hardening.

**Disposition:** Already in plan as Task 26 step 3. No action needed here.

## Phase 5 gate findings (deferred)

### checkExtraEntries placement (Phase 5 gate, M2)

`checkExtraEntries` is an unmentioned step inserted between manifest-shape (step 4) and hash-verify (step 5). Its placement is defensible (rejects junk before we waste CPU hashing files we'd reject), but the spec's normative ordering (lines 210-225) doesn't enumerate it, and `ErrFileUnexpected` isn't in the spec's error precedence list.

**Disposition:** v1.1 spec clarification. Either add as documented step 4.5 or fold into manifest-shape validation conceptually.

### parseAllWorkflows / detectCycles strictness asymmetry (Phase 5 gate, L2/L3)

`parseAllWorkflows` parses all manifest-listed workflows (stronger than spec post-condition #4 which says "every workflow parses"). `detectCycles` only DFS's from `m.Entry`, missing cycles in unreachable workflows. Inconsistent: the parse pass is broader than necessary, the cycle pass is narrower than necessary.

**Disposition:** v1.1 polish. Both passes should iterate every manifest-listed workflow, OR both should walk only entry-reachable subgraph. Pick one convention. Currently the broader-parse + narrower-cycle behavior is harmless but inconsistent.

## Phase 6 gate findings (deferred)

### dirSource doesn't apply Canonicalize-style validation (Phase 6, H2)

`dirSource.resolveDir` performs filesystem-style validation but doesn't reject NUL bytes, control chars, NFD-encoded paths, or Windows-reserved names. The spec is ambiguous — dirSource paths are filesystem-native and "trusted by virtue of being read from local disk" per the parseDipFile spec note. **Disposition:** v1.1. If a tightening is desired, factor a Canonicalize-equivalent for filesystem paths (which would need different rules — absolute paths are valid, etc.).

### dirSource cache is unbounded with global mutex (Phase 6, M1/M2)

Spec § "Source implementations" mandates "LRU of 256 entries with singleflight.Group for cold-call coalescing." Implementation is unbounded `map[string]*ir.Workflow` guarded by sync.Mutex held across parseDipFile (disk I/O). Two issues: long-running runtime processes leak parsed IR, and concurrent reads on different paths serialize on the global mutex. **Disposition:** v1.1. Drop in `golang.org/x/sync/singleflight` and a small LRU (e.g. `github.com/hashicorp/golang-lru/v2`) — or hand-roll. The singleflight key is the bundle-relative path; the LRU bounds memory.

### Extract atomicity edge cases (Phase 6, M3/M4) — SUPERSEDED by P10.1 resolution

`Extract` previously removed destDir before rename when `--force`, creating a window where neither old nor new existed; if `os.Rename` failed with EXDEV, staging was orphaned. **Resolved in v1.0** as part of the P10.1 fix below — see "Extract --force EXDEV data-loss vector (Phase 10, P10.1) — RESOLVED in v1.0" for the rename-old-aside / rename-new-into-place / remove-aside implementation and the `TestExtract_ForcePreservesDestOnRenameFailure` regression test.

### Source.Workflow doesn't take context (Phase 6, L4)

`Source.Workflow(refPath, relativeTo)` performs disk I/O for dirSource but no ctx. Spec is internally inconsistent: § "Cancellation" line 272 says I/O entry points take ctx, but § "Runtime integration contract" example shows ctx-less Workflow. **Disposition:** v1.1 spec clarification. Decide whether Source is a "fast lookup" or "I/O entry point." If the latter, breaking signature change.

## Phase 7 gate findings (deferred)

### Pack compression-ratio cap (Phase 7, H1 partial)

Spec § "Soft caps" lists 1000:1 compression-ratio as a producer-side limit. Pack only checks per-file uncompressed size (50MB); ratio is unchecked. Implementing requires either tracking compressed bytes during deflate (manual deflate chain + counter) or post-compression check via writer hooks. **Disposition:** v1.1. Per-file 50MB cap provides absolute DoS protection in the meantime.

### Pack parent-directory symlink check (Phase 7, M1) — RESOLVED in v1.0

`walkSourceTree` originally Lstat'd only the leaf file; parent components were silently followed, so a directory symlink in the path tree could silently re-root `rootDir`. **Re-classified as v1-blocking on multi-agent review** (PR #34): two of three external reviewers flagged the original disposition as understating host-file exfiltration risk in CI / contributor-PR / mono-repo build scenarios.

**Resolution:** `readNoFollowSymlinks` now takes `(path, rootDir)` and `assertNoSymlinkAncestor` walks every path component strictly between `rootDir` and the leaf, refusing any symlink along the way. `rootDir` itself is the trust anchor (CLI-supplied, may itself be a user-supplied symlinked working directory) and is not Lstat'd. Regression test: `TestPack_RejectsParentSymlink` in `dipx/dipx_test.go`.

### Pack O_NOFOLLOW (Phase 7, M2)

Spec § "Reproducible Pack" mandates `O_NOFOLLOW` on file open. Current implementation uses `os.Lstat` + `os.ReadFile`, leaving a TOCTOU window if a symlink is swapped between Lstat and ReadFile. **Disposition:** v1.1. Implementing `O_NOFOLLOW` requires platform-specific syscalls (Linux/macOS/BSD have `O_NOFOLLOW`; Windows uses different attributes). The Lstat-first approach mitigates the threat for non-adversarial source trees.

### Pack frozen golden test (Phase 7, L1, L2)

`encodeManifestCanonical` relies on Go's `json.Marshal` field-order behavior. `TestPack_Reproducible` proves in-process determinism but doesn't catch cross-Go-version drift via a frozen byte vector. Spec testing strategy mentions a `well-formed.dipx` golden file. **Disposition:** v1.1. Add as part of Task 25's hardening.

### Pack symlink coverage in tests (Phase 7, L3)

`TestPack_RejectsSymlink` only tests symlink-as-entry. Doesn't test transitive ref symlinks or directory symlinks. **Disposition:** v1.1.

## Phase 8 gate findings (deferred)

### Sentinel-to-exit-code mapping incomplete (Phase 8, M1)

`classifyExit` covers 5 integrity sentinels (HashMismatch, ManifestInvalid, ZipFeatureForbidden, ZipTruncated, UnsupportedFormatVersion) but spec error precedence treats more bundle errors as integrity (FileMissing, FileUnexpected, EntryNotInManifest, RefEscape, RefCycle, CapExceeded, PathUnsafe). Currently these route to user-error 1. **Disposition:** v1.1 spec clarification — decide which sentinels are "integrity" vs "user error" and update both spec table and classifier.

### I/O exit code unreachable in practice (Phase 8, M2) — RESOLVED in v1.0

`exitDipxIOError` (3) was only set by direct `os.Create`/`os.Close`/`os.Rename` failures in packToFile. I/O failures surfacing through `dipx.Pack`/`dipx.Open`/`dipx.Extract` (e.g., `*os.PathError` from walkSourceTree's read) routed to user-error 1 because they were not Canceled and not isIntegrityErr — contradicting the documented "I/O error = 3" contract. **Resolution:** classifyExit was split into `classifyExit` + `classifyDipxErr` to stay under the cyclomatic-5 cap, and a new `isIOErr` helper detects `*os.PathError` / `*os.LinkError` (or fs.ErrNotExist / fs.ErrPermission via `errors.Is`) and routes to exitDipxIOError. Order: cancellation > integrity > I/O > user-error.

### inspect --no-verify advisory in JSON (Phase 8, M4)

When `--no-verify` is requested, only stderr text reflects it. JSON output still says `"status": "VALID"`. **Disposition:** v1.1 polish — promote status to object with `verify_skipped` field.

### inspect text status footer omits byte total (Phase 8, L1)

Spec example: `status: VALID (3 files, 24831 bytes, format_version 1)`. Implementation: `status: VALID (3 files, format_version 1)`. **Disposition:** v1.1 cosmetic.

### inspect JSON status is a bare string, not an object (Phase 8, L2)

Spec line says "manifest plus a status object". Today it's `"status": "VALID"`. **Disposition:** v1.1 — promote to richer object with file count, byte total, etc.

### unpack lacks PATH_MAX + free-space preflight (Phase 8, L3)

Spec § "CLI" mandates these for unpack. Currently absent. **Disposition:** v1.1 — `statfs` + per-platform PATH_MAX checks before any write.

### pack --dry-run + -o - silently discards stdout intent (Phase 8, L4)

Combining flags fails silently. **Disposition:** v1.1 — add diagnostic + exit 1.

### parseFile/loadWorkflow error prefix inconsistency (Phase 8, L5)

Raw `BundleError.Error()` shown without "error:" prefix that other CLI commands use. **Disposition:** v1.1 polish.

## Phase 9 gate findings (deferred)

### OpenLax structured warning not emitted (Phase 9, P9.1)

Spec § "OpenLax discipline": "The function emits a structured warning to the caller-provided logger (when context.Context carries one via standard convention) on every invocation." Implementation does not emit any warning. **Disposition:** v1.1. Add ctx-key-based logger lookup or simple stderr warning gated by env var.

### ZIP entry case-fold uses ASCII strings.ToLower (Phase 9, P9.2)

Same issue as manifest case-fold (already-deferred Phase 2/3): `dipx/zipio.go` `recordEntry` uses `strings.ToLower(f.Name)` for case-fold-duplicate detection. Should use Unicode `cases.Fold()`. **Disposition:** v1.1, bundled with manifest-layer case-fold upgrade.

### ZIP entry canonicalization at admission (Phase 9, P9.3)

`cz.entries` stores by raw `f.Name` without applying `Canonicalize`. Non-canonical entries (e.g., `workflows//a.dip`) admitted then rejected later by strict-mode check. Practical effect: same rejection, but error class is `ErrFileUnexpected` rather than `ErrPathUnsafe`. **Disposition:** v1.1 polish — canonicalize at admission for cleaner error attribution.

### DIPX_DEBUG=1 diagnostic mode not implemented (Phase 9, P9.4)

Spec § "Operational ergonomics / Diagnostic mode": env-gated structured step trace. Currently no implementation. **Disposition:** v1.1. Add a small package-level tracer gated on `DIPX_DEBUG=1`. The "no default logging" rule is preserved because tracing only fires under explicit env var.

### Consumer deployment limits not configurable (Phase 9, P9.5)

Spec § "Soft caps": "A conformant reader MAY enforce stricter limits configured for its deployment context." Implementation hardcodes conformance caps. No `WithMaxFiles(n)`/`WithMaxBytes(n)` options. **Disposition:** v1.1. Add functional options on `Open`/`OpenReader` for deployment-stricter caps.

### CLI flag-position: flags must precede positional args (Phase 9, P9.6)

Go's `flag` package stops parsing at the first non-flag argument. Users running `dippin pack examples/foo.dip -o out.dipx` get usage errors; must be `dippin pack -o out.dipx examples/foo.dip`. Common Unix tools accept both orders. **Disposition:** v1.1. Either switch to `pflag` / `cobra` for any-position flag parsing, or document the constraint in `--help` output.

## Phase 10 final-gate findings (deferred)

### Extract --force EXDEV data-loss vector (Phase 10, P10.1) — RESOLVED in v1.0

**Re-classified as v1-blocking on multi-agent review** (PR #34): the original "v1.1 polish" disposition undersold a real data-loss vector triggered by routine cross-mount staging (TMPDIR on a separate APFS volume / tmpfs / CI work mount). Two of three external reviewers flagged it as v1-blocking.

**Resolution:** `Extract` now uses a `swapDestWithStaging` helper that performs the rename-old-aside / rename-new-into-place / remove-aside sequence. `swapWithBackup` renames `destDir → destDir.bak` first, then `staging → destDir`; on failure it restores `destDir.bak → destDir` so the user's original is preserved. Stale staging is removed via `defer`. Regression test: `TestExtract_ForcePreservesDestOnRenameFailure` in `dipx/dipx_test.go` injects a synthetic EXDEV at the second rename and asserts destDir is intact.

### Pack lacks mid-write ctx checks (Phase 10, P10.2)

`Pack` checks ctx between preparePackManifest and writeBundle, but `writeBundle`'s zip-writer loop has no ctx check. Asymmetric with Open (which checks ctx between every CPU-bound stage). **Disposition:** v1.1 polish.

### inspect JSON encoder error silently swallowed (Phase 10, P10.3) — RESOLVED in v1.0

`printInspectJSON` returns exitDipxIOError on `enc.Encode` failure but doesn't print the error to stderr. Operator gets exit 3 with no diagnostic. **Resolution:** signature now takes a `stderr io.Writer`; encode failures are surfaced via `fmt.Fprintln(stderr, err)` before returning the I/O exit code.

### inspect --no-verify is a no-op (Phase 10, P10.4)

The flag is parsed and prints a stderr warning, but `dipx.Open` is invoked unconditionally. Forensically-suspect bundles can't be inspected without triggering integrity errors. **Disposition:** v1.1. Either implement a no-verify code path or remove the flag.

### Extract writeOneFile lacks O_EXCL|O_NOFOLLOW (Phase 10, P10.5)

`writeOneFile` uses `os.WriteFile` without `O_EXCL`/`O_NOFOLLOW`. A hostile process running as the same user could race a symlink into the freshly-created staging dir between `MkdirAll` and `WriteFile`. Practical risk is low (single-user attacker scenario) but spec § "Reproducible Pack" mandates O_NOFOLLOW; same hardening applies to Extract. **Disposition:** v1.1, bundled with Pack O_NOFOLLOW.

### Pack-side BundleError.Path uses filesystem absolute paths (Phase 10, P10.6)

Spec § "Per-sentinel error context" says Path is "bundle-relative" but Pack errors legitimately can't be bundle-relative (no bundle yet). Defensible interpretation; spec wording should be tightened in v1.1.

## PR #34 review-cycle findings (deferred)

### walkSourceTree lacks ctx checks (Phase 10, P10.7 — extends P10.2)

Phase 10 P10.2 covers `writeBundle`'s zip-writer loop, but `walkSourceTree`'s `visitNext` / `readAndRecord` / `enqueueRefs` also have no ctx checks. A long Pack against a deep source tree cannot be cancelled until the walker finishes. Fix: `case <-ctx.Done()` at the top of `visitNext`. **Disposition:** v1.1, bundled with P10.2.

### Pack reads source files in full before per-file cap check (Phase 10, P10.8)

`readNoFollowSymlinks` calls `os.ReadFile` (full materialization) before `readAndRecord`'s `len(raw) > maxPerFileBytes` check, so an oversized source file fully allocates before being rejected. Producer-side risk only (Pack inputs are local trusted developer source trees, not adversarial bytes). Fix: pre-check `info.Size()` from `os.Lstat`, or switch to bounded read via `io.LimitedReader`. **Disposition:** v1.1.

### parsePackSource always classifies as ErrEntryParse (Phase 10, P10.9)

`parsePackSource` is invoked from `readAndRecord` for both the entry workflow AND every transitively-reachable subgraph. Subgraph parse failures during Pack therefore get `ErrEntryParse` (which structurally implies the entry, not a subgraph). Fix: thread an `isEntry` bool into `parsePackSource`, or check `path == s.entryAbs` at the caller and pick `ErrSubgraphParse` for refs. Cosmetic for Pack-side error attribution. **Disposition:** v1.1.

### verifyAllHashes lacks per-entry ctx (Phase 10, P10.10)

`verifyAllHashes` checks ctx once before starting (via `verifyHashesCtx`), then runs the inner loop without further ctx checks. A bundle with many large entries cannot be cancelled mid-verification. Bundle this with the broader Pack/walker ctx threading work. **Disposition:** v1.1, bundled with P10.2/P10.7.

### Justfile pack-examples uses `mktemp -u` and skips round-trip (Phase 10, P10.11)

The `pack-examples` recipe in `Justfile` uses `mktemp -u` (TOCTOU race) and never cleans up the generated `.dipx` files. The recipe also doesn't actually verify a round-trip — only `dippin pack` runs, no `inspect` or `unpack`. Practical risk is near-zero (CI/dev-loop recipe, parallel pack-races unlikely). **Disposition:** v1.1 polish — `mktemp -d` with cleanup + a `dippin inspect`/`unpack` verification step.
