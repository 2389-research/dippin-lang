// The Mermaid exporter converts an ir.Workflow into a Mermaid flowchart
// (https://mermaid.js.org/syntax/flowchart.html) — node shapes and colors by
// node kind, edges labeled by their routing condition, with the start and exit
// nodes emphasized. Mermaid renders natively on GitHub, in most docs tooling,
// and in the web playground, so this doubles as a lightweight visualization.
package export

import (
	"fmt"
	"strings"

	"github.com/2389-research/dippin-lang/ir"
)

// ExportMermaid renders w as a top-down Mermaid flowchart.
func ExportMermaid(w *ir.Workflow) string {
	var b strings.Builder
	b.WriteString("flowchart TD\n")
	writeMermaidClassDefs(&b)
	for _, n := range w.Nodes {
		writeMermaidNode(&b, n)
	}
	for _, e := range w.Edges {
		writeMermaidEdge(&b, e)
	}
	writeMermaidRoleClasses(&b, w)
	return b.String()
}

// writeMermaidClassDefs emits the per-kind fill colors plus the start/exit
// emphasis classes. Colors are chosen to read on both light and dark backgrounds.
func writeMermaidClassDefs(b *strings.Builder) {
	defs := []string{
		"classDef agent fill:#e8f5e9,stroke:#43a047,color:#1b5e20;",
		"classDef human fill:#e3f2fd,stroke:#1e88e5,color:#0d47a1;",
		"classDef tool fill:#fff3e0,stroke:#fb8c00,color:#e65100;",
		"classDef subgraph_ fill:#f3e5f5,stroke:#8e24aa,color:#4a148c;",
		"classDef manager_loop fill:#ede7f6,stroke:#5e35b1,color:#311b92;",
		"classDef conditional fill:#fffde7,stroke:#fdd835,color:#f57f17;",
		"classDef fork fill:#eceff1,stroke:#607d8b,color:#263238;",
		"classDef startNode stroke-width:3px;",
		"classDef exitNode stroke-width:3px,stroke-dasharray:4 2;",
	}
	for _, d := range defs {
		b.WriteString("  " + d + "\n")
	}
}

// mermaidShapes maps a node kind to the {open, close} bracket pair around its
// quoted display text. Shapes: agent → stadium, human → parallelogram (input/
// gate), subgraph/manager_loop → subroutine, conditional → rhombus,
// parallel/fan_in → hexagon (fork/join). Anything else → rectangle (tool).
var mermaidShapes = map[ir.NodeKind][2]string{
	ir.NodeAgent:       {`(["`, `"])`},
	ir.NodeHuman:       {`[/"`, `"/]`},
	ir.NodeSubgraph:    {`[["`, `"]]`},
	ir.NodeManagerLoop: {`[["`, `"]]`},
	ir.NodeConditional: {`{"`, `"}`},
	ir.NodeParallel:    {`{{"`, `"}}`},
	ir.NodeFanIn:       {`{{"`, `"}}`},
}

// mermaidClasses maps a node kind to its classDef name.
var mermaidClasses = map[ir.NodeKind]string{
	ir.NodeAgent:       "agent",
	ir.NodeHuman:       "human",
	ir.NodeSubgraph:    "subgraph_",
	ir.NodeManagerLoop: "manager_loop",
	ir.NodeConditional: "conditional",
	ir.NodeParallel:    "fork",
	ir.NodeFanIn:       "fork",
}

// writeMermaidNode emits one node with a kind-appropriate shape and class.
func writeMermaidNode(b *strings.Builder, n *ir.Node) {
	br, ok := mermaidShapes[n.Kind]
	if !ok {
		br = [2]string{`["`, `"]`} // tool / default → rectangle
	}
	cls, ok := mermaidClasses[n.Kind]
	if !ok {
		cls = "tool"
	}
	fmt.Fprintf(b, "  %s%s%s%s:::%s\n", mermaidID(n.ID), br[0], mermaidText(n.ID), br[1], cls)
}

// writeMermaidEdge emits one edge, labeled by its routing condition when present.
func writeMermaidEdge(b *strings.Builder, e *ir.Edge) {
	from, to := mermaidID(e.From), mermaidID(e.To)
	label := mermaidEdgeLabel(e)
	if label == "" {
		fmt.Fprintf(b, "  %s --> %s\n", from, to)
		return
	}
	fmt.Fprintf(b, "  %s -->|\"%s\"| %s\n", from, label, to)
}

// mermaidEdgeLabel builds a short edge label: the explicit label if set, else a
// compact form of the routing condition (`ctx.outcome = success` → `success`),
// with a loop glyph appended for restart edges.
func mermaidEdgeLabel(e *ir.Edge) string {
	label := e.Label
	if label == "" && e.Condition != nil {
		label = compactCondition(e.Condition.Raw)
	}
	if e.Restart {
		label = strings.TrimSpace(label + " ⟳")
	}
	return mermaidText(label)
}

// compactCondition shortens a routing condition for a label: an
// `<outcome-channel> = <value>` test renders as just `<value>`; anything else is
// returned trimmed.
func compactCondition(raw string) string {
	raw = strings.TrimSpace(raw)
	i := strings.Index(raw, "=")
	if i < 0 || strings.Count(raw, "=") != 1 {
		return raw
	}
	if isOutcomeChannel(strings.TrimSpace(raw[:i])) {
		return strings.TrimSpace(strings.Trim(raw[i+1:], " ="))
	}
	return raw
}

func isOutcomeChannel(lhs string) bool {
	return lhs == "ctx.outcome" || lhs == "outcome" || lhs == "ctx.tool_marker"
}

// mermaidID sanitizes a node ID into a Mermaid-safe identifier (Mermaid node IDs
// cannot contain spaces or most punctuation).
func mermaidID(id string) string {
	var b strings.Builder
	for _, r := range id {
		if isMermaidIDChar(r) {
			b.WriteRune(r)
		} else {
			b.WriteByte('_')
		}
	}
	if b.Len() == 0 {
		return "n"
	}
	return b.String()
}

func isMermaidIDChar(r rune) bool {
	return r == '_' || isASCIIAlpha(r) || isASCIIDigit(r)
}

func isASCIIAlpha(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z')
}

func isASCIIDigit(r rune) bool {
	return r >= '0' && r <= '9'
}

// mermaidText escapes text for a quoted Mermaid string: quotes become HTML
// entities and newlines collapse to spaces so a label never breaks the syntax.
func mermaidText(s string) string {
	s = strings.ReplaceAll(s, "\"", "&quot;")
	s = strings.ReplaceAll(s, "\n", " ")
	return strings.TrimSpace(s)
}

// writeMermaidRoleClasses tags the start and exit nodes with their emphasis class
// (in addition to the kind class already applied inline).
func writeMermaidRoleClasses(b *strings.Builder, w *ir.Workflow) {
	if w.Start != "" {
		fmt.Fprintf(b, "  class %s startNode\n", mermaidID(w.Start))
	}
	if w.Exit != "" {
		fmt.Fprintf(b, "  class %s exitNode\n", mermaidID(w.Exit))
	}
}
