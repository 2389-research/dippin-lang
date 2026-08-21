package pricing

import "encoding/json"

// fileEntry is the on-disk JSON shape of one catalog entry. Priced defaults to
// true; an unpriced entry sets "priced": false and omits the price fields.
type fileEntry struct {
	Provider        string   `json:"provider"`
	Model           string   `json:"model"`
	InputPerM       float64  `json:"input_per_m"`
	OutputPerM      float64  `json:"output_per_m"`
	CachedInputPerM float64  `json:"cached_input_per_m,omitempty"`
	CacheReadMult   float64  `json:"cache_read_mult,omitempty"`
	CacheWriteMult  float64  `json:"cache_write_mult,omitempty"`
	Aliases         []string `json:"aliases,omitempty"`
	Priced          *bool    `json:"priced,omitempty"` // nil = priced
	Deprecated      bool     `json:"deprecated,omitempty"`
	Source          string   `json:"source"`
	AsOf            string   `json:"as_of"`
	// Drift metadata (#264) + capability metadata (#267); all optional.
	Family        string   `json:"family,omitempty"`
	Rank          int      `json:"rank,omitempty"`
	Maturity      string   `json:"maturity,omitempty"`
	ContextWindow int      `json:"context_window,omitempty"`
	MaxOutput     int      `json:"max_output,omitempty"`
	Capabilities  []string `json:"capabilities,omitempty"`
	// Human-facing product name, verified from source (#285); optional (absent =
	// unknown, consumer derives or overlays).
	DisplayName string `json:"display_name,omitempty"`
}

type priceFile struct {
	ProviderAliases map[string]string `json:"provider_aliases"`
	Models          []fileEntry       `json:"models"`
}

// catalogIndex holds the parsed, lookup-ready catalog.
type catalogIndex struct {
	byProvider      map[string]map[string]ModelPrice // canonical provider -> model -> price
	byModel         map[string]ModelPrice            // model -> price (exact)
	byCanonModel    map[string]ModelPrice            // CanonicalModelID(model) -> price
	providerAliases map[string]string                // alias -> canonical
}

// index is the package-level catalog, built once from the embedded JSON. A
// malformed embed is a programming error (the file ships in the binary), so we
// panic rather than degrade silently.
var index = buildIndex()

func buildIndex() catalogIndex {
	var pf priceFile
	if err := json.Unmarshal(pricesJSON, &pf); err != nil {
		panic("pricing: invalid embedded prices.json: " + err.Error())
	}
	idx := catalogIndex{
		byProvider:      map[string]map[string]ModelPrice{},
		byModel:         map[string]ModelPrice{},
		byCanonModel:    map[string]ModelPrice{},
		providerAliases: pf.ProviderAliases,
	}
	for _, e := range pf.Models {
		idx.add(e)
	}
	return idx
}

// add inserts one file entry (and its aliases) into every lookup map.
func (idx *catalogIndex) add(e fileEntry) {
	p := ModelPrice{
		InputPerM:       e.InputPerM,
		OutputPerM:      e.OutputPerM,
		CachedInputPerM: e.CachedInputPerM,
		CacheReadMult:   e.CacheReadMult,
		CacheWriteMult:  e.CacheWriteMult,
		Aliases:         e.Aliases,
		Priced:          e.Priced == nil || *e.Priced,
		Deprecated:      e.Deprecated,
		Source:          e.Source,
		AsOf:            e.AsOf,
		Family:          e.Family,
		Rank:            e.Rank,
		Maturity:        e.Maturity,
		ContextWindow:   e.ContextWindow,
		MaxOutput:       e.MaxOutput,
		Capabilities:    e.Capabilities,
		DisplayName:     e.DisplayName,
	}
	idx.put(e.Provider, e.Model, p)
	for _, a := range e.Aliases {
		idx.put(e.Provider, a, p)
	}
}

// put registers one (provider, id) → price under all lookup maps.
func (idx *catalogIndex) put(provider, id string, p ModelPrice) {
	if idx.byProvider[provider] == nil {
		idx.byProvider[provider] = map[string]ModelPrice{}
	}
	idx.byProvider[provider][id] = p
	idx.byModel[id] = p
	idx.byCanonModel[CanonicalModelID(id)] = p
}
