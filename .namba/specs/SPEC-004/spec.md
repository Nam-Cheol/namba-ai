# SPEC-004

## Goal

Prepare an OSS-facing README with direct installation instructions and enforce UTF-8 output encoding for CLI execution.

## Context

- Project: namba-ai
- Language: go
- Mode: tdd
- Audience: external open-source users
- Current gap: README is minimal, English-only, and does not show a direct install path

## Requirements

- Rewrite README for open-source usage with Korean instructions and emoji-based section headers.
- Provide a direct install command that users can copy immediately.
- Align the Go module path with the actual GitHub repository path so `go install` works.
- Force Windows console output encoding to UTF-8 for CLI execution.
- Document the UTF-8 behavior in README.