# SPEC-063 Eval Plan

## Test Strategy

- Start with failing or missing assertions for the generated verification guide and navigation links.
- Add or update renderer tests near existing README/docs sync tests.
- Add a hook/output-contract validation case that starts from a detailed pre-hook answer and verifies the required Namba report frame still includes the important details.
- Keep tests deterministic and offline.

## Suggested Checks

- Generated `docs/verification-guide.md` exists after the sync renderer runs.
- README generated output links to the verification guide.
- Workflow-guide generated output links to the verification guide.
- Localized generated outputs do not contain broken verification-guide links.
- Verification-guide content includes anchors or sections for eval, report, quality script, `.namba` evidence, CI, and failure interpretation.
- Final report validation preserves sample details such as changed file path, validation command, artifact path, blocker, and next command.

## Validation Commands

1. Targeted Go tests for README/docs sync rendering.
2. Targeted Python or Go tests for hook/output-contract behavior.
3. `go test ./...`
4. `gofmt -l "cmd" "internal" "namba_test.go"`
5. `go vet ./...`
6. `namba sync`
7. Generated drift check with `git diff --exit-code` after the intended generated files are produced.
8. `scripts/quality.sh`

## Failure Evidence

If a command cannot run in the local environment, record:

- Exact command.
- Exit status or failing line.
- Relevant artifact path.
- Whether the failure is product behavior, test failure, missing tool, sandbox/network, or environment setup.
- Concrete next command or handoff.
