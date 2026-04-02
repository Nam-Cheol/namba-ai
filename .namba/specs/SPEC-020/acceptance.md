# Acceptance

- [ ] `namba plan "<description>"` preserves its current default feature-planning behavior and still creates the next sequential feature-oriented `SPEC-XXX` package with the expected review artifacts.
- [ ] `namba harness "<description>"` creates the next sequential `SPEC-XXX` package with the expected review artifacts and a harness-oriented scaffold.
- [ ] `namba harness --help` behaves as a read-only help flow and does not create or mutate any `SPEC-XXX` package.
- [ ] `namba plan --help` remains a read-only help flow, with the contract aligned to `SPEC-019` rather than reimplemented inconsistently inside this feature.
- [ ] Harness-template `spec.md` explicitly captures the intended Codex execution topology and does so using Namba/Codex terms rather than Claude Team API terms.
- [ ] Harness-template artifacts explicitly map reusable work to `.agents/skills/*`, built-in subagents, and `.codex/agents/*.toml` custom agents where relevant, and do not instruct users to generate `.claude/*` assets.
- [ ] Harness-template artifacts include progressive-disclosure guidance for reusable skills, covering metadata, `SKILL.md`, and optional `references/`, `scripts/`, or `assets/`.
- [ ] Harness-template artifacts include a concrete trigger strategy with should-trigger versus should-not-trigger guidance for skill descriptions.
- [ ] Harness-template artifacts include a concrete evaluation strategy with with-skill versus baseline comparison guidance where applicable, plus assertion/timing capture guidance for measurable workflows.
- [ ] Help text, repo-local skills, and generated docs clearly explain when to use default `namba plan` versus `namba harness`.
- [ ] No generated harness-template artifact emits `.claude/agents`, `.claude/skills`, `TeamCreate`, `SendMessage`, `TaskCreate`, or a mandatory `model: "opus"` requirement as if they were part of Namba's Codex contract.
- [ ] Regression tests cover template parsing, scaffold generation, non-portable primitive exclusion, and generated help/doc guidance where stable assertions are appropriate.
- [ ] Validation commands pass.
