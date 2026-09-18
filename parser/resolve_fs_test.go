package parser

import (
	"embed"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/2389-research/dippin-lang/ir"
)

// embeddedDirectives embeds the same fixture trees the disk tests read, so the
// embed.FS path (#304's motivating case) is exercised against real go:embed
// semantics rather than only fstest.MapFS.
//
//go:embed testdata/command_file testdata/prompt_file
var embeddedDirectives embed.FS

// mapFSFromDir loads every regular file under dir into a MapFS rooted at
// prefix (use "." for the root), mirroring the on-disk tree so the FS resolver
// can be compared against the disk resolver on identical inputs.
func mapFSFromDir(t *testing.T, dir, prefix string) fstest.MapFS {
	t.Helper()
	m := fstest.MapFS{}
	err := filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
		if err != nil || !d.Type().IsRegular() {
			return err
		}
		data, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(dir, p)
		if err != nil {
			return err
		}
		key := filepath.ToSlash(rel)
		if prefix != "." {
			key = prefix + "/" + key
		}
		m[key] = &fstest.MapFile{Data: data}
		return nil
	})
	if err != nil {
		t.Fatalf("load %s into MapFS: %v", dir, err)
	}
	return m
}

// parseInDir parses src as if it lived at dir/w.dip.
func parseInDir(t *testing.T, dir, src string) *ir.Workflow {
	t.Helper()
	w, err := NewParser(src, filepath.Join(dir, "w.dip")).Parse()
	if err != nil {
		t.Fatal(err)
	}
	return w
}

// parseExample parses examples/<name> from disk and returns the example dir.
func parseExample(t *testing.T, name string) (string, func() *ir.Workflow) {
	t.Helper()
	srcAbs, err := filepath.Abs(filepath.Join("..", "examples", name))
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(srcAbs)
	if err != nil {
		t.Fatal(err)
	}
	return filepath.Dir(srcAbs), func() *ir.Workflow {
		w, err := NewParser(string(data), srcAbs).Parse()
		if err != nil {
			t.Fatal(err)
		}
		return w
	}
}

func writeFiles(t *testing.T, dir string, files map[string]string) {
	t.Helper()
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

// parityCase describes one fixture tree plus a fresh-workflow constructor so
// the disk and FS resolvers each get an untouched copy of the same input.
type parityCase struct {
	name string
	dir  string
	mk   func() *ir.Workflow
}

func parityFixtureCases(t *testing.T) []parityCase {
	t.Helper()
	cmdDir, _ := filepath.Abs("testdata/command_file")
	promptDir, _ := filepath.Abs("testdata/prompt_file")
	return []parityCase{
		{"command_file", cmdDir, func() *ir.Workflow {
			return &ir.Workflow{Nodes: []*ir.Node{{ID: "A", Kind: ir.NodeTool, Config: ir.ToolConfig{CommandFile: "setup.sh"}}}}
		}},
		{"prompt_file both slots", promptDir, func() *ir.Workflow {
			return &ir.Workflow{Nodes: []*ir.Node{{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{
				PromptFile: "task.md", SystemPromptFile: "persona.md",
			}}}}
		}},
		{"prompt_file inline untouched", promptDir, func() *ir.Workflow {
			return &ir.Workflow{Nodes: []*ir.Node{{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{
				Prompt: "inline", SystemPromptFile: "persona.md",
			}}}}
		}},
	}
}

func parityTempTreeCases(t *testing.T) []parityCase {
	t.Helper()
	cascadeDir := t.TempDir()
	writeFiles(t, cascadeDir, map[string]string{"suffix.md": "END WITH STATUS", "extra.md": "EXTRA", "prefix.md": "PRE"})
	cascadeSrc := `workflow W
  start: A
  exit: B
  defaults
    prompt_prefix_file: prefix.md
    prompt_suffix_file: suffix.md
  agent A
    prompt: "body A"
    prompt_include: extra.md
  agent B
    prompt: "body B"
    prompt_suffix: none
    prompt_prefix: none
`
	sysDir := t.TempDir()
	writeFiles(t, sysDir, map[string]string{"persona.md": "SHARED PERSONA", "own.md": "OWN PERSONA"})
	sysSrc := `workflow W
  start: Inherit
  exit: OwnFile
  defaults
    system_prompt_file: persona.md
  agent Inherit
    prompt: "a"
  agent OwnInline
    prompt: "b"
    system_prompt: "OWN INLINE"
  agent OwnFile
    prompt: "c"
    system_prompt_file: own.md
`
	passDir := t.TempDir()
	writeFiles(t, passDir, map[string]string{"frag.md": "SHARED PREFIX", "inc.md": "INCLUDED"})
	passSrc := `workflow W
  start: S
  exit: E
  defaults
    prompt_prefix_file: frag.md
  agent S
    label: Start
  agent Work
    prompt: "do it"
  agent Inc
    prompt_include: inc.md
  agent E
    label: End
  edges
    S -> Work
    Work -> Inc
    Inc -> E
`
	return []parityCase{
		{"defaults cascade + include + opt-out", cascadeDir, func() *ir.Workflow { return parseInDir(t, cascadeDir, cascadeSrc) }},
		{"defaults system_prompt_file cascade", sysDir, func() *ir.Workflow { return parseInDir(t, sysDir, sysSrc) }},
		{"body-less passthrough skip", passDir, func() *ir.Workflow { return parseInDir(t, passDir, passSrc) }},
	}
}

func parityExampleCases(t *testing.T) []parityCase {
	t.Helper()
	var cases []parityCase
	for _, name := range []string{"external_files.dip", "external_prompts.dip", "shared_prompt_fragment.dip"} {
		dir, mk := parseExample(t, name)
		cases = append(cases, parityCase{"examples/" + name, dir, mk})
	}
	return cases
}

// TestResolveFileDirectivesFS_ParityWithDisk is the key guarantee of #304: for
// every fixture tree the disk tests use, resolving through a MapFS holding the
// identical files yields a deep-equal ir.Workflow to the disk resolver. Each
// tree is checked at the FS root (baseDir ".") and nested under a sub-path so
// path.Join(baseDir, p) is exercised both ways.
func TestResolveFileDirectivesFS_ParityWithDisk(t *testing.T) {
	var cases []parityCase
	cases = append(cases, parityFixtureCases(t)...)
	cases = append(cases, parityTempTreeCases(t)...)
	cases = append(cases, parityExampleCases(t)...)
	for _, tc := range cases {
		for _, base := range []string{".", "bundle/workflows"} {
			t.Run(tc.name+" @ "+base, func(t *testing.T) {
				want := tc.mk()
				if err := ResolveFileDirectives(want, tc.dir); err != nil {
					t.Fatalf("disk resolve: %v", err)
				}
				got := tc.mk()
				if err := ResolveFileDirectivesFS(got, mapFSFromDir(t, tc.dir, base), base); err != nil {
					t.Fatalf("fs resolve: %v", err)
				}
				if !reflect.DeepEqual(got, want) {
					t.Errorf("FS-resolved workflow differs from disk-resolved\n got: %#v\nwant: %#v", got, want)
				}
			})
		}
	}
}

func toolWF(commandFile string) *ir.Workflow {
	return &ir.Workflow{Nodes: []*ir.Node{{ID: "A", Kind: ir.NodeTool, Config: ir.ToolConfig{CommandFile: commandFile}}}}
}

func TestResolveFileDirectivesFS_EmbedFS(t *testing.T) {
	w := &ir.Workflow{Nodes: []*ir.Node{
		{ID: "T", Kind: ir.NodeTool, Config: ir.ToolConfig{CommandFile: "setup.sh"}},
	}}
	if err := ResolveFileDirectivesFS(w, embeddedDirectives, "testdata/command_file"); err != nil {
		t.Fatalf("ResolveFileDirectivesFS(embed): %v", err)
	}
	cfg := w.Nodes[0].Config.(ir.ToolConfig)
	if !strings.Contains(cfg.Command, "fixture: ResolveFileDirectives test") {
		t.Errorf("Command not populated from embed.FS; got %q", cfg.Command)
	}
	if cfg.CommandFile != "setup.sh" {
		t.Errorf("CommandFile = %q, want setup.sh (preserved post-resolve)", cfg.CommandFile)
	}

	// Sub-FS rooted at the fixture dir: baseDir "." against a nested embed.
	sub, err := fs.Sub(embeddedDirectives, "testdata/prompt_file")
	if err != nil {
		t.Fatal(err)
	}
	a := &ir.Workflow{Nodes: []*ir.Node{
		{ID: "A", Kind: ir.NodeAgent, Config: ir.AgentConfig{PromptFile: "task.md", SystemPromptFile: "persona.md"}},
	}}
	if err := ResolveFileDirectivesFS(a, sub, "."); err != nil {
		t.Fatalf("ResolveFileDirectivesFS(fs.Sub, \".\"): %v", err)
	}
	acfg := a.Nodes[0].Config.(ir.AgentConfig)
	if !strings.Contains(acfg.Prompt, "ResolveFileDirectives prompt test") {
		t.Errorf("Prompt not populated; got %q", acfg.Prompt)
	}
	if !strings.Contains(acfg.SystemPrompt, "ResolveFileDirectives system_prompt test") {
		t.Errorf("SystemPrompt not populated; got %q", acfg.SystemPrompt)
	}
}

func TestResolveFileDirectivesFS_RejectsAbsolutePath(t *testing.T) {
	fsys := fstest.MapFS{"etc/passwd": &fstest.MapFile{Data: []byte("x")}}
	err := ResolveFileDirectivesFS(toolWF("/etc/passwd"), fsys, ".")
	if err == nil || !strings.Contains(err.Error(), "absolute paths not allowed") {
		t.Errorf("expected absolute-path rejection; got %v", err)
	}
	if err != nil && !strings.Contains(err.Error(), "/etc/passwd") {
		t.Errorf("error should reference user-written path; got %v", err)
	}
}

func TestResolveFileDirectivesFS_RejectsParentRef(t *testing.T) {
	fsys := fstest.MapFS{
		"secret.txt":     &fstest.MapFile{Data: []byte("secret")},
		"wf/legit/x.txt": &fstest.MapFile{Data: []byte("x")},
	}
	for _, p := range []string{"../secret.txt", "legit/../../secret.txt", ".."} {
		w := toolWF(p)
		err := ResolveFileDirectivesFS(w, fsys, "wf")
		if err == nil || !strings.Contains(err.Error(), "resolves outside source directory") {
			t.Errorf("%q: expected parent-ref rejection; got %v", p, err)
		}
		if err != nil && !strings.Contains(err.Error(), p) {
			t.Errorf("%q: error should reference user-written path; got %v", p, err)
		}
		if cfg := w.Nodes[0].Config.(ir.ToolConfig); strings.Contains(cfg.Command, "secret") {
			t.Errorf("%q: read escaped baseDir", p)
		}
	}
}

func TestResolveFileDirectivesFS_MissingFileNamesUserPath(t *testing.T) {
	fsys := fstest.MapFS{"wf/other.sh": &fstest.MapFile{Data: []byte("x")}}
	err := ResolveFileDirectivesFS(toolWF("nonexistent.sh"), fsys, "wf")
	if err == nil {
		t.Fatal("expected missing-file error; got nil")
	}
	if !strings.Contains(err.Error(), `file "nonexistent.sh" not found`) {
		t.Errorf("error should name the user path only, same shape as disk; got %v", err)
	}
	if strings.Contains(err.Error(), "wf/nonexistent.sh") {
		t.Errorf("error leaked the joined FS path; got %v", err)
	}
	if !strings.Contains(err.Error(), `node "A" command_file`) {
		t.Errorf("error should identify node and directive; got %v", err)
	}
}

func TestResolveFileDirectivesFS_SizeCap(t *testing.T) {
	big := make([]byte, maxDirectiveFileSize+1)
	atCap := make([]byte, maxDirectiveFileSize)
	fsys := fstest.MapFS{
		"big.sh":   &fstest.MapFile{Data: big},
		"atcap.sh": &fstest.MapFile{Data: atCap},
	}
	err := ResolveFileDirectivesFS(toolWF("big.sh"), fsys, ".")
	if err == nil || !strings.Contains(err.Error(), "too large") || !strings.Contains(err.Error(), "max 4 MiB") {
		t.Errorf("expected oversize rejection with MiB cap; got %v", err)
	}
	w := toolWF("atcap.sh")
	if err := ResolveFileDirectivesFS(w, fsys, "."); err != nil {
		t.Fatalf("file exactly at cap must load; got %v", err)
	}
	if got := len(w.Nodes[0].Config.(ir.ToolConfig).Command); got != maxDirectiveFileSize {
		t.Errorf("at-cap file truncated: got %d bytes, want %d", got, maxDirectiveFileSize)
	}
}

func TestResolveFileDirectivesFS_RejectsDirectory(t *testing.T) {
	fsys := fstest.MapFS{"wf/sub/inner.sh": &fstest.MapFile{Data: []byte("x")}}
	err := ResolveFileDirectivesFS(toolWF("sub"), fsys, "wf")
	if err == nil {
		t.Fatal("expected directory-as-file rejection; got nil")
	}
	if !strings.Contains(err.Error(), `"sub"`) {
		t.Errorf("error should name the user path; got %v", err)
	}
	if strings.Contains(err.Error(), "wf/sub") {
		t.Errorf("error leaked the joined FS path; got %v", err)
	}
}

func TestResolveFileDirectivesFS_SkipsInlineAndEmpty(t *testing.T) {
	w := &ir.Workflow{Nodes: []*ir.Node{
		{ID: "A", Kind: ir.NodeTool, Config: ir.ToolConfig{Command: "inline"}},
		{ID: "B", Kind: ir.NodeAgent, Config: ir.AgentConfig{Prompt: "p"}},
	}}
	if err := ResolveFileDirectivesFS(w, fstest.MapFS{}, "."); err != nil {
		t.Fatalf("no directives should resolve cleanly against an empty FS; got %v", err)
	}
	if got := w.Nodes[0].Config.(ir.ToolConfig).Command; got != "inline" {
		t.Errorf("inline Command modified: %q", got)
	}
}

// symlinkRoot builds a temp root for os.DirFS symlink tests: root/wf/real.sh
// (regular), root/wf/inlink -> real.sh (in-root leaf symlink), root/wf/link ->
// <outside>/secret.txt, root/wf/linkdir -> <outside dir>. Skips when the
// platform cannot create symlinks.
func symlinkRoot(t *testing.T) (root, outside string) {
	t.Helper()
	root = t.TempDir()
	outside = t.TempDir()
	wf := filepath.Join(root, "wf")
	if err := os.Mkdir(wf, 0o755); err != nil {
		t.Fatal(err)
	}
	writeFiles(t, wf, map[string]string{"real.sh": "echo real"})
	writeFiles(t, outside, map[string]string{"secret.txt": "TOP-SECRET-OUTSIDE"})
	links := map[string]string{
		"inlink":  filepath.Join(wf, "real.sh"),
		"link":    filepath.Join(outside, "secret.txt"),
		"linkdir": outside,
	}
	for name, target := range links {
		if err := os.Symlink(target, filepath.Join(wf, name)); err != nil {
			t.Skipf("symlink not supported on this platform: %v", err)
		}
	}
	return root, outside
}

// TestResolveFileDirectivesFS_DirFSRejectsSymlinks covers the P1 review
// finding: os.DirFS follows symlinks on Open, so a ReadLinkFS-capable FS must
// have every component of the directive path Lstat-checked, mirroring the disk
// resolver (no leaf symlink, no symlinked parent).
func TestResolveFileDirectivesFS_DirFSRejectsSymlinks(t *testing.T) {
	root, outside := symlinkRoot(t)
	fsys := os.DirFS(root)
	cases := []struct{ p, want string }{
		{"link", "symlinks not allowed"},                            // leaf -> outside
		{"linkdir/secret.txt", "resolves outside source directory"}, // symlinked parent -> outside
		{"inlink", "symlinks not allowed"},                          // leaf symlink that stays in-root: O_NOFOLLOW parity
	}
	for _, tc := range cases {
		w := toolWF(tc.p)
		err := ResolveFileDirectivesFS(w, fsys, "wf")
		if err == nil || !strings.Contains(err.Error(), tc.want) {
			t.Errorf("%q: want %q rejection; got %v", tc.p, tc.want, err)
			continue
		}
		if !strings.Contains(err.Error(), tc.p) {
			t.Errorf("%q: error should name the user path; got %v", tc.p, err)
		}
		if strings.Contains(err.Error(), root) || strings.Contains(err.Error(), outside) || strings.Contains(err.Error(), "wf/") {
			t.Errorf("%q: error leaked a resolved or joined path; got %v", tc.p, err)
		}
		if cfg := w.Nodes[0].Config.(ir.ToolConfig); strings.Contains(cfg.Command, "SECRET") || strings.Contains(cfg.Command, "real") {
			t.Errorf("%q: content was loaded through a symlink", tc.p)
		}
	}
	// The real file next to the links still loads.
	w := toolWF("real.sh")
	if err := ResolveFileDirectivesFS(w, fsys, "wf"); err != nil {
		t.Fatalf("regular file beside symlinks must load; got %v", err)
	}
	if got := w.Nodes[0].Config.(ir.ToolConfig).Command; got != "echo real" {
		t.Errorf("Command = %q, want %q", got, "echo real")
	}
}

// TestResolveFileDirectivesFS_MapFSSymlinkEntryRejected: fstest.MapFS is a
// ReadLinkFS too, so a ModeSymlink entry is rejected the same way — and a plain
// MapFS without symlinks is unaffected (the parity test covers that at scale).
func TestResolveFileDirectivesFS_MapFSSymlinkEntryRejected(t *testing.T) {
	fsys := fstest.MapFS{
		"wf/real.sh": &fstest.MapFile{Data: []byte("echo real")},
		"wf/link.sh": &fstest.MapFile{Data: []byte("real.sh"), Mode: fs.ModeSymlink},
	}
	err := ResolveFileDirectivesFS(toolWF("link.sh"), fsys, "wf")
	if err == nil || !strings.Contains(err.Error(), "symlinks not allowed") {
		t.Errorf("expected MapFS symlink entry rejection; got %v", err)
	}
	if err := ResolveFileDirectivesFS(toolWF("real.sh"), fsys, "wf"); err != nil {
		t.Errorf("regular MapFS entry must still load; got %v", err)
	}
}
