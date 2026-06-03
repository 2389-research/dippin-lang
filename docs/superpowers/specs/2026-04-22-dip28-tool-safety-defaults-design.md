# DIP28 — Tool safety defaults in `WorkflowDefaults`

**Issue:** [#28 — feat(ir): add ToolCommandsAllow to WorkflowDefaults](https://github.com/2389-research/dippin-lang/issues/28)
**Status:** Design approved 2026-04-22
**Runtime enforcement:** Runtime consumes `graph.Attrs["tool_commands_allow"]`

## Problem

The runtime reads tool-node shell-command allowlists and denylist additions from two DOT graph attributes:

- `tool_commands_allow` — comma-separated glob allowlist (e.g., `"git *,make *"`).
- `tool_denylist_add` — comma-separated globs appended to the runtime's default denylist.

`.dip` authors currently cannot reach either from the `defaults:` block:

1. `ir.WorkflowDefaults` (`ir/ir.go:39`) is a closed struct with no fields for these.
2. `parser/parse_defaults.go` emits `"unknown defaults field %q"` for anything outside the known key set.

Result: to use the allowlist today, `.dip` authors must either drop to DOT or set `graph.Attrs` via the runtime's library API. This defeats the purpose of `.dip` as the authoring format.

## Scope

Ship both `tool_commands_allow` and `tool_denylist_add` together. The allowlist matches the just-shipped runtime; the denylist pairs symmetrically and avoids a second round of parser/export/migrate churn. Landing both together exercises the new plumbing once rather than twice.

Out of scope:

- Validating glob patterns. The runtime owns runtime semantics; IR stores strings verbatim.
- Emitting any *other* typed `WorkflowDefaults` fields (`Model`, `Provider`, `MaxRetries`, …) to DOT graph attributes. That gap pre-exists; fixing it is a separate issue.
- Lint rules (DIP codes) for the new keys. If field misuse becomes a real problem, file a follow-up.
- A blog post — this is operational plumbing, not an authoring-visible feature.

## Design

### IR (`ir/ir.go`)

Add two fields to `WorkflowDefaults`:

```go
ToolCommandsAllow string // Comma-separated glob allowlist (e.g., "git *,make *")
ToolDenylistAdd   string // Comma-separated glob patterns added to the runtime's default denylist
```

String rather than `[]string` — matches the runtime's existing `graph.Attrs[…]` contract, keeps the IR authoring-format-agnostic, and avoids parse-time normalization that could drift from the runtime's parser. Whitespace-around-commas handling stays runtime-side.

### Parser (`parser/parse_defaults.go`)

Add a leaf handler alongside the existing `applyDefaultCoreField` / `applyDefaultExtraField`:

```go
// applyDefaultToolField handles tool-safety defaults.
func applyDefaultToolField(d *ir.WorkflowDefaults, key, val string) bool {
    switch key {
    case "tool_commands_allow":
        d.ToolCommandsAllow = val
    case "tool_denylist_add":
        d.ToolDenylistAdd = val
    default:
        return false
    }
    return true
}
```

Wire it into `applyDefaultStringField` after the existing `applyDefaultExtraField` call. No validation — strings pass through verbatim.

### DOT export (`export/dot.go`)

Current state: `writeDOTHeader` emits only `w.Vars`, not typed defaults. To satisfy the round-trip AC, add a focused emit step after the vars block:

```go
if w.Defaults.ToolCommandsAllow != "" {
    fmt.Fprintf(b, "  %s=%s;\n", "tool_commands_allow", dotQuote(w.Defaults.ToolCommandsAllow))
}
if w.Defaults.ToolDenylistAdd != "" {
    fmt.Fprintf(b, "  %s=%s;\n", "tool_denylist_add", dotQuote(w.Defaults.ToolDenylistAdd))
}
```

Add both keys to `reservedGraphAttrs` (`export/dot.go:60`) so they don't collide if a user also has them in `vars:`.

Not touching the existing `Model`/`Provider`/`MaxRetries` emission gap — that's pre-existing and out of scope.

### DOT → dip migrate (`migrate/migrate.go`)

Add two entries to the `graphDefaultsHandlers` map:

```go
"tool_commands_allow": func(v string, w *ir.Workflow) { w.Defaults.ToolCommandsAllow = v },
"tool_denylist_add":   func(v string, w *ir.Workflow) { w.Defaults.ToolDenylistAdd = v },
```

Routes DOT graph attrs into `defaults:` on reverse migration, matching the forward export path.

## Testing

Four layers:

1. **Parser unit test** — parse a fixture with both keys in `defaults:`; assert raw string values on `Defaults.ToolCommandsAllow` / `Defaults.ToolDenylistAdd`.
2. **Export unit test** — construct a `Workflow` with both fields set; assert `ExportDOT` output contains `tool_commands_allow="git *,make *"` and the denylist equivalent as graph attrs.
3. **Round-trip integration test** — `.dip` → IR → DOT → `migrate` → IR; fields survive unchanged.
4. **Example `.dip`** — add `examples/tool_safety.dip` exercising both keys against a tool node. `TestLintExamples` and `just validate-examples` then guard against regression.

Happy path + "empty string stays empty" edge case per layer. No validation tests — there is no validation logic to cover.

## Release

- `CHANGELOG.md` — entry under "Unreleased" noting both new keys, with a pointer to the runtime release for runtime semantics.
- `QUICK_REFERENCE.md` — add both keys to the defaults-block field list.
- `site/content/language.md` — add a "Tool safety" subsection to the defaults reference with a short example.
- `site/static/skill.md` — add one-line descriptions for both keys so Claude Code authoring sessions know they exist.
- `site/static/llms-full.txt` — regenerated from docs (no manual edit).
- Version bump — tag `v0.23.0` after merge so downstream consumers can pin.

## Acceptance criteria

- [ ] `ir.WorkflowDefaults.ToolCommandsAllow string` and `ir.WorkflowDefaults.ToolDenylistAdd string`
- [ ] Parser accepts `tool_commands_allow: "pattern,pattern"` and `tool_denylist_add: "pattern,pattern"` in `defaults:`
- [ ] DOT export emits both as graph attributes when set
- [ ] DOT → dip migrate routes both back into `defaults:`
- [ ] Round-trip test: parse `.dip` → IR → export → DOT → migrate → IR preserves both fields
- [ ] `CHANGELOG.md`, `QUICK_REFERENCE.md`, `site/content/language.md`, `site/static/skill.md` updated
- [ ] `examples/tool_safety.dip` passes `just validate-examples`
- [ ] `just check` green
- [ ] Tagged as `v0.23.0`

## Runtime integration

Already documented in the runtime's adapter layer per issue #28. Once this lands and is tagged, the runtime bumps its dippin-lang dependency and adds the one-liner there.
