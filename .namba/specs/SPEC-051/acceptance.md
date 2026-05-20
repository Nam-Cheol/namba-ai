# Acceptance

- [x] `namba plan` caps the slug segment after `SPEC-XXX` when the description is long.
- [x] `namba fix --command plan` uses the same capped planning branch slug behavior.
- [x] The generated branch keeps the configured `spec/` prefix and `SPEC-XXX` id intact.
- [x] Existing short description branch names remain unchanged.
- [x] Long English and Korean descriptions are covered by targeted regression tests.
- [x] Validation commands pass
- [x] Existing branch reuse, uniqueness, and dirty-workspace safeguards around planning start are preserved
- [x] Namba report next-work wording consistently names the concrete command, review, validation, or handoff to do next.
- [x] Namba clarification and report text follow the configured init language even when the user input uses another language.
- [x] Configured-language output covers Japanese and Chinese instead of falling back to English.
- [x] `namba plan` and `$namba-plan-review` guidance points back to plan review when readiness is not `Cleared reviews: 3/3`.
- [x] Japanese and Chinese Goal/Scope/Constraints/Acceptance labels are recognized as clarification evidence and do not create false missing open points.
