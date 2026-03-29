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

type parallelWorkerResult struct {
	Name             string `json:"name"`
	Branch           string `json:"branch"`
	Worktree         string `json:"worktree"`
	SessionID        string `json:"session_id,omitempty"`
	RetryCount       int    `json:"retry_count,omitempty"`
	StartedAt        string `json:"started_at,omitempty"`
	FinishedAt       string `json:"finished_at,omitempty"`
	ExecutionPassed  bool   `json:"execution_passed"`
	ValidationPassed bool   `json:"validation_passed"`
	ExecutionError   string `json:"execution_error,omitempty"`
	ValidationError  string `json:"validation_error,omitempty"`
	MergeAttempted   bool   `json:"merge_attempted"`
	MergeSucceeded   bool   `json:"merge_succeeded"`
	MergeError       string `json:"merge_error,omitempty"`
	CleanupAttempted bool   `json:"cleanup_attempted"`
	WorktreeRemoved  bool   `json:"worktree_removed"`
	BranchRemoved    bool   `json:"branch_removed"`
	CleanupError     string `json:"cleanup_error,omitempty"`
	Preserved        bool   `json:"preserved"`
}

type parallelRunReport struct {
	SpecID         string                 `json:"spec_id"`
	BaseBranch     string                 `json:"base_branch"`
	DryRun         bool                   `json:"dry_run"`
	CleanupPolicy  string                 `json:"cleanup_policy"`
	MergeBlocked   bool                   `json:"merge_blocked"`
	PruneAttempted bool                   `json:"prune_attempted"`
	PruneError     string                 `json:"prune_error,omitempty"`
	Workers        []parallelWorkerResult `json:"workers"`
	StartedAt      string                 `json:"started_at"`
	FinishedAt     string                 `json:"finished_at"`
}

func (a *App) executeParallelRun(ctx context.Context, root string, specPkg specPackage, tasks []string, prompt string, qualityCfg qualityConfig, systemCfg systemConfig, codexCfg codexConfig, workflowCfg workflowConfig, dryRun bool) error {
	if _, err := a.lookPath("git"); err != nil {
		return errors.New("parallel execution requires git")
	}
	if !isGitRepository(root) {
		return errors.New("parallel execution requires a git repository")
	}

	baseBranch, err := a.currentBranch(ctx, root)
	if err != nil {
		return err
	}

	workers := minInt(len(tasks), maxInt(workflowCfg.MaxParallelWorkers, 1))
	if workers == 0 {
		workers = 1
	}
	chunks := chunkTasks(tasks, workers)
	logDir := filepath.Join(root, logsDir, "runs")
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		return err
	}

	if !dryRun {
		previewReq := a.newExecutionRequest(specPkg.ID, root, prompt, executionModeParallel, suggestDelegationPlan(executionModeParallel, prompt, "", ""), systemCfg, codexCfg)
		preflight, _, err := a.runPreflight(ctx, previewReq)
		if writeErr := writeJSONFile(filepath.Join(logDir, strings.ToLower(specPkg.ID)+"-parallel-preflight.json"), preflight); writeErr != nil {
			return writeErr
		}
		if err != nil {
			return fmt.Errorf("parallel preflight: %w", err)
		}
	}

	report := parallelRunReport{
		SpecID:        specPkg.ID,
		BaseBranch:    baseBranch,
		DryRun:        dryRun,
		CleanupPolicy: "Success removes temporary worktrees and deletes worker branches after every merge succeeds. Any execution, validation, or merge failure preserves all worker worktrees and branches for inspection.",
		StartedAt:     a.now().Format(time.RFC3339),
	}
	results := make([]parallelWorkerResult, len(chunks))

	for i, chunk := range chunks {
		name := strings.ToLower(specPkg.ID) + "-p" + strconv.Itoa(i+1)
		path := filepath.Join(root, worktreesDir, name)
		branch := "namba/" + name
		if _, err := a.runBinary(ctx, "git", []string{"worktree", "add", "-b", branch, path, baseBranch}, root); err != nil {
			return err
		}
		if err := os.MkdirAll(path, 0o755); err != nil {
			return err
		}

		workerPrompt := prompt + "\n\n## Assigned work package\n\n" + strings.Join(chunk, "\n")
		logPath := filepath.Join(logDir, name+"-request.md")
		if err := os.WriteFile(logPath, []byte(workerPrompt), 0o644); err != nil {
			return err
		}

		results[i] = parallelWorkerResult{Name: name, Branch: branch, Worktree: path}
		if dryRun {
			continue
		}
	}

	report.Workers = results
	if dryRun {
		report.FinishedAt = a.now().Format(time.RFC3339)
		if err := writeParallelRunReport(root, report); err != nil {
			return err
		}
		fmt.Fprintf(a.stdout, "Prepared %d parallel work packages for %s\n", len(results), specPkg.ID)
		fmt.Fprintf(a.stdout, "Cleanup policy: %s\n", report.CleanupPolicy)
		return nil
	}

	var wg sync.WaitGroup
	for i, chunk := range chunks {
		wg.Add(1)
		go func(index int, chunk []string) {
			defer wg.Done()

			name := results[index].Name
			path := results[index].Worktree
			workerPrompt := prompt + "\n\n## Assigned work package\n\n" + strings.Join(chunk, "\n")
			delegation := suggestDelegationPlan(executionModeParallel, workerPrompt, "", "")
			request := a.newExecutionRequest(specPkg.ID, path, workerPrompt, executionModeParallel, delegation, systemCfg, codexCfg)

			execResult, validationReport, runErr := a.executeRun(ctx, root, name, request, path, qualityCfg)
			results[index].SessionID = execResult.SessionID
			results[index].RetryCount = execResult.RetryCount
			results[index].StartedAt = execResult.StartedAt
			results[index].FinishedAt = execResult.FinishedAt
			results[index].ExecutionPassed = executionTurnsPassed(execResult)
			results[index].ValidationPassed = validationReport.Passed
			switch {
			case runErr == nil:
			case !execResult.Succeeded && len(validationReport.Steps) == 0:
				results[index].ExecutionError = firstNonEmptyString(execResult.Error, runErr.Error())
			case !validationReport.Passed && len(validationReport.Steps) > 0:
				results[index].ValidationError = validationFailureMessage(validationReport, runErr)
			case !execResult.Succeeded:
				results[index].ExecutionError = firstNonEmptyString(execResult.Error, runErr.Error())
			default:
				results[index].ExecutionError = runErr.Error()
			}
		}(i, chunk)
	}
	wg.Wait()

	report.Workers = results
	if hasParallelRunFailures(results) {
		report.MergeBlocked = true
		for i := range results {
			results[i].Preserved = true
		}
		report.Workers = results
		report.FinishedAt = a.now().Format(time.RFC3339)
		if err := writeParallelRunReport(root, report); err != nil {
			return err
		}
		return fmt.Errorf("parallel execution blocked merge: %s", summarizeParallelRunFailures(results))
	}

	for i := range results {
		results[i].MergeAttempted = true
		if _, err := a.runBinary(ctx, "git", []string{"merge", "--no-ff", results[i].Branch, "-m", "merge " + results[i].Branch}, root); err != nil {
			results[i].MergeError = err.Error()
			for j := range results {
				results[j].Preserved = true
			}
			report.Workers = results
			report.FinishedAt = a.now().Format(time.RFC3339)
			if writeErr := writeParallelRunReport(root, report); writeErr != nil {
				return writeErr
			}
			return fmt.Errorf("merge %s: %w", results[i].Branch, err)
		}
		results[i].MergeSucceeded = true
	}

	var cleanupFailures []string
	for i := range results {
		cleanupFailures = append(cleanupFailures, a.cleanupParallelWorker(ctx, root, &results[i])...)
	}
	report.PruneAttempted = true
	if _, err := a.runBinary(ctx, "git", []string{"worktree", "prune"}, root); err != nil {
		report.PruneError = err.Error()
		cleanupFailures = append(cleanupFailures, "worktree prune: "+err.Error())
	}

	report.Workers = results
	report.FinishedAt = a.now().Format(time.RFC3339)
	if err := writeParallelRunReport(root, report); err != nil {
		return err
	}
	if len(cleanupFailures) > 0 {
		return fmt.Errorf("parallel cleanup failed: %s", strings.Join(cleanupFailures, "; "))
	}

	fmt.Fprintf(a.stdout, "Executed %s in %d parallel worktrees with %s\n", specPkg.ID, len(results), normalizeRunner(systemCfg.Runner))
	fmt.Fprintf(a.stdout, "Cleanup policy: %s\n", report.CleanupPolicy)
	return nil
}

func (a *App) cleanupParallelWorker(ctx context.Context, root string, result *parallelWorkerResult) []string {
	result.CleanupAttempted = true
	var cleanupFailures []string

	if _, err := a.runBinary(ctx, "git", []string{"worktree", "remove", "--force", result.Worktree}, root); err != nil {
		result.CleanupError = appendCleanupError(result.CleanupError, "remove worktree: "+err.Error())
		cleanupFailures = append(cleanupFailures, fmt.Sprintf("remove worktree %s: %v", result.Worktree, err))
	} else {
		result.WorktreeRemoved = true
	}
	if _, err := a.runBinary(ctx, "git", []string{"branch", "-D", result.Branch}, root); err != nil {
		result.CleanupError = appendCleanupError(result.CleanupError, "delete branch: "+err.Error())
		cleanupFailures = append(cleanupFailures, fmt.Sprintf("delete branch %s: %v", result.Branch, err))
	} else {
		result.BranchRemoved = true
	}
	return cleanupFailures
}

func writeParallelRunReport(root string, report parallelRunReport) error {
	path := filepath.Join(root, logsDir, "runs", strings.ToLower(report.SpecID)+"-parallel.json")
	return writeJSONFile(path, report)
}

func hasParallelRunFailures(results []parallelWorkerResult) bool {
	for _, result := range results {
		if result.ExecutionError != "" || result.ValidationError != "" || !result.ExecutionPassed || !result.ValidationPassed {
			return true
		}
	}
	return false
}

func summarizeParallelRunFailures(results []parallelWorkerResult) string {
	issues := make([]string, 0)
	for _, result := range results {
		switch {
		case result.ExecutionError != "":
			issues = append(issues, result.Name+" execution failed: "+result.ExecutionError)
		case result.ValidationError != "":
			issues = append(issues, result.Name+" validation failed: "+result.ValidationError)
		}
	}
	if len(issues) == 0 {
		return "parallel workers did not all pass"
	}
	return strings.Join(issues, "; ")
}

func validationFailureMessage(report validationReport, fallback error) string {
	for _, step := range report.Steps {
		if step.Error != "" {
			if step.Name == "" {
				return step.Error
			}
			return step.Name + ": " + step.Error
		}
	}
	if fallback != nil {
		return fallback.Error()
	}
	return "validation failed"
}

func appendCleanupError(current, next string) string {
	if current == "" {
		return next
	}
	return current + "; " + next
}

func firstNonEmptyString(values ...string) string {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			return value
		}
	}
	return ""
}
