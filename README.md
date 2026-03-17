# 🚇 NambaAI

NambaAI는 **MoAI의 워크플로 감각을 Codex 환경으로 이식한 프로젝트 부트스트랩 + 실행 오케스트레이터**입니다.
핵심 목표는 `project -> plan -> run -> sync` 흐름, SPEC 중심 실행, TDD/DDD 방법론, 그리고 Codex 친화적인 skill/agent 자산을 한 번에 구성하는 것입니다.

## ✨ 핵심 개념

- `namba init .`를 실행하면 빈 디렉토리에서도 바로 프로젝트 bootstrap을 시작할 수 있습니다.
- 초기화 과정은 **MoAI 스타일 wizard**를 Codex에 맞게 재구성한 흐름입니다.
- Claude Code 전용 자산은 그대로 복제하지 않고, **Codex-compatible scaffold**로 변환합니다.
- 초기화가 끝나면 Codex는 `AGENTS.md`, `.agents/skills/`, `.codex/agents/`, `.codex/config.toml`, `.namba/`를 기반으로 Namba workflow를 사용할 수 있습니다.

## 🧭 Claude Code → Codex 매핑

| Claude Code / MoAI | NambaAI / Codex |
| --- | --- |
| `CLAUDE.md` | `AGENTS.md` |
| `.claude/skills/*` | `.agents/skills/*` |
| Claude 호환용 skill 경로 | `.codex/skills/*` compatibility mirror |
| `.claude/agents/*.md` | `.codex/agents/*.toml` custom agent |
| hooks | 명시적 validation + `namba sync` + structured logs |
| custom slash command workflow | `$namba` skill + built-in Codex slash commands + `namba` CLI |
| Claude settings / statusline | `.codex/config.toml` + `.namba/codex/statusline.example.toml` |

## 📦 설치

일반 사용자는 **Go가 필요 없습니다.**
GitHub Release 바이너리를 내려받아 전역 `namba` 명령으로 설치합니다.

### Windows

```powershell
irm https://raw.githubusercontent.com/Nam-Cheol/namba-ai/main/install.ps1 | iex
```

설치 경로:
- `%LOCALAPPDATA%\Programs\NambaAI\bin\namba.exe`

### macOS / Linux

```bash
curl -fsSL https://raw.githubusercontent.com/Nam-Cheol/namba-ai/main/install.sh | sh
```

설치 경로:
- `~/.local/bin/namba`

### 특정 버전 설치

```powershell
$env:NAMBA_VERSION = 'v0.1.0'
irm https://raw.githubusercontent.com/Nam-Cheol/namba-ai/main/install.ps1 | iex
```

```bash
NAMBA_VERSION=v0.1.0 curl -fsSL https://raw.githubusercontent.com/Nam-Cheol/namba-ai/main/install.sh | sh
```

### 아직 Release가 없을 때

Release가 아직 발행되지 않았다면 설치 스크립트는 `releases/latest`에서 404를 반환합니다.
이 경우에는 Go가 설치된 환경에서 아래처럼 소스 기준으로 설치합니다.

```powershell
go install github.com/Nam-Cheol/namba-ai/cmd/namba@main
```

```bash
go install github.com/Nam-Cheol/namba-ai/cmd/namba@main
```

## 🗑️ 제거 방법

NambaAI 제거는 `전역 명령 제거`와 `프로젝트 자산 제거`를 구분해서 보면 됩니다.

### 전역 `namba` 명령 제거

#### Windows

기본 설치 경로를 사용했다면 아래 파일을 삭제하면 됩니다.

- `%LOCALAPPDATA%\Programs\NambaAI\bin\namba.exe`

PowerShell 예시:

```powershell
Remove-Item "$env:LOCALAPPDATA\\Programs\\NambaAI\\bin\\namba.exe" -Force
```

추가로 사용자 `Path`에 남아 있는 설치 경로를 제거하면 깔끔합니다.

- 제거 대상: `%LOCALAPPDATA%\Programs\NambaAI\bin`

#### macOS / Linux

기본 설치 경로를 사용했다면 아래 파일을 삭제하면 됩니다.

- `~/.local/bin/namba`

예시:

```bash
rm -f ~/.local/bin/namba
```

그리고 셸 프로필에 추가된 PATH 줄을 제거합니다.

- `~/.profile` 또는 `~/.zshrc`
- 제거 대상 줄:

```bash
export PATH="$HOME/.local/bin:$PATH"
```

### 프로젝트에서 NambaAI 자산 제거

`namba init .`로 초기화한 프로젝트에서 NambaAI 자산만 제거하려면 아래 파일과 디렉토리를 삭제하면 됩니다.

- `AGENTS.md`
- `.agents/`
- `.codex/skills/`
- `.codex/agents/`
- `.codex/config.toml`
- `.namba/`

주의:
- 이미 직접 수정한 `AGENTS.md`나 `.codex/` 아래 다른 사용자 파일이 있다면 함께 삭제되지 않도록 확인 후 제거해야 합니다.
- Git 저장소라면 삭제 전에 `git status`로 추적 중인 변경을 먼저 확인하는 편이 안전합니다.

## ⚙️ 필수 전제조건

- `git`
- `codex` CLI
- 전역 `namba` 명령

확인 예시:

```powershell
git --version
cmd /c codex --version
namba doctor
```

## 🚀 빠른 시작

### 1. 새 프로젝트 디렉토리 준비

```powershell
mkdir C:\project\example
cd C:\project\example
```

### 2. NambaAI 초기화

```powershell
namba init .
```

기본적으로 interactive terminal에서는 wizard가 실행됩니다.
여기서 다음 항목을 선택할 수 있습니다.

- 프로젝트 이름
- 개발 방법론: `TDD` 또는 `DDD`
- 언어 / 프레임워크
- 대화 언어 / 문서 언어 / 코드 코멘트 언어
- Codex agent 모드: `single` / `multi`
- Git 자동화 모드: `manual` / `personal` / `team`
- Git provider / username
- status line preset
- 사용자 이름

자동화 환경에서는 `--yes`와 flags를 사용하면 됩니다.

```powershell
namba init . --yes --name example --mode tdd --agent-mode multi --git-mode team --git-provider github --git-username alice
```

### 3. 초기화 후 생성되는 주요 자산

- `AGENTS.md`
- `.agents/skills/`
- `.codex/skills/`
- `.codex/agents/`
- `.codex/config.toml`
- `.namba/config/sections/*.yaml`
- `.namba/project/*`
- `.namba/specs/*`
- `.namba/logs/*`

### 4. Codex에서 사용하기

Codex를 해당 프로젝트 디렉토리에서 열면 됩니다.

```powershell
cd C:\project\example
codex
```

그 다음 Codex 안에서는 이렇게 사용합니다.

- `$namba`를 직접 호출
- `namba project`
- `namba update`
- `namba plan "로그인 기능 추가"`
- `namba run SPEC-001`
- `namba sync`

중요한 점:
- interactive Codex 세션에서 `namba run SPEC-XXX`는 **Codex가 현재 세션에서 SPEC를 직접 수행하라**는 뜻입니다.
- 비대화형 자동 실행이 필요하면 standalone `namba run SPEC-XXX` CLI를 사용할 수 있습니다.
## Workflow Reference

```text
namba project
namba update   # only when template-generated Codex assets need regeneration
namba plan "work description"
namba run SPEC-XXX
namba sync
```

`namba update` and `namba sync` are different commands. You do not run both on every iteration unless the change actually requires both refresh paths.

### Command List

- `namba init [path] [--yes] [--name NAME] [--mode tdd|ddd] [--project-type new|existing]`
- `namba doctor`
- `namba status`
- `namba project`
- `namba update`
- `namba plan "<description>"`
- `namba fix "<description>"`
- `namba run SPEC-XXX [--parallel] [--dry-run]`
- `namba sync`
- `namba release [--bump patch|minor|major] [--version vX.Y.Z] [--push] [--remote origin]`
- `namba worktree <new|list|remove|clean>`

## Update vs Sync

- `namba update`: regenerate `AGENTS.md`, `.agents/skills/`, `.codex/skills/`, `.codex/agents/`, and `.codex/config.toml` from `.namba/config/sections/*.yaml`.
- `namba sync`: refresh `.namba/project/*` docs, codemaps, change summary, PR checklist, and release notes/checklists.
- If you edit template-generated assets directly, a later `namba update` can overwrite them. Durable settings belong in `.namba/config/sections/*.yaml`.

## Parallel Run Policy

`namba run SPEC-XXX --parallel` below refers to the standalone CLI runner path. In an interactive Codex session, `namba run SPEC-XXX` still means execute the SPEC directly in the current session.

- `--parallel` uses git worktree fan-out/fan-in execution.
- The current implementation splits work into at most three worker worktrees.
- Merge begins only after every worker passes execution and validation.
- Any execution, validation, or merge failure blocks fan-in merge and preserves worker worktrees and branches for inspection.
- Successful runs remove temporary worktrees, delete worker branches, and run `git worktree prune`.
- `--dry-run` prepares worker request docs and the parallel report only. It does not run the runner, merge, or cleanup.
- Reports are written under `.namba/logs/runs/<spec>-parallel.json`.

## Release Flow

- `namba release` requires a git repository, the `main` branch, and a clean working tree.
- Validators from `.namba/config/sections/quality.yaml` run before the release tag is created.
- Without an explicit version, `namba release` calculates the next `patch` tag. Use `--bump minor|major` or `--version vX.Y.Z` when needed.
- `--push` pushes both `main` and the new tag to the selected remote. Without it, Namba prints the next push commands only.
- The GitHub Release workflow publishes six platform archives plus `checksums.txt`.

## 🧪 예시: 빈 디렉토리에서 시작하기

```powershell
mkdir C:\project\example
cd C:\project\example
namba init .
namba project
namba plan "health check endpoint 추가"
namba run SPEC-001
namba sync
```

## 🧱 생성되는 설정 파일

### `.namba/config/sections/project.yaml`

```yaml
name: example
language: go
framework: cobra
created_at: 2026-03-15T12:00:00+09:00
```

### `.namba/config/sections/quality.yaml`

```yaml
development_mode: tdd
test_command: go test ./...
lint_command: gofmt -l "cmd" "internal" "namba_test.go"
typecheck_command: go vet ./...
```

### `.namba/config/sections/git-strategy.yaml`

```yaml
git_mode: team
git_provider: github
git_username: alice
gitlab_instance_url: https://gitlab.com
store_tokens: false
```

### `.namba/config/sections/codex.yaml`

```yaml
agent_mode: multi
status_line_preset: namba
repo_skills_path: .agents/skills
compat_skills_path: .codex/skills
repo_agents_path: .codex/agents
```

## 🤖 Codex agent 자산

NambaAI는 Claude의 sub-agent 개념을 그대로 복제하지 않습니다.
대신 Codex에서 읽기 쉬운 **custom agent TOML**을 생성합니다.

- `.codex/agents/namba-planner.toml`
- `.codex/agents/namba-implementer.toml`
- `.codex/agents/namba-reviewer.toml`

이 파일들은 Codex multi-agent delegation 시 역할 프롬프트의 기준점으로 사용합니다.

## 📚 status line

NambaAI는 `.namba/codex/statusline.example.toml`을 생성합니다.
원하면 이 내용을 `~/.codex/config.toml`에 병합해서 Namba에 맞는 status line을 사용할 수 있습니다.

## 🔐 보안 원칙

- GitHub/GitLab token은 scaffold에 저장하지 않습니다.
- 인증은 `gh auth login` 또는 `glab auth login`으로 처리합니다.
- 실행 로그는 `.namba/logs/` 아래에 남습니다.

## 🧾 개발자용 빌드

NambaAI 자체를 개발할 때만 Go가 필요합니다.

```bash
git clone https://github.com/Nam-Cheol/namba-ai.git
cd namba-ai
go build -o namba ./cmd/namba
```

## 📌 현재 상태

- Codex-native repo scaffold 지원
- MoAI-style init wizard 지원
- repo-local skills / compatibility mirror 지원
- repo-local Codex custom agent(.toml) 지원
- structured execution logs 지원
- README now documents update, release, and parallel worktree merge/cleanup policy
