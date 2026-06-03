package parser

import (
	"strings"
	"testing"

	"github.com/2389-research/dippin-lang/ir"
)

// parseToolFixture parses a minimal workflow with a single tool node and returns
// the parsed ToolConfig plus any diagnostics. Tests must parse real .dip text —
// hand-building ir.ToolConfig bypasses the parser and masks bugs (per CLAUDE.md).
func parseToolFixture(t *testing.T, body string) (ir.ToolConfig, []string) {
	t.Helper()
	src := "workflow W\n  start: T\n  exit: T\n\n  tool T\n" + body + "\n\n  edges\n    T -> T\n"
	p := NewParser(src, "test.dip")
	wf, _ := p.Parse()
	if len(wf.Nodes) == 0 {
		t.Fatalf("no nodes parsed; diagnostics: %v", p.Diagnostics())
	}
	cfg, ok := wf.Nodes[0].Config.(ir.ToolConfig)
	if !ok {
		t.Fatalf("node 0 is not a tool: %T", wf.Nodes[0].Config)
	}
	return cfg, p.Diagnostics()
}

func TestParseToolMarkerGrep(t *testing.T) {
	cfg, diags := parseToolFixture(t, `    marker_grep: "^(green|red)$"`)
	if len(diags) != 0 {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if cfg.MarkerGrep != "^(green|red)$" {
		t.Errorf("MarkerGrep = %q, want %q", cfg.MarkerGrep, "^(green|red)$")
	}
}

func TestParseToolMarkerGrepUnquoted(t *testing.T) {
	cfg, diags := parseToolFixture(t, "    marker_grep: tests_pass")
	if len(diags) != 0 {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if cfg.MarkerGrep != "tests_pass" {
		t.Errorf("MarkerGrep = %q, want %q", cfg.MarkerGrep, "tests_pass")
	}
}

func TestParseToolMarkerGrepRegexMetachars(t *testing.T) {
	// Stored verbatim — dippin does not validate the regex (the runtime does).
	cfg, diags := parseToolFixture(t, `    marker_grep: "^\[(PASS|FAIL)\]\s+\d+$"`)
	if len(diags) != 0 {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if !strings.Contains(cfg.MarkerGrep, `\[`) {
		t.Errorf("MarkerGrep lost metachars: %q", cfg.MarkerGrep)
	}
}

func TestParseToolRouteRequired(t *testing.T) {
	cfg, diags := parseToolFixture(t, "    route_required: true")
	if len(diags) != 0 {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if !cfg.RouteRequired {
		t.Error("RouteRequired = false, want true")
	}
}

func TestParseToolRouteRequiredExplicitFalse(t *testing.T) {
	cfg, diags := parseToolFixture(t, "    route_required: false")
	if len(diags) != 0 {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if cfg.RouteRequired {
		t.Error("RouteRequired = true, want false")
	}
}

func TestParseToolOutputLimit(t *testing.T) {
	cfg, diags := parseToolFixture(t, "    output_limit: 1048576")
	if len(diags) != 0 {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if cfg.OutputLimit != 1048576 {
		t.Errorf("OutputLimit = %d, want 1048576", cfg.OutputLimit)
	}
}

func TestParseToolOutputLimitNegative(t *testing.T) {
	_, diags := parseToolFixture(t, "    output_limit: -1")
	if len(diags) == 0 {
		t.Fatal("expected diagnostic for negative output_limit, got none")
	}
	found := false
	for _, d := range diags {
		if strings.Contains(d, "output_limit") && strings.Contains(d, "non-negative") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected 'non-negative' diagnostic, got: %v", diags)
	}
}

func TestParseToolOutputLimitNonNumeric(t *testing.T) {
	cfg, diags := parseToolFixture(t, "    output_limit: abc")
	if len(diags) == 0 {
		t.Fatal("expected diagnostic from parseInt, got none")
	}
	if cfg.OutputLimit != 0 {
		t.Errorf("OutputLimit should default to 0 on parse error, got %d", cfg.OutputLimit)
	}
}

func TestParseToolAllRoutingFields(t *testing.T) {
	body := "    marker_grep: pass\n    route_required: true\n    output_limit: 65536"
	cfg, diags := parseToolFixture(t, body)
	if len(diags) != 0 {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if cfg.MarkerGrep != "pass" || !cfg.RouteRequired || cfg.OutputLimit != 65536 {
		t.Errorf("got %+v", cfg)
	}
}
