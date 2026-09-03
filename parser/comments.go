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
// them instead of silently dropping them:
//
//   - a contiguous run of whole-line comments immediately above a node
//     declaration becomes that node's HeaderComment;
//   - comments on lines within a node's extent — trailing inline after a body
//     line, or whole-line — become the node's BodyComments, in source order;
//   - a contiguous run of whole-line comments immediately above an edge line
//     becomes the edge's HeaderComment, and the trailing inline comment on
//     the edge line itself becomes its TrailingComment.
//
// Comments on other lines (the dip pragma, workflow header fields, defaults,
// vars, inputs, stylesheet) have no IR slot yet and are dropped, as before.
// Node extents are capped at the next top-level line so a comment between a
// node and a later section attaches to the node, not the section. Header
// runs are maximal consecutive line-number runs ending at the declaration
// line minus one: a blank line between a comment and a node breaks the run
// (matching Workflow.HeaderComment's surrounding-blank trimming).
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

	// Node header blocks: the maximal run of whole-line comment lines ending
	// exactly at the declaration line minus one.
	for i := range p.nodeSpans {
		ns := &p.nodeSpans[i]
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

	// Edge header blocks + trailing inline comments. A run stops at a line
	// already claimed by a node header (a node declared after the edges
	// block in the source).
	for i := range p.edgeSpans {
		es := &p.edgeSpans[i]
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

	// Node body comments: every unclaimed comment on a line from the node's
	// declaration through the line before the next top-level construct.
	maxLine := len(p.lexer.lines)
	for i := range p.nodeSpans {
		ns := &p.nodeSpans[i]
		end := maxLine
		for _, l := range p.topLines {
			if l > ns.start && l < end {
				end = l
			}
		}
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
}
