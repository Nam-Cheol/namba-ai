# Design Review

- Status: completed
- Last Reviewed: 2026-05-20
- Reviewer: Codex (`namba-designer` perspective)
- Command Skill: `$namba-plan-design-review`
- Recommended Role: `namba-designer`

## Focus

- Clarify documentation hierarchy, information design, terminology control, and
  reader-comprehension risks before implementation starts.

- Evidence Status: spec.md, plan.md, acceptance.md reviewed
- Gate Decision: advisory proceed with documentation clarification before
  implementation wording is finalized
- Approved Direction: keep the current technical scope, but tighten the reader
  contract around ownership boundaries, terminology, and doc sequencing
- Banned Patterns: anthropomorphizing Codex UI state as Namba policy; implying
  repo config controls user approval mode, sandbox mode, service tier, or
  workspace roots; mixing validation boundary language with session guardrail
  language in the same sentence without contrast
- Negative-First Contract: avoid language that can be misread as "Namba sets or
  guarantees Codex permissions"; document what Namba observes, documents, or
  validates instead
- Default Library Fit: not a visual UI task; translate design review into
  documentation hierarchy, label precision, and low-ambiguity information flow
- Context-Specific Bans And Replacements: ban "Namba permissions",
  "Namba workspace roots", and similar ownership-confusing shorthand; replace
  with "Codex displays...", "repo-managed docs explain...", or
  "Namba-generated config does not set..."
- Reference-Driven Asset Manifest: none required
- Generated Image Plan: not applicable
- Visual Grammar: separate runtime display, repo-managed config, interactive
  hooks, and run-time validation into distinct explanation layers
- Generic-Section Proof: the most generic area is the broad documentation
  requirement list; it needs a clearer hierarchy for maintainers versus end
  users
- Architecture Handoff: documentation should present a three-layer model:
  Codex runtime/UI state, repo-managed generated assets, and Namba validation
  boundary
- Violation-Check Plan: check every user-facing sentence for accidental policy
  ownership claims or collapsed terminology around permissions and roots
- Open Questions: should generated docs include a compact terminology table for
  "approval mode", "permissions", "effective workspace roots", and
  "scoped write roots"?
- Unresolved Questions: none blocking at SPEC stage
- Design Review Axes: hierarchy, terminology, reader trust, policy ownership,
  maintainer comprehension, misuse resistance
- Keep / Fix / Quick Wins: keep the strong boundary emphasis; fix terminology
  collisions; add quick-win doc structure requirements for ordered explanation

## Review Checklist

- Documentation hierarchy matches the task context and distinguishes maintainer
  guidance from end-user explanation.
- Terminology stays precise enough that Codex runtime/UI state is not confused
  with Namba-owned policy or repo-managed config.
- Information grouping matches the content instead of collapsing runtime,
  config, hooks, and validation into one generic "safety" bucket.
- The do-not-mislead contract is complete enough to block ownership confusion,
  especially around approval mode, permissions, effective workspace roots, and
  scoped write roots.
- Reader-facing examples and phrasing define what Codex displays versus what
  Namba documents or validates.
- Sequence, if proposed, has a clear comprehension purpose: boundary first,
  then runtime display facts, then repo-generated surfaces, then limitations.
- The most generic documentation section is redesigned when a requirement block
  currently reads like an undifferentiated checklist.
- Anti-overcorrection guardrails hold: no needless jargon inflation, no false
  certainty about upstream Codex behavior, and no loss of implementation
  realism.

## Findings

1. Medium: The SPEC correctly names many Codex 0.131 surfaces, but the
   documentation requirements currently flatten them into a single narrative.
   A reader can easily miss the difference between:
   `Codex UI/runtime state` vs `repo-managed generated config` vs
   `interactive guardrails` vs `namba run validation`.
   Without an explicit hierarchy, maintainers may write docs that are accurate
   sentence-by-sentence but still misleading as a whole.
2. Medium: The wording around permissions, approval mode, and effective
   workspace roots is directionally correct, but still vulnerable to ownership
   drift. Phrases like "document how Codex now displays..." need a paired rule
   that these are observational surfaces, not settings Namba owns or persists.
   This is the highest communication-risk area in the SPEC.
3. Low: "Workspace roots" and "scoped write roots" are both present, but the
   SPEC does not yet require a reader-friendly contrast between them. That
   creates avoidable ambiguity for users who may infer broader write guarantees
   from a visible workspace root list.
4. Low: The documentation section is maintainable but generic. It tells authors
   what topics to include, not what order or framing to use. For a compatibility
   SPEC centered on trust boundaries, ordering is part of the design.

## Decisions

- Keep the technical direction and acceptance scope.
- Treat this as a documentation-information-design review, not a visual design
  review.
- Require docs to explain the boundary in a fixed order:
  `what Codex shows` -> `what repo-managed Namba assets generate` ->
  `what Namba validates` -> `what Namba cannot rely on or control`.
- Prefer contrastive wording in docs, for example:
  "Codex may display X; Namba does not set X in repo-managed config."
- Frame workspace-root language as visibility and boundary context, not as a
  promise that every shown root is equivalently writable or policy-owned.

## Follow-ups

- Add an explicit doc-writing requirement in implementation notes or downstream
  docs for a short terminology table covering:
  `approval mode`, `permissions`, `effective workspace roots`,
  `scoped write roots`, `.codex/hooks.json`, and `.namba/hooks.toml`.
- During implementation review, check README/AGENTS/generated-doc headings for a
  stable section order instead of a flat compatibility dump.
- Verify that no generated prose says or implies:
  "Namba sets Codex permissions", "Namba owns workspace roots", or
  "repo config controls approval mode/service tier".
- If examples are added, include at least one explicit limitation statement
  about Git helper hook behavior and one explicit limitation statement about
  Codex UI/runtime surfaces being session-dependent.

## Recommendation

- Recommendation: proceed, but treat documentation clarity as a first-class
  acceptance concern during implementation. The SPEC is technically ready, yet
  the wording must be executed with strict hierarchy and ownership discipline to
  avoid misleading users about Codex 0.131 permissions and workspace-root
  behavior.
