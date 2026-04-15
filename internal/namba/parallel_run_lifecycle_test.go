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

func TestRunParallelPreflightFailureStopsBeforeSetupAndWritesReport(t *testing.T) {
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
		t.Fatal("expected preflight failure")
	}
	if !strings.Contains(err.Error(), "parallel preflight") || !strings.Contains(err.Error(), "add_dir") {
		t.Fatalf("expected parallel preflight add_dir error, got %v", err)
	}
	if countCommandsContaining(h.commands, "git worktree add -b") != 0 {
		t.Fatalf("expected no worktree setup after preflight failure, got %v", h.commands)
	}
	if h.codexCount != 0 {
		t.Fatalf("expected no worker execution after preflight failure, got %v", h.commands)
	}

	report := mustReadPreflightReport(t, filepath.Join(h.tmp, ".namba", "logs", "runs", "spec-003-parallel-preflight.json"))
	if report.Passed {
		t.Fatalf("expected failed parallel preflight report, got %+v", report)
	}
	if len(report.Steps) == 0 {
		t.Fatalf("expected preflight steps to be recorded, got %+v", report)
	}
	parallelReport := mustReadParallelReport(t, filepath.Join(h.tmp, ".namba", "logs", "runs", "spec-003-parallel.json"))
	if !parallelReport.MergeBlocked || parallelReport.FinishedAt == "" {
		t.Fatalf("expected terminal parallel report on preflight failure, got %+v", parallelReport)
	}
	if parallelReport.ProgressLogFailed {
		t.Fatalf("did not expect progress log failure on preflight error, got %+v", parallelReport)
	}

	events := mustReadParallelProgressEvents(t, filepath.Join(h.tmp, ".namba", "logs", "runs", "spec-003-parallel.events.jsonl"))
	failedIdx := indexProgressEvent(events, func(event parallelProgressRecord) bool {
		return eventString(event, "scope") == parallelProgressScopeRun &&
			eventString(event, "phase") == "failed" &&
			eventString(event, "status") == "preflight_failed"
	})
	if failedIdx == -1 {
		t.Fatalf("expected terminal preflight failure event, got %+v", events)
	}

	_, statErr := os.Stat(filepath.Join(h.tmp, ".namba", "logs", "runs", "spec-003-p1-request.md"))
	if !errors.Is(statErr, os.ErrNotExist) {
		t.Fatalf("expected no request log to be written after preflight failure, got err=%v", statErr)
	}
}

func TestRunParallelStagingFailureWritesTerminalReportAndPreservesPreparedWorkers(t *testing.T) {
	h, restore := newParallelHarness(t)
	defer restore()

	blockedWorktree := filepath.Join(h.tmp, worktreesDir, "spec-003-p2")
	h.worktreeAddErr[blockedWorktree] = errors.New("worktree add denied")

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
	if err == nil || !strings.Contains(err.Error(), "worktree add denied") {
		t.Fatalf("expected staging failure, got %v", err)
	}
	if h.codexCount != 0 {
		t.Fatalf("expected no worker execution after staging failure, got %v", h.commands)
	}

	report := mustReadParallelReport(t, filepath.Join(h.tmp, ".namba", "logs", "runs", "spec-003-parallel.json"))
	if !report.MergeBlocked || report.FinishedAt == "" {
		t.Fatalf("expected merge-blocked terminal report, got %+v", report)
	}
	if len(report.Workers) != 1 {
		t.Fatalf("expected partial staged workers in report, got %+v", report.Workers)
	}
	worker := report.Workers[0]
	if worker.Name != "spec-003-p1" || !worker.Preserved {
		t.Fatalf("expected staged worker to remain preserved, got %+v", worker)
	}

	events := mustReadParallelProgressEvents(t, filepath.Join(h.tmp, ".namba", "logs", "runs", "spec-003-parallel.events.jsonl"))
	failedIdx := indexProgressEvent(events, func(event parallelProgressRecord) bool {
		return eventString(event, "scope") == parallelProgressScopeRun &&
			eventString(event, "phase") == "failed" &&
			eventString(event, "status") == "staging_failed"
	})
	preservedIdx := indexProgressEvent(events, func(event parallelProgressRecord) bool {
		return eventString(event, "scope") == parallelProgressScopeWorker &&
			eventString(event, "worker_name", "worker_id", "worker") == "spec-003-p1" &&
			eventString(event, "status") == "preserved"
	})
	if failedIdx == -1 || preservedIdx == -1 {
		t.Fatalf("expected staging failure and preserved worker events, got %+v", events)
	}
}

func TestRunParallelDryRunKeepsPreparedWorkerMetadataAndRequestLogs(t *testing.T) {
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
		true,
	)
	if err != nil {
		t.Fatalf("runParallel failed: %v", err)
	}

	report := mustReadParallelReport(t, filepath.Join(h.tmp, ".namba", "logs", "runs", "spec-003-parallel.json"))
	if !report.DryRun || report.MergeBlocked || report.PruneAttempted {
		t.Fatalf("expected dry-run report without merge or prune, got %+v", report)
	}
	if len(report.Workers) != 3 {
		t.Fatalf("expected three prepared workers, got %+v", report.Workers)
	}
	for _, worker := range report.Workers {
		if worker.Name == "" || worker.Branch == "" || worker.Worktree == "" {
			t.Fatalf("expected prepared worker metadata, got %+v", worker)
		}
		if worker.SessionID != "" || worker.ExecutionPassed || worker.ValidationPassed || worker.MergeAttempted || worker.CleanupAttempted || worker.Preserved {
			t.Fatalf("expected dry-run worker to remain unexecuted, got %+v", worker)
		}
		if _, err := os.Stat(worker.Worktree); err != nil {
			t.Fatalf("expected prepared worktree directory for %s: %v", worker.Name, err)
		}
		logPath := filepath.Join(h.tmp, ".namba", "logs", "runs", worker.Name+"-request.md")
		logBytes, err := os.ReadFile(logPath)
		if err != nil {
			t.Fatalf("expected request log for %s: %v", worker.Name, err)
		}
		logText := string(logBytes)
		if !strings.Contains(logText, "prompt") || !strings.Contains(logText, "## Assigned work package") {
			t.Fatalf("expected shaped worker prompt in request log for %s, got %q", worker.Name, logText)
		}
	}
}

func TestRunParallelDryRunWritesChunkedWorkerPromptContent(t *testing.T) {
	h, restore := newParallelHarness(t)
	defer restore()

	err := h.app.runParallel(
		context.Background(),
		h.tmp,
		specPackage{ID: "SPEC-003"},
		[]string{"one", "two", "three", "four"},
		"prompt",
		qualityConfig{TestCommand: "test", LintCommand: "none", TypecheckCommand: "none"},
		systemConfig{Runner: "codex"},
		codexConfig{},
		workflowConfig{MaxParallelWorkers: 3},
		true,
	)
	if err != nil {
		t.Fatalf("runParallel failed: %v", err)
	}

	request := mustReadTextFile(t, filepath.Join(h.tmp, ".namba", "logs", "runs", "spec-003-p1-request.md"))
	if !strings.Contains(request, "prompt") || !strings.Contains(request, "## Assigned work package") {
		t.Fatalf("expected worker prompt envelope, got %q", request)
	}
	for _, want := range []string{"one", "four"} {
		if !strings.Contains(request, want) {
			t.Fatalf("expected first worker chunk to include %q, got %q", want, request)
		}
	}
	for _, unwanted := range []string{"two", "three"} {
		if strings.Contains(request, unwanted) {
			t.Fatalf("expected first worker chunk to exclude %q, got %q", unwanted, request)
		}
	}
}

func TestRunParallelExecutionWritesWorkerRequestWithParallelRuntimeShape(t *testing.T) {
	h, restore := newParallelHarness(t)
	defer restore()

	err := h.app.runParallel(
		context.Background(),
		h.tmp,
		specPackage{ID: "SPEC-003"},
		[]string{"one", "two", "three", "four"},
		"prompt",
		qualityConfig{TestCommand: "test", LintCommand: "none", TypecheckCommand: "none"},
		systemConfig{Runner: "codex"},
		codexConfig{Profile: "namba", SessionMode: "stateful", RepairAttempts: 1},
		workflowConfig{MaxParallelWorkers: 3},
		false,
	)
	if err != nil {
		t.Fatalf("runParallel failed: %v", err)
	}

	request := mustReadParallelExecutionRequest(t, filepath.Join(h.tmp, ".namba", "logs", "runs", "spec-003-p2-request.json"))
	if request.Mode != executionModeParallel || request.SessionMode != "parallel-worker" {
		t.Fatalf("expected parallel worker runtime shape, got %+v", request)
	}
	expectedWorktree := filepath.Join(h.tmp, worktreesDir, "spec-003-p2")
	if request.WorkDir != expectedWorktree {
		t.Fatalf("expected worker request to target %q, got %q", expectedWorktree, request.WorkDir)
	}
	if !strings.Contains(request.Prompt, "prompt") || !strings.Contains(request.Prompt, "## Assigned work package") {
		t.Fatalf("expected worker prompt envelope, got %q", request.Prompt)
	}
	if !strings.Contains(request.Prompt, "two") {
		t.Fatalf("expected second worker chunk to include %q, got %q", "two", request.Prompt)
	}
	for _, unwanted := range []string{"one", "three", "four"} {
		if strings.Contains(request.Prompt, unwanted) {
			t.Fatalf("expected second worker chunk to exclude %q, got %q", unwanted, request.Prompt)
		}
	}
}

func TestRunParallelSuccessReportCapturesExecutionMergeAndCleanupState(t *testing.T) {
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

	report := mustReadParallelReport(t, filepath.Join(h.tmp, ".namba", "logs", "runs", "spec-003-parallel.json"))
	if report.MergeBlocked || !report.PruneAttempted || report.PruneError != "" {
		t.Fatalf("expected successful merge and prune report, got %+v", report)
	}
	for _, worker := range report.Workers {
		if worker.SessionID != worker.Name {
			t.Fatalf("expected worker session to follow log id, got %+v", worker)
		}
		if worker.StartedAt == "" || worker.FinishedAt == "" {
			t.Fatalf("expected worker timestamps to be recorded, got %+v", worker)
		}
		if !worker.ExecutionPassed || !worker.ValidationPassed || !worker.MergeAttempted || !worker.MergeSucceeded || !worker.CleanupAttempted || !worker.WorktreeRemoved || !worker.BranchRemoved {
			t.Fatalf("expected full lifecycle success state, got %+v", worker)
		}
		if worker.ExecutionError != "" || worker.ValidationError != "" || worker.MergeError != "" || worker.CleanupError != "" || worker.Preserved {
			t.Fatalf("expected clean worker report, got %+v", worker)
		}
	}
}

func TestParallelRunLifecycleBlockMergeOnWorkerFailuresPreservesWorkersWithoutWritingReport(t *testing.T) {
	h, restore := newParallelHarness(t)
	defer restore()

	lifecycle := &parallelRunLifecycle{
		app:  h.app,
		ctx:  context.Background(),
		root: h.tmp,
		plan: parallelRunPlan{Report: parallelRunReport{SpecID: "SPEC-003"}},
		results: []parallelWorkerResult{
			{Name: "spec-003-p1", Branch: "namba/spec-003-p1", ExecutionPassed: true, ValidationPassed: true},
			{Name: "spec-003-p2", Branch: "namba/spec-003-p2", ExecutionPassed: false, ValidationPassed: true, ExecutionError: "runner boom"},
		},
	}

	err := lifecycle.blockMergeOnWorkerFailures()
	if err == nil || !strings.Contains(err.Error(), "blocked merge") {
		t.Fatalf("expected blocked merge error, got %v", err)
	}
	for _, worker := range lifecycle.results {
		if !worker.Preserved {
			t.Fatalf("expected worker to be preserved after merge block, got %+v", worker)
		}
	}
	if countCommandsContaining(h.commands, "git merge --no-ff") != 0 {
		t.Fatalf("expected helper to stop before merge commands, got %v", h.commands)
	}
	_, statErr := os.Stat(filepath.Join(h.tmp, ".namba", "logs", "runs", "spec-003-parallel.json"))
	if !errors.Is(statErr, os.ErrNotExist) {
		t.Fatalf("expected no persisted report from merge-block helper, got err=%v", statErr)
	}
}

func TestParallelRunLifecycleMergeWorkersPreservesWorkersOnFailureWithoutWritingReport(t *testing.T) {
	h, restore := newParallelHarness(t)
	defer restore()

	h.mergeErr["namba/spec-003-p2"] = errors.New("merge conflict")
	lifecycle := &parallelRunLifecycle{
		app:  h.app,
		ctx:  context.Background(),
		root: h.tmp,
		plan: parallelRunPlan{Report: parallelRunReport{SpecID: "SPEC-003"}},
		results: []parallelWorkerResult{
			{Name: "spec-003-p1", Branch: "namba/spec-003-p1"},
			{Name: "spec-003-p2", Branch: "namba/spec-003-p2"},
		},
	}

	err := lifecycle.mergeWorkers()
	if err == nil || !strings.Contains(err.Error(), "merge namba/spec-003-p2") {
		t.Fatalf("expected merge failure error, got %v", err)
	}
	if !lifecycle.results[0].MergeAttempted || !lifecycle.results[0].MergeSucceeded {
		t.Fatalf("expected first worker merge success before failure, got %+v", lifecycle.results[0])
	}
	if !lifecycle.results[1].MergeAttempted || lifecycle.results[1].MergeSucceeded || lifecycle.results[1].MergeError == "" {
		t.Fatalf("expected second worker merge failure state, got %+v", lifecycle.results[1])
	}
	for _, worker := range lifecycle.results {
		if !worker.Preserved {
			t.Fatalf("expected all workers to remain preserved after merge failure, got %+v", worker)
		}
	}
	_, statErr := os.Stat(filepath.Join(h.tmp, ".namba", "logs", "runs", "spec-003-parallel.json"))
	if !errors.Is(statErr, os.ErrNotExist) {
		t.Fatalf("expected no persisted report from merge helper, got err=%v", statErr)
	}
}

func TestParallelRunLifecycleCleanupMergedWorkersReturnsWorkerCleanupOutcomeWithoutPrune(t *testing.T) {
	h, restore := newParallelHarness(t)
	defer restore()

	firstWorktree := filepath.Join(h.tmp, worktreesDir, "spec-003-p1")
	h.removeErr[firstWorktree] = errors.New("locked worktree")
	h.pruneErr = errors.New("prune denied")
	lifecycle := &parallelRunLifecycle{
		app:  h.app,
		ctx:  context.Background(),
		root: h.tmp,
		plan: parallelRunPlan{Report: parallelRunReport{SpecID: "SPEC-003"}},
		results: []parallelWorkerResult{
			{Name: "spec-003-p1", Branch: "namba/spec-003-p1", Worktree: firstWorktree},
			{Name: "spec-003-p2", Branch: "namba/spec-003-p2", Worktree: filepath.Join(h.tmp, worktreesDir, "spec-003-p2")},
		},
	}

	outcome := lifecycle.cleanupMergedWorkers()
	if outcome.PruneError != "" {
		t.Fatalf("expected cleanupMergedWorkers to exclude prune state, got %+v", outcome)
	}
	if len(outcome.Failures) != 1 {
		t.Fatalf("expected only worker cleanup failures, got %v", outcome.Failures)
	}
	if !lifecycle.results[0].CleanupAttempted || lifecycle.results[0].CleanupError == "" || lifecycle.results[0].WorktreeRemoved {
		t.Fatalf("expected first worker cleanup failure state, got %+v", lifecycle.results[0])
	}
	if !lifecycle.results[1].CleanupAttempted || !lifecycle.results[1].WorktreeRemoved || !lifecycle.results[1].BranchRemoved {
		t.Fatalf("expected second worker cleanup success state, got %+v", lifecycle.results[1])
	}
	if countCommandsContaining(h.commands, "git worktree prune") != 0 {
		t.Fatalf("expected cleanupMergedWorkers to stop before prune, got %v", h.commands)
	}
}

func TestReportWithParallelCleanupOutcomeRecordsPruneState(t *testing.T) {
	report := reportWithParallelCleanupOutcome(
		parallelRunReport{SpecID: "SPEC-003"},
		parallelCleanupOutcome{PruneError: "prune denied"},
	)

	if !report.PruneAttempted || report.PruneError != "prune denied" {
		t.Fatalf("expected cleanup outcome to update prune report state, got %+v", report)
	}
}

func TestApplyParallelPruneErrorAggregatesWorkerAndPruneFailures(t *testing.T) {
	outcome := applyParallelPruneError(
		parallelCleanupOutcome{Failures: []string{"remove worktree /repo/.namba/worktrees/spec-003-p1: locked worktree"}},
		"prune denied",
	)

	if outcome.PruneError != "prune denied" {
		t.Fatalf("expected prune error to be surfaced, got %+v", outcome)
	}
	if len(outcome.Failures) != 2 {
		t.Fatalf("expected cleanup and prune failures, got %v", outcome.Failures)
	}
	if !strings.Contains(strings.Join(outcome.Failures, "; "), "worktree prune: prune denied") {
		t.Fatalf("expected prune failure in cleanup outcome, got %v", outcome.Failures)
	}
}

func TestParallelRunLifecycleCleanupParallelWorkerArtifactsOutcomeExcludesPruneState(t *testing.T) {
	h, restore := newParallelHarness(t)
	defer restore()

	worktree := filepath.Join(h.tmp, worktreesDir, "spec-003-p1")
	h.removeErr[worktree] = errors.New("locked worktree")
	lifecycle := &parallelRunLifecycle{
		app:  h.app,
		ctx:  context.Background(),
		root: h.tmp,
		results: []parallelWorkerResult{
			{Name: "spec-003-p1", Branch: "namba/spec-003-p1", Worktree: worktree},
		},
	}

	outcome := lifecycle.cleanupParallelWorkerArtifactsOutcome()
	if outcome.PruneError != "" {
		t.Fatalf("expected worker cleanup outcome to exclude prune state, got %+v", outcome)
	}
	if len(outcome.Failures) != 1 || !strings.Contains(outcome.Failures[0], "remove worktree") {
		t.Fatalf("expected worker cleanup failures only, got %+v", outcome)
	}
}

func TestParallelRunLifecyclePruneParallelWorktreesReturnsErrorString(t *testing.T) {
	h, restore := newParallelHarness(t)
	defer restore()

	h.pruneErr = errors.New("prune denied")
	lifecycle := &parallelRunLifecycle{
		app:  h.app,
		ctx:  context.Background(),
		root: h.tmp,
	}

	if got := lifecycle.pruneParallelWorktrees(); got != "prune denied" {
		t.Fatalf("expected prune error string, got %q", got)
	}
}

func TestApplyParallelPruneErrorAppendsPruneFailureToCleanupOutcome(t *testing.T) {
	outcome := applyParallelPruneError(parallelCleanupOutcome{Failures: []string{"remove worktree failed"}}, "prune denied")
	if outcome.PruneError != "prune denied" {
		t.Fatalf("expected prune error to be recorded, got %+v", outcome)
	}
	if len(outcome.Failures) != 2 || outcome.Failures[1] != "worktree prune: prune denied" {
		t.Fatalf("expected prune failure to append to cleanup outcome, got %+v", outcome)
	}
}

func TestParallelRunLifecycleWriteFinishedReportSnapshotsLatestWorkerState(t *testing.T) {
	h, restore := newParallelHarness(t)
	defer restore()

	lifecycle := &parallelRunLifecycle{
		app:  h.app,
		ctx:  context.Background(),
		root: h.tmp,
		plan: parallelRunPlan{Report: parallelRunReport{SpecID: "SPEC-003", CleanupPolicy: "policy"}},
		results: []parallelWorkerResult{
			{Name: "spec-003-p1", MergeSucceeded: true},
		},
	}

	report := parallelRunReport{SpecID: "SPEC-003", CleanupPolicy: "policy", MergeBlocked: true}
	if err := lifecycle.writeFinishedReport(report); err != nil {
		t.Fatalf("writeFinishedReport failed: %v", err)
	}

	written := mustReadParallelReport(t, filepath.Join(h.tmp, ".namba", "logs", "runs", "spec-003-parallel.json"))
	if !written.MergeBlocked || written.FinishedAt == "" {
		t.Fatalf("expected finished report metadata, got %+v", written)
	}
	if len(written.Workers) != 1 || written.Workers[0].Name != "spec-003-p1" || !written.Workers[0].MergeSucceeded {
		t.Fatalf("expected latest worker snapshot in persisted report, got %+v", written.Workers)
	}
}

func TestParallelRunLifecycleFinishedReportSnapshotAddsWorkerStateAndTimestamp(t *testing.T) {
	h, restore := newParallelHarness(t)
	defer restore()

	lifecycle := &parallelRunLifecycle{
		app:  h.app,
		ctx:  context.Background(),
		root: h.tmp,
		plan: parallelRunPlan{Report: parallelRunReport{SpecID: "SPEC-003", CleanupPolicy: "policy"}},
		results: []parallelWorkerResult{
			{Name: "spec-003-p1", MergeSucceeded: true},
		},
	}

	snapshot := lifecycle.finishedReportSnapshot(parallelRunReport{SpecID: "SPEC-003", CleanupPolicy: "policy", MergeBlocked: true})
	if !snapshot.MergeBlocked || snapshot.FinishedAt == "" {
		t.Fatalf("expected finished report snapshot metadata, got %+v", snapshot)
	}
	if len(snapshot.Workers) != 1 || snapshot.Workers[0].Name != "spec-003-p1" || !snapshot.Workers[0].MergeSucceeded {
		t.Fatalf("expected latest worker snapshot, got %+v", snapshot.Workers)
	}
}

func TestBuildParallelWorkerPromptShapesAssignedWorkPackageSection(t *testing.T) {
	got := buildParallelWorkerPrompt("base prompt", []string{"task one", "task two"})
	want := "base prompt\n\n## Assigned work package\n\ntask one\ntask two"
	if got != want {
		t.Fatalf("expected shaped worker prompt %q, got %q", want, got)
	}
}

func TestApplyParallelWorkerOutcomeCopiesExecutionAndValidationState(t *testing.T) {
	base := parallelWorkerResult{Name: "spec-003-p1", Branch: "namba/spec-003-p1"}
	outcome := parallelWorkerOutcome{
		Index: 0,
		ExecutionResult: executionResult{
			SessionID:  "spec-003-p1",
			RetryCount: 2,
			StartedAt:  "2026-04-10T00:00:00Z",
			FinishedAt: "2026-04-10T00:01:00Z",
			Turns: []executionTurnResult{
				{Name: "implement", Succeeded: true},
			},
		},
		ValidationReport: validationReport{Passed: true},
	}

	got := applyParallelWorkerOutcome(base, outcome)
	if got.SessionID != "spec-003-p1" || got.RetryCount != 2 {
		t.Fatalf("expected execution metadata to copy onto worker result, got %+v", got)
	}
	if got.StartedAt != "2026-04-10T00:00:00Z" || got.FinishedAt != "2026-04-10T00:01:00Z" {
		t.Fatalf("expected worker timestamps to copy onto worker result, got %+v", got)
	}
	if !got.ExecutionPassed || !got.ValidationPassed || got.ExecutionError != "" || got.ValidationError != "" {
		t.Fatalf("expected successful outcome to stay clean, got %+v", got)
	}
}

func TestClassifyParallelWorkerFailureSeparatesExecutionAndValidationErrors(t *testing.T) {
	t.Run("execution failure", func(t *testing.T) {
		executionErr, validationErr := classifyParallelWorkerFailure(
			executionResult{Succeeded: false, Error: "runner boom"},
			validationReport{},
			errors.New("runner boom"),
		)
		if executionErr != "runner boom" || validationErr != "" {
			t.Fatalf("expected execution failure classification, got execution=%q validation=%q", executionErr, validationErr)
		}
	})

	t.Run("validation failure", func(t *testing.T) {
		executionErr, validationErr := classifyParallelWorkerFailure(
			executionResult{Succeeded: true},
			validationReport{Passed: false, Steps: []validationStep{{Name: "lint", Error: "failed"}}},
			errors.New("validation failed"),
		)
		if executionErr != "" || validationErr != "lint: failed" {
			t.Fatalf("expected validation failure classification, got execution=%q validation=%q", executionErr, validationErr)
		}
	})
}

func TestParallelRunLifecycleNewParallelWorkerExecutionBuildsRequestFromWorkerState(t *testing.T) {
	h, restore := newParallelHarness(t)
	defer restore()

	lifecycle := &parallelRunLifecycle{
		app:       h.app,
		ctx:       context.Background(),
		root:      h.tmp,
		specPkg:   specPackage{ID: "SPEC-003"},
		prompt:    "base prompt",
		systemCfg: systemConfig{Runner: "codex", ApprovalPolicy: "on-request", SandboxMode: "workspace-write"},
		codexCfg:  codexConfig{Profile: "namba", SessionMode: "stateful", RepairAttempts: 1},
		results: []parallelWorkerResult{
			{Name: "spec-003-p1", Branch: "namba/spec-003-p1", Worktree: filepath.Join(h.tmp, worktreesDir, "spec-003-p1")},
		},
	}

	worker := lifecycle.newParallelWorkerExecution(0, []string{"task one", "task two"})

	if worker.Result.Name != "spec-003-p1" || worker.Result.Worktree == "" {
		t.Fatalf("expected worker metadata from lifecycle state, got %+v", worker.Result)
	}
	if worker.Prompt != "base prompt\n\n## Assigned work package\n\ntask one\ntask two" {
		t.Fatalf("expected shaped worker prompt, got %q", worker.Prompt)
	}
	if worker.Request.SpecID != "SPEC-003" || worker.Request.WorkDir != worker.Result.Worktree {
		t.Fatalf("expected request to target worker worktree, got %+v", worker.Request)
	}
	if worker.Request.Mode != executionModeParallel || worker.Request.Prompt != worker.Prompt {
		t.Fatalf("expected parallel execution request with shaped prompt, got %+v", worker.Request)
	}
	if worker.Request.SessionMode != "parallel-worker" {
		t.Fatalf("expected parallel worker session mode, got %+v", worker.Request)
	}
	expectedDelegation := suggestDelegationPlan(executionModeParallel, worker.Prompt, "", "")
	if worker.Request.DelegationPlan.IntegratorRole != expectedDelegation.IntegratorRole || worker.Request.DelegationPlan.DelegationBudget != expectedDelegation.DelegationBudget {
		t.Fatalf("expected worker request delegation plan to follow prompt shaping, got %+v", worker.Request.DelegationPlan)
	}
}

func TestParallelRunLifecycleExecuteParallelWorkerReturnsOutcomeWithoutMutatingResults(t *testing.T) {
	h, restore := newParallelHarness(t)
	defer restore()

	worktree := filepath.Join(h.tmp, worktreesDir, "spec-003-p1")
	if err := os.MkdirAll(worktree, 0o755); err != nil {
		t.Fatalf("mkdir worktree: %v", err)
	}

	lifecycle := &parallelRunLifecycle{
		app:        h.app,
		ctx:        context.Background(),
		root:       h.tmp,
		specPkg:    specPackage{ID: "SPEC-003"},
		prompt:     "base prompt",
		qualityCfg: qualityConfig{TestCommand: "test", LintCommand: "none", TypecheckCommand: "none"},
		systemCfg:  systemConfig{Runner: "codex", ApprovalPolicy: "on-request", SandboxMode: "workspace-write"},
		codexCfg:   codexConfig{SessionMode: "stateful"},
		results: []parallelWorkerResult{
			{Name: "spec-003-p1", Branch: "namba/spec-003-p1", Worktree: worktree},
		},
	}

	worker := lifecycle.newParallelWorkerExecution(0, []string{"task one"})
	outcome := lifecycle.executeParallelWorker(worker)

	if outcome.Index != 0 {
		t.Fatalf("expected worker outcome index 0, got %+v", outcome)
	}
	if outcome.ExecutionResult.SessionID != "spec-003-p1" || !outcome.ValidationReport.Passed || outcome.RunErr != nil {
		t.Fatalf("expected successful worker outcome, got %+v", outcome)
	}
	if lifecycle.results[0].SessionID != "" || lifecycle.results[0].ExecutionPassed || lifecycle.results[0].ValidationPassed {
		t.Fatalf("expected executeParallelWorker to leave shared results untouched, got %+v", lifecycle.results[0])
	}
}

func TestParallelRunLifecycleRecordParallelWorkerOutcomeTargetsNamedWorker(t *testing.T) {
	lifecycle := &parallelRunLifecycle{
		results: []parallelWorkerResult{
			{Name: "spec-003-p1"},
			{Name: "spec-003-p2"},
		},
	}

	lifecycle.recordParallelWorkerOutcome(parallelWorkerOutcome{
		Index: 1,
		ExecutionResult: executionResult{
			SessionID:  "session-2",
			RetryCount: 1,
			StartedAt:  "2026-04-10T00:00:00Z",
			FinishedAt: "2026-04-10T00:01:00Z",
			Succeeded:  false,
			Error:      "runner boom",
			Turns:      []executionTurnResult{{Succeeded: false}},
		},
		ValidationReport: validationReport{},
		RunErr:           errors.New("runner boom"),
	})

	if lifecycle.results[0].SessionID != "" || lifecycle.results[0].ExecutionError != "" {
		t.Fatalf("expected first worker to remain unchanged, got %+v", lifecycle.results[0])
	}
	if lifecycle.results[1].SessionID != "session-2" || lifecycle.results[1].RetryCount != 1 || lifecycle.results[1].ExecutionError == "" {
		t.Fatalf("expected second worker execution state to be recorded, got %+v", lifecycle.results[1])
	}
}

func TestParallelRunLifecycleRecordParallelWorkerOutcomeClassifiesValidationFailure(t *testing.T) {
	lifecycle := &parallelRunLifecycle{
		results: []parallelWorkerResult{
			{Name: "spec-003-p1"},
		},
	}

	lifecycle.recordParallelWorkerOutcome(parallelWorkerOutcome{
		Index: 0,
		ExecutionResult: executionResult{
			SessionID:     "spec-003-p1",
			RetryCount:    1,
			StartedAt:     "started",
			FinishedAt:    "finished",
			Succeeded:     true,
			Turns:         []executionTurnResult{{Succeeded: true}},
			Output:        "ok",
			WorkDir:       "/tmp/spec-003-p1",
			ExecutionMode: string(executionModeParallel),
		},
		ValidationReport: validationReport{
			Passed: false,
			Steps:  []validationStep{{Name: "lint", Error: "failed"}},
		},
		RunErr: errors.New("lint failed"),
	})

	recorded := lifecycle.results[0]
	if recorded.SessionID != "spec-003-p1" || recorded.RetryCount != 1 || recorded.StartedAt != "started" || recorded.FinishedAt != "finished" {
		t.Fatalf("expected lifecycle to record execution metadata, got %+v", recorded)
	}
	if !recorded.ExecutionPassed || recorded.ValidationPassed {
		t.Fatalf("expected execution pass and validation failure state, got %+v", recorded)
	}
	if recorded.ValidationError != "lint: failed" || recorded.ExecutionError != "" {
		t.Fatalf("expected validation error classification, got %+v", recorded)
	}
}

func TestParallelRunLifecycleCleanupMergedWorkersAggregatesPruneFailures(t *testing.T) {
	h, restore := newParallelHarness(t)
	defer restore()

	worktree := filepath.Join(h.tmp, worktreesDir, "spec-006-p1")
	h.removeErr[worktree] = errors.New("locked worktree")
	h.pruneErr = errors.New("prune denied")
	lifecycle := &parallelRunLifecycle{
		app:  h.app,
		ctx:  context.Background(),
		root: h.tmp,
		plan: parallelRunPlan{
			Report: parallelRunReport{
				SpecID:        "SPEC-006",
				CleanupPolicy: "policy",
			},
		},
		results: []parallelWorkerResult{
			{Name: "spec-006-p1", Branch: "namba/spec-006-p1", Worktree: worktree, MergeSucceeded: true},
		},
	}

	outcome := lifecycle.cleanupMergedWorkers()
	outcome = applyParallelPruneError(outcome, lifecycle.pruneParallelWorktrees())
	report := reportWithParallelCleanupOutcome(lifecycle.report(), outcome)

	if !report.PruneAttempted || report.PruneError != "prune denied" {
		t.Fatalf("expected prune failure to be recorded in report, got %+v", report)
	}
	if len(outcome.Failures) != 2 {
		t.Fatalf("expected cleanup and prune failures, got %v", outcome.Failures)
	}
	if !strings.Contains(strings.Join(outcome.Failures, "; "), "worktree prune: prune denied") {
		t.Fatalf("expected prune failure in cleanup failures, got %v", outcome.Failures)
	}
	if !lifecycle.results[0].CleanupAttempted {
		t.Fatalf("expected worker cleanup attempt to be recorded, got %+v", lifecycle.results[0])
	}
}

func TestParallelRunLifecycleCompleteCleanupPhaseBuildsCleanupReportWithoutPersisting(t *testing.T) {
	h, restore := newParallelHarness(t)
	defer restore()

	worktree := filepath.Join(h.tmp, worktreesDir, "spec-006-p1")
	h.removeErr[worktree] = errors.New("locked worktree")
	h.pruneErr = errors.New("prune denied")
	lifecycle := &parallelRunLifecycle{
		app:  h.app,
		ctx:  context.Background(),
		root: h.tmp,
		plan: parallelRunPlan{
			Report: parallelRunReport{
				SpecID:        "SPEC-006",
				CleanupPolicy: "policy",
			},
		},
		results: []parallelWorkerResult{
			{Name: "spec-006-p1", Branch: "namba/spec-006-p1", Worktree: worktree, MergeSucceeded: true},
		},
	}

	report, outcome := lifecycle.completeCleanupPhase()

	if !report.PruneAttempted || report.PruneError != "prune denied" {
		t.Fatalf("expected cleanup phase report to record prune state, got %+v", report)
	}
	if len(outcome.Failures) != 2 {
		t.Fatalf("expected cleanup and prune failures, got %v", outcome.Failures)
	}
	if !strings.Contains(strings.Join(outcome.Failures, "; "), "worktree prune: prune denied") {
		t.Fatalf("expected prune failure in cleanup outcome, got %v", outcome.Failures)
	}
	if !lifecycle.results[0].CleanupAttempted {
		t.Fatalf("expected worker cleanup attempt to be recorded, got %+v", lifecycle.results[0])
	}
	_, statErr := os.Stat(filepath.Join(h.tmp, ".namba", "logs", "runs", "spec-006-parallel.json"))
	if !errors.Is(statErr, os.ErrNotExist) {
		t.Fatalf("expected cleanup phase helper to avoid persisting report directly, got err=%v", statErr)
	}
}

func TestCleanupParallelWorkerAggregatesBoundaryFailures(t *testing.T) {
	h, restore := newParallelHarness(t)
	defer restore()

	result := parallelWorkerResult{
		Name:     "spec-003-p1",
		Branch:   "namba/spec-003-p1",
		Worktree: filepath.Join(h.tmp, worktreesDir, "spec-003-p1"),
	}
	h.removeErr[result.Worktree] = errors.New("locked worktree")
	h.branchErr[result.Branch] = errors.New("branch delete denied")

	failures := h.app.cleanupParallelWorker(context.Background(), h.tmp, &result)

	if !result.CleanupAttempted {
		t.Fatalf("expected cleanup attempt to be recorded, got %+v", result)
	}
	if result.WorktreeRemoved || result.BranchRemoved {
		t.Fatalf("expected failed cleanup to keep removed flags false, got %+v", result)
	}
	if len(failures) != 2 {
		t.Fatalf("expected both cleanup failures to be returned, got %v", failures)
	}
	if !strings.Contains(result.CleanupError, "remove worktree: locked worktree") || !strings.Contains(result.CleanupError, "delete branch: branch delete denied") {
		t.Fatalf("expected aggregated cleanup error, got %+v", result)
	}
}

func TestParallelRunLifecycleHelpersFailLoudly(t *testing.T) {
	results := []parallelWorkerResult{
		{Name: "spec-003-p1", ExecutionPassed: true, ValidationPassed: true},
		{Name: "spec-003-p2", ExecutionPassed: false, ValidationPassed: true, ExecutionError: "runner boom"},
		{Name: "spec-003-p3", ExecutionPassed: true, ValidationPassed: false, ValidationError: "lint: failed"},
	}
	if !hasParallelRunFailures(results) {
		t.Fatal("expected failure helper to detect worker failures")
	}

	summary := summarizeParallelRunFailures(results)
	for _, want := range []string{
		"spec-003-p2 execution failed: runner boom",
		"spec-003-p3 validation failed: lint: failed",
	} {
		if !strings.Contains(summary, want) {
			t.Fatalf("expected summary to contain %q, got %q", want, summary)
		}
	}

	if got := summarizeParallelRunFailures([]parallelWorkerResult{{Name: "spec-003-p1", ExecutionPassed: false, ValidationPassed: false}}); got != "parallel workers did not all pass" {
		t.Fatalf("expected generic fallback summary, got %q", got)
	}
	if got := validationFailureMessage(validationReport{Steps: []validationStep{{Name: "lint", Error: "failed"}}}, errors.New("fallback")); got != "lint: failed" {
		t.Fatalf("expected named validation failure, got %q", got)
	}
	if got := validationFailureMessage(validationReport{Steps: []validationStep{{Error: "plain failure"}}}, nil); got != "plain failure" {
		t.Fatalf("expected unnamed validation failure, got %q", got)
	}
	if got := validationFailureMessage(validationReport{}, errors.New("fallback")); got != "fallback" {
		t.Fatalf("expected fallback validation failure, got %q", got)
	}
	if got := validationFailureMessage(validationReport{}, nil); got != "validation failed" {
		t.Fatalf("expected default validation failure, got %q", got)
	}
	if got := appendCleanupError("remove worktree: locked", "delete branch: denied"); got != "remove worktree: locked; delete branch: denied" {
		t.Fatalf("expected cleanup errors to append with separator, got %q", got)
	}
	if got := firstNonEmptyString("  ", "\n", " chosen ", "later"); got != "chosen" {
		t.Fatalf("expected first non-empty string to be trimmed, got %q", got)
	}
}

func mustReadParallelExecutionRequest(t *testing.T, path string) executionRequest {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read execution request: %v", err)
	}
	var req executionRequest
	if err := json.Unmarshal(data, &req); err != nil {
		t.Fatalf("unmarshal execution request: %v", err)
	}
	return req
}

func mustReadTextFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read text file: %v", err)
	}
	return string(data)
}
