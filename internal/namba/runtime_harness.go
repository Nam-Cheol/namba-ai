package namba

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func (a *App) runPreflight(ctx context.Context, req executionRequest) (preflightReport, codexCapabilityMatrix, error) {
	report := preflightReport{
		SpecID:    req.SpecID,
		Passed:    true,
		StartedAt: a.now().Format(time.RFC3339),
	}
	var capabilities codexCapabilityMatrix

	addStep := func(step preflightStep) {
		if !step.Passed {
			report.Passed = false
		}
		report.Steps = append(report.Steps, step)
	}

	workDir := filepath.Clean(req.WorkDir)
	if info, err := os.Stat(workDir); err != nil {
		addStep(preflightStep{Name: "project_root", Error: err.Error()})
	} else if !info.IsDir() {
		addStep(preflightStep{Name: "project_root", Error: fmt.Sprintf("%s is not a directory", workDir)})
	} else {
		addStep(preflightStep{Name: "project_root", Passed: true, Detail: workDir})
	}

	if _, err := a.lookPath("codex"); err != nil {
		addStep(preflightStep{Name: "codex", Error: err.Error()})
	} else {
		addStep(preflightStep{Name: "codex", Passed: true, Detail: "codex available"})
		detected, err := a.codexCapabilities(ctx, workDir, req)
		if err != nil {
			addStep(preflightStep{Name: "codex_cli_capabilities", Error: err.Error()})
		} else {
			capabilities = detected
			addStep(preflightStep{Name: "codex_cli_capabilities", Passed: true, Detail: firstNonBlank(capabilities.Version, "capabilities detected")})
			detail, contractErr := validateCodexExecutionContract(req, capabilities)
			if contractErr != nil {
				addStep(preflightStep{Name: "codex_cli_contract", Error: contractErr.Error()})
			} else {
				addStep(preflightStep{Name: "codex_cli_contract", Passed: true, Detail: detail})
			}
		}
	}

	if normalizeExecutionMode(req.Mode) == executionModeParallel || isGitRepository(workDir) {
		if _, err := a.lookPath("git"); err != nil {
			addStep(preflightStep{Name: "git", Error: err.Error()})
		} else {
			addStep(preflightStep{Name: "git", Passed: true, Detail: "git available"})
		}
	}

	resolvedAddDirs, err := resolveRuntimeAddDirs(workDir, req.AddDirs)
	if err != nil {
		addStep(preflightStep{Name: "add_dirs", Error: err.Error()})
	} else if len(resolvedAddDirs) == 0 {
		addStep(preflightStep{Name: "add_dirs", Passed: true, Detail: "none"})
	} else {
		addStep(preflightStep{Name: "add_dirs", Passed: true, Detail: strings.Join(resolvedAddDirs, ", ")})
	}

	missingEnv := make([]string, 0)
	for _, key := range req.RequiredEnv {
		if strings.TrimSpace(key) == "" {
			continue
		}
		if strings.TrimSpace(a.getenv(key)) == "" {
			missingEnv = append(missingEnv, key)
		}
	}
	if len(missingEnv) > 0 {
		addStep(preflightStep{Name: "required_env", Error: "missing " + strings.Join(missingEnv, ", ")})
	} else if len(req.RequiredEnv) > 0 {
		addStep(preflightStep{Name: "required_env", Passed: true, Detail: strings.Join(req.RequiredEnv, ", ")})
	}

	if req.RequiresNetwork {
		addStep(preflightStep{Name: "network", Passed: true, Detail: "run declares network access requirements"})
	}

	report.FinishedAt = a.now().Format(time.RFC3339)
	if !report.Passed {
		for _, step := range report.Steps {
			if step.Error != "" {
				return report, capabilities, fmt.Errorf("preflight failed at %s: %s", step.Name, step.Error)
			}
		}
		return report, capabilities, fmt.Errorf("preflight failed")
	}

	return report, capabilities, nil
}

func isInstructionSurfacePath(rel string) bool {
	switch {
	case rel == "AGENTS.md":
		return true
	case rel == repoCodexConfigPath:
		return true
	case strings.HasPrefix(rel, repoSkillsDir+"/"):
		return true
	case strings.HasPrefix(rel, repoCodexAgentsDir+"/"):
		return true
	case strings.HasPrefix(rel, codexStateDir+"/"):
		return true
	default:
		return false
	}
}

func (a *App) writeSessionRefreshNotice(root string, report outputWriteReport) error {
	if len(report.InstructionSurfacePaths) == 0 {
		return nil
	}
	notice := sessionRefreshNotice{
		Required:    true,
		Reason:      "Generated instruction surfaces changed. Start a fresh Codex session before continuing a long team or repair run.",
		Paths:       report.InstructionSurfacePaths,
		GeneratedAt: a.now().Format(time.RFC3339),
	}
	return writeJSONFile(filepath.Join(root, filepath.FromSlash(sessionRefreshNoticePath)), notice)
}
