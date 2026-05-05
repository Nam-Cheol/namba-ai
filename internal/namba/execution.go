package namba

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type executionTurnResult struct {
	Name            string   `json:"name"`
	Role            string   `json:"role,omitempty"`
	Model           string   `json:"model,omitempty"`
	Profile         string   `json:"profile,omitempty"`
	WebSearch       bool     `json:"web_search,omitempty"`
	AddDirs         []string `json:"add_dirs,omitempty"`
	SessionMode     string   `json:"session_mode,omitempty"`
	SessionAction   string   `json:"session_action,omitempty"`
	ReasoningEffort string   `json:"reasoning_effort,omitempty"`
	Output          string   `json:"output,omitempty"`
	Succeeded       bool     `json:"succeeded"`
	StartedAt       string   `json:"started_at"`
	FinishedAt      string   `json:"finished_at"`
	Error           string   `json:"error,omitempty"`
}

type executionResult struct {
	Runner             string                `json:"runner"`
	SpecID             string                `json:"spec_id"`
	WorkDir            string                `json:"work_dir"`
	HeadSHA            string                `json:"head_sha,omitempty"`
	ExecutionMode      string                `json:"execution_mode"`
	ApprovalPolicy     string                `json:"approval_policy"`
	SandboxMode        string                `json:"sandbox_mode"`
	Model              string                `json:"model,omitempty"`
	Profile            string                `json:"profile,omitempty"`
	WebSearch          bool                  `json:"web_search,omitempty"`
	AddDirs            []string              `json:"add_dirs,omitempty"`
	SessionMode        string                `json:"session_mode,omitempty"`
	SessionID          string                `json:"session_id,omitempty"`
	SessionContinuity  string                `json:"session_continuity,omitempty"`
	RetryCount         int                   `json:"retry_count,omitempty"`
	ValidationAttempts int                   `json:"validation_attempts,omitempty"`
	DelegationMode     string                `json:"delegation_mode,omitempty"`
	DelegationPlan     delegationPlan        `json:"delegation_plan,omitempty"`
	DelegationObserved bool                  `json:"delegation_observed"`
	DelegationSummary  string                `json:"delegation_summary,omitempty"`
	Turns              []executionTurnResult `json:"turns,omitempty"`
	Output             string                `json:"output"`
	Succeeded          bool                  `json:"succeeded"`
	StartedAt          string                `json:"started_at"`
	FinishedAt         string                `json:"finished_at"`
	Error              string                `json:"error,omitempty"`
}

type validationReport struct {
	SpecID     string           `json:"spec_id"`
	HeadSHA    string           `json:"head_sha,omitempty"`
	Passed     bool             `json:"passed"`
	Attempt    int              `json:"attempt"`
	StartedAt  string           `json:"started_at"`
	FinishedAt string           `json:"finished_at"`
	Steps      []validationStep `json:"steps"`
}

type validationStep struct {
	Name    string `json:"name"`
	Command string `json:"command"`
	Skipped bool   `json:"skipped"`
	Passed  bool   `json:"passed"`
	Output  string `json:"output,omitempty"`
	Error   string `json:"error,omitempty"`
}

type runner interface {
	Execute(context.Context, executionRequest, codexCapabilityMatrix) (executionTurnResult, error)
}

type codexRunner struct {
	lookPath  func(string) (string, error)
	runBinary func(context.Context, string, []string, string) (string, error)
	now       func() time.Time
}

func (r codexRunner) Execute(ctx context.Context, req executionRequest, capabilities codexCapabilityMatrix) (executionTurnResult, error) {
	result := executionTurnResult{
		Name:            firstNonBlank(req.TurnName, "implement"),
		Role:            strings.TrimSpace(req.TurnRole),
		Model:           strings.TrimSpace(req.Model),
		Profile:         strings.TrimSpace(req.Profile),
		WebSearch:       req.WebSearch,
		AddDirs:         append([]string(nil), req.AddDirs...),
		SessionMode:     normalizeSessionMode(req.SessionMode),
		ReasoningEffort: strings.TrimSpace(req.RequestedReasoningEffort),
		StartedAt:       r.now().Format(time.RFC3339),
	}
	if req.ResumeSession {
		result.SessionAction = "resume"
	} else {
		result.SessionAction = "exec"
	}

	args, err := buildCodexExecArgs(req, capabilities)
	if err != nil {
		result.FinishedAt = r.now().Format(time.RFC3339)
		result.Error = err.Error()
		return result, err
	}

	if _, err := r.lookPath("codex"); err != nil {
		result.FinishedAt = r.now().Format(time.RFC3339)
		result.Error = fmt.Sprintf("runner codex is not available: %v", err)
		return result, fmt.Errorf(result.Error)
	}

	output, err := r.runBinary(ctx, "codex", args, req.WorkDir)
	result.Output = output
	result.FinishedAt = r.now().Format(time.RFC3339)
	if err != nil {
		result.Error = err.Error()
		return result, err
	}

	result.Succeeded = true
	return result, nil
}

func buildCodexExecArgs(req executionRequest, capabilities codexCapabilityMatrix) ([]string, error) {
	invocation, err := resolveCodexInvocation(req, capabilities)
	if err != nil {
		return nil, err
	}
	return invocation.Args, nil
}

func (a *App) runnerFor(cfg systemConfig) (runner, error) {
	switch normalizeRunner(cfg.Runner) {
	case "", "codex":
		return codexRunner{
			lookPath:  a.lookPath,
			runBinary: a.runBinary,
			now:       a.now,
		}, nil
	default:
		return nil, fmt.Errorf("runner %q is not supported", cfg.Runner)
	}
}

func (a *App) executeRun(ctx context.Context, projectRoot, logID string, req executionRequest, validationRoot string, cfg qualityConfig, progress parallelProgressSink, progressWorkerName string) (executionResult, validationReport, error) {
	selectedRunner, err := a.runnerFor(systemConfig{Runner: req.Runner})
	if err != nil {
		return executionResult{}, validationReport{}, err
	}

	if resolvedAddDirs, err := resolveRuntimeAddDirs(req.WorkDir, req.AddDirs); err == nil {
		req.AddDirs = resolvedAddDirs
	}

	result := executionResult{
		Runner:            normalizeRunner(req.Runner),
		SpecID:            req.SpecID,
		WorkDir:           req.WorkDir,
		ExecutionMode:     string(normalizeExecutionMode(req.Mode)),
		ApprovalPolicy:    normalizeApprovalPolicy(req.ApprovalPolicy),
		SandboxMode:       normalizeSandboxMode(req.SandboxMode),
		Model:             strings.TrimSpace(req.Model),
		Profile:           strings.TrimSpace(req.Profile),
		WebSearch:         req.WebSearch,
		AddDirs:           append([]string(nil), req.AddDirs...),
		SessionMode:       normalizeSessionMode(req.SessionMode),
		SessionID:         logID,
		SessionContinuity: "single-turn",
		DelegationMode:    executionDelegationMode(req.Mode),
		DelegationPlan:    req.DelegationPlan,
		DelegationSummary: summarizeDelegationPlan(req.DelegationPlan),
		StartedAt:         a.now().Format(time.RFC3339),
	}
	if result.SessionMode == "" {
		result.SessionMode = "stateful"
	}

	publishProgress := func(phase, status, summary, detail string, metadata map[string]any) error {
		if progress == nil {
			return nil
		}
		return progress.Publish(parallelProgressEventInput{
			Source:     parallelProgressSourceLifecycle,
			Scope:      parallelProgressScopeWorker,
			WorkerName: progressWorkerName,
			Phase:      phase,
			Status:     status,
			Summary:    summary,
			Detail:     detail,
			Metadata:   metadata,
		})
	}
	progressPath := ""
	if progress != nil {
		progressPath = progress.Path()
	}
	hooks := newHookLifecycle(a, projectRoot, logID, req, progressPath)
	writeRunEvidence := func(status string, validationAttempts int, failureSummary string) error {
		return hooks.writeRunEvidence(ctx, status, validationAttempts, false, failureSummary)
	}
	writeRunEvidenceWithProgressFailure := func(status string, validationAttempts int, failureSummary string) error {
		return hooks.writeRunEvidence(ctx, status, validationAttempts, true, failureSummary)
	}

	if err := writeJSONFile(filepath.Join(projectRoot, logsDir, "runs", logID+"-request.json"), req); err != nil {
		return result, validationReport{}, err
	}
	if err := hooks.Trigger(ctx, hookTrigger{
		Event:       hookEventBeforePreflight,
		StageStatus: "pending",
	}); err != nil {
		result.FinishedAt = a.now().Format(time.RFC3339)
		result.Error = err.Error()
		if writeErr := a.writeExecutionArtifacts(projectRoot, logID, result); writeErr != nil {
			return result, validationReport{}, writeErr
		}
		if writeErr := writeRunEvidence("hook_failed", 0, result.Error); writeErr != nil {
			return result, validationReport{}, errors.Join(err, writeErr)
		}
		return result, validationReport{}, err
	}

	preflight, capabilities, preflightErr := a.runPreflight(ctx, req)
	if err := writeJSONFile(filepath.Join(projectRoot, logsDir, "runs", logID+"-preflight.json"), preflight); err != nil {
		return result, validationReport{}, err
	}
	afterPreflightErr := hooks.Trigger(ctx, hookTrigger{
		Event:        hookEventAfterPreflight,
		StageStatus:  hookStageStatus(preflight.Passed, preflightErr),
		ErrorSummary: hookErrorSummary(preflightErr),
		EventData: map[string]any{
			"preflight_path": filepath.ToSlash(filepath.Join(logsDir, "runs", logID+"-preflight.json")),
		},
	})
	if preflightErr == nil && afterPreflightErr != nil {
		result.FinishedAt = a.now().Format(time.RFC3339)
		result.Error = afterPreflightErr.Error()
		if err := writeJSONFile(filepath.Join(projectRoot, logsDir, "runs", logID+"-execution.json"), result); err != nil {
			return result, validationReport{}, err
		}
		if err := writeRunEvidence("hook_failed", 0, result.Error); err != nil {
			return result, validationReport{}, errors.Join(afterPreflightErr, err)
		}
		return result, validationReport{}, afterPreflightErr
	}
	if preflightErr != nil {
		result.FinishedAt = a.now().Format(time.RFC3339)
		result.Error = preflightErr.Error()
		if err := writeJSONFile(filepath.Join(projectRoot, logsDir, "runs", logID+"-execution.json"), result); err != nil {
			return result, validationReport{}, err
		}
		if err := writeRunEvidence("preflight_failed", 0, result.Error); err != nil {
			return result, validationReport{}, err
		}
		publishErr := publishProgress("failed", "preflight_failed", "Worker execution preflight failed", preflightErr.Error(), nil)
		if publishErr != nil {
			if err := writeRunEvidenceWithProgressFailure("preflight_failed", 0, result.Error); err != nil {
				return result, validationReport{}, errors.Join(preflightErr, afterPreflightErr, publishErr, err)
			}
		}
		return result, validationReport{}, errors.Join(preflightErr, afterPreflightErr, publishErr)
	}

	if err := hooks.Trigger(ctx, hookTrigger{
		Event:       hookEventBeforeExecution,
		StageStatus: "pending",
		EventData: map[string]any{
			"runner":          normalizeRunner(req.Runner),
			"delegation_mode": executionDelegationMode(req.Mode),
		},
	}); err != nil {
		result.FinishedAt = a.now().Format(time.RFC3339)
		result.Error = err.Error()
		if writeErr := a.writeExecutionArtifacts(projectRoot, logID, result); writeErr != nil {
			return result, validationReport{}, writeErr
		}
		if writeErr := writeRunEvidence("hook_failed", 0, result.Error); writeErr != nil {
			return result, validationReport{}, errors.Join(err, writeErr)
		}
		return result, validationReport{}, err
	}

	turnRequests := buildExecutionTurnRequests(req)
	teamContinuationMode := "degraded-fresh-exec"
	if codexSessionStateful(req.SessionMode) {
		teamContinuationMode = "codex-exec-resume-last"
	}

	if err := publishProgress(
		"running",
		"active",
		"Worker execution started",
		"Execution turns are starting",
		map[string]any{"session_id": logID},
	); err != nil {
		result.FinishedAt = a.now().Format(time.RFC3339)
		result.Error = err.Error()
		if writeErr := a.writeExecutionArtifacts(projectRoot, logID, result); writeErr != nil {
			return result, validationReport{}, writeErr
		}
		if writeErr := writeRunEvidence("progress_log_failed", 0, result.Error); writeErr != nil {
			return result, validationReport{}, writeErr
		}
		return result, validationReport{}, err
	}

	for _, turnReq := range turnRequests {
		turnResult, err := selectedRunner.Execute(ctx, turnReq, capabilities)
		result.Turns = append(result.Turns, turnResult)
		if turnReq.ResumeSession {
			result.SessionContinuity = teamContinuationMode
		}
		if turnReq.TurnRole != "" && turnReq.TurnRole != req.DelegationPlan.IntegratorRole {
			result.DelegationObserved = true
		}
		if err != nil {
			result.Output = joinExecutionOutputs(result.Turns)
			result.FinishedAt = a.now().Format(time.RFC3339)
			result.Error = err.Error()
			if writeErr := a.writeExecutionArtifacts(projectRoot, logID, result); writeErr != nil {
				return result, validationReport{}, writeErr
			}
			afterExecutionErr := hooks.Trigger(ctx, hookTrigger{
				Event:        hookEventAfterExecution,
				StageStatus:  "failed",
				ErrorSummary: err.Error(),
				EventData: map[string]any{
					"execution_path": filepath.ToSlash(filepath.Join(logsDir, "runs", logID+"-execution.json")),
				},
			})
			if writeErr := writeRunEvidence("execution_failed", 0, result.Error); writeErr != nil {
				return result, validationReport{}, errors.Join(err, afterExecutionErr, writeErr)
			}
			publishErr := publishProgress(
				"failed",
				"execution_failed",
				"Worker execution failed",
				err.Error(),
				map[string]any{"session_id": logID},
			)
			if publishErr != nil {
				if writeErr := writeRunEvidenceWithProgressFailure("execution_failed", 0, result.Error); writeErr != nil {
					return result, validationReport{}, errors.Join(err, publishErr, writeErr)
				}
			}
			return result, validationReport{}, errors.Join(err, afterExecutionErr, publishErr)
		}
	}

	if err := hooks.Trigger(ctx, hookTrigger{
		Event:       hookEventAfterExecution,
		StageStatus: "succeeded",
		EventData: map[string]any{
			"turn_count": len(result.Turns),
		},
	}); err != nil {
		result.Output = joinExecutionOutputs(result.Turns)
		result.FinishedAt = a.now().Format(time.RFC3339)
		result.Error = err.Error()
		if writeErr := a.writeExecutionArtifacts(projectRoot, logID, result); writeErr != nil {
			return result, validationReport{}, writeErr
		}
		if writeErr := writeRunEvidence("hook_failed", 0, result.Error); writeErr != nil {
			return result, validationReport{}, errors.Join(err, writeErr)
		}
		return result, validationReport{}, err
	}

	var finalReport validationReport
	maxAttempts := maxInt(req.RepairAttempts, 0) + 1
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		if err := hooks.Trigger(ctx, hookTrigger{
			Event:       hookEventBeforeValidation,
			StageStatus: "pending",
			Attempt:     attempt,
			EventData: map[string]any{
				"validation_attempt": attempt,
				"validation_root":    validationRoot,
			},
		}); err != nil {
			result.Output = joinExecutionOutputs(result.Turns)
			result.FinishedAt = a.now().Format(time.RFC3339)
			result.Error = err.Error()
			if writeErr := a.writeExecutionArtifacts(projectRoot, logID, result); writeErr != nil {
				return result, finalReport, writeErr
			}
			if writeErr := writeRunEvidence("hook_failed", result.ValidationAttempts, result.Error); writeErr != nil {
				return result, finalReport, errors.Join(err, writeErr)
			}
			return result, finalReport, err
		}

		if err := publishProgress(
			"validating",
			"active",
			"Worker validation started",
			fmt.Sprintf("Validation pipeline attempt %d is starting", attempt),
			map[string]any{"session_id": logID, "attempt": attempt},
		); err != nil {
			result.Output = joinExecutionOutputs(result.Turns)
			result.FinishedAt = a.now().Format(time.RFC3339)
			result.Error = err.Error()
			if writeErr := a.writeExecutionArtifacts(projectRoot, logID, result); writeErr != nil {
				return result, finalReport, writeErr
			}
			if writeErr := writeRunEvidence("progress_log_failed", result.ValidationAttempts, result.Error); writeErr != nil {
				return result, finalReport, writeErr
			}
			return result, finalReport, err
		}

		result.ValidationAttempts = attempt
		report, validationErr := a.runValidationReport(ctx, validationRoot, cfg, req.SpecID, attempt)
		finalReport = report
		if err := writeJSONFile(filepath.Join(projectRoot, logsDir, "runs", fmt.Sprintf("%s-validation-attempt-%d.json", logID, attempt)), report); err != nil {
			return result, finalReport, err
		}
		afterValidationErr := hooks.Trigger(ctx, hookTrigger{
			Event:        hookEventAfterValidation,
			StageStatus:  hookStageStatus(report.Passed, validationErr),
			ErrorSummary: hookErrorSummary(validationErr),
			Attempt:      attempt,
			EventData: map[string]any{
				"validation_path":    filepath.ToSlash(filepath.Join(logsDir, "runs", fmt.Sprintf("%s-validation-attempt-%d.json", logID, attempt))),
				"validation_passed":  report.Passed,
				"validation_attempt": attempt,
				"retry_remaining":    attempt <= req.RepairAttempts,
			},
		})
		if afterValidationErr != nil {
			result.Output = joinExecutionOutputs(result.Turns)
			result.FinishedAt = a.now().Format(time.RFC3339)
			status := "hook_failed"
			result.Error = afterValidationErr.Error()
			if validationErr != nil {
				status = "validation_failed"
				result.Error = fmt.Sprintf("validation failed after %d repair attempt(s): %s", result.RetryCount, validationFailureMessage(finalReport, validationErr))
			}
			if err := a.writeExecutionArtifacts(projectRoot, logID, result); err != nil {
				return result, finalReport, err
			}
			if err := writeJSONFile(filepath.Join(projectRoot, logsDir, "runs", logID+"-validation.json"), finalReport); err != nil {
				return result, finalReport, err
			}
			if err := writeRunEvidence(status, attempt, result.Error); err != nil {
				return result, finalReport, errors.Join(validationErr, afterValidationErr, err)
			}
			return result, finalReport, errors.Join(validationErr, afterValidationErr)
		}
		if validationErr == nil {
			result.Output = joinExecutionOutputs(result.Turns)
			result.FinishedAt = a.now().Format(time.RFC3339)
			result.Succeeded = true
			if err := a.writeExecutionArtifacts(projectRoot, logID, result); err != nil {
				return result, finalReport, err
			}
			if err := writeJSONFile(filepath.Join(projectRoot, logsDir, "runs", logID+"-validation.json"), finalReport); err != nil {
				return result, finalReport, err
			}
			publishErr := publishProgress(
				"merge_pending",
				"ready",
				"Worker ready to merge",
				"Execution and validation passed",
				map[string]any{"session_id": logID, "validation_attempts": attempt},
			)
			if publishErr != nil {
				if err := writeRunEvidence("progress_log_failed", attempt, publishErr.Error()); err != nil {
					return result, finalReport, errors.Join(publishErr, err)
				}
				return result, finalReport, publishErr
			}
			if err := writeRunEvidence("completed", attempt, ""); err != nil {
				return result, finalReport, err
			}
			return result, finalReport, nil
		}
		if attempt > req.RepairAttempts {
			result.Output = joinExecutionOutputs(result.Turns)
			result.FinishedAt = a.now().Format(time.RFC3339)
			result.Error = fmt.Sprintf("validation failed after %d repair attempt(s): %s", result.RetryCount, validationFailureMessage(finalReport, validationErr))
			if err := a.writeExecutionArtifacts(projectRoot, logID, result); err != nil {
				return result, finalReport, err
			}
			if err := writeJSONFile(filepath.Join(projectRoot, logsDir, "runs", logID+"-validation.json"), finalReport); err != nil {
				return result, finalReport, err
			}
			if err := writeRunEvidence("validation_failed", attempt, result.Error); err != nil {
				return result, finalReport, err
			}
			publishErr := publishProgress(
				"failed",
				"validation_failed",
				"Worker validation failed",
				validationFailureMessage(finalReport, validationErr),
				map[string]any{"session_id": logID, "validation_attempts": attempt},
			)
			if publishErr != nil {
				if err := writeRunEvidenceWithProgressFailure("validation_failed", attempt, result.Error); err != nil {
					return result, finalReport, errors.Join(validationErr, publishErr, err)
				}
			}
			return result, finalReport, errors.Join(validationErr, publishErr)
		}

		if err := publishProgress(
			"running",
			"repairing",
			"Worker repair attempt started",
			"Validation failed and a repair attempt will run",
			map[string]any{"session_id": logID, "attempt": attempt},
		); err != nil {
			result.Output = joinExecutionOutputs(result.Turns)
			result.FinishedAt = a.now().Format(time.RFC3339)
			result.Error = err.Error()
			if writeErr := a.writeExecutionArtifacts(projectRoot, logID, result); writeErr != nil {
				return result, finalReport, writeErr
			}
			if writeErr := writeJSONFile(filepath.Join(projectRoot, logsDir, "runs", logID+"-validation.json"), finalReport); writeErr != nil {
				return result, finalReport, writeErr
			}
			if writeErr := writeRunEvidence("progress_log_failed", attempt, result.Error); writeErr != nil {
				return result, finalReport, writeErr
			}
			return result, finalReport, err
		}

		repairReq := req
		repairReq.ResumeSession = codexSessionStateful(req.SessionMode)
		repairReq.TurnName = fmt.Sprintf("repair-%d", attempt)
		repairReq.TurnRole = req.DelegationPlan.IntegratorRole
		repairReq.Prompt = buildRepairPrompt(req, finalReport, attempt, !repairReq.ResumeSession)
		repairReq.RequestedReasoningEffort = ""

		repairResult, repairErr := selectedRunner.Execute(ctx, repairReq, capabilities)
		result.Turns = append(result.Turns, repairResult)
		result.RetryCount++
		if repairReq.ResumeSession {
			result.SessionContinuity = "codex-exec-resume-last"
		} else {
			result.SessionContinuity = "degraded-fresh-exec"
		}
		if repairErr != nil {
			result.Output = joinExecutionOutputs(result.Turns)
			result.FinishedAt = a.now().Format(time.RFC3339)
			result.Error = repairErr.Error()
			if err := a.writeExecutionArtifacts(projectRoot, logID, result); err != nil {
				return result, finalReport, err
			}
			if err := writeJSONFile(filepath.Join(projectRoot, logsDir, "runs", logID+"-validation.json"), finalReport); err != nil {
				return result, finalReport, err
			}
			if err := writeRunEvidence("repair_failed", attempt, result.Error); err != nil {
				return result, finalReport, err
			}
			publishErr := publishProgress(
				"failed",
				"repair_failed",
				"Worker repair attempt failed",
				repairErr.Error(),
				map[string]any{"session_id": logID, "attempt": attempt},
			)
			if publishErr != nil {
				if err := writeRunEvidenceWithProgressFailure("repair_failed", attempt, result.Error); err != nil {
					return result, finalReport, errors.Join(repairErr, publishErr, err)
				}
			}
			return result, finalReport, errors.Join(repairErr, publishErr)
		}
	}

	result.Output = joinExecutionOutputs(result.Turns)
	result.FinishedAt = a.now().Format(time.RFC3339)
	result.Error = "execution ended without a successful validation result"
	if err := a.writeExecutionArtifacts(projectRoot, logID, result); err != nil {
		return result, finalReport, err
	}
	if err := writeJSONFile(filepath.Join(projectRoot, logsDir, "runs", logID+"-validation.json"), finalReport); err != nil {
		return result, finalReport, err
	}
	if err := writeRunEvidence("validation_failed", result.ValidationAttempts, result.Error); err != nil {
		return result, finalReport, err
	}
	publishErr := publishProgress(
		"failed",
		"validation_failed",
		"Worker execution ended without validation success",
		result.Error,
		map[string]any{"session_id": logID},
	)
	if publishErr != nil {
		if err := writeRunEvidenceWithProgressFailure("validation_failed", result.ValidationAttempts, result.Error); err != nil {
			return result, finalReport, errors.Join(fmt.Errorf(result.Error), publishErr, err)
		}
	}
	return result, finalReport, errors.Join(fmt.Errorf(result.Error), publishErr)
}

func (a *App) writeExecutionArtifacts(projectRoot, logID string, result executionResult) error {
	if err := writeRunText(filepath.Join(projectRoot, logsDir, "runs", logID+"-result.txt"), result.Output); err != nil {
		return err
	}
	if err := writeJSONFile(filepath.Join(projectRoot, logsDir, "runs", logID+"-execution.json"), result); err != nil {
		return err
	}
	return nil
}

func (a *App) runValidationReport(ctx context.Context, root string, cfg qualityConfig, specID string, attempt int) (validationReport, error) {
	report := validationReport{
		SpecID:    specID,
		Passed:    true,
		Attempt:   attempt,
		StartedAt: a.now().Format(time.RFC3339),
	}

	for _, step := range validationPipelineSteps(cfg) {
		if step.Command == "" || step.Command == "none" {
			step.Skipped = true
			report.Steps = append(report.Steps, step)
			continue
		}

		output, err := runShellCommand(ctx, a.runCmd, step.Command, root)
		step.Output = output
		if err != nil {
			step.Error = err.Error()
			report.Passed = false
			report.Steps = append(report.Steps, step)
			report.FinishedAt = a.now().Format(time.RFC3339)
			return report, fmt.Errorf("validation failed for %q: %w", step.Command, err)
		}

		step.Passed = true
		report.Steps = append(report.Steps, step)
	}

	report.FinishedAt = a.now().Format(time.RFC3339)
	return report, nil
}

func validationPipelineSteps(cfg qualityConfig) []validationStep {
	return []validationStep{
		{Name: "test", Command: strings.TrimSpace(cfg.TestCommand)},
		{Name: "lint", Command: strings.TrimSpace(cfg.LintCommand)},
		{Name: "typecheck", Command: strings.TrimSpace(cfg.TypecheckCommand)},
		{Name: "build", Command: strings.TrimSpace(cfg.BuildCommand)},
		{Name: "migration-dry-run", Command: strings.TrimSpace(cfg.MigrationDryRunCommand)},
		{Name: "smoke-start", Command: strings.TrimSpace(cfg.SmokeStartCommand)},
		{Name: "output-contract", Command: strings.TrimSpace(cfg.OutputContractCommand)},
	}
}

func buildExecutionTurnRequests(req executionRequest) []executionRequest {
	base := req
	base.TurnName = "implement"
	base.TurnRole = req.DelegationPlan.IntegratorRole

	turns := []executionRequest{base}
	if normalizeExecutionMode(req.Mode) != executionModeTeam {
		return turns
	}

	stateful := codexSessionStateful(req.SessionMode)
	for _, profile := range req.DelegationPlan.SelectedRoleProfiles {
		turn := req
		turn.TurnName = roleTurnName(profile.Role)
		turn.TurnRole = profile.Role
		turn.ResumeSession = stateful
		turn.Model = firstNonBlank(profile.Model, req.Model)
		turn.Profile = req.Profile
		turn.RequestedReasoningEffort = profile.ModelReasoningEffort
		turn.Prompt = buildDelegationTurnPrompt(req, profile, !stateful)
		turns = append(turns, turn)
	}
	return turns
}

func roleTurnName(role string) string {
	role = strings.TrimSpace(strings.TrimPrefix(role, "namba-"))
	role = strings.ReplaceAll(role, "_", "-")
	if role == "" {
		return "specialist"
	}
	return role
}

func buildDelegationTurnPrompt(req executionRequest, profile agentRuntimeProfile, includeBasePrompt bool) string {
	lines := []string{
		fmt.Sprintf("Continue the current `%s` execution as `%s` in the same workspace.", req.SpecID, profile.Role),
		"Make direct repository changes for your specialty, then stop so the next turn or validator can continue.",
	}
	if profile.ModelReasoningEffort != "" {
		lines = append(lines, fmt.Sprintf("Requested reasoning effort for this turn: `%s`.", profile.ModelReasoningEffort))
	}
	if profile.Role == req.DelegationPlan.ReviewerRole {
		lines = append(lines, "Act as the final reviewer for the same-workspace team run. Close acceptance gaps you find instead of only describing them.")
	} else {
		lines = append(lines, "Focus on the acceptance items that match your specialty. Keep integration context intact for the next turn.")
	}
	if len(req.DelegationPlan.RoutingRationale) > 0 {
		lines = append(lines, "", "## Routing context")
		for _, reason := range req.DelegationPlan.RoutingRationale {
			lines = append(lines, "- "+reason)
		}
	}
	if includeBasePrompt {
		lines = append(lines, "", "## Base execution context", req.Prompt)
	}
	return strings.Join(lines, "\n")
}

func buildRepairPrompt(req executionRequest, report validationReport, attempt int, includeBasePrompt bool) string {
	lines := []string{
		fmt.Sprintf("Validation failed for `%s`. Repair the issues below and stop so validation can run again.", req.SpecID),
		fmt.Sprintf("This is repair attempt %d of %d.", attempt, req.RepairAttempts),
		"",
		"## Validation failures",
	}
	lines = append(lines, formatValidationFailures(report)...)
	if includeBasePrompt {
		lines = append(lines, "", "## Base execution context", req.Prompt)
	}
	return strings.Join(lines, "\n")
}

func formatValidationFailures(report validationReport) []string {
	lines := make([]string, 0)
	for _, step := range report.Steps {
		switch {
		case step.Error != "":
			lines = append(lines, fmt.Sprintf("- %s: %s", step.Name, step.Error))
			if strings.TrimSpace(step.Output) != "" {
				lines = append(lines, fmt.Sprintf("  output: %s", step.Output))
			}
		case step.Skipped:
			lines = append(lines, fmt.Sprintf("- %s: skipped", step.Name))
		}
	}
	if len(lines) == 0 {
		return []string{"- validation failed without a recorded failing step"}
	}
	return lines
}

func executionDelegationMode(mode executionMode) string {
	switch normalizeExecutionMode(mode) {
	case executionModeSolo:
		return "single-runner"
	case executionModeTeam:
		return "same-workspace-team"
	case executionModeParallel:
		return "worktree-parallel"
	default:
		return "standalone"
	}
}

func joinExecutionOutputs(turns []executionTurnResult) string {
	if len(turns) == 0 {
		return ""
	}
	if len(turns) == 1 {
		return strings.TrimSpace(turns[0].Output)
	}
	parts := make([]string, 0, len(turns))
	for _, turn := range turns {
		label := turn.Name
		if turn.Role != "" {
			label = label + " (" + turn.Role + ")"
		}
		if output := strings.TrimSpace(turn.Output); output != "" {
			parts = append(parts, fmt.Sprintf("## %s\n%s", label, output))
		}
	}
	return strings.Join(parts, "\n\n")
}

func executionTurnsPassed(result executionResult) bool {
	if len(result.Turns) == 0 {
		return false
	}
	for _, turn := range result.Turns {
		if !turn.Succeeded {
			return false
		}
	}
	return true
}

func writeRunText(path, content string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(content), 0o644)
}

func summarizeDelegationPlan(plan delegationPlan) string {
	if len(plan.SelectedRoles) == 0 {
		return "No delegated specialists planned; keep work inside the standalone runner."
	}

	parts := []string{fmt.Sprintf("Planned roles: %s.", strings.Join(plan.SelectedRoles, ", "))}
	if len(plan.SelectedRoleProfiles) > 0 {
		runtimeSummaries := make([]string, 0, len(plan.SelectedRoleProfiles))
		for _, profile := range plan.SelectedRoleProfiles {
			if summary := formatAgentRuntimeProfile(profile); summary != "" {
				runtimeSummaries = append(runtimeSummaries, summary)
			}
		}
		if len(runtimeSummaries) > 0 {
			parts = append(parts, fmt.Sprintf("Runtime profiles: %s.", strings.Join(runtimeSummaries, "; ")))
		}
	}
	if plan.DelegationBudget > 0 {
		parts = append(parts, fmt.Sprintf("Delegation budget: %d.", plan.DelegationBudget))
	}
	if plan.IntegratorRole != "" {
		parts = append(parts, fmt.Sprintf("Integrator: %s.", plan.IntegratorRole))
	}
	if plan.ReviewerRole != "" {
		parts = append(parts, fmt.Sprintf("Reviewer: %s.", plan.ReviewerRole))
	}
	return strings.Join(parts, " ")
}

func writeJSONFile(path string, value any) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}
