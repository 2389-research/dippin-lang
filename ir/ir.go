// Package ir defines the canonical intermediate representation for Dippin workflows.
//
// The IR is the contract between parsing and execution. It is explicit, normalized,
// and independent of both Dippin syntax and DOT syntax. All downstream consumers
// (engine, validator, formatter, DOT exporter) program against these types.
package ir

import "time"

// Workflow is the top-level IR structure representing a complete pipeline.
type Workflow struct {
	Name     string
	Version  string            // Dippin format version
	Goal     string            // Human-readable objective
	Start    string            // Explicit entry node ID (required)
	Exit     string            // Explicit exit node ID (required)
	Defaults WorkflowDefaults  // Graph-level config
	Vars     map[string]string // User-defined workflow variables
	Requires []string          // Environmental dependencies (e.g. ["git", "docker"]); semantics live in consumers
	// Inputs declares the values a caller must supply — a human at the entry
	// point, or a parent workflow via a subgraph node's params:. This is the
	// callee-side signature. Declaration order is significant: a host renders
	// these as an ordered form, so the formatter must not sort them. Values are
	// untrusted by construction and are read as ${inputs.name}. See issue #190.
	Inputs []*Input
	Nodes  []*Node // Ordered for deterministic processing
	Edges  []*Edge
	// ElseTarget is the section-level `else -> <node>` default of the edges block:
	// the destination for any node whose guard edges all fail to match and which
	// has no explicit unconditional edge of its own. It is the success-side default
	// (it never intercepts a genuine node failure, which routes via Defaults.OnFailure).
	// Empty when no `else` is declared. See docs/proposals/2026-06-16-error-funnel-default.md.
	ElseTarget       string
	ElseTargetSource SourceLocation   // Source location of the `else` entry, for diagnostics
	Stylesheet       []StylesheetRule // Theme/styling rules
	SourceMap        *SourceMap       // File/line mapping for diagnostics
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
	PromptPrefix      string        // Cascade: inline prefix prepended to every agent's prompt (#175)
	PromptSuffix      string        // Cascade: inline suffix appended to every agent's prompt (#175)
	PromptPrefixFile  string        // Cascade: prefix fragment loaded from a file (#175)
	PromptSuffixFile  string        // Cascade: suffix fragment loaded from a file (#175)
	SystemPromptFile  string        // Fallback default: shared system prompt (path) for agents that set none of their own (#72)
}

// Input declares one caller-supplied value bound at run start.
//
// Default, Min and Max are stored as raw source text rather than typed values so
// the formatter can round-trip a file byte-for-byte; the CLI's JSON projection
// coerces them per Type. Unknown Type values are carried verbatim and diagnosed
// by the validator (DIP155), never rejected by the parser — that keeps a .dip
// using a future type parseable, formattable, and packable on an older dippin.
type Input struct {
	Name        string
	Type        string   // v1: text | number | bool | enum | file | secret
	Required    bool     // Host must obtain a value even when Default is set
	Default     string   // Raw source text; a form prefill, not a substitute for Required
	HasDefault  bool     // Distinguishes an absent default from an empty-string default
	Prompt      string   // What a host asks the caller
	Description string   // Help text
	Options     []string // enum choices
	Pattern     string   // text: regex the host enforces
	Min         string   // number: inclusive lower bound, raw text
	Max         string   // number: inclusive upper bound, raw text
	MaxLength   int      // text: character cap
	Multiline   bool     // text: host renders a textarea
	Source      SourceLocation
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
	PromptInclude       string // Fragment file appended after the body, before the cascade suffix (#175). Composed into Prompt by ResolveFileDirectives.
	PromptPrefix        string // Node-level: "none" opts out of the defaults prompt_prefix cascade; "" = inherit (#175).
	PromptSuffix        string // Node-level: "none" opts out of the defaults prompt_suffix cascade; "" = inherit (#175).
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
	// LastResponseTruncate caps, at the runtime, the number of Unicode
	// characters of the auto-injected previous response ("last response") that
	// this agent receives in its prompt. 0 / unset = no truncation (full
	// response injected). A chain-attack mitigation (issue #56): it bounds how
	// much potentially-tainted upstream output reaches a privileged prompt.
	// dippin carries + lints (DIP148 flags a negative value); the runtime
	// enforces the truncation. Inert until a runtime reads it.
	LastResponseTruncate int
	Params               map[string]string // Generic key-value pairs passed through to runtime
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
	// LastResponseTruncate is a per-branch override of the target agent's
	// last_response_truncate. 0 INHERITS the target agent's value (never resets
	// to "no truncation") — the runtime resolves effective = branch if > 0 else
	// agent, mirroring ToolAccess / WritablePaths inheritance. See issue #56.
	LastResponseTruncate int
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
	Policy      string        // Named policy: "standard", "aggressive", "patient", "linear", "none"
	MaxRetries  int           // Override default
	BaseDelay   time.Duration // Override policy's default base delay (optional)
	RetryTarget string        // Node to jump to on retry (dip 1 + dip 2 spelling: retry_target)
	// FallbackTarget is the retry-EXHAUSTION route — where control goes when the
	// retry budget is spent. It is a distinct runtime channel from the edges block
	// (which routes on genuine node failure), which the engine reads from a node
	// attribute. Spelled fallback_target in dip 1 and fallback_retry_target in
	// dip 2 (the dip-2 name disambiguates it from `on fail` edges). See #186/#204.
	FallbackTarget string
}

// NodeIO declares what context keys a node reads and writes.
// Both use bare logical names (e.g., "human_response", not "ctx.human_response").
// Advisory in v1 — validated as warnings, not errors.
type NodeIO struct {
	Reads  []string // Context keys this node expects
	Writes []string // Context keys this node produces
}
