# Product Review

- Status: approved
- Last Reviewed: 2026-05-20
- Reviewer: Codex (`namba-product-manager`)
- Command Skill: `$namba-plan-pm-review`
- Recommended Role: `namba-product-manager`

## Focus

- Challenge the problem framing, scope, user value, and acceptance bar before implementation starts.

## Findings

1. High: The initial acceptance criteria did not explicitly limit where
   optional plugin, marketplace, remote-control, and remote-environment
   adoption guidance may appear. Without that limit, normal local workflows such
   as `namba plan`, `namba run`, and `namba pr` could become noisier even while
   claiming to remain optional.
2. Medium: The initial docs acceptance did not force a simple explanation of
   what users actually need for local NambaAI use versus what Codex 0.131 can
   expose as optional readiness.
3. Medium: "Stable CLI or API behavior" needed a product-readable evidence
   rule so docs, runtime summaries, and evidence output do not drift.

## Decisions

- Local-first remains the product default. Plugins, marketplace sharing,
  remote-control, remote environments, and the Python SDK are readiness paths,
  not prerequisites for normal NambaAI use.
- Plugin or remote adoption should become an action prompt only when the user
  explicitly asks for that path. Otherwise Namba may expose neutral capability
  or evidence status.
- `SPEC-050` continues to own core hook/config compatibility and `SPEC-052`
  continues to own baseline diagnostics evidence.

## Follow-ups

- Acceptance was revised to forbid adoption guidance in normal local workflows
  unless readiness inspection or already-collected status evidence is in scope.
- Acceptance was revised to require README, Korean README, and generated docs
  to separate local-use requirements from optional Codex platform readiness.
- Acceptance was revised to define stable remote-control or environment behavior
  evidence as local CLI/help output, stable local API behavior, authoritative
  config, or existing repository code evidence.

## Recommendation

- Proceed after the revised acceptance and visibility rules are preserved during
  implementation.
