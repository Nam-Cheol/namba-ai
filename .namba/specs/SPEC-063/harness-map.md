# SPEC-063 Harness Map

| Requirement | Source surface | Implementation owner | Validation |
| --- | --- | --- | --- |
| Sync-managed verification guide | README/docs sync renderer and docs config | `internal/namba` docs sync code or templates | Generated docs tests and `namba sync` drift check |
| README link | README generated sections | README renderer tests | Golden/link assertion |
| Workflow guide link | Workflow guide generated sections | Docs renderer tests | Golden/link assertion |
| Eval/report/quality content | Existing CLI behavior and `scripts/quality.sh` | Verification-guide content source | Content assertions plus command validation |
| `.namba` evidence explanation | Project/run/queue/hook evidence docs and code | Verification-guide content source | Content assertions |
| CI consumption | GitHub Actions and quality script parity | Verification-guide content source | Content assertions and CI parity review |
| Final-report detail preservation | Codex hook prompt/template/output contract | `.codex/hooks` and `.namba/codex` contract owners | Hook/output-contract regression |

## Review Handoff

- Product review owns scope, generated-doc discoverability, and user value.
- Engineering review owns renderer/source-of-truth choices, test coverage, idempotence, and compatibility.
- Design review owns documentation scanability and report-frame information architecture only.
