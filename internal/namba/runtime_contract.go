package namba

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const sessionRefreshNoticePath = ".namba/logs/session-refresh-required.json"

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
