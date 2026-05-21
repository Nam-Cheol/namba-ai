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
