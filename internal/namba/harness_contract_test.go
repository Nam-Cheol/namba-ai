package namba

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
)

type harnessRequestFixture struct {
	RequestKind      string   `json:"request_kind"`
	DeliveryMode     string   `json:"delivery_mode"`
	AdaptationMode   string   `json:"adaptation_mode"`
	BaseContractRef  string   `json:"base_contract_ref"`
	TouchesNambaCore bool     `json:"touches_namba_core"`
	ArtifactTargets  []string `json:"artifact_targets"`
	RequiredEvidence []string `json:"required_evidence"`
	RequiredReviews  []string `json:"required_reviews"`
}

func TestDefaultHarnessRequestIsCanonicalAndComplete(t *testing.T) {
	t.Parallel()

	req := defaultHarnessRequest()
	if req.RequestKind != harnessRequestKindDomain {
		t.Fatalf("unexpected request kind: %+v", req)
	}
	if req.DeliveryMode != harnessDeliveryModeSpec || req.AdaptationMode != harnessAdaptationExtendDomain {
		t.Fatalf("unexpected delivery/adaptation metadata: %+v", req)
	}
	if req.BaseContractRef != "" || req.TouchesNambaCore {
		t.Fatalf("unexpected core-boundary metadata: %+v", req)
	}
	if !equalHarnessArtifactTargets(req.ArtifactTargets, []harnessArtifactTarget{harnessArtifactTargetWorkflow, harnessArtifactTargetDocs}) {
		t.Fatalf("unexpected artifact targets: %+v", req.ArtifactTargets)
	}
	if !hasAllHarnessEvidence(req.RequiredEvidence, harnessEvidenceContract, harnessEvidenceBaseline, harnessEvidenceEvalPlan) {
		t.Fatalf("expected default request to require the harness evidence pack, got %+v", req.RequiredEvidence)
	}
	if !hasAllHarnessReviews(req.RequiredReviews, harnessReviewProduct, harnessReviewEngineering, harnessReviewDesign) {
		t.Fatalf("expected default request to require the current review runtime, got %+v", req.RequiredReviews)
	}
}

func TestInferCoreHarnessPlanRequestClassifiesCorePlanningWork(t *testing.T) {
	t.Parallel()

	req, ok := inferCoreHarnessPlanRequest("tighten Namba harness classification and readiness validator")
	if !ok {
		t.Fatal("expected core harness plan request to be classified")
	}
	if req.RequestKind != harnessRequestKindCore || req.DeliveryMode != harnessDeliveryModeSpec || req.AdaptationMode != harnessAdaptationModifyCore {
		t.Fatalf("unexpected core harness request metadata: %+v", req)
	}
	if req.BaseContractRef != "namba-core-harness" || !req.TouchesNambaCore {
		t.Fatalf("expected Namba core metadata, got %+v", req)
	}
	if !equalHarnessArtifactTargets(req.ArtifactTargets, []harnessArtifactTarget{harnessArtifactTargetWorkflow, harnessArtifactTargetValidator}) {
		t.Fatalf("unexpected artifact targets: %+v", req.ArtifactTargets)
	}
	if !reflect.DeepEqual(req.RequiredEvidence, []harnessEvidence{harnessEvidenceContract, harnessEvidenceBaseline, harnessEvidenceEvalPlan}) {
		t.Fatalf("unexpected required evidence: %+v", req.RequiredEvidence)
	}
}

func TestInferCoreHarnessPlanRequestSkipsRegularFeaturePlans(t *testing.T) {
	t.Parallel()

	if req, ok := inferCoreHarnessPlanRequest("add dashboard filters"); ok || !reflect.DeepEqual(req, harnessRequest{}) {
		t.Fatalf("expected regular feature plan to avoid harness sidecar, got ok=%v req=%+v", ok, req)
	}
}

func TestHarnessRequestTransportRoundTripUsesSpecSidecar(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	specID := "SPEC-777"
	specDir := filepath.Join(tmp, ".namba", "specs", specID)
	if err := os.MkdirAll(filepath.Join(specDir, "reviews"), 0o755); err != nil {
		t.Fatalf("mkdir spec reviews dir: %v", err)
	}
	for _, rel := range []string{"contract.md", "baseline.md", "eval-plan.md"} {
		writeTestFile(t, filepath.Join(specDir, rel), "# "+rel+"\n")
	}
	for _, rel := range []string{"product.md", "engineering.md", "design.md"} {
		writeTestFile(t, filepath.Join(specDir, "reviews", rel), "# "+rel+"\n")
	}

	req := defaultHarnessRequest()
	body, err := marshalHarnessRequest(req)
	if err != nil {
		t.Fatalf("marshal harness request: %v", err)
	}
	writeTestFile(t, filepath.Join(specDir, harnessRequestFileName), body)

	loaded, err := loadHarnessRequest(tmp, specID)
	if err != nil {
		t.Fatalf("load harness request: %v", err)
	}
	if loaded == nil {
		t.Fatal("expected harness request to load from spec sidecar")
	}
	if !reflect.DeepEqual(*loaded, req) {
		t.Fatalf("unexpected loaded request: got %+v want %+v", loaded, req)
	}

	report := validateHarnessEvidence(tmp, specID, *loaded)
	if report.Route != harnessRouteHarness {
		t.Fatalf("unexpected route: %+v", report)
	}
	if len(report.MissingEvidence) != 0 || len(report.MissingReviews) != 0 {
		t.Fatalf("expected complete evidence for default harness request, got %+v", report)
	}
}

func TestValidateHarnessEvidenceRequiresHarnessMapForComposedDomainRequests(t *testing.T) {
	t.Parallel()

	tmp := t.TempDir()
	specID := "SPEC-778"
	specDir := filepath.Join(tmp, ".namba", "specs", specID)
	if err := os.MkdirAll(filepath.Join(specDir, "reviews"), 0o755); err != nil {
		t.Fatalf("mkdir spec reviews dir: %v", err)
	}
	for _, rel := range []string{"contract.md", "baseline.md", "eval-plan.md"} {
		writeTestFile(t, filepath.Join(specDir, rel), "# "+rel+"\n")
	}
	for _, rel := range []string{"product.md", "engineering.md", "design.md"} {
		writeTestFile(t, filepath.Join(specDir, "reviews", rel), "# "+rel+"\n")
	}

	req := harnessRequest{
		RequestKind:      harnessRequestKindDomain,
		DeliveryMode:     harnessDeliveryModeSpec,
		AdaptationMode:   harnessAdaptationComposeDomain,
		BaseContractRef:  "finance-harness",
		TouchesNambaCore: false,
		ArtifactTargets:  []harnessArtifactTarget{harnessArtifactTargetWorkflow, harnessArtifactTargetDocs},
		RequiredEvidence: []harnessEvidence{harnessEvidenceContract, harnessEvidenceBaseline, harnessEvidenceEvalPlan},
		RequiredReviews:  []harnessReview{harnessReviewProduct, harnessReviewEngineering, harnessReviewDesign},
	}

	report := validateHarnessEvidence(tmp, specID, req)
	if report.Route != harnessRouteHarness {
		t.Fatalf("unexpected route for domain harness request: %+v", report)
	}
	if !containsString(report.RequiredEvidence, string(harnessEvidenceHarnessMap)) {
		t.Fatalf("expected harness map to be required for composed domain requests, got %+v", report)
	}
	if !containsString(report.MissingEvidence, filepath.ToSlash(filepath.Join(specsDir, specID, "harness-map.md"))) {
		t.Fatalf("expected missing harness map evidence, got %+v", report.MissingEvidence)
	}
	if len(report.MissingReviews) != 0 {
		t.Fatalf("expected review tracks to be complete, got %+v", report.MissingReviews)
	}
}

func TestValidateDirectCreateHarnessRequestRejectsCoreEscalation(t *testing.T) {
	t.Parallel()

	allowed := &HarnessRequest{
		RequestKind:      harnessRequestKindDirect,
		DeliveryMode:     harnessDeliveryModeDirect,
		AdaptationMode:   harnessAdaptationGenerateArtifact,
		TouchesNambaCore: false,
		ArtifactTargets:  []harnessArtifactTarget{harnessArtifactTargetSkill, harnessArtifactTargetAgent},
	}
	if err := validateDirectCreateHarnessRequest(allowed); err != nil {
		t.Fatalf("expected resolved direct request to remain on create route: %v", err)
	}

	blocked := &HarnessRequest{
		RequestKind:      harnessRequestKindDirect,
		DeliveryMode:     harnessDeliveryModeDirect,
		AdaptationMode:   harnessAdaptationGenerateArtifact,
		TouchesNambaCore: true,
		ArtifactTargets:  []harnessArtifactTarget{harnessArtifactTargetSkill, harnessArtifactTargetAgent},
	}
	if err := validateDirectCreateHarnessRequest(blocked); err == nil || !strings.Contains(err.Error(), "route to `namba plan`") {
		t.Fatalf("expected direct route escalation failure, got %v", err)
	}
}

func TestValidateHarnessEvidenceFlagsPersistedDirectRequest(t *testing.T) {
	t.Parallel()

	report := validateHarnessEvidence(t.TempDir(), "SPEC-779", harnessRequest{
		RequestKind:      harnessRequestKindDirect,
		DeliveryMode:     harnessDeliveryModeDirect,
		AdaptationMode:   harnessAdaptationGenerateArtifact,
		TouchesNambaCore: false,
		ArtifactTargets:  []harnessArtifactTarget{harnessArtifactTargetSkill},
	})
	if report.Route != harnessRouteCreate {
		t.Fatalf("expected persisted direct request to keep create route for diagnostics, got %+v", report)
	}
	if !containsString(report.Problems, "direct_artifact_creation must not be persisted to `harness-request.json`; keep it transient on `$namba-create` or escalate through `namba plan`") {
		t.Fatalf("expected persisted direct request to be flagged as invalid spec state, got %+v", report.Problems)
	}
	if len(report.MissingEvidence) != 0 || len(report.MissingReviews) != 0 {
		t.Fatalf("expected persisted direct request failure to come from problems, got %+v", report)
	}
}

func TestSpec032HarnessRequestFixtureIsCanonicalAndComplete(t *testing.T) {
	t.Parallel()

	path := repoFixturePath(t, filepath.Join(".namba", "specs", "SPEC-032", "harness-request.json"))
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read harness request fixture: %v", err)
	}

	var req harnessRequestFixture
	if err := json.Unmarshal(data, &req); err != nil {
		t.Fatalf("unmarshal harness request fixture: %v", err)
	}

	if req.RequestKind != "core_harness_change" {
		t.Fatalf("unexpected request kind: %+v", req)
	}
	if req.DeliveryMode != "spec" || req.AdaptationMode != "modify_core" {
		t.Fatalf("unexpected route metadata: %+v", req)
	}
	if req.BaseContractRef != "namba-core-harness" || !req.TouchesNambaCore {
		t.Fatalf("unexpected adaptation metadata: %+v", req)
	}

	for _, want := range []string{"workflow", "validator", "docs"} {
		if !containsString(req.ArtifactTargets, want) {
			t.Fatalf("expected artifact targets to contain %q, got %+v", want, req.ArtifactTargets)
		}
	}
	for _, want := range []string{"contract", "baseline", "eval-plan", "harness-map"} {
		if !containsString(req.RequiredEvidence, want) {
			t.Fatalf("expected required evidence to contain %q, got %+v", want, req.RequiredEvidence)
		}
	}
	for _, want := range []string{"product", "engineering", "design"} {
		if !containsString(req.RequiredReviews, want) {
			t.Fatalf("expected required reviews to contain %q, got %+v", want, req.RequiredReviews)
		}
	}
}

func TestSpec032ContractAndEvalPlanPreserveRoutePrecedenceAndEscalation(t *testing.T) {
	t.Parallel()

	contract := mustReadContractFixture(t, repoFixturePath(t, filepath.Join(".namba", "specs", "SPEC-032", "contract.md")))
	for _, want := range []string{
		"Routing Precedence",
		"1. `namba plan`",
		"2. `namba harness`",
		"3. `$namba-create`",
		"Never downgrade to `$namba-create` while core contract ambiguity remains",
		"Escalate to `core_harness_change` if `touches_namba_core=true` becomes true",
	} {
		if !strings.Contains(contract, want) {
			t.Fatalf("expected contract to contain %q, got %q", want, contract)
		}
	}

	evalPlan := mustReadContractFixture(t, repoFixturePath(t, filepath.Join(".namba", "specs", "SPEC-032", "eval-plan.md")))
	for _, want := range []string{
		"Add a brand-new reusable finance-analysis domain harness without touching built-in Namba routing",
		"Create a repo-local artifact that also changes Namba-managed harness semantics",
		"reject direct route, escalate to `core_harness_change`",
		"preview/apply evidence using the transient JSON transport",
	} {
		if !strings.Contains(evalPlan, want) {
			t.Fatalf("expected eval plan to contain %q, got %q", want, evalPlan)
		}
	}
}

func mustReadContractFixture(t *testing.T, path string) string {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(data)
}

func repoFixturePath(t *testing.T, rel string) string {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve caller path")
	}
	cwd := filepath.Dir(file)
	for {
		candidate := filepath.Join(cwd, rel)
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
		parent := filepath.Dir(cwd)
		if parent == cwd {
			break
		}
		cwd = parent
	}
	t.Fatalf("could not find fixture %s from %s", rel, cwd)
	return ""
}
