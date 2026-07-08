// ABOUTME: DIP152 — a tool node's marker_grep enumerates a literal marker that
// ABOUTME: no edge routes and no else/unconditional edge covers.
package validator

import (
	"strings"
)

// markerMetachars are the regex metacharacters that make a branch non-literal.
const markerMetachars = ".*+?[]{}()|\\^$"

// enumerateMarkers returns the literal marker set of a marker_grep value when it
// is a recognizable finite literal alternation (optional ^...$ anchors around a
// single literal token or a full-span (a|b|c) group of literal branches).
// Returns (nil, false) for anything else — those keep the blanket DIP101/DIP102
// exemption and get no DIP152.
func enumerateMarkers(markerGrep string) ([]string, bool) {
	branches, ok := splitAlternation(stripAnchors(markerGrep))
	if !ok {
		return nil, false
	}
	return collectUniqueLiterals(branches)
}

// collectUniqueLiterals converts a slice of regex branches into a deduplicated
// literal marker list, returning (nil, false) if any branch is empty or contains
// a metacharacter.
func collectUniqueLiterals(branches []string) ([]string, bool) {
	seen := make(map[string]struct{})
	var markers []string
	for _, b := range branches {
		if b == "" || !isLiteralToken(b) {
			return nil, false
		}
		if _, dup := seen[b]; !dup {
			seen[b] = struct{}{}
			markers = append(markers, b)
		}
	}
	return markers, true
}

// stripAnchors removes one leading ^ and one trailing $ if present.
func stripAnchors(s string) string {
	return strings.TrimSuffix(strings.TrimPrefix(s, "^"), "$")
}

// splitAlternation returns the branches of a full-span (a|b|c) group, or the
// whole string as a single branch when it is not a full-span group. Empty input
// is non-enumerable.
func splitAlternation(s string) ([]string, bool) {
	if s == "" {
		return nil, false
	}
	if strings.HasPrefix(s, "(") && strings.HasSuffix(s, ")") {
		return strings.Split(s[1:len(s)-1], "|"), true
	}
	return []string{s}, true
}

// isLiteralToken reports whether s contains no regex metacharacter.
func isLiteralToken(s string) bool {
	return !strings.ContainsAny(s, markerMetachars)
}
