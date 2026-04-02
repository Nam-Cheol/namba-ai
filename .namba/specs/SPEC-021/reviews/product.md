# Product Review

- Status: approved
- Last Reviewed: 2026-04-02
- Reviewer: Codex
- Command Skill: `$namba-plan-pm-review`
- Recommended Role: `namba-product-manager`

## Focus

- Challenge the problem framing, scope, user value, and acceptance bar before implementation starts.

## Findings

- 문제 정의는 적절합니다. 현재 README와 workflow guide는 "무엇이 있는가"는 어느 정도 보여주지만, "처음 보는 사용자가 지금 무엇을 해야 하는가"를 충분히 안내하지 못합니다. 특히 `namba project`, `namba plan`, `namba harness`, `namba fix`의 선택 기준이 빠르게 보이지 않는 점은 onboarding 품질에 직접적인 영향을 줍니다.
- 이번 SPEC가 정보 구조와 가독성을 우선순위로 둔 것은 맞는 방향입니다. 지금 저장소의 가장 큰 문서 문제는 기능 부족보다도 진입 순서와 command 선택 기준이 흐릿한 점이기 때문에, visual polish보다 onboarding 구조 개선이 먼저입니다.
- command-entry skill 설명을 강화하는 범위 설정도 적절합니다. 초심자는 CLI command와 `$skill` entry point의 관계를 한 번에 이해하기 어렵기 때문에, 각 skill이 "언제 쓰는가"를 user-facing 문장으로 설명하는 것은 문서 가치가 큽니다.
- acceptance는 전반적으로 좋아졌지만, 구현 시 README가 지나치게 길고 반복적으로 되지 않도록 주의가 필요합니다. 한 페이지에 모든 것을 넣기보다, root README는 빠른 선택과 기본 흐름에 집중하고 상세 내용은 workflow guide가 받치는 구조가 중요합니다.

## Decisions

- 진행합니다. `SPEC-021`은 문서 polish가 아니라 onboarding UX 개선으로 취급합니다.
- root README의 최우선 목적은 "첫 5분 안에 사용자가 다음 command를 고르게 하는 것"으로 둡니다.
- workflow guide는 README의 확장 설명서 역할을 맡고, lifecycle command / planning command / execution mode / review readiness / merge flow를 구조적으로 분리합니다.
- command/skill 설명은 단순한 이름 나열보다 "언제 선택하는지"를 먼저 쓰는 방식으로 구현합니다.

## Follow-ups

- README에 최소 하나의 명시적 command 선택 가이드를 포함하세요. 예를 들면 "repo 이해는 `namba project`", "기능 계획은 `namba plan`", "재사용 가능한 agent/skill 설계는 `namba harness`", "직접 수리는 `namba fix`"처럼 빠른 decision support가 필요합니다.
- command skill 섹션은 각 항목을 한 줄 정의가 아니라 사용 상황 중심으로 쓰세요. 초심자가 `$namba`와 `$namba-plan`의 차이를 읽고 바로 이해할 수 있어야 합니다.
- quick start에는 feature planning 경로뿐 아니라 harness planning의 존재도 드러나야 합니다. 다만 예시를 과도하게 늘리기보다 대표 흐름과 대체 경로를 명확히 구분하는 편이 좋습니다.
- localized README/workflow guide도 같은 정보 구조를 유지하는지 확인해야 합니다. 영어만 좋아지고 다른 언어 문서가 뒤처지면 onboarding 품질이 다시 깨집니다.

## Recommendation

- Advisory recommendation: approved. Proceed, but keep the implementation tightly centered on first-time-user onboarding, command-choice clarity, and non-redundant information architecture.
