package namba

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestEvalCommandRendersJSONAndComparesBaseline(t *testing.T) {
	t.Parallel()

	root := repoRootForHookTest(t)
	var stdout bytes.Buffer
	app := NewApp(&stdout, &bytes.Buffer{})
	app.now = func() time.Time { return time.Date(2026, 5, 20, 8, 0, 0, 0, time.UTC) }
	restore := chdirExecution(t, root)
	defer restore()

	err := app.Run(context.Background(), []string{"eval", "--suite", "harness", "--format", "json", "--fail-on-regression"})
	if err != nil {
		t.Fatalf("eval command failed: %v\n%s", err, stdout.String())
	}
	var result evalRunResult
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("eval output should be json: %v\n%s", err, stdout.String())
	}
	if result.SchemaVersion != evalResultSchemaVersion {
		t.Fatalf("unexpected result schema %q", result.SchemaVersion)
	}
	if result.Summary.Total < 24 || result.Summary.Failed != 0 {
		t.Fatalf("unexpected eval summary: %+v", result.Summary)
	}
	if !result.Baseline.Compared || !result.Baseline.Passed {
		t.Fatalf("expected checked-in baseline comparison to pass: %+v", result.Baseline)
	}
	if len(result.Regressions) != 0 {
		t.Fatalf("expected no regressions, got %+v", result.Regressions)
	}
}

func TestEvalCommandMarkdownShowsSummaryAndMetrics(t *testing.T) {
	t.Parallel()

	root := repoRootForHookTest(t)
	var stdout bytes.Buffer
	app := NewApp(&stdout, &bytes.Buffer{})
	restore := chdirExecution(t, root)
	defer restore()

	if err := app.Run(context.Background(), []string{"eval", "--suite", "harness", "--format", "markdown", "--case", "route_core_pr_review_opt_in_workflow"}); err != nil {
		t.Fatalf("eval markdown failed: %v\n%s", err, stdout.String())
	}
	got := stdout.String()
	for _, want := range []string{"# Namba Eval Report", "scenarios: 1 total, 1 passed, 0 failed", "`route_selection`"} {
		if !strings.Contains(got, want) {
			t.Fatalf("expected markdown output to contain %q, got %q", want, got)
		}
	}
}

func TestEvalCommandInvalidFixtureUsesExitCodeTwo(t *testing.T) {
	t.Parallel()

	root := canonicalTempDir(t)
	if err := NewApp(&bytes.Buffer{}, &bytes.Buffer{}).Run(context.Background(), []string{"init", root, "--yes"}); err != nil {
		t.Fatalf("init temp repo: %v", err)
	}
	badFixture := filepath.Join(root, "bad-scenarios.json")
	writeTestFile(t, badFixture, `{"schema_version":"wrong","suite":"harness","scenarios":[]}`)

	var stdout bytes.Buffer
	app := NewApp(&stdout, &bytes.Buffer{})
	restore := chdirExecution(t, root)
	defer restore()

	err := app.Run(context.Background(), []string{"eval", "--fixture", badFixture})
	if err == nil {
		t.Fatal("expected invalid fixture to fail")
	}
	if code := ExitCode(err); code != 2 {
		t.Fatalf("expected exit code 2 for invalid fixture, got %d (%v)", code, err)
	}
}

func TestEvalCommandBaselineRegressionUsesExitCodeOne(t *testing.T) {
	t.Parallel()

	root := repoRootForHookTest(t)
	tmp := canonicalTempDir(t)
	badBaseline := filepath.Join(tmp, "baseline.json")
	writeTestFile(t, badBaseline, `{
  "schema_version": "namba-eval-baseline/v1",
  "result_schema_version": "namba-eval-results/v1",
  "suite": "harness",
  "corpus_version": "2026-05-20.v1",
  "scenario_fingerprints": {"route_core_pr_review_opt_in_workflow": "wrong"},
  "metric_pass_counts": {"route_selection": 1},
  "required_coverage": ["core-runtime-or-harness-change"]
}`)

	var stdout bytes.Buffer
	app := NewApp(&stdout, &bytes.Buffer{})
	restore := chdirExecution(t, root)
	defer restore()

	err := app.Run(context.Background(), []string{"eval", "--format", "json", "--baseline", badBaseline, "--fail-on-regression", "--case", "route_core_pr_review_opt_in_workflow"})
	if err == nil {
		t.Fatal("expected baseline regression to fail")
	}
	var exitErr evalExitError
	if !errors.As(err, &exitErr) || exitErr.code != 1 {
		t.Fatalf("expected exit code 1 regression error, got %T %v", err, err)
	}
	if !strings.Contains(stdout.String(), "changed fingerprint") {
		t.Fatalf("expected regression diagnostic in output, got %q", stdout.String())
	}
}

func TestEvalCommandCaseBaselineComparisonIgnoresUnselectedScenarios(t *testing.T) {
	t.Parallel()

	root := repoRootForHookTest(t)
	var stdout bytes.Buffer
	app := NewApp(&stdout, &bytes.Buffer{})
	restore := chdirExecution(t, root)
	defer restore()

	err := app.Run(context.Background(), []string{"eval", "--format", "json", "--fail-on-regression", "--case", "route_core_pr_review_opt_in_workflow"})
	if err != nil {
		t.Fatalf("expected selected eval case to pass baseline comparison: %v\n%s", err, stdout.String())
	}
	var result evalRunResult
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("eval output should be json: %v\n%s", err, stdout.String())
	}
	if result.Summary.Total != 1 || result.Summary.RegressionCount != 0 || len(result.Regressions) != 0 {
		t.Fatalf("expected only the selected case without false regressions, got summary=%+v regressions=%+v", result.Summary, result.Regressions)
	}
}

func TestEvalCommandRejectsCaseBaselineUpdate(t *testing.T) {
	t.Parallel()

	root := repoRootForHookTest(t)
	var stdout bytes.Buffer
	app := NewApp(&stdout, &bytes.Buffer{})
	restore := chdirExecution(t, root)
	defer restore()

	err := app.Run(context.Background(), []string{"eval", "--case", "route_core_pr_review_opt_in_workflow", "--update-baseline"})
	if err == nil {
		t.Fatal("expected --case with --update-baseline to fail")
	}
	if code := ExitCode(err); code != 2 {
		t.Fatalf("expected exit code 2 for unsafe partial baseline update, got %d (%v)", code, err)
	}
	if !strings.Contains(err.Error(), "--update-baseline cannot be combined with --case") {
		t.Fatalf("expected partial update diagnostic, got %v", err)
	}
}

func TestEvalCommandDoesNotUpdateBaselineWhenScenariosFail(t *testing.T) {
	t.Parallel()

	root := repoRootForHookTest(t)
	tmp := canonicalTempDir(t)
	fixturePath := filepath.Join(tmp, "failing-scenarios.json")
	baselinePath := filepath.Join(tmp, "baseline.json")
	writeTestFile(t, fixturePath, `{
  "schema_version": "namba-eval-scenarios/v1",
  "suite": "harness",
  "corpus_version": "test",
  "scenarios": [
    {
      "id": "route_command_regression",
      "type": "route",
      "tags": ["coverage:core-runtime-or-harness-change"],
      "input": "Change the namba pr workflow so Codex review is opt-in",
      "expected": {
        "route_selection": "core",
        "command": "namba harness",
        "clarification_required": false,
        "execution_ready": true
      },
      "rationale": "Intentional command mismatch for baseline update guard."
    }
  ]
}`)
	writeTestFile(t, baselinePath, "sentinel\n")

	var stdout bytes.Buffer
	app := NewApp(&stdout, &bytes.Buffer{})
	restore := chdirExecution(t, root)
	defer restore()

	err := app.Run(context.Background(), []string{"eval", "--fixture", fixturePath, "--baseline", baselinePath, "--format", "json", "--update-baseline"})
	if err == nil {
		t.Fatal("expected failing scenarios to block baseline update")
	}
	if code := ExitCode(err); code != 1 {
		t.Fatalf("expected exit code 1 for scenario failure, got %d (%v)", code, err)
	}
	data, readErr := os.ReadFile(baselinePath)
	if readErr != nil {
		t.Fatalf("read baseline after failed update: %v", readErr)
	}
	if string(data) != "sentinel\n" {
		t.Fatalf("baseline should not be rewritten when scenarios fail, got %q", string(data))
	}
}

func TestEvalComparisonIncludesCommand(t *testing.T) {
	t.Parallel()

	failures := compareExpectedActual(
		map[string]any{"route_selection": "domain", "command": "namba harness"},
		map[string]any{"route_selection": "domain", "command": "namba plan"},
	)
	if len(failures) != 1 || !strings.Contains(failures[0], "command expected namba harness, got namba plan") {
		t.Fatalf("expected command mismatch failure, got %+v", failures)
	}
}

func TestEvalComparisonIncludesEventType(t *testing.T) {
	t.Parallel()

	failures := compareExpectedActual(
		map[string]any{"event_type": "PreToolUse"},
		map[string]any{"event_type": "PermissionRequest"},
	)
	if len(failures) != 1 || !strings.Contains(failures[0], "event_type expected PreToolUse, got PermissionRequest") {
		t.Fatalf("expected event_type mismatch failure, got %+v", failures)
	}
}

func TestEvalEvidenceBuilderRejectsEscapingExistingPaths(t *testing.T) {
	t.Parallel()

	_, err := buildEvalEvidenceManifest(evalEvidenceBuilderFixture{
		LogID:         "eval-evidence",
		SpecID:        "SPEC-001",
		ExecutionMode: string(executionModeDefault),
		Status:        "completed",
		ExistingPaths: []string{"../../outside.json"},
	})
	if err == nil {
		t.Fatal("expected escaping existing path to fail")
	}
	if !strings.Contains(err.Error(), "escapes eval evidence workspace") {
		t.Fatalf("expected workspace escape diagnostic, got %v", err)
	}
}

func TestEvalGuardrailScenarioUsesInputCommand(t *testing.T) {
	t.Parallel()

	actual, failures := evaluateGuardrailScenario(evalScenario{
		Input:    "git clean -fd",
		Expected: map[string]any{"event_type": "PreToolUse", "command": "go test ./..."},
	})
	if len(failures) != 0 {
		t.Fatalf("expected no direct guardrail failures, got %+v", failures)
	}
	if actual["command"] != "git clean -fd" {
		t.Fatalf("expected actual command from scenario input, got %v", actual["command"])
	}
	if actual["deny"] != true || actual["event_type"] != "PreToolUse" {
		t.Fatalf("expected denied pre-tool command from input, got %+v", actual)
	}
}

func TestEvalFileReadsDiskBeforeEmbeddedFixture(t *testing.T) {
	t.Parallel()

	root := canonicalTempDir(t)
	fixturePath := filepath.Join(root, filepath.FromSlash(defaultHarnessEvalFixture))
	writeTestFile(t, fixturePath, "disk fixture\n")

	data, err := NewApp(&bytes.Buffer{}, &bytes.Buffer{}).readEvalFile(root, defaultHarnessEvalFixture)
	if err != nil {
		t.Fatalf("read eval fixture: %v", err)
	}
	if string(data) != "disk fixture\n" {
		t.Fatalf("expected disk fixture to win, got %q", string(data))
	}
}

func TestEvalFileFallsBackToEmbeddedFixture(t *testing.T) {
	t.Parallel()

	data, err := NewApp(&bytes.Buffer{}, &bytes.Buffer{}).readEvalFile(canonicalTempDir(t), defaultHarnessEvalFixture)
	if err != nil {
		t.Fatalf("read embedded eval fixture: %v", err)
	}
	if !strings.Contains(string(data), `"schema_version": "namba-eval-scenarios/v1"`) {
		t.Fatalf("expected embedded eval fixture, got %q", string(data))
	}
}
