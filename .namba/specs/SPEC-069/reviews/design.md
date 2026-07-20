# Design Review

- Status: clear
- Last Reviewed: 2026-07-14
- Reviewer: Codex design review
- Command Skill: `$namba-plan-design-review`
- Recommended Role: `namba-designer`

## Focus

- This is a CLI/workflow information-design change; visual UI, image assets, and motion are not required.

## Findings

- Operators need stable ordered dry-run data, requested/effective separation, explicit unobserved state, and actionable fallback/blocked recovery.
- Generated role prompts must make read-only and writer responsibilities visible rather than sharing generic write instructions.

## Decisions

- Acceptance requires structured dry-run/report fields, human text summary, explicit status vocabulary, and generated operator documentation contract tests.
- Generated documentation covers policy mappings, unsupported override, state meanings, planned execution, and model-unavailable recovery.

## Follow-ups

- Add snapshot/string tests for CLI and generated-document wording as the feature is implemented.

## Recommendation

- Clear for implementation; no visual-design deliverable is needed.
