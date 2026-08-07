package validator

import (
	"fmt"
	"sort"
	"strings"

	"github.com/2389-research/dippin-lang/ir"
)

// knownModelProviders lists known valid model/provider combinations.
// This is a best-effort catalog — unknown combinations produce a warning,
// not an error, since new models may be added at any time.
//
// Last verified: 2026-06-10
//
// Sources:
//
//	Anthropic:  https://platform.claude.com/docs/en/docs/about-claude/models
//	Google:     https://ai.google.dev/gemini-api/docs/models
//	OpenAI:     https://developers.openai.com/api/docs/models/all
//	DeepSeek:   https://api-docs.deepseek.com/quick_start/pricing
//	xAI (Grok): https://docs.x.ai/developers/models
//	            https://docs.x.ai/developers/migration/may-15-retirement
//	Mistral:    https://docs.mistral.ai/getting-started/models/models_overview/
//	Cohere:     https://docs.cohere.com/docs/models
var knownModelProviders = map[string]map[string]bool{
	"anthropic": {
		// Claude 5 line (verified 2026-08-07, platform.claude.com models overview).
		"claude-opus-5":     true,
		"claude-sonnet-5":   true,
		"claude-opus-4-8":   true,
		"claude-opus-4-7":   true,
		"claude-opus-4-6":   true,
		"claude-sonnet-4-6": true,
		"claude-haiku-4-5":  true,
		// Fable 5 / Mythos 5 line (2026-06-09). Mythos 5 and the research
		// preview are invite-only (Project Glasswing) but are real IDs —
		// listed so approved users don't trip a spurious DIP108.
		"claude-fable-5":        true,
		"claude-mythos-5":       true,
		"claude-mythos-preview": true,
		// Legacy models still available via API.
		"claude-sonnet-4-5": true,
		"claude-opus-4-5":   true,
		// Deprecated; retires 2026-08-05 → claude-opus-4-8.
		"claude-opus-4-1": true,
		// Deprecated 2026-04-14, retires 2026-06-15 → claude-sonnet-4-6.
		"claude-sonnet-4-0": true,
		// Deprecated 2026-04-14, retires 2026-06-15 → claude-opus-4-8.
		"claude-opus-4-0": true,
		// Retired 2026-02-19 on first-party API; remains on Bedrock/Vertex AI.
		"claude-haiku-3-5": true,
	},
	"google": geminiModels(),
	"gemini": geminiModels(),
	"openai": {
		// GPT-5.6 line (verified 2026-08-07, developers.openai.com/api/docs/pricing).
		"gpt-5.6-sol":   true,
		"gpt-5.6-terra": true,
		"gpt-5.6-luna":  true,
		// Current frontier (May 2026).
		"gpt-5.5":      true,
		"gpt-5.5-pro":  true,
		"gpt-5.4":      true,
		"gpt-5.4-pro":  true,
		"gpt-5.4-mini": true,
		"gpt-5.4-nano": true,
		// GPT-5 base line.
		"gpt-5":      true,
		"gpt-5-pro":  true,
		"gpt-5-mini": true,
		"gpt-5-nano": true,
		// Coding line.
		"gpt-5.3-codex": true,
		// Previous-generation (still active).
		"gpt-5.2":      true,
		"gpt-5.2-pro":  true,
		"gpt-5.1":      true,
		"gpt-4.1":      true,
		"gpt-4.1-mini": true,
		"gpt-4o-mini":  true,
		// Reasoning line (still active).
		"o3":     true,
		"o3-pro": true,
		// Deprecated, scheduled retirement 2026-10-23.
		"gpt-4o":       true,
		"gpt-4.1-nano": true,
		"o3-mini":      true,
		"o4-mini":      true,
	},
	"deepseek": {
		// V4 models (current).
		"deepseek-v4-flash": true,
		"deepseek-v4-pro":   true,
		// Compatibility aliases, scheduled deprecation 2026-07-24 → deepseek-v4-flash.
		"deepseek-chat":     true,
		"deepseek-reasoner": true,
	},
	"xai":  grokModels(),
	"grok": grokModels(),
	"zai":  zaiModels(),
	// Moonshot AI (Kimi) — aliased under both provider keys.
	"moonshot": moonshotModels(),
	"kimi":     moonshotModels(),
	"minimax":  minimaxModels(),
	// Alibaba Qwen — recognized so DIP108 doesn't fire (IDs verified 2026-08-07,
	// alibabacloud.com/help/en/model-studio/models). Not in the cost table: the
	// international per-token USD pricing is console-gated and could not be
	// verified from an official page (tracked in a follow-up).
	"qwen": {
		"qwen3.7-max":   true,
		"qwen3.7-plus":  true,
		"qwen3.6-flash": true,
	},
	"mistral": {
		"mistral-large-3":         true,
		"mistral-medium-3":        true,
		"mistral-medium-3-1-2508": true, // Mistral Medium 3.1 (Aug 2025)
		"mistral-medium-3-5-2604": true, // Mistral Medium 3.5, new flagship-class (Apr 2026)
		"mistral-small-2603":      true, // Mistral Small 4 (March 2026)
		"mistral-small":           true, // Mistral Small 3.1 (legacy)
		"ministral-8b":            true,
		"ministral-3-3b-2512":     true, // Ministral 3 generation (Dec 2025)
		"ministral-3-8b-2512":     true,
		"ministral-3-14b-2512":    true,
		"codestral":               true,
		"magistral-medium":        true,
		"mistral-nemo":            true,
	},
	"cohere": {
		"command-a-03-2025":      true, // Current flagship
		"command-r-plus-08-2024": true,
		"command-r-08-2024":      true,
		"command-r7b-12-2024":    true,
		// Bare aliases — Cohere docs list these as resolving to versions deprecated
		// 2025-09-15. Keep callable for now; prefer the dated IDs above.
		"command-r-plus": true,
		"command-r":      true,
		"command-r7b":    true,
	},
}

// geminiModels returns the set of known Gemini model IDs.
func geminiModels() map[string]bool {
	return map[string]bool{
		// Gemini 3.5/3.6 flash line (verified 2026-08-07, ai.google.dev/gemini-api/docs/pricing).
		"gemini-3.6-flash":      true,
		"gemini-3.5-flash":      true,
		"gemini-3.5-flash-lite": true,
		// Gemini 3.x
		"gemini-3.1-pro-preview":             true,
		"gemini-3.1-pro-preview-customtools": true,
		"gemini-3-flash-preview":             true,
		"gemini-3.1-flash-lite-preview":      true,
		"gemini-3.1-flash-lite":              true, // GA promotion of the preview variant.
		// Gemini 2.x — gemini-2.5-* are stable/GA.
		"gemini-2.5-pro":        true,
		"gemini-2.5-flash":      true,
		"gemini-2.5-flash-lite": true,
		// Deprecated, shuts down 2026-06-01.
		"gemini-2.0-flash": true,
	}
}

// zaiModels returns the set of known Z.AI GLM model IDs (verified 2026-08-07,
// docs.z.ai/guides/overview/pricing). Includes the free-tier flash models, which
// are recognized here even though they carry no cost-table row.
func zaiModels() map[string]bool {
	return map[string]bool{
		"glm-5.2":        true,
		"glm-5.1":        true,
		"glm-5":          true,
		"glm-5-turbo":    true,
		"glm-4.7":        true,
		"glm-4.7-flashx": true,
		"glm-4.7-flash":  true,
		"glm-4.6":        true,
		"glm-4.5":        true,
		"glm-4.5-x":      true,
		"glm-4.5-air":    true,
		"glm-4.5-airx":   true,
		"glm-4.5-flash":  true,
	}
}

// moonshotModels returns the set of known Moonshot AI (Kimi) model IDs (verified
// 2026-08-07, platform.kimi.ai/docs/pricing).
func moonshotModels() map[string]bool {
	return map[string]bool{
		"kimi-k3": true,
	}
}

// minimaxModels returns the set of known MiniMax model IDs (verified 2026-08-07,
// platform.minimax.io/docs/guides/pricing-paygo).
func minimaxModels() map[string]bool {
	return map[string]bool{
		"MiniMax-M3":             true,
		"MiniMax-M2.7":           true,
		"MiniMax-M2.7-highspeed": true,
		"MiniMax-M2.5":           true,
		"MiniMax-M2.1":           true,
		"MiniMax-M2":             true,
	}
}

// grokModels returns the set of known xAI Grok model IDs.
//
// grok-4-1-fast-reasoning and grok-4-1-fast-non-reasoning were retired
// 2026-05-15 — requests are silently redirected to grok-4.3 by xAI and
// billed at grok-4.3 rates. They remain in the catalog because they are
// still functionally callable; surfacing DIP108 ("unknown model") on
// them would be indistinguishable from a typo. A future, more specific
// "deprecated alias" diagnostic can replace this comment.
func grokModels() map[string]bool {
	return map[string]bool{
		// Verified 2026-08-07 (docs.x.ai/docs/models).
		"grok-4.5":                     true,
		"grok-build-0.1":               true,
		"grok-4.3":                     true, // Current flagship (Apr 2026).
		"grok-4.20-0309-reasoning":     true,
		"grok-4.20-0309-non-reasoning": true,
		"grok-4.20-multi-agent-0309":   true,
		// Retired 2026-05-15; xAI redirects to grok-4.3.
		"grok-4-1-fast-reasoning":     true,
		"grok-4-1-fast-non-reasoning": true,
	}
}

// ExtraModels is a user-supplied catalog of provider→model sets that extends the
// known model catalog for a single lint invocation. It is passed via Options and
// never mutates the package-level base catalog, so it cannot leak across calls.
type ExtraModels map[string]map[string]bool

// ParseExtraModels parses user-provided entries into a scoped ExtraModels catalog.
// Format: "provider:model1,model2;provider2:model3". Malformed or empty entries
// are silently skipped. The result is always safe to pass to LintWithOptions.
func ParseExtraModels(spec string) ExtraModels {
	extra := ExtraModels{}
	for _, entry := range strings.Split(spec, ";") {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}
		extra.addEntry(entry)
	}
	return extra
}

// addEntry parses a single "provider:model1,model2" entry into the catalog.
// Silently ignores entries with empty provider or no valid model names.
func (e ExtraModels) addEntry(entry string) {
	parts := strings.SplitN(entry, ":", 2)
	if len(parts) != 2 {
		return
	}
	provider := strings.TrimSpace(parts[0])
	if provider == "" {
		return
	}
	models := parseModelNames(parts[1])
	if len(models) == 0 {
		return
	}
	e.addModels(provider, models)
}

// addModels registers models under the given provider name in the catalog.
func (e ExtraModels) addModels(provider string, models []string) {
	if e[provider] == nil {
		e[provider] = make(map[string]bool)
	}
	for _, m := range models {
		e[provider][m] = true
	}
}

// parseModelNames splits a comma-separated model list, trimming whitespace
// and discarding empty entries.
func parseModelNames(raw string) []string {
	var models []string
	for _, m := range strings.Split(raw, ",") {
		m = strings.TrimSpace(m)
		if m != "" {
			models = append(models, m)
		}
	}
	return models
}

// lintModelProvider checks DIP108: model/provider combinations should be
// in the known catalog (base ∪ extra). Unknown combinations may indicate typos.
func lintModelProvider(w *ir.Workflow, extra ExtraModels) []Diagnostic {
	var diags []Diagnostic
	for _, n := range w.Nodes {
		diags = append(diags, checkNodeModelProvider(w, n, extra)...)
	}
	return diags
}

// checkNodeModelProvider validates the model/provider for a single node.
func checkNodeModelProvider(w *ir.Workflow, n *ir.Node, extra ExtraModels) []Diagnostic {
	cfg, ok := n.Config.(ir.AgentConfig)
	if !ok {
		return nil
	}
	model, provider := resolveModelProvider(cfg, w)
	if model == "" || provider == "" {
		return nil
	}
	return validateModelProvider(n, model, provider, extra)
}

// resolveModelProvider resolves model and provider using node config and workflow defaults.
func resolveModelProvider(cfg ir.AgentConfig, w *ir.Workflow) (model, provider string) {
	model = cfg.Model
	provider = cfg.Provider
	if model == "" {
		model = w.Defaults.Model
	}
	if provider == "" {
		provider = w.Defaults.Provider
	}
	return
}

// validateModelProvider checks if a model/provider combination is known,
// consulting the base catalog plus the scoped extra catalog.
func validateModelProvider(n *ir.Node, model, provider string, extra ExtraModels) []Diagnostic {
	if !providerKnown(provider, extra) {
		return []Diagnostic{{
			Code:     DIP108,
			Severity: SeverityWarning,
			Message:  fmt.Sprintf("node %q uses unknown provider %q", n.ID, provider),
			Location: n.Source,
			Help:     fmt.Sprintf("known providers: %s", knownProviderList(extra)),
		}}
	}
	if !modelKnown(provider, model, extra) {
		return []Diagnostic{{
			Code:     DIP108,
			Severity: SeverityWarning,
			Message:  fmt.Sprintf("node %q uses unknown model %q for provider %q", n.ID, model, provider),
			Location: n.Source,
			Help:     fmt.Sprintf("known models for %s: %s", provider, knownModelList(provider, extra)),
		}}
	}
	return nil
}

// providerKnown reports whether the provider appears in the base catalog or the
// scoped extra catalog.
func providerKnown(provider string, extra ExtraModels) bool {
	if _, ok := knownModelProviders[provider]; ok {
		return true
	}
	_, ok := extra[provider]
	return ok
}

// modelKnown reports whether the model is registered for the provider in either
// the base catalog or the scoped extra catalog.
func modelKnown(provider, model string, extra ExtraModels) bool {
	if knownModelProviders[provider][model] || extra[provider][model] {
		return true
	}
	// Fall back to a version-separator-insensitive match so a dotted ID
	// (claude-haiku-4.5, the Vercel AI Gateway spelling) is recognized as the
	// dashed catalog key (claude-haiku-4-5), and vice versa (issue #188).
	return modelKnownNormalized(knownModelProviders[provider], model) ||
		modelKnownNormalized(extra[provider], model)
}

// modelKnownNormalized reports whether any key in catalog matches model once
// their version separators are folded (dots to dashes). See #188.
func modelKnownNormalized(catalog map[string]bool, model string) bool {
	want := canonicalModelID(model)
	for k := range catalog {
		if canonicalModelID(k) == want {
			return true
		}
	}
	return false
}

// canonicalModelID folds the model version separator so dotted and dashed
// spellings compare equal (claude-haiku-4.5 == claude-haiku-4-5). See #188.
func canonicalModelID(model string) string {
	return strings.ReplaceAll(model, ".", "-")
}

// knownProviderList returns a sorted comma-separated list of known providers,
// merging the base catalog with the scoped extra catalog.
func knownProviderList(extra ExtraModels) string {
	seen := map[string]bool{}
	for p := range knownModelProviders {
		seen[p] = true
	}
	for p := range extra {
		seen[p] = true
	}
	return sortedJoin(seen)
}

// knownModelList returns a sorted comma-separated list of known models for a
// provider, merging the base catalog with the scoped extra catalog.
func knownModelList(provider string, extra ExtraModels) string {
	seen := map[string]bool{}
	for m := range knownModelProviders[provider] {
		seen[m] = true
	}
	for m := range extra[provider] {
		seen[m] = true
	}
	return sortedJoin(seen)
}

// sortedJoin returns the keys of seen as a sorted, comma-separated string.
func sortedJoin(seen map[string]bool) string {
	list := make([]string, 0, len(seen))
	for k := range seen {
		list = append(list, k)
	}
	sort.Strings(list)
	return strings.Join(list, ", ")
}

// validReasoningEfforts is the set of reasoning effort levels recognized by LLM providers.
// Levels: none (disabled), minimal, low, medium, high, xhigh (extra-high), max.
// Not all providers support all levels — e.g., Anthropic Opus 4.7+ supports xhigh/max,
// OpenAI GPT-5.4 supports none/minimal/low/medium/high/xhigh, older o3 only low/medium/high.
var validReasoningEfforts = map[string]bool{
	"none":    true,
	"minimal": true,
	"low":     true,
	"medium":  true,
	"high":    true,
	"xhigh":   true,
	"max":     true,
}

// lintReasoningEffort checks DIP119: reasoning_effort must be a recognized level.
func lintReasoningEffort(w *ir.Workflow) []Diagnostic {
	var diags []Diagnostic
	for _, n := range w.Nodes {
		cfg, ok := n.Config.(ir.AgentConfig)
		if !ok {
			continue
		}
		if r := cfg.ReasoningEffort; r != "" && !validReasoningEfforts[r] {
			diags = append(diags, Diagnostic{
				Code:     DIP119,
				Severity: SeverityWarning,
				Message:  fmt.Sprintf("node %q has reasoning_effort %q which is not a recognized level", n.ID, r),
				Location: n.Source,
				Help:     "valid levels: none, minimal, low, medium, high, xhigh, max",
			})
		}
	}
	return diags
}
