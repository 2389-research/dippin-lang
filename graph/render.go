package graph

import (
	"fmt"
	"strings"

	"github.com/2389-research/dippin-lang/ir"
)

const minBoxWidth = 10

// renderCompact renders a single-line compact representation.
func renderCompact(layers [][]string, w *ir.Workflow) string {
	parts := make([]string, 0, len(layers))
	for _, layer := range layers {
		parts = append(parts, formatCompactLayer(layer, w))
	}
	return strings.Join(parts, " \u2192 ") + "\n"
}

// formatCompactLayer formats a single layer for compact output.
func formatCompactLayer(layer []string, w *ir.Workflow) string {
	labels := layerLabels(layer, w)
	if len(labels) == 1 {
		return "[" + labels[0] + "]"
	}
	return "[" + strings.Join(labels, " | ") + "]"
}

// layerLabels returns the display labels for nodes in a layer.
func layerLabels(nodeIDs []string, w *ir.Workflow) []string {
	labels := make([]string, 0, len(nodeIDs))
	for _, id := range nodeIDs {
		labels = append(labels, nodeLabel(id, w))
	}
	return labels
}

// nodeLabel returns the display label for a node, falling back to its ID.
func nodeLabel(id string, w *ir.Workflow) string {
	for _, n := range w.Nodes {
		if n.ID == id && n.Label != "" {
			return n.Label
		}
	}
	return id
}

// renderFull renders the full box-drawing representation.
func renderFull(layers [][]string, w *ir.Workflow) string {
	var lines []string
	for i, layer := range layers {
		lines = appendLayerLines(lines, layer, w)
		if i < len(layers)-1 {
			lines = appendConnector(lines, layer)
		}
	}
	return strings.Join(lines, "\n") + "\n"
}

// appendLayerLines adds the rendered box lines for a layer.
func appendLayerLines(lines []string, layer []string, w *ir.Workflow) []string {
	boxes := renderLayerBoxes(layer, w)
	return append(lines, boxes...)
}

// appendConnector adds the vertical connector and arrow between layers.
func appendConnector(lines []string, _ []string) []string {
	lines = append(lines, "       \u2502")
	lines = append(lines, "       \u25bc")
	return lines
}

// renderLayerBoxes renders all node boxes in a layer, side by side.
func renderLayerBoxes(layer []string, w *ir.Workflow) []string {
	if len(layer) == 1 {
		return renderNodeBox(nodeLabel(layer[0], w))
	}
	return renderSideBySide(layer, w)
}

// renderSideBySide renders multiple node boxes side by side.
func renderSideBySide(layer []string, w *ir.Workflow) []string {
	boxes := make([][]string, 0, len(layer))
	for _, id := range layer {
		boxes = append(boxes, renderNodeBox(nodeLabel(id, w)))
	}
	return mergeBoxes(boxes)
}

// mergeBoxes combines multiple box renderings side by side with a space gap.
func mergeBoxes(boxes [][]string) []string {
	height := boxHeight(boxes)
	lines := make([]string, height)
	for i := 0; i < height; i++ {
		lines[i] = mergeBoxLine(boxes, i)
	}
	return lines
}

// boxHeight returns the max height of any box.
func boxHeight(boxes [][]string) int {
	max := 0
	for _, b := range boxes {
		if len(b) > max {
			max = len(b)
		}
	}
	return max
}

// mergeBoxLine merges one line across all boxes. Every box comes from
// renderNodeBox, which always returns exactly 3 lines, so lineIdx is always in
// range and every box contributes a part.
func mergeBoxLine(boxes [][]string, lineIdx int) string {
	parts := make([]string, 0, len(boxes))
	for _, b := range boxes {
		parts = append(parts, b[lineIdx])
	}
	return strings.Join(parts, " ")
}

// renderNodeBox returns the three lines for a single box-drawn node.
func renderNodeBox(label string) []string {
	width := boxWidth(label)
	return []string{
		fmt.Sprintf("  \u250c%s\u2510", strings.Repeat("\u2500", width)),
		fmt.Sprintf("  \u2502%s\u2502", centerText(label, width)),
		fmt.Sprintf("  \u2514%s\u2518", strings.Repeat("\u2500", width)),
	}
}

// boxWidth returns the interior width for a node box.
func boxWidth(label string) int {
	w := len(label) + 2
	if w < minBoxWidth {
		return minBoxWidth
	}
	return w
}

// centerText centers text within a given width, padding with spaces.
func centerText(text string, width int) string {
	if len(text) >= width {
		return text
	}
	pad := width - len(text)
	left := pad / 2
	right := pad - left
	return strings.Repeat(" ", left) + text + strings.Repeat(" ", right)
}
