# Design Review

- Status: approved
- Last Reviewed: 2026-04-07
- Reviewer: Codex
- Command Skill: `$namba-plan-design-review`
- Recommended Role: `namba-designer`

## Focus

- Clarify interaction quality, responsive states, accessibility, and visual direction before implementation starts.

## Findings

- 정보 구조가 충분히 정리됐습니다. `product.md`가 first landing document, `tech.md`가 technical hub, `structure.md`가 appendix라는 읽기 경로가 이제 명시돼 있습니다.
- 시스템별 반복 패턴도 생겼습니다. purpose -> entry points/interfaces -> modules -> data/state -> auth/integrations -> deploy/runtime/test risks 순서는 멀티앱 저장소에서도 scan path를 안정화합니다.
- mismatch report와 evidence references도 핵심 문서에서 surfaced되어야 한다는 기준이 들어가 있어, 부록에만 갇히지 않고 실제 읽기 흐름 안에서 소비될 수 있습니다.
- 남은 디자인 과제는 표현 규칙의 절제입니다. evidence/confidence 표기는 읽기성을 해치지 않도록 스캔 가능한 수준으로 유지해야 합니다.

## Decisions

- 이 SPEC는 implementation planning으로 진행해도 됩니다.
- 문서 계층은 현재 정의한 three-layer reading model을 유지하세요: landing -> technical hub/per-system summary -> appendix/supporting artifacts.
- evidence/confidence 표기는 스캔 가능한 고정 패턴으로 제한하세요.

## Follow-ups

- 시스템별 summary template를 실제 generated docs에서 동일한 rhythm으로 유지하세요.
- evidence/confidence 표현을 과도한 메타데이터 나열이 아니라 읽기 보조 장치로 제한하세요.
- mismatch report 링크가 primary docs에서 실제로 발견 가능하도록 renderer 설계를 검증하세요.

## Recommendation

- Advisory recommendation: approved for implementation planning. The reading hierarchy is now explicit enough to implement without confusing the human scan path.
