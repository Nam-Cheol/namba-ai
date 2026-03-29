package namba

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type systemConfig struct {
	Runner         string
	ApprovalPolicy string
	SandboxMode    string
}

type executionMode string

const (
	executionModeDefault  executionMode = "default"
	executionModeSolo     executionMode = "solo"
	executionModeTeam     executionMode = "team"
	executionModeParallel executionMode = "parallel"
)

type delegationPlan struct {
	DominantDomains      []string              `json:"dominant_domains,omitempty"`
	SelectedRoles        []string              `json:"selected_roles,omitempty"`
	SelectedRoleProfiles []agentRuntimeProfile `json:"selected_role_profiles,omitempty"`
	DelegationBudget     int                   `json:"delegation_budget,omitempty"`
	IntegratorRole       string                `json:"integrator_role,omitempty"`
	ReviewerRole         string                `json:"reviewer_role,omitempty"`
	RoutingRationale     []string              `json:"routing_rationale,omitempty"`
}

type executionRequest struct {
	SpecID                   string         `json:"spec_id"`
	WorkDir                  string         `json:"work_dir"`
	Prompt                   string         `json:"prompt"`
	Mode                     executionMode  `json:"mode"`
	Runner                   string         `json:"runner"`
	ApprovalPolicy           string         `json:"approval_policy"`
	SandboxMode              string         `json:"sandbox_mode"`
	Model                    string         `json:"model,omitempty"`
	Profile                  string         `json:"profile,omitempty"`
	WebSearch                bool           `json:"web_search,omitempty"`
	AddDirs                  []string       `json:"add_dirs,omitempty"`
	SessionMode              string         `json:"session_mode,omitempty"`
	RepairAttempts           int            `json:"repair_attempts,omitempty"`
	RequiredEnv              []string       `json:"required_env,omitempty"`
	RequiresNetwork          bool           `json:"requires_network,omitempty"`
	DelegationPlan           delegationPlan `json:"delegation_plan,omitempty"`
	TurnName                 string         `json:"turn_name,omitempty"`
	TurnRole                 string         `json:"turn_role,omitempty"`
	RequestedReasoningEffort string         `json:"requested_reasoning_effort,omitempty"`
	ResumeSession            bool           `json:"resume_session,omitempty"`
}

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
	Execute(context.Context, executionRequest) (executionTurnResult, error)
}

type codexRunner struct {
	lookPath  func(string) (string, error)
	runBinary func(context.Context, string, []string, string) (string, error)
	now       func() time.Time
}

func (r codexRunner) Execute(ctx context.Context, req executionRequest) (executionTurnResult, error) {
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

	args, err := buildCodexExecArgs(req)
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

func buildCodexExecArgs(req executionRequest) ([]string, error) {
	approval := normalizeApprovalPolicy(req.ApprovalPolicy)
	if !isAllowedApprovalPolicy(approval) {
		return nil, fmt.Errorf("approval_policy %q is not supported", req.ApprovalPolicy)
	}

	sandbox := normalizeSandboxMode(req.SandboxMode)
	if !isAllowedSandboxMode(sandbox) {
		return nil, fmt.Errorf("sandbox_mode %q is not supported", req.SandboxMode)
	}

	sessionMode := normalizeSessionMode(req.SessionMode)
	if sessionMode == "" {
		sessionMode = "stateful"
	}
	if req.ResumeSession && !codexSessionStateful(sessionMode) {
		return nil, fmt.Errorf("session_mode %q does not support resume", sessionMode)
	}

	args := []string{"exec", "-a", approval, "-s", sandbox}
	if model := strings.TrimSpace(req.Model); model != "" {
		args = append(args, "-m", model)
	}
	if profile := strings.TrimSpace(req.Profile); profile != "" {
		args = append(args, "-p", profile)
	}
	if req.WebSearch {
		args = append(args, "--search")
	}
	for _, dir := range req.AddDirs {
		if trimmed := strings.TrimSpace(dir); trimmed != "" {
			args = append(args, "--add-dir", trimmed)
		}
	}

	if req.ResumeSession {
		args = append(args, "resume", "--last")
	} else if !codexSessionStateful(sessionMode) {
		args = append(args, "--ephemeral")
	}
	args = append(args, req.Prompt)
	return args, nil
}

func (a *App) loadSystemConfig(root string) (systemConfig, error) {
	values, err := readKeyValueFile(filepath.Join(root, configDir, "system.yaml"))
	if err != nil {
		return systemConfig{}, err
	}

	return systemConfig{
		Runner:         normalizeRunner(values["runner"]),
		ApprovalPolicy: normalizeApprovalPolicy(firstNonBlank(values["approval_policy"], values["approval_mode"])),
		SandboxMode:    normalizeSandboxMode(values["sandbox_mode"]),
	}, nil
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

func (a *App) newExecutionRequest(specID, workDir, prompt string, mode executionMode, plan delegationPlan, systemCfg systemConfig, codexCfg codexConfig) executionRequest {
	runtimeCfg := resolveCodexRuntimeForMode(codexCfg, mode)
	return executionRequest{
		SpecID:          specID,
		WorkDir:         workDir,
		Prompt:          prompt,
		Mode:            normalizeExecutionMode(mode),
		Runner:          normalizeRunner(systemCfg.Runner),
		ApprovalPolicy:  normalizeApprovalPolicy(systemCfg.ApprovalPolicy),
		SandboxMode:     normalizeSandboxMode(systemCfg.SandboxMode),
		Model:           runtimeCfg.Model,
		Profile:         runtimeCfg.Profile,
		WebSearch:       runtimeCfg.WebSearch,
		AddDirs:         append([]string(nil), runtimeCfg.AddDirs...),
		SessionMode:     runtimeCfg.SessionMode,
		RepairAttempts:  runtimeCfg.RepairAttempts,
		RequiredEnv:     append([]string(nil), runtimeCfg.RequiredEnv...),
		RequiresNetwork: runtimeCfg.RequiresNetwork,
		DelegationPlan:  plan,
	}
}

func (a *App) executeRun(ctx context.Context, projectRoot, logID string, req executionRequest, validationRoot string, cfg qualityConfig) (executionResult, validationReport, error) {
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

	preflight, preflightErr := a.runPreflight(ctx, req)
	if err := writeJSONFile(filepath.Join(projectRoot, logsDir, "runs", logID+"-preflight.json"), preflight); err != nil {
		return result, validationReport{}, err
	}
	if err := writeJSONFile(filepath.Join(projectRoot, logsDir, "runs", logID+"-request.json"), req); err != nil {
		return result, validationReport{}, err
	}
	if preflightErr != nil {
		result.FinishedAt = a.now().Format(time.RFC3339)
		result.Error = preflightErr.Error()
		if err := writeJSONFile(filepath.Join(projectRoot, logsDir, "runs", logID+"-execution.json"), result); err != nil {
			return result, validationReport{}, err
		}
		return result, validationReport{}, preflightErr
	}

	turnRequests := buildExecutionTurnRequests(req)
	teamContinuationMode := "degraded-fresh-exec"
	if codexSessionStateful(req.SessionMode) {
		teamContinuationMode = "codex-exec-resume-last"
	}

	for _, turnReq := range turnRequests {
		turnResult, err := selectedRunner.Execute(ctx, turnReq)
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
			return result, validationReport{}, err
		}
	}

	var finalReport validationReport
	maxAttempts := maxInt(req.RepairAttempts, 0) + 1
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		result.ValidationAttempts = attempt
		report, validationErr := a.runValidationReport(ctx, validationRoot, cfg, req.SpecID, attempt)
		finalReport = report
		if err := writeJSONFile(filepath.Join(projectRoot, logsDir, "runs", fmt.Sprintf("%s-validation-attempt-%d.json", logID, attempt)), report); err != nil {
			return result, finalReport, err
		}
		if validationErr == nil {
			result.Output = joinExecutionOutputs(result.Turns)
			result.Succeeded = true
			result.FinishedAt = a.now().Format(time.RFC3339)
			if err := a.writeExecutionArtifacts(projectRoot, logID, result); err != nil {
				return result, finalReport, err
			}
			if err := writeJSONFile(filepath.Join(projectRoot, logsDir, "runs", logID+"-validation.json"), finalReport); err != nil {
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
			return result, finalReport, validationErr
		}

		repairReq := req
		repairReq.ResumeSession = codexSessionStateful(req.SessionMode)
		repairReq.TurnName = fmt.Sprintf("repair-%d", attempt)
		repairReq.TurnRole = req.DelegationPlan.IntegratorRole
		repairReq.Prompt = buildRepairPrompt(req, finalReport, attempt, !repairReq.ResumeSession)
		repairReq.RequestedReasoningEffort = ""

		repairResult, repairErr := selectedRunner.Execute(ctx, repairReq)
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
			return result, finalReport, repairErr
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
	return result, finalReport, fmt.Errorf(result.Error)
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

func normalizeRunner(name string) string {
	normalized := strings.TrimSpace(strings.ToLower(name))
	if normalized == "" {
		return "codex"
	}
	return normalized
}

func normalizeApprovalPolicy(value string) string {
	normalized := strings.TrimSpace(strings.ToLower(value))
	if normalized == "" {
		return "on-request"
	}
	return normalized
}

func normalizeSandboxMode(value string) string {
	normalized := strings.TrimSpace(strings.ToLower(value))
	if normalized == "" {
		return "workspace-write"
	}
	return normalized
}

func normalizeExecutionMode(mode executionMode) executionMode {
	switch executionMode(strings.TrimSpace(strings.ToLower(string(mode)))) {
	case executionModeSolo:
		return executionModeSolo
	case executionModeTeam:
		return executionModeTeam
	case executionModeParallel:
		return executionModeParallel
	default:
		return executionModeDefault
	}
}

func isAllowedApprovalPolicy(value string) bool {
	switch value {
	case "untrusted", "on-failure", "on-request", "never":
		return true
	default:
		return false
	}
}

func isAllowedSandboxMode(value string) bool {
	switch value {
	case "read-only", "workspace-write", "danger-full-access":
		return true
	default:
		return false
	}
}
