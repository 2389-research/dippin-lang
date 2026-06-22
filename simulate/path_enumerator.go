package simulate

import (
	"fmt"
	"strings"

	"github.com/2389-research/dippin-lang/event"
	"github.com/2389-research/dippin-lang/ir"
)

// All-paths enumeration bounds. De-facto constants: every instance used these
// literals. They mirror the single-path simulator's maxSteps safety valve.
const (
	maxPathDepth   = 200
	maxPathResults = 100
)

// RunAllPaths enumerates all distinct execution paths through the workflow.
// Each Result represents one complete path from start to exit (or a dead end).
// Each path has a unique run ID. Only the scenario context is honored; the
// single-path Options (interactivity, visit caps, branch selection) do not
// apply to exhaustive enumeration.
func RunAllPaths(w *ir.Workflow, scenario map[string]string) ([]*Result, error) {
	if err := validateSimInput(w); err != nil {
		return nil, err
	}

	// Ensure all conditions are parsed into AST form.
	if err := EnsureConditionsParsed(w); err != nil {
		return nil, err
	}

	e := &pathEnumerator{
		workflow: w,
		scenario: scenario,
	}

	return e.enumerate()
}

// pathEnumerator performs depth-first search over all graph edges to enumerate
// every distinct execution path from start to exit (or dead end).
type pathEnumerator struct {
	workflow  *ir.Workflow
	scenario  map[string]string
	results   []*Result
	pathCount int // used to assign unique run IDs
}

// pathState is an immutable snapshot of simulator state passed down the DFS tree.
type pathState struct {
	nodeID  string
	ctx     map[string]string
	visited map[string]int // node → visit count (for loop detection)
	events  []event.Event
	path    []string
	depth   int
}

func (pe *pathEnumerator) enumerate() ([]*Result, error) {
	// Conditions are already parsed by RunAllPaths, the sole caller.
	initial := &pathState{
		nodeID:  pe.workflow.Start,
		ctx:     make(map[string]string),
		visited: make(map[string]int),
		events:  nil,
		path:    nil,
		depth:   0,
	}

	// Seed context with scenario values.
	for k, v := range pe.scenario {
		initial.ctx[k] = v
	}

	pe.explore(initial)

	if len(pe.results) == 0 {
		return nil, fmt.Errorf("no execution paths found")
	}

	return pe.results, nil
}

func (pe *pathEnumerator) explore(state *pathState) {
	if !pe.shouldExplore(state) {
		return
	}

	node := pe.workflow.Node(state.nodeID)
	if node == nil {
		return
	}

	// Clone mutable state for this branch to prevent cross-path contamination.
	ctx, visited, events, path := pe.cloneState(state)

	visited[state.nodeID]++
	path = append(path, state.nodeID)

	// Build and emit the node events.
	events = pe.appendNodeEvents(node, events)

	// Exit node reached — record this complete path.
	if state.nodeID == pe.workflow.Exit {
		pe.recordPath(events, visited, path, "success")
		return
	}

	// Find outgoing edges and explore each branch.
	edges := pe.workflow.EdgesFrom(state.nodeID)
	if len(edges) == 0 {
		pe.recordPath(events, visited, path, "dead_end")
		return
	}

	pe.exploreEdges(edges, events, ctx, visited, path, state.depth)
}

// shouldExplore checks whether the path should continue to be explored.
func (pe *pathEnumerator) shouldExplore(state *pathState) bool {
	if len(pe.results) >= maxPathResults {
		return false
	}
	if state.depth > maxPathDepth {
		return false
	}
	// Loop detection: allow revisiting a node up to 2 times (for retry loops).
	return state.visited[state.nodeID] < 2
}

// cloneState creates deep copies of all mutable state for a branch.
func (pe *pathEnumerator) cloneState(state *pathState) (map[string]string, map[string]int, []event.Event, []string) {
	ctx := cloneMap(state.ctx)
	visited := cloneMapInt(state.visited)
	events := cloneEvents(state.events)
	path := append([]string(nil), state.path...)

	// At the root of the traversal (no events yet), add PipelineStart.
	if len(events) == 0 {
		events = append(events, event.PipelineStart{
			Event:     event.TypePipelineStart,
			RunID:     "sim-path-pending",
			Workflow:  pe.workflow.Name,
			Timestamp: event.Now(),
		})
	}

	return ctx, visited, events, path
}

// appendNodeEvents adds the node_enter (with kind-specific events) and node_exit events.
func (pe *pathEnumerator) appendNodeEvents(node *ir.Node, events []event.Event) []event.Event {
	enterEvt := buildNodeEnterEvent(node, pe.workflow)

	switch cfg := node.Config.(type) {
	case ir.ParallelConfig:
		events = append(events, enterEvt)
		events = append(events, event.ParallelStart{
			Event: event.TypeParallelStart, Node: node.ID,
			Targets: cfg.Targets, Timestamp: event.Now(),
		})
	case ir.FanInConfig:
		events = append(events, enterEvt)
		events = append(events, event.ParallelEnd{
			Event: event.TypeParallelEnd, Node: node.ID,
			Sources: cfg.Sources, Timestamp: event.Now(),
		})
	default:
		events = append(events, enterEvt)
	}

	events = append(events, event.NodeExit{
		Event: event.TypeNodeExit, Node: node.ID,
		Status: "success", DurationMs: 0, Timestamp: event.Now(),
	})
	return events
}

// recordPath records a completed or dead-end path as a result.
func (pe *pathEnumerator) recordPath(events []event.Event, visited map[string]int, path []string, status string) {
	events = append(events, event.PipelineEnd{
		Event:        event.TypePipelineEnd,
		Status:       status,
		NodesVisited: len(visited),
		Timestamp:    event.Now(),
	})
	pe.pathCount++
	events = assignRunID(events, fmt.Sprintf("sim-path-%03d", pe.pathCount))
	pe.results = append(pe.results, &Result{
		Events:       events,
		NodesVisited: len(visited),
		Status:       status,
		Path:         path,
	})
}

// exploreEdges explores each outgoing edge as a separate branch.
func (pe *pathEnumerator) exploreEdges(edges []*ir.Edge, events []event.Event, ctx map[string]string, visited map[string]int, path []string, depth int) {
	for _, e := range edges {
		edgeEvt := buildEdgeTraverseEvent(e)
		branchEvents := append(cloneEvents(events), edgeEvt)

		branchCtx := cloneMap(ctx)
		if e.Condition != nil {
			seedConditionContext(e.Condition.Parsed, branchCtx)
		}

		pe.explore(&pathState{
			nodeID:  e.To,
			ctx:     branchCtx,
			visited: cloneMapInt(visited),
			events:  branchEvents,
			path:    append([]string(nil), path...),
			depth:   depth + 1,
		})
	}
}

// buildEdgeTraverseEvent constructs an EdgeTraverse event from an edge.
func buildEdgeTraverseEvent(e *ir.Edge) event.EdgeTraverse {
	evt := event.EdgeTraverse{
		Event:     event.TypeEdgeTraverse,
		From:      e.From,
		To:        e.To,
		Label:     e.Label,
		Restart:   e.Restart,
		Timestamp: event.Now(),
	}
	if e.Condition != nil {
		evt.Condition = e.Condition.Raw
	}
	return evt
}

// assignRunID replaces the RunID in the first PipelineStart event in the list.
// This is used to assign unique IDs to completed all-paths results.
func assignRunID(events []event.Event, runID string) []event.Event {
	for i, ev := range events {
		if ps, ok := ev.(event.PipelineStart); ok {
			ps.RunID = runID
			events[i] = ps
			return events
		}
	}
	return events
}

// seedConditionContext sets context values that would satisfy the condition.
// Used in all-paths mode to ensure each branch explores its "natural" path.
func seedConditionContext(expr ir.ConditionExpr, ctx map[string]string) {
	switch e := expr.(type) {
	case ir.CondCompare:
		seedCompareContext(e, ctx)
	case ir.CondAnd:
		seedConditionContext(e.Left, ctx)
		seedConditionContext(e.Right, ctx)
	case ir.CondOr:
		// For OR, seed only the left side.
		seedConditionContext(e.Left, ctx)
	case ir.CondNot:
		// Cannot trivially negate; leave context unchanged.
	}
}

// seedCompareContext seeds a single compare condition into the context.
func seedCompareContext(e ir.CondCompare, ctx map[string]string) {
	key := e.Variable
	if strings.HasPrefix(key, "ctx.") {
		key = strings.TrimPrefix(key, "ctx.")
	}
	switch e.Op {
	case "=", "==":
		ctx[key] = e.Value
	case "in":
		parts := strings.Split(e.Value, ",")
		if len(parts) > 0 {
			ctx[key] = strings.TrimSpace(parts[0])
		}
	}
}
