package namba

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
	"unicode"
)

type createTarget string

const (
	createTargetSkill createTarget = "skill"
	createTargetAgent createTarget = "agent"
	createTargetBoth  createTarget = "both"
)

type createRequest struct {
	Target               createTarget    `json:"target"`
	Name                 string          `json:"name"`
	Description          string          `json:"description"`
	Instructions         string          `json:"instructions"`
	SandboxMode          string          `json:"sandbox_mode,omitempty"`
	Model                string          `json:"model,omitempty"`
	ModelReasoningEffort string          `json:"model_reasoning_effort,omitempty"`
	HarnessRequest       *HarnessRequest `json:"harness_request,omitempty"`
	PreviewDigest        string          `json:"preview_digest,omitempty"`
	Confirmed            bool            `json:"confirmed,omitempty"`
	AllowOverwrite       bool            `json:"allow_overwrite,omitempty"`
}

type createPreview struct {
	Target           createTarget                `json:"target"`
	Name             string                      `json:"name"`
	Slug             string                      `json:"slug"`
	ExactOutputPaths []string                    `json:"exact_output_paths"`
	OverwriteImpact  createOverwriteImpact       `json:"overwrite_impact"`
	ValidationPlan   []string                    `json:"validation_plan"`
	SessionRefresh   createSessionRefreshPreview `json:"session_refresh"`
	HarnessRequest   *HarnessRequest             `json:"harness_request,omitempty"`
	PreviewDigest    string                      `json:"preview_digest"`
}

type createApplyResult struct {
	Preview        createPreview   `json:"preview"`
	WrittenPaths   []string        `json:"written_paths"`
	HarnessRequest *HarnessRequest `json:"harness_request,omitempty"`
}

type createOverwriteImpact struct {
	Required bool     `json:"required"`
	Paths    []string `json:"paths,omitempty"`
}

type createSessionRefreshPreview struct {
	Required bool   `json:"required"`
	Reason   string `json:"reason,omitempty"`
}

type createPlan struct {
	Preview createPreview
	Writes  []createWrite
}

type createWrite struct {
	Path    string
	Content string
}

type createPreviewDigestPayload struct {
	Version          string                      `json:"version"`
	Target           createTarget                `json:"target"`
	Name             string                      `json:"name"`
	Slug             string                      `json:"slug"`
	ExactOutputPaths []string                    `json:"exact_output_paths"`
	OverwriteImpact  createOverwriteImpact       `json:"overwrite_impact"`
	ValidationPlan   []string                    `json:"validation_plan"`
	SessionRefresh   createSessionRefreshPreview `json:"session_refresh"`
	HarnessRequest   *HarnessRequest             `json:"harness_request,omitempty"`
	Writes           []createWrite               `json:"writes"`
}

type createFileBackup struct {
	RelPath string
	AbsPath string
	Existed bool
	Content []byte
	Mode    os.FileMode
}

var createCollapseHyphenPattern = regexp.MustCompile(`-+`)
var createUnsafeInstructionPatterns = []struct {
	pattern *regexp.Regexp
	reason  string
}{
	{
		pattern: regexp.MustCompile(`(?i)\.claude(?:/|\\)`),
		reason:  "Claude-only paths are not allowed in durable instructions",
	},
	{
		pattern: regexp.MustCompile(`\b(?:TeamCreate|SendMessage|TaskCreate)\b`),
		reason:  "Claude-only runtime primitives are not allowed in durable instructions",
	},
	{
		pattern: regexp.MustCompile(`(?i)\bmodel\s*[:=]\s*["']?opus["']?`),
		reason:  "Claude-only model requirements are not allowed in durable instructions",
	},
	{
		pattern: regexp.MustCompile(`(?i)\.codex(?:/|\\)skills(?:/|\\)`),
		reason:  "Deprecated .codex/skills mirror paths are not allowed in durable instructions",
	},
}

func (a *App) previewCreate(root string, req createRequest) (createPreview, error) {
	plan, err := a.buildCreatePlan(root, req)
	if err != nil {
		return createPreview{}, err
	}
	return plan.Preview, nil
}

func (a *App) applyCreate(root string, req createRequest) (createApplyResult, error) {
	plan, err := a.buildCreatePlan(root, req)
	if err != nil {
		return createApplyResult{}, err
	}
	if !req.Confirmed {
		return createApplyResult{}, errors.New("create apply requires explicit confirmation before writing files")
	}
	if strings.TrimSpace(req.PreviewDigest) == "" {
		return createApplyResult{}, errors.New("create apply requires preview digest from a confirmed preview")
	}
	if req.PreviewDigest != plan.Preview.PreviewDigest {
		return createApplyResult{}, errors.New("create apply preview digest mismatch; request a fresh preview before writing")
	}
	if plan.Preview.OverwriteImpact.Required && !req.AllowOverwrite {
		return createApplyResult{}, fmt.Errorf("overwrite confirmation required for %s", strings.Join(plan.Preview.OverwriteImpact.Paths, ", "))
	}

	manifest, err := a.readManifest(root)
	if err != nil {
		return createApplyResult{}, err
	}
	now := a.now().Format(time.RFC3339)
	for _, write := range plan.Writes {
		manifest = upsertManifest(manifest, ManifestEntry{
			Path:      write.Path,
			Kind:      manifestKind(write.Path),
			Checksum:  checksum(write.Content),
			UpdatedAt: now,
		})
	}

	manifestBytes, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return createApplyResult{}, fmt.Errorf("marshal create manifest: %w", err)
	}

	report := outputWriteReport{
		ChangedPaths:            append([]string(nil), plan.Preview.ExactOutputPaths...),
		InstructionSurfacePaths: append([]string(nil), plan.Preview.ExactOutputPaths...),
	}
	noticeBytes, err := a.createSessionRefreshNoticeBytes(report)
	if err != nil {
		return createApplyResult{}, err
	}

	transactionWrites := append([]createWrite(nil), plan.Writes...)
	transactionWrites = append(transactionWrites, createWrite{
		Path:    manifestPath,
		Content: string(manifestBytes),
	})
	if len(noticeBytes) > 0 {
		transactionWrites = append(transactionWrites, createWrite{
			Path:    sessionRefreshNoticePath,
			Content: string(noticeBytes),
		})
	}
	if err := a.applyCreateWrites(root, transactionWrites); err != nil {
		return createApplyResult{}, err
	}

	return createApplyResult{
		Preview:        plan.Preview,
		WrittenPaths:   append([]string(nil), plan.Preview.ExactOutputPaths...),
		HarnessRequest: cloneHarnessRequest(plan.Preview.HarnessRequest),
	}, nil
}

func (a *App) buildCreatePlan(root string, req createRequest) (createPlan, error) {
	target, err := normalizeCreateTarget(req.Target)
	if err != nil {
		return createPlan{}, err
	}

	slug, err := normalizeCreateSlug(req.Name)
	if err != nil {
		return createPlan{}, err
	}

	description := normalizeCreateDescription(req.Description, target, slug)
	instructions := normalizeCreateInstructions(req.Instructions, target, slug)
	if err := validateCreateInstructionSafety(description, instructions); err != nil {
		return createPlan{}, err
	}
	harnessReq, err := validateCreateHarnessRequest(req.HarnessRequest, target)
	if err != nil {
		return createPlan{}, err
	}

	outputPaths := createOutputPaths(target, slug)
	if err := validateCreatePaths(root, outputPaths); err != nil {
		return createPlan{}, err
	}

	overwritePaths, err := detectCreateOverwritePaths(root, target, outputPaths)
	if err != nil {
		return createPlan{}, err
	}

	qualityCfg, err := a.loadQualityConfig(root)
	if err != nil {
		return createPlan{}, err
	}

	preview := createPreview{
		Target:           target,
		Name:             humanizeCreateSlug(slug),
		Slug:             slug,
		ExactOutputPaths: append([]string(nil), outputPaths...),
		ValidationPlan:   createValidationPlan(qualityCfg),
		HarnessRequest:   cloneHarnessRequest(harnessReq),
		OverwriteImpact: createOverwriteImpact{
			Required: len(overwritePaths) > 0,
			Paths:    append([]string(nil), overwritePaths...),
		},
		SessionRefresh: createSessionRefreshPreview{
			Required: true,
			Reason:   sessionRefreshNoticeReason(),
		},
	}

	writes := make([]createWrite, 0, len(outputPaths))
	for _, rel := range outputPaths {
		switch rel {
		case filepath.ToSlash(filepath.Join(repoSkillsDir, slug, "SKILL.md")):
			writes = append(writes, createWrite{
				Path:    rel,
				Content: renderUserSkill(slug, description, instructions),
			})
		case filepath.ToSlash(filepath.Join(repoCodexAgentsDir, slug+".toml")):
			writes = append(writes, createWrite{
				Path: rel,
				Content: renderUserAgentTOML(
					slug,
					description,
					instructions,
					req.SandboxMode,
					req.Model,
					req.ModelReasoningEffort,
				),
			})
		case filepath.ToSlash(filepath.Join(repoCodexAgentsDir, slug+".md")):
			writes = append(writes, createWrite{
				Path:    rel,
				Content: renderUserAgentMarkdown(slug, description, instructions),
			})
		default:
			return createPlan{}, fmt.Errorf("unsupported create output path %s", rel)
		}
	}

	previewDigest, err := computeCreatePreviewDigest(preview, writes)
	if err != nil {
		return createPlan{}, err
	}
	preview.PreviewDigest = previewDigest

	return createPlan{Preview: preview, Writes: writes}, nil
}

func normalizeCreateTarget(target createTarget) (createTarget, error) {
	switch createTarget(strings.TrimSpace(strings.ToLower(string(target)))) {
	case createTargetSkill:
		return createTargetSkill, nil
	case createTargetAgent:
		return createTargetAgent, nil
	case createTargetBoth:
		return createTargetBoth, nil
	default:
		return "", fmt.Errorf("unsupported create target %q", target)
	}
}

func normalizeCreateSlug(name string) (string, error) {
	raw := strings.TrimSpace(name)
	if raw == "" {
		return "", errors.New("invalid slug: name is required")
	}
	if strings.Contains(raw, "..") || strings.ContainsAny(raw, `/\`) {
		return "", fmt.Errorf("path traversal is not allowed in create names: %q", name)
	}

	var builder strings.Builder
	for _, r := range strings.ToLower(raw) {
		switch {
		case r >= 'a' && r <= 'z':
			builder.WriteRune(r)
		case r >= '0' && r <= '9':
			builder.WriteRune(r)
		case unicode.IsSpace(r), r == '-', r == '_':
			builder.WriteRune('-')
		default:
			builder.WriteRune('-')
		}
	}
	slug := createCollapseHyphenPattern.ReplaceAllString(builder.String(), "-")
	slug = strings.Trim(slug, "-")
	if slug == "" {
		return "", fmt.Errorf("invalid slug %q", name)
	}
	return slug, nil
}

func normalizeCreateDescription(description string, target createTarget, slug string) string {
	description = strings.TrimSpace(strings.ReplaceAll(description, "\n", " "))
	if description != "" {
		return description
	}
	switch target {
	case createTargetSkill:
		return fmt.Sprintf("User-authored skill generated for %s.", slug)
	case createTargetAgent:
		return fmt.Sprintf("User-authored custom agent generated for %s.", slug)
	default:
		return fmt.Sprintf("User-authored skill and agent generated for %s.", slug)
	}
}

func normalizeCreateInstructions(instructions string, target createTarget, slug string) string {
	instructions = strings.TrimSpace(instructions)
	if instructions != "" {
		return instructions
	}
	switch target {
	case createTargetSkill:
		return fmt.Sprintf("Use this skill when the user explicitly asks for `%s`.", slug)
	case createTargetAgent:
		return fmt.Sprintf("Use this custom agent when `%s` is the best fit for the assigned task.", slug)
	default:
		return fmt.Sprintf("Use these user-authored artifacts when `%s` is the explicit workflow target.", slug)
	}
}

func createOutputPaths(target createTarget, slug string) []string {
	switch target {
	case createTargetSkill:
		return []string{
			filepath.ToSlash(filepath.Join(repoSkillsDir, slug, "SKILL.md")),
		}
	case createTargetAgent:
		return []string{
			filepath.ToSlash(filepath.Join(repoCodexAgentsDir, slug+".toml")),
			filepath.ToSlash(filepath.Join(repoCodexAgentsDir, slug+".md")),
		}
	default:
		return []string{
			filepath.ToSlash(filepath.Join(repoSkillsDir, slug, "SKILL.md")),
			filepath.ToSlash(filepath.Join(repoCodexAgentsDir, slug+".toml")),
			filepath.ToSlash(filepath.Join(repoCodexAgentsDir, slug+".md")),
		}
	}
}

func validateCreatePaths(root string, paths []string) error {
	rootAbs, err := resolvedCreateRoot(root)
	if err != nil {
		return err
	}
	for _, rel := range paths {
		switch {
		case strings.HasPrefix(rel, repoSkillsDir+"/"):
			if isManagedRepoSkillPath(rel) {
				return fmt.Errorf("create path %s is reserved for Namba-managed skills", rel)
			}
		case strings.HasPrefix(rel, repoCodexAgentsDir+"/"):
			if isManagedRepoCodexAgentPath(rel) {
				return fmt.Errorf("create path %s is reserved for Namba-managed agents", rel)
			}
		default:
			return fmt.Errorf("create path %s is outside the allowlisted roots", rel)
		}
		abs := filepath.Clean(filepath.Join(rootAbs, filepath.FromSlash(rel)))
		if abs != rootAbs && !strings.HasPrefix(abs, rootAbs+string(os.PathSeparator)) {
			return fmt.Errorf("create path %s escapes the project root", rel)
		}
		if err := validateCreateResolvedPath(rootAbs, rel, abs); err != nil {
			return err
		}
	}
	return nil
}

func resolvedCreateRoot(root string) (string, error) {
	rootAbs, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}
	rootAbs, err = filepath.EvalSymlinks(rootAbs)
	if err != nil {
		return "", fmt.Errorf("resolve project root: %w", err)
	}
	return filepath.Clean(rootAbs), nil
}

func validateCreateResolvedPath(rootAbs, rel, absPath string) error {
	info, err := os.Lstat(absPath)
	switch {
	case err == nil:
		if info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("create output path %s is a symlink; symlinked create targets are not allowed", rel)
		}
		resolved, err := filepath.EvalSymlinks(absPath)
		if err != nil {
			return fmt.Errorf("resolve %s: %w", rel, err)
		}
		if !createPathWithinRoot(rootAbs, resolved) {
			return fmt.Errorf("create path %s resolves outside the project root", rel)
		}
		return nil
	case errors.Is(err, os.ErrNotExist):
		ancestor, err := nearestExistingCreateAncestor(rootAbs, absPath)
		if err != nil {
			return fmt.Errorf("inspect %s: %w", rel, err)
		}
		resolved, err := filepath.EvalSymlinks(ancestor)
		if err != nil {
			return fmt.Errorf("resolve %s: %w", rel, err)
		}
		if !createPathWithinRoot(rootAbs, resolved) {
			return fmt.Errorf("create path %s resolves outside the project root", rel)
		}
		return nil
	default:
		return fmt.Errorf("lstat %s: %w", rel, err)
	}
}

func nearestExistingCreateAncestor(rootAbs, absPath string) (string, error) {
	current := filepath.Dir(absPath)
	for {
		if _, err := os.Lstat(current); err == nil {
			return current, nil
		} else if !errors.Is(err, os.ErrNotExist) {
			return "", err
		}
		if current == rootAbs {
			return current, nil
		}
		next := filepath.Dir(current)
		if next == current {
			return "", fmt.Errorf("no existing ancestor for %s", absPath)
		}
		current = next
	}
}

func createPathWithinRoot(rootAbs, candidate string) bool {
	candidate = filepath.Clean(candidate)
	return candidate == rootAbs || strings.HasPrefix(candidate, rootAbs+string(os.PathSeparator))
}

func detectCreateOverwritePaths(root string, target createTarget, outputPaths []string) ([]string, error) {
	overwritePaths := make([]string, 0, len(outputPaths))
	if target == createTargetAgent || target == createTargetBoth {
		var agentPaths []string
		for _, rel := range outputPaths {
			if strings.HasPrefix(rel, repoCodexAgentsDir+"/") {
				agentPaths = append(agentPaths, rel)
			}
		}
		existsCount := 0
		for _, rel := range agentPaths {
			info, err := os.Stat(filepath.Join(root, filepath.FromSlash(rel)))
			if err != nil {
				if errors.Is(err, os.ErrNotExist) {
					continue
				}
				return nil, fmt.Errorf("inspect %s: %w", rel, err)
			}
			if info.IsDir() {
				return nil, fmt.Errorf("create output path %s is a directory, not a file", rel)
			}
			existsCount++
			overwritePaths = append(overwritePaths, rel)
		}
		if existsCount == 1 {
			return nil, fmt.Errorf("incomplete agent mirror state detected for %s", strings.TrimSuffix(filepath.Base(agentPaths[0]), filepath.Ext(agentPaths[0])))
		}
	}

	for _, rel := range outputPaths {
		if strings.HasPrefix(rel, repoCodexAgentsDir+"/") {
			continue
		}
		info, err := os.Stat(filepath.Join(root, filepath.FromSlash(rel)))
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				continue
			}
			return nil, fmt.Errorf("inspect %s: %w", rel, err)
		}
		if info.IsDir() {
			return nil, fmt.Errorf("create output path %s is a directory, not a file", rel)
		}
		overwritePaths = append(overwritePaths, rel)
	}

	sort.Strings(overwritePaths)
	return overwritePaths, nil
}

func createValidationPlan(cfg qualityConfig) []string {
	plan := make([]string, 0)
	for _, step := range validationPipelineSteps(cfg) {
		command := strings.TrimSpace(step.Command)
		if command == "" || command == "none" {
			continue
		}
		plan = append(plan, command)
	}
	return plan
}

func validateCreateInstructionSafety(description, instructions string) error {
	candidate := strings.TrimSpace(strings.Join([]string{description, instructions}, "\n"))
	for _, blocked := range createUnsafeInstructionPatterns {
		if blocked.pattern.MatchString(candidate) {
			return fmt.Errorf("unsafe create instructions: %s", blocked.reason)
		}
	}
	return nil
}

func renderUserSkill(slug, description, instructions string) string {
	return renderCommandSkill(slug, description, splitCreateInstructions(instructions))
}

func renderUserAgentTOML(slug, description, instructions, sandboxMode, model, reasoning string) string {
	profile := runtimeProfileForAgent(slug)
	model = firstNonBlank(strings.TrimSpace(model), strings.TrimSpace(profile.Model), "gpt-5.4-mini")
	reasoning = firstNonBlank(strings.TrimSpace(reasoning), strings.TrimSpace(profile.ModelReasoningEffort), "medium")
	sandboxMode = firstNonBlank(strings.TrimSpace(sandboxMode), "workspace-write")
	agentInstructions := strings.Join(splitCreateInstructions(withCreateAgentPreamble(slug, instructions)), "\n")
	lines := []string{
		"name = " + tomlString(slug),
		"description = " + tomlString(description),
		"sandbox_mode = " + tomlString(sandboxMode),
		"model = " + tomlString(model),
		"model_reasoning_effort = " + tomlString(reasoning),
		"developer_instructions = " + tomlString(agentInstructions),
	}
	return strings.Join(lines, "\n") + "\n"
}

func renderUserAgentMarkdown(slug, description, instructions string) string {
	lines := []string{
		fmt.Sprintf("# %s", humanizeCreateSlug(slug)),
		"",
		description,
		"",
	}
	lines = append(lines, splitCreateInstructions(withCreateAgentPreamble(slug, instructions))...)
	return strings.Join(lines, "\n") + "\n"
}

func withCreateAgentPreamble(slug, instructions string) string {
	instructions = strings.TrimSpace(instructions)
	if instructions == "" {
		return fmt.Sprintf("Use this custom agent when `%s` is the best fit for the assigned task.", slug)
	}
	return strings.Join([]string{
		fmt.Sprintf("Use this custom agent when `%s` is the best fit for the assigned task.", slug),
		"",
		instructions,
	}, "\n")
}

func splitCreateInstructions(instructions string) []string {
	lines := strings.Split(strings.TrimSpace(instructions), "\n")
	result := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimRight(line, " \t")
		if strings.TrimSpace(line) == "" && len(result) > 0 && result[len(result)-1] == "" {
			continue
		}
		result = append(result, line)
	}
	if len(result) == 0 {
		return []string{"Add durable instructions for this user-authored artifact."}
	}
	return result
}

func humanizeCreateSlug(slug string) string {
	parts := strings.FieldsFunc(slug, func(r rune) bool { return r == '-' || r == '_' })
	for i, part := range parts {
		if part == "" {
			continue
		}
		parts[i] = strings.ToUpper(part[:1]) + part[1:]
	}
	if len(parts) == 0 {
		return slug
	}
	return strings.Join(parts, " ")
}

func (a *App) createSessionRefreshNoticeBytes(report outputWriteReport) ([]byte, error) {
	if len(report.InstructionSurfacePaths) == 0 {
		return nil, nil
	}
	notice := sessionRefreshNotice{
		Required:    true,
		Reason:      sessionRefreshNoticeReason(),
		Paths:       report.InstructionSurfacePaths,
		GeneratedAt: a.now().Format(time.RFC3339),
	}
	data, err := json.MarshalIndent(notice, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal session refresh notice: %w", err)
	}
	return data, nil
}

func sessionRefreshNoticeReason() string {
	return "Generated instruction surfaces changed. Start a fresh Codex session before continuing a long team or repair run."
}

func computeCreatePreviewDigest(preview createPreview, writes []createWrite) (string, error) {
	payload := createPreviewDigestPayload{
		Version:          "create-preview-v2",
		Target:           preview.Target,
		Name:             preview.Name,
		Slug:             preview.Slug,
		ExactOutputPaths: append([]string(nil), preview.ExactOutputPaths...),
		OverwriteImpact: createOverwriteImpact{
			Required: preview.OverwriteImpact.Required,
			Paths:    append([]string(nil), preview.OverwriteImpact.Paths...),
		},
		ValidationPlan: append([]string(nil), preview.ValidationPlan...),
		SessionRefresh: createSessionRefreshPreview{
			Required: preview.SessionRefresh.Required,
			Reason:   preview.SessionRefresh.Reason,
		},
		HarnessRequest: cloneHarnessRequest(preview.HarnessRequest),
		Writes:         append([]createWrite(nil), writes...),
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("marshal create preview digest: %w", err)
	}
	return checksum(string(data)), nil
}

func cloneHarnessRequest(req *HarnessRequest) *HarnessRequest {
	if req == nil {
		return nil
	}
	copyReq := *req
	copyReq.ArtifactTargets = append([]harnessArtifactTarget(nil), req.ArtifactTargets...)
	copyReq.RequiredEvidence = append([]harnessEvidence(nil), req.RequiredEvidence...)
	copyReq.RequiredReviews = append([]harnessReview(nil), req.RequiredReviews...)
	return &copyReq
}

func (a *App) applyCreateWrites(root string, writes []createWrite) error {
	rootAbs, err := resolvedCreateRoot(root)
	if err != nil {
		return err
	}
	backups := make(map[string]createFileBackup, len(writes))
	for _, write := range writes {
		absPath := filepath.Join(root, filepath.FromSlash(write.Path))
		if _, ok := backups[absPath]; ok {
			continue
		}
		backup, err := snapshotCreateFile(rootAbs, write.Path, absPath)
		if err != nil {
			return err
		}
		backups[absPath] = backup
	}

	for _, write := range writes {
		absPath := filepath.Join(root, filepath.FromSlash(write.Path))
		if err := a.mkdirAll(filepath.Dir(absPath), 0o755); err != nil {
			rollbackErr := a.rollbackCreateWrites(backups)
			return joinCreateRollbackError(fmt.Errorf("create parent for %s: %w", write.Path, err), rollbackErr)
		}
		if err := a.writeFile(absPath, []byte(write.Content), 0o644); err != nil {
			rollbackErr := a.rollbackCreateWrites(backups)
			return joinCreateRollbackError(fmt.Errorf("write %s: %w", write.Path, err), rollbackErr)
		}
	}
	return nil
}

func snapshotCreateFile(rootAbs, relPath, absPath string) (createFileBackup, error) {
	if err := validateCreateResolvedPath(rootAbs, relPath, absPath); err != nil {
		return createFileBackup{}, err
	}
	info, err := os.Lstat(absPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return createFileBackup{RelPath: relPath, AbsPath: absPath}, nil
		}
		return createFileBackup{}, fmt.Errorf("stat %s: %w", relPath, err)
	}
	if info.IsDir() {
		return createFileBackup{}, fmt.Errorf("create output path %s is a directory, not a file", relPath)
	}
	content, err := os.ReadFile(absPath)
	if err != nil {
		return createFileBackup{}, fmt.Errorf("read %s: %w", relPath, err)
	}
	return createFileBackup{
		RelPath: relPath,
		AbsPath: absPath,
		Existed: true,
		Content: content,
		Mode:    info.Mode(),
	}, nil
}

func (a *App) rollbackCreateWrites(backups map[string]createFileBackup) error {
	keys := make([]string, 0, len(backups))
	for absPath := range backups {
		keys = append(keys, absPath)
	}
	sort.Sort(sort.Reverse(sort.StringSlice(keys)))

	var rollbackErr error
	for _, absPath := range keys {
		backup := backups[absPath]
		if !backup.Existed {
			if err := os.Remove(absPath); err != nil && !errors.Is(err, os.ErrNotExist) {
				rollbackErr = appendCreateError(rollbackErr, fmt.Errorf("rollback remove %s: %w", backup.RelPath, err))
			}
			continue
		}
		if err := a.mkdirAll(filepath.Dir(absPath), 0o755); err != nil {
			rollbackErr = appendCreateError(rollbackErr, fmt.Errorf("rollback mkdir %s: %w", backup.RelPath, err))
			continue
		}
		if err := a.writeFile(absPath, backup.Content, backup.Mode.Perm()); err != nil {
			rollbackErr = appendCreateError(rollbackErr, fmt.Errorf("rollback restore %s: %w", backup.RelPath, err))
		}
	}
	return rollbackErr
}

func joinCreateRollbackError(writeErr, rollbackErr error) error {
	if rollbackErr == nil {
		return writeErr
	}
	return fmt.Errorf("%w (rollback failed: %v)", writeErr, rollbackErr)
}

func appendCreateError(current error, next error) error {
	if next == nil {
		return current
	}
	if current == nil {
		return next
	}
	return fmt.Errorf("%v; %w", current, next)
}
