package dipx

import (
	"archive/zip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/2389-research/dippin-lang/ir"
	"github.com/2389-research/dippin-lang/parser"
	"github.com/2389-research/dippin-lang/simulate"
)

const (
	maxFiles            = 10000
	maxTotalUncompBytes = 100 << 20 // 100 MB
	maxPerFileBytes     = 50 << 20  // 50 MB
)

// readManifestEntry locates manifest.json in the constrained zip and reads
// up to maxManifestSize+1 bytes, rejecting oversized inputs before any further
// processing.
func readManifestEntry(cz *constrainedZip) ([]byte, error) {
	f, ok := cz.entries["manifest.json"]
	if !ok {
		return nil, newError(ErrManifestMissing, "", "manifest.json not at zip root", nil)
	}
	rc, err := f.Open()
	if err != nil {
		return nil, newError(ErrManifestInvalid, "manifest.json", "open failed", err)
	}
	defer rc.Close()
	limited := &io.LimitedReader{R: rc, N: int64(maxManifestSize) + 1}
	raw, err := io.ReadAll(limited)
	if err != nil {
		return nil, newError(ErrManifestInvalid, "manifest.json", "read failed", err)
	}
	if int64(len(raw)) > int64(maxManifestSize) {
		return nil, newError(ErrManifestInvalid, "manifest.json", "exceeds 1MB", nil)
	}
	return raw, nil
}

// verifyAllHashesCtx streams each file in m.Files through SHA-256 verification,
// enforcing per-file and total-uncompressed caps as streaming bounds during
// decompression. Returns the verified bytes (keyed by canonical bundle path)
// and the running total. The effective per-file cap is min(maxPerFileBytes,
// totalCap-total), so the running total cap is enforced as a streaming bound
// rather than after a per-file allocation has already happened. Checks ctx at
// entry and before each entry in the loop (P10.10).
func verifyAllHashesCtx(ctx context.Context, cz *constrainedZip, m Manifest, totalCap int64) (map[string]verifiedBytes, int64, error) {
	if err := ctx.Err(); err != nil {
		return nil, 0, err
	}
	if len(m.Files) > maxFiles {
		return nil, 0, newError(ErrCapExceeded, "", fmt.Sprintf("files exceeds %d", maxFiles), nil)
	}
	return verifyEntriesLoop(ctx, cz, m.Files, totalCap)
}

// verifyEntriesLoop iterates m.Files, checking ctx before each entry, and
// accumulates verified bytes up to totalCap. Extracted from verifyAllHashesCtx
// to keep each function within the project's cyclomatic-complexity cap.
func verifyEntriesLoop(ctx context.Context, cz *constrainedZip, files []ManifestEntry, totalCap int64) (map[string]verifiedBytes, int64, error) {
	verified := make(map[string]verifiedBytes, len(files))
	var total int64
	for _, e := range files {
		if err := ctx.Err(); err != nil {
			return nil, total, err
		}
		vb, err := verifyEntryWithBudget(cz, e, totalCap, total)
		if err != nil {
			return nil, total, err
		}
		total += int64(len(vb.Bytes()))
		verified[e.Path] = vb
	}
	return verified, total, nil
}

// verifyEntryWithBudget verifies a single manifest entry under an effective
// cap of min(maxPerFileBytes, totalCap-total), so the running total cap is a
// streaming bound rather than a post-allocation check.
func verifyEntryWithBudget(cz *constrainedZip, e ManifestEntry, totalCap, total int64) (verifiedBytes, error) {
	effectiveCap := totalCap - total
	if maxPerFileBytes < effectiveCap {
		effectiveCap = maxPerFileBytes
	}
	if effectiveCap <= 0 {
		return verifiedBytes{}, newError(ErrCapExceeded, e.Path, fmt.Sprintf("total uncompressed bytes would exceed %d", totalCap), nil)
	}
	return verifyAndReadEntry(cz, e.Path, e.SHA256, effectiveCap)
}

// walkRefs verifies that every transitive subgraph ref resolves to a
// manifest-listed entry, that no ref escapes workflows/, and that the
// ref graph is acyclic when rooted at any manifest-listed workflow.
//
// Cycle detection runs once per manifest-listed workflow (not only
// m.Entry) so that a cycle in a manifest-listed-but-unreachable
// workflow surfaces as ErrRefCycle instead of slipping through. See
// spec § "Open ordering" step 8.
func walkRefs(parsed map[string]*ir.Workflow, m Manifest) error {
	graph, err := buildRefGraph(parsed)
	if err != nil {
		return err
	}
	if err := verifyRefsListed(graph, m); err != nil {
		return err
	}
	return detectCyclesAll(graph, m)
}

// detectCyclesAll runs detectCycles rooted at every manifest-listed
// workflow. Each call uses a fresh color map; overlap across roots is
// re-explored, which is acceptable at manifest-cap scale (≤ a few
// hundred workflows in practice).
func detectCyclesAll(graph map[string][]string, m Manifest) error {
	for _, e := range m.Files {
		if !strings.HasSuffix(e.Path, ".dip") {
			continue // asset entries have no ref graph; cycle-check workflows only
		}
		if err := detectCycles(graph, e.Path); err != nil {
			return err
		}
	}
	return nil
}

// verifyRefsListed confirms every ref target exists in the manifest.
func verifyRefsListed(graph map[string][]string, m Manifest) error {
	listed := make(map[string]struct{}, len(m.Files))
	for _, e := range m.Files {
		listed[e.Path] = struct{}{}
	}
	for from, tos := range graph {
		for _, to := range tos {
			if _, ok := listed[to]; !ok {
				return newError(ErrRefEscape, from, "ref resolves to path not in manifest: "+to, nil)
			}
		}
	}
	return nil
}

// buildRefGraph extracts the per-workflow ref edges and resolves each ref
// against its parent's bundle path.
func buildRefGraph(parsed map[string]*ir.Workflow) (map[string][]string, error) {
	g := make(map[string][]string, len(parsed))
	for parentPath, wf := range parsed {
		out, err := refsForWorkflow(wf, parentPath)
		if err != nil {
			return nil, err
		}
		g[parentPath] = out
	}
	return g, nil
}

// refsForWorkflow resolves every ref-bearing node in wf against parentPath.
func refsForWorkflow(wf *ir.Workflow, parentPath string) ([]string, error) {
	var out []string
	for _, n := range wf.Nodes {
		refStr := refFromNode(n)
		if refStr == "" {
			continue
		}
		resolved, err := resolveLexically(refStr, parentPath)
		if err != nil {
			return nil, err
		}
		out = append(out, resolved)
	}
	return out, nil
}

// refFromNode returns the ref string for node kinds that carry one, or "".
func refFromNode(n *ir.Node) string {
	switch cfg := n.Config.(type) {
	case ir.SubgraphConfig:
		return cfg.Ref
	case ir.ManagerLoopConfig:
		return cfg.SubgraphRef
	}
	return ""
}

// normalizeConditions invokes simulate.EnsureConditionsParsed on every
// workflow so the runtime never has to call it on shared *ir.Workflow values
// (which would race in concurrent NodeParallel/NodeFanIn dispatch).
func normalizeConditions(parsed map[string]*ir.Workflow) error {
	for path, wf := range parsed {
		if err := simulate.EnsureConditionsParsed(wf); err != nil {
			return newError(ErrSubgraphParse, path, "condition normalization failed", err)
		}
	}
	return nil
}

// parseAllWorkflows parses every file in verified via parser.NewParser. THIS
// IS THE verifiedBytes-pathway CALL SITE OF parser.NewParser IN PACKAGE dipx.
// Bytes presented to the parser are obtained from verifiedBytes — a type whose
// only constructor is in the verifyHashes path — making "parse before verify"
// a structural impossibility.
//
// SPEC NOTE: The dipx package has THREE parser.NewParser sites total:
//  1. parseAllWorkflows here (Open pathway, verifiedBytes from .dipx).
//  2. parseDipFile in source.go (Source loader, raw disk bytes).
//  3. parsePackSource in helpers.go (Pack pathway, raw disk bytes parallel to
//     parseDipFile).
//
// The verifiedBytes invariant applies only to site 1. Sites 2 and 3 consume
// trusted local-disk bytes and never produce or consume verifiedBytes.
func parseAllWorkflows(verified map[string]verifiedBytes, entryPath string) (map[string]*ir.Workflow, error) {
	out := make(map[string]*ir.Workflow, len(verified))
	for path, vb := range verified {
		if !strings.HasSuffix(path, ".dip") {
			continue // non-.dip entries are opaque assets (format_version 2)
		}
		wf, err := parseOneVerifiedWorkflow(path, vb, entryPath)
		if err != nil {
			return nil, err
		}
		out[path] = wf
	}
	return out, nil
}

// parseOneVerifiedWorkflow parses a single verified .dip entry, attributing a
// parse failure to ErrEntryParse for the entry and ErrSubgraphParse otherwise.
// This is the verifiedBytes-pathway call site of parser.NewParser (site 1 of
// the three-site inventory documented on parseAllWorkflows).
func parseOneVerifiedWorkflow(path string, vb verifiedBytes, entryPath string) (*ir.Workflow, error) {
	wf, err := parser.NewParser(string(vb.Bytes()), path).Parse()
	if err != nil {
		sentinel := ErrSubgraphParse
		if path == entryPath {
			sentinel = ErrEntryParse
		}
		return nil, newError(sentinel, path, "parse failed", err)
	}
	return wf, nil
}

// packedFile is one source file collected by walkSourceTree.
type packedFile struct {
	bundlePath string // canonical, e.g. "workflows/foo.dip"
	bytes      []byte
	hash       string
}

// walkSourceTree collects the entry workflow plus every transitively-referenced
// subgraph from disk. Refuses to follow symlinks. Refuses if any ref escapes
// the entry's source root. When opts.NoInline is set it also collects directive
// and --include asset files as non-.dip bundle entries.
func walkSourceTree(ctx context.Context, entryPath string, opts PackOptions) (packedFile, []packedFile, error) {
	st, err := initPackWalkState(entryPath, opts.NoInline)
	if err != nil {
		return packedFile{}, nil, err
	}
	if err := runWalkLoop(ctx, st); err != nil {
		return packedFile{}, nil, err
	}
	return assemblePackFiles(ctx, st, opts.Include)
}

// initPackWalkState absolutizes entryPath and constructs the walk state.
func initPackWalkState(entryPath string, trackWfs bool) (*packWalkState, error) {
	entryAbs, err := filepath.Abs(entryPath)
	if err != nil {
		return nil, err
	}
	return newPackWalkState(entryAbs, filepath.Dir(entryAbs), trackWfs), nil
}

// runWalkLoop drives the BFS until the queue is empty, checking ctx per step.
func runWalkLoop(ctx context.Context, st *packWalkState) error {
	for st.hasMore() {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := st.visitNext(); err != nil {
			return err
		}
	}
	return nil
}

// assemblePackFiles returns the entry plus every packed file. In NoInline mode
// it appends the collected directive/--include assets before returning; the
// combined slice is sorted by path later in writeBundle/buildManifestForPack.
func assemblePackFiles(ctx context.Context, st *packWalkState, includes []string) (packedFile, []packedFile, error) {
	if !st.trackWfs {
		return st.entry, st.all, nil
	}
	assets, err := collectAllAssets(ctx, st, includes)
	if err != nil {
		return packedFile{}, nil, err
	}
	return st.entry, append(st.all, assets...), nil
}

// packWalkState carries iteration state for walkSourceTree so each step can be
// a small focused function under the project's complexity caps.
type packWalkState struct {
	entryAbs string
	rootDir  string
	visited  map[string]struct{}
	queue    []string
	entry    packedFile
	all      []packedFile
	trackWfs bool                    // NoInline: record parsed workflows for asset collection
	wfMap    map[string]*ir.Workflow // populated only when trackWfs; key = absPath
}

func newPackWalkState(entryAbs, rootDir string, trackWfs bool) *packWalkState {
	st := &packWalkState{
		entryAbs: entryAbs,
		rootDir:  rootDir,
		visited:  map[string]struct{}{},
		queue:    []string{entryAbs},
		trackWfs: trackWfs,
	}
	if trackWfs {
		st.wfMap = map[string]*ir.Workflow{}
	}
	return st
}

func (s *packWalkState) hasMore() bool { return len(s.queue) > 0 }

// visitNext pops the next path off the queue and processes it: read, parse,
// record as packedFile, and enqueue any transitive refs.
func (s *packWalkState) visitNext() error {
	cur := s.queue[0]
	s.queue = s.queue[1:]
	if _, ok := s.visited[cur]; ok {
		return nil
	}
	s.visited[cur] = struct{}{}
	pf, wf, err := s.readAndRecord(cur)
	if err != nil {
		return err
	}
	if cur == s.entryAbs {
		s.entry = pf
	}
	s.all = append(s.all, pf)
	if s.trackWfs {
		s.wfMap[cur] = wf
	}
	return s.enqueueRefs(cur, wf)
}

// readAndRecord reads the file, parses it, and constructs the packedFile.
// Enforces the per-file uncompressed cap (maxPerFileBytes) at Pack time so
// the producer cannot emit a bundle that fails its own round-trip in Open.
func (s *packWalkState) readAndRecord(cur string) (packedFile, *ir.Workflow, error) {
	raw, err := ReadNoFollowSymlinks(cur, s.rootDir)
	if err != nil {
		return packedFile{}, nil, err
	}
	if int64(len(raw)) > maxPerFileBytes {
		return packedFile{}, nil, newError(ErrCapExceeded, cur, fmt.Sprintf("source file exceeds %d bytes", maxPerFileBytes), nil)
	}
	wf, err := parsePackSource(cur, raw, cur == s.entryAbs)
	if err != nil {
		return packedFile{}, nil, err
	}
	bundlePath, err := bundlePathFor(cur, s.rootDir)
	if err != nil {
		return packedFile{}, nil, err
	}
	pf := packedFile{bundlePath: bundlePath, bytes: raw, hash: hashHex(raw)}
	return pf, wf, nil
}

// enqueueRefs walks wf.Nodes for refs, resolves each against cur's directory,
// confirms the result stays under the source root, and enqueues it.
func (s *packWalkState) enqueueRefs(cur string, wf *ir.Workflow) error {
	for _, n := range wf.Nodes {
		ref := refFromNode(n)
		if ref == "" {
			continue
		}
		target, err := s.resolveRefOnDisk(cur, ref)
		if err != nil {
			return err
		}
		s.queue = append(s.queue, target)
	}
	return nil
}

// resolveRefOnDisk joins ref against cur's directory and verifies the result
// stays under s.rootDir. The escape check is a literal-component compare:
// `..` alone or `../` prefix. A bare `strings.HasPrefix(rel, "..")` would
// false-positive on legitimate filenames like `..foo/bar.dip`.
func (s *packWalkState) resolveRefOnDisk(cur, ref string) (string, error) {
	target := filepath.Clean(filepath.Join(filepath.Dir(cur), ref))
	rel, err := filepath.Rel(s.rootDir, target)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", newError(ErrRefEscape, cur, "ref escapes source root: "+ref, nil)
	}
	return target, nil
}

// parsePackSource parses raw disk bytes for the Pack pathway. THIRD parser site
// in dipx (Pack pathway, parallel to parseDipFile in source.go). See note on
// parseAllWorkflows for the full inventory and justification.
//
// isEntry distinguishes the parent workflow (ErrEntryParse) from transitively-
// reached subgraphs (ErrSubgraphParse) so error attribution matches Open's
// parseAllWorkflows behavior.
func parsePackSource(path string, raw []byte, isEntry bool) (*ir.Workflow, error) {
	wf, err := parser.NewParser(string(raw), path).Parse()
	if err != nil {
		sentinel := ErrSubgraphParse
		if isEntry {
			sentinel = ErrEntryParse
		}
		return nil, newError(sentinel, path, "", err)
	}
	return wf, nil
}

// ReadNoFollowSymlinks reads a file, refusing to follow symlinks at the leaf OR
// at any intermediate path component between rootDir and path. Shared by the
// Pack walker and the cross-file tool_access lint (cmd/dippin) so both refuse
// symlinks identically. It closes a parent-component-symlink data-exfil vector:
// a tree containing `sub -> /etc` would otherwise let a leaf `sub/foo.dip` read
// `/etc/foo.dip`, because Lstat on the leaf reports a regular file, not a symlink.
//
// It does NOT perform the `..`-escape (containment) check — callers MUST run that
// separately BEFORE calling this (the pack walker via resolveRefOnDisk, the lint
// via ensureUnderRoot) and MUST pass a path already proven to be under rootDir.
// rootDir is only the ancestor-scan boundary here: if path is not under rootDir,
// filepath.Rel yields `..` components and the ancestor scan would Lstat paths
// outside rootDir.
//
// rootDir itself is treated as the trust anchor: it is an absolute path supplied
// by the caller, may itself be a user-specified symlink, and is not re-validated.
// Components strictly between rootDir and path's leaf MUST be directories that are
// not symlinks.
func ReadNoFollowSymlinks(path, rootDir string) ([]byte, error) {
	if err := assertNoSymlinkAncestor(path, rootDir); err != nil {
		return nil, err
	}
	info, err := os.Lstat(path)
	if err != nil {
		return nil, err
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return nil, newError(ErrPathUnsafe, path, "symlink in source tree", nil)
	}
	if !info.Mode().IsRegular() {
		return nil, newError(ErrPathUnsafe, path, "not a regular file", nil)
	}
	return os.ReadFile(path)
}

// assertNoSymlinkAncestor walks every path component strictly between rootDir
// and path's leaf and refuses any that is a symlink. rootDir itself is the
// trust anchor and is not Lstat'd.
func assertNoSymlinkAncestor(path, rootDir string) error {
	rel, err := filepath.Rel(rootDir, path)
	if err != nil {
		return newError(ErrPathUnsafe, path, "path not under source root", err)
	}
	parts := strings.Split(rel, string(filepath.Separator))
	cur := rootDir
	// All but the last component (which Lstat handles via the caller's
	// regular-file check). If parts has fewer than 2 elements the leaf is at
	// rootDir's level and there are no intermediate components to check.
	for i := 0; i < len(parts)-1; i++ {
		cur = filepath.Join(cur, parts[i])
		info, err := os.Lstat(cur)
		if err != nil {
			return err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return newError(ErrPathUnsafe, cur, "symlink in source tree ancestor", nil)
		}
	}
	return nil
}

// bundlePathFor converts an absolute source path under rootDir into its
// canonical bundle path (workflows/<rel>).
func bundlePathFor(absPath, rootDir string) (string, error) {
	rel, err := filepath.Rel(rootDir, absPath)
	if err != nil {
		return "", err
	}
	bundle := "workflows/" + filepath.ToSlash(rel)
	return Canonicalize(bundle)
}

// hashHex returns the lowercase hex SHA-256 of b.
func hashHex(b []byte) string {
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:])
}

// buildManifestForPack constructs a canonical Manifest from the packed files,
// with files[] sorted lexicographically by path for determinism.
func buildManifestForPack(entry packedFile, all []packedFile) Manifest {
	files := make([]ManifestEntry, 0, len(all))
	for _, pf := range all {
		files = append(files, ManifestEntry{Path: pf.bundlePath, SHA256: pf.hash})
	}
	sort.Slice(files, func(i, j int) bool { return files[i].Path < files[j].Path })
	return Manifest{
		FormatVersion: 1,
		Entry:         entry.bundlePath,
		Files:         files,
	}
}

// buildPackManifest enforces pack-time caps then routes to the v1 (inline) or
// v2 (no-inline) manifest builder.
func buildPackManifest(entry packedFile, all []packedFile, noInline bool) (Manifest, error) {
	if err := checkPackCaps(all); err != nil {
		return Manifest{}, err
	}
	if noInline {
		return buildManifestForPackV2(entry, all), nil
	}
	return buildManifestForPack(entry, all), nil
}

// buildManifestForPackV2 emits format_version 2. Kept separate from the v1
// builder so the v1 path stays byte-identical (goldens and Identity() stable).
func buildManifestForPackV2(entry packedFile, all []packedFile) Manifest {
	files := make([]ManifestEntry, 0, len(all))
	for _, pf := range all {
		files = append(files, ManifestEntry{Path: pf.bundlePath, SHA256: pf.hash})
	}
	sort.Slice(files, func(i, j int) bool { return files[i].Path < files[j].Path })
	return Manifest{
		FormatVersion: 2,
		Entry:         entry.bundlePath,
		Files:         files,
	}
}

// checkPackCaps enforces the 10 000-file and 100 MB-total caps at pack time so
// the producer cannot emit a bundle that fails its own Open (spec § Security 2).
func checkPackCaps(all []packedFile) error {
	if len(all) > maxFiles {
		return newError(ErrCapExceeded, "", fmt.Sprintf("files exceeds %d", maxFiles), nil)
	}
	var total int64
	for _, pf := range all {
		total += int64(len(pf.bytes))
	}
	if total > maxTotalUncompBytes {
		return newError(ErrCapExceeded, "", fmt.Sprintf("total uncompressed bytes would exceed %d", maxTotalUncompBytes), nil)
	}
	return nil
}

// ensureAssetUnderRoot rejects an absolute path that escapes rootDir, using the
// same lexical `..`-component check as resolveRefOnDisk.
func ensureAssetUnderRoot(abs, rootDir string) error {
	rel, err := filepath.Rel(rootDir, abs)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return newError(ErrPathUnsafe, abs, "path escapes source root", nil)
	}
	return nil
}

// readAssetFile reads a non-.dip asset through ReadNoFollowSymlinks (which
// refuses symlink leaves/ancestors and non-regular files), enforces the
// per-file cap, and returns a packedFile at its mirrored workflows/ bundle
// path. Callers MUST run ensureAssetUnderRoot first (ReadNoFollowSymlinks
// assumes containment).
func readAssetFile(abs, rootDir string) (packedFile, error) {
	if strings.HasSuffix(abs, ".dip") {
		return packedFile{}, newError(ErrPathUnsafe, abs, "assets may not end in .dip", nil)
	}
	raw, err := ReadNoFollowSymlinks(abs, rootDir)
	if err != nil {
		return packedFile{}, err
	}
	if int64(len(raw)) > maxPerFileBytes {
		return packedFile{}, newError(ErrCapExceeded, abs, fmt.Sprintf("asset file exceeds %d bytes", maxPerFileBytes), nil)
	}
	bp, err := bundlePathFor(abs, rootDir)
	if err != nil {
		return packedFile{}, err
	}
	return packedFile{bundlePath: bp, bytes: raw, hash: hashHex(raw)}, nil
}

// agentFileDirectives returns the non-empty file-directive paths on an agent
// config (prompt_file, system_prompt_file).
func agentFileDirectives(cfg ir.AgentConfig) []string {
	var out []string
	if cfg.PromptFile != "" {
		out = append(out, cfg.PromptFile)
	}
	if cfg.SystemPromptFile != "" {
		out = append(out, cfg.SystemPromptFile)
	}
	if cfg.PromptInclude != "" {
		out = append(out, cfg.PromptInclude)
	}
	return out
}

// fileDirectivesForNode returns the file-directive source paths for a node:
// command_file for tool nodes, prompt_file/system_prompt_file for agent nodes.
func fileDirectivesForNode(n *ir.Node) []string {
	switch cfg := n.Config.(type) {
	case ir.ToolConfig:
		if cfg.CommandFile != "" {
			return []string{cfg.CommandFile}
		}
	case ir.AgentConfig:
		return agentFileDirectives(cfg)
	}
	return nil
}

// resolveDirectiveTarget validates one file-directive path the same way
// parser.ResolveFileDirectives does for source/inline packs: it rejects
// absolute paths and any target that escapes the declaring workflow's own
// directory (wfDir). Without this, --no-inline could ship a file the source
// run rejects and preserve a directive that later resolution fails. It also
// enforces the assets-may-not-end-in-.dip rule before any dedup fast-path.
func resolveDirectiveTarget(rel, wfDir string) (string, error) {
	if filepath.IsAbs(rel) {
		return "", newError(ErrPathUnsafe, rel, "file directive must be a relative path", nil)
	}
	abs := filepath.Clean(filepath.Join(wfDir, rel))
	if err := ensureAssetUnderRoot(abs, wfDir); err != nil {
		return "", err
	}
	if strings.HasSuffix(abs, ".dip") {
		return "", newError(ErrPathUnsafe, abs, "assets may not end in .dip", nil)
	}
	return abs, nil
}

// collectDirectiveFile resolves one file-directive path (relative to wfDir),
// checks containment, deduplicates via visited, and reads the file as an asset.
func collectDirectiveFile(rel, wfDir, rootDir string, visited map[string]struct{}) ([]packedFile, error) {
	if rel == "" {
		return nil, nil
	}
	abs, err := resolveDirectiveTarget(rel, wfDir)
	if err != nil {
		return nil, err
	}
	if _, ok := visited[abs]; ok {
		return nil, nil
	}
	pf, err := readAssetFile(abs, rootDir)
	if err != nil {
		return nil, err
	}
	visited[abs] = struct{}{}
	return []packedFile{pf}, nil
}

// collectNodeDirectiveFiles gathers all directive assets referenced by one node.
func collectNodeDirectiveFiles(n *ir.Node, wfDir, rootDir string, visited map[string]struct{}) ([]packedFile, error) {
	var results []packedFile
	for _, rel := range fileDirectivesForNode(n) {
		pfs, err := collectDirectiveFile(rel, wfDir, rootDir, visited)
		if err != nil {
			return nil, err
		}
		results = append(results, pfs...)
	}
	return results, nil
}

// collectDirectiveAssets walks all nodes in wf collecting command_file /
// prompt_file / system_prompt_file targets, resolved relative to the
// workflow's own directory (wfAbsPath's dir).
func collectDirectiveAssets(wf *ir.Workflow, wfAbsPath, rootDir string, visited map[string]struct{}) ([]packedFile, error) {
	wfDir := filepath.Dir(wfAbsPath)
	var results []packedFile
	for _, n := range wf.Nodes {
		pfs, err := collectNodeDirectiveFiles(n, wfDir, rootDir, visited)
		if err != nil {
			return nil, err
		}
		results = append(results, pfs...)
	}
	cascade, err := collectDefaultsCascadeFiles(wf, wfDir, rootDir, visited)
	if err != nil {
		return nil, err
	}
	return append(results, cascade...), nil
}

// collectDefaultsCascadeFiles gathers the defaults-block cascade fragment files
// (prompt_prefix_file / prompt_suffix_file, #175) and the shared system-prompt
// fallback default (system_prompt_file, #72) for a workflow.
func collectDefaultsCascadeFiles(wf *ir.Workflow, wfDir, rootDir string, visited map[string]struct{}) ([]packedFile, error) {
	var results []packedFile
	for _, rel := range []string{wf.Defaults.PromptPrefixFile, wf.Defaults.PromptSuffixFile, wf.Defaults.SystemPromptFile} {
		if rel == "" {
			continue
		}
		pfs, err := collectDirectiveFile(rel, wfDir, rootDir, visited)
		if err != nil {
			return nil, err
		}
		results = append(results, pfs...)
	}
	return results, nil
}

// collectIncludeFile reads one --include leaf file as an asset and skips
// already-visited paths. The .dip rejection runs before the visited fast-path
// so an --include of a .dip already reachable as a subgraph still errors
// (rather than silently returning via dedup).
func collectIncludeFile(abs, rootDir string, visited map[string]struct{}) ([]packedFile, error) {
	if strings.HasSuffix(abs, ".dip") {
		return nil, newError(ErrPathUnsafe, abs, "assets may not end in .dip", nil)
	}
	if _, ok := visited[abs]; ok {
		return nil, nil
	}
	pf, err := readAssetFile(abs, rootDir)
	if err != nil {
		return nil, err
	}
	visited[abs] = struct{}{}
	return []packedFile{pf}, nil
}

// walkIncludeEntry is the filepath.WalkDir callback for resolveIncludeDir. It
// skips directory entries; file entries (including symlink leaves, which
// readAssetFile rejects) route through collectIncludeFile.
func walkIncludeEntry(ctx context.Context, path string, d fs.DirEntry, err error, rootDir string, visited map[string]struct{}, results *[]packedFile, found *int) error {
	if err != nil {
		return err
	}
	if ctxErr := ctx.Err(); ctxErr != nil {
		return ctxErr
	}
	if d.IsDir() {
		return nil
	}
	*found++
	pfs, err := collectIncludeFile(path, rootDir, visited)
	if err != nil {
		return err
	}
	*results = append(*results, pfs...)
	return nil
}

// resolveIncludeDir safely walks absDir collecting non-.dip files. WalkDir does
// not descend into symlinked directories; symlink leaves are re-validated by
// ReadNoFollowSymlinks inside readAssetFile. An include directory holding no
// files (empty, or only subdirectories) is an error, to catch typos rather
// than silently shipping nothing. `found` counts files before dedup, so a
// directory whose files are all already shipped does not spuriously error.
func resolveIncludeDir(ctx context.Context, absDir, rootDir string, visited map[string]struct{}) ([]packedFile, error) {
	var results []packedFile
	found := 0
	err := filepath.WalkDir(absDir, func(path string, d fs.DirEntry, err error) error {
		return walkIncludeEntry(ctx, path, d, err, rootDir, visited, &results, &found)
	})
	if err != nil {
		return nil, err
	}
	if found == 0 {
		return nil, newError(ErrPathUnsafe, absDir, "included directory contains no files", nil)
	}
	return results, nil
}

// resolveIncludePath resolves one --include value (file or directory) relative
// to entryDir, validates containment, then delegates to the dir or file reader.
func resolveIncludePath(ctx context.Context, rel, entryDir, rootDir string, visited map[string]struct{}) ([]packedFile, error) {
	if rel == "" {
		return nil, newError(ErrPathUnsafe, rel, "empty --include path", nil)
	}
	abs := filepath.Clean(filepath.Join(entryDir, rel))
	if err := ensureAssetUnderRoot(abs, rootDir); err != nil {
		return nil, err
	}
	info, err := os.Lstat(abs)
	if err != nil {
		return nil, err
	}
	if info.IsDir() {
		return resolveIncludeDir(ctx, abs, rootDir, visited)
	}
	return collectIncludeFile(abs, rootDir, visited)
}

// collectWorkflowAssets gathers directive assets from every walked workflow,
// honoring ctx cancellation between workflows.
func collectWorkflowAssets(ctx context.Context, st *packWalkState) ([]packedFile, error) {
	var assets []packedFile
	for absPath, wf := range st.wfMap {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		pfs, err := collectDirectiveAssets(wf, absPath, st.rootDir, st.visited)
		if err != nil {
			return nil, err
		}
		assets = append(assets, pfs...)
	}
	return assets, nil
}

// collectIncludeAssets resolves every --include path, honoring ctx cancellation
// between includes.
func collectIncludeAssets(ctx context.Context, st *packWalkState, includes []string) ([]packedFile, error) {
	entryDir := filepath.Dir(st.entryAbs)
	var assets []packedFile
	for _, inc := range includes {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		pfs, err := resolveIncludePath(ctx, inc, entryDir, st.rootDir, st.visited)
		if err != nil {
			return nil, err
		}
		assets = append(assets, pfs...)
	}
	return assets, nil
}

// collectAllAssets gathers directive assets from every walked workflow plus each
// --include path, deduplicating via st.visited, then sorts by bundle path for
// determinism (WalkDir and map iteration order are not guaranteed).
func collectAllAssets(ctx context.Context, st *packWalkState, includes []string) ([]packedFile, error) {
	assets, err := collectWorkflowAssets(ctx, st)
	if err != nil {
		return nil, err
	}
	inc, err := collectIncludeAssets(ctx, st, includes)
	if err != nil {
		return nil, err
	}
	assets = append(assets, inc...)
	sort.Slice(assets, func(i, j int) bool { return assets[i].bundlePath < assets[j].bundlePath })
	return assets, nil
}

// writeBundle writes a deterministic .dipx to w. manifest.json is always the
// first entry; payload entries follow in lexicographic order of bundlePath.
func writeBundle(ctx context.Context, w io.Writer, m Manifest, files []packedFile) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	zw := zip.NewWriter(w)
	manifestJSON, err := encodeManifestCanonical(m)
	if err != nil {
		return err
	}
	if err := writeZipEntry(zw, "manifest.json", manifestJSON); err != nil {
		return err
	}
	sort.Slice(files, func(i, j int) bool { return files[i].bundlePath < files[j].bundlePath })
	if err := writeAllPackedFiles(ctx, zw, files); err != nil {
		return err
	}
	return zw.Close()
}

// writeAllPackedFiles writes every packed file as a zip entry in the order
// supplied (callers sort first). Checks ctx between entries so a long
// Pack against many files can be canceled mid-write. P10.2.
func writeAllPackedFiles(ctx context.Context, zw *zip.Writer, files []packedFile) error {
	for _, pf := range files {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := writeZipEntry(zw, pf.bundlePath, pf.bytes); err != nil {
			return err
		}
	}
	return nil
}

// zipEpoch is the deterministic mtime stamped on every Pack output entry. Set
// to the ZIP epoch (1980-01-01) so two Pack runs over the same source tree
// produce byte-identical output regardless of file mtimes on disk.
var zipEpoch = time.Date(1980, 1, 1, 0, 0, 0, 0, time.UTC)

// writeZipEntry writes a single entry with fixed mtime (ZIP epoch) and 0644
// mode, no extra fields, with bit 11 (UTF-8 filename) set per spec.
//
// CRITICAL: Go's zip.Writer does NOT auto-set bit 11 for ASCII names, but
// openConstrainedZip requires it unconditionally. Setting hdr.Flags = 0x800
// here is non-negotiable for our own output to round-trip through Open.
func writeZipEntry(zw *zip.Writer, name string, body []byte) error {
	hdr := &zip.FileHeader{
		Name:     name,
		Method:   zip.Deflate,
		Modified: zipEpoch,
		Flags:    0x800,
	}
	hdr.SetMode(0o644)
	hdr.Extra = nil
	w, err := zw.CreateHeader(hdr)
	if err != nil {
		return err
	}
	_, err = w.Write(body)
	return err
}

// encodeManifestCanonical serializes m with alphabetical keys at every level
// (entry < files < format_version). Each files[] element preserves the
// ManifestEntry struct's (path, sha256) field order.
func encodeManifestCanonical(m Manifest) ([]byte, error) {
	type out struct {
		Entry         string          `json:"entry"`
		Files         []ManifestEntry `json:"files"`
		FormatVersion int             `json:"format_version"`
	}
	return json.Marshal(out{Entry: m.Entry, Files: m.Files, FormatVersion: m.FormatVersion})
}
