package validator

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/2389-research/dippin-lang/ir"
)

// lintWritablePaths fires DIP141 (writable_paths nullified by tool_access: none)
// and DIP142 (unsafe writable_paths entry) on agent nodes and per-branch overrides.
// dippin carries + lints the field; the runtime enforces the fs-level write jail.
func lintWritablePaths(w *ir.Workflow) []Diagnostic {
	var diags []Diagnostic
	for _, n := range w.Nodes {
		diags = append(diags, checkNodeWritablePathsByKind(n)...)
	}
	return diags
}

func checkNodeWritablePathsByKind(n *ir.Node) []Diagnostic {
	switch cfg := n.Config.(type) {
	case ir.AgentConfig:
		return checkWritablePathsObject(n, cfg.WritablePaths, cfg.ToolAccess, "")
	case ir.ParallelConfig:
		return checkBranchWritablePaths(n, cfg.Branches)
	default:
		return nil
	}
}

func checkBranchWritablePaths(n *ir.Node, branches []ir.BranchConfig) []Diagnostic {
	var diags []Diagnostic
	for _, b := range branches {
		diags = append(diags, checkWritablePathsObject(n, b.WritablePaths, b.ToolAccess, b.Target)...)
	}
	return diags
}

func checkWritablePathsObject(n *ir.Node, paths []string, toolAccess, branch string) []Diagnostic {
	if len(paths) == 0 {
		return nil
	}
	var diags []Diagnostic
	if strings.ToLower(strings.TrimSpace(toolAccess)) == "none" {
		diags = append(diags, dip141Diagnostic(n, branch))
	}
	for _, entry := range paths {
		if kind := unsafeEntryKind(entry); kind != "" {
			diags = append(diags, dip142Diagnostic(n, branch, entry, kind))
		}
	}
	return diags
}

func dip141Diagnostic(n *ir.Node, branch string) Diagnostic {
	msg := fmt.Sprintf("node %q has writable_paths but tool_access \"none\" — none strips all tools, so there is nothing to bound (dead config)", n.ID)
	if branch != "" {
		msg = fmt.Sprintf("node %q branch %q has writable_paths but tool_access \"none\" — none strips all tools, so there is nothing to bound (dead config)", n.ID, branch)
	}
	return Diagnostic{
		Code:     DIP141,
		Severity: SeverityWarning,
		Message:  msg,
		Location: n.Source,
		Help:     "remove writable_paths (no tools to bound) or drop tool_access: none to grant a bounded tool catalog.",
	}
}

// unsafeEntryKind classifies a writable_paths entry that will not bound writes as
// the author expects. Returns "" for a safe relative glob. This is a lexical
// clarity check, not the security boundary — the runtime fs-jail is authoritative.
func unsafeEntryKind(entry string) string {
	e := strings.TrimSpace(entry)
	if e == "" {
		return ""
	}
	if isAbsoluteEntry(e) {
		return "absolute"
	}
	if isParentEscape(e) {
		return "escape"
	}
	if strings.Count(e, "{") != strings.Count(e, "}") {
		return "brace"
	}
	return ""
}

func isAbsoluteEntry(e string) bool {
	return strings.HasPrefix(e, "/") || strings.HasPrefix(e, "~") ||
		strings.HasPrefix(e, "\\") || hasWindowsDrive(e)
}

func hasWindowsDrive(e string) bool {
	return len(e) >= 2 && e[1] == ':' && isDriveLetter(e[0])
}

func isDriveLetter(c byte) bool {
	return (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z')
}

// isParentEscape reports whether the cleaned entry contains a `..` segment that
// would escape its base. Local copy of parser.hasParentRef (validator imports ir only).
// Backslashes are normalised to forward slashes before cleaning so that Windows
// escape patterns like `..\secrets\**` are caught on all host platforms.
func isParentEscape(e string) bool {
	e = strings.ReplaceAll(e, "\\", "/")
	cleaned := filepath.ToSlash(filepath.Clean(e))
	for _, seg := range strings.Split(cleaned, "/") {
		if seg == ".." {
			return true
		}
	}
	return false
}

func dip142Diagnostic(n *ir.Node, branch, entry, kind string) Diagnostic {
	reason := writablePathReason(kind)
	msg := fmt.Sprintf("node %q writable_paths entry %q %s", n.ID, entry, reason)
	if branch != "" {
		msg = fmt.Sprintf("node %q branch %q writable_paths entry %q %s", n.ID, branch, entry, reason)
	}
	return Diagnostic{
		Code:     DIP142,
		Severity: SeverityWarning,
		Message:  msg,
		Location: n.Source,
		Help:     "use workspace-relative globs (e.g. .ai/sprints/**). Absolute, ~, and ..-escaping entries are rejected by the fs jail (it bounds writes to the session root); this lint catches obvious lexical cases only — the runtime jail is the real boundary. See #67/#77.",
	}
}

func writablePathReason(kind string) string {
	if kind == "brace" {
		return "is malformed (brace expansion is split on commas — enumerate entries instead)"
	}
	return "escapes the workspace (absolute / ~ / parent path) — the runtime write-jail will not honor it"
}
