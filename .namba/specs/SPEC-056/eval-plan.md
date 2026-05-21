# SPEC-056 Eval Plan

## Goal

Prove that `namba eval` can detect harness-quality regressions deterministically and explain them to both humans and CI.

## Metric Matrix

| Metric | What It Measures | Primary Sources |
| --- | --- | --- |
| `route_selection` | Correct command or route class for core, domain, direct, ordinary, and planned-fix work | harness route scenarios |
| `delivery_mode` | `spec` versus `direct` delivery decisions | harness request scenarios |
| `artifact_targets` | expected target set such as `workflow`, `validator`, `eval-pack`, `docs`, `skill`, `agent` | harness request metadata |
| `required_evidence` | expected evidence pack such as `contract`, `baseline`, `eval-plan`, `harness-map` | harness request metadata and evidence scenarios |
| `required_reviews` | expected review tracks such as product, engineering, design | harness request metadata and readiness scenarios |
| `clarification` | whether vague requests stop before SPEC creation | prompt-refinement scenarios |
| `spec_fields` | whether required SPEC files and content fields are complete | synthetic temporary SPEC scenarios |
| `execution_readiness` | whether missing evidence or review state is surfaced before execution | readiness and evidence scenarios |
| `guardrail` | whether unsafe commands are blocked or risk-noted | guardrail scenarios |
| `pr_review` | whether Codex review request remains explicit opt-in | PR review scenarios |

## Minimum Corpus

V1 target: at least 24 scenarios.

Required buckets:

- direct artifact generation: 3
- domain feature change: 3
- core runtime or harness change: 4
- security-sensitive change: 3
- release-related change: 2
- docs-only change: 2
- ambiguous request requiring clarification: 3
- unsafe or blocked command scenario: 2
- missing evidence scenario: 2
- review-required scenario: 2

Existing 39 fixture rows can be normalized into this corpus rather than discarded.

## Regression Cases

The suite must fail clearly when:

- a core harness request routes as domain or direct work
- a direct artifact request persists a SPEC sidecar
- a vague request no longer requires clarification
- missing `contract.md`, `baseline.md`, `eval-plan.md`, or required `harness-map.md` is not reported
- review-required work lacks product, engineering, or design review expectations
- `namba pr` requests Codex review without `--review`
- unsafe commands are no longer denied or risk-noted
- JSON output omits scenario failures or baseline regression details

## Output Checks

JSON tests should parse the command output and assert stable fields.

Markdown tests should assert that failed scenarios include:

- scenario id
- input
- expected field
- actual field
- rationale
- baseline regression note when applicable

## CI Command

```sh
go run ./cmd/namba eval --suite harness --format json --baseline internal/namba/testdata/evals/harness/baseline.json --fail-on-regression
```

The CI command should run after `go test ./...` or as a dedicated step that still fails the workflow on regression.
