#!/usr/bin/env bash
set -euo pipefail

coverage_threshold="${COVERAGE_THRESHOLD:-73.0}"
artifact_dir="${QUALITY_ARTIFACT_DIR:-${TMPDIR:-/tmp}/namba-ai-quality}"
coverage_out="${COVERAGE_OUT:-${artifact_dir}/coverage.out}"
coverage_report="${COVERAGE_REPORT:-${artifact_dir}/coverage.txt}"
eval_scorecard="${EVAL_SCORECARD_OUT:-${artifact_dir}/eval-scorecard.json}"
eval_summary="${EVAL_SUMMARY_OUT:-${artifact_dir}/eval-summary.md}"
report_json="${REPORT_JSON_OUT:-${artifact_dir}/report.json}"
schema_validation="${SCHEMA_VALIDATION_OUT:-${artifact_dir}/schema-validation.txt}"
staticcheck_install="${STATICCHECK_INSTALL:-go install honnef.co/go/tools/cmd/staticcheck@v0.7.0}"
govulncheck_install="${GOVULNCHECK_INSTALL:-go install golang.org/x/vuln/cmd/govulncheck@v1.1.4}"

go_bin="$(go env GOPATH)/bin"
if [[ ":$PATH:" != *":$go_bin:"* ]]; then
  export PATH="$go_bin:$PATH"
fi

mkdir -p "$artifact_dir"

run_step() {
  printf '\n==> %s\n' "$1"
  shift
  "$@"
}

require_tool() {
  local tool="$1"
  local install_cmd="$2"

  if ! command -v "$tool" >/dev/null 2>&1; then
    printf 'Missing required tool: %s\n' "$tool" >&2
    printf 'Install it with:\n  %s\n' "$install_cmd" >&2
    exit 127
  fi
}

check_gofmt() {
  local out
  out="$(git ls-files -z '*.go' | xargs -0 gofmt -l)"
  if [[ -n "$out" ]]; then
    printf 'Go files need gofmt:\n%s\n' "$out" >&2
    return 1
  fi
}

check_coverage_threshold() {
  local report="$1"
  local measured

  measured="$(awk '/^total:/ {gsub(/%/, "", $3); print $3}' "$report")"
  if [[ -z "$measured" ]]; then
    printf 'Could not find total coverage in %s\n' "$report" >&2
    return 1
  fi

  printf 'Aggregate Go coverage: %s%% (threshold: %s%%)\n' "$measured" "$coverage_threshold"
  awk -v measured="$measured" -v threshold="$coverage_threshold" 'BEGIN { exit measured + 0 < threshold + 0 ? 1 : 0 }'
}

run_step "Python unittest discovery" python3 -m unittest discover -s tests

run_step "gofmt check" check_gofmt
run_step "Go vet" go vet ./...
run_step "Go unit tests" go test ./...
run_step "Harness eval regression" go run ./cmd/namba eval --suite harness --format json --baseline internal/namba/testdata/evals/harness/baseline.json --fail-on-regression --scorecard-out "$eval_scorecard" --summary-out "$eval_summary"
run_step "Evidence schema validation" sh -c 'go test ./internal/namba -run "TestEvidenceSchemaFilesExistAndMatchContracts|TestEvidenceContractValidation|TestReportJSONOutputValidatesAgainstSchemaContract|TestEvalScorecardUsesFixedHarnessQualityMetrics" -count=1 | tee "$1"' sh "$schema_validation"
run_step "Report JSON artifact" sh -c 'go run ./cmd/namba report --format json > "$1"' sh "$report_json"
run_step "Go race tests" go test -race ./...

require_tool "staticcheck" "$staticcheck_install"
run_step "staticcheck" staticcheck ./...

require_tool "govulncheck" "$govulncheck_install"
run_step "govulncheck" govulncheck ./...

run_step "Go coverage profile" go test ./... -coverprofile="$coverage_out"
run_step "Go coverage report" sh -c 'go tool cover -func="$1" | tee "$2"' sh "$coverage_out" "$coverage_report"
run_step "Go coverage threshold" check_coverage_threshold "$coverage_report"

printf '\nQuality artifacts:\n'
printf -- '- %s\n' "$coverage_out" "$coverage_report" "$eval_scorecard" "$eval_summary" "$report_json" "$schema_validation"
printf '\nQuality checks passed.\n'
