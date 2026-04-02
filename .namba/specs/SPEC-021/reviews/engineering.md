# Engineering Review

- Status: approved
- Last Reviewed: 2026-04-02
- Reviewer: Codex
- Command Skill: `$namba-plan-eng-review`
- Recommended Role: `namba-planner`

## Focus

- Lock architecture, sequencing, failure modes, trust boundaries, and validation strategy before execution starts.

## Findings

- 구현 경계는 비교적 명확합니다. 이번 SPEC의 주된 수정 지점은 `internal/namba/readme.go`와 그 산출물, 그리고 해당 산출물을 고정하는 테스트입니다. CLI semantics를 바꾸는 작업이 아니므로 영향 범위를 문서 renderer와 generated docs로 제한하기 좋습니다.
- 현재 저장소에는 renderer source와 checked-in generated docs 사이의 drift가 실제로 존재합니다. 따라서 구현의 source of truth는 반드시 renderer로 두고, `README*.md`와 `docs/workflow-guide*.md`는 `namba sync` 결과로만 갱신해야 합니다.
- 테스트 전략은 보강이 필요합니다. 현재 `internal/namba/readme_sync_test.go`에는 일부 `namba harness` 관련 고정점이 있지만, root README를 first-session onboarding surface로 재구성하는 수준의 정보 구조 변화까지 충분히 보호하지는 못합니다.
- 다국어 문서까지 함께 관리하는 구조상, 테스트는 문단 전체의 완전 일치보다 안정적인 sentinel assertion 중심으로 가는 편이 안전합니다. 예를 들어 section heading, command-choice guidance, command/skill mapping, workflow guide 분리 구조 같은 안정적인 anchor를 확인해야 합니다.
- 구현 시 가장 큰 공학적 리스크는 README를 너무 많은 내용을 담는 단일 문서로 만들어 중복과 drift를 키우는 것입니다. root README는 빠른 진입과 command 선택을 맡고, workflow guide는 확장 설명을 맡도록 역할을 분리해야 유지보수 비용이 낮습니다.

## Decisions

- 진행합니다. `internal/namba/readme.go`를 단일 source of truth로 유지하고, generated docs는 `namba sync`로만 갱신합니다.
- 구현 범위는 root README, workflow guide, localized variants, 그리고 이를 보호하는 renderer/doc regression test까지로 고정합니다.
- root README와 workflow guide는 역할을 분리합니다. root README는 quick onboarding과 command 선택, workflow guide는 lifecycle/planning/execution/review/merge 구조 설명을 담당합니다.
- 테스트는 user-facing prose 전체를 하드코딩하지 말고, 구조와 핵심 guidance를 보여주는 안정적인 문구 단위로 고정합니다.

## Follow-ups

- `internal/namba/readme.go` 수정 후 `namba sync`로 generated docs를 다시 렌더링하세요. generated output만 직접 patch 하는 경로는 피해야 합니다.
- `internal/namba/readme_sync_test.go`에 아래 종류의 assertion을 추가하거나 강화하세요.
- root README가 `namba project`, `namba plan`, `namba harness`, `namba fix`의 선택 기준을 구분해 설명하는지
- command skill section이 `$namba-harness`를 포함해 주요 command-entry skill의 목적을 설명하는지
- workflow guide가 lifecycle command, planning command, execution mode를 구조적으로 분리하는지
- localized outputs가 같은 정보 구조를 유지하는지
- 필요하면 `internal/namba/update_command_test.go`에서도 regen/sync 이후 generated doc contract를 보강하세요.
- section heading과 짧은 핵심 문장을 stable anchor로 사용하고, 장문 문단 전체를 exact match로 고정하는 방식은 피하세요.

## Recommendation

- Advisory recommendation: approved. Proceed, but keep the implementation tightly scoped to renderer-driven doc generation, stable regression anchors, and a clean separation between quick onboarding content and detailed workflow reference content.
