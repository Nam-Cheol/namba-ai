# SPEC-048 Contract

## Contract

Add a small deterministic harness eval pack that measures existing NambaAI
harness behavior using golden fixtures.

## Invariants

- Fixtures live under `internal/namba/testdata/evals/harness/`.
- Go tests are the primary eval owner.
- Python tests are used only for real lifecycle hook subprocess boundaries.
- Eval tests do not call external LLMs, GitHub, Codex CLI, network services, or
  user-specific local configuration.
- Runtime behavior is preserved unless a tiny testability adapter is strictly
  required.
- Hook guard logic is not copied into Go.
- Route fixtures verify existing command/request outcomes, not a new runtime
  harness enum.
- Evidence fixtures distinguish raw persisted schema validation from builder
  normalization behavior.
- The eval pack stays small, curated, and deterministic.

## Required Coverage

- Harness contract routing.
- Prompt refinement boundaries.
- Shell guardrail boundaries.
- Execution evidence manifest structure.
- PR review explicit opt-in behavior.

## Failure Diagnostics

Every fixture-backed failure must include the case name, input or command or
manifest identifier, expected value, actual value, and rationale.
