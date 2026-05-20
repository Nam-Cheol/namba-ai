# SPEC-054

## Problem

Newly initialized Namba projects can inherit PR and release workflow
templates that stay too generic. When `namba pr` or `namba release` is
used from those projects, the generated handoff text must summarize the
work that actually happened instead of leaving maintainers with empty or
placeholder-style prose.

## Goal

Improve the templates and output contracts created by `namba init` so new
projects generate:

- `namba pr` bodies with concrete completed work.
- `namba release` GitHub release notes with actual changes, evidence
  sources, and validation results.

The generated text should preserve the evidence needed for review and
release decisions, including SPEC IDs, PR numbers, short commit hashes,
validation evidence, and source references when those are available.

## Scope

- Update the Namba CLI init and scaffold template surfaces that are copied
  into newly initialized projects.
- Update the scaffolded `namba-pr` and `namba-release` skill or workflow
  output contracts that shape PR body and GitHub release note generation.
- Add fixture or snapshot coverage for newly scaffolded projects and the
  generated PR and release handoff artifacts.

## Out Of Scope

- Migrating, backfilling, or rewriting already initialized downstream
  projects.
- Adding automatic `@codex review` requests. Review requests remain
  explicit opt-in only.
- Changing GitHub publishing mechanics except where required to pass the
  richer generated body text through existing workflow paths.

## Constraints

- Generated artifact language must prefer an explicit user-requested
  language.
- If no explicit language is requested, generated artifact language must
  follow the language stored by `namba init` or project configuration.
- Empty, template-like, or generic placeholder prose is not acceptable for
  PR bodies or GitHub Release bodies.
- Evidence should remain traceable from the final artifact back to the
  source of truth, such as SPEC files, commits, PR metadata, validation
  output, and release notes inputs.

## Context

- Project: namba-ai
- Project type: existing
- Language: go
- Mode: tdd
- Work type: plan

## Planning Evidence

- Project product docs describe NambaAI as a workflow that refines vague
  requests into goals, scope, constraints, and acceptance before execution.
- Project tech docs identify Go as the implementation language and
  `go test ./...`, `gofmt -l "cmd" "internal" "namba_test.go"`, and
  `go vet ./...` as configured validation commands.
- System docs identify `cmd/namba/main.go` as the CLI entry point and
  `.namba` as the generated state surface.
- `internal/namba/templates.go` renders the scaffolded `namba-pr` and
  `namba-release` skill bodies copied into initialized projects.
- `internal/namba/pr_land_command.go` currently builds PR bodies through
  `buildPullRequestBody`, pointing mostly to summary and checklist files.
- `internal/namba/release.go` currently renders commit-derived release notes
  through `renderReleaseNotes`, including category sections and commit suffix
  evidence.

## Required Output Contract

- PR body contract: completed work, changed areas or files, evidence sources,
  validation result, and review readiness reference when available.
- Release notes contract: actual changes grouped for users, evidence sources
  such as SPEC IDs, PR numbers, commit hashes, source files, and validation
  result or validation artifact reference.
- Language contract: explicit user request wins; otherwise the project init or
  git strategy language setting wins.
- Review contract: `@codex review` must be absent unless the existing explicit
  opt-in review path requests it.
