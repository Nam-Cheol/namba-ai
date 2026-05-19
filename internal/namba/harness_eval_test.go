package namba

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

const harnessEvalDir = "internal/namba/testdata/evals/harness"

type harnessRouteEvalCase struct {
	Name                             string   `json:"name"`
	Input                            string   `json:"input"`
	ExpectedCategory                 string   `json:"expected_category"`
	ExpectedDeliveryMode             string   `json:"expected_delivery_mode"`
	ExpectedRequiredEvidence         []string `json:"expected_required_evidence"`
	ExpectedReviewFlags              []string `json:"expected_review_flags"`
	ExpectedCommand                  string   `json:"expected_command"`
	ExpectedHarnessRequestKindOrNone string   `json:"expected_harness_request_kind_or_none"`
	ExpectedSidecarPersisted         bool     `json:"expected_sidecar_persisted"`
	Rationale                        string   `json:"rationale"`
}

type harnessEvidenceEvalCase struct {
	Name                           string                            `json:"name"`
	ValidationPath                 string                            `json:"validation_path"`
	Manifest                       json.RawMessage                   `json:"manifest"`
	Builder                        harnessEvidenceBuilderFixture     `json:"builder"`
	ExpectedValid                  bool                              `json:"expected_valid"`
	ExpectedMissingOrInvalidFields []string                          `json:"expected_missing_or_invalid_fields"`
	ExpectedSectionStates          map[string]executionEvidenceState `json:"expected_section_states"`
	ExpectedFailurePhase           string                            `json:"expected_failure_phase"`
	ExpectedProgressLogFailed      bool                              `json:"expected_progress_log_failed"`
	ExpectedHookCount              *int                              `json:"expected_hook_count"`
	Rationale                      string                            `json:"rationale"`
}

type harnessEvidenceBuilderFixture struct {
	LogID              string                  `json:"log_id"`
	SpecID             string                  `json:"spec_id"`
	ExecutionMode      string                  `json:"execution_mode"`
	Status             string                  `json:"status"`
	ValidationAttempts int                     `json:"validation_attempts"`
	ProgressLogFailed  bool                    `json:"progress_log_failed"`
	ExistingPaths      []string                `json:"existing_paths"`
	Progress           evidenceRefInputFixture `json:"progress"`
	Hooks              []hookResultFixture     `json:"hooks"`
}

type evidenceRefInputFixture struct {
	Kind string `json:"kind"`
	Path string `json:"path"`
}

type hookResultFixture struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	Path   string `json:"path"`
}

type harnessPREvalCase struct {
	Name                           string   `json:"name"`
	Argv                           []string `json:"argv"`
	LegacyAutoCodexReview          bool     `json:"legacy_auto_codex_review"`
	ExistingComments               []string `json:"existing_comments"`
	ExpectedReviewRequested        bool     `json:"expected_review_requested"`
	ExpectedDuplicateReviewComment bool     `json:"expected_duplicate_review_comment"`
	Rationale                      string   `json:"rationale"`
}

func TestHarnessEvalFixtureDirectoryIsComplete(t *testing.T) {
	t.Parallel()

	required := []string{
		"README.md",
		"route_cases.json",
		"prompt_refinement_cases.json",
		"guardrail_cases.json",
		"evidence_manifest_cases.json",
		"pr_review_cases.json",
	}
	for _, name := range required {
		path := repoFixturePath(t, filepath.Join(harnessEvalDir, name))
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("missing harness eval fixture %s: %v", name, err)
		}
	}
}

func TestHarnessRouteEvalCases(t *testing.T) {
	t.Parallel()

	var cases []harnessRouteEvalCase
	readHarnessEvalFixture(t, "route_cases.json", &cases)
	if len(cases) != 5 {
		t.Fatalf("route_cases.json should stay small and curated: got %d cases", len(cases))
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			requireHarnessEvalCaseBasics(t, tc.Name, tc.Input, tc.Rationale)

			actual := evaluateHarnessRouteCase(t, tc)
			if actual.command != tc.ExpectedCommand {
				t.Fatal(formatHarnessEvalDiagnostic(tc.Name, tc.Input, "expected_command", tc.ExpectedCommand, actual.command, tc.Rationale))
			}
			if actual.requestKindOrNone != tc.ExpectedHarnessRequestKindOrNone {
				t.Fatal(formatHarnessEvalDiagnostic(tc.Name, tc.Input, "expected_harness_request_kind_or_none", tc.ExpectedHarnessRequestKindOrNone, actual.requestKindOrNone, tc.Rationale))
			}
			if actual.sidecarPersisted != tc.ExpectedSidecarPersisted {
				t.Fatal(formatHarnessEvalDiagnostic(tc.Name, tc.Input, "expected_sidecar_persisted", tc.ExpectedSidecarPersisted, actual.sidecarPersisted, tc.Rationale))
			}
			if tc.ExpectedDeliveryMode != "" && actual.deliveryMode != tc.ExpectedDeliveryMode {
				t.Fatal(formatHarnessEvalDiagnostic(tc.Name, tc.Input, "expected_delivery_mode", tc.ExpectedDeliveryMode, actual.deliveryMode, tc.Rationale))
			}
			if len(tc.ExpectedRequiredEvidence) > 0 && !reflect.DeepEqual(actual.requiredEvidence, tc.ExpectedRequiredEvidence) {
				t.Fatal(formatHarnessEvalDiagnostic(tc.Name, tc.Input, "expected_required_evidence", tc.ExpectedRequiredEvidence, actual.requiredEvidence, tc.Rationale))
			}
			if len(tc.ExpectedReviewFlags) > 0 && !reflect.DeepEqual(actual.reviewFlags, tc.ExpectedReviewFlags) {
				t.Fatal(formatHarnessEvalDiagnostic(tc.Name, tc.Input, "expected_review_flags", tc.ExpectedReviewFlags, actual.reviewFlags, tc.Rationale))
			}
		})
	}
}

func TestHarnessEvidenceManifestEvalCases(t *testing.T) {
	t.Parallel()

	var cases []harnessEvidenceEvalCase
	readHarnessEvalFixture(t, "evidence_manifest_cases.json", &cases)
	if len(cases) == 0 {
		t.Fatal("evidence_manifest_cases.json must contain eval cases")
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			requireHarnessEvalCaseBasics(t, tc.Name, tc.ValidationPath, tc.Rationale)

			switch tc.ValidationPath {
			case "raw_schema":
				missing := validateRawExecutionEvidenceManifest(tc.Manifest)
				valid := len(missing) == 0
				if valid != tc.ExpectedValid {
					t.Fatal(formatHarnessEvalDiagnostic(tc.Name, "raw_schema", "expected_valid", tc.ExpectedValid, valid, tc.Rationale))
				}
				if !sameStringSet(missing, tc.ExpectedMissingOrInvalidFields) {
					t.Fatal(formatHarnessEvalDiagnostic(tc.Name, "raw_schema", "expected_missing_or_invalid_fields", tc.ExpectedMissingOrInvalidFields, missing, tc.Rationale))
				}
			case "builder_normalization":
				manifest := buildHarnessEvidenceEvalManifest(t, tc)
				assertHarnessEvidenceEvalManifest(t, tc, manifest)
			default:
				t.Fatalf("unknown evidence validation_path %q in case %s", tc.ValidationPath, tc.Name)
			}
		})
	}
}

func TestHarnessPRReviewEvalCases(t *testing.T) {
	t.Parallel()

	var cases []harnessPREvalCase
	readHarnessEvalFixture(t, "pr_review_cases.json", &cases)
	if len(cases) == 0 {
		t.Fatal("pr_review_cases.json must contain eval cases")
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			requireHarnessEvalCaseBasics(t, tc.Name, strings.Join(tc.Argv, " "), tc.Rationale)

			opts, err := parsePRArgs(tc.Argv)
			if err != nil {
				t.Fatalf("parse PR eval argv for %s: %v", tc.Name, err)
			}
			if opts.RequestReview != tc.ExpectedReviewRequested {
				t.Fatal(formatHarnessEvalDiagnostic(tc.Name, strings.Join(tc.Argv, " "), "expected_review_requested", tc.ExpectedReviewRequested, opts.RequestReview, tc.Rationale))
			}

			createdReviewComments := exerciseEnsureReviewComment(t, tc, opts)
			duplicateReviewComment := countReviewRequestComments(tc.ExistingComments) > 0 && createdReviewComments > 0
			if duplicateReviewComment != tc.ExpectedDuplicateReviewComment {
				t.Fatal(formatHarnessEvalDiagnostic(tc.Name, strings.Join(tc.Argv, " "), "expected_duplicate_review_comment", tc.ExpectedDuplicateReviewComment, duplicateReviewComment, tc.Rationale))
			}
			if tc.LegacyAutoCodexReview && !opts.RequestReview && createdReviewComments > 0 {
				t.Fatal(formatHarnessEvalDiagnostic(tc.Name, strings.Join(tc.Argv, " "), "legacy_auto_codex_review", "ignored", "created review comment", tc.Rationale))
			}
			if tc.ExpectedReviewRequested && countReviewRequestComments(tc.ExistingComments) == 0 && createdReviewComments != 1 {
				t.Fatal(formatHarnessEvalDiagnostic(tc.Name, strings.Join(tc.Argv, " "), "review comment creation", "one new marker comment", createdReviewComments, tc.Rationale))
			}
		})
	}
}

func TestHarnessEvalDiagnosticFormatIncludesDebugContext(t *testing.T) {
	t.Parallel()

	diagnostic := formatHarnessEvalDiagnostic("case-name", "input text", "field", "want", "got", "because")
	for _, want := range []string{"case-name", "input text", "field", "want", "got", "because"} {
		if !strings.Contains(diagnostic, want) {
			t.Fatalf("diagnostic should contain %q, got %q", want, diagnostic)
		}
	}
}

type harnessRouteEvaluation struct {
	command           string
	requestKindOrNone string
	deliveryMode      string
	requiredEvidence  []string
	reviewFlags       []string
	sidecarPersisted  bool
}

func evaluateHarnessRouteCase(t *testing.T, tc harnessRouteEvalCase) harnessRouteEvaluation {
	t.Helper()

	switch tc.ExpectedCategory {
	case "core_harness_change":
		req := inferredPlanningHarnessRequest("plan", tc.Input)
		if req == nil {
			t.Fatalf("expected core harness request for %s", tc.Name)
		}
		route, err := harnessRouteForRequest(*req)
		if err != nil {
			t.Fatalf("route core harness request for %s: %v", tc.Name, err)
		}
		return harnessRouteEvaluation{
			command:           string(route),
			requestKindOrNone: string(req.RequestKind),
			deliveryMode:      string(req.DeliveryMode),
			requiredEvidence:  harnessEvidenceStrings(req.RequiredEvidence),
			reviewFlags:       harnessReviewStrings(req.RequiredReviews),
			sidecarPersisted:  true,
		}
	case "domain_harness_change":
		req := inferredPlanningHarnessRequest("harness", tc.Input)
		if req == nil {
			t.Fatalf("expected domain harness request for %s", tc.Name)
		}
		route, err := harnessRouteForRequest(*req)
		if err != nil {
			t.Fatalf("route domain harness request for %s: %v", tc.Name, err)
		}
		return harnessRouteEvaluation{
			command:           string(route),
			requestKindOrNone: string(req.RequestKind),
			deliveryMode:      string(req.DeliveryMode),
			requiredEvidence:  harnessEvidenceStrings(req.RequiredEvidence),
			reviewFlags:       harnessReviewStrings(req.RequiredReviews),
			sidecarPersisted:  true,
		}
	case "direct_artifact_creation":
		req, err := previewDirectCreateHarnessRoute(t, tc.Input)
		if err != nil {
			t.Fatalf("preview direct create harness route for %s: %v", tc.Name, err)
		}
		route, err := harnessRouteForRequest(req)
		if err != nil {
			t.Fatalf("route direct harness request for %s: %v", tc.Name, err)
		}
		return harnessRouteEvaluation{
			command:           string(route),
			requestKindOrNone: string(req.RequestKind),
			deliveryMode:      string(req.DeliveryMode),
			sidecarPersisted:  false,
		}
	case "ordinary_feature_or_product_plan":
		if req := inferredPlanningHarnessRequest("plan", tc.Input); req != nil {
			t.Fatalf("ordinary route should not create harness request: %+v", req)
		}
		return harnessRouteEvaluation{command: "namba plan", requestKindOrNone: "none"}
	case "planned_fix":
		inv, err := parseFixArgs([]string{"--command", "plan", tc.Input})
		if err != nil {
			t.Fatalf("parse planned fix for %s: %v", tc.Name, err)
		}
		if inv.command != "plan" {
			t.Fatalf("expected planned fix command=plan, got %+v", inv)
		}
		return harnessRouteEvaluation{command: "namba fix --command plan", requestKindOrNone: "none"}
	default:
		t.Fatalf("unknown route eval category %q", tc.ExpectedCategory)
		return harnessRouteEvaluation{}
	}
}

func exerciseEnsureReviewComment(t *testing.T, tc harnessPREvalCase, opts prOptions) int {
	t.Helper()

	if !opts.RequestReview {
		return 0
	}

	app := NewApp(&strings.Builder{}, &strings.Builder{})
	createdReviewComments := 0
	app.runCmd = func(_ context.Context, name string, args []string, _ string) (string, error) {
		switch {
		case name == "gh" && len(args) >= 4 && args[0] == "pr" && args[1] == "view":
			comments := make([]githubPRComment, 0, len(tc.ExistingComments))
			for _, body := range tc.ExistingComments {
				comments = append(comments, githubPRComment{Body: body})
			}
			return mustMarshalJSON(t, githubPullRequest{Comments: comments}), nil
		case name == "gh" && len(args) >= 5 && args[0] == "pr" && args[1] == "comment":
			bodyIndex := indexOfArg(args, "--body")
			if bodyIndex == -1 || bodyIndex+1 >= len(args) {
				t.Fatalf("expected review comment body in args: %v", args)
			}
			if !isReviewRequestComment(args[bodyIndex+1], "@codex review") {
				t.Fatalf("expected Namba review marker comment, got %q", args[bodyIndex+1])
			}
			createdReviewComments++
			return "", nil
		default:
			t.Fatalf("unexpected command while exercising ensureReviewComment: %s %v", name, args)
			return "", nil
		}
	}

	if err := app.ensureReviewComment(context.Background(), t.TempDir(), 17, "@codex review"); err != nil {
		t.Fatalf("ensure review comment for %s: %v", tc.Name, err)
	}
	return createdReviewComments
}

func countReviewRequestComments(comments []string) int {
	count := 0
	for _, comment := range comments {
		if isReviewRequestComment(comment, "@codex review") {
			count++
		}
	}
	return count
}

func previewDirectCreateHarnessRoute(t *testing.T, input string) (harnessRequest, error) {
	t.Helper()

	req, err := directCreateRequestFromFixtureInput(input)
	if err != nil {
		return harnessRequest{}, err
	}
	tmp, app := prepareCreateProject(t)
	preview, err := app.previewCreate(tmp, req)
	if err != nil {
		return harnessRequest{}, err
	}
	if preview.HarnessRequest == nil {
		return harnessRequest{}, fmt.Errorf("direct create preview did not retain harness request")
	}
	return *preview.HarnessRequest, nil
}

func directCreateRequestFromFixtureInput(input string) (createRequest, error) {
	normalized := strings.ToLower(strings.TrimSpace(input))
	if !strings.Contains(normalized, "create") || !strings.Contains(normalized, "checklist") || !strings.Contains(normalized, "release validation") {
		return createRequest{}, fmt.Errorf("direct artifact fixture input is not recognized: %q", input)
	}
	return createRequest{
		Target:       createTargetSkill,
		Name:         "Release Validation Checklist",
		Description:  input,
		Instructions: "Create a markdown checklist for release validation.",
		HarnessRequest: &HarnessRequest{
			RequestKind:      harnessRequestKindDirect,
			DeliveryMode:     harnessDeliveryModeDirect,
			AdaptationMode:   harnessAdaptationGenerateArtifact,
			ArtifactTargets:  []harnessArtifactTarget{harnessArtifactTargetSkill},
			RequiredEvidence: nil,
			RequiredReviews:  nil,
		},
	}, nil
}

func buildHarnessEvidenceEvalManifest(t *testing.T, tc harnessEvidenceEvalCase) executionEvidenceManifest {
	t.Helper()

	tmp := t.TempDir()
	for _, rel := range tc.Builder.ExistingPaths {
		writeTestFile(t, filepath.Join(tmp, filepath.FromSlash(rel)), "{}")
	}
	hooks := make([]hookResult, 0, len(tc.Builder.Hooks))
	for _, hook := range tc.Builder.Hooks {
		hooks = append(hooks, hookResult{
			HookName:   hook.Name,
			Status:     hook.Status,
			StdoutPath: hook.Path,
		})
	}
	progress := executionEvidenceRefInput{}
	if tc.Builder.Progress.Kind != "" || tc.Builder.Progress.Path != "" {
		progress = executionEvidenceRefInput{Kind: tc.Builder.Progress.Kind, Path: tc.Builder.Progress.Path}
	}

	manifest, err := buildExecutionEvidenceManifest(tmp, executionEvidenceOptions{
		ProjectRoot:        tmp,
		LogID:              tc.Builder.LogID,
		SpecID:             tc.Builder.SpecID,
		ExecutionMode:      executionMode(tc.Builder.ExecutionMode),
		Status:             tc.Builder.Status,
		ValidationAttempts: tc.Builder.ValidationAttempts,
		ProgressLogFailed:  tc.Builder.ProgressLogFailed,
		GeneratedAt:        time.Date(2026, 5, 19, 0, 0, 0, 0, time.UTC),
		FinalizedBy:        "harness-eval",
		Progress:           progress,
		Hooks:              hooks,
	})
	if err != nil {
		t.Fatalf("build execution evidence eval manifest for %s: %v", tc.Name, err)
	}
	return manifest
}

func assertHarnessEvidenceEvalManifest(t *testing.T, tc harnessEvidenceEvalCase, manifest executionEvidenceManifest) {
	t.Helper()

	if !tc.ExpectedValid {
		t.Fatal(formatHarnessEvalDiagnostic(tc.Name, "builder_normalization", "expected_valid", tc.ExpectedValid, true, tc.Rationale))
	}
	states := map[string]executionEvidenceState{
		"request":    manifest.Request.State,
		"preflight":  manifest.Preflight.State,
		"execution":  manifest.Execution.State,
		"validation": manifest.Validation.State,
		"progress":   manifest.Progress.State,
		"runtime":    manifest.Extensions.Runtime.State,
		"browser":    manifest.Extensions.Browser.State,
	}
	for key, want := range tc.ExpectedSectionStates {
		if got := states[key]; got != want {
			t.Fatal(formatHarnessEvalDiagnostic(tc.Name, "builder_normalization", key, want, got, tc.Rationale))
		}
	}
	if tc.ExpectedFailurePhase != "" && manifest.Finalization.FailurePhase != tc.ExpectedFailurePhase {
		t.Fatal(formatHarnessEvalDiagnostic(tc.Name, "builder_normalization", "failure_phase", tc.ExpectedFailurePhase, manifest.Finalization.FailurePhase, tc.Rationale))
	}
	if tc.ExpectedProgressLogFailed != manifest.Finalization.ProgressLogFailed {
		t.Fatal(formatHarnessEvalDiagnostic(tc.Name, "builder_normalization", "progress_log_failed", tc.ExpectedProgressLogFailed, manifest.Finalization.ProgressLogFailed, tc.Rationale))
	}
	if tc.ExpectedHookCount != nil && len(manifest.Hooks) != *tc.ExpectedHookCount {
		t.Fatal(formatHarnessEvalDiagnostic(tc.Name, "builder_normalization", "hook_count", *tc.ExpectedHookCount, len(manifest.Hooks), tc.Rationale))
	}
}

func validateRawExecutionEvidenceManifest(data json.RawMessage) []string {
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
		if !rawEvidenceSectionHasState(raw, section) {
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

func rawEvidenceSectionHasState(raw map[string]any, section string) bool {
	child, ok := raw[section].(map[string]any)
	if !ok {
		return false
	}
	value, ok := child["state"].(string)
	return ok && strings.TrimSpace(value) != ""
}

func readHarnessEvalFixture(t *testing.T, name string, target any) {
	t.Helper()

	data, err := os.ReadFile(repoFixturePath(t, filepath.Join(harnessEvalDir, name)))
	if err != nil {
		t.Fatalf("read harness eval fixture %s: %v", name, err)
	}
	if err := json.Unmarshal(data, target); err != nil {
		t.Fatalf("unmarshal harness eval fixture %s: %v", name, err)
	}
}

func requireHarnessEvalCaseBasics(t *testing.T, name, subject, rationale string) {
	t.Helper()

	if strings.TrimSpace(name) == "" {
		t.Fatal("harness eval case is missing name")
	}
	if strings.TrimSpace(subject) == "" {
		t.Fatalf("harness eval case %s is missing input, command, or manifest identifier", name)
	}
	if strings.TrimSpace(rationale) == "" {
		t.Fatalf("harness eval case %s is missing rationale", name)
	}
}

func formatHarnessEvalDiagnostic(caseName, subject, expectedField string, expected, actual any, rationale string) string {
	return fmt.Sprintf("harness eval case %q failed\nsubject: %s\nfield: %s\nexpected: %v\nactual: %v\nrationale: %s", caseName, subject, expectedField, expected, actual, rationale)
}

func sameStringSet(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	seen := map[string]int{}
	for _, item := range left {
		seen[item]++
	}
	for _, item := range right {
		seen[item]--
	}
	for _, count := range seen {
		if count != 0 {
			return false
		}
	}
	return true
}
