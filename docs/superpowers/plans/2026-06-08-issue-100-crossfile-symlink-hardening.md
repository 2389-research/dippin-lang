# Cross-File Resolver Symlink/Root-Escape Hardening (#100) Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make DIP146's cross-file lint refuse symlinked / root-escaping child `.dip` refs — fail-soft (refusal → `postureUnresolved` → DIP143 retained), reaching parity with the pack walker.

**Architecture:** Reuse two existing primitives — `ensureUnderRoot` (`cmd/dippin/pack_shadow.go`, lexical `..`-escape check, already in `package main`) and `dipx`'s `readNoFollowSymlinks` (Lstat-based leaf+ancestor symlink refusal) — exporting only the latter (one new public symbol). Compute a fixed absolute containment root (the entry file's directory) once and thread it unchanged through the recursive walk; route the child read through the two checks before parsing. No new DIP code, no build tags, no pack-path refactor.

**Tech Stack:** Go; `just` recipes for build/test (never raw `go`); pre-commit hook is the CI gate.

**Spec:** `docs/superpowers/specs/2026-06-08-issue-100-design.md` (read it before starting).

**Worktree:** Work entirely in `/home/clint/code/2389/dippin-lang/.claude/worktrees/fix+100-crossfile-symlink-hardening` on branch `fix/100-crossfile-symlink-hardening`. Stage explicit paths only (never `git add -A` — `./dippin`, `./wasm`, `site/static/*.wasm` are gitignored build artifacts). Every commit ends with the trailer `Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>`.

---

## File Structure

- **`dipx/helpers.go`** (modify) — rename `readNoFollowSymlinks` → `ReadNoFollowSymlinks` (export), generalize its doc comment, update its one caller at `:317`.
- **`dipx/dipx_test.go`** (modify) — add `TestReadNoFollowSymlinks` (direct unit test of the newly-exported primitive).
- **`cmd/dippin/pack_shadow.go`** (modify) — generalize `ensureUnderRoot`'s error message + doc comment (now shared).
- **`cmd/dippin/crossfile_tool_access.go`** (modify) — replace `resolveBoundaryRefPath` with `resolveChildPath`; thread a fixed absolute `root` through the walk; route the read through `dipx.ReadNoFollowSymlinks`; swap the `os` import for `dipx`.
- **`cmd/dippin/crossfile_tool_access_test.go`** (modify) — add the refusal + regression tests (T1, T2, T3, T4, T5, T6, T7, T8, T9, T11).
- **`docs/architecture.md`** (modify) — one-line responsibility note on the `helpers.go` line.

---

## Task 1: Export the dipx no-follow read primitive (rename)

**Files:**
- Modify: `dipx/helpers.go:385-411` (doc comment + `func readNoFollowSymlinks`), `dipx/helpers.go:317` (caller)

This is a pure rename + doc-comment generalization. The existing dipx tests (`TestPack_RejectsSymlink`, `TestPack_RejectsParentSymlink`) are the regression guard — they exercise this code via `Pack` and must stay green.

- [ ] **Step 1: Update the caller at `helpers.go:317`**

In `func (s *packWalkState) readAndRecord`, change:

```go
	raw, err := readNoFollowSymlinks(cur, s.rootDir)
```

to:

```go
	raw, err := ReadNoFollowSymlinks(cur, s.rootDir)
```

- [ ] **Step 2: Rename the function and generalize its doc comment (`helpers.go:385-396`)**

Replace the doc comment + signature line (from `// readNoFollowSymlinks reads a file…` down through `func readNoFollowSymlinks(path, rootDir string) ([]byte, error) {`) with:

```go
// ReadNoFollowSymlinks reads a file, refusing to follow symlinks at the leaf OR
// at any intermediate path component between rootDir and path. Shared by the
// Pack walker and the cross-file tool_access lint (cmd/dippin) so both refuse
// symlinks identically. It closes a parent-component-symlink data-exfil vector:
// a tree containing `sub -> /etc` would otherwise let a leaf `sub/foo.dip` read
// `/etc/foo.dip`, because Lstat on the leaf reports a regular file, not a symlink.
//
// It does NOT perform the `..`-escape (containment) check — callers must run that
// separately BEFORE calling this (the pack walker via resolveRefOnDisk, the lint
// via ensureUnderRoot). rootDir is only the ancestor-scan boundary here.
//
// rootDir itself is treated as the trust anchor: it is an absolute path supplied
// by the caller, may itself be a user-specified symlink, and is not re-validated.
// Components strictly between rootDir and path's leaf MUST be directories that are
// not symlinks.
func ReadNoFollowSymlinks(path, rootDir string) ([]byte, error) {
```

- [ ] **Step 3: Run the dipx tests to verify the rename is clean**

Run: `just test-pkg dipx`
Expected: PASS (all dipx tests, including `TestPack_RejectsSymlink` / `TestPack_RejectsParentSymlink`). No "undefined: readNoFollowSymlinks" or "declared and not used".

- [ ] **Step 4: Commit**

```bash
git add dipx/helpers.go
git commit -m "refactor(dipx): export readNoFollowSymlinks as ReadNoFollowSymlinks

Capitalize + generalize the no-follow read primitive's doc comment so the
cross-file tool_access lint (#100) can share it with the Pack walker. Pure
rename of merged #79/#85 code; behavior unchanged; sole caller updated.

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

## Task 2: Unit-test the exported primitive (T10)

**Files:**
- Modify: `dipx/dipx_test.go` (add `TestReadNoFollowSymlinks`; `package dipx` internal test — `ReadNoFollowSymlinks`, `ErrPathUnsafe`, `minimalDipSrc` are all in scope)

- [ ] **Step 1: Write the test**

Append to `dipx/dipx_test.go`:

```go
// TestReadNoFollowSymlinks pins the contract of the now-exported no-follow read
// primitive: happy path reads, leaf-symlink refusal, ancestor-symlink refusal,
// and the not-regular-file branch (the documented fail-soft path for a ref that
// resolves to a directory, including the root dir itself).
func TestReadNoFollowSymlinks(t *testing.T) {
	root := t.TempDir()

	good := filepath.Join(root, "good.dip")
	if err := os.WriteFile(good, []byte(minimalDipSrc), 0o644); err != nil {
		t.Fatal(err)
	}
	data, err := ReadNoFollowSymlinks(good, root)
	if err != nil {
		t.Fatalf("happy path: unexpected err %v", err)
	}
	if string(data) != minimalDipSrc {
		t.Fatalf("happy path: content mismatch")
	}

	link := filepath.Join(root, "link.dip")
	if err := os.Symlink(good, link); err != nil {
		t.Skip("symlinks not supported on this platform")
	}
	if _, err := ReadNoFollowSymlinks(link, root); !errors.Is(err, ErrPathUnsafe) {
		t.Fatalf("leaf symlink: err = %v, want ErrPathUnsafe", err)
	}

	realdir := filepath.Join(root, "real")
	if err := os.Mkdir(realdir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(realdir, "f.dip"), []byte(minimalDipSrc), 0o644); err != nil {
		t.Fatal(err)
	}
	linkdir := filepath.Join(root, "linkdir")
	if err := os.Symlink(realdir, linkdir); err != nil {
		t.Skip("symlinks not supported on this platform")
	}
	if _, err := ReadNoFollowSymlinks(filepath.Join(linkdir, "f.dip"), root); !errors.Is(err, ErrPathUnsafe) {
		t.Fatalf("ancestor symlink: err = %v, want ErrPathUnsafe", err)
	}

	if _, err := ReadNoFollowSymlinks(realdir, root); !errors.Is(err, ErrPathUnsafe) {
		t.Fatalf("not-regular-file (directory): err = %v, want ErrPathUnsafe", err)
	}
}
```

- [ ] **Step 2: Run the test**

Run: `just test-pkg dipx`
Expected: PASS (including the new `TestReadNoFollowSymlinks`). If `dipx_test.go` is missing an `errors`/`os`/`path/filepath` import, the build will fail — add it (these are almost certainly already imported; verify before adding).

- [ ] **Step 3: Commit**

```bash
git add dipx/dipx_test.go
git commit -m "test(dipx): unit-test ReadNoFollowSymlinks (leaf/ancestor/not-regular)

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

## Task 3: Generalize `ensureUnderRoot` (now shared)

**Files:**
- Modify: `cmd/dippin/pack_shadow.go:16-29` (doc comment + error message)

`ensureUnderRoot` becomes shared between the pack-shadow writer and the cross-file lint. Drop the `"pack-inline:"` prefix from its error (it now fires from two subsystems) and note the shared use. Behavior is unchanged; existing pack tests guard it.

- [ ] **Step 1: Replace the doc comment + function body (`pack_shadow.go:16-29`)**

Replace:

```go
// ensureUnderRoot rejects any absolute path that resolves outside rootDir.
// Guards the shadow-tree writer against a malicious subgraph ref like
// `subgraph: ../../../escape.dip`, which would otherwise cause
// writeShadowFile to write outside the temp shadow dir (since
// filepath.Join(shadowDir, "../../../escape.dip") escapes). dipx.Pack does
// its own escape check, but it runs AFTER we've built the shadow tree —
// too late to prevent the out-of-tree write.
func ensureUnderRoot(absPath, rootDir string) error {
	rel, err := filepath.Rel(rootDir, absPath)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return fmt.Errorf("pack-inline: ref escapes source root: %s", absPath)
	}
	return nil
}
```

with:

```go
// ensureUnderRoot rejects any absolute path that resolves outside rootDir, via
// the same lexical `..`-component check the pack walker uses (resolveRefOnDisk).
// Shared by two CLI subsystems: the pack shadow-tree writer (guarding against a
// ref like `subgraph: ../../../escape.dip` that would make writeShadowFile write
// out of the temp dir) and the cross-file tool_access lint (#100), which uses it
// fail-soft to refuse root-escaping child refs. absPath MUST be absolute (the
// callers absolutize before calling); a relative rootDir makes filepath.Rel error
// and is reported as an escape.
func ensureUnderRoot(absPath, rootDir string) error {
	rel, err := filepath.Rel(rootDir, absPath)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return fmt.Errorf("ref escapes source root: %s", absPath)
	}
	return nil
}
```

- [ ] **Step 2: Run the dippin package tests**

Run: `just test-pkg dippin`
Expected: PASS. (No existing test asserts on the `"pack-inline:"` string — verify with `grep -rn "pack-inline" cmd/dippin` returning only the now-changed source if any; if a test asserts the old message, update that assertion.)

- [ ] **Step 3: Commit**

```bash
git add cmd/dippin/pack_shadow.go
git commit -m "refactor(cli): generalize ensureUnderRoot for shared use (#100)

Drop the pack-specific 'pack-inline:' error prefix and broaden the doc comment
now that the cross-file tool_access lint reuses this escape check. Behavior
unchanged.

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

## Task 4: Write the refusal tests (RED)

**Files:**
- Modify: `cmd/dippin/crossfile_tool_access_test.go` (add T1, T2, T6, T11)

These four assert the *new* refusal behavior and therefore FAIL against current code (which follows symlinks / escaping refs and fires DIP146). They use the existing `writeWorkflows` / `crossDiags` / `countCode` helpers and existing fixture constants (`entryRestrictsRefs`, `childZeroIntent`). Symlink tests skip if `os.Symlink` is unsupported.

- [ ] **Step 1: Add the four tests**

Append to `cmd/dippin/crossfile_tool_access_test.go`:

```go
// childWithBoundary is a full-restrict child that itself delegates to `ref`,
// so a refusal/resolution can be observed one hop below the entry.
func childWithBoundary(ref string) string {
	return `workflow Child
  start: Lock
  exit: Sup

  agent Lock
    prompt: "x"
    tool_access: none

  manager_loop Sup
    subgraph_ref: ` + ref + `
    max_cycles: 3

  edges
    Lock -> Sup
`
}

func hasPosture(classified map[ir.SourceLocation]childPosture, want childPosture) bool {
	for _, p := range classified {
		if p == want {
			return true
		}
	}
	return false
}

// T1 — a boundary child ref that is a SYMLINK (even to a legitimate in-root .dip)
// is refused: fail-soft -> postureUnresolved, DIP143 retained (supersedes==false),
// no DIP146, no error. Also pins T5's policy: a benign in-root symlink target is
// still refused unconditionally (do not "fix" into a false-positive exception).
func TestCrossFile_SymlinkedChildRefused(t *testing.T) {
	dir := writeWorkflows(t, map[string]string{
		"entry.dip":     entryRestrictsRefs("child.dip"),
		"realchild.dip": childZeroIntent,
	})
	if err := os.Symlink(filepath.Join(dir, "realchild.dip"), filepath.Join(dir, "child.dip")); err != nil {
		t.Skip("symlinks not supported on this platform")
	}
	diags, classified := crossDiags(t, dir, "entry.dip")
	if got := countCode(diags, validator.DIP146); got != 0 {
		t.Fatalf("want 0 DIP146 (symlinked child refused), got %d", got)
	}
	for _, p := range classified {
		if p != postureUnresolved {
			t.Errorf("want postureUnresolved (DIP143 retained), got %v", p)
		}
	}
}

// T2 — a child ref that ESCAPES the entry-file containment root is refused
// fail-soft. The entry lives in a subdir so its root is dir/sub; `../escape.dip`
// resolves to dir/escape.dip, outside root but inside the temp tree (so current,
// unhardened code WOULD read it and fire DIP146 — making this red-first).
func TestCrossFile_RootEscapingChildRefused(t *testing.T) {
	dir := writeWorkflows(t, map[string]string{
		"sub/entry.dip": entryRestrictsRefs("../escape.dip"),
		"escape.dip":    childZeroIntent,
	})
	diags, classified := crossDiags(t, dir, "sub/entry.dip")
	if got := countCode(diags, validator.DIP146); got != 0 {
		t.Fatalf("want 0 DIP146 (root-escaping child refused), got %d", got)
	}
	for _, p := range classified {
		if p != postureUnresolved {
			t.Errorf("want postureUnresolved (DIP143 retained), got %v", p)
		}
	}
}

// T6 — a child reached through a SYMLINKED ANCESTOR DIRECTORY is refused
// (the assertNoSymlinkAncestor vector), even though the leaf itself is a real file.
func TestCrossFile_SymlinkedAncestorRefused(t *testing.T) {
	dir := writeWorkflows(t, map[string]string{
		"entry.dip":      entryRestrictsRefs("linkdir/child.dip"),
		"realdir/child.dip": childZeroIntent,
	})
	if err := os.Symlink(filepath.Join(dir, "realdir"), filepath.Join(dir, "linkdir")); err != nil {
		t.Skip("symlinks not supported on this platform")
	}
	diags, classified := crossDiags(t, dir, "entry.dip")
	if got := countCode(diags, validator.DIP146); got != 0 {
		t.Fatalf("want 0 DIP146 (symlinked ancestor refused), got %d", got)
	}
	for _, p := range classified {
		if p != postureUnresolved {
			t.Errorf("want postureUnresolved (DIP143 retained), got %v", p)
		}
	}
}

// T11 — refusal at DEPTH: entry -> child (real, full-restrict) -> grandchild via
// a symlink. The child classifies normally (full-restrict, no DIP146); the
// grandchild boundary is refused (postureUnresolved); recursion continues without
// abort. Distinct code path from T1 (refusal mid-recursion).
func TestCrossFile_SymlinkRefusalAtDepth(t *testing.T) {
	dir := writeWorkflows(t, map[string]string{
		"entry.dip":     entryRestrictsRefs("child.dip"),
		"child.dip":     childWithBoundary("grandlink.dip"),
		"realgrand.dip": childZeroIntent,
	})
	if err := os.Symlink(filepath.Join(dir, "realgrand.dip"), filepath.Join(dir, "grandlink.dip")); err != nil {
		t.Skip("symlinks not supported on this platform")
	}
	diags, classified := crossDiags(t, dir, "entry.dip")
	if got := countCode(diags, validator.DIP146); got != 0 {
		t.Fatalf("want 0 DIP146 (grandchild symlink refused), got %d: %v", got, diags)
	}
	if !hasPosture(classified, postureFullRestrict) {
		t.Errorf("want the child boundary classified full-restrict: %v", classified)
	}
	if !hasPosture(classified, postureUnresolved) {
		t.Errorf("want the grandchild boundary classified unresolved: %v", classified)
	}
}
```

- [ ] **Step 2: Run the new tests to verify they FAIL**

Run: `just test-pkg dippin 2>&1 | grep -E "SymlinkedChildRefused|RootEscapingChildRefused|SymlinkedAncestorRefused|SymlinkRefusalAtDepth|FAIL|ok"`
Expected: all four FAIL — current code follows the symlink/escape and resolves the child, so `countCode(DIP146)` is `1`, not `0` (e.g. "want 0 DIP146 (symlinked child refused), got 1"). If any reports "skipped", symlinks are unsupported here — that's acceptable for the symlink cases, but `RootEscapingChildRefused` (no symlink) MUST fail.

- [ ] **Step 3: Commit the red tests**

```bash
git add cmd/dippin/crossfile_tool_access_test.go
git commit -m "test(cli): failing tests for cross-file symlink/escape refusal (#100)

T1/T6/T11 (symlink leaf/ancestor/depth) and T2 (root escape) — currently RED:
the unhardened resolver follows them and fires DIP146 instead of refusing.

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

## Task 5: Implement the hardening (GREEN)

**Files:**
- Modify: `cmd/dippin/crossfile_tool_access.go` (imports; `crossFileToolAccess`, `walkBoundaries`, `visitBoundary`, `maybeRecurse`, `resolveBoundaryChild`; replace `resolveBoundaryRefPath` with `resolveChildPath`)

Thread a fixed absolute `root` through the walk and route the child read through `ensureUnderRoot` + `dipx.ReadNoFollowSymlinks`.

- [ ] **Step 1: Swap the `os` import for `dipx`**

In the import block, remove `"os"` (its only use, `os.ReadFile`, is being deleted) and add the dipx import. The block becomes:

```go
import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/2389-research/dippin-lang/dipx"
	"github.com/2389-research/dippin-lang/ir"
	"github.com/2389-research/dippin-lang/validator"
)
```

- [ ] **Step 2: Derive the absolute root once and pass it into the walk**

In `crossFileToolAccess`, change the call:

```go
	intentSeen := validator.WorkflowDeclaresToolAccess(entry)
	walkBoundaries(entry, intentSeen, 0, visited, &diags, classified)
```

to:

```go
	intentSeen := validator.WorkflowDeclaresToolAccess(entry)
	// root is the containment anchor for every child ref, captured ONCE from the
	// entry's directory and threaded unchanged (never recomputed per-parent — see
	// spec D2/C2). absOrClean makes it absolute so ensureUnderRoot's filepath.Rel
	// can compare it against absolute child targets.
	root := filepath.Dir(absOrClean(entryPath))
	walkBoundaries(entry, intentSeen, 0, root, visited, &diags, classified)
```

- [ ] **Step 3: Thread `root` through `walkBoundaries`, `visitBoundary`, `maybeRecurse`**

Change `walkBoundaries`'s signature and its `visitBoundary` call:

```go
func walkBoundaries(w *ir.Workflow, intentSeen bool, depth int, root string, visited map[string]bool, diags *[]validator.Diagnostic, classified map[ir.SourceLocation]childPosture) {
	for _, n := range w.Nodes {
		_, ref := boundaryKindRef(n)
		if ref == "" || n.Source.File == "" {
			continue
		}
		visitBoundary(n, ref, intentSeen, depth, root, visited, diags, classified)
	}
}
```

Change `visitBoundary`'s signature and its two callees:

```go
func visitBoundary(n *ir.Node, ref string, intentSeen bool, depth int, root string, visited map[string]bool, diags *[]validator.Diagnostic, classified map[ir.SourceLocation]childPosture) {
	child, childPath := resolveBoundaryChild(n, ref, root)
	if child == nil {
		classified[n.Source] = postureUnresolved
		return
	}
	posture := classifyChild(child)
	classified[n.Source] = posture
	if intentSeen && posture == postureZeroIntent {
		*diags = append(*diags, boundaryDiag(n, ref))
	}
	maybeRecurse(child, childPath, intentSeen, depth, root, visited, diags, classified)
}
```

Change `maybeRecurse`'s signature and its `walkBoundaries` call (root is a pure pass-through here):

```go
func maybeRecurse(child *ir.Workflow, childPath string, intentSeen bool, depth int, root string, visited map[string]bool, diags *[]validator.Diagnostic, classified map[ir.SourceLocation]childPosture) {
	key := canonicalKey(childPath)
	if key == "" || visited[key] || depth+1 > crossFileMaxDepth {
		return
	}
	visited[key] = true
	childIntent := intentSeen || validator.WorkflowDeclaresToolAccess(child)
	walkBoundaries(child, childIntent, depth+1, root, visited, diags, classified)
}
```

- [ ] **Step 4: Replace `resolveBoundaryChild` and `resolveBoundaryRefPath` with the hardened pair**

Replace both `resolveBoundaryChild` (`:89-103`) and `resolveBoundaryRefPath` (`:105-112`) with:

```go
// resolveBoundaryChild resolves ref relative to the boundary node's source file,
// refusing root-escaping or symlinked targets, then parses the child workflow.
// Fail-soft: ANY error (escape, symlink, read, parse) yields (nil, ""), routing
// the boundary to postureUnresolved so the per-file DIP143 advisory is retained.
// This is detection parity with the pack walker (Lstat-based, not the parser's
// stronger O_NOFOLLOW open-once — the residual leaf TOCTOU matches pack's; a lint
// emits only DIP codes, never file content). Callers (walkBoundaries) guarantee
// ref != "" and n.Source.File != "".
func resolveBoundaryChild(n *ir.Node, ref, root string) (*ir.Workflow, string) {
	path, err := resolveChildPath(n.Source.File, ref, root)
	if err != nil {
		return nil, ""
	}
	data, err := dipx.ReadNoFollowSymlinks(path, root)
	if err != nil {
		return nil, ""
	}
	w, err := parseAndResolveDip(data, path)
	if err != nil {
		return nil, ""
	}
	return w, path
}

// resolveChildPath joins ref against parentFile's directory and refuses any result
// that escapes root. Pack-identical (mirrors dipx's resolveRefOnDisk): filepath.Join
// swallows the leading slash of an absolute ref, so abs refs are RE-ROOTED under
// root, not read out-of-root. The filepath.Abs is load-bearing — at the entry level
// parentFile (n.Source.File) may be relative while root is absolute, and
// ensureUnderRoot's filepath.Rel needs both absolute. Do not simplify it away.
func resolveChildPath(parentFile, ref, root string) (string, error) {
	target := filepath.Clean(filepath.Join(filepath.Dir(parentFile), ref))
	abs, err := filepath.Abs(target)
	if err != nil {
		return "", err
	}
	if err := ensureUnderRoot(abs, root); err != nil {
		return "", err
	}
	return abs, nil
}
```

- [ ] **Step 5: Run the Task 4 refusal tests — now GREEN**

Run: `just test-pkg dippin 2>&1 | grep -E "SymlinkedChildRefused|RootEscapingChildRefused|SymlinkedAncestorRefused|SymlinkRefusalAtDepth|FAIL|ok"`
Expected: all four PASS (or symlink cases SKIP where unsupported; `RootEscapingChildRefused` PASSES). No FAIL. The full `dippin` package (existing cross-file tests included) is `ok` — they use plain in-root sibling/subdir refs and are unaffected.

- [ ] **Step 6: Confirm no leftover references to the deleted function / import**

Run: `grep -n "resolveBoundaryRefPath\|os\\.ReadFile" cmd/dippin/crossfile_tool_access.go`
Expected: no output (function deleted, `os.ReadFile` gone).

- [ ] **Step 7: Check complexity stays within budget**

Run: `just complexity`
Expected: no violations. (`resolveBoundaryChild` cyclo 4, `resolveChildPath` cyclo 3, threaded functions gain no branch — all ≤ 5.)

- [ ] **Step 8: Commit**

```bash
git add cmd/dippin/crossfile_tool_access.go
git commit -m "feat(cli): refuse symlinked/root-escaping cross-file children (#100)

Thread a fixed absolute containment root (the entry file's directory) through
the DIP146 walk and route each child read through ensureUnderRoot (lexical
..-escape refusal) + dipx.ReadNoFollowSymlinks (leaf+ancestor symlink refusal),
replacing the unhardened os.ReadFile. Fail-soft: any refusal -> postureUnresolved
-> DIP143 retained, no DIP146, no abort. Deletes the absolute-ref special-case
(re-rooted, pack-identical). Detection parity with the pack walker; no tracker
dependency.

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

## Task 6: Add the regression / coverage guards (GREEN)

**Files:**
- Modify: `cmd/dippin/crossfile_tool_access_test.go` (add T3, T4, T5, T7, T8, T9)

These validate that correct behavior is preserved (legit children still resolve; relative entry works; abs refs re-root; symlink cycle terminates) and guard the lynchpin (absolute root, fixed not per-parent). They pass on the Task 5 implementation; T3 and T7 would fail under specific implementation bugs.

- [ ] **Step 1: Add the six tests**

Append to `cmd/dippin/crossfile_tool_access_test.go`:

```go
// T3 — HIGHEST-VALUE GUARD: a subdir child whose ref climbs back UP into the entry
// root must still resolve. entry -> sub/child.dip (full-restrict) -> ../other.dip
// (zero-intent, == root/other.dip, in-root). Correct (entry-anchored fixed root):
// other resolves, DIP146 fires on the child->other edge => count 1. A BUGGY
// per-parent-root recompute (root = dir(sub/child.dip) = root/sub) would refuse
// ../other.dip (escapes root/sub) => count 0. So this fails under that bug.
func TestCrossFile_SubdirChildClimbsBackIntoRoot(t *testing.T) {
	dir := writeWorkflows(t, map[string]string{
		"entry.dip":     entryRestrictsRefs("sub/child.dip"),
		"sub/child.dip": childWithBoundary("../other.dip"),
		"other.dip":     childZeroIntent,
	})
	diags, _ := crossDiags(t, dir, "entry.dip")
	if got := countCode(diags, validator.DIP146); got != 1 {
		t.Fatalf("want 1 DIP146 on the climb-back ../other.dip edge, got %d: %v", got, diags)
	}
}

// T4 — legit sibling AND subdirectory children resolve normally; DIP146
// supersession works exactly as before the hardening (regression guard).
func TestCrossFile_LegitSiblingAndSubdirResolve(t *testing.T) {
	sibling := writeWorkflows(t, map[string]string{
		"entry.dip": entryRestrictsRefs("child.dip"),
		"child.dip": childZeroIntent,
	})
	if got := countCode(firstDiags(t, sibling, "entry.dip"), validator.DIP146); got != 1 {
		t.Fatalf("sibling: want 1 DIP146, got %d", got)
	}
	subdir := writeWorkflows(t, map[string]string{
		"entry.dip":     entryRestrictsRefs("sub/child.dip"),
		"sub/child.dip": childZeroIntent,
	})
	if got := countCode(firstDiags(t, subdir, "entry.dip"), validator.DIP146); got != 1 {
		t.Fatalf("subdir: want 1 DIP146, got %d", got)
	}
}

// T5 — policy pin: a symlink whose target is a perfectly legitimate in-root file
// is STILL refused (all symlinks refused unconditionally, matching pack). Pins the
// conservative policy so it is not later "fixed" into a false-positive exception.
func TestCrossFile_BenignInRootSymlinkRefused(t *testing.T) {
	dir := writeWorkflows(t, map[string]string{
		"entry.dip":  entryRestrictsRefs("child.dip"),
		"target.dip": childZeroIntent,
	})
	if err := os.Symlink(filepath.Join(dir, "target.dip"), filepath.Join(dir, "child.dip")); err != nil {
		t.Skip("symlinks not supported on this platform")
	}
	diags, classified := crossDiags(t, dir, "entry.dip")
	if got := countCode(diags, validator.DIP146); got != 0 {
		t.Fatalf("want 0 DIP146 (benign in-root symlink still refused), got %d", got)
	}
	if !hasPosture(classified, postureUnresolved) {
		t.Errorf("want postureUnresolved, got %v", classified)
	}
}

// T7 — lynchpin guard (spec D2): a RELATIVE entry path must still resolve legit
// children. crossDiags always builds an absolute entry path, so this drives the
// pass with a relative path via os.Chdir. If root were left relative (no absOrClean),
// ensureUnderRoot's filepath.Rel(relRoot, absChild) would error and refuse every
// child => count 0. Not parallel-safe (mutates cwd); the file uses no t.Parallel.
func TestCrossFile_RelativeEntryResolves(t *testing.T) {
	dir := writeWorkflows(t, map[string]string{
		"entry.dip": entryRestrictsRefs("child.dip"),
		"child.dip": childZeroIntent,
	})
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(cwd) })
	w, err := loadWorkflow("entry.dip")
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	diags, _ := crossFileToolAccess(w, "entry.dip")
	if got := countCode(diags, validator.DIP146); got != 1 {
		t.Fatalf("want 1 DIP146 with a relative entry path, got %d", got)
	}
}

// T8 — an absolute ref is re-rooted under the parent dir (filepath.Join swallows
// the leading slash), not read out-of-root. /etc/passwd re-roots to <root>/etc/passwd
// which does not exist => postureUnresolved (DIP143 retained). Documents re-rooting;
// it does NOT assert literal absolute-path refusal.
func TestCrossFile_AbsoluteRefReRooted(t *testing.T) {
	dir := writeWorkflows(t, map[string]string{
		"entry.dip": entryRestrictsRefs("/etc/passwd"),
	})
	diags, classified := crossDiags(t, dir, "entry.dip")
	if got := countCode(diags, validator.DIP146); got != 0 {
		t.Fatalf("want 0 DIP146 (abs ref re-rooted, target absent), got %d", got)
	}
	for _, p := range classified {
		if p != postureUnresolved {
			t.Errorf("want postureUnresolved, got %v", p)
		}
	}
}

// T9 — a symlink that would form a cycle is refused at the first hop (so the cycle
// is never entered) and the walk terminates. Termination of NON-symlink cycles is
// covered by TestCrossFile_CycleTerminates / TestCrossFile_SelfReferenceTerminates;
// this only asserts the refusal short-circuit does not hang.
func TestCrossFile_SymlinkCycleRefused(t *testing.T) {
	a := childWithBoundary("blink.dip") // a.dip -> blink.dip
	dir := writeWorkflows(t, map[string]string{
		"a.dip": a,
	})
	// blink.dip is a symlink back to a.dip (a would-be a->blink->a cycle).
	if err := os.Symlink(filepath.Join(dir, "a.dip"), filepath.Join(dir, "blink.dip")); err != nil {
		t.Skip("symlinks not supported on this platform")
	}
	diags, classified := crossDiags(t, dir, "a.dip")
	if got := countCode(diags, validator.DIP146); got != 0 {
		t.Fatalf("want 0 DIP146 (symlink hop refused), got %d", got)
	}
	if !hasPosture(classified, postureUnresolved) {
		t.Errorf("want the symlink boundary classified unresolved: %v", classified)
	}
}

// firstDiags runs the pass and returns only the diagnostics (helper for cases that
// don't inspect the classified map).
func firstDiags(t *testing.T, dir, entry string) []validator.Diagnostic {
	t.Helper()
	diags, _ := crossDiags(t, dir, entry)
	return diags
}
```

- [ ] **Step 2: Run the new guards**

Run: `just test-pkg dippin 2>&1 | grep -E "SubdirChildClimbsBackIntoRoot|LegitSiblingAndSubdirResolve|BenignInRootSymlinkRefused|RelativeEntryResolves|AbsoluteRefReRooted|SymlinkCycleRefused|FAIL|ok"`
Expected: all PASS (symlink cases SKIP where unsupported). No FAIL.

- [ ] **Step 3: Run the full dippin package once more**

Run: `just test-pkg dippin`
Expected: PASS — all existing + new cross-file tests green.

- [ ] **Step 4: Commit**

```bash
git add cmd/dippin/crossfile_tool_access_test.go
git commit -m "test(cli): regression/lynchpin guards for #100 hardening

T3 (subdir climbs back into root — guards per-parent-root recompute), T4 (legit
sibling/subdir still resolve), T5 (benign in-root symlink still refused), T7
(relative entry — guards absolute-root lynchpin), T8 (abs ref re-rooted), T9
(symlink cycle refused/terminates).

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

## Task 7: Doc note + full verification

**Files:**
- Modify: `docs/architecture.md` (the `helpers.go` responsibility line)

- [ ] **Step 1: Add the responsibility note**

Open `docs/architecture.md`, find the line describing `helpers.go` (around `:148`):

```text
│   ├── helpers.go      # Hash verify, parse-and-link, walkSourceTree
```

Append the shared read primitive to the responsibility list, e.g.:

```text
│   ├── helpers.go      # Hash verify, parse-and-link, walkSourceTree, ReadNoFollowSymlinks (shared no-follow read)
```

(Match the existing comment style/width; this is a responsibility note, not an exhaustive export list.)

- [ ] **Step 2: Run the full race suite**

Run: `just test-race`
Expected: PASS across all packages.

- [ ] **Step 3: Confirm the example suites do not flip**

Run: `just validate-examples && just lint-examples`
Expected: both succeed; no example changes its diagnostics (all example subgraph/manager_loop refs are plain in-root relative paths — none are symlinked or escaping).

- [ ] **Step 4: Confirm the wasm build stays green**

Run: `just wasm`
Expected: build OK (dipx is not in the wasm closure; the lint pass is `cmd/dippin`-only; `ReadNoFollowSymlinks` uses no `syscall`).

- [ ] **Step 5: Commit the doc note**

```bash
git add docs/architecture.md
git commit -m "docs: note ReadNoFollowSymlinks as a shared dipx responsibility (#100)

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

- [ ] **Step 6: Final full-suite gate (pre-commit parity)**

Run: `just build && just vet && just test && just complexity && just validate-examples`
Expected: all green. (`just check` fails locally on tree-sitter-generate — that's the known environment gotcha; the pre-commit hook on the next commit is the real CI gate and runs build/vet/golangci-lint/gofmt/tests/releasecheck/complexity/validate-examples.)

---

## Self-Review notes (traceability to spec)

- **D1 (sharing):** Task 1 (export `ReadNoFollowSymlinks`), Task 3 (generalize `ensureUnderRoot`), Task 5 Step 4 (compose both). No combined function; no pack-path (`resolveRefOnDisk`) edit.
- **D2 (fixed absolute root):** Task 5 Step 2 (`root := filepath.Dir(absOrClean(entryPath))`, threaded by value). Guarded by T7 (relative entry) and T3 (no per-parent recompute).
- **D3 (resolution flow, no IsAbs):** Task 5 Step 4 (`resolveChildPath`, `resolveBoundaryRefPath` deleted). Abs re-rooting guarded by T8.
- **D4 (fail-soft):** Task 5 Step 4 (all errors → `(nil, "")`); T1/T2/T6/T11 assert `postureUnresolved` + 0 DIP146 + no abort.
- **D5 (cycle):** T9 (symlink hop refused, terminates); existing literal-ref cycle tests unchanged.
- **Scope — no new DIP, no build tags:** confirmed (no `validator` code edits; no `//go:build`). File directives untouched (out of scope by pack parity).
- **Tests T1–T11:** T1=SymlinkedChildRefused (absorbs T5 policy via comment; T5 also kept explicit as BenignInRootSymlinkRefused), T2=RootEscapingChildRefused, T3=SubdirChildClimbsBackIntoRoot, T4=LegitSiblingAndSubdirResolve, T5=BenignInRootSymlinkRefused, T6=SymlinkedAncestorRefused, T7=RelativeEntryResolves, T8=AbsoluteRefReRooted, T9=SymlinkCycleRefused, T10=TestReadNoFollowSymlinks (dipx), T11=SymlinkRefusalAtDepth.
- **Verification:** Task 7 (test-race, complexity, wasm, validate/lint-examples).
