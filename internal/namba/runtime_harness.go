package namba

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const sessionRefreshNoticePath = ".namba/logs/session-refresh-required.json"

type codexConfig struct {
	Model           string
	Profile         string
	WebSearch       bool
	AddDirs         []string
	SessionMode     string
	RepairAttempts  int
	RequiredEnv     []string
	RequiresNetwork bool
}

type workflowConfig struct {
	DefaultParallel             bool
	MaxParallelWorkers          int
	ParallelAcceptanceThreshold int
}

type preflightReport struct {
	SpecID     string          `json:"spec_id"`
	Passed     bool            `json:"passed"`
	StartedAt  string          `json:"started_at"`
	FinishedAt string          `json:"finished_at"`
	Steps      []preflightStep `json:"steps"`
}

type preflightStep struct {
	Name   string `json:"name"`
	Passed bool   `json:"passed"`
	Detail string `json:"detail,omitempty"`
	Error  string `json:"error,omitempty"`
}

type outputWriteReport struct {
	ChangedPaths            []string `json:"changed_paths,omitempty"`
	InstructionSurfacePaths []string `json:"instruction_surface_paths,omitempty"`
}

type sessionRefreshNotice struct {
	Required    bool     `json:"required"`
	Reason      string   `json:"reason"`
	Paths       []string `json:"paths,omitempty"`
	GeneratedAt string   `json:"generated_at"`
}

func (a *App) loadCodexConfig(root string) (codexConfig, error) {
	values, err := readKeyValueFile(filepath.Join(root, configDir, "codex.yaml"))
	if err != nil {
		return codexConfig{}, err
	}

	cfg := codexConfig{
		Model:           strings.TrimSpace(values["model"]),
		Profile:         strings.TrimSpace(values["profile"]),
		WebSearch:       parseBoolValue(values["web_search"], false),
		AddDirs:         parseCommaSeparatedList(values["add_dirs"]),
		SessionMode:     normalizeSessionMode(values["session_mode"]),
		RepairAttempts:  maxInt(parseIntValue(values["repair_attempts"], 1), 0),
		RequiredEnv:     parseCommaSeparatedList(values["required_env"]),
		RequiresNetwork: parseBoolValue(values["requires_network"], false),
	}
	if cfg.SessionMode == "" {
		cfg.SessionMode = "stateful"
	}
	return cfg, nil
}

func (a *App) loadWorkflowConfig(root string) (workflowConfig, error) {
	values, err := readKeyValueFile(filepath.Join(root, configDir, "workflow.yaml"))
	if err != nil {
		return workflowConfig{}, err
	}

	cfg := workflowConfig{
		DefaultParallel:             parseBoolValue(values["default_parallel"], false),
		MaxParallelWorkers:          maxInt(parseIntValue(values["max_parallel_workers"], 3), 1),
		ParallelAcceptanceThreshold: maxInt(parseIntValue(values["parallel_acceptance_threshold"], 3), 1),
	}
	return cfg, nil
}

func resolveCodexRuntimeForMode(cfg codexConfig, mode executionMode) codexConfig {
	cfg.SessionMode = firstNonBlank(normalizeSessionMode(cfg.SessionMode), "stateful")
	switch normalizeExecutionMode(mode) {
	case executionModeSolo:
		if cfg.SessionMode == "stateful" {
			cfg.SessionMode = "solo"
		}
	case executionModeTeam:
		if cfg.SessionMode == "stateful" {
			cfg.SessionMode = "team"
		}
	case executionModeParallel:
		if cfg.SessionMode == "stateful" {
			cfg.SessionMode = "parallel-worker"
		}
	default:
		if cfg.SessionMode == "stateful" {
			cfg.SessionMode = "stateful"
		}
	}
	return cfg
}

func normalizeSessionMode(value string) string {
	switch strings.TrimSpace(strings.ToLower(value)) {
	case "", "stateful":
		return "stateful"
	case "solo":
		return "solo"
	case "team":
		return "team"
	case "parallel-worker":
		return "parallel-worker"
	case "ephemeral":
		return "ephemeral"
	default:
		return strings.TrimSpace(strings.ToLower(value))
	}
}

func codexSessionStateful(mode string) bool {
	return normalizeSessionMode(mode) != "ephemeral"
}

func parseIntValue(raw string, fallback int) int {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return fallback
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}
	return value
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func resolveRuntimeAddDirs(base string, dirs []string) ([]string, error) {
	if len(dirs) == 0 {
		return nil, nil
	}

	resolved := make([]string, 0, len(dirs))
	seen := make(map[string]bool, len(dirs))
	for _, dir := range dirs {
		dir = strings.TrimSpace(dir)
		if dir == "" {
			continue
		}
		path := dir
		if !filepath.IsAbs(path) {
			path = filepath.Join(base, filepath.FromSlash(dir))
		}
		path = filepath.Clean(path)
		info, err := os.Stat(path)
		if err != nil {
			return nil, fmt.Errorf("add_dir %q: %w", dir, err)
		}
		if !info.IsDir() {
			return nil, fmt.Errorf("add_dir %q is not a directory", dir)
		}
		if seen[path] {
			continue
		}
		seen[path] = true
		resolved = append(resolved, path)
	}
	return resolved, nil
}

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
