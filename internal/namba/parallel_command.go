package namba

import (
	"context"
	"errors"
	"strings"
)

type parallelWorkerState struct {
	name   string
	path   string
	branch string
	err    error
}

func (a *App) runParallel(ctx context.Context, root string, specPkg specPackage, tasks []string, prompt string, qualityCfg qualityConfig, systemCfg systemConfig, codexCfg codexConfig, workflowCfg workflowConfig, dryRun bool) error {
	return a.executeParallelRun(ctx, root, specPkg, tasks, prompt, qualityCfg, systemCfg, codexCfg, workflowCfg, dryRun)
}

func (a *App) runCodexExec(ctx context.Context, dir, prompt string) (string, error) {
	if _, err := a.lookPath("codex"); err != nil {
		return "", errors.New("codex is not installed")
	}
	req := executionRequest{
		WorkDir:        dir,
		Prompt:         prompt,
		Runner:         "codex",
		ApprovalPolicy: "on-request",
		SandboxMode:    "workspace-write",
		SessionMode:    "stateful",
	}
	capabilities, err := a.codexCapabilities(ctx, dir, req)
	if err != nil {
		return "", err
	}
	args, err := buildCodexExecArgs(req, capabilities)
	if err != nil {
		return "", err
	}
	if a.runCodexCmdWithInput != nil {
		stdout, stderr, err := a.runCodexCmdWithInput(ctx, "codex", args, dir, prompt)
		return strings.TrimSpace(strings.Join(nonEmptyArgs([]string{stdout, stderr}), "\n")), err
	}
	return a.runBinary(ctx, "codex", args, dir)
}
