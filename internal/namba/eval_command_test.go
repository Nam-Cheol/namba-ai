package namba

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
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
