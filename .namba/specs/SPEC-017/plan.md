# SPEC-017 Plan

1. Add a Codex capability probe that distinguishes `exec` and `exec resume`
2. Refactor runtime argument construction to support direct flags, `-c` fallbacks, and explicit unsupported errors
3. Extend preflight so representability failures happen before execution starts
4. Gate `--parallel` before worktree fan-out using the same preflight contract
5. Add regression coverage for config fallbacks, unsupported resume profile, and parallel early-fail behavior
6. Run validation commands and sync artifacts with `namba sync`
