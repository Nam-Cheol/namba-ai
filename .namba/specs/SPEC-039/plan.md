# SPEC-039 Plan

1. Refresh planning context.
   - Read `.namba/project/product.md`, `.namba/project/tech.md`, mismatch and quality reports, `.namba/project/systems/workspace.md`, and current release/README docs.
   - Confirm `.namba/config/sections/docs.yaml`, `.namba/config/sections/git-strategy.yaml`, and `.namba/config/sections/language.yaml` because README generation, review comments, PR language, and release-note language depend on those configs.

2. Run plan review passes before implementation.
   - Use `$namba-plan-pm-review` to challenge scope, naming, and whether the two new skills are distinct enough from existing `$namba-pr`, `$namba-land`, GitHub plugin, and `namba release` behavior.
   - Use `$namba-plan-eng-review` to validate generated-surface registration, release-note handoff mechanics, GitHub thread resolution risks, and validation sequence.
   - Use `$namba-plan-design-review` for README visual structure, emoji density, multilingual readability, and accessibility.
   - Refresh `.namba/specs/SPEC-039/reviews/readiness.md`.

3. Implement the README visual upgrade.
   - Update `internal/namba/readme.go` Namba CLI root README sections with semantic emoji headings and scan markers.
   - Keep generated docs deterministic and avoid direct source-of-truth edits in `README*.md`.
   - Update existing README sync tests so they assert the richer section labels and do not become brittle around every emoji.

4. Implement `$namba-review-resolve`.
   - Add `renderReviewResolveCommandSkill()` in `internal/namba/templates.go`.
   - Register `namba-review-resolve` in `managedCodexSkillNames()` and `codexSkillTemplates()` in `internal/namba/codex.go`.
   - Expose the skill in `renderAgents()`, `renderNambaSkillCommandMappingSection()`, `renderNambaSkillExecutionRulesSection()`, Codex usage sections, README command skill sections, and skill-to-command mapping.
   - Define the meaningful/actionable triage contract, validation-before-reply rule, per-thread reply rule, resolve-only-when-addressed rule, and non-duplicated `@codex review` rerequest rule.
   - Add template/scaffold/docs tests proving the generated skill is present and its contract is visible.

5. Implement `$namba-release`.
   - Add `renderReleaseCommandSkill()` in `internal/namba/templates.go`.
   - Register `namba-release` in `managedCodexSkillNames()` and `codexSkillTemplates()` in `internal/namba/codex.go`.
   - Expose release routing in AGENTS, `$namba`, Codex usage docs, README command skills, skill mapping, workflow guide, and technical snapshot sections.
   - Keep release explicitly NambaAI-specific and document that `namba release` remains the guarded CLI primitive used by the skill.

6. Add release-note generation and publication support.
   - Decide the executable handoff path: either write `.namba/releases/<version>.md` before tagging and make the release workflow use it, or provide an equivalent notes-file path to the GitHub release publication step.
   - Build release notes from commits since the previous semver tag, grouping by user-visible change, fix, docs/workflow, and internal maintenance.
   - Preserve SPEC IDs, PR numbers, and short commit hashes when available.
   - Ensure the release workflow still publishes the existing archive matrix and `checksums.txt`.
   - Add unit tests for version selection, commit-range note rendering, and release workflow body selection where implemented.

7. Regenerate managed assets.
   - Run `namba regen` after template and managed skill changes.
   - Run `namba sync` after validation so `README*.md`, docs, `.namba/codex/README.md`, project release docs, and `.namba/manifest.json` reflect the new generated surfaces.

8. Validate.
   - Run `gofmt -l "cmd" "internal" "namba_test.go"`.
   - Run `go vet ./...`.
   - Run `go test ./...`.
   - Confirm `namba regen` and `namba sync` complete without removing the new managed skills.
   - Confirm generated README outputs contain the intended visual structure and release/review skill references.
