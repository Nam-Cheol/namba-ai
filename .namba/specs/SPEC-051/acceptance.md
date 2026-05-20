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
