# SPEC-063 Plan

1. Refresh project context with `namba project` if implementation starts from stale project docs.
2. Locate the sync-managed docs sources:
   - README and workflow guide renderer/config ownership.
   - Tests that assert generated README/docs content.
   - Existing generated-file warning comments.
3. Add the verification guide to the sync generation path:
   - Generate `docs/verification-guide.md`.
   - Add README read-next/navigation links.
   - Add workflow-guide navigation/reference links.
   - Decide and test localized behavior so no generated link points to a missing file.
4. Write the verification guide content:
   - `namba eval`
   - `namba report`
   - `scripts/quality.sh`
   - quality artifact paths and defaults
   - accumulated `.namba` evidence
   - CI consumption
   - failure interpretation and next actions
5. Fix final-report detail preservation:
   - Identify the hook prompt/template/output-contract source that emits the final Namba report-frame instruction.
   - Change it to preserve substantive details from the pre-hook response while enforcing section order.
   - Keep `.namba/codex/validate-output-contract.py` aligned if the validation contract owns this behavior.
6. Add targeted regression tests:
   - Generated docs and navigation golden assertions.
   - Sync idempotence or generated drift check.
   - Hook/output-contract detail preservation case.
7. Run validation:
   - Targeted tests changed by the implementation.
   - `namba sync`
   - Generated drift/idempotence check from the intended source state.
   - `go test ./...`
   - `gofmt -l "cmd" "internal" "namba_test.go"`
   - `go vet ./...`
   - `scripts/quality.sh`
8. Record the validation evidence:
   - Commands run.
   - Generated docs diff/idempotence result.
   - Quality artifact paths.
   - Any environment-only blocker with exact failure and next action.
