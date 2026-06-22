package optimize

import "strings"

// Complexity thresholds for prompt scoring.
const (
	complexityThresholdHigh = 30
)

// promptComplexity scores a prompt's complexity using a weighted scoring table.
// Higher scores indicate more complex prompts that may benefit from stronger models.
func promptComplexity(prompt string) int {
	score := 0
	for _, indicator := range complexityIndicators {
		score += indicator.score(prompt)
	}
	return score
}

// complexityIndicator represents a single complexity signal.
type complexityIndicator struct {
	score func(string) int
}

// complexityIndicators is the scoring table for prompt complexity.
var complexityIndicators = []complexityIndicator{
	// Length: longer prompts are generally more complex.
	{func(p string) int { return len(p) / 100 }},

	// Line count: multi-line prompts suggest multi-step instructions.
	{func(p string) int { return strings.Count(p, "\n") }},

	// Keyword signals for analytical/reasoning tasks.
	{keywordCounter([]string{"analyze", "evaluate", "compare", "reason", "trade-off"}, 5)},

	// Keyword signals for structured output.
	{keywordCounter([]string{"json", "xml", "schema", "format", "structured"}, 3)},

	// Keyword signals for multi-step reasoning.
	{keywordCounter([]string{"step 1", "step 2", "first", "then", "finally"}, 3)},

	// Code-related indicators.
	{keywordCounter([]string{"code", "implement", "refactor", "debug", "architect"}, 4)},
}

// containsAny reports whether the lowercased s contains any of subs as a substring.
func containsAny(s string, subs []string) bool {
	lower := strings.ToLower(s)
	for _, sub := range subs {
		if strings.Contains(lower, sub) {
			return true
		}
	}
	return false
}

// keywordCounter returns a scoring function that adds points per keyword match.
func keywordCounter(keywords []string, points int) func(string) int {
	return func(prompt string) int {
		lower := strings.ToLower(prompt)
		total := 0
		for _, kw := range keywords {
			if strings.Contains(lower, kw) {
				total += points
			}
		}
		return total
	}
}

// isExpensiveModel returns true for high-tier models.
func isExpensiveModel(model string) bool {
	return containsAny(model, expensiveModels)
}

// isCheapModel returns true for budget-tier models.
func isCheapModel(model string) bool {
	return containsAny(model, cheapModels)
}

// isBookkeepingPrompt detects prompts that are simple bookkeeping tasks.
func isBookkeepingPrompt(prompt string) bool {
	return containsAny(prompt, bookkeepingKeywords)
}

// suggestCheaperModel returns a budget model for the given provider.
func suggestCheaperModel(provider string) string {
	if s, ok := cheaperModelMap[provider]; ok {
		return s
	}
	return "claude-haiku-4-5"
}

// suggestStrongerModel returns a capable model for the given provider.
func suggestStrongerModel(provider string) string {
	if s, ok := strongerModelMap[provider]; ok {
		return s
	}
	return "claude-sonnet-4-6"
}

var expensiveModels = []string{"opus", "gpt-5.2", "gpt-5.4"}
var cheapModels = []string{"haiku", "mini"}
var bookkeepingKeywords = []string{"commit", "summary", "summarize", "cleanup", "clean up", "log", "record", "archive"}

var cheaperModelMap = map[string]string{
	"anthropic": "claude-haiku-4-5",
	"openai":    "gpt-4o-mini",
}

var strongerModelMap = map[string]string{
	"anthropic": "claude-sonnet-4-6",
	"openai":    "gpt-4o",
}
