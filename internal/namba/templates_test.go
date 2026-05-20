package namba

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestPlanningRoleTemplatesPreserveRoleCardAndCustomAgentContracts(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name                 string
		roleCard             string
		customAgent          string
		roleHeading          string
		roleUseWhen          string
		roleResponsibilities []string
		agentSnippets        []string
	}{
		{
			name:        "planner",
			roleCard:    renderPlannerRoleCard(),
			customAgent: renderPlannerCustomAgent(),
			roleHeading: "# Namba Planner",
			roleUseWhen: "Use this role when breaking down a SPEC package before implementation.",
			roleResponsibilities: []string{
				"Read `spec.md`, `plan.md`, and `acceptance.md`.",
				"Identify target files, risks, and validation commands.",
				"Produce a concise execution plan for the main session.",
				"Do not edit files directly.",
			},
			agentSnippets: []string{
				`name = "namba-planner"`,
				`sandbox_mode = "read-only"`,
				"You are Namba Planner.",
				"Use this custom agent when breaking down a SPEC package before implementation.",
				"- When repo-managed MCP presets are configured, consult them first when they can ground planning decisions with better source material or verification signals.",
			},
		},
		{
			name:        "plan reviewer",
			roleCard:    renderPlanReviewerRoleCard(),
			customAgent: renderPlanReviewerCustomAgent(),
			roleHeading: "# Namba Plan Reviewer",
			roleUseWhen: "Use this role for aggregate validation of plan-review artifacts before implementation starts.",
			roleResponsibilities: []string{
				"Read `spec.md`, `plan.md`, `acceptance.md`, and the review artifacts under `.namba/specs/<SPEC>/reviews/`.",
				"Check whether the product, engineering, and design review set is coherent, sufficiently deep, and reflected correctly in `readiness.md`.",
				"Call out contradictions, missing review depth, or weak acceptance coverage, and identify which review tracks need to rerun.",
				"Do not implement code or quietly turn the advisory review flow into a hidden hard gate.",
			},
			agentSnippets: []string{
				`name = "namba-plan-reviewer"`,
				`sandbox_mode = "read-only"`,
				"You are Namba Plan Reviewer.",
				"Use this custom agent for aggregate validation of plan-review artifacts before implementation starts.",
				"- Keep the review flow advisory unless the main session explicitly asks for a hard gate.",
				"- Do not implement code or rewrite unrelated files.",
			},
		},
		{
			name:        "product manager",
			roleCard:    renderProductManagerRoleCard(),
			customAgent: renderProductManagerCustomAgent(),
			roleHeading: "# Namba Product Manager",
			roleUseWhen: "Use this role when shaping scope, acceptance, and delivery slicing before implementation.",
			roleResponsibilities: []string{
				"Translate user goals into concrete scope, constraints, and success criteria.",
				"Tighten acceptance criteria, non-goals, and rollout boundaries.",
				"Break large ideas into deliverable slices the main session can schedule.",
				"Call out UX, data, and operational implications early.",
			},
			agentSnippets: []string{
				`name = "namba-product-manager"`,
				`sandbox_mode = "read-only"`,
				"You are Namba Product Manager.",
				"Use this custom agent when a request needs stronger product framing before implementation starts.",
				"- Do not implement code directly.",
			},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			if !strings.Contains(tc.roleCard, tc.roleHeading) {
				t.Fatalf("role card missing heading %q: %q", tc.roleHeading, tc.roleCard)
			}
			if !strings.Contains(tc.roleCard, tc.roleUseWhen) {
				t.Fatalf("role card missing use-when %q: %q", tc.roleUseWhen, tc.roleCard)
			}
			if !strings.Contains(tc.roleCard, "Responsibilities:") {
				t.Fatalf("role card missing responsibilities header: %q", tc.roleCard)
			}
			for _, responsibility := range tc.roleResponsibilities {
				if !strings.Contains(tc.roleCard, "- "+responsibility) {
					t.Fatalf("role card missing responsibility %q: %q", responsibility, tc.roleCard)
				}
			}
			if !strings.Contains(tc.customAgent, "Responsibilities:") {
				t.Fatalf("custom agent missing responsibilities header: %q", tc.customAgent)
			}
			for _, snippet := range tc.agentSnippets {
				if !strings.Contains(tc.customAgent, snippet) {
					t.Fatalf("custom agent missing snippet %q: %q", snippet, tc.customAgent)
				}
			}
		})
	}
}

func TestUIPlanningRoleTemplatesPreserveRoleCardAndCustomAgentContracts(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name                 string
		roleCard             string
		customAgent          string
		roleHeading          string
		roleUseWhen          string
		roleResponsibilities []string
		agentSnippets        []string
	}{
		{
			name:        "frontend architect",
			roleCard:    renderFrontendArchitectRoleCard(),
			customAgent: renderFrontendArchitectCustomAgent(),
			roleHeading: "# Namba Frontend Architect",
			roleUseWhen: "Use this role when frontend structure, state flow, or UI delivery planning needs to be clarified before editing.",
			roleResponsibilities: []string{
				"Identify component boundaries, state ownership, and data flow.",
				"Map UI changes to file targets, design-system constraints, and accessibility impact.",
				"Highlight responsive, performance, and browser-risk considerations.",
				"Recommend the smallest coherent UI implementation slice.",
			},
			agentSnippets: []string{
				`name = "namba-frontend-architect"`,
				`sandbox_mode = "read-only"`,
				"You are Namba Frontend Architect.",
				"Use this custom agent when a task needs frontend planning before implementation starts.",
				"- Do not edit files directly.",
			},
		},
		{
			name:        "designer",
			roleCard:    renderDesignerRoleCard(),
			customAgent: renderDesignerCustomAgent(),
			roleHeading: "# Namba Designer",
			roleUseWhen: "Use this role when art direction, palette/tone logic, composition, motion intent, or generic-looking UI surfaces need to be clarified before implementation or review.",
			roleResponsibilities: []string{
				"Lead with art direction: define the visual concept, hierarchy, and composition before defaulting to components or spacing tokens.",
				"Set palette logic with explicit temperature and undertone discipline, restrained saturation, and deliberate accent use instead of trend-chasing or washed-out minimalism.",
				"Choose semantic components and layout primitives that fit the content; do not default to interchangeable cards, border-heavy framing, or generic bento/grid patterns as the primary identity.",
				"Keep motion purposeful: use it only when it clarifies hierarchy, attention, or state change.",
				"For screen-, page-, or section-scale work, identify the most generic-looking section and propose a concrete redesign; for component-scale work, call out the risk without forcing gratuitous scope creep.",
				"Guard against overcorrection: do not flatten everything into gray minimalism, do not add novelty without payoff, and do not sacrifice accessibility, design-system fit, or implementation realism.",
			},
			agentSnippets: []string{
				`name = "namba-designer"`,
				`sandbox_mode = "read-only"`,
				"You are Namba Designer.",
				"Use this custom agent when a task needs design direction, taste correction, or visual review before implementation starts.",
				"- Do not edit files directly.",
			},
		},
		{
			name:        "mobile engineer",
			roleCard:    renderMobileEngineerRoleCard(),
			customAgent: renderMobileEngineerCustomAgent(),
			roleHeading: "# Namba Mobile Engineer",
			roleUseWhen: "Use this role when mobile-specific constraints, navigation, lifecycle, or platform behavior need to be clarified before editing.",
			roleResponsibilities: []string{
				"Define mobile component boundaries, platform-specific constraints, and ownership of shared versus native behavior.",
				"Map requested changes to navigation, lifecycle, offline, and responsive considerations.",
				"Highlight gesture, performance, and device-compatibility risks.",
				"Recommend the smallest mobile delivery slice the main session can delegate safely.",
			},
			agentSnippets: []string{
				`name = "namba-mobile-engineer"`,
				`sandbox_mode = "read-only"`,
				"You are Namba Mobile Engineer.",
				"Use this custom agent when a task needs mobile-specific planning before implementation starts.",
				"- Do not edit files directly.",
			},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			if !strings.Contains(tc.roleCard, tc.roleHeading) {
				t.Fatalf("role card missing heading %q: %q", tc.roleHeading, tc.roleCard)
			}
			if !strings.Contains(tc.roleCard, tc.roleUseWhen) {
				t.Fatalf("role card missing use-when %q: %q", tc.roleUseWhen, tc.roleCard)
			}
			if !strings.Contains(tc.roleCard, "Responsibilities:") {
				t.Fatalf("role card missing responsibilities header: %q", tc.roleCard)
			}
			for _, responsibility := range tc.roleResponsibilities {
				if !strings.Contains(tc.roleCard, "- "+responsibility) {
					t.Fatalf("role card missing responsibility %q: %q", responsibility, tc.roleCard)
				}
				if !strings.Contains(tc.customAgent, "- "+responsibility) {
					t.Fatalf("custom agent missing responsibility %q: %q", responsibility, tc.customAgent)
				}
			}
			if !strings.Contains(tc.customAgent, "Responsibilities:") {
				t.Fatalf("custom agent missing responsibilities header: %q", tc.customAgent)
			}
			for _, snippet := range tc.agentSnippets {
				if !strings.Contains(tc.customAgent, snippet) {
					t.Fatalf("custom agent missing snippet %q: %q", snippet, tc.customAgent)
				}
			}
		})
	}
}

func TestReadOnlyArchitectureAndReviewTemplatesPreserveRoleCardAndCustomAgentContracts(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name                   string
		roleCard               string
		customAgent            string
		roleHeading            string
		roleUseWhen            string
		roleResponsibilities   []string
		customResponsibilities []string
		agentSnippets          []string
	}{
		{
			name:        "backend architect",
			roleCard:    renderBackendArchitectRoleCard(),
			customAgent: renderBackendArchitectCustomAgent(),
			roleHeading: "# Namba Backend Architect",
			roleUseWhen: "Use this role when backend contracts, service boundaries, or persistence changes need to be clarified before implementation.",
			roleResponsibilities: []string{
				"Define API, service, and persistence boundaries for the requested change.",
				"Call out schema, transaction, idempotency, and rollback risks.",
				"Identify security, observability, and migration implications.",
				"Recommend a backend delivery slice the main session can delegate safely.",
			},
			customResponsibilities: []string{
				"Define API, service, and persistence boundaries for the requested change.",
				"Call out schema, transaction, idempotency, and rollback risks.",
				"Identify security, observability, and migration implications.",
				"Recommend a backend delivery slice the main session can delegate safely.",
				"Do not edit files directly.",
			},
			agentSnippets: []string{
				`name = "namba-backend-architect"`,
				`sandbox_mode = "read-only"`,
				"You are Namba Backend Architect.",
				"Use this custom agent when a task needs backend planning before implementation starts.",
			},
		},
		{
			name:        "reviewer",
			roleCard:    renderReviewerRoleCard(),
			customAgent: renderReviewerCustomAgent(),
			roleHeading: "# Namba Reviewer",
			roleUseWhen: "Use this role for acceptance and quality review before sync.",
			roleResponsibilities: []string{
				"Compare the implementation with `acceptance.md`.",
				"Check that validation output and artifacts exist.",
				"Call out regressions, missing tests, or documentation drift.",
				"Do not rewrite the implementation unless asked.",
			},
			customResponsibilities: []string{
				"Compare the implementation with `acceptance.md`.",
				"Check that validation output and expected artifacts exist.",
				"Call out regressions, missing tests, and documentation drift.",
				"Do not rewrite the implementation unless explicitly asked.",
			},
			agentSnippets: []string{
				`name = "namba-reviewer"`,
				`sandbox_mode = "read-only"`,
				"You are Namba Reviewer.",
				"Use this custom agent for acceptance and quality review before sync.",
			},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			if !strings.Contains(tc.roleCard, tc.roleHeading) {
				t.Fatalf("role card missing heading %q: %q", tc.roleHeading, tc.roleCard)
			}
			if !strings.Contains(tc.roleCard, tc.roleUseWhen) {
				t.Fatalf("role card missing use-when %q: %q", tc.roleUseWhen, tc.roleCard)
			}
			if !strings.Contains(tc.roleCard, "Responsibilities:") {
				t.Fatalf("role card missing responsibilities header: %q", tc.roleCard)
			}
			for _, responsibility := range tc.roleResponsibilities {
				if !strings.Contains(tc.roleCard, "- "+responsibility) {
					t.Fatalf("role card missing responsibility %q: %q", responsibility, tc.roleCard)
				}
			}
			if !strings.Contains(tc.customAgent, "Responsibilities:") {
				t.Fatalf("custom agent missing responsibilities header: %q", tc.customAgent)
			}
			for _, responsibility := range tc.customResponsibilities {
				if !strings.Contains(tc.customAgent, "- "+responsibility) {
					t.Fatalf("custom agent missing responsibility %q: %q", responsibility, tc.customAgent)
				}
			}
			for _, snippet := range tc.agentSnippets {
				if !strings.Contains(tc.customAgent, snippet) {
					t.Fatalf("custom agent missing snippet %q: %q", snippet, tc.customAgent)
				}
			}
		})
	}
}

func TestWorkspaceWriteRoleTemplatesPreserveRoleCardAndCustomAgentContracts(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name                   string
		roleCard               string
		customAgent            string
		roleHeading            string
		roleUseWhen            string
		roleResponsibilities   []string
		customResponsibilities []string
		agentSnippets          []string
	}{
		{
			name:        "backend implementer",
			roleCard:    renderBackendImplementerRoleCard(),
			customAgent: renderBackendImplementerCustomAgent(),
			roleHeading: "# Namba Backend Implementer",
			roleUseWhen: "Use this role when implementing approved server-side work.",
			roleResponsibilities: []string{
				"Change only the backend files assigned by the main session.",
				"Keep API contracts, validation, and persistence logic internally consistent.",
				"Add or update targeted backend tests when the change affects behavior.",
				"Report migration, rollout, or compatibility risks with the patch.",
			},
			customResponsibilities: []string{
				"Change only the backend files assigned by the main session.",
				"Keep API contracts, validation, and persistence logic internally consistent.",
				"Add or update targeted backend tests when the change affects behavior.",
				"Report migration, rollout, or compatibility risks with the patch.",
			},
			agentSnippets: []string{
				`name = "namba-backend-implementer"`,
				`sandbox_mode = "workspace-write"`,
				"You are Namba Backend Implementer.",
				"Use this custom agent when implementing approved server-side work.",
			},
		},
		{
			name:        "data engineer",
			roleCard:    renderDataEngineerRoleCard(),
			customAgent: renderDataEngineerCustomAgent(),
			roleHeading: "# Namba Data Engineer",
			roleUseWhen: "Use this role when schema, migration, pipeline, analytics, or transformation work is part of the change.",
			roleResponsibilities: []string{
				"Own data-model, migration, ETL, query, and analytics-facing code assigned by the main session.",
				"Keep schema changes, backfills, and data contracts internally consistent.",
				"Call out rollout sequencing, data quality risks, and irreversible migration concerns.",
				"Add or update focused validation for the changed data behavior when feasible.",
			},
			customResponsibilities: []string{
				"Own data-model, migration, ETL, query, and analytics-facing code assigned by the main session.",
				"Keep schema changes, backfills, and data contracts internally consistent.",
				"Call out rollout sequencing, data quality risks, and irreversible migration concerns.",
				"Add or update focused validation for the changed data behavior when feasible.",
			},
			agentSnippets: []string{
				`name = "namba-data-engineer"`,
				`sandbox_mode = "workspace-write"`,
				"You are Namba Data Engineer.",
				"Use this custom agent when schema, migration, pipeline, analytics, or transformation work is part of the change.",
			},
		},
		{
			name:        "security engineer",
			roleCard:    renderSecurityEngineerRoleCard(),
			customAgent: renderSecurityEngineerCustomAgent(),
			roleHeading: "# Namba Security Engineer",
			roleUseWhen: "Use this role when authentication, authorization, secrets, privacy, or hardening work is part of the change.",
			roleResponsibilities: []string{
				"Own security-sensitive code paths assigned by the main session.",
				"Tighten auth, permission, secret-handling, validation, and privacy boundaries without widening scope.",
				"Call out exploitability, compliance, rollback, and incident-response implications.",
				"Prefer the smallest defensible hardening patch plus explicit regression notes.",
			},
			customResponsibilities: []string{
				"Own security-sensitive code paths assigned by the main session.",
				"Tighten auth, permission, secret-handling, validation, and privacy boundaries without widening scope.",
				"Call out exploitability, compliance, rollback, and incident-response implications.",
				"Prefer the smallest defensible hardening patch plus explicit regression notes.",
			},
			agentSnippets: []string{
				`name = "namba-security-engineer"`,
				`sandbox_mode = "workspace-write"`,
				"You are Namba Security Engineer.",
				"Use this custom agent when authentication, authorization, secrets, privacy, or hardening work is part of the change.",
			},
		},
		{
			name:        "test engineer",
			roleCard:    renderTestEngineerRoleCard(),
			customAgent: renderTestEngineerCustomAgent(),
			roleHeading: "# Namba Test Engineer",
			roleUseWhen: "Use this role when acceptance coverage or regression protection needs to be strengthened.",
			roleResponsibilities: []string{
				"Turn acceptance criteria into concrete test scenarios and edge cases.",
				"Add the smallest high-value automated coverage for the changed behavior.",
				"Focus on regression detection rather than broad refactors.",
				"Report residual gaps when full automation is not practical.",
			},
			customResponsibilities: []string{
				"Turn acceptance criteria into concrete test scenarios and edge cases.",
				"Add the smallest high-value automated coverage for the changed behavior.",
				"Focus on regression detection rather than broad test refactors.",
				"Report residual gaps when full automation is not practical.",
			},
			agentSnippets: []string{
				`name = "namba-test-engineer"`,
				`sandbox_mode = "workspace-write"`,
				"You are Namba Test Engineer.",
				"Use this custom agent when acceptance coverage or regression protection needs to be strengthened.",
			},
		},
		{
			name:        "devops engineer",
			roleCard:    renderDevOpsEngineerRoleCard(),
			customAgent: renderDevOpsEngineerCustomAgent(),
			roleHeading: "# Namba DevOps Engineer",
			roleUseWhen: "Use this role when CI, runtime config, deployment, or operational automation is part of the change.",
			roleResponsibilities: []string{
				"Own pipeline, environment, container, and deployment-file changes assigned by the main session.",
				"Preserve release safety, rollback clarity, and secret-handling boundaries.",
				"Call out observability, operational risk, and environment drift.",
				"Keep infrastructure edits tightly scoped to the requested outcome.",
			},
			customResponsibilities: []string{
				"Own pipeline, environment, container, and deployment-file changes assigned by the main session.",
				"Preserve release safety, rollback clarity, and secret-handling boundaries.",
				"Call out observability, operational risk, and environment drift.",
				"Keep infrastructure edits tightly scoped to the requested outcome.",
			},
			agentSnippets: []string{
				`name = "namba-devops-engineer"`,
				`sandbox_mode = "workspace-write"`,
				"You are Namba DevOps Engineer.",
				"Use this custom agent when CI, runtime config, deployment, or operational automation is part of the change.",
			},
		},
		{
			name:        "implementer",
			roleCard:    renderImplementerRoleCard(),
			customAgent: renderImplementerCustomAgent(),
			roleHeading: "# Namba Implementer",
			roleUseWhen: "Use this role when implementing an approved portion of a SPEC package.",
			roleResponsibilities: []string{
				"Change only the files assigned by the main session.",
				"Preserve methodology rules from `.namba/config/sections/quality.yaml`.",
				"Run or report the relevant validation steps when feasible.",
				"Leave notes about validation status and residual risk.",
			},
			customResponsibilities: []string{
				"Change only the files assigned by the main session.",
				"Preserve methodology rules from `.namba/config/sections/quality.yaml`.",
				"Run or report the relevant validation steps when feasible.",
				"Leave notes about validation status and residual risk.",
			},
			agentSnippets: []string{
				`name = "namba-implementer"`,
				`sandbox_mode = "workspace-write"`,
				"You are Namba Implementer.",
				"Use this custom agent when implementing an approved portion of a SPEC package.",
			},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			if !strings.Contains(tc.roleCard, tc.roleHeading) {
				t.Fatalf("role card missing heading %q: %q", tc.roleHeading, tc.roleCard)
			}
			if !strings.Contains(tc.roleCard, tc.roleUseWhen) {
				t.Fatalf("role card missing use-when %q: %q", tc.roleUseWhen, tc.roleCard)
			}
			if !strings.Contains(tc.roleCard, "Responsibilities:") {
				t.Fatalf("role card missing responsibilities header: %q", tc.roleCard)
			}
			for _, responsibility := range tc.roleResponsibilities {
				if !strings.Contains(tc.roleCard, "- "+responsibility) {
					t.Fatalf("role card missing responsibility %q: %q", responsibility, tc.roleCard)
				}
			}
			if !strings.Contains(tc.customAgent, "Responsibilities:") {
				t.Fatalf("custom agent missing responsibilities header: %q", tc.customAgent)
			}
			for _, responsibility := range tc.customResponsibilities {
				if !strings.Contains(tc.customAgent, "- "+responsibility) {
					t.Fatalf("custom agent missing responsibility %q: %q", responsibility, tc.customAgent)
				}
			}
			for _, snippet := range tc.agentSnippets {
				if !strings.Contains(tc.customAgent, snippet) {
					t.Fatalf("custom agent missing snippet %q: %q", snippet, tc.customAgent)
				}
			}
		})
	}
}

func TestRenderCodexUsageTailSectionsPreserveDynamicAnchors(t *testing.T) {
	t.Parallel()

	profile := initProfile{
		ProjectName:          "repo",
		ConversationLanguage: "en",
		PRLanguage:           "ko",
		BranchBase:           "develop",
		PRBaseBranch:         "release",
		SpecBranchPrefix:     "spec/",
		TaskBranchPrefix:     "task/",
		CodexReviewComment:   "@codex review",
	}

	outputContract := strings.Join(renderCodexUsageOutputContractSection(profile), "\n")
	for _, want := range []string{
		"## Output Contract",
		outputContractHeaderExample(profile),
		outputContractSequence(profile),
		".namba/codex/validate-output-contract.py",
		"explicit repository enforcement path",
	} {
		if !strings.Contains(outputContract, want) {
			t.Fatalf("output-contract section missing %q: %q", want, outputContract)
		}
	}

	gitCollaboration := strings.Join(renderCodexUsageGitCollaborationSection(profile), "\n")
	for _, want := range []string{
		"## Git Collaboration Defaults",
		"`develop`",
		"`spec/<SPEC-ID>-<slug>`",
		"`task/<slug>`",
		"`release`",
		"Korean",
		"`@codex review`",
	} {
		if !strings.Contains(gitCollaboration, want) {
			t.Fatalf("git-collaboration section missing %q: %q", want, gitCollaboration)
		}
	}

	claudeMapping := strings.Join(renderCodexUsageClaudeMappingSection(), "\n")
	for _, want := range []string{
		"## Claude to Codex Mapping",
		"`CLAUDE.md` becomes `AGENTS.md`.",
		"`.agents/skills/`",
		"`.toml` custom agents",
		"`$namba`",
		"`$namba-review-resolve`",
		"`$namba-release`",
	} {
		if !strings.Contains(claudeMapping, want) {
			t.Fatalf("claude-mapping section missing %q: %q", want, claudeMapping)
		}
	}

	importantDistinction := strings.Join(renderCodexUsageImportantDistinctionSection(), "\n")
	for _, want := range []string{
		"## Important Distinction",
		"`namba run SPEC-XXX`",
		"`--solo`, `--team`, and worktree-based `--parallel`",
		"`gh auth login` or `glab auth login`",
	} {
		if !strings.Contains(importantDistinction, want) {
			t.Fatalf("important-distinction section missing %q: %q", want, importantDistinction)
		}
	}
}

func TestManagedCodexSkillRegistryIncludesQueueReviewResolveAndRelease(t *testing.T) {
	t.Parallel()

	for _, name := range []string{"namba-queue", "namba-review-resolve", "namba-release"} {
		found := false
		for _, managed := range managedCodexSkillNames() {
			if managed == name {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("managed skill registry missing %q: %v", name, managedCodexSkillNames())
		}

		rel := filepath.ToSlash(filepath.Join(repoSkillsDir, name, "SKILL.md"))
		if !isManagedRepoSkillPath(rel) {
			t.Fatalf("expected %s to be treated as a managed repo skill path", rel)
		}
	}

	templates := codexSkillTemplates(initProfile{})
	for _, rel := range []string{
		filepath.ToSlash(filepath.Join("namba-queue", "SKILL.md")),
		filepath.ToSlash(filepath.Join("namba-review-resolve", "SKILL.md")),
		filepath.ToSlash(filepath.Join("namba-release", "SKILL.md")),
	} {
		if _, ok := templates[rel]; !ok {
			t.Fatalf("codex skill templates missing %s", rel)
		}
	}
}

func TestRenderCodexUsageTailSectionsStayOrderedInIntegratedDoc(t *testing.T) {
	t.Parallel()

	content := renderCodexUsage(initProfile{
		ProjectName:          "repo",
		ConversationLanguage: "en",
		PRLanguage:           "ko",
		BranchBase:           "develop",
		PRBaseBranch:         "release",
		CodexReviewComment:   "@codex review",
	})

	lastIndex := -1
	for _, want := range []string{
		"prompt refinement before execution",
		"goals, scope, constraints, and acceptance criteria",
		"## Output Contract",
		"## Git Collaboration Defaults",
		"## Claude to Codex Mapping",
		"## Important Distinction",
	} {
		index := strings.Index(content, want)
		if index == -1 {
			t.Fatalf("codex usage missing %q: %q", want, content)
		}
		if index <= lastIndex {
			t.Fatalf("codex usage out of order for %q after index %d: %q", want, lastIndex, content)
		}
		lastIndex = index
	}
}

func TestRenderCodexUsageMidSectionsPreserveAnchors(t *testing.T) {
	t.Parallel()

	agentRoster := strings.Join(renderCodexUsageAgentRosterSection(), "\n")
	for _, want := range []string{
		"## Namba Custom Agent Roster",
		"`namba-product-manager`",
		"`namba-frontend-architect`",
		"`namba-designer` owns art direction",
		"`Plan the component/state split for this dashboard`",
		"`namba-backend-architect`",
		"`namba-security-engineer`",
		"`namba-implementer`",
		"`explorer` and `worker`",
	} {
		if !strings.Contains(agentRoster, want) {
			t.Fatalf("agent-roster section missing %q: %q", want, agentRoster)
		}
	}

	delegation := strings.Join(renderCodexUsageDelegationHeuristicsSection(), "\n")
	for _, want := range []string{
		"## Delegation Heuristics",
		"`--solo`",
		"`--team`",
		"`.codex/config.toml [agents].max_threads = 5`",
		"max_parallel_workers: 3",
		"`model` and `model_reasoning_effort`",
		"`namba-frontend-architect`",
		"`namba-reviewer`",
	} {
		if !strings.Contains(delegation, want) {
			t.Fatalf("delegation section missing %q: %q", want, delegation)
		}
	}
	for _, unwanted := range []string{"temperature and undertone discipline", "washed-out minimalism"} {
		if strings.Contains(delegation, unwanted) {
			t.Fatalf("delegation section should stay lightweight and not contain %q: %q", unwanted, delegation)
		}
	}

	planReview := strings.Join(renderCodexUsagePlanReviewReadinessSection(), "\n")
	for _, want := range []string{
		"## Plan Review Readiness",
		"`namba plan`, `namba harness`, and `namba fix --command plan`",
		"`$namba-plan-review`",
		"`$namba-plan-pm-review`, `$namba-plan-eng-review`, and `$namba-plan-design-review`",
		"`namba run`, `namba sync`, and `namba pr`",
	} {
		if !strings.Contains(planReview, want) {
			t.Fatalf("plan-review section missing %q: %q", want, planReview)
		}
	}
}

func TestDesignerContractStaysDesignSpecificAndDoesNotLeakIntoSharedSurfaces(t *testing.T) {
	t.Parallel()

	designer := renderDesignerRoleCard() + "\n" + renderDesignerCustomAgent()
	for _, want := range []string{
		"art direction",
		"temperature and undertone discipline",
		"restrained saturation",
		"generic bento/grid patterns",
		"most generic-looking section",
		"gray minimalism",
	} {
		if !strings.Contains(designer, want) {
			t.Fatalf("designer contract missing %q: %q", want, designer)
		}
	}

	frontendArchitect := renderFrontendArchitectRoleCard() + "\n" + renderFrontendArchitectCustomAgent()
	frontendImplementer := renderFrontendImplementerRoleCard() + "\n" + renderFrontendImplementerCustomAgent()
	for _, surface := range []string{frontendArchitect, frontendImplementer} {
		for _, unwanted := range []string{"temperature and undertone discipline", "gray minimalism", "generic bento/grid patterns"} {
			if strings.Contains(surface, unwanted) {
				t.Fatalf("non-design surface should not contain %q: %q", unwanted, surface)
			}
		}
	}

	runSkill := renderRunCommandSkill(initProfile{})
	for _, unwanted := range []string{"temperature and undertone discipline", "washed-out minimalism"} {
		if strings.Contains(runSkill, unwanted) {
			t.Fatalf("run skill should stay lightweight and not contain %q: %q", unwanted, runSkill)
		}
	}
}

func TestRenderInitCommandSkillMentionsHookReview(t *testing.T) {
	t.Parallel()

	skill := renderInitCommandSkill()
	for _, want := range []string{"repository-state-first", "should not ask for a starter app stack", "approachable emoji cues", "`b`/`back`", "do not ask the user for a GitHub username", "`/hooks`", "`6 hooks need review`", ".codex/hooks/namba_codex_guard.py", "prompt-refinement hooks"} {
		if !strings.Contains(skill, want) {
			t.Fatalf("init command skill missing hook-review guidance %q: %q", want, skill)
		}
	}
}

func TestRenderCodexUsageFrontSectionsPreserveAnchors(t *testing.T) {
	t.Parallel()

	initEnables := strings.Join(renderCodexUsageInitEnablesSection(), "\n")
	for _, want := range []string{
		"## What `namba init .` Enables",
		"`AGENTS.md`",
		"`.agents/skills/`",
		"`.codex/agents/*.toml`",
		"`.namba/codex/output-contract.md`",
		"`namba-review-resolve`",
		"`namba-release`",
		"`namba-queue`",
		"Creates `.namba/` project state, configs, docs, and SPEC storage.",
	} {
		if !strings.Contains(initEnables, want) {
			t.Fatalf("init-enables section missing %q: %q", want, initEnables)
		}
	}

	howCodexUses := strings.Join(renderCodexUsageHowCodexUsesNambaSection(), "\n")
	for _, want := range []string{
		"## How Codex Uses Namba After Init",
		"Open Codex in the initialized project directory.",
		"WSL workspace",
		"`6 hooks need review`",
		"`/hooks`",
		".codex/hooks/namba_codex_guard.py",
		"Codex loads `AGENTS.md` and repo skills.",
		"`default`, `worker`, and `explorer`",
		"`namba project`, `namba regen`, `namba update`, `namba codex access`, `namba plan`, `namba harness`, `namba fix`, `namba run SPEC-XXX`, `namba queue`, `namba sync`, `namba pr`, `namba land`, and `namba release`",
		"`$namba-queue`",
		"`$namba-review-resolve`",
		"`$namba-release`",
	} {
		if !strings.Contains(howCodexUses, want) {
			t.Fatalf("how-codex-uses section missing %q: %q", want, howCodexUses)
		}
	}
}

func TestRenderCodexUsageWorkflowCommandSemanticsSectionPreservesAnchors(t *testing.T) {
	t.Parallel()

	workflowSemantics := strings.Join(renderCodexUsageWorkflowCommandSemanticsSection(), "\n")
	for _, want := range []string{
		"## Workflow Command Semantics",
		"`$namba-help` explains how to use NambaAI",
		"`$namba-create` is the preview-first creation path",
		"`$namba-queue` operates the existing-SPEC queue conveyor",
		"`$namba-review-resolve` resolves GitHub review threads one by one",
		"`$namba-release` handles NambaAI release orchestration",
		"`namba codex access` inspects Namba runner `codex exec` access defaults",
		"Interactive Codex approval mode, permissions, service tier, effective workspace roots, models, auth, apps, web search, and platform sandbox choices stay user/session-owned",
		"Codex 0.131 may display approval mode, permissions, service tier, and effective workspace roots",
		"`codex update` updates the upstream Codex CLI itself",
		"`namba fix --command plan \"<issue description>\"` creates the next bugfix SPEC package plus review scaffolds.",
		"`namba queue start <SPEC-RANGE|SPEC-LIST>` processes existing SPEC packages",
		"`frontend-brief.md`",
		"`frontend-major`",
		"`namba run SPEC-XXX --parallel` still refers to the standalone worktree runner path.",
		"Codex `/goal` workflows are tracked as a future orchestration candidate",
	} {
		if !strings.Contains(workflowSemantics, want) {
			t.Fatalf("workflow-command-semantics section missing %q: %q", want, workflowSemantics)
		}
	}
}

func TestRunSkillAndUIRolesMentionFrontendGateContract(t *testing.T) {
	t.Parallel()

	runSkill := renderRunCommandSkill(initProfile{})
	for _, want := range []string{"`frontend-brief.md`", "`frontend-major`", "canonical source for frontend task classification", "Do-Not Design Contract", "Do-Not Design Violation Check", "Generated Asset Evidence"} {
		if !strings.Contains(runSkill, want) {
			t.Fatalf("run skill missing %q: %q", want, runSkill)
		}
	}

	frontendArchitect := renderFrontendArchitectRoleCard() + "\n" + renderFrontendArchitectCustomAgent()
	for _, want := range []string{"`frontend-major` synthesis", "design clearance", "architecture handoff", "Do-Not Design Contract", "asset manifest"} {
		if !strings.Contains(frontendArchitect, want) {
			t.Fatalf("frontend architect surface missing %q: %q", want, frontendArchitect)
		}
	}

	frontendImplementer := renderFrontendImplementerRoleCard() + "\n" + renderFrontendImplementerCustomAgent()
	for _, want := range []string{"frontend synthesis is cleared", "`frontend-brief.md`", "visual grammar", "Do-Not Design Violation Check", "Generated Asset Evidence"} {
		if !strings.Contains(frontendImplementer, want) {
			t.Fatalf("frontend implementer surface missing %q: %q", want, frontendImplementer)
		}
	}

	designer := renderDesignerRoleCard() + "\n" + renderDesignerCustomAgent()
	if !strings.Contains(designer, "reference collection and synthesis") && !strings.Contains(designer, "Collect or critique references") {
		t.Fatalf("designer surface should mention reference synthesis ownership: %q", designer)
	}
	for _, want := range []string{"Do-Not Design Contract", "allowed replacements", "visual grammar", "asset manifest", "generated-image plan", "generic-section proof"} {
		if !strings.Contains(designer, want) {
			t.Fatalf("designer surface missing %q: %q", want, designer)
		}
	}
}

func TestRenderCodexUsageWorkflowCommandSemanticsSectionStaysOrderedInIntegratedDoc(t *testing.T) {
	t.Parallel()

	content := renderCodexUsage(initProfile{
		ProjectName:          "repo",
		ConversationLanguage: "en",
		PRLanguage:           "ko",
		BranchBase:           "develop",
		PRBaseBranch:         "release",
		CodexReviewComment:   "@codex review",
	})

	lastIndex := -1
	for _, want := range []string{
		"## How Codex Uses Namba After Init",
		"## Workflow Command Semantics",
		"## Codex 0.131 Boundary",
		"## Namba Custom Agent Roster",
	} {
		index := strings.Index(content, want)
		if index == -1 {
			t.Fatalf("codex usage missing %q: %q", want, content)
		}
		if index <= lastIndex {
			t.Fatalf("codex usage out of order for %q after index %d: %q", want, lastIndex, content)
		}
		lastIndex = index
	}
}

func TestRenderCodexUsageMidAndTailSectionsStayOrderedInIntegratedDoc(t *testing.T) {
	t.Parallel()

	content := renderCodexUsage(initProfile{
		ProjectName:          "repo",
		ConversationLanguage: "en",
		PRLanguage:           "ko",
		BranchBase:           "develop",
		PRBaseBranch:         "release",
		CodexReviewComment:   "@codex review",
	})

	lastIndex := -1
	for _, want := range []string{
		"## Workflow Command Semantics",
		"## Codex 0.131 Boundary",
		"## Namba Custom Agent Roster",
		"## Delegation Heuristics",
		"## Plan Review Readiness",
		"## Output Contract",
		"## Git Collaboration Defaults",
		"## Claude to Codex Mapping",
		"## Important Distinction",
	} {
		index := strings.Index(content, want)
		if index == -1 {
			t.Fatalf("codex usage missing %q: %q", want, content)
		}
		if index <= lastIndex {
			t.Fatalf("codex usage out of order for %q after index %d: %q", want, lastIndex, content)
		}
		lastIndex = index
	}
}

func TestRenderCodexUsageFrontMidAndTailSectionsStayOrderedInIntegratedDoc(t *testing.T) {
	t.Parallel()

	content := renderCodexUsage(initProfile{
		ProjectName:          "repo",
		ConversationLanguage: "en",
		PRLanguage:           "ko",
		BranchBase:           "develop",
		PRBaseBranch:         "release",
		CodexReviewComment:   "@codex review",
	})

	lastIndex := -1
	for _, want := range []string{
		"## What `namba init .` Enables",
		"## How Codex Uses Namba After Init",
		"## Workflow Command Semantics",
		"## Codex 0.131 Boundary",
		"## Namba Custom Agent Roster",
		"## Delegation Heuristics",
		"## Plan Review Readiness",
		"## Output Contract",
		"## Git Collaboration Defaults",
		"## Claude to Codex Mapping",
		"## Important Distinction",
	} {
		index := strings.Index(content, want)
		if index == -1 {
			t.Fatalf("codex usage missing %q: %q", want, content)
		}
		if index <= lastIndex {
			t.Fatalf("codex usage out of order for %q after index %d: %q", want, lastIndex, content)
		}
		lastIndex = index
	}
}

func TestRenderNambaSkillSectionsPreserveAnchors(t *testing.T) {
	t.Parallel()

	commandMapping := strings.Join(renderNambaSkillCommandMappingSection(), "\n")
	for _, want := range []string{
		"Command mapping:",
		"`$namba-help`",
		"`$namba-create`",
		"`$namba-review-resolve`",
		"`$namba-release`",
		"`namba codex access`",
		"`$namba-plan-review`",
		"`namba run SPEC-XXX --solo|--team|--parallel`",
		"`namba queue start <SPEC-RANGE|SPEC-LIST>`",
		"`namba doctor`",
	} {
		if !strings.Contains(commandMapping, want) {
			t.Fatalf("namba-skill command-mapping section missing %q: %q", want, commandMapping)
		}
	}

	executionRules := strings.Join(renderNambaSkillExecutionRulesSection(initProfile{
		PRLanguage:         "ko",
		PRBaseBranch:       "release",
		CodexReviewComment: "@codex review",
	}), "\n")
	for _, want := range []string{
		"Execution rules:",
		"Treat `.namba/` as the source of truth.",
		"Prefer repo-local skills in `.agents/skills/`.",
		"`project`, `regen`, `update`, `codex access`, `plan`, `harness`, `fix`, `queue`, `pr`, `land`, `release`, and `sync`",
		"`--solo`, `--team`, `--parallel`, or `--dry-run`",
		"Prepare PRs against `release`, write the title/body in Korean, and request Codex review only when `namba pr --review` or queue `--review` is explicit",
	} {
		if !strings.Contains(executionRules, want) {
			t.Fatalf("namba-skill execution-rules section missing %q: %q", want, executionRules)
		}
	}
}

func TestRenderReviewResolveAndReleaseCommandSkillsPreserveContracts(t *testing.T) {
	t.Parallel()

	reviewResolve := renderReviewResolveCommandSkill()
	for _, want := range []string{
		"name: namba-review-resolve",
		"$namba-review-resolve",
		"thread-aware GitHub path such as `gh api graphql`",
		"`fixed-and-resolved`, `answered-open`, or `skipped-with-rationale`",
		"validation commands before replying or resolving",
		"configured `@codex review` marker is present exactly once",
	} {
		if !strings.Contains(reviewResolve, want) {
			t.Fatalf("review-resolve skill missing %q: %q", want, reviewResolve)
		}
	}

	release := renderReleaseCommandSkill()
	for _, want := range []string{
		"name: namba-release",
		"$namba-release",
		"clean working tree before the final tagging step",
		"`namba regen` and/or `namba sync`",
		"`.namba/releases/<version>.md`",
		"guarded `namba release --version <version> --push` path",
		"GitHub Release body uses the generated notes",
	} {
		if !strings.Contains(release, want) {
			t.Fatalf("release skill missing %q: %q", want, release)
		}
	}
}

func TestRenderQueueCommandSkillPreservesContract(t *testing.T) {
	t.Parallel()

	queue := renderQueueCommandSkill()
	for _, want := range []string{
		"name: namba-queue",
		"$namba-queue",
		"namba queue start <SPEC-RANGE|SPEC-LIST>",
		"`SPEC-001..SPEC-003`",
		"`namba queue status [--verbose]`",
		"`.namba/logs/queue/`",
		"one active SPEC at a time",
		"Block instead of skipping",
		"`waiting_for_land`",
		"`--skip-codex-review`",
	} {
		if !strings.Contains(queue, want) {
			t.Fatalf("queue skill missing %q: %q", want, queue)
		}
	}
}

func TestSkillSurfaceEvolutionHarnessContracts(t *testing.T) {
	t.Parallel()

	reviewResolve := renderReviewResolveCommandSkill()
	for _, want := range []string{
		"thread identity",
		"changed paths",
		"CI/check evidence when the review feedback or PR health depends on failing checks",
		"Inspect PR check status before re-requesting review",
	} {
		if !strings.Contains(reviewResolve, want) {
			t.Fatalf("review-resolve skill missing SPEC-040 contract %q: %q", want, reviewResolve)
		}
	}

	prSkill := renderPRCommandSkill(initProfile{PRBaseBranch: "main", PRLanguage: "ko", CodexReviewComment: "@codex review"})
	for _, want := range []string{
		"Inspect current PR check status before review handoff",
		"bounded GitHub Actions failure snippets",
		"external checks by status and details URL only",
		"When `--review` is present, ensure the configured Codex review marker exists exactly once",
	} {
		if !strings.Contains(prSkill, want) {
			t.Fatalf("pr skill missing SPEC-040 contract %q: %q", want, prSkill)
		}
	}

	harnessSkill := renderHarnessCommandSkill()
	for _, want := range []string{
		"same Namba clarification gate used by `namba plan`",
		"Goal/Scope/Constraints/Acceptance description to `namba harness",
		"deterministic helper-script candidates",
		"`--help`",
		"fixture or local-server tests",
		"mechanical versus behavioral edits",
		"update templates first, regenerate, review generated diffs, and validate",
		"workflow-first",
		"context-budgeted outputs",
		"actionable errors",
		"independent, read-only, realistic, verifiable, and stable",
	} {
		if !strings.Contains(harnessSkill, want) {
			t.Fatalf("harness skill missing SPEC-040 contract %q: %q", want, harnessSkill)
		}
	}

	fixSkill := renderFixCommandSkill()
	for _, want := range []string{
		"For `namba fix --command plan`, run the same clarification gate",
		"Goal/Scope/Constraints/Acceptance description to `namba fix --command plan",
		"direct-repair path",
	} {
		if !strings.Contains(fixSkill, want) {
			t.Fatalf("fix skill missing clarification contract %q: %q", want, fixSkill)
		}
	}

	runSkill := renderRunCommandSkill(initProfile{})
	executionSkill := renderExecutionSkill(initProfile{})
	for _, surface := range []struct {
		name    string
		content string
	}{
		{name: "run", content: runSkill},
		{name: "execution", content: executionSkill},
	} {
		for _, want := range []string{
			"managed server lifecycle",
			"rendered DOM",
			"screenshots",
			"console errors",
			"Playwright",
		} {
			if !strings.Contains(surface.content, want) {
				t.Fatalf("%s skill missing frontend validation evidence %q: %q", surface.name, want, surface.content)
			}
		}
	}

	createSkill := renderCreateCommandSkill()
	for _, want := range []string{
		"progressive disclosure",
		"references, assets, or deterministic helper candidates",
		"Do not add `$CODEX_HOME/skills`",
	} {
		if !strings.Contains(createSkill, want) {
			t.Fatalf("create skill missing progressive-disclosure contract %q: %q", want, createSkill)
		}
	}

	for _, surface := range []struct {
		name    string
		content string
	}{
		{name: "review-resolve", content: reviewResolve},
		{name: "pr", content: prSkill},
		{name: "harness", content: harnessSkill},
		{name: "fix", content: fixSkill},
		{name: "run", content: runSkill},
		{name: "execution", content: executionSkill},
		{name: "create", content: createSkill},
	} {
		for _, unwanted := range []string{"Composio CLI", "Slack", "Notion", ".codex/skills/"} {
			if strings.Contains(surface.content, unwanted) {
				t.Fatalf("%s skill should not introduce rejected flow %q: %q", surface.name, unwanted, surface.content)
			}
		}
	}
}

func TestRenderNambaSkillSectionsStayOrderedInIntegratedDoc(t *testing.T) {
	t.Parallel()

	content := renderNambaSkill(initProfile{
		PRLanguage:         "ko",
		PRBaseBranch:       "release",
		CodexReviewComment: "@codex review",
	})

	lastIndex := -1
	for _, want := range []string{
		"Use this skill whenever the user mentions `namba`",
		"Command mapping:",
		"Execution rules:",
		"Prepare PRs against `release`, write the title/body in Korean, and request Codex review only when `namba pr --review` or queue `--review` is explicit",
	} {
		index := strings.Index(content, want)
		if index == -1 {
			t.Fatalf("namba skill missing %q: %q", want, content)
		}
		if index <= lastIndex {
			t.Fatalf("namba skill out of order for %q after index %d: %q", want, lastIndex, content)
		}
		lastIndex = index
	}
}

func TestRenderNambaSkillRouterSectionsReserveCoachForAdvisoryRouting(t *testing.T) {
	t.Parallel()

	commandMapping := strings.Join(renderNambaSkillCommandMappingSection(), "\n")
	for _, want := range []string{
		"`$namba-coach`",
		"`$namba-help`",
		"`$namba-create`",
		"`namba plan \"<description>\"`",
		"`--no-review`",
		"`namba harness \"<description>\"`",
		"`namba fix --command plan \"<issue description>\"`",
		"`namba fix \"<issue description>\"` or `namba fix --command run \"<issue description>\"`",
		"`namba run SPEC-XXX`",
		"`namba queue start <SPEC-RANGE|SPEC-LIST>`",
		"`namba sync`",
		"`namba pr \"<title>\"`",
		"`namba land`",
	} {
		if !strings.Contains(commandMapping, want) {
			t.Fatalf("namba skill command-mapping section missing %q: %q", want, commandMapping)
		}
	}

	executionRules := strings.Join(renderNambaSkillExecutionRulesSection(initProfile{}), "\n")
	for _, want := range []string{
		"current-goal command coaching",
		"read-only usage guidance",
		"`$namba-coach`",
		"`$namba-create`",
		"`$namba-run`",
		"`$namba-queue`",
		"`$namba-pr`",
		"`$namba-land`",
		"`$namba-plan`",
		"`$namba-plan-review`",
		"`$namba-harness`",
		"`$namba-fix`",
	} {
		if !strings.Contains(executionRules, want) {
			t.Fatalf("namba skill execution-rules section missing %q: %q", want, executionRules)
		}
	}
}

func TestRenderPlanCommandSkillDefaultsToAutoReviewHandoff(t *testing.T) {
	t.Parallel()

	content := renderPlanCommandSkill()
	for _, want := range []string{
		"`--no-review`",
		"automatic handoff",
		"`$namba-plan-review SPEC-XXX`",
		"clarification gate",
		"`게시판 만들어줘`",
		"Goal, Scope, Constraints, and Acceptance",
		"Plan mode as the preferred clarification surface",
		"`namba plan` as the executor",
		"native Plan mode choice UI",
		"refined description",
	} {
		if !strings.Contains(content, want) {
			t.Fatalf("plan command skill missing %q: %q", want, content)
		}
	}

	planReview := renderPlanReviewLoopCommandSkill()
	for _, want := range []string{
		"same clarification gate as `$namba-plan`",
		"do not run the CLI while the raw request is still vague",
		"Goal/Scope/Constraints/Acceptance",
	} {
		if !strings.Contains(planReview, want) {
			t.Fatalf("plan-review command skill missing %q: %q", want, planReview)
		}
	}
}

func TestRenderRepoCodexConfigAvoidsSessionOwnedSettings(t *testing.T) {
	t.Parallel()

	profile := initProfile{
		AgentMode:         "multi",
		StatusLinePreset:  "namba",
		ApprovalPolicy:    "never",
		SandboxMode:       "danger-full-access",
		DefaultMCPServers: []string{"context7"},
	}
	first := renderRepoCodexConfig(profile)
	second := renderRepoCodexConfig(profile)
	if first != second {
		t.Fatalf("repo Codex config should render deterministically\nfirst=%q\nsecond=%q", first, second)
	}
	for _, want := range []string{
		"#:schema https://developers.openai.com/codex/config-schema.json",
		"# This file intentionally avoids session-owned Codex choices.",
		"[features]",
		"hooks = true",
		"goals = true",
		"[agents]",
		"max_threads = 5",
		"[tui]",
		"status_line",
		"[mcp_servers.context7]",
	} {
		if !strings.Contains(first, want) {
			t.Fatalf("repo Codex config missing %q: %q", want, first)
		}
	}
	for _, unwanted := range []string{
		"approval_policy =",
		"sandbox_mode =",
		"model =",
		"auth =",
		"web_search",
		"service_tier",
		"features.codex_hooks",
	} {
		if strings.Contains(first, unwanted) {
			t.Fatalf("repo Codex config must not render session-owned or legacy key %q: %q", unwanted, first)
		}
	}
}

func TestRenderNambaCodexHooksScaffold(t *testing.T) {
	t.Parallel()

	var parsed map[string]any
	hooksJSON := renderNambaCodexHooksJSON()
	if err := json.Unmarshal([]byte(hooksJSON), &parsed); err != nil {
		t.Fatalf("hooks JSON should be valid: %v\n%s", err, hooksJSON)
	}
	for _, want := range []string{
		`"SessionStart"`,
		`"PreToolUse"`,
		`"PermissionRequest"`,
		`"UserPromptSubmit"`,
		`"PostToolUse"`,
		`"Stop"`,
		`.codex/hooks/namba_codex_guard.sh`,
		`sh .codex/hooks/namba_codex_guard.sh`,
		`Checking NambaAI shell safety`,
		`Refining NambaAI prompt`,
		`Reviewing NambaAI surface changes`,
		`Checking NambaAI report format`,
	} {
		if !strings.Contains(hooksJSON, want) {
			t.Fatalf("hooks JSON missing %q: %q", want, hooksJSON)
		}
	}
	if strings.Contains(hooksJSON, "git rev-parse") {
		t.Fatalf("generated Codex hook commands must not require a git repository: %q", hooksJSON)
	}

	shellWrapper := renderNambaCodexHookGuardShellWrapper()
	for _, want := range []string{
		"POSIX shells",
		"python3",
		"NambaAI hook guard failed",
		"emit_failure",
		"extract_hook_event_name",
		"json_escape",
		"exit 1",
	} {
		if !strings.Contains(shellWrapper, want) {
			t.Fatalf("shell hook wrapper missing %q: %q", want, shellWrapper)
		}
	}

	script := renderNambaCodexHookGuardScript()
	for _, want := range []string{
		"permissionDecision",
		"PermissionRequest",
		"UserPromptSubmit",
		"prompt-refinement gate",
		"prompt_refinement_guidance",
		"matching SPEC creation command",
		"namba fix --command plan",
		`"additionalContext": guidance`,
		"updatedInput",
		"dedupe_key",
		"permissionDecision\": \"allow\"",
		"REPORT_SECTIONS",
		"approval_risk_note",
		"NAMBA_HOOK_TRACE_PATH",
		`record["prompt_refinement"]`,
		"git reset --hard",
		"git clean",
		"force-pushing requires an explicit human decision",
		"Namba-managed instruction or config surfaces are changed",
		"guardrails, not a complete security boundary",
		"emit_hook_error",
		"except Exception as exc",
		"sys.exit(main())",
	} {
		if !strings.Contains(script, want) {
			t.Fatalf("hook guard script missing %q: %q", want, script)
		}
	}
	userPromptHandlerIndex := strings.Index(script, "def handle_user_prompt_submit")
	if userPromptHandlerIndex < 0 {
		t.Fatalf("expected UserPromptSubmit handler in hook script: %q", script)
	}
	nextHandlerIndex := strings.Index(script[userPromptHandlerIndex+1:], "\ndef handle_")
	userPromptHandler := script[userPromptHandlerIndex:]
	if nextHandlerIndex >= 0 {
		userPromptHandler = script[userPromptHandlerIndex : userPromptHandlerIndex+1+nextHandlerIndex]
	}
	if strings.Contains(userPromptHandler, `"decision": "block"`) {
		t.Fatalf("prompt refinement must guide Codex without blocking user prompt submission: %q", userPromptHandler)
	}

	windowsHooksJSON := renderNambaCodexHooksJSONForOS("windows")
	for _, want := range []string{
		`powershell.exe -NoProfile -ExecutionPolicy Bypass -File .codex/hooks/namba_codex_guard.ps1`,
	} {
		if !strings.Contains(windowsHooksJSON, want) {
			t.Fatalf("windows hooks JSON missing %q: %q", want, windowsHooksJSON)
		}
	}
	for _, unwanted := range []string{` || `, `py -3 \".codex/hooks/namba_codex_guard.py\"`, `python \".codex/hooks/namba_codex_guard.py\"`, "Get-Command"} {
		if strings.Contains(windowsHooksJSON, unwanted) {
			t.Fatalf("windows hooks JSON must avoid PowerShell 7/cmd fallback syntax %q, got %q", unwanted, windowsHooksJSON)
		}
	}

	powerShellWrapper := renderNambaCodexHookGuardPowerShellWrapper()
	for _, want := range []string{
		"Windows PowerShell 5",
		"Write-NambaHookFailure",
		"Get-Command",
		"py.exe",
		"python.exe",
		"NambaAI hook guard failed",
		"HookEventName",
		"catch",
		"ConvertTo-Json -Compress",
		"--%",
		"exit 1",
		"exit 0",
	} {
		if want == "--%" {
			if strings.Contains(powerShellWrapper, want) {
				t.Fatalf("PowerShell hook wrapper must avoid unsupported stop-parsing token %q: %q", want, powerShellWrapper)
			}
			continue
		}
		if !strings.Contains(powerShellWrapper, want) {
			t.Fatalf("PowerShell hook wrapper missing %q: %q", want, powerShellWrapper)
		}
	}
}

func TestInitOnWindowsWritesCodexHooksWithWindowsCommand(t *testing.T) {
	t.Parallel()

	root := canonicalTempDir(t)
	app := NewApp(&bytes.Buffer{}, &bytes.Buffer{})
	app.goos = "windows"
	if err := app.Run(context.Background(), []string{"init", root, "--yes"}); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	hooksJSON := mustReadFile(t, filepath.Join(root, ".codex", "hooks.json"))
	for _, want := range []string{
		"UserPromptSubmit",
		`powershell.exe -NoProfile -ExecutionPolicy Bypass -File .codex/hooks/namba_codex_guard.ps1`,
	} {
		if !strings.Contains(hooksJSON, want) {
			t.Fatalf("expected Windows init hooks.json to contain %q, got %q", want, hooksJSON)
		}
	}
	for _, unwanted := range []string{` || `, `py -3 \".codex/hooks/namba_codex_guard.py\"`, `python \".codex/hooks/namba_codex_guard.py\"`, "Get-Command"} {
		if strings.Contains(hooksJSON, unwanted) {
			t.Fatalf("expected Windows init hooks.json to avoid unsupported fallback syntax %q, got %q", unwanted, hooksJSON)
		}
	}
	hookGuard := mustReadFile(t, filepath.Join(root, ".codex", "hooks", "namba_codex_guard.py"))
	if !strings.Contains(hookGuard, "NambaAI Codex lifecycle hook guard") {
		t.Fatalf("expected Windows init hook guard script, got %q", hookGuard)
	}
	hookWrapper := mustReadFile(t, filepath.Join(root, ".codex", "hooks", "namba_codex_guard.ps1"))
	if !strings.Contains(hookWrapper, "NambaAI hook guard failed") || !strings.Contains(hookWrapper, "Python 3 was not found or could not run") || !strings.Contains(hookWrapper, "exit 1") {
		t.Fatalf("expected Windows init PowerShell wrapper fallback, got %q", hookWrapper)
	}
}

func TestUserPromptSubmitHookGuidesWithoutBlockingSubmission(t *testing.T) {
	t.Parallel()

	tmp := canonicalTempDir(t)
	writeTestFile(t, filepath.Join(tmp, ".codex", "hooks.json"), renderNambaCodexHooksJSON())
	writeTestFile(t, filepath.Join(tmp, ".codex", "hooks", "namba_codex_guard.py"), renderNambaCodexHookGuardScript())
	writeTestFile(t, filepath.Join(tmp, ".codex", "hooks", "namba_codex_guard.sh"), renderNambaCodexHookGuardShellWrapper())

	for _, prompt := range []string{
		"namba plan 뭔가 개선해줘",
		"namba harness 뭔가 만들어줘",
		"namba fix --command plan 버그 고쳐줘",
	} {
		prompt := prompt
		t.Run(prompt, func(t *testing.T) {
			cmd := exec.Command("sh", "-c", codexHookCommandForEvent(t, renderNambaCodexHooksJSON(), "UserPromptSubmit"))
			cmd.Dir = tmp
			cmd.Stdin = strings.NewReader(`{"hook_event_name":"UserPromptSubmit","prompt":` + strconv.Quote(prompt) + `}`)
			output, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("expected UserPromptSubmit hook command to run, err=%v output=%s", err, output)
			}
			got := string(output)
			for _, want := range []string{`"hookSpecificOutput"`, `"hookEventName":"UserPromptSubmit"`, `"additionalContext"`, "Goal/Scope/Constraints/Acceptance", "namba harness", "namba fix --command plan"} {
				if !strings.Contains(got, want) {
					t.Fatalf("expected prompt guidance output to contain %q, got %q", want, got)
				}
			}
			if strings.Contains(got, `"decision":"block"`) || strings.Contains(got, `"decision": "block"`) {
				t.Fatalf("prompt guidance must not block submission, got %q", got)
			}
		})
	}
}

func TestRenderedCodexHookCommandsRunOutsideGitRepository(t *testing.T) {
	t.Parallel()

	tmp := canonicalTempDir(t)
	writeTestFile(t, filepath.Join(tmp, ".codex", "hooks.json"), renderNambaCodexHooksJSON())
	writeTestFile(t, filepath.Join(tmp, ".codex", "hooks", "namba_codex_guard.py"), renderNambaCodexHookGuardScript())
	writeTestFile(t, filepath.Join(tmp, ".codex", "hooks", "namba_codex_guard.sh"), renderNambaCodexHookGuardShellWrapper())

	payloads := map[string]string{
		"SessionStart":      `{"hook_event_name":"SessionStart"}`,
		"PreToolUse":        `{"hook_event_name":"PreToolUse","tool_input":{"command":"echo ok"}}`,
		"PermissionRequest": `{"hook_event_name":"PermissionRequest","tool_input":{"command":"echo ok"}}`,
		"UserPromptSubmit":  `{"hook_event_name":"UserPromptSubmit","prompt":"hello"}`,
		"PostToolUse":       `{"hook_event_name":"PostToolUse","cwd":"` + filepath.ToSlash(tmp) + `"}`,
		"Stop":              `{"hook_event_name":"Stop","last_assistant_message":"short","cwd":"` + filepath.ToSlash(tmp) + `"}`,
	}
	for event, payload := range payloads {
		t.Run(event, func(t *testing.T) {
			cmd := exec.Command("sh", "-c", codexHookCommandForEvent(t, renderNambaCodexHooksJSON(), event))
			cmd.Dir = tmp
			cmd.Stdin = strings.NewReader(payload)
			output, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("expected %s hook command to run outside a git repository, err=%v output=%s", event, err, output)
			}
		})
	}
}

func TestShellHookLauncherFailsLoudlyWhenPythonIsUnavailable(t *testing.T) {
	t.Parallel()

	tmp := canonicalTempDir(t)
	writeTestFile(t, filepath.Join(tmp, ".codex", "hooks", "namba_codex_guard.sh"), renderNambaCodexHookGuardShellWrapper())

	shellPath := "/bin/sh"
	if _, err := os.Stat(shellPath); err != nil {
		t.Skipf("%s is unavailable: %v", shellPath, err)
	}

	cmd := exec.Command(shellPath, filepath.Join(tmp, ".codex", "hooks", "namba_codex_guard.sh"))
	cmd.Dir = tmp
	cmd.Env = []string{"PATH=/definitely-no-python-here", "TMPDIR=" + tmp}
	cmd.Stdin = strings.NewReader(`{"hook_event_name":"UserPromptSubmit","prompt":"hello"}`)
	output, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("expected hook launcher to return a non-zero error when Python is unavailable, output=%s", output)
	}
	got := string(output)
	for _, want := range []string{"NambaAI hook guard failed", "Python 3 was not found", `"hookSpecificOutput"`} {
		if !strings.Contains(got, want) {
			t.Fatalf("expected loud hook launcher failure to contain %q, got %q", want, got)
		}
	}
	if !strings.Contains(got, `"hookEventName":"UserPromptSubmit"`) {
		t.Fatalf("expected launcher failure to preserve originating hook event, got %q", got)
	}
}

func TestShellHookLauncherHandlesQuotedPathsAndStdinJSON(t *testing.T) {
	t.Parallel()

	tmp := filepath.Join(canonicalTempDir(t), "repo with spaces")
	writeTestFile(t, filepath.Join(tmp, ".codex", "hooks", "namba_codex_guard.py"), renderNambaCodexHookGuardScript())
	writeTestFile(t, filepath.Join(tmp, ".codex", "hooks", "namba_codex_guard.sh"), renderNambaCodexHookGuardShellWrapper())

	shellPath := "/bin/sh"
	if _, err := os.Stat(shellPath); err != nil {
		t.Skipf("%s is unavailable: %v", shellPath, err)
	}

	cmd := exec.Command(shellPath, filepath.Join(tmp, ".codex", "hooks", "namba_codex_guard.sh"))
	cmd.Dir = tmp
	cmd.Env = append(os.Environ(), "NAMBA_HOOK_DEDUPE=0")
	cmd.Stdin = strings.NewReader(`{"hook_event_name":"SessionStart","session-id":"quoted-path-session"}`)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("expected quoted-path hook launcher to succeed, err=%v output=%s", err, output)
	}
	got := string(output)
	for _, want := range []string{`"hookSpecificOutput"`, `"hookEventName":"SessionStart"`, "NambaAI lifecycle hook is active"} {
		if !strings.Contains(got, want) {
			t.Fatalf("expected quoted-path launcher output to contain %q, got %q", want, got)
		}
	}
}

func TestHookGuardEmitsASCIIJSONUnderCP949Stdout(t *testing.T) {
	t.Parallel()

	pythonPath, err := exec.LookPath("python3")
	if err != nil {
		t.Skipf("python3 unavailable: %v", err)
	}

	tmp := canonicalTempDir(t)
	writeTestFile(t, filepath.Join(tmp, ".codex", "hooks", "namba_codex_guard.py"), renderNambaCodexHookGuardScript())
	writeTestFile(t, filepath.Join(tmp, "AGENTS.md"), "NambaAI\n")

	payload := map[string]any{
		"hook_event_name":        "Stop",
		"cwd":                    filepath.ToSlash(tmp),
		"last_assistant_message": strings.Repeat("namba missing report frame ", 30),
	}
	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}

	cmd := exec.Command(pythonPath, filepath.Join(tmp, ".codex", "hooks", "namba_codex_guard.py"))
	cmd.Dir = tmp
	cmd.Env = append(os.Environ(), "PYTHONIOENCODING=cp949")
	cmd.Stdin = strings.NewReader(string(body))
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("expected CP949 stdout hook execution to succeed, err=%v output=%s", err, output)
	}
	got := string(output)
	for _, want := range []string{`"decision":"block"`, `\ud83e\udded`, `\ud83d\udee0`} {
		if !strings.Contains(got, want) {
			t.Fatalf("expected ASCII-safe hook output to contain %q, got %q", want, got)
		}
	}
	if strings.Contains(got, "🧭") || strings.Contains(got, "🛠") {
		t.Fatalf("expected hook output to escape non-ASCII under CP949, got %q", got)
	}
}

func codexHookCommandForEvent(t *testing.T, hooksJSON, event string) string {
	t.Helper()

	var parsed struct {
		Hooks map[string][]struct {
			Hooks []struct {
				Command string `json:"command"`
			} `json:"hooks"`
		} `json:"hooks"`
	}
	if err := json.Unmarshal([]byte(hooksJSON), &parsed); err != nil {
		t.Fatalf("hooks JSON should parse: %v", err)
	}
	registrations := parsed.Hooks[event]
	if len(registrations) != 1 || len(registrations[0].Hooks) != 1 {
		t.Fatalf("expected exactly one command hook for %s, got %+v", event, registrations)
	}
	return registrations[0].Hooks[0].Command
}
