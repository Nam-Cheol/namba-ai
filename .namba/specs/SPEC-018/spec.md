# SPEC-018

## Problem

Codex capability preflight currently treats `codex exec` and `codex exec resume` as if they always share the same capability surface.

This creates two regressions:

- runs that never plan a resume turn can still fail preflight because `codex exec resume --help` is probed unconditionally
- resume turns reject valid requests such as `profile` or `sandbox` even when the installed Codex CLI accepts those options as exec-level flags before `resume`

## Goal

Make Codex capability preflight and invocation rendering match the real CLI contract for planned runs.

## Context

- Project: namba-ai
- Project type: existing
- Language: go
- Mode: tdd
- Work type: fix

## Desired Outcome

- probe `codex exec resume --help` only when the planned invocation set actually includes a resume turn
- allow resume invocations to place supported exec-level flags before `resume --last`
- keep `--team` and `--parallel` preflight aligned with the same representability logic
