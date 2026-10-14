package simulate

import (
	"bufio"
	"fmt"
	"strings"

	"github.com/2389-research/dippin-lang/ir"
)

// handleHumanInteraction prompts the user when the node is a human node in interactive mode.
func (s *simulator) handleHumanInteraction(node *ir.Node) error {
	hc, ok := node.Config.(ir.HumanConfig)
	if !ok || !s.opts.Interactive || s.opts.Stdin == nil {
		return nil
	}
	if hc.Mode == "interview" {
		return s.handleInterviewMode(node, hc)
	}
	return s.handleChoiceFreeform(node, hc)
}

func (s *simulator) handleChoiceFreeform(node *ir.Node, hc ir.HumanConfig) error {
	edges := s.workflow.EdgesFrom(node.ID)
	response, err := s.promptInteractive(node, hc, edges)
	if err != nil {
		return fmt.Errorf("interactive prompt at %q: %w", node.ID, err)
	}
	s.updateContext("human_response", response)
	s.recordGateAnswer(node, hc, edges, response)
	return nil
}

// recordGateAnswer routes the gate by the typed answer, as tracker routes the
// answer its interviewer collects (see answerEdge for the matching rule).
// Only a gate that chooses between edges records one: a lone unconditional
// edge is taken without consulting it, so recording there would leak the
// answer to a later gate. A blank answer, or the gate's own default: (what
// end of input answers), leaves the gate to its default: routing.
func (s *simulator) recordGateAnswer(node *ir.Node, hc ir.HumanConfig, edges []*ir.Edge, response string) {
	if !gateRoutes(edges) || response == "" || response == gateDefault(hc) {
		return
	}
	e := answerEdge(edges, hc.Mode, response)
	s.answer = &gateAnswer{edge: e}
	if e == nil {
		s.warnf("  no edge from %q matches %q; the gate routes without a label\n", node.ID, response)
		return
	}
	s.updateContext("preferred_label", routingKey(e))
}

// warnf writes a note to stderr if available.
func (s *simulator) warnf(format string, args ...any) {
	if s.opts.Stderr != nil {
		fmt.Fprintf(s.opts.Stderr, format, args...)
	}
}

// stdinScanner returns the one scanner every --interactive prompt reads
// from. A scanner per prompt would drop the input its read-ahead buffered,
// so a later gate would see end of input.
func (s *simulator) stdinScanner() *bufio.Scanner {
	if s.stdin == nil {
		s.stdin = bufio.NewScanner(s.opts.Stdin)
	}
	return s.stdin
}

// handleInterviewMode collects multi-line input until a blank line or EOF.
func (s *simulator) handleInterviewMode(node *ir.Node, hc ir.HumanConfig) error {
	s.writeInterviewPrompt(node, hc)

	scanner := s.stdinScanner()
	var lines []string
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			break
		}
		lines = append(lines, line)
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("interview input at %q: %w", node.ID, err)
	}

	combined := strings.Join(lines, "\n")
	answersKey := hc.AnswersKey
	if answersKey == "" {
		answersKey = "interview_answers"
	}
	s.updateContext(answersKey, combined)
	s.updateContext("human_response", combined)
	return nil
}

// writeInterviewPrompt writes the interview prompt to stderr if available.
func (s *simulator) writeInterviewPrompt(node *ir.Node, hc ir.HumanConfig) {
	if s.opts.Stderr == nil {
		return
	}
	label := node.Label
	if label == "" {
		label = node.ID
	}
	fmt.Fprintf(s.opts.Stderr, "\n[HUMAN] %s (interview mode)\n", label)
	fmt.Fprintf(s.opts.Stderr, "  Enter answers line by line (blank line to finish):\n")
}

func (s *simulator) promptInteractive(node *ir.Node, hc ir.HumanConfig, edges []*ir.Edge) (string, error) {
	s.writeInteractivePrompt(node, hc, edges)

	scanner := s.stdinScanner()
	if scanner.Scan() {
		return strings.TrimSpace(scanner.Text()), nil
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}
	// EOF — answer with the gate's default: (where tracker honors one), or empty.
	return gateDefault(hc), nil
}

// writeInteractivePrompt writes the human prompt to stderr if available,
// listing the numbered options a choice or yes_no gate offers.
func (s *simulator) writeInteractivePrompt(node *ir.Node, hc ir.HumanConfig, edges []*ir.Edge) {
	if s.opts.Stderr == nil {
		return
	}
	label := node.Label
	if label == "" {
		label = node.ID
	}
	fmt.Fprintf(s.opts.Stderr, "\n[HUMAN] %s\n", label)
	for i, opt := range promptOptions(hc.Mode, edges) {
		fmt.Fprintf(s.opts.Stderr, "  %d) %s\n", i+1, opt)
	}
	if hc.Mode == "freeform" {
		fmt.Fprintf(s.opts.Stderr, "  Enter response: ")
	} else {
		fmt.Fprintf(s.opts.Stderr, "  Enter choice: ")
	}
}
