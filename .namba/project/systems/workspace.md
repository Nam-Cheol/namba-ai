# System: workspace

- Root: `.`
- Kind: go-service

## Purpose

- NambaAI is a practical guide for working with Codex without guessing the next step. It sets up the repository, helps you choose the right command, turns bigger work into reviewable plans, and refreshes docs and checklists after the work is done. Confidence: medium. Evidence: `README.md`.

## Entry Points And Interfaces

- `cmd/namba/main.go`: Go command entry point Confidence: high. Evidence: `cmd/namba/main.go`.

## Module Boundaries

- `.github` is a visible module boundary in this system. Confidence: medium. Evidence: `.github`.
- `assets` is a visible module boundary in this system. Confidence: medium. Evidence: `assets`.
- `cmd` is a visible module boundary in this system. Confidence: medium. Evidence: `cmd`.
- `docs` is a visible module boundary in this system. Confidence: medium. Evidence: `docs`.
- `internal` is a visible module boundary in this system. Confidence: medium. Evidence: `internal`.

## Data And State

- Generated project state is persisted under `.namba`, including manifests and project documents. Confidence: high. Evidence: `.namba/manifest.json`.

## Auth And Integrations

- Codex, GitHub workflow, or security-policy integrations are explicitly represented in the repository surface. Confidence: high. Evidence: `.codex/config.toml`, `.github/workflows/codex-review-request.yml`, `.github/workflows/ci.yml`, `SECURITY.md`.

## Deploy Runtime And Test Risks

- System-local regression coverage exists, but end-to-end drift across generated planning docs still needs command-level validation. Confidence: medium. Evidence: `cmd/namba/encoding_test.go`, `cmd/namba/main_test.go`, `codex_review_request_workflow_test.go`.
