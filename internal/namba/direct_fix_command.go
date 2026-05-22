package namba

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
)

func (a *App) executeDirectFix(ctx context.Context, root, description string) error {
	fixCtx, err := a.loadDirectFixExecutionContext(root, description)
	if err != nil {
		return err
	}
	if err := a.materializeDirectFixExecutionPrompt(fixCtx); err != nil {
		return err
	}
	return a.dispatchDirectFixExecution(ctx, fixCtx)
}

func buildDirectFixPrompt(root, description string, projectCfg projectConfig, qualityCfg qualityConfig) (string, delegationPlan) {
	delegation := suggestDelegationPlan(executionModeDefault, description, "", "- [ ] Add targeted regression coverage\n- [ ] Validation commands pass")
	lines := []string{
		"# NambaAI Direct Repair Request",
		"",
		"Repair the reported issue directly in the current workspace without creating a SPEC package.",
		"",
		"## Issue",
		description,
		"",
		"## Repair Contract",
	}
	lines = append(lines, directFixRepairContractLines()...)
	lines = append(lines, "")
	lines = append(lines, directFixProjectContextPromptLines(projectCfg, qualityCfg)...)
	lines = append(lines, "")
	lines = append(lines, formatDelegationPlanPrompt(delegation)...)
	lines = append(lines, "")
	lines = append(lines, directFixValidationPromptLines(qualityCfg)...)
	lines = append(lines, "", fmt.Sprintf("Project root: %s", root))
	return strings.Join(lines, "\n"), delegation
}

func directFixRepairContractLines() []string {
	return []string{
		"- Inspect the relevant repository files plus `.namba/config/sections/*.yaml` and `.namba/project/*` context before editing.",
		"- Implement the smallest safe fix that resolves the reported issue in the current workspace.",
		"- Add targeted regression coverage for the affected area.",
		"- Run the configured validation commands from `.namba/config/sections/quality.yaml`.",
		"- Finish by running `namba sync` in the same workspace after validation passes.",
		"- Do not create or mutate `.namba/specs/<SPEC>` as part of this direct repair flow.",
		"- For bugfix SPEC scaffolding, use `namba fix --command plan \"<issue description>\"`.",
	}
}

func directFixProjectContextPromptLines(projectCfg projectConfig, qualityCfg qualityConfig) []string {
	return []string{
		"## Project Context",
		fmt.Sprintf("- Project: %s", firstNonBlank(projectCfg.Name, "unknown")),
		fmt.Sprintf("- Project type: %s", firstNonBlank(projectCfg.ProjectType, "unknown")),
		fmt.Sprintf("- Language: %s", firstNonBlank(projectCfg.Language, "unknown")),
		fmt.Sprintf("- Framework: %s", firstNonBlank(projectCfg.Framework, "unknown")),
		fmt.Sprintf("- Development mode: %s", firstNonBlank(qualityCfg.DevelopmentMode, "unknown")),
	}
}

func directFixValidationPromptLines(qualityCfg qualityConfig) []string {
	lines := []string{"## Validation"}
	for _, step := range validationPipelineSteps(qualityCfg) {
		command := strings.TrimSpace(step.Command)
		if command == "" || command == "none" {
			continue
		}
		lines = append(lines, fmt.Sprintf("- %s: %s", step.Name, command))
	}
	return lines
}

func (a *App) loadDirectFixExecutionContext(root, description string) (directFixExecutionContext, error) {
	projectCfg, err := a.loadProjectConfig(root)
	if err != nil {
		return directFixExecutionContext{}, err
	}
	runtimeCfg, err := a.loadExecutionRuntimeConfig(root)
	if err != nil {
		return directFixExecutionContext{}, err
	}

	prompt, delegation := buildDirectFixPrompt(root, description, projectCfg, runtimeCfg.QualityCfg)
	logID := "direct-fix"
	return directFixExecutionContext{
		Root:        root,
		Description: description,
		QualityCfg:  runtimeCfg.QualityCfg,
		SystemCfg:   runtimeCfg.SystemCfg,
		CodexCfg:    runtimeCfg.CodexCfg,
		Prompt:      prompt,
		PromptPath:  filepath.Join(root, logsDir, "runs", logID+"-request.md"),
		LogID:       logID,
		Delegation:  delegation,
	}, nil
}

func (a *App) materializeDirectFixExecutionPrompt(fixCtx directFixExecutionContext) error {
	return a.writeExecutionPrompt(fixCtx.PromptPath, fixCtx.Prompt)
}

func (a *App) dispatchDirectFixExecution(ctx context.Context, fixCtx directFixExecutionContext) error {
	request := a.newExecutionRequest("DIRECT-FIX", fixCtx.Root, fixCtx.Prompt, executionModeDefault, fixCtx.Delegation, fixCtx.SystemCfg, fixCtx.CodexCfg)
	request.TurnName = fixCtx.LogID
	request.TurnRole = fixCtx.Delegation.IntegratorRole
	if _, _, err := a.executeRun(ctx, fixCtx.Root, fixCtx.LogID, request, fixCtx.Root, fixCtx.QualityCfg, nil, ""); err != nil {
		return err
	}
	if err := a.runSync(ctx, nil); err != nil {
		return err
	}

	fmt.Fprintf(a.stdout, "Executed direct fix with %s\n", request.Runner)
	return nil
}
