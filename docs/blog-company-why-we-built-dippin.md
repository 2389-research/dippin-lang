# Why We Built a Language for AI Pipelines

Last March, one of our engineers spent forty minutes debugging a
broken pipeline. The fix: a missing backslash in a DOT file — one
character, buried inside a string that looked like this:

```text
tool_command="set -eu\nmkdir -p .ai .ai/drafts .ai/sprints\nif [ ! -f
.ai/ledger.tsv ]; then\n  now=$(date -u +%Y-%m-%dT%H:%M:%SZ)\n  printf
'sprint_id\\ttitle\\tstatus\\tcreated_at\\tupdated_at\\n001\\tBootstrap
sprint\\tplanned\\t%s\\t%s\\n' \"$now\" \"$now\" > .ai/ledger.tsv\nfi\n
printf 'ledger-ready'"
```

That's a shell script. Six lines of bash — create a directory,
write a TSV header if it doesn't exist, print a status message.
Nothing exotic. But inside a DOT attribute, every newline becomes
`\n`, every tab becomes `\\t`, every quote becomes `\"`. The script
is there, but you can't read it. You can't edit it with confidence.

We build AI pipeline workflows — multi-step processes where LLM
agents, tool calls, and human reviewers collaborate on complex
tasks: code review, sprint execution, API design. These pipelines
are directed graphs — nodes with prompts and models, edges with
conditions, retry loops, parallel branches. We defined them in
Graphviz DOT.

DOT worked when our pipelines were small. Five nodes, simple edges,
short prompts. But our pipelines grew. Twenty-node workflows with
multi-model consensus. Shell scripts that run test suites. System
prompts with embedded markdown and JSON schemas. The authoring
format stopped being invisible and started being the thing we
fought with most.

We were spending more time debugging escaped strings than writing
prompts. That was the signal.

## What DOT couldn't give us

The escaped-string problem was the most visible pain, but not the
only one.

DOT is a graph description language. It knows about nodes, edges,
and attributes. It does not know what an AI pipeline is. It has no
opinion about whether `claude-sonnet-4-6` is a valid model name or
`claude-sonet-4-6` is a typo. It won't tell you that a node is
unreachable, that a retry loop has no exit condition, or that a tool
command references a missing binary. You find these things out in
production, when the pipeline fails — or worse, produces subtly
wrong output.

Testing pipelines was its own problem. LLM calls are
non-deterministic; you can't assert on their output. But you can
assert on the *shape* of execution: which nodes were visited, in
what order, which branches were taken. We needed that. DOT had no
concept of it.

Then there was cost. A pipeline that fans out to three LLM providers
runs three sets of API calls. Before we could estimate the total,
someone had to manually count prompt tokens and look up pricing
tables. For twenty pipelines, that doesn't scale.

## A language, not a format

We could have written a YAML schema with a validation layer on top.
But validation is the floor, not the ceiling. We wanted a formatter
that normalizes style, a simulator that walks execution paths, a
cost estimator that reads prompt tokens, an LSP that shows
diagnostics in your editor. All of that requires a grammar — a
parser that produces a typed data model the entire toolchain can
share.

So we built Dippin.

That same shell script, in Dippin:

<!-- markdownlint-disable MD013 -->

```dip
tool EnsureLedger
  label: "Ensure Ledger"
  command:
    set -eu
    mkdir -p .ai .ai/drafts .ai/sprints
    if [ ! -f .ai/ledger.tsv ]; then
      now=$(date -u +%Y-%m-%dT%H:%M:%SZ)
      printf 'sprint_id\ttitle\tstatus\tcreated_at\tupdated_at\n001\tBootstrap sprint\tplanned\t%s\t%s\n' "$now" "$now" > .ai/ledger.tsv
    fi
    printf 'ledger-ready'
```

<!-- markdownlint-enable MD013 -->

Indent after the colon and write your script. No escaping, no
quoting, no `\n`. The same rule applies to prompts: multi-line
markdown with headers, bullet points, embedded code blocks, JSON
examples. Write it the way you'd write it in a document.

Here's a complete pipeline — a document gets drafted, reviewed, and
either published or sent back for revision:

```dip
workflow ReviewPipeline
  goal: "Draft, review, and publish a document"
  start: Start
  exit: Exit

  defaults
    provider: anthropic
    model: claude-sonnet-4-6

  agent Draft
    label: "Write Draft"
    prompt:
      Write a clear, concise technical document based on the
      provided requirements. Focus on accuracy and readability.

  agent Review
    label: "Review Draft"
    auto_status: true
    prompt:
      Review the draft for accuracy, clarity, and completeness.
      Return success if it meets standards, or fail with feedback.

  agent Publish
    label: Publish

  edges
    Start -> Draft
    Draft -> Review
    Review -> Publish  when ctx.outcome = success
    Review -> Draft    when ctx.outcome = fail
    Publish -> Exit
```

The conditional edges say what they mean: if the review passes,
publish; if it fails, go back to drafting.

## Tooling follows language

A config format stores data. A language has structure you can query,
check, and transform. That difference is what makes everything
below possible.

Dippin ships with 39 diagnostic checks. Nine catch structural
errors: your file references a node that doesn't exist, or declares
a start node with no outgoing edges. Thirty catch semantic
problems: an unknown model name, a tool command with no timeout, a
condition that references a variable without its namespace prefix.
Every diagnostic has a code, an explanation, and a fix suggestion.
Run `dippin explain DIP108` and it tells you what went wrong and
how to fix it.

Dippin's scenario testing lets you inject context values and assert
on execution paths — which nodes were visited, which weren't. The
tests are deterministic even though the underlying LLM calls are
not. Our CI runs `dippin check` on every push; a broken pipeline
fails the build before it reaches production.

Cost estimation came next. `dippin cost` counts prompt tokens,
applies per-model pricing, and accounts for retry loops:

```text
$ dippin cost complexity_cleanup.dip
═══ Cost Estimate ═════════════════════════════════════════
                                Min Expected      Max
  ──────────────────────── ──────── ──────── ────────
  TOTAL                       $0.65    $0.65    $2.66
```

`dippin optimize` then suggests where cheaper models would do the
same job. Our code review pipeline dropped from $0.65 to $0.02
expected cost after following its suggestions.

The toolchain kept growing. An LSP server catches errors as you
type. A semantic diff tool reports "the model changed from opus to
sonnet on this node" instead of a raw text diff. A migration tool
converts existing DOT files with structural parity verification.
There's a WASM playground, a file watcher, syntax highlighting —
the kind of tooling you can only build when you have a grammar to
build on.

## What this means in practice

Pipeline authors think about logic now, not string escaping. A new
team member reads a `.dip` file and understands the workflow without
a walkthrough. When someone pushes a change, CI validates structure,
checks semantics, and estimates the cost delta. The change either
passes or it doesn't — no ambiguity.

The feedback loop between the runtime and the language is tight. Last
week, the runtime needed to force LLM APIs to return structured JSON —
a feature that all three providers support, but that requires
specific API parameters to activate. In DOT, adding this would have
meant inventing an attribute convention, documenting it somewhere,
and hoping people used it correctly. In Dippin, we added
`response_format` and `response_schema` as first-class fields with
four lint rules to catch mistakes. The runtime adapter picked them
up automatically. Request to production in a day.

Because everything reads the same typed model, adding
`response_format` meant the linter, formatter, cost estimator, and
LSP all understood it immediately. One grammar change, and every
tool caught up. That's what a language gives you that a config
format can't.

## Try it

Dippin is [open source](https://github.com/2389-research/dippin-lang).
We built it for ourselves, but the problem it solves isn't unique
to us. If you're building multi-step LLM pipelines — with
conditional routing, human checkpoints, tool calls, retry logic —
and you're defining them in YAML or JSON or DOT or raw config
files, you know the friction.

Install it, point it at a workflow, and see what it finds.

```sh
go install github.com/2389-research/dippin-lang/cmd/dippin@latest
```

It's on [GitHub](https://github.com/2389-research/dippin-lang).
MIT license.
