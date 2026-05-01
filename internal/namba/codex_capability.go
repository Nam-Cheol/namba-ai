package namba

import (
	"context"
	"fmt"
	"strconv"
	"strings"
)

type codexCommandCapabilities struct {
	Config        bool `json:"config"`
	ApprovalFlag  bool `json:"approval_flag"`
	SandboxFlag   bool `json:"sandbox_flag"`
	ModelFlag     bool `json:"model_flag"`
	ProfileFlag   bool `json:"profile_flag"`
	WebSearch     bool `json:"web_search_flag"`
	AddDirFlag    bool `json:"add_dir_flag"`
	EphemeralFlag bool `json:"ephemeral_flag"`
	JSONFlag      bool `json:"json_flag"`
}

type codexCapabilityMatrix struct {
	Version string                   `json:"version,omitempty"`
	Exec    codexCommandCapabilities `json:"exec"`
	Resume  codexCommandCapabilities `json:"resume"`
}

type resolvedCodexInvocation struct {
	CommandShape    string
	Args            []string
	DirectFlags     []string
	ConfigOverrides []string
}

func (a *App) codexCapabilities(ctx context.Context, dir string, req executionRequest) (codexCapabilityMatrix, error) {
	if a.detectCodexCapabilities != nil {
		return a.detectCodexCapabilities(ctx, dir, req)
	}
	return a.probeCodexCapabilities(ctx, dir, req)
}

func (a *App) probeCodexCapabilities(ctx context.Context, dir string, req executionRequest) (codexCapabilityMatrix, error) {
	if _, err := a.lookPath("codex"); err != nil {
		return codexCapabilityMatrix{}, err
	}

	version, err := a.runBinary(ctx, "codex", []string{"--version"}, dir)
	if err != nil {
		return codexCapabilityMatrix{}, fmt.Errorf("codex --version: %w", err)
	}
	execHelp, err := a.runBinary(ctx, "codex", []string{"exec", "--help"}, dir)
	if err != nil {
		return codexCapabilityMatrix{}, fmt.Errorf("codex exec --help: %w", err)
	}
	matrix := codexCapabilityMatrix{
		Version: strings.TrimSpace(version),
		Exec:    parseCodexCommandCapabilities(execHelp),
	}
	if !plannedInvocationsNeedResume(plannedCodexRequests(req)) {
		return matrix, nil
	}

	resumeHelp, err := a.runBinary(ctx, "codex", []string{"exec", "resume", "--help"}, dir)
	if err != nil {
		return codexCapabilityMatrix{}, fmt.Errorf("codex exec resume --help: %w", err)
	}
	matrix.Resume = parseCodexCommandCapabilities(resumeHelp)
	return matrix, nil
}

func parseCodexCommandCapabilities(help string) codexCommandCapabilities {
	return codexCommandCapabilities{
		Config:        commandHelpContains(help, "--config"),
		ApprovalFlag:  commandHelpContains(help, "--ask-for-approval"),
		SandboxFlag:   commandHelpContains(help, "--sandbox"),
		ModelFlag:     commandHelpContains(help, "--model"),
		ProfileFlag:   commandHelpContains(help, "--profile"),
		WebSearch:     commandHelpContains(help, "--search"),
		AddDirFlag:    commandHelpContains(help, "--add-dir"),
		EphemeralFlag: commandHelpContains(help, "--ephemeral"),
		JSONFlag:      commandHelpContains(help, "--json"),
	}
}

func commandHelpContains(help, needle string) bool {
	return strings.Contains(help, needle)
}

func plannedInvocationsNeedResume(planned []executionRequest) bool {
	for _, req := range planned {
		if req.ResumeSession {
			return true
		}
	}
	return false
}

func validateCodexExecutionContract(req executionRequest, capabilities codexCapabilityMatrix) (string, error) {
	invocations, err := resolvePlannedCodexInvocations(req, capabilities)
	if err != nil {
		return "", err
	}

	seen := make(map[string]bool, len(invocations))
	summaries := make([]string, 0, len(invocations))
	for _, invocation := range invocations {
		summary := formatCodexInvocationSummary(invocation)
		if seen[summary] {
			continue
		}
		seen[summary] = true
		summaries = append(summaries, summary)
	}

	if len(summaries) == 0 {
		return firstNonBlank(capabilities.Version, "codex capabilities detected"), nil
	}
	if capabilities.Version == "" {
		return strings.Join(summaries, "; "), nil
	}
	return capabilities.Version + " | " + strings.Join(summaries, "; "), nil
}

func resolvePlannedCodexInvocations(req executionRequest, capabilities codexCapabilityMatrix) ([]resolvedCodexInvocation, error) {
	planned := plannedCodexRequests(req)
	invocations := make([]resolvedCodexInvocation, 0, len(planned))
	for _, plannedReq := range planned {
		invocation, err := resolveCodexInvocation(plannedReq, capabilities)
		if err != nil {
			return nil, err
		}
		invocations = append(invocations, invocation)
	}
	return invocations, nil
}

func plannedCodexRequests(req executionRequest) []executionRequest {
	planned := buildExecutionTurnRequests(req)
	if req.RepairAttempts > 0 && codexSessionStateful(req.SessionMode) {
		repairReq := req
		repairReq.ResumeSession = true
		repairReq.TurnName = "repair-preview"
		repairReq.TurnRole = req.DelegationPlan.IntegratorRole
		planned = append(planned, repairReq)
	}
	return planned
}

func resolveCodexInvocation(req executionRequest, capabilities codexCapabilityMatrix) (resolvedCodexInvocation, error) {
	approval := normalizeApprovalPolicy(req.ApprovalPolicy)
	if !isAllowedApprovalPolicy(approval) {
		return resolvedCodexInvocation{}, fmt.Errorf("approval_policy %q is not supported", req.ApprovalPolicy)
	}

	sandbox := normalizeSandboxMode(req.SandboxMode)
	if !isAllowedSandboxMode(sandbox) {
		return resolvedCodexInvocation{}, fmt.Errorf("sandbox_mode %q is not supported", req.SandboxMode)
	}

	sessionMode := normalizeSessionMode(req.SessionMode)
	if sessionMode == "" {
		sessionMode = "stateful"
	}
	if req.ResumeSession && !codexSessionStateful(sessionMode) {
		return resolvedCodexInvocation{}, fmt.Errorf("session_mode %q does not support resume", sessionMode)
	}

	invocation := resolvedCodexInvocation{
		CommandShape: "codex exec",
		Args:         []string{"exec"},
	}
	if !req.ResumeSession {
		return resolveSingleExecInvocation(invocation, req, capabilities.Exec, sandbox, sessionMode)
	}
	return resolveResumeInvocation(invocation, req, capabilities, sandbox)
}

func resolveSingleExecInvocation(invocation resolvedCodexInvocation, req executionRequest, surface codexCommandCapabilities, sandbox, sessionMode string) (resolvedCodexInvocation, error) {
	var err error
	invocation.Args, err = appendDirectFlagOrConfigOverride(invocation.Args, surface, surface.ApprovalFlag, "approval_policy", "-a", normalizeApprovalPolicy(req.ApprovalPolicy))
	if err != nil {
		return resolvedCodexInvocation{}, fmt.Errorf("%s: %w", invocation.CommandShape, err)
	}
	if surface.ApprovalFlag {
		invocation.DirectFlags = append(invocation.DirectFlags, "approval_policy")
	} else {
		invocation.ConfigOverrides = append(invocation.ConfigOverrides, "approval_policy")
	}

	invocation.Args, err = appendDirectFlagOrConfigOverride(invocation.Args, surface, surface.SandboxFlag, "sandbox_mode", "-s", sandbox)
	if err != nil {
		return resolvedCodexInvocation{}, fmt.Errorf("%s: %w", invocation.CommandShape, err)
	}
	if surface.SandboxFlag {
		invocation.DirectFlags = append(invocation.DirectFlags, "sandbox_mode")
	} else {
		invocation.ConfigOverrides = append(invocation.ConfigOverrides, "sandbox_mode")
	}

	if model := strings.TrimSpace(req.Model); model != "" {
		invocation.Args, err = appendDirectFlagOrConfigOverride(invocation.Args, surface, surface.ModelFlag, "model", "-m", model)
		if err != nil {
			return resolvedCodexInvocation{}, fmt.Errorf("%s: %w", invocation.CommandShape, err)
		}
		if surface.ModelFlag {
			invocation.DirectFlags = append(invocation.DirectFlags, "model")
		} else {
			invocation.ConfigOverrides = append(invocation.ConfigOverrides, "model")
		}
	}

	if profile := strings.TrimSpace(req.Profile); profile != "" {
		if !surface.ProfileFlag {
			return resolvedCodexInvocation{}, fmt.Errorf("%s: profile overrides require direct --profile support", invocation.CommandShape)
		}
		invocation.Args = append(invocation.Args, "-p", profile)
		invocation.DirectFlags = append(invocation.DirectFlags, "profile")
	}

	if req.WebSearch {
		if surface.WebSearch {
			invocation.Args = append(invocation.Args, "--search")
			invocation.DirectFlags = append(invocation.DirectFlags, "web_search")
		} else if surface.Config {
			invocation.Args = appendConfigOverride(invocation.Args, "web_search", tomlString("live"))
			invocation.ConfigOverrides = append(invocation.ConfigOverrides, "web_search")
		} else {
			return resolvedCodexInvocation{}, fmt.Errorf("%s: web_search cannot be represented by the installed Codex CLI", invocation.CommandShape)
		}
	}

	addDirs := nonEmptyArgs(req.AddDirs)
	if len(addDirs) > 0 {
		if surface.AddDirFlag {
			for _, dir := range addDirs {
				invocation.Args = append(invocation.Args, "--add-dir", dir)
			}
			invocation.DirectFlags = append(invocation.DirectFlags, "add_dirs")
		} else if surface.Config {
			invocation.Args = appendConfigOverride(invocation.Args, "sandbox_workspace_write.writable_roots", tomlStringArray(addDirs))
			invocation.ConfigOverrides = append(invocation.ConfigOverrides, "sandbox_workspace_write.writable_roots")
		} else {
			return resolvedCodexInvocation{}, fmt.Errorf("%s: add_dirs cannot be represented by the installed Codex CLI", invocation.CommandShape)
		}
	}

	if !req.ResumeSession && !codexSessionStateful(sessionMode) {
		invocation.Args = append(invocation.Args, "--ephemeral")
	}
	invocation.Args = append(invocation.Args, req.Prompt)
	return invocation, nil
}

func resolveResumeInvocation(invocation resolvedCodexInvocation, req executionRequest, capabilities codexCapabilityMatrix, sandbox string) (resolvedCodexInvocation, error) {
	invocation.CommandShape = "codex exec resume"
	prefix := append([]string{}, invocation.Args...)
	suffix := []string{"resume", "--last"}
	approval := normalizeApprovalPolicy(req.ApprovalPolicy)

	var err error
	prefix, suffix, err = appendResumeDirectOrConfig(prefix, suffix, capabilities, "approval_policy", "-a", approval)
	if err != nil {
		return resolvedCodexInvocation{}, fmt.Errorf("%s: %w", invocation.CommandShape, err)
	}
	if capabilities.Exec.ApprovalFlag || capabilities.Resume.ApprovalFlag {
		invocation.DirectFlags = append(invocation.DirectFlags, "approval_policy")
	} else {
		invocation.ConfigOverrides = append(invocation.ConfigOverrides, "approval_policy")
	}

	prefix, suffix, err = appendResumeDirectOrConfig(prefix, suffix, capabilities, "sandbox_mode", "-s", sandbox)
	if err != nil {
		return resolvedCodexInvocation{}, fmt.Errorf("%s: %w", invocation.CommandShape, err)
	}
	if capabilities.Exec.SandboxFlag || capabilities.Resume.SandboxFlag {
		invocation.DirectFlags = append(invocation.DirectFlags, "sandbox_mode")
	} else {
		invocation.ConfigOverrides = append(invocation.ConfigOverrides, "sandbox_mode")
	}

	if model := strings.TrimSpace(req.Model); model != "" {
		prefix, suffix, err = appendResumeDirectOrConfig(prefix, suffix, capabilities, "model", "-m", model)
		if err != nil {
			return resolvedCodexInvocation{}, fmt.Errorf("%s: %w", invocation.CommandShape, err)
		}
		if capabilities.Exec.ModelFlag || capabilities.Resume.ModelFlag {
			invocation.DirectFlags = append(invocation.DirectFlags, "model")
		} else {
			invocation.ConfigOverrides = append(invocation.ConfigOverrides, "model")
		}
	}

	if profile := strings.TrimSpace(req.Profile); profile != "" {
		if capabilities.Exec.ProfileFlag {
			prefix = append(prefix, "-p", profile)
		} else if capabilities.Resume.ProfileFlag {
			suffix = append(suffix, "-p", profile)
		} else {
			return resolvedCodexInvocation{}, fmt.Errorf("%s: profile overrides require direct --profile support", invocation.CommandShape)
		}
		invocation.DirectFlags = append(invocation.DirectFlags, "profile")
	}

	if req.WebSearch {
		switch {
		case capabilities.Exec.WebSearch:
			prefix = append(prefix, "--search")
			invocation.DirectFlags = append(invocation.DirectFlags, "web_search")
		case capabilities.Resume.WebSearch:
			suffix = append(suffix, "--search")
			invocation.DirectFlags = append(invocation.DirectFlags, "web_search")
		case capabilities.Resume.Config:
			suffix = appendConfigOverride(suffix, "web_search", tomlString("live"))
			invocation.ConfigOverrides = append(invocation.ConfigOverrides, "web_search")
		case capabilities.Exec.Config:
			prefix = appendConfigOverride(prefix, "web_search", tomlString("live"))
			invocation.ConfigOverrides = append(invocation.ConfigOverrides, "web_search")
		default:
			return resolvedCodexInvocation{}, fmt.Errorf("%s: web_search cannot be represented by the installed Codex CLI", invocation.CommandShape)
		}
	}

	addDirs := nonEmptyArgs(req.AddDirs)
	if len(addDirs) > 0 {
		switch {
		case capabilities.Exec.AddDirFlag:
			for _, dir := range addDirs {
				prefix = append(prefix, "--add-dir", dir)
			}
			invocation.DirectFlags = append(invocation.DirectFlags, "add_dirs")
		case capabilities.Resume.AddDirFlag:
			for _, dir := range addDirs {
				suffix = append(suffix, "--add-dir", dir)
			}
			invocation.DirectFlags = append(invocation.DirectFlags, "add_dirs")
		case capabilities.Resume.Config:
			suffix = appendConfigOverride(suffix, "sandbox_workspace_write.writable_roots", tomlStringArray(addDirs))
			invocation.ConfigOverrides = append(invocation.ConfigOverrides, "sandbox_workspace_write.writable_roots")
		case capabilities.Exec.Config:
			prefix = appendConfigOverride(prefix, "sandbox_workspace_write.writable_roots", tomlStringArray(addDirs))
			invocation.ConfigOverrides = append(invocation.ConfigOverrides, "sandbox_workspace_write.writable_roots")
		default:
			return resolvedCodexInvocation{}, fmt.Errorf("%s: add_dirs cannot be represented by the installed Codex CLI", invocation.CommandShape)
		}
	}

	invocation.Args = append(prefix, suffix...)
	invocation.Args = append(invocation.Args, req.Prompt)
	return invocation, nil
}

func appendResumeDirectOrConfig(prefix, suffix []string, capabilities codexCapabilityMatrix, key, flag, value string) ([]string, []string, error) {
	switch {
	case resumeControlExecSupport(capabilities.Exec, key):
		return append(prefix, flag, value), suffix, nil
	case resumeControlResumeSupport(capabilities.Resume, key):
		return prefix, append(suffix, flag, value), nil
	case capabilities.Resume.Config:
		return prefix, appendConfigOverride(suffix, key, tomlString(value)), nil
	case capabilities.Exec.Config:
		return appendConfigOverride(prefix, key, tomlString(value)), suffix, nil
	default:
		return nil, nil, fmt.Errorf("%s cannot be represented by the installed Codex CLI", key)
	}
}

func resumeControlExecSupport(surface codexCommandCapabilities, key string) bool {
	switch key {
	case "approval_policy":
		return surface.ApprovalFlag
	case "sandbox_mode":
		return surface.SandboxFlag
	case "model":
		return surface.ModelFlag
	default:
		return false
	}
}

func resumeControlResumeSupport(surface codexCommandCapabilities, key string) bool {
	switch key {
	case "approval_policy":
		return surface.ApprovalFlag
	case "sandbox_mode":
		return surface.SandboxFlag
	case "model":
		return surface.ModelFlag
	default:
		return false
	}
}

func appendDirectFlagOrConfigOverride(args []string, surface codexCommandCapabilities, directSupported bool, key, flag, value string) ([]string, error) {
	if directSupported {
		return append(args, flag, value), nil
	}
	if !surface.Config {
		return nil, fmt.Errorf("%s cannot be represented by the installed Codex CLI", key)
	}
	return appendConfigOverride(args, key, tomlString(value)), nil
}

func appendConfigOverride(args []string, key, value string) []string {
	return append(args, "-c", key+"="+value)
}

func tomlString(value string) string {
	return strconv.Quote(value)
}

func tomlStringArray(values []string) string {
	if len(values) == 0 {
		return "[]"
	}

	quoted := make([]string, 0, len(values))
	for _, value := range values {
		quoted = append(quoted, strconv.Quote(value))
	}
	return "[" + strings.Join(quoted, ", ") + "]"
}

func nonEmptyArgs(values []string) []string {
	filtered := make([]string, 0, len(values))
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			filtered = append(filtered, trimmed)
		}
	}
	return filtered
}

func formatCodexInvocationSummary(invocation resolvedCodexInvocation) string {
	parts := []string{invocation.CommandShape}
	if len(invocation.DirectFlags) > 0 {
		parts = append(parts, "direct["+strings.Join(invocation.DirectFlags, ",")+"]")
	}
	if len(invocation.ConfigOverrides) > 0 {
		parts = append(parts, "config["+strings.Join(invocation.ConfigOverrides, ",")+"]")
	}
	return strings.Join(parts, " ")
}
