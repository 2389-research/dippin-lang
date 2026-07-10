# #175 — Shared prompt fragments (defaults cascade + per-node include)

**Status:** design approved 2026-07-10
**Issue:** [#175](https://github.com/2389-research/dippin-lang/issues/175). Downstream driver: 2389-research/pipelines #111 (a STATUS/FINAL-LINE control-protocol block duplicated 65× across 11 files, already drifting).

## Problem

`prompt_file:` / `system_prompt_file:` load an **entire** prompt from a file — all or nothing. There is no way to compose a prompt from a shared **fragment** plus a per-node body, so control-protocol boilerplate that must be byte-identical across many agents gets hand-pasted and drifts.

## Goal

Let a shared fragment be declared **once** and applied to many agents:
- a `defaults`-block cascade for the dominant case ("every agent ends with the STATUS contract"), and
- a per-node include for agents that need an extra/different fragment.

Fully additive and non-breaking: every existing `.dip` file parses, validates, formats, and packs unchanged. Works identically in v1 and `dip 2`.

## Language surface

### defaults cascade (applies to every agent node)

```dippin
  defaults
    prompt_suffix_file: protocols/status-contract.md   # fragment from a file
    prompt_prefix: "You are part of an automated pipeline."   # inline literal
```

- `prompt_prefix:` / `prompt_suffix:` — inline literal text.
- `prompt_prefix_file:` / `prompt_suffix_file:` — fragment loaded from a file (the key to cross-file single-sourcing).
- Inline and file forms are mutually exclusive per side (a `prompt_suffix:` and a `prompt_suffix_file:` in the same defaults block is a structural error — ambiguous).
- The cascade reaches **agent nodes only** (tool/human/parallel/fan_in/subgraph/conditional/manager_loop have no `prompt`, so they are unaffected).

### per-node include (opt-in extra fragment)

```dippin
  agent Planner
    prompt: "Draft a plan."
    prompt_include: protocols/plan-extra.md
```

- `prompt_include: <file>` — file-only (an inline "extra" is just more prompt body). Appends after the node body, before the cascade suffix.

### per-node opt-out

```dippin
  agent Summarizer
    prompt: "Summarize."
    prompt_suffix: none      # this agent does NOT get the cascade suffix
    prompt_prefix: none      # ...nor the cascade prefix
```

- At the **node** level, `prompt_prefix`/`prompt_suffix` accept only the literal `none` (opt-out) in this release. Custom per-node override (text/file) is a noted future extension.

## Composition semantics

The formatter runs on **unresolved** IR (`dippin fmt` calls `parser.NewParser(...).Parse()` directly, never `ResolveFileDirectives`), so it always emits the authored directives and round-trips perfectly. Composition happens only at **resolve time** — in `parser.ResolveFileDirectives`, which is the shared path for `pack`, `check`, `validate`, `lint`, `simulate`, `cost`, and any runtime load (`parseAndResolveDip`). LSP/WASM intentionally view unresolved IR and are unaffected.

Effective prompt assembly (each part omitted when empty), per agent:

```
body_with_include = join_nonempty("\n\n", [ Prompt, <content of prompt_include> ])
effective         = join_nonempty("\n\n", [ prefix, body_with_include, suffix ])
```

where `prefix` / `suffix` are the resolved cascade values for that agent — the defaults value unless the node set `prompt_prefix: none` / `prompt_suffix: none`. The **suffix is always the final part**, satisfying the downstream "the very last line MUST be exactly STATUS: …" requirement.

`join_nonempty(sep, parts)` joins only the non-empty parts with `sep`, so a missing prefix/include/suffix introduces no stray blank lines.

Interaction with `prompt_file:`: the cascade and `prompt_include` compose around the **resolved** body — so an agent using `prompt_file:` still receives the cascade prefix/suffix and any include, appended to the file-loaded body.

## Data model

`ir.WorkflowDefaults` (authored fields):
```go
PromptPrefix      string // inline literal (defaults cascade)
PromptSuffix      string
PromptPrefixFile  string // fragment path
PromptSuffixFile  string
```

`ir.AgentConfig` (authored fields):
```go
PromptInclude string // fragment path appended after body
PromptPrefix  string // node-level: "none" (opt-out) or "" (inherit). Reserved for future custom override.
PromptSuffix  string
```

Authored fields are **never overwritten**; `ResolveFileDirectives` composes and writes the resolved text into the existing `AgentConfig.Prompt` (exactly as `prompt_file:` loads its file into `Prompt` today). Downstream consumers (pack-inline, runtime, cost, simulate) read the composed `Prompt`.

A single helper — `ir.EffectivePromptParts(w, n)` or composition inside resolve — centralizes the assembly so pack and any future consumer share one definition.

## Security

Fragment files (`prompt_prefix_file`, `prompt_suffix_file`, `prompt_include`) are loaded through the **existing** `loadDirectiveInto` envelope — the same one guarding `prompt_file`/`command_file`: relative-path containment (`checkContainment`), atomic leaf-symlink rejection (O_NOFOLLOW / ELOOP on unix; fd-based fstat→read elsewhere), size cap via `io.LimitReader`, and TOCTOU hardening (validate on the open fd, never re-resolve by pathname). No new security code — new directives register with the same loader. A fragment path that escapes the base dir, is a symlink, or exceeds the cap is a hard resolve-time error, identical to `prompt_file`.

## Packing

- `pack` (default inline): works automatically. `ResolveFileDirectives` composes the fragments into `cfg.Prompt`; the packed `format_version 1` bundle inlines the fully-composed prompt.
- `pack --no-inline` (#73): ships the `prompt_prefix_file` / `prompt_suffix_file` / `prompt_include` targets as separate `workflows/`-tree entries and keeps the directives, so the extracted tree composes identically — following the exact precedent `--no-inline` set for `prompt_file`. The collector that walks `*_file:` directives gains the three new directive fields.

## Validation & lint

- **Structural (resolve-time errors):** missing/unreadable/oob fragment file → hard error (as `prompt_file`). `prompt_suffix:` **and** `prompt_suffix_file:` both set in one defaults block → structural error (ambiguous); same for the prefix pair.
- **DIP154 (Hint):** an agent sets `prompt_prefix: none` or `prompt_suffix: none` while the `defaults` block declares no cascade of that kind — a no-op opt-out, likely a leftover or mistake. Conservative, no false positives (fires only on the exact none-without-cascade shape). Surfaces in lint/check/watch/doctor. Brings the catalog to 64 codes (DIP101–DIP154).

## Non-goals

- `system_prompt` cascade/include (future; pairs with #72 defaults-block `prompt_file`/`system_prompt_file`).
- Custom per-node prefix/suffix override with text/file (only `none` opt-out now).
- Recursive fragments (a fragment file that itself contains include directives) — fragments are plain text, not re-parsed.
- Interpolation-token placement (`${include:…}`) mid-prompt — rejected in favor of the cascade+include model.

## Surfaces to update (standing directive)

- `ir/` (defaults + agent fields), `parser/parse_nodes.go` + defaults parsing (new keywords), `parser/resolve.go` (composition + new directive registration), `formatter/format.go` (emit new directives, round-trip), the `*_file` collector for `pack --no-inline`.
- **Editor grammars** (new keywords, unlike DIP153): tree-sitter grammar + corpus (regen via `npx tree-sitter generate`), VS Code tmLanguage, Zed.
- Docs: `docs/nodes.md`, `docs/cli.md`, `docs/GRAMMAR.ebnf` (W3C EBNF — `/* */` comments), `docs/llm-reference.md`, `docs/validation.md` (DIP154); `site/content/{language,cli}.md`, `site/static/skill.md`; regenerate `cmd/dippin/generated-spec.md`. `site/content/validation.md` DIP154 section at release per `release-process` (or now, per always-current — decide at implementation).
- Diagnostic catalog count 63 → 64.

## Acceptance criteria

- [ ] `defaults` cascade (`prompt_prefix`/`prompt_suffix` + `_file` variants) applies to every agent; a fixture with 3 agents + one `prompt_suffix_file` composes the fragment onto all three.
- [ ] `prompt_include: <file>` appends a fragment after the body, before the suffix.
- [ ] `prompt_suffix: none` / `prompt_prefix: none` opts a node out of that side.
- [ ] Assembly order is `prefix → body → include → suffix`; suffix is always last; empty parts introduce no blank lines.
- [ ] Fragment files use the existing security envelope (symlink/oob/oversize → hard error); regression test mirrors a `prompt_file` security test.
- [ ] Formatter round-trips all new directives (unresolved IR); `fmt` idempotent.
- [ ] `pack` inline composes fragments into the bundle; `pack --no-inline` ships fragment files + keeps directives.
- [ ] DIP154 fires on `none` without a matching cascade; silent otherwise.
- [ ] Both-forms-set (`prompt_suffix` + `prompt_suffix_file`) is a structural error.
- [ ] Editor grammars + all docs/site/spec updated; catalog count 64.
- [ ] An example `.dip` demonstrates the cascade; example suite stays green.
