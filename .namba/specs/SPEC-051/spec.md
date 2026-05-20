# SPEC-051

## Problem

`namba plan` and `namba fix --command plan` derive the dedicated SPEC branch name from the full planning description. When the description is long, the branch keeps a readable `spec/SPEC-XXX-` prefix but the generated slug after the SPEC id can become excessively long. In the worst case this blocks planning entirely because Git rejects or struggles with the oversized ref path.

## Goal

Cap only the generated slug segment that follows `SPEC-XXX` for planning branches, while preserving the branch prefix, SPEC id, and existing behavior for already-short descriptions.

## Scope

- Planning branch generation for `namba plan`.
- Bugfix planning branch generation for `namba fix --command plan`.
- Regression coverage for long English and Korean descriptions.

## Constraints

- Preserve `spec/SPEC-XXX-` exactly when branch-per-work planning is enabled.
- Truncate only the generated slug after the SPEC id.
- Choose a conservative built-in slug length that is comfortably below normal Git and GitHub practical limits without requiring a new user-facing setting.
- Keep short generated branch names unchanged.
- Preserve existing branch uniqueness, branch reuse, and collision handling behavior.

## Context

- Project: namba-ai
- Project type: existing
- Language: go
- Mode: tdd
- Work type: fix
