// Package scaffold builds starter ir.Workflow values for common pipeline topologies.
// Templates are constructed programmatically to guarantee canonical output via formatter.Format().
package scaffold

import (
	"fmt"
	"time"

	"github.com/2389-research/dippin-lang/ir"
)

// TemplateNames returns the sorted list of available template names.
func TemplateNames() []string {
	return []string{"conditional", "human-gate", "manager_loop", "minimal", "parallel", "review-loop"}
}

// templateBuilders maps template names to their builder functions.
var templateBuilders = map[string]func(string) *ir.Workflow{
	"minimal":      buildMinimal,
	"parallel":     buildParallel,
	"conditional":  buildConditional,
	"review-loop":  buildReviewLoop,
	"human-gate":   buildHumanGate,
	"manager_loop": buildManagerLoop,
}

// Build constructs an ir.Workflow for the named template.
// The name parameter overrides the workflow name (defaults to the template name).
func Build(template, name string) (*ir.Workflow, error) {
	if name == "" {
		name = template
	}
	builder, ok := templateBuilders[template]
	if !ok {
		return nil, fmt.Errorf("unknown template %q (available: %v)", template, TemplateNames())
	}
	return builder(name), nil
}

func buildMinimal(name string) *ir.Workflow {
	return &ir.Workflow{
		Name:  name,
		Goal:  "A minimal two-node workflow",
		Start: "Start",
		Exit:  "Done",
		// A graph-level failure route so every generated template lints clean
		// (DIP144) and demonstrates the on_failure catch-all.
		Defaults: ir.WorkflowDefaults{OnFailure: "Done"},
		Nodes: []*ir.Node{
			{ID: "Start", Kind: ir.NodeAgent, Config: ir.AgentConfig{
				Prompt: "Begin the task.",
			}},
			{ID: "Done", Kind: ir.NodeAgent, Config: ir.AgentConfig{
				Prompt: "Summarize the result.",
			}},
		},
		Edges: []*ir.Edge{
			{From: "Start", To: "Done"},
		},
	}
}

func buildParallel(name string) *ir.Workflow {
	return &ir.Workflow{
		Name:  name,
		Goal:  "Fan-out to parallel workers then join",
		Start: "Init",
		Exit:  "Done",
		// A graph-level failure route so every generated template lints clean
		// (DIP144) and demonstrates the on_failure catch-all.
		Defaults: ir.WorkflowDefaults{OnFailure: "Done"},
		Nodes: []*ir.Node{
			{ID: "Init", Kind: ir.NodeAgent, Config: ir.AgentConfig{
				Prompt: "Prepare work for parallel execution.",
			}},
			{ID: "Fan", Kind: ir.NodeParallel, Config: ir.ParallelConfig{
				Targets: []string{"WorkerA", "WorkerB"},
			}},
			{ID: "WorkerA", Kind: ir.NodeAgent, Config: ir.AgentConfig{
				Prompt: "Handle the first work stream.",
			}},
			{ID: "WorkerB", Kind: ir.NodeAgent, Config: ir.AgentConfig{
				Prompt: "Handle the second work stream.",
			}},
			{ID: "Join", Kind: ir.NodeFanIn, Config: ir.FanInConfig{
				Sources: []string{"WorkerA", "WorkerB"},
			}},
			{ID: "Done", Kind: ir.NodeAgent, Config: ir.AgentConfig{
				Prompt: "Combine results and finish.",
			}},
		},
		Edges: []*ir.Edge{
			{From: "Init", To: "Fan"},
			{From: "Fan", To: "WorkerA"},
			{From: "Fan", To: "WorkerB"},
			{From: "WorkerA", To: "Join"},
			{From: "WorkerB", To: "Join"},
			{From: "Join", To: "Done"},
		},
	}
}

func buildConditional(name string) *ir.Workflow {
	return &ir.Workflow{
		Name:  name,
		Goal:  "Route based on status check",
		Start: "Check",
		Exit:  "Done",
		// A graph-level failure route so every generated template lints clean
		// (DIP144) and demonstrates the on_failure catch-all.
		Defaults: ir.WorkflowDefaults{OnFailure: "Done"},
		Nodes: []*ir.Node{
			{ID: "Check", Kind: ir.NodeConditional, Label: "Evaluate outcome", Config: ir.ConditionalConfig{}},
			{ID: "Pass", Kind: ir.NodeAgent, Config: ir.AgentConfig{
				Prompt: "Handle the success path.",
			}},
			{ID: "Fail", Kind: ir.NodeAgent, Config: ir.AgentConfig{
				Prompt: "Handle the failure path.",
			}},
			{ID: "Done", Kind: ir.NodeAgent, Config: ir.AgentConfig{
				Prompt: "Wrap up.",
			}},
		},
		Edges: []*ir.Edge{
			{From: "Check", To: "Pass", Condition: &ir.Condition{
				Raw: "ctx.outcome = success",
			}},
			{From: "Check", To: "Fail", Condition: &ir.Condition{
				Raw: "ctx.outcome = fail",
			}},
			{From: "Check", To: "Done"},
			{From: "Pass", To: "Done"},
			{From: "Fail", To: "Done"},
		},
	}
}

func buildReviewLoop(name string) *ir.Workflow {
	return &ir.Workflow{
		Name:  name,
		Goal:  "Implement, review, and loop until approved",
		Start: "Implement",
		Exit:  "Done",
		// A graph-level failure route so every generated template lints clean
		// (DIP144) and demonstrates the on_failure catch-all.
		Defaults: ir.WorkflowDefaults{OnFailure: "Done"},
		Nodes: []*ir.Node{
			{ID: "Implement", Kind: ir.NodeAgent, Config: ir.AgentConfig{
				Prompt: "Write the implementation.",
			}},
			{ID: "Review", Kind: ir.NodeAgent, Config: ir.AgentConfig{
				AutoStatus: true,
				Prompt:     "Review the implementation. Set STATUS: success if approved, STATUS: fail if changes needed.",
			}},
			{ID: "Done", Kind: ir.NodeAgent, Config: ir.AgentConfig{
				Prompt: "Finalize the approved result.",
			}},
		},
		Edges: []*ir.Edge{
			{From: "Implement", To: "Review"},
			{From: "Review", To: "Done", Condition: &ir.Condition{
				Raw: "ctx.outcome = success",
			}},
			{From: "Review", To: "Implement", Restart: true, Condition: &ir.Condition{
				Raw: "ctx.outcome = fail",
			}},
		},
	}
}

func buildHumanGate(name string) *ir.Workflow {
	return &ir.Workflow{
		Name:  name,
		Goal:  "Gate progress on human approval",
		Start: "Prepare",
		Exit:  "Done",
		// A graph-level failure route so every generated template lints clean
		// (DIP144) and demonstrates the on_failure catch-all.
		Defaults: ir.WorkflowDefaults{OnFailure: "Done"},
		Nodes: []*ir.Node{
			{ID: "Prepare", Kind: ir.NodeAgent, Config: ir.AgentConfig{
				Prompt: "Prepare the proposal for review.",
			}},
			{ID: "Gate", Kind: ir.NodeHuman, Config: ir.HumanConfig{
				Mode:    "choice",
				Default: "approve",
			}},
			{ID: "Approved", Kind: ir.NodeAgent, Config: ir.AgentConfig{
				Prompt: "Proceed with the approved plan.",
			}},
			{ID: "Rejected", Kind: ir.NodeAgent, Config: ir.AgentConfig{
				Prompt: "Handle the rejection.",
			}},
			{ID: "Done", Kind: ir.NodeAgent, Config: ir.AgentConfig{
				Prompt: "Wrap up.",
			}},
		},
		Edges: []*ir.Edge{
			{From: "Prepare", To: "Gate"},
			{From: "Gate", To: "Approved", Label: "approve", Choice: "approve"},
			{From: "Gate", To: "Rejected", Label: "reject", Choice: "reject"},
			{From: "Approved", To: "Done"},
			{From: "Rejected", To: "Done"},
		},
	}
}

func buildManagerLoop(name string) *ir.Workflow {
	return &ir.Workflow{
		Name:  name,
		Goal:  "Supervise a child pipeline with periodic steering",
		Start: "Supervise",
		Exit:  "Done",
		// A graph-level failure route (clears DIP144, demonstrates on_failure).
		// This template still surfaces DIP135 by design — it references an external
		// child_pipeline.dip the author must create — so it is the one template not
		// fully lint-clean out of the box (see TestTemplatesLintClean).
		Defaults: ir.WorkflowDefaults{OnFailure: "Done"},
		Nodes: []*ir.Node{
			{ID: "Supervise", Kind: ir.NodeManagerLoop, Label: "Quality Gate Supervisor", Config: ir.ManagerLoopConfig{
				SubgraphRef:    "child_pipeline.dip",
				PollInterval:   30 * time.Second,
				MaxCycles:      10,
				StopCondition:  &ir.Condition{Raw: "stack.child.outcome = success"},
				SteerCondition: &ir.Condition{Raw: "stack.child.cycles = 5"},
				SteerContext:   map[string]string{"hint": "halfway_through"},
			}},
			{ID: "Done", Kind: ir.NodeAgent, Config: ir.AgentConfig{
				Prompt: "Summarize the supervised run outcome.",
			}},
		},
		Edges: []*ir.Edge{
			{From: "Supervise", To: "Done"},
		},
	}
}
