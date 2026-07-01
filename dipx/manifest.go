package dipx

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
)

// Manifest is the parsed manifest.json.
//
// JSON tags use the canonical wire-format keys (`format_version`, `entry`,
// `files`) so external tools that pipe `dippin inspect --format=json` output
// can decode it back into a Manifest without renaming fields. Internal decode
// goes through decodeStrictly's private `raw` struct (which has its own tags),
// so these tags are decoupled from manifest-strict-decode behavior.
type Manifest struct {
	FormatVersion int             `json:"format_version"`
	Entry         string          `json:"entry"`
	Files         []ManifestEntry `json:"files"`
}

// ManifestEntry is one record in Manifest.Files.
type ManifestEntry struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
}

const (
	maxManifestSize  = 1 << 20 // 1 MB
	maxManifestDepth = 32
	bomPrefix        = "\ufeff"
)

// decodeManifest parses raw manifest bytes per the spec's JSON encoding rules.
// It rejects: BOM, oversized input (>1MB), duplicate keys (any level),
// trailing data, depth > 32, version != integer, presence of reserved
// "signatures" key at the top level. Field-presence and shape rules (entry must match a
// files[] entry, sha256 format, path canonicalization, etc.) are enforced
// separately by verifyManifestShape.
func decodeManifest(raw []byte) (Manifest, error) {
	if err := preflightManifestBytes(raw); err != nil {
		return Manifest{}, err
	}
	if err := validateJSONStructure(raw); err != nil {
		return Manifest{}, err
	}
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	var m Manifest
	if err := decodeStrictly(dec, &m); err != nil {
		return Manifest{}, err
	}
	if err := assertNoTrailingTokens(dec); err != nil {
		return Manifest{}, err
	}
	return m, nil
}

// preflightManifestBytes enforces the byte-level prerequisites (size cap,
// BOM rejection) before any JSON parsing is attempted.
func preflightManifestBytes(raw []byte) error {
	if len(raw) > maxManifestSize {
		return newError(ErrManifestInvalid, "", "manifest exceeds 1MB", nil)
	}
	if bytes.HasPrefix(raw, []byte(bomPrefix)) {
		return newError(ErrManifestInvalid, "", "BOM present", nil)
	}
	return nil
}

// assertNoTrailingTokens returns an error if dec has more tokens after the
// top-level value has been consumed.
func assertNoTrailingTokens(dec *json.Decoder) error {
	if dec.More() {
		return newError(ErrManifestInvalid, "", "trailing data after JSON object", nil)
	}
	if _, err := dec.Token(); err != nil && !errors.Is(err, io.EOF) {
		return newError(ErrManifestInvalid, "", "trailing data", err)
	}
	return nil
}

// jsonFrame tracks per-container state for validateJSONStructure.
type jsonFrame struct {
	isObj bool
	seen  map[string]struct{}
	key   string // most recently seen object key awaiting its value
}

// validateJSONStructure does a token-based pre-pass that enforces:
//   - depth <= maxManifestDepth
//   - no duplicate keys at any level
//   - no presence of "signatures" key at top level
func validateJSONStructure(raw []byte) error {
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	stack := []jsonFrame{}
	depth := 0
	sawTopLevel := false
	for {
		tok, err := dec.Token()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return newError(ErrManifestInvalid, "", "JSON parse error", err)
		}
		if err := dispatchValidationToken(tok, &stack, &depth, &sawTopLevel); err != nil {
			return err
		}
	}
	return nil
}

// dispatchValidationToken routes a single JSON token to the delim or value
// handler, then resets the pending object key when a value is consumed.
func dispatchValidationToken(tok json.Token, stack *[]jsonFrame, depth *int, sawTopLevel *bool) error {
	if d, ok := tok.(json.Delim); ok {
		return handleDelim(d, stack, depth, sawTopLevel)
	}
	return handleValueToken(tok, *stack, len(*stack) == 1 && !*sawTopLevel)
}

// handleDelim processes object/array open/close delimiters: enforces depth,
// pushes/pops the frame stack, and clears any pending key on the parent
// object when a container value completes.
func handleDelim(d json.Delim, stack *[]jsonFrame, depth *int, sawTopLevel *bool) error {
	if d == '{' || d == '[' {
		return openContainer(d, stack, depth)
	}
	closeContainer(stack, depth, sawTopLevel)
	return nil
}

// openContainer pushes a new frame for an object or array and enforces the
// depth cap.
func openContainer(d json.Delim, stack *[]jsonFrame, depth *int) error {
	*depth++
	if *depth > maxManifestDepth {
		return newError(ErrManifestInvalid, "", "JSON nesting too deep", nil)
	}
	if d == '{' {
		*stack = append(*stack, jsonFrame{isObj: true, seen: map[string]struct{}{}})
	} else {
		*stack = append(*stack, jsonFrame{isObj: false})
	}
	return nil
}

// closeContainer pops the current frame and clears the pending key on the
// parent so the next token there is interpreted as a key (object) or value
// (array).
func closeContainer(stack *[]jsonFrame, depth *int, sawTopLevel *bool) {
	*depth--
	*stack = (*stack)[:len(*stack)-1]
	clearPendingKey(*stack)
	if *depth == 0 {
		*sawTopLevel = true
	}
}

// handleValueToken processes a non-delimiter token. Inside an object, it
// alternates between key (string at expected key position) and value.
func handleValueToken(tok json.Token, stack []jsonFrame, atTopLevelObject bool) error {
	if len(stack) == 0 {
		return nil
	}
	top := &stack[len(stack)-1]
	if !top.isObj {
		return nil
	}
	if top.key == "" {
		s, ok := tok.(string)
		if !ok {
			return nil
		}
		return registerObjectKey(top, s, atTopLevelObject)
	}
	top.key = ""
	return nil
}

// registerObjectKey records a freshly-seen object key, enforcing dedup and
// rejecting the reserved "signatures" key at the top-level object.
func registerObjectKey(top *jsonFrame, key string, atTopLevelObject bool) error {
	if _, dup := top.seen[key]; dup {
		return newError(ErrManifestInvalid, "", fmt.Sprintf("duplicate key %q", key), nil)
	}
	top.seen[key] = struct{}{}
	top.key = key
	if atTopLevelObject && key == "signatures" {
		return newError(ErrManifestInvalid, "", "reserved key 'signatures' present", nil)
	}
	return nil
}

// clearPendingKey clears the pending key on the parent frame so the next
// token there is read as a key (object) or ignored (array).
func clearPendingKey(stack []jsonFrame) {
	if len(stack) == 0 {
		return
	}
	stack[len(stack)-1].key = ""
}

// decodeStrictly decodes the validated JSON into m, with format_version
// parsed via json.RawMessage so we can distinguish a JSON number from a
// JSON string literal (json.Number tolerates both).
func decodeStrictly(dec *json.Decoder, m *Manifest) error {
	type raw struct {
		FormatVersion json.RawMessage `json:"format_version"`
		Entry         string          `json:"entry"`
		Files         []ManifestEntry `json:"files"`
	}
	var r raw
	if err := dec.Decode(&r); err != nil {
		return newError(ErrManifestInvalid, "", "JSON decode error", err)
	}
	v, err := parseFormatVersion(r.FormatVersion)
	if err != nil {
		return err
	}
	m.FormatVersion = v
	m.Entry = r.Entry
	m.Files = r.Files
	return nil
}

// parseFormatVersion converts the raw JSON literal for format_version into a
// positive int value. It rejects string literals, non-canonical literals
// (e.g. "1.0", "1e0"), and out-of-range values.
func parseFormatVersion(rawMsg json.RawMessage) (int, error) {
	if err := assertNumberLiteral(rawMsg); err != nil {
		return 0, err
	}
	n := json.Number(rawMsg)
	v, err := n.Int64()
	if err != nil {
		return 0, newError(ErrManifestInvalid, "format_version", "must be an integer literal", err)
	}
	if err := assertCanonicalIntLiteral(n, v); err != nil {
		return 0, err
	}
	return int(v), nil
}

// assertNumberLiteral verifies the raw JSON literal is non-empty and not a
// string-encoded number.
func assertNumberLiteral(rawMsg json.RawMessage) error {
	if len(rawMsg) == 0 {
		return newError(ErrManifestInvalid, "format_version", "missing", nil)
	}
	if rawMsg[0] == '"' {
		return newError(ErrManifestInvalid, "format_version", "must be a JSON number, not string", nil)
	}
	return nil
}

// assertCanonicalIntLiteral checks that the source literal is in range and
// matches the canonical integer rendering of v (no "1.0", "1e0", "01" etc.).
func assertCanonicalIntLiteral(n json.Number, v int64) error {
	if v < 1 || v > (1<<31-1) {
		return newError(ErrManifestInvalid, "format_version", "out of range", nil)
	}
	if n.String() != fmt.Sprintf("%d", v) {
		return newError(ErrManifestInvalid, "format_version", "non-canonical literal", nil)
	}
	return nil
}

// verifyManifestShape applies the spec's "Schema rules" to a decoded Manifest:
//   - format_version supported
//   - every files[].path is canonical, ends in .dip, in workflows/
//   - every files[].sha256 is lowercase hex, 64 chars
//   - paths are unique (byte-equal AND case-fold-equal)
//   - entry byte-matches exactly one files[].path
func verifyManifestShape(m Manifest) error {
	if err := verifyManifestVersionAndCount(m); err != nil {
		return err
	}
	seenByte, err := verifyFiles(m)
	if err != nil {
		return err
	}
	return verifyEntryInFiles(m, seenByte)
}

// verifyManifestVersionAndCount checks format_version is supported and that
// files[] is non-empty.
func verifyManifestVersionAndCount(m Manifest) error {
	if !isSupportedVersion(m.FormatVersion) {
		return newError(ErrUnsupportedFormatVersion, "", fmt.Sprintf("got %d; supports %v", m.FormatVersion, SupportedFormatVersions()), nil)
	}
	if len(m.Files) == 0 {
		return newError(ErrManifestInvalid, "", "files[] is empty", nil)
	}
	return nil
}

// verifyFiles iterates files[] enforcing path canonicalization, hash format,
// and uniqueness (both byte-equal and case-fold). Returns the byte-seen set
// for the caller to use when validating the entry.
func verifyFiles(m Manifest) (map[string]struct{}, error) {
	seenByte := make(map[string]struct{}, len(m.Files))
	seenFold := make(map[string]struct{}, len(m.Files))
	for _, e := range m.Files {
		if err := verifyOneFile(e, m.FormatVersion, seenByte, seenFold); err != nil {
			return nil, err
		}
	}
	return seenByte, nil
}

// verifyOneFile validates a single ManifestEntry and records it in the seen
// sets. Mutates seenByte and seenFold on success.
func verifyOneFile(e ManifestEntry, version int, seenByte, seenFold map[string]struct{}) error {
	if err := verifyEntryFields(e, version); err != nil {
		return err
	}
	return checkAndRecordPath(e.Path, seenByte, seenFold)
}

// verifyEntryFields enforces presence and shape of an entry's path and sha256
// fields. Missing path is classified as ErrManifestInvalid (schema rule 3),
// not as ErrPathUnsafe. Path safety is checked by Canonicalize; the
// version-aware prefix/suffix POLICY is checked by checkPathPolicy.
func verifyEntryFields(e ManifestEntry, version int) error {
	if e.Path == "" {
		return newError(ErrManifestInvalid, "", "files[] entry missing required key: path", nil)
	}
	if _, err := Canonicalize(e.Path); err != nil {
		return err
	}
	if err := checkPathPolicy(e.Path, version); err != nil {
		return err
	}
	if !isValidHash(e.SHA256) {
		return newError(ErrManifestInvalid, e.Path, "sha256 not 64-char lowercase hex", nil)
	}
	return nil
}

// checkAndRecordPath enforces byte- and case-fold-uniqueness for path against
// the running seen sets, recording the path on success.
func checkAndRecordPath(path string, seenByte, seenFold map[string]struct{}) error {
	if _, dup := seenByte[path]; dup {
		return newError(ErrManifestInvalid, path, "duplicate path in files[]", nil)
	}
	fold := strings.ToLower(path)
	if _, dup := seenFold[fold]; dup {
		return newError(ErrManifestInvalid, path, "case-fold-duplicate path in files[]", nil)
	}
	seenByte[path] = struct{}{}
	seenFold[fold] = struct{}{}
	return nil
}

// verifyEntryInFiles canonicalizes m.Entry and confirms it byte-matches one of
// the previously-seen files[].path values.
func verifyEntryInFiles(m Manifest, seenByte map[string]struct{}) error {
	if _, err := Canonicalize(m.Entry); err != nil {
		return err
	}
	// entry is always a workflow; v1 policy (workflows/ prefix + .dip suffix)
	// applies regardless of bundle format version.
	if err := checkPathPolicy(m.Entry, 1); err != nil {
		return err
	}
	if _, ok := seenByte[m.Entry]; !ok {
		return newError(ErrEntryNotInManifest, m.Entry, "", nil)
	}
	return nil
}

func isSupportedVersion(v int) bool {
	for _, sv := range SupportedFormatVersions() {
		if sv == v {
			return true
		}
	}
	return false
}

func isValidHash(s string) bool {
	if len(s) != 64 {
		return false
	}
	for _, r := range s {
		if !isLowerHexRune(r) {
			return false
		}
	}
	return true
}

// isLowerHexRune reports whether r is a lowercase hex digit ([0-9a-f]).
func isLowerHexRune(r rune) bool {
	if r >= '0' && r <= '9' {
		return true
	}
	return r >= 'a' && r <= 'f'
}

// SupportedFormatVersions returns the format_version values this build accepts.
// Returns a fresh slice on every call to prevent mutation by callers.
func SupportedFormatVersions() []int { return []int{1, 2} }
