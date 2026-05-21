package namba

import (
	"fmt"
	"html"
	"net/url"
	"strings"
)

func renderHeroBlock(heroPath, alt string) string {
	if strings.TrimSpace(heroPath) == "" {
		return ""
	}
	return strings.Join([]string{
		"<p align=\"center\">",
		fmt.Sprintf("  <img src=\"%s\" alt=\"%s\" width=\"100%%\" />", heroPath, alt),
		"</p>",
		"",
	}, "\n")
}

func appendHeroOrFallback(lines []string, heroPath, alt, fallback string) []string {
	if hero := renderHeroBlock(heroPath, alt); hero != "" {
		return append(lines, hero)
	}
	return append(lines, fallback, "")
}

func readmeRootCommandChooserHeading(lang string) string {
	switch normalizeReadmeLanguage(lang) {
	case "ko":
		return "## 🧭 어떤 명령을 써야 하나요?"
	case "ja":
		return "## 🧭 どのコマンドを使うべきですか?"
	case "zh":
		return "## 🧭 应该使用哪个命令？"
	default:
		return "## 🧭 Which Command Should I Use?"
	}
}

func renderRootCTAButtonRow(lang string) string {
	return strings.Join(renderCenteredBadgeLinkRow(rootCTAButtons(lang)), "\n")
}

type readmeBadgeLink struct {
	href    string
	alt     string
	label   string
	message string
	color   string
}

func rootCTAButtons(lang string) []readmeBadgeLink {
	switch normalizeReadmeLanguage(lang) {
	case "ko":
		return []readmeBadgeLink{
			{href: guidePath("getting-started", lang), alt: "시작하기", label: "시작", message: "가이드", color: "2ea44f"},
			{href: guidePath("workflow-guide", lang), alt: localizeGuideLabel(lang, "workflow-guide"), label: "워크플로", message: "가이드", color: "0969da"},
			{href: nambaRepositoryURL + "/releases/latest", alt: "릴리스", label: "릴리스", message: "최신", color: "8250df"},
			{href: "SECURITY.md", alt: "보안", label: "보안", message: "정책", color: "1a7f37"},
		}
	case "ja":
		return []readmeBadgeLink{
			{href: guidePath("getting-started", lang), alt: "始める", label: "開始", message: "ガイド", color: "2ea44f"},
			{href: guidePath("workflow-guide", lang), alt: localizeGuideLabel(lang, "workflow-guide"), label: "ワークフロー", message: "ガイド", color: "0969da"},
			{href: nambaRepositoryURL + "/releases/latest", alt: "Release", label: "リリース", message: "最新", color: "8250df"},
			{href: "SECURITY.md", alt: "Security", label: "セキュリティ", message: "ポリシー", color: "1a7f37"},
		}
	case "zh":
		return []readmeBadgeLink{
			{href: guidePath("getting-started", lang), alt: "开始使用", label: "开始", message: "指南", color: "2ea44f"},
			{href: guidePath("workflow-guide", lang), alt: localizeGuideLabel(lang, "workflow-guide"), label: "工作流", message: "指南", color: "0969da"},
			{href: nambaRepositoryURL + "/releases/latest", alt: "Release", label: "版本", message: "最新", color: "8250df"},
			{href: "SECURITY.md", alt: "Security", label: "安全", message: "策略", color: "1a7f37"},
		}
	default:
		return []readmeBadgeLink{
			{href: guidePath("getting-started", lang), alt: "Start here", label: "Start", message: "Here", color: "2ea44f"},
			{href: guidePath("workflow-guide", lang), alt: localizeGuideLabel(lang, "workflow-guide"), label: "Workflow", message: "Guide", color: "0969da"},
			{href: nambaRepositoryURL + "/releases/latest", alt: "Latest release", label: "Latest", message: "Release", color: "8250df"},
			{href: "SECURITY.md", alt: "Security", label: "Security", message: "Policy", color: "1a7f37"},
		}
	}
}

func renderCenteredBadgeLinkRow(links []readmeBadgeLink) []string {
	lines := []string{"<p align=\"center\">"}
	for _, link := range links {
		lines = append(lines, fmt.Sprintf("  <a href=\"%s\"><img src=\"%s\" alt=\"%s\" /></a>", html.EscapeString(link.href), readmeShieldBadgeURL(link.label, link.message, link.color), html.EscapeString(link.alt)))
	}
	lines = append(lines, "</p>")
	return lines
}

func readmeShieldBadgeURL(label, message, color string) string {
	values := url.Values{}
	values.Set("color", color)
	values.Set("label", label)
	values.Set("labelColor", "24292f")
	values.Set("message", message)
	return "https://img.shields.io/static/v1?" + strings.ReplaceAll(values.Encode(), "&", "&amp;")
}

func renderMermaidSection(title, intro string, diagramLines []string) []string {
	lines := []string{title, "", intro, "", "```mermaid"}
	lines = append(lines, diagramLines...)
	lines = append(lines, "```", "")
	return lines
}

func localizedMermaidTitle(lang, diagram string) string {
	switch diagram {
	case "root-command-flow":
		switch normalizeReadmeLanguage(lang) {
		case "ko":
			return "### 명령 흐름"
		case "ja":
			return "### コマンドフロー"
		case "zh":
			return "### 命令流程"
		default:
			return "### Command Flow"
		}
	case "getting-started-first-run":
		switch normalizeReadmeLanguage(lang) {
		case "ko":
			return "### 첫 실행 경로"
		case "ja":
			return "### 初回実行パス"
		case "zh":
			return "### 首次运行路径"
		default:
			return "### First-Run Path"
		}
	case "workflow-lifecycle":
		switch normalizeReadmeLanguage(lang) {
		case "ko":
			return "### SPEC 생명주기"
		case "ja":
			return "### SPEC ライフサイクル"
		case "zh":
			return "### SPEC 生命周期"
		default:
			return "### SPEC Lifecycle"
		}
	default:
		return "### Workflow Diagram"
	}
}

func localizedMermaidIntro(lang, diagram string) string {
	switch diagram {
	case "root-command-flow":
		switch normalizeReadmeLanguage(lang) {
		case "ko":
			return "가장 자주 쓰는 NambaAI 명령 흐름입니다. 여러 기존 SPEC을 순서대로 처리할 때는 `namba queue`에서 시작합니다."
		case "ja":
			return "よく使う NambaAI command の流れです。既存 SPEC を複数処理するときは `namba queue` から始めます。"
		case "zh":
			return "这是最常用的 NambaAI 命令流程。顺序处理多个已有 SPEC 时，从 `namba queue` 开始。"
		default:
			return "This is the main NambaAI command path. Start from `namba queue` when you need to process multiple existing SPECs."
		}
	case "getting-started-first-run":
		switch normalizeReadmeLanguage(lang) {
		case "ko":
			return "설치 뒤 첫 PR까지 이어지는 최소 경로입니다."
		case "ja":
			return "インストール後、最初の PR まで進む最短経路です。"
		case "zh":
			return "这是安装后走到第一个 PR 的最短路径。"
		default:
			return "This is the shortest path from install to the first PR."
		}
	case "workflow-lifecycle":
		switch normalizeReadmeLanguage(lang) {
		case "ko":
			return "SPEC은 아이디어를 수용 기준까지 정리한 뒤 구현, 검증, 문서 동기화, PR 인계, merge/land로 이동합니다. 애매하거나 검증이 실패하면 blocked 상태에서 repair/retry로 돌아갑니다."
		case "ja":
			return "SPEC は idea を acceptance まで整理してから implementation、validation、docs sync、PR handoff、merge/land へ進みます。曖昧な状態や validation 失敗は blocked になり、repair/retry へ戻ります。"
		case "zh":
			return "SPEC 会把 idea 整理到 acceptance，再进入 implementation、validation、docs sync、PR handoff 和 merge/land。状态不明确或 validation 失败时会进入 blocked，并回到 repair/retry。"
		default:
			return "A SPEC moves from idea to acceptance, implementation, validation, docs sync, PR handoff, and merge/land. Ambiguous state or failed validation stops as blocked and returns through repair/retry."
		}
	default:
		return "This diagram summarizes the generated documentation flow."
	}
}

func renderNambaCLIRootCommandFlowDiagramSection(lang string) []string {
	return renderMermaidSection(localizedMermaidTitle(lang, "root-command-flow"), localizedMermaidIntro(lang, "root-command-flow"), []string{
		"flowchart LR",
		"    project[\"namba project\"] --> choose{\"choose path\"}",
		"    choose --> plan[\"namba plan\"]",
		"    choose --> harness[\"namba harness\"]",
		"    choose --> fix[\"namba fix\"]",
		"    plan --> run[\"namba run\"]",
		"    harness --> run",
		"    fix --> run",
		"    queue[\"namba queue\"] --> run",
		"    run --> sync[\"namba sync\"]",
		"    sync --> pr[\"namba pr\"]",
		"    pr --> land[\"namba land\"]",
	})
}

func renderNambaCLIGettingStartedFirstRunDiagramSection(lang string) []string {
	return renderMermaidSection(localizedMermaidTitle(lang, "getting-started-first-run"), localizedMermaidIntro(lang, "getting-started-first-run"), []string{
		"flowchart LR",
		"    install[\"install\"] --> init[\"namba init .\"]",
		"    init --> project[\"namba project\"]",
		"    project --> plan[\"namba plan\"]",
		"    plan --> run[\"namba run\"]",
		"    run --> sync[\"namba sync\"]",
		"    sync --> pr[\"namba pr\"]",
	})
}

func renderNambaCLIWorkflowLifecycleDiagramSection(lang string) []string {
	return renderMermaidSection(localizedMermaidTitle(lang, "workflow-lifecycle"), localizedMermaidIntro(lang, "workflow-lifecycle"), []string{
		"flowchart LR",
		"    idea[\"idea\"] --> clarified[\"goal/scope/constraints/acceptance\"]",
		"    clarified --> spec[\"SPEC\"]",
		"    spec --> implementation[\"implementation\"]",
		"    implementation --> validation[\"validation\"]",
		"    validation --> docs[\"docs sync\"]",
		"    docs --> pr[\"PR handoff\"]",
		"    pr --> land[\"merge/land\"]",
		"    clarified --> blocked[\"blocked\"]",
		"    validation --> blocked",
		"    blocked --> repair[\"repair/retry\"]",
		"    repair --> clarified",
		"    repair --> implementation",
	})
}
