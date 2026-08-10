# Issue #73 — File-shipping pack mode (`--no-inline` / `--include`)

**Date:** 2026-07-01
**Status:** Design approved (via multi-reviewer squad debate)
**Issue:** dippin-lang#73 (paired with tracker#430; motivating case pipelines#49)

## Problem

`dippin pack` today inlines every `command_file:` / `prompt_file:` /
`system_prompt_file:` body into the packed `.dip` text (see
`cmd/dippin/pack_shadow.go`). The resulting `.dipx` contains only inline
`command:` / `prompt:` / `system_prompt:` blocks — no external files.

That breaks any script that sources a sibling helper by relative path, e.g.

```sh
. "${graph.workflow_dir}/scripts/lib/bootstrap.sh"
```

Under a packed run the file isn't on disk, so under `set -eu` the run aborts.
The concrete motivation (pipelines#49) is collapsing 22 byte-identical inlined
bootstrap preambles into one *sourced* helper — impossible today without
breaking the packed distribution path.

Two distinct categories of files must reach the bundle:

- **(a) Directive-referenced files** — the targets of `command_file:` /
  `prompt_file:` / `system_prompt_file:`. dippin discovers these statically.
- **(b) Sibling / asset files** — like `bootstrap.sh`, referenced only from
  *inside shell bodies* via runtime interpolation. dippin's parser **cannot**
  discover these (it doesn't parse shell). They must be declared explicitly.

## Goal / acceptance (dippin half)

dippin's deliverable is **structural and tracker-free**: for every file in the
entry's reachable closure (`.dip` workflows + shipped assets),

```text
extract(pack(S, --no-inline --include …))/workflows/<rel>  ==  S/<rel>   (byte-for-byte)
```

with the entry `.dip` present, directives **preserved (not inlined)**, and each
declared asset present at its mirrored path. This is verifiable in dippin's own
test suite with no shell and no tracker.

The end-to-end shell-parity check (`tracker run dev_loop.dip` vs
`dippin pack … && tracker out.dipx` behaving identically under `set -eu`) is
**tracker#430's** acceptance test, not dippin's. dippin must never gate its own
"done" on the tracker runtime (project rule: never-gate-dippin-on-tracker).

## Non-goals

- Glob patterns in `--include` (e.g. `scripts/**`). Deferred; directory
  includes cover the motivating case. Go's stdlib `filepath.Glob` doesn't even
  implement recursive `**` and follows symlinks — not worth the surface now.
- Preserving file mode / executable bit (assets are *sourced*, not executed;
  see Security).
- A `Kind` discriminator field in the manifest (the `.dip` suffix already
  serves this role — see Format).
- Auto-detecting shell `source`/`.` references (a lint warning for the literal
  `${graph.workflow_dir}/<path>` case is a possible later enhancement, out of
  scope here).
- tracker#430 (seeding `graph.workflow_dir` in packed runs) — tracked
  separately; dippin ships nothing gated on it.

## CLI surface

```shell
dippin pack --no-inline [--include <path> ...] -o out.dipx entry.dip
```

- **`--no-inline`** — mode switch. Instead of inlining directive bodies, ship
  the directive-referenced files as bundle entries and keep the `*_file:`
  directives in the `.dip` text. Default (flag absent) = today's inline
  behavior, byte-for-byte unchanged.
- **`--include <path>`** (repeatable) — declare a sibling asset that dippin
  can't discover statically. `<path>` is a **file or a directory**, resolved
  relative to the entry `.dip`'s directory:
  - a **file** ships that one file;
  - a **directory** ships its entire subtree (recursive), via a safe
    `WalkDir` that refuses to descend symlinked directories.
- **`--include` requires `--no-inline`** — using `--include` without
  `--no-inline` is a usage error (an incoherent hybrid: inlined bodies plus
  loose assets).
- A `--include` path that **doesn't exist / is empty / escapes the root /
  is (or traverses) a symlink** is an error (fail fast; catches typos and
  path-safety violations).
- A `--include` that would ship a **`.dip` file** (directly, or found inside an
  included directory) is an error: `.dip` files reach the bundle as workflows
  via subgraph refs, not as opaque assets (assets may not end in `.dip`).

## Bundle format v2

Only `--no-inline` produces `format_version: 2`. Inline mode still emits
`format_version: 1`; existing bundles and readers are untouched.

### Layout — keep the `workflows/` prefix

Assets ship under `workflows/` as **siblings of the `.dip` that references
them**, reusing the existing `bundlePathFor` mapping (`"workflows/" + rel`,
`dipx/helpers.go:441-448`) unchanged.

```text
out.dipx (format_version: 2)
  manifest.json
  workflows/dev_loop.dip              # workflow (entry)
  workflows/sub/child.dip             # workflow
  workflows/scripts/dev_loop.sh       # asset (command_file target)
  workflows/scripts/lib/bootstrap.sh  # asset (via --include scripts/lib/)
```

**Why keep `workflows/` (rejected mirror-root):** `command_file:` and shell
`${graph.workflow_dir}/…` paths are all *relative*, so the absolute prefix is
invisible to resolution — a packed run resolves identically whether or not the
prefix is present, as long as the tree structure is mirrored. tracker#430 seeds
`graph.workflow_dir = <extract>/workflows/` (a one-line `filepath.Join`).
Keeping the prefix:

- leaves the escape-critical `Canonicalize` / `resolveLexically` path and
  ~162 existing test expectations untouched;
- **structurally prevents** an asset named `manifest.json` / `manifest.sig`
  from colliding with the reserved root entries (mirror-root would have needed
  an explicit collision guard).

### Discriminator — the `.dip` suffix, no new field

No `Kind` field is added to `ManifestEntry` (avoids the manifest-`Identity`
determinism risk of a new struct field).

- Manifest entries **ending in `.dip`** are workflows: parsed and cycle-checked
  by `Open` exactly as today (preserves the "cycle-check every listed
  workflow, even unreachable ones" hardening in `dipx/helpers.go:107-133`).
- Entries **not ending in `.dip`** are assets: hash-verified and shipped
  opaque, never parsed.
- **An asset path may not end in `.dip`** (enforced at pack time; removes all
  ambiguity).

### Path validation — version-aware policy

`Canonicalize` (`dipx/resolve.go:21`) currently bundles path *safety* (NFC,
reject `..` / absolute / `//` / backslash / NUL / control chars / component
caps / Windows-reserved) with path *policy* (`checkPathSuffix`: require
`workflows/` prefix **and** `.dip` suffix, `resolve.go:93-102`).

Refactor: pull the suffix/policy check out of `Canonicalize` into a
version-aware `checkPathPolicy(path, version)`:

- **v1:** every path must be `workflows/…​.dip` (unchanged).
- **v2:** every path must be under `workflows/`; workflow entries end in
  `.dip`; asset entries are any canonical relative path under `workflows/`
  that does **not** end in `.dip`.

All the version-independent safety checks stay in `Canonicalize` and apply to
both. `SupportedFormatVersions()` → `{1, 2}` (`dipx/manifest.go:400`).

## Pack pipeline changes

### cmd layer (`cmd/dippin/cmd_pack.go`)

- `parsePackArgs` gains `--no-inline` (bool) and repeatable `--include`
  (`flag.Var` collecting a `[]string`). Validate `--include` ⇒ `--no-inline`.
- `runPack`: in `--no-inline` mode, **skip** `prepShadowSourceTree` entirely
  (no inlining) and pass the real entry to `dipx.Pack` with options. Structural
  validation (`validateEntryPrePack`) runs first, unchanged.

### dipx layer (`dipx/dipx.go`, `dipx/helpers.go`)

Add options to the existing `Pack` (one walk, one entry point — preferred over
a parallel `PackFiles` that would fork the shared skeleton):

```go
type PackOptions struct {
    NoInline bool     // ship directive targets as asset entries, keep directives
    Include  []string // extra sibling files/dirs (relative to entry dir) to ship
}

func Pack(ctx context.Context, entryPath string, w io.Writer, opts PackOptions) (Manifest, error)
```

`PackOptions{}` reproduces today's behavior. The three existing call sites
(all in `cmd_pack.go`) pass the zero value except in `--no-inline` mode.

`walkSourceTree` (or a helper it calls) additionally, when `NoInline`:

1. **Directive assets** — for each walked workflow, collect its
   `CommandFile` / `PromptFile` / `SystemPromptFile` targets, resolved relative
   to *that* workflow's own directory. (dipx may import `parser`; it already
   parses every workflow.)
2. **Include assets** — resolve each `--include` path relative to the entry
   dir; a directory is expanded via safe `WalkDir`.
3. Each candidate is read via the existing `ReadNoFollowSymlinks`
   (`dipx/helpers.go:396` — refuses symlink leaf/ancestor and non-regular
   files) after an `ensureUnderRoot` containment check, hashed (SHA-256), and
   recorded as an asset `packedFile` at `bundlePathFor(...)`.
4. **Dedup** — a file referenced by two nodes, or referenced *and* `--include`d,
   is packed once (reuse the `visited`-map pattern).

Assets join the existing `all []packedFile` slice, which is already sorted by
path in `buildManifestForPack` (`helpers.go:463`) → deterministic bytes. Add a
v2 manifest builder emitting `FormatVersion: 2` (do **not** mutate the v1
builder).

### Open / Extract (`dipx/dipx.go`, `dipx/helpers.go`)

`Open` must **skip non-`.dip` entries** in its parse / ref-listed / cycle-check
passes (`parseAllWorkflows`, `verifyRefsListed`, `detectCyclesAll`) — otherwise
`Extract` of a v2 bundle dies parsing `bootstrap.sh` as a workflow. Assets are
still hash-verified. `Extract`/`writeOneFile` already mirror each entry's
`bundlePath` onto disk (`dipx.go:342-351`) — no change needed there beyond the
policy relaxation.

**No new runtime *resolution* logic:** because the extracted tree mirrors the
source tree under `workflows/`, `parser.ResolveFileDirectives` resolves
`command_file:` against the extracted disk exactly as in a source run. The only
new load-path behavior is the parse-skip above.

## Security (MUST-FIX items from review)

1. **Never trust traversal output.** Every `--include` candidate (glob/readdir/
   WalkDir result) is re-validated: absolutize → `ensureUnderRoot` →
   `ReadNoFollowSymlinks`. Prefer `WalkDir` (won't descend symlinked dirs) over
   pattern matching. A naive `os.ReadFile(match)` that follows symlinks is
   forbidden.
2. **Pack-time caps.** Today the 100 MB-total and 10 000-file caps are enforced
   only at Open. A directory include could sweep past them and emit a bundle
   that fails its *own* Open. Mirror both caps at pack time (across `.dip` +
   assets combined) so the producer gets a clean error instead of a
   self-invalid bundle.
3. **Keep `0644`; do not carry file mode / exec bit.** Assets are *sourced*
   (read), not executed — no exec bit needed. Carrying mode would widen the
   surface (setuid/setgid/sticky/world-writable). Matches the existing
   hardcoded `0644` in `writeZipEntry`/`writeOneFile`.
4. **Reserved-name safety is automatic.** Because assets live under
   `workflows/`, they cannot collide with the root-level reserved
   `manifest.json` / `manifest.sig`. (Assert it anyway for defense in depth.)
5. **Threat model is net-equivalent for code execution** (the runtime already
   executes inlined command bodies); the only net-new surface is files landing
   on the consumer's disk at mirrored paths, fully contained by items 1 & 4 and
   the existing zip-slip defenses. Opaque non-`.dip` assets bypass DIP lint but
   stay within the existing producer/consumer trust boundary (opening a `.dipx`
   was already "run their code").

## Determinism guards

- Sort included/asset candidates before hashing/writing (WalkDir order isn't
  guaranteed) — they flow through the existing path-sort, so ensure they're
  added before it runs.
- Separate v1/v2 manifest builders; the v1 path stays byte-identical, so
  existing v1 goldens and `Identity()` values are unaffected.
- No `ManifestEntry` struct change → no per-file wire-order / `Identity` shift.

## Complexity-cap plan (≤5 cyclomatic / ≤7 cognitive)

Pre-plan helper extraction to avoid blowing the caps:

- `collectDirectiveAssets(wf, srcDir, root) ([]packedFile, error)` — split out
  of `readAndRecord`/`walkSourceTree` (mirror the per-config split already in
  `pack_shadow.go`).
- `resolveIncludePath(path, root) ([]packedFile, error)` — file-vs-dir +
  WalkDir + containment + symlink-refuse + read/hash.
- `checkPathPolicy(path, version)` — the version-aware suffix/prefix check
  pulled out of `Canonicalize`.
- Asset dedup via the existing `visited`-map pattern, not inline.

## Testing

**dipx unit:**
- `Pack` with `PackOptions{NoInline:true}` round-trip: bundle carries asset
  entries at `workflows/<rel>` paths; the `.dip` **retains** its `command_file:`
  directive (proves no inlining); `Open` skips parsing assets; hashes verify.
- `--include` of a directory ships the whole subtree; of a file ships one file;
  dedup when a file is both directive-referenced and included.
- Manifest v2 decode/`checkPathPolicy` accepts non-`.dip` under `workflows/`,
  rejects asset paths ending in `.dip`, rejects escapes; v1 policy unchanged.
- Pack-time cap enforcement (file-count + total-size) triggers a clean error.
- Symlink/`..`/absolute `--include` rejected.

**CLI:**
- `--no-inline --include` produces a v2 bundle; default pack still yields
  deterministic v1 bytes (regression guard).
- `--include` without `--no-inline` is a usage error.
- Missing/empty/escaping `--include` path errors.

**Structural acceptance (tracker-free):** build a source tree with an entry
`.dip` whose `command_file:` script sources a sibling via
`${graph.workflow_dir}/scripts/lib/bootstrap.sh`; pack with
`--no-inline --include scripts/lib/`; extract; assert every `.dip` and asset in
the reachable closure is byte-identical to its source counterpart under
`workflows/<rel>`, entry present, directive preserved.

## Backward compatibility

- Default (inline) pack is byte-for-byte unchanged and still `format_version 1`.
- Existing v1 bundles open identically (v1 policy path untouched).
- No `ManifestEntry` wire change → existing v1 `Identity()` / goldens stable.
- An old dippin binary opening a v2 bundle fails closed loudly with
  `ErrUnsupportedFormatVersion` (integrity exit 2) — the honest tripwire.

## Out of scope / follow-ups

- **tracker#430** — seed `graph.workflow_dir = <extract>/workflows/` in packed
  runs. Separate repo/issue. dippin ships independently.
- **`--include` glob patterns** — later PR if requested.
- **Source-reference lint** — optional warning when a packed command body
  literally contains `${graph.workflow_dir}/<relpath>` or `source`/`.` of a
  relative literal not present in the bundle. Cheap, catches the exact footgun;
  deferred to keep PR1 tight.
