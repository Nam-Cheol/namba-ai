package namba

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunWritesExecutionEvidenceManifestOnSuccess(t *testing.T) {
	tmp, app, restore := prepareExecutionProject(t)
	defer restore()

	app.lookPath = func(name string) (string, error) {
		switch name {
		case "codex", "git":
			return name, nil
		default:
			return "", errors.New("missing dependency")
		}
	}
	app.runCmd = func(_ context.Context, name string, args []string, dir string) (string, error) {
		switch {
		case isCodexExec(name, args):
			return "runner output", nil
		case isShellCommand(name):
			return "validation ok", nil
		default:
			t.Fatalf("unexpected command: %s %v", name, args)
			return "", nil
		}
	}

	if err := app.Run(context.Background(), []string{"run", "SPEC-001"}); err != nil {
		t.Fatalf("run failed: %v", err)
	}

	manifest := mustReadExecutionEvidenceManifest(t, filepath.Join(tmp, ".namba", "logs", "runs", "spec-001-evidence.json"))
	if manifest.SchemaVersion != executionEvidenceSchemaVersion || manifest.LogID != "spec-001" || manifest.SpecID != "SPEC-001" {
		t.Fatalf("unexpected execution evidence identity: %+v", manifest)
	}
	if manifest.Status != "completed" || manifest.ExecutionMode != string(executionModeDefault) {
		t.Fatalf("expected completed default manifest, got %+v", manifest)
	}
	for name, ref := range map[string]executionEvidenceRef{
		"request":    manifest.Request,
		"preflight":  manifest.Preflight,
		"execution":  manifest.Execution,
		"validation": manifest.Validation,
	} {
		if ref.State != executionEvidenceStatePresent {
			t.Fatalf("expected %s evidence to be present, got %+v", name, ref)
		}
	}
	if manifest.Progress.State != executionEvidenceStateNotApplicable {
		t.Fatalf("expected progress evidence to stay not_applicable, got %+v", manifest.Progress)
	}
	if manifest.Extensions.Browser.State != executionEvidenceStateNotApplicable {
		t.Fatalf("expected browser evidence to stay not_applicable, got %+v", manifest.Extensions.Browser)
	}
	if manifest.Extensions.Runtime.State != executionEvidenceStatePresent {
		t.Fatalf("expected runtime extension to be present, got %+v", manifest.Extensions.Runtime)
	}
	if len(manifest.Extensions.Runtime.SignalBundles) != 1 || manifest.Extensions.Runtime.SignalBundles[0].Kind != "validation_attempts" {
		t.Fatalf("expected validation-attempt runtime bundle, got %+v", manifest.Extensions.Runtime.SignalBundles)
	}
	if !strings.Contains(strings.Join(manifest.Extensions.Runtime.SignalBundles[0].Paths, "\n"), "spec-001-validation-attempt-1.json") {
		t.Fatalf("expected validation-attempt path in runtime bundle, got %+v", manifest.Extensions.Runtime.SignalBundles[0])
	}
}

func TestBuildExecutionEvidenceManifestExcludesStaleValidationAttempts(t *testing.T) {
	tmp := t.TempDir()
	writeTestFile(t, filepath.Join(tmp, ".namba", "logs", "runs", "spec-001-validation-attempt-1.json"), "{}")
	writeTestFile(t, filepath.Join(tmp, ".namba", "logs", "runs", "spec-001-validation-attempt-2.json"), "{}")

	manifest, err := buildExecutionEvidenceManifest(tmp, executionEvidenceOptions{
		ProjectRoot:        tmp,
		LogID:              "spec-001",
		SpecID:             "SPEC-001",
		ExecutionMode:      executionModeDefault,
		Status:             "completed",
		ValidationAttempts: 1,
	})
	if err != nil {
		t.Fatalf("buildExecutionEvidenceManifest failed: %v", err)
	}
	if len(manifest.Extensions.Runtime.SignalBundles) != 1 {
		t.Fatalf("expected one validation-attempt bundle, got %+v", manifest.Extensions.Runtime.SignalBundles)
	}
	if paths := manifest.Extensions.Runtime.SignalBundles[0].Paths; len(paths) != 1 || paths[0] != ".namba/logs/runs/spec-001-validation-attempt-1.json" {
		t.Fatalf("expected only current-run validation attempt path, got %+v", manifest.Extensions.Runtime.SignalBundles[0])
	}

	manifest, err = buildExecutionEvidenceManifest(tmp, executionEvidenceOptions{
		ProjectRoot:        tmp,
		LogID:              "spec-001",
		SpecID:             "SPEC-001",
		ExecutionMode:      executionModeDefault,
		Status:             "execution_failed",
		ValidationAttempts: 0,
	})
	if err != nil {
		t.Fatalf("buildExecutionEvidenceManifest failed without validation attempts: %v", err)
	}
	if len(manifest.Extensions.Runtime.SignalBundles) != 0 {
		t.Fatalf("expected stale validation attempts to be ignored when the current run never validated, got %+v", manifest.Extensions.Runtime.SignalBundles)
	}
	if manifest.Extensions.Runtime.State != executionEvidenceStateNotApplicable {
		t.Fatalf("expected runtime extension to stay not_applicable without current-run attempts, got %+v", manifest.Extensions.Runtime)
	}
}

func TestExecuteRunWritesProgressFailureEvidenceWhenFinalPublishFails(t *testing.T) {
	tmp, app, restore := prepareExecutionProject(t)
	defer restore()

	app.lookPath = func(name string) (string, error) {
		switch name {
		case "codex", "git":
			return name, nil
		default:
			return "", errors.New("missing dependency")
		}
	}
	app.runCmd = func(_ context.Context, name string, args []string, dir string) (string, error) {
		if isCodexExec(name, args) {
			return "runner output", nil
		}
		t.Fatalf("unexpected command: %s %v", name, args)
		return "", nil
	}

	req := app.newExecutionRequest(
		"SPEC-001",
		tmp,
		"prompt",
		executionModeParallel,
		suggestDelegationPlan(executionModeParallel, "prompt", "", ""),
		systemConfig{Runner: "codex", ApprovalPolicy: "on-request", SandboxMode: "workspace-write"},
		codexConfig{},
	)

	progress := &stubParallelProgressSink{
		path: filepath.Join(tmp, ".namba", "logs", "runs", "spec-001-progress.events.jsonl"),
		failMatch: func(input parallelProgressEventInput, publishCount int) bool {
			return input.Phase == "merge_pending"
		},
		failPublishErr: errors.New("append denied"),
	}

	result, report, err := app.executeRun(
		context.Background(),
		tmp,
		"spec-001",
		req,
		tmp,
		qualityConfig{TestCommand: "none", LintCommand: "none", TypecheckCommand: "none"},
		progress,
		"spec-001-p1",
	)
	if err == nil || !strings.Contains(err.Error(), "append denied") {
		t.Fatalf("expected final publish failure, got err=%v result=%+v report=%+v", err, result, report)
	}

	manifest := mustReadExecutionEvidenceManifest(t, filepath.Join(tmp, ".namba", "logs", "runs", "spec-001-evidence.json"))
	if manifest.Status != "progress_log_failed" {
		t.Fatalf("expected progress_log_failed evidence after final publish failure, got %+v", manifest)
	}
	if manifest.Finalization.ProgressLogFailed != true {
		t.Fatalf("expected finalization to mark progress log failure, got %+v", manifest.Finalization)
	}
	if manifest.Progress.Path != ".namba/logs/runs/spec-001-progress.events.jsonl" || manifest.Progress.State != executionEvidenceStateMissing {
		t.Fatalf("expected progress reference to point at shared progress log path, got %+v", manifest.Progress)
	}
	if manifest.Validation.State != executionEvidenceStatePresent {
		t.Fatalf("expected validation artifact to remain present, got %+v", manifest.Validation)
	}
}

func TestExecuteRunPreservesPreviousValidationAttemptsWhenRetryPublishFails(t *testing.T) {
	tmp, app, restore := prepareExecutionProject(t)
	defer restore()

	app.lookPath = func(name string) (string, error) {
		switch name {
		case "codex", "git":
			return name, nil
		default:
			return "", errors.New("missing dependency")
		}
	}

	validationCalls := 0
	app.runCmd = func(_ context.Context, name string, args []string, dir string) (string, error) {
		switch {
		case isCodexExec(name, args):
			return "runner output", nil
		case isShellCommand(name):
			validationCalls++
			if validationCalls == 1 {
				return "validation failed", errors.New("lint failed")
			}
			t.Fatalf("validation should not start a second attempt after publish failure: %s %v", name, args)
			return "", nil
		default:
			t.Fatalf("unexpected command: %s %v", name, args)
			return "", nil
		}
	}

	req := app.newExecutionRequest(
		"SPEC-001",
		tmp,
		"prompt",
		executionModeParallel,
		suggestDelegationPlan(executionModeParallel, "prompt", "", ""),
		systemConfig{Runner: "codex", ApprovalPolicy: "on-request", SandboxMode: "workspace-write"},
		codexConfig{},
	)
	req.RepairAttempts = 1

	progress := &stubParallelProgressSink{
		path: filepath.Join(tmp, ".namba", "logs", "runs", "spec-001-progress.events.jsonl"),
		failMatch: func(input parallelProgressEventInput, publishCount int) bool {
			if input.Phase != "validating" {
				return false
			}
			attempt, _ := input.Metadata["attempt"].(int)
			return attempt == 2
		},
		failPublishErr: errors.New("append denied"),
	}

	result, report, err := app.executeRun(
		context.Background(),
		tmp,
		"spec-001",
		req,
		tmp,
		qualityConfig{TestCommand: "none", LintCommand: "lint", TypecheckCommand: "none"},
		progress,
		"spec-001-p1",
	)
	if err == nil || !strings.Contains(err.Error(), "append denied") {
		t.Fatalf("expected retry publish failure, got err=%v result=%+v report=%+v", err, result, report)
	}
	if validationCalls != 1 {
		t.Fatalf("expected exactly one completed validation attempt before publish failure, got %d", validationCalls)
	}

	manifest := mustReadExecutionEvidenceManifest(t, filepath.Join(tmp, ".namba", "logs", "runs", "spec-001-evidence.json"))
	if manifest.Status != "progress_log_failed" {
		t.Fatalf("expected progress_log_failed evidence after retry publish failure, got %+v", manifest)
	}
	if manifest.Finalization.ProgressLogFailed != true {
		t.Fatalf("expected finalization to mark progress log failure, got %+v", manifest.Finalization)
	}
	if manifest.Progress.Path != ".namba/logs/runs/spec-001-progress.events.jsonl" || manifest.Progress.State != executionEvidenceStateMissing {
		t.Fatalf("expected progress reference to point at shared progress log path, got %+v", manifest.Progress)
	}
	if len(manifest.Extensions.Runtime.SignalBundles) != 1 {
		t.Fatalf("expected validation-attempt bundle to remain present, got %+v", manifest.Extensions.Runtime.SignalBundles)
	}
	if paths := manifest.Extensions.Runtime.SignalBundles[0].Paths; len(paths) != 1 || paths[0] != ".namba/logs/runs/spec-001-validation-attempt-1.json" {
		t.Fatalf("expected first validation attempt to remain in evidence bundle, got %+v", manifest.Extensions.Runtime.SignalBundles[0])
	}
}

func TestExecuteRunExecutionFailureRecordsProgressFailureSeparately(t *testing.T) {
	tmp, app, restore := prepareExecutionProject(t)
	defer restore()

	app.lookPath = func(name string) (string, error) {
		switch name {
		case "codex", "git":
			return name, nil
		default:
			return "", errors.New("missing dependency")
		}
	}
	app.runCmd = func(_ context.Context, name string, args []string, dir string) (string, error) {
		switch {
		case isCodexExec(name, args):
			return "partial output", errors.New("runner failed")
		case isShellCommand(name):
			t.Fatal("validators should not run after runner failure")
			return "", nil
		default:
			t.Fatalf("unexpected command: %s %v", name, args)
			return "", nil
		}
	}

	req := app.newExecutionRequest(
		"SPEC-001",
		tmp,
		"prompt",
		executionModeParallel,
		suggestDelegationPlan(executionModeParallel, "prompt", "", ""),
		systemConfig{Runner: "codex", ApprovalPolicy: "on-request", SandboxMode: "workspace-write"},
		codexConfig{},
	)

	progress := &stubParallelProgressSink{
		path: filepath.Join(tmp, ".namba", "logs", "runs", "spec-001-progress.events.jsonl"),
		failMatch: func(input parallelProgressEventInput, publishCount int) bool {
			return input.Phase == "failed" && input.Status == "execution_failed"
		},
		failPublishErr: errors.New("append denied"),
	}

	result, report, err := app.executeRun(
		context.Background(),
		tmp,
		"spec-001",
		req,
		tmp,
		qualityConfig{TestCommand: "none", LintCommand: "none", TypecheckCommand: "none"},
		progress,
		"spec-001-p1",
	)
	if err == nil || !strings.Contains(err.Error(), "runner failed") || !strings.Contains(err.Error(), "append denied") {
		t.Fatalf("expected runner failure plus progress publish failure, got err=%v result=%+v report=%+v", err, result, report)
	}

	manifest := mustReadExecutionEvidenceManifest(t, filepath.Join(tmp, ".namba", "logs", "runs", "spec-001-evidence.json"))
	if manifest.Status != "execution_failed" {
		t.Fatalf("expected execution_failed evidence status to remain primary, got %+v", manifest)
	}
	if manifest.Finalization.ProgressLogFailed != true {
		t.Fatalf("expected finalization to record progress log failure separately, got %+v", manifest.Finalization)
	}
	if manifest.Progress.Path != ".namba/logs/runs/spec-001-progress.events.jsonl" || manifest.Progress.State != executionEvidenceStateMissing {
		t.Fatalf("expected progress reference to be preserved on execution failure, got %+v", manifest.Progress)
	}
}

func TestRunWritesExecutionEvidenceManifestOnPreflightFailure(t *testing.T) {
	tmp, app, restore := prepareExecutionProject(t)
	defer restore()

	writeTestFile(t, filepath.Join(tmp, ".namba", "config", "sections", "codex.yaml"), "agent_mode: multi\nstatus_line_preset: namba\nrepo_skills_path: .agents/skills\nrepo_agents_path: .codex/agents\nweb_search: false\nadd_dirs: missing-dir\nsession_mode: stateful\nrepair_attempts: 1\n")

	app.lookPath = func(name string) (string, error) {
		if name == "codex" || name == "git" {
			return name, nil
		}
		return "", errors.New("missing dependency")
	}

	err := app.Run(context.Background(), []string{"run", "SPEC-001"})
	if err == nil {
		t.Fatal("expected preflight failure")
	}

	manifest := mustReadExecutionEvidenceManifest(t, filepath.Join(tmp, ".namba", "logs", "runs", "spec-001-evidence.json"))
	if manifest.Status != "preflight_failed" {
		t.Fatalf("expected preflight_failed status, got %+v", manifest)
	}
	if manifest.Request.State != executionEvidenceStatePresent || manifest.Preflight.State != executionEvidenceStatePresent || manifest.Execution.State != executionEvidenceStatePresent {
		t.Fatalf("expected request/preflight/execution evidence on preflight failure, got %+v", manifest)
	}
	if manifest.Validation.State != executionEvidenceStateMissing {
		t.Fatalf("expected validation evidence to be missing on preflight failure, got %+v", manifest.Validation)
	}
}

func TestRunWritesExecutionEvidenceManifestOnExecutionFailure(t *testing.T) {
	tmp, app, restore := prepareExecutionProject(t)
	defer restore()

	app.lookPath = func(name string) (string, error) {
		if name == "codex" || name == "git" {
			return name, nil
		}
		return "", errors.New("missing dependency")
	}
	app.runCmd = func(_ context.Context, name string, args []string, dir string) (string, error) {
		switch {
		case isCodexExec(name, args):
			return "partial output", errors.New("runner failed")
		case isShellCommand(name):
			t.Fatal("validators should not run after runner failure")
			return "", nil
		default:
			t.Fatalf("unexpected command: %s %v", name, args)
			return "", nil
		}
	}

	err := app.Run(context.Background(), []string{"run", "SPEC-001"})
	if err == nil {
		t.Fatal("expected execution failure")
	}

	manifest := mustReadExecutionEvidenceManifest(t, filepath.Join(tmp, ".namba", "logs", "runs", "spec-001-evidence.json"))
	if manifest.Status != "execution_failed" {
		t.Fatalf("expected execution_failed status, got %+v", manifest)
	}
	if manifest.Execution.State != executionEvidenceStatePresent || manifest.Validation.State != executionEvidenceStateMissing {
		t.Fatalf("expected execution evidence without validation on runner failure, got %+v", manifest)
	}
}

func TestRunWritesExecutionEvidenceManifestOnValidationFailure(t *testing.T) {
	tmp, app, restore := prepareExecutionProject(t)
	defer restore()

	app.lookPath = func(name string) (string, error) {
		if name == "codex" || name == "git" {
			return name, nil
		}
		return "", errors.New("missing dependency")
	}
	app.runCmd = func(_ context.Context, name string, args []string, dir string) (string, error) {
		switch {
		case isCodexExec(name, args):
			return "runner output", nil
		case isShellCommand(name):
			command := args[len(args)-1]
			if strings.Contains(command, "gofmt") {
				return "formatting failed", errors.New("lint failed")
			}
			return "ok", nil
		default:
			t.Fatalf("unexpected command: %s %v", name, args)
			return "", nil
		}
	}

	err := app.Run(context.Background(), []string{"run", "SPEC-001"})
	if err == nil {
		t.Fatal("expected validation failure")
	}

	manifest := mustReadExecutionEvidenceManifest(t, filepath.Join(tmp, ".namba", "logs", "runs", "spec-001-evidence.json"))
	if manifest.Status != "validation_failed" {
		t.Fatalf("expected validation_failed status, got %+v", manifest)
	}
	if manifest.Validation.State != executionEvidenceStatePresent {
		t.Fatalf("expected validation artifact on validation failure, got %+v", manifest.Validation)
	}
	if manifest.Extensions.Runtime.State != executionEvidenceStatePresent {
		t.Fatalf("expected runtime extension on validation failure, got %+v", manifest.Extensions.Runtime)
	}
}

func TestRunParallelWritesExecutionEvidenceManifest(t *testing.T) {
	h, restore := newParallelHarness(t)
	defer restore()

	err := h.app.runParallel(
		context.Background(),
		h.tmp,
		specPackage{ID: "SPEC-003"},
		[]string{"one", "two", "three"},
		"prompt",
		qualityConfig{TestCommand: "test", LintCommand: "none", TypecheckCommand: "none"},
		systemConfig{Runner: "codex"},
		codexConfig{},
		workflowConfig{MaxParallelWorkers: 3},
		false,
	)
	if err != nil {
		t.Fatalf("runParallel failed: %v", err)
	}

	manifest := mustReadExecutionEvidenceManifest(t, filepath.Join(h.tmp, ".namba", "logs", "runs", "spec-003-parallel-evidence.json"))
	if manifest.Status != "completed" || manifest.ExecutionMode != string(executionModeParallel) {
		t.Fatalf("expected completed parallel manifest, got %+v", manifest)
	}
	if manifest.Request.State != executionEvidenceStateNotApplicable || manifest.Validation.State != executionEvidenceStateNotApplicable {
		t.Fatalf("expected request/validation to stay not_applicable for aggregate parallel manifest, got %+v", manifest)
	}
	if manifest.Preflight.State != executionEvidenceStatePresent || manifest.Execution.State != executionEvidenceStatePresent || manifest.Progress.State != executionEvidenceStatePresent {
		t.Fatalf("expected preflight/execution/progress evidence on parallel run, got %+v", manifest)
	}
}

func TestRunParallelPreflightFailureWritesExecutionEvidenceManifest(t *testing.T) {
	h, restore := newParallelHarness(t)
	defer restore()

	err := h.app.runParallel(
		context.Background(),
		h.tmp,
		specPackage{ID: "SPEC-003"},
		[]string{"one", "two", "three"},
		"prompt",
		qualityConfig{TestCommand: "test", LintCommand: "none", TypecheckCommand: "none"},
		systemConfig{Runner: "codex"},
		codexConfig{AddDirs: []string{"missing-dir"}},
		workflowConfig{MaxParallelWorkers: 3},
		false,
	)
	if err == nil {
		t.Fatal("expected parallel preflight failure")
	}

	manifest := mustReadExecutionEvidenceManifest(t, filepath.Join(h.tmp, ".namba", "logs", "runs", "spec-003-parallel-evidence.json"))
	if manifest.Status != "preflight_failed" {
		t.Fatalf("expected preflight_failed parallel manifest, got %+v", manifest)
	}
	if manifest.Preflight.State != executionEvidenceStatePresent || manifest.Execution.State != executionEvidenceStatePresent || manifest.Progress.State != executionEvidenceStatePresent {
		t.Fatalf("expected aggregate parallel evidence on preflight failure, got %+v", manifest)
	}
}

func TestRunParallelCloseFailurePreservesCompletedStatusInExecutionEvidence(t *testing.T) {
	h, restore := newParallelHarness(t)
	defer restore()

	h.app.newParallelProgressSink = func(cfg parallelProgressSinkConfig) (parallelProgressSink, error) {
		return &stubParallelProgressSink{
			path:     cfg.Path,
			closeErr: errors.New("close denied"),
		}, nil
	}

	err := h.app.runParallel(
		context.Background(),
		h.tmp,
		specPackage{ID: "SPEC-003"},
		[]string{"one", "two", "three"},
		"prompt",
		qualityConfig{TestCommand: "test", LintCommand: "none", TypecheckCommand: "none"},
		systemConfig{Runner: "codex"},
		codexConfig{},
		workflowConfig{MaxParallelWorkers: 3},
		false,
	)
	if err == nil || !strings.Contains(err.Error(), "close denied") {
		t.Fatalf("expected close failure to surface, got %v", err)
	}

	manifest := mustReadExecutionEvidenceManifest(t, filepath.Join(h.tmp, ".namba", "logs", "runs", "spec-003-parallel-evidence.json"))
	if manifest.Status != "completed" {
		t.Fatalf("expected completed parallel evidence after close failure, got %+v", manifest)
	}
	if manifest.Finalization.ProgressLogFailed != true {
		t.Fatalf("expected finalization to preserve progress log failure separately, got %+v", manifest.Finalization)
	}
	if manifest.Execution.State != executionEvidenceStatePresent {
		t.Fatalf("expected aggregate execution artifact to remain present, got %+v", manifest.Execution)
	}
}

func TestRunParallelMergeFailureCloseFailurePreservesMergeFailedExecutionEvidence(t *testing.T) {
	h, restore := newParallelHarness(t)
	defer restore()

	h.mergeErr["namba/spec-003-p2"] = errors.New("merge conflict")
	h.app.newParallelProgressSink = func(cfg parallelProgressSinkConfig) (parallelProgressSink, error) {
		return &stubParallelProgressSink{
			path:     cfg.Path,
			closeErr: errors.New("close denied"),
		}, nil
	}

	err := h.app.runParallel(
		context.Background(),
		h.tmp,
		specPackage{ID: "SPEC-003"},
		[]string{"one", "two", "three"},
		"prompt",
		qualityConfig{TestCommand: "test", LintCommand: "none", TypecheckCommand: "none"},
		systemConfig{Runner: "codex"},
		codexConfig{},
		workflowConfig{MaxParallelWorkers: 3},
		false,
	)
	if err == nil || !strings.Contains(err.Error(), "merge conflict") || !strings.Contains(err.Error(), "close denied") {
		t.Fatalf("expected merge and close failures to surface together, got %v", err)
	}

	manifest := mustReadExecutionEvidenceManifest(t, filepath.Join(h.tmp, ".namba", "logs", "runs", "spec-003-parallel-evidence.json"))
	if manifest.Status != "merge_failed" {
		t.Fatalf("expected merge_failed parallel evidence after merge+close failure, got %+v", manifest)
	}
	if manifest.Finalization.ProgressLogFailed != true {
		t.Fatalf("expected finalization to preserve progress log failure separately, got %+v", manifest.Finalization)
	}
	if manifest.Progress.Path != ".namba/logs/runs/spec-003-parallel.events.jsonl" || manifest.Progress.State != executionEvidenceStateMissing {
		t.Fatalf("expected progress reference to keep the shared event log path, got %+v", manifest.Progress)
	}
}

func TestBuildExecutionEvidenceManifestAllowsTypedBrowserArtifacts(t *testing.T) {
	tmp := t.TempDir()
	browserPath := filepath.Join(tmp, ".namba", "logs", "runs", "spec-033-browser-trace.zip")
	writeTestFile(t, browserPath, "trace")

	manifest, err := buildExecutionEvidenceManifest(tmp, executionEvidenceOptions{
		ProjectRoot:   tmp,
		LogID:         "spec-033",
		RunID:         "spec-033",
		SpecID:        "SPEC-033",
		ExecutionMode: executionModeTeam,
		Status:        "completed",
		Request: executionEvidenceRefInput{
			Kind:          "request",
			NotApplicable: true,
		},
		Preflight: executionEvidenceRefInput{
			Kind:          "preflight",
			NotApplicable: true,
		},
		Execution: executionEvidenceRefInput{
			Kind:          "execution",
			NotApplicable: true,
		},
		Validation: executionEvidenceRefInput{
			Kind:          "validation",
			NotApplicable: true,
		},
		Progress: executionEvidenceRefInput{
			Kind:          "progress",
			NotApplicable: true,
		},
		BrowserArtifacts: []executionEvidenceRef{
			{
				Kind: "trace",
				Path: filepath.ToSlash(filepath.Join(logsDir, "runs", "spec-033-browser-trace.zip")),
			},
		},
	})
	if err != nil {
		t.Fatalf("buildExecutionEvidenceManifest failed: %v", err)
	}
	if manifest.Extensions.Browser.State != executionEvidenceStatePresent {
		t.Fatalf("expected browser extension to be present, got %+v", manifest.Extensions.Browser)
	}
	if len(manifest.Extensions.Browser.Artifacts) != 1 || manifest.Extensions.Browser.Artifacts[0].State != executionEvidenceStatePresent {
		t.Fatalf("expected present typed browser artifact, got %+v", manifest.Extensions.Browser.Artifacts)
	}
}

func TestChangeSummaryLatestExecutionProofSectionSkipsLegacyRuns(t *testing.T) {
	tmp, _, restore := prepareExecutionProject(t)
	defer restore()

	if lines := changeSummaryLatestExecutionProofSection(tmp); len(lines) != 0 {
		t.Fatalf("expected no execution-proof section without manifest, got %+v", lines)
	}
	if lines := prChecklistLatestExecutionProofItem(tmp); len(lines) != 0 {
		t.Fatalf("expected no execution-proof checklist item without manifest, got %+v", lines)
	}
}

func TestExecutionProofConsumersUseLatestAvailableManifest(t *testing.T) {
	tmp, _, restore := prepareExecutionProject(t)
	defer restore()

	olderPath := filepath.Join(tmp, ".namba", "logs", "runs", "spec-001-evidence.json")
	newerPath := filepath.Join(tmp, ".namba", "logs", "runs", "direct-fix-evidence.json")
	writeExecutionEvidenceFixture(t, olderPath, executionEvidenceManifest{
		SchemaVersion: executionEvidenceSchemaVersion,
		LogID:         "spec-001",
		RunID:         "spec-001",
		SpecID:        "SPEC-001",
		GeneratedAt:   "2026-04-22T09:00:00Z",
		ExecutionMode: string(executionModeTeam),
		Advisory:      true,
		Status:        "completed",
		Request:       executionEvidenceRef{Kind: "request", State: executionEvidenceStatePresent},
		Preflight:     executionEvidenceRef{Kind: "preflight", State: executionEvidenceStatePresent},
		Execution:     executionEvidenceRef{Kind: "execution", State: executionEvidenceStatePresent},
		Validation:    executionEvidenceRef{Kind: "validation", State: executionEvidenceStatePresent},
		Progress:      executionEvidenceRef{Kind: "progress", State: executionEvidenceStateNotApplicable},
		Extensions: executionEvidenceExtensions{
			Browser: executionEvidenceExtension{State: executionEvidenceStateNotApplicable},
			Runtime: executionEvidenceExtension{State: executionEvidenceStatePresent},
		},
	})
	writeExecutionEvidenceFixture(t, newerPath, executionEvidenceManifest{
		SchemaVersion: executionEvidenceSchemaVersion,
		LogID:         "direct-fix",
		RunID:         "direct-fix",
		SpecID:        "DIRECT-FIX",
		GeneratedAt:   "2026-04-22T10:00:00Z",
		ExecutionMode: string(executionModeDefault),
		Advisory:      true,
		Status:        "completed",
		Request:       executionEvidenceRef{Kind: "request", State: executionEvidenceStatePresent},
		Preflight:     executionEvidenceRef{Kind: "preflight", State: executionEvidenceStatePresent},
		Execution:     executionEvidenceRef{Kind: "execution", State: executionEvidenceStatePresent},
		Validation:    executionEvidenceRef{Kind: "validation", State: executionEvidenceStatePresent},
		Progress:      executionEvidenceRef{Kind: "progress", State: executionEvidenceStateNotApplicable},
		Extensions: executionEvidenceExtensions{
			Browser: executionEvidenceExtension{State: executionEvidenceStateNotApplicable},
			Runtime: executionEvidenceExtension{State: executionEvidenceStatePresent},
		},
	})

	changeSummaryLines := strings.Join(changeSummaryLatestExecutionProofSection(tmp), "\n")
	if !strings.Contains(changeSummaryLines, "direct-fix-evidence.json") || !strings.Contains(changeSummaryLines, "Proof target: `DIRECT-FIX`") {
		t.Fatalf("expected latest available proof in change summary section, got %q", changeSummaryLines)
	}
	checklistLines := strings.Join(prChecklistLatestExecutionProofItem(tmp), "\n")
	if !strings.Contains(checklistLines, "direct-fix-evidence.json") || !strings.Contains(checklistLines, "target `DIRECT-FIX`") {
		t.Fatalf("expected latest available proof in checklist item, got %q", checklistLines)
	}
}

func TestSyncOutputsSurfaceExecutionProofSeparatelyFromReadiness(t *testing.T) {
	tmp, app, restore := prepareExecutionProject(t)
	defer restore()

	app.lookPath = func(name string) (string, error) {
		switch name {
		case "codex", "git":
			return name, nil
		default:
			return "", errors.New("missing dependency")
		}
	}
	app.runCmd = func(_ context.Context, name string, args []string, dir string) (string, error) {
		switch {
		case isCodexExec(name, args):
			return "runner output", nil
		case isShellCommand(name):
			return "validation ok", nil
		default:
			t.Fatalf("unexpected command: %s %v", name, args)
			return "", nil
		}
	}

	if err := app.Run(context.Background(), []string{"run", "SPEC-001"}); err != nil {
		t.Fatalf("run failed: %v", err)
	}

	syncCtx, err := app.loadSyncContext(tmp)
	if err != nil {
		t.Fatalf("loadSyncContext failed: %v", err)
	}
	if err := app.writeSyncProjectSupportDocs(syncCtx); err != nil {
		t.Fatalf("writeSyncProjectSupportDocs failed: %v", err)
	}

	changeSummary := mustReadFile(t, filepath.Join(tmp, ".namba", "project", "change-summary.md"))
	if !strings.Contains(changeSummary, "## Latest Review Readiness") || !strings.Contains(changeSummary, "## Latest Execution Proof") {
		t.Fatalf("expected change summary to separate readiness and execution proof, got %q", changeSummary)
	}
	if !strings.Contains(changeSummary, "Execution proof status: `completed`") {
		t.Fatalf("expected execution proof status in change summary, got %q", changeSummary)
	}

	prChecklist := mustReadFile(t, filepath.Join(tmp, ".namba", "project", "pr-checklist.md"))
	if !strings.Contains(prChecklist, "Latest SPEC review readiness checked") || !strings.Contains(prChecklist, "Latest execution proof checked") {
		t.Fatalf("expected PR checklist to include readiness and execution-proof items, got %q", prChecklist)
	}
}

func mustReadExecutionEvidenceManifest(t *testing.T, path string) executionEvidenceManifest {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read execution evidence manifest: %v", err)
	}
	var manifest executionEvidenceManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		t.Fatalf("unmarshal execution evidence manifest: %v", err)
	}
	return manifest
}

func writeExecutionEvidenceFixture(t *testing.T, path string, manifest executionEvidenceManifest) {
	t.Helper()
	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		t.Fatalf("marshal execution evidence fixture: %v", err)
	}
	writeTestFile(t, path, string(data))
}
