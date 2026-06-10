package migrate

import (
	"testing"

	"github.com/2389-research/dippin-lang/ir"
)

func TestParity_AgentWritablePathsDiff(t *testing.T) {
	a := ir.AgentConfig{WritablePaths: []string{"workspace/**"}}
	b := ir.AgentConfig{WritablePaths: []string{"workspace/**", ".ai/**"}}
	diffs := compareAgentConfigs("N", "", a, b)
	if len(diffs) == 0 {
		t.Fatalf("expected a writable_paths difference, got none")
	}
}

func TestParity_BranchWritablePathsDiff(t *testing.T) {
	a := ir.ParallelConfig{Branches: []ir.BranchConfig{{Target: "x", WritablePaths: []string{"workspace/**"}}}}
	b := ir.ParallelConfig{Branches: []ir.BranchConfig{{Target: "x", WritablePaths: []string{".ai/**"}}}}
	diffs := compareParallelBranches("N", a, b)
	if len(diffs) == 0 {
		t.Fatalf("expected a branch writable_paths difference, got none")
	}
}

func TestParity_AgentLastResponseTruncateDiff(t *testing.T) {
	a := ir.AgentConfig{LastResponseTruncate: 4096}
	b := ir.AgentConfig{LastResponseTruncate: 2048}
	diffs := compareAgentBehavior("A", a, b)
	if len(diffs) == 0 {
		t.Error("expected a last_response_truncate diff, got none")
	}
}

func TestParity_AgentLastResponseTruncateEqual(t *testing.T) {
	a := ir.AgentConfig{LastResponseTruncate: 4096}
	b := ir.AgentConfig{LastResponseTruncate: 4096}
	if diffs := compareAgentBehavior("A", a, b); len(diffs) != 0 {
		t.Errorf("expected no diff for equal values, got %v", diffs)
	}
}

func TestParity_AgentLastResponseTruncateZeroNoDiff(t *testing.T) {
	a := ir.AgentConfig{}
	b := ir.AgentConfig{}
	if diffs := compareAgentBehavior("A", a, b); len(diffs) != 0 {
		t.Errorf("expected no diff for zero/unset values, got %v", diffs)
	}
}

func TestParity_BranchLastResponseTruncateDiff(t *testing.T) {
	a := ir.ParallelConfig{Branches: []ir.BranchConfig{{Target: "x", LastResponseTruncate: 4096}}}
	b := ir.ParallelConfig{Branches: []ir.BranchConfig{{Target: "x", LastResponseTruncate: 2048}}}
	diffs := compareParallelBranches("N", a, b)
	if len(diffs) == 0 {
		t.Fatalf("expected a branch last_response_truncate difference, got none")
	}
}
