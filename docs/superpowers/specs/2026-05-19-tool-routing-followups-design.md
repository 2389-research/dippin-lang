# Tool-routing follow-ups (issues #42, #43, #44) — v0.29.0

**Date:** 2026-05-19
**Closes:** #42, #43, #44
**Release target:** v0.29.0
**Predecessor:** v0.28.0 shipped `marker_grep:` / `route_required:` / `output_limit:` (#39); this batch is the cleanup pass flagged in that spec as out-of-scope.

## Motivation

`marker_grep:` made `ctx.tool_marker` an authoritative routing channel on tool nodes. That changes the lint and round-trip expectations the toolchain inherited from the pre-#39 world:

1. **Lint false positives (#42).** A clean marker-routed workflow with single-conditional edges per marker value still trips DIP101 / DIP102 today because the existing exhaustive-source detection only knows about `ctx.outcome` pairs and equality partitions. Authors who follow the "TRK101 option (d)" guidance — enumerate every marker as a separate conditional edge — get false-positive coverage hits.
2. **Foot-gun bool parsing (#43).** The new `route_required:` field uses the existing `val == "true"` strict equality pattern shared with `goal_gate`, `auto_status`, `cache_tools`. Anything other than the literal `"true"` silently becomes `false` — `route_required: yes` parses as `false` and the safety net is disabled with no warning. This was acceptable for `goal_gate` (failure mode is conservative); it is not acceptable for `route_required`.
3. **`outputs:` dropped on DOT round-trip (#44).** The runtime exports a workflow to DOT to talk to legacy tooling. The export side never emitted `outputs`, and the migrate side never read it back. A `.dip → DOT → .dip` round-trip silently dropped the field, so coverage / output-set analysis on the migrated copy ran against an empty set.

Each is independently small. Together they round out the v0.28.0 tool-routing surface.

## Scope

Three independent code paths, bundled as one PR + version bump because each alone is a patch-level change and only #42 (which adds a reserved lint code) earns the minor bump.

### #42 — `marker_grep:` suppresses DIP101 / DIP102 on the source tool node

A tool node that declares `marker_grep:` has an authoritative typed routing channel (`ctx.tool_marker`). Outgoing edges that route on this marker are intentional routing, not unsafe reachability. Extend the existing "safe source" predicate in `validator/lint_reachability.go`:

- A node is **safe** today if `exhaustiveSources[id]` (outgoing conditions form a known/inferred partition) OR `hasUnconditionalEdge(outgoing[id])`.
- Add a third arm: **a `ToolConfig` source with non-empty `MarkerGrep`** is safe.

The change applies to both DIP101 propagation (`sourceIsSafe` in the reachability path) and DIP102 (the direct check on `hasMissingDefault`). DIP102 must guard against the new case symmetrically; otherwise a single-edge marker-routed node would skip DIP101 but still trip DIP102.

**DIP138 — reserved only.** Add the constant + `CodeDescription[DIP138]` + an entry in `explanations.go` describing the future advisory ("tool node parses stdout for routing but declares no marker_grep / outputs"). No firing logic in this PR. This keeps the PR focused while reserving the code so it isn't reused.

### #43 — `parseBoolAttr` helper, applied to all four bool fields

Add `func (p *Parser) parseBoolAttr(val, key string, loc ir.SourceLocation) bool` to `parser/parse_helpers.go`. Algorithm:

1. `s := strings.ToLower(strings.TrimSpace(val))`
2. truthy set: `{"true", "1", "yes", "on"}` → return `true`
3. falsy set: `{"false", "0", "no", "off"}` → return `false`
4. anything else: append diagnostic `invalid boolean %q for %s at %d:%d (use true/false, 1/0, yes/no, on/off)`, return `false`

Migrate four call sites in `parser/parse_nodes.go`:

- `applyAgentBoolField` — `goal_gate`, `auto_status`, `cache_tools` (currently three `val == "true"` lines)
- `applyToolBoolField` — `route_required` (one `val == "true"` line)

Both helper functions need their signatures widened to receive `p *Parser` and `loc ir.SourceLocation`. The wrappers `applyAgentField` / `applyToolField` already plumb these values.

**Migrate side is out of scope.** `migrate/migrate.go applyToolStringAttrs` parses DOT attrs that we emit, so the input is always machine-generated `"true"` / `"false"`. Strict equality is acceptable on the DOT path. Document this as an explicit non-change in the CHANGELOG.

### #44 — `outputs:` round-trips through DOT

Two missing pieces; one (parity comparison) already shipped during v0.28.0.

- `export/dot.go applyToolSemanticAttrs` (line ~304): emit `outputs = strings.Join(cfg.Outputs, ",")` when len > 0.
- `migrate/migrate.go applyToolStringAttrs` (line ~508): read `outputs` back via the existing `splitComma`-style helper.
- `migrate/parity.go compareToolSlices` (line ~306) **already** compares `Outputs`, and `TestCompareToolConfigsDifferentOutputs` already exists. Issue #44's body claims otherwise — stale at filing time; v0.28.0 shipped the parity equality. The spec acknowledges this and skips that step.

## Tests

### #42
- `validator/lint_test.go` — `TestLintMarkerGrepSuppression`: parse a real `.dip` snippet with a tool node declaring `marker_grep:`, single conditional outgoing edge `tool -> X when ctx.tool_marker = foo`. Assert zero DIP101 / DIP102.
- Negative case in the same table: same workflow without `marker_grep:` — expect DIP101 + DIP102 to fire.
- Existing exhaustive-conditions cases (success/fail, partitions, complementary pairs) must still pass — no regressions.

### #43
- New `parser/parse_bool_test.go`. Table-driven. Each row:
  - field name (one of the four)
  - value to assign
  - expected bool result
  - expected diagnostic substring (`""` if the form is valid)
- Cover for each of the four fields: at least one accepted truthy form, one accepted falsy form, one invalid form (`maybe`, `2`).
- Use the real parser to parse minimal `.dip` text (per CLAUDE.md: tests must parse real `.dip`, no hand-built IR).

### #44
- `migrate/roundtrip_test.go` — add `TestRoundtripPreservesToolOutputs`: read `examples/marker_routing.dip` → export DOT → import DOT → assert `cfg.Outputs == ["tests_green", "tests_red"]` and DOT contains `outputs="tests_green,tests_red"`.

## Docs

- `CHANGELOG.md` — new `## [v0.29.0] — 2026-05-19` entry following the v0.27/v0.28 precedent. Sections: Added (DIP138 reserved, outputs DOT round-trip), Changed (DIP101/DIP102 suppress on marker_grep, parseBoolAttr widens accepted forms), Closed (#42 #43 #44).
- `docs/nodes.md` — "Markers and Verbose Output" section: note that declaring `marker_grep:` suppresses DIP101/DIP102 on the source node. Mention that bool fields accept `true/false`, `1/0`, `yes/no`, `on/off`.
- `docs/llm-reference.md` — common-mistakes table: cross-reference the suppression behavior.
- `site/static/skill.md` — same updates as `docs/nodes.md` (this is the hosted Claude skill).
- `cmd/dippin/generated-spec.md` — regenerates on commit via the existing hook.

## Non-goals

- Implementing DIP138 firing logic (reserved only).
- Loosening bool parsing on the migrate (DOT-input) path — DOT attrs are machine-emitted and known canonical.
- Refactoring `findExhaustiveSources` to take the new marker_grep case (it stays focused on conditions; the marker check is a separate predicate composed at the call sites).

## Backwards compatibility

- `parseBoolAttr` is strictly **more** permissive than today. Existing `.dip` files written with literal `true` / `false` are unchanged. Files that wrote `yes` and quietly got `false` will now flip to `true` (the user's likely intent). Files with garbage values now get a diagnostic where previously they got a silent `false`. The latter is the only observable behavior change for already-parsing files; the spec accepts this trade in exchange for fixing the `route_required: yes` foot-gun.
- DIP101 / DIP102 only stop firing on tool nodes that declare `marker_grep:`. Pre-#39 files cannot have this field, so this change is invisible to them.
- DOT export now emits an additional `outputs="…"` attr on tool nodes that declare outputs. Consumers that ignore unknown attrs (which is most DOT consumers) are unaffected.
