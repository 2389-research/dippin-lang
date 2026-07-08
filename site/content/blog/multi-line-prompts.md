---
title: "Multi-line Prompts Without Escaping"
date: "2026-03-31"
description: "DOT's escaped strings are unreadable. Dippin's indentation-based blocks let you write prompts with markdown, JSON, and code blocks — no escaping required."
tagStyle: "deep-dive"
tagLabel: "DEEP DIVE"
category: "Language"
readTime: "6 min read"
related:
  - url: "conditional-edges.html"
    title: "Conditional Edges"
    summary: "Route pipelines based on LLM output with the when keyword."
  - url: "cost-estimation.html"
    title: "Cost Estimation"
    summary: "Know what your pipeline will cost before you run it."
---

If you've been authoring AI pipelines in Graphviz DOT, you know the pain.
Every multi-step workflow lives in a `.dot` file, and every string
-- prompts, commands, instructions -- has to fit on a single quoted line.
What that looks like in practice is this, a real `tool_command`
from a sprint management pipeline:

<pre>
tool_command="set -eu\nmkdir -p .ai .ai/drafts .ai/sprints\nif [ ! -f .ai/ledger.tsv ]; then\n  now=$(date -u +%Y-%m-%dT%H:%M:%SZ)\n  printf 'sprint_id\\ttitle\\tstatus\\tcreated_at\\tupdated_at\\n001\\tBootstrap sprint\\tplanned\\t%s\\t%s\\n' \"$now\" \"$now\" > .ai/ledger.tsv\nfi\nprintf 'ledger-ready'"
</pre>

If you've authored AI pipelines in Graphviz DOT, this looks familiar. And painful.
Every newline is `\n`, every quote is `\"`, every tab is
`\\t`. Reading this is archaeology. Editing it is anxiety.

This is the problem Dippin was built to solve. Not the only problem -- but the
original one, the one that made DOT untenable for real pipelines with real prompts.

## The Fix: Indentation-Based Blocks

Here's that same shell script as a Dippin `command:` block:

<pre>
<span class="hl-kw">tool</span> <span class="hl-node">EnsureLedger</span>
  <span class="hl-field">label</span>: <span class="hl-str">"Ensure Ledger"</span>
  <span class="hl-field">command</span>:
    set -eu
    mkdir -p .ai .ai/drafts .ai/sprints
    if [ ! -f .ai/ledger.tsv ]; then
      now=$(date -u +%Y-%m-%dT%H:%M:%SZ)
      printf 'sprint_id\ttitle\tstatus\tcreated_at\tupdated_at\n001\tBootstrap sprint\tplanned\t%s\t%s\n' "$now" "$now" > .ai/ledger.tsv
    fi
    printf 'ledger-ready'
</pre>

Indent after the colon. Write anything. That's it. No escaping, no backslash
sequences, no mental translation between what you're reading and what the
shell will actually execute. The content is exactly what it looks like.

The same rule applies to `prompt:` blocks, `goal:` fields,
and anywhere else you need multi-line text. If the first line after the colon is
blank (or the colon is followed by a newline), Dippin reads everything at the
next indentation level as the block content until indentation decreases.

## Prompts Get Even Better

Shell scripts are one thing. Prompts are where escaping really breaks down.
A useful system prompt has markdown headers, bullet lists, code examples,
and carefully structured instructions. In DOT, that's hundreds of
`\n`s and `\"`s. In Dippin:

<pre>
<span class="hl-kw">agent</span> <span class="hl-node">ReviewCode</span>
  <span class="hl-field">label</span>: <span class="hl-str">"Code Review"</span>
  <span class="hl-field">model</span>: claude-sonnet-4-6
  <span class="hl-field">provider</span>: anthropic
  <span class="hl-field">prompt</span>:
    Review the submitted code changes for:

    ## Quality Checks
    - Correctness: Does it do what it claims?
    - Edge cases: What happens with empty input?
    - Security: Any injection risks?

    ## Style
    Follow the project's existing patterns.
    Don't suggest changes that are purely aesthetic.

    Return "success" if the code passes all checks,
    or "fail" with specific feedback.
</pre>

Everything is preserved exactly: blank lines, markdown headers, bullet points,
quoted strings inside the prompt text. The `#` character is a comment
marker at the top level of a `.dip` file, but inside an indented block
it's just a literal character -- so markdown headers work without any special treatment.

The same applies to JSON snippets, YAML fragments, SQL, or any other structured
content you need to embed in a prompt. The formatter round-trips this exactly --
`dippin fmt` preserves your block content character-for-character.
What you write is what gets sent to the model.

## But It's Not Just Prompts

The block syntax solves the immediate escaping problem, but Dippin goes further
with two features that matter for real pipelines.

**Subgraph composition.** Instead of copy-pasting patterns across
workflows, you can define a reusable subgraph and reference it by file. Here's
how `api_design.dip` embeds an entire requirements-gathering interview
as a single node:

<pre>
<span class="hl-kw">subgraph</span> <span class="hl-node">Interview</span>
  <span class="hl-field">label</span>: <span class="hl-str">"Requirements interview"</span>
  <span class="hl-field">ref</span>: interview_loop.dip
  <span class="hl-field">writes</span>: requirements_summary
  <span class="hl-field">params</span>:
    topic: <span class="hl-str">"API design"</span>
    focus: <span class="hl-str">"resources, auth, consumers, scale, integrations, real-time needs"</span>
</pre>

The `interview_loop.dip` file is a self-contained Q&A subgraph --
it asks clarifying questions, collects answers, and writes a summary to the
context variable named in `writes`. `api_design.dip`
just references it and passes parameters. Swap in a different interview loop,
change the topic, reuse it in ten other workflows. No copy-paste, no drift.

**The playground.** Paste any example from this post and try it yourself
without installing anything -- [open the playground](../playground.html).
It runs the full parser and validator in your browser, so you get real diagnostics
and can experiment with the syntax immediately.

## What's Next: Install and Keep Going

Ready to use Dippin on your own pipelines? Install the binary:

<pre>
<span class="hl-dim">$</span> <span class="hl-shcmd">go install github.com/2389-research/dippin-lang/cmd/dippin@latest</span>
</pre>

Or via Homebrew (macOS and Linux):

<pre>
<span class="hl-dim">$</span> <span class="hl-shcmd">brew install 2389-research/tap/dippin-lang</span>
</pre>

If you're new to Dippin, [Getting Started](getting-started.html) walks
through writing your first workflow, validating it, linting it, and visualizing the
topology -- all in under 10 minutes.

In the next posts, we'll build branching pipelines with conditional edges and learn
how to estimate costs before running them. Multi-line prompts are the foundation;
routing and cost awareness are what make pipelines production-ready.
