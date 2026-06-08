package main

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/2389-research/dippin-lang/ir"
	"github.com/2389-research/dippin-lang/validator"
)

// --- fixtures: real .dip files parsed by the real parser ---

// entryRestrictsRefs is an entry that locks down agent Lock and delegates to a
// child via manager_loop. childRef is the child file name in the same dir.
func entryRestrictsRefs(childRef string) string {
	return `workflow Entry
  start: Lock
  exit: Sup

  agent Lock
    prompt: "x"
    tool_access: none

  manager_loop Sup
    subgraph_ref: ` + childRef + `
    max_cycles: 3

  edges
    Lock -> Sup
`
}

// entryNoIntentRefs delegates to a child but declares no tool_access itself.
func entryNoIntentRefs(childRef string) string {
	return `workflow Entry
  start: Go
  exit: Sup

  agent Go
    prompt: "x"

  manager_loop Sup
    subgraph_ref: ` + childRef + `
    max_cycles: 3

  edges
    Go -> Sup
`
}

// Children mirror the proven minimalDip shape (human start -> agent, with edges)
// so the parser accepts them; only the agents' tool_access differs.
const childZeroIntent = `workflow Child
  start: Ask
  exit: Done

  human Ask
    mode: freeform

  agent Done
    prompt: "do it"

  edges
    Ask -> Done
`

const childFullRestrict = `workflow Child
  start: Ask
  exit: Done

  human Ask
    mode: freeform

  agent Done
    prompt: "do it"
    tool_access: none

  edges
    Ask -> Done
`

const childPartial = `workflow Child
  start: Ask
  exit: B

  human Ask
    mode: freeform

  agent A
    prompt: "a"
    tool_access: none

  agent B
    prompt: "b"

  edges
    Ask -> A
    A -> B
`

const childAgentless = `workflow Child
  start: Ask
  exit: Done

  human Ask
    mode: freeform

  human Done
    mode: freeform

  edges
    Ask -> Done
`

// writeWorkflows writes name->content into a fresh temp dir and returns the dir.
func writeWorkflows(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	for name, content := range files {
		p := filepath.Join(dir, name)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", p, err)
		}
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			t.Fatalf("write %s: %v", p, err)
		}
	}
	return dir
}

// crossDiags loads dir/entry through the real loader and runs the pass.
func crossDiags(t *testing.T, dir, entry string) ([]validator.Diagnostic, map[ir.SourceLocation]childPosture) {
	t.Helper()
	entryPath := filepath.Join(dir, entry)
	w, err := loadWorkflow(entryPath)
	if err != nil {
		t.Fatalf("load %s: %v", entry, err)
	}
	return crossFileToolAccess(w, entryPath)
}

func countCode(diags []validator.Diagnostic, code string) int {
	n := 0
	for _, d := range diags {
		if d.Code == code {
			n++
		}
	}
	return n
}

func TestCrossFile_FiresOnZeroIntentChild(t *testing.T) {
	dir := writeWorkflows(t, map[string]string{
		"entry.dip": entryRestrictsRefs("child.dip"),
		"child.dip": childZeroIntent,
	})
	diags, classified := crossDiags(t, dir, "entry.dip")
	if got := countCode(diags, validator.DIP146); got != 1 {
		t.Fatalf("want 1 DIP146, got %d: %v", got, diags)
	}
	found := false
	for _, p := range classified {
		if p == postureZeroIntent {
			found = true
		}
	}
	if !found {
		t.Errorf("expected a zero-intent classification: %v", classified)
	}
}

func TestCrossFile_SilentOnFullRestrict(t *testing.T) {
	dir := writeWorkflows(t, map[string]string{
		"entry.dip": entryRestrictsRefs("child.dip"),
		"child.dip": childFullRestrict,
	})
	diags, _ := crossDiags(t, dir, "entry.dip")
	if got := countCode(diags, validator.DIP146); got != 0 {
		t.Fatalf("want 0 DIP146 (child fully restricted), got %d", got)
	}
}

func TestCrossFile_SilentOnPartialAudit(t *testing.T) {
	dir := writeWorkflows(t, map[string]string{
		"entry.dip": entryRestrictsRefs("child.dip"),
		"child.dip": childPartial,
	})
	diags, classified := crossDiags(t, dir, "entry.dip")
	if got := countCode(diags, validator.DIP146); got != 0 {
		t.Fatalf("want 0 DIP146 (partial audit keeps DIP143), got %d", got)
	}
	found := false
	for _, p := range classified {
		if p == posturePartial {
			found = true
		}
	}
	if !found {
		t.Errorf("expected a partial classification: %v", classified)
	}
}

func TestCrossFile_SilentWhenNoPathIntent(t *testing.T) {
	dir := writeWorkflows(t, map[string]string{
		"entry.dip": entryNoIntentRefs("child.dip"),
		"child.dip": childZeroIntent,
	})
	diags, _ := crossDiags(t, dir, "entry.dip")
	if got := countCode(diags, validator.DIP146); got != 0 {
		t.Fatalf("want 0 DIP146 (no intent on path), got %d", got)
	}
}

func TestCrossFile_AncestorPathIntent(t *testing.T) {
	// entry (no intent) -> mid (restricts) -> leaf (zero intent) => DIP146 on mid->leaf.
	mid := `workflow Mid
  start: Lock
  exit: Sup

  agent Lock
    prompt: "x"
    tool_access: none

  manager_loop Sup
    subgraph_ref: leaf.dip
    max_cycles: 3

  edges
    Lock -> Sup
`
	dir := writeWorkflows(t, map[string]string{
		"entry.dip": entryNoIntentRefs("mid.dip"),
		"mid.dip":   mid,
		"leaf.dip":  childZeroIntent,
	})
	diags, _ := crossDiags(t, dir, "entry.dip")
	if got := countCode(diags, validator.DIP146); got != 1 {
		t.Fatalf("want 1 DIP146 (ancestor-path intent), got %d: %v", got, diags)
	}
}

func TestCrossFile_Transitive(t *testing.T) {
	// entry (restricts) -> child (restricts) -> grandchild (zero) => DIP146 on grandchild edge.
	child := `workflow Child
  start: Lock
  exit: Sup

  agent Lock
    prompt: "x"
    tool_access: none

  manager_loop Sup
    subgraph_ref: grand.dip
    max_cycles: 3

  edges
    Lock -> Sup
`
	dir := writeWorkflows(t, map[string]string{
		"entry.dip": entryRestrictsRefs("child.dip"),
		"child.dip": child,
		"grand.dip": childZeroIntent,
	})
	diags, _ := crossDiags(t, dir, "entry.dip")
	if got := countCode(diags, validator.DIP146); got != 1 {
		t.Fatalf("want 1 DIP146 on the grandchild edge, got %d: %v", got, diags)
	}
}

func TestCrossFile_DiamondTwoFindings(t *testing.T) {
	// entry has no intent; P1 and P2 each restrict tools and both reference the same
	// zero-intent child. The path intent comes from P1/P2 (OR-threaded down the DFS),
	// so each P->child boundary fires its own DIP146 => 2 findings (one per edge).
	parent := func(name, ref string) string {
		return `workflow ` + name + `
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
	entry := `workflow Entry
  start: P1
  exit: P2

  subgraph P1
    ref: p1.dip

  subgraph P2
    ref: p2.dip

  edges
    P1 -> P2
`
	dir := writeWorkflows(t, map[string]string{
		"entry.dip": entry,
		"p1.dip":    parent("P1", "child.dip"),
		"p2.dip":    parent("P2", "child.dip"),
		"child.dip": childZeroIntent,
	})
	diags, _ := crossDiags(t, dir, "entry.dip")
	if got := countCode(diags, validator.DIP146); got != 2 {
		t.Fatalf("want 2 DIP146 (one per parent->child edge), got %d: %v", got, diags)
	}
}

func TestCrossFile_CycleTerminates(t *testing.T) {
	// A -> B -> A. Must terminate; A has intent so no DIP146 on the B->A edge.
	a := `workflow A
  start: Lock
  exit: Sup

  agent Lock
    prompt: "x"
    tool_access: none

  manager_loop Sup
    subgraph_ref: b.dip
    max_cycles: 3

  edges
    Lock -> Sup
`
	b := `workflow B
  start: Go
  exit: Sup

  human Go
    mode: freeform

  manager_loop Sup
    subgraph_ref: a.dip
    max_cycles: 3

  edges
    Go -> Sup
`
	dir := writeWorkflows(t, map[string]string{
		"a.dip": a,
		"b.dip": b,
	})
	// Completing at all proves termination (no hang/stack overflow). A restricts
	// tools, so on the B->A edge A classifies as full-restrict — no DIP146.
	diags, _ := crossDiags(t, dir, "a.dip")
	if got := countCode(diags, validator.DIP146); got != 0 {
		t.Fatalf("want 0 DIP146 on cycle (A is full-restrict), got %d", got)
	}
}

func TestCrossFile_SelfReferenceTerminates(t *testing.T) {
	// A -> A. Must terminate; A is agent-less here so no DIP146 either way.
	a := `workflow A
  start: Go
  exit: Sup

  human Go
    mode: freeform

  manager_loop Sup
    subgraph_ref: a.dip
    max_cycles: 3

  edges
    Go -> Sup
`
	dir := writeWorkflows(t, map[string]string{"a.dip": a})
	diags, _ := crossDiags(t, dir, "a.dip")
	if got := countCode(diags, validator.DIP146); got != 0 {
		t.Fatalf("want 0 DIP146 on self-reference, got %d", got)
	}
}

func TestCrossFile_DepthCapStopsTraversal(t *testing.T) {
	// A FULLY-RESOLVABLE chain longer than crossFileMaxDepth, ending in a real leaf
	// with no boundary. The visited set alone would let every distinct file through,
	// and nothing is unresolvable — so the ONLY thing that can stop traversal short
	// of the leaf is the depth cap, which bounds classified to ~crossFileMaxDepth.
	const chain = crossFileMaxDepth + 10
	files := map[string]string{
		"entry.dip": `workflow Entry
  start: Lock
  exit: Sup

  agent Lock
    prompt: "x"
    tool_access: none

  manager_loop Sup
    subgraph_ref: f1.dip
    max_cycles: 3

  edges
    Lock -> Sup
`,
	}
	for i := 1; i < chain; i++ {
		next := fmt.Sprintf("f%d.dip", i+1)
		files[fmt.Sprintf("f%d.dip", i)] = `workflow F
  start: Go
  exit: Sup

  human Go
    mode: freeform

  manager_loop Sup
    subgraph_ref: ` + next + `
    max_cycles: 3

  edges
    Go -> Sup
`
	}
	// Resolvable terminal with no boundary: without the cap, traversal would reach
	// it and classify the full chain; with the cap it never gets here.
	files[fmt.Sprintf("f%d.dip", chain)] = childAgentless
	dir := writeWorkflows(t, files)
	_, classified := crossDiags(t, dir, "entry.dip")
	if n := len(classified); n == 0 || n > crossFileMaxDepth+2 {
		t.Fatalf("depth cap did not bound traversal: classified %d (cap %d, chain %d)", n, crossFileMaxDepth, chain)
	}
}

func TestCrossFile_UnresolvableChildIsSilent(t *testing.T) {
	dir := writeWorkflows(t, map[string]string{
		"entry.dip": entryRestrictsRefs("missing.dip"), // no missing.dip on disk
	})
	diags, classified := crossDiags(t, dir, "entry.dip")
	if got := countCode(diags, validator.DIP146); got != 0 {
		t.Fatalf("want 0 DIP146 for unresolvable child, got %d", got)
	}
	for _, p := range classified {
		if p != postureUnresolved {
			t.Errorf("want postureUnresolved, got %v", p)
		}
	}
}

func TestCrossFile_AgentlessChildIsSilent(t *testing.T) {
	dir := writeWorkflows(t, map[string]string{
		"entry.dip": entryRestrictsRefs("child.dip"),
		"child.dip": childAgentless,
	})
	diags, _ := crossDiags(t, dir, "entry.dip")
	if got := countCode(diags, validator.DIP146); got != 0 {
		t.Fatalf("want 0 DIP146 (agent-less child has no tools to grant), got %d", got)
	}
}

// Known intentional false positive: a fully-tooled worker pool fires DIP146.
// This is mitigated by Hint severity, NOT by the gate. Do not "fix" this into a
// Warning or add a heuristic — see the design spec (#89).
func TestCrossFile_IntentionalOpenWorkerFires(t *testing.T) {
	dir := writeWorkflows(t, map[string]string{
		"entry.dip":  entryRestrictsRefs("worker.dip"),
		"worker.dip": childZeroIntent,
	})
	diags, _ := crossDiags(t, dir, "entry.dip")
	if got := countCode(diags, validator.DIP146); got != 1 {
		t.Fatalf("want 1 DIP146 (known intentional-open FP, Hint-mitigated), got %d", got)
	}
}

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
// no DIP146, no error. Also pins the policy that a benign in-root symlink target
// is still refused unconditionally (do not "fix" into a false-positive exception).
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
		"entry.dip":         entryRestrictsRefs("linkdir/child.dip"),
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

// --- #102: deep (depth>=1) partial-audit / unresolvable advisory (reuses DIP143) ---

// TestCrossFile_DeepPartialChildFiresDIP143 is the core #102 case: entry (restricts)
// -> child (full-restrict, so entry->child is DIP146-silent) -> grandchild (PARTIAL:
// restricts agent A, leaves tool-bearing B open). The grandchild's partial-audit gap
// is behind an already-audited intermediate, so validator.Lint (entry-only) never sees
// it. The cross pass must emit its own DIP143 (Hint) at the child->grandchild boundary.
func TestCrossFile_DeepPartialChildFiresDIP143(t *testing.T) {
	dir := writeWorkflows(t, map[string]string{
		"entry.dip": entryRestrictsRefs("child.dip"),
		"child.dip": childWithBoundary("grand.dip"),
		"grand.dip": childPartial,
	})
	diags, _ := crossDiags(t, dir, "entry.dip")
	if got := countCode(diags, validator.DIP143); got != 1 {
		t.Fatalf("want 1 deep DIP143 on the child->grandchild boundary, got %d: %v", got, diags)
	}
	if got := countCode(diags, validator.DIP146); got != 0 {
		t.Fatalf("want 0 DIP146 (grandchild is partial, not zero-intent), got %d", got)
	}
	for _, d := range diags {
		if d.Code == validator.DIP143 && d.Severity != validator.SeverityHint {
			t.Errorf("deep DIP143 must be Hint, got %v", d.Severity)
		}
	}
}

// TestCrossFile_DeepUnresolvableChildFiresDIP143: entry (restricts) -> child
// (full-restrict) -> grandchild that does not exist on disk. The grandchild boundary
// classifies as postureUnresolved at depth 1; the cross pass emits a deep DIP143.
func TestCrossFile_DeepUnresolvableChildFiresDIP143(t *testing.T) {
	dir := writeWorkflows(t, map[string]string{
		"entry.dip": entryRestrictsRefs("child.dip"),
		"child.dip": childWithBoundary("missing.dip"), // no missing.dip on disk
	})
	diags, classified := crossDiags(t, dir, "entry.dip")
	if got := countCode(diags, validator.DIP143); got != 1 {
		t.Fatalf("want 1 deep DIP143 on the unresolvable grandchild boundary, got %d: %v", got, diags)
	}
	if got := countCode(diags, validator.DIP146); got != 0 {
		t.Fatalf("want 0 DIP146 (unresolvable child), got %d", got)
	}
	if !hasPosture(classified, postureUnresolved) {
		t.Errorf("want the grandchild boundary classified unresolved: %v", classified)
	}
}

// TestCrossFile_EntryPartialNoDoubleEmit: an ENTRY-level (depth 0) partial child must
// yield EXACTLY ONE DIP143 end-to-end — the entry's, from validator.Lint — not two.
// The cross pass must not also emit a deep advisory for the depth-0 boundary (the
// depth>=1 guard). This pins the no-double-flag invariant through the real composition.
func TestCrossFile_EntryPartialNoDoubleEmit(t *testing.T) {
	dir := writeWorkflows(t, map[string]string{
		"entry.dip": entryRestrictsRefs("child.dip"),
		"child.dip": childPartial,
	})
	entryPath := filepath.Join(dir, "entry.dip")
	w, err := loadWorkflow(entryPath)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	base := validator.Lint(w).Diagnostics
	if got := countCode(base, validator.DIP143); got != 1 {
		t.Fatalf("precondition: want 1 entry DIP143 from validator.Lint, got %d", got)
	}
	final := applyCrossFileToolAccess(base, w, entryPath)
	if got := countCode(final, validator.DIP143); got != 1 {
		t.Fatalf("want exactly 1 DIP143 (entry's; no deep double-emit at depth 0), got %d: %v", got, final)
	}
}

// TestCrossFile_DeepPartialSilentWithoutIntent: when NO workflow on the path restricts
// tools (intentSeen false), a deep partial child gets no advisory — there is no
// restriction to escape. Matches DIP146's gate.
func TestCrossFile_DeepPartialSilentWithoutIntent(t *testing.T) {
	// entry (no intent) -> mid (no intent) -> grandchild (partial).
	midNoIntent := `workflow Mid
  start: Go
  exit: Sup

  human Go
    mode: freeform

  manager_loop Sup
    subgraph_ref: grand.dip
    max_cycles: 3

  edges
    Go -> Sup
`
	dir := writeWorkflows(t, map[string]string{
		"entry.dip": entryNoIntentRefs("mid.dip"),
		"mid.dip":   midNoIntent,
		"grand.dip": childPartial,
	})
	diags, _ := crossDiags(t, dir, "entry.dip")
	if got := countCode(diags, validator.DIP143); got != 0 {
		t.Fatalf("want 0 deep DIP143 (no intent on path), got %d: %v", got, diags)
	}
	if got := countCode(diags, validator.DIP146); got != 0 {
		t.Fatalf("want 0 DIP146 (no intent on path), got %d", got)
	}
}

// TestCrossFile_DeepZeroIntentStillDIP146: a zero-intent deep child still fires DIP146
// (the precise gap), NOT the new partial/unresolved DIP143 — unchanged by #102.
func TestCrossFile_DeepZeroIntentStillDIP146(t *testing.T) {
	dir := writeWorkflows(t, map[string]string{
		"entry.dip": entryRestrictsRefs("child.dip"),
		"child.dip": childWithBoundary("grand.dip"),
		"grand.dip": childZeroIntent,
	})
	diags, _ := crossDiags(t, dir, "entry.dip")
	if got := countCode(diags, validator.DIP146); got != 1 {
		t.Fatalf("want 1 DIP146 on the deep zero-intent boundary, got %d: %v", got, diags)
	}
	if got := countCode(diags, validator.DIP143); got != 0 {
		t.Fatalf("want 0 DIP143 (zero-intent is DIP146's job, not the new advisory), got %d", got)
	}
}

// TestCrossFile_DeepPartialAtDepth2 pins that the advisory is gated on depth >= 1,
// not "exactly 1": a partial child three hops down (entry -> child -> grand -> great,
// the great boundary living at depth 2) still fires exactly one deep DIP143. Guards
// against a future refactor mis-threading depth past the first hop.
func TestCrossFile_DeepPartialAtDepth2(t *testing.T) {
	dir := writeWorkflows(t, map[string]string{
		"entry.dip": entryRestrictsRefs("child.dip"),
		"child.dip": childWithBoundary("grand.dip"), // full-restrict, depth-0 boundary
		"grand.dip": childWithBoundary("great.dip"), // full-restrict, depth-1 boundary
		"great.dip": childPartial,                   // partial, classified at the depth-2 boundary
	})
	diags, _ := crossDiags(t, dir, "entry.dip")
	if got := countCode(diags, validator.DIP143); got != 1 {
		t.Fatalf("want 1 deep DIP143 at the depth-2 (grand->great) boundary, got %d: %v", got, diags)
	}
	if got := countCode(diags, validator.DIP146); got != 0 {
		t.Fatalf("want 0 DIP146, got %d", got)
	}
}

// TestCrossFile_DiamondOntoPartialChild pins the per-boundary cardinality for the new
// code path: two restricting parents (P1, P2) each delegate to the SAME partial child,
// so each P->child boundary (depth 1) fires its own deep DIP143 => 2 findings. The
// visited set gates recursion into the shared child, not emission — mirroring the
// DIP146 diamond (TestCrossFile_DiamondTwoFindings) for the partial/DIP143 case.
func TestCrossFile_DiamondOntoPartialChild(t *testing.T) {
	parent := func(name, ref string) string {
		return `workflow ` + name + `
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
	entry := `workflow Entry
  start: P1
  exit: P2

  subgraph P1
    ref: p1.dip

  subgraph P2
    ref: p2.dip

  edges
    P1 -> P2
`
	dir := writeWorkflows(t, map[string]string{
		"entry.dip":  entry,
		"p1.dip":     parent("P1", "shared.dip"),
		"p2.dip":     parent("P2", "shared.dip"),
		"shared.dip": childPartial,
	})
	diags, _ := crossDiags(t, dir, "entry.dip")
	if got := countCode(diags, validator.DIP143); got != 2 {
		t.Fatalf("want 2 deep DIP143 (one per parent->shared boundary), got %d: %v", got, diags)
	}
	if got := countCode(diags, validator.DIP146); got != 0 {
		t.Fatalf("want 0 DIP146 (shared child is partial), got %d", got)
	}
}

// firstDiags runs the pass and returns only the diagnostics (helper for cases that
// don't inspect the classified map).
func firstDiags(t *testing.T, dir, entry string) []validator.Diagnostic {
	t.Helper()
	diags, _ := crossDiags(t, dir, entry)
	return diags
}
