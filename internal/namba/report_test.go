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

func TestReportCommandRendersJSONAndAggregatesLocalState(t *testing.T) {
	root := newReportFixture(t)
	stdout := &bytes.Buffer{}
	app := NewApp(stdout, &bytes.Buffer{})
	app.getwd = func() (string, error) { return root, nil }
	app.now = func() time.Time { return time.Date(2026, 5, 21, 9, 0, 0, 0, time.FixedZone("KST", 9*60*60)) }

	if err := app.Run(context.Background(), []string{"report", "--json"}); err != nil {
		t.Fatalf("report --json failed: %v\n%s", err, stdout.String())
	}

	var report nambaReport
	if err := json.Unmarshal(stdout.Bytes(), &report); err != nil {
		t.Fatalf("report should be JSON: %v\n%s", err, stdout.String())
	}
	if report.SchemaVersion != reportSchemaVersion {
		t.Fatalf("unexpected schema version %q", report.SchemaVersion)
	}
	if report.Summary.Health != "blocked" {
		t.Fatalf("expected blocked health, got %+v", report.Summary)
	}
	if report.Summary.RunFailureCount == 0 || report.Summary.BlockedCommandCount == 0 {
		t.Fatalf("expected failure and blocked counts, got %+v", report.Summary)
	}
	if report.Summary.MissingEvidenceCount == 0 {
		t.Fatalf("expected missing evidence count, got %+v", report.Summary)
	}
	if report.Queue.PendingCount != 1 || report.Queue.RunningCount != 1 || report.Queue.BlockedCount != 1 || report.Queue.DoneCount != 1 {
		t.Fatalf("unexpected queue counts: %+v", report.Queue)
	}
	if report.Specs.ReviewReadyCount != 1 || report.Specs.ReviewRequiredCount != 1 {
		t.Fatalf("unexpected review counts: %+v", report.Specs)
	}
	if report.Diagnostics.CodexAvailable != "detected" || report.Diagnostics.RemoteControlStatus != "enabled" {
		t.Fatalf("expected diagnostics summary, got %+v", report.Diagnostics)
	}

	formatStdout := &bytes.Buffer{}
	app = NewApp(formatStdout, &bytes.Buffer{})
	app.getwd = func() (string, error) { return root, nil }
	app.now = func() time.Time { return time.Date(2026, 5, 21, 9, 0, 0, 0, time.FixedZone("KST", 9*60*60)) }
	if err := app.Run(context.Background(), []string{"report", "--format", "json"}); err != nil {
		t.Fatalf("report --format json failed: %v\n%s", err, formatStdout.String())
	}
	if formatStdout.String() != stdout.String() {
		t.Fatalf("--format json should match --json\n--json: %s\n--format json: %s", stdout.String(), formatStdout.String())
	}
}

func TestReportJSONOutputValidatesAgainstSchemaContract(t *testing.T) {
	t.Parallel()

	root := newReportFixture(t)
	stdout := &bytes.Buffer{}
	app := NewApp(stdout, &bytes.Buffer{})
	app.getwd = func() (string, error) { return root, nil }
	app.now = func() time.Time { return time.Date(2026, 5, 22, 10, 0, 0, 0, time.UTC) }

	if err := app.Run(context.Background(), []string{"report", "--format", "json"}); err != nil {
		t.Fatalf("report --format json failed: %v\n%s", err, stdout.String())
	}
	if err := validateEvidenceContractJSON(reportSchemaVersion, stdout.Bytes()); err != nil {
		t.Fatalf("report JSON should validate against %s: %v\n%s", reportSchemaVersion, err, stdout.String())
	}
}

func TestReportCommandHumanOutputSurfacesOperatorSections(t *testing.T) {
	root := newReportFixture(t)
	stdout := &bytes.Buffer{}
	app := NewApp(stdout, &bytes.Buffer{})
	app.getwd = func() (string, error) { return root, nil }
	app.now = func() time.Time { return time.Date(2026, 5, 21, 9, 0, 0, 0, time.FixedZone("KST", 9*60*60)) }

	if err := app.Run(context.Background(), []string{"report", "--format", "text"}); err != nil {
		t.Fatalf("report text failed: %v\n%s", err, stdout.String())
	}
	out := stdout.String()
	for _, want := range []string{"health:", "Top Issues", "Blocked Reasons", "Queue State", "Latest Run", "Stale Candidates", "Release", "Diagnostics", "Next Action"} {
		if !strings.Contains(out, want) {
			t.Fatalf("human report missing %q:\n%s", want, out)
		}
	}
}

func TestReportSinceFiltersTimeSeriesEvidence(t *testing.T) {
	root := newReportFixture(t)
	oldEvidence := executionEvidenceManifest{
		SchemaVersion: executionEvidenceSchemaVersion,
		LogID:         "spec-000",
		RunID:         "spec-000",
		SpecID:        "SPEC-000",
		GeneratedAt:   "2026-05-01T00:00:00Z",
		ExecutionMode: "solo",
		Status:        "completed",
		Request:       executionEvidenceRef{Kind: "request", State: executionEvidenceStatePresent},
		Preflight:     executionEvidenceRef{Kind: "preflight", State: executionEvidenceStatePresent},
		Execution:     executionEvidenceRef{Kind: "execution", State: executionEvidenceStatePresent},
		Validation:    executionEvidenceRef{Kind: "validation", State: executionEvidenceStatePresent},
		Progress:      executionEvidenceRef{Kind: "progress", State: executionEvidenceStateNotApplicable},
	}
	mustWriteJSON(t, filepath.Join(root, ".namba", "logs", "runs", "spec-000-evidence.json"), oldEvidence)

	now := time.Date(2026, 5, 21, 9, 0, 0, 0, time.FixedZone("KST", 9*60*60))
	fullReport := collectNambaReport(root, now, reportOptions{})
	recentReport := collectNambaReport(root, now, reportOptions{since: "1h"})

	if fullReport.Runs.SuccessCount != 1 {
		t.Fatalf("expected unfiltered report to include old success evidence, got %+v", fullReport.Runs)
	}
	if recentReport.Runs.SuccessCount != 0 {
		t.Fatalf("expected --since to exclude old success evidence, got %+v", recentReport.Runs)
	}
	if recentReport.Runs.FailureCount == 0 || recentReport.Runs.Latest == nil || recentReport.Runs.Latest.SpecID != "SPEC-001" {
		t.Fatalf("expected recent failed evidence to remain visible, got %+v", recentReport.Runs)
	}
}

func TestReportParserToleratesMissingCorruptAndPartialData(t *testing.T) {
	root := canonicalTempDir(t)
	mustMkdir(t, filepath.Join(root, ".namba", "logs", "runs"))
	mustMkdir(t, filepath.Join(root, ".namba", "specs"))
	mustWrite(t, filepath.Join(root, ".namba", "logs", "runs", "bad-evidence.json"), "{")
	mustWrite(t, filepath.Join(root, ".namba", "logs", "runs", "bad-parallel-progress.jsonl"), "{\"ok\":true}\n{bad\n")

	report := collectNambaReport(root, time.Date(2026, 5, 21, 9, 0, 0, 0, time.UTC), reportOptions{})
	if report.Runs.ParallelProgressLineCount != 1 {
		t.Fatalf("expected one valid JSONL line, got %+v", report.Runs)
	}
	if len(report.Warnings) < 2 {
		t.Fatalf("expected tolerant warnings for corrupt JSON and JSONL, got %+v", report.Warnings)
	}
	if report.Summary.Health == "" {
		t.Fatalf("expected health to be computed despite corrupt partial data")
	}
}

func TestStatusJSONUsesCompactReportSubset(t *testing.T) {
	root := newReportFixture(t)
	stdout := &bytes.Buffer{}
	app := NewApp(stdout, &bytes.Buffer{})
	app.getwd = func() (string, error) { return root, nil }

	if err := app.Run(context.Background(), []string{"status", "--json"}); err != nil {
		t.Fatalf("status --json failed: %v\n%s", err, stdout.String())
	}
	var status map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &status); err != nil {
		t.Fatalf("status should be JSON: %v\n%s", err, stdout.String())
	}
	if status["schema_version"] != reportSchemaVersion {
		t.Fatalf("unexpected status schema: %+v", status)
	}
	if _, ok := status["runs"]; ok {
		t.Fatalf("status --json should be compact, got full runs field: %+v", status)
	}

	text := &bytes.Buffer{}
	app = NewApp(text, &bytes.Buffer{})
	app.getwd = func() (string, error) { return root, nil }
	if err := app.Run(context.Background(), []string{"status"}); err != nil {
		t.Fatalf("status text failed: %v", err)
	}
	if !strings.Contains(text.String(), "SPEC packages:") || strings.Contains(text.String(), "schema_version") {
		t.Fatalf("plain status contract changed: %q", text.String())
	}
}

func TestReportRejectsUnsupportedFlags(t *testing.T) {
	stdout := &bytes.Buffer{}
	app := NewApp(stdout, &bytes.Buffer{})
	err := app.Run(context.Background(), []string{"report", "--format", "yaml"})
	if err == nil || !strings.Contains(err.Error(), `unsupported report format "yaml"`) {
		t.Fatalf("expected unsupported format usage error, got %v", err)
	}

	err = app.Run(context.Background(), []string{"report", "--since", "later"})
	if err == nil || !strings.Contains(err.Error(), `unsupported --since value "later"`) {
		t.Fatalf("expected unsupported since usage error, got %v", err)
	}
}

func newReportFixture(t *testing.T) string {
	t.Helper()
	root := canonicalTempDir(t)
	mustMkdir(t, filepath.Join(root, ".namba", "logs", "runs"))
	mustMkdir(t, filepath.Join(root, ".namba", "logs", "queue"))
	mustMkdir(t, filepath.Join(root, ".namba", "logs", "project"))
	mustMkdir(t, filepath.Join(root, ".namba", "specs", "SPEC-001", "reviews"))
	mustMkdir(t, filepath.Join(root, ".namba", "specs", "SPEC-002", "reviews"))
	mustMkdir(t, filepath.Join(root, ".namba", "project"))
	mustMkdir(t, filepath.Join(root, ".namba", "releases"))
	mustWrite(t, filepath.Join(root, ".namba", "config", "sections", "project.yaml"), "name: fixture\nproject_type: existing\nlanguage: go\nframework: cli\n")
	mustWrite(t, filepath.Join(root, ".namba", "config", "sections", "quality.yaml"), "development_mode: tdd\n")
	for _, specID := range []string{"SPEC-001", "SPEC-002"} {
		mustWrite(t, filepath.Join(root, ".namba", "specs", specID, "spec.md"), "# spec\n")
		mustWrite(t, filepath.Join(root, ".namba", "specs", specID, "plan.md"), "# plan\n")
		mustWrite(t, filepath.Join(root, ".namba", "specs", specID, "acceptance.md"), "# acceptance\n")
	}
	mustWrite(t, filepath.Join(root, ".namba", "specs", "SPEC-001", "reviews", "readiness.md"), "# Review Readiness\n\n## Summary\n\n- Cleared reviews: 3/3\n- Advisory status: all current review tracks are marked clear.\n")
	mustWrite(t, filepath.Join(root, ".namba", "specs", "SPEC-002", "reviews", "readiness.md"), "# Review Readiness\n\n## Summary\n\n- Cleared reviews: 1/3\n- Advisory status: follow up on engineering=pending before execution.\n")
	mustWrite(t, filepath.Join(root, ".namba", "project", "release-checklist.md"), "- [x] notes\n- [ ] tag\n")
	mustWrite(t, filepath.Join(root, ".namba", "releases", "v0.1.0.md"), "# v0.1.0\n")

	evidence := executionEvidenceManifest{
		SchemaVersion: executionEvidenceSchemaVersion,
		LogID:         "spec-001",
		RunID:         "spec-001",
		SpecID:        "SPEC-001",
		GeneratedAt:   "2026-05-21T00:00:00Z",
		ExecutionMode: "solo",
		Status:        "validation_failed",
		Request:       executionEvidenceRef{Kind: "request", State: executionEvidenceStatePresent, Path: ".namba/logs/runs/spec-001-request.json"},
		Preflight:     executionEvidenceRef{Kind: "preflight", State: executionEvidenceStateMissing, Path: ".namba/logs/runs/spec-001-preflight.json"},
		Execution:     executionEvidenceRef{Kind: "execution", State: executionEvidenceStatePresent, Path: ".namba/logs/runs/spec-001-execution.json"},
		Validation:    executionEvidenceRef{Kind: "validation", State: executionEvidenceStateMissing, Path: ".namba/logs/runs/spec-001-validation.json"},
		Progress:      executionEvidenceRef{Kind: "progress", State: executionEvidenceStateNotApplicable},
		Hooks: []hookResult{{
			HookName:      "validation-guard",
			Status:        "failed",
			Blocking:      true,
			FailureAction: "validation_hook_failed",
		}},
	}
	mustWriteJSON(t, filepath.Join(root, ".namba", "logs", "runs", "spec-001-evidence.json"), evidence)
	mustWriteJSON(t, filepath.Join(root, ".namba", "logs", "runs", "spec-001-validation.json"), validationReport{SpecID: "SPEC-001", Passed: false, Attempt: 1})
	mustWrite(t, filepath.Join(root, ".namba", "logs", "runs", "spec-001-parallel-progress.jsonl"), "{\"phase\":\"running\"}\n{bad\n")
	mustWriteJSON(t, filepath.Join(root, ".namba", "logs", "runs", "spec-001-heartbeat.json"), queueRunnerHeartbeat{SpecID: "SPEC-001", Runner: "cli", Status: "running", UpdatedAt: "2026-05-21T07:00:00+09:00"})
	mustWriteJSON(t, filepath.Join(root, ".namba", "logs", "queue", "state.json"), queueState{
		Status:        queueStateBlocked,
		OperatorState: queueOperatorBlocked,
		ActiveSpecID:  "SPEC-001",
		LastBlocker:   "checks_failed",
		Specs: map[string]queueSpec{
			"SPEC-001": {SpecID: "SPEC-001", OperatorState: queueOperatorRunning, Phase: queuePhaseRunning},
			"SPEC-002": {SpecID: "SPEC-002", OperatorState: queueOperatorBlocked, Phase: queuePhaseBlocked, Blocker: "review_required"},
			"SPEC-003": {SpecID: "SPEC-003", OperatorState: queueOperatorDone, Phase: queuePhaseLanded},
			"SPEC-004": {SpecID: "SPEC-004", OperatorState: queueOperatorWaiting, Phase: queuePhasePending},
		},
	})
	mustWriteJSON(t, filepath.Join(root, ".namba", "logs", "project", "codex-diagnostics-evidence.json"), projectCodexDiagnosticsEvidence{
		SchemaVersion: projectCodexDiagnosticsSchema,
		GeneratedAt:   "2026-05-21T00:00:00Z",
		ProjectRoot:   root,
		Diagnostics: codexDiagnosticsEvidence{
			GeneratedAt:             "2026-05-21T00:00:00Z",
			CodexAvailable:          "detected",
			Version:                 codexVersionEvidence{Status: "detected", BaselineComparison: "meets_baseline"},
			Doctor:                  codexDoctorEvidence{Status: "passed"},
			WorkspaceRootComparison: codexRootComparison{Status: "matched"},
			RemoteControl:           codexRemoteControl{Status: "enabled"},
			RemoteEnvironments:      codexRemoteEnvironments{Status: "detected"},
		},
	})
	return root
}

func mustMkdir(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", path, err)
	}
}

func mustWrite(t *testing.T, path, body string) {
	t.Helper()
	mustMkdir(t, filepath.Dir(path))
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func mustWriteJSON(t *testing.T, path string, value any) {
	t.Helper()
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		t.Fatalf("marshal %s: %v", path, err)
	}
	mustWrite(t, path, string(data))
}
