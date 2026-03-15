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
	Runner       string
	ApprovalMode string
	SandboxMode  string
}

type executionRequest struct {
	SpecID       string
	WorkDir      string
	Prompt       string
	Runner       string
	ApprovalMode string
	SandboxMode  string
}

type executionResult struct {
	Runner       string `json:"runner"`
	SpecID       string `json:"spec_id"`
	WorkDir      string `json:"work_dir"`
	ApprovalMode string `json:"approval_mode"`
	SandboxMode  string `json:"sandbox_mode"`
	Output       string `json:"output"`
	Succeeded    bool   `json:"succeeded"`
	StartedAt    string `json:"started_at"`
	FinishedAt   string `json:"finished_at"`
	Error        string `json:"error,omitempty"`
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
		Runner:       normalizeRunner(req.Runner),
		SpecID:       req.SpecID,
		WorkDir:      req.WorkDir,
		ApprovalMode: normalizeApprovalMode(req.ApprovalMode),
		SandboxMode:  normalizeSandboxMode(req.SandboxMode),
		StartedAt:    r.now().Format(time.RFC3339),
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
	approval := normalizeApprovalMode(req.ApprovalMode)
	if !isAllowedApprovalMode(approval) {
		return nil, fmt.Errorf("approval_mode %q is not supported", req.ApprovalMode)
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
		Runner:       normalizeRunner(values["runner"]),
		ApprovalMode: normalizeApprovalMode(values["approval_mode"]),
		SandboxMode:  normalizeSandboxMode(values["sandbox_mode"]),
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

func (a *App) newExecutionRequest(specID, workDir, prompt string, cfg systemConfig) executionRequest {
	return executionRequest{
		SpecID:       specID,
		WorkDir:      workDir,
		Prompt:       prompt,
		Runner:       normalizeRunner(cfg.Runner),
		ApprovalMode: normalizeApprovalMode(cfg.ApprovalMode),
		SandboxMode:  normalizeSandboxMode(cfg.SandboxMode),
	}
}

func (a *App) executeRun(ctx context.Context, projectRoot, logID string, req executionRequest, validationRoot string, cfg qualityConfig) (executionResult, validationReport, error) {
	selectedRunner, err := a.runnerFor(systemConfig{Runner: req.Runner})
	if err != nil {
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

func normalizeApprovalMode(value string) string {
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

func isAllowedApprovalMode(value string) bool {
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
