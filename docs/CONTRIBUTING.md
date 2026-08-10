# Documentation Accuracy Protocol

Every doc targets a specific persona. Before editing, know who you're writing for.

## Doc Personas

| Document | Persona | What they need |
|----------|---------|---------------|
| `README.md` | Developer evaluating dippin | Quick start, feature overview, "can it do X?" |
| `docs/syntax.md` | Developer writing .dip files | Every syntax rule, complete examples |
| `docs/nodes.md` | Developer configuring nodes | Every field, every type, every default |
| `docs/edges.md` | Developer writing routing logic | Conditions, operators, routing priority |
| `docs/context.md` | Developer debugging variable flow | What variables exist, who sets them, namespaces |
| `docs/validation.md` | Developer fixing lint warnings | What each DIP code means, how to fix it |
| `docs/cli.md` | Developer using CLI or scripting | Every command, every flag, output format |
| `docs/analysis.md` | Developer interpreting analysis output | JSON schemas, when to use each command |
| `docs/editor-setup.md` | Developer configuring their editor | Copy-paste configs for VS Code, Neovim, etc. |
| `docs/llm-reference.md` | LLM generating .dip files | Compact reference for system prompts |
| `docs/architecture.md` | Developer contributing to dippin-lang | Package map, dependency graph, design decisions |
| `docs/integration.md` | Developer consuming dippin as a Go library | API examples, import paths |
| `docs/GRAMMAR.ebnf` | Developer understanding the parser | Canonical syntax spec, must match parser exactly |
| `docs/testing.md` | Developer writing scenario tests | Scenario testing, `dippin test` command |
| `CHANGELOG.md` | Developer upgrading versions | What changed, what broke, what to migrate |

## Pre-Commit Doc Checklist

When changing code that affects observable behavior, check these before committing:

1. **New/changed fields** → update `nodes.md`, `syntax.md`, `README.md` field table, `docs/GRAMMAR.ebnf`
2. **New/changed DIP codes** → update `validation.md`, `README.md` diagnostics table, `llm-reference.md`
3. **New/changed CLI commands** → update `cli.md`, `README.md` commands table
4. **New/changed operators** → update `syntax.md`, `edges.md`, `llm-reference.md`, `docs/GRAMMAR.ebnf`
5. **New/changed providers or models** → verify against official provider docs, then edit the single source of truth **`pricing/prices.json`** (embedded by the leaf `pricing` package; both `cost` and `validator`'s DIP108 catalog derive from it — do **not** hand-edit `cost/pricing.go` or `validator/lint_model.go` model tables). Set each entry's `source` URL and `as_of` date; use `"priced": false` for a recognized-but-unverifiable model. Run `just check-prices` to surface overdue entries and `just sync-prices` to diff against models.dev. Update `docs/pricing-integration.md` and the supported-provider list if a provider is added.
6. **New analysis output** → update `analysis.md` JSON schemas
7. **Architecture changes** → update `architecture.md` package map and dependency graph
8. **Version release** → update `CHANGELOG.md`, create GitHub release with notes

## Accuracy Verification

When auditing docs for accuracy:

1. **EBNF vs parser**: Every production rule in `docs/GRAMMAR.ebnf` must match the parser implementation. Run: read each EBNF rule, find the corresponding parser code, verify they match.

2. **Field tables vs IR**: Every field listed in `nodes.md` and `README.md` must exist in `ir/ir.go` config structs. Every config struct field should be documented.

3. **Operators vs evaluator**: Every operator listed in docs must be implemented in `simulate/defaults.go` `operatorFuncs` and `simulate/condition.go` `validOperators`. No operator should be documented that silently fails at runtime.

4. **DIP codes vs implementation**: Every DIP code in `validation.md` must have a corresponding check function in `validator/`. The code constant in `codes.go`/`lint_codes.go` must exist.

5. **CLI commands vs implementation**: Every command in `cli.md` must have a handler in `cmd/dippin/`. Flags must match.

6. **JSON schemas vs struct tags**: Every field in `analysis.md` JSON examples must match the `json:` struct tags in the corresponding Go types.

7. **Model catalog vs reality**: Model names and pricing live in `pricing/prices.json` and must be verified against official provider documentation. Each entry's `source` URL must be current and its `as_of` date updated on verification (`just check-prices` flags overdue entries). `cost/pricing.go` and `validator/lint_model.go` derive from this file — they carry no hand-maintained model table.
