package namba

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

func renderEvalJSON(result evalRunResult) (string, error) {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", fmt.Errorf("render eval json: %w", err)
	}
	return string(append(data, '\n')), nil
}

func renderEvalMarkdown(result evalRunResult) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# Namba Eval Report\n\n")
	fmt.Fprintf(&b, "- suite: `%s`\n", result.Suite)
	fmt.Fprintf(&b, "- generated_at: `%s`\n", result.GeneratedAt)
	fmt.Fprintf(&b, "- scenarios: %d total, %d passed, %d failed\n", result.Summary.Total, result.Summary.Passed, result.Summary.Failed)
	if result.Baseline.Compared {
		status := "passed"
		if !result.Baseline.Passed {
			status = "failed"
		}
		fmt.Fprintf(&b, "- baseline: %s `%s`\n", status, result.Baseline.Path)
	}
	fmt.Fprintf(&b, "\n## Metrics\n\n")
	for _, metric := range result.Metrics {
		fmt.Fprintf(&b, "- `%s`: %d/%d (%.0f%%)\n", metric.Name, metric.Passed, metric.Total, metric.PassRate*100)
	}
	if len(result.Regressions) > 0 {
		fmt.Fprintf(&b, "\n## Baseline Regressions\n\n")
		for _, regression := range result.Regressions {
			fmt.Fprintf(&b, "- %s\n", regression)
		}
	}
	failed := make([]evalScenarioResult, 0)
	passed := make([]evalScenarioResult, 0)
	for _, scenario := range result.Scenarios {
		if scenario.Passed {
			passed = append(passed, scenario)
		} else {
			failed = append(failed, scenario)
		}
	}
	sortEvalScenarioResults(failed)
	sortEvalScenarioResults(passed)
	if len(failed) > 0 {
		fmt.Fprintf(&b, "\n## Failed Scenarios\n\n")
		for _, scenario := range failed {
			renderEvalScenarioMarkdown(&b, scenario)
		}
	}
	fmt.Fprintf(&b, "\n## Scenario Results\n\n")
	for _, scenario := range append(failed, passed...) {
		if scenario.Passed {
			fmt.Fprintf(&b, "- PASS `%s` (%s)\n", scenario.ID, scenario.Type)
			continue
		}
		fmt.Fprintf(&b, "- FAIL `%s` (%s)\n", scenario.ID, scenario.Type)
	}
	return b.String()
}

func renderEvalScorecardMarkdown(result evalRunResult) string {
	var b strings.Builder
	scorecard := result.Scorecard
	if scorecard == nil {
		fmt.Fprintf(&b, "# Namba Eval Scorecard\n\n")
		fmt.Fprintf(&b, "- status: `unavailable`\n")
		return b.String()
	}
	fmt.Fprintf(&b, "# Namba Eval Scorecard\n\n")
	fmt.Fprintf(&b, "- schema_version: `%s`\n", scorecard.SchemaVersion)
	fmt.Fprintf(&b, "- suite: `%s`\n", scorecard.Suite)
	fmt.Fprintf(&b, "- corpus_version: `%s`\n", scorecard.CorpusVersion)
	fmt.Fprintf(&b, "- generated_at: `%s`\n", scorecard.GeneratedAt)
	fmt.Fprintf(&b, "- scenarios: %d total, %d passed, %d failed\n", scorecard.Summary.Total, scorecard.Summary.Passed, scorecard.Summary.Failed)
	fmt.Fprintf(&b, "- regressions: %d\n", scorecard.Summary.RegressionCount)
	fmt.Fprintf(&b, "- schema_validation: `%s`\n", scorecard.SchemaValidation.Status)
	if scorecard.Baseline.Compared {
		status := "passed"
		if !scorecard.Baseline.Passed {
			status = "failed"
		}
		fmt.Fprintf(&b, "- baseline: %s `%s`\n", status, scorecard.Baseline.Path)
	}
	fmt.Fprintf(&b, "\n## 1.0 Metrics\n\n")
	for _, metric := range scorecard.Metrics {
		fmt.Fprintf(&b, "- `%s`: %d/%d (%.0f%%)\n", metric.Name, metric.Passed, metric.Total, metric.PassRate*100)
	}
	if len(scorecard.RequiredCoverage) > 0 {
		fmt.Fprintf(&b, "\n## Required Coverage\n\n")
		for _, bucket := range scorecard.RequiredCoverage {
			fmt.Fprintf(&b, "- `%s`\n", bucket)
		}
	}
	if len(scorecard.Regressions) > 0 {
		fmt.Fprintf(&b, "\n## Baseline Regressions\n\n")
		for _, regression := range scorecard.Regressions {
			fmt.Fprintf(&b, "- %s\n", regression)
		}
	}
	return b.String()
}

func renderEvalScenarioMarkdown(b *strings.Builder, scenario evalScenarioResult) {
	fmt.Fprintf(b, "### `%s`\n\n", scenario.ID)
	fmt.Fprintf(b, "- type: `%s`\n", scenario.Type)
	if scenario.Input != "" {
		fmt.Fprintf(b, "- input: %s\n", scenario.Input)
	}
	fmt.Fprintf(b, "- rationale: %s\n", scenario.Rationale)
	fmt.Fprintf(b, "- failures:\n")
	for _, failure := range scenario.Failures {
		fmt.Fprintf(b, "  - %s\n", failure)
	}
	fmt.Fprintf(b, "- expected: `%s`\n", compactEvalJSON(scenario.Expected))
	fmt.Fprintf(b, "- actual: `%s`\n\n", compactEvalJSON(scenario.Actual))
}

func compactEvalJSON(value any) string {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Sprint(value)
	}
	return string(data)
}

func sortEvalScenarioResults(results []evalScenarioResult) {
	sort.Slice(results, func(i, j int) bool {
		return results[i].ID < results[j].ID
	})
}
