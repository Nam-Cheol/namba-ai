package namba

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	codexDiagnosticsBaselineVersion = "0.131.0"
	projectCodexDiagnosticsSchema   = "project-codex-diagnostics-evidence/v1"
	codexDoctorTimeout              = 10 * time.Second
)

type codexDiagnosticsEvidence struct {
	SchemaVersion           string                 `json:"schema_version"`
	GeneratedAt             string                 `json:"generated_at"`
	CodexAvailable          string                 `json:"codex_available"`
	Version                 codexVersionEvidence   `json:"version"`
	Doctor                  codexDoctorEvidence    `json:"doctor"`
	Redaction               codexRedactionEvidence `json:"redaction"`
	Roots                   codexRootsEvidence     `json:"roots"`
	WorkspaceRootComparison codexRootComparison    `json:"workspace_root_comparison"`
	SandboxMode             string                 `json:"sandbox_mode,omitempty"`
	ApprovalPolicy          string                 `json:"approval_policy,omitempty"`
	PermissionProfile       string                 `json:"permission_profile,omitempty"`
}

type codexVersionEvidence struct {
	Status             string `json:"status"`
	Raw                string `json:"raw,omitempty"`
	Parsed             string `json:"parsed,omitempty"`
	ParseStatus        string `json:"parse_status"`
	BaselineVersion    string `json:"baseline_version"`
	BaselineComparison string `json:"baseline_comparison"`
	Error              string `json:"error,omitempty"`
}

type codexDoctorEvidence struct {
	Status        string `json:"status"`
	StdoutPath    string `json:"stdout_path,omitempty"`
	StderrPath    string `json:"stderr_path,omitempty"`
	TimedOut      bool   `json:"timed_out"`
	TimeoutMS     int    `json:"timeout_ms,omitempty"`
	Error         string `json:"error,omitempty"`
	ExitCode      int    `json:"exit_code,omitempty"`
	LocalFallback bool   `json:"local_fallback"`
}

type codexRedactionEvidence struct {
	Status       string   `json:"status"`
	Applied      bool     `json:"applied"`
	Patterns     []string `json:"patterns,omitempty"`
	Replacement  string   `json:"replacement,omitempty"`
	RedactedLogs []string `json:"redacted_logs,omitempty"`
}

type codexRootsEvidence struct {
	NambaRoot                      string   `json:"namba_root"`
	GitRoot                        string   `json:"git_root,omitempty"`
	ConfiguredWorkspaceRoots       []string `json:"configured_workspace_roots,omitempty"`
	CodexEffectiveWorkspaceRoots   []string `json:"codex_effective_workspace_roots,omitempty"`
	EffectiveWorkspaceRootsStatus  string   `json:"effective_workspace_roots_status"`
	ConfiguredWorkspaceRootsStatus string   `json:"configured_workspace_roots_status"`
}

type codexRootComparison struct {
	Status     string   `json:"status"`
	Compared   []string `json:"compared,omitempty"`
	Mismatches []string `json:"mismatches,omitempty"`
	Message    string   `json:"message,omitempty"`
}

type projectCodexDiagnosticsEvidence struct {
	SchemaVersion string                   `json:"schema_version"`
	GeneratedAt   string                   `json:"generated_at"`
	ProjectRoot   string                   `json:"project_root"`
	Diagnostics   codexDiagnosticsEvidence `json:"codex_diagnostics"`
}

type codexDiagnosticsOptions struct {
	LogDir                string
	LogPrefix             string
	RunCommands           bool
	Request               *executionRequest
	SystemConfig          systemConfig
	ConfiguredRoots       []string
	EffectiveRoots        []string
	EffectiveRootsStatus  string
	PermissionProfile     string
	IncludeDoctorLogFiles bool
	DoctorTimeout         time.Duration
}

func (a *App) buildCodexDiagnosticsEvidence(ctx context.Context, root string, opts codexDiagnosticsOptions) codexDiagnosticsEvidence {
	now := a.now()
	evidence := codexDiagnosticsEvidence{
		SchemaVersion:  "codex-diagnostics-evidence/v1",
		GeneratedAt:    now.Format(time.RFC3339),
		CodexAvailable: "unavailable",
		Version: codexVersionEvidence{
			Status:             "unavailable",
			ParseStatus:        "unavailable",
			BaselineVersion:    codexDiagnosticsBaselineVersion,
			BaselineComparison: "unavailable",
		},
		Doctor: codexDoctorEvidence{
			Status:        "local_fallback",
			LocalFallback: true,
		},
		Redaction: codexRedactionEvidence{
			Status:      "not_applicable",
			Replacement: "[REDACTED]",
		},
		SandboxMode:       normalizeSandboxMode(firstNonBlank(requestSandboxMode(opts.Request), opts.SystemConfig.SandboxMode)),
		ApprovalPolicy:    normalizeApprovalPolicy(firstNonBlank(requestApprovalPolicy(opts.Request), opts.SystemConfig.ApprovalPolicy)),
		PermissionProfile: firstNonBlank(strings.TrimSpace(opts.PermissionProfile), strings.TrimSpace(a.getenv("NAMBA_CODEX_PERMISSION_PROFILE")), "unavailable"),
	}
	evidence.Roots = a.codexDiagnosticsRoots(ctx, root, opts)
	evidence.WorkspaceRootComparison = compareCodexWorkspaceRoots(evidence.Roots)

	if !opts.RunCommands {
		evidence.Version.Status = "unavailable"
		evidence.Version.ParseStatus = "unavailable"
		evidence.Doctor.Status = "unavailable"
		evidence.Doctor.LocalFallback = true
		return evidence
	}

	if _, err := a.lookPath("codex"); err != nil {
		evidence.CodexAvailable = "not_detected"
		evidence.Version.Error = err.Error()
		evidence.Doctor.Error = err.Error()
		return evidence
	}
	evidence.CodexAvailable = "detected"

	rawVersion, err := a.runBinary(ctx, "codex", []string{"--version"}, root)
	evidence.Version = buildCodexVersionEvidence(rawVersion, err)
	evidence.Doctor, evidence.Redaction = a.runCodexDoctorEvidence(ctx, root, opts)
	return evidence
}

func (a *App) codexDiagnosticsRoots(ctx context.Context, root string, opts codexDiagnosticsOptions) codexRootsEvidence {
	configured := append([]string{}, opts.ConfiguredRoots...)
	if opts.Request != nil {
		configured = append(configured, opts.Request.WorkDir)
		configured = append(configured, opts.Request.AddDirs...)
	}
	if len(configured) == 0 {
		configured = append(configured, root)
	}
	effective := append([]string{}, opts.EffectiveRoots...)
	if len(effective) == 0 {
		effective = parsePathList(a.getenv("NAMBA_CODEX_EFFECTIVE_WORKSPACE_ROOTS"))
	}
	effectiveStatus := firstNonBlank(strings.TrimSpace(opts.EffectiveRootsStatus), "unavailable")
	if len(effective) > 0 {
		effectiveStatus = "detected"
	}
	return codexRootsEvidence{
		NambaRoot:                      filepath.Clean(root),
		GitRoot:                        detectGitRootLocal(root),
		ConfiguredWorkspaceRoots:       normalizeWorkspaceRoots(root, configured),
		CodexEffectiveWorkspaceRoots:   normalizeWorkspaceRoots(root, effective),
		EffectiveWorkspaceRootsStatus:  effectiveStatus,
		ConfiguredWorkspaceRootsStatus: "detected",
	}
}

func (a *App) runCodexDoctorEvidence(ctx context.Context, root string, opts codexDiagnosticsOptions) (codexDoctorEvidence, codexRedactionEvidence) {
	timeout := opts.DoctorTimeout
	if timeout <= 0 {
		timeout = codexDoctorTimeout
	}
	doctor := codexDoctorEvidence{
		Status:    "unavailable",
		TimeoutMS: int(timeout / time.Millisecond),
	}
	redaction := codexRedactionEvidence{
		Status:      "not_applicable",
		Replacement: "[REDACTED]",
	}
	if a.runCmdWithInput == nil {
		doctor.Status = "local_fallback"
		doctor.LocalFallback = true
		return doctor, redaction
	}

	doctorCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	stdout, stderr, err := a.runCmdWithInput(doctorCtx, "codex", []string{"doctor"}, root, "")
	timedOut := errors.Is(doctorCtx.Err(), context.DeadlineExceeded)
	stdout, stdoutRedacted := redactCodexDiagnosticText(stdout)
	stderr, stderrRedacted := redactCodexDiagnosticText(stderr)
	redactedLogs := []string{}

	logDir := firstNonBlank(strings.TrimSpace(opts.LogDir), filepath.ToSlash(filepath.Join(logsDir, "runs")))
	logPrefix := firstNonBlank(strings.TrimSpace(opts.LogPrefix), "codex")
	if opts.IncludeDoctorLogFiles {
		stdoutRel := filepath.ToSlash(filepath.Join(logDir, logPrefix+"-codex-doctor-stdout.txt"))
		stderrRel := filepath.ToSlash(filepath.Join(logDir, logPrefix+"-codex-doctor-stderr.txt"))
		if writeErr := writeRunText(filepath.Join(root, filepath.FromSlash(stdoutRel)), stdout); writeErr == nil {
			doctor.StdoutPath = stdoutRel
			redactedLogs = append(redactedLogs, stdoutRel)
		} else if err == nil {
			err = writeErr
		}
		if writeErr := writeRunText(filepath.Join(root, filepath.FromSlash(stderrRel)), stderr); writeErr == nil {
			doctor.StderrPath = stderrRel
			redactedLogs = append(redactedLogs, stderrRel)
		} else if err == nil {
			err = writeErr
		}
	}

	if stdoutRedacted || stderrRedacted {
		redaction.Status = "redacted"
		redaction.Applied = true
		redaction.Patterns = []string{"token", "api_key", "authorization", "password", "secret"}
		redaction.RedactedLogs = redactedLogs
	} else if opts.IncludeDoctorLogFiles {
		redaction.Status = "checked"
		redaction.RedactedLogs = redactedLogs
	}

	doctor.TimedOut = timedOut
	if timedOut {
		doctor.Status = "timed_out"
		doctor.LocalFallback = true
		if err != nil {
			doctor.Error = err.Error()
		}
		return doctor, redaction
	}
	if err != nil {
		doctor.Status = "failed"
		doctor.Error = err.Error()
		doctor.ExitCode = commandExitCode(err)
		return doctor, redaction
	}
	doctor.Status = "detected"
	return doctor, redaction
}

func buildCodexVersionEvidence(raw string, err error) codexVersionEvidence {
	evidence := codexVersionEvidence{
		Status:             "unavailable",
		Raw:                strings.TrimSpace(raw),
		ParseStatus:        "unavailable",
		BaselineVersion:    codexDiagnosticsBaselineVersion,
		BaselineComparison: "unavailable",
	}
	if err != nil {
		evidence.Status = "failed"
		evidence.Error = err.Error()
		return evidence
	}
	if evidence.Raw == "" {
		return evidence
	}
	evidence.Status = "detected"
	parsed, ok := parseCodexVersion(evidence.Raw)
	if !ok {
		evidence.ParseStatus = "unparsable"
		evidence.BaselineComparison = "unavailable"
		return evidence
	}
	evidence.Parsed = parsed.String()
	evidence.ParseStatus = "parsed"
	evidence.BaselineComparison = compareCodexVersions(parsed, mustParseCodexVersion(codexDiagnosticsBaselineVersion))
	return evidence
}

type codexSemver struct {
	Major int
	Minor int
	Patch int
}

func (v codexSemver) String() string {
	return fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
}

var codexVersionPattern = regexp.MustCompile(`v?([0-9]+)\.([0-9]+)(?:\.([0-9]+))?`)

func parseCodexVersion(raw string) (codexSemver, bool) {
	matches := codexVersionPattern.FindStringSubmatch(strings.TrimSpace(raw))
	if len(matches) == 0 {
		return codexSemver{}, false
	}
	major, _ := strconv.Atoi(matches[1])
	minor, _ := strconv.Atoi(matches[2])
	patch := 0
	if matches[3] != "" {
		patch, _ = strconv.Atoi(matches[3])
	}
	return codexSemver{Major: major, Minor: minor, Patch: patch}, true
}

func mustParseCodexVersion(raw string) codexSemver {
	version, _ := parseCodexVersion(raw)
	return version
}

func compareCodexVersions(got, baseline codexSemver) string {
	switch {
	case got.Major < baseline.Major:
		return "older"
	case got.Major > baseline.Major:
		return "newer"
	case got.Minor < baseline.Minor:
		return "older"
	case got.Minor > baseline.Minor:
		return "newer"
	case got.Patch < baseline.Patch:
		return "older"
	case got.Patch > baseline.Patch:
		return "newer"
	default:
		return "equal"
	}
}

func redactCodexDiagnosticText(input string) (string, bool) {
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`(?i)(authorization:\s*bearer\s+)[^\s]+`),
		regexp.MustCompile(`(?i)((?:api[_-]?key|token|password|secret)\s*[:=]\s*)[^\s]+`),
	}
	output := input
	redacted := false
	for _, pattern := range patterns {
		next := pattern.ReplaceAllString(output, `${1}[REDACTED]`)
		if next != output {
			redacted = true
			output = next
		}
	}
	return output, redacted
}

func compareCodexWorkspaceRoots(roots codexRootsEvidence) codexRootComparison {
	compared := uniqueSortedExecutionEvidencePaths(append(append([]string{}, roots.ConfiguredWorkspaceRoots...), roots.CodexEffectiveWorkspaceRoots...))
	if len(roots.CodexEffectiveWorkspaceRoots) == 0 {
		return codexRootComparison{
			Status:   "unavailable",
			Compared: compared,
			Message:  "Codex effective workspace roots were unavailable; workflow is not blocked.",
		}
	}
	effective := map[string]bool{}
	for _, root := range roots.CodexEffectiveWorkspaceRoots {
		effective[root] = true
	}
	var mismatches []string
	for _, root := range roots.ConfiguredWorkspaceRoots {
		if !effective[root] {
			mismatches = append(mismatches, root)
		}
	}
	if len(mismatches) > 0 {
		return codexRootComparison{
			Status:     "advisory_mismatch",
			Compared:   compared,
			Mismatches: mismatches,
			Message:    "Configured workspace roots differ from Codex effective workspace roots; workflow is not blocked.",
		}
	}
	return codexRootComparison{
		Status:   "detected",
		Compared: compared,
		Message:  "Configured workspace roots match Codex effective workspace roots.",
	}
}

func normalizeWorkspaceRoots(base string, roots []string) []string {
	if len(roots) == 0 {
		return nil
	}
	result := make([]string, 0, len(roots))
	seen := map[string]bool{}
	for _, root := range roots {
		root = strings.TrimSpace(root)
		if root == "" {
			continue
		}
		if !filepath.IsAbs(root) {
			root = filepath.Join(base, root)
		}
		root = filepath.Clean(root)
		if seen[root] {
			continue
		}
		seen[root] = true
		result = append(result, root)
	}
	sortStrings(result)
	return result
}

func parsePathList(raw string) []string {
	fields := strings.FieldsFunc(raw, func(r rune) bool {
		return r == ',' || r == '\n' || r == ';'
	})
	var out []string
	for _, field := range fields {
		if strings.TrimSpace(field) != "" {
			out = append(out, strings.TrimSpace(field))
		}
	}
	return out
}

func detectGitRootLocal(root string) string {
	current := filepath.Clean(root)
	for {
		if _, err := os.Stat(filepath.Join(current, ".git")); err == nil {
			return current
		}
		parent := filepath.Dir(current)
		if parent == current {
			return ""
		}
		current = parent
	}
}

func requestSandboxMode(req *executionRequest) string {
	if req == nil {
		return ""
	}
	return req.SandboxMode
}

func requestApprovalPolicy(req *executionRequest) string {
	if req == nil {
		return ""
	}
	return req.ApprovalPolicy
}

func sortStrings(values []string) {
	for i := 1; i < len(values); i++ {
		for j := i; j > 0 && values[j] < values[j-1]; j-- {
			values[j], values[j-1] = values[j-1], values[j]
		}
	}
}

func formatCodexBaselineAdvice(evidence codexDiagnosticsEvidence) string {
	switch evidence.Version.BaselineComparison {
	case "older":
		return fmt.Sprintf("Codex baseline advice: detected %s, older than advisory baseline %s. Update Codex separately with upstream Codex tooling when you need 0.131 runtime visibility.", evidence.Version.Parsed, codexDiagnosticsBaselineVersion)
	case "equal":
		return fmt.Sprintf("Codex baseline advice: detected %s, matching advisory baseline %s.", evidence.Version.Parsed, codexDiagnosticsBaselineVersion)
	case "newer":
		return fmt.Sprintf("Codex baseline advice: detected %s, newer than advisory baseline %s.", evidence.Version.Parsed, codexDiagnosticsBaselineVersion)
	case "unavailable":
		if evidence.Version.ParseStatus == "unparsable" {
			return fmt.Sprintf("Codex baseline advice: Codex version output was unparsable; advisory baseline is %s.", codexDiagnosticsBaselineVersion)
		}
		return fmt.Sprintf("Codex baseline advice: Codex version is unavailable; advisory baseline is %s.", codexDiagnosticsBaselineVersion)
	default:
		return ""
	}
}
