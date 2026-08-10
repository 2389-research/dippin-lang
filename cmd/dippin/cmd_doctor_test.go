package main

import (
	"testing"

	"github.com/2389-research/dippin-lang/doctor"
	"github.com/2389-research/dippin-lang/validator"
)

// doctorHints computes the doctor lint-hint count the CLI way: validate + lint,
// then the cross-file tool_access supersession, then DiagnoseFromDiagnostics —
// mirroring CmdDoctor.
func doctorHints(t *testing.T, dir, entry string) int {
	t.Helper()
	w, err := loadWorkflow(dir + "/" + entry)
	if err != nil {
		t.Fatalf("load %s: %v", entry, err)
	}
	diags := append(validator.Validate(w).Diagnostics,
		validator.LintWithOptions(w, validator.Options{}).Diagnostics...)
	diags = applyCrossFileToolAccess(diags, w, dir+"/"+entry)
	return doctor.DiagnoseFromDiagnostics(w, diags).Lint.Hints
}

// TestDoctor_CrossFileSupersedesDIP143 proves doctor's hint count now reflects
// the cross-file pass: a fully-restricted resolved child supersedes the per-file
// DIP143 advisory, so doctor reports one fewer hint than the pass-free summary
// (#101). This is exactly the "only a fully-restricted child differs by one hint"
// case the prior known-limitation note described.
func TestDoctor_CrossFileSupersedesDIP143(t *testing.T) {
	dir := writeWorkflows(t, map[string]string{
		"entry.dip": entryRestrictsRefs("child.dip"),
		"child.dip": childFullRestrict,
	})
	w, err := loadWorkflow(dir + "/entry.dip")
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	// Pass-free summary (old behavior): the entry declares tool_access and
	// references a child, so validator.Lint emits the DIP143 advisory (a hint).
	before := doctor.DiagnoseWithOptions(w, validator.Options{}).Lint.Hints
	after := doctorHints(t, dir, "entry.dip")

	if before == 0 {
		t.Fatal("expected the pass-free summary to carry a DIP143 hint to supersede")
	}
	if after != before-1 {
		t.Errorf("cross-file supersession should drop one hint: before=%d after=%d", before, after)
	}
}

// TestDoctor_CrossFileZeroIntentKeepsHintCount confirms a zero-intent child swaps
// DIP143 for DIP146 (both hints) — the count is unchanged, but the pass still runs.
func TestDoctor_CrossFileZeroIntentKeepsHintCount(t *testing.T) {
	dir := writeWorkflows(t, map[string]string{
		"entry.dip": entryRestrictsRefs("child.dip"),
		"child.dip": childZeroIntent,
	})
	w, err := loadWorkflow(dir + "/entry.dip")
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	before := doctor.DiagnoseWithOptions(w, validator.Options{}).Lint.Hints
	after := doctorHints(t, dir, "entry.dip")
	if after != before {
		t.Errorf("zero-intent child swaps DIP143→DIP146 (both hints); count should hold: before=%d after=%d", before, after)
	}
}
