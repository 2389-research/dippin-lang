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
