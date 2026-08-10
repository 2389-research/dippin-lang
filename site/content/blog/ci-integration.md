---
title: "CI Integration: Lint, Test, Format"
date: "2026-03-27"
description: "Set up GitHub Actions to validate, lint, test, and format-check your Dippin workflow files on every push and pull request."
tagStyle: "guide"
tagLabel: "GUIDE"
category: "Integration"
readTime: "11 min read"
related:
  - url: "scenario-testing.html"
    title: "Scenario Testing"
    summary: "Write the .test.json files that your CI pipeline will run on every push."
  - url: "editor-setup.html"
    title: "Editor Setup"
    summary: "Get the same diagnostics in real-time in your editor, before you even push."
---

Your Dippin workflows deserve the same CI treatment as your application code.
Here's how to set up [GitHub Actions](https://docs.github.com/en/actions)
to validate, lint, test, and format-check your `.dip` files on every push.
We'll build a workflow YAML from scratch, then look at how dippin-lang dogfoods
its own CI.

## What to Check in CI

A Dippin CI pipeline has four stages. Each catches a different class
of problem:

<div class="pipeline-visual">
  <div class="pipeline-box green">validate</div>
  <span class="pipeline-arrow">&rarr;</span>
  <div class="pipeline-box lavender">lint</div>
  <span class="pipeline-arrow">&rarr;</span>
  <div class="pipeline-box yellow">test</div>
  <span class="pipeline-arrow">&rarr;</span>
  <div class="pipeline-box green">fmt --check</div>
</div>

| Stage | Command | What it catches | Exit code on failure |
|-------|---------|----------------|---------------------|
| Validate | `dippin validate` | Structural errors ([DIP001-DIP010](../validation.html)) | 1 |
| Lint | `dippin lint` | Semantic warnings ([DIP101-DIP160](../validation.html)) | 0 (warnings don't fail) |
| Test | `dippin test` | [Scenario test](scenario-testing.html) failures | 1 |
| Format | `dippin fmt --check` | Non-canonical formatting | 1 |

<div class="callout">
  <h4>Exit code contract</h4>
  <p>
    <code>dippin validate</code> and <code>dippin test</code> exit 1 on failure,
    making them suitable as CI gates. <code>dippin lint</code> exits 0 even with
    warnings -- warnings are advisory, not blocking. To fail CI on warnings,
    use <code>dippin check</code>, which combines validate + lint and can
    be configured to fail on warnings.
  </p>
</div>

## A Complete GitHub Actions Workflow

Create `.github/workflows/dippin.yml` in your repository:

<pre>
name: Dippin CI

on:
  push:
    branches: [main]
    paths: ['**/*.dip', '**/*.test.json']
  pull_request:
    branches: [main]
    paths: ['**/*.dip', '**/*.test.json']

jobs:
  dippin:
    name: Validate, Lint, Test, Format
    runs-on: ubuntu-latest
    steps:
      - name: Check out code
        uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.25'
          cache: true

      - name: Install dippin
        run: go install github.com/2389-research/dippin-lang/cmd/dippin@latest

      - name: Validate all .dip files
        run: |
          failed=0
          for f in $(find . -name '*.dip' -not -path './.git/*'); do
            if ! dippin validate "$f"; then
              failed=1
            fi
          done
          exit $failed

      - name: Lint all .dip files
        run: |
          for f in $(find . -name '*.dip' -not -path './.git/*'); do
            dippin lint "$f"
          done

      - name: Run scenario tests
        run: |
          failed=0
          for f in $(find . -name '*.dip' -not -path './.git/*'); do
            test_file="${f%.dip}.test.json"
            if [ -f "$test_file" ]; then
              if ! dippin test "$f"; then
                failed=1
              fi
            fi
          done
          exit $failed

      - name: Check formatting
        run: |
          failed=0
          for f in $(find . -name '*.dip' -not -path './.git/*'); do
            if ! dippin fmt --check "$f"; then
              echo "::error::$f is not formatted. Run: dippin fmt --write $f"
              failed=1
            fi
          done
          exit $failed
</pre>

The `paths` filter means the job only runs when `.dip` or
`.test.json` files change, saving CI minutes.

## Understanding Each Step

### Validate: Structural Correctness

Validation catches fatal errors. If a workflow fails validation, nothing else
matters -- the file is structurally broken. Run this first.

<pre>
<span class="hl-dim">$</span> <span class="hl-shcmd">dippin validate pipelines/deploy.dip</span>
<span class="hl-fail">FAIL</span>  pipelines/deploy.dip
  <span class="hl-fail">DIP001</span>: start node "Begin" is not defined
  <span class="hl-fail">DIP004</span>: node "Cleanup" is unreachable from start
</pre>

Validation errors block CI: exit code 1. Fix the structural
issue before proceeding.

### Lint: Semantic Warnings

Linting checks for best-practice violations that aren't structurally wrong
but may indicate problems:

<pre>
<span class="hl-dim">$</span> <span class="hl-shcmd">dippin lint pipelines/deploy.dip</span>
<span class="hl-warn">WARN</span>  pipelines/deploy.dip
  <span class="hl-warn">DIP108</span>: node "Analyze": unknown model "gpt-4-turbo" for provider "openai"
  <span class="hl-warn">DIP111</span>: tool node "RunTests" has no timeout
  <span class="hl-warn">DIP103</span>: node "Review" has no prompt
</pre>

Warnings exit 0 by default. The CI step runs lint for visibility -- developers
see warnings in the job log. To block merges on warnings, swap in
`dippin check`.

### Test: Scenario Verification

Scenario tests verify that given specific context values, the workflow follows
the expected execution path. See the
[Scenario Testing](scenario-testing.html) guide for details on
writing test files.

<pre>
<span class="hl-dim">$</span> <span class="hl-shcmd">dippin test pipelines/deploy.dip</span>
<span class="hl-pass">PASS</span>  happy path -- all checks pass
<span class="hl-pass">PASS</span>  validation fails -- routes to manual review
<span class="hl-fail">FAIL</span>  timeout -- routes to fallback
  <span class="hl-fail">expected visited: [Fallback], got visited: [Retry]</span>

<span class="hl-dim">2/3 passed</span>  pipelines/deploy.dip
</pre>

Any test failure exits 1, blocking the pipeline.

### Format: Canonical Style

The `--check` flag verifies formatting without modifying the file.
Non-canonical formatting exits 1:

<pre>
<span class="hl-dim">$</span> <span class="hl-shcmd">dippin fmt --check pipelines/deploy.dip</span>
<span class="hl-fail">FAIL</span>  pipelines/deploy.dip is not formatted
</pre>

The fix is always the same: `dippin fmt --write pipelines/deploy.dip`.
The formatter is idempotent, so formatting an already-formatted file is a no-op.

## JSON Output for Job Summaries

Use `--format json` with `dippin check` to produce
machine-readable output for GitHub job summaries:

<pre>
<span class="hl-dim">$</span> <span class="hl-shcmd">dippin --format json check pipeline.dip</span>
{
  "file": "pipeline.dip",
  "valid": true,
  "errors": 0,
  "warnings": 2,
  "diagnostics": [
    {"code": "DIP111", "severity": "warning", "message": "tool node \"RunTests\" has no timeout"},
    {"code": "DIP103", "severity": "warning", "message": "node \"Review\" has no prompt"}
  ]
}
</pre>

Parse it with `jq` to build Markdown tables for
`$GITHUB_STEP_SUMMARY`:

<pre>
- name: Lint summary
  run: |
    summary="## Dippin Lint Results\n\n"
    summary+="| File | Status | Errors | Warnings |\n"
    summary+="|------|--------|--------|----------|\n"
    for f in pipelines/*.dip; do
      output=$(dippin --format json check "$f" 2>/dev/null) || true
      valid=$(echo "$output" | jq -r '.valid // false')
      errors=$(echo "$output" | jq -r '.errors // 0')
      warnings=$(echo "$output" | jq -r '.warnings // 0')
      status="pass"
      [ "$valid" != "true" ] &amp;&amp; status="fail"
      summary+="| \`$f\` | $status | $errors | $warnings |\n"
    done
    echo -e "$summary" >> "$GITHUB_STEP_SUMMARY"
</pre>

This renders a table in the GitHub Actions run summary showing the status
of every pipeline file at a glance.

## How Dippin Dogfoods Its Own CI

The [dippin-lang repository](https://github.com/2389-research/dippin-lang)
runs five CI jobs on every push and pull request. The most relevant for pipeline
authors is `validate-examples`, which runs seven checks against
all example files:

| Step | What it does |
|------|-------------|
| Parse all .dip examples | Verifies the parser accepts every file |
| Validate all .dip examples | Structural correctness |
| Lint all .dip examples | Semantic best practices |
| Format check (idempotency) | Formats twice, asserts output is identical |
| Full round-trip | Parse, format, re-parse, validate -- nothing lost |
| Validate migration parity | DOT/Dippin pairs remain structurally equivalent |
| Simulate all examples | Dry-run simulation for each workflow |

The idempotency check is worth highlighting. It formats the file, formats the
formatted output, and asserts they're identical. This catches formatter bugs where
a second pass changes the output:<sup>1</sup>

<pre>
- name: Format check all .dip examples (idempotency)
  run: |
    failed=0
    for f in examples/*.dip; do
      formatted=$(./dippin fmt "$f" 2>/dev/null)
      reformatted=$(echo "$formatted" | ./dippin fmt /dev/stdin 2>/dev/null)
      if [ "$formatted" != "$reformatted" ]; then
        echo "::error::$f is not format-idempotent"
        failed=1
      fi
    done
    exit $failed
</pre>

## The Dogfood Lint Job

A separate `dippin-lint` job uses `dippin check`
with JSON output to generate a summary table in the GitHub Actions UI --
Dippin linting its own example files:

<pre>
<span class="hl-dim">$</span> <span class="hl-shcmd">dippin --format json check examples/code_quality_sweep.dip</span>
{
  "file": "examples/code_quality_sweep.dip",
  "valid": true,
  "errors": 0,
  "warnings": 0,
  "diagnostics": []
}
</pre>

## Adding to an Existing CI Pipeline

If you already have a GitHub Actions workflow, add Dippin checks as
a new job or append steps to an existing one:

<pre>
<span class="hl-cmt"># Add to your existing workflow</span>
  dippin-check:
    name: Dippin Workflows
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.25'
      - run: go install github.com/2389-research/dippin-lang/cmd/dippin@latest
      - run: |
          for f in $(find . -name '*.dip' -not -path './.git/*'); do
            dippin check "$f"
          done
</pre>

The minimum viable CI step is a single `dippin check` loop --
it combines validate + lint in one command. Add `dippin test` and
`dippin fmt --check` when you want stricter enforcement.

## Pre-commit Hooks

For instant local feedback before pushing, add Dippin checks to a
[pre-commit hook](https://git-scm.com/book/en/v2/Customizing-Git-Git-Hooks):

<pre>
<span class="hl-cmt">#!/bin/sh</span>
<span class="hl-cmt"># .git/hooks/pre-commit</span>

<span class="hl-cmt"># Check all staged .dip files</span>
staged_dip=$(git diff --cached --name-only --diff-filter=ACM | grep '\.dip$')
if [ -n "$staged_dip" ]; then
  for f in $staged_dip; do
    if ! dippin validate "$f"; then
      echo "Validation failed for $f"
      exit 1
    fi
    if ! dippin fmt --check "$f"; then
      echo "$f is not formatted. Run: dippin fmt --write $f"
      exit 1
    fi
  done
fi
</pre>

## What's Next?

Your CI pipeline now catches structural errors, semantic issues, test failures,
and formatting drift.

<div class="footnotes">
  <h3>Notes</h3>
  <ol>
    <li id="fn1">Non-idempotent formatters are a classic source of CI frustration: format, commit, CI formats again and gets a different result, CI fails. Dippin's formatter was designed to be idempotent from day one, and this CI check is the proof. The implementation lives in <a href="https://github.com/2389-research/dippin-lang/tree/main/formatter"><code>formatter/</code></a>.</li>
  </ol>
</div>
