# `.dipx` — Distributable bundle format for Dippin workflows

**Status:** Design v3 — revised 2026-05-07 after second-round multi-reviewer pass (DoS, ops, crypto discipline, PAL)
**Primary use case:** Distribution / portability — share a self-contained pipeline by file, URL, or registry; recipient runs it without needing the original repo.
**Library-first:** A `dipx` Go package both the dippin CLI and downstream consumers import directly.

## Conventions

The capitalised keywords MUST, MUST NOT, REQUIRED, SHOULD, SHOULD NOT, MAY are used in the normative sense from RFC 2119/8174. Lowercase usage is descriptive prose.

This document is a **Go library + format specification** for the dippin-lang reference implementation. Cross-language conformance (test vectors, language-neutral pseudocode, alternative-implementation interop) is **out of scope for v1**; see *Known v1 limitations*.

## Problem

A Dippin workflow's only external dependencies today are other `.dip` files referenced via `subgraph ref:` and `manager_loop subgraph_ref:`. Prompts, system prompts, and JSON response schemas are inline strings in the IR (`ir/ir.go:90-109`). There is no `@file:` / include syntax.

To share a multi-file pipeline today, you ship a directory tree and hope the recipient preserves layout. There is no:

- single artifact to email, host, or version
- integrity check on the bundle
- guarantee that a recipient won't unwittingly execute a workflow whose subgraph escaped the source root
- shared loading contract between `dippin` (authoring) and the runtime (execution)

`.dipx` is that artifact.

## Scope

In scope (v1):

- Single-file ZIP container with extension `.dipx`.
- Bundling rules: parent `.dip` plus every transitively-referenced subgraph, plus a JSON manifest with SHA-256 per file.
- Hermetic invariant: refs inside the bundle resolve only to manifest-listed, hash-verified files.
- Library API in package `dipx`: `Pack`, `Open`, `OpenReader`, `OpenLax`, `Validate`, `Extract`, `Load`, plus a `Source` interface and `Bundle` type — all I/O operations accept `context.Context`.
- CLI: new `dippin pack` / `unpack` / `inspect`; existing analysis commands transparently accept `.dipx`. CLI uses atomic file-write patterns (write-temp + rename) for both Pack output and Extract output.
- Runtime integration: two call-site changes; identical execution semantics.
- Errors: typed `*BundleError` plus Go error sentinels in `dipx`; no DIP codes for bundle errors.
- Reproducible Pack: same source tree → byte-identical bundle.
- Bundle identity defined as SHA-256 of manifest bytes (forward-compatible with v2 detached signatures).

Out of scope (see *Known v1 limitations*):

- Cryptographic signatures (sketch included for v2 alignment).
- External asset bundling.
- Bundle-level budget aggregation.
- Distribution mechanism — networking, registries, caching.
- Cross-language conformance.
- Hash algorithm agility.
- `Bundle.Compact()` / `OpenLite` memory-shedding mode.
- Process-wide memory accountant.

## Established design decisions

| Decision | Choice | Why |
|---|---|---|
| Container | ZIP (constrained subset) | Stdlib `archive/zip`; seekable; universal `unzip` for forensics; matches `.whl`/`.jar` precedent. |
| Bundle scope | Workflows + manifest only | No external assets in the language today; manifest extensible via tolerant decoding. |
| Integrity | SHA-256 per file in manifest | Detects accidental corruption and casual tampering between trusted parties. **Does not authenticate origin** — see *Threat model*. |
| Signatures | Deferred to v2 (sketch in *Versioning*) | Don't gate v1 adoption on key-management; v1 design must not preclude v2. |
| Strict mode | **Default**: extra zip *file* entries rejected; directory entries always ignored | Executable artifacts must be strict by default. `OpenLax` is the explicit escape hatch. |
| Hermetic | Refs cannot escape `workflows/` AND must resolve to manifest-listed entries | A `.dipx` runs standalone or not at all. Manifest tampering cannot trick the runtime into serving extras. |
| Versioning | Integer `format_version`, fail-closed on unknown | Executable artifact; "warn and try" is unsafe. |
| Reproducibility | Pack output is deterministic | Same source tree → byte-identical `.dipx`. Fixed timestamps; sorted entries; no platform-specific metadata. |
| Bundle identity | SHA-256(manifest.json bytes-as-stored) | Manifest is the authoritative provenance object; zip-byte order is malleable. |
| Cancellation | `context.Context` on every I/O entry point | Required to bound CPU and time on adversarial input. |

## Threat model

`.dipx` v1 is designed for distribution **between trusted parties** over channels where integrity matters but authenticity is asserted out-of-band (e.g., a developer hands a teammate a bundle; a CI artifact moves between trusted services). SHA-256 in the manifest detects accidental corruption and casual tampering by parties who do not control the channel.

**`.dipx` v1 does NOT defend against an active attacker who controls the distribution channel.** Such an attacker can rewrite both the bundle bytes and the manifest hashes. Authenticity defense is the role of cryptographic signatures, deferred to v2 (sketch in *Versioning*).

**Cross-bundle replay is undefended in v1.** Two `.dipx` files with byte-identical manifest contents are indistinguishable to a v1 reader. Operators tracking provenance MUST record manifest digests (`SHA-256(manifest.json bytes)`), not bundle filenames, and MUST NOT compare bundle file digests (zip ordering is malleable under valid re-encoding).

**Hash comparison need not be constant-time.** The threat model is integrity in transit, not authentication. SHA-256 comparison against a known manifest value has no adaptive chosen-ciphertext analog. `bytes.Equal` is acceptable. Implementations MUST NOT introduce `crypto/subtle` comparisons here, as doing so obscures the actual threat model.

`OpenLax` weakens the strict-mode trust boundary and MUST NOT be invoked on bytes obtained from any non-local source (network, untrusted user upload, IPC from un-trusted peer). See *OpenLax discipline*.

Producers and consumers MUST treat `.dipx` files received from untrusted sources as un-authenticated until v2 ships signatures.

## Wire format

### Container: ZIP, constrained

A `.dipx` file is a ZIP archive (PKZIP APPNOTE.TXT). The following ZIP features are restricted to keep the format predictable and resistant to parser-confusion attacks:

- Entries MUST use compression method `Store` (0) or `Deflate` (8). Other methods MUST be rejected with `ErrZipFeatureForbidden`.
- Encryption (any form, including ZipCrypto and AES) MUST NOT be used. Encrypted entries MUST be rejected.
- Multi-disk / spanned archives MUST NOT be used and MUST be rejected.
- ZIP64 records MUST be used when, and only when, required by entry size or count (per APPNOTE.TXT).
- The ZIP archive comment field SHOULD be empty; if present, MUST be ignored.
- Per-file extra fields are permitted but MUST NOT alter file content; Pack MUST NOT emit any (see *Reproducibility*).
- Entry filenames MUST be encoded as UTF-8 with general-purpose bit 11 set (APPNOTE.TXT 4.4.4). CP437-encoded filenames MUST be rejected.
- An entry whose external file attributes encode a symlink (Unix mode bit `S_IFLNK`, `0o120000`) MUST be rejected.
- An entry whose external file attributes encode a non-regular, non-directory file (device, FIFO, socket) MUST be rejected.
- The central directory and local file headers MUST agree on filename and uncompressed size; mismatches MUST be rejected (defense against parser-confusion). **[Deferred to v1.1 — Phase 4 C2.](../plans/2026-05-07-dipx-followups.md) v1 mitigates by per-file SHA-256: any cross-record disagreement on bytes-as-decompressed will fail the manifest hash.**
- Duplicate entry names (case-sensitive byte equality of canonical paths) MUST be rejected.
- Two entries whose canonical paths case-fold to the same value (Unicode case-folding, not just ASCII tolower) MUST be rejected. **[Unicode case-folding deferred to v1.1 — Phase 9 P9.2.](../plans/2026-05-07-dipx-followups.md) v1 uses `strings.ToLower` (ASCII-only).**
- **Directory entries** (entries whose name ends with `/`) MUST be ignored on read in both `Open` and `OpenLax` modes. Pack MUST NOT emit them. Rationale: most ZIP producers emit directory entries by default; rejecting them is a trivial compatibility-DoS lever and adds no security benefit since strict mode is anchored by `files[]` byte-comparison plus hash verification, not by zip-entry inventory.
- A truncated zip (mid-stream EOF or central-directory missing) MUST be rejected with `ErrZipTruncated` and MUST NOT be silently coerced into `ErrHashMismatch`.
- Open MUST NOT enforce zip entry ordering. (Reproducibility is producer-side; receivers cannot rely on bundle-byte-equality.)

### Bundle layout

```text
manifest.json              # at zip root, exactly this name (lowercase ASCII)
workflows/                 # mirrors original directory structure
  api_design.dip
  interview_loop.dip
  phases/
    code_review.dip
```

`manifest.json` MUST be a zip entry named exactly `manifest.json`, at the archive root, with no leading directory.

The entry name `manifest.sig` is **reserved** in v1 for the future v2 detached-signature mechanism. Pack MUST NOT emit it; Open MUST reject its presence in v1 with `ErrZipFeatureForbidden` (treating it as a forward-compat marker that v1 cannot honor).

No `assets/` directory in v1. Adding one in a future format version is non-breaking provided the producer also bumps `format_version`.

### Manifest schema

```json
{
  "format_version": 1,
  "entry": "workflows/api_design.dip",
  "files": [
    { "path": "workflows/api_design.dip",         "sha256": "abc123…" },
    { "path": "workflows/interview_loop.dip",     "sha256": "def456…" },
    { "path": "workflows/phases/code_review.dip", "sha256": "789abc…" }
  ]
}
```

Three required keys.

#### Schema rules

1. **`format_version`** MUST be a non-negative integer JSON literal in `[1, 2^31-1]`. Decoders MUST use `json.Decoder.UseNumber()` and parse via `json.Number.Int64()` to avoid float64 silent rounding at high values. The values `1.0`, `1e0`, `"1"`, leading-zero forms, and negative values MUST be rejected. v1 accepts only `1`.
2. **`entry`** MUST be a path string that byte-equals exactly one `files[].path`. It MUST start with `workflows/` and end with `.dip`. Canonicalization (per *Path canonicalization*) is applied during manifest validation, before byte-equality comparison.
3. **`files[]`** MUST be a non-empty JSON array of objects. Each object MUST have exactly the required keys `path` (string) and `sha256` (string). Unknown keys inside the object are tolerated (silently ignored). Non-object members or members missing required keys MUST cause `ErrManifestInvalid`.
4. **`files[].path`** MUST start with `workflows/`, MUST end with `.dip`, MUST be byte-equal to its canonical form (per *Path canonicalization*), and MUST be unique across the array.
5. **`files[].sha256`** MUST be lowercase hex, exactly 64 characters, computed over the **uncompressed** (logical) bytes of the file as it would be written to disk by `Extract`. Uppercase, non-hex, leading/trailing whitespace, and incorrect length MUST be rejected as `ErrManifestInvalid` **during step 1 (manifest validation), never as `ErrHashMismatch`**.
6. **Reserved keys**: `signatures` is reserved at the top level for v2 detached-signatures support. v1 readers MUST reject the presence of `signatures` with `ErrManifestInvalid`. (The "reserved-but-tolerated" pattern is a known downgrade-attack vector — v2-signed bundles flipped to `format_version: 1` would silently lose authenticity. Reject in v1; v2 readers will require it.)

#### Manifest JSON encoding

The manifest MUST be UTF-8 encoded. A leading byte-order-mark MUST NOT be emitted by Pack and MUST be rejected on Open. A trailing newline is OPTIONAL.

The decoder MUST reject:

- Duplicate top-level keys (Go's `encoding/json` accepts last-wins by default; implementations MUST do a token-based pre-pass that rejects duplicates).
- Duplicate keys inside any object (including `files[]` members).
- Trailing data after the top-level JSON object.
- Manifests larger than 1 MB before parsing (`io.LimitReader`-bounded).
- Nesting depth > 32 (defense against stdlib JSON CVEs that recurse on the goroutine stack).

The decoder MUST tolerate (silently ignore):

- Unknown top-level keys (extensibility — see *Forward compatibility*) **except** the reserved `signatures` key, which MUST be rejected.
- Unknown keys inside objects in `files[]`.
- Whitespace, line endings, and indentation choices.

JSON key ordering inside the manifest is not significant for v1 semantics. Pack writes a canonical key order (alphabetical at every level) and a sorted `files[]` (by `path` byte-order) to support reproducibility.

### Path canonicalization

The following rules apply identically in `Pack`, `Open`, `Source.Workflow`, and `Extract`. The Go reference implementation expresses them via a single exported function `dipx.Canonicalize(path string) (string, error)`. **All four sites MUST call this single function**; CI MUST grep-check that no call site within `dipx` uses `path.Clean` or `filepath.Clean` outside that function.

A path is in **canonical form** if and only if all of:

1. It is a valid UTF-8 sequence in Unicode Normalization Form C (NFC).
2. Separators are forward-slash `/` only. Backslash `\` MUST be rejected.
3. There are no leading `./` segments, no `..` segments, no empty path components (`//`), no leading `/`.
4. It contains no NUL byte (U+0000), no ASCII control characters (< 0x20), no DEL (0x7F).
5. No path component has leading or trailing whitespace.
6. No path component, after stripping case and extension, equals a Windows reserved name: `CON`, `PRN`, `AUX`, `NUL`, `COM0`–`COM9`, `LPT0`–`LPT9`.
7. No path component ends in `.` or has a trailing space (Win32 strips these silently and creates collisions).
8. The path component count is ≤ 16 and total path length is ≤ 1024 bytes.
9. The path ends in `.dip` (lowercase, byte-exact). Other extensions and case variants MUST be rejected.

Path comparison is case-sensitive byte-equality of the canonical form. Pack MUST additionally reject source trees containing two paths whose canonical forms case-fold to the same value (Unicode case-folding).

The Unicode normalization library version MUST be pinned in `go.mod`. Bumping it requires regression testing of canonicalization output across all rules and SHOULD be coordinated with `format_version` planning.

### Subgraph ref resolution

A *ref* is the value of a `subgraph ref:` or `manager_loop subgraph_ref:` field in a workflow source. Refs MUST be static string literals; refs containing variable interpolation, expressions, or other dynamic forms are unpackable and Pack MUST reject them.

Given a parent workflow at bundle-relative path `relativeTo` and a ref string `refPath`, the resolved bundle-relative path is computed as follows (pseudocode):

```text
resolved := lexical-clean(join(directory(relativeTo), refPath))
```

Where `directory(p)` returns `p` up to and including its last `/` (or empty if none); `join(a, b)` concatenates with a single `/` separator unless `a` is empty; `lexical-clean` collapses redundant `./`, resolves `..` relative to earlier components without consulting the filesystem, and normalizes redundant `/`.

The resolved path MUST then satisfy *all* of:

- It is in canonical form (per *Path canonicalization*).
- It starts with `workflows/`.
- It byte-equals exactly one `files[].path` in the manifest.

Use of `..` *inside* `refPath` is permitted as long as the resolved path satisfies the above (e.g., `ref: ../sibling/foo.dip` from `workflows/sub/parent.dip` resolving to `workflows/sibling/foo.dip` is legal). Use of `..` *anywhere in `files[].path` or `entry`* MUST be rejected (manifest paths must already be canonical).

### Hash computation and verification

Each `files[].sha256` MUST be the SHA-256 digest of the **uncompressed** (logical) bytes of the file. This is the byte sequence that `Extract` would write to disk.

Hash binds to whole logical files. Sub-range hashing, Merkle trees over chunks, or partial-content hashing are out of scope for v1 and would require a `format_version` bump.

#### Open ordering (normative)

`Open` (and `OpenReader`/`OpenLax`) MUST execute the following steps in order:

1. **Open zip**; locate `manifest.json`; reject duplicate entry names; verify central-directory ↔ local-header agreement; cap-check manifest entry size (≤ 1 MB).
2. **Read manifest bytes** with a length-bounded reader; record `manifestDigest = SHA-256(manifest.json bytes-as-stored)` for later use as bundle identity.
3. **Decode manifest** with token-based duplicate-key rejection, depth cap 32, BOM rejection, `json.Number` for `format_version`. Validate `format_version` against `SupportedFormatVersions()` — unknown → `ErrUnsupportedFormatVersion`.
4. **Validate manifest shape**: every `files[].path` is canonical, every `files[].sha256` is well-formed (length, hex, lowercase), `entry` matches a `files[].path` byte-exact, no duplicate paths or case-fold-equal paths, reserved key `signatures` not present, paths within caps. Failures → `ErrManifestInvalid`.
5. **Verify no extra zip entries**: enumerate non-directory zip entries; reject any whose canonical path does not appear in `files[]` (`ErrFileUnexpected`). This step rejects junk before step 6 spends CPU hashing files the bundle would reject anyway. Implementations MAY fold this check into step 4's manifest-shape validation as long as the externally observable error precedence is preserved.
6. **Verify hashes**: for each `files[]` entry: locate the zip entry by canonical path; stream-decompress its bytes through a length-bounded reader (per-file size cap and ratio cap enforced *during* decompression, not after) into a fresh `[]byte`; compute SHA-256; compare to manifest. Mismatch → `*BundleError{Sentinel: ErrHashMismatch, Path: path, Detail: "expected: X, actual: Y"}`. Cap-check accumulated total uncompressed size during this loop.
7. **Parse**: only after step 6 verifies all hashes, parse each `.dip` file via `parser.NewParser(string(verifiedBytes), path).Parse()`. Implementations MUST ensure the bytes presented to `parser.NewParser` come from the same `[]byte` allocation produced by step 6; re-reading from the zip MUST be a structural impossibility (encoded in the type signature; see *Type-encoded ordering*). Failures → `ErrEntryParse` for entry; `ErrSubgraphParse` for subgraphs.
8. **Walk refs** with tri-color DFS for cycle detection rooted at **every manifest-listed workflow** (not only `entry`); reject cycles (`ErrRefCycle`), refs that escape `workflows/` (`ErrRefEscape`), and refs that resolve to paths not in `files[]` (`ErrFileMissing`). Walking from every manifest-listed workflow keeps cycle detection symmetric with step 7's parse pass, which already parses every workflow.
9. **Normalize**: for each parsed workflow, invoke `simulate.EnsureConditionsParsed` to populate `Condition.Parsed` on every edge condition and every `manager_loop` `StopCondition`/`SteerCondition`. Done eagerly while no other goroutines hold the workflow, to prevent runtime races on shared `*ir.Workflow` values.
10. **Build** the immutable in-memory `Bundle` with pre-sized maps (`make(map[string]*ir.Workflow, len(files))`) and return.

Bytes read in step 6 are the bytes used for parsing in step 7. **There is no second read from the zip** and no TOCTOU.

Implementations MUST close all file handles and zip-reader internals on every exit path from `Open`/`OpenReader`/`OpenLax`, **success and error alike**. `defer` statements covering `*os.File` and `*zip.Reader` are the recommended pattern.

#### Type-encoded ordering

The Go reference implementation MUST encode steps 6 and 7 such that step 7 cannot bypass step 6 as a structural matter, not just a documentation matter:

- Step 6 returns `map[string]verifiedBytes` where `verifiedBytes` is an unexported wrapping type.
- Step 7 takes `map[string]verifiedBytes`. It has no `*zip.Reader` parameter.
- The Open pathway invokes `parser.NewParser` from exactly one site, and that site takes its input from `verifiedBytes.Bytes()`. The `dipx` package as a whole has TWO additional `parser.NewParser` sites that consume trusted local-disk bytes: the dirSource pathway (`parseDipFile` in `source.go`) and the Pack pathway (`parsePackSource` in `helpers.go`). Neither produces nor consumes `verifiedBytes`; the type-encoded invariant binds only the Open pathway, which is the only one where bytes originate from a `.dipx` archive.
- A CI test (`TestInvariant_ParserNewParserSiteCount`) pins the total to exactly three sites: the Open verifiedBytes site, the dirSource site, and the Pack site. Any drift fails the test.

This makes "parse before verify" or "re-read from zip after verify" on the Open pathway a compile-time error, not a documentation invariant.

### Streaming cap enforcement

Caps MUST be enforced via streaming, not by trusting ZIP header `UncompressedSize64` fields and not by post-buffering the full decompressed output. Specifically:

- The per-file uncompressed cap (50 MB) MUST be enforced via `io.LimitReader` wrapping the decompressor. Excess bytes trip the cap before allocation.
- The per-file compression-ratio cap (1000:1) MUST be checked while reading: track compressed bytes consumed against decompressed bytes produced; abort once the ratio exceeds 1000. **[Deferred to v1.1 — Phase 4 C3.](../plans/2026-05-07-dipx-followups.md) v1 enforces only the absolute 50 MB per-file uncompressed cap.**
- The total uncompressed cap (100 MB) MUST be enforced as a running sum across files; abort the moment a step-6 read crosses the threshold, before allocating the offending file's full buffer.
- The 1 MB manifest cap MUST be enforced before JSON parsing begins.

Implementations MUST NOT fully buffer an entry into a `[]byte` and then check size — they MUST refuse to allocate beyond the cap.

### Soft caps (split: producer / consumer)

`.dipx` distinguishes between **bundle conformance limits** (producer-side) and **consumer-deployment limits** (reader-side):

**Bundle conformance limits** — a conformant bundle MUST NOT exceed:

- 10,000 files in `files[]`.
- 100 MB total uncompressed size.
- Ref-graph depth: 64.
- Per-workflow node count: 1,000.
- Per-node fan-out (outgoing subgraph refs): 256.
- Per-file uncompressed size: 50 MB.
- Per-file compression ratio: 1000:1.
- Manifest size: 1 MB.
- JSON nesting depth: 32.

**Consumer deployment limits** — a conformant reader MUST accept any bundle within the conformance limits AND MAY enforce stricter limits configured for its deployment context. When stricter limits trip, the reader MUST emit `ErrCapExceeded` with the cap name and observed value.

The conformance limits are intentionally generous to support today's production workflows. **Operators SHOULD configure stricter consumer-side limits for production deployments** (e.g., 10 MB total for memory-constrained runtime pods). The 100 MB conformance maximum is *not* a deployment recommendation.

### Cancellation and timeouts

All I/O entry points (`Open`, `OpenReader`, `OpenLax`, `Pack`, `Extract`, `Validate`) accept `context.Context` as their first parameter. Implementations MUST honor cancellation by:

- Checking `ctx.Err()` between each step of the Open ordering.
- Wrapping zip readers and decompressors in cancellation-aware reader chains so a cancelled context aborts the in-progress step within bounded time.
- Returning `ctx.Err()` (typically `context.Canceled` or `context.DeadlineExceeded`) without further wrapping, so callers can use `errors.Is(err, context.Canceled)` directly.

Callers fetching bundles from non-local sources (HTTP, network filesystems, slow disks) MUST pass a deadline-bearing context. The library does not impose a default timeout — that is the caller's policy.

### Reproducible Pack

`Pack(ctx, sameSource, w)` invoked twice MUST produce byte-identical bytes via `w`. Specifically:

- All ZIP entry mtimes MUST be set to the ZIP epoch `1980-01-01T00:00:00Z`.
- Entry order in the central directory MUST be lexicographic by canonical path, with `manifest.json` first.
- ZIP extra fields MUST NOT be emitted (no Unix UID/GID, no NTFS timestamps, no Info-ZIP Unicode path).
- File modes in external attributes MUST be `0644` for files. Directory entries are not emitted.
- The manifest is canonicalized (alphabetical keys at every object level; `files[]` sorted by `path`).
- Pack MUST NOT include OS-specific metadata files (`__MACOSX/*`, `.DS_Store`, `Thumbs.db`, `desktop.ini`, AppleDouble `._*` files, `.gitkeep`/`.gitignore` from source trees). Pack walks the source tree and only emits transitively-reachable workflow files.
- Pack MUST NOT follow filesystem symlinks anywhere in the source tree. Encountering a symlink during walk is an error. Files MUST be opened with `O_NOFOLLOW` where the OS supports it (Linux, macOS, BSD; on Windows the equivalent attribute check applies).
- Pack reads each source file exactly once: the same `[]byte` produces both the SHA-256 digest stored in the manifest and the bytes written to the zip. There is no Stat-then-Open TOCTOU window.

Open MUST NOT enforce zip ordering on read; reproducibility is a producer-side property. Receivers tracking provenance use `manifestDigest`, not the bundle file digest (zip-byte ordering is malleable under valid re-encoding).

## Library API (Go reference implementation)

### Package layout

```text
dipx/
  dipx.go        # Public API: Pack, Open, OpenReader, OpenLax, Validate, Extract, Load
  manifest.go    # Manifest type, JSON encoding/decoding with strict rules
  resolve.go     # Path canonicalization (single Canonicalize function), ref walking, cycle detection
  zipio.go       # Constrained zip reader/writer; verifiedBytes wrapping type
  errors.go      # BundleError type + sentinels
  helpers.go     # Helper decomposition for Open/Pack to satisfy 5/7 complexity caps
  testdata/
  *_test.go
```

`dipx` imports `ir`, `parser`, **and `simulate`**. The `simulate` import is required because step 9 invokes `simulate.EnsureConditionsParsed` to render returned workflows ready for execution. This extends the architectural rule in `CLAUDE.md` ("Packages import `ir` but not each other"). `CLAUDE.md` MUST be amended in the same change to declare a "loader" tier, parallel to the existing "analysis" tier exemption: `dipx` may compose `ir + parser + simulate`. The exemption is bounded — `dipx` MUST NOT import `validator`, `cost`, `formatter`, or any other analysis package.

`Open` and `Pack` are decomposed into helpers in `helpers.go` to keep each function under cyclomatic 5 / cognitive 7. Indicative decomposition:

```text
Open  → openZip → readManifest → decodeManifest → verifyManifestShape →
        verifyHashes → parseAllWorkflows → walkRefs → normalizeConditions → buildBundle
Pack  → walkSourceTree → resolveRefs → buildManifest → writeBundle
```

Each helper is a thin step with one purpose; the top-level functions are sequence orchestrators with `if err != nil` plumbing only.

The `dipx` package emits **no log output**. All observability is via returned errors and (future) optional tracer hooks. Implementations MUST NOT call `log.Printf`, `slog.Info`, or any package-level logger from production paths.

### Types

```go
package dipx

// Manifest is the parsed manifest.json. Returned by Bundle.Manifest() as a value.
type Manifest struct {
    FormatVersion int
    Entry         string
    Files         []ManifestEntry
}

// ManifestEntry is one record in Manifest.Files.
type ManifestEntry struct {
    Path   string
    SHA256 string
}

// SupportedFormatVersions returns the format_version values this build accepts.
// Returns a fresh slice on every call to prevent mutation by callers.
func SupportedFormatVersions() []int { return []int{1} }

// BundleError wraps a sentinel with structured context. Use errors.Is for sentinel
// discrimination and errors.As to extract structured fields.
type BundleError struct {
    Sentinel error  // one of the package-level sentinels
    Path     string // bundle-relative path, or filesystem path for Pack/Extract
    Detail   string // human-readable specifics
    Cause    error  // underlying error (e.g., parser error for ErrEntryParse)
}

func (e *BundleError) Error() string
func (e *BundleError) Is(target error) bool { return target == e.Sentinel }
func (e *BundleError) Unwrap() error        { return e.Cause }

// Source loads workflows, whether from a .dip on disk (refs resolved against
// the filesystem) or from a .dipx bundle (refs resolved against the bundle root).
//
// Argument order matches flatten.Resolver.Resolve(refPath, relativeTo) for
// codebase consistency.
//
// Source is safe for concurrent reads. Returned *ir.Workflow values MUST be
// treated as read-only by callers.
type Source interface {
    Entry() *ir.Workflow
    Workflow(refPath, relativeTo string) (*ir.Workflow, error)
}

// Bundle is an opened .dipx. All workflows are parsed and normalized eagerly
// on Open; no file handles are held after Open returns. Bundle implements Source
// and is immutable post-Open.
type Bundle struct {
    // unexported state
}

// Manifest returns a copy of the parsed manifest. Cost is O(len(Files)); for
// hot paths, callers SHOULD cache the return value. Callers may mutate the
// returned value without affecting the bundle.
func (b *Bundle) Manifest() Manifest

// Identity returns the SHA-256 of manifest.json bytes-as-stored. This is the
// authoritative bundle identity for provenance tracking. Two bundles with
// equal Identity are equivalent regardless of zip-byte ordering.
func (b *Bundle) Identity() [32]byte

func (b *Bundle) Entry() *ir.Workflow
func (b *Bundle) Workflow(refPath, relativeTo string) (*ir.Workflow, error)
func (b *Bundle) Lookup(bundlePath string) (*ir.Workflow, error)
func (b *Bundle) Resolve(refPath, relativeTo string) (string, error)
func (b *Bundle) ReadFile(bundlePath string) ([]byte, error)
```

### Constructors

```go
// Pack builds a .dipx from an entry .dip on disk and writes it to w.
// Walks every transitively-reachable subgraph/manager_loop ref. Validates
// structurally, applies all path-safety and ZIP-feature constraints, and
// produces a deterministic byte stream. Returns the resulting Manifest for
// caller logging/inspection.
//
// ctx cancellation aborts the walk and write within bounded time.
func Pack(ctx context.Context, entryPath string, w io.Writer) (Manifest, error)

// Open reads a .dipx from disk in strict mode (the default). Strict mode
// rejects any zip *file* entry not listed in the manifest. Directory entries
// are always ignored. Pack-produced bundles always pass strict mode.
//
// ctx cancellation aborts in-progress decompression / hash / parse steps.
func Open(ctx context.Context, path string) (*Bundle, error)

// OpenReader is Open from any io.ReaderAt of known size. Callers reading from
// network sources MUST pass a deadline-bearing ctx.
func OpenReader(ctx context.Context, r io.ReaderAt, size int64) (*Bundle, error)

// OpenLax is Open with extra zip *file* entries silently tolerated. For
// hand-edited bundles only. NEVER call OpenLax on bytes obtained from any
// non-local source. See OpenLax discipline.
func OpenLax(ctx context.Context, path string) (*Bundle, error)

// Validate is Open-and-discard. Provided for symmetry with the dippin CLI;
// library callers should typically use Open directly.
func Validate(ctx context.Context, path string) error

// Extract unpacks a .dipx into destDir. Atomic: writes to destDir+".tmp" and
// renames on success. Files are written with mode 0644 regardless of zip
// metadata; directories with 0755. Path safety, PATH_MAX (per-platform), and
// destination free-space preflight are checked before any write.
//
// Returns ErrPathUnsafe, ErrCapExceeded, or wrapped filesystem errors. On
// failure, the partially-extracted staging directory is removed before returning.
func Extract(ctx context.Context, path, destDir string, allowOverwrite bool) error

// Load opens either a .dip or a .dipx based on filename extension. Returns
// a Source with identical semantics for both formats. Callers needing
// bundle-only methods (Manifest, Identity, ReadFile, Lookup) should use Open
// directly.
func Load(ctx context.Context, path string) (Source, error)

// Canonicalize returns the canonical form of a bundle-relative path or an
// error if the path violates any rule in the Path canonicalization section.
// All callers within the dipx package and its consumers MUST use this single
// function for canonicalization; CI checks for `path.Clean`/`filepath.Clean`
// outside this function within `dipx/`.
func Canonicalize(p string) (string, error)
```

### Source implementations

Two implementations satisfy `Source`:

- **`*Bundle`** (`.dipx`) — holds parsed manifest plus every workflow in memory. `Workflow(refPath, relativeTo)` calls internal `Resolve`, then a map lookup. The lookup returns `ErrFileMissing` if the resolved path is not in `files[]` — this enforces the runtime hermetic invariant.
- **`*dirSource`** (unexported, for `.dip` on disk) — holds the entry workflow and a base directory. `Workflow(refPath, relativeTo)` joins via `filepath` (OS-aware separators, preserving today's `flatten.Resolver` behavior on Windows for the `.dip` path), parses lazily, applies path-safety checks identical in spirit to the bundle case (rejecting `..` escapes from the base directory), and normalizes via `simulate.EnsureConditionsParsed` before returning.

`dirSource`'s lazy cache is bounded by an LRU of 256 entries with `singleflight.Group` for cold-call coalescing. Eviction is transparent to callers.

### Path safety on every read

`Source.Workflow`, `Bundle.Lookup`, `Bundle.ReadFile`, and `Bundle.Resolve` all re-apply path canonicalization on every call via `dipx.Canonicalize`, not just at `Open`. This is defense-in-depth: even if a future bug let a non-canonical path into a workflow's ref string post-Open, runtime resolution will still reject it.

### `OpenLax` discipline

`OpenLax` weakens the strict-mode trust boundary. It MUST NOT be invoked on bytes obtained from any non-local source. The function is intended to emit a structured warning to a caller-provided logger (when `context.Context` carries one via standard convention) on every invocation. **[Structured-warning emission deferred to v1.1 — Phase 9 P9.1.](../plans/2026-05-07-dipx-followups.md) v1 OpenLax is silent at the package level; production deployments SHOULD audit invocation sites in source.**

`dipx.OpenLax` does not log to package-level loggers (see *Logging discipline*). Production deployments embedding `dipx` SHOULD audit invocation sites. CLI flags that opt into lax mode (`dippin validate --lax`, etc.) MUST be opt-in only, never default.

`dippin pack`, `dippin unpack`, and `dippin inspect` MUST NOT use `OpenLax`. Lax mode is exclusively a debug/forensic mode.

### Open post-conditions

After a successful `Open`/`OpenReader`/`OpenLax`:

1. The manifest is well-formed; `format_version` is supported; manifest digest (`Bundle.Identity()`) is recorded.
2. The bundle uses only permitted ZIP features.
3. Every file listed in `files[]` exists in the zip with matching SHA-256, computed over decompressed bytes that the bundle now holds in memory.
4. Every workflow parses successfully via `parser.NewParser`.
5. Every workflow is normalized: `Condition.Parsed` is populated for every edge condition and every `manager_loop` `StopCondition`/`SteerCondition`.
6. Every transitive `subgraph ref:` and `manager_loop subgraph_ref:` resolves to a manifest-listed entry inside `workflows/`.
7. Subgraph reference graph is acyclic (verified via tri-color DFS).
8. All caps are within both bundle conformance limits and the consumer's deployment limits.
9. **No file handles are held; no goroutines are running; the returned `*Bundle` is immutable.** This invariant holds on every exit path including error returns.

### Error precedence

When a bundle fails multiple invariants, `Open` returns the first error encountered in this precedence order:

1. ZIP structural / forbidden feature (encryption, BZIP2, multi-disk, central-dir mismatch, duplicate entries, symlink mode, truncation).
2. Manifest decoding (size cap, BOM, JSON syntax, depth cap, duplicate keys).
3. Manifest shape (format_version, missing required fields, malformed sha256, reserved-key presence, path canonicalization, duplicate paths).
4. Extra zip entries (non-directory entries not listed in `files[]`).
5. Cap exceeded (during streaming hash verification).
6. Hash mismatch (file body bytes don't match manifest).
7. Parse error (workflow source is invalid Dippin).
8. Ref resolution / cycle / escape.

This precedence ensures that "I have a malformed manifest" doesn't surface as "hash mismatch" and that operators can triage errors by category. Implementations SHOULD NOT short-circuit early errors when later errors might be more useful — first error wins.

### Per-sentinel error context (normative)

`BundleError.Path` carries one of three values depending on where the error originates:

- **(a) Bundle-relative path** — the canonical, normal case for read-side errors that fire after the bundle is open and a path context exists (e.g., `ErrHashMismatch`, `ErrFileMissing`, `ErrPathUnsafe`).
- **(b) JSON field name** — for manifest decode errors that fire before the bundle path is in scope inside the manifest decoder (`ErrManifestInvalid`, `ErrUnsupportedFormatVersion`). `Open` MUST enrich these errors to case (a) before returning, so external callers of `Open`/`OpenReader`/`OpenLax` always observe a bundle-relative `Path` for read-side errors.
- **(c) Source filesystem path** — for Pack-side errors that fire before any bundle file exists (e.g., `ErrPathUnsafe` raised by `Pack` against a source tree). Pack errors are not subject to the (b)→(a) enrichment requirement: there is no bundle file yet.

The table below describes the post-enrichment contract for read-side errors and the natural origin for Pack-side errors.

| Sentinel | `BundleError.Path` | `BundleError.Detail` | `BundleError.Cause` |
|---|---|---|---|
| `ErrUnsupportedFormatVersion` | bundle path | `"got N; supports [1]"` | nil |
| `ErrManifestMissing` | bundle path | reason (e.g., "not present at zip root") | nil |
| `ErrManifestInvalid` | bundle path | offending field name + value | nil |
| `ErrFileMissing` | missing path | `"listed in files[] but absent from zip"` | nil |
| `ErrFileUnexpected` | unexpected path | `"in zip but not in files[]"` | nil |
| `ErrHashMismatch` | file path | `"expected: <hex>; actual: <hex>"` | nil |
| `ErrPathUnsafe` | offending path | which canonicalization rule fired | nil |
| `ErrEntryNotInManifest` | entry path | nil | nil |
| `ErrRefEscape` | parent path | the ref string and resolved path | nil |
| `ErrRefCycle` | first cycle node | the cycle in `a → b → a` form | nil |
| `ErrCapExceeded` | offending entry path or empty | cap name + observed value | nil |
| `ErrEntryParse` | entry path | summary of parse failure | underlying parser error |
| `ErrSubgraphParse` | subgraph path | summary of parse failure | underlying parser error |
| `ErrZipFeatureForbidden` | offending entry path | feature name (e.g., "encryption", "BZIP2") | nil |
| `ErrZipTruncated` | bundle path | byte offset of truncation if known | nil |

Consumers use `errors.Is(err, ErrHashMismatch)` for sentinel discrimination and `errors.As(err, &be)` to extract the structured fields.

## CLI

### New commands

```sh
dippin pack <entry.dip> [-o output.dipx] [--dry-run]
```

- Walks every transitively-reachable subgraph ref from disk.
- Runs structural validation first (same checks as `dippin validate`); refuses to pack invalid input.
- Runs all path-safety checks; refuses to pack with informative errors.
- Lint is not run automatically; user can run it independently.
- Defaults output to `./<basename>.dipx` next to the entry file.
- `-o -` writes to stdout.
- **`--dry-run`** validates the source tree and computes the manifest without writing output. Useful for CI lint.
- **Atomic writes**: when writing to a path, writes to `<output>.tmp` and renames on success. Mid-Pack failure leaves no partial output. (When writing to stdout, atomicity is the caller's responsibility.)

```sh
dippin unpack <bundle.dipx> [-o destdir] [--force]
```

- Defaults to `./<basename>/`.
- Atomic: writes to `<destdir>.tmp` and renames on success. Mid-extract failure removes the staging directory.
- Without `--force`, refuses to overwrite existing destdir.
- All path-safety checks apply; PATH_MAX checked per platform; destination free-space preflight.

```sh
dippin inspect <bundle.dipx> [--no-verify] [--format=text|json]
```

- Default `--format=text` prints human-readable manifest with verification status footer:

```text
format: 1
entry:  workflows/api_design.dip
identity: sha256:0a7d9f...
files:
  workflows/api_design.dip                       sha256:abc123…
  workflows/interview_loop.dip                   sha256:def456…
  workflows/phases/code_review.dip               sha256:789abc…
status: VALID (3 files, 24831 bytes, format_version 1)
```

- `--format=json` emits the manifest plus a status object for machine consumption (`jq`-friendly). The status object schema is:

```json
{
  "format_version": 1,
  "entry": "workflows/api_design.dip",
  "files": [
    {"path": "workflows/api_design.dip", "sha256": "abc123..."}
  ],
  "status": {
    "valid": true,
    "verify_skipped": false,
    "file_count": 3,
    "byte_total": 24831,
    "format_version": 1
  }
}
```

When `--no-verify` is passed, `verify_skipped` is `true` and `valid` reflects only structural admission (manifest shape, extras check, refs); hash verification is skipped. When verification fails, `valid` is `false` and a sibling `"error"` field carries the sentinel name (e.g., `"error": "hash mismatch"`).
- Default runs full `Open`-equivalent validation. `--no-verify` skips hash checks for forensic inspection (status footer reads `INVALID: <reason>`).

### CLI exit codes

| Exit code | Meaning |
|---|---|
| 0 | Success |
| 1 | User error (missing entry, invalid arguments, validation failure pre-pack) |
| 2 | Integrity failure (`ErrHashMismatch`, `ErrManifestInvalid`, `ErrUnsupportedFormatVersion`, `ErrZipFeatureForbidden`, `ErrZipTruncated`, `ErrFileMissing`, `ErrFileUnexpected`, `ErrEntryNotInManifest`, `ErrRefEscape`, `ErrRefCycle`, `ErrCapExceeded`, `ErrPathUnsafe`) |
| 3 | I/O error (filesystem failure, write failure, disk full) |
| 4 | Cancelled (context cancellation, signal) |

These are stable; shell scripts wrapping `dippin` can rely on them.

### Existing commands — uniform extension

These accept `.dip` and `.dipx` interchangeably via `dipx.Load`:

`dippin validate`, `lint`, `doctor`, `simulate`, `parse`, `cost`, `coverage`, `unused`, `diff`.

Per-command behavior on a bundle: default operates on the entry workflow only; `--all` runs on every workflow.

### Validation layers

| Check | `.dip` | `.dipx` |
|---|---|---|
| Workflow structural validation (DIP001–009) | ✓ | ✓ on entry; ✓ on all with `--all` |
| Manifest well-formed + JSON encoding rules | — | ✓ |
| Format version recognized | — | ✓ |
| Path canonicalization | ✓ on resolution | ✓ |
| ZIP feature constraints | — | ✓ |
| Files in manifest exist with matching SHA-256 | — | ✓ |
| Entry listed in `files[]` | — | ✓ |
| Hermetic invariant | — | ✓ |
| No extra zip *file* entries | — | ✓ (default; `OpenLax` tolerates) |
| Cycle and depth-cap checks | ✓ via dirSource | ✓ |
| Conditions normalized (`Condition.Parsed`) | ✓ via dirSource | ✓ |
| Bundle identity recorded | — | ✓ |
| Signature verification *(future)* | — | ✓ when present (v2) |

## Versioning

- `format_version: 1` is the only valid value today.
- Bumping `format_version` requires (1) a documented breaking change, (2) a migration tool, (3) the prior version stays supported for at least one release cycle.
- Reading rules across all versions: known version → read; unknown version → reject; never "warn and try."
- The `dipx` Go module ships with `dippin-lang` semver tags. Format support is documented in `SupportedFormatVersions()`.

### Forward compatibility

Tolerant decoding of unknown JSON keys is the v1 mechanism for additive evolution. Three rules:

1. **Tolerant decoding applies only after `format_version` is recognized.** A v1 reader receiving `format_version: 2` rejects before parsing further keys.
2. **Tolerant decoding accommodates only optional additions.** A future feature whose presence requires consumer behavior change MUST bump `format_version`.
3. **Reserved keys are not tolerated; they are rejected.** A reserved-but-tolerated key would create a downgrade-attack vector. v1 reserves and rejects `signatures`; v1 reserves the zip entry name `manifest.sig`.

### Sketch of v2 detached signatures

This sketch is non-normative for v1 but constrains v1 design choices to remain forward-compatible:

```text
Bundle layout (v2):
  manifest.json
  manifest.sig          # detached signature
  workflows/...

Manifest schema (v2):
  {
    "format_version": 2,
    "entry": "...",
    "files": [...],
    "signatures": {       // mandatory in v2 if signing enforced; optional otherwise
      "alg": "ed25519",
      "key_id": "...",
      "scheme": "..."
    }
  }
```

Open ordering for v2:

0a. Open zip; read `manifest.json` bytes-as-stored (do not decode yet).
0b. Read `manifest.sig` bytes; verify signature over `manifest.json` bytes-as-stored against configured trust roots. Failure → `ErrSignatureInvalid`.
0c. Only after signature verification succeeds, decode the manifest and proceed with v1 ordering steps 3–10.

Signature is over the bytes-as-stored of `manifest.json` (which is reproducible because Pack canonicalizes manifest output). Signer identity is in the signature envelope, not the manifest, breaking the chicken-and-egg (you don't trust the manifest until you've verified it; the signature tells you who to trust).

## Concurrency and immutability

`Source` is safe for concurrent reads.

- `Open` reads each zip entry into its own `[]byte`, closes zip internals, parses each workflow once, normalizes conditions once, stores results in immutable maps (pre-sized), and returns. No file handles held; no mutable state remains.
- `Source.Workflow(refPath, relativeTo)` is path validation + map lookup over already-parsed-and-normalized IR. Lock-free.
- `*Bundle.Extract` writes to disk and is not concurrent with itself, but is safe alongside concurrent reads of the same `Bundle`.
- `Pack` is reentrant only when called with non-overlapping outputs. Two `Pack` calls writing to the same `io.Writer` is undefined.
- `dirSource` cache uses thread-safe LRU with `singleflight.Group`; concurrent first-call-misses on the same path are coalesced.

Returned `*ir.Workflow` values MUST be treated as read-only by callers. Any consumer mutating IR MUST clone the workflow first. `Condition.Parsed` is a derived view of `Condition.Raw`; it MUST NOT be serialized into any persistent cache (parse-version skew between `simulate` versions could otherwise serve stale ASTs).

## Runtime integration contract

### Behavioral contract

Workflow execution semantics, params propagation, retry policies, fidelity, fan-in/parallel, and `manager_loop` polling/steering are unchanged. Per-workflow budgets remain scoped exactly as in non-bundled execution. **`.dipx` does not introduce new budget scopes.**

Hermeticity applies to **ref resolution only**, not to runtime data flow. `manager_loop` `SteerContext` injection, agent `Params` propagation, and any runtime-injected context are out of scope of the hermetic invariant.

**Bundle vs disk**: A `*Bundle` is fully in memory; replacing the underlying `.dipx` file on disk does NOT propagate to in-flight executions. Operators deploying via "drop new bundle into place" MUST restart runtime workers (or use the runtime's reload mechanism, if any) to pick up the new bundle. The runtime MAY implement reload by calling `Open` on the new path and atomically swapping the in-flight `Source`.

**Failure isolation**: Each `Bundle` is independent; a failed `Open` of bundle B does not affect in-flight executions of bundle A. Cap enforcement is **per-bundle**, not per-process. Operators running long-lived runtime processes with many concurrent bundles SHOULD configure consumer-deployment limits (see *Soft caps*) appropriate for their resource budget.

### Migration: two call sites

The runtime's existing code:

```go
p := parser.NewParser(input, "foo.dip")
wf, err := p.Parse()
// ... when hitting a subgraph node:
childPath := filepath.Join(filepath.Dir(parentPath), sub.Ref)
data, _ := os.ReadFile(childPath)
childParser := parser.NewParser(string(data), childPath)
child, _ := childParser.Parse()
```

After:

```go
src, err := dipx.Load(ctx, path)        // path may be "foo.dip" OR "foo.dipx"
wf := src.Entry()
// ... when hitting a subgraph node:
child, err := src.Workflow(ctx, sub.Ref, parentPath)
```

`parentPath` semantics:

- For `.dipx`: `parentPath` is bundle-relative (e.g., `workflows/api_design.dip`).
- For `.dip`: `parentPath` is the filesystem path the runtime tracks today.

The `ctx` passed to `Load` SHOULD carry a deadline appropriate to the bundle's source; for HTTP fetch, set a hard timeout (e.g., 30 s) to bound CPU on adversarial input.

### Format version skew

A `.dipx` with unsupported `format_version` returns `*BundleError{Sentinel: ErrUnsupportedFormatVersion, Detail: "got N; supports [1]"}`. The runtime SHOULD surface to operators with a remediation hint such as "upgrade the runtime to a build supporting format_version N."

### Distribution surface

`.dipx` adds:

- File path on disk → `dipx.Load(ctx, path)`.
- HTTP/URL fetch → buffer fully; `dipx.OpenReader(ctx, bytes.NewReader(buf), int64(len(buf)))` with a deadline-bearing `ctx`.
- Stdin / pipe → buffer fully; `OpenReader`.
- In-memory generation → `OpenReader`.

`dipx` defines nothing about networking, registries, caching, or signature verification.

## Operational ergonomics

### Logging discipline

The `dipx` package emits no log output of its own. All observability MUST be via:

1. Returned `*BundleError` values (with structured fields for `errors.As`).
2. (Future v1.1) Optional tracer hooks injected via `context.Context` or functional options.

This makes `dipx` quiet by default and lets host applications (the runtime, dippin CLI) own log formatting and routing.

### Recommended runtime logging

The runtime SHOULD log bundle errors at `error` level with a stable structured field `error_class: bundle_integrity` plus the `BundleError` fields (`sentinel`, `path`, `detail`). Recommended user-facing message templates per sentinel are documented in `dipx/errors.go` doc comments and SHOULD be adopted for cross-deployment consistency.

### Diagnostic mode

When `Open` fails, operators MAY enable diagnostic tracing via:

- Setting `DIPX_DEBUG=1` in the process environment causes the `dipx` package to emit a structured trace of the 9-step Open ordering to stderr (one JSON object per step). This is the only exception to the no-default-logging rule and is gated on the env var. **[Deferred to v1.1 — Phase 9 P9.4.](../plans/2026-05-07-dipx-followups.md) v1 emits no diagnostic-mode output regardless of `DIPX_DEBUG`.**
- Library callers MAY pass a tracer via `context.Context` (post-v1 hook).

### Bundle equality

Two bundles are *byte-equal* if their `.dipx` file bytes are identical (Pack reproducibility makes this the common case for re-pack of identical sources).

Two bundles are *identity-equal* if `b1.Identity() == b2.Identity()` — i.e., their manifest digests match. This is the operator-meaningful definition. Bundle-byte-equality is *not* a reliable identity check because zip ordering is malleable under valid re-encoding.

## Known v1 limitations

The following are known design trade-offs and intended follow-ups:

1. **No cryptographic signatures.** Authenticity over untrusted channels is unaddressed in v1. Sketch in *Versioning*; v2 work.
2. **No bundle-level budget aggregation.** Per-workflow budgets remain. Operators wanting a run-level cap need a separate runtime feature.
3. **No cross-language conformance suite.** This spec is normative for the Go reference implementation. A Rust/Python re-implementation would need a v2 conformance suite.
4. **No hash algorithm agility.** SHA-256 is locked. Migrating to SHA-3 or BLAKE3 requires `format_version` bump and is a deliberate one-way door.
5. **No external asset bundling.** Only `.dip` files are bundled. Asset-reference syntax in the language → corresponding `.dipx` extension.
6. **No streaming Open.** All workflow bytes are read into memory at Open time. Practical limit: 100 MB total uncompressed (the cap).
7. **No `OpenLite` mode that drops file bytes.** `Bundle.ReadFile` requires bytes to be retained. The runtime uses ~2× workflow bytes (raw + parsed IR) per open bundle. Memory profile of long-running runtime processes SHOULD be measured before scale deployment.
8. **No process-wide memory accountant.** Each `Bundle` enforces per-bundle caps; the host is responsible for limiting concurrent open bundles. v1.1 may add a `WithMaxTotalBytes` hook.
9. **No parallel hash verification.** Step 6 of Open ordering is serial in v1. Future versions MAY parallelize across files; this is an implementation choice, not a wire-format change.
10. **Hash comparison is not constant-time.** This is deliberate: the threat model is integrity, not authentication. Constant-time would obscure the model.

## Testing strategy

### Unit tests in `dipx/`

```text
dipx/
  dipx_test.go        # Pack/Open/Load round-trips, format_version handling, context cancellation
  manifest_test.go    # JSON encoding rules, duplicate-key rejection, json.Number, depth cap
  resolve_test.go     # Path canonicalization (single-function invariant), ref resolution, cycles
  zipio_test.go       # Forbidden ZIP features rejected; verifiedBytes type-encoded ordering
  errors_test.go      # BundleError fields, errors.Is/As, per-sentinel context contract
  testdata/
    well-formed.dipx              # frozen byte-vector golden file
    [+ ~30 negative fixtures: see sub-list below]
```

Fixtures cover (one fixture per condition):

- Manifest: corrupt, BOM, duplicate keys, depth-33 nesting, oversized (> 1 MB), trailing data, mismatched-hash, missing-file, extra-file, signatures-key-rejected, bad versions (0, -1, 1.0, "1", 999, 2^53+1).
- Paths: NFD, NUL, control char, leading whitespace, trailing dot, `..` in manifest, absolute, backslash, > 1024 bytes, > 16 components, non-`.dip`, Windows-reserved, case-fold-collision.
- ZIP: encryption, multi-disk, BZIP2, symlink mode, CP437 filename, central-dir/local-header mismatch, duplicate entries, truncation, manifest.sig present.
- Refs: self-loop, 3-cycle, diamond (must succeed), depth-65, escape-attempt, fan-out-257.
- Strict vs lax: `__MACOSX/` extras rejected by `Open`, tolerated by `OpenLax`.

### Required test cases

- **Round-trip** (byte-equality + `reflect.DeepEqual`): Pack → Open → assert every file's bytes via `Bundle.ReadFile` match source AND parsed IR matches `parser.NewParser(source).Parse()`.
- **Reproducibility**: Pack same source twice → assert byte-identical `.dipx`.
- **Hash sentinels**: Corrupt one byte; assert `errors.Is(err, ErrHashMismatch)`, `errors.As(err, &be)` populates `be.Path`, `be.Detail`.
- **Hash format check ordering**: 65-char `sha256` field with body matching truncated-hash → assert `ErrManifestInvalid` (not `ErrHashMismatch`).
- **Truncation distinguishability**: truncated zip → `ErrZipTruncated`, never coerced to `ErrHashMismatch`.
- **Path safety, every entry**: each canonicalization rule failure → `ErrPathUnsafe`.
- **Path safety on every read**: `src.Workflow("../../../etc/passwd", validParent)` → `ErrPathUnsafe`.
- **Streaming cap enforcement**: custom `ReaderAt` that tracks decompressed bytes; assert `Open` aborts before allocating beyond cap. Wrap in a memory-bound assertion (use `runtime.MemStats` deltas across Open) to verify allocation never exceeds cap × small constant.
- **Context cancellation**: `Open` with a cancelled context → returns `context.Canceled`. Slow `ReaderAt` + 1 ms deadline → `context.DeadlineExceeded` within bounded time.
- **FD cleanup on error paths**: corrupt-manifest fixture; assert process FD count is unchanged after 1000 failed `Open` calls.
- **Type-encoded ordering**: CI grep test that `parser.NewParser` is called from exactly one site in `dipx/`. Static analysis test that no function takes both `*zip.Reader` and calls into parsing.
- **Signatures-key rejection**: manifest with `"signatures": [...]` → `ErrManifestInvalid` (not silently tolerated).
- **Manifest digest stability**: load same bundle twice → `b1.Identity() == b2.Identity()`. Re-encode zip with different ordering → same `Identity()`.
- **Pack TOCTOU defense**: source tree containing a symlink → Pack rejects with informative error.
- **Cycle detection variants**: self-loop, 3-cycle, diamond (must succeed), depth-65.
- **Concurrency**: 1000 goroutines, starter-channel adversarial timing, calling `Source.Workflow` on a shared `*Bundle` under `-race`. Repeat for `*dirSource` (cache-fill races).
- **`Condition.Parsed` populated**: assert non-nil for every conditional edge after Open.
- **Polymorphic `Load`**: same tree as `.dip` and `.dipx` → structurally-equal entry workflows. Asymmetry on broken-subgraph detection asserted explicitly.
- **`OpenLax` audit**: invocation triggers structured warning event (test verifies hook is called).
- **CLI exit codes**: each exit code reachable via crafted inputs.
- **`dippin inspect --format=json`**: output is valid JSON parseable into `Manifest` plus a status object.
- **Pack atomicity**: Pack failure mid-write → no partial file at output path.
- **Extract atomicity**: Extract failure mid-write → no partial files at destdir; staging dir removed.
- **In-test cap generators**: too-big and too-many-files generated in test code.

### Integration test (extends `validator/lint_examples_test.go`)

Add `TestPackExamples`: walk `examples/`, pack each `.dip` with subgraphs, round-trip through `dipx.Open`, then run lint. Plus a synthetic deep-tree fixture (3+ levels, diamond pattern) committed under `dipx/testdata/`.

### CLI tests

Each of `cmd_pack_test.go`, `cmd_unpack_test.go`, `cmd_inspect_test.go` covers ≥ 5 paths: happy path + at least 4 failure modes (missing entry, validation fails, ref escapes, output unwritable, etc.).

### Justfile

Add `pack-examples` recipe that descends `examples/**/*.dip`. Extend `just check` after `validate-examples`.

### Coverage

After implementation, run `just cover` and quote concrete `parser/`, `validator/` percentages as targets. The unit tests above cover every error sentinel; gaps surface in coverage and become follow-ups.

## CLAUDE.md amendment (companion change)

The architectural rule "Packages import `ir` but not each other (except analysis packages that compose: doctor → validator + coverage + cost, unused → coverage + cost)" MUST be amended in the same change as the `dipx` package, adding:

> **Loader tier**: `dipx` may compose `ir + parser + simulate`. The exemption is bounded — `dipx` MUST NOT import `validator`, `cost`, `formatter`, or any other analysis package.

This documents the architectural rationale for `dipx` reaching into `simulate` (for `EnsureConditionsParsed`) and prevents the boundary from drifting.
