package ir

import "strings"

// ComposePrompt assembles the effective prompt from already-resolved parts, in
// order prefix → body → include → suffix, joining only the non-empty parts with
// a blank line. The suffix is always last, satisfying downstream "the very last
// line must be exactly …" control-protocol contracts (#175).
func ComposePrompt(prefix, body, include, suffix string) string {
	parts := make([]string, 0, 4)
	for _, p := range []string{prefix, body, include, suffix} {
		if p != "" {
			parts = append(parts, p)
		}
	}
	return strings.Join(parts, "\n\n")
}
