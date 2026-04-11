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
	BaseBranch string
	Chunks     [][]string
	LogDir     string
	Report     parallelRunReport
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
	app         *App
	ctx         context.Context
	root        string
	specPkg     specPackage
	tasks       []string
	prompt      string
	qualityCfg  qualityConfig
	systemCfg   systemConfig
	codexCfg    codexConfig
	workflowCfg workflowConfig
	dryRun      bool
	plan        parallelRunPlan
	results     []parallelWorkerResult
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

func (a *App) prepareParallelRunPlan(ctx context.Context, root string, specPkg specPackage, tasks []string, prompt string, systemCfg systemConfig, codexCfg codexConfig, workflowCfg workflowConfig, dryRun bool) (parallelRunPlan, error) {
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

	if !dryRun {
		previewReq := a.newExecutionRequest(specPkg.ID, root, prompt, executionModeParallel, suggestDelegationPlan(executionModeParallel, prompt, "", ""), systemCfg, codexCfg)
		preflight, _, err := a.runPreflight(ctx, previewReq)
		if writeErr := writeJSONFile(filepath.Join(logDir, strings.ToLower(specPkg.ID)+"-parallel-preflight.json"), preflight); writeErr != nil {
			return parallelRunPlan{}, writeErr
		}
		if err != nil {
			return parallelRunPlan{}, fmt.Errorf("parallel preflight: %w", err)
		}
	}

	return parallelRunPlan{
		BaseBranch: baseBranch,
		Chunks:     chunks,
		LogDir:     logDir,
		Report: parallelRunReport{
			SpecID:        specPkg.ID,
			BaseBranch:    baseBranch,
			DryRun:        dryRun,
			CleanupPolicy: "Success removes temporary worktrees and deletes worker branches after every merge succeeds. Any execution, validation, or merge failure preserves all worker worktrees and branches for inspection.",
			StartedAt:     a.now().Format(time.RFC3339),
		},
	}, nil
}

func (a *App) stageParallelWorkers(ctx context.Context, root string, specPkg specPackage, baseBranch string, chunks [][]string, prompt, logDir string) ([]parallelWorkerResult, error) {
	results := make([]parallelWorkerResult, len(chunks))
	for i, chunk := range chunks {
		name := parallelWorkerName(specPkg.ID, i)
		path := filepath.Join(root, worktreesDir, name)
		branch := parallelWorkerBranch(name)
		if _, err := a.runBinary(ctx, "git", []string{"worktree", "add", "-b", branch, path, baseBranch}, root); err != nil {
			return nil, err
		}
		if err := os.MkdirAll(path, 0o755); err != nil {
			return nil, err
		}

		workerPrompt := buildParallelWorkerPrompt(prompt, chunk)
		logPath := filepath.Join(logDir, name+"-request.md")
		if err := os.WriteFile(logPath, []byte(workerPrompt), 0o644); err != nil {
			return nil, err
		}

		results[i] = parallelWorkerResult{Name: name, Branch: branch, Worktree: path}
	}
	return results, nil
}

func (l *parallelRunLifecycle) prepare() error {
	plan, err := l.app.prepareParallelRunPlan(l.ctx, l.root, l.specPkg, l.tasks, l.prompt, l.systemCfg, l.codexCfg, l.workflowCfg, l.dryRun)
	if err != nil {
		return err
	}
	results, err := l.app.stageParallelWorkers(l.ctx, l.root, l.specPkg, plan.BaseBranch, plan.Chunks, l.prompt, plan.LogDir)
	if err != nil {
		return err
	}
	l.plan = plan
	l.results = results
	return nil
}

func (l *parallelRunLifecycle) completeDryRun() error {
	report := l.report()
	if err := l.writeFinishedReport(report); err != nil {
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
	wg.Wait()
	close(outcomes)

	for outcome := range outcomes {
		l.recordParallelWorkerOutcome(outcome)
	}
}

func (l *parallelRunLifecycle) finalize() error {
	if err := l.blockMergeOnWorkerFailures(); err != nil {
		report := l.report()
		report.MergeBlocked = true
		if writeErr := l.writeFinishedReport(report); writeErr != nil {
			return writeErr
		}
		return err
	}
	if err := l.mergeWorkers(); err != nil {
		if writeErr := l.writeFinishedReport(l.report()); writeErr != nil {
			return writeErr
		}
		return err
	}

	report, cleanupOutcome := l.completeCleanupPhase()
	if err := l.writeFinishedReport(report); err != nil {
		return err
	}

	if len(cleanupOutcome.Failures) > 0 {
		return fmt.Errorf("parallel cleanup failed: %s", strings.Join(cleanupOutcome.Failures, "; "))
	}

	fmt.Fprintf(l.app.stdout, "Executed %s in %d parallel worktrees with %s\n", l.specPkg.ID, len(l.results), normalizeRunner(l.systemCfg.Runner))
	fmt.Fprintf(l.app.stdout, "Cleanup policy: %s\n", report.CleanupPolicy)
	return nil
}

func (l *parallelRunLifecycle) blockMergeOnWorkerFailures() error {
	if !hasParallelRunFailures(l.results) {
		return nil
	}

	l.markWorkersPreserved()
	return fmt.Errorf("parallel execution blocked merge: %s", summarizeParallelRunFailures(l.results))
}

func (l *parallelRunLifecycle) mergeWorkers() error {
	for i := range l.results {
		l.results[i].MergeAttempted = true
		if _, err := l.app.runBinary(l.ctx, "git", []string{"merge", "--no-ff", l.results[i].Branch, "-m", "merge " + l.results[i].Branch}, l.root); err != nil {
			l.results[i].MergeError = err.Error()
			l.markWorkersPreserved()
			return fmt.Errorf("merge %s: %w", l.results[i].Branch, err)
		}
		l.results[i].MergeSucceeded = true
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
	execResult, validationReport, runErr := l.app.executeRun(l.ctx, l.root, worker.Result.Name, worker.Request, worker.Result.Worktree, l.qualityCfg)
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
		cleanupFailures = append(cleanupFailures, l.app.cleanupParallelWorker(l.ctx, l.root, &l.results[i])...)
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

func (l *parallelRunLifecycle) report() parallelRunReport {
	report := l.plan.Report
	report.Workers = append([]parallelWorkerResult(nil), l.results...)
	return report
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
