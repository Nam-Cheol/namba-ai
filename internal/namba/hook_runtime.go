package namba

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	hookContextSchemaVersion = "namba-hook/v1"
	hookScopeWorker          = "worker"
	hookConfigErrorName      = "hook_config_error"

	hookStatusSucceeded = "succeeded"
	hookStatusFailed    = "failed"
	hookStatusTimeout   = "timeout"
	hookStatusError     = "error"

	hookFailureActionContinued     = "continued"
	hookFailureActionStopped       = "stopped"
	hookFailureActionNotApplicable = "not_applicable"
)

type hookEvent string

const (
	hookEventBeforePreflight  hookEvent = "before_preflight"
	hookEventAfterPreflight   hookEvent = "after_preflight"
	hookEventBeforeExecution  hookEvent = "before_execution"
	hookEventAfterExecution   hookEvent = "after_execution"
	hookEventAfterPatch       hookEvent = "after_patch"
	hookEventAfterBash        hookEvent = "after_bash"
	hookEventAfterMCPTool     hookEvent = "after_mcp_tool"
	hookEventBeforeValidation hookEvent = "before_validation"
	hookEventAfterValidation  hookEvent = "after_validation"
	hookEventOnFailure        hookEvent = "on_failure"
)

type hookResult struct {
	Event         string `json:"event"`
	HookName      string `json:"hook_name"`
	Command       string `json:"command"`
	CWD           string `json:"cwd"`
	StartedAt     string `json:"started_at"`
	EndedAt       string `json:"ended_at"`
	DurationMS    int64  `json:"duration_ms"`
	ExitCode      int    `json:"exit_code"`
	Status        string `json:"status"`
	StdoutPath    string `json:"stdout_path"`
	StderrPath    string `json:"stderr_path"`
	Blocking      bool   `json:"blocking"`
	FailureAction string `json:"failure_action"`
	ErrorSummary  string `json:"error_summary"`
	Scope         string `json:"scope"`
	Attempt       int    `json:"attempt,omitempty"`
	ToolName      string `json:"tool_name,omitempty"`
	ToolUseID     string `json:"tool_use_id,omitempty"`
}

type hookExecutionContext struct {
	SchemaVersion string            `json:"schema_version"`
	Event         string            `json:"event"`
	LogID         string            `json:"log_id"`
	RunID         string            `json:"run_id"`
	SpecID        string            `json:"spec_id"`
	ExecutionMode string            `json:"execution_mode"`
	WorkDir       string            `json:"work_dir"`
	ProjectRoot   string            `json:"project_root"`
	Artifacts     map[string]string `json:"artifacts"`
	StageStatus   string            `json:"stage_status,omitempty"`
	TriggeredAt   string            `json:"triggered_at"`
	FailurePhase  string            `json:"failure_phase,omitempty"`
	FailureStatus string            `json:"failure_status,omitempty"`
	ErrorSummary  string            `json:"error_summary,omitempty"`
	BlockingHook  *hookResult       `json:"blocking_hook,omitempty"`
	Attempt       int               `json:"attempt,omitempty"`
	ToolName      string            `json:"tool_name,omitempty"`
	ToolUseID     string            `json:"tool_use_id,omitempty"`
	ToolStatus    string            `json:"tool_status,omitempty"`
	ExitCode      int               `json:"exit_code,omitempty"`
	Command       string            `json:"command,omitempty"`
	CWD           string            `json:"cwd,omitempty"`
	EventData     map[string]any    `json:"event_data,omitempty"`
}

type hookConfig struct {
	Hooks []hookRegistration
}

type hookRegistration struct {
	Name              string
	Event             hookEvent
	Command           string
	CWD               string
	ResolvedCWD       string
	TimeoutSeconds    int
	Enabled           bool
	ContinueOnFailure bool
}

type hookLifecycle struct {
	app          *App
	artifactRoot string
	configRoot   string
	logID        string
	req          executionRequest
	progressPath string
	scope        string
	config       hookConfig
	configErr    error
	results      []hookResult
	outputCounts map[string]int
	failureSent  bool
	blockingHook *hookResult
}

type hookTrigger struct {
	Event         hookEvent
	StageStatus   string
	FailurePhase  string
	FailureStatus string
	ErrorSummary  string
	Attempt       int
	BlockingHook  *hookResult
	ToolName      string
	ToolUseID     string
	ToolStatus    string
	ExitCode      int
	Command       string
	CWD           string
	EventData     map[string]any
}

type hookFailureError struct {
	Result hookResult
}

func (e hookFailureError) Error() string {
	if strings.TrimSpace(e.Result.ErrorSummary) != "" {
		return fmt.Sprintf("blocking hook %s failed for %s: %s", e.Result.HookName, e.Result.Event, e.Result.ErrorSummary)
	}
	return fmt.Sprintf("blocking hook %s failed for %s", e.Result.HookName, e.Result.Event)
}

type runnerObservationType string

const (
	runnerObservationPatchApplied     runnerObservationType = "patch_applied"
	runnerObservationBashCompleted    runnerObservationType = "bash_completed"
	runnerObservationMCPToolCompleted runnerObservationType = "mcp_tool_completed"
)

type runnerObservation struct {
	ObservationType runnerObservationType
	ToolName        string
	ToolUseID       string
	StartedAt       string
	EndedAt         string
	Status          string
	ExitCode        int
	Command         string
	CWD             string
	InputSummary    string
	OutputArtifacts []string
}

func newHookLifecycle(app *App, artifactRoot, logID string, req executionRequest, progressPath string) *hookLifecycle {
	artifactRoot = strings.TrimSpace(artifactRoot)
	if abs, err := filepath.Abs(artifactRoot); err == nil {
		artifactRoot = filepath.Clean(abs)
	}
	configRoot := strings.TrimSpace(req.WorkDir)
	if configRoot == "" {
		configRoot = artifactRoot
	}
	if abs, err := filepath.Abs(configRoot); err == nil {
		configRoot = filepath.Clean(abs)
	}
	cfg, err := loadHookConfig(configRoot)
	return &hookLifecycle{
		app:          app,
		artifactRoot: artifactRoot,
		configRoot:   configRoot,
		logID:        strings.TrimSpace(logID),
		req:          req,
		progressPath: strings.TrimSpace(progressPath),
		scope:        hookScopeWorker,
		config:       cfg,
		configErr:    err,
		outputCounts: make(map[string]int),
	}
}

func (l *hookLifecycle) handleConfigError(ctx context.Context) error {
	if l == nil || l.configErr == nil {
		return nil
	}
	result := l.configErrorResult(l.configErr)
	l.results = append(l.results, result)
	l.blockingHook = &l.results[len(l.results)-1]
	l.configErr = nil
	_ = l.triggerOnFailure(ctx, hookTrigger{
		FailurePhase:  "hooks",
		FailureStatus: "hook_failed",
		ErrorSummary:  result.ErrorSummary,
		BlockingHook:  &result,
	})
	return hookFailureError{Result: result}
}

func (l *hookLifecycle) Trigger(ctx context.Context, trigger hookTrigger) error {
	if l == nil {
		return nil
	}
	if err := l.handleConfigError(ctx); err != nil {
		return err
	}
	if trigger.Event == hookEventOnFailure {
		return l.triggerOnFailure(ctx, trigger)
	}
	for _, hook := range l.config.enabledHooksForEvent(trigger.Event) {
		result := l.runHook(ctx, hook, trigger)
		l.results = append(l.results, result)
		if result.Blocking {
			l.blockingHook = &l.results[len(l.results)-1]
			return hookFailureError{Result: result}
		}
	}
	return nil
}

func (l *hookLifecycle) ObserveRunnerTool(ctx context.Context, observation runnerObservation) error {
	event := hookEvent("")
	switch observation.ObservationType {
	case runnerObservationPatchApplied:
		event = hookEventAfterPatch
	case runnerObservationBashCompleted:
		event = hookEventAfterBash
	case runnerObservationMCPToolCompleted:
		event = hookEventAfterMCPTool
	default:
		return nil
	}
	data := map[string]any{
		"observation_type": string(observation.ObservationType),
		"input_summary":    strings.TrimSpace(observation.InputSummary),
		"output_artifacts": append([]string(nil), observation.OutputArtifacts...),
		"started_at":       strings.TrimSpace(observation.StartedAt),
		"ended_at":         strings.TrimSpace(observation.EndedAt),
	}
	return l.Trigger(ctx, hookTrigger{
		Event:       event,
		StageStatus: strings.TrimSpace(observation.Status),
		ToolName:    strings.TrimSpace(observation.ToolName),
		ToolUseID:   strings.TrimSpace(observation.ToolUseID),
		ToolStatus:  strings.TrimSpace(observation.Status),
		ExitCode:    observation.ExitCode,
		Command:     strings.TrimSpace(observation.Command),
		CWD:         strings.TrimSpace(observation.CWD),
		EventData:   data,
	})
}

func (l *hookLifecycle) writeRunEvidence(ctx context.Context, status string, validationAttempts int, progressLogFailed bool, failureSummary string) error {
	if l == nil {
		return nil
	}
	if isHookFailureStatus(status) {
		if err := l.triggerOnFailure(ctx, hookTrigger{
			FailurePhase:  executionEvidenceFailurePhase(status),
			FailureStatus: status,
			ErrorSummary:  firstNonBlank(failureSummary, status),
			BlockingHook:  l.blockingHook,
		}); err != nil {
			return err
		}
	}

	progress := executionEvidenceRefInput{
		Kind:          "progress",
		NotApplicable: true,
	}
	if relPath := firstNonBlank(
		executionEvidenceRelativePath(l.artifactRoot, l.progressPath),
		relativeParallelProgressLogPath(l.req.SpecID),
	); normalizeExecutionMode(l.req.Mode) == executionModeParallel && relPath != "" {
		progress = executionEvidenceRefInput{
			Kind: "progress",
			Path: relPath,
		}
	}

	return l.app.writeExecutionEvidenceManifest(l.artifactRoot, executionEvidenceOptions{
		ProjectRoot:        l.artifactRoot,
		LogID:              l.logID,
		SpecID:             l.req.SpecID,
		ExecutionMode:      l.req.Mode,
		Status:             status,
		ValidationAttempts: validationAttempts,
		ProgressLogFailed:  progressLogFailed,
		GeneratedAt:        l.app.now(),
		FinalizedBy:        "executeRun",
		Progress:           progress,
		Hooks:              l.results,
	})
}

func (l *hookLifecycle) triggerOnFailure(ctx context.Context, trigger hookTrigger) error {
	if l == nil || l.failureSent {
		return nil
	}
	l.failureSent = true
	trigger.Event = hookEventOnFailure
	if trigger.FailureStatus == "" {
		trigger.FailureStatus = "hook_failed"
	}
	if trigger.FailurePhase == "" {
		trigger.FailurePhase = executionEvidenceFailurePhase(trigger.FailureStatus)
	}
	for _, hook := range l.config.enabledHooksForEvent(hookEventOnFailure) {
		result := l.runHook(ctx, hook, trigger)
		l.results = append(l.results, result)
	}
	return nil
}

func (l *hookLifecycle) runHook(ctx context.Context, hook hookRegistration, trigger hookTrigger) hookResult {
	start := l.app.now()
	result := hookResult{
		Event:         string(trigger.Event),
		HookName:      hook.Name,
		Command:       hook.Command,
		CWD:           hook.ResolvedCWD,
		StartedAt:     start.Format(time.RFC3339),
		ExitCode:      -1,
		Status:        hookStatusError,
		FailureAction: hookFailureActionNotApplicable,
		Scope:         l.scope,
		Attempt:       trigger.Attempt,
		ToolName:      trigger.ToolName,
		ToolUseID:     trigger.ToolUseID,
	}
	stdoutPath, stderrPath := l.nextOutputPaths(trigger.Event, hook.Name)
	result.StdoutPath = stdoutPath
	result.StderrPath = stderrPath

	contextJSON, err := json.Marshal(l.contextForTrigger(trigger))
	if err != nil {
		result.ErrorSummary = err.Error()
		l.finishHookResult(&result, start)
		if writeErr := l.writeHookOutputs(result, "", result.ErrorSummary); writeErr != nil {
			recordHookArtifactWriteFailure(&result, writeErr)
		}
		return l.applyHookFailurePolicy(result, hook)
	}

	if info, statErr := os.Stat(hook.ResolvedCWD); statErr != nil {
		result.ErrorSummary = fmt.Sprintf("resolve cwd: %v", statErr)
		l.finishHookResult(&result, start)
		if writeErr := l.writeHookOutputs(result, "", result.ErrorSummary); writeErr != nil {
			recordHookArtifactWriteFailure(&result, writeErr)
		}
		return l.applyHookFailurePolicy(result, hook)
	} else if !info.IsDir() {
		result.ErrorSummary = fmt.Sprintf("cwd is not a directory: %s", hook.ResolvedCWD)
		l.finishHookResult(&result, start)
		if writeErr := l.writeHookOutputs(result, "", result.ErrorSummary); writeErr != nil {
			recordHookArtifactWriteFailure(&result, writeErr)
		}
		return l.applyHookFailurePolicy(result, hook)
	}

	hookCtx, cancel := context.WithTimeout(ctx, time.Duration(hook.TimeoutSeconds)*time.Second)
	defer cancel()
	stdout, stderr, runErr := runShellCommandWithInput(hookCtx, l.app.hookCommandRunner(), hook.Command, hook.ResolvedCWD, string(contextJSON))
	end := l.app.now()
	result.EndedAt = end.Format(time.RFC3339)
	result.DurationMS = maxInt64(0, end.Sub(start).Milliseconds())
	result.ExitCode = hookExitCode(runErr)
	if runErr == nil {
		result.Status = hookStatusSucceeded
		result.ExitCode = 0
	} else if errors.Is(hookCtx.Err(), context.DeadlineExceeded) || errors.Is(runErr, context.DeadlineExceeded) {
		result.Status = hookStatusTimeout
		result.ExitCode = -1
		result.ErrorSummary = "hook timed out"
	} else if exitCode := hookExitCode(runErr); exitCode >= 0 {
		result.Status = hookStatusFailed
		result.ExitCode = exitCode
		result.ErrorSummary = "hook exited non-zero"
	} else {
		result.Status = hookStatusError
		result.ExitCode = -1
		result.ErrorSummary = runErr.Error()
	}
	if writeErr := l.writeHookOutputs(result, stdout, stderr); writeErr != nil {
		recordHookArtifactWriteFailure(&result, writeErr)
	}
	return l.applyHookFailurePolicy(result, hook)
}

func (l *hookLifecycle) configErrorResult(err error) hookResult {
	start := l.app.now()
	stdoutPath, stderrPath := l.nextOutputPaths(hookEventBeforePreflight, hookConfigErrorName)
	result := hookResult{
		Event:         string(hookEventBeforePreflight),
		HookName:      hookConfigErrorName,
		Command:       "",
		CWD:           l.configRoot,
		StartedAt:     start.Format(time.RFC3339),
		EndedAt:       start.Format(time.RFC3339),
		DurationMS:    0,
		ExitCode:      -1,
		Status:        hookStatusError,
		StdoutPath:    stdoutPath,
		StderrPath:    stderrPath,
		Blocking:      true,
		FailureAction: hookFailureActionStopped,
		ErrorSummary:  strings.TrimSpace(err.Error()),
		Scope:         l.scope,
	}
	if writeErr := l.writeHookOutputs(result, "", result.ErrorSummary); writeErr != nil {
		recordHookArtifactWriteFailure(&result, writeErr)
	}
	return result
}

func (l *hookLifecycle) finishHookResult(result *hookResult, start time.Time) {
	end := l.app.now()
	result.EndedAt = end.Format(time.RFC3339)
	result.DurationMS = maxInt64(0, end.Sub(start).Milliseconds())
	if result.FailureAction == "" {
		result.FailureAction = hookFailureActionNotApplicable
	}
	if result.Scope == "" {
		result.Scope = l.scope
	}
}

func (l *hookLifecycle) applyHookFailurePolicy(result hookResult, hook hookRegistration) hookResult {
	if result.Status == hookStatusSucceeded {
		result.Blocking = false
		result.FailureAction = hookFailureActionNotApplicable
		result.ErrorSummary = ""
		return result
	}
	if hook.ContinueOnFailure {
		result.Blocking = false
		result.FailureAction = hookFailureActionContinued
		return result
	}
	result.Blocking = true
	result.FailureAction = hookFailureActionStopped
	if result.ErrorSummary == "" {
		result.ErrorSummary = "blocking hook failed"
	}
	return result
}

func (l *hookLifecycle) contextForTrigger(trigger hookTrigger) hookExecutionContext {
	return hookExecutionContext{
		SchemaVersion: hookContextSchemaVersion,
		Event:         string(trigger.Event),
		LogID:         l.logID,
		RunID:         l.logID,
		SpecID:        strings.TrimSpace(l.req.SpecID),
		ExecutionMode: string(normalizeExecutionMode(l.req.Mode)),
		WorkDir:       strings.TrimSpace(l.req.WorkDir),
		ProjectRoot:   l.configRoot,
		Artifacts:     l.artifactPaths(),
		StageStatus:   strings.TrimSpace(trigger.StageStatus),
		TriggeredAt:   l.app.now().Format(time.RFC3339),
		FailurePhase:  strings.TrimSpace(trigger.FailurePhase),
		FailureStatus: strings.TrimSpace(trigger.FailureStatus),
		ErrorSummary:  strings.TrimSpace(trigger.ErrorSummary),
		BlockingHook:  trigger.BlockingHook,
		Attempt:       trigger.Attempt,
		ToolName:      strings.TrimSpace(trigger.ToolName),
		ToolUseID:     strings.TrimSpace(trigger.ToolUseID),
		ToolStatus:    strings.TrimSpace(trigger.ToolStatus),
		ExitCode:      trigger.ExitCode,
		Command:       strings.TrimSpace(trigger.Command),
		CWD:           strings.TrimSpace(trigger.CWD),
		EventData:     trigger.EventData,
	}
}

func (l *hookLifecycle) artifactPaths() map[string]string {
	paths := map[string]string{
		"request":    l.artifactPath(filepath.ToSlash(filepath.Join(logsDir, "runs", l.logID+"-request.json"))),
		"preflight":  l.artifactPath(filepath.ToSlash(filepath.Join(logsDir, "runs", l.logID+"-preflight.json"))),
		"execution":  l.artifactPath(filepath.ToSlash(filepath.Join(logsDir, "runs", l.logID+"-execution.json"))),
		"validation": l.artifactPath(filepath.ToSlash(filepath.Join(logsDir, "runs", l.logID+"-validation.json"))),
		"evidence":   l.artifactPath(executionEvidenceManifestPath(l.logID)),
	}
	if strings.TrimSpace(l.progressPath) != "" {
		if filepath.IsAbs(l.progressPath) {
			paths["progress"] = filepath.ToSlash(filepath.Clean(l.progressPath))
		} else {
			paths["progress"] = l.artifactPath(l.progressPath)
		}
	}
	return paths
}

func (l *hookLifecycle) artifactPath(rel string) string {
	rel = strings.TrimSpace(rel)
	if rel == "" {
		return ""
	}
	if filepath.IsAbs(rel) {
		return filepath.ToSlash(filepath.Clean(rel))
	}
	if strings.TrimSpace(l.artifactRoot) == "" {
		return filepath.ToSlash(filepath.Clean(filepath.FromSlash(rel)))
	}
	return filepath.ToSlash(filepath.Join(l.artifactRoot, filepath.FromSlash(rel)))
}

func (l *hookLifecycle) nextOutputPaths(event hookEvent, hookName string) (string, string) {
	key := string(event) + "\x00" + hookName
	l.outputCounts[key]++
	suffix := ""
	if l.outputCounts[key] > 1 {
		suffix = fmt.Sprintf("-%d", l.outputCounts[key])
	}
	base := filepath.ToSlash(filepath.Join(logsDir, "runs", l.logID+"-hooks", string(event), hookArtifactName(hookName)+suffix))
	return base + "-stdout.txt", base + "-stderr.txt"
}

func (l *hookLifecycle) writeHookOutputs(result hookResult, stdout, stderr string) error {
	if err := writeRunText(filepath.Join(l.artifactRoot, filepath.FromSlash(result.StdoutPath)), stdout); err != nil {
		return err
	}
	if err := writeRunText(filepath.Join(l.artifactRoot, filepath.FromSlash(result.StderrPath)), stderr); err != nil {
		return err
	}
	return nil
}

func recordHookArtifactWriteFailure(result *hookResult, err error) {
	if result == nil || err == nil {
		return
	}
	result.Status = hookStatusError
	writeSummary := fmt.Sprintf("write hook artifacts: %v", err)
	if summary := strings.TrimSpace(result.ErrorSummary); summary != "" {
		result.ErrorSummary = summary + "; " + writeSummary
		return
	}
	result.ErrorSummary = writeSummary
}

func (a *App) hookCommandRunner() func(context.Context, string, []string, string, string) (string, string, error) {
	if a.runCmdWithInput != nil {
		return a.runCmdWithInput
	}
	return func(ctx context.Context, name string, args []string, dir, input string) (string, string, error) {
		output, err := a.runCmd(ctx, name, args, dir)
		return output, "", err
	}
}

func loadHookConfig(root string) (hookConfig, error) {
	path := filepath.Join(root, ".namba", "hooks.toml")
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return hookConfig{}, nil
		}
		return hookConfig{}, err
	}
	rootAbs, err := filepath.Abs(root)
	if err != nil {
		return hookConfig{}, fmt.Errorf("resolve hook root: %w", err)
	}
	rootAbs = filepath.Clean(rootAbs)
	if resolvedRoot, resolveErr := filepath.EvalSymlinks(rootAbs); resolveErr == nil {
		rootAbs = filepath.Clean(resolvedRoot)
	}

	parsed, err := parseHookConfigTOML(rootAbs, string(data))
	if err != nil {
		return hookConfig{}, err
	}
	return parsed, nil
}

func parseHookConfigTOML(rootAbs, body string) (hookConfig, error) {
	type rawHook struct {
		name   string
		fields map[string]hookTOMLValue
	}
	var hooks []rawHook
	var current *rawHook
	seenHookNames := make(map[string]struct{})
	for lineNo, rawLine := range strings.Split(body, "\n") {
		line := strings.TrimSpace(stripHookComment(rawLine))
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "[") {
			if !strings.HasSuffix(line, "]") {
				return hookConfig{}, fmt.Errorf("parse hooks.toml line %d: missing closing bracket", lineNo+1)
			}
			section := strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(line, "["), "]"))
			current = nil
			if strings.HasPrefix(section, "hooks.") {
				name := strings.TrimSpace(strings.TrimPrefix(section, "hooks."))
				if !validHookName(name) {
					return hookConfig{}, fmt.Errorf("invalid hook name %q", name)
				}
				if _, exists := seenHookNames[name]; exists {
					return hookConfig{}, fmt.Errorf("parse hooks.toml line %d: duplicate hook table %q", lineNo+1, name)
				}
				seenHookNames[name] = struct{}{}
				hooks = append(hooks, rawHook{name: name, fields: make(map[string]hookTOMLValue)})
				current = &hooks[len(hooks)-1]
			}
			continue
		}
		if current == nil {
			return hookConfig{}, fmt.Errorf("parse hooks.toml line %d: key outside [hooks.<hook_name>] table", lineNo+1)
		}
		key, valueRaw, ok := strings.Cut(line, "=")
		if !ok {
			return hookConfig{}, fmt.Errorf("parse hooks.toml line %d: expected key = value", lineNo+1)
		}
		key = strings.TrimSpace(key)
		if key == "" {
			return hookConfig{}, fmt.Errorf("parse hooks.toml line %d: empty key", lineNo+1)
		}
		if _, exists := current.fields[key]; exists {
			return hookConfig{}, fmt.Errorf("parse hooks.toml line %d: duplicate key %q in hook %s", lineNo+1, key, current.name)
		}
		value, err := parseHookTOMLValue(strings.TrimSpace(valueRaw))
		if err != nil {
			return hookConfig{}, fmt.Errorf("parse hooks.toml line %d: %w", lineNo+1, err)
		}
		current.fields[key] = value
	}

	cfg := hookConfig{Hooks: make([]hookRegistration, 0, len(hooks))}
	for _, raw := range hooks {
		hook, err := buildHookRegistration(rootAbs, raw.name, raw.fields)
		if err != nil {
			return hookConfig{}, err
		}
		cfg.Hooks = append(cfg.Hooks, hook)
	}
	return cfg, nil
}

type hookTOMLValue struct {
	kind  string
	value string
}

func parseHookTOMLValue(raw string) (hookTOMLValue, error) {
	if raw == "" {
		return hookTOMLValue{}, fmt.Errorf("empty value")
	}
	if raw[0] == '"' || raw[0] == '\'' {
		quote := raw[0]
		if len(raw) < 2 || raw[len(raw)-1] != quote {
			return hookTOMLValue{}, fmt.Errorf("unterminated string")
		}
		value := raw[1 : len(raw)-1]
		if quote == '"' {
			value = strings.ReplaceAll(value, `\"`, `"`)
			value = strings.ReplaceAll(value, `\\`, `\`)
		}
		return hookTOMLValue{kind: "string", value: value}, nil
	}
	switch strings.ToLower(raw) {
	case "true", "false":
		return hookTOMLValue{kind: "bool", value: strings.ToLower(raw)}, nil
	}
	if _, err := strconv.Atoi(raw); err == nil {
		return hookTOMLValue{kind: "int", value: raw}, nil
	}
	return hookTOMLValue{}, fmt.Errorf("unsupported value %q", raw)
}

func buildHookRegistration(rootAbs, name string, fields map[string]hookTOMLValue) (hookRegistration, error) {
	required := []string{"event", "command", "cwd", "timeout", "enabled", "continue_on_failure"}
	for _, field := range required {
		if _, ok := fields[field]; !ok {
			return hookRegistration{}, fmt.Errorf("hook %s missing required field %s", name, field)
		}
	}
	eventRaw, err := hookStringField(fields, "event")
	if err != nil {
		return hookRegistration{}, fmt.Errorf("hook %s: %w", name, err)
	}
	event, ok := parseHookEvent(eventRaw)
	if !ok {
		return hookRegistration{}, fmt.Errorf("hook %s invalid hook event %q", name, eventRaw)
	}
	command, err := hookStringField(fields, "command")
	if err != nil {
		return hookRegistration{}, fmt.Errorf("hook %s: %w", name, err)
	}
	if strings.TrimSpace(command) == "" {
		return hookRegistration{}, fmt.Errorf("hook %s command is empty", name)
	}
	cwd, err := hookStringField(fields, "cwd")
	if err != nil {
		return hookRegistration{}, fmt.Errorf("hook %s: %w", name, err)
	}
	timeout, err := hookIntField(fields, "timeout")
	if err != nil {
		return hookRegistration{}, fmt.Errorf("hook %s: %w", name, err)
	}
	if timeout <= 0 {
		return hookRegistration{}, fmt.Errorf("hook %s timeout must be greater than zero", name)
	}
	enabled, err := hookBoolField(fields, "enabled")
	if err != nil {
		return hookRegistration{}, fmt.Errorf("hook %s: %w", name, err)
	}
	continueOnFailure, err := hookBoolField(fields, "continue_on_failure")
	if err != nil {
		return hookRegistration{}, fmt.Errorf("hook %s: %w", name, err)
	}
	resolvedCWD, err := resolveHookCWD(rootAbs, cwd)
	if err != nil {
		return hookRegistration{}, fmt.Errorf("hook %s: %w", name, err)
	}
	return hookRegistration{
		Name:              name,
		Event:             event,
		Command:           command,
		CWD:               cwd,
		ResolvedCWD:       resolvedCWD,
		TimeoutSeconds:    timeout,
		Enabled:           enabled,
		ContinueOnFailure: continueOnFailure,
	}, nil
}

func hookStringField(fields map[string]hookTOMLValue, name string) (string, error) {
	value := fields[name]
	if value.kind != "string" {
		return "", fmt.Errorf("%s must be a string", name)
	}
	return value.value, nil
}

func hookIntField(fields map[string]hookTOMLValue, name string) (int, error) {
	value := fields[name]
	if value.kind != "int" {
		return 0, fmt.Errorf("%s must be an integer", name)
	}
	parsed, err := strconv.Atoi(value.value)
	if err != nil {
		return 0, fmt.Errorf("%s must be an integer", name)
	}
	return parsed, nil
}

func hookBoolField(fields map[string]hookTOMLValue, name string) (bool, error) {
	value := fields[name]
	if value.kind != "bool" {
		return false, fmt.Errorf("%s must be a boolean", name)
	}
	return value.value == "true", nil
}

func parseHookEvent(value string) (hookEvent, bool) {
	event := hookEvent(strings.TrimSpace(value))
	switch event {
	case hookEventBeforePreflight, hookEventAfterPreflight, hookEventBeforeExecution, hookEventAfterExecution,
		hookEventAfterPatch, hookEventAfterBash, hookEventAfterMCPTool,
		hookEventBeforeValidation, hookEventAfterValidation, hookEventOnFailure:
		return event, true
	default:
		return "", false
	}
}

func (cfg hookConfig) enabledHooksForEvent(event hookEvent) []hookRegistration {
	hooks := make([]hookRegistration, 0)
	for _, hook := range cfg.Hooks {
		if hook.Enabled && hook.Event == event {
			hooks = append(hooks, hook)
		}
	}
	sort.SliceStable(hooks, func(i, j int) bool {
		return hooks[i].Name < hooks[j].Name
	})
	return hooks
}

func stripHookComment(line string) string {
	inSingle := false
	inDouble := false
	escaped := false
	for i, r := range line {
		switch {
		case escaped:
			escaped = false
		case r == '\\' && inDouble:
			escaped = true
		case r == '"' && !inSingle:
			inDouble = !inDouble
		case r == '\'' && !inDouble:
			inSingle = !inSingle
		case r == '#' && !inSingle && !inDouble:
			return line[:i]
		}
	}
	return line
}

func validHookName(name string) bool {
	if strings.TrimSpace(name) == "" || strings.ContainsAny(name, `/\`) || name == "." || name == ".." {
		return false
	}
	for _, r := range name {
		switch {
		case r >= 'a' && r <= 'z':
		case r >= 'A' && r <= 'Z':
		case r >= '0' && r <= '9':
		case r == '_' || r == '-' || r == '.':
		default:
			return false
		}
	}
	return true
}

func resolveHookCWD(rootAbs, cwd string) (string, error) {
	cwd = strings.TrimSpace(cwd)
	if cwd == "" {
		return "", fmt.Errorf("cwd is empty")
	}
	candidate := cwd
	if !filepath.IsAbs(candidate) {
		candidate = filepath.Join(rootAbs, filepath.FromSlash(candidate))
	}
	abs, err := filepath.Abs(candidate)
	if err != nil {
		return "", fmt.Errorf("resolve cwd: %w", err)
	}
	abs = filepath.Clean(abs)
	if abs == rootAbs {
		return rootAbs, nil
	}
	ancestor, err := nearestExistingCreateAncestor(rootAbs, abs)
	if err != nil {
		return "", fmt.Errorf("resolve cwd: %w", err)
	}
	resolvedAncestor, err := filepath.EvalSymlinks(ancestor)
	if err != nil {
		return "", fmt.Errorf("resolve cwd: %w", err)
	}
	if !createPathWithinRoot(rootAbs, resolvedAncestor) {
		return "", fmt.Errorf("cwd resolves outside the repository root")
	}
	if resolved, err := filepath.EvalSymlinks(abs); err == nil {
		abs = filepath.Clean(resolved)
	}
	if !createPathWithinRoot(rootAbs, abs) {
		return "", fmt.Errorf("cwd resolves outside the repository root")
	}
	return abs, nil
}

func runShellCommandWithInput(ctx context.Context, runner func(context.Context, string, []string, string, string) (string, string, error), command, dir, input string) (string, string, error) {
	if runtime.GOOS == "windows" {
		return runner(ctx, "powershell", []string{"-NoProfile", "-Command", command}, dir, input)
	}
	return runner(ctx, "sh", []string{"-c", command}, dir, input)
}

func hookExitCode(err error) int {
	if err == nil {
		return 0
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return exitErr.ExitCode()
	}
	text := err.Error()
	if strings.HasPrefix(text, "exit status ") {
		if code, parseErr := strconv.Atoi(strings.TrimSpace(strings.TrimPrefix(text, "exit status "))); parseErr == nil {
			return code
		}
	}
	return -1
}

func isHookFailureStatus(status string) bool {
	switch strings.TrimSpace(status) {
	case "", "completed":
		return false
	default:
		return true
	}
}

func hookStageStatus(passed bool, err error) string {
	if err != nil || !passed {
		return "failed"
	}
	return "passed"
}

func hookErrorSummary(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

func hookArtifactName(name string) string {
	var b strings.Builder
	for _, r := range name {
		switch {
		case r >= 'a' && r <= 'z':
			b.WriteRune(r)
		case r >= 'A' && r <= 'Z':
			b.WriteRune(r)
		case r >= '0' && r <= '9':
			b.WriteRune(r)
		case r == '_' || r == '-' || r == '.':
			b.WriteRune(r)
		default:
			b.WriteByte('_')
		}
	}
	if b.Len() == 0 {
		return "hook"
	}
	return b.String()
}

func maxInt64(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}
