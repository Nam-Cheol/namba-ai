# SPEC-010

## Goal

Add explicit PR handoff and merge commands so Namba can automate the repeatable review cycle without overloading `namba sync`.

## Context

- Project: namba-ai
- Project type: existing
- Language: go
- Mode: tdd
- Work type: plan
- `namba sync` should remain a local artifact refresh command.
- The repetitive review cycle is now: sync, validate, commit, push, open PR, confirm `@codex review`, wait for checks, merge, and update local `main`.
- This SPEC should introduce dedicated commands for that collaboration flow instead of expanding `sync` beyond its current responsibility.

## Scope

- Add `namba pr "<title>"` as the command that prepares the current branch for review.
- Add `namba land` as the command that finishes the review cycle after the PR is ready to merge.
- Use the configured PR base branch, PR language, and Codex review marker from `.namba/config/sections/git-strategy.yaml`.
- Use GitHub CLI as the automation surface for PR lookup, creation, checks, comments, and merge actions.

## Non-Goals

- Do not fold PR or merge automation into `namba sync`.
- Do not redesign `namba release` in this SPEC.
- Do not add broad Git hosting abstraction beyond the current GitHub-oriented workflow.
