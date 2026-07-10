package ir

import "strings"

// ComposePrompt assembles the effective prompt from already-resolved parts, in
// order prefix → body → include → suffix, joining only the non-empty parts with
// a blank line. The suffix is always last, satisfying downstream "the very last
// line must be exactly …" control-protocol contracts (#175).
//
// When any fragment is present the result is trailing-whitespace-normalized the
// same way the formatter normalizes a multiline block (per-line right-trim, then
// overall right-trim). This keeps the composed prompt byte-identical across a
// source run, an inline-packed bundle (whose prompt round-trips through the
// formatter), and a no-inline bundle. A plain body with no fragments is returned
// untouched, so existing prompt / prompt_file behavior is unchanged.
func ComposePrompt(prefix, body, include, suffix string) string {
	if hasNoFragments(prefix, include, suffix) {
		return body
	}
	return normalizeComposedPrompt(joinNonEmpty(prefix, body, include, suffix))
}

func hasNoFragments(prefix, include, suffix string) bool {
	return prefix == "" && include == "" && suffix == ""
}

func joinNonEmpty(parts ...string) string {
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p != "" {
			out = append(out, p)
		}
	}
	return strings.Join(out, "\n\n")
}

// normalizeComposedPrompt mirrors the formatter's multiline-block trimming so a
// fragment's trailing whitespace or newline does not survive composition (which
// would otherwise diverge from the formatter-round-tripped inline-pack form).
func normalizeComposedPrompt(s string) string {
	lines := strings.Split(s, "\n")
	for i, l := range lines {
		lines[i] = strings.TrimRight(l, " \t\r")
	}
	return strings.TrimRight(strings.Join(lines, "\n"), " \t\n\r")
}
