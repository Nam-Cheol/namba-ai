# Release Notes Draft

Project: namba-ai
Project type: existing
Reference SPEC: SPEC-008
Generated: 2026-03-16T17:45:00+09:00

## Release Target

- Candidate version: `v0.1.1`
- Base semantic version: `v0.1.0`

## Highlights

- Added explicit workflow permission handling for Codex execution.
- Expanded the init wizard with project type selection, Java support, and improved interactive terminal controls.
- Added `namba fix "<description>"` for bugfix-oriented SPEC packages.
- Added `namba release` to create and optionally push release tags after validators pass.
- Added `namba update` plus structured parallel run sync and reporting.
- Synced README, AGENTS, repo skills, and generated project docs with the current workflow.

## Validation Status

- Validation commands were completed before release.
- CI remains configured to run tests, `go vet`, formatting checks, and secret scanning.

## Release Command

```text
namba release --version v0.1.1
git push origin main
git push origin v0.1.1
```

## Expected Assets

- `namba_Windows_x86_64.zip`
- `namba_Windows_arm64.zip`
- `namba_Linux_x86_64.tar.gz`
- `namba_Linux_arm64.tar.gz`
- `namba_macOS_x86_64.tar.gz`
- `namba_macOS_arm64.tar.gz`
