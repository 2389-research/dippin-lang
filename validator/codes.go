package validator

// Diagnostic codes for graph structure validation (DIP001–DIP010).
const (
	DIP001 = "DIP001" // start node missing
	DIP002 = "DIP002" // exit node missing
	DIP003 = "DIP003" // unknown node reference (edge endpoint or on_failure target)
	DIP004 = "DIP004" // unreachable node(s) from start
	DIP005 = "DIP005" // unconditional cycle detected
	DIP006 = "DIP006" // exit node has outgoing edges
	DIP007 = "DIP007" // parallel/fan_in mismatch
	DIP008 = "DIP008" // duplicate node ID
	DIP009 = "DIP009" // duplicate edge
	DIP010 = "DIP010" // unparseable edge condition
)

// CodeDescription maps each code to a short human-readable description.
var CodeDescription = map[string]string{
	DIP001: "start node does not exist",
	DIP002: "exit node does not exist",
	DIP003: "unknown node reference in edge or on_failure",
	DIP004: "node unreachable from start",
	DIP005: "unconditional cycle detected",
	DIP006: "exit node has outgoing edges",
	DIP007: "parallel fan-out/fan-in mismatch",
	DIP008: "duplicate node ID",
	DIP009: "duplicate edge",
	DIP010: "edge condition cannot be parsed",
}
