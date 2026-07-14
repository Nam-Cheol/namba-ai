package namba

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
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
	ThreadID        string   `json:"thread_id,omitempty"`
	ReasoningEffort string   `json:"reasoning_effort,omitempty"`
	Output          string   `json:"output,omitempty"`
	CommandArgs     []string `json:"command_args,omitempty"`
	StdoutPath      string   `json:"stdout_path,omitempty"`
	StderrPath      string   `json:"stderr_path,omitempty"`
	ExitCode        int      `json:"exit_code,omitempty"`
	Succeeded       bool     `json:"succeeded"`
	StartedAt       string   `json:"started_at"`
	FinishedAt      string   `json:"finished_at"`
	Error           string   `json:"error,omitempty"`
}

type codexExecCommand struct {
	Args  []string
	Input string
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
	runCmd    func(context.Context, string, []string, string, string) (string, string, error)
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

	command, err := buildCodexExecCommand(req, capabilities)
	if err != nil {
		result.FinishedAt = r.now().Format(time.RFC3339)
		result.Error = err.Error()
		return result, err
	}

	if _, err := r.lookPath("codex"); err != nil {
		result.FinishedAt = r.now().Format(time.RFC3339)
		result.Error = fmt.Sprintf("runner codex is not available: %v", err)
		return result, fmt.Errorf("%s", result.Error)
	}

	result.CommandArgs = append([]string{"codex"}, command.Args...)
	var stdout, stderr string
	if r.runCmd != nil {
		stdout, stderr, err = r.runCmd(ctx, "codex", command.Args, req.WorkDir, command.Input)
		output := strings.TrimSpace(strings.Join(nonEmptyArgs([]string{stdout, stderr}), "\n"))
		result.Output = output
		if writeErr := writeRunnerStreamArtifacts(req.WorkDir, req.SpecID, result.Name, stdout, stderr, &result); writeErr != nil && err == nil {
			err = writeErr
		}
	} else {
		var output string
		output, err = r.runBinary(ctx, "codex", command.Args, req.WorkDir)
		result.Output = output
		if writeErr := writeRunnerStreamArtifacts(req.WorkDir, req.SpecID, result.Name, output, "", &result); writeErr != nil && err == nil {
			err = writeErr
		}
	}
	result.FinishedAt = r.now().Format(time.RFC3339)
	result.ThreadID = firstCodexThreadID(result.Output)
	if err != nil {
		result.ExitCode = commandExitCode(err)
		result.Error = err.Error()
		return result, err
	}

	result.Succeeded = true
	return result, nil
}

// firstCodexThreadID intentionally accepts only structured JSONL evidence. A
// human-readable UUID in an agent response is not safe resume authority.
func firstCodexThreadID(output string) string {
	for _, line := range strings.Split(output, "\n") {
		var event any
		if err := json.Unmarshal([]byte(strings.TrimSpace(line)), &event); err != nil {
			continue
		}
		if id := threadIDFromJSONValue(event); id != "" {
			return id
		}
	}
	return ""
}

func threadIDFromJSONValue(value any) string {
	switch typed := value.(type) {
	case map[string]any:
		for _, key := range []string{"thread_id", "threadId", "session_id", "sessionId"} {
			if id, ok := typed[key].(string); ok && strings.TrimSpace(id) != "" {
				return strings.TrimSpace(id)
			}
		}
		if thread, ok := typed["thread"].(map[string]any); ok {
			if id, ok := thread["id"].(string); ok && strings.TrimSpace(id) != "" {
				return strings.TrimSpace(id)
			}
		}
		for _, child := range typed {
			if id := threadIDFromJSONValue(child); id != "" {
				return id
			}
		}
	case []any:
		for _, child := range typed {
			if id := threadIDFromJSONValue(child); id != "" {
				return id
			}
		}
	}
	return ""
}

func buildCodexExecArgs(req executionRequest, capabilities codexCapabilityMatrix) ([]string, error) {
	command, err := buildCodexExecCommand(req, capabilities)
	if err != nil {
		return nil, err
	}
	return command.Args, nil
}

func buildCodexExecCommand(req executionRequest, capabilities codexCapabilityMatrix) (codexExecCommand, error) {
	invocation, err := resolveCodexInvocation(req, capabilities)
	if err != nil {
		return codexExecCommand{}, err
	}
	return codexExecCommand{Args: invocation.Args, Input: req.Prompt}, nil
}

func writeRunnerStreamArtifacts(workDir, specID, turnName, stdout, stderr string, result *executionTurnResult) error {
	logID := strings.ToLower(strings.TrimSpace(specID))
	if logID == "" {
		return nil
	}
	turn := normalizeArtifactToken(firstNonBlank(turnName, "implement"))
	stdoutRel := filepath.ToSlash(filepath.Join(logsDir, "runs", logID+"-"+turn+"-stdout.txt"))
	stderrRel := filepath.ToSlash(filepath.Join(logsDir, "runs", logID+"-"+turn+"-stderr.txt"))
	if err := writeRunText(filepath.Join(workDir, stdoutRel), stdout); err != nil {
		return err
	}
	if err := writeRunText(filepath.Join(workDir, stderrRel), stderr); err != nil {
		return err
	}
	result.StdoutPath = stdoutRel
	result.StderrPath = stderrRel
	return nil
}

func frontendViolationCheckFailure(root, specID, output string) error {
	report := loadFrontendBriefReport(root, specID)
	if !report.Exists || !report.Valid || report.Header.TaskClassification != frontendTaskClassificationMajor {
		return nil
	}

	section, ok := markdownSection(output, "Do-Not Design Violation Check", 2)
	if !ok {
		return errors.New("Do-Not Design Violation Check failed: frontend-major runner result must include `## Do-Not Design Violation Check` with changed-file evidence and banned-pattern findings.")
	}

	status := normalizeFrontendBriefEnum(parseLooseLabel(section, "Status"))
	if status == "" {
		return errors.New("Do-Not Design Violation Check failed: missing Status field.")
	}
	switch status {
	case "passed", "pass", "clear", "complete", "completed", "success", "succeeded":
		if frontendViolationCheckUsesException(section) && !frontendViolationCheckCitesException(section) {
			return errors.New("Do-Not Design Violation Check failed: exception-path usage must cite the contract evidence that allows the exception.")
		}
		if report.ImagegenRequirement == frontendImagegenRequirementRequired {
			if err := frontendGeneratedAssetEvidenceFailure(output); err != nil {
				return err
			}
		}
		return nil
	case "failed", "fail", "blocked", "violation", "violated", "unresolved":
		pattern := firstNonBlank(parseLooseLabel(section, "Banned pattern"), parseLooseLabel(section, "Banned Pattern"), "unspecified banned pattern")
		remediation := firstNonBlank(parseLooseLabel(section, "Remediation"), parseLooseLabel(section, "Remediation path"), "follow the allowed replacement or exception path in `frontend-brief.md`")
		return fmt.Errorf("Do-Not Design Violation Check failed: %s. Remediation: %s", pattern, remediation)
	default:
		return fmt.Errorf("Do-Not Design Violation Check failed: unsupported Status %q.", status)
	}
}

func frontendGeneratedAssetEvidenceFailure(output string) error {
	section, ok := markdownSection(output, "Generated Asset Evidence", 2)
	if !ok {
		return errors.New("Generated Asset Evidence failed: frontend-major work with Imagegen requirement `required` must include `## Generated Asset Evidence` with manifest path and per-asset generated file, saved asset path, prompt summary, intended UI usage, and rendered usage evidence.")
	}
	if isPendingMarkdownValue(parseLooseLabel(section, "Manifest path")) {
		return errors.New("Generated Asset Evidence failed: missing non-pending manifest path.")
	}
	blocks := parseGeneratedAssetEvidenceBlocks(section)
	if len(blocks) == 0 {
		return errors.New("Generated Asset Evidence failed: missing repeatable Asset ID evidence blocks.")
	}
	for _, block := range blocks {
		for _, label := range []string{"Asset ID", "Generated file", "Saved asset path", "Prompt summary", "Intended UI usage", "Rendered usage evidence"} {
			if isPendingMarkdownValue(block.Fields[label]) {
				return fmt.Errorf("Generated Asset Evidence failed: Asset ID %q missing non-pending %s.", block.ID, strings.ToLower(label))
			}
		}
	}
	return nil
}

type generatedAssetEvidenceBlock struct {
	ID     string
	Fields map[string]string
}

func parseGeneratedAssetEvidenceBlocks(section string) []generatedAssetEvidenceBlock {
	var blocks []generatedAssetEvidenceBlock
	var current *generatedAssetEvidenceBlock
	for _, rawLine := range strings.Split(section, "\n") {
		line := strings.TrimSpace(trimMarkdownListMarker(rawLine))
		if line == "" {
			continue
		}
		label, value, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		label = strings.TrimSpace(label)
		value = strings.TrimSpace(value)
		if label == "Asset ID" {
			blocks = append(blocks, generatedAssetEvidenceBlock{
				ID:     value,
				Fields: map[string]string{"Asset ID": value},
			})
			current = &blocks[len(blocks)-1]
			continue
		}
		if current == nil {
			continue
		}
		switch label {
		case "Generated file", "Saved asset path", "Prompt summary", "Intended UI usage", "Rendered usage evidence":
			current.Fields[label] = value
		}
	}
	return blocks
}

func frontendViolationCheckUsesException(section string) bool {
	return strings.Contains(strings.ToLower(section), "exception")
}

func frontendViolationCheckCitesException(section string) bool {
	for _, label := range []string{"Exception path cited", "Exception evidence", "Contract evidence"} {
		if !isPendingMarkdownValue(parseLooseLabel(section, label)) {
			return true
		}
	}
	return false
}

func normalizeArtifactToken(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "" {
		return "turn"
	}
	var b strings.Builder
	lastDash := false
	for _, r := range value {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
			lastDash = false
			continue
		}
		if !lastDash {
			b.WriteByte('-')
			lastDash = true
		}
	}
	return strings.Trim(b.String(), "-")
}

func commandExitCode(err error) int {
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return exitErr.ExitCode()
	}
	if err != nil {
		return -1
	}
	return 0
}

func (a *App) runnerFor(cfg systemConfig) (runner, error) {
	switch normalizeRunner(cfg.Runner) {
	case "", "codex":
		return codexRunner{
			lookPath:  a.lookPath,
			runBinary: a.runBinary,
			runCmd:    a.runCodexCmdWithInput,
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
		teamContinuationMode = "explicit-thread-resume"
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

	var observedThreadID string
	for index, turnReq := range turnRequests {
		if index > 0 && observedThreadID != "" && turnReq.Model == req.Model {
			turnReq.ResumeSession = true
			turnReq.ThreadID = observedThreadID
		}
		turnResult, err := selectedRunner.Execute(ctx, turnReq, capabilities)
		result.Turns = append(result.Turns, turnResult)
		if observedThreadID == "" && turnResult.ThreadID != "" {
			observedThreadID = turnResult.ThreadID
			result.SessionID = observedThreadID
		}
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

	result.Output = joinExecutionOutputs(result.Turns)
	if violationErr := frontendViolationCheckFailure(projectRoot, req.SpecID, result.Output); violationErr != nil {
		result.FinishedAt = a.now().Format(time.RFC3339)
		result.Error = violationErr.Error()
		if writeErr := a.writeExecutionArtifacts(projectRoot, logID, result); writeErr != nil {
			return result, validationReport{}, writeErr
		}
		afterExecutionErr := hooks.Trigger(ctx, hookTrigger{
			Event:        hookEventAfterExecution,
			StageStatus:  "failed",
			ErrorSummary: violationErr.Error(),
			EventData: map[string]any{
				"execution_path": filepath.ToSlash(filepath.Join(logsDir, "runs", logID+"-execution.json")),
			},
		})
		if writeErr := writeRunEvidence("execution_failed", 0, result.Error); writeErr != nil {
			return result, validationReport{}, errors.Join(violationErr, afterExecutionErr, writeErr)
		}
		publishErr := publishProgress(
			"failed",
			"execution_failed",
			"Worker execution failed",
			violationErr.Error(),
			map[string]any{"session_id": logID},
		)
		if publishErr != nil {
			if writeErr := writeRunEvidenceWithProgressFailure("execution_failed", 0, result.Error); writeErr != nil {
				return result, validationReport{}, errors.Join(violationErr, publishErr, writeErr)
			}
		}
		return result, validationReport{}, errors.Join(violationErr, afterExecutionErr, publishErr)
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
		repairReq.ResumeSession = codexSessionStateful(req.SessionMode) && result.SessionID != "" && result.SessionID != logID
		repairReq.ThreadID = result.SessionID
		repairReq.TurnName = fmt.Sprintf("repair-%d", attempt)
		repairReq.TurnRole = req.DelegationPlan.IntegratorRole
		repairReq.Prompt = buildRepairPrompt(req, finalReport, attempt, !repairReq.ResumeSession)
		repairReq.RequestedReasoningEffort = "high"

		repairResult, repairErr := selectedRunner.Execute(ctx, repairReq, capabilities)
		result.Turns = append(result.Turns, repairResult)
		result.RetryCount++
		if repairReq.ResumeSession {
			result.SessionContinuity = "explicit-thread-resume"
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
			return result, finalReport, errors.Join(fmt.Errorf("%s", result.Error), publishErr, err)
		}
	}
	return result, finalReport, errors.Join(fmt.Errorf("%s", result.Error), publishErr)
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
	base.Phase = routingPhaseImplement
	base.RoutingDecision = modelRoutingDecision(modelRoutingInput{Phase: base.Phase, Role: base.TurnRole, RepairCount: base.RepairAttempts})
	if base.ModelRoutingPolicy == modelRoutingPolicyCostBalancedV1 {
		base.Model = base.RoutingDecision.Model
		base.RequestedReasoningEffort = base.RoutingDecision.ReasoningEffort
	}

	turns := []executionRequest{base}
	if normalizeExecutionMode(req.Mode) != executionModeTeam {
		return turns
	}

	stateful := codexSessionStateful(req.SessionMode)
	for _, profile := range req.DelegationPlan.SelectedRoleProfiles {
		turn := req
		turn.TurnName = roleTurnName(profile.Role)
		turn.TurnRole = profile.Role
		turn.Phase = routingPhaseForRole(profile.Role)
		turn.RoutingDecision = modelRoutingDecision(modelRoutingInput{Phase: turn.Phase, Role: profile.Role})
		// A continuation is legal only after an explicit UUID is observed from
		// the immediately preceding same-model turn.
		turn.ResumeSession = false
		turn.Model = firstNonBlank(profile.Model, req.Model)
		if turn.ModelRoutingPolicy == modelRoutingPolicyCostBalancedV1 {
			turn.Model = turn.RoutingDecision.Model
			turn.RequestedReasoningEffort = turn.RoutingDecision.ReasoningEffort
			if turn.RoutingDecision.ReadOnly {
				turn.SandboxMode = "read-only"
			}
		}
		turn.Profile = req.Profile
		if turn.RequestedReasoningEffort == "" {
			turn.RequestedReasoningEffort = profile.ModelReasoningEffort
		}
		turn.Prompt = buildDelegationTurnPrompt(turn, profile, !stateful)
		turns = append(turns, turn)
	}
	return turns
}

func routingPhaseForRole(role string) routingPhase {
	role = strings.TrimSpace(strings.ToLower(role))
	switch {
	case strings.Contains(role, "planner"):
		return routingPhasePlan
	case strings.Contains(role, "architect"):
		return routingPhaseArchitecture
	case strings.Contains(role, "designer"):
		return routingPhaseDesign
	case strings.Contains(role, "reviewer"):
		return routingPhaseReview
	case strings.Contains(role, "test"):
		return routingPhaseTest
	default:
		return routingPhaseImplement
	}
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
	if req.RoutingDecision.ReadOnly {
		return strings.Join([]string{
			fmt.Sprintf("Act as the read-only `%s` checkpoint for `%s`.", profile.Role, req.SpecID),
			"Do not edit files or run mutating commands.",
			"Return the decision, evidence, risks, open questions, and a Terra/Luna writer handoff.",
		}, "\n")
	}
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
