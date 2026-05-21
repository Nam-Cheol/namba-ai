# Design Review

- Status: clear
- Last Reviewed: 2026-05-21
- Reviewer: namba-designer
- Command Skill: `$namba-plan-design-review`
- Recommended Role: `namba-designer`

## Focus

- GitHub-safe Markdown 제약 안에서 README 및 가이드 문서의 첫 화면 스캔성, 섹션 구조 동등성, 접근성, 다국어 내비게이션 일관성을 먼저 고정한다.

- Evidence Status: not-applicable
- Gate Decision: not-applicable
- Approved Direction: 히어로, 배지, CTA, 명령 선택, 워크플로 표, 빠른 시작, 고급 접기 섹션, 문서 간 내비게이션을 하나의 반복 가능한 Markdown 문법으로 통일한다. 시각적 개성은 이미지, 제목 계층, 표 밀도, 링크 배열로 만들고, GitHub가 불안정하게 처리할 수 있는 커스텀 HTML 연출에는 의존하지 않는다.
- Banned Patterns: 커스텀 CSS, JavaScript, iframe, 정렬용 HTML 남용, 언어별 섹션 순서 불일치, 배지만 남고 설명이 없는 첫 화면, 이미지 안에만 핵심 정보를 넣는 구성, CTA를 표 셀 장식처럼 과도하게 포장하는 구성.
- Negative-First Contract: 지원이 불확실한 렌더링보다 plain Markdown 우선. 필수 정보는 항상 접히지 않은 본문에 존재해야 하며, `details`는 보조 설명으로만 사용한다.
- Default Library Fit: 표준 GitHub Markdown 요소만 사용한다. 이미지, 제목, 목록, 링크, 표, `details`/`summary`, 상대 경로 링크, 배지 이미지만 허용 범위로 본다.
- Context-Specific Bans And Replacements: "centered hero"는 HTML 정렬 구현이 아니라 문서 최상단 주도권 확보로 해석한다. 카드형 워크플로는 CSS 카드가 아니라 짧은 설명이 포함된 Markdown 표 또는 굵은 제목 기반 리스트로 대체한다.
- Reference-Driven Asset Manifest: 기존 README 히어로 이미지를 재사용하되, 각 언어 문서가 동일 자산 또는 대응 자산을 같은 위치에 노출해야 한다. 새 자산 추가는 필수 아님.
- Generated Image Plan: 없음. 본 SPEC의 시각 차별점은 새 비트맵 생성보다 구조적 일관성과 스캔성에서 만든다.
- Visual Grammar: 1스크린 안에서 "정체성 -> 신뢰 신호 -> 즉시 행동 -> 명령 선택 -> 다음 읽을 문서" 순서가 유지되어야 한다. 하위 문서는 같은 순서를 축약 또는 확장하되 문법은 유지한다.
- Generic-Section Proof: 가장 흔하게 평범해질 구간은 명령 선택 섹션이다. 단순 명령 나열 대신 "언제 이 명령을 쓰는가"를 한 줄로 붙여 사용 맥락을 드러내야 한다.
- Architecture Handoff: 렌더러는 언어별 문자열만 바꾸고 섹션 순서, 링크 슬롯, 배지 슬롯, CTA 슬롯, 접기 슬롯은 공통 템플릿으로 유지해야 한다.
- Violation-Check Plan: 생성 결과 테스트에서 금지 태그 부재, 상대 링크 유지, 필수 섹션 순서, 다국어 구조 동등성, 접기 섹션의 비핵심성, 이미지 대체 텍스트 존재를 검증해야 한다.
- Open Questions: 배지 개수가 4개를 넘을 때 첫 화면 밀도를 어떻게 제한할지 구현 단계에서 확정이 필요하다.
- Unresolved Questions: "card style"의 표현 강도를 어디까지 허용할지 문서 샘플 확인이 필요하지만 구현 차단 사유는 아니다.
- Design Review Axes: scanability, hierarchy, parity, accessibility, markdown-safety, translation consistency, trust signaling
- Keep / Fix / Quick Wins: 히어로 이미지와 배지 행 조합은 유지해도 된다. 가장 먼저 고쳐야 할 부분은 언어별 구조 드리프트 가능성이다. 빠른 성과는 CTA와 명령 선택을 첫 화면에서 명확히 분리하는 것이다.

## Review Checklist

- Art direction is clear and fits the task context.
- Palette temperature and undertone logic are not a primary concern because this surface is GitHub Markdown, but image and badge usage must still avoid visual clutter.
- Semantic sections match the content instead of defaulting to a README dump or badge wall.
- The Do-Not Design Contract is complete enough to prevent unsupported GitHub rendering and generic documentation fallback.
- Asset usage is restrained and justified; image reliance does not replace text orientation.
- Motion is not applicable.
- The most generic section risk is explicitly addressed in the command selection and workflow-summary sections.
- Anti-overcorrection guardrails hold: the docs should feel modern and scannable without becoming marketing-heavy or layout-fragile.

## Findings

- 첫 화면 요구사항은 충분히 타당하지만, "centered hero"를 literal 정렬 요구로 해석하면 GitHub-safe Markdown 제약과 충돌한다. 구현은 중앙 정렬 트릭이 아니라 상단 주도권이 명확한 히어로 블록으로 해석해야 한다.
- 배지 행, CTA 행, 명령 선택, 빠른 시작, 문서 내비게이션이 모두 첫 화면에 들어오면 쉽게 과밀해진다. 따라서 각 블록은 한 가지 질문만 해결해야 한다. 배지는 신뢰, CTA는 시작 행동, 명령 선택은 상황별 분기, 빠른 시작은 최소 경로, 내비게이션은 다음 문서 이동으로 역할을 분리해야 한다.
- 다국어 문서에서 가장 큰 디자인 리스크는 번역 품질보다 구조 비동기화다. 영어, 한국어, 일본어, 중국어 README와 가이드는 같은 섹션 순서, 같은 링크 슬롯, 같은 접기 위치를 공유해야 사용자가 언어를 바꿔도 방향 감각을 잃지 않는다.
- 접근성 측면에서 히어로 이미지와 배지는 장식이 아니라 보조 신호여야 한다. 핵심 설명, 시작 명령, 다음 문서 링크는 이미지 없이도 읽혀야 하고, 이미지 대체 텍스트와 링크 레이블은 언어별로 명확해야 한다.
- `details`는 고급 내용 정리에 적합하지만, 설치 시작 경로나 핵심 명령 비교처럼 신규 사용자가 즉시 알아야 할 정보까지 접어 넣으면 안 된다. 접기 섹션은 "추가 옵션", "심화 설명", "참고 규칙"만 담당해야 한다.
- 워크플로 표는 GitHub-safe 범위 안에서 유효하지만, 너무 많은 열을 쓰면 모바일과 좁은 뷰포트에서 급격히 읽기 어려워진다. "상황", "추천 명령", "다음 문서" 정도의 얕은 표나 짧은 리스트가 더 안전하다.
- 교차 문서 내비게이션은 장식용 링크 모음이 아니라 정보 구조의 골격이어야 한다. README, getting started, workflow guide, release, CI, security 문서 모두 현재 문서의 위치와 다음 읽을 곳을 상대 링크로 알려줘야 한다.

## Decisions

- 승인 방향은 "GitHub가 안정적으로 렌더링하는 Markdown만으로 시각 문법을 만든다"이다.
- 히어로는 이미지, 제목, 짧은 설명, CTA 인접 배치로 해결하고, 정렬 표현을 위해 커스텀 HTML에 기대지 않는다.
- 배지 행은 신뢰 신호만 담고, 행동 유도는 별도의 CTA 행에서 처리한다.
- 명령 선택 섹션은 제품 기능 소개가 아니라 사용자의 상황별 분기표로 설계한다.
- 빠른 시작은 가장 짧은 성공 경로만 노출하고, 고급 옵션은 `details` 아래로 내린다.
- 다국어 문서는 번역 독립체가 아니라 구조 공유 산출물로 취급하고, 섹션 parity를 테스트 계약으로 강제한다.
- 카드 스타일 요구는 CSS 모사 없이 Markdown 표, 굵은 제목, 짧은 설명, 명확한 링크 조합으로 해석한다.

## Follow-ups

- [non-blocking] 구현 단계에서 README와 각 언어 README의 첫 40~60줄 구조를 비교하는 계약 테스트를 추가해 구조 parity를 자동 검증할 것.
- [non-blocking] 배지 수 상한과 우선순위를 정해 첫 화면에서 줄바꿈이 과도하게 발생하지 않도록 할 것.
- [non-blocking] 히어로 이미지가 없는 저장소에서도 문서가 어색하지 않도록 텍스트 우선 폴백을 명시할 것.
- [non-blocking] 상대 링크가 루트 README와 로컬라이즈드 README에서 각각 올바르게 계산되는지 문서별 검증 케이스를 둘 것.

## Recommendation

- 진행 권고. 방향성은 명확하며, 핵심 디자인 리스크는 미적 감각보다 GitHub-safe 제약 안에서의 정보 계층과 다국어 구조 동등성이다. 구현은 시각 장식 확장보다 렌더 안전성, 스캔성, 접근성, 구조 parity를 우선순위로 두고 진행하는 것이 맞다.
