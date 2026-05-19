package validator

// Diagnostic codes for semantic quality warnings (DIP101–DIP138).
const (
	DIP101 = "DIP101" // unreachable nodes after conditional branches
	DIP102 = "DIP102" // routing node without default/unconditional edge
	DIP103 = "DIP103" // overlapping or contradictory conditions
	DIP104 = "DIP104" // unbounded retry loop
	DIP105 = "DIP105" // no success path to exit
	DIP106 = "DIP106" // undefined variable in prompt
	DIP107 = "DIP107" // unused context key (written but never read)
	DIP108 = "DIP108" // unknown model/provider combination
	DIP109 = "DIP109" // namespace collision in imports
	DIP110 = "DIP110" // empty prompt on agent node
	DIP111 = "DIP111" // tool command without timeout
	DIP112 = "DIP112" // reads key not in any upstream writes
	DIP113 = "DIP113" // invalid retry policy name
	DIP114 = "DIP114" // invalid fidelity level
	DIP115 = "DIP115" // goal_gate without retry/fallback target
	DIP116 = "DIP116" // invalid compaction threshold or on_resume value
	DIP117 = "DIP117" // stylesheet .class references undefined class
	DIP118 = "DIP118" // stylesheet #id references unknown node
	DIP119 = "DIP119" // invalid reasoning_effort value
	DIP120 = "DIP120" // condition variable missing namespace prefix
	DIP121 = "DIP121" // condition references variable not in source node writes
	DIP122 = "DIP122" // condition tests value not in source tool outputs
	DIP123 = "DIP123" // tool command has shell syntax errors
	DIP124 = "DIP124" // tool command references runtime-only ${ctx.*} variable
	DIP125 = "DIP125" // tool command binary not found on PATH
	DIP126 = "DIP126" // subgraph ref file does not exist
	DIP127 = "DIP127" // invalid human node mode
	DIP128 = "DIP128" // interview mode with meaningless default
	DIP129 = "DIP129" // interview mode with conflicting choice-style edges
	DIP130 = "DIP130" // invalid response_format value or on non-agent node
	DIP131 = "DIP131" // response_schema/response_format mismatch
	DIP132 = "DIP132" // response_schema is not valid JSON
	DIP133 = "DIP133" // agent params key shadows a first-class field
	DIP134 = "DIP134" // max_retries set with restart edges but no max_restarts
	DIP135 = "DIP135" // manager_loop subgraph_ref missing or file does not exist
	DIP136 = "DIP136" // manager_loop control field has invalid value (poll_interval or max_cycles)
	DIP137 = "DIP137" // unbounded manager_loop: no stop_condition and no max_cycles
	DIP138 = "DIP138" // tool node routes on stdout but declares no marker_grep / outputs (reserved)
)

func init() {
	// Extend CodeDescription with linter codes.
	CodeDescription[DIP101] = "unreachable node after conditional branches"
	CodeDescription[DIP102] = "routing node has no default/unconditional edge"
	CodeDescription[DIP103] = "overlapping or contradictory conditions"
	CodeDescription[DIP104] = "unbounded retry loop (no max_retries or fallback)"
	CodeDescription[DIP105] = "no success path from start to exit"
	CodeDescription[DIP106] = "undefined variable reference in prompt"
	CodeDescription[DIP107] = "unused context key (written but never read)"
	CodeDescription[DIP108] = "unknown model/provider combination"
	CodeDescription[DIP109] = "namespace collision in imports"
	CodeDescription[DIP110] = "empty prompt on agent node"
	CodeDescription[DIP111] = "tool command has no timeout"
	CodeDescription[DIP112] = "reads key not produced by any upstream writes"
	CodeDescription[DIP113] = "invalid retry policy name"
	CodeDescription[DIP114] = "invalid fidelity level"
	CodeDescription[DIP115] = "goal_gate node without retry or fallback target"
	CodeDescription[DIP116] = "invalid compaction threshold or on_resume value"
	CodeDescription[DIP117] = "stylesheet class references undefined class"
	CodeDescription[DIP118] = "stylesheet ID references unknown node"
	CodeDescription[DIP119] = "invalid reasoning_effort value"
	CodeDescription[DIP120] = "condition variable missing namespace prefix"
	CodeDescription[DIP121] = "condition references variable not in source node writes"
	CodeDescription[DIP122] = "condition tests value not in source tool outputs"
	CodeDescription[DIP123] = "tool command has shell syntax errors"
	CodeDescription[DIP124] = "tool command references runtime-only ${ctx.*} variable"
	CodeDescription[DIP125] = "tool command binary not found on PATH"
	CodeDescription[DIP126] = "subgraph ref file does not exist"
	CodeDescription[DIP127] = "invalid human node mode"
	CodeDescription[DIP128] = "interview mode with meaningless default value"
	CodeDescription[DIP129] = "interview mode with conflicting choice-style edges"
	CodeDescription[DIP130] = "invalid response_format value or on non-agent node"
	CodeDescription[DIP131] = "response_schema and response_format mismatch"
	CodeDescription[DIP132] = "response_schema is not valid JSON"
	CodeDescription[DIP133] = "agent params key shadows a first-class field"
	CodeDescription[DIP134] = "max_retries set in defaults with restart edges but no max_restarts — did you mean max_restarts?"
	CodeDescription[DIP135] = "manager_loop subgraph_ref missing or file does not exist"
	CodeDescription[DIP136] = "manager_loop control field has invalid value (poll_interval or max_cycles)"
	CodeDescription[DIP137] = "unbounded manager_loop: no stop_condition and no max_cycles"
	CodeDescription[DIP138] = "tool node routes on stdout but declares no marker_grep / outputs"
}
