# Engineering Review

- Status: needs-revision
- Last Reviewed: 2026-05-20
- Reviewer: namba-planner
- Command Skill: `$namba-plan-eng-review`
- Recommended Role: `namba-planner`

## Focus

- Lock architecture, sequencing, failure modes, trust boundaries, and validation strategy before execution starts.

## Findings

- The biggest architecture ambiguity is config ownership. `spec.md` says
  repo-managed Codex config must not own sandbox or approval choices, but the
  current product surface still treats those as repo-owned defaults through
  `.codex/config.toml`, `.namba/config/sections/system.yaml`, and
  `namba codex access` (`internal/namba/codex_access.go`,
  `internal/namba/codex_access_command_test.go`, `.codex/config.toml`). This
  must be resolved before implementation or the team will either violate the
  SPEC or silently break an existing CLI contract.
- Duplicate-hook prevention must be designed at the interactive Codex boundary,
  not in `internal/namba/hook_runtime.go`. Plugin-bundled hooks and
  `.codex/hooks.json` both affect Codex session hooks, while
  `.namba/hooks.toml` governs `namba run` evidence only. If dedupe is attempted
  only in the Namba runner runtime, interactive duplicate executions will still
  happen and the trust boundary will remain wrong.
- `PreToolUse` rewrite support needs an explicit precedence rule before coding.
  The current Python guard only denies dangerous Bash commands and never emits
  `updatedInput` (`.codex/hooks/namba_codex_guard.py`). For Codex 0.131,
  implementation must state that block/deny decisions win over rewrite, and
  rewrites apply only to already-allowed inputs. Without that sequencing,
  command blocking can be weakened by an over-eager rewrite path.
- Validation scope is still too implicit for 0.131 compatibility. Current tests
  cover only a thin subset of interactive hook behavior
  (`internal/namba/hook_guard_test.go`, portions of
  `internal/namba/templates_test.go`), and the POSIX launcher currently emits
  `"Unknown"` on wrapper failure rather than preserving the original event
  (`.codex/hooks/namba_codex_guard.sh`). The SPEC should require a shared
  six-event payload corpus plus explicit launcher-failure assertions before docs
  and README work, otherwise compatibility regressions will hide behind partial
  green tests.

## Decisions

- Keep the trust boundary split explicit: `.codex/hooks.json` and the launcher
  scripts own interactive guardrails, while `.namba/hooks.toml` and
  `internal/namba/hook_runtime.go` remain the `namba run` evidence boundary.
- Treat repo-owned `approval_policy` / `sandbox_mode` behavior as a prerequisite
  product decision for this SPEC. Either preserve that contract as the allowed
  repo-safe baseline, or remove it deliberately with matching CLI, template,
  doc, and test changes in the same slice.
- Require implementation sequencing to start with generator/runtime contracts
  and tests, then launcher hardening, then docs. Documentation should trail
  verified behavior here, not define it.

## Follow-ups

- Clarify in `spec.md` or implementation notes whether `namba codex access`
  survives unchanged, is narrowed to inspect-only, or is removed from
  repo-managed Codex config ownership.
- Add an explicit `PreToolUse` contract: malformed payload -> valid JSON
  continue/failure response, deny beats rewrite, rewrite never bypasses command
  blocking, and tests cover allow/deny/rewrite for quoted and multiline input.
- Add acceptance language that duplicate-hook prevention is verified at the
  generated `.codex/hooks.json` plus launcher/guard level, not inferred from
  Namba runner hooks.
- Add launcher acceptance that both POSIX and PowerShell failure paths preserve
  the originating hook event when possible and never fall back to shell traces
  as the primary user-visible output.

## Recommendation

- Revise before implementation. The SPEC is close, but config ownership,
  interactive-hook dedupe placement, and `updatedInput` sequencing need to be
  pinned down first for a safe Codex 0.131-compatible implementation.
