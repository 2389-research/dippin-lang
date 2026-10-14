package simulate

import (
	"strconv"
	"strings"

	"github.com/2389-research/dippin-lang/ir"
)

// gateAnswer records the answer typed at an --interactive human gate: the
// edge it picks, or nil when it matched none. A nil edge still decides the
// gate — tracker then routes it without a label (conditions, else the
// fallback edge), never by its default:.
type gateAnswer struct{ edge *ir.Edge }

// yesNoOptions are the answers tracker offers a yes_no gate
// (handlers/human.go executeYesNo).
var yesNoOptions = []string{"Yes", "No"}

// gateRoutes reports whether a node with these outgoing edges chooses
// between them, i.e. resolveNext consults its answer. A lone unconditional
// edge is taken without one.
func gateRoutes(edges []*ir.Edge) bool {
	return len(edges) > 1 || (len(edges) == 1 && edges[0].Condition != nil)
}

// answerEdge resolves an answer typed at a gate in the given mode to the
// edge it picks, following tracker's interviewer for that mode:
//
//   - freeform: the answer must equal an edge's choice: or label:
//     (matchGateAnswer; handlers/human.go matchFreeformLabel).
//   - yes_no: the answer picks Yes or No (by name or 1-based number); the
//     gate routes down an edge keyed by that word, if any.
//   - choice / unset: the answer picks an offered option — an edge's label,
//     or its target ID when unlabeled — by name or 1-based number
//     (interviewer_console.go matchConsoleChoice).
//
// Names compare case-insensitively and exactly: no substrings, so "n" picks
// neither "Continue" nor "[N] No". Returns nil when the answer picks nothing.
func answerEdge(edges []*ir.Edge, mode, answer string) *ir.Edge {
	switch mode {
	case "freeform":
		return matchGateAnswer(edges, answer)
	case "yes_no":
		return yesNoEdge(edges, answer)
	}
	if i := pickOption(choiceOptions(edges), answer); i >= 0 {
		return edges[i]
	}
	return nil
}

// yesNoEdge returns the edge keyed by the Yes/No option answer picks.
func yesNoEdge(edges []*ir.Edge, answer string) *ir.Edge {
	i := pickOption(yesNoOptions, answer)
	if i < 0 {
		return nil
	}
	return matchGateAnswer(edges, yesNoOptions[i])
}

// choiceOptions are the options tracker offers a choice gate, one per
// outgoing edge: its label, else its target ID (handlers/human.go
// executeChoice).
func choiceOptions(edges []*ir.Edge) []string {
	opts := make([]string, len(edges))
	for i, e := range edges {
		opts[i] = e.Label
		if opts[i] == "" {
			opts[i] = e.To
		}
	}
	return opts
}

// pickOption returns the index of the option answer names — exactly,
// ignoring case — or numbers (1-based), else -1.
func pickOption(options []string, answer string) int {
	for i, o := range options {
		if strings.EqualFold(o, answer) {
			return i
		}
	}
	return optionNumber(answer, len(options))
}

// optionNumber returns the 0-based index of a 1-based option number among
// count options, or -1 when answer is not such a number.
func optionNumber(answer string, count int) int {
	n, err := strconv.Atoi(answer)
	if err != nil || n < 1 || n > count {
		return -1
	}
	return n - 1
}

// promptOptions are the numbered options an --interactive prompt lists for
// a gate: none for freeform (open text), Yes/No for yes_no, else one per
// edge.
func promptOptions(mode string, edges []*ir.Edge) []string {
	switch mode {
	case "freeform":
		return nil
	case "yes_no":
		return yesNoOptions
	}
	return choiceOptions(edges)
}

// routingKey is the key tracker routes an edge by: its choice:, else its
// label:, else its target ID (pipeline/engine_edges.go edgeRoutingKey).
func routingKey(e *ir.Edge) string {
	if e.Choice != "" {
		return e.Choice
	}
	if e.Label != "" {
		return e.Label
	}
	return e.To
}
