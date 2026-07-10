package parser

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/2389-research/dippin-lang/ir"
)

const maxDirectiveFileSize = 4 << 20 // 4 MiB

// ResolveFileDirectives loads file contents for every *_file directive on a
// node: a tool node's CommandFile into Command, and an agent node's PromptFile
// into Prompt and SystemPromptFile into SystemPrompt. Paths are
// resolved relative to baseDir (typically the directory of the .dip
// source file). Returns the first error encountered.
//
// Parser entry points (NewParser/Parse) do NOT call this — they stay pure.
// CLI entry points call it after parsing. LSP and WASM contexts skip it;
// the IR retains the CommandFile-set / Command-empty state, which is the
// correct unresolved view for those consumers.
func ResolveFileDirectives(w *ir.Workflow, baseDir string) error {
	cascade, err := loadPromptCascade(&w.Defaults, baseDir)
	if err != nil {
		return err
	}
	for _, n := range w.Nodes {
		if err := resolveNodeDirective(n, baseDir, cascade); err != nil {
			return err
		}
	}
	return nil
}

// promptCascade holds the resolved defaults prompt prefix/suffix (inline value or
// loaded file content), applied to every agent unless a node opts out (#175).
type promptCascade struct {
	prefix, suffix string
}

// loadPromptCascade resolves the defaults-block prompt cascade once per workflow
// so N agents do not re-read the fragment files N times.
func loadPromptCascade(d *ir.WorkflowDefaults, baseDir string) (promptCascade, error) {
	c := promptCascade{prefix: d.PromptPrefix, suffix: d.PromptSuffix}
	if err := loadDirectiveInto(&c.prefix, d.PromptPrefixFile, baseDir, "defaults", "prompt_prefix_file"); err != nil {
		return c, err
	}
	if err := loadDirectiveInto(&c.suffix, d.PromptSuffixFile, baseDir, "defaults", "prompt_suffix_file"); err != nil {
		return c, err
	}
	return c, nil
}

// resolveNodeDirective resolves any file-directive fields on a single node.
// Dispatches per node-config kind so the per-kind loader functions stay
// focused on their own field set.
func resolveNodeDirective(n *ir.Node, baseDir string, cascade promptCascade) error {
	switch cfg := n.Config.(type) {
	case ir.ToolConfig:
		return resolveToolDirective(n, cfg, baseDir)
	case ir.AgentConfig:
		return resolveAgentDirective(n, cfg, baseDir, cascade)
	}
	return nil
}

// resolveToolDirective populates ToolConfig.Command from CommandFile, if set.
func resolveToolDirective(n *ir.Node, cfg ir.ToolConfig, baseDir string) error {
	if err := loadDirectiveInto(&cfg.Command, cfg.CommandFile, baseDir, n.ID, "command_file"); err != nil {
		return err
	}
	n.Config = cfg
	return nil
}

// resolveAgentDirective populates Prompt and SystemPrompt from their *File
// twins on AgentConfig. The two slots are independent — either, both, or
// neither may be set.
func resolveAgentDirective(n *ir.Node, cfg ir.AgentConfig, baseDir string, cascade promptCascade) error {
	if err := loadDirectiveInto(&cfg.Prompt, cfg.PromptFile, baseDir, n.ID, "prompt_file"); err != nil {
		return err
	}
	if err := loadDirectiveInto(&cfg.SystemPrompt, cfg.SystemPromptFile, baseDir, n.ID, "system_prompt_file"); err != nil {
		return err
	}
	include := ""
	if err := loadDirectiveInto(&include, cfg.PromptInclude, baseDir, n.ID, "prompt_include"); err != nil {
		return err
	}
	cfg.Prompt = ir.ComposePrompt(effectivePrefix(cfg, cascade), cfg.Prompt, include, effectiveSuffix(cfg, cascade))
	n.Config = cfg
	return nil
}

// effectivePrefix returns the cascade prefix unless the node opted out with
// `prompt_prefix: none` (#175).
func effectivePrefix(cfg ir.AgentConfig, cascade promptCascade) string {
	if cfg.PromptPrefix == "none" {
		return ""
	}
	return cascade.prefix
}

// effectiveSuffix returns the cascade suffix unless the node opted out with
// `prompt_suffix: none` (#175).
func effectiveSuffix(cfg ir.AgentConfig, cascade promptCascade) string {
	if cfg.PromptSuffix == "none" {
		return ""
	}
	return cascade.suffix
}

// loadDirectiveInto reads path (relative to baseDir) into *dst, no-op if path
// is empty. When path is non-empty, *dst is always empty: the parser rejects a
// node declaring both an inline value and its *_file directive, and the CLI
// bails on that parse error before reaching the resolver.
func loadDirectiveInto(dst *string, path, baseDir, nodeID, directive string) error {
	if path == "" {
		return nil
	}
	contents, err := loadDirectiveFile(baseDir, path)
	if err != nil {
		return fmt.Errorf("node %q %s: %w", nodeID, directive, err)
	}
	*dst = string(contents)
	return nil
}

// loadDirectiveFile resolves p relative to baseDir, applies path security
// checks, and reads the file. Error messages reference the user-written
// path (p), never the resolved absolute path, to avoid leaking directory
// structure into diagnostics. In particular, *fs.PathError values from the
// os package carry the resolved absolute path in their stringification, so we
// never wrap them with %w — instead we branch on the error kind and emit a
// user-path-only message.
func loadDirectiveFile(baseDir, p string) ([]byte, error) {
	resolved, err := safeResolve(baseDir, p)
	if err != nil {
		return nil, err
	}
	if err := checkContainment(baseDir, resolved, p); err != nil {
		return nil, err
	}
	return openCheckRead(p, resolved)
}

// openCheckRead opens resolved exactly once and validates + reads the file
// through that single fd. Opening, fstat, and read all operate on the same
// descriptor, so nothing is re-resolved by pathname between the symlink/size
// checks and the read — closing the leaf check-to-read TOCTOU race that a
// separate Lstat+ReadFile pair left open (#79).
//
// O_NOFOLLOW (oNoFollow on Unix) makes leaf-symlink rejection atomic: open()
// fails with ELOOP when the final path component is a symlink. It affects only
// the final component, so contained parent symlinks stay followed and remain
// validated by checkContainment. On non-unix targets (oNoFollow == 0, see
// resolve_nofollow_other.go): the fd-based fstat→read still closes the
// fstat-to-read race there, but atomic leaf-symlink rejection is unix-only.
//
// Residual (out of scope for #79): checkContainment validates the parent chain
// at open-adjacent time, but a fully race-free parent walk needs
// openat-per-component, a much larger cross-platform lift. This closes the LEAF
// check-to-read race only.
func openCheckRead(p, resolved string) ([]byte, error) {
	f, err := os.OpenFile(resolved, os.O_RDONLY|oNoFollow, 0)
	if err != nil {
		return nil, openErr(p, err)
	}
	defer func() { _ = f.Close() }()
	info, err := f.Stat()
	if err != nil {
		return nil, pathErr(p, err, "stat")
	}
	if err := checkFileInfo(p, info); err != nil {
		return nil, err
	}
	return readFromFD(p, f)
}

// readFromFD reads the already-open fd, rewriting any error so it only mentions
// the user-written path p (never the absolute resolved one). The read is bounded
// by io.LimitReader as a belt to checkFileInfo's fstat size check: under
// untrusted-.dip-in-CI a concurrent writer could grow the file on the same fd
// between fstat and read, so capping the read at maxDirectiveFileSize+1 keeps
// memory bounded regardless of post-fstat growth.
func readFromFD(p string, f *os.File) ([]byte, error) {
	contents, err := io.ReadAll(io.LimitReader(f, maxDirectiveFileSize+1))
	if err != nil {
		return nil, pathErr(p, err, "read")
	}
	if int64(len(contents)) > maxDirectiveFileSize {
		return nil, fmt.Errorf("file %q is too large (exceeds max %s)", p, formatMiB(maxDirectiveFileSize))
	}
	return contents, nil
}

// openErr maps an os.OpenFile error to a user-path-only diagnostic. An
// O_NOFOLLOW open of a symlink leaf fails with ELOOP inside an *fs.PathError
// whose stringification embeds the resolved absolute path; we detect it via
// errors.Is and emit the same user-path-only "symlinks not allowed" message the
// old Lstat-based leaf check produced. Other errors defer to pathErr, which is
// also user-path-only.
//
// This check needs no build tag (unlike oNoFollow): syscall.ELOOP is a portable
// errno defined on every target, including js/wasm and windows, whereas
// syscall.O_NOFOLLOW exists only on unix. On a non-unix build oNoFollow is 0, so
// open never returns ELOOP and this branch is simply dead — never reached, but
// it compiles everywhere.
func openErr(p string, err error) error {
	if errors.Is(err, syscall.ELOOP) {
		return fmt.Errorf("symlinks not allowed: %q", p)
	}
	return pathErr(p, err, "open")
}

// pathErr maps a filesystem error to a user-path-only diagnostic. *fs.PathError
// values from the os package carry the resolved absolute path in their
// stringification, so we branch on the error kind and emit a message that names
// only the user-written path p (verb is the failed operation, e.g. "stat").
func pathErr(p string, err error, verb string) error {
	if errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("file %q not found", p)
	}
	return fmt.Errorf("cannot %s file %q: not accessible", verb, p)
}

// safeResolve joins baseDir/p and ensures the result stays under baseDir.
// Rejects absolute paths and any path that escapes via `..`. This is a purely
// lexical check; symlink-based escapes are caught separately by
// checkContainment.
func safeResolve(baseDir, p string) (string, error) {
	if filepath.IsAbs(p) {
		return "", fmt.Errorf("absolute paths not allowed: %q", p)
	}
	resolved := filepath.Join(baseDir, p)
	if escapesBase(baseDir, resolved) {
		return "", fmt.Errorf("path %q resolves outside source directory", p)
	}
	return resolved, nil
}

// checkContainment resolves the symlink chain of the file's parent directory
// and rejects the path if its real location escapes baseDir. safeResolve only
// does lexical containment and checkFileInfo's Lstat only inspects the leaf, so
// a symlinked *parent* directory (e.g. baseDir/sub -> /etc, then sub/passwd)
// defeats both. Resolving the parent — not the leaf — leaves the separate
// leaf-symlink rejection in checkFileInfo intact. (#67)
func checkContainment(baseDir, resolved, p string) error {
	realBase, err := realPath(baseDir)
	if err != nil {
		return pathErr(p, err, "stat")
	}
	realParent, err := realPath(filepath.Dir(resolved))
	if err != nil {
		return pathErr(p, err, "stat")
	}
	if escapesBase(realBase, realParent) {
		return fmt.Errorf("path %q resolves outside source directory", p)
	}
	return nil
}

// realPath resolves path's symlink chain and returns it absolute and cleaned.
// Both the base and the parent must be absolute before escapesBase compares
// them: with a relative baseDir, a symlink whose target is absolute yields a
// relative base but an absolute parent, and filepath.Rel errors on that pair —
// which would otherwise be misread as an escape and reject a valid in-dir file
// (e.g. `dippin validate workflow.dip`, baseDir "."). (#67)
func realPath(path string) (string, error) {
	resolved, err := filepath.EvalSymlinks(path)
	if err != nil {
		return "", err
	}
	return filepath.Abs(resolved)
}

// escapesBase reports whether target lies outside base via a `..` ancestor
// reference or an unrelated root. Shared by the lexical (safeResolve) and
// full-chain (checkContainment) containment checks.
func escapesBase(base, target string) bool {
	rel, err := filepath.Rel(base, target)
	return err != nil || hasParentRef(rel) || filepath.IsAbs(rel)
}

// checkFileInfo enforces symlink and size policies on the fd's fstat info. The
// size cap is authoritative. The symlink-mode check is defensive: on unix the
// leaf symlink was already rejected atomically by O_NOFOLLOW at open, so fstat
// never sees one; on non-unix targets (oNoFollow == 0) open follows the leaf,
// so fstat sees the target and this check cannot fire.
func checkFileInfo(p string, info os.FileInfo) error {
	if info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("symlinks not allowed: %q", p)
	}
	if info.Size() > maxDirectiveFileSize {
		return fmt.Errorf("file %q is too large (size %s, max %s)",
			p, formatMiB(info.Size()), formatMiB(maxDirectiveFileSize))
	}
	return nil
}

// formatMiB renders a byte count as a human-readable MiB string. The limit
// (4 MiB) is a whole-MiB value, so an integer cap renders cleanly; user-file
// sizes get one decimal place to distinguish, e.g., 5.0 MiB from 4.9 MiB.
func formatMiB(n int64) string {
	const mib = 1 << 20
	if n%mib == 0 {
		return fmt.Sprintf("%d MiB", n/mib)
	}
	return fmt.Sprintf("%.1f MiB", float64(n)/float64(mib))
}

// hasParentRef returns true if rel contains a `..` path segment.
// Used as defensive belt after filepath.Rel; should not fire on a
// non-symlink filesystem since Rel canonicalizes already.
func hasParentRef(rel string) bool {
	for _, seg := range strings.Split(filepath.ToSlash(rel), "/") {
		if seg == ".." {
			return true
		}
	}
	return false
}
