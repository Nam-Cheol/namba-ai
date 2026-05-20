package namba

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

const (
	evalSchemaVersion         = "namba-eval-scenarios/v1"
	evalResultSchemaVersion   = "namba-eval-results/v1"
	evalBaselineSchemaVersion = "namba-eval-baseline/v1"
	defaultEvalSuite          = "harness"
	defaultEvalFormat         = "markdown"
	defaultHarnessEvalFixture = "internal/namba/testdata/evals/harness/scenarios.json"
	defaultHarnessEvalBase    = "internal/namba/testdata/evals/harness/baseline.json"
)

type evalExitError struct {
	code int
	err  error
}

func (e evalExitError) Error() string {
	if e.err == nil {
		return ""
	}
	return e.err.Error()
}

func (e evalExitError) Unwrap() error {
	return e.err
}

// ExitCode returns the process exit code that should be used for command errors.
func ExitCode(err error) int {
	if err == nil {
		return 0
	}
	var exitErr evalExitError
	if errors.As(err, &exitErr) && exitErr.code != 0 {
		return exitErr.code
	}
	return 1
}

func commandExitError(code int, err error) error {
	return evalExitError{code: code, err: err}
}

type evalCorpus struct {
	SchemaVersion string         `json:"schema_version"`
	Suite         string         `json:"suite"`
	CorpusVersion string         `json:"corpus_version"`
	Scenarios     []evalScenario `json:"scenarios"`
}

type evalScenario struct {
	ID        string          `json:"id"`
	Type      string          `json:"type"`
	Tags      []string        `json:"tags"`
	Input     string          `json:"input,omitempty"`
	Expected  map[string]any  `json:"expected"`
	Fixture   json.RawMessage `json:"fixture,omitempty"`
	Rationale string          `json:"rationale"`
}

type evalBaseline struct {
	SchemaVersion        string            `json:"schema_version"`
	ResultSchemaVersion  string            `json:"result_schema_version"`
	Suite                string            `json:"suite"`
	CorpusVersion        string            `json:"corpus_version"`
	ScenarioFingerprints map[string]string `json:"scenario_fingerprints"`
	MetricPassCounts     map[string]int    `json:"metric_pass_counts"`
	RequiredCoverage     []string          `json:"required_coverage"`
}

type evalRunResult struct {
	SchemaVersion string               `json:"schema_version"`
	Suite         string               `json:"suite"`
	GeneratedAt   string               `json:"generated_at"`
	Summary       evalSummary          `json:"summary"`
	Metrics       []evalMetric         `json:"metrics"`
	Scenarios     []evalScenarioResult `json:"scenarios"`
	Baseline      evalBaselineResult   `json:"baseline"`
	Regressions   []string             `json:"regressions"`
}

type evalSummary struct {
	Total           int `json:"total"`
	Passed          int `json:"passed"`
	Failed          int `json:"failed"`
	RegressionCount int `json:"regression_count"`
}

type evalMetric struct {
	Name     string  `json:"name"`
	Passed   int     `json:"passed"`
	Total    int     `json:"total"`
	PassRate float64 `json:"pass_rate"`
}

type evalScenarioResult struct {
	ID          string         `json:"id"`
	Type        string         `json:"type"`
	Tags        []string       `json:"tags"`
	Passed      bool           `json:"passed"`
	Input       string         `json:"input,omitempty"`
	Expected    map[string]any `json:"expected"`
	Actual      map[string]any `json:"actual"`
	Failures    []string       `json:"failures"`
	Rationale   string         `json:"rationale"`
	Fingerprint string         `json:"fingerprint"`
}

type evalBaselineResult struct {
	Compared      bool   `json:"compared"`
	Path          string `json:"path,omitempty"`
	Passed        bool   `json:"passed"`
	CorpusVersion string `json:"corpus_version,omitempty"`
}

func normalizeEvalStringSlice(value any) []string {
	switch typed := value.(type) {
	case []string:
		return append([]string(nil), typed...)
	case []any:
		out := make([]string, 0, len(typed))
		for _, item := range typed {
			out = append(out, fmt.Sprint(item))
		}
		return out
	default:
		return nil
	}
}

func evalString(value any) string {
	return strings.TrimSpace(fmt.Sprint(value))
}
