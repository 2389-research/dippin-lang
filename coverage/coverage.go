// Package coverage analyzes workflow edge coverage and reachability.
//
// It examines tool nodes to determine whether their possible outputs are
// covered by outgoing edge conditions, checks that all nodes are reachable
// from start, and verifies that all reachable nodes have a path to exit.
package coverage

import (
	"sort"
	"strings"

	"github.com/2389-research/dippin-lang/ir"
	"github.com/2389-research/dippin-lang/simulate"
	"mvdan.cc/sh/v3/syntax"
)

// Report is the top-level result of a coverage analysis.
type Report struct {
	Nodes        map[string]NodeCoverage `json:"nodes"`
	Reachability ReachabilityReport      `json:"reachability"`
	Termination  TerminationReport       `json:"termination"`
}

// NodeCoverage describes edge coverage for a single node.
type NodeCoverage struct {
	NodeID           string   `json:"node_id"`
	Status           string   `json:"status"` // "covered", "partial", "unknown", "no_conditions"
	DeclaredOutputs  []string `json:"declared_outputs,omitempty"`
	ExtractedOutputs []string `json:"extracted_outputs,omitempty"`
	EdgeConditions   []string `json:"edge_conditions"`
	MissingEdges     []string `json:"missing_edges,omitempty"`
	HasFallback      bool     `json:"has_fallback"`
}

// ReachabilityReport summarizes forward reachability from the start node.
type ReachabilityReport struct {
	TotalNodes       int      `json:"total_nodes"`
	ReachableNodes   int      `json:"reachable_nodes"`
	UnreachableNodes []string `json:"unreachable_nodes,omitempty"`
}

// TerminationReport summarizes whether all reachable nodes can reach exit.
type TerminationReport struct {
	AllPathsTerminate bool     `json:"all_paths_terminate"`
	SinkNodes         []string `json:"sink_nodes,omitempty"`
}

// Analyze runs coverage analysis on a workflow and returns a report.
func Analyze(w *ir.Workflow) *Report {
	// Parse edge conditions into AST form. Errors are non-fatal —
	// unparseable conditions stay with Parsed=nil and are skipped
	// by appendConditionValues (the default case returns out unchanged).
	_ = simulate.EnsureConditionsParsed(w)

	r := &Report{
		Nodes: make(map[string]NodeCoverage),
	}
	analyzeNodes(w, r)
	r.Reachability = analyzeReachability(w)
	r.Termination = analyzeTermination(w)
	return r
}

// analyzeNodes evaluates edge coverage for each tool node.
func analyzeNodes(w *ir.Workflow, r *Report) {
	for _, n := range w.Nodes {
		if _, ok := n.Config.(ir.ToolConfig); !ok {
			continue
		}
		r.Nodes[n.ID] = analyzeToolNode(w, n)
	}
}

// analyzeToolNode computes coverage for a single tool node.
func analyzeToolNode(w *ir.Workflow, n *ir.Node) NodeCoverage {
	cfg, ok := n.Config.(ir.ToolConfig)
	if !ok {
		return NodeCoverage{NodeID: n.ID}
	}
	edges := w.EdgesFrom(n.ID)
	conditions, hasFallback := collectEdgeConditions(edges)

	cov := NodeCoverage{
		NodeID:          n.ID,
		DeclaredOutputs: cfg.Outputs,
		EdgeConditions:  conditions,
		HasFallback:     hasFallback,
	}

	// Prefer declared outputs over regex extraction.
	outputs := cfg.Outputs
	if len(outputs) == 0 {
		outputs = extractToolOutputs(cfg.Command)
		cov.ExtractedOutputs = outputs
	}

	cov.MissingEdges = findMissingEdges(outputs, conditions)
	cov.Status = determineStatus(cov)
	return cov
}

// determineStatus assigns a coverage status based on the node's edges and outputs.
func determineStatus(cov NodeCoverage) string {
	if len(cov.EdgeConditions) == 0 {
		return "no_conditions"
	}
	if len(cov.ExtractedOutputs) == 0 {
		return "unknown"
	}
	if len(cov.MissingEdges) == 0 || cov.HasFallback {
		return "covered"
	}
	return "partial"
}

// findMissingEdges returns extracted outputs that have no matching edge condition.
func findMissingEdges(outputs, conditions []string) []string {
	condSet := toSet(conditions)
	var missing []string
	for _, o := range outputs {
		if !condSet[o] {
			missing = append(missing, o)
		}
	}
	return missing
}

// toSet converts a string slice to a map for O(1) lookups.
func toSet(items []string) map[string]bool {
	s := make(map[string]bool, len(items))
	for _, item := range items {
		s[item] = true
	}
	return s
}

// echoFlags are echo invocations that should be skipped when extracting
// literal output values from an echo call's arguments.
var echoFlags = map[string]bool{
	"-n": true, "-e": true, "-E": true, "--": true,
}

// extractToolOutputs finds printf/echo string literals in a shell command
// whose stdout reaches the engine. Statements that redirect to files or
// feed into pipes are skipped, as are echo/printf calls nested inside
// command substitution. Returns nil on shell parse errors — DIP123 catches
// syntax problems upstream.
func extractToolOutputs(command string) []string {
	parser := syntax.NewParser(syntax.KeepComments(false))
	prog, err := parser.Parse(strings.NewReader(command), "")
	if err != nil {
		return nil
	}
	var outputs []string
	seen := make(map[string]bool)
	syntax.Walk(prog, walkForOutputs(&outputs, seen))
	return outputs
}

// walkForOutputs returns a syntax.Walk callback that collects literal
// echo/printf args from statements whose stdout reaches the engine.
func walkForOutputs(outputs *[]string, seen map[string]bool) func(syntax.Node) bool {
	return func(node syntax.Node) bool {
		switch n := node.(type) {
		case *syntax.Stmt:
			return !stmtIsRedirected(n)
		case *syntax.CmdSubst:
			return false
		case *syntax.CallExpr:
			collectFromCall(n, outputs, seen)
		}
		return true
	}
}

// stmtIsRedirected returns true if a statement's stdout does not reach
// the engine — either via a stdout-affecting redirect or a pipe. Stdin
// redirects (`<`) and stderr-only redirects (`2>err.log`, `2>&1`) leave
// stdout intact and do NOT cause the statement to be skipped.
func stmtIsRedirected(s *syntax.Stmt) bool {
	for _, r := range s.Redirs {
		if redirAffectsStdout(r) {
			return true
		}
	}
	if bin, ok := s.Cmd.(*syntax.BinaryCmd); ok {
		return bin.Op == syntax.Pipe || bin.Op == syntax.PipeAll
	}
	return false
}

// redirAffectsStdout reports whether a single redirect diverts fd 1 away
// from the engine's stdout. The `&>` / `&>>` operators always do; the
// per-fd operators (`>`, `>>`, `>&`, `>|`, `<>`) only do when the source
// fd is unset (defaults to 1) or explicitly "1".
func redirAffectsStdout(r *syntax.Redirect) bool {
	switch r.Op {
	case syntax.RdrAll, syntax.AppAll:
		return true
	case syntax.RdrOut, syntax.AppOut, syntax.DplOut, syntax.RdrClob, syntax.RdrInOut:
		return redirTargetsStdout(r.N)
	}
	return false
}

// redirTargetsStdout returns true if a redirect's fd number is unset
// (defaults to stdout) or explicitly the literal "1".
func redirTargetsStdout(n *syntax.Lit) bool {
	return n == nil || n.Value == "1"
}

// collectFromCall extracts literal output args from an echo or printf call.
func collectFromCall(call *syntax.CallExpr, outputs *[]string, seen map[string]bool) {
	if len(call.Args) == 0 {
		return
	}
	name := extractWordLiteralLocal(call.Args[0])
	switch name {
	case "echo":
		collectEchoArgs(call.Args[1:], outputs, seen)
	case "printf":
		collectPrintfArgs(call.Args[1:], outputs, seen)
	}
}

// collectEchoArgs extracts literal args from `echo`, skipping flag args.
func collectEchoArgs(args []*syntax.Word, outputs *[]string, seen map[string]bool) {
	for _, w := range args {
		lit := extractWordLiteralLocal(w)
		if lit == "" || echoFlags[lit] {
			continue
		}
		addOutput(lit, outputs, seen)
	}
}

// collectPrintfArgs extracts literal args from `printf`. The two-arg
// `printf '%s' 'value'` and `printf '%s\n' 'value'` patterns extract the
// value; otherwise the format-string itself is the output literal.
func collectPrintfArgs(args []*syntax.Word, outputs *[]string, seen map[string]bool) {
	if len(args) == 0 {
		return
	}
	first := extractWordLiteralLocal(args[0])
	if first == "" {
		return
	}
	if isFormatTwoArg(first, args) {
		if val := extractWordLiteralLocal(args[1]); val != "" {
			addOutput(val, outputs, seen)
		}
		return
	}
	addOutput(first, outputs, seen)
}

// isFormatTwoArg returns true if printf was invoked with a bare `%s` or
// `%s\n` format string and exactly one value arg.
func isFormatTwoArg(first string, args []*syntax.Word) bool {
	return (first == "%s" || first == "%s\\n") && len(args) == 2
}

// addOutput appends a literal output value if it hasn't been seen and
// isn't itself a format specifier.
func addOutput(val string, outputs *[]string, seen map[string]bool) {
	if seen[val] || isFormatSpecifier(val) {
		return
	}
	seen[val] = true
	*outputs = append(*outputs, val)
}

// extractWordLiteralLocal returns the literal string content of a simple
// Word arg, handling bare, single-quoted, and double-quoted forms.
// Returns "" if the word contains expansions or substitutions.
func extractWordLiteralLocal(w *syntax.Word) string {
	if w == nil || len(w.Parts) != 1 {
		return ""
	}
	return extractPartLiteral(w.Parts[0])
}

// extractPartLiteral returns the literal content of a single WordPart,
// or "" if the part is not a simple literal.
func extractPartLiteral(part syntax.WordPart) string {
	switch p := part.(type) {
	case *syntax.Lit:
		return p.Value
	case *syntax.SglQuoted:
		return p.Value
	case *syntax.DblQuoted:
		return extractDblQuotedLit(p)
	}
	return ""
}

// extractDblQuotedLit returns the literal content of a double-quoted
// word part, or "" if it contains expansions.
func extractDblQuotedLit(d *syntax.DblQuoted) string {
	if len(d.Parts) != 1 {
		return ""
	}
	lit, ok := d.Parts[0].(*syntax.Lit)
	if !ok {
		return ""
	}
	return lit.Value
}

// isFormatSpecifier returns true if the string looks like a printf format string.
func isFormatSpecifier(s string) bool {
	return strings.HasPrefix(s, "%") && len(s) <= 3
}

// collectEdgeConditions extracts condition values from edges and detects fallbacks.
func collectEdgeConditions(edges []*ir.Edge) (conditions []string, hasFallback bool) {
	for _, e := range edges {
		if e.Condition == nil {
			hasFallback = true
			continue
		}
		conditions = appendConditionValues(conditions, e.Condition.Parsed)
	}
	return conditions, hasFallback
}

// appendConditionValues extracts comparison values from a condition expression tree.
func appendConditionValues(out []string, expr ir.ConditionExpr) []string {
	switch e := expr.(type) {
	case ir.CondCompare:
		return append(out, e.Value)
	case ir.CondAnd:
		out = appendConditionValues(out, e.Left)
		return appendConditionValues(out, e.Right)
	case ir.CondOr:
		out = appendConditionValues(out, e.Left)
		return appendConditionValues(out, e.Right)
	case ir.CondNot:
		return appendConditionValues(out, e.Inner)
	default:
		return out
	}
}

// analyzeReachability performs BFS from start to find unreachable nodes.
func analyzeReachability(w *ir.Workflow) ReachabilityReport {
	adj := buildForwardAdjacency(w)
	reachable := bfsFrom(w.Start, adj)

	report := ReachabilityReport{
		TotalNodes:     len(w.Nodes),
		ReachableNodes: len(reachable),
	}
	report.UnreachableNodes = findUnreachableNodes(w, reachable)
	return report
}

// findUnreachableNodes returns sorted IDs of nodes not in the reachable set.
func findUnreachableNodes(w *ir.Workflow, reachable map[string]bool) []string {
	var unreachable []string
	for _, n := range w.Nodes {
		if !reachable[n.ID] {
			unreachable = append(unreachable, n.ID)
		}
	}
	sort.Strings(unreachable)
	return unreachable
}

// analyzeTermination checks that all reachable nodes can reach exit.
func analyzeTermination(w *ir.Workflow) TerminationReport {
	fwd := buildForwardAdjacency(w)
	reachable := bfsFrom(w.Start, fwd)

	rev := buildReverseAdjacency(w)
	canReachExit := bfsFrom(w.Exit, rev)

	sinks := findSinkNodes(reachable, canReachExit)
	return TerminationReport{
		AllPathsTerminate: len(sinks) == 0,
		SinkNodes:         sinks,
	}
}

// findSinkNodes returns sorted IDs of nodes reachable from start but not from exit (reverse).
func findSinkNodes(reachable, canReachExit map[string]bool) []string {
	var sinks []string
	for id := range reachable {
		if !canReachExit[id] {
			sinks = append(sinks, id)
		}
	}
	sort.Strings(sinks)
	return sinks
}

// buildForwardAdjacency creates a forward adjacency map from all edges.
func buildForwardAdjacency(w *ir.Workflow) map[string][]string {
	adj := make(map[string][]string)
	for _, e := range w.Edges {
		adj[e.From] = append(adj[e.From], e.To)
	}
	addImplicitEdges(adj, w)
	return adj
}

// buildReverseAdjacency creates a reverse adjacency map from all edges.
func buildReverseAdjacency(w *ir.Workflow) map[string][]string {
	adj := make(map[string][]string)
	for _, e := range w.Edges {
		adj[e.To] = append(adj[e.To], e.From)
	}
	addReverseImplicitEdges(adj, w)
	return adj
}

// addImplicitEdges adds forward edges from parallel and fan_in nodes.
func addImplicitEdges(adj map[string][]string, w *ir.Workflow) {
	for _, n := range w.Nodes {
		switch cfg := n.Config.(type) {
		case ir.ParallelConfig:
			adj[n.ID] = append(adj[n.ID], cfg.Targets...)
		case ir.FanInConfig:
			for _, src := range cfg.Sources {
				adj[src] = append(adj[src], n.ID)
			}
		}
	}
}

// addReverseImplicitEdges adds reverse edges from parallel and fan_in nodes.
func addReverseImplicitEdges(adj map[string][]string, w *ir.Workflow) {
	for _, n := range w.Nodes {
		addReverseNodeEdges(adj, n)
	}
}

// addReverseNodeEdges adds reverse edges for a single parallel or fan_in node.
func addReverseNodeEdges(adj map[string][]string, n *ir.Node) {
	switch cfg := n.Config.(type) {
	case ir.ParallelConfig:
		for _, t := range cfg.Targets {
			adj[t] = append(adj[t], n.ID)
		}
	case ir.FanInConfig:
		adj[n.ID] = append(adj[n.ID], cfg.Sources...)
	}
}

// bfsFrom returns the set of nodes reachable from start via BFS.
func bfsFrom(start string, adj map[string][]string) map[string]bool {
	visited := make(map[string]bool)
	queue := []string{start}
	visited[start] = true

	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]
		for _, next := range adj[curr] {
			if !visited[next] {
				visited[next] = true
				queue = append(queue, next)
			}
		}
	}
	return visited
}
