package namba

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type harnessRequestKind string

const (
	harnessRequestKindCore   harnessRequestKind = "core_harness_change"
	harnessRequestKindDomain harnessRequestKind = "domain_harness_change"
	harnessRequestKindDirect harnessRequestKind = "direct_artifact_creation"
)

type harnessDeliveryMode string

const (
	harnessDeliveryModeSpec   harnessDeliveryMode = "spec"
	harnessDeliveryModeDirect harnessDeliveryMode = "direct"
)

type harnessAdaptationMode string

const (
	harnessAdaptationModifyCore       harnessAdaptationMode = "modify_core"
	harnessAdaptationExtendDomain     harnessAdaptationMode = "extend_domain"
	harnessAdaptationComposeDomain    harnessAdaptationMode = "compose_domain"
	harnessAdaptationGenerateArtifact harnessAdaptationMode = "generate_artifact"
)

type harnessArtifactTarget string

const (
	harnessArtifactTargetSkill     harnessArtifactTarget = "skill"
	harnessArtifactTargetAgent     harnessArtifactTarget = "agent"
	harnessArtifactTargetWorkflow  harnessArtifactTarget = "workflow"
	harnessArtifactTargetValidator harnessArtifactTarget = "validator"
	harnessArtifactTargetEvalPack  harnessArtifactTarget = "eval-pack"
	harnessArtifactTargetDocs      harnessArtifactTarget = "docs"
)

type harnessEvidence string

const (
	harnessEvidenceContract   harnessEvidence = "contract"
	harnessEvidenceBaseline   harnessEvidence = "baseline"
	harnessEvidenceEvalPlan   harnessEvidence = "eval-plan"
	harnessEvidenceHarnessMap harnessEvidence = "harness-map"
)

type harnessReview string

const (
	harnessReviewProduct     harnessReview = "product"
	harnessReviewEngineering harnessReview = "engineering"
	harnessReviewDesign      harnessReview = "design"
)

type harnessRoute string

const (
	harnessRoutePlan       harnessRoute = "namba plan"
	harnessRouteHarness    harnessRoute = "namba harness"
	harnessRouteCreate     harnessRoute = "$namba-create"
	harnessRequestFileName              = "harness-request.json"
)

type harnessRequest struct {
	RequestKind      harnessRequestKind      `json:"request_kind"`
	DeliveryMode     harnessDeliveryMode     `json:"delivery_mode"`
	AdaptationMode   harnessAdaptationMode   `json:"adaptation_mode"`
	BaseContractRef  string                  `json:"base_contract_ref"`
	TouchesNambaCore bool                    `json:"touches_namba_core"`
	ArtifactTargets  []harnessArtifactTarget `json:"artifact_targets"`
	RequiredEvidence []harnessEvidence       `json:"required_evidence"`
	RequiredReviews  []harnessReview         `json:"required_reviews"`
}

type HarnessRequest = harnessRequest

type harnessValidationResult struct {
	Request harnessRequest
	Route   harnessRoute
	Issues  []string
}

type harnessEvidenceValidationReport struct {
	Request          harnessRequest
	Route            harnessRoute
	RequiredEvidence []string
	MissingEvidence  []string
	RequiredReviews  []string
	MissingReviews   []string
	Problems         []string
}

var harnessArtifactTargetOrder = map[harnessArtifactTarget]int{
	harnessArtifactTargetSkill:     0,
	harnessArtifactTargetAgent:     1,
	harnessArtifactTargetWorkflow:  2,
	harnessArtifactTargetValidator: 3,
	harnessArtifactTargetEvalPack:  4,
	harnessArtifactTargetDocs:      5,
}

var harnessEvidenceOrder = map[harnessEvidence]int{
	harnessEvidenceContract:   0,
	harnessEvidenceBaseline:   1,
	harnessEvidenceEvalPlan:   2,
	harnessEvidenceHarnessMap: 3,
}

var harnessReviewOrder = map[harnessReview]int{
	harnessReviewProduct:     0,
	harnessReviewEngineering: 1,
	harnessReviewDesign:      2,
}

func specHarnessRequestPath(specID string) string {
	return filepath.ToSlash(filepath.Join(specsDir, specID, harnessRequestFileName))
}

func harnessRouteForRequest(req harnessRequest) (harnessRoute, error) {
	switch req.RequestKind {
	case harnessRequestKindCore:
		return harnessRoutePlan, nil
	case harnessRequestKindDomain:
		return harnessRouteHarness, nil
	case harnessRequestKindDirect:
		return harnessRouteCreate, nil
	default:
		return "", fmt.Errorf("unknown harness request kind %q", req.RequestKind)
	}
}

func normalizeHarnessRequest(req harnessRequest) (harnessRequest, error) {
	req.BaseContractRef = strings.TrimSpace(req.BaseContractRef)
	req.ArtifactTargets = normalizeHarnessArtifactTargets(req.ArtifactTargets)
	req.RequiredEvidence = normalizeHarnessEvidence(req.RequiredEvidence)
	req.RequiredReviews = normalizeHarnessReviews(req.RequiredReviews)
	if err := validateHarnessRequest(req); err != nil {
		return harnessRequest{}, err
	}
	return req, nil
}

func validateHarnessRequest(req harnessRequest) error {
	if _, ok := harnessRequestKindSet()[req.RequestKind]; !ok {
		return fmt.Errorf("unknown harness request kind %q", req.RequestKind)
	}
	if _, ok := harnessDeliveryModeSet()[req.DeliveryMode]; !ok {
		return fmt.Errorf("unknown harness delivery mode %q", req.DeliveryMode)
	}
	if _, ok := harnessAdaptationModeSet()[req.AdaptationMode]; !ok {
		return fmt.Errorf("unknown harness adaptation mode %q", req.AdaptationMode)
	}
	if len(req.ArtifactTargets) == 0 {
		return fmt.Errorf("artifact_targets must not be empty")
	}
	if req.RequestKind == harnessRequestKindCore && !req.TouchesNambaCore {
		return fmt.Errorf("core_harness_change requires touches_namba_core=true")
	}
	if req.RequestKind == harnessRequestKindDomain && req.TouchesNambaCore {
		return fmt.Errorf("domain_harness_change must escalate to core_harness_change when touches_namba_core=true")
	}
	if req.RequestKind == harnessRequestKindDirect && req.TouchesNambaCore {
		return fmt.Errorf("direct_artifact_creation must route to `namba plan` before it touches Namba core")
	}
	if req.RequestKind == harnessRequestKindDirect && req.DeliveryMode != harnessDeliveryModeDirect {
		return fmt.Errorf("direct_artifact_creation requires delivery_mode=direct")
	}
	if req.RequestKind != harnessRequestKindDirect && req.DeliveryMode == harnessDeliveryModeDirect {
		return fmt.Errorf("direct delivery mode is reserved for direct_artifact_creation")
	}
	if req.RequestKind == harnessRequestKindDirect && req.AdaptationMode != harnessAdaptationGenerateArtifact {
		return fmt.Errorf("direct_artifact_creation requires adaptation_mode=generate_artifact")
	}
	if req.RequestKind != harnessRequestKindDirect && req.DeliveryMode == harnessDeliveryModeSpec {
		if !hasAllHarnessEvidence(req.RequiredEvidence, harnessEvidenceContract, harnessEvidenceBaseline, harnessEvidenceEvalPlan) {
			return fmt.Errorf("spec delivery requires contract, baseline, and eval-plan evidence")
		}
		if !hasAllHarnessReviews(req.RequiredReviews, harnessReviewProduct, harnessReviewEngineering, harnessReviewDesign) {
			return fmt.Errorf("spec delivery requires product, engineering, and design reviews")
		}
		if harnessRequestRequiresHarnessMap(req) && !hasHarnessEvidence(req.RequiredEvidence, harnessEvidenceHarnessMap) {
			return fmt.Errorf("compose/adaptation boundaries require harness-map evidence")
		}
	}
	if req.RequestKind == harnessRequestKindDirect {
		if len(req.RequiredEvidence) > 0 {
			return fmt.Errorf("direct_artifact_creation must not require the harness evidence pack")
		}
		if len(req.RequiredReviews) > 0 {
			return fmt.Errorf("direct_artifact_creation must not require harness review tracks")
		}
	}
	return nil
}

func harnessRequestRequiresHarnessMap(req harnessRequest) bool {
	if hasHarnessEvidence(req.RequiredEvidence, harnessEvidenceHarnessMap) {
		return true
	}
	return req.RequestKind == harnessRequestKindDomain &&
		(req.AdaptationMode == harnessAdaptationComposeDomain || req.BaseContractRef != "")
}

func validateCreateHarnessRequest(req *harnessRequest, target createTarget) (*harnessRequest, error) {
	if req == nil {
		return nil, nil
	}
	normalized, err := normalizeHarnessRequest(*req)
	if err != nil {
		return nil, err
	}
	route, err := harnessRouteForRequest(normalized)
	if err != nil {
		return nil, err
	}
	if route != harnessRouteCreate {
		return nil, fmt.Errorf("direct create request must route through `%s`, not `%s`", harnessRouteCreate, route)
	}

	expectedTargets := expectedCreateHarnessTargets(target)
	if !equalHarnessArtifactTargets(normalized.ArtifactTargets, expectedTargets) {
		return nil, fmt.Errorf("direct create target %q requires artifact_targets %v", target, harnessArtifactTargetsToStrings(expectedTargets))
	}
	return &normalized, nil
}

func validateDirectCreateHarnessRequest(req *HarnessRequest) error {
	if req == nil {
		return nil
	}
	normalized, err := normalizeHarnessRequest(*req)
	if err != nil {
		return err
	}
	route, err := harnessRouteForRequest(normalized)
	if err != nil {
		return err
	}
	if route != harnessRouteCreate {
		return fmt.Errorf("direct create request must route through `%s`, not `%s`", harnessRouteCreate, route)
	}
	return nil
}

func defaultHarnessRequest() harnessRequest {
	req, err := normalizeHarnessRequest(harnessRequest{
		RequestKind:      harnessRequestKindDomain,
		DeliveryMode:     harnessDeliveryModeSpec,
		AdaptationMode:   harnessAdaptationExtendDomain,
		BaseContractRef:  "",
		TouchesNambaCore: false,
		ArtifactTargets: []harnessArtifactTarget{
			harnessArtifactTargetWorkflow,
			harnessArtifactTargetDocs,
		},
		RequiredEvidence: []harnessEvidence{
			harnessEvidenceContract,
			harnessEvidenceBaseline,
			harnessEvidenceEvalPlan,
		},
		RequiredReviews: []harnessReview{
			harnessReviewProduct,
			harnessReviewEngineering,
			harnessReviewDesign,
		},
	})
	if err != nil {
		return harnessRequest{}
	}
	return req
}

func inferredPlanningHarnessRequest(kind, description string) *HarnessRequest {
	switch strings.TrimSpace(kind) {
	case "harness":
		req := defaultHarnessRequest()
		return &req
	case "plan":
		req, ok := inferCoreHarnessPlanRequest(description)
		if !ok {
			return nil
		}
		return &req
	default:
		return nil
	}
}

func inferCoreHarnessPlanRequest(description string) (harnessRequest, bool) {
	if !isCoreHarnessPlanDescription(description) {
		return harnessRequest{}, false
	}
	req, err := normalizeHarnessRequest(harnessRequest{
		RequestKind:      harnessRequestKindCore,
		DeliveryMode:     harnessDeliveryModeSpec,
		AdaptationMode:   harnessAdaptationModifyCore,
		BaseContractRef:  "namba-core-harness",
		TouchesNambaCore: true,
		ArtifactTargets:  inferCoreHarnessArtifactTargets(description),
		RequiredEvidence: inferCoreHarnessRequiredEvidence(description),
		RequiredReviews: []harnessReview{
			harnessReviewProduct,
			harnessReviewEngineering,
			harnessReviewDesign,
		},
	})
	if err != nil {
		return harnessRequest{}, false
	}
	return req, true
}

func isCoreHarnessPlanDescription(description string) bool {
	text := strings.ToLower(strings.TrimSpace(description))
	if text == "" {
		return false
	}

	explicitPlatform := containsAnyNormalizedToken(text,
		"namba",
		"codex",
		".namba",
		"built-in",
		"builtin",
		"command-entry",
		"core harness",
		"core contract",
	)
	contractSignals := containsAnyNormalizedToken(text,
		"harness",
		"route",
		"routing",
		"validator",
		"readiness",
		"classifier",
		"classification",
		"review",
		"workflow",
		"orchestration",
		"skill",
		"agent",
		"execution",
		"command",
	)
	if explicitPlatform && contractSignals {
		return true
	}

	if strings.Contains(text, "harness") && containsAnyNormalizedToken(text,
		"classifier",
		"classification",
		"route",
		"routing",
		"validator",
		"readiness",
		"namba plan",
		"namba harness",
		"$namba-create",
	) {
		return true
	}
	return false
}

func inferCoreHarnessArtifactTargets(description string) []harnessArtifactTarget {
	text := strings.ToLower(description)
	targets := []harnessArtifactTarget{harnessArtifactTargetWorkflow}
	if containsAnyNormalizedToken(text, "validator", "readiness", "classifier", "classification") {
		targets = append(targets, harnessArtifactTargetValidator)
	}
	if containsAnyNormalizedToken(text, "skill", "$namba-create", "artifact") {
		targets = append(targets, harnessArtifactTargetSkill)
	}
	if containsAnyNormalizedToken(text, "agent") {
		targets = append(targets, harnessArtifactTargetAgent)
	}
	if containsAnyNormalizedToken(text, "doc", "docs", "readme", "guide", "help", "contract") {
		targets = append(targets, harnessArtifactTargetDocs)
	}
	return normalizeHarnessArtifactTargets(targets)
}

func inferCoreHarnessRequiredEvidence(description string) []harnessEvidence {
	text := strings.ToLower(description)
	evidence := []harnessEvidence{
		harnessEvidenceContract,
		harnessEvidenceBaseline,
		harnessEvidenceEvalPlan,
	}
	if containsAnyNormalizedToken(text,
		"$namba-create",
		"artifact",
		"skill",
		"agent",
		"direct",
		"compose",
		"adapt",
		"domain harness",
	) {
		evidence = append(evidence, harnessEvidenceHarnessMap)
	}
	return normalizeHarnessEvidence(evidence)
}

func marshalHarnessRequest(req harnessRequest) (string, error) {
	normalized, err := normalizeHarnessRequest(req)
	if err != nil {
		return "", err
	}
	body, err := json.MarshalIndent(normalized, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshal harness request: %w", err)
	}
	return string(body), nil
}

func loadHarnessRequest(root, specID string) (*HarnessRequest, error) {
	if strings.TrimSpace(root) == "" || strings.TrimSpace(specID) == "" {
		return nil, nil
	}
	path := filepath.Join(root, filepath.FromSlash(specHarnessRequestPath(specID)))
	if !exists(path) {
		return nil, nil
	}
	body, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read harness request: %w", err)
	}
	req, err := decodeHarnessRequest(body)
	if err != nil {
		return nil, err
	}
	normalized, err := normalizeHarnessRequest(req)
	if err != nil {
		return nil, err
	}
	return &normalized, nil
}

func validateHarnessEvidence(root, specID string, req harnessRequest) harnessEvidenceValidationReport {
	requiredEvidence := normalizeHarnessEvidence(req.RequiredEvidence)
	if harnessRequestRequiresHarnessMap(req) && !hasHarnessEvidence(requiredEvidence, harnessEvidenceHarnessMap) {
		requiredEvidence = normalizeHarnessEvidence(append(requiredEvidence, harnessEvidenceHarnessMap))
	}
	requiredReviews := normalizeHarnessReviews(req.RequiredReviews)
	report := harnessEvidenceValidationReport{
		Request:          req,
		RequiredEvidence: harnessEvidenceStrings(requiredEvidence),
		RequiredReviews:  harnessReviewStrings(requiredReviews),
	}
	if route, err := harnessRouteForRequest(req); err == nil {
		report.Route = route
	}
	normalized, err := normalizeHarnessRequest(req)
	if err != nil {
		report.Problems = append(report.Problems, err.Error())
	} else {
		report.Request = normalized
		report.RequiredEvidence = harnessEvidenceStrings(normalized.RequiredEvidence)
		report.RequiredReviews = harnessReviewStrings(normalized.RequiredReviews)
		route, err := harnessRouteForRequest(normalized)
		if err != nil {
			report.Problems = append(report.Problems, err.Error())
		} else {
			report.Route = route
		}
		if err := validatePersistedSpecHarnessRequest(normalized); err != nil {
			report.Problems = append(report.Problems, err.Error())
		}
	}

	for _, evidence := range requiredEvidence {
		rel := specHarnessEvidencePath(specID, evidence)
		if rel == "" {
			continue
		}
		if !exists(filepath.Join(root, filepath.FromSlash(rel))) {
			report.MissingEvidence = append(report.MissingEvidence, rel)
		}
	}
	for _, review := range requiredReviews {
		rel := specReviewPath(specID, string(review))
		if !exists(filepath.Join(root, filepath.FromSlash(rel))) {
			report.MissingReviews = append(report.MissingReviews, rel)
		}
	}
	return report
}

func validatePersistedSpecHarnessRequest(req harnessRequest) error {
	if req.RequestKind == harnessRequestKindDirect {
		return fmt.Errorf("direct_artifact_creation must not be persisted to `%s`; keep it transient on `%s` or escalate through `namba plan`", harnessRequestFileName, harnessRouteCreate)
	}
	return nil
}

func inspectSpecHarnessRequest(root, specID string) (harnessValidationResult, bool) {
	path := filepath.Join(root, filepath.FromSlash(specHarnessRequestPath(specID)))
	if !exists(path) {
		return harnessValidationResult{}, false
	}

	body, err := os.ReadFile(path)
	if err != nil {
		return harnessValidationResult{Issues: []string{fmt.Sprintf("read harness metadata: %v", err)}}, true
	}
	req, err := decodeHarnessRequest(body)
	if err != nil {
		return harnessValidationResult{Issues: []string{err.Error()}}, true
	}
	req, err = normalizeHarnessRequest(req)
	if err != nil {
		return harnessValidationResult{Issues: []string{err.Error()}}, true
	}
	route, err := harnessRouteForRequest(req)
	if err != nil {
		return harnessValidationResult{Issues: []string{err.Error()}}, true
	}

	var issues []string
	for _, evidence := range req.RequiredEvidence {
		rel := specHarnessEvidencePath(specID, evidence)
		if strings.TrimSpace(rel) == "" {
			continue
		}
		if !exists(filepath.Join(root, filepath.FromSlash(rel))) {
			issues = append(issues, fmt.Sprintf("missing evidence artifact `%s`", rel))
		}
	}
	return harnessValidationResult{
		Request: req,
		Route:   route,
		Issues:  issues,
	}, true
}

func harnessAdvisorySummary(root, specID string) string {
	result, ok := inspectSpecHarnessRequest(root, specID)
	if !ok || len(result.Issues) == 0 {
		return ""
	}
	return strings.Join(result.Issues, "; ")
}

func specHarnessEvidencePath(specID string, evidence harnessEvidence) string {
	switch evidence {
	case harnessEvidenceContract:
		return filepath.ToSlash(filepath.Join(specsDir, specID, "contract.md"))
	case harnessEvidenceBaseline:
		return filepath.ToSlash(filepath.Join(specsDir, specID, "baseline.md"))
	case harnessEvidenceEvalPlan:
		return filepath.ToSlash(filepath.Join(specsDir, specID, "eval-plan.md"))
	case harnessEvidenceHarnessMap:
		return filepath.ToSlash(filepath.Join(specsDir, specID, "harness-map.md"))
	default:
		return ""
	}
}

func decodeHarnessRequest(data []byte) (harnessRequest, error) {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()

	var req harnessRequest
	if err := decoder.Decode(&req); err != nil {
		return harnessRequest{}, fmt.Errorf("decode harness request: %w", err)
	}
	var extra struct{}
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return harnessRequest{}, fmt.Errorf("decode harness request: unexpected trailing JSON input")
		}
		return harnessRequest{}, fmt.Errorf("decode harness request: %w", err)
	}
	return req, nil
}

func normalizeHarnessArtifactTargets(values []harnessArtifactTarget) []harnessArtifactTarget {
	seen := map[harnessArtifactTarget]struct{}{}
	var normalized []harnessArtifactTarget
	for _, value := range values {
		value = harnessArtifactTarget(strings.TrimSpace(string(value)))
		if value == "" {
			continue
		}
		if _, ok := harnessArtifactTargetOrder[value]; !ok {
			normalized = append(normalized, value)
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		normalized = append(normalized, value)
	}
	sort.Slice(normalized, func(i, j int) bool {
		return harnessArtifactSortKey(normalized[i]) < harnessArtifactSortKey(normalized[j])
	})
	return normalized
}

func normalizeHarnessEvidence(values []harnessEvidence) []harnessEvidence {
	seen := map[harnessEvidence]struct{}{}
	var normalized []harnessEvidence
	for _, value := range values {
		value = harnessEvidence(strings.TrimSpace(string(value)))
		if value == "" {
			continue
		}
		if _, ok := harnessEvidenceOrder[value]; !ok {
			normalized = append(normalized, value)
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		normalized = append(normalized, value)
	}
	sort.Slice(normalized, func(i, j int) bool {
		return harnessEvidenceSortKey(normalized[i]) < harnessEvidenceSortKey(normalized[j])
	})
	return normalized
}

func normalizeHarnessReviews(values []harnessReview) []harnessReview {
	seen := map[harnessReview]struct{}{}
	var normalized []harnessReview
	for _, value := range values {
		value = harnessReview(strings.TrimSpace(string(value)))
		if value == "" {
			continue
		}
		if _, ok := harnessReviewOrder[value]; !ok {
			normalized = append(normalized, value)
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		normalized = append(normalized, value)
	}
	sort.Slice(normalized, func(i, j int) bool {
		return harnessReviewSortKey(normalized[i]) < harnessReviewSortKey(normalized[j])
	})
	return normalized
}

func harnessRequestKindSet() map[harnessRequestKind]struct{} {
	return map[harnessRequestKind]struct{}{
		harnessRequestKindCore:   {},
		harnessRequestKindDomain: {},
		harnessRequestKindDirect: {},
	}
}

func harnessDeliveryModeSet() map[harnessDeliveryMode]struct{} {
	return map[harnessDeliveryMode]struct{}{
		harnessDeliveryModeSpec:   {},
		harnessDeliveryModeDirect: {},
	}
}

func harnessAdaptationModeSet() map[harnessAdaptationMode]struct{} {
	return map[harnessAdaptationMode]struct{}{
		harnessAdaptationModifyCore:       {},
		harnessAdaptationExtendDomain:     {},
		harnessAdaptationComposeDomain:    {},
		harnessAdaptationGenerateArtifact: {},
	}
}

func hasAllHarnessEvidence(values []harnessEvidence, required ...harnessEvidence) bool {
	for _, item := range required {
		if !hasHarnessEvidence(values, item) {
			return false
		}
	}
	return true
}

func hasHarnessEvidence(values []harnessEvidence, target harnessEvidence) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func hasAllHarnessReviews(values []harnessReview, required ...harnessReview) bool {
	for _, item := range required {
		if !hasHarnessReview(values, item) {
			return false
		}
	}
	return true
}

func hasHarnessReview(values []harnessReview, target harnessReview) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func expectedCreateHarnessTargets(target createTarget) []harnessArtifactTarget {
	switch target {
	case createTargetSkill:
		return []harnessArtifactTarget{harnessArtifactTargetSkill}
	case createTargetAgent:
		return []harnessArtifactTarget{harnessArtifactTargetAgent}
	case createTargetBoth:
		return []harnessArtifactTarget{harnessArtifactTargetSkill, harnessArtifactTargetAgent}
	default:
		return nil
	}
}

func equalHarnessArtifactTargets(left, right []harnessArtifactTarget) bool {
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if left[i] != right[i] {
			return false
		}
	}
	return true
}

func harnessArtifactTargetsToStrings(values []harnessArtifactTarget) []string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		out = append(out, string(value))
	}
	return out
}

func harnessEvidenceStrings(values []harnessEvidence) []string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		out = append(out, string(value))
	}
	return out
}

func harnessReviewStrings(values []harnessReview) []string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		out = append(out, string(value))
	}
	return out
}

func harnessArtifactSortKey(value harnessArtifactTarget) int {
	if order, ok := harnessArtifactTargetOrder[value]; ok {
		return order
	}
	return len(harnessArtifactTargetOrder) + 1
}

func harnessEvidenceSortKey(value harnessEvidence) int {
	if order, ok := harnessEvidenceOrder[value]; ok {
		return order
	}
	return len(harnessEvidenceOrder) + 1
}

func harnessReviewSortKey(value harnessReview) int {
	if order, ok := harnessReviewOrder[value]; ok {
		return order
	}
	return len(harnessReviewOrder) + 1
}

func containsAnyNormalizedToken(text string, tokens ...string) bool {
	for _, token := range tokens {
		if strings.Contains(text, strings.ToLower(strings.TrimSpace(token))) {
			return true
		}
	}
	return false
}
