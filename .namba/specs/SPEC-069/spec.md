# SPEC-069 — GPT-5.6 비용 인지형 단계 기반 모델 라우팅

## Problem

현재 NambaAI는 역할별 고정 런타임 프로필을 사용한다. 이 방식은 작업 단계, 위험도, repair 횟수, Sol 예산을 판단하지 못하고 `codex exec resume --last`에 의존한다. 결과적으로 비용·품질·세션 경계·읽기 전용 검토의 신뢰 계약을 보장할 수 없다.

## Goal

모든 Namba 관리 실행 경로에 순수하고 결정적인 `phase → role → risk → model` 턴 라우터를 도입한다. 기본은 `gpt-5.6-terra`, 엄격히 단순한 가역 구현은 `gpt-5.6-luna`, 초기 설계·아키텍처·고위험 판단은 읽기 전용 `gpt-5.6-sol`로 선택한다. 구현, 검증, 커밋, 푸시, 다음 메이저 릴리즈까지 이 계약을 완료한다.

## Scope

- Phase 상태기계: `intake`, `plan`, `design`, `architecture`, `implement`, `test`, `integration`, `review`, `repair`.
- 입력 `(phase, role, domain, risk flags, repair count, remaining Sol turns)`에서 `modelRoutingDecision`을 만드는 불변 policy registry와 순수 결정 함수.
- `plan`, plan-review, interactive `run`, direct fix/repair, queue, integrator, specialist, reviewer, parallel worker, managed custom agent에 공통 정책 적용.
- 모든 실제 `codex exec`/명시 resume에 `-m <explicit-model-id>`와 `-c model_reasoning_effort=<level>` 인코딩, capability probe 실패 시 실행 전 차단.
- 첫 JSONL에서 실제 thread UUID를 수집하고 같은 모델의 인접 phase만 UUID resume한다. 모델 경계 또는 UUID 부재는 structured checkpoint를 가진 새 세션이며 `--last`는 사용하지 않는다.
- optional `model-routing/v1` evidence/report 확장, dry-run turn plan, managed agent 권한·문서 생성, 정책 migration 및 legacy compatibility.

## Non-goals

- Codex 루트 대화의 모델을 저장소가 전환하는 기능.
- LLM을 사용한 라우팅, per-turn 사용자 model/effort override, hard token budget, `xhigh`·`max`·Pro 계열 자동 선택.
- 기존 사용자 작성 agent/config 또는 과거 SPEC/release/run-log의 수정.

## Safety and Compatibility Contract

- Luna 구현은 단일 subsystem, 명확한 파일·변환·결과, 가역성·결정적 acceptance, public API/schema/auth/deploy/build/dependency 변경 없음, 미해결 설계·리뷰 쟁점 없음의 **모든** 조건에서만 허용한다.
- Sol은 항상 read-only decision/checkpoint turn이다. 일반/solo/fix는 최대 1 Sol turn, team/integration plan-review는 최대 2, 어떤 모드에서도 Sol high는 최대 1이다. high는 `critical_risk && high_ambiguity && (cross_system || irreversible)`일 때만 허용한다.
- 필수 Sol 미지원은 `blocked_model_unavailable`이며 silent downgrade하지 않는다. 비필수 Sol은 Terra high fallback을 evidence에 기록한다. Terra는 Luna 단순 조건에서만 downgrade하고, Luna는 Terra로 승격한다.
- 새 init과 이 저장소는 `gpt-5.6-cost-balanced-v1`; 기존 config에 정책 키가 없으면 `legacy-static-v1`을 유지하고 doctor/status migration advisory를 보인다. adaptive 정책과 legacy global `model:`의 병존은 validation error다.
- 관측 불가 값은 추정하지 않고 `unavailable` 또는 `external_unobserved`로 기록한다.

## Release Contract

메이저 릴리즈 전 legacy 실행 보존, adaptive 새 init, policy conflict failure, dry-run/report/evidence compatibility, migration guide와 rollback 안내, regeneration/sync 멱등성, Go/Namba quality gates를 모두 증명한다.
