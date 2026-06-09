// Package ir defines the canonical intermediate representation for Dippin workflows.
//
// The IR is the contract between parsing and execution. It is explicit, normalized,
// and independent of both Dippin syntax and DOT syntax. All downstream consumers
// (engine, validator, formatter, DOT exporter) program against these types.
package ir

import "time"

// Workflow is the top-level IR structure representing a complete pipeline.
type Workflow struct {
	Name       string
	Version    string            // Dippin format version
	Goal       string            // Human-readable objective
	Start      string            // Explicit entry node ID (required)
	Exit       string            // Explicit exit node ID (required)
	Defaults   WorkflowDefaults  // Graph-level config
	Vars       map[string]string // User-defined workflow variables
	Requires   []string          // Environmental dependencies (e.g. ["git", "docker"]); semantics live in consumers
	Nodes      []*Node           // Ordered for deterministic processing
	Edges      []*Edge
	Stylesheet []StylesheetRule // Theme/styling rules
	SourceMap  *SourceMap       // File/line mapping for diagnostics
}

// StylesheetRule pairs a selector with a set of properties.
type StylesheetRule struct {
	Selector   StyleSelector
	Properties map[string]string
}

// StyleSelector identifies what a stylesheet rule targets.
type StyleSelector struct {
	Kind  string // "universal", "kind", "class", "id"
	Value string // "*", "agent", "coder", "critical_gate"
}

// WorkflowDefaults holds graph-level configuration that applies to all nodes
// unless overridden at the node level.
type WorkflowDefaults struct {
	Model             string        // Default LLM model
	Provider          string        // Default LLM provider
	RetryPolicy       string        // Default retry policy name
	MaxRetries        int           // Default max retries
	Fidelity          string        // Default fidelity level
	MaxRestarts       int           // Max loop restarts (default 5)
	RestartTarget     string        // Where to restart on loop
	OnFailure         string        // Default failure route (node to jump to on failure)
	CacheTools        bool          // Cache tool results
	Compaction        string        // Context compaction mode
	OnResume          string        // Fidelity behavior on resume: "preserve" or "degrade"
	MaxTotalTokens    int           // Hard ceiling on total tokens
	MaxCostCents      int           // Hard ceiling on cost in cents (USD)
	MaxWallTime       time.Duration // Hard ceiling on wall time
	StallTimeout      time.Duration // Abort/route when no progress for this wall-clock span (0 = disabled)
	ToolCommandsAllow string        // Comma-separated glob allowlist for tool shell commands
	ToolDenylistAdd   string        // Comma-separated globs appended to the runtime's default denylist
}

// Node represents a single step in the workflow.
type Node struct {
	ID      string
	Kind    NodeKind
	Label   string     // Human-readable display name
	Classes []string   // For stylesheet matching (post-v1)
	Config  NodeConfig // Kind-specific configuration
	Retry   RetryConfig
	IO      NodeIO // Declared inputs/outputs (advisory in v1)
	Source  SourceLocation
}

// NodeKind enumerates node types explicitly.
type NodeKind string

const (
	NodeAgent       NodeKind = "agent"
	NodeHuman       NodeKind = "human"
	NodeTool        NodeKind = "tool"
	NodeParallel    NodeKind = "parallel"
	NodeFanIn       NodeKind = "fan_in"
	NodeSubgraph    NodeKind = "subgraph"
	NodeConditional NodeKind = "conditional"
	NodeManagerLoop NodeKind = "manager_loop"
)

// NodeConfig is implemented by kind-specific configuration types.
// The sealed interface prevents invalid combinations structurally.
type NodeConfig interface {
	nodeConfig()
}

// AgentConfig holds configuration for LLM agent nodes.
type AgentConfig struct {
	Prompt              string
	PromptFile          string // Source path; coexists with Prompt after resolve (formatter prefers the directive form). Populated by parser.ResolveFileDirectives.
	SystemPrompt        string
	SystemPromptFile    string // Source path; coexists with SystemPrompt after resolve (formatter prefers the directive form). Populated by parser.ResolveFileDirectives.
	Model               string // Per-node override
	Provider            string
	MaxTurns            int
	CmdTimeout          time.Duration
	CacheTools          bool
	Compaction          string
	CompactionThreshold float64
	ReasoningEffort     string
	Fidelity            string
	AutoStatus          bool   // Parse STATUS: from response
	GoalGate            bool   // Pipeline fails if this node fails
	ResponseFormat      string // "json_object" or "json_schema"
	ResponseSchema      string // JSON schema (when ResponseFormat is "json_schema")
	Backend             string // Per-node backend override: "native", "claude-code", "acp"
	WorkingDir          string // Per-node working directory override
	ToolAccess          string // Raw value: "" or "none" recognized; other values lint as DIP139 and fail-closed to no-tools at runtime
	// WritablePaths bounds the file paths this agent's tools may write, as
	// author-chosen globs (e.g. "workspace/**", ".ai/sprints/**") resolved against
	// the session root. Empty/absent = unbounded. A present-but-empty or malformed
	// value fails CLOSED at the runtime (deny-all / refuse-to-start), never
	// unbounded. dippin carries + lints; the runtime enforces an fs-level write jail on
	// the native backend (Bash + its children included); claude-code/acp refuse to
	// start. See issue #75.
	WritablePaths []string
	Params        map[string]string // Generic key-value pairs passed through to runtime
}

func (AgentConfig) nodeConfig() {}

// HumanConfig holds configuration for human gate nodes.
type HumanConfig struct {
	Mode          string        // "choice" | "freeform" | "interview" | "yes_no"
	Default       string        // Default choice
	Prompt        string        // Instructions shown to the human
	QuestionsKey  string        // Context key to read questions from (interview mode)
	AnswersKey    string        // Context key to write answers to (interview mode)
	Timeout       time.Duration // Per-gate timeout; 0 = no timeout
	TimeoutAction string        // "fail" | "default" | "" (pick default-if-set else fail)
}

func (HumanConfig) nodeConfig() {}

// ToolConfig holds configuration for shell command nodes.
type ToolConfig struct {
	Command       string // Shell command (multiline OK)
	CommandFile   string // Source path; coexists with Command after resolve (formatter prefers the directive form). Populated by parser.ResolveFileDirectives.
	Timeout       time.Duration
	Outputs       []string // Declared possible stdout values for coverage analysis
	MarkerGrep    string   // Regex matched line-by-line against stdout; populates ctx.tool_marker
	RouteRequired bool     // True → node fails if the command emits no routing signal recognized by the runtime
	OutputLimit   int      // Bytes; > 0 = override engine default
}

func (ToolConfig) nodeConfig() {}

// ParallelConfig holds configuration for fan-out nodes.
type ParallelConfig struct {
	Targets  []string          // Fan-out target node IDs (inline form)
	Branches []BranchConfig    // Per-branch config (block form)
	Params   map[string]string // Generic key-value pairs passed through to runtime (e.g. fan-in aggregation policy); carried, not interpreted
}

func (ParallelConfig) nodeConfig() {}

// BranchConfig holds per-branch configuration for block-form parallel nodes.
type BranchConfig struct {
	Target   string
	Model    string
	Provider string
	Fidelity string
	// ToolAccess is a per-branch override of the target agent's tool_access.
	// Recognized values mirror AgentConfig.ToolAccess: "" (inherit) and "none"
	// (strip tools); other values lint as DIP139 and fail closed at runtime.
	// Empty INHERITS the target agent's tool_access (never resets to the full
	// catalog) — the runtime resolves effective = branch if non-empty else agent.
	// dippin carries + lints this field; the runtime enforces the override, exactly
	// as it does for Model/Provider/Fidelity.
	ToolAccess string
	// WritablePaths is a per-branch override of the target agent's writable_paths.
	// Empty INHERITS the target agent's writable_paths (never resets to unbounded) —
	// the runtime resolves effective = branch if non-empty else agent. dippin carries +
	// lints; the runtime enforces. See issue #75.
	WritablePaths []string
}

// FanInConfig holds configuration for join nodes.
type FanInConfig struct {
	Sources []string          // Source node IDs to join
	Params  map[string]string // Generic key-value pairs passed through to runtime (e.g. fan-in aggregation policy); carried, not interpreted
}

func (FanInConfig) nodeConfig() {}

// SubgraphConfig holds configuration for embedded sub-pipeline nodes.
type SubgraphConfig struct {
	Ref    string            // Workflow name or path
	Params map[string]string // Parameter overrides
}

func (SubgraphConfig) nodeConfig() {}

// ConditionalConfig holds configuration for pure conditional branching nodes.
// Conditional nodes evaluate outgoing edge conditions without making an LLM call.
type ConditionalConfig struct{}

func (ConditionalConfig) nodeConfig() {}

// ManagerLoopConfig holds configuration for manager_loop supervisor nodes.
// A manager_loop runs a child subgraph, polls at PollInterval, and may
// steer the child by injecting SteerContext when SteerCondition evaluates
// true against stack.child.* variables exposed by the runtime.
type ManagerLoopConfig struct {
	SubgraphRef    string            // Child subgraph to supervise (required)
	PollInterval   time.Duration     // Polling cadence; 0 = event-driven
	MaxCycles      int               // Hard cap on child cycles; 0 = unbounded
	StopCondition  *Condition        // Terminate supervision when true
	SteerCondition *Condition        // Inject SteerContext when true
	SteerContext   map[string]string // Key-value hints injected into child
}

func (ManagerLoopConfig) nodeConfig() {}

// RetryConfig specifies retry behavior for a node.
type RetryConfig struct {
	Policy         string        // Named policy: "standard", "aggressive", "patient", "linear", "none"
	MaxRetries     int           // Override default
	BaseDelay      time.Duration // Override policy's default base delay (optional)
	RetryTarget    string        // Node to jump to on retry
	FallbackTarget string        // Fallback if retries exhausted
}

// NodeIO declares what context keys a node reads and writes.
// Both use bare logical names (e.g., "human_response", not "ctx.human_response").
// Advisory in v1 — validated as warnings, not errors.
type NodeIO struct {
	Reads  []string // Context keys this node expects
	Writes []string // Context keys this node produces
}
