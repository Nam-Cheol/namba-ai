# SPEC-019 Plan

1. Inspect the current `plan`/`fix` parsing path, bugfix SPEC scaffolding helper, and generated skill or doc sources that currently encode the old contract.
2. Add explicit `plan`/`fix` argument parsing that treats `--help` and `-h` as read-only help flows and rejects unsupported flag-only invocations safely.
3. Add `--command <run|plan>` to `namba fix`, keep plain `namba fix "..."` as the default direct-repair path, and move the existing bugfix SPEC scaffolding under `namba fix --command plan`.
4. Define and implement the direct `namba fix` runtime contract: require repo context plus issue description, inspect relevant Namba context, edit the current workspace, add targeted regression coverage, run configured validation, and finish with `namba sync`.
5. Keep `namba plan` focused on feature planning and keep `namba pr`/`namba land` as the GitHub handoff path instead of overloading `fix`.
6. Rewrite README bundles and generated workflow guidance so they explain why each user-facing skill exists, when to use it, what CLI mapping it represents, and what detailed options matter, beginning with `namba fix --command run|plan`.
7. Update `.agents/skills/*`, `internal/namba/templates.go`, `internal/namba/readme.go`, `.namba/codex/*`, and any generated workflow docs so the planning-versus-repair contract stays aligned everywhere.
8. Add regression tests for help parsing, no-write safety, `fix --command plan`, default `fix` run semantics, failure-mode messaging, and generated guidance text where stable assertions are appropriate.
9. Run validation commands plus `namba regen`/`namba sync` as needed so repo-managed artifacts reflect the new contract.
