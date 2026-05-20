# Acceptance

- [ ] `namba plan` caps the slug segment after `SPEC-XXX` when the description is long.
- [ ] `namba fix --command plan` uses the same capped planning branch slug behavior.
- [ ] The generated branch keeps the configured `spec/` prefix and `SPEC-XXX` id intact.
- [ ] Existing short description branch names remain unchanged.
- [ ] Long English and Korean descriptions are covered by targeted regression tests.
- [ ] Validation commands pass
- [ ] Existing branch reuse, uniqueness, and dirty-workspace safeguards around planning start are preserved
