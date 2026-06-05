# Subgraph Composition

## The problem

Pipelines repeat themselves. A three-step interview loop — generate
questions, collect answers, assess completeness — shows up in API
design workflows, onboarding flows, and requirements gathering.
Without composition, you copy the same nodes and edges into every
workflow that needs them. When the pattern changes, you update it
in five places and miss the sixth.

## How subgraphs work

A subgraph node embeds one workflow inside another. It looks like
any other node in the graph — it has a label, it connects to other
nodes with edges, it participates in retry logic and conditional
routing. But instead of running an LLM call or a shell command, it
references a separate `.dip` file:

```dip
subgraph Interview
  label: "Requirements interview"
  ref: interview_loop.dip
  writes: requirements_summary
  params:
    topic: "API design"
    focus: "resources, auth, consumers, scale"
```

`ref` points to the workflow file. `params` passes key-value pairs
into it. Inside `interview_loop.dip`, those values are available as
`${params.topic}` and `${params.focus}` — the same interpolation
syntax used for context variables, but in a dedicated namespace that
keeps parent and child workflows from stepping on each other.

The referenced workflow is a complete, self-contained `.dip` file.
It has its own start node, exit node, edges, and node definitions.
You can validate, lint, format, and cost it independently. It
doesn't know or care that it's being embedded — it's just a
workflow.

## What this means for the toolchain

The subgraph node is opaque to the parent workflow's toolchain
passes. When `dippin lint` runs on the parent, it checks that the
referenced file exists on disk (DIP126) and warns if two subgraph
nodes reference the same file (DIP109, a namespace collision risk).
It does not inline or expand the subgraph. The child workflow gets
its own lint pass when you run `dippin lint` on it directly.

`dippin lint` also emits DIP143 (Hint) when this workflow declares
`tool_access` on any agent or parallel branch and also references an
external subgraph. `tool_access` is per-node and file-bounded — it
does not propagate across a `ref:` or `subgraph_ref:` boundary; the
child workflow's agents are governed solely by their own file. DIP143
fires only when the parent shows tool-containment intent (a non-empty
`tool_access` somewhere in this file) and the reference points at a
different file — a node that references its own source file is not
flagged. The lint never reads the child file. Cross-file
effective-access enforcement is deferred to issue #89.

The simulator treats subgraphs as atomic steps. It records that the
node was entered and exited, logs the ref path, and moves on to the
next edge. It doesn't recurse into the child workflow. This is
deliberate: the simulator provides a control-flow trace of the
parent pipeline, not a fully expanded execution tree. Runtime
expansion is the orchestrator's job.

The formatter emits subgraph fields in a fixed order — label, ref,
params — with param keys sorted alphabetically. This makes diffs
clean and round-trips deterministic.

Cost estimation sums the parent workflow's nodes. The child
workflow's cost is estimated separately when you run `dippin cost`
on it. This keeps cost reports scoped to one file at a time, which
matches how teams reason about budgets: "what does this workflow
cost?" not "what does this workflow plus everything it calls cost?"

## manager_loop and subgraph_ref

`manager_loop` is a supervised variant of the subgraph pattern. Where
a plain `subgraph` node embeds a child and moves on, a `manager_loop`
node runs the child repeatedly, polls on an interval, and can steer
the child by injecting context between cycles. It uses `subgraph_ref:`
instead of `ref:`:

```dip
manager_loop Supervise
  subgraph_ref: quality_loop.dip
  poll_interval: 30s
  stop_condition: stack.child.outcome = success
```

`stop_condition` and `steer_condition` are evaluated over the child's
runtime state in the `stack.child.*` namespace (`stack.child.outcome`,
`stack.child.cycles`, `stack.child.status`) — see [nodes.md](nodes.md)
for the full supervisor variable contract.

The same file-boundary rules apply. `quality_loop.dip` is a separate,
independently lintable `.dip` file — the parent treats it as an opaque
step. `manager_loop` nodes are subject to DIP143 for the same reason
plain `subgraph` nodes are: `tool_access` declared in the parent does
not reach into the child file. The lint uses `subgraph_ref:` as the
reference path for the same boundary check.

## Why not inline?

An earlier design inlined subgraphs at parse time — the parser
would read the referenced file, prefix node IDs to avoid collisions,
and splice the nodes and edges into the parent graph. This was
simpler conceptually but caused problems:

- Lint diagnostics pointed to line numbers in an expanded graph
  that didn't correspond to any file the author could edit.
- Cost estimates doubled when the same subgraph appeared twice.
- The formatter couldn't round-trip an inlined graph back to the
  original two-file structure.
- Error messages were confusing: "node Interview_Assess has no
  fallback" means nothing when the author named it "Assess" in
  `interview_loop.dip`.

Keeping subgraphs opaque at the IR level avoids all of this. Each
file is a self-contained unit. Tooling operates on one file at a
time. The runtime handles expansion.

## The runtime contract

Dippin defines the subgraph — the ref, the params, the edges into
and out of it. The runtime is responsible for loading the referenced
file, substituting params, and executing the child workflow as part
of the parent pipeline. The IR gives the runtime everything it needs:

```go
cfg := node.Config.(ir.SubgraphConfig)
cfg.Ref    // "interview_loop.dip"
cfg.Params // {"topic": "API design", "focus": "resources, auth, ..."}
```

The adapter reads these fields and hands them to the pipeline
engine. No special protocol, no registration step. If the file
exists and parses, it runs.

One thing the IR does not carry across the boundary is `tool_access`.
Restrictions on nodes in this file do not extend into the referenced
child file; the child's agents must declare their own `tool_access`.
DIP143 reminds you when tool-containment intent is present in the
parent but the child is opaque to this lint pass.

## A real example

`api_design.dip` is a 20-node pipeline that produces an API design
package — OpenAPI spec, SDK examples, error catalog. One of its
steps is a requirements interview. Rather than embedding the
interview logic (generate questions, collect answers, assess, loop
if incomplete), it references `interview_loop.dip`:

```dip
subgraph Interview
  label: "Requirements interview"
  ref: interview_loop.dip
  writes: requirements_summary
  params:
    topic: "API design"
    focus: "resources, auth, consumers, scale, integrations, real-time needs"
```

`interview_loop.dip` is parameterized by topic and focus areas. The
same file could be referenced by a user research workflow, an
onboarding pipeline, or a support triage flow — each passing
different params. The interview logic lives in one place.

When the interview pattern changes — say we add a confidence score
to the assessment step — we update `interview_loop.dip` once. Every
workflow that references it picks up the change on its next run.

## What subgraphs don't do (yet)

Subgraphs are file-based references with flat string params. There
is no module registry, no version pinning, no type-checked parameter
contracts. DIP109 warns about namespace collisions but doesn't
prevent them. Recursive subgraphs (a subgraph that references
itself) are not detected or prohibited — the runtime would loop.

These are real limitations. They're also the right trade-offs for
where the project is today. The file-based approach works with
standard tooling — editors, git, CI — without inventing a package
system. When the limitations bite, we'll address them. So far they
haven't.

## See also: packaging via `.dipx`

Subgraph refs are resolved against the filesystem at runtime, which
means the runtime needs every referenced file present at the right
relative path. `dippin pack` collects an entry `.dip` plus every
transitively-reachable ref into a single deterministic `.dipx`
bundle so the whole tree ships as one integrity-verified artifact.
Inside the bundle, refs are resolved against the bundle root with
the same lexical rules — no symlinks, no escapes from the workflow
tree, NFC-canonical paths. See `docs/cli.md` (pack/unpack/inspect)
and the `dipx` package documentation for details.
