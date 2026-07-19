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
	Version           string                   `json:"version,omitempty"`
	Exec              codexCommandCapabilities `json:"exec"`
	Resume            codexCommandCapabilities `json:"resume"`
	SolAvailable      *bool                    `json:"sol_available,omitempty"`
	ModelAvailability map[string]bool          `json:"model_availability,omitempty"`
	Probes            []lifecycleProbeOutcome  `json:"probes,omitempty"`
}

type resolvedCodexInvocation struct {
	CommandShape    string
	Args            []string
	DirectFlags     []string
	ConfigOverrides []string
}

const plannedCodexResumeThreadID = "00000000-0000-0000-0000-000000000000"

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

	matrix := codexCapabilityMatrix{}
	var version string
	versionProbe, err := runBoundedLifecycleProbe(ctx, lifecycleProbeCodexVersion, a.capabilityProbeTimeout, func(probeCtx context.Context) error {
		var probeErr error
		version, probeErr = a.runBinary(probeCtx, "codex", []string{"--version"}, dir)
		return probeErr
	})
	matrix.Probes = append(matrix.Probes, versionProbe)
	if err != nil {
		return matrix, fmt.Errorf("codex --version: %w", err)
	}
	matrix.Version = strings.TrimSpace(version)

	var execHelp string
	execHelpProbe, err := runBoundedLifecycleProbe(ctx, lifecycleProbeCodexExecHelp, a.capabilityProbeTimeout, func(probeCtx context.Context) error {
		var probeErr error
		execHelp, probeErr = a.runBinary(probeCtx, "codex", []string{"exec", "--help"}, dir)
		return probeErr
	})
	matrix.Probes = append(matrix.Probes, execHelpProbe)
	if err != nil {
		return matrix, fmt.Errorf("codex exec --help: %w", err)
	}
	matrix.Exec = parseCodexCommandCapabilities(execHelp)
	if req.ModelRoutingPolicy == modelRoutingPolicyCostBalancedV1 {
		matrix.ModelAvailability = make(map[string]bool)
		a.probePlannedRoutedModels(ctx, dir, req, &matrix)
	}
	planningReq := withModelAvailability(req, matrix)
	planned, immediateBlock := executablePlannedCodexRequests(plannedCodexRequests(planningReq))
	if immediateBlock || !plannedInvocationsNeedResume(planned) {
		return matrix, nil
	}

	var resumeHelp string
	resumeHelpProbe, err := runBoundedLifecycleProbe(ctx, lifecycleProbeCodexResumeExecHelp, a.capabilityProbeTimeout, func(probeCtx context.Context) error {
		var probeErr error
		resumeHelp, probeErr = a.runBinary(probeCtx, "codex", []string{"exec", "resume", "--help"}, dir)
		return probeErr
	})
	matrix.Probes = append(matrix.Probes, resumeHelpProbe)
	if err != nil {
		return matrix, fmt.Errorf("codex exec resume --help: %w", err)
	}
	matrix.Resume = parseCodexCommandCapabilities(resumeHelp)
	return matrix, nil
}

func (a *App) probePlannedRoutedModels(ctx context.Context, dir string, req executionRequest, matrix *codexCapabilityMatrix) {
	if matrix == nil {
		return
	}
	if matrix.ModelAvailability == nil {
		matrix.ModelAvailability = make(map[string]bool)
	}
	for {
		planningReq := withModelAvailability(req, *matrix)
		models := plannedRoutedModelIDs(plannedCodexRequests(planningReq))
		nextModel := ""
		for _, model := range models {
			if _, probed := matrix.ModelAvailability[model]; !probed {
				nextModel = model
				break
			}
		}
		if nextModel == "" {
			return
		}
		available, probe := a.probeModelAvailability(ctx, dir, req, nextModel, *matrix)
		matrix.ModelAvailability[nextModel] = available
		if nextModel == modelRoutingModelSol {
			solAvailable := available
			matrix.SolAvailable = &solAvailable
		}
		matrix.Probes = append(matrix.Probes, probe)
	}
}

func plannedRoutedModelIDs(planned []executionRequest) []string {
	models := make([]string, 0, len(planned))
	seen := make(map[string]bool, len(planned))
	for _, req := range planned {
		if req.ModelRoutingPolicy != modelRoutingPolicyCostBalancedV1 {
			continue
		}
		if req.RoutingDecision.Status == modelRoutingStatusBlocked {
			if req.TurnName != "repair-preview" {
				break
			}
			continue
		}
		model := strings.TrimSpace(req.RoutingDecision.Model)
		if model == "" || seen[model] {
			continue
		}
		seen[model] = true
		models = append(models, model)
	}
	return models
}

func withModelAvailability(req executionRequest, capabilities codexCapabilityMatrix) executionRequest {
	req.SolAvailable = capabilities.SolAvailable
	if len(capabilities.ModelAvailability) == 0 {
		req.ModelAvailability = nil
		return req
	}
	req.ModelAvailability = make(map[string]bool, len(capabilities.ModelAvailability))
	for model, available := range capabilities.ModelAvailability {
		req.ModelAvailability[model] = available
	}
	return req
}

func (a *App) probeModelAvailability(ctx context.Context, dir string, req executionRequest, model string, capabilities codexCapabilityMatrix) (bool, lifecycleProbeOutcome) {
	probeName := lifecycleProbeCodexModelAvailability
	if model == modelRoutingModelSol {
		probeName = lifecycleProbeSolAvailability
	}
	baseOutcome := lifecycleProbeOutcome{Name: probeName, Model: model}
	if (!capabilities.Exec.ModelFlag && !capabilities.Exec.Config) || !capabilities.Exec.JSONFlag || a.runCodexCmdWithInput == nil {
		baseOutcome.Status = lifecycleProbeStatusUnsupported
		return false, baseOutcome
	}

	sessionMode := "stateful"
	if capabilities.Exec.EphemeralFlag {
		sessionMode = "ephemeral"
	}
	command, err := buildCodexExecCommand(executionRequest{
		WorkDir:        dir,
		Prompt:         "Return exactly `namba-model-available`. Do not inspect or modify files.",
		ApprovalPolicy: "never",
		SandboxMode:    "read-only",
		Model:          model,
		Profile:        strings.TrimSpace(req.Profile),
		SessionMode:    sessionMode,
	}, capabilities)
	if err != nil {
		baseOutcome.Status = lifecycleProbeStatusError
		return false, baseOutcome
	}

	var stdout, stderr string
	probe, probeErr := runBoundedLifecycleProbe(ctx, probeName, a.capabilityProbeTimeout, func(probeCtx context.Context) error {
		var runErr error
		stdout, stderr, runErr = a.runCodexCmdWithInput(probeCtx, "codex", command.Args, dir, command.Input)
		return runErr
	})
	probe.Model = model
	if probeErr != nil {
		return false, probe
	}
	available := firstCodexThreadID(strings.Join(nonEmptyArgs([]string{stdout, stderr}), "\n")) != ""
	if available {
		probe.Status = lifecycleProbeStatusAvailable
	} else {
		probe.Status = lifecycleProbeStatusUnavailable
	}
	return available, probe
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
	planned, immediateBlock := executablePlannedCodexRequests(plannedCodexRequests(req))
	if immediateBlock {
		for _, plannedReq := range plannedCodexRequests(req) {
			if plannedReq.TurnName == "repair-preview" || plannedReq.RoutingDecision.Status != modelRoutingStatusBlocked {
				continue
			}
			return nil, validateModelRoutingTurnPlan([]executionRequest{plannedReq})
		}
		return nil, fmt.Errorf("model routing blocked before invocation planning")
	}
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

// executablePlannedCodexRequests separates turns that can run now from a
// conditional repair preview. An immediate blocked turn takes precedence over
// invocation validation, while a blocked future repair must not suppress
// validation of otherwise executable turns.
func executablePlannedCodexRequests(planned []executionRequest) ([]executionRequest, bool) {
	executable := make([]executionRequest, 0, len(planned))
	for _, req := range planned {
		blocked := req.ModelRoutingPolicy == modelRoutingPolicyCostBalancedV1 && req.RoutingDecision.Status == modelRoutingStatusBlocked
		if !blocked {
			executable = append(executable, req)
			continue
		}
		if req.TurnName != "repair-preview" {
			return nil, true
		}
	}
	return executable, false
}

func plannedCodexRequests(req executionRequest) []executionRequest {
	planned := plannedExecutionTurnRequests(req)
	resumeThreadID := firstNonBlank(strings.TrimSpace(req.ThreadID), plannedCodexResumeThreadID)
	if req.RepairAttempts > 0 {
		solRemaining, solHighRemaining := remainingSolTurnBudgets(req, planned)
		repairSessionID := ""
		if codexSessionStateful(req.SessionMode) {
			repairSessionID = resumeThreadID
		}
		repairReq, _, _ := buildRepairExecutionTurnRequest(req, validationReport{}, 1, repairSessionID, "", solRemaining, solHighRemaining)
		repairReq.TurnName = "repair-preview"
		planned = append(planned, repairReq)
	}
	return planned
}

func plannedExecutionTurnRequests(req executionRequest) []executionRequest {
	planned := buildExecutionTurnRequests(req)
	resumeThreadID := firstNonBlank(strings.TrimSpace(req.ThreadID), plannedCodexResumeThreadID)
	if codexSessionStateful(req.SessionMode) {
		for index := 1; index < len(planned); index++ {
			if !canResumeExecutionTurn(planned[index-1], planned[index]) {
				continue
			}
			planned[index].ResumeSession = true
			planned[index].ThreadID = resumeThreadID
		}
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
	if req.ResumeSession && strings.TrimSpace(req.ThreadID) == "" {
		return resolvedCodexInvocation{}, fmt.Errorf("resume requires an explicit Codex thread UUID")
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
	if effort := strings.TrimSpace(req.RequestedReasoningEffort); effort != "" {
		if !surface.Config {
			return resolvedCodexInvocation{}, fmt.Errorf("%s: model_reasoning_effort cannot be represented by the installed Codex CLI", invocation.CommandShape)
		}
		invocation.Args = appendConfigOverride(invocation.Args, "model_reasoning_effort", tomlString(effort))
		invocation.ConfigOverrides = append(invocation.ConfigOverrides, "model_reasoning_effort")
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
	if surface.JSONFlag {
		invocation.Args = append(invocation.Args, "--json")
		invocation.DirectFlags = append(invocation.DirectFlags, "json_output")
	}
	invocation.Args = append(invocation.Args, "-")
	return invocation, nil
}

func resolveResumeInvocation(invocation resolvedCodexInvocation, req executionRequest, capabilities codexCapabilityMatrix, sandbox string) (resolvedCodexInvocation, error) {
	invocation.CommandShape = "codex exec resume"
	prefix := append([]string{}, invocation.Args...)
	suffix := []string{"resume", strings.TrimSpace(req.ThreadID)}
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
	if effort := strings.TrimSpace(req.RequestedReasoningEffort); effort != "" {
		switch {
		case capabilities.Resume.Config:
			suffix = appendConfigOverride(suffix, "model_reasoning_effort", tomlString(effort))
		case capabilities.Exec.Config:
			prefix = appendConfigOverride(prefix, "model_reasoning_effort", tomlString(effort))
		default:
			return resolvedCodexInvocation{}, fmt.Errorf("%s: model_reasoning_effort cannot be represented by the installed Codex CLI", invocation.CommandShape)
		}
		invocation.ConfigOverrides = append(invocation.ConfigOverrides, "model_reasoning_effort")
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

	switch {
	case capabilities.Exec.JSONFlag:
		prefix = append(prefix, "--json")
		invocation.DirectFlags = append(invocation.DirectFlags, "json_output")
	case capabilities.Resume.JSONFlag:
		suffix = append(suffix, "--json")
		invocation.DirectFlags = append(invocation.DirectFlags, "json_output")
	}

	invocation.Args = append(prefix, suffix...)
	invocation.Args = append(invocation.Args, "-")
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
