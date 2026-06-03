# `.dipx` v1 Follow-ups — Tiered Roadmap (design)

Date: 2026-05-08
Source backlog: [`docs/superpowers/plans/2026-05-07-dipx-followups.md`](../plans/2026-05-07-dipx-followups.md)
Spec: `docs/dipx-spec.md`

## Purpose

The `.dipx` v1 release deferred ~30 findings from per-phase gate reviews into a single backlog file. This roadmap triages every entry in that backlog into a release tier (v1.1 / v1.2 / Later / Wontfix), bundles items that share infrastructure into coherent work units, and writes enough detail per v1.1 bundle that a future `writing-plans` session can produce a granular implementation plan without re-reading the gate-review history.

This is a planning artifact, not an implementation plan. Nothing in this doc proposes code changes directly; v1.1 bundles each become their own implementation plan in a follow-up session.

## Rubric

The two ranking axes are **runtime integration value** and **spec correctness**. Items that score high on both go to v1.1. Items that score high on only one go to v1.2. Items that score low on both — including the security-only ZIP I/O defenses, the Unicode-spoofing hardening, and most testing additions — go to Later.

The rubric explicitly omits two axes that could have driven a different cut:

- **Security risk in isolation** is deferred. Per spec § "Conventions", v1 is Go-only; defense-in-depth gaps that matter for cross-implementation interop don't yet have a real attack surface.
- **Cost / coupling** is not a primary axis but is reflected in the bundling: items that share infrastructure (ctx threading hits 4 sites; case-fold hits 3) are merged so cost is paid once.

## Tier definitions

**v1.1.** Next release. Items must score high on **both** runtime value AND spec correctness. Each bundle has a sketched approach, acceptance criteria, affected files, and any open spec questions that must be resolved during the work. Ready to feed into `writing-plans` without rereading the followups doc.

**v1.2.** Following release. Scores high on one axis, not both. Each bundle has title, original followup IDs, one-paragraph problem statement, and a one-line v1.2 rationale. No acceptance criteria — that comes when v1.2 planning starts.

**Later.** Deferred indefinitely. Low on both axes given the chosen rubric, OR blocked on something external (e.g., a non-Go `.dipx` reader). Re-tiering is fine when the rubric or external context changes. One-bullet entries.

**Wontfix.** Already disposed in the original doc as "leave as-is" / "defense in depth" / "cosmetic, current behavior acceptable." Listed here only so the cross-reference table is complete.

---

## v1.1

### Bundle 1 — Cancellable Pack/Open/Source.Workflow

**Covers:** Phase 6 L4, P10.2, P10.7, P10.10.

**Problem.** Spec § "Cancellation" line 272 says I/O entry points take ctx and check it between CPU-bound stages. Implementation only checks ctx in `Open` between stages and at `writeBundle`'s outermost level. `walkSourceTree` (Pack), `writeBundle`'s inner zip-writer loop, and `verifyAllHashes`' inner per-entry loop all run uninterruptible. `Source.Workflow` takes no ctx at all, so dirSource disk I/O cannot be cancelled either. The runtime is a long-running process; an uncancellable Pack against a deep source tree or an uncancellable verify against a many-entry bundle blocks request-handling goroutines for the duration of the operation.

**Open spec questions.**
- § "Runtime integration contract" example shows ctx-less `Source.Workflow`. Either fix the example to add ctx, or split `Source` into a fast-lookup interface and an I/O entry point. Pick one. Bundle 6 must resolve this.

**Approach sketch.**
- Add `case <-ctx.Done(): return ctx.Err()` at the top of `walkSourceTree.visitNext`, `writeBundle`'s per-entry zip-writer loop, and `verifyAllHashes`' per-entry loop.
- Change `Source.Workflow` signature to `Workflow(ctx, refPath, relativeTo)`.
- Update dirSource and all callers in `dipx/loader.go`, `parser/parser.go`, and any cmd/dippin call sites.
- Audit the rest of `dipx/dipx.go` for any other CPU-bound loops that should also check ctx (proactive, in scope).

**Acceptance criteria.**
- New test: Pack against a synthetic deep source tree (≥1000 files) responds to `context.WithTimeout(50ms)` cancellation within 200ms.
- New test: `verifyAllHashes` against a many-entry bundle (≥100 entries) responds to ctx cancellation between entries.
- New test: dirSource.Workflow honors ctx cancellation during disk I/O.
- All existing `Source.Workflow` callers updated; `go build ./...` clean.
- `gocyclo ≤ 5` and `gocognit ≤ 7` caps held; if a loop body grows, extract a helper.

**Affected files.** `dipx/pack.go`, `dipx/zipio.go`, `dipx/dipx.go`, `dipx/source.go`, `dipx/loader.go`, `parser/parser.go`, any `cmd/dippin/*` call sites, `CHANGELOG.md`.

**Risk / coupling.** `Source.Workflow` signature change is breaking for any external caller. The runtime is the primary consumer and pins to dippin-lang; the change ships in v0.25.0+ and the runtime bumps the import. CHANGELOG entry must call out the signature change explicitly.

---

### Bundle 2 — Inspect output overhaul (text + JSON status object, real --no-verify)

**Covers:** Phase 8 M4, L1, L2, P10.4.

**Problem.** Spec § "CLI / Inspect" example shows `status: VALID (3 files, 24831 bytes, format_version 1)` for text output and "manifest plus a status object" for JSON. Implementation: text footer omits the byte total; JSON `status` is a bare string `"VALID"`, not an object. Separately, `inspect --no-verify` is parsed and prints a stderr advisory but `dipx.Open` is invoked unconditionally — the flag is a no-op, and forensically-suspect bundles cannot be inspected without integrity errors firing. All three of these are surfaces the runtime UI consumes when presenting bundle metadata to operators.

**Open spec questions.**
- JSON status object schema is confirmed in Bundle 6 (criterion #6); this bundle implements to match. If Bundle 6 lands first the question is closed by the time work starts here.
- Decide what "no-verify" means precisely: does it skip hash verification only, or also skip strict-mode admission checks? Recommended: hash-verification skip only; admission checks still fire. This is a behavior decision, not a spec contract change — owned by this bundle.

**Approach sketch.**
- Define an `InspectStatus` struct shared by text and JSON renderers; centralize in `cmd/dippin/cmd_inspect.go` or a small `dipx/inspect.go` helper.
- Add a no-verify code path through `dipx`. Either a new `OpenNoVerify(ctx, ...)` function or a functional option `dipx.WithSkipHashVerify()`. Functional option preferred — it composes with future `WithMaxFiles` / `WithMaxBytes` from Bundle 3.
- Wire `--no-verify` to that path; populate `verify_skipped: true` in the status object.
- Update text footer template to include byte total.

**Acceptance criteria.**
- JSON output includes `"status": {"valid": true, "verify_skipped": false, "file_count": 3, "byte_total": 24831, "format_version": 1}`.
- Text footer is `status: VALID (3 files, 24831 bytes, format_version 1)`.
- `inspect --no-verify` against a tampered bundle exits 0 with `verify_skipped: true` and **does not** fire `ErrHashMismatch`.
- `inspect --no-verify` against a structurally invalid bundle (bad manifest, ZIP truncated) still exits non-zero — admission checks are not skipped.
- New tests cover all four scenarios.

**Affected files.** `cmd/dippin/cmd_inspect.go`, `dipx/dipx.go` (no-verify path), `dipx/inspect.go` (new helper if extracted), `docs/dipx-spec.md` (example update — coordinate with Bundle 6), `CHANGELOG.md`.

**Risk / coupling.** JSON shape change is breaking for any consumer parsing `status` as a string. The runtime is the main consumer; the bump must be coordinated. CHANGELOG must call out the breaking schema change.

---

### Bundle 5 — Manifest / Pack error attribution

**Covers:** Phase 3 gate (manifest decoder error-context), P10.6, P10.9.

**Problem.** Spec § "Per-sentinel error context" table calls for `BundleError.Path` to be the bundle path. The manifest decoder runs before the bundle path is in scope, so current implementation puts either an empty string or a JSON field name (e.g., `format_version`) in `Path` for `ErrManifestInvalid` and `ErrUnsupportedFormatVersion`. Pack-side errors put the source filesystem absolute path in `Path` because no bundle file exists yet. Separately, `parsePackSource` is invoked from `readAndRecord` for both the entry workflow and every transitively-reachable subgraph, but always classifies parse failures as `ErrEntryParse` — subgraph parse failures are misattributed to the entry. All three issues surface in operator UI through the runtime; current attribution is misleading.

**Open spec questions.**
- Bundle 6 must resolve "what does `Path` mean for: pre-bundle Pack errors? manifest decode errors before bundle path is known? subgraph parse during Pack?" before this bundle implements.

**Approach sketch.**
- `dipx.Open` enriches manifest decode errors with bundle path before returning. Wrap or rebuild the `BundleError` at the `Open` boundary.
- Pack errors document `Path = source filesystem path` (after spec edit in Bundle 6).
- `parsePackSource` takes an `isEntry bool`; returns `ErrSubgraphParse` when false. Caller in `readAndRecord` passes `path == s.entryAbs`.

**Acceptance criteria.**
- `ErrManifestInvalid` surfaced through `Open` carries the bundle path in `Path` (not empty, not a JSON field name).
- `ErrUnsupportedFormatVersion` same.
- Pack failure on a bad subgraph ref returns `ErrSubgraphParse` with the subgraph's filesystem path; Pack failure on a bad entry returns `ErrEntryParse` with the entry's filesystem path.
- New tests assert `error.Path` contents for each scenario.

**Affected files.** `dipx/dipx.go` (Open), `dipx/pack.go` (parsePackSource, readAndRecord), `dipx/manifest.go`, BundleError doc comments, `CHANGELOG.md`.

**Risk / coupling.** Blocked on Bundle 6's spec resolution. Land Bundle 6 first (or together as a single PR if scope is small enough).

---

### Bundle 6 — Spec wording clarifications (no-code spec PR)

**Covers:** Phase 2 ("any other separator"), Phase 3 (BundleError.Path semantics), Phase 5 M2 (checkExtraEntries placement), Phase 5 L2/L3 (parseAllWorkflows / detectCycles asymmetry), Phase 8 M1 (integrity sentinels precedence), P10.6 (bundle-relative wording vs Pack reality).

**Problem.** Each item is a specific wording fix in `docs/dipx-spec.md`. They are collected here because (a) all are no-code; (b) Bundles 1, 2, and 5 each depend on at least one resolution from this list before implementation can begin without re-arguing the contract.

**Open spec questions.** Each item below states the question and the recommended resolution.

**Approach sketch.**
- Edit `docs/dipx-spec.md` inline, one section at a time.
- `CHANGELOG.md` entry under v1.1: "Spec clarifications (no behavior change unless paired with bundle N)."
- If the spec has a revision log, annotate each edit.

**Acceptance criteria** (each is a separate spec edit):

1. **§ Path canonicalization rule 2** — narrow "Backslash `\` and any other separator MUST be rejected" to "Backslash `\` MUST be rejected." Drop the broader claim; the implementation does not (and need not) enforce it.

2. **§ Per-sentinel error context** — clarify that `Path` may be (a) bundle-relative when in-bundle context exists, (b) a JSON field name for manifest decode errors before bundle context is established and the error is later enriched at the `Open` boundary, or (c) a source filesystem path for Pack errors. State this explicitly in the table preamble. Recommend: also state that `Open` MUST enrich `Path` to bundle-relative before returning, so external callers always see (a) for read-side errors.

3. **§ Reading procedure** — insert step 4.5 "verify no extra entries beyond manifest-listed (`ErrFileUnexpected`)" between current step 4 (manifest-shape validation) and step 5 (hash-verify). Add `ErrFileUnexpected` to the error precedence list at integrity-class.

4. **§ Reading procedure** — pick a single convention: parse and cycle-check **every manifest-listed workflow** (matches current `parseAllWorkflows`; tighten `detectCycles` to match), OR parse and cycle-check the **entry-reachable subgraph only** (narrower; more work to change). Recommend: every manifest-listed workflow. Document the choice; update `detectCycles` in Bundle 6 (small enough to land alongside the spec edit).

5. **§ Error precedence** — enumerate which sentinels are integrity-class. Recommended set: `ErrHashMismatch`, `ErrManifestInvalid`, `ErrZipFeatureForbidden`, `ErrZipTruncated`, `ErrUnsupportedFormatVersion`, `ErrFileMissing`, `ErrFileUnexpected`, `ErrEntryNotInManifest`, `ErrRefEscape`, `ErrRefCycle`, `ErrCapExceeded`, `ErrPathUnsafe`. Update `cmd/dippin/cmd_inspect.go::classifyExit` to match — small mechanical change, lands in this bundle.

6. **§ CLI / Inspect example** — update the JSON output example so `status` is an object: `{"valid": bool, "verify_skipped": bool, "file_count": int, "byte_total": int, "format_version": int}`. Update the text output example footer to include the byte total: `status: VALID (3 files, 24831 bytes, format_version 1)`. Bundle 2 implements code to match.

7. **§ Runtime integration contract** — fix the `Source.Workflow` example to take ctx, resolving the internal inconsistency with § "Cancellation" line 272. Bundle 1 implements the signature change.

**Affected files.** `docs/dipx-spec.md`, `dipx/loader.go` (detectCycles tightening), `cmd/dippin/cmd_inspect.go` (classifyExit), `CHANGELOG.md`.

**Risk / coupling.** Land first or alongside Bundles 1, 2, and 5. Spec-only edits are reviewable independently; the small code changes (detectCycles, classifyExit) can ship in the same PR or a follow-up — author's choice.

---

## v1.2

### Bundle 3 — Configurable consumer caps

**Covers:** P9.5.

**Problem.** Spec § "Soft caps" says "A conformant reader MAY enforce stricter limits configured for its deployment context." Implementation hardcodes the conformance caps (50 MB per file, 50 MB total, file count, etc.). No `WithMaxFiles(n)` / `WithMaxBytes(n)` functional options on `Open` / `OpenReader`. Runtime deployments may want stricter limits; today they have to fork the package.

**v1.2 rationale.** High runtime value, medium spec correctness (the spec invites this as MAY). Goes to v1.2 because no spec contract is being violated. Composes naturally with Bundle 2's `WithSkipHashVerify` if both ship as functional options.

---

### Bundle 4 — CLI ergonomics

**Covers:** P9.6, Phase 8 L4, L5, L3.

**Problem.** Three independent CLI papercuts plus one preflight gap: (a) Go's `flag` package stops at the first non-flag arg, so `dippin pack examples/foo.dip -o out.dipx` errors; (b) `pack --dry-run -o -` silently discards stdout intent; (c) `parseFile` / `loadWorkflow` print `BundleError.Error()` without the `error:` prefix other commands use; (d) `unpack` lacks the `PATH_MAX` + free-space preflight the spec mandates.

**v1.2 rationale.** Medium runtime value (operator UX; the runtime likely uses the Go API directly, not the CLI), low-to-medium spec correctness. Bundle these to amortize the cost of switching to `pflag`/`cobra` if (a) drives that decision.

---

### Bundle 7 — Error-class refinement

**Covers:** Phase 4 I6/L3, I7.

**Problem.** `ErrZipTruncated` is over-broadened: mid-stream EOF, corrupt deflate, malformed central directory, and `archive/zip.ErrInsecurePath` all map to it. Operationally distinct conditions are conflated. Separately, `archive/zip.ErrInsecurePath` (Go-version dependent on `GODEBUG=zipinsecurepath`) should map to `ErrPathUnsafe`, not `ErrZipTruncated`.

**v1.2 rationale.** Medium runtime value (clearer operator diagnostics), high spec correctness (spec wants distinct sentinels). Held back from v1.1 because the breaking change to error sentinels is best paired with Bundle 6's spec edits if they land in the same release window — but Bundle 6 already had enough scope. Land in v1.2 with a CHANGELOG entry.

---

### Bundle 8 — Operator diagnostics

**Covers:** P9.1, P9.4.

**Problem.** Spec § "OpenLax discipline" mandates a structured warning to a caller-provided logger on every invocation; impl emits nothing. Spec § "Operational ergonomics / Diagnostic mode" calls for env-gated `DIPX_DEBUG=1` step trace; not implemented.

**v1.2 rationale.** Medium runtime value (runtime debugging integration issues), high spec correctness (spec mandates both). Both are additive; no breaking change. Held to v1.2 only because v1.1 has enough scope.

---

### Bundle 9 — dirSource LRU + singleflight + path validation

**Covers:** Phase 6 H2, M1, M2.

**Problem.** Spec § "Source implementations" mandates "LRU of 256 entries with `singleflight.Group` for cold-call coalescing." Impl is unbounded `map[string]*ir.Workflow` guarded by `sync.Mutex` held across `parseDipFile` (disk I/O). Two issues: long-running runtime processes leak parsed IR, and concurrent reads on different paths serialize on the global mutex. Separately, `dirSource.resolveDir` doesn't reject NUL bytes, control chars, NFD-encoded paths, or Windows-reserved names.

**v1.2 rationale.** High runtime value (process-leak + lock-contention are real for long-running processes), medium spec correctness (impl violates a normative line, but the violation is silent — correct results, just unbounded memory + serialization). Held back from v1.1 because the LRU+singleflight rewrite is a non-trivial restructure and the contention is bounded in practice (runtime reuse patterns aren't pathological today).

---

## Later

- **Bundle 10 — Unicode hardening.** Phase 2 spec gaps (NFKC, format-class characters, bidi controls, IDN homographs); manifest case-fold (`strings.ToLower` → `cases.Fold()`) at three sites — manifest `verifyOneFile`, ZIP `recordEntry` (P9.2), ZIP entry canonicalization at admission (P9.3). Deferred: low runtime value (no homograph attack vector through the runtime UI today), spec-correctness gap exists but the implementation is **stricter** than the spec, not laxer — current code rejects more than it must, so non-action does not break any spec-conformant bundle.

- **Bundle 11 — ZIP I/O defense in depth.** C2 (central-directory ↔ local-header mismatch), C3 (per-file compression-ratio cap 1000:1), I5 (hardlink / repeated central-dir pointer), Phase 7 H1 (Pack compression-ratio cap), Phase 7 M2 (Pack `O_NOFOLLOW`), P10.5 (Extract `writeOneFile` `O_EXCL` | `O_NOFOLLOW`). Deferred: per spec § "Conventions" v1 is Go-only; cross-implementation interop is the primary motivation for these defenses, and that surface doesn't yet exist. The 50 MB absolute caps + hash verification provide DoS / smuggling protection in the meantime. **Re-tier when a non-Go `.dipx` reader appears.**

- **Bundle 12 — Test additions.** Phase 2 adversarial input coverage (bidi marks U+200E/U+200F, ZWNBSP/BOM U+FEFF, IDN homograph pairs, percent-encoded `..`), Phase 7 frozen golden test for cross-Go-version stability of `encodeManifestCanonical`, Phase 7 pack symlink coverage (transitive ref symlinks, directory symlinks), P10.8 Pack large-file pre-check (use `info.Size()` from `Lstat` before `os.ReadFile`), P10.11 Justfile `pack-examples` recipe (replace `mktemp -u` + add `dippin inspect` round-trip). Deferred: low on both axes; pull individual tests into whichever bundle motivates them as that bundle lands.

---

## Wontfix

These were already disposed in the original followups doc as "leave as-is." Listed only for cross-reference completeness.

- **Phase 3 cosmetic — `assertCanonicalIntLiteral` partly dead code.** Defense in depth; Int64 + JSON grammar already reject most non-canonical literals. Leave.
- **Phase 3 cosmetic — `sha256` `Detail` message granularity.** Current "sha256 not 64-char lowercase hex" distinguishes the failure mode well enough; promoting to "missing" vs "malformed" adds no operator value.
- **Phase 3 cosmetic — `assertNoTrailingTokens` `dec.More()` check.** Belt-and-suspenders with the EOF Token check; harmless redundancy, leave.
- **Phase 3 — non-canonical `entry` returns `ErrPathUnsafe`.** Spec is ambiguous between `ErrPathUnsafe` and `ErrManifestInvalid`. Current behavior is defensible; document the choice in Bundle 6 (which it already does implicitly via the Per-sentinel error context clarification).

---

## Cross-reference table

Every entry from `docs/superpowers/plans/2026-05-07-dipx-followups.md` mapped to a bundle and tier. Items already resolved in v1.0 are tagged `v1.0 RESOLVED` for quick reference.

| Followup ID                                              | Bundle  | Tier            |
|----------------------------------------------------------|---------|-----------------|
| Phase 2 — NFKC / format-class / bidi / IDN homographs    | 10      | Later           |
| Phase 2 — "any other separator" wording                  | 6       | v1.1            |
| Phase 2 — adversarial test inputs                        | 12      | Later           |
| Phase 3 — manifest decoder error-context                 | 5, 6    | v1.1            |
| Phase 3 — non-canonical `entry` returns `ErrPathUnsafe`  | 6       | Wontfix         |
| Phase 3 — manifest case-fold (`strings.ToLower`)         | 10      | Later           |
| Phase 3 cosmetic — `assertCanonicalIntLiteral`           | —       | Wontfix         |
| Phase 3 cosmetic — `sha256` `Detail` granularity         | —       | Wontfix         |
| Phase 3 cosmetic — `assertNoTrailingTokens` `dec.More()` | —       | Wontfix         |
| Phase 4 C2 — central-dir ↔ local-header mismatch         | 11      | Later           |
| Phase 4 C3 — compression-ratio cap (1000:1)              | 11      | Later           |
| Phase 4 I5 — hardlink / repeated central-dir pointer     | 11      | Later           |
| Phase 4 I6/L3 — `ErrZipTruncated` over-broadens          | 7       | v1.2            |
| Phase 4 I7 — `archive/zip.ErrInsecurePath` handling      | 7       | v1.2            |
| Phase 4 M1 — CI grep for `verifiedBytes{` literals       | —       | v1.0 RESOLVED   |
| Phase 5 M2 — `checkExtraEntries` placement               | 6       | v1.1            |
| Phase 5 L2/L3 — parseAllWorkflows / detectCycles asymmetry | 6     | v1.1            |
| Phase 6 H2 — dirSource canonicalize-style validation     | 9       | v1.2            |
| Phase 6 M1/M2 — dirSource LRU + singleflight             | 9       | v1.2            |
| Phase 6 M3/M4 — Extract atomicity edge cases             | —       | v1.0 RESOLVED   |
| Phase 6 L4 — `Source.Workflow` doesn't take ctx          | 1, 6    | v1.1            |
| Phase 7 H1 — Pack compression-ratio cap                  | 11      | Later           |
| Phase 7 M1 — Pack parent-directory symlink check         | —       | v1.0 RESOLVED   |
| Phase 7 M2 — Pack `O_NOFOLLOW`                           | 11      | Later           |
| Phase 7 L1/L2 — Pack frozen golden test                  | 12      | Later           |
| Phase 7 L3 — Pack symlink coverage in tests              | 12      | Later           |
| Phase 8 M1 — sentinel-to-exit-code mapping               | 6       | v1.1            |
| Phase 8 M2 — I/O exit code unreachable                   | —       | v1.0 RESOLVED   |
| Phase 8 M4 — `inspect --no-verify` advisory in JSON      | 2       | v1.1            |
| Phase 8 L1 — inspect text status footer omits byte total | 2       | v1.1            |
| Phase 8 L2 — inspect JSON status is a bare string        | 2       | v1.1            |
| Phase 8 L3 — unpack PATH_MAX + free-space preflight      | 4       | v1.2            |
| Phase 8 L4 — pack `--dry-run` + `-o -` silent discard    | 4       | v1.2            |
| Phase 8 L5 — parseFile/loadWorkflow error prefix         | 4       | v1.2            |
| Phase 9 P9.1 — OpenLax structured warning                | 8       | v1.2            |
| Phase 9 P9.2 — ZIP entry case-fold (ASCII ToLower)       | 10      | Later           |
| Phase 9 P9.3 — ZIP entry canonicalization at admission   | 10      | Later           |
| Phase 9 P9.4 — `DIPX_DEBUG=1` diagnostic mode            | 8       | v1.2            |
| Phase 9 P9.5 — consumer deployment limits configurable   | 3       | v1.2            |
| Phase 9 P9.6 — CLI flag-position parsing                 | 4       | v1.2            |
| Phase 10 P10.1 — Extract `--force` EXDEV data-loss       | —       | v1.0 RESOLVED   |
| Phase 10 P10.2 — Pack lacks mid-write ctx checks         | 1       | v1.1            |
| Phase 10 P10.3 — inspect JSON encoder error swallowed    | —       | v1.0 RESOLVED   |
| Phase 10 P10.4 — inspect `--no-verify` is a no-op        | 2       | v1.1            |
| Phase 10 P10.5 — `writeOneFile` lacks `O_EXCL`/`O_NOFOLLOW` | 11    | Later           |
| Phase 10 P10.6 — Pack-side `BundleError.Path` is fs path | 5, 6    | v1.1            |
| Phase 10 P10.7 — `walkSourceTree` lacks ctx checks       | 1       | v1.1            |
| Phase 10 P10.8 — Pack reads source files in full         | 12      | Later           |
| Phase 10 P10.9 — `parsePackSource` always `ErrEntryParse` | 5      | v1.1            |
| Phase 10 P10.10 — `verifyAllHashes` lacks per-entry ctx  | 1       | v1.1            |
| Phase 10 P10.11 — Justfile `pack-examples` `mktemp -u`   | 12      | Later           |

---

## Next step

After approval of this roadmap, the v1.1 tier (Bundles 1, 2, 5, 6) goes to a `writing-plans` session — one plan per bundle, in the order: 6 (spec) → 1, 2, 5 (impl, parallelizable). Bundle 6's spec edits and small classifyExit/detectCycles changes land first because Bundles 1, 2, and 5 each reference resolutions from Bundle 6.
