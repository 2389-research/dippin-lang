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
	response, err := s.promptInteractive(node, hc)
	if err != nil {
		return fmt.Errorf("interactive prompt at %q: %w", node.ID, err)
	}
	s.updateContext("human_response", response)
	// The answer also routes the gate, as tracker routes the answer it
	// collects: resolveByLabel matches preferred_label against edge labels.
	if response != "" {
		s.updateContext("preferred_label", response)
	}
	return nil
}

// handleInterviewMode collects multi-line input until a blank line or EOF.
func (s *simulator) handleInterviewMode(node *ir.Node, hc ir.HumanConfig) error {
	s.writeInterviewPrompt(node, hc)

	scanner := bufio.NewScanner(s.opts.Stdin)
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

func (s *simulator) promptInteractive(node *ir.Node, hc ir.HumanConfig) (string, error) {
	s.writeInteractivePrompt(node, hc)

	scanner := bufio.NewScanner(s.opts.Stdin)
	if scanner.Scan() {
		return strings.TrimSpace(scanner.Text()), nil
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}
	// EOF — return default or empty.
	return hc.Default, nil
}

// writeInteractivePrompt writes the human prompt to stderr if available.
func (s *simulator) writeInteractivePrompt(node *ir.Node, hc ir.HumanConfig) {
	if s.opts.Stderr == nil {
		return
	}
	label := node.Label
	if label == "" {
		label = node.ID
	}
	fmt.Fprintf(s.opts.Stderr, "\n[HUMAN] %s\n", label)
	if hc.Mode == "freeform" {
		fmt.Fprintf(s.opts.Stderr, "  Enter response: ")
	} else {
		fmt.Fprintf(s.opts.Stderr, "  Enter choice: ")
	}
}
