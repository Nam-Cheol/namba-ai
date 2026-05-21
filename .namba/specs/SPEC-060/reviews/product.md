# Product Review

- Status: approved
- Last Reviewed: 2026-05-21
- Reviewer: Codex (Product Review)
- Command Skill: `$namba-plan-pm-review`
- Recommended Role: `namba-product-manager`

## Focus

- 수정된 SPEC이 이전 제품 블로커를 실제 acceptance 규칙으로 해소했는지 재검증합니다.

## Findings

- Resolved: 사용자 여정 우선순위가 충분히 고정됐습니다. `spec.md`의 Problem과 User Journey Rules가 README 첫 화면이 답해야 할 질문 순서를 명시하고, `acceptance.md`가 "무엇인가 -> 왜 신뢰할 수 있나 -> 무엇을 먼저 실행하나 -> 어떤 명령을 고르나 -> 어디를 더 읽나"를 직접 검증 대상으로 승격했습니다.
- Resolved: 다국어 구조 동등성 기준이 명확해졌습니다. `Localization Parity`와 acceptance가 섹션 순서, navigation slot, CTA 목적, command-selection row, advanced-details 위치까지 동일해야 한다고 고정해서 언어별 임의 해석 여지를 줄였습니다.
- Resolved: README와 가이드 문서의 책임 분리가 제품 관점에서 충분히 선명해졌습니다. `User Journey Rules`가 README는 오리엔테이션과 명령 선택, `getting-started`는 설치와 첫 성공 경로, `workflow-guide`는 심화 워크플로와 고급 참조를 담당한다고 분리했습니다.
- Resolved: badge 및 CTA fallback 정책이 acceptance 수준으로 보강됐습니다. badge는 데이터가 없거나 부적합할 때 텍스트 링크 또는 omission으로 안전하게 degrade해야 하고, CTA는 unsupported HTML 버튼이 아니라 linked badge 또는 plain Markdown link로 제한되어 렌더링 실패 리스크를 줄였습니다.
- Resolved: centered hero 관련 모호성이 해소됐습니다. 수정본은 "현재 GitHub-compatible centered image helper는 재사용 가능" 수준으로 후퇴했고, acceptance는 hero의 핵심 요구를 기존 이미지 사용, alt text, hero 미설정 시 text-first fallback으로 고정했습니다. 즉 중심 정렬은 선택적 구현 힌트이고, 제품 요구는 아닙니다.
- Minor: command-selection 섹션은 여전히 여러 명령을 한 화면에 노출하므로 구현 시 정보량 제어가 중요합니다. 다만 이번 수정으로 "각 명령은 사용자 상황과 다음 문서에 연결"되어야 하므로, 이전처럼 무맥락 나열로 흐를 가능성은 낮아졌습니다.

## Decisions

- 이전 제품 블로커 5건은 모두 해소된 것으로 판단합니다.
- 현재 SPEC은 시각 구성 요소 체크리스트를 넘어서 사용자 여정, 문서 책임, 다국어 구조, fallback 규칙까지 acceptance에 반영했습니다.
- 제품 관점 상태는 `approved`로 상향합니다.

## Follow-ups

- 구현 리뷰 단계에서는 command-selection 문구가 신규 사용자에게 과도한 선택 피로를 주지 않는지 확인이 필요합니다.
- `namba sync` 결과물 검토 시 README 첫 화면의 정보 밀도와 guide로 내려간 상세 정보의 경계가 실제 렌더링에서도 유지되는지 확인하는 것이 좋습니다.

## Recommendation

- 구현 진행을 권장합니다.
- 제품 기준선은 충분히 정리됐고, 이후 검증 포인트는 SPEC 보완이 아니라 실제 생성 결과가 이 구조를 흔들림 없이 재현하는지에 있습니다.
