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
	SpecID         string         `json:"spec_id"`
	WorkDir        string         `json:"work_dir"`
	Prompt         string         `json:"prompt"`
	Mode           executionMode  `json:"mode"`
	Runner         string         `json:"runner"`
	ApprovalPolicy string         `json:"approval_policy"`
	SandboxMode    string         `json:"sandbox_mode"`
	DelegationPlan delegationPlan `json:"delegation_plan,omitempty"`
}

type executionResult struct {
	Runner             string         `json:"runner"`
	SpecID             string         `json:"spec_id"`
	WorkDir            string         `json:"work_dir"`
	ExecutionMode      string         `json:"execution_mode"`
	ApprovalPolicy     string         `json:"approval_policy"`
	SandboxMode        string         `json:"sandbox_mode"`
	DelegationPlan     delegationPlan `json:"delegation_plan,omitempty"`
	DelegationObserved bool           `json:"delegation_observed"`
	DelegationSummary  string         `json:"delegation_summary,omitempty"`
	Output             string         `json:"output"`
	Succeeded          bool           `json:"succeeded"`
	StartedAt          string         `json:"started_at"`
	FinishedAt         string         `json:"finished_at"`
	Error              string         `json:"error,omitempty"`
}

type validationReport struct {
	SpecID     string           `json:"spec_id"`
	Passed     bool             `json:"passed"`
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
	Execute(context.Context, executionRequest) (executionResult, error)
}

type codexRunner struct {
	lookPath  func(string) (string, error)
	runBinary func(context.Context, string, []string, string) (string, error)
	now       func() time.Time
}

func (r codexRunner) Execute(ctx context.Context, req executionRequest) (executionResult, error) {
	result := executionResult{
		Runner:             normalizeRunner(req.Runner),
		SpecID:             req.SpecID,
		WorkDir:            req.WorkDir,
		ExecutionMode:      string(normalizeExecutionMode(req.Mode)),
		ApprovalPolicy:     normalizeApprovalPolicy(req.ApprovalPolicy),
		SandboxMode:        normalizeSandboxMode(req.SandboxMode),
		DelegationPlan:     req.DelegationPlan,
		DelegationObserved: false,
		DelegationSummary:  summarizeDelegationPlan(req.DelegationPlan),
		StartedAt:          r.now().Format(time.RFC3339),
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

	return []string{"exec", "-a", approval, "-s", sandbox, req.Prompt}, nil
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

func (a *App) newExecutionRequest(specID, workDir, prompt string, mode executionMode, plan delegationPlan, cfg systemConfig) executionRequest {
	return executionRequest{
		SpecID:         specID,
		WorkDir:        workDir,
		Prompt:         prompt,
		Mode:           normalizeExecutionMode(mode),
		Runner:         normalizeRunner(cfg.Runner),
		ApprovalPolicy: normalizeApprovalPolicy(cfg.ApprovalPolicy),
		SandboxMode:    normalizeSandboxMode(cfg.SandboxMode),
		DelegationPlan: plan,
	}
}

func (a *App) executeRun(ctx context.Context, projectRoot, logID string, req executionRequest, validationRoot string, cfg qualityConfig) (executionResult, validationReport, error) {
	selectedRunner, err := a.runnerFor(systemConfig{Runner: req.Runner})
	if err != nil {
		return executionResult{}, validationReport{}, err
	}
	if err := writeJSONFile(filepath.Join(projectRoot, logsDir, "runs", logID+"-request.json"), req); err != nil {
		return executionResult{}, validationReport{}, err
	}

	result, execErr := selectedRunner.Execute(ctx, req)
	if err := writeRunText(filepath.Join(projectRoot, logsDir, "runs", logID+"-result.txt"), result.Output); err != nil {
		return result, validationReport{}, err
	}
	if err := writeJSONFile(filepath.Join(projectRoot, logsDir, "runs", logID+"-execution.json"), result); err != nil {
		return result, validationReport{}, err
	}
	if execErr != nil {
		return result, validationReport{}, execErr
	}

	report, validationErr := a.runValidationReport(ctx, validationRoot, cfg, req.SpecID)
	if err := writeJSONFile(filepath.Join(projectRoot, logsDir, "runs", logID+"-validation.json"), report); err != nil {
		if validationErr != nil {
			return result, report, fmt.Errorf("%w; write validation log: %v", validationErr, err)
		}
		return result, report, err
	}
	if validationErr != nil {
		return result, report, validationErr
	}

	return result, report, nil
}

func (a *App) runValidationReport(ctx context.Context, root string, cfg qualityConfig, specID string) (validationReport, error) {
	report := validationReport{
		SpecID:    specID,
		Passed:    true,
		StartedAt: a.now().Format(time.RFC3339),
	}

	steps := []validationStep{
		{Name: "test", Command: strings.TrimSpace(cfg.TestCommand)},
		{Name: "lint", Command: strings.TrimSpace(cfg.LintCommand)},
		{Name: "typecheck", Command: strings.TrimSpace(cfg.TypecheckCommand)},
	}

	for _, step := range steps {
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
