package cost

// DefaultPricing returns a PricingTable with current model prices (USD per 1M tokens).
//
// Last verified: 2026-06-10 for the original providers; the rows dated
// "verified 2026-08-07" in inline comments (Claude 5, GPT-5.6, Gemini 3.5/3.6,
// grok-4.5/grok-build, and the zai/moonshot/minimax providers) were confirmed
// against official docs on that date. DeepSeek, Mistral, and Cohere were NOT
// re-verified on 2026-08-07 — their 2026-06-10 values stand.
//
// Sources:
//
//	Anthropic:  https://platform.claude.com/docs/en/about-claude/models/overview
//	Google:     https://ai.google.dev/gemini-api/docs/pricing
//	OpenAI:     https://developers.openai.com/api/docs/pricing
//	DeepSeek:   https://api-docs.deepseek.com/quick_start/pricing
//	xAI (Grok): https://docs.x.ai/docs/models
//	Mistral:    https://docs.mistral.ai/getting-started/models/models_overview/
//	Cohere:     https://docs.cohere.com/docs/models
//	Z.AI (GLM): https://docs.z.ai/guides/overview/pricing
//	Moonshot:   https://platform.kimi.ai/docs/pricing/chat-k3
//	MiniMax:    https://platform.minimax.io/docs/guides/pricing-paygo
//
// Notes on uncertainty (carried across this verification pass):
//   - Mistral nemo and mistral-small-2603: Mistral's official pricing tab is
//     JS-rendered and could not be read directly. Values below are unchanged
//     from the prior verification; third-party sources disagree and these
//     should be re-confirmed via the live page on the next pass.
//   - Cohere command-a-03-2025 and command-r7b-12-2024: Cohere removed
//     per-token pricing for these from the public pricing page; the values
//     below are unchanged from the prior verification.
//   - Gemini Pro tier: prompts >200K tokens are billed at 2x. This table
//     models only the ≤200K tier — callers should apply the multiplier where
//     it matters.
//   - OpenAI gpt-5.5 and gpt-5.4 / gpt-5.4 pro: prompts >272K tokens are
//     billed at 2x input / 1.5x output for the full session. Modeled at the
//     base tier only.
func DefaultPricing() PricingTable {
	gemini := geminiPricing()
	grok := grokPricing()
	return PricingTable{
		"zai":      zaiPricing(),
		"moonshot": moonshotPricing(),
		"kimi":     moonshotPricing(),
		"minimax":  minimaxPricing(),
		"anthropic": {
			// Claude 5 line (verified 2026-08-07, platform.claude.com models
			// overview). sonnet-5 durable list price is $3/$15; an introductory
			// $2/$10 rate applies through 2026-08-31 — the durable rate is modeled.
			"claude-opus-5":     {InputPer1M: 5.00, OutputPer1M: 25.00},
			"claude-sonnet-5":   {InputPer1M: 3.00, OutputPer1M: 15.00},
			"claude-opus-4-8":   {InputPer1M: 5.00, OutputPer1M: 25.00},
			"claude-opus-4-7":   {InputPer1M: 5.00, OutputPer1M: 25.00},
			"claude-opus-4-6":   {InputPer1M: 5.00, OutputPer1M: 25.00},
			"claude-sonnet-4-6": {InputPer1M: 3.00, OutputPer1M: 15.00},
			"claude-haiku-4-5":  {InputPer1M: 1.00, OutputPer1M: 5.00},
			// Fable 5 / Mythos 5 line (2026-06-09); base input/output rates
			// only (cache/batch/fast-mode tiers are out of dippin's schema).
			"claude-fable-5":    {InputPer1M: 10.00, OutputPer1M: 50.00},
			"claude-mythos-5":   {InputPer1M: 10.00, OutputPer1M: 50.00},
			"claude-sonnet-4-5": {InputPer1M: 3.00, OutputPer1M: 15.00},
			"claude-opus-4-5":   {InputPer1M: 5.00, OutputPer1M: 25.00},
			// Deprecated; retires 2026-08-05 → claude-opus-4-8.
			"claude-opus-4-1": {InputPer1M: 15.00, OutputPer1M: 75.00},
			// Both deprecated 2026-04-14, retire 2026-06-15.
			"claude-sonnet-4-0": {InputPer1M: 3.00, OutputPer1M: 15.00},
			"claude-opus-4-0":   {InputPer1M: 15.00, OutputPer1M: 75.00},
			// Retired 2026-02-19 on first-party API; Bedrock/Vertex passthrough rate.
			"claude-haiku-3-5": {InputPer1M: 0.80, OutputPer1M: 4.00},
		},
		"openai": {
			// GPT-5.6 line (verified 2026-08-07, developers.openai.com/api/docs/pricing).
			"gpt-5.6-sol":   {InputPer1M: 5.00, OutputPer1M: 30.00},
			"gpt-5.6-terra": {InputPer1M: 2.00, OutputPer1M: 12.00},
			"gpt-5.6-luna":  {InputPer1M: 0.20, OutputPer1M: 1.20},
			// Current frontier.
			"gpt-5.5":      {InputPer1M: 5.00, OutputPer1M: 30.00},
			"gpt-5.5-pro":  {InputPer1M: 30.00, OutputPer1M: 180.00},
			"gpt-5.4":      {InputPer1M: 2.50, OutputPer1M: 15.00},
			"gpt-5.4-pro":  {InputPer1M: 30.00, OutputPer1M: 180.00},
			"gpt-5.4-mini": {InputPer1M: 0.75, OutputPer1M: 4.50},
			"gpt-5.4-nano": {InputPer1M: 0.20, OutputPer1M: 1.25},
			// GPT-5 base line.
			"gpt-5":      {InputPer1M: 1.25, OutputPer1M: 10.00},
			"gpt-5-pro":  {InputPer1M: 15.00, OutputPer1M: 120.00},
			"gpt-5-mini": {InputPer1M: 0.25, OutputPer1M: 2.00},
			"gpt-5-nano": {InputPer1M: 0.05, OutputPer1M: 0.40},
			// Coding.
			"gpt-5.3-codex": {InputPer1M: 1.75, OutputPer1M: 14.00},
			// Previous-generation (still active).
			"gpt-5.2":      {InputPer1M: 1.75, OutputPer1M: 14.00},
			"gpt-5.2-pro":  {InputPer1M: 21.00, OutputPer1M: 168.00},
			"gpt-5.1":      {InputPer1M: 1.25, OutputPer1M: 10.00},
			"gpt-4.1":      {InputPer1M: 2.00, OutputPer1M: 8.00},
			"gpt-4.1-mini": {InputPer1M: 0.40, OutputPer1M: 1.60},
			"gpt-4o-mini":  {InputPer1M: 0.15, OutputPer1M: 0.60},
			"o3":           {InputPer1M: 2.00, OutputPer1M: 8.00},
			"o3-pro":       {InputPer1M: 20.00, OutputPer1M: 80.00},
			// Deprecated, scheduled retirement 2026-10-23.
			"gpt-4o":       {InputPer1M: 2.50, OutputPer1M: 10.00},
			"gpt-4.1-nano": {InputPer1M: 0.05, OutputPer1M: 0.20},
			"o3-mini":      {InputPer1M: 1.10, OutputPer1M: 4.40},
			"o4-mini":      {InputPer1M: 1.10, OutputPer1M: 4.40},
		},
		"google": gemini,
		"gemini": gemini,
		"deepseek": {
			// V4 models (current).
			"deepseek-v4-flash": {InputPer1M: 0.14, OutputPer1M: 0.28},
			// v4-pro list price; a 75% launch discount applies through 2026-05-31.
			"deepseek-v4-pro": {InputPer1M: 1.74, OutputPer1M: 3.48},
			// Compatibility aliases — sunset 2026-07-24 → deepseek-v4-flash.
			"deepseek-chat":     {InputPer1M: 0.14, OutputPer1M: 0.28},
			"deepseek-reasoner": {InputPer1M: 0.14, OutputPer1M: 0.28},
		},
		"xai":  grok,
		"grok": grok,
		"mistral": {
			"mistral-large-3":         {InputPer1M: 0.50, OutputPer1M: 1.50},
			"mistral-medium-3":        {InputPer1M: 0.40, OutputPer1M: 2.00},
			"mistral-medium-3-1-2508": {InputPer1M: 0.40, OutputPer1M: 2.00},
			"mistral-medium-3-5-2604": {InputPer1M: 1.50, OutputPer1M: 7.50},
			"mistral-small-2603":      {InputPer1M: 0.10, OutputPer1M: 0.30}, // uncertain; see header note
			"mistral-small":           {InputPer1M: 0.10, OutputPer1M: 0.30},
			"ministral-8b":            {InputPer1M: 0.10, OutputPer1M: 0.10},
			"ministral-3-3b-2512":     {InputPer1M: 0.10, OutputPer1M: 0.10},
			"ministral-3-8b-2512":     {InputPer1M: 0.15, OutputPer1M: 0.15},
			"ministral-3-14b-2512":    {InputPer1M: 0.20, OutputPer1M: 0.20},
			"codestral":               {InputPer1M: 0.30, OutputPer1M: 0.90},
			"magistral-medium":        {InputPer1M: 2.00, OutputPer1M: 5.00},
			"mistral-nemo":            {InputPer1M: 0.02, OutputPer1M: 0.04}, // uncertain; see header note
		},
		"cohere": {
			"command-a-03-2025":      {InputPer1M: 2.50, OutputPer1M: 10.00}, // uncertain; see header note
			"command-r-plus-08-2024": {InputPer1M: 2.50, OutputPer1M: 10.00},
			"command-r-08-2024":      {InputPer1M: 0.50, OutputPer1M: 1.50},
			"command-r7b-12-2024":    {InputPer1M: 0.0375, OutputPer1M: 0.15}, // uncertain; see header note
			// Bare aliases — Cohere docs resolve these to versions deprecated 2025-09-15.
			"command-r-plus": {InputPer1M: 2.50, OutputPer1M: 10.00},
			"command-r":      {InputPer1M: 0.50, OutputPer1M: 1.50},
			"command-r7b":    {InputPer1M: 0.0375, OutputPer1M: 0.15},
		},
	}
}

// grokPricing returns pricing for xAI Grok models.
//
// grok-4-1-fast-* IDs were retired 2026-05-15 — xAI silently redirects
// them to grok-4.3 server-side and bills at grok-4.3 rates. They remain
// in this table at grok-4.3 prices so cost analysis reflects what the
// runtime actually bills for workflows that still reference them.
func grokPricing() map[string]ModelPrice {
	return map[string]ModelPrice{
		// Verified 2026-08-07 (docs.x.ai/docs/models); base tier only —
		// grok-4.5 is $4/$12 for prompts ≥200k tokens.
		"grok-4.5":                     {InputPer1M: 2.00, OutputPer1M: 6.00},
		"grok-build-0.1":               {InputPer1M: 1.00, OutputPer1M: 2.00},
		"grok-4.3":                     {InputPer1M: 1.25, OutputPer1M: 2.50},
		"grok-4.20-0309-reasoning":     {InputPer1M: 1.25, OutputPer1M: 2.50},
		"grok-4.20-0309-non-reasoning": {InputPer1M: 1.25, OutputPer1M: 2.50},
		"grok-4.20-multi-agent-0309":   {InputPer1M: 1.25, OutputPer1M: 2.50},
		// Retired 2026-05-15; redirected to grok-4.3.
		"grok-4-1-fast-reasoning":     {InputPer1M: 1.25, OutputPer1M: 2.50},
		"grok-4-1-fast-non-reasoning": {InputPer1M: 1.25, OutputPer1M: 2.50},
	}
}

// geminiPricing returns pricing for all known Gemini models.
//
// Pro models (gemini-3.1-pro-preview, gemini-2.5-pro) charge 2x for prompts
// >200K tokens; this table reflects the base tier only.
func geminiPricing() map[string]ModelPrice {
	return map[string]ModelPrice{
		// Gemini 3.5/3.6 flash line (verified 2026-08-07, ai.google.dev/gemini-api/docs/pricing).
		"gemini-3.6-flash":                   {InputPer1M: 1.50, OutputPer1M: 7.50},
		"gemini-3.5-flash":                   {InputPer1M: 1.50, OutputPer1M: 9.00},
		"gemini-3.5-flash-lite":              {InputPer1M: 0.30, OutputPer1M: 2.50},
		"gemini-3.1-pro-preview":             {InputPer1M: 2.00, OutputPer1M: 12.00},
		"gemini-3.1-pro-preview-customtools": {InputPer1M: 2.00, OutputPer1M: 12.00},
		"gemini-3-flash-preview":             {InputPer1M: 0.50, OutputPer1M: 3.00},
		"gemini-3.1-flash-lite-preview":      {InputPer1M: 0.25, OutputPer1M: 1.50},
		"gemini-3.1-flash-lite":              {InputPer1M: 0.25, OutputPer1M: 1.50},
		"gemini-2.5-pro":                     {InputPer1M: 1.25, OutputPer1M: 10.00},
		"gemini-2.5-flash":                   {InputPer1M: 0.30, OutputPer1M: 2.50},
		"gemini-2.5-flash-lite":              {InputPer1M: 0.10, OutputPer1M: 0.40},
		// Deprecated, shuts down 2026-06-01.
		"gemini-2.0-flash": {InputPer1M: 0.10, OutputPer1M: 0.40},
	}
}

// zaiPricing returns pricing for Z.AI GLM models.
//
// Verified 2026-08-07 against https://docs.z.ai/guides/overview/pricing.
// Base list prices only (cached-input tiers are out of dippin's schema).
// glm-4.7-flash / glm-4.5-flash are free tiers and are omitted (a $0 row is
// indistinguishable from an unknown-model miss in cost output).
func zaiPricing() map[string]ModelPrice {
	return map[string]ModelPrice{
		"glm-5.2":        {InputPer1M: 1.40, OutputPer1M: 4.40},
		"glm-5.1":        {InputPer1M: 1.40, OutputPer1M: 4.40},
		"glm-5":          {InputPer1M: 1.00, OutputPer1M: 3.20},
		"glm-5-turbo":    {InputPer1M: 1.20, OutputPer1M: 4.00},
		"glm-4.7":        {InputPer1M: 0.60, OutputPer1M: 2.20},
		"glm-4.7-flashx": {InputPer1M: 0.07, OutputPer1M: 0.40},
		"glm-4.6":        {InputPer1M: 0.60, OutputPer1M: 2.20},
		"glm-4.5":        {InputPer1M: 0.60, OutputPer1M: 2.20},
		"glm-4.5-x":      {InputPer1M: 2.20, OutputPer1M: 8.90},
		"glm-4.5-air":    {InputPer1M: 0.20, OutputPer1M: 1.10},
		"glm-4.5-airx":   {InputPer1M: 1.10, OutputPer1M: 4.50},
	}
}

// moonshotPricing returns pricing for Moonshot AI (Kimi) models. Aliased under
// both the "moonshot" and "kimi" provider keys.
//
// Verified 2026-08-07 against https://platform.kimi.ai/docs/pricing/chat-k3.
// Input is the cache-miss rate (cache-hit tiers are out of dippin's schema).
func moonshotPricing() map[string]ModelPrice {
	return map[string]ModelPrice{
		"kimi-k3": {InputPer1M: 3.00, OutputPer1M: 15.00},
	}
}

// minimaxPricing returns pricing for MiniMax models.
//
// Verified 2026-08-07 against https://platform.minimax.io/docs/guides/pricing-paygo.
// Base tier only — MiniMax-M3 is $0.60/$2.40 for prompts >512k tokens. The
// published rates already reflect MiniMax's standing "permanent 50% off".
func minimaxPricing() map[string]ModelPrice {
	return map[string]ModelPrice{
		"MiniMax-M3":             {InputPer1M: 0.30, OutputPer1M: 1.20},
		"MiniMax-M2.7":           {InputPer1M: 0.30, OutputPer1M: 1.20},
		"MiniMax-M2.7-highspeed": {InputPer1M: 0.60, OutputPer1M: 2.40},
		"MiniMax-M2.5":           {InputPer1M: 0.30, OutputPer1M: 1.20},
		"MiniMax-M2.1":           {InputPer1M: 0.30, OutputPer1M: 1.20},
		"MiniMax-M2":             {InputPer1M: 0.30, OutputPer1M: 1.20},
	}
}
