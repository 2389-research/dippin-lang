package simulate

import (
	"fmt"

	"github.com/2389-research/dippin-lang/event"
	"github.com/2389-research/dippin-lang/ir"
)

// buildNodeEnterEvent constructs a NodeEnter event with kind-specific fields.
func buildNodeEnterEvent(node *ir.Node, w *ir.Workflow) event.NodeEnter {
	enterEvt := event.NodeEnter{
		Event:     event.TypeNodeEnter,
		Node:      node.ID,
		Kind:      string(node.Kind),
		Label:     node.Label,
		Timestamp: event.Now(),
	}
	populateEnterFields(&enterEvt, node, w)
	return enterEvt
}

// populateEnterFields fills kind-specific fields on a NodeEnter event.
func populateEnterFields(evt *event.NodeEnter, node *ir.Node, w *ir.Workflow) {
	switch cfg := node.Config.(type) {
	case ir.AgentConfig:
		evt.Model = resolveModel(cfg.Model, w.Defaults.Model)
		evt.Provider = resolveProvider(cfg.Provider, w.Defaults.Provider)
		evt.Prompt = cfg.Prompt
		evt.Fidelity = resolveFidelity(cfg.Fidelity, w.Defaults.Fidelity)
	case ir.ToolConfig:
		evt.Command = cfg.Command
	case ir.HumanConfig:
		evt.Mode = cfg.Mode
		evt.Prompt = cfg.Prompt
	case ir.SubgraphConfig:
		evt.Label = fmt.Sprintf("subgraph:%s", cfg.Ref)
	}
}

// buildPipelineStartEvent constructs a PipelineStart event.
func buildPipelineStartEvent(runID, workflowName string) event.PipelineStart {
	return event.PipelineStart{
		Event:     event.TypePipelineStart,
		RunID:     runID,
		Workflow:  workflowName,
		Timestamp: event.Now(),
	}
}

// buildPipelineEndEvent constructs a PipelineEnd event.
func buildPipelineEndEvent(status string, nodesVisited int) event.PipelineEnd {
	return event.PipelineEnd{
		Event:        event.TypePipelineEnd,
		Status:       status,
		NodesVisited: nodesVisited,
		Timestamp:    event.Now(),
	}
}

// newNodeExitSuccess constructs a successful NodeExit event for a node.
func newNodeExitSuccess(nodeID string) event.NodeExit {
	return event.NodeExit{
		Event:      event.TypeNodeExit,
		Node:       nodeID,
		Status:     "success",
		DurationMs: 0,
		Timestamp:  event.Now(),
	}
}

// buildNodeEventSequence returns the ordered events for visiting a node:
// the node_enter event, any kind-specific parallel/fan-in event, and the
// trailing successful node_exit. Used by both the single-path simulator and
// the all-paths enumerator so they emit identical sequences.
func buildNodeEventSequence(node *ir.Node, enterEvt event.NodeEnter) []event.Event {
	events := []event.Event{enterEvt}
	switch cfg := node.Config.(type) {
	case ir.ParallelConfig:
		events = append(events, event.ParallelStart{
			Event: event.TypeParallelStart, Node: node.ID,
			Targets: cfg.Targets, Timestamp: event.Now(),
		})
	case ir.FanInConfig:
		events = append(events, event.ParallelEnd{
			Event: event.TypeParallelEnd, Node: node.ID,
			Sources: cfg.Sources, Timestamp: event.Now(),
		})
	}
	return append(events, newNodeExitSuccess(node.ID))
}

func (s *simulator) emitEdgeTraverse(e *ir.Edge) {
	s.emit(buildEdgeTraverseEvent(e))
}

// emitFanOutIn handles parallel and fan-in nodes, returning true if the node was handled.
func (s *simulator) emitFanOutIn(node *ir.Node, enterEvt event.NodeEnter) bool {
	switch node.Config.(type) {
	case ir.ParallelConfig, ir.FanInConfig:
		for _, evt := range buildNodeEventSequence(node, enterEvt) {
			s.emit(evt)
		}
		return true
	default:
		return false
	}
}

func resolveModel(nodeModel, defaultModel string) string {
	if nodeModel != "" {
		return nodeModel
	}
	return defaultModel
}

func resolveProvider(nodeProvider, defaultProvider string) string {
	if nodeProvider != "" {
		return nodeProvider
	}
	return defaultProvider
}

func resolveFidelity(nodeFidelity, defaultFidelity string) string {
	if nodeFidelity != "" {
		return nodeFidelity
	}
	return defaultFidelity
}
