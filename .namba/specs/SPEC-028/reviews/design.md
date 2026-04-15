# Design Review

- Status: clear
- Last Reviewed: 2026-04-15
- Reviewer: Codex design review
- Command Skill: `$namba-plan-design-review`
- Recommended Role: `namba-designer`

## Focus

- Clarify interaction quality, responsive states, accessibility, and visual direction before implementation starts.

## Findings

- The core framing is now clear and well-bounded: v1 is explicitly a machine-consumable, tail-friendly progress source, not a direct live UI. That removes the earlier ambiguity about whether SPEC-028 is supposed to deliver an operator-ready watch experience in the same slice.
- The event contract is stronger after the clarifications. `source`, `scope`, stable baseline phases, and the human-readable summary/detail convention now give future watch surfaces a usable operator-facing layer without forcing them to expose raw internal execution phrasing.
- The phase model is mostly credible for downstream consumers. The explicit note that baseline watcher-facing phases are compatibility-sensitive and that extra phases must be additive is the right guardrail for later Codex-native streaming integration.
- The remaining semantic soft spot is `done`. The spec now says consumers must interpret it together with `scope`, `status`, and stable summary/detail fields, which is a meaningful improvement, but the meaning is still intentionally broad. That is acceptable for this slice as long as later watch UIs do not collapse those distinctions into a single human phrase.
- Operator readability is materially improved by the clarified summary/detail requirement. The contract now gives later plain-text readers enough stable material to describe failures, preserved workers, and merge-blocked states without depending on color, icons, or hidden aggregate state.

## Decisions

- Advisory decision: proceed with SPEC-028 as designed; the event-source framing is now sufficiently clear for implementation.
- Treat the JSONL artifact as the canonical progress source for future watchers, with summary/detail conventions intentionally reserved for later operator-facing rendering.
- Keep deferred UI work out of scope for SPEC-028 and preserve the rule that future watch/readout layers render from the stable contract rather than inventing alternate human phase semantics.

## Follow-ups

- Keep implementation notes explicit about how `done` should be rendered for worker scope versus run scope so future watch surfaces do not flatten distinct completion moments into one ambiguous label.
- Ensure summary/detail text stays concise and stable enough for plain-text readouts; rich diagnostic payloads should remain in metadata instead of leaking into the human-facing path by default.
- When Codex-native streaming is added later, require that finer-grained upstream events map onto the same operator-facing baseline phases and summary/detail conventions rather than creating a second readability model.

## Recommendation

- Recommend proceeding with SPEC-028. Design is now clear enough for implementation, with only non-blocking follow-ups around how later watch surfaces present `done` and how summary/detail text remains disciplined as richer event sources are added.
