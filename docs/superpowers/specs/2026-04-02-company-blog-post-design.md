# Company Blog Post: Why We Built a Language for AI Pipelines — Design Spec

## Overview

A ~2000-word blog post for the 2389 Research company blog. Origin story arc with before/after code woven in. Audience: customers, investors, partners, and engineers at other companies (mixed technical). Storytelling tone with technical substance.

## Narrative Arc

Credibility (we actually built this) → Thesis (pipelines deserve real tooling) → Payoff (here's what good DX gets you)

## Title Candidates

Pick the best during implementation:
- "Why We Built a Language for AI Pipelines"
- "From Escaped Strings to a Toolchain: Building Dippin"
- "AI Pipelines Deserve Better Than Config Files"

---

## Section 1: Opening (~200 words)

Start with a concrete moment: someone editing a 20-line system prompt inside a DOT file. Escaped quotes, `\n` everywhere, a JSON block requiring `\"` on every key. A one-character mistake silently breaks the pipeline at runtime — no validation, no error, just wrong output.

Step back: We build an AI pipeline orchestration system. Pipelines are directed graphs — nodes, edges, prompts, models, conditions. For two years, DOT was the authoring format. It worked until it didn't.

Thesis: "When your pipeline definitions get complex enough — multi-model consensus, human gates, retry loops, tool nodes with shell scripts — the authoring format becomes either an accelerator or a bottleneck. DOT had become our bottleneck."

## Section 2: The Breaking Point (~300 words)

Three pain points, each with a brief before/after code snippet:

**Multi-line prompts**: Show real DOT `tool_command` with `\n` and `\"` escapes (from `examples/sprint_exec.dot` line 15). "Every prompt edit was an exercise in escape-character archaeology." Then show the Dippin equivalent — clean, indented, readable.

**No validation**: "DOT is a graph description language. It doesn't know what an AI pipeline is. Misspell a model name? Reference a node that doesn't exist? You find out at runtime, in production." No way to lint, test, or catch mistakes before deployment.

**No testing story**: "We had no way to write deterministic tests for non-deterministic pipelines. Every change was a manual spot-check against a running system."

**The tipping point**: "We'd spent more time debugging escaped strings than writing prompts. That's when we knew DOT wasn't the right tool anymore."

## Section 3: What We Built (~500 words)

Introduce Dippin — "a domain-specific language purpose-built for AI pipeline workflows." Walk through DX features as narrative, not feature list. Each flows from the pain points above.

**The language**: Indentation-based blocks, node types (agent, tool, human, subgraph), edges block. Show a complete ~15-line workflow (Draft→Review→Publish with conditional edges). Reader should understand it at a glance — that's the point.

**39 diagnostics**: "Misspell a model name? DIP108. Forget a timeout? DIP111. Missing namespace prefix? DIP120." Every diagnostic has a code, explanation, and fix suggestion. 9 structural errors vs 30 semantic warnings.

**Scenario testing**: "We can now write deterministic tests for non-deterministic pipelines." Inject context values, assert visited nodes, check execution paths. Testing AI pipelines is a solved problem for us.

**Cost estimation**: "`dippin cost` tells us what a run costs before we spend money. `dippin optimize` suggests cheaper models." Real numbers: "Our code review pipeline runs ~$0.65 expected, $2.66 worst case."

**The toolchain**: Quick hits — LSP server, WASM playground, `dippin watch`, tree-sitter grammar, semantic diff, DOT migration tool. Once you have a proper language, tooling follows naturally.

**Structured output**: "Last week downstream consumers needed to force LLM APIs to return JSON instead of prose. In DOT, this would have required a new attribute convention, no validation, and hope. In Dippin, we added `response_format` and `response_schema` as first-class fields with four lint rules — and the runtime adapter picked them up with zero changes." The payoff moment: the language makes the whole system more capable.

## Section 4: The DX Payoff (~300 words)

Zoom out from features to impact.

**Velocity**: Pipeline authors think about logic, not escape characters. New team members read `.dip` files and understand them. CI runs `dippin check` on every push.

**Confidence**: From "deploy and pray" to 39 diagnostics, scenario testing with edge coverage, and cost estimation. Changes land validated, sound, and within budget.

**The feedback loop**: Downstream consumers file a request. We add a field. Parser, formatter, linter handle it. Adapter picks it up. Turnaround from "we need this" to "it's in production" went from weeks to days. Structured output is the concrete example.

**Open source**: Dippin is open source. Link to repo, playground, docs. "We built this for ourselves, but the problem isn't unique to us. Anyone building multi-step LLM pipelines is dealing with some version of the same authoring pain."

**Closing line**: Forward-looking but grounded. Not "changing the world" — more like: "The best developer tools are the ones that disappear — they let you think about the problem, not the format."

---

## Code Examples to Include

1. DOT `tool_command` with escapes (from `examples/sprint_exec.dot:15`) — the "before"
2. Same content as Dippin `command:` block (from `examples/sprint_exec.dip:17-24`) — the "after"
3. A complete small workflow showing conditional edges (~15 lines, Draft→Review→Publish pattern)
4. `dippin cost` terminal output (from `examples/complexity_cleanup.dip` — real numbers)

## What NOT to Include

- Deep runtime architecture (keep it "our AI pipeline orchestration system")
- Implementation details of the parser/formatter/IR (no one cares about the Go code)
- Feature catalog style listing (weave into narrative)
- Pricing details of LLM providers
- Anything that reads as a product pitch

## Format

- Company blog HTML or markdown (determine during implementation)
- No syntax highlighting spans (this is a company blog, not the Dippin docs site)
- Code blocks use standard markdown fencing
- ~2000 words total
