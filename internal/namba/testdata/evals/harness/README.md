# NambaAI Harness Eval Fixtures

This directory is a small golden eval pack for NambaAI harness quality boundaries. It is intentionally fixture-driven and deterministic: cases must not call LLMs, GitHub, Codex CLI, network services, wall-clock dependent services, or user-specific local configuration.

Run the eval pack with the normal project checks:

```sh
go test ./...
python3 -m unittest discover -s tests
```

## Files

- `route_cases.json` measures command-selection and harness sidecar outcomes for core harness planning, domain harness planning, direct artifact creation, ordinary planning, and bugfix SPEC planning.
- `mention_plugin_cases.json` measures Codex unified `@` mention and plugin/list-backed metadata ambiguity without making mention search the Namba command router or requiring plugin installation.
- `prompt_refinement_cases.json` measures the Codex lifecycle hook prompt-refinement gate through the real `.codex/hooks/namba_codex_guard.py` subprocess.
- `guardrail_cases.json` measures dangerous shell denials, safe commands, and approval-risk notes through the real hook subprocess.
- `evidence_manifest_cases.json` measures execution evidence manifest boundaries. Raw-schema cases validate fixture shape without builder defaults. Builder-normalization cases call the existing manifest builder and assert normalized state.
- `pr_review_cases.json` measures explicit Codex review opt-in behavior using existing PR parsing and review-comment helpers.

## Fixture Schemas

Every fixture item must include:

- `name`: stable case identifier shown in diagnostics.
- `rationale`: why the case exists and what regression it protects.
- One of `input`, `command`, or `manifest`, depending on fixture type.
- Expected result fields for the measured behavior.

Route cases use:

- `input`
- `expected_category`
- `expected_command`
- `expected_harness_request_kind_or_none`
- `expected_sidecar_persisted`
- optional `expected_delivery_mode`
- optional `expected_required_evidence`
- optional `expected_review_flags`

Mention/plugin cases use:

- `input`
- `mention_kinds`
- `expected_namba_routing`
- `expected_platform_readiness`
- `expected_requires_plugin_install`

Prompt refinement cases use:

- `input`
- `expected_refinement_required`
- optional `expected_language_behavior`

Guardrail cases use:

- `event_type`
- `command`
- `expected_deny`
- `expected_risk_note`
- optional `expected_reason_substring`

Evidence manifest cases use:

- `validation_path`: `raw_schema` or `builder_normalization`
- `manifest` for raw-schema cases
- `builder` for builder-normalization cases
- `expected_valid`
- `expected_missing_or_invalid_fields` when invalid
- optional expected section states and hook/artifact expectations

PR review cases use:

- `argv`
- `legacy_auto_codex_review`
- `existing_comments`
- `expected_review_requested`
- `expected_duplicate_review_comment`

## Adding Cases

Add the smallest case that protects a real boundary and keep diagnostics obvious. Prefer one new fixture row over a new test helper unless the same pattern appears in several files. When a failure is expected, include the exact missing or invalid field so the failure message points at the broken contract.

These evals differ from ordinary unit tests by grouping cross-surface contract examples in readable fixtures. Unit tests should still cover low-level behavior; eval fixtures should preserve curated user-visible boundaries and regression examples.

Do not add large corpora, generated reports, runtime logs, external model checks, GitHub API calls, network installs, or tests that write persistent state to the real repository.
