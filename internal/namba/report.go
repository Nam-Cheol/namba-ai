package namba

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	reportSchemaVersion = "namba-report/v1"
	defaultReportFormat = "markdown"
)

type reportOptions struct {
	format string
	specID string
	since  string
	failOn string
	help   bool
}

type nambaReport struct {
	SchemaVersion string            `json:"schema_version"`
	GeneratedAt   string            `json:"generated_at"`
	Project       reportProject     `json:"project"`
	Summary       reportSummary     `json:"summary"`
	Runs          reportRuns        `json:"runs"`
	Queue         reportQueue       `json:"queue"`
	Specs         reportSpecs       `json:"specs"`
	Release       reportRelease     `json:"release"`
	Diagnostics   reportDiagnostics `json:"diagnostics"`
	Issues        []reportIssue     `json:"issues"`
	Warnings      []reportWarning   `json:"warnings"`
}

type reportProject struct {
	Name      string `json:"name"`
	Type      string `json:"type"`
	Language  string `json:"language"`
	Framework string `json:"framework"`
	Root      string `json:"root"`
	NambaDir  string `json:"namba_dir"`
	Manifest  string `json:"manifest"`
}

type reportSummary struct {
	Health                 string         `json:"health"`
	RunSuccessCount        int            `json:"run_success_count"`
	RunFailureCount        int            `json:"run_failure_count"`
	BlockedCommandCount    int            `json:"blocked_command_count"`
	BlockedReasons         map[string]int `json:"blocked_reasons"`
	MissingEvidenceCount   int            `json:"missing_evidence_count"`
	ValidationFailureCount int            `json:"validation_failure_count"`
	ReviewRequiredCount    int            `json:"review_required_count"`
	ReviewReadyCount       int            `json:"review_ready_count"`
	QueuePendingCount      int            `json:"queue_pending_count"`
	QueueRunningCount      int            `json:"queue_running_count"`
	QueueBlockedCount      int            `json:"queue_blocked_count"`
	QueueDoneCount         int            `json:"queue_done_count"`
	IssueCount             int            `json:"issue_count"`
	WarningCount           int            `json:"warning_count"`
	NextAction             string         `json:"next_action"`
}

type reportRuns struct {
	State                     string              `json:"state"`
	EvidenceCount             int                 `json:"evidence_count"`
	ExecutionCount            int                 `json:"execution_count"`
	ValidationCount           int                 `json:"validation_count"`
	ValidationAttemptCount    int                 `json:"validation_attempt_count"`
	ParallelCount             int                 `json:"parallel_count"`
	ParallelProgressLineCount int                 `json:"parallel_progress_line_count"`
	QueueEvidenceCount        int                 `json:"queue_evidence_count"`
	HeartbeatCount            int                 `json:"heartbeat_count"`
	SuccessCount              int                 `json:"success_count"`
	FailureCount              int                 `json:"failure_count"`
	Latest                    *reportRunEvidence  `json:"latest,omitempty"`
	ModelRouting              *reportModelRouting `json:"model_routing,omitempty"`
}

type reportModelRouting struct {
	Version               string         `json:"version"`
	TurnsByModel          map[string]int `json:"turns_by_model"`
	FallbackCount         int            `json:"fallback_count"`
	BlockedCount          int            `json:"blocked_count"`
	UnavailableUsageCount int            `json:"unavailable_usage_count"`
}

type reportRunEvidence struct {
	Path            string `json:"path"`
	SpecID          string `json:"spec_id,omitempty"`
	Status          string `json:"status"`
	GeneratedAt     string `json:"generated_at,omitempty"`
	ExecutionMode   string `json:"execution_mode,omitempty"`
	MissingEvidence int    `json:"missing_evidence"`
}

type reportQueue struct {
	State           string            `json:"state"`
	Status          string            `json:"status,omitempty"`
	OperatorState   string            `json:"operator_state,omitempty"`
	ActiveSpecID    string            `json:"active_spec_id,omitempty"`
	PendingCount    int               `json:"pending_count"`
	RunningCount    int               `json:"running_count"`
	BlockedCount    int               `json:"blocked_count"`
	DoneCount       int               `json:"done_count"`
	BlockedReasons  map[string]int    `json:"blocked_reasons"`
	StaleCandidates []reportStaleItem `json:"stale_candidates,omitempty"`
}

type reportSpecs struct {
	State                 string                `json:"state"`
	Total                 int                   `json:"total"`
	CompleteCount         int                   `json:"complete_count"`
	IncompleteCount       int                   `json:"incomplete_count"`
	ReviewReadyCount      int                   `json:"review_ready_count"`
	ReviewRequiredCount   int                   `json:"review_required_count"`
	MissingReadinessCount int                   `json:"missing_readiness_count"`
	MissingEvidenceCount  int                   `json:"missing_evidence_count"`
	Incomplete            []reportSpecSummary   `json:"incomplete,omitempty"`
	Readiness             []reportReviewSummary `json:"readiness,omitempty"`
	StaleCandidates       []reportStaleItem     `json:"stale_candidates,omitempty"`
}

type reportSpecSummary struct {
	SpecID          string   `json:"spec_id"`
	Path            string   `json:"path"`
	MissingFiles    []string `json:"missing_files,omitempty"`
	MissingEvidence []string `json:"missing_evidence,omitempty"`
}

type reportReviewSummary struct {
	SpecID  string `json:"spec_id"`
	Path    string `json:"path"`
	State   string `json:"state"`
	Cleared string `json:"cleared,omitempty"`
}

type reportRelease struct {
	State                    string `json:"state"`
	ChecklistPath            string `json:"checklist_path,omitempty"`
	ChecklistTotal           int    `json:"checklist_total"`
	ChecklistComplete        int    `json:"checklist_complete"`
	LatestReleaseNotePath    string `json:"latest_release_note_path,omitempty"`
	LatestReleaseNotePresent bool   `json:"latest_release_note_present"`
	Readiness                string `json:"readiness"`
}

type reportDiagnostics struct {
	State                    string `json:"state"`
	Path                     string `json:"path,omitempty"`
	GeneratedAt              string `json:"generated_at,omitempty"`
	CodexAvailable           string `json:"codex_available,omitempty"`
	VersionStatus            string `json:"version_status,omitempty"`
	BaselineComparison       string `json:"baseline_comparison,omitempty"`
	DoctorStatus             string `json:"doctor_status,omitempty"`
	WorkspaceRootStatus      string `json:"workspace_root_status,omitempty"`
	RemoteControlStatus      string `json:"remote_control_status,omitempty"`
	RemoteEnvironmentsStatus string `json:"remote_environments_status,omitempty"`
}

type reportIssue struct {
	Severity string `json:"severity"`
	Kind     string `json:"kind"`
	Path     string `json:"path,omitempty"`
	Message  string `json:"message"`
}

type reportWarning struct {
	Kind    string `json:"kind"`
	Path    string `json:"path,omitempty"`
	Message string `json:"message"`
}

type reportStaleItem struct {
	Kind      string `json:"kind"`
	ID        string `json:"id"`
	Path      string `json:"path,omitempty"`
	Reason    string `json:"reason"`
	UpdatedAt string `json:"updated_at,omitempty"`
}

func parseReportOptions(args []string) (reportOptions, error) {
	opts := reportOptions{format: defaultReportFormat}
	for i := 0; i < len(args); i++ {
		switch arg := args[i]; arg {
		case "--help", "-h":
			opts.help = true
		case "--json":
			opts.format = "json"
		case "--format":
			value, err := consumeFlagValue(args, &i, arg)
			if err != nil {
				return reportOptions{}, err
			}
			opts.format = strings.TrimSpace(value)
		case "--spec":
			value, err := consumeFlagValue(args, &i, arg)
			if err != nil {
				return reportOptions{}, err
			}
			opts.specID = strings.TrimSpace(value)
		case "--since":
			value, err := consumeFlagValue(args, &i, arg)
			if err != nil {
				return reportOptions{}, err
			}
			opts.since = strings.TrimSpace(value)
		case "--fail-on":
			value, err := consumeFlagValue(args, &i, arg)
			if err != nil {
				return reportOptions{}, err
			}
			opts.failOn = strings.TrimSpace(value)
		default:
			if strings.HasPrefix(arg, "--format=") {
				opts.format = strings.TrimSpace(strings.TrimPrefix(arg, "--format="))
				continue
			}
			if strings.HasPrefix(arg, "--spec=") {
				opts.specID = strings.TrimSpace(strings.TrimPrefix(arg, "--spec="))
				continue
			}
			if strings.HasPrefix(arg, "--since=") {
				opts.since = strings.TrimSpace(strings.TrimPrefix(arg, "--since="))
				continue
			}
			if strings.HasPrefix(arg, "--fail-on=") {
				opts.failOn = strings.TrimSpace(strings.TrimPrefix(arg, "--fail-on="))
				continue
			}
			return reportOptions{}, fmt.Errorf("unknown flag %q", arg)
		}
	}
	if opts.format == "" {
		return reportOptions{}, errors.New("--format cannot be empty")
	}
	switch opts.format {
	case "markdown", "text", "json":
	default:
		return reportOptions{}, fmt.Errorf("unsupported report format %q", opts.format)
	}
	switch opts.failOn {
	case "", "attention", "blocked", "stale":
	default:
		return reportOptions{}, fmt.Errorf("unsupported --fail-on value %q", opts.failOn)
	}
	if opts.since != "" {
		if _, err := parseReportSinceDuration(opts.since); err != nil {
			return reportOptions{}, err
		}
	}
	return opts, nil
}

func parseReportSinceDuration(value string) (time.Duration, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return 0, errors.New("--since cannot be empty")
	}
	if duration, err := time.ParseDuration(trimmed); err == nil {
		if duration <= 0 {
			return 0, errors.New("--since must be positive")
		}
		return duration, nil
	}
	unit := trimmed[len(trimmed)-1:]
	number := strings.TrimSpace(trimmed[:len(trimmed)-1])
	amount, err := strconv.Atoi(number)
	if err != nil || amount <= 0 {
		return 0, fmt.Errorf("unsupported --since value %q", value)
	}
	switch unit {
	case "d":
		return time.Duration(amount) * 24 * time.Hour, nil
	case "w":
		return time.Duration(amount) * 7 * 24 * time.Hour, nil
	default:
		return 0, fmt.Errorf("unsupported --since value %q", value)
	}
}

func reportSinceCutoff(now time.Time, since string) (time.Time, bool) {
	duration, err := parseReportSinceDuration(since)
	if err != nil {
		return time.Time{}, false
	}
	return now.Add(-duration), true
}

func (a *App) runReport(_ context.Context, args []string) error {
	opts, err := parseReportOptions(args)
	if err != nil {
		return commandUsageError("report", err)
	}
	if opts.help {
		return a.printCommandUsage("report")
	}
	root, err := a.requireProjectRoot()
	if err != nil {
		return err
	}
	report := collectNambaReport(root, a.now(), opts)
	output, err := renderNambaReport(report, opts.format)
	if err != nil {
		return commandExitError(2, err)
	}
	fmt.Fprint(a.stdout, output)
	if shouldFailReport(report, opts.failOn) {
		return commandExitError(1, fmt.Errorf("namba report health is %s", report.Summary.Health))
	}
	return nil
}

func reportUsageText() string {
	return strings.Join([]string{
		"namba report",
		"",
		"Usage:",
		"  namba report [--format markdown|text|json] [--json] [--spec SPEC-XXX] [--since 7d] [--fail-on attention|blocked|stale]",
		"",
		"Behavior:",
		"  Build a read-only local observability report from .namba state for operators, CI, and harness evidence checks.",
	}, "\n") + "\n"
}

func collectNambaReport(root string, now time.Time, opts reportOptions) nambaReport {
	projectCfg, _ := (&App{}).loadProjectConfig(root)
	sinceCutoff, hasSinceCutoff := reportSinceCutoff(now, opts.since)
	report := nambaReport{
		SchemaVersion: reportSchemaVersion,
		GeneratedAt:   now.Format(time.RFC3339),
		Project: reportProject{
			Name:      projectCfg.Name,
			Type:      projectCfg.ProjectType,
			Language:  projectCfg.Language,
			Framework: projectCfg.Framework,
			Root:      ".",
			NambaDir:  nambaDir,
			Manifest:  manifestPath,
		},
		Summary:  reportSummary{BlockedReasons: map[string]int{}},
		Queue:    reportQueue{BlockedReasons: map[string]int{}},
		Issues:   []reportIssue{},
		Warnings: []reportWarning{},
	}
	report.Runs = collectReportRuns(root, &report, sinceCutoff, hasSinceCutoff)
	report.Queue = collectReportQueue(root, now, &report, sinceCutoff, hasSinceCutoff)
	report.Specs = collectReportSpecs(root, now, opts.specID, &report)
	report.Release = collectReportRelease(root, &report)
	report.Diagnostics = collectReportDiagnostics(root, &report)
	finalizeReportSummary(&report)
	return report
}

func collectReportRuns(root string, report *nambaReport, sinceCutoff time.Time, hasSinceCutoff bool) reportRuns {
	runs := reportRuns{State: "missing"}
	runDir := filepath.Join(root, logsDir, "runs")
	entries, err := os.ReadDir(runDir)
	if err != nil {
		addReportWarning(report, "missing", filepath.ToSlash(filepath.Join(logsDir, "runs")), "run evidence directory is missing")
		return runs
	}
	runs.State = "present"
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		rel := filepath.ToSlash(filepath.Join(logsDir, "runs", name))
		path := filepath.Join(runDir, name)
		if !reportFileWithinWindow(path, sinceCutoff, hasSinceCutoff) {
			continue
		}
		switch {
		case strings.HasSuffix(name, "-evidence.json"):
			var evidence executionEvidenceManifest
			if readReportJSON(path, rel, &evidence, report) {
				if !reportArtifactWithinWindow(evidence.GeneratedAt, path, sinceCutoff, hasSinceCutoff) {
					continue
				}
				runs.EvidenceCount++
				missing := countMissingEvidenceRefs(evidence)
				if missing > 0 {
					report.Summary.MissingEvidenceCount += missing
				}
				if isRunSuccessStatus(evidence.Status) {
					runs.SuccessCount++
				} else if isRunFailureStatus(evidence.Status) {
					runs.FailureCount++
				}
				for _, hook := range evidence.Hooks {
					if hook.Blocking && hook.Status != "" && hook.Status != "passed" {
						report.Summary.BlockedCommandCount++
						reason := firstNonBlank(hook.FailureAction, hook.ErrorSummary, hook.HookName, "hook_blocked")
						report.Summary.BlockedReasons[reason]++
					}
				}
				routings := evidence.ModelRoutingTurns
				if len(routings) == 0 && evidence.ModelRouting != nil {
					routings = []modelRoutingEvidence{*evidence.ModelRouting}
				}
				for _, routing := range routings {
					if runs.ModelRouting == nil {
						runs.ModelRouting = &reportModelRouting{Version: routing.Version, TurnsByModel: map[string]int{}}
					}
					runs.ModelRouting.TurnsByModel[routing.RequestedModel]++
					if routing.State == modelRoutingStatusFallback {
						runs.ModelRouting.FallbackCount++
					}
					if routing.State == modelRoutingStatusBlocked {
						runs.ModelRouting.BlockedCount++
					}
					if routing.UsageState == modelRoutingUsageUnavailable {
						runs.ModelRouting.UnavailableUsageCount++
					}
				}
				candidate := reportRunEvidence{Path: rel, SpecID: evidence.SpecID, Status: evidence.Status, GeneratedAt: evidence.GeneratedAt, ExecutionMode: evidence.ExecutionMode, MissingEvidence: missing}
				if isLatestRun(runs.Latest, candidate) {
					runs.Latest = &candidate
				}
			}
		case strings.HasSuffix(name, "-execution.json"):
			var execution executionResult
			if readReportJSON(path, rel, &execution, report) && reportArtifactWithinWindow(firstNonBlank(execution.FinishedAt, execution.StartedAt), path, sinceCutoff, hasSinceCutoff) {
				runs.ExecutionCount++
				if !execution.Succeeded {
					runs.FailureCount++
				}
			}
		case strings.HasSuffix(name, "-validation.json"):
			var validation validationReport
			if readReportJSON(path, rel, &validation, report) && reportArtifactWithinWindow(firstNonBlank(validation.FinishedAt, validation.StartedAt), path, sinceCutoff, hasSinceCutoff) {
				runs.ValidationCount++
				if !validation.Passed {
					report.Summary.ValidationFailureCount++
				}
			}
		case strings.Contains(name, "-validation-attempt-") && strings.HasSuffix(name, ".json"):
			var validation validationReport
			if readReportJSON(path, rel, &validation, report) && reportArtifactWithinWindow(firstNonBlank(validation.FinishedAt, validation.StartedAt), path, sinceCutoff, hasSinceCutoff) {
				runs.ValidationAttemptCount++
				if !validation.Passed {
					report.Summary.ValidationFailureCount++
				}
			}
		case strings.HasSuffix(name, "-parallel.json"):
			runs.ParallelCount++
		case strings.HasSuffix(name, "-parallel-progress.jsonl"):
			runs.ParallelProgressLineCount += countReportJSONLLines(path, rel, report, sinceCutoff, hasSinceCutoff)
		case strings.HasSuffix(name, "-queue-evidence.json"):
			runs.QueueEvidenceCount++
		case strings.HasSuffix(name, "-heartbeat.json"):
			var heartbeat queueRunnerHeartbeat
			if readReportJSON(path, rel, &heartbeat, report) && reportArtifactWithinWindow(heartbeat.UpdatedAt, path, sinceCutoff, hasSinceCutoff) {
				runs.HeartbeatCount++
			}
		}
	}
	return runs
}

func collectReportQueue(root string, now time.Time, report *nambaReport, sinceCutoff time.Time, hasSinceCutoff bool) reportQueue {
	queue := reportQueue{State: "none", BlockedReasons: map[string]int{}}
	rel := filepath.ToSlash(filepath.Join(logsDir, "queue", "state.json"))
	path := filepath.Join(root, filepath.FromSlash(rel))
	var state queueState
	if _, err := os.Stat(path); err != nil {
		addReportWarning(report, "missing", rel, "queue state is absent")
		return queue
	}
	if !readReportJSON(path, rel, &state, report) {
		queue.State = "corrupt"
		return queue
	}
	queue.State = "present"
	queue.Status = state.Status
	queue.OperatorState = state.OperatorState
	queue.ActiveSpecID = state.ActiveSpecID
	for _, spec := range state.Specs {
		switch {
		case spec.OperatorState == queueOperatorBlocked || spec.Phase == queuePhaseBlocked || spec.Status == queueStateBlocked:
			queue.BlockedCount++
			reason := firstNonBlank(spec.Blocker, state.LastBlocker, "queue_blocked")
			queue.BlockedReasons[reason]++
			report.Summary.BlockedCommandCount++
			report.Summary.BlockedReasons[reason]++
		case spec.OperatorState == queueOperatorRunning || spec.Phase == queuePhaseRunning:
			queue.RunningCount++
		case spec.OperatorState == queueOperatorDone || spec.Phase == queuePhaseLanded || spec.Status == queueStateDone:
			queue.DoneCount++
		default:
			queue.PendingCount++
		}
	}
	for _, stale := range collectQueueHeartbeatStaleCandidates(root, now, report, sinceCutoff, hasSinceCutoff) {
		queue.StaleCandidates = append(queue.StaleCandidates, stale)
	}
	return queue
}

func collectReportSpecs(root string, now time.Time, specFilter string, report *nambaReport) reportSpecs {
	specs := reportSpecs{State: "missing"}
	specRoot := filepath.Join(root, specsDir)
	entries, err := os.ReadDir(specRoot)
	if err != nil {
		addReportWarning(report, "missing", specsDir, "SPEC directory is missing")
		return specs
	}
	specs.State = "present"
	for _, entry := range entries {
		if !entry.IsDir() || !strings.HasPrefix(entry.Name(), "SPEC-") {
			continue
		}
		specID := entry.Name()
		if specFilter != "" && specID != specFilter {
			continue
		}
		specs.Total++
		relDir := filepath.ToSlash(filepath.Join(specsDir, specID))
		missingFiles := missingSpecCoreFiles(root, specID)
		missingEvidence := missingHarnessEvidence(root, specID, report)
		if len(missingFiles) == 0 && len(missingEvidence) == 0 {
			specs.CompleteCount++
		} else {
			specs.IncompleteCount++
			specs.MissingEvidenceCount += len(missingEvidence)
			specs.Incomplete = append(specs.Incomplete, reportSpecSummary{
				SpecID:          specID,
				Path:            relDir,
				MissingFiles:    missingFiles,
				MissingEvidence: missingEvidence,
			})
			if specLooksStale(root, specID, now) {
				specs.StaleCandidates = append(specs.StaleCandidates, reportStaleItem{Kind: "spec", ID: specID, Path: relDir, Reason: "incomplete SPEC package is older than 30 days"})
			}
		}
		readiness := collectSpecReadiness(root, specID, report)
		specs.Readiness = append(specs.Readiness, readiness)
		switch readiness.State {
		case "ready":
			specs.ReviewReadyCount++
		case "missing":
			specs.MissingReadinessCount++
			specs.ReviewRequiredCount++
		default:
			specs.ReviewRequiredCount++
		}
	}
	sort.Slice(specs.Incomplete, func(i, j int) bool { return specs.Incomplete[i].SpecID < specs.Incomplete[j].SpecID })
	sort.Slice(specs.Readiness, func(i, j int) bool { return specs.Readiness[i].SpecID < specs.Readiness[j].SpecID })
	return specs
}

func collectReportRelease(root string, report *nambaReport) reportRelease {
	release := reportRelease{State: "unknown", Readiness: "unknown"}
	checklistRel := filepath.ToSlash(filepath.Join(projectDir, "release-checklist.md"))
	checklistPath := filepath.Join(root, filepath.FromSlash(checklistRel))
	if data, err := os.ReadFile(checklistPath); err == nil {
		release.State = "present"
		release.ChecklistPath = checklistRel
		release.ChecklistTotal, release.ChecklistComplete = countMarkdownChecks(string(data))
	} else {
		addReportWarning(report, "missing", checklistRel, "release checklist is missing")
	}
	releaseRel := latestReleaseNote(root)
	if releaseRel != "" {
		release.LatestReleaseNotePath = releaseRel
		release.LatestReleaseNotePresent = true
	}
	switch {
	case release.ChecklistTotal > 0 && release.ChecklistComplete < release.ChecklistTotal:
		release.Readiness = "attention"
	case !release.LatestReleaseNotePresent:
		release.Readiness = "attention"
	case release.State == "present":
		release.Readiness = "ready"
	default:
		release.Readiness = "unknown"
	}
	return release
}

func collectReportDiagnostics(root string, report *nambaReport) reportDiagnostics {
	rel := filepath.ToSlash(filepath.Join(logsDir, "project", "codex-diagnostics-evidence.json"))
	path := filepath.Join(root, filepath.FromSlash(rel))
	diagnostics := reportDiagnostics{State: "missing"}
	var evidence projectCodexDiagnosticsEvidence
	if _, err := os.Stat(path); err != nil {
		addReportWarning(report, "missing", rel, "Codex diagnostics evidence is missing")
		return diagnostics
	}
	if !readReportJSON(path, rel, &evidence, report) {
		diagnostics.State = "corrupt"
		return diagnostics
	}
	diagnostics.State = "present"
	diagnostics.Path = rel
	diagnostics.GeneratedAt = evidence.Diagnostics.GeneratedAt
	diagnostics.CodexAvailable = evidence.Diagnostics.CodexAvailable
	diagnostics.VersionStatus = evidence.Diagnostics.Version.Status
	diagnostics.BaselineComparison = evidence.Diagnostics.Version.BaselineComparison
	diagnostics.DoctorStatus = evidence.Diagnostics.Doctor.Status
	diagnostics.WorkspaceRootStatus = evidence.Diagnostics.WorkspaceRootComparison.Status
	diagnostics.RemoteControlStatus = evidence.Diagnostics.RemoteControl.Status
	diagnostics.RemoteEnvironmentsStatus = evidence.Diagnostics.RemoteEnvironments.Status
	return diagnostics
}

func finalizeReportSummary(report *nambaReport) {
	report.Summary.RunSuccessCount = report.Runs.SuccessCount
	report.Summary.RunFailureCount = report.Runs.FailureCount
	report.Summary.ReviewReadyCount = report.Specs.ReviewReadyCount
	report.Summary.ReviewRequiredCount = report.Specs.ReviewRequiredCount
	report.Summary.MissingEvidenceCount += report.Specs.MissingEvidenceCount
	report.Summary.QueuePendingCount = report.Queue.PendingCount
	report.Summary.QueueRunningCount = report.Queue.RunningCount
	report.Summary.QueueBlockedCount = report.Queue.BlockedCount
	report.Summary.QueueDoneCount = report.Queue.DoneCount
	report.Summary.IssueCount = len(report.Issues)
	report.Summary.WarningCount = len(report.Warnings)
	switch {
	case report.Summary.QueueBlockedCount > 0 || report.Summary.BlockedCommandCount > 0:
		report.Summary.Health = "blocked"
	case report.Summary.ValidationFailureCount > 0 || report.Summary.MissingEvidenceCount > 0 || report.Summary.ReviewRequiredCount > 0 || len(report.Issues) > 0:
		report.Summary.Health = "attention"
	case report.Runs.State == "missing" && report.Queue.State == "none" && report.Specs.Total == 0:
		report.Summary.Health = "unknown"
	default:
		report.Summary.Health = "ok"
	}
	report.Summary.NextAction = reportNextAction(*report)
}

func renderNambaReport(report nambaReport, format string) (string, error) {
	if format == "json" {
		data, err := json.MarshalIndent(report, "", "  ")
		if err != nil {
			return "", fmt.Errorf("render report json: %w", err)
		}
		return string(append(data, '\n')), nil
	}
	var b strings.Builder
	heading := "# Namba Observability Report"
	if format == "text" {
		heading = "Namba Observability Report"
	}
	fmt.Fprintf(&b, "%s\n\n", heading)
	fmt.Fprintf(&b, "- health: `%s`\n", report.Summary.Health)
	fmt.Fprintf(&b, "- generated_at: `%s`\n", report.GeneratedAt)
	fmt.Fprintf(&b, "- runs: %d success, %d failure\n", report.Summary.RunSuccessCount, report.Summary.RunFailureCount)
	fmt.Fprintf(&b, "- blocked commands: %d\n", report.Summary.BlockedCommandCount)
	fmt.Fprintf(&b, "- missing evidence: %d\n", report.Summary.MissingEvidenceCount)
	fmt.Fprintf(&b, "- validation failures: %d\n", report.Summary.ValidationFailureCount)
	fmt.Fprintf(&b, "- review readiness: %d ready, %d require follow-up\n", report.Summary.ReviewReadyCount, report.Summary.ReviewRequiredCount)
	fmt.Fprintf(&b, "- queue: %d pending, %d running, %d blocked, %d done\n", report.Summary.QueuePendingCount, report.Summary.QueueRunningCount, report.Summary.QueueBlockedCount, report.Summary.QueueDoneCount)
	fmt.Fprintf(&b, "\n## Top Issues\n\n")
	if len(report.Issues) == 0 {
		fmt.Fprintf(&b, "- none\n")
	} else {
		for _, issue := range report.Issues {
			fmt.Fprintf(&b, "- %s `%s`: %s\n", issue.Severity, issue.Kind, issue.Message)
		}
	}
	fmt.Fprintf(&b, "\n## Blocked Reasons\n\n")
	renderCountMap(&b, report.Summary.BlockedReasons)
	fmt.Fprintf(&b, "\n## Queue State\n\n")
	fmt.Fprintf(&b, "- state: `%s`\n- status: `%s`\n- active_spec: `%s`\n", report.Queue.State, fallbackOrValue(report.Queue.Status, "none"), fallbackOrValue(report.Queue.ActiveSpecID, "none"))
	fmt.Fprintf(&b, "\n## Latest Run\n\n")
	if report.Runs.Latest == nil {
		fmt.Fprintf(&b, "- no execution evidence found\n")
	} else {
		fmt.Fprintf(&b, "- `%s`: status `%s`, spec `%s`, missing evidence %d\n", report.Runs.Latest.Path, report.Runs.Latest.Status, fallbackOrValue(report.Runs.Latest.SpecID, "unknown"), report.Runs.Latest.MissingEvidence)
	}
	fmt.Fprintf(&b, "\n## Stale Candidates\n\n")
	stale := append([]reportStaleItem{}, report.Queue.StaleCandidates...)
	stale = append(stale, report.Specs.StaleCandidates...)
	if len(stale) == 0 {
		fmt.Fprintf(&b, "- none\n")
	} else {
		for _, item := range stale {
			fmt.Fprintf(&b, "- `%s` %s: %s\n", item.Kind, item.ID, item.Reason)
		}
	}
	fmt.Fprintf(&b, "\n## Release\n\n")
	fmt.Fprintf(&b, "- readiness: `%s`\n- checklist: %d/%d complete\n- latest release note: `%s`\n", report.Release.Readiness, report.Release.ChecklistComplete, report.Release.ChecklistTotal, fallbackOrValue(report.Release.LatestReleaseNotePath, "missing"))
	fmt.Fprintf(&b, "\n## Diagnostics\n\n")
	fmt.Fprintf(&b, "- state: `%s`\n- codex: `%s`\n- version: `%s`\n- doctor: `%s`\n- remote_control: `%s`\n- remote_environments: `%s`\n", report.Diagnostics.State, fallbackOrValue(report.Diagnostics.CodexAvailable, "unknown"), fallbackOrValue(report.Diagnostics.BaselineComparison, report.Diagnostics.VersionStatus), fallbackOrValue(report.Diagnostics.DoctorStatus, "unknown"), fallbackOrValue(report.Diagnostics.RemoteControlStatus, "unknown"), fallbackOrValue(report.Diagnostics.RemoteEnvironmentsStatus, "unknown"))
	fmt.Fprintf(&b, "\n## Next Action\n\n- %s\n", report.Summary.NextAction)
	return b.String(), nil
}

func renderCountMap(b *strings.Builder, counts map[string]int) {
	if len(counts) == 0 {
		fmt.Fprintf(b, "- none\n")
		return
	}
	keys := make([]string, 0, len(counts))
	for key := range counts {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		fmt.Fprintf(b, "- `%s`: %d\n", key, counts[key])
	}
}

func renderStatusJSON(report nambaReport) (string, error) {
	status := map[string]any{
		"schema_version": report.SchemaVersion,
		"generated_at":   report.GeneratedAt,
		"project":        report.Project,
		"summary": map[string]any{
			"health":                   report.Summary.Health,
			"run_success_count":        report.Summary.RunSuccessCount,
			"run_failure_count":        report.Summary.RunFailureCount,
			"blocked_command_count":    report.Summary.BlockedCommandCount,
			"missing_evidence_count":   report.Summary.MissingEvidenceCount,
			"validation_failure_count": report.Summary.ValidationFailureCount,
			"review_required_count":    report.Summary.ReviewRequiredCount,
			"review_ready_count":       report.Summary.ReviewReadyCount,
			"queue_pending_count":      report.Summary.QueuePendingCount,
			"queue_running_count":      report.Summary.QueueRunningCount,
			"queue_blocked_count":      report.Summary.QueueBlockedCount,
			"queue_done_count":         report.Summary.QueueDoneCount,
			"next_action":              report.Summary.NextAction,
		},
		"queue":       report.Queue,
		"diagnostics": report.Diagnostics,
	}
	data, err := json.MarshalIndent(status, "", "  ")
	if err != nil {
		return "", err
	}
	return string(append(data, '\n')), nil
}

func shouldFailReport(report nambaReport, failOn string) bool {
	switch failOn {
	case "blocked":
		return report.Summary.Health == "blocked"
	case "attention":
		return report.Summary.Health == "blocked" || report.Summary.Health == "attention"
	case "stale":
		return len(report.Queue.StaleCandidates) > 0 || len(report.Specs.StaleCandidates) > 0
	default:
		return false
	}
}

func readReportJSON(path, rel string, target any, report *nambaReport) bool {
	data, err := os.ReadFile(path)
	if err != nil {
		addReportWarning(report, "missing", rel, err.Error())
		return false
	}
	if err := json.Unmarshal(data, target); err != nil {
		addReportWarning(report, "corrupt", rel, err.Error())
		return false
	}
	return true
}

func countReportJSONLLines(path, rel string, report *nambaReport, sinceCutoff time.Time, hasSinceCutoff bool) int {
	file, err := os.Open(path)
	if err != nil {
		addReportWarning(report, "missing", rel, err.Error())
		return 0
	}
	defer file.Close()
	count := 0
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		var raw map[string]any
		if err := json.Unmarshal([]byte(line), &raw); err != nil {
			addReportWarning(report, "corrupt", rel, "partial JSONL corruption: "+err.Error())
			continue
		}
		if timestamp, ok := raw["timestamp"].(string); ok && !reportTimeWithinWindow(timestamp, sinceCutoff, hasSinceCutoff) {
			continue
		}
		count++
	}
	if err := scanner.Err(); err != nil {
		addReportWarning(report, "unknown", rel, err.Error())
	}
	return count
}

func reportFileWithinWindow(path string, sinceCutoff time.Time, hasSinceCutoff bool) bool {
	if !hasSinceCutoff {
		return true
	}
	info, err := os.Stat(path)
	return err == nil && !info.ModTime().Before(sinceCutoff)
}

func reportTimeWithinWindow(value string, sinceCutoff time.Time, hasSinceCutoff bool) bool {
	if !hasSinceCutoff {
		return true
	}
	timestamp, err := time.Parse(time.RFC3339, strings.TrimSpace(value))
	return err == nil && !timestamp.Before(sinceCutoff)
}

func reportArtifactWithinWindow(timestamp, path string, sinceCutoff time.Time, hasSinceCutoff bool) bool {
	if !hasSinceCutoff {
		return true
	}
	if parsed, err := time.Parse(time.RFC3339, strings.TrimSpace(timestamp)); err == nil {
		return !parsed.Before(sinceCutoff)
	}
	return reportFileWithinWindow(path, sinceCutoff, hasSinceCutoff)
}

func countMissingEvidenceRefs(evidence executionEvidenceManifest) int {
	refs := []executionEvidenceRef{evidence.Request, evidence.Preflight, evidence.Execution, evidence.Validation, evidence.Progress}
	count := 0
	for _, ref := range refs {
		if ref.State == executionEvidenceStateMissing {
			count++
		}
	}
	if evidence.Extensions.Browser.State == executionEvidenceStateMissing {
		count++
	}
	if evidence.Extensions.Runtime.State == executionEvidenceStateMissing {
		count++
	}
	return count
}

func missingSpecCoreFiles(root, specID string) []string {
	var missing []string
	for _, name := range []string{"spec.md", "plan.md", "acceptance.md"} {
		rel := filepath.ToSlash(filepath.Join(specsDir, specID, name))
		if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(rel))); err != nil {
			missing = append(missing, rel)
		}
	}
	return missing
}

func missingHarnessEvidence(root, specID string, report *nambaReport) []string {
	req, err := loadHarnessRequest(root, specID)
	if err != nil {
		addReportWarning(report, "corrupt", filepath.ToSlash(filepath.Join(specsDir, specID, "harness-request.json")), err.Error())
		return nil
	}
	if req == nil {
		return nil
	}
	harnessReport := validateHarnessEvidence(root, specID, *req)
	return append([]string{}, harnessReport.MissingEvidence...)
}

func collectSpecReadiness(root, specID string, report *nambaReport) reportReviewSummary {
	rel := specReviewReadinessPath(specID)
	path := filepath.Join(root, filepath.FromSlash(rel))
	data, err := os.ReadFile(path)
	if err != nil {
		return reportReviewSummary{SpecID: specID, Path: rel, State: "missing"}
	}
	body := string(data)
	cleared := parseClearedReviews(body)
	state := "required"
	if strings.Contains(body, "Advisory status: all current review tracks are marked clear") || strings.Contains(cleared, "3/3") {
		state = "ready"
	}
	if strings.TrimSpace(body) == "" {
		state = "corrupt"
		addReportWarning(report, "corrupt", rel, "readiness summary is empty")
	}
	return reportReviewSummary{SpecID: specID, Path: rel, State: state, Cleared: cleared}
}

func parseClearedReviews(body string) string {
	for _, line := range strings.Split(body, "\n") {
		if strings.Contains(line, "Cleared reviews:") {
			return strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(line), "- Cleared reviews:"))
		}
	}
	return ""
}

func specLooksStale(root, specID string, now time.Time) bool {
	info, err := os.Stat(filepath.Join(root, filepath.FromSlash(filepath.Join(specsDir, specID))))
	return err == nil && now.Sub(info.ModTime()) > 30*24*time.Hour
}

func collectQueueHeartbeatStaleCandidates(root string, now time.Time, report *nambaReport, sinceCutoff time.Time, hasSinceCutoff bool) []reportStaleItem {
	var stale []reportStaleItem
	matches, _ := filepath.Glob(filepath.Join(root, logsDir, "runs", "*-heartbeat.json"))
	for _, path := range matches {
		if !reportFileWithinWindow(path, sinceCutoff, hasSinceCutoff) {
			continue
		}
		rel, _ := filepath.Rel(root, path)
		rel = filepath.ToSlash(rel)
		var heartbeat queueRunnerHeartbeat
		if !readReportJSON(path, rel, &heartbeat, report) {
			continue
		}
		if !reportArtifactWithinWindow(heartbeat.UpdatedAt, path, sinceCutoff, hasSinceCutoff) {
			continue
		}
		updated, err := time.Parse(time.RFC3339, heartbeat.UpdatedAt)
		if err != nil {
			addReportWarning(report, "corrupt", rel, "heartbeat updated_at is invalid")
			continue
		}
		if now.Sub(updated) > queueRunnerStaleAfter {
			stale = append(stale, reportStaleItem{Kind: "queue_heartbeat", ID: heartbeat.SpecID, Path: rel, Reason: "runner heartbeat is stale", UpdatedAt: heartbeat.UpdatedAt})
		}
	}
	return stale
}

func countMarkdownChecks(body string) (int, int) {
	total, complete := 0, 0
	for _, line := range strings.Split(body, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "- [") || strings.HasPrefix(trimmed, "* [") {
			total++
			if strings.HasPrefix(trimmed, "- [x]") || strings.HasPrefix(trimmed, "- [X]") || strings.HasPrefix(trimmed, "* [x]") || strings.HasPrefix(trimmed, "* [X]") {
				complete++
			}
		}
	}
	return total, complete
}

func latestReleaseNote(root string) string {
	matches, _ := filepath.Glob(filepath.Join(root, ".namba", "releases", "*.md"))
	if len(matches) == 0 {
		return ""
	}
	sort.Strings(matches)
	rel, err := filepath.Rel(root, matches[len(matches)-1])
	if err != nil {
		return ""
	}
	return filepath.ToSlash(rel)
}

func isLatestRun(current *reportRunEvidence, candidate reportRunEvidence) bool {
	if current == nil {
		return true
	}
	return candidate.GeneratedAt > current.GeneratedAt || (candidate.GeneratedAt == current.GeneratedAt && candidate.Path > current.Path)
}

func isRunSuccessStatus(status string) bool {
	switch strings.TrimSpace(status) {
	case "completed", "passed", "success", "done":
		return true
	default:
		return false
	}
}

func isRunFailureStatus(status string) bool {
	status = strings.TrimSpace(status)
	return strings.Contains(status, "failed") || strings.Contains(status, "blocked")
}

func reportNextAction(report nambaReport) string {
	switch {
	case report.Summary.BlockedCommandCount > 0 || report.Summary.QueueBlockedCount > 0:
		return "Resolve the blocked queue or hook reason, then rerun validation."
	case report.Summary.ValidationFailureCount > 0:
		return "Inspect the latest validation report and repair failing checks."
	case report.Summary.MissingEvidenceCount > 0:
		return "Regenerate or attach the missing execution and harness evidence before handoff."
	case report.Summary.ReviewRequiredCount > 0:
		return "Refresh SPEC review readiness for packages that still require follow-up."
	default:
		return "No immediate recovery action is required."
	}
}

func addReportWarning(report *nambaReport, kind, path, message string) {
	report.Warnings = append(report.Warnings, reportWarning{Kind: kind, Path: path, Message: message})
}
