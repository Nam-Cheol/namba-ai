# SPEC-035

## Goal

Make the `namba init` onboarding wizard clearer for Codex access choices and add a safe project-root `namba` command for changing those access defaults after initialization.

## Context

- Project: namba-ai
- Project type: existing
- Language: go
- Mode: tdd
- Work type: plan
- Verified local context as of 2026-04-23:
  - `runInitWizard` already prompts `approval_policy` and `sandbox_mode`, but it does so as two raw low-level selects without a higher-level access preset or post-selection preview. (`internal/namba/namba.go`)
  - `parseInitArgs` already accepts `--approval-policy` and `--sandbox-mode`, and bootstrap tests assert those values land in `.namba/config/sections/system.yaml` and generated `.codex/config.toml`. (`internal/namba/namba.go`, `namba_test.go`, `internal/namba/init_wizard_test.go`)
  - Repo-owned Codex access defaults live in `.namba/config/sections/system.yaml`, while `.codex/config.toml` is generated from that state. (`internal/namba/templates.go`, `.namba/config/sections/system.yaml`, `.codex/config.toml`)
  - `runInit` always rewrites the full scaffold directly, including placeholder project docs and README bundles, so rerunning `namba init` in an existing repository is too broad to serve as a safe “permissions only” edit path. (`internal/namba/namba.go`)
  - `namba regen` already provides the safe managed-output path for regenerating repo-owned Codex assets from config without clobbering unrelated files. (`internal/namba/update_command.go`)
  - `namba init --help` currently advertises only a subset of supported flags, so the existing non-interactive access controls are under-documented. (`internal/namba/namba.go`)

## Problem

Namba has the raw configuration knobs for Codex access, but the user experience is incomplete in three ways:

1. The `init` wizard asks for `approval_policy` and `sandbox_mode` separately, which makes users infer the operational meaning of each combination on their own.
2. After initialization there is no safe, explicit `namba` command for adjusting repo-owned Codex access defaults from the project root. The practical workaround is manual YAML editing or rerunning `namba init`, both of which are too error-prone.
3. The repository already separates source-of-truth config from generated Codex assets, but there is no focused mutation path that updates the config, regenerates the managed outputs, and warns about session refresh requirements in one coherent flow.

## Desired Outcome

- `namba init` presents Codex access as a guided choice with clearer labels, tradeoff explanations, and a concise preview of the resulting `approval_policy` and `sandbox_mode`.
- A dedicated project-root command, `namba codex access`, lets users inspect and change Codex access defaults after initialization without hand-editing `.namba/config/sections/system.yaml`.
- `namba codex access` without mutation flags is an inspect path first: it prints the current effective access state in a terminal-friendly format, including the resolved preset label and the effective `approval_policy` / `sandbox_mode`, and it does not apply changes implicitly.
- Both bootstrap-time and post-init access changes resolve through one shared validation/normalization layer and keep `.namba/config/sections/system.yaml` as the source of truth.
- `namba init` and `namba codex access` share one user-facing access model: the same preset names, consequence statements, raw-value mapping, and preview semantics.
- Post-init access edits regenerate the affected repo-managed Codex assets and emit a session-refresh notice when instruction-surface files changed.
- Help text, docs, and tests make the first-run flow and the follow-up reconfiguration flow equally discoverable.

## Scope

- Add a shared Codex access configuration model that maps user-facing choices onto `approval_policy` and `sandbox_mode`.
- Enhance the interactive `namba init` flow so the Codex access step is more guided than today and surfaces the effective values more clearly.
- Add a new top-level command surface following the existing nested-command pattern:
  - `namba codex access`
  - `namba codex access --approval-policy <value> --sandbox-mode <value>`
  - `namba codex access --help`
- Define the zero-argument behavior for `namba codex access` as inspect-only across interactive and non-interactive contexts so the read path is deterministic and script-safe.
- Persist post-init access changes by updating `.namba/config/sections/system.yaml` and regenerating a narrow managed-output set centered on `.codex/config.toml` through the existing managed-output pipeline rather than the full init scaffold path.
- Update the minimum discoverability surfaces:
  - `namba init --help`
  - `namba codex access --help`
  - generated getting-started guidance
- Add explicit product/CLI semantics for:
  - shared preset labels and consequence statements across init and post-init flows
  - no-op mutation behavior
  - when session-refresh notice appears and when it is suppressed
  - invalid flag combinations and non-managed-repo errors
- Add regression coverage for parser/help behavior, interactive/default access selection, existing-repo reconfiguration, generated config outputs, and the guarantee that non-managed files are not clobbered.

## Non-Goals

- Do not turn this slice into a general-purpose editor for every `.namba/config/sections/*.yaml` file.
- Do not store user-specific auth, tokens, or machine-local Codex preferences in repo-managed config.
- Do not redesign `namba regen`, `namba project`, or runtime delegation semantics outside the access-setting surfaces touched by this change.
- Do not silently widen permissions during migration; risky combinations must stay explicit and reviewable.

## Design Constraints

- `.namba/config/sections/system.yaml` remains the source of truth; `.codex/config.toml` stays generated output.
- `namba codex access` is a first-class command contract, not an undocumented `init` side path or a thin YAML-editing helper.
- The new post-init edit path must be safe to run in an already initialized repo and must not overwrite user-authored project docs, README content, or other non-managed files that `runInit` writes during first bootstrap.
- `namba codex access` must fail clearly outside a Namba-managed repository instead of guessing or creating partial state.
- The effective access profile must be resolved and validated fully in memory before any write so invalid flag combinations or invalid managed config never leave source-of-truth config and generated output out of sync.
- Interactive flows must stay terminal-friendly and degrade cleanly to flag-driven non-interactive execution.
- If the effective access pair is unchanged, the command should avoid write/manifest churn and suppress session-refresh warnings in favor of a clear no-change confirmation.
- Implementation should follow the repository TDD contract: add failing tests for the new parser, mutation path, and no-clobber behavior before the production changes.
