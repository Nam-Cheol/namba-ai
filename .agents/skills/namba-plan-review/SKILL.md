---
name: namba-plan-review
description: Command-style entry point for bundling SPEC creation, parallel plan reviews, and advisory readiness validation.
---

Use this skill when the user explicitly says `$namba-plan-review`, asks to create a SPEC and run the full pre-implementation review loop, or wants `namba plan` plus the review flow bundled into one skill.

Behavior:
- Resolve the target SPEC from an explicit `SPEC-XXX`; otherwise create the next SPEC with `namba plan` for feature work, `namba harness` for reusable agent/skill/workflow/orchestration work, or `namba fix --command plan` for bugfix planning.
- Prefer the installed `namba` CLI for the SPEC-creation step when it is available; keep `.namba/` as the source of truth if you need to do the setup manually.
- Inherit the same safe-by-default planning branch contract as `namba plan`: create or switch to the dedicated `spec/...` branch in the current workspace by default, treat `--current-workspace` as the explicit current-branch escape hatch, and do not create planning worktrees.
- Read `.namba/specs/<SPEC>/spec.md`, `plan.md`, and `acceptance.md` before launching reviews or revising the planning artifacts.
- Launch product, engineering, and design review passes in parallel when subagent routing is available, using `$namba-plan-pm-review`, `$namba-plan-eng-review`, and `$namba-plan-design-review` as the three authoritative review tracks.
- Prefer `namba-product-manager`, `namba-planner`, and `namba-designer` for the three review tracks, and use `namba-plan-reviewer` as the aggregate validator when custom-agent routing is available.
- After the three review tracks finish, run an aggregate validation pass over `spec.md`, `plan.md`, `acceptance.md`, and `.namba/specs/<SPEC>/reviews/*.md` to check coverage gaps, contradictions, and whether the advisory readiness state is credible.
- If the aggregate validator finds issues, revise the SPEC or review artifacts directly, rerun only the affected review tracks, and repeat the validation loop instead of restarting every pass blindly.
- Refresh `.namba/specs/<SPEC>/reviews/readiness.md` after each review and validation cycle so the advisory summary stays current.
- Keep the loop bounded and explicit: stop when the readiness state is clear enough to proceed or when the remaining blockers are concrete enough that another loop would be redundant.
- Keep the whole flow advisory by default; missing depth or blockers should be visible, not silently converted into a hard gate.
