package simulate

import (
	"fmt"

	"github.com/2389-research/dippin-lang/ir"
)

// walkParallelBranches visits all targets of a parallel fan-out.
// Each branch is walked from its target node until it reaches the
// corresponding fan-in join. Returns the join node ID as the next
// node to continue from.
//
// opts.Branch is a set of target node IDs: when non-empty, only the targets it
// contains are walked (for targeted test scenarios); an empty set walks all.
// If a non-empty set matches none of this parallel's targets, the filter is
// ignored and all targets are walked (no matches → all).
func (s *simulator) walkParallelBranches(cfg ir.ParallelConfig) (string, *Result, error) {
	targets := resolveTargets(cfg)
	joinID := s.findJoinNode(targets)

	selected := s.selectBranches(targets)
	for _, target := range selected {
		if err := s.walkBranchToJoin(target, joinID); err != nil {
			return "", nil, err
		}
	}

	if joinID == "" {
		return "", s.finishRun("dead_end"), nil
	}
	return joinID, nil, nil
}

// resolveTargets extracts target node IDs from a ParallelConfig.
// Handles both inline (Targets) and block (Branches) forms.
func resolveTargets(cfg ir.ParallelConfig) []string {
	if len(cfg.Targets) > 0 {
		return cfg.Targets
	}
	targets := make([]string, len(cfg.Branches))
	for i, b := range cfg.Branches {
		targets[i] = b.Target
	}
	return targets
}

// selectBranches filters targets based on opts.Branch overrides.
// If no override exists for any of the targets, all are returned.
func (s *simulator) selectBranches(targets []string) []string {
	if len(s.opts.Branch) == 0 {
		return targets
	}
	var selected []string
	for _, t := range targets {
		if s.opts.Branch[t] {
			selected = append(selected, t)
		}
	}
	if len(selected) == 0 {
		return targets
	}
	return selected
}

// findJoinNode locates the fan_in node whose sources match the given targets.
func (s *simulator) findJoinNode(targets []string) string {
	targetSet := make(map[string]bool, len(targets))
	for _, t := range targets {
		targetSet[t] = true
	}
	for _, n := range s.workflow.Nodes {
		if cfg, ok := n.Config.(ir.FanInConfig); ok {
			if sourcesMatchTargets(cfg.Sources, targetSet) {
				return n.ID
			}
		}
	}
	return ""
}

// sourcesMatchTargets checks if all fan-in sources are in the target set.
func sourcesMatchTargets(sources []string, targets map[string]bool) bool {
	for _, src := range sources {
		if !targets[src] {
			return false
		}
	}
	return len(sources) > 0
}

// walkBranchToJoin walks from a branch target node until it reaches
// the join node (or a dead end). Each node along the way is visited
// and its events are emitted into the main simulator.
func (s *simulator) walkBranchToJoin(target, joinID string) error {
	current := target
	for s.steps < maxSteps {
		if current == joinID {
			return nil
		}
		next, err := s.walkBranchStep(current)
		if err != nil {
			return err
		}
		if next == "" {
			return nil
		}
		current = next
	}
	return fmt.Errorf("branch exceeded %d steps", maxSteps)
}

// walkBranchStep visits a single node in a branch and returns the next node.
func (s *simulator) walkBranchStep(current string) (string, error) {
	node := s.workflow.Node(current)
	if node == nil {
		return "", fmt.Errorf("branch node %q not found", current)
	}
	if err := s.visitNode(node); err != nil {
		return "", err
	}
	s.steps++
	return s.resolveNext(node)
}
