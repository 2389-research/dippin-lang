# Issue #41 — Terror Squad Findings (shipped in v0.32.0)

**Date:** 2026-05-19
**Status:** Shipped in v0.32.0 — see [`docs/superpowers/specs/2026-05-26-issue-41-design.md`](../specs/2026-05-26-issue-41-design.md). The simplification process is documented in the spec's "Design journey" section.
**Raw reports:** five reviewer JSONL transcripts under `/tmp/claude-1000/-home-clint-code-2389-dippin-lang/.../tasks/` (ephemeral; consolidate any nuance from there before they age out).

## Why this document exists

Mid-brainstorming for v0.30.0, a five-reviewer panel (language-design / security / IR-API / parser-compiler / future-maintainer) gutted the initial design. Rather than re-discover the same findings the next time #41 is attempted, this doc captures them for the spec writer who picks it up next.

**Do not treat this as a spec.** It is a research artifact. The eventual v0.31.0 spec must make its own decisions — but it must address (or knowingly skip with rationale) every item below.

## The parking decision

The original framing: dippin v0.30.0 ships parser + IR + validator + DOT export + migrate for new `tools:` and `disallowed_tools:` agent-node fields. The runtime untouched until a follow-up. The fields would be a runtime no-op until the runtime plumbed them through.

**This is the load-bearing critical finding.** The Security and Future-maintainer reviewers independently concluded that shipping lint-validated, runtime-no-op safety fields is *materially worse* than the status quo: authors get green diagnostics on `tools: none` and ship believing they're sandboxed when they aren't. The project's own track record (v0.22, v0.23) shipped runtime-side in the same batch when the language surface had security semantics. Shipping advisory-only contradicts that pattern.

Three coupled scope choices were on the table:

1. **Park #41**, do a smaller v0.30.0 (chosen).
2. Ship narrowed dippin-only with a hard-warn lint on every use of the new fields until the runtime lands.
3. Pull the runtime into the same batch as a joint release.

Choice 1 was selected. Choice 3 is the natural return path for v0.31.0 — design dippin and the runtime together, ship together, no advisory window.

## What blocks unparking

- Approval to design a runtime-side change concurrent with the dippin spec (the runtime has been intentionally untouched through v0.28/v0.29/v0.30).
- Time budget for a cross-repo design pass (two specs, two plans, coordinated PRs, joint release).
- Decision on the criticals below (a fresh brainstorming pass for #41 should *start* with this list, not rediscover it).

## Critical findings — every one must be addressed in the v0.31 spec

### 1. Per-node-only misses the bug class

The v0.28.2 incident was an *author-judgment* failure — the author didn't identify the dangerous node. Per-node-only `tools:` requires the same judgment to deploy. A `defaults: tools: read_only` cascade is the only thing that catches "safe by default, opt out per node." DIP28 precedent (`ToolCommandsAllow` in defaults) was specifically chosen for this reason.

**Sources:** Language Critic C1, Security Threat-Modeler I4, IR/API Designer I3.

### 2. DIP141 reserved-only ships a known typo footgun

`disallowed_tools: bash, applypatch` (lowercase) silently no-ops against the runtime's CamelCase catalog (`Bash`, `ApplyPatch`). DIP141 was reserved to catch this but with no firing logic. The v0.29.0 spec explicitly rejected this exact pattern (`route_required: yes` silently parsing as `false`) as "**not acceptable**" when the failure mode silently disables a safety net. Reserving DIP141 contradicts the project's own established precedent.

**Resolutions:**
- Ship `ir.KnownAgentTools = []string{"Read","Write","Edit","ApplyPatch","Glob","GrepSearch","Bash"}` — already implicitly coupled to the runtime; making it explicit beats silent failure.
- Case-normalize on parse to canonical CamelCase.
- Or: drop `disallowed_tools` from the surface entirely (see #5 below).

**Sources:** Language Critic C2, Security Threat-Modeler C3, Future-maintainer #6.

### 3. `Params` map is a parallel attack surface

`agent X { tools: none; params: { allowed_tools: Bash, Edit } }` re-enables tools the typed field disabled. The runtime already reads runtime-affecting keys from `Params` (e.g., `permission_mode`). The original design has zero coverage.

**Resolution:** new lint code for known-dangerous param keys (`allowed_tools`, `disallowed_tools`, `tool_choice`, `permission_mode`) shadowing the typed `tools:` field. Extend `agentFirstClassFields` in `validator/lint_response.go` to include the new tool-access fields so DIP133 fires automatically.

**Sources:** Security Threat-Modeler C2.

### 4. Precedence rules between `tools:`, `disallowed_tools:`, `allowed_tools:` are undefined

Five combinations have no spec answer. The issue body itself names `allowed_tools` as a likely future. By the time `allowed_tools` lands, v0.30.0 workflows have already shipped combining `tools:` + `disallowed_tools:` — precedent set by accident.

**Resolution:** lock the precedence rules in the v0.31 spec even if `allowed_tools` is deferred. Suggested order: `tools:` resolves to a starting catalog → `allowed_tools:` intersects → `disallowed_tools:` subtracts.

**Sources:** Language Critic C3, Future-maintainer #2.

### 5. `read_only` doesn't bound the v0.28.2 vector

Chain: agent A (full tools) writes payload to disk → A's `last_response` auto-injects into B's prompt → B (`read_only`) Reads/Globs/Greps the payload file → B's response launders the payload into trusted text → C's prompt receives it → C executes with full tools.

`read_only` is content laundering, not safety. It also exfiltrates `.env`, `~/.ssh/*`, `.git/config` with embedded tokens directly into the 39k-token output.

**Resolutions:**
- Rename `read_only` to `read_filesystem` or `introspect` — makes threat surface obvious.
- Add explicit doc: "`tools: read_only` constrains side effects, not information flow. Use `tools: none` for pure text-in / text-out summarizers."
- Address `${ctx.last_response}` auto-injection at the language level — the runtime auto-injection bypasses the registry filter entirely. Possibly a new `last_response_truncate:` field on agent nodes, or a structural change to context threading.

**Sources:** Security Threat-Modeler C4 + I1, Language Critic I3.

### 6. `Tools` field name collides with `ToolConfig` and `tool` node kind

`agent.Tools` reads as "downstream tool nodes" — a different concept. Three reviewers flagged this independently. Worst possible name in the package.

**Suggested replacements:**
- Source surface: `tool_access:` (Language Critic), `tool_policy:` (IR Designer), `tools_mode:` (Future-maintainer).
- IR struct field: `LLMTools`, `ToolCatalog`, or `ToolPolicy`.

Future-maintainer also notes: reserve `tools:` as a *name* for a future richer map-shaped field rather than burning it on the enum.

**Sources:** Language Critic I1, IR Designer C1, Future-maintainer closing.

### 7. Fail-open on parse error is the wrong default for a safety field

Invalid `tools: nono` → `cfg.Tools = ""` (full catalog). Author who skips `just check` ships full-catalog nodes labeled with a typo. The proper default for a safety field is fail-closed: invalid value → `cfg.Tools = "none"`. Worth calling out in the spec since current code analogues (`parseInt → 0`, `parseBoolAttr → false`) happen to fail safe-by-luck.

**Sources:** Parser Engineer I2.

### 8. DIP139-in-parser violates established pattern

Parser has no coded-diagnostic infrastructure today — only untyped freeform strings. DIP127 (invalid human mode), DIP130 (invalid response_format), every other enum-validation lint fires in the **validator** with a `validX` map. Putting DIP139 in the parser is net-new infrastructure not in the spec.

**Resolution:** new `validator/lint_tool_access.go` with `validToolPolicies` map; parser stores `cfg.Tools = val` verbatim; lint emits the coded diagnostic.

**Sources:** IR Designer I6, Parser Engineer C1.

### 9. `DisallowedTools []string` contradicts DIP28's `ToolCommandsAllow string`

Same struct neighborhood, same problem shape, different decision. DIP28's rationale ("the runtime owns parsing, IR stays authoring-format-agnostic, avoids drift") still applies to `disallowed_tools`. Two representations of the same concept across two structs is a bug factory.

**Resolutions:**
- Keep both `string` (DIP28's choice — preferred).
- Or change DIP28 to `[]string` in the same PR.
- Or drop `disallowed_tools` from the surface entirely (see below).

**Sources:** IR Designer C2.

### 10. Cross-node coverage holes in `BranchConfig` / `ManagerLoopConfig`

`BranchConfig` (parallel-node branches) and `ManagerLoopConfig` (managed child subgraphs) don't inherit `tools:` from the parent. Parallel agents and managed children have the same v0.28.2 risk profile. Safety primitive with documented holes day-one.

**Resolution:** at minimum an explicit non-goal in the spec; ideally a reserved lint code that fires when a parent's `tools: none|read_only` isn't replicated on children.

**Sources:** IR Designer I3.

## Important findings

| Finding | Sources |
|---|---|
| `tools: all` vs `""` is dead bits — drop `"all"` | Language I2, IR C3, Future #5 |
| DIP140 lints the *least* dangerous combo (belt-and-suspenders); the dangerous combo (`tools: all` + `disallowed_tools: X` opting out of safe defaults) goes unlinted | Security M1 |
| Missing symmetric lint: `tools: read_only` + `disallowed_tools: Read, Glob` makes agent effectively `none` with no diagnostic. This is DIP141's actual home | Language I5, Parser edge case |
| No DOT round-trip parity tests proposed. Security-critical attr survives only by convention. Need `compareAgentConfigs` extension + roundtrip test (matches v0.29.0's `TestRoundtripPreservesToolOutputs` precedent) | Security I3 |
| `disallowed_tools: "Bash, ApplyPatch"` — quotes don't protect commas; single-quotes leak as literal chars | Parser C2 |
| `splitComma` produces empty-string entries on edge inputs — use `splitCommaNoEmpty`; the proposed `parseListAttr` helper is redundant | Parser C3 + I3, IR I4 |
| Casing not normalized — `tools: NONE` silently becomes default-full-catalog | Parser edge case, Security M3 |
| Backend coupling: `tools: none` semantics differ between `backend: native` and `backend: claude-code`. Original spec didn't disambiguate | Security I5 |
| Authoring guidance debt — skill.md would have to hedge ("ineffective until the runtime enforces it") forever | Future #7 |
| `examples/agent_tool_access.dip` must lint-clean; DIP140-tripping example would be permanent CI noise. Mirror the issue body's `ReportFinalStatus` framing | Future #10 |

## Predicted regret (Future-maintainer's closing line)

> *Six months from now: shipped `tools: none` as an authoring surface with no runtime enforcement; no clean way to add deny-by-default-for-unknown-tools without breaking already-shipped workflows whose authors believed the field meant something it didn't. The specific lock-in is the field-name `tools:` — a maximally-generic name that will need to mean something fancier (per-tool config, policy bundles, runtime enforcement levels) within two minors.*

## Suggested starting shape for the v0.31 spec

This is **not** a design — it is a starting point that addresses every critical above. The v0.31 brainstorming should reconsider every choice.

- **Cross-repo scope.** dippin + the runtime as a joint release. No advisory window.
- **Field name.** `tool_access:` (source) / `ToolAccess` (IR) — avoids the `tools`/`tool` collision.
- **Surface.** `tool_access: none | read_filesystem | full`. Drop `"all"`, drop `"read_only"` naming. No `disallowed_tools` in the first cut — defer to a richer follow-up if needed (precedent rules already locked).
- **Defaults cascade.** `defaults: tool_access: read_filesystem` works; per-node `tool_access:` overrides (matches `Model`/`Provider` precedent: full override, no merge).
- **Parser.** Stores `cfg.ToolAccess = val` verbatim; no validation. On unknown value, fail closed to `"none"`.
- **Linter.** New `validator/lint_tool_access.go` with `validToolAccess` map. DIP139 fires on invalid value, DIP140 fires when `Params` shadows `tool_access` (`allowed_tools`/`disallowed_tools`/`tool_choice`/`permission_mode` in `params:`), DIP141 fires when parent of a parallel/manager-loop block has restrictive `tool_access` but children don't.
- **Precedence rules.** Locked: `tool_access:` resolves to a starting catalog. Future `allowed_tools:` intersects. Future `disallowed_tools:` subtracts.
- **Runtime side.** Honors `ToolAccess == "none"` by setting `tool_choice: none` (Anthropic) / equivalent (OpenAI). Filter tool registry for `read_filesystem`. System-prompt scrubbing: when `ToolAccess == "none"`, omit the "File tool arguments MUST use..." line.
- **Cross-node injection.** Out of scope for this primitive, but spec should call out the `${ctx.last_response}` auto-injection vector and link to a separate follow-up issue.

## Raw reports

Each reviewer's full output was saved to `/tmp/claude-1000/.../tasks/{id}.output` as JSONL. They will age out; consolidate any nuance into this doc before they do.

- Language design critic: `aa626c5c8e5e8674c.output`
- Security threat modeler: `a0303cec8bb9ddf02.output`
- IR/API designer: `ac6049503f0f834ae.output`
- Parser/compiler engineer: `aa522dc32ef32774f.output`
- Future maintainer: `a572215e50f30a56d.output`
