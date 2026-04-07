# Engineering Review

- Status: clear
- Last Reviewed: 2026-04-07
- Reviewer: namba-planner
- Command Skill: `$namba-plan-eng-review`
- Recommended Role: `namba-planner`

## Focus

- Lock architecture, sequencing, failure modes, trust boundaries, and validation strategy before execution starts.

## Findings

- The SPEC correctly identifies the most dangerous failure mode: commands such as `project` and `sync` that currently ignore args and can perform mutations even when the caller is clearly probing with `--help`.
- The implementation area is appropriately scoped to shared parsing/rendering helpers plus the command handlers that currently parse too late or not at all.
- The acceptance criteria are specific enough to drive a test-first implementation: help-before-preconditions, non-mutating malformed invocations, delimiter-preserved literal flag text, and regression coverage for commands that used to ignore args.
- The scope is disciplined. It hardens parsing and help semantics without reopening unrelated command semantics such as how `pr`, `land`, or `run` fundamentally behave.

## Decisions

- Prefer a shared top-level helper for help detection and no-extra-args enforcement rather than bespoke `--help` branches spread across each command.
- Treat help parsing order as a first-class implementation requirement: parse before `requireProjectRoot`, git checks, auth checks, validators, or writes.
- Keep usage rendering stable and short enough that tests can assert on anchors rather than on every line of prose.

## Follow-ups

- During implementation, explicitly test commands that currently ignore args (`project`, `sync`, `doctor`, `status`) because they are the easiest place for regressions to hide.
- Add at least one regression for literal description text after `--`, because that is the main place where the hardened parser could accidentally become too aggressive.

## Recommendation

- Clear to proceed from an engineering perspective. The architecture, failure modes, and validation bar are specific enough to start implementation.
