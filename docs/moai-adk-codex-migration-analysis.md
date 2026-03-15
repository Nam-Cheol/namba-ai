# MoAI-ADK -> Codex Migration Analysis

Date: 2026-03-15
Analyst: Codex

## 1. Objective

Primary goal: migrate the operating model of `modu-ai/moai-adk` from a Claude Code-centered environment to a Codex-centered environment without losing the core harness engineering approach.

This is not a simple prompt conversion task.
`moai-adk` is a provider-specific development operating system composed of:

1. Claude-native control surfaces
2. MoAI Go runtime and template deployment
3. Project assets and workflow state

The migration strategy should preserve layers 2 and 3 as much as possible, and replace layer 1 with Codex-native equivalents.

## 2. Executive Summary

`moai-adk` is structurally closer to an AI development platform than to a prompt pack.
Its real core is a Go binary that deploys and updates a project-local operating environment:

- `CLAUDE.md`
- `.claude/agents`
- `.claude/commands`
- `.claude/hooks`
- `.claude/skills`
- `.claude/rules`
- `.claude/settings.json`
- `.mcp.json`
- `.moai/config/sections/*`
- `.moai/specs/*`
- `.moai/project/codemaps/*`

The most important conclusion is this:

- The MoAI core workflow is portable.
- The Claude event model is not portable 1:1.
- Codex can absorb a large part of the system, but not by direct file-for-file transplantation.

The right direction is a `provider abstraction` approach:

1. Keep `.moai` as the vendor-neutral workflow/state layer
2. Keep the Go binary as the deployment/orchestration core
3. Add a new `codex` target that emits `AGENTS.md`, Codex skills, Codex agent definitions where possible, and Codex-oriented wrapper commands
4. Redesign Claude-only hooks and team-runtime features instead of trying to emulate them exactly

## 3. What MoAI-ADK Actually Is

### 3.1 Core system shape

From the repository and generated codemaps, `moai-adk` has three major layers:

1. Control layer
   - Claude slash commands
   - Claude subagents
   - Claude Agent Teams
   - Claude hooks
   - Claude MCP wiring
2. Runtime layer
   - Go CLI
   - template deployer
   - hook dispatcher
   - worktree orchestrator
   - config loader
   - quality gates
   - update/merge engine
3. Project state layer
   - `.moai/specs`
   - `.moai/project/codemaps`
   - `.moai/config/sections`
   - manifest and logs

This explains why the repository contains both high-level prompt assets and a large Go codebase.

### 3.2 Initialization model

`moai init` is the real installation engine.
It does not just print instructions.
It deploys a full project-local agent operating environment from embedded templates and generates runtime config programmatically.

Observed flow:

1. Run init wizard
2. Detect language/framework/methodology
3. Deploy embedded templates into the target project
4. Generate `settings.json` and `.mcp.json`
5. Write `CLAUDE.md`, `.claude/*`, `.moai/*`
6. Record deployed files in a manifest
7. Enable future safe updates via 3-way merge

This means the migration target should not be "rewrite prompts first".
It should be "extend the template deployment system to support Codex".

## 4. Claude Code-Centric Structural Flow

### 4.1 Main workflow

MoAI documents a canonical `project -> plan -> run -> sync` pipeline:

1. `/moai project`
   - generate product/structure/tech docs and codemaps
2. `/moai plan`
   - analyze request and create a SPEC in EARS format
3. `/moai run`
   - implement with TDD or DDD
4. `/moai sync`
   - update docs, codemaps, changelog, PR-ready state

There is also a default autonomous route where `/moai` can execute the full pipeline when complexity warrants it.

### 4.2 Methodology selection

The methodology engine is portable and valuable:

- TDD for new projects or projects with enough test coverage
- DDD-style preservation/improvement flow for brownfield code with weak tests
- TRUST 5 as the quality gate

This part is mostly vendor-neutral and should be retained unchanged in a Codex migration.

### 4.3 Agent topology

MoAI organizes work into 27 specialized agents:

- Manager agents
- Expert agents
- Builder agents
- Team agents

Important detail:
the real operational unit is not "one giant Claude prompt", but a dispatcher that selects specialized agents and passes structured responsibilities downstream.

### 4.4 Claude-native execution modes

MoAI relies on two Claude-native execution styles:

1. Sub-Agent mode
   - sequential delegation through Claude's task/subagent model
2. Agent Teams mode
   - parallel teammate collaboration with team lifecycle events

Auto-selection is based on complexity, file count, and domain count.

This is a major migration fault line because Codex can support delegation and parallelism, but Claude's Agent Teams lifecycle is materially different from Codex's current model.

## 5. Claude-Specific Dependencies Inside MoAI

These are the most important Claude-bound mechanisms.

### 5.1 `CLAUDE.md` as the root instruction contract

`CLAUDE.md` is the top-level execution directive.
It defines routing, hard rules, workflow structure, delegation norms, quality gates, and user interaction rules.

Migration note:
this maps conceptually to Codex `AGENTS.md`, and this is one of the cleanest translation paths.

### 5.2 `.claude/commands` as slash-command entry points

Each `/moai ...` command is surfaced through Claude custom slash commands.
In practice these command files are thin wrappers that forward into the unified MoAI skill/runtime.

Migration note:
this is partially portable if Codex custom slash commands are used, but the command surface will likely need a different packaging strategy.

### 5.3 `.claude/agents` as project-shipped subagent definitions

Agent files contain structured frontmatter such as:

- `name`
- `description`
- `tools`
- `model`
- `permissionMode`
- `maxTurns`
- `memory`
- `skills`
- `hooks`
- team-only controls like `background` and `isolation: worktree`

Migration note:
Codex supports multi-agent patterns and custom agents, but the packaging and lifecycle model is different enough that these files cannot be assumed portable as-is.

### 5.4 `.claude/hooks` + `.claude/settings.json`

This is the strongest Claude lock-in.
MoAI wires many lifecycle events directly into Claude Code:

- `SessionStart`
- `SessionEnd`
- `PreCompact`
- `PreToolUse`
- `PostToolUse`
- `Stop`
- `SubagentStart`
- `SubagentStop`
- `PostToolUseFailure`
- `Notification`
- `UserPromptSubmit`
- `PermissionRequest`
- `TeammateIdle`
- `TaskCompleted`

These events drive real logic:

- security scanning before dangerous tool use
- LSP diagnostics after edits
- token/task metrics logging
- prompt preprocessing
- permission routing
- team quality gates
- rejection of idle/completed states when standards are not met

Migration note:
Codex currently does not expose an equivalent always-on hook framework with the same lifecycle semantics.
This entire subsystem needs redesign, not translation.

### 5.5 Agent Teams protocol

MoAI depends on Claude Team features such as:

- shared task list semantics
- teammate idle checks
- task completion interception
- worktree/background teammate roles
- message passing patterns

Migration note:
this is the second strongest Claude lock-in after hooks.

### 5.6 Claude settings and environment controls

MoAI edits/generated Claude-specific runtime controls such as:

- `.claude/settings.json`
- `outputStyle`
- `statusLine`
- `permissions`
- `CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS=1`
- model routing via Claude-oriented fields

Migration note:
Codex has its own config, sandbox, approval, and TUI controls, but not this same config surface.

## 6. What Is Portable to Codex

These parts are strong migration candidates.

### 6.1 `.moai` project layer

Portable with minimal redesign:

- `.moai/specs/*`
- `.moai/project/codemaps/*`
- `.moai/config/sections/*`
- `.moai/manifest.json`
- quality state and workflow metadata

Why:
these represent workflow state, documentation, and policy rather than provider-specific protocol.

### 6.2 Go runtime core

Portable or reusable:

- template deployer
- update engine
- manifest tracking
- worktree orchestration
- config loading
- quality validation
- status reporting

The CLI/runtime is already provider-adjacent rather than provider-exclusive.
It needs new providers and adapters, not a total rewrite.

### 6.3 Workflow philosophy

Highly portable:

- SPEC-first execution
- TDD/DDD mode selection
- TRUST 5 gates
- codemap generation
- worktree-based isolation
- scaffolding-first
- fail/fix loops
- progressive disclosure

This is exactly the layer that should survive the migration unchanged.

## 7. Codex Capability Fit

Based on current Codex docs and the local Codex CLI environment, Codex already supports several pieces MoAI needs.

### 7.1 Good fit

Codex has strong equivalents for:

- repository-scoped instructions via `AGENTS.md`
- reusable skills
- MCP integration
- non-interactive execution via `codex exec`
- review mode
- sandbox and approval controls
- session resume/fork
- optional web search
- custom automation surfaces
- multi-agent/delegation support

This means Codex can host:

- root orchestration instructions
- domain skills
- provider-specific operating conventions
- scripted workflow execution
- autonomous or semi-autonomous execution modes

### 7.2 Partial fit

Codex likely supports these in a usable but not identical way:

- custom slash command workflows
- custom agents
- multi-agent decomposition
- project-local skill loading

These are viable for migration, but need a Codex-native information architecture instead of mirroring `.claude/*`.

### 7.3 Weak fit or current gap

These Claude mechanisms do not currently map cleanly:

- global lifecycle hooks on prompt/tool/session events
- teammate lifecycle interception (`TeammateIdle`, `TaskCompleted`)
- Claude Agent Teams collaboration protocol
- provider-specific output-style runtime layer
- `.claude/settings.json`-style per-project hook/permission wiring
- direct reuse of Claude permission frontmatter

This means a Codex edition should shift from event-driven internal interception to explicit orchestration and external enforcement.

## 8. Claude Code vs Codex: Structural Difference

### 8.1 Core control philosophy

Claude Code in MoAI is used as an event-rich runtime.
Codex is better viewed as an instruction-driven execution runtime with configurable autonomy, approvals, tools, skills, and automation entry points.

So:

- Claude favors in-session lifecycle interception
- Codex favors instruction control, execution policy, and orchestrated tasks

### 8.2 Best migration mental model

Do not ask:
"How do we recreate every Claude hook in Codex?"

Ask instead:
"How do we preserve the harness so Codex naturally produces the same quality outcomes?"

That is also consistent with OpenAI's own recent harness engineering framing:

- reduce entropy
- front-load scaffolding and context
- encode checks into the environment
- automate verification loops
- make autonomy safe through constraints

This aligns with MoAI philosophically, even where the concrete runtime differs.

## 9. Migration Mapping Matrix

### 9.1 Direct or near-direct mapping

| MoAI / Claude asset | Codex target | Migration difficulty | Notes |
|---|---|---:|---|
| `CLAUDE.md` | `AGENTS.md` | Low | Core orchestration rules can be translated |
| `.claude/skills/*` | `.codex/skills/*` | Low-Medium | Rework metadata and trigger language |
| `.mcp.json` | Codex MCP config | Low-Medium | Server definitions portable, format needs adjustment |
| `.moai/specs/*` | keep as-is | Low | Provider-neutral workflow state |
| `.moai/project/codemaps/*` | keep as-is | Low | Valuable for both providers |
| `.moai/config/sections/*` | keep as-is with provider section split | Medium | Add provider abstraction |

### 9.2 Adaptation required

| MoAI / Claude asset | Codex target | Migration difficulty | Notes |
|---|---|---:|---|
| `.claude/agents/*` | Codex agents or skill+dispatcher model | Medium-High | Depends on Codex custom agent packaging constraints |
| `.claude/commands/*` | Codex slash commands / exec wrappers / prompt entrypoints | Medium | Better to simplify command surface |
| `permissions` frontmatter | Codex sandbox + approval policy | Medium | Semantic translation, not file translation |
| model policy in agent frontmatter | Codex model/config/profile policy | Medium | Needs new configuration strategy |

### 9.3 Must redesign

| MoAI / Claude asset | Codex target | Migration difficulty | Notes |
|---|---|---:|---|
| `.claude/settings.json` hook graph | explicit CLI orchestration + CI + wrappers | High | No clean Codex equivalent |
| `TeammateIdle` / `TaskCompleted` quality interception | explicit phase gates | High | Move checks out of event loop |
| Agent Teams live collaboration model | fan-out/fan-in orchestrator | High | Convert to orchestrated parallel tasks |
| `moai cc/glm/cg` editing Claude settings | Codex runtime profiles | High | Provider-specific mode system must change |

## 10. Recommended Codex Migration Architecture

### 10.1 Target shape

Create a dual-provider architecture:

1. Vendor-neutral core
   - `.moai/*`
   - workflow engine
   - spec engine
   - codemap engine
   - quality engine
   - worktree engine
2. Claude adapter
   - existing `.claude/*`
   - existing Claude settings/hook generation
3. Codex adapter
   - `AGENTS.md`
   - `.codex/skills/*`
   - optional `.codex/agents/*` if supported in project scope
   - Codex-specific workflow wrappers
   - Codex config/profile guidance

### 10.2 New abstraction boundary

Introduce provider-aware deployment in `moai init`:

`moai init --target claude`
`moai init --target codex`
`moai init --target dual`

`dual` is strategically useful because it allows teams to transition incrementally.

### 10.3 Recommended file strategy

Keep:

- `.moai/**`
- shared workflow docs
- shared methodology/config files

Generate per provider:

- `CLAUDE.md`
- `.claude/**`
- `AGENTS.md`
- `.codex/skills/**`
- `.codex/agents/**` if viable

Avoid:

- mixing provider-specific lifecycle logic into shared `.moai` sections without namespacing

## 11. How to Replace the Claude Hook System in Codex

This is the most important design choice.

Do not attempt a fake hook-by-hook emulation layer inside Codex.
Use explicit orchestration instead.

### 11.1 Replacement model

Replace hook-time enforcement with four explicit control points:

1. Pre-execution gate
   - validate config, branch/worktree, spec presence, task scope
2. In-phase gate
   - after each plan/run/sync step, run quality/lint/test checks explicitly
3. Post-execution gate
   - verify TRUST 5, codemap freshness, docs sync, diff quality
4. External enforcement
   - CI, Git hooks, scripted validators

### 11.2 Example mapping

| Claude hook behavior | Codex replacement |
|---|---|
| `PreToolUse` security gate | explicit secure command wrapper before shell-heavy tasks |
| `PostToolUse` diagnostics collection | run diagnostics after edit batches or phase completion |
| `UserPromptSubmit` preprocessing | encode routing rules in `AGENTS.md` and command wrappers |
| `PermissionRequest` policy | Codex approval mode + sandbox policy |
| `TeammateIdle` quality rejection | worker completion validator before fan-in |
| `TaskCompleted` rejection | orchestrator refuses merge/apply until checks pass |
| `SessionStart` banner/update/status | startup script or `moai status` + Codex automation |

This is less "magical" than Claude hooks, but more predictable and easier to reason about.

## 12. How to Replace Claude Agent Teams in Codex

The best Codex analogue is not live teammate collaboration.
It is controlled parallel decomposition.

### 12.1 Recommended Codex pattern

Use:

1. orchestrator prompt in `AGENTS.md`
2. specialized skills and optional custom agents
3. parallel worker runs over isolated worktrees
4. explicit fan-in merge/review stage

### 12.2 Execution pattern

1. Orchestrator reads SPEC
2. Split into independent work packages
3. Create git worktrees
4. Launch parallel Codex workers, each bound to a worktree
5. Run tests in each worktree
6. Collect outputs
7. Merge only validated work

This preserves the high-value outcome of Agent Teams without depending on Claude's team event protocol.

## 13. Proposed Codex Edition Scope

### 13.1 MVP

The first Codex edition should target parity on the core workflow, not the full Claude feature set.

MVP scope:

- `AGENTS.md` root orchestration
- `.codex/skills` for MoAI methodology and domains
- `moai init --target codex`
- `moai update --target codex`
- SPEC workflow using `project / plan / run / sync`
- explicit quality gate commands
- worktree-based parallel mode driven by Codex workers
- Codex MCP support

Do not block MVP on:

- full hook parity
- full live agent-team parity
- Claude GLM hybrid features
- output style cosmetics

### 13.2 Phase 2

- custom agent support if project-portable in Codex
- custom slash command equivalents
- Codex automations for recurring maintenance
- autonomous cleanup/refactor loops
- provider-aware quality dashboards

### 13.3 Phase 3

- dual-provider distribution
- migration utilities from `.claude/*` to Codex assets
- unified provider-neutral agent registry

## 14. Concrete Project Plan

### Phase A. Inventory and decoupling

1. Classify every deployed asset as `shared`, `claude-only`, or `provider-adapter`
2. Extract provider-neutral workflow concepts from Claude-authored docs
3. Define a provider abstraction in template deployment and config generation

### Phase B. Codex adapter bootstrap

1. Add `--target codex|dual` to init/update
2. Generate `AGENTS.md`
3. Generate `.codex/skills/*`
4. Add Codex-specific setup guidance
5. Add Codex MCP configuration writer

### Phase C. Workflow reconstruction

1. Recreate `project -> plan -> run -> sync` with Codex-oriented command entrypoints
2. Replace hook-based checks with explicit validators
3. Implement worktree-based parallel Codex workers

### Phase D. Autonomy layer

1. Add `codex exec`-based orchestration for headless runs
2. Add automation recipes
3. Add resume/fork-based long-running workflow guidance

### Phase E. Dual-provider hardening

1. Validate same SPEC can run in Claude and Codex modes
2. Compare output quality and throughput
3. Document operational differences clearly

## 15. Priority Risks

### Risk 1. Overfitting to Claude file layout

If the migration starts by copying `.claude/*` to `.codex/*`, the project will inherit the wrong abstraction.

Action:
start from workflow semantics, not from folder symmetry.

### Risk 2. Treating hooks as mandatory for quality

MoAI quality comes from the harness, not from hooks alone.
Hooks are one implementation of the harness.

Action:
encode checks into explicit run gates and external validators.

### Risk 3. Assuming Codex agent packaging is identical

Codex may support custom agents differently from Claude, including different scope and discovery rules.

Action:
design a fallback where skills + `AGENTS.md` carry most orchestration logic even if custom agents are limited.

### Risk 4. Chasing total feature parity too early

The fastest path to failure is trying to port:

- hooks
- teams
- GLM hybrid mode
- output styles

before the core workflow works.

Action:
ship Codex MVP around the SPEC pipeline first.

## 16. Immediate Next Steps

Recommended next implementation order:

1. Create a provider matrix for all deployed templates
2. Design `AGENTS.md` from `CLAUDE.md`
3. Convert the highest-value Claude skills into Codex skills
4. Add `--target codex` to `moai init`
5. Implement explicit phase validators to replace hook dependency
6. Add a Codex worktree runner for parallel execution

## 17. Final Judgment

Can `moai-adk` be applied to Codex now?

Yes, partially and meaningfully.

But not by direct transplantation.

What can be migrated now:

- methodology
- spec system
- codemaps
- trust gates
- worktree orchestration
- skills
- root instruction layer
- MCP usage
- scripted autonomous flows

What must be redesigned:

- Claude hooks
- Claude Agent Teams semantics
- Claude settings/runtime manipulation
- Claude-specific command and permission surfaces

The strategic recommendation is:

Build `MoAI-Codex` as a new provider adapter on top of the existing MoAI core, not as a forked prompt pack.
