package namba

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

type parallelRunPlan struct {
	BaseBranch   string
	Chunks       [][]string
	LogDir       string
	RunID        string
	EventLogPath string
	Report       parallelRunReport
}

type parallelWorkerExecution struct {
	Index   int
	Result  parallelWorkerResult
	Prompt  string
	Request executionRequest
}

type parallelWorkerOutcome struct {
	Index            int
	ExecutionResult  executionResult
	ValidationReport validationReport
	RunErr           error
}

type parallelCleanupOutcome struct {
	Failures   []string
	PruneError string
}

type parallelRunLifecycle struct {
	app              *App
	ctx              context.Context
	root             string
	specPkg          specPackage
	tasks            []string
	prompt           string
	qualityCfg       qualityConfig
	systemCfg        systemConfig
	codexCfg         codexConfig
	workflowCfg      workflowConfig
	dryRun           bool
	plan             parallelRunPlan
	results          []parallelWorkerResult
	progress         parallelProgressSink
	progressErr      error
	progressCloseErr error
	progressClosed   bool
}

func (a *App) newParallelRunLifecycle(ctx context.Context, root string, specPkg specPackage, tasks []string, prompt string, qualityCfg qualityConfig, systemCfg systemConfig, codexCfg codexConfig, workflowCfg workflowConfig, dryRun bool) (*parallelRunLifecycle, error) {
	return &parallelRunLifecycle{
		app:         a,
		ctx:         ctx,
		root:        root,
		specPkg:     specPkg,
		tasks:       append([]string(nil), tasks...),
		prompt:      prompt,
		qualityCfg:  qualityCfg,
		systemCfg:   systemCfg,
		codexCfg:    codexCfg,
		workflowCfg: workflowCfg,
		dryRun:      dryRun,
	}, nil
}

func (a *App) prepareParallelRunPlan(ctx context.Context, root string, specPkg specPackage, tasks []string, workflowCfg workflowConfig, dryRun bool) (parallelRunPlan, error) {
	if _, err := a.lookPath("git"); err != nil {
		return parallelRunPlan{}, errors.New("parallel execution requires git")
	}
	if !isGitRepository(root) {
		return parallelRunPlan{}, errors.New("parallel execution requires a git repository")
	}

	baseBranch, err := a.currentBranch(ctx, root)
	if err != nil {
		return parallelRunPlan{}, err
	}

	workers := minInt(len(tasks), maxInt(workflowCfg.MaxParallelWorkers, 1))
	if workers == 0 {
		workers = 1
	}
	chunks := chunkTasks(tasks, workers)
	logDir := filepath.Join(root, logsDir, "runs")
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		return parallelRunPlan{}, err
	}

	runID := newParallelRunID(specPkg.ID, a.now())
	eventLogPath := parallelProgressLogPath(root, specPkg.ID)
	return parallelRunPlan{
		BaseBranch:   baseBranch,
		Chunks:       chunks,
		LogDir:       logDir,
		RunID:        runID,
		EventLogPath: eventLogPath,
		Report: parallelRunReport{
			SpecID:        specPkg.ID,
			RunID:         runID,
			BaseBranch:    baseBranch,
			DryRun:        dryRun,
			CleanupPolicy: "Success removes temporary worktrees and deletes worker branches after every merge succeeds. Any execution, validation, or merge failure preserves all worker worktrees and branches for inspection.",
			EventLogPath:  relativeParallelProgressLogPath(specPkg.ID),
			StartedAt:     a.now().Format(time.RFC3339),
		},
	}, nil
}

func (a *App) stageParallelWorkers(ctx context.Context, root string, specPkg specPackage, baseBranch string, chunks [][]string, prompt, logDir string) ([]parallelWorkerResult, error) {
	results := make([]parallelWorkerResult, 0, len(chunks))
	for i, chunk := range chunks {
		name := parallelWorkerName(specPkg.ID, i)
		path := filepath.Join(root, worktreesDir, name)
		branch := parallelWorkerBranch(name)
		if _, err := a.runBinary(ctx, "git", []string{"worktree", "add", "-b", branch, path, baseBranch}, root); err != nil {
			return results, err
		}
		if err := os.MkdirAll(path, 0o755); err != nil {
			return results, err
		}

		workerPrompt := buildParallelWorkerPrompt(prompt, chunk)
		logPath := filepath.Join(logDir, name+"-request.md")
		if err := os.WriteFile(logPath, []byte(workerPrompt), 0o644); err != nil {
			return results, err
		}

		results = append(results, parallelWorkerResult{Name: name, Branch: branch, Worktree: path})
	}
	return results, nil
}

func (l *parallelRunLifecycle) prepare() error {
	plan, err := l.app.prepareParallelRunPlan(l.ctx, l.root, l.specPkg, l.tasks, l.workflowCfg, l.dryRun)
	if err != nil {
		return err
	}
	l.plan = plan

	if err := l.initProgress(); err != nil {
		return l.finishPrepareFailure(
			err,
			"progress_log_failed",
			"Parallel progress log failed",
			err.Error(),
			map[string]any{"stage": "initialize"},
			false,
		)
	}
	if err := l.publishRunEvent("queued", "ready", "Parallel run queued", "Parallel run lifecycle is preparing worker staging and validation", nil); err != nil {
		return l.finishPrepareFailure(
			err,
			"progress_log_failed",
			"Parallel progress log failed",
			err.Error(),
			map[string]any{"stage": "prepare"},
			false,
		)
	}
	if !l.dryRun {
		if err := l.runPreflight(); err != nil {
			return l.finishPrepareFailure(
				fmt.Errorf("parallel preflight: %w", err),
				"preflight_failed",
				"Parallel run preflight failed",
				err.Error(),
				map[string]any{"stage": "preflight"},
				false,
			)
		}
		if err := l.publishRunEvent(
			"queued",
			"ready",
			"Parallel run preflight passed",
			"Runtime prerequisites are satisfied for parallel execution",
			map[string]any{"stage": "preflight"},
		); err != nil {
			return l.finishPrepareFailure(
				err,
				"progress_log_failed",
				"Parallel progress log failed",
				err.Error(),
				map[string]any{"stage": "preflight"},
				false,
			)
		}
	}

	results, err := l.app.stageParallelWorkers(l.ctx, l.root, l.specPkg, l.plan.BaseBranch, l.plan.Chunks, l.prompt, l.plan.LogDir)
	l.results = results
	if err != nil {
		return l.finishPrepareFailure(
			err,
			"staging_failed",
			"Parallel worker staging failed",
			err.Error(),
			map[string]any{"stage": "staging"},
			len(l.results) > 0,
		)
	}

	for i := range l.results {
		if err := l.publishWorkerEvent(
			l.results[i].Name,
			"queued",
			"ready",
			"Worker staged and queued",
			"Worker worktree and prompt are ready for execution",
			map[string]any{
				"branch":   l.results[i].Branch,
				"worktree": l.results[i].Worktree,
				"index":    i,
			},
		); err != nil {
			return l.finishPrepareFailure(
				err,
				"progress_log_failed",
				"Parallel progress log failed",
				err.Error(),
				map[string]any{"stage": "worker_staging"},
				true,
			)
		}
	}

	return nil
}

func (l *parallelRunLifecycle) runPreflight() error {
	previewReq := l.app.newExecutionRequest(
		l.specPkg.ID,
		l.root,
		l.prompt,
		executionModeParallel,
		suggestDelegationPlan(executionModeParallel, l.prompt, "", ""),
		l.systemCfg,
		l.codexCfg,
	)
	preflight, _, err := l.app.runPreflight(l.ctx, previewReq)
	if writeErr := writeJSONFile(filepath.Join(l.plan.LogDir, strings.ToLower(l.specPkg.ID)+"-parallel-preflight.json"), preflight); writeErr != nil {
		return writeErr
	}
	return err
}

func (l *parallelRunLifecycle) completeDryRun() error {
	report := l.report()
	if err := l.finishRun(
		report,
		nil,
		"done",
		"dry_run",
		"Parallel run prepared",
		"Dry-run preparation completed without executing workers",
		map[string]any{"dry_run": true},
	); err != nil {
		return err
	}

	fmt.Fprintf(l.app.stdout, "Prepared %d parallel work packages for %s\n", len(l.results), l.specPkg.ID)
	fmt.Fprintf(l.app.stdout, "Cleanup policy: %s\n", report.CleanupPolicy)
	return nil
}

func (l *parallelRunLifecycle) executeWorkers() {
	workers := l.parallelWorkerExecutions()
	outcomes := make(chan parallelWorkerOutcome, len(workers))
	var wg sync.WaitGroup
	for _, worker := range workers {
		wg.Add(1)
		go func(worker parallelWorkerExecution) {
			defer wg.Done()
			outcomes <- l.executeParallelWorker(worker)
		}(worker)
	}
	go func() {
		wg.Wait()
		close(outcomes)
	}()

	for outcome := range outcomes {
		l.recordParallelWorkerOutcome(outcome)
	}
}

func (l *parallelRunLifecycle) finalize() error {
	if err := l.blockMergeOnProgressFailure(); err != nil {
		report := l.report()
		report.MergeBlocked = true
		return l.finishRun(
			report,
			err,
			"failed",
			"progress_log_failed",
			"Parallel progress log failed",
			l.progressErr.Error(),
			map[string]any{"merge_blocked": true},
		)
	}
	if err := l.blockMergeOnWorkerFailures(); err != nil {
		report := l.report()
		report.MergeBlocked = true
		return l.finishRun(
			report,
			err,
			"failed",
			"merge_blocked",
			"Parallel merge blocked",
			err.Error(),
			map[string]any{"merge_blocked": true},
		)
	}
	if err := l.publishRunEvent(
		"merge_pending",
		"ready",
		"Parallel workers ready to merge",
		"All workers passed execution and validation",
		map[string]any{"workers": len(l.results)},
	); err != nil {
		preserveErr := l.preserveWorkers(
			"Worker artifacts preserved after progress log failure before merge",
			map[string]any{"reason": "progress_log_failed"},
		)
		report := l.report()
		report.MergeBlocked = true
		return l.finishRun(
			report,
			errors.Join(err, preserveErr),
			"failed",
			"progress_log_failed",
			"Parallel progress log failed",
			err.Error(),
			map[string]any{"merge_blocked": true},
		)
	}
	if err := l.publishRunEvent(
		"merging",
		"active",
		"Parallel merge started",
		"Merging worker branches into the base branch",
		nil,
	); err != nil {
		preserveErr := l.preserveWorkers(
			"Worker artifacts preserved after progress log failure before merge",
			map[string]any{"reason": "progress_log_failed"},
		)
		report := l.report()
		report.MergeBlocked = true
		return l.finishRun(
			report,
			errors.Join(err, preserveErr),
			"failed",
			"progress_log_failed",
			"Parallel progress log failed",
			err.Error(),
			map[string]any{"merge_blocked": true},
		)
	}
	if err := l.mergeWorkers(); err != nil {
		report := l.report()
		if l.progressErr != nil && !hasParallelMergeErrors(l.results) {
			preserveErr := l.preserveWorkers(
				"Worker artifacts preserved after progress log failure during merge",
				map[string]any{"reason": "progress_log_failed"},
			)
			return l.finishRun(
				report,
				errors.Join(err, preserveErr),
				"failed",
				"progress_log_failed",
				"Parallel progress log failed",
				l.progressErr.Error(),
				nil,
			)
		}
		return l.finishRun(
			report,
			err,
			"failed",
			"merge_failed",
			"Parallel merge failed",
			err.Error(),
			nil,
		)
	}

	report, cleanupOutcome := l.completeCleanupPhase()
	if len(cleanupOutcome.Failures) > 0 {
		cleanupErr := fmt.Errorf("parallel cleanup failed: %s", strings.Join(cleanupOutcome.Failures, "; "))
		return l.finishRun(
			report,
			cleanupErr,
			"failed",
			"cleanup_failed",
			"Parallel cleanup failed",
			cleanupErr.Error(),
			map[string]any{"failures": len(cleanupOutcome.Failures)},
		)
	}
	if l.progressErr != nil {
		return l.finishRun(
			report,
			l.progressErr,
			"failed",
			"progress_log_failed",
			"Parallel progress log failed",
			l.progressErr.Error(),
			nil,
		)
	}
	if err := l.finishRun(
		report,
		nil,
		"done",
		"completed",
		"Parallel run completed",
		"All workers merged and cleanup finished successfully",
		nil,
	); err != nil {
		return err
	}

	fmt.Fprintf(l.app.stdout, "Executed %s in %d parallel worktrees with %s\n", l.specPkg.ID, len(l.results), normalizeRunner(l.systemCfg.Runner))
	fmt.Fprintf(l.app.stdout, "Cleanup policy: %s\n", report.CleanupPolicy)
	return nil
}

func (l *parallelRunLifecycle) blockMergeOnProgressFailure() error {
	if l.progressErr == nil {
		return nil
	}
	preserveErr := l.preserveWorkers(
		"Worker artifacts preserved after progress log failure during execution",
		map[string]any{"reason": "progress_log_failed"},
	)
	return errors.Join(fmt.Errorf("parallel progress log failed: %w", l.progressErr), preserveErr)
}

func (l *parallelRunLifecycle) blockMergeOnWorkerFailures() error {
	if !hasParallelRunFailures(l.results) {
		return nil
	}

	preserveErr := l.preserveWorkers(
		"Worker artifacts were preserved because at least one worker failed before merge",
		map[string]any{"reason": "worker_failure"},
	)
	return errors.Join(
		fmt.Errorf("parallel execution blocked merge: %s", summarizeParallelRunFailures(l.results)),
		preserveErr,
	)
}

func (l *parallelRunLifecycle) mergeWorkers() error {
	for i := range l.results {
		if err := l.publishWorkerEvent(
			l.results[i].Name,
			"merging",
			"active",
			"Worker merge started",
			"Merging worker branch into the base branch",
			map[string]any{"branch": l.results[i].Branch},
		); err != nil {
			return err
		}

		l.results[i].MergeAttempted = true
		if _, err := l.app.runBinary(l.ctx, "git", []string{"merge", "--no-ff", l.results[i].Branch, "-m", "merge " + l.results[i].Branch}, l.root); err != nil {
			l.results[i].MergeError = err.Error()
			publishErr := l.publishWorkerEvent(
				l.results[i].Name,
				"failed",
				"merge_failed",
				"Worker merge failed",
				err.Error(),
				map[string]any{"branch": l.results[i].Branch},
			)
			preserveErr := l.preserveWorkers(
				"Worker artifacts preserved after merge failure",
				map[string]any{"reason": "merge_failed"},
			)
			return errors.Join(fmt.Errorf("merge %s: %w", l.results[i].Branch, err), publishErr, preserveErr)
		}

		l.results[i].MergeSucceeded = true
		if err := l.publishWorkerEvent(
			l.results[i].Name,
			"done",
			"merged",
			"Worker merge completed",
			"Worker branch merged into the base branch",
			map[string]any{"branch": l.results[i].Branch},
		); err != nil {
			return err
		}
	}
	return nil
}

func (l *parallelRunLifecycle) cleanupMergedWorkers() parallelCleanupOutcome {
	return l.cleanupParallelWorkerArtifactsOutcome()
}

func (l *parallelRunLifecycle) completeCleanupPhase() (parallelRunReport, parallelCleanupOutcome) {
	cleanupOutcome := applyParallelPruneError(l.cleanupMergedWorkers(), l.pruneParallelWorktrees())
	return reportWithParallelCleanupOutcome(l.report(), cleanupOutcome), cleanupOutcome
}

func (l *parallelRunLifecycle) writeFinishedReport(report parallelRunReport) error {
	return writeParallelRunReport(l.root, l.finishedReportSnapshot(report))
}

func (l *parallelRunLifecycle) finishedReportSnapshot(report parallelRunReport) parallelRunReport {
	report.Workers = append([]parallelWorkerResult(nil), l.results...)
	if progressErr := l.progressFailure(); progressErr != nil {
		report.ProgressLogFailed = true
		report.ProgressLogError = progressErr.Error()
	}
	report.FinishedAt = l.app.now().Format(time.RFC3339)
	return report
}

func reportWithParallelCleanupOutcome(report parallelRunReport, outcome parallelCleanupOutcome) parallelRunReport {
	report.PruneAttempted = true
	report.PruneError = outcome.PruneError
	return report
}

func (l *parallelRunLifecycle) newParallelWorkerExecution(index int, chunk []string) parallelWorkerExecution {
	worker := l.results[index]
	workerPrompt := buildParallelWorkerPrompt(l.prompt, chunk)
	delegation := suggestDelegationPlan(executionModeParallel, workerPrompt, "", "")
	request := l.app.newExecutionRequest(l.specPkg.ID, worker.Worktree, workerPrompt, executionModeParallel, delegation, l.systemCfg, l.codexCfg)
	return parallelWorkerExecution{
		Index:   index,
		Result:  worker,
		Prompt:  workerPrompt,
		Request: request,
	}
}

func (l *parallelRunLifecycle) parallelWorkerExecutions() []parallelWorkerExecution {
	workers := make([]parallelWorkerExecution, len(l.plan.Chunks))
	for i, chunk := range l.plan.Chunks {
		workers[i] = l.newParallelWorkerExecution(i, chunk)
	}
	return workers
}

func (l *parallelRunLifecycle) executeParallelWorker(worker parallelWorkerExecution) parallelWorkerOutcome {
	execResult, validationReport, runErr := l.app.executeRun(
		l.ctx,
		l.root,
		worker.Result.Name,
		worker.Request,
		worker.Result.Worktree,
		l.qualityCfg,
		l.progress,
		worker.Result.Name,
	)
	return parallelWorkerOutcome{
		Index:            worker.Index,
		ExecutionResult:  execResult,
		ValidationReport: validationReport,
		RunErr:           runErr,
	}
}

func (l *parallelRunLifecycle) recordParallelWorkerOutcome(outcome parallelWorkerOutcome) {
	l.results[outcome.Index] = applyParallelWorkerOutcome(l.results[outcome.Index], outcome)
}

func applyParallelWorkerOutcome(result parallelWorkerResult, outcome parallelWorkerOutcome) parallelWorkerResult {
	result.SessionID = outcome.ExecutionResult.SessionID
	result.RetryCount = outcome.ExecutionResult.RetryCount
	result.StartedAt = outcome.ExecutionResult.StartedAt
	result.FinishedAt = outcome.ExecutionResult.FinishedAt
	result.ExecutionPassed = executionTurnsPassed(outcome.ExecutionResult)
	result.ValidationPassed = outcome.ValidationReport.Passed
	result.ExecutionError, result.ValidationError = classifyParallelWorkerFailure(outcome.ExecutionResult, outcome.ValidationReport, outcome.RunErr)
	return result
}

func classifyParallelWorkerFailure(execResult executionResult, validationReport validationReport, runErr error) (string, string) {
	switch {
	case runErr == nil:
		return "", ""
	case execResult.Succeeded && validationReport.Passed && isParallelProgressFailure(runErr):
		return "", ""
	case !execResult.Succeeded && len(validationReport.Steps) == 0:
		return firstNonEmptyString(execResult.Error, runErr.Error()), ""
	case !validationReport.Passed && len(validationReport.Steps) > 0:
		return "", validationFailureMessage(validationReport, runErr)
	case !execResult.Succeeded:
		return firstNonEmptyString(execResult.Error, runErr.Error()), ""
	default:
		return runErr.Error(), ""
	}
}

func (l *parallelRunLifecycle) cleanupParallelWorkerArtifactsOutcome() parallelCleanupOutcome {
	return parallelCleanupOutcome{
		Failures: l.cleanupParallelWorkerArtifacts(),
	}
}

func (l *parallelRunLifecycle) cleanupParallelWorkerArtifacts() []string {
	var cleanupFailures []string
	for i := range l.results {
		beforeFailures := len(cleanupFailures)
		cleanupFailures = append(cleanupFailures, l.app.cleanupParallelWorker(l.ctx, l.root, &l.results[i])...)
		if len(cleanupFailures) > beforeFailures {
			_ = l.publishWorkerEvent(
				l.results[i].Name,
				"failed",
				"cleanup_failed",
				"Worker cleanup failed",
				firstNonEmptyString(l.results[i].CleanupError, "worker cleanup failed"),
				map[string]any{
					"branch":   l.results[i].Branch,
					"worktree": l.results[i].Worktree,
				},
			)
			continue
		}
		_ = l.publishWorkerEvent(
			l.results[i].Name,
			"done",
			"cleaned",
			"Worker cleanup completed",
			"Worker worktree and branch were removed after merge",
			map[string]any{
				"branch":   l.results[i].Branch,
				"worktree": l.results[i].Worktree,
			},
		)
	}
	return cleanupFailures
}

func (l *parallelRunLifecycle) pruneParallelWorktrees() string {
	if _, err := l.app.runBinary(l.ctx, "git", []string{"worktree", "prune"}, l.root); err != nil {
		return err.Error()
	}
	return ""
}

func applyParallelPruneError(outcome parallelCleanupOutcome, pruneError string) parallelCleanupOutcome {
	if pruneError == "" {
		return outcome
	}
	outcome.PruneError = pruneError
	outcome.Failures = append(outcome.Failures, "worktree prune: "+pruneError)
	return outcome
}

func (l *parallelRunLifecycle) markWorkersPreserved() {
	for i := range l.results {
		l.results[i].Preserved = true
	}
}

func (l *parallelRunLifecycle) preserveWorkers(detail string, metadata map[string]any) error {
	var errs []error
	for i := range l.results {
		alreadyPreserved := l.results[i].Preserved
		l.results[i].Preserved = true
		if alreadyPreserved {
			continue
		}
		workerMetadata := copyParallelProgressMetadata(metadata)
		if workerMetadata == nil {
			workerMetadata = map[string]any{}
		}
		if l.results[i].Branch != "" {
			workerMetadata["branch"] = l.results[i].Branch
		}
		if l.results[i].Worktree != "" {
			workerMetadata["worktree"] = l.results[i].Worktree
		}
		if err := l.publishWorkerEvent(
			l.results[i].Name,
			"done",
			"preserved",
			"Worker preserved for inspection",
			detail,
			workerMetadata,
		); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

func (l *parallelRunLifecycle) report() parallelRunReport {
	report := l.plan.Report
	report.Workers = append([]parallelWorkerResult(nil), l.results...)
	return report
}

func (l *parallelRunLifecycle) finishPrepareFailure(runErr error, status, summary, detail string, metadata map[string]any, preserveWorkers bool) error {
	if preserveWorkers && len(l.results) > 0 {
		runErr = errors.Join(
			runErr,
			l.preserveWorkers(
				"Worker artifacts preserved after prepare-stage failure",
				map[string]any{"reason": status},
			),
		)
	}
	report := l.report()
	report.MergeBlocked = true
	return l.finishRun(report, runErr, "failed", status, summary, detail, metadata)
}

func (l *parallelRunLifecycle) initProgress() error {
	if l.progress != nil {
		return nil
	}
	progress, err := l.app.newParallelProgressSink(parallelProgressSinkConfig{
		Path:   l.plan.EventLogPath,
		SpecID: l.specPkg.ID,
		RunID:  l.plan.RunID,
		Now:    l.app.now,
	})
	if err != nil {
		l.progressErr = errors.Join(l.progressErr, err)
		return err
	}
	l.progress = progress
	return nil
}

func (l *parallelRunLifecycle) publishRunEvent(phase, status, summary, detail string, metadata map[string]any) error {
	return l.publishProgress(parallelProgressEventInput{
		Source:   parallelProgressSourceLifecycle,
		Scope:    parallelProgressScopeRun,
		Phase:    phase,
		Status:   status,
		Summary:  summary,
		Detail:   detail,
		Metadata: metadata,
	})
}

func (l *parallelRunLifecycle) publishWorkerEvent(workerName, phase, status, summary, detail string, metadata map[string]any) error {
	return l.publishProgress(parallelProgressEventInput{
		Source:     parallelProgressSourceLifecycle,
		Scope:      parallelProgressScopeWorker,
		WorkerName: workerName,
		Phase:      phase,
		Status:     status,
		Summary:    summary,
		Detail:     detail,
		Metadata:   metadata,
	})
}

func (l *parallelRunLifecycle) publishProgress(event parallelProgressEventInput) error {
	if l.progress == nil {
		return nil
	}
	if l.progressErr != nil {
		return l.progressErr
	}
	if err := l.progress.Publish(event); err != nil {
		l.progressErr = errors.Join(l.progressErr, err)
		return err
	}
	return nil
}

func (l *parallelRunLifecycle) closeProgress() error {
	if l.progress == nil || l.progressClosed {
		return nil
	}
	l.progressClosed = true
	if err := l.progress.Close(); err != nil {
		l.progressCloseErr = errors.Join(l.progressCloseErr, err)
		return err
	}
	return nil
}

func (l *parallelRunLifecycle) progressFailure() error {
	return errors.Join(l.progressErr, l.progressCloseErr)
}

func (l *parallelRunLifecycle) finishRun(report parallelRunReport, runErr error, phase, status, summary, detail string, metadata map[string]any) error {
	publishErr := l.publishRunEvent(phase, status, summary, detail, metadata)
	closeErr := l.closeProgress()
	if err := l.writeFinishedReport(report); err != nil {
		return errors.Join(runErr, publishErr, closeErr, err)
	}
	if err := l.app.writeParallelExecutionEvidence(l.root, l.specPkg.ID, l.plan.RunID, status, l.dryRun); err != nil {
		return errors.Join(runErr, publishErr, closeErr, err)
	}
	return errors.Join(runErr, publishErr, closeErr)
}

func hasParallelMergeErrors(results []parallelWorkerResult) bool {
	for _, result := range results {
		if result.MergeError != "" {
			return true
		}
	}
	return false
}

func parallelWorkerName(specID string, index int) string {
	return strings.ToLower(specID) + "-p" + strconv.Itoa(index+1)
}

func parallelWorkerBranch(name string) string {
	return "namba/" + name
}

func buildParallelWorkerPrompt(basePrompt string, chunk []string) string {
	return basePrompt + "\n\n## Assigned work package\n\n" + strings.Join(chunk, "\n")
}
