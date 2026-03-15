# 🚉 NambaAI

NambaAI는 Codex 중심의 SPEC 기반 개발 오케스트레이터입니다.
작업을 바로 코드로 밀어 넣지 않고 `project -> plan -> run -> sync` 흐름으로 정리해서 실행합니다.

핵심 목표는 다음과 같습니다.

- 📦 작업을 `.namba/specs/` 아래 SPEC 문서로 남기기
- 🧪 실행 뒤 품질 게이트를 명시적으로 통과시키기
- 🤖 Codex를 실행 엔진으로 사용해 자동화된 구현 흐름 만들기
- 🌲 필요할 때 `git worktree` 기반 병렬 실행으로 확장하기

## ✨ 주요 기능

- `namba project`: 프로젝트 구조, 제품 문서, codemap 갱신
- `namba plan`: 작업 설명을 SPEC 패키지로 변환
- `namba run`: SPEC를 읽고 Codex로 실행
- `namba sync`: 변경 요약, 체크리스트, 문서 산출물 정리
- `approval_mode`, `sandbox_mode`를 실제 `codex exec` 인자로 반영
- 실행 로그를 `request.md`, `result.txt`, `execution.json`, `validation.json`으로 저장

## 🚀 설치

일반 사용자는 Go가 필요 없습니다.
최신 GitHub Release 바이너리를 받아서 사용자 PATH에 등록하는 설치 스크립트를 기본 경로로 제공합니다.

### Windows

```powershell
irm https://raw.githubusercontent.com/Nam-Cheol/namba-ai/main/install.ps1 | iex
```

설치 위치:
- `%LOCALAPPDATA%\Programs\NambaAI\bin\namba.exe`

설치가 끝나면 새 터미널을 열고 바로 전역 명령으로 실행합니다.

```powershell
namba doctor
namba status
```

### macOS / Linux

```bash
curl -fsSL https://raw.githubusercontent.com/Nam-Cheol/namba-ai/main/install.sh | sh
```

설치 위치:
- `~/.local/bin/namba`

설치 스크립트가 사용자 PATH를 갱신합니다. 새 셸을 열거나 아래를 실행하면 됩니다.

```bash
exec $SHELL -l
namba doctor
namba status
```

### 특정 버전 설치

```powershell
$env:NAMBA_VERSION = 'v0.1.0'
irm https://raw.githubusercontent.com/Nam-Cheol/namba-ai/main/install.ps1 | iex
```

```bash
NAMBA_VERSION=v0.1.0 curl -fsSL https://raw.githubusercontent.com/Nam-Cheol/namba-ai/main/install.sh | sh
```

## 🛠 개발자용 설치

NambaAI 자체를 개발하거나 로컬에서 직접 빌드하고 싶을 때만 Go가 필요합니다.

```bash
git clone https://github.com/Nam-Cheol/namba-ai.git
cd namba-ai
go build -o namba ./cmd/namba
```

직접 빌드한 경우에도 PATH에 등록하면 `namba` 전역 명령으로 사용할 수 있습니다.

## 📋 필수 전제조건

- `git`
- `codex` CLI
- `namba` 실행 파일 또는 설치 스크립트로 설치된 전역 명령

Codex 확인 예시:

```powershell
cmd /c codex --version
```

PowerShell 실행 정책 때문에 `codex`를 직접 입력하면 `codex.ps1`이 막힐 수 있습니다.
NambaAI는 내부적으로 Windows에서 `cmd /c codex` 경로를 사용하므로 일반적인 `namba run`은 그대로 동작합니다.

## ⚡ 빠른 시작

현재 저장소에서 바로 쓰는 흐름은 아래와 같습니다.

```bash
namba doctor
namba status
namba project
namba plan "README 개선 작업"
namba run SPEC-004 --dry-run
namba run SPEC-004
namba sync
```

다른 저장소에서는 이렇게 시작합니다.

```bash
mkdir my-project
cd my-project
namba init .
namba project
namba plan "사용자 인증 플로우 추가"
namba run SPEC-001 --dry-run
namba run SPEC-001
namba sync
```

## 🧭 기본 워크플로

1. `namba project`
   현재 저장소를 읽고 `.namba/project/*`와 codemap을 갱신합니다.
2. `namba plan "<작업 설명>"`
   `.namba/specs/SPEC-XXX/` 아래에 `spec.md`, `plan.md`, `acceptance.md`를 생성합니다.
3. `namba run SPEC-XXX`
   SPEC를 읽고 Codex로 실행합니다.
4. `namba sync`
   변경 요약, PR 체크리스트, 문서 산출물을 정리합니다.

## 🧩 명령어

- `namba init [path]`
- `namba doctor`
- `namba status`
- `namba project`
- `namba plan "<description>"`
- `namba run SPEC-XXX [--parallel] [--dry-run]`
- `namba sync`
- `namba worktree <new|list|remove|clean>`

## ⚙ 실행 설정

시스템 설정은 `.namba/config/sections/system.yaml`에서 관리합니다.

```yaml
runner: codex
approval_mode: on-request
sandbox_mode: workspace-write
```

품질 게이트는 `.namba/config/sections/quality.yaml`에서 관리합니다.

```yaml
development_mode: tdd
test_command: go test ./...
lint_command: gofmt -l "cmd" "internal" "namba_test.go"
typecheck_command: go vet ./...
```

## 🧾 로그와 산출물

실행 결과는 보통 아래 위치에 남습니다.

- `.namba/logs/runs/<spec>-request.md`
- `.namba/logs/runs/<spec>-result.txt`
- `.namba/logs/runs/<spec>-execution.json`
- `.namba/logs/runs/<spec>-validation.json`
- `.namba/project/change-summary.md`
- `.namba/project/pr-checklist.md`

## 🌲 병렬 실행 주의사항

병렬 실행은 아래처럼 사용할 수 있습니다.

```bash
namba run SPEC-003 --parallel
```

현재는 `git worktree` 기반 fan-out/fan-in 뼈대까지 구현되어 있습니다.
실패 정책과 merge gate는 계속 고도화 중이므로, 지금 시점에서는 **serial run을 기본 경로로 사용하는 편이 가장 안전합니다.**

## 🌐 UTF-8 출력

NambaAI는 생성 문서를 UTF-8로 기록하고, Windows 콘솔에서는 출력 코드 페이지를 UTF-8(65001)로 고정합니다.
README, 로그, JSON 산출물, CLI 메시지를 같은 인코딩 기준으로 맞추기 위한 설정입니다.

## 🗂 저장소 구조

- `cmd/namba`: CLI 진입점
- `internal/namba`: 워크플로, runner, validation 구현
- `.namba`: 프로젝트 상태, SPEC, 문서, 로그
- `.codex/skills`: Codex 세션용 로컬 스킬
- `install.ps1`, `install.sh`: 릴리스 바이너리 설치 스크립트
- `.github/workflows/release.yml`: 다중 플랫폼 릴리스 패키징

## 🛣 현재 로드맵

- ✅ SPEC 기반 실행 코어
- ✅ Runner abstraction
- ✅ approval / sandbox 설정 반영
- ✅ GitHub Release 기반 설치 스크립트
- 🟡 parallel worktree failure policy / merge gate
- 🟡 cleanup policy와 fan-in 안정화
