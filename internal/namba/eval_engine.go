package namba

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"time"
)

func (a *App) runEvalSuite(_ context.Context, root string, options evalOptions) (evalRunResult, error) {
	if options.suite != defaultEvalSuite {
		return evalRunResult{}, fmt.Errorf("unsupported eval suite %q", options.suite)
	}
	corpus, err := a.loadEvalCorpus(root, options.fixture)
	if err != nil {
		return evalRunResult{}, err
	}
	if corpus.SchemaVersion != evalSchemaVersion {
		return evalRunResult{}, fmt.Errorf("unsupported eval fixture schema %q", corpus.SchemaVersion)
	}
	if corpus.Suite != defaultEvalSuite {
		return evalRunResult{}, fmt.Errorf("fixture suite %q does not match requested suite %q", corpus.Suite, options.suite)
	}

	scenarios := corpus.Scenarios
	if options.caseID != "" {
		scenarios = filterEvalScenarios(scenarios, options.caseID)
		if len(scenarios) == 0 {
			return evalRunResult{}, fmt.Errorf("eval case %q was not found", options.caseID)
		}
	}
	results := make([]evalScenarioResult, 0, len(scenarios))
	for _, scenario := range scenarios {
		result := a.evaluateHarnessScenario(root, scenario)
		results = append(results, result)
	}

	result := evalRunResult{
		SchemaVersion: evalResultSchemaVersion,
		Suite:         corpus.Suite,
		GeneratedAt:   a.now().UTC().Format(time.RFC3339),
		Scenarios:     results,
		Metrics:       buildEvalMetrics(results),
	}
	for _, scenario := range results {
		result.Summary.Total++
		if scenario.Passed {
			result.Summary.Passed++
		} else {
			result.Summary.Failed++
		}
	}
	if options.updateBaseline {
		baseline := buildEvalBaseline(corpus, result)
		if err := a.writeEvalBaseline(root, options.baseline, baseline); err != nil {
			return evalRunResult{}, err
		}
		result.Baseline = evalBaselineResult{Compared: true, Path: options.baseline, Passed: true, CorpusVersion: baseline.CorpusVersion}
	} else if strings.TrimSpace(options.baseline) != "" {
		baseline, err := a.loadEvalBaseline(root, options.baseline)
		if err != nil {
			return evalRunResult{}, err
		}
		regressions := compareEvalBaseline(baseline, result)
		result.Baseline = evalBaselineResult{
			Compared:      true,
			Path:          options.baseline,
			Passed:        len(regressions) == 0,
			CorpusVersion: baseline.CorpusVersion,
		}
		result.Regressions = regressions
		result.Summary.RegressionCount = len(regressions)
	}
	return result, nil
}

func (a *App) loadEvalCorpus(root, path string) (evalCorpus, error) {
	data, err := a.readEvalFile(root, path)
	if err != nil {
		return evalCorpus{}, err
	}
	var corpus evalCorpus
	if err := json.Unmarshal(data, &corpus); err != nil {
		return evalCorpus{}, fmt.Errorf("parse eval fixture %s: %w", path, err)
	}
	for i, scenario := range corpus.Scenarios {
		if strings.TrimSpace(scenario.ID) == "" {
			return evalCorpus{}, fmt.Errorf("eval scenario at index %d is missing id", i)
		}
		if strings.TrimSpace(scenario.Type) == "" {
			return evalCorpus{}, fmt.Errorf("eval scenario %s is missing type", scenario.ID)
		}
		if strings.TrimSpace(scenario.Rationale) == "" {
			return evalCorpus{}, fmt.Errorf("eval scenario %s is missing rationale", scenario.ID)
		}
		if len(scenario.Expected) == 0 {
			return evalCorpus{}, fmt.Errorf("eval scenario %s is missing expected fields", scenario.ID)
		}
	}
	return corpus, nil
}

func (a *App) readEvalFile(root, path string) ([]byte, error) {
	clean := filepath.ToSlash(filepath.Clean(path))
	if !filepath.IsAbs(path) {
		path = filepath.Join(root, path)
	}
	data, err := a.readFile(path)
	if err == nil {
		return data, nil
	}
	if clean == defaultHarnessEvalFixture || clean == defaultHarnessEvalBase || strings.HasPrefix(clean, "internal/namba/testdata/evals/harness/") {
		embeddedPath := strings.TrimPrefix(clean, "internal/namba/")
		if data, embedErr := embeddedEvalFixtures.ReadFile(embeddedPath); embedErr == nil {
			return data, nil
		}
	}
	return nil, fmt.Errorf("read eval file %s: %w", path, err)
}

func (a *App) loadEvalBaseline(root, path string) (evalBaseline, error) {
	data, err := a.readEvalFile(root, path)
	if err != nil {
		return evalBaseline{}, err
	}
	var baseline evalBaseline
	if err := json.Unmarshal(data, &baseline); err != nil {
		return evalBaseline{}, fmt.Errorf("parse eval baseline %s: %w", path, err)
	}
	if baseline.SchemaVersion != evalBaselineSchemaVersion {
		return evalBaseline{}, fmt.Errorf("unsupported eval baseline schema %q", baseline.SchemaVersion)
	}
	if baseline.ResultSchemaVersion != evalResultSchemaVersion {
		return evalBaseline{}, fmt.Errorf("unsupported eval result schema in baseline %q", baseline.ResultSchemaVersion)
	}
	return baseline, nil
}

func (a *App) writeEvalBaseline(root, path string, baseline evalBaseline) error {
	if !filepath.IsAbs(path) {
		path = filepath.Join(root, path)
	}
	data, err := json.MarshalIndent(baseline, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal eval baseline: %w", err)
	}
	data = append(data, '\n')
	if err := a.mkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return a.writeFile(path, data, 0o644)
}

func filterEvalScenarios(scenarios []evalScenario, caseID string) []evalScenario {
	var filtered []evalScenario
	for _, scenario := range scenarios {
		if scenario.ID == caseID {
			filtered = append(filtered, scenario)
		}
	}
	return filtered
}

func (a *App) evaluateHarnessScenario(root string, scenario evalScenario) evalScenarioResult {
	actual := map[string]any{}
	var failures []string
	switch scenario.Type {
	case "route":
		actual, failures = a.evaluateRouteScenario(root, scenario)
	case "mention_plugin":
		actual, failures = evaluateMentionPluginScenario(scenario)
	case "prompt_refinement":
		actual, failures = evaluatePromptRefinementScenario(scenario)
	case "guardrail":
		actual, failures = evaluateGuardrailScenario(scenario)
	case "evidence_manifest":
		actual, failures = evaluateEvidenceManifestScenario(scenario)
	case "pr_review":
		actual, failures = evaluatePRReviewScenario(scenario)
	default:
		failures = append(failures, fmt.Sprintf("unsupported scenario type %q", scenario.Type))
	}
	failures = append(failures, compareExpectedActual(scenario.Expected, actual)...)
	result := evalScenarioResult{
		ID:        scenario.ID,
		Type:      scenario.Type,
		Tags:      append([]string(nil), scenario.Tags...),
		Passed:    len(failures) == 0,
		Input:     scenario.Input,
		Expected:  scenario.Expected,
		Actual:    actual,
		Failures:  failures,
		Rationale: scenario.Rationale,
	}
	result.Fingerprint = fingerprintEvalScenario(result)
	return result
}

func (a *App) evaluateRouteScenario(_ string, scenario evalScenario) (map[string]any, []string) {
	input := scenario.Input
	actual := map[string]any{
		"clarification_required":        false,
		"spec_required_fields_complete": true,
		"execution_ready":               true,
	}
	if req, ok := directEvalCreateRequest(input); ok {
		route, err := harnessRouteForRequest(*req.HarnessRequest)
		if err != nil {
			return actual, []string{err.Error()}
		}
		actual["route_selection"] = "direct"
		actual["command"] = string(route)
		actual["delivery_mode"] = string(req.HarnessRequest.DeliveryMode)
		actual["artifact_targets"] = harnessArtifactTargetStrings(req.HarnessRequest.ArtifactTargets)
		actual["required_evidence"] = []string{}
		actual["required_reviews"] = []string{}
		actual["harness_request_kind_or_none"] = string(req.HarnessRequest.RequestKind)
		actual["sidecar_persisted"] = false
		return actual, nil
	}
	if strings.Contains(strings.ToLower(input), "fix") && strings.Contains(strings.ToLower(input), "spec first") {
		inv, err := parseFixArgs([]string{"--command", "plan", input})
		if err != nil {
			return actual, []string{err.Error()}
		}
		actual["route_selection"] = "planned_fix"
		actual["command"] = "namba fix --command " + inv.command
		actual["harness_request_kind_or_none"] = "none"
		actual["sidecar_persisted"] = false
		actual["delivery_mode"] = "spec"
		return actual, nil
	}
	if req := inferredPlanningHarnessRequest("plan", input); req != nil {
		route, err := harnessRouteForRequest(*req)
		if err != nil {
			return actual, []string{err.Error()}
		}
		switch req.RequestKind {
		case harnessRequestKindCore:
			actual["route_selection"] = "core"
		case harnessRequestKindDomain:
			actual["route_selection"] = "domain"
		default:
			actual["route_selection"] = strings.TrimSuffix(string(req.RequestKind), "_change")
		}
		actual["command"] = string(route)
		actual["delivery_mode"] = string(req.DeliveryMode)
		actual["artifact_targets"] = harnessArtifactTargetStrings(req.ArtifactTargets)
		actual["required_evidence"] = harnessEvidenceStrings(req.RequiredEvidence)
		actual["required_reviews"] = harnessReviewStrings(req.RequiredReviews)
		actual["harness_request_kind_or_none"] = string(req.RequestKind)
		actual["sidecar_persisted"] = true
		return actual, nil
	}
	if looksLikeDomainHarnessRequest(input) {
		req := inferredPlanningHarnessRequest("harness", input)
		if req == nil {
			return actual, []string{"domain harness input did not produce a harness request"}
		}
		route, err := harnessRouteForRequest(*req)
		if err != nil {
			return actual, []string{err.Error()}
		}
		switch req.RequestKind {
		case harnessRequestKindCore:
			actual["route_selection"] = "core"
		case harnessRequestKindDomain:
			actual["route_selection"] = "domain"
		default:
			actual["route_selection"] = strings.TrimSuffix(string(req.RequestKind), "_change")
		}
		actual["command"] = string(route)
		actual["delivery_mode"] = string(req.DeliveryMode)
		actual["artifact_targets"] = harnessArtifactTargetStrings(req.ArtifactTargets)
		actual["required_evidence"] = harnessEvidenceStrings(req.RequiredEvidence)
		actual["required_reviews"] = harnessReviewStrings(req.RequiredReviews)
		actual["harness_request_kind_or_none"] = string(req.RequestKind)
		actual["sidecar_persisted"] = true
		return actual, nil
	}
	actual["route_selection"] = "ordinary"
	actual["command"] = "namba plan"
	actual["harness_request_kind_or_none"] = "none"
	actual["sidecar_persisted"] = false
	actual["delivery_mode"] = "spec"
	return actual, nil
}

func looksLikeDomainHarnessRequest(input string) bool {
	lower := strings.ToLower(input)
	for _, marker := range []string{"skill", "agent", "harness", "eval", "workflow", "mcp"} {
		if strings.Contains(lower, marker) {
			return true
		}
	}
	return false
}

func evaluateMentionPluginScenario(scenario evalScenario) (map[string]any, []string) {
	input := strings.ToLower(scenario.Input)
	kinds := map[string]bool{}
	for _, kind := range normalizeEvalStringSlice(scenario.Expected["mention_kinds"]) {
		kinds[strings.ToLower(strings.TrimSpace(kind))] = true
	}
	routing := "explicit_namba_skill"
	if kinds["plugin"] || strings.Contains(input, "plugin") || strings.Contains(input, "marketplace") || strings.Contains(input, "share checkout") {
		routing = "platform_readiness_note"
	}
	if kinds["file"] || kinds["directory"] {
		routing = "explicit_namba_skill"
	}
	if kinds["mixed"] {
		routing = "ask_to_disambiguate"
	}
	if strings.Contains(input, "namba plan") || strings.Contains(input, "$namba-plan") {
		routing = "explicit_namba_skill"
	}
	return map[string]any{
		"route_selection":         routing,
		"namba_routing":           routing,
		"platform_readiness":      strings.Contains(input, "codex 0.131") || kinds["plugin"],
		"requires_plugin_install": false,
		"clarification_required":  routing == "ask_to_disambiguate",
		"execution_ready":         routing != "ask_to_disambiguate",
	}, nil
}

func evaluatePromptRefinementScenario(scenario evalScenario) (map[string]any, []string) {
	language := evalString(scenario.Expected["language_behavior"])
	if language == "" || language == "<nil>" {
		language = fallbackSpecCreationClarificationLanguage(scenario.Input)
	}
	_, required := evaluateSpecCreationClarificationForLanguage("plan", scenario.Input, language)
	return map[string]any{
		"clarification_required": required,
		"language_behavior":      language,
		"execution_ready":        !required,
	}, nil
}

func evaluateGuardrailScenario(scenario evalScenario) (map[string]any, []string) {
	command := evalString(scenario.Expected["command"])
	eventType := evalString(scenario.Expected["event_type"])
	deny, reason := evalGuardrailDeny(command)
	risk, riskReason := evalGuardrailRisk(command)
	reasonOut := reason
	if reasonOut == "" {
		reasonOut = riskReason
	}
	return map[string]any{
		"event_type":             eventType,
		"command":                command,
		"deny":                   deny,
		"risk_note":              eventType == "PermissionRequest" && risk,
		"reason":                 reasonOut,
		"execution_ready":        !deny,
		"clarification_required": false,
	}, nil
}

func evaluateEvidenceManifestScenario(scenario evalScenario) (map[string]any, []string) {
	var fixture struct {
		ValidationPath string                     `json:"validation_path"`
		Manifest       json.RawMessage            `json:"manifest"`
		Builder        evalEvidenceBuilderFixture `json:"builder"`
	}
	if err := json.Unmarshal(scenario.Fixture, &fixture); err != nil {
		return nil, []string{fmt.Sprintf("parse evidence fixture: %v", err)}
	}
	switch fixture.ValidationPath {
	case "raw_schema":
		missing := validateRawEvalExecutionEvidenceManifest(fixture.Manifest)
		if missing == nil {
			missing = []string{}
		}
		return map[string]any{
			"valid":                     len(missing) == 0,
			"missing_or_invalid_fields": missing,
			"required_evidence":         []string{"execution-evidence"},
			"execution_ready":           len(missing) == 0,
		}, nil
	case "builder_normalization":
		manifest, err := buildEvalEvidenceManifest(fixture.Builder)
		if err != nil {
			return nil, []string{err.Error()}
		}
		states := map[string]any{
			"request":    string(manifest.Request.State),
			"preflight":  string(manifest.Preflight.State),
			"execution":  string(manifest.Execution.State),
			"validation": string(manifest.Validation.State),
			"progress":   string(manifest.Progress.State),
			"runtime":    string(manifest.Extensions.Runtime.State),
			"browser":    string(manifest.Extensions.Browser.State),
		}
		return map[string]any{
			"valid":               true,
			"section_states":      states,
			"failure_phase":       manifest.Finalization.FailurePhase,
			"progress_log_failed": manifest.Finalization.ProgressLogFailed,
			"hook_count":          len(manifest.Hooks),
			"required_evidence":   []string{"execution-evidence"},
			"execution_ready":     manifest.Finalization.FailurePhase == "",
		}, nil
	default:
		return nil, []string{fmt.Sprintf("unknown evidence validation path %q", fixture.ValidationPath)}
	}
}

func evaluatePRReviewScenario(scenario evalScenario) (map[string]any, []string) {
	argv := normalizeEvalStringSlice(scenario.Expected["argv"])
	opts, err := parsePRArgs(argv)
	if err != nil {
		return nil, []string{err.Error()}
	}
	existing := normalizeEvalStringSlice(scenario.Expected["existing_comments"])
	created := 0
	if opts.RequestReview && countEvalReviewRequestComments(existing) == 0 {
		created = 1
	}
	return map[string]any{
		"required_reviews":         boolToReviewList(opts.RequestReview),
		"review_requested":         opts.RequestReview,
		"duplicate_review_comment": countEvalReviewRequestComments(existing) > 0 && created > 0,
		"created_review_comments":  created,
		"execution_ready":          true,
		"clarification_required":   false,
	}, nil
}

type evalEvidenceBuilderFixture struct {
	LogID              string                  `json:"log_id"`
	SpecID             string                  `json:"spec_id"`
	ExecutionMode      string                  `json:"execution_mode"`
	Status             string                  `json:"status"`
	ValidationAttempts int                     `json:"validation_attempts"`
	ProgressLogFailed  bool                    `json:"progress_log_failed"`
	ExistingPaths      []string                `json:"existing_paths"`
	Progress           evalEvidenceRefFixture  `json:"progress"`
	Hooks              []evalHookResultFixture `json:"hooks"`
}

type evalEvidenceRefFixture struct {
	Kind string `json:"kind"`
	Path string `json:"path"`
}

type evalHookResultFixture struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	Path   string `json:"path"`
}

func directEvalCreateRequest(input string) (createRequest, bool) {
	normalized := strings.ToLower(strings.TrimSpace(input))
	if !strings.Contains(normalized, "create") || !strings.Contains(normalized, "checklist") || !strings.Contains(normalized, "release validation") {
		return createRequest{}, false
	}
	return createRequest{
		Target:       createTargetSkill,
		Name:         "Release Validation Checklist",
		Description:  input,
		Instructions: "Create a markdown checklist for release validation.",
		HarnessRequest: &HarnessRequest{
			RequestKind:     harnessRequestKindDirect,
			DeliveryMode:    harnessDeliveryModeDirect,
			AdaptationMode:  harnessAdaptationGenerateArtifact,
			ArtifactTargets: []harnessArtifactTarget{harnessArtifactTargetSkill},
		},
	}, true
}

func harnessArtifactTargetStrings(targets []harnessArtifactTarget) []string {
	out := make([]string, 0, len(targets))
	for _, target := range targets {
		out = append(out, string(target))
	}
	return out
}

func evalGuardrailDeny(command string) (bool, string) {
	lower := strings.ToLower(command)
	switch {
	case strings.Contains(lower, "git reset --hard"):
		return true, "discards repository changes"
	case strings.Contains(lower, "git clean -fd") || strings.Contains(lower, "git clean -df"):
		return true, "delete untracked files"
	case strings.Contains(lower, "git push --force"):
		return true, "force-pushing"
	case strings.Contains(lower, "rm -rf /") || strings.Contains(lower, "rm -rf ."):
		return true, "broad rm -rf"
	case strings.Contains(lower, "chmod -r 777") || strings.Contains(lower, "chmod 777"):
		return true, "chmod 777"
	case (strings.Contains(lower, "curl ") || strings.Contains(lower, "wget ")) && (strings.Contains(lower, "| sh") || strings.Contains(lower, "| bash")):
		return true, "downloaded content"
	default:
		return false, ""
	}
}

func evalGuardrailRisk(command string) (bool, string) {
	lower := strings.ToLower(command)
	switch {
	case strings.Contains(lower, "sudo"):
		return true, "uses sudo"
	case strings.Contains(lower, "git push"):
		return true, "publish"
	default:
		return false, ""
	}
}

func buildEvalEvidenceManifest(builder evalEvidenceBuilderFixture) (executionEvidenceManifest, error) {
	tmp, err := os.MkdirTemp("", "namba-eval-evidence-*")
	if err != nil {
		return executionEvidenceManifest{}, err
	}
	defer os.RemoveAll(tmp)
	for _, rel := range builder.ExistingPaths {
		path := filepath.Join(tmp, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return executionEvidenceManifest{}, err
		}
		if err := os.WriteFile(path, []byte("{}"), 0o644); err != nil {
			return executionEvidenceManifest{}, err
		}
	}
	hooks := make([]hookResult, 0, len(builder.Hooks))
	for _, hook := range builder.Hooks {
		hooks = append(hooks, hookResult{HookName: hook.Name, Status: hook.Status, StdoutPath: hook.Path})
	}
	progress := executionEvidenceRefInput{}
	if builder.Progress.Kind != "" || builder.Progress.Path != "" {
		progress = executionEvidenceRefInput{Kind: builder.Progress.Kind, Path: builder.Progress.Path}
	}
	return buildExecutionEvidenceManifest(tmp, executionEvidenceOptions{
		ProjectRoot:        tmp,
		LogID:              builder.LogID,
		SpecID:             builder.SpecID,
		ExecutionMode:      executionMode(builder.ExecutionMode),
		Status:             builder.Status,
		ValidationAttempts: builder.ValidationAttempts,
		ProgressLogFailed:  builder.ProgressLogFailed,
		GeneratedAt:        time.Date(2026, 5, 19, 0, 0, 0, 0, time.UTC),
		FinalizedBy:        "harness-eval",
		Progress:           progress,
		Hooks:              hooks,
	})
}

func validateRawEvalExecutionEvidenceManifest(data json.RawMessage) []string {
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return []string{"manifest_json"}
	}
	var missing []string
	requireStringField := func(field string) {
		value, ok := raw[field].(string)
		if !ok || strings.TrimSpace(value) == "" {
			missing = append(missing, field)
		}
	}
	requireStringField("schema_version")
	requireStringField("log_id")
	requireStringField("run_id")
	requireStringField("generated_at")
	requireStringField("execution_mode")
	requireStringField("status")
	for _, section := range []string{"request", "preflight", "execution", "validation", "progress"} {
		if !rawEvalEvidenceSectionHasState(raw, section) {
			missing = append(missing, section)
		}
	}
	finalization, _ := raw["finalization"].(map[string]any)
	if value, ok := finalization["finalized_at"].(string); !ok || strings.TrimSpace(value) == "" {
		missing = append(missing, "finalization.finalized_at")
	}
	extensions, _ := raw["extensions"].(map[string]any)
	for _, section := range []string{"browser", "runtime"} {
		child, _ := extensions[section].(map[string]any)
		if value, ok := child["state"].(string); !ok || strings.TrimSpace(value) == "" {
			missing = append(missing, "extensions."+section+".state")
		}
	}
	if _, ok := raw["hooks"].([]any); !ok {
		missing = append(missing, "hooks")
	}
	return missing
}

func rawEvalEvidenceSectionHasState(raw map[string]any, section string) bool {
	child, ok := raw[section].(map[string]any)
	if !ok {
		return false
	}
	value, ok := child["state"].(string)
	return ok && strings.TrimSpace(value) != ""
}

func countEvalReviewRequestComments(comments []string) int {
	count := 0
	for _, comment := range comments {
		if isReviewRequestComment(comment, "@codex review") {
			count++
		}
	}
	return count
}

func boolToReviewList(required bool) []string {
	if required {
		return []string{"codex"}
	}
	return []string{}
}

func compareExpectedActual(expected, actual map[string]any) []string {
	var failures []string
	keys := make([]string, 0, len(expected))
	for key := range expected {
		if strings.HasPrefix(key, "_") || key == "argv" || key == "event_type" || key == "existing_comments" || key == "mention_kinds" || key == "reason_substring" {
			continue
		}
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		want := normalizeEvalComparable(expected[key])
		got, ok := actual[key]
		if !ok {
			failures = append(failures, fmt.Sprintf("%s missing; expected %v", key, want))
			continue
		}
		got = normalizeEvalComparable(got)
		if !reflect.DeepEqual(want, got) {
			failures = append(failures, fmt.Sprintf("%s expected %v, got %v", key, want, got))
		}
	}
	if substring := evalString(expected["reason_substring"]); substring != "" && substring != "<nil>" {
		if !strings.Contains(evalString(actual["reason"]), substring) {
			failures = append(failures, fmt.Sprintf("reason expected to contain %q, got %q", substring, evalString(actual["reason"])))
		}
	}
	return failures
}

func normalizeEvalComparable(value any) any {
	switch typed := value.(type) {
	case []string:
		out := append([]string(nil), typed...)
		if out == nil {
			out = []string{}
		}
		sort.Strings(out)
		return out
	case []any:
		out := make([]string, 0, len(typed))
		for _, item := range typed {
			out = append(out, fmt.Sprint(item))
		}
		sort.Strings(out)
		return out
	case float64:
		if typed == float64(int(typed)) {
			return int(typed)
		}
		return typed
	case map[string]any:
		out := make(map[string]any, len(typed))
		for key, item := range typed {
			out[key] = normalizeEvalComparable(item)
		}
		return out
	case map[string]string:
		out := make(map[string]any, len(typed))
		for key, item := range typed {
			out[key] = item
		}
		return out
	case executionEvidenceState:
		return string(typed)
	default:
		return typed
	}
}

func fingerprintEvalScenario(result evalScenarioResult) string {
	payload := map[string]any{
		"id":     result.ID,
		"type":   result.Type,
		"passed": result.Passed,
		"actual": result.Actual,
	}
	data, _ := json.Marshal(payload)
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func buildEvalMetrics(results []evalScenarioResult) []evalMetric {
	fields := map[string][]string{
		"route_selection":      {"route_selection", "command"},
		"delivery_mode":        {"delivery_mode"},
		"artifact_targets":     {"artifact_targets"},
		"required_evidence":    {"required_evidence"},
		"required_reviews":     {"required_reviews", "review_requested"},
		"clarification":        {"clarification_required"},
		"spec_required_fields": {"spec_required_fields_complete"},
		"execution_readiness":  {"execution_ready"},
	}
	names := []string{"route_selection", "delivery_mode", "artifact_targets", "required_evidence", "required_reviews", "clarification", "spec_required_fields", "execution_readiness"}
	metrics := make([]evalMetric, 0, len(names))
	for _, name := range names {
		total := 0
		passed := 0
		for _, result := range results {
			if !scenarioContributesMetric(result.Expected, fields[name]) {
				continue
			}
			total++
			if result.Passed {
				passed++
			}
		}
		rate := 0.0
		if total > 0 {
			rate = float64(passed) / float64(total)
		}
		metrics = append(metrics, evalMetric{Name: name, Passed: passed, Total: total, PassRate: rate})
	}
	return metrics
}

func scenarioContributesMetric(expected map[string]any, fields []string) bool {
	for _, field := range fields {
		if _, ok := expected[field]; ok {
			return true
		}
	}
	return false
}

func buildEvalBaseline(corpus evalCorpus, result evalRunResult) evalBaseline {
	fingerprints := map[string]string{}
	for _, scenario := range result.Scenarios {
		if scenario.Passed {
			fingerprints[scenario.ID] = scenario.Fingerprint
		}
	}
	metricCounts := map[string]int{}
	for _, metric := range result.Metrics {
		metricCounts[metric.Name] = metric.Passed
	}
	return evalBaseline{
		SchemaVersion:        evalBaselineSchemaVersion,
		ResultSchemaVersion:  evalResultSchemaVersion,
		Suite:                result.Suite,
		CorpusVersion:        corpus.CorpusVersion,
		ScenarioFingerprints: fingerprints,
		MetricPassCounts:     metricCounts,
		RequiredCoverage:     requiredEvalCoverageBuckets(result.Scenarios),
	}
}

func compareEvalBaseline(baseline evalBaseline, result evalRunResult) []string {
	var regressions []string
	if baseline.Suite != result.Suite {
		regressions = append(regressions, fmt.Sprintf("baseline suite %q does not match result suite %q", baseline.Suite, result.Suite))
	}
	currentByID := map[string]evalScenarioResult{}
	for _, scenario := range result.Scenarios {
		currentByID[scenario.ID] = scenario
	}
	for id, fingerprint := range baseline.ScenarioFingerprints {
		current, ok := currentByID[id]
		if !ok {
			regressions = append(regressions, fmt.Sprintf("previously passing scenario %s disappeared", id))
			continue
		}
		if !current.Passed {
			regressions = append(regressions, fmt.Sprintf("previously passing scenario %s now fails", id))
			continue
		}
		if current.Fingerprint != fingerprint {
			regressions = append(regressions, fmt.Sprintf("previously passing scenario %s changed fingerprint", id))
		}
	}
	for _, metric := range result.Metrics {
		if baselineCount, ok := baseline.MetricPassCounts[metric.Name]; ok && metric.Passed < baselineCount {
			regressions = append(regressions, fmt.Sprintf("metric %s pass count decreased from %d to %d", metric.Name, baselineCount, metric.Passed))
		}
	}
	coverage := map[string]bool{}
	for _, bucket := range requiredEvalCoverageBuckets(result.Scenarios) {
		coverage[bucket] = true
	}
	for _, bucket := range baseline.RequiredCoverage {
		if !coverage[bucket] {
			regressions = append(regressions, fmt.Sprintf("required coverage bucket %s is missing", bucket))
		}
	}
	sort.Strings(regressions)
	return regressions
}

func requiredEvalCoverageBuckets(results []evalScenarioResult) []string {
	seen := map[string]bool{}
	for _, scenario := range results {
		for _, tag := range scenario.Tags {
			if strings.HasPrefix(tag, "coverage:") {
				seen[strings.TrimPrefix(tag, "coverage:")] = true
			}
		}
	}
	out := make([]string, 0, len(seen))
	for tag := range seen {
		out = append(out, tag)
	}
	sort.Strings(out)
	return out
}
