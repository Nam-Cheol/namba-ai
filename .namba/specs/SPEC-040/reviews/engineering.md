# Engineering Review

- Status: clear
- Last Reviewed: 2026-04-27
- Reviewer: Codex acting as `namba-planner`
- Command Skill: `$namba-plan-eng-review`
- Recommended Role: `namba-planner`

## Focus

- Lock architecture, sequencing, failure modes, trust boundaries, and validation strategy before execution starts.

## Findings

- The ownership model is correct: source edits belong in `internal/namba/templates.go`; generated `.agents/skills/*/SKILL.md` files should change through `namba regen`; project artifacts should refresh through `namba sync`.
- The typed harness metadata now matches the real implementation risk. Although the SPEC was requested through `$namba-harness`, `core_harness_change`, `modify_core`, and `touches_namba_core=true` are the right execution signals because this work changes Namba-managed templates and command-entry guidance.
- The trust boundary around upstream content is explicit. The plan adapts patterns from `ComposioHQ/awesome-codex-skills` but rejects wholesale copying, install flows, Composio dependencies, and third-party app automation.
- The helper-candidate criteria are implementation-ready. The `--help`, fixture/local testability, bounded output, no silent mutation, explicit network/auth assumptions, and no app/OAuth coupling requirements are specific enough to test.
- The validation strategy points at the right local surfaces: template rendering tests, update/regen tests, no `.codex/skills` or `$CODEX_HOME/skills` regression, review-loop checks if executable behavior is added, and final repo quality commands.

## Decisions

- Implement template wording before generated artifacts; do not hand-edit generated skill files except as part of reviewing `namba regen` output.
- Add regression tests for contractual strings only where wording is part of routing, safety, or evidence behavior. Avoid brittle tests for incidental prose.
- If helper scripts are implemented, keep them read-only by default and test them with fixtures or local dummy servers before any live GitHub or browser workflow is relied upon.
- Keep external CI providers out of automation scope. Report their status and URL only.

## Follow-ups

- Before coding, inventory exact render functions and tests to touch so behavior changes do not spread across unrelated generated docs.
- After `namba regen`, review source-template diffs and generated skill diffs separately.
- If new helper/reference files are added, ensure manifest ownership and `namba regen` or `namba sync` behavior are explicit and covered by tests.
- Confirm `namba pr` guidance does not imply networked GitHub inspection is mandatory for local-only validation paths where no PR exists yet; it should apply at handoff time.

## Recommendation

- Clear for implementation. The SPEC has enough architecture, sequencing, and validation detail to proceed without another planning revision.
