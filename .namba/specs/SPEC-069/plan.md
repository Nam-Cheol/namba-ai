# SPEC-069 Plan

## 1. Contract-first TDD

1. policy registry, phase/role/risk model, `modelRoutingDecision`, reason/rule/fallback/status vocabulary를 새 unit tests로 고정한다.
2. Luna eligibility, Sol read-only and budget, `xhigh/max` ban, required/optional Sol fallback, legacy/adaptive config conflict를 table-driven tests로 먼저 고정한다.

## 2. Execution architecture

1. 기존 implement-first request 배열을 phase-aware turn plan으로 교체하고 모든 execution entry point가 이를 만들게 한다.
2. syntax capability(`-m`, config override, JSONL)와 exact model/effective reasoning capability를 분리 probe한다. reasoning encoding 불가 시 실패한다.
3. typed JSONL parser로 첫 thread UUID를 저장한다. same-model adjacent phase + explicit UUID만 resume; 그 외 fresh session/checkpoint. `--last` 제거를 source and invocation tests로 증명한다.
4. invocation 별 write permission을 강제해 Sol advisor/reviewer는 read-only sandbox와 decision-only prompt, writer는 checkpoint consumer prompt를 쓴다.

## 3. Configuration, evidence, and operator experience

1. immutable registry mappings `efficient/standard/deep` and `model_routing_policy: gpt-5.6-cost-balanced-v1`을 config/generator에 추가한다.
2. `model-routing/v1` optional extension에 requested/effective model/effort, rule/reason, Sol state, fallback, thread/session strategy, observable usage를 추가한다.
3. dry-run에는 deterministic ordered turn-plan table/JSON (`phase, role, tier, model, effort, rule, reason, session, Sol remaining`)과 fallback/blocked recovery를 출력한다. report JSON/text에는 aggregate 및 observed/unobserved state를 추가한다.
4. regen은 managed agents와 `.namba/codex/model-routing.md`만 생성하고 user artifacts를 보존한다. generated prompts have explicit read-only vs writer contracts.

## 4. Matrix and regression validation

1. simple, general, cross-system design, irreversible security, repeated behavioral failure scenarios and direct-fix/queue/parallel/integrator/specialist/reviewer paths.
2. JSONL missing/malformed/long line, missing usage, unavailable model, policy conflict, checkpoint/new-session and no-resume cases.
3. old evidence/report/historical fixtures; regen and sync twice; past SPEC/release/log checksums unchanged.

## 5. Integration and release

1. Run `go test ./...`, gofmt check, `go vet ./...`, `namba eval`, `namba regen`, `namba sync`, and `scripts/quality.sh`.
2. Inspect generated diff, refresh project/sync artifacts, commit the scoped branch, push, prepare Korean PR handoff, and use guarded Namba release on clean main only after all release contract proof is present.
