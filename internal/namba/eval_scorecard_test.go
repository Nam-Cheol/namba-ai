package namba

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestEvalScorecardUsesFixedHarnessQualityMetrics(t *testing.T) {
	t.Parallel()

	root := repoRootForHookTest(t)
	app := NewApp(&bytes.Buffer{}, &bytes.Buffer{})
	app.now = func() time.Time { return time.Date(2026, 5, 22, 0, 0, 0, 0, time.UTC) }

	result, err := app.runEvalSuite(context.Background(), root, evalOptions{
		suite:    defaultEvalSuite,
		format:   "json",
		fixture:  defaultHarnessEvalFixture,
		baseline: defaultHarnessEvalBase,
	})
	if err != nil {
		t.Fatalf("run eval suite: %v", err)
	}
	if result.Scorecard == nil {
		t.Fatal("expected eval scorecard")
	}
	data, err := json.Marshal(result.Scorecard)
	if err != nil {
		t.Fatalf("marshal scorecard: %v", err)
	}
	if err := validateEvidenceContractJSON(evalScorecardSchemaVersion, data); err != nil {
		t.Fatalf("scorecard should validate: %v\n%s", err, string(data))
	}

	wantNames := []string{
		"route_selection",
		"clarification_quality",
		"spec_completeness",
		"execution_readiness",
		"review_readiness",
		"dangerous_command_blocking",
		"evidence_completeness",
		"schema_validity",
		"offline_e2e_workflow_health",
	}
	got := map[string]evalScorecardMetric{}
	for _, metric := range result.Scorecard.Metrics {
		got[metric.Name] = metric
	}
	for _, name := range wantNames {
		metric, ok := got[name]
		if !ok {
			t.Fatalf("scorecard missing metric %q; got %+v", name, result.Scorecard.Metrics)
		}
		if metric.Total == 0 {
			t.Fatalf("scorecard metric %q should have scenario coverage", name)
		}
	}
	if result.Scorecard.Summary.Total != result.Summary.Total || result.Scorecard.Summary.Failed != result.Summary.Failed {
		t.Fatalf("scorecard summary should match eval summary: scorecard=%+v result=%+v", result.Scorecard.Summary, result.Summary)
	}
}

func TestEvalCommandWritesScorecardAndSummaryArtifacts(t *testing.T) {
	t.Parallel()

	root := repoRootForHookTest(t)
	tmp := canonicalTempDir(t)
	scorecardPath := filepath.Join(tmp, "scorecard.json")
	summaryPath := filepath.Join(tmp, "summary.md")
	var stdout bytes.Buffer
	app := NewApp(&stdout, &bytes.Buffer{})
	app.now = func() time.Time { return time.Date(2026, 5, 22, 0, 0, 0, 0, time.UTC) }
	restore := chdirExecution(t, root)
	defer restore()

	err := app.Run(context.Background(), []string{
		"eval",
		"--suite", "harness",
		"--format", "json",
		"--baseline", defaultHarnessEvalBase,
		"--fail-on-regression",
		"--scorecard-out", scorecardPath,
		"--summary-out", summaryPath,
	})
	if err != nil {
		t.Fatalf("eval command failed: %v\n%s", err, stdout.String())
	}
	scorecardData, err := os.ReadFile(scorecardPath)
	if err != nil {
		t.Fatalf("read scorecard artifact: %v", err)
	}
	if err := validateEvidenceContractJSON(evalScorecardSchemaVersion, scorecardData); err != nil {
		t.Fatalf("scorecard artifact should validate: %v\n%s", err, string(scorecardData))
	}
	summaryData, err := os.ReadFile(summaryPath)
	if err != nil {
		t.Fatalf("read summary artifact: %v", err)
	}
	for _, want := range []string{"# Namba Eval Scorecard", "`route_selection`", "`offline_e2e_workflow_health`"} {
		if !strings.Contains(string(summaryData), want) {
			t.Fatalf("summary artifact missing %q:\n%s", want, string(summaryData))
		}
	}
}
