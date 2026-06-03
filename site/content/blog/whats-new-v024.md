---
title: "What's New in Dippin v0.24"
date: "2026-05-08"
description: "The .dipx bundle format — pack a workflow tree into one verifiable file you can ship anywhere. Three new commands, every analysis command extended."
tagStyle: "release"
tagLabel: "RELEASE"
category: "Releases"
readTime: "6 min read"
related:
  - url: "whats-new-v025.html"
    title: "What's New in v0.25"
    summary: ".dipx format v1.1 — real ctx cancellation, an inspect that actually inspects, and exit code 2 that matches the spec."
  - url: "whats-new-v023.html"
    title: "What's New in v0.23"
    summary: "tool_commands_allow / tool_denylist_add defaults plus a cleaner DOT header."
---

You've spent a week stitching together a workflow. It's six `.dip` files: an entry pipeline, three reusable subgraphs, two interview loops your teammate built and you finally got around to wiring in. It runs locally. Now you need to ship it to the runtime.

Today, that means tarring up the directory, scp'ing it, and hoping nobody renames a file in transit. There's no way to ask "is this the same workflow tree I packaged on Friday?" without diffing source files by hand. And if your CI runner unpacks the tarball into a directory that contains a leftover `phases/old_review.dip` from yesterday's build, your pipeline runs against the wrong subgraph and you don't notice until the bill arrives.

`v0.24.0` is a fix for that. It introduces a new file format — `.dipx` — that bundles a workflow and every file it references into one artifact you can verify, address by content, and ship as a unit. Three new commands manage them, and every existing analysis command (`validate`, `lint`, `doctor`, `cost`, all of them) now accepts a `.dipx` directly. You author as `.dip`. You distribute as `.dipx`.

## Why a new format

The reason `.dip` projects spread across multiple files is `subgraph` refs. Even a small project tends to grow them — once you've written one decent review loop, you want to reuse it from three places.

```dippin
subgraph S
  ref: phases/review.dip
```

That `ref:` is a relative filesystem path. At runtime, the executor (`dippin simulate` or any other consumer) walks the directory and opens the referenced file. Works perfectly when you control the filesystem.

It stops working as soon as the workflow needs to leave your machine. Four scenarios where this bites:

- **Shipping to a runtime.** The runtime has to know how to find every file before it can run the entry. A directory tree is fine; a single artifact is better.
- **Pinning a version in CI.** "We deployed workflow `v3.7.2` last Tuesday" should mean exactly the bytes that ran. A directory is hard to pin; a hash is easy.
- **Verifying nothing changed in transit.** Did the workflow you built in CI match the one that hit production? With a directory, you compare files by hand. With one file and a hash, you compare the hash.
- **Caching by content.** A runtime that opens the same workflow twenty times shouldn't re-parse it twenty times. A content-addressed id makes the cache trivial.

The DOT-flattening route — paste every subgraph inline, ship one giant file — solves the "single file" problem but loses the per-file identity. You can't tell which part of the flattened workflow came from `phases/review.dip` anymore. `.dipx` keeps the structure intact and adds the integrity guarantees.

## What a .dipx actually is

It's a ZIP archive. That's the whole punch line. Three rules turn an ordinary ZIP into a `.dipx`:

1. **`manifest.json` lives at the archive root.** It lists every other file's bundle path and its SHA-256, plus a `format_version` and an `entry` field that names the workflow root.
2. **Workflow files live under `workflows/`.** The directory layout mirrors your source tree, so refs that worked on disk still resolve inside the bundle.
3. **The bytes are deterministic.** Two `dippin pack` runs over the same source tree produce byte-for-byte identical bundles. Mtimes are pinned to the ZIP epoch (1980-01-01). Entries are sorted. No platform metadata leaks in. The UTF-8 filename bit is always set.

Determinism is the unglamorous foundation. It means you can hash a bundle and have that hash mean something — across machines, across CI runs, across years. Without it, you'd be hashing whatever timestamp your filesystem happened to produce.

Every `Open` verifies every file. Hashes are checked before any `.dip` content reaches the parser, so a bundle that doesn't match its manifest never has a chance to do anything weird. The Go library encodes this in the type system: the parser only accepts `verifiedBytes`, an unexported wrapper that the verification step is the only place allowed to construct. "Parse before verify" isn't a discipline you have to remember — it's a compile error if you try.

## Pack, inspect, unpack

Three new commands, one obvious purpose each.

```sh
$ dippin pack pipeline.dip
$ ls
pipeline.dip   pipeline.dipx

$ dippin inspect pipeline.dipx
format: 1
entry:    workflows/pipeline.dip
identity: sha256:0a7d9f...
files:
  workflows/pipeline.dip          sha256:abc123…
  workflows/phases/review.dip     sha256:def456…
status: VALID (2 files, format_version 1)

$ dippin unpack pipeline.dipx -o ./out
```

`pack` walks every transitive `ref:` from the entry, collects the files, hashes them, and writes the bundle. The output is atomic — it goes through a temp file and renames into place, so a half-written bundle never sits on disk. `--dry-run` validates and walks without writing if you just want to confirm the tree is healthy.

`unpack` reverses the process. It stages everything in `<destdir>.tmp`, then renames into the final location. With `--force`, the existing destination is moved aside to a `.bak` first and restored if anything goes wrong — so a failed extract can never destroy the directory you were trying to overwrite. (We learned that the hard way: an earlier draft did the obvious thing and would silently nuke your old directory if the staging rename failed across a mount boundary. The reviewers caught it.)

`inspect` shows you what's inside without unpacking. The `identity` line is a SHA-256 over the manifest bytes — a stable content-addressable id you can paste into a Slack message and have someone else verify they have the exact same bundle.

| Command | What it does |
|---------|---------|
| `dippin pack <entry.dip>` | Build a deterministic `.dipx` from a `.dip` entry. `-o <out>` (default `<entry>.dipx`; `-` for stdout). `--dry-run` to validate without writing. |
| `dippin unpack <bundle.dipx>` | Atomically extract a bundle. `-o <dir>`, `--force` to overwrite with rollback. |
| `dippin inspect <bundle.dipx>` | Print manifest, identity, and file list. `--format text\|json`. |

Bundle commands use a finer exit-code ladder than the standard analysis commands: `0` ok, `1` user error, `2` integrity error, `3` I/O error, `4` cancelled. Five buckets so a script can tell whether `dippin pack` failed because your `.dip` had a syntax error or because the disk was full.

## Every analysis command accepts a .dipx

This is the part that makes the format actually pull its weight day to day.

You don't have to choose between authoring as `.dip` and analyzing as `.dipx`. `validate`, `lint`, `doctor`, `parse`, `cost`, `coverage`, `simulate`, `optimize`, `unused`, `graph`, `diff`, `check`, `explain`, `export-dot` — all of them accept either. Internally they route the file through `dipx.Load`, which sniffs the extension and either parses a `.dip` from disk or opens a `.dipx`, hash-verifies it, and feeds the entry workflow into whatever the analyzer wanted in the first place.

```go
src, err := dipx.Load(ctx, "pipeline.dipx")  // also accepts pipeline.dip
entry := src.Entry()                          // *ir.Workflow
```

For runtime integration, the contract is the `dipx.Source` interface. A `dipx.Bundle` satisfies it. So does the on-disk `dirSource`. Any runtime can swap its directory-walking loader for `dipx.Load` and get content-addressed bundle support without touching anything downstream.

The intended rhythm: author your workflow as `.dip` files in your editor, with all the linting and the LSP. When you're ready to ship — to the runtime, to a teammate, to your future self — `dippin pack` it. The bundle is what travels.

## What "content-addressed" buys you

`Bundle.Identity()` returns the SHA-256 of the manifest bytes exactly as they appear in the archive. Two byte-identical bundles have the same identity. Two bundles that contain the same workflows but were packed five minutes apart with different entry ordering would *not* have the same identity — and that's the point. The identity verifies bytes, not intent.

You can use it as a cache key. You can use it as a deployment id. You can put it in a deploy log and answer "is the workflow the runtime is running right now the one that came out of CI on Tuesday?" with a single string comparison.

This is the thing the directory-of-files setup couldn't give you. A directory has no canonical hash. You'd have to invent one (which directories? which file order? what about ignored files?), and then everyone would have to agree. A `.dipx` has one identity, one definition, no debate.

## What v1 doesn't do

`.dipx` v1 guarantees integrity, not authenticity. The hash on every file means the bundle bytes weren't corrupted between `pack` and `open`. But anyone who can run `dippin pack` can produce a *valid* bundle. There's no signature. There's no key. If you receive a `.dipx` from a stranger, all you know is that nothing tampered with it after they made it — not that they're who they claim to be.

That's fine for distribution between parties that already trust each other. It's not fine for "download a `.dipx` from the internet and run it." The spec carries a v2 sketch with detached signatures (`manifest.sig`, `ed25519`, key id). It's not in v1.

A few other things deferred to v1.1, all tracked in [`2026-05-07-dipx-followups.md`](https://github.com/2389-research/dippin-lang/blob/main/docs/superpowers/plans/2026-05-07-dipx-followups.md):

- **Per-file compression-ratio cap (1000:1).** v1 enforces only the absolute 50 MB per-file ceiling. A maximally-compressed entry could expand within the absolute cap. Real risk is low for hand-authored `.dip` files; we'll close it in v1.1.
- **Unicode case-folding for duplicate detection.** The current `strings.ToLower` covers ASCII but misses things like German `ß` ↔ `ss`.
- **Central-directory ↔ local-header agreement.** Defense against ZIP parser-confusion attacks. v1 is mitigated by per-file SHA-256: any bytes that get decompressed get hashed, and if they don't match the manifest, the open fails.
- **`DIPX_DEBUG=1` step trace.** Spec calls for it; v1 doesn't ship it. The 9-step Open ordering is documented in the spec if you need to follow along.

What v1 *does* take seriously is the producer side. `dippin pack` refuses symlinks anywhere in the source tree — not just at the leaf, but at every intermediate directory. We added that defense after a reviewer pointed out that `workflows/phases -> /etc` would happily exfiltrate `/etc/secret.dip` into the bundle. (Two of three independent reviewers flagged it as v1-blocking. They were right.)

## What's next

Full notes in [CHANGELOG.md](https://github.com/2389-research/dippin-lang/blob/main/CHANGELOG.md). The runtime side of this — swapping the loader for `dipx.Load` so workflows can be deployed as bundles — is its own conversation. Ping us if you want to follow along, or if you have a use case for `.dipx` we haven't thought about.

For now: `dippin pack pipeline.dip`, ship the `.dipx`, sleep better.
