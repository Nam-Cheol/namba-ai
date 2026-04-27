# SPEC-040 Plan

1. Refresh and lock context.
   - Use the upstream audit in `baseline.md` as pattern evidence.
   - Reconfirm local ownership: Namba skills are generated from `internal/namba/templates.go`; generated skill files are outputs.
   - Inventory the current command-entry skills that will be affected: `$namba-review-resolve`, `$namba-pr`, `$namba-harness`, `$namba-run`, `namba-workflow-execution`, `$namba-create`, and `$namba`.

2. Define the adaptation contract.
   - Use `contract.md` and `harness-map.md` to separate adopted, adapted, and rejected upstream patterns.
   - Keep the implementation limited to existing Namba-managed skills and workflow docs.
   - Treat helper scripts as candidates first; implement only if they pass the deterministic-helper criteria.

3. Tighten skill trigger metadata.
   - Review `name` and `description` frontmatter for the affected skills.
   - Ensure descriptions state both explicit command triggers and natural-language equivalents where useful.
   - Add regression assertions for stable trigger text only where the wording is part of the contract.

4. Apply progressive-disclosure discipline.
   - Keep each command-entry `SKILL.md` focused on trigger and execution behavior.
   - If a recipe becomes long, add an existing-skill reference file or Namba-owned helper path instead of creating a new standalone skill.
   - Add tests or manifest expectations for any new managed reference/helper output.

5. Strengthen PR and review-resolution loops.
   - Update `$namba-review-resolve` guidance so each meaningful thread records thread identity, outcome, concrete fix or answer, validation, CI/check evidence when relevant, and resolution state.
   - Update `namba pr` guidance so PR handoff includes sync, configured validation, current PR check status, GitHub Actions failure snippets when checks fail, and exactly one configured Codex review marker.
   - Keep external checks scoped to status and details URL only.

6. Define deterministic helper-script candidates.
   - Candidate A: fetch unresolved PR review threads through GitHub GraphQL and emit stable JSON with thread ids, paths, comments, authors, and resolved/outdated state.
   - Candidate B: inspect current-branch PR checks, detect failing GitHub Actions runs, fetch run/job logs, and emit bounded failure snippets.
   - Candidate C: manage local frontend server lifecycle, run Playwright checks, capture screenshots, rendered DOM facts, and console errors.
   - Implement candidates only when they can be tested with fixtures or local dummy servers and invoked as black boxes with `--help`.

7. Add batch-migration guidance.
   - Require a target inventory, mechanical-versus-behavioral classification, template-first edits, generated diff review, and validation.
   - Keep large refactors split when trigger metadata, behavior, helper plumbing, and docs would otherwise blur together.

8. Improve harness/MCP quality criteria.
   - Add workflow-first criteria: complete task workflows over raw endpoint wrappers.
   - Add context-budget criteria: concise default outputs, detailed mode only when needed, and truncation/pagination expectations.
   - Add actionable-error criteria: failures explain the next recoverable step.
   - Add evaluation criteria: independent, read-only, realistic, verifiable, stable scenarios with expected answers or observable evidence.

9. Regenerate and validate.
   - Run `namba regen` after template changes.
   - Run `namba sync` after implementation and generated artifact review.
   - Run configured quality commands:
     - `gofmt -l "cmd" "internal" "namba_test.go"`
     - `GOCACHE=/tmp/namba-go-cache go test ./...`
     - `GOCACHE=/tmp/namba-go-cache go vet ./...`
   - Run `git diff --check`.

10. Review readiness.
    - Keep product, engineering, and design review artifacts current.
    - Product review should challenge scope and exclusion discipline.
    - Engineering review should challenge helper ownership, fixture testing, generated-template safety, and GitHub failure modes.
    - Design review should focus on operator clarity for command-entry guidance, PR evidence, and frontend validation reports.
