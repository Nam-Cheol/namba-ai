# Acceptance

- [ ] `namba plan "<description>"` still creates the next sequential feature-oriented `SPEC-XXX` package under `.namba/specs/` and seeds the expected review artifacts.
- [ ] `namba fix --command plan "<issue description>"` creates the next sequential bugfix-oriented `SPEC-XXX` package under `.namba/specs/` with the expected review artifacts and bugfix-specific scaffolding.
- [ ] `namba fix "<issue description>"` defaults to direct repair behavior in the current workspace and is behaviorally equivalent to `namba fix --command run "<issue description>"`.
- [ ] `namba plan --help` and `namba plan -h` print subcommand-specific help and exit without creating or mutating any `SPEC-XXX` package.
- [ ] `namba fix --help` and `namba fix -h` print subcommand-specific help, including the available `--command` values, and exit without creating or mutating any `SPEC-XXX` package.
- [ ] Unsupported or flag-only `plan`/`fix` invocations fail safely without writing into `.namba/specs/`.
- [ ] Direct `namba fix` requires repository context plus an issue description and does not create or mutate `.namba/specs/<SPEC>` as part of normal execution.
- [ ] In an interactive Codex session, direct `namba fix` follows Codex-native in-session execution semantics in the current workspace instead of recursively creating a SPEC or routing through `namba run`.
- [ ] Direct `namba fix` inspects relevant project/config context, implements the smallest safe fix, adds targeted regression coverage, runs configured validation commands from `.namba/config/sections/quality.yaml`, and finishes with `namba sync` on the same workspace path.
- [ ] When repo preconditions or validation fail, `namba fix` exits with actionable error text and no implicit SPEC creation.
- [ ] README bundles, generated workflow docs, and generated Codex guidance explain each user-facing repo-local Namba skill in terms of intent, main command mapping, and detailed options where relevant, including `namba fix --command run|plan`.
- [ ] Existing feature-planning behavior of `namba plan` is preserved while accidental SPEC creation via help probing is removed.
- [ ] Validation commands pass.
- [ ] Tests covering CLI parsing, no-write help behavior, `fix --command` semantics, failure-mode messaging, and generated guidance text are present.
