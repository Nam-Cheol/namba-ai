package namba

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
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
	SpecID            string                 `json:"spec_id"`
	RunID             string                 `json:"run_id,omitempty"`
	BaseBranch        string                 `json:"base_branch"`
	DryRun            bool                   `json:"dry_run"`
	CleanupPolicy     string                 `json:"cleanup_policy"`
	EventLogPath      string                 `json:"event_log_path,omitempty"`
	ProgressLogFailed bool                   `json:"progress_log_failed,omitempty"`
	ProgressLogError  string                 `json:"progress_log_error,omitempty"`
	MergeBlocked      bool                   `json:"merge_blocked"`
	PruneAttempted    bool                   `json:"prune_attempted"`
	PruneError        string                 `json:"prune_error,omitempty"`
	Workers           []parallelWorkerResult `json:"workers"`
	StartedAt         string                 `json:"started_at"`
	FinishedAt        string                 `json:"finished_at"`
}

func (a *App) executeParallelRun(ctx context.Context, root string, specPkg specPackage, tasks []string, prompt string, qualityCfg qualityConfig, systemCfg systemConfig, codexCfg codexConfig, workflowCfg workflowConfig, dryRun bool) error {
	lifecycle, err := a.newParallelRunLifecycle(ctx, root, specPkg, tasks, prompt, qualityCfg, systemCfg, codexCfg, workflowCfg, dryRun)
	if err != nil {
		return err
	}
	if err := lifecycle.prepare(); err != nil {
		return err
	}
	if dryRun {
		return lifecycle.completeDryRun()
	}
	lifecycle.executeWorkers()
	return lifecycle.finalize()
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
