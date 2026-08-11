package scaffold

import (
	"testing"

	"github.com/2389-research/dippin-lang/formatter"
	"github.com/2389-research/dippin-lang/ir"
	"github.com/2389-research/dippin-lang/parser"
	"github.com/2389-research/dippin-lang/validator"
)

func TestBuild_AllTemplates_RoundTrip(t *testing.T) {
	for _, tmpl := range TemplateNames() {
		t.Run(tmpl, func(t *testing.T) {
			// Build the template.
			w, err := Build(tmpl, "")
			if err != nil {
				t.Fatalf("Build(%q) error: %v", tmpl, err)
			}

			// Format to .dip source.
			source := formatter.Format(w)
			if source == "" {
				t.Fatal("formatter produced empty output")
			}

			// Parse back from source.
			p := parser.NewParser(source, tmpl+".dip")
			w2, err := p.Parse()
			if err != nil {
				t.Fatalf("re-parse failed: %v\nsource:\n%s", err, source)
			}

			// Validate the parsed workflow.
			res := validator.Validate(w2)
			if res.HasErrors() {
				for _, d := range res.Diagnostics {
					t.Errorf("validation: %s", d.String())
				}
				t.Fatalf("template %q failed validation after round-trip\nsource:\n%s", tmpl, source)
			}
		})
	}
}

func TestBuild_CustomName(t *testing.T) {
	w, err := Build("minimal", "MyPipeline")
	if err != nil {
		t.Fatal(err)
	}
	if w.Name != "MyPipeline" {
		t.Errorf("expected name MyPipeline, got %s", w.Name)
	}
}

func TestBuild_DefaultName(t *testing.T) {
	w, err := Build("minimal", "")
	if err != nil {
		t.Fatal(err)
	}
	if w.Name != "minimal" {
		t.Errorf("expected name minimal, got %s", w.Name)
	}
}

func TestBuild_UnknownTemplate(t *testing.T) {
	_, err := Build("nosuch", "")
	if err == nil {
		t.Fatal("expected error for unknown template")
	}
}

func TestTemplateNames(t *testing.T) {
	names := TemplateNames()
	if len(names) != 6 {
		t.Errorf("expected 6 templates, got %d: %v", len(names), names)
	}
}

func TestBuildManagerLoop(t *testing.T) {
	w, err := Build("manager_loop", "Supervisor")
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if w.Name != "Supervisor" {
		t.Errorf("Name = %q, want %q", w.Name, "Supervisor")
	}
	hasManager := false
	for _, n := range w.Nodes {
		if n.Kind == ir.NodeManagerLoop {
			hasManager = true
			cfg, ok := n.Config.(ir.ManagerLoopConfig)
			if !ok {
				t.Fatalf("manager_loop node Config = %T, want ManagerLoopConfig", n.Config)
			}
			if cfg.SubgraphRef == "" {
				t.Errorf("ManagerLoopConfig.SubgraphRef is empty in template output")
			}
			if cfg.MaxCycles == 0 && cfg.StopCondition == nil {
				t.Errorf("template is unbounded — would trigger DIP137")
			}
		}
	}
	if !hasManager {
		t.Errorf("no NodeManagerLoop node in template output")
	}
}

func TestTemplateNames_IncludesManagerLoop(t *testing.T) {
	names := TemplateNames()
	for _, n := range names {
		if n == "manager_loop" {
			return
		}
	}
	t.Errorf("manager_loop missing from TemplateNames: %v", names)
}

// TestTemplatesLintClean guards that `dippin new <template>` produces a workflow
// that lints completely clean — no author sees a warning on freshly-generated
// output. manager_loop is excluded: it references an external child_pipeline.dip
// that cannot exist yet, so its DIP135 ("referenced file does not exist") is an
// informative prompt to create the child, not noise.
func TestTemplatesLintClean(t *testing.T) {
	for _, name := range TemplateNames() {
		t.Run(name, func(t *testing.T) {
			built, err := Build(name, "")
			if err != nil {
				t.Fatalf("build %s: %v", name, err)
			}
			// Mirror `dippin new` → `dippin lint`: the generated text is the
			// formatter's canonical output (which strips redundant fan edges,
			// etc.), so lint the parsed-back-from-formatted workflow, not the raw
			// builder IR.
			src := formatter.Format(built)
			w, err := parser.NewParser(src, name+".dip").Parse()
			if err != nil {
				t.Fatalf("reparse %s: %v\n%s", name, err, src)
			}
			var diags []validator.Diagnostic
			diags = append(diags, validator.Validate(w).Diagnostics...)
			diags = append(diags, validator.Lint(w).Diagnostics...)
			for _, d := range diags {
				// manager_loop legitimately emits DIP135: it references an external
				// child_pipeline.dip the author must create — informative, not noise.
				if name == "manager_loop" && d.Code == validator.DIP135 {
					continue
				}
				t.Errorf("template %q is not clean: %s[%s] %s", name, d.Severity, d.Code, d.Message)
			}
		})
	}
}
