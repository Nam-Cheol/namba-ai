# SPEC-066 Evaluation Plan

## Deterministic Checks

- Renderer tests assert the generated contract snippets are present in every managed skill template and every role-card/custom-agent template.
- Create-flow tests assert user-generated custom-agent TOML includes the shared contract while preserving TOML escaping.
- Init tests assert a fresh scaffold contains the contract in representative skill, role-card, and custom-agent files.
- Regen tests assert existing scaffold regeneration writes the same contract across generated skill and agent surfaces.

## Manual Evidence Checks

- Run local-source `regen` in the current repository and inspect generated diffs.
- Create a fresh temporary repository with the local CLI and confirm contract snippets appear under `.agents/skills` and `.codex/agents`.
- Commit the temporary repository baseline, rerun local-source `regen`, and confirm instruction surfaces have no diff.

## Validation Commands

- `gofmt -l "cmd" "internal" "namba_test.go"`
- `go test ./...`
- `go vet ./...`
