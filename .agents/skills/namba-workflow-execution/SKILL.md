---
name: namba-workflow-execution
description: Execute NambaAI SPEC packages with Codex-native workflow and explicit validation.
---

Use this skill when implementing a SPEC package.

Execution pattern:
1. Read `.namba/specs/<SPEC>/spec.md`
2. Read `.namba/specs/<SPEC>/plan.md`
3. Read `.namba/specs/<SPEC>/acceptance.md`
4. Implement the work directly in the current Codex session
5. Run configured validation commands
6. Summarize results in `.namba/logs` and sync artifacts

Do not call `namba run` from inside Codex unless the user explicitly requests the non-interactive CLI runner.
