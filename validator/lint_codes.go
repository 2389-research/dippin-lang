package validator

// Diagnostic codes for semantic quality warnings (DIP101–DIP151).
const (
	DIP101 = "DIP101" // unreachable nodes after conditional branches
	DIP102 = "DIP102" // routing node without default/unconditional edge
	DIP103 = "DIP103" // duplicate condition on edges from the same node
	DIP104 = "DIP104" // unbounded retry loop
	DIP105 = "DIP105" // no success path to exit
	DIP106 = "DIP106" // unrecognized variable reference in prompt
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
	DIP130 = "DIP130" // invalid response_format value
	DIP131 = "DIP131" // response_schema/response_format mismatch
	DIP132 = "DIP132" // response_schema is not valid JSON
	DIP133 = "DIP133" // agent params key shadows a first-class field
	DIP134 = "DIP134" // max_retries set with restart edges but no max_restarts
	DIP135 = "DIP135" // manager_loop subgraph_ref missing or file does not exist
	DIP136 = "DIP136" // manager_loop control field has invalid value (poll_interval or max_cycles)
	DIP137 = "DIP137" // unbounded manager_loop: no stop_condition and no max_cycles
	DIP138 = "DIP138" // tool node routes on stdout but declares no marker_grep / outputs (reserved)
	DIP139 = "DIP139" // invalid tool_access value on agent node or parallel branch
	DIP140 = "DIP140" // params re-enables tools that tool_access strips
	DIP141 = "DIP141" // writable_paths nullified by tool_access: none (dead config)
	DIP142 = "DIP142" // unsafe writable_paths entry (absolute / ~ / parent-escape / brace)
	DIP143 = "DIP143" // referenced subgraph does not inherit this workflow's tool_access restrictions
	DIP144 = "DIP144" // agent node has no failure route
	DIP145 = "DIP145" // graph budget default is negative
	// DIP146 is emitted by the CLI's native cross-file pass (cmd/dippin), NOT by
	// validator.Lint() — the validator cannot read the child .dip. Registered here
	// only so it appears in the catalog / `dippin explain` / docs.
	DIP146 = "DIP146" // child subgraph re-grants tools the parent restricted (cross-file)
	DIP147 = "DIP147" // restricted (tool_access:none) agent output flows into a tool-bearing agent (chain-attack / info-flow)
	DIP148 = "DIP148" // last_response_truncate is negative
	DIP149 = "DIP149" // ambiguous routing: 2+ unconditional outgoing edges resolved only by the lexical tiebreak
	DIP150 = "DIP150" // human gate routes by label: which doubles as the routing key — use choice: to mark it explicitly
	DIP151 = "DIP151" // edge sets weight: which routing does not use (soft-deprecated, slated for removal in dip 2)
)

func init() {
	// Extend CodeDescription with linter codes.
	for code, desc := range linterCodeDescriptions {
		CodeDescription[code] = desc
	}
}

var linterCodeDescriptions = map[string]string{
	DIP101: "unreachable node after conditional branches",
	DIP102: "routing node has no default/unconditional edge",
	DIP103: "duplicate condition on edges from the same node",
	DIP104: "unbounded retry loop (no max_retries or fallback)",
	DIP105: "no success path from start to exit",
	DIP106: "unrecognized variable reference in prompt",
	DIP107: "unused context key (written but never read)",
	DIP108: "unknown model/provider combination",
	DIP109: "namespace collision in imports",
	DIP110: "empty prompt on agent node",
	DIP111: "tool command has no timeout",
	DIP112: "reads key not produced by any upstream writes",
	DIP113: "invalid retry policy name",
	DIP114: "invalid fidelity level",
	DIP115: "goal_gate node without retry or fallback target",
	DIP116: "invalid compaction threshold or on_resume value",
	DIP117: "stylesheet class references undefined class",
	DIP118: "stylesheet ID references unknown node",
	DIP119: "invalid reasoning_effort value",
	DIP120: "condition variable missing namespace prefix",
	DIP121: "condition references variable not in source node writes",
	DIP122: "condition tests value not in source tool outputs",
	DIP123: "tool command has shell syntax errors",
	DIP124: "tool command references runtime-only ${ctx.*} variable",
	DIP125: "tool command binary not found on PATH",
	DIP126: "subgraph ref file does not exist",
	DIP127: "invalid human node mode",
	DIP128: "interview mode with meaningless default value",
	DIP129: "interview mode with conflicting choice-style edges",
	DIP130: "invalid response_format value",
	DIP131: "response_schema and response_format mismatch",
	DIP132: "response_schema is not valid JSON",
	DIP133: "agent params key shadows a first-class field",
	DIP134: "max_retries set in defaults with restart edges but no max_restarts — did you mean max_restarts?",
	DIP135: "manager_loop subgraph_ref missing or file does not exist",
	DIP136: "manager_loop control field has invalid value (poll_interval or max_cycles)",
	DIP137: "unbounded manager_loop: no stop_condition and no max_cycles",
	DIP138: "tool node routes on stdout but declares no marker_grep / outputs",
	DIP139: "invalid tool_access value on agent node or parallel branch",
	DIP140: "params re-enables tools that tool_access strips",
	DIP141: "writable_paths nullified by tool_access: none (dead config)",
	DIP142: "unsafe writable_paths entry (absolute / ~ / parent-escape / brace)",
	DIP143: "referenced subgraph does not inherit this workflow's tool_access restrictions",
	DIP144: "agent node has no failure route",
	DIP145: "graph budget default is negative",
	DIP146: "child subgraph re-grants tools the parent restricted (cross-file)",
	DIP147: "restricted agent output flows into a tool-bearing agent (chain-attack)",
	DIP148: "last_response_truncate is negative",
	DIP149: "ambiguous routing: multiple unconditional outgoing edges",
	DIP150: "human gate routes by label; use choice: to mark the routing key",
	DIP151: "edge weight: is unused by routing (soft-deprecated)",
}
