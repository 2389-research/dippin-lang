# Issue #52 — `command_file:` Implementation Plan (v0.33.0)

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Ship `command_file:` on tool nodes so authors can replace inline `command:` heredocs with external file references. v0.33.0 is dippin-internal — no tracker coordination required.

**Architecture:** Parser stays pure. Stores the path verbatim in a new `ToolConfig.CommandFile` field. A separate `parser.ResolveFileDirectives(w, baseDir)` pass loads files at CLI entry points; LSP and WASM skip the resolver and see the unresolved IR view without error. Path security (absolute / parent-tree / symlink / 4MB) lives entirely in the resolver. Mutual exclusion with inline `command:` is a parser-time error (no DIP code). Formatter preserves directive form via `cfg.CommandFile != ""` branch.

**Tech Stack:** Go, dippin's recursive-descent parser, validator's existing diagnostic-emission pattern.

**Spec:** `docs/superpowers/specs/2026-05-27-issue-52-command-file-design.md`

**Build/test commands:** Always use `just` recipes (see `CLAUDE.md`). Never raw `go test`.

---

## Task 0: File 6 follow-up issues + spec backlinks

**Purpose:** Mirror the #41 task-#0 pattern that breaks the DIP28 "file follow-up if needed" anti-pattern. Numbered issues are referenced in the spec text so deferred scope stays visible.

**Files:**
- Modify: `docs/superpowers/specs/2026-05-27-issue-52-command-file-design.md` (record issue numbers in § Non-goals + § Follow-up issues)

- [ ] **Step 1: File 6 follow-up issues against `2389-research/dippin-lang`**

Use `gh issue create -R 2389-research/dippin-lang`. Title pattern: `follow-up: <topic> (deferred from #52)`. Body template:

```
Filed per v0.33.0 spec § Follow-up issues #N.
See: https://github.com/2389-research/dippin-lang/blob/main/docs/superpowers/specs/2026-05-27-issue-52-command-file-design.md

<one-paragraph context from spec's non-goal text>
```

Label each with `enhancement` (or whatever label fits — `safety-follow-up` was issue-#41-specific; if no general "follow-up" label exists, plain `enhancement` is fine).

The six issues, in spec order:
1. Add `prompt_file:` and `system_prompt_file:` directives
2. Configurable size cap for `*_file:` directives
3. Symlink full-chain resolution via `EvalSymlinks`
4. Glob support for `command_file:` patterns
5. Preserve `command_file:` path through DOT round-trip
6. Graceful LSP/WASM "file directive seen; not loaded in this context" signal

Record the assigned issue numbers for Step 2.

- [ ] **Step 2: Update spec to reference filed issue numbers**

Edit `docs/superpowers/specs/2026-05-27-issue-52-command-file-design.md`. For each non-goal bullet (§ Non-goals 1–6) and each follow-up entry (§ Follow-up issues 1–6), append ` ([#N](https://github.com/2389-research/dippin-lang/issues/N))` to the bullet.

- [ ] **Step 3: Commit**

```bash
git add docs/superpowers/specs/2026-05-27-issue-52-command-file-design.md
git commit -m "docs(spec): link v0.33.0 follow-up issues into spec"
```

---

## Task 1: Add `CommandFile` field to `ir.ToolConfig`

**Files:**
- Modify: `ir/ir.go`

- [ ] **Step 1: Add the field**

Open `ir/ir.go`. Find `type ToolConfig struct`. Add `CommandFile string` immediately after `Command`:

```go
type ToolConfig struct {
	Command       string
	CommandFile   string // Source path when Command was loaded from command_file:; empty if inline. Populated by parser.ResolveFileDirectives.
	Timeout       time.Duration
	Outputs       []string
	MarkerGrep    string
	RouteRequired bool
	OutputLimit   int
}
```

- [ ] **Step 2: Verify build**

Run: `just build`
Expected: success (no test changes; new field that nothing reads yet).

- [ ] **Step 3: Run tests**

Run: `just test`
Expected: all packages pass (new field is zero-value in all existing constructions).

- [ ] **Step 4: Commit**

```bash
git add ir/ir.go
git commit -m "feat(ir): add ToolConfig.CommandFile field"
```

---

## Task 2: Parser handler for `command_file:` + mutual-exclusion error

**Files:**
- Test: `parser/parser_test.go` (extend)
- Modify: `parser/parse_nodes.go` (`applyToolField` to detect both-set; `applyToolStringField` for the new case)

- [ ] **Step 1: Write the failing tests**

Append to `parser/parser_test.go`. (First check the existing convention via `grep -n "func TestParseTool\|func TestParseAgent_ToolAccess" parser/parser_test.go` and mirror it — same-package `parser`, bare `NewParser`, real `.dip` source via `parser.NewParser(src, "test.dip").Parse()`.)

```go
func TestParseTool_CommandFile(t *testing.T) {
	cases := []struct {
		name             string
		src              string
		wantFile         string
		wantCommand      string
		wantDiagContains string // substring expected in any parser diagnostic; "" = no diagnostic expected
	}{
		{
			name: "command_file only",
			src: `workflow X
  start: A
  exit: A

  tool A
    command_file: scripts/setup.sh
`,
			wantFile:    "scripts/setup.sh",
			wantCommand: "",
		},
		{
			name: "command only (unchanged behavior)",
			src: `workflow X
  start: A
  exit: A

  tool A
    command: echo hi
`,
			wantFile:    "",
			wantCommand: "echo hi",
		},
		{
			name: "both set (command first)",
			src: `workflow X
  start: A
  exit: A

  tool A
    command: echo hi
    command_file: scripts/setup.sh
`,
			wantDiagContains: "both `command` and `command_file`",
		},
		{
			name: "both set (command_file first)",
			src: `workflow X
  start: A
  exit: A

  tool A
    command_file: scripts/setup.sh
    command: echo hi
`,
			wantDiagContains: "both `command` and `command_file`",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p := NewParser(tc.src, "test.dip")
			w, _ := p.Parse() // some cases expect diagnostics; don't fail on parse error
			if tc.wantDiagContains != "" {
				found := false
				for _, d := range p.Diagnostics() {
					if strings.Contains(d.Message, tc.wantDiagContains) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected diagnostic containing %q; got %v", tc.wantDiagContains, p.Diagnostics())
				}
				return
			}
			if w == nil {
				t.Fatalf("unexpected nil workflow")
			}
			node := w.Node("A")
			if node == nil {
				t.Fatalf("node A not found")
			}
			cfg, ok := node.Config.(ir.ToolConfig)
			if !ok {
				t.Fatalf("expected ToolConfig, got %T", node.Config)
			}
			if cfg.CommandFile != tc.wantFile {
				t.Errorf("CommandFile = %q, want %q", cfg.CommandFile, tc.wantFile)
			}
			if cfg.Command != tc.wantCommand {
				t.Errorf("Command = %q, want %q", cfg.Command, tc.wantCommand)
			}
		})
	}
}
```

If `p.Diagnostics()` isn't the exact accessor name, grep the test file for how other tests inspect diagnostics and match that pattern. Likely candidates: `p.diagnostics` (unexported, accessible since same-package) or a helper like `parseAndCollectDiags`.

- [ ] **Step 2: Run test to verify it fails**

Run: `just test-pkg parser`
Expected: FAIL — `TestParseTool_CommandFile` failures because `command_file` isn't a recognized key (gets the "unknown field" hint) and there's no mutual-exclusion check.

- [ ] **Step 3: Extend `applyToolStringField` for the new key**

Edit `parser/parse_nodes.go`. Find `applyToolStringField` (around line 422). Add a `case "command_file":` arm:

```go
func applyToolStringField(cfg *ir.ToolConfig, key, val string) bool {
	switch key {
	case "command":
		cfg.Command = val
	case "command_file":
		cfg.CommandFile = val
	case "outputs":
		cfg.Outputs = splitComma(val)
	case "marker_grep":
		cfg.MarkerGrep = val
	default:
		return false
	}
	return true
}
```

- [ ] **Step 4: Add the mutual-exclusion check**

Still in `parser/parse_nodes.go`, in `applyToolField` (the entry point at line 408), add the check AFTER the string-field-dispatch returns:

```go
func (p *Parser) applyToolField(cfg *ir.ToolConfig, key, val string, loc ir.SourceLocation) {
	if applyToolStringField(cfg, key, val) {
		if cfg.Command != "" && cfg.CommandFile != "" {
			p.emitDiag(loc, fmt.Sprintf(
				"tool node has both `command` and `command_file` set; choose one"))
		}
		return
	}
	if p.applyToolBoolField(cfg, key, val, loc) {
		return
	}
	if p.applyToolParsedField(cfg, key, val, loc) {
		return
	}
	p.emitUnknownFieldHint("tool", key, loc)
}
```

If `p.emitDiag` isn't the right call (it might be `p.appendDiagnostic`, `p.diag`, or appended-to-slice directly), check how `applyToolStringField`'s siblings emit diagnostics in the same file. Grep for `p.diagnostics = append` or `p.emit` to find the convention.

Note: the diagnostic message intentionally doesn't reference the node ID — the parser may not have the ID available at this point in field-application. The source location (line/col) is enough context for the author.

- [ ] **Step 5: Run test to verify it passes**

Run: `just test-pkg parser`
Expected: PASS — all four subtests of `TestParseTool_CommandFile`.

- [ ] **Step 6: Commit**

```bash
git add parser/parse_nodes.go parser/parser_test.go
git commit -m "feat(parser): accept command_file: on tool nodes + mutual-exclusion check"
```

---

## Task 3: Resolver (`parser/resolve.go`)

**Files:**
- Create: `parser/resolve.go`
- Create: `parser/resolve_test.go`
- Create: `parser/testdata/command_file/setup.sh` (fixture)
- Create: `parser/testdata/command_file/oversize.sh` (fixture, optional)
- Create: `parser/testdata/command_file/escape.sh` (used in symlink test only if `Lstat`-creatable)

- [ ] **Step 1: Create the fixture file**

```bash
mkdir -p parser/testdata/command_file
cat > parser/testdata/command_file/setup.sh << 'EOF'
#!/usr/bin/env bash
set -eu
echo "fixture: ResolveFileDirectives test"
EOF
```

- [ ] **Step 2: Write the failing tests**

Create `parser/resolve_test.go`:

```go
package parser

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/2389-research/dippin-lang/ir"
)

func TestResolveFileDirectives_LoadsContent(t *testing.T) {
	w := &ir.Workflow{
		Nodes: []*ir.Node{
			{ID: "A", Kind: ir.NodeTool, Config: ir.ToolConfig{
				CommandFile: "setup.sh",
			}},
		},
	}
	baseDir := "testdata/command_file"
	if err := ResolveFileDirectives(w, baseDir); err != nil {
		t.Fatalf("ResolveFileDirectives: %v", err)
	}
	cfg := w.Nodes[0].Config.(ir.ToolConfig)
	if !strings.Contains(cfg.Command, "fixture: ResolveFileDirectives test") {
		t.Errorf("Command not populated from file; got %q", cfg.Command)
	}
	if cfg.CommandFile != "setup.sh" {
		t.Errorf("CommandFile = %q, want %q", cfg.CommandFile, "setup.sh")
	}
}

func TestResolveFileDirectives_SkipsInline(t *testing.T) {
	w := &ir.Workflow{
		Nodes: []*ir.Node{
			{ID: "A", Kind: ir.NodeTool, Config: ir.ToolConfig{
				Command: "inline content",
			}},
		},
	}
	if err := ResolveFileDirectives(w, "testdata/command_file"); err != nil {
		t.Fatalf("ResolveFileDirectives: %v", err)
	}
	cfg := w.Nodes[0].Config.(ir.ToolConfig)
	if cfg.Command != "inline content" {
		t.Errorf("inline Command modified: %q", cfg.Command)
	}
}

func TestResolveFileDirectives_RejectsAbsolutePath(t *testing.T) {
	w := &ir.Workflow{
		Nodes: []*ir.Node{
			{ID: "A", Kind: ir.NodeTool, Config: ir.ToolConfig{
				CommandFile: "/etc/passwd",
			}},
		},
	}
	err := ResolveFileDirectives(w, "testdata/command_file")
	if err == nil || !strings.Contains(err.Error(), "absolute paths not allowed") {
		t.Errorf("expected absolute-path rejection; got %v", err)
	}
	// Error must reference the user-written path, NOT the resolved absolute path's parent
	if !strings.Contains(err.Error(), "/etc/passwd") {
		t.Errorf("error should reference user-written path /etc/passwd; got %v", err)
	}
}

func TestResolveFileDirectives_RejectsParentEscape(t *testing.T) {
	w := &ir.Workflow{
		Nodes: []*ir.Node{
			{ID: "A", Kind: ir.NodeTool, Config: ir.ToolConfig{
				CommandFile: "../../../etc/passwd",
			}},
		},
	}
	err := ResolveFileDirectives(w, "testdata/command_file")
	if err == nil || !strings.Contains(err.Error(), "resolves outside source directory") {
		t.Errorf("expected parent-escape rejection; got %v", err)
	}
}

func TestResolveFileDirectives_RejectsPostCleanEscape(t *testing.T) {
	w := &ir.Workflow{
		Nodes: []*ir.Node{
			{ID: "A", Kind: ir.NodeTool, Config: ir.ToolConfig{
				CommandFile: "legit/../../../etc/passwd",
			}},
		},
	}
	err := ResolveFileDirectives(w, "testdata/command_file")
	if err == nil || !strings.Contains(err.Error(), "resolves outside source directory") {
		t.Errorf("expected post-Clean escape rejection; got %v", err)
	}
}

func TestResolveFileDirectives_RejectsSymlink(t *testing.T) {
	// Create a symlink in a temp dir (avoids checked-in symlinks for cross-platform).
	tmp := t.TempDir()
	// Create the target file outside tmp first
	external := filepath.Join(t.TempDir(), "outside.txt")
	if err := os.WriteFile(external, []byte("external"), 0o644); err != nil {
		t.Fatalf("write external: %v", err)
	}
	linkPath := filepath.Join(tmp, "link.sh")
	if err := os.Symlink(external, linkPath); err != nil {
		t.Skipf("symlink not supported on this platform: %v", err)
	}
	w := &ir.Workflow{
		Nodes: []*ir.Node{
			{ID: "A", Kind: ir.NodeTool, Config: ir.ToolConfig{
				CommandFile: "link.sh",
			}},
		},
	}
	err := ResolveFileDirectives(w, tmp)
	if err == nil || !strings.Contains(err.Error(), "symlinks not allowed") {
		t.Errorf("expected symlink rejection; got %v", err)
	}
}

func TestResolveFileDirectives_RejectsOversize(t *testing.T) {
	tmp := t.TempDir()
	bigPath := filepath.Join(tmp, "big.sh")
	// Write 4 MiB + 1 byte
	big := make([]byte, (4<<20)+1)
	if err := os.WriteFile(bigPath, big, 0o644); err != nil {
		t.Fatalf("write big: %v", err)
	}
	w := &ir.Workflow{
		Nodes: []*ir.Node{
			{ID: "A", Kind: ir.NodeTool, Config: ir.ToolConfig{
				CommandFile: "big.sh",
			}},
		},
	}
	err := ResolveFileDirectives(w, tmp)
	if err == nil || !strings.Contains(err.Error(), "exceeds") {
		t.Errorf("expected oversize rejection; got %v", err)
	}
}

func TestResolveFileDirectives_MissingFile(t *testing.T) {
	w := &ir.Workflow{
		Nodes: []*ir.Node{
			{ID: "A", Kind: ir.NodeTool, Config: ir.ToolConfig{
				CommandFile: "nonexistent.sh",
			}},
		},
	}
	err := ResolveFileDirectives(w, "testdata/command_file")
	if err == nil {
		t.Errorf("expected missing-file error; got nil")
	}
	// Error must reference the user-written path
	if !strings.Contains(err.Error(), "nonexistent.sh") {
		t.Errorf("error should reference user-written path; got %v", err)
	}
}
```

- [ ] **Step 3: Run tests to verify they fail**

Run: `just test-pkg parser`
Expected: FAIL — `ResolveFileDirectives` undefined (compile error).

- [ ] **Step 4: Create `parser/resolve.go`**

```go
package parser

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/2389-research/dippin-lang/ir"
)

const maxDirectiveFileSize = 4 << 20 // 4 MiB

// ResolveFileDirectives loads file contents for every tool node with
// CommandFile set, populating Command from the file's bytes. Paths are
// resolved relative to baseDir (typically the directory of the .dip
// source file). Returns the first error encountered.
//
// Parser entry points (NewParser/Parse) do NOT call this — they stay pure.
// CLI entry points call it after parsing. LSP and WASM contexts skip it;
// the IR retains the CommandFile-set / Command-empty state, which is the
// correct unresolved view for those consumers.
func ResolveFileDirectives(w *ir.Workflow, baseDir string) error {
	for _, n := range w.Nodes {
		tc, ok := n.Config.(ir.ToolConfig)
		if !ok || tc.CommandFile == "" {
			continue
		}
		// Skip if Command is already populated (inline takes precedence on the
		// off chance both are set — Task 2's parser check should prevent this,
		// but defensive).
		if tc.Command != "" {
			continue
		}
		contents, err := loadDirectiveFile(baseDir, tc.CommandFile)
		if err != nil {
			return fmt.Errorf("node %q command_file: %w", n.ID, err)
		}
		tc.Command = string(contents)
		n.Config = tc
	}
	return nil
}

// loadDirectiveFile resolves p relative to baseDir, applies path security
// checks, and reads the file. Returns the file's bytes or a clean error.
// Error messages reference the user-written path (p), never the resolved
// absolute path, to avoid leaking directory structure into diagnostics.
func loadDirectiveFile(baseDir, p string) ([]byte, error) {
	if filepath.IsAbs(p) {
		return nil, fmt.Errorf("absolute paths not allowed: %q", p)
	}
	resolved := filepath.Join(baseDir, p)
	rel, err := filepath.Rel(baseDir, resolved)
	if err != nil || hasParentRef(rel) || filepath.IsAbs(rel) {
		return nil, fmt.Errorf("path %q resolves outside source directory", p)
	}
	info, err := os.Lstat(resolved)
	if err != nil {
		return nil, fmt.Errorf("path %q: %w", p, err)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return nil, fmt.Errorf("symlinks not allowed: %q", p)
	}
	if info.Size() > maxDirectiveFileSize {
		return nil, fmt.Errorf("file %q exceeds %d byte limit (size %d)", p, maxDirectiveFileSize, info.Size())
	}
	return os.ReadFile(resolved)
}

// hasParentRef returns true if rel contains a `..` path segment.
// Used as defensive belt after filepath.Rel; should not fire on a
// non-symlink filesystem since Rel canonicalizes already.
func hasParentRef(rel string) bool {
	for _, seg := range strings.Split(filepath.ToSlash(rel), "/") {
		if seg == ".." {
			return true
		}
	}
	return false
}
```

- [ ] **Step 5: Run tests to verify they pass**

Run: `just test-pkg parser`
Expected: PASS — all 7+ `TestResolveFileDirectives_*` tests + the existing `TestParseTool_CommandFile` from Task 2.

- [ ] **Step 6: Check complexity**

Run: `just complexity`
Expected: pass. If `loadDirectiveFile` trips the ≤5 cyc cap, extract each check to a predicate helper (`isAbsolute`, `escapesBase`, `isSymlink`, `tooLarge`) per CLAUDE.md.

- [ ] **Step 7: Commit**

```bash
git add parser/resolve.go parser/resolve_test.go parser/testdata/command_file/
git commit -m "feat(parser): add ResolveFileDirectives with path security + size cap"
```

---

## Task 4: CLI integration — call resolver after parse

**Files:**
- Modify: `cmd/dippin/cli.go` (extend the two parse helpers around lines 245-275)

- [ ] **Step 1: Inspect the parse helpers**

Run: `grep -n "p := parser.NewParser" cmd/dippin/cli.go`

Expected: two locations (around line 249 and 273). Both end with `return p.Parse()`. Confirm by reading the surrounding 5-10 lines.

- [ ] **Step 2: Edit both to call the resolver**

For each helper, change:

```go
p := parser.NewParser(string(data), path)
return p.Parse()
```

to:

```go
p := parser.NewParser(string(data), path)
w, err := p.Parse()
if err != nil {
    return nil, err
}
if err := parser.ResolveFileDirectives(w, filepath.Dir(path)); err != nil {
    return nil, err
}
return w, nil
```

The `filepath.Dir(path)` is the baseDir. `path` is the workflow source file's path (from the function's arg). Add `"path/filepath"` to the file's imports if not already present (it almost certainly is).

- [ ] **Step 3: Build + test**

Run: `just build && just test`
Expected: success. No existing tests use `command_file:` yet, so the new resolver call is a no-op for them.

- [ ] **Step 4: Manual sanity check**

Create a quick fixture:

```bash
mkdir -p /tmp/dippin-cf-smoke
cat > /tmp/dippin-cf-smoke/setup.sh << 'EOF'
echo "from external file"
EOF
cat > /tmp/dippin-cf-smoke/wf.dip << 'EOF'
workflow X
  start: A
  exit: A

  tool A
    command_file: setup.sh
EOF

./bin/dippin lint /tmp/dippin-cf-smoke/wf.dip
```

Expected: lint passes (no diagnostics).

- [ ] **Step 5: Manual error-path check**

```bash
cat > /tmp/dippin-cf-smoke/wf-bad.dip << 'EOF'
workflow X
  start: A
  exit: A

  tool A
    command_file: nonexistent.sh
EOF

./bin/dippin lint /tmp/dippin-cf-smoke/wf-bad.dip
```

Expected: error message referencing `nonexistent.sh`, not the absolute path of `/tmp/dippin-cf-smoke/`.

- [ ] **Step 6: Commit**

```bash
git add cmd/dippin/cli.go
git commit -m "feat(cli): call ResolveFileDirectives after parse in CLI entry points"
```

---

## Task 5: Formatter — preserve directive form

**Files:**
- Test: `formatter/format_test.go` (extend)
- Modify: `formatter/format.go` (`writeToolFields` around line 503-516)

- [ ] **Step 1: Write the failing tests**

Append to `formatter/format_test.go`:

```go
func TestFormat_ToolCommandFile(t *testing.T) {
	src := `workflow X
  start: A
  exit: A

  tool A
    command_file: scripts/setup.sh
`
	p := parser.NewParser(src, "test.dip")
	w, err := p.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	out := Format(w)
	if !strings.Contains(out, "command_file: scripts/setup.sh") {
		t.Errorf("formatted output missing command_file directive:\n%s", out)
	}
	// And it should NOT also emit a multi-line command: block
	if strings.Contains(out, "command:") {
		t.Errorf("formatted output should not contain inline command: when file form used:\n%s", out)
	}
}

func TestFormat_ToolCommandInline(t *testing.T) {
	src := `workflow X
  start: A
  exit: A

  tool A
    command: echo hi
`
	p := parser.NewParser(src, "test.dip")
	w, err := p.Parse()
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	out := Format(w)
	if !strings.Contains(out, "command:") {
		t.Errorf("formatted output missing inline command:\n%s", out)
	}
	if strings.Contains(out, "command_file:") {
		t.Errorf("formatted output should not contain command_file when inline used:\n%s", out)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `just test-pkg formatter`
Expected: FAIL — formatter doesn't emit `command_file:`.

- [ ] **Step 3: Update `writeToolFields`**

Edit `formatter/format.go`, around line 503. Replace the final `if cfg.Command != ""` block:

```go
func writeToolFields(wr *writer, n *ir.Node, cfg ir.ToolConfig) {
	writeCommonNodeFields(wr, n)
	if len(cfg.Outputs) > 0 {
		wr.line("outputs: %s", strings.Join(cfg.Outputs, ", "))
	}
	writeToolRoutingFields(wr, cfg)
	if cfg.Timeout != 0 {
		wr.line("timeout: %s", formatDuration(cfg.Timeout))
	}
	writeIOFields(wr, n)
	if cfg.CommandFile != "" {
		wr.line("command_file: %s", quoteValue(cfg.CommandFile))
	} else if cfg.Command != "" {
		wr.multilineBlock("command", cfg.Command)
	}
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `just test-pkg formatter`
Expected: PASS — both `TestFormat_ToolCommandFile` and `TestFormat_ToolCommandInline`.

- [ ] **Step 5: Commit**

```bash
git add formatter/format.go formatter/format_test.go
git commit -m "feat(formatter): preserve command_file directive across dippin fmt"
```

---

## Task 6: Migrate parity comparator

**Files:**
- Modify: `migrate/parity.go` (`compareToolConfigs` / `compareToolScalars`)

- [ ] **Step 1: Find the comparator**

Run: `grep -n "compareToolScalars\|compareToolConfigs" migrate/parity.go`

Expected: hits around lines 260-275. `compareToolConfigs` calls `compareToolScalars`.

- [ ] **Step 2: Read the comparator structure**

Open `migrate/parity.go`. Read `compareToolScalars` and any sibling per-field comparator functions (`compareToolCommandAndTimeout`, `compareToolMarkerAndRoute`, `compareToolOutputLimit` per the v0.32 context). Find where `Command` is compared.

- [ ] **Step 3: Add `CommandFile` to the field-equality list**

If `Command` is compared in a function called `compareToolCommandAndTimeout` (or similar), add `CommandFile` alongside it:

```go
func compareToolCommandAndTimeout(id string, ac, bc ir.ToolConfig) []Difference {
	var diffs []Difference
	if ac.Command != bc.Command {
		diffs = append(diffs, fieldDiff(id, "command", fmt.Sprintf("node %q command differs", id)))
	}
	if ac.CommandFile != bc.CommandFile {
		diffs = append(diffs, fieldDiff(id, "command_file", fmt.Sprintf("node %q command_file: %q vs %q", id, ac.CommandFile, bc.CommandFile)))
	}
	if ac.Timeout != bc.Timeout {
		// ... existing timeout check ...
	}
	return diffs
}
```

Match the existing style in the file — don't invent a new function shape.

- [ ] **Step 4: Run tests**

Run: `just test-pkg migrate`
Expected: PASS — existing parity tests don't use `command_file:` so the new comparison is a no-op (both sides empty).

- [ ] **Step 5: Commit**

```bash
git add migrate/parity.go
git commit -m "feat(migrate): include CommandFile in tool parity comparator"
```

---

## Task 7: Example file + integration test

**Files:**
- Create: `examples/external_files.dip`
- Create: `examples/external_files/setup.sh`

- [ ] **Step 1: Create the script fixture**

```bash
mkdir -p examples/external_files
cat > examples/external_files/setup.sh << 'EOF'
set -eu
echo "Running external setup script..."
mkdir -p .ai/scratch
echo "ready" > .ai/scratch/setup.done
EOF
chmod +x examples/external_files/setup.sh
```

- [ ] **Step 2: Create the example workflow**

```bash
cat > examples/external_files.dip << 'EOF'
workflow ExternalFiles
  goal: "Demonstrate command_file: directive (issue #52)"
  start: Setup
  exit: Verify

  tool Setup
    command_file: external_files/setup.sh

  agent Verify
    model: claude-sonnet-4-6
    prompt: "Confirm setup ran successfully. STATUS: success or STATUS: fail."
    auto_status: true

  edges
    Setup -> Verify
EOF
```

- [ ] **Step 3: Run validate-examples**

Run: `just validate-examples`
Expected: success — the new example must pass.

Note: `dippin validate` runs through `loadWorkflow` in `cmd/dippin/cli.go` which (after Task 4) calls `ResolveFileDirectives`. So the example actually exercises the resolver end-to-end.

- [ ] **Step 4: Run `TestLintExamples`**

Run: `just test-pkg validator`
Expected: PASS — `TestLintExamples` iterates all `examples/*.dip` through the linter. The new example must produce zero lint warnings.

**Important:** `TestLintExamples` parses via `parser.NewParser(...).Parse()` directly (no CLI), so it does NOT call `ResolveFileDirectives`. That means `cfg.Command == ""` for the Setup node. Most lints don't care, but if any lint requires `cfg.Command` to be non-empty for tool nodes, that lint would fire here. Check the output; if a regression appears, the lint needs a defensive check for `cfg.CommandFile != ""` (the unresolved-but-intentional case).

- [ ] **Step 5: Run `just check`**

Run: `just check`
Expected: PASS — full Go suite + complexity + validate-examples.

- [ ] **Step 6: Commit**

```bash
git add examples/external_files.dip examples/external_files/
git commit -m "examples: add external_files.dip demonstrating command_file directive (#52)"
```

---

## Task 8: Documentation updates

**Files:**
- Modify: `docs/nodes.md`
- Modify: `docs/llm-reference.md`
- Modify: `docs/GRAMMAR.ebnf`
- Modify: `site/static/skill.md`
- Modify: `CHANGELOG.md`

- [ ] **Step 1: Add field row to `docs/nodes.md`**

Open `docs/nodes.md`. Find the **tool node** fields table (search for `command` field row). Add a row immediately after `command`:

```markdown
| `command` | String (multiline) | Required (unless `command_file`) | Shell command(s) to execute. |
| `command_file` | String | — | Path (relative to the `.dip` source directory) to an external file whose contents replace inline `command:`. Mutually exclusive with `command`. Loaded at CLI entry points (`dippin lint`, `pack`, `validate`); LSP and playground see the path unresolved. Path security: absolute paths rejected, parent-tree escape rejected, symlinks rejected, 4 MiB size cap. |
```

(Adjust the wording for the existing `command` row if it's currently labeled "Required" — it's now "Required unless command_file is set". Be conservative; don't rewrite unrelated rows.)

- [ ] **Step 2: Update `docs/llm-reference.md`**

Open `docs/llm-reference.md`. Find the tool optional-fields list. Add `command_file` after `command`:

```markdown
| `tool` | `command` (or `command_file`) | `timeout` (e.g. 30s, 5m), `outputs` (CSV), `marker_grep` (regex), `route_required` (bool), `output_limit` (bytes), `command_file` (path to external script, relative to .dip dir) |
```

- [ ] **Step 3: Update `docs/GRAMMAR.ebnf`**

Open `docs/GRAMMAR.ebnf`. Find the tool field productions (search `"command"`):

```ebnf
tool_field = "command" ":" field_value
           | "command_file" ":" field_value
           | "timeout" ":" DURATION
           | "outputs" ":" identifier_list
           | ... ;
```

Add the `"command_file"` line directly after `"command"`.

- [ ] **Step 4: Update `site/static/skill.md`**

Open `site/static/skill.md`. Find the **tool fields** table (likely near the agent table). Add a row for `command_file:`. Plus a short subsection:

```markdown
**Tool Command File (`command_file:`)** — *added v0.33.0.*

Reference an external file for a tool node's command instead of inlining a heredoc:

```dip
tool Setup
  command_file: scripts/setup.sh
```

Path resolution: relative to the `.dip` source directory. Absolute paths rejected. Symlinks rejected. Parent-tree escape (`../../etc/passwd`) rejected. 4 MiB size cap.

Mutually exclusive with `command:` — specifying both is a parse error.

Loading: CLI entry points (`dippin lint`, `dippin pack`, `dippin validate`, `dippin doctor`) load the file contents into the IR after parse. The LSP and the playground (cmd/wasm) skip loading; they show the path unresolved. Tracker reads `.dipx` bundles where content is already inlined, so the runtime sees no difference from inline `command:`.

Non-goals (deferred): `prompt_file:` / `system_prompt_file:` directives, glob expansion, configurable size cap, DOT round-trip preservation of the directive form. See follow-up issues linked from issue [#52](https://github.com/2389-research/dippin-lang/issues/52).
```

- [ ] **Step 5: Add CHANGELOG entry stub**

Edit `CHANGELOG.md`. Add at the top (above v0.32.0):

```markdown
## [v0.33.0] — RELEASE_DATE

New `command_file:` directive on tool nodes replaces inline `command:` heredocs with external file references. Solves the heredoc-bloat pattern seen in long tracker workflows. Dippin-only release — no tracker coordination required (tracker reads inlined `Command` from `.dipx` bundles unchanged).

### Added
- `command_file: <path>` directive on tool nodes. Path is relative to the `.dip` source directory.
- `parser.ResolveFileDirectives(w, baseDir)` — separate-pass file loader called by CLI entry points. Parser itself stays pure.
- Path security in the resolver: absolute-path reject, parent-tree-escape reject, symlink reject (via `Lstat`), 4 MiB size cap. Error messages reference user-written paths, not resolved absolute paths.
- Parser-time error when both `command:` and `command_file:` are set on the same tool node.
- `examples/external_files.dip` + `examples/external_files/setup.sh` demonstrate the directive.

### Notes
- LSP and `cmd/wasm` (playground) skip the resolver. They see `cfg.CommandFile != "" && cfg.Command == ""`, which is the correct unresolved-IR view. Lints that need command content (none today on tool nodes) would need a defensive check for the unresolved case.
- DOT round-trip is lossy for the directive form — pack-then-unpack rewrites `command_file:` to inline `command:`. Tracker reads from `.dipx` bundles where content is already inlined; no current consumer needs the path preserved through DOT. Deferred to a follow-up issue.
```

The `RELEASE_DATE` placeholder gets filled at Task 10 tag time.

- [ ] **Step 6: Run `just check`**

Run: `just check`
Expected: PASS. Pre-commit hook auto-regenerates `cmd/dippin/generated-spec.md` (and the site-content changelog mirror via the changelog-md recipe if it runs in the hook).

- [ ] **Step 7: Commit**

```bash
git add docs/nodes.md docs/llm-reference.md docs/GRAMMAR.ebnf site/static/skill.md CHANGELOG.md
git commit -m "docs: document command_file directive + v0.33.0 changelog stub"
```

(Pre-commit may also stage `cmd/dippin/generated-spec.md` and `site/content/changelog.md` if it auto-regenerates them — that's expected.)

---

## Task 9: LSP completion + VSCode grammar

**Files:**
- Modify: `lsp/completion.go`
- Modify: `lsp/lsp_test.go`
- Modify: `editors/vscode/syntaxes/dippin.tmLanguage.json` (line ~148)

- [ ] **Step 1: Add LSP completion entry**

Open `lsp/completion.go`. Find `fieldCompletions` (around line 42-67). Add an entry alongside `marker_grep:` (since both are tool-node string fields):

```go
{"marker_grep:", "Regex matched against tool stdout; sets ctx.tool_marker"},
{"command_file:", "External script reference for tool node (path relative to .dip dir)"},
{"route_required:", "Require _TRACKER_ROUTE= sentinel line from tool stdout"},
```

- [ ] **Step 2: Add LSP completion regression test**

Open `lsp/lsp_test.go`. Find `TestFieldCompletionsIncludesToolAccess` (from v0.32 work) — mirror its shape:

```go
func TestFieldCompletionsIncludesCommandFile(t *testing.T) {
	items := fieldCompletions()
	for _, it := range items {
		if it.Label == "command_file:" {
			return
		}
	}
	t.Error("missing completion for 'command_file:' (v0.33.0 directive)")
}
```

- [ ] **Step 3: Update VSCode TextMate grammar**

Open `editors/vscode/syntaxes/dippin.tmLanguage.json`. Find the field-name regex (line ~148). Add `command_file` to the alternation:

```json
"match": "^\\s*(label|model|provider|mode|default|ref|subgraph_ref|timeout|poll_interval|max_cycles|fidelity|reasoning_effort|goal_gate|auto_status|max_turns|max_retries|marker_grep|command_file|route_required|output_limit|base_delay|retry_policy|retry_target|fallback_target|max_restarts|restart_target|cache_tools|compaction|stop_condition|steer_condition|steer_context|tool_access|class|reads|writes)\\s*(:)",
```

(Add `command_file` after `marker_grep`, keeping the alphabetical-ish grouping near sibling tool fields.)

- [ ] **Step 4: Run LSP tests**

Run: `just test-pkg lsp`
Expected: PASS — `TestFieldCompletionsIncludesCommandFile` plus all existing completion tests.

- [ ] **Step 5: Verify VSCode grammar parses**

The TextMate file is JSON — `just check` will catch malformed JSON via the gen-spec step. Run `just check` to confirm.

- [ ] **Step 6: Commit**

```bash
git add lsp/completion.go lsp/lsp_test.go editors/vscode/syntaxes/dippin.tmLanguage.json
git commit -m "feat(integration): LSP completion + VSCode highlighting for command_file"
```

---

## Task 10: Tag v0.33.0

**Files:**
- Modify: `CHANGELOG.md` (substitute `RELEASE_DATE`)

- [ ] **Step 1: Open PR (if not already on a feature branch + open PR)**

If working on `main` directly:

```bash
git push origin main
```

If working on a feature branch:

```bash
git push -u origin <branch-name>
gh pr create --title "v0.33.0: command_file: directive on tool nodes (#52)" --body "$(cat <<'EOF'
## Summary

Closes [#52](https://github.com/2389-research/dippin-lang/issues/52).

Adds `command_file:` directive on tool nodes. Author can replace inline `command:` heredoc with external file reference. Parser stays pure; a separate `ResolveFileDirectives` pass runs at CLI entry points. LSP and WASM playground skip the resolver and see the unresolved IR view.

Dippin-internal — no tracker coordination needed.

## Spec
docs/superpowers/specs/2026-05-27-issue-52-command-file-design.md

## Test plan
- [ ] `just check` passes
- [ ] `examples/external_files.dip` lints clean
- [ ] Six follow-up issues filed and linked in the spec § Non-goals

🤖 Generated with [Claude Code](https://claude.com/claude-code)
EOF
)"
```

- [ ] **Step 2: Address PR review feedback**

Use the prbuddy:reviews pattern (or just `gh pr-review threads list <PR#> --unresolved`) to triage CodeRabbit + Copilot comments. Fix in small follow-up commits; reply + resolve each thread.

- [ ] **Step 3: Merge PR**

After approval:

```bash
gh pr merge <PR#> --squash  # or --merge per project convention
git checkout main && git pull
```

- [ ] **Step 4: Substitute the RELEASE_DATE placeholder**

```bash
# Today's date in YYYY-MM-DD form
DATE=$(date -u +%Y-%m-%d)
sed -i "s/RELEASE_DATE/$DATE/" CHANGELOG.md
# Pre-commit will regenerate site/content/changelog.md
git add CHANGELOG.md
git commit -m "release: prep v0.33.0"
```

- [ ] **Step 5: Tag and push**

```bash
git tag -a v0.33.0 -m "command_file: directive on tool nodes (#52)"
git push origin main
git push origin v0.33.0
```

GoReleaser handles cross-platform binary build + Homebrew tap update.

- [ ] **Step 6: Verify release**

```bash
gh run list --workflow=release.yml --limit 1 -R 2389-research/dippin-lang
gh release view v0.33.0 -R 2389-research/dippin-lang
```

Expected: GoReleaser run in_progress → success; release published with 5 assets (checksums + Darwin/Linux × arm64/x86_64 tarballs).

- [ ] **Step 7: Close issue #52**

```bash
gh issue close 52 -R 2389-research/dippin-lang --comment "Shipped in [v0.33.0](https://github.com/2389-research/dippin-lang/releases/tag/v0.33.0). See [CHANGELOG](https://github.com/2389-research/dippin-lang/blob/main/CHANGELOG.md). Six follow-up issues remain open for deferred scope (linked from the spec § Non-goals)."
```

---

## Notes for the executing engineer

- **`just check` is your friend.** Run after every task. Pre-commit hook runs it too.
- **Cyclomatic ≤ 5 / cognitive ≤ 7 per function.** If `loadDirectiveFile` trips the cap, extract predicate helpers per CLAUDE.md. Never add `//nolint`.
- **No hand-built IR in tests where avoidable.** Parser tests should parse real `.dip` text. Resolver tests build IR directly (no `.dip` source needed — they test the resolver, not the parser).
- **Pre-commit auto-regenerates** `cmd/dippin/generated-spec.md` and `site/content/changelog.md`. Stage them when they appear in `git status`.
- **WASM verification:** the playground (`cmd/wasm/main.go`) parses but doesn't call `ResolveFileDirectives`. If you find a path where playground users would see broken behavior with `command_file:` (e.g., a downstream consumer panicking on empty `Command`), file follow-up #6 and add a graceful diagnostic in this PR.
- **Don't merge with bot review state in play.** Branch protection on dippin `main` is not enforced; always confirm with the user before merging if ANY review state is non-APPROVED (per CLAUDE.md notes).
