# SPEC-021

## Problem

NambaAI의 README와 workflow guide는 현재 "처음 보는 사용자가 바로 쓸 수 있는 onboarding surface" 역할을 충분히 하지 못하고 있다.

현재 저장소 기준으로 확인되는 문제:

- checked-in `README*.md`는 `namba harness`를 아직 일관되게 소개하지 못한다.
- root README의 quick summary는 `namba plan`과 `namba fix` 중심으로만 설명되어 있어, `namba project`, `namba harness`, `namba run`, `namba sync`, `namba pr`, `namba land`의 역할 차이가 초심자에게 즉시 드러나지 않는다.
- command skill 섹션은 일부 command-entry skill의 존재나 목적을 충분히 설명하지 못한다.
- workflow guide도 planning surface와 review readiness 문구가 최신 renderer 의도와 완전히 맞지 않는다.
- renderer source에는 더 나은 방향의 문구가 일부 반영되어 있지만, 실제 generated docs와 완전히 동기화되지 않아 정보 구조와 checked-in docs 사이에 drift가 있다.

결과적으로 초심자는 아래 질문에 즉시 답을 얻기 어렵다.

- "처음에는 `namba project`, `namba plan`, `namba harness`, `namba fix` 중 무엇을 써야 하는가?"
- "`$namba`, `$namba-plan`, `$namba-harness`, `$namba-run`, `$namba-pr`는 각각 어떤 상황에서 쓰는가?"
- "설치 후 첫 작업을 시작해서 PR 머지까지 가는 기본 경로는 무엇인가?"

## Goal

README와 workflow guide를 "처음 보는 사용자를 위한 빠른 진입 문서"로 고도화하고, renderer source와 checked-in generated docs를 일치시킨다.

## Context

- Project: namba-ai
- Project type: existing
- Language: go
- Mode: tdd
- Work type: plan

## Target Reader

- NambaAI를 처음 접한 사용자
- Codex에서 어떤 command/skill을 먼저 써야 하는지 빠르게 판단해야 하는 사용자
- `namba init` 이후 첫 SPEC 생성부터 PR handoff/merge까지의 기본 경로를 이해하려는 사용자

## Desired Outcome

- root README만 읽어도 NambaAI의 전체 흐름을 빠르게 이해할 수 있다.
- 초심자가 `namba project`, `namba plan`, `namba harness`, `namba fix`, `namba run`, `namba sync`, `namba pr`, `namba land`의 차이를 즉시 구분할 수 있다.
- command-entry skill 섹션은 각 skill이 "무엇을 위한 것인지"를 user-facing 언어로 설명한다.
- quick start는 설치 후 첫 작업을 시작하는 대표 경로를 보여주며, 일반 feature planning과 harness planning의 차이를 놓치지 않는다.
- workflow guide는 lifecycle command, planning command, execution mode, review readiness, PR/merge flow를 구조적으로 설명한다.
- generated README/workflow docs는 renderer source와 동기화되어 stale copy가 남지 않는다.

## Primary Risk To Control

- 이번 작업의 가장 큰 위험은 renderer source와 checked-in generated docs가 다시 분리되는 것이다.
- root README와 workflow guide가 서로 비슷한 설명을 중복해서 가지면, 이후 한쪽만 수정되고 다른 쪽이 stale해질 가능성이 높다.
- localized generated docs까지 포함한 문서 세트가 renderer 변화와 함께 갱신되지 않으면 `namba harness` 같은 최신 surface가 다시 일부 문서에서 누락될 수 있다.
- 따라서 이번 SPEC은 문장 개선만이 아니라 "어디를 source of truth로 둘지"와 "drift를 어떻게 테스트로 잡을지"까지 함께 해결해야 한다.

## Implementation Guardrails

- `internal/namba/readme.go`를 README/workflow guide 계열 문서의 단일 source of truth로 유지한다.
- checked-in generated docs는 renderer 수정 후 `namba sync`로만 갱신한다.
- root README와 workflow guide는 역할을 분리해 중복 설명을 최소화한다.
- regression test는 전체 장문 일치보다 command selection guidance, skill purpose guidance, `namba harness` 노출, section structure 같은 stable anchor를 고정한다.
- localized variants도 같은 정보 구조와 command surface를 유지해야 한다.

## Information Architecture Priority

- 정보 구조와 가독성 개선을 우선한다.
- "명령 목록 나열"보다 "언제 무엇을 쓰는지"가 먼저 보이도록 구성한다.
- 각 command/skill 설명은 초심자 기준의 decision support 문장으로 작성한다.
- 세부 구현 원리보다 실제 사용 흐름과 command 선택 기준을 먼저 제시한다.

## Scope

- `internal/namba/readme.go` 기반 README/workflow guide renderer 고도화
- `README.md`, `README.ko.md`, `README.ja.md`, `README.zh.md` 동기화
- `docs/workflow-guide.md`, `docs/workflow-guide.ko.md`, `docs/workflow-guide.ja.md`, `docs/workflow-guide.zh.md` 동기화
- command/skill purpose 설명 고도화
- quick start와 command selection guidance 고도화
- renderer output 관련 회귀 테스트 보강

## Non-Goals

- CLI command semantics 자체를 바꾸는 작업은 이번 SPEC의 주된 목적이 아니다.
- onboarding/worflow guide 외 다른 장문 문서를 전면 재작성하지 않는다.
- visual redesign 자체보다 정보 구조, 용도 설명, generated doc consistency를 우선한다.
- generated docs를 임시로 수작업 patch해서 drift를 숨기는 방식은 허용하지 않는다.
