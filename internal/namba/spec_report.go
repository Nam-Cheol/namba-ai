package namba

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
)

func formatSpecCreationReport(scaffoldCtx specPackageScaffoldContext, start planningStartResolution, outputs map[string]string, autoReview bool, language string) string {
	switch normalizeReadmeLanguage(language) {
	case "ko":
		return formatSpecCreationReportKorean(scaffoldCtx, start, outputs, autoReview)
	case "ja":
		return formatSpecCreationReportJapanese(scaffoldCtx, start, outputs, autoReview)
	case "zh":
		return formatSpecCreationReportChinese(scaffoldCtx, start, outputs, autoReview)
	default:
		return formatSpecCreationReportEnglish(scaffoldCtx, start, outputs, autoReview)
	}
}

func formatSpecCreationReportKorean(scaffoldCtx specPackageScaffoldContext, start planningStartResolution, outputs map[string]string, autoReview bool) string {
	kind := scaffoldCtx.Kind
	specID := scaffoldCtx.SpecID
	lines := []string{
		"",
		"SPEC 결정 보고서:",
		"쉬운 요약:",
		fmt.Sprintf("- 진행 이유: %s", specReportWhyKorean(kind, scaffoldCtx.Description)),
		fmt.Sprintf("- 앞으로 할 일: %s", specReportWhatKorean(kind)),
		fmt.Sprintf("- 다음에 해야 할 작업: %s", specReportNextWorkKorean(kind, specID, autoReview)),
		fmt.Sprintf("- 진행 판단: %s", specReportProceedSignalKorean(kind)),
		"",
		"검토할 애매한 부분:",
	}
	lines = append(lines, specReportOpenPointsKorean(kind, scaffoldCtx.Description)...)
	lines = append(lines,
		"",
		"보안 및 안전 메모:",
	)
	lines = append(lines, specReportSecurityNotesKorean(scaffoldCtx.Description)...)
	lines = append(lines,
		"",
		"개발자 상세:",
		fmt.Sprintf("- SPEC 패키지: %s", filepath.ToSlash(filepath.Join(specsDir, specID))),
		fmt.Sprintf("- 명령 surface: %s", specCreationInvocation(kind)),
		fmt.Sprintf("- 브랜치/작업공간: %s | %s", firstNonBlank(start.Branch, "n/a"), firstNonBlank(start.WorkspaceAction, "n/a")),
	)
	if next := specReportNextCommand(kind, specID, autoReview); next != "" {
		lines = append(lines, fmt.Sprintf("- 다음 명령: %s", next))
	}
	lines = append(lines, "- 생성된 파일:")
	for _, path := range specReportGeneratedPaths(outputs) {
		lines = append(lines, fmt.Sprintf("  - %s", path))
	}
	return strings.Join(lines, "\n") + "\n"
}

func formatSpecCreationReportEnglish(scaffoldCtx specPackageScaffoldContext, start planningStartResolution, outputs map[string]string, autoReview bool) string {
	kind := scaffoldCtx.Kind
	specID := scaffoldCtx.SpecID
	lines := []string{
		"",
		"SPEC decision report:",
		"Plain-language summary:",
		fmt.Sprintf("- Why this SPEC exists: %s", specReportWhy(kind, scaffoldCtx.Description)),
		fmt.Sprintf("- What this SPEC will do: %s", specReportWhat(kind)),
		fmt.Sprintf("- Next work to do: %s", specReportNextWork(kind, specID, autoReview)),
		fmt.Sprintf("- Proceed signal: %s", specReportProceedSignal(kind)),
		"",
		"Open points to review:",
	}
	lines = append(lines, specReportOpenPoints(kind, scaffoldCtx.Description)...)
	lines = append(lines,
		"",
		"Security and safety notes:",
	)
	lines = append(lines, specReportSecurityNotes(scaffoldCtx.Description)...)
	lines = append(lines,
		"",
		"Developer detail:",
		fmt.Sprintf("- SPEC package: %s", filepath.ToSlash(filepath.Join(specsDir, specID))),
		fmt.Sprintf("- Command surface: %s", specCreationInvocation(kind)),
		fmt.Sprintf("- Branch/workspace: %s | %s", firstNonBlank(start.Branch, "n/a"), firstNonBlank(start.WorkspaceAction, "n/a")),
	)
	if next := specReportNextCommand(kind, specID, autoReview); next != "" {
		lines = append(lines, fmt.Sprintf("- Next command: %s", next))
	}
	lines = append(lines, "- Generated files:")
	for _, path := range specReportGeneratedPaths(outputs) {
		lines = append(lines, fmt.Sprintf("  - %s", path))
	}
	return strings.Join(lines, "\n") + "\n"
}

func formatSpecCreationReportJapanese(scaffoldCtx specPackageScaffoldContext, start planningStartResolution, outputs map[string]string, autoReview bool) string {
	kind := scaffoldCtx.Kind
	specID := scaffoldCtx.SpecID
	lines := []string{
		"",
		"SPEC 判断レポート:",
		"要約:",
		fmt.Sprintf("- 進める理由: %s", specReportWhyJapanese(kind, scaffoldCtx.Description)),
		fmt.Sprintf("- これから行う作業: %s", specReportWhatJapanese(kind)),
		fmt.Sprintf("- 次にやる作業: %s", specReportNextWorkJapanese(kind, specID, autoReview)),
		fmt.Sprintf("- 判断: %s", specReportProceedSignalJapanese(kind)),
		"",
		"確認が必要な点:",
	}
	lines = append(lines, specReportOpenPointsJapanese(kind, scaffoldCtx.Description)...)
	lines = append(lines,
		"",
		"セキュリティと安全性のメモ:",
	)
	lines = append(lines, specReportSecurityNotesJapanese(scaffoldCtx.Description)...)
	lines = append(lines,
		"",
		"開発者向け詳細:",
		fmt.Sprintf("- SPEC パッケージ: %s", filepath.ToSlash(filepath.Join(specsDir, specID))),
		fmt.Sprintf("- コマンド surface: %s", specCreationInvocation(kind)),
		fmt.Sprintf("- ブランチ/ワークスペース: %s | %s", firstNonBlank(start.Branch, "n/a"), firstNonBlank(start.WorkspaceAction, "n/a")),
	)
	if next := specReportNextCommand(kind, specID, autoReview); next != "" {
		lines = append(lines, fmt.Sprintf("- 次のコマンド: %s", next))
	}
	lines = append(lines, "- 生成されたファイル:")
	for _, path := range specReportGeneratedPaths(outputs) {
		lines = append(lines, fmt.Sprintf("  - %s", path))
	}
	return strings.Join(lines, "\n") + "\n"
}

func formatSpecCreationReportChinese(scaffoldCtx specPackageScaffoldContext, start planningStartResolution, outputs map[string]string, autoReview bool) string {
	kind := scaffoldCtx.Kind
	specID := scaffoldCtx.SpecID
	lines := []string{
		"",
		"SPEC 决策报告:",
		"简要摘要:",
		fmt.Sprintf("- 推进原因: %s", specReportWhyChinese(kind, scaffoldCtx.Description)),
		fmt.Sprintf("- 接下来要做的事: %s", specReportWhatChinese(kind)),
		fmt.Sprintf("- 下一步工作: %s", specReportNextWorkChinese(kind, specID, autoReview)),
		fmt.Sprintf("- 判断: %s", specReportProceedSignalChinese(kind)),
		"",
		"需要确认的点:",
	}
	lines = append(lines, specReportOpenPointsChinese(kind, scaffoldCtx.Description)...)
	lines = append(lines,
		"",
		"安全与风险备注:",
	)
	lines = append(lines, specReportSecurityNotesChinese(scaffoldCtx.Description)...)
	lines = append(lines,
		"",
		"开发者详情:",
		fmt.Sprintf("- SPEC 包: %s", filepath.ToSlash(filepath.Join(specsDir, specID))),
		fmt.Sprintf("- 命令 surface: %s", specCreationInvocation(kind)),
		fmt.Sprintf("- 分支/工作区: %s | %s", firstNonBlank(start.Branch, "n/a"), firstNonBlank(start.WorkspaceAction, "n/a")),
	)
	if next := specReportNextCommand(kind, specID, autoReview); next != "" {
		lines = append(lines, fmt.Sprintf("- 下一条命令: %s", next))
	}
	lines = append(lines, "- 已生成文件:")
	for _, path := range specReportGeneratedPaths(outputs) {
		lines = append(lines, fmt.Sprintf("  - %s", path))
	}
	return strings.Join(lines, "\n") + "\n"
}

func specReportWhy(kind, description string) string {
	description = strings.TrimSpace(description)
	switch kind {
	case "fix":
		return fmt.Sprintf("the reported issue needs a reviewable repair plan before code changes: %s", description)
	case "harness":
		return fmt.Sprintf("the requested reusable Namba/Codex workflow work needs a harness-oriented plan: %s", description)
	default:
		return fmt.Sprintf("the requested feature or product change needs an implementation-ready plan: %s", description)
	}
}

func specReportWhyKorean(kind, description string) string {
	description = strings.TrimSpace(description)
	switch kind {
	case "fix":
		return fmt.Sprintf("코드를 바로 고치기 전에 문제를 재현하고 안전한 수정 계획을 검토해야 합니다: %s", description)
	case "harness":
		return fmt.Sprintf("재사용할 Namba/Codex 구성요소 작업이라서 일반 기능 계획보다 harness 전용 계획이 필요합니다: %s", description)
	default:
		return fmt.Sprintf("요청한 기능 또는 제품 변경을 구현 가능한 계획으로 바꿔야 합니다: %s", description)
	}
}

func specReportWhyJapanese(kind, description string) string {
	description = strings.TrimSpace(description)
	switch kind {
	case "fix":
		return fmt.Sprintf("コードを直接直す前に、問題の再現方法と安全な修正計画を確認する必要があります: %s", description)
	case "harness":
		return fmt.Sprintf("再利用する Namba/Codex 構成要素の作業なので、通常の機能計画ではなく harness 向けの計画が必要です: %s", description)
	default:
		return fmt.Sprintf("依頼された機能またはプロダクト変更を、実装可能な計画に落とし込む必要があります: %s", description)
	}
}

func specReportWhyChinese(kind, description string) string {
	description = strings.TrimSpace(description)
	switch kind {
	case "fix":
		return fmt.Sprintf("在直接修改代码前，需要先确认复现路径和安全修复计划: %s", description)
	case "harness":
		return fmt.Sprintf("这是可复用的 Namba/Codex 组件工作，需要使用 harness 取向的计划: %s", description)
	default:
		return fmt.Sprintf("需要把请求的功能或产品变更整理成可实施的计划: %s", description)
	}
}

func specReportWhat(kind string) string {
	switch kind {
	case "fix":
		return "reproduce or inspect the issue, choose the smallest safe fix, add regression coverage, and validate before sync."
	case "harness":
		return "define the reusable skill, agent, workflow, or orchestration change inside the normal SPEC review flow."
	default:
		return "turn the requested change into scoped implementation steps, acceptance checks, and review artifacts."
	}
}

func specReportWhatKorean(kind string) string {
	switch kind {
	case "fix":
		return "문제를 재현하거나 확인한 뒤, 가장 작은 안전한 수정과 회귀 테스트, 검증 절차를 진행합니다."
	case "harness":
		return "skill, agent, workflow, orchestration 같은 재사용 구성요소의 경계와 평가 방법을 SPEC 리뷰 흐름 안에서 정리합니다."
	default:
		return "요청을 구현 단계, 완료 기준, 리뷰 산출물로 나눠 다음 실행자가 바로 판단할 수 있게 만듭니다."
	}
}

func specReportWhatJapanese(kind string) string {
	switch kind {
	case "fix":
		return "問題を再現または確認し、最小で安全な修正、回帰テスト、検証手順を進めます。"
	case "harness":
		return "skill、agent、workflow、orchestration など再利用構成要素の境界と評価方法を SPEC レビューの流れで整理します。"
	default:
		return "依頼を実装手順、完了基準、レビュー成果物に分け、次の実行者がすぐ判断できる状態にします。"
	}
}

func specReportWhatChinese(kind string) string {
	switch kind {
	case "fix":
		return "先复现或确认问题，再推进最小且安全的修复、回归测试和验证步骤。"
	case "harness":
		return "在 SPEC 审查流程中整理 skill、agent、workflow、orchestration 等可复用组件的边界和评估方法。"
	default:
		return "把请求拆成实施步骤、完成标准和审查产物，让下一位执行者可以直接判断。"
	}
}

func specReportNextWork(kind, specID string, autoReview bool) string {
	readinessPath := specReviewReadinessPath(specID)
	switch kind {
	case "fix":
		return fmt.Sprintf("confirm the reproduction path and regression test in `$namba-plan-review %s`; if `%s` does not say `Cleared reviews: 3/3`, rerun `$namba-plan-review %s` or the missing review skills before `namba run %s`.", specID, readinessPath, specID, specID)
	case "harness":
		return fmt.Sprintf("confirm reusable boundaries and evaluation evidence in `$namba-plan-review %s`; if `%s` does not say `Cleared reviews: 3/3`, rerun `$namba-plan-review %s` or the missing review skills before `namba run %s`.", specID, readinessPath, specID, specID)
	case "plan":
		if autoReview {
			return fmt.Sprintf("run `$namba-plan-review %s` first; if `%s` does not say `Cleared reviews: 3/3`, rerun `$namba-plan-review %s` or the missing review skills before `namba run %s`.", specID, readinessPath, specID, specID)
		}
		return fmt.Sprintf("review product, engineering, and design concerns if needed; if `%s` does not say `Cleared reviews: 3/3`, run the missing individual review skills before `namba run %s`.", readinessPath, specID)
	default:
		return fmt.Sprintf("review the SPEC package; if `%s` does not say `Cleared reviews: 3/3`, rerun `$namba-plan-review %s` or the missing review skills before `namba run %s`.", readinessPath, specID, specID)
	}
}

func specReportNextWorkKorean(kind, specID string, autoReview bool) string {
	readinessPath := specReviewReadinessPath(specID)
	switch kind {
	case "fix":
		return fmt.Sprintf("먼저 `$namba-plan-review %s`로 재현 경로와 회귀 테스트를 확인하세요. `%s`가 `Cleared reviews: 3/3`이 아니면 `namba run %s` 전에 `$namba-plan-review %s` 또는 누락된 review skill을 다시 진행하세요.", specID, readinessPath, specID, specID)
	case "harness":
		return fmt.Sprintf("먼저 `$namba-plan-review %s`로 재사용 경계와 평가 증거를 확인하세요. `%s`가 `Cleared reviews: 3/3`이 아니면 `namba run %s` 전에 `$namba-plan-review %s` 또는 누락된 review skill을 다시 진행하세요.", specID, readinessPath, specID, specID)
	case "plan":
		if autoReview {
			return fmt.Sprintf("먼저 `$namba-plan-review %s`로 product/engineering/design 쟁점을 확인하세요. `%s`가 `Cleared reviews: 3/3`이 아니면 `namba run %s` 전에 `$namba-plan-review %s` 또는 누락된 review skill을 다시 진행하세요.", specID, readinessPath, specID, specID)
		}
		return fmt.Sprintf("리뷰를 생략한 상태입니다. 필요하면 product/engineering/design 쟁점을 확인하고, `%s`가 `Cleared reviews: 3/3`이 아니면 `namba run %s` 전에 누락된 개별 review skill을 진행하세요.", readinessPath, specID)
	default:
		return fmt.Sprintf("SPEC 패키지를 확인하세요. `%s`가 `Cleared reviews: 3/3`이 아니면 `namba run %s` 전에 `$namba-plan-review %s` 또는 누락된 review skill을 진행하세요.", readinessPath, specID, specID)
	}
}

func specReportNextWorkJapanese(kind, specID string, autoReview bool) string {
	readinessPath := specReviewReadinessPath(specID)
	switch kind {
	case "fix":
		return fmt.Sprintf("まず `$namba-plan-review %s` で再現方法と回帰テストを確認してください。`%s` が `Cleared reviews: 3/3` でなければ、`namba run %s` の前に `$namba-plan-review %s` または不足している review skill を再実行してください。", specID, readinessPath, specID, specID)
	case "harness":
		return fmt.Sprintf("まず `$namba-plan-review %s` で再利用境界と評価証拠を確認してください。`%s` が `Cleared reviews: 3/3` でなければ、`namba run %s` の前に `$namba-plan-review %s` または不足している review skill を再実行してください。", specID, readinessPath, specID, specID)
	case "plan":
		if autoReview {
			return fmt.Sprintf("まず `$namba-plan-review %s` で product/engineering/design の論点を確認してください。`%s` が `Cleared reviews: 3/3` でなければ、`namba run %s` の前に `$namba-plan-review %s` または不足している review skill を再実行してください。", specID, readinessPath, specID, specID)
		}
		return fmt.Sprintf("レビューを省略した状態です。必要なら product/engineering/design の論点を確認し、`%s` が `Cleared reviews: 3/3` でなければ `namba run %s` の前に不足している個別 review skill を実行してください。", readinessPath, specID)
	default:
		return fmt.Sprintf("SPEC パッケージを確認してください。`%s` が `Cleared reviews: 3/3` でなければ `namba run %s` の前に `$namba-plan-review %s` または不足している review skill を実行してください。", readinessPath, specID, specID)
	}
}

func specReportNextWorkChinese(kind, specID string, autoReview bool) string {
	readinessPath := specReviewReadinessPath(specID)
	switch kind {
	case "fix":
		return fmt.Sprintf("先用 `$namba-plan-review %s` 确认复现路径和回归测试；如果 `%s` 不是 `Cleared reviews: 3/3`，请在 `namba run %s` 前重新运行 `$namba-plan-review %s` 或缺失的 review skill。", specID, readinessPath, specID, specID)
	case "harness":
		return fmt.Sprintf("先用 `$namba-plan-review %s` 确认可复用边界和评估证据；如果 `%s` 不是 `Cleared reviews: 3/3`，请在 `namba run %s` 前重新运行 `$namba-plan-review %s` 或缺失的 review skill。", specID, readinessPath, specID, specID)
	case "plan":
		if autoReview {
			return fmt.Sprintf("先用 `$namba-plan-review %s` 确认 product/engineering/design 议题；如果 `%s` 不是 `Cleared reviews: 3/3`，请在 `namba run %s` 前重新运行 `$namba-plan-review %s` 或缺失的 review skill。", specID, readinessPath, specID, specID)
		}
		return fmt.Sprintf("当前跳过了自动审查。需要时先确认 product/engineering/design 议题；如果 `%s` 不是 `Cleared reviews: 3/3`，请在 `namba run %s` 前运行缺失的单项 review skill。", readinessPath, specID)
	default:
		return fmt.Sprintf("请先确认 SPEC 包；如果 `%s` 不是 `Cleared reviews: 3/3`，请在 `namba run %s` 前运行 `$namba-plan-review %s` 或缺失的 review skill。", readinessPath, specID, specID)
	}
}

func specReportProceedSignal(kind string) string {
	switch kind {
	case "fix":
		return "ready for bugfix plan review; implementation should wait until the issue shape and regression check are clear."
	case "harness":
		return "ready for harness plan review; implementation should wait until reusable boundaries and evaluation evidence are clear."
	default:
		return "ready for plan review; implementation should wait until product, engineering, and design concerns are visible."
	}
}

func specReportProceedSignalKorean(kind string) string {
	switch kind {
	case "fix":
		return "버그 수정 계획 검토를 시작할 수 있지만, 재현 경로와 회귀 테스트가 분명해진 뒤 구현하는 편이 안전합니다."
	case "harness":
		return "harness 계획 검토를 시작할 수 있지만, 재사용 경계와 평가 증거가 분명해진 뒤 구현하는 편이 안전합니다."
	default:
		return "계획 리뷰를 시작할 수 있지만, 제품/엔지니어링/디자인 우려를 확인한 뒤 구현하는 편이 안전합니다."
	}
}

func specReportProceedSignalJapanese(kind string) string {
	switch kind {
	case "fix":
		return "バグ修正計画レビューを始められますが、再現方法と回帰テストが明確になってから実装する方が安全です。"
	case "harness":
		return "harness 計画レビューを始められますが、再利用境界と評価証拠が明確になってから実装する方が安全です。"
	default:
		return "計画レビューを始められますが、product/engineering/design の懸念を確認してから実装する方が安全です。"
	}
}

func specReportProceedSignalChinese(kind string) string {
	switch kind {
	case "fix":
		return "可以开始 bugfix 计划审查，但最好在复现路径和回归测试明确后再实现。"
	case "harness":
		return "可以开始 harness 计划审查，但最好在可复用边界和评估证据明确后再实现。"
	default:
		return "可以开始计划审查，但最好先确认 product/engineering/design 风险后再实现。"
	}
}

func specReportOpenPoints(kind, description string) []string {
	points := []string{
		"- Product, engineering, and design review files are seeded but still pending.",
	}
	lower := strings.ToLower(description)
	groups := planClarificationEvidenceGroups()
	labels := []string{"goal", "scope", "constraints", "acceptance"}
	var missing []string
	for i, group := range groups {
		if !containsAnyFolded(lower, group) && i < len(labels) {
			missing = append(missing, labels[i])
		}
	}
	if len(missing) > 0 {
		points = append(points, fmt.Sprintf("- The original request did not explicitly name %s; confirm these during review if they matter.", strings.Join(missing, ", ")))
	}
	switch kind {
	case "fix":
		points = append(points, "- Confirm the reproduction path and the regression test before coding.")
	case "harness":
		points = append(points, "- Confirm whether the work changes Namba core behavior or only repo-local reusable artifacts.")
	default:
		points = append(points, "- Confirm any user-facing edge cases before `namba run`.")
	}
	return points
}

func specReportOpenPointsKorean(kind, description string) []string {
	points := []string{
		"- product, engineering, design 리뷰 파일은 만들어졌지만 아직 pending 상태입니다.",
	}
	lower := strings.ToLower(description)
	groups := planClarificationEvidenceGroups()
	labels := []string{"목표", "범위", "제약", "완료 기준"}
	var missing []string
	for i, group := range groups {
		if !containsAnyFolded(lower, group) && i < len(labels) {
			missing = append(missing, labels[i])
		}
	}
	if len(missing) > 0 {
		points = append(points, fmt.Sprintf("- 원 요청에 %s가 명시적으로 드러나지 않았습니다. 중요하면 리뷰에서 먼저 확정하세요.", strings.Join(missing, ", ")))
	}
	switch kind {
	case "fix":
		points = append(points, "- 코딩 전에 재현 경로와 회귀 테스트를 확인하세요.")
	case "harness":
		points = append(points, "- Namba core 동작 변경인지, repo-local 재사용 산출물 변경인지 확인하세요.")
	default:
		points = append(points, "- `namba run` 전에 사용자 흐름의 예외 상황을 확인하세요.")
	}
	return points
}

func specReportOpenPointsJapanese(kind, description string) []string {
	points := []string{
		"- product、engineering、design のレビュー ファイルは作成済みですが、まだ pending です。",
	}
	lower := strings.ToLower(description)
	groups := planClarificationEvidenceGroups()
	labels := []string{"目標", "範囲", "制約", "完了基準"}
	var missing []string
	for i, group := range groups {
		if !containsAnyFolded(lower, group) && i < len(labels) {
			missing = append(missing, labels[i])
		}
	}
	if len(missing) > 0 {
		points = append(points, fmt.Sprintf("- 元の依頼では %s が明示されていません。重要な場合はレビューで先に確定してください。", strings.Join(missing, ", ")))
	}
	switch kind {
	case "fix":
		points = append(points, "- コーディング前に再現方法と回帰テストを確認してください。")
	case "harness":
		points = append(points, "- Namba core の動作変更なのか、repo-local の再利用成果物変更なのかを確認してください。")
	default:
		points = append(points, "- `namba run` の前にユーザーフローの例外ケースを確認してください。")
	}
	return points
}

func specReportOpenPointsChinese(kind, description string) []string {
	points := []string{
		"- product、engineering、design 审查文件已创建，但仍处于 pending 状态。",
	}
	lower := strings.ToLower(description)
	groups := planClarificationEvidenceGroups()
	labels := []string{"目标", "范围", "约束", "完成标准"}
	var missing []string
	for i, group := range groups {
		if !containsAnyFolded(lower, group) && i < len(labels) {
			missing = append(missing, labels[i])
		}
	}
	if len(missing) > 0 {
		points = append(points, fmt.Sprintf("- 原始请求没有明确说明 %s；如果重要，请先在审查中确认。", strings.Join(missing, ", ")))
	}
	switch kind {
	case "fix":
		points = append(points, "- 编码前请确认复现路径和回归测试。")
	case "harness":
		points = append(points, "- 请确认这是 Namba core 行为变更，还是 repo-local 可复用产物变更。")
	default:
		points = append(points, "- 在 `namba run` 前请确认用户流程的例外情况。")
	}
	return points
}

func specReportSecurityNotes(description string) []string {
	notes := []string{
		"- No application code has changed yet; this command only created planning artifacts.",
	}
	lower := strings.ToLower(description)
	if containsAnyFolded(lower, []string{"auth", "login", "permission", "role", "token", "secret", "password", "admin", "payment", "billing", "pii", "personal data", "개인정보", "인증", "권한", "결제", "토큰", "비밀번호", "관리자"}) {
		notes = append(notes, "- Security-sensitive wording was detected; review authentication, authorization, secrets, and personal-data handling before implementation.")
	} else {
		notes = append(notes, "- During review, still check whether the implementation might touch auth, permissions, secrets, payments, personal data, or destructive data changes.")
	}
	return notes
}

func specReportSecurityNotesKorean(description string) []string {
	notes := []string{
		"- 아직 애플리케이션 코드는 바뀌지 않았고, 이번 명령은 계획 산출물만 만들었습니다.",
	}
	lower := strings.ToLower(description)
	if containsAnyFolded(lower, []string{"auth", "login", "permission", "role", "token", "secret", "password", "admin", "payment", "billing", "pii", "personal data", "개인정보", "인증", "권한", "결제", "토큰", "비밀번호", "관리자"}) {
		notes = append(notes, "- 보안 민감 단어가 감지되었습니다. 구현 전에 인증, 권한, 비밀값, 개인정보 처리를 꼭 검토하세요.")
	} else {
		notes = append(notes, "- 리뷰 중 인증, 권한, 비밀값, 결제, 개인정보, 삭제성 데이터 변경이 포함되는지 확인하세요.")
	}
	return notes
}

func specReportSecurityNotesJapanese(description string) []string {
	notes := []string{
		"- まだアプリケーションコードは変更されておらず、このコマンドは計画成果物だけを作成しました。",
	}
	lower := strings.ToLower(description)
	if containsAnyFolded(lower, []string{"auth", "login", "permission", "role", "token", "secret", "password", "admin", "payment", "billing", "pii", "personal data", "個人情報", "認証", "権限", "支払い", "トークン", "パスワード", "管理者"}) {
		notes = append(notes, "- セキュリティに関わる語が検出されました。実装前に認証、認可、秘密情報、個人情報の扱いを必ず確認してください。")
	} else {
		notes = append(notes, "- レビュー中に、認証、権限、秘密情報、支払い、個人情報、破壊的なデータ変更が含まれるか確認してください。")
	}
	return notes
}

func specReportSecurityNotesChinese(description string) []string {
	notes := []string{
		"- 目前还没有修改应用代码；这个命令只创建了计划产物。",
	}
	lower := strings.ToLower(description)
	if containsAnyFolded(lower, []string{"auth", "login", "permission", "role", "token", "secret", "password", "admin", "payment", "billing", "pii", "personal data", "个人信息", "认证", "权限", "支付", "令牌", "密码", "管理员"}) {
		notes = append(notes, "- 检测到安全敏感词。实现前请务必审查认证、授权、密钥和个人信息处理。")
	} else {
		notes = append(notes, "- 审查时仍需确认是否涉及认证、权限、密钥、支付、个人信息或破坏性数据变更。")
	}
	return notes
}

func specReportNextCommand(kind, specID string, autoReview bool) string {
	switch kind {
	case "plan":
		if autoReview {
			return fmt.Sprintf("$namba-plan-review %s", specID)
		}
		return fmt.Sprintf("individual product/engineering/design review skills for %s when review is wanted", specID)
	case "harness", "fix":
		return fmt.Sprintf("$namba-plan-review %s, or individual product/engineering/design review skills", specID)
	default:
		return ""
	}
}

func specReportGeneratedPaths(outputs map[string]string) []string {
	paths := make([]string, 0, len(outputs))
	for path := range outputs {
		paths = append(paths, filepath.ToSlash(path))
	}
	sort.Strings(paths)
	return paths
}
