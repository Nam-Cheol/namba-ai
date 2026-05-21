# Design Review

- Status: clear
- Last Reviewed: 2026-05-21
- Reviewer: namba-designer
- Command Skill: `$namba-plan-design-review`
- Recommended Role: `namba-designer`

## Focus

- This is a CLI workflow review, so the design surface is operator comprehension: queue state clarity, recovery messaging, Windows-safe wording, and failure modes that do not misclassify queue-owned writes as user mistakes.

- Evidence Status: spec-only
- Gate Decision: clear
- Approved Direction: keep the queue conveyor terse and operational: show what SPEC is active, why progress stopped when it stops, which writes are queue-owned, and what exact recovery path is available when remote PR handoff is unavailable.
- Banned Patterns: vague "workspace is dirty" failures, Unix-only wording, exposing transport implementation details as user-facing jargon, silent fallback behavior, and queue states that make local automation writes indistinguishable from user edits.
- Negative-First Contract: if the queue cannot continue, lead with the blocking reason first, then the affected SPEC, then the exact next command or manual recovery path.
- Default Library Fit: existing Namba CLI text-first output patterns are the right base; no new visual language is needed beyond disciplined status labels and ordering.
- Context-Specific Bans And Replacements: replace generic "clean your branch" language with ownership-aware guidance that names whether the change came from queue metadata, run evidence, or unrelated user work. Replace "PR handoff failed" with a message that states whether the queue can continue locally and what branch/main action happens next.
- Reference-Driven Asset Manifest: not-applicable.
- Generated Image Plan: not-applicable.
- Visual Grammar: `queue target -> current SPEC -> state -> blocker or next step -> recovery command`. Detail may follow, but the first screen must answer whether the conveyor is continuing, paused, or stopped.
- Generic-Section Proof: the most generic failure would be a raw error dump after each SPEC. The CLI should instead collapse incidental detail and preserve one stable operator mental model across dry-run, normal run, validation failure, and remote-handoff fallback.
- Architecture Handoff: queue execution should classify outcomes before rendering so messaging can stay consistent across success, blocked, fallback, and interrupted states.
- Violation-Check Plan: implementation validation should assert human-readable output for queue-owned dirty writes, unrelated user dirty work, Windows-safe request transport fallback, remote PR handoff fallback, and the terminal condition when the selected SPEC list is fully processed.
- Open Questions: none blocking from the review artifact alone.
- Unresolved Questions: exact wording for local main merge fallback should be finalized during implementation so it is accurate on Windows shells as well as POSIX shells.
- Design Review Axes: operator hierarchy, recovery clarity, cross-platform wording, state ownership, interruption handling, plain-text scanability
- Keep / Fix / Quick Wins: keep the scope narrow around continuity. Fix wording that conflates queue automation with user edits. Quick win: standardize one sentence pattern for blocked, resumed, and completed states.

## Review Checklist

- Art direction is clear and fits the task context.
- Palette, imagery, and motion are not applicable to this CLI workflow surface.
- Status labels and message ordering match the content instead of defaulting to raw logs or stack-first dumps.
- The review contract is complete enough to prevent generic CLI fallback, especially around ownership-aware dirty checks and recovery messaging.
- Cross-platform wording avoids assuming Bash, Unix path conventions, or argv-safe payload sizes.
- The most generic section risk is addressed by defining how queue state should read in blocked, resumed, fallback, and completed conditions.
- Anti-overcorrection guardrails hold: the CLI stays terse and operational rather than becoming verbose tutorial text.

## Findings

- The main UX risk is not the continuity fix itself but misreporting why continuity stopped. If queue-owned metadata or run evidence is treated like user dirtiness without an ownership explanation, operators will not trust retries and may manually clean state incorrectly.
- Windows compatibility is partially a messaging problem. Even if request transport moves off argv internally, user-visible copy should avoid implying shell-specific invocation details unless the operator must act on them.
- Remote PR handoff fallback needs an explicit state transition. "Handoff failed" is insufficient because the operator needs to know whether the queue stopped, continued on a local branch, or expects a later manual merge into `main`.
- Queue processing over a SPEC list needs a stable completion grammar. Without it, operators can mistake "current SPEC finished" for "entire requested queue completed," especially after intermittent fallback or validation interruptions.
- Dirty-branch checks should distinguish three cases in human output: queue-owned writes that are safe, unrelated user edits that block, and mixed state that requires inspection. Collapsing these into one branch cleanliness error would be confusing.
- Recovery copy should be command-oriented and specific. "Resolve the issue and rerun" is too weak; the CLI should indicate the relevant queue command or the exact manual fallback milestone already reached.

## Decisions

- Approve the implementation direction if queue messaging becomes ownership-aware and stateful rather than error-dump oriented.
- Require one consistent vocabulary set for `continuing`, `blocked`, `fallback`, `resumed`, and `completed` so operators do not have to infer queue state from surrounding prose.
- Prefer plain-language references to "request payload" or "temporary request file" over low-level transport wording unless debug mode explicitly asks for internals.
- Treat local branch creation and local `main` merge fallback as first-class, documented outcomes rather than exceptional edge notes buried in logs.

## Follow-ups

- [non-blocking] Define canonical one-line status messages for dry-run, active run, validation block, remote handoff fallback, and queue completion before implementation spreads wording across multiple call sites.
- [non-blocking] Add regression expectations for mixed dirty-state messaging so queue-owned writes and unrelated edits cannot regress back into one generic failure string.
- [non-blocking] Validate that any recommended manual commands or branch references remain readable and unambiguous in both Windows and POSIX terminal contexts.
- [non-blocking] Ensure the final queue summary reports both per-SPEC outcomes and the aggregate requested-range outcome.

## Recommendation

- Proceed. The UX risk is moderate but well-bounded: success depends on clear operator state transitions, ownership-aware dirty-check messaging, and explicit fallback/completion wording more than on any visual design concern.
