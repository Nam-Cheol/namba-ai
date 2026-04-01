# SPEC-019

## Problem

The current `plan` and `fix` surfaces collapse two different intents into one generic SPEC generator.

Local evidence in the current repository:

- `internal/namba/namba.go` routes both `runPlan` and `runFix` into the same `createSpecPackage` helper.
- `createSpecPackage` only rejects an empty arg list, so `namba plan --help` and `namba fix --help` are interpreted as descriptions instead of help requests.
- `.agents/skills/namba-fix/SKILL.md` and the generated templates in `internal/namba/templates.go` currently describe `$namba-fix` as a bug-fix SPEC package creator, even though the intended user value is harness-driven repair execution against a reported issue.
- A bugfix-planning path still needs to exist, but moving that path away from `fix` entirely weakens discoverability because users naturally expect the `fix` family to cover both planning and execution decisions for bug work.
- The current README bundles and generated workflow docs explain command mappings only shallowly. They do not explain why each user-facing skill exists, what workflow problem it solves, or what option surfaces such as `--command` mean in practice.

## Goal

Restore a clear `fix` command family where bugfix planning and direct repair are both explicit under `namba fix`, help and probe flows stay read-only, and README or generated docs teach the purpose and option surface of each user-facing Namba skill clearly enough that users do not have to infer behavior from source code.

## Context

- Project: namba-ai
- Project type: existing
- Language: go
- Mode: tdd
- Work type: plan
- Affected area: `internal/namba/namba.go`, `.agents/skills/namba-fix/SKILL.md`, `.agents/skills/namba-plan/SKILL.md`, `.agents/skills/namba/SKILL.md`, `internal/namba/templates.go`, `internal/namba/readme.go`, generated README and workflow docs

## Desired Outcome

- `namba plan "<description>"` remains the feature-planning path and creates the next feature-oriented `SPEC-XXX` package plus review scaffolds.
- `namba fix --command plan "<issue description>"` becomes the explicit authored bugfix-planning path and creates the next bugfix-oriented `SPEC-XXX` package with review artifacts.
- `namba fix "<issue description>"` defaults to direct repair behavior in the current workspace, and `namba fix --command run "<issue description>"` is the explicit equivalent form.
- `namba fix` no longer behaves like an ambiguous hidden alias for SPEC generation. Planning and execution are both available in the `fix` family, but the selected `--command` makes the behavior obvious.
- The direct-repair runtime contract is explicit: require repository context plus issue description, inspect relevant repo and Namba project/config context, implement the smallest safe fix, add targeted regression coverage, run validation commands from `.namba/config/sections/quality.yaml`, and finish with `namba sync`.
- `namba plan --help`, `namba plan -h`, `namba fix --help`, and `namba fix -h` print subcommand help and exit without writing into `.namba/specs/`.
- Flag-only probing or malformed invocations for `plan` and `fix` fail safely without mutating repo state.
- README bundles, generated workflow docs, and generated Codex guidance explain each user-facing repo-local Namba skill in terms of intent, when to use it, primary command mapping, and relevant option surface, including `namba fix --command ...`.
- Regression tests lock the parsing, no-write help behavior, `fix --command` contract, README or generated-doc guidance, and generated-surface consistency so future template changes cannot silently reintroduce ambiguity.

## Scope

- Add explicit subcommand argument parsing for `plan` and `fix`, including `--help`/`-h` handling and safe rejection of unsupported flag-only invocations.
- Add `--command <run|plan>` to `namba fix`.
- Keep `namba fix "<issue description>"` as the default direct-repair path, equivalent to `namba fix --command run "<issue description>"`.
- Move the existing bugfix SPEC scaffolding behavior under `namba fix --command plan`, preserving review artifacts and bugfix-oriented problem/goal/acceptance language.
- Preserve `namba plan` as the feature-planning path rather than turning it into a mixed feature/bugfix planning surface.
- Redefine `namba fix` direct repair as a harness-driven workflow that requires a repo root plus issue description, edits the current workspace directly, adds targeted regression coverage, runs configured validation commands, and finishes with `namba sync`.
- Keep GitHub handoff outside the `fix` command. Users still use `namba pr` and `namba land` after a successful repair flow.
- Rewrite README bundles and generated workflow docs so they explain the intent behind each user-facing repo-local skill, the main command mapping, and the detailed option surface for commands that need it, starting with `namba fix --command run|plan`.
- Update repo-managed skill files, generated templates, and generated docs so `$namba`, `$namba-plan`, `$namba-fix`, the review skills, and top-level workflow guidance all match the CLI contract exactly.
- Add regression coverage for CLI parsing, no-write help flows, `fix --command plan`, default `fix` run behavior, failure-mode messaging, and generated guidance text where stable assertions are appropriate.

## Non-Goals

- Do not redesign `namba run`, `namba sync`, `namba pr`, or `namba land`.
- Do not broaden this work into a general CLI framework rewrite outside the `plan` and `fix` command surface.
- Do not move bugfix planning into `namba plan`; this SPEC intentionally keeps bugfix planning discoverable inside the `fix` family.
- Do not turn `namba fix` into PR creation, merge automation, or a second `namba run` mode surface with unrelated execution flags.
- Do not add decorative README copy at the expense of concrete command, option, and workflow guidance.

## Design Constraints

- Keep `.namba/` as the source of truth for authored SPEC packages. After this change, reviewable bugfix planning must go through `namba fix --command plan`, not through plain `namba fix`.
- Never create a `SPEC-XXX` directory as a side effect of help discovery, malformed flag probing, or normal direct-repair execution.
- Keep the user-facing distinction between feature planning (`plan`), bugfix planning (`fix --command plan`), and direct repair (`fix` or `fix --command run`) obvious in code, tests, skills, help text, and generated docs.
- Keep `namba fix` aligned with the Namba repair posture: smallest safe fix, explicit regression coverage, configured validation, synced artifacts, and actionable failures when preconditions are not met.
- Treat `internal/namba/templates.go`, `internal/namba/readme.go`, and other renderer sources as authoritative for generated instruction surfaces; update sources first, then regenerate and sync outputs.
- Keep user-facing copy concise, example-led, and text-first so it works well in plain terminals and across generated multilingual docs.
- Preserve deterministic, testable behavior in local CI without relying on network access or interactive prompts.
