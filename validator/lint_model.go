package validator

import (
	"fmt"
	"sort"
	"strings"

	"github.com/2389-research/dippin-lang/ir"
	"github.com/2389-research/dippin-lang/pricing"
)

// The base model/provider catalog (DIP108) is the embedded pricing catalog
// (github.com/2389-research/dippin-lang/pricing) — the single source of truth
// for models and their prices. This file adds only the per-invocation
// ExtraModels overlay and the DIP108 checks over the union of the two.

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
// consulting the base catalog plus the scoped extra catalog. A family alias
// (family@selector, #264) is handled first: a resolvable alias is valid (no
// diagnostic), an unresolvable one is DIP162.
func validateModelProvider(n *ir.Node, model, provider string, extra ExtraModels) []Diagnostic {
	if diags, isAlias := aliasModelDiag(n, model, provider); isAlias {
		return diags
	}
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
	return deprecatedModelDiag(n, model, provider)
}

// deprecatedModelDiag emits DIP161 when a known model resolves to a pricing
// catalog entry flagged Deprecated (retired on the first-party API, still
// billed on passthrough platforms like Bedrock/Vertex). Only the base pricing
// catalog carries the flag — extra-catalog models never trip this.
func deprecatedModelDiag(n *ir.Node, model, provider string) []Diagnostic {
	mp, ok := pricing.LookupProvider(provider, model)
	if !ok || !mp.Deprecated {
		return nil
	}
	return []Diagnostic{{
		Code:     DIP161,
		Severity: SeverityWarning,
		Message:  fmt.Sprintf("model %q (provider %q) is deprecated — retired on the first-party API (still billed on passthrough); pin a current model", model, provider),
		Location: n.Source,
		Help:     "pin a current, non-deprecated model for this provider",
	}}
}

// aliasModelDiag handles a family-alias model value (family@selector). isAlias
// reports whether model is an alias at all; when it is, a resolvable alias
// yields no diagnostic and an unresolvable one yields DIP162. A non-alias value
// returns isAlias=false so the caller falls through to the DIP108 catalog check.
func aliasModelDiag(n *ir.Node, model, provider string) (diags []Diagnostic, isAlias bool) {
	_, resolved, isAlias := pricing.ResolveModelRef(provider, model)
	if !isAlias || resolved {
		return nil, isAlias
	}
	return []Diagnostic{{
		Code:     DIP162,
		Severity: SeverityWarning,
		Message:  fmt.Sprintf("node %q model alias %q resolves to no eligible model in that family/selector for provider %q", n.ID, model, provider),
		Location: n.Source,
		Help:     "check the family and selector (latest, stable, sota); the family may be unknown or all its members deprecated/preview",
	}}, true
}

// providerKnown reports whether the provider appears in the pricing catalog
// (alias-resolved) or the scoped extra catalog.
func providerKnown(provider string, extra ExtraModels) bool {
	if pricing.KnownProvider(provider) {
		return true
	}
	_, ok := extra[provider]
	return ok
}

// modelKnown reports whether the model is registered for the provider in the
// pricing catalog or the scoped extra catalog. The pricing lookup already folds
// the version separator (claude-haiku-4.5 == claude-haiku-4-5, issue #188) and
// resolves provider aliases; the extra overlay gets the same fold here.
func modelKnown(provider, model string, extra ExtraModels) bool {
	if _, ok := pricing.LookupProvider(provider, model); ok {
		return true
	}
	if extra[provider][model] {
		return true
	}
	return modelKnownNormalized(extra[provider], model)
}

// modelKnownNormalized reports whether any key in an extra-catalog entry matches
// model once their version separators are folded (dots to dashes). See #188.
func modelKnownNormalized(catalog map[string]bool, model string) bool {
	want := pricing.CanonicalModelID(model)
	for k := range catalog {
		if pricing.CanonicalModelID(k) == want {
			return true
		}
	}
	return false
}

// knownProviderList returns a sorted comma-separated list of known providers,
// merging the pricing catalog with the scoped extra catalog.
func knownProviderList(extra ExtraModels) string {
	seen := map[string]bool{}
	for _, p := range pricing.ProviderNames() {
		seen[p] = true
	}
	for p := range extra {
		seen[p] = true
	}
	return sortedJoin(seen)
}

// knownModelList returns a sorted comma-separated list of known models for a
// provider, merging the pricing catalog with the scoped extra catalog.
func knownModelList(provider string, extra ExtraModels) string {
	seen := map[string]bool{}
	for _, m := range pricing.ModelIDs(provider) {
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
