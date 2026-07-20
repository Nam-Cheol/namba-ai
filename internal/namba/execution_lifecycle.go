package namba

import (
	"context"
	"errors"
	"strings"
	"time"
)

const defaultCodexCapabilityProbeTimeout = 10 * time.Second

const (
	lifecycleProbeCodexVersion           = "codex_version"
	lifecycleProbeCodexExecHelp          = "codex_exec_help"
	lifecycleProbeCodexModelAvailability = "codex_model_availability"
	lifecycleProbeSolAvailability        = "codex_sol_availability"
	lifecycleProbeCodexResumeExecHelp    = "codex_resume_exec_help"
)

const (
	lifecycleProbeStatusSucceeded   = "succeeded"
	lifecycleProbeStatusAvailable   = "available"
	lifecycleProbeStatusUnavailable = "unavailable"
	lifecycleProbeStatusUnsupported = "unsupported"
	lifecycleProbeStatusError       = "error"
	lifecycleProbeStatusTimedOut    = "timed_out"
	lifecycleProbeStatusCanceled    = "canceled"
)

type lifecycleProbeOutcome struct {
	Name       string `json:"name"`
	Model      string `json:"model,omitempty"`
	Status     string `json:"status"`
	TimeoutMS  int64  `json:"timeout_ms,omitempty"`
	DurationMS int64  `json:"duration_ms"`
}

func runBoundedLifecycleProbe(ctx context.Context, name string, timeout time.Duration, probe func(context.Context) error) (lifecycleProbeOutcome, error) {
	if timeout <= 0 {
		timeout = defaultCodexCapabilityProbeTimeout
	}
	outcome := lifecycleProbeOutcome{
		Name:      strings.TrimSpace(name),
		Status:    lifecycleProbeStatusSucceeded,
		TimeoutMS: timeout.Milliseconds(),
	}
	started := time.Now()
	probeCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	err := probe(probeCtx)
	outcome.DurationMS = time.Since(started).Milliseconds()
	switch {
	case errors.Is(probeCtx.Err(), context.DeadlineExceeded):
		outcome.Status = lifecycleProbeStatusTimedOut
		return outcome, context.DeadlineExceeded
	case errors.Is(probeCtx.Err(), context.Canceled):
		outcome.Status = lifecycleProbeStatusCanceled
		return outcome, context.Canceled
	case err != nil:
		outcome.Status = lifecycleProbeStatusError
		return outcome, err
	default:
		return outcome, nil
	}
}

// executionLifecycleState is the single mutable boundary for one execution:
// the capability snapshot, full-input routing plan, reached routing evidence,
// and writable resume authority all advance together and feed artifact output.
type executionLifecycleState struct {
	request                 executionRequest
	capabilities            codexCapabilityMatrix
	modelRoutingPlan        []executionRequest
	modelRoutingTurns       []executionRequest
	latestWritableThreadIDs map[string]string
}

func newExecutionLifecycleState(req executionRequest) *executionLifecycleState {
	return &executionLifecycleState{
		request:                 req,
		latestWritableThreadIDs: make(map[string]string),
	}
}

func (s *executionLifecycleState) requestSnapshot() executionRequest {
	if s == nil {
		return executionRequest{}
	}
	return s.request
}

func (s *executionLifecycleState) recordPreflight(capabilities codexCapabilityMatrix) {
	if s == nil {
		return
	}
	s.capabilities = capabilities
	s.request = withModelAvailability(s.request, capabilities)
	s.modelRoutingPlan = append(s.modelRoutingPlan[:0], plannedExecutionTurnRequests(s.request)...)
}

func (s *executionLifecycleState) plannedExecutionTurns() []executionRequest {
	if s == nil {
		return nil
	}
	return append([]executionRequest(nil), s.modelRoutingPlan...)
}

func (s *executionLifecycleState) replaceReachedRoutingTurns(turns []executionRequest) {
	if s == nil {
		return
	}
	s.modelRoutingTurns = append(s.modelRoutingTurns[:0], turns...)
}

func (s *executionLifecycleState) appendReachedRoutingTurn(turn executionRequest) {
	if s == nil {
		return
	}
	s.modelRoutingTurns = append(s.modelRoutingTurns, turn)
}

// recordReachedTurnObservation attaches a canonical UUID observed from a fresh
// execution to its reached routing evidence. Resume turns retain the UUID they
// actually used as input instead of being rewritten by a later output value.
func (s *executionLifecycleState) recordReachedTurnObservation(turn executionRequest, threadID string) executionRequest {
	if s == nil {
		return turn
	}
	threadID = strings.TrimSpace(threadID)
	if !turn.ResumeSession && isCodexThreadUUID(threadID) {
		turn.ThreadID = threadID
	}
	if len(s.modelRoutingTurns) > 0 {
		s.modelRoutingTurns[len(s.modelRoutingTurns)-1] = turn
	}
	return turn
}

func (s *executionLifecycleState) modelRoutingEvidence() (*modelRoutingEvidence, []modelRoutingEvidence) {
	if s == nil {
		return nil, nil
	}
	reached := modelRoutingEvidenceForRequests(s.modelRoutingTurns)
	primary := modelRoutingEvidenceForRequest(s.request)
	if planned := modelRoutingEvidenceForRequests(s.modelRoutingPlan); len(planned) > 0 {
		primary = &planned[0]
	}
	if len(reached) > 0 {
		primary = &reached[0]
	}
	return primary, reached
}

func (s *executionLifecycleState) latestWritableThreadID(model string) string {
	if s == nil {
		return ""
	}
	return s.latestWritableThreadIDs[strings.TrimSpace(model)]
}

// observeWritableThread makes resume authority match the latest writable turn
// for a model. A missing or malformed UUID revokes older authority, while a
// read-only checkpoint cannot create or revoke writable continuation state.
func (s *executionLifecycleState) observeWritableThread(req executionRequest, threadID string) bool {
	if s == nil || req.RoutingDecision.ReadOnly {
		return false
	}
	model := strings.TrimSpace(req.Model)
	if model == "" {
		return false
	}
	threadID = strings.TrimSpace(threadID)
	if !isCodexThreadUUID(threadID) {
		delete(s.latestWritableThreadIDs, model)
		return false
	}
	s.latestWritableThreadIDs[model] = threadID
	return true
}
