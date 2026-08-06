# Pipeline Inputs — Design

**Status:** Approved design, ready for implementation planning
**Date:** 2026-08-06
**Issues:** dippin-lang [#190](https://github.com/2389-research/dippin-lang/issues/190) (this repo — grammar/IR/validate),
tracker #553 (engine collect/validate/inject), tracker-runner #210 (host collection UI)

---

## Problem

A `.dip` file has no way to declare what a caller must supply. `idea-to-pr` needs a
free-text idea from a human, so authors smuggle it through the `goal` string or an
undeclared `ctx.*` key — no schema, no prompt text, no required/optional semantics, no
validation. Runs proceed with nothing supplied and the agent invents work.

The gap is wider than the entry point. Dippin currently has:

- **A declaration with no namespace** — `vars` carries author-set key/values to the
  runtime, but there is no `${vars.x}` accessor in the language.
- **A namespace with no declaration** — `${params.x}` is readable inside a subgraph, but
  nothing in a `.dip` states which params that file accepts. A `subgraph … params:` call
  site is an argument list against a signature that does not exist.

A top-level run and a subgraph call are the same act with different value sources. One
construct closes both gaps.

## Core decision

**`inputs` is the callee-side signature.** It declares what a `.dip` accepts, whoever is
calling — a human at the entry point, or a parent workflow through `subgraph … params:`.

This is the decision everything else follows from. It yields a compile-time check that is
otherwise unreachable: *subgraph node `Interview` omits required input `topic` declared by
`interview_loop.dip`*. The tracker engine mirrors it at runtime — the subgraph handler
binds a child's `inputs.` namespace from the parent's `params:` map and runs the same
value validation against the child's declared signature. One validation path, two value
sources.

## Syntax

```dip
workflow IdeaToPR
  goal: "Turn a user's idea into a shipped PR."
  start: Plan
  exit: Done

  inputs
    idea: text
      required: true
      prompt: "What do you want built?"
      description: "One or two sentences describing the change."
      multiline: true
      max_length: 4000
    target_branch: text
      default: main
      pattern: "^[A-Za-z0-9._/-]+$"
    spec: file
      default: SPEC.md
    risk: enum
      options: low, medium, high
      default: medium
    dry_run: bool
      default: false
```

A declaration is a header line `name: type`, optionally extended by an indented attribute
block. The minimal form is one line — `idea: text` — which reads like `vars`. The block
appears only when there is metadata to carry.

The shape follows `branch:` inside a `parallel` block, the existing precedent in this
grammar for "header line, optional indented body". The issue's original sketch
(`- name: idea`) is a YAML list-of-maps; dippin has no such construct anywhere, and
introducing one for a single feature would be a new syntactic category to teach, lex,
format, and highlight.

Attribute names are snake_case (`max_length`, not `maxLength`), matching every other field
in the language.

### Grammar additions

```ebnf
workflow_body ::= ( workflow_field | defaults_section | vars_section | inputs_section
                  | node_decl | edges_section | stylesheet_section | NEWLINE )*

inputs_section ::= "inputs" NEWLINE INDENT input_decl* OUTDENT

input_decl ::= IDENTIFIER ":" input_type NEWLINE ( INDENT input_field* OUTDENT )?

/* The named types are the v1 closed set. IDENTIFIER makes the production
   forward-compatible: the parser accepts any type token, and an unrecognized
   one is diagnosed by the validator (DIP155), never by the parser. */
input_type ::= "text" | "number" | "bool" | "enum" | "file" | "secret" | IDENTIFIER

input_field ::= "required" ":" BOOLEAN
              | "default" ":" field_value
              | "prompt" ":" field_value        /* what a host asks the caller */
              | "description" ":" field_value   /* help text */
              | "options" ":" field_value       /* enum: comma-separated choices */
              | "pattern" ":" field_value       /* text: regex the host enforces */
              | "min" ":" field_value           /* number: inclusive lower bound */
              | "max" ":" field_value           /* number: inclusive upper bound */
              | "max_length" ":" INTEGER        /* text: character cap */
              | "multiline" ":" BOOLEAN         /* text: host renders a textarea */
```

### Types — v1 closed set

| Type | Meaning |
|---|---|
| `text` | String. `multiline: true` is an attribute, not a separate type. |
| `number` | Integer or float; the JSON projection emits a JSON number. |
| `bool` | `true`/`false`. |
| `enum` | One of `options:`. |
| `file` | A path. Existence is a host preflight; dippin never touches the filesystem for it. |
| `secret` | `text` plus redaction obligations (see *Secrets*). |

`bundle` (raised in #190) is deferred. Neither side could pin down a concrete meaning that
`file` plus host-side staging does not already cover; confirmed on the engine side in
tracker #553.

### `required` and `default` coexist

`default` is a **form prefill**. `required: true` means the host must obtain a confirmed
value regardless — the engine errors `missing_required` when a required input has no
supplied value even if a default is declared. Defaults auto-fill omitted **non-required**
inputs.

The combination is therefore meaningful and must not be rejected. Confirmed against the
engine's `ValidateInputs` semantics in tracker #553.

### Ordering is significant

`Workflow.Inputs` preserves **declaration order** and the formatter must not sort it. A
host renders these as an ordered form or asks them conversationally in sequence; the
author's order is the intended order. This differs deliberately from `vars`, which the
formatter sorts alphabetically because it has no such semantics.

### Formatter placement

Canonical section order becomes:

```
header (goal, requires, start, exit) → inputs → defaults → vars → nodes → stylesheet → edges
```

`inputs` sits directly after the header because it is the file's contract — the first
thing a reader needs — while `defaults` and `vars` are configuration.

## Namespace: `${inputs.name}`

Declared inputs are read as `${inputs.name}`. Not `${ctx.name}`, not `${params.name}`.

This was the one point of pushback against #190's framing, and it is load-bearing on both
sides of the contract.

**Why not `ctx.`** — `ctx` is an open namespace; anything a node writes lands there.
Folding inputs into it forfeits the trust distinction *and* any ability to flag a
misspelled reference, because dippin cannot know which `ctx` keys are legitimate.

**Why not `params.`** — those are author-set values written by a parent `.dip`, i.e.
author-trusted. Mixing caller input into them destroys precisely the distinction the
engine is being asked to expose.

**What a closed `inputs.` namespace buys:**

- **Structural taint.** Everything under it is caller-supplied by construction, so a host
  cannot forget to frame it as data rather than instructions.
- **A guarantee `ctx.` cannot offer.** `${inputs.x}` where `x` is not declared is a lint
  error (DIP156). This is the only namespace in the language that is closed and checked
  against a declaration.
- **Engine-side enforcement, not just lint.** tracker's shell-interpolation safety is a
  safe-key allowlist in `pipeline/expand.go`; LLM-origin `ctx.*` keys are blocked from
  reaching a shell, and #177 namespaces steer values under `steer.*` for the same reason.
  A closed `inputs.` namespace lets the engine keep the **entire namespace** off that
  allowlist by construction.
- **Composition with the chain-attack family.** "Untrusted input reaches a tool-bearing
  agent" becomes a lintable shape alongside DIP147 and `last_response_truncate`, which it
  cannot be while hidden inside `ctx`.

`${params.x}` keeps working unchanged for today's undeclared pass-through and becomes the
legacy path. One rule to teach: **declared it → read it as `inputs.`; `params.` is the old
undeclared form.** Reconciling the call-site keyword — `params:` binding an `inputs`
signature — is a dip 2 rename, out of scope here.

### Consequence: `${inputs.*}` never interpolates in a tool `command:`

Because the engine keeps the whole namespace off the shell allowlist, an `${inputs.x}`
reference inside a tool node's `command:` or `command_file:` body is **dead text** — it
expands to nothing, silently, for every input type and not only secrets.

Authors will hit this immediately and the failure mode is an empty shell variable rather
than an error, so DIP157 flags it as an error in Phase 1. To get an input into a shell
command, route it through an agent or a declared context key.

## IR

Additive and `omitempty`; a file with no `inputs` block is unchanged in every respect.

```go
// Workflow gains:
Inputs []*Input   // Declaration order preserved; nil when no inputs block

// Input declares a caller-supplied value bound at run start. Values are
// untrusted by construction — see docs/context.md.
type Input struct {
    Name        string
    Type        string   // v1: text|number|bool|enum|file|secret. Unknown types are
                         // carried verbatim and diagnosed by the validator (DIP155).
    Required    bool
    Default     string   // Raw source text; typed at projection time per Type
    HasDefault  bool     // Distinguishes an absent default from an empty-string default
    Prompt      string
    Description string
    Options     []string // enum
    Pattern     string   // text
    Min         string   // number; raw text, typed at projection
    Max         string   // number; raw text, typed at projection
    MaxLength   int      // text
    Multiline   bool     // text
    Source      SourceLocation
}
```

**Why `Default`/`Min`/`Max` are raw strings.** The IR stays text-faithful so the formatter
round-trips a file byte-for-byte; typing happens in the JSON projection, which emits
`number` as a JSON number and `bool` as a JSON bool. The host therefore receives typed
values and injects them typed — the direct fix for tracker-runner's `map[string]string`
unmarshal, which currently fails any run with a non-string input.

## Forward compatibility

#190 asks for unknown types to be a validate error *and* for extensibility that does not
break older consumers. As stated those pull opposite ways: an older dippin erroring on a
type a newer one introduces **is** breaking an older consumer.

It resolves by layering, following the existing `emitUnknownFieldHint` precedent:

- **Parser** never fails on an unknown type or an unknown input attribute. It records the
  value verbatim into the IR.
- **Validator** emits an error-severity diagnostic for an unknown type (DIP155) and a hint
  for an unknown attribute (parser diagnostic, no DIP code — matching how unknown node
  fields behave today).

A `.dip` using a future `type: duration` therefore still parses, formats, round-trips, and
packs on an older dippin. Only the lint complains. Both requirements are satisfied without
compromise.

## Division of labor

**dippin lints the declaration. The engine validates values.** dippin never sees a value,
so `ValidateInputs` lives entirely on the tracker side. Confirmed in tracker #553.

### Diagnostics

Next free code is DIP155.

| Code | Severity | Rule | Phase |
|---|---|---|---|
| DIP155 | Error | Unknown input type | 1 |
| DIP156 | Error | `${inputs.x}` references an undeclared input | 1 |
| DIP157 | Error | `${inputs.*}` inside a tool `command:` / `command_file:` — never interpolates | 1 |
| DIP158 | Error | Invalid or inapplicable constraint: enum `default` not in `options`, `min` > `max`, malformed `pattern` regex, or a constraint on a type that has none (e.g. `max_length` on a `bool`) | 2 |
| DIP159 | Warning | Declared input never referenced anywhere — dead input (mirrors DIP107) | 2 |
| DIP160 | Warning | Subgraph `params:` omits a required input of the referenced child (cross-file) | 3 |
| DIP161 | Warning | Untrusted input flows into a tool-bearing agent (DIP147 family) | 3 |

DIP157 subsumes the narrower "secret into a shell command" rule originally proposed; a
`secret` reference in that position is the same defect with a sharper help message, not a
separate code.

**DIP156 must cover two scan paths.** Prompt bodies go through `lint_context.go` (the
DIP106 path); edge conditions go through `lint_conditions.go` (the DIP120 path). Both
currently consult `knownNamespaces` for prefix validity only. `inputs.` is the first
namespace requiring closed-set key checking, so each path needs it — a shared helper
resolving a reference against `w.Inputs`.

`knownNamespaces` gains `"inputs"`, but prefix membership alone is no longer sufficient
for this namespace.

## Introspection

The single most important export for the downstream host (tracker #553, ask 1).

**JSON projection.** A stable schema document a host can render directly into a form or
walk conversationally. Per input: name, type, required, default (typed), prompt,
description, and constraints. Declaration order preserved.

**CLI surface:** `dippin inputs <file.dip> [--format=json|text]`. A new noun command
matching `dippin cost` / `dippin coverage` / `dippin unused`. `dippin inspect` is
bundle-scoped and cannot serve this — but it should additionally surface the **entry
workflow's** declared inputs when inspecting a `.dipx`, so a host can enumerate what to
collect without unpacking.

**Go API:** `w.Inputs` is exported IR. No accessor method is needed; the field is the API.

## Secrets

dippin's obligations: mark the input as `secret`, refuse to inline its value anywhere
(dippin never has a value to inline), and flag a `secret` reference reaching a shell
command via DIP157.

The non-persistence guarantee belongs to the host: redaction in the activity log and
`--json` stream, exclusion from the checkpoint snapshot and the ExportBundle run
directory. Confirmed in tracker #553 — neither side writes a secret value to disk in the
clear.

## Fulfillment timing

**Run-start binding, no mid-run fulfillment.** Inputs are the run's signature, fixed
before the first node executes. A `human` node does **not** satisfy an input — human gates
produce `human_response` / interview answers into `ctx.*`, an orthogonal mechanism. The
`needs_input` park is pre-run collection in the host, not an engine pause.

dippin can therefore assume inputs are bound before execution and needs no syntax for
mid-run collection.

## Backward compatibility

- A `.dip` with no `inputs` block behaves identically — no IR change, no formatter change,
  no new diagnostics.
- Purely additive to dip 1. No version bump, no migration.
- `requires:` is untouched. It means "tools on PATH" and answers a different question;
  repurposing it would break whoever reads `Workflow.Requires` today.
- `${params.x}` is untouched.

## Implementation notes

**Complexity budget.** Field application must respect cyclomatic ≤ 5 / cognitive ≤ 7. A
single switch over ten attributes will not fit, so input field application splits along
the existing `applyCommonStringField` / `applyCommonComplexField` pattern — plain string
fields, boolean/integer fields, and constraint fields as separate helpers each returning
`bool`.

**Package boundaries.** `parser` and `formatter` import `ir` only. The DIP160 cross-file
check belongs with the existing cross-file machinery at the CLI layer
(`crossfile_tool_access.go` is the precedent), not inside `validator`.

## Phasing

**Phase 1 — unblocks tracker #553**
Grammar, IR, parser, formatter, DIP155/156/157, the JSON projection, and
`dippin inputs`. Plus the standing surface sweep: `docs/GRAMMAR.ebnf`, generated spec,
`docs/syntax.md`, `docs/context.md`, tree-sitter, LSP, editors, `skill.md`, site. Ships
with a worked example `.dip` under `examples/`.

**Phase 2** — DIP158, DIP159.

**Phase 3** — DIP160 (cross-file), DIP161 (chain-attack integration).

The engine team can build their introspection/validation/typed-injection contract against
their own `InputSpec` now and land the adapter the moment the IR carries `Inputs`.

## Testing

- Parser round-trip: parse → format → parse is byte-identical, including declaration order
  and an unknown type carried verbatim.
- Minimal form (`idea: text`, no block) and full form both parse.
- A file with no `inputs` block produces a nil `Inputs` and unchanged output.
- Each diagnostic gets a positive and a negative fixture.
- DIP156 is exercised through **both** a prompt body and an edge condition.
- `TestLintExamples` covers the new example, asserting zero warnings.
- JSON projection: typed defaults (`number` → JSON number, `bool` → JSON bool), absent vs.
  empty-string default, declaration order preserved.

Fixtures are built by parsing real `.dip` source, never by hand-assembling IR — the DIP101
regression came from hand-populated fields masking that production code never set them.

## Out of scope

- Value validation, collection, and injection — tracker #553.
- Host collection surfaces — tracker-runner #210.
- `bundle` type.
- Mid-run input fulfillment.
- Renaming the subgraph call-site keyword from `params:` to `inputs:` — dip 2.
- A `${vars.x}` accessor. `vars` remaining author-set with no accessor will read as
  arbitrary once `inputs` exists beside it; that deserves its own issue rather than
  scope creep here.
