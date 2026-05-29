# DOT round-trip for block-form parallel branches (#76) — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make block-form `parallel` nodes round-trip per-branch model/provider/fidelity (and their fan-out targets) losslessly through `dippin export` → DOT → `migrate`.

**Architecture:** Add a new `branches=` DOT node attribute encoding each branch as percent-encoded `k=v` tokens (`;`-joined within a branch, `,`-joined across branches), mirroring the existing `steer_context` encoder. Export always co-emits `targets=`. Migrate parses `branches=` back into `[]ir.BranchConfig` and treats branch order as the source of truth for `Targets`. Two correctness fixes: stop `inferParallelTargets` from clobbering `Branches`, and make the parity check compare `Branches`.

**Tech Stack:** Go. Packages `export`, `migrate`, `ir`. Spec: `docs/superpowers/specs/2026-05-28-issue-76-dot-parallel-branches-design.md`.

**Conventions (CLAUDE.md):**
- All ops via `just`. Never run raw `go build/test`.
- Complexity caps (pre-commit + CI): cyclomatic ≤ 5, cognitive ≤ 7 per function. No `//nolint`.
- TDD: write the failing test, watch it fail, then implement.
- Commit trailer: `Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>`
- Leave `.claude/settings.local.json` uncommitted.
- Already on branch `fix/76-dot-parallel-branches`.

---

## File map

| File | Responsibility | Change |
|---|---|---|
| `export/dot.go` | DOT emission | Replace `applyParallelAttrs`; add branch encoder + helpers |
| `export/dot_test.go` | Export tests | Add branch emission tests |
| `migrate/migrate.go` | DOT import | Replace `buildParallelConfig`; add branch decoder + helpers; harden `inferParallelTargets` |
| `migrate/migrate_test.go` | Migrate tests | Add branch parse + inference-preservation tests |
| `migrate/parity.go` | `dippin --check` parity | Make `compareParallelConfigs` compare `Branches` |
| `migrate/coverage_test.go` | Parity tests | Add branch-mismatch test |
| `migrate/roundtrip_test.go` | Full round-trip tests | Add reserved-char + block-form acceptance tests |
| `docs/nodes.md` | User docs | Document block form + `branches=` attr |

---

## Task 1: Export — encode `branches=`

**Files:**
- Modify: `export/dot.go` (replace `applyParallelAttrs` at lines 376-381; add helpers)
- Test: `export/dot_test.go`

- [ ] **Step 1: Write the failing tests**

Add to `export/dot_test.go`:

```go
func TestExportDOTParallelBranches(t *testing.T) {
	w := &ir.Workflow{
		Name:  "test",
		Start: "S",
		Exit:  "J",
		Nodes: []*ir.Node{
			{ID: "S", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "go."}},
			{ID: "P", Kind: ir.NodeParallel, Config: ir.ParallelConfig{
				Targets: []string{"fast", "accurate"},
				Branches: []ir.BranchConfig{
					{Target: "fast", Model: "claude-haiku-4-5", Provider: "anthropic", Fidelity: "summary"},
					{Target: "accurate", Model: "claude-opus-4-7", Provider: "anthropic", Fidelity: "full"},
				},
			}},
			{ID: "fast", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "f."}},
			{ID: "accurate", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "a."}},
			{ID: "J", Kind: ir.NodeFanIn, Config: ir.FanInConfig{Sources: []string{"fast", "accurate"}}},
		},
	}
	out := ExportDOT(w, ExportOptions{})
	assertContains(t, out, `targets="fast,accurate"`)
	assertContains(t, out, `branches="target=fast;model=claude-haiku-4-5;provider=anthropic;fidelity=summary,target=accurate;model=claude-opus-4-7;provider=anthropic;fidelity=full"`)
}

// Block form with only targets (no per-branch overrides) must still emit
// branches= so re-import keeps block form instead of downgrading to inline.
func TestExportDOTParallelBranchesTargetOnly(t *testing.T) {
	attrs := map[string]string{}
	applyParallelAttrs(attrs, ir.ParallelConfig{
		Branches: []ir.BranchConfig{{Target: "A"}, {Target: "B"}},
	})
	if attrs["targets"] != "A,B" {
		t.Errorf("targets = %q, want A,B", attrs["targets"])
	}
	if attrs["branches"] != "target=A,target=B" {
		t.Errorf("branches = %q, want target=A,target=B", attrs["branches"])
	}
}
```

- [ ] **Step 2: Run the tests to verify they fail**

Run: `just test-pkg export`
Expected: FAIL — `branches=` not present in output; `TestExportDOTParallelBranchesTargetOnly` sees empty `attrs["branches"]`.

- [ ] **Step 3: Replace `applyParallelAttrs` and add helpers**

In `export/dot.go`, replace the existing `applyParallelAttrs` (lines 376-381):

```go
// applyParallelAttrs adds parallel-specific attributes: targets= (always) and
// branches= (when per-branch config is present).
func applyParallelAttrs(attrs map[string]string, cfg ir.ParallelConfig) {
	applyParallelTargetsAttr(attrs, cfg)
	applyParallelBranchesAttr(attrs, cfg)
}

// applyParallelTargetsAttr emits targets= from cfg.Targets, deriving from branch
// targets when Targets is empty but Branches is present (so block-form-only IR
// keeps its fan-out).
func applyParallelTargetsAttr(attrs map[string]string, cfg ir.ParallelConfig) {
	targets := cfg.Targets
	if len(targets) == 0 && len(cfg.Branches) > 0 {
		targets = parallelBranchTargets(cfg.Branches)
	}
	if len(targets) > 0 {
		attrs["targets"] = strings.Join(targets, ",")
	}
}

// applyParallelBranchesAttr emits branches= whenever per-branch config exists.
// Emitted even when every branch is target-only, so block form round-trips
// (does not downgrade to inline form on re-import).
func applyParallelBranchesAttr(attrs map[string]string, cfg ir.ParallelConfig) {
	if len(cfg.Branches) > 0 {
		attrs["branches"] = encodeBranches(cfg.Branches)
	}
}

// parallelBranchTargets extracts target IDs from branch configs. Local copy of
// the parser's branchTargets (packages import ir only, not each other).
func parallelBranchTargets(branches []ir.BranchConfig) []string {
	targets := make([]string, len(branches))
	for i, b := range branches {
		targets[i] = b.Target
	}
	return targets
}

// encodeBranches encodes a branch slice as comma-joined branch tokens, in slice
// order (order maps positionally to targets).
func encodeBranches(branches []ir.BranchConfig) string {
	parts := make([]string, 0, len(branches))
	for _, b := range branches {
		parts = append(parts, encodeBranch(b))
	}
	return strings.Join(parts, ",")
}

// encodeBranch encodes one branch as ';'-joined k=v tokens. target is always
// first; model/provider/fidelity only when non-empty.
func encodeBranch(b ir.BranchConfig) string {
	parts := []string{"target=" + encodeBranchToken(b.Target)}
	parts = appendBranchField(parts, "model", b.Model)
	parts = appendBranchField(parts, "provider", b.Provider)
	parts = appendBranchField(parts, "fidelity", b.Fidelity)
	return strings.Join(parts, ";")
}

// appendBranchField appends key=value only when value is non-empty.
func appendBranchField(parts []string, key, val string) []string {
	if val == "" {
		return parts
	}
	return append(parts, key+"="+encodeBranchToken(val))
}

// branchEncoder percent-encodes the reserved characters of the branches
// encoding: ';' (field sep), ',' (branch sep), '=' (k/v sep), '%' (escape char),
// and '\' (backslash, so a literal '\' + n/l/r cannot survive the DOT-quote
// layer as a DOT escape and be decoded to a newline by the migrate lexer). This
// is a superset of steerContextEncoder. Order is irrelevant (single-pass
// Replacer, mutually-exclusive patterns). The DOT-quote layer still handles '"'.
var branchEncoder = strings.NewReplacer(
	"%", "%25",
	",", "%2C",
	";", "%3B",
	"=", "%3D",
	"\\", "%5C",
)

// encodeBranchToken percent-encodes the reserved characters in a key
// or value so the round-trip through DOT → migrate stays lossless.
func encodeBranchToken(s string) string {
	if !strings.ContainsAny(s, ",;=%\\") {
		return s
	}
	return branchEncoder.Replace(s)
}
```

- [ ] **Step 4: Run the tests to verify they pass**

Run: `just test-pkg export`
Expected: PASS (including the existing `TestExportDOTParallelConfig`).

- [ ] **Step 5: Check complexity**

Run: `just complexity`
Expected: no output (no violations).

- [ ] **Step 6: Commit**

```bash
git add export/dot.go export/dot_test.go
git commit -m "feat: encode block-form parallel branches into DOT branches= attr (#76)

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

## Task 2: Migrate — decode `branches=`

**Files:**
- Modify: `migrate/migrate.go` (replace `buildParallelConfig` at lines 583-589; add helpers)
- Test: `migrate/migrate_test.go`

- [ ] **Step 1: Write the failing test**

Add to `migrate/migrate_test.go`:

```go
func TestMigrateParallelBranches(t *testing.T) {
	dot := `digraph G {
		Start [shape=Mdiamond];
		P [shape=component, targets="fast,accurate", branches="target=fast;model=claude-haiku-4-5;provider=anthropic;fidelity=summary,target=accurate;model=claude-opus-4-7;provider=anthropic;fidelity=full"];
		fast [shape=box];
		accurate [shape=box];
		Exit [shape=Msquare];
		Start -> P;
		P -> fast;
		P -> accurate;
		fast -> Exit;
		accurate -> Exit;
	}`
	w, err := Migrate(dot)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	cfg, ok := w.Node("P").Config.(ir.ParallelConfig)
	if !ok {
		t.Fatalf("config type = %T, want ParallelConfig", w.Node("P").Config)
	}
	if len(cfg.Branches) != 2 {
		t.Fatalf("branches = %d, want 2", len(cfg.Branches))
	}
	// Targets derived from branch order.
	if len(cfg.Targets) != 2 || cfg.Targets[0] != "fast" || cfg.Targets[1] != "accurate" {
		t.Errorf("targets = %v, want [fast accurate]", cfg.Targets)
	}
	b0 := cfg.Branches[0]
	if b0.Target != "fast" || b0.Model != "claude-haiku-4-5" || b0.Provider != "anthropic" || b0.Fidelity != "summary" {
		t.Errorf("branch[0] = %+v", b0)
	}
	b1 := cfg.Branches[1]
	if b1.Target != "accurate" || b1.Model != "claude-opus-4-7" || b1.Provider != "anthropic" || b1.Fidelity != "full" {
		t.Errorf("branch[1] = %+v", b1)
	}
}

// Inline form (no branches=) must leave Branches nil so the formatter keeps
// inline output.
func TestMigrateParallelInlineNoBranches(t *testing.T) {
	cfg := buildParallelConfig(map[string]string{"targets": "A,B"})
	if cfg.Branches != nil {
		t.Errorf("branches = %+v, want nil", cfg.Branches)
	}
	if len(cfg.Targets) != 2 {
		t.Errorf("targets = %v, want [A B]", cfg.Targets)
	}
}

// A branch token with no target= is dropped (a zero-target branch would corrupt
// the edge mapping); unknown keys are skipped.
func TestMigrateParallelBranchesMalformed(t *testing.T) {
	cfg := buildParallelConfig(map[string]string{
		"branches": "model=x,target=B;bogus=1;model=y",
	})
	if len(cfg.Branches) != 1 {
		t.Fatalf("branches = %d, want 1 (target-less branch dropped)", len(cfg.Branches))
	}
	if cfg.Branches[0].Target != "B" || cfg.Branches[0].Model != "y" {
		t.Errorf("branch[0] = %+v, want {Target:B Model:y}", cfg.Branches[0])
	}
}
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `just test-pkg migrate`
Expected: FAIL — `buildParallelConfig` ignores `branches=`; `cfg.Branches` is nil in `TestMigrateParallelBranches`.

- [ ] **Step 3: Replace `buildParallelConfig` and add helpers**

In `migrate/migrate.go`, replace the existing `buildParallelConfig` (lines 583-589):

```go
func buildParallelConfig(attrs map[string]string) ir.ParallelConfig {
	cfg := ir.ParallelConfig{}
	if v, ok := attrs["branches"]; ok && v != "" {
		cfg.Branches = parseBranches(v)
	}
	applyParallelTargets(&cfg, attrs)
	return cfg
}

// applyParallelTargets sets cfg.Targets. When branches are present they are the
// source of truth for target order (reproducing the parser invariant
// Targets[i] == Branches[i].Target). Otherwise targets= is used; if both are
// absent, inferParallelFanIn backfills from edges.
func applyParallelTargets(cfg *ir.ParallelConfig, attrs map[string]string) {
	if len(cfg.Branches) > 0 {
		cfg.Targets = parallelBranchTargets(cfg.Branches)
		return
	}
	if v, ok := attrs["targets"]; ok {
		cfg.Targets = splitComma(v)
	}
}

// parallelBranchTargets extracts target IDs from branch configs (local copy;
// packages import ir only).
func parallelBranchTargets(branches []ir.BranchConfig) []string {
	targets := make([]string, len(branches))
	for i, b := range branches {
		targets[i] = b.Target
	}
	return targets
}

// parseBranches parses the branches= attribute ("k=v;k=v,k=v;...") into
// BranchConfigs in order. A branch with an empty target is dropped (a branch
// with no fan-out target would corrupt the edge mapping). Unknown field keys
// are skipped, mirroring parseFlattenedSteerContext's non-erroring contract.
func parseBranches(s string) []ir.BranchConfig {
	var out []ir.BranchConfig
	for _, raw := range strings.Split(s, ",") {
		b := parseBranchToken(raw)
		if b.Target != "" {
			out = append(out, b)
		}
	}
	return out
}

// parseBranchToken parses a single ';'-joined branch token into a BranchConfig.
func parseBranchToken(s string) ir.BranchConfig {
	var b ir.BranchConfig
	for _, field := range strings.Split(s, ";") {
		applyBranchToken(&b, field)
	}
	return b
}

// branchFieldSetters maps a branch field key to the BranchConfig field it sets.
// Table-driven (rather than a switch) keeps applyBranchToken under the cyclo≤5
// cap with the extra target case + malformed guard, and makes adding a key
// (e.g. tool_access for #58) a one-line change.
var branchFieldSetters = map[string]func(*ir.BranchConfig, string){
	"target":   func(b *ir.BranchConfig, v string) { b.Target = v },
	"model":    func(b *ir.BranchConfig, v string) { b.Model = v },
	"provider": func(b *ir.BranchConfig, v string) { b.Provider = v },
	"fidelity": func(b *ir.BranchConfig, v string) { b.Fidelity = v },
}

// applyBranchToken sets one "key=value" field on a BranchConfig. The value is
// percent-decoded. Tokens without '=' and unknown keys are skipped. The value is
// NOT TrimSpace'd: branch targets are node IDs that must match edge endpoints
// exactly.
func applyBranchToken(b *ir.BranchConfig, field string) {
	kv := strings.SplitN(field, "=", 2)
	if len(kv) != 2 {
		return
	}
	if set, ok := branchFieldSetters[kv[0]]; ok {
		set(b, decodeBranchToken(kv[1]))
	}
}

// branchDecoder reverses branchEncoder from export/dot.go.
var branchDecoder = strings.NewReplacer(
	"%25", "%",
	"%2C", ",",
	"%3B", ";",
	"%3D", "=",
	"%5C", "\\",
)

// decodeBranchToken reverses encodeBranchToken from export. Returns the input
// unchanged when it contains no '%' escapes.
func decodeBranchToken(s string) string {
	if !strings.Contains(s, "%") {
		return s
	}
	return branchDecoder.Replace(s)
}
```

- [ ] **Step 4: Run the test to verify it passes**

Run: `just test-pkg migrate`
Expected: PASS (including existing `TestMigrateParallelExplicitTargets`, `TestMigrateParallelInference`).

- [ ] **Step 5: Check complexity**

Run: `just complexity`
Expected: no output.

- [ ] **Step 6: Commit**

```bash
git add migrate/migrate.go migrate/migrate_test.go
git commit -m "feat: decode DOT branches= attr into ParallelConfig.Branches (#76)

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

## Task 3: Harden `inferParallelTargets` to preserve `Branches`

**Files:**
- Modify: `migrate/migrate.go` (`inferParallelTargets`, line ~1023)
- Test: `migrate/migrate_test.go`

- [ ] **Step 1: Write the failing test**

Add to `migrate/migrate_test.go`:

```go
// inferParallelFanIn rebuilds ParallelConfig when Targets is empty; it must not
// discard Branches in the process.
func TestInferParallelTargetsPreservesBranches(t *testing.T) {
	w := &ir.Workflow{
		Nodes: []*ir.Node{
			{ID: "P", Kind: ir.NodeParallel, Config: ir.ParallelConfig{
				Branches: []ir.BranchConfig{{Target: "A", Model: "m"}},
			}},
		},
		Edges: []*ir.Edge{{From: "P", To: "A"}},
	}
	inferParallelFanIn(w)
	cfg := w.Node("P").Config.(ir.ParallelConfig)
	if len(cfg.Branches) != 1 || cfg.Branches[0].Model != "m" {
		t.Errorf("branches lost during inference: %+v", cfg.Branches)
	}
	if len(cfg.Targets) != 1 || cfg.Targets[0] != "A" {
		t.Errorf("targets = %v, want [A]", cfg.Targets)
	}
}
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `just test-pkg migrate`
Expected: FAIL — `branches lost during inference: []` (the rebuild at line 1023 drops `Branches`).

- [ ] **Step 3: Preserve `Branches` on rebuild**

In `migrate/migrate.go`, in `inferParallelTargets`, change the final assignment (line ~1023) from:

```go
	n.Config = ir.ParallelConfig{Targets: targets}
```

to:

```go
	n.Config = ir.ParallelConfig{Targets: targets, Branches: cfg.Branches}
```

- [ ] **Step 4: Run the test to verify it passes**

Run: `just test-pkg migrate`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add migrate/migrate.go migrate/migrate_test.go
git commit -m "fix: preserve Branches when inferring parallel targets from edges (#76)

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

## Task 4: Parity — compare `Branches` in `compareParallelConfigs`

**Files:**
- Modify: `migrate/parity.go` (`compareParallelConfigs`, lines 347-356)
- Test: `migrate/coverage_test.go`

- [ ] **Step 1: Write the failing test**

Add to `migrate/coverage_test.go`:

```go
func TestCompareParallelConfigs_DifferentBranches(t *testing.T) {
	a := ir.ParallelConfig{
		Targets:  []string{"A"},
		Branches: []ir.BranchConfig{{Target: "A", Model: "x"}},
	}
	b := ir.ParallelConfig{
		Targets:  []string{"A"},
		Branches: []ir.BranchConfig{{Target: "A", Model: "y"}},
	}
	diffs := compareParallelConfigs("P", "nodes.P", a, b)
	if len(diffs) == 0 {
		t.Fatal("expected a branches difference, got none")
	}
}
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `just test-pkg migrate`
Expected: FAIL — `compareParallelConfigs` only compares `Targets` (which are equal), so it returns no diffs.

- [ ] **Step 3: Compare `Branches` field-by-field**

In `migrate/parity.go`, replace `compareParallelConfigs` (lines 347-356):

```go
func compareParallelConfigs(id, path string, ac ir.ParallelConfig, bCfg interface{}) []Difference {
	bc, ok := bCfg.(ir.ParallelConfig)
	if !ok {
		return []Difference{configMismatchDiff(id, path, "ParallelConfig", bCfg)}
	}
	var diffs []Difference
	diffs = append(diffs, compareParallelTargets(id, ac, bc)...)
	diffs = append(diffs, compareParallelBranches(id, ac, bc)...)
	return diffs
}

func compareParallelTargets(id string, ac, bc ir.ParallelConfig) []Difference {
	if strings.Join(ac.Targets, ",") != strings.Join(bc.Targets, ",") {
		return []Difference{fieldDiff(id, "targets", fmt.Sprintf("node %q targets: %v vs %v", id, ac.Targets, bc.Targets))}
	}
	return nil
}

func compareParallelBranches(id string, ac, bc ir.ParallelConfig) []Difference {
	if branchesKey(ac.Branches) != branchesKey(bc.Branches) {
		return []Difference{fieldDiff(id, "branches", fmt.Sprintf("node %q branches differ", id))}
	}
	return nil
}

// branchesKey produces a stable comparison key for a branch slice (order
// matters — it maps positionally to targets).
func branchesKey(branches []ir.BranchConfig) string {
	parts := make([]string, len(branches))
	for i, b := range branches {
		parts[i] = fmt.Sprintf("%s|%s|%s|%s", b.Target, b.Model, b.Provider, b.Fidelity)
	}
	return strings.Join(parts, ",")
}
```

- [ ] **Step 4: Run the test to verify it passes**

Run: `just test-pkg migrate`
Expected: PASS (including existing `TestCompareParallelConfigs_TypeMismatch`, `TestCompareParallelConfigs_DifferentTargets`).

- [ ] **Step 5: Check complexity**

Run: `just complexity`
Expected: no output.

- [ ] **Step 6: Commit**

```bash
git add migrate/parity.go migrate/coverage_test.go
git commit -m "fix: parity check compares parallel Branches, not just Targets (#76)

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

## Task 5: Reserved-char + acceptance round-trip tests

**Files:**
- Test: `migrate/roundtrip_test.go` (package `migrate`; already imports `export`, `ir`, `dipparser "…/parser"`, and has the `extractDOTAttr` helper)

- [ ] **Step 1: Write the reserved-char round-trip test**

Add to `migrate/roundtrip_test.go`:

```go
// Branch values containing reserved delimiters (',', ';', '=', '%') must
// round-trip losslessly through encodeBranches → parseBranches.
func TestExportDOT_Parallel_BranchesReservedCharsRoundTrip(t *testing.T) {
	original := []ir.BranchConfig{
		{Target: "fast", Model: "a,b", Provider: "x=y", Fidelity: "100%"},
		{Target: "accurate", Model: "p;q"},
	}
	w := &ir.Workflow{
		Name:  "W",
		Start: "S",
		Exit:  "E",
		Nodes: []*ir.Node{
			{ID: "S", Kind: ir.NodeAgent, Config: ir.AgentConfig{}},
			{ID: "P", Kind: ir.NodeParallel, Config: ir.ParallelConfig{
				Targets:  []string{"fast", "accurate"},
				Branches: original,
			}},
			{ID: "fast", Kind: ir.NodeAgent, Config: ir.AgentConfig{}},
			{ID: "accurate", Kind: ir.NodeAgent, Config: ir.AgentConfig{}},
			{ID: "E", Kind: ir.NodeAgent, Config: ir.AgentConfig{}},
		},
		Edges: []*ir.Edge{{From: "S", To: "P"}, {From: "P", To: "fast"}, {From: "P", To: "accurate"}, {From: "fast", To: "E"}, {From: "accurate", To: "E"}},
	}
	dot := export.ExportDOT(w, export.ExportOptions{})
	attrVal := extractDOTAttr(dot, "branches")
	if attrVal == "" {
		t.Fatalf("branches attr not found in DOT output:\n%s", dot)
	}
	got := parseBranches(attrVal)
	if len(got) != len(original) {
		t.Fatalf("branches = %d, want %d", len(got), len(original))
	}
	for i, want := range original {
		if got[i] != want {
			t.Errorf("branch[%d] round-trip: got %+v want %+v", i, got[i], want)
		}
	}
}
```

- [ ] **Step 2: Run it to verify it passes**

Run: `just test-pkg migrate`
Expected: PASS (Tasks 1–2 already implemented the encode/decode; this confirms reserved-char safety). If it FAILS, the encoder/decoder reserved set is wrong — fix in `export/dot.go` / `migrate/migrate.go` before continuing.

- [ ] **Step 3: Write the block-form acceptance round-trip test (`.dip` → DOT → migrate)**

Add to `migrate/roundtrip_test.go`:

```go
// Acceptance test for #76: a block-form-only parallel node survives
// parse → ExportDOT → Migrate with fan-out targets AND per-branch config intact.
func TestRoundtripBlockFormParallel(t *testing.T) {
	src := `workflow RT
  start: split
  exit: join

  agent fast
    prompt: "f."

  agent accurate
    prompt: "a."

  parallel split
    branch: fast
      model: claude-haiku-4-5
      provider: anthropic
      fidelity: summary
    branch: accurate
      model: claude-opus-4-7
      provider: anthropic
      fidelity: full

  fan_in join <- fast, accurate

  edges
    split -> fast
    split -> accurate
    fast -> join
    accurate -> join
`
	p := dipparser.NewParser(src, "rt.dip")
	w, err := p.Parse()
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	dot := export.ExportDOT(w, export.ExportOptions{})
	got, err := Migrate(dot)
	if err != nil {
		t.Fatalf("migrate: %v", err)
	}
	cfg, ok := got.Node("split").Config.(ir.ParallelConfig)
	if !ok {
		t.Fatalf("config type = %T, want ParallelConfig", got.Node("split").Config)
	}
	if len(cfg.Targets) != 2 || cfg.Targets[0] != "fast" || cfg.Targets[1] != "accurate" {
		t.Errorf("targets = %v, want [fast accurate]", cfg.Targets)
	}
	if len(cfg.Branches) != 2 {
		t.Fatalf("branches = %d, want 2", len(cfg.Branches))
	}
	want := []ir.BranchConfig{
		{Target: "fast", Model: "claude-haiku-4-5", Provider: "anthropic", Fidelity: "summary"},
		{Target: "accurate", Model: "claude-opus-4-7", Provider: "anthropic", Fidelity: "full"},
	}
	for i := range want {
		if cfg.Branches[i] != want[i] {
			t.Errorf("branch[%d]: got %+v want %+v", i, cfg.Branches[i], want[i])
		}
	}
}
```

> Note: `dipparser.NewParser(src, name)` + `.Parse()` matches the parser API used in `parser/parser_test.go`. If the parser constructor signature differs at implementation time, check `parser/parser.go` for the exact `NewParser` signature and adjust.

- [ ] **Step 4: Run it to verify it passes**

Run: `just test-pkg migrate`
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add migrate/roundtrip_test.go
git commit -m "test: reserved-char + block-form parallel DOT round-trip (#76)

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

## Task 6: Document block-form parallel + `branches=` attribute

**Files:**
- Modify: `docs/nodes.md` (parallel section, ~lines 344-362)

- [ ] **Step 1: Read the current parallel section**

Run: `just` is not needed — open `docs/nodes.md` around lines 344-362 and read the existing inline-only `parallel` documentation and (for style) the `steer_context` percent-encoding note (~line 507).

- [ ] **Step 2: Add block-form + DOT-attr documentation**

After the existing inline `parallel` documentation in `docs/nodes.md`, add a block-form subsection. Use this content (adjust headings to match the file's existing style):

```markdown
#### Block form (per-branch config)

When branches need different models, providers, or fidelity, use block form:

    parallel split
      branch: fast
        model: claude-haiku-4-5
        provider: anthropic
        fidelity: summary
      branch: accurate
        model: claude-opus-4-7
        provider: anthropic
        fidelity: full

Block form populates both the fan-out targets and per-branch overrides.

**DOT mapping.** Block-form parallels export a `branches=` node attribute
alongside `targets=`. Each branch is encoded as `;`-joined `key=value` tokens
(`target` plus any of `model`/`provider`/`fidelity`), branches joined by `,`:

    branches="target=fast;model=claude-haiku-4-5;provider=anthropic;fidelity=summary,target=accurate;model=claude-opus-4-7;provider=anthropic;fidelity=full"

The reserved characters `%`, `,`, `;`, `=` are percent-encoded in keys and values
(as with `steer_context`), so they round-trip losslessly through `migrate`.
```

- [ ] **Step 3: Verify docs build / examples still pass**

Run: `just validate-examples`
Expected: all examples validate (no behavior change; this confirms nothing broke). Note: `docs/generated-spec.md` may be auto-regenerated by the pre-commit hook — that is expected.

- [ ] **Step 4: Commit**

```bash
git add docs/nodes.md
git commit -m "docs: document block-form parallel and branches= DOT attribute (#76)

Co-Authored-By: Claude Opus 4.8 (1M context) <noreply@anthropic.com>"
```

---

## Task 7: Full gate + PR

- [ ] **Step 1: Run the full gate**

Run: `just check`
Expected: all checks pass (build, vet, fmt, test-race, complexity, validate-examples).

- [ ] **Step 2: Open the PR**

```bash
git push -u origin fix/76-dot-parallel-branches
gh pr create --title "fix: DOT round-trip for block-form parallel branches (#76)" --body "$(cat <<'EOF'
## Summary
Block-form `parallel` nodes now round-trip through `dippin export` → DOT → `migrate` without losing per-branch `model`/`provider`/`fidelity` or their fan-out targets. Fixes #76.

- Export emits a new `branches=` node attribute (percent-encoded `k=v` tokens, mirroring `steer_context`), always co-emitting `targets=`.
- Migrate parses `branches=` back into `[]ir.BranchConfig`; branch order is the source of truth for `Targets`.
- Fixed two latent gaps: `inferParallelTargets` no longer discards `Branches`, and the `--check` parity comparison now compares `Branches` (previously it compared only `Targets`, which is why the loss was silent).

## Unblocks
This unblocks #58 (per-branch `tool_access` override): per-branch attributes can now survive DOT, and the `k=v` encoding makes adding a `tool_access` key a one-line change.

## Testing
- Export/migrate unit tests, reserved-char round-trip, and a block-form-only `.dip` → DOT → `migrate` acceptance test.
- `just check` green.

Note: CHANGELOG is intentionally untouched (updated at tag time per CLAUDE.md).
EOF
)"
```

---

## Self-review notes (author)

- **Spec coverage:** encoding (T1), decoding+source-of-truth+malformed policy (T2), inference clobber (T3), parity blind spot (T4), reserved-char + idempotence/acceptance tests (T5), docs (T6), example file deliberately omitted (spec "Out of scope"). All spec sections map to a task.
- **Complexity:** `applyBranchToken` uses a table-driven setter map (not a switch) specifically because the 4 cases + malformed guard would hit cyclo 6. All other new functions are single-loop/single-guard. `just complexity` is run in T1, T2, T4.
- **Type consistency:** `parallelBranchTargets` is defined locally in both `export` and `migrate` (separate packages, same name by design). Encoder/decoder names: `branchEncoder`/`encodeBranchToken` (export), `branchDecoder`/`decodeBranchToken` (migrate). `branchesKey` only in parity.
