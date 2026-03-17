# Namba Output Contract

This repository uses a NambaAI-specific output contract for substantial task responses.

## Contract

- Use a decorated header such as `# NAMBA-AI 작업 결과 보고`.
- Keep the report sections in this order:
1. `🧭 작업 정의`
2. `🧠 판단`
3. `🛠 수행한 작업`
4. `🚧 현재 이슈`
5. `⚠ 잠재 문제`
6. `➡ 다음 스텝`

## Namba Style

- The header and label palette should follow the init-selected language: Korean.
- The semantic order is fixed, but the exact labels may vary within the selected language palette.
- Light visual styling such as simple emoji section markers is encouraged when it improves scanability.
- Recommended label palette:
  - `🧭 작업 정의`: `작업 정의`, `정의`, `정의한 범위`, `문제 정의`
  - `🧠 판단`: `판단`, `내린 판단`, `핵심 판단`, `결정`
  - `🛠 수행한 작업`: `수행한 작업`, `진행한 작업`, `작업 내용`, `적용한 작업`
  - `🚧 현재 이슈`: `현재 이슈`, `이슈`, `남은 이슈`, `현재 문제`
  - `⚠ 잠재 문제`: `잠재 문제`, `잠재 리스크`, `위험 요소`, `잠재 이슈`
  - `➡ 다음 스텝`: `다음 스텝`, `다음 단계`, `추천`, `권장 흐름`
- The answer should read like a concise engineering field report rather than a stiff checklist.

## Scope

- Apply the full contract to implementation summaries, design decisions, operational guidance, code reviews, and other substantial responses.
- Very short acknowledgements or one-line factual replies may stay shorter, but substantial responses should keep the same semantic order.

## Validation

- Use `.namba/codex/validate-output-contract.py --file <response.md>` to validate a saved response.
- Use `.namba/codex/validate-output-contract.py` and pipe UTF-8 text through stdin to validate ad hoc content.

## Hook Status

- OpenAI Codex docs currently document AGENTS, repo skills, built-in slash commands, and config, but they do not document a repository-configurable stop-hook surface.
- Treat the validator script as the fallback enforcement path until upstream hook support is documented.
