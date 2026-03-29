# Engineering Review

- Status: approved
- Last Reviewed: 2026-03-29
- Reviewer: Codex
- Command Skill: `$namba-plan-eng-review`
- Recommended Role: `namba-planner`

## Focus

- Lock architecture, sequencing, failure modes, trust boundaries, and validation strategy before execution starts.

## Findings

- The acceptance contract is now implementation-ready for sync stability. It explicitly covers transient-path exclusion, no-op `namba sync` cleanliness, and the requirement that tracked generated artifacts including `.namba/manifest.json` stay clean when nothing meaningful changed.
- The versioning slice now has a usable engineering contract: `namba --version` must print `namba <version>`, tagged releases must report the tag, and local development builds must fall back to the literal `dev` label.
- The `namba update` UX slice is now specific enough to build and test. The success path requires target-version visibility, asset/platform context, and next-step guidance including deferred Windows restart messaging; the failure path requires requested-version context plus actionable release, asset, checksum, or network guidance.

## Decisions

- Keep sync stability, version visibility, update UX, and multilingual install lifecycle docs in one SPEC. They live in the same user-facing CLI/documentation surface and should be validated together.
- Treat release tags as the authoritative version source for published binaries, injected during release builds, with a clearly defined fallback label for local development builds.
- Keep `namba update` UX as structured terminal guidance rather than introducing interactive or stateful UI. The goal is predictable, high-signal console output that remains testable across platforms.

## Follow-ups

- During implementation, keep one shared version source wired through release builds, local builds, `namba --version`, and update messaging so the contract does not fragment.
- Add tests that cover both Windows deferred-restart messaging and Unix in-place update messaging.
- Keep uninstall guidance synchronized across `README.md`, `README.ko.md`, `README.ja.md`, `README.zh.md`, and any generated guide that owns install/update lifecycle documentation.

## Recommendation

- Advisory recommendation: approved. Proceed with implementation; the engineering contract is now specific enough for low-risk execution as long as the shared version-source wiring is preserved end to end.
