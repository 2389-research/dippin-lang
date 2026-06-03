# Blog Posts #6, #7, #8 — Design Spec

Three blog posts using a hub-and-spokes model: one prospect-facing hub post that tees up two existing-user tutorials.

## Structure: Hub and Spokes

- Post #1 (hub) introduces Dippin through the multi-line prompt lens, tees up posts #2 and #3
- Posts #2 and #3 open with a one-line callback to the hub but are fully self-contained
- Sequential readers get narrative momentum; direct-link readers lose nothing

---

## Post 1: Multi-line Prompts Without Escaping

**Blog ID**: #17 from blog-ideas.md (Deep Dives)
**File**: `site/blog/multi-line-prompts.html`
**Category tag**: Deep Dive
**Audience**: Prospects evaluating Dippin
**Length**: ~800-1000 words
**Tone**: Conversational. "Here's your pain. Here's the fix."

### Flow

1. **Hook** — Open with a real DOT prompt: 4-5 lines of escaped quotes, `\n` characters, nested JSON. One paragraph: "If you've authored AI pipelines in Graphviz DOT, this looks familiar. And painful."
2. **The fix** — Show the same prompt in Dippin. Clean, indented, readable. Explain the mechanic: indent after `:`, everything is literal.
3. **What's preserved** — Blank lines, markdown, code blocks, shell scripts, JSON, `#` characters (not comments inside prompts). The formatter round-trips it exactly.
4. **"But it's not just prompts"** — Quick showcase of two more capabilities:
   - **Subgraph composition**: embed `interview_loop.dip` inside `api_design.dip` with params. Show 3-4 lines of syntax, not a full walkthrough.
   - **The playground**: "Paste this example and try it yourself" with a link to the WASM playground.
5. **CTA** — Install command (`go install` / Homebrew), link to Getting Started post, tease: "Next up: conditional edges and cost estimation."

### DOT Example Source

Use a real prompt from one of the migrated pipelines in `examples/`. During implementation, grep the `.dot` files for the longest escaped `prompt=` string — ideally one with embedded JSON or markdown. If no sufficiently painful example exists in the repo, construct a representative one based on real DOT patterns.

### SEO

- Title: "Multi-line Prompts Without Escaping — Dippin Blog"
- Description: "DOT's escaped strings are unreadable. Dippin's indentation-based blocks let you write prompts with markdown, JSON, and code blocks — no escaping required."

---

## Post 2: Conditional Edges: Routing Pipelines with `when`

**Blog ID**: #6 from blog-ideas.md (Tutorials)
**File**: `site/blog/conditional-edges.html`
**Category tag**: Tutorial
**Audience**: Existing Dippin users
**Length**: ~1200-1500 words
**Tone**: "Let's build this together" — each step shows .dip snippet + lint/simulate output

### Flow

1. **Opening** — One-line callback: "In the last post we showed how Dippin handles prompts. Now let's make pipelines that think." Straight into building.
2. **Linear baseline** — A 3-node pipeline: `Draft -> Review -> Publish`. Works, but what if the review fails?
3. **Basic condition** — Add `when ctx.outcome = success` on the Review->Publish edge and `when ctx.outcome = fail` with a retry edge back to Draft. Show the .dip, run lint, show clean output.
4. **Operators** — Introduce progressively with one-liner examples each:
   - `contains` / `not contains`
   - `startswith` / `endswith`
   - `!=`
   - `in` (value-in-set)
5. **Compound conditions** — `and` / `or` / `not`, parentheses for precedence. One example combining two conditions.
6. **`auto_status: true`** — Let the LLM self-assess. Show how it populates `ctx.outcome`. Brief: 2-3 sentences + snippet.
7. **Exhaustive detection** — Why `success`/`fail` pairs suppress DIP101/DIP102. Show lint output with only one edge (warning), then with the complementary edge (clean). Same for `contains X` / `not contains X`.
8. **Common mistake** — Forgetting the `ctx.` prefix (DIP120). Show the warning, show the one-character fix.
9. **Teaser** — "What does this cost? Next post."

### Running Example

Build a single workflow across all sections — don't switch examples. Start simple, add complexity. The reader should feel the workflow growing under their hands.

### SEO

- Title: "Conditional Edges: Routing Pipelines with when — Dippin Blog"
- Description: "Build branching AI pipelines that route based on LLM output. Learn Dippin's condition syntax, operators, and exhaustive detection."

---

## Post 3: Cost Estimation: Know Before You Run

**Blog ID**: #12 from blog-ideas.md (Tutorials)
**File**: `site/blog/cost-estimation.html`
**Category tag**: Tutorial
**Audience**: Existing Dippin users
**Length**: ~800-1000 words
**Tone**: Practical, dollars-and-cents. The tool does the talking.

### Flow

1. **Opening** — One-line callback: "You've got a branching pipeline with conditional edges. Before you run it against real APIs, let's find out what it'll cost."
2. **`dippin cost`** — Run it on the conditional workflow from post #2 (or a similar branching workflow if full self-containment is needed). Show per-node and total breakdown. Reproduce realistic terminal output.
3. **How estimation works** — Prompt length maps to input tokens, heuristic for output. Brief — 2-3 sentences, not implementation detail.
4. **`max_turns` multiplier** — Agentic loops cost more per iteration. Show how a retry loop multiplies the estimate. Before/after adding `max_turns: 3`.
5. **Free nodes** — Tool nodes and human nodes cost $0. One sentence.
6. **`dippin optimize`** — Run it on the same workflow. Show it suggesting cheaper model substitutions: "Swap to claude-haiku for this review node? Saves $X/run." (Actual dollar amount determined by running `dippin optimize` during implementation.) Show the output.
7. **Closing** — "Between conditional edges and cost estimation, you can build pipelines that route intelligently and stay within budget." Tease a future post (e.g., The Doctor, or Edge Coverage).

### Terminal Output

Show real `dippin cost` and `dippin optimize` output. If the current CLI output format isn't ideal for a blog post, note what would need to change (but don't design CLI changes here).

### SEO

- Title: "Cost Estimation: Know Before You Run — Dippin Blog"
- Description: "Estimate per-run pipeline costs before spending real money on LLM calls. Use dippin cost and dippin optimize to find savings."

---

## Cross-Cutting Concerns

### Blog Index Update

Add three new cards to `site/blog/index.html` in their respective positions. Each card needs:
- Category tag (deep-dive, tutorial)
- Title, description, "Read" link
- Published date

### Homepage "From the Blog"

Update the 3 featured posts on `site/index.html` to include the newest post(s) if appropriate.

### Navigation

No nav changes needed — blog link already exists in all pages.

### Existing Post Links

The "What's Next" sections in existing posts (especially Getting Started and Editor Setup) may benefit from links to these new posts. Evaluate during implementation.

### HTML Template

Follow the established pattern from `getting-started.html`:
- Full SEO meta tags (OG, Twitter Cards, canonical URL)
- Same CSS variables and post-header/post-meta-bar structure
- Syntax-highlighted code blocks using existing `highlight.js`
- "What's Next" footer with card links to related posts
- Nav from `site/_layout/nav.html` (sync-nav handles propagation)

### blog-ideas.md Update

Mark posts #6, #12, and #17 as published with file paths.
