# SPEC-012

## Problem

Codex review가 여전히 진행되지 않고 있으며, Codex 웹에서 pull review trigger와 Linear 연동을 설정한 상태라 전제했지만 실제 요구 설정이 모두 충족되었는지 다시 확인할 필요가 있음

## Goal

Apply the smallest safe fix that resolves the reported issue.

## Context

- Project: namba-ai
- Project type: existing
- Language: go
- Mode: tdd
- Work type: fix
- Verified external context:
  - OpenAI Codex GitHub docs say PR review requires Code review to be enabled for the repository, and Automatic reviews is a separate toggle for auto-review on PR open-for-review events.
  - OpenAI Codex GitHub docs also say explicit `@codex review` comments should trigger a review after the repository code-review feature is enabled.
  - OpenAI Codex Linear docs say Linear delegation requires Codex cloud tasks enabled for the workspace, Codex for Linear enabled in connector settings, and the user linked by mentioning `@Codex` in a Linear issue.
  - GitHub review and Linear delegation are documented as separate integrations, so a valid Linear connection alone does not prove GitHub PR review is configured correctly.