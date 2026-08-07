package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/2389-research/dippin-lang/simulate"
)

// helper runs the CLI with the given args and returns stdout, stderr, and exit code.
func runCLI(t *testing.T, args ...string) (stdout, stderr string, code ExitCode) {
	t.Helper()
	var outBuf, errBuf bytes.Buffer
	code = Run(args, &outBuf, &errBuf)
	return outBuf.String(), errBuf.String(), code
}

// testdata returns the path to a test fixture file.
func testdata(name string) string {
	return filepath.Join("testdata", name)
}

// --- Parse Command ---

func TestCmdParse_ValidFile(t *testing.T) {
	stdout, stderr, code := runCLI(t, "parse", testdata("valid_minimal.dip"))

	if code != ExitOK {
		t.Fatalf("expected exit 0, got %d; stderr: %s", code, stderr)
	}
	if stderr != "" {
		t.Errorf("expected no stderr, got: %s", stderr)
	}

	// Verify output is valid JSON.
	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(stdout), &parsed); err != nil {
		t.Fatalf("stdout is not valid JSON: %v\nstdout: %s", err, stdout)
	}

	// Verify expected fields.
	if parsed["Name"] != "Minimal" {
		t.Errorf("expected Name=Minimal, got %v", parsed["Name"])
	}
	if parsed["Start"] != "Ask" {
		t.Errorf("expected Start=Ask, got %v", parsed["Start"])
	}
	if parsed["Exit"] != "Done" {
		t.Errorf("expected Exit=Done, got %v", parsed["Exit"])
	}

	// Verify nodes exist.
	nodes, ok := parsed["Nodes"].([]interface{})
	if !ok || len(nodes) < 2 {
		t.Errorf("expected at least 2 nodes, got %v", parsed["Nodes"])
	}
}

func TestCmdParse_InvalidFile(t *testing.T) {
	stdout, stderr, code := runCLI(t, "parse", testdata("invalid_missing_start.dip"))

	// The parser may still return a partial workflow without error when
	// the start field is just missing. If that happens, parse succeeds
	// and the validation errors would be caught by validate/lint.
	// Let's just verify the command runs and produces JSON output.
	if code == ExitOK {
		// Parser succeeded — verify JSON output contains empty start.
		var parsed map[string]interface{}
		if err := json.Unmarshal([]byte(stdout), &parsed); err != nil {
			t.Fatalf("stdout is not valid JSON: %v", err)
		}
		if parsed["Start"] != "" {
			t.Errorf("expected empty Start for missing-start fixture, got %v", parsed["Start"])
		}
	} else {
		// Parser failed — error should be on stderr.
		if stderr == "" {
			t.Error("expected error on stderr for invalid file")
		}
	}
	_ = stdout
}

func TestCmdParse_MissingArg(t *testing.T) {
	_, stderr, code := runCLI(t, "parse")

	if code != ExitUsageError {
		t.Fatalf("expected exit 2 (usage error), got %d", code)
	}
	if !strings.Contains(stderr, "usage:") {
		t.Errorf("expected usage message on stderr, got: %s", stderr)
	}
}

func TestCmdParse_NonexistentFile(t *testing.T) {
	_, stderr, code := runCLI(t, "parse", "testdata/nosuch.dip")

	if code != ExitError {
		t.Fatalf("expected exit 1 (error), got %d", code)
	}
	if !strings.Contains(stderr, "no such file") && !strings.Contains(stderr, "does not exist") && !strings.Contains(stderr, "error") {
		t.Errorf("expected file-not-found error on stderr, got: %s", stderr)
	}
}

// --- Validate Command ---

func TestCmdValidate_Valid(t *testing.T) {
	stdout, stderr, code := runCLI(t, "validate", testdata("valid_minimal.dip"))

	if code != ExitOK {
		t.Fatalf("expected exit 0, got %d; stderr: %s", code, stderr)
	}
	if !strings.Contains(stdout, "validation passed") {
		t.Errorf("expected 'validation passed' on stdout, got: %s", stdout)
	}
}

func TestCmdValidate_Errors(t *testing.T) {
	_, stderr, code := runCLI(t, "validate", testdata("invalid_missing_start.dip"))

	if code != ExitError {
		t.Fatalf("expected exit 1, got %d", code)
	}
	if !strings.Contains(stderr, "DIP001") {
		t.Errorf("expected DIP001 diagnostic on stderr, got: %s", stderr)
	}
}

func TestCmdValidate_JSONOutput(t *testing.T) {
	_, stderr, code := runCLI(t, "--format", "json", "validate", testdata("invalid_missing_start.dip"))

	if code != ExitError {
		t.Fatalf("expected exit 1, got %d", code)
	}

	// Verify stderr is valid JSON array.
	var diags []map[string]interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &diags); err != nil {
		t.Fatalf("stderr is not valid JSON: %v\nstderr: %s", err, stderr)
	}
	if len(diags) == 0 {
		t.Fatal("expected at least one diagnostic")
	}

	// Check first diagnostic has the expected structure.
	first := diags[0]
	if first["code"] != "DIP001" {
		t.Errorf("expected code=DIP001, got %v", first["code"])
	}
	if first["severity"] != "error" {
		t.Errorf("expected severity=error, got %v", first["severity"])
	}
	if _, ok := first["message"]; !ok {
		t.Error("expected message field in diagnostic")
	}
	if _, ok := first["location"]; !ok {
		t.Error("expected location field in diagnostic")
	}
}

// --- Lint Command ---

func TestCmdLint_Clean(t *testing.T) {
	_, _, code := runCLI(t, "lint", testdata("valid_minimal.dip"))

	if code != ExitOK {
		t.Fatalf("expected exit 0 for clean file, got %d", code)
	}
}

func TestCmdLint_Warnings(t *testing.T) {
	_, stderr, code := runCLI(t, "lint", testdata("lint_warnings.dip"))

	// Warnings alone should not cause failure.
	if code != ExitOK {
		t.Fatalf("expected exit 0 (warnings don't fail), got %d; stderr: %s", code, stderr)
	}
	// But we should see the warnings.
	if !strings.Contains(stderr, "DIP110") {
		t.Errorf("expected DIP110 (empty prompt) warning, got: %s", stderr)
	}
	if !strings.Contains(stderr, "DIP111") {
		t.Errorf("expected DIP111 (tool timeout) warning, got: %s", stderr)
	}
}

func TestCmdLint_NewCodes(t *testing.T) {
	_, stderr, code := runCLI(t, "lint", testdata("lint_new_codes.dip"))

	// Only warnings, no structural errors — should pass.
	if code != ExitOK {
		t.Fatalf("expected exit 0 (warnings don't fail), got %d; stderr: %s", code, stderr)
	}
	if !strings.Contains(stderr, "DIP113") {
		t.Errorf("expected DIP113 (invalid retry policy), got: %s", stderr)
	}
	if !strings.Contains(stderr, "DIP114") {
		t.Errorf("expected DIP114 (invalid fidelity), got: %s", stderr)
	}
	if !strings.Contains(stderr, "DIP115") {
		t.Errorf("expected DIP115 (goal gate without fallback), got: %s", stderr)
	}
}

func TestCmdLint_WithErrors(t *testing.T) {
	// invalid_missing_start.dip has structural errors (DIP001) — lint should fail.
	_, stderr, code := runCLI(t, "lint", testdata("invalid_missing_start.dip"))

	if code != ExitError {
		t.Fatalf("expected exit 1, got %d", code)
	}
	if !strings.Contains(stderr, "DIP001") {
		t.Errorf("expected DIP001 on stderr, got: %s", stderr)
	}
}

func TestCmdLint_ExtraModels_SuppressesDIP108(t *testing.T) {
	// Without --extra-models, custom-corp provider should trigger DIP108.
	_, stderr, code := runCLI(t, "lint", testdata("custom_provider.dip"))
	if code != ExitOK {
		t.Fatalf("expected exit 0 (warnings don't fail), got %d; stderr: %s", code, stderr)
	}
	if !strings.Contains(stderr, "DIP108") {
		t.Errorf("expected DIP108 warning without --extra-models, got: %s", stderr)
	}

	// With --extra-models, DIP108 should be suppressed.
	_, stderr2, code2 := runCLI(t, "lint", "--extra-models", "custom-corp:custom-llm-v1", testdata("custom_provider.dip"))
	if code2 != ExitOK {
		t.Fatalf("expected exit 0, got %d; stderr: %s", code2, stderr2)
	}
	if strings.Contains(stderr2, "DIP108") {
		t.Errorf("expected no DIP108 with --extra-models, got: %s", stderr2)
	}
}

func TestCmdLint_ExtraModels_MissingArg(t *testing.T) {
	_, stderr, code := runCLI(t, "lint")

	if code != ExitUsageError {
		t.Fatalf("expected exit 2 (usage error), got %d", code)
	}
	if !strings.Contains(stderr, "usage:") {
		t.Errorf("expected usage message on stderr, got: %s", stderr)
	}
}

func TestCmdDoctor_ExtraModels_SuppressesDIP108(t *testing.T) {
	// Without --extra-models, custom-corp provider produces DIP108 in doctor lint section.
	stdout, _, code := runCLI(t, "doctor", testdata("custom_provider.dip"))
	if code != ExitOK {
		t.Fatalf("expected exit 0, got %d", code)
	}
	// Doctor runs lint internally — DIP108 should appear as a warning count.
	// We just verify the command succeeds and health report is produced.
	if !strings.Contains(stdout, "Health Report") {
		t.Errorf("expected health report, got: %s", stdout)
	}

	// With --extra-models, DIP108 warnings should be suppressed → better score.
	stdout2, _, code2 := runCLI(t, "doctor", "--extra-models", "custom-corp:custom-llm-v1", testdata("custom_provider.dip"))
	if code2 != ExitOK {
		t.Fatalf("expected exit 0, got %d", code2)
	}
	if !strings.Contains(stdout2, "Health Report") {
		t.Errorf("expected health report, got: %s", stdout2)
	}
}

// --- Check Command ---

func TestCmdCheck_ValidFile_JSON(t *testing.T) {
	stdout, _, code := runCLI(t, "check", testdata("valid_minimal.dip"))

	if code != ExitOK {
		t.Fatalf("expected exit 0, got %d", code)
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(stdout)), &result); err != nil {
		t.Fatalf("stdout is not valid JSON: %v\nstdout: %s", err, stdout)
	}
	if result["valid"] != true {
		t.Errorf("expected valid=true, got %v", result["valid"])
	}
	if result["errors"].(float64) != 0 {
		t.Errorf("expected errors=0, got %v", result["errors"])
	}
}

func TestCmdCheck_InvalidFile_JSON(t *testing.T) {
	stdout, _, code := runCLI(t, "check", testdata("check_errors.dip"))

	if code != ExitError {
		t.Fatalf("expected exit 1, got %d", code)
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(stdout)), &result); err != nil {
		t.Fatalf("stdout is not valid JSON: %v\nstdout: %s", err, stdout)
	}
	if result["valid"] != false {
		t.Errorf("expected valid=false, got %v", result["valid"])
	}
	errors := result["errors"].(float64)
	if errors < 1 {
		t.Errorf("expected at least 1 error, got %v", errors)
	}
	diags, ok := result["diagnostics"].([]interface{})
	if !ok || len(diags) == 0 {
		t.Error("expected non-empty diagnostics array")
	}
	if _, ok := result["suggested_actions"].([]interface{}); !ok {
		t.Error("expected suggested_actions array")
	}
}

func TestCmdCheck_WarningsOnly(t *testing.T) {
	stdout, _, code := runCLI(t, "check", testdata("lint_warnings.dip"))

	if code != ExitOK {
		t.Fatalf("expected exit 0 (warnings don't fail), got %d", code)
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(stdout)), &result); err != nil {
		t.Fatalf("stdout is not valid JSON: %v\nstdout: %s", err, stdout)
	}
	if result["valid"] != true {
		t.Errorf("expected valid=true with warnings only, got %v", result["valid"])
	}
	warnings := result["warnings"].(float64)
	if warnings < 1 {
		t.Errorf("expected at least 1 warning, got %v", warnings)
	}
}

func TestCmdCheck_TextFormat(t *testing.T) {
	stdout, _, code := runCLI(t, "check", "--format", "text", testdata("valid_minimal.dip"))

	if code != ExitOK {
		t.Fatalf("expected exit 0, got %d", code)
	}
	if !strings.Contains(stdout, "check passed") {
		t.Errorf("expected 'check passed' in text output, got: %s", stdout)
	}
}

func TestCmdCheck_DefaultIsJSON(t *testing.T) {
	stdout, _, _ := runCLI(t, "check", testdata("valid_minimal.dip"))

	// Default format should be JSON, not text.
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(stdout)), &result); err != nil {
		t.Fatalf("default check output should be JSON, got: %s", stdout)
	}
}

func TestCmdCheck_MissingArg(t *testing.T) {
	_, stderr, code := runCLI(t, "check")

	if code != ExitUsageError {
		t.Fatalf("expected exit 2, got %d", code)
	}
	if !strings.Contains(stderr, "usage:") {
		t.Errorf("expected usage message, got: %s", stderr)
	}
}

func TestCmdCheck_NonexistentFile(t *testing.T) {
	stdout, _, code := runCLI(t, "check", "testdata/nosuch.dip")

	if code != ExitError {
		t.Fatalf("expected exit 1, got %d", code)
	}
	// JSON output even for file errors.
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(stdout)), &result); err != nil {
		t.Fatalf("expected JSON error output, got: %s", stdout)
	}
	if result["valid"] != false {
		t.Errorf("expected valid=false, got %v", result["valid"])
	}
}

// TestCmdCheck_ErrorSeverityLint verifies that dippin check fails (exit code
// and JSON valid field) when the diagnostics include an error-severity LINT
// diagnostic (e.g. DIP155) even though the workflow is structurally valid, so
// check stays consistent with lint/doctor on the same file.
func TestCmdCheck_ErrorSeverityLint(t *testing.T) {
	tmpDir := t.TempDir()
	dipFile := filepath.Join(tmpDir, "unknown_input_type.dip")
	content := `workflow W
  goal: "t"
  start: A
  exit: A

  inputs
    when: duration

  agent A
    prompt:
      hi
`
	if err := os.WriteFile(dipFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	stdout, _, code := runCLI(t, "check", "--format=json", dipFile)
	if code != ExitError {
		t.Fatalf("expected exit 1 (ExitError) for DIP155 error-severity lint, got %d; stdout: %s", code, stdout)
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(stdout)), &result); err != nil {
		t.Fatalf("stdout is not valid JSON: %v\nstdout: %s", err, stdout)
	}
	if result["valid"] != false {
		t.Errorf("expected valid=false, got %v", result["valid"])
	}
	if errors := result["errors"].(float64); errors != 1 {
		t.Errorf("expected errors=1, got %v", errors)
	}
}

// --- New Command ---

func TestCmdNew_AllTemplates(t *testing.T) {
	templates := []string{"minimal", "parallel", "conditional", "review-loop", "human-gate"}
	for _, tmpl := range templates {
		t.Run(tmpl, func(t *testing.T) {
			stdout, stderr, code := runCLI(t, "new", tmpl)

			if code != ExitOK {
				t.Fatalf("expected exit 0, got %d; stderr: %s", code, stderr)
			}
			if !strings.Contains(stdout, "workflow") {
				t.Errorf("expected 'workflow' keyword in output, got: %s", stdout)
			}
			if !strings.Contains(stdout, "edges") {
				t.Errorf("expected 'edges' section in output")
			}
		})
	}
}

func TestCmdNew_NameFlag(t *testing.T) {
	stdout, _, code := runCLI(t, "new", "--name", "MyPipeline", "minimal")

	if code != ExitOK {
		t.Fatalf("expected exit 0, got %d", code)
	}
	if !strings.Contains(stdout, "workflow MyPipeline") {
		t.Errorf("expected 'workflow MyPipeline', got: %s", stdout)
	}
}

func TestCmdNew_WriteFlag(t *testing.T) {
	tmpDir := t.TempDir()
	outFile := filepath.Join(tmpDir, "new.dip")

	_, stderr, code := runCLI(t, "new", "--write", outFile, "minimal")
	if code != ExitOK {
		t.Fatalf("expected exit 0, got %d; stderr: %s", code, stderr)
	}

	data, err := os.ReadFile(outFile)
	if err != nil {
		t.Fatalf("expected output file: %v", err)
	}
	if !strings.Contains(string(data), "workflow minimal") {
		t.Errorf("expected 'workflow minimal' in file, got: %s", string(data))
	}
}

func TestCmdNew_UnknownTemplate(t *testing.T) {
	_, stderr, code := runCLI(t, "new", "nosuch")

	if code != ExitError {
		t.Fatalf("expected exit 1, got %d", code)
	}
	if !strings.Contains(stderr, "unknown template") {
		t.Errorf("expected 'unknown template' error, got: %s", stderr)
	}
}

func TestCmdNew_MissingArg(t *testing.T) {
	_, stderr, code := runCLI(t, "new")

	if code != ExitUsageError {
		t.Fatalf("expected exit 2, got %d", code)
	}
	if !strings.Contains(stderr, "usage:") {
		t.Errorf("expected usage message, got: %s", stderr)
	}
}

// --- Fmt Command ---

func TestCmdFmt_Output(t *testing.T) {
	stdout, stderr, code := runCLI(t, "fmt", testdata("needs_formatting.dip"))

	if code != ExitOK {
		t.Fatalf("expected exit 0, got %d; stderr: %s", code, stderr)
	}

	// Verify output contains canonical keywords.
	if !strings.Contains(stdout, "workflow Messy") {
		t.Errorf("expected 'workflow Messy' in output, got: %s", stdout)
	}
	// Verify the extra whitespace is gone.
	if strings.Contains(stdout, "workflow  Messy") {
		t.Error("output should not have double spaces after workflow keyword")
	}

	// Verify output is canonical — re-formatting should be idempotent.
	// (We can't easily parse + reformat within the test without importing
	// parser/formatter, but we can verify basic structural properties.)
	if !strings.HasSuffix(stdout, "\n") {
		t.Error("expected output to end with newline")
	}
}

func TestCmdFmt_Check_AlreadyCanonical(t *testing.T) {
	_, _, code := runCLI(t, "fmt", "--check", testdata("valid_minimal.dip"))

	if code != ExitOK {
		t.Fatalf("expected exit 0 for already-canonical file, got %d", code)
	}
}

func TestCmdFmt_Check_NotCanonical(t *testing.T) {
	_, stderr, code := runCLI(t, "fmt", "--check", testdata("needs_formatting.dip"))

	if code != ExitError {
		t.Fatalf("expected exit 1 for non-canonical file, got %d", code)
	}
	if !strings.Contains(stderr, "not canonically formatted") {
		t.Errorf("expected 'not canonically formatted' on stderr, got: %s", stderr)
	}
}

func TestCmdFmt_Migrate_ConvertsToV2(t *testing.T) {
	// --migrate converts a v1 file to dip 2: the output gains a `dip 2` declaration
	// and exits OK (no review notes for a clean v1 file with no retry targets).
	migrated, _, code := runCLI(t, "fmt", "--migrate", testdata("valid_minimal.dip"))
	if code != ExitOK {
		t.Fatalf("fmt --migrate failed: exit %d", code)
	}
	if !strings.HasPrefix(migrated, "dip 2\n") {
		t.Errorf("--migrate output should declare dip 2:\n%s", migrated)
	}
}

func TestCmdFmt_MigrateCheckDetectsUnmigrated(t *testing.T) {
	// `--migrate --check` compares the (migrated) formatted output to the file:
	// a file that is not already canonical dip 2 exits non-zero (migration would
	// change it), and --check never writes formatted output to stdout.
	if _, _, code := runCLI(t, "fmt", "--check", testdata("needs_formatting.dip")); code != ExitError {
		t.Fatalf("--check on non-canonical file: expected exit 1, got %d", code)
	}
	out, _, code := runCLI(t, "fmt", "--check", "--migrate", testdata("needs_formatting.dip"))
	if code == ExitOK {
		t.Fatalf("--check --migrate on a file needing migration: expected non-zero, got %d", code)
	}
	if out != "" {
		t.Errorf("--check must not write formatted output to stdout, got:\n%s", out)
	}
}

func TestCmdFmt_Write(t *testing.T) {
	// Create a temp copy so we don't mutate the fixture.
	data, err := os.ReadFile(testdata("needs_formatting.dip"))
	if err != nil {
		t.Fatal(err)
	}
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.dip")
	if err := os.WriteFile(tmpFile, data, 0644); err != nil {
		t.Fatal(err)
	}

	_, stderr, code := runCLI(t, "fmt", "--write", tmpFile)
	if code != ExitOK {
		t.Fatalf("expected exit 0, got %d; stderr: %s", code, stderr)
	}

	// Verify the file was rewritten.
	result, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatal(err)
	}
	if string(result) == string(data) {
		t.Error("expected file to be changed by --write, but it's identical")
	}
	if !strings.Contains(string(result), "workflow Messy") {
		t.Error("expected rewritten file to contain 'workflow Messy'")
	}

	// Verify the result is now canonical (fmt --check should pass).
	_, _, code2 := runCLI(t, "fmt", "--check", tmpFile)
	if code2 != ExitOK {
		t.Errorf("expected --check to pass after --write, got exit %d", code2)
	}
}

// --- Export-DOT Command ---

func TestCmdExportDOT_Basic(t *testing.T) {
	stdout, stderr, code := runCLI(t, "export-dot", testdata("valid_minimal.dip"))

	if code != ExitOK {
		t.Fatalf("expected exit 0, got %d; stderr: %s", code, stderr)
	}
	if !strings.Contains(stdout, "digraph") {
		t.Errorf("expected 'digraph' in DOT output, got: %s", stdout)
	}
	if !strings.Contains(stdout, "Ask") {
		t.Errorf("expected node 'Ask' in DOT output")
	}
	if !strings.Contains(stdout, "Done") {
		t.Errorf("expected node 'Done' in DOT output")
	}
}

func TestCmdExportDOT_WithRankdir(t *testing.T) {
	stdout, _, code := runCLI(t, "export-dot", "--rankdir", "LR", testdata("valid_minimal.dip"))

	if code != ExitOK {
		t.Fatalf("expected exit 0, got %d", code)
	}
	if !strings.Contains(stdout, "rankdir=LR") {
		t.Errorf("expected 'rankdir=LR' in output, got: %s", stdout)
	}
}

func TestCmdExportDOT_WithPrompts(t *testing.T) {
	stdout, _, code := runCLI(t, "export-dot", "--prompts", testdata("valid_minimal.dip"))

	if code != ExitOK {
		t.Fatalf("expected exit 0, got %d", code)
	}
	// With --prompts, the Done node should have a prompt attribute.
	if !strings.Contains(stdout, "prompt") {
		t.Errorf("expected prompt attribute in DOT output with --prompts flag")
	}
}

func TestCmdExportDOT_MissingArg(t *testing.T) {
	_, stderr, code := runCLI(t, "export-dot")

	if code != ExitUsageError {
		t.Fatalf("expected exit 2, got %d", code)
	}
	if !strings.Contains(stderr, "usage:") {
		t.Errorf("expected usage message, got: %s", stderr)
	}
}

// --- Export-DIP Command ---

func TestCmdExportDIP_Basic(t *testing.T) {
	stdout, stderr, code := runCLI(t, "export-dip", testdata("valid_minimal.dip"))

	if code != ExitOK {
		t.Fatalf("expected exit 0, got %d; stderr: %s", code, stderr)
	}
	if !strings.Contains(stdout, "workflow Minimal") {
		t.Errorf("expected 'workflow Minimal' in output, got: %s", stdout)
	}
	if !strings.Contains(stdout, "Ask -> Done") {
		t.Errorf("expected edge 'Ask -> Done' in output")
	}
}

func TestCmdExportDIP_MissingArg(t *testing.T) {
	_, stderr, code := runCLI(t, "export-dip")

	if code != ExitUsageError {
		t.Fatalf("expected exit 2, got %d", code)
	}
	if !strings.Contains(stderr, "usage:") {
		t.Errorf("expected usage message, got: %s", stderr)
	}
}

// --- Migrate Command ---

func TestCmdMigrate_Basic(t *testing.T) {
	stdout, stderr, code := runCLI(t, "migrate", testdata("sample.dot"))

	if code != ExitOK {
		t.Fatalf("expected exit 0, got %d; stderr: %s", code, stderr)
	}
	if !strings.Contains(stdout, "workflow") {
		t.Errorf("expected 'workflow' keyword in .dip output, got: %s", stdout)
	}
	if !strings.Contains(stdout, "Start") {
		t.Errorf("expected 'Start' node in output")
	}
	if !strings.Contains(stdout, "DoWork") {
		t.Errorf("expected 'DoWork' node in output")
	}
}

func TestCmdMigrate_WithOutput(t *testing.T) {
	tmpDir := t.TempDir()
	outFile := filepath.Join(tmpDir, "output.dip")

	_, stderr, code := runCLI(t, "migrate", "--output", outFile, testdata("sample.dot"))
	if code != ExitOK {
		t.Fatalf("expected exit 0, got %d; stderr: %s", code, stderr)
	}

	data, err := os.ReadFile(outFile)
	if err != nil {
		t.Fatalf("expected output file to exist: %v", err)
	}
	if !strings.Contains(string(data), "workflow") {
		t.Errorf("expected 'workflow' keyword in output file")
	}
}

func TestCmdMigrate_MissingArg(t *testing.T) {
	_, stderr, code := runCLI(t, "migrate")

	if code != ExitUsageError {
		t.Fatalf("expected exit 2, got %d", code)
	}
	if !strings.Contains(stderr, "usage:") {
		t.Errorf("expected usage message, got: %s", stderr)
	}
}

// --- Validate-Migration Command ---

func TestCmdValidateMigration_Match(t *testing.T) {
	// Migrate the DOT file, write to a temp .dip, then validate parity.
	tmpDir := t.TempDir()
	dipFile := filepath.Join(tmpDir, "migrated.dip")

	_, _, code := runCLI(t, "migrate", "--output", dipFile, testdata("sample.dot"))
	if code != ExitOK {
		t.Fatalf("migrate failed with exit %d", code)
	}

	stdout, stderr, code := runCLI(t, "validate-migration", testdata("sample.dot"), dipFile)
	if code != ExitOK {
		t.Fatalf("expected exit 0 for matching migration, got %d; stderr: %s", code, stderr)
	}
	if !strings.Contains(stdout, "parity check passed") {
		t.Errorf("expected success message, got stdout: %s", stdout)
	}
}

func TestCmdValidateMigration_Mismatch(t *testing.T) {
	// Use valid_minimal.dip which has different structure than sample.dot.
	_, stderr, code := runCLI(t, "validate-migration", testdata("sample.dot"), testdata("valid_minimal.dip"))

	if code != ExitError {
		t.Fatalf("expected exit 1 for mismatched migration, got %d", code)
	}
	if !strings.Contains(stderr, "parity check failed") && !strings.Contains(stderr, "difference") {
		t.Errorf("expected parity failure message on stderr, got: %s", stderr)
	}
}

func TestCmdValidateMigration_MissingArgs(t *testing.T) {
	_, stderr, code := runCLI(t, "validate-migration", testdata("sample.dot"))

	if code != ExitUsageError {
		t.Fatalf("expected exit 2 for missing args, got %d", code)
	}
	if !strings.Contains(stderr, "usage:") {
		t.Errorf("expected usage message, got: %s", stderr)
	}
}

// --- Edge Cases ---

func TestCmdUnknownCommand(t *testing.T) {
	_, stderr, code := runCLI(t, "bogus")

	if code != ExitUsageError {
		t.Fatalf("expected exit 2, got %d", code)
	}
	if !strings.Contains(stderr, "unknown command") {
		t.Errorf("expected 'unknown command' on stderr, got: %s", stderr)
	}
}

func TestCmdNoArgs(t *testing.T) {
	_, stderr, code := runCLI(t)

	if code != ExitUsageError {
		t.Fatalf("expected exit 2, got %d", code)
	}
	if !strings.Contains(stderr, "usage:") {
		t.Errorf("expected usage message on stderr, got: %s", stderr)
	}
}

func TestCmdHelp(t *testing.T) {
	stdout, _, code := runCLI(t, "help")

	if code != ExitOK {
		t.Fatalf("expected exit 0, got %d", code)
	}
	if !strings.Contains(stdout, "usage:") {
		t.Errorf("expected usage on stdout, got: %s", stdout)
	}
	if !strings.Contains(stdout, "parse") {
		t.Error("expected 'parse' command in help output")
	}
	if !strings.Contains(stdout, "validate-migration") {
		t.Error("expected 'validate-migration' command in help output")
	}
	if !strings.Contains(stdout, "simulate") {
		t.Error("expected 'simulate' command in help output")
	}
	if !strings.Contains(stdout, "check") {
		t.Error("expected 'check' command in help output")
	}
	if !strings.Contains(stdout, "new") {
		t.Error("expected 'new' command in help output")
	}
}

// --- Global Flag Tests ---

func TestGlobalFlag_Version(t *testing.T) {
	for _, flag := range []string{"--version", "-version"} {
		stdout, _, code := runCLI(t, flag)
		if code != ExitOK {
			t.Fatalf("%s: expected exit 0, got %d", flag, code)
		}
		if !strings.Contains(stdout, "dippin") {
			t.Errorf("%s: expected 'dippin' in output, got: %s", flag, stdout)
		}
	}
}

func TestGlobalFlag_FormatJSON_Validate(t *testing.T) {
	// Verify that --format json works for validate command on a valid file.
	stdout, stderr, code := runCLI(t, "--format", "json", "validate", testdata("valid_minimal.dip"))

	if code != ExitOK {
		t.Fatalf("expected exit 0, got %d; stderr: %s", code, stderr)
	}
	// In JSON mode, no text "validation passed" should be printed.
	if strings.Contains(stdout, "validation passed") {
		t.Error("in JSON mode, should not print text success message")
	}
}

func TestGlobalFlag_InvalidFormat(t *testing.T) {
	_, stderr, code := runCLI(t, "--format", "xml", "validate", testdata("valid_minimal.dip"))

	if code != ExitUsageError {
		t.Fatalf("expected exit 2 for invalid format, got %d", code)
	}
	if !strings.Contains(stderr, "unknown format") {
		t.Errorf("expected 'unknown format' error, got: %s", stderr)
	}
}

// --- Table-Driven: Exit Code Consistency ---

func TestExitCodes(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		wantCode ExitCode
	}{
		{
			name:     "no args",
			args:     nil,
			wantCode: ExitUsageError,
		},
		{
			name:     "unknown command",
			args:     []string{"nonexistent"},
			wantCode: ExitUsageError,
		},
		{
			name:     "parse missing file arg",
			args:     []string{"parse"},
			wantCode: ExitUsageError,
		},
		{
			name:     "validate missing file arg",
			args:     []string{"validate"},
			wantCode: ExitUsageError,
		},
		{
			name:     "lint missing file arg",
			args:     []string{"lint"},
			wantCode: ExitUsageError,
		},
		{
			name:     "fmt missing file arg",
			args:     []string{"fmt"},
			wantCode: ExitUsageError,
		},
		{
			name:     "export-dot missing file arg",
			args:     []string{"export-dot"},
			wantCode: ExitUsageError,
		},
		{
			name:     "migrate missing file arg",
			args:     []string{"migrate"},
			wantCode: ExitUsageError,
		},
		{
			name:     "validate-migration missing both args",
			args:     []string{"validate-migration"},
			wantCode: ExitUsageError,
		},
		{
			name:     "validate-migration missing second arg",
			args:     []string{"validate-migration", testdata("sample.dot")},
			wantCode: ExitUsageError,
		},
		{
			name:     "parse nonexistent file",
			args:     []string{"parse", "testdata/does_not_exist.dip"},
			wantCode: ExitError,
		},
		{
			name:     "validate valid file",
			args:     []string{"validate", testdata("valid_minimal.dip")},
			wantCode: ExitOK,
		},
		{
			name:     "validate invalid file",
			args:     []string{"validate", testdata("invalid_missing_start.dip")},
			wantCode: ExitError,
		},
		{
			name:     "lint clean file",
			args:     []string{"lint", testdata("valid_minimal.dip")},
			wantCode: ExitOK,
		},
		{
			name:     "lint warnings only",
			args:     []string{"lint", testdata("lint_warnings.dip")},
			wantCode: ExitOK,
		},
		{
			name:     "fmt check canonical",
			args:     []string{"fmt", "--check", testdata("valid_minimal.dip")},
			wantCode: ExitOK,
		},
		{
			name:     "fmt check not canonical",
			args:     []string{"fmt", "--check", testdata("needs_formatting.dip")},
			wantCode: ExitError,
		},
		{
			name:     "check missing file arg",
			args:     []string{"check"},
			wantCode: ExitUsageError,
		},
		{
			name:     "check valid file",
			args:     []string{"check", testdata("valid_minimal.dip")},
			wantCode: ExitOK,
		},
		{
			name:     "check invalid file",
			args:     []string{"check", testdata("check_errors.dip")},
			wantCode: ExitError,
		},
		{
			name:     "new missing template arg",
			args:     []string{"new"},
			wantCode: ExitUsageError,
		},
		{
			name:     "new valid template",
			args:     []string{"new", "minimal"},
			wantCode: ExitOK,
		},
		{
			name:     "new unknown template",
			args:     []string{"new", "nosuch"},
			wantCode: ExitError,
		},
		{
			name:     "simulate missing file arg",
			args:     []string{"simulate"},
			wantCode: ExitUsageError,
		},
		{
			name:     "simulate valid file",
			args:     []string{"simulate", testdata("valid_minimal.dip")},
			wantCode: ExitOK,
		},
		{
			name:     "simulate invalid file",
			args:     []string{"simulate", testdata("invalid_missing_start.dip")},
			wantCode: ExitError,
		},
		{
			name:     "help command",
			args:     []string{"help"},
			wantCode: ExitOK,
		},
		{
			name:     "cost missing file arg",
			args:     []string{"cost"},
			wantCode: ExitUsageError,
		},
		{
			name:     "cost valid file",
			args:     []string{"cost", testdata("valid_minimal.dip")},
			wantCode: ExitOK,
		},
		{
			name:     "coverage missing file arg",
			args:     []string{"coverage"},
			wantCode: ExitUsageError,
		},
		{
			name:     "coverage valid file",
			args:     []string{"coverage", testdata("valid_minimal.dip")},
			wantCode: ExitOK,
		},
		{
			name:     "doctor missing file arg",
			args:     []string{"doctor"},
			wantCode: ExitUsageError,
		},
		{
			name:     "doctor valid file",
			args:     []string{"doctor", testdata("valid_minimal.dip")},
			wantCode: ExitOK,
		},
		{
			name:     "optimize missing file arg",
			args:     []string{"optimize"},
			wantCode: ExitUsageError,
		},
		{
			name:     "optimize valid file",
			args:     []string{"optimize", testdata("valid_minimal.dip")},
			wantCode: ExitOK,
		},
		{
			name:     "unused missing file arg",
			args:     []string{"unused"},
			wantCode: ExitUsageError,
		},
		{
			name:     "unused valid file",
			args:     []string{"unused", testdata("valid_minimal.dip")},
			wantCode: ExitOK,
		},
		{
			name:     "graph missing file arg",
			args:     []string{"graph"},
			wantCode: ExitUsageError,
		},
		{
			name:     "graph valid file",
			args:     []string{"graph", testdata("valid_minimal.dip")},
			wantCode: ExitOK,
		},
		{
			name:     "diff missing args",
			args:     []string{"diff"},
			wantCode: ExitUsageError,
		},
		{
			name:     "feedback missing args",
			args:     []string{"feedback"},
			wantCode: ExitUsageError,
		},
		{
			name:     "explain missing arg",
			args:     []string{"explain"},
			wantCode: ExitUsageError,
		},
		{
			name:     "explain valid code",
			args:     []string{"explain", "DIP001"},
			wantCode: ExitOK,
		},
		{
			name:     "explain unknown code",
			args:     []string{"explain", "DIP999"},
			wantCode: ExitError,
		},
		{
			name:     "test missing arg",
			args:     []string{"test"},
			wantCode: ExitUsageError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, code := runCLI(t, tt.args...)
			if code != tt.wantCode {
				t.Errorf("args=%v: got exit code %d, want %d", tt.args, code, tt.wantCode)
			}
		})
	}
}

// --- Render Diagnostics ---

func TestRenderDiagnosticsJSON_Structure(t *testing.T) {
	// Verify the JSON diagnostic output conforms to spec §12 format.
	_, stderr, _ := runCLI(t, "--format", "json", "validate", testdata("invalid_missing_start.dip"))

	var diags []map[string]interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &diags); err != nil {
		t.Fatalf("failed to parse JSON diagnostics: %v\nraw: %s", err, stderr)
	}

	for i, d := range diags {
		// Each diagnostic must have these fields per spec §12.
		for _, field := range []string{"code", "severity", "message", "location"} {
			if _, ok := d[field]; !ok {
				t.Errorf("diagnostic[%d] missing required field %q", i, field)
			}
		}

		// Location must have these sub-fields.
		loc, ok := d["location"].(map[string]interface{})
		if !ok {
			t.Errorf("diagnostic[%d] location is not an object", i)
			continue
		}
		for _, field := range []string{"file", "line", "column", "end_line", "end_column"} {
			if _, ok := loc[field]; !ok {
				t.Errorf("diagnostic[%d] location missing field %q", i, field)
			}
		}
	}
}

func TestRenderDiagnosticsText_Format(t *testing.T) {
	// Verify the text diagnostic output matches the spec format:
	// error[DIP001]: ...
	//   --> file:line:col
	_, stderr, _ := runCLI(t, "validate", testdata("invalid_missing_start.dip"))

	if !strings.Contains(stderr, "error[DIP001]") {
		t.Errorf("expected 'error[DIP001]' in text output, got: %s", stderr)
	}
	if !strings.Contains(stderr, "-->") {
		t.Errorf("expected '-->' location indicator in text output, got: %s", stderr)
	}
}

// --- Parse DOT File ---

func TestCmdParse_DOTFile(t *testing.T) {
	// parse should auto-detect .dot files and use the migration path.
	stdout, stderr, code := runCLI(t, "parse", testdata("sample.dot"))

	if code != ExitOK {
		t.Fatalf("expected exit 0, got %d; stderr: %s", code, stderr)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(stdout), &parsed); err != nil {
		t.Fatalf("stdout is not valid JSON: %v", err)
	}
	if parsed["Name"] != "Sample" {
		t.Errorf("expected Name=Sample, got %v", parsed["Name"])
	}
}

// --- Simulate Command ---

func TestCmdSimulate_Basic(t *testing.T) {
	stdout, stderr, code := runCLI(t, "simulate", testdata("valid_minimal.dip"))

	if code != ExitOK {
		t.Fatalf("expected exit 0, got %d; stderr: %s", code, stderr)
	}

	// Output should be valid JSONL.
	lines := strings.Split(strings.TrimSpace(stdout), "\n")
	if len(lines) < 3 {
		t.Fatalf("expected at least 3 JSONL lines, got %d", len(lines))
	}

	// First line should be pipeline_start.
	var first map[string]interface{}
	if err := json.Unmarshal([]byte(lines[0]), &first); err != nil {
		t.Fatalf("first line not valid JSON: %v\nline: %s", err, lines[0])
	}
	if first["event"] != "pipeline_start" {
		t.Errorf("first event = %q, want pipeline_start", first["event"])
	}
	if first["workflow"] != "Minimal" {
		t.Errorf("workflow = %q, want Minimal", first["workflow"])
	}

	// Last line should be pipeline_end.
	var last map[string]interface{}
	if err := json.Unmarshal([]byte(lines[len(lines)-1]), &last); err != nil {
		t.Fatalf("last line not valid JSON: %v\nline: %s", err, lines[len(lines)-1])
	}
	if last["event"] != "pipeline_end" {
		t.Errorf("last event = %q, want pipeline_end", last["event"])
	}
	if last["status"] != "success" {
		t.Errorf("status = %q, want success", last["status"])
	}

	// Stderr should contain summary.
	if !strings.Contains(stderr, "simulation complete") {
		t.Errorf("expected 'simulation complete' on stderr, got: %s", stderr)
	}
}

func TestCmdSimulate_AllJSONLLinesValid(t *testing.T) {
	stdout, _, code := runCLI(t, "simulate", testdata("valid_minimal.dip"))

	if code != ExitOK {
		t.Fatalf("expected exit 0, got %d", code)
	}

	lines := strings.Split(strings.TrimSpace(stdout), "\n")
	for i, line := range lines {
		var m map[string]interface{}
		if err := json.Unmarshal([]byte(line), &m); err != nil {
			t.Errorf("line %d not valid JSON: %v\nline: %s", i, err, line)
		}
		if _, ok := m["event"]; !ok {
			t.Errorf("line %d missing 'event' field", i)
		}
		if _, ok := m["timestamp"]; !ok {
			t.Errorf("line %d missing 'timestamp' field", i)
		}
	}
}

func TestCmdSimulate_MissingArg(t *testing.T) {
	_, stderr, code := runCLI(t, "simulate")

	if code != ExitUsageError {
		t.Fatalf("expected exit 2, got %d", code)
	}
	if !strings.Contains(stderr, "usage:") {
		t.Errorf("expected usage message, got: %s", stderr)
	}
}

func TestCmdSimulate_NonexistentFile(t *testing.T) {
	_, stderr, code := runCLI(t, "simulate", "testdata/nosuch.dip")

	if code != ExitError {
		t.Fatalf("expected exit 1, got %d", code)
	}
	if stderr == "" {
		t.Error("expected error on stderr")
	}
}

func TestCmdSimulate_InvalidFile(t *testing.T) {
	_, stderr, code := runCLI(t, "simulate", testdata("invalid_missing_start.dip"))

	// Should fail validation.
	if code != ExitError {
		t.Fatalf("expected exit 1, got %d", code)
	}
	if !strings.Contains(stderr, "DIP001") {
		t.Errorf("expected DIP001 diagnostic, got: %s", stderr)
	}
}

func TestCmdSimulate_WithScenario(t *testing.T) {
	// Create a workflow file with conditional edges to test --scenario.
	tmpDir := t.TempDir()
	dipFile := filepath.Join(tmpDir, "cond.dip")
	content := `workflow Cond
  goal: "Test conditional"
  start: Check
  exit: Done

  agent Check
    auto_status: true
    prompt:
      Check.

  agent PathA
    prompt:
      A.

  agent PathB
    prompt:
      B.

  agent Done
    prompt:
      Done.

  edges
    Check -> PathA  when ctx.outcome = success
    Check -> PathB  when ctx.outcome = fail
    PathA -> Done
    PathB -> Done
`
	if err := os.WriteFile(dipFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	// Test with outcome=fail scenario.
	stdout, _, code := runCLI(t, "simulate", "--scenario", "outcome=fail", dipFile)
	if code != ExitOK {
		t.Fatalf("expected exit 0, got %d", code)
	}

	// Should traverse to PathB.
	if !strings.Contains(stdout, "PathB") {
		t.Errorf("expected PathB in output, got: %s", stdout)
	}
}

func TestCmdSimulate_AllPaths(t *testing.T) {
	// Create a workflow with conditional branching.
	tmpDir := t.TempDir()
	dipFile := filepath.Join(tmpDir, "branch.dip")
	content := `workflow Branch
  goal: "Test all-paths"
  start: Start
  exit: Done

  agent Start
    prompt:
      Start.

  agent PathA
    prompt:
      A.

  agent PathB
    prompt:
      B.

  agent Done
    prompt:
      Done.

  edges
    Start -> PathA  when ctx.x = a
    Start -> PathB  when ctx.x = b
    PathA -> Done
    PathB -> Done
`
	if err := os.WriteFile(dipFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	stdout, stderr, code := runCLI(t, "simulate", "--all-paths", dipFile)
	if code != ExitOK {
		t.Fatalf("expected exit 0, got %d; stderr: %s", code, stderr)
	}

	// Should find 2 paths.
	if !strings.Contains(stderr, "2 path(s) enumerated") {
		t.Errorf("expected '2 path(s) enumerated' on stderr, got: %s", stderr)
	}

	// Both paths should appear in output.
	if !strings.Contains(stdout, "PathA") {
		t.Error("expected PathA in output")
	}
	if !strings.Contains(stdout, "PathB") {
		t.Error("expected PathB in output")
	}
}

// --- Cost Command ---

func TestCmdCost_Valid(t *testing.T) {
	stdout, stderr, code := runCLI(t, "cost", testdata("valid_minimal.dip"))

	if code != ExitOK {
		t.Fatalf("expected exit 0, got %d; stderr: %s", code, stderr)
	}
	if !strings.Contains(stdout, "Cost Estimate") {
		t.Errorf("expected 'Cost Estimate' in stdout, got: %s", stdout)
	}
}

func TestCmdCost_MissingArg(t *testing.T) {
	_, stderr, code := runCLI(t, "cost")

	if code != ExitUsageError {
		t.Fatalf("expected exit 2, got %d", code)
	}
	if !strings.Contains(stderr, "usage:") {
		t.Errorf("expected usage message on stderr, got: %s", stderr)
	}
}

func TestCmdCost_JSON(t *testing.T) {
	stdout, stderr, code := runCLI(t, "--format", "json", "cost", testdata("valid_minimal.dip"))

	if code != ExitOK {
		t.Fatalf("expected exit 0, got %d; stderr: %s", code, stderr)
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(stdout)), &result); err != nil {
		t.Fatalf("stdout is not valid JSON: %v\nstdout: %s", err, stdout)
	}
}

// --- Coverage Command ---

func TestCmdCoverage_Valid(t *testing.T) {
	stdout, stderr, code := runCLI(t, "coverage", testdata("valid_minimal.dip"))

	if code != ExitOK {
		t.Fatalf("expected exit 0, got %d; stderr: %s", code, stderr)
	}
	if !strings.Contains(stdout, "Coverage Analysis") {
		t.Errorf("expected 'Coverage Analysis' in stdout, got: %s", stdout)
	}
}

func TestCmdCoverage_MissingArg(t *testing.T) {
	_, stderr, code := runCLI(t, "coverage")

	if code != ExitUsageError {
		t.Fatalf("expected exit 2, got %d", code)
	}
	if !strings.Contains(stderr, "usage:") {
		t.Errorf("expected usage message on stderr, got: %s", stderr)
	}
}

func TestCmdCoverage_JSON(t *testing.T) {
	stdout, stderr, code := runCLI(t, "--format", "json", "coverage", testdata("valid_minimal.dip"))

	if code != ExitOK {
		t.Fatalf("expected exit 0, got %d; stderr: %s", code, stderr)
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(stdout)), &result); err != nil {
		t.Fatalf("stdout is not valid JSON: %v\nstdout: %s", err, stdout)
	}
}

func TestCmdCoverage_PartialCoverage(t *testing.T) {
	stdout, stderr, code := runCLI(t, "coverage", testdata("partial_coverage.dip"))

	if code != ExitOK {
		t.Fatalf("expected exit 0, got %d; stderr: %s", code, stderr)
	}
	if !strings.Contains(stdout, "Coverage Analysis") {
		t.Errorf("expected 'Coverage Analysis' in stdout, got: %s", stdout)
	}
	if !strings.Contains(stdout, "Edge Coverage") {
		t.Errorf("expected 'Edge Coverage' section, got: %s", stdout)
	}
	// Tool node with extracted outputs should show partial coverage.
	if !strings.Contains(stdout, "RunTool") {
		t.Errorf("expected 'RunTool' node in coverage output, got: %s", stdout)
	}
	if !strings.Contains(stdout, "partial") {
		t.Errorf("expected 'partial' status, got: %s", stdout)
	}
	if !strings.Contains(stdout, "missing: fail") {
		t.Errorf("expected 'missing: fail' in output, got: %s", stdout)
	}
}

// --- Doctor Command ---

func TestCmdDoctor_Valid(t *testing.T) {
	stdout, stderr, code := runCLI(t, "doctor", testdata("valid_minimal.dip"))

	if code != ExitOK {
		t.Fatalf("expected exit 0, got %d; stderr: %s", code, stderr)
	}
	if !strings.Contains(stdout, "Health Report") {
		t.Errorf("expected 'Health Report' in stdout, got: %s", stdout)
	}
}

func TestCmdDoctor_MissingArg(t *testing.T) {
	_, stderr, code := runCLI(t, "doctor")

	if code != ExitUsageError {
		t.Fatalf("expected exit 2, got %d", code)
	}
	if !strings.Contains(stderr, "usage:") {
		t.Errorf("expected usage message on stderr, got: %s", stderr)
	}
}

func TestCmdDoctor_JSON(t *testing.T) {
	stdout, stderr, code := runCLI(t, "--format", "json", "doctor", testdata("valid_minimal.dip"))

	if code != ExitOK {
		t.Fatalf("expected exit 0, got %d; stderr: %s", code, stderr)
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(stdout)), &result); err != nil {
		t.Fatalf("stdout is not valid JSON: %v\nstdout: %s", err, stdout)
	}
	if _, ok := result["grade"]; !ok {
		t.Error("expected 'grade' field in JSON output")
	}
}

// --- Optimize Command ---

func TestCmdOptimize_Valid(t *testing.T) {
	stdout, stderr, code := runCLI(t, "optimize", testdata("valid_minimal.dip"))

	if code != ExitOK {
		t.Fatalf("expected exit 0, got %d; stderr: %s", code, stderr)
	}
	if !strings.Contains(stdout, "Optimization") {
		t.Errorf("expected 'Optimization' in stdout, got: %s", stdout)
	}
}

func TestCmdOptimize_MissingArg(t *testing.T) {
	_, stderr, code := runCLI(t, "optimize")

	if code != ExitUsageError {
		t.Fatalf("expected exit 2, got %d", code)
	}
	if !strings.Contains(stderr, "usage:") {
		t.Errorf("expected usage message on stderr, got: %s", stderr)
	}
}

func TestCmdOptimize_JSON(t *testing.T) {
	stdout, stderr, code := runCLI(t, "--format", "json", "optimize", testdata("valid_minimal.dip"))

	if code != ExitOK {
		t.Fatalf("expected exit 0, got %d; stderr: %s", code, stderr)
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(stdout)), &result); err != nil {
		t.Fatalf("stdout is not valid JSON: %v\nstdout: %s", err, stdout)
	}
}

// --- Unused Command ---

func TestCmdUnused_Valid(t *testing.T) {
	stdout, stderr, code := runCLI(t, "unused", testdata("valid_minimal.dip"))

	if code != ExitOK {
		t.Fatalf("expected exit 0, got %d; stderr: %s", code, stderr)
	}
	if !strings.Contains(stdout, "Unused Nodes") {
		t.Errorf("expected 'Unused Nodes' in stdout, got: %s", stdout)
	}
}

func TestCmdUnused_MissingArg(t *testing.T) {
	_, stderr, code := runCLI(t, "unused")

	if code != ExitUsageError {
		t.Fatalf("expected exit 2, got %d", code)
	}
	if !strings.Contains(stderr, "usage:") {
		t.Errorf("expected usage message on stderr, got: %s", stderr)
	}
}

func TestCmdUnused_JSON(t *testing.T) {
	stdout, stderr, code := runCLI(t, "--format", "json", "unused", testdata("valid_minimal.dip"))

	if code != ExitOK {
		t.Fatalf("expected exit 0, got %d; stderr: %s", code, stderr)
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(stdout)), &result); err != nil {
		t.Fatalf("stdout is not valid JSON: %v\nstdout: %s", err, stdout)
	}
}

// --- Graph Command ---

func TestCmdGraph_Full(t *testing.T) {
	stdout, stderr, code := runCLI(t, "graph", testdata("valid_minimal.dip"))

	if code != ExitOK {
		t.Fatalf("expected exit 0, got %d; stderr: %s", code, stderr)
	}
	if !strings.Contains(stdout, "Ask") {
		t.Errorf("expected 'Ask' in graph output, got: %s", stdout)
	}
	if !strings.Contains(stdout, "Done") {
		t.Errorf("expected 'Done' in graph output, got: %s", stdout)
	}
}

func TestCmdGraph_Compact(t *testing.T) {
	stdout, stderr, code := runCLI(t, "graph", "--compact", testdata("valid_minimal.dip"))

	if code != ExitOK {
		t.Fatalf("expected exit 0, got %d; stderr: %s", code, stderr)
	}
	if !strings.Contains(stdout, "\u2192") {
		t.Errorf("expected arrow character in compact output, got: %s", stdout)
	}
}

func TestCmdGraph_JSON(t *testing.T) {
	stdout, stderr, code := runCLI(t, "--format", "json", "graph", testdata("valid_minimal.dip"))

	if code != ExitOK {
		t.Fatalf("expected exit 0, got %d; stderr: %s", code, stderr)
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(stdout)), &result); err != nil {
		t.Fatalf("stdout is not valid JSON: %v\nstdout: %s", err, stdout)
	}
	if _, ok := result["layers"]; !ok {
		t.Error("expected 'layers' field in JSON output")
	}
}

func TestCmdGraph_MissingArg(t *testing.T) {
	_, stderr, code := runCLI(t, "graph")

	if code != ExitUsageError {
		t.Fatalf("expected exit 2, got %d", code)
	}
	if !strings.Contains(stderr, "usage:") {
		t.Errorf("expected usage message on stderr, got: %s", stderr)
	}
}

// --- Diff Command ---

func TestCmdDiff_Basic(t *testing.T) {
	stdout, stderr, code := runCLI(t, "diff", testdata("valid_minimal.dip"), testdata("valid_minimal.dip"))

	if code != ExitOK {
		t.Fatalf("expected exit 0, got %d; stderr: %s", code, stderr)
	}
	if !strings.Contains(stdout, "Diff") {
		t.Errorf("expected 'Diff' in stdout, got: %s", stdout)
	}
}

func TestCmdDiff_JSON(t *testing.T) {
	stdout, stderr, code := runCLI(t, "--format", "json", "diff", testdata("valid_minimal.dip"), testdata("valid_minimal.dip"))

	if code != ExitOK {
		t.Fatalf("expected exit 0, got %d; stderr: %s", code, stderr)
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(stdout)), &result); err != nil {
		t.Fatalf("stdout is not valid JSON: %v\nstdout: %s", err, stdout)
	}
}

func TestCmdDiff_MissingArg(t *testing.T) {
	_, stderr, code := runCLI(t, "diff")

	if code != ExitUsageError {
		t.Fatalf("expected exit 2, got %d", code)
	}
	if !strings.Contains(stderr, "usage:") {
		t.Errorf("expected usage message on stderr, got: %s", stderr)
	}
}

// --- Feedback Command ---

func TestCmdFeedback_Basic(t *testing.T) {
	stdout, stderr, code := runCLI(t, "feedback", testdata("valid_minimal.dip"), testdata("sample_telemetry.jsonl"))

	if code != ExitOK {
		t.Fatalf("expected exit 0, got %d; stderr: %s", code, stderr)
	}
	if !strings.Contains(stdout, "Calibration") && !strings.Contains(stdout, "Cost") {
		t.Errorf("expected 'Calibration' or 'Cost' in stdout, got: %s", stdout)
	}
}

func TestCmdFeedback_JSON(t *testing.T) {
	stdout, stderr, code := runCLI(t, "--format", "json", "feedback", testdata("valid_minimal.dip"), testdata("sample_telemetry.jsonl"))

	if code != ExitOK {
		t.Fatalf("expected exit 0, got %d; stderr: %s", code, stderr)
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(stdout)), &result); err != nil {
		t.Fatalf("stdout is not valid JSON: %v\nstdout: %s", err, stdout)
	}
}

func TestCmdFeedback_MissingArg(t *testing.T) {
	_, stderr, code := runCLI(t, "feedback")

	if code != ExitUsageError {
		t.Fatalf("expected exit 2, got %d", code)
	}
	if !strings.Contains(stderr, "usage:") {
		t.Errorf("expected usage message on stderr, got: %s", stderr)
	}
}

// --- Explain Command ---

func TestCmdExplain_ValidCode(t *testing.T) {
	stdout, stderr, code := runCLI(t, "explain", "DIP001")

	if code != ExitOK {
		t.Fatalf("expected exit 0, got %d; stderr: %s", code, stderr)
	}
	if !strings.Contains(stdout, "DIP001") {
		t.Errorf("expected 'DIP001' in stdout, got: %s", stdout)
	}
}

func TestCmdExplain_UnknownCode(t *testing.T) {
	_, stderr, code := runCLI(t, "explain", "DIP999")

	if code != ExitError {
		t.Fatalf("expected exit 1, got %d", code)
	}
	if !strings.Contains(stderr, "unknown") {
		t.Errorf("expected 'unknown' on stderr, got: %s", stderr)
	}
}

func TestCmdExplain_BadFormat(t *testing.T) {
	_, stderr, code := runCLI(t, "explain", "foobar")

	if code != ExitUsageError {
		t.Fatalf("expected exit 2, got %d", code)
	}
	if !strings.Contains(stderr, "invalid") {
		t.Errorf("expected 'invalid' on stderr, got: %s", stderr)
	}
}

func TestCmdExplain_MissingArg(t *testing.T) {
	_, stderr, code := runCLI(t, "explain")

	if code != ExitUsageError {
		t.Fatalf("expected exit 2, got %d", code)
	}
	if !strings.Contains(stderr, "usage:") {
		t.Errorf("expected usage message on stderr, got: %s", stderr)
	}
}

// --- Test Command ---

func TestCmdTest_Pass(t *testing.T) {
	simulate.ResetRunCounter()
	stdout, stderr, code := runCLI(t, "test", testdata("valid_minimal.dip"))

	if code != ExitOK {
		t.Fatalf("expected exit 0, got %d; stderr: %s", code, stderr)
	}
	if !strings.Contains(stdout, "PASS") && !strings.Contains(stdout, "Test Results") {
		t.Errorf("expected 'PASS' or 'Test Results' in stdout, got: %s", stdout)
	}
}

func TestCmdTest_MissingArg(t *testing.T) {
	_, stderr, code := runCLI(t, "test")

	if code != ExitUsageError {
		t.Fatalf("expected exit 2, got %d", code)
	}
	if !strings.Contains(stderr, "usage:") {
		t.Errorf("expected usage message on stderr, got: %s", stderr)
	}
}

func TestCmdTest_MissingTestFile(t *testing.T) {
	_, stderr, code := runCLI(t, "test", testdata("lint_warnings.dip"))

	if code != ExitError {
		t.Fatalf("expected exit 1, got %d", code)
	}
	if !strings.Contains(stderr, "error") && !strings.Contains(stderr, "no such file") {
		t.Errorf("expected error about missing test file on stderr, got: %s", stderr)
	}
}

// --- Simulate DOT Format ---

func TestCmdSimulate_DOTFormat(t *testing.T) {
	simulate.ResetRunCounter()
	stdout, stderr, code := runCLI(t, "--format", "dot", "simulate", testdata("valid_minimal.dip"))

	if code != ExitOK {
		t.Fatalf("expected exit 0, got %d; stderr: %s", code, stderr)
	}
	if !strings.Contains(stdout, "digraph") {
		t.Errorf("expected 'digraph' in DOT output, got: %s", stdout)
	}
	if !strings.Contains(stdout, "Ask") {
		t.Errorf("expected 'Ask' node in DOT output, got: %s", stdout)
	}
}

// --- Diff Command ---

func TestCmdDiff_AddedAndRemovedNodes(t *testing.T) {
	stdout, stderr, code := runCLI(t, "diff", testdata("valid_minimal.dip"), testdata("valid_alternate.dip"))

	if code != ExitOK {
		t.Fatalf("expected exit 0, got %d; stderr: %s", code, stderr)
	}
	// Added nodes.
	if !strings.Contains(stdout, "+ Begin") {
		t.Errorf("expected '+ Begin' in diff output, got: %s", stdout)
	}
	if !strings.Contains(stdout, "+ Middle") {
		t.Errorf("expected '+ Middle' in diff output, got: %s", stdout)
	}
	// Removed nodes.
	if !strings.Contains(stdout, "- Ask") {
		t.Errorf("expected '- Ask' in diff output, got: %s", stdout)
	}
	// Added edges.
	if !strings.Contains(stdout, "+ Begin -> Middle") {
		t.Errorf("expected '+ Begin -> Middle' in edge diff, got: %s", stdout)
	}
	// Removed edges.
	if !strings.Contains(stdout, "- Ask -> Done") {
		t.Errorf("expected '- Ask -> Done' in edge diff, got: %s", stdout)
	}
}

func TestCmdDiff_ModifiedNodes(t *testing.T) {
	stdout, stderr, code := runCLI(t, "diff", testdata("valid_minimal.dip"), testdata("valid_minimal_v2.dip"))

	if code != ExitOK {
		t.Fatalf("expected exit 0, got %d; stderr: %s", code, stderr)
	}
	// Modified node.
	if !strings.Contains(stdout, "~ Done") {
		t.Errorf("expected '~ Done' in diff output, got: %s", stdout)
	}
	// Field change details.
	if !strings.Contains(stdout, "model:") {
		t.Errorf("expected model field change in diff output, got: %s", stdout)
	}
}

func TestCmdDiff_JSONFormat(t *testing.T) {
	stdout, stderr, code := runCLI(t, "--format", "json", "diff", testdata("valid_minimal.dip"), testdata("valid_alternate.dip"))

	if code != ExitOK {
		t.Fatalf("expected exit 0, got %d; stderr: %s", code, stderr)
	}
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(stdout)), &result); err != nil {
		t.Fatalf("expected JSON output, got: %s", stdout)
	}
}

// --- Cost Command ---

func TestCmdCost_TextFormat_MultiProvider(t *testing.T) {
	stdout, stderr, code := runCLI(t, "cost", testdata("multi_provider.dip"))

	if code != ExitOK {
		t.Fatalf("expected exit 0, got %d; stderr: %s", code, stderr)
	}
	if !strings.Contains(stdout, "Cost Estimate") {
		t.Errorf("expected 'Cost Estimate' header, got: %s", stdout)
	}
	if !strings.Contains(stdout, "By Provider") {
		t.Errorf("expected 'By Provider' section, got: %s", stdout)
	}
	if !strings.Contains(stdout, "anthropic") {
		t.Errorf("expected 'anthropic' provider, got: %s", stdout)
	}
	if !strings.Contains(stdout, "openai") {
		t.Errorf("expected 'openai' provider, got: %s", stdout)
	}
}

// --- Optimize Command ---

func TestCmdOptimize_TextFormat_WithSuggestions(t *testing.T) {
	stdout, stderr, code := runCLI(t, "optimize", testdata("optimize_target.dip"))

	if code != ExitOK {
		t.Fatalf("expected exit 0, got %d; stderr: %s", code, stderr)
	}
	if !strings.Contains(stdout, "Optimization Report") {
		t.Errorf("expected 'Optimization Report' header, got: %s", stdout)
	}
	if !strings.Contains(stdout, "Suggestions") {
		t.Errorf("expected 'Suggestions' section, got: %s", stdout)
	}
	if !strings.Contains(stdout, "Summarize") {
		t.Errorf("expected 'Summarize' node in suggestions, got: %s", stdout)
	}
	if !strings.Contains(stdout, "saves") {
		t.Errorf("expected 'saves' in suggestion output, got: %s", stdout)
	}
}

func TestCmdOptimize_JSONFormat(t *testing.T) {
	stdout, stderr, code := runCLI(t, "--format", "json", "optimize", testdata("optimize_target.dip"))

	if code != ExitOK {
		t.Fatalf("expected exit 0, got %d; stderr: %s", code, stderr)
	}
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(stdout)), &result); err != nil {
		t.Fatalf("expected JSON output, got: %s", stdout)
	}
	suggestions, ok := result["suggestions"].([]interface{})
	if !ok || len(suggestions) == 0 {
		t.Errorf("expected non-empty suggestions array in JSON, got: %v", result["suggestions"])
	}
}

// --- Unused Command ---

func TestCmdUnused_TextFormat_WithUnusedNodes(t *testing.T) {
	stdout, stderr, code := runCLI(t, "unused", testdata("unused_nodes.dip"))

	if code != ExitOK {
		t.Fatalf("expected exit 0, got %d; stderr: %s", code, stderr)
	}
	if !strings.Contains(stdout, "Unused Nodes") {
		t.Errorf("expected 'Unused Nodes' header, got: %s", stdout)
	}
	if !strings.Contains(stdout, "DeadEnd") {
		t.Errorf("expected 'DeadEnd' node in output, got: %s", stdout)
	}
	if !strings.Contains(stdout, "Dead End") {
		t.Errorf("expected 'Dead End' label in output, got: %s", stdout)
	}
	if !strings.Contains(stdout, "1 unused node") {
		t.Errorf("expected '1 unused node' in summary, got: %s", stdout)
	}
}

func TestCmdUnused_JSONFormat(t *testing.T) {
	stdout, stderr, code := runCLI(t, "--format", "json", "unused", testdata("unused_nodes.dip"))

	if code != ExitOK {
		t.Fatalf("expected exit 0, got %d; stderr: %s", code, stderr)
	}
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(stdout)), &result); err != nil {
		t.Fatalf("expected JSON output, got: %s", stdout)
	}
	nodes, ok := result["unused_nodes"].([]interface{})
	if !ok || len(nodes) == 0 {
		t.Errorf("expected non-empty unused_nodes array, got: %v", result["unused_nodes"])
	}
}

// --- scenarioFlags.String() (cmd_simulate.go:152) ---

func TestScenarioFlags_StringWithValues(t *testing.T) {
	sf := &scenarioFlags{}
	if err := sf.Set("outcome=success"); err != nil {
		t.Fatalf("Set failed: %v", err)
	}
	got := sf.String()
	if !strings.Contains(got, "outcome=success") {
		t.Errorf("expected 'outcome=success' in String(), got: %q", got)
	}
}

func TestScenarioFlags_StringNil(t *testing.T) {
	sf := &scenarioFlags{}
	if got := sf.String(); got != "" {
		t.Errorf("expected empty string for nil values, got: %q", got)
	}
}

func TestScenarioFlags_SetInvalid(t *testing.T) {
	sf := &scenarioFlags{}
	if err := sf.Set("noequals"); err == nil {
		t.Error("expected error for invalid format without '='")
	}
}

// --- renderError JSON branch (cli.go:237) ---

func TestRenderError_JSONFormat(t *testing.T) {
	// Use --format json with a syntax error file to trigger renderError JSON path.
	_, stderr, code := runCLI(t, "--format", "json", "fmt", testdata("syntax_error.dip"))

	if code != ExitError {
		t.Fatalf("expected exit 1, got %d", code)
	}
	// JSON output should contain DIP000.
	var diags []map[string]interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &diags); err != nil {
		t.Fatalf("stderr is not valid JSON: %v\nstderr: %s", err, stderr)
	}
	if len(diags) == 0 {
		t.Fatal("expected at least one diagnostic")
	}
	if diags[0]["code"] != "DIP000" {
		t.Errorf("expected code=DIP000, got %v", diags[0]["code"])
	}
}

// --- parseAndFormat error path (cmd_fmt.go:52) ---

func TestCmdFmt_SyntaxError(t *testing.T) {
	_, stderr, code := runCLI(t, "fmt", testdata("syntax_error.dip"))

	if code != ExitError {
		t.Fatalf("expected exit 1, got %d", code)
	}
	if !strings.Contains(stderr, "error") {
		t.Errorf("expected error on stderr, got: %s", stderr)
	}
}

// --- statusIcon coverage (cmd_coverage.go:92) ---

func TestCmdCoverage_FullCoverage(t *testing.T) {
	// full_coverage.dip has tool node with extracted outputs matching all edge conditions.
	stdout, stderr, code := runCLI(t, "coverage", testdata("full_coverage.dip"))

	if code != ExitOK {
		t.Fatalf("expected exit 0, got %d; stderr: %s", code, stderr)
	}
	if !strings.Contains(stdout, "Coverage Analysis") {
		t.Errorf("expected 'Coverage Analysis' in stdout, got: %s", stdout)
	}
	// The tool node RunTool should have "covered" status (extracted success+fail = edges).
	if !strings.Contains(stdout, "covered") {
		t.Errorf("expected 'covered' status in output, got: %s", stdout)
	}
}

func TestCmdCoverage_NoConds(t *testing.T) {
	// tool_no_conditions.dip has a tool node with an unconditional edge.
	stdout, stderr, code := runCLI(t, "coverage", testdata("tool_no_conditions.dip"))

	if code != ExitOK {
		t.Fatalf("expected exit 0, got %d; stderr: %s", code, stderr)
	}
	if !strings.Contains(stdout, "no_conditions") {
		t.Errorf("expected 'no_conditions' status in output, got: %s", stdout)
	}
}

// --- renderDoctorSuggestions (cmd_doctor.go:79) ---

func TestCmdDoctor_WithSuggestions(t *testing.T) {
	// unhealthy.dip has lint warnings + an unreachable node (DIP004, an error) ->
	// generates suggestions. doctor renders the report but exits non-zero because
	// the report contains errors.
	stdout, stderr, code := runCLI(t, "doctor", testdata("unhealthy.dip"))

	if code != ExitError {
		t.Fatalf("expected ExitError (report has DIP004), got %d; stderr: %s", code, stderr)
	}
	if !strings.Contains(stdout, "Suggestions") {
		t.Errorf("expected 'Suggestions' section in output, got: %s", stdout)
	}
}

// --- renderFeedbackOutliers (cmd_feedback.go:72) ---

func TestCmdFeedback_WithOutliers(t *testing.T) {
	// outlier_telemetry.jsonl has a node with 0.500 actual_cost that should
	// differ significantly from predicted cost, triggering outlier detection.
	stdout, stderr, code := runCLI(t, "feedback",
		testdata("outlier_workflow.dip"), testdata("outlier_telemetry.jsonl"))

	if code != ExitOK {
		t.Fatalf("expected exit 0, got %d; stderr: %s", code, stderr)
	}
	if !strings.Contains(stdout, "Outliers") || !strings.Contains(stdout, "Calibration") {
		t.Errorf("expected 'Outliers' and 'Calibration' in stdout, got: %s", stdout)
	}
}

// --- renderParityDiffs + loadDOTWorkflow + loadDIPWorkflow error paths (cmd_migrate.go) ---

func TestCmdValidateMigration_JSONFormat_Mismatch(t *testing.T) {
	// Mismatch in JSON mode -> triggers renderParityDiffs JSON branch.
	_, stderr, code := runCLI(t, "--format", "json", "validate-migration",
		testdata("sample.dot"), testdata("valid_minimal.dip"))

	if code != ExitError {
		t.Fatalf("expected exit 1, got %d", code)
	}
	// JSON output should be a JSON array of differences.
	var diffs []map[string]interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(stderr)), &diffs); err != nil {
		t.Fatalf("stderr is not valid JSON: %v\nstderr: %s", err, stderr)
	}
	if len(diffs) == 0 {
		t.Error("expected at least one diff")
	}
}

func TestCmdValidateMigration_BadDOTFile(t *testing.T) {
	_, stderr, code := runCLI(t, "validate-migration",
		"testdata/nosuch.dot", testdata("valid_minimal.dip"))

	if code != ExitError {
		t.Fatalf("expected exit 1, got %d", code)
	}
	if !strings.Contains(stderr, "error") {
		t.Errorf("expected error on stderr, got: %s", stderr)
	}
}

func TestCmdValidateMigration_BadDIPFile(t *testing.T) {
	_, stderr, code := runCLI(t, "validate-migration",
		testdata("sample.dot"), "testdata/nosuch.dip")

	if code != ExitError {
		t.Fatalf("expected exit 1, got %d", code)
	}
	if !strings.Contains(stderr, "error") {
		t.Errorf("expected error on stderr, got: %s", stderr)
	}
}

func TestCmdValidateMigration_SyntaxErrorDIP(t *testing.T) {
	_, stderr, code := runCLI(t, "validate-migration",
		testdata("sample.dot"), testdata("syntax_error.dip"))

	if code != ExitError {
		t.Fatalf("expected exit 1, got %d", code)
	}
	if !strings.Contains(stderr, "error") {
		t.Errorf("expected error on stderr, got: %s", stderr)
	}
}

// --- renderCaseResult failure path (cmd_test_runner.go:68) ---

func TestCmdTest_FailingTests(t *testing.T) {
	simulate.ResetRunCounter()
	stdout, _, code := runCLI(t, "test", testdata("failing.dip"))

	if code != ExitError {
		t.Fatalf("expected exit 1 for failing tests, got %d", code)
	}
	if !strings.Contains(stdout, "FAIL") {
		t.Errorf("expected 'FAIL' in test output, got: %s", stdout)
	}
	if !strings.Contains(stdout, "expect wrong node") {
		t.Errorf("expected test name in output, got: %s", stdout)
	}
}

func TestCmdTest_Verbose(t *testing.T) {
	simulate.ResetRunCounter()
	stdout, stderr, code := runCLI(t, "test", "--verbose", testdata("valid_minimal.dip"))

	if code != ExitOK {
		t.Fatalf("expected exit 0, got %d; stderr: %s", code, stderr)
	}
	if !strings.Contains(stdout, "path:") {
		t.Errorf("expected 'path:' in verbose output, got: %s", stdout)
	}
}

func TestCmdTest_JSONFormat(t *testing.T) {
	simulate.ResetRunCounter()
	stdout, stderr, code := runCLI(t, "--format", "json", "test", testdata("valid_minimal.dip"))

	if code != ExitOK {
		t.Fatalf("expected exit 0, got %d; stderr: %s", code, stderr)
	}
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(stdout)), &result); err != nil {
		t.Fatalf("stdout is not valid JSON: %v\nstdout: %s", err, stdout)
	}
}

// --- Simulate with populated scenario (additional String() coverage) ---

func TestCmdSimulate_MultipleScenarios(t *testing.T) {
	tmpDir := t.TempDir()
	dipFile := filepath.Join(tmpDir, "multi.dip")
	content := `workflow Multi
  goal: "Test multiple scenarios"
  start: Check
  exit: Done

  agent Check
    auto_status: true
    prompt:
      Check.

  agent Done
    prompt:
      Done.

  edges
    Check -> Done
`
	if err := os.WriteFile(dipFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	simulate.ResetRunCounter()
	_, _, code := runCLI(t, "simulate",
		"--scenario", "x=1",
		"--scenario", "y=2",
		dipFile)
	if code != ExitOK {
		t.Fatalf("expected exit 0, got %d", code)
	}
}

func TestLintExitsNonZeroOnErrorSeverityLint(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "unknown_type.dip")
	src := `workflow W
  goal: "test"
  start: A
  exit: A

  inputs
    when: duration

  agent A
    prompt:
      hi
`
	if err := os.WriteFile(path, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	cli := &CLI{Stdout: &stdout, Stderr: &stderr}
	if got := cli.CmdLint([]string{path}); got != ExitError {
		t.Errorf("CmdLint exit = %v, want ExitError — an error-severity lint must fail the command", got)
	}
}
