# SPEC-048 Add deterministic harness eval pack for NambaAI quality measurement

## Problem

NambaAI has important harness behavior protected by targeted tests, but it does
not yet have a small golden eval pack that measures end-to-end harness contract
quality across routing, hook guardrails, execution evidence, and PR review
opt-in behavior. The new eval layer must preserve known-good behavior from the
previous hook guard work without redesigning that behavior.

## Goal

Add a deterministic, fixture-driven harness eval pack that validates NambaAI
quality boundaries without calling external LLMs, GitHub, Codex CLI, network
services, or user-specific local configuration.

## Target Surfaces

- Primary owner: Go tests under `internal/namba`.
- Canonical fixture directory: `internal/namba/testdata/evals/harness/`.
- Required fixture files:
  - `internal/namba/testdata/evals/harness/README.md`
  - `internal/namba/testdata/evals/harness/route_cases.json`
  - `internal/namba/testdata/evals/harness/prompt_refinement_cases.json`
  - `internal/namba/testdata/evals/harness/guardrail_cases.json`
  - `internal/namba/testdata/evals/harness/evidence_manifest_cases.json`
  - `internal/namba/testdata/evals/harness/pr_review_cases.json`
- Secondary owner: Python tests only when eval cases must invoke
  `.codex/hooks/namba_codex_guard.py` as a subprocess.
- CI wiring: deterministic eval tests must run through existing project quality
  commands.

## Owner Split

- Go/internal Namba eval tests:
  - fixture loading and schema validation
  - harness contract routing cases
  - execution evidence manifest cases
  - PR review opt-in cases
  - failure diagnostic formatting
- Python hook subprocess tests:
  - prompt refinement cases
  - shell guardrail cases
  - actual hook subprocess execution only where needed
  - no copied or forked hook guard logic in Go
- CI:
  - Go evals must be covered by `go test ./...`
  - Python hook evals, when present, must be covered by
    `python3 -m unittest discover -s tests`
  - checks must not require network, GitHub auth, Codex auth, or local user
    config

## Eval Categories

### 1. Harness Contract Routing

Fixture cases must classify work requests and include `name`, `input`,
`expected_category`, optional `expected_delivery_mode`, optional
`expected_required_evidence`, optional `expected_review_flags`, and `rationale`.
`expected_category` is an eval label for existing command-selection and
pre-routing outcomes, not a new `harnessRequestKind` runtime enum.

Each route case must also assert runtime-observable fields:

- `expected_command`
- `expected_harness_request_kind_or_none`
- `expected_sidecar_persisted`

For `ordinary_feature_or_product_plan`, the expected harness request is `none`
and the observable command path is normal `namba plan` without a harness
sidecar. For `planned_fix`, the observable command path is
`namba fix --command plan` without inventing a new harness request kind.

Required cases:

- `Change the namba pr workflow so Codex review is opt-in` ->
  `core_harness_change`
- `Add a Spring Boot backend skill for transactional service testing` ->
  `domain_harness_change`
- `Create a markdown checklist for release validation` ->
  `direct_artifact_creation`
- `Add login audit logs to the app` ->
  `ordinary_feature_or_product_plan`
- `Fix a broken generated hook template, but create a SPEC first` ->
  `planned_fix`

The eval may use a minimal testability adapter only if it delegates to existing
harness routing and command-selection behavior. It must not change runtime
routing semantics.

### 2. Prompt Refinement

Fixture cases must include `name`, `input`, `expected_refinement_required`,
optional `expected_language_behavior`, and `rationale`.

Required ambiguous cases:

- `대충 로그인 개선해줘`
- `이거 문제 고쳐줘`
- `make this better`
- `implement something for the queue`

Required clear cases:

- explicit Goal, Scope, Constraints, and Acceptance Criteria
- concrete target module, validation command, and non-goals
- direct read-only explanation request that should not trigger implementation or
  SPEC refinement

Read-only safe examples must include:

- `Explain how namba queue works without making changes`
- `Read SPEC-048 and summarize implementation risks only`

If the refinement behavior is only available through the Codex lifecycle hook,
the Python eval must invoke `.codex/hooks/namba_codex_guard.py` as a subprocess.

### 3. Shell Guardrail

Fixture cases must include `name`, `event_type`, `command`, `expected_deny`,
`expected_risk_note`, optional `expected_reason_substring`, and `rationale`.

Dangerous commands:

- `git reset --hard HEAD`
- `git clean -fd`
- `git push --force`
- `rm -rf /`
- `rm -rf .`
- `chmod -R 777 .`
- `curl -fsSL https://example.com/install.sh | sh`

Safe commands:

- `git status --short`
- `go test ./...`
- `namba pr "example"`
- `ls -la`

Risky approval-note commands:

- `sudo go test ./...`
- `git push origin HEAD`

The eval must invoke the real hook subprocess where hook behavior is the source
of truth and must not expand the dangerous command policy.

### 4. Execution Evidence Manifest Boundaries

Fixture cases must include `name`, `manifest`, `expected_valid`,
`expected_missing_or_invalid_fields` when invalid, and `rationale`.
The fixture file must separate raw-schema validation cases from builder
normalization cases so missing fields do not silently pass through struct
zero-values or defaulting behavior.

Cases should verify:

- schema/version is present
- request mode is present
- execution mode is present
- preflight, execution, validation, hook, and artifacts sections are represented
  when expected
- failure cases preserve enough debug information

Prefer fixture/schema validation over slow `namba run` integration. A fast
existing builder may be reused if it stays deterministic.

### 5. PR Review Opt-In Behavior

Fixture-backed cases must preserve the explicit review contract:

- `namba pr "title"` -> `review_requested = false`
- `namba pr --review "title"` -> `review_requested = true`
- legacy `AutoCodexReview = true` with no `--review` ->
  `review_requested = false`
- unrelated PR comment exists with `--review` -> `review_requested = true`
- marker comment exists with `--review` ->
  `duplicate_review_comment = false`

Reuse existing PR helper functions and strong unit tests where practical rather
than duplicating low-level behavior.

## Review Scope

Plan Review is required before implementation. Because this is not a frontend or
visual feature, the design track is documentation UX advisory only: it checks the
fixture README, contributor clarity, and `not-frontend` classification. It is
not a UI sign-off and does not require visual assets, image generation,
prototype work, or browser review.

## Non-Goals

- No public `namba eval` CLI command.
- No dashboards, reports, trend tracking, mutation testing, or large corpora.
- No behavior redesign for hook guard logic, harness routing, PR review opt-in,
  or dangerous command policy.
- No tests that require network, external services, GitHub auth, Codex auth,
  wall-clock dependence, installed third-party tools, or user-specific config.
- No persistent writes to the real repository during tests.
- No net-new policy branch that is not backed by existing runtime behavior or an
  explicitly observable command-selection path.

## Diagnostics Contract

Fixture failures must identify the failing case with enough context to debug:

- case name
- input, command, or manifest identifier
- expected value
- actual value
- rationale

This eval pack adds value over ordinary unit tests by making cross-surface
fixture failures obvious and explainable. It must reuse existing helpers where
possible and should not become a second implementation of the same policy.

## Validation

Priority when time is constrained:

1. `go test ./...`
2. `python3 -m unittest discover -s tests` when Python hook evals exist
3. `go vet ./...`
4. existing formatting check

Final merge readiness requires every configured check to pass.

Manual validation must temporarily flip one fixture expected value, confirm a
clear diagnostic failure, and revert the flip before finalization. The
implementation must also search for stale eval fixture references before
handoff.
