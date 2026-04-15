# Product Review

- Status: clear
- Last Reviewed: 2026-04-15
- Reviewer: Codex product reviewer
- Command Skill: `$namba-plan-pm-review`
- Recommended Role: `namba-product-manager`

## Focus

- Challenge the problem framing, scope, user value, and acceptance bar before implementation starts.

## Findings

- Low: The previous concern about overpromising visible UX is now adequately addressed. `spec.md` and `plan.md` explicitly frame v1 as durable, machine-consumable progress persistence, not a shipped watch surface or `namba status` upgrade.
- Low: The previous concern about phase stability is now adequately addressed. `spec.md` and `acceptance.md` both treat the seven watcher-facing phases as compatibility-sensitive baseline semantics and constrain additional phases to additive extensions.
- Low: The previous concern about operator-readable detail is now adequately addressed. The event contract and acceptance bar now explicitly reserve stable summary/detail or metadata capacity for failures, preserved workers, and merge-blocked outcomes, which is enough for later watch UX without forcing a schema break.
- Low: One product caution remains for implementation discipline rather than planning scope. Because `done` is intentionally broad and interpreted with `scope`, `status`, and stable detail fields, follow-on consumers should avoid collapsing it into a simplistic "success" label in future readouts.

## Decisions

- The current delivery boundary is now product-clear. This remains an infrastructure slice that delivers trustworthy progress data first and deliberately defers the visible watch surface.
- The JSONL event stream should be treated as the primary product contract for future live-progress features. The terminal aggregate report remains useful, but later watcher UX should read from the event stream rather than inventing a second live-state model.
- Source-awareness remains the right product decision. It keeps this slice extensible toward Codex-native streaming without reframing the v1 contract later.

## Follow-ups

- Keep implementation and later docs aligned with the now-correct v1 expectation: durable live progress data now, visible watch UX later.
- Preserve the baseline meaning of `queued`, `running`, `validating`, `merge_pending`, `merging`, `done`, and `failed` as a compatibility surface; extend additively only.
- When the follow-up watch/readout slice starts, define human-facing wording rules so broad terminal phases such as `done` are rendered with enough context instead of being treated as raw user-facing copy.
- In the next follow-up SPEC, prioritize either `namba run --watch` or another thin event-log reader so this slice's value becomes directly visible to users.

## Recommendation

- Proceed. From a product standpoint, the earlier concerns are adequately addressed and the slice is now clear enough to implement, with the remaining caution limited to how future watch UX interprets the stable event contract.
