package namba

import (
	"bytes"
	"context"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunSyncWritesRunModeDocs(t *testing.T) {
	tmp := t.TempDir()
	app := NewApp(&bytes.Buffer{}, &bytes.Buffer{})
	if err := app.Run(context.Background(), []string{"init", tmp, "--yes"}); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	restore := chdirExecution(t, tmp)
	defer restore()

	if err := app.Run(context.Background(), []string{"sync"}); err != nil {
		t.Fatalf("sync failed: %v", err)
	}

	readme := mustReadFile(t, filepath.Join(tmp, "README.md"))
	for _, want := range []string{"`--solo` for a single runner in one workspace", "`--team` for same-workspace multi-agent execution", "`--parallel` for worktree fan-out/fan-in", "`namba queue start SPEC-001..SPEC-003`", "## Command Skills In Codex", "## 🗺️ Skill To Command Mapping", "## Custom Agents In Codex", "`$namba-help`", "`$namba-run`", "`$namba-queue`", "`$namba-harness`", "`$namba-plan-review`", "`$namba-review-resolve`", "`$namba-release`", "`$namba-sync`", "`$namba-pr`", "`$namba-regen`", "`$namba-plan-pm-review`", "`$namba-plan-eng-review`", "`$namba-plan-design-review`", "`namba-product-manager`", "`namba-plan-reviewer`", "`namba-mobile-engineer`", "`namba-designer`", "`namba-data-engineer`", "`namba-security-engineer`", "`namba harness \"description\"`", "`namba fix --command plan \"issue description\"`", "direct repair in the current workspace", "frontend-brief.md", "`frontend-major`"} {
		if !strings.Contains(readme, want) {
			t.Fatalf("expected README to contain %q, got %q", want, readme)
		}
	}

	workflowGuide := mustReadFile(t, filepath.Join(tmp, "docs", "workflow-guide.md"))
	for _, want := range []string{"## Run modes", "## SPEC queue conveyor", "## Role routing", "## Review readiness", "## Planning commands", "## PR and merge flow", "`$namba-help`", "`namba project`: refresh current repository docs and codemaps without creating a SPEC package.", "`namba codex access`", "`namba harness \"description\"`: create the next harness-oriented SPEC package", "`namba run SPEC-XXX --solo`: a single runner in one workspace.", "`namba run SPEC-XXX --team`: same-workspace multi-agent execution.", "`namba run SPEC-XXX --parallel`: Namba-managed git worktree fan-out/fan-in, not Codex subagent orchestration.", "`namba queue status`", "`waiting_for_land`", "`namba fix \"issue description\"`: direct repair in the current workspace.", "`namba fix --command plan \"issue description\"`: create a bugfix SPEC package plus review artifacts.", "`namba plan`, `namba harness`, and `namba fix --command plan` seed", "fresh Codex session", "`namba-mobile-engineer`", "`namba-security-engineer`", "`$namba-plan-review`", "`$namba-plan-pm-review`", "frontend-brief.md", "`frontend-major`"} {
		if !strings.Contains(workflowGuide, want) {
			t.Fatalf("expected workflow guide to contain %q, got %q", want, workflowGuide)
		}
	}
}

func TestRenderNambaCLIWorkflowGuideIncludesRoleRouting(t *testing.T) {
	guide := renderReadmeGuide("en", "workflow-guide", projectConfig{}, initProfile{}, docsConfig{ReadmeProfile: readmeProfileNambaCLI})
	for _, want := range []string{"## Role routing", "## SPEC queue conveyor", "## Planning commands", "## PR and merge flow", "`$namba-help`", "`namba codex access`", "`namba harness \"description\"`: create the next harness-oriented SPEC package", "`namba run SPEC-XXX --team`: same-workspace multi-agent execution.", "`namba run SPEC-XXX --solo`: a single runner in one workspace.", "`namba queue start SPEC-001..SPEC-003`", "`namba fix --command plan \"issue description\"`: create a bugfix SPEC package plus review artifacts.", "`namba-mobile-engineer`", "`namba-security-engineer`", "`namba-reviewer`", "`$namba-plan-review`", "fresh Codex session"} {
		if !strings.Contains(guide, want) {
			t.Fatalf("expected namba-cli workflow guide to contain %q, got %q", want, guide)
		}
	}
}

func TestRenderNambaCLICoachGuidanceIsExposedInSyncedDocs(t *testing.T) {
	outputs := buildReadmeOutputs(projectConfig{}, initProfile{}, docsConfig{
		ManageReadme:        true,
		ReadmeProfile:       readmeProfileNambaCLI,
		DefaultLanguage:     "en",
		AdditionalLanguages: []string{"ko", "ja", "zh"},
	})

	for _, lang := range []string{"en", "ko", "ja", "zh"} {
		root := outputs[readmePath(lang)]
		for _, want := range []string{"`$namba-coach`", "$namba-coach"} {
			if !strings.Contains(root, want) {
				t.Fatalf("%s README missing coach routing anchor %q: %q", lang, want, root)
			}
		}

		guide := outputs[guidePath("workflow-guide", lang)]
		for _, want := range []string{"`$namba-coach`", "$namba-coach"} {
			if !strings.Contains(guide, want) {
				t.Fatalf("%s workflow guide missing coach routing anchor %q: %q", lang, want, guide)
			}
		}
	}
}

func TestRenderReadmeGuidePreludePreservesLocalizedHeaderAndLinks(t *testing.T) {
	cases := []struct {
		lang  string
		guide string
	}{
		{lang: "en", guide: "getting-started"},
		{lang: "ko", guide: "workflow-guide"},
	}

	for _, tc := range cases {
		got := strings.Join(renderReadmeGuidePrelude(tc.lang, tc.guide), "\n")
		want := strings.Join([]string{
			renderGeneratedDocHeader(),
			"# " + localizeGuideLabel(tc.lang, tc.guide),
			"",
			renderLanguageLinks("../"),
			"",
			renderDocLinkBar(tc.lang),
			"",
		}, "\n")
		if got != want {
			t.Fatalf("renderReadmeGuidePrelude(%q, %q) mismatch\n got: %q\nwant: %q", tc.lang, tc.guide, got, want)
		}
	}
}

func TestBuildReadmeOutputsForNambaCLIIncludesLocalizedLifecycleDocs(t *testing.T) {
	outputs := buildReadmeOutputs(projectConfig{}, initProfile{}, docsConfig{
		ManageReadme:        true,
		ReadmeProfile:       readmeProfileNambaCLI,
		DefaultLanguage:     "en",
		AdditionalLanguages: []string{"ko", "ja", "zh"},
	})

	if got, want := len(outputs), 12; got != want {
		t.Fatalf("buildReadmeOutputs() produced %d outputs, want %d", got, want)
	}

	cases := []struct {
		lang                    string
		rootLifecycleHeading    string
		gettingStartedUpdate    string
		gettingStartedUninstall string
		workflowModesHeading    string
		workflowReviewHeading   string
	}{
		{
			lang:                    "en",
			rootLifecycleHeading:    "## 📦 Install, Update, and Uninstall",
			gettingStartedUpdate:    "## 2. Update",
			gettingStartedUninstall: "## 3. Uninstall",
			workflowModesHeading:    "## `namba run` modes",
			workflowReviewHeading:   "## Review readiness",
		},
		{
			lang:                    "ko",
			rootLifecycleHeading:    "## 📦 설치, 업데이트, 제거",
			gettingStartedUpdate:    "## 2. 업데이트",
			gettingStartedUninstall: "## 3. 제거",
			workflowModesHeading:    "## `namba run` 모드",
			workflowReviewHeading:   "## 리뷰 준비도",
		},
		{
			lang:                    "ja",
			rootLifecycleHeading:    "## 📦 インストール、アップデート、アンインストール",
			gettingStartedUpdate:    "## 2. アップデート",
			gettingStartedUninstall: "## 3. アンインストール",
			workflowModesHeading:    "## `namba run` モード",
			workflowReviewHeading:   "## レビュー準備度",
		},
		{
			lang:                    "zh",
			rootLifecycleHeading:    "## 📦 安装、更新与卸载",
			gettingStartedUpdate:    "## 2. 更新",
			gettingStartedUninstall: "## 3. 卸载",
			workflowModesHeading:    "## `namba run` 模式",
			workflowReviewHeading:   "## 评审准备度",
		},
	}

	for _, tc := range cases {
		root := outputs[readmePath(tc.lang)]
		for _, want := range []string{
			tc.rootLifecycleHeading,
			"`$namba-help`",
			"`$namba-run`",
			"`$namba-harness`",
			"`$namba-plan-review`",
			"`$namba-review-resolve`",
			"`$namba-release`",
			"`$namba-update`",
			"`namba-mobile-engineer`",
			"`namba-plan-reviewer`",
			"`namba-security-engineer`",
			"`namba update --version vX.Y.Z`",
			"`codex update`",
			"namba pr",
			"namba land",
			nambaWindowsBinaryPath,
			nambaUnixBinaryPath,
		} {
			if !strings.Contains(root, want) {
				t.Fatalf("%s README missing %q: %q", tc.lang, want, root)
			}
		}

		gettingStarted := outputs[guidePath("getting-started", tc.lang)]
		for _, want := range []string{
			tc.gettingStartedUpdate,
			tc.gettingStartedUninstall,
			"`namba update`",
			"`namba update --version vX.Y.Z`",
			"`codex update`",
			"`namba codex access`",
			"`NAMBA_INSTALL_DIR`",
			"namba pr",
			"namba land",
			nambaWindowsBinaryPath,
			nambaUnixBinaryPath,
		} {
			if !strings.Contains(gettingStarted, want) {
				t.Fatalf("%s getting-started guide missing %q: %q", tc.lang, want, gettingStarted)
			}
		}

		workflowGuide := outputs[guidePath("workflow-guide", tc.lang)]
		for _, want := range []string{
			tc.workflowModesHeading,
			tc.workflowReviewHeading,
			"`namba codex access`",
			"`namba harness",
			"`$namba-help`",
			"`$namba-plan-review`",
			"`namba run SPEC-XXX --team`",
			"`namba queue status`",
			"`waiting_for_land`",
			"`namba fix --command plan",
			"`namba fix --command run",
			"`namba-reviewer`",
			"`$namba-plan-eng-review`",
			"`namba pr`",
			"`namba land`",
			"`codex update`",
		} {
			if !strings.Contains(workflowGuide, want) {
				t.Fatalf("%s workflow guide missing %q: %q", tc.lang, want, workflowGuide)
			}
		}

		if tc.lang == "en" {
			for _, want := range []string{
				"`namba run SPEC-XXX --solo`: a single runner in one workspace.",
				"`namba run SPEC-XXX --team`: same-workspace multi-agent execution.",
				"`namba run SPEC-XXX --parallel`: Namba-managed git worktree fan-out/fan-in, not Codex subagent orchestration.",
				"Codex subagent threads are controlled by `.codex/config.toml [agents].max_threads = 5`; Namba worktree workers stay separate at `.namba/config/sections/workflow.yaml max_parallel_workers: 3`.",
				"Persisted Codex `/goal` workflows are a future orchestration candidate, not a required Namba runtime dependency.",
				"fresh Codex session",
			} {
				if !strings.Contains(workflowGuide, want) {
					t.Fatalf("%s workflow guide missing %q: %q", tc.lang, want, workflowGuide)
				}
			}
		}
	}
}

func TestRenderNambaCLIGettingStartedSectionHelpersPreserveLocalizedAnchors(t *testing.T) {
	cases := []struct {
		lang             string
		installHeading   string
		bootstrapHeading string
		flowHeading      string
		nextDocsHeading  string
	}{
		{
			lang:             "en",
			installHeading:   "## 1. Install",
			bootstrapHeading: "## 4. Bootstrap a new repository",
			flowHeading:      "## 5. Run the basic Codex flow",
			nextDocsHeading:  "## 6. Read next",
		},
		{
			lang:             "ko",
			installHeading:   "## 1. 설치",
			bootstrapHeading: "## 4. 새 저장소 부트스트랩",
			flowHeading:      "## 5. 기본 Codex 흐름 실행",
			nextDocsHeading:  "## 6. 다음 문서",
		},
		{
			lang:             "ja",
			installHeading:   "## 1. インストール",
			bootstrapHeading: "## 4. 新しいリポジトリをブートストラップ",
			flowHeading:      "## 5. 基本の Codex フローを実行",
			nextDocsHeading:  "## 6. 次に読む文書",
		},
		{
			lang:             "zh",
			installHeading:   "## 1. 安装",
			bootstrapHeading: "## 4. 初始化新仓库",
			flowHeading:      "## 5. 运行基础 Codex 流程",
			nextDocsHeading:  "## 6. 接下来阅读",
		},
	}

	for _, tc := range cases {
		install := strings.Join(renderNambaCLIGettingStartedInstallSection(tc.lang), "\n")
		for _, want := range []string{tc.installHeading, nambaInstallPowerShell, nambaInstallShell, "`NAMBA_INSTALL_DIR`"} {
			if !strings.Contains(install, want) {
				t.Fatalf("%s install section missing %q: %q", tc.lang, want, install)
			}
		}

		bootstrap := strings.Join(renderNambaCLIGettingStartedBootstrapSection(tc.lang), "\n")
		for _, want := range []string{tc.bootstrapHeading, "mkdir my-project", "namba init .", "namba codex access"} {
			if !strings.Contains(bootstrap, want) {
				t.Fatalf("%s bootstrap section missing %q: %q", tc.lang, want, bootstrap)
			}
		}

		flow := strings.Join(renderNambaCLIGettingStartedBasicFlowSection(tc.lang), "\n")
		for _, want := range []string{tc.flowHeading, "namba project", "namba run SPEC-001", "namba land"} {
			if !strings.Contains(flow, want) {
				t.Fatalf("%s basic flow section missing %q: %q", tc.lang, want, flow)
			}
		}

		nextDocs := strings.Join(renderNambaCLIGettingStartedNextDocsSection(tc.lang), "\n")
		for _, want := range []string{tc.nextDocsHeading, guideFilename("workflow-guide", tc.lang), "Codex Upstream Reference"} {
			if !strings.Contains(nextDocs, want) {
				t.Fatalf("%s next-docs section missing %q: %q", tc.lang, want, nextDocs)
			}
		}
	}
}

func TestRenderNambaCLIGettingStartedSectionHelpersFallbackToEnglish(t *testing.T) {
	install := strings.Join(renderNambaCLIGettingStartedInstallSection("fr"), "\n")
	if !strings.Contains(install, "## 1. Install") {
		t.Fatalf("install section fallback missing English heading: %q", install)
	}

	nextDocs := strings.Join(renderNambaCLIGettingStartedNextDocsSection("fr"), "\n")
	for _, want := range []string{"## 6. Read next", guideFilename("workflow-guide", "en")} {
		if !strings.Contains(nextDocs, want) {
			t.Fatalf("next-docs section fallback missing %q: %q", want, nextDocs)
		}
	}
}

func TestRenderManagedProjectGettingStartedSectionHelpersPreserveLocalizedAnchors(t *testing.T) {
	cases := []struct {
		lang             string
		projectName      string
		openHeading      string
		refreshHeading   string
		workHeading      string
		reviewHeading    string
		implementHeading string
		handoffHeading   string
	}{
		{
			lang:             "en",
			projectName:      "demo-repo",
			openHeading:      "## 1. Open the repository",
			refreshHeading:   "## 2. Refresh current context",
			workHeading:      "## 3. Create a work package",
			reviewHeading:    "## 4. Review plan readiness",
			implementHeading: "## 5. Implement and sync",
			handoffHeading:   "## 6. Hand off and merge",
		},
		{
			lang:             "ko",
			projectName:      "demo-repo",
			openHeading:      "## 1. 저장소 열기",
			refreshHeading:   "## 2. 현재 컨텍스트 갱신",
			workHeading:      "## 3. 작업 패키지 만들기",
			reviewHeading:    "## 4. 계획 준비도 검토",
			implementHeading: "## 5. 구현 및 sync",
			handoffHeading:   "## 6. 인계 및 머지",
		},
		{
			lang:             "ja",
			projectName:      "demo-repo",
			openHeading:      "## 1. リポジトリを開く",
			refreshHeading:   "## 2. 現在のコンテキストを更新",
			workHeading:      "## 3. 作業パッケージを作る",
			reviewHeading:    "## 4. 計画の準備度を確認",
			implementHeading: "## 5. 実装と sync",
			handoffHeading:   "## 6. 引き渡しとマージ",
		},
		{
			lang:             "zh",
			projectName:      "demo-repo",
			openHeading:      "## 1. 打开仓库",
			refreshHeading:   "## 2. 刷新当前上下文",
			workHeading:      "## 3. 创建工作包",
			reviewHeading:    "## 4. 检查计划准备度",
			implementHeading: "## 5. 实现并 sync",
			handoffHeading:   "## 6. 交接与合并",
		},
	}

	for _, tc := range cases {
		open := strings.Join(renderManagedProjectGettingStartedOpenSection(tc.lang, tc.projectName), "\n")
		for _, want := range []string{tc.openHeading, tc.projectName, "`namba doctor`"} {
			if !strings.Contains(open, want) {
				t.Fatalf("%s open section missing %q: %q", tc.lang, want, open)
			}
		}

		refresh := strings.Join(renderManagedProjectGettingStartedRefreshContextSection(tc.lang), "\n")
		for _, want := range []string{tc.refreshHeading, "namba project"} {
			if !strings.Contains(refresh, want) {
				t.Fatalf("%s refresh section missing %q: %q", tc.lang, want, refresh)
			}
		}

		workPackage := strings.Join(renderManagedProjectGettingStartedWorkPackageSection(tc.lang), "\n")
		for _, want := range []string{tc.workHeading, "namba plan", "namba harness", "namba fix --command plan", "namba fix --command run", "namba codex access"} {
			if !strings.Contains(workPackage, want) {
				t.Fatalf("%s work-package section missing %q: %q", tc.lang, want, workPackage)
			}
		}

		review := strings.Join(renderManagedProjectGettingStartedReviewReadinessSection(tc.lang), "\n")
		for _, want := range []string{tc.reviewHeading, "$namba-plan-pm-review", "readiness.md"} {
			if !strings.Contains(review, want) {
				t.Fatalf("%s review-readiness section missing %q: %q", tc.lang, want, review)
			}
		}

		implement := strings.Join(renderManagedProjectGettingStartedImplementSection(tc.lang), "\n")
		for _, want := range []string{tc.implementHeading, "namba run SPEC-001", "namba sync"} {
			if !strings.Contains(implement, want) {
				t.Fatalf("%s implement section missing %q: %q", tc.lang, want, implement)
			}
		}

		handoff := strings.Join(renderManagedProjectGettingStartedHandoffSection(tc.lang), "\n")
		for _, want := range []string{tc.handoffHeading, "namba pr \"work description\"", "namba land"} {
			if !strings.Contains(handoff, want) {
				t.Fatalf("%s handoff section missing %q: %q", tc.lang, want, handoff)
			}
		}
	}
}

func TestRenderManagedProjectGettingStartedSectionHelpersFallbackToEnglish(t *testing.T) {
	open := strings.Join(renderManagedProjectGettingStartedOpenSection("fr", "demo-repo"), "\n")
	for _, want := range []string{"## 1. Open the repository", "demo-repo", "`namba doctor`"} {
		if !strings.Contains(open, want) {
			t.Fatalf("open section fallback missing %q: %q", want, open)
		}
	}

	handoff := strings.Join(renderManagedProjectGettingStartedHandoffSection("fr"), "\n")
	for _, want := range []string{"## 6. Hand off and merge", "namba pr \"work description\"", "namba land"} {
		if !strings.Contains(handoff, want) {
			t.Fatalf("handoff section fallback missing %q: %q", want, handoff)
		}
	}
}

func TestRenderManagedProjectRootSectionHelpersPreserveLocalizedAnchors(t *testing.T) {
	profile := initProfile{
		DocumentationLanguage: "ko",
		PRBaseBranch:          "develop",
		CodexReviewComment:    "@codex review",
	}

	cases := []struct {
		lang                   string
		workSummaryHeading     string
		quickStartHeading      string
		readMoreHeading        string
		currentDefaultsHeading string
		workingLanguageNeedle  string
		prBaseNeedle           string
		reviewCommentNeedle    string
	}{
		{
			lang:                   "en",
			workSummaryHeading:     "## What You Can Do In This Repository",
			quickStartHeading:      "## Start Fast",
			readMoreHeading:        "## Read More",
			currentDefaultsHeading: "## Current Defaults",
			workingLanguageNeedle:  "Working language: Korean",
			prBaseNeedle:           "PR base branch: `develop`",
			reviewCommentNeedle:    "Codex review comment: `@codex review`",
		},
		{
			lang:                   "ko",
			workSummaryHeading:     "## 이 저장소에서 할 수 있는 일",
			quickStartHeading:      "## 빠르게 시작하기",
			readMoreHeading:        "## 더 읽기",
			currentDefaultsHeading: "## 현재 기본값",
			workingLanguageNeedle:  "작업 언어: Korean",
			prBaseNeedle:           "PR 대상 브랜치: `develop`",
			reviewCommentNeedle:    "Codex 리뷰 코멘트: `@codex review`",
		},
		{
			lang:                   "ja",
			workSummaryHeading:     "## このリポジトリでできること",
			quickStartHeading:      "## すぐ始める",
			readMoreHeading:        "## さらに読む",
			currentDefaultsHeading: "## 現在の既定値",
			workingLanguageNeedle:  "作業言語: Korean",
			prBaseNeedle:           "PR 対象ブランチ: `develop`",
			reviewCommentNeedle:    "Codex review コメント: `@codex review`",
		},
		{
			lang:                   "zh",
			workSummaryHeading:     "## 这个仓库里可以做什么",
			quickStartHeading:      "## 快速开始",
			readMoreHeading:        "## 继续阅读",
			currentDefaultsHeading: "## 当前默认值",
			workingLanguageNeedle:  "工作语言: Korean",
			prBaseNeedle:           "PR 目标分支: `develop`",
			reviewCommentNeedle:    "Codex review 评论: `@codex review`",
		},
	}

	for _, tc := range cases {
		workSummary := strings.Join(renderManagedProjectRootWhatYouCanDoSection(tc.lang), "\n")
		for _, want := range []string{tc.workSummaryHeading, "namba project", "namba plan", "namba fix --command plan", "$namba-help", "$namba-plan-review", "namba run SPEC-XXX"} {
			if !strings.Contains(workSummary, want) {
				t.Fatalf("%s work-summary section missing %q: %q", tc.lang, want, workSummary)
			}
		}

		quickStart := strings.Join(renderManagedProjectRootQuickStartSection(tc.lang), "\n")
		for _, want := range []string{tc.quickStartHeading, "namba project", "namba plan \"work description\"", "namba run SPEC-001", "namba land"} {
			if !strings.Contains(quickStart, want) {
				t.Fatalf("%s quick-start section missing %q: %q", tc.lang, want, quickStart)
			}
		}

		readMore := strings.Join(renderManagedProjectRootReadMoreSection(tc.lang), "\n")
		for _, want := range []string{tc.readMoreHeading, localizeGuideLabel(tc.lang, "getting-started"), localizeGuideLabel(tc.lang, "workflow-guide"), "Codex Upstream Reference"} {
			if !strings.Contains(readMore, want) {
				t.Fatalf("%s read-more section missing %q: %q", tc.lang, want, readMore)
			}
		}

		currentDefaults := strings.Join(renderManagedProjectRootCurrentDefaultsSection(tc.lang, profile), "\n")
		for _, want := range []string{tc.currentDefaultsHeading, tc.workingLanguageNeedle, tc.prBaseNeedle, tc.reviewCommentNeedle, "source of truth: `.namba/`"} {
			if !strings.Contains(currentDefaults, want) {
				t.Fatalf("%s current-defaults section missing %q: %q", tc.lang, want, currentDefaults)
			}
		}
	}
}

func TestRenderManagedProjectRootSectionHelpersFallbackToEnglish(t *testing.T) {
	workSummary := strings.Join(renderManagedProjectRootWhatYouCanDoSection("fr"), "\n")
	for _, want := range []string{"## What You Can Do In This Repository", "namba project", "namba fix --command plan", "$namba-plan-review", "namba run SPEC-XXX"} {
		if !strings.Contains(workSummary, want) {
			t.Fatalf("work-summary section fallback missing %q: %q", want, workSummary)
		}
	}

	quickStart := strings.Join(renderManagedProjectRootQuickStartSection("fr"), "\n")
	for _, want := range []string{"## Start Fast", "namba project", "namba run SPEC-001"} {
		if !strings.Contains(quickStart, want) {
			t.Fatalf("quick-start section fallback missing %q: %q", want, quickStart)
		}
	}

	readMore := strings.Join(renderManagedProjectRootReadMoreSection("fr"), "\n")
	for _, want := range []string{"## Read More", localizeGuideLabel("fr", "getting-started"), "Codex Upstream Reference"} {
		if !strings.Contains(readMore, want) {
			t.Fatalf("read-more section fallback missing %q: %q", want, readMore)
		}
	}

	currentDefaults := strings.Join(renderManagedProjectRootCurrentDefaultsSection("fr", initProfile{
		DocumentationLanguage: "en",
		PRBaseBranch:          "main",
		CodexReviewComment:    "@codex review",
	}), "\n")
	for _, want := range []string{"## Current Defaults", "Working language: English", "PR base branch: `main`", "Codex review comment: `@codex review`"} {
		if !strings.Contains(currentDefaults, want) {
			t.Fatalf("current-defaults section fallback missing %q: %q", want, currentDefaults)
		}
	}
}

func TestRenderManagedProjectRootWhatYouCanDoSectionPreservesLocalizedAnchors(t *testing.T) {
	cases := []struct {
		lang      string
		heading   string
		reviewRef string
	}{
		{lang: "en", heading: "## What You Can Do In This Repository", reviewRef: "Codex review"},
		{lang: "ko", heading: "## 이 저장소에서 할 수 있는 일", reviewRef: "Codex review"},
		{lang: "ja", heading: "## このリポジトリでできること", reviewRef: "Codex review"},
		{lang: "zh", heading: "## 这个仓库里可以做什么", reviewRef: "Codex review"},
	}

	for _, tc := range cases {
		section := strings.Join(renderManagedProjectRootWhatYouCanDoSection(tc.lang), "\n")
		for _, want := range []string{tc.heading, "namba project", "namba fix --command plan", "$namba-help", "$namba-plan-review", "$namba-plan-pm-review", "namba run SPEC-XXX", "--parallel", tc.reviewRef} {
			if !strings.Contains(section, want) {
				t.Fatalf("%s what-you-can-do section missing %q: %q", tc.lang, want, section)
			}
		}
	}
}

func TestRenderManagedProjectRootWhatYouCanDoSectionFallbackToEnglish(t *testing.T) {
	section := strings.Join(renderManagedProjectRootWhatYouCanDoSection("fr"), "\n")
	for _, want := range []string{"## What You Can Do In This Repository", "namba project", "namba fix --command plan", "$namba-help", "$namba-plan-pm-review", "namba run SPEC-XXX", "--parallel", "Codex review"} {
		if !strings.Contains(section, want) {
			t.Fatalf("what-you-can-do section fallback missing %q: %q", want, section)
		}
	}
}

func TestRenderManagedProjectRootEnglishCommandSurfaceHelpersPreserveAnchors(t *testing.T) {
	commandSkills := strings.Join(renderManagedProjectRootCommandSkillsSection("en"), "\n")
	for _, want := range []string{"## Command Skills In Codex", "`$namba-help`", "`$namba-queue`", "`$namba-plan-review`", "`$namba-review-resolve`", "`$namba-release`", "`$namba-update`", "right Namba entry point", "reviewable bugfix SPEC"} {
		if !strings.Contains(commandSkills, want) {
			t.Fatalf("managed-project command-skills section missing %q: %q", want, commandSkills)
		}
	}

	mapping := strings.Join(renderManagedProjectRootSkillMappingSection("en"), "\n")
	wantMapping := strings.Join(renderNambaCLIRootSkillMappingSection("en"), "\n")
	if mapping != wantMapping {
		t.Fatalf("managed-project skill-mapping section should reuse namba-cli mapping helper\n got: %q\nwant: %q", mapping, wantMapping)
	}

	customAgents := strings.Join(renderManagedProjectRootCustomAgentsSection("en"), "\n")
	for _, want := range []string{"## Custom Agents In Codex", "`namba-product-manager`", "`namba-plan-reviewer`", "`namba-security-engineer`", "`namba-implementer`", "review set is coherent enough"} {
		if !strings.Contains(customAgents, want) {
			t.Fatalf("managed-project custom-agents section missing %q: %q", want, customAgents)
		}
	}
}

func TestRenderManagedProjectRootEnglishCommandSurfaceHelpersFallbackToEnglish(t *testing.T) {
	commandSkills := strings.Join(renderManagedProjectRootCommandSkillsSection("fr"), "\n")
	if !strings.Contains(commandSkills, "## Command Skills In Codex") {
		t.Fatalf("managed-project command-skills fallback missing English heading: %q", commandSkills)
	}

	mapping := strings.Join(renderManagedProjectRootSkillMappingSection("fr"), "\n")
	wantMapping := strings.Join(renderNambaCLIRootSkillMappingSection("fr"), "\n")
	if mapping != wantMapping {
		t.Fatalf("managed-project skill-mapping fallback should reuse namba-cli mapping helper\n got: %q\nwant: %q", mapping, wantMapping)
	}

	customAgents := strings.Join(renderManagedProjectRootCustomAgentsSection("fr"), "\n")
	if !strings.Contains(customAgents, "## Custom Agents In Codex") {
		t.Fatalf("managed-project custom-agents fallback missing English heading: %q", customAgents)
	}
}

func TestRenderManagedProjectWorkflowGuideCollaborationRulesSectionPreservesLocalizedAnchors(t *testing.T) {
	profile := initProfile{
		BranchBase:         "main",
		PRBaseBranch:       "develop",
		PRLanguage:         "ko",
		CodexReviewComment: "@codex review",
	}

	cases := []struct {
		lang          string
		heading       string
		prLanguage    string
		handoffNeedle string
	}{
		{lang: "en", heading: "## Collaboration rules", prLanguage: "PR language: Korean", handoffNeedle: "`namba sync` stays local; `namba pr` and `namba land` handle GitHub handoff plus merge."},
		{lang: "ko", heading: "## 협업 규칙", prLanguage: "PR language: Korean", handoffNeedle: "`namba sync`는 로컬 산출물 갱신에 머물고, `namba pr`와 `namba land`가 GitHub handoff와 merge를 담당합니다."},
		{lang: "ja", heading: "## 協業ルール", prLanguage: "PR language: Korean", handoffNeedle: "`namba sync` はローカル成果物の更新にとどまり、`namba pr` と `namba land` が GitHub handoff と merge を担当します。"},
		{lang: "zh", heading: "## 协作规则", prLanguage: "PR language: Korean", handoffNeedle: "`namba sync` 只负责刷新本地产物，`namba pr` 和 `namba land` 负责 GitHub handoff 和 merge。"},
	}

	for _, tc := range cases {
		section := strings.Join(renderManagedProjectWorkflowGuideCollaborationRulesSection(tc.lang, profile), "\n")
		for _, want := range []string{tc.heading, "base branch: `main`", "PR base: `develop`", tc.prLanguage, "review request: `@codex review`", tc.handoffNeedle} {
			if !strings.Contains(section, want) {
				t.Fatalf("%s collaboration-rules section missing %q: %q", tc.lang, want, section)
			}
		}
	}
}

func TestRenderManagedProjectWorkflowGuideCollaborationRulesSectionFallbackToEnglish(t *testing.T) {
	section := strings.Join(renderManagedProjectWorkflowGuideCollaborationRulesSection("fr", initProfile{
		BranchBase:         "main",
		PRBaseBranch:       "main",
		PRLanguage:         "en",
		CodexReviewComment: "@codex review",
	}), "\n")

	for _, want := range []string{"## Collaboration rules", "base branch: `main`", "PR base: `main`", "PR language: English", "review request: `@codex review`", "`namba sync` stays local; `namba pr` and `namba land` handle GitHub handoff plus merge."} {
		if !strings.Contains(section, want) {
			t.Fatalf("collaboration-rules section fallback missing %q: %q", want, section)
		}
	}
}

func TestRenderManagedProjectWorkflowGuideKeyLocationsAndWorkOrderSectionsPreserveLocalizedAnchors(t *testing.T) {
	cases := []struct {
		lang                string
		keyLocationsHeading string
		workOrderHeading    string
		workOrderThirdStep  string
	}{
		{
			lang:                "en",
			keyLocationsHeading: "## Key locations",
			workOrderHeading:    "## Work order",
			workOrderThirdStep:  "3. Run the relevant plan-review skills and refresh `.namba/specs/<SPEC>/reviews/readiness.md` when the SPEC needs product, engineering, or design critique",
		},
		{
			lang:                "ko",
			keyLocationsHeading: "## 핵심 위치",
			workOrderHeading:    "## 작업 순서",
			workOrderThirdStep:  "3. 필요하면 관련 plan-review skill을 실행하고 `.namba/specs/<SPEC>/reviews/readiness.md`를 갱신",
		},
		{
			lang:                "ja",
			keyLocationsHeading: "## 主要な場所",
			workOrderHeading:    "## 作業順序",
			workOrderThirdStep:  "3. 必要なら関連する plan-review skill を実行し、`.namba/specs/<SPEC>/reviews/readiness.md` を更新",
		},
		{
			lang:                "zh",
			keyLocationsHeading: "## 关键位置",
			workOrderHeading:    "## 工作顺序",
			workOrderThirdStep:  "3. 必要时执行相关 plan-review skill，并更新 `.namba/specs/<SPEC>/reviews/readiness.md`",
		},
	}

	for _, tc := range cases {
		keyLocations := strings.Join(renderManagedProjectWorkflowGuideKeyLocationsSection(tc.lang), "\n")
		for _, want := range []string{tc.keyLocationsHeading, ".namba/", ".namba/specs/<SPEC>/reviews/", ".agents/skills/", ".codex/agents/*.toml"} {
			if !strings.Contains(keyLocations, want) {
				t.Fatalf("%s key-locations section missing %q: %q", tc.lang, want, keyLocations)
			}
		}

		workOrder := strings.Join(renderManagedProjectWorkflowGuideWorkOrderSection(tc.lang), "\n")
		for _, want := range []string{tc.workOrderHeading, "1. `namba project`", "4. `namba run SPEC-XXX`", "7. `namba land`", tc.workOrderThirdStep} {
			if !strings.Contains(workOrder, want) {
				t.Fatalf("%s work-order section missing %q: %q", tc.lang, want, workOrder)
			}
		}
	}
}

func TestRenderManagedProjectWorkflowGuideKeyLocationsAndWorkOrderSectionsFallbackToEnglish(t *testing.T) {
	keyLocations := strings.Join(renderManagedProjectWorkflowGuideKeyLocationsSection("fr"), "\n")
	for _, want := range []string{"## Key locations", ".namba/", ".agents/skills/", ".codex/agents/*.toml"} {
		if !strings.Contains(keyLocations, want) {
			t.Fatalf("key-locations section fallback missing %q: %q", want, keyLocations)
		}
	}

	workOrder := strings.Join(renderManagedProjectWorkflowGuideWorkOrderSection("fr"), "\n")
	for _, want := range []string{"## Work order", "1. `namba project`", "4. `namba run SPEC-XXX`", "7. `namba land`"} {
		if !strings.Contains(workOrder, want) {
			t.Fatalf("work-order section fallback missing %q: %q", want, workOrder)
		}
	}
}

func TestRenderManagedProjectWorkflowGuideRunModesSectionPreservesLocalizedAnchors(t *testing.T) {
	cases := []struct {
		lang       string
		heading    string
		teamLine   string
		parallelLn string
		workerLine string
	}{
		{
			lang:       "en",
			heading:    "## Run modes",
			teamLine:   "- `namba run SPEC-XXX --team`: same-workspace multi-agent execution.",
			parallelLn: "- `namba run SPEC-XXX --parallel`: Namba-managed git worktree fan-out/fan-in, not Codex subagent orchestration.",
			workerLine: "Codex subagent threads are controlled by `.codex/config.toml [agents].max_threads = 5`; Namba worktree workers stay separate at `.namba/config/sections/workflow.yaml max_parallel_workers: 3`.",
		},
		{
			lang:       "ko",
			heading:    "## Run 모드",
			teamLine:   "- `namba run SPEC-XXX --team`: 같은 workspace의 멀티에이전트 실행입니다.",
			parallelLn: "- `namba run SPEC-XXX --parallel`: Codex subagent orchestration이 아니라 Namba-managed git worktree fan-out/fan-in 입니다.",
			workerLine: "Codex subagent threads는 `.codex/config.toml [agents].max_threads = 5`, Namba worktree workers는 `.namba/config/sections/workflow.yaml max_parallel_workers: 3`로 따로 관리합니다.",
		},
		{
			lang:       "ja",
			heading:    "## Run モード",
			teamLine:   "- `namba run SPEC-XXX --team`: 同じ workspace の multi-agent execution です。",
			parallelLn: "- `namba run SPEC-XXX --parallel`: Codex subagent orchestration ではなく、Namba-managed git worktree fan-out/fan-in です。",
			workerLine: "Codex subagent threads は `.codex/config.toml [agents].max_threads = 5`、Namba worktree workers は `.namba/config/sections/workflow.yaml max_parallel_workers: 3` で別々に管理します。",
		},
		{
			lang:       "zh",
			heading:    "## Run 模式",
			teamLine:   "- `namba run SPEC-XXX --team`: 同一 workspace 内的 multi-agent execution。",
			parallelLn: "- `namba run SPEC-XXX --parallel`: 这不是 Codex subagent orchestration，而是 Namba-managed git worktree fan-out/fan-in。",
			workerLine: "Codex subagent threads 由 `.codex/config.toml [agents].max_threads = 5` 管理，Namba worktree workers 由 `.namba/config/sections/workflow.yaml max_parallel_workers: 3` 分开管理。",
		},
	}

	for _, tc := range cases {
		section := strings.Join(renderManagedProjectWorkflowGuideRunModesSection(tc.lang), "\n")
		for _, want := range []string{tc.heading, "- `namba run SPEC-XXX`: ", "- `namba run SPEC-XXX --solo`: ", tc.teamLine, tc.parallelLn, tc.workerLine} {
			if !strings.Contains(section, want) {
				t.Fatalf("%s run-modes section missing %q: %q", tc.lang, want, section)
			}
		}
	}
}

func TestRenderManagedProjectWorkflowGuideRunModesSectionFallbackToEnglish(t *testing.T) {
	section := strings.Join(renderManagedProjectWorkflowGuideRunModesSection("fr"), "\n")
	for _, want := range []string{
		"## Run modes",
		"- `namba run SPEC-XXX`: standard standalone Codex flow in one workspace.",
		"- `namba run SPEC-XXX --solo`: a single runner in one workspace.",
		"- `namba run SPEC-XXX --team`: same-workspace multi-agent execution.",
		"- `namba run SPEC-XXX --parallel`: Namba-managed git worktree fan-out/fan-in, not Codex subagent orchestration.",
		"Codex subagent threads are controlled by `.codex/config.toml [agents].max_threads = 5`; Namba worktree workers stay separate at `.namba/config/sections/workflow.yaml max_parallel_workers: 3`.",
	} {
		if !strings.Contains(section, want) {
			t.Fatalf("run-modes section fallback missing %q: %q", want, section)
		}
	}
}

func TestRenderManagedProjectWorkflowGuideReviewReadinessSectionPreservesLocalizedAnchors(t *testing.T) {
	cases := []struct {
		lang           string
		heading        string
		refreshNeedle  string
		advisoryNeedle string
	}{
		{
			lang:           "en",
			heading:        "## Review readiness",
			refreshNeedle:  "- If `namba regen` or `namba sync` changes generated instruction surfaces, start a fresh Codex session so the updated guidance is loaded before continuing a long repair loop.",
			advisoryNeedle: "- Review readiness is advisory by default: missing review passes are surfaced clearly by `namba run`, `namba sync`, and `namba pr`, but they do not silently become a hard gate.",
		},
		{
			lang:           "ko",
			heading:        "## 리뷰 준비도",
			refreshNeedle:  "- `namba regen` 또는 `namba sync`가 생성된 instruction surface를 바꾸면, 긴 repair loop 전에 fresh Codex session을 시작해 갱신된 지침을 다시 읽어오세요.",
			advisoryNeedle: "- 리뷰 통과가 빠져 있어도 advisory 상태를 유지합니다. `namba run`, `namba sync`, `namba pr`는 readiness 요약을 노출하지만 hard gate로 삼지는 않습니다.",
		},
		{
			lang:           "ja",
			heading:        "## レビュー準備度",
			refreshNeedle:  "- `namba regen` または `namba sync` が生成された instruction surface を変えた場合は、長い repair loop の前に fresh Codex session を開始して更新済み guidance を読み直してください。",
			advisoryNeedle: "- review pass が欠けていても advisory のままです。`namba run`、`namba sync`、`namba pr` は readiness summary を表示しますが hard gate にはしません。",
		},
		{
			lang:           "zh",
			heading:        "## 评审准备度",
			refreshNeedle:  "- 如果 `namba regen` 或 `namba sync` 改变了生成的 instruction surface，请在长 repair loop 之前启动 fresh Codex session，重新加载更新后的 guidance。",
			advisoryNeedle: "- 即使 review pass 缺失，默认也保持 advisory。`namba run`、`namba sync`、`namba pr` 会展示 readiness summary，但不会变成 hard gate。",
		},
	}

	for _, tc := range cases {
		section := strings.Join(renderManagedProjectWorkflowGuideReviewReadinessSection(tc.lang), "\n")
		for _, want := range []string{
			tc.heading,
			".namba/specs/<SPEC>/reviews/product.md",
			"$namba-plan-review",
			tc.refreshNeedle,
			tc.advisoryNeedle,
		} {
			if !strings.Contains(section, want) {
				t.Fatalf("%s review-readiness section missing %q: %q", tc.lang, want, section)
			}
		}
	}
}

func TestRenderManagedProjectWorkflowGuideReviewReadinessSectionFallbackToEnglish(t *testing.T) {
	section := strings.Join(renderManagedProjectWorkflowGuideReviewReadinessSection("fr"), "\n")
	for _, want := range []string{
		"## Review readiness",
		".namba/specs/<SPEC>/reviews/product.md",
		"$namba-plan-review",
		"- If `namba regen` or `namba sync` changes generated instruction surfaces, start a fresh Codex session so the updated guidance is loaded before continuing a long repair loop.",
		"- Review readiness is advisory by default: missing review passes are surfaced clearly by `namba run`, `namba sync`, and `namba pr`, but they do not silently become a hard gate.",
	} {
		if !strings.Contains(section, want) {
			t.Fatalf("review-readiness section fallback missing %q: %q", want, section)
		}
	}
}

func TestRenderManagedProjectWorkflowGuideRoleRoutingSectionPreservesLocalizedAnchors(t *testing.T) {
	cases := []struct {
		lang              string
		heading           string
		teamNeedle        string
		finalReviewNeedle string
	}{
		{
			lang:              "en",
			heading:           "## Role routing",
			teamNeedle:        "- `--team` keeps the work in one workspace while coordinating multiple specialists plus a final reviewer when acceptance spans multiple domains.",
			finalReviewNeedle: "- Keep the standalone runner as integrator and validation owner, and use `namba-reviewer` last when multiple specialists contribute.",
		},
		{
			lang:              "ko",
			heading:           "## 역할 라우팅",
			teamNeedle:        "- `--solo`는 한 명의 specialist로 위험을 줄일 수 있을 때만 분기하고, `--team`은 여러 specialist와 마지막 reviewer를 같은 workspace 안에서 조율합니다.",
			finalReviewNeedle: "- standalone runner는 integrator와 validation owner를 맡고, final acceptance는 `namba-reviewer`가 담당합니다.",
		},
		{
			lang:              "ja",
			heading:           "## ロールルーティング",
			teamNeedle:        "- `--solo` は 1 人の specialist でリスクを下げられる場合だけ分岐し、`--team` は複数 specialist と最終 reviewer を同じ workspace 内で調整します。",
			finalReviewNeedle: "- standalone runner は integrator と validation owner を担い、final acceptance は `namba-reviewer` が担当します。",
		},
		{
			lang:              "zh",
			heading:           "## 角色路由",
			teamNeedle:        "- `--solo` 只在一个 specialist 能明显降低风险时才分流，`--team` 会在同一 workspace 内协调多个 specialist 和最终 reviewer。",
			finalReviewNeedle: "- standalone runner 继续担任 integrator 和 validation owner，final acceptance 由 `namba-reviewer` 负责。",
		},
	}

	for _, tc := range cases {
		section := strings.Join(renderManagedProjectWorkflowGuideRoleRoutingSection(tc.lang), "\n")
		for _, want := range []string{
			tc.heading,
			"`namba-frontend-implementer`",
			"`namba-security-engineer`",
			tc.teamNeedle,
			tc.finalReviewNeedle,
		} {
			if !strings.Contains(section, want) {
				t.Fatalf("%s role-routing section missing %q: %q", tc.lang, want, section)
			}
		}
	}
}

func TestRenderManagedProjectWorkflowGuideRoleRoutingSectionFallbackToEnglish(t *testing.T) {
	section := strings.Join(renderManagedProjectWorkflowGuideRoleRoutingSection("fr"), "\n")
	for _, want := range []string{
		"## Role routing",
		"`namba-frontend-implementer`",
		"`namba-security-engineer`",
		"- `--team` keeps the work in one workspace while coordinating multiple specialists plus a final reviewer when acceptance spans multiple domains.",
		"- Keep the standalone runner as integrator and validation owner, and use `namba-reviewer` last when multiple specialists contribute.",
	} {
		if !strings.Contains(section, want) {
			t.Fatalf("role-routing section fallback missing %q: %q", want, section)
		}
	}
}

func TestRenderManagedProjectWorkflowGuidePlanAndFixSectionPreservesLocalizedAnchors(t *testing.T) {
	cases := []struct {
		lang       string
		heading    string
		leadNeedle string
		planNeedle string
		runNeedle  string
		helpNeedle string
	}{
		{
			lang:       "en",
			heading:    "## `namba plan` and `namba fix`",
			leadNeedle: "- `namba project`: refresh current repository docs and codemaps without creating a SPEC package.",
			planNeedle: "- `namba plan \"description\"`: create the next feature SPEC package and review artifacts.",
			runNeedle:  "- `namba fix --command run \"issue description\"`: explicit form of the same direct-repair path.",
			helpNeedle: "- `namba <command> --help`, `namba <command> -h`, and `namba help <command>`: read-only help flows for every top-level command; they must not mutate repository state.",
		},
		{
			lang:       "ko",
			heading:    "## `namba plan`과 `namba fix`",
			leadNeedle: "`$namba-help`",
			planNeedle: "- `namba plan \"description\"`: 다음 기능 SPEC 패키지와 review artifact를 만듭니다.",
			runNeedle:  "- `namba fix --command run \"issue description\"`: 같은 direct-repair 경로를 명시적으로 선택합니다.",
			helpNeedle: "- `namba <command> --help`, `namba <command> -h`, `namba help <command>`: 모든 top-level command에서 read-only help 경로로 끝나며 repo state를 바꾸지 않습니다.",
		},
		{
			lang:       "ja",
			heading:    "## `namba plan` と `namba fix`",
			leadNeedle: "`$namba-help`",
			planNeedle: "- `namba plan \"description\"`: 次の機能 SPEC パッケージと review artifact を作成します。",
			runNeedle:  "- `namba fix --command run \"issue description\"`: 同じ direct-repair path を明示的に選びます。",
			helpNeedle: "- `namba <command> --help`、`namba <command> -h`、`namba help <command>`: すべての top-level command で read-only の help に入り、repo state を変更しません。",
		},
		{
			lang:       "zh",
			heading:    "## `namba plan` 和 `namba fix`",
			leadNeedle: "`$namba-help`",
			planNeedle: "- `namba plan \"description\"`: 创建下一个功能 SPEC 包和 review artifact。",
			runNeedle:  "- `namba fix --command run \"issue description\"`: 显式选择同一个 direct-repair 路径。",
			helpNeedle: "- `namba <command> --help`、`namba <command> -h`、`namba help <command>`: 所有 top-level command 都会走 read-only help 路径，不会改变 repo state。",
		},
	}

	for _, tc := range cases {
		section := strings.Join(renderManagedProjectWorkflowGuidePlanAndFixSection(tc.lang), "\n")
		for _, want := range []string{
			tc.heading,
			tc.leadNeedle,
			tc.planNeedle,
			tc.runNeedle,
			tc.helpNeedle,
		} {
			if !strings.Contains(section, want) {
				t.Fatalf("%s plan-and-fix section missing %q: %q", tc.lang, want, section)
			}
		}
	}
}

func TestRenderManagedProjectWorkflowGuidePlanAndFixSectionFallbackToEnglish(t *testing.T) {
	section := strings.Join(renderManagedProjectWorkflowGuidePlanAndFixSection("fr"), "\n")
	for _, want := range []string{
		"## `namba plan` and `namba fix`",
		"- `namba plan \"description\"`: create the next feature SPEC package and review artifacts.",
		"- `namba fix --command run \"issue description\"`: explicit form of the same direct-repair path.",
		"- `namba <command> --help`, `namba <command> -h`, and `namba help <command>`: read-only help flows for every top-level command; they must not mutate repository state.",
	} {
		if !strings.Contains(section, want) {
			t.Fatalf("plan-and-fix section fallback missing %q: %q", want, section)
		}
	}
}

func TestRenderManagedProjectWorkflowGuidePlanningCommandsSectionPreservesEnglishAnchors(t *testing.T) {
	section := strings.Join(renderManagedProjectWorkflowGuidePlanningCommandsSection(), "\n")
	for _, want := range []string{
		"## Planning commands",
		"- `$namba-help`: read-only guidance on how to use NambaAI, which command or skill to choose next, and where the authoritative docs live.",
		"- `$namba-create`: use the preview-first creation flow when you need repo-local skills or project-scoped custom agents directly instead of another SPEC package.",
		"- `namba project`: refresh current repository docs and codemaps before choosing work.",
		"- `namba plan`: create the next feature SPEC package.",
		"- `namba harness`: create the next harness-oriented SPEC package for reusable agent, skill, workflow, or orchestration work.",
		"- `namba fix --command plan`: create a reviewable bugfix SPEC package.",
		"- `namba fix`: start direct repair in the current workspace.",
	} {
		if !strings.Contains(section, want) {
			t.Fatalf("planning-commands section missing %q: %q", want, section)
		}
	}
}

func TestRenderManagedProjectWorkflowGuidePRAndMergeFlowSectionPreservesEnglishAnchors(t *testing.T) {
	section := strings.Join(renderManagedProjectWorkflowGuidePRAndMergeFlowSection(), "\n")
	for _, want := range []string{
		"## PR and merge flow",
		"- `namba sync` stays local and refreshes generated artifacts only.",
		"- `namba pr` handles validation, commit, push, and PR handoff.",
		"bounded GitHub Actions failure snippets",
		"external checks by status plus details URL only",
		"- `namba land` merges a clean PR and updates local `main`.",
	} {
		if !strings.Contains(section, want) {
			t.Fatalf("pr-and-merge-flow section missing %q: %q", want, section)
		}
	}
}

func TestRenderManagedProjectWorkflowGuidePreludePreservesLocalizedAnchors(t *testing.T) {
	cases := []struct {
		lang          string
		heading       string
		introSentence string
	}{
		{
			lang:          "en",
			heading:       "# Workflow Guide",
			introSentence: "`repo` is managed with NambaAI conventions.",
		},
		{
			lang:          "ko",
			heading:       "# 워크플로 가이드",
			introSentence: "`repo`는 NambaAI 규칙으로 관리됩니다.",
		},
		{
			lang:          "ja",
			heading:       "# ワークフローガイド",
			introSentence: "`repo` は NambaAI の規約で管理されています。",
		},
		{
			lang:          "zh",
			heading:       "# 工作流指南",
			introSentence: "`repo` 由 NambaAI 约定管理。",
		},
	}

	for _, tc := range cases {
		section := strings.Join(renderManagedProjectWorkflowGuidePrelude(tc.lang, "repo"), "\n")
		for _, want := range []string{
			renderGeneratedDocHeader(),
			tc.heading,
			"[Codex Upstream Reference](./codex-upstream-reference.md)",
			tc.introSentence,
		} {
			if !strings.Contains(section, want) {
				t.Fatalf("%s workflow-guide prelude missing %q: %q", tc.lang, want, section)
			}
		}
	}
}

func TestRenderManagedProjectWorkflowGuidePreludeFallsBackToEnglish(t *testing.T) {
	section := strings.Join(renderManagedProjectWorkflowGuidePrelude("fr", "repo"), "\n")
	for _, want := range []string{
		"# Workflow Guide",
		"`repo` is managed with NambaAI conventions.",
	} {
		if !strings.Contains(section, want) {
			t.Fatalf("workflow-guide prelude fallback missing %q: %q", want, section)
		}
	}
}

func TestRenderManagedProjectWorkflowGuideSectionAssemblyPreservesOrderAndLanguageSpecificBlocks(t *testing.T) {
	assertOrderedContains := func(t *testing.T, content string, wants []string) {
		t.Helper()

		lastIndex := -1
		for _, want := range wants {
			index := strings.Index(content, want)
			if index == -1 {
				t.Fatalf("guide missing %q: %q", want, content)
			}
			if index <= lastIndex {
				t.Fatalf("guide out of order for %q after index %d: %q", want, lastIndex, content)
			}
			lastIndex = index
		}
	}

	englishGuide := renderManagedProjectWorkflowGuide("en", projectConfig{Name: "repo"}, initProfile{})
	assertOrderedContains(t, englishGuide, []string{
		"## Key locations",
		"## Work order",
		"## Planning commands",
		"## `namba plan` and `namba fix`",
		"## Run modes",
		"## Role routing",
		"## Review readiness",
		"## PR and merge flow",
		"## Collaboration rules",
	})

	koreanGuide := renderManagedProjectWorkflowGuide("ko", projectConfig{Name: "repo"}, initProfile{})
	assertOrderedContains(t, koreanGuide, []string{
		"## 핵심 위치",
		"## 작업 순서",
		"## `namba plan`과 `namba fix`",
		"## Run 모드",
		"## 역할 라우팅",
		"## 리뷰 준비도",
		"## 협업 규칙",
	})

	for _, unwanted := range []string{"## Planning commands", "## PR and merge flow"} {
		if strings.Contains(koreanGuide, unwanted) {
			t.Fatalf("korean guide unexpectedly contains %q: %q", unwanted, koreanGuide)
		}
	}
}

func TestRenderNambaCLIWorkflowGuideTailSectionHelpersPreserveLocalizedAnchors(t *testing.T) {
	cases := []struct {
		lang                 string
		prMergeHeading       string
		assetsHeading        string
		collaborationHeading string
		releaseHeading       string
	}{
		{
			lang:                 "en",
			prMergeHeading:       "## PR and merge flow",
			assetsHeading:        "## Key generated assets",
			collaborationHeading: "## Collaboration defaults",
			releaseHeading:       "## 🚢 Release Flow",
		},
		{
			lang:                 "ko",
			prMergeHeading:       "## PR 및 머지 흐름",
			assetsHeading:        "## 주요 생성 산출물",
			collaborationHeading: "## 협업 기본값",
			releaseHeading:       "## 🚢 릴리스 흐름",
		},
		{
			lang:                 "ja",
			prMergeHeading:       "## PR とマージの流れ",
			assetsHeading:        "## 主な生成物",
			collaborationHeading: "## 協業の既定値",
			releaseHeading:       "## 🚢 リリースフロー",
		},
		{
			lang:                 "zh",
			prMergeHeading:       "## PR 与合并流程",
			assetsHeading:        "## 主要生成产物",
			collaborationHeading: "## 协作默认值",
			releaseHeading:       "## 🚢 发布流程",
		},
	}

	for _, tc := range cases {
		prMerge := strings.Join(renderNambaCLIWorkflowGuidePRMergeSection(tc.lang), "\n")
		for _, want := range []string{tc.prMergeHeading, "`namba sync`", "`namba pr`", "`namba land`"} {
			if !strings.Contains(prMerge, want) {
				t.Fatalf("%s pr-merge section missing %q: %q", tc.lang, want, prMerge)
			}
		}

		assets := strings.Join(renderNambaCLIWorkflowGuideAssetsSection(tc.lang), "\n")
		for _, want := range []string{tc.assetsHeading, "`.namba/`", "`.agents/skills/`", "`.codex/agents/*.toml`", "`.namba/project/*`"} {
			if !strings.Contains(assets, want) {
				t.Fatalf("%s assets section missing %q: %q", tc.lang, want, assets)
			}
		}

		collaboration := strings.Join(renderNambaCLIWorkflowGuideCollaborationDefaultsSection(tc.lang), "\n")
		for _, want := range []string{tc.collaborationHeading, "`namba pr`", "`namba land`", "`@codex review`"} {
			if !strings.Contains(collaboration, want) {
				t.Fatalf("%s collaboration section missing %q: %q", tc.lang, want, collaboration)
			}
		}

		release := strings.Join(renderNambaCLIWorkflowGuideReleaseFlowSection(tc.lang), "\n")
		for _, want := range []string{tc.releaseHeading, "`$namba-release`", "`namba release`", "`--push`", ".namba/releases/<version>.md", "`checksums.txt`"} {
			if !strings.Contains(release, want) {
				t.Fatalf("%s release section missing %q: %q", tc.lang, want, release)
			}
		}
	}
}

func TestRenderNambaCLIWorkflowGuideTailSectionHelpersFallbackToEnglish(t *testing.T) {
	prMerge := strings.Join(renderNambaCLIWorkflowGuidePRMergeSection("fr"), "\n")
	if !strings.Contains(prMerge, "## PR and merge flow") {
		t.Fatalf("pr-merge section fallback missing English heading: %q", prMerge)
	}

	assets := strings.Join(renderNambaCLIWorkflowGuideAssetsSection("fr"), "\n")
	if !strings.Contains(assets, "## Key generated assets") {
		t.Fatalf("assets section fallback missing English heading: %q", assets)
	}

	collaboration := strings.Join(renderNambaCLIWorkflowGuideCollaborationDefaultsSection("fr"), "\n")
	if !strings.Contains(collaboration, "## Collaboration defaults") {
		t.Fatalf("collaboration section fallback missing English heading: %q", collaboration)
	}

	release := strings.Join(renderNambaCLIWorkflowGuideReleaseFlowSection("fr"), "\n")
	if !strings.Contains(release, "## 🚢 Release Flow") {
		t.Fatalf("release section fallback missing English heading: %q", release)
	}
}

func TestRenderNambaCLIRootQuickStartSectionPreservesLocalizedAnchors(t *testing.T) {
	cases := []struct {
		lang              string
		quickStartHeading string
		installHeading    string
		windowsLabel      string
		unixLabel         string
		bootstrapHeading  string
		workHeading       string
		harnessNote       string
	}{
		{
			lang:              "en",
			quickStartHeading: "## 🚀 Quick Start",
			installHeading:    "### 1. Install NambaAI",
			windowsLabel:      "Windows:",
			unixLabel:         "macOS / Linux:",
			bootstrapHeading:  "### 2. Bootstrap a new repository",
			workHeading:       "### 3. Start working from Codex",
			harnessNote:       "swap `namba plan` for `namba harness \"description\"`",
		},
		{
			lang:              "ko",
			quickStartHeading: "## 🚀 빠른 시작",
			installHeading:    "### 1. NambaAI 설치",
			windowsLabel:      "Windows:",
			unixLabel:         "macOS / Linux:",
			bootstrapHeading:  "### 2. 새 저장소 초기화",
			workHeading:       "### 3. Codex에서 작업 시작",
			harnessNote:       "`namba plan` 대신 `namba harness \"description\"`",
		},
		{
			lang:              "ja",
			quickStartHeading: "## 🚀 クイックスタート",
			installHeading:    "### 1. NambaAI をインストール",
			windowsLabel:      "Windows:",
			unixLabel:         "macOS / Linux:",
			bootstrapHeading:  "### 2. 新しいリポジトリを初期化",
			workHeading:       "### 3. Codex から作業を開始",
			harnessNote:       "`namba plan` の代わりに `namba harness \"description\"`",
		},
		{
			lang:              "zh",
			quickStartHeading: "## 🚀 快速开始",
			installHeading:    "### 1. 安装 NambaAI",
			windowsLabel:      "Windows:",
			unixLabel:         "macOS / Linux:",
			bootstrapHeading:  "### 2. 初始化新仓库",
			workHeading:       "### 3. 从 Codex 开始工作",
			harnessNote:       "`namba plan` 换成 `namba harness \"description\"`",
		},
	}

	for _, tc := range cases {
		section := strings.Join(renderNambaCLIRootQuickStartSection(tc.lang), "\n")
		for _, want := range []string{
			tc.quickStartHeading,
			tc.installHeading,
			tc.windowsLabel,
			tc.unixLabel,
			tc.bootstrapHeading,
			tc.workHeading,
			nambaInstallPowerShell,
			nambaInstallShell,
			"namba init .",
			"namba run SPEC-001",
			"namba land",
			tc.harnessNote,
		} {
			if !strings.Contains(section, want) {
				t.Fatalf("%s quick-start section missing %q: %q", tc.lang, want, section)
			}
		}
	}
}

func TestRenderNambaCLIRootQuickStartSectionFallbackToEnglish(t *testing.T) {
	section := strings.Join(renderNambaCLIRootQuickStartSection("fr"), "\n")
	for _, want := range []string{"## 🚀 Quick Start", "### 1. Install NambaAI", "### 2. Bootstrap a new repository", "### 3. Start working from Codex"} {
		if !strings.Contains(section, want) {
			t.Fatalf("quick-start section fallback missing %q: %q", want, section)
		}
	}
}

func TestRenderNambaCLIWorkflowGuideCommandDifferencesSectionPreservesLocalizedAnchors(t *testing.T) {
	cases := []struct {
		lang    string
		heading string
	}{
		{lang: "en", heading: "## `update`, `regen`, `sync`, `pr`, and `land` are different commands"},
		{lang: "ko", heading: "## `update`, `regen`, `sync`, `pr`, `land`는 서로 다른 명령입니다"},
		{lang: "ja", heading: "## `update`, `regen`, `sync`, `pr`, `land` はそれぞれ別のコマンドです"},
		{lang: "zh", heading: "## `update`、`regen`、`sync`、`pr`、`land` 是不同的命令"},
	}

	for _, tc := range cases {
		section := strings.Join(renderNambaCLIWorkflowGuideCommandDifferencesSection(tc.lang), "\n")
		for _, want := range []string{tc.heading, "`namba update`", "`codex update`", "`namba regen`", "`namba sync`", "`namba pr`", "`namba land`"} {
			if !strings.Contains(section, want) {
				t.Fatalf("%s command-differences section missing %q: %q", tc.lang, want, section)
			}
		}
	}
}

func TestRenderNambaCLIWorkflowGuideCommandDifferencesSectionFallbackToEnglish(t *testing.T) {
	section := strings.Join(renderNambaCLIWorkflowGuideCommandDifferencesSection("fr"), "\n")
	if !strings.Contains(section, "## `update`, `regen`, `sync`, `pr`, and `land` are different commands") {
		t.Fatalf("command-differences section fallback missing English heading: %q", section)
	}
}

func TestRenderNambaCLIRootCommandSurfaceSectionHelpersPreserveLocalizedAnchors(t *testing.T) {
	cases := []struct {
		lang                 string
		commandSkillsHeading string
		skillMappingHeading  string
		customAgentsHeading  string
	}{
		{
			lang:                 "en",
			commandSkillsHeading: "## 🧩 Command Skills In Codex",
			skillMappingHeading:  "## 🗺️ Skill To Command Mapping",
			customAgentsHeading:  "## 👥 Custom Agents In Codex",
		},
		{
			lang:                 "ko",
			commandSkillsHeading: "## 🧩 Codex에서 쓰는 Command Skill",
			skillMappingHeading:  "## 🗺️ Skill To Command Mapping",
			customAgentsHeading:  "## 👥 Codex용 Custom Agents",
		},
		{
			lang:                 "ja",
			commandSkillsHeading: "## 🧩 Codex で使う Command Skill",
			skillMappingHeading:  "## 🗺️ Skill To Command Mapping",
			customAgentsHeading:  "## 👥 Codex 用 Custom Agents",
		},
		{
			lang:                 "zh",
			commandSkillsHeading: "## 🧩 Codex 中的 Command Skill",
			skillMappingHeading:  "## 🗺️ Skill To Command Mapping",
			customAgentsHeading:  "## 👥 Codex 自定义 Agents",
		},
	}

	for _, tc := range cases {
		commandSkills := strings.Join(renderNambaCLIRootCommandSkillsSection(tc.lang), "\n")
		for _, want := range []string{tc.commandSkillsHeading, "`$namba-help`", "`$namba-run`", "`$namba-queue`", "`$namba-plan-review`", "`$namba-review-resolve`", "`$namba-release`", "`$namba-update`"} {
			if !strings.Contains(commandSkills, want) {
				t.Fatalf("%s command-skills section missing %q: %q", tc.lang, want, commandSkills)
			}
		}

		mapping := strings.Join(renderNambaCLIRootSkillMappingSection(tc.lang), "\n")
		for _, want := range []string{tc.skillMappingHeading, "`$namba-project` -> `namba project`", "`$namba-run` -> `namba run SPEC-XXX`", "`$namba-queue` -> `namba queue start <SPEC-RANGE|SPEC-LIST>`", "`$namba-review-resolve`", "`$namba-release`", "`$namba-update` -> `namba update [--version vX.Y.Z]`"} {
			if !strings.Contains(mapping, want) {
				t.Fatalf("%s skill-mapping section missing %q: %q", tc.lang, want, mapping)
			}
		}

		customAgents := strings.Join(renderNambaCLIRootCustomAgentsSection(tc.lang), "\n")
		for _, want := range []string{tc.customAgentsHeading, "`namba-product-manager`", "`namba-planner`", "`namba-security-engineer`", "`namba-implementer`"} {
			if !strings.Contains(customAgents, want) {
				t.Fatalf("%s custom-agents section missing %q: %q", tc.lang, want, customAgents)
			}
		}
	}
}

func TestRenderNambaCLIRootCommandSurfaceSectionHelpersFallbackToEnglish(t *testing.T) {
	commandSkills := strings.Join(renderNambaCLIRootCommandSkillsSection("fr"), "\n")
	if !strings.Contains(commandSkills, "## 🧩 Command Skills In Codex") {
		t.Fatalf("command-skills section fallback missing English heading: %q", commandSkills)
	}

	mapping := strings.Join(renderNambaCLIRootSkillMappingSection("fr"), "\n")
	if !strings.Contains(mapping, "## 🗺️ Skill To Command Mapping") {
		t.Fatalf("skill-mapping section fallback missing English heading: %q", mapping)
	}

	customAgents := strings.Join(renderNambaCLIRootCustomAgentsSection("fr"), "\n")
	if !strings.Contains(customAgents, "## 👥 Custom Agents In Codex") {
		t.Fatalf("custom-agents section fallback missing English heading: %q", customAgents)
	}
}

func TestRenderNambaCLIRootTailSectionHelpersPreserveLocalizedAnchors(t *testing.T) {
	cases := []struct {
		lang                      string
		readMoreHeading           string
		technicalSnapshotHeading  string
		workflowGuideNeedle       string
		technicalSnapshotLangNeed string
	}{
		{
			lang:                      "en",
			readMoreHeading:           "## 📚 Need More Detail?",
			technicalSnapshotHeading:  "## 🧱 Technical Snapshot",
			workflowGuideNeedle:       "generated assets, and collaboration defaults",
			technicalSnapshotLangNeed: "solve different problems and should not be mixed",
		},
		{
			lang:                      "ko",
			readMoreHeading:           "## 📚 더 읽기",
			technicalSnapshotHeading:  "## 🧱 기술 스냅샷",
			workflowGuideNeedle:       "생성 산출물, 협업 기본값",
			technicalSnapshotLangNeed: "서로 다른 문제를 푸는 명령",
		},
		{
			lang:                      "ja",
			readMoreHeading:           "## 📚 さらに詳しく",
			technicalSnapshotHeading:  "## 🧱 技術スナップショット",
			workflowGuideNeedle:       "生成物、協業ルール",
			technicalSnapshotLangNeed: "別々の問題を解くコマンド",
		},
		{
			lang:                      "zh",
			readMoreHeading:           "## 📚 继续阅读",
			technicalSnapshotHeading:  "## 🧱 技术概览",
			workflowGuideNeedle:       "生成产物和协作默认值",
			technicalSnapshotLangNeed: "各自解决不同的问题",
		},
	}

	for _, tc := range cases {
		readMore := strings.Join(renderNambaCLIRootReadMoreSection(tc.lang), "\n")
		for _, want := range []string{tc.readMoreHeading, localizeGuideLabel(tc.lang, "getting-started"), localizeGuideLabel(tc.lang, "workflow-guide"), "SECURITY.md", tc.workflowGuideNeedle} {
			if !strings.Contains(readMore, want) {
				t.Fatalf("%s read-more section missing %q: %q", tc.lang, want, readMore)
			}
		}

		technicalSnapshot := strings.Join(renderNambaCLIRootTechnicalSnapshotSection(tc.lang), "\n")
		for _, want := range []string{tc.technicalSnapshotHeading, ".namba/", ".agents/skills/", ".codex/agents/*.toml", "Emoji density rule", tc.technicalSnapshotLangNeed, "namba update", "namba regen", "namba sync", "namba pr", "namba land"} {
			if !strings.Contains(technicalSnapshot, want) {
				t.Fatalf("%s technical-snapshot section missing %q: %q", tc.lang, want, technicalSnapshot)
			}
		}
	}
}

func TestRenderNambaCLIRootTailSectionHelpersFallbackToEnglish(t *testing.T) {
	readMore := strings.Join(renderNambaCLIRootReadMoreSection("fr"), "\n")
	for _, want := range []string{"## 📚 Need More Detail?", "installation, updates, uninstall, init, and first-run flow", "SECURITY.md"} {
		if !strings.Contains(readMore, want) {
			t.Fatalf("read-more section fallback missing %q: %q", want, readMore)
		}
	}

	technicalSnapshot := strings.Join(renderNambaCLIRootTechnicalSnapshotSection("fr"), "\n")
	for _, want := range []string{"## 🧱 Technical Snapshot", ".namba/", ".agents/skills/", ".codex/agents/*.toml", "Emoji density rule", "solve different problems and should not be mixed"} {
		if !strings.Contains(technicalSnapshot, want) {
			t.Fatalf("technical-snapshot section fallback missing %q: %q", want, technicalSnapshot)
		}
	}
}
