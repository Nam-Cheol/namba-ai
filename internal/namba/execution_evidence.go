package namba

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const executionEvidenceSchemaVersion = "execution-evidence/v1"

type executionEvidenceState string

const (
	executionEvidenceStatePresent       executionEvidenceState = "present"
	executionEvidenceStateMissing       executionEvidenceState = "missing"
	executionEvidenceStateNotApplicable executionEvidenceState = "not_applicable"
)

type executionEvidenceRef struct {
	Kind  string                 `json:"kind"`
	Path  string                 `json:"path,omitempty"`
	State executionEvidenceState `json:"state"`
}

type executionEvidenceSignalBundle struct {
	Kind  string                 `json:"kind"`
	State executionEvidenceState `json:"state"`
	Paths []string               `json:"paths,omitempty"`
}

type executionEvidenceExtension struct {
	State         executionEvidenceState          `json:"state"`
	Artifacts     []executionEvidenceRef          `json:"artifacts,omitempty"`
	SignalBundles []executionEvidenceSignalBundle `json:"signal_bundles,omitempty"`
}

type executionEvidenceExtensions struct {
	Browser executionEvidenceExtension `json:"browser"`
	Runtime executionEvidenceExtension `json:"runtime"`
}

type executionEvidenceFinalization struct {
	FinalizedAt  string `json:"finalized_at"`
	FinalizedBy  string `json:"finalized_by"`
	FailurePhase string `json:"failure_phase,omitempty"`
}

type executionEvidenceManifest struct {
	SchemaVersion string                        `json:"schema_version"`
	LogID         string                        `json:"log_id"`
	RunID         string                        `json:"run_id"`
	SpecID        string                        `json:"spec_id,omitempty"`
	GeneratedAt   string                        `json:"generated_at"`
	ExecutionMode string                        `json:"execution_mode,omitempty"`
	Advisory      bool                          `json:"advisory"`
	Status        string                        `json:"status"`
	Finalization  executionEvidenceFinalization `json:"finalization"`
	Request       executionEvidenceRef          `json:"request"`
	Preflight     executionEvidenceRef          `json:"preflight"`
	Execution     executionEvidenceRef          `json:"execution"`
	Validation    executionEvidenceRef          `json:"validation"`
	Progress      executionEvidenceRef          `json:"progress"`
	Extensions    executionEvidenceExtensions   `json:"extensions"`
}

type executionEvidenceRefInput struct {
	Kind          string
	Path          string
	NotApplicable bool
}

type executionEvidenceOptions struct {
	ProjectRoot          string
	LogID                string
	RunID                string
	SpecID               string
	ExecutionMode        executionMode
	Status               string
	GeneratedAt          time.Time
	FinalizedBy          string
	Request              executionEvidenceRefInput
	Preflight            executionEvidenceRefInput
	Execution            executionEvidenceRefInput
	Validation           executionEvidenceRefInput
	Progress             executionEvidenceRefInput
	BrowserArtifacts     []executionEvidenceRef
	RuntimeArtifacts     []executionEvidenceRef
	RuntimeSignalBundles []executionEvidenceSignalBundle
}

func executionEvidenceManifestPath(logID string) string {
	return filepath.ToSlash(filepath.Join(logsDir, "runs", strings.TrimSpace(logID)+"-evidence.json"))
}

func defaultExecutionEvidenceRef(logID, suffix, kind string) executionEvidenceRefInput {
	return executionEvidenceRefInput{
		Kind: kind,
		Path: filepath.ToSlash(filepath.Join(logsDir, "runs", strings.TrimSpace(logID)+suffix)),
	}
}

func (a *App) writeExecutionEvidenceManifest(projectRoot string, options executionEvidenceOptions) error {
	manifest, err := buildExecutionEvidenceManifest(projectRoot, options)
	if err != nil {
		return err
	}
	return writeJSONFile(filepath.Join(projectRoot, filepath.FromSlash(executionEvidenceManifestPath(manifest.LogID))), manifest)
}

func (a *App) writeRunExecutionEvidence(projectRoot, logID string, req executionRequest, status string) error {
	return a.writeExecutionEvidenceManifest(projectRoot, executionEvidenceOptions{
		ProjectRoot:   projectRoot,
		LogID:         logID,
		SpecID:        req.SpecID,
		ExecutionMode: req.Mode,
		Status:        status,
		GeneratedAt:   a.now(),
		FinalizedBy:   "executeRun",
	})
}

func (a *App) writeParallelExecutionEvidence(root, specID, runID, status string, dryRun bool) error {
	logID := strings.ToLower(strings.TrimSpace(specID)) + "-parallel"
	return a.writeExecutionEvidenceManifest(root, executionEvidenceOptions{
		ProjectRoot:   root,
		LogID:         logID,
		RunID:         runID,
		SpecID:        specID,
		ExecutionMode: executionModeParallel,
		Status:        status,
		GeneratedAt:   a.now(),
		FinalizedBy:   "parallelRunLifecycle.finishRun",
		Request: executionEvidenceRefInput{
			Kind:          "request",
			NotApplicable: true,
		},
		Preflight: executionEvidenceRefInput{
			Kind:          "preflight",
			Path:          filepath.ToSlash(filepath.Join(logsDir, "runs", strings.ToLower(strings.TrimSpace(specID))+"-parallel-preflight.json")),
			NotApplicable: dryRun,
		},
		Execution: executionEvidenceRefInput{
			Kind: "execution",
			Path: filepath.ToSlash(filepath.Join(logsDir, "runs", strings.ToLower(strings.TrimSpace(specID))+"-parallel.json")),
		},
		Validation: executionEvidenceRefInput{
			Kind:          "validation",
			NotApplicable: true,
		},
		Progress: executionEvidenceRefInput{
			Kind: "parallel_progress",
			Path: relativeParallelProgressLogPath(specID),
		},
	})
}

func buildExecutionEvidenceManifest(projectRoot string, options executionEvidenceOptions) (executionEvidenceManifest, error) {
	root := strings.TrimSpace(projectRoot)
	logID := strings.TrimSpace(options.LogID)
	if root == "" {
		return executionEvidenceManifest{}, fmt.Errorf("execution evidence requires project root")
	}
	if logID == "" {
		return executionEvidenceManifest{}, fmt.Errorf("execution evidence requires log id")
	}

	if options.GeneratedAt.IsZero() {
		options.GeneratedAt = time.Now()
	}

	requestInput := options.Request
	if requestInput.Kind == "" {
		requestInput = defaultExecutionEvidenceRef(logID, "-request.json", "request")
	}
	preflightInput := options.Preflight
	if preflightInput.Kind == "" {
		preflightInput = defaultExecutionEvidenceRef(logID, "-preflight.json", "preflight")
	}
	executionInput := options.Execution
	if executionInput.Kind == "" {
		executionInput = defaultExecutionEvidenceRef(logID, "-execution.json", "execution")
	}
	validationInput := options.Validation
	if validationInput.Kind == "" {
		validationInput = defaultExecutionEvidenceRef(logID, "-validation.json", "validation")
	}
	progressInput := options.Progress
	if progressInput.Kind == "" {
		progressInput = executionEvidenceRefInput{Kind: "progress", NotApplicable: true}
	}

	runtimeSignalBundles := append([]executionEvidenceSignalBundle(nil), options.RuntimeSignalBundles...)
	if validationAttempts := executionEvidenceValidationAttempts(root, logID); len(validationAttempts.Paths) > 0 {
		runtimeSignalBundles = append(runtimeSignalBundles, validationAttempts)
	}

	browser := normalizeExecutionEvidenceExtension(root, executionEvidenceExtension{
		Artifacts: append([]executionEvidenceRef(nil), options.BrowserArtifacts...),
	})
	runtime := normalizeExecutionEvidenceExtension(root, executionEvidenceExtension{
		Artifacts:     append([]executionEvidenceRef(nil), options.RuntimeArtifacts...),
		SignalBundles: runtimeSignalBundles,
	})

	status := strings.TrimSpace(options.Status)
	if status == "" {
		status = inferExecutionEvidenceStatus(requestInput, preflightInput, executionInput, validationInput)
	}

	return executionEvidenceManifest{
		SchemaVersion: executionEvidenceSchemaVersion,
		LogID:         logID,
		RunID:         firstNonBlank(strings.TrimSpace(options.RunID), logID),
		SpecID:        strings.TrimSpace(options.SpecID),
		GeneratedAt:   options.GeneratedAt.Format(time.RFC3339),
		ExecutionMode: string(normalizeExecutionMode(options.ExecutionMode)),
		Advisory:      true,
		Status:        status,
		Finalization: executionEvidenceFinalization{
			FinalizedAt:  options.GeneratedAt.Format(time.RFC3339),
			FinalizedBy:  firstNonBlank(strings.TrimSpace(options.FinalizedBy), "namba"),
			FailurePhase: executionEvidenceFailurePhase(status),
		},
		Request:    resolveExecutionEvidenceRef(root, requestInput),
		Preflight:  resolveExecutionEvidenceRef(root, preflightInput),
		Execution:  resolveExecutionEvidenceRef(root, executionInput),
		Validation: resolveExecutionEvidenceRef(root, validationInput),
		Progress:   resolveExecutionEvidenceRef(root, progressInput),
		Extensions: executionEvidenceExtensions{
			Browser: browser,
			Runtime: runtime,
		},
	}, nil
}

func inferExecutionEvidenceStatus(request, preflight, execution, validation executionEvidenceRefInput) string {
	if validation.NotApplicable {
		return "completed"
	}
	return "partial"
}

func executionEvidenceFailurePhase(status string) string {
	switch strings.TrimSpace(status) {
	case "preflight_failed":
		return "preflight"
	case "execution_failed":
		return "execution"
	case "validation_failed":
		return "validation"
	case "repair_failed":
		return "repair"
	case "progress_log_failed":
		return "progress"
	case "merge_blocked":
		return "merge"
	case "merge_failed":
		return "merge"
	case "cleanup_failed":
		return "cleanup"
	default:
		return ""
	}
}

func resolveExecutionEvidenceRef(root string, input executionEvidenceRefInput) executionEvidenceRef {
	ref := executionEvidenceRef{
		Kind: strings.TrimSpace(input.Kind),
		Path: strings.TrimSpace(input.Path),
	}
	if input.NotApplicable {
		ref.State = executionEvidenceStateNotApplicable
		return ref
	}
	if ref.Kind == "" {
		ref.Kind = "artifact"
	}
	if ref.Path == "" {
		ref.State = executionEvidenceStateMissing
		return ref
	}
	if executionEvidencePathExists(root, ref.Path) {
		ref.State = executionEvidenceStatePresent
		return ref
	}
	ref.State = executionEvidenceStateMissing
	return ref
}

func normalizeExecutionEvidenceExtension(root string, ext executionEvidenceExtension) executionEvidenceExtension {
	for i := range ext.Artifacts {
		if ext.Artifacts[i].Kind == "" {
			ext.Artifacts[i].Kind = "artifact"
		}
		if ext.Artifacts[i].State == "" {
			if strings.TrimSpace(ext.Artifacts[i].Path) == "" {
				ext.Artifacts[i].State = executionEvidenceStateMissing
			} else if executionEvidencePathExists(root, ext.Artifacts[i].Path) {
				ext.Artifacts[i].State = executionEvidenceStatePresent
			} else {
				ext.Artifacts[i].State = executionEvidenceStateMissing
			}
		}
	}
	for i := range ext.SignalBundles {
		ext.SignalBundles[i].Paths = uniqueSortedExecutionEvidencePaths(ext.SignalBundles[i].Paths)
		if ext.SignalBundles[i].Kind == "" {
			ext.SignalBundles[i].Kind = "signal_bundle"
		}
		if ext.SignalBundles[i].State == "" {
			if len(ext.SignalBundles[i].Paths) == 0 {
				ext.SignalBundles[i].State = executionEvidenceStateMissing
			} else {
				ext.SignalBundles[i].State = executionEvidenceStatePresent
			}
		}
	}

	if len(ext.Artifacts) == 0 && len(ext.SignalBundles) == 0 {
		ext.State = executionEvidenceStateNotApplicable
		return ext
	}

	ext.State = executionEvidenceStateMissing
	for _, artifact := range ext.Artifacts {
		if artifact.State == executionEvidenceStatePresent {
			ext.State = executionEvidenceStatePresent
			return ext
		}
	}
	for _, bundle := range ext.SignalBundles {
		if bundle.State == executionEvidenceStatePresent {
			ext.State = executionEvidenceStatePresent
			return ext
		}
	}
	return ext
}

func executionEvidenceValidationAttempts(root, logID string) executionEvidenceSignalBundle {
	pattern := filepath.Join(root, logsDir, "runs", strings.TrimSpace(logID)+"-validation-attempt-*.json")
	matches, err := filepath.Glob(pattern)
	if err != nil || len(matches) == 0 {
		return executionEvidenceSignalBundle{}
	}

	paths := make([]string, 0, len(matches))
	for _, match := range matches {
		rel, relErr := filepath.Rel(root, match)
		if relErr != nil {
			continue
		}
		paths = append(paths, filepath.ToSlash(rel))
	}
	if len(paths) == 0 {
		return executionEvidenceSignalBundle{}
	}
	return executionEvidenceSignalBundle{
		Kind:  "validation_attempts",
		State: executionEvidenceStatePresent,
		Paths: uniqueSortedExecutionEvidencePaths(paths),
	}
}

func executionEvidencePathExists(root, rel string) bool {
	if strings.TrimSpace(rel) == "" {
		return false
	}
	_, err := os.Stat(filepath.Join(root, filepath.FromSlash(rel)))
	return err == nil
}

func uniqueSortedExecutionEvidencePaths(paths []string) []string {
	if len(paths) == 0 {
		return nil
	}
	seen := make(map[string]bool, len(paths))
	result := make([]string, 0, len(paths))
	for _, path := range paths {
		normalized := filepath.ToSlash(strings.TrimSpace(path))
		if normalized == "" || seen[normalized] {
			continue
		}
		seen[normalized] = true
		result = append(result, normalized)
	}
	sort.Strings(result)
	return result
}

func changeSummaryLatestExecutionProofSection(root string) []string {
	manifest, relPath, ok := latestExecutionEvidence(root)
	if !ok {
		return nil
	}
	return []string{
		"## Latest Execution Proof",
		"",
		fmt.Sprintf("- Latest execution proof artifact: `%s`", relPath),
		fmt.Sprintf("- Proof target: `%s`", fallbackOrValue(manifest.SpecID, "unknown")),
		fmt.Sprintf("- Execution proof status: `%s`", manifest.Status),
		fmt.Sprintf("- Execution mode: `%s`", fallbackOrValue(manifest.ExecutionMode, "unknown")),
		fmt.Sprintf("- Base artifacts: `request=%s`, `preflight=%s`, `execution=%s`, `validation=%s`, `progress=%s`", manifest.Request.State, manifest.Preflight.State, manifest.Execution.State, manifest.Validation.State, manifest.Progress.State),
		fmt.Sprintf("- Browser evidence: `%s`", manifest.Extensions.Browser.State),
		fmt.Sprintf("- Runtime evidence: `%s`", manifest.Extensions.Runtime.State),
	}
}

func prChecklistLatestExecutionProofItem(root string) []string {
	manifest, relPath, ok := latestExecutionEvidence(root)
	if !ok {
		return nil
	}
	return []string{fmt.Sprintf("- [ ] Latest execution proof checked: `%s` (`%s`, target `%s`)", relPath, manifest.Status, fallbackOrValue(manifest.SpecID, "unknown"))}
}

func latestExecutionEvidence(root string) (executionEvidenceManifest, string, bool) {
	return latestExecutionEvidenceMatching(root, func(executionEvidenceManifest) bool { return true })
}

func latestExecutionEvidenceForSpec(root, specID string) (executionEvidenceManifest, string, bool) {
	if strings.TrimSpace(specID) == "" || normalizedLatestSpec(specID) == "none" {
		return executionEvidenceManifest{}, "", false
	}
	return latestExecutionEvidenceMatching(root, func(candidate executionEvidenceManifest) bool {
		return strings.TrimSpace(candidate.SpecID) == strings.TrimSpace(specID)
	})
}

func latestExecutionEvidenceMatching(root string, match func(executionEvidenceManifest) bool) (executionEvidenceManifest, string, bool) {
	matches, err := filepath.Glob(filepath.Join(root, logsDir, "runs", "*-evidence.json"))
	if err != nil || len(matches) == 0 {
		return executionEvidenceManifest{}, "", false
	}

	var (
		found      bool
		latest     executionEvidenceManifest
		latestPath string
		latestTime time.Time
	)
	for _, candidatePath := range matches {
		data, readErr := os.ReadFile(candidatePath)
		if readErr != nil {
			continue
		}
		var candidate executionEvidenceManifest
		if err := json.Unmarshal(data, &candidate); err != nil {
			continue
		}
		if !match(candidate) {
			continue
		}

		candidateTime := executionEvidenceTimestamp(candidatePath, candidate)
		if found && !candidateTime.After(latestTime) {
			continue
		}

		rel, relErr := filepath.Rel(root, candidatePath)
		if relErr != nil {
			continue
		}
		found = true
		latest = candidate
		latestPath = filepath.ToSlash(rel)
		latestTime = candidateTime
	}

	return latest, latestPath, found
}

func executionEvidenceTimestamp(path string, manifest executionEvidenceManifest) time.Time {
	if ts, err := time.Parse(time.RFC3339, manifest.GeneratedAt); err == nil {
		return ts
	}
	info, err := os.Stat(path)
	if err != nil {
		return time.Time{}
	}
	return info.ModTime()
}
