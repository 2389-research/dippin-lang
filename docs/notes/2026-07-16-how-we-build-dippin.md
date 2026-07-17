# What's actually made Dippin work

*A reflection on techniques, collaboration, and approach — 2026-07-16.*

If I had to name the single organizing principle behind how we've built Dippin, it isn't speed, cleverness, or any particular tool. It's this: **we treat the author as fallible and build a dense mesh of cheap, early error-catching around them.** "The author" here is usually the AI — and the whole system is arranged to catch mistakes before they reach a consumer, locally and quickly, rather than after they've shipped. Velocity comes not from moving fast but from making errors cheap to catch.

That principle shows up as a handful of concrete techniques.

## Design before code, always

Every non-trivial change went through brainstorm → spec → plan → implement, with a hard gate: no code until a design is presented and approved. This felt like ceremony at first and turned out to be where the real work happens. The gate isn't about documentation — it's about surfacing the *one decision that actually matters* before it's buried in an implementation. For #175 it was the composition model (cascade vs. include vs. token). For #182 it was where to reject unterminated quotes. And most tellingly, for **#186 the design conversation itself revealed that #134 — a headline feature we'd already shipped — was fundamentally wrong.** We never wrote a line of the "fix" before discovering the fix was actually a revert. That only happens if you think structurally before you type.

## Adversarial review by independent squads

After each feature, before landing, we spin up fresh subagents with distinct lenses — prompted not to bless the work but to *refute* it. This is the highest-leverage thing we do, because an author is structurally blind to their own errors. The squads have earned their keep repeatedly: they found dip 2's migration bugs, they found #175's trailing-whitespace pack-parity divergence (a byte-level mismatch across pack modes that would've silently broken the exact "final line" contract the feature existed to serve), and they found a docs corruption we'd introduced. The instruction *"I do not review, you review with competent squads"* was the key structural move: offload review to a mechanism that scales, instead of a human bottleneck, while keeping the decisions that genuinely need a human.

## Reproduce first, gate hard

Bugs start with a reproduction against the *real* built binary, not a mental model. For #182 both failure halves were reproduced before touching code; for #186 the very first action was confirming the bug reproduced against the actual migration code. And the pre-commit gate is ground truth, not advisory: complexity caps (≤5 cyclomatic, ≤7 cognitive) that force small focused helpers instead of `//nolint`, a pinned linter, full test+race. When the gate was once bypassed with `--no-verify`, the correction was flat — *"--no-verify is never admissible"* — and correct; every commit was re-validated and it never happened again.

## Keep every surface current, in the same batch

A standing rule: a language change lands with its docs, site, embedded LLM spec, editor grammars, and examples swept in the *same* change — never as a follow-up. Drift is how a toolchain rots quietly, and it had bitten us (a stale brew formula, downstream surveys reading old docs). Treating the sweep as part of the feature keeps the whole surface honest.

## How the collaboration works

The working relationship has a distinct, effective shape:

- **Decisive, high-level delegation** — "go", "Lets do all that" — trusting the AI to drive long multi-step execution, but with **tight design checkpoints** posed as sharp multiple-choice questions. The human stays out of the mechanics and firmly in the load-bearing decisions.
- **Spot-checks that catch loose ends.** "Is 175 done?" caught code shipped but an issue left open. "I'm confused, I see PR 183 open for 182?" caught a collision that had been walked right past. Delegate execution, but *verify state*.
- **Direct correction, no coddling.** When CI had been red since #136 (GoReleaser is a separate workflow, so releases shipped green while the CI workflow was failing), the feedback was blunt and technical: fix it.
- **The human brings knowledge the AI can't have.** #186 is the clearest case: the tracker engine is a separate repo, and the engine-level analysis ("`OutcomeRetry` never consults `selectEdge`") is what exposed that two disjoint runtime channels were being conflated. The AI's value is dippin-side rigor; the human's is runtime reality and product judgment. The bug report *was* the hard part.

## The honest core

What stands out most is how much of this is oriented around **catching one's own mistakes without ego**. The #183 story is the archetype: a failure to check for an existing PR, a parallel fix implemented and merged, a contributor's work broken — and then the discovery that their tests caught a real bug *ours had missed*. The right response wasn't to defend our version; it was to salvage their tests, land the gap fix, credit them, and apologize. Each lapse became a durable memory — verify the whole CI run, check for existing PRs, never bypass the gate — so the same mistake gets cheaper each time. The system learns.

And that's why #186 is the perfect capstone. We are about to **calmly revert a shipped headline feature** because a careful design conversation proved it wrong — no defensiveness, no sunk-cost, no attempt to save face. That's only possible because everything upstream — the design gates, the squads, the reproductions, the spot-checks — has made "I was wrong, here's the correction" the *normal* state of the system rather than an emergency.

Dippin works because correctness is cheaper than pride here, and we've built the machinery to keep it that way.
