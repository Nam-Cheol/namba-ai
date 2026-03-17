# Namba Output Contract

This repository uses a NambaAI-specific output contract for substantial task responses.

## Contract

Keep the semantic flow in this order:
1. `오늘의 결정`
2. `판단 근거`
3. `검증 경로`
4. `무너지는 조건`
5. `다음 수`

## Namba Style

- The semantic order is fixed, but the exact labels may vary.
- Recommended Namba label palette:
  - `오늘의 결정`, `핵심 판단`, `이번 수`
  - `판단 근거`, `왜 이렇게 봤나`, `근거`
  - `검증 경로`, `검증 방법`, `확인 루트`
  - `무너지는 조건`, `실패 조건`, `경계 조건`
  - `다음 수`, `추천`, `권장 흐름`
- The answer should read like a concise engineering handoff rather than a stiff checklist.

## Scope

- Apply the full contract to implementation summaries, design decisions, operational guidance, code reviews, and other substantial responses.
- Very short acknowledgements or one-line factual replies may stay shorter, but substantial responses should keep the same semantic order.

## Validation

- Use `.namba/codex/validate-output-contract.py --file <response.md>` to validate a saved response.
- Use `.namba/codex/validate-output-contract.py` and pipe UTF-8 text through stdin to validate ad hoc content.

## Hook Status

- OpenAI Codex docs currently document AGENTS, repo skills, built-in slash commands, and config, but they do not document a repository-configurable stop-hook surface.
- Treat the validator script as the fallback enforcement path until upstream hook support is documented.
