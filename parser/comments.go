package parser

import "github.com/2389-research/dippin-lang/ir"

// nodeSpan pairs a parsed node with the 1-based source line of its
// declaration, for comment attachment (#259).
type nodeSpan struct {
	node  *ir.Node
	start int
}

// edgeSpan pairs a parsed edge with the 1-based source line of its
// declaration, for comment attachment (#259).
type edgeSpan struct {
	edge *ir.Edge
	line int
}

// recordNodeSpan remembers a node's declaration line so attachComments can
// scope its body comments to the node's source extent (#259).
func (p *Parser) recordNodeSpan(n *ir.Node, declLine int) {
	p.nodeSpans = append(p.nodeSpans, nodeSpan{node: n, start: declLine})
}

// attachComments transfers the comments the lexer stripped during
// tokenization onto the parsed workflow (#259), so `dippin fmt` can re-emit
// them instead of silently dropping them. It indexes the recorded comments by
// line, then runs three attachment phases: node header runs, edge header +
// trailing inline, and node body comments.
//
// Comments on other lines (the dip pragma, workflow header fields, defaults,
// vars, inputs, stylesheet) have no IR slot yet and are dropped, as before.
func (p *Parser) attachComments() {
	comms := p.lexer.Comments()
	if len(comms) == 0 {
		return
	}
	standalone := make(map[int]string, len(comms))
	trailing := make(map[int]string, len(comms))
	for _, c := range comms {
		if c.kind == commentStandalone {
			standalone[c.line] = c.text
		} else {
			trailing[c.line] = c.text
		}
	}
	claimed := make(map[int]bool, len(comms))

	for i := range p.nodeSpans {
		p.attachNodeHeader(&p.nodeSpans[i], standalone, claimed)
	}
	for i := range p.edgeSpans {
		p.attachEdgeComments(&p.edgeSpans[i], standalone, trailing, claimed)
	}
	for i := range p.nodeSpans {
		p.attachNodeBody(&p.nodeSpans[i], standalone, trailing, claimed)
	}
}

// attachNodeHeader sets the node's HeaderComment to the maximal run of
// whole-line comment lines ending exactly at the declaration line minus one.
// A blank line between a comment and a node breaks the run (matching
// Workflow.HeaderComment's surrounding-blank trimming). Lines in the run are
// marked claimed so later phases do not double-attach them.
func (p *Parser) attachNodeHeader(ns *nodeSpan, standalone map[int]string, claimed map[int]bool) {
	var block []string
	for l := ns.start - 1; ; l-- {
		text, ok := standalone[l]
		if !ok {
			break
		}
		block = append([]string{text}, block...)
		claimed[l] = true
	}
	ns.node.HeaderComment = block
}

// attachEdgeComments sets the edge's HeaderComment to the run of whole-line
// comment lines immediately above the edge line, and its TrailingComment to
// the inline comment on the edge line itself. The header run stops at a line
// already claimed by a node header (a node declared after the edges block in
// the source) as well as at any non-comment line.
func (p *Parser) attachEdgeComments(es *edgeSpan, standalone, trailing map[int]string, claimed map[int]bool) {
	var block []string
	for l := es.line - 1; ; l-- {
		if claimed[l] {
			break
		}
		text, ok := standalone[l]
		if !ok {
			break
		}
		block = append([]string{text}, block...)
		claimed[l] = true
	}
	es.edge.HeaderComment = block
	if text, ok := trailing[es.line]; ok {
		es.edge.TrailingComment = text
	}
}

// attachNodeBody sets the node's BodyComments to every unclaimed comment on a
// line within the node's extent, in source order. The extent runs from the
// declaration line to the line before the next top-level construct, so a
// comment between a node and a later section attaches to the node, not the
// section.
func (p *Parser) attachNodeBody(ns *nodeSpan, standalone, trailing map[int]string, claimed map[int]bool) {
	end := p.nodeBodyEnd(ns.start)
	var body []string
	for l := ns.start; l < end; l++ {
		if claimed[l] {
			continue
		}
		if text, ok := trailing[l]; ok {
			body = append(body, text)
		}
		if text, ok := standalone[l]; ok {
			body = append(body, text)
		}
	}
	ns.node.BodyComments = body
}

// nodeBodyEnd returns the one-past-the-last source line belonging to a node's
// extent: the smallest top-level line greater than start, or the input's line
// count when the node is the last top-level construct.
func (p *Parser) nodeBodyEnd(start int) int {
	end := len(p.lexer.lines)
	for _, l := range p.topLines {
		if l > start && l < end {
			end = l
		}
	}
	return end
}
