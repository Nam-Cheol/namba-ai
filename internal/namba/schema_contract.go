package namba

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

const (
	queueRunnerEvidenceSchemaVersion     = "queue-runner-evidence/v1"
	codexDiagnosticsSchemaVersion        = "codex-diagnostics-evidence/v1"
	projectCodexDiagnosticsSchemaVersion = "project-codex-diagnostics-evidence/v1"
)

func requiredEvidenceSchemaVersions() []string {
	return []string{
		executionEvidenceSchemaVersion,
		queueRunnerEvidenceSchemaVersion,
		codexDiagnosticsSchemaVersion,
		projectCodexDiagnosticsSchemaVersion,
		evalResultSchemaVersion,
		evalBaselineSchemaVersion,
		evalScorecardSchemaVersion,
		reportSchemaVersion,
	}
}

func validateEvidenceContractJSON(contract string, data []byte) error {
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("parse %s JSON: %w", contract, err)
	}
	if got := stringField(raw, "schema_version"); got != contract {
		return fmt.Errorf("schema_version = %q, want %q", got, contract)
	}
	var failures []string
	for _, field := range evidenceRequiredFields(contract) {
		if err := requireEvidenceField(raw, field.path, field.kind); err != nil {
			failures = append(failures, err.Error())
		}
	}
	failures = append(failures, validateEvidenceSemantics(contract, raw)...)
	if len(failures) > 0 {
		sort.Strings(failures)
		return fmt.Errorf("%s contract violation: %s", contract, strings.Join(failures, "; "))
	}
	return nil
}

type evidenceRequiredField struct {
	path string
	kind string
}

func evidenceRequiredFields(contract string) []evidenceRequiredField {
	switch contract {
	case executionEvidenceSchemaVersion:
		return []evidenceRequiredField{
			{"schema_version", "string"}, {"log_id", "string"}, {"run_id", "string"}, {"generated_at", "string"}, {"status", "string"},
			{"finalization.finalized_at", "string"}, {"request.state", "string"}, {"preflight.state", "string"}, {"execution.state", "string"}, {"validation.state", "string"}, {"progress.state", "string"},
			{"extensions.browser.state", "string"}, {"extensions.runtime.state", "string"}, {"hooks", "array"},
		}
	case queueRunnerEvidenceSchemaVersion:
		return []evidenceRequiredField{{"schema_version", "string"}, {"spec_id", "string"}, {"status", "string"}, {"runner", "string"}, {"validation.passed", "bool"}}
	case codexDiagnosticsSchemaVersion:
		return []evidenceRequiredField{
			{"schema_version", "string"}, {"generated_at", "string"}, {"codex_available", "string"}, {"version.status", "string"}, {"version.parse_status", "string"}, {"doctor.status", "string"},
			{"redaction.status", "string"}, {"roots.namba_root", "string"}, {"workspace_root_comparison.status", "string"}, {"remote_control.status", "string"}, {"remote_environments.status", "string"},
		}
	case projectCodexDiagnosticsSchemaVersion:
		return []evidenceRequiredField{{"schema_version", "string"}, {"generated_at", "string"}, {"project_root", "string"}, {"codex_diagnostics.schema_version", "string"}}
	case evalResultSchemaVersion:
		return []evidenceRequiredField{{"schema_version", "string"}, {"suite", "string"}, {"generated_at", "string"}, {"summary.total", "number"}, {"metrics", "array"}, {"scenarios", "array"}, {"regressions", "array"}}
	case evalBaselineSchemaVersion:
		return []evidenceRequiredField{{"schema_version", "string"}, {"result_schema_version", "string"}, {"suite", "string"}, {"corpus_version", "string"}, {"scenario_fingerprints", "object"}, {"metric_pass_counts", "object"}, {"required_coverage", "array"}}
	case evalScorecardSchemaVersion:
		return []evidenceRequiredField{{"schema_version", "string"}, {"suite", "string"}, {"corpus_version", "string"}, {"generated_at", "string"}, {"summary.total", "number"}, {"metrics", "array"}, {"required_coverage", "array"}, {"scenarios", "array"}, {"schema_validation.status", "string"}}
	case reportSchemaVersion:
		return []evidenceRequiredField{{"schema_version", "string"}, {"generated_at", "string"}, {"project.name", "string"}, {"summary.health", "string"}, {"runs.state", "string"}, {"queue.state", "string"}, {"specs.state", "string"}, {"release.state", "string"}, {"diagnostics.state", "string"}, {"issues", "array"}, {"warnings", "array"}}
	default:
		return nil
	}
}

func validateEvidenceSemantics(contract string, raw map[string]any) []string {
	var failures []string
	switch contract {
	case executionEvidenceSchemaVersion:
		for _, path := range []string{"request.state", "preflight.state", "execution.state", "validation.state", "progress.state", "extensions.browser.state", "extensions.runtime.state"} {
			if state := nestedStringField(raw, path); state != string(executionEvidenceStatePresent) && state != string(executionEvidenceStateMissing) && state != string(executionEvidenceStateNotApplicable) {
				failures = append(failures, fmt.Sprintf("%s has unsupported state %q", path, state))
			}
		}
	case evalBaselineSchemaVersion:
		if got := stringField(raw, "result_schema_version"); got != evalResultSchemaVersion {
			failures = append(failures, fmt.Sprintf("result_schema_version = %q, want %q", got, evalResultSchemaVersion))
		}
	case evalScorecardSchemaVersion:
		seen := map[string]bool{}
		for _, item := range arrayField(raw, "metrics") {
			if metric, ok := item.(map[string]any); ok {
				seen[stringField(metric, "name")] = true
			}
		}
		for _, definition := range evalScorecardMetricDefinitions() {
			if !seen[definition.name] {
				failures = append(failures, fmt.Sprintf("scorecard metric %s is missing", definition.name))
			}
		}
	}
	return failures
}

func requireEvidenceField(raw map[string]any, path, kind string) error {
	value, ok := nestedField(raw, path)
	if !ok {
		return fmt.Errorf("%s is missing", path)
	}
	switch kind {
	case "string":
		if text, ok := value.(string); !ok || strings.TrimSpace(text) == "" {
			return fmt.Errorf("%s must be a non-empty string", path)
		}
	case "array":
		if _, ok := value.([]any); !ok {
			return fmt.Errorf("%s must be an array", path)
		}
	case "object":
		if _, ok := value.(map[string]any); !ok {
			return fmt.Errorf("%s must be an object", path)
		}
	case "number":
		if _, ok := value.(float64); !ok {
			return fmt.Errorf("%s must be a number", path)
		}
	case "bool":
		if _, ok := value.(bool); !ok {
			return fmt.Errorf("%s must be a boolean", path)
		}
	}
	return nil
}

func nestedField(raw map[string]any, path string) (any, bool) {
	current := any(raw)
	for _, part := range strings.Split(path, ".") {
		object, ok := current.(map[string]any)
		if !ok {
			return nil, false
		}
		current, ok = object[part]
		if !ok {
			return nil, false
		}
	}
	return current, true
}

func stringField(raw map[string]any, field string) string {
	value, _ := raw[field].(string)
	return strings.TrimSpace(value)
}

func nestedStringField(raw map[string]any, path string) string {
	value, ok := nestedField(raw, path)
	if !ok {
		return ""
	}
	text, _ := value.(string)
	return strings.TrimSpace(text)
}

func arrayField(raw map[string]any, field string) []any {
	value, _ := raw[field].([]any)
	return value
}
