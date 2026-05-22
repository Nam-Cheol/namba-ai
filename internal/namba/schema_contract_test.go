package namba

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestEvidenceSchemaFilesExistAndMatchContracts(t *testing.T) {
	t.Parallel()

	root := repoRootForHookTest(t)
	for _, contract := range requiredEvidenceSchemaVersions() {
		path := filepath.Join(root, "internal", "namba", "testdata", "schemas", schemaFileName(contract))
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read schema file for %s: %v", contract, err)
		}
		var schema map[string]any
		if err := json.Unmarshal(data, &schema); err != nil {
			t.Fatalf("schema %s should be JSON: %v", contract, err)
		}
		if got := stringField(schema, "$id"); got != contract {
			t.Fatalf("schema %s id = %q", contract, got)
		}
		if got, ok := schema["additionalProperties"].(bool); !ok || !got {
			t.Fatalf("schema %s should allow append-only optional fields", contract)
		}
	}
}

func TestEvidenceContractValidationAcceptsCurrentFixturesAndOptionalFields(t *testing.T) {
	t.Parallel()

	root := repoRootForHookTest(t)
	app := NewApp(nil, nil)
	app.now = func() time.Time { return time.Date(2026, 5, 22, 0, 0, 0, 0, time.UTC) }
	execution, err := buildExecutionEvidenceManifest(root, executionEvidenceOptions{
		ProjectRoot:      root,
		LogID:            "spec-062",
		RunID:            "spec-062",
		SpecID:           "SPEC-062",
		ExecutionMode:    executionModeSolo,
		Status:           "completed",
		GeneratedAt:      app.now(),
		FinalizedBy:      "schema-test",
		Validation:       executionEvidenceRefInput{Kind: "validation", NotApplicable: true},
		Progress:         executionEvidenceRefInput{Kind: "progress", NotApplicable: true},
		CodexDiagnostics: &codexDiagnosticsEvidence{SchemaVersion: codexDiagnosticsSchemaVersion},
	})
	if err != nil {
		t.Fatalf("build execution evidence: %v", err)
	}

	scorecardResult, err := app.runEvalSuite(context.Background(), root, evalOptions{suite: defaultEvalSuite, fixture: defaultHarnessEvalFixture, baseline: defaultHarnessEvalBase})
	if err != nil {
		t.Fatalf("run eval suite: %v", err)
	}
	report := collectNambaReport(newReportFixture(t), app.now(), reportOptions{})
	fixtures := map[string]any{
		executionEvidenceSchemaVersion:       execution,
		queueRunnerEvidenceSchemaVersion:     queueRunnerEvidence{SchemaVersion: queueRunnerEvidenceSchemaVersion, SpecID: "SPEC-062", Status: "completed", Runner: "cli", Validation: queueRunnerValidationEvidence{Passed: true}},
		codexDiagnosticsSchemaVersion:        completeCodexDiagnosticsFixture(),
		projectCodexDiagnosticsSchemaVersion: projectCodexDiagnosticsEvidence{SchemaVersion: projectCodexDiagnosticsSchemaVersion, GeneratedAt: app.now().Format(time.RFC3339), ProjectRoot: root, Diagnostics: completeCodexDiagnosticsFixture()},
		evalResultSchemaVersion:              scorecardResult,
		evalBaselineSchemaVersion:            evalBaseline{SchemaVersion: evalBaselineSchemaVersion, ResultSchemaVersion: evalResultSchemaVersion, Suite: defaultEvalSuite, CorpusVersion: "fixture", ScenarioFingerprints: map[string]string{"scenario": "fingerprint"}, MetricPassCounts: map[string]int{"route_selection": 1}, RequiredCoverage: []string{"fixture"}},
		evalScorecardSchemaVersion:           scorecardResult.Scorecard,
		reportSchemaVersion:                  report,
	}
	for contract, fixture := range fixtures {
		data, err := json.Marshal(fixture)
		if err != nil {
			t.Fatalf("marshal %s fixture: %v", contract, err)
		}
		var raw map[string]any
		if err := json.Unmarshal(data, &raw); err != nil {
			t.Fatalf("unmarshal %s fixture: %v", contract, err)
		}
		raw["future_optional_field"] = map[string]any{"allowed": true}
		data, err = json.Marshal(raw)
		if err != nil {
			t.Fatalf("marshal %s optional fixture: %v", contract, err)
		}
		if err := validateEvidenceContractJSON(contract, data); err != nil {
			t.Fatalf("%s fixture should validate: %v\n%s", contract, err, string(data))
		}
	}
}

func TestEvidenceContractValidationRejectsBreakingChanges(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		contract string
		body     string
		want     string
	}{
		{
			name:     "missing required field",
			contract: queueRunnerEvidenceSchemaVersion,
			body:     `{"schema_version":"queue-runner-evidence/v1","status":"completed","runner":"cli","validation":{"passed":true}}`,
			want:     "spec_id",
		},
		{
			name:     "type change",
			contract: queueRunnerEvidenceSchemaVersion,
			body:     `{"schema_version":"queue-runner-evidence/v1","spec_id":"SPEC-062","status":"completed","runner":"cli","validation":{"passed":"yes"}}`,
			want:     "validation.passed",
		},
		{
			name:     "unsupported schema version",
			contract: queueRunnerEvidenceSchemaVersion,
			body:     `{"schema_version":"queue-runner-evidence/v2","spec_id":"SPEC-062","status":"completed","runner":"cli","validation":{"passed":true}}`,
			want:     "schema_version",
		},
		{
			name:     "semantic breaking state",
			contract: executionEvidenceSchemaVersion,
			body:     `{"schema_version":"execution-evidence/v1","log_id":"x","run_id":"x","generated_at":"2026-05-22T00:00:00Z","status":"completed","finalization":{"finalized_at":"2026-05-22T00:00:00Z"},"request":{"state":"present"},"preflight":{"state":"invalid"},"execution":{"state":"present"},"validation":{"state":"present"},"progress":{"state":"not_applicable"},"extensions":{"browser":{"state":"not_applicable"},"runtime":{"state":"not_applicable"}},"hooks":[]}`,
			want:     "unsupported state",
		},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			err := validateEvidenceContractJSON(tc.contract, []byte(tc.body))
			if err == nil {
				t.Fatalf("expected %s to fail", tc.name)
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("expected error to contain %q, got %v", tc.want, err)
			}
		})
	}
}

func schemaFileName(contract string) string {
	return strings.NewReplacer("/", "-").Replace(contract) + ".schema.json"
}

func completeCodexDiagnosticsFixture() codexDiagnosticsEvidence {
	return codexDiagnosticsEvidence{
		SchemaVersion:  codexDiagnosticsSchemaVersion,
		GeneratedAt:    "2026-05-22T00:00:00Z",
		CodexAvailable: "detected",
		Version: codexVersionEvidence{
			Status:             "detected",
			ParseStatus:        "parsed",
			BaselineVersion:    codexDiagnosticsBaselineVersion,
			BaselineComparison: "meets_baseline",
		},
		Doctor:                  codexDoctorEvidence{Status: "passed"},
		Redaction:               codexRedactionEvidence{Status: "not_applicable"},
		Roots:                   codexRootsEvidence{NambaRoot: ".", EffectiveWorkspaceRootsStatus: "matched", ConfiguredWorkspaceRootsStatus: "matched"},
		WorkspaceRootComparison: codexRootComparison{Status: "matched"},
		RemoteControl:           codexRemoteControl{Status: "enabled", Source: "fixture"},
		RemoteEnvironments:      codexRemoteEnvironments{Status: "detected", Source: "fixture"},
	}
}
