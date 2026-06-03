# `.dipx` Implementation — Phase Gate Log

## Phase 0 gate (2026-05-07)

- **Reviewer:** fast-track (build only)
- **Result:** PASS
- **Summary:** Bootstrap complete. `just build` and pre-commit hook (vet, golangci-lint, race tests, complexity, validate-examples) all green. `golang.org/x/text` already transitively pinned at v0.3.3, which includes `unicode/norm`. No `go.mod`/`go.sum` changes needed. Commit `a98f369`.

## Phase 1 gate (2026-05-07)

- **Reviewer:** fast-track + spec/quality reviewers
- **Result:** PASS
- **Summary:** 15 sentinels per spec, `BundleError` with `Is`/`Unwrap` extension protocol correctly implemented, idiomatic `Error()` formatting, 5 tests pass. Spec reviewer confirmed exact match. Code-quality reviewer: no critical/important issues; stylistic minors only. `just check` clean. Implementer adjusted `TestBundleErrorUnwrap` to satisfy `errorlint` (intent preserved per spec reviewer); added `TestNewError` to keep unexported `newError` helper from tripping `unused` lint (CLAUDE.md forbids `//nolint`). Commit `9d8322b`.

## Phase 2 gate (2026-05-07)

- **Reviewer:** mandatory squad — security subagent (parallel) + PAL spot-check
- **Result:** REMEDIATED → PASS
- **Summary:** Tasks 2 (Canonicalize), 3 (resolveLexically), 4 (cycle detection) implemented and committed. Code-quality reviewer (Task 2) caught a spec-rule-6 bypass (multi-extension Windows-reserved `CON.tar.dip` accepted because `stripExt` used `LastIndexByte`); fixed in commit `2de1a20`. Phase 2 gate (security + PAL) confirmed composition is sound, cycle detection correct, recently-fixed reserved check correct. Two important findings remediated in commit `92bd4da`: ErrRefCycle now reports full cycle path (`a → b → c → a`); resolveLexically's path.Clean usage documented for the future Task 26 CI grep allowlist. 4 new tests added (full-cycle-path, at-cap-success, empty-graph, adversarial-composition). Spec-level gaps (NFKC, ZWJ/ZWNJ, bidi, IDN homographs, "any other separator" wording) deferred to v1.1 in `docs/superpowers/plans/2026-05-07-dipx-followups.md` (commit `2de4034`). 48 dipx tests passing.

## Phase 3 gate (2026-05-07)

- **Reviewer:** mandatory — crypto-discipline subagent + PAL spot-check
- **Result:** REMEDIATED → PASS
- **Summary:** Tasks 5 (Manifest types + decoder) and 6 (verifyManifestShape) implemented. Spec compliance + code quality reviews PASS (one important finding: doc comment over-promised; fixed in commit `eb40a51`). Phase 3 gate confirmed architectural soundness: duplicate-key detection at every level, top-level-only signatures rejection, integer-only `format_version` gauntlet, depth cap. One important spec gap (missing `path` returned `ErrPathUnsafe` instead of `ErrManifestInvalid`) and two test gaps remediated in commit `6c18d0d`. Deferred findings (per-sentinel `Path` semantics across layers; ErrPathUnsafe vs ErrManifestInvalid for non-canonical entry; case-fold via strings.ToLower vs Unicode `cases.Fold`) appended to followups in commit `d825fed`. 67 dipx tests passing.

## Phase 10 final gate (2026-05-07)

- **Reviewer:** comprehensive — 3 squad subagents in parallel (security + crypto-discipline + ops-reliability)
- **Result:** REMEDIATED → PASS — implementation ready for human PR review
- **Summary:** Tasks 25 (cap/concurrency/FD-leak tests) and 26 (final self-review) complete. Three squad reviewers + 9 prior phase gates confirmed: type-encoded ordering invariant intact (verifiedBytes single constructor, parser.NewParser at 3 documented sites with SPEC NOTEs), Bundle.Identity correct, hash-format check fires in step 4 before hash-compare, signatures key rejected in v1, format_version gauntlet hard, FD cleanup proven, Pack atomicity correct, Extract atomicity correct, no package-level logging in dipx, exit codes consistent. Two cheap fixes applied in commit `72b34e1`: (1) new `TestInvariant_ParserNewParserSiteCount` enforces the spec's CI grep requirement at test time (3 sites verified); (2) stale doc comment in source.go updated from "second of two" to "one of three". Six new findings deferred to followups in commit `9127877`: P10.1 (Extract --force EXDEV data-loss strengthening), P10.2 (Pack mid-write ctx), P10.3 (inspect JSON encoder error silent), P10.4 (inspect --no-verify no-op), P10.5 (Extract O_EXCL/O_NOFOLLOW), P10.6 (Pack-side BundleError.Path semantics). End-to-end smoke verified across pack/inspect/unpack/validate on real `examples/orchestrator.dip`. **All 26 implementation tasks done; 9 gates passed (5 mandatory squad, 4 standard) with remediation; full followups file documents all v1.1 deferrals.**

## Phase 9 gate (2026-05-07)

- **Reviewer:** PAL spec-coverage audit + end-to-end smoke
- **Result:** PASS (with deferrals)
- **Summary:** Tasks 22 (TestPackExamples integration test — all 26 examples round-trip cleanly), 23 (justfile pack-examples recipe wired into `just check`), 24 (CLAUDE.md amendment for dipx loader-tier exemption). End-to-end smoke verified: `dippin pack -o /tmp/orch.dipx examples/orchestrator.dip` → `dippin inspect --format=json` (valid JSON, identity hash, manifest with 2 files) → `dippin unpack -o /tmp/orch-unpacked` (preserves directory tree including phases/) → `dippin validate /tmp/orch.dipx` (passes). PAL spec-coverage audit identified 6 additional spec gaps not previously deferred: OpenLax warning emission, ZIP-layer case-fold (Unicode), ZIP entry canonicalization at admission, DIPX_DEBUG diagnostic mode, deployment-limits configurability, CLI flag-position constraint. All 6 added to followups; none block v1.

## Phase 8 gate (2026-05-07)

- **Reviewer:** standard — ops-reliability subagent
- **Result:** REMEDIATED → PASS (with deferrals)
- **Summary:** Tasks 18-21 implemented. New CLI commands `dippin pack`/`unpack`/`inspect` plus existing analysis commands routed through `loadWorkflow`/`parseFile` to accept `.dipx`. Three HIGH fixes applied in commit `5f92ee3`: H1 (pack runs validation pre-pack via `validator.Validate` at CLI layer, honoring loader-tier exemption); H2 (inspect uses classifyExit for proper exit-code mapping); H3 (Manifest gets JSON tags so inspect --format=json output is parseable into dipx.Manifest); M3 (inspect rejects unknown --format values). 2 new tests + 2 updated. Deferred (commit `cf2cdb1`): incomplete sentinel-to-exit-code mapping (M1), I/O exit code (3) unreachable in practice (M2), --no-verify JSON status field (M4), inspect text byte total / JSON status object (L1/L2), unpack PATH_MAX + free-space preflight (L3), pack --dry-run -o - combo diagnostic (L4), error prefix inconsistency (L5). 81+ dipx tests + new CLI tests all passing.

## Phase 7 gate (2026-05-07)

- **Reviewer:** mandatory — security subagent (Pack TOCTOU + reproducibility)
- **Result:** REMEDIATED → PASS (with deferrals)
- **Summary:** Task 17 (Pack) implemented with reproducibility (fixed timestamps, sorted entries, no extras, bit 11 set explicitly per Phase 4 finding) and symlink defense via Lstat. Phase 7 security gate confirmed: TOCTOU single-read invariant, hermetic Pack rejection of escaping refs, manifest canonicalization, bit-11 round-trip success. One HIGH remediated in commit `5cc2288`: per-file size cap (50MB) now enforced at Pack time so producers don't generate bundles their own Open rejects. New `TestPack_RejectsOversizedSource` test. Deferred (commit `865099f`): compression-ratio cap (1000:1; needs manual deflate chain), parent-directory symlink walk, O_NOFOLLOW (platform-specific), frozen golden test, transitive-ref symlink coverage. 81 dipx tests passing. Three parser.NewParser sites now exist; verifiedBytes invariant intact (only Open pathway consumes verifiedBytes; Pack and dirSource consume trusted disk bytes).

## Phase 6 gate (2026-05-07)

- **Reviewer:** standard — runtime integration subagent
- **Result:** REMEDIATED → PASS (with deferrals)
- **Summary:** Tasks 14 (Source interface), 15 (dirSource + Load), 16 (Extract atomic). Runtime-integration gate confirmed: Source.Workflow argument order matches flatten.Resolver, dirSource preserves Windows filepath behavior, simulate.EnsureConditionsParsed eagerly applied, single verifiedBytes-pathway parser.NewParser invariant intact. Two cheap fixes applied in commit `3476dce`: H1 (dirSource boundary check uses canonical `rel == ".." || HasPrefix(rel, ".."+sep)` idiom) and H3 (Load strict extension allowlist; .DIPX uppercase routes to Open; non-.dip/.dipx returns ErrPathUnsafe). 2 new tests. Deferred findings (H2 dirSource Canonicalize-style validation, M1/M2 LRU+singleflight cache, M3/M4 Extract atomicity edge cases, L4 Source.Workflow ctx) appended to followups in commit `d38f065`. Tests passing.

## Phase 5 gate (2026-05-07)

- **Reviewer:** mandatory — crypto-discipline subagent + PAL end-to-end
- **Result:** REMEDIATED → PASS
- **Summary:** Tasks 9-13 implemented. Phase 5 gate confirmed: 9-step Open ordering preserved (modulo extras-check insertion), verifiedBytes type-encoded ordering invariant holds at integration point, parser.NewParser called from exactly one site (verified by grep), error precedence honored (first-error-wins), ctx returned bare for errors.Is, FD cleanup proven on every exit path, Bundle.Manifest()/Identity() correct. Three actionable findings remediated in commit `d64bfd8`: H1 (total-cap now enforced via streaming `min(perFileCap, totalCap-total)`); H2 (Bundle.Lookup and ReadFile now Canonicalize on every call per spec § "Path safety on every read"); M1 (ctx checks added before parse/walkRefs/normalize CPU stages). 3 new defense tests (TestBundle_Lookup_RejectsUnsafePath, TestBundle_ReadFile_RejectsUnsafePath, TestOpen_ContextCancelled). Deferred findings (checkExtraEntries placement; parseAllWorkflows/detectCycles strictness asymmetry) appended to followups in commit `59fcda3`. 71 dipx tests passing.

## Phase 4 gate (2026-05-07)

- **Reviewer:** mandatory dual squad — security subagent + crypto-discipline subagent (parallel)
- **Result:** REMEDIATED → PASS (with deferrals)
- **Summary:** Tasks 7 (constrainedZip + verifiedBytes) and 8 (verifyAndReadEntry streaming hash) implemented. Crypto-discipline gate confirmed type-encoded ordering invariant is structurally sound: `verifiedBytes` only constructible inside package via `newVerifiedBytes`, all error paths return zero-value, hash via TeeReader is single-read with no TOCTOU, perFileCap+1 trick correctly distinguishes at-cap from over-cap. Security gate found 4 critical and 7 important issues. Cheap critical fixes applied in commit `0a9ac3c`: C1 (CP437-on-ASCII bypass — bit 11 now required unconditionally), C4 (short-read mis-classification → ErrZipTruncated), I1 (manifest.sig rejected per spec), I3 (all non-regular mode bits rejected: device/FIFO/socket/char-device), I4 (10000-entry cap). Tests added: at-cap-accepted, one-over-cap-rejected, ascii-without-utf8-flag, manifest-sig-rejected. Expensive findings deferred to followups in commit `c644836`: C2 (central-dir/local-header mismatch — needs custom zip reader; mitigated by hash verification), C3 (compression-ratio cap 1000:1 — needs OpenRaw + manual deflate; absolute 50MB cap provides DoS protection), I5 (central-dir pointer collisions; bundled with C2), I6/L3 (ErrZipTruncated specificity), I7 (ErrInsecurePath handling), M1 (CI grep for verifiedBytes literals — already in Task 26). **Important note for Task 17 (Pack)**: Go's archive/zip writer does NOT auto-set bit 11 for ASCII names — Pack MUST set `h.Flags |= 0x800` explicitly or its own output will fail openConstrainedZip.
