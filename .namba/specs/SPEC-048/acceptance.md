# Acceptance

- [ ] `internal/namba/testdata/evals/harness/` exists.
- [ ] The fixture directory contains `README.md`, `route_cases.json`,
  `prompt_refinement_cases.json`, `guardrail_cases.json`,
  `evidence_manifest_cases.json`, and `pr_review_cases.json`.
- [ ] Fixture schemas are documented in the README.
- [ ] Every fixture item has a name, input or command or manifest, expected
  result, and rationale.
- [ ] Route fixtures cover core harness change, domain harness change, direct
  artifact creation, ordinary feature/product planning, and planned fix.
- [ ] Route fixtures validate runtime-observable command and request outcomes,
  not a new standalone policy classifier.
- [ ] Prompt refinement fixtures cover required ambiguous prompts, clear
  structured prompts, concrete target-module prompts, and read-only explanation
  requests.
- [ ] Prompt refinement fixtures include `Explain how namba queue works without
  making changes` and `Read SPEC-048 and summarize implementation risks only`
  as clear read-only cases.
- [ ] Guardrail fixtures cover required dangerous commands, safe commands, and
  risky approval-note commands.
- [ ] Evidence manifest fixtures validate schema/version, request mode,
  execution mode, expected evidence sections, hook/artifact boundaries, and
  invalid failure diagnostics.
- [ ] Evidence manifest fixtures separate raw-schema validation from
  builder-normalization validation.
- [ ] PR review fixtures validate default no-review behavior, explicit
  `--review`, ignored legacy auto-review config, unrelated comments, and marker
  duplicate prevention.
- [ ] Go tests load and validate the relevant eval fixtures.
- [ ] Python tests load hook-related fixtures only if actual hook subprocess
  validation is required.
- [ ] Hook logic is not copied or forked into Go tests.
- [ ] Existing helpers are reused where practical; no net-new policy branch is
  added without an existing runtime or command-selection source of truth.
- [ ] No eval requires network, GitHub auth, Codex auth, external model access,
  user-specific local config, or persistent writes to the repository.
- [ ] Fixture failures report case name, input or command, expected value,
  actual value, and rationale.
- [ ] Fixture diagnostics make the eval pack's added value over ordinary unit
  tests visible at failure time.
- [ ] The fixture README explains why the eval pack exists, how to run it, how
  to add cases, what each fixture measures, what not to add, how this differs
  from unit tests, and why external LLM/API calls are prohibited.
- [ ] CI runs the deterministic eval tests through existing project validation
  commands.
- [ ] `go test ./...` passes.
- [ ] `python3 -m unittest discover -s tests` passes if Python eval or hook
  tests are present.
- [ ] `go vet ./...` passes.
- [ ] Existing formatting check passes.
- [ ] Manual flip validation is performed and reverted.
- [ ] Stale eval fixture references are searched before handoff.
- [ ] No runtime logs, eval reports, or temporary artifacts are committed.
- [ ] The eval pack remains small and curated.
