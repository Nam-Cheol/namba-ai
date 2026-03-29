# SPEC-018 Plan

1. Inspect the current capability probe and resume invocation rendering paths.
2. Gate resume probing on the planned invocation set instead of probing it unconditionally.
3. Render resume commands so exec-level flags can appear before `resume --last` when the installed CLI supports that shape.
4. Update team and parallel preflight coverage to match the new representability rules.
5. Add regression tests for conditional resume probing and exec-level resume flags.
6. Run validation and sync the `.namba` artifacts.
