# Engineering Review

- Status: approved
- Last Reviewed: 2026-05-21
- Reviewer: namba-planner
- Command Skill: `$namba-plan-eng-review`
- Recommended Role: `namba-planner`

## Focus

- 이전 `needs-revision` 지적 사항이 실제로 해소되었는지 다시 확인하고, 구현 착수 기준으로 충분한지 판단한다.

## Findings

- `plan.md`가 이전보다 충분히 구체화되었습니다. 공용 Markdown primitive 정의 -> 루트 README 우선 적용 -> guide 문법 확장 -> locale parity 정리 -> config 보강 -> 집중/전체 검증 -> `namba sync` 반영 순서가 명시되어 있어, 이전 리뷰에서 지적한 실행 순서 공백은 해소됐습니다.
- 증분 렌더러 전략도 문서에 고정되었습니다. `spec.md`와 `plan.md` 모두 전면 재작성 대신 renderer/helper 추출과 root-first 적용을 전제로 하고 있어, `internal/namba/readme.go`와 `buildReadmeOutputs`의 회귀 범위를 단계적으로 통제할 수 있습니다.
- 결정적 출력 보호 장치가 명확해졌습니다. `spec.md`의 Constraints가 시간/환경/네트워크/locale drift를 금지하고, `acceptance.md`와 `plan.md`가 deterministic output test와 localized structural parity assertions를 요구하므로 이전 핵심 리스크가 문서상 관리 가능합니다.
- 테스트 범위는 기존 계약 확장 방식으로 정리되었습니다. helper-level 검증과 output/sync 계약 검증을 나누고, hero, badge, CTA, command selection, quick-start, workflow/navigation table, advanced details, relative links, unsupported HTML 부재, locale parity, deterministic output까지 acceptance에 포함해 이전 테스트 전략 부족 이슈가 해소됐습니다.
- `namba sync` 계약도 충분히 명시됐습니다. renderer/config가 source of truth라는 점, managed-output replacement 및 no-op stability 기대치를 유지해야 한다는 점, 구현 후 `namba sync`로 README 및 guide 산출물을 재생성해야 한다는 점이 `spec.md`, `plan.md`, `acceptance.md` 전반에 반영돼 있습니다.
- 프런트엔드 자산 blocker는 해소됐습니다. `frontend-brief.md`가 이 작업을 `frontend-minor`이자 `not-applicable` 게이트로 재분류했고, 기존 hero asset 재사용만 허용하며 새 이미지 생성과 브라우저 검증을 범위 밖으로 못 박고 있어 이전의 asset/reference 공백이 구현 blocker로 남아 있지 않습니다.
- 현재 문서 패키지 기준으로 남은 이슈는 구현 단계의 규율 준수입니다. 생성 결과를 직접 손대지 않고 renderer/tests/sync 경로에서만 변경을 수렴하면 엔지니어링 측면에서 추가적인 사전 blocker는 보이지 않습니다.

## Decisions

- 이전 `needs-revision` 사유였던 여섯 항목은 모두 해소된 것으로 판단합니다.
- 구현은 문서에 적힌 순서대로 renderer helper/section 단위의 증분 변경으로 진행해도 안전합니다.
- 엔지니어링 관점에서 SPEC-060은 구현 착수 가능 상태입니다. 다만 구현 중에도 GitHub-safe allowlist, locale parity, deterministic sync 계약은 비타협 항목으로 유지해야 합니다.

## Follow-ups

- 구현 시 생성 파일 직접 수정이 아니라 `internal/namba` renderer와 관련 테스트, sync 경로만 변경 대상으로 유지하세요.
- 검증 순서는 `go test ./internal/namba -run 'Readme|Sync|Frontend|Contract'` 우선, 이후 `go test ./...`, 마지막으로 `namba sync` 결과 드리프트 확인 순으로 고정하는 편이 좋습니다.
- 이번 재검토에서는 소유권 제약 때문에 `readiness.md`를 갱신하지 않았습니다. 후속 워크플로에서 리뷰 요약 아티팩트는 별도로 동기화하는 것이 맞습니다.

## Recommendation

- 수정 반영 후 구현 진행을 권고합니다. 이전 보완 요청이 문서 수준에서 해소됐고, 남은 리스크는 구현 중 계약 위반 여부에 가깝습니다.
