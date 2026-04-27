# Acceptance

## README Visual Upgrade

- [ ] Generated Namba CLI README output uses purposeful emoji in major section headings or scan markers for command choice, workflow stages, install/update/release concepts, and technical snapshot areas.
- [ ] Emoji usage improves scanning without replacing precise command names or making every bullet decorative.
- [ ] The source of truth is `internal/namba/readme.go`; generated `README*.md` files are refreshed by `namba sync`, not hand-edited as primary source.
- [ ] The generated header, hero image, language links, Latest Release/CI/Security links, command examples, and multilingual outputs remain intact.
- [ ] English, Korean, Japanese, and Chinese root README outputs remain coherent after the visual update.
- [ ] README sync tests cover the new visual structure with stable anchors rather than brittle full-file snapshots.
- [ ] README visual changes do not rename, reorder, or visually overpower the existing command-selection model.
- [ ] Emoji usage has an explicit density rule: section headings by default, selected lifecycle/caution bullets only when they add scan value, and no emoji inside command literals, language links, release/CI/security links, or shell snippets.
- [ ] Stable visual anchors are validated for at least command choice, quick start, hook runtime, command skills, release flow, and technical snapshot sections across all generated root README languages.

## `$namba-review-resolve`

- [ ] `namba init` and `namba regen` generate `.agents/skills/namba-review-resolve/SKILL.md`.
- [ ] `namba-review-resolve` is included in the managed skill registry and manifest after regeneration.
- [ ] `$namba-review-resolve` is exposed consistently in `AGENTS.md`, `.agents/skills/namba/SKILL.md`, `.namba/codex/README.md`, README command skill sections, workflow guide content, and skill-to-command mapping.
- [ ] The skill triggers on `$namba-review-resolve`, review-addressing requests, and Korean wording equivalent to "리뷰 확인하고 의미있는 리뷰면 수정 후에 해당 리뷰에 답변을 달고, resolve한 다음 다시 리뷰 요청해".
- [ ] The skill resolves the target PR from the current branch when no PR number is given.
- [ ] The skill reads unresolved review threads with thread-level resolution state when available.
- [ ] The skill distinguishes meaningful/actionable comments from duplicates, stale remarks, preference-only comments, or comments already satisfied by current code.
- [ ] Meaningful comments are fixed with scoped edits, validated, answered on the original thread with evidence, and resolved only after being addressed.
- [ ] Non-actionable comments are not silently resolved; they are left open or answered with a clear rationale according to the thread state.
- [ ] Every reviewed thread receives an explicit outcome: `fixed-and-resolved`, `answered-open`, or `skipped-with-rationale`.
- [ ] Mixed PRs are handled safely: re-review is not requested while any meaningful/actionable thread remains unresolved or lacks a concrete follow-up owner.
- [ ] Review is requested again only after meaningful comments are addressed and validation passes.
- [ ] The configured `@codex review` marker is confirmed or added once without duplication.
- [ ] Thread discovery, per-thread replies, and thread resolution use a thread-aware GitHub path such as `gh` GraphQL when connector comments cannot expose review-thread state.
- [ ] Flat PR comments are allowed only for PR-level context and confirming or adding the configured review request marker.

## `$namba-release`

- [ ] `namba init` and `namba regen` generate `.agents/skills/namba-release/SKILL.md`.
- [ ] `namba-release` is included in the managed skill registry and manifest after regeneration.
- [ ] `$namba-release` is exposed consistently in `AGENTS.md`, `.agents/skills/namba/SKILL.md`, `.namba/codex/README.md`, README command skill sections, workflow guide content, release docs, and skill-to-command mapping.
- [ ] The skill triggers on `$namba-release`, `namba release` in Codex workflow context, and Korean wording such as "릴리즈 진행해".
- [ ] The release skill is described as NambaAI-specific, not a generic release helper.
- [ ] Release orchestration starts from clean `main` and runs configured validation before tagging.
- [ ] The release version can be explicit or derived from the existing semver bump behavior.
- [ ] Release notes are generated from commits since the previous semver tag before the release is published.
- [ ] Release notes group meaningful work areas and preserve SPEC IDs, PR numbers, or short commit hashes when available.
- [ ] Release notes use the configured repo language for release/PR-facing content, Korean for this repository unless config changes.
- [ ] A durable per-version notes artifact exists before tagging, for example `.namba/releases/<version>.md`, or an equivalent explicit file handoff consumed by the publish workflow.
- [ ] Generated release notes include minimum sections for user-visible changes, fixes, docs/workflow, and internal maintenance; empty sections are omitted or marked explicitly.
- [ ] If commit history is too noisy or ambiguous to generate reviewer-safe notes, `$namba-release` stops for manual clarification instead of publishing generic notes.
- [ ] If a committed notes artifact is used, the release-note prep commit is excluded from the release-note commit range.
- [ ] Any release-note artifact generation, `namba regen`, or `namba sync` work happens before the final clean-tree `namba release --version <version> --push` invocation.
- [ ] The GitHub Release body uses the generated notes rather than an empty or generic body.
- [ ] Existing release assets and `checksums.txt` publication remain unchanged.
- [ ] `namba release` clean-main, clean-tree, validation, duplicate-tag, semver, and push guardrails remain intact.

## Tests And Validation

- [ ] `internal/namba/templates_test.go` covers both new skill templates and `$namba`/AGENTS exposure anchors.
- [ ] Init/regen scaffold tests cover generation and managed registry inclusion for both new skills.
- [ ] README sync tests cover emoji section anchors and both new skill references.
- [ ] Release command or workflow tests cover commit-based release-note rendering and GitHub Release body handoff.
- [ ] `gofmt -l "cmd" "internal" "namba_test.go"` reports no files.
- [ ] `go vet ./...` passes.
- [ ] `go test ./...` passes.
- [ ] `namba regen` passes.
- [ ] `namba sync` passes.
