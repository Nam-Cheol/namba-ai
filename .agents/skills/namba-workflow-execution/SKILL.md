---
name: namba-workflow-execution
description: Execute NambaAI SPEC packages with Codex-native workflow and explicit validation.
---

Use this skill when implementing a SPEC package.

Execution pattern:
1. Read `.namba/specs/<SPEC>/spec.md`
2. Read `.namba/specs/<SPEC>/plan.md`
3. Read `.namba/specs/<SPEC>/acceptance.md`
4. Read `.namba/specs/<SPEC>/reviews/readiness.md` when present so advisory review status informs execution
5. Implement the work directly in the current Codex session
6. Run configured validation commands
7. Summarize results in `.namba/logs` and sync artifacts

Collaboration defaults: use a dedicated branch from `main` for the SPEC, open the PR into `main`, write the PR in Korean, and request `@codex review` on GitHub after the PR is open.

Do not call `namba run` from inside Codex unless the user explicitly requests the non-interactive CLI runner.
