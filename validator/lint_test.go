package validator

import (
	"strings"
	"testing"
	"time"

	"github.com/2389-research/dippin-lang/ir"
)

// --- Test fixtures ---

// cleanMinimalWorkflow returns a minimal valid workflow with no lint warnings.
func cleanMinimalWorkflow() *ir.Workflow {
	return &ir.Workflow{
		Name:  "clean",
		Start: "Begin",
		Exit:  "End",
		Nodes: []*ir.Node{
			{ID: "Begin", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "Hello."}},
			{ID: "End", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "Done."}},
		},
		Edges: []*ir.Edge{
			{From: "Begin", To: "End"},
		},
	}
}

// cleanComplexWorkflow returns a complex valid workflow with no lint warnings.
func cleanComplexWorkflow() *ir.Workflow {
	return &ir.Workflow{
		Name:  "clean_complex",
		Start: "Ask",
		Exit:  "Done",
		Defaults: ir.WorkflowDefaults{
			Model:    "claude-opus-4-6",
			Provider: "anthropic",
		},
		Nodes: []*ir.Node{
			{ID: "Ask", Kind: ir.NodeHuman, Config: ir.HumanConfig{Mode: "freeform"},
				IO: ir.NodeIO{Writes: []string{"human_response"}}},
			{ID: "Plan", Kind: ir.NodeAgent, Config: ir.AgentConfig{
				Prompt: "Plan the work based on ${ctx.human_response}.",
			}, IO: ir.NodeIO{Reads: []string{"human_response"}, Writes: []string{"plan"}}},
			{ID: "Execute", Kind: ir.NodeAgent, Config: ir.AgentConfig{
				Prompt: "Execute the plan.",
			}, IO: ir.NodeIO{Reads: []string{"plan"}}},
			{ID: "Done", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "Ship it."}},
		},
		Edges: []*ir.Edge{
			{From: "Ask", To: "Plan"},
			{From: "Plan", To: "Execute"},
			{From: "Execute", To: "Done"},
		},
	}
}

// --- Table-driven tests ---

func TestLint(t *testing.T) {
	tests := []struct {
		name       string
		workflow   *ir.Workflow
		wantCodes  []string // Expected diagnostic codes (empty = no diagnostics)
		wantNoDiag bool     // If true, expect zero diagnostics
	}{
		// --- Happy path ---
		{
			name:       "clean minimal workflow",
			workflow:   cleanMinimalWorkflow(),
			wantNoDiag: true,
		},
		{
			name:       "clean complex workflow with IO contracts",
			workflow:   cleanComplexWorkflow(),
			wantNoDiag: true,
		},
		{
			name: "workflow with known model/provider is clean",
			workflow: &ir.Workflow{
				Name:  "known_model",
				Start: "A",
				Exit:  "A",
				Nodes: []*ir.Node{
					{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{
						Prompt:   "Hello.",
						Model:    "gpt-5.4",
						Provider: "openai",
					}},
				},
			},
			wantNoDiag: true,
		},
		{
			name: "tool with timeout is clean",
			workflow: &ir.Workflow{
				Name:  "tool_timeout",
				Start: "T",
				Exit:  "T",
				Nodes: []*ir.Node{
					{ID: "T", Kind: ir.NodeTool, Config: ir.ToolConfig{
						Command: "echo test",
						Timeout: 30 * time.Second,
					}},
				},
			},
			wantNoDiag: true,
		},

		// --- DIP101: Unreachable nodes after conditional branches ---
		{
			name: "DIP101: source has unconditional edge, conditional branch is safe",
			workflow: &ir.Workflow{
				Name:  "cond_only",
				Start: "A",
				Exit:  "C",
				Nodes: []*ir.Node{
					{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "a"}},
					{ID: "B", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "b"}},
					{ID: "C", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "c"}},
				},
				Edges: []*ir.Edge{
					{From: "A", To: "B", Condition: &ir.Condition{
						Raw:    "ctx.x = 1",
						Parsed: ir.CondCompare{Variable: "ctx.x", Op: "=", Value: "1"},
					}},
					{From: "A", To: "C"},
					{From: "B", To: "C"},
				},
			},
			// A has unconditional edge to C → conditional branch to B is intentional.
			wantNoDiag: true,
		},
		{
			name: "DIP101: node with unconditional incoming edge is fine",
			workflow: &ir.Workflow{
				Name:  "uncond_ok",
				Start: "A",
				Exit:  "B",
				Nodes: []*ir.Node{
					{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "a"}},
					{ID: "B", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "b"}},
				},
				Edges: []*ir.Edge{
					{From: "A", To: "B"},
				},
			},
			wantNoDiag: true,
		},

		// --- DIP102: Routing node without default edge ---
		{
			name: "DIP102: single conditional outgoing with no default",
			workflow: &ir.Workflow{
				Name:  "no_default",
				Start: "A",
				Exit:  "C",
				Nodes: []*ir.Node{
					{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "a"}},
					{ID: "B", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "b"}},
					{ID: "C", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "c"}},
				},
				Edges: []*ir.Edge{
					{From: "A", To: "B", Condition: &ir.Condition{
						Raw: "ctx.x = 1",
					}},
					{From: "B", To: "C"},
				},
			},
			// Single conditional edge, no default — DIP102 fires.
			// Also DIP101 on B (only reachable via conditional, not a partition).
			wantCodes: []string{DIP101, DIP102},
		},
		{
			name: "DIP102: mixed conditional + unconditional is fine",
			workflow: &ir.Workflow{
				Name:  "with_default",
				Start: "A",
				Exit:  "C",
				Nodes: []*ir.Node{
					{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "a"}},
					{ID: "B", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "b"}},
					{ID: "C", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "c"}},
				},
				Edges: []*ir.Edge{
					{From: "A", To: "B", Condition: &ir.Condition{
						Raw:    "ctx.x = 1",
						Parsed: ir.CondCompare{Variable: "ctx.x", Op: "=", Value: "1"},
					}},
					{From: "A", To: "C"}, // unconditional default
					{From: "B", To: "C"},
				},
			},
			// A has unconditional edge → conditional branch to B is intentional. No DIP101 or DIP102.
			wantNoDiag: true,
		},

		// --- DIP103: Overlapping conditions ---
		{
			name: "DIP103: two edges with same condition",
			workflow: &ir.Workflow{
				Name:  "overlap",
				Start: "A",
				Exit:  "C",
				Nodes: []*ir.Node{
					{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "a"}},
					{ID: "B", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "b"}},
					{ID: "C", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "c"}},
				},
				Edges: []*ir.Edge{
					{From: "A", To: "B", Condition: &ir.Condition{
						Raw:    "ctx.outcome = success",
						Parsed: ir.CondCompare{Variable: "ctx.outcome", Op: "=", Value: "success"},
					}},
					{From: "A", To: "C", Condition: &ir.Condition{
						Raw:    "ctx.outcome = success",
						Parsed: ir.CondCompare{Variable: "ctx.outcome", Op: "=", Value: "success"},
					}},
				},
			},
			wantCodes: []string{DIP103},
		},
		{
			name: "DIP103: different conditions from same node is fine for overlap",
			workflow: &ir.Workflow{
				Name:  "no_overlap",
				Start: "A",
				Exit:  "C",
				Nodes: []*ir.Node{
					{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "a"}},
					{ID: "B", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "b"}},
					{ID: "C", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "c"}},
				},
				Edges: []*ir.Edge{
					{From: "A", To: "B", Condition: &ir.Condition{
						Raw:    "ctx.outcome = success",
						Parsed: ir.CondCompare{Variable: "ctx.outcome", Op: "=", Value: "success"},
					}},
					{From: "A", To: "C", Condition: &ir.Condition{
						Raw:    "ctx.outcome = fail",
						Parsed: ir.CondCompare{Variable: "ctx.outcome", Op: "=", Value: "fail"},
					}},
					{From: "B", To: "C"},
				},
			},
			// No DIP103, and success/fail is an exhaustive condition set, so no DIP101/DIP102 either.
			wantNoDiag: true,
		},
		{
			name: "DIP101: exhaustive conditions from raw text (not pre-parsed)",
			workflow: &ir.Workflow{
				Name:  "raw_exhaustive",
				Start: "A",
				Exit:  "C",
				Nodes: []*ir.Node{
					{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "a"}},
					{ID: "B", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "b"}},
					{ID: "C", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "c"}},
				},
				Edges: []*ir.Edge{
					{From: "A", To: "B", Condition: &ir.Condition{Raw: "ctx.outcome = success"}},
					{From: "A", To: "C", Condition: &ir.Condition{Raw: "ctx.outcome = fail"}},
					{From: "B", To: "C"},
				},
			},
			// Lint must parse raw conditions and recognize success/fail as exhaustive.
			wantNoDiag: true,
		},
		{
			name: "DIP101: exhaustive conditions + unconditional fallback (pattern 1)",
			workflow: &ir.Workflow{
				Name:  "exhaustive_with_fallback",
				Start: "Start",
				Exit:  "Done",
				Nodes: []*ir.Node{
					{ID: "Start", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "go"}},
					{ID: "Gate", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "gate"}},
					{ID: "Pass", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "pass"}},
					{ID: "Fix", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "fix"}},
					{ID: "Escalate", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "escalate"}},
					{ID: "Done", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "done"}},
				},
				Edges: []*ir.Edge{
					{From: "Start", To: "Gate"},
					{From: "Gate", To: "Pass", Condition: &ir.Condition{Raw: "ctx.outcome = success"}},
					{From: "Gate", To: "Fix", Condition: &ir.Condition{Raw: "ctx.outcome = fail"}},
					{From: "Gate", To: "Escalate"}, // unconditional fallback
					{From: "Pass", To: "Done"},
					{From: "Fix", To: "Gate"},
					{From: "Escalate", To: "Done"},
				},
			},
			// Gate has exhaustive conditions (success+fail) plus an unconditional edge.
			// Pass and Fix should NOT get DIP101 since they're behind exhaustive conditions.
			wantNoDiag: true,
		},
		{
			name: "DIP101: exhaustive conditions + extra variable conditions (pattern 2)",
			workflow: &ir.Workflow{
				Name:  "exhaustive_with_extra_conds",
				Start: "Start",
				Exit:  "Done",
				Nodes: []*ir.Node{
					{ID: "Start", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "go"}},
					{ID: "Gate", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "gate"}},
					{ID: "Pass", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "pass"}},
					{ID: "Fix", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "fix"}},
					{ID: "Escalate", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "escalate"}},
					{ID: "Done", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "done"}},
				},
				Edges: []*ir.Edge{
					{From: "Start", To: "Gate"},
					{From: "Gate", To: "Pass", Condition: &ir.Condition{Raw: "ctx.outcome = success"}},
					{From: "Gate", To: "Fix", Condition: &ir.Condition{Raw: "ctx.outcome = fail"}},
					{From: "Gate", To: "Escalate", Condition: &ir.Condition{Raw: "ctx.tool_stdout contains escalate"}},
					{From: "Gate", To: "Escalate"}, // unconditional fallback
					{From: "Pass", To: "Done"},
					{From: "Fix", To: "Gate"},
					{From: "Escalate", To: "Done"},
				},
			},
			// Gate has outcome=success + outcome=fail (exhaustive) plus an extra
			// condition on tool_stdout. Pass and Fix should not get DIP101.
			wantNoDiag: true,
		},

		{
			name: "DIP101: complementary contains/not contains pair (pattern 3)",
			workflow: &ir.Workflow{
				Name:  "complementary_pair",
				Start: "Start",
				Exit:  "Done",
				Nodes: []*ir.Node{
					{ID: "Start", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "go"}},
					{ID: "Pick", Kind: ir.NodeTool, Config: ir.ToolConfig{Command: "echo done", Timeout: 60 * 1e9}},
					{ID: "Implement", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "implement"}},
					{ID: "Review", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "review"}},
					{ID: "Done", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "done"}},
				},
				Edges: []*ir.Edge{
					{From: "Start", To: "Pick"},
					{From: "Pick", To: "Implement", Condition: &ir.Condition{Raw: "ctx.tool_stdout not contains all-done"}},
					{From: "Pick", To: "Review", Condition: &ir.Condition{Raw: "ctx.tool_stdout contains all-done"}},
					{From: "Implement", To: "Pick"},
					{From: "Review", To: "Done"},
				},
			},
			// "contains X" + "not contains X" is complementary — no DIP101.
			wantNoDiag: true,
		},
		{
			name: "DIP101: custom binary partition on tool_stdout (done/more_questions)",
			workflow: &ir.Workflow{
				Name:  "custom_partition",
				Start: "Start",
				Exit:  "Done",
				Nodes: []*ir.Node{
					{ID: "Start", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "go"}},
					{ID: "Check", Kind: ir.NodeTool, Config: ir.ToolConfig{Command: "echo done", Timeout: 30e9}},
					{ID: "KeepGoing", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "more"}},
					{ID: "Finalize", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "wrap up"}},
					{ID: "Done", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "done"}},
				},
				Edges: []*ir.Edge{
					{From: "Start", To: "Check"},
					{From: "Check", To: "KeepGoing", Condition: &ir.Condition{Raw: "ctx.tool_stdout = more_questions"}},
					{From: "Check", To: "Finalize", Condition: &ir.Condition{Raw: "ctx.tool_stdout = done"}},
					{From: "KeepGoing", To: "Check"},
					{From: "Finalize", To: "Done"},
				},
			},
			// All conditional edges test the same variable — complete partition.
			wantNoDiag: true,
		},
		{
			name: "DIP101: custom partition with 3 values (tasks_remain/in_progress/all_done)",
			workflow: &ir.Workflow{
				Name:  "ternary_partition",
				Start: "Start",
				Exit:  "Done",
				Nodes: []*ir.Node{
					{ID: "Start", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "go"}},
					{ID: "Check", Kind: ir.NodeTool, Config: ir.ToolConfig{Command: "echo status", Timeout: 30e9}},
					{ID: "Work", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "work"}},
					{ID: "Wait", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "wait"}},
					{ID: "Finish", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "finish"}},
					{ID: "Done", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "done"}},
				},
				Edges: []*ir.Edge{
					{From: "Start", To: "Check"},
					{From: "Check", To: "Work", Condition: &ir.Condition{Raw: "ctx.tool_stdout = tasks_remain"}},
					{From: "Check", To: "Wait", Condition: &ir.Condition{Raw: "ctx.tool_stdout = in_progress"}},
					{From: "Check", To: "Finish", Condition: &ir.Condition{Raw: "ctx.tool_stdout = all_done"}},
					{From: "Work", To: "Check"},
					{From: "Wait", To: "Check"},
					{From: "Finish", To: "Done"},
				},
			},
			wantNoDiag: true,
		},
		{
			name: "DIP101: single conditional edge is NOT a partition",
			workflow: &ir.Workflow{
				Name:  "single_cond",
				Start: "Start",
				Exit:  "Done",
				Nodes: []*ir.Node{
					{ID: "Start", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "go"}},
					{ID: "Check", Kind: ir.NodeTool, Config: ir.ToolConfig{Command: "echo ok", Timeout: 30e9}},
					{ID: "Next", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "next"}},
					{ID: "Done", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "done"}},
				},
				Edges: []*ir.Edge{
					{From: "Start", To: "Check"},
					{From: "Check", To: "Next", Condition: &ir.Condition{Raw: "ctx.tool_stdout = ok"}},
					{From: "Next", To: "Done"},
				},
			},
			// Single conditional edge — not exhaustive, DIP101 should fire.
			wantCodes: []string{DIP101},
		},
		{
			name: "DIP101: mixed unconditional + conditional from same source (pattern 3)",
			workflow: &ir.Workflow{
				Name:  "mixed_routing",
				Start: "Start",
				Exit:  "Done",
				Nodes: []*ir.Node{
					{ID: "Start", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "go"}},
					{ID: "Impl", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "impl"}},
					{ID: "Review", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "review"}},
					{ID: "Debug", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "debug"}},
					{ID: "Done", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "done"}},
				},
				Edges: []*ir.Edge{
					{From: "Start", To: "Impl"},
					{From: "Impl", To: "Review"}, // unconditional
					{From: "Impl", To: "Debug", Condition: &ir.Condition{Raw: "ctx.outcome = fail"}}, // conditional
					{From: "Review", To: "Done"},
					{From: "Debug", To: "Impl"},
				},
			},
			// Source has unconditional edge → conditional branch is intentional.
			wantNoDiag: true,
		},
		{
			name: "DIP101: conditional edge with labeled fallback from source (pattern 4)",
			workflow: &ir.Workflow{
				Name:  "labeled_fallback",
				Start: "Start",
				Exit:  "Done",
				Nodes: []*ir.Node{
					{ID: "Start", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "go"}},
					{ID: "Gate", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "gate"}},
					{ID: "Pass", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "pass"}},
					{ID: "Fix", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "fix"}},
					{ID: "Fallback", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "fallback"}},
					{ID: "Done", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "done"}},
				},
				Edges: []*ir.Edge{
					{From: "Start", To: "Gate"},
					{From: "Gate", To: "Pass", Condition: &ir.Condition{Raw: "ctx.outcome = success"}},
					{From: "Gate", To: "Fix", Condition: &ir.Condition{Raw: "ctx.outcome = fail"}},
					{From: "Gate", To: "Fallback"}, // unconditional labeled fallback
					{From: "Pass", To: "Done"},
					{From: "Fix", To: "Gate"},
					{From: "Fallback", To: "Done"},
				},
			},
			// Source has exhaustive conditions + unconditional fallback.
			wantNoDiag: true,
		},
		{
			name: "DIP101: only conditional edges, source has no unconditional out, NOT safe",
			workflow: &ir.Workflow{
				Name:  "unsafe_conditional",
				Start: "Start",
				Exit:  "Done",
				Nodes: []*ir.Node{
					{ID: "Start", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "go"}},
					{ID: "Gate", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "gate"}},
					{ID: "Only", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "only"}},
					{ID: "Done", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "done"}},
				},
				Edges: []*ir.Edge{
					{From: "Start", To: "Gate"},
					{From: "Gate", To: "Only", Condition: &ir.Condition{Raw: "ctx.outcome = success"}},
					{From: "Only", To: "Done"},
				},
			},
			// Source has only conditional edges, no unconditional out, not exhaustive.
			wantCodes: []string{DIP101},
		},

		// --- DIP104: Unbounded retry ---
		{
			name: "DIP104: retry config but no max_retries or fallback",
			workflow: &ir.Workflow{
				Name:  "unbounded_retry",
				Start: "A",
				Exit:  "A",
				Nodes: []*ir.Node{
					{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "a"},
						Retry: ir.RetryConfig{Policy: "standard"}},
				},
			},
			wantCodes: []string{DIP104},
		},
		{
			name: "DIP104: retry with max_retries is fine",
			workflow: &ir.Workflow{
				Name:  "bounded_retry",
				Start: "A",
				Exit:  "A",
				Nodes: []*ir.Node{
					{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "a"},
						Retry: ir.RetryConfig{Policy: "standard", MaxRetries: 3}},
				},
			},
			wantNoDiag: true,
		},
		{
			name: "DIP104: retry with fallback is fine",
			workflow: &ir.Workflow{
				Name:  "retry_fallback",
				Start: "A",
				Exit:  "B",
				Nodes: []*ir.Node{
					{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "a"},
						Retry: ir.RetryConfig{Policy: "standard", FallbackTarget: "B"}},
					{ID: "B", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "b"}},
				},
				Edges: []*ir.Edge{
					{From: "A", To: "B"},
				},
			},
			wantNoDiag: true,
		},

		// --- DIP105: No success path to exit ---
		{
			name: "DIP105: no forward path to exit",
			workflow: &ir.Workflow{
				Name:  "no_path",
				Start: "A",
				Exit:  "C",
				Nodes: []*ir.Node{
					{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "a"}},
					{ID: "B", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "b"}},
					{ID: "C", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "c"}},
				},
				Edges: []*ir.Edge{
					{From: "A", To: "B"},
					{From: "B", To: "A", Restart: true}, // only restart edges back
				},
			},
			wantCodes: []string{DIP105},
		},
		{
			name: "DIP105: forward path exists even with restart edges",
			workflow: &ir.Workflow{
				Name:  "has_path",
				Start: "A",
				Exit:  "C",
				Nodes: []*ir.Node{
					{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "a"}},
					{ID: "B", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "b"}},
					{ID: "C", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "c"}},
				},
				Edges: []*ir.Edge{
					{From: "A", To: "B"},
					{From: "B", To: "C"},
					{From: "B", To: "A", Restart: true},
				},
			},
			wantNoDiag: true,
		},

		// --- DIP106: Undefined variables in prompts ---
		{
			name: "DIP106: unnamespaced variable reference",
			workflow: &ir.Workflow{
				Name:  "bad_var",
				Start: "A",
				Exit:  "A",
				Nodes: []*ir.Node{
					{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{
						Prompt: "Process ${unknown_var} now.",
					}},
				},
			},
			wantCodes: []string{DIP106},
		},
		{
			name: "DIP106: known namespace is fine",
			workflow: &ir.Workflow{
				Name:  "good_var",
				Start: "A",
				Exit:  "A",
				Nodes: []*ir.Node{
					{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{
						Prompt: "Process ${ctx.data} and ${graph.goal} and ${params.strict}.",
					}},
				},
			},
			wantNoDiag: true,
		},
		{
			name: "DIP106: ctx.node.<existingNode>.* is valid — no DIP106",
			workflow: &ir.Workflow{
				Name:  "node_scoped_ref_valid",
				Start: "Planner",
				Exit:  "Builder",
				Nodes: []*ir.Node{
					{ID: "Planner", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "plan"}},
					{ID: "Builder", Kind: ir.NodeAgent, Config: ir.AgentConfig{
						Prompt: "Build based on ${ctx.node.Planner.last_response}",
					}},
				},
				Edges: []*ir.Edge{{From: "Planner", To: "Builder"}},
			},
			wantNoDiag: true,
		},
		{
			name: "DIP106: ctx.node.<unknownNode>.* triggers DIP106",
			workflow: &ir.Workflow{
				Name:  "node_scoped_ref_invalid",
				Start: "A",
				Exit:  "A",
				Nodes: []*ir.Node{
					{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{
						Prompt: "Use ${ctx.node.Ghost.result}",
					}},
				},
			},
			wantCodes: []string{DIP106},
		},
		{
			name: "DIP106: bare node.<unknownNode>.* triggers DIP106",
			workflow: &ir.Workflow{
				Name:  "bare_node_ref_invalid",
				Start: "A",
				Exit:  "A",
				Nodes: []*ir.Node{
					{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{
						Prompt: "Use ${node.Ghost.result}",
					}},
				},
			},
			wantCodes: []string{DIP106},
		},
		{
			name: "DIP106: bare node.<existingNode>.* is valid — no DIP106",
			workflow: &ir.Workflow{
				Name:  "bare_node_ref_valid",
				Start: "Planner",
				Exit:  "Builder",
				Nodes: []*ir.Node{
					{ID: "Planner", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "plan"}},
					{ID: "Builder", Kind: ir.NodeAgent, Config: ir.AgentConfig{
						Prompt: "Build based on ${node.Planner.last_response}",
					}},
				},
				Edges: []*ir.Edge{{From: "Planner", To: "Builder"}},
			},
			wantNoDiag: true,
		},
		{
			name: "DIP106: malformed node.<id> missing key triggers DIP106",
			workflow: &ir.Workflow{
				Name:  "malformed_node_ref",
				Start: "Planner",
				Exit:  "Builder",
				Nodes: []*ir.Node{
					{ID: "Planner", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "plan"}},
					{ID: "Builder", Kind: ir.NodeAgent, Config: ir.AgentConfig{
						Prompt: "Build based on ${node.Planner}",
					}},
				},
				Edges: []*ir.Edge{{From: "Planner", To: "Builder"}},
			},
			wantCodes: []string{DIP106},
		},

		// --- DIP107: Unused writes ---
		{
			name: "DIP107: writes key that nobody reads",
			workflow: &ir.Workflow{
				Name:  "unused_write",
				Start: "A",
				Exit:  "B",
				Nodes: []*ir.Node{
					{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "a"},
						IO: ir.NodeIO{Writes: []string{"orphan_key"}}},
					{ID: "B", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "b"}},
				},
				Edges: []*ir.Edge{
					{From: "A", To: "B"},
				},
			},
			wantCodes: []string{DIP107},
		},
		{
			name: "DIP107: writes key that is read downstream is fine",
			workflow: &ir.Workflow{
				Name:  "used_write",
				Start: "A",
				Exit:  "B",
				Nodes: []*ir.Node{
					{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "a"},
						IO: ir.NodeIO{Writes: []string{"data"}}},
					{ID: "B", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "b"},
						IO: ir.NodeIO{Reads: []string{"data"}}},
				},
				Edges: []*ir.Edge{
					{From: "A", To: "B"},
				},
			},
			wantNoDiag: true,
		},

		// --- DIP108: Unknown model/provider ---
		{
			name: "DIP108: unknown provider",
			workflow: &ir.Workflow{
				Name:  "bad_provider",
				Start: "A",
				Exit:  "A",
				Nodes: []*ir.Node{
					{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{
						Prompt:   "a",
						Model:    "some-model",
						Provider: "unknown-provider",
					}},
				},
			},
			wantCodes: []string{DIP108},
		},
		{
			name: "DIP108: unknown model for known provider",
			workflow: &ir.Workflow{
				Name:  "bad_model",
				Start: "A",
				Exit:  "A",
				Nodes: []*ir.Node{
					{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{
						Prompt:   "a",
						Model:    "gpt-99",
						Provider: "openai",
					}},
				},
			},
			wantCodes: []string{DIP108},
		},
		{
			name: "DIP108: inherits from defaults",
			workflow: &ir.Workflow{
				Name:  "defaults_model",
				Start: "A",
				Exit:  "A",
				Defaults: ir.WorkflowDefaults{
					Model:    "gpt-99",
					Provider: "openai",
				},
				Nodes: []*ir.Node{
					{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "a"}},
				},
			},
			wantCodes: []string{DIP108},
		},

		// --- DIP109: Namespace collisions in imports ---
		{
			name: "DIP109: two subgraphs referencing same file",
			workflow: &ir.Workflow{
				Name:  "collision",
				Start: "A",
				Exit:  "C",
				Nodes: []*ir.Node{
					{ID: "A", Kind: ir.NodeSubgraph, Config: ir.SubgraphConfig{Ref: "./review.dip"}},
					{ID: "B", Kind: ir.NodeSubgraph, Config: ir.SubgraphConfig{Ref: "./review.dip"}},
					{ID: "C", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "done"}},
				},
				Edges: []*ir.Edge{
					{From: "A", To: "B"},
					{From: "B", To: "C"},
				},
			},
			wantCodes: []string{DIP109},
		},
		{
			name: "DIP109: different subgraph refs is fine",
			workflow: &ir.Workflow{
				Name:  "no_collision",
				Start: "A",
				Exit:  "C",
				Nodes: []*ir.Node{
					{ID: "A", Kind: ir.NodeSubgraph, Config: ir.SubgraphConfig{Ref: "./review.dip"}},
					{ID: "B", Kind: ir.NodeSubgraph, Config: ir.SubgraphConfig{Ref: "./test.dip"}},
					{ID: "C", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "done"}},
				},
				Edges: []*ir.Edge{
					{From: "A", To: "B"},
					{From: "B", To: "C"},
				},
			},
			wantNoDiag: true,
		},

		// --- DIP110: Empty prompts ---
		{
			name: "DIP110: agent with empty prompt",
			workflow: &ir.Workflow{
				Name:  "empty_prompt",
				Start: "Begin",
				Exit:  "End",
				Nodes: []*ir.Node{
					{ID: "Begin", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "start"}},
					{ID: "Mid", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: ""}},
					{ID: "End", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "done"}},
				},
				Edges: []*ir.Edge{
					{From: "Begin", To: "Mid"},
					{From: "Mid", To: "End"},
				},
			},
			wantCodes: []string{DIP110},
		},
		{
			name: "DIP110: agent with whitespace-only prompt",
			workflow: &ir.Workflow{
				Name:  "ws_prompt",
				Start: "Begin",
				Exit:  "End",
				Nodes: []*ir.Node{
					{ID: "Begin", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "start"}},
					{ID: "Mid", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "   \n  "}},
					{ID: "End", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "done"}},
				},
				Edges: []*ir.Edge{
					{From: "Begin", To: "Mid"},
					{From: "Mid", To: "End"},
				},
			},
			wantCodes: []string{DIP110},
		},
		{
			name: "DIP110: start/exit lifecycle nodes exempt from empty prompt",
			workflow: &ir.Workflow{
				Name:  "lifecycle_exempt",
				Start: "Init",
				Exit:  "Done",
				Nodes: []*ir.Node{
					{ID: "Init", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: ""}},
					{ID: "Done", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: ""}},
				},
				Edges: []*ir.Edge{
					{From: "Init", To: "Done"},
				},
			},
			wantNoDiag: true,
		},
		{
			name: "DIP110: non-agent node types do not trigger",
			workflow: &ir.Workflow{
				Name:  "no_prompt_needed",
				Start: "A",
				Exit:  "A",
				Nodes: []*ir.Node{
					{ID: "A", Kind: ir.NodeHuman, Config: ir.HumanConfig{Mode: "freeform"}},
				},
			},
			wantNoDiag: true,
		},

		// --- DIP111: Tool without timeout ---
		{
			name: "DIP111: tool with no timeout",
			workflow: &ir.Workflow{
				Name:  "no_timeout",
				Start: "T",
				Exit:  "T",
				Nodes: []*ir.Node{
					{ID: "T", Kind: ir.NodeTool, Config: ir.ToolConfig{Command: "echo hello"}},
				},
			},
			wantCodes: []string{DIP111},
		},
		{
			name: "DIP111: tool with timeout is clean",
			workflow: &ir.Workflow{
				Name:  "has_timeout",
				Start: "T",
				Exit:  "T",
				Nodes: []*ir.Node{
					{ID: "T", Kind: ir.NodeTool, Config: ir.ToolConfig{
						Command: "echo hello",
						Timeout: 10 * time.Second,
					}},
				},
			},
			wantNoDiag: true,
		},

		// --- DIP112: Reads without upstream writes ---
		{
			name: "DIP112: reads key with no upstream writer",
			workflow: &ir.Workflow{
				Name:  "no_writer",
				Start: "A",
				Exit:  "B",
				Nodes: []*ir.Node{
					{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "a"}},
					{ID: "B", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "b"},
						IO: ir.NodeIO{Reads: []string{"missing_key"}}},
				},
				Edges: []*ir.Edge{
					{From: "A", To: "B"},
				},
			},
			wantCodes: []string{DIP112},
		},
		{
			name: "DIP112: reads key with upstream writer is fine",
			workflow: &ir.Workflow{
				Name:  "has_writer",
				Start: "A",
				Exit:  "B",
				Nodes: []*ir.Node{
					{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "a"},
						IO: ir.NodeIO{Writes: []string{"data"}}},
					{ID: "B", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "b"},
						IO: ir.NodeIO{Reads: []string{"data"}}},
				},
				Edges: []*ir.Edge{
					{From: "A", To: "B"},
				},
			},
			wantNoDiag: true,
		},
		{
			name: "DIP112: node.<existingNode>.* in reads is valid — no DIP112",
			workflow: &ir.Workflow{
				Name:  "node_scoped_read_valid",
				Start: "Planner",
				Exit:  "Builder",
				Nodes: []*ir.Node{
					{ID: "Planner", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "plan"}},
					{ID: "Builder", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "build"},
						IO: ir.NodeIO{Reads: []string{"node.Planner.last_response"}}},
				},
				Edges: []*ir.Edge{
					{From: "Planner", To: "Builder"},
				},
			},
			wantNoDiag: true,
		},
		{
			name: "DIP112: node.<existingNode>.* not upstream triggers DIP112",
			workflow: &ir.Workflow{
				Name:  "node_scoped_read_not_upstream",
				Start: "A",
				Exit:  "B",
				Nodes: []*ir.Node{
					{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "a"}},
					{ID: "B", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "b"},
						// B reads from A but A is not upstream of B (no edge A→B)
						IO: ir.NodeIO{Reads: []string{"node.A.result"}}},
				},
				// No edge from A to B — A is not upstream of B
			},
			wantCodes: []string{DIP112},
		},

		// --- Edge cases ---
		{
			name: "empty workflow only reports DIP105 if start/exit are missing",
			workflow: &ir.Workflow{
				Name: "empty",
			},
			// DIP105 skips if start/exit are empty, DIP110/111/etc have no nodes to check
			wantNoDiag: true,
		},
		{
			name: "multiple lint warnings at once",
			workflow: &ir.Workflow{
				Name:  "multi_warn",
				Start: "Begin",
				Exit:  "End",
				Nodes: []*ir.Node{
					{ID: "Begin", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "go"}},
					{ID: "Mid", Kind: ir.NodeAgent, Config: ir.AgentConfig{
						Prompt: "", // DIP110
					}, Retry: ir.RetryConfig{Policy: "aggressive"}}, // DIP104
					{ID: "End", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "done"}},
				},
				Edges: []*ir.Edge{
					{From: "Begin", To: "Mid"},
					{From: "Mid", To: "End"},
				},
			},
			wantCodes: []string{DIP110, DIP104},
		},
		{
			name: "DIP106: multiple undefined vars in one prompt",
			workflow: &ir.Workflow{
				Name:  "multi_var",
				Start: "A",
				Exit:  "A",
				Nodes: []*ir.Node{
					{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{
						Prompt: "Use ${foo} and ${bar} now.",
					}},
				},
			},
			wantCodes: []string{DIP106, DIP106},
		},
		{
			name: "DIP104: no retry config at all does not trigger",
			workflow: &ir.Workflow{
				Name:  "no_retry",
				Start: "A",
				Exit:  "A",
				Nodes: []*ir.Node{
					{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "a"}},
				},
			},
			wantNoDiag: true,
		},
		{
			name: "DIP112: transitive writes propagation",
			workflow: &ir.Workflow{
				Name:  "transitive",
				Start: "A",
				Exit:  "C",
				Nodes: []*ir.Node{
					{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "a"},
						IO: ir.NodeIO{Writes: []string{"key1"}}},
					{ID: "B", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "b"},
						IO: ir.NodeIO{Reads: []string{"key1"}, Writes: []string{"key2"}}},
					{ID: "C", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "c"},
						IO: ir.NodeIO{Reads: []string{"key1", "key2"}}},
				},
				Edges: []*ir.Edge{
					{From: "A", To: "B"},
					{From: "B", To: "C"},
				},
			},
			wantNoDiag: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Lint(tt.workflow)

			if tt.wantNoDiag {
				if len(result.Diagnostics) != 0 {
					t.Errorf("expected no diagnostics, got %d:", len(result.Diagnostics))
					for _, d := range result.Diagnostics {
						t.Errorf("  %s", d.String())
					}
				}
				return
			}

			if tt.wantCodes != nil {
				gotCodes := make([]string, len(result.Diagnostics))
				for i, d := range result.Diagnostics {
					gotCodes[i] = d.Code
				}

				wantCount := make(map[string]int)
				for _, c := range tt.wantCodes {
					wantCount[c]++
				}
				gotCount := make(map[string]int)
				for _, c := range gotCodes {
					gotCount[c]++
				}
				for code, want := range wantCount {
					if got := gotCount[code]; got < want {
						t.Errorf("expected at least %d %s diagnostic(s), got %d. All codes: %v", want, code, got, gotCodes)
					}
				}
				// Also check total count matches expectations for single-diagnostic tests.
				if len(tt.wantCodes) == 1 {
					codeCount := 0
					for _, d := range result.Diagnostics {
						if d.Code == tt.wantCodes[0] {
							codeCount++
						}
					}
					if codeCount < 1 {
						t.Errorf("expected at least 1 %s diagnostic, got 0. All: %v", tt.wantCodes[0], gotCodes)
					}
				}
			}
		})
	}
}

func TestLintDiagnosticSeverity(t *testing.T) {
	// All lint diagnostics should be warnings, not errors.
	w := &ir.Workflow{
		Name:  "severity_check",
		Start: "A",
		Exit:  "A",
		Nodes: []*ir.Node{
			{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: ""},
				Retry: ir.RetryConfig{Policy: "standard"}},
		},
	}

	result := Lint(w)
	for _, d := range result.Diagnostics {
		if d.Severity != SeverityWarning {
			t.Errorf("lint diagnostic %s has severity %s, want warning", d.Code, d.Severity)
		}
	}
}

func TestLintDIP101MessageContent(t *testing.T) {
	// Source has ONLY conditional outgoing edges (no unconditional) and is not
	// exhaustive → DIP101 fires on the conditional destination.
	w := &ir.Workflow{
		Name:  "msg_check",
		Start: "A",
		Exit:  "C",
		Nodes: []*ir.Node{
			{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "a"}},
			{ID: "B", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "b"}},
			{ID: "C", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "c"}},
		},
		Edges: []*ir.Edge{
			{From: "A", To: "B", Condition: &ir.Condition{
				Raw: "ctx.x = 1",
			}},
			{From: "B", To: "C"},
		},
	}

	result := Lint(w)
	var found bool
	for _, d := range result.Diagnostics {
		if d.Code == DIP101 {
			found = true
			if !strings.Contains(d.Message, `"B"`) {
				t.Errorf("DIP101 message should mention node B, got: %s", d.Message)
			}
			if !strings.Contains(d.Message, "conditional") {
				t.Errorf("DIP101 message should mention 'conditional', got: %s", d.Message)
			}
		}
	}
	if !found {
		t.Error("expected DIP101 diagnostic")
	}
}

func TestLintDIP102MessageContent(t *testing.T) {
	// Single conditional edge — not a complete partition, so DIP102 fires.
	w := &ir.Workflow{
		Name:  "msg_check",
		Start: "A",
		Exit:  "C",
		Nodes: []*ir.Node{
			{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "a"}},
			{ID: "B", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "b"}},
			{ID: "C", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "c"}},
		},
		Edges: []*ir.Edge{
			{From: "A", To: "B", Condition: &ir.Condition{
				Raw: "ctx.x = 1",
			}},
			{From: "B", To: "C"},
		},
	}

	result := Lint(w)
	var found bool
	for _, d := range result.Diagnostics {
		if d.Code == DIP102 {
			found = true
			if !strings.Contains(d.Message, `"A"`) {
				t.Errorf("DIP102 message should mention node A, got: %s", d.Message)
			}
			if !strings.Contains(d.Message, "unconditional") || !strings.Contains(d.Message, "default") {
				t.Errorf("DIP102 message should mention 'unconditional default', got: %s", d.Message)
			}
		}
	}
	if !found {
		t.Error("expected DIP102 diagnostic")
	}
}

func TestLintDIP103OverlappingANDConditions(t *testing.T) {
	// Overlapping condition buried in an AND expression.
	w := &ir.Workflow{
		Name:  "and_overlap",
		Start: "A",
		Exit:  "C",
		Nodes: []*ir.Node{
			{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "a"}},
			{ID: "B", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "b"}},
			{ID: "C", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "c"}},
		},
		Edges: []*ir.Edge{
			{From: "A", To: "B", Condition: &ir.Condition{
				Raw: "ctx.x = 1 and ctx.y = 2",
				Parsed: ir.CondAnd{
					Left:  ir.CondCompare{Variable: "ctx.x", Op: "=", Value: "1"},
					Right: ir.CondCompare{Variable: "ctx.y", Op: "=", Value: "2"},
				},
			}},
			{From: "A", To: "C", Condition: &ir.Condition{
				Raw:    "ctx.x = 1",
				Parsed: ir.CondCompare{Variable: "ctx.x", Op: "=", Value: "1"},
			}},
		},
	}

	result := Lint(w)
	found := false
	for _, d := range result.Diagnostics {
		if d.Code == DIP103 {
			found = true
		}
	}
	if !found {
		t.Error("expected DIP103 for overlapping condition in AND expression")
	}
}

func TestLintDIP105StartEqualsExit(t *testing.T) {
	// When start == exit, trivially reachable.
	w := &ir.Workflow{
		Name:  "trivial",
		Start: "A",
		Exit:  "A",
		Nodes: []*ir.Node{
			{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "a"}},
		},
	}

	result := Lint(w)
	for _, d := range result.Diagnostics {
		if d.Code == DIP105 {
			t.Errorf("DIP105 should not trigger when start == exit, got: %s", d.Message)
		}
	}
}

func TestLintDIP106NoPromptNodes(t *testing.T) {
	// Human and tool nodes should not trigger DIP106.
	w := &ir.Workflow{
		Name:  "no_prompt",
		Start: "A",
		Exit:  "B",
		Nodes: []*ir.Node{
			{ID: "A", Kind: ir.NodeHuman, Config: ir.HumanConfig{Mode: "freeform"}},
			{ID: "B", Kind: ir.NodeTool, Config: ir.ToolConfig{Command: "echo ${not_a_prompt}", Timeout: 10 * time.Second}},
		},
		Edges: []*ir.Edge{
			{From: "A", To: "B"},
		},
	}

	result := Lint(w)
	for _, d := range result.Diagnostics {
		if d.Code == DIP106 {
			t.Errorf("DIP106 should not trigger for non-agent nodes, got: %s", d.Message)
		}
	}
}

func TestLintDIP108NoModelOrProvider(t *testing.T) {
	// If model or provider is not specified (and no defaults), don't check.
	w := &ir.Workflow{
		Name:  "no_model",
		Start: "A",
		Exit:  "A",
		Nodes: []*ir.Node{
			{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "a"}},
		},
	}

	result := Lint(w)
	for _, d := range result.Diagnostics {
		if d.Code == DIP108 {
			t.Errorf("DIP108 should not trigger when model/provider are unset, got: %s", d.Message)
		}
	}
}

func TestRegisterExtraModels(t *testing.T) {
	const provider = "custom-corp"
	const model = "custom-llm-v1"

	// Cleanup: remove the test provider after the test.
	t.Cleanup(func() {
		delete(knownModelProviders, provider)
	})

	w := &ir.Workflow{
		Name:  "custom_model",
		Start: "A",
		Exit:  "A",
		Nodes: []*ir.Node{
			{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{
				Prompt:   "hello",
				Model:    model,
				Provider: provider,
			}},
		},
	}

	// Before registration, DIP108 should fire.
	result := Lint(w)
	hasDIP108 := false
	for _, d := range result.Diagnostics {
		if d.Code == DIP108 {
			hasDIP108 = true
			break
		}
	}
	if !hasDIP108 {
		t.Error("expected DIP108 before RegisterExtraModels, but got none")
	}

	// Register the custom model.
	RegisterExtraModels(provider + ":" + model)

	// After registration, DIP108 should not fire.
	result2 := Lint(w)
	for _, d := range result2.Diagnostics {
		if d.Code == DIP108 {
			t.Errorf("unexpected DIP108 after RegisterExtraModels: %s", d.Message)
		}
	}
}

func TestRegisterExtraModels_MultipleProviders(t *testing.T) {
	t.Cleanup(func() {
		delete(knownModelProviders, "corp-a")
		delete(knownModelProviders, "corp-b")
	})

	RegisterExtraModels("corp-a:model-x,model-y;corp-b:model-z")

	if !knownModelProviders["corp-a"]["model-x"] {
		t.Error("expected corp-a:model-x to be registered")
	}
	if !knownModelProviders["corp-a"]["model-y"] {
		t.Error("expected corp-a:model-y to be registered")
	}
	if !knownModelProviders["corp-b"]["model-z"] {
		t.Error("expected corp-b:model-z to be registered")
	}
}

func TestRegisterExtraModels_EmptyAndMalformed(t *testing.T) {
	// Should not panic or register garbage.
	RegisterExtraModels("")
	RegisterExtraModels(";;")
	RegisterExtraModels("no-colon-here")
	RegisterExtraModels("provider-only:")
}

func TestLintDIP112CycleDoesNotPanic(t *testing.T) {
	// A workflow with a cycle (via non-restart edges) should not cause
	// the topological sort in DIP112 to hang or panic.
	w := &ir.Workflow{
		Name:  "cycle_safe",
		Start: "A",
		Exit:  "C",
		Nodes: []*ir.Node{
			{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "a"},
				IO: ir.NodeIO{Writes: []string{"data"}}},
			{ID: "B", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "b"},
				IO: ir.NodeIO{Reads: []string{"data"}}},
			{ID: "C", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "c"}},
		},
		Edges: []*ir.Edge{
			{From: "A", To: "B"},
			{From: "B", To: "A"}, // cycle
			{From: "B", To: "C"},
		},
	}

	// Should not panic.
	result := Lint(w)
	_ = result
}

func TestLintCodeDescriptionCoverage(t *testing.T) {
	codes := []string{DIP101, DIP102, DIP103, DIP104, DIP105, DIP106, DIP107, DIP108, DIP109, DIP110, DIP111, DIP112}
	for _, c := range codes {
		if desc, ok := CodeDescription[c]; !ok || desc == "" {
			t.Errorf("CodeDescription[%q] is missing or empty", c)
		}
	}
}

func TestExtractComparisons(t *testing.T) {
	tests := []struct {
		name string
		expr ir.ConditionExpr
		want int // expected number of comparisons
	}{
		{
			name: "nil",
			expr: nil,
			want: 0,
		},
		{
			name: "single compare",
			expr: ir.CondCompare{Variable: "ctx.x", Op: "=", Value: "1"},
			want: 1,
		},
		{
			name: "AND of two compares",
			expr: ir.CondAnd{
				Left:  ir.CondCompare{Variable: "ctx.x", Op: "=", Value: "1"},
				Right: ir.CondCompare{Variable: "ctx.y", Op: "=", Value: "2"},
			},
			want: 2,
		},
		{
			name: "OR of two compares",
			expr: ir.CondOr{
				Left:  ir.CondCompare{Variable: "ctx.x", Op: "=", Value: "1"},
				Right: ir.CondCompare{Variable: "ctx.y", Op: "=", Value: "2"},
			},
			want: 2,
		},
		{
			name: "NOT of compare",
			expr: ir.CondNot{Inner: ir.CondCompare{Variable: "ctx.x", Op: "=", Value: "1"}},
			want: 1,
		},
		{
			name: "nested AND/OR/NOT",
			expr: ir.CondAnd{
				Left: ir.CondOr{
					Left:  ir.CondCompare{Variable: "ctx.a", Op: "=", Value: "1"},
					Right: ir.CondCompare{Variable: "ctx.b", Op: "=", Value: "2"},
				},
				Right: ir.CondNot{
					Inner: ir.CondCompare{Variable: "ctx.c", Op: "!=", Value: "3"},
				},
			},
			want: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractComparisons(tt.expr)
			if len(got) != tt.want {
				t.Errorf("extractComparisons() returned %d comparisons, want %d", len(got), tt.want)
			}
		})
	}
}

// --- DIP113: Invalid retry policy ---

func TestLint_DIP113_InvalidRetryPolicy_Node(t *testing.T) {
	w := cleanMinimalWorkflow()
	w.Nodes[0].Retry.Policy = "bogus"
	res := Lint(w)
	assertHasCode(t, res, DIP113)
}

func TestLint_DIP113_InvalidRetryPolicy_Default(t *testing.T) {
	w := cleanMinimalWorkflow()
	w.Defaults.RetryPolicy = "nope"
	res := Lint(w)
	assertHasCode(t, res, DIP113)
}

func TestLint_DIP113_ValidRetryPolicies(t *testing.T) {
	for _, policy := range []string{"standard", "aggressive", "patient", "linear", "none"} {
		t.Run(policy, func(t *testing.T) {
			w := cleanMinimalWorkflow()
			w.Nodes[0].Retry.Policy = policy
			res := Lint(w)
			assertNoCode(t, res, DIP113)
		})
	}
}

func TestLint_DIP113_EmptyPolicy_NoDiag(t *testing.T) {
	w := cleanMinimalWorkflow()
	// No retry policy set — should not warn.
	res := Lint(w)
	assertNoCode(t, res, DIP113)
}

// --- DIP114: Invalid fidelity level ---

func TestLint_DIP114_InvalidFidelity_Node(t *testing.T) {
	w := cleanMinimalWorkflow()
	w.Nodes[0].Config = ir.AgentConfig{Prompt: "X", Fidelity: "sumary:high"}
	res := Lint(w)
	assertHasCode(t, res, DIP114)
}

func TestLint_DIP114_InvalidFidelity_Default(t *testing.T) {
	w := cleanMinimalWorkflow()
	w.Defaults.Fidelity = "hi"
	res := Lint(w)
	assertHasCode(t, res, DIP114)
}

func TestLint_DIP114_ValidFidelityLevels(t *testing.T) {
	for _, level := range []string{"full", "summary:high", "summary:medium", "summary:low", "compact", "truncate"} {
		t.Run(level, func(t *testing.T) {
			w := cleanMinimalWorkflow()
			w.Nodes[0].Config = ir.AgentConfig{Prompt: "X", Fidelity: level}
			res := Lint(w)
			assertNoCode(t, res, DIP114)
		})
	}
}

func TestLint_DIP114_EmptyFidelity_NoDiag(t *testing.T) {
	w := cleanMinimalWorkflow()
	res := Lint(w)
	assertNoCode(t, res, DIP114)
}

// --- DIP115: Goal gate without fallback ---

func TestLint_DIP115_GoalGateNoFallback(t *testing.T) {
	w := cleanMinimalWorkflow()
	w.Nodes[0].Config = ir.AgentConfig{Prompt: "X", GoalGate: true}
	res := Lint(w)
	assertHasCode(t, res, DIP115)
}

func TestLint_DIP115_GoalGateWithRetryTarget(t *testing.T) {
	w := cleanMinimalWorkflow()
	w.Nodes[0].Config = ir.AgentConfig{Prompt: "X", GoalGate: true}
	w.Nodes[0].Retry.RetryTarget = "Begin"
	res := Lint(w)
	assertNoCode(t, res, DIP115)
}

func TestLint_DIP115_GoalGateWithFallbackTarget(t *testing.T) {
	w := cleanMinimalWorkflow()
	w.Nodes[0].Config = ir.AgentConfig{Prompt: "X", GoalGate: true}
	w.Nodes[0].Retry.FallbackTarget = "End"
	res := Lint(w)
	assertNoCode(t, res, DIP115)
}

func TestLint_DIP115_NoGoalGate_NoDiag(t *testing.T) {
	w := cleanMinimalWorkflow()
	res := Lint(w)
	assertNoCode(t, res, DIP115)
}

// --- DIP120: condition variable namespace ---

func TestLint_DIP120_BareVariable(t *testing.T) {
	w := cleanMinimalWorkflow()
	w.Edges[0].Condition = &ir.Condition{
		Raw:    "outcome = success",
		Parsed: ir.CondCompare{Variable: "outcome", Op: "=", Value: "success"},
	}
	res := Lint(w)
	assertHasCode(t, res, DIP120)
}

func TestLint_DIP120_NamespacedVariable_NoDiag(t *testing.T) {
	w := cleanMinimalWorkflow()
	w.Edges[0].Condition = &ir.Condition{
		Raw:    "ctx.outcome = success",
		Parsed: ir.CondCompare{Variable: "ctx.outcome", Op: "=", Value: "success"},
	}
	res := Lint(w)
	assertNoCode(t, res, DIP120)
}

func TestLint_DIP120_GraphNamespace_NoDiag(t *testing.T) {
	w := cleanMinimalWorkflow()
	w.Edges[0].Condition = &ir.Condition{
		Raw:    "graph.goal = build",
		Parsed: ir.CondCompare{Variable: "graph.goal", Op: "=", Value: "build"},
	}
	res := Lint(w)
	assertNoCode(t, res, DIP120)
}

func TestLint_DIP120_BareVariableInAnd(t *testing.T) {
	w := cleanMinimalWorkflow()
	w.Edges[0].Condition = &ir.Condition{
		Raw: "outcome = success and tool_stdout != empty",
		Parsed: ir.CondAnd{
			Left:  ir.CondCompare{Variable: "outcome", Op: "=", Value: "success"},
			Right: ir.CondCompare{Variable: "tool_stdout", Op: "!=", Value: "empty"},
		},
	}
	res := Lint(w)
	assertHasCode(t, res, DIP120)
}

func TestLint_DIP120_NoCondition_NoDiag(t *testing.T) {
	w := cleanMinimalWorkflow()
	res := Lint(w)
	assertNoCode(t, res, DIP120)
}

// --- DIP114: BranchConfig fidelity ---

func TestLint_DIP114_InvalidBranchFidelity(t *testing.T) {
	w := cleanMinimalWorkflow()
	w.Nodes = append(w.Nodes, &ir.Node{
		ID: "Fan", Kind: ir.NodeParallel, Config: ir.ParallelConfig{
			Branches: []ir.BranchConfig{
				{Target: "Begin", Fidelity: "bogus"},
			},
		},
	})
	res := Lint(w)
	assertHasCode(t, res, DIP114)
}

func TestLint_DIP114_ValidBranchFidelity(t *testing.T) {
	w := cleanMinimalWorkflow()
	w.Nodes = append(w.Nodes, &ir.Node{
		ID: "Fan", Kind: ir.NodeParallel, Config: ir.ParallelConfig{
			Branches: []ir.BranchConfig{
				{Target: "Begin", Fidelity: "compact"},
			},
		},
	})
	res := Lint(w)
	assertNoCode(t, res, DIP114)
}

// --- DIP119: reasoning_effort ---

func TestLint_DIP119_InvalidReasoningEffort(t *testing.T) {
	w := cleanMinimalWorkflow()
	w.Nodes[0].Config = ir.AgentConfig{Prompt: "Go.", ReasoningEffort: "hihg"}
	res := Lint(w)
	assertHasCode(t, res, DIP119)
}

func TestLint_DIP119_ValidReasoningEffort(t *testing.T) {
	for _, level := range []string{"none", "minimal", "low", "medium", "high", "xhigh", "max"} {
		w := cleanMinimalWorkflow()
		w.Nodes[0].Config = ir.AgentConfig{Prompt: "Go.", ReasoningEffort: level}
		res := Lint(w)
		assertNoCode(t, res, DIP119)
	}
}

func TestLint_DIP119_EmptyReasoningEffort_NoDiag(t *testing.T) {
	w := cleanMinimalWorkflow()
	res := Lint(w)
	assertNoCode(t, res, DIP119)
}

func TestLint_DIP127_InvalidHumanMode(t *testing.T) {
	w := cleanMinimalWorkflow()
	w.Nodes[0].Kind = ir.NodeHuman
	w.Nodes[0].Config = ir.HumanConfig{Mode: "invalid"}
	res := Lint(w)
	assertHasCode(t, res, DIP127)
}

func TestLint_DIP127_ValidModes(t *testing.T) {
	for _, mode := range []string{"choice", "freeform", "interview", "yes_no", ""} {
		w := cleanMinimalWorkflow()
		w.Nodes[0].Kind = ir.NodeHuman
		w.Nodes[0].Config = ir.HumanConfig{Mode: mode}
		res := Lint(w)
		assertNoCode(t, res, DIP127)
	}
}

func TestLint_DIP128_InterviewWithDefault(t *testing.T) {
	w := cleanMinimalWorkflow()
	w.Nodes[0].Kind = ir.NodeHuman
	w.Nodes[0].Config = ir.HumanConfig{Mode: "interview", Default: "yes"}
	res := Lint(w)
	assertHasCode(t, res, DIP128)
}

func TestLint_DIP128_ChoiceWithDefault_NoDiag(t *testing.T) {
	w := cleanMinimalWorkflow()
	w.Nodes[0].Kind = ir.NodeHuman
	w.Nodes[0].Config = ir.HumanConfig{Mode: "choice", Default: "yes"}
	res := Lint(w)
	assertNoCode(t, res, DIP128)
}

func TestLint_DIP129_InterviewWithLabeledEdges(t *testing.T) {
	w := cleanMinimalWorkflow()
	w.Nodes[0].Kind = ir.NodeHuman
	w.Nodes[0].Config = ir.HumanConfig{Mode: "interview"}
	w.Nodes = append(w.Nodes, &ir.Node{
		ID: "Alt", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "Alt."},
	})
	w.Edges = []*ir.Edge{
		{From: w.Nodes[0].ID, To: w.Exit, Label: "approve"},
		{From: w.Nodes[0].ID, To: "Alt", Label: "reject"},
		{From: "Alt", To: w.Exit},
	}
	res := Lint(w)
	assertHasCode(t, res, DIP129)
}

func TestLint_DIP129_InterviewSingleEdge_NoDiag(t *testing.T) {
	w := cleanMinimalWorkflow()
	w.Nodes[0].Kind = ir.NodeHuman
	w.Nodes[0].Config = ir.HumanConfig{Mode: "interview"}
	res := Lint(w)
	assertNoCode(t, res, DIP129)
}

// assertHasCode checks that a result contains at least one diagnostic with the given code.
func assertHasCode(t *testing.T, res Result, code string) {
	t.Helper()
	for _, d := range res.Diagnostics {
		if d.Code == code {
			return
		}
	}
	codes := make([]string, len(res.Diagnostics))
	for i, d := range res.Diagnostics {
		codes[i] = d.Code
	}
	t.Errorf("expected diagnostic %s, got codes: %v", code, codes)
}

// assertNoCode checks that a result does not contain any diagnostic with the given code.
func assertNoCode(t *testing.T, res Result, code string) {
	t.Helper()
	for _, d := range res.Diagnostics {
		if d.Code == code {
			t.Errorf("unexpected diagnostic %s: %s", code, d.Message)
		}
	}
}

// --- DIP121: Condition references variable not in source node writes ---

func TestLint_DIP121_UndeclaredVariable(t *testing.T) {
	w := &ir.Workflow{
		Name:  "dip121_warn",
		Start: "A",
		Exit:  "B",
		Nodes: []*ir.Node{
			{ID: "A", Kind: ir.NodeTool, Config: ir.ToolConfig{
				Command: "echo ok", Timeout: 30e9,
			}, IO: ir.NodeIO{Writes: []string{"result"}}},
			{ID: "B", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "done"}},
		},
		Edges: []*ir.Edge{
			{From: "A", To: "B", Condition: &ir.Condition{Raw: "ctx.score = high"}},
		},
	}
	res := Lint(w)
	assertHasCode(t, res, DIP121)
}

func TestLint_DIP121_EmptyWrites_NoWarning(t *testing.T) {
	w := &ir.Workflow{
		Name:  "dip121_skip",
		Start: "A",
		Exit:  "B",
		Nodes: []*ir.Node{
			{ID: "A", Kind: ir.NodeTool, Config: ir.ToolConfig{
				Command: "echo ok", Timeout: 30e9,
			}},
			{ID: "B", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "done"}},
		},
		Edges: []*ir.Edge{
			{From: "A", To: "B", Condition: &ir.Condition{Raw: "ctx.score = high"}},
		},
	}
	res := Lint(w)
	assertNoCode(t, res, DIP121)
}

func TestLint_DIP121_ReservedVariable_NoWarning(t *testing.T) {
	w := &ir.Workflow{
		Name:  "dip121_reserved",
		Start: "A",
		Exit:  "B",
		Nodes: []*ir.Node{
			{ID: "A", Kind: ir.NodeTool, Config: ir.ToolConfig{
				Command: "echo ok", Timeout: 30e9,
			}, IO: ir.NodeIO{Writes: []string{"result"}}},
			{ID: "B", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "done"}},
		},
		Edges: []*ir.Edge{
			{From: "A", To: "B", Condition: &ir.Condition{Raw: "ctx.outcome = success"}},
		},
	}
	res := Lint(w)
	assertNoCode(t, res, DIP121)
}

func TestLint_DIP121_DeclaredWrite_NoWarning(t *testing.T) {
	w := &ir.Workflow{
		Name:  "dip121_declared",
		Start: "A",
		Exit:  "B",
		Nodes: []*ir.Node{
			{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "go"},
				IO: ir.NodeIO{Writes: []string{"score"}}},
			{ID: "B", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "done"}},
		},
		Edges: []*ir.Edge{
			{From: "A", To: "B", Condition: &ir.Condition{Raw: "ctx.score = high"}},
		},
	}
	res := Lint(w)
	assertNoCode(t, res, DIP121)
}

func TestLint_DIP121_CompoundCondition(t *testing.T) {
	w := &ir.Workflow{
		Name:  "dip121_compound",
		Start: "A",
		Exit:  "B",
		Nodes: []*ir.Node{
			{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "a"},
				IO: ir.NodeIO{Writes: []string{"score"}}},
			{ID: "B", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "b"}},
		},
		Edges: []*ir.Edge{
			{From: "A", To: "B", Condition: &ir.Condition{
				Raw: "ctx.score = high and ctx.mood = happy",
			}},
		},
	}
	res := Lint(w)
	// score IS in writes, so no DIP121 for it.
	// mood is NOT in writes → should get DIP121.
	assertHasCode(t, res, DIP121)
	// Verify the warning is specifically about "mood", not "score"
	found := false
	for _, d := range res.Diagnostics {
		if d.Code == DIP121 && strings.Contains(d.Message, "mood") {
			found = true
		}
	}
	if !found {
		t.Error("expected DIP121 warning about 'mood' variable")
	}
	// Verify no DIP121 for "score" (which IS declared in writes)
	for _, d := range res.Diagnostics {
		if d.Code == DIP121 && strings.Contains(d.Message, "score") {
			t.Error("should not warn about 'score' which is in writes")
		}
	}
}

// --- DIP122: Condition tests value not in source tool outputs ---

func TestLint_DIP122_UndeclaredValue(t *testing.T) {
	w := &ir.Workflow{
		Name:  "dip122_warn",
		Start: "T",
		Exit:  "B",
		Nodes: []*ir.Node{
			{ID: "T", Kind: ir.NodeTool, Config: ir.ToolConfig{
				Command: "check", Timeout: 30e9,
				Outputs: []string{"success", "fail"},
			}},
			{ID: "B", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "done"}},
		},
		Edges: []*ir.Edge{
			{From: "T", To: "B", Condition: &ir.Condition{Raw: "ctx.outcome = retry"}},
		},
	}
	res := Lint(w)
	assertHasCode(t, res, DIP122)
}

func TestLint_DIP122_EmptyOutputs_NoWarning(t *testing.T) {
	w := &ir.Workflow{
		Name:  "dip122_skip",
		Start: "T",
		Exit:  "B",
		Nodes: []*ir.Node{
			{ID: "T", Kind: ir.NodeTool, Config: ir.ToolConfig{
				Command: "check", Timeout: 30e9,
			}},
			{ID: "B", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "done"}},
		},
		Edges: []*ir.Edge{
			{From: "T", To: "B", Condition: &ir.Condition{Raw: "ctx.outcome = retry"}},
		},
	}
	res := Lint(w)
	assertNoCode(t, res, DIP122)
}

func TestLint_DIP122_AgentNode_NoWarning(t *testing.T) {
	w := &ir.Workflow{
		Name:  "dip122_agent",
		Start: "A",
		Exit:  "B",
		Nodes: []*ir.Node{
			{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "go"}},
			{ID: "B", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "done"}},
		},
		Edges: []*ir.Edge{
			{From: "A", To: "B", Condition: &ir.Condition{Raw: "ctx.outcome = retry"}},
		},
	}
	res := Lint(w)
	assertNoCode(t, res, DIP122)
}

func TestLint_DIP122_DeclaredOutput_NoWarning(t *testing.T) {
	w := &ir.Workflow{
		Name:  "dip122_declared",
		Start: "T",
		Exit:  "B",
		Nodes: []*ir.Node{
			{ID: "T", Kind: ir.NodeTool, Config: ir.ToolConfig{
				Command: "check", Timeout: 30e9,
				Outputs: []string{"success", "fail"},
			}},
			{ID: "B", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "done"}},
		},
		Edges: []*ir.Edge{
			{From: "T", To: "B", Condition: &ir.Condition{Raw: "ctx.outcome = success"}},
		},
	}
	res := Lint(w)
	assertNoCode(t, res, DIP122)
}

// --- DIP130: invalid response_format value ---

func TestLintResponseFormatInvalid(t *testing.T) {
	w := &ir.Workflow{
		Name:  "dip130_invalid",
		Start: "A",
		Exit:  "A",
		Nodes: []*ir.Node{
			{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{
				Prompt:         "Hello.",
				ResponseFormat: "xml",
			}},
		},
	}
	res := Lint(w)
	assertHasCode(t, res, DIP130)
}

func TestLintResponseFormatValid(t *testing.T) {
	w := &ir.Workflow{
		Name:  "dip130_valid",
		Start: "A",
		Exit:  "A",
		Nodes: []*ir.Node{
			{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{
				Prompt:         "Hello.",
				ResponseFormat: "json_object",
			}},
		},
	}
	res := Lint(w)
	assertNoCode(t, res, DIP130)
}

// --- DIP131: response_schema/response_format mismatch ---

func TestLintResponseSchemaWithoutJsonSchema(t *testing.T) {
	w := &ir.Workflow{
		Name:  "dip131_schema_no_format",
		Start: "A",
		Exit:  "A",
		Nodes: []*ir.Node{
			{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{
				Prompt:         "Hello.",
				ResponseFormat: "json_object",
				ResponseSchema: `{"type":"object"}`,
			}},
		},
	}
	res := Lint(w)
	found := false
	for _, d := range res.Diagnostics {
		if d.Code == DIP131 && d.Severity == SeverityWarning {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected DIP131 warning, got: %v", res.Diagnostics)
	}
}

func TestLintJsonSchemaWithoutResponseSchema(t *testing.T) {
	w := &ir.Workflow{
		Name:  "dip131_format_no_schema",
		Start: "A",
		Exit:  "A",
		Nodes: []*ir.Node{
			{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{
				Prompt:         "Hello.",
				ResponseFormat: "json_schema",
			}},
		},
	}
	res := Lint(w)
	found := false
	for _, d := range res.Diagnostics {
		if d.Code == DIP131 && d.Severity == SeverityHint {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected DIP131 hint, got: %v", res.Diagnostics)
	}
}

// --- DIP132: response_schema is not valid JSON ---

func TestLintResponseSchemaInvalidJSON(t *testing.T) {
	w := &ir.Workflow{
		Name:  "dip132_invalid_json",
		Start: "A",
		Exit:  "A",
		Nodes: []*ir.Node{
			{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{
				Prompt:         "Hello.",
				ResponseFormat: "json_schema",
				ResponseSchema: `{not valid json`,
			}},
		},
	}
	res := Lint(w)
	assertHasCode(t, res, DIP132)
}

func TestLintResponseSchemaValidJSON(t *testing.T) {
	w := &ir.Workflow{
		Name:  "dip132_valid_json",
		Start: "A",
		Exit:  "A",
		Nodes: []*ir.Node{
			{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{
				Prompt:         "Hello.",
				ResponseFormat: "json_schema",
				ResponseSchema: `{"type":"object","properties":{"name":{"type":"string"}}}`,
			}},
		},
	}
	res := Lint(w)
	assertNoCode(t, res, DIP132)
}

// --- DIP133: agent params key shadows a first-class field ---

func TestLintParamsShadowsField(t *testing.T) {
	w := &ir.Workflow{
		Name:  "dip133_shadow",
		Start: "A",
		Exit:  "A",
		Nodes: []*ir.Node{
			{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{
				Prompt: "Hello.",
				Params: map[string]string{"model": "gpt-4"},
			}},
		},
	}
	res := Lint(w)
	assertHasCode(t, res, DIP133)
}

func TestLintParamsNoShadow(t *testing.T) {
	w := &ir.Workflow{
		Name:  "dip133_no_shadow",
		Start: "A",
		Exit:  "A",
		Nodes: []*ir.Node{
			{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{
				Prompt: "Hello.",
				Params: map[string]string{"custom_key": "value", "another_key": "data"},
			}},
		},
	}
	res := Lint(w)
	assertNoCode(t, res, DIP133)
}

// --- DIP131: response_schema with response_format absent entirely ---

func TestLintResponseSchemaFormatAbsent(t *testing.T) {
	w := &ir.Workflow{
		Name:  "dip131_format_absent",
		Start: "A",
		Exit:  "A",
		Nodes: []*ir.Node{
			{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{
				Prompt:         "Hello.",
				ResponseSchema: `{"type": "object"}`,
				// ResponseFormat intentionally empty
			}},
		},
	}
	res := Lint(w)
	assertHasCode(t, res, DIP131)
	for _, d := range res.Diagnostics {
		if d.Code == DIP131 && d.Severity != SeverityWarning {
			t.Errorf("DIP131 for absent format should be warning, got %s", d.Severity)
		}
	}
}

// --- All structured output fields valid together ---

func TestLintAllStructuredFieldsValidNoWarnings(t *testing.T) {
	w := &ir.Workflow{
		Name:  "all_structured_valid",
		Start: "A",
		Exit:  "A",
		Nodes: []*ir.Node{
			{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{
				Prompt:         "Return structured output.",
				ResponseFormat: "json_schema",
				ResponseSchema: `{"type":"object","properties":{"result":{"type":"string"}}}`,
				Params:         map[string]string{"backend": "claude-code"},
			}},
		},
	}
	res := Lint(w)
	assertNoCode(t, res, DIP130)
	assertNoCode(t, res, DIP131)
	assertNoCode(t, res, DIP132)
	assertNoCode(t, res, DIP133)
}

func TestLintRetryRestartConfusion(t *testing.T) {
	buildWorkflow := func(maxRetries, maxRestarts int, hasRestart bool) *ir.Workflow {
		edges := []*ir.Edge{{From: "A", To: "B"}}
		if hasRestart {
			edges = append(edges, &ir.Edge{From: "B", To: "A", Restart: true})
		}
		return &ir.Workflow{
			Start: "A",
			Exit:  "B",
			Defaults: ir.WorkflowDefaults{
				MaxRetries:  maxRetries,
				MaxRestarts: maxRestarts,
			},
			Nodes: []*ir.Node{
				{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "a"}},
				{ID: "B", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "b"}},
			},
			Edges: edges,
		}
	}

	t.Run("fires when max_retries set with restart edges and no max_restarts", func(t *testing.T) {
		assertHasCode(t, Lint(buildWorkflow(5, 0, true)), DIP134)
	})
	t.Run("silent when max_restarts also set", func(t *testing.T) {
		assertNoCode(t, Lint(buildWorkflow(5, 10, true)), DIP134)
	})
	t.Run("silent when no max_retries", func(t *testing.T) {
		assertNoCode(t, Lint(buildWorkflow(0, 0, true)), DIP134)
	})
	t.Run("silent when no restart edges", func(t *testing.T) {
		assertNoCode(t, Lint(buildWorkflow(5, 0, false)), DIP134)
	})
}

func TestDIP139_BranchInvalidToolAccess(t *testing.T) {
	w := &ir.Workflow{
		Nodes: []*ir.Node{
			{ID: "split", Kind: ir.NodeParallel, Config: ir.ParallelConfig{
				Targets: []string{"a", "b"},
				Branches: []ir.BranchConfig{
					{Target: "a", ToolAccess: "none"}, // valid
					{Target: "b", ToolAccess: "nono"}, // invalid -> DIP139
				},
			}},
		},
	}
	diags := lintToolAccessValues(w)
	if len(diags) != 1 {
		t.Fatalf("got %d diagnostics, want 1; %+v", len(diags), diags)
	}
	if diags[0].Code != DIP139 {
		t.Errorf("code = %s, want DIP139", diags[0].Code)
	}
	if !strings.Contains(diags[0].Message, `branch "b"`) {
		t.Errorf("message %q should name branch \"b\"", diags[0].Message)
	}
}

func TestDIP139_BranchValidToolAccessSilent(t *testing.T) {
	w := &ir.Workflow{
		Nodes: []*ir.Node{
			{ID: "split", Kind: ir.NodeParallel, Config: ir.ParallelConfig{
				Targets: []string{"a", "b"},
				Branches: []ir.BranchConfig{
					{Target: "a", ToolAccess: "none"},
					{Target: "b"}, // empty -> inherit, silent
				},
			}},
		},
	}
	if diags := lintToolAccessValues(w); len(diags) != 0 {
		t.Errorf("got %d diagnostics, want 0; %+v", len(diags), diags)
	}
}

// TestLintEmptyPrompts_PromptFileSuppresses verifies that an agent with
// prompt_file: set does not trigger DIP110, even if Prompt is empty.
// Regression test for issue #65: parser-pure design means PromptFile
// is set but Prompt is empty until the resolver runs.
func TestLintEmptyPrompts_PromptFileSuppresses(t *testing.T) {
	w := &ir.Workflow{
		Name:  "W",
		Start: "Start",
		Exit:  "End",
		Nodes: []*ir.Node{
			{ID: "Start", Kind: ir.NodeAgent, Config: ir.AgentConfig{}},
			{ID: "Middle", Kind: ir.NodeAgent, Config: ir.AgentConfig{
				PromptFile: "prompts/task.md",
			}},
			{ID: "End", Kind: ir.NodeAgent, Config: ir.AgentConfig{}},
		},
	}
	result := Lint(w)
	for _, d := range result.Diagnostics {
		if d.Code == DIP110 && strings.Contains(d.Message, "Middle") {
			t.Errorf("DIP110 fired on agent with prompt_file: set; diag: %+v", d)
		}
	}
}
