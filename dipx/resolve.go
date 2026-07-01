package dipx

import (
	"fmt"
	"path"
	"strings"
	"unicode/utf8"

	"golang.org/x/text/unicode/norm"
)

const (
	maxPathBytes      = 1024
	maxPathComponents = 16
)

// Canonicalize applies the version-independent SAFETY rules (NFC, reject
// absolute paths, .. segments, backslash, NUL/control chars, component-count
// cap, Windows-reserved names) and returns the canonical form of a
// bundle-relative path or an error if any safety rule is violated. It does
// NOT enforce the workflows/ prefix or the .dip suffix — those are
// version-aware POLICY, enforced separately inside dipx by checkPathPolicy;
// external callers of Canonicalize get safety canonicalization only. Within
// dipx, all bundle-path handling MUST go through this function; no other code
// in dipx is permitted to call path.Clean / filepath.Clean.
func Canonicalize(p string) (string, error) {
	checks := []func(string) error{
		checkPathBasics,
		checkPathStructure,
		checkPathComponents,
		checkPathIdempotent,
	}
	for _, fn := range checks {
		if err := fn(p); err != nil {
			return "", err
		}
	}
	return p, nil
}

// checkPathIdempotent verifies path.Clean leaves the input unchanged.
func checkPathIdempotent(p string) error {
	if cleaned := path.Clean(p); cleaned != p {
		return newError(ErrPathUnsafe, p, "not canonical", nil)
	}
	return nil
}

// checkPathBasics handles empty, byte-level, NFC, and length checks.
func checkPathBasics(p string) error {
	if p == "" {
		return newError(ErrPathUnsafe, p, "empty path", nil)
	}
	if err := checkBytes(p); err != nil {
		return err
	}
	if normed := norm.NFC.String(p); normed != p {
		return newError(ErrPathUnsafe, p, "not in NFC form", nil)
	}
	if len(p) > maxPathBytes {
		return newError(ErrPathUnsafe, p, "path exceeds 1024 bytes", nil)
	}
	return nil
}

// checkPathStructure handles absolute / leading-./ / repeated-slash / dot-dot.
func checkPathStructure(p string) error {
	if strings.HasPrefix(p, "/") {
		return newError(ErrPathUnsafe, p, "absolute path", nil)
	}
	if strings.HasPrefix(p, "./") {
		return newError(ErrPathUnsafe, p, "leading ./", nil)
	}
	if strings.Contains(p, "//") {
		return newError(ErrPathUnsafe, p, "empty path component", nil)
	}
	if hasDotDotSegment(p) {
		return newError(ErrPathUnsafe, p, "contains .. segment", nil)
	}
	return nil
}

// checkPathComponents enforces the component-count cap and per-component rules.
func checkPathComponents(p string) error {
	parts := strings.Split(p, "/")
	if len(parts) > maxPathComponents {
		return newError(ErrPathUnsafe, p, "too many path components", nil)
	}
	for _, c := range parts {
		if err := checkComponent(p, c); err != nil {
			return err
		}
	}
	return nil
}

// isReservedBundleName reports whether p collides with a root-level reserved
// bundle entry. Because assets live under workflows/ they can never collide,
// but checkPathPolicy asserts it anyway as defense in depth (spec § Security 4).
func isReservedBundleName(p string) bool {
	return p == "manifest.json" || p == "manifest.sig"
}

// checkPathPolicy enforces the version-aware path POLICY that Canonicalize no
// longer applies. Both versions require the workflows/ prefix and reject the
// reserved bundle names; format_version 1 additionally requires the .dip
// suffix (every listed path is a workflow). format_version 2 allows any
// canonical path under workflows/ that does not end in .dip to be an asset.
func checkPathPolicy(p string, version int) error {
	if isReservedBundleName(p) {
		return newError(ErrPathUnsafe, p, "reserved bundle name", nil)
	}
	if !strings.HasPrefix(p, "workflows/") {
		return newError(ErrPathUnsafe, p, "must start with workflows/", nil)
	}
	if version == 1 && !strings.HasSuffix(p, ".dip") {
		return newError(ErrPathUnsafe, p, "must end with .dip", nil)
	}
	return nil
}

func checkBytes(p string) error {
	if !utf8.ValidString(p) {
		return newError(ErrPathUnsafe, p, "invalid UTF-8", nil)
	}
	for _, r := range p {
		if err := checkRune(r, p); err != nil {
			return err
		}
	}
	return nil
}

func checkRune(r rune, p string) error {
	if r == '\\' {
		return newError(ErrPathUnsafe, p, "backslash separator", nil)
	}
	if r == 0 {
		return newError(ErrPathUnsafe, p, "NUL byte", nil)
	}
	if r < 0x20 || r == 0x7f {
		return newError(ErrPathUnsafe, p, "control character", nil)
	}
	return nil
}

func hasDotDotSegment(p string) bool {
	for _, c := range strings.Split(p, "/") {
		if c == ".." {
			return true
		}
	}
	return false
}

func checkComponent(p, c string) error {
	if c == "" {
		return newError(ErrPathUnsafe, p, "empty component", nil)
	}
	if err := checkComponentWhitespaceAndDots(p, c); err != nil {
		return err
	}
	if isWindowsReserved(c) {
		return newError(ErrPathUnsafe, p, fmt.Sprintf("Windows reserved name: %q", c), nil)
	}
	return nil
}

func checkComponentWhitespaceAndDots(p, c string) error {
	if strings.HasPrefix(c, " ") || strings.HasSuffix(c, " ") {
		return newError(ErrPathUnsafe, p, fmt.Sprintf("leading/trailing whitespace in component %q", c), nil)
	}
	if strings.HasSuffix(stripExt(c), " ") {
		return newError(ErrPathUnsafe, p, fmt.Sprintf("trailing whitespace before extension in component %q", c), nil)
	}
	if strings.HasSuffix(c, ".") {
		return newError(ErrPathUnsafe, p, fmt.Sprintf("trailing dot in component %q", c), nil)
	}
	return nil
}

func isWindowsReserved(c string) bool {
	upper := strings.ToUpper(stripExt(c))
	switch upper {
	case "CON", "PRN", "AUX", "NUL":
		return true
	}
	return isWindowsNumberedReserved(upper)
}

func isWindowsNumberedReserved(upper string) bool {
	if len(upper) != 4 {
		return false
	}
	if !strings.HasPrefix(upper, "COM") && !strings.HasPrefix(upper, "LPT") {
		return false
	}
	r := upper[3]
	return r >= '0' && r <= '9'
}

// stripExt returns the component prefix before the FIRST dot. Using the first
// dot (not the last) is required so multi-extension forms like "CON.tar.dip"
// are still classified as Windows-reserved — on Windows, "CON.anything" maps
// to the CON device regardless of how many extensions follow.
func stripExt(c string) string {
	if i := strings.IndexByte(c, '.'); i >= 0 {
		return c[:i]
	}
	return c
}

// Tri-color DFS marker values for detectCycles.
const (
	colorWhite = 0
	colorGray  = 1
	colorBlack = 2
)

// maxRefDepth bounds the ref-graph DFS depth. A de-facto constant: every
// caller used the literal 64.
const maxRefDepth = 64

// detectCycles runs a tri-color DFS over the ref graph rooted at start.
// Returns ErrRefCycle on the first cycle found, ErrCapExceeded when depth
// exceeds maxRefDepth.
func detectCycles(graph map[string][]string, start string) error {
	color := make(map[string]int, len(graph))
	stack := make([]string, 0, 16)
	return dfsVisit(graph, color, &stack, start, 0)
}

// dfsVisit is the recursive worker for detectCycles. Hoisted to a top-level
// helper so detectCycles stays under the project's complexity caps. The
// stack tracks the active DFS path so cycle errors can include the full
// path from the cycle entry node back to itself.
func dfsVisit(graph map[string][]string, color map[string]int, stack *[]string, node string, depth int) error {
	if depth > maxRefDepth {
		return newError(ErrCapExceeded, node, fmt.Sprintf("ref-graph depth exceeds %d", maxRefDepth), nil)
	}
	color[node] = colorGray
	*stack = append(*stack, node)
	for _, next := range graph[node] {
		if err := dfsVisitEdge(graph, color, stack, next, depth); err != nil {
			return err
		}
	}
	*stack = (*stack)[:len(*stack)-1]
	color[node] = colorBlack
	return nil
}

// dfsVisitEdge inspects a single outgoing edge to next, recursing when next
// is unvisited and reporting a cycle when next is on the active path.
// The reported error's Path field is the cycle entry node (next, where the
// back-edge points), and Detail is the full cycle path "n1 -> n2 -> ... -> n1".
func dfsVisitEdge(graph map[string][]string, color map[string]int, stack *[]string, next string, depth int) error {
	switch color[next] {
	case colorGray:
		return newError(ErrRefCycle, next, formatCycle(*stack, next), nil)
	case colorWhite:
		return dfsVisit(graph, color, stack, next, depth+1)
	}
	return nil
}

// formatCycle renders the active DFS stack as "n1 -> n2 -> ... -> nk -> n1"
// where n1 is the cycle entry node (where the back-edge points). The caller
// (dfsVisitEdge) only invokes this when target is colorGray, which the tri-color
// DFS sets exactly while target is on the active stack — so target is always
// found. A miss would be an invariant violation, hence the panic.
func formatCycle(stack []string, target string) string {
	idx := -1
	for i, n := range stack {
		if n == target {
			idx = i
			break
		}
	}
	if idx < 0 {
		panic(fmt.Sprintf("formatCycle: target %q not on active DFS stack %v", target, stack))
	}
	cycle := append([]string{}, stack[idx:]...)
	cycle = append(cycle, target)
	return strings.Join(cycle, " -> ")
}

// resolveLexically computes the resolved bundle-relative path of a ref string
// relative to a parent workflow's bundle path. The resolved path is then
// validated by Canonicalize.
//
// SPEC NOTE: This function uses path.Clean and path.Join for lexical-join
// (resolving '..' and '/.' segments before validation). The spec mandates
// that "all four sites (Pack, Open, Source.Workflow, Extract) call exactly
// one Canonicalize function." resolveLexically is part of the dipx package's
// internal canonicalization pipeline; its path.Clean usage is for input
// preparation, and the function ALWAYS calls Canonicalize on the result
// before returning. The future CI grep added in Task 26 must allowlist this
// helper.
//
// refPath comes from a workflow's source (subgraph ref:); relativeTo is the
// bundle-relative path of the parent workflow.
func resolveLexically(refPath, relativeTo string) (string, error) {
	if refPath == "" {
		return "", newError(ErrPathUnsafe, refPath, "empty ref", nil)
	}
	dir := path.Dir(relativeTo)
	if dir == "." {
		dir = ""
	}
	joined := path.Join(dir, refPath)
	cleaned := path.Clean(joined)
	// Run through Canonicalize for safety checks, then apply v1 path policy:
	// subgraph refs are always workflows (workflows/ prefix + .dip suffix),
	// which holds for both format_version 1 and 2 bundles. Note: refPath may
	// have originally contained "..", which path.Clean resolves; the resulting
	// cleaned path must itself be canonical. The future CI grep (Task 26) must
	// allowlist both the Canonicalize and checkPathPolicy calls here.
	canon, err := Canonicalize(cleaned)
	if err != nil {
		return "", err
	}
	return canon, checkPathPolicy(canon, 1)
}
