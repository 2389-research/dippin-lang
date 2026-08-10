# `dippin spec` Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a `dippin spec` command that dumps a comprehensive language reference to stdout, assembled at build time from existing docs and embedded via `go:embed`.

**Architecture:** A shell script (`scripts/gen-spec.sh`) concatenates sections from `docs/llm-reference.md` and `site/static/skill.md` into `docs/generated-spec.md`. A new Go file embeds that file and exposes it via `dippin spec`. The same file is copied to `site/static/llms-full.txt` during site builds for web access at `dippin.org/llms-full.txt`.

**Tech Stack:** Go (`go:embed`), shell (`sh`), Hugo site, Netlify

---

## File Map

| File | Action | Responsibility |
|------|--------|----------------|
| `scripts/gen-spec.sh` | Create | Assembles `docs/generated-spec.md` from source docs |
| `docs/generated-spec.md` | Generated | The assembled spec (gitignored) |
| `cmd/dippin/cmd_spec.go` | Create | `dippin spec` command — embeds and prints the spec |
| `cmd/dippin/cmd_spec_test.go` | Create | Tests for the spec command |
| `cmd/dippin/cli.go` | Modify | Add `"spec"` to dispatch map and help text |
| `Justfile` | Modify | Add `gen-spec` recipe, update `build` and `site-build` deps |
| `netlify.toml` | Modify | Add `llms-full.txt` header, add gen-spec to build commands |
| `.gitignore` | Modify | Add `docs/generated-spec.md` |

---

### Task 1: Create the spec generator script

**Files:**
- Create: `scripts/gen-spec.sh`

This script concatenates sections from existing maintained docs into a single flat markdown file. It extracts content by heading markers using `sed`. The output must NOT include the skill.md frontmatter (the `---`/`name:`/`description:`/`---` block) or documentation links section at the bottom.

- [ ] **Step 1: Create `scripts/gen-spec.sh`**

```sh
#!/bin/sh
# ABOUTME: Assemble docs/generated-spec.md from existing doc sources.
# ABOUTME: Concatenates llm-reference.md and skill.md sections into one agent-facing spec.
set -e

OUTPUT="docs/generated-spec.md"
LLM_REF="docs/llm-reference.md"
SKILL="site/static/skill.md"

# Verify sources exist
for f in "$LLM_REF" "$SKILL"; do
  if [ ! -f "$f" ]; then
    echo "gen-spec: $f not found" >&2
    exit 1
  fi
done

cat > "$OUTPUT" << 'HEADER'
# Dippin Language Specification

Complete reference for AI agents generating `.dip` workflow files. This document is the canonical, self-contained spec for the dippin DSL and CLI toolchain.

Install: `go install github.com/2389-research/dippin-lang/cmd/dippin@latest`

Generate from template: `dippin new minimal`, `dippin new parallel`, `dippin new conditional`, `dippin new review-loop`, `dippin new human-gate`

HEADER

# --- Section 1: Grammar from llm-reference.md ---
# Extract everything from "## Grammar" through end of file, skipping the title line and first ---
sed -n '/^## Grammar/,$ p' "$LLM_REF" >> "$OUTPUT"

# --- Section 2: Syntax details from skill.md ---
# Extract "## File Structure" through "## Node Types" (exclusive), then
# "## Node Types" through "## Common Fields" (inclusive), etc.
# We grab the bulk of skill.md from "## File Structure" to "## Documentation" (exclusive),
# skipping the frontmatter and Quick Start section which overlap with the header above.

{
  echo ""
  echo "---"
  echo ""
  # Extract from "## File Structure" up to but not including "## Documentation"
  sed -n '/^## File Structure/,/^## Documentation/ { /^## Documentation/d; p; }' "$SKILL"

  echo ""
  echo "---"
  echo ""
  echo "## Documentation"
  echo ""
  echo "- [Language Reference](https://dippin.org/language/)"
  echo "- [CLI Reference](https://dippin.org/cli/)"
  echo "- [Validation & Linting](https://dippin.org/validation/)"
  echo "- [Scenario Testing](https://dippin.org/testing/)"
  echo "- [Analysis Tools](https://dippin.org/analysis/)"
  echo "- [Playground](https://dippin.org/playground/)"
  echo "- [GitHub](https://github.com/2389-research/dippin-lang)"
} >> "$OUTPUT"

echo "gen-spec: wrote $OUTPUT"
```

- [ ] **Step 2: Make the script executable**

Run: `chmod +x scripts/gen-spec.sh`

- [ ] **Step 3: Run the script and verify output**

Run: `./scripts/gen-spec.sh`
Expected: prints `gen-spec: wrote docs/generated-spec.md`

Then verify the output contains expected sections:

Run: `head -5 docs/generated-spec.md`
Expected: starts with `# Dippin Language Specification`

Run: `grep -c '^## ' docs/generated-spec.md`
Expected: 15+ section headings

Run: `grep 'DIP001' docs/generated-spec.md | head -1`
Expected: a line containing the DIP001 diagnostic

Run: `grep '^---$' docs/generated-spec.md | head -1`
Expected: at least one horizontal rule (section separator from llm-reference.md)

Verify NO frontmatter leaked in:

Run: `grep '^name: dippin-lang' docs/generated-spec.md`
Expected: no output (exit code 1)

- [ ] **Step 4: Commit**

```bash
git add scripts/gen-spec.sh
git commit -m "feat: add spec generator script

Assembles docs/generated-spec.md from llm-reference.md and skill.md
for embedding into the dippin binary."
```

---

### Task 2: Add `.gitignore` entry for generated spec

**Files:**
- Modify: `.gitignore`

The generated spec is a build artifact and should not be tracked.

- [ ] **Step 1: Add the gitignore entry**

Add this line after the `# Temp site files` section in `.gitignore`:

```text
# Generated spec (build artifact)
docs/generated-spec.md
```

- [ ] **Step 2: Verify it's ignored**

Run: `git status docs/generated-spec.md`
Expected: no output (file is ignored)

- [ ] **Step 3: Commit**

```bash
git add .gitignore
git commit -m "chore: gitignore generated spec file"
```

---

### Task 3: Create the `dippin spec` command

**Files:**
- Create: `cmd/dippin/cmd_spec.go`
- Modify: `cmd/dippin/cli.go`

The command embeds the generated spec via `go:embed` and prints it to stdout. It follows the same pattern as other simple commands (like `CmdExplain`).

Important: `go:embed` paths are relative to the Go source file. `cmd/dippin/cmd_spec.go` is two directories deep from the repo root, so the embed path is `../../docs/generated-spec.md`.

- [ ] **Step 1: Ensure the generated spec exists**

Run: `./scripts/gen-spec.sh`
Expected: `gen-spec: wrote docs/generated-spec.md`

(The embed directive requires the file to exist at compile time.)

- [ ] **Step 2: Create `cmd/dippin/cmd_spec.go`**

```go
// ABOUTME: Implements the `dippin spec` command.
// ABOUTME: Embeds the build-time-generated language spec and prints it to stdout.
package main

import (
	_ "embed"
	"fmt"
)

//go:embed ../../docs/generated-spec.md
var specContent string

// CmdSpec prints the full dippin language specification to stdout.
func (c *CLI) CmdSpec(args []string) ExitCode {
	fmt.Fprint(c.Stdout, specContent)
	return ExitOK
}
```

- [ ] **Step 3: Register in dispatch map**

In `cmd/dippin/cli.go`, add `"spec"` to the `commandDispatch()` map. Add it after the `"lsp"` entry:

```go
		"lsp":                c.CmdLSP,
		"spec":               c.CmdSpec,
```

- [ ] **Step 4: Add to help text**

In `cmd/dippin/cli.go`, in `printGlobalUsage()`, add the spec command after the `lsp` line and before the `version` line:

```go
	fmt.Fprintln(w, "  spec                              Print full language specification")
```

- [ ] **Step 5: Verify it compiles and runs**

Run: `go build -ldflags "-X main.version=dev -X main.commit=test -X main.date=now" -o dippin ./cmd/dippin/`
Expected: builds successfully

Run: `./dippin spec | head -3`
Expected:
```text
# Dippin Language Specification

Complete reference for AI agents generating `.dip` workflow files. This document is the canonical, self-contained spec for the dippin DSL and CLI toolchain.
```

Run: `./dippin help | grep spec`
Expected: `  spec                              Print full language specification`

- [ ] **Step 6: Commit**

```bash
git add cmd/dippin/cmd_spec.go cmd/dippin/cli.go
git commit -m "feat: add dippin spec command

Embeds the generated language spec via go:embed and prints it to stdout.
Registered in dispatch map and help text."
```

---

### Task 4: Write tests for `dippin spec`

**Files:**
- Create: `cmd/dippin/cmd_spec_test.go`

Tests use the existing `runCLI` helper from `main_test.go`. The spec content is embedded at compile time, so the tests verify the embedded content has the expected structure.

- [ ] **Step 1: Create `cmd/dippin/cmd_spec_test.go`**

```go
// ABOUTME: Tests for the dippin spec command.
// ABOUTME: Verifies the embedded spec contains expected sections and markers.
package main

import (
	"strings"
	"testing"
)

func TestCmdSpec_OutputContainsExpectedSections(t *testing.T) {
	stdout, stderr, code := runCLI(t, "spec")

	if code != ExitOK {
		t.Fatalf("expected exit 0, got %d; stderr: %s", code, stderr)
	}
	if stderr != "" {
		t.Errorf("expected no stderr, got: %s", stderr)
	}

	// Verify the spec is non-trivial (at least 100 lines).
	lines := strings.Count(stdout, "\n")
	if lines < 100 {
		t.Errorf("expected at least 100 lines, got %d", lines)
	}

	// Verify key sections are present.
	markers := []string{
		"# Dippin Language Specification",
		"## Grammar",
		"## Node Kinds",
		"## Edge Conditions",
		"## File Structure",
		"## Node Types",
		"## Diagnostic Codes",
		"## CLI Reference",
		"DIP001",
		"DIP101",
		"DIP133",
	}
	for _, m := range markers {
		if !strings.Contains(stdout, m) {
			t.Errorf("spec output missing expected marker: %q", m)
		}
	}
}

func TestCmdSpec_NoFrontmatter(t *testing.T) {
	stdout, _, code := runCLI(t, "spec")

	if code != ExitOK {
		t.Fatalf("expected exit 0, got %d", code)
	}

	// Verify no skill.md frontmatter leaked through.
	if strings.Contains(stdout, "name: dippin-lang") {
		t.Error("spec output contains skill.md frontmatter (name: dippin-lang)")
	}
	if strings.Contains(stdout, "description: Use when working") {
		t.Error("spec output contains skill.md frontmatter (description)")
	}
}
```

- [ ] **Step 2: Run the tests**

Run: `go test ./cmd/dippin/ -run TestCmdSpec -v -count=1`
Expected: both tests PASS

- [ ] **Step 3: Commit**

```bash
git add cmd/dippin/cmd_spec_test.go
git commit -m "test: add tests for dippin spec command

Verifies embedded spec contains expected sections, markers,
and no leaked frontmatter."
```

---

### Task 5: Update Justfile with `gen-spec` recipe and build dependencies

**Files:**
- Modify: `Justfile`

The `gen-spec` recipe runs the generator script. The `build` recipe depends on it so the embedded file is always fresh before compilation. The `site-build` recipe copies the generated spec to `site/static/llms-full.txt`.

- [ ] **Step 1: Add `gen-spec` recipe**

Add after the `install` recipe and before the `test` recipe in the Justfile:

```just
# Generate the language spec from source docs
gen-spec:
    ./scripts/gen-spec.sh
```

- [ ] **Step 2: Add `gen-spec` as dependency of `build`**

Change the `build` recipe from:

```just
build:
    go build -ldflags ...
```

to:

```just
build: gen-spec
    go build -ldflags "-X main.version=dev -X main.commit=$(git rev-parse --short HEAD) -X main.date=$(date -u +%Y-%m-%dT%H:%M:%SZ)" -o dippin ./cmd/dippin/
```

- [ ] **Step 3: Update `site-build` recipe**

Change the `site-build` recipe from:

```just
site-build: build wasm changelog-md
    cd site && hugo --minify
```

to:

```just
site-build: build wasm changelog-md
    cp docs/generated-spec.md site/static/llms-full.txt
    cd site && hugo --minify
```

(`build` already depends on `gen-spec`, so no need to list `gen-spec` again here.)

- [ ] **Step 4: Verify the recipes work**

Run: `just gen-spec`
Expected: `gen-spec: wrote docs/generated-spec.md`

Run: `just build`
Expected: builds successfully (gen-spec runs first, then go build)

Run: `just site-build`
Expected: builds site, `site/static/llms-full.txt` exists

Run: `head -3 site/static/llms-full.txt`
Expected: `# Dippin Language Specification`

- [ ] **Step 5: Commit**

```bash
git add Justfile
git commit -m "build: add gen-spec recipe, wire into build and site-build

gen-spec assembles the language spec before compilation so go:embed
picks up the latest content. site-build copies it to llms-full.txt."
```

---

### Task 6: Update Netlify config for `llms-full.txt`

**Files:**
- Modify: `netlify.toml`

Add the Content-Type header for `llms-full.txt` and add `gen-spec` to the Netlify build commands so the spec is generated before Hugo runs.

- [ ] **Step 1: Add `llms-full.txt` header block**

Add this block after the existing `/skill.md` header block in `netlify.toml`:

```toml
[[headers]]
  for = "/llms-full.txt"
  [headers.values]
    Content-Type = "text/plain; charset=utf-8"
```

- [ ] **Step 2: Add gen-spec to Netlify build commands**

In the `[build]` command, add `./scripts/gen-spec.sh &&` before the `cd site` step and add a `cp` step. Update the command to:

```toml
[build]
  base = "/"
  command = """
    mkdir -p site/public && \
    GOOS=js GOARCH=wasm go build -o site/static/dippin.wasm ./cmd/wasm/ && \
    cp "$(go env GOROOT)/lib/wasm/wasm_exec.js" site/static/wasm_exec.js && \
    ./scripts/gen-changelog.sh && \
    ./scripts/gen-spec.sh && \
    cp docs/generated-spec.md site/static/llms-full.txt && \
    cd site && hugo --minify
  """
  publish = "site/public"
```

Apply the same change to `[context.production]` and `[context.deploy-preview]` command blocks — add `./scripts/gen-spec.sh && \` and `cp docs/generated-spec.md site/static/llms-full.txt && \` before the `cd site` lines.

For `[context.production]`:

```toml
[context.production]
  command = """
    mkdir -p site/public && \
    GOOS=js GOARCH=wasm go build -o site/static/dippin.wasm ./cmd/wasm/ && \
    cp "$(go env GOROOT)/lib/wasm/wasm_exec.js" site/static/wasm_exec.js && \
    ./scripts/gen-changelog.sh && \
    ./scripts/gen-spec.sh && \
    cp docs/generated-spec.md site/static/llms-full.txt && \
    cd site && hugo --minify --baseURL $URL
  """
```

For `[context.deploy-preview]`:

```toml
[context.deploy-preview]
  command = """
    mkdir -p site/public && \
    GOOS=js GOARCH=wasm go build -o site/static/dippin.wasm ./cmd/wasm/ && \
    cp "$(go env GOROOT)/lib/wasm/wasm_exec.js" site/static/wasm_exec.js && \
    ./scripts/gen-changelog.sh && \
    ./scripts/gen-spec.sh && \
    cp docs/generated-spec.md site/static/llms-full.txt && \
    cd site && hugo --minify --baseURL $DEPLOY_PRIME_URL
  """
```

- [ ] **Step 3: Commit**

```bash
git add netlify.toml
git commit -m "build: add llms-full.txt to Netlify build and headers

Generates spec and copies to site/static during Netlify builds.
Serves llms-full.txt with text/plain content type."
```

---

### Task 7: Run full check suite and verify end-to-end

**Files:** None (verification only)

- [ ] **Step 1: Run full check suite**

Run: `just check`
Expected: `All checks passed.`

This runs build (which runs gen-spec), vet, fmt-check, lint-go, test-race, complexity, and validate-examples.

- [ ] **Step 2: Verify `dippin spec` end-to-end**

Run: `./dippin spec | wc -l`
Expected: 500+ lines

Run: `./dippin spec | grep -c '^## '`
Expected: 15+ section headings

Run: `./dippin spec | head -1`
Expected: `# Dippin Language Specification`

Run: `./dippin spec | tail -1`
Expected: a documentation link or blank line (not frontmatter)

- [ ] **Step 3: Verify `llms-full.txt` for site build**

Run: `just site-build`
Expected: builds without errors

Run: `diff docs/generated-spec.md site/static/llms-full.txt`
Expected: no differences (files are identical)

- [ ] **Step 4: Final commit (if any fixes were needed)**

Only commit if earlier tasks required adjustments discovered during verification. Otherwise skip this step.
